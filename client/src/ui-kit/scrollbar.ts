export type ScrollbarTrackRect = {
  // Верх полосы прокрутки в пикселях окна.
  top: number;
  // Высота полосы прокрутки в пикселях окна.
  height: number;
};

export type ScrollbarDragState = {
  // Смещение точки захвата от верхнего края ползунка в пикселях.
  grabOffsetPx: number;
};

export type ScrollbarThumbInput = ScrollbarTrackRect & {
  // Верх ползунка в процентах высоты полосы.
  thumbTopPercent: number;
  // Высота ползунка в процентах высоты полосы.
  thumbHeightPercent: number;
};

export type ScrollbarOffsetInput = {
  // Верх ползунка в процентах высоты полосы.
  thumbTopPercent: number;
  // Высота ползунка в процентах высоты полосы.
  thumbHeightPercent: number;
  // Максимальное значение прокрутки в пикселях контента.
  maxOffsetPx: number;
  // Признак обратного направления, когда нижняя позиция означает нулевую прокрутку.
  reverse: boolean;
};

// Создает состояние захвата так, чтобы ползунок продолжал идти под той же точкой курсора.
export const startScrollbarDrag = (input: ScrollbarThumbInput, cursorY: number): ScrollbarDragState => {
  const thumb = getScrollbarThumbRect(input);
  return {
    grabOffsetPx: clamp(cursorY - thumb.top, 0, thumb.height),
  };
};

// Вычисляет новый верх ползунка по вертикальной координате игрового указателя.
export const getScrollbarThumbTopPercentFromCursor = (
  input: ScrollbarTrackRect & { thumbHeightPercent: number; drag: ScrollbarDragState },
  cursorY: number,
): number => {
  const thumbHeightPx = input.height * input.thumbHeightPercent / 100;
  const availableTrackPx = Math.max(0, input.height - thumbHeightPx);
  if (availableTrackPx <= 0) {
    return 0;
  }

  const thumbTopPx = clamp(cursorY - input.drag.grabOffsetPx, input.top, input.top + availableTrackPx);
  return (thumbTopPx - input.top) / availableTrackPx * (100 - input.thumbHeightPercent);
};

// Переводит позицию ползунка в сдвиг контента.
export const getScrollOffsetFromThumbTopPercent = (input: ScrollbarOffsetInput): number => {
  const availablePercent = 100 - input.thumbHeightPercent;
  if (availablePercent <= 0 || input.maxOffsetPx <= 0) {
    return 0;
  }

  const ratio = clamp(input.thumbTopPercent / availablePercent, 0, 1);
  return (input.reverse ? 1 - ratio : ratio) * input.maxOffsetPx;
};

// Возвращает пиксельные границы ползунка внутри полосы.
export const getScrollbarThumbRect = (input: ScrollbarThumbInput): { top: number; height: number } => {
  const height = input.height * input.thumbHeightPercent / 100;
  return {
    top: input.top + input.height * input.thumbTopPercent / 100,
    height,
  };
};

const clamp = (value: number, min: number, max: number): number => Math.min(max, Math.max(min, value));
