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
)

// Собирает справочники и игровые сущности, нужные симуляции мира.
type Data struct {
	Accounts                *data.Accounts                // Учетные записи, доступные игровой симуляции.
	Characters              *data.Characters              // Персонажи, доступные игровой симуляции.
	CosmicObjects           *data.CosmicObjects           // Экземпляры объектов, участвующие в мире.
	CosmicObjectTypes       *data.CosmicObjectTypes       // Справочник типов объектов для правил мира.
	CosmicObjectModels      *data.CosmicObjectModels      // Справочник моделей объектов для физики и отображения.
	Itemtypes               *data.Itemtypes               // Справочник типов предметов для серверной логики.
	EquipmentGroups         *data.EquipmentGroups         // Группы оборудования, установленные на объектах мира.
	Assemblies              *data.Assemblies              // Справочник сборок для расчета характеристик кораблей.
	AssemblyEquipmentGroups *data.AssemblyEquipmentGroups // Группы оборудования, заданные в сборках.
}

// Управляет подключенными аккаунтами, вводом игроков и пошаговой симуляцией объектов.
type World struct {
	mu               sync.Mutex               // Защищает изменяемое состояние мира от параллельных горутин.
	tick             int64                    // Номер последнего выполненного шага симуляции.
	data             Data                     // Справочники и игровые сущности, которыми управляет мир.
	accountObjectIDs map[int64]int64          // Связь подключенных аккаунтов с управляемыми объектами.
	inputs           map[int64]game.ShipInput // Последний принятый ввод для каждого подключенного аккаунта.
	random           *rand.Rand               // Источник случайности для воспроизводимых команд.
}

// Создает мир поверх уже загруженных серверных данных.
func New(seed int64, serverData Data) *World {
	created := &World{
		data:             serverData,
		accountObjectIDs: map[int64]int64{},
		inputs:           map[int64]game.ShipInput{},
		random:           rand.New(rand.NewSource(seed)),
	}
	created.applyAssembliesToLoadedShips()
	return created
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
	assembly, ok := world.firstPublicDeveloperAssembly(model.ID)
	if !ok {
		return nil, fmt.Errorf("public developer assembly for starter ship model %q not found", defaultStarterShipAcronym)
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

	cosmicObject := world.cosmicObjectFromModelAndAssembly(model, assembly)
	cosmicObject.OwnerCharacterID = character.ID
	cosmicObject.CreatorCharacterID = character.ID
	createdObject, err := world.data.CosmicObjects.Add(cosmicObject)
	if err != nil {
		world.data.Characters.Delete(character.ID)
		world.data.Accounts.Delete(account.ID)
		return nil, err
	}
	if err := world.replaceEquipmentFromAssembly(createdObject.ID, assembly); err != nil {
		world.data.CosmicObjects.Delete(createdObject.ID)
		world.data.Characters.Delete(character.ID)
		world.data.Accounts.Delete(account.ID)
		return nil, err
	}

	character.LocationCosmicObjectID = createdObject.ID
	if err := world.data.Accounts.SetCurrentCharacter(account.ID, character.ID); err != nil {
		world.data.CosmicObjects.Delete(createdObject.ID)
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
			if _, ok := world.firstPublicDeveloperAssembly(modelID); ok {
				candidateIDs = append(candidateIDs, modelID)
			}
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
	assembly, ok := world.firstPublicDeveloperAssembly(model.ID)
	if !ok {
		return false
	}

	world.applyModelAndAssembly(cosmicObject, model, assembly)
	if err := world.replaceEquipmentFromAssembly(cosmicObject.ID, assembly); err != nil {
		return false
	}

	return world.data.CosmicObjects.RebuildIndexes() == nil
}

// Выполняет один шаг симуляции и возвращает общий снимок мира.
func (world *World) Tick(dtSeconds float64) game.Snapshot {
	world.mu.Lock()
	defer world.mu.Unlock()

	accountIDs := make([]int64, 0, len(world.accountObjectIDs))
	for accountID := range world.accountObjectIDs {
		accountIDs = append(accountIDs, accountID)
	}
	sort.Slice(accountIDs, func(left int, right int) bool {
		return accountIDs[left] < accountIDs[right]
	})

	for _, accountID := range accountIDs {
		objectID := world.accountObjectIDs[accountID]
		cosmicObject, ok := world.data.CosmicObjects.Get(objectID)
		if !ok || cosmicObject.Anchored || !cosmicObject.Enabled {
			continue
		}

		model, ok := world.data.CosmicObjectModels.Get(cosmicObject.CosmicObjectModelID)
		if !ok {
			continue
		}

		next := physics.StepShip(*cosmicObject, *model, world.inputs[accountID], dtSeconds)
		next = world.resolveCollisions(next, cosmicObject.ID)
		*cosmicObject = next
	}

	world.tick++
	return world.snapshotLocked(0)
}

// Раздвигает движущийся объект со всеми включенными телами мира.
func (world *World) resolveCollisions(moving data.CosmicObject, movingID int64) data.CosmicObject {
	movingModel, ok := world.data.CosmicObjectModels.Get(moving.CosmicObjectModelID)
	if !ok {
		return moving
	}

	objectIDs := make([]int64, 0, len(world.data.CosmicObjects.Items))
	for objectID := range world.data.CosmicObjects.Items {
		objectIDs = append(objectIDs, objectID)
	}
	sort.Slice(objectIDs, func(left int, right int) bool {
		return objectIDs[left] < objectIDs[right]
	})

	for _, objectID := range objectIDs {
		if objectID == movingID {
			continue
		}
		obstacle, ok := world.data.CosmicObjects.Get(objectID)
		if !ok {
			continue
		}
		obstacleModel, ok := world.data.CosmicObjectModels.Get(obstacle.CosmicObjectModelID)
		if !ok {
			continue
		}
		correction, collided := physics.CollisionCorrection(moving, *movingModel, *obstacle, *obstacleModel)
		if !collided {
			continue
		}
		nextMoving, nextObstacle := physics.ApplyCollisionResponse(moving, *obstacle, correction)
		moving = nextMoving
		*obstacle = nextObstacle
	}

	return moving
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
	if world.data.Assemblies != nil {
		if err := world.data.Assemblies.SaveToFile(filepath.Join(dataDirectory, "Assemblies.json")); err != nil {
			return err
		}
	}
	if world.data.EquipmentGroups != nil {
		if err := world.data.EquipmentGroups.SaveToFile(filepath.Join(dataDirectory, "EquipmentGroups.json")); err != nil {
			return err
		}
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

// Ищет первую публичную системную сборку для модели корпуса.
func (world *World) firstPublicDeveloperAssembly(cosmicObjectModelID int64) (*data.Assembly, bool) {
	if world.data.Assemblies == nil {
		return nil, false
	}
	return world.data.Assemblies.FirstPublicDeveloperAssembly(cosmicObjectModelID)
}

// Обновляет сохраненные корабли по системным публичным сборкам, сохраняя их движение и владельцев.
func (world *World) applyAssembliesToLoadedShips() {
	if world.data.CosmicObjects == nil || world.data.CosmicObjectModels == nil || world.data.CosmicObjectTypes == nil {
		return
	}

	shipType, ok := world.data.CosmicObjectTypes.GetByAcronym("Ship")
	if !ok {
		return
	}

	for _, cosmicObject := range world.data.CosmicObjects.Items {
		model, ok := world.data.CosmicObjectModels.Get(cosmicObject.CosmicObjectModelID)
		if !ok || model.CosmicObjectTypeID != shipType.ID {
			continue
		}
		assembly, ok := world.firstPublicDeveloperAssembly(model.ID)
		if !ok {
			continue
		}
		world.applyModelAndAssembly(cosmicObject, model, assembly)
		_ = world.ensureEquipmentFromAssembly(cosmicObject.ID, assembly)
	}
}

// Собирает новый объект из модели корпуса и рассчитанной сборки.
func (world *World) cosmicObjectFromModelAndAssembly(model *data.CosmicObjectModel, assembly *data.Assembly) *data.CosmicObject {
	cosmicObject := &data.CosmicObject{Enabled: true}
	world.applyModelAndAssembly(cosmicObject, model, assembly)
	cosmicObject.Armor = assembly.MaxArmor
	cosmicObject.Fuel = assembly.MaxFuel
	return cosmicObject
}

// Применяет рассчитанные характеристики сборки, не трогая движение и владение объектом.
func (world *World) applyModelAndAssembly(cosmicObject *data.CosmicObject, model *data.CosmicObjectModel, assembly *data.Assembly) {
	cosmicObject.Title = model.TitleRu
	cosmicObject.CosmicObjectModelID = model.ID
	cosmicObject.Mass = assembly.Mass
	cosmicObject.Capacity = model.Capacity
	cosmicObject.MaxArmor = assembly.MaxArmor
	cosmicObject.MaxSpeed = model.MaxSpeed
	cosmicObject.MaxAngularSpeed = model.MaxAngularSpeed
	cosmicObject.MaxAlongForce = assembly.MaxAlongForce
	cosmicObject.MaxAcrossForce = assembly.MaxAcrossForce
	cosmicObject.MaxTorque = assembly.MaxTorque
	cosmicObject.GeneratingPower = assembly.GeneratingPower
	cosmicObject.ConsumingPower = assembly.ConsumingPower
	cosmicObject.Complexity = assembly.Complexity
	cosmicObject.OccupiedVolume = assembly.OccupiedVolume
	cosmicObject.MaxFuel = assembly.MaxFuel
	if cosmicObject.Armor > assembly.MaxArmor {
		cosmicObject.Armor = assembly.MaxArmor
	}
	if cosmicObject.Fuel > assembly.MaxFuel {
		cosmicObject.Fuel = assembly.MaxFuel
	}
}

// Устанавливает оборудование из сборки, если у объекта еще нет оборудования.
func (world *World) ensureEquipmentFromAssembly(cosmicObjectID int64, assembly *data.Assembly) error {
	if world.data.EquipmentGroups == nil || len(world.data.EquipmentGroups.GetByCosmicObjectID(cosmicObjectID)) > 0 {
		return nil
	}
	return world.installEquipmentFromAssembly(cosmicObjectID, assembly)
}

// Заменяет оборудование объекта на оборудование выбранной сборки.
func (world *World) replaceEquipmentFromAssembly(cosmicObjectID int64, assembly *data.Assembly) error {
	if world.data.EquipmentGroups == nil {
		return nil
	}
	world.data.EquipmentGroups.DeleteByCosmicObjectID(cosmicObjectID)
	return world.installEquipmentFromAssembly(cosmicObjectID, assembly)
}

// Копирует группы оборудования сборки на конкретный объект.
func (world *World) installEquipmentFromAssembly(cosmicObjectID int64, assembly *data.Assembly) error {
	if world.data.EquipmentGroups == nil || world.data.AssemblyEquipmentGroups == nil {
		return nil
	}

	for _, group := range world.data.AssemblyEquipmentGroups.GetByAssemblyID(assembly.ID) {
		if _, err := world.data.EquipmentGroups.Add(&data.EquipmentGroup{
			CosmicObjectID:       cosmicObjectID,
			Title:                group.Title,
			EquipmentItemModelID: group.EquipmentItemModelID,
			Count:                group.Count,
			EnabledCount:         group.Count,
			Enabled:              true,
			Active:               true,
		}); err != nil {
			return err
		}
	}
	return nil
}
