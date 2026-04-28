# Local Phaser Prototype Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use `superpowers:executing-plans` to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Build the first local browser prototype: a controllable `ship_bat` flying on a tiled star background with pilot camera, Pointer Lock mouse rotation, mouse-wheel zoom, one station, one asteroid, and an always-visible debug overlay.

**Architecture:** Keep simulation state outside Phaser. Phaser only renders sprites, reads input, and applies a pilot-camera transform that keeps the player ship in the lower screen half with its nose up.

**Tech Stack:** TypeScript, Vite, Phaser 4.0.0+, Vitest for unit tests, DOM/HTML debug overlay without SolidJS.

---

## File Structure

- Create: `client/package.json` — npm scripts and dependencies.
- Create: `client/tsconfig.json` — strict TypeScript config.
- Create: `client/index.html` — Vite entry HTML with game root and debug overlay root.
- Create: `client/src/main.ts` — Phaser bootstrapping.
- Create: `client/src/style.css` — full-screen canvas and debug overlay styles.
- Create: `client/src/data/assets.ts` — stable asset keys and public asset paths.
- Create: `client/src/data/prototypeObjects.ts` — temporary prototype object data derived from `shared/data/cosmic_object_models.json`.
- Create: `client/src/domain/types.ts` — simulation-facing types.
- Create: `client/src/domain/physics.ts` — ship force, torque, braking, and integration rules.
- Create: `client/src/domain/camera.ts` — world-to-screen pilot camera transform.
- Create: `client/src/domain/format.ts` — debug number formatting.
- Create: `client/src/game/InputController.ts` — keyboard, wheel, and Pointer Lock input state.
- Create: `client/src/game/DebugOverlay.ts` — DOM overlay updater.
- Create: `client/src/game/GameScene.ts` — thin Phaser scene that renders state and calls simulation.
- Create: `client/src/domain/*.test.ts` — unit tests for scaling, physics, and camera transforms.

---

## Task 1: Client Scaffold

**Files:**
- Create: `client/package.json`
- Create: `client/tsconfig.json`
- Create: `client/index.html`
- Create: `client/src/main.ts`
- Create: `client/src/style.css`

- [ ] **Step 1: Create npm project files**

`client/package.json`:

```json
{
  "name": "space-game-07-client",
  "private": true,
  "type": "module",
  "scripts": {
    "dev": "vite --host 127.0.0.1",
    "build": "tsc && vite build",
    "test": "vitest run"
  },
  "dependencies": {
    "@vitejs/plugin-basic-ssl": "^1.1.0",
    "phaser": "^4.0.0",
    "vite": "^5.3.0"
  },
  "devDependencies": {
    "typescript": "^5.4.0",
    "vitest": "^1.6.0"
  }
}
```

`client/tsconfig.json`:

```json
{
  "compilerOptions": {
    "target": "ES2022",
    "useDefineForClassFields": true,
    "module": "ESNext",
    "lib": ["ES2022", "DOM", "DOM.Iterable"],
    "skipLibCheck": true,
    "moduleResolution": "Bundler",
    "allowImportingTsExtensions": false,
    "resolveJsonModule": true,
    "isolatedModules": true,
    "noEmit": true,
    "strict": true,
    "noUnusedLocals": true,
    "noUnusedParameters": true
  },
  "include": ["src"]
}
```

- [ ] **Step 2: Create HTML and CSS**

`client/index.html`:

```html
<!doctype html>
<html lang="ru">
  <head>
    <meta charset="UTF-8" />
    <meta name="viewport" content="width=device-width, initial-scale=1.0" />
    <title>Space Game 07</title>
  </head>
  <body>
    <div id="game-root"></div>
    <div id="debug-overlay"></div>
    <script type="module" src="/src/main.ts"></script>
  </body>
</html>
```

`client/src/style.css`:

```css
html,
body,
#game-root {
  width: 100%;
  height: 100%;
  margin: 0;
  overflow: hidden;
  background: #000;
}

#debug-overlay {
  position: fixed;
  top: 8px;
  left: 8px;
  z-index: 10;
  min-width: 260px;
  padding: 8px 10px;
  font: 12px/1.4 Consolas, monospace;
  color: #d8f3ff;
  background: rgba(0, 8, 16, 0.78);
  border: 1px solid rgba(130, 210, 255, 0.35);
  pointer-events: none;
  white-space: pre;
}
```

- [ ] **Step 3: Add Phaser bootstrap**

`client/src/main.ts`:

```ts
import Phaser from "phaser";
import "./style.css";
import { GameScene } from "./game/GameScene";

const game = new Phaser.Game({
  type: Phaser.AUTO,
  parent: "game-root",
  backgroundColor: "#000000",
  scale: {
    mode: Phaser.Scale.RESIZE,
    width: window.innerWidth,
    height: window.innerHeight,
  },
  render: {
    pixelArt: false,
    antialias: true,
  },
  scene: [GameScene],
});

// Глобальная ссылка нужна только для корректного уничтожения Phaser при горячей перезагрузке Vite.
if (import.meta.hot) {
  import.meta.hot.dispose(() => game.destroy(true));
}
```

- [ ] **Step 4: Install dependencies**

Run:

```powershell
Set-Location client
npm install
```

Expected: dependencies installed and `client/package-lock.json` created. If network access fails in sandbox, rerun with escalation.

- [ ] **Step 5: Verify scaffold**

Run:

```powershell
Set-Location client
npm run build
```

Expected: TypeScript build reaches missing-module errors only until later tasks create `GameScene`; after Task 6 it must pass.

---

## Task 2: Prototype Data and Asset Manifest

**Files:**
- Create: `client/src/data/assets.ts`
- Create: `client/src/data/prototypeObjects.ts`
- Create: `client/src/domain/types.ts`
- Test: `client/src/domain/prototypeObjects.test.ts`

- [ ] **Step 1: Define domain types**

`client/src/domain/types.ts`:

```ts
export type CosmicObjectKind = "ship" | "asteroid" | "station";

export type WorldVector = {
  x: number;
  y: number;
};

export type CosmicObjectModel = {
  acronym: string;
  titleRu: string;
  kind: CosmicObjectKind;
  textureKey: string;
  texturePath: string;
  textureWidth: number;
  textureHeight: number;
  textureBodyOriginX: number;
  textureBodyOriginY: number;
  textureBodyWidth: number;
  textureBodyLength: number;
  textureScale: number;
  massKg: number;
  thrustN: number;
  torqueNm: number;
};

export type SimObject = {
  model: CosmicObjectModel;
  position: WorldVector;
  rotation: number;
};

export type ShipState = SimObject & {
  velocity: WorldVector;
  angularVelocity: number;
};
```

- [ ] **Step 2: Define stable asset keys**

`client/src/data/assets.ts`:

```ts
export const ASSET_KEYS = {
  background: "world.background.space",
  shipBat: "world.ship.ship_bat",
  asteroid0002: "world.asteroid.asteroid_0002",
  stationTinyCrumb: "world.station.station_tiny_crumb",
} as const;

export const ASSET_PATHS = {
  [ASSET_KEYS.background]: "/assets/world/backgrounds/space-background.jpg",
  [ASSET_KEYS.shipBat]: "/assets/world/cosmic-objects/ships/256x512/ship_256x512_0008.png",
  [ASSET_KEYS.asteroid0002]: "/assets/world/cosmic-objects/asteroids/asteroid_0002.png",
  [ASSET_KEYS.stationTinyCrumb]: "/assets/world/cosmic-objects/stations/station_0064.png",
} as const;
```

- [ ] **Step 3: Encode prototype object models**

`client/src/data/prototypeObjects.ts`:

```ts
import { ASSET_KEYS, ASSET_PATHS } from "./assets";
import type { CosmicObjectModel, ShipState, SimObject } from "../domain/types";

const TEXTURE_SCALE = 4;

export const SHIP_BAT: CosmicObjectModel = {
  acronym: "ship_bat",
  titleRu: "Летучая мышь",
  kind: "ship",
  textureKey: ASSET_KEYS.shipBat,
  texturePath: ASSET_PATHS[ASSET_KEYS.shipBat],
  textureWidth: 256,
  textureHeight: 512,
  textureBodyOriginX: 126,
  textureBodyOriginY: 259,
  textureBodyWidth: 88,
  textureBodyLength: 90,
  textureScale: TEXTURE_SCALE,
  massKg: 7.92 * 1000,
  thrustN: 0.006439507649442245 * 10000000,
  torqueNm: 653.565 * 1000,
};

export const ASTEROID_0002: CosmicObjectModel = {
  acronym: "asteroid_0002",
  titleRu: "Астероид",
  kind: "asteroid",
  textureKey: ASSET_KEYS.asteroid0002,
  texturePath: ASSET_PATHS[ASSET_KEYS.asteroid0002],
  textureWidth: 2048,
  textureHeight: 2048,
  textureBodyOriginX: 988,
  textureBodyOriginY: 1289,
  textureBodyWidth: 804,
  textureBodyLength: 783,
  textureScale: TEXTURE_SCALE,
  massKg: 629.532 * 1000,
  thrustN: 0,
  torqueNm: 0,
};

export const STATION_TINY_CRUMB: CosmicObjectModel = {
  acronym: "station_tiny_crumb",
  titleRu: "Крошка",
  kind: "station",
  textureKey: ASSET_KEYS.stationTinyCrumb,
  texturePath: ASSET_PATHS[ASSET_KEYS.stationTinyCrumb],
  textureWidth: 2048,
  textureHeight: 2048,
  textureBodyOriginX: 996,
  textureBodyOriginY: 738,
  textureBodyWidth: 225,
  textureBodyLength: 825,
  textureScale: TEXTURE_SCALE,
  massKg: 185.625 * 1000,
  thrustN: 0,
  torqueNm: 0,
};

export const createInitialShipState = (): ShipState => ({
  model: SHIP_BAT,
  position: { x: 0, y: 0 },
  velocity: { x: 0, y: 0 },
  rotation: 0,
  angularVelocity: 0,
});

export const STATIC_OBJECTS: SimObject[] = [
  { model: ASTEROID_0002, position: { x: -500, y: 800 }, rotation: 0 },
  { model: STATION_TINY_CRUMB, position: { x: 500, y: 500 }, rotation: 0 },
];
```

- [ ] **Step 4: Test scaling**

`client/src/domain/prototypeObjects.test.ts`:

```ts
import { describe, expect, it } from "vitest";
import { SHIP_BAT } from "../data/prototypeObjects";

describe("prototype object data", () => {
  it("масштабирует массу и тягу ship_bat в физические единицы", () => {
    expect(SHIP_BAT.massKg).toBeCloseTo(7920);
    expect(SHIP_BAT.thrustN).toBeCloseTo(64395.07649442245);
    expect(SHIP_BAT.thrustN / SHIP_BAT.massKg).toBeCloseTo(8.13069147656849);
  });
});
```

- [ ] **Step 5: Run test**

Run:

```powershell
Set-Location client
npm test -- prototypeObjects
```

Expected: PASS.

---

## Task 3: Physics System

**Files:**
- Create: `client/src/domain/physics.ts`
- Test: `client/src/domain/physics.test.ts`

- [ ] **Step 1: Define physics inputs and helpers**

`client/src/domain/physics.ts`:

```ts
import type { ShipState, WorldVector } from "./types";

export type ShipInput = {
  thrustForward: boolean;
  thrustBackward: boolean;
  thrustLeft: boolean;
  thrustRight: boolean;
  turnClockwise: boolean;
  turnCounterClockwise: boolean;
};

export const EPSILON = 0.000001;

export const getBodySizeMeters = (model: ShipState["model"]) => ({
  width: model.textureBodyWidth / model.textureScale,
  length: model.textureBodyLength / model.textureScale,
});

export const getMomentOfInertia = (ship: ShipState): number => {
  const body = getBodySizeMeters(ship.model);
  return (ship.model.massKg * (body.width ** 2 + body.length ** 2)) / 16;
};

export const getForwardVector = (rotation: number): WorldVector => ({
  x: Math.sin(rotation),
  y: Math.cos(rotation),
});

export const getRightVector = (rotation: number): WorldVector => ({
  x: Math.cos(rotation),
  y: -Math.sin(rotation),
});
```

- [ ] **Step 2: Implement integration**

Append to `physics.ts`:

```ts
const brakeValue = (value: number, acceleration: number, dtSeconds: number): number => {
  const delta = acceleration * dtSeconds;
  if (Math.abs(value) <= delta) {
    return 0;
  }
  return value - Math.sign(value) * delta;
};

export const hasAnyThrust = (input: ShipInput): boolean =>
  input.thrustForward ||
  input.thrustBackward ||
  input.thrustLeft ||
  input.thrustRight ||
  input.turnClockwise ||
  input.turnCounterClockwise;

export const stepShipPhysics = (
  ship: ShipState,
  input: ShipInput,
  dtSeconds: number,
): ShipState => {
  const forward = getForwardVector(ship.rotation);
  const right = getRightVector(ship.rotation);
  const along = (input.thrustForward ? 1 : 0) - (input.thrustBackward ? 1 : 0);
  const across = (input.thrustRight ? 1 : 0) - (input.thrustLeft ? 1 : 0);
  const torqueDirection = (input.turnClockwise ? 1 : 0) - (input.turnCounterClockwise ? 1 : 0);

  let velocityX = ship.velocity.x;
  let velocityY = ship.velocity.y;
  let angularVelocity = ship.angularVelocity;

  if (hasAnyThrust(input)) {
    const forceX = forward.x * along * ship.model.thrustN + right.x * across * ship.model.thrustN;
    const forceY = forward.y * along * ship.model.thrustN + right.y * across * ship.model.thrustN;
    velocityX += (forceX / ship.model.massKg) * dtSeconds;
    velocityY += (forceY / ship.model.massKg) * dtSeconds;
    angularVelocity += (torqueDirection * ship.model.torqueNm / getMomentOfInertia(ship)) * dtSeconds;
  } else {
    const speed = Math.hypot(velocityX, velocityY);
    if (speed > EPSILON) {
      const brakeDelta = Math.min((ship.model.thrustN / ship.model.massKg) * dtSeconds, speed);
      velocityX -= (velocityX / speed) * brakeDelta;
      velocityY -= (velocityY / speed) * brakeDelta;
    }
    angularVelocity = brakeValue(
      angularVelocity,
      ship.model.torqueNm / getMomentOfInertia(ship),
      dtSeconds,
    );
  }

  return {
    ...ship,
    position: {
      x: ship.position.x + velocityX * dtSeconds,
      y: ship.position.y + velocityY * dtSeconds,
    },
    velocity: { x: velocityX, y: velocityY },
    rotation: ship.rotation + angularVelocity * dtSeconds,
    angularVelocity,
  };
};
```

- [ ] **Step 3: Test physics**

`client/src/domain/physics.test.ts`:

```ts
import { describe, expect, it } from "vitest";
import { createInitialShipState } from "../data/prototypeObjects";
import { getMomentOfInertia, stepShipPhysics, type ShipInput } from "./physics";

const idle: ShipInput = {
  thrustForward: false,
  thrustBackward: false,
  thrustLeft: false,
  thrustRight: false,
  turnClockwise: false,
  turnCounterClockwise: false,
};

describe("ship physics", () => {
  it("считает момент инерции заполненного эллипса", () => {
    expect(getMomentOfInertia(createInitialShipState())).toBeCloseTo(490173.75);
  });

  it("ускоряет корабль вперед вдоль +Y при нулевом угле", () => {
    const next = stepShipPhysics(createInitialShipState(), { ...idle, thrustForward: true }, 1);
    expect(next.velocity.x).toBeCloseTo(0);
    expect(next.velocity.y).toBeCloseTo(8.13069147656849);
  });

  it("тормозит линейную и угловую скорость только когда нет тяги", () => {
    const ship = {
      ...createInitialShipState(),
      velocity: { x: 10, y: 0 },
      angularVelocity: 1,
    };
    const next = stepShipPhysics(ship, idle, 1);
    expect(next.velocity.x).toBeLessThan(10);
    expect(next.angularVelocity).toBe(0);
  });
});
```

- [ ] **Step 4: Run physics tests**

Run:

```powershell
Set-Location client
npm test -- physics
```

Expected: PASS.

---

## Task 4: Pilot Camera Transform

**Files:**
- Create: `client/src/domain/camera.ts`
- Test: `client/src/domain/camera.test.ts`

- [ ] **Step 1: Implement camera transform**

`client/src/domain/camera.ts`:

```ts
import type { WorldVector } from "./types";

export type PilotCamera = {
  shipPosition: WorldVector;
  shipRotation: number;
  zoom: number;
  viewportWidth: number;
  viewportHeight: number;
};

export const MIN_ZOOM = 0.01;
export const MAX_ZOOM = 100;
export const INITIAL_ZOOM = 4;

export const clampZoom = (zoom: number): number => Math.min(MAX_ZOOM, Math.max(MIN_ZOOM, zoom));

export const getPilotShipScreenPosition = (viewportWidth: number, viewportHeight: number): WorldVector => ({
  x: viewportWidth / 2,
  y: viewportHeight * 0.75,
});

export const worldToPilotScreen = (worldPosition: WorldVector, camera: PilotCamera): WorldVector => {
  const dx = worldPosition.x - camera.shipPosition.x;
  const dy = worldPosition.y - camera.shipPosition.y;
  const cos = Math.cos(camera.shipRotation);
  const sin = Math.sin(camera.shipRotation);
  const localRight = dx * cos - dy * sin;
  const localForward = dx * sin + dy * cos;
  const shipScreen = getPilotShipScreenPosition(camera.viewportWidth, camera.viewportHeight);

  return {
    x: shipScreen.x + localRight * camera.zoom,
    y: shipScreen.y - localForward * camera.zoom,
  };
};

export const rotationToPilotScreen = (objectRotation: number, shipRotation: number): number =>
  objectRotation - shipRotation;
```

- [ ] **Step 2: Test camera transform**

`client/src/domain/camera.test.ts`:

```ts
import { describe, expect, it } from "vitest";
import {
  INITIAL_ZOOM,
  clampZoom,
  getPilotShipScreenPosition,
  rotationToPilotScreen,
  worldToPilotScreen,
} from "./camera";

describe("pilot camera", () => {
  it("держит корабль в центре нижней половины экрана", () => {
    expect(getPilotShipScreenPosition(1200, 800)).toEqual({ x: 600, y: 600 });
  });

  it("переводит точку впереди корабля вверх экрана", () => {
    const point = worldToPilotScreen(
      { x: 0, y: 100 },
      { shipPosition: { x: 0, y: 0 }, shipRotation: 0, zoom: INITIAL_ZOOM, viewportWidth: 1200, viewportHeight: 800 },
    );
    expect(point).toEqual({ x: 600, y: 200 });
  });

  it("ограничивает zoom диапазоном 0.01..100", () => {
    expect(clampZoom(0)).toBe(0.01);
    expect(clampZoom(1000)).toBe(100);
  });

  it("оставляет корабль носом вверх через относительный поворот", () => {
    expect(rotationToPilotScreen(Math.PI / 2, Math.PI / 2)).toBe(0);
  });
});
```

- [ ] **Step 3: Run camera tests**

Run:

```powershell
Set-Location client
npm test -- camera
```

Expected: PASS.

---

## Task 5: Input and Debug Overlay

**Files:**
- Create: `client/src/game/InputController.ts`
- Create: `client/src/game/DebugOverlay.ts`
- Create: `client/src/domain/format.ts`

- [ ] **Step 1: Add input controller**

`client/src/game/InputController.ts`:

```ts
import { clampZoom } from "../domain/camera";
import type { ShipInput } from "../domain/physics";

export class InputController {
  private readonly keys: Record<string, boolean> = {};
  private mouseDeltaX = 0;
  private zoom = 4;

  constructor(private readonly canvas: HTMLCanvasElement) {
    window.addEventListener("keydown", (event) => {
      this.keys[event.code] = true;
    });
    window.addEventListener("keyup", (event) => {
      this.keys[event.code] = false;
    });
    window.addEventListener("mousemove", (event) => {
      if (document.pointerLockElement === this.canvas) {
        this.mouseDeltaX += event.movementX;
      }
    });
    window.addEventListener("wheel", (event) => {
      this.zoom = clampZoom(this.zoom * (event.deltaY > 0 ? 0.9 : 1.1));
    }, { passive: true });
    this.canvas.addEventListener("click", () => {
      void this.canvas.requestPointerLock();
    });
  }

  getZoom(): number {
    return this.zoom;
  }

  consumeShipInput(): ShipInput {
    const turnClockwise = this.mouseDeltaX > 0;
    const turnCounterClockwise = this.mouseDeltaX < 0;
    this.mouseDeltaX = 0;

    return {
      thrustForward: Boolean(this.keys.KeyW),
      thrustBackward: Boolean(this.keys.KeyS),
      thrustLeft: Boolean(this.keys.KeyA),
      thrustRight: Boolean(this.keys.KeyD),
      turnClockwise,
      turnCounterClockwise,
    };
  }
}
```

- [ ] **Step 2: Add debug formatting and overlay**

`client/src/domain/format.ts`:

```ts
export const formatNumber = (value: number, digits = 2): string =>
  Number.isFinite(value) ? value.toFixed(digits) : "NaN";
```

`client/src/game/DebugOverlay.ts`:

```ts
import type { ShipState } from "../domain/types";
import { formatNumber } from "../domain/format";

export class DebugOverlay {
  constructor(private readonly element: HTMLElement) {}

  update(ship: ShipState, fps: number, zoom: number): void {
    const speed = Math.hypot(ship.velocity.x, ship.velocity.y);
    this.element.textContent = [
      `X: ${formatNumber(ship.position.x)} м`,
      `Y: ${formatNumber(ship.position.y)} м`,
      `Speed: ${formatNumber(speed)} м/с`,
      `Angle: ${formatNumber(ship.rotation, 4)} рад`,
      `Angular speed: ${formatNumber(ship.angularVelocity, 4)} рад/с`,
      `Zoom: ${formatNumber(zoom, 2)}`,
      `FPS: ${formatNumber(fps, 0)}`,
    ].join("\n");
  }
}
```

- [ ] **Step 3: Run unit tests**

Run:

```powershell
Set-Location client
npm test
```

Expected: all unit tests pass.

---

## Task 6: Phaser Scene Rendering

**Files:**
- Create: `client/src/game/GameScene.ts`
- Modify: `client/src/main.ts`

- [ ] **Step 1: Implement scene preload and create**

`client/src/game/GameScene.ts`:

```ts
import Phaser from "phaser";
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
    this.background = this.add.tileSprite(0, 0, this.scale.width, this.scale.height, ASSET_KEYS.background).setOrigin(0);
    this.shipSprite = this.add.image(0, 0, ASSET_KEYS.shipBat).setOrigin(0.5);
    this.staticSprites = STATIC_OBJECTS.map((object) => ({
      object,
      sprite: this.add.image(0, 0, object.model.textureKey).setOrigin(0.5),
    }));

    const canvas = this.game.canvas;
    const overlay = document.getElementById("debug-overlay");
    if (!overlay) {
      throw new Error("debug-overlay element not found");
    }
    this.inputController = new InputController(canvas);
    this.debugOverlay = new DebugOverlay(overlay);
  }
```

- [ ] **Step 2: Implement update and render transform**

Append to `GameScene.ts`:

```ts
  update(_time: number, deltaMs: number): void {
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
```

- [ ] **Step 3: Run build**

Run:

```powershell
Set-Location client
npm run build
```

Expected: TypeScript and Vite build pass.

---

## Task 7: Browser Verification

**Files:**
- No source files unless verification exposes issues.

- [ ] **Step 1: Start dev server**

Run:

```powershell
Set-Location client
npm run dev
```

Expected: Vite prints a local URL.

- [ ] **Step 2: Manual playtest**

Check:
- Page opens with star background.
- Click enters Pointer Lock.
- `W/S/A/D` move ship in local axes.
- Mouse movement rotates world while ship remains nose-up.
- Ship remains at center of lower half of screen.
- Mouse wheel changes zoom between `0.01` and `100`.
- One asteroid and one station are visible when flying around.
- Debug overlay always shows coordinates, speed, angle, angular speed, zoom, FPS.

- [ ] **Step 3: Final verification commands**

Run:

```powershell
Set-Location client
npm test
npm run build
```

Expected: both commands exit with code `0`.

---

## Review Notes

Spec coverage:
- Local-only prototype: covered by client-only Vite/Phaser tasks.
- Ship `ship_bat`: covered by prototype data.
- One asteroid and one station: covered by `STATIC_OBJECTS`.
- Physics: covered by `stepShipPhysics` tests.
- Pilot camera: covered by camera transform tests and scene render transform.
- Pointer Lock and wheel zoom: covered by `InputController`.
- Debug overlay always visible: covered by `DebugOverlay` and CSS.

Intentional omissions for this stage:
- Server, WebSocket, WebTransport, protobuf.
- SolidJS UI.
- Persistence and JSON table migration.
- Combat, mining, inventory, construction.
