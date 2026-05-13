package world

import (
	"encoding/json"
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
	"space-game-07-server/internal/storage"
)

const (
	defaultAccountEmailDomain = "auto.local"
	defaultAccountPassword    = "auto"
	defaultStarterShipAcronym = "ship_bat"
)

// Собирает справочники и игровые сущности, нужные симуляции мира.
type Data struct {
	RelationTypes              *data.RelationTypes              // Справочник видов связей групп оборудования.
	EquipmentGroupRelations    *data.EquipmentGroupRelations    // Сохранённые связи выбранных групп оборудования.
	Accounts                   *data.Accounts                   // Учетные записи, доступные игровой симуляции.
	Characters                 *data.Characters                 // Персонажи, доступные игровой симуляции.
	CosmicObjects              *data.CosmicObjects              // Экземпляры объектов, участвующие в мире.
	CosmicObjectTypes          *data.CosmicObjectTypes          // Справочник типов объектов для правил мира.
	CosmicObjectModels         *data.CosmicObjectModels         // Справочник моделей объектов для физики и отображения.
	Itemtypes                  *data.Itemtypes                  // Справочник типов предметов для серверной логики.
	ItemModels                 *data.ItemModels                 // Справочник моделей предметов для оборудования и содержимого контейнеров.
	Blueprints                 *storage.RawReferenceTable       // Справочник чертежей объектов для изготовления в конструкторе.
	BlueprintComponents        *storage.RawReferenceTable       // Справочник компонентов чертежей объектов для списания материалов.
	Schemas                    *storage.RawReferenceTable       // Справочник схем предметов для изготовления в конструкторе.
	SchemaComponents           *storage.RawReferenceTable       // Справочник компонентов схем предметов для списания материалов.
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
	mu                             sync.Mutex                 // Защищает изменяемое состояние мира от параллельных горутин.
	tick                           int64                      // Номер последнего выполненного шага симуляции.
	data                           Data                       // Справочники и игровые сущности, которыми управляет мир.
	accountObjectIDs               map[int64]int64            // Связь подключенных аккаунтов с управляемыми объектами.
	inputs                         map[int64]game.ShipInput   // Последний принятый ввод для каждого подключенного аккаунта.
	mutationAcks                   map[string]int64           // Последний обработанный номер команды панели по аккаунту и сессии.
	random                         *rand.Rand                 // Источник случайности для воспроизводимых команд.
	nextConstructorProductionJobID int64                      // Следующий идентификатор задания изготовления.
	constructorProductionJobs      []constructorProductionJob // Задания изготовления, ожидающие или выполняющиеся в конструкторах.
}

type constructorProductionJob struct {
	ID                                int64   // Уникальный числовой идентификатор задания.
	ConstructorEquipmentGroupID       int64   // Конструктор, к очереди которого относится задание.
	MaterialContainerEquipmentGroupID int64   // Контейнер, из которого списываются компоненты.
	ProductContainerEquipmentGroupID  int64   // Контейнер, в который кладется результат.
	QueueType                         string  // Очередь задания: основная или вспомогательная.
	SchemaID                          int64   // Схема, по которой изготавливается предмет.
	BlueprintID                       int64   // Чертёж, по которому изготавливается космический объект.
	ProductItemModelID                int64   // Модель предмета, который получится после завершения.
	ProductCosmicObjectModelID        int64   // Модель космического объекта, который появится после завершения.
	ProductCount                      float64 // Количество предметов, которое получится после завершения.
	RemainingBatches                  int64   // ���������� ��������, ������� ��� �������� ��������� �� ������.
	TotalBatches                      int64   // ����� ���������� ��������, ��������������� �� ������.
	RemainingTime                     float64 // Оставшееся время изготовления в секундах.
	TotalTime                         float64 // Полное время изготовления в секундах.
	Running                           bool    // Показывает, что задание сейчас выполняется.
	ParentJobID                       int64   // ������������ ������, ��� ���������� ������� ����� ��� ��������������� ������.
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
// ControlPanelEquipmentGroupRelationUpdate описывает сохранение выбранной связанной группы оборудования.
type ControlPanelEquipmentGroupRelationUpdate struct {
	EquipmentGroupID        int64  // Группа оборудования, для которой сохраняется выбор.
	RelationTypeAcronym     string // Вид связи, выбранный по неизменяемому строковому идентификатору.
	RelatedEquipmentGroupID int64  // Группа оборудования, выбранная игроком в связанной панели.
}

type ControlPanelContainerTransfer struct {
	SourceContainerEquipmentGroupID int64   // Группа контейнеров, из которой переносятся все предметы.
	TargetContainerEquipmentGroupID int64   // Группа контейнеров, в которую переносятся все предметы.
	ItemGroupIDs                    []int64 // Группы предметов, выбранные для переноса.
	Amount                          float64 // Количество предметов для частичного переноса одной выбранной группы.
}

// Создает мир поверх уже загруженных серверных данных.
// ControlPanelFuelTransfer описывает перенос топлива между контейнером и баком текущего объекта.
type ControlPanelFuelTransfer struct {
	ContainerEquipmentGroupID int64   // Группа контейнеров, из которой берется или куда сливается топливо.
	FuelTankEquipmentGroupID  int64   // Группа топливных баков, с которой работает игрок.
	ItemGroupIDs              []int64 // Группы топлива, выбранные для заливки из контейнера.
	Amount                    float64 // Количество топлива для слива из бака в контейнер.
}

// ControlPanelConstructorProduceItem описывает изготовление предмета по схеме конструктора.
type ControlPanelConstructorProduceItem struct {
	ConstructorEquipmentGroupID       int64 // Группа конструкторов, которая выполняет изготовление.
	MaterialContainerEquipmentGroupID int64 // Контейнер, из которого списываются компоненты схемы.
	ProductContainerEquipmentGroupID  int64 // Контейнер, в который кладется готовая продукция.
	SchemaID                          int64 // Схема предмета, выбранная игроком для изготовления.
	BlueprintID                       int64 // Чертёж объекта, выбранный игроком для изготовления.
	Amount                            int64 // Количество запусков изготовления по выбранной схеме.
}

// controlPanelItemSchema хранит нужные серверу поля одной сырой записи схемы.
type ControlPanelConstructorQueueCommand struct {
	ConstructorEquipmentGroupID int64  // Группа конструкторов, очередь которой меняется.
	JobID                       int64  // Строка основной очереди, выбранная игроком.
	Command                     string // Действие над выбранной строкой и следующими строками.
}

type controlPanelItemSchema struct {
	ID                 int64   `json:"ID"`                 // Уникальный числовой идентификатор схемы.
	ItemModelID        int64   `json:"ItemModelID"`        // Модель предмета, получаемого по схеме.
	Count              float64 `json:"Count"`              // Количество предметов, получаемое за одно изготовление.
	ProductionBaseTime float64 `json:"ProductionBaseTime"` // Базовое время изготовления, пока не используемое мгновенной командой.
}

// controlPanelItemSchemaComponent хранит нужные серверу поля одного компонента схемы.
type controlPanelItemSchemaComponent struct {
	ID                   int64   `json:"ID"`                   // Уникальный числовой идентификатор компонента схемы.
	SchemaID             int64   `json:"SchemaID"`             // Схема, к которой относится компонент.
	ComponentItemModelID int64   `json:"ComponentItemModelID"` // Модель предмета, которую нужно списать как компонент.
	Count                float64 `json:"Count"`                // Количество компонента, требуемое для одного изготовления.
}

type controlPanelObjectBlueprint struct {
	ID                  int64   `json:"ID"`                  // Уникальный числовой идентификатор чертежа.
	CosmicObjectModelID int64   `json:"CosmicObjectModelID"` // Модель космического объекта, получаемого по чертежу.
	ProductionBaseTime  float64 `json:"ProductionBaseTime"`  // Базовое время изготовления одного объекта.
}

type controlPanelObjectBlueprintComponent struct {
	ID                   int64   `json:"ID"`                   // Уникальный числовой идентификатор компонента чертежа.
	BlueprintID          int64   `json:"BlueprintID"`          // Чертёж, к которому относится компонент.
	ComponentItemModelID int64   `json:"ComponentItemModelID"` // Модель предмета, которую нужно списать как компонент.
	Count                float64 `json:"Count"`                // Количество компонента, требуемое для одного изготовления.
}

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
// ApplyControlPanelEquipmentGroupRelationUpdate сохраняет выбранную связанную группу оборудования для текущего объекта.
func (world *World) ApplyControlPanelEquipmentGroupRelationUpdate(accountID int64, sessionID string, mutationSeq int64, update ControlPanelEquipmentGroupRelationUpdate) error {
	world.mu.Lock()
	defer world.mu.Unlock()

	objectID, ok := world.accountObjectIDs[accountID]
	if !ok {
		return errors.New("account is not connected")
	}
	if world.data.EquipmentGroups == nil || world.data.RelationTypes == nil || world.data.EquipmentGroupRelations == nil {
		return errors.New("equipment relation data is not loaded")
	}
	group, ok := world.data.EquipmentGroups.Get(update.EquipmentGroupID)
	if !ok {
		return errors.New("equipment group not found")
	}
	if group.CosmicObjectID != objectID {
		return errors.New("equipment group does not belong to controlled object")
	}
	if _, err := world.controlledContainerEquipmentLocked(objectID, update.RelatedEquipmentGroupID); err != nil {
		return err
	}
	relationType, ok := world.data.RelationTypes.GetByAcronym(update.RelationTypeAcronym)
	if !ok {
		return errors.New("relation type not found")
	}
	if _, err := world.data.EquipmentGroupRelations.Upsert(&data.EquipmentGroupRelation{
		EquipmentGroupID:        update.EquipmentGroupID,
		RelationTypeID:          relationType.ID,
		RelatedEquipmentGroupID: update.RelatedEquipmentGroupID,
	}); err != nil {
		return err
	}
	world.ackMutationLocked(accountID, sessionID, mutationSeq)
	return nil
}

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
		moved := itemGroup.Count
		if transfer.Amount > 0 && len(transfer.ItemGroupIDs) == 1 {
			moved = math.Min(itemGroup.Count, transfer.Amount)
		}
		if existing := targetByModel[itemGroup.ContentItemModelID]; existing != nil {
			existing.Count += moved
		} else if moved < itemGroup.Count {
			created, err := world.data.ItemGroups.Add(&data.ItemGroup{ContainerEquipmentGroupID: target.ID, ContentItemModelID: itemGroup.ContentItemModelID, Count: moved})
			if err != nil {
				return err
			}
			targetByModel[created.ContentItemModelID] = created
		} else {
			itemGroup.ContainerEquipmentGroupID = target.ID
			targetByModel[itemGroup.ContentItemModelID] = itemGroup
			continue
		}
		itemGroup.Count -= moved
		if itemGroup.Count <= physics.Epsilon {
			delete(world.data.ItemGroups.Items, itemGroup.ID)
		}
	}
	if err := world.data.ItemGroups.RebuildIndexes(); err != nil {
		return err
	}
	world.ackMutationLocked(accountID, sessionID, mutationSeq)
	return nil
}

// controlledContainerEquipmentLocked возвращает контейнер текущего объекта; вызывается только под mutex.
// ApplyControlPanelFuelTransfer переносит топливо между контейнером и общим запасом текущего объекта.
func (world *World) ApplyControlPanelFuelTransfer(accountID int64, sessionID string, mutationSeq int64, transfer ControlPanelFuelTransfer) error {
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
	if world.data.EquipmentGroups == nil || world.data.ItemGroups == nil {
		return errors.New("equipment or item groups are not loaded")
	}
	container, err := world.controlledContainerEquipmentLocked(objectID, transfer.ContainerEquipmentGroupID)
	if err != nil {
		return err
	}
	fuelModelID, err := world.controlledFuelTankFuelModelIDLocked(objectID, transfer.FuelTankEquipmentGroupID)
	if err != nil {
		return err
	}
	if len(transfer.ItemGroupIDs) > 0 {
		if err := world.fillFuelFromContainerLocked(cosmicObject, container.ID, fuelModelID, transfer.ItemGroupIDs, transfer.Amount); err != nil {
			return err
		}
	} else if transfer.Amount > 0 {
		if err := world.drainFuelToContainerLocked(cosmicObject, container.ID, fuelModelID, transfer.Amount); err != nil {
			return err
		}
	}
	if err := world.data.ItemGroups.RebuildIndexes(); err != nil {
		return err
	}
	world.ackMutationLocked(accountID, sessionID, mutationSeq)
	return nil
}

// ApplyControlPanelConstructorProduceItem изготавливает одну партию предметов по выбранной схеме.
func (world *World) ApplyControlPanelConstructorProduceItem(accountID int64, sessionID string, mutationSeq int64, production ControlPanelConstructorProduceItem) error {
	world.mu.Lock()
	defer world.mu.Unlock()

	objectID, ok := world.accountObjectIDs[accountID]
	if !ok {
		return errors.New("account is not connected")
	}
	if world.data.EquipmentGroups == nil || world.data.ItemGroups == nil || world.data.ItemModels == nil || world.data.Itemtypes == nil {
		return errors.New("constructor data is not loaded")
	}
	if _, err := world.controlledEquipmentItemtypeLocked(objectID, production.ConstructorEquipmentGroupID, "Constructor"); err != nil {
		return err
	}
	materialContainer, err := world.constructorRelatedContainerOrFallbackLocked(objectID, production.ConstructorEquipmentGroupID, "Source", production.MaterialContainerEquipmentGroupID)
	if err != nil {
		return err
	}
	if (production.SchemaID <= 0) == (production.BlueprintID <= 0) {
		return errors.New("constructor production must select schema or blueprint")
	}
	mainJob, components, amount, err := world.newMainConstructorProductionJobLocked(objectID, production, materialContainer.ID)
	if err != nil {
		return err
	}

	requiredByModel := map[int64]float64{}
	for _, component := range components {
		if component.Count <= 0 {
			return errors.New("constructor component count is invalid")
		}
		if _, ok := world.data.ItemModels.Get(component.ComponentItemModelID); !ok {
			return errors.New("constructor component item model not found")
		}
		requiredByModel[component.ComponentItemModelID] += component.Count
	}
	availableByModel := map[int64]float64{}
	for _, itemGroup := range world.data.ItemGroups.GetByContainerEquipmentGroupID(materialContainer.ID) {
		availableByModel[itemGroup.ContentItemModelID] += itemGroup.Count
	}
	plannedJobs := make([]constructorProductionJob, 0)
	for itemModelID, required := range requiredByModel {
		if err := world.planMissingConstructorComponentsLocked(&plannedJobs, production.ConstructorEquipmentGroupID, materialContainer.ID, availableByModel, itemModelID, required*float64(amount), mainJob.ID, map[int64]bool{}); err != nil {
			return err
		}
	}
	plannedJobs = append(plannedJobs, mainJob)
	world.constructorProductionJobs = append(world.constructorProductionJobs, plannedJobs...)
	world.ackMutationLocked(accountID, sessionID, mutationSeq)
	return nil
}

// ApplyControlPanelConstructorQueueCommand меняет основную очередь конструктора по выбранной строке.
func (world *World) ApplyControlPanelConstructorQueueCommand(accountID int64, sessionID string, mutationSeq int64, command ControlPanelConstructorQueueCommand) error {
	world.mu.Lock()
	defer world.mu.Unlock()

	objectID, ok := world.accountObjectIDs[accountID]
	if !ok {
		return errors.New("account is not connected")
	}
	if _, err := world.controlledEquipmentItemtypeLocked(objectID, command.ConstructorEquipmentGroupID, "Constructor"); err != nil {
		return err
	}
	jobIndex := world.constructorMainJobIndexLocked(command.ConstructorEquipmentGroupID, command.JobID)
	if jobIndex < 0 {
		return errors.New("constructor main job not found")
	}
	switch command.Command {
	case "skipNext":
		world.skipConstructorMainJobNextLocked(jobIndex)
	case "skipAllNext":
		world.removeConstructorMainJobsAfterLocked(command.ConstructorEquipmentGroupID, command.JobID)
		world.skipConstructorMainJobNextLocked(jobIndex)
	case "cancel":
		world.removeConstructorMainJobAtLocked(jobIndex)
	case "cancelAll":
		world.removeConstructorMainJobsFromLocked(command.ConstructorEquipmentGroupID, command.JobID)
	default:
		return errors.New("constructor queue command is invalid")
	}
	world.ackMutationLocked(accountID, sessionID, mutationSeq)
	return nil
}

// planMissingConstructorComponentsLocked добавляет во вспомогательную очередь только недостающее количество компонентов.
// newMainConstructorProductionJobLocked собирает основную строку очереди по схеме или чертежу.
func (world *World) newMainConstructorProductionJobLocked(objectID int64, production ControlPanelConstructorProduceItem, materialContainerID int64) (constructorProductionJob, []controlPanelItemSchemaComponent, int64, error) {
	amount := production.Amount
	if amount <= 0 {
		amount = 1
	}
	if production.SchemaID > 0 {
		productContainer, err := world.constructorRelatedContainerOrFallbackLocked(objectID, production.ConstructorEquipmentGroupID, "Destination", production.ProductContainerEquipmentGroupID)
		if err != nil {
			return constructorProductionJob{}, nil, 0, err
		}
		schema, err := world.itemSchemaLocked(production.SchemaID)
		if err != nil {
			return constructorProductionJob{}, nil, 0, err
		}
		if schema.Count <= 0 {
			return constructorProductionJob{}, nil, 0, errors.New("schema product count is invalid")
		}
		if _, ok := world.data.ItemModels.Get(schema.ItemModelID); !ok {
			return constructorProductionJob{}, nil, 0, errors.New("schema product item model not found")
		}
		components, err := world.itemSchemaComponentsLocked(schema.ID)
		if err != nil {
			return constructorProductionJob{}, nil, 0, err
		}
		if len(components) == 0 {
			return constructorProductionJob{}, nil, 0, errors.New("schema components not found")
		}
		return world.newConstructorProductionJobLocked(production.ConstructorEquipmentGroupID, materialContainerID, productContainer.ID, "main", schema, amount, 0), components, amount, nil
	}
	if world.data.Blueprints == nil || world.data.BlueprintComponents == nil || world.data.CosmicObjectModels == nil {
		return constructorProductionJob{}, nil, 0, errors.New("blueprint data is not loaded")
	}
	blueprint, err := world.objectBlueprintLocked(production.BlueprintID)
	if err != nil {
		return constructorProductionJob{}, nil, 0, err
	}
	if _, ok := world.data.CosmicObjectModels.Get(blueprint.CosmicObjectModelID); !ok {
		return constructorProductionJob{}, nil, 0, errors.New("blueprint object model not found")
	}
	components, err := world.objectBlueprintComponentsLocked(blueprint.ID)
	if err != nil {
		return constructorProductionJob{}, nil, 0, err
	}
	if len(components) == 0 {
		return constructorProductionJob{}, nil, 0, errors.New("blueprint components not found")
	}
	return world.newConstructorObjectProductionJobLocked(production.ConstructorEquipmentGroupID, materialContainerID, "main", blueprint, 1, 0), components, 1, nil
}

func (world *World) planMissingConstructorComponentsLocked(plannedJobs *[]constructorProductionJob, constructorID int64, materialContainerID int64, availableByModel map[int64]float64, itemModelID int64, required float64, parentJobID int64, visiting map[int64]bool) error {
	shortage := required - availableByModel[itemModelID]
	if shortage <= physics.Epsilon {
		availableByModel[itemModelID] -= required
		return nil
	}
	if world.data.Schemas == nil || world.data.SchemaComponents == nil {
		return errors.New("item schema data is not loaded")
	}
	schema, err := world.itemSchemaByProductModelLocked(itemModelID)
	if err != nil {
		return errors.New("not enough schema components")
	}
	if schema.Count <= 0 {
		return errors.New("schema product count is invalid")
	}
	if visiting[schema.ID] {
		return errors.New("schema dependency cycle")
	}
	availableByModel[itemModelID] = 0
	batchCount := int64(math.Ceil(shortage / schema.Count))
	visiting[schema.ID] = true
	job := world.newConstructorProductionJobLocked(constructorID, materialContainerID, materialContainerID, "auxiliary", schema, batchCount, parentJobID)
	components, err := world.itemSchemaComponentsLocked(schema.ID)
	if err != nil {
		return err
	}
	if len(components) == 0 {
		return errors.New("schema components not found")
	}
	for _, component := range components {
		if component.Count <= 0 {
			return errors.New("schema component count is invalid")
		}
		if err := world.planMissingConstructorComponentsLocked(plannedJobs, constructorID, materialContainerID, availableByModel, component.ComponentItemModelID, component.Count*float64(batchCount), job.ID, visiting); err != nil {
			return err
		}
	}
	*plannedJobs = append(*plannedJobs, job)
	availableByModel[itemModelID] += schema.Count * float64(batchCount)
	delete(visiting, schema.ID)
	availableByModel[itemModelID] -= shortage
	return nil
}

// newConstructorProductionJobLocked создает строку очереди по схеме; вызывается только под mutex.
func (world *World) newConstructorProductionJobLocked(constructorID int64, materialContainerID int64, productContainerID int64, queueType string, schema controlPanelItemSchema, batches int64, parentJobID int64) constructorProductionJob {
	world.nextConstructorProductionJobID++
	totalTime := math.Max(0, schema.ProductionBaseTime)
	if batches <= 0 {
		batches = 1
	}
	return constructorProductionJob{
		ID:                                world.nextConstructorProductionJobID,
		ConstructorEquipmentGroupID:       constructorID,
		MaterialContainerEquipmentGroupID: materialContainerID,
		ProductContainerEquipmentGroupID:  productContainerID,
		QueueType:                         queueType,
		SchemaID:                          schema.ID,
		ProductItemModelID:                schema.ItemModelID,
		ProductCount:                      schema.Count,
		RemainingBatches:                  batches,
		TotalBatches:                      batches,
		RemainingTime:                     totalTime,
		TotalTime:                         totalTime,
		ParentJobID:                       parentJobID,
	}
}

// newConstructorObjectProductionJobLocked создаёт строку очереди по чертежу объекта; вызывается только под mutex.
func (world *World) newConstructorObjectProductionJobLocked(constructorID int64, materialContainerID int64, queueType string, blueprint controlPanelObjectBlueprint, batches int64, parentJobID int64) constructorProductionJob {
	world.nextConstructorProductionJobID++
	totalTime := math.Max(0, blueprint.ProductionBaseTime)
	if batches <= 0 {
		batches = 1
	}
	return constructorProductionJob{
		ID:                                world.nextConstructorProductionJobID,
		ConstructorEquipmentGroupID:       constructorID,
		MaterialContainerEquipmentGroupID: materialContainerID,
		QueueType:                         queueType,
		BlueprintID:                       blueprint.ID,
		ProductCosmicObjectModelID:        blueprint.CosmicObjectModelID,
		ProductCount:                      1,
		RemainingBatches:                  batches,
		TotalBatches:                      batches,
		RemainingTime:                     totalTime,
		TotalTime:                         totalTime,
		ParentJobID:                       parentJobID,
	}
}

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

// controlledEquipmentItemtypeLocked возвращает оборудование текущего объекта с ожидаемым типом предмета; вызывается только под mutex.
// relatedContainerEquipmentLocked возвращает сохранённый контейнер для указанной группы и вида связи.
func (world *World) relatedContainerEquipmentLocked(objectID int64, equipmentGroupID int64, relationTypeAcronym string) (*data.EquipmentGroup, error) {
	if world.data.RelationTypes == nil || world.data.EquipmentGroupRelations == nil {
		return nil, errors.New("equipment relation data is not loaded")
	}
	relationType, ok := world.data.RelationTypes.GetByAcronym(relationTypeAcronym)
	if !ok {
		return nil, errors.New("relation type not found")
	}
	relation, ok := world.data.EquipmentGroupRelations.GetByEquipmentGroupAndType(equipmentGroupID, relationType.ID)
	if !ok {
		return nil, errors.New("equipment group relation not found")
	}
	return world.controlledContainerEquipmentLocked(objectID, relation.RelatedEquipmentGroupID)
}

// constructorRelatedContainerOrFallbackLocked возвращает сохранённый контейнер или старое значение команды для совместимости.
func (world *World) constructorRelatedContainerOrFallbackLocked(objectID int64, constructorID int64, relationTypeAcronym string, fallbackContainerID int64) (*data.EquipmentGroup, error) {
	container, err := world.relatedContainerEquipmentLocked(objectID, constructorID, relationTypeAcronym)
	if err == nil {
		return container, nil
	}
	if fallbackContainerID <= 0 {
		return nil, err
	}
	return world.controlledContainerEquipmentLocked(objectID, fallbackContainerID)
}

func (world *World) controlledEquipmentItemtypeLocked(objectID int64, groupID int64, itemtypeAcronym string) (*data.EquipmentGroup, error) {
	group, ok := world.data.EquipmentGroups.Get(groupID)
	if !ok {
		return nil, errors.New("equipment group not found")
	}
	if group.CosmicObjectID != objectID {
		return nil, errors.New("equipment group does not belong to controlled object")
	}
	model, ok := world.data.ItemModels.Get(group.EquipmentItemModelID)
	if !ok {
		return nil, errors.New("equipment model not found")
	}
	itemtype, ok := world.data.Itemtypes.Get(model.ItemtypeID)
	if !ok || itemtype.Acronym != itemtypeAcronym {
		return nil, errors.New("equipment group has unexpected item type")
	}
	return group, nil
}

// itemSchemaLocked разбирает одну сырую запись схемы предмета; вызывается только под mutex.
func (world *World) itemSchemaLocked(schemaID int64) (controlPanelItemSchema, error) {
	raw, ok := world.data.Schemas.Items[fmt.Sprintf("%d", schemaID)]
	if !ok {
		return controlPanelItemSchema{}, errors.New("item schema not found")
	}
	var schema controlPanelItemSchema
	if err := json.Unmarshal(raw, &schema); err != nil {
		return controlPanelItemSchema{}, err
	}
	if schema.ID != schemaID || schema.ItemModelID <= 0 {
		return controlPanelItemSchema{}, errors.New("item schema is invalid")
	}
	return schema, nil
}

// itemSchemaComponentsLocked возвращает компоненты выбранной схемы из сырого справочника; вызывается только под mutex.
// itemSchemaByProductModelLocked находит схему, производящую указанную модель предмета; вызывается только под mutex.
func (world *World) itemSchemaByProductModelLocked(itemModelID int64) (controlPanelItemSchema, error) {
	schemaIDs := make([]int64, 0, len(world.data.Schemas.Items))
	for key := range world.data.Schemas.Items {
		var schemaID int64
		if _, err := fmt.Sscanf(key, "%d", &schemaID); err == nil {
			schemaIDs = append(schemaIDs, schemaID)
		}
	}
	sort.Slice(schemaIDs, func(left int, right int) bool {
		return schemaIDs[left] < schemaIDs[right]
	})
	for _, schemaID := range schemaIDs {
		schema, err := world.itemSchemaLocked(schemaID)
		if err != nil {
			return controlPanelItemSchema{}, err
		}
		if schema.ItemModelID == itemModelID {
			return schema, nil
		}
	}
	return controlPanelItemSchema{}, errors.New("item schema not found")
}

func (world *World) itemSchemaComponentsLocked(schemaID int64) ([]controlPanelItemSchemaComponent, error) {
	components := make([]controlPanelItemSchemaComponent, 0)
	for _, raw := range world.data.SchemaComponents.Items {
		var component controlPanelItemSchemaComponent
		if err := json.Unmarshal(raw, &component); err != nil {
			return nil, err
		}
		if component.SchemaID == schemaID {
			components = append(components, component)
		}
	}
	sort.Slice(components, func(left int, right int) bool {
		return components[left].ID < components[right].ID
	})
	return components, nil
}

// consumeItemModelFromContainerLocked списывает требуемое количество предметов одной модели из контейнера; вызывается только под mutex.
// objectBlueprintLocked разбирает одну сырую запись чертежа объекта; вызывается только под mutex.
func (world *World) objectBlueprintLocked(blueprintID int64) (controlPanelObjectBlueprint, error) {
	raw, ok := world.data.Blueprints.Items[fmt.Sprintf("%d", blueprintID)]
	if !ok {
		return controlPanelObjectBlueprint{}, errors.New("object blueprint not found")
	}
	var blueprint controlPanelObjectBlueprint
	if err := json.Unmarshal(raw, &blueprint); err != nil {
		return controlPanelObjectBlueprint{}, err
	}
	if blueprint.ID != blueprintID || blueprint.CosmicObjectModelID <= 0 {
		return controlPanelObjectBlueprint{}, errors.New("object blueprint is invalid")
	}
	return blueprint, nil
}

// objectBlueprintComponentsLocked возвращает компоненты выбранного чертежа; вызывается только под mutex.
func (world *World) objectBlueprintComponentsLocked(blueprintID int64) ([]controlPanelItemSchemaComponent, error) {
	components := make([]controlPanelItemSchemaComponent, 0)
	for _, raw := range world.data.BlueprintComponents.Items {
		var component controlPanelObjectBlueprintComponent
		if err := json.Unmarshal(raw, &component); err != nil {
			return nil, err
		}
		if component.BlueprintID == blueprintID {
			components = append(components, controlPanelItemSchemaComponent{
				ID:                   component.ID,
				ComponentItemModelID: component.ComponentItemModelID,
				Count:                component.Count,
			})
		}
	}
	sort.Slice(components, func(left int, right int) bool {
		return components[left].ID < components[right].ID
	})
	return components, nil
}

func (world *World) consumeItemModelFromContainerLocked(containerID int64, itemModelID int64, amount float64) {
	remaining := amount
	for _, itemGroup := range world.data.ItemGroups.GetByContainerEquipmentGroupID(containerID) {
		if remaining <= physics.Epsilon {
			return
		}
		if itemGroup.ContentItemModelID != itemModelID {
			continue
		}
		consumed := math.Min(itemGroup.Count, remaining)
		itemGroup.Count -= consumed
		remaining -= consumed
		if itemGroup.Count <= physics.Epsilon {
			delete(world.data.ItemGroups.Items, itemGroup.ID)
		}
	}
}

// addItemModelToContainerLocked добавляет продукцию в существующую группу или создает новую; вызывается только под mutex.
func (world *World) addItemModelToContainerLocked(containerID int64, itemModelID int64, amount float64) error {
	for _, itemGroup := range world.data.ItemGroups.GetByContainerEquipmentGroupID(containerID) {
		if itemGroup.ContentItemModelID == itemModelID {
			itemGroup.Count += amount
			return nil
		}
	}
	_, err := world.data.ItemGroups.Add(&data.ItemGroup{
		ContainerEquipmentGroupID: containerID,
		ContentItemModelID:        itemModelID,
		Count:                     amount,
	})
	return err
}

// controlledFuelTankFuelModelIDLocked возвращает модель топлива для выбранного бака текущего объекта.
func (world *World) controlledFuelTankFuelModelIDLocked(objectID int64, groupID int64) (int64, error) {
	group, ok := world.data.EquipmentGroups.Get(groupID)
	if !ok {
		return 0, errors.New("equipment group not found")
	}
	if group.CosmicObjectID != objectID {
		return 0, errors.New("equipment group does not belong to controlled object")
	}
	model, ok := world.data.ItemModels.Get(group.EquipmentItemModelID)
	if !ok {
		return 0, errors.New("fuel tank model not found")
	}
	itemtype, ok := world.data.Itemtypes.Get(model.ItemtypeID)
	if !ok || itemtype.Acronym != "FuelTank" {
		return 0, errors.New("equipment group is not a fuel tank")
	}
	if model.ConsumingItemModelID <= 0 {
		return 0, errors.New("fuel tank fuel model is not set")
	}
	return model.ConsumingItemModelID, nil
}

// fillFuelFromContainerLocked забирает выбранное топливо из контейнера до заполнения общего запаса.
func (world *World) fillFuelFromContainerLocked(cosmicObject *data.CosmicObject, containerID int64, fuelModelID int64, itemGroupIDs []int64, amount float64) error {
	freeFuel := math.Max(0, cosmicObject.MaxFuel-cosmicObject.Fuel)
	if freeFuel <= 0 {
		return nil
	}
	remainingAmount := freeFuel
	if amount > 0 {
		remainingAmount = math.Min(freeFuel, amount)
	}
	for _, itemGroupID := range itemGroupIDs {
		if remainingAmount <= 0 {
			break
		}
		itemGroup, ok := world.data.ItemGroups.Get(itemGroupID)
		if !ok {
			return errors.New("item group not found")
		}
		if itemGroup.ContainerEquipmentGroupID != containerID {
			return errors.New("item group does not belong to source container")
		}
		if itemGroup.ContentItemModelID != fuelModelID {
			return errors.New("item group is not fuel for selected tank")
		}
		moved := math.Min(itemGroup.Count, remainingAmount)
		cosmicObject.Fuel += moved
		remainingAmount -= moved
		itemGroup.Count -= moved
		if itemGroup.Count <= physics.Epsilon {
			delete(world.data.ItemGroups.Items, itemGroup.ID)
		}
	}
	return nil
}

// drainFuelToContainerLocked сливает указанное топливо из общего запаса в выбранный контейнер.
func (world *World) drainFuelToContainerLocked(cosmicObject *data.CosmicObject, containerID int64, fuelModelID int64, amount float64) error {
	moved := math.Min(math.Max(0, amount), cosmicObject.Fuel)
	if moved <= 0 {
		return nil
	}
	for _, itemGroup := range world.data.ItemGroups.GetByContainerEquipmentGroupID(containerID) {
		if itemGroup.ContentItemModelID == fuelModelID {
			itemGroup.Count += moved
			cosmicObject.Fuel -= moved
			return nil
		}
	}
	if _, err := world.data.ItemGroups.Add(&data.ItemGroup{ContainerEquipmentGroupID: containerID, ContentItemModelID: fuelModelID, Count: moved}); err != nil {
		return err
	}
	cosmicObject.Fuel -= moved
	return nil
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
	world.stepConstructorProductionJobsLocked(dtSeconds)

	world.tick++
	return world.snapshotLocked(0)
}

// stepConstructorProductionJobsLocked продвигает по одному заданию на каждый конструктор за текущий шаг мира.
func (world *World) stepConstructorProductionJobsLocked(dtSeconds float64) {
	if dtSeconds <= 0 || len(world.constructorProductionJobs) == 0 {
		return
	}
	constructorIDs := world.constructorIDsWithProductionJobsLocked()
	for _, constructorID := range constructorIDs {
		jobIndex := world.activeConstructorProductionJobIndexLocked(constructorID)
		if jobIndex < 0 {
			continue
		}
		job := &world.constructorProductionJobs[jobIndex]
		if !job.Running {
			if !world.startConstructorProductionJobLocked(job) {
				continue
			}
		}
		job.RemainingTime = math.Max(0, job.RemainingTime-dtSeconds)
		if job.RemainingTime > physics.Epsilon {
			continue
		}
		if err := world.completeConstructorProductionJobLocked(job); err != nil {
			continue
		}
		job.RemainingBatches--
		if job.RemainingBatches > 0 {
			job.Running = false
			job.RemainingTime = job.TotalTime
			_ = world.data.ItemGroups.RebuildIndexes()
			continue
		}
		world.constructorProductionJobs = append(world.constructorProductionJobs[:jobIndex], world.constructorProductionJobs[jobIndex+1:]...)
		_ = world.data.ItemGroups.RebuildIndexes()
	}
}

// completeConstructorProductionJobLocked кладёт результат задания в контейнер или создаёт объект в космосе.
func (world *World) completeConstructorProductionJobLocked(job *constructorProductionJob) error {
	if job.ProductCosmicObjectModelID > 0 {
		return world.createConstructedCosmicObjectLocked(job)
	}
	productContainer, err := world.currentConstructorJobContainerLocked(job, "Destination", job.ProductContainerEquipmentGroupID)
	if err != nil {
		return err
	}
	return world.addItemModelToContainerLocked(productContainer.ID, job.ProductItemModelID, job.ProductCount)
}

// constructorIDsWithProductionJobsLocked возвращает конструкторы с очередями в стабильном порядке.
// createConstructedCosmicObjectLocked создаёт результат чертежа перед объектом-изготовителем.
func (world *World) createConstructedCosmicObjectLocked(job *constructorProductionJob) error {
	constructor, ok := world.data.EquipmentGroups.Get(job.ConstructorEquipmentGroupID)
	if !ok {
		return errors.New("constructor equipment group not found")
	}
	builder, ok := world.data.CosmicObjects.Get(constructor.CosmicObjectID)
	if !ok {
		return errors.New("builder object not found")
	}
	model, ok := world.data.CosmicObjectModels.Get(job.ProductCosmicObjectModelID)
	if !ok {
		return errors.New("blueprint object model not found")
	}
	assembly, ok := world.firstPublicDeveloperAssembly(model.ID)
	if !ok {
		return errors.New("blueprint object assembly not found")
	}
	cosmicObject := world.cosmicObjectFromModelAndAssembly(model, assembly)
	cosmicObject.OwnerCharacterID = builder.OwnerCharacterID
	cosmicObject.CreatorCharacterID = builder.OwnerCharacterID
	cosmicObject.Rotation = builder.Rotation
	cosmicObject.TargetRotation = builder.Rotation
	world.placeConstructedCosmicObjectLocked(cosmicObject, *model, *builder)
	createdObject, err := world.data.CosmicObjects.Add(cosmicObject)
	if err != nil {
		return err
	}
	return world.ensureEquipmentFromAssembly(createdObject.ID, assembly)
}

// placeConstructedCosmicObjectLocked ищет свободную точку прямо по ходу объекта-изготовителя.
func (world *World) placeConstructedCosmicObjectLocked(created *data.CosmicObject, createdModel data.CosmicObjectModel, builder data.CosmicObject) {
	builderModel, ok := world.data.CosmicObjectModels.Get(builder.CosmicObjectModelID)
	if !ok {
		created.X = builder.X
		created.Y = builder.Y
		return
	}
	forward := physics.ForwardVector(builder.Rotation)
	gap := 1.0
	baseDistance := builderModel.BodyLength/2 + createdModel.BodyLength/2 + gap
	stepDistance := math.Max(1, createdModel.BodyLength/2)
	for index := 0; index < 1000; index++ {
		distance := baseDistance + float64(index)*stepDistance
		created.X = builder.X + forward.X*distance
		created.Y = builder.Y + forward.Y*distance
		if !world.cosmicObjectIntersectsExistingLocked(*created, createdModel) {
			return
		}
	}
}

// cosmicObjectIntersectsExistingLocked проверяет пересечение кандидата с уже существующими объектами.
func (world *World) cosmicObjectIntersectsExistingLocked(candidate data.CosmicObject, candidateModel data.CosmicObjectModel) bool {
	for _, existing := range world.data.CosmicObjects.Items {
		if existing == nil {
			continue
		}
		existingModel, ok := world.data.CosmicObjectModels.Get(existing.CosmicObjectModelID)
		if !ok {
			continue
		}
		if _, collided := physics.CollisionInfo(candidate, candidateModel, *existing, *existingModel); collided {
			return true
		}
	}
	return false
}

func (world *World) constructorIDsWithProductionJobsLocked() []int64 {
	seen := map[int64]bool{}
	constructorIDs := make([]int64, 0)
	for _, job := range world.constructorProductionJobs {
		if seen[job.ConstructorEquipmentGroupID] {
			continue
		}
		seen[job.ConstructorEquipmentGroupID] = true
		constructorIDs = append(constructorIDs, job.ConstructorEquipmentGroupID)
	}
	sort.Slice(constructorIDs, func(left int, right int) bool {
		return constructorIDs[left] < constructorIDs[right]
	})
	return constructorIDs
}

// constructorMainJobIndexLocked ищет выбранную строку основной очереди конструктора.
func (world *World) constructorMainJobIndexLocked(constructorID int64, jobID int64) int {
	for index, job := range world.constructorProductionJobs {
		if job.ID == jobID && job.ConstructorEquipmentGroupID == constructorID && job.QueueType == "main" {
			return index
		}
	}
	return -1
}

// skipConstructorMainJobNextLocked оставляет только начатую единицу или убирает не начатую строку.
func (world *World) skipConstructorMainJobNextLocked(jobIndex int) {
	if jobIndex < 0 || jobIndex >= len(world.constructorProductionJobs) {
		return
	}
	job := &world.constructorProductionJobs[jobIndex]
	if !job.Running {
		world.removeConstructorMainJobAtLocked(jobIndex)
		return
	}
	job.RemainingBatches = 1
	job.TotalBatches = 1
}

// removeConstructorMainJobsAfterLocked убирает основные строки, следующие за выбранной.
func (world *World) removeConstructorMainJobsAfterLocked(constructorID int64, jobID int64) {
	seenSelected := false
	for index := 0; index < len(world.constructorProductionJobs); {
		job := world.constructorProductionJobs[index]
		if job.ConstructorEquipmentGroupID == constructorID && job.QueueType == "main" {
			if seenSelected {
				world.removeConstructorMainJobAtLocked(index)
				continue
			}
			if job.ID == jobID {
				seenSelected = true
			}
		}
		index++
	}
}

// removeConstructorMainJobsFromLocked убирает выбранную основную строку и все следующие основные строки.
func (world *World) removeConstructorMainJobsFromLocked(constructorID int64, jobID int64) {
	seenSelected := false
	for index := 0; index < len(world.constructorProductionJobs); {
		job := world.constructorProductionJobs[index]
		if job.ConstructorEquipmentGroupID == constructorID && job.QueueType == "main" && (seenSelected || job.ID == jobID) {
			seenSelected = true
			world.removeConstructorMainJobAtLocked(index)
			continue
		}
		index++
	}
}

// removeConstructorMainJobAtLocked убирает основную строку и ее вспомогательные строки.
func (world *World) removeConstructorMainJobAtLocked(jobIndex int) {
	if jobIndex < 0 || jobIndex >= len(world.constructorProductionJobs) {
		return
	}
	jobID := world.constructorProductionJobs[jobIndex].ID
	world.constructorProductionJobs = append(world.constructorProductionJobs[:jobIndex], world.constructorProductionJobs[jobIndex+1:]...)
	for index := 0; index < len(world.constructorProductionJobs); {
		if world.constructorProductionJobs[index].ParentJobID == jobID {
			world.constructorProductionJobs = append(world.constructorProductionJobs[:index], world.constructorProductionJobs[index+1:]...)
			continue
		}
		index++
	}
}

// activeConstructorProductionJobIndexLocked выбирает текущее задание конструктора с приоритетом вспомогательной очереди.
func (world *World) activeConstructorProductionJobIndexLocked(constructorID int64) int {
	for index, job := range world.constructorProductionJobs {
		if job.ConstructorEquipmentGroupID == constructorID && job.Running {
			return index
		}
	}
	for index, job := range world.constructorProductionJobs {
		if job.ConstructorEquipmentGroupID == constructorID && job.QueueType == "auxiliary" {
			return index
		}
	}
	for index, job := range world.constructorProductionJobs {
		if job.ConstructorEquipmentGroupID == constructorID && job.QueueType == "main" {
			return index
		}
	}
	return -1
}

// startConstructorProductionJobLocked списывает компоненты задания и переводит его в выполнение.
func (world *World) startConstructorProductionJobLocked(job *constructorProductionJob) bool {
	requiredByModel, ok := world.constructorProductionRequirementsLocked(job)
	if !ok {
		return false
	}
	materialContainer, err := world.currentConstructorJobContainerLocked(job, "Source", job.MaterialContainerEquipmentGroupID)
	if err != nil {
		return false
	}
	for itemModelID, required := range requiredByModel {
		world.consumeItemModelFromContainerLocked(materialContainer.ID, itemModelID, required)
	}
	job.Running = true
	job.RemainingTime = job.TotalTime
	_ = world.data.ItemGroups.RebuildIndexes()
	return true
}

// Собирает ввод подключенных аккаунтов по управляемым объектам.
// constructorJobComponentsLocked возвращает компоненты задания по схеме или чертежу.
// constructorEquipmentIsWorkingLocked проверяет, выполняет ли группа конструктора текущее или готовое к старту задание.
func (world *World) constructorEquipmentIsWorkingLocked(groupID int64) bool {
	jobIndex := world.activeConstructorProductionJobIndexLocked(groupID)
	if jobIndex < 0 {
		return false
	}
	job := &world.constructorProductionJobs[jobIndex]
	if job.Running {
		return true
	}
	_, ok := world.constructorProductionRequirementsLocked(job)
	return ok
}

// constructorProductionRequirementsLocked собирает доступные к списанию компоненты для старта строки очереди.
func (world *World) constructorProductionRequirementsLocked(job *constructorProductionJob) (map[int64]float64, bool) {
	if world.data.ItemGroups == nil {
		return nil, false
	}
	components, err := world.constructorJobComponentsLocked(job)
	if err != nil || len(components) == 0 {
		return nil, false
	}
	requiredByModel := map[int64]float64{}
	for _, component := range components {
		requiredByModel[component.ComponentItemModelID] += component.Count
	}
	availableByModel := map[int64]float64{}
	materialContainer, err := world.currentConstructorJobContainerLocked(job, "Source", job.MaterialContainerEquipmentGroupID)
	if err != nil {
		return nil, false
	}
	for _, itemGroup := range world.data.ItemGroups.GetByContainerEquipmentGroupID(materialContainer.ID) {
		availableByModel[itemGroup.ContentItemModelID] += itemGroup.Count
	}
	for itemModelID, required := range requiredByModel {
		if availableByModel[itemModelID]+physics.Epsilon < required {
			return nil, false
		}
	}
	return requiredByModel, true
}

func (world *World) constructorJobComponentsLocked(job *constructorProductionJob) ([]controlPanelItemSchemaComponent, error) {
	if job.BlueprintID > 0 {
		return world.objectBlueprintComponentsLocked(job.BlueprintID)
	}
	return world.itemSchemaComponentsLocked(job.SchemaID)
}

// currentConstructorJobContainerLocked находит актуальный контейнер задания по текущим сохранённым связям конструктора.
func (world *World) currentConstructorJobContainerLocked(job *constructorProductionJob, relationTypeAcronym string, fallbackContainerID int64) (*data.EquipmentGroup, error) {
	constructor, ok := world.data.EquipmentGroups.Get(job.ConstructorEquipmentGroupID)
	if !ok {
		return nil, errors.New("constructor equipment group not found")
	}
	return world.constructorRelatedContainerOrFallbackLocked(constructor.CosmicObjectID, job.ConstructorEquipmentGroupID, relationTypeAcronym, fallbackContainerID)
}

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

		group.Active = equipmentIsActive(*cosmicObject, *model) || world.constructorEquipmentIsWorkingLocked(group.ID)
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
	if world.data.RelationTypes != nil {
		if err := world.data.RelationTypes.SaveToFile(filepath.Join(dataDirectory, "RelationTypes.json")); err != nil {
			return err
		}
	}
	if world.data.EquipmentGroupRelations != nil {
		if err := world.data.EquipmentGroupRelations.SaveToFile(filepath.Join(dataDirectory, "EquipmentGroupRelations.json")); err != nil {
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

	equipmentGroupRelations := make([]data.EquipmentGroupRelation, 0)
	if world.data.EquipmentGroupRelations != nil {
		relationIDs := make([]int64, 0, len(world.data.EquipmentGroupRelations.Items))
		for relationID := range world.data.EquipmentGroupRelations.Items {
			relationIDs = append(relationIDs, relationID)
		}
		sort.Slice(relationIDs, func(left int, right int) bool {
			return relationIDs[left] < relationIDs[right]
		})
		for _, relationID := range relationIDs {
			relation := world.data.EquipmentGroupRelations.Items[relationID]
			if relation == nil {
				continue
			}
			equipmentGroupRelations = append(equipmentGroupRelations, *relation)
		}
	}

	constructorProductionJobs := make([]game.ConstructorProductionJob, 0, len(world.constructorProductionJobs))
	for _, job := range world.constructorProductionJobs {
		constructorProductionJobs = append(constructorProductionJobs, game.ConstructorProductionJob{
			ID:                          job.ID,
			ConstructorEquipmentGroupID: job.ConstructorEquipmentGroupID,
			QueueType:                   job.QueueType,
			SchemaID:                    job.SchemaID,
			BlueprintID:                 job.BlueprintID,
			ProductItemModelID:          job.ProductItemModelID,
			ProductCosmicObjectModelID:  job.ProductCosmicObjectModelID,
			ProductCount:                job.ProductCount,
			RemainingCount:              float64(job.RemainingBatches) * job.ProductCount,
			TotalCount:                  float64(job.TotalBatches) * job.ProductCount,
			RemainingTime:               job.RemainingTime,
			TotalTime:                   job.TotalTime,
			Running:                     job.Running,
			ParentJobID:                 job.ParentJobID,
		})
	}

	return game.Snapshot{
		Type:                      "snapshot",
		Tick:                      world.tick,
		SelfObjectID:              selfObjectID,
		Objects:                   objects,
		EquipmentGroups:           equipmentGroups,
		EquipmentGroupRelations:   equipmentGroupRelations,
		ItemGroups:                itemGroups,
		ConstructorProductionJobs: constructorProductionJobs,
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
	resourceModelIDs := world.resourceItemModelIDs()
	for _, itemModelID := range resourceModelIDs {
		_, _ = world.data.ItemGroups.Add(&data.ItemGroup{
			ContainerEquipmentGroupID: containerIDs[0],
			ContentItemModelID:        itemModelID,
			Count:                     1000,
		})
	}
}

// Возвращает первую модель предмета каждого типа в стабильном порядке типов.
func (world *World) firstItemModelIDsByType() []int64 {
	firstByType := make(map[int64]int64)
	resourceTypeID := int64(0)
	if resourceType, ok := world.data.Itemtypes.GetByAcronym("Resource"); ok {
		resourceTypeID = resourceType.ID
	}
	for itemModelID, model := range world.data.ItemModels.Items {
		if model == nil || model.ItemtypeID <= 0 || model.ItemtypeID == resourceTypeID {
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

// resourceItemModelIDs возвращает все модели ресурсов в стабильном порядке для тестового запаса нового корабля.
func (world *World) resourceItemModelIDs() []int64 {
	resourceType, ok := world.data.Itemtypes.GetByAcronym("Resource")
	if !ok || world.data.ItemModels == nil {
		return nil
	}
	result := make([]int64, 0)
	for itemModelID, model := range world.data.ItemModels.Items {
		if model != nil && model.ItemtypeID == resourceType.ID {
			result = append(result, itemModelID)
		}
	}
	sort.Slice(result, func(left int, right int) bool {
		return result[left] < result[right]
	})
	return result
}
