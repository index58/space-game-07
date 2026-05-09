import { describe, expect, it } from "vitest";
import type { CosmicObject, EquipmentGroup } from "../network/protocol";
import { applyControlPanelPendingToEquipmentGroups, applyControlPanelPendingToObject, pruneControlPanelPending, rejectControlPanelPending, type ControlPanelPendingState } from "./controlPanelMutations";

const object = (partial: Partial<CosmicObject> = {}): CosmicObject => ({
  ID: 1,
  Title: "Ship",
  CosmicObjectModelID: 1,
  OwnerCharacterID: 1,
  OwnerNpcClanID: 0,
  CreatorCharacterID: 1,
  Mass: 1,
  Capacity: 1,
  MaxArmor: 1,
  MaxSpeed: 1,
  MaxAngularSpeed: 1,
  X: 0,
  Y: 0,
  Rotation: 0,
  Armor: 1,
  MaxAlongForce: 1,
  MaxAcrossForce: 1,
  MaxTorque: 1,
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

const equipmentGroup = (partial: Partial<EquipmentGroup> = {}): EquipmentGroup => ({
  ID: 11,
  CosmicObjectID: 1,
  Title: "Generator",
  EquipmentItemModelID: 2,
  Count: 4,
  EnabledCount: 2,
  Enabled: true,
  Active: false,
  LastRechargeStartTime: 0,
  ...partial,
});

describe("control panel mutations", () => {
  // Проверяет, что старый снимок не перетирает ожидающее изменение объекта.
  it("keeps pending object value until matching ack arrives", () => {
    const pending: ControlPanelPendingState = {
      object: {
        enabled: { sessionId: "session-1", seq: 2, value: false },
      },
      equipment: {},
    };

    expect(applyControlPanelPendingToObject(object({ Enabled: true }), pending)?.Enabled).toBe(false);
    expect(pruneControlPanelPending(pending, { sessionId: "session-1", lastAppliedSeq: 1 }).object.enabled).toBeDefined();
    expect(pruneControlPanelPending(pending, { sessionId: "session-1", lastAppliedSeq: 2 }).object.enabled).toBeUndefined();
  });

  // Проверяет, что старый снимок не перетирает ожидающее изменение оборудования.
  it("keeps pending equipment value until matching ack arrives", () => {
    const pending: ControlPanelPendingState = {
      object: {},
      equipment: {
        11: {
          enabledCount: { sessionId: "session-1", seq: 3, value: 4 },
        },
      },
    };

    const groups = applyControlPanelPendingToEquipmentGroups([equipmentGroup({ EnabledCount: 2 })], pending);

    expect(groups[0].EnabledCount).toBe(4);
    expect(pruneControlPanelPending(pending, { sessionId: "session-2", lastAppliedSeq: 99 }).equipment[11]?.enabledCount).toBeDefined();
    expect(pruneControlPanelPending(pending, { sessionId: "session-1", lastAppliedSeq: 3 }).equipment[11]?.enabledCount).toBeUndefined();
  });

  // Проверяет, что отказ сервера снимает только изменение с указанным номером.
  it("removes only rejected pending mutation", () => {
    const pending: ControlPanelPendingState = {
      object: {
        enabled: { sessionId: "session-1", seq: 2, value: false },
      },
      equipment: {
        11: {
          enabledCount: { sessionId: "session-1", seq: 3, value: 4 },
        },
      },
    };

    const next = rejectControlPanelPending(pending, { clientSessionId: "session-1", mutationSeq: 2 });

    expect(next.object.enabled).toBeUndefined();
    expect(next.equipment[11]?.enabledCount?.value).toBe(4);
  });
});
