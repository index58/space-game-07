import { describe, expect, it } from "vitest";
import { MOUSE_TURN_IMPULSE_SECONDS, MouseTurnImpulse, toShipInput } from "./inputState";

// Эти тесты фиксируют границу между заблокированным системным курсором и игровым вводом.
describe("toShipInput", () => {
  it("не отдаёт управление кораблю без Pointer Lock", () => {
    const input = toShipInput(false, { KeyW: true }, 10);

    expect(input).toEqual({
      thrustForward: false,
      thrustBackward: false,
      thrustLeft: false,
      thrustRight: false,
      turnClockwise: false,
      turnCounterClockwise: false,
    });
  });

  it("начинает поворот при любом ненулевом сдвиге мыши", () => {
    expect(toShipInput(true, {}, Number.EPSILON).turnClockwise).toBe(true);
    expect(toShipInput(true, {}, -Number.EPSILON).turnCounterClockwise).toBe(true);
  });

  it("держит поворот от одиночного сдвига мыши ровно 50 миллисекунд", () => {
    const impulse = new MouseTurnImpulse();

    impulse.addMouseDelta(1);

    expect(impulse.consume(0.016).turnClockwise).toBe(true);
    expect(impulse.consume(0.016).turnClockwise).toBe(true);
    expect(impulse.consume(MOUSE_TURN_IMPULSE_SECONDS - 0.032).turnClockwise).toBe(true);
    expect(impulse.consume(Number.EPSILON).turnClockwise).toBe(false);
  });
});
