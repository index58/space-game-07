import { INITIAL_ZOOM, clampZoom } from "../domain/camera";
import type { ChatSendMessage, ChatStateMessage, ClientInputState, CosmicObject } from "../network/protocol";
import { GameUiRuntime } from "../ui-kit/runtime";
import { getScrollOffsetFromThumbTopPercent, getScrollbarThumbTopPercentFromCursor, startScrollbarDrag, type ScrollbarDragState } from "../ui-kit/scrollbar";
import { TextEditController } from "../ui-kit/textEdit";
import type { GameUiAction, GameUiControlState, TextEditState } from "../ui-kit/types";
import { isFreshKeyboardBinding, isFreshKeyboardEventBinding, isFreshKeyDown, toShipInput, type InputBindingMap } from "./inputState";

export type ChatInputAction = Omit<ChatSendMessage, "type">;

export type DockingAction = "request" | "approve" | "reject" | "undock" | "landing";

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

export type GameUiWheelAction = {
  // Список или его полоса, над которыми игрок прокрутил колесо.
  controlId: string;
  // Вертикальный сдвиг колеса из браузерного события.
  deltaY: number;
};

type ChatInputDraft = {
  // Текст, набранный игроком в конкретной вкладке чата.
  text: string;
  // Позиция каретки в тексте этой вкладки.
  cursorIndex: number;
  // Начало выделенного диапазона в тексте этой вкладки.
  selectionStart: number;
  // Конец выделенного диапазона в тексте этой вкладки.
  selectionEnd: number;
};

const hudEdgeVh = 1;
const chatPanelWidthVh = 48;
const chatPanelHalfHeightVh = 17;
const chatPanelPaddingVh = 0.8;
const chatPanelGapVh = 0.7;
const chatInputHeightVh = 3.1;
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
  // Одноразовый тестовый запрос на присвоение объекта в фокусе.
  private focusedObjectOwnerClaimRequested = false;
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
  // Черновики ввода, сохраненные отдельно для каждой вкладки чата.
  private readonly chatInputDrafts = new Map<number, ChatInputDraft>();
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
  // Очередь прокрутки списков общего игрового UI.
  private uiWheelActions: GameUiWheelAction[] = [];
  // Очередь одноразовых команд стыковки.
  private dockingActions: DockingAction[] = [];
  // Показывает отладочное окно с примерами всех контролов.
  private uiKitShowcaseVisible = false;
  // Показывает окно настроек игрока.
  private settingsVisible = false;
  // Показывает окно панели управления объектом.
  private controlPanelVisible = false;
  // Накопленная прокрутка окна настроек игровым колесом.
  private settingsWheelDeltaY = 0;
  // Текущие системные привязки действий, полученные из настроек аккаунта.
  private inputBindings: InputBindingMap = {};
  // Нативный движок редактирования строки чата.
  private readonly chatEdit = new TextEditController({ id: "chat-input", mode: "singleLine" });
  // Нативный движок редактирования названия объекта в панели управления.
  private readonly controlPanelObjectTitleEdit = new TextEditController({ id: "control-panel-object-title-input", mode: "singleLine" });
  // Нативный движок редактирования названия группы оборудования.
  private readonly controlPanelEquipmentTitleEdit = new TextEditController({ id: "control-panel-equipment-title-input", mode: "singleLine" });
  // Нативный движок редактирования количества сливаемого топлива.
  private readonly controlPanelFuelDrainAmountEdit = new TextEditController({ id: "control-panel-fuel-drain-amount-input", mode: "singleLine" });
  // Объект, для которого сейчас хранится черновик панели управления.
  private controlPanelObjectId: number | null = null;
  // Черновик состояния включения объекта в панели управления.
  private controlPanelObjectEnabledDraft: boolean | null = null;
  // Черновик пользовательского названия объекта в панели управления.
  private controlPanelObjectTitleDraft: string | null = null;
  // Текст, который нужно отправить на сервер после завершения редактирования названия.
  private controlPanelObjectTitleCommit: string | null = null;
  // Группа оборудования, для которой сейчас хранится черновик названия.
  private controlPanelEquipmentTitleGroupId: number | null = null;
  // Черновик названия группы оборудования в панели управления.
  private controlPanelEquipmentTitleDraft: string | null = null;
  // Текст, который нужно отправить на сервер после завершения редактирования названия группы.
  private controlPanelEquipmentTitleCommit: string | null = null;
  // Показывает, что игровой курсор сейчас выделяет текст чата мышью.
  private chatEditDragActive = false;
  // Позиция начала выделения текстового поля игровым курсором.
  private chatEditDragAnchorIndex = 0;
  // Показывает, что игровой курсор сейчас выделяет название объекта мышью.
  private controlPanelObjectTitleEditDragActive = false;
  // Показывает, что игровой курсор сейчас выделяет название группы оборудования мышью.
  private controlPanelEquipmentTitleEditDragActive = false;
  // Позиция начала выделения названия объекта игровым курсором.
  private controlPanelObjectTitleEditDragAnchorIndex = 0;
  // Позиция начала выделения названия группы оборудования игровым курсором.
  private controlPanelEquipmentTitleEditDragAnchorIndex = 0;
  // Блокирует click после перетаскивания, чтобы выделение не схлопывалось на отпускании.
  private controlPanelObjectTitleSuppressClick = false;
  // Блокирует click после перетаскивания, чтобы выделение не схлопывалось на отпускании.
  private controlPanelEquipmentTitleSuppressClick = false;
  // Черновой текст количества сливаемого топлива.
  private controlPanelFuelDrainAmountText = "0";

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
        this.controlPanelObjectTitleEdit.blur();
        this.controlPanelEquipmentTitleEdit.blur();
        this.controlPanelFuelDrainAmountEdit.blur();
        this.keys[event.code] = true;
        return;
      }
      if (isFreshKeyDown(event.code, Boolean(this.keys[event.code]), "F10") || isFreshKeyboardBinding(event.code, Boolean(this.keys[event.code]), this.inputBindings.ToggleSettingsWindow)) {
        this.toggleSettingsModal();
        this.keys[event.code] = true;
        event.preventDefault();
        return;
      }
      if (isFreshKeyDown(event.code, Boolean(this.keys[event.code]), "F9") || isFreshKeyboardBinding(event.code, Boolean(this.keys[event.code]), this.inputBindings.ToggleUiKitShowcase)) {
        this.toggleUiKitShowcaseModal();
        this.keys[event.code] = true;
        event.preventDefault();
        return;
      }
      if (this.handleChatKeyDown(event)) {
        return;
      }
      if (this.handleControlPanelObjectTitleKeyDown(event)) {
        return;
      }
      if (this.handleControlPanelEquipmentTitleKeyDown(event)) {
        return;
      }
      if (this.handleControlPanelFuelDrainAmountKeyDown(event)) {
        return;
      }
      if (isFreshKeyDown(event.code, Boolean(this.keys[event.code]), "KeyI")) {
        this.toggleControlPanelModal();
        this.keys[event.code] = true;
        event.preventDefault();
        return;
      }
      if (this.isGameCursorVisible()) {
        this.keys[event.code] = true;
        event.preventDefault();
        return;
      }
      const dockingAction = this.getDockingActionFromKey(event, Boolean(this.keys[event.code]));
      if (dockingAction) {
        this.dockingActions.push(dockingAction);
        this.keys[event.code] = true;
        event.preventDefault();
        return;
      }
      if (isFreshKeyDown(event.code, Boolean(this.keys[event.code]), "Backslash") && event.altKey) {
        this.focusedObjectOwnerClaimRequested = true;
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
        if (this.controlPanelObjectTitleEditDragActive) {
          this.updateControlPanelObjectTitleEditSelectionFromCursor();
        }
        if (this.controlPanelEquipmentTitleEditDragActive) {
          this.updateControlPanelEquipmentTitleEditSelectionFromCursor();
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
          this.enqueueGameUiWheelAction(event.deltaY);
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
      this.enqueueUiAction(this.uiRuntime.pointerDown(this.cursorX, this.cursorY, event.button, uiActionModifiers(event)));
      if (this.isDropdownOutsideBlocker(uiTarget)) {
        event.preventDefault();
        return;
      }
      if (this.startChatEditPointerSelection(event.detail)) {
        event.preventDefault();
        return;
      }
      if (this.startControlPanelObjectTitleEditPointerSelection(event.detail)) {
        event.preventDefault();
        return;
      }
      if (this.startControlPanelEquipmentTitleEditPointerSelection(event.detail)) {
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
        this.enqueueUiAction(this.uiRuntime.pointerUp(this.cursorX, this.cursorY, event.button, uiActionModifiers(event)));
        this.chatEditDragActive = false;
        this.controlPanelObjectTitleEditDragActive = false;
        this.controlPanelEquipmentTitleEditDragActive = false;
        this.controlPanelObjectTitleSuppressClick = false;
        this.controlPanelEquipmentTitleSuppressClick = false;
        this.chatScrollbarDragActive = false;
        this.chatScrollbarDrag = null;
      }
    });

    window.addEventListener("resize", () => {
      this.cursorX = clamp(this.cursorX, 0, Math.max(0, window.innerWidth - 1));
      this.cursorY = clamp(this.cursorY, 0, Math.max(0, window.innerHeight - 1));
    });

    // Захват мыши включается только по клику и только когда клиент уже готов принимать управление.
    this.canvas.addEventListener("click", (event) => {
      if (this.isPointerLocked()) {
        return;
      }
      if (!this.canRequestPointerLock()) {
        return;
      }

      this.placeCursorAtSystemPointer(event.clientX, event.clientY);
      void this.canvas.requestPointerLock();
    });

    this.chatEdit.element().addEventListener("input", () => this.syncChatInputFromNativeEdit());
    this.chatEdit.element().addEventListener("select", () => this.syncChatInputFromNativeEdit());
    this.chatEdit.element().addEventListener("keyup", () => this.syncChatInputFromNativeEdit());
    this.controlPanelObjectTitleEdit.element().addEventListener("input", () => this.syncControlPanelObjectTitleFromNativeEdit());
    this.controlPanelObjectTitleEdit.element().addEventListener("select", () => this.syncControlPanelObjectTitleFromNativeEdit());
    this.controlPanelObjectTitleEdit.element().addEventListener("keyup", () => this.syncControlPanelObjectTitleFromNativeEdit());
    this.controlPanelObjectTitleEdit.element().addEventListener("keydown", (event) => this.handleControlPanelObjectTitleKeyDown(event));
    this.controlPanelEquipmentTitleEdit.element().addEventListener("input", () => this.syncControlPanelEquipmentTitleFromNativeEdit());
    this.controlPanelEquipmentTitleEdit.element().addEventListener("select", () => this.syncControlPanelEquipmentTitleFromNativeEdit());
    this.controlPanelEquipmentTitleEdit.element().addEventListener("keyup", () => this.syncControlPanelEquipmentTitleFromNativeEdit());
    this.controlPanelEquipmentTitleEdit.element().addEventListener("keydown", (event) => this.handleControlPanelEquipmentTitleKeyDown(event));
    this.controlPanelFuelDrainAmountEdit.element().addEventListener("input", () => this.syncControlPanelFuelDrainAmountFromNativeEdit());
    this.controlPanelFuelDrainAmountEdit.element().addEventListener("select", () => this.syncControlPanelFuelDrainAmountFromNativeEdit());
    this.controlPanelFuelDrainAmountEdit.element().addEventListener("keyup", () => this.syncControlPanelFuelDrainAmountFromNativeEdit());
    this.controlPanelFuelDrainAmountEdit.element().addEventListener("keydown", (event) => this.handleControlPanelFuelDrainAmountKeyDown(event));
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
      this.saveCurrentChatInputDraft();
      this.selectedChatId = 0;
      this.chatInputText = "";
      this.chatCursorIndex = 0;
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
      this.setSelectedChatId(chatState.selectedChatId);
    }
    if (!tabs.some((tab) => tab.chatId === this.selectedChatId)) {
      this.setSelectedChatId(tabs[0].chatId);
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

  // Синхронизирует черновик панели управления с текущим объектом без перезаписи уже введенных правок.
  syncControlPanelObject(object: CosmicObject | null): void {
    if (!object) {
      this.controlPanelObjectId = null;
      this.controlPanelObjectEnabledDraft = null;
      this.controlPanelObjectTitleDraft = null;
      this.controlPanelObjectTitleEdit.blur();
      return;
    }
    if (this.controlPanelObjectId !== object.ID) {
      this.controlPanelObjectId = object.ID;
      this.controlPanelObjectEnabledDraft = object.Enabled;
      this.controlPanelObjectTitleDraft = object.Title;
      this.controlPanelObjectTitleEdit.blur();
      return;
    }
    this.controlPanelObjectEnabledDraft = object.Enabled;
    if (!this.controlPanelObjectTitleEdit.snapshot().focused) {
      this.controlPanelObjectTitleDraft = object.Title;
    }
  }

  // Возвращает черновик переключателя панели управления или серверное значение до синхронизации.
  getControlPanelObjectEnabled(fallback: boolean): boolean {
    return this.controlPanelObjectEnabledDraft ?? fallback;
  }

  // Возвращает черновик названия объекта или серверное значение до синхронизации.
  getControlPanelObjectTitle(fallback: string): string {
    return this.controlPanelObjectTitleDraft ?? fallback;
  }

  // Возвращает состояние редактирования названия объекта для отрисовки UI Kit поля.
  getControlPanelObjectTitleEditState(fallback = ""): TextEditState {
    const snapshot = this.controlPanelObjectTitleEdit.snapshot();
    if (snapshot.focused) {
      return snapshot;
    }
    const text = this.getControlPanelObjectTitle(fallback);
    return {
      text,
      selectionStart: text.length,
      selectionEnd: text.length,
      selectionDirection: "none",
      scrollX: 0,
      scrollY: 0,
      focused: false,
    };
  }

  // Синхронизирует черновик названия выбранной группы оборудования без перезаписи активного ввода.
  syncControlPanelEquipmentTitle(groupId: number | null, title: string): void {
    if (!groupId) {
      this.controlPanelEquipmentTitleGroupId = null;
      this.controlPanelEquipmentTitleDraft = null;
      this.controlPanelEquipmentTitleEdit.blur();
      return;
    }
    if (this.controlPanelEquipmentTitleGroupId !== groupId) {
      this.controlPanelEquipmentTitleGroupId = groupId;
      this.controlPanelEquipmentTitleDraft = title;
      this.controlPanelEquipmentTitleEdit.blur();
      return;
    }
    if (!this.controlPanelEquipmentTitleEdit.snapshot().focused) {
      this.controlPanelEquipmentTitleDraft = title;
    }
  }

  // Возвращает черновик названия группы оборудования или серверное значение до синхронизации.
  getControlPanelEquipmentTitle(fallback: string): string {
    return this.controlPanelEquipmentTitleDraft ?? fallback;
  }

  // Возвращает состояние редактирования названия группы оборудования для отрисовки UI Kit поля.
  getControlPanelEquipmentTitleEditState(fallback = ""): TextEditState {
    const snapshot = this.controlPanelEquipmentTitleEdit.snapshot();
    if (snapshot.focused) {
      return snapshot;
    }
    const text = this.getControlPanelEquipmentTitle(fallback);
    return {
      text,
      selectionStart: text.length,
      selectionEnd: text.length,
      selectionDirection: "none",
      scrollX: 0,
      scrollY: 0,
      focused: false,
    };
  }

  // Задает количество слива топлива из внешнего состояния панели.
  setControlPanelFuelDrainAmount(value: number): void {
    this.controlPanelFuelDrainAmountText = formatFuelDrainAmount(value);
    this.controlPanelFuelDrainAmountEdit.blur();
  }

  // Возвращает числовое количество слива, набранное в поле.
  getControlPanelFuelDrainAmount(fallback = 0): number {
    const value = Number(this.controlPanelFuelDrainAmountText.replace(",", "."));
    return Number.isFinite(value) ? Math.max(0, value) : fallback;
  }

  // Возвращает состояние редактирования количества топлива для отрисовки поля.
  getControlPanelFuelDrainAmountEditState(): TextEditState {
    const snapshot = this.controlPanelFuelDrainAmountEdit.snapshot();
    if (snapshot.focused) {
      return snapshot;
    }
    const text = this.controlPanelFuelDrainAmountText;
    return {
      text,
      selectionStart: text.length,
      selectionEnd: text.length,
      selectionDirection: "none",
      scrollX: 0,
      scrollY: 0,
      focused: false,
    };
  }

  // Снимает фокус с поля количества слива топлива.
  blurControlPanelFuelDrainAmount(): void {
    this.controlPanelFuelDrainAmountEdit.blur();
  }

  // Обновляет registry общего UI runtime из актуально отрисованных HUD-контролов.
  updateGameUiControls(controls: GameUiControlState[]): void {
    this.uiRuntime.updateControls(controls);
  }

  // Возвращает контрол HUD, на который сейчас наведён игровой указатель.
  getHoveredGameUiControlId(): string | null {
    return this.uiRuntime.snapshot().hoveredControlId;
  }

  // Возвращает очередное действие общего UI, если оно было создано игровым курсором.
  consumeGameUiAction(): GameUiAction | null {
    return this.uiActions.shift() ?? null;
  }

  // Возвращает очередную прокрутку списка общего UI, если она была создана игровым колесом.
  consumeGameUiWheelAction(): GameUiWheelAction | null {
    return this.uiWheelActions.shift() ?? null;
  }

  // Возвращает очередную команду стыковки для отправки отдельным сетевым сообщением.
  consumeDockingAction(): DockingAction | null {
    return this.dockingActions.shift() ?? null;
  }

  // Преобразует настроенные или базовые сочетания клавиш в команды стыковки.
  private getDockingActionFromKey(event: KeyboardEvent, wasPressed: boolean): DockingAction | null {
    if (isFreshKeyboardEventBinding(event, wasPressed, this.inputBindings.DockingRequest)) {
      return "request";
    }
    if (isFreshKeyboardEventBinding(event, wasPressed, this.inputBindings.ApproveRequest) || isFreshKeyboardEventBinding(event, wasPressed, this.inputBindings.DockingApprove)) {
      return "approve";
    }
    if (isFreshKeyboardEventBinding(event, wasPressed, this.inputBindings.RejectRequest) || isFreshKeyboardEventBinding(event, wasPressed, this.inputBindings.DockingReject)) {
      return "reject";
    }
    if (isFreshKeyboardEventBinding(event, wasPressed, this.inputBindings.DockingUndock)) {
      return "undock";
    }
    if (isFreshKeyboardEventBinding(event, wasPressed, this.inputBindings.LandingBegin)) {
      return "landing";
    }
    return getDefaultDockingActionFromKey(event, wasPressed);
  }

  // Возвращает видимость отладочной панели UI Kit.
  isUiKitShowcaseVisible(): boolean {
    return this.uiKitShowcaseVisible;
  }

  // Возвращает видимость окна настроек.
  isSettingsVisible(): boolean {
    return this.settingsVisible;
  }

  // Возвращает видимость панели управления.
  isControlPanelVisible(): boolean {
    return this.controlPanelVisible;
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

  // Возвращает тестовую команду присвоения объекта один раз на одно нажатие.
  consumeFocusedObjectOwnerClaimRequest(): boolean {
    const requested = this.focusedObjectOwnerClaimRequested;
    this.focusedObjectOwnerClaimRequested = false;
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

  // Возвращает завершенное редактирование названия объекта и сразу очищает событие.
  consumeControlPanelObjectTitleCommit(): string | null {
    const title = this.controlPanelObjectTitleCommit;
    this.controlPanelObjectTitleCommit = null;
    return title;
  }

  // Возвращает завершенное редактирование названия группы оборудования и сразу очищает событие.
  consumeControlPanelEquipmentTitleCommit(): string | null {
    const title = this.controlPanelEquipmentTitleCommit;
    this.controlPanelEquipmentTitleCommit = null;
    return title;
  }

  // Отдает ввод за текущий кадр и сбрасывает накопленное движение мыши.
  consumeShipInput(): ClientInputState {
    // Захват указателя отдает относительное движение мыши; после кадра накопление сбрасывается.
    const isPointerLocked = this.isPointerLocked();
    if (!isPointerLocked) {
      this.chatInputFocused = false;
      this.chatEdit.blur();
      this.controlPanelObjectTitleEdit.blur();
      this.controlPanelEquipmentTitleEdit.blur();
      this.controlPanelFuelDrainAmountEdit.blur();
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

  // Отдает клавиатуру native-полю названия объекта, пока оно находится в фокусе.
  private handleControlPanelObjectTitleKeyDown(event: KeyboardEvent): boolean {
    if (!this.controlPanelObjectTitleEdit.snapshot().focused) {
      return false;
    }
    if (event.target !== this.controlPanelObjectTitleEdit.element()) {
      return false;
    }
    if (event.code === "Escape" || event.code === "Enter") {
      this.syncControlPanelObjectTitleFromNativeEdit();
      this.commitAndBlurControlPanelObjectTitle();
      event.preventDefault();
      return true;
    }

    window.setTimeout(() => this.syncControlPanelObjectTitleFromNativeEdit(), 0);
    return true;
  }

  // Отдает клавиатуру native-полю названия группы оборудования, пока оно находится в фокусе.
  private handleControlPanelEquipmentTitleKeyDown(event: KeyboardEvent): boolean {
    if (!this.controlPanelEquipmentTitleEdit.snapshot().focused) {
      return false;
    }
    if (event.target !== this.controlPanelEquipmentTitleEdit.element()) {
      return false;
    }
    if (event.code === "Escape" || event.code === "Enter") {
      this.syncControlPanelEquipmentTitleFromNativeEdit();
      this.commitAndBlurControlPanelEquipmentTitle();
      event.preventDefault();
      return true;
    }

    window.setTimeout(() => this.syncControlPanelEquipmentTitleFromNativeEdit(), 0);
    return true;
  }

  // Отдает клавиатуру native-полю количества слива топлива, пока оно находится в фокусе.
  private handleControlPanelFuelDrainAmountKeyDown(event: KeyboardEvent): boolean {
    if (!this.controlPanelFuelDrainAmountEdit.snapshot().focused) {
      return false;
    }
    if (event.target !== this.controlPanelFuelDrainAmountEdit.element()) {
      return false;
    }
    if (event.code === "Escape" || event.code === "Enter") {
      this.syncControlPanelFuelDrainAmountFromNativeEdit();
      this.controlPanelFuelDrainAmountEdit.blur();
      event.preventDefault();
      return true;
    }

    window.setTimeout(() => this.syncControlPanelFuelDrainAmountFromNativeEdit(), 0);
    return true;
  }

  // Превращает локальную строку в обычную или адресную сетевую команду.
  private queueChatAction(): void {
    const text = this.chatInputText.trim();
    if (text === "") {
      this.clearChatInput();
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

    this.clearChatInput();
  }

  // Проверяет, что системный указатель передан игровому canvas.
  private isPointerLocked(): boolean {
    return document.pointerLockElement === this.canvas;
  }

  // Ставит игровой указатель туда, где браузерный указатель был перед захватом.
  private placeCursorAtSystemPointer(x: number, y: number): void {
    this.cursorX = clamp(x, 0, Math.max(0, window.innerWidth - 1));
    this.cursorY = clamp(y, 0, Math.max(0, window.innerHeight - 1));
  }

  // Показывает, что игровой указатель нужен для текущего UI-взаимодействия.
  private isGameCursorVisible(): boolean {
    return this.isPointerLocked() && (this.chatInputFocused || this.chatContextMenu !== null || this.uiKitShowcaseVisible || this.settingsVisible || this.controlPanelVisible);
  }

  // Переключает окно настроек так, чтобы остальные модальные окна были закрыты.
  private toggleSettingsModal(): void {
    const nextVisible = !this.settingsVisible;
    this.closeModalWindows();
    this.settingsVisible = nextVisible;
  }

  // Переключает витрину UI Kit так, чтобы остальные модальные окна были закрыты.
  private toggleUiKitShowcaseModal(): void {
    const nextVisible = !this.uiKitShowcaseVisible;
    this.closeModalWindows();
    this.uiKitShowcaseVisible = nextVisible;
  }

  // Переключает панель управления так, чтобы остальные модальные окна были закрыты.
  private toggleControlPanelModal(): void {
    const nextVisible = !this.controlPanelVisible;
    this.closeModalWindows();
    this.controlPanelVisible = nextVisible;
  }

  // Закрывает все модальные окна, которыми управляет игровой ввод.
  private closeModalWindows(): void {
    this.settingsVisible = false;
    this.uiKitShowcaseVisible = false;
    this.controlPanelVisible = false;
    this.controlPanelObjectTitleEdit.blur();
    this.controlPanelEquipmentTitleEdit.blur();
    this.controlPanelFuelDrainAmountEdit.blur();
  }

  // Сохраняет действие общего runtime до обработки сценой.
  private enqueueUiAction(action: GameUiAction | null): void {
    if (action) {
      if (this.consumeInternalUiAction(action)) {
        return;
      }
      this.uiActions.push(action);
    }
  }

  // Сохраняет прокрутку только для списков, чтобы колесо не мешало управлению кораблем.
  private enqueueGameUiWheelAction(deltaY: number): void {
    const target = this.uiRuntime.hitTest(this.cursorX, this.cursorY);
    const controlId = target ? scrollableListIdFromControl(target) : null;
    if (controlId) {
      this.uiWheelActions.push({ controlId, deltaY });
    }
  }

  // Обрабатывает команды, которые принадлежат самому каркасу игрового UI.
  private consumeInternalUiAction(action: GameUiAction): boolean {
    if (action.type === "click" && action.controlId.endsWith("-close-button")) {
      this.closeModalWindows();
      return true;
    }
    if (action.type === "click" && action.controlId === "control-panel-object-enabled") {
      this.commitAndBlurControlPanelObjectTitle();
      this.commitAndBlurControlPanelEquipmentTitle();
      return false;
    }
    if (action.type === "click" && action.controlId === "control-panel-object-title-input") {
      if (this.controlPanelObjectTitleSuppressClick) {
        return true;
      }
      const text = this.getControlPanelObjectTitle("");
      const index = this.textInputIndexAtCursor("control-panel-object-title-input", text);
      this.controlPanelObjectTitleEdit.focus(text, index, index);
      this.syncControlPanelObjectTitleFromNativeEdit();
      return true;
    }
    if (action.type === "click" && action.controlId === "control-panel-equipment-title-input") {
      if (this.controlPanelEquipmentTitleSuppressClick) {
        return true;
      }
      const text = this.getControlPanelEquipmentTitle("");
      const index = this.textInputIndexAtCursor("control-panel-equipment-title-input", text);
      this.controlPanelEquipmentTitleEdit.focus(text, index, index);
      this.syncControlPanelEquipmentTitleFromNativeEdit();
      return true;
    }
    if (action.type === "click" && action.controlId === "control-panel-fuel-drain-amount-input") {
      this.controlPanelFuelDrainAmountEdit.focus(this.controlPanelFuelDrainAmountText, 0, this.controlPanelFuelDrainAmountText.length);
      this.syncControlPanelFuelDrainAmountFromNativeEdit();
      return true;
    }
    if (action.type === "click" && action.controlId.startsWith("control-panel-")) {
      this.commitAndBlurControlPanelObjectTitle();
      this.commitAndBlurControlPanelEquipmentTitle();
    }
    return false;
  }

  // Переносит данные native-поля в черновик панели управления.
  private syncControlPanelObjectTitleFromNativeEdit(): void {
    this.controlPanelObjectTitleDraft = this.controlPanelObjectTitleEdit.snapshot().text;
  }

  // Переносит данные native-поля названия группы оборудования в черновик панели управления.
  private syncControlPanelEquipmentTitleFromNativeEdit(): void {
    this.controlPanelEquipmentTitleDraft = this.controlPanelEquipmentTitleEdit.snapshot().text;
  }

  // Переносит данные native-поля количества слива в черновик панели.
  private syncControlPanelFuelDrainAmountFromNativeEdit(): void {
    this.controlPanelFuelDrainAmountText = this.controlPanelFuelDrainAmountEdit.snapshot().text;
  }

  // Завершает native-редактирование названия и запоминает изменение для сцены.
  private commitAndBlurControlPanelObjectTitle(): void {
    const snapshot = this.controlPanelObjectTitleEdit.snapshot();
    if (snapshot.focused) {
      this.controlPanelObjectTitleDraft = snapshot.text;
      this.controlPanelObjectTitleCommit = snapshot.text;
    }
    this.controlPanelObjectTitleEdit.blur();
  }

  // Завершает native-редактирование названия группы оборудования и запоминает изменение для сцены.
  private commitAndBlurControlPanelEquipmentTitle(): void {
    const snapshot = this.controlPanelEquipmentTitleEdit.snapshot();
    if (snapshot.focused) {
      this.controlPanelEquipmentTitleDraft = snapshot.text;
      this.controlPanelEquipmentTitleCommit = snapshot.text;
    }
    this.controlPanelEquipmentTitleEdit.blur();
  }

  // Узнает слой, который должен только закрыть раскрытый список и не запускать нижний UI.
  private isDropdownOutsideBlocker(control: GameUiControlState | null): boolean {
    return control?.id.endsWith("-outside-blocker") ?? false;
  }

  // Очищает локальную строку и закрывает native-редактор.
  private clearChatInput(): void {
    this.chatInputText = "";
    this.chatCursorIndex = 0;
    if (this.selectedChatId > 0) {
      this.chatInputDrafts.delete(this.selectedChatId);
    }
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

    this.setSelectedChatId(tab.chatId);
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

    this.setSelectedChatId(tab.chatId);
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
    this.clearChatInput();
    this.chatContextMenu = null;
    this.uiKitShowcaseVisible = false;
    this.settingsVisible = false;
    this.controlPanelVisible = false;
    this.controlPanelObjectTitleEdit.blur();
    this.controlPanelEquipmentTitleEdit.blur();
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

  // Начинает постановку каретки или выделение названия объекта игровым курсором.
  private startControlPanelObjectTitleEditPointerSelection(clickCount: number): boolean {
    if (!this.isGameCursorVisible() || !this.isCursorOverTextInput("control-panel-object-title-input")) {
      return false;
    }

    const text = this.getControlPanelObjectTitle("");
    const index = this.textInputIndexAtCursor("control-panel-object-title-input", text);
    if (clickCount >= 2) {
      this.controlPanelObjectTitleEdit.focus(text, index, index);
      this.controlPanelObjectTitleEdit.selectWordAt(index);
      this.syncControlPanelObjectTitleFromNativeEdit();
      this.controlPanelObjectTitleSuppressClick = true;
      return true;
    }

    this.controlPanelObjectTitleEditDragActive = true;
    this.controlPanelObjectTitleEditDragAnchorIndex = index;
    this.controlPanelObjectTitleSuppressClick = false;
    this.controlPanelObjectTitleEdit.focus(text, index, index);
    this.syncControlPanelObjectTitleFromNativeEdit();
    return true;
  }

  // Начинает постановку каретки или выделение названия группы оборудования игровым курсором.
  private startControlPanelEquipmentTitleEditPointerSelection(clickCount: number): boolean {
    if (!this.isGameCursorVisible() || !this.isCursorOverTextInput("control-panel-equipment-title-input")) {
      return false;
    }

    const text = this.getControlPanelEquipmentTitle("");
    const index = this.textInputIndexAtCursor("control-panel-equipment-title-input", text);
    if (clickCount >= 2) {
      this.controlPanelEquipmentTitleEdit.focus(text, index, index);
      this.controlPanelEquipmentTitleEdit.selectWordAt(index);
      this.syncControlPanelEquipmentTitleFromNativeEdit();
      this.controlPanelEquipmentTitleSuppressClick = true;
      return true;
    }

    this.controlPanelEquipmentTitleEditDragActive = true;
    this.controlPanelEquipmentTitleEditDragAnchorIndex = index;
    this.controlPanelEquipmentTitleSuppressClick = false;
    this.controlPanelEquipmentTitleEdit.focus(text, index, index);
    this.syncControlPanelEquipmentTitleFromNativeEdit();
    return true;
  }

  // Обновляет выделение названия объекта при перетаскивании по общему полю ввода.
  private updateControlPanelObjectTitleEditSelectionFromCursor(): void {
    const text = this.getControlPanelObjectTitle("");
    const index = this.textInputIndexAtCursor("control-panel-object-title-input", text);
    this.controlPanelObjectTitleSuppressClick = this.controlPanelObjectTitleEditDragAnchorIndex !== index;
    this.controlPanelObjectTitleEdit.focus(text, this.controlPanelObjectTitleEditDragAnchorIndex, index);
    this.syncControlPanelObjectTitleFromNativeEdit();
  }

  // Обновляет выделение названия группы оборудования при перетаскивании по общему полю ввода.
  private updateControlPanelEquipmentTitleEditSelectionFromCursor(): void {
    const text = this.getControlPanelEquipmentTitle("");
    const index = this.textInputIndexAtCursor("control-panel-equipment-title-input", text);
    this.controlPanelEquipmentTitleSuppressClick = this.controlPanelEquipmentTitleEditDragAnchorIndex !== index;
    this.controlPanelEquipmentTitleEdit.focus(text, this.controlPanelEquipmentTitleEditDragAnchorIndex, index);
    this.syncControlPanelEquipmentTitleFromNativeEdit();
  }

  // Проверяет попадание игрового курсора в визуальную строку ввода.
  private isCursorOverChatInput(): boolean {
    return this.isCursorOverTextInput("chat-input");
  }

  // Проверяет попадание игрового курсора в общее поле ввода.
  private isCursorOverTextInput(inputId: string): boolean {
    const rect = document.getElementById(inputId)?.getBoundingClientRect();
    return Boolean(rect &&
      this.cursorX >= rect.left &&
      this.cursorX <= rect.right &&
      this.cursorY >= rect.top &&
      this.cursorY <= rect.bottom);
  }

  // Находит ближайшую позицию текста в строке чата по координате игрового курсора.
  private chatTextIndexAtCursor(): number {
    return this.textInputIndexAtCursor("chat-input", this.chatInputText);
  }

  // Находит ближайшую позицию текста в общем поле ввода по координате игрового курсора.
  private textInputIndexAtCursor(inputId: string, value: string): number {
    const viewport = document.querySelector<HTMLElement>(`#${inputId} .ui-kit-text-input__viewport`);
    const text = document.querySelector<HTMLElement>(`#${inputId} .ui-kit-text-input__text`);
    const measure = document.querySelector<HTMLElement>(`#${inputId} .ui-kit-text-input__measure`);
    const viewportRect = viewport?.getBoundingClientRect();
    if (!viewportRect || value.length === 0) {
      return value.length;
    }

    const textWidth = measure?.getBoundingClientRect().width ?? value.length * 8;
    const charWidth = Math.max(1, textWidth / Math.max(1, value.length));
    const transform = text?.style.transform ?? "";
    const offset = Number(transform.match(/translateX\((-?\d+(?:\.\d+)?)px\)/)?.[1] ?? 0);
    const localX = this.cursorX - viewportRect.left - offset;
    return clamp(Math.round(localX / charWidth), 0, value.length);
  }

  private syncChatInputFromNativeEdit(): void {
    const snapshot = this.chatEdit.snapshot();
    this.chatInputText = snapshot.text;
    this.chatCursorIndex = snapshot.selectionEnd;
    this.saveCurrentChatInputDraft();
  }

  // Переключает активную вкладку чата вместе с ее локальным черновиком ввода.
  private setSelectedChatId(chatId: number): void {
    if (this.selectedChatId === chatId) {
      return;
    }

    this.saveCurrentChatInputDraft();
    this.selectedChatId = chatId;
    this.loadSelectedChatInputDraft();
  }

  // Запоминает текущий текст, каретку и выделение для выбранной вкладки.
  private saveCurrentChatInputDraft(): void {
    if (this.selectedChatId <= 0) {
      return;
    }

    const snapshot = this.chatEdit.snapshot();
    const text = this.chatInputFocused ? snapshot.text : this.chatInputText;
    const selectionStart = this.chatInputFocused ? snapshot.selectionStart : this.chatCursorIndex;
    const selectionEnd = this.chatInputFocused ? snapshot.selectionEnd : this.chatCursorIndex;
    if (text === "") {
      this.chatInputDrafts.delete(this.selectedChatId);
      return;
    }

    this.chatInputDrafts.set(this.selectedChatId, {
      text,
      cursorIndex: selectionEnd,
      selectionStart,
      selectionEnd,
    });
  }

  // Возвращает строку ввода к черновику выбранной вкладки.
  private loadSelectedChatInputDraft(): void {
    const draft = this.chatInputDrafts.get(this.selectedChatId) ?? {
      text: "",
      cursorIndex: 0,
      selectionStart: 0,
      selectionEnd: 0,
    };
    this.chatInputText = draft.text;
    this.chatCursorIndex = clamp(draft.cursorIndex, 0, draft.text.length);
    if (this.chatInputFocused) {
      this.chatEdit.focus(
        draft.text,
        clamp(draft.selectionStart, 0, draft.text.length),
        clamp(draft.selectionEnd, 0, draft.text.length),
      );
    }
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
    const tabTop = panelTop + (chatPanelPaddingVh + chatMessagesHeightVh + chatPanelGapVh + chatInputHeightVh + chatPanelGapVh) * vh;
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

const formatFuelDrainAmount = (value: number): string => {
  if (!Number.isFinite(value)) {
    return "0";
  }
  return String(Math.max(0, value));
};

// Возвращает ID списка по корню, строке или общей полосе прокрутки ListBox.
const scrollableListIdFromControl = (control: GameUiControlState): string | null => {
  if (control.kind === "scrollbar" && control.id.endsWith("-scrollbar")) {
    return control.id.slice(0, -"-scrollbar".length);
  }
  if (control.kind !== "list") {
    return null;
  }
  const element = document.getElementById(control.id);
  const listRoot = element?.closest<HTMLElement>(".ui-kit-list");
  return listRoot?.id ?? control.id;
};

// Преобразует базовые сочетания Alt в одноразовые команды запросов и стыковки.
const getDefaultDockingActionFromKey = (event: KeyboardEvent, wasPressed: boolean): DockingAction | null => {
  if (!event.altKey || wasPressed) {
    return null;
  }
  if (event.code === "Equal" || event.code === "NumpadAdd") {
    return "request";
  }
  if (event.code === "Minus" || event.code === "NumpadSubtract") {
    return "undock";
  }
  if (event.code === "Digit1" || event.code === "Numpad1") {
    return "approve";
  }
  if (event.code === "Digit2" || event.code === "Numpad2") {
    return "reject";
  }
  if (event.code === "Slash" || event.code === "NumpadDivide") {
    return "landing";
  }
  return null;
};

// Возвращает клавиши-модификаторы для действия игрового UI.
const uiActionModifiers = (event: MouseEvent): Pick<GameUiAction, "ctrlKey" | "metaKey" | "shiftKey"> => ({
  ctrlKey: event.ctrlKey,
  metaKey: event.metaKey,
  shiftKey: event.shiftKey,
});
