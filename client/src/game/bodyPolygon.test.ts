import { describe, expect, it } from "vitest";
import type { CosmicObject, CosmicObjectModelReference } from "../network/protocol";
import { bodyPolygonToPilotScreen, buildBodyPolygon } from "./bodyPolygon";

const model = {
  ID: 1,
  TextureFilePath: "assets/ships/ship.png",
  TextureScale: 4,
  BodyWidth: 10,
  BodyLength: 20,
} satisfies CosmicObjectModelReference;

const object = {
  ID: 1,
  CosmicObjectModelID: 1,
  X: 100,
  Y: 200,
  Rotation: 0,
} as CosmicObject;

describe("buildBodyPolygon", () => {
  it("строит шестнадцать локальных точек по эллипсу серверного тела", () => {
    const points = buildBodyPolygon(model);

    expect(points).toHaveLength(16);
    expect(points[0]).toEqual({ x: 0, y: 10 });
    expect(points[4]).toEqual({ x: 5, y: 0 });
    expect(points[8]).toEqual({ x: 0, y: -10 });
    expect(points[12]).toEqual({ x: -5, y: 0 });
  });
});

describe("bodyPolygonToPilotScreen", () => {
  it("переводит серверное тело в экранные точки камеры пилота", () => {
    const points = bodyPolygonToPilotScreen(object, model, {
      shipPosition: { x: 100, y: 200 },
      shipRotation: 0,
      zoom: 2,
      viewportWidth: 800,
      viewportHeight: 600,
    });

    expect(points[0]).toEqual({ x: 400, y: 430 });
    expect(points[4]).toEqual({ x: 410, y: 450 });
    expect(points[8]).toEqual({ x: 400, y: 470 });
    expect(points[12]).toEqual({ x: 390, y: 450 });
  });
});
