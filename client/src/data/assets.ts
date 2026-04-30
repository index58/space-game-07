// Ключи ассетов стабильны внутри клиента, а пути соответствуют каталогу client/public.
export const ASSET_KEYS = {
  background: "world.background.space",
  shipBat: "world.ship.ship_bat",
  asteroid0002: "world.asteroid.asteroid_0002",
  stationTinyCrumb: "world.station.station_tiny_crumb",
} as const;

// связывает стабильные ключи Phaser с реальными файлами, отдаваемыми Vite из public.
export const ASSET_PATHS = {
  [ASSET_KEYS.background]: "/assets/world/backgrounds/space-background.jpg",
  [ASSET_KEYS.shipBat]: "/assets/world/cosmic-objects/ships/256x512/ship_256x512_0008.png",
  [ASSET_KEYS.asteroid0002]: "/assets/world/cosmic-objects/asteroids/asteroid_0002.png",
  [ASSET_KEYS.stationTinyCrumb]: "/assets/world/cosmic-objects/stations/station_0064.png",
} as const;
