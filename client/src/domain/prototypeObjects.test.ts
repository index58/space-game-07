import { describe, expect, it } from "vitest";
import { SHIP_BAT } from "../data/prototypeObjects";

// Тест фиксирует договорённые коэффициенты перевода старых условных единиц.
describe("prototype object data", () => {
  it("масштабирует массу и тягу ship_bat в физические единицы", () => {
    expect(SHIP_BAT.massKg).toBeCloseTo(7920);
    expect(SHIP_BAT.thrustN).toBeCloseTo(64395.07649442245);
    expect(SHIP_BAT.thrustN / SHIP_BAT.massKg).toBeCloseTo(8.13069147656849);
  });
});
