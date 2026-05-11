import { describe, expect, it } from "vitest";
import type { CosmicObject, EquipmentGroup, ReferenceDataMessage } from "../network/protocol";
import { getControlPanelFuelFillMaxAmount } from "./controlPanelFuelAmount";

const object = (fuel: number, maxFuel: number): CosmicObject => ({
  // Создаёт минимальный управляемый объект с запасом топлива.
  ID: 1,
  Title: "",
  CosmicObjectModelID: 1,
  OwnerCharacterID: 0,
  OwnerNpcClanID: 0,
  CreatorCharacterID: 0,
  Mass: 1,
  Capacity: 1,
  MaxArmor: 1,
  MaxSpeed: 0,
  MaxAngularSpeed: 0,
  X: 0,
  Y: 0,
  Rotation: 0,
  Armor: 1,
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
  Complexity: 1,
  OccupiedVolume: 0,
  MaxFuel: maxFuel,
  Fuel: fuel,
  Speed: 0,
  VelocityX: 0,
  VelocityY: 0,
  AngularSpeed: 0,
  TargetRotation: 0,
});

const fuelTank = (): EquipmentGroup => ({
  // Создаёт группу топливного бака, модель которого потребляет выбранное топливо.
  ID: 20,
  CosmicObjectID: 1,
  Title: "Бак",
  EquipmentItemModelID: 200,
  Count: 1,
  EnabledCount: 1,
  Enabled: true,
  Active: false,
  LastRechargeStartTime: 0,
});

const referenceData = (): ReferenceDataMessage => ({
  // Создаёт справочник моделей с привязкой топливного бака к модели топлива.
  type: "referenceData",
  NpcClan: { MaxID: 0, Items: {} },
  CosmicObjectType: { MaxID: 0, Items: {} },
  Itemtype: { MaxID: 0, Items: {} },
  CosmicObjectModel: { MaxID: 0, Items: {} },
  ItemModel: {
    MaxID: 2,
    Items: {
      "100": { ID: 100, ItemtypeID: 1, Acronym: "Fuel" },
      "200": { ID: 200, ItemtypeID: 2, Acronym: "FuelTank", ConsumingItemModelID: 100 },
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

describe("getControlPanelFuelFillMaxAmount", () => {
  // Проверяет, что максимум залива ограничен свободным местом бака и выбранным топливом.
  it("uses selected fuel and free object fuel capacity", () => {
    const amount = getControlPanelFuelFillMaxAmount({
      object: object(20, 50),
      fuelTankGroup: fuelTank(),
      referenceData: referenceData(),
      selectedItemGroupIds: [1, 2, 3],
      itemGroups: [
        { ID: 1, ContainerEquipmentGroupID: 10, ContentItemModelID: 100, Count: 15 },
        { ID: 2, ContainerEquipmentGroupID: 10, ContentItemModelID: 100, Count: 40 },
        { ID: 3, ContainerEquipmentGroupID: 10, ContentItemModelID: 300, Count: 100 },
      ],
    });

    expect(amount).toBe(30);
  });
});
