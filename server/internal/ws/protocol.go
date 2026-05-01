package ws

import (
	"encoding/json"

	"space-game-07-server/internal/game"
)

type clientMessageType struct {
	Type string `json:"type"` // Вид входящего пакета для выбора серверного обработчика.
}

// Передает клиенту секрет, который нужно сохранить для следующих подключений.
type AuthMessage struct {
	Type  string `json:"type"`  // Вид пакета для отличия от снимков мира.
	Token string `json:"token"` // Секрет созданной учетной записи.
}

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

// Проверяет, что клиент прислал команду смены корабля.
func DecodeRandomShipMessage(payload []byte) bool {
	var message clientMessageType
	if err := json.Unmarshal(payload, &message); err != nil {
		return false
	}

	return message.Type == "randomShip"
}

// Сериализует снимок мира в формат WebSocket-сообщения.
func EncodeSnapshotMessage(snapshot game.Snapshot) ([]byte, error) {
	return json.Marshal(snapshot)
}

// Сериализует результат автоматической авторизации.
func EncodeAuthMessage(token string) ([]byte, error) {
	return json.Marshal(AuthMessage{
		Type:  "auth",
		Token: token,
	})
}
