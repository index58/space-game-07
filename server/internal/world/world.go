package world

import (
	"math"
	"path/filepath"
	"sort"
	"sync"

	"space-game-07-server/internal/data"
	"space-game-07-server/internal/game"
	"space-game-07-server/internal/physics"
)

const modelMassScale = 1000

type Data struct {
	Accounts           *data.Accounts
	Characters         *data.Characters
	CosmicObjects      *data.CosmicObjects
	CosmicObjectTypes  *data.CosmicObjectTypes
	CosmicObjectModels *data.CosmicObjectModels
	Itemtypes          *data.Itemtypes
}

type objectRuntime struct {
	Velocity       game.WorldVector
	TargetRotation float64
}

type World struct {
	mu               sync.Mutex
	tick             int64
	data             Data
	accountObjectIDs map[int64]int64
	inputs           map[int64]game.ShipInput
	runtime          map[int64]objectRuntime
}

func New(seed int64, serverData Data) *World {
	return &World{
		data:             serverData,
		accountObjectIDs: map[int64]int64{},
		inputs:           map[int64]game.ShipInput{},
		runtime:          map[int64]objectRuntime{},
	}
}

func (world *World) ConnectAccount(accountID int64) (int64, bool) {
	world.mu.Lock()
	defer world.mu.Unlock()

	account, ok := world.data.Accounts.Get(accountID)
	if !ok || account.CurrentCharacterID <= 0 {
		return 0, false
	}

	character, ok := world.data.Characters.Get(account.CurrentCharacterID)
	if !ok || character.AccountID != account.ID || character.LocationCosmicObjectID <= 0 {
		return 0, false
	}

	if _, ok := world.data.CosmicObjects.Get(character.LocationCosmicObjectID); !ok {
		return 0, false
	}

	world.accountObjectIDs[accountID] = character.LocationCosmicObjectID
	world.inputs[accountID] = game.ShipInput{}
	return character.LocationCosmicObjectID, true
}

func (world *World) DisconnectAccount(accountID int64) {
	world.mu.Lock()
	defer world.mu.Unlock()

	delete(world.accountObjectIDs, accountID)
	delete(world.inputs, accountID)
}

func (world *World) SetInput(accountID int64, input game.ShipInput) {
	world.mu.Lock()
	defer world.mu.Unlock()

	if _, ok := world.accountObjectIDs[accountID]; !ok {
		return
	}

	world.inputs[accountID] = input
}

func (world *World) Tick(dtSeconds float64) game.Snapshot {
	world.mu.Lock()
	defer world.mu.Unlock()

	for accountID, objectID := range world.accountObjectIDs {
		cosmicObject, ok := world.data.CosmicObjects.Get(objectID)
		if !ok || cosmicObject.Anchored || !cosmicObject.Enabled {
			continue
		}

		runtimeObject, ok := world.gameObjectLocked(cosmicObject)
		if !ok {
			continue
		}

		input := world.inputs[accountID]
		next := physics.StepShip(runtimeObject, input, dtSeconds)
		world.saveObjectStateLocked(cosmicObject, next, input)
	}

	world.tick++
	return world.snapshotLocked(0)
}

func (world *World) SnapshotForAccount(accountID int64) game.Snapshot {
	world.mu.Lock()
	defer world.mu.Unlock()

	objectID := world.accountObjectIDs[accountID]
	return world.snapshotLocked(objectID)
}

func (world *World) ObjectIDForAccount(accountID int64) (int64, bool) {
	world.mu.Lock()
	defer world.mu.Unlock()

	objectID, ok := world.accountObjectIDs[accountID]
	return objectID, ok
}

func (world *World) SaveData(workingDirectory string) error {
	world.mu.Lock()
	defer world.mu.Unlock()

	dataDirectory := filepath.Join(workingDirectory, "data")
	if err := world.data.Accounts.SaveToFile(filepath.Join(dataDirectory, "Accounts.json")); err != nil {
		return err
	}
	if err := world.data.Characters.SaveToFile(filepath.Join(dataDirectory, "Characters.json")); err != nil {
		return err
	}
	if err := world.data.CosmicObjects.SaveToFile(filepath.Join(dataDirectory, "CosmicObjects.json")); err != nil {
		return err
	}
	if err := world.data.CosmicObjectTypes.SaveToFile(filepath.Join(dataDirectory, "CosmicObjectTypes.json")); err != nil {
		return err
	}
	if err := world.data.CosmicObjectModels.SaveToFile(filepath.Join(dataDirectory, "CosmicObjectModels.json")); err != nil {
		return err
	}
	if world.data.Itemtypes != nil {
		return world.data.Itemtypes.SaveToFile(filepath.Join(dataDirectory, "Itemtypes.json"))
	}
	return nil
}

func (world *World) snapshotLocked(selfObjectID int64) game.Snapshot {
	objectIDs := make([]int64, 0, len(world.data.CosmicObjects.Items))
	for objectID := range world.data.CosmicObjects.Items {
		objectIDs = append(objectIDs, objectID)
	}
	sort.Slice(objectIDs, func(left int, right int) bool {
		return objectIDs[left] < objectIDs[right]
	})

	objects := make([]game.SnapshotObject, 0, len(objectIDs))
	for _, objectID := range objectIDs {
		cosmicObject, ok := world.data.CosmicObjects.Get(objectID)
		if !ok {
			continue
		}
		runtimeObject, ok := world.gameObjectLocked(cosmicObject)
		if !ok {
			continue
		}
		objects = append(objects, game.NewSnapshotObject(runtimeObject))
	}

	return game.Snapshot{
		Type:         "snapshot",
		Tick:         world.tick,
		SelfObjectID: selfObjectID,
		Objects:      objects,
	}
}

func (world *World) gameObjectLocked(cosmicObject *data.CosmicObject) (game.WorldObject, bool) {
	model, ok := world.data.CosmicObjectModels.Get(cosmicObject.CosmicObjectModelID)
	if !ok {
		return game.WorldObject{}, false
	}

	runtime := world.runtime[cosmicObject.ID]
	return game.WorldObject{
		ID:              cosmicObject.ID,
		Model:           world.gameModelLocked(cosmicObject, model),
		Position:        game.WorldVector{X: cosmicObject.X, Y: cosmicObject.Y},
		Velocity:        runtime.Velocity,
		Rotation:        cosmicObject.Rotation,
		AngularVelocity: cosmicObject.AngularSpeed,
		TargetRotation:  runtime.TargetRotation,
	}, true
}

func (world *World) gameModelLocked(cosmicObject *data.CosmicObject, model *data.CosmicObjectModel) game.CosmicObjectModel {
	kind := game.ObjectKindStation
	if cosmicObjectType, ok := world.data.CosmicObjectTypes.Get(model.CosmicObjectTypeID); ok {
		switch cosmicObjectType.Acronym {
		case "Ship":
			kind = game.ObjectKindShip
		case "Asteroid":
			kind = game.ObjectKindAsteroid
		case "Station":
			kind = game.ObjectKindStation
		}
	}

	return game.CosmicObjectModel{
		Acronym:                     model.Acronym,
		TitleRu:                     model.TitleRu,
		Kind:                        kind,
		TexturePath:                 model.TextureFilePath,
		TextureWidth:                model.TextureWidth,
		TextureHeight:               model.TextureHeight,
		TextureBodyOriginX:          model.TextureBodyOriginX,
		TextureBodyOriginY:          model.TextureBodyOriginY,
		TextureBodyWidth:            model.TextureBodyWidth,
		TextureBodyLength:           model.TextureBodyLength,
		TextureScale:                model.TextureScale,
		MassKg:                      cosmicObject.Mass * modelMassScale,
		ThrustN:                     cosmicObject.MaxAlongForce,
		MaxSpeedMps:                 cosmicObject.MaxSpeed,
		TorqueNm:                    cosmicObject.MaxTorque,
		MaxAngularSpeedRadPerSecond: cosmicObject.MaxAngularSpeed,
	}
}

func (world *World) saveObjectStateLocked(cosmicObject *data.CosmicObject, object game.WorldObject, input game.ShipInput) {
	cosmicObject.X = object.Position.X
	cosmicObject.Y = object.Position.Y
	cosmicObject.Rotation = object.Rotation
	cosmicObject.Speed = objectVelocityLength(object.Velocity)
	cosmicObject.AngularSpeed = object.AngularVelocity
	cosmicObject.AlongForce = axisForce(input.ThrustForward, input.ThrustBackward, cosmicObject.MaxAlongForce)
	cosmicObject.AcrossForce = axisForce(input.ThrustRight, input.ThrustLeft, cosmicObject.MaxAcrossForce)
	if input.TargetRotationDelta == 0 {
		cosmicObject.Torque = 0
	} else {
		cosmicObject.Torque = object.Model.TorqueNm
	}
	world.runtime[cosmicObject.ID] = objectRuntime{
		Velocity:       object.Velocity,
		TargetRotation: object.TargetRotation,
	}
}

func axisForce(positive bool, negative bool, maxForce float64) float64 {
	if positive == negative {
		return 0
	}
	if positive {
		return maxForce
	}
	return -maxForce
}

func objectVelocityLength(velocity game.WorldVector) float64 {
	return math.Hypot(velocity.X, velocity.Y)
}
