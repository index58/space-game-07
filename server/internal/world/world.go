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
	"time"

	"space-game-07-server/internal/data"
	"space-game-07-server/internal/game"
	"space-game-07-server/internal/physics"
	"space-game-07-server/internal/storage"
)

const (
	defaultAccountEmailDomain = "auto.local"
	defaultAccountPassword    = "auto"
	defaultStarterShipAcronym = "ship_bat"
	dockingDurationSeconds    = 10
	dockingProbeDistance      = 10
	miningNotificationSeconds = 1
	pilotToolSlotCount        = 10
	simpleDrillAcronym        = "SimpleDrill"
	drillRayAcronym           = "DrillRay"
	weaponItemTypeAcronym     = "Weapon"
)

// cross2D считает псевдоскалярное произведение двух плоских векторов.
type pilotInstrumentModel struct {
	ModelID      int64 // Уникальный идентификатор модели предмета.
	FirstGroupID int64 // Уникальный идентификатор первой группы оборудования этой модели.
	EnabledCount int64 // Количество включенных единиц оборудования этой модели.
}

// cross2D считает псевдоскалярное произведение двух плоских векторов.
// cross2D считает псевдоскалярное произведение двух плоских векторов.
type drillMiningParameters struct {
	Range        float64 // Дальность действия луча в метрах.
	MiningSpeed  float64 // Масса добываемого ресурса в килограммах за секунду одной установленной единицей.
	EnabledCount int64   // Количество включенных единиц выбранной модели.
}

// cross2D считает псевдоскалярное произведение двух плоских векторов.
type weaponAttackParameters struct {
	ItemModelID            int64               // Модель выбранного оружия для учёта темпа выстрелов.
	ProjectileModelID      int64               // Модель космического объекта, который появляется после выстрела.
	Damage                 float64             // Урон броне при одном попадании.
	ProjectileSpeed        float64             // Скорость движения в метрах за секунду.
	ShotsPerSecond         float64             // Количество выстрелов за секунду с учётом включённых единиц.
	InitialProjectileCount int64               // Количество объектов, появляющихся сразу при начале стрельбы.
	Range                  float64             // Дальность попадания в метрах.
	Groups                 []weaponAttackGroup // Группы оружия, которые участвуют в выбранном огне.
}

// cross2D считает псевдоскалярное произведение двух плоских векторов.
type weaponAttackGroup struct {
	BarrelStartIndex       int64                // Первый поперечный номер орудия этой группы среди всех выбранных орудий модели.
	EquipmentGroup         *data.EquipmentGroup // Заряженная группа, из которой расходуются боеприпасы.
	EnabledCount           int64                // Количество работающих единиц, создающих выстрелы.
	ShotsPerSecond         float64              // Темп появления снарядов с учетом работающих единиц.
	InitialProjectileCount int64                // Стартовое число снарядов при начале удержания огня.
}

// cross2D считает псевдоскалярное произведение двух плоских векторов.
type activeProjectile struct {
	ID              int64             // Временный отрицательный идентификатор для снимков мира.
	SourceObjectID  int64             // Объект, выпустивший боеприпас.
	CosmicObject    data.CosmicObject // Видимое положение и модель в снимке мира.
	Damage          float64           // Урон броне при попадании.
	VelocityX       float64           // Горизонтальная часть скорости в метрах за секунду.
	VelocityY       float64           // Вертикальная часть скорости в метрах за секунду.
	ProjectileSpeed float64           // Собственная скорость снаряда относительно выпустившего корабля.
	RemainingRange  float64           // Оставшаяся дальность полета в метрах.
}

// cross2D считает псевдоскалярное произведение двух плоских векторов.
type weaponShotKey struct {
	ObjectID         int64 // Объект, который ведёт огонь.
	EquipmentGroupID int64 // Группа оружия, для которой считается накопленный темп.
}

// cross2D считает псевдоскалярное произведение двух плоских векторов.
type miningNotificationKey struct {
	ObjectID    int64 // Корабль, которому нужно отправлять уведомление.
	ItemModelID int64 // Ресурс, количество которого накапливается.
}

// cross2D считает псевдоскалярное произведение двух плоских векторов.
type miningNotificationAccumulator struct {
	Seconds float64 // Время активной добычи после прошлого уведомления.
	Count   float64 // Количество ресурса, добытое за накопленное время.
}

type Data struct {
	Accounts                   *data.Accounts                   // Учетные записи, доступные игровой симуляции.
	Characters                 *data.Characters                 // Персонажи, доступные игровой симуляции.
	CosmicObjects              *data.CosmicObjects              // Экземпляры объектов, участвующие в мире.
	CosmicObjectTypes          *data.CosmicObjectTypes          // Справочник типов объектов для правил мира.
	CosmicObjectModels         *data.CosmicObjectModels         // Справочник моделей объектов для физики и отображения.
	ItemTypes                  *data.ItemTypes                  // Справочник типов предметов для серверной логики.
	ItemModels                 *data.ItemModels                 // Справочник моделей предметов для оборудования и содержимого контейнеров.
	Blueprints                 *storage.RawReferenceTable       // Справочник чертежей объектов для изготовления в конструкторе.
	BlueprintComponents        *storage.RawReferenceTable       // Справочник компонентов чертежей объектов для списания материалов.
	Schemas                    *storage.RawReferenceTable       // Справочник схем предметов для изготовления в конструкторе.
	SchemaComponents           *storage.RawReferenceTable       // Справочник компонентов схем предметов для списания материалов.
	TaskTypes                  *data.TaskTypes                  // Справочник типов заданий.
	Tasks                      *data.Tasks                      // Сохраненные задания оборудования.
	TaskItemGroups             *data.TaskItemGroups             // Зарезервированные предметы заданий.
	Implementers               *data.Implementers               // Исполнители типов заданий.
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

// cross2D считает псевдоскалярное произведение двух плоских векторов.
type World struct {
	mu                             sync.Mutex                                              // Защищает изменяемое состояние мира от параллельных горутин.
	tick                           int64                                                   // Номер последнего выполненного шага симуляции.
	currentTimeMillis              int64                                                   // Текущее время симуляции в миллисекундах для перезарядки.
	data                           Data                                                    // Справочники и игровые сущности, которыми управляет мир.
	accountObjectIDs               map[int64]int64                                         // Связь подключенных аккаунтов с управляемыми объектами.
	inputs                         map[int64]game.ShipInput                                // Последний принятый ввод для каждого подключенного аккаунта.
	mutationAcks                   map[string]int64                                        // Последний обработанный номер команды панели по аккаунту и сессии.
	random                         *rand.Rand                                              // Источник случайности для воспроизводимых команд.
	nextConstructorProductionJobID int64                                                   // Следующий идентификатор задания изготовления.
	constructorProductionJobs      []constructorProductionJob                              // Задания изготовления, ожидающие или выполняющиеся в конструкторах.
	dockingRequests                []dockingRequest                                        // Активные запросы стыковки, ожидающие ответа.
	dockingProcesses               []dockingProcess                                        // Активные автоматические стыковки, хранящиеся только в памяти.
	landingRequests                []landingRequest                                        // Активные запросы посадки персонажа, ожидающие ответа.
	dockingEvents                  []game.DockingEvent                                     // Накопленные клиентские события стыковки до ближайшей рассылки.
	exchangeRequests               []exchangeRequest                                       // Ожидающие ответы на запросы обмена.
	exchangeSessions               []exchangeSession                                       // Открытые и выполняющиеся обмены.
	exchangeEvents                 []game.ExchangeEvent                                    // Накопленные клиентские события обмена.
	nextProjectileID               int64                                                   // Следующий отрицательный номер для временных боеприпасов.
	projectiles                    []activeProjectile                                      // Летящие боеприпасы, видимые в снимках мира.
	weaponShotAccumulators         map[weaponShotKey]float64                               // Накопители темпа стрельбы для удерживаемого огня.
	miningNotifications            map[miningNotificationKey]miningNotificationAccumulator // Накопленная добыча для периодических уведомлений.
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
	RemainingBatches                  int64   // пїЅпїЅпїЅпїЅпїЅпїЅпїЅпїЅпїЅпїЅ пїЅпїЅпїЅпїЅпїЅпїЅпїЅпїЅ, пїЅпїЅпїЅпїЅпїЅпїЅпїЅ пїЅпїЅпїЅ пїЅпїЅпїЅпїЅпїЅпїЅпїЅпїЅ пїЅпїЅпїЅпїЅпїЅпїЅпїЅпїЅпїЅ пїЅпїЅ пїЅпїЅпїЅпїЅпїЅпїЅ.
	TotalBatches                      int64   // пїЅпїЅпїЅпїЅпїЅ пїЅпїЅпїЅпїЅпїЅпїЅпїЅпїЅпїЅпїЅ пїЅпїЅпїЅпїЅпїЅпїЅпїЅпїЅ, пїЅпїЅпїЅпїЅпїЅпїЅпїЅпїЅпїЅпїЅпїЅпїЅпїЅпїЅпїЅ пїЅпїЅ пїЅпїЅпїЅпїЅпїЅпїЅ.
	RemainingTime                     float64 // Оставшееся время изготовления в секундах.
	TotalTime                         float64 // Полное время изготовления в секундах.
	Running                           bool    // Показывает, что задание сейчас выполняется.
	ParentJobID                       int64   // пїЅпїЅпїЅпїЅпїЅпїЅпїЅпїЅпїЅпїЅпїЅпїЅ пїЅпїЅпїЅпїЅпїЅпїЅ, пїЅпїЅпїЅ пїЅпїЅпїЅпїЅпїЅпїЅпїЅпїЅпїЅпїЅ пїЅпїЅпїЅпїЅпїЅпїЅпїЅ пїЅпїЅпїЅпїЅпїЅ пїЅпїЅпїЅ пїЅпїЅпїЅпїЅпїЅпїЅпїЅпїЅпїЅпїЅпїЅпїЅпїЅпїЅпїЅ пїЅпїЅпїЅпїЅпїЅпїЅ.
}

type dockingProcess struct {
	SenderCosmicObjectID   int64   // Объект, который отправил запрос.
	ReceiverCosmicObjectID int64   // Объект, который должен принять решение.
	RemainingSeconds       float64 // Оставшееся время ожидания ответа.
}

type dockingRequest struct {
	SenderCosmicObjectID   int64   // Объект, который отправил запрос.
	ReceiverCosmicObjectID int64   // Объект, который должен принять решение.
	RemainingSeconds       float64 // Оставшееся время ожидания ответа.
}

type landingRequest struct {
	CharacterID            int64 // Персонаж, который пересаживается.
	SenderCosmicObjectID   int64 // Объект отправления, где персонаж остается до решения.
	ReceiverCosmicObjectID int64 // Объект назначения, который должен принять решение.
}

// cross2D считает псевдоскалярное произведение двух плоских векторов.
type ControlPanelObjectUpdate struct {
	Enabled *bool   // Новое состояние включения объекта, если оно меняется.
	Title   *string // Новое пользовательское название объекта, если оно меняется.
}

// cross2D считает псевдоскалярное произведение двух плоских векторов.
type ControlPanelEquipmentUpdate struct {
	Title            *string // Новое пользовательское название группы, если оно меняется.
	EquipmentGroupID int64   // Группа оборудования, которую нужно изменить.
	Enabled          *bool   // Новое состояние включения группы, если оно меняется.
	EnabledCount     *int64  // Новое количество включенных единиц, если оно меняется.
}

// cross2D считает псевдоскалярное произведение двух плоских векторов.
// cross2D считает псевдоскалярное произведение двух плоских векторов.
type ControlPanelEquipmentGroupRelationUpdate struct {
	EquipmentGroupID        int64  // Группа оборудования, для которой сохраняется выбор.
	RelationTypeAcronym     string // Вид связи, выбранный по неизменяемому строковому идентификатору.
	RelatedEquipmentGroupID int64  // Группа оборудования, выбранная игроком в связанной панели.
}

type ControlPanelContainerTransfer struct {
	ControllerEquipmentGroupID      int64   // Правая группа контейнеров, управляющая очередью перемещений.
	LeftToRightDirection            bool    // Перемещается ли груз из левого контейнера в правый.
	SourceContainerEquipmentGroupID int64   // Контейнер с предметами, выбранный в правой части панели.
	TargetContainerEquipmentGroupID int64   // Контейнер результата, выбранный в левой части панели.
	ItemGroupIDs                    []int64 // Строки предметов, выбранные для деконструкции.
	Amount                          float64 // Максимальное количество предметов одной выбранной строки для деконструкции.
}

// cross2D считает псевдоскалярное произведение двух плоских векторов.
// cross2D считает псевдоскалярное произведение двух плоских векторов.
type ControlPanelFuelTransfer struct {
	ContainerEquipmentGroupID int64   // Группа контейнеров, из которой берется или куда сливается топливо.
	FuelTankEquipmentGroupID  int64   // Группа топливных баков, с которой работает игрок.
	ItemGroupIDs              []int64 // Группы топлива, выбранные для заливки из контейнера.
	Amount                    float64 // Количество топлива для слива из бака в контейнер.
}

// cross2D считает псевдоскалярное произведение двух плоских векторов.
// cross2D считает псевдоскалярное произведение двух плоских векторов.
type ControlPanelItemDeconstruction struct {
	DeconstructorEquipmentGroupID   int64   // Группа деконструктора, которая управляет очередью заданий.
	SourceContainerEquipmentGroupID int64   // Контейнер с предметами, выбранный в правой части панели.
	TargetContainerEquipmentGroupID int64   // Контейнер результата, выбранный в левой части панели.
	ItemGroupIDs                    []int64 // Строки предметов, выбранные для деконструкции.
	Amount                          float64 // Максимальное количество предметов одной выбранной строки для деконструкции.
}

type ControlPanelConstructorProduceItem struct {
	ConstructorEquipmentGroupID       int64 // Группа конструкторов, которая выполняет изготовление.
	MaterialContainerEquipmentGroupID int64 // Контейнер, из которого списываются компоненты схемы.
	ProductContainerEquipmentGroupID  int64 // Контейнер, в который кладется готовая продукция.
	SchemaID                          int64 // Схема предмета, выбранная игроком для изготовления.
	BlueprintID                       int64 // Чертёж объекта, выбранный игроком для изготовления.
	Amount                            int64 // Количество запусков изготовления по выбранной схеме.
}

// cross2D считает псевдоскалярное произведение двух плоских векторов.
type ControlPanelConstructorQueueCommand struct {
	ConstructorEquipmentGroupID int64  // Группа конструкторов, очередь которой меняется.
	JobID                       int64  // Строка основной очереди, выбранная игроком.
	Command                     string // Действие над выбранной строкой и следующими строками.
}

type controlPanelItemSchema struct {
	ID               int64   `json:"ID"`               // Уникальный числовой идентификатор схемы.
	ItemModelID      int64   `json:"ItemModelID"`      // Модель предмета, получаемого по схеме.
	Count            float64 `json:"Count"`            // Количество предметов, получаемое за одно изготовление.
	ProductionEnergy float64 `json:"ProductionEnergy"` // Базовое время изготовления, пока не используемое мгновенной командой.
}

// cross2D считает псевдоскалярное произведение двух плоских векторов.
type controlPanelItemSchemaComponent struct {
	ID                   int64   `json:"ID"`                   // Уникальный числовой идентификатор компонента чертежа.
	SchemaID             int64   `json:"SchemaID"`             // Схема, к которой относится компонент.
	ComponentItemModelID int64   `json:"ComponentItemModelID"` // Модель предмета, которую нужно списать как компонент.
	Count                float64 `json:"Count"`                // Количество компонента, требуемое для одного изготовления.
}

type controlPanelObjectBlueprint struct {
	ID                  int64   `json:"ID"`                  // Уникальный числовой идентификатор чертежа.
	CosmicObjectModelID int64   `json:"CosmicObjectModelID"` // Модель космического объекта, получаемого по чертежу.
	ProductionEnergy    float64 `json:"ProductionEnergy"`    // Базовое время изготовления одного объекта.
}

type controlPanelObjectBlueprintComponent struct {
	ID                   int64   `json:"ID"`                   // Уникальный числовой идентификатор компонента чертежа.
	BlueprintID          int64   `json:"BlueprintID"`          // Чертёж, к которому относится компонент.
	ComponentItemModelID int64   `json:"ComponentItemModelID"` // Модель предмета, которую нужно списать как компонент.
	Count                float64 `json:"Count"`                // Количество компонента, требуемое для одного изготовления.
}

func New(seed int64, serverData Data) *World {
	created := &World{
		data:                   serverData,
		accountObjectIDs:       map[int64]int64{},
		inputs:                 map[int64]game.ShipInput{},
		mutationAcks:           map[string]int64{},
		currentTimeMillis:      time.Now().UnixMilli(),
		nextProjectileID:       -1,
		weaponShotAccumulators: map[weaponShotKey]float64{},
		miningNotifications:    map[miningNotificationKey]miningNotificationAccumulator{},
		random:                 rand.New(rand.NewSource(seed)),
	}
	created.ensureChatData()
	created.applyAssembliesToLoadedShips()
	return created
}

// cross2D считает псевдоскалярное произведение двух плоских векторов.
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

// cross2D считает псевдоскалярное произведение двух плоских векторов.
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

// cross2D считает псевдоскалярное произведение двух плоских векторов.
func (world *World) DisconnectAccount(accountID int64) {
	world.mu.Lock()
	defer world.mu.Unlock()

	delete(world.accountObjectIDs, accountID)
	delete(world.inputs, accountID)
}

// cross2D считает псевдоскалярное произведение двух плоских векторов.
func (world *World) SetInput(accountID int64, input game.ShipInput) {
	world.mu.Lock()
	defer world.mu.Unlock()

	objectID, ok := world.accountObjectIDs[accountID]
	if !ok {
		return
	}

	if input.ToggleAnchor {
		if cosmicObject, ok := world.data.CosmicObjects.Get(objectID); ok {
			if cosmicObject.ClusterMainCosmicObjectID == 0 && (cosmicObject.Anchored || cosmicObjectIsFullyStopped(*cosmicObject)) {
				cosmicObject.Anchored = !cosmicObject.Anchored
			}
		}
		input.ToggleAnchor = false
	}
	world.inputs[accountID] = input
}

// cross2D считает псевдоскалярное произведение двух плоских векторов.
func (world *World) SendDockingRequest(accountID int64) error {
	world.mu.Lock()
	defer world.mu.Unlock()

	sender, err := world.controlledCosmicObjectLocked(accountID)
	if err != nil {
		return err
	}
	if err := world.validateDockingSenderLocked(sender); err != nil {
		return err
	}
	receiver, err := world.findDockingReceiverLocked(sender)
	if err != nil {
		return err
	}
	if err := world.validateDockingReceiverLocked(receiver); err != nil {
		return err
	}
	if world.dockingObjectIsBusyLocked(sender.ID) || world.dockingObjectIsBusyLocked(receiver.ID) {
		return errors.New("object already participates in docking")
	}
	if receiver.OwnerCharacterID == sender.OwnerCharacterID {
		world.startDockingProcessLocked(sender.ID, receiver.ID)
		return nil
	}
	if !world.dockingReceiverHasDecisionMakerLocked(receiver.ID) {
		world.addDockingNotificationLocked([]int64{sender.ID}, "В Получателе нет персонажа для принятия решения")
		return nil
	}
	world.dockingRequests = append(world.dockingRequests, dockingRequest{
		SenderCosmicObjectID:   sender.ID,
		ReceiverCosmicObjectID: receiver.ID,
		RemainingSeconds:       dockingDurationSeconds,
	})
	world.addDockingWindowEventsLocked("dockingRequestStarted", sender.ID, receiver.ID, dockingDurationSeconds)
	return nil
}

// cross2D считает псевдоскалярное произведение двух плоских векторов.
func (world *World) ApproveDockingRequest(accountID int64) error {
	world.mu.Lock()
	defer world.mu.Unlock()

	receiver, err := world.controlledCosmicObjectLocked(accountID)
	if err != nil {
		return err
	}
	requestIndex := world.dockingRequestIndexByReceiverLocked(receiver.ID)
	if requestIndex < 0 {
		return errors.New("docking request not found")
	}
	request := world.dockingRequests[requestIndex]
	sender, ok := world.data.CosmicObjects.Get(request.SenderCosmicObjectID)
	if !ok {
		world.removeDockingRequestLocked(requestIndex)
		return errors.New("sender object not found")
	}
	if err := world.validateDockingSenderLocked(sender); err != nil {
		world.removeDockingRequestLocked(requestIndex)
		world.closeDockingRequestWindowLocked(request)
		world.addDockingNotificationLocked([]int64{request.SenderCosmicObjectID, request.ReceiverCosmicObjectID}, "Условия стыковки больше не выполняются")
		return err
	}
	if err := world.validateDockingReceiverLocked(receiver); err != nil {
		world.removeDockingRequestLocked(requestIndex)
		world.closeDockingRequestWindowLocked(request)
		world.addDockingNotificationLocked([]int64{request.SenderCosmicObjectID, request.ReceiverCosmicObjectID}, "Условия стыковки больше не выполняются")
		return err
	}
	world.removeDockingRequestLocked(requestIndex)
	world.startDockingProcessLocked(sender.ID, receiver.ID)
	return nil
}

// cross2D считает псевдоскалярное произведение двух плоских векторов.
func (world *World) RejectDockingRequest(accountID int64) error {
	world.mu.Lock()
	defer world.mu.Unlock()

	receiver, err := world.controlledCosmicObjectLocked(accountID)
	if err != nil {
		return err
	}
	requestIndex := world.dockingRequestIndexByReceiverLocked(receiver.ID)
	if requestIndex < 0 {
		return errors.New("docking request not found")
	}
	request := world.dockingRequests[requestIndex]
	world.removeDockingRequestLocked(requestIndex)
	world.closeDockingRequestWindowLocked(request)
	world.addDockingNotificationLocked([]int64{request.SenderCosmicObjectID, request.ReceiverCosmicObjectID}, "Отказ на запрос стыковки")
	return nil
}

// cross2D считает псевдоскалярное произведение двух плоских векторов.
func (world *World) UndockControlledObject(accountID int64) error {
	world.mu.Lock()
	defer world.mu.Unlock()

	cosmicObject, err := world.controlledCosmicObjectLocked(accountID)
	if err != nil {
		return err
	}
	if world.exchangeClusterIsBusyLocked(cosmicObject.ID) {
		return errors.New("object participates in exchange")
	}
	mainID := cosmicObject.ClusterMainCosmicObjectID
	if mainID <= 0 {
		return errors.New("object is not docked")
	}
	notificationObjectIDs := world.clusterObjectIDsLocked(mainID)
	if mainID == cosmicObject.ID {
		world.disbandClusterLocked(mainID)
		world.addDockingNotificationLocked(notificationObjectIDs, "Объект отстыкован")
		return nil
	}
	cosmicObject.ClusterMainCosmicObjectID = 0
	cosmicObject.Anchored = false
	if len(world.clusterObjectIDsLocked(mainID)) <= 1 {
		world.disbandClusterLocked(mainID)
	}
	world.addDockingNotificationLocked(notificationObjectIDs, "Объект отстыкован")
	return nil
}

// cross2D считает псевдоскалярное произведение двух плоских векторов.
func (world *World) BeginCharacterTransfer(accountID int64) error {
	world.mu.Lock()
	defer world.mu.Unlock()

	character, err := world.currentCharacterLocked(accountID)
	if err != nil {
		return err
	}
	sender, err := world.controlledCosmicObjectLocked(accountID)
	if err != nil {
		return err
	}
	if sender.ClusterMainCosmicObjectID <= 0 {
		world.addDockingNotificationLocked([]int64{sender.ID}, "Объект не пристыкован")
		return nil
	}
	targetID, ok := world.autoLandingTargetIDLocked(sender)
	if !ok {
		world.addLandingTargetSelectionLocked(sender.ID)
		return nil
	}
	return world.requestCharacterLandingLocked(character.ID, sender.ID, targetID)
}

// cross2D считает псевдоскалярное произведение двух плоских векторов.
func (world *World) RequestCharacterLanding(accountID int64, targetID int64) error {
	world.mu.Lock()
	defer world.mu.Unlock()

	character, err := world.currentCharacterLocked(accountID)
	if err != nil {
		return err
	}
	sender, err := world.controlledCosmicObjectLocked(accountID)
	if err != nil {
		return err
	}
	if sender.ClusterMainCosmicObjectID <= 0 {
		world.addDockingNotificationLocked([]int64{sender.ID}, "Объект не пристыкован")
		return nil
	}
	return world.requestCharacterLandingLocked(character.ID, sender.ID, targetID)
}

// cross2D считает псевдоскалярное произведение двух плоских векторов.
func (world *World) ApproveCharacterLanding(accountID int64) error {
	world.mu.Lock()
	defer world.mu.Unlock()

	receiver, err := world.controlledCosmicObjectLocked(accountID)
	if err != nil {
		return err
	}
	requestIndex := world.landingRequestIndexByReceiverLocked(receiver.ID)
	if requestIndex < 0 {
		return errors.New("landing request not found")
	}
	request := world.landingRequests[requestIndex]
	world.removeLandingRequestLocked(requestIndex)
	world.moveCharacterToObjectLocked(request.CharacterID, request.ReceiverCosmicObjectID)
	world.addLandingWindowEventsLocked("landingFinished", request.SenderCosmicObjectID, request.ReceiverCosmicObjectID)
	return nil
}

// cross2D считает псевдоскалярное произведение двух плоских векторов.
func (world *World) RejectCharacterLanding(accountID int64) error {
	world.mu.Lock()
	defer world.mu.Unlock()

	receiver, err := world.controlledCosmicObjectLocked(accountID)
	if err != nil {
		return err
	}
	requestIndex := world.landingRequestIndexByReceiverLocked(receiver.ID)
	if requestIndex < 0 {
		return errors.New("landing request not found")
	}
	request := world.landingRequests[requestIndex]
	world.removeLandingRequestLocked(requestIndex)
	world.addLandingWindowEventsLocked("landingFinished", request.SenderCosmicObjectID, request.ReceiverCosmicObjectID)
	return nil
}

// cross2D считает псевдоскалярное произведение двух плоских векторов.
func (world *World) ClaimFocusedObjectOwnerForTesting(accountID int64) error {
	world.mu.Lock()
	defer world.mu.Unlock()

	character, err := world.currentCharacterLocked(accountID)
	if err != nil {
		return err
	}
	cosmicObject, err := world.controlledCosmicObjectLocked(accountID)
	if err != nil {
		return err
	}
	target, err := world.findDockingReceiverLocked(cosmicObject)
	if err != nil {
		return err
	}
	target.OwnerCharacterID = character.ID
	target.OwnerNpcClanID = 0
	return world.data.CosmicObjects.RebuildIndexes()
}

// cross2D считает псевдоскалярное произведение двух плоских векторов.
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
	character, err := world.currentCharacterLocked(accountID)
	if err != nil {
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
	cosmicObject.OwnerCharacterID = character.ID
	cosmicObject.OwnerNpcClanID = 0
	cosmicObject.TargetRotation = cosmicObject.Rotation
	if err := world.replaceEquipmentFromAssembly(cosmicObject.ID, assembly); err != nil {
		return false
	}
	cosmicObject.Armor = cosmicObject.MaxArmor
	world.fillShipSupplies(cosmicObject)

	return world.data.CosmicObjects.RebuildIndexes() == nil
}

// cross2D считает псевдоскалярное произведение двух плоских векторов.
func (world *World) controlledCosmicObjectLocked(accountID int64) (*data.CosmicObject, error) {
	objectID, ok := world.accountObjectIDs[accountID]
	if !ok {
		return nil, errors.New("account is not connected")
	}
	cosmicObject, ok := world.data.CosmicObjects.Get(objectID)
	if !ok {
		return nil, errors.New("controlled object not found")
	}
	return cosmicObject, nil
}

// cross2D считает псевдоскалярное произведение двух плоских векторов.
func (world *World) currentCharacterLocked(accountID int64) (*data.Character, error) {
	account, ok := world.data.Accounts.Get(accountID)
	if !ok || account.CurrentCharacterID <= 0 {
		return nil, errors.New("account is not connected")
	}
	character, ok := world.data.Characters.Get(account.CurrentCharacterID)
	if !ok || character.AccountID != account.ID {
		return nil, errors.New("current character not found")
	}
	return character, nil
}

// cross2D считает псевдоскалярное произведение двух плоских векторов.
func (world *World) validateDockingSenderLocked(sender *data.CosmicObject) error {
	if sender == nil {
		return errors.New("sender object not found")
	}
	if !world.cosmicObjectHasTypeLocked(sender, "Ship") {
		return errors.New("sender must be a ship")
	}
	if sender.ClusterMainCosmicObjectID > 0 {
		return errors.New("sender is already docked")
	}
	if !cosmicObjectIsFullyStopped(*sender) {
		return errors.New("sender is not stopped")
	}
	return nil
}

// cross2D считает псевдоскалярное произведение двух плоских векторов.
func (world *World) validateDockingReceiverLocked(receiver *data.CosmicObject) error {
	if receiver == nil {
		return errors.New("receiver object not found")
	}
	if !world.cosmicObjectHasTypeLocked(receiver, "Ship") && !world.cosmicObjectHasTypeLocked(receiver, "Station") {
		return errors.New("receiver must be a ship or station")
	}
	if receiver.ClusterMainCosmicObjectID > 0 && receiver.ClusterMainCosmicObjectID != receiver.ID {
		return errors.New("receiver is secondary object")
	}
	return nil
}

// cross2D считает псевдоскалярное произведение двух плоских векторов.
func (world *World) cosmicObjectHasTypeLocked(cosmicObject *data.CosmicObject, acronym string) bool {
	if cosmicObject == nil {
		return false
	}
	model, ok := world.data.CosmicObjectModels.Get(cosmicObject.CosmicObjectModelID)
	if !ok {
		return false
	}
	objectType, ok := world.data.CosmicObjectTypes.Get(model.CosmicObjectTypeID)
	return ok && objectType.Acronym == acronym
}

// cross2D считает псевдоскалярное произведение двух плоских векторов.
func (world *World) findDockingReceiverLocked(sender *data.CosmicObject) (*data.CosmicObject, error) {
	senderModel, ok := world.data.CosmicObjectModels.Get(sender.CosmicObjectModelID)
	if !ok {
		return nil, errors.New("sender model not found")
	}
	startX := sender.X + math.Sin(sender.Rotation)*senderModel.BodyLength/2
	startY := sender.Y + math.Cos(sender.Rotation)*senderModel.BodyLength/2
	endX := startX + math.Sin(sender.Rotation)*dockingProbeDistance
	endY := startY + math.Cos(sender.Rotation)*dockingProbeDistance

	var selected *data.CosmicObject
	selectedDistance := math.Inf(1)
	for _, candidate := range world.data.CosmicObjects.Items {
		if candidate == nil || candidate.ID == sender.ID {
			continue
		}
		candidateModel, ok := world.data.CosmicObjectModels.Get(candidate.CosmicObjectModelID)
		if !ok {
			continue
		}
		distance, ok := raySegmentPolygonDistance(startX, startY, endX, endY, *candidate, *candidateModel)
		if !ok || distance >= selectedDistance {
			continue
		}
		selected = candidate
		selectedDistance = distance
	}
	if selected == nil {
		return nil, errors.New("docking receiver not found")
	}
	return selected, nil
}

// cross2D считает псевдоскалярное произведение двух плоских векторов.
func (world *World) startDockingProcessLocked(senderID int64, receiverID int64) {
	if sender, ok := world.data.CosmicObjects.Get(senderID); ok {
		sender.Anchored = true
	}
	if receiver, ok := world.data.CosmicObjects.Get(receiverID); ok {
		receiver.Anchored = true
	}
	world.dockingProcesses = append(world.dockingProcesses, dockingProcess{
		SenderCosmicObjectID:   senderID,
		ReceiverCosmicObjectID: receiverID,
		RemainingSeconds:       dockingDurationSeconds,
	})
	world.addDockingWindowEventsLocked("dockingProcessStarted", senderID, receiverID, dockingDurationSeconds)
}

// cross2D считает псевдоскалярное произведение двух плоских векторов.
func (world *World) dockingObjectIsBusyLocked(objectID int64) bool {
	for _, request := range world.dockingRequests {
		if request.SenderCosmicObjectID == objectID || request.ReceiverCosmicObjectID == objectID {
			return true
		}
	}
	for _, process := range world.dockingProcesses {
		if process.SenderCosmicObjectID == objectID || process.ReceiverCosmicObjectID == objectID {
			return true
		}
	}
	return false
}

// cross2D считает псевдоскалярное произведение двух плоских векторов.
func (world *World) dockingRequestIndexByReceiverLocked(receiverID int64) int {
	for index, request := range world.dockingRequests {
		if request.ReceiverCosmicObjectID == receiverID {
			return index
		}
	}
	return -1
}

// cross2D считает псевдоскалярное произведение двух плоских векторов.
func (world *World) removeDockingRequestLocked(index int) {
	world.dockingRequests = append(world.dockingRequests[:index], world.dockingRequests[index+1:]...)
}

// cross2D считает псевдоскалярное произведение двух плоских векторов.
func (world *World) closeDockingRequestWindowLocked(request dockingRequest) {
	world.addDockingWindowEventsLocked("dockingFinished", request.SenderCosmicObjectID, request.ReceiverCosmicObjectID, 0)
}

// cross2D считает псевдоскалярное произведение двух плоских векторов.
func (world *World) autoLandingTargetIDLocked(sender *data.CosmicObject) (int64, bool) {
	if sender == nil || sender.ClusterMainCosmicObjectID <= 0 {
		return 0, false
	}
	mainID := sender.ClusterMainCosmicObjectID
	if sender.ID != mainID {
		return mainID, true
	}
	secondaryIDs := world.secondaryClusterObjectIDsLocked(mainID)
	if len(secondaryIDs) != 1 {
		return 0, false
	}
	return secondaryIDs[0], true
}

// cross2D считает псевдоскалярное произведение двух плоских векторов.
func (world *World) requestCharacterLandingLocked(characterID int64, senderID int64, receiverID int64) error {
	character, ok := world.data.Characters.Get(characterID)
	if !ok {
		return errors.New("character not found")
	}
	sender, ok := world.data.CosmicObjects.Get(senderID)
	if !ok {
		return errors.New("sender object not found")
	}
	receiver, ok := world.data.CosmicObjects.Get(receiverID)
	if !ok {
		return errors.New("receiver object not found")
	}
	if !world.objectsInSameClusterLocked(sender, receiver) || sender.ID == receiver.ID {
		return errors.New("landing target is not in the same cluster")
	}
	if receiver.OwnerCharacterID == character.ID {
		world.moveCharacterToObjectLocked(character.ID, receiver.ID)
		return nil
	}
	if !world.cosmicObjectHasPassengerSeatLocked(receiver.ID) {
		world.addDockingNotificationLocked([]int64{sender.ID}, "В объекте назначения не установлено пассажирское кресло")
		return nil
	}
	if world.landingRequestIndexByReceiverLocked(receiver.ID) >= 0 {
		return errors.New("landing request already exists")
	}
	world.landingRequests = append(world.landingRequests, landingRequest{
		CharacterID:            character.ID,
		SenderCosmicObjectID:   sender.ID,
		ReceiverCosmicObjectID: receiver.ID,
	})
	world.addLandingWindowEventsLocked("landingRequestStarted", sender.ID, receiver.ID)
	return nil
}

// cross2D считает псевдоскалярное произведение двух плоских векторов.
func (world *World) moveCharacterToObjectLocked(characterID int64, objectID int64) {
	character, ok := world.data.Characters.Get(characterID)
	if !ok {
		return
	}
	character.LocationCosmicObjectID = objectID
	world.accountObjectIDs[character.AccountID] = objectID
}

// cross2D считает псевдоскалярное произведение двух плоских векторов.
func (world *World) objectsInSameClusterLocked(left *data.CosmicObject, right *data.CosmicObject) bool {
	return left != nil && right != nil && left.ClusterMainCosmicObjectID > 0 && left.ClusterMainCosmicObjectID == right.ClusterMainCosmicObjectID
}

// cross2D считает псевдоскалярное произведение двух плоских векторов.
func (world *World) secondaryClusterObjectIDsLocked(mainID int64) []int64 {
	objectIDs := make([]int64, 0)
	for _, cosmicObject := range world.data.CosmicObjects.Items {
		if cosmicObject != nil && cosmicObject.ClusterMainCosmicObjectID == mainID && cosmicObject.ID != mainID {
			objectIDs = append(objectIDs, cosmicObject.ID)
		}
	}
	sort.Slice(objectIDs, func(left int, right int) bool {
		return objectIDs[left] < objectIDs[right]
	})
	return objectIDs
}

// cross2D считает псевдоскалярное произведение двух плоских векторов.
func (world *World) cosmicObjectHasPassengerSeatLocked(objectID int64) bool {
	if world.data.EquipmentGroups == nil || world.data.ItemModels == nil || world.data.ItemTypes == nil {
		return false
	}
	for _, group := range world.data.EquipmentGroups.GetByCosmicObjectID(objectID) {
		if group == nil || group.Count <= 0 {
			continue
		}
		model, ok := world.data.ItemModels.Get(group.EquipmentItemModelID)
		if !ok {
			continue
		}
		itemType, ok := world.data.ItemTypes.Get(model.ItemTypeID)
		if ok && itemType.Acronym == "PassengerSeat" {
			return true
		}
	}
	return false
}

// cross2D считает псевдоскалярное произведение двух плоских векторов.
func (world *World) landingRequestIndexByReceiverLocked(receiverID int64) int {
	for index, request := range world.landingRequests {
		if request.ReceiverCosmicObjectID == receiverID {
			return index
		}
	}
	return -1
}

// cross2D считает псевдоскалярное произведение двух плоских векторов.
func (world *World) removeLandingRequestLocked(index int) {
	world.landingRequests = append(world.landingRequests[:index], world.landingRequests[index+1:]...)
}

// cross2D считает псевдоскалярное произведение двух плоских векторов.
func (world *World) addLandingWindowEventsLocked(kind string, senderID int64, receiverID int64) {
	world.dockingEvents = append(world.dockingEvents,
		game.DockingEvent{Type: "dockingEvent", Kind: kind, Role: "sender", ObjectIDs: []int64{senderID}},
		game.DockingEvent{Type: "dockingEvent", Kind: kind, Role: "receiver", ObjectIDs: []int64{receiverID}},
	)
}

// cross2D считает псевдоскалярное произведение двух плоских векторов.
func (world *World) addLandingTargetSelectionLocked(senderID int64) {
	sender, ok := world.data.CosmicObjects.Get(senderID)
	if !ok {
		return
	}
	world.dockingEvents = append(world.dockingEvents, game.DockingEvent{
		Type:      "dockingEvent",
		Kind:      "landingTargetSelection",
		ObjectIDs: []int64{senderID},
		TargetIDs: world.secondaryClusterObjectIDsLocked(sender.ClusterMainCosmicObjectID),
	})
}

// cross2D считает псевдоскалярное произведение двух плоских векторов.
func (world *World) dockingReceiverHasDecisionMakerLocked(receiverID int64) bool {
	for _, objectID := range world.accountObjectIDs {
		if objectID == receiverID {
			return true
		}
	}
	return false
}

// cross2D считает псевдоскалярное произведение двух плоских векторов.
func (world *World) addDockingWindowEventsLocked(kind string, senderID int64, receiverID int64, duration float64) {
	world.dockingEvents = append(world.dockingEvents,
		game.DockingEvent{Type: "dockingEvent", Kind: kind, Role: "sender", Duration: duration, ObjectIDs: []int64{senderID}},
		game.DockingEvent{Type: "dockingEvent", Kind: kind, Role: "receiver", Duration: duration, ObjectIDs: []int64{receiverID}},
	)
}

// cross2D считает псевдоскалярное произведение двух плоских векторов.
func (world *World) addDockingNotificationLocked(objectIDs []int64, message string) {
	seen := map[int64]bool{}
	recipients := make([]int64, 0, len(objectIDs))
	for _, objectID := range objectIDs {
		if objectID <= 0 || seen[objectID] {
			continue
		}
		seen[objectID] = true
		recipients = append(recipients, objectID)
	}
	if len(recipients) == 0 {
		return
	}
	world.dockingEvents = append(world.dockingEvents, game.DockingEvent{
		Type:      "dockingEvent",
		Kind:      "dockingNotification",
		Message:   message,
		ObjectIDs: recipients,
	})
}

// cross2D считает псевдоскалярное произведение двух плоских векторов.
func (world *World) clusterObjectIDsLocked(mainID int64) []int64 {
	objectIDs := make([]int64, 0)
	for _, cosmicObject := range world.data.CosmicObjects.Items {
		if cosmicObject != nil && cosmicObject.ClusterMainCosmicObjectID == mainID {
			objectIDs = append(objectIDs, cosmicObject.ID)
		}
	}
	sort.Slice(objectIDs, func(left int, right int) bool {
		return objectIDs[left] < objectIDs[right]
	})
	return objectIDs
}

// cross2D считает псевдоскалярное произведение двух плоских векторов.
func (world *World) disbandClusterLocked(mainID int64) {
	for _, clusterObject := range world.data.CosmicObjects.Items {
		if clusterObject != nil && clusterObject.ClusterMainCosmicObjectID == mainID {
			clusterObject.ClusterMainCosmicObjectID = 0
			clusterObject.Anchored = false
		}
	}
}

// cross2D считает псевдоскалярное произведение двух плоских векторов.
func (world *World) ownerNameForTestingLocked(ownerCharacterID int64) string {
	if ownerCharacterID <= 0 {
		return ""
	}
	character, ok := world.data.Characters.Get(ownerCharacterID)
	if !ok {
		return ""
	}
	account, ok := world.data.Accounts.Get(character.AccountID)
	if !ok {
		return ""
	}
	return account.Nickname
}

// cross2D считает псевдоскалярное произведение двух плоских векторов.
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

// cross2D считает псевдоскалярное произведение двух плоских векторов.
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
	if err := world.ensureControlledClusterEquipmentLocked(objectID, group.CosmicObjectID); err != nil {
		return err
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
	if update.Title != nil {
		group.Title = *update.Title
	}
	world.ackMutationLocked(accountID, sessionID, mutationSeq)
	return nil
}

// cross2D считает псевдоскалярное произведение двух плоских векторов.
// cross2D считает псевдоскалярное произведение двух плоских векторов.
func (world *World) ApplyControlPanelEquipmentGroupRelationUpdate(accountID int64, sessionID string, mutationSeq int64, update ControlPanelEquipmentGroupRelationUpdate) error {
	world.mu.Lock()
	defer world.mu.Unlock()

	objectID, ok := world.accountObjectIDs[accountID]
	if !ok {
		return errors.New("account is not connected")
	}
	if world.data.EquipmentGroups == nil {
		return errors.New("equipment group data is not loaded")
	}
	group, ok := world.data.EquipmentGroups.Get(update.EquipmentGroupID)
	if !ok {
		return errors.New("equipment group not found")
	}
	if err := world.ensureControlledClusterEquipmentLocked(objectID, group.CosmicObjectID); err != nil {
		return err
	}
	if _, err := world.controlledContainerEquipmentLocked(objectID, update.RelatedEquipmentGroupID); err != nil {
		return err
	}
	switch update.RelationTypeAcronym {
	case "Source":
		group.SourceEquipmentGroupID = update.RelatedEquipmentGroupID
	case "Destination":
		group.DestinationEquipmentGroupID = update.RelatedEquipmentGroupID
	case "Opposite":
		group.OppositeEquipmentGroupID = update.RelatedEquipmentGroupID
	default:
		return errors.New("unknown equipment group relation")
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
	if world.data.EquipmentGroups == nil || world.data.ItemGroups == nil || world.data.ItemModels == nil || world.data.Tasks == nil || world.data.TaskTypes == nil || world.data.TaskItemGroups == nil {
		return errors.New("cargo movement data is not loaded")
	}
	controllerID := transfer.ControllerEquipmentGroupID
	if controllerID <= 0 {
		controllerID = transfer.TargetContainerEquipmentGroupID
	}
	controller, err := world.controlledContainerEquipmentLocked(objectID, controllerID)
	if err != nil {
		return err
	}
	source, target, err := world.cargoMovementEndpointsLocked(controller, transfer.LeftToRightDirection)
	if err != nil {
		return err
	}
	if source.ID == target.ID {
		return errors.New("source and target containers must be different")
	}
	taskType, ok := world.data.TaskTypes.GetByAcronym("CargoMovement")
	if !ok {
		return errors.New("cargo movement task type not found")
	}

	movedByModel := make(map[int64]float64)
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
		if moved <= physics.Epsilon {
			continue
		}
		movedByModel[itemGroup.ContentItemModelID] += moved
	}
	if len(movedByModel) == 0 {
		return errors.New("cargo movement amount is empty")
	}
	modelIDs := make([]int64, 0, len(movedByModel))
	for itemModelID := range movedByModel {
		modelIDs = append(modelIDs, itemModelID)
	}
	sort.Slice(modelIDs, func(left int, right int) bool { return modelIDs[left] < modelIDs[right] })
	for _, itemModelID := range modelIDs {
		count := movedByModel[itemModelID]
		itemModel, ok := world.data.ItemModels.Get(itemModelID)
		if !ok {
			return errors.New("cargo item model not found")
		}
		distance, err := world.cargoMovementDistanceLocked(source.CosmicObjectID, target.CosmicObjectID)
		if err != nil {
			return err
		}
		totalEnergy := itemModel.Mass * count * distance
		if totalEnergy <= physics.Epsilon {
			totalEnergy = physics.Epsilon
		}
		task, err := world.data.Tasks.Add(&data.Task{
			ControllerEquipmentGroupID: controller.ID,
			TaskTypeID:                 taskType.ID,
			RemainingEnergy:            totalEnergy,
			TotalEnergy:                totalEnergy,
			LeftToRightDirection:       transfer.LeftToRightDirection,
			BatchCount:                 1,
		})
		if err != nil {
			return err
		}
		if _, err := world.data.TaskItemGroups.Add(&data.TaskItemGroup{TaskID: task.ID, ItemModelID: itemModelID, Count: count}); err != nil {
			return err
		}
	}
	world.ackMutationLocked(accountID, sessionID, mutationSeq)
	return nil
}

// cross2D считает псевдоскалярное произведение двух плоских векторов.
// cross2D считает псевдоскалярное произведение двух плоских векторов.
// cross2D считает псевдоскалярное произведение двух плоских векторов.
func (world *World) cargoMovementEndpointsLocked(controller *data.EquipmentGroup, leftToRight bool) (*data.EquipmentGroup, *data.EquipmentGroup, error) {
	if controller == nil {
		return nil, nil, errors.New("movement controller container not found")
	}
	opposite, ok := world.data.EquipmentGroups.Get(controller.OppositeEquipmentGroupID)
	if !ok {
		return nil, nil, errors.New("opposite container equipment group not found")
	}
	if !world.equipmentGroupIsContainerLocked(opposite) {
		return nil, nil, errors.New("opposite equipment group is not a container")
	}
	if err := world.ensureControlledClusterEquipmentLocked(controller.CosmicObjectID, opposite.CosmicObjectID); err != nil {
		return nil, nil, err
	}
	if leftToRight {
		return opposite, controller, nil
	}
	return controller, opposite, nil
}

func (world *World) ApplyControlPanelFuelTransfer(accountID int64, sessionID string, mutationSeq int64, transfer ControlPanelFuelTransfer) error {
	world.mu.Lock()
	defer world.mu.Unlock()

	objectID, ok := world.accountObjectIDs[accountID]
	if !ok {
		return errors.New("account is not connected")
	}
	if world.data.EquipmentGroups == nil || world.data.ItemGroups == nil || world.data.Tasks == nil || world.data.TaskTypes == nil || world.data.TaskItemGroups == nil || world.data.ItemModels == nil {
		return errors.New("equipment or item groups are not loaded")
	}
	container, err := world.controlledContainerEquipmentLocked(objectID, transfer.ContainerEquipmentGroupID)
	if err != nil {
		return err
	}
	fuelTankGroup, ok := world.data.EquipmentGroups.Get(transfer.FuelTankEquipmentGroupID)
	if !ok {
		return errors.New("equipment group not found")
	}
	cosmicObject, ok := world.data.CosmicObjects.Get(fuelTankGroup.CosmicObjectID)
	if !ok {
		return errors.New("fuel tank object not found")
	}
	fuelModelID, err := world.controlledFuelTankFuelModelIDLocked(objectID, transfer.FuelTankEquipmentGroupID)
	if err != nil {
		return err
	}
	taskType, ok := world.data.TaskTypes.GetByAcronym("Fueling")
	if !ok {
		return errors.New("fueling task type not found")
	}
	if len(transfer.ItemGroupIDs) > 0 {
		amount, err := world.fuelFillAmountLocked(cosmicObject, container.ID, fuelModelID, transfer.ItemGroupIDs, transfer.Amount)
		if err != nil {
			return err
		}
		if amount > physics.Epsilon {
			totalEnergy, err := world.fuelingTaskEnergyLocked(container.ID, fuelTankGroup.ID, fuelModelID, amount)
			if err != nil {
				return err
			}
			task, err := world.data.Tasks.Add(&data.Task{
				ControllerEquipmentGroupID:      fuelTankGroup.ID,
				TaskTypeID:                      taskType.ID,
				RemainingEnergy:                 totalEnergy,
				TotalEnergy:                     totalEnergy,
				BatchCount:                      1,
				LeftToRightDirection:            true,
				SourceContainerEquipmentGroupID: container.ID,
				FuelTankEquipmentGroupID:        fuelTankGroup.ID,
			})
			if err != nil {
				return err
			}
			if _, err := world.data.TaskItemGroups.Add(&data.TaskItemGroup{TaskID: task.ID, ItemModelID: fuelModelID, Count: amount}); err != nil {
				return err
			}
		}
	} else if transfer.Amount > 0 {
		amount := math.Min(cosmicObject.Fuel, transfer.Amount)
		if amount <= physics.Epsilon {
			world.ackMutationLocked(accountID, sessionID, mutationSeq)
			return nil
		}
		totalEnergy, err := world.fuelingTaskEnergyLocked(container.ID, fuelTankGroup.ID, fuelModelID, amount)
		if err != nil {
			return err
		}
		task, err := world.data.Tasks.Add(&data.Task{
			ControllerEquipmentGroupID:      fuelTankGroup.ID,
			TaskTypeID:                      taskType.ID,
			RemainingEnergy:                 totalEnergy,
			TotalEnergy:                     totalEnergy,
			BatchCount:                      1,
			LeftToRightDirection:            false,
			SourceContainerEquipmentGroupID: container.ID,
			FuelTankEquipmentGroupID:        fuelTankGroup.ID,
		})
		if err != nil {
			return err
		}
		if _, err := world.data.TaskItemGroups.Add(&data.TaskItemGroup{TaskID: task.ID, ItemModelID: fuelModelID, Count: amount}); err != nil {
			return err
		}
	}
	world.ackMutationLocked(accountID, sessionID, mutationSeq)
	return nil
}

// cross2D считает псевдоскалярное произведение двух плоских векторов.
// cross2D считает псевдоскалярное произведение двух плоских векторов.
func (world *World) ApplyControlPanelItemDeconstruction(accountID int64, sessionID string, mutationSeq int64, deconstruction ControlPanelItemDeconstruction) error {
	world.mu.Lock()
	defer world.mu.Unlock()

	objectID, ok := world.accountObjectIDs[accountID]
	if !ok {
		return errors.New("account is not connected")
	}
	if world.data.EquipmentGroups == nil || world.data.ItemGroups == nil || world.data.ItemModels == nil || world.data.Tasks == nil || world.data.TaskTypes == nil || world.data.TaskItemGroups == nil || world.data.Schemas == nil || world.data.SchemaComponents == nil {
		return errors.New("deconstruction data is not loaded")
	}
	if _, err := world.controlledEquipmentitemTypeLocked(objectID, deconstruction.DeconstructorEquipmentGroupID, "Deconstructor"); err != nil {
		return err
	}
	sourceContainer, err := world.controlledContainerEquipmentLocked(objectID, deconstruction.SourceContainerEquipmentGroupID)
	if err != nil {
		return err
	}
	targetContainer, err := world.controlledContainerEquipmentLocked(objectID, deconstruction.TargetContainerEquipmentGroupID)
	if err != nil {
		return err
	}
	taskType, ok := world.data.TaskTypes.GetByAcronym("ItemDeconstruction")
	if !ok {
		return errors.New("item deconstruction task type not found")
	}
	queued := 0
	for _, itemGroupID := range deconstruction.ItemGroupIDs {
		itemGroup, ok := world.data.ItemGroups.Get(itemGroupID)
		if !ok || itemGroup.ContainerEquipmentGroupID != sourceContainer.ID {
			continue
		}
		schema, err := world.cheapestItemSchemaByProductModelLocked(itemGroup.ContentItemModelID)
		if err != nil || schema.Count <= 0 {
			continue
		}
		selectedCount := itemGroup.Count
		if deconstruction.Amount > 0 && selectedCount > deconstruction.Amount {
			selectedCount = deconstruction.Amount
		}
		batches := math.Floor(selectedCount/schema.Count + physics.Epsilon)
		if batches <= 0 {
			continue
		}
		reservedCount := batches * schema.Count
		task, err := world.data.Tasks.Add(&data.Task{
			ControllerEquipmentGroupID:      deconstruction.DeconstructorEquipmentGroupID,
			TaskTypeID:                      taskType.ID,
			RemainingEnergy:                 schema.ProductionEnergy * batches,
			TotalEnergy:                     schema.ProductionEnergy * batches,
			BatchCount:                      int64(batches),
			SchemaID:                        schema.ID,
			SourceContainerEquipmentGroupID: sourceContainer.ID,
			TargetContainerEquipmentGroupID: targetContainer.ID,
		})
		if err != nil {
			return err
		}
		if _, err := world.data.TaskItemGroups.Add(&data.TaskItemGroup{TaskID: task.ID, ItemModelID: itemGroup.ContentItemModelID, Count: reservedCount}); err != nil {
			return err
		}
		queued++
	}
	if queued == 0 {
		return errors.New("deconstruction items not found")
	}
	world.ackMutationLocked(accountID, sessionID, mutationSeq)
	return nil
}

func (world *World) ApplyControlPanelConstructorProduceItem(accountID int64, sessionID string, mutationSeq int64, production ControlPanelConstructorProduceItem) error {
	world.mu.Lock()
	defer world.mu.Unlock()

	objectID, ok := world.accountObjectIDs[accountID]
	if !ok {
		return errors.New("account is not connected")
	}
	if world.data.EquipmentGroups == nil || world.data.ItemGroups == nil || world.data.ItemModels == nil || world.data.ItemTypes == nil || world.data.Tasks == nil || world.data.TaskTypes == nil || world.data.TaskItemGroups == nil {
		return errors.New("constructor data is not loaded")
	}
	if _, err := world.controlledEquipmentitemTypeLocked(objectID, production.ConstructorEquipmentGroupID, "Constructor"); err != nil {
		return err
	}
	materialContainer, err := world.constructorRelatedContainerOrFallbackLocked(objectID, production.ConstructorEquipmentGroupID, "Source", production.MaterialContainerEquipmentGroupID)
	if err != nil {
		return err
	}
	constructorGroup, ok := world.data.EquipmentGroups.Get(production.ConstructorEquipmentGroupID)
	if !ok {
		return errors.New("constructor equipment group not found")
	}
	constructorGroup.SourceEquipmentGroupID = materialContainer.ID
	if (production.SchemaID <= 0) == (production.BlueprintID <= 0) {
		return errors.New("constructor production must select schema or blueprint")
	}
	mainJob, components, amount, err := world.newMainConstructorProductionJobLocked(objectID, production, materialContainer.ID)
	if err != nil {
		return err
	}
	if mainJob.ProductContainerEquipmentGroupID > 0 {
		constructorGroup.DestinationEquipmentGroupID = mainJob.ProductContainerEquipmentGroupID
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
	if err := world.addConstructorTasksLocked(plannedJobs); err != nil {
		return err
	}
	world.ackMutationLocked(accountID, sessionID, mutationSeq)
	return nil
}

// cross2D считает псевдоскалярное произведение двух плоских векторов.
func (world *World) ApplyControlPanelConstructorQueueCommand(accountID int64, sessionID string, mutationSeq int64, command ControlPanelConstructorQueueCommand) error {
	world.mu.Lock()
	defer world.mu.Unlock()

	objectID, ok := world.accountObjectIDs[accountID]
	if !ok {
		return errors.New("account is not connected")
	}
	constructorController := true
	if _, err := world.controlledEquipmentitemTypeLocked(objectID, command.ConstructorEquipmentGroupID, "Constructor"); err != nil {
		constructorController = false
		if _, err := world.controlledContainerEquipmentLocked(objectID, command.ConstructorEquipmentGroupID); err != nil {
			if _, fuelTankErr := world.controlledFuelTankFuelModelIDLocked(objectID, command.ConstructorEquipmentGroupID); fuelTankErr != nil {
				if _, deconstructorErr := world.controlledEquipmentitemTypeLocked(objectID, command.ConstructorEquipmentGroupID, "Deconstructor"); deconstructorErr != nil {
					return err
				}
			}
		}
	}
	task := world.queueTaskLocked(command.ConstructorEquipmentGroupID, command.JobID)
	if constructorController {
		task = world.constructorMainTaskLocked(command.ConstructorEquipmentGroupID, command.JobID)
	}
	if task == nil {
		return errors.New("queue job not found")
	}
	switch command.Command {
	case "skipNext":
		if task.RemainingEnergy < task.TotalEnergy || world.taskHasReserveLocked(task.ID) {
			world.trimTaskToStartedCountLocked(task)
			world.removeConstructorMainTasksAfterLocked(command.ConstructorEquipmentGroupID, command.JobID)
		} else {
			world.removeConstructorTaskLocked(task.ID)
		}
	case "skipAllNext":
		if task.RemainingEnergy < task.TotalEnergy || world.taskHasReserveLocked(task.ID) {
			world.trimTaskToStartedCountLocked(task)
		}
		world.removeConstructorMainTasksAfterLocked(command.ConstructorEquipmentGroupID, command.JobID)
		world.removeConstructorTaskLocked(task.ID)
	case "cancel":
		world.removeConstructorTaskLocked(task.ID)
	case "cancelAll":
		world.removeConstructorMainTasksFromLocked(command.ConstructorEquipmentGroupID, command.JobID)
	default:
		return errors.New("constructor queue command is invalid")
	}
	world.ackMutationLocked(accountID, sessionID, mutationSeq)
	return nil
}

// cross2D считает псевдоскалярное произведение двух плоских векторов.
func (world *World) trimTaskToStartedCountLocked(task *data.Task) {
	amount := taskCount(task)
	if amount <= 1 || task.TotalEnergy <= physics.Epsilon {
		return
	}
	perUnitEnergy := task.TotalEnergy / amount
	if perUnitEnergy <= physics.Epsilon {
		return
	}
	elapsedEnergy := math.Max(0, task.TotalEnergy-task.RemainingEnergy)
	startedCount := math.Floor(elapsedEnergy/perUnitEnergy) + 1
	if startedCount < 1 {
		startedCount = 1
	}
	if startedCount >= amount {
		return
	}
	if container, err := world.taskRelatedContainerLocked(task, "Source"); err == nil {
		components, err := world.taskComponentsLocked(task)
		if err == nil {
			keptByModel := map[int64]float64{}
			for _, component := range components {
				keptByModel[component.ComponentItemModelID] += component.Count * startedCount
			}
			for _, group := range world.data.TaskItemGroups.GetByTaskID(task.ID) {
				if group == nil {
					continue
				}
				keptCount := keptByModel[group.ItemModelID]
				if group.Count > keptCount {
					_ = world.addItemModelToContainerLocked(container.ID, group.ItemModelID, group.Count-keptCount)
					group.Count = keptCount
				}
			}
			_ = world.data.ItemGroups.RebuildIndexes()
		}
	}
	task.BatchCount = int64(startedCount)
	task.TotalEnergy = perUnitEnergy * startedCount
	task.RemainingEnergy = math.Max(0, task.TotalEnergy-elapsedEnergy)
}

// cross2D считает псевдоскалярное произведение двух плоских векторов.
// cross2D считает псевдоскалярное произведение двух плоских векторов.
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

// cross2D считает псевдоскалярное произведение двух плоских векторов.
func (world *World) constructorMainTaskLocked(constructorID int64, taskID int64) *data.Task {
	if world.data.Tasks == nil || world.data.TaskTypes == nil {
		return nil
	}
	itemProductionType, _ := world.data.TaskTypes.GetByAcronym("ItemProduction")
	objectProductionType, _ := world.data.TaskTypes.GetByAcronym("ObjectProduction")
	for _, task := range world.data.Tasks.GetByControllerEquipmentGroupID(constructorID) {
		if task.ID != taskID || task.ParentTaskID != 0 {
			continue
		}
		if itemProductionType != nil && task.TaskTypeID == itemProductionType.ID {
			return task
		}
		if objectProductionType != nil && task.TaskTypeID == objectProductionType.ID {
			return task
		}
	}
	return nil
}

// cross2D считает псевдоскалярное произведение двух плоских векторов.
// cross2D считает псевдоскалярное произведение двух плоских векторов.
func (world *World) queueTaskLocked(controllerID int64, taskID int64) *data.Task {
	if world.data.Tasks == nil {
		return nil
	}
	for _, task := range world.data.Tasks.GetByControllerEquipmentGroupID(controllerID) {
		if task.ID == taskID && task.ParentTaskID == 0 {
			return task
		}
	}
	return nil
}

func (world *World) removeConstructorMainTasksAfterLocked(constructorID int64, taskID int64) {
	seenSelected := false
	for _, task := range append([]*data.Task(nil), world.data.Tasks.GetByControllerEquipmentGroupID(constructorID)...) {
		if task.ParentTaskID != 0 {
			continue
		}
		if seenSelected {
			world.removeConstructorTaskLocked(task.ID)
			continue
		}
		if task.ID == taskID {
			seenSelected = true
		}
	}
}

// cross2D считает псевдоскалярное произведение двух плоских векторов.
func (world *World) removeConstructorMainTasksFromLocked(constructorID int64, taskID int64) {
	seenSelected := false
	for _, task := range append([]*data.Task(nil), world.data.Tasks.GetByControllerEquipmentGroupID(constructorID)...) {
		if task.ParentTaskID != 0 {
			continue
		}
		if seenSelected || task.ID == taskID {
			seenSelected = true
			world.removeConstructorTaskLocked(task.ID)
		}
	}
}

// cross2D считает псевдоскалярное произведение двух плоских векторов.
func (world *World) removeConstructorTaskLocked(taskID int64) {
	world.returnTaskReserveLocked(taskID)
	world.data.TaskItemGroups.DeleteByTaskID(taskID)
	world.data.Tasks.Delete(taskID)
	dependentTasks := make([]*data.Task, 0)
	for _, task := range world.data.Tasks.Items {
		dependentTasks = append(dependentTasks, task)
	}
	for _, task := range dependentTasks {
		if task != nil && task.ParentTaskID == taskID {
			world.removeConstructorTaskLocked(task.ID)
		}
	}
}

// cross2D считает псевдоскалярное произведение двух плоских векторов.
func (world *World) returnTaskReserveLocked(taskID int64) {
	task, ok := world.data.Tasks.Get(taskID)
	if !ok {
		return
	}
	if world.taskTypeAcronymLocked(task) == "CargoMovement" {
		sourceContainerID := task.SourceContainerEquipmentGroupID
		if sourceContainerID <= 0 {
			return
		}
		for _, group := range world.data.TaskItemGroups.GetByTaskID(taskID) {
			if group == nil || !group.IsStored {
				continue
			}
			_ = world.addItemModelToContainerLocked(sourceContainerID, group.ItemModelID, group.Count)
			group.IsStored = false
		}
		_ = world.data.ItemGroups.RebuildIndexes()
		return
	}
	if world.taskTypeAcronymLocked(task) == "Fueling" {
		world.returnFuelingReserveLocked(task)
		return
	}
	if world.taskTypeAcronymLocked(task) == "ItemDeconstruction" {
		world.returnStoredTaskReserveToContainerLocked(task, task.SourceContainerEquipmentGroupID)
		return
	}
	reservedGroups := world.data.TaskItemGroups.GetByTaskID(taskID)
	if len(reservedGroups) == 0 {
		return
	}
	container, err := world.taskRelatedContainerLocked(task, "Source")
	if err != nil {
		world.moveTaskReserveThroughCargoTaskLocked(task, reservedGroups)
		return
	}
	for _, group := range world.data.TaskItemGroups.GetByTaskID(taskID) {
		_ = world.addItemModelToContainerLocked(container.ID, group.ItemModelID, group.Count)
	}
	_ = world.data.ItemGroups.RebuildIndexes()
}

// cross2D считает псевдоскалярное произведение двух плоских векторов.
// cross2D считает псевдоскалярное произведение двух плоских векторов.
func (world *World) returnStoredTaskReserveToContainerLocked(task *data.Task, containerID int64) {
	if task == nil || containerID <= 0 {
		return
	}
	for _, group := range world.data.TaskItemGroups.GetByTaskID(task.ID) {
		if group == nil || !group.IsStored {
			continue
		}
		_ = world.addItemModelToContainerLocked(containerID, group.ItemModelID, group.Count)
		group.IsStored = false
	}
	_ = world.data.ItemGroups.RebuildIndexes()
}

func (world *World) returnFuelingReserveLocked(task *data.Task) {
	if task == nil || task.SourceContainerEquipmentGroupID <= 0 || task.FuelTankEquipmentGroupID <= 0 {
		return
	}
	fuelTank, ok := world.data.EquipmentGroups.Get(task.FuelTankEquipmentGroupID)
	if !ok {
		return
	}
	cosmicObject, ok := world.data.CosmicObjects.Get(fuelTank.CosmicObjectID)
	if !ok {
		return
	}
	for _, group := range world.data.TaskItemGroups.GetByTaskID(task.ID) {
		if group == nil || !group.IsStored {
			continue
		}
		if task.LeftToRightDirection {
			_ = world.addItemModelToContainerLocked(task.SourceContainerEquipmentGroupID, group.ItemModelID, group.Count)
		} else {
			cosmicObject.Fuel = math.Min(cosmicObject.MaxFuel, cosmicObject.Fuel+group.Count)
		}
		group.IsStored = false
	}
	_ = world.data.ItemGroups.RebuildIndexes()
}

// cross2D считает псевдоскалярное произведение двух плоских векторов.
func (world *World) moveTaskReserveThroughCargoTaskLocked(task *data.Task, reservedGroups []*data.TaskItemGroup) {
	targetContainer := world.firstAvailableContainerForTaskLocked(task, reservedGroups)
	if targetContainer == nil {
		for _, group := range reservedGroups {
			_ = world.addItemModelToContainerLocked(task.ControllerEquipmentGroupID, group.ItemModelID, group.Count)
		}
		_ = world.data.ItemGroups.RebuildIndexes()
		return
	}
	cargoTaskType, ok := world.data.TaskTypes.GetByAcronym("CargoMovement")
	if !ok {
		for _, group := range reservedGroups {
			_ = world.addItemModelToContainerLocked(targetContainer.ID, group.ItemModelID, group.Count)
		}
		_ = world.data.ItemGroups.RebuildIndexes()
		return
	}
	totalEnergy := world.cargoMovementEnergyLocked(reservedGroups)
	cargoTask, err := world.data.Tasks.Add(&data.Task{
		ControllerEquipmentGroupID: targetContainer.ID,
		TaskTypeID:                 cargoTaskType.ID,
		RemainingEnergy:            totalEnergy,
		TotalEnergy:                totalEnergy,
	})
	if err != nil {
		return
	}
	for _, group := range reservedGroups {
		_, _ = world.data.TaskItemGroups.Add(&data.TaskItemGroup{TaskID: cargoTask.ID, ItemModelID: group.ItemModelID, Count: group.Count, IsStored: true})
	}
}

// cross2D считает псевдоскалярное произведение двух плоских векторов.
func (world *World) firstAvailableContainerForTaskLocked(task *data.Task, reservedGroups []*data.TaskItemGroup) *data.EquipmentGroup {
	controller, ok := world.data.EquipmentGroups.Get(task.ControllerEquipmentGroupID)
	if !ok {
		return nil
	}
	containers := make([]*data.EquipmentGroup, 0)
	for _, group := range world.data.EquipmentGroups.Items {
		if group == nil || !world.equipmentGroupIsContainerLocked(group) {
			continue
		}
		if err := world.ensureControlledClusterEquipmentLocked(controller.CosmicObjectID, group.CosmicObjectID); err != nil {
			continue
		}
		containers = append(containers, group)
	}
	sort.Slice(containers, func(left int, right int) bool { return containers[left].ID < containers[right].ID })
	for _, container := range containers {
		if world.containerCanAcceptTaskReserveLocked(container.ID, reservedGroups) {
			return container
		}
	}
	return nil
}

// cross2D считает псевдоскалярное произведение двух плоских векторов.
func (world *World) containerCanAcceptTaskReserveLocked(containerID int64, reservedGroups []*data.TaskItemGroup) bool {
	return containerID > 0 && len(reservedGroups) > 0
}

// cross2D считает псевдоскалярное произведение двух плоских векторов.
func (world *World) cargoMovementEnergyLocked(reservedGroups []*data.TaskItemGroup) float64 {
	totalMass := 0.0
	for _, group := range reservedGroups {
		if group == nil {
			continue
		}
		itemModel, ok := world.data.ItemModels.Get(group.ItemModelID)
		if !ok {
			continue
		}
		totalMass += itemModel.Mass * group.Count
	}
	if totalMass <= physics.Epsilon {
		return 1
	}
	return totalMass
}

// cross2D считает псевдоскалярное произведение двух плоских векторов.
// cross2D считает псевдоскалярное произведение двух плоских векторов.
func (world *World) cargoMovementTaskEnergyLocked(sourceContainerID int64, targetContainerID int64, groups []*data.TaskItemGroup) (float64, error) {
	source, ok := world.data.EquipmentGroups.Get(sourceContainerID)
	if !ok {
		return 0, errors.New("movement source container not found")
	}
	target, ok := world.data.EquipmentGroups.Get(targetContainerID)
	if !ok {
		return 0, errors.New("movement target container not found")
	}
	distance, err := world.cargoMovementDistanceLocked(source.CosmicObjectID, target.CosmicObjectID)
	if err != nil {
		return 0, err
	}
	totalMass := 0.0
	for _, group := range groups {
		if group == nil {
			continue
		}
		itemModel, ok := world.data.ItemModels.Get(group.ItemModelID)
		if !ok {
			return 0, errors.New("cargo item model not found")
		}
		totalMass += itemModel.Mass * group.Count
	}
	return math.Max(totalMass*distance, physics.Epsilon), nil
}

func (world *World) cargoMovementDistanceLocked(sourceObjectID int64, targetObjectID int64) (float64, error) {
	if sourceObjectID == targetObjectID {
		return world.cosmicObjectHalfSizeLocked(sourceObjectID)
	}
	source, ok := world.data.CosmicObjects.Get(sourceObjectID)
	if !ok {
		return 0, errors.New("source object not found")
	}
	target, ok := world.data.CosmicObjects.Get(targetObjectID)
	if !ok {
		return 0, errors.New("target object not found")
	}
	sourceMainID := world.clusterMainObjectIDLocked(source)
	targetMainID := world.clusterMainObjectIDLocked(target)
	if sourceMainID <= 0 || sourceMainID != targetMainID {
		return 0, errors.New("cargo movement objects are not in one cluster")
	}
	sourceHalfSize, err := world.cosmicObjectHalfSizeLocked(sourceObjectID)
	if err != nil {
		return 0, err
	}
	targetHalfSize, err := world.cosmicObjectHalfSizeLocked(targetObjectID)
	if err != nil {
		return 0, err
	}
	if sourceObjectID == sourceMainID || targetObjectID == sourceMainID {
		return sourceHalfSize + targetHalfSize, nil
	}
	mainHalfSize, err := world.cosmicObjectHalfSizeLocked(sourceMainID)
	if err != nil {
		return 0, err
	}
	return sourceHalfSize + mainHalfSize*2 + targetHalfSize, nil
}

// cross2D считает псевдоскалярное произведение двух плоских векторов.
func (world *World) clusterMainObjectIDLocked(cosmicObject *data.CosmicObject) int64 {
	if cosmicObject == nil {
		return 0
	}
	if cosmicObject.ClusterMainCosmicObjectID > 0 {
		return cosmicObject.ClusterMainCosmicObjectID
	}
	return cosmicObject.ID
}

// cross2D считает псевдоскалярное произведение двух плоских векторов.
func (world *World) cosmicObjectHalfSizeLocked(cosmicObjectID int64) (float64, error) {
	cosmicObject, ok := world.data.CosmicObjects.Get(cosmicObjectID)
	if !ok {
		return 0, errors.New("cosmic object not found")
	}
	model, ok := world.data.CosmicObjectModels.Get(cosmicObject.CosmicObjectModelID)
	if !ok {
		return 0, errors.New("cosmic object model not found")
	}
	halfSize := (model.BodyLength + model.BodyWidth) / 4
	if halfSize <= physics.Epsilon {
		return 1, nil
	}
	return halfSize, nil
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

// cross2D считает псевдоскалярное произведение двух плоских векторов.
// cross2D считает псевдоскалярное произведение двух плоских векторов.
func (world *World) addConstructorTasksLocked(plannedJobs []constructorProductionJob) error {
	taskIDByJobID := map[int64]int64{}
	orderedJobs := append([]constructorProductionJob(nil), plannedJobs...)
	sort.SliceStable(orderedJobs, func(left int, right int) bool {
		if (orderedJobs[left].ParentJobID == 0) != (orderedJobs[right].ParentJobID == 0) {
			return orderedJobs[left].ParentJobID == 0
		}
		return orderedJobs[left].ID < orderedJobs[right].ID
	})
	for _, job := range orderedJobs {
		taskTypeAcronym := "ItemProduction"
		totalEnergy := 0.0
		if job.BlueprintID > 0 {
			taskTypeAcronym = "ObjectProduction"
			blueprint, err := world.objectBlueprintLocked(job.BlueprintID)
			if err != nil {
				return err
			}
			totalEnergy = math.Max(0, blueprint.ProductionEnergy)
		} else {
			schema, err := world.itemSchemaLocked(job.SchemaID)
			if err != nil {
				return err
			}
			totalEnergy = math.Max(0, schema.ProductionEnergy)
		}
		taskType, ok := world.data.TaskTypes.GetByAcronym(taskTypeAcronym)
		if !ok {
			return errors.New("task type not found")
		}
		batches := job.TotalBatches
		if batches <= 0 {
			batches = 1
		}
		parentTaskID := taskIDByJobID[job.ParentJobID]
		task, err := world.data.Tasks.Add(&data.Task{
			ControllerEquipmentGroupID: job.ConstructorEquipmentGroupID,
			ParentTaskID:               parentTaskID,
			TaskTypeID:                 taskType.ID,
			RemainingEnergy:            totalEnergy * float64(batches),
			TotalEnergy:                totalEnergy * float64(batches),
			BatchCount:                 batches,
			SchemaID:                   job.SchemaID,
			BlueprintID:                job.BlueprintID,
		})
		if err != nil {
			return err
		}
		taskIDByJobID[job.ID] = task.ID
	}
	return nil
}
func (world *World) newConstructorProductionJobLocked(constructorID int64, materialContainerID int64, productContainerID int64, queueType string, schema controlPanelItemSchema, batches int64, parentJobID int64) constructorProductionJob {
	world.nextConstructorProductionJobID++
	totalTime := math.Max(0, schema.ProductionEnergy)
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

// cross2D считает псевдоскалярное произведение двух плоских векторов.
func (world *World) newConstructorObjectProductionJobLocked(constructorID int64, materialContainerID int64, queueType string, blueprint controlPanelObjectBlueprint, batches int64, parentJobID int64) constructorProductionJob {
	world.nextConstructorProductionJobID++
	totalTime := math.Max(0, blueprint.ProductionEnergy)
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
	if err := world.ensureControlledClusterEquipmentLocked(objectID, group.CosmicObjectID); err != nil {
		return nil, err
	}
	if !world.equipmentGroupIsContainerLocked(group) {
		return nil, errors.New("equipment group is not a container")
	}
	return group, nil
}

// cross2D считает псевдоскалярное произведение двух плоских векторов.
func (world *World) ensureControlledClusterEquipmentLocked(controlledObjectID int64, equipmentObjectID int64) error {
	if equipmentObjectID == controlledObjectID {
		return nil
	}
	controlledObject, ok := world.data.CosmicObjects.Get(controlledObjectID)
	if !ok {
		return errors.New("controlled object not found")
	}
	equipmentObject, ok := world.data.CosmicObjects.Get(equipmentObjectID)
	if !ok {
		return errors.New("equipment object not found")
	}
	mainID := controlledObject.ClusterMainCosmicObjectID
	if mainID <= 0 || equipmentObject.ClusterMainCosmicObjectID != mainID {
		return errors.New("equipment group does not belong to controlled object")
	}
	mainObject, ok := world.data.CosmicObjects.Get(mainID)
	if !ok || mainObject.OwnerCharacterID != controlledObject.OwnerCharacterID || controlledObject.OwnerCharacterID <= 0 {
		return errors.New("equipment group does not belong to controlled object")
	}
	if equipmentObject.OwnerCharacterID != controlledObject.OwnerCharacterID {
		return errors.New("equipment object does not belong to character")
	}
	return nil
}

// cross2D считает псевдоскалярное произведение двух плоских векторов.
// cross2D считает псевдоскалярное произведение двух плоских векторов.
func (world *World) relatedContainerEquipmentLocked(objectID int64, equipmentGroupID int64, relationTypeAcronym string) (*data.EquipmentGroup, error) {
	group, ok := world.data.EquipmentGroups.Get(equipmentGroupID)
	if !ok {
		return nil, errors.New("equipment group not found")
	}
	var relatedEquipmentGroupID int64
	switch relationTypeAcronym {
	case "Source":
		relatedEquipmentGroupID = group.SourceEquipmentGroupID
	case "Destination":
		relatedEquipmentGroupID = group.DestinationEquipmentGroupID
	case "Opposite":
		relatedEquipmentGroupID = group.OppositeEquipmentGroupID
	default:
		return nil, errors.New("unknown equipment group relation")
	}
	if relatedEquipmentGroupID <= 0 {
		return nil, errors.New("equipment group relation not found")
	}
	return world.controlledContainerEquipmentLocked(objectID, relatedEquipmentGroupID)
}

// cross2D считает псевдоскалярное произведение двух плоских векторов.
func (world *World) constructorRelatedContainerOrFallbackLocked(objectID int64, constructorID int64, relationTypeAcronym string, fallbackContainerID int64) (*data.EquipmentGroup, error) {
	if fallbackContainerID > 0 {
		return world.controlledContainerEquipmentLocked(objectID, fallbackContainerID)
	}
	return world.relatedContainerEquipmentLocked(objectID, constructorID, relationTypeAcronym)
}

func (world *World) controlledEquipmentitemTypeLocked(objectID int64, groupID int64, itemTypeAcronym string) (*data.EquipmentGroup, error) {
	group, ok := world.data.EquipmentGroups.Get(groupID)
	if !ok {
		return nil, errors.New("equipment group not found")
	}
	if err := world.ensureControlledClusterEquipmentLocked(objectID, group.CosmicObjectID); err != nil {
		return nil, err
	}
	model, ok := world.data.ItemModels.Get(group.EquipmentItemModelID)
	if !ok {
		return nil, errors.New("equipment model not found")
	}
	itemType, ok := world.data.ItemTypes.Get(model.ItemTypeID)
	if !ok || itemType.Acronym != itemTypeAcronym {
		return nil, errors.New("equipment group has unexpected item type")
	}
	return group, nil
}

// cross2D считает псевдоскалярное произведение двух плоских векторов.
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

// cross2D считает псевдоскалярное произведение двух плоских векторов.
// cross2D считает псевдоскалярное произведение двух плоских векторов.
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

// cross2D считает псевдоскалярное произведение двух плоских векторов.
func (world *World) cheapestItemSchemaByProductModelLocked(itemModelID int64) (controlPanelItemSchema, error) {
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
	var best controlPanelItemSchema
	found := false
	for _, schemaID := range schemaIDs {
		schema, err := world.itemSchemaLocked(schemaID)
		if err != nil {
			return controlPanelItemSchema{}, err
		}
		if schema.ItemModelID != itemModelID {
			continue
		}
		if !found || schema.ProductionEnergy < best.ProductionEnergy {
			best = schema
			found = true
		}
	}
	if !found {
		return controlPanelItemSchema{}, errors.New("item schema not found")
	}
	return best, nil
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

// cross2D считает псевдоскалярное произведение двух плоских векторов.
// cross2D считает псевдоскалярное произведение двух плоских векторов.
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

// cross2D считает псевдоскалярное произведение двух плоских векторов.
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
			world.deleteItemGroupLocked(itemGroup)
		}
	}
}

// cross2D считает псевдоскалярное произведение двух плоских векторов.
// cross2D считает псевдоскалярное произведение двух плоских векторов.
func (world *World) deleteItemGroupLocked(itemGroup *data.ItemGroup) {
	if itemGroup == nil || world.data.ItemGroups == nil {
		return
	}
	delete(world.data.ItemGroups.Items, itemGroup.ID)
	groups := world.data.ItemGroups.ByContainerEquipmentGroupID[itemGroup.ContainerEquipmentGroupID]
	for index, indexedGroup := range groups {
		if indexedGroup == nil || indexedGroup.ID != itemGroup.ID {
			continue
		}
		groups = append(groups[:index], groups[index+1:]...)
		break
	}
	if len(groups) == 0 {
		delete(world.data.ItemGroups.ByContainerEquipmentGroupID, itemGroup.ContainerEquipmentGroupID)
		return
	}
	world.data.ItemGroups.ByContainerEquipmentGroupID[itemGroup.ContainerEquipmentGroupID] = groups
}

// cross2D считает псевдоскалярное произведение двух плоских векторов.
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

// cross2D считает псевдоскалярное произведение двух плоских векторов.
// cross2D считает псевдоскалярное произведение двух плоских векторов.
func (world *World) setItemModelInContainerLocked(containerID int64, itemModelID int64, amount float64) error {
	for _, itemGroup := range world.data.ItemGroups.GetByContainerEquipmentGroupID(containerID) {
		if itemGroup.ContentItemModelID == itemModelID {
			itemGroup.Count = amount
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

func (world *World) controlledFuelTankFuelModelIDLocked(objectID int64, groupID int64) (int64, error) {
	group, ok := world.data.EquipmentGroups.Get(groupID)
	if !ok {
		return 0, errors.New("equipment group not found")
	}
	if err := world.ensureControlledClusterEquipmentLocked(objectID, group.CosmicObjectID); err != nil {
		return 0, err
	}
	model, ok := world.data.ItemModels.Get(group.EquipmentItemModelID)
	if !ok {
		return 0, errors.New("fuel tank model not found")
	}
	itemType, ok := world.data.ItemTypes.Get(model.ItemTypeID)
	if !ok || itemType.Acronym != "FuelTank" {
		return 0, errors.New("equipment group is not a fuel tank")
	}
	if model.ConsumingItemModelID <= 0 {
		return 0, errors.New("fuel tank fuel model is not set")
	}
	return model.ConsumingItemModelID, nil
}

// cross2D считает псевдоскалярное произведение двух плоских векторов.
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
			world.deleteItemGroupLocked(itemGroup)
		}
	}
	return nil
}

// cross2D считает псевдоскалярное произведение двух плоских векторов.
func (world *World) fuelFillAmountLocked(cosmicObject *data.CosmicObject, containerID int64, fuelModelID int64, itemGroupIDs []int64, amount float64) (float64, error) {
	freeFuel := math.Max(0, cosmicObject.MaxFuel-cosmicObject.Fuel)
	if freeFuel <= 0 {
		return 0, nil
	}
	selectedFuel := 0.0
	for _, itemGroupID := range itemGroupIDs {
		itemGroup, ok := world.data.ItemGroups.Get(itemGroupID)
		if !ok {
			return 0, errors.New("item group not found")
		}
		if itemGroup.ContainerEquipmentGroupID != containerID {
			return 0, errors.New("item group does not belong to source container")
		}
		if itemGroup.ContentItemModelID != fuelModelID {
			return 0, errors.New("item group is not fuel for selected tank")
		}
		selectedFuel += itemGroup.Count
	}
	if amount > 0 {
		selectedFuel = math.Min(selectedFuel, amount)
	}
	return math.Min(freeFuel, selectedFuel), nil
}

// cross2D считает псевдоскалярное произведение двух плоских векторов.
func (world *World) fuelingTaskEnergyLocked(containerID int64, fuelTankID int64, fuelModelID int64, amount float64) (float64, error) {
	container, ok := world.data.EquipmentGroups.Get(containerID)
	if !ok {
		return 0, errors.New("fueling container not found")
	}
	fuelTank, ok := world.data.EquipmentGroups.Get(fuelTankID)
	if !ok {
		return 0, errors.New("fueling tank not found")
	}
	itemModel, ok := world.data.ItemModels.Get(fuelModelID)
	if !ok {
		return 0, errors.New("fuel item model not found")
	}
	distance, err := world.cargoMovementDistanceLocked(container.CosmicObjectID, fuelTank.CosmicObjectID)
	if err != nil {
		return 0, err
	}
	totalEnergy := itemModel.Mass * amount * distance
	if totalEnergy <= physics.Epsilon {
		totalEnergy = physics.Epsilon
	}
	return totalEnergy, nil
}

// cross2D считает псевдоскалярное произведение двух плоских векторов.
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

// cross2D считает псевдоскалярное произведение двух плоских векторов.
func (world *World) equipmentGroupIsContainerLocked(group *data.EquipmentGroup) bool {
	if group == nil || world.data.ItemModels == nil || world.data.ItemTypes == nil {
		return false
	}
	model, ok := world.data.ItemModels.Get(group.EquipmentItemModelID)
	if !ok {
		return false
	}
	itemType, ok := world.data.ItemTypes.Get(model.ItemTypeID)
	return ok && itemType.Acronym == "Container"
}

// cross2D считает псевдоскалярное произведение двух плоских векторов.
// cross2D считает псевдоскалярное произведение двух плоских векторов.
func (world *World) stepMiningLocked(dtSeconds float64, inputsByObjectID map[int64]game.ShipInput) {
	if dtSeconds <= 0 || world.data.CosmicObjects == nil || world.data.CosmicObjectModels == nil || world.data.EquipmentGroups == nil || world.data.ItemGroups == nil || world.data.ItemModels == nil || world.data.ItemTypes == nil {
		return
	}

	objectIDs := make([]int64, 0, len(inputsByObjectID))
	for objectID := range inputsByObjectID {
		objectIDs = append(objectIDs, objectID)
	}
	sort.Slice(objectIDs, func(left int, right int) bool {
		return objectIDs[left] < objectIDs[right]
	})

	for _, objectID := range objectIDs {
		input := inputsByObjectID[objectID]
		cosmicObject, ok := world.data.CosmicObjects.Get(objectID)
		if !ok || !cosmicObject.Enabled {
			continue
		}
		parameters, ok := world.selectedSimpleDrillMiningParametersLocked(objectID, input)
		if !ok {
			continue
		}
		electricShare := world.electricShareForInput(*cosmicObject, input)
		if electricShare <= physics.Epsilon {
			continue
		}
		targetContainer, ok := world.firstContainerEquipmentGroupLocked(objectID)
		if !ok {
			continue
		}
		hitObject, hitModel, ok := world.nearestRayHitObjectLocked(*cosmicObject, parameters.Range)
		if !ok || !world.cosmicObjectModelHasTypeAcronymLocked(hitModel, "Asteroid") {
			continue
		}
		sourceContainer, ok := world.firstContainerEquipmentGroupLocked(hitObject.ID)
		if !ok {
			continue
		}
		sourceItem, resourceModel, ok := world.firstResourceItemGroupLocked(sourceContainer.ID)
		if !ok || resourceModel.Mass <= physics.Epsilon {
			continue
		}

		maxUnits := parameters.MiningSpeed * float64(parameters.EnabledCount) * electricShare * dtSeconds / resourceModel.Mass
		transferUnits := math.Min(sourceItem.Count, maxUnits)
		if transferUnits <= physics.Epsilon {
			continue
		}
		if err := world.addItemModelToContainerLocked(targetContainer.ID, sourceItem.ContentItemModelID, transferUnits); err != nil {
			continue
		}
		world.consumeItemModelFromContainerLocked(sourceContainer.ID, sourceItem.ContentItemModelID, transferUnits)
		world.recordMiningNotificationLocked(objectID, resourceModel, transferUnits, dtSeconds)
	}
}

// cross2D считает псевдоскалярное произведение двух плоских векторов.
func (world *World) stepWeaponFireLocked(dtSeconds float64, inputsByObjectID map[int64]game.ShipInput) {
	if dtSeconds <= 0 || world.data.CosmicObjects == nil || world.data.CosmicObjectModels == nil || world.data.CosmicObjectTypes == nil || world.data.EquipmentGroups == nil || world.data.ItemModels == nil || world.data.ItemTypes == nil {
		return
	}
	if world.weaponShotAccumulators == nil {
		world.weaponShotAccumulators = map[weaponShotKey]float64{}
	}

	objectIDs := make([]int64, 0, len(inputsByObjectID))
	for objectID := range inputsByObjectID {
		objectIDs = append(objectIDs, objectID)
	}
	sort.Slice(objectIDs, func(left int, right int) bool {
		return objectIDs[left] < objectIDs[right]
	})

	for _, objectID := range objectIDs {
		input := inputsByObjectID[objectID]
		cosmicObject, ok := world.data.CosmicObjects.Get(objectID)
		if !ok || !cosmicObject.Enabled {
			continue
		}
		// Выбранное оружие нужно найти и тогда, когда действие отпущено.
		firingInput := input
		firingInput.PrimaryPointerAction = true
		parameters, ok := world.selectedWeaponAttackParametersLocked(objectID, firingInput)
		if !ok {
			continue
		}
		for _, attackGroup := range parameters.Groups {
			if attackGroup.EquipmentGroup == nil {
				continue
			}
			groupParameters := parameters
			groupParameters.ShotsPerSecond = attackGroup.ShotsPerSecond
			groupParameters.InitialProjectileCount = attackGroup.InitialProjectileCount
			key := weaponShotKey{ObjectID: objectID, EquipmentGroupID: attackGroup.EquipmentGroup.ID}
			if _, ok := world.weaponShotAccumulators[key]; !ok {
				world.weaponShotAccumulators[key] = 1
			}
			if !input.PrimaryPointerAction {
				// Пауза восстанавливает готовность одного залпа без накопления очереди выстрелов.
				world.weaponShotAccumulators[key] += groupParameters.ShotsPerSecond * dtSeconds
				if world.weaponShotAccumulators[key] > 1 {
					world.weaponShotAccumulators[key] = 1
				}
				continue
			}
			electricShare := world.electricShareForInput(*cosmicObject, input)
			if electricShare <= physics.Epsilon {
				continue
			}
			world.weaponShotAccumulators[key] += groupParameters.ShotsPerSecond * electricShare * dtSeconds
			volleyCount := int(math.Floor(world.weaponShotAccumulators[key]))
			if volleyCount <= 0 {
				continue
			}
			shotCount := volleyCount * int(groupParameters.InitialProjectileCount)
			firedCount := 0
			for firedCount < shotCount && world.consumeWeaponMagazineShotLocked(attackGroup.EquipmentGroup, groupParameters.ItemModelID) {
				firedCount++
			}
			world.weaponShotAccumulators[key] -= float64(volleyCount)
			if firedCount <= 0 {
				continue
			}
			world.spawnWeaponProjectilesLocked(*cosmicObject, groupParameters, firedCount, attackGroup.BarrelStartIndex, attackGroup.EnabledCount, parameters.InitialProjectileCount)
		}
	}
}

// cross2D считает псевдоскалярное произведение двух плоских векторов.
// cross2D считает псевдоскалярное произведение двух плоских векторов.
func (world *World) stepWeaponReloadsLocked() {
	if world.data.EquipmentGroups == nil || world.data.ItemModels == nil {
		return
	}
	groups := make([]*data.EquipmentGroup, 0, len(world.data.EquipmentGroups.Items))
	for _, group := range world.data.EquipmentGroups.Items {
		groups = append(groups, group)
	}
	sort.Slice(groups, func(left int, right int) bool {
		return groups[left].ID < groups[right].ID
	})
	for _, group := range groups {
		if group == nil || group.LastRechargeStartTime <= 0 {
			continue
		}
		model, ok := world.data.ItemModels.Get(group.EquipmentItemModelID)
		if !ok {
			continue
		}
		world.finishWeaponReloadLocked(group, model)
	}
}

// cross2D считает псевдоскалярное произведение двух плоских векторов.
func (world *World) consumeWeaponMagazineShotLocked(group *data.EquipmentGroup, itemModelID int64) bool {
	if group == nil || world.data.ItemModels == nil {
		return true
	}
	model, ok := world.data.ItemModels.Get(itemModelID)
	if !ok || weaponMagazineCapacity(group, model) <= 0 || model.AmmoItemModelID <= 0 {
		return true
	}
	if group.MagazineCount <= 0 {
		world.startWeaponReloadLocked(group, model)
		return false
	}
	group.MagazineCount--
	if group.MagazineCount <= 0 {
		world.startWeaponReloadLocked(group, model)
	}
	return true
}

// cross2D считает псевдоскалярное произведение двух плоских векторов.
func (world *World) startWeaponReloadLocked(group *data.EquipmentGroup, model *data.ItemModel) {
	if group == nil || model == nil || weaponMagazineCapacity(group, model) <= 0 || model.AmmoItemModelID <= 0 || group.LastRechargeStartTime > 0 {
		return
	}
	if model.RechargeTime <= 0 {
		world.reloadWeaponMagazineFromContainersLocked(group, model)
		return
	}
	group.LastRechargeStartTime = world.currentTimeMillis
}

// cross2D считает псевдоскалярное произведение двух плоских векторов.
func (world *World) finishWeaponReloadLocked(group *data.EquipmentGroup, model *data.ItemModel) {
	if group == nil || model == nil || group.LastRechargeStartTime <= 0 {
		return
	}
	rechargeMilliseconds := int64(math.Ceil(model.RechargeTime * 1000))
	if rechargeMilliseconds > 0 && world.currentTimeMillis-group.LastRechargeStartTime < rechargeMilliseconds {
		return
	}
	if world.reloadWeaponMagazineFromContainersLocked(group, model) > 0 {
		group.LastRechargeStartTime = 0
	}
}

// cross2D считает псевдоскалярное произведение двух плоских векторов.
func (world *World) reloadWeaponMagazineFromContainersLocked(group *data.EquipmentGroup, model *data.ItemModel) int64 {
	if group == nil || model == nil || world.data.EquipmentGroups == nil || world.data.ItemGroups == nil || model.AmmoItemModelID <= 0 {
		return 0
	}
	capacity := weaponMagazineCapacity(group, model)
	need := capacity - group.MagazineCount
	if need <= 0 {
		return 0
	}
	containerIDs := world.containerEquipmentGroupIDsLocked(group.CosmicObjectID)
	remaining := float64(need)
	loaded := float64(0)
	for _, containerID := range containerIDs {
		for _, itemGroup := range append([]*data.ItemGroup(nil), world.data.ItemGroups.GetByContainerEquipmentGroupID(containerID)...) {
			if remaining <= physics.Epsilon {
				break
			}
			if itemGroup.ContentItemModelID != model.AmmoItemModelID {
				continue
			}
			consumed := math.Min(itemGroup.Count, remaining)
			itemGroup.Count -= consumed
			remaining -= consumed
			loaded += consumed
			if itemGroup.Count <= physics.Epsilon {
				world.deleteItemGroupLocked(itemGroup)
			}
		}
	}
	loadedCount := int64(math.Floor(loaded + physics.Epsilon))
	group.MagazineCount += loadedCount
	if group.MagazineCount > capacity {
		group.MagazineCount = capacity
	}
	return loadedCount
}

// cross2D считает псевдоскалярное произведение двух плоских векторов.
func (world *World) containerEquipmentGroupIDsLocked(objectID int64) []int64 {
	if world.data.EquipmentGroups == nil || world.data.ItemModels == nil || world.data.ItemTypes == nil {
		return nil
	}
	containerIDs := make([]int64, 0)
	for _, group := range world.data.EquipmentGroups.GetByCosmicObjectID(objectID) {
		model, ok := world.data.ItemModels.Get(group.EquipmentItemModelID)
		if !ok || !world.itemModelHasTypeAcronymLocked(model, "Container") {
			continue
		}
		containerIDs = append(containerIDs, group.ID)
	}
	sort.Slice(containerIDs, func(left int, right int) bool {
		return containerIDs[left] < containerIDs[right]
	})
	return containerIDs
}

// cross2D считает псевдоскалярное произведение двух плоских векторов.
func weaponMagazineCapacity(group *data.EquipmentGroup, model *data.ItemModel) int64 {
	if group == nil || model == nil || group.Count <= 0 || model.MagazineCapacity <= 0 {
		return 0
	}
	return group.Count * model.MagazineCapacity
}

func (world *World) spawnWeaponProjectilesLocked(source data.CosmicObject, parameters weaponAttackParameters, shotCount int, barrelStartIndex int64, groupBarrelCount int64, totalBarrelCount int64) {
	projectileModel, ok := world.data.CosmicObjectModels.Get(parameters.ProjectileModelID)
	if !ok || shotCount <= 0 || groupBarrelCount <= 0 || totalBarrelCount <= 0 {
		return
	}
	sourceModel, ok := world.data.CosmicObjectModels.Get(source.CosmicObjectModelID)
	if !ok {
		return
	}
	forward := physics.ForwardVector(source.Rotation)
	right := physics.RightVector(source.Rotation)
	startDistance := modelVisualForwardOffsetMeters(*sourceModel)
	for index := 0; index < shotCount; index++ {
		barrelIndex := barrelStartIndex + int64(index)%groupBarrelCount
		sideOffset := weaponBarrelLateralOffsetMeters(*sourceModel, barrelIndex, totalBarrelCount)
		projectileID := world.nextProjectileID
		world.nextProjectileID--
		velocityX := source.VelocityX + forward.X*parameters.ProjectileSpeed
		velocityY := source.VelocityY + forward.Y*parameters.ProjectileSpeed
		speed := math.Hypot(velocityX, velocityY)
		cosmicObject := data.CosmicObject{
			ID:                  projectileID,
			Title:               projectileModel.Acronym,
			CosmicObjectModelID: projectileModel.ID,
			X:                   source.X + forward.X*startDistance + right.X*sideOffset,
			Y:                   source.Y + forward.Y*startDistance + right.Y*sideOffset,
			Rotation:            source.Rotation,
			TargetRotation:      source.Rotation,
			Speed:               speed,
			VelocityX:           velocityX,
			VelocityY:           velocityY,
			Enabled:             true,
		}
		world.projectiles = append(world.projectiles, activeProjectile{
			ID:              projectileID,
			SourceObjectID:  source.ID,
			CosmicObject:    cosmicObject,
			Damage:          parameters.Damage,
			VelocityX:       velocityX,
			VelocityY:       velocityY,
			ProjectileSpeed: parameters.ProjectileSpeed,
			RemainingRange:  parameters.Range,
		})
	}
}

// cross2D считает псевдоскалярное произведение двух плоских векторов.
func (world *World) stepProjectilesLocked(dtSeconds float64) {
	if dtSeconds <= 0 || len(world.projectiles) == 0 || world.data.CosmicObjects == nil || world.data.CosmicObjectModels == nil || world.data.CosmicObjectTypes == nil {
		return
	}

	remaining := world.projectiles[:0]
	for _, projectile := range world.projectiles {
		speed := math.Hypot(projectile.VelocityX, projectile.VelocityY)
		if projectile.ProjectileSpeed <= physics.Epsilon || projectile.RemainingRange <= physics.Epsilon {
			continue
		}
		moveDistance := math.Min(projectile.ProjectileSpeed*dtSeconds, projectile.RemainingRange)
		moveSeconds := moveDistance / projectile.ProjectileSpeed
		startX := projectile.CosmicObject.X
		startY := projectile.CosmicObject.Y
		endX := startX + projectile.VelocityX*moveSeconds
		endY := startY + projectile.VelocityY*moveSeconds

		hitObject, hitModel, ok := world.nearestProjectileHitObjectLocked(projectile, startX, startY, endX, endY)
		if ok {
			if world.cosmicObjectModelCanReceiveWeaponDamageLocked(hitModel) {
				world.damageObjectArmorLocked(hitObject, projectile.Damage)
			}
			continue
		}

		projectile.CosmicObject.X = endX
		projectile.CosmicObject.Y = endY
		projectile.CosmicObject.Speed = speed
		projectile.CosmicObject.VelocityX = projectile.VelocityX
		projectile.CosmicObject.VelocityY = projectile.VelocityY
		projectile.RemainingRange -= moveDistance
		if projectile.RemainingRange > physics.Epsilon {
			remaining = append(remaining, projectile)
		}
	}
	world.projectiles = remaining
}

// cross2D считает псевдоскалярное произведение двух плоских векторов.
func (world *World) nearestProjectileHitObjectLocked(projectile activeProjectile, startX float64, startY float64, endX float64, endY float64) (*data.CosmicObject, *data.CosmicObjectModel, bool) {
	bestDistance := math.Inf(1)
	var bestObject *data.CosmicObject
	var bestModel *data.CosmicObjectModel
	objectIDs := make([]int64, 0, len(world.data.CosmicObjects.Items))
	for objectID := range world.data.CosmicObjects.Items {
		objectIDs = append(objectIDs, objectID)
	}
	sort.Slice(objectIDs, func(left int, right int) bool {
		return objectIDs[left] < objectIDs[right]
	})
	for _, objectID := range objectIDs {
		if objectID == projectile.SourceObjectID {
			continue
		}
		candidate, ok := world.data.CosmicObjects.Get(objectID)
		if !ok || !candidate.Enabled {
			continue
		}
		candidateModel, ok := world.data.CosmicObjectModels.Get(candidate.CosmicObjectModelID)
		if !ok || world.cosmicObjectModelIgnoresProjectileHitLocked(candidateModel) {
			continue
		}
		distance, ok := raySegmentPolygonDistance(startX, startY, endX, endY, *candidate, *candidateModel)
		if ok && distance < bestDistance {
			bestDistance = distance
			bestObject = candidate
			bestModel = candidateModel
		}
	}
	if bestObject == nil || bestModel == nil {
		return nil, nil, false
	}
	return bestObject, bestModel, true
}

// cross2D считает псевдоскалярное произведение двух плоских векторов.
func (world *World) cosmicObjectModelIgnoresProjectileHitLocked(model *data.CosmicObjectModel) bool {
	return world.cosmicObjectModelHasProjectileTypeLocked(model)
}

// cross2D считает псевдоскалярное произведение двух плоских векторов.
func (world *World) cosmicObjectModelCanReceiveWeaponDamageLocked(model *data.CosmicObjectModel) bool {
	return world.cosmicObjectModelHasTypeAcronymLocked(model, "Ship") || world.cosmicObjectModelHasTypeAcronymLocked(model, "Station")
}

// cross2D считает псевдоскалярное произведение двух плоских векторов.
func (world *World) damageObjectArmorLocked(cosmicObject *data.CosmicObject, damage float64) {
	if cosmicObject == nil || damage <= physics.Epsilon || cosmicObject.Armor <= 0 {
		return
	}
	cosmicObject.Armor = math.Max(0, cosmicObject.Armor-damage)
	cosmicObject.LastReceivedDamageTime = time.Now().UnixMilli()
}

// cross2D считает псевдоскалярное произведение двух плоских векторов.
func (world *World) recordMiningNotificationLocked(objectID int64, resourceModel *data.ItemModel, minedCount float64, dtSeconds float64) {
	if resourceModel == nil || minedCount <= physics.Epsilon || dtSeconds <= physics.Epsilon {
		return
	}
	if world.miningNotifications == nil {
		world.miningNotifications = map[miningNotificationKey]miningNotificationAccumulator{}
	}

	key := miningNotificationKey{ObjectID: objectID, ItemModelID: resourceModel.ID}
	accumulator := world.miningNotifications[key]
	accumulator.Seconds += dtSeconds
	accumulator.Count += minedCount

	for accumulator.Seconds+physics.Epsilon >= miningNotificationSeconds {
		intervalShare := miningNotificationSeconds / accumulator.Seconds
		intervalCount := accumulator.Count * intervalShare
		world.addExchangeNotificationLocked(
			[]int64{objectID},
			fmt.Sprintf("+ %.0f %s", intervalCount, resourceModel.TitleRu),
		)
		accumulator.Seconds -= miningNotificationSeconds
		accumulator.Count -= intervalCount
	}

	if accumulator.Seconds <= physics.Epsilon || accumulator.Count <= physics.Epsilon {
		delete(world.miningNotifications, key)
		return
	}
	world.miningNotifications[key] = accumulator
}

// cross2D считает псевдоскалярное произведение двух плоских векторов.
func (world *World) firstContainerEquipmentGroupLocked(objectID int64) (*data.EquipmentGroup, bool) {
	for _, group := range sortedEquipmentGroups(world.data.EquipmentGroups.GetByCosmicObjectID(objectID)) {
		if group != nil && group.Count > 0 && world.equipmentGroupIsContainerLocked(group) {
			return group, true
		}
	}
	return nil, false
}

// cross2D считает псевдоскалярное произведение двух плоских векторов.
func (world *World) firstResourceItemGroupLocked(containerID int64) (*data.ItemGroup, *data.ItemModel, bool) {
	groups := append([]*data.ItemGroup(nil), world.data.ItemGroups.GetByContainerEquipmentGroupID(containerID)...)
	sort.Slice(groups, func(left int, right int) bool {
		return groups[left].ID < groups[right].ID
	})
	for _, group := range groups {
		if group == nil || group.Count <= physics.Epsilon {
			continue
		}
		model, ok := world.data.ItemModels.Get(group.ContentItemModelID)
		if !ok || !world.itemModelHasTypeAcronymLocked(model, "Resource") {
			continue
		}
		return group, model, true
	}
	return nil, nil, false
}

// cross2D считает псевдоскалярное произведение двух плоских векторов.
func (world *World) nearestRayHitObjectLocked(source data.CosmicObject, rayRange float64) (*data.CosmicObject, *data.CosmicObjectModel, bool) {
	sourceModel, ok := world.data.CosmicObjectModels.Get(source.CosmicObjectModelID)
	if !ok || rayRange <= 0 {
		return nil, nil, false
	}
	forward := physics.ForwardVector(source.Rotation)
	startDistance := modelVisualForwardOffsetMeters(*sourceModel)
	startX := source.X + forward.X*startDistance
	startY := source.Y + forward.Y*startDistance
	endX := startX + forward.X*rayRange
	endY := startY + forward.Y*rayRange

	bestDistance := math.Inf(1)
	var bestObject *data.CosmicObject
	var bestModel *data.CosmicObjectModel
	objectIDs := make([]int64, 0, len(world.data.CosmicObjects.Items))
	for objectID := range world.data.CosmicObjects.Items {
		objectIDs = append(objectIDs, objectID)
	}
	sort.Slice(objectIDs, func(left int, right int) bool {
		return objectIDs[left] < objectIDs[right]
	})
	for _, objectID := range objectIDs {
		if objectID == source.ID {
			continue
		}
		candidate, ok := world.data.CosmicObjects.Get(objectID)
		if !ok || !candidate.Enabled {
			continue
		}
		candidateModel, ok := world.data.CosmicObjectModels.Get(candidate.CosmicObjectModelID)
		if !ok {
			continue
		}
		distance, ok := raySegmentPolygonDistance(startX, startY, endX, endY, *candidate, *candidateModel)
		if ok && distance < bestDistance {
			bestDistance = distance
			bestObject = candidate
			bestModel = candidateModel
		}
	}
	if bestObject == nil || bestModel == nil {
		return nil, nil, false
	}
	return bestObject, bestModel, true
}

// cross2D считает псевдоскалярное произведение двух плоских векторов.
func (world *World) cosmicObjectModelHasTypeAcronymLocked(model *data.CosmicObjectModel, acronym string) bool {
	if model == nil {
		return false
	}
	cosmicObjectType, ok := world.data.CosmicObjectTypes.Get(model.CosmicObjectTypeID)
	return ok && cosmicObjectType.Acronym == acronym
}

func (world *World) cosmicObjectModelHasProjectileTypeLocked(model *data.CosmicObjectModel) bool {
	if model == nil {
		return false
	}
	cosmicObjectType, ok := world.data.CosmicObjectTypes.Get(model.CosmicObjectTypeID)
	return ok && cosmicObjectType.IsProjectile
}

// cross2D считает псевдоскалярное произведение двух плоских векторов.
func (world *World) itemModelHasTypeAcronymLocked(model *data.ItemModel, acronym string) bool {
	if model == nil {
		return false
	}
	itemType, ok := world.data.ItemTypes.Get(model.ItemTypeID)
	return ok && itemType.Acronym == acronym
}

// cross2D считает псевдоскалярное произведение двух плоских векторов.
func (world *World) projectileModelForWeaponLocked(model *data.ItemModel) (*data.CosmicObjectModel, bool) {
	if model == nil || world.data.CosmicObjectModels == nil {
		return nil, false
	}
	return world.data.CosmicObjectModels.Get(model.ProjectileObjectModelID)
}

func (world *World) Tick(dtSeconds float64) game.Snapshot {
	world.mu.Lock()
	defer world.mu.Unlock()

	world.stepTasksLocked(dtSeconds)
	world.stepExchangeSessionsLocked(dtSeconds)
	if dtSeconds > 0 {
		world.currentTimeMillis += int64(math.Round(dtSeconds * 1000))
	}
	inputsByObjectID := world.inputsByObjectID()
	world.stepMovableObjects(dtSeconds, inputsByObjectID)
	world.resolveAllCollisions()
	world.stepMiningLocked(dtSeconds, inputsByObjectID)
	world.stepWeaponReloadsLocked()
	world.stepWeaponFireLocked(dtSeconds, inputsByObjectID)
	world.stepProjectilesLocked(dtSeconds)
	world.stepExchangeRequestsLocked(dtSeconds)
	world.stepDockingRequestsLocked(dtSeconds)
	world.stepDockingProcessesLocked(dtSeconds)

	world.tick++
	return world.snapshotLocked(0)
}

// cross2D считает псевдоскалярное произведение двух плоских векторов.
func (world *World) stepDockingRequestsLocked(dtSeconds float64) {
	if dtSeconds <= 0 || len(world.dockingRequests) == 0 {
		return
	}
	remaining := world.dockingRequests[:0]
	for _, request := range world.dockingRequests {
		request.RemainingSeconds -= dtSeconds
		if request.RemainingSeconds > physics.Epsilon {
			remaining = append(remaining, request)
			continue
		}
		world.closeDockingRequestWindowLocked(request)
		world.addDockingNotificationLocked([]int64{request.SenderCosmicObjectID, request.ReceiverCosmicObjectID}, "Истекло время ожидания ответа на запрос стыковки")
	}
	world.dockingRequests = remaining
}

// cross2D считает псевдоскалярное произведение двух плоских векторов.
func (world *World) stepDockingProcessesLocked(dtSeconds float64) {
	if dtSeconds <= 0 || len(world.dockingProcesses) == 0 {
		return
	}
	remaining := world.dockingProcesses[:0]
	for _, process := range world.dockingProcesses {
		process.RemainingSeconds -= dtSeconds
		if process.RemainingSeconds > physics.Epsilon {
			remaining = append(remaining, process)
			continue
		}
		world.completeDockingProcessLocked(process)
	}
	world.dockingProcesses = remaining
}

// cross2D считает псевдоскалярное произведение двух плоских векторов.
func (world *World) completeDockingProcessLocked(process dockingProcess) {
	sender, senderOK := world.data.CosmicObjects.Get(process.SenderCosmicObjectID)
	receiver, receiverOK := world.data.CosmicObjects.Get(process.ReceiverCosmicObjectID)
	if !senderOK || !receiverOK {
		return
	}
	mainID := receiver.ID
	if receiver.ClusterMainCosmicObjectID == receiver.ID {
		mainID = receiver.ClusterMainCosmicObjectID
	}
	receiver.ClusterMainCosmicObjectID = mainID
	sender.ClusterMainCosmicObjectID = mainID
	for _, cosmicObject := range world.data.CosmicObjects.Items {
		if cosmicObject != nil && cosmicObject.ClusterMainCosmicObjectID == mainID {
			cosmicObject.Anchored = true
		}
	}
	world.addDockingWindowEventsLocked("dockingFinished", sender.ID, receiver.ID, 0)
	world.openExchangeAfterDockingLocked(sender.ID, receiver.ID)
	world.addDockingNotificationLocked(world.clusterObjectIDsLocked(mainID), "Объект пристыкован")
}

// cross2D считает псевдоскалярное произведение двух плоских векторов.
// cross2D считает псевдоскалярное произведение двух плоских векторов.
func (world *World) stepTasksLocked(dtSeconds float64) {
	if dtSeconds <= 0 || world.data.Tasks == nil || world.data.TaskTypes == nil || world.data.TaskItemGroups == nil {
		return
	}
	controllerIDs := world.controllerIDsWithTasksLocked()
	for _, controllerID := range controllerIDs {
		if world.exchangePausesControllerLocked(controllerID) {
			continue
		}
		task := world.activeTaskLocked(controllerID)
		if task == nil {
			continue
		}
		if !world.taskHasReserveLocked(task.ID) {
			if !world.reserveTaskItemsLocked(task) {
				continue
			}
		}
		workPower := world.taskWorkPowerLocked(task)
		if workPower <= physics.Epsilon {
			continue
		}
		if controller, ok := world.data.EquipmentGroups.Get(task.ControllerEquipmentGroupID); ok {
			controller.Active = true
		}
		task.RemainingEnergy = math.Max(0, task.RemainingEnergy-workPower*dtSeconds)
		if task.RemainingEnergy > physics.Epsilon {
			continue
		}
		if err := world.completeTaskLocked(task); err != nil {
			continue
		}
		world.data.TaskItemGroups.DeleteByTaskID(task.ID)
		world.data.Tasks.Delete(task.ID)
		_ = world.data.ItemGroups.RebuildIndexes()
	}
}

// cross2D считает псевдоскалярное произведение двух плоских векторов.
func (world *World) controllerIDsWithTasksLocked() []int64 {
	seen := map[int64]bool{}
	controllerIDs := make([]int64, 0)
	for _, task := range world.data.Tasks.Items {
		if task == nil || seen[task.ControllerEquipmentGroupID] {
			continue
		}
		seen[task.ControllerEquipmentGroupID] = true
		controllerIDs = append(controllerIDs, task.ControllerEquipmentGroupID)
	}
	sort.Slice(controllerIDs, func(left int, right int) bool { return controllerIDs[left] < controllerIDs[right] })
	return controllerIDs
}

// cross2D считает псевдоскалярное произведение двух плоских векторов.
func (world *World) activeTaskLocked(controllerID int64) *data.Task {
	tasks := append([]*data.Task(nil), world.data.Tasks.GetByControllerEquipmentGroupID(controllerID)...)
	sort.Slice(tasks, func(left int, right int) bool { return tasks[left].ID < tasks[right].ID })
	for _, task := range tasks {
		if task.RemainingEnergy < task.TotalEnergy || world.taskHasReserveLocked(task.ID) {
			return task
		}
	}
	for _, task := range tasks {
		if task.ParentTaskID > 0 {
			return task
		}
	}
	if len(tasks) == 0 {
		return nil
	}
	return tasks[0]
}

// cross2D считает псевдоскалярное произведение двух плоских векторов.
func (world *World) taskHasReserveLocked(taskID int64) bool {
	task, ok := world.data.Tasks.Get(taskID)
	if ok && (world.taskTypeAcronymLocked(task) == "CargoMovement" || world.taskTypeAcronymLocked(task) == "Fueling" || world.taskTypeAcronymLocked(task) == "ItemDeconstruction") {
		return taskItemGroupsAreStored(world.data.TaskItemGroups.GetByTaskID(taskID))
	}
	return len(world.data.TaskItemGroups.GetByTaskID(taskID)) > 0
}

// cross2D считает псевдоскалярное произведение двух плоских векторов.
func taskItemGroupsAreStored(groups []*data.TaskItemGroup) bool {
	if len(groups) == 0 {
		return false
	}
	for _, group := range groups {
		if group == nil || !group.IsStored {
			return false
		}
	}
	return true
}

// cross2D считает псевдоскалярное произведение двух плоских векторов.
func (world *World) reserveTaskItemsLocked(task *data.Task) bool {
	if world.taskTypeAcronymLocked(task) == "CargoMovement" {
		return world.reserveCargoMovementItemsLocked(task)
	}
	if world.taskTypeAcronymLocked(task) == "Fueling" {
		return world.reserveFuelingItemsLocked(task)
	}
	if world.taskTypeAcronymLocked(task) == "ItemDeconstruction" {
		return world.reserveItemDeconstructionItemsLocked(task)
	}
	requiredByModel, ok := world.taskRequirementsLocked(task)
	if !ok {
		return false
	}
	if len(requiredByModel) == 0 {
		return true
	}
	materialContainer, err := world.taskRelatedContainerLocked(task, "Source")
	if err != nil {
		return false
	}
	availableByModel := map[int64]float64{}
	for _, itemGroup := range world.data.ItemGroups.GetByContainerEquipmentGroupID(materialContainer.ID) {
		availableByModel[itemGroup.ContentItemModelID] += itemGroup.Count
	}
	for itemModelID, required := range requiredByModel {
		if availableByModel[itemModelID]+physics.Epsilon < required {
			return false
		}
	}
	for itemModelID, required := range requiredByModel {
		world.consumeItemModelFromContainerLocked(materialContainer.ID, itemModelID, required)
		if _, err := world.data.TaskItemGroups.Add(&data.TaskItemGroup{TaskID: task.ID, ItemModelID: itemModelID, Count: required}); err != nil {
			return false
		}
	}
	_ = world.data.ItemGroups.RebuildIndexes()
	return true
}

// cross2D считает псевдоскалярное произведение двух плоских векторов.
// cross2D считает псевдоскалярное произведение двух плоских векторов.
func (world *World) reserveItemDeconstructionItemsLocked(task *data.Task) bool {
	requiredGroups := world.data.TaskItemGroups.GetByTaskID(task.ID)
	if taskItemGroupsAreStored(requiredGroups) {
		return true
	}
	if len(requiredGroups) == 0 || task.SourceContainerEquipmentGroupID <= 0 {
		return false
	}
	availableByModel := map[int64]float64{}
	for _, itemGroup := range world.data.ItemGroups.GetByContainerEquipmentGroupID(task.SourceContainerEquipmentGroupID) {
		availableByModel[itemGroup.ContentItemModelID] += itemGroup.Count
	}
	for _, group := range requiredGroups {
		if group != nil && group.IsStored {
			continue
		}
		if group == nil || availableByModel[group.ItemModelID]+physics.Epsilon < group.Count {
			return false
		}
	}
	for _, group := range requiredGroups {
		if group == nil || group.IsStored {
			continue
		}
		world.consumeItemModelFromContainerLocked(task.SourceContainerEquipmentGroupID, group.ItemModelID, group.Count)
		group.IsStored = true
	}
	_ = world.data.ItemGroups.RebuildIndexes()
	return true
}

func (world *World) reserveFuelingItemsLocked(task *data.Task) bool {
	requiredGroups := world.data.TaskItemGroups.GetByTaskID(task.ID)
	if taskItemGroupsAreStored(requiredGroups) {
		return true
	}
	if len(requiredGroups) == 0 || task.SourceContainerEquipmentGroupID <= 0 || task.FuelTankEquipmentGroupID <= 0 {
		return false
	}
	fuelTank, ok := world.data.EquipmentGroups.Get(task.FuelTankEquipmentGroupID)
	if !ok {
		return false
	}
	cosmicObject, ok := world.data.CosmicObjects.Get(fuelTank.CosmicObjectID)
	if !ok {
		return false
	}
	for _, group := range requiredGroups {
		if group == nil || group.IsStored {
			continue
		}
		if task.LeftToRightDirection {
			freeFuel := math.Max(0, cosmicObject.MaxFuel-cosmicObject.Fuel)
			if freeFuel+physics.Epsilon < group.Count {
				return false
			}
			available := 0.0
			for _, itemGroup := range world.data.ItemGroups.GetByContainerEquipmentGroupID(task.SourceContainerEquipmentGroupID) {
				if itemGroup.ContentItemModelID == group.ItemModelID {
					available += itemGroup.Count
				}
			}
			if available+physics.Epsilon < group.Count {
				return false
			}
			world.consumeItemModelFromContainerLocked(task.SourceContainerEquipmentGroupID, group.ItemModelID, group.Count)
		} else {
			if cosmicObject.Fuel+physics.Epsilon < group.Count {
				return false
			}
			cosmicObject.Fuel = math.Max(0, cosmicObject.Fuel-group.Count)
		}
		group.IsStored = true
	}
	_ = world.data.ItemGroups.RebuildIndexes()
	return true
}

// cross2D считает псевдоскалярное произведение двух плоских векторов.
func (world *World) reserveCargoMovementItemsLocked(task *data.Task) bool {
	requiredGroups := world.data.TaskItemGroups.GetByTaskID(task.ID)
	if taskItemGroupsAreStored(requiredGroups) {
		return true
	}
	controller, ok := world.data.EquipmentGroups.Get(task.ControllerEquipmentGroupID)
	if !ok {
		return false
	}
	source, target, err := world.cargoMovementEndpointsLocked(controller, task.LeftToRightDirection)
	if err != nil {
		return false
	}
	if len(requiredGroups) == 0 {
		return false
	}
	availableByModel := map[int64]float64{}
	for _, itemGroup := range world.data.ItemGroups.GetByContainerEquipmentGroupID(source.ID) {
		availableByModel[itemGroup.ContentItemModelID] += itemGroup.Count
	}
	for _, group := range requiredGroups {
		if group != nil && group.IsStored {
			continue
		}
		if group == nil || availableByModel[group.ItemModelID]+physics.Epsilon < group.Count {
			return false
		}
	}
	for _, group := range requiredGroups {
		if group == nil || group.IsStored {
			continue
		}
		world.consumeItemModelFromContainerLocked(source.ID, group.ItemModelID, group.Count)
		group.IsStored = true
	}
	totalEnergy, err := world.cargoMovementTaskEnergyLocked(source.ID, target.ID, requiredGroups)
	if err != nil {
		return false
	}
	task.SourceContainerEquipmentGroupID = source.ID
	task.TargetContainerEquipmentGroupID = target.ID
	task.TotalEnergy = totalEnergy
	task.RemainingEnergy = totalEnergy
	_ = world.data.ItemGroups.RebuildIndexes()
	return true
}

// cross2D считает псевдоскалярное произведение двух плоских векторов.
func (world *World) taskCanReserveItemsLocked(task *data.Task) bool {
	if world.taskTypeAcronymLocked(task) == "CargoMovement" {
		if world.taskHasReserveLocked(task.ID) {
			return true
		}
		controller, ok := world.data.EquipmentGroups.Get(task.ControllerEquipmentGroupID)
		if !ok {
			return false
		}
		source, _, err := world.cargoMovementEndpointsLocked(controller, task.LeftToRightDirection)
		if err != nil {
			return false
		}
		availableByModel := map[int64]float64{}
		for _, itemGroup := range world.data.ItemGroups.GetByContainerEquipmentGroupID(source.ID) {
			availableByModel[itemGroup.ContentItemModelID] += itemGroup.Count
		}
		for _, group := range world.data.TaskItemGroups.GetByTaskID(task.ID) {
			if group != nil && group.IsStored {
				continue
			}
			if group == nil || availableByModel[group.ItemModelID]+physics.Epsilon < group.Count {
				return false
			}
		}
		return true
	}
	if world.taskTypeAcronymLocked(task) == "Fueling" {
		if world.taskHasReserveLocked(task.ID) {
			return true
		}
		if task.SourceContainerEquipmentGroupID <= 0 || task.FuelTankEquipmentGroupID <= 0 {
			return false
		}
		fuelTank, ok := world.data.EquipmentGroups.Get(task.FuelTankEquipmentGroupID)
		if !ok {
			return false
		}
		cosmicObject, ok := world.data.CosmicObjects.Get(fuelTank.CosmicObjectID)
		if !ok {
			return false
		}
		for _, group := range world.data.TaskItemGroups.GetByTaskID(task.ID) {
			if group != nil && group.IsStored {
				continue
			}
			if group == nil {
				return false
			}
			if task.LeftToRightDirection {
				if math.Max(0, cosmicObject.MaxFuel-cosmicObject.Fuel)+physics.Epsilon < group.Count {
					return false
				}
				available := 0.0
				for _, itemGroup := range world.data.ItemGroups.GetByContainerEquipmentGroupID(task.SourceContainerEquipmentGroupID) {
					if itemGroup.ContentItemModelID == group.ItemModelID {
						available += itemGroup.Count
					}
				}
				if available+physics.Epsilon < group.Count {
					return false
				}
			} else if cosmicObject.Fuel+physics.Epsilon < group.Count {
				return false
			}
		}
		return true
	}
	requiredByModel, ok := world.taskRequirementsLocked(task)
	if !ok {
		return false
	}
	if len(requiredByModel) == 0 {
		return true
	}
	materialContainer, err := world.taskRelatedContainerLocked(task, "Source")
	if err != nil {
		return false
	}
	availableByModel := map[int64]float64{}
	for _, itemGroup := range world.data.ItemGroups.GetByContainerEquipmentGroupID(materialContainer.ID) {
		availableByModel[itemGroup.ContentItemModelID] += itemGroup.Count
	}
	for itemModelID, required := range requiredByModel {
		if availableByModel[itemModelID]+physics.Epsilon < required {
			return false
		}
	}
	return true
}

// cross2D считает псевдоскалярное произведение двух плоских векторов.
func (world *World) taskRequirementsLocked(task *data.Task) (map[int64]float64, bool) {
	components, err := world.taskComponentsLocked(task)
	if err != nil {
		return nil, false
	}
	requiredByModel := map[int64]float64{}
	amount := taskCount(task)
	for _, component := range components {
		requiredByModel[component.ComponentItemModelID] += component.Count * amount
	}
	return requiredByModel, true
}

// cross2D считает псевдоскалярное произведение двух плоских векторов.
func taskCount(task *data.Task) float64 {
	if task == nil || task.BatchCount <= 0 {
		return 1
	}
	return float64(task.BatchCount)
}

// cross2D считает псевдоскалярное произведение двух плоских векторов.
func (world *World) taskComponentsLocked(task *data.Task) ([]controlPanelItemSchemaComponent, error) {
	if world.taskTypeAcronymLocked(task) == "CargoMovement" || world.taskTypeAcronymLocked(task) == "Fueling" {
		return nil, nil
	}
	if task.BlueprintID > 0 {
		return world.objectBlueprintComponentsLocked(task.BlueprintID)
	}
	return world.itemSchemaComponentsLocked(task.SchemaID)
}

// cross2D считает псевдоскалярное произведение двух плоских векторов.
func (world *World) taskTypeAcronymLocked(task *data.Task) string {
	if task == nil || world.data.TaskTypes == nil {
		return ""
	}
	taskType, ok := world.data.TaskTypes.Get(task.TaskTypeID)
	if !ok {
		return ""
	}
	return taskType.Acronym
}

// cross2D считает псевдоскалярное произведение двух плоских векторов.
func (world *World) taskRelatedContainerLocked(task *data.Task, relationTypeAcronym string) (*data.EquipmentGroup, error) {
	controller, ok := world.data.EquipmentGroups.Get(task.ControllerEquipmentGroupID)
	if !ok {
		return nil, errors.New("task controller equipment group not found")
	}
	return world.constructorRelatedContainerOrFallbackLocked(controller.CosmicObjectID, task.ControllerEquipmentGroupID, relationTypeAcronym, 0)
}

// cross2D считает псевдоскалярное произведение двух плоских векторов.
func (world *World) taskWorkPowerLocked(task *data.Task) float64 {
	if world.data.Implementers == nil || world.data.ItemTypes == nil || world.data.ItemModels == nil || world.data.EquipmentGroups == nil {
		return 0
	}
	controller, ok := world.data.EquipmentGroups.Get(task.ControllerEquipmentGroupID)
	if !ok {
		return 0
	}
	power := 0.0
	for _, implementer := range world.data.Implementers.GetByTaskTypeID(task.TaskTypeID) {
		for _, group := range world.data.EquipmentGroups.GetByCosmicObjectID(controller.CosmicObjectID) {
			model, ok := world.data.ItemModels.Get(group.EquipmentItemModelID)
			if !ok || model.ItemTypeID != implementer.ImplementerEquipmentItemTypeID {
				continue
			}
			efficiency := model.Efficiency
			if efficiency <= 0 {
				efficiency = 1
			}
			modelPower := model.ConsumingPower
			if modelPower <= 0 {
				modelPower = 1
			}
			power += modelPower * float64(enabledEquipmentCount(group)) * implementer.WorkPart * efficiency
		}
	}
	return power
}

// cross2D считает псевдоскалярное произведение двух плоских векторов.
func (world *World) completeTaskLocked(task *data.Task) error {
	if world.taskTypeAcronymLocked(task) == "CargoMovement" {
		targetContainerID := task.TargetContainerEquipmentGroupID
		if targetContainerID <= 0 {
			targetContainerID = task.ControllerEquipmentGroupID
		}
		for _, group := range world.data.TaskItemGroups.GetByTaskID(task.ID) {
			if group == nil || !group.IsStored {
				continue
			}
			if err := world.addItemModelToContainerLocked(targetContainerID, group.ItemModelID, group.Count); err != nil {
				return err
			}
		}
		return nil
	}
	if world.taskTypeAcronymLocked(task) == "Fueling" {
		return world.completeFuelingTaskLocked(task)
	}
	if world.taskTypeAcronymLocked(task) == "ItemDeconstruction" {
		return world.completeItemDeconstructionTaskLocked(task)
	}
	job := constructorProductionJob{ConstructorEquipmentGroupID: task.ControllerEquipmentGroupID, SchemaID: task.SchemaID, BlueprintID: task.BlueprintID}
	if task.BlueprintID > 0 {
		blueprint, err := world.objectBlueprintLocked(task.BlueprintID)
		if err != nil {
			return err
		}
		job.ProductCosmicObjectModelID = blueprint.CosmicObjectModelID
		amount := int(math.Ceil(taskCount(task) - physics.Epsilon))
		if amount < 1 {
			amount = 1
		}
		for index := 0; index < amount; index++ {
			if err := world.createConstructedCosmicObjectLocked(&job); err != nil {
				return err
			}
		}
		return nil
	}
	schema, err := world.itemSchemaLocked(task.SchemaID)
	if err != nil {
		return err
	}
	job.ProductItemModelID = schema.ItemModelID
	job.ProductCount = schema.Count * taskCount(task)
	relationTypeAcronym := "Destination"
	if task.ParentTaskID > 0 {
		relationTypeAcronym = "Source"
	}
	productContainer, err := world.taskRelatedContainerLocked(task, relationTypeAcronym)
	if err != nil {
		return err
	}
	return world.addItemModelToContainerLocked(productContainer.ID, job.ProductItemModelID, job.ProductCount)
}

// cross2D считает псевдоскалярное произведение двух плоских векторов.
// cross2D считает псевдоскалярное произведение двух плоских векторов.
func (world *World) completeItemDeconstructionTaskLocked(task *data.Task) error {
	if task == nil || task.SchemaID <= 0 || task.TargetContainerEquipmentGroupID <= 0 {
		return errors.New("item deconstruction task is invalid")
	}
	components, err := world.itemSchemaComponentsLocked(task.SchemaID)
	if err != nil {
		return err
	}
	batches := taskCount(task)
	for _, component := range components {
		if component.Count <= 0 {
			continue
		}
		if err := world.addItemModelToContainerLocked(task.TargetContainerEquipmentGroupID, component.ComponentItemModelID, component.Count*batches); err != nil {
			return err
		}
	}
	return nil
}

func (world *World) completeFuelingTaskLocked(task *data.Task) error {
	if task == nil || task.FuelTankEquipmentGroupID <= 0 || task.SourceContainerEquipmentGroupID <= 0 {
		return errors.New("fueling task is invalid")
	}
	fuelTank, ok := world.data.EquipmentGroups.Get(task.FuelTankEquipmentGroupID)
	if !ok {
		return errors.New("fuel tank equipment group not found")
	}
	cosmicObject, ok := world.data.CosmicObjects.Get(fuelTank.CosmicObjectID)
	if !ok {
		return errors.New("fuel tank object not found")
	}
	for _, group := range world.data.TaskItemGroups.GetByTaskID(task.ID) {
		if group == nil || !group.IsStored {
			continue
		}
		if task.LeftToRightDirection {
			cosmicObject.Fuel = math.Min(cosmicObject.MaxFuel, cosmicObject.Fuel+group.Count)
			continue
		}
		if err := world.addItemModelToContainerLocked(task.SourceContainerEquipmentGroupID, group.ItemModelID, group.Count); err != nil {
			return err
		}
	}
	return nil
}
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

// cross2D считает псевдоскалярное произведение двух плоских векторов.
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

// cross2D считает псевдоскалярное произведение двух плоских векторов.
// cross2D считает псевдоскалярное произведение двух плоских векторов.
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

// cross2D считает псевдоскалярное произведение двух плоских векторов.
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

// cross2D считает псевдоскалярное произведение двух плоских векторов.
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

// cross2D считает псевдоскалярное произведение двух плоских векторов.
func (world *World) constructorMainJobIndexLocked(constructorID int64, jobID int64) int {
	for index, job := range world.constructorProductionJobs {
		if job.ID == jobID && job.ConstructorEquipmentGroupID == constructorID && job.QueueType == "main" {
			return index
		}
	}
	return -1
}

// cross2D считает псевдоскалярное произведение двух плоских векторов.
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

// cross2D считает псевдоскалярное произведение двух плоских векторов.
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

// cross2D считает псевдоскалярное произведение двух плоских векторов.
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

// cross2D считает псевдоскалярное произведение двух плоских векторов.
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

// cross2D считает псевдоскалярное произведение двух плоских векторов.
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

// cross2D считает псевдоскалярное произведение двух плоских векторов.
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

// cross2D считает псевдоскалярное произведение двух плоских векторов.
// cross2D считает псевдоскалярное произведение двух плоских векторов.
// cross2D считает псевдоскалярное произведение двух плоских векторов.
func (world *World) constructorEquipmentIsWorkingLocked(groupID int64) bool {
	if world.data.Tasks != nil {
		task := world.activeTaskLocked(groupID)
		if task == nil {
			group, ok := world.data.EquipmentGroups.Get(groupID)
			if ok && world.equipmentGroupHasItemTypeLocked(group, "Constructor") {
				for _, candidate := range world.data.Tasks.Items {
					controller, controllerOk := world.data.EquipmentGroups.Get(candidate.ControllerEquipmentGroupID)
					if candidate != nil && controllerOk && controller.CosmicObjectID == group.CosmicObjectID {
						task = candidate
						break
					}
				}
			}
		}
		if task == nil {
			return false
		}
		return task.RemainingEnergy < task.TotalEnergy || world.taskHasReserveLocked(task.ID) || (world.taskWorkPowerLocked(task) > physics.Epsilon && world.taskCanReserveItemsLocked(task))
	}
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

// cross2D считает псевдоскалярное произведение двух плоских векторов.
func (world *World) equipmentGroupHasItemTypeLocked(group *data.EquipmentGroup, itemTypeAcronym string) bool {
	if group == nil || world.data.ItemModels == nil || world.data.ItemTypes == nil {
		return false
	}
	model, ok := world.data.ItemModels.Get(group.EquipmentItemModelID)
	if !ok {
		return false
	}
	itemType, ok := world.data.ItemTypes.Get(model.ItemTypeID)
	return ok && itemType.Acronym == itemTypeAcronym
}

// cross2D считает псевдоскалярное произведение двух плоских векторов.
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

// cross2D считает псевдоскалярное произведение двух плоских векторов.
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

// cross2D считает псевдоскалярное произведение двух плоских векторов.
func cosmicObjectIsFullyStopped(cosmicObject data.CosmicObject) bool {
	return math.Abs(cosmicObject.VelocityX) <= physics.Epsilon &&
		math.Abs(cosmicObject.VelocityY) <= physics.Epsilon &&
		math.Abs(cosmicObject.Speed) <= physics.Epsilon &&
		math.Abs(cosmicObject.AngularSpeed) <= physics.Epsilon
}

// cross2D считает псевдоскалярное произведение двух плоских векторов.
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
		input, controlled := inputsByObjectID[objectID]
		if !cosmicObject.Enabled {
			world.updateEquipmentUsage(cosmicObject, game.ShipInput{}, dtSeconds)
			continue
		}
		model, ok := world.data.CosmicObjectModels.Get(cosmicObject.CosmicObjectModelID)
		if !ok {
			world.updateEquipmentUsage(cosmicObject, input, dtSeconds)
			continue
		}
		if cosmicObject.Anchored {
			world.updateEquipmentUsage(cosmicObject, input, dtSeconds)
			continue
		}

		isShip := world.isShipModel(model)
		if controlled && (!isShip || shipHasFuel(*cosmicObject)) {
			*cosmicObject = world.stepControlledObject(*cosmicObject, *model, input, dtSeconds)
		} else if isShip {
			*cosmicObject = physics.StepUnpilotedShip(*cosmicObject, dtSeconds)
		} else {
			*cosmicObject = physics.StepFreeBody(*cosmicObject, dtSeconds)
		}
		world.updateEquipmentUsage(cosmicObject, input, dtSeconds)
	}
}

// cross2D считает псевдоскалярное произведение двух плоских векторов.
func (world *World) stepControlledObject(cosmicObject data.CosmicObject, model data.CosmicObjectModel, input game.ShipInput, dtSeconds float64) data.CosmicObject {
	effectiveObject := world.objectWithEnabledEquipmentForces(cosmicObject, input)
	next := physics.StepShip(effectiveObject, model, input, dtSeconds)
	next.MaxAlongForce = cosmicObject.MaxAlongForce
	next.MaxAcrossForce = cosmicObject.MaxAcrossForce
	next.MaxTorque = cosmicObject.MaxTorque
	return next
}

// cross2D считает псевдоскалярное произведение двух плоских векторов.
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

// cross2D считает псевдоскалярное произведение двух плоских векторов.
func (world *World) electricShareForInput(cosmicObject data.CosmicObject, input game.ShipInput) float64 {
	if world.data.EquipmentGroups == nil || world.data.ItemModels == nil {
		return 1
	}

	generatedPower := 0.0
	neededPower := 0.0
	primaryModelID, primaryModelOK := world.activePrimaryPilotToolModelIDLocked(cosmicObject.ID, input)
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
		if equipmentNeedsElectricityForInput(input, *model, primaryModelID, primaryModelOK) {
			neededPower += model.ConsumingPower * count
		}
	}

	return electricWorkShare(generatedPower, neededPower)
}

// cross2D считает псевдоскалярное произведение двух плоских векторов.
func (world *World) updateEquipmentUsage(cosmicObject *data.CosmicObject, input game.ShipInput, dtSeconds float64) {
	cosmicObject.ConsumingPower = 0
	cosmicObject.GeneratingPower = 0
	if world.data.EquipmentGroups == nil || world.data.ItemModels == nil {
		return
	}

	consumerFuelConsumptionPerSecond := 0.0
	generatorFuelConsumptionPerSecond := 0.0
	neededPower := 0.0
	generatorGroups := make([]*data.EquipmentGroup, 0)
	primaryModelID, primaryModelOK := world.activePrimaryPilotToolModelIDLocked(cosmicObject.ID, input)
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

		group.Active = equipmentIsActive(*cosmicObject, *model) ||
			equipmentIsActiveForPrimaryAction(*model, primaryModelID, primaryModelOK) ||
			world.constructorEquipmentIsWorkingLocked(group.ID)
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

// cross2D считает псевдоскалярное произведение двух плоских векторов.
func enabledEquipmentCount(group *data.EquipmentGroup) int64 {
	if group == nil || !group.Enabled {
		return 0
	}
	return group.EnabledCount
}

// cross2D считает псевдоскалярное произведение двух плоских векторов.
func shipHasFuel(cosmicObject data.CosmicObject) bool {
	return cosmicObject.Fuel > physics.Epsilon
}

// cross2D считает псевдоскалярное произведение двух плоских векторов.
func equipmentConsumesStoredFuel(model data.ItemModel) bool {
	return model.ConsumingItemModelID > 0 && model.ConsumingCount > 0
}

// cross2D считает псевдоскалярное произведение двух плоских векторов.
func electricWorkShare(generatedPower float64, neededPower float64) float64 {
	if neededPower <= 0 || generatedPower >= neededPower {
		return 1
	}
	if generatedPower <= 0 {
		return 0
	}
	return generatedPower / neededPower
}

// cross2D считает псевдоскалярное произведение двух плоских векторов.
func equipmentNeedsElectricityForInput(input game.ShipInput, model data.ItemModel, primaryModelID int64, primaryModelOK bool) bool {
	usesAlongForce := model.MaxAlongForce != 0 && (input.ThrustForward || input.ThrustBackward)
	usesAcrossForce := model.MaxAcrossForce != 0 && (input.ThrustLeft || input.ThrustRight)
	usesTorque := model.MaxTorque != 0 && input.TargetRotationDelta != 0
	usesPrimaryAction := equipmentIsActiveForPrimaryAction(model, primaryModelID, primaryModelOK)
	return usesAlongForce || usesAcrossForce || usesTorque || usesPrimaryAction
}

// cross2D считает псевдоскалярное произведение двух плоских векторов.
func equipmentIsActiveForPrimaryAction(model data.ItemModel, primaryModelID int64, primaryModelOK bool) bool {
	return primaryModelOK && model.ID == primaryModelID
}

// cross2D считает псевдоскалярное произведение двух плоских векторов.
func equipmentIsActive(cosmicObject data.CosmicObject, model data.ItemModel) bool {
	usesLinearForce := model.MaxAlongForce != 0 || model.MaxAcrossForce != 0
	usesTorque := model.MaxTorque != 0
	if usesLinearForce || usesTorque {
		return (usesLinearForce && (math.Abs(cosmicObject.AlongForce) > physics.Epsilon || math.Abs(cosmicObject.AcrossForce) > physics.Epsilon)) ||
			(usesTorque && math.Abs(cosmicObject.Torque) > physics.Epsilon)
	}

	return false
}

// cross2D считает псевдоскалярное произведение двух плоских векторов.
func sortedEquipmentGroups(groups []*data.EquipmentGroup) []*data.EquipmentGroup {
	result := append([]*data.EquipmentGroup(nil), groups...)
	sort.Slice(result, func(left int, right int) bool {
		return result[left].ID < result[right].ID
	})
	return result
}

// cross2D считает псевдоскалярное произведение двух плоских векторов.
func (world *World) ackMutationLocked(accountID int64, sessionID string, mutationSeq int64) {
	if sessionID == "" || mutationSeq <= 0 {
		return
	}
	key := mutationAckKey(accountID, sessionID)
	if mutationSeq > world.mutationAcks[key] {
		world.mutationAcks[key] = mutationSeq
	}
}

// cross2D считает псевдоскалярное произведение двух плоских векторов.
func mutationAckKey(accountID int64, sessionID string) string {
	return fmt.Sprintf("%d:%s", accountID, sessionID)
}

// cross2D считает псевдоскалярное произведение двух плоских векторов.
func (world *World) isShipModel(model *data.CosmicObjectModel) bool {
	cosmicObjectType, ok := world.data.CosmicObjectTypes.Get(model.CosmicObjectTypeID)
	return ok && cosmicObjectType.Acronym == "Ship"
}

// cross2D считает псевдоскалярное произведение двух плоских векторов.
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

// cross2D считает псевдоскалярное произведение двух плоских векторов.
func (world *World) SnapshotForAccount(accountID int64) game.Snapshot {
	world.mu.Lock()
	defer world.mu.Unlock()

	objectID := world.accountObjectIDs[accountID]
	return world.snapshotLocked(objectID)
}

// cross2D считает псевдоскалярное произведение двух плоских векторов.
func (world *World) ObjectIDForAccount(accountID int64) (int64, bool) {
	world.mu.Lock()
	defer world.mu.Unlock()

	objectID, ok := world.accountObjectIDs[accountID]
	return objectID, ok
}

// cross2D считает псевдоскалярное произведение двух плоских векторов.
func (world *World) DrainDockingEvents() []game.DockingEvent {
	world.mu.Lock()
	defer world.mu.Unlock()

	events := append([]game.DockingEvent(nil), world.dockingEvents...)
	world.dockingEvents = nil
	return events
}

// cross2D считает псевдоскалярное произведение двух плоских векторов.
func (world *World) ClientMutationAck(accountID int64, sessionID string) game.ClientMutationAck {
	world.mu.Lock()
	defer world.mu.Unlock()

	return game.ClientMutationAck{
		SessionID:      sessionID,
		LastAppliedSeq: world.mutationAcks[mutationAckKey(accountID, sessionID)],
	}
}

// cross2D считает псевдоскалярное произведение двух плоских векторов.
func (world *World) CharacterByID(id int64) (*data.Character, bool) {
	world.mu.Lock()
	defer world.mu.Unlock()

	return world.data.Characters.Get(id)
}

// cross2D считает псевдоскалярное произведение двух плоских векторов.
func (world *World) CosmicObjectByID(id int64) (*data.CosmicObject, bool) {
	world.mu.Lock()
	defer world.mu.Unlock()

	return world.data.CosmicObjects.Get(id)
}

// cross2D считает псевдоскалярное произведение двух плоских векторов.
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
	if world.data.Tasks != nil {
		if err := world.data.Tasks.SaveToFile(filepath.Join(dataDirectory, "Tasks.json")); err != nil {
			return err
		}
	}
	if world.data.TaskItemGroups != nil {
		if err := world.data.TaskItemGroups.SaveToFile(filepath.Join(dataDirectory, "TaskItemGroups.json")); err != nil {
			return err
		}
	}
	if world.data.ItemTypes != nil {
		if err := world.data.ItemTypes.SaveToFile(filepath.Join(dataDirectory, "ItemTypes.json")); err != nil {
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

// cross2D считает псевдоскалярное произведение двух плоских векторов.
// cross2D считает псевдоскалярное произведение двух плоских векторов.
func (world *World) activeDrillRayObjectsLocked() []game.SnapshotCosmicObject {
	if world.data.CosmicObjects == nil || world.data.CosmicObjectModels == nil {
		return nil
	}
	rayModel, ok := world.data.CosmicObjectModels.GetByAcronym(drillRayAcronym)
	if !ok {
		return nil
	}

	inputs := world.inputsByObjectID()
	objectIDs := make([]int64, 0, len(inputs))
	for objectID, input := range inputs {
		if input.PrimaryPointerAction {
			objectIDs = append(objectIDs, objectID)
		}
	}
	sort.Slice(objectIDs, func(left int, right int) bool {
		return objectIDs[left] < objectIDs[right]
	})

	rays := make([]game.SnapshotCosmicObject, 0, len(objectIDs))
	for _, objectID := range objectIDs {
		input := inputs[objectID]
		cosmicObject, ok := world.data.CosmicObjects.Get(objectID)
		if !ok || !cosmicObject.Enabled {
			continue
		}
		model, ok := world.data.CosmicObjectModels.Get(cosmicObject.CosmicObjectModelID)
		if !ok {
			continue
		}
		if _, ok := world.selectedSimpleDrillRangeLocked(objectID, input); !ok {
			continue
		}

		forward := physics.ForwardVector(cosmicObject.Rotation)
		centerDistance := modelVisualForwardOffsetMeters(*model) + rayModel.BodyLength/2
		ray := data.CosmicObject{
			ID:                  -objectID,
			Title:               drillRayAcronym,
			CosmicObjectModelID: rayModel.ID,
			X:                   cosmicObject.X + forward.X*centerDistance,
			Y:                   cosmicObject.Y + forward.Y*centerDistance,
			Rotation:            cosmicObject.Rotation,
			TargetRotation:      cosmicObject.Rotation,
			Enabled:             true,
		}
		rays = append(rays, game.SnapshotCosmicObject{CosmicObject: ray})
	}

	return rays
}

// cross2D считает псевдоскалярное произведение двух плоских векторов.
func modelVisualForwardOffsetMeters(model data.CosmicObjectModel) float64 {
	if len(model.BodyPolygon) > 0 {
		forwardOffset := model.BodyPolygon[0].Y
		for _, point := range model.BodyPolygon[1:] {
			if point.Y > forwardOffset {
				forwardOffset = point.Y
			}
		}
		return forwardOffset
	}

	return model.BodyLength / 2
}

// cross2D считает псевдоскалярное произведение двух плоских векторов.
// Возвращает поперечное смещение орудия с удвоенным боковым отступом от края корпуса.
func weaponBarrelLateralOffsetMeters(model data.CosmicObjectModel, barrelIndex int64, barrelCount int64) float64 {
	if barrelCount <= 1 {
		return 0
	}
	halfWidth := modelVisualHalfWidthMeters(model)
	gap := halfWidth * 2 / float64(barrelCount+3)
	return -halfWidth + gap*(float64(barrelIndex)+2)
}

// Возвращает половину визуальной ширины физического корпуса.
func modelVisualHalfWidthMeters(model data.CosmicObjectModel) float64 {
	if len(model.BodyPolygon) > 0 {
		halfWidth := math.Abs(model.BodyPolygon[0].X)
		for _, point := range model.BodyPolygon[1:] {
			if width := math.Abs(point.X); width > halfWidth {
				halfWidth = width
			}
		}
		return halfWidth
	}

	return model.BodyWidth / 2
}

func (world *World) activePrimaryPilotToolModelIDLocked(objectID int64, input game.ShipInput) (int64, bool) {
	if !input.PrimaryPointerAction {
		return 0, false
	}
	tools := world.pilotInstrumentModelsLocked(objectID)
	if len(tools) == 0 {
		return 0, false
	}
	index := normalizePilotToolIndex(input.SelectedPilotToolIndex)
	if index >= len(tools) || tools[index].EnabledCount <= 0 {
		return 0, false
	}
	return tools[index].ModelID, true
}

// cross2D считает псевдоскалярное произведение двух плоских векторов.
func (world *World) pilotInstrumentModelsLocked(objectID int64) []pilotInstrumentModel {
	if world.data.EquipmentGroups == nil || world.data.ItemModels == nil || world.data.ItemTypes == nil {
		return nil
	}

	byModelID := map[int64]*pilotInstrumentModel{}
	for _, group := range sortedEquipmentGroups(world.data.EquipmentGroups.GetByCosmicObjectID(objectID)) {
		if group == nil || group.Count <= 0 {
			continue
		}
		model, ok := world.data.ItemModels.Get(group.EquipmentItemModelID)
		if !ok {
			continue
		}
		itemType, ok := world.data.ItemTypes.Get(model.ItemTypeID)
		if !ok || !itemType.IsPilotInstrument {
			continue
		}
		tool, ok := byModelID[model.ID]
		if !ok {
			tool = &pilotInstrumentModel{ModelID: model.ID, FirstGroupID: group.ID}
			byModelID[model.ID] = tool
		}
		if group.ID < tool.FirstGroupID {
			tool.FirstGroupID = group.ID
		}
		tool.EnabledCount += enabledEquipmentCount(group)
	}

	result := make([]pilotInstrumentModel, 0, len(byModelID))
	for _, tool := range byModelID {
		result = append(result, *tool)
	}
	sort.Slice(result, func(left int, right int) bool {
		return result[left].FirstGroupID < result[right].FirstGroupID
	})
	if len(result) > pilotToolSlotCount {
		return result[:pilotToolSlotCount]
	}
	return result
}

// cross2D считает псевдоскалярное произведение двух плоских векторов.
func normalizePilotToolIndex(index int) int {
	result := index % pilotToolSlotCount
	if result < 0 {
		result += pilotToolSlotCount
	}
	return result
}

// cross2D считает псевдоскалярное произведение двух плоских векторов.
func (world *World) activeProjectileObjectsLocked() []game.SnapshotCosmicObject {
	if len(world.projectiles) == 0 {
		return nil
	}
	objects := make([]game.SnapshotCosmicObject, 0, len(world.projectiles))
	for _, projectile := range world.projectiles {
		objects = append(objects, game.SnapshotCosmicObject{CosmicObject: projectile.CosmicObject})
	}
	return objects
}

// cross2D считает псевдоскалярное произведение двух плоских векторов.
func (world *World) selectedWeaponAttackParametersLocked(objectID int64, input game.ShipInput) (weaponAttackParameters, bool) {
	modelID, ok := world.activePrimaryPilotToolModelIDLocked(objectID, input)
	if !ok || world.data.EquipmentGroups == nil || world.data.ItemModels == nil || world.data.ItemTypes == nil {
		return weaponAttackParameters{}, false
	}

	parameters := weaponAttackParameters{}
	for _, group := range sortedEquipmentGroups(world.data.EquipmentGroups.GetByCosmicObjectID(objectID)) {
		enabledCount := enabledEquipmentCount(group)
		if group.EquipmentItemModelID != modelID || enabledCount <= 0 {
			continue
		}
		model, ok := world.data.ItemModels.Get(group.EquipmentItemModelID)
		if !ok || !world.itemModelHasTypeAcronymLocked(model, weaponItemTypeAcronym) || model.Range <= 0 || model.FiringRate <= 0 || model.ProjectileObjectModelID <= 0 {
			continue
		}
		projectileModel, ok := world.projectileModelForWeaponLocked(model)
		if !ok || projectileModel.Damage <= 0 || projectileModel.MaxSpeed <= 0 || !world.cosmicObjectModelHasProjectileTypeLocked(projectileModel) {
			continue
		}
		parameters.ItemModelID = model.ID
		parameters.ProjectileModelID = projectileModel.ID
		parameters.Damage = projectileModel.Damage
		parameters.ProjectileSpeed = projectileModel.MaxSpeed
		if model.Range > parameters.Range {
			parameters.Range = model.Range
		}
		barrelStartIndex := parameters.InitialProjectileCount
		shotsPerSecond := model.FiringRate
		parameters.ShotsPerSecond += shotsPerSecond
		parameters.InitialProjectileCount += enabledCount
		parameters.Groups = append(parameters.Groups, weaponAttackGroup{
			BarrelStartIndex:       barrelStartIndex,
			EquipmentGroup:         group,
			EnabledCount:           enabledCount,
			ShotsPerSecond:         shotsPerSecond,
			InitialProjectileCount: enabledCount,
		})
	}

	return parameters, parameters.ItemModelID > 0 && parameters.ProjectileModelID > 0 && parameters.Range > 0 && parameters.Damage > 0 && parameters.ProjectileSpeed > 0 && parameters.ShotsPerSecond > 0 && parameters.InitialProjectileCount > 0 && len(parameters.Groups) > 0
}

// cross2D считает псевдоскалярное произведение двух плоских векторов.
func (world *World) selectedSimpleDrillRangeLocked(objectID int64, input game.ShipInput) (float64, bool) {
	modelID, ok := world.activePrimaryPilotToolModelIDLocked(objectID, input)
	if !ok || world.data.EquipmentGroups == nil || world.data.ItemModels == nil {
		return 0, false
	}

	var selectedRange float64
	for _, group := range sortedEquipmentGroups(world.data.EquipmentGroups.GetByCosmicObjectID(objectID)) {
		if group.EquipmentItemModelID != modelID || enabledEquipmentCount(group) <= 0 {
			continue
		}
		model, ok := world.data.ItemModels.Get(group.EquipmentItemModelID)
		if !ok || model.Acronym != simpleDrillAcronym || model.Range <= 0 {
			continue
		}
		if model.Range > selectedRange {
			selectedRange = model.Range
		}
	}

	return selectedRange, selectedRange > 0
}

// cross2D считает псевдоскалярное произведение двух плоских векторов.
func (world *World) selectedSimpleDrillMiningParametersLocked(objectID int64, input game.ShipInput) (drillMiningParameters, bool) {
	modelID, ok := world.activePrimaryPilotToolModelIDLocked(objectID, input)
	if !ok || world.data.EquipmentGroups == nil || world.data.ItemModels == nil {
		return drillMiningParameters{}, false
	}

	parameters := drillMiningParameters{}
	for _, group := range sortedEquipmentGroups(world.data.EquipmentGroups.GetByCosmicObjectID(objectID)) {
		if group.EquipmentItemModelID != modelID || enabledEquipmentCount(group) <= 0 {
			continue
		}
		model, ok := world.data.ItemModels.Get(group.EquipmentItemModelID)
		if !ok || model.Acronym != simpleDrillAcronym || model.Range <= 0 || model.MiningSpeed <= 0 {
			continue
		}
		if model.Range > parameters.Range {
			parameters.Range = model.Range
		}
		parameters.MiningSpeed = model.MiningSpeed
		parameters.EnabledCount += enabledEquipmentCount(group)
	}

	return parameters, parameters.Range > 0 && parameters.MiningSpeed > 0 && parameters.EnabledCount > 0
}

func (world *World) snapshotLocked(selfObjectID int64) game.Snapshot {
	objectIDs := make([]int64, 0, len(world.data.CosmicObjects.Items))
	for objectID := range world.data.CosmicObjects.Items {
		objectIDs = append(objectIDs, objectID)
	}
	sort.Slice(objectIDs, func(left int, right int) bool {
		return objectIDs[left] < objectIDs[right]
	})

	objects := make([]game.SnapshotCosmicObject, 0, len(objectIDs))
	for _, objectID := range objectIDs {
		cosmicObject, ok := world.data.CosmicObjects.Get(objectID)
		if !ok {
			continue
		}
		objects = append(objects, game.SnapshotCosmicObject{
			CosmicObject: *cosmicObject,
			OwnerName:    world.ownerNameForTestingLocked(cosmicObject.OwnerCharacterID),
		})
	}
	objects = append(objects, world.activeDrillRayObjectsLocked()...)
	objects = append(objects, world.activeProjectileObjectsLocked()...)

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

	tasks := make([]data.Task, 0)
	if world.data.Tasks != nil {
		taskIDs := make([]int64, 0, len(world.data.Tasks.Items))
		for taskID := range world.data.Tasks.Items {
			taskIDs = append(taskIDs, taskID)
		}
		sort.Slice(taskIDs, func(left int, right int) bool { return taskIDs[left] < taskIDs[right] })
		for _, taskID := range taskIDs {
			task, ok := world.data.Tasks.Get(taskID)
			if !ok {
				continue
			}
			tasks = append(tasks, *task)
		}
	}

	taskItemGroups := make([]data.TaskItemGroup, 0)
	if world.data.TaskItemGroups != nil {
		groupIDs := make([]int64, 0, len(world.data.TaskItemGroups.Items))
		for groupID := range world.data.TaskItemGroups.Items {
			groupIDs = append(groupIDs, groupID)
		}
		sort.Slice(groupIDs, func(left int, right int) bool { return groupIDs[left] < groupIDs[right] })
		for _, groupID := range groupIDs {
			group, ok := world.data.TaskItemGroups.Items[groupID]
			if !ok || group == nil {
				continue
			}
			taskItemGroups = append(taskItemGroups, *group)
		}
	}
	constructorProductionJobs := world.constructorProductionJobsForTestsLocked(tasks)

	return game.Snapshot{
		Type:                      "snapshot",
		Tick:                      world.tick,
		SelfObjectID:              selfObjectID,
		Objects:                   objects,
		EquipmentGroups:           equipmentGroups,
		ItemGroups:                itemGroups,
		Tasks:                     tasks,
		TaskItemGroups:            taskItemGroups,
		ConstructorProductionJobs: constructorProductionJobs,
	}
}

// cross2D считает псевдоскалярное произведение двух плоских векторов.
// cross2D считает псевдоскалярное произведение двух плоских векторов.
func (world *World) constructorProductionJobsForTestsLocked(tasks []data.Task) []game.ConstructorProductionJob {
	byKey := map[string]*game.ConstructorProductionJob{}
	keys := make([]string, 0)
	for _, task := range tasks {
		queueType := "main"
		if task.ParentTaskID > 0 {
			queueType = "auxiliary"
		}
		key := fmt.Sprintf("%d:%s:%d:%d:%d", task.ControllerEquipmentGroupID, queueType, task.ParentTaskID, task.SchemaID, task.BlueprintID)
		job := byKey[key]
		if job == nil {
			job = &game.ConstructorProductionJob{
				QueueType:      queueType,
				RemainingCount: 0,
				TotalCount:     0,
				RemainingTime:  task.RemainingEnergy,
				TotalTime:      task.TotalEnergy,
				Running:        false,
				ParentJobID:    task.ParentTaskID,
			}
			byKey[key] = job
			keys = append(keys, key)
		}
		if job.ID == 0 || task.ID < job.ID {
			job.ID = task.ID
			job.ConstructorEquipmentGroupID = task.ControllerEquipmentGroupID
			job.SchemaID = task.SchemaID
			job.BlueprintID = task.BlueprintID
		}
		if task.RemainingEnergy < task.TotalEnergy || world.taskHasReserveLocked(task.ID) {
			job.Running = true
			job.RemainingTime = task.RemainingEnergy
		}
		if task.SchemaID > 0 {
			if schema, err := world.itemSchemaLocked(task.SchemaID); err == nil {
				amount := taskCount(&task)
				job.ProductItemModelID = schema.ItemModelID
				job.ProductCount = schema.Count
				job.RemainingCount += schema.Count * remainingTaskCount(task.RemainingEnergy, task.TotalEnergy, amount)
				job.TotalCount += schema.Count * amount
			}
		}
		if task.BlueprintID > 0 {
			if blueprint, err := world.objectBlueprintLocked(task.BlueprintID); err == nil {
				amount := taskCount(&task)
				job.ProductCosmicObjectModelID = blueprint.CosmicObjectModelID
				job.ProductCount = 1
				job.RemainingCount += remainingTaskCount(task.RemainingEnergy, task.TotalEnergy, amount)
				job.TotalCount += amount
			}
		}
	}
	result := make([]game.ConstructorProductionJob, 0, len(keys))
	for _, key := range keys {
		result = append(result, *byKey[key])
	}
	sort.Slice(result, func(left int, right int) bool {
		if result[left].QueueType != result[right].QueueType {
			return result[left].QueueType == "auxiliary"
		}
		return result[left].ID < result[right].ID
	})
	return result
}

// cross2D считает псевдоскалярное произведение двух плоских векторов.
func remainingTaskCount(remainingEnergy float64, totalEnergy float64, amount float64) float64 {
	if amount <= 0 {
		return 1
	}
	if totalEnergy <= physics.Epsilon {
		return amount
	}
	if remainingEnergy <= physics.Epsilon {
		return 0
	}
	completed := math.Floor(((totalEnergy - remainingEnergy) / totalEnergy * amount) + physics.Epsilon)
	remaining := amount - completed
	if remaining < 0 {
		return 0
	}
	return remaining
}

func (world *World) firstPublicDeveloperAssembly(cosmicObjectModelID int64) (*data.Assembly, bool) {
	if world.data.Assemblies == nil {
		return nil, false
	}
	return world.data.Assemblies.FirstPublicDeveloperAssembly(cosmicObjectModelID)
}

// cross2D считает псевдоскалярное произведение двух плоских векторов.
// cross2D считает псевдоскалярное произведение двух плоских векторов.
func (world *World) initializeWeaponMagazinesLocked() {
	if world.data.EquipmentGroups == nil || world.data.ItemModels == nil {
		return
	}
	for _, group := range world.data.EquipmentGroups.Items {
		if group == nil || group.MagazineCount > 0 || group.LastRechargeStartTime > 0 {
			continue
		}
		model, ok := world.data.ItemModels.Get(group.EquipmentItemModelID)
		if !ok {
			continue
		}
		group.MagazineCount = weaponMagazineCapacity(group, model)
	}
}

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

// cross2D считает псевдоскалярное произведение двух плоских векторов.
func (world *World) cosmicObjectFromModelAndAssembly(model *data.CosmicObjectModel, assembly *data.Assembly) *data.CosmicObject {
	cosmicObject := &data.CosmicObject{Enabled: true}
	world.applyModelAndAssembly(cosmicObject, model, assembly)
	cosmicObject.Armor = assembly.MaxArmor
	cosmicObject.Fuel = assembly.MaxFuel
	return cosmicObject
}

// cross2D считает псевдоскалярное произведение двух плоских векторов.
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

// cross2D считает псевдоскалярное произведение двух плоских векторов.
func (world *World) ensureEquipmentFromAssembly(cosmicObjectID int64, assembly *data.Assembly) error {
	if world.data.EquipmentGroups == nil || len(world.data.EquipmentGroups.GetByCosmicObjectID(cosmicObjectID)) > 0 {
		return nil
	}
	return world.installEquipmentFromAssembly(cosmicObjectID, assembly)
}

// cross2D считает псевдоскалярное произведение двух плоских векторов.
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

// cross2D считает псевдоскалярное произведение двух плоских векторов.
func (world *World) installEquipmentFromAssembly(cosmicObjectID int64, assembly *data.Assembly) error {
	if world.data.EquipmentGroups == nil || world.data.AssemblyEquipmentGroups == nil || world.data.ItemModels == nil {
		return nil
	}

	for _, group := range world.data.AssemblyEquipmentGroups.GetByAssemblyID(assembly.ID) {
		equipmentGroup := &data.EquipmentGroup{
			CosmicObjectID:       cosmicObjectID,
			Title:                group.Title,
			EquipmentItemModelID: group.EquipmentItemModelID,
			Count:                group.Count,
			EnabledCount:         group.Count,
			Enabled:              true,
			Active:               true,
		}
		if model, ok := world.data.ItemModels.Get(group.EquipmentItemModelID); ok {
			equipmentGroup.MagazineCount = weaponMagazineCapacity(equipmentGroup, model)
		}
		if _, err := world.data.EquipmentGroups.Add(equipmentGroup); err != nil {
			return err
		}
	}
	return nil
}

// cross2D считает псевдоскалярное произведение двух плоских векторов.
func (world *World) fillShipSupplies(cosmicObject *data.CosmicObject) {
	if cosmicObject == nil {
		return
	}
	cosmicObject.Fuel = cosmicObject.MaxFuel
	if world.data.EquipmentGroups == nil || world.data.ItemGroups == nil || world.data.ItemModels == nil || world.data.ItemTypes == nil {
		return
	}

	containerType, ok := world.data.ItemTypes.GetByAcronym("Container")
	if !ok {
		return
	}

	containerIDs := make([]int64, 0)
	for _, group := range world.data.EquipmentGroups.GetByCosmicObjectID(cosmicObject.ID) {
		model, ok := world.data.ItemModels.Get(group.EquipmentItemModelID)
		if !ok {
			continue
		}
		if model.ItemTypeID == containerType.ID {
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
		_ = world.setItemModelInContainerLocked(containerIDs[0], itemModelID, 10)
	}
	resourceModelIDs := world.resourceItemModelIDs()
	for _, itemModelID := range resourceModelIDs {
		_ = world.setItemModelInContainerLocked(containerIDs[0], itemModelID, 1000)
	}
	ammoModelIDs := world.installedAmmoItemModelIDs(cosmicObject.ID)
	for _, itemModelID := range ammoModelIDs {
		_ = world.setItemModelInContainerLocked(containerIDs[0], itemModelID, 10000)
	}
}

// cross2D считает псевдоскалярное произведение двух плоских векторов.
func (world *World) firstItemModelIDsByType() []int64 {
	firstByType := make(map[int64]int64)
	resourceTypeID := int64(0)
	if resourceType, ok := world.data.ItemTypes.GetByAcronym("Resource"); ok {
		resourceTypeID = resourceType.ID
	}
	ammunitionTypeID := int64(0)
	if ammunitionType, ok := world.data.ItemTypes.GetByAcronym("Ammunition"); ok {
		ammunitionTypeID = ammunitionType.ID
	}
	for itemModelID, model := range world.data.ItemModels.Items {
		if model == nil || model.ItemTypeID <= 0 || model.ItemTypeID == resourceTypeID || model.ItemTypeID == ammunitionTypeID {
			continue
		}
		current, ok := firstByType[model.ItemTypeID]
		if !ok || itemModelID < current {
			firstByType[model.ItemTypeID] = itemModelID
		}
	}

	typeIDs := make([]int64, 0, len(firstByType))
	for ItemTypeID := range firstByType {
		typeIDs = append(typeIDs, ItemTypeID)
	}
	sort.Slice(typeIDs, func(left int, right int) bool {
		return typeIDs[left] < typeIDs[right]
	})

	result := make([]int64, 0, len(typeIDs))
	for _, ItemTypeID := range typeIDs {
		result = append(result, firstByType[ItemTypeID])
	}
	return result
}

// cross2D считает псевдоскалярное произведение двух плоских векторов.
// cross2D считает псевдоскалярное произведение двух плоских векторов.
func (world *World) installedAmmoItemModelIDs(objectID int64) []int64 {
	if world.data.EquipmentGroups == nil || world.data.ItemModels == nil {
		return nil
	}
	seen := map[int64]bool{}
	for _, group := range world.data.EquipmentGroups.GetByCosmicObjectID(objectID) {
		model, ok := world.data.ItemModels.Get(group.EquipmentItemModelID)
		if !ok || model.AmmoItemModelID <= 0 {
			continue
		}
		seen[model.AmmoItemModelID] = true
	}
	result := make([]int64, 0, len(seen))
	for itemModelID := range seen {
		result = append(result, itemModelID)
	}
	sort.Slice(result, func(left int, right int) bool {
		return result[left] < result[right]
	})
	return result
}

func (world *World) resourceItemModelIDs() []int64 {
	resourceType, ok := world.data.ItemTypes.GetByAcronym("Resource")
	if !ok || world.data.ItemModels == nil {
		return nil
	}
	result := make([]int64, 0)
	for itemModelID, model := range world.data.ItemModels.Items {
		if model != nil && model.ItemTypeID == resourceType.ID {
			result = append(result, itemModelID)
		}
	}
	sort.Slice(result, func(left int, right int) bool {
		return result[left] < result[right]
	})
	return result
}

// cross2D считает псевдоскалярное произведение двух плоских векторов.
func raySegmentPolygonDistance(startX float64, startY float64, endX float64, endY float64, object data.CosmicObject, model data.CosmicObjectModel) (float64, bool) {
	points := transformedBodyPolygon(object, model)
	if len(points) < 3 {
		return 0, false
	}
	if pointInsidePolygon(startX, startY, points) {
		return 0, true
	}
	best := math.Inf(1)
	for index := range points {
		nextIndex := (index + 1) % len(points)
		distance, ok := segmentIntersectionDistance(startX, startY, endX, endY, points[index].X, points[index].Y, points[nextIndex].X, points[nextIndex].Y)
		if ok && distance < best {
			best = distance
		}
	}
	if math.IsInf(best, 1) {
		return 0, false
	}
	return best, true
}

// cross2D считает псевдоскалярное произведение двух плоских векторов.
func transformedBodyPolygon(object data.CosmicObject, model data.CosmicObjectModel) []data.BodyPoint {
	points := model.BodyPolygon
	if len(points) == 0 {
		points = fallbackBodyPolygon(model)
	}
	result := make([]data.BodyPoint, 0, len(points))
	cosRotation := math.Cos(object.Rotation)
	sinRotation := math.Sin(object.Rotation)
	for _, point := range points {
		result = append(result, data.BodyPoint{
			X: object.X + point.X*cosRotation + point.Y*sinRotation,
			Y: object.Y - point.X*sinRotation + point.Y*cosRotation,
		})
	}
	return result
}

// cross2D считает псевдоскалярное произведение двух плоских векторов.
func fallbackBodyPolygon(model data.CosmicObjectModel) []data.BodyPoint {
	halfWidth := model.BodyWidth / 2
	halfLength := model.BodyLength / 2
	if halfWidth <= 0 && model.TextureScale > 0 {
		halfWidth = float64(model.TextureBodyWidth) / model.TextureScale / 2
	}
	if halfLength <= 0 && model.TextureScale > 0 {
		halfLength = float64(model.TextureBodyLength) / model.TextureScale / 2
	}
	return []data.BodyPoint{
		{X: -halfWidth, Y: -halfLength},
		{X: halfWidth, Y: -halfLength},
		{X: halfWidth, Y: halfLength},
		{X: -halfWidth, Y: halfLength},
	}
}

// cross2D считает псевдоскалярное произведение двух плоских векторов.
func pointInsidePolygon(x float64, y float64, points []data.BodyPoint) bool {
	inside := false
	for index, point := range points {
		previous := points[(index+len(points)-1)%len(points)]
		if ((point.Y > y) != (previous.Y > y)) &&
			(x < (previous.X-point.X)*(y-point.Y)/(previous.Y-point.Y)+point.X) {
			inside = !inside
		}
	}
	return inside
}

// cross2D считает псевдоскалярное произведение двух плоских векторов.
func segmentIntersectionDistance(ax float64, ay float64, bx float64, by float64, cx float64, cy float64, dx float64, dy float64) (float64, bool) {
	rx := bx - ax
	ry := by - ay
	sx := dx - cx
	sy := dy - cy
	denominator := cross2D(rx, ry, sx, sy)
	if math.Abs(denominator) <= physics.Epsilon {
		return 0, false
	}
	qpx := cx - ax
	qpy := cy - ay
	t := cross2D(qpx, qpy, sx, sy) / denominator
	u := cross2D(qpx, qpy, rx, ry) / denominator
	if t < -physics.Epsilon || t > 1+physics.Epsilon || u < -physics.Epsilon || u > 1+physics.Epsilon {
		return 0, false
	}
	return math.Hypot(rx, ry) * math.Max(0, math.Min(1, t)), true
}

// cross2D считает псевдоскалярное произведение двух плоских векторов.
func cross2D(ax float64, ay float64, bx float64, by float64) float64 {
	return ax*by - ay*bx
}
