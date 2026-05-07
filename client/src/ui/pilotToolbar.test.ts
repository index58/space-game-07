import { describe, expect, it } from "vitest";
import type { CosmicObject, EquipmentGroup, ReferenceDataMessage } from "../network/protocol";
import { getPilotToolbarView, getNextPilotToolIndex } from "./pilotToolbar";

const emptyObject = (): CosmicObject => ({
  ID: 10,
  Title: "Ship",
  CosmicObjectModelID: 1,
  OwnerCharacterID: 1,
  OwnerNpcClanID: 0,
  CreatorCharacterID: 1,
  Mass: 1,
  Capacity: 0,
  MaxArmor: 0,
  MaxSpeed: 0,
  MaxAngularSpeed: 0,
  X: 0,
  Y: 0,
  Rotation: 0,
  Armor: 0,
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
});

const group = (id: number, modelId: number, enabledCount: number, objectId = 10): EquipmentGroup => ({
  ID: id,
  CosmicObjectID: objectId,
  Title: "Group",
  EquipmentItemModelID: modelId,
  Count: enabledCount,
  EnabledCount: enabledCount,
  Enabled: true,
  Active: false,
  LastRechargeStartTime: 0,
});

const referenceData: ReferenceDataMessage = {
  type: "referenceData",
  NpcClan: { MaxID: 0, Items: {} },
  CosmicObjectType: { MaxID: 0, Items: {} },
  Itemtype: {
    MaxID: 2,
    Items: {
      "1": { ID: 1, Acronym: "Weapon", IsPilotInstrument: true },
      "2": { ID: 2, Acronym: "Container", IsPilotInstrument: false },
    },
  },
  CosmicObjectModel: { MaxID: 0, Items: {} },
  ItemModel: {
    MaxID: 3,
    Items: {
      "101": { ID: 101, ItemtypeID: 1, Acronym: "Laser", TitleRu: "Лазер", IconFilePath: "", MagazineCapacity: 6 },
      "102": { ID: 102, ItemtypeID: 1, Acronym: "Drill", TitleRu: "Бур", IconFilePath: "", MagazineCapacity: 0 },
      "201": { ID: 201, ItemtypeID: 2, Acronym: "Box", TitleRu: "Контейнер", IconFilePath: "" },
    },
  },
  Blueprint: { MaxID: 0, Items: {} },
  BlueprintComponent: { MaxID: 0, Items: {} },
  Schema: { MaxID: 0, Items: {} },
  SchemaComponent: { MaxID: 0, Items: {} },
  ActionType: { MaxID: 0, Items: {} },
  InputEventType: { MaxID: 0, Items: {} },
  DefaultActionInputSetting: { MaxID: 0, Items: {} },
};

describe("pilot toolbar", () => {
  it("groups pilot tools by item model and sums enabled counts", () => {
    const toolbar = getPilotToolbarView({
      selfObject: emptyObject(),
      equipmentGroups: [
        group(5, 101, 2),
        group(7, 101, 3),
        group(6, 201, 8),
        group(8, 102, 1),
      ],
      referenceData,
      selectedToolIndex: 1,
    });

    expect(toolbar.slots).toHaveLength(10);
    expect(toolbar.slots[0].tool).toMatchObject({
      itemModelId: 101,
      acronym: "Laser",
      title: "Лазер",
      enabledCount: 5,
    });
    expect(toolbar.slots[0].isSelected).toBe(false);
    expect(toolbar.slots[1].tool).toMatchObject({
      itemModelId: 102,
      enabledCount: 1,
    });
    expect(toolbar.slots[1].isSelected).toBe(true);
    expect(toolbar.magazine).toBeNull();
  });

  it("wraps selected tool index over available tools", () => {
    expect(getNextPilotToolIndex(0, -1)).toBe(9);
    expect(getNextPilotToolIndex(9, 1)).toBe(0);
    expect(getNextPilotToolIndex(4, 1)).toBe(5);
  });

  it("can select an empty slot", () => {
    const toolbar = getPilotToolbarView({
      selfObject: emptyObject(),
      equipmentGroups: [group(5, 101, 2)],
      referenceData,
      selectedToolIndex: 4,
    });

    expect(toolbar.slots[4].tool).toBeNull();
    expect(toolbar.slots[4].isSelected).toBe(true);
    expect(toolbar.magazine).toBeNull();
  });
});
