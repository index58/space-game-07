package ws

import (
	"encoding/json"

	"space-game-07-server/internal/game"
)

// Разбирает клиентский JSON и пропускает только сообщения управления кораблем.
func DecodeInputMessage(payload []byte) (game.ShipInput, bool) {
	var input game.ShipInput
	if err := json.Unmarshal(payload, &input); err != nil {
		return game.ShipInput{}, false
	}

	if input.Type != "input" {
		return game.ShipInput{}, false
	}

	return input, true
}

// Сериализует снимок мира в формат WebSocket-сообщения.
func EncodeSnapshotMessage(snapshot game.Snapshot) ([]byte, error) {
	return json.Marshal(snapshot)
}
