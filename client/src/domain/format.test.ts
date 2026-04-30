import { describe, expect, it } from "vitest";
import { formatNumber } from "./format";

// Отладочная панель должна получать предсказуемые строки независимо от текущей локали браузера.
describe("formatNumber", () => {
  it("форматирует конечные числа с заданной точностью", () => {
    expect(formatNumber(12.3456, 2)).toBe("12.35");
    expect(formatNumber(12.3456, 0)).toBe("12");
  });

  it("явно показывает некорректное число", () => {
    expect(formatNumber(Number.NaN)).toBe("NaN");
  });
});
