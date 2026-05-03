import { describe, expect, it } from "vitest";
import type { CosmicObject } from "../network/protocol";
import { getObjectIndicators } from "./ObjectIndicatorsOverlay";

const selfObject = {
  Armor: 420,
  MaxArmor: 500,
  ConsumingPower: 18,
  GeneratingPower: 24,
  Fuel: 76,
  MaxFuel: 130,
} as CosmicObject;

describe("getObjectIndicators", () => {
  it("готовит основные показатели посещаемого объекта для нижней левой панели", () => {
    const indicators = getObjectIndicators(selfObject);

    expect(indicators).toEqual([
      {
        acronym: "Armor",
        title: "Броня",
        valueText: "420 / 500",
        fillPercent: 84,
      },
      {
        acronym: "Power",
        title: "Энергия",
        valueText: "18 / 24",
        fillPercent: 75,
      },
      {
        acronym: "Fuel",
        title: "Топливо",
        valueText: "76 / 130",
        fillPercent: 58.46153846153847,
      },
    ]);
  });
});
