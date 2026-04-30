import type { WorldVector } from "./types";

// Описывает камеру пилота: корабль игрока является центром ориентации экрана.
export type PilotCamera = {
  // Положение корабля игрока в мировых координатах.
  shipPosition: WorldVector;
  // Угол поворота корабля игрока в радианах.
  shipRotation: number;
  // Текущий масштаб перевода метров мира в пиксели экрана.
  zoom: number;
  // Ширина области просмотра в пикселях.
  viewportWidth: number;
  // Высота области просмотра в пикселях.
  viewportHeight: number;
};

// Содержит готовые параметры тайлового спрайта для бесшовного космического фона.
export type PilotBackgroundTransform = {
  // Экранный центр тайлового фона.
  position: WorldVector;
  // Размер квадратной области фона в метрах мира.
  size: number;
  // Поворот фона относительно камеры.
  rotation: number;
  // Масштаб отрисовки фона в пикселях на метр.
  scale: number;
  // Масштаб тайловой текстуры внутри спрайта.
  tileScale: number;
  // Горизонтальное смещение тайловой текстуры.
  tilePositionX: number;
  // Вертикальное смещение тайловой текстуры.
  tilePositionY: number;
};

// Ограничения зума защищают сцену от слишком мелких или слишком крупных значений.
export const MIN_ZOOM = -100;
export const MAX_ZOOM = 100;
export const INITIAL_ZOOM = 0;
// Переводит мировые метры в координаты тайла фоновой текстуры.
export const BACKGROUND_TEXTURE_SCALE = 2;
// Задает базовую высоту видимого мира при нулевом зуме.
export const BASE_VIEWPORT_HEIGHT_METERS = 1000;
// Определяет множитель одного шага колесика мыши.
export const ZOOM_STEP_FACTOR = 1.1;

// Удерживает пользовательский зум в допустимых пределах.
export const clampZoom = (zoom: number): number => Math.min(MAX_ZOOM, Math.max(MIN_ZOOM, zoom));

// Переводит уровень зума в пиксели на метр с учетом высоты окна.
export const getViewportZoomScale = (zoom: number, viewportHeight: number): number =>
  (viewportHeight / BASE_VIEWPORT_HEIGHT_METERS) * ZOOM_STEP_FACTOR ** zoom;

// Фиксирует корабль пилота ниже центра, оставляя больше обзора впереди.
export const getPilotShipScreenPosition = (
  viewportWidth: number,
  viewportHeight: number,
): WorldVector => ({
  x: viewportWidth / 2,
  y: viewportHeight * 0.75,
});

// Переводит мировые координаты объекта в экранные координаты камеры пилота.
export const worldToPilotScreen = (worldPosition: WorldVector, camera: PilotCamera): WorldVector => {
  const dx = worldPosition.x - camera.shipPosition.x;
  const dy = worldPosition.y - camera.shipPosition.y;
  const cos = Math.cos(camera.shipRotation);
  const sin = Math.sin(camera.shipRotation);

  // Переводим мировое смещение в локальные оси корабля: вправо и вперед.
  const localRight = dx * cos - dy * sin;
  const localForward = dx * sin + dy * cos;
  const shipScreen = getPilotShipScreenPosition(camera.viewportWidth, camera.viewportHeight);

  return {
    x: shipScreen.x + localRight * camera.zoom,
    y: shipScreen.y - localForward * camera.zoom,
  };
};

// Делает поворот объекта относительным к повороту корабля игрока.
export const rotationToPilotScreen = (objectRotation: number, shipRotation: number): number =>
  objectRotation - shipRotation;

// Рассчитывает фон так, чтобы он всегда закрывал весь повернутый viewport.
export const getPilotBackgroundTransform = (camera: PilotCamera): PilotBackgroundTransform => {
  const shipScreen = getPilotShipScreenPosition(camera.viewportWidth, camera.viewportHeight);
  const maxDistanceToViewportCorner = Math.max(
    Math.hypot(shipScreen.x, shipScreen.y),
    Math.hypot(camera.viewportWidth - shipScreen.x, shipScreen.y),
    Math.hypot(shipScreen.x, camera.viewportHeight - shipScreen.y),
    Math.hypot(camera.viewportWidth - shipScreen.x, camera.viewportHeight - shipScreen.y),
  );
  const displaySize = maxDistanceToViewportCorner * 2;
  const size = displaySize / camera.zoom;

  // Тайловый спрайт хранит тайлы в локальных координатах, поэтому компенсируем центрирование через половину размера.
  return {
    position: shipScreen,
    size,
    rotation: -camera.shipRotation,
    scale: camera.zoom,
    tileScale: 1 / BACKGROUND_TEXTURE_SCALE,
    tilePositionX: (camera.shipPosition.x - size / 2) * BACKGROUND_TEXTURE_SCALE,
    tilePositionY: (-camera.shipPosition.y - size / 2) * BACKGROUND_TEXTURE_SCALE,
  };
};
