import { formatNumber } from "../domain/format";
import type { ConnectionStatus, SnapshotObject } from "../network/protocol";

// Overlay намеренно остается DOM-слоем, чтобы не смешивать отладочный UI с игровым canvas.
export class DebugOverlay {
  constructor(private readonly element: HTMLElement) {}

  update(
    status: ConnectionStatus,
    selfObject: SnapshotObject | null,
    fps: number,
    zoom: number,
  ): void {
    if (!selfObject) {
      this.element.textContent = [
        `Status: ${status}`,
        "Ожидание подключения к серверу",
        `Zoom: ${formatNumber(zoom, 2)}`,
        `FPS: ${formatNumber(fps, 0)}`,
      ].join("\n");
      return;
    }

    const speed = Math.hypot(selfObject.velocityX, selfObject.velocityY);

    this.element.textContent = [
      `Status: ${status}`,
      `Self ID: ${selfObject.id}`,
      `X: ${formatNumber(selfObject.x)} м`,
      `Y: ${formatNumber(selfObject.y)} м`,
      `Скорость: ${formatNumber(speed)} м/с`,
      `Угол: ${formatNumber(selfObject.rotation, 4)} рад`,
      `Угл. скорость: ${formatNumber(selfObject.angularVelocity, 4)} рад/с`,
      `Zoom: ${formatNumber(zoom, 2)}`,
      `FPS: ${formatNumber(fps, 0)}`,
    ].join("\n");
  }
}
