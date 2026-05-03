package ws_test

import (
	"encoding/json"
	"strings"
	"testing"

	"space-game-07-server/internal/data"
	"space-game-07-server/internal/game"
	transport "space-game-07-server/internal/ws"
)

// Проверяет, что входное сообщение управления читается из согласованных JSON-полей.
func TestDecodeInputMessageUsesAgreedJSONFields(t *testing.T) {
	input, ok := transport.DecodeInputMessage([]byte(`{
		"type": "input",
		"seq": 42,
		"thrustForward": true,
		"thrustBackward": false,
		"thrustLeft": false,
		"thrustRight": true,
		"toggleAnchor": true,
		"targetRotationDelta": 0.0125
	}`))

	if !ok {
		t.Fatalf("input message was not accepted")
	}
	if input.Seq != 42 ||
		!input.ThrustForward ||
		input.ThrustBackward ||
		input.ThrustLeft ||
		!input.ThrustRight ||
		!input.ToggleAnchor ||
		input.TargetRotationDelta != 0.0125 {
		t.Fatalf("decoded input mismatch: %+v", input)
	}
}

// Проверяет, что сообщение с неподдерживаемым типом не принимается как управление.
func TestDecodeInputMessageRejectsUnknownType(t *testing.T) {
	_, ok := transport.DecodeInputMessage([]byte(`{"type":"unknown"}`))

	if ok {
		t.Fatalf("unknown message type was accepted")
	}
}

// Проверяет, что команда случайной смены корабля принимается по согласованному типу.
func TestDecodeRandomShipMessageAcceptsAgreedType(t *testing.T) {
	if !transport.DecodeRandomShipMessage([]byte(`{"type":"randomShip"}`)) {
		t.Fatalf("random ship message was not accepted")
	}
}

// Проверяет, что другие типы сообщений не распознаются как команда смены корабля.
func TestDecodeRandomShipMessageRejectsOtherTypes(t *testing.T) {
	if transport.DecodeRandomShipMessage([]byte(`{"type":"input"}`)) {
		t.Fatalf("input message was accepted as random ship command")
	}
}

// Проверяет, что снимок мира кодируется с текущими именами полей и без удалённых полей.
func TestEncodeSnapshotMessageUsesAgreedCamelCaseFields(t *testing.T) {
	snapshot := game.Snapshot{
		Type:         "snapshot",
		Tick:         123,
		SelfObjectID: 7,
		Objects: []data.CosmicObject{
			{
				ID:                  7,
				CosmicObjectModelID: 1,
				X:                   10.5,
				Y:                   -3.2,
				VelocityX:           1.1,
				VelocityY:           0.4,
				Rotation:            0.2,
				AngularSpeed:        0.01,
				TargetRotation:      0.25,
			},
		},
		EquipmentGroups: []data.EquipmentGroup{
			{
				ID:                   3,
				CosmicObjectID:       7,
				EquipmentItemModelID: 101,
				EnabledCount:         2,
			},
		},
	}

	payload, err := transport.EncodeSnapshotMessage(snapshot)
	if err != nil {
		t.Fatalf("encode snapshot: %v", err)
	}
	jsonText := string(payload)
	for _, field := range []string{
		`"selfObjectId":7`,
		`"CosmicObjectModelID":1`,
		`"VelocityX":1.1`,
		`"VelocityY":0.4`,
		`"AngularSpeed":0.01`,
		`"TargetRotation":0.25`,
		`"equipmentGroups":[`,
		`"EquipmentItemModelID":101`,
		`"EnabledCount":2`,
	} {
		if !strings.Contains(jsonText, field) {
			t.Fatalf("encoded snapshot %s does not contain %s", jsonText, field)
		}
	}
	for _, removedField := range []string{
		`"modelAcronym"`,
		`"textureScale"`,
		`"angularVelocity"`,
		`"targetRotation"`,
	} {
		if strings.Contains(jsonText, removedField) {
			t.Fatalf("encoded snapshot %s still contains removed field %s", jsonText, removedField)
		}
	}

	var decoded game.Snapshot
	if err := json.Unmarshal(payload, &decoded); err != nil {
		t.Fatalf("decode snapshot: %v", err)
	}
}

// Проверяет, что сообщение авторизации содержит тип и токен для отправки до снимков мира.
func TestEncodeAuthMessageSendsTokenBeforeSnapshots(t *testing.T) {
	payload, err := transport.EncodeAuthMessage("secret-token")
	if err != nil {
		t.Fatalf("encode auth: %v", err)
	}

	jsonText := string(payload)
	for _, field := range []string{
		`"type":"auth"`,
		`"token":"secret-token"`,
	} {
		if !strings.Contains(jsonText, field) {
			t.Fatalf("encoded auth %s does not contain %s", jsonText, field)
		}
	}
}
