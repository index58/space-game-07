import { buildBodyPolygon } from "../game/bodyPolygon";
import type { CosmicObject, CosmicObjectModelReference, ReferenceDataMessage } from "../network/protocol";

const probeLengthMeters = 100;
const intersectionEpsilon = 0.000000001;

export type InformationPanelRow = {
  // Подпись строки в информационной панели.
  label: string;
  // Значение строки в информационной панели.
  value: string;
};

export type InformationPanelView = {
  // Объект, физическое тело которого первым касается луч обзора.
  object: CosmicObject;
  // Строки, которые нужно показать в HUD.
  rows: InformationPanelRow[];
};

export type InformationPanelInput = {
  // Посещаемый объект, от носа которого выпускается луч обзора.
  selfObject: CosmicObject;
  // Объекты последнего серверного снимка.
  objects: CosmicObject[];
  // Справочники клиента для названий моделей и типов.
  referenceData: ReferenceDataMessage;
};

type WorldPoint = {
  // Горизонтальная координата в метрах мира.
  x: number;
  // Вертикальная координата в метрах мира.
  y: number;
};

// Собирает данные панели для первого объекта перед носом корабля.
export const getInformationPanelView = (input: InformationPanelInput): InformationPanelView | null => {
  const target = getProbeTarget(input);
  if (!target) {
    return null;
  }

  const rows = getInformationRows(target, input.referenceData);
  return rows.length > 0 ? { object: target, rows } : null;
};

// Находит ближайшее тело, которое пересекает 100-метровый луч обзора.
const getProbeTarget = (input: InformationPanelInput): CosmicObject | null => {
  const selfModel = input.referenceData.CosmicObjectModel.Items[String(input.selfObject.CosmicObjectModelID)];
  if (!selfModel) {
    return null;
  }

  const start = getProbeStart(input.selfObject, selfModel);
  const forward = getForward(input.selfObject.Rotation);
  const end = {
    x: start.x + forward.x * probeLengthMeters,
    y: start.y + forward.y * probeLengthMeters,
  };

  let best: { object: CosmicObject; distance: number } | null = null;
  for (const object of input.objects) {
    if (object.ID === input.selfObject.ID) {
      continue;
    }
    const model = input.referenceData.CosmicObjectModel.Items[String(object.CosmicObjectModelID)];
    if (!model) {
      continue;
    }
    const distance = intersectSegmentWithBody(start, end, object, model);
    if (distance === null) {
      continue;
    }
    if (!best || distance < best.distance || (distance === best.distance && object.ID < best.object.ID)) {
      best = { object, distance };
    }
  }

  return best?.object ?? null;
};

// Возвращает самую переднюю точку физического тела в мировых координатах.
const getProbeStart = (object: CosmicObject, model: CosmicObjectModelReference): WorldPoint => {
  const frontPoint = buildBodyPolygon(model).reduce((best, point) => point.y > best.y ? point : best);
  return localToWorld(object, frontPoint);
};

// Ищет первую точку пересечения отрезка с полигоном тела.
const intersectSegmentWithBody = (
  start: WorldPoint,
  end: WorldPoint,
  object: CosmicObject,
  model: CosmicObjectModelReference,
): number | null => {
  const polygon = buildBodyPolygon(model).map((point) => localToWorld(object, point));
  if (pointInPolygon(start, polygon)) {
    return 0;
  }

  let bestDistance: number | null = null;
  for (let index = 0; index < polygon.length; index += 1) {
    const distance = intersectSegments(start, end, polygon[index], polygon[(index + 1) % polygon.length]);
    if (distance === null) {
      continue;
    }
    if (bestDistance === null || distance < bestDistance) {
      bestDistance = distance;
    }
  }
  return bestDistance;
};

// Переводит локальную точку тела в мировые координаты объекта.
const localToWorld = (object: CosmicObject, point: WorldPoint): WorldPoint => {
  const forward = getForward(object.Rotation);
  const right = getRight(object.Rotation);
  return {
    x: object.X + right.x * point.x + forward.x * point.y,
    y: object.Y + right.y * point.x + forward.y * point.y,
  };
};

// Возвращает направление взгляда объекта.
const getForward = (rotation: number): WorldPoint => ({
  x: Math.sin(rotation),
  y: Math.cos(rotation),
});

// Возвращает правую ось объекта.
const getRight = (rotation: number): WorldPoint => ({
  x: Math.cos(rotation),
  y: -Math.sin(rotation),
});

// Проверяет попадание точки внутрь выпуклого тела.
const pointInPolygon = (point: WorldPoint, polygon: WorldPoint[]): boolean => {
  let inside = false;
  for (let index = 0, previous = polygon.length - 1; index < polygon.length; previous = index, index += 1) {
    const currentPoint = polygon[index];
    const previousPoint = polygon[previous];
    if (((currentPoint.y > point.y) !== (previousPoint.y > point.y)) &&
      point.x < (previousPoint.x - currentPoint.x) * (point.y - currentPoint.y) / (previousPoint.y - currentPoint.y) + currentPoint.x) {
      inside = !inside;
    }
  }
  return inside;
};

// Возвращает расстояние от начала первого отрезка до пересечения двух отрезков.
const intersectSegments = (firstStart: WorldPoint, firstEnd: WorldPoint, secondStart: WorldPoint, secondEnd: WorldPoint): number | null => {
  const ray = subtract(firstEnd, firstStart);
  const edge = subtract(secondEnd, secondStart);
  const denominator = cross(ray, edge);
  if (Math.abs(denominator) <= intersectionEpsilon) {
    return null;
  }

  const offset = subtract(secondStart, firstStart);
  const rayRatio = cross(offset, edge) / denominator;
  const edgeRatio = cross(offset, ray) / denominator;
  if (rayRatio < -intersectionEpsilon || rayRatio > 1 + intersectionEpsilon || edgeRatio < -intersectionEpsilon || edgeRatio > 1 + intersectionEpsilon) {
    return null;
  }
  return Math.max(0, rayRatio) * Math.hypot(ray.x, ray.y);
};

// Вычитает две мировые точки как векторы.
const subtract = (left: WorldPoint, right: WorldPoint): WorldPoint => ({
  x: left.x - right.x,
  y: left.y - right.y,
});

// Возвращает псевдоскалярное произведение двух векторов.
const cross = (left: WorldPoint, right: WorldPoint): number => left.x * right.y - left.y * right.x;

// Собирает строки панели по доступным клиенту данным.
const getInformationRows = (object: CosmicObject, referenceData: ReferenceDataMessage): InformationPanelRow[] => {
  const rows: InformationPanelRow[] = [];
  if (object.Title.trim()) {
    rows.push({ label: "Название", value: object.Title });
  }

  const model = referenceData.CosmicObjectModel.Items[String(object.CosmicObjectModelID)];
  const modelTitle = model ? getTitle(model) : "";
  if (modelTitle) {
    rows.push({ label: "Модель", value: modelTitle });
  }

  const ownerName = getOptionalString(object, "OwnerName") || getOptionalString(object, "OwnerAccountNickname");
  if (ownerName) {
    rows.push({ label: "Владелец", value: ownerName });
  }

  if (object.OwnerNpcClanID > 0) {
    const npcClan = referenceData.NpcClan.Items[String(object.OwnerNpcClanID)];
    const npcClanTitle = npcClan ? getTitle(npcClan) : "";
    if (npcClanTitle) {
      rows.push({ label: "NPC-клан", value: npcClanTitle });
    }
  }

  return rows;
};

// Выбирает человекочитаемое название из справочной записи.
const getTitle = (record: Record<string, unknown>): string => (
  getOptionalString(record, "TitleRu") ||
  getOptionalString(record, "TitleEn") ||
  getOptionalString(record, "Title") ||
  getOptionalString(record, "Acronym")
);

// Безопасно достает строковое поле из неизвестной справочной записи.
const getOptionalString = (record: Record<string, unknown>, field: string): string => {
  const value = record[field];
  return typeof value === "string" ? value.trim() : "";
};
