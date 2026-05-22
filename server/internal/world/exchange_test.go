package world_test

import (
	"testing"

	"space-game-07-server/internal/data"
	"space-game-07-server/internal/game"
	"space-game-07-server/internal/world"
)

// Проверяет, что обмен после одобрения не открывается до завершения стыковки.
func TestExchangeRequestOpensExchangeAfterDocking(t *testing.T) {
	serverData := exchangeWorldData(t)
	serverData.CosmicObjects.Items[3].X = 0
	serverData.CosmicObjects.Items[3].Y = 14
	gameWorld := world.New(1, serverData)
	if _, ok := gameWorld.ConnectAccount(1); !ok {
		t.Fatal("sender account was not connected")
	}
	if _, ok := gameWorld.ConnectAccount(2); !ok {
		t.Fatal("receiver account was not connected")
	}

	if err := gameWorld.SendExchangeRequest(1); err != nil {
		t.Fatalf("send exchange request: %v", err)
	}
	_ = gameWorld.DrainExchangeEvents()
	if err := gameWorld.ApproveExchangeRequest(2); err != nil {
		t.Fatalf("approve exchange request: %v", err)
	}

	if exchangeEventsContainKind(gameWorld.DrainExchangeEvents(), "exchangeState") {
		t.Fatal("exchange window was opened before docking finished")
	}
	gameWorld.Tick(10)
	if !exchangeEventsContainKind(gameWorld.DrainExchangeEvents(), "exchangeState") {
		t.Fatal("exchange window was not opened after docking finished")
	}
}

// Проверяет, что обычная отстыковка запрещена, пока окно обмена открыто.
func TestExchangeSessionBlocksManualUndock(t *testing.T) {
	serverData := exchangeWorldData(t)
	serverData.CosmicObjects.Items[1].ClusterMainCosmicObjectID = 3
	serverData.CosmicObjects.Items[3].ClusterMainCosmicObjectID = 3
	gameWorld := world.New(1, serverData)
	if _, ok := gameWorld.ConnectAccount(1); !ok {
		t.Fatal("sender account was not connected")
	}
	if _, ok := gameWorld.ConnectAccount(2); !ok {
		t.Fatal("receiver account was not connected")
	}

	if err := gameWorld.SendExchangeRequest(1); err != nil {
		t.Fatalf("send exchange request: %v", err)
	}
	if err := gameWorld.ApproveExchangeRequest(2); err != nil {
		t.Fatalf("approve exchange request: %v", err)
	}

	if err := gameWorld.UndockControlledObject(1); err == nil {
		t.Fatal("undock was allowed during exchange")
	}
}

// Проверяет, что при открытии обмена первый контейнер сразу выбран как источник.
func TestExchangeSessionUsesDefaultSourceContainer(t *testing.T) {
	serverData := exchangeWorldData(t)
	serverData.CosmicObjects.Items[1].ClusterMainCosmicObjectID = 3
	serverData.CosmicObjects.Items[3].ClusterMainCosmicObjectID = 3
	sourceContainer := serverData.EquipmentGroups.GetByCosmicObjectID(1)[0]
	itemGroup, err := serverData.ItemGroups.Add(&data.ItemGroup{ContainerEquipmentGroupID: sourceContainer.ID, ContentItemModelID: 303, Count: 10})
	if err != nil {
		t.Fatal(err)
	}
	gameWorld := world.New(1, serverData)
	if _, ok := gameWorld.ConnectAccount(1); !ok {
		t.Fatal("sender account was not connected")
	}
	if _, ok := gameWorld.ConnectAccount(2); !ok {
		t.Fatal("receiver account was not connected")
	}

	if err := gameWorld.SendExchangeRequest(1); err != nil {
		t.Fatalf("send exchange request: %v", err)
	}
	if err := gameWorld.ApproveExchangeRequest(2); err != nil {
		t.Fatalf("approve exchange request: %v", err)
	}
	_ = gameWorld.DrainExchangeEvents()
	if err := gameWorld.AddExchangeItems(1, []int64{itemGroup.ID}, 4); err != nil {
		t.Fatalf("add exchange items: %v", err)
	}

	events := gameWorld.DrainExchangeEvents()
	if !exchangeEventsContainSelfQueueCount(events, 4) {
		t.Fatalf("exchange queue did not receive selected amount: %+v", events)
	}
}

// Проверяет, что полный обмен одинаковыми предметами не теряет содержимое при финальном переносе.
func TestExchangeCompletionMovesItemsToReceiverContainers(t *testing.T) {
	serverData := exchangeWorldData(t)
	serverData.CosmicObjects.Items[1].ClusterMainCosmicObjectID = 3
	serverData.CosmicObjects.Items[3].ClusterMainCosmicObjectID = 3
	leftContainer := serverData.EquipmentGroups.GetByCosmicObjectID(1)[0]
	rightContainer := serverData.EquipmentGroups.GetByCosmicObjectID(3)[0]
	leftItem, err := serverData.ItemGroups.Add(&data.ItemGroup{ContainerEquipmentGroupID: leftContainer.ID, ContentItemModelID: 303, Count: 10})
	if err != nil {
		t.Fatal(err)
	}
	rightItem, err := serverData.ItemGroups.Add(&data.ItemGroup{ContainerEquipmentGroupID: rightContainer.ID, ContentItemModelID: 303, Count: 7})
	if err != nil {
		t.Fatal(err)
	}
	gameWorld := world.New(1, serverData)
	if _, ok := gameWorld.ConnectAccount(1); !ok {
		t.Fatal("sender account was not connected")
	}
	if _, ok := gameWorld.ConnectAccount(2); !ok {
		t.Fatal("receiver account was not connected")
	}

	if err := gameWorld.SendExchangeRequest(1); err != nil {
		t.Fatalf("send exchange request: %v", err)
	}
	if err := gameWorld.ApproveExchangeRequest(2); err != nil {
		t.Fatalf("approve exchange request: %v", err)
	}
	if err := gameWorld.AddExchangeItems(1, []int64{leftItem.ID}, 10); err != nil {
		t.Fatalf("add sender items: %v", err)
	}
	if err := gameWorld.AddExchangeItems(2, []int64{rightItem.ID}, 7); err != nil {
		t.Fatalf("add receiver items: %v", err)
	}
	if err := gameWorld.ConfirmExchange(1); err != nil {
		t.Fatalf("confirm sender exchange: %v", err)
	}
	if err := gameWorld.ConfirmExchange(2); err != nil {
		t.Fatalf("confirm receiver exchange: %v", err)
	}
	gameWorld.Tick(1000000)
	gameWorld.Tick(1000000)

	if count := exchangeContainerItemCount(serverData, leftContainer.ID, 303); count != 7 {
		t.Fatalf("left container item count = %v, want 7", count)
	}
	if count := exchangeContainerItemCount(serverData, rightContainer.ID, 303); count != 10 {
		t.Fatalf("right container item count = %v, want 10", count)
	}
	if len(serverData.TaskItemGroups.Items) != 0 {
		t.Fatalf("exchange task reserves were not removed: %+v", serverData.TaskItemGroups.Items)
	}
}

// Проверяет, что при повторном подключении ожидающий запрос обмена можно снова отправить клиенту.
func TestExchangeEventsForAccountRestoresPendingRequestAfterReconnect(t *testing.T) {
	serverData := exchangeWorldData(t)
	serverData.CosmicObjects.Items[3].X = 0
	serverData.CosmicObjects.Items[3].Y = 14
	gameWorld := world.New(1, serverData)
	if _, ok := gameWorld.ConnectAccount(1); !ok {
		t.Fatal("sender account was not connected")
	}
	if _, ok := gameWorld.ConnectAccount(2); !ok {
		t.Fatal("receiver account was not connected")
	}

	if err := gameWorld.SendExchangeRequest(1); err != nil {
		t.Fatalf("send exchange request: %v", err)
	}
	_ = gameWorld.DrainExchangeEvents()

	senderEvents := gameWorld.ExchangeEventsForAccount(1)
	receiverEvents := gameWorld.ExchangeEventsForAccount(2)
	if len(senderEvents) != 1 || senderEvents[0].Kind != "exchangeRequestStarted" || senderEvents[0].Role != "sender" {
		t.Fatalf("sender reconnect events = %+v", senderEvents)
	}
	if len(receiverEvents) != 1 || receiverEvents[0].Kind != "exchangeRequestStarted" || receiverEvents[0].Role != "receiver" {
		t.Fatalf("receiver reconnect events = %+v", receiverEvents)
	}
}

// Проверяет, что при повторном подключении открытое окно обмена можно снова отправить клиенту.
func TestExchangeEventsForAccountRestoresOpenSessionAfterReconnect(t *testing.T) {
	serverData := exchangeWorldData(t)
	serverData.CosmicObjects.Items[1].ClusterMainCosmicObjectID = 3
	serverData.CosmicObjects.Items[3].ClusterMainCosmicObjectID = 3
	gameWorld := world.New(1, serverData)
	if _, ok := gameWorld.ConnectAccount(1); !ok {
		t.Fatal("sender account was not connected")
	}
	if _, ok := gameWorld.ConnectAccount(2); !ok {
		t.Fatal("receiver account was not connected")
	}

	if err := gameWorld.SendExchangeRequest(1); err != nil {
		t.Fatalf("send exchange request: %v", err)
	}
	if err := gameWorld.ApproveExchangeRequest(2); err != nil {
		t.Fatalf("approve exchange request: %v", err)
	}
	_ = gameWorld.DrainExchangeEvents()

	events := gameWorld.ExchangeEventsForAccount(1)
	if len(events) != 1 || events[0].Kind != "exchangeState" || events[0].Role != "sender" {
		t.Fatalf("sender reconnect events = %+v", events)
	}
	if events[0].State.SelfObjectID != 1 || events[0].State.OtherObjectID != 3 {
		t.Fatalf("sender exchange state = %+v", events[0].State)
	}
}

func exchangeWorldData(t *testing.T) world.Data {
	t.Helper()

	serverData := testWorldData(t)
	serverData.Accounts.Items[2] = &data.Account{ID: 2, Email: "receiver@email.net", Nickname: "receiver", PasswordHash: "hash", Token: "token-2", CurrentCharacterID: 2}
	serverData.Characters.Items[2] = &data.Character{ID: 2, AccountID: 2, LocationCosmicObjectID: 3}
	serverData.CosmicObjects.Items[3].OwnerCharacterID = 2
	if _, err := serverData.TaskTypes.Add(&data.TaskType{TitleRu: "Обмен", TitleEn: "Exchange", Acronym: "Exchange"}); err != nil {
		t.Fatal(err)
	}
	if _, err := serverData.EquipmentGroups.Add(&data.EquipmentGroup{CosmicObjectID: 1, Title: "Left container", EquipmentItemModelID: 301, Count: 1, EnabledCount: 1, Enabled: true}); err != nil {
		t.Fatal(err)
	}
	if _, err := serverData.EquipmentGroups.Add(&data.EquipmentGroup{CosmicObjectID: 3, Title: "Right container", EquipmentItemModelID: 301, Count: 1, EnabledCount: 1, Enabled: true}); err != nil {
		t.Fatal(err)
	}
	if err := serverData.Accounts.RebuildIndexes(); err != nil {
		t.Fatal(err)
	}
	if err := serverData.Characters.RebuildIndexes(); err != nil {
		t.Fatal(err)
	}
	if err := serverData.TaskTypes.RebuildIndexes(); err != nil {
		t.Fatal(err)
	}
	if err := serverData.EquipmentGroups.RebuildIndexes(); err != nil {
		t.Fatal(err)
	}
	return serverData
}

func exchangeEventsContainKind(events []game.ExchangeEvent, kind string) bool {
	for _, event := range events {
		if event.Kind == kind {
			return true
		}
	}
	return false
}

// Проверяет, что в событиях есть строка своей очереди с нужным количеством.
func exchangeEventsContainSelfQueueCount(events []game.ExchangeEvent, count float64) bool {
	for _, event := range events {
		if event.Kind != "exchangeState" {
			continue
		}
		for _, item := range event.State.SelfQueue {
			if item.Count == count {
				return true
			}
		}
	}
	return false
}

// Возвращает количество предметов указанной модели в контейнере после обмена.
func exchangeContainerItemCount(serverData world.Data, containerID int64, itemModelID int64) float64 {
	total := 0.0
	for _, itemGroup := range serverData.ItemGroups.Items {
		if itemGroup.ContainerEquipmentGroupID == containerID && itemGroup.ContentItemModelID == itemModelID {
			total += itemGroup.Count
		}
	}
	return total
}
