import type { WorldVector } from "./types";

export type PilotCamera = {
  shipPosition: WorldVector;
  shipRotation: number;
  zoom: number;
  viewportWidth: number;
  viewportHeight: number;
};

export type PilotBackgroundTransform = {
  position: WorldVector;
  size: number;
  rotation: number;
  scale: number;
  tileScale: number;
  tilePositionX: number;
  tilePositionY: number;
};

export const MIN_ZOOM = -100;
export const MAX_ZOOM = 100;
export const INITIAL_ZOOM = 0;
export const BACKGROUND_TEXTURE_SCALE = 2;
export const BASE_VIEWPORT_HEIGHT_METERS = 1000;
export const ZOOM_STEP_FACTOR = 1.1;

export const clampZoom = (zoom: number): number => Math.min(MAX_ZOOM, Math.max(MIN_ZOOM, zoom));

export const getViewportZoomScale = (zoom: number, viewportHeight: number): number =>
  (viewportHeight / BASE_VIEWPORT_HEIGHT_METERS) * ZOOM_STEP_FACTOR ** zoom;

export const getPilotShipScreenPosition = (
  viewportWidth: number,
  viewportHeight: number,
): WorldVector => ({
  x: viewportWidth / 2,
  y: viewportHeight * 0.75,
});

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

export const rotationToPilotScreen = (objectRotation: number, shipRotation: number): number =>
  objectRotation - shipRotation;

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

  // TileSprite хранит тайлы в локальных координатах, поэтому компенсируем центрирование через половину размера.
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
