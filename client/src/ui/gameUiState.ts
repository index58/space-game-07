import type { Accessor } from "solid-js";
import { createStore } from "solid-js/store";
import type { ChatContextMenuState, ChatScrollState, GameCursorState } from "../game/InputController";
import type { ChatStateMessage, ConnectionStatus, ConstructorProductionJob, CosmicObject, DockingNotification, DockingWindowState, EquipmentGroup, EquipmentGroupRelation, ItemGroup, ReferenceDataMessage, Task, TaskItemGroup } from "../network/protocol";
import { createInitialUiKitDemoState, type UiKitDemoState } from "../ui-kit/showcaseState";
import type { GameUiControlState } from "../ui-kit/types";

export type SettingsTabValue = "video" | "audio" | "input";
export type ControlPanelTabValue = "object" | "equipment" | "pilotTools" | "schemas" | "blueprints" | "map";
export type ControlPanelEquipmentSubTabValue = "setup" | "usage";
export type ControlPanelConstructorTabValue = "items" | "objects";
export type ControlPanelUsageSelectValue = "leftObject" | "left" | "rightObject" | "right" | "constructorMaterialObject" | "constructorMaterials" | "constructorProductObject" | "constructorProducts";

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
  equipmentGroupRelations?: EquipmentGroupRelation[];
  // Р“СЂСѓРїРїС‹ РїСЂРµРґРјРµС‚РѕРІ РїРѕСЃР»РµРґРЅРµРіРѕ СЃРµСЂРІРµСЂРЅРѕРіРѕ СЃРЅРёРјРєР°, Р»РµР¶Р°С‰РёРµ РІРЅСѓС‚СЂРё РєРѕРЅС‚РµР№РЅРµСЂРѕРІ.
  itemGroups: ItemGroup[];
  // Задания изготовления в очередях конструкторов.
  // Задания оборудования в очередях.
  tasks?: Task[];
  // Предметы, зарезервированные заданиями.
  taskItemGroups?: TaskItemGroup[];
  constructorProductionJobs: ConstructorProductionJob[];
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
  // Маленькое окно ожидания или выполнения стыковки.
  dockingWindow?: DockingWindowState | null;
  // Отдельные уведомления стыковки, видимые до автоматического скрытия.
  dockingNotifications?: DockingNotification[];
  // Список объектов назначения для окна выбора пересадки.
  landingTargetObjectIds?: number[];
  // Выбранный объект назначения для пересадки.
  selectedLandingTargetObjectId?: number | null;
  // Открытое игровое меню вкладки, если игрок вызвал его правым кликом.
  chatContextMenu: ChatContextMenuState | null;
  // Положение и видимость игрового указателя мыши.
  gameCursor: GameCursorState;
  // ID контрола HUD, на который сейчас наведён игровой указатель.
  hoveredGameUiControlId: string | null;
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
  // ID объекта для левого контейнера использования.
  selectedControlPanelUsageLeftObjectId: number | null;
  // ID объекта для правого оборудования использования.
  selectedControlPanelUsageRightObjectId: number | null;
  // ID объекта для контейнера материалов конструктора.
  selectedControlPanelConstructorMaterialObjectId: number | null;
  // ID объекта для контейнера продукции конструктора.
  selectedControlPanelConstructorProductObjectId: number | null;
  // ID контейнера в левой панели подвкладки использования.
  selectedControlPanelUsageLeftContainerGroupId: number | null;
  // ID оборудования в правой панели подвкладки использования.
  selectedControlPanelUsageRightEquipmentGroupId: number | null;
  // Открытый выпадающий список подвкладки использования.
  openControlPanelUsageSelect: ControlPanelUsageSelectValue | null;
  // ID выбранных строк содержимого левого контейнера.
  selectedControlPanelUsageLeftItemGroupIds: number[];
  // ID выбранных строк содержимого правого контейнера.
  selectedControlPanelUsageRightItemGroupIds: number[];
  // ID контейнера материалов в использовании конструктора.
  selectedControlPanelConstructorMaterialContainerGroupId: number | null;
  // ID контейнера продукции в использовании конструктора.
  selectedControlPanelConstructorProductContainerGroupId: number | null;
  // Активная вкладка верхней части конструктора.
  selectedControlPanelConstructorTab: ControlPanelConstructorTabValue;
  // ID выбранной схемы предмета в конструкторе.
  selectedControlPanelConstructorSchemaId: number | null;
  // ID выбранного чертежа объекта в конструкторе.
  selectedControlPanelConstructorBlueprintId: number | null;
  // ID выбранной строки основной очереди конструктора.
  selectedControlPanelConstructorMainJobId: number | null;
  // Черновики признака включения групп оборудования по ID группы.
  // Показывает окно подтверждения слива топлива из бака.
  controlPanelFuelDrainDialogOpen: boolean;
  // Показывает окно подтверждения залива топлива в бак.
  controlPanelFuelFillDialogOpen: boolean;
  // Показывает окно подтверждения частичного переноса предметов между контейнерами.
  controlPanelContainerTransferDialogOpen: boolean;
  // Показывает окно выбора количества запусков изготовления.
  controlPanelConstructorProduceDialogOpen: boolean;
  // Максимальное количество предметов для частичного переноса между контейнерами.
  controlPanelContainerTransferMaxAmount: number;
  // Максимальное количество топлива, доступное для залива в бак.
  controlPanelFuelFillMaxAmount: number;
  // Максимальное количество запусков изготовления, доступное в окне выбора.
  controlPanelConstructorProduceMaxAmount: number;
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
  // Черновик пользовательского названия выбранной группы оборудования.
  controlPanelEquipmentTitleText: string;
  // Начало выделения в поле названия группы оборудования.
  controlPanelEquipmentTitleSelectionStart: number;
  // Конец выделения в поле названия группы оборудования.
  controlPanelEquipmentTitleSelectionEnd: number;
  // Признак активного фокуса поля названия группы оборудования.
  controlPanelEquipmentTitleFocused: boolean;
  // Состояние прокрутки списка групп оборудования.
  controlPanelEquipmentListScroll: ChatScrollState;
  // Состояния прокрутки обычных списков по ID списка.
  listScroll: Record<string, ChatScrollState>;
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
  equipmentGroupRelations: [],
  itemGroups: [],
  tasks: [],
  taskItemGroups: [],
  constructorProductionJobs: [],
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
  dockingWindow: null,
  dockingNotifications: [],
  landingTargetObjectIds: [],
  selectedLandingTargetObjectId: null,
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
  controlPanelContainerTransferMaxAmount: 0,
  controlPanelFuelFillMaxAmount: 0,
  controlPanelConstructorProduceMaxAmount: 100,
  controlPanelFuelDrainAmount: 0,
  controlPanelFuelDrainAmountText: "0",
  controlPanelFuelDrainAmountSelectionStart: 1,
  controlPanelFuelDrainAmountSelectionEnd: 1,
  controlPanelFuelDrainAmountFocused: false,
  controlPanelEquipmentEnabledDrafts: {},
  controlPanelEquipmentEnabledCountDrafts: {},
  controlPanelEquipmentTitleText: "",
  controlPanelEquipmentTitleSelectionStart: 0,
  controlPanelEquipmentTitleSelectionEnd: 0,
  controlPanelEquipmentTitleFocused: false,
  controlPanelEquipmentListScroll: { visible: false, thumbTopPercent: 0, thumbHeightPercent: 100, contentOffsetPx: 0, dragging: false },
  listScroll: {},
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
