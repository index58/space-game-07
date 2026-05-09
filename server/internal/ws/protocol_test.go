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

// Проверяет, что запрос свежих настроек ввода принимается по согласованному типу.
func TestDecodeInputSettingsRequestMessageAcceptsAgreedType(t *testing.T) {
	if !transport.DecodeInputSettingsRequestMessage([]byte(`{"type":"inputSettingsRequest"}`)) {
		t.Fatalf("input settings request was not accepted")
	}
}

// Проверяет, что команда панели управления объектом читает идентификатор мутации и изменяемые поля.
func TestDecodeControlPanelObjectUpdateMessageUsesAgreedJSONFields(t *testing.T) {
	message, ok := transport.DecodeControlPanelObjectUpdateMessage([]byte(`{
		"type": "controlPanelObjectUpdate",
		"clientSessionId": "session-1",
		"mutationSeq": 7,
		"enabled": false,
		"title": "Новый корабль"
	}`))

	if !ok {
		t.Fatalf("control panel object update was not accepted")
	}
	if message.ClientSessionID != "session-1" || message.MutationSeq != 7 || message.Enabled == nil || *message.Enabled || message.Title == nil || *message.Title != "Новый корабль" {
		t.Fatalf("decoded object update mismatch: %+v", message)
	}
}

// Проверяет, что команда панели управления оборудованием читает идентификатор группы и значения.
func TestDecodeControlPanelEquipmentUpdateMessageUsesAgreedJSONFields(t *testing.T) {
	message, ok := transport.DecodeControlPanelEquipmentUpdateMessage([]byte(`{
		"type": "controlPanelEquipmentUpdate",
		"clientSessionId": "session-1",
		"mutationSeq": 8,
		"equipmentGroupId": 12,
		"enabled": true,
		"enabledCount": 3
	}`))

	if !ok {
		t.Fatalf("control panel equipment update was not accepted")
	}
	if message.ClientSessionID != "session-1" || message.MutationSeq != 8 || message.EquipmentGroupID != 12 || message.Enabled == nil || !*message.Enabled || message.EnabledCount == nil || *message.EnabledCount != 3 {
		t.Fatalf("decoded equipment update mismatch: %+v", message)
	}
}

// Проверяет, что команда чата читает выбранную вкладку и адресный ник из JSON.
func TestDecodeChatSendMessageUsesAgreedJSONFields(t *testing.T) {
	message, ok := transport.DecodeChatSendMessage([]byte(`{
		"type": "chatSend",
		"chatId": 7,
		"targetNickname": "Pilot2",
		"text": "hello"
	}`))

	if !ok {
		t.Fatalf("chat message was not accepted")
	}
	if message.ChatID != 7 || message.TargetNickname != "Pilot2" || message.Text != "hello" {
		t.Fatalf("decoded chat message mismatch: %+v", message)
	}
}

// Проверяет, что другие входящие типы не распознаются как отправка текста.
func TestDecodeChatSendMessageRejectsOtherTypes(t *testing.T) {
	if _, ok := transport.DecodeChatSendMessage([]byte(`{"type":"input","text":"hello"}`)); ok {
		t.Fatalf("input message was accepted as chat command")
	}
}

// Проверяет, что выбор вкладки чата читает ID выбранного чата из JSON.
func TestDecodeChatSelectMessageUsesAgreedJSONFields(t *testing.T) {
	message, ok := transport.DecodeChatSelectMessage([]byte(`{"type":"chatSelect","chatId":7}`))

	if !ok {
		t.Fatalf("chat selection was not accepted")
	}
	if message.ChatID != 7 {
		t.Fatalf("decoded chat selection mismatch: %+v", message)
	}
}

// Проверяет, что состояние чата кодируется с согласованными именами полей.
func TestEncodeChatStateMessageUsesAgreedCamelCaseFields(t *testing.T) {
	payload, err := transport.EncodeChatStateMessage(game.ChatState{
		Type:           "chatState",
		SelectedChatID: 3,
		Tabs: []game.ChatTab{
			{
				ChatID:               3,
				Title:                "Server",
				CommunityTypeAcronym: "Server",
				UnreadCount:          2,
				Messages: []game.ChatMessage{
					{ID: 9, ChatID: 3, SenderNickname: "Pilot1", Text: "hello"},
				},
			},
		},
	})
	if err != nil {
		t.Fatalf("encode chat state: %v", err)
	}

	jsonText := string(payload)
	for _, field := range []string{
		`"type":"chatState"`,
		`"selectedChatId":3`,
		`"chatId":3`,
		`"communityTypeAcronym":"Server"`,
		`"unreadCount":2`,
		`"senderNickname":"Pilot1"`,
	} {
		if !strings.Contains(jsonText, field) {
			t.Fatalf("encoded chat state %s does not contain %s", jsonText, field)
		}
	}
}

// Проверяет, что отказ команды чата возвращается отдельным сетевым типом.
func TestEncodeChatErrorMessageUsesAgreedFields(t *testing.T) {
	payload, err := transport.EncodeChatErrorMessage("Адресат не найден")
	if err != nil {
		t.Fatalf("encode chat error: %v", err)
	}

	jsonText := string(payload)
	for _, field := range []string{
		`"type":"chatError"`,
		`"message":"Адресат не найден"`,
	} {
		if !strings.Contains(jsonText, field) {
			t.Fatalf("encoded chat error %s does not contain %s", jsonText, field)
		}
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
		ClientMutationAck: &game.ClientMutationAck{
			SessionID:      "session-1",
			LastAppliedSeq: 8,
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
		`"clientMutationAck":{"sessionId":"session-1","lastAppliedSeq":8}`,
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
