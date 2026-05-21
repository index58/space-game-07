package ws

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"space-game-07-server/internal/data"
	"space-game-07-server/internal/storage"
)

// РџСЂРѕРІРµСЂСЏРµС‚, С‡С‚Рѕ HTTP-РѕС‚РІРµС‚ СЃРїСЂР°РІРѕС‡РЅРёРєРѕРІ СЃРѕРґРµСЂР¶РёС‚ РІСЃРµ С‚Р°Р±Р»РёС†С‹, РЅРµРѕР±С…РѕРґРёРјС‹Рµ РєР»РёРµРЅС‚Сѓ РїСЂРё СЃС‚Р°СЂС‚Рµ.
func TestReferenceDataHandlerReturnsAllStartupTables(t *testing.T) {
	serverData := &storage.ServerData{
		CosmicObjectTypes: data.NewCosmicObjectTypes(),
		CosmicObjectModels: &data.CosmicObjectModels{
			MaxID: 23,
			Items: map[int64]*data.CosmicObjectModel{
				23: {
					ID:              23,
					TitleRu:         "РљРѕСЂР°Р±Р»СЊ",
					TitleEn:         "Ship",
					Acronym:         "ship_bat",
					TextureFilePath: "assets/world/cosmic-objects/ships/256x512/ship_256x512_0008.png",
					TextureScale:    4,
				},
			},
		},
		ItemTypes:           data.NewItemTypes(),
		NpcClans:            storage.NewRawReferenceTable(),
		ItemModels:          data.NewItemModels(),
		Blueprints:          storage.NewRawReferenceTable(),
		BlueprintComponents: storage.NewRawReferenceTable(),
		Schemas:             storage.NewRawReferenceTable(),
		SchemaComponents:    storage.NewRawReferenceTable(),
	}
	response := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodGet, "/reference-data", nil)

	NewReferenceDataHandler(serverData).ServeHTTP(response, request)

	if response.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", response.Code, http.StatusOK)
	}

	var payload map[string]json.RawMessage
	if err := json.Unmarshal(response.Body.Bytes(), &payload); err != nil {
		t.Fatalf("decode response: %v", err)
	}

	for _, field := range []string{
		"type",
		"NpcClan",
		"CosmicObjectType",
		"ItemType",
		"CosmicObjectModel",
		"ItemModel",
		"Blueprint",
		"BlueprintComponent",
		"Schema",
		"SchemaComponent",
	} {
		if _, ok := payload[field]; !ok {
			t.Fatalf("response does not contain %s", field)
		}
	}
}
