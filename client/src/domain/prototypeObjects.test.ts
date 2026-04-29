import { describe, expect, it } from "vitest";
import { SHIP_BAT } from "../data/prototypeObjects";

// Тест фиксирует договорённые коэффициенты перевода старых условных единиц.
describe("prototype object data", () => {
  it("масштабирует массу и тягу ship_bat в физические единицы", () => {
    expect(SHIP_BAT.massKg).toBeCloseTo(7920);
    expect(SHIP_BAT.thrustN).toBeCloseTo(1287901.529888449);
    expect(SHIP_BAT.thrustN / SHIP_BAT.massKg).toBeCloseTo(162.61382953137096);
  });
});
