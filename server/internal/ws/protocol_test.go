package ws_test

import (
	"encoding/json"
	"strings"
	"testing"

	"space-game-07-server/internal/data"
	"space-game-07-server/internal/game"
	transport "space-game-07-server/internal/ws"
)

// РџСЂРѕРІРµСЂСЏРµС‚, С‡С‚Рѕ РІС…РѕРґРЅРѕРµ СЃРѕРѕР±С‰РµРЅРёРµ СѓРїСЂР°РІР»РµРЅРёСЏ С‡РёС‚Р°РµС‚СЃСЏ РёР· СЃРѕРіР»Р°СЃРѕРІР°РЅРЅС‹С… JSON-РїРѕР»РµР№.
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

// РџСЂРѕРІРµСЂСЏРµС‚, С‡С‚Рѕ СЃРѕРѕР±С‰РµРЅРёРµ СЃ РЅРµРїРѕРґРґРµСЂР¶РёРІР°РµРјС‹Рј С‚РёРїРѕРј РЅРµ РїСЂРёРЅРёРјР°РµС‚СЃСЏ РєР°Рє СѓРїСЂР°РІР»РµРЅРёРµ.
func TestDecodeInputMessageRejectsUnknownType(t *testing.T) {
	_, ok := transport.DecodeInputMessage([]byte(`{"type":"unknown"}`))

	if ok {
		t.Fatalf("unknown message type was accepted")
	}
}

// РџСЂРѕРІРµСЂСЏРµС‚, С‡С‚Рѕ РєРѕРјР°РЅРґР° СЃР»СѓС‡Р°Р№РЅРѕР№ СЃРјРµРЅС‹ РєРѕСЂР°Р±Р»СЏ РїСЂРёРЅРёРјР°РµС‚СЃСЏ РїРѕ СЃРѕРіР»Р°СЃРѕРІР°РЅРЅРѕРјСѓ С‚РёРїСѓ.
func TestDecodeRandomShipMessageAcceptsAgreedType(t *testing.T) {
	if !transport.DecodeRandomShipMessage([]byte(`{"type":"randomShip"}`)) {
		t.Fatalf("random ship message was not accepted")
	}
}

// РџСЂРѕРІРµСЂСЏРµС‚, С‡С‚Рѕ РґСЂСѓРіРёРµ С‚РёРїС‹ СЃРѕРѕР±С‰РµРЅРёР№ РЅРµ СЂР°СЃРїРѕР·РЅР°СЋС‚СЃСЏ РєР°Рє РєРѕРјР°РЅРґР° СЃРјРµРЅС‹ РєРѕСЂР°Р±Р»СЏ.
func TestDecodeRandomShipMessageRejectsOtherTypes(t *testing.T) {
	if transport.DecodeRandomShipMessage([]byte(`{"type":"input"}`)) {
		t.Fatalf("input message was accepted as random ship command")
	}
}

// РџСЂРѕРІРµСЂСЏРµС‚, С‡С‚Рѕ РєРѕРјР°РЅРґР° РѕС‚РїСЂР°РІРєРё Р·Р°РїСЂРѕСЃР° СЃС‚С‹РєРѕРІРєРё РїСЂРёРЅРёРјР°РµС‚СЃСЏ РѕС‚РґРµР»СЊРЅС‹Рј WebSocket-СЃРѕРѕР±С‰РµРЅРёРµРј.
// РџСЂРѕРІРµСЂСЏРµС‚, С‡С‚Рѕ С‚РµСЃС‚РѕРІР°СЏ РєРѕРјР°РЅРґР° РїСЂРёСЃРІРѕРµРЅРёСЏ РѕР±СЉРµРєС‚Р° РїСЂРёРЅРёРјР°РµС‚СЃСЏ РїРѕ СЃРѕРіР»Р°СЃРѕРІР°РЅРЅРѕРјСѓ С‚РёРїСѓ.
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

// РџСЂРѕРІРµСЂСЏРµС‚, С‡С‚Рѕ РєРѕРјР°РЅРґР° РѕРґРѕР±СЂРµРЅРёСЏ Р·Р°РїСЂРѕСЃР° СЃС‚С‹РєРѕРІРєРё РїСЂРёРЅРёРјР°РµС‚СЃСЏ Р±РµР· РёРґРµРЅС‚РёС„РёРєР°С‚РѕСЂР° Р·Р°РїСЂРѕСЃР°.
func TestDecodeDockingApproveMessageAcceptsAgreedType(t *testing.T) {
	if !transport.DecodeDockingApproveMessage([]byte(`{"type":"dockingApprove"}`)) {
		t.Fatalf("docking approve message was not accepted")
	}
}

// РџСЂРѕРІРµСЂСЏРµС‚, С‡С‚Рѕ РєРѕРјР°РЅРґР° РѕС‚РєР°Р·Р° Р·Р°РїСЂРѕСЃР° СЃС‚С‹РєРѕРІРєРё РїСЂРёРЅРёРјР°РµС‚СЃСЏ Р±РµР· РёРґРµРЅС‚РёС„РёРєР°С‚РѕСЂР° Р·Р°РїСЂРѕСЃР°.
func TestDecodeDockingRejectMessageAcceptsAgreedType(t *testing.T) {
	if !transport.DecodeDockingRejectMessage([]byte(`{"type":"dockingReject"}`)) {
		t.Fatalf("docking reject message was not accepted")
	}
}

// РџСЂРѕРІРµСЂСЏРµС‚, С‡С‚Рѕ РєРѕРјР°РЅРґР° РѕС‚СЃС‚С‹РєРѕРІРєРё РїСЂРёРјРµРЅСЏРµС‚СЃСЏ Рє С‚РµРєСѓС‰РµРјСѓ РѕР±СЉРµРєС‚Сѓ Р±РµР· РґРѕРїРѕР»РЅРёС‚РµР»СЊРЅС‹С… РїР°СЂР°РјРµС‚СЂРѕРІ.
func TestDecodeDockingUndockMessageAcceptsAgreedType(t *testing.T) {
	if !transport.DecodeDockingUndockMessage([]byte(`{"type":"dockingUndock"}`)) {
		t.Fatalf("docking undock message was not accepted")
	}
}

// РџСЂРѕРІРµСЂСЏРµС‚, С‡С‚Рѕ РєРѕРјР°РЅРґР° РЅР°С‡Р°Р»Р° РїРµСЂРµСЃР°РґРєРё РїСЂРёРЅРёРјР°РµС‚СЃСЏ Р±РµР· РґРѕРїРѕР»РЅРёС‚РµР»СЊРЅС‹С… РїРѕР»РµР№.
func TestDecodeLandingBeginMessageAcceptsAgreedType(t *testing.T) {
	if !transport.DecodeLandingBeginMessage([]byte(`{"type":"landingBegin"}`)) {
		t.Fatalf("landing begin message was not accepted")
	}
}

// РџСЂРѕРІРµСЂСЏРµС‚, С‡С‚Рѕ РєРѕРјР°РЅРґР° РѕРґРѕР±СЂРµРЅРёСЏ РїРѕСЃР°РґРєРё РїСЂРёРЅРёРјР°РµС‚СЃСЏ Р±РµР· РёРґРµРЅС‚РёС„РёРєР°С‚РѕСЂР° Р·Р°РїСЂРѕСЃР°.
func TestDecodeLandingApproveMessageAcceptsAgreedType(t *testing.T) {
	if !transport.DecodeLandingApproveMessage([]byte(`{"type":"landingApprove"}`)) {
		t.Fatalf("landing approve message was not accepted")
	}
}

// РџСЂРѕРІРµСЂСЏРµС‚, С‡С‚Рѕ РєРѕРјР°РЅРґР° РѕС‚РєР°Р·Р° РїРѕСЃР°РґРєРё РїСЂРёРЅРёРјР°РµС‚СЃСЏ Р±РµР· РёРґРµРЅС‚РёС„РёРєР°С‚РѕСЂР° Р·Р°РїСЂРѕСЃР°.
func TestDecodeLandingRejectMessageAcceptsAgreedType(t *testing.T) {
	if !transport.DecodeLandingRejectMessage([]byte(`{"type":"landingReject"}`)) {
		t.Fatalf("landing reject message was not accepted")
	}
}

// РџСЂРѕРІРµСЂСЏРµС‚, С‡С‚Рѕ Р·Р°РїСЂРѕСЃ СЃРІРµР¶РёС… РЅР°СЃС‚СЂРѕРµРє РІРІРѕРґР° РїСЂРёРЅРёРјР°РµС‚СЃСЏ РїРѕ СЃРѕРіР»Р°СЃРѕРІР°РЅРЅРѕРјСѓ С‚РёРїСѓ.
func TestDecodeInputSettingsRequestMessageAcceptsAgreedType(t *testing.T) {
	if !transport.DecodeInputSettingsRequestMessage([]byte(`{"type":"inputSettingsRequest"}`)) {
		t.Fatalf("input settings request was not accepted")
	}
}

// РџСЂРѕРІРµСЂСЏРµС‚, С‡С‚Рѕ РєРѕРјР°РЅРґР° РїР°РЅРµР»Рё СѓРїСЂР°РІР»РµРЅРёСЏ РѕР±СЉРµРєС‚РѕРј С‡РёС‚Р°РµС‚ РёРґРµРЅС‚РёС„РёРєР°С‚РѕСЂ РјСѓС‚Р°С†РёРё Рё РёР·РјРµРЅСЏРµРјС‹Рµ РїРѕР»СЏ.
func TestDecodeControlPanelObjectUpdateMessageUsesAgreedJSONFields(t *testing.T) {
	message, ok := transport.DecodeControlPanelObjectUpdateMessage([]byte(`{
		"type": "controlPanelObjectUpdate",
		"clientSessionId": "session-1",
		"mutationSeq": 7,
		"enabled": false,
		"title": "РќРѕРІС‹Р№ РєРѕСЂР°Р±Р»СЊ"
	}`))

	if !ok {
		t.Fatalf("control panel object update was not accepted")
	}
	if message.ClientSessionID != "session-1" || message.MutationSeq != 7 || message.Enabled == nil || *message.Enabled || message.Title == nil || *message.Title != "РќРѕРІС‹Р№ РєРѕСЂР°Р±Р»СЊ" {
		t.Fatalf("decoded object update mismatch: %+v", message)
	}
}

// РџСЂРѕРІРµСЂСЏРµС‚, С‡С‚Рѕ РєРѕРјР°РЅРґР° РїР°РЅРµР»Рё СѓРїСЂР°РІР»РµРЅРёСЏ РѕР±РѕСЂСѓРґРѕРІР°РЅРёРµРј С‡РёС‚Р°РµС‚ РёРґРµРЅС‚РёС„РёРєР°С‚РѕСЂ РіСЂСѓРїРїС‹ Рё Р·РЅР°С‡РµРЅРёСЏ.
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

// РџСЂРѕРІРµСЂСЏРµС‚, С‡С‚Рѕ РєРѕРјР°РЅРґР° РїРµСЂРµРјРµС‰РµРЅРёСЏ СЃРѕРґРµСЂР¶РёРјРѕРіРѕ РєРѕРЅС‚РµР№РЅРµСЂР° С‡РёС‚Р°РµС‚СЃСЏ РёР· СЃРѕРіР»Р°СЃРѕРІР°РЅРЅС‹С… JSON-РїРѕР»РµР№.
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

// РџСЂРѕРІРµСЂСЏРµС‚, С‡С‚Рѕ РєРѕРјР°РЅРґР° С‡Р°С‚Р° С‡РёС‚Р°РµС‚ РІС‹Р±СЂР°РЅРЅСѓСЋ РІРєР»Р°РґРєСѓ Рё Р°РґСЂРµСЃРЅС‹Р№ РЅРёРє РёР· JSON.
// РџСЂРѕРІРµСЂСЏРµС‚, С‡С‚Рѕ РєРѕРјР°РЅРґР° РїРµСЂРµРЅРѕСЃР° С‚РѕРїР»РёРІР° С‡РёС‚Р°РµС‚СЃСЏ РёР· СЃРѕРіР»Р°СЃРѕРІР°РЅРЅС‹С… JSON-РїРѕР»РµР№.
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

// РџСЂРѕРІРµСЂСЏРµС‚, С‡С‚Рѕ РєРѕРјР°РЅРґР° РёР·РіРѕС‚РѕРІР»РµРЅРёСЏ РїСЂРµРґРјРµС‚Р° С‡РёС‚Р°РµС‚ РІС‹Р±СЂР°РЅРЅС‹Рµ РіСЂСѓРїРїС‹ РѕР±РѕСЂСѓРґРѕРІР°РЅРёСЏ Рё СЃС…РµРјСѓ РёР· СЃРѕРіР»Р°СЃРѕРІР°РЅРЅС‹С… JSON-РїРѕР»РµР№.
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

// РџСЂРѕРІРµСЂСЏРµС‚, С‡С‚Рѕ РєРѕРјР°РЅРґР° РёР·РіРѕС‚РѕРІР»РµРЅРёСЏ РѕР±СЉРµРєС‚Р° С‡РёС‚Р°РµС‚ С‡РµСЂС‚С‘Р¶ РёР· СЃРѕРіР»Р°СЃРѕРІР°РЅРЅС‹С… JSON-РїРѕР»РµР№.
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

// РџСЂРѕРІРµСЂСЏРµС‚, С‡С‚Рѕ РєРѕРјР°РЅРґР° РёР·РјРµРЅРµРЅРёСЏ РѕС‡РµСЂРµРґРё РєРѕРЅСЃС‚СЂСѓРєС‚РѕСЂР° С‡РёС‚Р°РµС‚ РІС‹Р±СЂР°РЅРЅСѓСЋ СЃС‚СЂРѕРєСѓ Рё РґРµР№СЃС‚РІРёРµ.
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

// РџСЂРѕРІРµСЂСЏРµС‚, С‡С‚Рѕ РґСЂСѓРіРёРµ РІС…РѕРґСЏС‰РёРµ С‚РёРїС‹ РЅРµ СЂР°СЃРїРѕР·РЅР°СЋС‚СЃСЏ РєР°Рє РѕС‚РїСЂР°РІРєР° С‚РµРєСЃС‚Р°.
func TestDecodeChatSendMessageRejectsOtherTypes(t *testing.T) {
	if _, ok := transport.DecodeChatSendMessage([]byte(`{"type":"input","text":"hello"}`)); ok {
		t.Fatalf("input message was accepted as chat command")
	}
}

// РџСЂРѕРІРµСЂСЏРµС‚, С‡С‚Рѕ РІС‹Р±РѕСЂ РІРєР»Р°РґРєРё С‡Р°С‚Р° С‡РёС‚Р°РµС‚ ID РІС‹Р±СЂР°РЅРЅРѕРіРѕ С‡Р°С‚Р° РёР· JSON.
func TestDecodeChatSelectMessageUsesAgreedJSONFields(t *testing.T) {
	message, ok := transport.DecodeChatSelectMessage([]byte(`{"type":"chatSelect","chatId":7}`))

	if !ok {
		t.Fatalf("chat selection was not accepted")
	}
	if message.ChatID != 7 {
		t.Fatalf("decoded chat selection mismatch: %+v", message)
	}
}

// РџСЂРѕРІРµСЂСЏРµС‚, С‡С‚Рѕ СЃРѕСЃС‚РѕСЏРЅРёРµ С‡Р°С‚Р° РєРѕРґРёСЂСѓРµС‚СЃСЏ СЃ СЃРѕРіР»Р°СЃРѕРІР°РЅРЅС‹РјРё РёРјРµРЅР°РјРё РїРѕР»РµР№.
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

// РџСЂРѕРІРµСЂСЏРµС‚, С‡С‚Рѕ РѕС‚РєР°Р· РєРѕРјР°РЅРґС‹ С‡Р°С‚Р° РІРѕР·РІСЂР°С‰Р°РµС‚СЃСЏ РѕС‚РґРµР»СЊРЅС‹Рј СЃРµС‚РµРІС‹Рј С‚РёРїРѕРј.
func TestEncodeChatErrorMessageUsesAgreedFields(t *testing.T) {
	payload, err := transport.EncodeChatErrorMessage("РђРґСЂРµСЃР°С‚ РЅРµ РЅР°Р№РґРµРЅ")
	if err != nil {
		t.Fatalf("encode chat error: %v", err)
	}

	jsonText := string(payload)
	for _, field := range []string{
		`"type":"chatError"`,
		`"message":"РђРґСЂРµСЃР°С‚ РЅРµ РЅР°Р№РґРµРЅ"`,
	} {
		if !strings.Contains(jsonText, field) {
			t.Fatalf("encoded chat error %s does not contain %s", jsonText, field)
		}
	}
}

// РџСЂРѕРІРµСЂСЏРµС‚, С‡С‚Рѕ СЃРЅРёРјРѕРє РјРёСЂР° РєРѕРґРёСЂСѓРµС‚СЃСЏ СЃ С‚РµРєСѓС‰РёРјРё РёРјРµРЅР°РјРё РїРѕР»РµР№ Рё Р±РµР· СѓРґР°Р»С‘РЅРЅС‹С… РїРѕР»РµР№.
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

// РџСЂРѕРІРµСЂСЏРµС‚, С‡С‚Рѕ СЃРѕРѕР±С‰РµРЅРёРµ Р°РІС‚РѕСЂРёР·Р°С†РёРё СЃРѕРґРµСЂР¶РёС‚ С‚РёРї Рё С‚РѕРєРµРЅ РґР»СЏ РѕС‚РїСЂР°РІРєРё РґРѕ СЃРЅРёРјРєРѕРІ РјРёСЂР°.
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
