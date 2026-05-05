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
  chatState: null,
  chatInputText: "",
  chatInputFocused: false,
  chatError: null,
  chatContextMenu: null,
  gameCursor: { visible: false, x: 0, y: 0 },
  chatScroll: { visible: false, thumbTopPercent: 0, thumbHeightPercent: 100, contentOffsetPx: 0, dragging: false },
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

  it("renders minimap zone and anchor statuses in the same monochrome HUD style", () => {
    const root = document.createElement("div");
    document.body.append(root);

    dispose = render(() => <GameUi state={state} />, root);

    const zone = root.querySelector(".minimap-status__zone");
    const anchor = root.querySelector(".minimap-status__anchor");

    expect(zone?.textContent).toBe("PVE");
    expect(Array.from(zone?.classList ?? [])).toContain("minimap-status__item");
    expect(Array.from(anchor?.classList ?? [])).toContain("minimap-status__item");
  });

  it("renders chat panel with tabs, selected history and local input text", () => {
    const root = document.createElement("div");
    document.body.append(root);
    const chatState = (): GameUiState => ({
      ...state(),
      chatState: {
        type: "chatState",
        selectedChatId: 2,
        tabs: [
          { chatId: 1, title: "Server", communityTypeAcronym: "Server", duoChatKey: "", messages: [] },
          {
            chatId: 2,
            title: "Pilot2",
            communityTypeAcronym: "Duo",
            duoChatKey: "1:2",
            messages: [
              { id: 10, chatId: 2, senderCharacterId: 1, senderNickname: "Pilot1", messageTypeAcronym: "FromCharacter", text: "old", color: "D8F3FF", sentTime: "" },
              { id: 11, chatId: 2, senderCharacterId: 2, senderNickname: "Pilot2", messageTypeAcronym: "FromCharacter", text: "new", color: "E8FFD8", sentTime: "" },
            ],
          },
        ],
      },
      chatInputText: "draft",
      chatInputFocused: true,
      chatError: "Адресат не найден",
      chatContextMenu: { chatId: 2, communityTypeAcronym: "Duo", x: 120, y: 220 },
      gameCursor: { visible: true, x: 320, y: 240 },
      chatScroll: { visible: true, thumbTopPercent: 25, thumbHeightPercent: 60, contentOffsetPx: 42, dragging: true },
    });

    dispose = render(() => <GameUi state={chatState} />, root);

    expect(root.querySelector(".chat-panel")).not.toBeNull();
    expect(Array.from(root.querySelectorAll(".chat-tab")).map((tab) => tab.textContent)).toEqual(["SServer", "DPilot2"]);
    expect(root.querySelector(".chat-tab.is-selected")?.textContent).toBe("DPilot2");
    expect(Array.from(root.querySelectorAll(".chat-message__text")).map((message) => message.textContent)).toEqual(["old", "new"]);
    expect(root.querySelector(".chat-input__text")?.textContent).toBe("draft");
    expect(root.querySelector(".chat-error")?.textContent).toBe("Адресат не найден");
    expect(root.querySelector(".chat-context-menu__item")?.textContent).toBe("Закрыть");
    expect(root.querySelector(".chat-scrollbar__thumb")).not.toBeNull();
    expect(Array.from(root.querySelector(".chat-scrollbar")?.classList ?? [])).toContain("is-dragging");
    expect(root.querySelector<HTMLElement>(".chat-messages__content")?.style.transform).toBe("translateY(42px)");
    expect(root.querySelector(".game-cursor")).not.toBeNull();
  });
});
