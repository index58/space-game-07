import { describe, expect, it } from "vitest";
import { getCountSliderValue } from "./slider";

describe("getCountSliderValue", () => {
  // Проверяет, что значение количества совпадает с долей заполнения шкалы.
  it("maps pointer position to the same fraction used by fill width", () => {
    expect(getCountSliderValue(0.25, 4)).toBe(1);
    expect(getCountSliderValue(0.5, 4)).toBe(2);
    expect(getCountSliderValue(0.75, 4)).toBe(3);
    expect(getCountSliderValue(1, 4)).toBe(4);
  });

  // Проверяет, что количество не уходит ниже одной включенной единицы.
  it("keeps the minimum enabled count at one", () => {
    expect(getCountSliderValue(0, 4)).toBe(1);
    expect(getCountSliderValue(-0.2, 4)).toBe(1);
  });
});
