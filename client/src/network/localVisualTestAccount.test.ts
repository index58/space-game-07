import { describe, expect, it } from "vitest";
import { installLocalVisualTestAccount } from "./localVisualTestAccount";

// Имитирует минимальный ответ загрузчика без реального HTTP-запроса.
const response = (ok: boolean, payload: unknown) => ({
  ok,
  json: async () => payload,
});

// Имитирует браузерное хранилище для проверки записи секрета.
const createStorage = () => {
  const values = new Map<string, string>();

  return {
    getItem: (key: string) => values.get(key) ?? null,
    setItem: (key: string, value: string) => values.set(key, value),
  };
};

describe("installLocalVisualTestAccount", () => {
  // Проверяет, что отсутствие локального файла не ломает обычный вход.
  it("keeps regular startup when local visual account file is absent", async () => {
    const storage = createStorage();

    const installed = await installLocalVisualTestAccount({
      fetcher: async () => response(false, null),
      storage,
    });

    expect(installed).toBe(false);
    expect(storage.getItem("accountToken")).toBeNull();
  });

  // Проверяет, что локальный тестовый секрет ставится до сетевого клиента.
  it("stores local visual account token for websocket identity", async () => {
    const storage = createStorage();

    const installed = await installLocalVisualTestAccount({
      fetcher: async () => response(true, { accountToken: "visual-token" }),
      storage,
    });

    expect(installed).toBe(true);
    expect(storage.getItem("accountToken")).toBe("visual-token");
  });

  // Проверяет, что поврежденный локальный файл не превращается в гостевое подключение.
  it("fails loudly when local visual account file has no token", async () => {
    const storage = createStorage();

    await expect(installLocalVisualTestAccount({
      fetcher: async () => response(true, { accountToken: "" }),
      storage,
    })).rejects.toThrow("Не задан accountToken");
  });
});
