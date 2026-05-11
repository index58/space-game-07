import type { Accessor } from "solid-js";
import { createStore } from "solid-js/store";
import type { ChatContextMenuState, ChatScrollState, GameCursorState } from "../game/InputController";
import type { ChatStateMessage, ConnectionStatus, CosmicObject, EquipmentGroup, ItemGroup, ReferenceDataMessage } from "../network/protocol";
import { createInitialUiKitDemoState, type UiKitDemoState } from "../ui-kit/showcaseState";
import type { GameUiControlState } from "../ui-kit/types";

export type SettingsTabValue = "video" | "audio" | "input";
export type ControlPanelTabValue = "object" | "equipment" | "pilotTools" | "schemas" | "blueprints" | "map";
export type ControlPanelEquipmentSubTabValue = "setup" | "usage";

export type GameUiState = {
  // Состояние сетевого подключения.
  status: ConnectionStatus;
  // Посещаемый объект игрока, если он уже получен.
  selfObject: CosmicObject | null;
  // Объекты последнего серверного снимка, доступные клиенту.
  objects: CosmicObject[];
  // Группы оборудования последнего серверного снимка, нужные UI для панели пилота.
  equipmentGroups: EquipmentGroup[];
  // Группы предметов последнего серверного снимка, лежащие внутри контейнеров.
  itemGroups: ItemGroup[];
  // Выбранный индекс среди десяти ячеек панели пилота.
  selectedPilotToolIndex: number;
  // Справочники клиента, нужные UI для определения типов объектов.
  referenceData: ReferenceDataMessage | null;
  // Путь к текстуре текущего объекта.
  textureFilePath: string | null;
  // Сетевое состояние доступных вкладок и истории чата.
  chatState: ChatStateMessage | null;
  // Локальная строка ввода, которую HUD только показывает.
  chatInputText: string;
  // Позиция каретки внутри локальной строки ввода.
  chatCursorIndex: number;
  // Начало выделенного диапазона в строке чата.
  chatSelectionStart: number;
  // Конец выделенного диапазона в строке чата.
  chatSelectionEnd: number;
  // Признак направления печатных клавиш в панель чата.
  chatInputFocused: boolean;
  // Последняя ошибка отправки текста, полученная от сервера.
  chatError: string | null;
  // Порядковый номер ошибки, чтобы одинаковый текст заново запускал анимацию.
  chatErrorSeq: number;
  // Открытое игровое меню вкладки, если игрок вызвал его правым кликом.
  chatContextMenu: ChatContextMenuState | null;
  // Положение и видимость игрового указателя мыши.
  gameCursor: GameCursorState;
  // Состояние полосы прокрутки выбранного чата.
  chatScroll: ChatScrollState;
  // Показывает отладочную панель примеров единого UI Kit.
  uiKitShowcaseVisible: boolean;
  // Показывает модальное окно настроек игрока.
  settingsVisible: boolean;
  // Показывает модальное окно панели управления объектом.
  controlPanelVisible: boolean;
  // Текущая страница модального окна настроек.
  selectedSettingsTab: SettingsTabValue;
  // Текущая страница модального окна панели управления.
  selectedControlPanelTab: ControlPanelTabValue;
  // Текущая подстраница оборудования в панели управления.
  selectedControlPanelEquipmentTab: ControlPanelEquipmentSubTabValue;
  // ID выбранной группы оборудования в панели управления.
  selectedControlPanelEquipmentGroupId: number | null;
  // ID контейнера в левой панели подвкладки использования.
  selectedControlPanelUsageLeftContainerGroupId: number | null;
  // ID оборудования в правой панели подвкладки использования.
  selectedControlPanelUsageRightEquipmentGroupId: number | null;
  // Открытый выпадающий список подвкладки использования.
  openControlPanelUsageSelect: "left" | "right" | null;
  // ID выбранных строк содержимого левого контейнера.
  selectedControlPanelUsageLeftItemGroupIds: number[];
  // ID выбранных строк содержимого правого контейнера.
  selectedControlPanelUsageRightItemGroupIds: number[];
  // Черновики признака включения групп оборудования по ID группы.
  // Показывает окно подтверждения слива топлива из бака.
  controlPanelFuelDrainDialogOpen: boolean;
  // Показывает окно подтверждения залива топлива в бак.
  controlPanelFuelFillDialogOpen: boolean;
  // Показывает окно подтверждения частичного переноса предметов между контейнерами.
  controlPanelContainerTransferDialogOpen: boolean;
  // Максимальное количество предметов для частичного переноса между контейнерами.
  controlPanelContainerTransferMaxAmount: number;
  // Максимальное количество топлива, доступное для залива в бак.
  controlPanelFuelFillMaxAmount: number;
  // Количество топлива, выбранное для слива из бака.
  controlPanelFuelDrainAmount: number;
  // Текст количества топлива в поле слива из бака.
  controlPanelFuelDrainAmountText: string;
  // Начало выделения в поле количества слива топлива.
  controlPanelFuelDrainAmountSelectionStart: number;
  // Конец выделения в поле количества слива топлива.
  controlPanelFuelDrainAmountSelectionEnd: number;
  // Признак активного фокуса поля количества слива топлива.
  controlPanelFuelDrainAmountFocused: boolean;
  controlPanelEquipmentEnabledDrafts: Record<number, boolean>;
  // Черновики количества включенных единиц оборудования по ID группы.
  controlPanelEquipmentEnabledCountDrafts: Record<number, number>;
  // Состояние прокрутки списка групп оборудования.
  controlPanelEquipmentListScroll: ChatScrollState;
  // Черновик признака включения объекта для интерактивного переключателя панели управления.
  controlPanelObjectEnabled: boolean;
  // Черновик пользовательского названия объекта для интерактивного поля панели управления.
  controlPanelObjectTitleText: string;
  // Начало выделения в поле названия объекта.
  controlPanelObjectTitleSelectionStart: number;
  // Конец выделения в поле названия объекта.
  controlPanelObjectTitleSelectionEnd: number;
  // Признак активного фокуса поля названия объекта.
  controlPanelObjectTitleFocused: boolean;
  // Черновик выбранных привязок ввода по ID действия.
  inputSettingsValues: Record<number, number>;
  // ID действия с раскрытым списком событий ввода.
  openInputSettingsActionId: number | null;
  // Последняя ошибка сохранения настроек ввода.
  inputSettingsError: string | null;
  // Признак ожидания подтверждения сохранения настроек.
  inputSettingsSaving: boolean;
  // Состояние прокрутки списка действий в окне настроек.
  inputSettingsScroll: ChatScrollState;
  // Состояние прокрутки раскрытого списка событий ввода.
  inputSettingsDropdownScroll: ChatScrollState;
  // Состояние интерактивных примеров единого UI Kit.
  uiKitDemoState: UiKitDemoState;
  // Последний снимок зарегистрированных интерактивных контролов HUD.
  uiControls: GameUiControlState[];
  // Текущая частота кадров.
  fps: number;
  // Рассчитанный масштаб камеры.
  zoom: number;
};

export type GameUiController = {
  // Реактивный снимок состояния для Solid-компонентов.
  state: Accessor<GameUiState>;
  // Передаёт свежие данные игрового кадра в UI.
  update: (state: GameUiState) => void;
};

const initialGameUiState: GameUiState = {
  status: "connecting",
  selfObject: null,
  objects: [],
  equipmentGroups: [],
  itemGroups: [],
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
  gameCursor: { visible: false, x: 0, y: 0 },
  chatScroll: { visible: false, thumbTopPercent: 0, thumbHeightPercent: 100, contentOffsetPx: 0, dragging: false },
  uiKitShowcaseVisible: false,
  settingsVisible: false,
  controlPanelVisible: false,
  selectedSettingsTab: "input",
  selectedControlPanelTab: "object",
  selectedControlPanelEquipmentTab: "setup",
  selectedControlPanelEquipmentGroupId: null,
  selectedControlPanelUsageLeftContainerGroupId: null,
  selectedControlPanelUsageRightEquipmentGroupId: null,
  openControlPanelUsageSelect: null,
  selectedControlPanelUsageLeftItemGroupIds: [],
  selectedControlPanelUsageRightItemGroupIds: [],
  controlPanelFuelDrainDialogOpen: false,
  controlPanelFuelFillDialogOpen: false,
  controlPanelContainerTransferDialogOpen: false,
  controlPanelContainerTransferMaxAmount: 0,
  controlPanelFuelFillMaxAmount: 0,
  controlPanelFuelDrainAmount: 0,
  controlPanelFuelDrainAmountText: "0",
  controlPanelFuelDrainAmountSelectionStart: 1,
  controlPanelFuelDrainAmountSelectionEnd: 1,
  controlPanelFuelDrainAmountFocused: false,
  controlPanelEquipmentEnabledDrafts: {},
  controlPanelEquipmentEnabledCountDrafts: {},
  controlPanelEquipmentListScroll: { visible: false, thumbTopPercent: 0, thumbHeightPercent: 100, contentOffsetPx: 0, dragging: false },
  controlPanelObjectEnabled: false,
  controlPanelObjectTitleText: "",
  controlPanelObjectTitleSelectionStart: 0,
  controlPanelObjectTitleSelectionEnd: 0,
  controlPanelObjectTitleFocused: false,
  inputSettingsValues: {},
  openInputSettingsActionId: null,
  inputSettingsError: null,
  inputSettingsSaving: false,
  inputSettingsScroll: { visible: false, thumbTopPercent: 0, thumbHeightPercent: 100, contentOffsetPx: 0, dragging: false },
  inputSettingsDropdownScroll: { visible: false, thumbTopPercent: 0, thumbHeightPercent: 100, contentOffsetPx: 0, dragging: false },
  uiKitDemoState: createInitialUiKitDemoState(),
  uiControls: [],
  fps: 0,
  zoom: 1,
};

// Создаёт единый мост состояния между Phaser-сценой и SolidJS.
export const createGameUiController = (): GameUiController => {
  const [state, setState] = createStore<GameUiState>(initialGameUiState);

  return {
    state: () => state,
    update: setState,
  };
};
