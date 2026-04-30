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

// собирает все загруженные справочники и игровые сущности, нужные миру.
type Data struct {
	Accounts           *data.Accounts
	Characters         *data.Characters
	CosmicObjects      *data.CosmicObjects
	CosmicObjectTypes  *data.CosmicObjectTypes
	CosmicObjectModels *data.CosmicObjectModels
	Itemtypes          *data.Itemtypes
}

// хранит состояние, которое участвует в симуляции, но пока не сериализуется в JSON-данные.
type objectRuntime struct {
	Velocity       game.WorldVector
	TargetRotation float64
}

// управляет подключенными аккаунтами, входами игроков и пошаговой симуляцией объектов.
type World struct {
	mu               sync.Mutex
	tick             int64
	data             Data
	accountObjectIDs map[int64]int64
	inputs           map[int64]game.ShipInput
	runtime          map[int64]objectRuntime
}

// создает мир поверх уже загруженных серверных данных.
func New(seed int64, serverData Data) *World {
	return &World{
		data:             serverData,
		accountObjectIDs: map[int64]int64{},
		inputs:           map[int64]game.ShipInput{},
		runtime:          map[int64]objectRuntime{},
	}
}

// привязывает аккаунт к текущему объекту его персонажа и разрешает получать ввод.
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

// убирает активную привязку аккаунта и последний ввод игрока.
func (world *World) DisconnectAccount(accountID int64) {
	world.mu.Lock()
	defer world.mu.Unlock()

	delete(world.accountObjectIDs, accountID)
	delete(world.inputs, accountID)
}

// сохраняет последний пакет управления только для уже подключенного аккаунта.
func (world *World) SetInput(accountID int64, input game.ShipInput) {
	world.mu.Lock()
	defer world.mu.Unlock()

	if _, ok := world.accountObjectIDs[accountID]; !ok {
		return
	}

	world.inputs[accountID] = input
}

// выполняет один шаг симуляции и возвращает общий снимок мира.
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

// возвращает снимок мира с заполненным ID объекта текущего игрока.
func (world *World) SnapshotForAccount(accountID int64) game.Snapshot {
	world.mu.Lock()
	defer world.mu.Unlock()

	objectID := world.accountObjectIDs[accountID]
	return world.snapshotLocked(objectID)
}

// возвращает объект, которым сейчас управляет подключенный аккаунт.
func (world *World) ObjectIDForAccount(accountID int64) (int64, bool) {
	world.mu.Lock()
	defer world.mu.Unlock()

	objectID, ok := world.accountObjectIDs[accountID]
	return objectID, ok
}

// сохраняет изменяемое состояние мира обратно в JSON-файлы сервера.
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

// собирает детерминированно отсортированный снимок; вызывается только под mutex.
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

// склеивает сохраненное JSON-состояние и runtime-поля в объект симуляции.
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

// переводит модель из слоя данных в модель, понятную физике и клиенту.
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

// записывает результат физики в постоянные поля и runtime-кэш объекта.
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

// переводит пару противоположных кнопок в силу по одной локальной оси.
func axisForce(positive bool, negative bool, maxForce float64) float64 {
	if positive == negative {
		return 0
	}
	if positive {
		return maxForce
	}
	return -maxForce
}

// возвращает модуль скорости объекта для сохранения в данных.
func objectVelocityLength(velocity game.WorldVector) float64 {
	return math.Hypot(velocity.X, velocity.Y)
}
