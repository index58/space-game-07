import type { PilotCamera } from "../domain/camera";
import { worldToPilotScreen } from "../domain/camera";
import type { WorldVector } from "../domain/types";
import type { CosmicObject, CosmicObjectModelReference } from "../network/protocol";

const BODY_POLYGON_VERTEX_COUNT = 16;
const AXIS_EPSILON = 0.000000000001;

// Строит локальные точки тела по той же формуле, что и сервер.
export const buildBodyPolygon = (model: CosmicObjectModelReference): WorldVector[] => {
  const offset = bodyPolygonOffset(model);
  const radiusX = model.BodyWidth / 2;
  const radiusY = model.BodyLength / 2;
  const points: WorldVector[] = [];

  for (let index = 0; index < BODY_POLYGON_VERTEX_COUNT; index++) {
    const angle = 2 * Math.PI * index / BODY_POLYGON_VERTEX_COUNT;
    points.push({
      x: zeroSmallValue(offset.x + zeroSmallValue(radiusX * Math.sin(angle))),
      y: zeroSmallValue(offset.y + zeroSmallValue(radiusY * Math.cos(angle))),
    });
  }

  return points;
};

// Рассчитывает смещение центра тела относительно центра текстуры.
const bodyPolygonOffset = (model: CosmicObjectModelReference): WorldVector => {
  if (model.TextureScale <= 0 || model.TextureWidth <= 0 || model.TextureHeight <= 0) {
    return { x: 0, y: 0 };
  }

  return {
    x: (model.TextureBodyOriginX - model.TextureWidth / 2) / model.TextureScale,
    y: (model.TextureBodyOriginY - model.TextureHeight / 2) / model.TextureScale,
  };
};

// Переводит локальное тело объекта в экранные точки камеры пилота.
export const bodyPolygonToPilotScreen = (
  object: CosmicObject,
  model: CosmicObjectModelReference,
  camera: PilotCamera,
): WorldVector[] => {
  const forward = {
    x: Math.sin(object.Rotation),
    y: Math.cos(object.Rotation),
  };
  const right = {
    x: Math.cos(object.Rotation),
    y: -Math.sin(object.Rotation),
  };

  return buildBodyPolygon(model).map((point) => worldToPilotScreen({
    x: object.X + right.x * point.x + forward.x * point.y,
    y: object.Y + right.y * point.x + forward.y * point.y,
  }, camera));
};

// Убирает микроскопическую погрешность у вершин на осях эллипса.
const zeroSmallValue = (value: number): number => Math.abs(value) < AXIS_EPSILON ? 0 : value;
