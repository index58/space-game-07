package ws

import (
	"encoding/json"

	"space-game-07-server/internal/game"
)

// ExchangeContainerMessage передает выбранный контейнер обмена.
type ExchangeContainerMessage struct {
	Type                      string `json:"type"`                      // Вид команды обмена.
	ContainerEquipmentGroupID int64  `json:"containerEquipmentGroupId"` // Выбранный контейнер.
}

// ExchangeAddItemsMessage передает строки предметов для добавления в очередь.
type ExchangeAddItemsMessage struct {
	Type         string  `json:"type"`         // Вид команды обмена.
	ItemGroupIDs []int64 `json:"itemGroupIds"` // Выбранные строки предметов.
	Amount       float64 `json:"amount"`       // Количество предметов из одной выбранной строки.
}

// DecodeExchangeSimpleMessage проверяет команду обмена без дополнительных данных.
func DecodeExchangeSimpleMessage(payload []byte, messageType string) bool {
	var message clientMessageType
	if err := json.Unmarshal(payload, &message); err != nil {
		return false
	}
	return message.Type == messageType
}

// DecodeExchangeContainerMessage проверяет команду выбора контейнера.
func DecodeExchangeContainerMessage(payload []byte, messageType string) (ExchangeContainerMessage, bool) {
	var message ExchangeContainerMessage
	if err := json.Unmarshal(payload, &message); err != nil {
		return ExchangeContainerMessage{}, false
	}
	if message.Type != messageType || message.ContainerEquipmentGroupID <= 0 {
		return ExchangeContainerMessage{}, false
	}
	return message, true
}

// DecodeExchangeAddItemsMessage проверяет команду добавления предметов в очередь.
func DecodeExchangeAddItemsMessage(payload []byte) (ExchangeAddItemsMessage, bool) {
	var message ExchangeAddItemsMessage
	if err := json.Unmarshal(payload, &message); err != nil {
		return ExchangeAddItemsMessage{}, false
	}
	if message.Type != "exchangeAddItems" || len(message.ItemGroupIDs) == 0 {
		return ExchangeAddItemsMessage{}, false
	}
	return message, true
}

// EncodeExchangeEventMessage кодирует событие обмена для клиента.
func EncodeExchangeEventMessage(event game.ExchangeEvent) ([]byte, error) {
	return json.Marshal(event)
}
