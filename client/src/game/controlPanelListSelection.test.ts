import { describe, expect, it } from "vitest";
import { applyControlPanelListSelection } from "./controlPanelListSelection";

describe("applyControlPanelListSelection", () => {
  // Проверяет, что Ctrl добавляет строку к текущему выбору.
  it("adds clicked row with ctrl", () => {
    const result = applyControlPanelListSelection({
      orderedIds: [1, 2, 3],
      selectedIds: [1],
      clickedId: 3,
      anchorId: 1,
      action: { ctrlKey: true },
    });

    expect(result).toEqual({ selectedIds: [1, 3], anchorId: 3 });
  });

  // Проверяет, что Ctrl снимает выбор с уже выбранной строки.
  it("removes clicked row with ctrl", () => {
    const result = applyControlPanelListSelection({
      orderedIds: [1, 2, 3],
      selectedIds: [1, 3],
      clickedId: 3,
      anchorId: 1,
      action: { ctrlKey: true },
    });

    expect(result).toEqual({ selectedIds: [1], anchorId: 3 });
  });

  // Проверяет, что Shift выбирает диапазон от опорной строки до кликнутой.
  it("selects range with shift", () => {
    const result = applyControlPanelListSelection({
      orderedIds: [1, 2, 3, 4],
      selectedIds: [1],
      clickedId: 4,
      anchorId: 2,
      action: { shiftKey: true },
    });

    expect(result).toEqual({ selectedIds: [2, 3, 4], anchorId: 2 });
  });
});
