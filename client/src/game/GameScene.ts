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
  CosmicObject,
  CosmicObjectModelReference,
  EquipmentGroup,
  ReferenceDataMessage,
} from "../network/protocol";
import { fetchReferenceData } from "../network/referenceData";
import type { ControlPanelConstructorTabValue, ControlPanelEquipmentSubTabValue, ControlPanelTabValue, GameUiController, GameUiState, SettingsTabValue } from "../ui/gameUiState";
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
import { normalizeControlPanelUsageSelection } from "./controlPanelUsageSelection";
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

// Связывает Phaser-отрисовку, сетевой клиент, ввод и SolidJS UI-слой.
export class GameScene extends Phaser.Scene {
  // Спрайты объектов, переиспользуемые между серверными снимками.
  private objectSprites = new Map<number, Phaser.GameObjects.Image>();
  // Векторный слой отладочной отрисовки физических тел.
  private bodyPolygonGraphics!: Phaser.GameObjects.Graphics;
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
  private openControlPanelUsageSelect: "left" | "right" | "constructorMaterials" | null = null;
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
  // Активная вкладка списка заданий конструктора.
  private selectedControlPanelConstructorTab: ControlPanelConstructorTabValue = "items";
  // ID схемы, выбранной в списке конструктора.
  private selectedControlPanelConstructorSchemaId: number | null = null;
  // ID чертежа, выбранного в списке конструктора.
  private selectedControlPanelConstructorBlueprintId: number | null = null;
  // Показывает окно подтверждения слива топлива из бака.
  private controlPanelFuelDrainDialogOpen = false;
  // Показывает окно подтверждения залива топлива в бак.
  private controlPanelFuelFillDialogOpen = false;
  // Показывает окно подтверждения частичного переноса предметов между контейнерами.
  private controlPanelContainerTransferDialogOpen = false;
  // Максимальное количество предметов для частичного переноса между контейнерами.
  private controlPanelContainerTransferMaxAmount = 0;
  // Источник ожидающего подтверждения переноса между контейнерами.
  private controlPanelContainerTransferSourceGroupId: number | null = null;
  // Получатель ожидающего подтверждения переноса между контейнерами.
  private controlPanelContainerTransferTargetGroupId: number | null = null;
  // Строка содержимого, ожидающая частичного переноса между контейнерами.
  private controlPanelContainerTransferItemGroupIds: number[] = [];
  // Максимальное количество топлива, доступное для залива в бак.
  private controlPanelFuelFillMaxAmount = 0;
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
    this.consumeSettingsWheel();
    const inputSettingsScroll = this.getSettingsInputScrollState();
    const inputSettingsDropdownScroll = this.getSettingsDropdownScrollState();
    const serverSelfObject = snapshot?.objects.find((object) => object.ID === snapshot.selfObjectId) ?? null;
    const effectiveEquipmentGroups = snapshot ? applyControlPanelPendingToEquipmentGroups(snapshot.equipmentGroups ?? [], this.controlPanelPending) : [];
    const selfObject = applyControlPanelPendingToObject(serverSelfObject, this.controlPanelPending);
    this.syncControlPanelUsageSelection(selfObject?.ID ?? null, effectiveEquipmentGroups);
    this.controlPanelFuelFillMaxAmount = this.getControlPanelFuelFillMaxAmount(selfObject, effectiveEquipmentGroups, snapshot?.itemGroups ?? []);
    this.inputController.syncControlPanelObject(selfObject);
    if (this.controlPanelFuelDrainDialogOpen || this.controlPanelFuelFillDialogOpen || this.controlPanelContainerTransferDialogOpen) {
      const maxAmount = this.controlPanelContainerTransferDialogOpen ? this.controlPanelContainerTransferMaxAmount : this.controlPanelFuelFillDialogOpen ? this.controlPanelFuelFillMaxAmount : Math.max(0, selfObject?.Fuel ?? 0);
      this.controlPanelFuelDrainAmount = clamp(this.inputController.getControlPanelFuelDrainAmount(this.controlPanelFuelDrainAmount), 0, maxAmount);
    }
    const controlPanelObjectTitleEditState = this.inputController.getControlPanelObjectTitleEditState(selfObject?.Title ?? "");
    const controlPanelFuelDrainAmountEditState = this.inputController.getControlPanelFuelDrainAmountEditState();
    this.commitControlPanelObjectTitleIfNeeded(serverSelfObject);

    this.zoomScale = getViewportZoomScale(this.zoomLevel, this.scale.height);

    if (status !== "connected" || !snapshot || !selfObject) {
      this.renderWaiting(status);
      this.gameUi.update({
        status,
        selfObject: null,
        objects: snapshot?.objects ?? [],
        equipmentGroups: effectiveEquipmentGroups,
        itemGroups: snapshot?.itemGroups ?? [],
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
        chatContextMenu: null,
        gameCursor: this.inputController.getGameCursor(),
        chatScroll: { visible: false, thumbTopPercent: 0, thumbHeightPercent: 100, contentOffsetPx: 0, dragging: false },
        uiKitShowcaseVisible: this.inputController.isUiKitShowcaseVisible(),
        settingsVisible,
        controlPanelVisible,
        selectedSettingsTab: this.selectedSettingsTab,
        selectedControlPanelTab: this.selectedControlPanelTab,
        selectedControlPanelEquipmentTab: this.selectedControlPanelEquipmentTab,
        selectedControlPanelEquipmentGroupId: this.selectedControlPanelEquipmentGroupId,
        selectedControlPanelUsageLeftContainerGroupId: this.selectedControlPanelUsageLeftContainerGroupId,
        selectedControlPanelUsageRightEquipmentGroupId: this.selectedControlPanelUsageRightEquipmentGroupId,
        openControlPanelUsageSelect: this.openControlPanelUsageSelect,
        selectedControlPanelUsageLeftItemGroupIds: this.selectedControlPanelUsageLeftItemGroupIds,
        selectedControlPanelUsageRightItemGroupIds: this.selectedControlPanelUsageRightItemGroupIds,
        selectedControlPanelConstructorMaterialContainerGroupId: this.selectedControlPanelConstructorMaterialContainerGroupId,
        selectedControlPanelConstructorTab: this.selectedControlPanelConstructorTab,
        selectedControlPanelConstructorSchemaId: this.selectedControlPanelConstructorSchemaId,
        selectedControlPanelConstructorBlueprintId: this.selectedControlPanelConstructorBlueprintId,
        controlPanelFuelDrainDialogOpen: this.controlPanelFuelDrainDialogOpen,
        controlPanelFuelFillDialogOpen: this.controlPanelFuelFillDialogOpen,
        controlPanelContainerTransferDialogOpen: this.controlPanelContainerTransferDialogOpen,
        controlPanelContainerTransferMaxAmount: this.controlPanelContainerTransferMaxAmount,
        controlPanelFuelFillMaxAmount: this.controlPanelFuelFillMaxAmount,
        controlPanelFuelDrainAmount: this.controlPanelFuelDrainAmount,
        controlPanelFuelDrainAmountText: controlPanelFuelDrainAmountEditState.text,
        controlPanelFuelDrainAmountSelectionStart: controlPanelFuelDrainAmountEditState.selectionStart,
        controlPanelFuelDrainAmountSelectionEnd: controlPanelFuelDrainAmountEditState.selectionEnd,
        controlPanelFuelDrainAmountFocused: controlPanelFuelDrainAmountEditState.focused,
        controlPanelEquipmentEnabledDrafts: {},
        controlPanelEquipmentEnabledCountDrafts: {},
        controlPanelEquipmentListScroll: this.getControlPanelEquipmentListScrollState(),
        controlPanelObjectEnabled: this.inputController.getControlPanelObjectEnabled(false),
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
    this.renderWorld(snapshot.objects, selfObject);
    this.gameUi.update({
      status,
      selfObject,
      objects: snapshot.objects,
      equipmentGroups: effectiveEquipmentGroups,
      itemGroups: snapshot.itemGroups ?? [],
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
      chatContextMenu: this.inputController.getChatContextMenu(),
      gameCursor: this.inputController.getGameCursor(),
      chatScroll: this.inputController.getChatScrollState(),
      uiKitShowcaseVisible: this.inputController.isUiKitShowcaseVisible(),
      settingsVisible,
      controlPanelVisible,
      selectedSettingsTab: this.selectedSettingsTab,
      selectedControlPanelTab: this.selectedControlPanelTab,
      selectedControlPanelEquipmentTab: this.selectedControlPanelEquipmentTab,
      selectedControlPanelEquipmentGroupId: this.selectedControlPanelEquipmentGroupId,
      selectedControlPanelUsageLeftContainerGroupId: this.selectedControlPanelUsageLeftContainerGroupId,
      selectedControlPanelUsageRightEquipmentGroupId: this.selectedControlPanelUsageRightEquipmentGroupId,
      openControlPanelUsageSelect: this.openControlPanelUsageSelect,
      selectedControlPanelUsageLeftItemGroupIds: this.selectedControlPanelUsageLeftItemGroupIds,
      selectedControlPanelUsageRightItemGroupIds: this.selectedControlPanelUsageRightItemGroupIds,
      selectedControlPanelConstructorMaterialContainerGroupId: this.selectedControlPanelConstructorMaterialContainerGroupId,
      selectedControlPanelConstructorTab: this.selectedControlPanelConstructorTab,
      selectedControlPanelConstructorSchemaId: this.selectedControlPanelConstructorSchemaId,
      selectedControlPanelConstructorBlueprintId: this.selectedControlPanelConstructorBlueprintId,
      controlPanelFuelDrainDialogOpen: this.controlPanelFuelDrainDialogOpen,
      controlPanelFuelFillDialogOpen: this.controlPanelFuelFillDialogOpen,
      controlPanelContainerTransferDialogOpen: this.controlPanelContainerTransferDialogOpen,
      controlPanelContainerTransferMaxAmount: this.controlPanelContainerTransferMaxAmount,
      controlPanelFuelFillMaxAmount: this.controlPanelFuelFillMaxAmount,
      controlPanelFuelDrainAmount: this.controlPanelFuelDrainAmount,
      controlPanelFuelDrainAmountText: controlPanelFuelDrainAmountEditState.text,
      controlPanelFuelDrainAmountSelectionStart: controlPanelFuelDrainAmountEditState.selectionStart,
      controlPanelFuelDrainAmountSelectionEnd: controlPanelFuelDrainAmountEditState.selectionEnd,
      controlPanelFuelDrainAmountFocused: controlPanelFuelDrainAmountEditState.focused,
      controlPanelEquipmentEnabledDrafts: {},
      controlPanelEquipmentEnabledCountDrafts: {},
      controlPanelEquipmentListScroll: this.getControlPanelEquipmentListScrollState(),
      controlPanelObjectEnabled: this.inputController.getControlPanelObjectEnabled(selfObject.Enabled),
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
    this.bodyPolygonGraphics.clear();
  }

  // Размещает все объекты в экранных координатах камеры пилота.
  private renderWorld(objects: CosmicObject[], selfObject: CosmicObject): void {
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
      activeObjectIds.add(object.ID);

      const sprite = this.getOrCreateObjectSprite(object);
      if (!sprite) {
        continue;
      }
      const screen = worldToPilotScreen({ x: object.X, y: object.Y }, camera);

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
        const clippingViewport = element.closest(".ui-kit-dropdown__menu-viewport, .settings-input-table, .control-panel-equipment-list .ui-kit-list");
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
  private syncControlPanelUsageSelection(objectId: number | null, equipmentGroups: EquipmentGroup[]): void {
    const selection = normalizeControlPanelUsageSelection({
      objectId,
      equipmentGroups,
      referenceData: this.referenceData,
      selection: {
        leftContainerGroupId: this.selectedControlPanelUsageLeftContainerGroupId,
        rightEquipmentGroupId: this.selectedControlPanelUsageRightEquipmentGroupId,
      },
    });

    this.selectedControlPanelUsageLeftContainerGroupId = selection.leftContainerGroupId;
    this.selectedControlPanelUsageRightEquipmentGroupId = selection.rightEquipmentGroupId;
    this.selectedControlPanelConstructorMaterialContainerGroupId = normalizeSelectedControlPanelGroupId(
      equipmentGroups.filter((group) => group.CosmicObjectID === objectId && this.isEquipmentGroupItemtype(group, "Container")),
      this.selectedControlPanelConstructorMaterialContainerGroupId,
    );
  }

  // Возвращает количество топлива, доступное для залива из текущего выбора в левом контейнере.
  private getControlPanelFuelFillMaxAmount(object: CosmicObject | null, equipmentGroups: EquipmentGroup[], itemGroups: GameUiState["itemGroups"]): number {
    return getControlPanelFuelFillMaxAmount({
      object,
      fuelTankGroup: equipmentGroups.find((group) => group.ID === this.selectedControlPanelUsageRightEquipmentGroupId) ?? null,
      itemGroups,
      selectedItemGroupIds: this.selectedControlPanelUsageLeftItemGroupIds,
      referenceData: this.referenceData,
    });
  }

  // Отправляет завершенное редактирование названия объекта, если текст отличается от серверного снимка.
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

  // Применяет накопленные действия общего UI к локальной витрине контролов.
  private consumeGameUiActions(): void {
    let action = this.inputController.consumeGameUiAction();
    while (action) {
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

  // Применяет действия панели управления и не пропускает их в отладочную витрину.
  private consumeControlPanelUiAction(action: GameUiAction): boolean {
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
    if (action.controlId === "control-panel-usage-right-equipment-select") {
      this.openControlPanelUsageSelect = this.openControlPanelUsageSelect === "right" ? null : "right";
      return true;
    }
    if (action.controlId === "control-panel-constructor-material-select") {
      this.openControlPanelUsageSelect = this.openControlPanelUsageSelect === "constructorMaterials" ? null : "constructorMaterials";
      return true;
    }
    if (action.controlId.startsWith("control-panel-usage-left-container-select-") && typeof action.value === "string") {
      const groupId = Number(action.value);
      if (this.getControlPanelEquipmentGroup(groupId)) {
        this.selectedControlPanelUsageLeftContainerGroupId = groupId;
      }
      this.openControlPanelUsageSelect = null;
      return true;
    }
    if (action.controlId.startsWith("control-panel-usage-right-equipment-select-") && typeof action.value === "string") {
      const groupId = Number(action.value);
      if (this.getControlPanelEquipmentGroup(groupId)) {
        this.selectedControlPanelUsageRightEquipmentGroupId = groupId;
      }
      this.openControlPanelUsageSelect = null;
      return true;
    }
    if (action.controlId.startsWith("control-panel-constructor-material-select-") && typeof action.value === "string") {
      const groupId = Number(action.value);
      if (this.getControlPanelEquipmentGroup(groupId)) {
        this.selectedControlPanelConstructorMaterialContainerGroupId = groupId;
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
      this.startControlPanelContainerTransfer(this.selectedControlPanelUsageLeftContainerGroupId, this.selectedControlPanelUsageRightEquipmentGroupId, this.selectedControlPanelUsageLeftItemGroupIds);
      return true;
    }
    if (action.controlId === "control-panel-container-transfer-to-left") {
      this.startControlPanelContainerTransfer(this.selectedControlPanelUsageRightEquipmentGroupId, this.selectedControlPanelUsageLeftContainerGroupId, this.selectedControlPanelUsageRightItemGroupIds);
      return true;
    }
    if (action.controlId === "control-panel-container-transfer-cancel") {
      this.controlPanelContainerTransferDialogOpen = false;
      this.inputController.blurControlPanelFuelDrainAmount();
      return true;
    }
    if (action.controlId === "control-panel-container-transfer-ok") {
      this.controlPanelFuelDrainAmount = clamp(this.inputController.getControlPanelFuelDrainAmount(this.controlPanelFuelDrainAmount), 0, this.controlPanelContainerTransferMaxAmount);
      this.sendControlPanelContainerTransfer(this.controlPanelContainerTransferSourceGroupId, this.controlPanelContainerTransferTargetGroupId, this.controlPanelContainerTransferItemGroupIds, this.controlPanelFuelDrainAmount);
      this.controlPanelContainerTransferDialogOpen = false;
      this.inputController.blurControlPanelFuelDrainAmount();
      return true;
    }
    if (action.controlId === "control-panel-fuel-transfer-to-tank") {
      this.controlPanelFuelFillMaxAmount = this.getControlPanelFuelFillMaxAmount(this.gameUi.state().selfObject, this.gameUi.state().equipmentGroups, this.gameUi.state().itemGroups);
      if (this.controlPanelFuelFillMaxAmount > 0) {
        this.controlPanelFuelFillDialogOpen = true;
        this.controlPanelFuelDrainDialogOpen = false;
        this.controlPanelContainerTransferDialogOpen = false;
        this.controlPanelFuelDrainAmount = this.controlPanelFuelFillMaxAmount;
        this.inputController.setControlPanelFuelDrainAmount(this.controlPanelFuelDrainAmount);
      }
      return true;
    }
    if (action.controlId === "control-panel-fuel-fill-cancel") {
      this.controlPanelFuelFillDialogOpen = false;
      this.inputController.blurControlPanelFuelDrainAmount();
      return true;
    }
    if (action.controlId === "control-panel-fuel-fill-ok") {
      this.controlPanelFuelDrainAmount = clamp(this.inputController.getControlPanelFuelDrainAmount(this.controlPanelFuelDrainAmount), 0, this.controlPanelFuelFillMaxAmount);
      this.sendControlPanelFuelTransfer(this.selectedControlPanelUsageLeftContainerGroupId, this.selectedControlPanelUsageRightEquipmentGroupId, this.selectedControlPanelUsageLeftItemGroupIds, this.controlPanelFuelDrainAmount);
      this.controlPanelFuelFillDialogOpen = false;
      this.inputController.blurControlPanelFuelDrainAmount();
      return true;
    }
    if (action.controlId === "control-panel-fuel-drain-open") {
      this.controlPanelFuelDrainDialogOpen = true;
      this.controlPanelFuelFillDialogOpen = false;
      this.controlPanelContainerTransferDialogOpen = false;
      this.controlPanelFuelDrainAmount = Math.max(0, this.gameUi.state().selfObject?.Fuel ?? 0);
      this.inputController.setControlPanelFuelDrainAmount(this.controlPanelFuelDrainAmount);
      return true;
    }
    if (action.controlId === "control-panel-fuel-drain-cancel") {
      this.controlPanelFuelDrainDialogOpen = false;
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
      this.controlPanelFuelDrainAmount = clamp(this.inputController.getControlPanelFuelDrainAmount(this.controlPanelFuelDrainAmount), 0, Math.max(0, this.gameUi.state().selfObject?.Fuel ?? 0));
      this.sendControlPanelFuelTransfer(this.selectedControlPanelUsageLeftContainerGroupId, this.selectedControlPanelUsageRightEquipmentGroupId, [], this.controlPanelFuelDrainAmount);
      this.controlPanelFuelDrainDialogOpen = false;
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
    const maxFuel = this.controlPanelContainerTransferDialogOpen ? this.controlPanelContainerTransferMaxAmount : this.controlPanelFuelFillDialogOpen ? this.controlPanelFuelFillMaxAmount : Math.max(0, this.gameUi.state().selfObject?.Fuel ?? 0);
    this.controlPanelFuelDrainAmount = clamp(this.controlPanelFuelDrainAmount + delta, 0, maxFuel);
    this.inputController.setControlPanelFuelDrainAmount(this.controlPanelFuelDrainAmount);
  }

  // Меняет количество слива топлива по позиции курсора на полосе.
  private updateControlPanelFuelDrainAmountFromSlider(action: GameUiAction): void {
    if (!action.controlRect || (action.type !== "dragStart" && action.type !== "dragMove")) {
      return;
    }
    const maxFuel = this.controlPanelContainerTransferDialogOpen ? this.controlPanelContainerTransferMaxAmount : this.controlPanelFuelFillDialogOpen ? this.controlPanelFuelFillMaxAmount : Math.max(0, this.gameUi.state().selfObject?.Fuel ?? 0);
    const position = clamp((action.x - action.controlRect.left) / Math.max(1, action.controlRect.width), 0, 1);
    this.controlPanelFuelDrainAmount = Math.round(maxFuel * position);
    this.inputController.setControlPanelFuelDrainAmount(this.controlPanelFuelDrainAmount);
  }

  // Запускает перенос между контейнерами сразу или через окно количества для одной строки.
  private startControlPanelContainerTransfer(sourceContainerEquipmentGroupId: number | null, targetContainerEquipmentGroupId: number | null, itemGroupIds: number[]): void {
    if (itemGroupIds.length !== 1) {
      this.sendControlPanelContainerTransfer(sourceContainerEquipmentGroupId, targetContainerEquipmentGroupId, itemGroupIds);
      return;
    }
    const itemGroup = this.gameUi.state().itemGroups.find((group) => group.ID === itemGroupIds[0]);
    if (!itemGroup || itemGroup.Count <= 0) {
      return;
    }
    this.controlPanelContainerTransferSourceGroupId = sourceContainerEquipmentGroupId;
    this.controlPanelContainerTransferTargetGroupId = targetContainerEquipmentGroupId;
    this.controlPanelContainerTransferItemGroupIds = itemGroupIds;
    this.controlPanelContainerTransferMaxAmount = itemGroup.Count;
    this.controlPanelFuelDrainAmount = itemGroup.Count;
    this.controlPanelContainerTransferDialogOpen = true;
    this.controlPanelFuelDrainDialogOpen = false;
    this.controlPanelFuelFillDialogOpen = false;
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
    if (rightGroup && this.isEquipmentGroupItemtype(rightGroup, "Constructor")) {
      return this.selectedControlPanelConstructorMaterialContainerGroupId;
    }
    return this.selectedControlPanelUsageRightEquipmentGroupId;
  }

  // Отправляет изменение оборудования и кладет его поверх снимков до серверного подтверждения.
  private sendControlPanelEquipmentMutation(groupId: number, update: { enabled?: boolean; enabledCount?: number }): void {
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
        },
      },
    };
  }

  // Отправляет перенос между двумя выбранными контейнерами панели управления.
  private sendControlPanelContainerTransfer(sourceContainerEquipmentGroupId: number | null, targetContainerEquipmentGroupId: number | null, itemGroupIds: number[], amount = 0): void {
    if (!sourceContainerEquipmentGroupId || !targetContainerEquipmentGroupId || sourceContainerEquipmentGroupId === targetContainerEquipmentGroupId || itemGroupIds.length === 0) {
      return;
    }
    this.gameClient?.sendControlPanelContainerTransfer({
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

  // Возвращает серверную группу оборудования по ID из последнего UI-снимка.
  private getControlPanelEquipmentGroup(groupId: number): EquipmentGroup | null {
    return this.getControlPanelEquipmentGroups().find((group) => group.ID === groupId) ?? null;
  }

  // Возвращает выбранную группу или первую доступную группу оборудования текущего объекта.
  private getSelectedControlPanelEquipmentGroup(): EquipmentGroup | null {
    const groups = this.getControlPanelEquipmentGroups();
    return groups.find((group) => group.ID === this.selectedControlPanelEquipmentGroupId) ?? groups[0] ?? null;
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
  private isEquipmentGroupItemtype(group: EquipmentGroup, itemtypeAcronym: string): boolean {
    const itemModel = this.referenceData?.ItemModel.Items[String(group.EquipmentItemModelID)];
    const itemtype = this.referenceData?.Itemtype.Items[String(itemModel?.ItemtypeID)];
    return itemtype?.Acronym === itemtypeAcronym;
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
const normalizeSelectedControlPanelGroupId = (groups: EquipmentGroup[], selectedGroupId: number | null): number | null =>
  groups.some((group) => group.ID === selectedGroupId) ? selectedGroupId : groups[0]?.ID ?? null;

const clamp = (value: number, min: number, max: number): number => Math.min(max, Math.max(min, value));
