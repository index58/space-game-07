import { INITIAL_ZOOM, clampZoom } from "../domain/camera";
import type { ShipInput } from "../domain/physics";
import { toShipInput } from "./inputState";

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
        this.zoom = clampZoom(this.zoom + (event.deltaY > 0 ? -1 : 1));
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
    // Pointer Lock отдает относительное движение мыши; после кадра накопление сбрасывается.
    const isPointerLocked = document.pointerLockElement === this.canvas;
    const input = toShipInput(isPointerLocked, this.keys, this.mouseDeltaX);
    this.mouseDeltaX = 0;

    return input;
  }
}
