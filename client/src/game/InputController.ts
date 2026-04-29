import { INITIAL_ZOOM, clampZoom } from "../domain/camera";
import type { ShipInput } from "../domain/physics";

export class InputController {
  private readonly keys: Record<string, boolean> = {};
  private mouseDeltaX = 0;
  private zoom = INITIAL_ZOOM;

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

    window.addEventListener(
      "wheel",
      (event) => {
        this.zoom = clampZoom(this.zoom * (event.deltaY > 0 ? 0.9 : 1.1));
      },
      { passive: true },
    );

    this.canvas.addEventListener("click", () => {
      void this.canvas.requestPointerLock();
    });
  }

  getZoom(): number {
    return this.zoom;
  }

  consumeShipInput(): ShipInput {
    // Pointer Lock отдаёт относительное движение мыши; после кадра накопление сбрасывается.
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
