import { describe, expect, it } from "vitest";
import { MOUSE_TARGET_ROTATION_RADIANS_PER_PIXEL, toShipInput } from "./inputState";

// Эти тесты фиксируют границу между заблокированным системным курсором и игровым вводом.
describe("toShipInput", () => {
  it("не отдает управление кораблю без захвата указателя", () => {
    const input = toShipInput(false, { KeyW: true }, 10);

    expect(input).toEqual({
      thrustForward: false,
      thrustBackward: false,
      thrustLeft: false,
      thrustRight: false,
      targetRotationDelta: 0,
    });
  });

  it("сохраняет линейный ввод клавиш при захвате указателя", () => {
    const input = toShipInput(true, { KeyW: true, KeyA: true }, 0);

    expect(input.thrustForward).toBe(true);
    expect(input.thrustLeft).toBe(true);
    expect(input.targetRotationDelta).toBe(0);
  });

  it("превращает сдвиг мыши в изменение целевого угла", () => {
    expect(MOUSE_TARGET_ROTATION_RADIANS_PER_PIXEL).toBe(0.0025);
    expect(toShipInput(true, {}, 3).targetRotationDelta).toBeCloseTo(
      3 * MOUSE_TARGET_ROTATION_RADIANS_PER_PIXEL,
    );
    expect(toShipInput(true, {}, -2).targetRotationDelta).toBeCloseTo(
      -2 * MOUSE_TARGET_ROTATION_RADIANS_PER_PIXEL,
    );
  });
});
