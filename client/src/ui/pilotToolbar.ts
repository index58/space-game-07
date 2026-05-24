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
  // Количество боеприпасов, уже готовых к ближайшим выстрелам.
  magazineCount: number;
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
  // Признак подготовки новой порции зарядов.
  isReloading: boolean;
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
  // Выбранный индекс среди десяти ячеек панели.
  selectedToolIndex: number;
  // Текущее время кадра в миллисекундах Unix.
  nowMs: number;
  // Локальные моменты первого отображения подготовки зарядов.
  reloadDisplayStartMsByGroupId?: Record<number, number>;
};

type AggregatedPilotTool = {
  // Модель предмета, общая для объединенных групп оборудования.
  itemModel: ItemModelReference;
  // Сумма включенных единиц оборудования этой модели.
  enabledCount: number;
  // Суммарная вместимость магазинов всех групп этой модели.
  magazineCapacity: number;
  // Суммарное количество снарядов, уже заряженных в эти магазины.
  magazineCount: number;
  // Плавное значение для отображения подготовки зарядов.
  displayedMagazineCount: number;
  // Признак подготовки хотя бы одной группы этой модели.
  isReloading: boolean;
  // Минимальный ID группы, задающий временный порядок панели.
  firstGroupId: number;
};

// Собирает данные панели пилота из групп оборудования текущего объекта.
export const getPilotToolbarView = (input: PilotToolbarInput): PilotToolbarView => {
  const tools = getAggregatedPilotTools(input).slice(0, PILOT_TOOL_SLOT_COUNT);
  const selectedToolIndex = normalizePilotToolIndex(input.selectedToolIndex);
  const selectedTool = tools[selectedToolIndex] ?? null;
  const slots = Array.from({ length: PILOT_TOOL_SLOT_COUNT }, (_, slotIndex): PilotToolSlotView => {
    const tool = tools[slotIndex] ? toPilotToolView(tools[slotIndex]) : null;

    return {
      index: slotIndex + 1,
      tool,
      isSelected: slotIndex === selectedToolIndex,
    };
  });

  return {
    slots,
    magazine: selectedTool && selectedTool.magazineCapacity > 0
      ? {
        fillPercent: Math.max(0, Math.min(100, selectedTool.displayedMagazineCount / selectedTool.magazineCapacity * 100)),
        valueText: selectedTool.isReloading ? "Перезарядка" : `${selectedTool.magazineCount} / ${selectedTool.magazineCapacity}`,
        isReloading: selectedTool.isReloading,
      }
      : null,
  };
};

// Возвращает следующий индекс ячейки с кольцевым переходом через границы.
export const getNextPilotToolIndex = (currentIndex: number, delta: number): number => {
  return normalizePilotToolIndex(currentIndex + delta);
};

// Приводит индекс к диапазону десяти ячеек панели.
export const normalizePilotToolIndex = (index: number): number => {
  return ((index % PILOT_TOOL_SLOT_COUNT) + PILOT_TOOL_SLOT_COUNT) % PILOT_TOOL_SLOT_COUNT;
};

// Объединяет группы одной модели и оставляет только типы, разрешенные для панели пилота.
const getAggregatedPilotTools = (input: PilotToolbarInput): AggregatedPilotTool[] => {
  const byModelId = new Map<number, AggregatedPilotTool>();

  for (const group of input.equipmentGroups) {
    if (group.CosmicObjectID !== input.selfObject.ID || group.Count <= 0) {
      continue;
    }

    const itemModel = input.referenceData.ItemModel.Items[String(group.EquipmentItemModelID)];
    if (!itemModel || !isPilotInstrument(itemModel, input.referenceData)) {
      continue;
    }

    const existing = byModelId.get(itemModel.ID);
    if (existing) {
      existing.enabledCount += group.EnabledCount;
      existing.magazineCapacity += getEquipmentGroupMagazineCapacity(group, itemModel);
      existing.magazineCount += getEquipmentGroupMagazineCount(group, itemModel);
      existing.displayedMagazineCount += getEquipmentGroupDisplayedMagazineCount(group, itemModel, input.nowMs, input.reloadDisplayStartMsByGroupId);
      existing.isReloading = existing.isReloading || isEquipmentGroupReloading(group, itemModel);
      existing.firstGroupId = Math.min(existing.firstGroupId, group.ID);
      continue;
    }

    byModelId.set(itemModel.ID, {
      itemModel,
      enabledCount: group.EnabledCount,
      magazineCapacity: getEquipmentGroupMagazineCapacity(group, itemModel),
      magazineCount: getEquipmentGroupMagazineCount(group, itemModel),
      displayedMagazineCount: getEquipmentGroupDisplayedMagazineCount(group, itemModel, input.nowMs, input.reloadDisplayStartMsByGroupId),
      isReloading: isEquipmentGroupReloading(group, itemModel),
      firstGroupId: group.ID,
    });
  }

  return Array.from(byModelId.values()).sort((left, right) => left.firstGroupId - right.firstGroupId);
};

// Проверяет флаг типа предмета через справочник.
const isPilotInstrument = (itemModel: ItemModelReference, referenceData: ReferenceDataMessage): boolean => {
  return Boolean(referenceData.ItemType.Items[String(itemModel.ItemTypeID)]?.IsPilotInstrument);
};

// Приводит объединенную модель к данным для отрисовки ячейки.
const toPilotToolView = (tool: AggregatedPilotTool): PilotToolView => ({
  itemModelId: tool.itemModel.ID,
  acronym: tool.itemModel.Acronym,
  title: tool.itemModel.TitleRu || tool.itemModel.TitleEn || tool.itemModel.Acronym,
  iconFilePath: tool.itemModel.IconFilePath || null,
  enabledCount: tool.enabledCount,
  magazineCapacity: tool.magazineCapacity,
  magazineCount: tool.magazineCount,
});

// Считает суммарный размер магазина с учётом количества установленных единиц.
const getEquipmentGroupMagazineCapacity = (group: EquipmentGroup, itemModel: ItemModelReference): number => {
  const singleCapacity = itemModel.MagazineCapacity ?? 0;
  return Math.max(0, group.Count) * Math.max(0, singleCapacity);
};

// Берёт серверное значение магазина, а для старых снимков считает магазин полным.
const getEquipmentGroupMagazineCount = (group: EquipmentGroup, itemModel: ItemModelReference): number => {
  const capacity = getEquipmentGroupMagazineCapacity(group, itemModel);
  if (capacity <= 0) {
    return 0;
  }
  return Math.max(0, Math.min(capacity, group.MagazineCount ?? capacity));
};

// Возвращает значение для шкалы с учётом незавершённой подготовки зарядов.
const getEquipmentGroupDisplayedMagazineCount = (group: EquipmentGroup, itemModel: ItemModelReference, nowMs: number, reloadDisplayStartMsByGroupId?: Record<number, number>): number => {
  const capacity = getEquipmentGroupMagazineCapacity(group, itemModel);
  const magazineCount = getEquipmentGroupMagazineCount(group, itemModel);
  if (capacity <= 0 || !isEquipmentGroupReloading(group, itemModel)) {
    return magazineCount;
  }

  const rechargeDurationMs = Math.max(0, (itemModel.RechargeTime ?? 0) * 1000);
  if (rechargeDurationMs <= 0) {
    return capacity;
  }

  const displayStartMs = reloadDisplayStartMsByGroupId?.[group.ID] ?? group.LastRechargeStartTime;
  const progress = Math.max(0, Math.min(1, (nowMs - displayStartMs) / rechargeDurationMs));
  return Math.max(0, Math.min(capacity, magazineCount + (capacity - magazineCount) * progress));
};

// Проверяет, что сервер уже начал подготовку новой порции зарядов.
const isEquipmentGroupReloading = (group: EquipmentGroup, itemModel: ItemModelReference): boolean => {
  return group.LastRechargeStartTime > 0 && (itemModel.RechargeTime ?? 0) > 0;
};
