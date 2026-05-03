import type { CosmicObject, EquipmentGroup, ItemModelReference, ReferenceDataMessage } from "../network/protocol";

const PILOT_TOOL_SLOT_COUNT = 10;

export type PilotToolView = {
  // Модель предмета, назначенная в слот панели пилота.
  itemModelId: number;
  // Неизменяемый строковый идентификатор модели.
  acronym: string;
  // Отображаемое название инструмента.
  title: string;
  // Путь к файлу иконки, если он задан в справочнике.
  iconFilePath: string | null;
  // Сумма включенных единиц оборудования этой модели на объекте.
  enabledCount: number;
  // Вместимость магазина, если у инструмента есть магазин.
  magazineCapacity: number;
};

export type PilotToolSlotView = {
  // Номер ячейки панели от 1 до 10.
  index: number;
  // Инструмент, назначенный в ячейку, если он есть.
  tool: PilotToolView | null;
  // Признак выбранной ячейки.
  isSelected: boolean;
};

export type PilotToolMagazineView = {
  // Процент заполнения прогресс-бара магазина.
  fillPercent: number;
  // Текстовое значение заполненности магазина.
  valueText: string;
};

export type PilotToolbarView = {
  // Десять ячеек панели инструментов пилота.
  slots: PilotToolSlotView[];
  // Прогресс магазина выбранного инструмента, если у него есть магазин.
  magazine: PilotToolMagazineView | null;
};

type PilotToolbarInput = {
  // Посещаемый объект, для которого строится панель пилота.
  selfObject: CosmicObject;
  // Группы оборудования из последнего серверного снимка.
  equipmentGroups: EquipmentGroup[];
  // Справочники клиента для определения моделей и типов оборудования.
  referenceData: ReferenceDataMessage;
  // Выбранный индекс среди доступных инструментов.
  selectedToolIndex: number;
};

type AggregatedPilotTool = {
  // Модель предмета, общая для объединенных групп оборудования.
  itemModel: ItemModelReference;
  // Сумма включенных единиц оборудования этой модели.
  enabledCount: number;
  // Минимальный ID группы, задающий временный порядок панели.
  firstGroupId: number;
};

// Собирает данные панели пилота из групп оборудования текущего объекта.
export const getPilotToolbarView = (input: PilotToolbarInput): PilotToolbarView => {
  const tools = getAggregatedPilotTools(input).slice(0, PILOT_TOOL_SLOT_COUNT);
  const selectedToolIndex = normalizePilotToolIndex(input.selectedToolIndex, tools.length);
  const selectedTool = tools[selectedToolIndex] ?? null;
  const slots = Array.from({ length: PILOT_TOOL_SLOT_COUNT }, (_, slotIndex): PilotToolSlotView => {
    const tool = tools[slotIndex] ? toPilotToolView(tools[slotIndex]) : null;

    return {
      index: slotIndex + 1,
      tool,
      isSelected: slotIndex === selectedToolIndex && tool !== null,
    };
  });

  return {
    slots,
    magazine: selectedTool && selectedTool.itemModel.MagazineCapacity && selectedTool.itemModel.MagazineCapacity > 0
      ? {
        fillPercent: 100,
        valueText: `${selectedTool.itemModel.MagazineCapacity} / ${selectedTool.itemModel.MagazineCapacity}`,
      }
      : null,
  };
};

// Возвращает следующий индекс инструмента с кольцевым переходом через границы.
export const getNextPilotToolIndex = (currentIndex: number, delta: number, toolCount: number): number => {
  if (toolCount <= 0) {
    return 0;
  }
  return normalizePilotToolIndex(currentIndex + delta, toolCount);
};

// Приводит индекс к доступному диапазону инструментов.
export const normalizePilotToolIndex = (index: number, toolCount: number): number => {
  if (toolCount <= 0) {
    return 0;
  }
  return ((index % toolCount) + toolCount) % toolCount;
};

// Объединяет группы одной модели и оставляет только типы, разрешенные для панели пилота.
const getAggregatedPilotTools = (input: PilotToolbarInput): AggregatedPilotTool[] => {
  const byModelId = new Map<number, AggregatedPilotTool>();

  for (const group of input.equipmentGroups) {
    if (group.CosmicObjectID !== input.selfObject.ID || group.EnabledCount <= 0) {
      continue;
    }

    const itemModel = input.referenceData.ItemModel.Items[String(group.EquipmentItemModelID)];
    if (!itemModel || !isPilotInstrument(itemModel, input.referenceData)) {
      continue;
    }

    const existing = byModelId.get(itemModel.ID);
    if (existing) {
      existing.enabledCount += group.EnabledCount;
      existing.firstGroupId = Math.min(existing.firstGroupId, group.ID);
      continue;
    }

    byModelId.set(itemModel.ID, {
      itemModel,
      enabledCount: group.EnabledCount,
      firstGroupId: group.ID,
    });
  }

  return Array.from(byModelId.values()).sort((left, right) => left.firstGroupId - right.firstGroupId);
};

// Проверяет флаг типа предмета через справочник.
const isPilotInstrument = (itemModel: ItemModelReference, referenceData: ReferenceDataMessage): boolean => {
  return Boolean(referenceData.Itemtype.Items[String(itemModel.ItemtypeID)]?.IsPilotInstrument);
};

// Приводит объединенную модель к данным для отрисовки ячейки.
const toPilotToolView = (tool: AggregatedPilotTool): PilotToolView => ({
  itemModelId: tool.itemModel.ID,
  acronym: tool.itemModel.Acronym,
  title: tool.itemModel.TitleRu || tool.itemModel.TitleEn || tool.itemModel.Acronym,
  iconFilePath: tool.itemModel.IconFilePath || null,
  enabledCount: tool.enabledCount,
  magazineCapacity: tool.itemModel.MagazineCapacity ?? 0,
});
