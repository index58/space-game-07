import type { InputSettingPayload, ReferenceDataMessage } from "../network/protocol";
import type { UiKitOption } from "../ui-kit/components";

export type InputSettingsRow = {
  // Действие, которое отображается в левой колонке.
  actionTypeId: number;
  // Акроним действия для связи с игровой логикой.
  actionAcronym: string;
  // Видимое название действия.
  actionTitle: string;
  // Выбранное событие ввода.
  inputEventTypeId: number;
};

export type InputBindingMap = Record<string, string>;

// Возвращает число строк в левой половине окна настроек с лишней строкой слева.
export const getInputSettingsLeftColumnRowCount = (rowCount: number): number => Math.ceil(Math.max(0, rowCount) / 2);

// Возвращает настройки по умолчанию как карту действия к событию.
export const getDefaultInputSettingValues = (referenceData: ReferenceDataMessage | null): Record<number, number> => {
  if (!referenceData) {
    return {};
  }

  const values: Record<number, number> = {};
  for (const setting of Object.values(referenceData.DefaultActionInputSetting.Items)) {
    values[setting.ActionTypeID] = setting.InputEventTypeID;
  }
  return values;
};

// Накладывает аккаунтные настройки поверх серверных значений по умолчанию.
export const getMergedInputSettingValues = (
  referenceData: ReferenceDataMessage | null,
  accountSettings: InputSettingPayload[],
): Record<number, number> => {
  const values = getDefaultInputSettingValues(referenceData);
  for (const setting of accountSettings) {
    values[setting.actionTypeId] = setting.inputEventTypeId;
  }
  return values;
};

// Собирает строки вкладки ввода в стабильном порядке справочника действий.
export const getInputSettingsRows = (
  referenceData: ReferenceDataMessage | null,
  values: Record<number, number>,
): InputSettingsRow[] => {
  if (!referenceData) {
    return [];
  }

  return Object.values(referenceData.ActionType.Items)
    .sort((left, right) => left.ID - right.ID)
    .map((action) => ({
      actionTypeId: action.ID,
      actionAcronym: action.Acronym,
      actionTitle: action.TitleRu,
      inputEventTypeId: values[action.ID] ?? 0,
    }));
};

// Собирает пункты выпадающего списка в стабильном порядке справочника событий.
export const getInputEventOptions = (referenceData: ReferenceDataMessage | null): UiKitOption[] => {
  if (!referenceData) {
    return [];
  }

  return Object.values(referenceData.InputEventType.Items)
    .sort((left, right) => left.TitleRu.localeCompare(right.TitleRu, "ru"))
    .map((eventType) => ({
      value: String(eventType.ID),
      label: eventType.TitleRu,
    }));
};

// Преобразует выбранные ID событий в системные строки для обработчика ввода.
export const getInputBindingMap = (
  referenceData: ReferenceDataMessage | null,
  values: Record<number, number>,
): InputBindingMap => {
  if (!referenceData) {
    return {};
  }

  const result: InputBindingMap = {};
  for (const action of Object.values(referenceData.ActionType.Items)) {
    const eventID = values[action.ID];
    const eventType = referenceData.InputEventType.Items[String(eventID)];
    if (eventType) {
      result[action.Acronym] = eventType.SystemStringValue;
    }
  }
  return result;
};

// Готовит полный список настроек для отправки на сервер.
export const toInputSettingsPayload = (values: Record<number, number>): InputSettingPayload[] => Object.entries(values)
  .map(([actionTypeId, inputEventTypeId]) => ({
    actionTypeId: Number(actionTypeId),
    inputEventTypeId,
  }))
  .filter((setting) => setting.actionTypeId > 0 && setting.inputEventTypeId > 0)
  .sort((left, right) => left.actionTypeId - right.actionTypeId);
