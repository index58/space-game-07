import { describe, expect, it } from "vitest";
import type { CosmicObject } from "../network/protocol";
import { DebugOverlay } from "./DebugOverlay";

const selfObject = {
  ID: 7,
  X: 1,
  Y: 2,
  VelocityX: 3,
  VelocityY: 4,
  Rotation: 0.5,
  AngularSpeed: 0.25,
} as CosmicObject;

describe("DebugOverlay", () => {
  it("показывает путь к файлу модели текущего объекта", () => {
    const element = { textContent: "" } as HTMLElement;
    const overlay = new DebugOverlay(element);

    overlay.update("connected", selfObject, "assets/ships/ship.png", 60, 1);

    expect(element.textContent).toContain("Файл объекта: assets/ships/ship.png");
  });
});
