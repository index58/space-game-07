// отражает текущее состояние WebSocket-клиента для сцены и отладочного слоя.
export type ConnectionStatus = "connecting" | "connected" | "waiting";

// хранит последний пользовательский ввод, еще не упакованный в сетевое сообщение.
export type ClientInputState = {
  thrustForward: boolean;
  thrustBackward: boolean;
  thrustLeft: boolean;
  thrustRight: boolean;
  targetRotationDelta: number;
};

// добавляет к вводу тип сообщения и порядковый номер для серверного протокола.
export type ClientInputMessage = ClientInputState & {
  type: "input";
  seq: number;
};

// повторяет серверные категории объектов, от которых зависит выбор текстуры.
export type SnapshotObjectKind = "ship" | "asteroid" | "station";

// описывает один объект из серверного снимка мира.
export type SnapshotObject = {
  id: number;
  modelAcronym: string;
  kind: SnapshotObjectKind;
  textureScale: number;
  x: number;
  y: number;
  velocityX: number;
  velocityY: number;
  rotation: number;
  angularVelocity: number;
  targetRotation: number;
};

// является полным состоянием мира, которое сервер регулярно отправляет клиенту.
export type SnapshotMessage = {
  type: "snapshot";
  tick: number;
  selfObjectId: number;
  objects: SnapshotObject[];
};
