package main

import (
	"log"
	"net/http"
	"time"

	"space-game-07-server/internal/storage"
	"space-game-07-server/internal/world"
	transport "space-game-07-server/internal/ws"
)

// Р—Р°РґР°РµС‚ С‡Р°СЃС‚РѕС‚Сѓ СЃРµСЂРІРµСЂРЅРѕР№ СЃРёРјСѓР»СЏС†РёРё Рё РѕС‚РїСЂР°РІРєРё СЃРЅРёРјРєРѕРІ РєР»РёРµРЅС‚Р°Рј.
const tickRate = 30

func main() {
	// РџСЂРё СЃС‚Р°СЂС‚Рµ СЃРµСЂРІРµСЂ РїРѕРґРЅРёРјР°РµС‚ РІРµСЃСЊ РёРіСЂРѕРІРѕР№ РјРёСЂ РёР· Р»РѕРєР°Р»СЊРЅС‹С… JSON-С„Р°Р№Р»РѕРІ.
	serverData, err := storage.LoadServerData(".")
	if err != nil {
		log.Fatal(err)
	}
	log.Printf(
		"loaded %d accounts, %d characters, %d cosmic objects, %d cosmic object types, %d cosmic object models and %d ItemTypes from data",
		len(serverData.Accounts.Items),
		len(serverData.Characters.Items),
		len(serverData.CosmicObjects.Items),
		len(serverData.CosmicObjectTypes.Items),
		len(serverData.CosmicObjectModels.Items),
		len(serverData.ItemTypes.Items),
	)

	gameWorld := world.New(time.Now().UnixNano(), world.Data{
		Accounts:                   serverData.Accounts,
		Characters:                 serverData.Characters,
		CosmicObjects:              serverData.CosmicObjects,
		CosmicObjectTypes:          serverData.CosmicObjectTypes,
		CosmicObjectModels:         serverData.CosmicObjectModels,
		ItemTypes:                  serverData.ItemTypes,
		ItemModels:                 serverData.ItemModels,
		Blueprints:                 serverData.Blueprints,
		BlueprintComponents:        serverData.BlueprintComponents,
		Schemas:                    serverData.Schemas,
		SchemaComponents:           serverData.SchemaComponents,
		TaskTypes:                  serverData.TaskTypes,
		Tasks:                      serverData.Tasks,
		TaskItemGroups:             serverData.TaskItemGroups,
		Implementers:               serverData.Implementers,
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

	// HTTP-С‚РѕС‡РєРё РѕР±СЃР»СѓР¶РёРІР°СЋС‚ РёРіСЂРѕРІРѕР№ РїРѕС‚РѕРє Рё СЃС‚Р°СЂС‚РѕРІС‹Р№ РїР°РєРµС‚ СЃРїСЂР°РІРѕС‡РЅРёРєРѕРІ.
	http.Handle("/ws", handler)
	http.Handle("/reference-data", transport.NewReferenceDataHandler(serverData))

	ticker := time.NewTicker(time.Second / tickRate)
	defer ticker.Stop()
	saveTicker := time.NewTicker(30 * time.Second)
	defer saveTicker.Stop()

	go func() {
		for range ticker.C {
			// РљР°Р¶РґС‹Р№ С‚РёРє РґРІРёРіР°РµС‚ РјРёСЂ РІРїРµСЂРµРґ Рё СЂР°СЃСЃС‹Р»Р°РµС‚ РєР»РёРµРЅС‚Р°Рј РЅРѕРІС‹Р№ СЃРЅРёРјРѕРє.
			snapshot := gameWorld.Tick(1.0 / tickRate)
			hub.Broadcast(snapshot)
		}
	}()

	go func() {
		for range saveTicker.C {
			// РџРµСЂРёРѕРґРёС‡РµСЃРєРѕРµ СЃРѕС…СЂР°РЅРµРЅРёРµ С„РёРєСЃРёСЂСѓРµС‚ РёР·РјРµРЅРёРІС€РёРµСЃСЏ РєРѕРѕСЂРґРёРЅР°С‚С‹ Рё РїР°СЂР°РјРµС‚СЂС‹ РѕР±СЉРµРєС‚РѕРІ.
			if err := gameWorld.SaveData("."); err != nil {
				log.Printf("save server data: %v", err)
			}
		}
	}()

	log.Println("space-game-server listening on 127.0.0.1:8080")
	log.Fatal(http.ListenAndServe("127.0.0.1:8080", nil))
}
