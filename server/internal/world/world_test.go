package world_test

import (
	"os"
	"path/filepath"
	"testing"

	"space-game-07-server/internal/data"
	"space-game-07-server/internal/game"
	"space-game-07-server/internal/world"
)

// Ищет объект в снимке мира по ID, чтобы тесты не зависели от порядка массива.
func findCosmicObjectInSnapshot(snapshot game.Snapshot, objectID int64) (data.CosmicObject, bool) {
	for _, object := range snapshot.Objects {
		if object.ID == objectID {
			return object, true
		}
	}

	return data.CosmicObject{}, false
}

// Собирает минимальный игровой мир с кораблем, астероидом и станцией.
func testWorldData(t *testing.T) world.Data {
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
		MaxID: 3,
		Items: map[int64]*data.CosmicObjectType{
			1: {ID: 1, TitleRu: "Корабль", TitleEn: "Ship", Acronym: "Ship"},
			2: {ID: 2, TitleRu: "Станция", TitleEn: "Station", Acronym: "Station"},
			3: {ID: 3, TitleRu: "Астероид", TitleEn: "Asteroid", Acronym: "Asteroid"},
		},
	}
	cosmicObjectModels := &data.CosmicObjectModels{
		MaxID: 4,
		Items: map[int64]*data.CosmicObjectModel{
			1: {ID: 1, TitleRu: "Корабль", TitleEn: "Ship", Acronym: "ship_bat", CosmicObjectTypeID: 1, TextureScale: 4, TextureBodyWidth: 27, TextureBodyLength: 63, Mass: 7.92, Capacity: 10, MaxArmor: 100, MaxSpeed: 497, MaxAngularSpeed: 3},
			2: {ID: 2, TitleRu: "Астероид", TitleEn: "Asteroid", Acronym: "asteroid_0002", CosmicObjectTypeID: 3, TextureScale: 4, TextureBodyWidth: 30, TextureBodyLength: 30},
			3: {ID: 3, TitleRu: "Станция", TitleEn: "Station", Acronym: "station_tiny_crumb", CosmicObjectTypeID: 2, TextureScale: 4, TextureBodyWidth: 30, TextureBodyLength: 30},
			4: {ID: 4, TitleRu: "Новый корабль", TitleEn: "New Ship", Acronym: "ship_new", CosmicObjectTypeID: 1, TextureScale: 4, TextureBodyWidth: 40, TextureBodyLength: 80, Mass: 12, Capacity: 20, MaxArmor: 150, MaxSpeed: 600, MaxAngularSpeed: 4},
		},
	}
	cosmicObjects := &data.CosmicObjects{
		MaxID: 3,
		Items: map[int64]*data.CosmicObject{
			1: {ID: 1, Title: "Ship", CosmicObjectModelID: 1, OwnerCharacterID: 1, Mass: 7.92, MaxSpeed: 497, MaxAngularSpeed: 3, MaxAlongForce: 1287901.529888449, MaxTorque: 653565, Enabled: true},
			2: {ID: 2, Title: "Asteroid", CosmicObjectModelID: 2, X: -500, Y: 800, Mass: 629.532, MaxSpeed: 475, MaxAngularSpeed: 3, Enabled: true, Anchored: true},
			3: {ID: 3, Title: "Station", CosmicObjectModelID: 3, X: 500, Y: 500, Mass: 185.625, MaxSpeed: 486, MaxAngularSpeed: 3, Enabled: true, Anchored: true},
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

func TestConnectAccountUsesCurrentCharacterLocation(t *testing.T) {
	gameWorld := world.New(1, testWorldData(t))

	objectID, ok := gameWorld.ConnectAccount(1)
	if !ok {
		t.Fatalf("account was not connected")
	}

	if objectID != 1 {
		t.Fatalf("got object ID %v, want 1", objectID)
	}
}

func TestTickAppliesAccountInputToExistingShip(t *testing.T) {
	serverData := testWorldData(t)
	gameWorld := world.New(1, serverData)

	objectID, ok := gameWorld.ConnectAccount(1)
	if !ok {
		t.Fatalf("account was not connected")
	}
	gameWorld.SetInput(1, game.ShipInput{ThrustForward: true})
	snapshot := gameWorld.Tick(1.0 / 30.0)
	object, ok := findCosmicObjectInSnapshot(snapshot, objectID)
	if !ok {
		t.Fatalf("object %v not found in snapshot", objectID)
	}

	if object.VelocityY <= 0 {
		t.Fatalf("got velocity Y %v, want positive", object.VelocityY)
	}
	if serverData.CosmicObjects.Items[1].Y <= 0 {
		t.Fatalf("got stored Y %v, want positive", serverData.CosmicObjects.Items[1].Y)
	}
}

func TestChangeControlledShipToRandomModelUsesOnlyOtherShipModels(t *testing.T) {
	serverData := testWorldData(t)
	gameWorld := world.New(1, serverData)

	objectID, ok := gameWorld.ConnectAccount(1)
	if !ok {
		t.Fatalf("account was not connected")
	}
	if !gameWorld.ChangeControlledShipToRandomModel(1) {
		t.Fatalf("ship model was not changed")
	}
	snapshot := gameWorld.SnapshotForAccount(1)
	object, ok := findCosmicObjectInSnapshot(snapshot, objectID)
	if !ok {
		t.Fatalf("object %v not found in snapshot", objectID)
	}

	if object.CosmicObjectModelID != 4 {
		t.Fatalf("got model %v, want another ship model 4", object.CosmicObjectModelID)
	}
	if object.Mass != 12 || object.Capacity != 20 || object.MaxArmor != 150 || object.MaxSpeed != 600 || object.MaxAngularSpeed != 4 {
		t.Fatalf("ship stats were not copied from selected model: %+v", object)
	}
	if len(serverData.CosmicObjects.GetByCosmicObjectModelID(1)) != 0 {
		t.Fatalf("old model index still contains changed object")
	}
	if len(serverData.CosmicObjects.GetByCosmicObjectModelID(4)) != 1 {
		t.Fatalf("new model index does not contain changed object")
	}
}

func TestDisconnectAccountDoesNotDeleteStoredShip(t *testing.T) {
	gameWorld := world.New(1, testWorldData(t))

	objectID, ok := gameWorld.ConnectAccount(1)
	if !ok {
		t.Fatalf("account was not connected")
	}
	gameWorld.DisconnectAccount(1)
	snapshot := gameWorld.Tick(1.0 / 30.0)

	if _, ok := findCosmicObjectInSnapshot(snapshot, objectID); !ok {
		t.Fatalf("object %v was removed after disconnect", objectID)
	}
}

func TestSnapshotContainsObjectsLoadedFromData(t *testing.T) {
	gameWorld := world.New(1, testWorldData(t))

	snapshot := gameWorld.Tick(1.0 / 30.0)
	models := map[int64]bool{}
	for _, object := range snapshot.Objects {
		models[object.CosmicObjectModelID] = true
	}

	if !models[1] {
		t.Fatalf("snapshot does not contain model 1")
	}
	if !models[2] {
		t.Fatalf("snapshot does not contain model 2")
	}
	if !models[3] {
		t.Fatalf("snapshot does not contain model 3")
	}
}

func TestSaveDataWritesCosmicObjectPosition(t *testing.T) {
	serverData := testWorldData(t)
	gameWorld := world.New(1, serverData)
	workingDirectory := t.TempDir()
	if err := os.Mkdir(filepath.Join(workingDirectory, "data"), 0o700); err != nil {
		t.Fatal(err)
	}

	if _, ok := gameWorld.ConnectAccount(1); !ok {
		t.Fatalf("account was not connected")
	}
	gameWorld.SetInput(1, game.ShipInput{ThrustForward: true})
	gameWorld.Tick(1)
	if err := gameWorld.SaveData(workingDirectory); err != nil {
		t.Fatal(err)
	}

	loadedCosmicObjects := data.NewCosmicObjects()
	if err := loadedCosmicObjects.LoadFromFile(filepath.Join(workingDirectory, "data", "CosmicObjects.json")); err != nil {
		t.Fatal(err)
	}
	if loadedCosmicObjects.Items[1].Y <= 0 {
		t.Fatalf("got saved Y %v, want positive", loadedCosmicObjects.Items[1].Y)
	}
}
