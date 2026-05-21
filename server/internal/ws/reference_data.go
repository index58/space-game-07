package ws

import (
	"encoding/json"
	"net/http"
	"strings"

	"space-game-07-server/internal/data"
	"space-game-07-server/internal/storage"
)

// Р С›Р С—Р С‘РЎРѓРЎвЂ№Р Р†Р В°Р ВµРЎвЂљ Р С—Р В°Р С”Р ВµРЎвЂљ РЎРѓР С—РЎР‚Р В°Р Р†Р С•РЎвЂЎР Р…Р С‘Р С”Р С•Р Р†, Р С”Р С•РЎвЂљР С•РЎР‚РЎвЂ№Р в„– Р С”Р В»Р С‘Р ВµР Р…РЎвЂљ Р С—Р С•Р В»РЎС“РЎвЂЎР В°Р ВµРЎвЂљ Р С—Р ВµРЎР‚Р ВµР Т‘ Р С—Р С•Р Т‘Р С”Р В»РЎР‹РЎвЂЎР ВµР Р…Р С‘Р ВµР С Р С” Р С‘Р С–РЎР‚Р С•Р Р†Р С•Р СРЎС“ Р С—Р С•РЎвЂљР С•Р С”РЎС“.
type ReferenceDataResponse struct {
	Type                      string                           `json:"type"`                      // Р СћР С‘Р С— РЎРѓР С•Р С•Р В±РЎвЂ°Р ВµР Р…Р С‘РЎРЏ Р Т‘Р В»РЎРЏ Р С—РЎР‚Р С•Р Р†Р ВµРЎР‚Р С”Р С‘ Р С”Р В»Р С‘Р ВµР Р…РЎвЂљРЎРѓР С”Р С•Р С–Р С• Р С”Р С•Р Р…РЎвЂљРЎР‚Р В°Р С”РЎвЂљР В°.
	NpcClan                   *storage.RawReferenceTable       `json:"NpcClan"`                   // Р РЋР С—РЎР‚Р В°Р Р†Р С•РЎвЂЎР Р…Р С‘Р С” NPC-Р С”Р В»Р В°Р Р…Р С•Р Р†.
	CosmicObjectType          *data.CosmicObjectTypes          `json:"CosmicObjectType"`          // Р РЋР С—РЎР‚Р В°Р Р†Р С•РЎвЂЎР Р…Р С‘Р С” РЎвЂљР С‘Р С—Р С•Р Р† Р С”Р С•РЎРѓР СР С‘РЎвЂЎР ВµРЎРѓР С”Р С‘РЎвЂ¦ Р С•Р В±РЎР‰Р ВµР С”РЎвЂљР С•Р Р†.
	ItemType                  *data.ItemTypes                  `json:"ItemType"`                  // Р РЋР С—РЎР‚Р В°Р Р†Р С•РЎвЂЎР Р…Р С‘Р С” РЎвЂљР С‘Р С—Р С•Р Р† Р С—РЎР‚Р ВµР Т‘Р СР ВµРЎвЂљР С•Р Р†.
	CosmicObjectModel         *data.CosmicObjectModels         `json:"CosmicObjectModel"`         // Р РЋР С—РЎР‚Р В°Р Р†Р С•РЎвЂЎР Р…Р С‘Р С” Р СР С•Р Т‘Р ВµР В»Р ВµР в„– Р С”Р С•РЎРѓР СР С‘РЎвЂЎР ВµРЎРѓР С”Р С‘РЎвЂ¦ Р С•Р В±РЎР‰Р ВµР С”РЎвЂљР С•Р Р†.
	ItemModel                 *data.ItemModels                 `json:"ItemModel"`                 // Р РЋР С—РЎР‚Р В°Р Р†Р С•РЎвЂЎР Р…Р С‘Р С” Р СР С•Р Т‘Р ВµР В»Р ВµР в„– Р С—РЎР‚Р ВµР Т‘Р СР ВµРЎвЂљР С•Р Р†.
	TaskType                  *data.TaskTypes                  `json:"TaskType"`                  // Справочник типов заданий для интерфейса.
	Implementer               *data.Implementers               `json:"Implementer"`               // Справочник исполнителей заданий для расчета времени.
	Blueprint                 *storage.RawReferenceTable       `json:"Blueprint"`                 // Р РЋР С—РЎР‚Р В°Р Р†Р С•РЎвЂЎР Р…Р С‘Р С” РЎвЂЎР ВµРЎР‚РЎвЂљР ВµР В¶Р ВµР в„– Р С•Р В±РЎР‰Р ВµР С”РЎвЂљР С•Р Р†.
	BlueprintComponent        *storage.RawReferenceTable       `json:"BlueprintComponent"`        // Р РЋР С—РЎР‚Р В°Р Р†Р С•РЎвЂЎР Р…Р С‘Р С” Р С”Р С•Р СР С—Р С•Р Р…Р ВµР Р…РЎвЂљР С•Р Р† РЎвЂЎР ВµРЎР‚РЎвЂљР ВµР В¶Р ВµР в„–.
	Schema                    *storage.RawReferenceTable       `json:"Schema"`                    // Р РЋР С—РЎР‚Р В°Р Р†Р С•РЎвЂЎР Р…Р С‘Р С” РЎРѓРЎвЂ¦Р ВµР С Р С—РЎР‚Р ВµР Т‘Р СР ВµРЎвЂљР С•Р Р†.
	SchemaComponent           *storage.RawReferenceTable       `json:"SchemaComponent"`           // Р РЋР С—РЎР‚Р В°Р Р†Р С•РЎвЂЎР Р…Р С‘Р С” Р С”Р С•Р СР С—Р С•Р Р…Р ВµР Р…РЎвЂљР С•Р Р† РЎРѓРЎвЂ¦Р ВµР С.
	ActionType                *data.ActionTypes                `json:"ActionType"`                // Р РЋР С—РЎР‚Р В°Р Р†Р С•РЎвЂЎР Р…Р С‘Р С” Р С‘Р С–РЎР‚Р С•Р Р†РЎвЂ№РЎвЂ¦ Р Т‘Р ВµР в„–РЎРѓРЎвЂљР Р†Р С‘Р в„–.
	InputEventType            *data.InputEventTypes            `json:"InputEventType"`            // Р РЋР С—РЎР‚Р В°Р Р†Р С•РЎвЂЎР Р…Р С‘Р С” РЎРѓР С•Р В±РЎвЂ№РЎвЂљР С‘Р в„– Р Р†Р Р†Р С•Р Т‘Р В°.
	DefaultActionInputSetting *data.DefaultActionInputSettings `json:"DefaultActionInputSetting"` // Р СџРЎР‚Р С‘Р Р†РЎРЏР В·Р С”Р С‘ Р Р†Р Р†Р С•Р Т‘Р В° Р С—Р С• РЎС“Р СР С•Р В»РЎвЂЎР В°Р Р…Р С‘РЎР‹.
}

// Р РЋР С•Р В·Р Т‘Р В°Р ВµРЎвЂљ HTTP-Р С•Р В±РЎР‚Р В°Р В±Р С•РЎвЂљРЎвЂЎР С‘Р С” Р Т‘Р В»РЎРЏ Р Р†РЎвЂ№Р Т‘Р В°РЎвЂЎР С‘ Р Р†РЎРѓР ВµРЎвЂ¦ РЎРѓР С—РЎР‚Р В°Р Р†Р С•РЎвЂЎР Р…Р С‘Р С”Р С•Р Р† Р С•Р Т‘Р Р…Р С‘Р С Р В·Р В°Р С—РЎР‚Р С•РЎРѓР С•Р С Р С—РЎР‚Р С‘ Р Р†РЎвЂ¦Р С•Р Т‘Р Вµ Р С”Р В»Р С‘Р ВµР Р…РЎвЂљР В°.
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

// Р РЋР С•Р В±Р С‘РЎР‚Р В°Р ВµРЎвЂљ Р С•РЎвЂљР Р†Р ВµРЎвЂљ Р В±Р ВµР В· Р С”Р В»Р С‘Р ВµР Р…РЎвЂљРЎРѓР С”Р С‘РЎвЂ¦ РЎвЂљР В°Р В±Р В»Р С‘РЎвЂ  Р С‘ Р В±Р ВµР В· РЎРѓР ВµРЎР‚Р Р†Р ВµРЎР‚Р Р…РЎвЂ№РЎвЂ¦ Р С‘Р Р…Р Т‘Р ВµР С”РЎРѓР С•Р Р†, РЎРѓР С”РЎР‚РЎвЂ№РЎвЂљРЎвЂ№РЎвЂ¦ JSON-РЎвЂљР ВµР С–Р В°Р СР С‘.
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

// Р В Р В°Р В·РЎР‚Р ВµРЎв‚¬Р В°Р ВµРЎвЂљ Р В±РЎР‚Р В°РЎС“Р В·Р ВµРЎР‚Р Р…Р С•Р СРЎС“ Р С”Р В»Р С‘Р ВµР Р…РЎвЂљРЎС“ РЎвЂЎР С‘РЎвЂљР В°РЎвЂљРЎРЉ РЎРѓР С—РЎР‚Р В°Р Р†Р С•РЎвЂЎР Р…Р С‘Р С”Р С‘ РЎРѓ Р В»Р С•Р С”Р В°Р В»РЎРЉР Р…Р С•Р С–Р С• HTTP-РЎРѓР ВµРЎР‚Р Р†Р ВµРЎР‚Р В°.
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
