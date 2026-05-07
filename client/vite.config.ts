import { defineConfig } from "vite";
import solid from "vite-plugin-solid";

// Конфигурация подключает компилятор SolidJS для клиентского UI-слоя.
export default defineConfig({
  plugins: [solid()],
  build: {
    // Phaser вынесен в отдельный кэшируемый chunk, поэтому лимит предупреждения выставлен выше его текущего размера.
    chunkSizeWarningLimit: 1800,
    rollupOptions: {
      output: {
        manualChunks: (id) => {
          const normalizedId = id.replaceAll("\\", "/");
          if (normalizedId.includes("/node_modules/phaser/")) {
            return "phaser";
          }
          if (normalizedId.includes("/node_modules/solid-js/")) {
            return "solid";
          }
          if (normalizedId.includes("/node_modules/")) {
            return "vendor";
          }
          return undefined;
        },
      },
    },
  },
});
