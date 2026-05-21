import type { CosmicObject, EquipmentGroup, ReferenceDataMessage } from "../network/protocol";

export type ControlPanelUsageSelection = {
  // Объект, выбранный для левого контейнера.
  leftContainerObjectId: number | null;
  // Контейнер, выбранный в левой панели использования.
  leftContainerGroupId: number | null;
  // Объект, выбранный для правого оборудования.
  rightEquipmentObjectId: number | null;
  // Оборудование, выбранное в правой панели использования.
  rightEquipmentGroupId: number | null;
  // Объект, выбранный для контейнера материалов конструктора.
  constructorMaterialObjectId: number | null;
  // Контейнер материалов конструктора.
  constructorMaterialGroupId: number | null;
  // Объект, выбранный для контейнера продукции конструктора.
  constructorProductObjectId: number | null;
  // Контейнер продукции конструктора.
  constructorProductGroupId: number | null;
};

export type NormalizeControlPanelUsageSelectionInput = {
  // Объект, из которого открыта панель управления.
  objectId: number | null;
  // Объекты мира из текущего снимка.
  objects?: CosmicObject[];
  // Группы оборудования из текущего снимка с учётом ожидающих изменений.
  equipmentGroups: EquipmentGroup[];
  // Справочники, по которым определяется тип оборудования.
  referenceData: ReferenceDataMessage | null;
  // Текущий выбор панелей использования.
  selection: ControlPanelUsageSelection;
};

export type ApplyActiveControlPanelUsageRelationsInput = {
  // Выбор, полученный после обычной нормализации доступных списков.
  selection: ControlPanelUsageSelection;
  // Показывает, выполняет ли выбранная правая группа свою работу.
  rightEquipmentActive: boolean;
  // Сохранённый контейнер противоположной панели.
  relatedOppositeGroupId: number | null;
  // Сохранённый контейнер материалов.
  relatedSourceGroupId: number | null;
  // Сохранённый контейнер продукции.
  relatedDestinationGroupId: number | null;
  // Возвращает объект, которому принадлежит группа.
  groupObjectId: (groupId: number) => number | null;
};

// Возвращает текущий выбор или первый доступный элемент того списка, который показывает UI.
export const normalizeControlPanelUsageSelection = (input: NormalizeControlPanelUsageSelectionInput): ControlPanelUsageSelection => {
  if (!input.objectId || !input.referenceData) {
    return emptySelection();
  }

  const referenceData = input.referenceData;
  const availableObjectIDs = getAvailableClusterObjectIDs(input.objectId, input.objects ?? []);
  const groups = input.equipmentGroups
    .filter((group) => availableObjectIDs.includes(group.CosmicObjectID))
    .sort((left, right) => left.ID - right.ID);
  const containers = groups.filter((group) => isEquipmentGroupItemType(group, referenceData, "Container"));
  const internalEquipment = groups.filter((group) => isEquipmentGroupInternalUsable(group, referenceData));
  const leftContainer = normalizeSelectedGroupOnObject(availableObjectIDs, containers, input.selection.leftContainerObjectId, input.selection.leftContainerGroupId);
  const rightEquipment = normalizeSelectedGroupOnObject(availableObjectIDs, internalEquipment, input.selection.rightEquipmentObjectId, input.selection.rightEquipmentGroupId);
  const constructorMaterial = normalizeSelectedGroupOnObject(availableObjectIDs, containers, input.selection.constructorMaterialObjectId, input.selection.constructorMaterialGroupId);
  const constructorProduct = normalizeSelectedGroupOnObject(availableObjectIDs, containers, input.selection.constructorProductObjectId, input.selection.constructorProductGroupId);

  return {
    leftContainerObjectId: leftContainer.objectId,
    leftContainerGroupId: leftContainer.groupId,
    rightEquipmentObjectId: rightEquipment.objectId,
    rightEquipmentGroupId: rightEquipment.groupId,
    constructorMaterialObjectId: constructorMaterial.objectId,
    constructorMaterialGroupId: constructorMaterial.groupId,
    constructorProductObjectId: constructorProduct.objectId,
    constructorProductGroupId: constructorProduct.groupId,
  };
};

// Применяет сохранённые связи левой панели только для работающего правого оборудования.
export const applyActiveControlPanelUsageRelations = (input: ApplyActiveControlPanelUsageRelationsInput): ControlPanelUsageSelection => {
  if (!input.rightEquipmentActive) {
    return input.selection;
  }

  let selection = input.selection;
  if (input.relatedOppositeGroupId !== null) {
    selection = applyLeftContainerRelation(selection, input.relatedOppositeGroupId, input.groupObjectId);
  }
  if (input.relatedSourceGroupId !== null) {
    selection = applyConstructorMaterialRelation(selection, input.relatedSourceGroupId, input.groupObjectId);
  }
  if (input.relatedDestinationGroupId !== null) {
    selection = applyConstructorProductRelation(selection, input.relatedDestinationGroupId, input.groupObjectId);
  }
  return selection;
};

// Подставляет сохранённый контейнер в левую нижнюю часть панели.
const applyLeftContainerRelation = (selection: ControlPanelUsageSelection, groupId: number, groupObjectId: (groupId: number) => number | null): ControlPanelUsageSelection => {
  const objectId = groupObjectId(groupId);
  if (objectId === null) {
    return selection;
  }
  return {
    ...selection,
    leftContainerObjectId: objectId,
    leftContainerGroupId: groupId,
  };
};

// Подставляет сохранённый контейнер материалов в левую верхнюю часть панели.
const applyConstructorMaterialRelation = (selection: ControlPanelUsageSelection, groupId: number, groupObjectId: (groupId: number) => number | null): ControlPanelUsageSelection => {
  const objectId = groupObjectId(groupId);
  if (objectId === null) {
    return selection;
  }
  return {
    ...selection,
    constructorMaterialObjectId: objectId,
    constructorMaterialGroupId: groupId,
  };
};

// Подставляет сохранённый контейнер продукции в отдельный выбор конструктора.
const applyConstructorProductRelation = (selection: ControlPanelUsageSelection, groupId: number, groupObjectId: (groupId: number) => number | null): ControlPanelUsageSelection => {
  const objectId = groupObjectId(groupId);
  if (objectId === null) {
    return selection;
  }
  return {
    ...selection,
    constructorProductObjectId: objectId,
    constructorProductGroupId: groupId,
  };
};

// Возвращает объекты, оборудованием которых можно управлять из посещаемого объекта.
export const getAvailableClusterObjectIDs = (objectId: number, objects: CosmicObject[]): number[] => {
  const current = objects.find((object) => object.ID === objectId);
  if (!current) {
    return [objectId];
  }
  const mainID = current.ClusterMainCosmicObjectID ?? 0;
  const main = mainID > 0 ? objects.find((object) => object.ID === mainID) : null;
  if (!main || current.OwnerCharacterID <= 0 || main.OwnerCharacterID !== current.OwnerCharacterID) {
    return [objectId];
  }
  const clusterObjectIDs = objects
    .filter((object) => object.ClusterMainCosmicObjectID === mainID && object.OwnerCharacterID === current.OwnerCharacterID)
    .sort((left, right) => left.ID - right.ID)
    .map((object) => object.ID);
  return clusterObjectIDs.length > 0 ? clusterObjectIDs : [objectId];
};

// Возвращает выбранную группу на выбранном объекте или ближайший доступный вариант.
const normalizeSelectedGroupOnObject = (objectIDs: number[], groups: EquipmentGroup[], selectedObjectId: number | null, selectedGroupId: number | null): { objectId: number | null; groupId: number | null } => {
  const selectedGroup = groups.find((group) => group.ID === selectedGroupId);
  const objectId = selectedGroup?.CosmicObjectID ?? (objectIDs.includes(selectedObjectId ?? 0) ? selectedObjectId : objectIDs[0] ?? null);
  const objectGroups = groups.filter((group) => group.CosmicObjectID === objectId);
  return {
    objectId,
    groupId: selectedGroup?.CosmicObjectID === objectId ? selectedGroup.ID : objectGroups[0]?.ID ?? null,
  };
};

// Возвращает пустой выбор для состояния без объекта или справочников.
const emptySelection = (): ControlPanelUsageSelection => ({
  leftContainerObjectId: null,
  leftContainerGroupId: null,
  rightEquipmentObjectId: null,
  rightEquipmentGroupId: null,
  constructorMaterialObjectId: null,
  constructorMaterialGroupId: null,
  constructorProductObjectId: null,
  constructorProductGroupId: null,
});

// Проверяет принадлежность группы оборудования к типу предмета по акрониму.
const isEquipmentGroupItemType = (group: EquipmentGroup, referenceData: ReferenceDataMessage, itemTypeAcronym: string): boolean => {
  const itemModel = referenceData.ItemModel.Items[String(group.EquipmentItemModelID)];
  const itemType = referenceData.ItemType.Items[String(itemModel?.ItemTypeID)];
  return itemType?.Acronym === itemTypeAcronym;
};

// Проверяет, что группу можно использовать из правой панели использования.
const isEquipmentGroupInternalUsable = (group: EquipmentGroup, referenceData: ReferenceDataMessage): boolean => {
  const itemModel = referenceData.ItemModel.Items[String(group.EquipmentItemModelID)];
  const itemType = referenceData.ItemType.Items[String(itemModel?.ItemTypeID)];
  return Boolean(itemType?.IsInternalUsable);
};
