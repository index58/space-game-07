import type { ReferenceDataMessage, ReferenceTable } from "./protocol";

// Настраивает загрузку справочников для тестов и нестандартных адресов сервера.
export type ReferenceDataOptions = {
  // HTTP-адрес единого пакета справочников.
  url?: string;
  // Фабрика HTTP-запросов, позволяющая тестам не обращаться к сети.
  fetchFactory?: typeof fetch;
};

const DEFAULT_REFERENCE_DATA_URL = "http://127.0.0.1:8080/reference-data";

// Проверяет общий контейнер таблицы без знания предметной структуры записей.
const isReferenceTable = (value: unknown): value is ReferenceTable => {
  if (!value || typeof value !== "object") {
    return false;
  }

  const table = value as ReferenceTable;
  return typeof table.MaxID === "number" &&
    !!table.Items &&
    typeof table.Items === "object" &&
    !Array.isArray(table.Items);
};

// Проверяет, что сервер вернул все справочники, необходимые клиенту сразу после входа.
const isReferenceDataMessage = (value: unknown): value is ReferenceDataMessage => {
  if (!value || typeof value !== "object") {
    return false;
  }

  const message = value as ReferenceDataMessage;
  return message.type === "referenceData" &&
    isReferenceTable(message.NpcClan) &&
    isReferenceTable(message.CosmicObjectType) &&
    isReferenceTable(message.ItemType) &&
    isReferenceTable(message.CosmicObjectModel) &&
    isReferenceTable(message.ItemModel) &&
    (!message.TaskType || isReferenceTable(message.TaskType)) &&
    (!message.Implementer || isReferenceTable(message.Implementer)) &&
    isReferenceTable(message.Blueprint) &&
    isReferenceTable(message.BlueprintComponent) &&
    isReferenceTable(message.Schema) &&
    isReferenceTable(message.SchemaComponent) &&
    isReferenceTable(message.ActionType) &&
    isReferenceTable(message.InputEventType) &&
    isReferenceTable(message.DefaultActionInputSetting);
};

// Загружает справочники с сервера и не допускает запуск клиента с неполным пакетом данных.
export const fetchReferenceData = async (
  options: ReferenceDataOptions = {},
): Promise<ReferenceDataMessage> => {
  const response = await (options.fetchFactory ?? fetch)(options.url ?? DEFAULT_REFERENCE_DATA_URL);
  if (!response.ok) {
    throw new Error(`reference data request failed: ${response.status}`);
  }

  const payload: unknown = await response.json();
  if (!isReferenceDataMessage(payload)) {
    throw new Error("reference data response has invalid shape");
  }

  return payload;
};
