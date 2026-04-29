package ws

import (
	"sync"

	"github.com/gorilla/websocket"
	"space-game-07-server/internal/game"
	"space-game-07-server/internal/world"
)

type Client struct {
	connection *websocket.Conn
	playerID   int64
	objectID   int64
	send       chan []byte
	done       chan struct{}
	closeOnce  sync.Once
}

type Hub struct {
	mu      sync.Mutex
	world   *world.World
	clients map[*Client]struct{}
}

func NewHub(gameWorld *world.World) *Hub {
	return &Hub{
		world:   gameWorld,
		clients: map[*Client]struct{}{},
	}
}

func (hub *Hub) AddConnection(connection *websocket.Conn) {
	playerID, objectID := hub.world.AddPlayer()
	client := &Client{
		connection: connection,
		playerID:   playerID,
		objectID:   objectID,
		send:       make(chan []byte, 8),
		done:       make(chan struct{}),
	}

	hub.mu.Lock()
	hub.clients[client] = struct{}{}
	hub.mu.Unlock()

	go hub.readLoop(client)
	go hub.writeLoop(client)
}

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

func (hub *Hub) readLoop(client *Client) {
	defer hub.removeClient(client)

	for {
		_, payload, err := client.connection.ReadMessage()
		if err != nil {
			return
		}

		input, ok := DecodeInputMessage(payload)
		if !ok {
			continue
		}

		hub.world.SetInput(client.playerID, input)
	}
}

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

func (hub *Hub) removeClient(client *Client) {
	hub.mu.Lock()
	if _, ok := hub.clients[client]; !ok {
		hub.mu.Unlock()
		return
	}
	delete(hub.clients, client)
	hub.mu.Unlock()

	hub.world.RemovePlayer(client.playerID)
	client.closeOnce.Do(func() {
		close(client.done)
	})
	_ = client.connection.Close()
}
