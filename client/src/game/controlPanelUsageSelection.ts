import type { EquipmentGroup, ReferenceDataMessage } from "../network/protocol";

export type ControlPanelUsageSelection = {
  // ID выбранного контейнера в левой панели использования.
  leftContainerGroupId: number | null;
  // ID выбранного оборудования в правой панели использования.
  rightEquipmentGroupId: number | null;
};

export type NormalizeControlPanelUsageSelectionInput = {
  // ID объекта, для которого открыта панель управления.
  objectId: number | null;
  // Группы оборудования из текущего снимка с учётом ожидающих изменений.
  equipmentGroups: EquipmentGroup[];
  // Справочники, по которым определяется тип оборудования.
  referenceData: ReferenceDataMessage | null;
  // Текущий выбор панелей использования.
  selection: ControlPanelUsageSelection;
};

// Возвращает текущий выбор или первый доступный элемент того же списка, который показывает UI.
export const normalizeControlPanelUsageSelection = (input: NormalizeControlPanelUsageSelectionInput): ControlPanelUsageSelection => {
  if (!input.objectId || !input.referenceData) {
    return { leftContainerGroupId: null, rightEquipmentGroupId: null };
  }

  const referenceData = input.referenceData;
  const groups = input.equipmentGroups
    .filter((group) => group.CosmicObjectID === input.objectId)
    .sort((left, right) => left.ID - right.ID);
  const leftContainers = groups.filter((group) => isEquipmentGroupItemtype(group, referenceData, "Container"));
  const rightEquipment = groups.filter((group) => isEquipmentGroupInternalUsable(group, referenceData));

  return {
    leftContainerGroupId: normalizeSelectedGroupId(leftContainers, input.selection.leftContainerGroupId),
    rightEquipmentGroupId: normalizeSelectedGroupId(rightEquipment, input.selection.rightEquipmentGroupId),
  };
};

// Возвращает текущий ID, если он ещё есть в списке, иначе выбирает первый доступный ID.
const normalizeSelectedGroupId = (groups: EquipmentGroup[], selectedGroupId: number | null): number | null =>
  groups.some((group) => group.ID === selectedGroupId) ? selectedGroupId : groups[0]?.ID ?? null;

// Проверяет принадлежность группы оборудования к типу предмета по акрониму.
const isEquipmentGroupItemtype = (group: EquipmentGroup, referenceData: ReferenceDataMessage, itemtypeAcronym: string): boolean => {
  const itemModel = referenceData.ItemModel.Items[String(group.EquipmentItemModelID)];
  const itemtype = referenceData.Itemtype.Items[String(itemModel?.ItemtypeID)];
  return itemtype?.Acronym === itemtypeAcronym;
};

// Проверяет, что группу можно использовать из правой панели использования.
const isEquipmentGroupInternalUsable = (group: EquipmentGroup, referenceData: ReferenceDataMessage): boolean => {
  const itemModel = referenceData.ItemModel.Items[String(group.EquipmentItemModelID)];
  const itemtype = referenceData.Itemtype.Items[String(itemModel?.ItemtypeID)];
  return Boolean(itemtype?.IsInternalUsable);
};
