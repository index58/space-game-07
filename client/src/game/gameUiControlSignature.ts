import type { GameUiState } from "../ui/gameUiState";

export type GameUiControlLayoutViewport = {
  // Ширина браузерного окна в пикселях.
  width: number;
  // Высота браузерного окна в пикселях.
  height: number;
  // Ширина Phaser-области в пикселях.
  scaleWidth: number;
  // Высота Phaser-области в пикселях.
  scaleHeight: number;
};

// Собирает только те признаки, которые могут изменить набор или геометрию DOM-контролов.
export const getGameUiControlLayoutSignature = (
  state: GameUiState,
  viewport: GameUiControlLayoutViewport,
): string => JSON.stringify({
  viewport,
  status: state.status,
  hasSelfObject: state.selfObject !== null,
  chat: {
    selectedChatId: state.chatState?.selectedChatId ?? null,
    tabIds: state.chatState?.tabs.map((tab) => `${tab.chatId}:${tab.unreadCount ?? 0}`).join(",") ?? "",
    inputFocused: state.chatInputFocused,
    contextMenu: state.chatContextMenu ? {
      chatId: state.chatContextMenu.chatId,
      x: state.chatContextMenu.x,
      y: state.chatContextMenu.y,
    } : null,
    scrollbarVisible: state.chatScroll.visible,
  },
  showcase: {
    visible: state.uiKitShowcaseVisible,
    state: state.uiKitDemoState,
  },
  settings: {
    visible: state.settingsVisible,
    selectedTab: state.selectedSettingsTab,
    actionCount: state.referenceData?.ActionType.MaxID ?? 0,
    inputEventCount: state.referenceData?.InputEventType.MaxID ?? 0,
    openActionId: state.openInputSettingsActionId,
    saving: state.inputSettingsSaving,
    hasError: state.inputSettingsError !== null,
    listScroll: scrollSignature(state.inputSettingsScroll),
    dropdownScroll: scrollSignature(state.inputSettingsDropdownScroll),
  },
  controlPanel: {
    visible: state.controlPanelVisible,
    selectedTab: state.selectedControlPanelTab,
    selectedEquipmentTab: state.selectedControlPanelEquipmentTab,
    selectedEquipmentGroupId: state.selectedControlPanelEquipmentGroupId,
    selectedUsageLeftContainerGroupId: state.selectedControlPanelUsageLeftContainerGroupId,
    selectedUsageRightEquipmentGroupId: state.selectedControlPanelUsageRightEquipmentGroupId,
    openUsageSelect: state.openControlPanelUsageSelect,
    selectedUsageLeftItemGroupIds: state.selectedControlPanelUsageLeftItemGroupIds.join(","),
    selectedUsageRightItemGroupIds: state.selectedControlPanelUsageRightItemGroupIds.join(","),
    selectedConstructorMaterialContainerGroupId: state.selectedControlPanelConstructorMaterialContainerGroupId,
    selectedConstructorTab: state.selectedControlPanelConstructorTab,
    selectedConstructorSchemaId: state.selectedControlPanelConstructorSchemaId,
    selectedConstructorBlueprintId: state.selectedControlPanelConstructorBlueprintId,
    fuelDrainDialogOpen: state.controlPanelFuelDrainDialogOpen,
    fuelFillDialogOpen: state.controlPanelFuelFillDialogOpen,
    containerTransferDialogOpen: state.controlPanelContainerTransferDialogOpen,
    containerTransferMaxAmount: state.controlPanelContainerTransferMaxAmount,
    fuelFillMaxAmount: state.controlPanelFuelFillMaxAmount,
    fuelDrainAmount: state.controlPanelFuelDrainAmount,
    fuelDrainText: state.controlPanelFuelDrainAmountText,
    fuelDrainFocused: state.controlPanelFuelDrainAmountFocused,
    equipmentIds: state.equipmentGroups.map((group) => `${group.ID}:${group.CosmicObjectID}:${group.EquipmentItemModelID}`).join(","),
    itemGroupIds: state.itemGroups.map((group) => `${group.ID}:${group.ContainerEquipmentGroupID}:${group.ContentItemModelID}:${group.Count}`).join(","),
    itemtypeInternalUsable: Object.values(state.referenceData?.Itemtype.Items ?? {}).sort((left, right) => left.ID - right.ID).map((itemtype) => `${itemtype.ID}:${itemtype.IsInternalUsable}`).join(","),
    schemas: Object.values(state.referenceData?.Schema.Items ?? {}).sort((left, right) => left.ID - right.ID).map((schema) => `${schema.ID}:${schema.ItemModelID}`).join(","),
    blueprints: Object.values(state.referenceData?.Blueprint.Items ?? {}).sort((left, right) => left.ID - right.ID).map((blueprint) => `${blueprint.ID}:${blueprint.CosmicObjectModelID}`).join(","),
    equipmentListScroll: scrollSignature(state.controlPanelEquipmentListScroll),
    listScroll: Object.entries(state.listScroll).sort(([left], [right]) => left.localeCompare(right)).map(([id, scroll]) => `${id}:${scrollSignatureText(scroll)}`).join("|"),
    hasSelfObject: state.selfObject !== null,
    objectId: state.selfObject?.ID ?? null,
  },
});

const scrollSignature = (scroll: GameUiState["inputSettingsScroll"]) => ({
  visible: scroll.visible,
  thumbTopPercent: scroll.thumbTopPercent,
  thumbHeightPercent: scroll.thumbHeightPercent,
  contentOffsetPx: scroll.contentOffsetPx,
  dragging: scroll.dragging,
});

// Возвращает компактную строку состояния прокрутки для признака раскладки.
const scrollSignatureText = (scroll: GameUiState["inputSettingsScroll"]): string =>
  `${Number(scroll.visible)}:${Math.round(scroll.thumbTopPercent)}:${Math.round(scroll.thumbHeightPercent)}:${Math.round(scroll.contentOffsetPx)}:${Number(scroll.dragging)}`;
