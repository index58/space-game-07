package ws

import (
	"sync"

	"github.com/gorilla/websocket"
	"space-game-07-server/internal/game"
	"space-game-07-server/internal/world"
)

// Хранит одно WebSocket-подключение и служебные каналы записи.
type Client struct {
	connection *websocket.Conn // Активное сетевое соединение с браузером.
	accountID  int64           // Подключенный аккаунт, которому принадлежит соединение.
	objectID   int64           // Объект мира, которым управляет подключенный аккаунт.
	send       chan []byte     // Очередь исходящих сообщений для отдельной горутины записи.
	done       chan struct{}   // Сигнал завершения чтения, записи и очистки соединения.
	closeOnce  sync.Once       // Защита от повторного закрытия служебного канала.
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

	hub.mu.Lock()
	hub.clients[client] = struct{}{}
	hub.mu.Unlock()

	go hub.readLoop(client)
	go hub.writeLoop(client)
}

// Отправляет снимок всем клиентам, подставляя каждому его собственный объект.
func (hub *Hub) Broadcast(snapshot game.Snapshot) {
	hub.mu.Lock()
	clients := make([]*Client, 0, len(hub.clients))
	for client := range hub.clients {
		clients = append(clients, client)
	}
	hub.mu.Unlock()

	for _, client := range clients {
		clientSnapshot := snapshot
		clientSnapshot.SelfObjectID = client.objectID
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
			continue
		}

		hub.world.SetInput(client.accountID, input)
	}
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
