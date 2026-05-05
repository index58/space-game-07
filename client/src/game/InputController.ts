import { INITIAL_ZOOM, clampZoom } from "../domain/camera";
import type { ChatSendMessage, ChatStateMessage, ClientInputState } from "../network/protocol";
import { isFreshKeyDown, toShipInput } from "./inputState";

export type ChatInputAction = Omit<ChatSendMessage, "type">;

export type ChatContextMenuState = {
  // Чат, к которому относится открытое меню.
  chatId: number;
  // Тип вкладки, определяющий доступность локального закрытия.
  communityTypeAcronym: string;
  // Горизонтальная координата меню в пикселях окна.
  x: number;
  // Вертикальная координата меню в пикселях окна.
  y: number;
};

// Изолирует браузерные события ввода от игровой сцены.
export class InputController {
  // Текущее состояние клавиш по DOM-кодам.
  private readonly keys: Record<string, boolean> = {};
  // Накопленное горизонтальное движение мыши между кадрами.
  private mouseDeltaX = 0;
  // Дискретный пользовательский уровень приближения.
  private zoom = INITIAL_ZOOM;
  // Одноразовый запрос на смену модели корабля.
  private randomShipChangeRequested = false;
  // Одноразовый запрос на переключение отладочной отрисовки тел.
  private bodyPolygonDebugToggleRequested = false;
  // Накопленное переключение выбранного инструмента пилота.
  private pilotToolSelectionDelta = 0;
  // Одноразовый запрос на переключение якоря.
  private anchorToggleRequested = false;
  // Показывает, что печатные клавиши сейчас направлены в чат.
  private chatInputFocused = false;
  // Локальная строка, которую игрок набирает до отправки.
  private chatInputText = "";
  // Позиция вставки внутри локальной строки.
  private chatCursorIndex = 0;
  // Последний выбранный сервером или локальным вводом чат.
  private selectedChatId = 0;
  // Очередь одноразовых команд отправки текста.
  private chatActions: ChatInputAction[] = [];
  // Локально закрытые дуэты, которые не должны отображаться в HUD.
  private readonly closedDuoChatIds = new Set<number>();
  // Последнее состояние вкладок для hit-test игрового указателя.
  private visibleChatState: ChatStateMessage | null = null;
  // Открытое игровое меню вкладки чата.
  private chatContextMenu: ChatContextMenuState | null = null;

  constructor(
    // Игровой canvas, который получает захват указателя.
    private readonly canvas: HTMLCanvasElement,
    // Проверка готовности сцены к захвату мыши.
    private readonly canRequestPointerLock: () => boolean = () => true,
  ) {
    // Состояние клавиш хранится непрерывно, потому что сетевой ввод отправляется реже кадров браузера.
    window.addEventListener("keydown", (event) => {
      if (!this.isPointerLocked()) {
        this.chatInputFocused = false;
        this.chatContextMenu = null;
        this.keys[event.code] = true;
        return;
      }
      if (this.handleChatKeyDown(event)) {
        return;
      }
      if (isFreshKeyDown(event.code, Boolean(this.keys[event.code]), "Backslash")) {
        this.randomShipChangeRequested = true;
      }
      if (isFreshKeyDown(event.code, Boolean(this.keys[event.code]), "KeyO")) {
        this.bodyPolygonDebugToggleRequested = true;
      }
      if (isFreshKeyDown(event.code, Boolean(this.keys[event.code]), "KeyP")) {
        this.anchorToggleRequested = true;
      }

      this.keys[event.code] = true;
    });

    window.addEventListener("keyup", (event) => {
      this.keys[event.code] = false;
    });

    window.addEventListener("mousemove", (event) => {
      if (document.pointerLockElement === this.canvas) {
        this.mouseDeltaX += event.movementX;
      }
    });

    // Колесо мыши меняет дискретный уровень зума, который затем переводится камерой в масштаб.
    window.addEventListener(
      "wheel",
      (event) => {
        if (!this.isPointerLocked()) {
          return;
        }
        if (event.shiftKey) {
          this.pilotToolSelectionDelta += event.deltaY > 0 ? 1 : -1;
          return;
        }
        this.zoom = clampZoom(this.zoom + (event.deltaY > 0 ? -1 : 1));
      },
      { passive: true },
    );

    window.addEventListener("contextmenu", (event) => {
      if (this.isPointerLocked() && this.openChatContextMenu(event.clientX, event.clientY)) {
        event.preventDefault();
      }
    });

    window.addEventListener("mousedown", (event) => {
      if (event.button !== 0) {
        return;
      }
      if (!this.isPointerLocked()) {
        this.chatContextMenu = null;
        return;
      }
      if (this.closeChatContextMenuItem(event.clientX, event.clientY)) {
        event.preventDefault();
      } else {
        this.chatContextMenu = null;
      }
    });

    // Захват мыши включается только по клику и только когда клиент уже готов принимать управление.
    this.canvas.addEventListener("click", () => {
      if (!this.canRequestPointerLock()) {
        return;
      }

      void this.canvas.requestPointerLock();
    });
  }

  // Возвращает пользовательский уровень зума без пересчета в пиксели.
  getZoom(): number {
    return this.zoom;
  }

  // Синхронизирует выбранную вкладку с сервером и убирает закрытые чаты из локального списка.
  getVisibleChatState(chatState: ChatStateMessage | null): ChatStateMessage | null {
    if (!chatState) {
      return null;
    }

    const tabs = chatState.tabs.filter((tab) => tab.communityTypeAcronym === "Server" || !this.closedDuoChatIds.has(tab.chatId));
    if (tabs.length === 0) {
      this.selectedChatId = 0;
      this.visibleChatState = { ...chatState, tabs: [], selectedChatId: 0 };
      return this.visibleChatState;
    }

    if (chatState.selectedChatId && !this.closedDuoChatIds.has(chatState.selectedChatId)) {
      this.selectedChatId = chatState.selectedChatId;
    }
    if (!tabs.some((tab) => tab.chatId === this.selectedChatId)) {
      this.selectedChatId = tabs[0].chatId;
    }

    this.visibleChatState = { ...chatState, tabs, selectedChatId: this.selectedChatId };
    return this.visibleChatState;
  }

  // Возвращает текущую строку, которую SolidJS только отображает.
  getChatInputText(): string {
    return this.chatInputText;
  }

  // Возвращает состояние фокуса, чтобы HUD мог показать каретку.
  isChatInputFocused(): boolean {
    return this.chatInputFocused;
  }

  // Выдает одну готовую команду отправки текста.
  consumeChatAction(): ChatInputAction | null {
    return this.chatActions.shift() ?? null;
  }

  // Возвращает открытое меню вкладки для отображения в SolidJS.
  getChatContextMenu(): ChatContextMenuState | null {
    return this.chatContextMenu;
  }

  // Закрывает локальную вкладку дуэта без изменения серверных данных.
  closeChatTab(chatId: number, communityTypeAcronym: string): void {
    if (communityTypeAcronym !== "Duo") {
      return;
    }

    this.closedDuoChatIds.add(chatId);
    this.chatContextMenu = null;
    if (this.selectedChatId === chatId) {
      this.selectedChatId = 0;
    }
  }

  // Возвращает дискретную команду один раз на одно нажатие клавиши.
  consumeRandomShipChangeRequest(): boolean {
    const requested = this.randomShipChangeRequested;
    this.randomShipChangeRequested = false;
    return requested;
  }

  // Возвращает запрос переключения отладочного слоя один раз на одно нажатие.
  consumeBodyPolygonDebugToggleRequest(): boolean {
    const requested = this.bodyPolygonDebugToggleRequested;
    this.bodyPolygonDebugToggleRequested = false;
    return requested;
  }

  // Возвращает накопленное переключение инструмента пилота и сразу сбрасывает его.
  consumePilotToolSelectionDelta(): number {
    const delta = this.pilotToolSelectionDelta;
    this.pilotToolSelectionDelta = 0;
    return delta;
  }

  // Отдает ввод за текущий кадр и сбрасывает накопленное движение мыши.
  consumeShipInput(): ClientInputState {
    // Захват указателя отдает относительное движение мыши; после кадра накопление сбрасывается.
    const isPointerLocked = this.isPointerLocked();
    if (!isPointerLocked) {
      this.chatInputFocused = false;
      this.chatContextMenu = null;
    }
    const input = toShipInput(isPointerLocked, this.keys, this.mouseDeltaX);
    input.toggleAnchor = isPointerLocked && this.anchorToggleRequested;
    this.mouseDeltaX = 0;
    this.anchorToggleRequested = false;

    return input;
  }

  // Обрабатывает клавиши чата раньше управления кораблем, когда строка активна.
  private handleChatKeyDown(event: KeyboardEvent): boolean {
    const wasPressed = Boolean(this.keys[event.code]);
    if (isFreshKeyDown(event.code, wasPressed, "Enter")) {
      if (!this.chatInputFocused) {
        this.chatInputFocused = true;
        this.keys[event.code] = true;
        event.preventDefault();
        return true;
      }

      this.queueChatAction();
      this.keys[event.code] = true;
      event.preventDefault();
      return true;
    }

    if (!this.chatInputFocused) {
      return false;
    }

    if (event.code === "Backspace") {
      if (this.chatCursorIndex > 0) {
        this.chatInputText = `${this.chatInputText.slice(0, this.chatCursorIndex - 1)}${this.chatInputText.slice(this.chatCursorIndex)}`;
        this.chatCursorIndex -= 1;
      }
      event.preventDefault();
      return true;
    }

    if (event.code === "Delete") {
      if (this.chatCursorIndex < this.chatInputText.length) {
        this.chatInputText = `${this.chatInputText.slice(0, this.chatCursorIndex)}${this.chatInputText.slice(this.chatCursorIndex + 1)}`;
      }
      event.preventDefault();
      return true;
    }

    if (event.code === "ArrowLeft") {
      this.chatCursorIndex = Math.max(0, this.chatCursorIndex - 1);
      event.preventDefault();
      return true;
    }

    if (event.code === "ArrowRight") {
      this.chatCursorIndex = Math.min(this.chatInputText.length, this.chatCursorIndex + 1);
      event.preventDefault();
      return true;
    }

    if (event.key.length === 1 && !event.ctrlKey && !event.metaKey && !event.altKey) {
      this.chatInputText = `${this.chatInputText.slice(0, this.chatCursorIndex)}${event.key}${this.chatInputText.slice(this.chatCursorIndex)}`;
      this.chatCursorIndex += event.key.length;
      event.preventDefault();
      return true;
    }

    return true;
  }

  // Превращает локальную строку в обычную или адресную сетевую команду.
  private queueChatAction(): void {
    const text = this.chatInputText.trim();
    if (text === "") {
      return;
    }

    const directMessage = text.match(/^@([^\s]+)\s+(.+)$/);
    if (directMessage) {
      this.chatActions.push({
        targetNickname: directMessage[1],
        text: directMessage[2].trim(),
      });
    } else {
      this.chatActions.push({
        chatId: this.selectedChatId,
        text,
      });
    }

    this.chatInputText = "";
    this.chatCursorIndex = 0;
    this.chatInputFocused = false;
  }

  // Проверяет, что системный указатель передан игровому canvas.
  private isPointerLocked(): boolean {
    return document.pointerLockElement === this.canvas;
  }

  // Открывает игровое меню, если правый клик попал в область вкладки.
  private openChatContextMenu(x: number, y: number): boolean {
    const chatState = this.visibleChatState;
    if (!chatState) {
      return false;
    }

    const tab = this.chatTabAtPoint(x, y);
    if (!tab) {
      return false;
    }

    this.selectedChatId = tab.chatId;
    this.chatContextMenu = {
      chatId: tab.chatId,
      communityTypeAcronym: tab.communityTypeAcronym,
      x,
      y,
    };
    return true;
  }

  // Выполняет пункт закрытия, если левый клик попал в меню.
  private closeChatContextMenuItem(x: number, y: number): boolean {
    if (!this.chatContextMenu) {
      return false;
    }

    const vh = window.innerHeight / 100;
    const width = 16 * vh;
    const height = 3.2 * vh;
    const inside = x >= this.chatContextMenu.x &&
      x <= this.chatContextMenu.x + width &&
      y >= this.chatContextMenu.y &&
      y <= this.chatContextMenu.y + height;
    if (!inside) {
      return false;
    }

    this.closeChatTab(this.chatContextMenu.chatId, this.chatContextMenu.communityTypeAcronym);
    return true;
  }

  // Находит вкладку по координатам с теми же размерами, что использует HUD.
  private chatTabAtPoint(x: number, y: number): ChatStateMessage["tabs"][number] | null {
    if (!this.visibleChatState || this.visibleChatState.tabs.length === 0) {
      return null;
    }

    const vh = window.innerHeight / 100;
    const panelLeft = 3 * vh;
    const panelTop = window.innerHeight / 2 - 17 * vh;
    const tabLeft = panelLeft + 1 * vh;
    const tabTop = panelTop + 1 * vh;
    const tabWidth = 14 * vh;
    const tabHeight = 3 * vh;
    const tabGap = 0.5 * vh;

    if (y < tabTop || y > tabTop + tabHeight || x < tabLeft) {
      return null;
    }

    const index = Math.floor((x - tabLeft) / (tabWidth + tabGap));
    const tabStart = tabLeft + index * (tabWidth + tabGap);
    if (x > tabStart + tabWidth) {
      return null;
    }

    return this.visibleChatState.tabs[index] ?? null;
  }
}
