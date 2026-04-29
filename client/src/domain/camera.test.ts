import { describe, expect, it } from "vitest";
import {
  INITIAL_ZOOM,
  clampZoom,
  getPilotShipScreenPosition,
  rotationToPilotScreen,
  worldToPilotScreen,
} from "./camera";

// Камера тестируется отдельно от Phaser, чтобы зафиксировать математическую модель экрана пилота.
describe("pilot camera", () => {
  it("держит корабль в центре нижней половины экрана", () => {
    expect(getPilotShipScreenPosition(1200, 800)).toEqual({ x: 600, y: 600 });
  });

  it("переводит точку впереди корабля вверх экрана", () => {
    const point = worldToPilotScreen(
      { x: 0, y: 100 },
      {
        shipPosition: { x: 0, y: 0 },
        shipRotation: 0,
        zoom: INITIAL_ZOOM,
        viewportWidth: 1200,
        viewportHeight: 800,
      },
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
