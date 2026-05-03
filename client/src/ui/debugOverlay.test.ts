import { describe, expect, it } from "vitest";
import type { CosmicObject } from "../network/protocol";
import { getDebugOverlayLines } from "./debugOverlay";

const selfObject = {
  ID: 7,
  X: 1,
  Y: 2,
  VelocityX: 3,
  VelocityY: 4,
  Rotation: 0.5,
  AngularSpeed: 0.25,
} as CosmicObject;

describe("getDebugOverlayLines", () => {
  it("показывает путь к файлу модели текущего объекта", () => {
    const lines = getDebugOverlayLines({
      status: "connected",
      selfObject,
      textureFilePath: "assets/ships/ship.png",
      fps: 60,
      zoom: 1,
    });

    expect(lines).toContain("Файл объекта: assets/ships/ship.png");
  });
});
