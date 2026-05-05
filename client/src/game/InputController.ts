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

export type GameCursorState = {
  // Показывает, что игровой указатель сейчас должен быть видимым.
  visible: boolean;
  // Горизонтальная координата указателя в пикселях окна.
  x: number;
  // Вертикальная координата указателя в пикселях окна.
  y: number;
};

export type ChatScrollState = {
  // Показывает, что у выбранной вкладки есть скрытая история.
  visible: boolean;
  // Верх ползунка в процентах высоты полосы.
  thumbTopPercent: number;
  // Высота ползунка в процентах высоты полосы.
  thumbHeightPercent: number;
  // Плавное смещение списка сообщений в пикселях.
  contentOffsetPx: number;
  // Показывает, что игрок сейчас держит ползунок мышью.
  dragging: boolean;
};

const hudEdgeVh = 1;
const chatPanelWidthVh = 48;
const chatPanelHalfHeightVh = 17;
const chatPanelPaddingVh = 1;
const chatTabsHeightVh = 3;
const chatTabsGapBottomVh = 0.8;
const chatMessagesHeightVh = 24;
const chatMessagesBorderVh = 0.14;
const chatScrollbarWidthVh = 1.1;
const chatScrollbarRightInsetVh = 0.25;
const chatScrollbarTopInsetVh = 0.8;
const chatScrollbarVerticalInsetVh = 1.6;
const chatEstimatedMessageHeightVh = 1.75;
const chatWheelPixels = 42;

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
  // Горизонтальная позиция игрового указателя.
  private cursorX = window.innerWidth / 2;
  // Вертикальная позиция игрового указателя.
  private cursorY = window.innerHeight / 2;
  // Локальный сдвиг истории выбранного чата от нижнего края.
  private chatScrollOffsetPx = 0;
  // Показывает, что левая кнопка двигает ползунок истории.
  private chatScrollbarDragActive = false;
  // Вертикальное расстояние от указателя до верха ползунка при начале перетаскивания.
  private chatScrollbarDragOffsetPx = 0;
  // Последний рассчитанный вид полосы прокрутки.
  private chatScrollState: ChatScrollState = { visible: false, thumbTopPercent: 0, thumbHeightPercent: 100, contentOffsetPx: 0, dragging: false };
  // Полный размер истории выбранной вкладки для расчета высоты прокрутки.
  private selectedChatMessageCount = 0;
  // Последний чат, для которого рассчитан локальный сдвиг истории.
  private scrolledChatId = 0;
  // Последнее сообщение выбранной вкладки для определения пополнения истории.
  private selectedChatLastMessageId = 0;

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
      if (this.isGameCursorVisible()) {
        this.keys[event.code] = true;
        event.preventDefault();
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
      if (this.isPointerLocked()) {
        if (!this.isGameCursorVisible()) {
          this.mouseDeltaX += event.movementX;
        }
        this.cursorX = clamp(this.cursorX + event.movementX, 0, Math.max(0, window.innerWidth - 1));
        this.cursorY = clamp(this.cursorY + event.movementY, 0, Math.max(0, window.innerHeight - 1));
        if (this.chatScrollbarDragActive) {
          this.updateChatScrollFromCursor();
        }
      }
    });

    // Колесо мыши меняет дискретный уровень зума, который затем переводится камерой в масштаб.
    window.addEventListener(
      "wheel",
      (event) => {
        if (!this.isPointerLocked()) {
          return;
        }
        if (this.chatInputFocused && this.isCursorOverChatPanel()) {
          this.scrollChatByWheel(event.deltaY);
          event.preventDefault();
          return;
        }
        if (this.isGameCursorVisible()) {
          event.preventDefault();
          return;
        }
        if (event.shiftKey) {
          this.pilotToolSelectionDelta += event.deltaY > 0 ? 1 : -1;
          return;
        }
        this.zoom = clampZoom(this.zoom + (event.deltaY > 0 ? -1 : 1));
      },
      { passive: false },
    );

    window.addEventListener("contextmenu", (event) => {
      if (this.isPointerLocked() && this.isGameCursorVisible() && (this.openChatContextMenu(this.cursorX, this.cursorY) || this.openChatContextMenu(event.clientX, event.clientY))) {
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
      if (this.startChatScrollbarDrag()) {
        event.preventDefault();
        return;
      }
      if (this.closeChatContextMenuItem(this.cursorX, this.cursorY) || this.closeChatContextMenuItem(event.clientX, event.clientY)) {
        event.preventDefault();
      } else {
        this.chatContextMenu = null;
      }
    });

    window.addEventListener("mouseup", (event) => {
      if (event.button === 0) {
        this.chatScrollbarDragActive = false;
        this.chatScrollbarDragOffsetPx = 0;
      }
    });

    window.addEventListener("resize", () => {
      this.cursorX = clamp(this.cursorX, 0, Math.max(0, window.innerWidth - 1));
      this.cursorY = clamp(this.cursorY, 0, Math.max(0, window.innerHeight - 1));
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

    const tabs = chatState.tabs
      .filter((tab) => tab.communityTypeAcronym === "Server" || !this.closedDuoChatIds.has(tab.chatId))
      .map((tab) => ({ ...tab, messages: [...tab.messages] }));
    if (tabs.length === 0) {
      this.selectedChatId = 0;
      this.selectedChatMessageCount = 0;
      this.scrolledChatId = 0;
      this.selectedChatLastMessageId = 0;
      this.visibleChatState = { ...chatState, tabs: [], selectedChatId: 0 };
      this.chatScrollState = { visible: false, thumbTopPercent: 0, thumbHeightPercent: 100, contentOffsetPx: 0, dragging: false };
      return this.visibleChatState;
    }

    if (chatState.selectedChatId && !this.closedDuoChatIds.has(chatState.selectedChatId)) {
      this.selectedChatId = chatState.selectedChatId;
    }
    if (!tabs.some((tab) => tab.chatId === this.selectedChatId)) {
      this.selectedChatId = tabs[0].chatId;
    }

    this.applyChatScroll(tabs);
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

  // Возвращает положение и видимость игрового указателя.
  getGameCursor(): GameCursorState {
    return {
      visible: this.isGameCursorVisible(),
      x: this.cursorX,
      y: this.cursorY,
    };
  }

  // Возвращает состояние полосы истории выбранной вкладки.
  getChatScrollState(): ChatScrollState {
    return this.chatScrollState;
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
      this.chatScrollbarDragActive = false;
      this.chatScrollbarDragOffsetPx = 0;
    }
    const canControlShip = isPointerLocked && !this.isGameCursorVisible();
    const input = toShipInput(canControlShip, this.keys, this.mouseDeltaX);
    input.toggleAnchor = canControlShip && this.anchorToggleRequested;
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
      this.chatInputText = "";
      this.chatCursorIndex = 0;
      this.chatInputFocused = false;
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

  // Показывает, что игровой указатель нужен для текущего UI-взаимодействия.
  private isGameCursorVisible(): boolean {
    return this.isPointerLocked() && (this.chatInputFocused || this.chatContextMenu !== null);
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
    const panelLeft = hudEdgeVh * vh;
    const panelTop = window.innerHeight / 2 - chatPanelHalfHeightVh * vh;
    const tabLeft = panelLeft + chatPanelPaddingVh * vh;
    const tabTop = panelTop + chatPanelPaddingVh * vh;
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

  // Применяет локальный сдвиг истории к выбранной вкладке.
  private applyChatScroll(tabs: ChatStateMessage["tabs"]): void {
    const selectedTab = tabs.find((tab) => tab.chatId === this.selectedChatId);
    if (!selectedTab) {
      this.scrolledChatId = 0;
      this.selectedChatLastMessageId = 0;
      this.chatScrollState = { visible: false, thumbTopPercent: 0, thumbHeightPercent: 100, contentOffsetPx: 0, dragging: this.chatScrollbarDragActive };
      return;
    }

    const total = selectedTab.messages.length;
    const lastMessageId = selectedTab.messages.at(-1)?.id ?? 0;
    const chatChanged = this.scrolledChatId !== selectedTab.chatId;
    const receivedNewMessage = !chatChanged && lastMessageId > this.selectedChatLastMessageId;
    if (chatChanged || (receivedNewMessage && !this.chatScrollbarDragActive)) {
      this.chatScrollOffsetPx = 0;
    }
    this.scrolledChatId = selectedTab.chatId;
    this.selectedChatLastMessageId = lastMessageId;
    this.selectedChatMessageCount = total;
    const maxOffsetPx = this.selectedChatMaxScrollOffset();
    this.chatScrollOffsetPx = clamp(this.chatScrollOffsetPx, 0, maxOffsetPx);

    if (maxOffsetPx <= 0) {
      this.chatScrollState = { visible: false, thumbTopPercent: 0, thumbHeightPercent: 100, contentOffsetPx: 0, dragging: this.chatScrollbarDragActive };
      return;
    }

    const viewportHeightPx = chatMessagesHeightVh * window.innerHeight / 100;
    const contentHeightPx = this.selectedChatContentHeightPx();
    const thumbHeightPercent = Math.max(14, (viewportHeightPx / contentHeightPx) * 100);
    const freeTrackPercent = 100 - thumbHeightPercent;
    const thumbTopPercent = freeTrackPercent * (1 - this.chatScrollOffsetPx / maxOffsetPx);
    this.chatScrollState = { visible: true, thumbTopPercent, thumbHeightPercent, contentOffsetPx: this.chatScrollOffsetPx, dragging: this.chatScrollbarDragActive };
  }

  // Прокручивает историю под игровым указателем колесом мыши.
  private scrollChatByWheel(deltaY: number): void {
    const maxOffsetPx = this.selectedChatMaxScrollOffset();
    if (maxOffsetPx <= 0) {
      this.chatScrollOffsetPx = 0;
      return;
    }
    this.chatScrollOffsetPx = clamp(this.chatScrollOffsetPx + (deltaY < 0 ? chatWheelPixels : -chatWheelPixels), 0, maxOffsetPx);
  }

  // Начинает перетаскивание, если указатель находится на видимой вертикальной полосе истории.
  private startChatScrollbarDrag(): boolean {
    if (!this.isGameCursorVisible() || !this.chatScrollState.visible || !this.isCursorOverChatScrollbarTrack()) {
      return false;
    }

    const thumb = this.chatScrollbarThumbRect();
    this.chatScrollbarDragOffsetPx = clamp(this.cursorY - thumb.top, 0, thumb.height);
    this.chatScrollbarDragActive = true;
    return true;
  }

  // Пересчитывает сдвиг истории так, чтобы ползунок шел вместе с указателем без смены точки хвата.
  private updateChatScrollFromCursor(): void {
    const maxOffsetPx = this.selectedChatMaxScrollOffset();
    if (maxOffsetPx <= 0) {
      this.chatScrollOffsetPx = 0;
      return;
    }

    const track = this.chatScrollbarTrackRect();
    const thumbHeightPx = track.height * this.chatScrollState.thumbHeightPercent / 100;
    const availableTrackPx = Math.max(0, track.height - thumbHeightPx);
    if (availableTrackPx <= 0) {
      this.chatScrollOffsetPx = 0;
      return;
    }

    const thumbTopPx = clamp(this.cursorY - this.chatScrollbarDragOffsetPx, track.top, track.top + availableTrackPx);
    const ratio = (thumbTopPx - track.top) / availableTrackPx;
    this.chatScrollOffsetPx = (1 - ratio) * maxOffsetPx;
  }

  // Возвращает максимальный локальный сдвиг истории выбранной вкладки.
  private selectedChatMaxScrollOffset(): number {
    return Math.max(0, this.selectedChatContentHeightPx() - chatMessagesHeightVh * window.innerHeight / 100);
  }

  // Оценивает высоту истории в пикселях по текущему размеру шрифта HUD.
  private selectedChatContentHeightPx(): number {
    return this.selectedChatMessageCount * chatEstimatedMessageHeightVh * window.innerHeight / 100;
  }

  // Проверяет попадание игрового указателя в панель чата.
  private isCursorOverChatPanel(): boolean {
    const vh = window.innerHeight / 100;
    const left = hudEdgeVh * vh;
    const top = window.innerHeight / 2 - chatPanelHalfHeightVh * vh;
    const width = chatPanelWidthVh * vh;
    const height = chatPanelHalfHeightVh * 2 * vh;
    return this.cursorX >= left && this.cursorX <= left + width && this.cursorY >= top && this.cursorY <= top + height;
  }

  // Проверяет попадание игрового указателя в видимую полосу истории.
  private isCursorOverChatScrollbarTrack(): boolean {
    const track = this.chatScrollbarTrackRect();
    return this.cursorX >= track.left &&
      this.cursorX <= track.left + track.width &&
      this.cursorY >= track.top &&
      this.cursorY <= track.top + track.height;
  }

  // Возвращает экранную область трека полосы истории.
  private chatScrollbarTrackRect(): { left: number; top: number; width: number; height: number } {
    const rect = this.chatMessagesRect();
    const vh = window.innerHeight / 100;
    const width = chatScrollbarWidthVh * vh;
    const border = chatMessagesBorderVh * vh;
    return {
      left: rect.left + rect.width - border - (chatScrollbarRightInsetVh + chatScrollbarWidthVh) * vh,
      top: rect.top + border + chatScrollbarTopInsetVh * vh,
      width,
      height: rect.height - border * 2 - chatScrollbarVerticalInsetVh * vh,
    };
  }

  // Возвращает экранную область ползунка истории.
  private chatScrollbarThumbRect(): { left: number; top: number; width: number; height: number } {
    const track = this.chatScrollbarTrackRect();
    const top = track.top + track.height * this.chatScrollState.thumbTopPercent / 100;
    const height = track.height * this.chatScrollState.thumbHeightPercent / 100;
    return {
      left: track.left,
      top,
      width: track.width,
      height,
    };
  }

  // Возвращает экранную область блока сообщений.
  private chatMessagesRect(): { left: number; top: number; width: number; height: number } {
    const vh = window.innerHeight / 100;
    const panelLeft = hudEdgeVh * vh;
    const panelTop = window.innerHeight / 2 - chatPanelHalfHeightVh * vh;
    return {
      left: panelLeft + chatPanelPaddingVh * vh,
      top: panelTop + (chatPanelPaddingVh + chatTabsHeightVh + chatTabsGapBottomVh) * vh,
      width: (chatPanelWidthVh - chatPanelPaddingVh * 2) * vh,
      height: chatMessagesHeightVh * vh,
    };
  }
}

const clamp = (value: number, min: number, max: number): number => Math.min(max, Math.max(min, value));
