package ws

import (
	"sync"

	"github.com/gorilla/websocket"
	"space-game-07-server/internal/data"
	"space-game-07-server/internal/game"
	"space-game-07-server/internal/world"
)

// РҐСЂР°РЅРёС‚ РѕРґРЅРѕ WebSocket-РїРѕРґРєР»СЋС‡РµРЅРёРµ Рё СЃР»СѓР¶РµР±РЅС‹Рµ РєР°РЅР°Р»С‹ Р·Р°РїРёСЃРё.
type Client struct {
	connection        *websocket.Conn // РђРєС‚РёРІРЅРѕРµ СЃРµС‚РµРІРѕРµ СЃРѕРµРґРёРЅРµРЅРёРµ СЃ Р±СЂР°СѓР·РµСЂРѕРј.
	accountID         int64           // РџРѕРґРєР»СЋС‡РµРЅРЅС‹Р№ Р°РєРєР°СѓРЅС‚, РєРѕС‚РѕСЂРѕРјСѓ РїСЂРёРЅР°РґР»РµР¶РёС‚ СЃРѕРµРґРёРЅРµРЅРёРµ.
	objectID          int64           // РћР±СЉРµРєС‚ РјРёСЂР°, РєРѕС‚РѕСЂС‹Рј СѓРїСЂР°РІР»СЏРµС‚ РїРѕРґРєР»СЋС‡РµРЅРЅС‹Р№ Р°РєРєР°СѓРЅС‚.
	selectedChatID    int64           // РџРѕСЃР»РµРґРЅСЏСЏ РІС‹Р±СЂР°РЅРЅР°СЏ РІРєР»Р°РґРєР° С‡Р°С‚Р° РІ СЌС‚РѕРј Р±СЂР°СѓР·РµСЂРµ.
	mutationSessionID string          // РџРѕСЃР»РµРґРЅСЏСЏ СЃРµСЃСЃРёСЏ РєР»РёРµРЅС‚Р°, РѕС‚РїСЂР°РІР»СЏРІС€Р°СЏ РєРѕРјР°РЅРґС‹ РїР°РЅРµР»Рё.
	send              chan []byte     // РћС‡РµСЂРµРґСЊ РёСЃС…РѕРґСЏС‰РёС… СЃРѕРѕР±С‰РµРЅРёР№ РґР»СЏ РѕС‚РґРµР»СЊРЅРѕР№ РіРѕСЂСѓС‚РёРЅС‹ Р·Р°РїРёСЃРё.
	done              chan struct{}   // РЎРёРіРЅР°Р» Р·Р°РІРµСЂС€РµРЅРёСЏ С‡С‚РµРЅРёСЏ, Р·Р°РїРёСЃРё Рё РѕС‡РёСЃС‚РєРё СЃРѕРµРґРёРЅРµРЅРёСЏ.
	closeOnce         sync.Once       // Р—Р°С‰РёС‚Р° РѕС‚ РїРѕРІС‚РѕСЂРЅРѕРіРѕ Р·Р°РєСЂС‹С‚РёСЏ СЃР»СѓР¶РµР±РЅРѕРіРѕ РєР°РЅР°Р»Р°.
}

// РљРѕРѕСЂРґРёРЅРёСЂСѓРµС‚ РІСЃРµ Р°РєС‚РёРІРЅС‹Рµ WebSocket-РєР»РёРµРЅС‚С‹ РІРѕРєСЂСѓРі РѕРґРЅРѕРіРѕ РёРіСЂРѕРІРѕРіРѕ РјРёСЂР°.
type Hub struct {
	mu      sync.Mutex           // Р—Р°С‰РёС‰Р°РµС‚ РЅР°Р±РѕСЂ Р°РєС‚РёРІРЅС‹С… РєР»РёРµРЅС‚РѕРІ.
	world   *world.World         // РРіСЂРѕРІРѕР№ РјРёСЂ, РїРѕР»СѓС‡Р°СЋС‰РёР№ РІРІРѕРґ Рё РѕС‚РґР°СЋС‰РёР№ СЃРЅРёРјРєРё.
	clients map[*Client]struct{} // РќР°Р±РѕСЂ С‚РµРєСѓС‰РёС… WebSocket-РїРѕРґРєР»СЋС‡РµРЅРёР№.
}

// РЎРѕР·РґР°РµС‚ РїСѓСЃС‚РѕР№ РґРёСЃРїРµС‚С‡РµСЂ РїРѕРґРєР»СЋС‡РµРЅРёР№ РґР»СЏ СѓРєР°Р·Р°РЅРЅРѕРіРѕ РјРёСЂР°.
func NewHub(gameWorld *world.World) *Hub {
	return &Hub{
		world:   gameWorld,
		clients: map[*Client]struct{}{},
	}
}

// Р РµРіРёСЃС‚СЂРёСЂСѓРµС‚ РЅРѕРІРѕРµ РїРѕРґРєР»СЋС‡РµРЅРёРµ Рё Р·Р°РїСѓСЃРєР°РµС‚ РѕС‚РґРµР»СЊРЅС‹Рµ С†РёРєР»С‹ С‡С‚РµРЅРёСЏ Рё Р·Р°РїРёСЃРё.
func (hub *Hub) AddConnection(connection *websocket.Conn, accountID int64, initialMessages ...[]byte) {
	objectID, ok := hub.world.ConnectAccount(accountID)
	if !ok {
		_ = connection.Close()
		return
	}

	client := &Client{
		connection: connection,
		accountID:  accountID,
		objectID:   objectID,
		send:       make(chan []byte, 8),
		done:       make(chan struct{}),
	}
	for _, payload := range initialMessages {
		client.send <- payload
	}
	if chatState, ok := hub.world.ChatStateForAccount(accountID, 0); ok {
		client.selectedChatID = chatState.SelectedChatID
		if payload, err := EncodeChatStateMessage(chatState); err == nil {
			client.send <- payload
		}
	}
	if payload, err := EncodeInputSettingsMessage(hub.world.AccountInputSettings(accountID)); err == nil {
		client.send <- payload
	}

	hub.mu.Lock()
	hub.clients[client] = struct{}{}
	hub.mu.Unlock()

	go hub.readLoop(client)
	go hub.writeLoop(client)
}

// РћС‚РїСЂР°РІР»СЏРµС‚ СЃРЅРёРјРѕРє РІСЃРµРј РєР»РёРµРЅС‚Р°Рј, РїРѕРґСЃС‚Р°РІР»СЏСЏ РєР°Р¶РґРѕРјСѓ РµРіРѕ СЃРѕР±СЃС‚РІРµРЅРЅС‹Р№ РѕР±СЉРµРєС‚.
func (hub *Hub) Broadcast(snapshot game.Snapshot) {
	hub.mu.Lock()
	clients := make([]struct {
		client            *Client
		mutationSessionID string
	}, 0, len(hub.clients))
	for client := range hub.clients {
		if objectID, ok := hub.world.ObjectIDForAccount(client.accountID); ok {
			client.objectID = objectID
		}
		clients = append(clients, struct {
			client            *Client
			mutationSessionID string
		}{client: client, mutationSessionID: client.mutationSessionID})
	}
	hub.mu.Unlock()

	for _, item := range clients {
		client := item.client
		clientSnapshot := snapshot
		clientSnapshot.SelfObjectID = client.objectID
		if item.mutationSessionID != "" {
			ack := hub.world.ClientMutationAck(client.accountID, item.mutationSessionID)
			clientSnapshot.ClientMutationAck = &ack
		}
		payload, err := EncodeSnapshotMessage(clientSnapshot)
		if err != nil {
			continue
		}

		select {
		case client.send <- payload:
		case <-client.done:
		default:
			hub.removeClient(client)
		}
	}
	hub.sendDockingEvents(hub.world.DrainDockingEvents())
}

// РџСЂРёРЅРёРјР°РµС‚ РІРІРѕРґ РѕС‚ РєР»РёРµРЅС‚Р° Рё РїРµСЂРµРґР°РµС‚ РїРѕСЃР»РµРґРЅРёР№ РІР°Р»РёРґРЅС‹Р№ РїР°РєРµС‚ РІ РјРёСЂ.
func (hub *Hub) readLoop(client *Client) {
	defer hub.removeClient(client)

	for {
		_, payload, err := client.connection.ReadMessage()
		if err != nil {
			return
		}

		input, ok := DecodeInputMessage(payload)
		if !ok {
			if DecodeRandomShipMessage(payload) {
				hub.world.ChangeControlledShipToRandomModel(client.accountID)
			}
			if DecodeTestClaimFocusedObjectOwnerMessage(payload) {
				_ = hub.world.ClaimFocusedObjectOwnerForTesting(client.accountID)
			}
			if message, ok := DecodeChatSendMessage(payload); ok {
				hub.handleChatSend(client, message)
			}
			if message, ok := DecodeChatSelectMessage(payload); ok {
				hub.handleChatSelect(client, message)
			}
			if message, ok := DecodeInputSettingsSaveMessage(payload); ok {
				hub.handleInputSettingsSave(client, message)
			}
			if DecodeInputSettingsRequestMessage(payload) {
				hub.handleInputSettingsRequest(client)
			}
			if DecodeDockingRequestMessage(payload) {
				hub.handleDockingRequest(client)
			}
			if DecodeDockingApproveMessage(payload) {
				hub.handleDockingApprove(client)
			}
			if DecodeDockingRejectMessage(payload) {
				hub.handleDockingReject(client)
			}
			if DecodeDockingUndockMessage(payload) {
				hub.handleDockingUndock(client)
			}
			if DecodeLandingBeginMessage(payload) {
				hub.handleLandingBegin(client)
			}
			if DecodeLandingApproveMessage(payload) {
				hub.handleLandingApprove(client)
			}
			if DecodeLandingRejectMessage(payload) {
				hub.handleLandingReject(client)
			}
			if message, ok := DecodeLandingRequestMessage(payload); ok {
				hub.handleLandingRequest(client, message)
			}
			if message, ok := DecodeControlPanelObjectUpdateMessage(payload); ok {
				hub.handleControlPanelObjectUpdate(client, message)
			}
			if message, ok := DecodeControlPanelEquipmentUpdateMessage(payload); ok {
				hub.handleControlPanelEquipmentUpdate(client, message)
			}
			if message, ok := DecodeControlPanelEquipmentGroupRelationUpdateMessage(payload); ok {
				hub.handleControlPanelEquipmentGroupRelationUpdate(client, message)
			}
			if message, ok := DecodeControlPanelContainerTransferMessage(payload); ok {
				hub.handleControlPanelContainerTransfer(client, message)
			}
			if message, ok := DecodeControlPanelFuelTransferMessage(payload); ok {
				hub.handleControlPanelFuelTransfer(client, message)
			}
			if message, ok := DecodeControlPanelConstructorProduceItemMessage(payload); ok {
				hub.handleControlPanelConstructorProduceItem(client, message)
			}
			if message, ok := DecodeControlPanelConstructorQueueCommandMessage(payload); ok {
				hub.handleControlPanelConstructorQueueCommand(client, message)
			}
			continue
		}

		hub.world.SetInput(client.accountID, input)
	}
}

// handleDockingRequest Р·Р°РїСѓСЃРєР°РµС‚ РёСЃС…РѕРґСЏС‰РёР№ Р·Р°РїСЂРѕСЃ РёР»Рё РІРѕР·РІСЂР°С‰Р°РµС‚ РїСЂРёС‡РёРЅСѓ РѕС‚РєР°Р·Р° С‚РµРєСѓС‰РµРјСѓ РїРѕРґРєР»СЋС‡РµРЅРёСЋ.
func (hub *Hub) handleDockingRequest(client *Client) {
	if err := hub.world.SendDockingRequest(client.accountID); err != nil {
		hub.sendDockingError(client, err.Error())
		return
	}
	hub.sendDockingEvents(hub.world.DrainDockingEvents())
}

// handleDockingApprove РїСЂРёРЅРёРјР°РµС‚ РІС…РѕРґСЏС‰РёР№ Р·Р°РїСЂРѕСЃ РёР»Рё РІРѕР·РІСЂР°С‰Р°РµС‚ РїСЂРёС‡РёРЅСѓ РѕС‚РєР°Р·Р° С‚РµРєСѓС‰РµРјСѓ РїРѕРґРєР»СЋС‡РµРЅРёСЋ.
func (hub *Hub) handleDockingApprove(client *Client) {
	if err := hub.world.ApproveDockingRequest(client.accountID); err != nil {
		hub.sendDockingError(client, err.Error())
	}
	hub.sendDockingEvents(hub.world.DrainDockingEvents())
}

// handleDockingReject РѕС‚РєР»РѕРЅСЏРµС‚ РІС…РѕРґСЏС‰РёР№ Р·Р°РїСЂРѕСЃ РёР»Рё РІРѕР·РІСЂР°С‰Р°РµС‚ РїСЂРёС‡РёРЅСѓ РѕС‚РєР°Р·Р° С‚РµРєСѓС‰РµРјСѓ РїРѕРґРєР»СЋС‡РµРЅРёСЋ.
func (hub *Hub) handleDockingReject(client *Client) {
	if err := hub.world.RejectDockingRequest(client.accountID); err != nil {
		hub.sendDockingError(client, err.Error())
		return
	}
	hub.sendDockingEvents(hub.world.DrainDockingEvents())
}

// handleDockingUndock РІС‹РІРѕРґРёС‚ РѕР±СЉРµРєС‚ РёР· РєР»Р°СЃС‚РµСЂР° РёР»Рё РІРѕР·РІСЂР°С‰Р°РµС‚ РїСЂРёС‡РёРЅСѓ РѕС‚РєР°Р·Р° С‚РµРєСѓС‰РµРјСѓ РїРѕРґРєР»СЋС‡РµРЅРёСЋ.
func (hub *Hub) handleDockingUndock(client *Client) {
	if err := hub.world.UndockControlledObject(client.accountID); err != nil {
		hub.sendDockingError(client, err.Error())
		return
	}
	hub.sendDockingEvents(hub.world.DrainDockingEvents())
}

// handleLandingBegin Р·Р°РїСѓСЃРєР°РµС‚ РїРµСЂРµСЃР°РґРєСѓ РїРµСЂСЃРѕРЅР°Р¶Р° РёР»Рё РІРѕР·РІСЂР°С‰Р°РµС‚ РїСЂРёС‡РёРЅСѓ РѕС‚РєР°Р·Р° С‚РµРєСѓС‰РµРјСѓ РїРѕРґРєР»СЋС‡РµРЅРёСЋ.
func (hub *Hub) handleLandingBegin(client *Client) {
	if err := hub.world.BeginCharacterTransfer(client.accountID); err != nil {
		hub.sendDockingError(client, err.Error())
		return
	}
	hub.sendDockingEvents(hub.world.DrainDockingEvents())
}

// handleLandingApprove РїСЂРёРЅРёРјР°РµС‚ РІС…РѕРґСЏС‰РёР№ Р·Р°РїСЂРѕСЃ РїРѕСЃР°РґРєРё РёР»Рё РІРѕР·РІСЂР°С‰Р°РµС‚ РїСЂРёС‡РёРЅСѓ РѕС‚РєР°Р·Р° С‚РµРєСѓС‰РµРјСѓ РїРѕРґРєР»СЋС‡РµРЅРёСЋ.
func (hub *Hub) handleLandingApprove(client *Client) {
	if err := hub.world.ApproveCharacterLanding(client.accountID); err != nil {
		hub.sendDockingError(client, err.Error())
		return
	}
	hub.sendDockingEvents(hub.world.DrainDockingEvents())
}

// handleLandingReject РѕС‚РєР»РѕРЅСЏРµС‚ РІС…РѕРґСЏС‰РёР№ Р·Р°РїСЂРѕСЃ РїРѕСЃР°РґРєРё РёР»Рё РІРѕР·РІСЂР°С‰Р°РµС‚ РїСЂРёС‡РёРЅСѓ РѕС‚РєР°Р·Р° С‚РµРєСѓС‰РµРјСѓ РїРѕРґРєР»СЋС‡РµРЅРёСЋ.
func (hub *Hub) handleLandingReject(client *Client) {
	if err := hub.world.RejectCharacterLanding(client.accountID); err != nil {
		hub.sendDockingError(client, err.Error())
		return
	}
	hub.sendDockingEvents(hub.world.DrainDockingEvents())
}

// handleLandingRequest РѕС‚РїСЂР°РІР»СЏРµС‚ Р·Р°РїСЂРѕСЃ РїРѕСЃР°РґРєРё РІ РІС‹Р±СЂР°РЅРЅС‹Р№ РѕР±СЉРµРєС‚ РЅР°Р·РЅР°С‡РµРЅРёСЏ.
func (hub *Hub) handleLandingRequest(client *Client, message LandingRequestMessage) {
	if err := hub.world.RequestCharacterLanding(client.accountID, message.TargetObjectID); err != nil {
		hub.sendDockingError(client, err.Error())
		return
	}
	hub.sendDockingEvents(hub.world.DrainDockingEvents())
}

// handleControlPanelObjectUpdate РїСЂРёРјРµРЅСЏРµС‚ РёР·РјРµРЅРµРЅРёРµ РѕР±СЉРµРєС‚Р° РёР»Рё РІРѕР·РІСЂР°С‰Р°РµС‚ РѕС‚РєР°Р· СЃ РЅРѕРјРµСЂРѕРј РјСѓС‚Р°С†РёРё.
func (hub *Hub) handleControlPanelObjectUpdate(client *Client, message ControlPanelObjectUpdateMessage) {
	hub.setClientMutationSession(client, message.ClientSessionID)
	err := hub.world.ApplyControlPanelObjectUpdate(client.accountID, message.ClientSessionID, message.MutationSeq, world.ControlPanelObjectUpdate{
		Enabled: message.Enabled,
		Title:   message.Title,
	})
	if err != nil {
		hub.sendControlPanelError(client, message.ClientSessionID, message.MutationSeq, err.Error())
	}
}

// handleControlPanelEquipmentUpdate РїСЂРёРјРµРЅСЏРµС‚ РёР·РјРµРЅРµРЅРёРµ РѕР±РѕСЂСѓРґРѕРІР°РЅРёСЏ РёР»Рё РІРѕР·РІСЂР°С‰Р°РµС‚ РѕС‚РєР°Р· СЃ РЅРѕРјРµСЂРѕРј РјСѓС‚Р°С†РёРё.
func (hub *Hub) handleControlPanelEquipmentUpdate(client *Client, message ControlPanelEquipmentUpdateMessage) {
	hub.setClientMutationSession(client, message.ClientSessionID)
	err := hub.world.ApplyControlPanelEquipmentUpdate(client.accountID, message.ClientSessionID, message.MutationSeq, world.ControlPanelEquipmentUpdate{
		EquipmentGroupID: message.EquipmentGroupID,
		Enabled:          message.Enabled,
		EnabledCount:     message.EnabledCount,
		Title:            message.Title,
	})
	if err != nil {
		hub.sendControlPanelError(client, message.ClientSessionID, message.MutationSeq, err.Error())
	}
}

// handleControlPanelContainerTransfer РїСЂРёРјРµРЅСЏРµС‚ РїРµСЂРµРЅРѕСЃ РјРµР¶РґСѓ РєРѕРЅС‚РµР№РЅРµСЂР°РјРё РёР»Рё РІРѕР·РІСЂР°С‰Р°РµС‚ РѕС‚РєР°Р· СЃ РЅРѕРјРµСЂРѕРј РјСѓС‚Р°С†РёРё.
// handleControlPanelEquipmentGroupRelationUpdate РїСЂРёРјРµРЅСЏРµС‚ СЃРѕС…СЂР°РЅРµРЅРёРµ СЃРІСЏР·Рё РіСЂСѓРїРї РѕР±РѕСЂСѓРґРѕРІР°РЅРёСЏ РёР»Рё РІРѕР·РІСЂР°С‰Р°РµС‚ РѕС‚РєР°Р· СЃ РЅРѕРјРµСЂРѕРј РјСѓС‚Р°С†РёРё.
func (hub *Hub) handleControlPanelEquipmentGroupRelationUpdate(client *Client, message ControlPanelEquipmentGroupRelationUpdateMessage) {
	hub.setClientMutationSession(client, message.ClientSessionID)
	err := hub.world.ApplyControlPanelEquipmentGroupRelationUpdate(client.accountID, message.ClientSessionID, message.MutationSeq, world.ControlPanelEquipmentGroupRelationUpdate{
		EquipmentGroupID:        message.EquipmentGroupID,
		RelationTypeAcronym:     message.RelationTypeAcronym,
		RelatedEquipmentGroupID: message.RelatedEquipmentGroupID,
	})
	if err != nil {
		hub.sendControlPanelError(client, message.ClientSessionID, message.MutationSeq, err.Error())
	}
}

func (hub *Hub) handleControlPanelContainerTransfer(client *Client, message ControlPanelContainerTransferMessage) {
	hub.setClientMutationSession(client, message.ClientSessionID)
	err := hub.world.ApplyControlPanelContainerTransfer(client.accountID, message.ClientSessionID, message.MutationSeq, world.ControlPanelContainerTransfer{
		ControllerEquipmentGroupID:      message.ControllerEquipmentGroupID,
		LeftToRightDirection:            message.LeftToRightDirection,
		SourceContainerEquipmentGroupID: message.SourceContainerEquipmentGroupID,
		TargetContainerEquipmentGroupID: message.TargetContainerEquipmentGroupID,
		ItemGroupIDs:                    message.ItemGroupIDs,
		Amount:                          message.Amount,
	})
	if err != nil {
		hub.sendControlPanelError(client, message.ClientSessionID, message.MutationSeq, err.Error())
	}
}

// setClientMutationSession Р·Р°РїРѕРјРёРЅР°РµС‚ СЃРµСЃСЃРёСЋ РєРѕРјР°РЅРґ РїР°РЅРµР»Рё РїРѕРґ mutex РґРёСЃРїРµС‚С‡РµСЂР°.
// handleControlPanelFuelTransfer РїСЂРёРјРµРЅСЏРµС‚ РїРµСЂРµРЅРѕСЃ С‚РѕРїР»РёРІР° РёР»Рё РІРѕР·РІСЂР°С‰Р°РµС‚ РѕС‚РєР°Р· СЃ РЅРѕРјРµСЂРѕРј РјСѓС‚Р°С†РёРё.
func (hub *Hub) handleControlPanelFuelTransfer(client *Client, message ControlPanelFuelTransferMessage) {
	hub.setClientMutationSession(client, message.ClientSessionID)
	err := hub.world.ApplyControlPanelFuelTransfer(client.accountID, message.ClientSessionID, message.MutationSeq, world.ControlPanelFuelTransfer{
		ContainerEquipmentGroupID: message.ContainerEquipmentGroupID,
		FuelTankEquipmentGroupID:  message.FuelTankEquipmentGroupID,
		ItemGroupIDs:              message.ItemGroupIDs,
		Amount:                    message.Amount,
	})
	if err != nil {
		hub.sendControlPanelError(client, message.ClientSessionID, message.MutationSeq, err.Error())
	}
}

// handleControlPanelConstructorProduceItem РїСЂРёРјРµРЅСЏРµС‚ РёР·РіРѕС‚РѕРІР»РµРЅРёРµ РїСЂРµРґРјРµС‚Р° РёР»Рё РІРѕР·РІСЂР°С‰Р°РµС‚ РѕС‚РєР°Р· СЃ РЅРѕРјРµСЂРѕРј РјСѓС‚Р°С†РёРё.
func (hub *Hub) handleControlPanelConstructorProduceItem(client *Client, message ControlPanelConstructorProduceItemMessage) {
	hub.setClientMutationSession(client, message.ClientSessionID)
	err := hub.world.ApplyControlPanelConstructorProduceItem(client.accountID, message.ClientSessionID, message.MutationSeq, world.ControlPanelConstructorProduceItem{
		ConstructorEquipmentGroupID:       message.ConstructorEquipmentGroupID,
		MaterialContainerEquipmentGroupID: message.MaterialContainerEquipmentGroupID,
		ProductContainerEquipmentGroupID:  message.ProductContainerEquipmentGroupID,
		SchemaID:                          message.SchemaID,
		BlueprintID:                       message.BlueprintID,
		Amount:                            message.Amount,
	})
	if err != nil {
		hub.sendControlPanelError(client, message.ClientSessionID, message.MutationSeq, err.Error())
	}
}

// handleControlPanelConstructorQueueCommand РїСЂРёРјРµРЅСЏРµС‚ РёР·РјРµРЅРµРЅРёРµ РѕС‡РµСЂРµРґРё РєРѕРЅСЃС‚СЂСѓРєС‚РѕСЂР° РёР»Рё РІРѕР·РІСЂР°С‰Р°РµС‚ РѕС‚РєР°Р· СЃ РЅРѕРјРµСЂРѕРј РјСѓС‚Р°С†РёРё.
func (hub *Hub) handleControlPanelConstructorQueueCommand(client *Client, message ControlPanelConstructorQueueCommandMessage) {
	hub.setClientMutationSession(client, message.ClientSessionID)
	err := hub.world.ApplyControlPanelConstructorQueueCommand(client.accountID, message.ClientSessionID, message.MutationSeq, world.ControlPanelConstructorQueueCommand{
		ConstructorEquipmentGroupID: message.ConstructorEquipmentGroupID,
		JobID:                       message.JobID,
		Command:                     message.Command,
	})
	if err != nil {
		hub.sendControlPanelError(client, message.ClientSessionID, message.MutationSeq, err.Error())
	}
}

func (hub *Hub) setClientMutationSession(client *Client, sessionID string) {
	hub.mu.Lock()
	defer hub.mu.Unlock()

	client.mutationSessionID = sessionID
}

// sendControlPanelError РѕС‚РїСЂР°РІР»СЏРµС‚ РѕС‚РєР°Р· РєРѕРјР°РЅРґС‹ РїР°РЅРµР»Рё СѓРїСЂР°РІР»РµРЅРёСЏ С‚РµРєСѓС‰РµРјСѓ РїРѕРґРєР»СЋС‡РµРЅРёСЋ.
func (hub *Hub) sendControlPanelError(client *Client, sessionID string, mutationSeq int64, message string) {
	payload, err := EncodeControlPanelErrorMessage(sessionID, mutationSeq, message)
	if err != nil {
		return
	}
	hub.sendToClient(client, payload)
}

// sendDockingError РѕС‚РїСЂР°РІР»СЏРµС‚ РѕС‚РєР°Р· РєРѕРјР°РЅРґС‹ СЃС‚С‹РєРѕРІРєРё С‚РѕР»СЊРєРѕ С‚РµРєСѓС‰РµРјСѓ РїРѕРґРєР»СЋС‡РµРЅРёСЋ.
func (hub *Hub) sendDockingError(client *Client, message string) {
	payload, err := EncodeDockingEventMessage(game.DockingEvent{
		Type:    "dockingEvent",
		Kind:    "dockingNotification",
		Message: message,
	})
	if err != nil {
		return
	}
	hub.sendToClient(client, payload)
}

// handleInputSettingsRequest РІРѕР·РІСЂР°С‰Р°РµС‚ С‚РµРєСѓС‰РёРµ СЃРѕС…СЂР°РЅРµРЅРЅС‹Рµ РЅР°СЃС‚СЂРѕР№РєРё Р°РєРєР°СѓРЅС‚Р° Р±РµР· РёР·РјРµРЅРµРЅРёСЏ РјРёСЂР°.
func (hub *Hub) handleInputSettingsRequest(client *Client) {
	payload, err := EncodeInputSettingsMessage(hub.world.AccountInputSettings(client.accountID))
	if err != nil {
		return
	}
	hub.sendToClient(client, payload)
}

// handleInputSettingsSave СЃРѕС…СЂР°РЅСЏРµС‚ РЅРѕРІС‹Рµ РїСЂРёРІСЏР·РєРё Рё РІРѕР·РІСЂР°С‰Р°РµС‚ Р°РєС‚СѓР°Р»СЊРЅРѕРµ СЃРѕСЃС‚РѕСЏРЅРёРµ Р°РєРєР°СѓРЅС‚Р°.
func (hub *Hub) handleInputSettingsSave(client *Client, message InputSettingsSaveMessage) {
	settings := make([]data.AccountActionInputSetting, 0, len(message.Settings))
	for _, item := range message.Settings {
		settings = append(settings, data.AccountActionInputSetting{
			ActionTypeID:     item.ActionTypeID,
			InputEventTypeID: item.InputEventTypeID,
		})
	}
	saved, err := hub.world.SaveAccountInputSettings(client.accountID, settings)
	if err != nil {
		payload, encodeErr := EncodeInputSettingsErrorMessage(err.Error())
		if encodeErr != nil {
			return
		}
		hub.sendToClient(client, payload)
		return
	}
	payload, err := EncodeInputSettingsMessage(saved)
	if err != nil {
		return
	}
	hub.sendToAccount(client.accountID, payload)
}

// РћР±СЂР°Р±Р°С‚С‹РІР°РµС‚ РѕС‚РїСЂР°РІРєСѓ С‚РµРєСЃС‚Р° Рё СЂР°СЃСЃС‹Р»Р°РµС‚ РѕР±РЅРѕРІР»РµРЅРёСЏ РІРєР»Р°РґРѕРє РІСЃРµРј РґРѕСЃС‚СѓРїРЅС‹Рј РїРѕР»СѓС‡Р°С‚РµР»СЏРј.
func (hub *Hub) handleChatSend(client *Client, message ChatSendMessage) {
	chatState, recipientIDs, chatError := hub.world.SendChatMessage(client.accountID, message.ChatID, message.TargetNickname, message.Text)
	if chatError != "" {
		payload, err := EncodeChatErrorMessage(chatError)
		if err != nil {
			return
		}
		hub.sendToClient(client, payload)
		return
	}

	if len(recipientIDs) == 0 {
		recipientIDs = []int64{client.accountID}
	}
	client.selectedChatID = chatState.SelectedChatID
	hub.mu.Lock()
	clients := make([]*Client, 0, len(hub.clients))
	for recipient := range hub.clients {
		if containsAccountID(recipientIDs, recipient.accountID) {
			clients = append(clients, recipient)
		}
	}
	hub.mu.Unlock()
	for _, recipient := range clients {
		state, ok := hub.world.ChatStateForAccount(recipient.accountID, recipient.selectedChatID)
		if !ok {
			continue
		}
		payload, err := EncodeChatStateMessage(state)
		if err != nil {
			continue
		}
		hub.sendToClient(recipient, payload)
	}
}

// РћР±СЂР°Р±Р°С‚С‹РІР°РµС‚ РІС‹Р±РѕСЂ РІРєР»Р°РґРєРё Рё РІРѕР·РІСЂР°С‰Р°РµС‚ СЃРѕСЃС‚РѕСЏРЅРёРµ СЃ РѕР±РЅРѕРІР»РµРЅРЅРѕР№ РїРѕР·РёС†РёРµР№ С‡С‚РµРЅРёСЏ.
func (hub *Hub) handleChatSelect(client *Client, message ChatSelectMessage) {
	chatState, ok := hub.world.ChatStateForAccount(client.accountID, message.ChatID)
	if !ok {
		return
	}
	client.selectedChatID = chatState.SelectedChatID
	payload, err := EncodeChatStateMessage(chatState)
	if err != nil {
		return
	}
	hub.sendToClient(client, payload)
}

// РљР»Р°РґРµС‚ РїР°РєРµС‚ РІ РѕС‡РµСЂРµРґСЊ РєРѕРЅРєСЂРµС‚РЅРѕРіРѕ РїРѕРґРєР»СЋС‡РµРЅРёСЏ.
func (hub *Hub) sendToClient(client *Client, payload []byte) {
	select {
	case client.send <- payload:
	case <-client.done:
	default:
		hub.removeClient(client)
	}
}

// РљР»Р°РґРµС‚ РїР°РєРµС‚ РІРѕ РІСЃРµ СЃРѕРµРґРёРЅРµРЅРёСЏ СѓРєР°Р·Р°РЅРЅРѕР№ СѓС‡РµС‚РЅРѕР№ Р·Р°РїРёСЃРё.
func (hub *Hub) sendToAccount(accountID int64, payload []byte) {
	hub.mu.Lock()
	clients := make([]*Client, 0)
	for client := range hub.clients {
		if client.accountID == accountID {
			clients = append(clients, client)
		}
	}
	hub.mu.Unlock()

	for _, client := range clients {
		hub.sendToClient(client, payload)
	}
}

// sendDockingEvents СЂР°СЃСЃС‹Р»Р°РµС‚ СЃРѕР±С‹С‚РёСЏ РёРіСЂРѕРєР°Рј, СѓРїСЂР°РІР»СЏСЋС‰РёРј СѓРєР°Р·Р°РЅРЅС‹РјРё РѕР±СЉРµРєС‚Р°РјРё.
func (hub *Hub) sendDockingEvents(events []game.DockingEvent) {
	if len(events) == 0 {
		return
	}
	hub.mu.Lock()
	clients := make([]*Client, 0, len(hub.clients))
	for client := range hub.clients {
		clients = append(clients, client)
	}
	hub.mu.Unlock()

	for _, event := range events {
		recipients := make([]*Client, 0)
		for _, client := range clients {
			if containsObjectID(event.ObjectIDs, client.objectID) {
				recipients = append(recipients, client)
			}
		}
		if len(recipients) == 0 {
			continue
		}
		event.ObjectIDs = nil
		payload, err := EncodeDockingEventMessage(event)
		if err != nil {
			continue
		}
		for _, client := range recipients {
			hub.sendToClient(client, payload)
		}
	}
}

// РџСЂРѕРІРµСЂСЏРµС‚, С‡С‚Рѕ СЃРїРёСЃРѕРє РїРѕР»СѓС‡Р°С‚РµР»РµР№ СЃРѕРґРµСЂР¶РёС‚ СѓРєР°Р·Р°РЅРЅС‹Р№ Р°РєРєР°СѓРЅС‚.
func containsAccountID(accountIDs []int64, accountID int64) bool {
	for _, candidateID := range accountIDs {
		if candidateID == accountID {
			return true
		}
	}
	return false
}

// containsObjectID РїСЂРѕРІРµСЂСЏРµС‚, С‡С‚Рѕ СЃРїРёСЃРѕРє РїРѕР»СѓС‡Р°С‚РµР»РµР№ СЃРѕРґРµСЂР¶РёС‚ СѓРєР°Р·Р°РЅРЅС‹Р№ РѕР±СЉРµРєС‚.
func containsObjectID(objectIDs []int64, objectID int64) bool {
	for _, candidateID := range objectIDs {
		if candidateID == objectID {
			return true
		}
	}
	return false
}

// РџРёС€РµС‚ РёСЃС…РѕРґСЏС‰РёРµ СЃРЅРёРјРєРё РІ WebSocket, РїРѕРєР° РєР»РёРµРЅС‚ РЅРµ РѕС‚РєР»СЋС‡РёР»СЃСЏ.
func (hub *Hub) writeLoop(client *Client) {
	defer hub.removeClient(client)

	for {
		select {
		case payload := <-client.send:
			if err := client.connection.WriteMessage(websocket.TextMessage, payload); err != nil {
				return
			}
		case <-client.done:
			return
		}
	}
}

// РРґРµРјРїРѕС‚РµРЅС‚РЅРѕ Р·Р°РєСЂС‹РІР°РµС‚ РїРѕРґРєР»СЋС‡РµРЅРёРµ Рё РѕС‡РёС‰Р°РµС‚ РїСЂРёРІСЏР·РєСѓ Р°РєРєР°СѓРЅС‚Р° РІ РјРёСЂРµ.
func (hub *Hub) removeClient(client *Client) {
	hub.mu.Lock()
	if _, ok := hub.clients[client]; !ok {
		hub.mu.Unlock()
		return
	}
	delete(hub.clients, client)
	hub.mu.Unlock()

	hub.world.DisconnectAccount(client.accountID)
	client.closeOnce.Do(func() {
		close(client.done)
	})
	_ = client.connection.Close()
}
