package main

import (
	"log"
	"net/http"
	"time"

	"space-game-07-server/internal/storage"
	"space-game-07-server/internal/world"
	transport "space-game-07-server/internal/ws"
)

const tickRate = 30

func main() {
	serverData, err := storage.LoadServerData(".")
	if err != nil {
		log.Fatal(err)
	}
	log.Printf(
		"loaded %d accounts, %d characters and %d cosmic object types from data",
		len(serverData.Accounts.Items),
		len(serverData.Characters.Items),
		len(serverData.CosmicObjectTypes.Items),
	)

	gameWorld := world.New(time.Now().UnixNano())
	hub := transport.NewHub(gameWorld)
	handler := transport.NewHandler(hub)

	http.Handle("/ws", handler)

	ticker := time.NewTicker(time.Second / tickRate)
	defer ticker.Stop()

	go func() {
		for range ticker.C {
			snapshot := gameWorld.Tick(1.0 / tickRate)
			hub.Broadcast(snapshot)
		}
	}()

	log.Println("space-game-server listening on 127.0.0.1:8080")
	log.Fatal(http.ListenAndServe("127.0.0.1:8080", nil))
}
