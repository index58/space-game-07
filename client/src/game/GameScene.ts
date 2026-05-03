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
import type { GameUiController } from "../ui/gameUiState";
import { bodyPolygonToPilotScreen } from "./bodyPolygon";
import { InputController } from "./InputController";

const BODY_POLYGON_DEBUG_COLOR = 0x35d7ff;

// Связывает Phaser-отрисовку, сетевой клиент, ввод и SolidJS UI-слой.
export class GameScene extends Phaser.Scene {
  // Спрайты объектов, переиспользуемые между серверными снимками.
  private objectSprites = new Map<number, Phaser.GameObjects.Image>();
  // Векторный слой отладочной отрисовки физических тел.
  private bodyPolygonGraphics!: Phaser.GameObjects.Graphics;
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
  update(_time: number, _deltaMs: number): void {
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

    const status = this.gameClient?.getStatus() ?? "connecting";
    const snapshot = this.gameClient?.getLatestSnapshot() ?? null;
    const selfObject = snapshot?.objects.find((object) => object.ID === snapshot.selfObjectId) ?? null;

    this.zoomScale = getViewportZoomScale(this.zoomLevel, this.scale.height);

    if (status !== "connected" || !snapshot || !selfObject) {
      this.renderWaiting(status);
      this.gameUi.update({
        status,
        selfObject: null,
        textureFilePath: null,
        fps: this.game.loop.actualFps,
        zoom: this.zoomScale,
      });
      return;
    }

    this.waitingText.setVisible(false);
    this.renderWorld(snapshot.objects, selfObject);
    this.gameUi.update({
      status,
      selfObject,
      textureFilePath: this.modelForObject(selfObject)?.TextureFilePath ?? null,
      fps: this.game.loop.actualFps,
      zoom: this.zoomScale,
    });
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
}
