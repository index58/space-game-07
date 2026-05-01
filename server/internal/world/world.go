package world

import (
	"fmt"
	"math/rand"
	"path/filepath"
	"sort"
	"sync"

	"space-game-07-server/internal/data"
	"space-game-07-server/internal/game"
	"space-game-07-server/internal/physics"
)

const (
	defaultAccountEmailDomain = "auto.local"
	defaultAccountPassword    = "auto"
	defaultStarterShipAcronym = "ship_bat"
	defaultStarterShipForce   = 1287901.529888449
	defaultStarterShipTorque  = 653565
)

// Собирает все загруженные справочники и игровые сущности, нужные миру.
type Data struct {
	Accounts           *data.Accounts           // Учетные записи, доступные игровой симуляции.
	Characters         *data.Characters         // Персонажи, доступные игровой симуляции.
	CosmicObjects      *data.CosmicObjects      // Экземпляры объектов, которые участвуют в мире.
	CosmicObjectTypes  *data.CosmicObjectTypes  // Справочник типов объектов для правил мира.
	CosmicObjectModels *data.CosmicObjectModels // Справочник моделей объектов для физики и отображения.
	Itemtypes          *data.Itemtypes          // Справочник типов предметов для серверной логики.
}

// Управляет подключенными аккаунтами, входами игроков и пошаговой симуляцией объектов.
type World struct {
	mu               sync.Mutex               // Защищает изменяемое состояние мира от параллельных горутин.
	tick             int64                    // Номер последнего выполненного шага симуляции.
	data             Data                     // Справочники и игровые сущности, которыми управляет мир.
	accountObjectIDs map[int64]int64          // Связь подключенных аккаунтов с управляемыми объектами.
	inputs           map[int64]game.ShipInput // Последний принятый ввод для каждого подключенного аккаунта.
	random           *rand.Rand               // Источник случайности для команд, воспроизводимых по начальному зерну.
}

// Создает мир поверх уже загруженных серверных данных.
func New(seed int64, serverData Data) *World {
	return &World{
		data:             serverData,
		accountObjectIDs: map[int64]int64{},
		inputs:           map[int64]game.ShipInput{},
		random:           rand.New(rand.NewSource(seed)),
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

// Создает учетную запись, персонажа и стартовый корабль для первого входа.
func (world *World) CreateStarterAccount() (*data.Account, error) {
	world.mu.Lock()
	defer world.mu.Unlock()

	model, ok := world.data.CosmicObjectModels.GetByAcronym(defaultStarterShipAcronym)
	if !ok {
		return nil, fmt.Errorf("starter ship model %q not found", defaultStarterShipAcronym)
	}

	nextAccountID := world.data.Accounts.MaxID + 1
	account, err := world.data.Accounts.Add(&data.Account{
		Email:        fmt.Sprintf("auto%d@%s", nextAccountID, defaultAccountEmailDomain),
		Nickname:     fmt.Sprintf("Pilot%d", nextAccountID),
		PasswordHash: defaultAccountPassword,
	})
	if err != nil {
		return nil, err
	}

	character, err := world.data.Characters.Add(&data.Character{
		AccountID: account.ID,
	})
	if err != nil {
		world.data.Accounts.Delete(account.ID)
		return nil, err
	}

	cosmicObject, err := world.data.CosmicObjects.Add(&data.CosmicObject{
		Title:               model.TitleRu,
		CosmicObjectModelID: model.ID,
		OwnerCharacterID:    character.ID,
		CreatorCharacterID:  character.ID,
		Mass:                model.Mass,
		Capacity:            model.Capacity,
		MaxArmor:            model.MaxArmor,
		MaxSpeed:            model.MaxSpeed,
		MaxAngularSpeed:     model.MaxAngularSpeed,
		Armor:               model.MaxArmor,
		MaxAlongForce:       defaultStarterShipForce,
		MaxAcrossForce:      defaultStarterShipForce,
		MaxTorque:           defaultStarterShipTorque,
		Enabled:             true,
		TargetRotation:      0,
	})
	if err != nil {
		world.data.Characters.Delete(character.ID)
		world.data.Accounts.Delete(account.ID)
		return nil, err
	}

	character.LocationCosmicObjectID = cosmicObject.ID
	if err := world.data.Accounts.SetCurrentCharacter(account.ID, character.ID); err != nil {
		world.data.CosmicObjects.Delete(cosmicObject.ID)
		world.data.Characters.Delete(character.ID)
		world.data.Accounts.Delete(account.ID)
		return nil, err
	}

	return account, nil
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

// Меняет управляемый объект на другую случайную модель корабля из справочника.
func (world *World) ChangeControlledShipToRandomModel(accountID int64) bool {
	world.mu.Lock()
	defer world.mu.Unlock()

	objectID, ok := world.accountObjectIDs[accountID]
	if !ok {
		return false
	}

	cosmicObject, ok := world.data.CosmicObjects.Get(objectID)
	if !ok {
		return false
	}

	shipType, ok := world.data.CosmicObjectTypes.GetByAcronym("Ship")
	if !ok {
		return false
	}

	candidateIDs := make([]int64, 0)
	for modelID, model := range world.data.CosmicObjectModels.Items {
		if model.CosmicObjectTypeID == shipType.ID && modelID != cosmicObject.CosmicObjectModelID {
			candidateIDs = append(candidateIDs, modelID)
		}
	}
	if len(candidateIDs) == 0 {
		return false
	}
	sort.Slice(candidateIDs, func(left int, right int) bool {
		return candidateIDs[left] < candidateIDs[right]
	})

	model, ok := world.data.CosmicObjectModels.Get(candidateIDs[world.random.Intn(len(candidateIDs))])
	if !ok {
		return false
	}

	cosmicObject.Title = model.TitleRu
	cosmicObject.CosmicObjectModelID = model.ID
	cosmicObject.Mass = model.Mass
	cosmicObject.Capacity = model.Capacity
	cosmicObject.MaxArmor = model.MaxArmor
	cosmicObject.Armor = model.MaxArmor
	cosmicObject.MaxSpeed = model.MaxSpeed
	cosmicObject.MaxAngularSpeed = model.MaxAngularSpeed

	return world.data.CosmicObjects.RebuildIndexes() == nil
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

// Возвращает персонажа по идентификатору из защищенного состояния.
func (world *World) CharacterByID(id int64) (*data.Character, bool) {
	world.mu.Lock()
	defer world.mu.Unlock()

	return world.data.Characters.Get(id)
}

// Возвращает космический объект по идентификатору из защищенного состояния.
func (world *World) CosmicObjectByID(id int64) (*data.CosmicObject, bool) {
	world.mu.Lock()
	defer world.mu.Unlock()

	return world.data.CosmicObjects.Get(id)
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

	objects := make([]data.CosmicObject, 0, len(objectIDs))
	for _, objectID := range objectIDs {
		cosmicObject, ok := world.data.CosmicObjects.Get(objectID)
		if !ok {
			continue
		}
		objects = append(objects, *cosmicObject)
	}

	return game.Snapshot{
		Type:         "snapshot",
		Tick:         world.tick,
		SelfObjectID: selfObjectID,
		Objects:      objects,
	}
}
