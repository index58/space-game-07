// Ключи ассетов стабильны внутри клиента, а пути соответствуют каталогу client/public.
export const ASSET_KEYS = {
  background: "world.background.space",
} as const;

// Связывает стабильные ключи Phaser с реальными файлами, отдаваемыми Vite из public.
export const ASSET_PATHS = {
  [ASSET_KEYS.background]: "/assets/world/backgrounds/space-background.jpg",
} as const;
