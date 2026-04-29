import { ASSET_KEYS, ASSET_PATHS } from "./assets";
import type { CosmicObjectModel, ShipState, SimObject } from "../domain/types";

// Временная конфигурация прототипа взята из старого shared/data/cosmic_object_models.json.
const TEXTURE_SCALE = 4;
const MASS_SCALE = 1000;
const THRUST_SCALE = 10_000_000;
const TORQUE_SCALE = 1000;

export const SHIP_BAT: CosmicObjectModel = {
  acronym: "ship_bat",
  titleRu: "Летучая мышь",
  kind: "ship",
  textureKey: ASSET_KEYS.shipBat,
  texturePath: ASSET_PATHS[ASSET_KEYS.shipBat],
  textureWidth: 256,
  textureHeight: 512,
  textureBodyOriginX: 126,
  textureBodyOriginY: 259,
  textureBodyWidth: 88,
  textureBodyLength: 90,
  textureScale: TEXTURE_SCALE,
  massKg: 7.92 * MASS_SCALE,
  thrustN: 0.006439507649442245 * THRUST_SCALE,
  torqueNm: 653.5649999999999 * TORQUE_SCALE,
};

export const ASTEROID_0002: CosmicObjectModel = {
  acronym: "asteroid_0002",
  titleRu: "Астероид",
  kind: "asteroid",
  textureKey: ASSET_KEYS.asteroid0002,
  texturePath: ASSET_PATHS[ASSET_KEYS.asteroid0002],
  textureWidth: 2048,
  textureHeight: 2048,
  textureBodyOriginX: 988,
  textureBodyOriginY: 1289,
  textureBodyWidth: 804,
  textureBodyLength: 783,
  textureScale: TEXTURE_SCALE,
  massKg: 629.532 * MASS_SCALE,
  thrustN: 0,
  torqueNm: 0,
};

export const STATION_TINY_CRUMB: CosmicObjectModel = {
  acronym: "station_tiny_crumb",
  titleRu: "Крошка",
  kind: "station",
  textureKey: ASSET_KEYS.stationTinyCrumb,
  texturePath: ASSET_PATHS[ASSET_KEYS.stationTinyCrumb],
  textureWidth: 2048,
  textureHeight: 2048,
  textureBodyOriginX: 996,
  textureBodyOriginY: 738,
  textureBodyWidth: 225,
  textureBodyLength: 825,
  textureScale: TEXTURE_SCALE,
  massKg: 185.625 * MASS_SCALE,
  thrustN: 0,
  torqueNm: 0,
};

export const createInitialShipState = (): ShipState => ({
  model: SHIP_BAT,
  position: { x: 0, y: 0 },
  velocity: { x: 0, y: 0 },
  rotation: 0,
  angularVelocity: 0,
});

export const STATIC_OBJECTS: SimObject[] = [
  { model: ASTEROID_0002, position: { x: -500, y: 800 }, rotation: 0 },
  { model: STATION_TINY_CRUMB, position: { x: 500, y: 500 }, rotation: 0 },
];
