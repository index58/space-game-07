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
	Type                      string                           `json:"type"`                      // Тип сообщения для проверки клиентского контракта.
	NpcClan                   *storage.RawReferenceTable       `json:"NpcClan"`                   // Справочник NPC-кланов.
	CosmicObjectType          *data.CosmicObjectTypes          `json:"CosmicObjectType"`          // Справочник типов космических объектов.
	ItemType                  *data.ItemTypes                  `json:"ItemType"`                  // Справочник типов предметов.
	CosmicObjectModel         *data.CosmicObjectModels         `json:"CosmicObjectModel"`         // Справочник моделей космических объектов.
	ItemModel                 *data.ItemModels                 `json:"ItemModel"`                 // Справочник моделей предметов.
	TaskType                  *data.TaskTypes                  `json:"TaskType"`                  // Справочник типов заданий для интерфейса.
	Implementer               *data.Implementers               `json:"Implementer"`               // Справочник исполнителей заданий для расчета времени.
	Blueprint                 *storage.RawReferenceTable       `json:"Blueprint"`                 // Справочник чертежей объектов.
	BlueprintComponent        *storage.RawReferenceTable       `json:"BlueprintComponent"`        // Справочник компонентов чертежей.
	Schema                    *storage.RawReferenceTable       `json:"Schema"`                    // Справочник схем предметов.
	SchemaComponent           *storage.RawReferenceTable       `json:"SchemaComponent"`           // Справочник компонентов схем.
	ActionType                *data.ActionTypes                `json:"ActionType"`                // Справочник игровых действий.
	InputEventType            *data.InputEventTypes            `json:"InputEventType"`            // Справочник событий ввода.
	DefaultActionInputSetting *data.DefaultActionInputSettings `json:"DefaultActionInputSetting"` // Привязки ввода по умолчанию.
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
		Type:                      "referenceData",
		NpcClan:                   serverData.NpcClans,
		CosmicObjectType:          serverData.CosmicObjectTypes,
		ItemType:                  serverData.ItemTypes,
		CosmicObjectModel:         serverData.CosmicObjectModels,
		ItemModel:                 serverData.ItemModels,
		TaskType:                  serverData.TaskTypes,
		Implementer:               serverData.Implementers,
		Blueprint:                 serverData.Blueprints,
		BlueprintComponent:        serverData.BlueprintComponents,
		Schema:                    serverData.Schemas,
		SchemaComponent:           serverData.SchemaComponents,
		ActionType:                serverData.ActionTypes,
		InputEventType:            serverData.InputEventTypes,
		DefaultActionInputSetting: serverData.DefaultActionInputSettings,
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
