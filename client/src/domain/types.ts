// Типы локальной симуляции повторяют только тот срез модели мира, который нужен первому прототипу.
export type CosmicObjectKind = "ship" | "asteroid" | "station";

export type WorldVector = {
  x: number;
  y: number;
};

export type CosmicObjectModel = {
  acronym: string;
  titleRu: string;
  kind: CosmicObjectKind;
  textureKey: string;
  texturePath: string;
  textureWidth: number;
  textureHeight: number;
  textureBodyOriginX: number;
  textureBodyOriginY: number;
  textureBodyWidth: number;
  textureBodyLength: number;
  textureScale: number;
  massKg: number;
  thrustN: number;
  torqueNm: number;
};

export type SimObject = {
  model: CosmicObjectModel;
  position: WorldVector;
  rotation: number;
};

export type ShipState = SimObject & {
  velocity: WorldVector;
  angularVelocity: number;
};
