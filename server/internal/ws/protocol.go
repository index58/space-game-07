package ws

import (
	"encoding/json"

	"space-game-07-server/internal/game"
)

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

func EncodeSnapshotMessage(snapshot game.Snapshot) ([]byte, error) {
	return json.Marshal(snapshot)
}
