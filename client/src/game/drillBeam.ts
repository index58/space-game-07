export const SIMPLE_DRILL_RAY_ACRONYM = "SimpleDrillRay";

const INTERSECTION_EPSILON = 0.000000001;

export type DrillBeamPoint = {
  // Горизонтальная координата в пикселях экрана.
  x: number;
  // Вертикальная координата в пикселях экрана.
  y: number;
};

export type DrillBeamGeometry = {
  // Точка выхода луча из носа корабля.
  start: DrillBeamPoint;
  // Точка контакта на предельной дальности инструмента.
  end: DrillBeamPoint;
  // Длина видимой части в пикселях экрана.
  lengthPx: number;
  // Базовая толщина яркого ядра в пикселях экрана.
  widthPx: number;
};

export type DrillBeamInput = {
  // Экранный центр объекта луча из серверного снимка.
  center: DrillBeamPoint;
  // Экранный поворот объекта луча относительно камеры пилота.
  rotation: number;
  // Длина луча в метрах мира.
  lengthMeters: number;
  // Текущий масштаб камеры в пикселях на метр.
  zoomScale: number;
};

// Считает экранные точки луча, не привязывая эффект к Phaser.
export const getDrillBeamGeometry = (input: DrillBeamInput): DrillBeamGeometry | null => {
  const lengthPx = input.lengthMeters * input.zoomScale;
  if (lengthPx <= 0) {
    return null;
  }

  const halfLength = lengthPx / 2;
  const forward = {
    x: Math.sin(input.rotation),
    y: -Math.cos(input.rotation),
  };

  return {
    start: {
      x: input.center.x - forward.x * halfLength,
      y: input.center.y - forward.y * halfLength,
    },
    end: {
      x: input.center.x + forward.x * halfLength,
      y: input.center.y + forward.y * halfLength,
    },
    lengthPx,
    widthPx: input.zoomScale * 3,
  };
};

// Обрезает видимый след по ближайшей границе объектов, которые стоят на пути.
export const clipDrillBeamGeometryToPolygons = (
  geometry: DrillBeamGeometry,
  polygons: DrillBeamPoint[][],
): DrillBeamGeometry => {
  const segmentLength = Math.hypot(geometry.end.x - geometry.start.x, geometry.end.y - geometry.start.y);
  if (segmentLength <= 0) {
    return geometry;
  }

  let nearestDistance = segmentLength;
  for (const polygon of polygons) {
    if (polygon.length < 2) {
      continue;
    }

    for (let index = 0; index < polygon.length; index += 1) {
      const edgeStart = polygon[index];
      const edgeEnd = polygon[(index + 1) % polygon.length];
      const distance = drillBeamIntersectionDistance(geometry.start, geometry.end, edgeStart, edgeEnd);
      if (distance !== null && distance < nearestDistance) {
        nearestDistance = distance;
      }
    }
  }

  if (Math.abs(nearestDistance - segmentLength) <= INTERSECTION_EPSILON) {
    return geometry;
  }

  const ratio = nearestDistance / segmentLength;
  return {
    ...geometry,
    end: {
      x: geometry.start.x + (geometry.end.x - geometry.start.x) * ratio,
      y: geometry.start.y + (geometry.end.y - geometry.start.y) * ratio,
    },
    lengthPx: nearestDistance,
  };
};

// Находит расстояние от начала луча до пересечения с одной стороной тела.
const drillBeamIntersectionDistance = (
  beamStart: DrillBeamPoint,
  beamEnd: DrillBeamPoint,
  edgeStart: DrillBeamPoint,
  edgeEnd: DrillBeamPoint,
): number | null => {
  const beam = subtractPoints(beamEnd, beamStart);
  const edge = subtractPoints(edgeEnd, edgeStart);
  const denominator = crossPoints(beam, edge);
  if (Math.abs(denominator) <= INTERSECTION_EPSILON) {
    return null;
  }

  const offset = subtractPoints(edgeStart, beamStart);
  const beamRatio = crossPoints(offset, edge) / denominator;
  const edgeRatio = crossPoints(offset, beam) / denominator;
  if (
    beamRatio < -INTERSECTION_EPSILON
    || beamRatio > 1 + INTERSECTION_EPSILON
    || edgeRatio < -INTERSECTION_EPSILON
    || edgeRatio > 1 + INTERSECTION_EPSILON
  ) {
    return null;
  }

  return Math.max(0, beamRatio) * Math.hypot(beam.x, beam.y);
};

// Вычитает экранные точки как обычные двумерные векторы.
const subtractPoints = (left: DrillBeamPoint, right: DrillBeamPoint): DrillBeamPoint => ({
  x: left.x - right.x,
  y: left.y - right.y,
});

// Возвращает ориентированную площадь пары векторов для проверки пересечения.
const crossPoints = (left: DrillBeamPoint, right: DrillBeamPoint): number =>
  left.x * right.y - left.y * right.x;

// Возвращает положение внутренних штрихов так, чтобы движение шло от цели к кораблю.
export const getDrillBeamIntakeProgress = (timeMs: number, index: number): number => {
  const progress = (timeMs * 0.00042 + index * 0.173) % 1;
  return 1 - progress;
};
