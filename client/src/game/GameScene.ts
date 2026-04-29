import * as Phaser from "phaser";
import { ASSET_KEYS, ASSET_PATHS } from "../data/assets";
import {
  ASTEROID_0002,
  SHIP_BAT,
  STATION_TINY_CRUMB,
} from "../data/prototypeObjects";
import {
  INITIAL_ZOOM,
  getViewportZoomScale,
  getPilotBackgroundTransform,
  rotationToPilotScreen,
  worldToPilotScreen,
} from "../domain/camera";
import type { CosmicObjectModel } from "../domain/types";
import { GameClient } from "../network/GameClient";
import type { ConnectionStatus, SnapshotObject } from "../network/protocol";
import { DebugOverlay } from "./DebugOverlay";
import { InputController } from "./InputController";

const MODELS_BY_ACRONYM: Record<string, CosmicObjectModel> = {
  [SHIP_BAT.acronym]: SHIP_BAT,
  [ASTEROID_0002.acronym]: ASTEROID_0002,
  [STATION_TINY_CRUMB.acronym]: STATION_TINY_CRUMB,
};

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

  preload(): void {
    for (const [key, path] of Object.entries(ASSET_PATHS)) {
      this.load.image(key, path);
    }
  }

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
      const model = MODELS_BY_ACRONYM[object.modelAcronym];
      if (!model) {
        continue;
      }

      const sprite = this.getOrCreateObjectSprite(object, model);
      const screen = worldToPilotScreen({ x: object.x, y: object.y }, camera);

      sprite.setVisible(true);
      sprite.setPosition(screen.x, screen.y);
      sprite.setRotation(object.id === selfObject.id ? 0 : rotationToPilotScreen(object.rotation, selfObject.rotation));
      sprite.setScale(this.zoomScale / model.textureScale);
    }

    for (const [objectId, sprite] of this.objectSprites) {
      if (!activeObjectIds.has(objectId)) {
        sprite.destroy();
        this.objectSprites.delete(objectId);
      }
    }
  }

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

  private getOrCreateObjectSprite(
    object: SnapshotObject,
    model: CosmicObjectModel,
  ): Phaser.GameObjects.Image {
    const existing = this.objectSprites.get(object.id);
    if (existing) {
      return existing;
    }

    const sprite = this.add.image(0, 0, model.textureKey).setOrigin(0.5);
    this.objectSprites.set(object.id, sprite);

    return sprite;
  }
}
