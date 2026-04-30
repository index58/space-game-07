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
import type { ConnectionStatus, SnapshotObject } from "../network/protocol";
import { DebugOverlay } from "./DebugOverlay";
import { InputController } from "./InputController";

// Связывает Phaser-отрисовку, сетевой клиент, ввод и отладочный DOM-слой.
export class GameScene extends Phaser.Scene {
  private objectSprites = new Map<number, Phaser.GameObjects.Image>();
  private background!: Phaser.GameObjects.TileSprite;
  private inputController!: InputController;
  private debugOverlay!: DebugOverlay;
  private gameClient!: GameClient;
  private waitingText!: Phaser.GameObjects.Text;
  private zoomLevel = INITIAL_ZOOM;
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
    const selfObject = snapshot?.objects.find((object) => object.id === snapshot.selfObjectId) ?? null;

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
  private renderWorld(objects: SnapshotObject[], selfObject: SnapshotObject): void {
    const viewportWidth = this.scale.width;
    const viewportHeight = this.scale.height;
    const camera = {
      shipPosition: { x: selfObject.x, y: selfObject.y },
      shipRotation: selfObject.rotation,
      zoom: this.zoomScale,
      viewportWidth,
      viewportHeight,
    };

    this.renderBackground(camera);

    const activeObjectIds = new Set<number>();
    for (const object of objects) {
      activeObjectIds.add(object.id);

      const sprite = this.getOrCreateObjectSprite(object);
      const screen = worldToPilotScreen({ x: object.x, y: object.y }, camera);

      sprite.setVisible(true);
      sprite.setPosition(screen.x, screen.y);
      // Корабль игрока всегда смотрит вверх экрана, остальные объекты вращаются относительно него.
      sprite.setRotation(object.id === selfObject.id ? 0 : rotationToPilotScreen(object.rotation, selfObject.rotation));
      sprite.setScale(this.zoomScale / object.textureScale);
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
    object: SnapshotObject,
  ): Phaser.GameObjects.Image {
    const existing = this.objectSprites.get(object.id);
    if (existing) {
      return existing;
    }

    const sprite = this.add.image(0, 0, this.textureKeyForObject(object)).setOrigin(0.5);
    this.objectSprites.set(object.id, sprite);

    return sprite;
  }

  // Строит ключ ассета из категории объекта и акронима модели.
  private textureKeyForObject(object: SnapshotObject): string {
    return `world.${object.kind}.${object.modelAcronym}`;
  }
}
