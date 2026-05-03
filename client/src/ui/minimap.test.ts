import { describe, expect, it } from "vitest";
import type { CosmicObject, ReferenceDataMessage } from "../network/protocol";
import { getMinimapView } from "./minimap";

const referenceData = {
  CosmicObjectType: {
    Items: {
      "1": { ID: 1, Acronym: "Ship" },
      "2": { ID: 2, Acronym: "Station" },
      "3": { ID: 3, Acronym: "Asteroid" },
      "4": { ID: 4, Acronym: "Loot" },
    },
  },
  CosmicObjectModel: {
    Items: {
      "10": { ID: 10, CosmicObjectTypeID: 1 },
      "20": { ID: 20, CosmicObjectTypeID: 2 },
      "30": { ID: 30, CosmicObjectTypeID: 3 },
      "40": { ID: 40, CosmicObjectTypeID: 4 },
    },
  },
} as unknown as ReferenceDataMessage;

const object = (partial: Partial<CosmicObject>): CosmicObject => ({
  ID: 0,
  Title: "",
  CosmicObjectModelID: 10,
  OwnerCharacterID: 0,
  OwnerNpcClanID: 0,
  CreatorCharacterID: 0,
  Mass: 0,
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
  ...partial,
});

describe("getMinimapView", () => {
  it("центрирует карту на посещаемом объекте и масштабирует точки в пределах панели", () => {
    const selfObject = object({ ID: 1, X: 100, Y: 200, CosmicObjectModelID: 10, OwnerCharacterID: 7 });
    const view = getMinimapView({
      selfObject,
      objects: [
        selfObject,
        object({ ID: 2, X: 300, Y: 300, CosmicObjectModelID: 30 }),
        object({ ID: 3, X: 5000, Y: 200, CosmicObjectModelID: 30 }),
      ],
      referenceData,
    });

    expect(view.points).toEqual([
      { id: 1, xPercent: 50, yPercent: 50, kind: "owned", isSelf: true },
      { id: 2, xPercent: 52, yPercent: 49, kind: "asteroid", isSelf: false },
      { id: 3, xPercent: 99, yPercent: 50, kind: "asteroid", isSelf: false },
    ]);
  });

  it("оставляет объект сверху карты при нулевом повороте, если он сверху экрана", () => {
    const selfObject = object({ ID: 1, X: 100, Y: 200, CosmicObjectModelID: 10, OwnerCharacterID: 7 });
    const view = getMinimapView({
      selfObject,
      objects: [
        selfObject,
        object({ ID: 2, X: 100, Y: 400, CosmicObjectModelID: 30 }),
      ],
      referenceData,
    });

    expect(view.points[1]).toMatchObject({
      id: 2,
      xPercent: 50,
      yPercent: 48,
    });
  });

  it("поворачивает стоячие объекты против часовой стрелки при повороте посещаемого объекта по часовой стрелке", () => {
    const selfObject = object({
      ID: 1,
      X: 100,
      Y: 200,
      Rotation: Math.PI / 2,
      CosmicObjectModelID: 10,
      OwnerCharacterID: 7,
    });
    const view = getMinimapView({
      selfObject,
      objects: [
        selfObject,
        object({ ID: 2, X: 100, Y: 400, CosmicObjectModelID: 30 }),
      ],
      referenceData,
    });

    expect(view.points[1]).toMatchObject({
      id: 2,
      xPercent: 48,
      yPercent: 50,
      kind: "asteroid",
      isSelf: false,
    });
  });

  it("держит север по центру компаса при нулевом повороте", () => {
    const view = getMinimapView({
      selfObject: object({ ID: 1, Rotation: 0, CosmicObjectModelID: 10 }),
      objects: [],
      referenceData,
    });

    expect(view.compassMarks.find((mark) => mark.label === "N")).toEqual({ label: "N", xPercent: 50 });
  });

  it("смещает деления компаса по горизонтали вместо поворота всей панели", () => {
    const view = getMinimapView({
      selfObject: object({ ID: 1, Rotation: Math.PI / 2, CosmicObjectModelID: 10 }),
      objects: [],
      referenceData,
    });

    expect(view.compassMarks.find((mark) => mark.label === "E")).toEqual({ label: "E", xPercent: 50 });
    expect(view.compassMarks.find((mark) => mark.label === "N")).toEqual({ label: "N", xPercent: 25 });
  });

  it("показывает только станции, корабли, астероиды и свой лут", () => {
    const selfObject = object({ ID: 1, CosmicObjectModelID: 10, OwnerCharacterID: 7 });
    const view = getMinimapView({
      selfObject,
      objects: [
        selfObject,
        object({ ID: 2, CosmicObjectModelID: 40, OwnerCharacterID: 7 }),
        object({ ID: 3, CosmicObjectModelID: 40, OwnerCharacterID: 8 }),
      ],
      referenceData,
    });

    expect(view.points.map((point) => point.id)).toEqual([1, 2]);
  });

  it("выбирает вид точки по типу и владельцу объекта", () => {
    const selfObject = object({ ID: 1, CosmicObjectModelID: 10, OwnerCharacterID: 7 });
    const view = getMinimapView({
      selfObject,
      objects: [
        selfObject,
        object({ ID: 2, CosmicObjectModelID: 20, OwnerNpcClanID: 1 }),
        object({ ID: 3, CosmicObjectModelID: 10, OwnerCharacterID: 8 }),
        object({ ID: 4, CosmicObjectModelID: 30 }),
        object({ ID: 5, CosmicObjectModelID: 20 }),
      ],
      referenceData,
    });

    expect(view.points.map((point) => point.kind)).toEqual([
      "owned",
      "npc",
      "player",
      "asteroid",
      "neutral",
    ]);
  });
});
