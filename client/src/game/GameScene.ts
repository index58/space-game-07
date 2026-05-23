import * as Phaser from "phaser";
import { ASSET_KEYS, ASSET_PATHS } from "../data/assets";
import {
  INITIAL_ZOOM,
  getViewportZoomScale,
  getPilotBackgroundTransform,
  rotationToPilotScreen,
  worldToPilotScreen,
} from "../domain/camera";
import { GameClient } from "../network/GameClient";
import type {
  ConnectionStatus,
  ConstructorProductionJob,
  CosmicObject,
  CosmicObjectModelReference,
  DockingEventMessage,
  DockingNotification,
  DockingWindowState,
  EquipmentGroup,
  ExchangeEventMessage,
  ExchangeStateMessage,
  ReferenceDataMessage,
  Task,
} from "../network/protocol";
import { fetchReferenceData } from "../network/referenceData";
import type { ControlPanelConstructorTabValue, ControlPanelEquipmentSubTabValue, ControlPanelTabValue, ControlPanelUsageSelectValue, ExchangeSelectValue, GameUiController, GameUiState, SettingsTabValue } from "../ui/gameUiState";
import { getInformationPanelView } from "../ui/informationPanel";
import { getInputBindingMap, getInputSettingsLeftColumnRowCount, getMergedInputSettingValues, toInputSettingsPayload } from "../ui/inputSettings";
import { getNextPilotToolIndex } from "../ui/pilotToolbar";
import { getScrollOffsetFromThumbTopPercent, getScrollbarThumbTopPercentFromCursor, startScrollbarDrag, type ScrollbarDragState } from "../ui-kit/scrollbar";
import { getUiKitControlHitRect } from "../ui-kit/hitRect";
import { getCountSliderValue } from "../ui-kit/slider";
import { applyUiKitDemoAction, createInitialUiKitDemoState, type UiKitDemoState } from "../ui-kit/showcaseState";
import type { GameUiAction, GameUiControlKind, GameUiControlState } from "../ui-kit/types";
import { bodyPolygonToPilotScreen } from "./bodyPolygon";
import { applyControlPanelPendingToEquipmentGroups, applyControlPanelPendingToObject, emptyControlPanelPendingState, pruneControlPanelPending, rejectControlPanelPending, type ControlPanelPendingState } from "./controlPanelMutations";
import { getControlPanelFuelFillMaxAmount } from "./controlPanelFuelAmount";
import { applyControlPanelListSelection } from "./controlPanelListSelection";
import { applyActiveControlPanelUsageRelations, normalizeControlPanelUsageSelection } from "./controlPanelUsageSelection";
import {
  SIMPLE_DRILL_RAY_ACRONYM,
  clipDrillBeamGeometryToPolygons,
  getDrillBeamGeometry,
  getDrillBeamIntakeProgress,
  type DrillBeamGeometry,
  type DrillBeamPoint,
} from "./drillBeam";
import { FrameRateMeter } from "./frameRateMeter";
import { getGameUiControlLayoutSignature } from "./gameUiControlSignature";
import { InputController, type ChatScrollState } from "./InputController";

const BODY_POLYGON_DEBUG_COLOR = 0x35d7ff;
// Базовая высота строки настроек ввода в единицах высоты экрана.
const SETTINGS_INPUT_ROW_HEIGHT_VH = 2.7;
// Базовая высота пункта раскрытого списка в единицах высоты экрана.
const SETTINGS_DROPDOWN_ITEM_HEIGHT_VH = 2.3;
// Суммарный вертикальный отступ содержимого раскрытого списка в единицах высоты экрана.
const SETTINGS_DROPDOWN_CONTENT_PADDING_VH = 0.7;
// Видимая высота раскрытого списка в единицах высоты экрана.
const SETTINGS_DROPDOWN_VIEWPORT_HEIGHT_VH = 22;
// Базовая высота пункта списка оборудования в единицах высоты экрана.
const CONTROL_PANEL_EQUIPMENT_LIST_ITEM_HEIGHT_VH = 2.35;
// Время показа маленького уведомления стыковки.
const DOCKING_NOTIFICATION_DURATION_MS = 5000;
// Базовая длительность окна запроса обмена для обратной полоски.
const EXCHANGE_REQUEST_DURATION_MS = 10000;

// Защищает анимацию запроса обмена от пустой длительности.
const getExchangeRequestDurationMs = (durationSeconds: number | undefined): number => {
  if (durationSeconds === undefined || durationSeconds <= 0) {
    return EXCHANGE_REQUEST_DURATION_MS;
  }
  return durationSeconds * 1000;
};

// Связывает Phaser-отрисовку, сетевой клиент, ввод и SolidJS UI-слой.
export class GameScene extends Phaser.Scene {
  // Спрайты объектов, переиспользуемые между серверными снимками.
  private objectSprites = new Map<number, Phaser.GameObjects.Image>();
  // Векторный слой отладочной отрисовки физических тел.
  private bodyPolygonGraphics!: Phaser.GameObjects.Graphics;
  // Векторный слой активных эффектов инструментов пилота.
  private pilotToolEffectGraphics!: Phaser.GameObjects.Graphics;
  // Измеряет простую среднюю частоту кадров за последнюю секунду.
  private frameRateMeter = new FrameRateMeter();
  // Включает показ серверных физических тел поверх текстур.
  private bodyPolygonDebugVisible = false;
  // Тайловое изображение космоса под всеми объектами.
  private background!: Phaser.GameObjects.TileSprite;
  // Контроллер клавиатуры, мыши и захвата указателя.
  private inputController!: InputController;
  // Сетевой клиент для получения снимков и отправки ввода.
  private gameClient: GameClient | null = null;
  // Справочники, полученные с сервера перед подключением к игровому потоку.
  private referenceData: ReferenceDataMessage | null = null;
  // Ключи текстур, для которых уже запущена асинхронная загрузка.
  private loadingTextureKeys = new Set<string>();
  // Сообщение об ошибке начальной загрузки, если сервер не отдал справочники.
  private startupErrorMessage: string | null = null;
  // Текст ожидания, показанный до появления валидного снимка.
  private waitingText!: Phaser.GameObjects.Text;
  // Дискретный пользовательский уровень приближения.
  private zoomLevel = INITIAL_ZOOM;
  // Рассчитанный масштаб мира в пикселях на метр.
  private zoomScale = getViewportZoomScale(INITIAL_ZOOM, 1000);
  // Выбранный индекс среди десяти ячеек панели пилота.
  private selectedPilotToolIndex = 0;
  // Состояние интерактивной витрины базовых контролов UI Kit.
  private uiKitDemoState: UiKitDemoState = createInitialUiKitDemoState();
  // Черновик настроек ввода, показанный в модальном окне.
  private inputSettingsValues: Record<number, number> = {};
  // Последний примененный номер серверных настроек ввода.
  private inputSettingsSeq = -1;
  // ID действия с раскрытым выпадающим списком в окне настроек.
  private openInputSettingsActionId: number | null = null;
  // Последний примененный номер ошибки сохранения настроек.
  private inputSettingsErrorSeq = -1;
  // Локальные выбранные связи групп оборудования, ещё не пришедшие в серверном снимке.
  private controlPanelEquipmentGroupRelationDrafts: Record<string, number> = {};
  // Текст ошибки сохранения настроек.
  private inputSettingsError: string | null = null;
  // Показывает ожидание ответа сервера после нажатия кнопки сохранения.
  private inputSettingsSaving = false;
  // Активная страница окна настроек.
  private selectedSettingsTab: SettingsTabValue = "input";
  // Активная страница панели управления.
  private selectedControlPanelTab: ControlPanelTabValue = "object";
  // Активная подстраница оборудования в панели управления.
  private selectedControlPanelEquipmentTab: ControlPanelEquipmentSubTabValue = "setup";
  // ID выбранной группы оборудования в панели управления.
  private selectedControlPanelEquipmentGroupId: number | null = null;
  // ID контейнера в левой панели использования оборудования.
  private selectedControlPanelUsageLeftContainerGroupId: number | null = null;
  // ID оборудования в правой панели использования оборудования.
  private selectedControlPanelUsageRightEquipmentGroupId: number | null = null;
  // Открытый выпадающий список использования оборудования.
  private openControlPanelUsageSelect: ControlPanelUsageSelectValue | null = null;
  // Выбранный объект для левого контейнера использования.
  private selectedControlPanelUsageLeftObjectId: number | null = null;
  // Выбранный объект для правого оборудования использования.
  private selectedControlPanelUsageRightObjectId: number | null = null;
  // Выбранный объект для контейнера материалов конструктора.
  private selectedControlPanelConstructorMaterialObjectId: number | null = null;
  // Выбранный объект для контейнера продукции конструктора.
  private selectedControlPanelConstructorProductObjectId: number | null = null;
  // Выбранные строки содержимого левого контейнера.
  private selectedControlPanelUsageLeftItemGroupIds: number[] = [];
  // Опорная строка левого контейнера для выбора диапазона через Shift.
  private selectedControlPanelUsageLeftAnchorItemGroupId: number | null = null;
  // Выбранные строки содержимого правого контейнера.
  private selectedControlPanelUsageRightItemGroupIds: number[] = [];
  // Опорная строка правого контейнера для выбора диапазона через Shift.
  private selectedControlPanelUsageRightAnchorItemGroupId: number | null = null;
  // ID контейнера, из которого конструктор берёт материалы.
  private selectedControlPanelConstructorMaterialContainerGroupId: number | null = null;
  // ID контейнера, в который конструктор кладёт продукцию.
  private selectedControlPanelConstructorProductContainerGroupId: number | null = null;
  // Активная вкладка списка заданий конструктора.
  private selectedControlPanelConstructorTab: ControlPanelConstructorTabValue = "items";
  // ID схемы, выбранной в списке конструктора.
  private selectedControlPanelConstructorSchemaId: number | null = null;
  // ID чертежа, выбранного в списке конструктора.
  private selectedControlPanelConstructorBlueprintId: number | null = null;
  // ID выбранной строки основной очереди конструктора.
  private selectedControlPanelConstructorMainJobId: number | null = null;
  // Показывает окно подтверждения слива топлива из бака.
  private controlPanelFuelDrainDialogOpen = false;
  // Показывает окно подтверждения залива топлива в бак.
  private controlPanelFuelFillDialogOpen = false;
  // Показывает окно подтверждения частичного переноса предметов между контейнерами.
  private controlPanelContainerTransferDialogOpen = false;
  // Показывает окно выбора количества запусков изготовления.
  private controlPanelConstructorProduceDialogOpen = false;
  // Показывает окно выбора количества предметов для деконструкции одной строки.
  private controlPanelItemDeconstructionDialogOpen = false;
  // Максимальное количество предметов для частичного переноса между контейнерами.
  private controlPanelContainerTransferMaxAmount = 0;
  // Источник ожидающего подтверждения переноса между контейнерами.
  private controlPanelContainerTransferSourceGroupId: number | null = null;
  // Получатель ожидающего подтверждения переноса между контейнерами.
  private controlPanelContainerTransferTargetGroupId: number | null = null;
  // Правый контейнер, управляющий ожидающим подтверждения перемещением.
  private controlPanelContainerTransferControllerGroupId: number | null = null;
  // Направление ожидающего подтверждения перемещения относительно правого контейнера.
  private controlPanelContainerTransferLeftToRightDirection = true;
  // Строка содержимого, ожидающая частичного переноса между контейнерами.
  private controlPanelContainerTransferItemGroupIds: number[] = [];
  // Деконструктор, ожидающий подтверждения частичной деконструкции.
  private controlPanelItemDeconstructionDeconstructorGroupId: number | null = null;
  // Контейнер-источник, ожидающий подтверждения частичной деконструкции.
  private controlPanelItemDeconstructionSourceGroupId: number | null = null;
  // Контейнер-приёмник, ожидающий подтверждения частичной деконструкции.
  private controlPanelItemDeconstructionTargetGroupId: number | null = null;
  // Строка содержимого, ожидающая подтверждения частичной деконструкции.
  private controlPanelItemDeconstructionItemGroupIds: number[] = [];
  // Максимальное количество топлива, доступное для залива в бак.
  private controlPanelFuelFillMaxAmount = 0;
  // Максимальное количество запусков изготовления в окне конструктора.
  private controlPanelConstructorProduceMaxAmount = 100;
  // Максимальное количество предметов для частичной деконструкции.
  private controlPanelItemDeconstructionMaxAmount = 0;
  // Количество топлива, выбранное для слива из бака.
  private controlPanelFuelDrainAmount = 0;
  // Ожидающие подтверждения сервером изменения панели управления.
  private controlPanelPending: ControlPanelPendingState = emptyControlPanelPendingState();
  // Последний обработанный номер ошибки панели управления.
  private controlPanelErrorSeq = -1;
  // Вертикальный сдвиг списка групп оборудования.
  private controlPanelEquipmentListScrollOffsetPx = 0;
  // Захват ползунка списка групп оборудования.
  private controlPanelEquipmentListScrollbarDrag: ScrollbarDragState | null = null;
  // Вертикальные сдвиги обычных списков панели управления по ID списка.
  private readonly listScrollOffsets = new Map<string, number>();
  // Захваты ползунков обычных списков панели управления по ID списка.
  private readonly listScrollbarDrags = new Map<string, ScrollbarDragState>();
  // Предыдущее состояние видимости окна настроек для запроса свежих данных при открытии.
  private previousSettingsVisible = false;
  // Вертикальный сдвиг списка действий в окне настроек.
  private settingsInputScrollOffsetPx = 0;
  // Захват ползунка списка действий.
  private settingsInputScrollbarDrag: ScrollbarDragState | null = null;
  // Вертикальный сдвиг раскрытого списка событий ввода.
  private settingsDropdownScrollOffsetPx = 0;
  // Захват ползунка раскрытого списка событий ввода.
  private settingsDropdownScrollbarDrag: ScrollbarDragState | null = null;
  // Последняя раскладка DOM-контролов, для которой уже обновлен hit-test.
  private lastGameUiControlLayoutSignature = "";
  // Количество ближайших кадров, в которых нужно перечитать DOM после изменения раскладки.
  private gameUiControlRefreshFrames = 0;
  // Текущее маленькое окно ожидания или процесса.
  private dockingWindow: DockingWindowState | null = null;
  // Отдельные всплывающие уведомления стыковки.
  private dockingNotifications: DockingNotification[] = [];
  // Следующий локальный идентификатор уведомления.
  private nextDockingNotificationID = 1;
  // Объекты, доступные для выбора при пересадке из главного объекта.
  private landingTargetObjectIds: number[] = [];
  // Текущий выбор в окне пересадки.
  private selectedLandingTargetObjectId: number | null = null;
  // Текущее состояние окна обмена от сервера.
  private exchangeState: ExchangeStateMessage | null = null;
  // Роль входящего или исходящего запроса обмена до открытия окна.
  private exchangeRequestRole: "sender" | "receiver" | null = null;
  // Открытый выпадающий список обмена.
  private openExchangeSelect: ExchangeSelectValue | null = null;
  // Выбранный объект для контейнера-приемника обмена.
  private selectedExchangeReceiverObjectId: number | null = null;
  // Выбранный объект для контейнера-источника обмена.
  private selectedExchangeSourceObjectId: number | null = null;
  // Выделенные строки контейнера-источника обмена.
  private selectedExchangeSourceItemGroupIds: number[] = [];
  // Опорная строка источника обмена для выбора диапазона через Shift.
  private selectedExchangeSourceAnchorItemGroupId: number | null = null;

  constructor(
    // Мост передачи состояния из Phaser в SolidJS UI.
    private readonly gameUi: GameUiController,
  ) {
    super("GameScene");
  }

  // Регистрирует только фон; объектные текстуры приходят из серверных справочников.
  preload(): void {
    for (const [key, path] of Object.entries(ASSET_PATHS)) {
      this.load.image(key, path);
    }
  }

  // Создает постоянные объекты сцены и подключает клиентские контроллеры.
  create(): void {
    this.background = this.add
      .tileSprite(0, 0, this.scale.width, this.scale.height, ASSET_KEYS.background)
      .setOrigin(0.5);
    this.waitingText = this.add
      .text(this.scale.width / 2, this.scale.height / 2, "Ожидание подключения к серверу", {
        color: "#d8f3ff",
        fontFamily: "Consolas, monospace",
        fontSize: "18px",
      })
      .setOrigin(0.5);
    this.pilotToolEffectGraphics = this.add.graphics().setDepth(900).setBlendMode(Phaser.BlendModes.ADD);
    this.bodyPolygonGraphics = this.add.graphics().setDepth(1000);

    this.inputController = new InputController(
      this.game.canvas,
      () => this.gameClient?.getStatus() === "connected",
    );
    void this.loadStartupData();
  }

  // Каждый кадр отправляет свежий ввод и рисует последний серверный снимок мира.
  update(time: number, _deltaMs: number): void {
    const measuredFps = this.frameRateMeter.recordFrame(time);
    const input = this.inputController.consumeShipInput();
    this.gameClient?.setInput(input);
    if (this.inputController.consumeRandomShipChangeRequest()) {
      this.gameClient?.requestRandomShipChange();
    }
    this.consumeDockingActions();
    this.syncDockingEventFromServer(time);
    this.syncExchangeEventFromServer(time);
    this.dockingNotifications = this.dockingNotifications.filter((notification) => notification.expiresAtMs > time);
    if (this.inputController.consumeBodyPolygonDebugToggleRequest()) {
      this.bodyPolygonDebugVisible = !this.bodyPolygonDebugVisible;
      if (!this.bodyPolygonDebugVisible) {
        this.bodyPolygonGraphics.clear();
      }
    }
    this.zoomLevel = this.inputController.getZoom();
    const pilotToolSelectionDelta = this.inputController.consumePilotToolSelectionDelta();

    const status = this.gameClient?.getStatus() ?? "connecting";
    const settingsVisible = this.inputController.isSettingsVisible();
    const controlPanelVisible = this.inputController.isControlPanelVisible();
    this.requestInputSettingsOnOpen(settingsVisible);
    this.syncInputSettingsFromServer();
    const snapshot = this.gameClient?.getLatestSnapshot() ?? null;
    this.syncControlPanelPendingFromServer(snapshot);
    const isWorldReady = status === "connected" && Boolean(snapshot);
    const chatState = isWorldReady ? this.inputController.getVisibleChatState(this.gameClient?.getLatestChatState() ?? null) : null;
    const chatAction = this.inputController.consumeChatAction();
    if (chatAction) {
      this.gameClient?.sendChatMessage(chatAction);
    }
    const chatSelectAction = this.inputController.consumeChatSelectAction();
    if (chatSelectAction) {
      this.gameClient?.selectChat(chatSelectAction.chatId);
    }
    this.consumeGameUiActions();
    this.consumeGameUiWheelActions();
    this.consumeSettingsWheel();
    const inputSettingsScroll = this.getSettingsInputScrollState();
    const inputSettingsDropdownScroll = this.getSettingsDropdownScrollState();
    const serverSelfObject = snapshot?.objects.find((object) => object.ID === snapshot.selfObjectId) ?? null;
    const effectiveEquipmentGroups = snapshot ? applyControlPanelPendingToEquipmentGroups(snapshot.equipmentGroups ?? [], this.controlPanelPending) : [];
    const snapshotTasks = snapshot?.tasks ?? [];
    const snapshotTaskItemGroups = snapshot?.taskItemGroups ?? [];
    const constructorProductionJobs = this.constructorProductionJobsFromTasks(snapshotTasks, effectiveEquipmentGroups);
    const selfObject = applyControlPanelPendingToObject(serverSelfObject, this.controlPanelPending);
    if (this.inputController.consumeFocusedObjectOwnerClaimRequest() && snapshot && selfObject && this.referenceData && getInformationPanelView({ selfObject, objects: snapshot.objects, referenceData: this.referenceData })) {
      this.gameClient?.requestFocusedObjectOwnerClaim();
    }
    this.syncControlPanelUsageSelection(selfObject?.ID ?? null, snapshot?.objects ?? [], effectiveEquipmentGroups);
    this.syncControlPanelConstructorMainJobSelection(constructorProductionJobs, snapshotTasks);
    this.controlPanelFuelFillMaxAmount = this.getControlPanelFuelFillMaxAmount(snapshot?.objects ?? [], effectiveEquipmentGroups, snapshot?.itemGroups ?? []);
    this.inputController.syncControlPanelObject(selfObject);
    if (this.controlPanelFuelDrainDialogOpen || this.controlPanelFuelFillDialogOpen || this.controlPanelContainerTransferDialogOpen || this.controlPanelConstructorProduceDialogOpen || this.controlPanelItemDeconstructionDialogOpen) {
      const maxAmount = this.currentControlPanelAmountMax(selfObject);
      this.controlPanelFuelDrainAmount = clamp(this.inputController.getControlPanelFuelDrainAmount(this.controlPanelFuelDrainAmount), 0, maxAmount);
    }
    const controlPanelObjectTitleEditState = this.inputController.getControlPanelObjectTitleEditState(selfObject?.Title ?? "");
    const selectedControlPanelEquipmentGroup = this.getSelectedControlPanelEquipmentGroupFromList(effectiveEquipmentGroups, selfObject?.ID ?? null);
    this.inputController.syncControlPanelEquipmentTitle(selectedControlPanelEquipmentGroup?.ID ?? null, selectedControlPanelEquipmentGroup?.Title ?? "");
    const controlPanelEquipmentTitleEditState = this.inputController.getControlPanelEquipmentTitleEditState(selectedControlPanelEquipmentGroup?.Title ?? "");
    const controlPanelFuelDrainAmountEditState = this.inputController.getControlPanelFuelDrainAmountEditState();
    this.commitControlPanelObjectTitleIfNeeded(serverSelfObject);
    this.commitControlPanelEquipmentTitleIfNeeded(selectedControlPanelEquipmentGroup);

    this.zoomScale = getViewportZoomScale(this.zoomLevel, this.scale.height);
    this.inputController.setExchangeVisible(this.exchangeState !== null);

    if (status !== "connected" || !snapshot || !selfObject) {
      this.renderWaiting(status);
      this.gameUi.update({
        status,
        selfObject: null,
        objects: snapshot?.objects ?? [],
        equipmentGroups: effectiveEquipmentGroups,
        equipmentGroupRelations: snapshot?.equipmentGroupRelations ?? [],
        itemGroups: snapshot?.itemGroups ?? [],
        tasks: snapshotTasks,
        taskItemGroups: snapshotTaskItemGroups,
        constructorProductionJobs,
        selectedPilotToolIndex: this.selectedPilotToolIndex,
        referenceData: this.referenceData,
        textureFilePath: null,
        chatState: null,
        chatInputText: "",
        chatCursorIndex: 0,
        chatSelectionStart: 0,
        chatSelectionEnd: 0,
        chatInputFocused: false,
        chatError: null,
        chatErrorSeq: 0,
        dockingWindow: this.dockingWindow,
        dockingNotifications: this.dockingNotifications,
        landingTargetObjectIds: this.landingTargetObjectIds,
        exchangeState: this.exchangeState,
        openExchangeSelect: this.openExchangeSelect,
        selectedExchangeReceiverObjectId: this.selectedExchangeReceiverObjectId,
        selectedExchangeSourceObjectId: this.selectedExchangeSourceObjectId,
        selectedExchangeSourceItemGroupIds: this.selectedExchangeSourceItemGroupIds,
        selectedLandingTargetObjectId: this.selectedLandingTargetObjectId,
        chatContextMenu: null,
        gameCursor: this.inputController.getGameCursor(),
        hoveredGameUiControlId: this.inputController.getHoveredGameUiControlId(),
        chatScroll: { visible: false, thumbTopPercent: 0, thumbHeightPercent: 100, contentOffsetPx: 0, dragging: false },
        uiKitShowcaseVisible: this.inputController.isUiKitShowcaseVisible(),
        settingsVisible,
        controlPanelVisible,
        selectedSettingsTab: this.selectedSettingsTab,
        selectedControlPanelTab: this.selectedControlPanelTab,
        selectedControlPanelEquipmentTab: this.selectedControlPanelEquipmentTab,
        selectedControlPanelEquipmentGroupId: this.selectedControlPanelEquipmentGroupId,
        selectedControlPanelUsageLeftObjectId: this.selectedControlPanelUsageLeftObjectId,
        selectedControlPanelUsageRightObjectId: this.selectedControlPanelUsageRightObjectId,
        selectedControlPanelConstructorMaterialObjectId: this.selectedControlPanelConstructorMaterialObjectId,
        selectedControlPanelConstructorProductObjectId: this.selectedControlPanelConstructorProductObjectId,
        selectedControlPanelUsageLeftContainerGroupId: this.selectedControlPanelUsageLeftContainerGroupId,
        selectedControlPanelUsageRightEquipmentGroupId: this.selectedControlPanelUsageRightEquipmentGroupId,
        openControlPanelUsageSelect: this.openControlPanelUsageSelect,
        selectedControlPanelUsageLeftItemGroupIds: this.selectedControlPanelUsageLeftItemGroupIds,
        selectedControlPanelUsageRightItemGroupIds: this.selectedControlPanelUsageRightItemGroupIds,
        selectedControlPanelConstructorMaterialContainerGroupId: this.selectedControlPanelConstructorMaterialContainerGroupId,
        selectedControlPanelConstructorProductContainerGroupId: this.selectedControlPanelConstructorProductContainerGroupId,
        selectedControlPanelConstructorTab: this.selectedControlPanelConstructorTab,
        selectedControlPanelConstructorSchemaId: this.selectedControlPanelConstructorSchemaId,
        selectedControlPanelConstructorBlueprintId: this.selectedControlPanelConstructorBlueprintId,
        selectedControlPanelConstructorMainJobId: this.selectedControlPanelConstructorMainJobId,
        controlPanelFuelDrainDialogOpen: this.controlPanelFuelDrainDialogOpen,
        controlPanelFuelFillDialogOpen: this.controlPanelFuelFillDialogOpen,
        controlPanelContainerTransferDialogOpen: this.controlPanelContainerTransferDialogOpen,
        controlPanelConstructorProduceDialogOpen: this.controlPanelConstructorProduceDialogOpen,
        controlPanelItemDeconstructionDialogOpen: this.controlPanelItemDeconstructionDialogOpen,
        controlPanelContainerTransferMaxAmount: this.controlPanelContainerTransferMaxAmount,
        controlPanelFuelFillMaxAmount: this.controlPanelFuelFillMaxAmount,
        controlPanelConstructorProduceMaxAmount: this.controlPanelConstructorProduceMaxAmount,
        controlPanelItemDeconstructionMaxAmount: this.controlPanelItemDeconstructionMaxAmount,
        controlPanelFuelDrainAmount: this.controlPanelFuelDrainAmount,
        controlPanelFuelDrainAmountText: controlPanelFuelDrainAmountEditState.text,
        controlPanelFuelDrainAmountSelectionStart: controlPanelFuelDrainAmountEditState.selectionStart,
        controlPanelFuelDrainAmountSelectionEnd: controlPanelFuelDrainAmountEditState.selectionEnd,
        controlPanelFuelDrainAmountFocused: controlPanelFuelDrainAmountEditState.focused,
        controlPanelEquipmentEnabledDrafts: {},
        controlPanelEquipmentEnabledCountDrafts: {},
        controlPanelEquipmentListScroll: this.getControlPanelEquipmentListScrollState(),
        listScroll: this.getListScrollStates(),
        controlPanelObjectEnabled: this.inputController.getControlPanelObjectEnabled(false),
        controlPanelEquipmentTitleText: controlPanelEquipmentTitleEditState.text,
        controlPanelEquipmentTitleSelectionStart: controlPanelEquipmentTitleEditState.selectionStart,
        controlPanelEquipmentTitleSelectionEnd: controlPanelEquipmentTitleEditState.selectionEnd,
        controlPanelEquipmentTitleFocused: controlPanelEquipmentTitleEditState.focused,
        controlPanelObjectTitleText: controlPanelObjectTitleEditState.text,
        controlPanelObjectTitleSelectionStart: controlPanelObjectTitleEditState.selectionStart,
        controlPanelObjectTitleSelectionEnd: controlPanelObjectTitleEditState.selectionEnd,
        controlPanelObjectTitleFocused: controlPanelObjectTitleEditState.focused,
        inputSettingsValues: this.inputSettingsValues,
        openInputSettingsActionId: this.openInputSettingsActionId,
        inputSettingsError: this.inputSettingsError,
        inputSettingsSaving: this.inputSettingsSaving,
        inputSettingsScroll,
        inputSettingsDropdownScroll,
        uiKitDemoState: this.uiKitDemoState,
        uiControls: [],
        fps: measuredFps,
        zoom: this.zoomScale,
      });
      this.updateGameUiControlsIfNeeded(this.gameUi.state());
      return;
    }

    if (pilotToolSelectionDelta !== 0) {
      this.selectedPilotToolIndex = getNextPilotToolIndex(this.selectedPilotToolIndex, pilotToolSelectionDelta);
    }

    this.waitingText.setVisible(false);
    this.renderWorld(snapshot.objects, selfObject, time);
    this.gameUi.update({
      status,
      selfObject,
      objects: snapshot.objects,
      equipmentGroups: effectiveEquipmentGroups,
      equipmentGroupRelations: snapshot.equipmentGroupRelations ?? [],
      itemGroups: snapshot.itemGroups ?? [],
      tasks: snapshotTasks,
      taskItemGroups: snapshotTaskItemGroups,
      constructorProductionJobs,
      selectedPilotToolIndex: this.selectedPilotToolIndex,
      referenceData: this.referenceData,
      textureFilePath: this.modelForObject(selfObject)?.TextureFilePath ?? null,
      chatState,
      chatInputText: this.inputController.getChatInputText(),
      chatCursorIndex: this.inputController.getChatCursorIndex(),
      chatSelectionStart: this.inputController.getChatSelectionStart(),
      chatSelectionEnd: this.inputController.getChatSelectionEnd(),
      chatInputFocused: this.inputController.isChatInputFocused(),
      chatError: this.gameClient?.getLatestChatError() ?? null,
      chatErrorSeq: this.gameClient?.getLatestChatErrorSeq() ?? 0,
      dockingWindow: this.dockingWindow,
      dockingNotifications: this.dockingNotifications,
      landingTargetObjectIds: this.landingTargetObjectIds,
      exchangeState: this.exchangeState,
      openExchangeSelect: this.openExchangeSelect,
      selectedExchangeReceiverObjectId: this.selectedExchangeReceiverObjectId,
      selectedExchangeSourceObjectId: this.selectedExchangeSourceObjectId,
      selectedExchangeSourceItemGroupIds: this.selectedExchangeSourceItemGroupIds,
      selectedLandingTargetObjectId: this.selectedLandingTargetObjectId,
      chatContextMenu: this.inputController.getChatContextMenu(),
      gameCursor: this.inputController.getGameCursor(),
      hoveredGameUiControlId: this.inputController.getHoveredGameUiControlId(),
      chatScroll: this.inputController.getChatScrollState(),
      uiKitShowcaseVisible: this.inputController.isUiKitShowcaseVisible(),
      settingsVisible,
      controlPanelVisible,
      selectedSettingsTab: this.selectedSettingsTab,
      selectedControlPanelTab: this.selectedControlPanelTab,
      selectedControlPanelEquipmentTab: this.selectedControlPanelEquipmentTab,
      selectedControlPanelEquipmentGroupId: this.selectedControlPanelEquipmentGroupId,
      selectedControlPanelUsageLeftObjectId: this.selectedControlPanelUsageLeftObjectId,
      selectedControlPanelUsageRightObjectId: this.selectedControlPanelUsageRightObjectId,
      selectedControlPanelConstructorMaterialObjectId: this.selectedControlPanelConstructorMaterialObjectId,
      selectedControlPanelConstructorProductObjectId: this.selectedControlPanelConstructorProductObjectId,
      selectedControlPanelUsageLeftContainerGroupId: this.selectedControlPanelUsageLeftContainerGroupId,
      selectedControlPanelUsageRightEquipmentGroupId: this.selectedControlPanelUsageRightEquipmentGroupId,
      openControlPanelUsageSelect: this.openControlPanelUsageSelect,
      selectedControlPanelUsageLeftItemGroupIds: this.selectedControlPanelUsageLeftItemGroupIds,
      selectedControlPanelUsageRightItemGroupIds: this.selectedControlPanelUsageRightItemGroupIds,
      selectedControlPanelConstructorMaterialContainerGroupId: this.selectedControlPanelConstructorMaterialContainerGroupId,
      selectedControlPanelConstructorProductContainerGroupId: this.selectedControlPanelConstructorProductContainerGroupId,
      selectedControlPanelConstructorTab: this.selectedControlPanelConstructorTab,
      selectedControlPanelConstructorSchemaId: this.selectedControlPanelConstructorSchemaId,
      selectedControlPanelConstructorBlueprintId: this.selectedControlPanelConstructorBlueprintId,
      selectedControlPanelConstructorMainJobId: this.selectedControlPanelConstructorMainJobId,
      controlPanelFuelDrainDialogOpen: this.controlPanelFuelDrainDialogOpen,
      controlPanelFuelFillDialogOpen: this.controlPanelFuelFillDialogOpen,
      controlPanelContainerTransferDialogOpen: this.controlPanelContainerTransferDialogOpen,
      controlPanelConstructorProduceDialogOpen: this.controlPanelConstructorProduceDialogOpen,
      controlPanelItemDeconstructionDialogOpen: this.controlPanelItemDeconstructionDialogOpen,
      controlPanelContainerTransferMaxAmount: this.controlPanelContainerTransferMaxAmount,
      controlPanelFuelFillMaxAmount: this.controlPanelFuelFillMaxAmount,
      controlPanelConstructorProduceMaxAmount: this.controlPanelConstructorProduceMaxAmount,
      controlPanelItemDeconstructionMaxAmount: this.controlPanelItemDeconstructionMaxAmount,
      controlPanelFuelDrainAmount: this.controlPanelFuelDrainAmount,
      controlPanelFuelDrainAmountText: controlPanelFuelDrainAmountEditState.text,
      controlPanelFuelDrainAmountSelectionStart: controlPanelFuelDrainAmountEditState.selectionStart,
      controlPanelFuelDrainAmountSelectionEnd: controlPanelFuelDrainAmountEditState.selectionEnd,
      controlPanelFuelDrainAmountFocused: controlPanelFuelDrainAmountEditState.focused,
      controlPanelEquipmentEnabledDrafts: {},
      controlPanelEquipmentEnabledCountDrafts: {},
      controlPanelEquipmentListScroll: this.getControlPanelEquipmentListScrollState(),
      listScroll: this.getListScrollStates(),
      controlPanelObjectEnabled: this.inputController.getControlPanelObjectEnabled(selfObject.Enabled),
      controlPanelEquipmentTitleText: controlPanelEquipmentTitleEditState.text,
      controlPanelEquipmentTitleSelectionStart: controlPanelEquipmentTitleEditState.selectionStart,
      controlPanelEquipmentTitleSelectionEnd: controlPanelEquipmentTitleEditState.selectionEnd,
      controlPanelEquipmentTitleFocused: controlPanelEquipmentTitleEditState.focused,
      controlPanelObjectTitleText: controlPanelObjectTitleEditState.text,
      controlPanelObjectTitleSelectionStart: controlPanelObjectTitleEditState.selectionStart,
      controlPanelObjectTitleSelectionEnd: controlPanelObjectTitleEditState.selectionEnd,
      controlPanelObjectTitleFocused: controlPanelObjectTitleEditState.focused,
      inputSettingsValues: this.inputSettingsValues,
      openInputSettingsActionId: this.openInputSettingsActionId,
      inputSettingsError: this.inputSettingsError,
      inputSettingsSaving: this.inputSettingsSaving,
      inputSettingsScroll,
      inputSettingsDropdownScroll,
      uiKitDemoState: this.uiKitDemoState,
      uiControls: [],
      fps: measuredFps,
      zoom: this.zoomScale,
    });
    this.updateGameUiControlsIfNeeded(this.gameUi.state());
  }

  // Показывает фон и статус, пока нет валидного снимка объекта игрока.
  private renderWaiting(status: ConnectionStatus): void {
    this.waitingText.setVisible(true);
    this.waitingText.setPosition(this.scale.width / 2, this.scale.height / 2);
    this.waitingText.setText(
      this.startupErrorMessage ??
        (status === "connecting" ? "Подключение к серверу" : "Ожидание подключения к серверу"),
    );

    this.renderBackground({
      shipPosition: { x: 0, y: 0 },
      shipRotation: 0,
      zoom: this.zoomScale,
      viewportWidth: this.scale.width,
      viewportHeight: this.scale.height,
    });

    for (const sprite of this.objectSprites.values()) {
      sprite.setVisible(false);
    }
    this.pilotToolEffectGraphics.clear();
    this.bodyPolygonGraphics.clear();
  }

  // Размещает все объекты в экранных координатах камеры пилота.
  private renderWorld(
    objects: CosmicObject[],
    selfObject: CosmicObject,
    timeMs: number,
  ): void {
    const viewportWidth = this.scale.width;
    const viewportHeight = this.scale.height;
    const camera = {
      shipPosition: { x: selfObject.X, y: selfObject.Y },
      shipRotation: selfObject.Rotation,
      zoom: this.zoomScale,
      viewportWidth,
      viewportHeight,
    };

    this.renderBackground(camera);

    const activeObjectIds = new Set<number>();
    for (const object of objects) {
      if (this.isDrillRayObject(object)) {
        continue;
      }
      activeObjectIds.add(object.ID);

      const sprite = this.getOrCreateObjectSprite(object);
      if (!sprite) {
        continue;
      }
      const screen = worldToPilotScreen({ x: object.X, y: object.Y }, camera);
      this.updateObjectSpriteOrigin(sprite, object);

      sprite.setVisible(true);
      sprite.setPosition(screen.x, screen.y);
      // Корабль игрока всегда смотрит вверх экрана, остальные объекты вращаются относительно него.
      sprite.setRotation(object.ID === selfObject.ID ? 0 : rotationToPilotScreen(object.Rotation, selfObject.Rotation));
      sprite.setScale(this.zoomScale / this.textureScaleForObject(object));
    }

    for (const [objectId, sprite] of this.objectSprites) {
      if (!activeObjectIds.has(objectId)) {
        sprite.destroy();
        this.objectSprites.delete(objectId);
      }
    }
    this.renderBodyPolygons(objects, camera);
    this.renderDrillBeams(objects, camera, selfObject.Rotation, timeMs);
  }

  // Рисует полупрозрачные физические тела поверх видимых объектов.
  private renderBodyPolygons(
    objects: CosmicObject[],
    camera: Parameters<typeof worldToPilotScreen>[1],
  ): void {
    this.bodyPolygonGraphics.clear();
    if (!this.bodyPolygonDebugVisible) {
      return;
    }

    this.bodyPolygonGraphics.fillStyle(BODY_POLYGON_DEBUG_COLOR, 0.24);
    this.bodyPolygonGraphics.lineStyle(1, BODY_POLYGON_DEBUG_COLOR, 0.9);
    for (const object of objects) {
      const model = this.modelForObject(object);
      if (!model) {
        continue;
      }
      const points = bodyPolygonToPilotScreen(object, model, camera);
      if (points.length === 0) {
        continue;
      }

      this.bodyPolygonGraphics.beginPath();
      this.bodyPolygonGraphics.moveTo(points[0].x, points[0].y);
      for (const point of points.slice(1)) {
        this.bodyPolygonGraphics.lineTo(point.x, point.y);
      }
      this.bodyPolygonGraphics.closePath();
      this.bodyPolygonGraphics.fillPath();
      this.bodyPolygonGraphics.strokePath();
    }
  }

  // Рисует энергетический след активного бура поверх мира пилота.
  private renderDrillBeams(
    objects: CosmicObject[],
    camera: Parameters<typeof worldToPilotScreen>[1],
    selfRotation: number,
    timeMs: number,
  ): void {
    const graphics = this.pilotToolEffectGraphics;
    graphics.clear();

    for (const object of objects) {
      if (!this.isDrillRayObject(object)) {
        continue;
      }
      const model = this.modelForObject(object);
      if (!model) {
        continue;
      }
      const geometry = getDrillBeamGeometry({
        center: worldToPilotScreen({ x: object.X, y: object.Y }, camera),
        rotation: rotationToPilotScreen(object.Rotation, selfRotation),
        lengthMeters: model.BodyLength,
        zoomScale: this.zoomScale,
      });
      if (geometry) {
        const sourceObjectId = -object.ID;
        const obstaclePolygons = objects.flatMap((obstacle) => {
          if (obstacle.ID === object.ID || obstacle.ID === sourceObjectId || this.isDrillRayObject(obstacle)) {
            return [];
          }
          const obstacleModel = this.modelForObject(obstacle);
          return obstacleModel ? [bodyPolygonToPilotScreen(obstacle, obstacleModel, camera)] : [];
        });
        this.renderDrillBeamGeometry(clipDrillBeamGeometryToPolygons(geometry, obstaclePolygons), timeMs);
      }
    }
  }

  // Рисует один световой отрезок бура по готовой экранной геометрии.
  private renderDrillBeamGeometry(geometry: DrillBeamGeometry, timeMs: number): void {
    const graphics = this.pilotToolEffectGraphics;

    const pulse = 0.82 + Math.sin(timeMs * 0.018) * 0.16;
    const width = geometry.widthPx;
    const sideWidth = width * 2.4;
    const direction = normalizeScreenVector({
      x: geometry.end.x - geometry.start.x,
      y: geometry.end.y - geometry.start.y,
    });
    const normal = { x: -direction.y, y: direction.x };

    graphics.fillStyle(0x1fdcff, 0.08 * pulse);
    graphics.beginPath();
    graphics.moveTo(geometry.start.x - normal.x * width * 1.4, geometry.start.y - normal.y * width * 1.4);
    graphics.lineTo(geometry.end.x - normal.x * sideWidth, geometry.end.y - normal.y * sideWidth);
    graphics.lineTo(geometry.end.x + normal.x * sideWidth, geometry.end.y + normal.y * sideWidth);
    graphics.lineTo(geometry.start.x + normal.x * width * 1.4, geometry.start.y + normal.y * width * 1.4);
    graphics.closePath();
    graphics.fillPath();

    this.strokePilotToolEffectLine(geometry.start, geometry.end, width * 9, 0x0b6f9e, 0.18 * pulse);
    this.strokePilotToolEffectLine(geometry.start, geometry.end, width * 5.5, 0x20d8ff, 0.28 * pulse);
    this.strokePilotToolEffectLine(geometry.start, geometry.end, width * 2.2, 0x8cf8ff, 0.58);
    this.strokePilotToolEffectLine(geometry.start, geometry.end, width * 0.65, 0xffffff, 0.95);

    for (let index = 0; index < 10; index += 1) {
      const progress = getDrillBeamIntakeProgress(timeMs, index);
      const center = {
        x: lerp(geometry.start.x, geometry.end.x, progress),
        y: lerp(geometry.start.y, geometry.end.y, progress),
      };
      const shimmer = Math.sin(timeMs * 0.01 + index * 2.17);
      const offset = shimmer * sideWidth * (0.35 + progress * 0.85);
      const segmentLength = clamp(geometry.lengthPx * 0.045, 12, 46);
      this.strokePilotToolEffectLine(
        { x: center.x + normal.x * offset, y: center.y + normal.y * offset },
        {
          x: center.x - direction.x * segmentLength + normal.x * shimmer * width * 0.5,
          y: center.y - direction.y * segmentLength + normal.y * shimmer * width * 0.5,
        },
        width * 0.75,
        index % 2 === 0 ? 0xffffff : 0x69f0ff,
        0.32,
      );
    }

    graphics.fillStyle(0x26e6ff, 0.2);
    graphics.fillCircle(geometry.end.x, geometry.end.y, sideWidth * 2.2 * pulse);
    graphics.fillStyle(0xe9ffff, 0.62);
    graphics.fillCircle(geometry.end.x, geometry.end.y, width * 1.8);
    graphics.lineStyle(width * 0.55, 0x9ffcff, 0.72);
    graphics.strokeCircle(geometry.end.x, geometry.end.y, sideWidth * 1.35 * pulse);
  }

  // Находит активный бур в выбранной ячейке и готовит геометрию для эффекта.
  // Проводит одну светящуюся линию без накопления состояния между штрихами.
  private strokePilotToolEffectLine(
    start: DrillBeamGeometry["start"],
    end: DrillBeamGeometry["end"],
    width: number,
    color: number,
    alpha: number,
  ): void {
    this.pilotToolEffectGraphics.lineStyle(width, color, alpha);
    this.pilotToolEffectGraphics.beginPath();
    this.pilotToolEffectGraphics.moveTo(start.x, start.y);
    this.pilotToolEffectGraphics.lineTo(end.x, end.y);
    this.pilotToolEffectGraphics.strokePath();
  }

  // Обновляет тайловый фон так, чтобы движение камеры выглядело непрерывным.
  private renderBackground(camera: Parameters<typeof getPilotBackgroundTransform>[0]): void {
    const transform = getPilotBackgroundTransform(camera);

    this.background.setPosition(transform.position.x, transform.position.y);
    this.background.setSize(transform.size, transform.size);
    this.background.setRotation(transform.rotation);
    this.background.setScale(transform.scale);
    this.background.setTileScale(transform.tileScale, transform.tileScale);
    this.background.tilePositionX = transform.tilePositionX;
    this.background.tilePositionY = transform.tilePositionY;
  }

  // Переиспользует спрайты между снимками, чтобы не создавать их каждый кадр.
  private getOrCreateObjectSprite(object: CosmicObject): Phaser.GameObjects.Image | null {
    const textureKey = this.textureKeyForObject(object);
    if (!textureKey) {
      return this.objectSprites.get(object.ID) ?? null;
    }

    const existing = this.objectSprites.get(object.ID);
    if (existing) {
      if (existing.texture.key !== textureKey) {
        existing.setTexture(textureKey);
      }
      return existing;
    }

    const sprite = this.add.image(0, 0, textureKey).setOrigin(0.5);
    this.objectSprites.set(object.ID, sprite);

    return sprite;
  }

  // Загружает справочники до подключения к WebSocket, чтобы клиент не держал локальные таблицы.
  private async loadStartupData(): Promise<void> {
    try {
      this.referenceData = await fetchReferenceData();
      this.gameClient = new GameClient();
    } catch (error) {
      console.error(error);
      this.startupErrorMessage = "Не удалось загрузить справочники с сервера";
    }
  }

  // Возвращает модель объекта из серверного справочника.
  private modelForObject(object: CosmicObject): CosmicObjectModelReference | null {
    return this.referenceData?.CosmicObjectModel.Items[String(object.CosmicObjectModelID)] ?? null;
  }

  // Возвращает ключ текстуры модели и запускает загрузку, если Phaser еще не получил файл.
  // Проверяет, что объект является серверным лучом простого бура.
  private isDrillRayObject(object: CosmicObject): boolean {
    const model = this.modelForObject(object);
    if (!model || model.Acronym !== SIMPLE_DRILL_RAY_ACRONYM) {
      return false;
    }
    const objectType = this.referenceData?.CosmicObjectType.Items[String(model.CosmicObjectTypeID)] as { Acronym?: unknown } | undefined;
    return objectType?.Acronym === "Ray";
  }

  // Ставит точку привязки спрайта на физический центр модели из справочника.
  private updateObjectSpriteOrigin(sprite: Phaser.GameObjects.Image, object: CosmicObject): void {
    const model = this.modelForObject(object);
    if (!model || model.TextureWidth <= 0 || model.TextureHeight <= 0) {
      sprite.setOrigin(0.5);
      return;
    }

    sprite.setOrigin(model.TextureBodyOriginX / model.TextureWidth, model.TextureBodyOriginY / model.TextureHeight);
  }

  private textureKeyForObject(object: CosmicObject): string | null {
    const model = this.modelForObject(object);
    if (!model) {
      return null;
    }

    const textureKey = `world.cosmicObjectModel.${model.ID}`;
    if (!this.textures.exists(textureKey)) {
      this.loadTextureForModel(model, textureKey);
      return null;
    }
    return textureKey;
  }

  // Подключает файл текстуры из public/assets по пути, полученному с сервера.
  private loadTextureForModel(model: CosmicObjectModelReference, textureKey: string): void {
    if (this.loadingTextureKeys.has(textureKey)) {
      return;
    }

    this.loadingTextureKeys.add(textureKey);
    this.load.image(textureKey, this.texturePathForModel(model));
    this.load.once(Phaser.Loader.Events.COMPLETE, () => this.loadingTextureKeys.delete(textureKey));
    if (!this.load.isLoading()) {
      this.load.start();
    }
  }

  // Переводит серверный путь данных в URL ассета, отдаваемый Vite из client/public.
  private texturePathForModel(model: CosmicObjectModelReference): string {
    return `/assets/world/cosmic-objects/${model.TextureFilePath.replace(/^assets\//, "")}`;
  }

  // Возвращает масштаб текстуры из серверной модели.
  private textureScaleForObject(object: CosmicObject): number {
    return this.modelForObject(object)?.TextureScale ?? 4;
  }

  // Собирает реальные DOM-границы UI Kit, чтобы игровой курсор работал без pointer events.
  private collectGameUiControls(): GameUiControlState[] {
    return Array.from(document.querySelectorAll<HTMLElement>(".ui-kit-control"))
      .map((element, index): GameUiControlState | null => {
        const rect = getUiKitControlHitRect(element);
        if (rect.width <= 0 || rect.height <= 0) {
          return null;
        }
        const clippingViewport = element.closest(".ui-kit-dropdown__menu-viewport, .settings-input-table, .ui-kit-list");
        if (clippingViewport) {
          const viewportRect = clippingViewport.getBoundingClientRect();
          if (rect.bottom < viewportRect.top || rect.top > viewportRect.bottom || rect.right < viewportRect.left || rect.left > viewportRect.right) {
            return null;
          }
        }
        return {
          id: element.id || `ui-kit-control-${index}`,
          kind: uiKind(element.dataset.uiKind),
          rect: { left: rect.left, top: rect.top, width: rect.width, height: rect.height },
          zIndex: Number(element.dataset.uiZIndex ?? index),
          disabled: element.classList.contains("is-disabled"),
          visible: true,
          focusable: element.dataset.uiFocusable !== "false",
          value: element.dataset.uiValue ?? null,
        };
      })
      .filter((control): control is GameUiControlState => control !== null);
  }

  // Обновляет hit-test DOM-контролов только после изменения раскладки UI, а не на каждом кадре.
  private updateGameUiControlsIfNeeded(state: GameUiState): void {
    const signature = getGameUiControlLayoutSignature(state, {
      width: window.innerWidth,
      height: window.innerHeight,
      scaleWidth: this.scale.width,
      scaleHeight: this.scale.height,
    });
    if (signature !== this.lastGameUiControlLayoutSignature) {
      this.lastGameUiControlLayoutSignature = signature;
      this.gameUiControlRefreshFrames = 2;
    }
    if (this.gameUiControlRefreshFrames <= 0) {
      return;
    }

    this.gameUiControlRefreshFrames--;
    this.inputController.updateGameUiControls(this.collectGameUiControls());
  }

  // Снимает подтвержденные или отклоненные сервером pending-изменения панели управления.
  private syncControlPanelPendingFromServer(snapshot: { clientMutationAck?: { sessionId: string; lastAppliedSeq: number } } | null): void {
    this.controlPanelPending = pruneControlPanelPending(this.controlPanelPending, snapshot?.clientMutationAck);
    const errorSeq = this.gameClient?.getLatestControlPanelErrorSeq() ?? 0;
    if (errorSeq === this.controlPanelErrorSeq) {
      return;
    }
    this.controlPanelErrorSeq = errorSeq;
    const error = this.gameClient?.getLatestControlPanelError();
    if (!error) {
      return;
    }
    this.controlPanelPending = rejectControlPanelPending(this.controlPanelPending, {
      clientSessionId: error.clientSessionId,
      mutationSeq: error.mutationSeq,
    });
  }

  // Синхронизирует реальный выбор использования с первым доступным значением, которое показывает UI.
  private syncControlPanelUsageSelection(objectId: number | null, objects: CosmicObject[], equipmentGroups: EquipmentGroup[]): void {
    const selection = normalizeControlPanelUsageSelection({
      objectId,
      objects,
      equipmentGroups,
      referenceData: this.referenceData,
      selection: {
        leftContainerObjectId: this.selectedControlPanelUsageLeftObjectId,
        leftContainerGroupId: this.selectedControlPanelUsageLeftContainerGroupId,
        rightEquipmentObjectId: this.selectedControlPanelUsageRightObjectId,
        rightEquipmentGroupId: this.selectedControlPanelUsageRightEquipmentGroupId,
        constructorMaterialObjectId: this.selectedControlPanelConstructorMaterialObjectId,
        constructorMaterialGroupId: this.selectedControlPanelConstructorMaterialContainerGroupId,
        constructorProductObjectId: this.selectedControlPanelConstructorProductObjectId,
        constructorProductGroupId: this.selectedControlPanelConstructorProductContainerGroupId,
      },
    });

    const rightEquipmentGroupId = selection.rightEquipmentGroupId;
    const relatedLeftContainerGroupId = this.relatedEquipmentGroupId(rightEquipmentGroupId, "Opposite");
    const relatedSourceContainerGroupId = this.relatedEquipmentGroupId(rightEquipmentGroupId, "Source");
    const relatedDestinationContainerGroupId = this.relatedEquipmentGroupId(rightEquipmentGroupId, "Destination");
    const relatedSelection = applyActiveControlPanelUsageRelations({
      selection,
      rightEquipmentActive: this.getControlPanelEquipmentGroup(rightEquipmentGroupId ?? 0)?.Active ?? false,
      relatedOppositeGroupId: relatedLeftContainerGroupId,
      relatedSourceGroupId: relatedSourceContainerGroupId,
      relatedDestinationGroupId: relatedDestinationContainerGroupId,
      groupObjectId: (groupId) => this.getControlPanelEquipmentGroup(groupId)?.CosmicObjectID ?? null,
    });
    this.selectedControlPanelUsageLeftObjectId = relatedSelection.leftContainerObjectId;
    this.selectedControlPanelUsageLeftContainerGroupId = relatedSelection.leftContainerGroupId;
    this.selectedControlPanelUsageRightObjectId = relatedSelection.rightEquipmentObjectId;
    this.selectedControlPanelUsageRightEquipmentGroupId = relatedSelection.rightEquipmentGroupId;
    this.selectedControlPanelConstructorMaterialObjectId = relatedSelection.constructorMaterialObjectId;
    this.selectedControlPanelConstructorMaterialContainerGroupId = relatedSelection.constructorMaterialGroupId;
    this.selectedControlPanelConstructorProductObjectId = relatedSelection.constructorProductObjectId;
    this.selectedControlPanelConstructorProductContainerGroupId = relatedSelection.constructorProductGroupId;
    this.selectedControlPanelConstructorMaterialContainerGroupId = normalizeSelectedControlPanelGroupId(
      equipmentGroups.filter((group) => group.CosmicObjectID === this.selectedControlPanelConstructorMaterialObjectId && this.isEquipmentGroupItemType(group, "Container")),
      this.selectedControlPanelConstructorMaterialContainerGroupId,
    );
    this.selectedControlPanelConstructorProductContainerGroupId = normalizeSelectedControlPanelGroupId(
      equipmentGroups.filter((group) => group.CosmicObjectID === this.selectedControlPanelConstructorProductObjectId && this.isEquipmentGroupItemType(group, "Container")),
      this.selectedControlPanelConstructorProductContainerGroupId,
    );
  }

  // Возвращает количество топлива, доступное для залива из текущего выбора в левом контейнере.
  private getControlPanelFuelFillMaxAmount(objects: CosmicObject[], equipmentGroups: EquipmentGroup[], itemGroups: GameUiState["itemGroups"]): number {
    const fuelTankGroup = equipmentGroups.find((group) => group.ID === this.selectedControlPanelUsageRightEquipmentGroupId) ?? null;
    return getControlPanelFuelFillMaxAmount({
      object: objects.find((object) => object.ID === fuelTankGroup?.CosmicObjectID) ?? null,
      fuelTankGroup,
      itemGroups,
      selectedItemGroupIds: this.selectedControlPanelUsageLeftItemGroupIds,
      referenceData: this.referenceData,
    });
  }

  // Отправляет завершенное редактирование названия объекта, если текст отличается от серверного снимка.
  // Возвращает связанную группу оборудования по сохранённому виду связи.
  private relatedEquipmentGroupId(equipmentGroupId: number | null, relationTypeAcronym: "Source" | "Destination" | "Opposite"): number | null {
    if (!equipmentGroupId) {
      return null;
    }
    const draft = this.controlPanelEquipmentGroupRelationDrafts[this.equipmentGroupRelationDraftKey(equipmentGroupId, relationTypeAcronym)];
    if (draft) {
      return draft;
    }
    const group = this.getControlPanelEquipmentGroup(equipmentGroupId);
    if (relationTypeAcronym === "Source") {
      return group?.SourceEquipmentGroupID || null;
    }
    if (relationTypeAcronym === "Destination") {
      return group?.DestinationEquipmentGroupID || null;
    }
    return group?.OppositeEquipmentGroupID || null;
  }

  // Сохраняет выбор контейнера для текущей правой группы оборудования.
  private saveControlPanelUsageRelatedContainer(relatedEquipmentGroupId: number, relationTypeAcronym: "Source" | "Destination" | "Opposite"): void {
    if (!this.selectedControlPanelUsageRightEquipmentGroupId) {
      return;
    }
    this.controlPanelEquipmentGroupRelationDrafts[this.equipmentGroupRelationDraftKey(this.selectedControlPanelUsageRightEquipmentGroupId, relationTypeAcronym)] = relatedEquipmentGroupId;
    this.gameClient?.sendControlPanelEquipmentGroupRelationUpdate({
      equipmentGroupId: this.selectedControlPanelUsageRightEquipmentGroupId,
      relationTypeAcronym,
      relatedEquipmentGroupId,
    });
  }

  // Собирает ключ черновика связи из группы и вида связи.
  private equipmentGroupRelationDraftKey(equipmentGroupId: number, relationTypeAcronym: "Source" | "Destination" | "Opposite"): string {
    return `${equipmentGroupId}:${relationTypeAcronym}`;
  }

  private commitControlPanelObjectTitleIfNeeded(serverSelfObject: CosmicObject | null): void {
    const title = this.inputController.consumeControlPanelObjectTitleCommit();
    if (title === null || !serverSelfObject || title === serverSelfObject.Title) {
      return;
    }
    const mutation = this.gameClient?.sendControlPanelObjectUpdate({ title });
    if (!mutation) {
      return;
    }
    this.controlPanelPending = {
      ...this.controlPanelPending,
      object: {
        ...this.controlPanelPending.object,
        title: { ...mutation, value: title },
      },
    };
  }

  // Отправляет завершенное редактирование названия группы, если оно отличается от показанного снимка.
  private commitControlPanelEquipmentTitleIfNeeded(group: EquipmentGroup | null): void {
    const title = this.inputController.consumeControlPanelEquipmentTitleCommit();
    if (title === null || !group || title === group.Title) {
      return;
    }
    this.sendControlPanelEquipmentMutation(group.ID, { title });
  }

  // Применяет накопленные действия общего UI к локальной витрине контролов.
  private consumeGameUiActions(): void {
    let action = this.inputController.consumeGameUiAction();
    while (action) {
      if (this.consumeLandingTargetUiAction(action)) {
        action = this.inputController.consumeGameUiAction();
        continue;
      }
      if (this.consumeExchangeUiAction(action)) {
        action = this.inputController.consumeGameUiAction();
        continue;
      }
      if (this.inputController.isSettingsVisible() && this.consumeSettingsUiAction(action)) {
        action = this.inputController.consumeGameUiAction();
        continue;
      }
      if (this.inputController.isControlPanelVisible() && this.consumeControlPanelUiAction(action)) {
        action = this.inputController.consumeGameUiAction();
        continue;
      }
      this.uiKitDemoState = applyUiKitDemoAction(this.uiKitDemoState, action);
      action = this.inputController.consumeGameUiAction();
    }
  }

  // Обрабатывает окно выбора объекта назначения для пересадки.
  private consumeLandingTargetUiAction(action: GameUiAction): boolean {
    if (this.landingTargetObjectIds.length === 0 || action.type !== "click") {
      return false;
    }
    if (action.controlId === "landing-target-modal-close-button") {
      this.closeLandingTargetModal();
      return true;
    }
    if (action.controlId.startsWith("landing-target-list-") && typeof action.value === "string") {
      const targetID = Number(action.value);
      if (this.landingTargetObjectIds.includes(targetID)) {
        this.selectedLandingTargetObjectId = targetID;
      }
      return true;
    }
    if (action.controlId === "landing-target-send-button") {
      const targetID = this.selectedLandingTargetObjectId ?? this.landingTargetObjectIds[0] ?? 0;
      if (targetID > 0) {
        this.gameClient?.sendLandingRequest(targetID);
      }
      this.closeLandingTargetModal();
      return true;
    }
    return action.controlId.startsWith("landing-target-");
  }

  // Закрывает окно выбора назначения и сбрасывает локальный выбор.
  private closeLandingTargetModal(): void {
    this.landingTargetObjectIds = [];
    this.selectedLandingTargetObjectId = null;
  }

  // Отправляет накопленные клавиатурные команды стыковки отдельными сетевыми пакетами.
  private consumeDockingActions(): void {
    let action = this.inputController.consumeDockingAction();
    while (action) {
      if (action === "request") {
        this.gameClient?.sendDockingRequest();
      } else if (action === "approve") {
        if (this.exchangeRequestRole === "receiver") {
          this.gameClient?.sendExchangeApprove();
        } else if (this.dockingWindow?.kind === "landingRequest") {
          this.gameClient?.sendLandingApprove();
        } else {
          this.gameClient?.sendDockingApprove();
        }
      } else if (action === "reject") {
        if (this.exchangeRequestRole === "receiver") {
          this.gameClient?.sendExchangeReject();
        } else if (this.dockingWindow?.kind === "landingRequest") {
          this.gameClient?.sendLandingReject();
        } else {
          this.gameClient?.sendDockingReject();
        }
      } else if (action === "landing") {
        this.gameClient?.sendLandingBegin();
      } else if (action === "exchange") {
        this.gameClient?.sendExchangeRequest();
      } else {
        this.gameClient?.sendDockingUndock();
      }
      action = this.inputController.consumeDockingAction();
    }
  }

  // Обновляет окно и уведомления по последнему событию сервера.
  private syncDockingEventFromServer(nowMs: number): void {
    for (const event of this.gameClient?.consumeDockingEvents() ?? []) {
      this.applyDockingEvent(event, nowMs);
    }
  }

  // Применяет одно событие стыковки к локальному HUD-состоянию.
  private applyDockingEvent(event: DockingEventMessage, nowMs: number): void {
    if (event.kind === "dockingNotification") {
      if (event.message) {
        this.dockingNotifications = [
          ...this.dockingNotifications,
          { id: this.nextDockingNotificationID++, message: event.message, expiresAtMs: nowMs + DOCKING_NOTIFICATION_DURATION_MS },
        ];
      }
      return;
    }
    if (event.kind === "dockingFinished" || event.kind === "landingFinished") {
      this.dockingWindow = null;
      return;
    }
    if (event.kind === "landingTargetSelection") {
      this.landingTargetObjectIds = event.targetIds ?? [];
      this.selectedLandingTargetObjectId = this.landingTargetObjectIds[0] ?? null;
      return;
    }
    if (!event.role) {
      return;
    }
    this.dockingWindow = {
      kind: event.kind === "landingRequestStarted" ? "landingRequest" : event.kind === "dockingRequestStarted" ? "request" : "process",
      role: event.role,
      startedAtMs: nowMs,
      durationMs: Math.max(0, (event.duration ?? 0) * 1000),
    };
  }

  // Обновляет окно обмена по событиям сервера.
  private syncExchangeEventFromServer(nowMs: number): void {
    for (const event of this.gameClient?.consumeExchangeEvents() ?? []) {
      this.applyExchangeEvent(event, nowMs);
    }
  }

  // Применяет одно событие обмена к локальному состоянию HUD.
  private applyExchangeEvent(event: ExchangeEventMessage, nowMs: number): void {
    if (event.kind === "exchangeRequestStarted") {
      this.exchangeRequestRole = event.role ?? null;
      if (event.role) {
        this.dockingWindow = {
          kind: "exchangeRequest",
          role: event.role,
          startedAtMs: nowMs,
          durationMs: getExchangeRequestDurationMs(event.duration),
        };
      }
      return;
    }
    if (event.kind === "exchangeState" && event.state) {
      this.exchangeRequestRole = null;
      this.dockingWindow = null;
      this.exchangeState = event.state;
      this.syncExchangeSelectionFromState(event.state);
      return;
    }
    if (event.kind === "exchangeNotification" && event.message) {
      this.dockingNotifications = [
        ...this.dockingNotifications,
        { id: this.nextDockingNotificationID++, message: event.message, expiresAtMs: this.time.now + DOCKING_NOTIFICATION_DURATION_MS },
      ];
      return;
    }
    if (event.kind === "exchangeRejected" || event.kind === "exchangeCancelled" || event.kind === "exchangeFinished") {
      this.exchangeRequestRole = null;
      this.dockingWindow = null;
      this.exchangeState = null;
      this.openExchangeSelect = null;
      this.selectedExchangeSourceItemGroupIds = [];
      this.selectedExchangeSourceAnchorItemGroupId = null;
    }
  }

  // Синхронизирует выбранные выпадающие списки с серверным состоянием.
  private syncExchangeSelectionFromState(state: ExchangeStateMessage): void {
    this.selectedExchangeReceiverObjectId = this.objectIDForEquipmentGroup(state.selfReceiverContainerEquipmentGroupId) ?? state.selfObjectId;
    this.selectedExchangeSourceObjectId = this.objectIDForEquipmentGroup(state.selfSourceContainerEquipmentGroupId) ?? state.selfObjectId;
  }

  // Возвращает объект, на котором стоит группа оборудования.
  private objectIDForEquipmentGroup(groupId: number): number | null {
    return this.gameUi.state().equipmentGroups.find((group) => group.ID === groupId)?.CosmicObjectID ?? null;
  }

  // Применяет колесо мыши к обычным спискам общего UI.
  private consumeGameUiWheelActions(): void {
    let action = this.inputController.consumeGameUiWheelAction();
    while (action) {
      const maxOffsetPx = this.listMaxScrollOffset(action.controlId);
      if (maxOffsetPx > 0) {
        const currentOffsetPx = this.listScrollOffsets.get(action.controlId) ?? 0;
        this.listScrollOffsets.set(action.controlId, clamp(currentOffsetPx + action.deltaY, 0, maxOffsetPx));
      }
      action = this.inputController.consumeGameUiWheelAction();
    }
  }

  // Применяет действия панели управления и не пропускает их в отладочную витрину.
  private consumeControlPanelUiAction(action: GameUiAction): boolean {
    if (action.kind === "scrollbar" && action.controlId.endsWith("-scrollbar") && this.consumeListScrollbarAction(action)) {
      return true;
    }
    if (action.kind === "scrollbar" && action.controlId === "control-panel-equipment-list-scrollbar") {
      return this.consumeControlPanelEquipmentListScrollbarAction(action);
    }
    if (action.kind === "slider" && action.controlId === "control-panel-equipment-enabled-slider") {
      return this.consumeControlPanelEquipmentSliderAction(action);
    }
    if (action.kind === "slider" && action.controlId === "control-panel-fuel-drain-amount-slider") {
      this.updateControlPanelFuelDrainAmountFromSlider(action);
      return true;
    }
    if (action.type !== "click") {
      return action.controlId.startsWith("control-panel-");
    }
    if (action.type === "click" && action.controlId.endsWith("-outside-blocker")) {
      this.openControlPanelUsageSelect = null;
      return true;
    }
    if (action.type === "click" && action.controlId.startsWith("control-panel-tab-")) {
      if (isControlPanelTabValue(action.value)) {
        this.selectedControlPanelTab = action.value;
      }
      return true;
    }
    if (action.type === "click" && action.controlId.startsWith("control-panel-equipment-tab-")) {
      if (isControlPanelEquipmentSubTabValue(action.value)) {
        this.selectedControlPanelEquipmentTab = action.value;
        this.openControlPanelUsageSelect = null;
      }
      return true;
    }
    if (action.type === "click" && action.controlId === "control-panel-equipment-usage-button") {
      this.selectedControlPanelEquipmentTab = "usage";
      this.selectedControlPanelUsageRightEquipmentGroupId = this.getSelectedControlPanelEquipmentGroup()?.ID ?? null;
      this.openControlPanelUsageSelect = null;
      return true;
    }
    if (action.controlId === "control-panel-object-enabled") {
      const enabled = !this.gameUi.state().controlPanelObjectEnabled;
      const mutation = this.gameClient?.sendControlPanelObjectUpdate({ enabled });
      if (mutation) {
        this.controlPanelPending = {
          ...this.controlPanelPending,
          object: {
            ...this.controlPanelPending.object,
            enabled: { ...mutation, value: enabled },
          },
        };
      }
      return true;
    }
    if (action.controlId.startsWith("control-panel-equipment-list-") && typeof action.value === "string") {
      const groupId = Number(action.value);
      if (this.getControlPanelEquipmentGroup(groupId)) {
        this.selectedControlPanelEquipmentGroupId = groupId;
        this.openControlPanelUsageSelect = null;
        this.controlPanelEquipmentListScrollOffsetPx = clamp(this.controlPanelEquipmentListScrollOffsetPx, 0, this.controlPanelEquipmentListMaxScrollOffset());
      }
      return true;
    }
    if (action.controlId === "control-panel-usage-left-container-select") {
      this.openControlPanelUsageSelect = this.openControlPanelUsageSelect === "left" ? null : "left";
      return true;
    }
    if (action.controlId === "control-panel-usage-left-object-select") {
      this.openControlPanelUsageSelect = this.openControlPanelUsageSelect === "leftObject" ? null : "leftObject";
      return true;
    }
    if (action.controlId === "control-panel-usage-right-equipment-select") {
      this.openControlPanelUsageSelect = this.openControlPanelUsageSelect === "right" ? null : "right";
      return true;
    }
    if (action.controlId === "control-panel-usage-right-object-select") {
      this.openControlPanelUsageSelect = this.openControlPanelUsageSelect === "rightObject" ? null : "rightObject";
      return true;
    }
    if (action.controlId === "control-panel-constructor-material-select") {
      this.openControlPanelUsageSelect = this.openControlPanelUsageSelect === "constructorMaterials" ? null : "constructorMaterials";
      return true;
    }
    if (action.controlId === "control-panel-constructor-material-object-select") {
      this.openControlPanelUsageSelect = this.openControlPanelUsageSelect === "constructorMaterialObject" ? null : "constructorMaterialObject";
      return true;
    }
    if (action.controlId === "control-panel-constructor-product-select") {
      this.openControlPanelUsageSelect = this.openControlPanelUsageSelect === "constructorProducts" ? null : "constructorProducts";
      return true;
    }
    if (action.controlId === "control-panel-constructor-product-object-select") {
      this.openControlPanelUsageSelect = this.openControlPanelUsageSelect === "constructorProductObject" ? null : "constructorProductObject";
      return true;
    }
    if (action.controlId.startsWith("control-panel-usage-left-object-select-") && typeof action.value === "string") {
      this.selectedControlPanelUsageLeftObjectId = Number(action.value);
      this.selectedControlPanelUsageLeftContainerGroupId = this.firstControlPanelGroupIdOnObject(this.selectedControlPanelUsageLeftObjectId, "Container");
      this.openControlPanelUsageSelect = null;
      return true;
    }
    if (action.controlId.startsWith("control-panel-usage-left-container-select-") && typeof action.value === "string") {
      const groupId = Number(action.value);
      if (this.getControlPanelEquipmentGroup(groupId)) {
        this.selectedControlPanelUsageLeftContainerGroupId = groupId;
        this.selectedControlPanelUsageLeftObjectId = this.getControlPanelEquipmentGroup(groupId)?.CosmicObjectID ?? this.selectedControlPanelUsageLeftObjectId;
        this.saveControlPanelUsageRelatedContainer(groupId, "Opposite");
      }
      this.openControlPanelUsageSelect = null;
      return true;
    }
    if (action.controlId.startsWith("control-panel-usage-right-object-select-") && typeof action.value === "string") {
      this.selectedControlPanelUsageRightObjectId = Number(action.value);
      this.selectedControlPanelUsageRightEquipmentGroupId = this.firstControlPanelInternalGroupIdOnObject(this.selectedControlPanelUsageRightObjectId);
      this.openControlPanelUsageSelect = null;
      return true;
    }
    if (action.controlId.startsWith("control-panel-usage-right-equipment-select-") && typeof action.value === "string") {
      const groupId = Number(action.value);
      if (this.getControlPanelEquipmentGroup(groupId)) {
        this.selectedControlPanelUsageRightEquipmentGroupId = groupId;
        this.selectedControlPanelUsageRightObjectId = this.getControlPanelEquipmentGroup(groupId)?.CosmicObjectID ?? this.selectedControlPanelUsageRightObjectId;
      }
      this.openControlPanelUsageSelect = null;
      return true;
    }
    if (action.controlId.startsWith("control-panel-constructor-material-object-select-") && typeof action.value === "string") {
      this.selectedControlPanelConstructorMaterialObjectId = Number(action.value);
      this.selectedControlPanelConstructorMaterialContainerGroupId = this.firstControlPanelGroupIdOnObject(this.selectedControlPanelConstructorMaterialObjectId, "Container");
      this.openControlPanelUsageSelect = null;
      return true;
    }
    if (action.controlId.startsWith("control-panel-constructor-material-select-") && typeof action.value === "string") {
      const groupId = Number(action.value);
      if (this.getControlPanelEquipmentGroup(groupId)) {
        this.selectedControlPanelConstructorMaterialContainerGroupId = groupId;
        this.selectedControlPanelConstructorMaterialObjectId = this.getControlPanelEquipmentGroup(groupId)?.CosmicObjectID ?? this.selectedControlPanelConstructorMaterialObjectId;
        this.saveControlPanelUsageRelatedContainer(groupId, "Source");
      }
      this.openControlPanelUsageSelect = null;
      return true;
    }
    if (action.controlId.startsWith("control-panel-constructor-product-object-select-") && typeof action.value === "string") {
      this.selectedControlPanelConstructorProductObjectId = Number(action.value);
      this.selectedControlPanelConstructorProductContainerGroupId = this.firstControlPanelGroupIdOnObject(this.selectedControlPanelConstructorProductObjectId, "Container");
      this.openControlPanelUsageSelect = null;
      return true;
    }
    if (action.controlId.startsWith("control-panel-constructor-product-select-") && typeof action.value === "string") {
      const groupId = Number(action.value);
      if (this.getControlPanelEquipmentGroup(groupId)) {
        this.selectedControlPanelConstructorProductContainerGroupId = groupId;
        this.selectedControlPanelConstructorProductObjectId = this.getControlPanelEquipmentGroup(groupId)?.CosmicObjectID ?? this.selectedControlPanelConstructorProductObjectId;
        this.saveControlPanelUsageRelatedContainer(groupId, "Destination");
      }
      this.openControlPanelUsageSelect = null;
      return true;
    }
    if (action.controlId.startsWith("control-panel-constructor-tab-")) {
      if (isControlPanelConstructorTabValue(action.value)) {
        this.selectedControlPanelConstructorTab = action.value;
      }
      return true;
    }
    if (action.controlId.startsWith("control-panel-constructor-schema-list-") && typeof action.value === "string") {
      this.selectedControlPanelConstructorSchemaId = Number(action.value);
      return true;
    }
    if (action.controlId.startsWith("control-panel-constructor-blueprint-list-") && typeof action.value === "string") {
      this.selectedControlPanelConstructorBlueprintId = Number(action.value);
      return true;
    }
    if (action.controlId === "control-panel-constructor-make-button") {
      this.startControlPanelConstructorProduceItem();
      return true;
    }
    if (action.controlId === "control-panel-deconstructor-make-button") {
      this.startControlPanelItemDeconstruction();
      return true;
    }
    if (action.controlId.startsWith("control-panel-constructor-main-queue-") && typeof action.value === "string") {
      this.selectedControlPanelConstructorMainJobId = Number(action.value);
      return true;
    }
    if (action.controlId.startsWith("control-panel-container-queue-") && typeof action.value === "string") {
      this.selectedControlPanelConstructorMainJobId = Number(action.value);
      return true;
    }
    if (action.controlId.startsWith("control-panel-fuel-queue-") && typeof action.value === "string") {
      this.selectedControlPanelConstructorMainJobId = Number(action.value);
      return true;
    }
    if (action.controlId.startsWith("control-panel-deconstructor-main-queue-") && typeof action.value === "string") {
      this.selectedControlPanelConstructorMainJobId = Number(action.value);
      return true;
    }
    if (action.controlId === "control-panel-constructor-skip-next") {
      this.sendControlPanelConstructorQueueCommand("skipNext");
      return true;
    }
    if (action.controlId === "control-panel-constructor-skip-all-next") {
      this.sendControlPanelConstructorQueueCommand("skipAllNext");
      return true;
    }
    if (action.controlId === "control-panel-constructor-cancel") {
      this.sendControlPanelConstructorQueueCommand("cancel");
      return true;
    }
    if (action.controlId === "control-panel-constructor-cancel-all") {
      this.sendControlPanelConstructorQueueCommand("cancelAll");
      return true;
    }
    if (action.controlId === "control-panel-container-skip-next") {
      this.sendControlPanelConstructorQueueCommand("skipNext");
      return true;
    }
    if (action.controlId === "control-panel-container-skip-all-next") {
      this.sendControlPanelConstructorQueueCommand("skipAllNext");
      return true;
    }
    if (action.controlId === "control-panel-container-cancel") {
      this.sendControlPanelConstructorQueueCommand("cancel");
      return true;
    }
    if (action.controlId === "control-panel-container-cancel-all") {
      this.sendControlPanelConstructorQueueCommand("cancelAll");
      return true;
    }
    if (action.controlId === "control-panel-fuel-skip-next") {
      this.sendControlPanelConstructorQueueCommand("skipNext");
      return true;
    }
    if (action.controlId === "control-panel-fuel-skip-all-next") {
      this.sendControlPanelConstructorQueueCommand("skipAllNext");
      return true;
    }
    if (action.controlId === "control-panel-fuel-cancel") {
      this.sendControlPanelConstructorQueueCommand("cancel");
      return true;
    }
    if (action.controlId === "control-panel-fuel-cancel-all") {
      this.sendControlPanelConstructorQueueCommand("cancelAll");
      return true;
    }
    if (action.controlId === "control-panel-deconstructor-skip-next") {
      this.sendControlPanelConstructorQueueCommand("skipNext");
      return true;
    }
    if (action.controlId === "control-panel-deconstructor-skip-all-next") {
      this.sendControlPanelConstructorQueueCommand("skipAllNext");
      return true;
    }
    if (action.controlId === "control-panel-deconstructor-cancel") {
      this.sendControlPanelConstructorQueueCommand("cancel");
      return true;
    }
    if (action.controlId === "control-panel-deconstructor-cancel-all") {
      this.sendControlPanelConstructorQueueCommand("cancelAll");
      return true;
    }
    if (action.controlId === "control-panel-constructor-produce-cancel") {
      this.controlPanelConstructorProduceDialogOpen = false;
      this.controlPanelItemDeconstructionDialogOpen = false;
      this.inputController.blurControlPanelFuelDrainAmount();
      return true;
    }
    if (action.controlId === "control-panel-constructor-produce-ok") {
      this.controlPanelFuelDrainAmount = clamp(this.inputController.getControlPanelFuelDrainAmount(this.controlPanelFuelDrainAmount), 1, this.controlPanelConstructorProduceMaxAmount);
      this.sendControlPanelConstructorProduceItem(this.controlPanelFuelDrainAmount);
      this.controlPanelConstructorProduceDialogOpen = false;
      this.controlPanelItemDeconstructionDialogOpen = false;
      this.inputController.blurControlPanelFuelDrainAmount();
      return true;
    }
    if (action.controlId === "control-panel-item-deconstruction-cancel") {
      this.controlPanelItemDeconstructionDialogOpen = false;
      this.inputController.blurControlPanelFuelDrainAmount();
      return true;
    }
    if (action.controlId === "control-panel-item-deconstruction-ok") {
      this.controlPanelFuelDrainAmount = clamp(this.inputController.getControlPanelFuelDrainAmount(this.controlPanelFuelDrainAmount), 1, this.controlPanelItemDeconstructionMaxAmount);
      this.sendControlPanelItemDeconstruction(
        this.controlPanelItemDeconstructionDeconstructorGroupId,
        this.controlPanelItemDeconstructionSourceGroupId,
        this.controlPanelItemDeconstructionTargetGroupId,
        this.controlPanelItemDeconstructionItemGroupIds,
        this.controlPanelFuelDrainAmount,
      );
      this.controlPanelItemDeconstructionDialogOpen = false;
      this.inputController.blurControlPanelFuelDrainAmount();
      return true;
    }
    if (action.controlId.startsWith("control-panel-usage-left-container-content-") && typeof action.value === "string") {
      const selection = this.updateControlPanelUsageItemSelection(this.selectedControlPanelUsageLeftContainerGroupId, this.selectedControlPanelUsageLeftItemGroupIds, this.selectedControlPanelUsageLeftAnchorItemGroupId, Number(action.value), action);
      this.selectedControlPanelUsageLeftItemGroupIds = selection.selectedIds;
      this.selectedControlPanelUsageLeftAnchorItemGroupId = selection.anchorId;
      return true;
    }
    if (action.controlId.startsWith("control-panel-usage-right-container-content-") && typeof action.value === "string") {
      const selection = this.updateControlPanelUsageItemSelection(this.getControlPanelUsageRightContentContainerGroupId(), this.selectedControlPanelUsageRightItemGroupIds, this.selectedControlPanelUsageRightAnchorItemGroupId, Number(action.value), action);
      this.selectedControlPanelUsageRightItemGroupIds = selection.selectedIds;
      this.selectedControlPanelUsageRightAnchorItemGroupId = selection.anchorId;
      return true;
    }
    if (action.controlId === "control-panel-container-transfer-to-right") {
      this.startControlPanelContainerTransfer(this.selectedControlPanelUsageLeftContainerGroupId, this.selectedControlPanelUsageRightEquipmentGroupId, this.selectedControlPanelUsageRightEquipmentGroupId, true, this.selectedControlPanelUsageLeftItemGroupIds);
      return true;
    }
    if (action.controlId === "control-panel-container-transfer-to-left") {
      this.startControlPanelContainerTransfer(this.selectedControlPanelUsageRightEquipmentGroupId, this.selectedControlPanelUsageLeftContainerGroupId, this.selectedControlPanelUsageRightEquipmentGroupId, false, this.selectedControlPanelUsageRightItemGroupIds);
      return true;
    }
    if (action.controlId === "control-panel-container-transfer-cancel") {
      this.controlPanelContainerTransferDialogOpen = false;
      this.controlPanelItemDeconstructionDialogOpen = false;
      this.inputController.blurControlPanelFuelDrainAmount();
      return true;
    }
    if (action.controlId === "control-panel-container-transfer-ok") {
      this.controlPanelFuelDrainAmount = clamp(this.inputController.getControlPanelFuelDrainAmount(this.controlPanelFuelDrainAmount), 0, this.controlPanelContainerTransferMaxAmount);
      this.sendControlPanelContainerTransfer(this.controlPanelContainerTransferSourceGroupId, this.controlPanelContainerTransferTargetGroupId, this.controlPanelContainerTransferControllerGroupId, this.controlPanelContainerTransferLeftToRightDirection, this.controlPanelContainerTransferItemGroupIds, this.controlPanelFuelDrainAmount);
      this.controlPanelContainerTransferDialogOpen = false;
      this.controlPanelItemDeconstructionDialogOpen = false;
      this.inputController.blurControlPanelFuelDrainAmount();
      return true;
    }
    if (action.controlId === "control-panel-fuel-transfer-to-tank") {
      this.controlPanelFuelFillMaxAmount = this.getControlPanelFuelFillMaxAmount(this.gameUi.state().objects, this.gameUi.state().equipmentGroups, this.gameUi.state().itemGroups);
      if (this.controlPanelFuelFillMaxAmount > 0) {
        this.controlPanelFuelFillDialogOpen = true;
        this.controlPanelFuelDrainDialogOpen = false;
        this.controlPanelContainerTransferDialogOpen = false;
        this.controlPanelConstructorProduceDialogOpen = false;
        this.controlPanelItemDeconstructionDialogOpen = false;
        this.controlPanelFuelDrainAmount = this.controlPanelFuelFillMaxAmount;
        this.inputController.setControlPanelFuelDrainAmount(this.controlPanelFuelDrainAmount);
      }
      return true;
    }
    if (action.controlId === "control-panel-fuel-fill-cancel") {
      this.controlPanelFuelFillDialogOpen = false;
      this.controlPanelItemDeconstructionDialogOpen = false;
      this.inputController.blurControlPanelFuelDrainAmount();
      return true;
    }
    if (action.controlId === "control-panel-fuel-fill-ok") {
      this.controlPanelFuelDrainAmount = clamp(this.inputController.getControlPanelFuelDrainAmount(this.controlPanelFuelDrainAmount), 0, this.controlPanelFuelFillMaxAmount);
      this.sendControlPanelFuelTransfer(this.selectedControlPanelUsageLeftContainerGroupId, this.selectedControlPanelUsageRightEquipmentGroupId, this.selectedControlPanelUsageLeftItemGroupIds, this.controlPanelFuelDrainAmount);
      this.controlPanelFuelFillDialogOpen = false;
      this.controlPanelItemDeconstructionDialogOpen = false;
      this.inputController.blurControlPanelFuelDrainAmount();
      return true;
    }
    if (action.controlId === "control-panel-fuel-drain-open") {
      this.controlPanelFuelDrainDialogOpen = true;
      this.controlPanelFuelFillDialogOpen = false;
      this.controlPanelContainerTransferDialogOpen = false;
      this.controlPanelConstructorProduceDialogOpen = false;
      this.controlPanelItemDeconstructionDialogOpen = false;
      this.controlPanelFuelDrainAmount = Math.max(0, this.getSelectedFuelTankObject()?.Fuel ?? 0);
      this.inputController.setControlPanelFuelDrainAmount(this.controlPanelFuelDrainAmount);
      return true;
    }
    if (action.controlId === "control-panel-fuel-drain-cancel") {
      this.controlPanelFuelDrainDialogOpen = false;
      this.controlPanelItemDeconstructionDialogOpen = false;
      this.inputController.blurControlPanelFuelDrainAmount();
      return true;
    }
    if (action.controlId === "control-panel-fuel-drain-amount-decrement") {
      this.changeControlPanelFuelDrainAmount(-1);
      return true;
    }
    if (action.controlId === "control-panel-fuel-drain-amount" || action.controlId === "control-panel-fuel-drain-amount-increment") {
      this.changeControlPanelFuelDrainAmount(1);
      return true;
    }
    if (action.controlId === "control-panel-fuel-drain-ok") {
      this.controlPanelFuelDrainAmount = clamp(this.inputController.getControlPanelFuelDrainAmount(this.controlPanelFuelDrainAmount), 0, Math.max(0, this.getSelectedFuelTankObject()?.Fuel ?? 0));
      this.sendControlPanelFuelTransfer(this.selectedControlPanelUsageLeftContainerGroupId, this.selectedControlPanelUsageRightEquipmentGroupId, [], this.controlPanelFuelDrainAmount);
      this.controlPanelFuelDrainDialogOpen = false;
      this.controlPanelItemDeconstructionDialogOpen = false;
      this.inputController.blurControlPanelFuelDrainAmount();
      return true;
    }
    if (action.controlId === "control-panel-equipment-enabled") {
      const group = this.getSelectedControlPanelEquipmentGroup();
      if (group) {
        const enabled = !this.getControlPanelEquipmentEnabled(group);
        this.sendControlPanelEquipmentMutation(group.ID, { enabled });
      }
      return true;
    }
    if (action.controlId === "control-panel-equipment-enabled-count-decrement") {
      this.changeControlPanelEquipmentEnabledCount(-1);
      return true;
    }
    if (action.controlId === "control-panel-equipment-enabled-count" || action.controlId === "control-panel-equipment-enabled-count-increment") {
      this.changeControlPanelEquipmentEnabledCount(1);
      return true;
    }
    if (action.controlId === "control-panel-modal") {
      return true;
    }
    return action.controlId.startsWith("control-panel-");
  }

  // Обрабатывает полосу прокрутки любого обычного списка панели управления.
  private consumeListScrollbarAction(action: GameUiAction): boolean {
    if (!action.controlRect) {
      return false;
    }
    const listId = listIdFromScrollbarId(action.controlId);
    if (!listId || !this.isScrollableListId(listId)) {
      return false;
    }
    const scrollState = this.getListScrollState(listId);
    if (action.type === "dragStart") {
      this.listScrollbarDrags.set(listId, startScrollbarDrag({
        top: action.controlRect.top,
        height: action.controlRect.height,
        thumbTopPercent: scrollState.thumbTopPercent,
        thumbHeightPercent: scrollState.thumbHeightPercent,
      }, action.y));
      return true;
    }
    if (action.type === "dragEnd") {
      this.listScrollbarDrags.delete(listId);
      return true;
    }
    const drag = this.listScrollbarDrags.get(listId);
    if (action.type === "dragMove" && drag) {
      const thumbTopPercent = getScrollbarThumbTopPercentFromCursor({
        top: action.controlRect.top,
        height: action.controlRect.height,
        thumbHeightPercent: scrollState.thumbHeightPercent,
        drag,
      }, action.y);
      this.listScrollOffsets.set(listId, getScrollOffsetFromThumbTopPercent({
        thumbTopPercent,
        thumbHeightPercent: scrollState.thumbHeightPercent,
        maxOffsetPx: this.listMaxScrollOffset(listId),
        reverse: false,
      }));
      return true;
    }
    return true;
  }

  // Обрабатывает полосу прокрутки списка оборудования через общий механизм UI Kit.
  private consumeControlPanelEquipmentListScrollbarAction(action: GameUiAction): boolean {
    if (!action.controlRect) {
      return true;
    }
    const scrollState = this.getControlPanelEquipmentListScrollState();
    if (action.type === "dragStart") {
      this.controlPanelEquipmentListScrollbarDrag = startScrollbarDrag({
        top: action.controlRect.top,
        height: action.controlRect.height,
        thumbTopPercent: scrollState.thumbTopPercent,
        thumbHeightPercent: scrollState.thumbHeightPercent,
      }, action.y);
      return true;
    }
    if (action.type === "dragEnd") {
      this.controlPanelEquipmentListScrollbarDrag = null;
      return true;
    }
    if (action.type === "dragMove" && this.controlPanelEquipmentListScrollbarDrag) {
      const thumbTopPercent = getScrollbarThumbTopPercentFromCursor({
        top: action.controlRect.top,
        height: action.controlRect.height,
        thumbHeightPercent: scrollState.thumbHeightPercent,
        drag: this.controlPanelEquipmentListScrollbarDrag,
      }, action.y);
      this.controlPanelEquipmentListScrollOffsetPx = getScrollOffsetFromThumbTopPercent({
        thumbTopPercent,
        thumbHeightPercent: scrollState.thumbHeightPercent,
        maxOffsetPx: this.controlPanelEquipmentListMaxScrollOffset(),
        reverse: false,
      });
      return true;
    }
    return true;
  }

  // Пересчитывает черновое количество оборудования по позиции курсора на общем слайдере.
  private consumeControlPanelEquipmentSliderAction(action: GameUiAction): boolean {
    const group = this.getSelectedControlPanelEquipmentGroup();
    if (!group || !action.controlRect) {
      return true;
    }
    if (action.type === "dragStart" || action.type === "dragMove") {
      const position = clamp((action.x - action.controlRect.left) / Math.max(1, action.controlRect.width), 0, 1);
      const maxCount = Math.max(1, group.Count);
      this.setControlPanelEquipmentEnabledCount(group, getCountSliderValue(position, maxCount));
    }
    return true;
  }

  // Меняет черновое количество выбранной группы оборудования дискретным шагом.
  private changeControlPanelEquipmentEnabledCount(delta: number): void {
    const group = this.getSelectedControlPanelEquipmentGroup();
    if (!group) {
      return;
    }
    this.setControlPanelEquipmentEnabledCount(group, this.getControlPanelEquipmentEnabledCount(group) + delta);
  }

  // Сохраняет черновое количество с ограничением по доступному числу единиц.
  private setControlPanelEquipmentEnabledCount(group: EquipmentGroup, value: number): void {
    this.sendControlPanelEquipmentMutation(group.ID, { enabledCount: clamp(value, 1, Math.max(1, group.Count)) });
  }

  // Меняет количество топлива для слива в пределах текущего запаса.
  private changeControlPanelFuelDrainAmount(delta: number): void {
    const maxFuel = this.currentControlPanelAmountMax();
    this.controlPanelFuelDrainAmount = clamp(this.controlPanelFuelDrainAmount + delta, 0, maxFuel);
    this.inputController.setControlPanelFuelDrainAmount(this.controlPanelFuelDrainAmount);
  }

  // Меняет количество слива топлива по позиции курсора на полосе.
  private updateControlPanelFuelDrainAmountFromSlider(action: GameUiAction): void {
    if (!action.controlRect || (action.type !== "dragStart" && action.type !== "dragMove")) {
      return;
    }
    const maxFuel = this.currentControlPanelAmountMax();
    const position = clamp((action.x - action.controlRect.left) / Math.max(1, action.controlRect.width), 0, 1);
    this.controlPanelFuelDrainAmount = Math.round(maxFuel * position);
    this.inputController.setControlPanelFuelDrainAmount(this.controlPanelFuelDrainAmount);
  }

  // Возвращает верхнюю границу текущего окна выбора количества.
  private currentControlPanelAmountMax(_selfObject?: GameUiState["selfObject"]): number {
    if (this.controlPanelContainerTransferDialogOpen) {
      return this.controlPanelContainerTransferMaxAmount;
    }
    if (this.controlPanelFuelFillDialogOpen) {
      return this.controlPanelFuelFillMaxAmount;
    }
    if (this.controlPanelConstructorProduceDialogOpen) {
      return this.controlPanelConstructorProduceMaxAmount;
    }
    if (this.controlPanelItemDeconstructionDialogOpen) {
      return this.controlPanelItemDeconstructionMaxAmount;
    }
    return Math.max(0, this.getSelectedFuelTankObject()?.Fuel ?? 0);
  }

  // Запускает перенос между контейнерами сразу или через окно количества для одной строки.
  private startControlPanelContainerTransfer(sourceContainerEquipmentGroupId: number | null, targetContainerEquipmentGroupId: number | null, controllerEquipmentGroupId: number | null, leftToRightDirection: boolean, itemGroupIds: number[]): void {
    if (itemGroupIds.length !== 1) {
      this.sendControlPanelContainerTransfer(sourceContainerEquipmentGroupId, targetContainerEquipmentGroupId, controllerEquipmentGroupId, leftToRightDirection, itemGroupIds);
      return;
    }
    const itemGroup = this.gameUi.state().itemGroups.find((group) => group.ID === itemGroupIds[0]);
    if (!itemGroup || itemGroup.Count <= 0) {
      return;
    }
    this.controlPanelContainerTransferSourceGroupId = sourceContainerEquipmentGroupId;
    this.controlPanelContainerTransferTargetGroupId = targetContainerEquipmentGroupId;
    this.controlPanelContainerTransferControllerGroupId = controllerEquipmentGroupId;
    this.controlPanelContainerTransferLeftToRightDirection = leftToRightDirection;
    this.controlPanelContainerTransferItemGroupIds = itemGroupIds;
    this.controlPanelContainerTransferMaxAmount = itemGroup.Count;
    this.controlPanelFuelDrainAmount = itemGroup.Count;
    this.controlPanelContainerTransferDialogOpen = true;
    this.controlPanelFuelDrainDialogOpen = false;
    this.controlPanelFuelFillDialogOpen = false;
    this.controlPanelConstructorProduceDialogOpen = false;
    this.controlPanelItemDeconstructionDialogOpen = false;
    this.inputController.setControlPanelFuelDrainAmount(this.controlPanelFuelDrainAmount);
  }

  // Открывает окно выбора количества запусков изготовления по выбранной схеме.
  private startControlPanelConstructorProduceItem(): void {
    if (!this.selectedControlPanelUsageRightEquipmentGroupId || !this.selectedControlPanelConstructorMaterialContainerGroupId) {
      return;
    }
    if (this.selectedControlPanelConstructorTab === "objects") {
      this.controlPanelItemDeconstructionDialogOpen = false;
      this.sendControlPanelConstructorProduceItem(1);
      return;
    }
    if (!this.selectedControlPanelConstructorProductContainerGroupId || !this.selectedControlPanelConstructorSchemaId) {
      return;
    }
    this.controlPanelConstructorProduceDialogOpen = true;
    this.controlPanelFuelDrainDialogOpen = false;
    this.controlPanelFuelFillDialogOpen = false;
    this.controlPanelContainerTransferDialogOpen = false;
    this.controlPanelItemDeconstructionDialogOpen = false;
    this.controlPanelFuelDrainAmount = 1;
    this.inputController.setControlPanelFuelDrainAmount(this.controlPanelFuelDrainAmount);
  }

  // Открывает выбор количества для одной строки деконструкции или сразу запускает разбор нескольких строк.
  private startControlPanelItemDeconstruction(): void {
    const deconstructorGroupId = this.selectedControlPanelUsageRightEquipmentGroupId;
    const sourceGroupId = this.selectedControlPanelConstructorMaterialContainerGroupId;
    const targetGroupId = this.selectedControlPanelUsageLeftContainerGroupId;
    const itemGroupIds = this.selectedControlPanelUsageRightItemGroupIds;
    if (!deconstructorGroupId || !sourceGroupId || !targetGroupId || itemGroupIds.length === 0) {
      return;
    }
    if (itemGroupIds.length !== 1) {
      this.controlPanelItemDeconstructionDialogOpen = false;
      this.sendControlPanelItemDeconstruction(deconstructorGroupId, sourceGroupId, targetGroupId, itemGroupIds);
      return;
    }
    const itemGroup = this.gameUi.state().itemGroups.find((group) => group.ID === itemGroupIds[0] && group.ContainerEquipmentGroupID === sourceGroupId);
    if (!itemGroup) {
      return;
    }
    this.controlPanelItemDeconstructionDeconstructorGroupId = deconstructorGroupId;
    this.controlPanelItemDeconstructionSourceGroupId = sourceGroupId;
    this.controlPanelItemDeconstructionTargetGroupId = targetGroupId;
    this.controlPanelItemDeconstructionItemGroupIds = itemGroupIds;
    this.controlPanelItemDeconstructionMaxAmount = itemGroup.Count;
    this.controlPanelFuelDrainAmount = itemGroup.Count;
    this.controlPanelItemDeconstructionDialogOpen = true;
    this.controlPanelFuelDrainDialogOpen = false;
    this.controlPanelFuelFillDialogOpen = false;
    this.controlPanelContainerTransferDialogOpen = false;
    this.controlPanelConstructorProduceDialogOpen = false;
    this.inputController.setControlPanelFuelDrainAmount(this.controlPanelFuelDrainAmount);
  }

  // Возвращает новый выбор строк содержимого контейнера с учетом Ctrl и Shift.
  private updateControlPanelUsageItemSelection(containerGroupId: number | null, selectedIds: number[], anchorId: number | null, clickedId: number, action: GameUiAction): { selectedIds: number[]; anchorId: number } {
    const orderedIds = this.gameUi.state().itemGroups
      .filter((itemGroup) => itemGroup.ContainerEquipmentGroupID === containerGroupId)
      .map((itemGroup) => itemGroup.ID);
    return applyControlPanelListSelection({
      orderedIds,
      selectedIds,
      clickedId,
      anchorId,
      action,
    });
  }

  // Возвращает контейнер, содержимое которого сейчас показано в правой части использования.
  private getControlPanelUsageRightContentContainerGroupId(): number | null {
    const rightGroup = this.selectedControlPanelUsageRightEquipmentGroupId ? this.getControlPanelEquipmentGroup(this.selectedControlPanelUsageRightEquipmentGroupId) : null;
    if (rightGroup && this.isEquipmentGroupItemType(rightGroup, "Constructor")) {
      return this.selectedControlPanelConstructorMaterialContainerGroupId;
    }
    if (rightGroup && this.isEquipmentGroupItemType(rightGroup, "Deconstructor")) {
      return this.selectedControlPanelConstructorMaterialContainerGroupId;
    }
    return this.selectedControlPanelUsageRightEquipmentGroupId;
  }

  // Отправляет изменение оборудования и кладет его поверх снимков до серверного подтверждения.
  private sendControlPanelEquipmentMutation(groupId: number, update: { enabled?: boolean; enabledCount?: number; title?: string }): void {
    const mutation = this.gameClient?.sendControlPanelEquipmentUpdate({ equipmentGroupId: groupId, ...update });
    if (!mutation) {
      return;
    }
    const current = this.controlPanelPending.equipment[groupId] ?? {};
    this.controlPanelPending = {
      ...this.controlPanelPending,
      equipment: {
        ...this.controlPanelPending.equipment,
        [groupId]: {
          ...current,
          enabled: update.enabled === undefined ? current.enabled : { ...mutation, value: update.enabled },
          enabledCount: update.enabledCount === undefined ? current.enabledCount : { ...mutation, value: update.enabledCount },
          title: update.title === undefined ? current.title : { ...mutation, value: update.title },
        },
      },
    };
  }

  // Отправляет перенос между двумя выбранными контейнерами панели управления.
  private sendControlPanelContainerTransfer(sourceContainerEquipmentGroupId: number | null, targetContainerEquipmentGroupId: number | null, controllerEquipmentGroupId: number | null, leftToRightDirection: boolean, itemGroupIds: number[], amount = 0): void {
    if (!sourceContainerEquipmentGroupId || !targetContainerEquipmentGroupId || !controllerEquipmentGroupId || sourceContainerEquipmentGroupId === targetContainerEquipmentGroupId || itemGroupIds.length === 0) {
      return;
    }
    this.gameClient?.sendControlPanelContainerTransfer({
      controllerEquipmentGroupId,
      leftToRightDirection,
      sourceContainerEquipmentGroupId,
      targetContainerEquipmentGroupId,
      itemGroupIds,
      amount: amount > 0 ? amount : undefined,
    });
  }

  // Отправляет перенос топлива между левым контейнером и выбранным топливным баком.
  private sendControlPanelFuelTransfer(containerEquipmentGroupId: number | null, fuelTankEquipmentGroupId: number | null, itemGroupIds: number[], amount = 0): void {
    if (!containerEquipmentGroupId || !fuelTankEquipmentGroupId || (itemGroupIds.length === 0 && amount <= 0)) {
      return;
    }
    this.gameClient?.sendControlPanelFuelTransfer({
      containerEquipmentGroupId,
      fuelTankEquipmentGroupId,
      itemGroupIds,
      amount: amount > 0 ? amount : undefined,
    });
  }

  // Отправляет запуск деконструкции выбранных предметов из правого контейнера в левый контейнер.
  private sendControlPanelItemDeconstruction(deconstructorGroupId: number | null, sourceGroupId: number | null, targetGroupId: number | null, itemGroupIds: number[], amount = 0): void {
    if (!deconstructorGroupId || !sourceGroupId || !targetGroupId || itemGroupIds.length === 0) {
      return;
    }
    this.gameClient?.sendControlPanelItemDeconstruction({
      deconstructorEquipmentGroupId: deconstructorGroupId,
      sourceContainerEquipmentGroupId: sourceGroupId,
      targetContainerEquipmentGroupId: targetGroupId,
      itemGroupIds,
      amount: amount > 0 ? amount : undefined,
    });
  }

  // Отправляет изготовление одной партии предметов по выбранной схеме конструктора.
  private sendControlPanelConstructorProduceItem(amount: number): void {
    if (
      !this.selectedControlPanelUsageRightEquipmentGroupId ||
      !this.selectedControlPanelConstructorMaterialContainerGroupId
    ) {
      return;
    }
    if (this.selectedControlPanelConstructorTab === "objects") {
      if (!this.selectedControlPanelConstructorBlueprintId) {
        return;
      }
      this.gameClient?.sendControlPanelConstructorProduceItem({
        constructorEquipmentGroupId: this.selectedControlPanelUsageRightEquipmentGroupId,
        materialContainerEquipmentGroupId: this.selectedControlPanelConstructorMaterialContainerGroupId,
        blueprintId: this.selectedControlPanelConstructorBlueprintId,
        amount: 1,
      });
      return;
    }
    if (!this.selectedControlPanelConstructorProductContainerGroupId || !this.selectedControlPanelConstructorSchemaId) {
      return;
    }
    this.gameClient?.sendControlPanelConstructorProduceItem({
      constructorEquipmentGroupId: this.selectedControlPanelUsageRightEquipmentGroupId,
      materialContainerEquipmentGroupId: this.selectedControlPanelConstructorMaterialContainerGroupId,
      productContainerEquipmentGroupId: this.selectedControlPanelConstructorProductContainerGroupId,
      schemaId: this.selectedControlPanelConstructorSchemaId,
      amount,
    });
  }

  // Возвращает серверную группу оборудования по ID из последнего UI-снимка.
  // Отправляет команду изменения основной очереди выбранного конструктора.
  private sendControlPanelConstructorQueueCommand(command: "skipNext" | "skipAllNext" | "cancel" | "cancelAll"): void {
    if (!this.selectedControlPanelUsageRightEquipmentGroupId || !this.selectedControlPanelConstructorMainJobId) {
      return;
    }
    this.gameClient?.sendControlPanelConstructorQueueCommand({
      constructorEquipmentGroupId: this.selectedControlPanelUsageRightEquipmentGroupId,
      jobId: this.selectedControlPanelConstructorMainJobId,
      command,
    });
  }

  // Сбрасывает выбор строки основной очереди, если сервер больше не присылает эту строку.
  private syncControlPanelConstructorMainJobSelection(jobs: ConstructorProductionJob[], tasks: Task[]): void {
    if (!this.selectedControlPanelConstructorMainJobId) {
      return;
    }
    const selectedConstructorJobExists = jobs.some((job) =>
      job.id === this.selectedControlPanelConstructorMainJobId &&
      job.queueType === "main" &&
      job.constructorEquipmentGroupId === this.selectedControlPanelUsageRightEquipmentGroupId,
    );
    const selectedTaskJobExists = tasks.some((task) =>
      task.ID === this.selectedControlPanelConstructorMainJobId &&
      task.ParentTaskID === 0 &&
      task.ControllerEquipmentGroupID === this.selectedControlPanelUsageRightEquipmentGroupId &&
      this.isControlPanelSelectableTaskQueueType(task.TaskTypeID),
    );
    if (!selectedConstructorJobExists && !selectedTaskJobExists) {
      this.selectedControlPanelConstructorMainJobId = null;
    }
  }

  // Проверяет, что тип задания относится к очередям, строки которых можно выбирать в панели оборудования.
  private isControlPanelSelectableTaskQueueType(taskTypeId: number): boolean {
    const taskType = this.referenceData?.TaskType?.Items[String(taskTypeId)];
    return Boolean(taskType && controlPanelSelectableTaskQueueTypeAcronyms.has(taskType.Acronym));
  }

  // Создает старую форму строк очереди из новой таблицы заданий для существующих компонентов UI.
  private constructorProductionJobsFromTasks(tasks: Task[], equipmentGroups: EquipmentGroup[]): ConstructorProductionJob[] {
    const itemProductionTypeId = this.taskTypeId("ItemProduction");
    const objectProductionTypeId = this.taskTypeId("ObjectProduction");
    return tasks
      .filter((task) => task.TaskTypeID === itemProductionTypeId || task.TaskTypeID === objectProductionTypeId)
      .sort((left, right) => left.ID - right.ID)
      .map((task) => {
        const schema = task.SchemaID > 0 ? this.referenceData?.Schema.Items[String(task.SchemaID)] : undefined;
        const blueprint = task.BlueprintID > 0 ? this.referenceData?.Blueprint.Items[String(task.BlueprintID)] : undefined;
        const amount = task.BatchCount > 0 ? task.BatchCount : 1;
        const remainingAmount = this.remainingTaskCount(task.RemainingEnergy, task.TotalEnergy, amount);
        const remainingTime = this.taskEnergyToSeconds(task, task.RemainingEnergy, equipmentGroups);
        const totalTime = this.taskEnergyToSeconds(task, task.TotalEnergy, equipmentGroups);
        return {
          id: task.ID,
          constructorEquipmentGroupId: task.ControllerEquipmentGroupID,
          queueType: task.ParentTaskID > 0 ? "auxiliary" : "main",
          schemaId: task.SchemaID,
          blueprintId: task.BlueprintID,
          productItemModelId: schema?.ItemModelID ?? 0,
          productCosmicObjectModelId: blueprint?.CosmicObjectModelID ?? 0,
          productCount: schema?.Count ?? (blueprint ? 1 : 0),
          remainingCount: (schema?.Count ?? (blueprint ? 1 : 0)) * remainingAmount,
          totalCount: (schema?.Count ?? (blueprint ? 1 : 0)) * amount,
          remainingTime,
          totalTime,
          running: task.RemainingEnergy < task.TotalEnergy || (this.gameUi.state().taskItemGroups ?? []).some((group) => group.TaskID === task.ID),
          parentJobId: task.ParentTaskID,
        };
      });
  }

  // Оценивает оставшееся количество результата по доле невыполненной работы.
  private remainingTaskCount(remainingEnergy: number, totalEnergy: number, amount: number): number {
    if (amount <= 0) {
      return 1;
    }
    if (totalEnergy <= 0) {
      return amount;
    }
    if (remainingEnergy <= 0) {
      return 0;
    }
    const completed = Math.floor(((totalEnergy - remainingEnergy) / totalEnergy) * amount + 1e-9);
    return Math.max(0, amount - completed);
  }

  // Возвращает числовой код типа задания по акрониму.
  private taskTypeId(acronym: string): number | null {
    const taskType = Object.values(this.referenceData?.TaskType?.Items ?? {}).find((item) => item.Acronym === acronym);
    return taskType?.ID ?? null;
  }

  // Переводит работу задания в секунды по доступной мощности исполнителей.
  private taskEnergyToSeconds(task: Task, energy: number, equipmentGroups: EquipmentGroup[]): number {
    const power = this.taskWorkPower(task, equipmentGroups);
    return power > 0 ? energy / power : energy;
  }

  // Считает мощность так же, как сервер, но использует только данные последнего снимка.
  private taskWorkPower(task: Task, equipmentGroups: EquipmentGroup[]): number {
    const controller = equipmentGroups.find((group) => group.ID === task.ControllerEquipmentGroupID);
    if (!controller || !this.referenceData) {
      return 0;
    }
    let power = 0;
    const implementers = Object.values(this.referenceData.Implementer?.Items ?? {}).filter((item) => item.TaskTypeID === task.TaskTypeID);
    for (const implementer of implementers) {
      for (const group of equipmentGroups.filter((item) => item.CosmicObjectID === controller.CosmicObjectID)) {
        const model = this.referenceData.ItemModel.Items[String(group.EquipmentItemModelID)];
        if (!model || model.ItemTypeID !== implementer.ImplementerEquipmentItemTypeID) {
          continue;
        }
        const enabledCount = Math.max(0, group.Enabled ? group.EnabledCount : 0);
        const modelPower = typeof model.ConsumingPower === "number" && model.ConsumingPower > 0 ? model.ConsumingPower : 1;
        const efficiency = typeof model.Efficiency === "number" && model.Efficiency > 0 ? model.Efficiency : 1;
        power += modelPower * enabledCount * implementer.WorkPart * efficiency;
      }
    }
    return power;
  }

  private getControlPanelEquipmentGroup(groupId: number): EquipmentGroup | null {
    return this.gameUi.state().equipmentGroups.find((group) => group.ID === groupId) ?? null;
  }

  // Возвращает первую группу указанного типа на выбранном объекте.
  private firstControlPanelGroupIdOnObject(objectId: number | null, itemTypeAcronym: string): number | null {
    return this.gameUi.state().equipmentGroups
      .filter((group) => group.CosmicObjectID === objectId && this.isEquipmentGroupItemType(group, itemTypeAcronym))
      .sort((left, right) => left.ID - right.ID)[0]?.ID ?? null;
  }

  // Возвращает первую группу, пригодную для правой панели на выбранном объекте.
  private firstControlPanelInternalGroupIdOnObject(objectId: number | null): number | null {
    return this.gameUi.state().equipmentGroups
      .filter((group) => group.CosmicObjectID === objectId && this.isEquipmentGroupInternalUsable(group))
      .sort((left, right) => left.ID - right.ID)[0]?.ID ?? null;
  }

  // Возвращает объект выбранного топливного бака.
  private getSelectedFuelTankObject(): CosmicObject | null {
    const fuelTankGroup = this.selectedControlPanelUsageRightEquipmentGroupId ? this.getControlPanelEquipmentGroup(this.selectedControlPanelUsageRightEquipmentGroupId) : null;
    return this.gameUi.state().objects.find((object) => object.ID === fuelTankGroup?.CosmicObjectID) ?? null;
  }

  // Возвращает выбранную группу или первую доступную группу оборудования текущего объекта.
  private getSelectedControlPanelEquipmentGroup(): EquipmentGroup | null {
    const groups = this.getControlPanelEquipmentGroups();
    return groups.find((group) => group.ID === this.selectedControlPanelEquipmentGroupId) ?? groups[0] ?? null;
  }

  // Выбирает группу из переданного снимка, чтобы поле названия не зависело от предыдущего UI-кадра.
  private getSelectedControlPanelEquipmentGroupFromList(groups: EquipmentGroup[], objectId: number | null): EquipmentGroup | null {
    const objectGroups = groups
      .filter((group) => group.CosmicObjectID === objectId)
      .sort((left, right) => left.ID - right.ID);
    return objectGroups.find((group) => group.ID === this.selectedControlPanelEquipmentGroupId) ?? objectGroups[0] ?? null;
  }

  // Возвращает группы оборудования текущего объекта из последнего UI-снимка.
  private getControlPanelEquipmentGroups(): EquipmentGroup[] {
    const state = this.gameUi.state();
    const objectId = state.selfObject?.ID;
    if (!objectId) {
      return [];
    }
    return state.equipmentGroups
      .filter((group) => group.CosmicObjectID === objectId)
      .sort((left, right) => left.ID - right.ID);
  }

  // Проверяет тип модели оборудования по стабильному акрониму.
  private isEquipmentGroupItemType(group: EquipmentGroup, itemTypeAcronym: string): boolean {
    const itemModel = this.referenceData?.ItemModel.Items[String(group.EquipmentItemModelID)];
    const itemType = this.referenceData?.ItemType.Items[String(itemModel?.ItemTypeID)];
    return itemType?.Acronym === itemTypeAcronym;
  }

  // Проверяет, что оборудование можно выбрать в правой панели использования.
  private isEquipmentGroupInternalUsable(group: EquipmentGroup): boolean {
    const itemModel = this.referenceData?.ItemModel.Items[String(group.EquipmentItemModelID)];
    const itemType = this.referenceData?.ItemType.Items[String(itemModel?.ItemTypeID)];
    return Boolean(itemType?.IsInternalUsable);
  }

  // Возвращает effective-признак включения группы оборудования из снимка с учетом pending.
  private getControlPanelEquipmentEnabled(group: EquipmentGroup): boolean {
    return group.Enabled;
  }

  // Возвращает effective-количество включенных единиц из снимка с учетом pending.
  private getControlPanelEquipmentEnabledCount(group: EquipmentGroup): number {
    return clamp(group.EnabledCount, 1, Math.max(1, group.Count));
  }

  // Применяет действия модального окна настроек и не пропускает их в демонстрационную панель.
  private consumeSettingsUiAction(action: GameUiAction): boolean {
    if (action.type === "cancel") {
      this.openInputSettingsActionId = null;
      return true;
    }
    if (action.type !== "click") {
      return this.consumeSettingsScrollbarAction(action);
    }
    if (action.controlId === "settings-cancel-button") {
      this.resetInputSettingsDraftFromServer();
      this.openInputSettingsActionId = null;
      this.inputSettingsSaving = false;
      this.inputSettingsError = null;
      this.inputController.closeSettings();
      return true;
    }
    if (action.controlId === "settings-save-button") {
      this.inputSettingsSaving = true;
      this.inputSettingsError = null;
      this.gameClient?.saveInputSettings(toInputSettingsPayload(this.inputSettingsValues));
      return true;
    }
    const inputSelectMatch = action.controlId.match(/^settings-input-select-(\d+)$/);
    if (inputSelectMatch) {
      const actionTypeID = Number(inputSelectMatch[1]);
      this.openInputSettingsActionId = this.openInputSettingsActionId === actionTypeID ? null : actionTypeID;
      this.settingsDropdownScrollOffsetPx = 0;
      this.settingsDropdownScrollbarDrag = null;
      return true;
    }
    if (action.controlId.startsWith("settings-input-select-") && typeof action.value === "string") {
      const parts = action.controlId.split("-");
      const actionTypeID = Number(parts[3]);
      const inputEventTypeID = Number(action.value);
      if (actionTypeID > 0 && inputEventTypeID > 0) {
        this.inputSettingsValues = { ...this.inputSettingsValues, [actionTypeID]: inputEventTypeID };
        this.openInputSettingsActionId = null;
        this.settingsDropdownScrollOffsetPx = 0;
        this.settingsDropdownScrollbarDrag = null;
        this.inputController.updateInputBindings(getInputBindingMap(this.referenceData, this.inputSettingsValues));
        return true;
      }
    }
    if (action.controlId.startsWith("settings-tab-")) {
      if (action.value === "video" || action.value === "audio" || action.value === "input") {
        this.selectedSettingsTab = action.value;
      }
      this.openInputSettingsActionId = null;
      return true;
    }
    if (action.controlId === "settings-modal") {
      return true;
    }
    this.openInputSettingsActionId = null;
    return action.controlId.startsWith("settings-");
  }

  // Применяет колесо мыши к активной прокручиваемой области окна настроек.
  private consumeSettingsWheel(): void {
    const deltaY = this.inputController.consumeSettingsWheelDeltaY();
    if (deltaY === 0 || !this.inputController.isSettingsVisible()) {
      return;
    }
    if (this.openInputSettingsActionId !== null) {
      this.settingsDropdownScrollOffsetPx = clamp(this.settingsDropdownScrollOffsetPx + deltaY, 0, this.settingsDropdownMaxScrollOffset());
      return;
    }
    this.settingsInputScrollOffsetPx = clamp(this.settingsInputScrollOffsetPx + deltaY, 0, this.settingsInputMaxScrollOffset());
  }

  // Обрабатывает перетаскивание единых UI Kit полос прокрутки в окне настроек.
  private consumeSettingsScrollbarAction(action: GameUiAction): boolean {
    if (action.kind !== "scrollbar" || !action.controlRect) {
      return false;
    }
    if (action.controlId === "settings-input-scrollbar") {
      return this.consumeSettingsInputScrollbarAction(action);
    }
    if (action.controlId.startsWith("settings-input-select-") && action.controlId.endsWith("-scrollbar")) {
      return this.consumeSettingsDropdownScrollbarAction(action);
    }
    return false;
  }

  // Обрабатывает мышиные действия окна обмена.
  private consumeExchangeUiAction(action: GameUiAction): boolean {
    if (!this.exchangeState) {
      return false;
    }
    if (action.kind === "scrollbar" && action.controlId.endsWith("-scrollbar") && this.consumeListScrollbarAction(action)) {
      return true;
    }
    if (action.kind === "slider" && action.controlId === "control-panel-fuel-drain-amount-slider") {
      this.updateControlPanelFuelDrainAmountFromSlider(action);
      return true;
    }
    if (action.type !== "click") {
      return false;
    }
    const receiverLocked = this.exchangeState.selfConfirmed;
    const sourceLocked = this.exchangeState.otherConfirmed;
    const cancelLocked = receiverLocked && sourceLocked;
    if (action.controlId === "exchange-modal-close-button" || action.controlId === "exchange-cancel-button") {
      if (cancelLocked) {
        return true;
      }
      this.gameClient?.sendExchangeCancel();
      return true;
    }
    if (action.controlId === "exchange-confirm-button") {
      this.gameClient?.sendExchangeConfirm();
      return true;
    }
    if (action.controlId === "exchange-move-to-queue-button") {
      if (sourceLocked) {
        return true;
      }
      this.startExchangeAddItems();
      return true;
    }
    if (action.controlId === "exchange-add-items-cancel") {
      this.controlPanelContainerTransferDialogOpen = false;
      this.inputController.blurControlPanelFuelDrainAmount();
      return true;
    }
    if (action.controlId === "exchange-add-items-ok") {
      const amount = clamp(this.inputController.getControlPanelFuelDrainAmount(this.controlPanelFuelDrainAmount), 1, this.controlPanelContainerTransferMaxAmount);
      if (this.controlPanelContainerTransferSourceGroupId && this.exchangeState.selfSourceContainerEquipmentGroupId !== this.controlPanelContainerTransferSourceGroupId) {
        this.gameClient?.sendExchangeSelectSource(this.controlPanelContainerTransferSourceGroupId);
      }
      this.gameClient?.sendExchangeAddItems(this.controlPanelContainerTransferItemGroupIds, amount);
      this.selectedExchangeSourceItemGroupIds = [];
      this.selectedExchangeSourceAnchorItemGroupId = null;
      this.controlPanelContainerTransferDialogOpen = false;
      this.inputController.blurControlPanelFuelDrainAmount();
      return true;
    }
    if (action.controlId.startsWith("exchange-source-list-") && typeof action.value === "string") {
      if (sourceLocked) {
        return true;
      }
      const itemGroupID = Number(action.value);
      const selection = this.updateExchangeSourceItemSelection(itemGroupID, action);
      this.selectedExchangeSourceItemGroupIds = selection.selectedIds;
      this.selectedExchangeSourceAnchorItemGroupId = selection.anchorId;
      return true;
    }
    if (action.controlId === "exchange-receiver-object-select") {
      if (receiverLocked) {
        return true;
      }
      this.openExchangeSelect = this.openExchangeSelect === "receiverObject" ? null : "receiverObject";
      return true;
    }
    if (action.controlId === "exchange-receiver-container-select") {
      if (receiverLocked) {
        return true;
      }
      this.openExchangeSelect = this.openExchangeSelect === "receiverContainer" ? null : "receiverContainer";
      return true;
    }
    if (action.controlId === "exchange-source-object-select") {
      if (sourceLocked) {
        return true;
      }
      this.openExchangeSelect = this.openExchangeSelect === "sourceObject" ? null : "sourceObject";
      return true;
    }
    if (action.controlId === "exchange-source-container-select") {
      if (sourceLocked) {
        return true;
      }
      this.openExchangeSelect = this.openExchangeSelect === "sourceContainer" ? null : "sourceContainer";
      return true;
    }
    if (typeof action.value !== "string") {
      return false;
    }
    if (action.controlId.startsWith("exchange-receiver-object-select-")) {
      if (receiverLocked) {
        return true;
      }
      this.selectedExchangeReceiverObjectId = Number(action.value);
      this.openExchangeSelect = null;
      return true;
    }
    if (action.controlId.startsWith("exchange-source-object-select-")) {
      if (sourceLocked) {
        return true;
      }
      this.selectedExchangeSourceObjectId = Number(action.value);
      this.openExchangeSelect = null;
      this.selectedExchangeSourceItemGroupIds = [];
      this.selectedExchangeSourceAnchorItemGroupId = null;
      return true;
    }
    if (action.controlId.startsWith("exchange-receiver-container-select-")) {
      if (receiverLocked) {
        return true;
      }
      this.openExchangeSelect = null;
      this.gameClient?.sendExchangeSelectReceiver(Number(action.value));
      return true;
    }
    if (action.controlId.startsWith("exchange-source-container-select-")) {
      if (sourceLocked) {
        return true;
      }
      this.openExchangeSelect = null;
      this.selectedExchangeSourceItemGroupIds = [];
      this.selectedExchangeSourceAnchorItemGroupId = null;
      this.gameClient?.sendExchangeSelectSource(Number(action.value));
      return true;
    }
    return false;
  }

  // Открывает выбор количества для одной строки обмена или сразу отправляет несколько строк.
  private startExchangeAddItems(): void {
    if (!this.exchangeState || this.selectedExchangeSourceItemGroupIds.length === 0) {
      return;
    }
    const sourceContainerID = this.currentExchangeSourceContainerID();
    if (this.selectedExchangeSourceItemGroupIds.length !== 1) {
      if (sourceContainerID && this.exchangeState.selfSourceContainerEquipmentGroupId !== sourceContainerID) {
        this.gameClient?.sendExchangeSelectSource(sourceContainerID);
      }
      this.gameClient?.sendExchangeAddItems(this.selectedExchangeSourceItemGroupIds, 1);
      this.selectedExchangeSourceItemGroupIds = [];
      this.selectedExchangeSourceAnchorItemGroupId = null;
      return;
    }
    const itemGroup = this.gameUi.state().itemGroups.find((group) => group.ID === this.selectedExchangeSourceItemGroupIds[0]);
    if (!itemGroup || itemGroup.Count <= 0) {
      return;
    }
    this.controlPanelContainerTransferSourceGroupId = sourceContainerID;
    this.controlPanelContainerTransferItemGroupIds = [...this.selectedExchangeSourceItemGroupIds];
    this.controlPanelContainerTransferMaxAmount = itemGroup.Count;
    this.controlPanelFuelDrainAmount = itemGroup.Count;
    this.controlPanelContainerTransferDialogOpen = true;
    this.controlPanelFuelDrainDialogOpen = false;
    this.controlPanelFuelFillDialogOpen = false;
    this.controlPanelConstructorProduceDialogOpen = false;
    this.controlPanelItemDeconstructionDialogOpen = false;
    this.inputController.setControlPanelFuelDrainAmount(this.controlPanelFuelDrainAmount);
  }

  // Возвращает фактически выбранный или показанный по умолчанию контейнер-источник обмена.
  private currentExchangeSourceContainerID(): number | null {
    if (!this.exchangeState) {
      return null;
    }
    if (this.exchangeState.selfSourceContainerEquipmentGroupId > 0) {
      return this.exchangeState.selfSourceContainerEquipmentGroupId;
    }
    const objectID = this.selectedExchangeSourceObjectId ?? this.exchangeState.selfObjectId;
    const state = this.gameUi.state();
    return state.equipmentGroups
      .filter((group) => group.CosmicObjectID === objectID && this.isExchangeContainerEquipmentGroup(group.EquipmentItemModelID))
      .sort((left, right) => left.ID - right.ID)[0]?.ID ?? null;
  }

  // Возвращает фактически выбранный или показанный по умолчанию контейнер-приемник обмена.
  private currentExchangeReceiverContainerID(): number | null {
    if (!this.exchangeState) {
      return null;
    }
    if (this.exchangeState.selfReceiverContainerEquipmentGroupId > 0) {
      return this.exchangeState.selfReceiverContainerEquipmentGroupId;
    }
    const objectID = this.selectedExchangeReceiverObjectId ?? this.exchangeState.selfObjectId;
    const state = this.gameUi.state();
    return state.equipmentGroups
      .filter((group) => group.CosmicObjectID === objectID && this.isExchangeContainerEquipmentGroup(group.EquipmentItemModelID))
      .sort((left, right) => left.ID - right.ID)[0]?.ID ?? null;
  }

  // Возвращает новый выбор строк источника обмена с учётом Ctrl и Shift.
  private updateExchangeSourceItemSelection(clickedId: number, action: GameUiAction): { selectedIds: number[]; anchorId: number } {
    const sourceContainerID = this.currentExchangeSourceContainerID();
    const orderedIds = this.gameUi.state().itemGroups
      .filter((itemGroup) => itemGroup.ContainerEquipmentGroupID === sourceContainerID)
      .map((itemGroup) => itemGroup.ID);
    return applyControlPanelListSelection({
      orderedIds,
      selectedIds: this.selectedExchangeSourceItemGroupIds,
      clickedId,
      anchorId: this.selectedExchangeSourceAnchorItemGroupId,
      action,
    });
  }

  // Проверяет, что модель оборудования является контейнером для окна обмена.
  private isExchangeContainerEquipmentGroup(equipmentItemModelID: number): boolean {
    const itemModel = this.gameUi.state().referenceData?.ItemModel.Items[String(equipmentItemModelID)];
    const itemType = this.gameUi.state().referenceData?.ItemType.Items[String(itemModel?.ItemTypeID)];
    return itemType?.Acronym === "Container";
  }

  // Пересчитывает сдвиг списка действий по положению его ползунка.
  private consumeSettingsInputScrollbarAction(action: GameUiAction): boolean {
    const scrollState = this.getSettingsInputScrollState();
    if (action.type === "dragStart") {
      this.settingsInputScrollbarDrag = startScrollbarDrag({
        top: action.controlRect?.top ?? 0,
        height: action.controlRect?.height ?? 1,
        thumbTopPercent: scrollState.thumbTopPercent,
        thumbHeightPercent: scrollState.thumbHeightPercent,
      }, action.y);
      return true;
    }
    if (action.type === "dragEnd") {
      this.settingsInputScrollbarDrag = null;
      return true;
    }
    if (action.type === "dragMove" && this.settingsInputScrollbarDrag) {
      const thumbTopPercent = getScrollbarThumbTopPercentFromCursor({
        top: action.controlRect?.top ?? 0,
        height: action.controlRect?.height ?? 1,
        thumbHeightPercent: scrollState.thumbHeightPercent,
        drag: this.settingsInputScrollbarDrag,
      }, action.y);
      this.settingsInputScrollOffsetPx = getScrollOffsetFromThumbTopPercent({
        thumbTopPercent,
        thumbHeightPercent: scrollState.thumbHeightPercent,
        maxOffsetPx: this.settingsInputMaxScrollOffset(),
        reverse: false,
      });
      return true;
    }
    return true;
  }

  // Пересчитывает сдвиг раскрытого списка событий по положению его ползунка.
  private consumeSettingsDropdownScrollbarAction(action: GameUiAction): boolean {
    const scrollState = this.getSettingsDropdownScrollState();
    if (action.type === "dragStart") {
      this.settingsDropdownScrollbarDrag = startScrollbarDrag({
        top: action.controlRect?.top ?? 0,
        height: action.controlRect?.height ?? 1,
        thumbTopPercent: scrollState.thumbTopPercent,
        thumbHeightPercent: scrollState.thumbHeightPercent,
      }, action.y);
      return true;
    }
    if (action.type === "dragEnd") {
      this.settingsDropdownScrollbarDrag = null;
      return true;
    }
    if (action.type === "dragMove" && this.settingsDropdownScrollbarDrag) {
      const thumbTopPercent = getScrollbarThumbTopPercentFromCursor({
        top: action.controlRect?.top ?? 0,
        height: action.controlRect?.height ?? 1,
        thumbHeightPercent: scrollState.thumbHeightPercent,
        drag: this.settingsDropdownScrollbarDrag,
      }, action.y);
      this.settingsDropdownScrollOffsetPx = getScrollOffsetFromThumbTopPercent({
        thumbTopPercent,
        thumbHeightPercent: scrollState.thumbHeightPercent,
        maxOffsetPx: this.settingsDropdownMaxScrollOffset(),
        reverse: false,
      });
      return true;
    }
    return true;
  }

  // Запрашивает актуальные серверные значения один раз при открытии окна настроек.
  private requestInputSettingsOnOpen(settingsVisible: boolean): void {
    if (settingsVisible && !this.previousSettingsVisible) {
      this.gameClient?.requestInputSettings();
      this.resetInputSettingsDraftFromServer();
      this.openInputSettingsActionId = null;
      this.inputSettingsSaving = false;
      this.inputSettingsError = null;
    }
    this.previousSettingsVisible = settingsVisible;
  }

  // Возвращает черновик окна к последним значениям, которые уже подтвердил сервер.
  private resetInputSettingsDraftFromServer(): void {
    this.inputSettingsValues = getMergedInputSettingValues(this.referenceData, this.gameClient?.getLatestInputSettings() ?? []);
    this.settingsInputScrollOffsetPx = clamp(this.settingsInputScrollOffsetPx, 0, this.settingsInputMaxScrollOffset());
    this.settingsDropdownScrollOffsetPx = clamp(this.settingsDropdownScrollOffsetPx, 0, this.settingsDropdownMaxScrollOffset());
    this.inputController.updateInputBindings(getInputBindingMap(this.referenceData, this.inputSettingsValues));
  }

  // Синхронизирует черновик и фактические привязки после серверного ответа.
  private syncInputSettingsFromServer(): void {
    const seq = this.gameClient?.getLatestInputSettingsSeq() ?? 0;
    if (seq !== this.inputSettingsSeq) {
      this.inputSettingsSeq = seq;
      this.resetInputSettingsDraftFromServer();
      this.inputSettingsSaving = false;
      this.inputSettingsError = null;
    }
    const errorSeq = this.gameClient?.getLatestInputSettingsErrorSeq() ?? 0;
    if (errorSeq !== this.inputSettingsErrorSeq) {
      this.inputSettingsErrorSeq = errorSeq;
      this.inputSettingsError = this.gameClient?.getLatestInputSettingsError() ?? null;
      this.inputSettingsSaving = false;
    }
  }

  // Возвращает состояние полосы прокрутки списка действий.
  private getSettingsInputScrollState(): ChatScrollState {
    return this.getScrollState(
      this.settingsInputScrollOffsetPx,
      this.settingsInputViewportHeightPx(),
      this.settingsInputContentHeightPx(),
      this.settingsInputScrollbarDrag !== null,
    );
  }

  // Возвращает состояние полосы прокрутки раскрытого списка событий.
  private getSettingsDropdownScrollState(): ChatScrollState {
    return this.getScrollState(
      this.settingsDropdownScrollOffsetPx,
      this.settingsDropdownViewportHeightPx(),
      this.settingsDropdownContentHeightPx(),
      this.settingsDropdownScrollbarDrag !== null,
    );
  }

  // Возвращает состояние полосы прокрутки списка групп оборудования.
  private getControlPanelEquipmentListScrollState(): ChatScrollState {
    return this.getScrollState(
      this.controlPanelEquipmentListScrollOffsetPx,
      this.controlPanelEquipmentListViewportHeightPx(),
      this.controlPanelEquipmentListContentHeightPx(),
      this.controlPanelEquipmentListScrollbarDrag !== null,
    );
  }

  // Возвращает состояния прокрутки для всех видимых обычных списков.
  private getListScrollStates(): Record<string, ChatScrollState> {
    const scrollStates: Record<string, ChatScrollState> = {};
    for (const listId of scrollableListIds) {
      if (this.isScrollableListId(listId)) {
        scrollStates[listId] = this.getListScrollState(listId);
      }
    }
    return scrollStates;
  }

  // Возвращает состояние прокрутки одного обычного списка.
  private getListScrollState(listId: string): ChatScrollState {
    const maxOffsetPx = this.listMaxScrollOffset(listId);
    const offsetPx = clamp(this.listScrollOffsets.get(listId) ?? 0, 0, maxOffsetPx);
    this.listScrollOffsets.set(listId, offsetPx);
    if (listId === "control-panel-equipment-list") {
      this.controlPanelEquipmentListScrollOffsetPx = offsetPx;
    }
    return this.getScrollState(
      offsetPx,
      this.listViewportHeightPx(listId),
      this.listContentHeightPx(listId),
      this.listScrollbarDrags.has(listId),
    );
  }

  // Формирует универсальное состояние скролла для UI Kit полосы.
  private getScrollState(offsetPx: number, viewportHeightPx: number, contentHeightPx: number, dragging: boolean): ChatScrollState {
    const maxOffsetPx = Math.max(0, contentHeightPx - viewportHeightPx);
    if (viewportHeightPx <= 0 || contentHeightPx <= viewportHeightPx || maxOffsetPx <= 0) {
      return { visible: false, thumbTopPercent: 0, thumbHeightPercent: 100, contentOffsetPx: 0, dragging };
    }
    const contentOffsetPx = clamp(offsetPx, 0, maxOffsetPx);
    const thumbHeightPercent = Math.max(14, (viewportHeightPx / contentHeightPx) * 100);
    const thumbTopPercent = (100 - thumbHeightPercent) * contentOffsetPx / maxOffsetPx;
    return { visible: true, thumbTopPercent, thumbHeightPercent, contentOffsetPx, dragging };
  }

  // Возвращает максимальный сдвиг списка действий.
  private settingsInputMaxScrollOffset(): number {
    return Math.max(0, this.settingsInputContentHeightPx() - this.settingsInputViewportHeightPx());
  }

  // Возвращает максимальный сдвиг раскрытого списка событий.
  private settingsDropdownMaxScrollOffset(): number {
    return Math.max(0, this.settingsDropdownContentHeightPx() - this.settingsDropdownViewportHeightPx());
  }

  // Возвращает максимальный сдвиг списка групп оборудования.
  private controlPanelEquipmentListMaxScrollOffset(): number {
    return Math.max(0, this.controlPanelEquipmentListContentHeightPx() - this.controlPanelEquipmentListViewportHeightPx());
  }

  // Возвращает максимальный сдвиг обычного списка.
  private listMaxScrollOffset(listId: string): number {
    return Math.max(0, this.listContentHeightPx(listId) - this.listViewportHeightPx(listId));
  }

  // Измеряет видимую высоту списка действий.
  private settingsInputViewportHeightPx(): number {
    return document.querySelector<HTMLElement>(".settings-input-table__left")?.getBoundingClientRect().height ?? window.innerHeight * 0.31;
  }

  // Измеряет полную высоту самой длинной колонки действий.
  private settingsInputContentHeightPx(): number {
    const rowCount = Object.keys(this.referenceData?.ActionType.Items ?? {}).length;
    return getInputSettingsLeftColumnRowCount(rowCount) * SETTINGS_INPUT_ROW_HEIGHT_VH * window.innerHeight / 100;
  }

  // Измеряет видимую высоту раскрытого списка событий.
  private settingsDropdownViewportHeightPx(): number {
    return SETTINGS_DROPDOWN_VIEWPORT_HEIGHT_VH * window.innerHeight / 100;
  }

  // Измеряет полную высоту раскрытого списка событий.
  private settingsDropdownContentHeightPx(): number {
    const optionCount = Object.keys(this.referenceData?.InputEventType.Items ?? {}).length;
    return (optionCount * SETTINGS_DROPDOWN_ITEM_HEIGHT_VH + SETTINGS_DROPDOWN_CONTENT_PADDING_VH) * window.innerHeight / 100;
  }

  // Измеряет видимую высоту списка групп оборудования.
  private controlPanelEquipmentListViewportHeightPx(): number {
    return document.querySelector<HTMLElement>(".control-panel-equipment-list .ui-kit-list")?.getBoundingClientRect().height ?? window.innerHeight * 0.31;
  }

  // Измеряет полную высоту списка групп оборудования.
  private controlPanelEquipmentListContentHeightPx(): number {
    const rowCount = this.getControlPanelEquipmentGroups().length;
    return (rowCount * CONTROL_PANEL_EQUIPMENT_LIST_ITEM_HEIGHT_VH + SETTINGS_DROPDOWN_CONTENT_PADDING_VH) * window.innerHeight / 100;
  }

  // Проверяет, что ID относится к обычному списку с общей прокруткой.
  private isScrollableListId(listId: string): boolean {
    return scrollableListIds.has(listId);
  }

  // Измеряет видимую высоту обычного списка.
  private listViewportHeightPx(listId: string): number {
    if (listId === "control-panel-equipment-list") {
      return this.controlPanelEquipmentListViewportHeightPx();
    }
    return document.getElementById(listId)?.getBoundingClientRect().height ?? window.innerHeight * 0.2;
  }

  // Возвращает расчетную полную высоту содержимого обычного списка.
  private listContentHeightPx(listId: string): number {
    if (listId === "control-panel-equipment-list") {
      return this.controlPanelEquipmentListContentHeightPx();
    }
    const rowCount = this.listRowCount(listId);
    return (rowCount * CONTROL_PANEL_EQUIPMENT_LIST_ITEM_HEIGHT_VH + SETTINGS_DROPDOWN_CONTENT_PADDING_VH) * window.innerHeight / 100;
  }

  // Возвращает количество строк в обычном списке по его назначению.
  private listRowCount(listId: string): number {
    if (listId === "control-panel-usage-left-container-content") {
      return this.getControlPanelContainerItemGroupCount(this.selectedControlPanelUsageLeftContainerGroupId);
    }
    if (listId === "control-panel-usage-right-container-content") {
      return this.getControlPanelContainerItemGroupCount(this.getControlPanelUsageRightContentContainerGroupId());
    }
    if (listId === "exchange-receiver-list") {
      return this.getControlPanelContainerItemGroupCount(this.currentExchangeReceiverContainerID());
    }
    if (listId === "exchange-source-list") {
      return this.getControlPanelContainerItemGroupCount(this.currentExchangeSourceContainerID());
    }
    if (listId === "exchange-other-queue") {
      return this.exchangeState?.otherQueue.length ?? 0;
    }
    if (listId === "exchange-self-queue") {
      return this.exchangeState?.selfQueue.length ?? 0;
    }
    if (listId === "control-panel-constructor-schema-list") {
      return Object.keys(this.referenceData?.Schema.Items ?? {}).length;
    }
    if (listId === "control-panel-constructor-blueprint-list") {
      return Object.keys(this.referenceData?.Blueprint.Items ?? {}).length;
    }
    if (listId === "control-panel-constructor-main-queue") {
      return this.getControlPanelConstructorProductionJobCount("main");
    }
    if (listId === "control-panel-constructor-required-queue") {
      return this.getControlPanelConstructorProductionJobCount("auxiliary");
    }
    if (listId === "control-panel-container-queue") {
      return this.getControlPanelTaskQueueCount("CargoMovement");
    }
    if (listId === "control-panel-fuel-queue") {
      return this.getControlPanelTaskQueueCount("Fueling");
    }
    if (listId === "control-panel-deconstructor-main-queue") {
      return this.getControlPanelTaskQueueCount("ItemDeconstruction");
    }
    if (listId === "control-panel-deconstructor-required-queue") {
      return 0;
    }
    return 0;
  }

  // Возвращает количество строк очереди выбранного конструктора.
  private getControlPanelConstructorProductionJobCount(queueType: "main" | "auxiliary"): number {
    const constructorID = this.gameUi.state().selectedControlPanelUsageRightEquipmentGroupId;
    if (!constructorID) {
      return 0;
    }
    return this.gameUi.state().constructorProductionJobs.filter((job) => job.constructorEquipmentGroupId === constructorID && job.queueType === queueType).length;
  }

  // Возвращает количество строк очереди заданий выбранного контроллера по акрониму типа работы.
  private getControlPanelTaskQueueCount(taskTypeAcronym: string): number {
    const controllerID = this.gameUi.state().selectedControlPanelUsageRightEquipmentGroupId;
    const taskType = Object.values(this.referenceData?.TaskType?.Items ?? {}).find((item) => item.Acronym === taskTypeAcronym);
    if (!controllerID || !taskType) {
      return 0;
    }
    return (this.gameUi.state().tasks ?? []).filter((task) => task.ControllerEquipmentGroupID === controllerID && task.TaskTypeID === taskType.ID).length;
  }

  // Возвращает количество групп предметов в выбранном контейнере.
  private getControlPanelContainerItemGroupCount(containerGroupId: number | null): number {
    if (!containerGroupId) {
      return 0;
    }
    return this.gameUi.state().itemGroups.filter((group) => group.ContainerEquipmentGroupID === containerGroupId).length;
  }
}

const uiKind = (value: string | undefined): GameUiControlKind => {
  const allowed = new Set<GameUiControlKind>(["edit", "button", "checkbox", "radio", "select", "list", "tree", "tabs", "menu", "modal", "tooltip", "scrollbar", "slider", "stepper", "hotkey", "splitter", "dragItem"]);
  return value && allowed.has(value as GameUiControlKind) ? value as GameUiControlKind : "button";
};

const isControlPanelTabValue = (value: unknown): value is ControlPanelTabValue =>
  value === "object" ||
  value === "equipment" ||
  value === "pilotTools" ||
  value === "schemas" ||
  value === "blueprints" ||
  value === "map";

const isControlPanelEquipmentSubTabValue = (value: unknown): value is ControlPanelEquipmentSubTabValue =>
  value === "setup" ||
  value === "usage";

const isControlPanelConstructorTabValue = (value: unknown): value is ControlPanelConstructorTabValue =>
  value === "items" ||
  value === "objects";

// Возвращает текущую группу, если она ещё доступна, иначе первую доступную.
const scrollableListIds = new Set([
  "control-panel-equipment-list",
  "control-panel-usage-left-container-content",
  "control-panel-usage-right-container-content",
  "control-panel-constructor-schema-list",
  "control-panel-constructor-blueprint-list",
  "control-panel-constructor-main-queue",
  "control-panel-constructor-required-queue",
  "control-panel-container-queue",
  "control-panel-fuel-queue",
  "control-panel-deconstructor-main-queue",
  "control-panel-deconstructor-required-queue",
  "exchange-receiver-list",
  "exchange-source-list",
  "exchange-other-queue",
  "exchange-self-queue",
]);

const controlPanelSelectableTaskQueueTypeAcronyms = new Set([
  "CargoMovement",
  "Fueling",
  "ItemDeconstruction",
]);

// Возвращает ID списка по ID встроенной полосы прокрутки.
const listIdFromScrollbarId = (scrollbarId: string): string | null =>
  scrollbarId.endsWith("-scrollbar") ? scrollbarId.slice(0, -"-scrollbar".length) : null;

const normalizeSelectedControlPanelGroupId = (groups: EquipmentGroup[], selectedGroupId: number | null): number | null =>
  groups.some((group) => group.ID === selectedGroupId) ? selectedGroupId : groups[0]?.ID ?? null;

const clamp = (value: number, min: number, max: number): number => Math.min(max, Math.max(min, value));

const lerp = (start: number, end: number, progress: number): number => start + (end - start) * progress;

// Нормализует экранное направление и защищает от нулевой длины отрезка.
const normalizeScreenVector = (vector: DrillBeamPoint): DrillBeamPoint => {
  const length = Math.hypot(vector.x, vector.y);
  if (length <= Number.EPSILON) {
    return { x: 0, y: -1 };
  }

  return {
    x: vector.x / length,
    y: vector.y / length,
  };
};
