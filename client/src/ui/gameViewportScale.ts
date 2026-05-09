export type GameUiViewportScale = {
  // Видимая ширина игровой области в пикселях браузера.
  gameWidthPx: number;
  // Видимая высота игровой области в пикселях браузера.
  gameHeightPx: number;
  // Коэффициент, который переводит размеры UI из высоты браузера в высоту игровой области.
  scale: number;
  // Ширина внутренней раскладки UI до применения визуального масштабирования.
  layoutWidthPx: number;
  // Высота внутренней раскладки UI до применения визуального масштабирования.
  layoutHeightPx: number;
  // Один процент внутренней ширины UI до применения визуального масштабирования.
  layoutViewportWidthUnitPx: number;
};

// Синхронизирует DOM UI с letterbox-областью так, чтобы vh-размеры выглядели рассчитанными от высоты игры.
export const syncGameUiViewportScale = (gameRoot: HTMLElement, uiRoot: HTMLElement): GameUiViewportScale => {
  const gameRect = gameRoot.getBoundingClientRect();
  const browserHeightPx = Math.max(1, window.innerHeight);
  const gameWidthPx = Math.max(1, gameRect.width);
  const gameHeightPx = Math.max(1, gameRect.height);
  const scale = Math.min(1, gameHeightPx / browserHeightPx);
  const layoutWidthPx = gameWidthPx / scale;
  const layoutHeightPx = gameHeightPx / scale;
  const layoutViewportWidthUnitPx = layoutWidthPx / 100;

  uiRoot.style.setProperty("--game-ui-scale", String(scale));
  uiRoot.style.setProperty("--game-ui-layout-width", `${layoutWidthPx}px`);
  uiRoot.style.setProperty("--game-ui-layout-height", `${layoutHeightPx}px`);
  uiRoot.style.setProperty("--game-ui-vw", `${layoutViewportWidthUnitPx}px`);

  return {
    gameWidthPx,
    gameHeightPx,
    scale,
    layoutWidthPx,
    layoutHeightPx,
    layoutViewportWidthUnitPx,
  };
};
