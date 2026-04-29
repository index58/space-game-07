package world

import (
	"math"
	"math/rand"
	"sort"
	"sync"

	"space-game-07-server/internal/game"
	"space-game-07-server/internal/physics"
)

const spawnRadius = 500

type World struct {
	mu              sync.Mutex
	nextPlayerID    int64
	nextObjectID    int64
	tick            int64
	random          *rand.Rand
	playerObjectIDs map[int64]int64
	objects         map[int64]game.WorldObject
	inputs          map[int64]game.ShipInput
	staticObjects   map[int64]game.WorldObject
}

func New(seed int64) *World {
	staticObjects := map[int64]game.WorldObject{}
	maxObjectID := int64(0)
	for _, object := range game.StaticObjects() {
		staticObjects[object.ID] = object
		if object.ID > maxObjectID {
			maxObjectID = object.ID
		}
	}

	return &World{
		nextPlayerID:    1,
		nextObjectID:    maxObjectID + 1,
		random:          rand.New(rand.NewSource(seed)),
		playerObjectIDs: map[int64]int64{},
		objects:         map[int64]game.WorldObject{},
		inputs:          map[int64]game.ShipInput{},
		staticObjects:   staticObjects,
	}
}

func (world *World) AddPlayer() (int64, int64) {
	world.mu.Lock()
	defer world.mu.Unlock()

	playerID := world.nextPlayerID
	world.nextPlayerID++
	objectID := world.nextObjectID
	world.nextObjectID++

	ship := game.NewPlayerShip(objectID, world.randomSpawnPosition())
	world.playerObjectIDs[playerID] = objectID
	world.objects[objectID] = ship
	world.inputs[playerID] = game.ShipInput{}

	return playerID, objectID
}

func (world *World) RemovePlayer(playerID int64) {
	world.mu.Lock()
	defer world.mu.Unlock()

	objectID, ok := world.playerObjectIDs[playerID]
	if !ok {
		return
	}

	delete(world.playerObjectIDs, playerID)
	delete(world.objects, objectID)
	delete(world.inputs, playerID)
}

func (world *World) SetInput(playerID int64, input game.ShipInput) {
	world.mu.Lock()
	defer world.mu.Unlock()

	if _, ok := world.playerObjectIDs[playerID]; !ok {
		return
	}

	world.inputs[playerID] = input
}

func (world *World) Tick(dtSeconds float64) game.Snapshot {
	world.mu.Lock()
	defer world.mu.Unlock()

	for playerID, objectID := range world.playerObjectIDs {
		object, ok := world.objects[objectID]
		if !ok {
			continue
		}

		world.objects[objectID] = physics.StepShip(object, world.inputs[playerID], dtSeconds)
	}

	world.tick++

	return world.snapshotLocked(0)
}

func (world *World) SnapshotForPlayer(selfObjectID int64) game.Snapshot {
	world.mu.Lock()
	defer world.mu.Unlock()

	return world.snapshotLocked(selfObjectID)
}

func (world *World) ObjectIDForPlayer(playerID int64) (int64, bool) {
	world.mu.Lock()
	defer world.mu.Unlock()

	objectID, ok := world.playerObjectIDs[playerID]
	return objectID, ok
}

func (world *World) snapshotLocked(selfObjectID int64) game.Snapshot {
	objectIDs := make([]int64, 0, len(world.objects)+len(world.staticObjects))
	for objectID := range world.objects {
		objectIDs = append(objectIDs, objectID)
	}
	for objectID := range world.staticObjects {
		objectIDs = append(objectIDs, objectID)
	}
	sort.Slice(objectIDs, func(i int, j int) bool {
		return objectIDs[i] < objectIDs[j]
	})

	objects := make([]game.SnapshotObject, 0, len(objectIDs))
	for _, objectID := range objectIDs {
		if object, ok := world.objects[objectID]; ok {
			objects = append(objects, game.NewSnapshotObject(object))
			continue
		}
		objects = append(objects, game.NewSnapshotObject(world.staticObjects[objectID]))
	}

	return game.Snapshot{
		Type:         "snapshot",
		Tick:         world.tick,
		SelfObjectID: selfObjectID,
		Objects:      objects,
	}
}

func (world *World) randomSpawnPosition() game.WorldVector {
	angle := world.random.Float64() * math.Pi * 2
	radius := math.Sqrt(world.random.Float64()) * spawnRadius

	return game.WorldVector{
		X: math.Cos(angle) * radius,
		Y: math.Sin(angle) * radius,
	}
}
