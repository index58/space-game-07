export type ConnectionStatus = "connecting" | "connected" | "waiting";

export type ClientInputState = {
  thrustForward: boolean;
  thrustBackward: boolean;
  thrustLeft: boolean;
  thrustRight: boolean;
  targetRotationDelta: number;
};

export type ClientInputMessage = ClientInputState & {
  type: "input";
  seq: number;
};

export type SnapshotObjectKind = "ship" | "asteroid" | "station";

export type SnapshotObject = {
  id: number;
  modelAcronym: string;
  kind: SnapshotObjectKind;
  x: number;
  y: number;
  velocityX: number;
  velocityY: number;
  rotation: number;
  angularVelocity: number;
  targetRotation: number;
};

export type SnapshotMessage = {
  type: "snapshot";
  tick: number;
  selfObjectId: number;
  objects: SnapshotObject[];
};
