import { createSignal, type Accessor } from "solid-js";
import type { ConnectionStatus, CosmicObject, EquipmentGroup, ReferenceDataMessage } from "../network/protocol";

export type GameUiState = {
  // Состояние сетевого подключения.
  status: ConnectionStatus;
  // Посещаемый объект игрока, если он уже получен.
  selfObject: CosmicObject | null;
  // Объекты последнего серверного снимка, доступные клиенту.
  objects: CosmicObject[];
  // Группы оборудования последнего серверного снимка, нужные UI для панели пилота.
  equipmentGroups: EquipmentGroup[];
  // Справочники клиента, нужные UI для определения типов объектов.
  referenceData: ReferenceDataMessage | null;
  // Путь к текстуре текущего объекта.
  textureFilePath: string | null;
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
  referenceData: null,
  textureFilePath: null,
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
