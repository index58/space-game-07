import { render } from "solid-js/web";
import { createSignal } from "solid-js";
import { afterEach, describe, expect, it } from "vitest";
import type { CosmicObject, ReferenceDataMessage } from "../network/protocol";
import { createInitialUiKitDemoState } from "../ui-kit/showcaseState";
import { createGameUiController, type GameUiState } from "./gameUiState";
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
  // Проверяет, что окно количества изготовления использует общий шаблон выбора числа.
  it("renders control panel constructor amount dialog", () => {
    const root = document.createElement("div");
    document.body.append(root);

    dispose = render(() => <GameUi state={() => ({
      ...state(),
      controlPanelVisible: true,
      controlPanelConstructorProduceDialogOpen: true,
      controlPanelConstructorProduceMaxAmount: 100,
      controlPanelFuelDrainAmount: 5,
      controlPanelFuelDrainAmountText: "5",
      controlPanelFuelDrainAmountSelectionStart: 1,
      controlPanelFuelDrainAmountSelectionEnd: 1,
    })} />, root);

    expect(root.querySelector("#control-panel-constructor-produce-dialog .control-panel-fuel-drain-dialog__title")?.textContent).toBe("Изготовление");
    expect(root.querySelector("#control-panel-fuel-drain-amount-input .ui-kit-text-input__text")?.textContent).toBe("5");
    expect(root.querySelector("#control-panel-fuel-drain-amount-slider .ui-kit-slider__fill")).not.toBeNull();
    expect(root.querySelector("#control-panel-constructor-produce-ok")?.textContent).toBe("ОК");
    expect(root.querySelector("#control-panel-constructor-produce-cancel")?.textContent).toBe("Отмена");
  });

  // Проверяет, что окно количества деконструкции использует общий шаблон выбора числа.
  it("renders control panel item deconstruction amount dialog", () => {
    const root = document.createElement("div");
    document.body.append(root);

    dispose = render(() => <GameUi state={() => ({
      ...state(),
      controlPanelVisible: true,
      controlPanelItemDeconstructionDialogOpen: true,
      controlPanelItemDeconstructionMaxAmount: 12,
      controlPanelFuelDrainAmount: 4,
      controlPanelFuelDrainAmountText: "4",
      controlPanelFuelDrainAmountSelectionStart: 1,
      controlPanelFuelDrainAmountSelectionEnd: 1,
    })} />, root);

    expect(root.querySelector("#control-panel-item-deconstruction-dialog .control-panel-fuel-drain-dialog__title")?.textContent).toBe("Деконструкция");
    expect(root.querySelector("#control-panel-fuel-drain-amount-input .ui-kit-text-input__text")?.textContent).toBe("4");
    expect(root.querySelector("#control-panel-fuel-drain-amount-slider .ui-kit-slider__fill")).not.toBeNull();
    expect(root.querySelector("#control-panel-item-deconstruction-ok")?.textContent).toBe("ОК");
    expect(root.querySelector("#control-panel-item-deconstruction-cancel")?.textContent).toBe("Отмена");
  });

const visibleControlText = (element: Element | null): string | null =>
  element?.querySelector(".ui-kit-text-input__text")?.textContent ?? element?.textContent ?? null;

const referenceData = {
  type: "referenceData",
  NpcClan: { MaxID: 0, Items: {} },
  CosmicObjectType: {
    MaxID: 1,
    Items: {
      "1": { ID: 1, Acronym: "Ship" },
    },
  },
  ItemType: { MaxID: 0, Items: {} },
  CosmicObjectModel: {
    MaxID: 10,
    Items: {
      "10": { ID: 10, CosmicObjectTypeID: 1, TitleRu: "Корабль", TextureWidth: 40, TextureHeight: 40, TextureBodyOriginX: 20, TextureBodyOriginY: 20, TextureScale: 1, BodyWidth: 20, BodyLength: 20 },
      "11": { ID: 11, CosmicObjectTypeID: 1, TitleRu: "Цель", TextureWidth: 40, TextureHeight: 40, TextureBodyOriginX: 20, TextureBodyOriginY: 20, TextureScale: 1, BodyWidth: 20, BodyLength: 20 },
    },
  },
  ItemModel: { MaxID: 0, Items: {} },
  Blueprint: { MaxID: 0, Items: {} },
  BlueprintComponent: { MaxID: 0, Items: {} },
  Schema: { MaxID: 0, Items: {} },
  SchemaComponent: { MaxID: 0, Items: {} },
  ActionType: {
    MaxID: 1,
    Items: {
      "1": { ID: 1, TitleRu: "Продольная тяга вперед", TitleEn: "Forward thrust", Acronym: "ThrustForward" },
    },
  },
  InputEventType: {
    MaxID: 1,
    Items: {
      "1": { ID: 1, TitleRu: "Клавиша W", TitleEn: "Keyboard W", Acronym: "KeyboardKeyW", SystemStringValue: "KeyboardEvent.code:KeyW", SystemIntegerValue: 0 },
    },
  },
  DefaultActionInputSetting: {
    MaxID: 1,
    Items: {
      "1": { ID: 1, ActionTypeID: 1, InputEventTypeID: 1 },
    },
  },
} as unknown as ReferenceDataMessage;

const rect = (width: number): DOMRect => ({
  x: 0,
  y: 0,
  left: 0,
  top: 0,
  right: width,
  bottom: 20,
  width,
  height: 20,
  toJSON: () => ({}),
} as DOMRect);

const state = (): GameUiState => ({
  status: "connected",
  nowMs: 0,
  reloadDisplayStartMsByGroupId: {},
  selfObject: object(),
  objects: [object()],
  equipmentGroups: [],
  itemGroups: [],
  constructorProductionJobs: [],
  selectedPilotToolIndex: 0,
  referenceData,
  textureFilePath: null,
  chatState: null,
  chatInputText: "",
  chatCursorIndex: 0,
  chatSelectionStart: 0,
  chatSelectionEnd: 0,
  chatInputFocused: false,
  chatError: null,
  chatErrorSeq: 0,
  chatContextMenu: null,
  gameCursor: { visible: false, x: 0, y: 0 },
  hoveredGameUiControlId: null,
  chatScroll: { visible: false, thumbTopPercent: 0, thumbHeightPercent: 100, contentOffsetPx: 0, dragging: false },
  uiKitShowcaseVisible: false,
  settingsVisible: false,
  controlPanelVisible: false,
  selectedSettingsTab: "input",
  selectedControlPanelTab: "object",
  selectedControlPanelEquipmentTab: "setup",
  selectedControlPanelEquipmentGroupId: null,
  selectedControlPanelUsageLeftObjectId: null,
  selectedControlPanelUsageRightObjectId: null,
  selectedControlPanelConstructorMaterialObjectId: null,
  selectedControlPanelConstructorProductObjectId: null,
  selectedControlPanelUsageLeftContainerGroupId: null,
  selectedControlPanelUsageRightEquipmentGroupId: null,
  openControlPanelUsageSelect: null,
  selectedControlPanelUsageLeftItemGroupIds: [],
  selectedControlPanelUsageRightItemGroupIds: [],
  selectedControlPanelConstructorMaterialContainerGroupId: null,
  selectedControlPanelConstructorProductContainerGroupId: null,
  selectedControlPanelConstructorTab: "items",
  selectedControlPanelConstructorSchemaId: null,
  selectedControlPanelConstructorBlueprintId: null,
  selectedControlPanelConstructorMainJobId: null,
  controlPanelFuelDrainDialogOpen: false,
  controlPanelFuelFillDialogOpen: false,
  controlPanelContainerTransferDialogOpen: false,
  controlPanelConstructorProduceDialogOpen: false,
  controlPanelItemDeconstructionDialogOpen: false,
  controlPanelContainerTransferMaxAmount: 0,
  controlPanelFuelFillMaxAmount: 0,
  controlPanelConstructorProduceMaxAmount: 100,
  controlPanelItemDeconstructionMaxAmount: 0,
  controlPanelFuelDrainAmount: 0,
  controlPanelFuelDrainAmountText: "0",
  controlPanelFuelDrainAmountSelectionStart: 1,
  controlPanelFuelDrainAmountSelectionEnd: 1,
  controlPanelFuelDrainAmountFocused: false,
  controlPanelEquipmentEnabledDrafts: {},
  controlPanelEquipmentEnabledCountDrafts: {},
  controlPanelEquipmentListScroll: { visible: false, thumbTopPercent: 0, thumbHeightPercent: 100, contentOffsetPx: 0, dragging: false },
  listScroll: {},
  controlPanelObjectEnabled: true,
  controlPanelEquipmentTitleText: "",
  controlPanelEquipmentTitleSelectionStart: 0,
  controlPanelEquipmentTitleSelectionEnd: 0,
  controlPanelEquipmentTitleFocused: false,
  controlPanelObjectTitleText: "Ship",
  controlPanelObjectTitleSelectionStart: 4,
  controlPanelObjectTitleSelectionEnd: 4,
  controlPanelObjectTitleFocused: false,
  inputSettingsValues: { 1: 1 },
  openInputSettingsActionId: null,
  inputSettingsError: null,
  inputSettingsSaving: false,
  inputSettingsScroll: { visible: false, thumbTopPercent: 0, thumbHeightPercent: 100, contentOffsetPx: 0, dragging: false },
  inputSettingsDropdownScroll: { visible: false, thumbTopPercent: 0, thumbHeightPercent: 100, contentOffsetPx: 0, dragging: false },
  uiKitDemoState: createInitialUiKitDemoState(),
  uiControls: [],
  fps: 60,
  zoom: 4,
});

// Создаёт справочник с нужным числом действий, чтобы проверить раскладку строк настроек.
const referenceDataWithActionTitles = (titles: string[]): ReferenceDataMessage => ({
  ...referenceData,
  ActionType: {
    MaxID: titles.length,
    Items: Object.fromEntries(titles.map((title, index) => {
      const id = index + 1;
      return [String(id), { ID: id, TitleRu: title, TitleEn: title, Acronym: `Action${id}` }];
    })),
  },
  DefaultActionInputSetting: {
    MaxID: titles.length,
    Items: Object.fromEntries(titles.map((_, index) => {
      const id = index + 1;
      return [String(id), { ID: id, ActionTypeID: id, InputEventTypeID: 1 }];
    })),
  },
}) as unknown as ReferenceDataMessage;

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

  // Проверяет, что панель пилота показывает процесс подготовки зарядов отдельным состоянием.
  it("renders pilot toolbar reload state", () => {
    const root = document.createElement("div");
    document.body.append(root);

    dispose = render(() => <GameUi state={() => ({
      ...state(),
      nowMs: 4000,
      referenceData: {
        ...referenceData,
        ItemType: {
          MaxID: 1,
          Items: {
            "1": { ID: 1, Acronym: "Weapon", IsPilotInstrument: true, IsInternalUsable: false },
          },
        },
        ItemModel: {
          MaxID: 1,
          Items: {
            "1": { ID: 1, ItemTypeID: 1, Acronym: "Laser", TitleRu: "Лазер", IconFilePath: "", MagazineCapacity: 6, RechargeTime: 6 },
          },
        },
      },
      equipmentGroups: [
        { ID: 10, CosmicObjectID: 1, Title: "Лазер", EquipmentItemModelID: 1, Count: 2, EnabledCount: 2, Enabled: true, Active: false, LastRechargeStartTime: 1000, MagazineCount: 0 },
      ],
    })} />, root);

    expect(root.querySelector(".pilot-toolbar__magazine")?.classList.contains("is-reloading")).toBe(true);
    expect(root.querySelector(".pilot-toolbar__magazine-fill")?.getAttribute("style")).toContain("width: 50%");
    expect(root.querySelector(".pilot-toolbar__magazine-value")?.textContent).toBe("Перезарядка");
  });

  it("renders docking process progress as increasing fill", () => {
    const root = document.createElement("div");
    document.body.append(root);

    dispose = render(() => <GameUi state={() => ({
      ...state(),
      dockingWindow: { kind: "process", role: "sender", startedAtMs: 0, durationMs: 10000 },
    })} />, root);

    expect(root.querySelector(".docking-window")).not.toBeNull();
    expect(root.querySelector(".docking-window__fill")?.classList.contains("is-increasing")).toBe(true);
  });

  // Проверяет, что входящий запрос выделяет рамкой только сами клавиши.
  it("shows only keys in boxes for receiver docking request hints", () => {
    const root = document.createElement("div");
    document.body.append(root);

    dispose = render(() => <GameUi state={() => ({
      ...state(),
      dockingWindow: { kind: "request", role: "receiver", startedAtMs: 0, durationMs: 10000 },
    })} />, root);

    expect(Array.from(root.querySelectorAll(".docking-window__hint")).map((hint) => hint.textContent)).toEqual([
      "Одобрить — Alt + 1",
      "Отклонить — Alt + 2",
    ]);
    expect(root.querySelector(".docking-window__hint-action--approve")?.textContent).toBe("Одобрить");
    expect(root.querySelector(".docking-window__hint-action--reject")?.textContent).toBe("Отклонить");
    expect(Array.from(root.querySelectorAll(".docking-window__hint-key")).map((key) => key.textContent)).toEqual([
      "Alt",
      "1",
      "Alt",
      "2",
    ]);
    expect(root.querySelector(".docking-window__hint")?.classList.contains("docking-window__hint-key")).toBe(false);
  });

  // Проверяет, что входящий запрос обмена виден второму игроку и использует те же клавиши ответа.
  it("shows exchange request window for receiver", () => {
    const root = document.createElement("div");
    document.body.append(root);

    dispose = render(() => <GameUi state={() => ({
      ...state(),
      dockingWindow: { kind: "exchangeRequest", role: "receiver", startedAtMs: 0, durationMs: 10000 } as GameUiState["dockingWindow"],
    })} />, root);

    expect(root.querySelector(".docking-window__title")?.textContent).toBe("Запрос обмена");
    expect(root.querySelector(".docking-window__text")?.textContent).toBe("Входящий запрос на обмен");
    expect(Array.from(root.querySelectorAll(".docking-window__hint-key")).map((key) => key.textContent)).toEqual(["Alt", "1", "Alt", "2"]);
    expect(root.querySelector(".docking-window__fill")?.classList.contains("is-decreasing")).toBe(true);
    expect((root.querySelector(".docking-window__fill") as HTMLElement | null)?.style.animationDuration).toBe("10s");
  });

  // Проверяет, что окно обмена показывает общий выбор количества для переноса в очередь.
  it("renders exchange amount dialog", () => {
    const root = document.createElement("div");
    document.body.append(root);

    dispose = render(() => <GameUi state={() => ({
      ...state(),
      exchangeState: {
        selfObjectId: 1,
        otherObjectId: 2,
        selfNickname: "Pilot1",
        otherNickname: "Pilot2",
        selfReceiverContainerEquipmentGroupId: 0,
        selfSourceContainerEquipmentGroupId: 0,
        selfConfirmed: false,
        otherConfirmed: false,
        notEnoughSpace: false,
        selfQueue: [],
        otherQueue: [],
      },
      controlPanelContainerTransferDialogOpen: true,
      controlPanelContainerTransferMaxAmount: 7,
      controlPanelFuelDrainAmount: 3,
      controlPanelFuelDrainAmountText: "3",
      controlPanelFuelDrainAmountSelectionStart: 1,
      controlPanelFuelDrainAmountSelectionEnd: 1,
    })} />, root);

    expect(root.querySelector("#exchange-add-items-dialog .control-panel-fuel-drain-dialog__title")?.textContent).toBe("Перенос предметов");
    expect(root.querySelector("#control-panel-fuel-drain-amount-input .ui-kit-text-input__text")?.textContent).toBe("3");
    expect(root.querySelector("#control-panel-fuel-drain-amount-slider .ui-kit-slider__fill")).not.toBeNull();
    expect(root.querySelector("#exchange-add-items-ok")?.textContent).toBe("ОК");
    expect(root.querySelector("#exchange-add-items-cancel")?.textContent).toBe("Отмена");
  });

  // Проверяет, что списки предметов обмена используют общий шаблон списка с прокруткой.
  it("renders exchange item lists with shared scrollbars", () => {
    const root = document.createElement("div");
    document.body.append(root);

    dispose = render(() => <GameUi state={() => ({
      ...state(),
      exchangeState: {
        selfObjectId: 1,
        otherObjectId: 2,
        selfNickname: "Pilot1",
        otherNickname: "Pilot2",
        selfReceiverContainerEquipmentGroupId: 10,
        selfSourceContainerEquipmentGroupId: 11,
        selfConfirmed: false,
        otherConfirmed: false,
        notEnoughSpace: false,
        selfQueue: [],
        otherQueue: [],
      },
      referenceData: {
        ...referenceData,
        ItemModel: {
          MaxID: 1,
          Items: {
            "1": { ID: 1, ItemTypeID: 1, Acronym: "Melit", TitleRu: "Мелит" },
          },
        },
      } as ReferenceDataMessage,
      itemGroups: [
        { ID: 101, ContainerEquipmentGroupID: 10, ContentItemModelID: 1, Count: 5 },
        { ID: 201, ContainerEquipmentGroupID: 11, ContentItemModelID: 1, Count: 7 },
      ],
      listScroll: {
        "exchange-receiver-list": { visible: true, thumbTopPercent: 20, thumbHeightPercent: 40, contentOffsetPx: 12, dragging: false },
        "exchange-source-list": { visible: true, thumbTopPercent: 30, thumbHeightPercent: 35, contentOffsetPx: 18, dragging: true },
      },
    })} />, root);

    expect(root.querySelector<HTMLElement>("#exchange-receiver-list .ui-kit-list__content")?.style.transform).toBe("translateY(-12px)");
    expect(root.querySelector("#exchange-receiver-list-scrollbar")).not.toBeNull();
    expect(root.querySelector<HTMLElement>("#exchange-source-list .ui-kit-list__content")?.style.transform).toBe("translateY(-18px)");
    expect(Array.from(root.querySelector("#exchange-source-list-scrollbar")?.classList ?? [])).toContain("is-dragging");
  });

  // Проверяет, что окно обмена показывает подтверждение второго игрока под моей очередью.
  it("shows other player confirmation status in exchange window", () => {
    const root = document.createElement("div");
    document.body.append(root);

    dispose = render(() => <GameUi state={() => ({
      ...state(),
      exchangeState: {
        selfObjectId: 1,
        otherObjectId: 2,
        selfNickname: "Pilot1",
        otherNickname: "Pilot2",
        selfReceiverContainerEquipmentGroupId: 0,
        selfSourceContainerEquipmentGroupId: 0,
        selfConfirmed: false,
        otherConfirmed: true,
        notEnoughSpace: false,
        selfQueue: [],
        otherQueue: [],
      },
    })} />, root);

    expect(root.querySelector(".exchange-window__status")?.textContent).toBe("✓ Подтверждено");
    expect(root.querySelector(".exchange-window__status")?.classList.contains("is-confirmed")).toBe(true);
  });

  // Проверяет, что очереди обмена используют общий шаблон очереди с нижней полосой прогресса и прокруткой.
  it("renders exchange queues with shared queue progress and scrollbars", () => {
    const root = document.createElement("div");
    document.body.append(root);

    dispose = render(() => <GameUi state={() => ({
      ...state(),
      exchangeState: {
        selfObjectId: 1,
        otherObjectId: 2,
        selfNickname: "Pilot1",
        otherNickname: "Pilot2",
        selfReceiverContainerEquipmentGroupId: 0,
        selfSourceContainerEquipmentGroupId: 0,
        selfConfirmed: false,
        otherConfirmed: false,
        notEnoughSpace: false,
        selfQueue: [{ taskItemGroupId: 301, itemModelId: 1, count: 4, progress: 0.5, isReady: true }],
        otherQueue: [{ taskItemGroupId: 401, itemModelId: 1, count: 2, progress: 0.25, isReady: false }],
      },
      referenceData: {
        ...referenceData,
        ItemModel: {
          MaxID: 1,
          Items: {
            "1": { ID: 1, ItemTypeID: 1, Acronym: "Melit", TitleRu: "Мелит" },
          },
        },
      } as ReferenceDataMessage,
      listScroll: {
        "exchange-self-queue": { visible: true, thumbTopPercent: 10, thumbHeightPercent: 40, contentOffsetPx: 16, dragging: true },
        "exchange-other-queue": { visible: true, thumbTopPercent: 20, thumbHeightPercent: 45, contentOffsetPx: 12, dragging: false },
      },
    })} />, root);

    expect(root.querySelector("#exchange-self-queue-301")?.classList.contains("control-panel-constructor-queue__item")).toBe(true);
    expect(root.querySelector("#exchange-self-queue-301 .ui-kit-list__item-label-prefix")?.textContent).toBe("✓");
    expect(root.querySelector("#exchange-self-queue-301")?.getAttribute("style")).toContain("--constructor-queue-unit-progress: 0%");
    expect(root.querySelector("#exchange-other-queue-401")?.getAttribute("style")).toContain("--constructor-queue-unit-progress: 25%");
    expect(root.querySelector<HTMLElement>("#exchange-self-queue .ui-kit-list__content")?.style.transform).toBe("translateY(-16px)");
    expect(Array.from(root.querySelector("#exchange-self-queue-scrollbar")?.classList ?? [])).toContain("is-dragging");
    expect(root.querySelector("#exchange-other-queue-scrollbar")).not.toBeNull();
  });

  // Проверяет, что собственное подтверждение показывается зеленым текстом с галочкой.
  it("renders self confirmation with check mark", () => {
    const root = document.createElement("div");
    document.body.append(root);

    dispose = render(() => <GameUi state={() => ({
      ...state(),
      exchangeState: {
        selfObjectId: 1,
        otherObjectId: 2,
        selfNickname: "Pilot1",
        otherNickname: "Pilot2",
        selfReceiverContainerEquipmentGroupId: 0,
        selfSourceContainerEquipmentGroupId: 0,
        selfConfirmed: true,
        otherConfirmed: false,
        notEnoughSpace: false,
        selfQueue: [],
        otherQueue: [],
      },
    })} />, root);

    expect(root.querySelector(".exchange-window__confirmed")?.textContent).toBe("✓ Подтверждено");
  });

  // Проверяет, что после двух подтверждений отмена обмена становится недоступной на время переноса.
  it("disables exchange cancel button while items are moving", () => {
    const root = document.createElement("div");
    document.body.append(root);

    dispose = render(() => <GameUi state={() => ({
      ...state(),
      exchangeState: {
        selfObjectId: 1,
        otherObjectId: 2,
        selfNickname: "Pilot1",
        otherNickname: "Pilot2",
        selfReceiverContainerEquipmentGroupId: 0,
        selfSourceContainerEquipmentGroupId: 0,
        selfConfirmed: true,
        otherConfirmed: true,
        notEnoughSpace: false,
        selfQueue: [],
        otherQueue: [],
      },
    })} />, root);

    expect(root.querySelector("#exchange-cancel-button")?.classList.contains("is-disabled")).toBe(true);
  });

  it("does not render docking window after docking finish clears state", () => {
    const root = document.createElement("div");
    document.body.append(root);

    dispose = render(() => <GameUi state={() => ({
      ...state(),
      dockingWindow: null,
    })} />, root);

    expect(root.querySelector(".docking-window")).toBeNull();
  });

  // Проверяет, что при отсутствии подключения остаётся видимой только отладочная панель.
  it("hides panels and windows while server connection is missing", () => {
    const root = document.createElement("div");
    document.body.append(root);

    dispose = render(() => <GameUi state={() => ({
      ...state(),
      status: "waiting",
      chatState: {
        type: "chatState",
        selectedChatId: 1,
        tabs: [{ chatId: 1, title: "Server", communityTypeAcronym: "Server", duoChatKey: "", unreadCount: 0, messages: [] }],
      },
      gameCursor: { visible: true, x: 120, y: 80 },
      hoveredGameUiControlId: "control-panel-constructor-schema-list-1",
      uiKitShowcaseVisible: true,
      settingsVisible: true,
      controlPanelVisible: true,
      controlPanelFuelDrainDialogOpen: true,
      controlPanelConstructorProduceDialogOpen: true,
    })} />, root);

    expect(root.querySelector(".debug-overlay")).not.toBeNull();
    expect(Array.from(root.querySelectorAll(".hud-panel")).map((panel) => panel.className)).toEqual(["hud-panel hud-panel--left-top debug-overlay"]);
    expect(root.querySelector(".game-window-layer")).toBeNull();
    expect(root.querySelector(".chat-panel")).toBeNull();
    expect(root.querySelector(".object-indicators")).toBeNull();
    expect(root.querySelector(".pilot-toolbar")).toBeNull();
    expect(root.querySelector(".minimap")).toBeNull();
    expect(document.body.querySelector(".game-cursor")).toBeNull();
    expect(document.body.querySelector(".control-panel-constructor-recipe-tooltip")).toBeNull();
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

  // Проверяет, что информационная панель появляется справа по объекту перед носом корабля.
  it("renders information panel for object touched by forward probe", () => {
    const root = document.createElement("div");
    document.body.append(root);
    const infoState = (): GameUiState => ({
      ...state(),
      objects: [
        object({ ID: 1, CosmicObjectModelID: 10, X: 0, Y: 0, Rotation: 0 }),
        object({ ID: 2, Title: "Target", CosmicObjectModelID: 11, X: 0, Y: 55, OwnerName: "Pilot2" }),
      ],
    });

    dispose = render(() => <GameUi state={infoState} />, root);

    expect(root.querySelector(".information-panel")).not.toBeNull();
    expect(Array.from(root.querySelector(".information-panel")?.classList ?? [])).toContain("hud-panel--right-middle");
    expect(Array.from(root.querySelectorAll(".information-panel__label")).map((row) => row.textContent)).toEqual(["Название", "Модель", "Владелец"]);
    expect(Array.from(root.querySelectorAll(".information-panel__value")).map((row) => row.textContent)).toEqual(["Target", "Цель", "Pilot2"]);
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
          { chatId: 1, title: "Server", communityTypeAcronym: "Server", duoChatKey: "", unreadCount: 3, messages: [] },
          {
            chatId: 2,
            title: "Pilot2",
            communityTypeAcronym: "Duo",
            duoChatKey: "1:2",
            unreadCount: 0,
            messages: [
              { id: 10, chatId: 2, senderCharacterId: 1, senderNickname: "Pilot1", messageTypeAcronym: "FromCharacter", text: "old", color: "D8F3FF", sentTime: "" },
              { id: 11, chatId: 2, senderCharacterId: 2, senderNickname: "Pilot2", messageTypeAcronym: "FromCharacter", text: "new", color: "E8FFD8", sentTime: "" },
            ],
          },
        ],
      },
      chatInputText: "draft",
      chatCursorIndex: 2,
      chatInputFocused: true,
      chatError: "Адресат не найден",
      chatErrorSeq: 1,
      chatContextMenu: { chatId: 2, communityTypeAcronym: "Duo", x: 120, y: 220 },
      gameCursor: { visible: true, x: 320, y: 240 },
      chatScroll: { visible: true, thumbTopPercent: 25, thumbHeightPercent: 60, contentOffsetPx: 42, dragging: true },
    });

    dispose = render(() => <GameUi state={chatState} />, root);

    expect(root.querySelector(".chat-panel")).not.toBeNull();
    expect(Array.from(root.querySelectorAll(".chat-tab")).map((tab) => tab.textContent)).toEqual(["SServer3", "Pilot2"]);
    expect(root.querySelector(".chat-tab.is-selected")?.textContent).toBe("Pilot2");
    expect(root.querySelector("#chat-tab-2 .ui-kit-tab__marker")).toBeNull();
    expect(root.querySelector(".chat-tab .ui-kit-tab__badge")?.textContent).toBe("3");
    expect(Array.from(root.querySelector(".chat-panel")?.children ?? []).map((element) => element.className.replace(/\s+/g, " ").trim())).toEqual([
      "chat-messages",
      "chat-error",
      "ui-kit-control ui-kit-text-input chat-input is-focused",
      "ui-kit-control ui-kit-tabs chat-tabs",
      "ui-kit-control chat-context-menu",
    ]);
    expect(Array.from(root.querySelectorAll(".chat-message")).map((message) => message.textContent)).toEqual(["Pilot1: old", "Pilot2: new"]);
    expect(Array.from(root.querySelectorAll(".chat-message__separator")).map((separator) => separator.textContent)).toEqual([": ", ": "]);
    expect(Array.from(root.querySelectorAll(".chat-message__text")).map((message) => message.textContent)).toEqual(["old", "new"]);
    expect(root.querySelector(".ui-kit-text-input__text")?.textContent).toBe("draft");
    expect(root.querySelector<HTMLElement>(".ui-kit-text-input__caret")?.style.left).toBe("0px");
    expect(root.querySelector(".chat-error")?.textContent).toBe("Адресат не найден");
    expect(root.querySelector(".chat-context-menu__item")?.textContent).toBe("Закрыть");
    expect(root.querySelector(".chat-scrollbar .ui-kit-scrollbar__thumb")).not.toBeNull();
    expect(Array.from(root.querySelector(".chat-scrollbar")?.classList ?? [])).toContain("is-dragging");
    expect(root.querySelector<HTMLElement>(".chat-messages__content")?.style.transform).toBe("translateY(42px)");
    expect(document.body.querySelector(".game-cursor")).not.toBeNull();
  });

  // Проверяет, что значок чата рисуется только для общих каналов и одиночной вкладки, но не для прямого диалога.
  it("renders chat tab markers only for non-duo community types", () => {
    const root = document.createElement("div");
    document.body.append(root);
    const chatState = (): GameUiState => ({
      ...state(),
      chatState: {
        type: "chatState",
        selectedChatId: 1,
        tabs: [
          { chatId: 1, title: "Server", communityTypeAcronym: "Server", duoChatKey: "", unreadCount: 0, messages: [] },
          { chatId: 2, title: "Clan", communityTypeAcronym: "Clan", duoChatKey: "", unreadCount: 0, messages: [] },
          { chatId: 3, title: "Alliance", communityTypeAcronym: "Alliance", duoChatKey: "", unreadCount: 0, messages: [] },
          { chatId: 4, title: "Solo", communityTypeAcronym: "Solo", duoChatKey: "", unreadCount: 0, messages: [] },
          { chatId: 5, title: "Pilot2", communityTypeAcronym: "Duo", duoChatKey: "1:2", unreadCount: 0, messages: [] },
        ],
      },
    });

    dispose = render(() => <GameUi state={chatState} />, root);

    expect(Array.from(root.querySelectorAll(".ui-kit-tab__marker")).map((marker) => marker.textContent)).toEqual(["S", "C", "A", "S"]);
    expect(root.querySelector("#chat-tab-5 .ui-kit-tab__marker")).toBeNull();
    expect(Array.from(root.querySelector("#chat-tab-1")?.classList ?? [])).toContain("ui-kit-tab--with-marker");
    expect(Array.from(root.querySelector("#chat-tab-5")?.classList ?? [])).not.toContain("ui-kit-tab--with-marker");
  });

  // Проверяет, что игровой указатель не попадает в слой HUD, который ниже портальных меню.
  it("portals game cursor above body-level overlays", () => {
    const root = document.createElement("div");
    root.id = "ui-root";
    document.body.append(root);

    dispose = render(() => <GameUi state={() => ({ ...state(), gameCursor: { visible: true, x: 320, y: 240 } })} />, root);

    const cursor = document.body.querySelector<HTMLElement>(".game-cursor");
    expect(cursor?.closest("#ui-root")).toBeNull();
    expect(cursor ? document.body.contains(cursor) : false).toBe(true);
  });

  // Проверяет, что отладочная витрина UI Kit показывает расширенный набор контролов.
  it("renders UI kit showcase with reusable controls", () => {
    const root = document.createElement("div");
    document.body.append(root);

    dispose = render(() => <GameUi state={() => ({ ...state(), uiKitShowcaseVisible: true, uiKitDemoState: { ...state().uiKitDemoState, buttonClicks: 2, checkboxChecked: false, tabValue: "one" } })} />, root);

    expect(root.querySelector(".ui-kit-showcase")).not.toBeNull();
    expect(root.querySelector("#ui-kit-demo-button")?.textContent).toBe("Button 2");
    expect(Array.from(root.querySelector(".ui-kit-checkbox")?.classList ?? [])).not.toContain("is-checked");
    expect(root.querySelector(".ui-kit-dropdown")).not.toBeNull();
    expect(root.querySelector(".ui-kit-tree")).not.toBeNull();
    expect(root.querySelector(".ui-kit-tab.is-selected")?.textContent).toBe("One");
    expect(root.querySelector(".ui-kit-slider")).not.toBeNull();
    expect(root.querySelector(".ui-kit-tooltip")).not.toBeNull();
  });

  // Проверяет, что модальные окна используют один каркас игрового окна.
  it("renders modal windows through the same game window shell", () => {
    const root = document.createElement("div");
    document.body.append(root);

    dispose = render(() => <GameUi state={() => ({ ...state(), uiKitShowcaseVisible: true, settingsVisible: true, controlPanelVisible: true })} />, root);

    expect(root.querySelector("#ui-kit-showcase-modal")?.parentElement?.className).toBe("game-window-layer game-window-layer--showcase");
    expect(root.querySelector("#settings-modal")?.parentElement?.className).toBe("game-window-layer game-window-layer--settings");
    expect(root.querySelector("#control-panel-modal")?.parentElement?.className).toBe("game-window-layer game-window-layer--control-panel");
    expect(Array.from(root.querySelectorAll(".game-window-layer .ui-kit-modal__close")).map((button) => button.id)).toEqual(["ui-kit-showcase-modal-close-button", "settings-modal-close-button", "control-panel-modal-close-button"]);
  });

  // Проверяет, что панель управления показывает все вкладки и наполняет только страницу объекта.
  it("renders control panel tabs and object page", () => {
    const root = document.createElement("div");
    document.body.append(root);

    dispose = render(() => <GameUi state={() => ({ ...state(), controlPanelVisible: true })} />, root);

    expect(root.querySelector("#control-panel-modal .ui-kit-modal__title")?.textContent).toBe("Панель управления");
    expect(Array.from(root.querySelectorAll(".control-panel-tabs .ui-kit-tab")).map((tab) => tab.textContent)).toEqual([
      "Объект",
      "Оборудование",
      "Инструменты пилота",
      "Схемы",
      "Чертежи",
      "Карта",
    ]);
    expect(root.querySelector(".control-panel-tabs .ui-kit-tab.is-selected")?.textContent).toBe("Объект");
    expect(root.querySelector("#control-panel-object-enabled")?.textContent).toBe("");
    expect(root.querySelector("#control-panel-object-title-input .ui-kit-text-input__text")?.textContent).toBe("Ship");
    expect(root.querySelector("#control-panel-object-title-input")?.classList.contains("ui-kit-text-input")).toBe(true);
    expect(Array.from(root.querySelectorAll(".control-panel-object-row__value--control")).map((row) => row.querySelector(".ui-kit-control")?.id)).toEqual(["control-panel-object-enabled", "control-panel-object-title-input"]);
    expect(Array.from(root.querySelectorAll(".control-panel-object-row__label")).every((row) => row.classList.contains("game-form-row-label"))).toBe(true);
    expect(Array.from(root.querySelectorAll(".control-panel-object-row__value--readonly")).map(visibleControlText)).toEqual([
      "Корабль",
      "—",
      "—",
      "1",
      "0 / 0",
      "100 / 100",
      "0",
      "0.00",
      "0.00",
      "0.00",
      "0.00",
      "0.00",
      "5 / 10",
      "50 / 100",
      "0 / 0",
    ]);
    expect(Array.from(root.querySelectorAll(".control-panel-object-row__label")).map((row) => row.textContent)).toEqual([
      "Название модели космического объекта",
      "Включен",
      "Пользовательское название объекта",
      "Никнейм аккаунта персонажа-владельца",
      "Никнейм аккаунта персонажа-создателя",
      "Масса (кг)",
      "Объём оборудования / Вместимость (м³)",
      "Броня / Максимум брони",
      "Сложность",
      "Максимальная скорость (м/с)",
      "Максимальная угловая скорость (рад/с)",
      "Продольная сила тяги (максимальная) (Н)",
      "Поперечная сила тяги (максимальная) (Н)",
      "Крутящий момент (максимальный) (Н·м)",
      "Потребляемая мощность / Вырабатываемая мощность (Вт)",
      "Запас топлива / Максимальный запас топлива",
      "Занято на складе / Объём склада (м³)",
    ]);
    expect(Array.from(root.querySelectorAll(".control-panel-object-row__value")).map(visibleControlText)).toEqual([
      "Корабль",
      "",
      "Ship",
      "—",
      "—",
      "1",
      "0 / 0",
      "100 / 100",
      "0",
      "0.00",
      "0.00",
      "0.00",
      "0.00",
      "0.00",
      "5 / 10",
      "50 / 100",
      "0 / 0",
    ]);
  });

  // Проверяет, что панель управления рисует интерактивные черновики, а не только серверное состояние объекта.
  it("renders control panel object drafts for active controls", () => {
    const root = document.createElement("div");
    document.body.append(root);

    dispose = render(() => <GameUi state={() => ({
      ...state(),
      controlPanelVisible: true,
      controlPanelObjectEnabled: false,
      controlPanelObjectTitleText: "Draft",
      controlPanelObjectTitleSelectionStart: 2,
      controlPanelObjectTitleSelectionEnd: 5,
      controlPanelObjectTitleFocused: true,
    })} />, root);

    expect(root.querySelector("#control-panel-object-enabled")?.textContent).toBe("");
    expect(Array.from(root.querySelector("#control-panel-object-enabled")?.classList ?? [])).not.toContain("is-checked");
    expect(root.querySelector("#control-panel-object-title-input .ui-kit-text-input__text")?.textContent).toBe("Draft");
    expect(root.querySelector("#control-panel-object-title-input .ui-kit-text-input__selection")?.textContent).toBe("aft");
    expect(root.querySelector("#control-panel-object-title-input .ui-kit-text-input__caret")).not.toBeNull();
  });

  // Проверяет, что вкладка оборудования заполняет под-вкладку настройки по ФЗ.
  it("renders control panel equipment setup tab from installed equipment groups", () => {
    const root = document.createElement("div");
    document.body.append(root);
    const equipmentReferenceData = {
      ...referenceData,
      ItemType: {
        MaxID: 1,
        Items: {
          "1": { ID: 1, Acronym: "Container", IsPilotInstrument: false, IsInternalUsable: true },
          "2": { ID: 2, Acronym: "Generator", IsPilotInstrument: false, IsInternalUsable: false },
        },
      },
      ItemModel: {
        MaxID: 2,
        Items: {
          "1": { ID: 1, ItemTypeID: 1, Acronym: "SimpleContainer", TitleRu: "Простой контейнер", Mass: 12, Volume: 3, ConsumingPower: 4, GeneratingPower: 0, MaxAlongForce: 15, MaxAcrossForce: 6, MaxTorque: 2, Complexity: 7 },
          "2": { ID: 2, ItemTypeID: 2, Acronym: "CompactGenerator", TitleRu: "Компактный генератор", Mass: 8, Volume: 2, ConsumingPower: 0, GeneratingPower: 20, MaxAlongForce: 0, MaxAcrossForce: 0, MaxTorque: 0, Complexity: 3 },
        },
      },
    } as unknown as ReferenceDataMessage;

    dispose = render(() => <GameUi state={() => ({
      ...state(),
      controlPanelVisible: true,
      selectedControlPanelTab: "equipment",
      selectedControlPanelEquipmentGroupId: 11,
      controlPanelEquipmentEnabledDrafts: { 11: true },
      controlPanelEquipmentEnabledCountDrafts: { 11: 1 },
      controlPanelEquipmentTitleText: "Генератор",
      controlPanelEquipmentTitleSelectionStart: 9,
      controlPanelEquipmentTitleSelectionEnd: 9,
      controlPanelEquipmentTitleFocused: false,
      controlPanelEquipmentListScroll: { visible: true, thumbTopPercent: 20, thumbHeightPercent: 55, contentOffsetPx: 17, dragging: true },
      referenceData: equipmentReferenceData,
      equipmentGroups: [
        { ID: 20, CosmicObjectID: 2, Title: "Чужое", EquipmentItemModelID: 2, Count: 1, EnabledCount: 1, Enabled: true, Active: false, LastRechargeStartTime: 0 },
        { ID: 10, CosmicObjectID: 1, Title: "Маршевые", EquipmentItemModelID: 1, Count: 4, EnabledCount: 3, Enabled: true, Active: false, LastRechargeStartTime: 0 },
        { ID: 11, CosmicObjectID: 1, Title: "Генератор", EquipmentItemModelID: 2, Count: 2, EnabledCount: 2, Enabled: false, Active: true, LastRechargeStartTime: 0 },
      ],
    })} />, root);

    expect(Array.from(root.querySelectorAll(".control-panel-equipment-tabs .ui-kit-tab")).map((tab) => tab.textContent)).toEqual(["Настройка", "Использование"]);
    expect(root.querySelector(".control-panel-equipment-tabs .ui-kit-tab.is-selected")?.textContent).toBe("Настройка");
    expect(Array.from(root.querySelectorAll(".control-panel-equipment-list .ui-kit-list__item")).map((item) => item.textContent)).toEqual(["Простой контейнер", "Компактный генератор"]);
    expect(root.querySelector(".control-panel-equipment-list .ui-kit-list__item.is-selected")?.textContent).toBe("Компактный генератор");
    expect(root.querySelector<HTMLElement>(".control-panel-equipment-list .ui-kit-list__content")?.style.transform).toBe("translateY(-17px)");
    expect(Array.from(root.querySelector("#control-panel-equipment-list-scrollbar")?.classList ?? [])).toContain("is-dragging");
    expect(root.querySelector("#control-panel-equipment-enabled")?.textContent).toBe("");
    expect(Array.from(root.querySelector("#control-panel-equipment-enabled")?.classList ?? [])).toContain("is-checked");
    expect(root.querySelector("#control-panel-equipment-enabled-count")).toBeNull();
    expect(root.querySelector<HTMLElement>("#control-panel-equipment-enabled-slider .ui-kit-slider__fill")?.style.width).toBe("50%");
    expect(root.querySelector("#control-panel-equipment-enabled-slider .ui-kit-slider__label")?.textContent).toBe("1 / 2");
    expect(root.querySelector("#control-panel-equipment-title-input")?.classList.contains("ui-kit-text-input")).toBe(true);
    expect(root.querySelector("#control-panel-equipment-title-input .ui-kit-text-input__text")?.textContent).toBe("Генератор");
    expect(root.querySelector("#control-panel-equipment-usage-button")?.textContent).toBe("Использовать");
    expect(root.querySelector("#control-panel-equipment-usage-button")?.classList.contains("is-disabled")).toBe(true);
    expect(root.querySelector(".control-panel-equipment-action #control-panel-equipment-usage-button")?.textContent).toBe("Использовать");
    expect(root.querySelector(".control-panel-equipment-info > .control-panel-equipment-action #control-panel-equipment-usage-button")?.textContent).toBe("Использовать");
    expect(root.querySelector(".control-panel-equipment-layout > .control-panel-equipment-action")).toBeNull();
    expect(Array.from(root.querySelectorAll(".control-panel-equipment-info .control-panel-object-row .control-panel-object-row__label")).map((row) => row.textContent)).toEqual([
      "Название модели оборудования",
      "Включено",
      "Количество включенных единиц",
      "Активно",
      "Масса (кг)",
      "Объём (м³)",
      "Потребляемая мощность (Вт)",
      "Вырабатываемая мощность (Вт)",
      "Продольная сила тяги (Н)",
      "Поперечная сила тяги (Н)",
      "Крутящий момент (Н·м)",
      "Сложность",
    ]);
    expect(Array.from(root.querySelectorAll(".control-panel-equipment-info .control-panel-object-row .control-panel-object-row__value")).map(visibleControlText)).toEqual([
      "Компактный генератор",
      "",
      "1 / 2",
      "Да",
      "8",
      "2",
      "0",
      "20",
      "0",
      "0",
      "0",
      "3",
    ]);
  });

  // Проверяет, что подвкладка использования показывает контейнер слева и выбранный контейнер справа по ФЗ.
  it("renders control panel equipment usage container UI", () => {
    const root = document.createElement("div");
    document.body.append(root);
    const equipmentReferenceData = {
      ...referenceData,
      ItemType: {
        MaxID: 2,
        Items: {
          "1": { ID: 1, Acronym: "Container", IsPilotInstrument: false, IsInternalUsable: true },
          "2": { ID: 2, Acronym: "Generator", IsPilotInstrument: false, IsInternalUsable: true },
        },
      },
      ItemModel: {
        MaxID: 4,
        Items: {
          "1": { ID: 1, ItemTypeID: 1, Acronym: "CargoContainer", TitleRu: "Грузовой контейнер" },
          "2": { ID: 2, ItemTypeID: 2, Acronym: "Generator", TitleRu: "Генератор" },
          "3": { ID: 3, ItemTypeID: 1, Acronym: "ReserveContainer", TitleRu: "Резервный контейнер" },
          "4": { ID: 4, ItemTypeID: 2, Acronym: "Ore", TitleRu: "Руда" },
        },
      },
    } as unknown as ReferenceDataMessage;

    dispose = render(() => <GameUi state={() => ({
      ...state(),
      controlPanelVisible: true,
      selectedControlPanelTab: "equipment",
      selectedControlPanelEquipmentTab: "usage",
      gameCursor: { visible: true, x: 240, y: 180 },
      hoveredGameUiControlId: "control-panel-constructor-schema-list-1",
      selectedControlPanelUsageLeftContainerGroupId: 10,
      selectedControlPanelUsageRightEquipmentGroupId: 12,
      openControlPanelUsageSelect: null,
      selectedControlPanelUsageLeftItemGroupIds: [1],
      selectedControlPanelUsageRightItemGroupIds: [2],
      referenceData: equipmentReferenceData,
      equipmentGroups: [
        { ID: 10, CosmicObjectID: 1, Title: "Левый", EquipmentItemModelID: 1, Count: 1, EnabledCount: 1, Enabled: true, Active: false, LastRechargeStartTime: 0 },
        { ID: 11, CosmicObjectID: 1, Title: "Энергия", EquipmentItemModelID: 2, Count: 1, EnabledCount: 1, Enabled: true, Active: false, LastRechargeStartTime: 0 },
        { ID: 12, CosmicObjectID: 1, Title: "Правый", EquipmentItemModelID: 3, Count: 1, EnabledCount: 1, Enabled: true, Active: false, LastRechargeStartTime: 0 },
      ],
      itemGroups: [
        { ID: 1, ContainerEquipmentGroupID: 10, ContentItemModelID: 4, Count: 7 },
        { ID: 2, ContainerEquipmentGroupID: 12, ContentItemModelID: 4, Count: 3 },
      ],
    })} />, root);

    expect(root.querySelector(".control-panel-equipment-tabs .ui-kit-tab.is-selected")?.textContent).toBe("Использование");
    expect(root.querySelector("#control-panel-usage-left-container-select .ui-kit-dropdown__value")?.textContent).toBe("Левый");
    expect(root.querySelector("#control-panel-usage-right-equipment-select .ui-kit-dropdown__value")?.textContent).toBe("Правый");
    expect(root.querySelector(".control-panel-container-content__title")).toBeNull();
    expect(Array.from(root.querySelectorAll(".control-panel-container-content .ui-kit-list__item")).map((row) => row.textContent)).toEqual(["Руда7", "Руда3"]);
    expect(Array.from(root.querySelectorAll(".control-panel-container-content .ui-kit-list__item.is-selected")).map((row) => row.textContent)).toEqual(["Руда7", "Руда3"]);
    expect(root.querySelector("#control-panel-container-transfer-to-right")?.getAttribute("aria-label")).toBe("Переместить выбранные предметы в правый контейнер");
    expect(root.querySelector("#control-panel-container-transfer-to-left")?.getAttribute("aria-label")).toBe("Переместить выбранные предметы в левый контейнер");
  });

  // Проверяет, что выбранный конструктор показывает схемы, чертежи и отдельный контейнер материалов.
  // Проверяет, что выбор использования разделён на объект кластера и группу оборудования выбранного объекта.
  it("renders cluster object selects before usage equipment selects", () => {
    const root = document.createElement("div");
    document.body.append(root);
    const equipmentReferenceData = {
      ...referenceData,
      ItemType: {
        MaxID: 2,
        Items: {
          "1": { ID: 1, Acronym: "Container", IsPilotInstrument: false, IsInternalUsable: true },
          "2": { ID: 2, Acronym: "FuelTank", IsPilotInstrument: false, IsInternalUsable: true },
        },
      },
      ItemModel: {
        MaxID: 2,
        Items: {
          "1": { ID: 1, ItemTypeID: 1, Acronym: "CargoContainer", TitleRu: "Грузовой контейнер" },
          "2": { ID: 2, ItemTypeID: 2, Acronym: "FuelTank", TitleRu: "Топливный бак" },
        },
      },
    } as unknown as ReferenceDataMessage;

    dispose = render(() => <GameUi state={() => ({
      ...state(),
      controlPanelVisible: true,
      selectedControlPanelTab: "equipment",
      selectedControlPanelEquipmentTab: "usage",
      selfObject: object({ ID: 1, Title: "База", OwnerCharacterID: 7, ClusterMainCosmicObjectID: 1 }),
      objects: [
        object({ ID: 1, Title: "База", OwnerCharacterID: 7, ClusterMainCosmicObjectID: 1 }),
        object({ ID: 2, Title: "Буксир", OwnerCharacterID: 7, ClusterMainCosmicObjectID: 1 }),
        object({ ID: 3, Title: "Чужой", OwnerCharacterID: 8, ClusterMainCosmicObjectID: 1 }),
      ],
      selectedControlPanelUsageLeftObjectId: 2,
      selectedControlPanelUsageLeftContainerGroupId: 20,
      selectedControlPanelUsageRightObjectId: 1,
      selectedControlPanelUsageRightEquipmentGroupId: 11,
      selectedControlPanelConstructorMaterialObjectId: 2,
      selectedControlPanelConstructorMaterialContainerGroupId: 20,
      openControlPanelUsageSelect: "leftObject",
      referenceData: equipmentReferenceData,
      equipmentGroups: [
        { ID: 10, CosmicObjectID: 1, Title: "Склад базы", EquipmentItemModelID: 1, Count: 1, EnabledCount: 1, Enabled: true, Active: false, LastRechargeStartTime: 0 },
        { ID: 11, CosmicObjectID: 1, Title: "Бак базы", EquipmentItemModelID: 2, Count: 1, EnabledCount: 1, Enabled: true, Active: false, LastRechargeStartTime: 0 },
        { ID: 20, CosmicObjectID: 2, Title: "Склад буксира", EquipmentItemModelID: 1, Count: 1, EnabledCount: 1, Enabled: true, Active: false, LastRechargeStartTime: 0 },
        { ID: 30, CosmicObjectID: 3, Title: "Чужой склад", EquipmentItemModelID: 1, Count: 1, EnabledCount: 1, Enabled: true, Active: false, LastRechargeStartTime: 0 },
      ],
    })} />, root);

    expect(root.querySelector("#control-panel-usage-left-object-select .ui-kit-dropdown__value")?.textContent).toBe("Буксир");
    expect(root.querySelector("#control-panel-usage-left-container-select .ui-kit-dropdown__value")?.textContent).toBe("Склад буксира");
    expect(root.querySelector("#control-panel-usage-right-object-select .ui-kit-dropdown__value")?.textContent).toBe("База");
    expect(root.querySelector("#control-panel-usage-right-equipment-select .ui-kit-dropdown__value")?.textContent).toBe("Бак базы");
    expect(document.querySelector("#control-panel-usage-left-object-select .ui-kit-dropdown__menu")?.textContent ?? "").not.toContain("Чужой");
  });

  it("renders control panel equipment usage constructor UI", async () => {
    const root = document.createElement("div");
    document.body.append(root);
    const equipmentReferenceData = {
      ...referenceData,
      ItemType: {
        MaxID: 3,
        Items: {
          "1": { ID: 1, Acronym: "Container", IsPilotInstrument: false, IsInternalUsable: true },
          "2": { ID: 2, Acronym: "Constructor", IsPilotInstrument: true, IsInternalUsable: true },
          "3": { ID: 3, Acronym: "Resource", IsPilotInstrument: false, IsInternalUsable: false },
        },
      },
      ItemModel: {
        MaxID: 4,
        Items: {
          "1": { ID: 1, ItemTypeID: 1, Acronym: "CargoContainer", TitleRu: "Контейнер" },
          "2": { ID: 2, ItemTypeID: 2, Acronym: "Constructor", TitleRu: "Конструктор", ConsumingPower: 18, Efficiency: 1 },
          "3": { ID: 3, ItemTypeID: 3, Acronym: "Ferrogel", TitleRu: "Феррогель" },
          "4": { ID: 4, ItemTypeID: 3, Acronym: "Plate", TitleRu: "Пластина" },
        },
      },
      CosmicObjectModel: {
        ...referenceData.CosmicObjectModel,
        Items: {
          ...referenceData.CosmicObjectModel.Items,
          "12": { ID: 12, CosmicObjectTypeID: 1, TitleRu: "Катер", TextureWidth: 40, TextureHeight: 40, TextureBodyOriginX: 20, TextureBodyOriginY: 20, TextureScale: 1, BodyWidth: 20, BodyLength: 20 },
        },
      },
      Schema: {
        MaxID: 1,
        Items: {
          "1": { ID: 1, TitleRu: "Схема: Пластина", TitleEn: "Schema: Plate", ItemModelID: 4, Count: 2, ProductionEnergy: 180 },
        },
      },
      SchemaComponent: {
        MaxID: 1,
        Items: {
          "1": { ID: 1, SchemaID: 1, ComponentItemModelID: 3, Count: 5 },
        },
      },
      Blueprint: {
        MaxID: 1,
        Items: {
          "1": { ID: 1, TitleRu: "Чертёж: Катер", TitleEn: "Blueprint: Boat", CosmicObjectModelID: 12, ProductionEnergy: 90 },
        },
      },
      BlueprintComponent: {
        MaxID: 1,
        Items: {
          "1": { ID: 1, BlueprintID: 1, ComponentItemModelID: 4, Count: 6 },
        },
      },
      TaskType: {
        MaxID: 2,
        Items: {
          "1": { ID: 1, TitleRu: "Производство предметов", TitleEn: "Item production", Acronym: "ItemProduction" },
          "2": { ID: 2, TitleRu: "Производство объектов", TitleEn: "Object production", Acronym: "ObjectProduction" },
        },
      },
      Implementer: {
        MaxID: 2,
        Items: {
          "1": { ID: 1, TaskTypeID: 1, ImplementerEquipmentItemTypeID: 2, WorkPart: 1 },
          "2": { ID: 2, TaskTypeID: 2, ImplementerEquipmentItemTypeID: 2, WorkPart: 1 },
        },
      },
    } as unknown as ReferenceDataMessage;

    dispose = render(() => <GameUi state={() => ({
      ...state(),
      controlPanelVisible: true,
      selectedControlPanelTab: "equipment",
      selectedControlPanelEquipmentTab: "usage",
      gameCursor: { visible: true, x: 240, y: 180 },
      hoveredGameUiControlId: "control-panel-constructor-schema-list-1",
      selectedControlPanelUsageLeftContainerGroupId: 13,
      selectedControlPanelUsageRightEquipmentGroupId: 11,
      selectedControlPanelConstructorMaterialContainerGroupId: 12,
      selectedControlPanelConstructorProductContainerGroupId: 10,
      selectedControlPanelConstructorSchemaId: 1,
      selectedControlPanelConstructorMainJobId: 1,
      referenceData: equipmentReferenceData,
      equipmentGroups: [
        { ID: 10, CosmicObjectID: 1, Title: "Продукция", EquipmentItemModelID: 1, Count: 1, EnabledCount: 1, Enabled: true, Active: false, LastRechargeStartTime: 0 },
        { ID: 11, CosmicObjectID: 1, Title: "Сборщик", EquipmentItemModelID: 2, Count: 1, EnabledCount: 1, Enabled: true, Active: false, LastRechargeStartTime: 0 },
        { ID: 12, CosmicObjectID: 1, Title: "Материалы", EquipmentItemModelID: 1, Count: 1, EnabledCount: 1, Enabled: true, Active: false, LastRechargeStartTime: 0 },
        { ID: 13, CosmicObjectID: 1, Title: "Обычный левый", EquipmentItemModelID: 1, Count: 1, EnabledCount: 1, Enabled: true, Active: false, LastRechargeStartTime: 0 },
      ],
      itemGroups: [
        { ID: 1, ContainerEquipmentGroupID: 12, ContentItemModelID: 3, Count: 15 },
        { ID: 2, ContainerEquipmentGroupID: 10, ContentItemModelID: 4, Count: 2 },
      ],
      constructorProductionJobs: [
        { id: 1, constructorEquipmentGroupId: 11, queueType: "main", schemaId: 1, blueprintId: 0, productItemModelId: 4, productCosmicObjectModelId: 0, productCount: 2, remainingCount: 6, totalCount: 8, remainingTime: 12, totalTime: 30, running: true, parentJobId: 0 },
        { id: 2, constructorEquipmentGroupId: 11, queueType: "auxiliary", schemaId: 1, blueprintId: 0, productItemModelId: 4, productCosmicObjectModelId: 0, productCount: 2, remainingCount: 4, totalCount: 8, remainingTime: 20, totalTime: 30, running: false, parentJobId: 1 },
        { id: 3, constructorEquipmentGroupId: 11, queueType: "auxiliary", schemaId: 1, blueprintId: 0, productItemModelId: 4, productCosmicObjectModelId: 0, productCount: 2, remainingCount: 2, totalCount: 4, remainingTime: 10, totalTime: 30, running: true, parentJobId: 1 },
        { id: 4, constructorEquipmentGroupId: 11, queueType: "auxiliary", schemaId: 1, blueprintId: 0, productItemModelId: 4, productCosmicObjectModelId: 0, productCount: 2, remainingCount: 3, totalCount: 3, remainingTime: 30, totalTime: 30, running: false, parentJobId: 99 },
      ],
    })} />, root);

    expect(root.querySelector(".control-panel-equipment-usage--constructor")).not.toBeNull();
    expect(root.querySelector("#control-panel-constructor-product-select .ui-kit-dropdown__value")?.textContent).toBe("Продукция");
    expect(root.querySelector("#control-panel-usage-left-container-select")).toBeNull();
    expect(root.querySelector("#control-panel-constructor-material-select .ui-kit-dropdown__value")?.textContent).toBe("Материалы");
    expect(root.querySelector("#control-panel-usage-left-container-content-2")?.textContent).toBe("Пластина2");
    expect(root.querySelector(".control-panel-constructor-storage")?.closest(".control-panel-equipment-usage__panel--left")).not.toBeNull();
    expect(root.querySelector(".control-panel-constructor-recipes")?.closest(".control-panel-equipment-usage__panel--right")).not.toBeNull();
    expect(root.querySelector(".control-panel-constructor-queues")?.closest(".control-panel-equipment-usage__panel--right")).not.toBeNull();
    expect(root.querySelector(".control-panel-constructor-usage")?.children[0]?.classList.contains("control-panel-controller-work")).toBe(true);
    expect(root.querySelector(".control-panel-constructor-usage")?.children[1]?.classList.contains("control-panel-equipment-right-stack")).toBe(true);
    expect(root.querySelector(".control-panel-constructor-recipes")?.closest(".control-panel-equipment-right-stack")).not.toBeNull();
    expect(root.querySelector("#control-panel-usage-right-equipment-select")?.closest(".control-panel-controller-work")).not.toBeNull();
    expect(root.querySelector("#control-panel-constructor-main-queue-1")?.textContent).toBe("Пластина2 / 8");
    expect(root.querySelector("#control-panel-constructor-required-queue-2")?.textContent).toBe("Пластина6 / 12");
    expect(root.querySelector("#control-panel-constructor-required-queue-4")?.textContent).toBe("Пластина0 / 3");
    expect(root.querySelector("#control-panel-constructor-main-queue-1")?.classList.contains("is-selected")).toBe(true);
    expect(root.querySelector("#control-panel-constructor-required-queue-2")?.getAttribute("style")).not.toContain("--constructor-queue-total-progress");
    expect(root.querySelector("#control-panel-constructor-skip-next")?.textContent).toBe("Не делать следующие");
    expect(root.querySelector("#control-panel-constructor-skip-all-next")?.textContent).toBe("Не делать все следующие");
    expect(root.querySelector("#control-panel-constructor-cancel")?.textContent).toBe("Отменить");
    expect(root.querySelector("#control-panel-constructor-cancel-all")?.textContent).toBe("Отменить все");
    expect(root.querySelector("#control-panel-constructor-schema-list-1")?.textContent).toBe("Пластина");
    expect(root.querySelector("#control-panel-constructor-schema-list-1")?.getAttribute("title")).toBe("2 шт, 10 с, Феррогель: 5");
    expect(root.querySelector("#control-panel-constructor-make-button")?.textContent).toBe("Изготовить");
    expect(root.querySelector("#control-panel-usage-right-container-content-1")?.textContent).toBe("Феррогель15");
    await Promise.resolve();
    expect(document.querySelector(".control-panel-constructor-recipe-tooltip")?.textContent).toBe("ПластинаФеррогель: 5Получается: 2Время: 10 с");
    expect(document.querySelector(".control-panel-constructor-recipe-tooltip__component")?.textContent).toBe("Феррогель: 5");
  });

  // Проверяет, что выбранная строка очереди деконструкции получает общий класс выделения списка.
  it("highlights selected deconstructor queue row", () => {
    const root = document.createElement("div");
    document.body.append(root);
    const equipmentReferenceData = {
      ...referenceData,
      ItemType: {
        MaxID: 2,
        Items: {
          "1": { ID: 1, Acronym: "Container", IsPilotInstrument: false, IsInternalUsable: true },
          "2": { ID: 2, Acronym: "Deconstructor", IsPilotInstrument: false, IsInternalUsable: true },
        },
      },
      ItemModel: {
        MaxID: 3,
        Items: {
          "1": { ID: 1, TitleRu: "Контейнер", TitleEn: "Container", Acronym: "Container", ItemTypeID: 1 },
          "2": { ID: 2, TitleRu: "Деконструктор", TitleEn: "Deconstructor", Acronym: "Deconstructor", ItemTypeID: 2 },
          "3": { ID: 3, TitleRu: "Пластина", TitleEn: "Plate", Acronym: "Plate", ItemTypeID: 1 },
        },
      },
      TaskType: {
        MaxID: 3,
        Items: {
          "3": { ID: 3, TitleRu: "Деконструкция предметов", TitleEn: "Item deconstruction", Acronym: "ItemDeconstruction" },
        },
      },
    } as unknown as ReferenceDataMessage;

    dispose = render(() => <GameUi state={() => ({
      ...state(),
      controlPanelVisible: true,
      selectedControlPanelTab: "equipment",
      selectedControlPanelEquipmentTab: "usage",
      selectedControlPanelUsageRightEquipmentGroupId: 11,
      selectedControlPanelConstructorMaterialContainerGroupId: 10,
      selectedControlPanelUsageLeftContainerGroupId: 10,
      selectedControlPanelConstructorMainJobId: 31,
      referenceData: equipmentReferenceData,
      equipmentGroups: [
        { ID: 10, CosmicObjectID: 1, Title: "Контейнер", EquipmentItemModelID: 1, Count: 1, EnabledCount: 1, Enabled: true, Active: false, LastRechargeStartTime: 0 },
        { ID: 11, CosmicObjectID: 1, Title: "Деконструктор", EquipmentItemModelID: 2, Count: 1, EnabledCount: 1, Enabled: true, Active: false, LastRechargeStartTime: 0 },
      ],
      itemGroups: [
        { ID: 21, ContainerEquipmentGroupID: 10, ContentItemModelID: 3, Count: 4 },
      ],
      tasks: [
        { ID: 31, ControllerEquipmentGroupID: 11, ParentTaskID: 0, TaskTypeID: 3, RemainingEnergy: 5, TotalEnergy: 10, SchemaID: 0, BlueprintID: 0, LeftToRightDirection: true, BatchCount: 1 },
      ],
      taskItemGroups: [
        { ID: 41, TaskID: 31, ItemModelID: 3, Count: 4, IsStored: true },
      ],
    })} />, root);

    expect(root.querySelector("#control-panel-deconstructor-main-queue-31")?.classList.contains("is-selected")).toBe(true);
  });

  // Проверяет, что выбранный чертёж объекта включает кнопку запуска изготовления.
  it("enables constructor make button for selected object blueprint", () => {
    const root = document.createElement("div");
    document.body.append(root);
    const equipmentReferenceData = {
      ...referenceData,
      ItemType: {
        MaxID: 2,
        Items: {
          "1": { ID: 1, Acronym: "Container", IsPilotInstrument: false, IsInternalUsable: true },
          "2": { ID: 2, Acronym: "Constructor", IsPilotInstrument: false, IsInternalUsable: true },
        },
      },
      ItemModel: {
        MaxID: 2,
        Items: {
          "1": { ID: 1, TitleRu: "Контейнер", TitleEn: "Container", Acronym: "Container", ItemTypeID: 1 },
          "2": { ID: 2, TitleRu: "Конструктор", TitleEn: "Constructor", Acronym: "Constructor", ItemTypeID: 2 },
        },
      },
      Blueprint: {
        MaxID: 1,
        Items: {
          "1": { ID: 1, TitleRu: "Чертёж: Катер", TitleEn: "Blueprint: Boat", CosmicObjectModelID: 12, ProductionEnergy: 90 },
        },
      },
      BlueprintComponent: {
        MaxID: 1,
        Items: {
          "1": { ID: 1, BlueprintID: 1, ComponentItemModelID: 4, Count: 6 },
        },
      },
    } as unknown as ReferenceDataMessage;

    const dispose = render(() => <GameUi state={() => ({
      ...state(),
      controlPanelVisible: true,
      selectedControlPanelTab: "equipment",
      selectedControlPanelEquipmentTab: "usage",
      selectedControlPanelUsageRightEquipmentGroupId: 11,
      selectedControlPanelConstructorMaterialContainerGroupId: 12,
      selectedControlPanelConstructorTab: "objects",
      selectedControlPanelConstructorBlueprintId: 1,
      referenceData: equipmentReferenceData,
      equipmentGroups: [
        { ID: 11, CosmicObjectID: 1, Title: "Сборщик", EquipmentItemModelID: 2, Count: 1, EnabledCount: 1, Enabled: true, Active: false, LastRechargeStartTime: 0 },
        { ID: 12, CosmicObjectID: 1, Title: "Материалы", EquipmentItemModelID: 1, Count: 1, EnabledCount: 1, Enabled: true, Active: false, LastRechargeStartTime: 0 },
      ],
    })} />, root);

    expect(root.querySelector("#control-panel-constructor-make-button")).not.toBeNull();
    expect(root.querySelector("#control-panel-constructor-make-button")?.classList.contains("is-disabled")).toBe(false);

    dispose();
    root.remove();
  });

  // Проверяет, что выбранный топливный бак в правой панели использования показывает шкалу общего топлива объекта.
  it("renders control panel equipment usage fuel tank UI", () => {
    const root = document.createElement("div");
    document.body.append(root);
    const equipmentReferenceData = {
      ...referenceData,
      ItemType: {
        MaxID: 3,
        Items: {
          "1": { ID: 1, Acronym: "Container", IsPilotInstrument: false, IsInternalUsable: true },
          "2": { ID: 2, Acronym: "FuelTank", IsPilotInstrument: false, IsInternalUsable: true },
          "3": { ID: 3, Acronym: "Resource", IsPilotInstrument: false, IsInternalUsable: false },
        },
      },
      ItemModel: {
        MaxID: 3,
        Items: {
          "1": { ID: 1, ItemTypeID: 1, Acronym: "CargoContainer", TitleRu: "Грузовой контейнер" },
          "2": { ID: 2, ItemTypeID: 2, Acronym: "FuelTank", TitleRu: "Топливный бак" },
          "3": { ID: 3, ItemTypeID: 3, Acronym: "Melit", TitleRu: "Мелит" },
        },
      },
    } as unknown as ReferenceDataMessage;

    dispose = render(() => <GameUi state={() => ({
      ...state(),
      selfObject: { ...object(), Fuel: 40, MaxFuel: 100 },
      controlPanelVisible: true,
      selectedControlPanelTab: "equipment",
      selectedControlPanelEquipmentTab: "usage",
      selectedControlPanelUsageLeftContainerGroupId: 10,
      selectedControlPanelUsageRightEquipmentGroupId: 11,
      selectedControlPanelUsageLeftItemGroupIds: [1],
      controlPanelFuelDrainDialogOpen: true,
      controlPanelFuelDrainAmount: 12,
      controlPanelFuelDrainAmountText: "12",
      controlPanelFuelDrainAmountSelectionStart: 2,
      controlPanelFuelDrainAmountSelectionEnd: 2,
      referenceData: equipmentReferenceData,
      equipmentGroups: [
        { ID: 10, CosmicObjectID: 1, Title: "Склад", EquipmentItemModelID: 1, Count: 1, EnabledCount: 1, Enabled: true, Active: false, LastRechargeStartTime: 0 },
        { ID: 11, CosmicObjectID: 1, Title: "Основной бак", EquipmentItemModelID: 2, Count: 1, EnabledCount: 1, Enabled: true, Active: false, LastRechargeStartTime: 0 },
      ],
      itemGroups: [
        { ID: 1, ContainerEquipmentGroupID: 10, ContentItemModelID: 3, Count: 70 },
      ],
    })} />, root);

    expect(root.querySelector("#control-panel-usage-right-equipment-select .ui-kit-dropdown__value")?.textContent).toBe("Основной бак");
    expect(root.querySelector(".control-panel-fuel-tank__label")?.textContent).toBe("40 / 100");
    expect((root.querySelector(".control-panel-fuel-tank__fill") as HTMLElement | null)?.style.height).toBe("40%");
    expect(root.querySelector("#control-panel-fuel-transfer-to-tank")?.getAttribute("aria-label")).toBe("Переместить выбранное топливо в топливный бак");
    expect(root.querySelector("#control-panel-fuel-drain-open")?.getAttribute("aria-label")).toBe("Слить топливо в левый контейнер");
    expect(root.querySelector("#control-panel-fuel-drain-dialog .control-panel-fuel-drain-dialog__title")?.textContent).toBe("Слив топлива");
    expect(root.querySelector("#control-panel-fuel-drain-amount-input .ui-kit-text-input__text")?.textContent).toBe("12");
    expect(root.querySelector("#control-panel-fuel-drain-amount-slider .ui-kit-slider__fill")).not.toBeNull();
    expect(root.querySelector("#control-panel-fuel-drain-ok")?.textContent).toBe("ОК");
  });

  // Проверяет, что топливный бак показывает окно выбора количества для залива топлива.
  it("renders fuel fill amount dialog for selected fuel tank", () => {
    const root = document.createElement("div");
    document.body.append(root);
    const equipmentReferenceData = {
      ...referenceData,
      ItemType: {
        MaxID: 3,
        Items: {
          "1": { ID: 1, Acronym: "Container", IsPilotInstrument: false, IsInternalUsable: true },
          "2": { ID: 2, Acronym: "FuelTank", IsPilotInstrument: false, IsInternalUsable: true },
          "3": { ID: 3, Acronym: "Resource", IsPilotInstrument: false, IsInternalUsable: false },
        },
      },
      ItemModel: {
        MaxID: 3,
        Items: {
          "1": { ID: 1, ItemTypeID: 1, Acronym: "CargoContainer", TitleRu: "Грузовой контейнер" },
          "2": { ID: 2, ItemTypeID: 2, Acronym: "FuelTank", TitleRu: "Топливный бак" },
          "3": { ID: 3, ItemTypeID: 3, Acronym: "Melit", TitleRu: "Мелит" },
        },
      },
    } as unknown as ReferenceDataMessage;

    dispose = render(() => <GameUi state={() => ({
      ...state(),
      selfObject: { ...object(), Fuel: 40, MaxFuel: 100 },
      controlPanelVisible: true,
      selectedControlPanelTab: "equipment",
      selectedControlPanelEquipmentTab: "usage",
      selectedControlPanelUsageLeftContainerGroupId: 10,
      selectedControlPanelUsageRightEquipmentGroupId: 11,
      selectedControlPanelUsageLeftItemGroupIds: [1],
      controlPanelFuelFillDialogOpen: true,
      controlPanelFuelFillMaxAmount: 30,
      controlPanelFuelDrainAmount: 18,
      controlPanelFuelDrainAmountText: "18",
      controlPanelFuelDrainAmountSelectionStart: 2,
      controlPanelFuelDrainAmountSelectionEnd: 2,
      referenceData: equipmentReferenceData,
      equipmentGroups: [
        { ID: 10, CosmicObjectID: 1, Title: "Склад", EquipmentItemModelID: 1, Count: 1, EnabledCount: 1, Enabled: true, Active: false, LastRechargeStartTime: 0 },
        { ID: 11, CosmicObjectID: 1, Title: "Основной бак", EquipmentItemModelID: 2, Count: 1, EnabledCount: 1, Enabled: true, Active: false, LastRechargeStartTime: 0 },
      ],
      itemGroups: [
        { ID: 1, ContainerEquipmentGroupID: 10, ContentItemModelID: 3, Count: 70 },
      ],
    })} />, root);

    expect(root.querySelector("#control-panel-fuel-fill-dialog .control-panel-fuel-drain-dialog__title")?.textContent).toBe("Залив топлива");
    expect(root.querySelector("#control-panel-fuel-drain-amount-input .ui-kit-text-input__text")?.textContent).toBe("18");
    expect(root.querySelector("#control-panel-fuel-drain-amount-slider .ui-kit-slider__fill")).not.toBeNull();
    expect(root.querySelector("#control-panel-fuel-fill-ok")?.textContent).toBe("ОК");
  });

  // Проверяет, что контейнерный перенос одной строки показывает окно выбора количества.
  it("renders container transfer amount dialog", () => {
    const root = document.createElement("div");
    document.body.append(root);

    dispose = render(() => <GameUi state={() => ({
      ...state(),
      controlPanelVisible: true,
      selectedControlPanelTab: "equipment",
      selectedControlPanelEquipmentTab: "usage",
      controlPanelContainerTransferDialogOpen: true,
      controlPanelContainerTransferMaxAmount: 7,
      controlPanelFuelDrainAmount: 3,
      controlPanelFuelDrainAmountText: "3",
      controlPanelFuelDrainAmountSelectionStart: 1,
      controlPanelFuelDrainAmountSelectionEnd: 1,
    })} />, root);

    expect(root.querySelector("#control-panel-container-transfer-dialog .control-panel-fuel-drain-dialog__title")?.textContent).toBe("Перенос предметов");
    expect(root.querySelector("#control-panel-fuel-drain-amount-input .ui-kit-text-input__text")?.textContent).toBe("3");
    expect(root.querySelector("#control-panel-fuel-drain-amount-slider .ui-kit-slider__fill")).not.toBeNull();
    expect(root.querySelector("#control-panel-container-transfer-ok")?.textContent).toBe("ОК");
  });

  // Проверяет, что будущие вкладки панели управления уже переключают пустую страницу.
  it("renders empty control panel page for future tabs", () => {
    const root = document.createElement("div");
    document.body.append(root);

    dispose = render(() => <GameUi state={() => ({ ...state(), controlPanelVisible: true, selectedControlPanelTab: "pilotTools" })} />, root);

    expect(root.querySelector(".control-panel-empty-page")?.textContent).toBe("");
    expect(root.querySelector(".control-panel-object-page")).toBeNull();
  });

  // Проверяет, что частые счетчики кадра не пересоздают строки окна настроек.
  it("keeps settings rows mounted when only frame counters change", () => {
    const root = document.createElement("div");
    document.body.append(root);
    const controller = createGameUiController();
    controller.update({
      ...state(),
      settingsVisible: true,
      gameCursor: { visible: true, x: 320, y: 240 },
      fps: 180,
    });

    dispose = render(() => <GameUi state={controller.state} />, root);

    const row = root.querySelector(".settings-input-row");
    const dropdown = root.querySelector(".settings-input-row .ui-kit-dropdown");
    controller.update({
      ...controller.state(),
      gameCursor: { visible: true, x: 360, y: 260 },
      fps: 81,
    });

    expect(root.querySelector(".settings-input-row")).toBe(row);
    expect(root.querySelector(".settings-input-row .ui-kit-dropdown")).toBe(dropdown);
  });

  // Проверяет, что окно настроек показывает будущие разделы перед вкладкой ввода.
  it("renders video and audio tabs before input settings tab", () => {
    const root = document.createElement("div");
    document.body.append(root);

    dispose = render(() => <GameUi state={() => ({ ...state(), settingsVisible: true, inputSettingsScroll: { visible: true, thumbTopPercent: 0, thumbHeightPercent: 50, contentOffsetPx: 0, dragging: false } })} />, root);

    expect(Array.from(root.querySelectorAll(".settings-tabs .ui-kit-tab")).map((tab) => tab.textContent)).toEqual(["Видео", "Аудио", "Ввод"]);
    expect(root.querySelector("#settings-tabs")?.classList.contains("ui-kit-tabs--center")).toBe(true);
    expect(root.querySelector(".settings-tabs .ui-kit-tab.is-selected")?.textContent).toBe("Ввод");
    expect(root.querySelector(".settings-input-row__action")?.classList.contains("game-form-row-label")).toBe(true);
    expect(root.querySelector(".settings-input-table__left .settings-input-row")).not.toBeNull();
    expect(root.querySelector(".settings-input-table__right")?.textContent).toBe("");
    expect(root.querySelector("#settings-input-scrollbar")?.parentElement?.className).toBe("settings-input-table");
    expect(root.querySelector(".settings-modal__actions")).not.toBeNull();
    expect(Array.from(root.querySelectorAll(".settings-modal__footer .ui-kit-button")).map((button) => button.id)).toEqual(["settings-save-button", "settings-cancel-button"]);
  });

  // Проверяет, что нечётное число строк настроек делится между половинами окна с лишней строкой слева.
  it("splits input settings rows between settings window halves with the extra row on the left", () => {
    const root = document.createElement("div");
    document.body.append(root);

    const inputReferenceData = referenceDataWithActionTitles(["Первая", "Вторая", "Третья", "Четвёртая", "Пятая"]);
    dispose = render(() => <GameUi state={() => ({ ...state(), settingsVisible: true, referenceData: inputReferenceData, inputSettingsValues: { 1: 1, 2: 1, 3: 1, 4: 1, 5: 1 } })} />, root);

    const leftRows = Array.from(root.querySelectorAll(".settings-input-table__left .settings-input-row__action")).map((row) => row.textContent);
    const rightRows = Array.from(root.querySelectorAll(".settings-input-table__right .settings-input-row__action")).map((row) => row.textContent);

    expect(leftRows).toEqual(["Первая", "Вторая", "Третья"]);
    expect(rightRows).toEqual(["Четвёртая", "Пятая"]);
  });

  // Проверяет, что каждая вкладка настроек управляет собственной страницей.
  it("renders selected settings tab page", () => {
    const root = document.createElement("div");
    document.body.append(root);
    const [settingsState, setSettingsState] = createSignal<GameUiState>({ ...state(), settingsVisible: true, selectedSettingsTab: "video" });

    dispose = render(() => <GameUi state={settingsState} />, root);

    expect(root.querySelector(".settings-empty-page")?.textContent).toBe("");
    expect(root.querySelector(".settings-input-table")).toBeNull();

    setSettingsState({ ...settingsState(), selectedSettingsTab: "audio" });
    expect(root.querySelector(".settings-empty-page")?.textContent).toBe("");
    expect(root.querySelector(".settings-input-table")).toBeNull();

    setSettingsState({ ...settingsState(), selectedSettingsTab: "input" });
    expect(root.querySelector(".settings-input-table")).not.toBeNull();
  });

  // Проверяет, что длинная строка сдвигается влево и оставляет каретку внутри поля ввода.
  it("keeps long chat input caret inside viewport by moving text", async () => {
    const root = document.createElement("div");
    document.body.append(root);
    const [chatState, setChatState] = createSignal<GameUiState>({
      ...state(),
      chatState: {
        type: "chatState",
        selectedChatId: 1,
        tabs: [{ chatId: 1, title: "Server", communityTypeAcronym: "Server", duoChatKey: "", unreadCount: 0, messages: [] }],
      },
      chatInputText: "abcdefghijklmnopqrstuvwxyz0123456789",
      chatCursorIndex: 29,
      chatInputFocused: true,
    });

    dispose = render(() => <GameUi state={chatState} />, root);
    const viewport = root.querySelector<HTMLElement>(".ui-kit-text-input__viewport");
    const textMeasure = root.querySelectorAll<HTMLElement>(".ui-kit-text-input__measure")[0];
    const caretMeasure = root.querySelectorAll<HTMLElement>(".ui-kit-text-input__measure")[1];
    if (!viewport || !textMeasure || !caretMeasure) {
      throw new Error("Строка ввода чата не отрисована.");
    }
    viewport.getBoundingClientRect = () => rect(100);
    textMeasure.getBoundingClientRect = () => rect(260);
    caretMeasure.getBoundingClientRect = () => rect(210);

    setChatState({ ...chatState(), chatCursorIndex: 30 });
    await Promise.resolve();

    expect(root.querySelector<HTMLElement>(".ui-kit-text-input__text")?.style.transform).toBe("translateX(-118px)");
    expect(root.querySelector<HTMLElement>(".ui-kit-text-input__caret")?.style.left).toBe("92px");
  });

  // Проверяет, что каретка в конце длинной строки не обрезается правой границей поля.
  it("keeps long chat input caret visible at text end", async () => {
    const root = document.createElement("div");
    document.body.append(root);
    const [chatState, setChatState] = createSignal<GameUiState>({
      ...state(),
      chatState: {
        type: "chatState",
        selectedChatId: 1,
        tabs: [{ chatId: 1, title: "Server", communityTypeAcronym: "Server", duoChatKey: "", unreadCount: 0, messages: [] }],
      },
      chatInputText: "abcdefghijklmnopqrstuvwxyz0123456789",
      chatCursorIndex: 35,
      chatInputFocused: true,
    });

    dispose = render(() => <GameUi state={chatState} />, root);
    const viewport = root.querySelector<HTMLElement>(".ui-kit-text-input__viewport");
    const textMeasure = root.querySelectorAll<HTMLElement>(".ui-kit-text-input__measure")[0];
    const caretMeasure = root.querySelectorAll<HTMLElement>(".ui-kit-text-input__measure")[1];
    if (!viewport || !textMeasure || !caretMeasure) {
      throw new Error("Строка ввода чата не отрисована.");
    }
    viewport.getBoundingClientRect = () => rect(100);
    textMeasure.getBoundingClientRect = () => rect(260);
    caretMeasure.getBoundingClientRect = () => rect(260);

    setChatState({ ...chatState(), chatCursorIndex: 36 });
    await Promise.resolve();

    expect(root.querySelector<HTMLElement>(".ui-kit-text-input__text")?.style.transform).toBe("translateX(-168px)");
    expect(root.querySelector<HTMLElement>(".ui-kit-text-input__caret")?.style.left).toBe("92px");
  });

  it("remounts repeated chat error when sequence changes", () => {
    const root = document.createElement("div");
    document.body.append(root);
    const [chatState, setChatState] = createSignal<GameUiState>({
      ...state(),
      chatState: {
        type: "chatState",
        selectedChatId: 1,
        tabs: [{ chatId: 1, title: "Server", communityTypeAcronym: "Server", duoChatKey: "", unreadCount: 0, messages: [] }],
      },
      chatError: "Адресат не найден",
      chatErrorSeq: 1,
    });

    dispose = render(() => <GameUi state={chatState} />, root);
    const firstError = root.querySelector<HTMLElement>(".chat-error");
    expect(firstError?.style.animationName).toBe("chat-error-fade-odd");
    setChatState({ ...chatState(), chatErrorSeq: 2 });

    expect(root.querySelector(".chat-error")).toBe(firstError);
    expect(root.querySelector<HTMLElement>(".chat-error")?.style.animationName).toBe("chat-error-fade-even");
    expect(root.querySelector(".chat-error")?.textContent).toBe("Адресат не найден");
  });
});
