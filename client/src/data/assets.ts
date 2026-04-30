// Ключи ассетов стабильны внутри клиента, а пути соответствуют каталогу client/public.
export const ASSET_KEYS = {
  background: "world.background.space",
  shipBat: "world.ship.ship_bat",
  asteroid0002: "world.asteroid.asteroid_0002",
  stationTinyCrumb: "world.station.station_tiny_crumb",
} as const;

// Связывает стабильные ключи Phaser с реальными файлами, отдаваемыми Vite из public.
export const ASSET_PATHS = {
  [ASSET_KEYS.background]: "/assets/world/backgrounds/space-background.jpg",
  [ASSET_KEYS.shipBat]: "/assets/world/cosmic-objects/ships/256x512/ship_256x512_0008.png",
  [ASSET_KEYS.asteroid0002]: "/assets/world/cosmic-objects/asteroids/asteroid_0002.png",
  [ASSET_KEYS.stationTinyCrumb]: "/assets/world/cosmic-objects/stations/station_0064.png",
} as const;

// Связывает ID модели из серверных данных с уже загруженным ключом текстуры.
export const ASSET_KEY_BY_COSMIC_OBJECT_MODEL_ID: Record<number, string> = {
  1: ASSET_KEYS.shipBat,
  2: ASSET_KEYS.asteroid0002,
  3: ASSET_KEYS.stationTinyCrumb,
};

// Повторяет масштаб текстуры из серверного справочника моделей для клиентского рендера.
export const TEXTURE_SCALE_BY_COSMIC_OBJECT_MODEL_ID: Record<number, number> = {
  1: 4,
  2: 4,
  3: 4,
};
