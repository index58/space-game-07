package world_test

import (
	"fmt"
	"math"
	"os"
	"path/filepath"
	"testing"

	"space-game-07-server/internal/data"
	"space-game-07-server/internal/game"
	"space-game-07-server/internal/physics"
	"space-game-07-server/internal/world"
)

// Ищет объект в снимке мира по ID, чтобы тесты не зависели от порядка массива.
func findCosmicObjectInSnapshot(snapshot game.Snapshot, objectID int64) (data.CosmicObject, bool) {
	for _, object := range snapshot.Objects {
		if object.ID == objectID {
			return object, true
		}
	}

	return data.CosmicObject{}, false
}

// Сравнивает дробные значения с малым допуском для проверок симуляции.
func closeWorldFloat(t *testing.T, actual float64, expected float64) {
	t.Helper()

	if math.Abs(actual-expected) > 0.000001 {
		t.Fatalf("got %v, want %v", actual, expected)
	}
}

// Возвращает указатель на логическое значение для проверки частичных команд.
func boolPointer(value bool) *bool {
	return &value
}

// Возвращает указатель на строку для проверки частичных команд.
func stringPointer(value string) *string {
	return &value
}

// Возвращает указатель на целое значение для проверки частичных команд.
func int64Pointer(value int64) *int64 {
	return &value
}

// Ищет группу оборудования в снимке мира по ID, чтобы тесты не зависели от порядка массива.
func findEquipmentGroupInSnapshot(snapshot game.Snapshot, groupID int64) (data.EquipmentGroup, bool) {
	for _, group := range snapshot.EquipmentGroups {
		if group.ID == groupID {
			return group, true
		}
	}

	return data.EquipmentGroup{}, false
}

// Собирает минимальный игровой мир с кораблем, астероидом и станцией.
// Устанавливает тестовый генератор на указанный объект.
func addTestGenerator(t *testing.T, serverData world.Data, cosmicObjectID int64) *data.EquipmentGroup {
	t.Helper()

	serverData.ItemModels.Items[104] = &data.ItemModel{
		ID:                   104,
		TitleRu:              "Generator",
		TitleEn:              "Generator",
		Acronym:              "Generator",
		ItemtypeID:           1,
		GeneratingPower:      60000,
		ConsumingItemModelID: 7,
		ConsumingCount:       2,
	}
	if err := serverData.ItemModels.RebuildIndexes(); err != nil {
		t.Fatal(err)
	}
	group, err := serverData.EquipmentGroups.Add(&data.EquipmentGroup{
		CosmicObjectID:       cosmicObjectID,
		Title:                "Generator",
		EquipmentItemModelID: 104,
		Count:                1,
		EnabledCount:         1,
		Enabled:              true,
	})
	if err != nil {
		t.Fatal(err)
	}
	return group
}

// Устанавливает тестовый источник электроэнергии без расхода топлива.
func addTestPowerProducer(t *testing.T, serverData world.Data, cosmicObjectID int64) *data.EquipmentGroup {
	t.Helper()

	serverData.ItemModels.Items[105] = &data.ItemModel{
		ID:              105,
		TitleRu:         "Power Producer",
		TitleEn:         "Power Producer",
		Acronym:         "PowerProducer",
		ItemtypeID:      1,
		GeneratingPower: 60000,
	}
	if err := serverData.ItemModels.RebuildIndexes(); err != nil {
		t.Fatal(err)
	}
	group, err := serverData.EquipmentGroups.Add(&data.EquipmentGroup{
		CosmicObjectID:       cosmicObjectID,
		Title:                "Power Producer",
		EquipmentItemModelID: 105,
		Count:                1,
		EnabledCount:         1,
		Enabled:              true,
	})
	if err != nil {
		t.Fatal(err)
	}
	return group
}

// Проверяет, что команда панели управления меняет только объект текущего аккаунта и подтверждает мутацию.
func TestApplyControlPanelObjectUpdateChangesControlledObjectAndAcknowledgesMutation(t *testing.T) {
	serverData := testWorldData(t)
	gameWorld := world.New(1, serverData)
	if _, ok := gameWorld.ConnectAccount(1); !ok {
		t.Fatal("account was not connected")
	}

	err := gameWorld.ApplyControlPanelObjectUpdate(1, "session-1", 4, world.ControlPanelObjectUpdate{
		Enabled: boolPointer(false),
		Title:   stringPointer("Новый корабль"),
	})

	if err != nil {
		t.Fatalf("apply object update: %v", err)
	}
	snapshot := gameWorld.SnapshotForAccount(1)
	object, ok := findCosmicObjectInSnapshot(snapshot, 1)
	if !ok {
		t.Fatal("controlled object not found")
	}
	if object.Enabled || object.Title != "Новый корабль" {
		t.Fatalf("object was not updated: %+v", object)
	}
	ack := gameWorld.ClientMutationAck(1, "session-1")
	if ack.SessionID != "session-1" || ack.LastAppliedSeq != 4 {
		t.Fatalf("mutation ack mismatch: %+v", ack)
	}
}

// Проверяет, что команда панели управления оборудованием ограничена оборудованием текущего объекта.
func TestApplyControlPanelEquipmentUpdateChangesOwnedGroupOnly(t *testing.T) {
	serverData := testWorldData(t)
	gameWorld := world.New(1, serverData)
	if _, ok := gameWorld.ConnectAccount(1); !ok {
		t.Fatal("account was not connected")
	}
	ownedGroup := serverData.EquipmentGroups.GetByCosmicObjectID(1)[0]
	foreignGroup, err := serverData.EquipmentGroups.Add(&data.EquipmentGroup{
		CosmicObjectID:       2,
		Title:                "Foreign",
		EquipmentItemModelID: ownedGroup.EquipmentItemModelID,
		Count:                1,
		EnabledCount:         1,
		Enabled:              true,
	})
	if err != nil {
		t.Fatal(err)
	}

	err = gameWorld.ApplyControlPanelEquipmentUpdate(1, "session-1", 5, world.ControlPanelEquipmentUpdate{
		EquipmentGroupID: ownedGroup.ID,
		Enabled:          boolPointer(false),
		EnabledCount:     int64Pointer(1),
	})

	if err != nil {
		t.Fatalf("apply equipment update: %v", err)
	}
	if ownedGroup.Enabled || ownedGroup.EnabledCount != 1 {
		t.Fatalf("owned group was not updated: %+v", ownedGroup)
	}
	if err := gameWorld.ApplyControlPanelEquipmentUpdate(1, "session-1", 6, world.ControlPanelEquipmentUpdate{EquipmentGroupID: foreignGroup.ID, Enabled: boolPointer(false)}); err == nil {
		t.Fatal("foreign group update was accepted")
	}
	snapshot := gameWorld.SnapshotForAccount(1)
	group, ok := findEquipmentGroupInSnapshot(snapshot, ownedGroup.ID)
	if !ok || group.Enabled || group.EnabledCount != 1 {
		t.Fatalf("snapshot group mismatch: %+v", group)
	}
}

// Собирает минимальный игровой мир с кораблем, астероидом и станцией.
func testWorldData(t *testing.T) world.Data {
	t.Helper()

	accounts := &data.Accounts{
		MaxID: 1,
		Items: map[int64]*data.Account{
			1: {ID: 1, Email: "index@email.net", Nickname: "index", PasswordHash: "hash", Token: "token", CurrentCharacterID: 1},
		},
	}
	characters := &data.Characters{
		MaxID: 1,
		Items: map[int64]*data.Character{
			1: {ID: 1, AccountID: 1, LocationCosmicObjectID: 1},
		},
	}
	cosmicObjectTypes := &data.CosmicObjectTypes{
		MaxID: 3,
		Items: map[int64]*data.CosmicObjectType{
			1: {ID: 1, TitleRu: "Корабль", TitleEn: "Ship", Acronym: "Ship"},
			2: {ID: 2, TitleRu: "Станция", TitleEn: "Station", Acronym: "Station"},
			3: {ID: 3, TitleRu: "Астероид", TitleEn: "Asteroid", Acronym: "Asteroid"},
		},
	}
	cosmicObjectModels := &data.CosmicObjectModels{
		MaxID: 4,
		Items: map[int64]*data.CosmicObjectModel{
			1: {ID: 1, TitleRu: "Корабль", TitleEn: "Ship", Acronym: "ship_bat", CosmicObjectTypeID: 1, TextureScale: 4, TextureBodyWidth: 27, TextureBodyLength: 63, Mass: 7.92, Capacity: 10, MaxArmor: 100, MaxSpeed: 497, MaxAngularSpeed: 3},
			2: {ID: 2, TitleRu: "Астероид", TitleEn: "Asteroid", Acronym: "asteroid_0002", CosmicObjectTypeID: 3, TextureScale: 4, TextureBodyWidth: 30, TextureBodyLength: 30},
			3: {ID: 3, TitleRu: "Станция", TitleEn: "Station", Acronym: "station_tiny_crumb", CosmicObjectTypeID: 2, TextureScale: 4, TextureBodyWidth: 30, TextureBodyLength: 30},
			4: {ID: 4, TitleRu: "Новый корабль", TitleEn: "New Ship", Acronym: "ship_new", CosmicObjectTypeID: 1, TextureScale: 4, TextureBodyWidth: 40, TextureBodyLength: 80, Mass: 12, Capacity: 20, MaxArmor: 150, MaxSpeed: 600, MaxAngularSpeed: 4},
		},
	}
	cosmicObjects := &data.CosmicObjects{
		MaxID: 3,
		Items: map[int64]*data.CosmicObject{
			1: {ID: 1, Title: "Ship", CosmicObjectModelID: 1, OwnerCharacterID: 1, Mass: 7.92, MaxSpeed: 497, MaxAngularSpeed: 3, MaxAlongForce: 1287901.529888449, MaxTorque: 653565, Enabled: true},
			2: {ID: 2, Title: "Asteroid", CosmicObjectModelID: 2, X: -500, Y: 800, Mass: 629.532, MaxSpeed: 475, MaxAngularSpeed: 3, Enabled: true, Anchored: true},
			3: {ID: 3, Title: "Station", CosmicObjectModelID: 3, X: 500, Y: 500, Mass: 185.625, MaxSpeed: 486, MaxAngularSpeed: 3, Enabled: true, Anchored: true},
		},
	}
	assemblies := &data.Assemblies{
		MaxID: 3,
		Items: map[int64]*data.Assembly{
			1: {ID: 1, AuthorCharacterID: 1, Title: "Private Ship", CosmicObjectModelID: 1, IsPublic: true, Mass: 999, MaxArmor: 999, MaxAlongForce: 999, MaxAcrossForce: 999, MaxTorque: 999},
			2: {ID: 2, AuthorCharacterID: 0, Title: "Default Ship", CosmicObjectModelID: 1, IsPublic: true, Mass: 900, MaxArmor: 300, MaxAlongForce: 90000, MaxAcrossForce: 80000, MaxTorque: 70000, GeneratingPower: 10, ConsumingPower: 20, Complexity: 30, OccupiedVolume: 40, MaxFuel: 50},
			3: {ID: 3, AuthorCharacterID: 0, Title: "Default New Ship", CosmicObjectModelID: 4, IsPublic: true, Mass: 1200, MaxArmor: 400, MaxAlongForce: 120000, MaxAcrossForce: 110000, MaxTorque: 100000, GeneratingPower: 11, ConsumingPower: 21, Complexity: 31, OccupiedVolume: 41, MaxFuel: 51},
		},
	}
	assemblyEquipmentGroups := &data.AssemblyEquipmentGroups{
		MaxID: 7,
		Items: map[int64]*data.AssemblyEquipmentGroup{
			1: {ID: 1, AssemblyID: 2, Title: "Thrusters", EquipmentItemModelID: 101, Count: 2},
			2: {ID: 2, AssemblyID: 2, Title: "Torquers", EquipmentItemModelID: 102, Count: 1},
			3: {ID: 3, AssemblyID: 2, Title: "Containers", EquipmentItemModelID: 301, Count: 2},
			4: {ID: 4, AssemblyID: 2, Title: "Cannons", EquipmentItemModelID: 302, Count: 3},
			5: {ID: 5, AssemblyID: 3, Title: "New Thrusters", EquipmentItemModelID: 201, Count: 4},
			6: {ID: 6, AssemblyID: 3, Title: "New Containers", EquipmentItemModelID: 401, Count: 1},
			7: {ID: 7, AssemblyID: 3, Title: "New Cannons", EquipmentItemModelID: 402, Count: 4},
		},
	}
	itemtypes := &data.Itemtypes{
		MaxID: 18,
		Items: map[int64]*data.Itemtype{
			1:  {ID: 1, TitleRu: "Weapon", TitleEn: "Weapon", Acronym: "Weapon", CountMustBeInteger: true},
			7:  {ID: 7, TitleRu: "Fuel", TitleEn: "Fuel", Acronym: "Fuel"},
			9:  {ID: 9, TitleRu: "Container", TitleEn: "Container", Acronym: "Container", CountMustBeInteger: true},
			18: {ID: 18, TitleRu: "Ammunition", TitleEn: "Ammunition", Acronym: "Ammunition", CountMustBeInteger: true},
		},
	}
	itemModels := &data.ItemModels{
		MaxID: 403,
		Items: map[int64]*data.ItemModel{
			101: {ID: 101, TitleRu: "Thruster", TitleEn: "Thruster", Acronym: "Thruster", ItemtypeID: 1, ConsumingPower: 15000, ConsumingItemModelID: 7, ConsumingCount: 0.5, MaxAlongForce: 45000, MaxAcrossForce: 40000},
			102: {ID: 102, TitleRu: "Torquer", TitleEn: "Torquer", Acronym: "Torquer", ItemtypeID: 1, ConsumingPower: 5000, ConsumingItemModelID: 7, ConsumingCount: 0.25, MaxTorque: 70000},
			201: {ID: 201, TitleRu: "New Thruster", TitleEn: "New Thruster", Acronym: "NewThruster", ItemtypeID: 1, ConsumingPower: 10000, ConsumingItemModelID: 7, ConsumingCount: 0.4, MaxAlongForce: 30000, MaxAcrossForce: 27500},
			301: {ID: 301, TitleRu: "Container", TitleEn: "Container", Acronym: "Container", ItemtypeID: 9},
			302: {ID: 302, TitleRu: "Cannon", TitleEn: "Cannon", Acronym: "Cannon", ItemtypeID: 1, AmmoItemModelID: 303, FiringRate: 0.5},
			303: {ID: 303, TitleRu: "Shell", TitleEn: "Shell", Acronym: "Shell", ItemtypeID: 18},
			401: {ID: 401, TitleRu: "New Container", TitleEn: "New Container", Acronym: "NewContainer", ItemtypeID: 9},
			402: {ID: 402, TitleRu: "New Cannon", TitleEn: "New Cannon", Acronym: "NewCannon", ItemtypeID: 1, AmmoItemModelID: 403, FiringRate: 1},
			403: {ID: 403, TitleRu: "New Shell", TitleEn: "New Shell", Acronym: "NewShell", ItemtypeID: 18},
		},
	}
	equipmentGroups := data.NewEquipmentGroups()
	itemGroups := data.NewItemGroups()

	if err := accounts.RebuildIndexes(); err != nil {
		t.Fatal(err)
	}
	if err := characters.RebuildIndexes(); err != nil {
		t.Fatal(err)
	}
	if err := cosmicObjectTypes.RebuildIndexes(); err != nil {
		t.Fatal(err)
	}
	if err := cosmicObjectModels.RebuildIndexes(); err != nil {
		t.Fatal(err)
	}
	if err := cosmicObjects.RebuildIndexes(); err != nil {
		t.Fatal(err)
	}
	if err := assemblies.RebuildIndexes(); err != nil {
		t.Fatal(err)
	}
	if err := assemblyEquipmentGroups.RebuildIndexes(); err != nil {
		t.Fatal(err)
	}
	if err := itemtypes.RebuildIndexes(); err != nil {
		t.Fatal(err)
	}
	if err := itemModels.RebuildIndexes(); err != nil {
		t.Fatal(err)
	}

	return world.Data{
		Accounts:                accounts,
		Characters:              characters,
		CosmicObjects:           cosmicObjects,
		CosmicObjectTypes:       cosmicObjectTypes,
		CosmicObjectModels:      cosmicObjectModels,
		Itemtypes:               itemtypes,
		ItemModels:              itemModels,
		EquipmentGroups:         equipmentGroups,
		ItemGroups:              itemGroups,
		Assemblies:              assemblies,
		AssemblyEquipmentGroups: assemblyEquipmentGroups,
	}
}

// Проверяет, что подключение аккаунта берёт управляемый объект из текущего местоположения персонажа.
func TestConnectAccountUsesCurrentCharacterLocation(t *testing.T) {
	gameWorld := world.New(1, testWorldData(t))

	objectID, ok := gameWorld.ConnectAccount(1)
	if !ok {
		t.Fatalf("account was not connected")
	}

	if objectID != 1 {
		t.Fatalf("got object ID %v, want 1", objectID)
	}
}

// Проверяет, что подключение аккаунта не оставляет кораблю устаревший целевой угол поворота.
func TestConnectAccountAlignsTargetRotationWithCurrentRotation(t *testing.T) {
	serverData := testWorldData(t)
	serverData.CosmicObjects.Items[1].Rotation = 1.75
	serverData.CosmicObjects.Items[1].TargetRotation = 0
	gameWorld := world.New(1, serverData)

	if _, ok := gameWorld.ConnectAccount(1); !ok {
		t.Fatalf("account was not connected")
	}
	object, ok := serverData.CosmicObjects.Get(1)
	if !ok {
		t.Fatalf("object was not found")
	}

	if object.TargetRotation != object.Rotation {
		t.Fatalf("target rotation = %v, want current rotation %v", object.TargetRotation, object.Rotation)
	}
}

// Проверяет, что новый мир создает серверный чат и показывает его подключенному аккаунту.
func TestChatStateIncludesServerChatAfterConnect(t *testing.T) {
	gameWorld := world.New(1, testWorldData(t))
	if _, ok := gameWorld.ConnectAccount(1); !ok {
		t.Fatalf("account was not connected")
	}

	chatState, ok := gameWorld.ChatStateForAccount(1, 0)
	if !ok {
		t.Fatalf("chat state is not available")
	}
	if len(chatState.Tabs) != 1 {
		t.Fatalf("chat tabs count = %d, want 1", len(chatState.Tabs))
	}
	if chatState.Tabs[0].CommunityTypeAcronym != "Server" {
		t.Fatalf("first chat type = %q, want Server", chatState.Tabs[0].CommunityTypeAcronym)
	}
}

// Проверяет, что сообщение в серверный чат сохраняется от имени текущего персонажа.
func TestSendServerChatMessageStoresMessageFromCurrentCharacter(t *testing.T) {
	gameWorld := world.New(1, testWorldData(t))
	if _, ok := gameWorld.ConnectAccount(1); !ok {
		t.Fatalf("account was not connected")
	}
	chatState, ok := gameWorld.ChatStateForAccount(1, 0)
	if !ok {
		t.Fatalf("chat state is not available")
	}

	nextState, recipients, chatError := gameWorld.SendChatMessage(1, chatState.SelectedChatID, "", "Привет всем")
	if chatError != "" {
		t.Fatalf("SendChatMessage returned error: %s", chatError)
	}
	if len(recipients) != 1 || recipients[0] != 1 {
		t.Fatalf("recipients = %v, want [1]", recipients)
	}
	if len(nextState.Tabs[0].Messages) != 1 {
		t.Fatalf("message count = %d, want 1", len(nextState.Tabs[0].Messages))
	}
	message := nextState.Tabs[0].Messages[0]
	if message.Text != "Привет всем" || message.SenderCharacterID != 1 || message.SenderNickname != "index" {
		t.Fatalf("message = %+v, want text from current account", message)
	}
}

// Проверяет, что адресация через ник аккаунта создает дуэтный чат с ключом из ID персонажей.
func TestSendDuoChatMessageUsesAccountNicknameAndDuoKey(t *testing.T) {
	serverData := testWorldData(t)
	serverData.Accounts.Items[2] = &data.Account{ID: 2, Email: "pilot2@email.net", Nickname: "Pilot2", PasswordHash: "hash", Token: "token-2", CurrentCharacterID: 2}
	serverData.Characters.Items[2] = &data.Character{ID: 2, AccountID: 2, LocationCosmicObjectID: 3}
	if err := serverData.Accounts.RebuildIndexes(); err != nil {
		t.Fatal(err)
	}
	if err := serverData.Characters.RebuildIndexes(); err != nil {
		t.Fatal(err)
	}
	gameWorld := world.New(1, serverData)
	if _, ok := gameWorld.ConnectAccount(1); !ok {
		t.Fatalf("first account was not connected")
	}
	if _, ok := gameWorld.ConnectAccount(2); !ok {
		t.Fatalf("second account was not connected")
	}

	chatState, recipients, chatError := gameWorld.SendChatMessage(1, 0, "Pilot2", "Личное сообщение")
	if chatError != "" {
		t.Fatalf("SendChatMessage returned error: %s", chatError)
	}
	if len(recipients) != 2 || recipients[0] != 1 || recipients[1] != 2 {
		t.Fatalf("recipients = %v, want [1 2]", recipients)
	}

	var duoTab game.ChatTab
	for _, tab := range chatState.Tabs {
		if tab.CommunityTypeAcronym == "Duo" {
			duoTab = tab
		}
	}
	if duoTab.ChatID == 0 {
		t.Fatal("duo tab was not opened")
	}
	if duoTab.Title != "Pilot2" || duoTab.DuoChatKey != "1:2" {
		t.Fatalf("duo tab = %+v, want Pilot2 with key 1:2", duoTab)
	}
}

// Проверяет, что входящий дуэт не меняет выбранный чат получателя и учитывается как непрочитанный.
func TestIncomingDuoMessageKeepsRecipientSelectionAndUpdatesUnreadCount(t *testing.T) {
	serverData := testWorldData(t)
	serverData.Accounts.Items[2] = &data.Account{ID: 2, Email: "pilot2@email.net", Nickname: "Pilot2", PasswordHash: "hash", Token: "token-2", CurrentCharacterID: 2}
	serverData.Characters.Items[2] = &data.Character{ID: 2, AccountID: 2, LocationCosmicObjectID: 3}
	if err := serverData.Accounts.RebuildIndexes(); err != nil {
		t.Fatal(err)
	}
	if err := serverData.Characters.RebuildIndexes(); err != nil {
		t.Fatal(err)
	}
	gameWorld := world.New(1, serverData)
	if _, ok := gameWorld.ConnectAccount(1); !ok {
		t.Fatalf("sender account was not connected")
	}
	if _, ok := gameWorld.ConnectAccount(2); !ok {
		t.Fatalf("recipient account was not connected")
	}
	recipientState, ok := gameWorld.ChatStateForAccount(2, 0)
	if !ok {
		t.Fatalf("recipient chat state is not available")
	}
	serverChatID := recipientState.SelectedChatID

	if _, _, chatError := gameWorld.SendChatMessage(1, 0, "Pilot2", "Личное сообщение"); chatError != "" {
		t.Fatalf("SendChatMessage returned error: %s", chatError)
	}
	recipientState, ok = gameWorld.ChatStateForAccount(2, serverChatID)
	if !ok {
		t.Fatalf("recipient chat state is not available after incoming message")
	}

	if recipientState.SelectedChatID != serverChatID {
		t.Fatalf("selected chat = %d, want server chat %d", recipientState.SelectedChatID, serverChatID)
	}
	var duoTab *game.ChatTab
	for index := range recipientState.Tabs {
		if recipientState.Tabs[index].CommunityTypeAcronym == "Duo" {
			duoTab = &recipientState.Tabs[index]
		}
	}
	if duoTab == nil {
		t.Fatalf("recipient duo tab was not opened")
	}
	if duoTab.UnreadCount != 1 {
		t.Fatalf("duo unread count = %d, want 1", duoTab.UnreadCount)
	}

	readState, ok := gameWorld.ChatStateForAccount(2, duoTab.ChatID)
	if !ok || readState.SelectedChatID != duoTab.ChatID {
		t.Fatalf("recipient could not select duo chat")
	}
	var readDuoTab *game.ChatTab
	for index := range readState.Tabs {
		if readState.Tabs[index].ChatID == duoTab.ChatID {
			readDuoTab = &readState.Tabs[index]
		}
	}
	if readDuoTab == nil || readDuoTab.UnreadCount != 0 {
		t.Fatalf("duo unread count after selection = %+v, want 0", readDuoTab)
	}
}

// Проверяет, что неизвестный ник аккаунта дает ошибку без создания сообщения.
func TestSendDuoChatMessageRejectsUnknownAccountNickname(t *testing.T) {
	gameWorld := world.New(1, testWorldData(t))
	if _, ok := gameWorld.ConnectAccount(1); !ok {
		t.Fatalf("account was not connected")
	}

	_, recipients, chatError := gameWorld.SendChatMessage(1, 0, "Nobody", "Личное сообщение")
	if chatError == "" {
		t.Fatal("unknown nickname was accepted")
	}
	if len(recipients) != 0 {
		t.Fatalf("recipients = %v, want empty list", recipients)
	}
}

// Проверяет, что состояние чата отдает всю доступную историю сообщений.
func TestChatStateIncludesFullMessageHistory(t *testing.T) {
	gameWorld := world.New(1, testWorldData(t))
	if _, ok := gameWorld.ConnectAccount(1); !ok {
		t.Fatalf("account was not connected")
	}
	chatState, ok := gameWorld.ChatStateForAccount(1, 0)
	if !ok {
		t.Fatalf("chat state is not available")
	}
	for index := 1; index <= 12; index++ {
		if _, _, chatError := gameWorld.SendChatMessage(1, chatState.SelectedChatID, "", fmt.Sprintf("msg-%02d", index)); chatError != "" {
			t.Fatalf("SendChatMessage returned error: %s", chatError)
		}
	}

	chatState, ok = gameWorld.ChatStateForAccount(1, chatState.SelectedChatID)
	if !ok {
		t.Fatalf("chat state is not available")
	}
	messages := chatState.Tabs[0].Messages
	if len(messages) != 12 {
		t.Fatalf("message count = %d, want 12", len(messages))
	}
	if messages[0].Text != "msg-01" || messages[11].Text != "msg-12" {
		t.Fatalf("message range = %q..%q, want msg-01..msg-12", messages[0].Text, messages[11].Text)
	}
}

// Проверяет, что ввод подключённого аккаунта двигает уже существующий корабль и сохраняет новое положение.
func TestTickAppliesAccountInputToExistingShip(t *testing.T) {
	serverData := testWorldData(t)
	serverData.CosmicObjects.Items[1].Fuel = 50
	gameWorld := world.New(1, serverData)

	objectID, ok := gameWorld.ConnectAccount(1)
	if !ok {
		t.Fatalf("account was not connected")
	}
	addTestPowerProducer(t, serverData, objectID)
	gameWorld.SetInput(1, game.ShipInput{ThrustForward: true})
	snapshot := gameWorld.Tick(1.0 / 30.0)
	object, ok := findCosmicObjectInSnapshot(snapshot, objectID)
	if !ok {
		t.Fatalf("object %v not found in snapshot", objectID)
	}

	if object.VelocityY <= 0 {
		t.Fatalf("got velocity Y %v, want positive", object.VelocityY)
	}
	if serverData.CosmicObjects.Items[1].Y <= 0 {
		t.Fatalf("got stored Y %v, want positive", serverData.CosmicObjects.Items[1].Y)
	}
}

// Проверяет, что выключенные двигатели не создают тягу и не двигают корабль.
func TestTickDoesNotApplyThrustFromDisabledEngines(t *testing.T) {
	serverData := testWorldData(t)
	gameWorld := world.New(1, serverData)

	objectID, ok := gameWorld.ConnectAccount(1)
	if !ok {
		t.Fatalf("account was not connected")
	}
	serverData.CosmicObjects.Items[objectID].Fuel = 50
	for _, group := range serverData.EquipmentGroups.GetByCosmicObjectID(objectID) {
		group.Enabled = false
	}

	gameWorld.SetInput(1, game.ShipInput{ThrustForward: true, ThrustRight: true, TargetRotationDelta: 1})
	snapshot := gameWorld.Tick(1)
	object, ok := findCosmicObjectInSnapshot(snapshot, objectID)
	if !ok {
		t.Fatalf("object %v not found in snapshot", objectID)
	}

	if object.VelocityX != 0 || object.VelocityY != 0 || object.AlongForce != 0 || object.AcrossForce != 0 || object.Torque != 0 {
		t.Fatalf("disabled engines affected movement: %+v", object)
	}
}

// Проверяет, что двигатели без выработки электроэнергии не создают тягу.
func TestTickDoesNotApplyThrustWithoutGeneratedPower(t *testing.T) {
	serverData := testWorldData(t)
	gameWorld := world.New(1, serverData)

	objectID, ok := gameWorld.ConnectAccount(1)
	if !ok {
		t.Fatalf("account was not connected")
	}
	serverData.CosmicObjects.Items[objectID].Fuel = 50

	gameWorld.SetInput(1, game.ShipInput{ThrustForward: true})
	snapshot := gameWorld.Tick(1)
	object, ok := findCosmicObjectInSnapshot(snapshot, objectID)
	if !ok {
		t.Fatalf("object %v not found in snapshot", objectID)
	}

	if object.VelocityY != 0 || object.AlongForce != 0 || object.ConsumingPower != 0 {
		t.Fatalf("engines worked without generated power: %+v", object)
	}
}

// Проверяет, что одноразовая команда ввода переключает якорь управляемого объекта.
func TestSetInputTogglesControlledShipAnchor(t *testing.T) {
	serverData := testWorldData(t)
	gameWorld := world.New(1, serverData)

	objectID, ok := gameWorld.ConnectAccount(1)
	if !ok {
		t.Fatalf("account was not connected")
	}
	gameWorld.SetInput(1, game.ShipInput{ToggleAnchor: true})
	firstSnapshot := gameWorld.SnapshotForAccount(1)
	firstObject, ok := findCosmicObjectInSnapshot(firstSnapshot, objectID)
	if !ok {
		t.Fatalf("object %v not found in snapshot", objectID)
	}
	if !firstObject.Anchored {
		t.Fatalf("anchor was not enabled: %+v", firstObject)
	}

	gameWorld.SetInput(1, game.ShipInput{})
	secondSnapshot := gameWorld.SnapshotForAccount(1)
	secondObject, ok := findCosmicObjectInSnapshot(secondSnapshot, objectID)
	if !ok {
		t.Fatalf("object %v not found in snapshot", objectID)
	}
	if !secondObject.Anchored {
		t.Fatalf("anchor changed without toggle command: %+v", secondObject)
	}

	gameWorld.SetInput(1, game.ShipInput{ToggleAnchor: true})
	thirdSnapshot := gameWorld.SnapshotForAccount(1)
	thirdObject, ok := findCosmicObjectInSnapshot(thirdSnapshot, objectID)
	if !ok {
		t.Fatalf("object %v not found in snapshot", objectID)
	}
	if thirdObject.Anchored {
		t.Fatalf("anchor was not disabled: %+v", thirdObject)
	}
}

// Проверяет, что якорь нельзя включить до полной остановки объекта.
func TestSetInputDoesNotEnableAnchorUntilControlledObjectStops(t *testing.T) {
	serverData := testWorldData(t)
	serverData.CosmicObjects.Items[1].VelocityX = 0.01
	serverData.CosmicObjects.Items[1].AngularSpeed = 0.01
	gameWorld := world.New(1, serverData)

	objectID, ok := gameWorld.ConnectAccount(1)
	if !ok {
		t.Fatalf("account was not connected")
	}
	gameWorld.SetInput(1, game.ShipInput{ToggleAnchor: true})
	firstSnapshot := gameWorld.SnapshotForAccount(1)
	firstObject, ok := findCosmicObjectInSnapshot(firstSnapshot, objectID)
	if !ok {
		t.Fatalf("object %v not found in snapshot", objectID)
	}
	if firstObject.Anchored {
		t.Fatalf("moving object enabled anchor: %+v", firstObject)
	}

	serverData.CosmicObjects.Items[1].VelocityX = 0
	serverData.CosmicObjects.Items[1].VelocityY = 0
	serverData.CosmicObjects.Items[1].Speed = 0
	serverData.CosmicObjects.Items[1].AngularSpeed = 0
	gameWorld.SetInput(1, game.ShipInput{ToggleAnchor: true})
	secondSnapshot := gameWorld.SnapshotForAccount(1)
	secondObject, ok := findCosmicObjectInSnapshot(secondSnapshot, objectID)
	if !ok {
		t.Fatalf("object %v not found in snapshot", objectID)
	}
	if !secondObject.Anchored {
		t.Fatalf("stopped object did not enable anchor: %+v", secondObject)
	}

	serverData.CosmicObjects.Items[1].VelocityX = 0.01
	gameWorld.SetInput(1, game.ShipInput{ToggleAnchor: true})
	thirdSnapshot := gameWorld.SnapshotForAccount(1)
	thirdObject, ok := findCosmicObjectInSnapshot(thirdSnapshot, objectID)
	if !ok {
		t.Fatalf("object %v not found in snapshot", objectID)
	}
	if thirdObject.Anchored {
		t.Fatalf("active anchor was not disabled while moving: %+v", thirdObject)
	}
}

// Проверяет, что тяга включает нужное оборудование, обновляет потребление энергии и тратит топливо.
func TestTickUpdatesActiveEquipmentPowerAndFuelFromShipInput(t *testing.T) {
	serverData := testWorldData(t)
	gameWorld := world.New(1, serverData)

	objectID, ok := gameWorld.ConnectAccount(1)
	if !ok {
		t.Fatalf("account was not connected")
	}
	serverData.CosmicObjects.Items[objectID].Fuel = 50
	addTestPowerProducer(t, serverData, objectID)

	gameWorld.SetInput(1, game.ShipInput{ThrustForward: true})
	gameWorld.Tick(2)

	cosmicObject := serverData.CosmicObjects.Items[objectID]
	if cosmicObject.ConsumingPower != 30000 {
		t.Fatalf("consuming power = %v, want 30000", cosmicObject.ConsumingPower)
	}
	if cosmicObject.GeneratingPower != 60000 {
		t.Fatalf("generating power = %v, want 60000", cosmicObject.GeneratingPower)
	}
	closeWorldFloat(t, cosmicObject.Fuel, 48)

	installed := serverData.EquipmentGroups.GetByCosmicObjectID(objectID)
	if !installed[0].Active {
		t.Fatalf("thrusters must be active while creating thrust: %+v", installed[0])
	}
	if installed[1].Active {
		t.Fatalf("torquer must not be active without target rotation delta: %+v", installed[1])
	}
}

// Проверяет, что без ввода двигатели не потребляют энергию и не расходуют топливо.
func TestTickDisablesEngineConsumptionWhenInputStops(t *testing.T) {
	serverData := testWorldData(t)
	gameWorld := world.New(1, serverData)

	objectID, ok := gameWorld.ConnectAccount(1)
	if !ok {
		t.Fatalf("account was not connected")
	}
	serverData.CosmicObjects.Items[objectID].Fuel = 50

	gameWorld.SetInput(1, game.ShipInput{})
	gameWorld.Tick(2)

	cosmicObject := serverData.CosmicObjects.Items[objectID]
	if cosmicObject.ConsumingPower != 0 {
		t.Fatalf("consuming power = %v, want 0", cosmicObject.ConsumingPower)
	}
	closeWorldFloat(t, cosmicObject.Fuel, 50)

	installed := serverData.EquipmentGroups.GetByCosmicObjectID(objectID)
	if installed[0].Active || installed[1].Active {
		t.Fatalf("engine equipment must be inactive without thrust or torque: %+v %+v", installed[0], installed[1])
	}
}

// Проверяет, что генератор без потребителей не активируется и не тратит топливо.
func TestTickDoesNotSpendGeneratorFuelWithoutPowerDemand(t *testing.T) {
	serverData := testWorldData(t)
	gameWorld := world.New(1, serverData)

	objectID, ok := gameWorld.ConnectAccount(1)
	if !ok {
		t.Fatalf("account was not connected")
	}
	generator := addTestGenerator(t, serverData, objectID)
	serverData.CosmicObjects.Items[objectID].Fuel = 50

	gameWorld.SetInput(1, game.ShipInput{})
	gameWorld.Tick(2)

	cosmicObject := serverData.CosmicObjects.Items[objectID]
	if cosmicObject.ConsumingPower != 0 {
		t.Fatalf("consuming power = %v, want 0", cosmicObject.ConsumingPower)
	}
	if cosmicObject.GeneratingPower != 60000 {
		t.Fatalf("generating power = %v, want 60000", cosmicObject.GeneratingPower)
	}
	closeWorldFloat(t, cosmicObject.Fuel, 50)
	if generator.Active {
		t.Fatalf("generator must be inactive without power demand: %+v", generator)
	}
}

// Проверяет, что генератор при частичной нагрузке тратит топливо пропорционально спросу на энергию.
func TestTickSpendsGeneratorFuelProportionallyToPowerDemand(t *testing.T) {
	serverData := testWorldData(t)
	gameWorld := world.New(1, serverData)

	objectID, ok := gameWorld.ConnectAccount(1)
	if !ok {
		t.Fatalf("account was not connected")
	}
	generator := addTestGenerator(t, serverData, objectID)
	serverData.CosmicObjects.Items[objectID].Fuel = 50

	gameWorld.SetInput(1, game.ShipInput{ThrustForward: true})
	gameWorld.Tick(2)

	cosmicObject := serverData.CosmicObjects.Items[objectID]
	if cosmicObject.ConsumingPower != 30000 {
		t.Fatalf("consuming power = %v, want 30000", cosmicObject.ConsumingPower)
	}
	if cosmicObject.GeneratingPower != 60000 {
		t.Fatalf("generating power = %v, want 60000", cosmicObject.GeneratingPower)
	}
	closeWorldFloat(t, cosmicObject.Fuel, 46)
	if !generator.Active {
		t.Fatalf("generator must be active with power demand: %+v", generator)
	}
}

// Проверяет, что расход топлива растёт сверх базового уровня, когда потребление выше генерации.
func TestTickSpendsGeneratorFuelAboveBaseRateWhenPowerDemandExceedsGeneration(t *testing.T) {
	serverData := testWorldData(t)
	gameWorld := world.New(1, serverData)

	objectID, ok := gameWorld.ConnectAccount(1)
	if !ok {
		t.Fatalf("account was not connected")
	}
	generator := addTestGenerator(t, serverData, objectID)
	serverData.CosmicObjects.Items[objectID].Fuel = 50
	serverData.EquipmentGroups.GetByCosmicObjectID(objectID)[0].Count = 6
	serverData.EquipmentGroups.GetByCosmicObjectID(objectID)[0].EnabledCount = 6

	gameWorld.SetInput(1, game.ShipInput{ThrustForward: true})
	gameWorld.Tick(2)

	cosmicObject := serverData.CosmicObjects.Items[objectID]
	if cosmicObject.ConsumingPower != 60000 {
		t.Fatalf("consuming power = %v, want 60000", cosmicObject.ConsumingPower)
	}
	if cosmicObject.GeneratingPower != 60000 {
		t.Fatalf("generating power = %v, want 60000", cosmicObject.GeneratingPower)
	}
	closeWorldFloat(t, cosmicObject.Fuel, 42)
	if !generator.Active {
		t.Fatalf("generator must be active with power demand: %+v", generator)
	}
}

// Проверяет, что при пустом баке топливозависимое оборудование не работает.
func TestTickDisablesFuelConsumingEquipmentWithoutFuel(t *testing.T) {
	serverData := testWorldData(t)
	gameWorld := world.New(1, serverData)

	objectID, ok := gameWorld.ConnectAccount(1)
	if !ok {
		t.Fatalf("account was not connected")
	}
	generator := addTestGenerator(t, serverData, objectID)
	serverData.CosmicObjects.Items[objectID].Fuel = 0

	gameWorld.SetInput(1, game.ShipInput{ThrustForward: true, TargetRotationDelta: 1})
	gameWorld.Tick(1)

	cosmicObject := serverData.CosmicObjects.Items[objectID]
	if cosmicObject.ConsumingPower != 0 {
		t.Fatalf("consuming power = %v, want 0", cosmicObject.ConsumingPower)
	}
	if cosmicObject.GeneratingPower != 0 {
		t.Fatalf("generating power = %v, want 0", cosmicObject.GeneratingPower)
	}
	if cosmicObject.Fuel != 0 {
		t.Fatalf("fuel = %v, want 0", cosmicObject.Fuel)
	}
	if cosmicObject.AlongForce != 0 || cosmicObject.AcrossForce != 0 || cosmicObject.Torque != 0 {
		t.Fatalf("fuel-less ship produced force: %+v", cosmicObject)
	}

	installed := serverData.EquipmentGroups.GetByCosmicObjectID(objectID)
	if installed[0].Active || installed[1].Active || generator.Active {
		t.Fatalf("fuel-consuming equipment must be inactive without fuel: %+v %+v %+v", installed[0], installed[1], generator)
	}
}

// Проверяет, что подключенный корабль без топлива тормозит как непилотируемый.
func TestTickTreatsConnectedShipWithoutFuelAsUnpilotedForBrake(t *testing.T) {
	serverData := testWorldData(t)
	serverData.CosmicObjects.Items[1].Fuel = 0
	serverData.CosmicObjects.Items[1].VelocityX = 100
	gameWorld := world.New(1, serverData)

	if _, ok := gameWorld.ConnectAccount(1); !ok {
		t.Fatalf("account was not connected")
	}
	gameWorld.SetInput(1, game.ShipInput{ThrustForward: true})
	gameWorld.Tick(0.1)

	ship := serverData.CosmicObjects.Items[1]
	closeWorldFloat(t, ship.VelocityX, 90)
	closeWorldFloat(t, ship.X, 9)
}

// Проверяет, что управляемый корабль после столкновения не остаётся внутри закреплённого объекта.
func TestTickPreventsControlledShipFromIntersectingAnchoredObject(t *testing.T) {
	serverData := testWorldData(t)
	serverData.CosmicObjects.Items[2].X = 0
	serverData.CosmicObjects.Items[2].Y = 11.7
	serverData.CosmicObjects.Items[2].Enabled = false
	gameWorld := world.New(1, serverData)

	objectID, ok := gameWorld.ConnectAccount(1)
	if !ok {
		t.Fatalf("account was not connected")
	}
	gameWorld.SetInput(1, game.ShipInput{ThrustForward: true})
	snapshot := gameWorld.Tick(1.0 / 30.0)
	object, ok := findCosmicObjectInSnapshot(snapshot, objectID)
	if !ok {
		t.Fatalf("object %v not found in snapshot", objectID)
	}

	obstacle := serverData.CosmicObjects.Items[2]
	objectModel := serverData.CosmicObjectModels.Items[object.CosmicObjectModelID]
	obstacleModel := serverData.CosmicObjectModels.Items[obstacle.CosmicObjectModelID]
	if _, collided := physics.CollisionCorrection(object, *objectModel, *obstacle, *obstacleModel); collided {
		t.Fatalf("controlled object still intersects anchored object: %+v", object)
	}
}

// Проверяет, что столкновение двух управляемых кораблей меняет скорости обоих и разделяет их.
func TestTickAppliesCollisionImpulseToBothControlledShips(t *testing.T) {
	serverData := testWorldData(t)
	serverData.Accounts.Items[2] = &data.Account{ID: 2, Email: "second@email.net", Nickname: "second", PasswordHash: "hash", Token: "token-2", CurrentCharacterID: 2}
	serverData.Characters.Items[2] = &data.Character{ID: 2, AccountID: 2, LocationCosmicObjectID: 4}
	serverData.CosmicObjects.MaxID = 4
	serverData.CosmicObjects.Items[1].X = -1
	serverData.CosmicObjects.Items[1].Y = 0
	serverData.CosmicObjects.Items[1].VelocityX = 10
	serverData.CosmicObjects.Items[1].VelocityY = 0
	serverData.CosmicObjects.Items[4] = &data.CosmicObject{
		ID:                  4,
		Title:               "Second ship",
		CosmicObjectModelID: 1,
		Mass:                900,
		MaxSpeed:            497,
		MaxAngularSpeed:     3,
		MaxAlongForce:       90000,
		MaxAcrossForce:      80000,
		MaxTorque:           70000,
		Enabled:             true,
		X:                   1,
		Y:                   0,
	}
	if err := serverData.Accounts.RebuildIndexes(); err != nil {
		t.Fatal(err)
	}
	if err := serverData.Characters.RebuildIndexes(); err != nil {
		t.Fatal(err)
	}
	if err := serverData.CosmicObjects.RebuildIndexes(); err != nil {
		t.Fatal(err)
	}
	gameWorld := world.New(1, serverData)

	if _, ok := gameWorld.ConnectAccount(1); !ok {
		t.Fatalf("first account was not connected")
	}
	if _, ok := gameWorld.ConnectAccount(2); !ok {
		t.Fatalf("second account was not connected")
	}
	gameWorld.Tick(0)

	first := serverData.CosmicObjects.Items[1]
	second := serverData.CosmicObjects.Items[4]
	if first.VelocityX >= 10 {
		t.Fatalf("first ship velocity was not reduced by collision: %+v", first)
	}
	if second.VelocityX <= 0 {
		t.Fatalf("second ship did not receive collision impulse: %+v", second)
	}
	if first.X >= second.X {
		t.Fatalf("ships were not separated in stable order: first=%+v second=%+v", first, second)
	}
}

// Проверяет, что подвижный объект без управления продолжает смещаться по своей скорости.
func TestTickMovesUncontrolledMovableObjectByVelocity(t *testing.T) {
	serverData := testWorldData(t)
	serverData.CosmicObjects.MaxID = 4
	serverData.CosmicObjects.Items[4] = &data.CosmicObject{
		ID:                  4,
		Title:               "Drifting ship",
		CosmicObjectModelID: 1,
		Mass:                900,
		MaxSpeed:            497,
		MaxAngularSpeed:     3,
		MaxAlongForce:       90000,
		MaxAcrossForce:      80000,
		MaxTorque:           70000,
		Enabled:             true,
		X:                   100,
		Y:                   0,
		VelocityX:           100,
		VelocityY:           0,
	}
	if err := serverData.CosmicObjects.RebuildIndexes(); err != nil {
		t.Fatal(err)
	}
	gameWorld := world.New(1, serverData)

	gameWorld.Tick(0.1)

	driftingShip := serverData.CosmicObjects.Items[4]
	if driftingShip.X <= 100 {
		t.Fatalf("uncontrolled movable object did not move by velocity: %+v", driftingShip)
	}
}

// Проверяет, что подключённый корабль без тяги автоматически снижает скорость.
func TestTickAutobrakesConnectedShipWithoutThrust(t *testing.T) {
	serverData := testWorldData(t)
	serverData.CosmicObjects.Items[1].VelocityX = 100
	gameWorld := world.New(1, serverData)

	if _, ok := gameWorld.ConnectAccount(1); !ok {
		t.Fatalf("account was not connected")
	}
	gameWorld.Tick(0.1)

	ship := serverData.CosmicObjects.Items[1]
	if ship.VelocityX >= 100 {
		t.Fatalf("connected ship did not autobrake: %+v", ship)
	}
}

// Проверяет, что после отключения корабль тормозит постоянным замедлением и продолжает движение.
func TestTickAppliesConstantBrakeToShipAfterDisconnect(t *testing.T) {
	serverData := testWorldData(t)
	serverData.CosmicObjects.Items[1].VelocityX = 100
	gameWorld := world.New(1, serverData)

	if _, ok := gameWorld.ConnectAccount(1); !ok {
		t.Fatalf("account was not connected")
	}
	gameWorld.DisconnectAccount(1)
	gameWorld.Tick(0.1)

	ship := serverData.CosmicObjects.Items[1]
	closeWorldFloat(t, ship.VelocityX, 90)
	closeWorldFloat(t, ship.X, 9)
}

// Проверяет, что торможение неподключённых кораблей одинаково по величине для разных скоростей.
func TestTickDisconnectedShipBrakeDoesNotDependOnSpeed(t *testing.T) {
	serverData := testWorldData(t)
	serverData.CosmicObjects.Items[1].VelocityX = 100
	serverData.CosmicObjects.MaxID = 4
	serverData.CosmicObjects.Items[4] = &data.CosmicObject{
		ID:                  4,
		Title:               "Fast ship",
		CosmicObjectModelID: 1,
		Mass:                900,
		MaxSpeed:            497,
		MaxAngularSpeed:     3,
		MaxAlongForce:       90000,
		MaxAcrossForce:      80000,
		MaxTorque:           70000,
		Enabled:             true,
		X:                   1000,
		Y:                   0,
		VelocityX:           200,
	}
	if err := serverData.CosmicObjects.RebuildIndexes(); err != nil {
		t.Fatal(err)
	}
	gameWorld := world.New(1, serverData)

	gameWorld.Tick(0.1)

	slowShip := serverData.CosmicObjects.Items[1]
	fastShip := serverData.CosmicObjects.Items[4]
	closeWorldFloat(t, 100-slowShip.VelocityX, 10)
	closeWorldFloat(t, 200-fastShip.VelocityX, 10)
}

// Проверяет, что неуправляемый астероид движется без корабельного торможения.
func TestTickDoesNotApplyShipDragToUncontrolledAsteroid(t *testing.T) {
	serverData := testWorldData(t)
	serverData.CosmicObjects.Items[2].Anchored = false
	serverData.CosmicObjects.Items[2].X = 1000
	serverData.CosmicObjects.Items[2].Y = 0
	serverData.CosmicObjects.Items[2].VelocityX = 100
	gameWorld := world.New(1, serverData)

	gameWorld.Tick(0.1)

	asteroid := serverData.CosmicObjects.Items[2]
	closeWorldFloat(t, asteroid.VelocityX, 100)
	closeWorldFloat(t, asteroid.X, 1010)
}

// Проверяет, что при создании мира загруженный корабль получает характеристики сборки без потери движения.
func TestNewAppliesAssemblyToLoadedShip(t *testing.T) {
	serverData := testWorldData(t)
	serverData.CosmicObjects.Items[1].X = 10
	serverData.CosmicObjects.Items[1].VelocityY = 3

	world.New(1, serverData)
	cosmicObject := serverData.CosmicObjects.Items[1]

	if cosmicObject.Mass != 900 || cosmicObject.MaxArmor != 300 || cosmicObject.MaxAlongForce != 90000 || cosmicObject.MaxAcrossForce != 80000 || cosmicObject.MaxTorque != 70000 {
		t.Fatalf("loaded ship stats were not copied from assembly: %+v", cosmicObject)
	}
	if cosmicObject.X != 10 || cosmicObject.VelocityY != 3 {
		t.Fatalf("loaded ship movement state was not preserved: %+v", cosmicObject)
	}
}

// Проверяет, что при создании мира оборудование сборки устанавливается на загруженный корабль.
func TestNewInstallsAssemblyEquipmentOnLoadedShip(t *testing.T) {
	serverData := testWorldData(t)

	world.New(1, serverData)
	installed := serverData.EquipmentGroups.GetByCosmicObjectID(1)

	if len(installed) != 4 {
		t.Fatalf("got %d equipment groups, want 4", len(installed))
	}
	if installed[0].EquipmentItemModelID != 101 || installed[0].Count != 2 || installed[0].EnabledCount != 2 || !installed[0].Enabled || !installed[0].Active {
		t.Fatalf("first equipment group was not copied from assembly: %+v", installed[0])
	}
	if installed[1].EquipmentItemModelID != 102 || installed[1].Count != 1 || installed[1].EnabledCount != 1 || !installed[1].Enabled || !installed[1].Active {
		t.Fatalf("second equipment group was not copied from assembly: %+v", installed[1])
	}
}

// Проверяет, что стартовый аккаунт получает корабль из первой публичной разработческой сборки.
func TestCreateStarterAccountUsesFirstPublicDeveloperAssembly(t *testing.T) {
	serverData := testWorldData(t)
	gameWorld := world.New(1, serverData)

	account, err := gameWorld.CreateStarterAccount()
	if err != nil {
		t.Fatalf("CreateStarterAccount returned error: %v", err)
	}
	character, ok := serverData.Characters.Get(account.CurrentCharacterID)
	if !ok {
		t.Fatalf("created character was not stored")
	}
	cosmicObject, ok := serverData.CosmicObjects.Get(character.LocationCosmicObjectID)
	if !ok {
		t.Fatalf("created ship was not stored")
	}

	if cosmicObject.Mass != 900 || cosmicObject.MaxArmor != 300 || cosmicObject.MaxAlongForce != 90000 || cosmicObject.MaxAcrossForce != 80000 || cosmicObject.MaxTorque != 70000 {
		t.Fatalf("starter ship stats were not copied from assembly: %+v", cosmicObject)
	}
	installed := serverData.EquipmentGroups.GetByCosmicObjectID(cosmicObject.ID)
	if len(installed) != 4 {
		t.Fatalf("got %d starter equipment groups, want 4", len(installed))
	}
}

// Проверяет, что стартовый корабль получает полный бак и по одному виду предметов каждого доступного типа в первом контейнере.
func TestCreateStarterAccountFillsFuelAndFirstContainerWithTypeSamples(t *testing.T) {
	serverData := testWorldData(t)
	gameWorld := world.New(1, serverData)

	account, err := gameWorld.CreateStarterAccount()
	if err != nil {
		t.Fatalf("CreateStarterAccount returned error: %v", err)
	}
	character, ok := serverData.Characters.Get(account.CurrentCharacterID)
	if !ok {
		t.Fatalf("created character was not stored")
	}
	cosmicObject, ok := serverData.CosmicObjects.Get(character.LocationCosmicObjectID)
	if !ok {
		t.Fatalf("created ship was not stored")
	}

	if cosmicObject.Fuel != cosmicObject.MaxFuel || cosmicObject.Fuel != 50 {
		t.Fatalf("starter ship fuel = %v/%v, want 50/50", cosmicObject.Fuel, cosmicObject.MaxFuel)
	}

	var container *data.EquipmentGroup
	for _, group := range serverData.EquipmentGroups.GetByCosmicObjectID(cosmicObject.ID) {
		if group.EquipmentItemModelID == 301 {
			container = group
			break
		}
	}
	if container == nil {
		t.Fatalf("starter ship container was not installed")
	}

	items := serverData.ItemGroups.GetByContainerEquipmentGroupID(container.ID)
	if len(items) != 3 {
		t.Fatalf("got %d container item groups, want 3", len(items))
	}
	counts := map[int64]float64{}
	for _, item := range items {
		counts[item.ContentItemModelID] = item.Count
	}
	for _, itemModelID := range []int64{101, 301, 303} {
		if counts[itemModelID] != 10 {
			t.Fatalf("starter container item model %d count = %v, want 10; all counts: %+v", itemModelID, counts[itemModelID], counts)
		}
	}
}

// Проверяет, что команда панели переносит всё содержимое из одного контейнера объекта в другой и объединяет одинаковые модели.
func TestApplyControlPanelContainerTransferMovesAllItemsToTargetContainer(t *testing.T) {
	serverData := testWorldData(t)
	gameWorld := world.New(1, serverData)

	objectID, ok := gameWorld.ConnectAccount(1)
	if !ok {
		t.Fatalf("account was not connected")
	}
	var source *data.EquipmentGroup
	for _, group := range serverData.EquipmentGroups.GetByCosmicObjectID(objectID) {
		if group.EquipmentItemModelID == 301 {
			source = group
			break
		}
	}
	if source == nil {
		t.Fatalf("source container was not installed")
	}
	target, err := serverData.EquipmentGroups.Add(&data.EquipmentGroup{
		CosmicObjectID:       objectID,
		Title:                "Target Container",
		EquipmentItemModelID: 301,
		Count:                1,
		EnabledCount:         1,
		Enabled:              true,
	})
	if err != nil {
		t.Fatal(err)
	}
	selectedItem, err := serverData.ItemGroups.Add(&data.ItemGroup{ContainerEquipmentGroupID: source.ID, ContentItemModelID: 303, Count: 10})
	if err != nil {
		t.Fatal(err)
	}
	remainingItem, err := serverData.ItemGroups.Add(&data.ItemGroup{ContainerEquipmentGroupID: source.ID, ContentItemModelID: 403, Count: 2})
	if err != nil {
		t.Fatal(err)
	}
	if _, err := serverData.ItemGroups.Add(&data.ItemGroup{ContainerEquipmentGroupID: target.ID, ContentItemModelID: 303, Count: 5}); err != nil {
		t.Fatal(err)
	}

	if err := gameWorld.ApplyControlPanelContainerTransfer(1, "session-1", 4, world.ControlPanelContainerTransfer{
		SourceContainerEquipmentGroupID: source.ID,
		TargetContainerEquipmentGroupID: target.ID,
		ItemGroupIDs:                    []int64{selectedItem.ID},
	}); err != nil {
		t.Fatalf("container transfer returned error: %v", err)
	}

	sourceItems := serverData.ItemGroups.GetByContainerEquipmentGroupID(source.ID)
	if len(sourceItems) != 1 || sourceItems[0].ID != remainingItem.ID {
		t.Fatalf("source container still has items: %+v", serverData.ItemGroups.GetByContainerEquipmentGroupID(source.ID))
	}
	targetItems := serverData.ItemGroups.GetByContainerEquipmentGroupID(target.ID)
	if len(targetItems) != 1 {
		t.Fatalf("got %d target item groups, want 1: %+v", len(targetItems), targetItems)
	}
	counts := map[int64]float64{}
	for _, item := range targetItems {
		counts[item.ContentItemModelID] = item.Count
	}
	if counts[303] != 15 {
		t.Fatalf("target container contents were not merged correctly: %+v", counts)
	}
}

// Проверяет, что команда панели не переносит предметы в группу оборудования, которая не является контейнером.
// Проверяет, что команда панели переносит из одной выбранной строки только указанное количество.
func TestApplyControlPanelContainerTransferMovesRequestedAmount(t *testing.T) {
	serverData := testWorldData(t)
	gameWorld := world.New(1, serverData)

	objectID, ok := gameWorld.ConnectAccount(1)
	if !ok {
		t.Fatalf("account was not connected")
	}
	var source *data.EquipmentGroup
	for _, group := range serverData.EquipmentGroups.GetByCosmicObjectID(objectID) {
		if group.EquipmentItemModelID == 301 {
			source = group
			break
		}
	}
	if source == nil {
		t.Fatalf("source container was not installed")
	}
	target, err := serverData.EquipmentGroups.Add(&data.EquipmentGroup{
		CosmicObjectID:       objectID,
		Title:                "Target Container",
		EquipmentItemModelID: 301,
		Count:                1,
		EnabledCount:         1,
		Enabled:              true,
	})
	if err != nil {
		t.Fatal(err)
	}
	selectedItem, err := serverData.ItemGroups.Add(&data.ItemGroup{ContainerEquipmentGroupID: source.ID, ContentItemModelID: 303, Count: 10})
	if err != nil {
		t.Fatal(err)
	}

	if err := gameWorld.ApplyControlPanelContainerTransfer(1, "session-1", 9, world.ControlPanelContainerTransfer{
		SourceContainerEquipmentGroupID: source.ID,
		TargetContainerEquipmentGroupID: target.ID,
		ItemGroupIDs:                    []int64{selectedItem.ID},
		Amount:                          4,
	}); err != nil {
		t.Fatalf("container transfer returned error: %v", err)
	}

	sourceItems := serverData.ItemGroups.GetByContainerEquipmentGroupID(source.ID)
	if len(sourceItems) != 1 || sourceItems[0].Count != 6 {
		t.Fatalf("source item was not reduced by requested amount: %+v", sourceItems)
	}
	targetItems := serverData.ItemGroups.GetByContainerEquipmentGroupID(target.ID)
	if len(targetItems) != 1 || targetItems[0].Count != 4 {
		t.Fatalf("target item was not created with requested amount: %+v", targetItems)
	}
}

func TestApplyControlPanelContainerTransferRejectsNonContainerTarget(t *testing.T) {
	serverData := testWorldData(t)
	gameWorld := world.New(1, serverData)

	objectID, ok := gameWorld.ConnectAccount(1)
	if !ok {
		t.Fatalf("account was not connected")
	}
	var source *data.EquipmentGroup
	var target *data.EquipmentGroup
	for _, group := range serverData.EquipmentGroups.GetByCosmicObjectID(objectID) {
		if group.EquipmentItemModelID == 301 {
			source = group
		}
		if group.EquipmentItemModelID == 302 {
			target = group
		}
	}
	if source == nil || target == nil {
		t.Fatalf("required equipment was not installed: source=%+v target=%+v", source, target)
	}
	selectedItem, err := serverData.ItemGroups.Add(&data.ItemGroup{ContainerEquipmentGroupID: source.ID, ContentItemModelID: 303, Count: 10})
	if err != nil {
		t.Fatal(err)
	}

	if err := gameWorld.ApplyControlPanelContainerTransfer(1, "session-1", 5, world.ControlPanelContainerTransfer{
		SourceContainerEquipmentGroupID: source.ID,
		TargetContainerEquipmentGroupID: target.ID,
		ItemGroupIDs:                    []int64{selectedItem.ID},
	}); err == nil {
		t.Fatalf("container transfer to non-container target succeeded")
	}
	if items := serverData.ItemGroups.GetByContainerEquipmentGroupID(source.ID); len(items) != 1 || items[0].Count != 10 {
		t.Fatalf("source container changed after rejected transfer: %+v", items)
	}
}

// Проверяет, что случайная смена выбирает другую корабельную модель и заменяет характеристики с оборудованием.
// Проверяет, что команда панели переливает выбранное топливо из контейнера в общий запас топлива объекта до свободного места.
func TestApplyControlPanelFuelTransferFillsObjectFuelFromContainer(t *testing.T) {
	serverData := testWorldData(t)
	serverData.Itemtypes.Items[10] = &data.Itemtype{ID: 10, TitleRu: "Fuel Tank", TitleEn: "Fuel Tank", Acronym: "FuelTank", IsInternalUsable: true, CountMustBeInteger: true}
	serverData.ItemModels.Items[7] = &data.ItemModel{ID: 7, TitleRu: "Fuel", TitleEn: "Fuel", Acronym: "Fuel", ItemtypeID: 7}
	serverData.ItemModels.Items[304] = &data.ItemModel{ID: 304, TitleRu: "Fuel Tank", TitleEn: "Fuel Tank", Acronym: "FuelTank", ItemtypeID: 10, Capacity: 100, ConsumingItemModelID: 7}
	if err := serverData.Itemtypes.RebuildIndexes(); err != nil {
		t.Fatal(err)
	}
	if err := serverData.ItemModels.RebuildIndexes(); err != nil {
		t.Fatal(err)
	}
	gameWorld := world.New(1, serverData)

	objectID, ok := gameWorld.ConnectAccount(1)
	if !ok {
		t.Fatalf("account was not connected")
	}
	cosmicObject := serverData.CosmicObjects.Items[objectID]
	cosmicObject.Fuel = 20
	cosmicObject.MaxFuel = 50
	var container *data.EquipmentGroup
	for _, group := range serverData.EquipmentGroups.GetByCosmicObjectID(objectID) {
		if group.EquipmentItemModelID == 301 {
			container = group
			break
		}
	}
	if container == nil {
		t.Fatalf("container was not installed")
	}
	fuelTank, err := serverData.EquipmentGroups.Add(&data.EquipmentGroup{
		CosmicObjectID:       objectID,
		Title:                "Fuel Tank",
		EquipmentItemModelID: 304,
		Count:                1,
		EnabledCount:         1,
		Enabled:              true,
	})
	if err != nil {
		t.Fatal(err)
	}
	fuelItem, err := serverData.ItemGroups.Add(&data.ItemGroup{ContainerEquipmentGroupID: container.ID, ContentItemModelID: 7, Count: 40})
	if err != nil {
		t.Fatal(err)
	}

	if err := gameWorld.ApplyControlPanelFuelTransfer(1, "session-1", 6, world.ControlPanelFuelTransfer{
		ContainerEquipmentGroupID: container.ID,
		FuelTankEquipmentGroupID:  fuelTank.ID,
		ItemGroupIDs:              []int64{fuelItem.ID},
	}); err != nil {
		t.Fatalf("fuel transfer returned error: %v", err)
	}

	if cosmicObject.Fuel != 50 {
		t.Fatalf("object fuel = %v, want 50", cosmicObject.Fuel)
	}
	items := serverData.ItemGroups.GetByContainerEquipmentGroupID(container.ID)
	if len(items) != 1 || items[0].Count != 10 {
		t.Fatalf("container fuel item was not reduced to remainder: %+v", items)
	}
}

// Проверяет, что команда панели сливает указанное топливо из общего запаса объекта в левый контейнер.
// Проверяет, что команда панели заливает из контейнера только указанное количество топлива.
func TestApplyControlPanelFuelTransferFillsOnlyRequestedAmount(t *testing.T) {
	serverData := testWorldData(t)
	serverData.Itemtypes.Items[10] = &data.Itemtype{ID: 10, TitleRu: "Fuel Tank", TitleEn: "Fuel Tank", Acronym: "FuelTank", IsInternalUsable: true, CountMustBeInteger: true}
	serverData.ItemModels.Items[7] = &data.ItemModel{ID: 7, TitleRu: "Fuel", TitleEn: "Fuel", Acronym: "Fuel", ItemtypeID: 7}
	serverData.ItemModels.Items[304] = &data.ItemModel{ID: 304, TitleRu: "Fuel Tank", TitleEn: "Fuel Tank", Acronym: "FuelTank", ItemtypeID: 10, Capacity: 100, ConsumingItemModelID: 7}
	if err := serverData.Itemtypes.RebuildIndexes(); err != nil {
		t.Fatal(err)
	}
	if err := serverData.ItemModels.RebuildIndexes(); err != nil {
		t.Fatal(err)
	}
	gameWorld := world.New(1, serverData)

	objectID, ok := gameWorld.ConnectAccount(1)
	if !ok {
		t.Fatalf("account was not connected")
	}
	cosmicObject := serverData.CosmicObjects.Items[objectID]
	cosmicObject.Fuel = 20
	cosmicObject.MaxFuel = 50
	var container *data.EquipmentGroup
	for _, group := range serverData.EquipmentGroups.GetByCosmicObjectID(objectID) {
		if group.EquipmentItemModelID == 301 {
			container = group
			break
		}
	}
	if container == nil {
		t.Fatalf("container was not installed")
	}
	fuelTank, err := serverData.EquipmentGroups.Add(&data.EquipmentGroup{
		CosmicObjectID:       objectID,
		Title:                "Fuel Tank",
		EquipmentItemModelID: 304,
		Count:                1,
		EnabledCount:         1,
		Enabled:              true,
	})
	if err != nil {
		t.Fatal(err)
	}
	fuelItem, err := serverData.ItemGroups.Add(&data.ItemGroup{ContainerEquipmentGroupID: container.ID, ContentItemModelID: 7, Count: 40})
	if err != nil {
		t.Fatal(err)
	}

	if err := gameWorld.ApplyControlPanelFuelTransfer(1, "session-1", 8, world.ControlPanelFuelTransfer{
		ContainerEquipmentGroupID: container.ID,
		FuelTankEquipmentGroupID:  fuelTank.ID,
		ItemGroupIDs:              []int64{fuelItem.ID},
		Amount:                    12,
	}); err != nil {
		t.Fatalf("fuel transfer returned error: %v", err)
	}

	if cosmicObject.Fuel != 32 {
		t.Fatalf("object fuel = %v, want 32", cosmicObject.Fuel)
	}
	items := serverData.ItemGroups.GetByContainerEquipmentGroupID(container.ID)
	if len(items) != 1 || items[0].Count != 28 {
		t.Fatalf("container fuel item was not reduced by requested amount: %+v", items)
	}
}

func TestApplyControlPanelFuelTransferDrainsObjectFuelToContainer(t *testing.T) {
	serverData := testWorldData(t)
	serverData.Itemtypes.Items[10] = &data.Itemtype{ID: 10, TitleRu: "Fuel Tank", TitleEn: "Fuel Tank", Acronym: "FuelTank", IsInternalUsable: true, CountMustBeInteger: true}
	serverData.ItemModels.Items[7] = &data.ItemModel{ID: 7, TitleRu: "Fuel", TitleEn: "Fuel", Acronym: "Fuel", ItemtypeID: 7}
	serverData.ItemModels.Items[304] = &data.ItemModel{ID: 304, TitleRu: "Fuel Tank", TitleEn: "Fuel Tank", Acronym: "FuelTank", ItemtypeID: 10, Capacity: 100, ConsumingItemModelID: 7}
	if err := serverData.Itemtypes.RebuildIndexes(); err != nil {
		t.Fatal(err)
	}
	if err := serverData.ItemModels.RebuildIndexes(); err != nil {
		t.Fatal(err)
	}
	gameWorld := world.New(1, serverData)

	objectID, ok := gameWorld.ConnectAccount(1)
	if !ok {
		t.Fatalf("account was not connected")
	}
	cosmicObject := serverData.CosmicObjects.Items[objectID]
	cosmicObject.Fuel = 20
	cosmicObject.MaxFuel = 50
	var container *data.EquipmentGroup
	for _, group := range serverData.EquipmentGroups.GetByCosmicObjectID(objectID) {
		if group.EquipmentItemModelID == 301 {
			container = group
			break
		}
	}
	if container == nil {
		t.Fatalf("container was not installed")
	}
	fuelTank, err := serverData.EquipmentGroups.Add(&data.EquipmentGroup{
		CosmicObjectID:       objectID,
		Title:                "Fuel Tank",
		EquipmentItemModelID: 304,
		Count:                1,
		EnabledCount:         1,
		Enabled:              true,
	})
	if err != nil {
		t.Fatal(err)
	}

	if err := gameWorld.ApplyControlPanelFuelTransfer(1, "session-1", 7, world.ControlPanelFuelTransfer{
		ContainerEquipmentGroupID: container.ID,
		FuelTankEquipmentGroupID:  fuelTank.ID,
		Amount:                    12,
	}); err != nil {
		t.Fatalf("fuel drain returned error: %v", err)
	}

	if cosmicObject.Fuel != 8 {
		t.Fatalf("object fuel = %v, want 8", cosmicObject.Fuel)
	}
	items := serverData.ItemGroups.GetByContainerEquipmentGroupID(container.ID)
	if len(items) != 1 || items[0].ContentItemModelID != 7 || items[0].Count != 12 {
		t.Fatalf("container fuel item was not created: %+v", items)
	}
}

func TestChangeControlledShipToRandomModelUsesOnlyOtherShipModels(t *testing.T) {
	serverData := testWorldData(t)
	gameWorld := world.New(1, serverData)

	objectID, ok := gameWorld.ConnectAccount(1)
	if !ok {
		t.Fatalf("account was not connected")
	}
	if !gameWorld.ChangeControlledShipToRandomModel(1) {
		t.Fatalf("ship model was not changed")
	}
	snapshot := gameWorld.SnapshotForAccount(1)
	object, ok := findCosmicObjectInSnapshot(snapshot, objectID)
	if !ok {
		t.Fatalf("object %v not found in snapshot", objectID)
	}

	if object.CosmicObjectModelID != 4 {
		t.Fatalf("got model %v, want another ship model 4", object.CosmicObjectModelID)
	}
	if object.Mass != 1200 || object.Capacity != 20 || object.MaxArmor != 400 || object.MaxSpeed != 600 || object.MaxAngularSpeed != 4 || object.MaxAlongForce != 120000 || object.MaxAcrossForce != 110000 || object.MaxTorque != 100000 {
		t.Fatalf("ship stats were not copied from selected assembly and model: %+v", object)
	}
	if object.Armor != object.MaxArmor || object.Armor != 400 {
		t.Fatalf("changed ship armor = %v/%v, want 400/400", object.Armor, object.MaxArmor)
	}
	if len(serverData.CosmicObjects.GetByCosmicObjectModelID(1)) != 0 {
		t.Fatalf("old model index still contains changed object")
	}
	if len(serverData.CosmicObjects.GetByCosmicObjectModelID(4)) != 1 {
		t.Fatalf("new model index does not contain changed object")
	}
	installed := serverData.EquipmentGroups.GetByCosmicObjectID(objectID)
	if len(installed) != 3 {
		t.Fatalf("got %d equipment groups after model change, want 3", len(installed))
	}
	if installed[0].EquipmentItemModelID != 201 || installed[0].Count != 4 {
		t.Fatalf("equipment was not replaced from selected assembly: %+v", installed[0])
	}
}

// Проверяет, что после смены модели топливо и боезапас пересоздаются для нового контейнера.
func TestChangeControlledShipToRandomModelRefillsFuelAndContainerAmmo(t *testing.T) {
	serverData := testWorldData(t)
	gameWorld := world.New(1, serverData)

	objectID, ok := gameWorld.ConnectAccount(1)
	if !ok {
		t.Fatalf("account was not connected")
	}
	oldContainer := serverData.EquipmentGroups.GetByCosmicObjectID(objectID)[2]
	if _, err := serverData.ItemGroups.Add(&data.ItemGroup{
		ContainerEquipmentGroupID: oldContainer.ID,
		ContentItemModelID:        303,
		Count:                     1,
	}); err != nil {
		t.Fatal(err)
	}
	serverData.CosmicObjects.Items[objectID].Fuel = 1

	if !gameWorld.ChangeControlledShipToRandomModel(1) {
		t.Fatalf("ship model was not changed")
	}
	cosmicObject, ok := serverData.CosmicObjects.Get(objectID)
	if !ok {
		t.Fatalf("object was not found")
	}
	if cosmicObject.Fuel != cosmicObject.MaxFuel || cosmicObject.Fuel != 51 {
		t.Fatalf("changed ship fuel = %v/%v, want 51/51", cosmicObject.Fuel, cosmicObject.MaxFuel)
	}
	if len(serverData.ItemGroups.GetByContainerEquipmentGroupID(oldContainer.ID)) != 0 {
		t.Fatalf("old container contents were not removed")
	}

	var container *data.EquipmentGroup
	for _, group := range serverData.EquipmentGroups.GetByCosmicObjectID(objectID) {
		if group.EquipmentItemModelID == 401 {
			container = group
			break
		}
	}
	if container == nil {
		t.Fatalf("changed ship container was not installed")
	}

	items := serverData.ItemGroups.GetByContainerEquipmentGroupID(container.ID)
	if len(items) != 3 {
		t.Fatalf("got %d changed container item groups, want 3", len(items))
	}
	counts := map[int64]float64{}
	for _, item := range items {
		counts[item.ContentItemModelID] = item.Count
	}
	for _, itemModelID := range []int64{101, 301, 303} {
		if counts[itemModelID] != 10 {
			t.Fatalf("changed container item model %d count = %v, want 10; all counts: %+v", itemModelID, counts[itemModelID], counts)
		}
	}
}

// Проверяет, что смена модели сохраняет положение и скорость управляемого корабля.
func TestChangeControlledShipToRandomModelKeepsMovementState(t *testing.T) {
	serverData := testWorldData(t)
	serverData.CosmicObjects.Items[1].X = 10
	serverData.CosmicObjects.Items[1].Y = 20
	serverData.CosmicObjects.Items[1].VelocityX = 3
	serverData.CosmicObjects.Items[1].VelocityY = 4
	serverData.CosmicObjects.Items[1].Rotation = 2.5
	serverData.CosmicObjects.Items[1].TargetRotation = 0
	gameWorld := world.New(1, serverData)

	if _, ok := gameWorld.ConnectAccount(1); !ok {
		t.Fatalf("account was not connected")
	}
	if !gameWorld.ChangeControlledShipToRandomModel(1) {
		t.Fatalf("ship model was not changed")
	}
	object, ok := serverData.CosmicObjects.Get(1)
	if !ok {
		t.Fatalf("object was not found")
	}

	if object.X != 10 || object.Y != 20 || object.VelocityX != 3 || object.VelocityY != 4 || object.Rotation != 2.5 {
		t.Fatalf("movement state was not preserved: %+v", object)
	}
	if object.TargetRotation != object.Rotation {
		t.Fatalf("target rotation = %v, want current rotation %v", object.TargetRotation, object.Rotation)
	}
}

// Проверяет, что отключение аккаунта не удаляет корабль из снимка мира.
func TestDisconnectAccountDoesNotDeleteStoredShip(t *testing.T) {
	gameWorld := world.New(1, testWorldData(t))

	objectID, ok := gameWorld.ConnectAccount(1)
	if !ok {
		t.Fatalf("account was not connected")
	}
	gameWorld.DisconnectAccount(1)
	snapshot := gameWorld.Tick(1.0 / 30.0)

	if _, ok := findCosmicObjectInSnapshot(snapshot, objectID); !ok {
		t.Fatalf("object %v was removed after disconnect", objectID)
	}
}

// Проверяет, что снимок мира содержит объекты, загруженные из серверных данных.
func TestSnapshotContainsObjectsLoadedFromData(t *testing.T) {
	gameWorld := world.New(1, testWorldData(t))

	snapshot := gameWorld.Tick(1.0 / 30.0)
	models := map[int64]bool{}
	for _, object := range snapshot.Objects {
		models[object.CosmicObjectModelID] = true
	}

	if !models[1] {
		t.Fatalf("snapshot does not contain model 1")
	}
	if !models[2] {
		t.Fatalf("snapshot does not contain model 2")
	}
	if !models[3] {
		t.Fatalf("snapshot does not contain model 3")
	}
}

// Проверяет, что сохранение данных записывает обновлённое положение космического объекта.
func TestSaveDataWritesCosmicObjectPosition(t *testing.T) {
	serverData := testWorldData(t)
	serverData.CosmicObjects.Items[1].Fuel = 50
	gameWorld := world.New(1, serverData)
	workingDirectory := t.TempDir()
	if err := os.Mkdir(filepath.Join(workingDirectory, "data"), 0o700); err != nil {
		t.Fatal(err)
	}

	if _, ok := gameWorld.ConnectAccount(1); !ok {
		t.Fatalf("account was not connected")
	}
	addTestPowerProducer(t, serverData, 1)
	gameWorld.SetInput(1, game.ShipInput{ThrustForward: true})
	gameWorld.Tick(1)
	if err := gameWorld.SaveData(workingDirectory); err != nil {
		t.Fatal(err)
	}

	loadedCosmicObjects := data.NewCosmicObjects()
	if err := loadedCosmicObjects.LoadFromFile(filepath.Join(workingDirectory, "data", "CosmicObjects.json")); err != nil {
		t.Fatal(err)
	}
	if loadedCosmicObjects.Items[1].Y <= 0 {
		t.Fatalf("got saved Y %v, want positive", loadedCosmicObjects.Items[1].Y)
	}
}
