import * as Phaser from "phaser";
import {
  ASSET_KEY_BY_COSMIC_OBJECT_MODEL_ID,
  ASSET_KEYS,
  ASSET_PATHS,
  TEXTURE_SCALE_BY_COSMIC_OBJECT_MODEL_ID,
} from "../data/assets";
import {
  INITIAL_ZOOM,
  getViewportZoomScale,
  getPilotBackgroundTransform,
  rotationToPilotScreen,
  worldToPilotScreen,
} from "../domain/camera";
import { GameClient } from "../network/GameClient";
import type { ConnectionStatus, CosmicObject } from "../network/protocol";
import { DebugOverlay } from "./DebugOverlay";
import { InputController } from "./InputController";

// Связывает Phaser-отрисовку, сетевой клиент, ввод и отладочный DOM-слой.
export class GameScene extends Phaser.Scene {
  // Спрайты объектов, переиспользуемые между серверными снимками.
  private objectSprites = new Map<number, Phaser.GameObjects.Image>();
  // Тайловое изображение космоса под всеми объектами.
  private background!: Phaser.GameObjects.TileSprite;
  // Контроллер клавиатуры, мыши и захвата указателя.
  private inputController!: InputController;
  // DOM-слой с диагностикой текущего состояния.
  private debugOverlay!: DebugOverlay;
  // Сетевой клиент для получения снимков и отправки ввода.
  private gameClient!: GameClient;
  // Текст ожидания, показанный до появления валидного снимка.
  private waitingText!: Phaser.GameObjects.Text;
  // Дискретный пользовательский уровень приближения.
  private zoomLevel = INITIAL_ZOOM;
  // Рассчитанный масштаб мира в пикселях на метр.
  private zoomScale = getViewportZoomScale(INITIAL_ZOOM, 1000);

  constructor() {
    super("GameScene");
  }

  // Регистрирует все статические изображения до создания объектов сцены.
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

    const overlay = document.getElementById("debug-overlay");

    if (!overlay) {
      throw new Error("debug-overlay element not found");
    }

    this.gameClient = new GameClient();
    this.inputController = new InputController(
      this.game.canvas,
      () => this.gameClient.getStatus() === "connected",
    );
    this.debugOverlay = new DebugOverlay(overlay);
  }

  // Каждый кадр отправляет свежий ввод и рисует последний серверный снимок мира.
  update(_time: number, _deltaMs: number): void {
    const input = this.inputController.consumeShipInput();
    this.gameClient.setInput(input);
    this.zoomLevel = this.inputController.getZoom();

    const status = this.gameClient.getStatus();
    const snapshot = this.gameClient.getLatestSnapshot();
    const selfObject = snapshot?.objects.find((object) => object.ID === snapshot.selfObjectId) ?? null;

    this.zoomScale = getViewportZoomScale(this.zoomLevel, this.scale.height);

    if (status !== "connected" || !snapshot || !selfObject) {
      this.renderWaiting(status);
      this.debugOverlay.update(status, null, this.game.loop.actualFps, this.zoomScale);
      return;
    }

    this.waitingText.setVisible(false);
    this.renderWorld(snapshot.objects, selfObject);
    this.debugOverlay.update(status, selfObject, this.game.loop.actualFps, this.zoomScale);
  }

  // Показывает фон и статус, пока нет валидного снимка объекта игрока.
  private renderWaiting(status: ConnectionStatus): void {
    this.waitingText.setVisible(true);
    this.waitingText.setPosition(this.scale.width / 2, this.scale.height / 2);
    this.waitingText.setText(
      status === "connecting" ? "Подключение к серверу" : "Ожидание подключения к серверу",
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
  private getOrCreateObjectSprite(
    object: CosmicObject,
  ): Phaser.GameObjects.Image {
    const existing = this.objectSprites.get(object.ID);
    if (existing) {
      return existing;
    }

    const sprite = this.add.image(0, 0, this.textureKeyForObject(object)).setOrigin(0.5);
    this.objectSprites.set(object.ID, sprite);

    return sprite;
  }

  // Возвращает ключ ассета по ID модели из серверных данных.
  private textureKeyForObject(object: CosmicObject): string {
    return ASSET_KEY_BY_COSMIC_OBJECT_MODEL_ID[object.CosmicObjectModelID] ?? ASSET_KEYS.shipBat;
  }

  // Возвращает масштаб текстуры по ID модели из серверных данных.
  private textureScaleForObject(object: CosmicObject): number {
    return TEXTURE_SCALE_BY_COSMIC_OBJECT_MODEL_ID[object.CosmicObjectModelID] ?? 4;
  }
}
