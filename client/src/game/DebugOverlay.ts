import { formatNumber } from "../domain/format";
import type { ConnectionStatus, CosmicObject } from "../network/protocol";

// Намеренно остается DOM-слоем, чтобы не смешивать отладочный UI с игровым canvas.
export class DebugOverlay {
  constructor(
    // DOM-узел, в котором показываются отладочные значения.
    private readonly element: HTMLElement,
  ) {}

  // Перезаписывает текстовый отладочный блок последними сетевыми и физическими данными.
  update(
    status: ConnectionStatus,
    selfObject: CosmicObject | null,
    fps: number,
    zoom: number,
  ): void {
    if (!selfObject) {
      this.element.textContent = [
        `Статус: ${status}`,
        "Ожидание подключения к серверу",
        `Зум: ${formatNumber(zoom, 2)}`,
        `FPS: ${formatNumber(fps, 0)}`,
      ].join("\n");
      return;
    }

    const speed = Math.hypot(selfObject.VelocityX, selfObject.VelocityY);

    this.element.textContent = [
      `Статус: ${status}`,
      `ID своего объекта: ${selfObject.ID}`,
      `X: ${formatNumber(selfObject.X)} м`,
      `Y: ${formatNumber(selfObject.Y)} м`,
      `Скорость: ${formatNumber(speed)} м/с`,
      `Угол: ${formatNumber(selfObject.Rotation, 4)} рад`,
      `Угл. скорость: ${formatNumber(selfObject.AngularSpeed, 4)} рад/с`,
      `Зум: ${formatNumber(zoom, 2)}`,
      `FPS: ${formatNumber(fps, 0)}`,
    ].join("\n");
  }
}
