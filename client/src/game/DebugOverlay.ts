import type { ShipState } from "../domain/types";
import { formatNumber } from "../domain/format";

// Overlay намеренно остаётся DOM-слоем, чтобы не смешивать отладочный UI с игровым canvas.
export class DebugOverlay {
  constructor(private readonly element: HTMLElement) {}

  update(ship: ShipState, fps: number, zoom: number): void {
    const speed = Math.hypot(ship.velocity.x, ship.velocity.y);

    this.element.textContent = [
      `X: ${formatNumber(ship.position.x)} м`,
      `Y: ${formatNumber(ship.position.y)} м`,
      `Скорость: ${formatNumber(speed)} м/с`,
      `Угол: ${formatNumber(ship.rotation, 4)} рад`,
      `Угл. скорость: ${formatNumber(ship.angularVelocity, 4)} рад/с`,
      `Zoom: ${formatNumber(zoom, 2)}`,
      `FPS: ${formatNumber(fps, 0)}`,
    ].join("\n");
  }
}
