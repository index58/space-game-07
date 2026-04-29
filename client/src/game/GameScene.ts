import * as Phaser from "phaser";
import { ASSET_KEYS, ASSET_PATHS } from "../data/assets";
import { STATIC_OBJECTS, createInitialShipState } from "../data/prototypeObjects";
import { INITIAL_ZOOM, rotationToPilotScreen, worldToPilotScreen } from "../domain/camera";
import { stepShipPhysics } from "../domain/physics";
import type { ShipState, SimObject } from "../domain/types";
import { DebugOverlay } from "./DebugOverlay";
import { InputController } from "./InputController";

export class GameScene extends Phaser.Scene {
  private ship!: ShipState;
  private shipSprite!: Phaser.GameObjects.Image;
  private staticSprites: Array<{ object: SimObject; sprite: Phaser.GameObjects.Image }> = [];
  private background!: Phaser.GameObjects.TileSprite;
  private inputController!: InputController;
  private debugOverlay!: DebugOverlay;
  private zoom = INITIAL_ZOOM;

  constructor() {
    super("GameScene");
  }

  preload(): void {
    for (const [key, path] of Object.entries(ASSET_PATHS)) {
      this.load.image(key, path);
    }
  }

  create(): void {
    this.ship = createInitialShipState();
    this.background = this.add
      .tileSprite(0, 0, this.scale.width, this.scale.height, ASSET_KEYS.background)
      .setOrigin(0);
    this.shipSprite = this.add.image(0, 0, ASSET_KEYS.shipBat).setOrigin(0.5);
    this.staticSprites = STATIC_OBJECTS.map((object) => ({
      object,
      sprite: this.add.image(0, 0, object.model.textureKey).setOrigin(0.5),
    }));

    const overlay = document.getElementById("debug-overlay");

    if (!overlay) {
      throw new Error("debug-overlay element not found");
    }

    this.inputController = new InputController(this.game.canvas);
    this.debugOverlay = new DebugOverlay(overlay);
  }

  update(_time: number, deltaMs: number): void {
    // Ограничиваем шаг, чтобы после паузы вкладки физика не делала огромный скачок.
    const dtSeconds = Math.min(deltaMs / 1000, 0.05);
    const input = this.inputController.consumeShipInput();

    this.zoom = this.inputController.getZoom();
    this.ship = stepShipPhysics(this.ship, input, dtSeconds);
    this.renderWorld();
    this.debugOverlay.update(this.ship, this.game.loop.actualFps, this.zoom);
  }

  private renderWorld(): void {
    const viewportWidth = this.scale.width;
    const viewportHeight = this.scale.height;

    this.background.setSize(viewportWidth, viewportHeight);
    this.background.tilePositionX = this.ship.position.x * this.zoom;
    this.background.tilePositionY = -this.ship.position.y * this.zoom;

    this.renderShip(viewportWidth, viewportHeight);
    this.renderStaticObjects(viewportWidth, viewportHeight);
  }

  private renderShip(viewportWidth: number, viewportHeight: number): void {
    const shipScreen = worldToPilotScreen(this.ship.position, {
      shipPosition: this.ship.position,
      shipRotation: this.ship.rotation,
      zoom: this.zoom,
      viewportWidth,
      viewportHeight,
    });

    this.shipSprite.setPosition(shipScreen.x, shipScreen.y);
    this.shipSprite.setRotation(0);
    this.shipSprite.setScale(this.zoom / this.ship.model.textureScale);
  }

  private renderStaticObjects(viewportWidth: number, viewportHeight: number): void {
    for (const item of this.staticSprites) {
      const screen = worldToPilotScreen(item.object.position, {
        shipPosition: this.ship.position,
        shipRotation: this.ship.rotation,
        zoom: this.zoom,
        viewportWidth,
        viewportHeight,
      });

      item.sprite.setPosition(screen.x, screen.y);
      item.sprite.setRotation(rotationToPilotScreen(item.object.rotation, this.ship.rotation));
      item.sprite.setScale(this.zoom / item.object.model.textureScale);
    }
  }
}
