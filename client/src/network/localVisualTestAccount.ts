const DEFAULT_ACCOUNT_PATH = "/local/visual-test-account.local.json";

type LocalVisualTestAccountConfig = {
  // Секрет существующего аккаунта для локальных визуальных проверок.
  accountToken: string;
};

export type LocalVisualTestAccountFetchResponse = {
  // Признак успешной загрузки локального файла.
  ok: boolean;
  // Возвращает разобранное тело локального файла.
  json(): Promise<unknown>;
};

export type LocalVisualTestAccountFetch = (
  input: string,
  init?: RequestInit,
) => Promise<LocalVisualTestAccountFetchResponse>;

export type InstallLocalVisualTestAccountOptions = {
  // URL, по которому dev-сервер отдает локальный файл.
  accountPath?: string;
  // Загрузчик, подменяемый в тестах без реального браузерного запроса.
  fetcher?: LocalVisualTestAccountFetch;
  // Хранилище, куда клиент уже умеет складывать секрет аккаунта.
  storage?: Pick<Storage, "setItem">;
};

const getDefaultFetcher = (): LocalVisualTestAccountFetch | null => {
  if (typeof fetch === "undefined") {
    return null;
  }

  return fetch;
};

const getDefaultStorage = (): Pick<Storage, "setItem"> | null => {
  if (typeof localStorage === "undefined") {
    return null;
  }

  return localStorage;
};

const parseConfig = (value: unknown): LocalVisualTestAccountConfig => {
  if (!value || typeof value !== "object") {
    throw new Error("Не задан accountToken для визуального тестового аккаунта.");
  }

  const token = (value as Partial<LocalVisualTestAccountConfig>).accountToken;
  if (typeof token !== "string" || token.trim() === "") {
    throw new Error("Не задан accountToken для визуального тестового аккаунта.");
  }

  return { accountToken: token.trim() };
};

// Подкладывает локальный секрет до создания WebSocket, если dev-сервер отдал файл.
export const installLocalVisualTestAccount = async (
  options: InstallLocalVisualTestAccountOptions = {},
): Promise<boolean> => {
  const fetcher = options.fetcher ?? getDefaultFetcher();
  const storage = options.storage ?? getDefaultStorage();

  if (!fetcher || !storage) {
    return false;
  }

  let response: LocalVisualTestAccountFetchResponse;
  try {
    response = await fetcher(options.accountPath ?? DEFAULT_ACCOUNT_PATH, {
      cache: "no-store",
    });
  } catch {
    return false;
  }

  if (!response.ok) {
    return false;
  }

  const config = parseConfig(await response.json());
  storage.setItem("accountToken", config.accountToken);
  return true;
};
