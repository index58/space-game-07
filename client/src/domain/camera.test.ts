import { describe, expect, it } from "vitest";
import {
  INITIAL_ZOOM,
  BACKGROUND_TEXTURE_SCALE,
  clampZoom,
  getViewportZoomScale,
  getPilotBackgroundTransform,
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
        zoom: getViewportZoomScale(INITIAL_ZOOM, 800),
        viewportWidth: 1200,
        viewportHeight: 800,
      },
    );

    expect(point).toEqual({ x: 600, y: 520 });
  });

  it("ограничивает уровень zoom диапазоном -100..100", () => {
    expect(clampZoom(-1000)).toBe(-100);
    expect(clampZoom(0)).toBe(0);
    expect(clampZoom(1000)).toBe(100);
  });

  it("привязывает нулевой zoom к высоте экрана в 1000 метров", () => {
    expect(getViewportZoomScale(0, 1000)).toBe(1);
    expect(getViewportZoomScale(0, 800)).toBe(0.8);
  });

  it("оставляет корабль носом вверх через относительный поворот", () => {
    expect(rotationToPilotScreen(Math.PI / 2, Math.PI / 2)).toBe(0);
  });

  it("держит звёздный фон в той же системе координат, что и мир", () => {
    const zoomScale = getViewportZoomScale(INITIAL_ZOOM, 800);
    const transform = getPilotBackgroundTransform({
      shipPosition: { x: 100, y: 50 },
      shipRotation: Math.PI / 2,
      zoom: zoomScale,
      viewportWidth: 1200,
      viewportHeight: 800,
    });

    expect(transform.position).toEqual({ x: 600, y: 600 });
    expect(transform.rotation).toBeCloseTo(-Math.PI / 2);
    expect(transform.scale).toBe(zoomScale);
    expect(BACKGROUND_TEXTURE_SCALE).toBe(2);
    expect(transform.tileScale).toBe(1 / BACKGROUND_TEXTURE_SCALE);
    expect(transform.tilePositionX).toBeCloseTo((100 - transform.size / 2) * BACKGROUND_TEXTURE_SCALE);
    expect(transform.tilePositionY).toBeCloseTo((-50 - transform.size / 2) * BACKGROUND_TEXTURE_SCALE);
  });
});
