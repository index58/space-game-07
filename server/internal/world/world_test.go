package world_test

import (
	"math"
	"testing"

	"space-game-07-server/internal/game"
	"space-game-07-server/internal/world"
)

func findSnapshotObject(snapshot game.Snapshot, objectID int64) (game.SnapshotObject, bool) {
	for _, object := range snapshot.Objects {
		if object.ID == objectID {
			return object, true
		}
	}

	return game.SnapshotObject{}, false
}

func TestAddPlayerCreatesDifferentShipObjects(t *testing.T) {
	gameWorld := world.New(1)

	_, firstObjectID := gameWorld.AddPlayer()
	_, secondObjectID := gameWorld.AddPlayer()

	if firstObjectID == secondObjectID {
		t.Fatalf("got same object ID %v", firstObjectID)
	}
}

func TestAddPlayerSpawnsInsideFiveHundredMeterRadius(t *testing.T) {
	gameWorld := world.New(1)

	_, objectID := gameWorld.AddPlayer()
	snapshot := gameWorld.Tick(1.0 / 30.0)
	object, ok := findSnapshotObject(snapshot, objectID)
	if !ok {
		t.Fatalf("object %v not found in snapshot", objectID)
	}

	if math.Hypot(object.X, object.Y) > 500 {
		t.Fatalf("spawn position (%v, %v) is outside radius 500", object.X, object.Y)
	}
}

func TestRemovePlayerRemovesShipFromSnapshot(t *testing.T) {
	gameWorld := world.New(1)

	playerID, objectID := gameWorld.AddPlayer()
	gameWorld.RemovePlayer(playerID)
	snapshot := gameWorld.Tick(1.0 / 30.0)

	if _, ok := findSnapshotObject(snapshot, objectID); ok {
		t.Fatalf("object %v still exists after player removal", objectID)
	}
}

func TestTickAppliesLatestPlayerInput(t *testing.T) {
	gameWorld := world.New(1)

	playerID, objectID := gameWorld.AddPlayer()
	gameWorld.SetInput(playerID, game.ShipInput{ThrustForward: true})
	snapshot := gameWorld.Tick(1.0 / 30.0)
	object, ok := findSnapshotObject(snapshot, objectID)
	if !ok {
		t.Fatalf("object %v not found in snapshot", objectID)
	}

	if object.VelocityY <= 0 {
		t.Fatalf("got velocity Y %v, want positive", object.VelocityY)
	}
}

func TestSnapshotContainsStaticObjects(t *testing.T) {
	gameWorld := world.New(1)

	snapshot := gameWorld.Tick(1.0 / 30.0)
	models := map[string]bool{}
	for _, object := range snapshot.Objects {
		models[object.ModelAcronym] = true
	}

	if !models["asteroid_0002"] {
		t.Fatalf("snapshot does not contain asteroid_0002")
	}
	if !models["station_tiny_crumb"] {
		t.Fatalf("snapshot does not contain station_tiny_crumb")
	}
}
