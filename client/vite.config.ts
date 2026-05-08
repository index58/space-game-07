import { readFileSync } from "node:fs";
import { dirname, resolve } from "node:path";
import { fileURLToPath } from "node:url";
import { defineConfig } from "vite";
import solid from "vite-plugin-solid";

const configDirectory = dirname(fileURLToPath(import.meta.url));
const visualTestAccountRoute = "/local/visual-test-account.local.json";
const visualTestAccountFile = resolve(configDirectory, ".visual-test-account.local.json");

// Отдает локальный секрет только dev-сервером, чтобы файл не копировался в сборку.
const visualTestAccountPlugin = () => ({
  name: "space-game-visual-test-account",
  apply: "serve" as const,
  configureServer(server: {
    middlewares: {
      use(
        route: string,
        handler: (
          request: { method?: string },
          response: { setHeader(name: string, value: string): void; end(body?: string): void },
          next: () => void,
        ) => void,
      ): void;
    };
  }) {
    server.middlewares.use(visualTestAccountRoute, (request, response, next) => {
      if (request.method !== "GET" && request.method !== "HEAD") {
        next();
        return;
      }

      let body: string;
      try {
        body = readFileSync(visualTestAccountFile, "utf8");
      } catch {
        next();
        return;
      }

      response.setHeader("Content-Type", "application/json; charset=utf-8");
      response.setHeader("Cache-Control", "no-store");
      response.end(request.method === "HEAD" ? undefined : body);
    });
  },
});

// Конфигурация подключает компилятор SolidJS для клиентского UI-слоя.
export default defineConfig({
  plugins: [solid(), visualTestAccountPlugin()],
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
