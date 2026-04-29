import { describe, expect, it } from "vitest";
import { createInitialShipState } from "../data/prototypeObjects";
import { getMomentOfInertia, stepShipPhysics, type ShipInput } from "./physics";

// Пустой ввод нужен для проверки автоматического торможения без активной тяги.
const idle: ShipInput = {
  thrustForward: false,
  thrustBackward: false,
  thrustLeft: false,
  thrustRight: false,
  turnClockwise: false,
  turnCounterClockwise: false,
};

describe("ship physics", () => {
  it("считает момент инерции заполненного эллипса", () => {
    expect(getMomentOfInertia(createInitialShipState())).toBeCloseTo(490173.75);
  });

  it("ускоряет корабль вперед вдоль +Y при нулевом угле", () => {
    const next = stepShipPhysics(createInitialShipState(), { ...idle, thrustForward: true }, 1);

    expect(next.velocity.x).toBeCloseTo(0);
    expect(next.velocity.y).toBeCloseTo(8.13069147656849);
  });

  it("тормозит линейную и угловую скорость только когда нет тяги", () => {
    const ship = {
      ...createInitialShipState(),
      velocity: { x: 10, y: 0 },
      angularVelocity: 1,
    };

    const next = stepShipPhysics(ship, idle, 1);

    expect(next.velocity.x).toBeLessThan(10);
    expect(next.angularVelocity).toBe(0);
  });
});
