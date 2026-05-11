import type { GameUiAction } from "../ui-kit/types";

export type ControlPanelListSelectionInput = {
  // ID строк списка в текущем визуальном порядке.
  orderedIds: number[];
  // ID строк, выбранных до текущего клика.
  selectedIds: number[];
  // ID строки, по которой кликнул игрок.
  clickedId: number;
  // ID опорной строки для выбора диапазона через Shift.
  anchorId: number | null;
  // Действие игрового UI с клавишами-модификаторами.
  action: Pick<GameUiAction, "ctrlKey" | "metaKey" | "shiftKey">;
};

export type ControlPanelListSelectionResult = {
  // Новый список выбранных ID.
  selectedIds: number[];
  // Новая опорная строка для следующего Shift-выбора.
  anchorId: number;
};

// Применяет обычный, Ctrl и Shift выбор к списку содержимого контейнера.
export const applyControlPanelListSelection = (input: ControlPanelListSelectionInput): ControlPanelListSelectionResult => {
  const clickedId = input.clickedId;
  const anchorId = input.anchorId ?? clickedId;
  const additive = Boolean(input.action.ctrlKey || input.action.metaKey);

  if (input.action.shiftKey) {
    const rangeIds = rangeBetween(input.orderedIds, anchorId, clickedId);
    return {
      selectedIds: additive ? uniqueIds([...input.selectedIds, ...rangeIds]) : rangeIds,
      anchorId,
    };
  }

  if (additive) {
    return {
      selectedIds: input.selectedIds.includes(clickedId)
        ? input.selectedIds.filter((id) => id !== clickedId)
        : [...input.selectedIds, clickedId],
      anchorId: clickedId,
    };
  }

  return { selectedIds: [clickedId], anchorId: clickedId };
};

// Возвращает ID между двумя строками включительно в порядке списка.
const rangeBetween = (orderedIds: number[], anchorId: number, clickedId: number): number[] => {
  const anchorIndex = orderedIds.indexOf(anchorId);
  const clickedIndex = orderedIds.indexOf(clickedId);
  if (anchorIndex < 0 || clickedIndex < 0) {
    return [clickedId];
  }
  const start = Math.min(anchorIndex, clickedIndex);
  const end = Math.max(anchorIndex, clickedIndex);
  return orderedIds.slice(start, end + 1);
};

// Убирает повторы без изменения порядка выбора.
const uniqueIds = (ids: number[]): number[] => Array.from(new Set(ids));
