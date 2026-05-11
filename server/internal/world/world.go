package world

import (
	"errors"
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
	Accounts                   *data.Accounts                   // Учетные записи, доступные игровой симуляции.
	Characters                 *data.Characters                 // Персонажи, доступные игровой симуляции.
	CosmicObjects              *data.CosmicObjects              // Экземпляры объектов, участвующие в мире.
	CosmicObjectTypes          *data.CosmicObjectTypes          // Справочник типов объектов для правил мира.
	CosmicObjectModels         *data.CosmicObjectModels         // Справочник моделей объектов для физики и отображения.
	Itemtypes                  *data.Itemtypes                  // Справочник типов предметов для серверной логики.
	ItemModels                 *data.ItemModels                 // Справочник моделей предметов для оборудования и содержимого контейнеров.
	EquipmentGroups            *data.EquipmentGroups            // Группы оборудования, установленные на объектах мира.
	ItemGroups                 *data.ItemGroups                 // Группы предметов внутри контейнерного оборудования.
	Assemblies                 *data.Assemblies                 // Справочник сборок для расчета характеристик кораблей.
	AssemblyEquipmentGroups    *data.AssemblyEquipmentGroups    // Группы оборудования, заданные в сборках.
	Chats                      *data.Chats                      // Чаты игрового мира.
	ChatMembers                *data.ChatMembers                // Участники чатов.
	CommunityTypes             *data.CommunityTypes             // Справочник типов сообществ.
	CommunityChatRoles         *data.CommunityChatRoles         // Справочник ролей в чатах сообществ.
	Messages                   *data.Messages                   // Сообщения чатов.
	MessageReads               *data.MessageReads               // Позиции чтения сообщений персонажами.
	MessageTypes               *data.MessageTypes               // Справочник типов сообщений.
	ActionTypes                *data.ActionTypes                // Справочник игровых действий для настроек ввода.
	InputEventTypes            *data.InputEventTypes            // Справочник событий ввода для настроек.
	DefaultActionInputSettings *data.DefaultActionInputSettings // Привязки ввода по умолчанию.
	AccountActionInputSettings *data.AccountActionInputSettings // Привязки ввода, выбранные аккаунтами.
}

// Управляет подключенными аккаунтами, вводом игроков и пошаговой симуляцией объектов.
type World struct {
	mu               sync.Mutex               // Защищает изменяемое состояние мира от параллельных горутин.
	tick             int64                    // Номер последнего выполненного шага симуляции.
	data             Data                     // Справочники и игровые сущности, которыми управляет мир.
	accountObjectIDs map[int64]int64          // Связь подключенных аккаунтов с управляемыми объектами.
	inputs           map[int64]game.ShipInput // Последний принятый ввод для каждого подключенного аккаунта.
	mutationAcks     map[string]int64         // Последний обработанный номер команды панели по аккаунту и сессии.
	random           *rand.Rand               // Источник случайности для воспроизводимых команд.
}

// ControlPanelObjectUpdate описывает частичное изменение управляемого объекта.
type ControlPanelObjectUpdate struct {
	Enabled *bool   // Новое состояние включения объекта, если оно меняется.
	Title   *string // Новое пользовательское название объекта, если оно меняется.
}

// ControlPanelEquipmentUpdate описывает частичное изменение группы оборудования.
type ControlPanelEquipmentUpdate struct {
	EquipmentGroupID int64  // Группа оборудования, которую нужно изменить.
	Enabled          *bool  // Новое состояние включения группы, если оно меняется.
	EnabledCount     *int64 // Новое количество включенных единиц, если оно меняется.
}

// ControlPanelContainerTransfer описывает перенос предметов между контейнерами текущего объекта.
type ControlPanelContainerTransfer struct {
	SourceContainerEquipmentGroupID int64   // Группа контейнеров, из которой переносятся все предметы.
	TargetContainerEquipmentGroupID int64   // Группа контейнеров, в которую переносятся все предметы.
	ItemGroupIDs                    []int64 // Группы предметов, выбранные для переноса.
}

// Создает мир поверх уже загруженных серверных данных.
func New(seed int64, serverData Data) *World {
	created := &World{
		data:             serverData,
		accountObjectIDs: map[int64]int64{},
		inputs:           map[int64]game.ShipInput{},
		mutationAcks:     map[string]int64{},
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

// ApplyControlPanelObjectUpdate применяет подтвержденное изменение панели к объекту текущего аккаунта.
func (world *World) ApplyControlPanelObjectUpdate(accountID int64, sessionID string, mutationSeq int64, update ControlPanelObjectUpdate) error {
	world.mu.Lock()
	defer world.mu.Unlock()

	objectID, ok := world.accountObjectIDs[accountID]
	if !ok {
		return errors.New("account is not connected")
	}
	cosmicObject, ok := world.data.CosmicObjects.Get(objectID)
	if !ok {
		return errors.New("controlled object not found")
	}

	if update.Enabled != nil {
		cosmicObject.Enabled = *update.Enabled
	}
	if update.Title != nil {
		cosmicObject.Title = *update.Title
	}
	world.ackMutationLocked(accountID, sessionID, mutationSeq)
	return nil
}

// ApplyControlPanelEquipmentUpdate применяет подтвержденное изменение панели к оборудованию текущего объекта.
func (world *World) ApplyControlPanelEquipmentUpdate(accountID int64, sessionID string, mutationSeq int64, update ControlPanelEquipmentUpdate) error {
	world.mu.Lock()
	defer world.mu.Unlock()

	objectID, ok := world.accountObjectIDs[accountID]
	if !ok {
		return errors.New("account is not connected")
	}
	if world.data.EquipmentGroups == nil {
		return errors.New("equipment groups are not loaded")
	}
	group, ok := world.data.EquipmentGroups.Get(update.EquipmentGroupID)
	if !ok {
		return errors.New("equipment group not found")
	}
	if group.CosmicObjectID != objectID {
		return errors.New("equipment group does not belong to controlled object")
	}

	if update.EnabledCount != nil {
		if *update.EnabledCount < 1 || *update.EnabledCount > group.Count {
			return errors.New("enabled equipment count is out of range")
		}
		group.EnabledCount = *update.EnabledCount
	}
	if update.Enabled != nil {
		group.Enabled = *update.Enabled
	}
	world.ackMutationLocked(accountID, sessionID, mutationSeq)
	return nil
}

// ApplyControlPanelContainerTransfer переносит всё содержимое из одного контейнера текущего объекта в другой.
func (world *World) ApplyControlPanelContainerTransfer(accountID int64, sessionID string, mutationSeq int64, transfer ControlPanelContainerTransfer) error {
	world.mu.Lock()
	defer world.mu.Unlock()

	objectID, ok := world.accountObjectIDs[accountID]
	if !ok {
		return errors.New("account is not connected")
	}
	if world.data.EquipmentGroups == nil {
		return errors.New("equipment groups are not loaded")
	}
	if world.data.ItemGroups == nil {
		return errors.New("item groups are not loaded")
	}
	source, err := world.controlledContainerEquipmentLocked(objectID, transfer.SourceContainerEquipmentGroupID)
	if err != nil {
		return err
	}
	target, err := world.controlledContainerEquipmentLocked(objectID, transfer.TargetContainerEquipmentGroupID)
	if err != nil {
		return err
	}
	if source.ID == target.ID {
		return errors.New("source and target containers must be different")
	}

	targetByModel := make(map[int64]*data.ItemGroup)
	for _, itemGroup := range world.data.ItemGroups.GetByContainerEquipmentGroupID(target.ID) {
		targetByModel[itemGroup.ContentItemModelID] = itemGroup
	}
	for _, itemGroupID := range transfer.ItemGroupIDs {
		itemGroup, ok := world.data.ItemGroups.Get(itemGroupID)
		if !ok {
			return errors.New("item group not found")
		}
		if itemGroup.ContainerEquipmentGroupID != source.ID {
			return errors.New("item group does not belong to source container")
		}
		if existing := targetByModel[itemGroup.ContentItemModelID]; existing != nil {
			existing.Count += itemGroup.Count
			delete(world.data.ItemGroups.Items, itemGroup.ID)
			continue
		}
		itemGroup.ContainerEquipmentGroupID = target.ID
		targetByModel[itemGroup.ContentItemModelID] = itemGroup
	}
	if err := world.data.ItemGroups.RebuildIndexes(); err != nil {
		return err
	}
	world.ackMutationLocked(accountID, sessionID, mutationSeq)
	return nil
}

// controlledContainerEquipmentLocked возвращает контейнер текущего объекта; вызывается только под mutex.
func (world *World) controlledContainerEquipmentLocked(objectID int64, groupID int64) (*data.EquipmentGroup, error) {
	group, ok := world.data.EquipmentGroups.Get(groupID)
	if !ok {
		return nil, errors.New("equipment group not found")
	}
	if group.CosmicObjectID != objectID {
		return nil, errors.New("equipment group does not belong to controlled object")
	}
	if !world.equipmentGroupIsContainerLocked(group) {
		return nil, errors.New("equipment group is not a container")
	}
	return group, nil
}

// equipmentGroupIsContainerLocked проверяет тип установленного предмета; вызывается только под mutex.
func (world *World) equipmentGroupIsContainerLocked(group *data.EquipmentGroup) bool {
	if group == nil || world.data.ItemModels == nil || world.data.Itemtypes == nil {
		return false
	}
	model, ok := world.data.ItemModels.Get(group.EquipmentItemModelID)
	if !ok {
		return false
	}
	itemtype, ok := world.data.Itemtypes.Get(model.ItemtypeID)
	return ok && itemtype.Acronym == "Container"
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
			*cosmicObject = world.stepControlledObject(*cosmicObject, *model, input, dtSeconds)
		} else if isShip {
			*cosmicObject = physics.StepUnpilotedShip(*cosmicObject, dtSeconds)
		} else {
			*cosmicObject = physics.StepFreeBody(*cosmicObject, dtSeconds)
		}
		world.updateEquipmentUsage(cosmicObject, dtSeconds)
	}
}

// Выполняет физический шаг с силами, доступными только от включенного оборудования.
func (world *World) stepControlledObject(cosmicObject data.CosmicObject, model data.CosmicObjectModel, input game.ShipInput, dtSeconds float64) data.CosmicObject {
	effectiveObject := world.objectWithEnabledEquipmentForces(cosmicObject, input)
	next := physics.StepShip(effectiveObject, model, input, dtSeconds)
	next.MaxAlongForce = cosmicObject.MaxAlongForce
	next.MaxAcrossForce = cosmicObject.MaxAcrossForce
	next.MaxTorque = cosmicObject.MaxTorque
	return next
}

// objectWithEnabledEquipmentForces рассчитывает доступную тягу и момент по включенным группам оборудования.
func (world *World) objectWithEnabledEquipmentForces(cosmicObject data.CosmicObject, input game.ShipInput) data.CosmicObject {
	if world.data.EquipmentGroups == nil || world.data.ItemModels == nil {
		return cosmicObject
	}

	electricShare := world.electricShareForInput(cosmicObject, input)
	cosmicObject.MaxAlongForce = 0
	cosmicObject.MaxAcrossForce = 0
	cosmicObject.MaxTorque = 0
	for _, group := range sortedEquipmentGroups(world.data.EquipmentGroups.GetByCosmicObjectID(cosmicObject.ID)) {
		enabledCount := enabledEquipmentCount(group)
		if enabledCount <= 0 {
			continue
		}
		model, ok := world.data.ItemModels.Get(group.EquipmentItemModelID)
		if !ok {
			continue
		}
		cosmicObject.MaxAlongForce += model.MaxAlongForce * float64(enabledCount) * electricShare
		cosmicObject.MaxAcrossForce += model.MaxAcrossForce * float64(enabledCount) * electricShare
		cosmicObject.MaxTorque += model.MaxTorque * float64(enabledCount) * electricShare
	}
	return cosmicObject
}

// electricShareForInput возвращает долю электричества, доступную всем потребителям при текущем вводе.
func (world *World) electricShareForInput(cosmicObject data.CosmicObject, input game.ShipInput) float64 {
	if world.data.EquipmentGroups == nil || world.data.ItemModels == nil {
		return 1
	}

	generatedPower := 0.0
	neededPower := 0.0
	for _, group := range sortedEquipmentGroups(world.data.EquipmentGroups.GetByCosmicObjectID(cosmicObject.ID)) {
		enabledCount := enabledEquipmentCount(group)
		if enabledCount <= 0 {
			continue
		}
		model, ok := world.data.ItemModels.Get(group.EquipmentItemModelID)
		if !ok {
			continue
		}
		count := float64(enabledCount)
		generatedPower += model.GeneratingPower * count
		if equipmentNeedsElectricityForInput(input, *model) {
			neededPower += model.ConsumingPower * count
		}
	}

	return electricWorkShare(generatedPower, neededPower)
}

// Обновляет активность оборудования, мощность и запас топлива после шага объекта.
func (world *World) updateEquipmentUsage(cosmicObject *data.CosmicObject, dtSeconds float64) {
	cosmicObject.ConsumingPower = 0
	cosmicObject.GeneratingPower = 0
	if world.data.EquipmentGroups == nil || world.data.ItemModels == nil {
		return
	}

	consumerFuelConsumptionPerSecond := 0.0
	generatorFuelConsumptionPerSecond := 0.0
	neededPower := 0.0
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

		count := float64(enabledCount)
		cosmicObject.GeneratingPower += model.GeneratingPower * count
		if model.GeneratingPower != 0 {
			group.Active = false
			generatorGroups = append(generatorGroups, group)
			if model.ConsumingItemModelID > 0 && model.ConsumingCount > 0 {
				generatorFuelConsumptionPerSecond += model.ConsumingCount * count
			}
		}

		group.Active = equipmentIsActive(*cosmicObject, *model)
		if !group.Active {
			continue
		}

		neededPower += model.ConsumingPower * count
		if model.ConsumingItemModelID > 0 && model.ConsumingCount > 0 {
			consumerFuelConsumptionPerSecond += model.ConsumingCount * count
		}
	}

	electricShare := electricWorkShare(cosmicObject.GeneratingPower, neededPower)
	cosmicObject.ConsumingPower = neededPower * electricShare

	fuelConsumptionPerSecond := consumerFuelConsumptionPerSecond * electricShare
	if math.Abs(cosmicObject.ConsumingPower) > physics.Epsilon && math.Abs(cosmicObject.GeneratingPower) > physics.Epsilon {
		generatorLoad := math.Abs(cosmicObject.ConsumingPower / cosmicObject.GeneratingPower)
		fuelConsumptionPerSecond += generatorFuelConsumptionPerSecond * generatorLoad
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

// Определяет долю работы, которую можно обеспечить имеющейся электроэнергией.
func electricWorkShare(generatedPower float64, neededPower float64) float64 {
	if neededPower <= 0 || generatedPower >= neededPower {
		return 1
	}
	if generatedPower <= 0 {
		return 0
	}
	return generatedPower / neededPower
}

// Проверяет, будет ли модель тратить электричество при текущем управлении.
func equipmentNeedsElectricityForInput(input game.ShipInput, model data.ItemModel) bool {
	usesAlongForce := model.MaxAlongForce != 0 && (input.ThrustForward || input.ThrustBackward)
	usesAcrossForce := model.MaxAcrossForce != 0 && (input.ThrustLeft || input.ThrustRight)
	usesTorque := model.MaxTorque != 0 && input.TargetRotationDelta != 0
	return usesAlongForce || usesAcrossForce || usesTorque
}

// Определяет, выполняет ли оборудование работу в текущем тике.
func equipmentIsActive(cosmicObject data.CosmicObject, model data.ItemModel) bool {
	usesLinearForce := model.MaxAlongForce != 0 || model.MaxAcrossForce != 0
	usesTorque := model.MaxTorque != 0
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

// ackMutationLocked запоминает последний обработанный номер команды панели под уже взятым mutex.
func (world *World) ackMutationLocked(accountID int64, sessionID string, mutationSeq int64) {
	if sessionID == "" || mutationSeq <= 0 {
		return
	}
	key := mutationAckKey(accountID, sessionID)
	if mutationSeq > world.mutationAcks[key] {
		world.mutationAcks[key] = mutationSeq
	}
}

// mutationAckKey собирает ключ подтверждения команд панели для аккаунта и сессии.
func mutationAckKey(accountID int64, sessionID string) string {
	return fmt.Sprintf("%d:%s", accountID, sessionID)
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

// ClientMutationAck возвращает последний обработанный номер команды панели для клиентской сессии.
func (world *World) ClientMutationAck(accountID int64, sessionID string) game.ClientMutationAck {
	world.mu.Lock()
	defer world.mu.Unlock()

	return game.ClientMutationAck{
		SessionID:      sessionID,
		LastAppliedSeq: world.mutationAcks[mutationAckKey(accountID, sessionID)],
	}
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
	if world.data.ActionTypes != nil {
		if err := world.data.ActionTypes.SaveToFile(filepath.Join(dataDirectory, "ActionTypes.json")); err != nil {
			return err
		}
	}
	if world.data.InputEventTypes != nil {
		if err := world.data.InputEventTypes.SaveToFile(filepath.Join(dataDirectory, "InputEventTypes.json")); err != nil {
			return err
		}
	}
	if world.data.DefaultActionInputSettings != nil {
		if err := world.data.DefaultActionInputSettings.SaveToFile(filepath.Join(dataDirectory, "DefaultActionInputSettings.json")); err != nil {
			return err
		}
	}
	if world.data.AccountActionInputSettings != nil {
		if err := world.data.AccountActionInputSettings.SaveToFile(filepath.Join(dataDirectory, "AccountActionInputSettings.json")); err != nil {
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

	itemGroups := make([]data.ItemGroup, 0)
	if world.data.ItemGroups != nil {
		itemGroupIDs := make([]int64, 0, len(world.data.ItemGroups.Items))
		for itemGroupID := range world.data.ItemGroups.Items {
			itemGroupIDs = append(itemGroupIDs, itemGroupID)
		}
		sort.Slice(itemGroupIDs, func(left int, right int) bool {
			return itemGroupIDs[left] < itemGroupIDs[right]
		})
		for _, itemGroupID := range itemGroupIDs {
			itemGroup, ok := world.data.ItemGroups.Get(itemGroupID)
			if !ok {
				continue
			}
			itemGroups = append(itemGroups, *itemGroup)
		}
	}

	return game.Snapshot{
		Type:            "snapshot",
		Tick:            world.tick,
		SelfObjectID:    selfObjectID,
		Objects:         objects,
		EquipmentGroups: equipmentGroups,
		ItemGroups:      itemGroups,
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

// Заполняет новый корабль топливом и кладет образцы предметов в первый контейнер.
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
	for _, group := range world.data.EquipmentGroups.GetByCosmicObjectID(cosmicObject.ID) {
		model, ok := world.data.ItemModels.Get(group.EquipmentItemModelID)
		if !ok {
			continue
		}
		if model.ItemtypeID == containerType.ID {
			containerIDs = append(containerIDs, group.ID)
		}
	}
	if len(containerIDs) == 0 {
		return
	}
	sort.Slice(containerIDs, func(left int, right int) bool {
		return containerIDs[left] < containerIDs[right]
	})
	world.data.ItemGroups.DeleteByContainerEquipmentGroupIDs(containerIDs)

	sampleModelIDs := world.firstItemModelIDsByType()
	for _, itemModelID := range sampleModelIDs {
		_, _ = world.data.ItemGroups.Add(&data.ItemGroup{
			ContainerEquipmentGroupID: containerIDs[0],
			ContentItemModelID:        itemModelID,
			Count:                     10,
		})
	}
}

// Возвращает первую модель предмета каждого типа в стабильном порядке типов.
func (world *World) firstItemModelIDsByType() []int64 {
	firstByType := make(map[int64]int64)
	for itemModelID, model := range world.data.ItemModels.Items {
		if model == nil || model.ItemtypeID <= 0 {
			continue
		}
		current, ok := firstByType[model.ItemtypeID]
		if !ok || itemModelID < current {
			firstByType[model.ItemtypeID] = itemModelID
		}
	}

	typeIDs := make([]int64, 0, len(firstByType))
	for itemtypeID := range firstByType {
		typeIDs = append(typeIDs, itemtypeID)
	}
	sort.Slice(typeIDs, func(left int, right int) bool {
		return typeIDs[left] < typeIDs[right]
	})

	result := make([]int64, 0, len(typeIDs))
	for _, itemtypeID := range typeIDs {
		result = append(result, firstByType[itemtypeID])
	}
	return result
}
