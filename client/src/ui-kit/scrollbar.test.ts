import { describe, expect, it } from "vitest";
import { getScrollbarThumbRect, getScrollbarThumbTopPercentFromCursor, getScrollOffsetFromThumbTopPercent, startScrollbarDrag } from "./scrollbar";

describe("scrollbar helpers", () => {
  // Проверяет, что точка захвата сохраняется внутри ползунка и не заставляет его прыгать под курсор.
  it("keeps the captured cursor offset inside the thumb", () => {
    const drag = startScrollbarDrag({ top: 10, height: 100, thumbTopPercent: 20, thumbHeightPercent: 40 }, 38);

    expect(drag).toEqual({ grabOffsetPx: 8 });
  });

  // Проверяет, что вертикальное движение указателя дает такое же вертикальное движение ползунка.
  it("maps cursor movement to matching thumb movement", () => {
    const drag = startScrollbarDrag({ top: 0, height: 100, thumbTopPercent: 20, thumbHeightPercent: 45 }, 36);
    const nextTop = getScrollbarThumbTopPercentFromCursor({ top: 0, height: 100, thumbHeightPercent: 45, drag }, 56);

    expect(nextTop).toBeCloseTo(40);
  });

  // Проверяет обратное направление истории чата, где нижний ползунок означает просмотр самых новых сообщений.
  it("converts reversed thumb position to chat content offset", () => {
    expect(getScrollOffsetFromThumbTopPercent({ thumbTopPercent: 0, thumbHeightPercent: 25, maxOffsetPx: 300, reverse: true })).toBe(300);
    expect(getScrollOffsetFromThumbTopPercent({ thumbTopPercent: 75, thumbHeightPercent: 25, maxOffsetPx: 300, reverse: true })).toBe(0);
  });

  // Проверяет пиксельные границы ползунка, которые используются и чатом, и витриной UI Kit.
  it("returns thumb pixel bounds from shared percentages", () => {
    expect(getScrollbarThumbRect({ top: 10, height: 200, thumbTopPercent: 25, thumbHeightPercent: 30 })).toEqual({ top: 60, height: 60 });
  });
});
