import { describe, expect, it } from "vitest";
import type { GameUiState } from "../ui/gameUiState";
import { getGameUiControlLayoutSignature } from "./gameUiControlSignature";

const scrollState = {
  visible: false,
  thumbTopPercent: 0,
  thumbHeightPercent: 100,
  contentOffsetPx: 0,
  dragging: false,
};

const state = (partial: Partial<GameUiState> = {}): GameUiState => ({
  status: "connected",
  selfObject: null,
  objects: [],
  equipmentGroups: [],
  selectedPilotToolIndex: 0,
  referenceData: null,
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
  gameCursor: { visible: true, x: 100, y: 100 },
  chatScroll: scrollState,
  uiKitShowcaseVisible: false,
  settingsVisible: true,
  controlPanelVisible: false,
  selectedSettingsTab: "input",
  selectedControlPanelTab: "object",
  selectedControlPanelEquipmentTab: "setup",
  selectedControlPanelEquipmentGroupId: null,
  controlPanelEquipmentEnabledDrafts: {},
  controlPanelEquipmentEnabledCountDrafts: {},
  controlPanelEquipmentListScroll: scrollState,
  controlPanelObjectEnabled: true,
  controlPanelObjectTitleText: "Ship",
  controlPanelObjectTitleSelectionStart: 4,
  controlPanelObjectTitleSelectionEnd: 4,
  controlPanelObjectTitleFocused: false,
  inputSettingsValues: { 1: 1 },
  openInputSettingsActionId: null,
  inputSettingsError: null,
  inputSettingsSaving: false,
  inputSettingsScroll: scrollState,
  inputSettingsDropdownScroll: scrollState,
  uiKitDemoState: {
    buttonClicks: 0,
    checkboxChecked: true,
    radioValue: "a",
    dropdownOpen: false,
    dropdownValue: "one",
    listValue: "1",
    treeValue: "root",
    virtualStartIndex: 20,
    tabValue: "one",
    editText: "text",
    editSelectionStart: 1,
    editSelectionEnd: 3,
    scrollbarTopPercent: 20,
    scrollbarDrag: null,
    sliderValue: 30,
    stepperValue: 7,
    splitterVertical: true,
    menuOpen: false,
    tooltipVisible: false,
  },
  uiControls: [],
  fps: 180,
  zoom: 4,
  ...partial,
});

const viewport = {
  width: 1200,
  height: 800,
  scaleWidth: 1200,
  scaleHeight: 800,
};

const referenceDataWithInternalUsableItemtype = (isInternalUsable: boolean): NonNullable<GameUiState["referenceData"]> => ({
  ActionType: { MaxID: 0, Items: {} },
  Blueprint: { MaxID: 0, Items: {} },
  BlueprintComponent: { MaxID: 0, Items: {} },
  CosmicObjectModel: { MaxID: 0, Items: {} },
  CosmicObjectType: { MaxID: 0, Items: {} },
  DefaultActionInputSetting: { MaxID: 0, Items: {} },
  InputEventType: { MaxID: 0, Items: {} },
  ItemModel: { MaxID: 0, Items: {} },
  Itemtype: { MaxID: 1, Items: { "1": { ID: 1, Acronym: "Weapon", IsPilotInstrument: true, IsInternalUsable: isInternalUsable } } },
  NpcClan: { MaxID: 0, Items: {} },
  Schema: { MaxID: 0, Items: {} },
  SchemaComponent: { MaxID: 0, Items: {} },
  type: "referenceData",
});

describe("getGameUiControlLayoutSignature", () => {
  // Проверяет, что частые значения кадра не заставляют заново измерять DOM-контролы.
  it("ignores frame-only fields", () => {
    const first = getGameUiControlLayoutSignature(state(), viewport);
    const second = getGameUiControlLayoutSignature(state({
      fps: 81,
      gameCursor: { visible: true, x: 700, y: 500 },
      zoom: 5,
    }), viewport);

    expect(second).toBe(first);
  });

  // Проверяет, что изменение раскладки окна настроек помечает геометрию UI устаревшей.
  it("changes when settings layout changes", () => {
    const first = getGameUiControlLayoutSignature(state(), viewport);
    const second = getGameUiControlLayoutSignature(state({
      openInputSettingsActionId: 1,
      inputSettingsDropdownScroll: { ...scrollState, visible: true, contentOffsetPx: 32 },
    }), viewport);

    expect(second).not.toBe(first);
  });

  // Проверяет, что смена страницы настроек пересобирает DOM-области интерактивных контролов.
  it("changes when selected settings tab changes", () => {
    const first = getGameUiControlLayoutSignature(state({ selectedSettingsTab: "input" }), viewport);
    const second = getGameUiControlLayoutSignature(state({ selectedSettingsTab: "video" }), viewport);

    expect(second).not.toBe(first);
  });

  // Проверяет, что смена страницы панели управления пересобирает DOM-области ее вкладок и содержимого.
  it("changes when selected control panel tab changes", () => {
    const first = getGameUiControlLayoutSignature(state({ controlPanelVisible: true, selectedControlPanelTab: "object" }), viewport);
    const second = getGameUiControlLayoutSignature(state({ controlPanelVisible: true, selectedControlPanelTab: "equipment" }), viewport);

    expect(second).not.toBe(first);
  });

  // Проверяет, что прокрутка списка оборудования пересобирает hit-test области пунктов.
  it("changes when equipment list scroll changes", () => {
    const first = getGameUiControlLayoutSignature(state({ controlPanelVisible: true, selectedControlPanelTab: "equipment" }), viewport);
    const second = getGameUiControlLayoutSignature(state({
      controlPanelVisible: true,
      selectedControlPanelTab: "equipment",
      controlPanelEquipmentListScroll: { ...scrollState, visible: true, contentOffsetPx: 32 },
    }), viewport);

    expect(second).not.toBe(first);
  });

  // Проверяет, что изменение внутреннего использования типа оборудования пересобирает состояние кнопки использования.
  it("changes when selected equipment usability changes", () => {
    const first = getGameUiControlLayoutSignature(state({
      controlPanelVisible: true,
      selectedControlPanelTab: "equipment",
      referenceData: referenceDataWithInternalUsableItemtype(false),
    }), viewport);
    const second = getGameUiControlLayoutSignature(state({
      controlPanelVisible: true,
      selectedControlPanelTab: "equipment",
      referenceData: referenceDataWithInternalUsableItemtype(true),
    }), viewport);

    expect(second).not.toBe(first);
  });
});
