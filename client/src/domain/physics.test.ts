import { describe, expect, it } from "vitest";
import { createInitialShipState } from "../data/prototypeObjects";
import { getMomentOfInertia, stepShipPhysics, type ShipInput } from "./physics";

// Пустой ввод нужен для проверки автоматического торможения без активной линейной тяги.
const idle: ShipInput = {
  thrustForward: false,
  thrustBackward: false,
  thrustLeft: false,
  thrustRight: false,
  targetRotationDelta: 0,
};

describe("ship physics", () => {
  it("считает момент инерции заполненного эллипса", () => {
    expect(getMomentOfInertia(createInitialShipState())).toBeCloseTo(490173.75);
  });

  it("ускоряет корабль вперед вдоль +Y при нулевом угле", () => {
    const next = stepShipPhysics(createInitialShipState(), { ...idle, thrustForward: true }, 1);

    expect(next.velocity.x).toBeCloseTo(0);
    expect(next.velocity.y).toBeCloseTo(162.61382953137096);
  });

  it("ограничивает линейную скорость максимумом модели", () => {
    const next = stepShipPhysics(createInitialShipState(), { ...idle, thrustForward: true }, 10);

    expect(Math.hypot(next.velocity.x, next.velocity.y)).toBeCloseTo(497);
  });

  it("обновляет целевой угол от ввода мыши", () => {
    const next = stepShipPhysics(createInitialShipState(), { ...idle, targetRotationDelta: 0.25 }, 0.016);

    expect(next.targetRotation).toBeCloseTo(0.25);
  });

  it("начинает вращение к целевому углу", () => {
    const ship = {
      ...createInitialShipState(),
      targetRotation: 1,
    };

    const next = stepShipPhysics(ship, idle, 0.05);

    expect(next.angularVelocity).toBeGreaterThan(0);
  });

  it("не нормализует ошибку угла через границу -pi/pi", () => {
    const ship = {
      ...createInitialShipState(),
      rotation: Math.PI - 0.1,
      targetRotation: -Math.PI + 0.1,
    };

    const next = stepShipPhysics(ship, idle, 0.05);

    expect(next.angularVelocity).toBeLessThan(0);
  });

  it("гасит угловую скорость возле целевого угла", () => {
    const ship = {
      ...createInitialShipState(),
      rotation: 1,
      targetRotation: 1,
      angularVelocity: 0.1,
    };

    const next = stepShipPhysics(ship, idle, 1);

    expect(next.angularVelocity).toBe(0);
    expect(next.rotation).toBeCloseTo(1);
  });

  it("останавливает поворот на целевом угле без перескока", () => {
    const ship = {
      ...createInitialShipState(),
      rotation: 0,
      targetRotation: 0.01,
      angularVelocity: 1,
    };

    const next = stepShipPhysics(ship, idle, 0.05);

    expect(next.rotation).toBeCloseTo(0.01);
    expect(next.angularVelocity).toBe(0);
  });

  it("гасит угловую скорость без разгона обратно при малой ошибке угла", () => {
    const ship = {
      ...createInitialShipState(),
      rotation: 0.011,
      targetRotation: 0.01,
      angularVelocity: 0.01,
    };

    const next = stepShipPhysics(ship, idle, 0.05);

    expect(next.rotation).toBeCloseTo(0.01);
    expect(next.angularVelocity).toBe(0);
  });

  it("снижает угловую скорость перед финальной остановкой у целевого угла", () => {
    const dtSeconds = 0.05;
    const ship = {
      ...createInitialShipState(),
      targetRotation: 0.5,
    };
    const maxAngularVelocityChange = (ship.model.torqueNm / getMomentOfInertia(ship)) * dtSeconds;
    let current = ship;
    let angularVelocityBeforeStop = 0;

    for (let step = 0; step < 100; step++) {
      const next = stepShipPhysics(current, idle, dtSeconds);

      if (next.angularVelocity === 0 && next.rotation === next.targetRotation) {
        angularVelocityBeforeStop = Math.abs(current.angularVelocity);
        break;
      }

      current = next;
    }

    expect(angularVelocityBeforeStop).toBeGreaterThan(0);
    expect(angularVelocityBeforeStop).toBeLessThanOrEqual(maxAngularVelocityChange + 0.000001);
  });

  it("ограничивает угловую скорость максимумом модели", () => {
    const ship = {
      ...createInitialShipState(),
      targetRotation: 100,
    };

    const next = stepShipPhysics(ship, idle, 10);

    expect(next.angularVelocity).toBeCloseTo(3);
  });

  it("тормозит линейную и угловую скорость когда цель уже достигнута", () => {
    const ship = {
      ...createInitialShipState(),
      velocity: { x: 10, y: 0 },
      angularVelocity: 1,
    };

    const next = stepShipPhysics(ship, idle, 1);

    expect(next.velocity.x).toBeLessThan(10);
    expect(next.angularVelocity).toBe(0);
  });

  it("не отключает линейное автоторможение при движении к целевому углу", () => {
    const ship = {
      ...createInitialShipState(),
      velocity: { x: 200, y: 0 },
      angularVelocity: 0,
      targetRotation: 1,
    };

    const next = stepShipPhysics(ship, idle, 0.05);

    expect(next.velocity.x).toBeLessThan(200);
    expect(next.angularVelocity).toBeGreaterThan(0);
  });

  it("автоматически тормозит поперечную скорость при активной продольной тяге", () => {
    const ship = {
      ...createInitialShipState(),
      velocity: { x: 100, y: 0 },
    };

    const next = stepShipPhysics(ship, { ...idle, thrustForward: true }, 0.1);

    expect(next.velocity.x).toBeLessThan(100);
    expect(next.velocity.y).toBeGreaterThan(0);
  });

  it("автоматически тормозит продольную скорость при активной поперечной тяге", () => {
    const ship = {
      ...createInitialShipState(),
      velocity: { x: 0, y: 100 },
    };

    const next = stepShipPhysics(ship, { ...idle, thrustRight: true }, 0.1);

    expect(next.velocity.x).toBeGreaterThan(0);
    expect(next.velocity.y).toBeLessThan(100);
  });

  it("не автотормозит продольную ось при одновременных W и S", () => {
    const ship = {
      ...createInitialShipState(),
      velocity: { x: 0, y: 100 },
    };

    const next = stepShipPhysics(ship, { ...idle, thrustForward: true, thrustBackward: true }, 0.1);

    expect(next.velocity.x).toBeCloseTo(0);
    expect(next.velocity.y).toBeCloseTo(100);
  });

  it("не автотормозит поперечную ось при одновременных A и D", () => {
    const ship = {
      ...createInitialShipState(),
      velocity: { x: 100, y: 0 },
    };

    const next = stepShipPhysics(ship, { ...idle, thrustLeft: true, thrustRight: true }, 0.1);

    expect(next.velocity.x).toBeCloseTo(100);
    expect(next.velocity.y).toBeCloseTo(0);
  });
});
