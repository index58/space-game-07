import { INITIAL_ZOOM, clampZoom } from "../domain/camera";
import type { ClientInputState } from "../network/protocol";
import { toShipInput } from "./inputState";

// изолирует браузерные события ввода от игровой сцены.
export class InputController {
  private readonly keys: Record<string, boolean> = {};
  private mouseDeltaX = 0;
  private zoom = INITIAL_ZOOM;

  constructor(
    private readonly canvas: HTMLCanvasElement,
    private readonly canRequestPointerLock: () => boolean = () => true,
  ) {
    // Состояние клавиш хранится непрерывно, потому что сетевой ввод отправляется реже кадров браузера.
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

    // Колесо мыши меняет дискретный уровень зума, который затем переводится камерой в масштаб.
    window.addEventListener(
      "wheel",
      (event) => {
        this.zoom = clampZoom(this.zoom + (event.deltaY > 0 ? -1 : 1));
      },
      { passive: true },
    );

    // Захват мыши включается только по клику и только когда клиент уже готов принимать управление.
    this.canvas.addEventListener("click", () => {
      if (!this.canRequestPointerLock()) {
        return;
      }

      void this.canvas.requestPointerLock();
    });
  }

  // возвращает пользовательский уровень зума без пересчета в пиксели.
  getZoom(): number {
    return this.zoom;
  }

  // отдает ввод за текущий кадр и сбрасывает накопленное движение мыши.
  consumeShipInput(): ClientInputState {
    // Захват указателя отдает относительное движение мыши; после кадра накопление сбрасывается.
    const isPointerLocked = document.pointerLockElement === this.canvas;
    const input = toShipInput(isPointerLocked, this.keys, this.mouseDeltaX);
    this.mouseDeltaX = 0;

    return input;
  }
}
