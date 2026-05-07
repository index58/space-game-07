package main

import (
	"log"
	"net/http"
	"time"

	"space-game-07-server/internal/storage"
	"space-game-07-server/internal/world"
	transport "space-game-07-server/internal/ws"
)

// Задает частоту серверной симуляции и отправки снимков клиентам.
const tickRate = 30

func main() {
	// При старте сервер поднимает весь игровой мир из локальных JSON-файлов.
	serverData, err := storage.LoadServerData(".")
	if err != nil {
		log.Fatal(err)
	}
	log.Printf(
		"loaded %d accounts, %d characters, %d cosmic objects, %d cosmic object types, %d cosmic object models and %d itemtypes from data",
		len(serverData.Accounts.Items),
		len(serverData.Characters.Items),
		len(serverData.CosmicObjects.Items),
		len(serverData.CosmicObjectTypes.Items),
		len(serverData.CosmicObjectModels.Items),
		len(serverData.Itemtypes.Items),
	)

	gameWorld := world.New(time.Now().UnixNano(), world.Data{
		Accounts:                   serverData.Accounts,
		Characters:                 serverData.Characters,
		CosmicObjects:              serverData.CosmicObjects,
		CosmicObjectTypes:          serverData.CosmicObjectTypes,
		CosmicObjectModels:         serverData.CosmicObjectModels,
		Itemtypes:                  serverData.Itemtypes,
		ItemModels:                 serverData.ItemModels,
		EquipmentGroups:            serverData.EquipmentGroups,
		ItemGroups:                 serverData.ItemGroups,
		Assemblies:                 serverData.Assemblies,
		AssemblyEquipmentGroups:    serverData.AssemblyEquipmentGroups,
		Chats:                      serverData.Chats,
		ChatMembers:                serverData.ChatMembers,
		CommunityTypes:             serverData.CommunityTypes,
		CommunityChatRoles:         serverData.CommunityChatRoles,
		Messages:                   serverData.Messages,
		MessageReads:               serverData.MessageReads,
		MessageTypes:               serverData.MessageTypes,
		ActionTypes:                serverData.ActionTypes,
		InputEventTypes:            serverData.InputEventTypes,
		DefaultActionInputSettings: serverData.DefaultActionInputSettings,
		AccountActionInputSettings: serverData.AccountActionInputSettings,
	})
	hub := transport.NewHub(gameWorld)
	handler := transport.NewHandler(hub, serverData.Accounts)

	// HTTP-точки обслуживают игровой поток и стартовый пакет справочников.
	http.Handle("/ws", handler)
	http.Handle("/reference-data", transport.NewReferenceDataHandler(serverData))

	ticker := time.NewTicker(time.Second / tickRate)
	defer ticker.Stop()
	saveTicker := time.NewTicker(30 * time.Second)
	defer saveTicker.Stop()

	go func() {
		for range ticker.C {
			// Каждый тик двигает мир вперед и рассылает клиентам новый снимок.
			snapshot := gameWorld.Tick(1.0 / tickRate)
			hub.Broadcast(snapshot)
		}
	}()

	go func() {
		for range saveTicker.C {
			// Периодическое сохранение фиксирует изменившиеся координаты и параметры объектов.
			if err := gameWorld.SaveData("."); err != nil {
				log.Printf("save server data: %v", err)
			}
		}
	}()

	log.Println("space-game-server listening on 127.0.0.1:8080")
	log.Fatal(http.ListenAndServe("127.0.0.1:8080", nil))
}
