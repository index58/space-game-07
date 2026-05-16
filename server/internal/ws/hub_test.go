package ws

import (
	"encoding/json"
	"testing"

	"space-game-07-server/internal/data"
	"space-game-07-server/internal/game"
	"space-game-07-server/internal/world"
)

// Проверяет, что после пересадки снимок для WebSocket-клиента указывает новый управляемый объект.
func TestBroadcastUsesTransferredCharacterObject(t *testing.T) {
	gameWorld := world.New(1, testHubWorldData(t))
	initialObjectID, ok := gameWorld.ConnectAccount(1)
	if !ok {
		t.Fatal("account was not connected")
	}
	hub := NewHub(gameWorld)
	client := &Client{
		accountID: 1,
		objectID:  initialObjectID,
		send:      make(chan []byte, 1),
		done:      make(chan struct{}),
	}
	hub.clients[client] = struct{}{}

	if err := gameWorld.BeginCharacterTransfer(1); err != nil {
		t.Fatalf("character transfer returned error: %v", err)
	}
	if objectID, ok := gameWorld.ObjectIDForAccount(1); !ok || objectID != 2 {
		t.Fatalf("world account object = %v, %v, want 2, true", objectID, ok)
	}

	hub.Broadcast(gameWorld.Tick(0))

	var snapshot game.Snapshot
	select {
	case payload := <-client.send:
		if err := json.Unmarshal(payload, &snapshot); err != nil {
			t.Fatalf("snapshot decode returned error: %v", err)
		}
	default:
		t.Fatal("snapshot was not sent")
	}
	if snapshot.SelfObjectID != 2 {
		t.Fatalf("self object ID = %v, want 2", snapshot.SelfObjectID)
	}
}

// Собирает минимальный мир с персонажем и двумя пристыкованными объектами одного владельца.
func testHubWorldData(t *testing.T) world.Data {
	t.Helper()

	accounts := &data.Accounts{
		MaxID: 1,
		Items: map[int64]*data.Account{
			1: {ID: 1, Email: "index@email.net", Nickname: "index", PasswordHash: "hash", Token: "token", CurrentCharacterID: 1},
		},
	}
	characters := &data.Characters{
		MaxID: 1,
		Items: map[int64]*data.Character{
			1: {ID: 1, AccountID: 1, LocationCosmicObjectID: 1},
		},
	}
	cosmicObjectTypes := &data.CosmicObjectTypes{
		MaxID: 1,
		Items: map[int64]*data.CosmicObjectType{
			1: {ID: 1, TitleRu: "Корабль", TitleEn: "Ship", Acronym: "Ship"},
		},
	}
	cosmicObjectModels := &data.CosmicObjectModels{
		MaxID: 1,
		Items: map[int64]*data.CosmicObjectModel{
			1: {ID: 1, TitleRu: "Корабль", TitleEn: "Ship", Acronym: "Ship", CosmicObjectTypeID: 1, TextureScale: 4, TextureBodyWidth: 20, TextureBodyLength: 40},
		},
	}
	cosmicObjects := &data.CosmicObjects{
		MaxID: 2,
		Items: map[int64]*data.CosmicObject{
			1: {ID: 1, Title: "Main", CosmicObjectModelID: 1, OwnerCharacterID: 1, ClusterMainCosmicObjectID: 1, Enabled: true, Anchored: true},
			2: {ID: 2, Title: "Secondary", CosmicObjectModelID: 1, OwnerCharacterID: 1, ClusterMainCosmicObjectID: 1, Enabled: true, Anchored: true},
		},
	}
	if err := accounts.RebuildIndexes(); err != nil {
		t.Fatal(err)
	}
	if err := characters.RebuildIndexes(); err != nil {
		t.Fatal(err)
	}
	if err := cosmicObjectTypes.RebuildIndexes(); err != nil {
		t.Fatal(err)
	}
	if err := cosmicObjectModels.RebuildIndexes(); err != nil {
		t.Fatal(err)
	}
	if err := cosmicObjects.RebuildIndexes(); err != nil {
		t.Fatal(err)
	}

	return world.Data{
		Accounts:           accounts,
		Characters:         characters,
		CosmicObjects:      cosmicObjects,
		CosmicObjectTypes:  cosmicObjectTypes,
		CosmicObjectModels: cosmicObjectModels,
	}
}
