package ws_test

import (
	"encoding/json"
	"strings"
	"testing"

	"space-game-07-server/internal/game"
	transport "space-game-07-server/internal/ws"
)

func TestDecodeInputMessageUsesAgreedJSONFields(t *testing.T) {
	input, ok := transport.DecodeInputMessage([]byte(`{
		"type": "input",
		"seq": 42,
		"thrustForward": true,
		"thrustBackward": false,
		"thrustLeft": false,
		"thrustRight": true,
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
		input.TargetRotationDelta != 0.0125 {
		t.Fatalf("decoded input mismatch: %+v", input)
	}
}

func TestDecodeInputMessageRejectsUnknownType(t *testing.T) {
	_, ok := transport.DecodeInputMessage([]byte(`{"type":"unknown"}`))

	if ok {
		t.Fatalf("unknown message type was accepted")
	}
}

func TestEncodeSnapshotMessageUsesAgreedCamelCaseFields(t *testing.T) {
	snapshot := game.Snapshot{
		Type:         "snapshot",
		Tick:         123,
		SelfObjectID: 7,
		Objects: []game.SnapshotObject{
			{
				ID:              7,
				ModelAcronym:    "ship_bat",
				Kind:            game.ObjectKindShip,
				TextureScale:    4,
				X:               10.5,
				Y:               -3.2,
				VelocityX:       1.1,
				VelocityY:       0.4,
				Rotation:        0.2,
				AngularVelocity: 0.01,
				TargetRotation:  0.25,
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
		`"modelAcronym":"ship_bat"`,
		`"textureScale":4`,
		`"velocityX":1.1`,
		`"velocityY":0.4`,
		`"angularVelocity":0.01`,
		`"targetRotation":0.25`,
	} {
		if !strings.Contains(jsonText, field) {
			t.Fatalf("encoded snapshot %s does not contain %s", jsonText, field)
		}
	}

	var decoded game.Snapshot
	if err := json.Unmarshal(payload, &decoded); err != nil {
		t.Fatalf("decode snapshot: %v", err)
	}
}
