import { describe, expect, it } from "vitest";
import { getInputSettingsLeftColumnRowCount } from "./inputSettings";

describe("input settings", () => {
  // Проверяет, что при нечётном количестве строк лишняя строка остаётся в левой половине окна.
  it("keeps the extra row in the left settings column", () => {
    expect(getInputSettingsLeftColumnRowCount(5)).toBe(3);
  });

  // Проверяет, что при чётном количестве строк половины окна получают одинаковое число строк.
  it("splits an even row count equally between settings columns", () => {
    expect(getInputSettingsLeftColumnRowCount(6)).toBe(3);
  });

  // Проверяет, что пустой список настроек не создаёт строк для левой половины окна.
  it("keeps an empty settings list empty", () => {
    expect(getInputSettingsLeftColumnRowCount(0)).toBe(0);
  });
});
