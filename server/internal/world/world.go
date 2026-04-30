package world

import (
	"path/filepath"
	"sort"
	"strings"
	"sync"

	"space-game-07-server/internal/data"
	"space-game-07-server/internal/game"
	"space-game-07-server/internal/physics"
)

// Собирает все загруженные справочники и игровые сущности, нужные миру.
type Data struct {
	Accounts           *data.Accounts
	Characters         *data.Characters
	CosmicObjects      *data.CosmicObjects
	CosmicObjectTypes  *data.CosmicObjectTypes
	CosmicObjectModels *data.CosmicObjectModels
	Itemtypes          *data.Itemtypes
}

// Управляет подключенными аккаунтами, входами игроков и пошаговой симуляцией объектов.
type World struct {
	mu               sync.Mutex
	tick             int64
	data             Data
	accountObjectIDs map[int64]int64
	inputs           map[int64]game.ShipInput
}

// Создает мир поверх уже загруженных серверных данных.
func New(seed int64, serverData Data) *World {
	return &World{
		data:             serverData,
		accountObjectIDs: map[int64]int64{},
		inputs:           map[int64]game.ShipInput{},
	}
}

// Привязывает аккаунт к текущему объекту его персонажа и разрешает получать ввод.
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

// Убирает активную привязку аккаунта и последний ввод игрока.
func (world *World) DisconnectAccount(accountID int64) {
	world.mu.Lock()
	defer world.mu.Unlock()

	delete(world.accountObjectIDs, accountID)
	delete(world.inputs, accountID)
}

// Сохраняет последний пакет управления только для уже подключенного аккаунта.
func (world *World) SetInput(accountID int64, input game.ShipInput) {
	world.mu.Lock()
	defer world.mu.Unlock()

	if _, ok := world.accountObjectIDs[accountID]; !ok {
		return
	}

	world.inputs[accountID] = input
}

// Выполняет один шаг симуляции и возвращает общий снимок мира.
func (world *World) Tick(dtSeconds float64) game.Snapshot {
	world.mu.Lock()
	defer world.mu.Unlock()

	for accountID, objectID := range world.accountObjectIDs {
		cosmicObject, ok := world.data.CosmicObjects.Get(objectID)
		if !ok || cosmicObject.Anchored || !cosmicObject.Enabled {
			continue
		}

		model, ok := world.data.CosmicObjectModels.Get(cosmicObject.CosmicObjectModelID)
		if !ok {
			continue
		}

		next := physics.StepShip(*cosmicObject, *model, world.inputs[accountID], dtSeconds)
		*cosmicObject = next
	}

	world.tick++
	return world.snapshotLocked(0)
}

// Возвращает снимок мира с заполненным ID объекта текущего игрока.
func (world *World) SnapshotForAccount(accountID int64) game.Snapshot {
	world.mu.Lock()
	defer world.mu.Unlock()

	objectID := world.accountObjectIDs[accountID]
	return world.snapshotLocked(objectID)
}

// Возвращает объект, которым сейчас управляет подключенный аккаунт.
func (world *World) ObjectIDForAccount(accountID int64) (int64, bool) {
	world.mu.Lock()
	defer world.mu.Unlock()

	objectID, ok := world.accountObjectIDs[accountID]
	return objectID, ok
}

// Сохраняет изменяемое состояние мира обратно в JSON-файлы сервера.
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

// Собирает детерминированно отсортированный снимок; вызывается только под mutex.
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
		model, ok := world.data.CosmicObjectModels.Get(cosmicObject.CosmicObjectModelID)
		if !ok {
			continue
		}
		objects = append(objects, world.snapshotObjectLocked(cosmicObject, model))
	}

	return game.Snapshot{
		Type:         "snapshot",
		Tick:         world.tick,
		SelfObjectID: selfObjectID,
		Objects:      objects,
	}
}

// Переводит объект данных в сетевой DTO без промежуточной игровой модели.
func (world *World) snapshotObjectLocked(cosmicObject *data.CosmicObject, model *data.CosmicObjectModel) game.SnapshotObject {
	return game.SnapshotObject{
		ID:              cosmicObject.ID,
		ModelAcronym:    model.Acronym,
		Kind:            world.objectKindLocked(model),
		TextureScale:    model.TextureScale,
		X:               cosmicObject.X,
		Y:               cosmicObject.Y,
		VelocityX:       cosmicObject.VelocityX,
		VelocityY:       cosmicObject.VelocityY,
		Rotation:        cosmicObject.Rotation,
		AngularVelocity: cosmicObject.AngularSpeed,
		TargetRotation:  cosmicObject.TargetRotation,
	}
}

// Возвращает клиентскую категорию объекта из справочника типов.
func (world *World) objectKindLocked(model *data.CosmicObjectModel) string {
	cosmicObjectType, ok := world.data.CosmicObjectTypes.Get(model.CosmicObjectTypeID)
	if !ok {
		return ""
	}

	return strings.ToLower(cosmicObjectType.Acronym)
}
