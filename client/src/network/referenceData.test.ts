import { describe, expect, it, vi } from "vitest";
import { fetchReferenceData } from "./referenceData";

// Создает пустую таблицу справочника с тем же контейнером, что приходит с сервера.
const emptyTable = () => ({ MaxID: 0, Items: {} });

// Собирает полный пакет справочников, чтобы тест менял только проверяемую часть.
const referenceDataPayload = () => ({
  type: "referenceData",
  NpcClan: emptyTable(),
  CosmicObjectType: emptyTable(),
  ItemType: emptyTable(),
  CosmicObjectModel: {
    MaxID: 23,
    Items: {
      "23": {
        ID: 23,
        TextureFilePath: "assets/world/cosmic-objects/ships/256x512/ship_256x512_0008.png",
        TextureScale: 4,
      },
    },
  },
  ItemModel: emptyTable(),
  Blueprint: emptyTable(),
  BlueprintComponent: emptyTable(),
  Schema: emptyTable(),
  SchemaComponent: emptyTable(),
  ActionType: emptyTable(),
  InputEventType: emptyTable(),
  DefaultActionInputSetting: emptyTable(),
});

describe("fetchReferenceData", () => {
  it("loads all reference tables from the server", async () => {
    const fetchFactory = vi.fn(async () => ({
      ok: true,
      status: 200,
      json: async () => referenceDataPayload(),
    } as Response));

    const referenceData = await fetchReferenceData({ fetchFactory });

    expect(fetchFactory).toHaveBeenCalledWith("http://127.0.0.1:8080/reference-data");
    expect(referenceData.CosmicObjectModel.Items["23"].TextureFilePath).toBe(
      "assets/world/cosmic-objects/ships/256x512/ship_256x512_0008.png",
    );
    expect(referenceData.SchemaComponent.Items).toEqual({});
  });

  it("rejects a response without the required tables", async () => {
    const fetchFactory = vi.fn(async () => ({
      ok: true,
      status: 200,
      json: async () => ({ type: "referenceData" }),
    } as Response));

    await expect(fetchReferenceData({ fetchFactory })).rejects.toThrow("invalid shape");
  });
});
