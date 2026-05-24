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
		"primaryPointerAction": true,
		"selectedPilotToolIndex": 3,
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
		!input.PrimaryPointerAction ||
		input.SelectedPilotToolIndex != 3 ||
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

// Проверяет, что команда отправки запроса стыковки принимается отдельным WebSocket-сообщением.
// Проверяет, что тестовая команда присвоения объекта принимается по согласованному типу.
func TestDecodeTestClaimFocusedObjectOwnerMessageAcceptsAgreedType(t *testing.T) {
	if !transport.DecodeTestClaimFocusedObjectOwnerMessage([]byte(`{"type":"testClaimFocusedObjectOwner"}`)) {
		t.Fatalf("test owner claim message was not accepted")
	}
}

func TestDecodeDockingRequestMessageAcceptsAgreedType(t *testing.T) {
	if !transport.DecodeDockingRequestMessage([]byte(`{"type":"dockingRequest"}`)) {
		t.Fatalf("docking request message was not accepted")
	}
}

// Проверяет, что команда одобрения запроса стыковки принимается без идентификатора запроса.
func TestDecodeDockingApproveMessageAcceptsAgreedType(t *testing.T) {
	if !transport.DecodeDockingApproveMessage([]byte(`{"type":"dockingApprove"}`)) {
		t.Fatalf("docking approve message was not accepted")
	}
}

// Проверяет, что команда отказа запроса стыковки принимается без идентификатора запроса.
func TestDecodeDockingRejectMessageAcceptsAgreedType(t *testing.T) {
	if !transport.DecodeDockingRejectMessage([]byte(`{"type":"dockingReject"}`)) {
		t.Fatalf("docking reject message was not accepted")
	}
}

// Проверяет, что команда отстыковки применяется к текущему объекту без дополнительных параметров.
func TestDecodeDockingUndockMessageAcceptsAgreedType(t *testing.T) {
	if !transport.DecodeDockingUndockMessage([]byte(`{"type":"dockingUndock"}`)) {
		t.Fatalf("docking undock message was not accepted")
	}
}

// Проверяет, что команда начала пересадки принимается без дополнительных полей.
func TestDecodeLandingBeginMessageAcceptsAgreedType(t *testing.T) {
	if !transport.DecodeLandingBeginMessage([]byte(`{"type":"landingBegin"}`)) {
		t.Fatalf("landing begin message was not accepted")
	}
}

// Проверяет, что команда одобрения посадки принимается без идентификатора запроса.
func TestDecodeLandingApproveMessageAcceptsAgreedType(t *testing.T) {
	if !transport.DecodeLandingApproveMessage([]byte(`{"type":"landingApprove"}`)) {
		t.Fatalf("landing approve message was not accepted")
	}
}

// Проверяет, что команда отказа посадки принимается без идентификатора запроса.
func TestDecodeLandingRejectMessageAcceptsAgreedType(t *testing.T) {
	if !transport.DecodeLandingRejectMessage([]byte(`{"type":"landingReject"}`)) {
		t.Fatalf("landing reject message was not accepted")
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
		"title": "Renamed equipment",
		"enabled": true,
		"enabledCount": 3
	}`))

	if !ok {
		t.Fatalf("control panel equipment update was not accepted")
	}
	if message.ClientSessionID != "session-1" || message.MutationSeq != 8 || message.EquipmentGroupID != 12 || message.Title == nil || *message.Title != "Renamed equipment" || message.Enabled == nil || !*message.Enabled || message.EnabledCount == nil || *message.EnabledCount != 3 {
		t.Fatalf("decoded equipment update mismatch: %+v", message)
	}
}

// Проверяет, что команда перемещения содержимого контейнера читается из согласованных JSON-полей.
func TestDecodeControlPanelContainerTransferMessageUsesAgreedJSONFields(t *testing.T) {
	message, ok := transport.DecodeControlPanelContainerTransferMessage([]byte(`{
		"type": "controlPanelContainerTransfer",
		"clientSessionId": "session-1",
		"mutationSeq": 5,
		"controllerEquipmentGroupId": 12,
		"leftToRightDirection": true,
		"sourceContainerEquipmentGroupId": 11,
		"targetContainerEquipmentGroupId": 12,
		"itemGroupIds": [21, 22]
	}`))

	if !ok {
		t.Fatalf("container transfer message was not decoded")
	}
	if message.ClientSessionID != "session-1" || message.MutationSeq != 5 {
		t.Fatalf("mutation fields were decoded incorrectly: %+v", message)
	}
	if message.SourceContainerEquipmentGroupID != 11 || message.TargetContainerEquipmentGroupID != 12 {
		t.Fatalf("container fields were decoded incorrectly: %+v", message)
	}
	if message.ControllerEquipmentGroupID != 12 || !message.LeftToRightDirection {
		t.Fatalf("movement controller fields were decoded incorrectly: %+v", message)
	}
	if len(message.ItemGroupIDs) != 2 || message.ItemGroupIDs[0] != 21 || message.ItemGroupIDs[1] != 22 {
		t.Fatalf("item group fields were decoded incorrectly: %+v", message)
	}
}

// Проверяет, что команда добавления в обмен читает выбранное количество.
func TestDecodeExchangeAddItemsMessageUsesAmount(t *testing.T) {
	message, ok := transport.DecodeExchangeAddItemsMessage([]byte(`{
		"type": "exchangeAddItems",
		"itemGroupIds": [21],
		"amount": 4
	}`))

	if !ok {
		t.Fatalf("exchange add items message was not decoded")
	}
	if len(message.ItemGroupIDs) != 1 || message.ItemGroupIDs[0] != 21 || message.Amount != 4 {
		t.Fatalf("exchange add items fields were decoded incorrectly: %+v", message)
	}
}

// Проверяет, что команда чата читает выбранную вкладку и адресный ник из JSON.
// Проверяет, что команда переноса топлива читается из согласованных JSON-полей.
func TestDecodeControlPanelFuelTransferMessageUsesAgreedJSONFields(t *testing.T) {
	message, ok := transport.DecodeControlPanelFuelTransferMessage([]byte(`{
		"type": "controlPanelFuelTransfer",
		"clientSessionId": "session-1",
		"mutationSeq": 6,
		"containerEquipmentGroupId": 11,
		"fuelTankEquipmentGroupId": 12,
		"itemGroupIds": [21],
		"amount": 15
	}`))

	if !ok {
		t.Fatalf("fuel transfer message was not decoded")
	}
	if message.ClientSessionID != "session-1" || message.MutationSeq != 6 {
		t.Fatalf("mutation fields were decoded incorrectly: %+v", message)
	}
	if message.ContainerEquipmentGroupID != 11 || message.FuelTankEquipmentGroupID != 12 || message.Amount != 15 {
		t.Fatalf("fuel transfer fields were decoded incorrectly: %+v", message)
	}
	if len(message.ItemGroupIDs) != 1 || message.ItemGroupIDs[0] != 21 {
		t.Fatalf("item group fields were decoded incorrectly: %+v", message)
	}
}

// Проверяет, что команда изготовления предмета читает выбранные группы оборудования и схему из согласованных JSON-полей.
// Проверяет, что команда деконструкции предметов читает выбранные группы оборудования и предметов из согласованных JSON-полей.
func TestDecodeControlPanelItemDeconstructionMessageUsesAgreedJSONFields(t *testing.T) {
	message, ok := transport.DecodeControlPanelItemDeconstructionMessage([]byte(`{
		"type": "controlPanelItemDeconstruction",
		"clientSessionId": "session-1",
		"mutationSeq": 10,
		"deconstructorEquipmentGroupId": 13,
		"sourceContainerEquipmentGroupId": 11,
		"targetContainerEquipmentGroupId": 12,
		"itemGroupIds": [21, 22],
		"amount": 4
	}`))

	if !ok {
		t.Fatalf("item deconstruction message was not decoded")
	}
	if message.ClientSessionID != "session-1" || message.MutationSeq != 10 {
		t.Fatalf("mutation fields were decoded incorrectly: %+v", message)
	}
	if message.DeconstructorEquipmentGroupID != 13 || message.SourceContainerEquipmentGroupID != 11 || message.TargetContainerEquipmentGroupID != 12 || message.Amount != 4 {
		t.Fatalf("item deconstruction fields were decoded incorrectly: %+v", message)
	}
	if len(message.ItemGroupIDs) != 2 || message.ItemGroupIDs[0] != 21 || message.ItemGroupIDs[1] != 22 {
		t.Fatalf("item group fields were decoded incorrectly: %+v", message)
	}
}

func TestDecodeControlPanelConstructorProduceItemMessageUsesAgreedJSONFields(t *testing.T) {
	message, ok := transport.DecodeControlPanelConstructorProduceItemMessage([]byte(`{
		"type": "controlPanelConstructorProduceItem",
		"clientSessionId": "session-1",
		"mutationSeq": 7,
		"constructorEquipmentGroupId": 13,
		"materialContainerEquipmentGroupId": 11,
		"productContainerEquipmentGroupId": 12,
		"schemaId": 21,
		"amount": 4
	}`))

	if !ok {
		t.Fatalf("constructor production message was not decoded")
	}
	if message.ClientSessionID != "session-1" || message.MutationSeq != 7 {
		t.Fatalf("mutation fields were decoded incorrectly: %+v", message)
	}
	if message.ConstructorEquipmentGroupID != 13 || message.MaterialContainerEquipmentGroupID != 11 || message.ProductContainerEquipmentGroupID != 12 || message.SchemaID != 21 || message.Amount != 4 {
		t.Fatalf("constructor production fields were decoded incorrectly: %+v", message)
	}
}

// Проверяет, что команда изготовления объекта читает чертёж из согласованных JSON-полей.
func TestDecodeControlPanelConstructorProduceItemMessageAcceptsBlueprintID(t *testing.T) {
	message, ok := transport.DecodeControlPanelConstructorProduceItemMessage([]byte(`{
		"type": "controlPanelConstructorProduceItem",
		"clientSessionId": "session-1",
		"mutationSeq": 8,
		"constructorEquipmentGroupId": 13,
		"materialContainerEquipmentGroupId": 11,
		"blueprintId": 31,
		"amount": 1
	}`))

	if !ok {
		t.Fatalf("constructor blueprint production message was not decoded")
	}
	if message.BlueprintID != 31 || message.SchemaID != 0 || message.Amount != 1 {
		t.Fatalf("constructor blueprint production fields were decoded incorrectly: %+v", message)
	}
}

// Проверяет, что команда изменения очереди конструктора читает выбранную строку и действие.
func TestDecodeControlPanelConstructorQueueCommandMessageUsesAgreedJSONFields(t *testing.T) {
	message, ok := transport.DecodeControlPanelConstructorQueueCommandMessage([]byte(`{
		"type": "controlPanelConstructorQueueCommand",
		"clientSessionId": "session-1",
		"mutationSeq": 9,
		"constructorEquipmentGroupId": 13,
		"jobId": 31,
		"command": "cancelAll"
	}`))

	if !ok {
		t.Fatalf("constructor queue command message was not decoded")
	}
	if message.ClientSessionID != "session-1" || message.MutationSeq != 9 || message.ConstructorEquipmentGroupID != 13 || message.JobID != 31 || message.Command != "cancelAll" {
		t.Fatalf("constructor queue command fields were decoded incorrectly: %+v", message)
	}
}

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
		Objects: []game.SnapshotCosmicObject{
			{
				CosmicObject: data.CosmicObject{
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
		},
		EquipmentGroups: []data.EquipmentGroup{
			{
				ID:                   3,
				CosmicObjectID:       7,
				EquipmentItemModelID: 101,
				EnabledCount:         2,
			},
		},
		ItemGroups: []data.ItemGroup{
			{
				ID:                        4,
				ContainerEquipmentGroupID: 3,
				ContentItemModelID:        303,
				Count:                     12,
			},
		},
		Tasks: []data.Task{
			{
				ID:                         5,
				ControllerEquipmentGroupID: 6,
				ParentTaskID:               1,
				TaskTypeID:                 2,
				RemainingEnergy:            7,
				TotalEnergy:                10,
				BatchCount:                 3,
				SchemaID:                   9,
				BlueprintID:                0,
			},
		},
		TaskItemGroups: []data.TaskItemGroup{
			{
				ID:          6,
				TaskID:      5,
				ItemModelID: 302,
				Count:       4,
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
		`"itemGroups":[`,
		`"ContainerEquipmentGroupID":3`,
		`"ContentItemModelID":303`,
		`"Count":12`,
		`"tasks":[`,
		`"ControllerEquipmentGroupID":6`,
		`"ParentTaskID":1`,
		`"TaskTypeID":2`,
		`"RemainingEnergy":7`,
		`"TotalEnergy":10`,
		`"BatchCount":3`,
		`"SchemaID":9`,
		`"taskItemGroups":[`,
		`"TaskID":5`,
		`"ItemModelID":302`,
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
