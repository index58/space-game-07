package ws

import (
	"encoding/json"
	"net/http"
	"strings"

	"space-game-07-server/internal/data"
	"space-game-07-server/internal/storage"
)

// Описывает пакет справочников, который клиент получает перед подключением к игровому потоку.
type ReferenceDataResponse struct {
	Type               string                     `json:"type"`               // Тип сообщения для проверки клиентского контракта.
	NpcClan            *storage.RawReferenceTable `json:"NpcClan"`            // Справочник NPC-кланов.
	CosmicObjectType   *data.CosmicObjectTypes    `json:"CosmicObjectType"`   // Справочник типов космических объектов.
	Itemtype           *data.Itemtypes            `json:"Itemtype"`           // Справочник типов предметов.
	CosmicObjectModel  *data.CosmicObjectModels   `json:"CosmicObjectModel"`  // Справочник моделей космических объектов.
	ItemModel          *storage.RawReferenceTable `json:"ItemModel"`          // Справочник моделей предметов.
	Blueprint          *storage.RawReferenceTable `json:"Blueprint"`          // Справочник чертежей объектов.
	BlueprintComponent *storage.RawReferenceTable `json:"BlueprintComponent"` // Справочник компонентов чертежей.
	Schema             *storage.RawReferenceTable `json:"Schema"`             // Справочник схем предметов.
	SchemaComponent    *storage.RawReferenceTable `json:"SchemaComponent"`    // Справочник компонентов схем.
}

// Создает HTTP-обработчик для выдачи всех справочников одним запросом при входе клиента.
func NewReferenceDataHandler(serverData *storage.ServerData) http.Handler {
	return http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		setLocalCORSHeaders(writer, request)
		if request.Method == http.MethodOptions {
			writer.WriteHeader(http.StatusNoContent)
			return
		}
		if request.Method != http.MethodGet {
			http.Error(writer, "method not allowed", http.StatusMethodNotAllowed)
			return
		}

		writer.Header().Set("Content-Type", "application/json; charset=utf-8")
		if err := json.NewEncoder(writer).Encode(NewReferenceDataResponse(serverData)); err != nil {
			http.Error(writer, "reference data encoding failed", http.StatusInternalServerError)
		}
	})
}

// Собирает ответ без клиентских таблиц и без серверных индексов, скрытых JSON-тегами.
func NewReferenceDataResponse(serverData *storage.ServerData) ReferenceDataResponse {
	return ReferenceDataResponse{
		Type:               "referenceData",
		NpcClan:            serverData.NpcClans,
		CosmicObjectType:   serverData.CosmicObjectTypes,
		Itemtype:           serverData.Itemtypes,
		CosmicObjectModel:  serverData.CosmicObjectModels,
		ItemModel:          serverData.ItemModels,
		Blueprint:          serverData.Blueprints,
		BlueprintComponent: serverData.BlueprintComponents,
		Schema:             serverData.Schemas,
		SchemaComponent:    serverData.SchemaComponents,
	}
}

// Разрешает браузерному клиенту читать справочники с локального HTTP-сервера.
func setLocalCORSHeaders(writer http.ResponseWriter, request *http.Request) {
	origin := request.Header.Get("Origin")
	if origin == "" ||
		strings.HasPrefix(origin, "http://127.0.0.1:") ||
		strings.HasPrefix(origin, "http://localhost:") {
		if origin == "" {
			writer.Header().Set("Access-Control-Allow-Origin", "*")
		} else {
			writer.Header().Set("Access-Control-Allow-Origin", origin)
		}
	}
	writer.Header().Set("Access-Control-Allow-Methods", "GET, OPTIONS")
}
