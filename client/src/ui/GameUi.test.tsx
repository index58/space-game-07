import { render } from "solid-js/web";
import { afterEach, describe, expect, it } from "vitest";
import type { CosmicObject, ReferenceDataMessage } from "../network/protocol";
import type { GameUiState } from "./gameUiState";
import { GameUi } from "./GameUi";

let dispose: (() => void) | null = null;

afterEach(() => {
  dispose?.();
  dispose = null;
  document.body.innerHTML = "";
});

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
  GeneratingPower: 10,
  ConsumingPower: 5,
  AlongForce: 0,
  AcrossForce: 0,
  Torque: 0,
  Enabled: true,
  LastReceivedDamageTime: 0,
  Anchored: false,
  Complexity: 0,
  OccupiedVolume: 0,
  MaxFuel: 100,
  Fuel: 50,
  Speed: 0,
  VelocityX: 0,
  VelocityY: 0,
  AngularSpeed: 0,
  TargetRotation: 0,
  ...partial,
});

const referenceData = {
  type: "referenceData",
  NpcClan: { MaxID: 0, Items: {} },
  CosmicObjectType: {
    MaxID: 1,
    Items: {
      "1": { ID: 1, Acronym: "Ship" },
    },
  },
  Itemtype: { MaxID: 0, Items: {} },
  CosmicObjectModel: {
    MaxID: 10,
    Items: {
      "10": { ID: 10, CosmicObjectTypeID: 1 },
    },
  },
  ItemModel: { MaxID: 0, Items: {} },
  Blueprint: { MaxID: 0, Items: {} },
  BlueprintComponent: { MaxID: 0, Items: {} },
  Schema: { MaxID: 0, Items: {} },
  SchemaComponent: { MaxID: 0, Items: {} },
} as unknown as ReferenceDataMessage;

const state = (): GameUiState => ({
  status: "connected",
  selfObject: object(),
  objects: [object()],
  equipmentGroups: [],
  selectedPilotToolIndex: 0,
  referenceData,
  textureFilePath: null,
  fps: 60,
  zoom: 4,
});

describe("GameUi", () => {
  it("renders every top-level game panel through the shared HUD panel shell", () => {
    const root = document.createElement("div");
    document.body.append(root);

    dispose = render(() => <GameUi state={state} />, root);

    const panels = Array.from(root.querySelectorAll(".hud-panel"));

    expect(panels.map((panel) => panel.getAttribute("aria-label") ?? panel.id)).toEqual([
      "Основные показатели посещаемого объекта",
      "Панель инструментов пилота",
      "Мини-карта",
      "debug-overlay",
    ]);
    expect(panels.map((panel) => Array.from(panel.classList))).toEqual([
      ["hud-panel", "hud-panel--left-bottom", "object-indicators"],
      ["hud-panel", "hud-panel--bottom-center", "pilot-toolbar"],
      ["hud-panel", "hud-panel--right-bottom", "minimap"],
      ["hud-panel", "hud-panel--left-top", "debug-overlay"],
    ]);
  });

  it("renders the minimap anchor status as a sea anchor icon", () => {
    const root = document.createElement("div");
    document.body.append(root);

    dispose = render(() => <GameUi state={state} />, root);

    const anchorIcon = root.querySelector(".minimap-status__anchor svg");

    expect(anchorIcon?.getAttribute("viewBox")).toBe("0 0 24 24");
    expect(anchorIcon?.querySelector('[data-icon-part="anchor-ring"]')).not.toBeNull();
    expect(anchorIcon?.querySelector('[data-icon-part="anchor-stock"]')).not.toBeNull();
    expect(anchorIcon?.querySelector('[data-icon-part="anchor-flukes"]')).not.toBeNull();
  });
});
