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

// Преобразует серверный объект в компактную модель HUD без привязки к DOM.
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

// Рисует основные показатели посещаемого объекта отдельным DOM-слоем поверх canvas.
export class ObjectIndicatorsOverlay {
  constructor(
    // Корневой DOM-узел панели.
    private readonly element: HTMLElement,
  ) {}

  // Обновляет строки панели или скрывает её до появления объекта игрока.
  update(selfObject: CosmicObject | null): void {
    if (!selfObject) {
      this.element.replaceChildren();
      this.element.hidden = true;
      return;
    }

    this.element.hidden = false;
    this.element.replaceChildren(
      ...getObjectIndicators(selfObject).map((indicator) => this.renderIndicator(indicator)),
    );
  }

  // Создаёт одну строку со значком, полосой и числовым значением.
  private renderIndicator(indicator: ObjectIndicatorView): HTMLElement {
    const row = document.createElement("div");
    row.className = "object-indicator";
    row.title = indicator.title;

    const icon = document.createElement("div");
    icon.className = `object-indicator__icon object-indicator__icon--${indicator.acronym}`;
    icon.innerHTML = iconSvgByAcronym[indicator.acronym];

    const bar = document.createElement("div");
    bar.className = "object-indicator__bar";

    const fill = document.createElement("div");
    fill.className = "object-indicator__fill";
    fill.style.width = `${indicator.fillPercent}%`;

    const value = document.createElement("div");
    value.className = "object-indicator__value";
    value.textContent = indicator.valueText;

    bar.append(fill, value);
    row.append(icon, bar);

    return row;
  }
}

const iconSvgByAcronym: Record<ObjectIndicatorView["acronym"], string> = {
  Armor:
    '<svg viewBox="0 0 24 24" aria-hidden="true"><path d="M12 3l7 3v5c0 5-3 8-7 10-4-2-7-5-7-10V6l7-3z"/></svg>',
  Power:
    '<svg viewBox="0 0 24 24" aria-hidden="true"><path d="M13 2L5 13h6l-1 9 9-13h-6l1-7z"/></svg>',
  Fuel:
    '<svg viewBox="0 0 24 24" aria-hidden="true"><path d="M7 3h8v18H7z"/><path d="M9 6h4"/><path d="M15 8h2l2 3v7c0 1-1 2-2 2h-2"/><path d="M19 11h-2"/></svg>',
};
