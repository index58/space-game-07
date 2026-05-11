import type { CosmicObject, EquipmentGroup, ItemGroup, ReferenceDataMessage } from "../network/protocol";

export type ControlPanelFuelFillAmountInput = {
  // Текущий объект, запас топлива которого будет пополняться.
  object: CosmicObject | null;
  // Группа выбранного топливного бака.
  fuelTankGroup: EquipmentGroup | null;
  // Группы предметов в контейнерах из текущего снимка.
  itemGroups: ItemGroup[];
  // ID выбранных строк топлива в левом контейнере.
  selectedItemGroupIds: number[];
  // Справочники, по которым определяется модель топлива для бака.
  referenceData: ReferenceDataMessage | null;
};

// Возвращает максимальное количество топлива, которое можно залить из выбранных строк.
export const getControlPanelFuelFillMaxAmount = (input: ControlPanelFuelFillAmountInput): number => {
  if (!input.object || !input.fuelTankGroup || !input.referenceData) {
    return 0;
  }

  const freeFuel = Math.max(0, input.object.MaxFuel - input.object.Fuel);
  const fuelModelId = numericField(input.referenceData.ItemModel.Items[String(input.fuelTankGroup.EquipmentItemModelID)], "ConsumingItemModelID");
  if (freeFuel <= 0 || fuelModelId <= 0) {
    return 0;
  }

  const selectedIds = new Set(input.selectedItemGroupIds);
  const selectedFuel = input.itemGroups
    .filter((itemGroup) => selectedIds.has(itemGroup.ID) && itemGroup.ContentItemModelID === fuelModelId)
    .reduce((total, itemGroup) => total + itemGroup.Count, 0);

  return Math.min(freeFuel, Math.max(0, selectedFuel));
};

// Возвращает числовое поле справочника или ноль, если поле не задано.
const numericField = (record: Record<string, unknown> | undefined, key: string): number => {
  const value = record?.[key];
  return typeof value === "number" && Number.isFinite(value) ? value : 0;
};
