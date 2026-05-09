import { describe, expect, it } from "vitest";
import { syncGameUiViewportScale } from "./gameViewportScale";

const setWindowHeight = (height: number): void => {
  Object.defineProperty(window, "innerHeight", { value: height, configurable: true });
};

const createGameRoot = (width: number, height: number): HTMLElement => {
  const element = document.createElement("div");
  element.getBoundingClientRect = () => ({
    x: 0,
    y: 0,
    left: 0,
    top: 0,
    right: width,
    bottom: height,
    width,
    height,
    toJSON: () => ({}),
  } as DOMRect);
  return element;
};

describe("game viewport ui scale", () => {
  // Проверяет, что в узком окне интерфейс сжимается от высоты игры, а не от высоты браузера.
  it("scales ui layout from letterboxed game height", () => {
    const gameRoot = createGameRoot(900, 675);
    const uiRoot = document.createElement("div");
    setWindowHeight(900);

    const result = syncGameUiViewportScale(gameRoot, uiRoot);

    expect(result).toEqual({
      gameWidthPx: 900,
      gameHeightPx: 675,
      scale: 0.75,
      layoutWidthPx: 1200,
      layoutHeightPx: 900,
      layoutViewportWidthUnitPx: 12,
    });
    expect(uiRoot.style.getPropertyValue("--game-ui-scale")).toBe("0.75");
    expect(uiRoot.style.getPropertyValue("--game-ui-layout-width")).toBe("1200px");
    expect(uiRoot.style.getPropertyValue("--game-ui-layout-height")).toBe("900px");
    expect(uiRoot.style.getPropertyValue("--game-ui-vw")).toBe("12px");
  });

  // Проверяет, что в обычном широком окне интерфейс не увеличивается сверх исходного масштаба.
  it("keeps ui scale at normal size when game uses browser height", () => {
    const gameRoot = createGameRoot(1365, 768);
    const uiRoot = document.createElement("div");
    setWindowHeight(768);

    const result = syncGameUiViewportScale(gameRoot, uiRoot);

    expect(result.scale).toBe(1);
    expect(result.layoutWidthPx).toBe(1365);
    expect(result.layoutHeightPx).toBe(768);
    expect(result.layoutViewportWidthUnitPx).toBe(13.65);
    expect(uiRoot.style.getPropertyValue("--game-ui-scale")).toBe("1");
    expect(uiRoot.style.getPropertyValue("--game-ui-vw")).toBe("13.65px");
  });
});
