import { describe, expect, it } from "vitest";
import { clipDrillBeamGeometryToPolygons, getDrillBeamGeometry, getDrillBeamIntakeProgress } from "./drillBeam";

describe("getDrillBeamGeometry", () => {
  // Проверяет, что центр мирового объекта луча превращается в экранный отрезок с учетом масштаба.
  it("строит экранный отрезок вокруг центра луча", () => {
    const geometry = getDrillBeamGeometry({
      center: { x: 400, y: 300 },
      rotation: 0,
      lengthMeters: 100,
      zoomScale: 2,
    });

    expect(geometry).toMatchObject({
      start: { x: 400, y: 400 },
      end: { x: 400, y: 200 },
      lengthPx: 200,
      hitObject: false,
    });
    expect(geometry?.widthPx).toBeGreaterThan(0);
  });

  // Проверяет, что направление луча берется из экранного поворота объекта.
  it("поворачивает отрезок вместе с объектом", () => {
    const geometry = getDrillBeamGeometry({
      center: { x: 400, y: 300 },
      rotation: Math.PI / 2,
      lengthMeters: 100,
      zoomScale: 2,
    });

    expect(geometry).toMatchObject({
      start: { x: 300, y: 300 },
      end: { x: 500, y: 300 },
      lengthPx: 200,
    });
  });

  // Проверяет, что видимая толщина луча меняется вместе с масштабом камеры.
  it("масштабирует толщину вместе с экранной длиной", () => {
    const narrow = getDrillBeamGeometry({
      center: { x: 400, y: 300 },
      rotation: 0,
      lengthMeters: 100,
      zoomScale: 1,
    });
    const wide = getDrillBeamGeometry({
      center: { x: 400, y: 300 },
      rotation: 0,
      lengthMeters: 100,
      zoomScale: 3,
    });

    expect(wide?.lengthPx).toBe((narrow?.lengthPx ?? 0) * 3);
    expect(wide?.widthPx).toBe((narrow?.widthPx ?? 0) * 3);
  });

  // Проверяет, что видимый след останавливается на первой границе физического тела перед кораблем.
  it("обрезает отрезок по ближайшему пересеченному телу", () => {
    const geometry = getDrillBeamGeometry({
      center: { x: 0, y: -50 },
      rotation: 0,
      lengthMeters: 100,
      zoomScale: 1,
    });

    expect(geometry).not.toBeNull();
    const clipped = clipDrillBeamGeometryToPolygons(geometry!, [[
      { x: -10, y: -70 },
      { x: 10, y: -70 },
      { x: 10, y: -50 },
      { x: -10, y: -50 },
    ]]);

    expect(clipped.end).toEqual({ x: 0, y: -50 });
    expect(clipped.lengthPx).toBe(50);
    expect(clipped.widthPx).toBe(geometry?.widthPx);
    expect(clipped.hitObject).toBe(true);
  });

  // Проверяет, что внутреннее движение идет от цели к кораблю.
  it("двигает внутренние линии от конца луча к началу", () => {
    const first = getDrillBeamIntakeProgress(0, 0);
    const next = getDrillBeamIntakeProgress(100, 0);

    expect(next).toBeLessThan(first);
  });
});
