import { describe, expect, it } from "vitest";
import type { EquipmentGroup, ReferenceDataMessage } from "../network/protocol";
import { normalizeControlPanelUsageSelection } from "./controlPanelUsageSelection";

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

const referenceData = (): ReferenceDataMessage => ({
  // Создаёт справочники с контейнером, топливным баком и внешним предметом.
  type: "referenceData",
  NpcClan: { MaxID: 0, Items: {} },
  CosmicObjectType: { MaxID: 0, Items: {} },
  Itemtype: {
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
      "10": { ID: 10, ItemtypeID: 1, Acronym: "SmallContainer" },
      "20": { ID: 20, ItemtypeID: 2, Acronym: "SmallFuelTank" },
      "30": { ID: 30, ItemtypeID: 3, Acronym: "Ore" },
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
      selection: { leftContainerGroupId: null, rightEquipmentGroupId: null },
    });

    expect(selection).toEqual({ leftContainerGroupId: 10, rightEquipmentGroupId: 10 });
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
      selection: { leftContainerGroupId: 10, rightEquipmentGroupId: 20 },
    });

    expect(selection).toEqual({ leftContainerGroupId: 10, rightEquipmentGroupId: 20 });
  });
});
