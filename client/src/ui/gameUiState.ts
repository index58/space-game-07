import { createSignal, type Accessor } from "solid-js";
import type { ChatContextMenuState, ChatScrollState, GameCursorState } from "../game/InputController";
import type { ChatStateMessage, ConnectionStatus, CosmicObject, EquipmentGroup, ReferenceDataMessage } from "../network/protocol";
import { createInitialUiKitDemoState, type UiKitDemoState } from "../ui-kit/showcaseState";
import type { GameUiControlState } from "../ui-kit/types";

export type GameUiState = {
  // Состояние сетевого подключения.
  status: ConnectionStatus;
  // Посещаемый объект игрока, если он уже получен.
  selfObject: CosmicObject | null;
  // Объекты последнего серверного снимка, доступные клиенту.
  objects: CosmicObject[];
  // Группы оборудования последнего серверного снимка, нужные UI для панели пилота.
  equipmentGroups: EquipmentGroup[];
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
  const [state, setState] = createSignal<GameUiState>(initialGameUiState);

  return {
    state,
    update: setState,
  };
};
