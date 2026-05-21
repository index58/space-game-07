import { describe, expect, it } from "vitest";
import type { CosmicObject, EquipmentGroup, ReferenceDataMessage } from "../network/protocol";
import { applyActiveControlPanelUsageRelations, normalizeControlPanelUsageSelection } from "./controlPanelUsageSelection";

const equipmentGroup = (partial: Partial<EquipmentGroup>): EquipmentGroup => ({
  // Создаёт минимальную группу оборудования для проверки выбора в панели использования.
  ID: 1,
  CosmicObjectID: 100,
  Title: "",
  EquipmentItemModelID: 10,
  Count: 1,
  EnabledCount: 1,
  Enabled: true,
  Active: false,
  LastRechargeStartTime: 0,
  ...partial,
});

const cosmicObject = (partial: Partial<CosmicObject>): CosmicObject => ({
  // Создает минимальный объект мира для проверки состава доступного кластера.
  ID: 100,
  Title: "Ship",
  CosmicObjectModelID: 1,
  OwnerCharacterID: 7,
  OwnerNpcClanID: 0,
  CreatorCharacterID: 7,
  Mass: 1,
  Capacity: 0,
  MaxArmor: 100,
  MaxSpeed: 0,
  MaxAngularSpeed: 0,
  X: 0,
  Y: 0,
  Rotation: 0,
  Armor: 100,
  MaxAlongForce: 0,
  MaxAcrossForce: 0,
  MaxTorque: 0,
  GeneratingPower: 0,
  ConsumingPower: 0,
  AlongForce: 0,
  AcrossForce: 0,
  Torque: 0,
  Enabled: true,
  LastReceivedDamageTime: 0,
  Anchored: false,
  Complexity: 0,
  OccupiedVolume: 0,
  MaxFuel: 0,
  Fuel: 0,
  Speed: 0,
  VelocityX: 0,
  VelocityY: 0,
  AngularSpeed: 0,
  TargetRotation: 0,
  ...partial,
});

const referenceData = (): ReferenceDataMessage => ({
  // Создаёт справочники с контейнером, топливным баком и внешним предметом.
  type: "referenceData",
  NpcClan: { MaxID: 0, Items: {} },
  CosmicObjectType: { MaxID: 0, Items: {} },
  ItemType: {
    MaxID: 3,
    Items: {
      "1": { ID: 1, Acronym: "Container", IsPilotInstrument: false, IsInternalUsable: true },
      "2": { ID: 2, Acronym: "FuelTank", IsPilotInstrument: false, IsInternalUsable: true },
      "3": { ID: 3, Acronym: "Ore", IsPilotInstrument: false, IsInternalUsable: false },
    },
  },
  CosmicObjectModel: { MaxID: 0, Items: {} },
  ItemModel: {
    MaxID: 3,
    Items: {
      "10": { ID: 10, ItemTypeID: 1, Acronym: "SmallContainer" },
      "20": { ID: 20, ItemTypeID: 2, Acronym: "SmallFuelTank" },
      "30": { ID: 30, ItemTypeID: 3, Acronym: "Ore" },
    },
  },
  Blueprint: { MaxID: 0, Items: {} },
  BlueprintComponent: { MaxID: 0, Items: {} },
  Schema: { MaxID: 0, Items: {} },
  SchemaComponent: { MaxID: 0, Items: {} },
  ActionType: { MaxID: 0, Items: {} },
  InputEventType: { MaxID: 0, Items: {} },
  DefaultActionInputSetting: { MaxID: 0, Items: {} },
});

describe("normalizeControlPanelUsageSelection", () => {
  // Проверяет, что после перезагрузки пустой выбор становится теми же первыми группами, которые показывает UI.
  it("selects first visible container and internal equipment when selection is empty", () => {
    const selection = normalizeControlPanelUsageSelection({
      objectId: 100,
      referenceData: referenceData(),
      equipmentGroups: [
        equipmentGroup({ ID: 30, EquipmentItemModelID: 30 }),
        equipmentGroup({ ID: 10, EquipmentItemModelID: 10 }),
        equipmentGroup({ ID: 20, EquipmentItemModelID: 20 }),
      ],
      objects: [cosmicObject({ ID: 100 })],
      selection: { leftContainerObjectId: null, leftContainerGroupId: null, rightEquipmentObjectId: null, rightEquipmentGroupId: null, constructorMaterialObjectId: null, constructorMaterialGroupId: null, constructorProductObjectId: null, constructorProductGroupId: null },
    });

    expect(selection).toEqual({ leftContainerObjectId: 100, leftContainerGroupId: 10, rightEquipmentObjectId: 100, rightEquipmentGroupId: 10, constructorMaterialObjectId: 100, constructorMaterialGroupId: 10, constructorProductObjectId: 100, constructorProductGroupId: 10 });
  });

  // Проверяет, что ручной выбор сохраняется, пока выбранные группы остаются доступными.
  it("keeps selected groups while they still exist", () => {
    const selection = normalizeControlPanelUsageSelection({
      objectId: 100,
      referenceData: referenceData(),
      equipmentGroups: [
        equipmentGroup({ ID: 10, EquipmentItemModelID: 10 }),
        equipmentGroup({ ID: 20, EquipmentItemModelID: 20 }),
      ],
      objects: [cosmicObject({ ID: 100 })],
      selection: { leftContainerObjectId: 100, leftContainerGroupId: 10, rightEquipmentObjectId: 100, rightEquipmentGroupId: 20, constructorMaterialObjectId: 100, constructorMaterialGroupId: 10, constructorProductObjectId: 100, constructorProductGroupId: 10 },
    });

    expect(selection).toEqual({ leftContainerObjectId: 100, leftContainerGroupId: 10, rightEquipmentObjectId: 100, rightEquipmentGroupId: 20, constructorMaterialObjectId: 100, constructorMaterialGroupId: 10, constructorProductObjectId: 100, constructorProductGroupId: 10 });
  });

  // Проверяет, что для своего кластера доступны только собственные объекты, а выбор объекта берется из выбранной группы.
  it("keeps cluster object from selected group and excludes foreign cluster objects", () => {
    const selection = normalizeControlPanelUsageSelection({
      objectId: 100,
      referenceData: referenceData(),
      objects: [
        cosmicObject({ ID: 100, OwnerCharacterID: 7, ClusterMainCosmicObjectID: 100 }),
        cosmicObject({ ID: 101, OwnerCharacterID: 7, ClusterMainCosmicObjectID: 100 }),
        cosmicObject({ ID: 102, OwnerCharacterID: 8, ClusterMainCosmicObjectID: 100 }),
      ],
      equipmentGroups: [
        equipmentGroup({ ID: 10, CosmicObjectID: 100, EquipmentItemModelID: 10 }),
        equipmentGroup({ ID: 20, CosmicObjectID: 101, EquipmentItemModelID: 10 }),
        equipmentGroup({ ID: 30, CosmicObjectID: 102, EquipmentItemModelID: 10 }),
      ],
      selection: { leftContainerObjectId: null, leftContainerGroupId: 20, rightEquipmentObjectId: null, rightEquipmentGroupId: 20, constructorMaterialObjectId: null, constructorMaterialGroupId: 20, constructorProductObjectId: null, constructorProductGroupId: 20 },
    });

    expect(selection).toEqual({ leftContainerObjectId: 101, leftContainerGroupId: 20, rightEquipmentObjectId: 101, rightEquipmentGroupId: 20, constructorMaterialObjectId: 101, constructorMaterialGroupId: 20, constructorProductObjectId: 101, constructorProductGroupId: 20 });
  });

  // Проверяет, что при недоступной выбранной группе используется первый собственный объект кластера.
  it("falls back to first owned cluster object when selected group is unavailable", () => {
    const selection = normalizeControlPanelUsageSelection({
      objectId: 100,
      referenceData: referenceData(),
      objects: [
        cosmicObject({ ID: 100, OwnerCharacterID: 7, ClusterMainCosmicObjectID: 100 }),
        cosmicObject({ ID: 101, OwnerCharacterID: 7, ClusterMainCosmicObjectID: 100 }),
      ],
      equipmentGroups: [
        equipmentGroup({ ID: 20, CosmicObjectID: 101, EquipmentItemModelID: 10 }),
      ],
      selection: { leftContainerObjectId: null, leftContainerGroupId: 999, rightEquipmentObjectId: null, rightEquipmentGroupId: 999, constructorMaterialObjectId: null, constructorMaterialGroupId: 999, constructorProductObjectId: null, constructorProductGroupId: 999 },
    });

    expect(selection).toEqual({ leftContainerObjectId: 100, leftContainerGroupId: null, rightEquipmentObjectId: 100, rightEquipmentGroupId: null, constructorMaterialObjectId: 100, constructorMaterialGroupId: null, constructorProductObjectId: 100, constructorProductGroupId: null });
  });
});

describe("applyActiveControlPanelUsageRelations", () => {
  // Проверяет, что сохранённые связи левой панели не применяются для бездействующего правого оборудования.
  it("keeps current left selections when right equipment is not active", () => {
    const selection = applyActiveControlPanelUsageRelations({
      selection: { leftContainerObjectId: 100, leftContainerGroupId: 10, rightEquipmentObjectId: 100, rightEquipmentGroupId: 20, constructorMaterialObjectId: 100, constructorMaterialGroupId: 11, constructorProductObjectId: 100, constructorProductGroupId: 12 },
      rightEquipmentActive: false,
      relatedOppositeGroupId: 30,
      relatedSourceGroupId: 31,
      relatedDestinationGroupId: 32,
      groupObjectId: (groupId) => groupId + 1000,
    });

    expect(selection).toEqual({ leftContainerObjectId: 100, leftContainerGroupId: 10, rightEquipmentObjectId: 100, rightEquipmentGroupId: 20, constructorMaterialObjectId: 100, constructorMaterialGroupId: 11, constructorProductObjectId: 100, constructorProductGroupId: 12 });
  });

  // Проверяет, что сохранённые связи левой панели применяются для работающего правого оборудования.
  it("restores related left selections when right equipment is active", () => {
    const selection = applyActiveControlPanelUsageRelations({
      selection: { leftContainerObjectId: 100, leftContainerGroupId: 10, rightEquipmentObjectId: 100, rightEquipmentGroupId: 20, constructorMaterialObjectId: 100, constructorMaterialGroupId: 11, constructorProductObjectId: 100, constructorProductGroupId: 12 },
      rightEquipmentActive: true,
      relatedOppositeGroupId: 30,
      relatedSourceGroupId: 31,
      relatedDestinationGroupId: 32,
      groupObjectId: (groupId) => groupId + 1000,
    });

    expect(selection).toEqual({ leftContainerObjectId: 1030, leftContainerGroupId: 30, rightEquipmentObjectId: 100, rightEquipmentGroupId: 20, constructorMaterialObjectId: 1031, constructorMaterialGroupId: 31, constructorProductObjectId: 1032, constructorProductGroupId: 32 });
  });

  // Проверяет, что контейнер продукции конструктора не смешивается с обычным левым контейнером.
  it("restores constructor product selection separately from regular left selection", () => {
    const selection = applyActiveControlPanelUsageRelations({
      selection: {
        leftContainerObjectId: 100,
        leftContainerGroupId: 10,
        rightEquipmentObjectId: 100,
        rightEquipmentGroupId: 20,
        constructorMaterialObjectId: 100,
        constructorMaterialGroupId: 11,
        constructorProductObjectId: 100,
        constructorProductGroupId: 12,
      },
      rightEquipmentActive: true,
      relatedOppositeGroupId: null,
      relatedSourceGroupId: null,
      relatedDestinationGroupId: 32,
      groupObjectId: (groupId) => groupId + 1000,
    });

    expect(selection).toEqual({
      leftContainerObjectId: 100,
      leftContainerGroupId: 10,
      rightEquipmentObjectId: 100,
      rightEquipmentGroupId: 20,
      constructorMaterialObjectId: 100,
      constructorMaterialGroupId: 11,
      constructorProductObjectId: 1032,
      constructorProductGroupId: 32,
    });
  });

  // Проверяет, что сохранённая связь без доступной группы не меняет левую панель.
  it("ignores related groups that are no longer available", () => {
    const selection = applyActiveControlPanelUsageRelations({
      selection: { leftContainerObjectId: 100, leftContainerGroupId: 10, rightEquipmentObjectId: 100, rightEquipmentGroupId: 20, constructorMaterialObjectId: 100, constructorMaterialGroupId: 11, constructorProductObjectId: 100, constructorProductGroupId: 12 },
      rightEquipmentActive: true,
      relatedOppositeGroupId: 30,
      relatedSourceGroupId: 31,
      relatedDestinationGroupId: null,
      groupObjectId: () => null,
    });

    expect(selection).toEqual({ leftContainerObjectId: 100, leftContainerGroupId: 10, rightEquipmentObjectId: 100, rightEquipmentGroupId: 20, constructorMaterialObjectId: 100, constructorMaterialGroupId: 11, constructorProductObjectId: 100, constructorProductGroupId: 12 });
  });
});
