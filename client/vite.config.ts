import { defineConfig } from "vite";
import solid from "vite-plugin-solid";

// Конфигурация подключает компилятор SolidJS для клиентского UI-слоя.
export default defineConfig({
  plugins: [solid()],
});
