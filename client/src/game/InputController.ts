import { INITIAL_ZOOM, clampZoom } from "../domain/camera";
import type { ChatSendMessage, ChatStateMessage, ClientInputState } from "../network/protocol";
import { GameUiRuntime } from "../ui-kit/runtime";
import { getScrollOffsetFromThumbTopPercent, getScrollbarThumbTopPercentFromCursor, startScrollbarDrag, type ScrollbarDragState } from "../ui-kit/scrollbar";
import { TextEditController } from "../ui-kit/textEdit";
import type { GameUiAction, GameUiControlState } from "../ui-kit/types";
import { isFreshKeyboardBinding, isFreshKeyDown, toShipInput, type InputBindingMap } from "./inputState";

export type ChatInputAction = Omit<ChatSendMessage, "type">;

export type ChatSelectAction = {
  // Чат, который игрок выбрал игровым указателем.
  chatId: number;
};

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
const chatTabsGapBottomVh = 0.8;
const chatTabHeightVh = 2.45;
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
  // Показывает, что локальный выбор вкладки ждет подтверждения сервера.
  private chatSelectionPending = false;
  // Очередь одноразовых команд отправки текста.
  private chatActions: ChatInputAction[] = [];
  // Очередь одноразовых команд выбора вкладки.
  private chatSelectActions: ChatSelectAction[] = [];
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
  // Состояние захвата ползунка истории.
  private chatScrollbarDrag: ScrollbarDragState | null = null;
  // Последний рассчитанный вид полосы прокрутки.
  private chatScrollState: ChatScrollState = { visible: false, thumbTopPercent: 0, thumbHeightPercent: 100, contentOffsetPx: 0, dragging: false };
  // Полный размер истории выбранной вкладки для расчета высоты прокрутки.
  private selectedChatMessageCount = 0;
  // Последний чат, для которого рассчитан локальный сдвиг истории.
  private scrolledChatId = 0;
  // Последнее сообщение выбранной вкладки для определения пополнения истории.
  private selectedChatLastMessageId = 0;
  // Общий runtime интерактивных игровых контролов HUD.
  private readonly uiRuntime = new GameUiRuntime();
  // Очередь действий общего игрового UI.
  private uiActions: GameUiAction[] = [];
  // Показывает отладочное окно с примерами всех контролов.
  private uiKitShowcaseVisible = false;
  // Показывает окно настроек игрока.
  private settingsVisible = false;
  // Накопленная прокрутка окна настроек игровым колесом.
  private settingsWheelDeltaY = 0;
  // Текущие системные привязки действий, полученные из настроек аккаунта.
  private inputBindings: InputBindingMap = {};
  // Нативный движок редактирования строки чата.
  private readonly chatEdit = new TextEditController({ id: "chat-input", mode: "singleLine" });
  // Показывает, что игровой курсор сейчас выделяет текст чата мышью.
  private chatEditDragActive = false;
  // Позиция начала выделения текстового поля игровым курсором.
  private chatEditDragAnchorIndex = 0;

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
        this.chatEdit.blur();
        this.keys[event.code] = true;
        return;
      }
      if (isFreshKeyDown(event.code, Boolean(this.keys[event.code]), "F10") || isFreshKeyboardBinding(event.code, Boolean(this.keys[event.code]), this.inputBindings.ToggleSettingsWindow)) {
        this.settingsVisible = !this.settingsVisible;
        this.keys[event.code] = true;
        event.preventDefault();
        return;
      }
      if (isFreshKeyDown(event.code, Boolean(this.keys[event.code]), "F9") || isFreshKeyboardBinding(event.code, Boolean(this.keys[event.code]), this.inputBindings.ToggleUiKitShowcase)) {
        this.uiKitShowcaseVisible = !this.uiKitShowcaseVisible;
        this.keys[event.code] = true;
        event.preventDefault();
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
      if (isFreshKeyboardBinding(event.code, Boolean(this.keys[event.code]), this.inputBindings.RandomShipChange) || isFreshKeyDown(event.code, Boolean(this.keys[event.code]), "Backslash")) {
        this.randomShipChangeRequested = true;
      }
      if (isFreshKeyboardBinding(event.code, Boolean(this.keys[event.code]), this.inputBindings.ToggleBodyPolygonDebug) || isFreshKeyDown(event.code, Boolean(this.keys[event.code]), "KeyO")) {
        this.bodyPolygonDebugToggleRequested = true;
      }
      if (isFreshKeyboardBinding(event.code, Boolean(this.keys[event.code]), this.inputBindings.ToggleAnchor) || isFreshKeyDown(event.code, Boolean(this.keys[event.code]), "KeyP")) {
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
        this.enqueueUiAction(this.uiRuntime.pointerMove(this.cursorX, this.cursorY));
        if (this.chatEditDragActive) {
          this.updateChatEditSelectionFromCursor();
        }
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
        if (this.settingsVisible) {
          this.settingsWheelDeltaY += event.deltaY;
          event.preventDefault();
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
      const uiTarget = this.uiRuntime.hitTest(this.cursorX, this.cursorY);
      this.enqueueUiAction(this.uiRuntime.pointerDown(this.cursorX, this.cursorY, event.button));
      if (this.isDropdownOutsideBlocker(uiTarget)) {
        event.preventDefault();
        return;
      }
      if (this.startChatEditPointerSelection(event.detail)) {
        event.preventDefault();
        return;
      }
      if (this.closeChatContextMenuItem(this.cursorX, this.cursorY) || this.closeChatContextMenuItem(event.clientX, event.clientY)) {
        event.preventDefault();
        return;
      }
      if (this.startChatScrollbarDrag()) {
        event.preventDefault();
        return;
      }
      if (this.selectChatTabAtCursor()) {
        event.preventDefault();
        return;
      }
      if (this.closeUiModeFromSpaceClick()) {
        event.preventDefault();
        return;
      }
      this.chatContextMenu = null;
    });

    window.addEventListener("mouseup", (event) => {
      if (event.button === 0) {
        this.enqueueUiAction(this.uiRuntime.pointerUp(this.cursorX, this.cursorY, event.button));
        this.chatEditDragActive = false;
        this.chatScrollbarDragActive = false;
        this.chatScrollbarDrag = null;
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

    this.chatEdit.element().addEventListener("input", () => this.syncChatInputFromNativeEdit());
    this.chatEdit.element().addEventListener("select", () => this.syncChatInputFromNativeEdit());
    this.chatEdit.element().addEventListener("keyup", () => this.syncChatInputFromNativeEdit());
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
      .filter((tab) => tab.communityTypeAcronym === "Server" || !this.closedDuoChatIds.has(tab.chatId) || (tab.unreadCount ?? 0) > 0)
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

    if (this.chatSelectionPending && chatState.selectedChatId === this.selectedChatId) {
      this.chatSelectionPending = false;
    }
    if (!this.chatSelectionPending && chatState.selectedChatId && !this.closedDuoChatIds.has(chatState.selectedChatId)) {
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

  // Возвращает позицию каретки внутри локальной строки.
  getChatCursorIndex(): number {
    return this.chatCursorIndex;
  }

  // Возвращает начало выделения в строке чата.
  getChatSelectionStart(): number {
    return this.chatEdit.snapshot().selectionStart;
  }

  // Возвращает конец выделения в строке чата.
  getChatSelectionEnd(): number {
    return this.chatEdit.snapshot().selectionEnd;
  }

  // Возвращает состояние фокуса, чтобы HUD мог показать каретку.
  isChatInputFocused(): boolean {
    return this.chatInputFocused;
  }

  // Выдает одну готовую команду отправки текста.
  consumeChatAction(): ChatInputAction | null {
    return this.chatActions.shift() ?? null;
  }

  // Выдает одну готовую команду выбора вкладки.
  consumeChatSelectAction(): ChatSelectAction | null {
    return this.chatSelectActions.shift() ?? null;
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

  // Обновляет registry общего UI runtime из актуально отрисованных HUD-контролов.
  updateGameUiControls(controls: GameUiControlState[]): void {
    this.uiRuntime.updateControls(controls);
  }

  // Возвращает очередное действие общего UI, если оно было создано игровым курсором.
  consumeGameUiAction(): GameUiAction | null {
    return this.uiActions.shift() ?? null;
  }

  // Возвращает видимость отладочной панели UI Kit.
  isUiKitShowcaseVisible(): boolean {
    return this.uiKitShowcaseVisible;
  }

  // Возвращает видимость окна настроек.
  isSettingsVisible(): boolean {
    return this.settingsVisible;
  }

  // Закрывает окно настроек после внешнего UI-действия.
  closeSettings(): void {
    this.settingsVisible = false;
  }

  // Возвращает прокрутку окна настроек и сразу начинает новое накопление.
  consumeSettingsWheelDeltaY(): number {
    const delta = this.settingsWheelDeltaY;
    this.settingsWheelDeltaY = 0;
    return delta;
  }

  // Обновляет привязки, используемые фактическим вводом.
  updateInputBindings(bindings: InputBindingMap): void {
    this.inputBindings = bindings;
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
      this.chatEdit.blur();
      this.chatContextMenu = null;
      this.chatScrollbarDragActive = false;
      this.chatEditDragActive = false;
      this.chatScrollbarDrag = null;
    }
    const canControlShip = isPointerLocked && !this.isGameCursorVisible();
    const input = toShipInput(canControlShip, this.keys, this.mouseDeltaX, this.inputBindings);
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
        this.chatEdit.focus(this.chatInputText, this.chatCursorIndex, this.chatCursorIndex);
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

    if (event.target === this.chatEdit.element()) {
      if (event.code === "Escape") {
        this.clearChatInput();
        event.preventDefault();
        return true;
      }
      window.setTimeout(() => this.syncChatInputFromNativeEdit(), 0);
      return true;
    }

    if (event.code === "Backspace") {
      if (this.chatCursorIndex > 0) {
        this.chatInputText = `${this.chatInputText.slice(0, this.chatCursorIndex - 1)}${this.chatInputText.slice(this.chatCursorIndex)}`;
        this.chatCursorIndex -= 1;
        this.chatEdit.focus(this.chatInputText, this.chatCursorIndex, this.chatCursorIndex);
      }
      event.preventDefault();
      return true;
    }

    if (event.code === "Delete") {
      if (this.chatCursorIndex < this.chatInputText.length) {
        this.chatInputText = `${this.chatInputText.slice(0, this.chatCursorIndex)}${this.chatInputText.slice(this.chatCursorIndex + 1)}`;
        this.chatEdit.focus(this.chatInputText, this.chatCursorIndex, this.chatCursorIndex);
      }
      event.preventDefault();
      return true;
    }

    if (event.code === "ArrowLeft") {
      this.chatCursorIndex = Math.max(0, this.chatCursorIndex - 1);
      this.chatEdit.focus(this.chatInputText, this.chatCursorIndex, this.chatCursorIndex);
      event.preventDefault();
      return true;
    }

    if (event.code === "ArrowRight") {
      this.chatCursorIndex = Math.min(this.chatInputText.length, this.chatCursorIndex + 1);
      this.chatEdit.focus(this.chatInputText, this.chatCursorIndex, this.chatCursorIndex);
      event.preventDefault();
      return true;
    }

    if (event.code === "Home") {
      this.chatCursorIndex = 0;
      this.chatEdit.focus(this.chatInputText, this.chatCursorIndex, this.chatCursorIndex);
      event.preventDefault();
      return true;
    }

    if (event.code === "End") {
      this.chatCursorIndex = this.chatInputText.length;
      this.chatEdit.focus(this.chatInputText, this.chatCursorIndex, this.chatCursorIndex);
      event.preventDefault();
      return true;
    }

    if (event.key.length === 1 && !event.ctrlKey && !event.metaKey && !event.altKey) {
      this.chatInputText = `${this.chatInputText.slice(0, this.chatCursorIndex)}${event.key}${this.chatInputText.slice(this.chatCursorIndex)}`;
      this.chatCursorIndex += event.key.length;
      this.chatEdit.focus(this.chatInputText, this.chatCursorIndex, this.chatCursorIndex);
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
      this.chatEdit.blur();
      return;
    }

    const directMessage = text.match(/^@(?:"([^"]+)"|([^\s]+))\s+([\s\S]+)$/);
    if (directMessage) {
      this.chatActions.push({
        targetNickname: directMessage[1] ?? directMessage[2],
        text: directMessage[3].trim(),
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
    this.chatEdit.blur();
  }

  // Проверяет, что системный указатель передан игровому canvas.
  private isPointerLocked(): boolean {
    return document.pointerLockElement === this.canvas;
  }

  // Показывает, что игровой указатель нужен для текущего UI-взаимодействия.
  private isGameCursorVisible(): boolean {
    return this.isPointerLocked() && (this.chatInputFocused || this.chatContextMenu !== null || this.uiKitShowcaseVisible || this.settingsVisible);
  }

  // Сохраняет действие общего runtime до обработки сценой.
  private enqueueUiAction(action: GameUiAction | null): void {
    if (action) {
      this.uiActions.push(action);
    }
  }

  // Узнает слой, который должен только закрыть раскрытый список и не запускать нижний UI.
  private isDropdownOutsideBlocker(control: GameUiControlState | null): boolean {
    return control?.id.endsWith("-outside-blocker") ?? false;
  }

  // Очищает локальную строку и закрывает native-редактор.
  private clearChatInput(): void {
    this.chatInputText = "";
    this.chatCursorIndex = 0;
    this.chatInputFocused = false;
    this.chatEdit.blur();
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

  // Выбирает вкладку под игровым указателем и ставит сетевое подтверждение чтения в очередь.
  private selectChatTabAtCursor(): boolean {
    if (!this.isGameCursorVisible()) {
      return false;
    }

    const tab = this.chatTabAtPoint(this.cursorX, this.cursorY);
    if (!tab) {
      return false;
    }

    this.selectedChatId = tab.chatId;
    this.chatSelectionPending = true;
    this.chatContextMenu = null;
    this.chatSelectActions.push({ chatId: tab.chatId });
    return true;
  }

  // Возвращает ввод корабля кликом по космосу вне панелей и меню.
  private closeUiModeFromSpaceClick(): boolean {
    if (!this.isGameCursorVisible()) {
      return false;
    }
    if (this.isCursorOverAnyHudPanel() || this.chatContextMenu !== null) {
      return false;
    }

    this.chatInputFocused = false;
    this.chatInputText = "";
    this.chatCursorIndex = 0;
    this.chatContextMenu = null;
    this.chatEdit.blur();
    this.uiKitShowcaseVisible = false;
    this.settingsVisible = false;
    return true;
  }

  // Начинает постановку каретки или выделение текста игровым курсором.
  private startChatEditPointerSelection(clickCount: number): boolean {
    if (!this.chatInputFocused || !this.isGameCursorVisible() || !this.isCursorOverChatInput()) {
      return false;
    }

    const index = this.chatTextIndexAtCursor();
    if (clickCount >= 2) {
      this.chatEdit.focus(this.chatInputText, index, index);
      this.chatEdit.selectWordAt(index);
      this.syncChatInputFromNativeEdit();
      return true;
    }

    this.chatEditDragActive = true;
    this.chatEditDragAnchorIndex = index;
    this.chatEdit.focus(this.chatInputText, index, index);
    this.syncChatInputFromNativeEdit();
    return true;
  }

  // Обновляет выделение при перетаскивании по строке ввода.
  private updateChatEditSelectionFromCursor(): void {
    const index = this.chatTextIndexAtCursor();
    this.chatEdit.focus(this.chatInputText, this.chatEditDragAnchorIndex, index);
    this.syncChatInputFromNativeEdit();
  }

  // Проверяет попадание игрового курсора в визуальную строку ввода.
  private isCursorOverChatInput(): boolean {
    const rect = document.getElementById("chat-input")?.getBoundingClientRect();
    return Boolean(rect &&
      this.cursorX >= rect.left &&
      this.cursorX <= rect.right &&
      this.cursorY >= rect.top &&
      this.cursorY <= rect.bottom);
  }

  // Находит ближайшую позицию текста в строке чата по координате игрового курсора.
  private chatTextIndexAtCursor(): number {
    const viewport = document.querySelector<HTMLElement>("#chat-input .chat-input__viewport");
    const text = document.querySelector<HTMLElement>("#chat-input .chat-input__text");
    const measure = document.querySelector<HTMLElement>("#chat-input .chat-input__measure");
    const viewportRect = viewport?.getBoundingClientRect();
    if (!viewportRect || this.chatInputText.length === 0) {
      return this.chatInputText.length;
    }

    const textWidth = measure?.getBoundingClientRect().width ?? this.chatInputText.length * 8;
    const charWidth = Math.max(1, textWidth / Math.max(1, this.chatInputText.length));
    const transform = text?.style.transform ?? "";
    const offset = Number(transform.match(/translateX\((-?\d+(?:\.\d+)?)px\)/)?.[1] ?? 0);
    const localX = this.cursorX - viewportRect.left - offset;
    return clamp(Math.round(localX / charWidth), 0, this.chatInputText.length);
  }

  private syncChatInputFromNativeEdit(): void {
    const snapshot = this.chatEdit.snapshot();
    this.chatInputText = snapshot.text;
    this.chatCursorIndex = snapshot.selectionEnd;
  }

  // Находит вкладку по координатам с теми же размерами, что использует HUD.
  private chatTabAtPoint(x: number, y: number): ChatStateMessage["tabs"][number] | null {
    if (!this.visibleChatState || this.visibleChatState.tabs.length === 0) {
      return null;
    }

    const renderedTab = this.renderedChatTabAtPoint(x, y);
    if (renderedTab !== undefined) {
      return renderedTab;
    }

    const vh = window.innerHeight / 100;
    const panelLeft = hudEdgeVh * vh;
    const panelTop = window.innerHeight / 2 - chatPanelHalfHeightVh * vh;
    const tabLeft = panelLeft + chatPanelPaddingVh * vh;
    const tabTop = panelTop + (chatPanelPaddingVh + chatMessagesHeightVh + chatTabsGapBottomVh) * vh;
    const tabWidth = 14 * vh;
    const tabHeight = chatTabHeightVh * vh;
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

  // Использует фактические DOM-границы, потому что вкладки чата имеют ширину по содержимому.
  private renderedChatTabAtPoint(x: number, y: number): ChatStateMessage["tabs"][number] | null | undefined {
    if (!this.visibleChatState) {
      return undefined;
    }

    let hasRenderedTabs = false;
    for (const tab of this.visibleChatState.tabs) {
      const element = document.getElementById(`chat-tab-${tab.chatId}`);
      const rect = element?.getBoundingClientRect();
      if (!rect || rect.width <= 0 || rect.height <= 0) {
        continue;
      }
      hasRenderedTabs = true;
      if (x >= rect.left && x <= rect.right && y >= rect.top && y <= rect.bottom) {
        return tab;
      }
    }

    return hasRenderedTabs ? null : undefined;
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

    const viewportHeightPx = this.chatMessagesViewportHeightPx();
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

    const track = this.chatScrollbarTrackRect();
    this.chatScrollbarDrag = startScrollbarDrag({
      top: track.top,
      height: track.height,
      thumbTopPercent: this.chatScrollState.thumbTopPercent,
      thumbHeightPercent: this.chatScrollState.thumbHeightPercent,
    }, this.cursorY);
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

    if (!this.chatScrollbarDrag) {
      this.chatScrollOffsetPx = 0;
      return;
    }

    const track = this.chatScrollbarTrackRect();
    const thumbTopPercent = getScrollbarThumbTopPercentFromCursor({
      top: track.top,
      height: track.height,
      thumbHeightPercent: this.chatScrollState.thumbHeightPercent,
      drag: this.chatScrollbarDrag,
    }, this.cursorY);
    this.chatScrollOffsetPx = getScrollOffsetFromThumbTopPercent({
      thumbTopPercent,
      thumbHeightPercent: this.chatScrollState.thumbHeightPercent,
      maxOffsetPx,
      reverse: true,
    });
  }

  // Возвращает максимальный локальный сдвиг истории выбранной вкладки.
  private selectedChatMaxScrollOffset(): number {
    return Math.max(0, this.selectedChatContentHeightPx() - this.chatMessagesViewportHeightPx());
  }

  // Возвращает видимую высоту истории без внутренних отступов блока сообщений.
  private chatMessagesViewportHeightPx(): number {
    const vh = window.innerHeight / 100;
    const rect = document.querySelector<HTMLElement>(".chat-messages")?.getBoundingClientRect();
    const height = (rect?.height && rect.height > 0) ? rect.height : chatMessagesHeightVh * vh;
    return Math.max(0, height - chatScrollbarVerticalInsetVh * vh);
  }

  // Возвращает фактическую высоту истории, а до отрисовки использует оценку по числу сообщений.
  private selectedChatContentHeightPx(): number {
    const measuredHeight = document.querySelector<HTMLElement>(".chat-messages__content")?.getBoundingClientRect().height ?? 0;
    const estimatedHeight = this.selectedChatMessageCount * chatEstimatedMessageHeightVh * window.innerHeight / 100;
    return Math.max(measuredHeight, estimatedHeight);
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

  // Проверяет попадание игрового указателя в любую постоянную HUD-панель.
  private isCursorOverAnyHudPanel(): boolean {
    if (this.uiRuntime.hitTest(this.cursorX, this.cursorY)) {
      return true;
    }
    for (const panel of document.querySelectorAll<HTMLElement>(".hud-panel")) {
      const rect = panel.getBoundingClientRect();
      if (rect.width <= 0 || rect.height <= 0) {
        continue;
      }
      if (this.cursorX >= rect.left &&
        this.cursorX <= rect.right &&
        this.cursorY >= rect.top &&
        this.cursorY <= rect.bottom) {
        return true;
      }
    }
    return this.isCursorOverChatPanel();
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

  // Возвращает экранную область блока сообщений.
  private chatMessagesRect(): { left: number; top: number; width: number; height: number } {
    const vh = window.innerHeight / 100;
    const panelLeft = hudEdgeVh * vh;
    const panelTop = window.innerHeight / 2 - chatPanelHalfHeightVh * vh;
    return {
      left: panelLeft + chatPanelPaddingVh * vh,
      top: panelTop + chatPanelPaddingVh * vh,
      width: (chatPanelWidthVh - chatPanelPaddingVh * 2) * vh,
      height: chatMessagesHeightVh * vh,
    };
  }
}

const clamp = (value: number, min: number, max: number): number => Math.min(max, Math.max(min, value));
