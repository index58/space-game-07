package ws

import "space-game-07-server/internal/game"

func (hub *Hub) handleExchangeRequest(client *Client) {
	if err := hub.world.SendExchangeRequest(client.accountID); err != nil {
		hub.sendExchangeError(client, err.Error())
	}
	hub.sendExchangeEvents(hub.world.DrainExchangeEvents())
}

func (hub *Hub) handleExchangeApprove(client *Client) {
	if err := hub.world.ApproveExchangeRequest(client.accountID); err != nil {
		hub.sendExchangeError(client, err.Error())
	}
	hub.sendExchangeEvents(hub.world.DrainExchangeEvents())
}

func (hub *Hub) handleExchangeReject(client *Client) {
	if err := hub.world.RejectExchangeRequest(client.accountID); err != nil {
		hub.sendExchangeError(client, err.Error())
	}
	hub.sendExchangeEvents(hub.world.DrainExchangeEvents())
}

func (hub *Hub) handleExchangeCancel(client *Client) {
	if err := hub.world.CancelExchange(client.accountID); err != nil {
		hub.sendExchangeError(client, err.Error())
	}
	hub.sendExchangeEvents(hub.world.DrainExchangeEvents())
}

func (hub *Hub) handleExchangeSelectReceiver(client *Client, message ExchangeContainerMessage) {
	if err := hub.world.SelectExchangeReceiver(client.accountID, message.ContainerEquipmentGroupID); err != nil {
		hub.sendExchangeError(client, err.Error())
	}
	hub.sendExchangeEvents(hub.world.DrainExchangeEvents())
}

func (hub *Hub) handleExchangeSelectSource(client *Client, message ExchangeContainerMessage) {
	if err := hub.world.SelectExchangeSource(client.accountID, message.ContainerEquipmentGroupID); err != nil {
		hub.sendExchangeError(client, err.Error())
	}
	hub.sendExchangeEvents(hub.world.DrainExchangeEvents())
}

func (hub *Hub) handleExchangeAddItems(client *Client, message ExchangeAddItemsMessage) {
	if err := hub.world.AddExchangeItems(client.accountID, message.ItemGroupIDs, message.Amount); err != nil {
		hub.sendExchangeError(client, err.Error())
	}
	hub.sendExchangeEvents(hub.world.DrainExchangeEvents())
}

func (hub *Hub) handleExchangeConfirm(client *Client) {
	if err := hub.world.ConfirmExchange(client.accountID); err != nil {
		hub.sendExchangeError(client, err.Error())
	}
	hub.sendExchangeEvents(hub.world.DrainExchangeEvents())
}

func (hub *Hub) sendExchangeError(client *Client, message string) {
	payload, err := EncodeExchangeEventMessage(game.ExchangeEvent{
		Type:    "exchangeEvent",
		Kind:    "exchangeNotification",
		Message: exchangeErrorText(message),
	})
	if err != nil {
		return
	}
	hub.sendToClient(client, payload)
}

// Кладет текущее состояние обмена в очередь нового подключения.
func (hub *Hub) enqueueInitialExchangeEvents(client *Client) {
	for _, event := range hub.world.ExchangeEventsForAccount(client.accountID) {
		payload, err := EncodeExchangeEventMessage(event)
		if err != nil {
			continue
		}
		client.send <- payload
	}
}

// Переводит технические причины отказа обмена в текст для игрока.
func exchangeErrorText(message string) string {
	switch message {
	case "object already participates in exchange":
		return "Объект уже участвует в обмене"
	case "exchange requires another player":
		return "Для обмена нужен другой игрок"
	case "exchange request not found":
		return "Запрос обмена не найден"
	case "sender object not found":
		return "Объект отправителя не найден"
	case "receiver container is locked after confirmation":
		return "Контейнер-приёмник заблокирован после подтверждения"
	case "queue is locked by other player confirmation":
		return "Очередь заблокирована подтверждением второго игрока"
	case "source container is not selected":
		return "Контейнер-источник не выбран"
	case "exchange amount is empty":
		return "Количество не выбрано"
	case "item group not found":
		return "Строка предметов не найдена"
	case "item group does not belong to source container":
		return "Строка предметов не относится к контейнеру-источнику"
	case "receiver container is not selected":
		return "Контейнер-приёмник не выбран"
	case "exchange is already moving":
		return "Обмен уже выполняется"
	case "exchange target not found":
		return "Объект для обмена не найден"
	case "exchange participant not found":
		return "Участник обмена не найден"
	case "controlled object not found":
		return "Управляемый объект не найден"
	case "exchange session not found":
		return "Обмен не найден"
	case "exchange task type not found":
		return "Тип задачи обмена не найден"
	default:
		return message
	}
}

func (hub *Hub) sendExchangeEvents(events []game.ExchangeEvent) {
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
		payload, err := EncodeExchangeEventMessage(event)
		if err != nil {
			continue
		}
		for _, client := range recipients {
			hub.sendToClient(client, payload)
		}
	}
}
