import type { CosmicObject, ReferenceDataMessage } from "../network/protocol";

const MINIMAP_WORLD_RADIUS_METERS = 800;
const FIRST_PVE_RADIUS_METERS = 100000;
const ZONE_RING_WIDTH_METERS = 10000;

export type MinimapPointKind = "npc" | "player" | "asteroid" | "owned" | "neutral";

export type MinimapPointView = {
  // Уникальный идентификатор объекта, отмеченного на карте.
  id: number;
  // Горизонтальная позиция точки внутри карты в процентах.
  xPercent: number;
  // Вертикальная позиция точки внутри карты в процентах.
  yPercent: number;
  // Визуальный тип точки, зависящий от типа и владельца объекта.
  kind: MinimapPointKind;
  // Признак посещаемого объекта игрока.
  isSelf: boolean;
};

export type MinimapCompassMarkView = {
  // Текст деления компаса.
  label: "N" | "E" | "S" | "W";
  // Горизонтальная позиция деления внутри линейной шкалы.
  xPercent: number;
};

export type MinimapView = {
  // Точки объектов, которые должны быть показаны в основной области карты.
  points: MinimapPointView[];
  // Признак ПВЕ-зоны в текущей позиции посещаемого объекта.
  isPveZone: boolean;
  // Признак активного якоря посещаемого объекта.
  isAnchored: boolean;
  // Деления горизонтальной шкалы компаса.
  compassMarks: MinimapCompassMarkView[];
};

export type MinimapInput = {
  // Посещаемый объект игрока, относительно которого центрируется карта.
  selfObject: CosmicObject;
  // Объекты из последнего серверного снимка, доступные клиенту.
  objects: CosmicObject[];
  // Справочники клиента, нужные для определения типов объектов.
  referenceData: ReferenceDataMessage;
};

// Готовит данные мини-карты для SolidJS-компонента без привязки к DOM.
export const getMinimapView = (input: MinimapInput): MinimapView => ({
  points: input.objects
    .map((object) => getMinimapPoint(input.selfObject, object, input.referenceData))
    .filter((point): point is MinimapPointView => point !== null),
  isPveZone: isPveZone(input.selfObject),
  isAnchored: input.selfObject.Anchored,
  compassMarks: getCompassMarks(input.selfObject.Rotation),
});

// Преобразует объект мира в точку мини-карты.
const getMinimapPoint = (
  selfObject: CosmicObject,
  object: CosmicObject,
  referenceData: ReferenceDataMessage,
): MinimapPointView | null => {
  const typeAcronym = getObjectTypeAcronym(object, referenceData);
  const isSelf = object.ID === selfObject.ID;
  const kind = getPointKind(selfObject, object, typeAcronym, isSelf);
  if (!kind) {
    return null;
  }

  return {
    id: object.ID,
    ...getRotatedPointPosition(selfObject, object),
    kind,
    isSelf,
  };
};

// Переводит мировое смещение объекта в локальные оси посещаемого объекта.
const getRotatedPointPosition = (
  selfObject: CosmicObject,
  object: CosmicObject,
): Pick<MinimapPointView, "xPercent" | "yPercent"> => {
  const dx = object.X - selfObject.X;
  const dy = object.Y - selfObject.Y;
  const cos = Math.cos(selfObject.Rotation);
  const sin = Math.sin(selfObject.Rotation);
  const localRight = dx * cos - dy * sin;
  const localForward = dx * sin + dy * cos;

  return {
    xPercent: clampPercent(50 + (localRight / MINIMAP_WORLD_RADIUS_METERS) * 50),
    yPercent: clampPercent(50 - (localForward / MINIMAP_WORLD_RADIUS_METERS) * 50),
  };
};

// Определяет визуальный тип точки по типу объекта и владельцам.
const getPointKind = (
  selfObject: CosmicObject,
  object: CosmicObject,
  typeAcronym: string | null,
  isSelf: boolean,
): MinimapPointKind | null => {
  if (isSelf) {
    return "owned";
  }
  if (typeAcronym === "Asteroid") {
    return "asteroid";
  }
  if (typeAcronym === "Loot") {
    return object.OwnerCharacterID > 0 && object.OwnerCharacterID === selfObject.OwnerCharacterID ? "owned" : null;
  }
  if (typeAcronym !== "Ship" && typeAcronym !== "Station") {
    return null;
  }
  if (object.OwnerNpcClanID > 0) {
    return "npc";
  }
  if (object.OwnerCharacterID > 0) {
    return object.OwnerCharacterID === selfObject.OwnerCharacterID ? "owned" : "player";
  }
  return "neutral";
};

// Возвращает акроним типа объекта через модель объекта.
const getObjectTypeAcronym = (object: CosmicObject, referenceData: ReferenceDataMessage): string | null => {
  const model = referenceData.CosmicObjectModel.Items[String(object.CosmicObjectModelID)];
  const typeID = model?.CosmicObjectTypeID;
  if (typeof typeID !== "number") {
    return null;
  }
  const type = referenceData.CosmicObjectType.Items[String(typeID)];
  return typeof type?.Acronym === "string" ? type.Acronym : null;
};

// Определяет тип зоны по расстоянию от центра мира.
const isPveZone = (object: CosmicObject): boolean => {
  const distanceFromCenter = Math.hypot(object.X, object.Y);
  if (distanceFromCenter <= FIRST_PVE_RADIUS_METERS) {
    return true;
  }
  const ringIndex = Math.floor((distanceFromCenter - FIRST_PVE_RADIUS_METERS) / ZONE_RING_WIDTH_METERS);
  return ringIndex % 2 === 1;
};

// Ограничивает позицию точки границами мини-карты.
const clampPercent = (value: number): number => Math.min(100, Math.max(0, value));

// Готовит видимые деления компаса с текущим направлением в центре шкалы.
const getCompassMarks = (rotation: number): MinimapCompassMarkView[] => {
  const headingDegrees = normalizeDegrees((rotation * 180) / Math.PI);
  return [
    { label: "N", xPercent: compassMarkPosition(0, headingDegrees) },
    { label: "E", xPercent: compassMarkPosition(90, headingDegrees) },
    { label: "S", xPercent: compassMarkPosition(180, headingDegrees) },
    { label: "W", xPercent: compassMarkPosition(270, headingDegrees) },
  ];
};

// Рассчитывает положение одного деления относительно текущего направления.
const compassMarkPosition = (bearingDegrees: number, headingDegrees: number): number =>
  clampPercent(50 + (signedDegreesDelta(bearingDegrees, headingDegrees) / 360) * 100);

// Нормализует угол в диапазон одного полного оборота.
const normalizeDegrees = (degrees: number): number => ((degrees % 360) + 360) % 360;

// Возвращает кратчайшую разницу направлений в градусах.
const signedDegreesDelta = (bearingDegrees: number, headingDegrees: number): number => {
  const delta = normalizeDegrees(bearingDegrees - headingDegrees);
  return delta > 180 ? delta - 360 : delta;
};
