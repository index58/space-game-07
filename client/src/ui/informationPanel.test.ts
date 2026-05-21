import { describe, expect, it } from "vitest";
import type { CosmicObject, ReferenceDataMessage } from "../network/protocol";
import { getInformationPanelView } from "./informationPanel";

const object = (partial: Partial<CosmicObject> = {}): CosmicObject => ({
  ID: 1,
  Title: "Ship",
  CosmicObjectModelID: 10,
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

const referenceData = {
  type: "referenceData",
  NpcClan: { MaxID: 1, Items: { "1": { ID: 1, TitleRu: "Клан добытчиков" } } },
  CosmicObjectType: {
    MaxID: 2,
    Items: {
      "1": { ID: 1, Acronym: "Ship" },
      "2": { ID: 2, Acronym: "Asteroid" },
    },
  },
  ItemType: { MaxID: 0, Items: {} },
  CosmicObjectModel: {
    MaxID: 20,
    Items: {
      "10": { ID: 10, CosmicObjectTypeID: 1, TitleRu: "Корабль", TextureWidth: 40, TextureHeight: 40, TextureBodyOriginX: 20, TextureBodyOriginY: 20, TextureScale: 1, BodyWidth: 20, BodyLength: 20 },
      "20": { ID: 20, CosmicObjectTypeID: 2, TitleRu: "Астероид", TextureWidth: 40, TextureHeight: 40, TextureBodyOriginX: 20, TextureBodyOriginY: 20, TextureScale: 1, BodyWidth: 20, BodyLength: 20 },
    },
  },
  ItemModel: { MaxID: 0, Items: {} },
  Blueprint: { MaxID: 0, Items: {} },
  BlueprintComponent: { MaxID: 0, Items: {} },
  Schema: { MaxID: 0, Items: {} },
  SchemaComponent: { MaxID: 0, Items: {} },
} as unknown as ReferenceDataMessage;

describe("getInformationPanelView", () => {
  // Проверяет, что панель выбирает ближайшее физическое тело на 100-метровом луче перед кораблем.
  it("selects nearest object touched by the forward probe", () => {
    const selfObject = object({ ID: 1, CosmicObjectModelID: 10, X: 0, Y: 0, Rotation: 0 });
    const farther = object({ ID: 3, Title: "Far", CosmicObjectModelID: 20, X: 0, Y: 80 });
    const nearer = object({ ID: 2, Title: "Near", CosmicObjectModelID: 20, X: 0, Y: 55, OwnerNpcClanID: 1 });

    const view = getInformationPanelView({ selfObject, objects: [selfObject, farther, nearer], referenceData });

    expect(view?.object.ID).toBe(2);
    expect(view?.rows).toEqual([
      { label: "Название", value: "Near" },
      { label: "Модель", value: "Астероид" },
      { label: "NPC-клан", value: "Клан добытчиков" },
    ]);
  });

  // Проверяет, что панель не появляется для объекта за пределами 100 метров от носа корабля.
  it("does not show panel when probe misses all bodies", () => {
    const selfObject = object({ ID: 1, CosmicObjectModelID: 10, X: 0, Y: 0, Rotation: 0 });
    const target = object({ ID: 2, CosmicObjectModelID: 20, X: 0, Y: 150 });

    expect(getInformationPanelView({ selfObject, objects: [selfObject, target], referenceData })).toBeNull();
  });

  // Проверяет временное поле имени владельца, если сервер уже прислал его в снимке.
  it("shows owner name when snapshot object contains it", () => {
    const selfObject = object({ ID: 1, CosmicObjectModelID: 10, X: 0, Y: 0, Rotation: 0 });
    const target = { ...object({ ID: 2, CosmicObjectModelID: 20, X: 0, Y: 55 }), OwnerName: "Pilot2" };

    const view = getInformationPanelView({ selfObject, objects: [selfObject, target], referenceData });

    expect(view?.rows).toContainEqual({ label: "Владелец", value: "Pilot2" });
  });
});
