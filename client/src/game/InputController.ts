import { INITIAL_ZOOM, clampZoom } from "../domain/camera";
import type { ClientInputState } from "../network/protocol";
import { toShipInput } from "./inputState";

// Изолирует браузерные события ввода от игровой сцены.
export class InputController {
  // Текущее состояние клавиш по DOM-кодам.
  private readonly keys: Record<string, boolean> = {};
  // Накопленное горизонтальное движение мыши между кадрами.
  private mouseDeltaX = 0;
  // Дискретный пользовательский уровень приближения.
  private zoom = INITIAL_ZOOM;
  // Одноразовый запрос на смену модели корабля.
  private randomShipChangeRequested = false;

  constructor(
    // Игровой canvas, который получает захват указателя.
    private readonly canvas: HTMLCanvasElement,
    // Проверка готовности сцены к захвату мыши.
    private readonly canRequestPointerLock: () => boolean = () => true,
  ) {
    // Состояние клавиш хранится непрерывно, потому что сетевой ввод отправляется реже кадров браузера.
    window.addEventListener("keydown", (event) => {
      if (event.code === "Backslash" && !this.keys[event.code]) {
        this.randomShipChangeRequested = true;
      }

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

  // Возвращает пользовательский уровень зума без пересчета в пиксели.
  getZoom(): number {
    return this.zoom;
  }

  // Возвращает дискретную команду один раз на одно нажатие клавиши.
  consumeRandomShipChangeRequest(): boolean {
    const requested = this.randomShipChangeRequested;
    this.randomShipChangeRequested = false;
    return requested;
  }

  // Отдает ввод за текущий кадр и сбрасывает накопленное движение мыши.
  consumeShipInput(): ClientInputState {
    // Захват указателя отдает относительное движение мыши; после кадра накопление сбрасывается.
    const isPointerLocked = document.pointerLockElement === this.canvas;
    const input = toShipInput(isPointerLocked, this.keys, this.mouseDeltaX);
    this.mouseDeltaX = 0;

    return input;
  }
}
