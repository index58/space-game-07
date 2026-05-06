package world

import (
	"fmt"
	"math"
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
	ItemModels              *data.ItemModels              // Справочник моделей предметов для оборудования и содержимого контейнеров.
	EquipmentGroups         *data.EquipmentGroups         // Группы оборудования, установленные на объектах мира.
	ItemGroups              *data.ItemGroups              // Группы предметов внутри контейнерного оборудования.
	Assemblies              *data.Assemblies              // Справочник сборок для расчета характеристик кораблей.
	AssemblyEquipmentGroups *data.AssemblyEquipmentGroups // Группы оборудования, заданные в сборках.
	Chats                   *data.Chats                   // Чаты игрового мира.
	ChatMembers             *data.ChatMembers             // Участники чатов.
	CommunityTypes          *data.CommunityTypes          // Справочник типов сообществ.
	CommunityChatRoles      *data.CommunityChatRoles      // Справочник ролей в чатах сообществ.
	Messages                *data.Messages                // Сообщения чатов.
	MessageReads            *data.MessageReads            // Позиции чтения сообщений персонажами.
	MessageTypes            *data.MessageTypes            // Справочник типов сообщений.
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
	created.ensureChatData()
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

	cosmicObject, ok := world.data.CosmicObjects.Get(character.LocationCosmicObjectID)
	if !ok {
		return 0, false
	}
	cosmicObject.TargetRotation = cosmicObject.Rotation

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
	world.fillShipSupplies(createdObject)

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

	objectID, ok := world.accountObjectIDs[accountID]
	if !ok {
		return
	}

	if input.ToggleAnchor {
		if cosmicObject, ok := world.data.CosmicObjects.Get(objectID); ok {
			if cosmicObject.Anchored || cosmicObjectIsFullyStopped(*cosmicObject) {
				cosmicObject.Anchored = !cosmicObject.Anchored
			}
		}
		input.ToggleAnchor = false
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
	cosmicObject.TargetRotation = cosmicObject.Rotation
	if err := world.replaceEquipmentFromAssembly(cosmicObject.ID, assembly); err != nil {
		return false
	}
	cosmicObject.Armor = cosmicObject.MaxArmor
	world.fillShipSupplies(cosmicObject)

	return world.data.CosmicObjects.RebuildIndexes() == nil
}

// Выполняет один шаг симуляции и возвращает общий снимок мира.
func (world *World) Tick(dtSeconds float64) game.Snapshot {
	world.mu.Lock()
	defer world.mu.Unlock()

	world.stepMovableObjects(dtSeconds, world.inputsByObjectID())
	world.resolveAllCollisions()

	world.tick++
	return world.snapshotLocked(0)
}

// Собирает ввод подключенных аккаунтов по управляемым объектам.
func (world *World) inputsByObjectID() map[int64]game.ShipInput {
	accountIDs := make([]int64, 0, len(world.accountObjectIDs))
	for accountID := range world.accountObjectIDs {
		accountIDs = append(accountIDs, accountID)
	}
	sort.Slice(accountIDs, func(left int, right int) bool {
		return accountIDs[left] < accountIDs[right]
	})

	result := make(map[int64]game.ShipInput, len(accountIDs))
	for _, accountID := range accountIDs {
		result[world.accountObjectIDs[accountID]] = world.inputs[accountID]
	}
	return result
}

// Проверяет, что объект не имеет ни линейного, ни углового движения.
func cosmicObjectIsFullyStopped(cosmicObject data.CosmicObject) bool {
	return math.Abs(cosmicObject.VelocityX) <= physics.Epsilon &&
		math.Abs(cosmicObject.VelocityY) <= physics.Epsilon &&
		math.Abs(cosmicObject.Speed) <= physics.Epsilon &&
		math.Abs(cosmicObject.AngularSpeed) <= physics.Epsilon
}

// Двигает все подвижные объекты мира до общего решения столкновений.
func (world *World) stepMovableObjects(dtSeconds float64, inputsByObjectID map[int64]game.ShipInput) {
	objectIDs := make([]int64, 0, len(world.data.CosmicObjects.Items))
	for objectID := range world.data.CosmicObjects.Items {
		objectIDs = append(objectIDs, objectID)
	}
	sort.Slice(objectIDs, func(left int, right int) bool {
		return objectIDs[left] < objectIDs[right]
	})

	for _, objectID := range objectIDs {
		cosmicObject, ok := world.data.CosmicObjects.Get(objectID)
		if !ok {
			continue
		}
		if !cosmicObject.Enabled {
			world.updateEquipmentUsage(cosmicObject, dtSeconds)
			continue
		}
		model, ok := world.data.CosmicObjectModels.Get(cosmicObject.CosmicObjectModelID)
		if !ok {
			world.updateEquipmentUsage(cosmicObject, dtSeconds)
			continue
		}
		if cosmicObject.Anchored {
			world.updateEquipmentUsage(cosmicObject, dtSeconds)
			continue
		}

		input, controlled := inputsByObjectID[objectID]
		isShip := world.isShipModel(model)
		if controlled && (!isShip || shipHasFuel(*cosmicObject)) {
			*cosmicObject = physics.StepShip(*cosmicObject, *model, input, dtSeconds)
		} else if isShip {
			*cosmicObject = physics.StepUnpilotedShip(*cosmicObject, dtSeconds)
		} else {
			*cosmicObject = physics.StepFreeBody(*cosmicObject, dtSeconds)
		}
		world.updateEquipmentUsage(cosmicObject, dtSeconds)
	}
}

// Обновляет активность оборудования, мощность и запас топлива после шага объекта.
func (world *World) updateEquipmentUsage(cosmicObject *data.CosmicObject, dtSeconds float64) {
	cosmicObject.ConsumingPower = 0
	cosmicObject.GeneratingPower = 0
	if world.data.EquipmentGroups == nil || world.data.ItemModels == nil {
		return
	}

	fuelConsumptionPerSecond := 0.0
	generatorFuelConsumptionPerSecond := 0.0
	generatorPower := 0.0
	generatorGroups := make([]*data.EquipmentGroup, 0)
	for _, group := range sortedEquipmentGroups(world.data.EquipmentGroups.GetByCosmicObjectID(cosmicObject.ID)) {
		model, ok := world.data.ItemModels.Get(group.EquipmentItemModelID)
		if !ok {
			group.Active = false
			continue
		}

		enabledCount := enabledEquipmentCount(group)
		if enabledCount <= 0 {
			group.Active = false
			continue
		}
		if !shipHasFuel(*cosmicObject) && equipmentConsumesStoredFuel(*model) {
			group.Active = false
			continue
		}

		if model.GeneratingPower > 0 {
			cosmicObject.GeneratingPower += model.GeneratingPower * float64(enabledCount)
			generatorPower += model.GeneratingPower * float64(enabledCount)
			if model.ConsumingItemModelID > 0 && model.ConsumingCount > 0 {
				generatorFuelConsumptionPerSecond += model.ConsumingCount * float64(enabledCount)
			}
			group.Active = false
			generatorGroups = append(generatorGroups, group)
			continue
		}

		group.Active = equipmentIsActive(*cosmicObject, *model)
		if !group.Active {
			continue
		}

		cosmicObject.ConsumingPower += model.ConsumingPower * float64(enabledCount)
		if model.ConsumingItemModelID > 0 && model.ConsumingCount > 0 {
			fuelConsumptionPerSecond += model.ConsumingCount * float64(enabledCount)
		}
	}

	if generatorPower > 0 && cosmicObject.ConsumingPower > 0 && generatorFuelConsumptionPerSecond > 0 {
		generatorLoadRatio := cosmicObject.ConsumingPower / generatorPower
		fuelConsumptionPerSecond += generatorFuelConsumptionPerSecond * generatorLoadRatio
		for _, group := range generatorGroups {
			group.Active = true
		}
	}

	if dtSeconds > 0 && fuelConsumptionPerSecond > 0 {
		cosmicObject.Fuel = math.Max(0, cosmicObject.Fuel-fuelConsumptionPerSecond*dtSeconds)
	}
}

// Возвращает фактически включенное количество единиц оборудования.
func enabledEquipmentCount(group *data.EquipmentGroup) int64 {
	if group == nil || !group.Enabled {
		return 0
	}
	return group.EnabledCount
}

// Проверяет, что в баке есть ресурс для топливозависимого оборудования.
func shipHasFuel(cosmicObject data.CosmicObject) bool {
	return cosmicObject.Fuel > physics.Epsilon
}

// Проверяет, что модель тратит хранимый ресурс корабля.
func equipmentConsumesStoredFuel(model data.ItemModel) bool {
	return model.ConsumingItemModelID > 0 && model.ConsumingCount > 0
}

// Определяет, выполняет ли оборудование работу в текущем тике.
func equipmentIsActive(cosmicObject data.CosmicObject, model data.ItemModel) bool {
	usesLinearForce := model.MaxAlongForce > 0 || model.MaxAcrossForce > 0
	usesTorque := model.MaxTorque > 0
	if usesLinearForce || usesTorque {
		return (usesLinearForce && (math.Abs(cosmicObject.AlongForce) > physics.Epsilon || math.Abs(cosmicObject.AcrossForce) > physics.Epsilon)) ||
			(usesTorque && math.Abs(cosmicObject.Torque) > physics.Epsilon)
	}

	return false
}

// Возвращает группы оборудования в стабильном порядке ID.
func sortedEquipmentGroups(groups []*data.EquipmentGroup) []*data.EquipmentGroup {
	result := append([]*data.EquipmentGroup(nil), groups...)
	sort.Slice(result, func(left int, right int) bool {
		return result[left].ID < result[right].ID
	})
	return result
}

// Проверяет, что модель относится к кораблям.
func (world *World) isShipModel(model *data.CosmicObjectModel) bool {
	cosmicObjectType, ok := world.data.CosmicObjectTypes.Get(model.CosmicObjectTypeID)
	return ok && cosmicObjectType.Acronym == "Ship"
}

// Решает столкновения всех пар тел после движения объектов.
func (world *World) resolveAllCollisions() {
	objectIDs := make([]int64, 0, len(world.data.CosmicObjects.Items))
	for objectID := range world.data.CosmicObjects.Items {
		objectIDs = append(objectIDs, objectID)
	}
	sort.Slice(objectIDs, func(left int, right int) bool {
		return objectIDs[left] < objectIDs[right]
	})

	for leftIndex := 0; leftIndex < len(objectIDs); leftIndex++ {
		first, ok := world.data.CosmicObjects.Get(objectIDs[leftIndex])
		if !ok {
			continue
		}
		firstModel, ok := world.data.CosmicObjectModels.Get(first.CosmicObjectModelID)
		if !ok {
			continue
		}

		for rightIndex := leftIndex + 1; rightIndex < len(objectIDs); rightIndex++ {
			second, ok := world.data.CosmicObjects.Get(objectIDs[rightIndex])
			if !ok {
				continue
			}
			secondModel, ok := world.data.CosmicObjectModels.Get(second.CosmicObjectModelID)
			if !ok {
				continue
			}
			collision, collided := physics.CollisionInfo(*first, *firstModel, *second, *secondModel)
			if !collided {
				continue
			}
			nextFirst, nextSecond := physics.ApplyCollisionResponse(*first, *firstModel, *second, *secondModel, collision)
			*first = nextFirst
			*second = nextSecond
		}
	}
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
	if world.data.ItemGroups != nil {
		if err := world.data.ItemGroups.SaveToFile(filepath.Join(dataDirectory, "ItemGroups.json")); err != nil {
			return err
		}
	}
	if world.data.Itemtypes != nil {
		if err := world.data.Itemtypes.SaveToFile(filepath.Join(dataDirectory, "Itemtypes.json")); err != nil {
			return err
		}
	}
	if world.data.Chats != nil {
		if err := world.data.Chats.SaveToFile(filepath.Join(dataDirectory, "Chats.json")); err != nil {
			return err
		}
	}
	if world.data.ChatMembers != nil {
		if err := world.data.ChatMembers.SaveToFile(filepath.Join(dataDirectory, "ChatMembers.json")); err != nil {
			return err
		}
	}
	if world.data.CommunityTypes != nil {
		if err := world.data.CommunityTypes.SaveToFile(filepath.Join(dataDirectory, "CommunityTypes.json")); err != nil {
			return err
		}
	}
	if world.data.CommunityChatRoles != nil {
		if err := world.data.CommunityChatRoles.SaveToFile(filepath.Join(dataDirectory, "CommunityChatRoles.json")); err != nil {
			return err
		}
	}
	if world.data.Messages != nil {
		if err := world.data.Messages.SaveToFile(filepath.Join(dataDirectory, "Messages.json")); err != nil {
			return err
		}
	}
	if world.data.MessageReads != nil {
		if err := world.data.MessageReads.SaveToFile(filepath.Join(dataDirectory, "MessageReads.json")); err != nil {
			return err
		}
	}
	if world.data.MessageTypes != nil {
		if err := world.data.MessageTypes.SaveToFile(filepath.Join(dataDirectory, "MessageTypes.json")); err != nil {
			return err
		}
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

	equipmentGroups := make([]data.EquipmentGroup, 0)
	if world.data.EquipmentGroups != nil {
		groupIDs := make([]int64, 0, len(world.data.EquipmentGroups.Items))
		for groupID := range world.data.EquipmentGroups.Items {
			groupIDs = append(groupIDs, groupID)
		}
		sort.Slice(groupIDs, func(left int, right int) bool {
			return groupIDs[left] < groupIDs[right]
		})
		for _, groupID := range groupIDs {
			group, ok := world.data.EquipmentGroups.Get(groupID)
			if !ok {
				continue
			}
			equipmentGroups = append(equipmentGroups, *group)
		}
	}

	return game.Snapshot{
		Type:            "snapshot",
		Tick:            world.tick,
		SelfObjectID:    selfObjectID,
		Objects:         objects,
		EquipmentGroups: equipmentGroups,
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
	if world.data.ItemGroups != nil {
		equipmentGroupIDs := make([]int64, 0)
		for _, group := range world.data.EquipmentGroups.GetByCosmicObjectID(cosmicObjectID) {
			equipmentGroupIDs = append(equipmentGroupIDs, group.ID)
		}
		world.data.ItemGroups.DeleteByContainerEquipmentGroupIDs(equipmentGroupIDs)
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

// Заполняет новый корабль топливом и кладет боеприпасы в установленные контейнеры.
func (world *World) fillShipSupplies(cosmicObject *data.CosmicObject) {
	if cosmicObject == nil {
		return
	}
	cosmicObject.Fuel = cosmicObject.MaxFuel
	if world.data.EquipmentGroups == nil || world.data.ItemGroups == nil || world.data.ItemModels == nil || world.data.Itemtypes == nil {
		return
	}

	containerType, ok := world.data.Itemtypes.GetByAcronym("Container")
	if !ok {
		return
	}

	containerIDs := make([]int64, 0)
	ammoByModelID := make(map[int64]float64)
	for _, group := range world.data.EquipmentGroups.GetByCosmicObjectID(cosmicObject.ID) {
		model, ok := world.data.ItemModels.Get(group.EquipmentItemModelID)
		if !ok {
			continue
		}
		if model.ItemtypeID == containerType.ID {
			containerIDs = append(containerIDs, group.ID)
		}
		if model.AmmoItemModelID > 0 && model.FiringRate > 0 && group.Count > 0 {
			ammoByModelID[model.AmmoItemModelID] += float64(group.Count) * model.FiringRate * 15 * 60
		}
	}
	if len(containerIDs) == 0 || len(ammoByModelID) == 0 {
		return
	}
	sort.Slice(containerIDs, func(left int, right int) bool {
		return containerIDs[left] < containerIDs[right]
	})
	world.data.ItemGroups.DeleteByContainerEquipmentGroupIDs(containerIDs)

	ammoModelIDs := make([]int64, 0, len(ammoByModelID))
	for ammoModelID := range ammoByModelID {
		ammoModelIDs = append(ammoModelIDs, ammoModelID)
	}
	sort.Slice(ammoModelIDs, func(left int, right int) bool {
		return ammoModelIDs[left] < ammoModelIDs[right]
	})
	for _, ammoModelID := range ammoModelIDs {
		_, _ = world.data.ItemGroups.Add(&data.ItemGroup{
			ContainerEquipmentGroupID: containerIDs[0],
			ContentItemModelID:        ammoModelID,
			Count:                     math.Ceil(ammoByModelID[ammoModelID]),
		})
	}
}
