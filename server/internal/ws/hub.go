package ws

import (
	"sync"

	"github.com/gorilla/websocket"
	"space-game-07-server/internal/data"
	"space-game-07-server/internal/game"
	"space-game-07-server/internal/world"
)

// Хранит одно WebSocket-подключение и служебные каналы записи.
type Client struct {
	connection        *websocket.Conn // Активное сетевое соединение с браузером.
	accountID         int64           // Подключенный аккаунт, которому принадлежит соединение.
	objectID          int64           // Объект мира, которым управляет подключенный аккаунт.
	selectedChatID    int64           // Последняя выбранная вкладка чата в этом браузере.
	mutationSessionID string          // Последняя сессия клиента, отправлявшая команды панели.
	send              chan []byte     // Очередь исходящих сообщений для отдельной горутины записи.
	done              chan struct{}   // Сигнал завершения чтения, записи и очистки соединения.
	closeOnce         sync.Once       // Защита от повторного закрытия служебного канала.
}

// Координирует все активные WebSocket-клиенты вокруг одного игрового мира.
type Hub struct {
	mu      sync.Mutex           // Защищает набор активных клиентов.
	world   *world.World         // Игровой мир, получающий ввод и отдающий снимки.
	clients map[*Client]struct{} // Набор текущих WebSocket-подключений.
}

// Создает пустой диспетчер подключений для указанного мира.
func NewHub(gameWorld *world.World) *Hub {
	return &Hub{
		world:   gameWorld,
		clients: map[*Client]struct{}{},
	}
}

// Регистрирует новое подключение и запускает отдельные циклы чтения и записи.
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

// Отправляет снимок всем клиентам, подставляя каждому его собственный объект.
func (hub *Hub) Broadcast(snapshot game.Snapshot) {
	hub.mu.Lock()
	clients := make([]struct {
		client            *Client
		mutationSessionID string
	}, 0, len(hub.clients))
	for client := range hub.clients {
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

// Принимает ввод от клиента и передает последний валидный пакет в мир.
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

// handleDockingRequest запускает исходящий запрос или возвращает причину отказа текущему подключению.
func (hub *Hub) handleDockingRequest(client *Client) {
	if err := hub.world.SendDockingRequest(client.accountID); err != nil {
		hub.sendDockingError(client, err.Error())
		return
	}
	hub.sendDockingEvents(hub.world.DrainDockingEvents())
}

// handleDockingApprove принимает входящий запрос или возвращает причину отказа текущему подключению.
func (hub *Hub) handleDockingApprove(client *Client) {
	if err := hub.world.ApproveDockingRequest(client.accountID); err != nil {
		hub.sendDockingError(client, err.Error())
	}
	hub.sendDockingEvents(hub.world.DrainDockingEvents())
}

// handleDockingReject отклоняет входящий запрос или возвращает причину отказа текущему подключению.
func (hub *Hub) handleDockingReject(client *Client) {
	if err := hub.world.RejectDockingRequest(client.accountID); err != nil {
		hub.sendDockingError(client, err.Error())
		return
	}
	hub.sendDockingEvents(hub.world.DrainDockingEvents())
}

// handleDockingUndock выводит объект из кластера или возвращает причину отказа текущему подключению.
func (hub *Hub) handleDockingUndock(client *Client) {
	if err := hub.world.UndockControlledObject(client.accountID); err != nil {
		hub.sendDockingError(client, err.Error())
		return
	}
	hub.sendDockingEvents(hub.world.DrainDockingEvents())
}

// handleControlPanelObjectUpdate применяет изменение объекта или возвращает отказ с номером мутации.
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

// handleControlPanelEquipmentUpdate применяет изменение оборудования или возвращает отказ с номером мутации.
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

// handleControlPanelContainerTransfer применяет перенос между контейнерами или возвращает отказ с номером мутации.
// handleControlPanelEquipmentGroupRelationUpdate применяет сохранение связи групп оборудования или возвращает отказ с номером мутации.
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
		SourceContainerEquipmentGroupID: message.SourceContainerEquipmentGroupID,
		TargetContainerEquipmentGroupID: message.TargetContainerEquipmentGroupID,
		ItemGroupIDs:                    message.ItemGroupIDs,
		Amount:                          message.Amount,
	})
	if err != nil {
		hub.sendControlPanelError(client, message.ClientSessionID, message.MutationSeq, err.Error())
	}
}

// setClientMutationSession запоминает сессию команд панели под mutex диспетчера.
// handleControlPanelFuelTransfer применяет перенос топлива или возвращает отказ с номером мутации.
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

// handleControlPanelConstructorProduceItem применяет изготовление предмета или возвращает отказ с номером мутации.
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

// handleControlPanelConstructorQueueCommand применяет изменение очереди конструктора или возвращает отказ с номером мутации.
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

// sendControlPanelError отправляет отказ команды панели управления текущему подключению.
func (hub *Hub) sendControlPanelError(client *Client, sessionID string, mutationSeq int64, message string) {
	payload, err := EncodeControlPanelErrorMessage(sessionID, mutationSeq, message)
	if err != nil {
		return
	}
	hub.sendToClient(client, payload)
}

// sendDockingError отправляет отказ команды стыковки только текущему подключению.
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

// handleInputSettingsRequest возвращает текущие сохраненные настройки аккаунта без изменения мира.
func (hub *Hub) handleInputSettingsRequest(client *Client) {
	payload, err := EncodeInputSettingsMessage(hub.world.AccountInputSettings(client.accountID))
	if err != nil {
		return
	}
	hub.sendToClient(client, payload)
}

// handleInputSettingsSave сохраняет новые привязки и возвращает актуальное состояние аккаунта.
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

// Обрабатывает отправку текста и рассылает обновления вкладок всем доступным получателям.
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

// Обрабатывает выбор вкладки и возвращает состояние с обновленной позицией чтения.
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

// Кладет пакет в очередь конкретного подключения.
func (hub *Hub) sendToClient(client *Client, payload []byte) {
	select {
	case client.send <- payload:
	case <-client.done:
	default:
		hub.removeClient(client)
	}
}

// Кладет пакет во все соединения указанной учетной записи.
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

// sendDockingEvents рассылает события игрокам, управляющим указанными объектами.
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

// Проверяет, что список получателей содержит указанный аккаунт.
func containsAccountID(accountIDs []int64, accountID int64) bool {
	for _, candidateID := range accountIDs {
		if candidateID == accountID {
			return true
		}
	}
	return false
}

// containsObjectID проверяет, что список получателей содержит указанный объект.
func containsObjectID(objectIDs []int64, objectID int64) bool {
	for _, candidateID := range objectIDs {
		if candidateID == objectID {
			return true
		}
	}
	return false
}

// Пишет исходящие снимки в WebSocket, пока клиент не отключился.
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

// Идемпотентно закрывает подключение и очищает привязку аккаунта в мире.
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
