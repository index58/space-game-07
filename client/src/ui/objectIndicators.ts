import { formatNumber } from "../domain/format";
import type { CosmicObject } from "../network/protocol";

export type ObjectIndicatorView = {
  // Стабильный строковый идентификатор показателя.
  acronym: "Armor" | "Power" | "Fuel";
  // Текст всплывающей подсказки.
  title: string;
  // Значение, отображаемое внутри полосы.
  valueText: string;
  // Заполнение полосы в процентах.
  fillPercent: number;
};

const clampPercent = (current: number, maximum: number): number => {
  if (!Number.isFinite(current) || !Number.isFinite(maximum) || maximum <= 0) {
    return 0;
  }
  return Math.min(100, Math.max(0, (current / maximum) * 100));
};

const formatIndicatorValue = (current: number, maximum: number): string =>
  `${formatNumber(current, 0)} / ${formatNumber(maximum, 0)}`;

// Преобразует серверный объект в компактную модель HUD без привязки к компонентам.
export const getObjectIndicators = (object: CosmicObject): ObjectIndicatorView[] => [
  {
    acronym: "Armor",
    title: "Броня",
    valueText: formatIndicatorValue(object.Armor, object.MaxArmor),
    fillPercent: clampPercent(object.Armor, object.MaxArmor),
  },
  {
    acronym: "Power",
    title: "Энергия",
    valueText: formatIndicatorValue(object.ConsumingPower, object.GeneratingPower),
    fillPercent: clampPercent(object.ConsumingPower, object.GeneratingPower),
  },
  {
    acronym: "Fuel",
    title: "Топливо",
    valueText: formatIndicatorValue(object.Fuel, object.MaxFuel),
    fillPercent: clampPercent(object.Fuel, object.MaxFuel),
  },
];
