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
  ReferenceDataMessage,
} from "../network/protocol";
import { fetchReferenceData } from "../network/referenceData";
import type { GameUiController, GameUiState, SettingsTabValue } from "../ui/gameUiState";
import { getInputBindingMap, getMergedInputSettingValues, toInputSettingsPayload } from "../ui/inputSettings";
import { getNextPilotToolIndex } from "../ui/pilotToolbar";
import { getScrollOffsetFromThumbTopPercent, getScrollbarThumbTopPercentFromCursor, startScrollbarDrag, type ScrollbarDragState } from "../ui-kit/scrollbar";
import { applyUiKitDemoAction, createInitialUiKitDemoState, type UiKitDemoState } from "../ui-kit/showcaseState";
import type { GameUiAction, GameUiControlKind, GameUiControlState } from "../ui-kit/types";
import { bodyPolygonToPilotScreen } from "./bodyPolygon";
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
    this.requestInputSettingsOnOpen(settingsVisible);
    this.syncInputSettingsFromServer();
    const snapshot = this.gameClient?.getLatestSnapshot() ?? null;
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
    const selfObject = snapshot?.objects.find((object) => object.ID === snapshot.selfObjectId) ?? null;

    this.zoomScale = getViewportZoomScale(this.zoomLevel, this.scale.height);

    if (status !== "connected" || !snapshot || !selfObject) {
      this.renderWaiting(status);
      this.gameUi.update({
        status,
        selfObject: null,
        objects: snapshot?.objects ?? [],
        equipmentGroups: snapshot?.equipmentGroups ?? [],
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
        selectedSettingsTab: this.selectedSettingsTab,
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
      equipmentGroups: snapshot.equipmentGroups ?? [],
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
      selectedSettingsTab: this.selectedSettingsTab,
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
        const rect = element.getBoundingClientRect();
        if (rect.width <= 0 || rect.height <= 0) {
          return null;
        }
        const clippingViewport = element.closest(".ui-kit-dropdown__menu-viewport, .settings-input-table");
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

  // Применяет накопленные действия общего UI к локальной витрине контролов.
  private consumeGameUiActions(): void {
    let action = this.inputController.consumeGameUiAction();
    while (action) {
      if (this.inputController.isSettingsVisible() && this.consumeSettingsUiAction(action)) {
        action = this.inputController.consumeGameUiAction();
        continue;
      }
      this.uiKitDemoState = applyUiKitDemoAction(this.uiKitDemoState, action);
      action = this.inputController.consumeGameUiAction();
    }
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

  // Измеряет видимую высоту списка действий.
  private settingsInputViewportHeightPx(): number {
    return document.querySelector<HTMLElement>(".settings-input-table")?.getBoundingClientRect().height ?? window.innerHeight * 0.31;
  }

  // Измеряет полную высоту списка действий.
  private settingsInputContentHeightPx(): number {
    const rowCount = Object.keys(this.referenceData?.ActionType.Items ?? {}).length;
    return rowCount * SETTINGS_INPUT_ROW_HEIGHT_VH * window.innerHeight / 100;
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
}

const uiKind = (value: string | undefined): GameUiControlKind => {
  const allowed = new Set<GameUiControlKind>(["edit", "button", "checkbox", "radio", "select", "list", "tree", "tabs", "menu", "modal", "tooltip", "scrollbar", "slider", "stepper", "hotkey", "splitter", "dragItem"]);
  return value && allowed.has(value as GameUiControlKind) ? value as GameUiControlKind : "button";
};

const clamp = (value: number, min: number, max: number): number => Math.min(max, Math.max(min, value));
