import { describe, expect, it } from "vitest";
import { MOUSE_TARGET_ROTATION_RADIANS_PER_PIXEL, isFreshKeyboardEventBinding, isFreshKeyDown, toShipInput } from "./inputState";

// Эти тесты фиксируют границу между заблокированным системным курсором и игровым вводом.
describe("toShipInput", () => {
  it("не отдает управление кораблю без захвата указателя", () => {
    const input = toShipInput(false, { KeyW: true }, 10);

    expect(input).toEqual({
      thrustForward: false,
      thrustBackward: false,
      thrustLeft: false,
      thrustRight: false,
      toggleAnchor: false,
      primaryPointerAction: false,
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

  it("использует переназначенную клавишу для продольной тяги", () => {
    const input = toShipInput(true, { ArrowUp: true, KeyW: false }, 0, {
      ThrustForward: "KeyboardEvent.code:ArrowUp",
    });

    expect(input.thrustForward).toBe(true);
    expect(input.thrustBackward).toBe(false);
  });

  // Проверяет, что удержание основной кнопки попадает в игровой ввод только при управлении кораблем.
  it("передает основное действие указателя из удержанной кнопки", () => {
    const activeInput = toShipInput(true, {}, 0, {}, { 0: true });
    const inactiveInput = toShipInput(false, {}, 0, {}, { 0: true });

    expect(activeInput.primaryPointerAction).toBe(true);
    expect(inactiveInput.primaryPointerAction).toBe(false);
  });
});

describe("isFreshKeyDown", () => {
  it("срабатывает только при первом нажатии нужной клавиши", () => {
    expect(isFreshKeyDown("KeyO", false, "KeyO")).toBe(true);
    expect(isFreshKeyDown("KeyO", true, "KeyO")).toBe(false);
    expect(isFreshKeyDown("KeyP", false, "KeyO")).toBe(false);
  });
});

describe("isFreshKeyboardEventBinding", () => {
  // Проверяет, что переназначаемые сочетания клавиш учитывают Alt и Shift.
  it("matches keyboard bindings with modifiers", () => {
    const event = new KeyboardEvent("keydown", { code: "Equal", altKey: true, shiftKey: true });

    expect(isFreshKeyboardEventBinding(event, false, "KeyboardEvent.altKey&&KeyboardEvent.shiftKey&&KeyboardEvent.code:Equal")).toBe(true);
    expect(isFreshKeyboardEventBinding(event, false, "KeyboardEvent.altKey&&KeyboardEvent.code:Equal")).toBe(false);
    expect(isFreshKeyboardEventBinding(event, true, "KeyboardEvent.altKey&&KeyboardEvent.shiftKey&&KeyboardEvent.code:Equal")).toBe(false);
  });
});
