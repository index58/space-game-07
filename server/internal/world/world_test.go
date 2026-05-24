package world_test

import (
	"encoding/json"
	"fmt"
	"math"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"testing"

	"space-game-07-server/internal/data"
	"space-game-07-server/internal/game"
	"space-game-07-server/internal/physics"
	"space-game-07-server/internal/storage"
	"space-game-07-server/internal/world"
)

// Р В РЎСџР РЋР вЂљР В РЎвЂўР В Р вЂ Р В Р’ВµР РЋР вЂљР РЋР РЏР В Р’ВµР РЋРІР‚С™, Р РЋРІР‚РЋР РЋРІР‚С™Р В РЎвЂў Р РЋР С“Р В РЎвЂўР РЋРІР‚В¦Р РЋР вЂљР В Р’В°Р В Р вЂ¦Р В Р’ВµР В Р вЂ¦Р В РЎвЂР В Р’Вµ Р В РўвЂР В Р’В°Р В Р вЂ¦Р В Р вЂ¦Р РЋРІР‚в„–Р РЋРІР‚В¦ Р В Р’В·Р В Р’В°Р В РЎвЂ”Р В РЎвЂР РЋР С“Р РЋРІР‚в„–Р В Р вЂ Р В Р’В°Р В Р’ВµР РЋРІР‚С™ Р В РЎвЂўР В Р’В±Р В Р вЂ¦Р В РЎвЂўР В Р вЂ Р В Р’В»Р РЋРІР‚ВР В Р вЂ¦Р В Р вЂ¦Р В РЎвЂўР В Р’Вµ Р В РЎвЂ”Р В РЎвЂўР В Р’В»Р В РЎвЂўР В Р’В¶Р В Р’ВµР В Р вЂ¦Р В РЎвЂР В Р’Вµ Р В РЎвЂќР В РЎвЂўР РЋР С“Р В РЎВР В РЎвЂР РЋРІР‚РЋР В Р’ВµР РЋР С“Р В РЎвЂќР В РЎвЂўР В РЎвЂ“Р В РЎвЂў Р В РЎвЂўР В Р’В±Р РЋР вЂ°Р В Р’ВµР В РЎвЂќР РЋРІР‚С™Р В Р’В°.
func findCosmicObjectInSnapshot(snapshot game.Snapshot, objectID int64) (game.SnapshotCosmicObject, bool) {
	for _, object := range snapshot.Objects {
		if object.ID == objectID {
			return object, true
		}
	}

	return game.SnapshotCosmicObject{}, false
}

// Р В РЎСџР РЋР вЂљР В РЎвЂўР В Р вЂ Р В Р’ВµР РЋР вЂљР РЋР РЏР В Р’ВµР РЋРІР‚С™, Р РЋРІР‚РЋР РЋРІР‚С™Р В РЎвЂў Р РЋР С“Р В РЎвЂўР РЋРІР‚В¦Р РЋР вЂљР В Р’В°Р В Р вЂ¦Р В Р’ВµР В Р вЂ¦Р В РЎвЂР В Р’Вµ Р В РўвЂР В Р’В°Р В Р вЂ¦Р В Р вЂ¦Р РЋРІР‚в„–Р РЋРІР‚В¦ Р В Р’В·Р В Р’В°Р В РЎвЂ”Р В РЎвЂР РЋР С“Р РЋРІР‚в„–Р В Р вЂ Р В Р’В°Р В Р’ВµР РЋРІР‚С™ Р В РЎвЂўР В Р’В±Р В Р вЂ¦Р В РЎвЂўР В Р вЂ Р В Р’В»Р РЋРІР‚ВР В Р вЂ¦Р В Р вЂ¦Р В РЎвЂўР В Р’Вµ Р В РЎвЂ”Р В РЎвЂўР В Р’В»Р В РЎвЂўР В Р’В¶Р В Р’ВµР В Р вЂ¦Р В РЎвЂР В Р’Вµ Р В РЎвЂќР В РЎвЂўР РЋР С“Р В РЎВР В РЎвЂР РЋРІР‚РЋР В Р’ВµР РЋР С“Р В РЎвЂќР В РЎвЂўР В РЎвЂ“Р В РЎвЂў Р В РЎвЂўР В Р’В±Р РЋР вЂ°Р В Р’ВµР В РЎвЂќР РЋРІР‚С™Р В Р’В°.
func findCosmicObjectModelInSnapshot(snapshot game.Snapshot, modelID int64) (game.SnapshotCosmicObject, bool) {
	for _, object := range snapshot.Objects {
		if object.CosmicObjectModelID == modelID {
			return object, true
		}
	}

	return game.SnapshotCosmicObject{}, false
}

// Находит все объекты указанной модели в снимке мира.
func findCosmicObjectModelsInSnapshot(snapshot game.Snapshot, modelID int64) []game.SnapshotCosmicObject {
	objects := make([]game.SnapshotCosmicObject, 0)
	for _, object := range snapshot.Objects {
		if object.CosmicObjectModelID == modelID {
			objects = append(objects, object)
		}
	}

	return objects
}

// Р В РЎСџР РЋР вЂљР В РЎвЂўР В Р вЂ Р В Р’ВµР РЋР вЂљР РЋР РЏР В Р’ВµР РЋРІР‚С™, Р РЋРІР‚РЋР РЋРІР‚С™Р В РЎвЂў Р РЋР С“Р В РЎвЂўР РЋРІР‚В¦Р РЋР вЂљР В Р’В°Р В Р вЂ¦Р В Р’ВµР В Р вЂ¦Р В РЎвЂР В Р’Вµ Р В РўвЂР В Р’В°Р В Р вЂ¦Р В Р вЂ¦Р РЋРІР‚в„–Р РЋРІР‚В¦ Р В Р’В·Р В Р’В°Р В РЎвЂ”Р В РЎвЂР РЋР С“Р РЋРІР‚в„–Р В Р вЂ Р В Р’В°Р В Р’ВµР РЋРІР‚С™ Р В РЎвЂўР В Р’В±Р В Р вЂ¦Р В РЎвЂўР В Р вЂ Р В Р’В»Р РЋРІР‚ВР В Р вЂ¦Р В Р вЂ¦Р В РЎвЂўР В Р’Вµ Р В РЎвЂ”Р В РЎвЂўР В Р’В»Р В РЎвЂўР В Р’В¶Р В Р’ВµР В Р вЂ¦Р В РЎвЂР В Р’Вµ Р В РЎвЂќР В РЎвЂўР РЋР С“Р В РЎВР В РЎвЂР РЋРІР‚РЋР В Р’ВµР РЋР С“Р В РЎвЂќР В РЎвЂўР В РЎвЂ“Р В РЎвЂў Р В РЎвЂўР В Р’В±Р РЋР вЂ°Р В Р’ВµР В РЎвЂќР РЋРІР‚С™Р В Р’В°.
// Р В РЎСџР РЋР вЂљР В РЎвЂўР В Р вЂ Р В Р’ВµР РЋР вЂљР РЋР РЏР В Р’ВµР РЋРІР‚С™, Р РЋРІР‚РЋР РЋРІР‚С™Р В РЎвЂў Р РЋР С“Р В РЎвЂўР РЋРІР‚В¦Р РЋР вЂљР В Р’В°Р В Р вЂ¦Р В Р’ВµР В Р вЂ¦Р В РЎвЂР В Р’Вµ Р В РўвЂР В Р’В°Р В Р вЂ¦Р В Р вЂ¦Р РЋРІР‚в„–Р РЋРІР‚В¦ Р В Р’В·Р В Р’В°Р В РЎвЂ”Р В РЎвЂР РЋР С“Р РЋРІР‚в„–Р В Р вЂ Р В Р’В°Р В Р’ВµР РЋРІР‚С™ Р В РЎвЂўР В Р’В±Р В Р вЂ¦Р В РЎвЂўР В Р вЂ Р В Р’В»Р РЋРІР‚ВР В Р вЂ¦Р В Р вЂ¦Р В РЎвЂўР В Р’Вµ Р В РЎвЂ”Р В РЎвЂўР В Р’В»Р В РЎвЂўР В Р’В¶Р В Р’ВµР В Р вЂ¦Р В РЎвЂР В Р’Вµ Р В РЎвЂќР В РЎвЂўР РЋР С“Р В РЎВР В РЎвЂР РЋРІР‚РЋР В Р’ВµР РЋР С“Р В РЎвЂќР В РЎвЂўР В РЎвЂ“Р В РЎвЂў Р В РЎвЂўР В Р’В±Р РЋР вЂ°Р В Р’ВµР В РЎвЂќР РЋРІР‚С™Р В Р’В°.
func dockingEventsContainKind(events []game.DockingEvent, kind string) bool {
	for _, event := range events {
		if event.Kind == kind {
			return true
		}
	}
	return false
}

// Р В РЎСџР РЋР вЂљР В РЎвЂўР В Р вЂ Р В Р’ВµР РЋР вЂљР РЋР РЏР В Р’ВµР РЋРІР‚С™, Р РЋРІР‚РЋР РЋРІР‚С™Р В РЎвЂў Р РЋР С“Р В РЎвЂўР РЋРІР‚В¦Р РЋР вЂљР В Р’В°Р В Р вЂ¦Р В Р’ВµР В Р вЂ¦Р В РЎвЂР В Р’Вµ Р В РўвЂР В Р’В°Р В Р вЂ¦Р В Р вЂ¦Р РЋРІР‚в„–Р РЋРІР‚В¦ Р В Р’В·Р В Р’В°Р В РЎвЂ”Р В РЎвЂР РЋР С“Р РЋРІР‚в„–Р В Р вЂ Р В Р’В°Р В Р’ВµР РЋРІР‚С™ Р В РЎвЂўР В Р’В±Р В Р вЂ¦Р В РЎвЂўР В Р вЂ Р В Р’В»Р РЋРІР‚ВР В Р вЂ¦Р В Р вЂ¦Р В РЎвЂўР В Р’Вµ Р В РЎвЂ”Р В РЎвЂўР В Р’В»Р В РЎвЂўР В Р’В¶Р В Р’ВµР В Р вЂ¦Р В РЎвЂР В Р’Вµ Р В РЎвЂќР В РЎвЂўР РЋР С“Р В РЎВР В РЎвЂР РЋРІР‚РЋР В Р’ВµР РЋР С“Р В РЎвЂќР В РЎвЂўР В РЎвЂ“Р В РЎвЂў Р В РЎвЂўР В Р’В±Р РЋР вЂ°Р В Р’ВµР В РЎвЂќР РЋРІР‚С™Р В Р’В°.
func addDrillRayTestData(t *testing.T, serverData *world.Data) {
	t.Helper()
	serverData.ItemTypes.Items[1].IsPilotInstrument = true

	serverData.CosmicObjectTypes.Items[6] = &data.CosmicObjectType{ID: 6, TitleRu: "Луч", TitleEn: "Ray", Acronym: "Ray", Movable: true, Rotatable: true}
	serverData.CosmicObjectTypes.MaxID = 6
	serverData.CosmicObjectModels.Items[900] = &data.CosmicObjectModel{
		ID:                 900,
		TitleRu:            "Луч простого бура",
		TitleEn:            "Drill Ray",
		Acronym:            "DrillRay",
		TextureWidth:       8,
		TextureHeight:      500,
		TextureBodyOriginX: 4,
		TextureBodyOriginY: 250,
		TextureBodyWidth:   8,
		TextureBodyLength:  500,
		TextureScale:       0.95,
		CosmicObjectTypeID: 6,
	}
	serverData.ItemModels.Items[700] = &data.ItemModel{
		ID:         700,
		TitleRu:    "Простой бур",
		TitleEn:    "Simple Drill",
		Acronym:    "SimpleDrill",
		ItemTypeID: 1,
		Range:      500,
	}
	serverData.EquipmentGroups.Items[700] = &data.EquipmentGroup{
		ID:                   700,
		CosmicObjectID:       1,
		Title:                "Simple Drill",
		EquipmentItemModelID: 700,
		Count:                1,
		EnabledCount:         1,
		Enabled:              true,
	}

	if err := serverData.CosmicObjectTypes.RebuildIndexes(); err != nil {
		t.Fatal(err)
	}
	if err := serverData.CosmicObjectModels.RebuildIndexes(); err != nil {
		t.Fatal(err)
	}
	if err := serverData.ItemTypes.RebuildIndexes(); err != nil {
		t.Fatal(err)
	}
	if err := serverData.ItemModels.RebuildIndexes(); err != nil {
		t.Fatal(err)
	}
	if err := serverData.EquipmentGroups.RebuildIndexes(); err != nil {
		t.Fatal(err)
	}
}

// Р В РЎСџР РЋР вЂљР В РЎвЂўР В Р вЂ Р В Р’ВµР РЋР вЂљР РЋР РЏР В Р’ВµР РЋРІР‚С™, Р РЋРІР‚РЋР РЋРІР‚С™Р В РЎвЂў Р РЋР С“Р В РЎвЂўР РЋРІР‚В¦Р РЋР вЂљР В Р’В°Р В Р вЂ¦Р В Р’ВµР В Р вЂ¦Р В РЎвЂР В Р’Вµ Р В РўвЂР В Р’В°Р В Р вЂ¦Р В Р вЂ¦Р РЋРІР‚в„–Р РЋРІР‚В¦ Р В Р’В·Р В Р’В°Р В РЎвЂ”Р В РЎвЂР РЋР С“Р РЋРІР‚в„–Р В Р вЂ Р В Р’В°Р В Р’ВµР РЋРІР‚С™ Р В РЎвЂўР В Р’В±Р В Р вЂ¦Р В РЎвЂўР В Р вЂ Р В Р’В»Р РЋРІР‚ВР В Р вЂ¦Р В Р вЂ¦Р В РЎвЂўР В Р’Вµ Р В РЎвЂ”Р В РЎвЂўР В Р’В»Р В РЎвЂўР В Р’В¶Р В Р’ВµР В Р вЂ¦Р В РЎвЂР В Р’Вµ Р В РЎвЂќР В РЎвЂўР РЋР С“Р В РЎВР В РЎвЂР РЋРІР‚РЋР В Р’ВµР РЋР С“Р В РЎвЂќР В РЎвЂўР В РЎвЂ“Р В РЎвЂў Р В РЎвЂўР В Р’В±Р РЋР вЂ°Р В Р’ВµР В РЎвЂќР РЋРІР‚С™Р В Р’В°.
// Р В РЎСџР РЋР вЂљР В РЎвЂўР В Р вЂ Р В Р’ВµР РЋР вЂљР РЋР РЏР В Р’ВµР РЋРІР‚С™, Р РЋРІР‚РЋР РЋРІР‚С™Р В РЎвЂў Р РЋР С“Р В РЎвЂўР РЋРІР‚В¦Р РЋР вЂљР В Р’В°Р В Р вЂ¦Р В Р’ВµР В Р вЂ¦Р В РЎвЂР В Р’Вµ Р В РўвЂР В Р’В°Р В Р вЂ¦Р В Р вЂ¦Р РЋРІР‚в„–Р РЋРІР‚В¦ Р В Р’В·Р В Р’В°Р В РЎвЂ”Р В РЎвЂР РЋР С“Р РЋРІР‚в„–Р В Р вЂ Р В Р’В°Р В Р’ВµР РЋРІР‚С™ Р В РЎвЂўР В Р’В±Р В Р вЂ¦Р В РЎвЂўР В Р вЂ Р В Р’В»Р РЋРІР‚ВР В Р вЂ¦Р В Р вЂ¦Р В РЎвЂўР В Р’Вµ Р В РЎвЂ”Р В РЎвЂўР В Р’В»Р В РЎвЂўР В Р’В¶Р В Р’ВµР В Р вЂ¦Р В РЎвЂР В Р’Вµ Р В РЎвЂќР В РЎвЂўР РЋР С“Р В РЎВР В РЎвЂР РЋРІР‚РЋР В Р’ВµР РЋР С“Р В РЎвЂќР В РЎвЂўР В РЎвЂ“Р В РЎвЂў Р В РЎвЂўР В Р’В±Р РЋР вЂ°Р В Р’ВµР В РЎвЂќР РЋРІР‚С™Р В Р’В°.
func addWeaponTestData(t *testing.T, serverData *world.Data) {
	t.Helper()
	serverData.ItemTypes.Items[1].IsPilotInstrument = true
	serverData.CosmicObjectTypes.Items[5] = &data.CosmicObjectType{ID: 5, TitleRu: "Снаряд", TitleEn: "Projectile", Acronym: "Projectile", Movable: true, Rotatable: true}
	serverData.CosmicObjectTypes.MaxID = 5
	serverData.CosmicObjectModels.Items[901] = &data.CosmicObjectModel{
		ID:                 901,
		TitleRu:            "Баллистический снаряд",
		TitleEn:            "Ballistic Projectile",
		Acronym:            "BallisticProjectile",
		TextureWidth:       12,
		TextureHeight:      32,
		TextureBodyOriginX: 6,
		TextureBodyOriginY: 16,
		TextureBodyWidth:   12,
		TextureBodyLength:  32,
		TextureScale:       1,
		CosmicObjectTypeID: 5,
		BodyLength:         32,
		BodyWidth:          12,
	}
	serverData.ItemModels.Items[302].Range = 500
	serverData.ItemModels.Items[302].Damage = 120
	serverData.ItemModels.Items[302].FiringRate = 1
	serverData.ItemModels.Items[302].ProjectileSpeed = 100
	serverData.EquipmentGroups.Items[600] = &data.EquipmentGroup{
		ID:                   600,
		CosmicObjectID:       1,
		Title:                "Cannon",
		EquipmentItemModelID: 302,
		Count:                1,
		EnabledCount:         1,
		Enabled:              true,
	}
	if err := serverData.CosmicObjectTypes.RebuildIndexes(); err != nil {
		t.Fatal(err)
	}
	if err := serverData.CosmicObjectModels.RebuildIndexes(); err != nil {
		t.Fatal(err)
	}
	if err := serverData.ItemTypes.RebuildIndexes(); err != nil {
		t.Fatal(err)
	}
	if err := serverData.ItemModels.RebuildIndexes(); err != nil {
		t.Fatal(err)
	}
	if err := serverData.EquipmentGroups.RebuildIndexes(); err != nil {
		t.Fatal(err)
	}
}

func dockingEventsContainMessage(events []game.DockingEvent, message string) bool {
	for _, event := range events {
		if event.Message != "" && (strings.Contains(event.Message, message) || strings.Contains(message, event.Message)) {
			return true
		}
	}
	return false
}

// Р В РЎСџР РЋР вЂљР В РЎвЂўР В Р вЂ Р В Р’ВµР РЋР вЂљР РЋР РЏР В Р’ВµР РЋРІР‚С™, Р РЋРІР‚РЋР РЋРІР‚С™Р В РЎвЂў Р РЋР С“Р В РЎвЂўР РЋРІР‚В¦Р РЋР вЂљР В Р’В°Р В Р вЂ¦Р В Р’ВµР В Р вЂ¦Р В РЎвЂР В Р’Вµ Р В РўвЂР В Р’В°Р В Р вЂ¦Р В Р вЂ¦Р РЋРІР‚в„–Р РЋРІР‚В¦ Р В Р’В·Р В Р’В°Р В РЎвЂ”Р В РЎвЂР РЋР С“Р РЋРІР‚в„–Р В Р вЂ Р В Р’В°Р В Р’ВµР РЋРІР‚С™ Р В РЎвЂўР В Р’В±Р В Р вЂ¦Р В РЎвЂўР В Р вЂ Р В Р’В»Р РЋРІР‚ВР В Р вЂ¦Р В Р вЂ¦Р В РЎвЂўР В Р’Вµ Р В РЎвЂ”Р В РЎвЂўР В Р’В»Р В РЎвЂўР В Р’В¶Р В Р’ВµР В Р вЂ¦Р В РЎвЂР В Р’Вµ Р В РЎвЂќР В РЎвЂўР РЋР С“Р В РЎВР В РЎвЂР РЋРІР‚РЋР В Р’ВµР РЋР С“Р В РЎвЂќР В РЎвЂўР В РЎвЂ“Р В РЎвЂў Р В РЎвЂўР В Р’В±Р РЋР вЂ°Р В Р’ВµР В РЎвЂќР РЋРІР‚С™Р В Р’В°.
func exchangeEventsContainNotificationMessage(events []game.ExchangeEvent, message string) bool {
	for _, event := range events {
		if event.Kind == "exchangeNotification" && event.Message == message {
			return true
		}
	}
	return false
}

func addRawReferenceItem(t *testing.T, table *storage.RawReferenceTable, id int64, item any) {
	t.Helper()
	raw, err := json.Marshal(item)
	if err != nil {
		t.Fatal(err)
	}
	table.Items[fmt.Sprintf("%d", id)] = raw
	if table.MaxID < id {
		table.MaxID = id
	}
}

// Р В РЎСџР РЋР вЂљР В РЎвЂўР В Р вЂ Р В Р’ВµР РЋР вЂљР РЋР РЏР В Р’ВµР РЋРІР‚С™, Р РЋРІР‚РЋР РЋРІР‚С™Р В РЎвЂў Р РЋР С“Р В РЎвЂўР РЋРІР‚В¦Р РЋР вЂљР В Р’В°Р В Р вЂ¦Р В Р’ВµР В Р вЂ¦Р В РЎвЂР В Р’Вµ Р В РўвЂР В Р’В°Р В Р вЂ¦Р В Р вЂ¦Р РЋРІР‚в„–Р РЋРІР‚В¦ Р В Р’В·Р В Р’В°Р В РЎвЂ”Р В РЎвЂР РЋР С“Р РЋРІР‚в„–Р В Р вЂ Р В Р’В°Р В Р’ВµР РЋРІР‚С™ Р В РЎвЂўР В Р’В±Р В Р вЂ¦Р В РЎвЂўР В Р вЂ Р В Р’В»Р РЋРІР‚ВР В Р вЂ¦Р В Р вЂ¦Р В РЎвЂўР В Р’Вµ Р В РЎвЂ”Р В РЎвЂўР В Р’В»Р В РЎвЂўР В Р’В¶Р В Р’ВµР В Р вЂ¦Р В РЎвЂР В Р’Вµ Р В РЎвЂќР В РЎвЂўР РЋР С“Р В РЎВР В РЎвЂР РЋРІР‚РЋР В Р’ВµР РЋР С“Р В РЎвЂќР В РЎвЂўР В РЎвЂ“Р В РЎвЂў Р В РЎвЂўР В Р’В±Р РЋР вЂ°Р В Р’ВµР В РЎвЂќР РЋРІР‚С™Р В Р’В°.
func closeWorldFloat(t *testing.T, actual float64, expected float64) {
	t.Helper()

	if math.Abs(actual-expected) > 0.000001 {
		t.Fatalf("got %v, want %v", actual, expected)
	}
}

// Р В РЎСџР РЋР вЂљР В РЎвЂўР В Р вЂ Р В Р’ВµР РЋР вЂљР РЋР РЏР В Р’ВµР РЋРІР‚С™, Р РЋРІР‚РЋР РЋРІР‚С™Р В РЎвЂў Р РЋР С“Р В РЎвЂўР РЋРІР‚В¦Р РЋР вЂљР В Р’В°Р В Р вЂ¦Р В Р’ВµР В Р вЂ¦Р В РЎвЂР В Р’Вµ Р В РўвЂР В Р’В°Р В Р вЂ¦Р В Р вЂ¦Р РЋРІР‚в„–Р РЋРІР‚В¦ Р В Р’В·Р В Р’В°Р В РЎвЂ”Р В РЎвЂР РЋР С“Р РЋРІР‚в„–Р В Р вЂ Р В Р’В°Р В Р’ВµР РЋРІР‚С™ Р В РЎвЂўР В Р’В±Р В Р вЂ¦Р В РЎвЂўР В Р вЂ Р В Р’В»Р РЋРІР‚ВР В Р вЂ¦Р В Р вЂ¦Р В РЎвЂўР В Р’Вµ Р В РЎвЂ”Р В РЎвЂўР В Р’В»Р В РЎвЂўР В Р’В¶Р В Р’ВµР В Р вЂ¦Р В РЎвЂР В Р’Вµ Р В РЎвЂќР В РЎвЂўР РЋР С“Р В РЎВР В РЎвЂР РЋРІР‚РЋР В Р’ВµР РЋР С“Р В РЎвЂќР В РЎвЂўР В РЎвЂ“Р В РЎвЂў Р В РЎвЂўР В Р’В±Р РЋР вЂ°Р В Р’ВµР В РЎвЂќР РЋРІР‚С™Р В Р’В°.
func boolPointer(value bool) *bool {
	return &value
}

// Р В РЎСџР РЋР вЂљР В РЎвЂўР В Р вЂ Р В Р’ВµР РЋР вЂљР РЋР РЏР В Р’ВµР РЋРІР‚С™, Р РЋРІР‚РЋР РЋРІР‚С™Р В РЎвЂў Р РЋР С“Р В РЎвЂўР РЋРІР‚В¦Р РЋР вЂљР В Р’В°Р В Р вЂ¦Р В Р’ВµР В Р вЂ¦Р В РЎвЂР В Р’Вµ Р В РўвЂР В Р’В°Р В Р вЂ¦Р В Р вЂ¦Р РЋРІР‚в„–Р РЋРІР‚В¦ Р В Р’В·Р В Р’В°Р В РЎвЂ”Р В РЎвЂР РЋР С“Р РЋРІР‚в„–Р В Р вЂ Р В Р’В°Р В Р’ВµР РЋРІР‚С™ Р В РЎвЂўР В Р’В±Р В Р вЂ¦Р В РЎвЂўР В Р вЂ Р В Р’В»Р РЋРІР‚ВР В Р вЂ¦Р В Р вЂ¦Р В РЎвЂўР В Р’Вµ Р В РЎвЂ”Р В РЎвЂўР В Р’В»Р В РЎвЂўР В Р’В¶Р В Р’ВµР В Р вЂ¦Р В РЎвЂР В Р’Вµ Р В РЎвЂќР В РЎвЂўР РЋР С“Р В РЎВР В РЎвЂР РЋРІР‚РЋР В Р’ВµР РЋР С“Р В РЎвЂќР В РЎвЂўР В РЎвЂ“Р В РЎвЂў Р В РЎвЂўР В Р’В±Р РЋР вЂ°Р В Р’ВµР В РЎвЂќР РЋРІР‚С™Р В Р’В°.
func stringPointer(value string) *string {
	return &value
}

// Р В РЎСџР РЋР вЂљР В РЎвЂўР В Р вЂ Р В Р’ВµР РЋР вЂљР РЋР РЏР В Р’ВµР РЋРІР‚С™, Р РЋРІР‚РЋР РЋРІР‚С™Р В РЎвЂў Р РЋР С“Р В РЎвЂўР РЋРІР‚В¦Р РЋР вЂљР В Р’В°Р В Р вЂ¦Р В Р’ВµР В Р вЂ¦Р В РЎвЂР В Р’Вµ Р В РўвЂР В Р’В°Р В Р вЂ¦Р В Р вЂ¦Р РЋРІР‚в„–Р РЋРІР‚В¦ Р В Р’В·Р В Р’В°Р В РЎвЂ”Р В РЎвЂР РЋР С“Р РЋРІР‚в„–Р В Р вЂ Р В Р’В°Р В Р’ВµР РЋРІР‚С™ Р В РЎвЂўР В Р’В±Р В Р вЂ¦Р В РЎвЂўР В Р вЂ Р В Р’В»Р РЋРІР‚ВР В Р вЂ¦Р В Р вЂ¦Р В РЎвЂўР В Р’Вµ Р В РЎвЂ”Р В РЎвЂўР В Р’В»Р В РЎвЂўР В Р’В¶Р В Р’ВµР В Р вЂ¦Р В РЎвЂР В Р’Вµ Р В РЎвЂќР В РЎвЂўР РЋР С“Р В РЎВР В РЎвЂР РЋРІР‚РЋР В Р’ВµР РЋР С“Р В РЎвЂќР В РЎвЂўР В РЎвЂ“Р В РЎвЂў Р В РЎвЂўР В Р’В±Р РЋР вЂ°Р В Р’ВµР В РЎвЂќР РЋРІР‚С™Р В Р’В°.
func int64Pointer(value int64) *int64 {
	return &value
}

// Р В РЎСџР РЋР вЂљР В РЎвЂўР В Р вЂ Р В Р’ВµР РЋР вЂљР РЋР РЏР В Р’ВµР РЋРІР‚С™, Р РЋРІР‚РЋР РЋРІР‚С™Р В РЎвЂў Р РЋР С“Р В РЎвЂўР РЋРІР‚В¦Р РЋР вЂљР В Р’В°Р В Р вЂ¦Р В Р’ВµР В Р вЂ¦Р В РЎвЂР В Р’Вµ Р В РўвЂР В Р’В°Р В Р вЂ¦Р В Р вЂ¦Р РЋРІР‚в„–Р РЋРІР‚В¦ Р В Р’В·Р В Р’В°Р В РЎвЂ”Р В РЎвЂР РЋР С“Р РЋРІР‚в„–Р В Р вЂ Р В Р’В°Р В Р’ВµР РЋРІР‚С™ Р В РЎвЂўР В Р’В±Р В Р вЂ¦Р В РЎвЂўР В Р вЂ Р В Р’В»Р РЋРІР‚ВР В Р вЂ¦Р В Р вЂ¦Р В РЎвЂўР В Р’Вµ Р В РЎвЂ”Р В РЎвЂўР В Р’В»Р В РЎвЂўР В Р’В¶Р В Р’ВµР В Р вЂ¦Р В РЎвЂР В Р’Вµ Р В РЎвЂќР В РЎвЂўР РЋР С“Р В РЎВР В РЎвЂР РЋРІР‚РЋР В Р’ВµР РЋР С“Р В РЎвЂќР В РЎвЂўР В РЎвЂ“Р В РЎвЂў Р В РЎвЂўР В Р’В±Р РЋР вЂ°Р В Р’ВµР В РЎвЂќР РЋРІР‚С™Р В Р’В°.
func findEquipmentGroupInSnapshot(snapshot game.Snapshot, groupID int64) (data.EquipmentGroup, bool) {
	for _, group := range snapshot.EquipmentGroups {
		if group.ID == groupID {
			return group, true
		}
	}

	return data.EquipmentGroup{}, false
}

// Р В РЎСџР РЋР вЂљР В РЎвЂўР В Р вЂ Р В Р’ВµР РЋР вЂљР РЋР РЏР В Р’ВµР РЋРІР‚С™, Р РЋРІР‚РЋР РЋРІР‚С™Р В РЎвЂў Р РЋР С“Р В РЎвЂўР РЋРІР‚В¦Р РЋР вЂљР В Р’В°Р В Р вЂ¦Р В Р’ВµР В Р вЂ¦Р В РЎвЂР В Р’Вµ Р В РўвЂР В Р’В°Р В Р вЂ¦Р В Р вЂ¦Р РЋРІР‚в„–Р РЋРІР‚В¦ Р В Р’В·Р В Р’В°Р В РЎвЂ”Р В РЎвЂР РЋР С“Р РЋРІР‚в„–Р В Р вЂ Р В Р’В°Р В Р’ВµР РЋРІР‚С™ Р В РЎвЂўР В Р’В±Р В Р вЂ¦Р В РЎвЂўР В Р вЂ Р В Р’В»Р РЋРІР‚ВР В Р вЂ¦Р В Р вЂ¦Р В РЎвЂўР В Р’Вµ Р В РЎвЂ”Р В РЎвЂўР В Р’В»Р В РЎвЂўР В Р’В¶Р В Р’ВµР В Р вЂ¦Р В РЎвЂР В Р’Вµ Р В РЎвЂќР В РЎвЂўР РЋР С“Р В РЎВР В РЎвЂР РЋРІР‚РЋР В Р’ВµР РЋР С“Р В РЎвЂќР В РЎвЂўР В РЎвЂ“Р В РЎвЂў Р В РЎвЂўР В Р’В±Р РЋР вЂ°Р В Р’ВµР В РЎвЂќР РЋРІР‚С™Р В Р’В°.
// Р В РЎСџР РЋР вЂљР В РЎвЂўР В Р вЂ Р В Р’ВµР РЋР вЂљР РЋР РЏР В Р’ВµР РЋРІР‚С™, Р РЋРІР‚РЋР РЋРІР‚С™Р В РЎвЂў Р РЋР С“Р В РЎвЂўР РЋРІР‚В¦Р РЋР вЂљР В Р’В°Р В Р вЂ¦Р В Р’ВµР В Р вЂ¦Р В РЎвЂР В Р’Вµ Р В РўвЂР В Р’В°Р В Р вЂ¦Р В Р вЂ¦Р РЋРІР‚в„–Р РЋРІР‚В¦ Р В Р’В·Р В Р’В°Р В РЎвЂ”Р В РЎвЂР РЋР С“Р РЋРІР‚в„–Р В Р вЂ Р В Р’В°Р В Р’ВµР РЋРІР‚С™ Р В РЎвЂўР В Р’В±Р В Р вЂ¦Р В РЎвЂўР В Р вЂ Р В Р’В»Р РЋРІР‚ВР В Р вЂ¦Р В Р вЂ¦Р В РЎвЂўР В Р’Вµ Р В РЎвЂ”Р В РЎвЂўР В Р’В»Р В РЎвЂўР В Р’В¶Р В Р’ВµР В Р вЂ¦Р В РЎвЂР В Р’Вµ Р В РЎвЂќР В РЎвЂўР РЋР С“Р В РЎВР В РЎвЂР РЋРІР‚РЋР В Р’ВµР РЋР С“Р В РЎвЂќР В РЎвЂўР В РЎвЂ“Р В РЎвЂў Р В РЎвЂўР В Р’В±Р РЋР вЂ°Р В Р’ВµР В РЎвЂќР РЋРІР‚С™Р В Р’В°.
func addTestGenerator(t *testing.T, serverData world.Data, cosmicObjectID int64) *data.EquipmentGroup {
	t.Helper()

	serverData.ItemModels.Items[104] = &data.ItemModel{
		ID:                   104,
		TitleRu:              "Generator",
		TitleEn:              "Generator",
		Acronym:              "Generator",
		ItemTypeID:           1,
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

// Р В РЎСџР РЋР вЂљР В РЎвЂўР В Р вЂ Р В Р’ВµР РЋР вЂљР РЋР РЏР В Р’ВµР РЋРІР‚С™, Р РЋРІР‚РЋР РЋРІР‚С™Р В РЎвЂў Р РЋР С“Р В РЎвЂўР РЋРІР‚В¦Р РЋР вЂљР В Р’В°Р В Р вЂ¦Р В Р’ВµР В Р вЂ¦Р В РЎвЂР В Р’Вµ Р В РўвЂР В Р’В°Р В Р вЂ¦Р В Р вЂ¦Р РЋРІР‚в„–Р РЋРІР‚В¦ Р В Р’В·Р В Р’В°Р В РЎвЂ”Р В РЎвЂР РЋР С“Р РЋРІР‚в„–Р В Р вЂ Р В Р’В°Р В Р’ВµР РЋРІР‚С™ Р В РЎвЂўР В Р’В±Р В Р вЂ¦Р В РЎвЂўР В Р вЂ Р В Р’В»Р РЋРІР‚ВР В Р вЂ¦Р В Р вЂ¦Р В РЎвЂўР В Р’Вµ Р В РЎвЂ”Р В РЎвЂўР В Р’В»Р В РЎвЂўР В Р’В¶Р В Р’ВµР В Р вЂ¦Р В РЎвЂР В Р’Вµ Р В РЎвЂќР В РЎвЂўР РЋР С“Р В РЎВР В РЎвЂР РЋРІР‚РЋР В Р’ВµР РЋР С“Р В РЎвЂќР В РЎвЂўР В РЎвЂ“Р В РЎвЂў Р В РЎвЂўР В Р’В±Р РЋР вЂ°Р В Р’ВµР В РЎвЂќР РЋРІР‚С™Р В Р’В°.
func addTestPowerProducer(t *testing.T, serverData world.Data, cosmicObjectID int64) *data.EquipmentGroup {
	t.Helper()

	serverData.ItemModels.Items[105] = &data.ItemModel{
		ID:              105,
		TitleRu:         "Power Producer",
		TitleEn:         "Power Producer",
		Acronym:         "PowerProducer",
		ItemTypeID:      1,
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

// Р В РЎСџР РЋР вЂљР В РЎвЂўР В Р вЂ Р В Р’ВµР РЋР вЂљР РЋР РЏР В Р’ВµР РЋРІР‚С™, Р РЋРІР‚РЋР РЋРІР‚С™Р В РЎвЂў Р РЋР С“Р В РЎвЂўР РЋРІР‚В¦Р РЋР вЂљР В Р’В°Р В Р вЂ¦Р В Р’ВµР В Р вЂ¦Р В РЎвЂР В Р’Вµ Р В РўвЂР В Р’В°Р В Р вЂ¦Р В Р вЂ¦Р РЋРІР‚в„–Р РЋРІР‚В¦ Р В Р’В·Р В Р’В°Р В РЎвЂ”Р В РЎвЂР РЋР С“Р РЋРІР‚в„–Р В Р вЂ Р В Р’В°Р В Р’ВµР РЋРІР‚С™ Р В РЎвЂўР В Р’В±Р В Р вЂ¦Р В РЎвЂўР В Р вЂ Р В Р’В»Р РЋРІР‚ВР В Р вЂ¦Р В Р вЂ¦Р В РЎвЂўР В Р’Вµ Р В РЎвЂ”Р В РЎвЂўР В Р’В»Р В РЎвЂўР В Р’В¶Р В Р’ВµР В Р вЂ¦Р В РЎвЂР В Р’Вµ Р В РЎвЂќР В РЎвЂўР РЋР С“Р В РЎВР В РЎвЂР РЋРІР‚РЋР В Р’ВµР РЋР С“Р В РЎвЂќР В РЎвЂўР В РЎвЂ“Р В РЎвЂў Р В РЎвЂўР В Р’В±Р РЋР вЂ°Р В Р’ВµР В РЎвЂќР РЋРІР‚С™Р В Р’В°.
func addTestRobots(t *testing.T, serverData world.Data, cosmicObjectID int64) *data.EquipmentGroup {
	t.Helper()

	serverData.ItemTypes.Items[20] = &data.ItemType{ID: 20, TitleRu: "Robot", TitleEn: "Robot", Acronym: "Robot", CountMustBeInteger: true}
	serverData.ItemModels.Items[404] = &data.ItemModel{ID: 404, TitleRu: "Robot", TitleEn: "Robot", Acronym: "Robot", ItemTypeID: 20, ConsumingPower: 10}
	if err := serverData.ItemTypes.RebuildIndexes(); err != nil {
		t.Fatal(err)
	}
	if err := serverData.ItemModels.RebuildIndexes(); err != nil {
		t.Fatal(err)
	}
	group, err := serverData.EquipmentGroups.Add(&data.EquipmentGroup{
		CosmicObjectID:       cosmicObjectID,
		Title:                "Robots",
		EquipmentItemModelID: 404,
		Count:                10,
		EnabledCount:         10,
		Enabled:              true,
	})
	if err != nil {
		t.Fatal(err)
	}
	return group
}

// Р В РЎСџР РЋР вЂљР В РЎвЂўР В Р вЂ Р В Р’ВµР РЋР вЂљР РЋР РЏР В Р’ВµР РЋРІР‚С™, Р РЋРІР‚РЋР РЋРІР‚С™Р В РЎвЂў Р РЋР С“Р В РЎвЂўР РЋРІР‚В¦Р РЋР вЂљР В Р’В°Р В Р вЂ¦Р В Р’ВµР В Р вЂ¦Р В РЎвЂР В Р’Вµ Р В РўвЂР В Р’В°Р В Р вЂ¦Р В Р вЂ¦Р РЋРІР‚в„–Р РЋРІР‚В¦ Р В Р’В·Р В Р’В°Р В РЎвЂ”Р В РЎвЂР РЋР С“Р РЋРІР‚в„–Р В Р вЂ Р В Р’В°Р В Р’ВµР РЋРІР‚С™ Р В РЎвЂўР В Р’В±Р В Р вЂ¦Р В РЎвЂўР В Р вЂ Р В Р’В»Р РЋРІР‚ВР В Р вЂ¦Р В Р вЂ¦Р В РЎвЂўР В Р’Вµ Р В РЎвЂ”Р В РЎвЂўР В Р’В»Р В РЎвЂўР В Р’В¶Р В Р’ВµР В Р вЂ¦Р В РЎвЂР В Р’Вµ Р В РЎвЂќР В РЎвЂўР РЋР С“Р В РЎВР В РЎвЂР РЋРІР‚РЋР В Р’ВµР РЋР С“Р В РЎвЂќР В РЎвЂўР В РЎвЂ“Р В РЎвЂў Р В РЎвЂўР В Р’В±Р РЋР вЂ°Р В Р’ВµР В РЎвЂќР РЋРІР‚С™Р В Р’В°.
func addTestDeconstructor(t *testing.T, serverData world.Data, cosmicObjectID int64) *data.EquipmentGroup {
	t.Helper()

	serverData.ItemTypes.Items[12] = &data.ItemType{ID: 12, TitleRu: "Deconstructor", TitleEn: "Deconstructor", Acronym: "Deconstructor", CountMustBeInteger: true}
	serverData.ItemModels.Items[405] = &data.ItemModel{ID: 405, TitleRu: "Deconstructor", TitleEn: "Deconstructor", Acronym: "Deconstructor", ItemTypeID: 12, ConsumingPower: 10}
	if err := serverData.ItemTypes.RebuildIndexes(); err != nil {
		t.Fatal(err)
	}
	if err := serverData.ItemModels.RebuildIndexes(); err != nil {
		t.Fatal(err)
	}
	group, err := serverData.EquipmentGroups.Add(&data.EquipmentGroup{
		CosmicObjectID:       cosmicObjectID,
		Title:                "Deconstructor",
		EquipmentItemModelID: 405,
		Count:                1,
		EnabledCount:         1,
		Enabled:              true,
	})
	if err != nil {
		t.Fatal(err)
	}
	return group
}

// Р В РЎСџР РЋР вЂљР В РЎвЂўР В Р вЂ Р В Р’ВµР РЋР вЂљР РЋР РЏР В Р’ВµР РЋРІР‚С™, Р РЋРІР‚РЋР РЋРІР‚С™Р В РЎвЂў Р РЋР С“Р В РЎвЂўР РЋРІР‚В¦Р РЋР вЂљР В Р’В°Р В Р вЂ¦Р В Р’ВµР В Р вЂ¦Р В РЎвЂР В Р’Вµ Р В РўвЂР В Р’В°Р В Р вЂ¦Р В Р вЂ¦Р РЋРІР‚в„–Р РЋРІР‚В¦ Р В Р’В·Р В Р’В°Р В РЎвЂ”Р В РЎвЂР РЋР С“Р РЋРІР‚в„–Р В Р вЂ Р В Р’В°Р В Р’ВµР РЋРІР‚С™ Р В РЎвЂўР В Р’В±Р В Р вЂ¦Р В РЎвЂўР В Р вЂ Р В Р’В»Р РЋРІР‚ВР В Р вЂ¦Р В Р вЂ¦Р В РЎвЂўР В Р’Вµ Р В РЎвЂ”Р В РЎвЂўР В Р’В»Р В РЎвЂўР В Р’В¶Р В Р’ВµР В Р вЂ¦Р В РЎвЂР В Р’Вµ Р В РЎвЂќР В РЎвЂўР РЋР С“Р В РЎВР В РЎвЂР РЋРІР‚РЋР В Р’ВµР РЋР С“Р В РЎвЂќР В РЎвЂўР В РЎвЂ“Р В РЎвЂў Р В РЎвЂўР В Р’В±Р РЋР вЂ°Р В Р’ВµР В РЎвЂќР РЋРІР‚С™Р В Р’В°.
func TestApplyControlPanelObjectUpdateChangesControlledObjectAndAcknowledgesMutation(t *testing.T) {
	serverData := testWorldData(t)
	gameWorld := world.New(1, serverData)
	if _, ok := gameWorld.ConnectAccount(1); !ok {
		t.Fatal("account was not connected")
	}

	err := gameWorld.ApplyControlPanelObjectUpdate(1, "session-1", 4, world.ControlPanelObjectUpdate{
		Enabled: boolPointer(false),
		Title:   stringPointer("Р В РЎСљР В РЎвЂўР В Р вЂ Р РЋРІР‚в„–Р В РІвЂћвЂ“ Р В РЎвЂќР В РЎвЂўР РЋР вЂљР В Р’В°Р В Р’В±Р В Р’В»Р РЋР Р‰"),
	})

	if err != nil {
		t.Fatalf("apply object update: %v", err)
	}
	snapshot := gameWorld.SnapshotForAccount(1)
	object, ok := findCosmicObjectInSnapshot(snapshot, 1)
	if !ok {
		t.Fatal("controlled object not found")
	}
	if object.Enabled || object.Title != "Р В РЎСљР В РЎвЂўР В Р вЂ Р РЋРІР‚в„–Р В РІвЂћвЂ“ Р В РЎвЂќР В РЎвЂўР РЋР вЂљР В Р’В°Р В Р’В±Р В Р’В»Р РЋР Р‰" {
		t.Fatalf("object was not updated: %+v", object)
	}
	ack := gameWorld.ClientMutationAck(1, "session-1")
	if ack.SessionID != "session-1" || ack.LastAppliedSeq != 4 {
		t.Fatalf("mutation ack mismatch: %+v", ack)
	}
}

// Р В РЎСџР РЋР вЂљР В РЎвЂўР В Р вЂ Р В Р’ВµР РЋР вЂљР РЋР РЏР В Р’ВµР РЋРІР‚С™, Р РЋРІР‚РЋР РЋРІР‚С™Р В РЎвЂў Р РЋР С“Р В РЎвЂўР РЋРІР‚В¦Р РЋР вЂљР В Р’В°Р В Р вЂ¦Р В Р’ВµР В Р вЂ¦Р В РЎвЂР В Р’Вµ Р В РўвЂР В Р’В°Р В Р вЂ¦Р В Р вЂ¦Р РЋРІР‚в„–Р РЋРІР‚В¦ Р В Р’В·Р В Р’В°Р В РЎвЂ”Р В РЎвЂР РЋР С“Р РЋРІР‚в„–Р В Р вЂ Р В Р’В°Р В Р’ВµР РЋРІР‚С™ Р В РЎвЂўР В Р’В±Р В Р вЂ¦Р В РЎвЂўР В Р вЂ Р В Р’В»Р РЋРІР‚ВР В Р вЂ¦Р В Р вЂ¦Р В РЎвЂўР В Р’Вµ Р В РЎвЂ”Р В РЎвЂўР В Р’В»Р В РЎвЂўР В Р’В¶Р В Р’ВµР В Р вЂ¦Р В РЎвЂР В Р’Вµ Р В РЎвЂќР В РЎвЂўР РЋР С“Р В РЎВР В РЎвЂР РЋРІР‚РЋР В Р’ВµР РЋР С“Р В РЎвЂќР В РЎвЂўР В РЎвЂ“Р В РЎвЂў Р В РЎвЂўР В Р’В±Р РЋР вЂ°Р В Р’ВµР В РЎвЂќР РЋРІР‚С™Р В Р’В°.
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
		Title:            stringPointer("Renamed equipment"),
	})

	if err != nil {
		t.Fatalf("apply equipment update: %v", err)
	}
	if ownedGroup.Enabled || ownedGroup.EnabledCount != 1 || ownedGroup.Title != "Renamed equipment" {
		t.Fatalf("owned group was not updated: %+v", ownedGroup)
	}
	if err := gameWorld.ApplyControlPanelEquipmentUpdate(1, "session-1", 6, world.ControlPanelEquipmentUpdate{EquipmentGroupID: foreignGroup.ID, Enabled: boolPointer(false)}); err == nil {
		t.Fatal("foreign group update was accepted")
	}
	snapshot := gameWorld.SnapshotForAccount(1)
	group, ok := findEquipmentGroupInSnapshot(snapshot, ownedGroup.ID)
	if !ok || group.Enabled || group.EnabledCount != 1 || group.Title != "Renamed equipment" {
		t.Fatalf("snapshot group mismatch: %+v", group)
	}
}

// Р В РЎСџР РЋР вЂљР В РЎвЂўР В Р вЂ Р В Р’ВµР РЋР вЂљР РЋР РЏР В Р’ВµР РЋРІР‚С™, Р РЋРІР‚РЋР РЋРІР‚С™Р В РЎвЂў Р РЋР С“Р В РЎвЂўР РЋРІР‚В¦Р РЋР вЂљР В Р’В°Р В Р вЂ¦Р В Р’ВµР В Р вЂ¦Р В РЎвЂР В Р’Вµ Р В РўвЂР В Р’В°Р В Р вЂ¦Р В Р вЂ¦Р РЋРІР‚в„–Р РЋРІР‚В¦ Р В Р’В·Р В Р’В°Р В РЎвЂ”Р В РЎвЂР РЋР С“Р РЋРІР‚в„–Р В Р вЂ Р В Р’В°Р В Р’ВµР РЋРІР‚С™ Р В РЎвЂўР В Р’В±Р В Р вЂ¦Р В РЎвЂўР В Р вЂ Р В Р’В»Р РЋРІР‚ВР В Р вЂ¦Р В Р вЂ¦Р В РЎвЂўР В Р’Вµ Р В РЎвЂ”Р В РЎвЂўР В Р’В»Р В РЎвЂўР В Р’В¶Р В Р’ВµР В Р вЂ¦Р В РЎвЂР В Р’Вµ Р В РЎвЂќР В РЎвЂўР РЋР С“Р В РЎВР В РЎвЂР РЋРІР‚РЋР В Р’ВµР РЋР С“Р В РЎвЂќР В РЎвЂўР В РЎвЂ“Р В РЎвЂў Р В РЎвЂўР В Р’В±Р РЋР вЂ°Р В Р’ВµР В РЎвЂќР РЋРІР‚С™Р В Р’В°.
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
			1: {ID: 1, TitleRu: "Р В РЎв„ўР В РЎвЂўР РЋР вЂљР В Р’В°Р В Р’В±Р В Р’В»Р РЋР Р‰", TitleEn: "Ship", Acronym: "Ship"},
			2: {ID: 2, TitleRu: "Р В Р Р‹Р РЋРІР‚С™Р В Р’В°Р В Р вЂ¦Р РЋРІР‚В Р В РЎвЂР РЋР РЏ", TitleEn: "Station", Acronym: "Station"},
			3: {ID: 3, TitleRu: "Р В РЎвЂ™Р РЋР С“Р РЋРІР‚С™Р В Р’ВµР РЋР вЂљР В РЎвЂўР В РЎвЂР В РўвЂ", TitleEn: "Asteroid", Acronym: "Asteroid"},
		},
	}
	cosmicObjectModels := &data.CosmicObjectModels{
		MaxID: 4,
		Items: map[int64]*data.CosmicObjectModel{
			1: {ID: 1, TitleRu: "Р В РЎв„ўР В РЎвЂўР РЋР вЂљР В Р’В°Р В Р’В±Р В Р’В»Р РЋР Р‰", TitleEn: "Ship", Acronym: "ship_bat", CosmicObjectTypeID: 1, TextureScale: 4, TextureBodyWidth: 27, TextureBodyLength: 63, Mass: 7.92, Capacity: 10, MaxArmor: 100, MaxSpeed: 497, MaxAngularSpeed: 3},
			2: {ID: 2, TitleRu: "Р В РЎвЂ™Р РЋР С“Р РЋРІР‚С™Р В Р’ВµР РЋР вЂљР В РЎвЂўР В РЎвЂР В РўвЂ", TitleEn: "Asteroid", Acronym: "asteroid_0002", CosmicObjectTypeID: 3, TextureScale: 4, TextureBodyWidth: 30, TextureBodyLength: 30},
			3: {ID: 3, TitleRu: "Р В Р Р‹Р РЋРІР‚С™Р В Р’В°Р В Р вЂ¦Р РЋРІР‚В Р В РЎвЂР РЋР РЏ", TitleEn: "Station", Acronym: "station_tiny_crumb", CosmicObjectTypeID: 2, TextureScale: 4, TextureBodyWidth: 30, TextureBodyLength: 30},
			4: {ID: 4, TitleRu: "Р В РЎСљР В РЎвЂўР В Р вЂ Р РЋРІР‚в„–Р В РІвЂћвЂ“ Р В РЎвЂќР В РЎвЂўР РЋР вЂљР В Р’В°Р В Р’В±Р В Р’В»Р РЋР Р‰", TitleEn: "New Ship", Acronym: "ship_new", CosmicObjectTypeID: 1, TextureScale: 4, TextureBodyWidth: 40, TextureBodyLength: 80, Mass: 12, Capacity: 20, MaxArmor: 150, MaxSpeed: 600, MaxAngularSpeed: 4},
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
	itemTypes := &data.ItemTypes{
		MaxID: 18,
		Items: map[int64]*data.ItemType{
			1:  {ID: 1, TitleRu: "Weapon", TitleEn: "Weapon", Acronym: "Weapon", CountMustBeInteger: true},
			7:  {ID: 7, TitleRu: "Fuel", TitleEn: "Fuel", Acronym: "Fuel"},
			9:  {ID: 9, TitleRu: "Container", TitleEn: "Container", Acronym: "Container", CountMustBeInteger: true},
			18: {ID: 18, TitleRu: "Ammunition", TitleEn: "Ammunition", Acronym: "Ammunition", CountMustBeInteger: true},
		},
	}
	itemModels := &data.ItemModels{
		MaxID: 403,
		Items: map[int64]*data.ItemModel{
			101: {ID: 101, TitleRu: "Thruster", TitleEn: "Thruster", Acronym: "Thruster", ItemTypeID: 1, ConsumingPower: 15000, ConsumingItemModelID: 7, ConsumingCount: 0.5, MaxAlongForce: 45000, MaxAcrossForce: 40000},
			102: {ID: 102, TitleRu: "Torquer", TitleEn: "Torquer", Acronym: "Torquer", ItemTypeID: 1, ConsumingPower: 5000, ConsumingItemModelID: 7, ConsumingCount: 0.25, MaxTorque: 70000},
			201: {ID: 201, TitleRu: "New Thruster", TitleEn: "New Thruster", Acronym: "NewThruster", ItemTypeID: 1, ConsumingPower: 10000, ConsumingItemModelID: 7, ConsumingCount: 0.4, MaxAlongForce: 30000, MaxAcrossForce: 27500},
			301: {ID: 301, TitleRu: "Container", TitleEn: "Container", Acronym: "Container", ItemTypeID: 9},
			302: {ID: 302, TitleRu: "Cannon", TitleEn: "Cannon", Acronym: "Cannon", ItemTypeID: 1, AmmoItemModelID: 303, FiringRate: 0.5},
			303: {ID: 303, TitleRu: "Shell", TitleEn: "Shell", Acronym: "Shell", ItemTypeID: 18},
			401: {ID: 401, TitleRu: "New Container", TitleEn: "New Container", Acronym: "NewContainer", ItemTypeID: 9},
			402: {ID: 402, TitleRu: "New Cannon", TitleEn: "New Cannon", Acronym: "NewCannon", ItemTypeID: 1, AmmoItemModelID: 403, FiringRate: 1},
			403: {ID: 403, TitleRu: "New Shell", TitleEn: "New Shell", Acronym: "NewShell", ItemTypeID: 18},
		},
	}
	equipmentGroups := data.NewEquipmentGroups()
	itemGroups := data.NewItemGroups()
	taskTypes := data.NewTaskTypes()
	tasks := data.NewTasks()
	taskItemGroups := data.NewTaskItemGroups()
	implementers := data.NewImplementers()
	itemProductionType, err := taskTypes.Add(&data.TaskType{TitleRu: "Р СџРЎР‚Р С•Р С‘Р В·Р Р†Р С•Р Т‘РЎРѓРЎвЂљР Р†Р С• Р С—РЎР‚Р ВµР Т‘Р СР ВµРЎвЂљР С•Р Р†", TitleEn: "Item production", Acronym: "ItemProduction"})
	if err != nil {
		t.Fatal(err)
	}
	objectProductionType, err := taskTypes.Add(&data.TaskType{TitleRu: "Р СџРЎР‚Р С•Р С‘Р В·Р Р†Р С•Р Т‘РЎРѓРЎвЂљР Р†Р С• Р С•Р В±РЎР‰Р ВµР С”РЎвЂљР С•Р Р†", TitleEn: "Object production", Acronym: "ObjectProduction"})
	if err != nil {
		t.Fatal(err)
	}
	if _, err := implementers.Add(&data.Implementer{TaskTypeID: itemProductionType.ID, ImplementerEquipmentItemTypeID: 19, WorkPart: 1}); err != nil {
		t.Fatal(err)
	}
	if _, err := implementers.Add(&data.Implementer{TaskTypeID: objectProductionType.ID, ImplementerEquipmentItemTypeID: 19, WorkPart: 1}); err != nil {
		t.Fatal(err)
	}
	cargoMovementType, err := taskTypes.Add(&data.TaskType{TitleRu: "Перемещение груза", TitleEn: "Cargo movement", Acronym: "CargoMovement"})
	if err != nil {
		t.Fatal(err)
	}
	if _, err := implementers.Add(&data.Implementer{TaskTypeID: cargoMovementType.ID, ImplementerEquipmentItemTypeID: 9, WorkPart: 0}); err != nil {
		t.Fatal(err)
	}
	if _, err := implementers.Add(&data.Implementer{TaskTypeID: cargoMovementType.ID, ImplementerEquipmentItemTypeID: 20, WorkPart: 1}); err != nil {
		t.Fatal(err)
	}
	fuelingType, err := taskTypes.Add(&data.TaskType{TitleRu: "Заправка топлива", TitleEn: "Fueling", Acronym: "Fueling"})
	if err != nil {
		t.Fatal(err)
	}
	if _, err := implementers.Add(&data.Implementer{TaskTypeID: fuelingType.ID, ImplementerEquipmentItemTypeID: 10, WorkPart: 0}); err != nil {
		t.Fatal(err)
	}
	if _, err := implementers.Add(&data.Implementer{TaskTypeID: fuelingType.ID, ImplementerEquipmentItemTypeID: 20, WorkPart: 1}); err != nil {
		t.Fatal(err)
	}
	itemDeconstructionType, err := taskTypes.Add(&data.TaskType{TitleRu: "Деконструкция предметов", TitleEn: "Item deconstruction", Acronym: "ItemDeconstruction"})
	if err != nil {
		t.Fatal(err)
	}
	if _, err := implementers.Add(&data.Implementer{TaskTypeID: itemDeconstructionType.ID, ImplementerEquipmentItemTypeID: 12, WorkPart: 0.5}); err != nil {
		t.Fatal(err)
	}
	if _, err := implementers.Add(&data.Implementer{TaskTypeID: itemDeconstructionType.ID, ImplementerEquipmentItemTypeID: 20, WorkPart: 0.5}); err != nil {
		t.Fatal(err)
	}

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
	if err := itemTypes.RebuildIndexes(); err != nil {
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
		ItemTypes:               itemTypes,
		ItemModels:              itemModels,
		TaskTypes:               taskTypes,
		Tasks:                   tasks,
		TaskItemGroups:          taskItemGroups,
		Implementers:            implementers,
		EquipmentGroups:         equipmentGroups,
		ItemGroups:              itemGroups,
		Assemblies:              assemblies,
		AssemblyEquipmentGroups: assemblyEquipmentGroups,
	}
}

// Р В РЎСџР РЋР вЂљР В РЎвЂўР В Р вЂ Р В Р’ВµР РЋР вЂљР РЋР РЏР В Р’ВµР РЋРІР‚С™, Р РЋРІР‚РЋР РЋРІР‚С™Р В РЎвЂў Р РЋР С“Р В РЎвЂўР РЋРІР‚В¦Р РЋР вЂљР В Р’В°Р В Р вЂ¦Р В Р’ВµР В Р вЂ¦Р В РЎвЂР В Р’Вµ Р В РўвЂР В Р’В°Р В Р вЂ¦Р В Р вЂ¦Р РЋРІР‚в„–Р РЋРІР‚В¦ Р В Р’В·Р В Р’В°Р В РЎвЂ”Р В РЎвЂР РЋР С“Р РЋРІР‚в„–Р В Р вЂ Р В Р’В°Р В Р’ВµР РЋРІР‚С™ Р В РЎвЂўР В Р’В±Р В Р вЂ¦Р В РЎвЂўР В Р вЂ Р В Р’В»Р РЋРІР‚ВР В Р вЂ¦Р В Р вЂ¦Р В РЎвЂўР В Р’Вµ Р В РЎвЂ”Р В РЎвЂўР В Р’В»Р В РЎвЂўР В Р’В¶Р В Р’ВµР В Р вЂ¦Р В РЎвЂР В Р’Вµ Р В РЎвЂќР В РЎвЂўР РЋР С“Р В РЎВР В РЎвЂР РЋРІР‚РЋР В Р’ВµР РЋР С“Р В РЎвЂќР В РЎвЂўР В РЎвЂ“Р В РЎвЂў Р В РЎвЂўР В Р’В±Р РЋР вЂ°Р В Р’ВµР В РЎвЂќР РЋРІР‚С™Р В Р’В°.
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

// Р В РЎСџР РЋР вЂљР В РЎвЂўР В Р вЂ Р В Р’ВµР РЋР вЂљР РЋР РЏР В Р’ВµР РЋРІР‚С™, Р РЋРІР‚РЋР РЋРІР‚С™Р В РЎвЂў Р РЋР С“Р В РЎвЂўР РЋРІР‚В¦Р РЋР вЂљР В Р’В°Р В Р вЂ¦Р В Р’ВµР В Р вЂ¦Р В РЎвЂР В Р’Вµ Р В РўвЂР В Р’В°Р В Р вЂ¦Р В Р вЂ¦Р РЋРІР‚в„–Р РЋРІР‚В¦ Р В Р’В·Р В Р’В°Р В РЎвЂ”Р В РЎвЂР РЋР С“Р РЋРІР‚в„–Р В Р вЂ Р В Р’В°Р В Р’ВµР РЋРІР‚С™ Р В РЎвЂўР В Р’В±Р В Р вЂ¦Р В РЎвЂўР В Р вЂ Р В Р’В»Р РЋРІР‚ВР В Р вЂ¦Р В Р вЂ¦Р В РЎвЂўР В Р’Вµ Р В РЎвЂ”Р В РЎвЂўР В Р’В»Р В РЎвЂўР В Р’В¶Р В Р’ВµР В Р вЂ¦Р В РЎвЂР В Р’Вµ Р В РЎвЂќР В РЎвЂўР РЋР С“Р В РЎВР В РЎвЂР РЋРІР‚РЋР В Р’ВµР РЋР С“Р В РЎвЂќР В РЎвЂўР В РЎвЂ“Р В РЎвЂў Р В РЎвЂўР В Р’В±Р РЋР вЂ°Р В Р’ВµР В РЎвЂќР РЋРІР‚С™Р В Р’В°.
// Р В РЎСџР РЋР вЂљР В РЎвЂўР В Р вЂ Р В Р’ВµР РЋР вЂљР РЋР РЏР В Р’ВµР РЋРІР‚С™, Р РЋРІР‚РЋР РЋРІР‚С™Р В РЎвЂў Р РЋР С“Р В РЎвЂўР РЋРІР‚В¦Р РЋР вЂљР В Р’В°Р В Р вЂ¦Р В Р’ВµР В Р вЂ¦Р В РЎвЂР В Р’Вµ Р В РўвЂР В Р’В°Р В Р вЂ¦Р В Р вЂ¦Р РЋРІР‚в„–Р РЋРІР‚В¦ Р В Р’В·Р В Р’В°Р В РЎвЂ”Р В РЎвЂР РЋР С“Р РЋРІР‚в„–Р В Р вЂ Р В Р’В°Р В Р’ВµР РЋРІР‚С™ Р В РЎвЂўР В Р’В±Р В Р вЂ¦Р В РЎвЂўР В Р вЂ Р В Р’В»Р РЋРІР‚ВР В Р вЂ¦Р В Р вЂ¦Р В РЎвЂўР В Р’Вµ Р В РЎвЂ”Р В РЎвЂўР В Р’В»Р В РЎвЂўР В Р’В¶Р В Р’ВµР В Р вЂ¦Р В РЎвЂР В Р’Вµ Р В РЎвЂќР В РЎвЂўР РЋР С“Р В РЎВР В РЎвЂР РЋРІР‚РЋР В Р’ВµР РЋР С“Р В РЎвЂќР В РЎвЂўР В РЎвЂ“Р В РЎвЂў Р В РЎвЂўР В Р’В±Р РЋР вЂ°Р В Р’ВµР В РЎвЂќР РЋРІР‚С™Р В Р’В°.
func TestSnapshotAddsTemporaryDrillRayObject(t *testing.T) {
	serverData := testWorldData(t)
	addDrillRayTestData(t, &serverData)
	serverData.CosmicObjectModels.Items[1].TextureHeight = 120
	serverData.CosmicObjectModels.Items[1].TextureBodyOriginY = 80
	serverData.CosmicObjectModels.Items[1].TextureBodyLength = 40
	serverData.CosmicObjectModels.Items[1].TextureScale = 1
	if err := serverData.CosmicObjectModels.RebuildIndexes(); err != nil {
		t.Fatalf("rebuild cosmic object model indexes: %v", err)
	}
	serverData.CosmicObjects.Items[1].X = 100
	serverData.CosmicObjects.Items[1].Y = -50
	serverData.CosmicObjects.Items[1].Rotation = 0.4
	gameWorld := world.New(1, serverData)
	if _, ok := gameWorld.ConnectAccount(1); !ok {
		t.Fatal("account was not connected")
	}

	withoutAction := gameWorld.SnapshotForAccount(1)
	if _, ok := findCosmicObjectInSnapshot(withoutAction, -1); ok {
		t.Fatalf("temporary ray exists without primary action")
	}

	gameWorld.SetInput(1, game.ShipInput{PrimaryPointerAction: true})
	snapshot := gameWorld.SnapshotForAccount(1)

	ray, ok := findCosmicObjectInSnapshot(snapshot, -1)
	if !ok {
		t.Fatalf("temporary ray was not added to snapshot")
	}
	if _, ok := serverData.CosmicObjects.Get(-1); ok {
		t.Fatalf("temporary ray was saved in persistent objects")
	}
	if ray.CosmicObjectModelID != 900 || ray.Title != "DrillRay" || !ray.Enabled {
		t.Fatalf("temporary ray uses wrong model or disabled state: %+v", ray.CosmicObject)
	}
	ship := serverData.CosmicObjects.Items[1]
	shipModel := serverData.CosmicObjectModels.Items[1]
	rayModel := serverData.CosmicObjectModels.Items[900]
	forward := physics.ForwardVector(ship.Rotation)
	physicalNoseDistance := shipModel.BodyPolygon[0].Y
	for _, point := range shipModel.BodyPolygon {
		if point.Y > physicalNoseDistance {
			physicalNoseDistance = point.Y
		}
	}
	expectedDistance := physicalNoseDistance + rayModel.BodyLength/2
	expectedX := ship.X + forward.X*expectedDistance
	expectedY := ship.Y + forward.Y*expectedDistance
	if math.Abs(ray.X-expectedX) > physics.Epsilon || math.Abs(ray.Y-expectedY) > physics.Epsilon {
		t.Fatalf("ray position got (%.6f, %.6f), want (%.6f, %.6f)", ray.X, ray.Y, expectedX, expectedY)
	}
	startDistance := math.Hypot(ray.X-ship.X, ray.Y-ship.Y) - rayModel.BodyLength/2
	if math.Abs(startDistance-physicalNoseDistance) > physics.Epsilon {
		t.Fatalf("ray start distance got %.6f, want %.6f", startDistance, physicalNoseDistance)
	}
	if math.Abs(ray.Rotation-ship.Rotation) > physics.Epsilon {
		t.Fatalf("ray rotation got %.6f, want %.6f", ray.Rotation, ship.Rotation)
	}
}

// Р В РЎСџР РЋР вЂљР В РЎвЂўР В Р вЂ Р В Р’ВµР РЋР вЂљР РЋР РЏР В Р’ВµР РЋРІР‚С™, Р РЋРІР‚РЋР РЋРІР‚С™Р В РЎвЂў Р РЋР С“Р В РЎвЂўР РЋРІР‚В¦Р РЋР вЂљР В Р’В°Р В Р вЂ¦Р В Р’ВµР В Р вЂ¦Р В РЎвЂР В Р’Вµ Р В РўвЂР В Р’В°Р В Р вЂ¦Р В Р вЂ¦Р РЋРІР‚в„–Р РЋРІР‚В¦ Р В Р’В·Р В Р’В°Р В РЎвЂ”Р В РЎвЂР РЋР С“Р РЋРІР‚в„–Р В Р вЂ Р В Р’В°Р В Р’ВµР РЋРІР‚С™ Р В РЎвЂўР В Р’В±Р В Р вЂ¦Р В РЎвЂўР В Р вЂ Р В Р’В»Р РЋРІР‚ВР В Р вЂ¦Р В Р вЂ¦Р В РЎвЂўР В Р’Вµ Р В РЎвЂ”Р В РЎвЂўР В Р’В»Р В РЎвЂўР В Р’В¶Р В Р’ВµР В Р вЂ¦Р В РЎвЂР В Р’Вµ Р В РЎвЂќР В РЎвЂўР РЋР С“Р В РЎВР В РЎвЂР РЋРІР‚РЋР В Р’ВµР РЋР С“Р В РЎвЂќР В РЎвЂўР В РЎвЂ“Р В РЎвЂў Р В РЎвЂўР В Р’В±Р РЋР вЂ°Р В Р’ВµР В РЎвЂќР РЋРІР‚С™Р В Р’В°.
func TestSnapshotAddsTemporaryDrillRayOnlyForSelectedDrill(t *testing.T) {
	serverData := testWorldData(t)
	addDrillRayTestData(t, &serverData)
	serverData.EquipmentGroups.Items[600] = &data.EquipmentGroup{
		ID:                   600,
		CosmicObjectID:       1,
		Title:                "Cannon",
		EquipmentItemModelID: 302,
		Count:                1,
		EnabledCount:         1,
		Enabled:              true,
	}
	if err := serverData.EquipmentGroups.RebuildIndexes(); err != nil {
		t.Fatal(err)
	}
	gameWorld := world.New(1, serverData)
	if _, ok := gameWorld.ConnectAccount(1); !ok {
		t.Fatal("account was not connected")
	}

	gameWorld.SetInput(1, game.ShipInput{SelectedPilotToolIndex: 0, PrimaryPointerAction: true})
	cannonSnapshot := gameWorld.SnapshotForAccount(1)
	if _, ok := findCosmicObjectInSnapshot(cannonSnapshot, -1); ok {
		t.Fatalf("temporary ray was added for selected cannon")
	}

	gameWorld.SetInput(1, game.ShipInput{SelectedPilotToolIndex: 1, PrimaryPointerAction: true})
	drillSnapshot := gameWorld.SnapshotForAccount(1)
	if _, ok := findCosmicObjectInSnapshot(drillSnapshot, -1); !ok {
		t.Fatalf("temporary ray was not added for selected drill")
	}
}

// Р В РЎСџР РЋР вЂљР В РЎвЂўР В Р вЂ Р В Р’ВµР РЋР вЂљР РЋР РЏР В Р’ВµР РЋРІР‚С™, Р РЋРІР‚РЋР РЋРІР‚С™Р В РЎвЂў Р РЋР С“Р В РЎвЂўР РЋРІР‚В¦Р РЋР вЂљР В Р’В°Р В Р вЂ¦Р В Р’ВµР В Р вЂ¦Р В РЎвЂР В Р’Вµ Р В РўвЂР В Р’В°Р В Р вЂ¦Р В Р вЂ¦Р РЋРІР‚в„–Р РЋРІР‚В¦ Р В Р’В·Р В Р’В°Р В РЎвЂ”Р В РЎвЂР РЋР С“Р РЋРІР‚в„–Р В Р вЂ Р В Р’В°Р В Р’ВµР РЋРІР‚С™ Р В РЎвЂўР В Р’В±Р В Р вЂ¦Р В РЎвЂўР В Р вЂ Р В Р’В»Р РЋРІР‚ВР В Р вЂ¦Р В Р вЂ¦Р В РЎвЂўР В Р’Вµ Р В РЎвЂ”Р В РЎвЂўР В Р’В»Р В РЎвЂўР В Р’В¶Р В Р’ВµР В Р вЂ¦Р В РЎвЂР В Р’Вµ Р В РЎвЂќР В РЎвЂўР РЋР С“Р В РЎВР В РЎвЂР РЋРІР‚РЋР В Р’ВµР РЋР С“Р В РЎвЂќР В РЎвЂўР В РЎвЂ“Р В РЎвЂў Р В РЎвЂўР В Р’В±Р РЋР вЂ°Р В Р’ВµР В РЎвЂќР РЋРІР‚С™Р В Р’В°.
func TestSelectedSimpleDrillConsumesPowerOnPrimaryAction(t *testing.T) {
	serverData := testWorldData(t)
	addDrillRayTestData(t, &serverData)
	serverData.ItemModels.Items[700].ConsumingPower = 12000
	if err := serverData.ItemModels.RebuildIndexes(); err != nil {
		t.Fatal(err)
	}
	addTestPowerProducer(t, serverData, 1)
	gameWorld := world.New(1, serverData)
	if _, ok := gameWorld.ConnectAccount(1); !ok {
		t.Fatal("account was not connected")
	}

	gameWorld.SetInput(1, game.ShipInput{SelectedPilotToolIndex: 0, PrimaryPointerAction: true})
	snapshot := gameWorld.Tick(1)

	ship, ok := findCosmicObjectInSnapshot(snapshot, 1)
	if !ok {
		t.Fatal("ship was not found")
	}
	closeWorldFloat(t, ship.ConsumingPower, 12000)
	drill, ok := findEquipmentGroupInSnapshot(snapshot, 700)
	if !ok {
		t.Fatal("drill group was not found")
	}
	if !drill.Active {
		t.Fatalf("drill group was not marked active")
	}
}

// Р В РЎСџР РЋР вЂљР В РЎвЂўР В Р вЂ Р В Р’ВµР РЋР вЂљР РЋР РЏР В Р’ВµР РЋРІР‚С™, Р РЋРІР‚РЋР РЋРІР‚С™Р В РЎвЂў Р РЋР С“Р В РЎвЂўР РЋРІР‚В¦Р РЋР вЂљР В Р’В°Р В Р вЂ¦Р В Р’ВµР В Р вЂ¦Р В РЎвЂР В Р’Вµ Р В РўвЂР В Р’В°Р В Р вЂ¦Р В Р вЂ¦Р РЋРІР‚в„–Р РЋРІР‚В¦ Р В Р’В·Р В Р’В°Р В РЎвЂ”Р В РЎвЂР РЋР С“Р РЋРІР‚в„–Р В Р вЂ Р В Р’В°Р В Р’ВµР РЋРІР‚С™ Р В РЎвЂўР В Р’В±Р В Р вЂ¦Р В РЎвЂўР В Р вЂ Р В Р’В»Р РЋРІР‚ВР В Р вЂ¦Р В Р вЂ¦Р В РЎвЂўР В Р’Вµ Р В РЎвЂ”Р В РЎвЂўР В Р’В»Р В РЎвЂўР В Р’В¶Р В Р’ВµР В Р вЂ¦Р В РЎвЂР В Р’Вµ Р В РЎвЂќР В РЎвЂўР РЋР С“Р В РЎВР В РЎвЂР РЋРІР‚РЋР В Р’ВµР РЋР С“Р В РЎвЂќР В РЎвЂўР В РЎвЂ“Р В РЎвЂў Р В РЎвЂўР В Р’В±Р РЋР вЂ°Р В Р’ВµР В РЎвЂќР РЋРІР‚С™Р В Р’В°.
// Проверяет расход энергии оружием, выбранным для основного действия пилота.
func TestSelectedWeaponConsumesPowerOnPrimaryAction(t *testing.T) {
	serverData := testWorldData(t)
	addWeaponTestData(t, &serverData)
	serverData.ItemModels.Items[302].ConsumingPower = 14000
	if err := serverData.ItemModels.RebuildIndexes(); err != nil {
		t.Fatal(err)
	}
	addTestPowerProducer(t, serverData, 1)
	gameWorld := world.New(1, serverData)
	if _, ok := gameWorld.ConnectAccount(1); !ok {
		t.Fatal("account was not connected")
	}

	gameWorld.SetInput(1, game.ShipInput{SelectedPilotToolIndex: 0, PrimaryPointerAction: true})
	snapshot := gameWorld.Tick(0.1)

	ship, ok := findCosmicObjectInSnapshot(snapshot, 1)
	if !ok {
		t.Fatal("ship was not found")
	}
	closeWorldFloat(t, ship.ConsumingPower, 14000)
	weapon, ok := findEquipmentGroupInSnapshot(snapshot, 600)
	if !ok {
		t.Fatal("weapon group was not found")
	}
	if !weapon.Active {
		t.Fatalf("weapon group was not marked active")
	}
}

func TestSelectedSimpleDrillMinesAsteroidResourceToShipContainer(t *testing.T) {
	serverData := testWorldData(t)
	addDrillRayTestData(t, &serverData)
	serverData.ItemTypes.Items[17] = &data.ItemType{ID: 17, TitleRu: "Ресурс", TitleEn: "Resource", Acronym: "Resource"}
	serverData.ItemModels.Items[700].MiningSpeed = 20
	serverData.ItemModels.Items[710] = &data.ItemModel{ID: 710, TitleRu: "Руда", TitleEn: "Ore", Acronym: "Ore", ItemTypeID: 17, Mass: 2}
	if err := serverData.ItemTypes.RebuildIndexes(); err != nil {
		t.Fatal(err)
	}
	if err := serverData.ItemModels.RebuildIndexes(); err != nil {
		t.Fatal(err)
	}
	shipContainer, err := serverData.EquipmentGroups.Add(&data.EquipmentGroup{
		CosmicObjectID:       1,
		Title:                "Ship container",
		EquipmentItemModelID: 301,
		Count:                1,
		EnabledCount:         1,
		Enabled:              true,
	})
	if err != nil {
		t.Fatal(err)
	}
	asteroidContainer, err := serverData.EquipmentGroups.Add(&data.EquipmentGroup{
		CosmicObjectID:       2,
		Title:                "Asteroid container",
		EquipmentItemModelID: 301,
		Count:                1,
		EnabledCount:         1,
		Enabled:              true,
	})
	if err != nil {
		t.Fatal(err)
	}
	if _, err := serverData.ItemGroups.Add(&data.ItemGroup{ContainerEquipmentGroupID: asteroidContainer.ID, ContentItemModelID: 710, Count: 100}); err != nil {
		t.Fatal(err)
	}
	serverData.CosmicObjects.Items[1].X = 0
	serverData.CosmicObjects.Items[1].Y = 0
	serverData.CosmicObjects.Items[1].Rotation = 0
	serverData.CosmicObjects.Items[1].Anchored = true
	serverData.CosmicObjects.Items[2].X = 0
	serverData.CosmicObjects.Items[2].Y = 100
	serverData.CosmicObjects.Items[2].Anchored = true
	gameWorld := world.New(1, serverData)
	if _, ok := gameWorld.ConnectAccount(1); !ok {
		t.Fatal("account was not connected")
	}

	gameWorld.SetInput(1, game.ShipInput{SelectedPilotToolIndex: 0, PrimaryPointerAction: true})
	gameWorld.Tick(1)

	closeWorldFloat(t, exchangeContainerItemCount(serverData, asteroidContainer.ID, 710), 90)
	closeWorldFloat(t, exchangeContainerItemCount(serverData, shipContainer.ID, 710), 10)
}

// Р В РЎСџР РЋР вЂљР В РЎвЂўР В Р вЂ Р В Р’ВµР РЋР вЂљР РЋР РЏР В Р’ВµР РЋРІР‚С™, Р РЋРІР‚РЋР РЋРІР‚С™Р В РЎвЂў Р РЋР С“Р В РЎвЂўР РЋРІР‚В¦Р РЋР вЂљР В Р’В°Р В Р вЂ¦Р В Р’ВµР В Р вЂ¦Р В РЎвЂР В Р’Вµ Р В РўвЂР В Р’В°Р В Р вЂ¦Р В Р вЂ¦Р РЋРІР‚в„–Р РЋРІР‚В¦ Р В Р’В·Р В Р’В°Р В РЎвЂ”Р В РЎвЂР РЋР С“Р РЋРІР‚в„–Р В Р вЂ Р В Р’В°Р В Р’ВµР РЋРІР‚С™ Р В РЎвЂўР В Р’В±Р В Р вЂ¦Р В РЎвЂўР В Р вЂ Р В Р’В»Р РЋРІР‚ВР В Р вЂ¦Р В Р вЂ¦Р В РЎвЂўР В Р’Вµ Р В РЎвЂ”Р В РЎвЂўР В Р’В»Р В РЎвЂўР В Р’В¶Р В Р’ВµР В Р вЂ¦Р В РЎвЂР В Р’Вµ Р В РЎвЂќР В РЎвЂўР РЋР С“Р В РЎВР В РЎвЂР РЋРІР‚РЋР В Р’ВµР РЋР С“Р В РЎвЂќР В РЎвЂўР В РЎвЂ“Р В РЎвЂў Р В РЎвЂўР В Р’В±Р РЋР вЂ°Р В Р’ВµР В РЎвЂќР РЋРІР‚С™Р В Р’В°.
func TestSelectedSimpleDrillReportsMinedResourceEverySecond(t *testing.T) {
	serverData := testWorldData(t)
	addDrillRayTestData(t, &serverData)
	serverData.ItemTypes.Items[17] = &data.ItemType{ID: 17, TitleRu: "Ресурс", TitleEn: "Resource", Acronym: "Resource"}
	serverData.ItemModels.Items[700].MiningSpeed = 20
	serverData.ItemModels.Items[710] = &data.ItemModel{ID: 710, TitleRu: "Руда", TitleEn: "Ore", Acronym: "Ore", ItemTypeID: 17, Mass: 2}
	if err := serverData.ItemTypes.RebuildIndexes(); err != nil {
		t.Fatal(err)
	}
	if err := serverData.ItemModels.RebuildIndexes(); err != nil {
		t.Fatal(err)
	}
	if _, err := serverData.EquipmentGroups.Add(&data.EquipmentGroup{
		CosmicObjectID:       1,
		Title:                "Ship container",
		EquipmentItemModelID: 301,
		Count:                1,
		EnabledCount:         1,
		Enabled:              true,
	}); err != nil {
		t.Fatal(err)
	}
	asteroidContainer, err := serverData.EquipmentGroups.Add(&data.EquipmentGroup{
		CosmicObjectID:       2,
		Title:                "Asteroid container",
		EquipmentItemModelID: 301,
		Count:                1,
		EnabledCount:         1,
		Enabled:              true,
	})
	if err != nil {
		t.Fatal(err)
	}
	if _, err := serverData.ItemGroups.Add(&data.ItemGroup{ContainerEquipmentGroupID: asteroidContainer.ID, ContentItemModelID: 710, Count: 300}); err != nil {
		t.Fatal(err)
	}
	serverData.CosmicObjects.Items[1].X = 0
	serverData.CosmicObjects.Items[1].Y = 0
	serverData.CosmicObjects.Items[1].Rotation = 0
	serverData.CosmicObjects.Items[1].Anchored = true
	serverData.CosmicObjects.Items[2].X = 0
	serverData.CosmicObjects.Items[2].Y = 100
	serverData.CosmicObjects.Items[2].Anchored = true
	gameWorld := world.New(1, serverData)
	if _, ok := gameWorld.ConnectAccount(1); !ok {
		t.Fatal("account was not connected")
	}

	gameWorld.SetInput(1, game.ShipInput{SelectedPilotToolIndex: 0, PrimaryPointerAction: true})
	gameWorld.Tick(0.5)
	if exchangeEventsContainKind(gameWorld.DrainExchangeEvents(), "exchangeNotification") {
		t.Fatalf("mining notification was sent before one second")
	}

	gameWorld.Tick(0.5)

	events := gameWorld.DrainExchangeEvents()
	if !exchangeEventsContainNotificationMessage(events, "+ 10 Руда") {
		t.Fatalf("mining notification was not sent with expected text: %+v", events)
	}
}

// Р В РЎСџР РЋР вЂљР В РЎвЂўР В Р вЂ Р В Р’ВµР РЋР вЂљР РЋР РЏР В Р’ВµР РЋРІР‚С™, Р РЋРІР‚РЋР РЋРІР‚С™Р В РЎвЂў Р РЋР С“Р В РЎвЂўР РЋРІР‚В¦Р РЋР вЂљР В Р’В°Р В Р вЂ¦Р В Р’ВµР В Р вЂ¦Р В РЎвЂР В Р’Вµ Р В РўвЂР В Р’В°Р В Р вЂ¦Р В Р вЂ¦Р РЋРІР‚в„–Р РЋРІР‚В¦ Р В Р’В·Р В Р’В°Р В РЎвЂ”Р В РЎвЂР РЋР С“Р РЋРІР‚в„–Р В Р вЂ Р В Р’В°Р В Р’ВµР РЋРІР‚С™ Р В РЎвЂўР В Р’В±Р В Р вЂ¦Р В РЎвЂўР В Р вЂ Р В Р’В»Р РЋРІР‚ВР В Р вЂ¦Р В Р вЂ¦Р В РЎвЂўР В Р’Вµ Р В РЎвЂ”Р В РЎвЂўР В Р’В»Р В РЎвЂўР В Р’В¶Р В Р’ВµР В Р вЂ¦Р В РЎвЂР В Р’Вµ Р В РЎвЂќР В РЎвЂўР РЋР С“Р В РЎВР В РЎвЂР РЋРІР‚РЋР В Р’ВµР РЋР С“Р В РЎвЂќР В РЎвЂўР В РЎвЂ“Р В РЎвЂў Р В РЎвЂўР В Р’В±Р РЋР вЂ°Р В Р’ВµР В РЎвЂќР РЋРІР‚С™Р В Р’В°.
func TestSelectedWeaponCreatesProjectileObjectOnPrimaryAction(t *testing.T) {
	serverData := testWorldData(t)
	addWeaponTestData(t, &serverData)
	serverData.CosmicObjects.Items[1].X = 0
	serverData.CosmicObjects.Items[1].Y = 0
	serverData.CosmicObjects.Items[1].Rotation = 0
	serverData.CosmicObjects.Items[1].Anchored = true
	serverData.CosmicObjects.Items[4] = &data.CosmicObject{ID: 4, Title: "Target ship", CosmicObjectModelID: 1, X: 0, Y: 100, MaxArmor: 300, Armor: 300, Enabled: true, Anchored: true}
	serverData.CosmicObjects.MaxID = 4
	if err := serverData.CosmicObjects.RebuildIndexes(); err != nil {
		t.Fatal(err)
	}
	gameWorld := world.New(1, serverData)
	if _, ok := gameWorld.ConnectAccount(1); !ok {
		t.Fatal("account was not connected")
	}

	gameWorld.SetInput(1, game.ShipInput{SelectedPilotToolIndex: 0, PrimaryPointerAction: true})
	snapshot := gameWorld.Tick(0.1)

	projectile, ok := findCosmicObjectModelInSnapshot(snapshot, 901)
	if !ok {
		t.Fatalf("projectile was not added to snapshot: %+v", snapshot.Objects)
	}
	if projectile.ID >= 0 || projectile.Title != "BallisticProjectile" || !projectile.Enabled {
		t.Fatalf("projectile uses wrong temporary state: %+v", projectile.CosmicObject)
	}
	if _, ok := serverData.CosmicObjects.Get(projectile.ID); ok {
		t.Fatalf("projectile was saved in persistent objects")
	}
	closeWorldFloat(t, serverData.CosmicObjects.Items[4].Armor, 300)
}

// Р В РЎСџР РЋР вЂљР В РЎвЂўР В Р вЂ Р В Р’ВµР РЋР вЂљР РЋР РЏР В Р’ВµР РЋРІР‚С™, Р РЋРІР‚РЋР РЋРІР‚С™Р В РЎвЂў Р РЋР С“Р В РЎвЂўР РЋРІР‚В¦Р РЋР вЂљР В Р’В°Р В Р вЂ¦Р В Р’ВµР В Р вЂ¦Р В РЎвЂР В Р’Вµ Р В РўвЂР В Р’В°Р В Р вЂ¦Р В Р вЂ¦Р РЋРІР‚в„–Р РЋРІР‚В¦ Р В Р’В·Р В Р’В°Р В РЎвЂ”Р В РЎвЂР РЋР С“Р РЋРІР‚в„–Р В Р вЂ Р В Р’В°Р В Р’ВµР РЋРІР‚С™ Р В РЎвЂўР В Р’В±Р В Р вЂ¦Р В РЎвЂўР В Р вЂ Р В Р’В»Р РЋРІР‚ВР В Р вЂ¦Р В Р вЂ¦Р В РЎвЂўР В Р’Вµ Р В РЎвЂ”Р В РЎвЂўР В Р’В»Р В РЎвЂўР В Р’В¶Р В Р’ВµР В Р вЂ¦Р В РЎвЂР В Р’Вµ Р В РЎвЂќР В РЎвЂўР РЋР С“Р В РЎВР В РЎвЂР РЋРІР‚РЋР В Р’ВµР РЋР С“Р В РЎвЂќР В РЎвЂўР В РЎвЂ“Р В РЎвЂў Р В РЎвЂўР В Р’В±Р РЋР вЂ°Р В Р’ВµР В РЎвЂќР РЋРІР‚С™Р В Р’В°.
// Проверяет, что снаряд появляется у переднего края корпуса без дополнительного выноса вперед.
func TestSelectedWeaponProjectileStartsAtShipNose(t *testing.T) {
	serverData := testWorldData(t)
	addWeaponTestData(t, &serverData)
	serverData.CosmicObjects.Items[1].X = 0
	serverData.CosmicObjects.Items[1].Y = 0
	serverData.CosmicObjects.Items[1].Rotation = 0
	serverData.CosmicObjects.Items[1].Anchored = true
	gameWorld := world.New(1, serverData)
	if _, ok := gameWorld.ConnectAccount(1); !ok {
		t.Fatal("account was not connected")
	}

	gameWorld.SetInput(1, game.ShipInput{SelectedPilotToolIndex: 0, PrimaryPointerAction: true})
	snapshot := gameWorld.Tick(0.1)

	projectile, ok := findCosmicObjectModelInSnapshot(snapshot, 901)
	if !ok {
		t.Fatalf("projectile was not added to snapshot: %+v", snapshot.Objects)
	}
	shipModel := serverData.CosmicObjectModels.Items[1]
	expectedY := shipModel.BodyLength/2 + serverData.ItemModels.Items[302].ProjectileSpeed*0.1
	closeWorldFloat(t, projectile.X, 0)
	closeWorldFloat(t, projectile.Y, expectedY)
}

// Проверяет, что скорость корабля не меняет время жизни выпущенного снаряда.
func TestSelectedWeaponProjectileLifetimeUsesWeaponRangeAndProjectileSpeed(t *testing.T) {
	serverData := testWorldData(t)
	addWeaponTestData(t, &serverData)
	serverData.ItemModels.Items[302].Range = 100
	serverData.ItemModels.Items[302].ProjectileSpeed = 100
	serverData.CosmicObjects.Items[1].X = 0
	serverData.CosmicObjects.Items[1].Y = 0
	serverData.CosmicObjects.Items[1].Rotation = 0
	serverData.CosmicObjects.Items[1].VelocityY = 100
	serverData.CosmicObjects.Items[1].Speed = 100
	serverData.CosmicObjects.Items[1].Anchored = true
	gameWorld := world.New(1, serverData)
	if _, ok := gameWorld.ConnectAccount(1); !ok {
		t.Fatal("account was not connected")
	}

	gameWorld.SetInput(1, game.ShipInput{SelectedPilotToolIndex: 0, PrimaryPointerAction: true})
	gameWorld.Tick(0.1)
	snapshot := gameWorld.Tick(0.4)

	if _, ok := findCosmicObjectModelInSnapshot(snapshot, 901); !ok {
		t.Fatalf("projectile disappeared before range time: %+v", snapshot.Objects)
	}
}

// Проверяет, что боковой отступ до крайних снарядов вдвое больше расстояния между соседними снарядами.
func TestSelectedWeaponMultipleCannonsUseDoubleEdgeMarginAcrossShipWidth(t *testing.T) {
	serverData := testWorldData(t)
	addWeaponTestData(t, &serverData)
	serverData.CosmicObjects.Items[1].X = 0
	serverData.CosmicObjects.Items[1].Y = 0
	serverData.CosmicObjects.Items[1].Rotation = 0
	serverData.CosmicObjects.Items[1].Anchored = true
	serverData.EquipmentGroups.Items[600].Count = 3
	serverData.EquipmentGroups.Items[600].EnabledCount = 3
	gameWorld := world.New(1, serverData)
	if _, ok := gameWorld.ConnectAccount(1); !ok {
		t.Fatal("account was not connected")
	}

	gameWorld.SetInput(1, game.ShipInput{SelectedPilotToolIndex: 0, PrimaryPointerAction: true})
	snapshot := gameWorld.Tick(0.1)

	projectiles := findCosmicObjectModelsInSnapshot(snapshot, 901)
	if len(projectiles) != 3 {
		t.Fatalf("projectile count = %d, want 3: %+v", len(projectiles), snapshot.Objects)
	}
	sort.Slice(projectiles, func(left int, right int) bool {
		return projectiles[left].X < projectiles[right].X
	})
	shipModel := serverData.CosmicObjectModels.Items[1]
	expectedY := shipModel.BodyLength/2 + serverData.ItemModels.Items[302].ProjectileSpeed*0.1
	expectedGap := shipModel.BodyWidth / 6
	closeWorldFloat(t, projectiles[0].X, -expectedGap)
	closeWorldFloat(t, projectiles[1].X, 0)
	closeWorldFloat(t, projectiles[2].X, expectedGap)
	closeWorldFloat(t, projectiles[0].X+shipModel.BodyWidth/2, (projectiles[1].X-projectiles[0].X)*2)
	closeWorldFloat(t, shipModel.BodyWidth/2-projectiles[2].X, (projectiles[2].X-projectiles[1].X)*2)
	for _, projectile := range projectiles {
		closeWorldFloat(t, projectile.Y, expectedY)
	}
}

func TestSelectedWeaponProjectileAddsSourceVelocityOnLaunch(t *testing.T) {
	serverData := testWorldData(t)
	addWeaponTestData(t, &serverData)
	serverData.CosmicObjects.Items[1].X = 0
	serverData.CosmicObjects.Items[1].Y = 0
	serverData.CosmicObjects.Items[1].Rotation = 0
	serverData.CosmicObjects.Items[1].VelocityX = 20
	serverData.CosmicObjects.Items[1].VelocityY = 50
	serverData.CosmicObjects.Items[1].Speed = math.Hypot(20, 50)
	serverData.CosmicObjects.Items[1].Anchored = true
	gameWorld := world.New(1, serverData)
	if _, ok := gameWorld.ConnectAccount(1); !ok {
		t.Fatal("account was not connected")
	}

	gameWorld.SetInput(1, game.ShipInput{SelectedPilotToolIndex: 0, PrimaryPointerAction: true})
	snapshot := gameWorld.Tick(0.1)

	projectile, ok := findCosmicObjectModelInSnapshot(snapshot, 901)
	if !ok {
		t.Fatalf("projectile was not added to snapshot: %+v", snapshot.Objects)
	}
	closeWorldFloat(t, projectile.VelocityX, 20)
	closeWorldFloat(t, projectile.VelocityY, 150)
	closeWorldFloat(t, projectile.Speed, math.Hypot(20, 150))
}

// Р В РЎСџР РЋР вЂљР В РЎвЂўР В Р вЂ Р В Р’ВµР РЋР вЂљР РЋР РЏР В Р’ВµР РЋРІР‚С™, Р РЋРІР‚РЋР РЋРІР‚С™Р В РЎвЂў Р РЋР С“Р В РЎвЂўР РЋРІР‚В¦Р РЋР вЂљР В Р’В°Р В Р вЂ¦Р В Р’ВµР В Р вЂ¦Р В РЎвЂР В Р’Вµ Р В РўвЂР В Р’В°Р В Р вЂ¦Р В Р вЂ¦Р РЋРІР‚в„–Р РЋРІР‚В¦ Р В Р’В·Р В Р’В°Р В РЎвЂ”Р В РЎвЂР РЋР С“Р РЋРІР‚в„–Р В Р вЂ Р В Р’В°Р В Р’ВµР РЋРІР‚С™ Р В РЎвЂўР В Р’В±Р В Р вЂ¦Р В РЎвЂўР В Р вЂ Р В Р’В»Р РЋРІР‚ВР В Р вЂ¦Р В Р вЂ¦Р В РЎвЂўР В Р’Вµ Р В РЎвЂ”Р В РЎвЂўР В Р’В»Р В РЎвЂўР В Р’В¶Р В Р’ВµР В Р вЂ¦Р В РЎвЂР В Р’Вµ Р В РЎвЂќР В РЎвЂўР РЋР С“Р В РЎВР В РЎвЂР РЋРІР‚РЋР В Р’ВµР РЋР С“Р В РЎвЂќР В РЎвЂўР В РЎвЂ“Р В РЎвЂў Р В РЎвЂўР В Р’В±Р РЋР вЂ°Р В Р’ВµР В РЎвЂќР РЋРІР‚С™Р В Р’В°.
// Р В РЎСџР РЋР вЂљР В РЎвЂўР В Р вЂ Р В Р’ВµР РЋР вЂљР РЋР РЏР В Р’ВµР РЋРІР‚С™, Р РЋРІР‚РЋР РЋРІР‚С™Р В РЎвЂў Р РЋР С“Р В РЎвЂўР РЋРІР‚В¦Р РЋР вЂљР В Р’В°Р В Р вЂ¦Р В Р’ВµР В Р вЂ¦Р В РЎвЂР В Р’Вµ Р В РўвЂР В Р’В°Р В Р вЂ¦Р В Р вЂ¦Р РЋРІР‚в„–Р РЋРІР‚В¦ Р В Р’В·Р В Р’В°Р В РЎвЂ”Р В РЎвЂР РЋР С“Р РЋРІР‚в„–Р В Р вЂ Р В Р’В°Р В Р’ВµР РЋРІР‚С™ Р В РЎвЂўР В Р’В±Р В Р вЂ¦Р В РЎвЂўР В Р вЂ Р В Р’В»Р РЋРІР‚ВР В Р вЂ¦Р В Р вЂ¦Р В РЎвЂўР В Р’Вµ Р В РЎвЂ”Р В РЎвЂўР В Р’В»Р В РЎвЂўР В Р’В¶Р В Р’ВµР В Р вЂ¦Р В РЎвЂР В Р’Вµ Р В РЎвЂќР В РЎвЂўР РЋР С“Р В РЎВР В РЎвЂР РЋРІР‚РЋР В Р’ВµР РЋР С“Р В РЎвЂќР В РЎвЂўР В РЎвЂ“Р В РЎвЂў Р В РЎвЂўР В Р’В±Р РЋР вЂ°Р В Р’ВµР В РЎвЂќР РЋРІР‚С™Р В Р’В°.
// Проверяет, что оружие стреляет сразу после паузы дольше интервала между выстрелами.
func TestSelectedWeaponFiresImmediatelyAfterCooldownPause(t *testing.T) {
	serverData := testWorldData(t)
	addWeaponTestData(t, &serverData)
	serverData.ItemModels.Items[302].FiringRate = 1
	serverData.ItemModels.Items[302].Range = 500
	serverData.CosmicObjects.Items[1].X = 0
	serverData.CosmicObjects.Items[1].Y = 0
	serverData.CosmicObjects.Items[1].Rotation = 0
	serverData.CosmicObjects.Items[1].Anchored = true
	gameWorld := world.New(1, serverData)
	if _, ok := gameWorld.ConnectAccount(1); !ok {
		t.Fatal("account was not connected")
	}

	gameWorld.SetInput(1, game.ShipInput{SelectedPilotToolIndex: 0, PrimaryPointerAction: true})
	gameWorld.Tick(0.1)
	gameWorld.SetInput(1, game.ShipInput{SelectedPilotToolIndex: 0, PrimaryPointerAction: false})
	gameWorld.Tick(1.1)
	gameWorld.SetInput(1, game.ShipInput{SelectedPilotToolIndex: 0, PrimaryPointerAction: true})
	snapshot := gameWorld.Tick(0.01)

	projectiles := findCosmicObjectModelsInSnapshot(snapshot, 901)
	if len(projectiles) != 2 {
		t.Fatalf("projectile count = %d, want 2: %+v", len(projectiles), snapshot.Objects)
	}
}

func TestSelectedWeaponConsumesMagazineBeforeSpawningProjectile(t *testing.T) {
	serverData := testWorldData(t)
	addWeaponTestData(t, &serverData)
	serverData.ItemModels.Items[302].MagazineCapacity = 2
	serverData.ItemModels.Items[302].FiringRate = 0.1
	serverData.EquipmentGroups.Items[600].MagazineCount = 2
	gameWorld := world.New(1, serverData)
	if _, ok := gameWorld.ConnectAccount(1); !ok {
		t.Fatal("account was not connected")
	}

	gameWorld.SetInput(1, game.ShipInput{SelectedPilotToolIndex: 0, PrimaryPointerAction: true})
	gameWorld.Tick(0.1)

	if serverData.EquipmentGroups.Items[600].MagazineCount != 1 {
		t.Fatalf("magazine count = %d, want 1", serverData.EquipmentGroups.Items[600].MagazineCount)
	}
}

// Р В РЎСџР РЋР вЂљР В РЎвЂўР В Р вЂ Р В Р’ВµР РЋР вЂљР РЋР РЏР В Р’ВµР РЋРІР‚С™, Р РЋРІР‚РЋР РЋРІР‚С™Р В РЎвЂў Р РЋР С“Р В РЎвЂўР РЋРІР‚В¦Р РЋР вЂљР В Р’В°Р В Р вЂ¦Р В Р’ВµР В Р вЂ¦Р В РЎвЂР В Р’Вµ Р В РўвЂР В Р’В°Р В Р вЂ¦Р В Р вЂ¦Р РЋРІР‚в„–Р РЋРІР‚В¦ Р В Р’В·Р В Р’В°Р В РЎвЂ”Р В РЎвЂР РЋР С“Р РЋРІР‚в„–Р В Р вЂ Р В Р’В°Р В Р’ВµР РЋРІР‚С™ Р В РЎвЂўР В Р’В±Р В Р вЂ¦Р В РЎвЂўР В Р вЂ Р В Р’В»Р РЋРІР‚ВР В Р вЂ¦Р В Р вЂ¦Р В РЎвЂўР В Р’Вµ Р В РЎвЂ”Р В РЎвЂўР В Р’В»Р В РЎвЂўР В Р’В¶Р В Р’ВµР В Р вЂ¦Р В РЎвЂР В Р’Вµ Р В РЎвЂќР В РЎвЂўР РЋР С“Р В РЎВР В РЎвЂР РЋРІР‚РЋР В Р’ВµР РЋР С“Р В РЎвЂќР В РЎвЂўР В РЎвЂ“Р В РЎвЂў Р В РЎвЂўР В Р’В±Р РЋР вЂ°Р В Р’ВµР В РЎвЂќР РЋРІР‚С™Р В Р’В°.
func TestSelectedWeaponReloadsMagazineFromContainerAmmo(t *testing.T) {
	serverData := testWorldData(t)
	addWeaponTestData(t, &serverData)
	serverData.ItemModels.Items[302].MagazineCapacity = 2
	serverData.ItemModels.Items[302].RechargeTime = 0.1
	serverData.EquipmentGroups.Items[600].MagazineCount = 0
	serverData.EquipmentGroups.Items[601] = &data.EquipmentGroup{
		ID:                   601,
		CosmicObjectID:       1,
		Title:                "Ammo box",
		EquipmentItemModelID: 301,
		Count:                1,
		EnabledCount:         1,
		Enabled:              true,
	}
	serverData.EquipmentGroups.MaxID = 601
	if _, err := serverData.ItemGroups.Add(&data.ItemGroup{
		ContainerEquipmentGroupID: 601,
		ContentItemModelID:        303,
		Count:                     5,
	}); err != nil {
		t.Fatal(err)
	}
	if err := serverData.EquipmentGroups.RebuildIndexes(); err != nil {
		t.Fatal(err)
	}
	gameWorld := world.New(1, serverData)
	if _, ok := gameWorld.ConnectAccount(1); !ok {
		t.Fatal("account was not connected")
	}

	gameWorld.SetInput(1, game.ShipInput{SelectedPilotToolIndex: 0, PrimaryPointerAction: true})
	gameWorld.Tick(0.01)
	gameWorld.SetInput(1, game.ShipInput{SelectedPilotToolIndex: 0, PrimaryPointerAction: false})
	gameWorld.Tick(0.11)

	if serverData.EquipmentGroups.Items[600].MagazineCount != 2 {
		t.Fatalf("magazine count = %d, want 2", serverData.EquipmentGroups.Items[600].MagazineCount)
	}
	items := serverData.ItemGroups.GetByContainerEquipmentGroupID(601)
	if len(items) != 1 || items[0].Count != 3 {
		t.Fatalf("ammo in container = %+v, want one group with 3", items)
	}
}

func TestSelectedWeaponProjectileDamagesShipArmorOnBodyHit(t *testing.T) {
	serverData := testWorldData(t)
	addWeaponTestData(t, &serverData)
	serverData.CosmicObjects.Items[1].X = 0
	serverData.CosmicObjects.Items[1].Y = 0
	serverData.CosmicObjects.Items[1].Rotation = 0
	serverData.CosmicObjects.Items[1].Anchored = true
	serverData.CosmicObjects.Items[4] = &data.CosmicObject{ID: 4, Title: "Target ship", CosmicObjectModelID: 1, X: 0, Y: 100, MaxArmor: 300, Armor: 300, Enabled: true, Anchored: true}
	serverData.CosmicObjects.MaxID = 4
	if err := serverData.CosmicObjects.RebuildIndexes(); err != nil {
		t.Fatal(err)
	}
	gameWorld := world.New(1, serverData)
	if _, ok := gameWorld.ConnectAccount(1); !ok {
		t.Fatal("account was not connected")
	}

	gameWorld.SetInput(1, game.ShipInput{SelectedPilotToolIndex: 0, PrimaryPointerAction: true})
	gameWorld.Tick(0.1)
	closeWorldFloat(t, serverData.CosmicObjects.Items[4].Armor, 300)
	gameWorld.SetInput(1, game.ShipInput{SelectedPilotToolIndex: 0, PrimaryPointerAction: false})
	snapshot := gameWorld.Tick(1)

	closeWorldFloat(t, serverData.CosmicObjects.Items[4].Armor, 180)
	if serverData.CosmicObjects.Items[4].LastReceivedDamageTime == 0 {
		t.Fatalf("damage time was not updated")
	}
	if _, ok := findCosmicObjectModelInSnapshot(snapshot, 901); ok {
		t.Fatalf("projectile was not removed after hit: %+v", snapshot.Objects)
	}
}

// Р В РЎСџР РЋР вЂљР В РЎвЂўР В Р вЂ Р В Р’ВµР РЋР вЂљР РЋР РЏР В Р’ВµР РЋРІР‚С™, Р РЋРІР‚РЋР РЋРІР‚С™Р В РЎвЂў Р РЋР С“Р В РЎвЂўР РЋРІР‚В¦Р РЋР вЂљР В Р’В°Р В Р вЂ¦Р В Р’ВµР В Р вЂ¦Р В РЎвЂР В Р’Вµ Р В РўвЂР В Р’В°Р В Р вЂ¦Р В Р вЂ¦Р РЋРІР‚в„–Р РЋРІР‚В¦ Р В Р’В·Р В Р’В°Р В РЎвЂ”Р В РЎвЂР РЋР С“Р РЋРІР‚в„–Р В Р вЂ Р В Р’В°Р В Р’ВµР РЋРІР‚С™ Р В РЎвЂўР В Р’В±Р В Р вЂ¦Р В РЎвЂўР В Р вЂ Р В Р’В»Р РЋРІР‚ВР В Р вЂ¦Р В Р вЂ¦Р В РЎвЂўР В Р’Вµ Р В РЎвЂ”Р В РЎвЂўР В Р’В»Р В РЎвЂўР В Р’В¶Р В Р’ВµР В Р вЂ¦Р В РЎвЂР В Р’Вµ Р В РЎвЂќР В РЎвЂўР РЋР С“Р В РЎВР В РЎвЂР РЋРІР‚РЋР В Р’ВµР РЋР С“Р В РЎвЂќР В РЎвЂўР В РЎвЂ“Р В РЎвЂў Р В РЎвЂўР В Р’В±Р РЋР вЂ°Р В Р’ВµР В РЎвЂќР РЋРІР‚С™Р В Р’В°.
func TestSelectedWeaponProjectileDamagesStationArmorOnBodyHit(t *testing.T) {
	serverData := testWorldData(t)
	addWeaponTestData(t, &serverData)
	serverData.CosmicObjects.Items[1].X = 0
	serverData.CosmicObjects.Items[1].Y = 0
	serverData.CosmicObjects.Items[1].Rotation = 0
	serverData.CosmicObjects.Items[1].Anchored = true
	serverData.CosmicObjects.Items[3].X = 0
	serverData.CosmicObjects.Items[3].Y = 100
	serverData.CosmicObjects.Items[3].MaxArmor = 300
	serverData.CosmicObjects.Items[3].Armor = 300
	serverData.CosmicObjects.Items[3].Enabled = true
	serverData.CosmicObjects.Items[3].Anchored = true
	gameWorld := world.New(1, serverData)
	if _, ok := gameWorld.ConnectAccount(1); !ok {
		t.Fatal("account was not connected")
	}

	gameWorld.SetInput(1, game.ShipInput{SelectedPilotToolIndex: 0, PrimaryPointerAction: true})
	gameWorld.Tick(0.1)
	gameWorld.SetInput(1, game.ShipInput{SelectedPilotToolIndex: 0, PrimaryPointerAction: false})
	gameWorld.Tick(1)

	closeWorldFloat(t, serverData.CosmicObjects.Items[3].Armor, 180)
}

// Р В РЎСџР РЋР вЂљР В РЎвЂўР В Р вЂ Р В Р’ВµР РЋР вЂљР РЋР РЏР В Р’ВµР РЋРІР‚С™, Р РЋРІР‚РЋР РЋРІР‚С™Р В РЎвЂў Р РЋР С“Р В РЎвЂўР РЋРІР‚В¦Р РЋР вЂљР В Р’В°Р В Р вЂ¦Р В Р’ВµР В Р вЂ¦Р В РЎвЂР В Р’Вµ Р В РўвЂР В Р’В°Р В Р вЂ¦Р В Р вЂ¦Р РЋРІР‚в„–Р РЋРІР‚В¦ Р В Р’В·Р В Р’В°Р В РЎвЂ”Р В РЎвЂР РЋР С“Р РЋРІР‚в„–Р В Р вЂ Р В Р’В°Р В Р’ВµР РЋРІР‚С™ Р В РЎвЂўР В Р’В±Р В Р вЂ¦Р В РЎвЂўР В Р вЂ Р В Р’В»Р РЋРІР‚ВР В Р вЂ¦Р В Р вЂ¦Р В РЎвЂўР В Р’Вµ Р В РЎвЂ”Р В РЎвЂўР В Р’В»Р В РЎвЂўР В Р’В¶Р В Р’ВµР В Р вЂ¦Р В РЎвЂР В Р’Вµ Р В РЎвЂќР В РЎвЂўР РЋР С“Р В РЎВР В РЎвЂР РЋРІР‚РЋР В Р’ВµР РЋР С“Р В РЎвЂќР В РЎвЂўР В РЎвЂ“Р В РЎвЂў Р В РЎвЂўР В Р’В±Р РЋР вЂ°Р В Р’ВµР В РЎвЂќР РЋРІР‚С™Р В Р’В°.
func TestSelectedWeaponProjectileIgnoresAsteroidArmorOnBodyHit(t *testing.T) {
	serverData := testWorldData(t)
	addWeaponTestData(t, &serverData)
	serverData.CosmicObjects.Items[1].X = 0
	serverData.CosmicObjects.Items[1].Y = 0
	serverData.CosmicObjects.Items[1].Rotation = 0
	serverData.CosmicObjects.Items[1].Anchored = true
	serverData.CosmicObjects.Items[2].X = 0
	serverData.CosmicObjects.Items[2].Y = 100
	serverData.CosmicObjects.Items[2].MaxArmor = 300
	serverData.CosmicObjects.Items[2].Armor = 300
	serverData.CosmicObjects.Items[2].Enabled = true
	serverData.CosmicObjects.Items[2].Anchored = true
	gameWorld := world.New(1, serverData)
	if _, ok := gameWorld.ConnectAccount(1); !ok {
		t.Fatal("account was not connected")
	}

	gameWorld.SetInput(1, game.ShipInput{SelectedPilotToolIndex: 0, PrimaryPointerAction: true})
	gameWorld.Tick(0.1)
	gameWorld.SetInput(1, game.ShipInput{SelectedPilotToolIndex: 0, PrimaryPointerAction: false})
	snapshot := gameWorld.Tick(1)

	closeWorldFloat(t, serverData.CosmicObjects.Items[2].Armor, 300)
	if _, ok := findCosmicObjectModelInSnapshot(snapshot, 901); ok {
		t.Fatalf("projectile was not removed after asteroid hit: %+v", snapshot.Objects)
	}
}

func TestDockingRequestAutoApprovesOwnedReceiverAndCompletesCluster(t *testing.T) {
	serverData := testWorldData(t)
	serverData.CosmicObjects.Items[1].Anchored = true
	serverData.CosmicObjects.Items[1].Speed = 0
	serverData.CosmicObjects.Items[1].AngularSpeed = 0
	serverData.CosmicObjects.Items[3].OwnerCharacterID = 1
	serverData.CosmicObjects.Items[3].X = 0
	serverData.CosmicObjects.Items[3].Y = 14
	serverData.CosmicObjects.Items[3].Anchored = true
	gameWorld := world.New(1, serverData)
	if _, ok := gameWorld.ConnectAccount(1); !ok {
		t.Fatal("account was not connected")
	}

	if err := gameWorld.SendDockingRequest(1); err != nil {
		t.Fatalf("send docking request: %v", err)
	}
	gameWorld.Tick(10)

	sender := serverData.CosmicObjects.Items[1]
	receiver := serverData.CosmicObjects.Items[3]
	if sender.ClusterMainCosmicObjectID != receiver.ID {
		t.Fatalf("sender cluster main = %d, want %d", sender.ClusterMainCosmicObjectID, receiver.ID)
	}
	if receiver.ClusterMainCosmicObjectID != receiver.ID {
		t.Fatalf("receiver cluster main = %d, want %d", receiver.ClusterMainCosmicObjectID, receiver.ID)
	}
	if !sender.Anchored || !receiver.Anchored {
		t.Fatalf("cluster objects must stay anchored: sender=%v receiver=%v", sender.Anchored, receiver.Anchored)
	}
}

// Р В РЎСџР РЋР вЂљР В РЎвЂўР В Р вЂ Р В Р’ВµР РЋР вЂљР РЋР РЏР В Р’ВµР РЋРІР‚С™, Р РЋРІР‚РЋР РЋРІР‚С™Р В РЎвЂў Р РЋР С“Р В РЎвЂўР РЋРІР‚В¦Р РЋР вЂљР В Р’В°Р В Р вЂ¦Р В Р’ВµР В Р вЂ¦Р В РЎвЂР В Р’Вµ Р В РўвЂР В Р’В°Р В Р вЂ¦Р В Р вЂ¦Р РЋРІР‚в„–Р РЋРІР‚В¦ Р В Р’В·Р В Р’В°Р В РЎвЂ”Р В РЎвЂР РЋР С“Р РЋРІР‚в„–Р В Р вЂ Р В Р’В°Р В Р’ВµР РЋРІР‚С™ Р В РЎвЂўР В Р’В±Р В Р вЂ¦Р В РЎвЂўР В Р вЂ Р В Р’В»Р РЋРІР‚ВР В Р вЂ¦Р В Р вЂ¦Р В РЎвЂўР В Р’Вµ Р В РЎвЂ”Р В РЎвЂўР В Р’В»Р В РЎвЂўР В Р’В¶Р В Р’ВµР В Р вЂ¦Р В РЎвЂР В Р’Вµ Р В РЎвЂќР В РЎвЂўР РЋР С“Р В РЎВР В РЎвЂР РЋРІР‚РЋР В Р’ВµР РЋР С“Р В РЎвЂќР В РЎвЂўР В РЎвЂ“Р В РЎвЂў Р В РЎвЂўР В Р’В±Р РЋР вЂ°Р В Р’ВµР В РЎвЂќР РЋРІР‚С™Р В Р’В°.
// Р В РЎСџР РЋР вЂљР В РЎвЂўР В Р вЂ Р В Р’ВµР РЋР вЂљР РЋР РЏР В Р’ВµР РЋРІР‚С™, Р РЋРІР‚РЋР РЋРІР‚С™Р В РЎвЂў Р РЋР С“Р В РЎвЂўР РЋРІР‚В¦Р РЋР вЂљР В Р’В°Р В Р вЂ¦Р В Р’ВµР В Р вЂ¦Р В РЎвЂР В Р’Вµ Р В РўвЂР В Р’В°Р В Р вЂ¦Р В Р вЂ¦Р РЋРІР‚в„–Р РЋРІР‚В¦ Р В Р’В·Р В Р’В°Р В РЎвЂ”Р В РЎвЂР РЋР С“Р РЋРІР‚в„–Р В Р вЂ Р В Р’В°Р В Р’ВµР РЋРІР‚С™ Р В РЎвЂўР В Р’В±Р В Р вЂ¦Р В РЎвЂўР В Р вЂ Р В Р’В»Р РЋРІР‚ВР В Р вЂ¦Р В Р вЂ¦Р В РЎвЂўР В Р’Вµ Р В РЎвЂ”Р В РЎвЂўР В Р’В»Р В РЎвЂўР В Р’В¶Р В Р’ВµР В Р вЂ¦Р В РЎвЂР В Р’Вµ Р В РЎвЂќР В РЎвЂўР РЋР С“Р В РЎВР В РЎвЂР РЋРІР‚РЋР В Р’ВµР РЋР С“Р В РЎвЂќР В РЎвЂўР В РЎвЂ“Р В РЎвЂў Р В РЎвЂўР В Р’В±Р РЋР вЂ°Р В Р’ВµР В РЎвЂќР РЋРІР‚С™Р В Р’В°.
func TestDockingRequestDoesNotRequireAnchoredObjects(t *testing.T) {
	serverData := testWorldData(t)
	serverData.CosmicObjects.Items[1].Anchored = false
	serverData.CosmicObjects.Items[1].Speed = 0
	serverData.CosmicObjects.Items[1].AngularSpeed = 0
	serverData.CosmicObjects.Items[3].OwnerCharacterID = 1
	serverData.CosmicObjects.Items[3].X = 0
	serverData.CosmicObjects.Items[3].Y = 14
	serverData.CosmicObjects.Items[3].Anchored = false
	gameWorld := world.New(1, serverData)
	if _, ok := gameWorld.ConnectAccount(1); !ok {
		t.Fatal("account was not connected")
	}

	if err := gameWorld.SendDockingRequest(1); err != nil {
		t.Fatalf("send docking request without anchors: %v", err)
	}
}

// Р В РЎСџР РЋР вЂљР В РЎвЂўР В Р вЂ Р В Р’ВµР РЋР вЂљР РЋР РЏР В Р’ВµР РЋРІР‚С™, Р РЋРІР‚РЋР РЋРІР‚С™Р В РЎвЂў Р РЋР С“Р В РЎвЂўР РЋРІР‚В¦Р РЋР вЂљР В Р’В°Р В Р вЂ¦Р В Р’ВµР В Р вЂ¦Р В РЎвЂР В Р’Вµ Р В РўвЂР В Р’В°Р В Р вЂ¦Р В Р вЂ¦Р РЋРІР‚в„–Р РЋРІР‚В¦ Р В Р’В·Р В Р’В°Р В РЎвЂ”Р В РЎвЂР РЋР С“Р РЋРІР‚в„–Р В Р вЂ Р В Р’В°Р В Р’ВµР РЋРІР‚С™ Р В РЎвЂўР В Р’В±Р В Р вЂ¦Р В РЎвЂўР В Р вЂ Р В Р’В»Р РЋРІР‚ВР В Р вЂ¦Р В Р вЂ¦Р В РЎвЂўР В Р’Вµ Р В РЎвЂ”Р В РЎвЂўР В Р’В»Р В РЎвЂўР В Р’В¶Р В Р’ВµР В Р вЂ¦Р В РЎвЂР В Р’Вµ Р В РЎвЂќР В РЎвЂўР РЋР С“Р В РЎВР В РЎвЂР РЋРІР‚РЋР В Р’ВµР РЋР С“Р В РЎвЂќР В РЎвЂўР В РЎвЂ“Р В РЎвЂў Р В РЎвЂўР В Р’В±Р РЋР вЂ°Р В Р’ВµР В РЎвЂќР РЋРІР‚С™Р В Р’В°.
func TestClaimFocusedObjectOwnerForTestingChangesProbeTargetOwner(t *testing.T) {
	serverData := testWorldData(t)
	serverData.CosmicObjects.Items[3].OwnerCharacterID = 0
	serverData.CosmicObjects.Items[3].X = 0
	serverData.CosmicObjects.Items[3].Y = 14
	gameWorld := world.New(1, serverData)
	if _, ok := gameWorld.ConnectAccount(1); !ok {
		t.Fatal("account was not connected")
	}

	if err := gameWorld.ClaimFocusedObjectOwnerForTesting(1); err != nil {
		t.Fatalf("claim focused object owner: %v", err)
	}

	if serverData.CosmicObjects.Items[3].OwnerCharacterID != 1 {
		t.Fatalf("owner was not changed: %d", serverData.CosmicObjects.Items[3].OwnerCharacterID)
	}
}

// Р В РЎСџР РЋР вЂљР В РЎвЂўР В Р вЂ Р В Р’ВµР РЋР вЂљР РЋР РЏР В Р’ВµР РЋРІР‚С™, Р РЋРІР‚РЋР РЋРІР‚С™Р В РЎвЂў Р РЋР С“Р В РЎвЂўР РЋРІР‚В¦Р РЋР вЂљР В Р’В°Р В Р вЂ¦Р В Р’ВµР В Р вЂ¦Р В РЎвЂР В Р’Вµ Р В РўвЂР В Р’В°Р В Р вЂ¦Р В Р вЂ¦Р РЋРІР‚в„–Р РЋРІР‚В¦ Р В Р’В·Р В Р’В°Р В РЎвЂ”Р В РЎвЂР РЋР С“Р РЋРІР‚в„–Р В Р вЂ Р В Р’В°Р В Р’ВµР РЋРІР‚С™ Р В РЎвЂўР В Р’В±Р В Р вЂ¦Р В РЎвЂўР В Р вЂ Р В Р’В»Р РЋРІР‚ВР В Р вЂ¦Р В Р вЂ¦Р В РЎвЂўР В Р’Вµ Р В РЎвЂ”Р В РЎвЂўР В Р’В»Р В РЎвЂўР В Р’В¶Р В Р’ВµР В Р вЂ¦Р В РЎвЂР В Р’Вµ Р В РЎвЂќР В РЎвЂўР РЋР С“Р В РЎВР В РЎвЂР РЋРІР‚РЋР В Р’ВµР РЋР С“Р В РЎвЂќР В РЎвЂўР В РЎвЂ“Р В РЎвЂў Р В РЎвЂўР В Р’В±Р РЋР вЂ°Р В Р’ВµР В РЎвЂќР РЋРІР‚С™Р В Р’В°.
func TestSnapshotForAccountIncludesOwnerNameForTesting(t *testing.T) {
	serverData := testWorldData(t)
	serverData.Accounts.Items[2] = &data.Account{ID: 2, Email: "owner@email.net", Nickname: "OwnerName", PasswordHash: "hash", Token: "token-2", CurrentCharacterID: 2}
	serverData.Characters.Items[2] = &data.Character{ID: 2, AccountID: 2, LocationCosmicObjectID: 3}
	serverData.CosmicObjects.Items[3].OwnerCharacterID = 2
	gameWorld := world.New(1, serverData)
	if _, ok := gameWorld.ConnectAccount(1); !ok {
		t.Fatal("account was not connected")
	}

	snapshot := gameWorld.SnapshotForAccount(1)
	object, ok := findCosmicObjectInSnapshot(snapshot, 3)
	if !ok {
		t.Fatal("owned object was not found in snapshot")
	}

	if object.OwnerName != "OwnerName" {
		t.Fatalf("owner name = %q, want OwnerName", object.OwnerName)
	}
}

func TestApproveDockingRequestStartsProcessForForeignReceiver(t *testing.T) {
	serverData := testWorldData(t)
	serverData.Accounts.Items[2] = &data.Account{ID: 2, Email: "receiver@email.net", Nickname: "receiver", PasswordHash: "hash", Token: "token-2", CurrentCharacterID: 2}
	serverData.Characters.Items[2] = &data.Character{ID: 2, AccountID: 2, LocationCosmicObjectID: 3}
	serverData.CosmicObjects.Items[1].Anchored = true
	serverData.CosmicObjects.Items[3].OwnerCharacterID = 2
	serverData.CosmicObjects.Items[3].X = 0
	serverData.CosmicObjects.Items[3].Y = 14
	serverData.CosmicObjects.Items[3].Anchored = true
	if err := serverData.Accounts.RebuildIndexes(); err != nil {
		t.Fatal(err)
	}
	if err := serverData.Characters.RebuildIndexes(); err != nil {
		t.Fatal(err)
	}
	gameWorld := world.New(1, serverData)
	if _, ok := gameWorld.ConnectAccount(1); !ok {
		t.Fatal("sender account was not connected")
	}
	if _, ok := gameWorld.ConnectAccount(2); !ok {
		t.Fatal("receiver account was not connected")
	}

	if err := gameWorld.SendDockingRequest(1); err != nil {
		t.Fatalf("send docking request: %v", err)
	}
	if err := gameWorld.ApproveDockingRequest(2); err != nil {
		t.Fatalf("approve docking request: %v", err)
	}
	gameWorld.Tick(10)

	if serverData.CosmicObjects.Items[1].ClusterMainCosmicObjectID != 3 || serverData.CosmicObjects.Items[3].ClusterMainCosmicObjectID != 3 {
		t.Fatalf("cluster was not completed: sender=%d receiver=%d", serverData.CosmicObjects.Items[1].ClusterMainCosmicObjectID, serverData.CosmicObjects.Items[3].ClusterMainCosmicObjectID)
	}
}

// Р В РЎСџР РЋР вЂљР В РЎвЂўР В Р вЂ Р В Р’ВµР РЋР вЂљР РЋР РЏР В Р’ВµР РЋРІР‚С™, Р РЋРІР‚РЋР РЋРІР‚С™Р В РЎвЂў Р РЋР С“Р В РЎвЂўР РЋРІР‚В¦Р РЋР вЂљР В Р’В°Р В Р вЂ¦Р В Р’ВµР В Р вЂ¦Р В РЎвЂР В Р’Вµ Р В РўвЂР В Р’В°Р В Р вЂ¦Р В Р вЂ¦Р РЋРІР‚в„–Р РЋРІР‚В¦ Р В Р’В·Р В Р’В°Р В РЎвЂ”Р В РЎвЂР РЋР С“Р РЋРІР‚в„–Р В Р вЂ Р В Р’В°Р В Р’ВµР РЋРІР‚С™ Р В РЎвЂўР В Р’В±Р В Р вЂ¦Р В РЎвЂўР В Р вЂ Р В Р’В»Р РЋРІР‚ВР В Р вЂ¦Р В Р вЂ¦Р В РЎвЂўР В Р’Вµ Р В РЎвЂ”Р В РЎвЂўР В Р’В»Р В РЎвЂўР В Р’В¶Р В Р’ВµР В Р вЂ¦Р В РЎвЂР В Р’Вµ Р В РЎвЂќР В РЎвЂўР РЋР С“Р В РЎВР В РЎвЂР РЋРІР‚РЋР В Р’ВµР РЋР С“Р В РЎвЂќР В РЎвЂўР В РЎвЂ“Р В РЎвЂў Р В РЎвЂўР В Р’В±Р РЋР вЂ°Р В Р’ВµР В РЎвЂќР РЋРІР‚С™Р В Р’В°.
// Р В РЎСџР РЋР вЂљР В РЎвЂўР В Р вЂ Р В Р’ВµР РЋР вЂљР РЋР РЏР В Р’ВµР РЋРІР‚С™, Р РЋРІР‚РЋР РЋРІР‚С™Р В РЎвЂў Р РЋР С“Р В РЎвЂўР РЋРІР‚В¦Р РЋР вЂљР В Р’В°Р В Р вЂ¦Р В Р’ВµР В Р вЂ¦Р В РЎвЂР В Р’Вµ Р В РўвЂР В Р’В°Р В Р вЂ¦Р В Р вЂ¦Р РЋРІР‚в„–Р РЋРІР‚В¦ Р В Р’В·Р В Р’В°Р В РЎвЂ”Р В РЎвЂР РЋР С“Р РЋРІР‚в„–Р В Р вЂ Р В Р’В°Р В Р’ВµР РЋРІР‚С™ Р В РЎвЂўР В Р’В±Р В Р вЂ¦Р В РЎвЂўР В Р вЂ Р В Р’В»Р РЋРІР‚ВР В Р вЂ¦Р В Р вЂ¦Р В РЎвЂўР В Р’Вµ Р В РЎвЂ”Р В РЎвЂўР В Р’В»Р В РЎвЂўР В Р’В¶Р В Р’ВµР В Р вЂ¦Р В РЎвЂР В Р’Вµ Р В РЎвЂќР В РЎвЂўР РЋР С“Р В РЎВР В РЎвЂР РЋРІР‚РЋР В Р’ВµР РЋР С“Р В РЎвЂќР В РЎвЂўР В РЎвЂ“Р В РЎвЂў Р В РЎвЂўР В Р’В±Р РЋР вЂ°Р В Р’ВµР В РЎвЂќР РЋРІР‚С™Р В Р’В°.
func TestDockingRequestTimeoutClosesRequestWindow(t *testing.T) {
	serverData := testWorldData(t)
	serverData.Accounts.Items[2] = &data.Account{ID: 2, Email: "receiver@email.net", Nickname: "receiver", PasswordHash: "hash", Token: "token-2", CurrentCharacterID: 2}
	serverData.Characters.Items[2] = &data.Character{ID: 2, AccountID: 2, LocationCosmicObjectID: 3}
	serverData.CosmicObjects.Items[3].OwnerCharacterID = 2
	serverData.CosmicObjects.Items[3].X = 0
	serverData.CosmicObjects.Items[3].Y = 14
	if err := serverData.Accounts.RebuildIndexes(); err != nil {
		t.Fatal(err)
	}
	if err := serverData.Characters.RebuildIndexes(); err != nil {
		t.Fatal(err)
	}
	gameWorld := world.New(1, serverData)
	if _, ok := gameWorld.ConnectAccount(1); !ok {
		t.Fatal("sender account was not connected")
	}
	if _, ok := gameWorld.ConnectAccount(2); !ok {
		t.Fatal("receiver account was not connected")
	}

	if err := gameWorld.SendDockingRequest(1); err != nil {
		t.Fatalf("send docking request: %v", err)
	}
	_ = gameWorld.DrainDockingEvents()
	gameWorld.Tick(10)

	if !dockingEventsContainKind(gameWorld.DrainDockingEvents(), "dockingFinished") {
		t.Fatal("timeout did not close docking request window")
	}
}

// Р В РЎСџР РЋР вЂљР В РЎвЂўР В Р вЂ Р В Р’ВµР РЋР вЂљР РЋР РЏР В Р’ВµР РЋРІР‚С™, Р РЋРІР‚РЋР РЋРІР‚С™Р В РЎвЂў Р РЋР С“Р В РЎвЂўР РЋРІР‚В¦Р РЋР вЂљР В Р’В°Р В Р вЂ¦Р В Р’ВµР В Р вЂ¦Р В РЎвЂР В Р’Вµ Р В РўвЂР В Р’В°Р В Р вЂ¦Р В Р вЂ¦Р РЋРІР‚в„–Р РЋРІР‚В¦ Р В Р’В·Р В Р’В°Р В РЎвЂ”Р В РЎвЂР РЋР С“Р РЋРІР‚в„–Р В Р вЂ Р В Р’В°Р В Р’ВµР РЋРІР‚С™ Р В РЎвЂўР В Р’В±Р В Р вЂ¦Р В РЎвЂўР В Р вЂ Р В Р’В»Р РЋРІР‚ВР В Р вЂ¦Р В Р вЂ¦Р В РЎвЂўР В Р’Вµ Р В РЎвЂ”Р В РЎвЂўР В Р’В»Р В РЎвЂўР В Р’В¶Р В Р’ВµР В Р вЂ¦Р В РЎвЂР В Р’Вµ Р В РЎвЂќР В РЎвЂўР РЋР С“Р В РЎВР В РЎвЂР РЋРІР‚РЋР В Р’ВµР РЋР С“Р В РЎвЂќР В РЎвЂўР В РЎвЂ“Р В РЎвЂў Р В РЎвЂўР В Р’В±Р РЋР вЂ°Р В Р’ВµР В РЎвЂќР РЋРІР‚С™Р В Р’В°.
// Проверяет, что чужой корабль без пилота не открывает окно запроса и отправляет уведомление.
func TestDockingRequestRejectsForeignShipWithoutPilotImmediately(t *testing.T) {
	serverData := testWorldData(t)
	serverData.CosmicObjects.Items[3].OwnerCharacterID = 2
	serverData.CosmicObjects.Items[3].X = 0
	serverData.CosmicObjects.Items[3].Y = 14
	gameWorld := world.New(1, serverData)
	if _, ok := gameWorld.ConnectAccount(1); !ok {
		t.Fatal("sender account was not connected")
	}

	if err := gameWorld.SendDockingRequest(1); err != nil {
		t.Fatalf("send docking request: %v", err)
	}
	events := gameWorld.DrainDockingEvents()

	if dockingEventsContainKind(events, "dockingRequestStarted") {
		t.Fatal("request window was shown for receiver without pilot")
	}
	if !dockingEventsContainMessage(events, "В Получателе нет персонажа для принятия решения") {
		t.Fatalf("missing no-pilot notification: %+v", events)
	}
}

// Р В РЎСџР РЋР вЂљР В РЎвЂўР В Р вЂ Р В Р’ВµР РЋР вЂљР РЋР РЏР В Р’ВµР РЋРІР‚С™, Р РЋРІР‚РЋР РЋРІР‚С™Р В РЎвЂў Р РЋР С“Р В РЎвЂўР РЋРІР‚В¦Р РЋР вЂљР В Р’В°Р В Р вЂ¦Р В Р’ВµР В Р вЂ¦Р В РЎвЂР В Р’Вµ Р В РўвЂР В Р’В°Р В Р вЂ¦Р В Р вЂ¦Р РЋРІР‚в„–Р РЋРІР‚В¦ Р В Р’В·Р В Р’В°Р В РЎвЂ”Р В РЎвЂР РЋР С“Р РЋРІР‚в„–Р В Р вЂ Р В Р’В°Р В Р’ВµР РЋРІР‚С™ Р В РЎвЂўР В Р’В±Р В Р вЂ¦Р В РЎвЂўР В Р вЂ Р В Р’В»Р РЋРІР‚ВР В Р вЂ¦Р В Р вЂ¦Р В РЎвЂўР В Р’Вµ Р В РЎвЂ”Р В РЎвЂўР В Р’В»Р В РЎвЂўР В Р’В¶Р В Р’ВµР В Р вЂ¦Р В РЎвЂР В Р’Вµ Р В РЎвЂќР В РЎвЂўР РЋР С“Р В РЎВР В РЎвЂР РЋРІР‚РЋР В Р’ВµР РЋР С“Р В РЎвЂќР В РЎвЂўР В РЎвЂ“Р В РЎвЂў Р В РЎвЂўР В Р’В±Р РЋР вЂ°Р В Р’ВµР В РЎвЂќР РЋРІР‚С™Р В Р’В°.
func TestBeginCharacterTransferFromSecondaryToOwnedMainMovesCharacter(t *testing.T) {
	serverData := testWorldData(t)
	serverData.CosmicObjects.Items[1].ClusterMainCosmicObjectID = 3
	serverData.CosmicObjects.Items[3].ClusterMainCosmicObjectID = 3
	serverData.CosmicObjects.Items[3].OwnerCharacterID = 1
	gameWorld := world.New(1, serverData)
	if _, ok := gameWorld.ConnectAccount(1); !ok {
		t.Fatal("account was not connected")
	}

	if err := gameWorld.BeginCharacterTransfer(1); err != nil {
		t.Fatalf("begin character transfer: %v", err)
	}

	if serverData.Characters.Items[1].LocationCosmicObjectID != 3 {
		t.Fatalf("character location = %d, want 3", serverData.Characters.Items[1].LocationCosmicObjectID)
	}
	if objectID, ok := gameWorld.ConnectAccount(1); !ok || objectID != 3 {
		t.Fatalf("controlled object after reconnect = %d, %v; want 3, true", objectID, ok)
	}
}

// Р В РЎСџР РЋР вЂљР В РЎвЂўР В Р вЂ Р В Р’ВµР РЋР вЂљР РЋР РЏР В Р’ВµР РЋРІР‚С™, Р РЋРІР‚РЋР РЋРІР‚С™Р В РЎвЂў Р РЋР С“Р В РЎвЂўР РЋРІР‚В¦Р РЋР вЂљР В Р’В°Р В Р вЂ¦Р В Р’ВµР В Р вЂ¦Р В РЎвЂР В Р’Вµ Р В РўвЂР В Р’В°Р В Р вЂ¦Р В Р вЂ¦Р РЋРІР‚в„–Р РЋРІР‚В¦ Р В Р’В·Р В Р’В°Р В РЎвЂ”Р В РЎвЂР РЋР С“Р РЋРІР‚в„–Р В Р вЂ Р В Р’В°Р В Р’ВµР РЋРІР‚С™ Р В РЎвЂўР В Р’В±Р В Р вЂ¦Р В РЎвЂўР В Р вЂ Р В Р’В»Р РЋРІР‚ВР В Р вЂ¦Р В Р вЂ¦Р В РЎвЂўР В Р’Вµ Р В РЎвЂ”Р В РЎвЂўР В Р’В»Р В РЎвЂўР В Р’В¶Р В Р’ВµР В Р вЂ¦Р В РЎвЂР В Р’Вµ Р В РЎвЂќР В РЎвЂўР РЋР С“Р В РЎВР В РЎвЂР РЋРІР‚РЋР В Р’ВµР РЋР С“Р В РЎвЂќР В РЎвЂўР В РЎвЂ“Р В РЎвЂў Р В РЎвЂўР В Р’В±Р РЋР вЂ°Р В Р’ВµР В РЎвЂќР РЋРІР‚С™Р В Р’В°.
// Проверяет, что пересадка вне стыкованной группы оставляет персонажа на месте и отправляет уведомление.
func TestBeginCharacterTransferOutsideClusterShowsNotification(t *testing.T) {
	serverData := testWorldData(t)
	gameWorld := world.New(1, serverData)
	if _, ok := gameWorld.ConnectAccount(1); !ok {
		t.Fatal("account was not connected")
	}

	if err := gameWorld.BeginCharacterTransfer(1); err != nil {
		t.Fatalf("begin character transfer: %v", err)
	}

	if serverData.Characters.Items[1].LocationCosmicObjectID != 1 {
		t.Fatalf("character location changed to %d", serverData.Characters.Items[1].LocationCosmicObjectID)
	}
	if !dockingEventsContainMessage(gameWorld.DrainDockingEvents(), "Объект не пристыкован") {
		t.Fatal("missing not-docked notification")
	}
}

// Р В РЎСџР РЋР вЂљР В РЎвЂўР В Р вЂ Р В Р’ВµР РЋР вЂљР РЋР РЏР В Р’ВµР РЋРІР‚С™, Р РЋРІР‚РЋР РЋРІР‚С™Р В РЎвЂў Р РЋР С“Р В РЎвЂўР РЋРІР‚В¦Р РЋР вЂљР В Р’В°Р В Р вЂ¦Р В Р’ВµР В Р вЂ¦Р В РЎвЂР В Р’Вµ Р В РўвЂР В Р’В°Р В Р вЂ¦Р В Р вЂ¦Р РЋРІР‚в„–Р РЋРІР‚В¦ Р В Р’В·Р В Р’В°Р В РЎвЂ”Р В РЎвЂР РЋР С“Р РЋРІР‚в„–Р В Р вЂ Р В Р’В°Р В Р’ВµР РЋРІР‚С™ Р В РЎвЂўР В Р’В±Р В Р вЂ¦Р В РЎвЂўР В Р вЂ Р В Р’В»Р РЋРІР‚ВР В Р вЂ¦Р В Р вЂ¦Р В РЎвЂўР В Р’Вµ Р В РЎвЂ”Р В РЎвЂўР В Р’В»Р В РЎвЂўР В Р’В¶Р В Р’ВµР В Р вЂ¦Р В РЎвЂР В Р’Вµ Р В РЎвЂќР В РЎвЂўР РЋР С“Р В РЎВР В РЎвЂР РЋРІР‚РЋР В Р’ВµР РЋР С“Р В РЎвЂќР В РЎвЂўР В РЎвЂ“Р В РЎвЂў Р В РЎвЂўР В Р’В±Р РЋР вЂ°Р В Р’ВµР В РЎвЂќР РЋРІР‚С™Р В Р’В°.
func TestCharacterTransferToForeignObjectWithPassengerSeatWaitsForApproval(t *testing.T) {
	serverData := testWorldData(t)
	serverData.Accounts.Items[2] = &data.Account{ID: 2, Email: "receiver@email.net", Nickname: "receiver", PasswordHash: "hash", Token: "token-2", CurrentCharacterID: 2}
	serverData.Characters.Items[2] = &data.Character{ID: 2, AccountID: 2, LocationCosmicObjectID: 3}
	serverData.CosmicObjects.Items[1].ClusterMainCosmicObjectID = 3
	serverData.CosmicObjects.Items[3].ClusterMainCosmicObjectID = 3
	serverData.CosmicObjects.Items[3].OwnerCharacterID = 2
	serverData.ItemTypes.Items[19] = &data.ItemType{ID: 19, TitleRu: "Р В РЎСџР В Р’В°Р РЋР С“Р РЋР С“Р В Р’В°Р В Р’В¶Р В РЎвЂР РЋР вЂљР РЋР С“Р В РЎвЂќР В РЎвЂўР В Р’Вµ Р В РЎвЂќР РЋР вЂљР В Р’ВµР РЋР С“Р В Р’В»Р В РЎвЂў", TitleEn: "Passenger Seat", Acronym: "PassengerSeat", CountMustBeInteger: true}
	serverData.ItemModels.Items[501] = &data.ItemModel{ID: 501, TitleRu: "Р В РЎСџР В Р’В°Р РЋР С“Р РЋР С“Р В Р’В°Р В Р’В¶Р В РЎвЂР РЋР вЂљР РЋР С“Р В РЎвЂќР В РЎвЂўР В Р’Вµ Р В РЎвЂќР РЋР вЂљР В Р’ВµР РЋР С“Р В Р’В»Р В РЎвЂў", TitleEn: "Passenger Seat", Acronym: "PassengerSeat", ItemTypeID: 19}
	if _, err := serverData.EquipmentGroups.Add(&data.EquipmentGroup{CosmicObjectID: 3, Title: "Passenger Seat", EquipmentItemModelID: 501, Count: 1, EnabledCount: 1, Enabled: true}); err != nil {
		t.Fatal(err)
	}
	if err := serverData.Accounts.RebuildIndexes(); err != nil {
		t.Fatal(err)
	}
	if err := serverData.Characters.RebuildIndexes(); err != nil {
		t.Fatal(err)
	}
	if err := serverData.ItemTypes.RebuildIndexes(); err != nil {
		t.Fatal(err)
	}
	if err := serverData.ItemModels.RebuildIndexes(); err != nil {
		t.Fatal(err)
	}
	gameWorld := world.New(1, serverData)
	if _, ok := gameWorld.ConnectAccount(1); !ok {
		t.Fatal("sender account was not connected")
	}
	if _, ok := gameWorld.ConnectAccount(2); !ok {
		t.Fatal("receiver account was not connected")
	}

	if err := gameWorld.BeginCharacterTransfer(1); err != nil {
		t.Fatalf("begin character transfer: %v", err)
	}
	if serverData.Characters.Items[1].LocationCosmicObjectID != 1 {
		t.Fatalf("character moved before approval to %d", serverData.Characters.Items[1].LocationCosmicObjectID)
	}
	if !dockingEventsContainKind(gameWorld.DrainDockingEvents(), "landingRequestStarted") {
		t.Fatal("landing request window was not shown")
	}
	if err := gameWorld.ApproveCharacterLanding(2); err != nil {
		t.Fatalf("approve character landing: %v", err)
	}

	if serverData.Characters.Items[1].LocationCosmicObjectID != 3 {
		t.Fatalf("character location = %d, want 3", serverData.Characters.Items[1].LocationCosmicObjectID)
	}
}

func TestUndockMainObjectDisbandsWholeCluster(t *testing.T) {
	serverData := testWorldData(t)
	serverData.CosmicObjects.Items[1].ClusterMainCosmicObjectID = 3
	serverData.CosmicObjects.Items[3].ClusterMainCosmicObjectID = 3
	serverData.Accounts.Items[1].CurrentCharacterID = 1
	serverData.Characters.Items[1].LocationCosmicObjectID = 3
	gameWorld := world.New(1, serverData)
	if _, ok := gameWorld.ConnectAccount(1); !ok {
		t.Fatal("account was not connected")
	}

	if err := gameWorld.UndockControlledObject(1); err != nil {
		t.Fatalf("undock main object: %v", err)
	}

	if serverData.CosmicObjects.Items[1].ClusterMainCosmicObjectID != 0 || serverData.CosmicObjects.Items[3].ClusterMainCosmicObjectID != 0 {
		t.Fatalf("cluster was not disbanded: sender=%d receiver=%d", serverData.CosmicObjects.Items[1].ClusterMainCosmicObjectID, serverData.CosmicObjects.Items[3].ClusterMainCosmicObjectID)
	}
}

// Р В РЎСџР РЋР вЂљР В РЎвЂўР В Р вЂ Р В Р’ВµР РЋР вЂљР РЋР РЏР В Р’ВµР РЋРІР‚С™, Р РЋРІР‚РЋР РЋРІР‚С™Р В РЎвЂў Р РЋР С“Р В РЎвЂўР РЋРІР‚В¦Р РЋР вЂљР В Р’В°Р В Р вЂ¦Р В Р’ВµР В Р вЂ¦Р В РЎвЂР В Р’Вµ Р В РўвЂР В Р’В°Р В Р вЂ¦Р В Р вЂ¦Р РЋРІР‚в„–Р РЋРІР‚В¦ Р В Р’В·Р В Р’В°Р В РЎвЂ”Р В РЎвЂР РЋР С“Р РЋРІР‚в„–Р В Р вЂ Р В Р’В°Р В Р’ВµР РЋРІР‚С™ Р В РЎвЂўР В Р’В±Р В Р вЂ¦Р В РЎвЂўР В Р вЂ Р В Р’В»Р РЋРІР‚ВР В Р вЂ¦Р В Р вЂ¦Р В РЎвЂўР В Р’Вµ Р В РЎвЂ”Р В РЎвЂўР В Р’В»Р В РЎвЂўР В Р’В¶Р В Р’ВµР В Р вЂ¦Р В РЎвЂР В Р’Вµ Р В РЎвЂќР В РЎвЂўР РЋР С“Р В РЎВР В РЎвЂР РЋРІР‚РЋР В Р’ВµР РЋР С“Р В РЎвЂќР В РЎвЂўР В РЎвЂ“Р В РЎвЂў Р В РЎвЂўР В Р’В±Р РЋР вЂ°Р В Р’ВµР В РЎвЂќР РЋРІР‚С™Р В Р’В°.
func TestUndockSecondaryObjectDisbandsTwoObjectCluster(t *testing.T) {
	serverData := testWorldData(t)
	serverData.CosmicObjects.Items[1].Anchored = true
	serverData.CosmicObjects.Items[1].ClusterMainCosmicObjectID = 3
	serverData.CosmicObjects.Items[3].CosmicObjectModelID = 1
	serverData.CosmicObjects.Items[3].Anchored = true
	serverData.CosmicObjects.Items[3].ClusterMainCosmicObjectID = 3
	gameWorld := world.New(1, serverData)
	if _, ok := gameWorld.ConnectAccount(1); !ok {
		t.Fatal("account was not connected")
	}

	if err := gameWorld.UndockControlledObject(1); err != nil {
		t.Fatalf("undock secondary object: %v", err)
	}

	if serverData.CosmicObjects.Items[1].ClusterMainCosmicObjectID != 0 || serverData.CosmicObjects.Items[3].ClusterMainCosmicObjectID != 0 {
		t.Fatalf("two-object cluster was not disbanded: secondary=%d main=%d", serverData.CosmicObjects.Items[1].ClusterMainCosmicObjectID, serverData.CosmicObjects.Items[3].ClusterMainCosmicObjectID)
	}
	if serverData.CosmicObjects.Items[3].Anchored {
		t.Fatal("single remaining object stayed forcibly anchored")
	}
}

// Р В РЎСџР РЋР вЂљР В РЎвЂўР В Р вЂ Р В Р’ВµР РЋР вЂљР РЋР РЏР В Р’ВµР РЋРІР‚С™, Р РЋРІР‚РЋР РЋРІР‚С™Р В РЎвЂў Р РЋР С“Р В РЎвЂўР РЋРІР‚В¦Р РЋР вЂљР В Р’В°Р В Р вЂ¦Р В Р’ВµР В Р вЂ¦Р В РЎвЂР В Р’Вµ Р В РўвЂР В Р’В°Р В Р вЂ¦Р В Р вЂ¦Р РЋРІР‚в„–Р РЋРІР‚В¦ Р В Р’В·Р В Р’В°Р В РЎвЂ”Р В РЎвЂР РЋР С“Р РЋРІР‚в„–Р В Р вЂ Р В Р’В°Р В Р’ВµР РЋРІР‚С™ Р В РЎвЂўР В Р’В±Р В Р вЂ¦Р В РЎвЂўР В Р вЂ Р В Р’В»Р РЋРІР‚ВР В Р вЂ¦Р В Р вЂ¦Р В РЎвЂўР В Р’Вµ Р В РЎвЂ”Р В РЎвЂўР В Р’В»Р В РЎвЂўР В Р’В¶Р В Р’ВµР В Р вЂ¦Р В РЎвЂР В Р’Вµ Р В РЎвЂќР В РЎвЂўР РЋР С“Р В РЎВР В РЎвЂР РЋРІР‚РЋР В Р’ВµР РЋР С“Р В РЎвЂќР В РЎвЂўР В РЎвЂ“Р В РЎвЂў Р В РЎвЂўР В Р’В±Р РЋР вЂ°Р В Р’ВµР В РЎвЂќР РЋРІР‚С™Р В Р’В°.
func TestSetInputDoesNotDisableAnchorForClusterObject(t *testing.T) {
	serverData := testWorldData(t)
	serverData.CosmicObjects.Items[1].Anchored = true
	serverData.CosmicObjects.Items[1].ClusterMainCosmicObjectID = 1
	gameWorld := world.New(1, serverData)
	if _, ok := gameWorld.ConnectAccount(1); !ok {
		t.Fatal("account was not connected")
	}

	gameWorld.SetInput(1, game.ShipInput{ToggleAnchor: true})

	if !serverData.CosmicObjects.Items[1].Anchored {
		t.Fatal("cluster object anchor was disabled")
	}
}

// Р В РЎСџР РЋР вЂљР В РЎвЂўР В Р вЂ Р В Р’ВµР РЋР вЂљР РЋР РЏР В Р’ВµР РЋРІР‚С™, Р РЋРІР‚РЋР РЋРІР‚С™Р В РЎвЂў Р РЋР С“Р В РЎвЂўР РЋРІР‚В¦Р РЋР вЂљР В Р’В°Р В Р вЂ¦Р В Р’ВµР В Р вЂ¦Р В РЎвЂР В Р’Вµ Р В РўвЂР В Р’В°Р В Р вЂ¦Р В Р вЂ¦Р РЋРІР‚в„–Р РЋРІР‚В¦ Р В Р’В·Р В Р’В°Р В РЎвЂ”Р В РЎвЂР РЋР С“Р РЋРІР‚в„–Р В Р вЂ Р В Р’В°Р В Р’ВµР РЋРІР‚С™ Р В РЎвЂўР В Р’В±Р В Р вЂ¦Р В РЎвЂўР В Р вЂ Р В Р’В»Р РЋРІР‚ВР В Р вЂ¦Р В Р вЂ¦Р В РЎвЂўР В Р’Вµ Р В РЎвЂ”Р В РЎвЂўР В Р’В»Р В РЎвЂўР В Р’В¶Р В Р’ВµР В Р вЂ¦Р В РЎвЂР В Р’Вµ Р В РЎвЂќР В РЎвЂўР РЋР С“Р В РЎВР В РЎвЂР РЋРІР‚РЋР В Р’ВµР РЋР С“Р В РЎвЂќР В РЎвЂўР В РЎвЂ“Р В РЎвЂў Р В РЎвЂўР В Р’В±Р РЋР вЂ°Р В Р’ВµР В РЎвЂќР РЋРІР‚С™Р В Р’В°.
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

// Р В РЎСџР РЋР вЂљР В РЎвЂўР В Р вЂ Р В Р’ВµР РЋР вЂљР РЋР РЏР В Р’ВµР РЋРІР‚С™, Р РЋРІР‚РЋР РЋРІР‚С™Р В РЎвЂў Р РЋР С“Р В РЎвЂўР РЋРІР‚В¦Р РЋР вЂљР В Р’В°Р В Р вЂ¦Р В Р’ВµР В Р вЂ¦Р В РЎвЂР В Р’Вµ Р В РўвЂР В Р’В°Р В Р вЂ¦Р В Р вЂ¦Р РЋРІР‚в„–Р РЋРІР‚В¦ Р В Р’В·Р В Р’В°Р В РЎвЂ”Р В РЎвЂР РЋР С“Р РЋРІР‚в„–Р В Р вЂ Р В Р’В°Р В Р’ВµР РЋРІР‚С™ Р В РЎвЂўР В Р’В±Р В Р вЂ¦Р В РЎвЂўР В Р вЂ Р В Р’В»Р РЋРІР‚ВР В Р вЂ¦Р В Р вЂ¦Р В РЎвЂўР В Р’Вµ Р В РЎвЂ”Р В РЎвЂўР В Р’В»Р В РЎвЂўР В Р’В¶Р В Р’ВµР В Р вЂ¦Р В РЎвЂР В Р’Вµ Р В РЎвЂќР В РЎвЂўР РЋР С“Р В РЎВР В РЎвЂР РЋРІР‚РЋР В Р’ВµР РЋР С“Р В РЎвЂќР В РЎвЂўР В РЎвЂ“Р В РЎвЂў Р В РЎвЂўР В Р’В±Р РЋР вЂ°Р В Р’ВµР В РЎвЂќР РЋРІР‚С™Р В Р’В°.
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

// Р В РЎСџР РЋР вЂљР В РЎвЂўР В Р вЂ Р В Р’ВµР РЋР вЂљР РЋР РЏР В Р’ВµР РЋРІР‚С™, Р РЋРІР‚РЋР РЋРІР‚С™Р В РЎвЂў Р РЋР С“Р В РЎвЂўР РЋРІР‚В¦Р РЋР вЂљР В Р’В°Р В Р вЂ¦Р В Р’ВµР В Р вЂ¦Р В РЎвЂР В Р’Вµ Р В РўвЂР В Р’В°Р В Р вЂ¦Р В Р вЂ¦Р РЋРІР‚в„–Р РЋРІР‚В¦ Р В Р’В·Р В Р’В°Р В РЎвЂ”Р В РЎвЂР РЋР С“Р РЋРІР‚в„–Р В Р вЂ Р В Р’В°Р В Р’ВµР РЋРІР‚С™ Р В РЎвЂўР В Р’В±Р В Р вЂ¦Р В РЎвЂўР В Р вЂ Р В Р’В»Р РЋРІР‚ВР В Р вЂ¦Р В Р вЂ¦Р В РЎвЂўР В Р’Вµ Р В РЎвЂ”Р В РЎвЂўР В Р’В»Р В РЎвЂўР В Р’В¶Р В Р’ВµР В Р вЂ¦Р В РЎвЂР В Р’Вµ Р В РЎвЂќР В РЎвЂўР РЋР С“Р В РЎВР В РЎвЂР РЋРІР‚РЋР В Р’ВµР РЋР С“Р В РЎвЂќР В РЎвЂўР В РЎвЂ“Р В РЎвЂў Р В РЎвЂўР В Р’В±Р РЋР вЂ°Р В Р’ВµР В РЎвЂќР РЋРІР‚С™Р В Р’В°.
func TestSendServerChatMessageStoresMessageFromCurrentCharacter(t *testing.T) {
	gameWorld := world.New(1, testWorldData(t))
	if _, ok := gameWorld.ConnectAccount(1); !ok {
		t.Fatalf("account was not connected")
	}
	chatState, ok := gameWorld.ChatStateForAccount(1, 0)
	if !ok {
		t.Fatalf("chat state is not available")
	}

	nextState, recipients, chatError := gameWorld.SendChatMessage(1, chatState.SelectedChatID, "", "Р В РЎСџР РЋР вЂљР В РЎвЂР В Р вЂ Р В Р’ВµР РЋРІР‚С™ Р В Р вЂ Р РЋР С“Р В Р’ВµР В РЎВ")
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
	if message.Text != "Р В РЎСџР РЋР вЂљР В РЎвЂР В Р вЂ Р В Р’ВµР РЋРІР‚С™ Р В Р вЂ Р РЋР С“Р В Р’ВµР В РЎВ" || message.SenderCharacterID != 1 || message.SenderNickname != "index" {
		t.Fatalf("message = %+v, want text from current account", message)
	}
}

// Р В РЎСџР РЋР вЂљР В РЎвЂўР В Р вЂ Р В Р’ВµР РЋР вЂљР РЋР РЏР В Р’ВµР РЋРІР‚С™, Р РЋРІР‚РЋР РЋРІР‚С™Р В РЎвЂў Р РЋР С“Р В РЎвЂўР РЋРІР‚В¦Р РЋР вЂљР В Р’В°Р В Р вЂ¦Р В Р’ВµР В Р вЂ¦Р В РЎвЂР В Р’Вµ Р В РўвЂР В Р’В°Р В Р вЂ¦Р В Р вЂ¦Р РЋРІР‚в„–Р РЋРІР‚В¦ Р В Р’В·Р В Р’В°Р В РЎвЂ”Р В РЎвЂР РЋР С“Р РЋРІР‚в„–Р В Р вЂ Р В Р’В°Р В Р’ВµР РЋРІР‚С™ Р В РЎвЂўР В Р’В±Р В Р вЂ¦Р В РЎвЂўР В Р вЂ Р В Р’В»Р РЋРІР‚ВР В Р вЂ¦Р В Р вЂ¦Р В РЎвЂўР В Р’Вµ Р В РЎвЂ”Р В РЎвЂўР В Р’В»Р В РЎвЂўР В Р’В¶Р В Р’ВµР В Р вЂ¦Р В РЎвЂР В Р’Вµ Р В РЎвЂќР В РЎвЂўР РЋР С“Р В РЎВР В РЎвЂР РЋРІР‚РЋР В Р’ВµР РЋР С“Р В РЎвЂќР В РЎвЂўР В РЎвЂ“Р В РЎвЂў Р В РЎвЂўР В Р’В±Р РЋР вЂ°Р В Р’ВµР В РЎвЂќР РЋРІР‚С™Р В Р’В°.
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

	chatState, recipients, chatError := gameWorld.SendChatMessage(1, 0, "Pilot2", "Р В РІР‚С”Р В РЎвЂР РЋРІР‚РЋР В Р вЂ¦Р В РЎвЂўР В Р’Вµ Р РЋР С“Р В РЎвЂўР В РЎвЂўР В Р’В±Р РЋРІР‚В°Р В Р’ВµР В Р вЂ¦Р В РЎвЂР В Р’Вµ")
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

// Р В РЎСџР РЋР вЂљР В РЎвЂўР В Р вЂ Р В Р’ВµР РЋР вЂљР РЋР РЏР В Р’ВµР РЋРІР‚С™, Р РЋРІР‚РЋР РЋРІР‚С™Р В РЎвЂў Р РЋР С“Р В РЎвЂўР РЋРІР‚В¦Р РЋР вЂљР В Р’В°Р В Р вЂ¦Р В Р’ВµР В Р вЂ¦Р В РЎвЂР В Р’Вµ Р В РўвЂР В Р’В°Р В Р вЂ¦Р В Р вЂ¦Р РЋРІР‚в„–Р РЋРІР‚В¦ Р В Р’В·Р В Р’В°Р В РЎвЂ”Р В РЎвЂР РЋР С“Р РЋРІР‚в„–Р В Р вЂ Р В Р’В°Р В Р’ВµР РЋРІР‚С™ Р В РЎвЂўР В Р’В±Р В Р вЂ¦Р В РЎвЂўР В Р вЂ Р В Р’В»Р РЋРІР‚ВР В Р вЂ¦Р В Р вЂ¦Р В РЎвЂўР В Р’Вµ Р В РЎвЂ”Р В РЎвЂўР В Р’В»Р В РЎвЂўР В Р’В¶Р В Р’ВµР В Р вЂ¦Р В РЎвЂР В Р’Вµ Р В РЎвЂќР В РЎвЂўР РЋР С“Р В РЎВР В РЎвЂР РЋРІР‚РЋР В Р’ВµР РЋР С“Р В РЎвЂќР В РЎвЂўР В РЎвЂ“Р В РЎвЂў Р В РЎвЂўР В Р’В±Р РЋР вЂ°Р В Р’ВµР В РЎвЂќР РЋРІР‚С™Р В Р’В°.
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

	if _, _, chatError := gameWorld.SendChatMessage(1, 0, "Pilot2", "Р В РІР‚С”Р В РЎвЂР РЋРІР‚РЋР В Р вЂ¦Р В РЎвЂўР В Р’Вµ Р РЋР С“Р В РЎвЂўР В РЎвЂўР В Р’В±Р РЋРІР‚В°Р В Р’ВµР В Р вЂ¦Р В РЎвЂР В Р’Вµ"); chatError != "" {
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

// Р В РЎСџР РЋР вЂљР В РЎвЂўР В Р вЂ Р В Р’ВµР РЋР вЂљР РЋР РЏР В Р’ВµР РЋРІР‚С™, Р РЋРІР‚РЋР РЋРІР‚С™Р В РЎвЂў Р РЋР С“Р В РЎвЂўР РЋРІР‚В¦Р РЋР вЂљР В Р’В°Р В Р вЂ¦Р В Р’ВµР В Р вЂ¦Р В РЎвЂР В Р’Вµ Р В РўвЂР В Р’В°Р В Р вЂ¦Р В Р вЂ¦Р РЋРІР‚в„–Р РЋРІР‚В¦ Р В Р’В·Р В Р’В°Р В РЎвЂ”Р В РЎвЂР РЋР С“Р РЋРІР‚в„–Р В Р вЂ Р В Р’В°Р В Р’ВµР РЋРІР‚С™ Р В РЎвЂўР В Р’В±Р В Р вЂ¦Р В РЎвЂўР В Р вЂ Р В Р’В»Р РЋРІР‚ВР В Р вЂ¦Р В Р вЂ¦Р В РЎвЂўР В Р’Вµ Р В РЎвЂ”Р В РЎвЂўР В Р’В»Р В РЎвЂўР В Р’В¶Р В Р’ВµР В Р вЂ¦Р В РЎвЂР В Р’Вµ Р В РЎвЂќР В РЎвЂўР РЋР С“Р В РЎВР В РЎвЂР РЋРІР‚РЋР В Р’ВµР РЋР С“Р В РЎвЂќР В РЎвЂўР В РЎвЂ“Р В РЎвЂў Р В РЎвЂўР В Р’В±Р РЋР вЂ°Р В Р’ВµР В РЎвЂќР РЋРІР‚С™Р В Р’В°.
func TestSendDuoChatMessageRejectsUnknownAccountNickname(t *testing.T) {
	gameWorld := world.New(1, testWorldData(t))
	if _, ok := gameWorld.ConnectAccount(1); !ok {
		t.Fatalf("account was not connected")
	}

	_, recipients, chatError := gameWorld.SendChatMessage(1, 0, "Nobody", "Р В РІР‚С”Р В РЎвЂР РЋРІР‚РЋР В Р вЂ¦Р В РЎвЂўР В Р’Вµ Р РЋР С“Р В РЎвЂўР В РЎвЂўР В Р’В±Р РЋРІР‚В°Р В Р’ВµР В Р вЂ¦Р В РЎвЂР В Р’Вµ")
	if chatError == "" {
		t.Fatal("unknown nickname was accepted")
	}
	if len(recipients) != 0 {
		t.Fatalf("recipients = %v, want empty list", recipients)
	}
}

// Р В РЎСџР РЋР вЂљР В РЎвЂўР В Р вЂ Р В Р’ВµР РЋР вЂљР РЋР РЏР В Р’ВµР РЋРІР‚С™, Р РЋРІР‚РЋР РЋРІР‚С™Р В РЎвЂў Р РЋР С“Р В РЎвЂўР РЋРІР‚В¦Р РЋР вЂљР В Р’В°Р В Р вЂ¦Р В Р’ВµР В Р вЂ¦Р В РЎвЂР В Р’Вµ Р В РўвЂР В Р’В°Р В Р вЂ¦Р В Р вЂ¦Р РЋРІР‚в„–Р РЋРІР‚В¦ Р В Р’В·Р В Р’В°Р В РЎвЂ”Р В РЎвЂР РЋР С“Р РЋРІР‚в„–Р В Р вЂ Р В Р’В°Р В Р’ВµР РЋРІР‚С™ Р В РЎвЂўР В Р’В±Р В Р вЂ¦Р В РЎвЂўР В Р вЂ Р В Р’В»Р РЋРІР‚ВР В Р вЂ¦Р В Р вЂ¦Р В РЎвЂўР В Р’Вµ Р В РЎвЂ”Р В РЎвЂўР В Р’В»Р В РЎвЂўР В Р’В¶Р В Р’ВµР В Р вЂ¦Р В РЎвЂР В Р’Вµ Р В РЎвЂќР В РЎвЂўР РЋР С“Р В РЎВР В РЎвЂР РЋРІР‚РЋР В Р’ВµР РЋР С“Р В РЎвЂќР В РЎвЂўР В РЎвЂ“Р В РЎвЂў Р В РЎвЂўР В Р’В±Р РЋР вЂ°Р В Р’ВµР В РЎвЂќР РЋРІР‚С™Р В Р’В°.
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

// Р В РЎСџР РЋР вЂљР В РЎвЂўР В Р вЂ Р В Р’ВµР РЋР вЂљР РЋР РЏР В Р’ВµР РЋРІР‚С™, Р РЋРІР‚РЋР РЋРІР‚С™Р В РЎвЂў Р РЋР С“Р В РЎвЂўР РЋРІР‚В¦Р РЋР вЂљР В Р’В°Р В Р вЂ¦Р В Р’ВµР В Р вЂ¦Р В РЎвЂР В Р’Вµ Р В РўвЂР В Р’В°Р В Р вЂ¦Р В Р вЂ¦Р РЋРІР‚в„–Р РЋРІР‚В¦ Р В Р’В·Р В Р’В°Р В РЎвЂ”Р В РЎвЂР РЋР С“Р РЋРІР‚в„–Р В Р вЂ Р В Р’В°Р В Р’ВµР РЋРІР‚С™ Р В РЎвЂўР В Р’В±Р В Р вЂ¦Р В РЎвЂўР В Р вЂ Р В Р’В»Р РЋРІР‚ВР В Р вЂ¦Р В Р вЂ¦Р В РЎвЂўР В Р’Вµ Р В РЎвЂ”Р В РЎвЂўР В Р’В»Р В РЎвЂўР В Р’В¶Р В Р’ВµР В Р вЂ¦Р В РЎвЂР В Р’Вµ Р В РЎвЂќР В РЎвЂўР РЋР С“Р В РЎВР В РЎвЂР РЋРІР‚РЋР В Р’ВµР РЋР С“Р В РЎвЂќР В РЎвЂўР В РЎвЂ“Р В РЎвЂў Р В РЎвЂўР В Р’В±Р РЋР вЂ°Р В Р’ВµР В РЎвЂќР РЋРІР‚С™Р В Р’В°.
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

// Р В РЎСџР РЋР вЂљР В РЎвЂўР В Р вЂ Р В Р’ВµР РЋР вЂљР РЋР РЏР В Р’ВµР РЋРІР‚С™, Р РЋРІР‚РЋР РЋРІР‚С™Р В РЎвЂў Р РЋР С“Р В РЎвЂўР РЋРІР‚В¦Р РЋР вЂљР В Р’В°Р В Р вЂ¦Р В Р’ВµР В Р вЂ¦Р В РЎвЂР В Р’Вµ Р В РўвЂР В Р’В°Р В Р вЂ¦Р В Р вЂ¦Р РЋРІР‚в„–Р РЋРІР‚В¦ Р В Р’В·Р В Р’В°Р В РЎвЂ”Р В РЎвЂР РЋР С“Р РЋРІР‚в„–Р В Р вЂ Р В Р’В°Р В Р’ВµР РЋРІР‚С™ Р В РЎвЂўР В Р’В±Р В Р вЂ¦Р В РЎвЂўР В Р вЂ Р В Р’В»Р РЋРІР‚ВР В Р вЂ¦Р В Р вЂ¦Р В РЎвЂўР В Р’Вµ Р В РЎвЂ”Р В РЎвЂўР В Р’В»Р В РЎвЂўР В Р’В¶Р В Р’ВµР В Р вЂ¦Р В РЎвЂР В Р’Вµ Р В РЎвЂќР В РЎвЂўР РЋР С“Р В РЎВР В РЎвЂР РЋРІР‚РЋР В Р’ВµР РЋР С“Р В РЎвЂќР В РЎвЂўР В РЎвЂ“Р В РЎвЂў Р В РЎвЂўР В Р’В±Р РЋР вЂ°Р В Р’ВµР В РЎвЂќР РЋРІР‚С™Р В Р’В°.
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

// Р В РЎСџР РЋР вЂљР В РЎвЂўР В Р вЂ Р В Р’ВµР РЋР вЂљР РЋР РЏР В Р’ВµР РЋРІР‚С™, Р РЋРІР‚РЋР РЋРІР‚С™Р В РЎвЂў Р РЋР С“Р В РЎвЂўР РЋРІР‚В¦Р РЋР вЂљР В Р’В°Р В Р вЂ¦Р В Р’ВµР В Р вЂ¦Р В РЎвЂР В Р’Вµ Р В РўвЂР В Р’В°Р В Р вЂ¦Р В Р вЂ¦Р РЋРІР‚в„–Р РЋРІР‚В¦ Р В Р’В·Р В Р’В°Р В РЎвЂ”Р В РЎвЂР РЋР С“Р РЋРІР‚в„–Р В Р вЂ Р В Р’В°Р В Р’ВµР РЋРІР‚С™ Р В РЎвЂўР В Р’В±Р В Р вЂ¦Р В РЎвЂўР В Р вЂ Р В Р’В»Р РЋРІР‚ВР В Р вЂ¦Р В Р вЂ¦Р В РЎвЂўР В Р’Вµ Р В РЎвЂ”Р В РЎвЂўР В Р’В»Р В РЎвЂўР В Р’В¶Р В Р’ВµР В Р вЂ¦Р В РЎвЂР В Р’Вµ Р В РЎвЂќР В РЎвЂўР РЋР С“Р В РЎВР В РЎвЂР РЋРІР‚РЋР В Р’ВµР РЋР С“Р В РЎвЂќР В РЎвЂўР В РЎвЂ“Р В РЎвЂў Р В РЎвЂўР В Р’В±Р РЋР вЂ°Р В Р’ВµР В РЎвЂќР РЋРІР‚С™Р В Р’В°.
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

// Р В РЎСџР РЋР вЂљР В РЎвЂўР В Р вЂ Р В Р’ВµР РЋР вЂљР РЋР РЏР В Р’ВµР РЋРІР‚С™, Р РЋРІР‚РЋР РЋРІР‚С™Р В РЎвЂў Р РЋР С“Р В РЎвЂўР РЋРІР‚В¦Р РЋР вЂљР В Р’В°Р В Р вЂ¦Р В Р’ВµР В Р вЂ¦Р В РЎвЂР В Р’Вµ Р В РўвЂР В Р’В°Р В Р вЂ¦Р В Р вЂ¦Р РЋРІР‚в„–Р РЋРІР‚В¦ Р В Р’В·Р В Р’В°Р В РЎвЂ”Р В РЎвЂР РЋР С“Р РЋРІР‚в„–Р В Р вЂ Р В Р’В°Р В Р’ВµР РЋРІР‚С™ Р В РЎвЂўР В Р’В±Р В Р вЂ¦Р В РЎвЂўР В Р вЂ Р В Р’В»Р РЋРІР‚ВР В Р вЂ¦Р В Р вЂ¦Р В РЎвЂўР В Р’Вµ Р В РЎвЂ”Р В РЎвЂўР В Р’В»Р В РЎвЂўР В Р’В¶Р В Р’ВµР В Р вЂ¦Р В РЎвЂР В Р’Вµ Р В РЎвЂќР В РЎвЂўР РЋР С“Р В РЎВР В РЎвЂР РЋРІР‚РЋР В Р’ВµР РЋР С“Р В РЎвЂќР В РЎвЂўР В РЎвЂ“Р В РЎвЂў Р В РЎвЂўР В Р’В±Р РЋР вЂ°Р В Р’ВµР В РЎвЂќР РЋРІР‚С™Р В Р’В°.
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

// Р В РЎСџР РЋР вЂљР В РЎвЂўР В Р вЂ Р В Р’ВµР РЋР вЂљР РЋР РЏР В Р’ВµР РЋРІР‚С™, Р РЋРІР‚РЋР РЋРІР‚С™Р В РЎвЂў Р РЋР С“Р В РЎвЂўР РЋРІР‚В¦Р РЋР вЂљР В Р’В°Р В Р вЂ¦Р В Р’ВµР В Р вЂ¦Р В РЎвЂР В Р’Вµ Р В РўвЂР В Р’В°Р В Р вЂ¦Р В Р вЂ¦Р РЋРІР‚в„–Р РЋРІР‚В¦ Р В Р’В·Р В Р’В°Р В РЎвЂ”Р В РЎвЂР РЋР С“Р РЋРІР‚в„–Р В Р вЂ Р В Р’В°Р В Р’ВµР РЋРІР‚С™ Р В РЎвЂўР В Р’В±Р В Р вЂ¦Р В РЎвЂўР В Р вЂ Р В Р’В»Р РЋРІР‚ВР В Р вЂ¦Р В Р вЂ¦Р В РЎвЂўР В Р’Вµ Р В РЎвЂ”Р В РЎвЂўР В Р’В»Р В РЎвЂўР В Р’В¶Р В Р’ВµР В Р вЂ¦Р В РЎвЂР В Р’Вµ Р В РЎвЂќР В РЎвЂўР РЋР С“Р В РЎВР В РЎвЂР РЋРІР‚РЋР В Р’ВµР РЋР С“Р В РЎвЂќР В РЎвЂўР В РЎвЂ“Р В РЎвЂў Р В РЎвЂўР В Р’В±Р РЋР вЂ°Р В Р’ВµР В РЎвЂќР РЋРІР‚С™Р В Р’В°.
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

// Р В РЎСџР РЋР вЂљР В РЎвЂўР В Р вЂ Р В Р’ВµР РЋР вЂљР РЋР РЏР В Р’ВµР РЋРІР‚С™, Р РЋРІР‚РЋР РЋРІР‚С™Р В РЎвЂў Р РЋР С“Р В РЎвЂўР РЋРІР‚В¦Р РЋР вЂљР В Р’В°Р В Р вЂ¦Р В Р’ВµР В Р вЂ¦Р В РЎвЂР В Р’Вµ Р В РўвЂР В Р’В°Р В Р вЂ¦Р В Р вЂ¦Р РЋРІР‚в„–Р РЋРІР‚В¦ Р В Р’В·Р В Р’В°Р В РЎвЂ”Р В РЎвЂР РЋР С“Р РЋРІР‚в„–Р В Р вЂ Р В Р’В°Р В Р’ВµР РЋРІР‚С™ Р В РЎвЂўР В Р’В±Р В Р вЂ¦Р В РЎвЂўР В Р вЂ Р В Р’В»Р РЋРІР‚ВР В Р вЂ¦Р В Р вЂ¦Р В РЎвЂўР В Р’Вµ Р В РЎвЂ”Р В РЎвЂўР В Р’В»Р В РЎвЂўР В Р’В¶Р В Р’ВµР В Р вЂ¦Р В РЎвЂР В Р’Вµ Р В РЎвЂќР В РЎвЂўР РЋР С“Р В РЎВР В РЎвЂР РЋРІР‚РЋР В Р’ВµР РЋР С“Р В РЎвЂќР В РЎвЂўР В РЎвЂ“Р В РЎвЂў Р В РЎвЂўР В Р’В±Р РЋР вЂ°Р В Р’ВµР В РЎвЂќР РЋРІР‚С™Р В Р’В°.
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
	gameWorld.Tick(1)

	cosmicObject := serverData.CosmicObjects.Items[objectID]
	if cosmicObject.ConsumingPower != 30000 {
		t.Fatalf("consuming power = %v, want 30000", cosmicObject.ConsumingPower)
	}
	if cosmicObject.GeneratingPower != 60000 {
		t.Fatalf("generating power = %v, want 60000", cosmicObject.GeneratingPower)
	}
	closeWorldFloat(t, cosmicObject.Fuel, 49)

	installed := serverData.EquipmentGroups.GetByCosmicObjectID(objectID)
	if !installed[0].Active {
		t.Fatalf("thrusters must be active while creating thrust: %+v", installed[0])
	}
	if installed[1].Active {
		t.Fatalf("torquer must not be active without target rotation delta: %+v", installed[1])
	}
}

// Р В РЎСџР РЋР вЂљР В РЎвЂўР В Р вЂ Р В Р’ВµР РЋР вЂљР РЋР РЏР В Р’ВµР РЋРІР‚С™, Р РЋРІР‚РЋР РЋРІР‚С™Р В РЎвЂў Р РЋР С“Р В РЎвЂўР РЋРІР‚В¦Р РЋР вЂљР В Р’В°Р В Р вЂ¦Р В Р’ВµР В Р вЂ¦Р В РЎвЂР В Р’Вµ Р В РўвЂР В Р’В°Р В Р вЂ¦Р В Р вЂ¦Р РЋРІР‚в„–Р РЋРІР‚В¦ Р В Р’В·Р В Р’В°Р В РЎвЂ”Р В РЎвЂР РЋР С“Р РЋРІР‚в„–Р В Р вЂ Р В Р’В°Р В Р’ВµР РЋРІР‚С™ Р В РЎвЂўР В Р’В±Р В Р вЂ¦Р В РЎвЂўР В Р вЂ Р В Р’В»Р РЋРІР‚ВР В Р вЂ¦Р В Р вЂ¦Р В РЎвЂўР В Р’Вµ Р В РЎвЂ”Р В РЎвЂўР В Р’В»Р В РЎвЂўР В Р’В¶Р В Р’ВµР В Р вЂ¦Р В РЎвЂР В Р’Вµ Р В РЎвЂќР В РЎвЂўР РЋР С“Р В РЎВР В РЎвЂР РЋРІР‚РЋР В Р’ВµР РЋР С“Р В РЎвЂќР В РЎвЂўР В РЎвЂ“Р В РЎвЂў Р В РЎвЂўР В Р’В±Р РЋР вЂ°Р В Р’ВµР В РЎвЂќР РЋРІР‚С™Р В Р’В°.
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

// Р В РЎСџР РЋР вЂљР В РЎвЂўР В Р вЂ Р В Р’ВµР РЋР вЂљР РЋР РЏР В Р’ВµР РЋРІР‚С™, Р РЋРІР‚РЋР РЋРІР‚С™Р В РЎвЂў Р РЋР С“Р В РЎвЂўР РЋРІР‚В¦Р РЋР вЂљР В Р’В°Р В Р вЂ¦Р В Р’ВµР В Р вЂ¦Р В РЎвЂР В Р’Вµ Р В РўвЂР В Р’В°Р В Р вЂ¦Р В Р вЂ¦Р РЋРІР‚в„–Р РЋРІР‚В¦ Р В Р’В·Р В Р’В°Р В РЎвЂ”Р В РЎвЂР РЋР С“Р РЋРІР‚в„–Р В Р вЂ Р В Р’В°Р В Р’ВµР РЋРІР‚С™ Р В РЎвЂўР В Р’В±Р В Р вЂ¦Р В РЎвЂўР В Р вЂ Р В Р’В»Р РЋРІР‚ВР В Р вЂ¦Р В Р вЂ¦Р В РЎвЂўР В Р’Вµ Р В РЎвЂ”Р В РЎвЂўР В Р’В»Р В РЎвЂўР В Р’В¶Р В Р’ВµР В Р вЂ¦Р В РЎвЂР В Р’Вµ Р В РЎвЂќР В РЎвЂўР РЋР С“Р В РЎВР В РЎвЂР РЋРІР‚РЋР В Р’ВµР РЋР С“Р В РЎвЂќР В РЎвЂўР В РЎвЂ“Р В РЎвЂў Р В РЎвЂўР В Р’В±Р РЋР вЂ°Р В Р’ВµР В РЎвЂќР РЋРІР‚С™Р В Р’В°.
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

// Р В РЎСџР РЋР вЂљР В РЎвЂўР В Р вЂ Р В Р’ВµР РЋР вЂљР РЋР РЏР В Р’ВµР РЋРІР‚С™, Р РЋРІР‚РЋР РЋРІР‚С™Р В РЎвЂў Р РЋР С“Р В РЎвЂўР РЋРІР‚В¦Р РЋР вЂљР В Р’В°Р В Р вЂ¦Р В Р’ВµР В Р вЂ¦Р В РЎвЂР В Р’Вµ Р В РўвЂР В Р’В°Р В Р вЂ¦Р В Р вЂ¦Р РЋРІР‚в„–Р РЋРІР‚В¦ Р В Р’В·Р В Р’В°Р В РЎвЂ”Р В РЎвЂР РЋР С“Р РЋРІР‚в„–Р В Р вЂ Р В Р’В°Р В Р’ВµР РЋРІР‚С™ Р В РЎвЂўР В Р’В±Р В Р вЂ¦Р В РЎвЂўР В Р вЂ Р В Р’В»Р РЋРІР‚ВР В Р вЂ¦Р В Р вЂ¦Р В РЎвЂўР В Р’Вµ Р В РЎвЂ”Р В РЎвЂўР В Р’В»Р В РЎвЂўР В Р’В¶Р В Р’ВµР В Р вЂ¦Р В РЎвЂР В Р’Вµ Р В РЎвЂќР В РЎвЂўР РЋР С“Р В РЎВР В РЎвЂР РЋРІР‚РЋР В Р’ВµР РЋР С“Р В РЎвЂќР В РЎвЂўР В РЎвЂ“Р В РЎвЂў Р В РЎвЂўР В Р’В±Р РЋР вЂ°Р В Р’ВµР В РЎвЂќР РЋРІР‚С™Р В Р’В°.
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

// Р В РЎСџР РЋР вЂљР В РЎвЂўР В Р вЂ Р В Р’ВµР РЋР вЂљР РЋР РЏР В Р’ВµР РЋРІР‚С™, Р РЋРІР‚РЋР РЋРІР‚С™Р В РЎвЂў Р РЋР С“Р В РЎвЂўР РЋРІР‚В¦Р РЋР вЂљР В Р’В°Р В Р вЂ¦Р В Р’ВµР В Р вЂ¦Р В РЎвЂР В Р’Вµ Р В РўвЂР В Р’В°Р В Р вЂ¦Р В Р вЂ¦Р РЋРІР‚в„–Р РЋРІР‚В¦ Р В Р’В·Р В Р’В°Р В РЎвЂ”Р В РЎвЂР РЋР С“Р РЋРІР‚в„–Р В Р вЂ Р В Р’В°Р В Р’ВµР РЋРІР‚С™ Р В РЎвЂўР В Р’В±Р В Р вЂ¦Р В РЎвЂўР В Р вЂ Р В Р’В»Р РЋРІР‚ВР В Р вЂ¦Р В Р вЂ¦Р В РЎвЂўР В Р’Вµ Р В РЎвЂ”Р В РЎвЂўР В Р’В»Р В РЎвЂўР В Р’В¶Р В Р’ВµР В Р вЂ¦Р В РЎвЂР В Р’Вµ Р В РЎвЂќР В РЎвЂўР РЋР С“Р В РЎВР В РЎвЂР РЋРІР‚РЋР В Р’ВµР РЋР С“Р В РЎвЂќР В РЎвЂўР В РЎвЂ“Р В РЎвЂў Р В РЎвЂўР В Р’В±Р РЋР вЂ°Р В Р’ВµР В РЎвЂќР РЋРІР‚С™Р В Р’В°.
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

// Р В РЎСџР РЋР вЂљР В РЎвЂўР В Р вЂ Р В Р’ВµР РЋР вЂљР РЋР РЏР В Р’ВµР РЋРІР‚С™, Р РЋРІР‚РЋР РЋРІР‚С™Р В РЎвЂў Р РЋР С“Р В РЎвЂўР РЋРІР‚В¦Р РЋР вЂљР В Р’В°Р В Р вЂ¦Р В Р’ВµР В Р вЂ¦Р В РЎвЂР В Р’Вµ Р В РўвЂР В Р’В°Р В Р вЂ¦Р В Р вЂ¦Р РЋРІР‚в„–Р РЋРІР‚В¦ Р В Р’В·Р В Р’В°Р В РЎвЂ”Р В РЎвЂР РЋР С“Р РЋРІР‚в„–Р В Р вЂ Р В Р’В°Р В Р’ВµР РЋРІР‚С™ Р В РЎвЂўР В Р’В±Р В Р вЂ¦Р В РЎвЂўР В Р вЂ Р В Р’В»Р РЋРІР‚ВР В Р вЂ¦Р В Р вЂ¦Р В РЎвЂўР В Р’Вµ Р В РЎвЂ”Р В РЎвЂўР В Р’В»Р В РЎвЂўР В Р’В¶Р В Р’ВµР В Р вЂ¦Р В РЎвЂР В Р’Вµ Р В РЎвЂќР В РЎвЂўР РЋР С“Р В РЎВР В РЎвЂР РЋРІР‚РЋР В Р’ВµР РЋР С“Р В РЎвЂќР В РЎвЂўР В РЎвЂ“Р В РЎвЂў Р В РЎвЂўР В Р’В±Р РЋР вЂ°Р В Р’ВµР В РЎвЂќР РЋРІР‚С™Р В Р’В°.
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
	gameWorld.Tick(2)

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

// Р В РЎСџР РЋР вЂљР В РЎвЂўР В Р вЂ Р В Р’ВµР РЋР вЂљР РЋР РЏР В Р’ВµР РЋРІР‚С™, Р РЋРІР‚РЋР РЋРІР‚С™Р В РЎвЂў Р РЋР С“Р В РЎвЂўР РЋРІР‚В¦Р РЋР вЂљР В Р’В°Р В Р вЂ¦Р В Р’ВµР В Р вЂ¦Р В РЎвЂР В Р’Вµ Р В РўвЂР В Р’В°Р В Р вЂ¦Р В Р вЂ¦Р РЋРІР‚в„–Р РЋРІР‚В¦ Р В Р’В·Р В Р’В°Р В РЎвЂ”Р В РЎвЂР РЋР С“Р РЋРІР‚в„–Р В Р вЂ Р В Р’В°Р В Р’ВµР РЋРІР‚С™ Р В РЎвЂўР В Р’В±Р В Р вЂ¦Р В РЎвЂўР В Р вЂ Р В Р’В»Р РЋРІР‚ВР В Р вЂ¦Р В Р вЂ¦Р В РЎвЂўР В Р’Вµ Р В РЎвЂ”Р В РЎвЂўР В Р’В»Р В РЎвЂўР В Р’В¶Р В Р’ВµР В Р вЂ¦Р В РЎвЂР В Р’Вµ Р В РЎвЂќР В РЎвЂўР РЋР С“Р В РЎВР В РЎвЂР РЋРІР‚РЋР В Р’ВµР РЋР С“Р В РЎвЂќР В РЎвЂўР В РЎвЂ“Р В РЎвЂў Р В РЎвЂўР В Р’В±Р РЋР вЂ°Р В Р’ВµР В РЎвЂќР РЋРІР‚С™Р В Р’В°.
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

// Р В РЎСџР РЋР вЂљР В РЎвЂўР В Р вЂ Р В Р’ВµР РЋР вЂљР РЋР РЏР В Р’ВµР РЋРІР‚С™, Р РЋРІР‚РЋР РЋРІР‚С™Р В РЎвЂў Р РЋР С“Р В РЎвЂўР РЋРІР‚В¦Р РЋР вЂљР В Р’В°Р В Р вЂ¦Р В Р’ВµР В Р вЂ¦Р В РЎвЂР В Р’Вµ Р В РўвЂР В Р’В°Р В Р вЂ¦Р В Р вЂ¦Р РЋРІР‚в„–Р РЋРІР‚В¦ Р В Р’В·Р В Р’В°Р В РЎвЂ”Р В РЎвЂР РЋР С“Р РЋРІР‚в„–Р В Р вЂ Р В Р’В°Р В Р’ВµР РЋРІР‚С™ Р В РЎвЂўР В Р’В±Р В Р вЂ¦Р В РЎвЂўР В Р вЂ Р В Р’В»Р РЋРІР‚ВР В Р вЂ¦Р В Р вЂ¦Р В РЎвЂўР В Р’Вµ Р В РЎвЂ”Р В РЎвЂўР В Р’В»Р В РЎвЂўР В Р’В¶Р В Р’ВµР В Р вЂ¦Р В РЎвЂР В Р’Вµ Р В РЎвЂќР В РЎвЂўР РЋР С“Р В РЎВР В РЎвЂР РЋРІР‚РЋР В Р’ВµР РЋР С“Р В РЎвЂќР В РЎвЂўР В РЎвЂ“Р В РЎвЂў Р В РЎвЂўР В Р’В±Р РЋР вЂ°Р В Р’ВµР В РЎвЂќР РЋРІР‚С™Р В Р’В°.
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
	if _, collided := physics.CollisionCorrection(object.CosmicObject, *objectModel, *obstacle, *obstacleModel); collided {
		t.Fatalf("controlled object still intersects anchored object: %+v", object)
	}
}

// Р В РЎСџР РЋР вЂљР В РЎвЂўР В Р вЂ Р В Р’ВµР РЋР вЂљР РЋР РЏР В Р’ВµР РЋРІР‚С™, Р РЋРІР‚РЋР РЋРІР‚С™Р В РЎвЂў Р РЋР С“Р В РЎвЂўР РЋРІР‚В¦Р РЋР вЂљР В Р’В°Р В Р вЂ¦Р В Р’ВµР В Р вЂ¦Р В РЎвЂР В Р’Вµ Р В РўвЂР В Р’В°Р В Р вЂ¦Р В Р вЂ¦Р РЋРІР‚в„–Р РЋРІР‚В¦ Р В Р’В·Р В Р’В°Р В РЎвЂ”Р В РЎвЂР РЋР С“Р РЋРІР‚в„–Р В Р вЂ Р В Р’В°Р В Р’ВµР РЋРІР‚С™ Р В РЎвЂўР В Р’В±Р В Р вЂ¦Р В РЎвЂўР В Р вЂ Р В Р’В»Р РЋРІР‚ВР В Р вЂ¦Р В Р вЂ¦Р В РЎвЂўР В Р’Вµ Р В РЎвЂ”Р В РЎвЂўР В Р’В»Р В РЎвЂўР В Р’В¶Р В Р’ВµР В Р вЂ¦Р В РЎвЂР В Р’Вµ Р В РЎвЂќР В РЎвЂўР РЋР С“Р В РЎВР В РЎвЂР РЋРІР‚РЋР В Р’ВµР РЋР С“Р В РЎвЂќР В РЎвЂўР В РЎвЂ“Р В РЎвЂў Р В РЎвЂўР В Р’В±Р РЋР вЂ°Р В Р’ВµР В РЎвЂќР РЋРІР‚С™Р В Р’В°.
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

// Р В РЎСџР РЋР вЂљР В РЎвЂўР В Р вЂ Р В Р’ВµР РЋР вЂљР РЋР РЏР В Р’ВµР РЋРІР‚С™, Р РЋРІР‚РЋР РЋРІР‚С™Р В РЎвЂў Р РЋР С“Р В РЎвЂўР РЋРІР‚В¦Р РЋР вЂљР В Р’В°Р В Р вЂ¦Р В Р’ВµР В Р вЂ¦Р В РЎвЂР В Р’Вµ Р В РўвЂР В Р’В°Р В Р вЂ¦Р В Р вЂ¦Р РЋРІР‚в„–Р РЋРІР‚В¦ Р В Р’В·Р В Р’В°Р В РЎвЂ”Р В РЎвЂР РЋР С“Р РЋРІР‚в„–Р В Р вЂ Р В Р’В°Р В Р’ВµР РЋРІР‚С™ Р В РЎвЂўР В Р’В±Р В Р вЂ¦Р В РЎвЂўР В Р вЂ Р В Р’В»Р РЋРІР‚ВР В Р вЂ¦Р В Р вЂ¦Р В РЎвЂўР В Р’Вµ Р В РЎвЂ”Р В РЎвЂўР В Р’В»Р В РЎвЂўР В Р’В¶Р В Р’ВµР В Р вЂ¦Р В РЎвЂР В Р’Вµ Р В РЎвЂќР В РЎвЂўР РЋР С“Р В РЎВР В РЎвЂР РЋРІР‚РЋР В Р’ВµР РЋР С“Р В РЎвЂќР В РЎвЂўР В РЎвЂ“Р В РЎвЂў Р В РЎвЂўР В Р’В±Р РЋР вЂ°Р В Р’ВµР В РЎвЂќР РЋРІР‚С™Р В Р’В°.
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

// Р В РЎСџР РЋР вЂљР В РЎвЂўР В Р вЂ Р В Р’ВµР РЋР вЂљР РЋР РЏР В Р’ВµР РЋРІР‚С™, Р РЋРІР‚РЋР РЋРІР‚С™Р В РЎвЂў Р РЋР С“Р В РЎвЂўР РЋРІР‚В¦Р РЋР вЂљР В Р’В°Р В Р вЂ¦Р В Р’ВµР В Р вЂ¦Р В РЎвЂР В Р’Вµ Р В РўвЂР В Р’В°Р В Р вЂ¦Р В Р вЂ¦Р РЋРІР‚в„–Р РЋРІР‚В¦ Р В Р’В·Р В Р’В°Р В РЎвЂ”Р В РЎвЂР РЋР С“Р РЋРІР‚в„–Р В Р вЂ Р В Р’В°Р В Р’ВµР РЋРІР‚С™ Р В РЎвЂўР В Р’В±Р В Р вЂ¦Р В РЎвЂўР В Р вЂ Р В Р’В»Р РЋРІР‚ВР В Р вЂ¦Р В Р вЂ¦Р В РЎвЂўР В Р’Вµ Р В РЎвЂ”Р В РЎвЂўР В Р’В»Р В РЎвЂўР В Р’В¶Р В Р’ВµР В Р вЂ¦Р В РЎвЂР В Р’Вµ Р В РЎвЂќР В РЎвЂўР РЋР С“Р В РЎВР В РЎвЂР РЋРІР‚РЋР В Р’ВµР РЋР С“Р В РЎвЂќР В РЎвЂўР В РЎвЂ“Р В РЎвЂў Р В РЎвЂўР В Р’В±Р РЋР вЂ°Р В Р’ВµР В РЎвЂќР РЋРІР‚С™Р В Р’В°.
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

// Р В РЎСџР РЋР вЂљР В РЎвЂўР В Р вЂ Р В Р’ВµР РЋР вЂљР РЋР РЏР В Р’ВµР РЋРІР‚С™, Р РЋРІР‚РЋР РЋРІР‚С™Р В РЎвЂў Р РЋР С“Р В РЎвЂўР РЋРІР‚В¦Р РЋР вЂљР В Р’В°Р В Р вЂ¦Р В Р’ВµР В Р вЂ¦Р В РЎвЂР В Р’Вµ Р В РўвЂР В Р’В°Р В Р вЂ¦Р В Р вЂ¦Р РЋРІР‚в„–Р РЋРІР‚В¦ Р В Р’В·Р В Р’В°Р В РЎвЂ”Р В РЎвЂР РЋР С“Р РЋРІР‚в„–Р В Р вЂ Р В Р’В°Р В Р’ВµР РЋРІР‚С™ Р В РЎвЂўР В Р’В±Р В Р вЂ¦Р В РЎвЂўР В Р вЂ Р В Р’В»Р РЋРІР‚ВР В Р вЂ¦Р В Р вЂ¦Р В РЎвЂўР В Р’Вµ Р В РЎвЂ”Р В РЎвЂўР В Р’В»Р В РЎвЂўР В Р’В¶Р В Р’ВµР В Р вЂ¦Р В РЎвЂР В Р’Вµ Р В РЎвЂќР В РЎвЂўР РЋР С“Р В РЎВР В РЎвЂР РЋРІР‚РЋР В Р’ВµР РЋР С“Р В РЎвЂќР В РЎвЂўР В РЎвЂ“Р В РЎвЂў Р В РЎвЂўР В Р’В±Р РЋР вЂ°Р В Р’ВµР В РЎвЂќР РЋРІР‚С™Р В Р’В°.
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

// Р В РЎСџР РЋР вЂљР В РЎвЂўР В Р вЂ Р В Р’ВµР РЋР вЂљР РЋР РЏР В Р’ВµР РЋРІР‚С™, Р РЋРІР‚РЋР РЋРІР‚С™Р В РЎвЂў Р РЋР С“Р В РЎвЂўР РЋРІР‚В¦Р РЋР вЂљР В Р’В°Р В Р вЂ¦Р В Р’ВµР В Р вЂ¦Р В РЎвЂР В Р’Вµ Р В РўвЂР В Р’В°Р В Р вЂ¦Р В Р вЂ¦Р РЋРІР‚в„–Р РЋРІР‚В¦ Р В Р’В·Р В Р’В°Р В РЎвЂ”Р В РЎвЂР РЋР С“Р РЋРІР‚в„–Р В Р вЂ Р В Р’В°Р В Р’ВµР РЋРІР‚С™ Р В РЎвЂўР В Р’В±Р В Р вЂ¦Р В РЎвЂўР В Р вЂ Р В Р’В»Р РЋРІР‚ВР В Р вЂ¦Р В Р вЂ¦Р В РЎвЂўР В Р’Вµ Р В РЎвЂ”Р В РЎвЂўР В Р’В»Р В РЎвЂўР В Р’В¶Р В Р’ВµР В Р вЂ¦Р В РЎвЂР В Р’Вµ Р В РЎвЂќР В РЎвЂўР РЋР С“Р В РЎВР В РЎвЂР РЋРІР‚РЋР В Р’ВµР РЋР С“Р В РЎвЂќР В РЎвЂўР В РЎвЂ“Р В РЎвЂў Р В РЎвЂўР В Р’В±Р РЋР вЂ°Р В Р’ВµР В РЎвЂќР РЋРІР‚С™Р В Р’В°.
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

// Р В РЎСџР РЋР вЂљР В РЎвЂўР В Р вЂ Р В Р’ВµР РЋР вЂљР РЋР РЏР В Р’ВµР РЋРІР‚С™, Р РЋРІР‚РЋР РЋРІР‚С™Р В РЎвЂў Р РЋР С“Р В РЎвЂўР РЋРІР‚В¦Р РЋР вЂљР В Р’В°Р В Р вЂ¦Р В Р’ВµР В Р вЂ¦Р В РЎвЂР В Р’Вµ Р В РўвЂР В Р’В°Р В Р вЂ¦Р В Р вЂ¦Р РЋРІР‚в„–Р РЋРІР‚В¦ Р В Р’В·Р В Р’В°Р В РЎвЂ”Р В РЎвЂР РЋР С“Р РЋРІР‚в„–Р В Р вЂ Р В Р’В°Р В Р’ВµР РЋРІР‚С™ Р В РЎвЂўР В Р’В±Р В Р вЂ¦Р В РЎвЂўР В Р вЂ Р В Р’В»Р РЋРІР‚ВР В Р вЂ¦Р В Р вЂ¦Р В РЎвЂўР В Р’Вµ Р В РЎвЂ”Р В РЎвЂўР В Р’В»Р В РЎвЂўР В Р’В¶Р В Р’ВµР В Р вЂ¦Р В РЎвЂР В Р’Вµ Р В РЎвЂќР В РЎвЂўР РЋР С“Р В РЎВР В РЎвЂР РЋРІР‚РЋР В Р’ВµР РЋР С“Р В РЎвЂќР В РЎвЂўР В РЎвЂ“Р В РЎвЂў Р В РЎвЂўР В Р’В±Р РЋР вЂ°Р В Р’ВµР В РЎвЂќР РЋРІР‚С™Р В Р’В°.
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

// Р В РЎСџР РЋР вЂљР В РЎвЂўР В Р вЂ Р В Р’ВµР РЋР вЂљР РЋР РЏР В Р’ВµР РЋРІР‚С™, Р РЋРІР‚РЋР РЋРІР‚С™Р В РЎвЂў Р РЋР С“Р В РЎвЂўР РЋРІР‚В¦Р РЋР вЂљР В Р’В°Р В Р вЂ¦Р В Р’ВµР В Р вЂ¦Р В РЎвЂР В Р’Вµ Р В РўвЂР В Р’В°Р В Р вЂ¦Р В Р вЂ¦Р РЋРІР‚в„–Р РЋРІР‚В¦ Р В Р’В·Р В Р’В°Р В РЎвЂ”Р В РЎвЂР РЋР С“Р РЋРІР‚в„–Р В Р вЂ Р В Р’В°Р В Р’ВµР РЋРІР‚С™ Р В РЎвЂўР В Р’В±Р В Р вЂ¦Р В РЎвЂўР В Р вЂ Р В Р’В»Р РЋРІР‚ВР В Р вЂ¦Р В Р вЂ¦Р В РЎвЂўР В Р’Вµ Р В РЎвЂ”Р В РЎвЂўР В Р’В»Р В РЎвЂўР В Р’В¶Р В Р’ВµР В Р вЂ¦Р В РЎвЂР В Р’Вµ Р В РЎвЂќР В РЎвЂўР РЋР С“Р В РЎВР В РЎвЂР РЋРІР‚РЋР В Р’ВµР РЋР С“Р В РЎвЂќР В РЎвЂўР В РЎвЂ“Р В РЎвЂў Р В РЎвЂўР В Р’В±Р РЋР вЂ°Р В Р’ВµР В РЎвЂќР РЋРІР‚С™Р В Р’В°.
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

// Р В РЎСџР РЋР вЂљР В РЎвЂўР В Р вЂ Р В Р’ВµР РЋР вЂљР РЋР РЏР В Р’ВµР РЋРІР‚С™, Р РЋРІР‚РЋР РЋРІР‚С™Р В РЎвЂў Р РЋР С“Р В РЎвЂўР РЋРІР‚В¦Р РЋР вЂљР В Р’В°Р В Р вЂ¦Р В Р’ВµР В Р вЂ¦Р В РЎвЂР В Р’Вµ Р В РўвЂР В Р’В°Р В Р вЂ¦Р В Р вЂ¦Р РЋРІР‚в„–Р РЋРІР‚В¦ Р В Р’В·Р В Р’В°Р В РЎвЂ”Р В РЎвЂР РЋР С“Р РЋРІР‚в„–Р В Р вЂ Р В Р’В°Р В Р’ВµР РЋРІР‚С™ Р В РЎвЂўР В Р’В±Р В Р вЂ¦Р В РЎвЂўР В Р вЂ Р В Р’В»Р РЋРІР‚ВР В Р вЂ¦Р В Р вЂ¦Р В РЎвЂўР В Р’Вµ Р В РЎвЂ”Р В РЎвЂўР В Р’В»Р В РЎвЂўР В Р’В¶Р В Р’ВµР В Р вЂ¦Р В РЎвЂР В Р’Вµ Р В РЎвЂќР В РЎвЂўР РЋР С“Р В РЎВР В РЎвЂР РЋРІР‚РЋР В Р’ВµР РЋР С“Р В РЎвЂќР В РЎвЂўР В РЎвЂ“Р В РЎвЂў Р В РЎвЂўР В Р’В±Р РЋР вЂ°Р В Р’ВµР В РЎвЂќР РЋРІР‚С™Р В Р’В°.
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

// Р В РЎСџР РЋР вЂљР В РЎвЂўР В Р вЂ Р В Р’ВµР РЋР вЂљР РЋР РЏР В Р’ВµР РЋРІР‚С™, Р РЋРІР‚РЋР РЋРІР‚С™Р В РЎвЂў Р РЋР С“Р В РЎвЂўР РЋРІР‚В¦Р РЋР вЂљР В Р’В°Р В Р вЂ¦Р В Р’ВµР В Р вЂ¦Р В РЎвЂР В Р’Вµ Р В РўвЂР В Р’В°Р В Р вЂ¦Р В Р вЂ¦Р РЋРІР‚в„–Р РЋРІР‚В¦ Р В Р’В·Р В Р’В°Р В РЎвЂ”Р В РЎвЂР РЋР С“Р РЋРІР‚в„–Р В Р вЂ Р В Р’В°Р В Р’ВµР РЋРІР‚С™ Р В РЎвЂўР В Р’В±Р В Р вЂ¦Р В РЎвЂўР В Р вЂ Р В Р’В»Р РЋРІР‚ВР В Р вЂ¦Р В Р вЂ¦Р В РЎвЂўР В Р’Вµ Р В РЎвЂ”Р В РЎвЂўР В Р’В»Р В РЎвЂўР В Р’В¶Р В Р’ВµР В Р вЂ¦Р В РЎвЂР В Р’Вµ Р В РЎвЂќР В РЎвЂўР РЋР С“Р В РЎВР В РЎвЂР РЋРІР‚РЋР В Р’ВµР РЋР С“Р В РЎвЂќР В РЎвЂўР В РЎвЂ“Р В РЎвЂў Р В РЎвЂўР В Р’В±Р РЋР вЂ°Р В Р’ВµР В РЎвЂќР РЋРІР‚С™Р В Р’В°.
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

// Р В РЎСџР РЋР вЂљР В РЎвЂўР В Р вЂ Р В Р’ВµР РЋР вЂљР РЋР РЏР В Р’ВµР РЋРІР‚С™, Р РЋРІР‚РЋР РЋРІР‚С™Р В РЎвЂў Р РЋР С“Р В РЎвЂўР РЋРІР‚В¦Р РЋР вЂљР В Р’В°Р В Р вЂ¦Р В Р’ВµР В Р вЂ¦Р В РЎвЂР В Р’Вµ Р В РўвЂР В Р’В°Р В Р вЂ¦Р В Р вЂ¦Р РЋРІР‚в„–Р РЋРІР‚В¦ Р В Р’В·Р В Р’В°Р В РЎвЂ”Р В РЎвЂР РЋР С“Р РЋРІР‚в„–Р В Р вЂ Р В Р’В°Р В Р’ВµР РЋРІР‚С™ Р В РЎвЂўР В Р’В±Р В Р вЂ¦Р В РЎвЂўР В Р вЂ Р В Р’В»Р РЋРІР‚ВР В Р вЂ¦Р В Р вЂ¦Р В РЎвЂўР В Р’Вµ Р В РЎвЂ”Р В РЎвЂўР В Р’В»Р В РЎвЂўР В Р’В¶Р В Р’ВµР В Р вЂ¦Р В РЎвЂР В Р’Вµ Р В РЎвЂќР В РЎвЂўР РЋР С“Р В РЎВР В РЎвЂР РЋРІР‚РЋР В Р’ВµР РЋР С“Р В РЎвЂќР В РЎвЂўР В РЎвЂ“Р В РЎвЂў Р В РЎвЂўР В Р’В±Р РЋР вЂ°Р В Р’ВµР В РЎвЂќР РЋРІР‚С™Р В Р’В°.
func TestCreateStarterAccountFillsFuelAndFirstContainerWithTypeSamples(t *testing.T) {
	serverData := testWorldData(t)
	serverData.ItemTypes.Items[17] = &data.ItemType{ID: 17, TitleRu: "Resource", TitleEn: "Resource", Acronym: "Resource"}
	serverData.ItemModels.Items[501] = &data.ItemModel{ID: 501, TitleRu: "Ore", TitleEn: "Ore", Acronym: "Ore", ItemTypeID: 17}
	serverData.ItemModels.Items[502] = &data.ItemModel{ID: 502, TitleRu: "Dust", TitleEn: "Dust", Acronym: "Dust", ItemTypeID: 17}
	if err := serverData.ItemTypes.RebuildIndexes(); err != nil {
		t.Fatal(err)
	}
	if err := serverData.ItemModels.RebuildIndexes(); err != nil {
		t.Fatal(err)
	}
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
	if len(items) != 5 {
		t.Fatalf("got %d container item groups, want 5", len(items))
	}
	counts := map[int64]float64{}
	for _, item := range items {
		counts[item.ContentItemModelID] = item.Count
	}
	for _, itemModelID := range []int64{101, 301} {
		if counts[itemModelID] != 10 {
			t.Fatalf("starter container item model %d count = %v, want 10; all counts: %+v", itemModelID, counts[itemModelID], counts)
		}
	}
	if counts[303] != 10000 {
		t.Fatalf("starter container ammo count = %v, want 10000; all counts: %+v", counts[303], counts)
	}
	for _, itemModelID := range []int64{501, 502} {
		if counts[itemModelID] != 1000 {
			t.Fatalf("starter container resource model %d count = %v, want 1000; all counts: %+v", itemModelID, counts[itemModelID], counts)
		}
	}
}

// Р В РЎСџР РЋР вЂљР В РЎвЂўР В Р вЂ Р В Р’ВµР РЋР вЂљР РЋР РЏР В Р’ВµР РЋРІР‚С™, Р РЋРІР‚РЋР РЋРІР‚С™Р В РЎвЂў Р РЋР С“Р В РЎвЂўР РЋРІР‚В¦Р РЋР вЂљР В Р’В°Р В Р вЂ¦Р В Р’ВµР В Р вЂ¦Р В РЎвЂР В Р’Вµ Р В РўвЂР В Р’В°Р В Р вЂ¦Р В Р вЂ¦Р РЋРІР‚в„–Р РЋРІР‚В¦ Р В Р’В·Р В Р’В°Р В РЎвЂ”Р В РЎвЂР РЋР С“Р РЋРІР‚в„–Р В Р вЂ Р В Р’В°Р В Р’ВµР РЋРІР‚С™ Р В РЎвЂўР В Р’В±Р В Р вЂ¦Р В РЎвЂўР В Р вЂ Р В Р’В»Р РЋРІР‚ВР В Р вЂ¦Р В Р вЂ¦Р В РЎвЂўР В Р’Вµ Р В РЎвЂ”Р В РЎвЂўР В Р’В»Р В РЎвЂўР В Р’В¶Р В Р’ВµР В Р вЂ¦Р В РЎвЂР В Р’Вµ Р В РЎвЂќР В РЎвЂўР РЋР С“Р В РЎВР В РЎвЂР РЋРІР‚РЋР В Р’ВµР РЋР С“Р В РЎвЂќР В РЎвЂўР В РЎвЂ“Р В РЎвЂў Р В РЎвЂўР В Р’В±Р РЋР вЂ°Р В Р’ВµР В РЎвЂќР РЋРІР‚С™Р В Р’В°.
func TestApplyControlPanelContainerTransferMovesAllItemsToTargetContainer(t *testing.T) {
	serverData := testWorldData(t)
	serverData.ItemTypes.Items[20] = &data.ItemType{ID: 20, TitleRu: "Робот", TitleEn: "Robot", Acronym: "Robot", CountMustBeInteger: true}
	serverData.ItemModels.Items[404] = &data.ItemModel{ID: 404, TitleRu: "Робот", TitleEn: "Robot", Acronym: "Robot", ItemTypeID: 20, ConsumingPower: 10}
	serverData.ItemModels.Items[303].Mass = 2
	cargoMovementType, ok := serverData.TaskTypes.GetByAcronym("CargoMovement")
	if !ok {
		t.Fatal("cargo movement task type was not loaded")
	}
	if err := serverData.ItemTypes.RebuildIndexes(); err != nil {
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
	if _, err := serverData.EquipmentGroups.Add(&data.EquipmentGroup{
		CosmicObjectID:       objectID,
		Title:                "Robots",
		EquipmentItemModelID: 404,
		Count:                10,
		EnabledCount:         10,
		Enabled:              true,
	}); err != nil {
		t.Fatal(err)
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
	target.OppositeEquipmentGroupID = source.ID
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
		ControllerEquipmentGroupID: target.ID,
		LeftToRightDirection:       true,
		ItemGroupIDs:               []int64{selectedItem.ID},
	}); err != nil {
		t.Fatalf("container transfer returned error: %v", err)
	}

	sourceItems := serverData.ItemGroups.GetByContainerEquipmentGroupID(source.ID)
	if len(sourceItems) != 2 {
		t.Fatalf("source container changed before movement started: %+v", serverData.ItemGroups.GetByContainerEquipmentGroupID(source.ID))
	}
	targetItems := serverData.ItemGroups.GetByContainerEquipmentGroupID(target.ID)
	if len(targetItems) != 1 {
		t.Fatalf("target container changed before movement finished: %+v", targetItems)
	}
	counts := map[int64]float64{}
	for _, item := range targetItems {
		counts[item.ContentItemModelID] = item.Count
	}
	if counts[303] != 5 {
		t.Fatalf("target container received cargo before movement finished: %+v", counts)
	}
	tasks := serverData.Tasks.GetByControllerEquipmentGroupID(target.ID)
	if len(tasks) != 1 {
		t.Fatalf("movement task was not queued: %+v", tasks)
	}
	task := tasks[0]
	if task.TaskTypeID != cargoMovementType.ID {
		t.Fatalf("task type = %d, want %d", task.TaskTypeID, cargoMovementType.ID)
	}
	if task.ControllerEquipmentGroupID != target.ID || !task.LeftToRightDirection {
		t.Fatalf("movement controller was not saved: %+v", task)
	}
	if task.BatchCount != 1 {
		t.Fatalf("movement batch count = %d, want 1", task.BatchCount)
	}
	reserved := serverData.TaskItemGroups.GetByTaskID(task.ID)
	if len(reserved) != 1 || reserved[0].ItemModelID != 303 || reserved[0].Count != 10 {
		t.Fatalf("cargo requirement was not saved in task item group: %+v", reserved)
	}
	// Проверяет, что работа считается по массе и полуразмеру текущего объекта.
	if math.Abs(task.TotalEnergy-106.875) > physics.Epsilon {
		t.Fatalf("movement energy = %v, want 106.875", task.TotalEnergy)
	}

	gameWorld.Tick(3)

	sourceItems = serverData.ItemGroups.GetByContainerEquipmentGroupID(source.ID)
	if len(sourceItems) != 1 || sourceItems[0].ID != remainingItem.ID {
		t.Fatalf("source container still has moved items: %+v", serverData.ItemGroups.GetByContainerEquipmentGroupID(source.ID))
	}

	targetItems = serverData.ItemGroups.GetByContainerEquipmentGroupID(target.ID)
	counts = map[int64]float64{}
	for _, item := range targetItems {
		counts[item.ContentItemModelID] = item.Count
	}
	if counts[303] != 15 {
		t.Fatalf("target container contents were not merged after movement: %+v", counts)
	}
	if reserved := serverData.TaskItemGroups.GetByTaskID(task.ID); len(reserved) != 0 {
		t.Fatalf("movement reserve was not cleared: %+v", reserved)
	}
}

// Р В РЎСџР РЋР вЂљР В РЎвЂўР В Р вЂ Р В Р’ВµР РЋР вЂљР РЋР РЏР В Р’ВµР РЋРІР‚С™, Р РЋРІР‚РЋР РЋРІР‚С™Р В РЎвЂў Р РЋР С“Р В РЎвЂўР РЋРІР‚В¦Р РЋР вЂљР В Р’В°Р В Р вЂ¦Р В Р’ВµР В Р вЂ¦Р В РЎвЂР В Р’Вµ Р В РўвЂР В Р’В°Р В Р вЂ¦Р В Р вЂ¦Р РЋРІР‚в„–Р РЋРІР‚В¦ Р В Р’В·Р В Р’В°Р В РЎвЂ”Р В РЎвЂР РЋР С“Р РЋРІР‚в„–Р В Р вЂ Р В Р’В°Р В Р’ВµР РЋРІР‚С™ Р В РЎвЂўР В Р’В±Р В Р вЂ¦Р В РЎвЂўР В Р вЂ Р В Р’В»Р РЋРІР‚ВР В Р вЂ¦Р В Р вЂ¦Р В РЎвЂўР В Р’Вµ Р В РЎвЂ”Р В РЎвЂўР В Р’В»Р В РЎвЂўР В Р’В¶Р В Р’ВµР В Р вЂ¦Р В РЎвЂР В Р’Вµ Р В РЎвЂќР В РЎвЂўР РЋР С“Р В РЎВР В РЎвЂР РЋРІР‚РЋР В Р’ВµР РЋР С“Р В РЎвЂќР В РЎвЂўР В РЎвЂ“Р В РЎвЂў Р В РЎвЂўР В Р’В±Р РЋР вЂ°Р В Р’ВµР В РЎвЂќР РЋРІР‚С™Р В Р’В°.
// Р В РЎСџР РЋР вЂљР В РЎвЂўР В Р вЂ Р В Р’ВµР РЋР вЂљР РЋР РЏР В Р’ВµР РЋРІР‚С™, Р РЋРІР‚РЋР РЋРІР‚С™Р В РЎвЂў Р РЋР С“Р В РЎвЂўР РЋРІР‚В¦Р РЋР вЂљР В Р’В°Р В Р вЂ¦Р В Р’ВµР В Р вЂ¦Р В РЎвЂР В Р’Вµ Р В РўвЂР В Р’В°Р В Р вЂ¦Р В Р вЂ¦Р РЋРІР‚в„–Р РЋРІР‚В¦ Р В Р’В·Р В Р’В°Р В РЎвЂ”Р В РЎвЂР РЋР С“Р РЋРІР‚в„–Р В Р вЂ Р В Р’В°Р В Р’ВµР РЋРІР‚С™ Р В РЎвЂўР В Р’В±Р В Р вЂ¦Р В РЎвЂўР В Р вЂ Р В Р’В»Р РЋРІР‚ВР В Р вЂ¦Р В Р вЂ¦Р В РЎвЂўР В Р’Вµ Р В РЎвЂ”Р В РЎвЂўР В Р’В»Р В РЎвЂўР В Р’В¶Р В Р’ВµР В Р вЂ¦Р В РЎвЂР В Р’Вµ Р В РЎвЂќР В РЎвЂўР РЋР С“Р В РЎВР В РЎвЂР РЋРІР‚РЋР В Р’ВµР РЋР С“Р В РЎвЂќР В РЎвЂўР В РЎвЂ“Р В РЎвЂў Р В РЎвЂўР В Р’В±Р РЋР вЂ°Р В Р’ВµР В РЎвЂќР РЋРІР‚С™Р В Р’В°.
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
	target.OppositeEquipmentGroupID = source.ID
	selectedItem, err := serverData.ItemGroups.Add(&data.ItemGroup{ContainerEquipmentGroupID: source.ID, ContentItemModelID: 303, Count: 10})
	if err != nil {
		t.Fatal(err)
	}

	if err := gameWorld.ApplyControlPanelContainerTransfer(1, "session-1", 9, world.ControlPanelContainerTransfer{
		ControllerEquipmentGroupID: target.ID,
		LeftToRightDirection:       true,
		ItemGroupIDs:               []int64{selectedItem.ID},
		Amount:                     4,
	}); err != nil {
		t.Fatalf("container transfer returned error: %v", err)
	}

	sourceItems := serverData.ItemGroups.GetByContainerEquipmentGroupID(source.ID)
	if len(sourceItems) != 1 || sourceItems[0].Count != 10 {
		t.Fatalf("source item changed before task start: %+v", sourceItems)
	}
	targetItems := serverData.ItemGroups.GetByContainerEquipmentGroupID(target.ID)
	if len(targetItems) != 0 {
		t.Fatalf("target item was moved before task completion: %+v", targetItems)
	}
	tasks := serverData.Tasks.GetByControllerEquipmentGroupID(target.ID)
	if len(tasks) != 1 {
		t.Fatalf("movement task was not queued: %+v", tasks)
	}
	if tasks[0].BatchCount != 1 {
		t.Fatalf("movement batch count = %d, want 1", tasks[0].BatchCount)
	}
	reserved := serverData.TaskItemGroups.GetByTaskID(tasks[0].ID)
	if len(reserved) != 1 || reserved[0].ItemModelID != 303 || reserved[0].Count != 4 {
		t.Fatalf("requested amount was not saved as requirement: %+v", reserved)
	}
}

// Р В РЎСџР РЋР вЂљР В РЎвЂўР В Р вЂ Р В Р’ВµР РЋР вЂљР РЋР РЏР В Р’ВµР РЋРІР‚С™, Р РЋРІР‚РЋР РЋРІР‚С™Р В РЎвЂў Р РЋР С“Р В РЎвЂўР РЋРІР‚В¦Р РЋР вЂљР В Р’В°Р В Р вЂ¦Р В Р’ВµР В Р вЂ¦Р В РЎвЂР В Р’Вµ Р В РўвЂР В Р’В°Р В Р вЂ¦Р В Р вЂ¦Р РЋРІР‚в„–Р РЋРІР‚В¦ Р В Р’В·Р В Р’В°Р В РЎвЂ”Р В РЎвЂР РЋР С“Р РЋРІР‚в„–Р В Р вЂ Р В Р’В°Р В Р’ВµР РЋРІР‚С™ Р В РЎвЂўР В Р’В±Р В Р вЂ¦Р В РЎвЂўР В Р вЂ Р В Р’В»Р РЋРІР‚ВР В Р вЂ¦Р В Р вЂ¦Р В РЎвЂўР В Р’Вµ Р В РЎвЂ”Р В РЎвЂўР В Р’В»Р В РЎвЂўР В Р’В¶Р В Р’ВµР В Р вЂ¦Р В РЎвЂР В Р’Вµ Р В РЎвЂќР В РЎвЂўР РЋР С“Р В РЎВР В РЎвЂР РЋРІР‚РЋР В Р’ВµР РЋР С“Р В РЎвЂќР В РЎвЂўР В РЎвЂ“Р В РЎвЂў Р В РЎвЂўР В Р’В±Р РЋР вЂ°Р В Р’ВµР В РЎвЂќР РЋРІР‚С™Р В Р’В°.
// Р В РЎСџР РЋР вЂљР В РЎвЂўР В Р вЂ Р В Р’ВµР РЋР вЂљР РЋР РЏР В Р’ВµР РЋРІР‚С™, Р РЋРІР‚РЋР РЋРІР‚С™Р В РЎвЂў Р РЋР С“Р В РЎвЂўР РЋРІР‚В¦Р РЋР вЂљР В Р’В°Р В Р вЂ¦Р В Р’ВµР В Р вЂ¦Р В РЎвЂР В Р’Вµ Р В РўвЂР В Р’В°Р В Р вЂ¦Р В Р вЂ¦Р РЋРІР‚в„–Р РЋРІР‚В¦ Р В Р’В·Р В Р’В°Р В РЎвЂ”Р В РЎвЂР РЋР С“Р РЋРІР‚в„–Р В Р вЂ Р В Р’В°Р В Р’ВµР РЋРІР‚С™ Р В РЎвЂўР В Р’В±Р В Р вЂ¦Р В РЎвЂўР В Р вЂ Р В Р’В»Р РЋРІР‚ВР В Р вЂ¦Р В Р вЂ¦Р В РЎвЂўР В Р’Вµ Р В РЎвЂ”Р В РЎвЂўР В Р’В»Р В РЎвЂўР В Р’В¶Р В Р’ВµР В Р вЂ¦Р В РЎвЂР В Р’Вµ Р В РЎвЂќР В РЎвЂўР РЋР С“Р В РЎВР В РЎвЂР РЋРІР‚РЋР В Р’ВµР РЋР С“Р В РЎвЂќР В РЎвЂўР В РЎвЂ“Р В РЎвЂў Р В РЎвЂўР В Р’В±Р РЋР вЂ°Р В Р’ВµР В РЎвЂќР РЋРІР‚С™Р В Р’В°.
func TestCargoMovementUsesCurrentOppositeContainerOnlyBeforeStart(t *testing.T) {
	serverData := testWorldData(t)
	serverData.ItemTypes.Items[20] = &data.ItemType{ID: 20, TitleRu: "Робот", TitleEn: "Robot", Acronym: "Robot", CountMustBeInteger: true}
	serverData.ItemModels.Items[404] = &data.ItemModel{ID: 404, TitleRu: "Робот", TitleEn: "Robot", Acronym: "Robot", ItemTypeID: 20, ConsumingPower: 10}
	serverData.ItemModels.Items[303].Mass = 2
	if err := serverData.ItemTypes.RebuildIndexes(); err != nil {
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
	var leftOld *data.EquipmentGroup
	for _, group := range serverData.EquipmentGroups.GetByCosmicObjectID(objectID) {
		if group.EquipmentItemModelID == 301 {
			leftOld = group
			break
		}
	}
	if leftOld == nil {
		t.Fatalf("left container was not installed")
	}
	right, err := serverData.EquipmentGroups.Add(&data.EquipmentGroup{
		CosmicObjectID:           objectID,
		Title:                    "Right Container",
		EquipmentItemModelID:     301,
		Count:                    1,
		EnabledCount:             1,
		Enabled:                  true,
		OppositeEquipmentGroupID: leftOld.ID,
	})
	if err != nil {
		t.Fatal(err)
	}
	leftNew, err := serverData.EquipmentGroups.Add(&data.EquipmentGroup{
		CosmicObjectID:       objectID,
		Title:                "New Left Container",
		EquipmentItemModelID: 301,
		Count:                1,
		EnabledCount:         1,
		Enabled:              true,
	})
	if err != nil {
		t.Fatal(err)
	}
	if _, err := serverData.EquipmentGroups.Add(&data.EquipmentGroup{
		CosmicObjectID:       objectID,
		Title:                "Robots",
		EquipmentItemModelID: 404,
		Count:                10,
		EnabledCount:         10,
		Enabled:              true,
	}); err != nil {
		t.Fatal(err)
	}
	selectedOld, err := serverData.ItemGroups.Add(&data.ItemGroup{ContainerEquipmentGroupID: leftOld.ID, ContentItemModelID: 303, Count: 10})
	if err != nil {
		t.Fatal(err)
	}
	selectedRight, err := serverData.ItemGroups.Add(&data.ItemGroup{ContainerEquipmentGroupID: right.ID, ContentItemModelID: 303, Count: 7})
	if err != nil {
		t.Fatal(err)
	}

	if err := gameWorld.ApplyControlPanelContainerTransfer(1, "session-1", 11, world.ControlPanelContainerTransfer{
		ControllerEquipmentGroupID: right.ID,
		LeftToRightDirection:       true,
		ItemGroupIDs:               []int64{selectedOld.ID},
	}); err != nil {
		t.Fatalf("left-to-right transfer returned error: %v", err)
	}
	if items := serverData.ItemGroups.GetByContainerEquipmentGroupID(leftOld.ID); len(items) != 1 || items[0].Count != 10 {
		t.Fatalf("cargo was reserved before task start: %+v", items)
	}
	right.OppositeEquipmentGroupID = leftNew.ID
	if _, err := serverData.ItemGroups.Add(&data.ItemGroup{ContainerEquipmentGroupID: leftNew.ID, ContentItemModelID: 303, Count: 10}); err != nil {
		t.Fatal(err)
	}
	gameWorld.Tick(1)
	runningTasks := serverData.Tasks.GetByControllerEquipmentGroupID(right.ID)
	if len(runningTasks) == 0 {
		t.Fatalf("running task was not created")
	}
	if reserves := serverData.TaskItemGroups.GetByTaskID(runningTasks[0].ID); len(reserves) != 1 || reserves[0].Count != 10 || !reserves[0].IsStored {
		t.Fatalf("running task did not move cargo into task storage: %+v", reserves)
	}
	if items := serverData.ItemGroups.GetByContainerEquipmentGroupID(leftOld.ID); len(items) != 1 || items[0].Count != 10 {
		t.Fatalf("running task unexpectedly used old source after opposite change: %+v", items)
	}
	if items := serverData.ItemGroups.GetByContainerEquipmentGroupID(leftNew.ID); len(items) != 0 {
		t.Fatalf("running task did not use new source after opposite change: %+v", items)
	}

	if err := gameWorld.ApplyControlPanelContainerTransfer(1, "session-1", 12, world.ControlPanelContainerTransfer{
		ControllerEquipmentGroupID: right.ID,
		LeftToRightDirection:       false,
		ItemGroupIDs:               []int64{selectedRight.ID},
	}); err != nil {
		t.Fatalf("right-to-left transfer returned error: %v", err)
	}
	gameWorld.Tick(20)
	gameWorld.Tick(20)
	counts := map[int64]float64{}
	for _, item := range serverData.ItemGroups.GetByContainerEquipmentGroupID(leftNew.ID) {
		counts[item.ContentItemModelID] += item.Count
	}
	if counts[303] != 7 {
		t.Fatalf("new left container did not receive completed cargo: %+v", counts)
	}
}

func TestApplyControlPanelContainerTransferUsesOwnedClusterObject(t *testing.T) {
	serverData := testWorldData(t)
	gameWorld := world.New(1, serverData)

	objectID, ok := gameWorld.ConnectAccount(1)
	if !ok {
		t.Fatalf("account was not connected")
	}
	mainObject := serverData.CosmicObjects.Items[objectID]
	mainObject.ClusterMainCosmicObjectID = mainObject.ID
	targetObject, err := serverData.CosmicObjects.Add(&data.CosmicObject{
		Title:                     "Cluster Container Ship",
		CosmicObjectModelID:       mainObject.CosmicObjectModelID,
		OwnerCharacterID:          mainObject.OwnerCharacterID,
		CreatorCharacterID:        mainObject.OwnerCharacterID,
		Enabled:                   true,
		ClusterMainCosmicObjectID: mainObject.ID,
	})
	if err != nil {
		t.Fatal(err)
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
		CosmicObjectID:       targetObject.ID,
		Title:                "Target Cluster Container",
		EquipmentItemModelID: 301,
		Count:                1,
		EnabledCount:         1,
		Enabled:              true,
	})
	if err != nil {
		t.Fatal(err)
	}
	target.OppositeEquipmentGroupID = source.ID
	selectedItem, err := serverData.ItemGroups.Add(&data.ItemGroup{ContainerEquipmentGroupID: source.ID, ContentItemModelID: 303, Count: 10})
	if err != nil {
		t.Fatal(err)
	}

	if err := gameWorld.ApplyControlPanelContainerTransfer(1, "session-1", 10, world.ControlPanelContainerTransfer{
		ControllerEquipmentGroupID: target.ID,
		LeftToRightDirection:       true,
		ItemGroupIDs:               []int64{selectedItem.ID},
	}); err != nil {
		t.Fatalf("cluster container transfer returned error: %v", err)
	}

	targetItems := serverData.ItemGroups.GetByContainerEquipmentGroupID(target.ID)
	if len(targetItems) != 0 {
		t.Fatalf("target cluster container changed before task completion: %+v", targetItems)
	}
	tasks := serverData.Tasks.GetByControllerEquipmentGroupID(target.ID)
	if len(tasks) != 1 || tasks[0].ControllerEquipmentGroupID != target.ID || !tasks[0].LeftToRightDirection || tasks[0].BatchCount != 1 {
		t.Fatalf("cluster movement task was not queued: %+v", tasks)
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

// Р В РЎСџР РЋР вЂљР В РЎвЂўР В Р вЂ Р В Р’ВµР РЋР вЂљР РЋР РЏР В Р’ВµР РЋРІР‚С™, Р РЋРІР‚РЋР РЋРІР‚С™Р В РЎвЂў Р РЋР С“Р В РЎвЂўР РЋРІР‚В¦Р РЋР вЂљР В Р’В°Р В Р вЂ¦Р В Р’ВµР В Р вЂ¦Р В РЎвЂР В Р’Вµ Р В РўвЂР В Р’В°Р В Р вЂ¦Р В Р вЂ¦Р РЋРІР‚в„–Р РЋРІР‚В¦ Р В Р’В·Р В Р’В°Р В РЎвЂ”Р В РЎвЂР РЋР С“Р РЋРІР‚в„–Р В Р вЂ Р В Р’В°Р В Р’ВµР РЋРІР‚С™ Р В РЎвЂўР В Р’В±Р В Р вЂ¦Р В РЎвЂўР В Р вЂ Р В Р’В»Р РЋРІР‚ВР В Р вЂ¦Р В Р вЂ¦Р В РЎвЂўР В Р’Вµ Р В РЎвЂ”Р В РЎвЂўР В Р’В»Р В РЎвЂўР В Р’В¶Р В Р’ВµР В Р вЂ¦Р В РЎвЂР В Р’Вµ Р В РЎвЂќР В РЎвЂўР РЋР С“Р В РЎВР В РЎвЂР РЋРІР‚РЋР В Р’ВµР РЋР С“Р В РЎвЂќР В РЎвЂўР В РЎвЂ“Р В РЎвЂў Р В РЎвЂўР В Р’В±Р РЋР вЂ°Р В Р’ВµР В РЎвЂќР РЋРІР‚С™Р В Р’В°.
func TestApplyControlPanelConstructorProduceItemQueuesAndCompletesAfterProductionTime(t *testing.T) {
	serverData := testWorldData(t)
	serverData.ItemTypes.Items[19] = &data.ItemType{ID: 19, TitleRu: "Constructor", TitleEn: "Constructor", Acronym: "Constructor", CountMustBeInteger: true}
	serverData.ItemModels.Items[501] = &data.ItemModel{ID: 501, TitleRu: "Constructor", TitleEn: "Constructor", Acronym: "Constructor", ItemTypeID: 19}
	if err := serverData.ItemTypes.RebuildIndexes(); err != nil {
		t.Fatal(err)
	}
	if err := serverData.ItemModels.RebuildIndexes(); err != nil {
		t.Fatal(err)
	}
	serverData.Schemas = storage.NewRawReferenceTable()
	serverData.SchemaComponents = storage.NewRawReferenceTable()
	addRawReferenceItem(t, serverData.Schemas, 1, map[string]any{"ID": 1, "ItemModelID": 302, "Count": 1, "ProductionEnergy": 10})
	addRawReferenceItem(t, serverData.SchemaComponents, 1, map[string]any{"ID": 1, "SchemaID": 1, "ComponentItemModelID": 303, "Count": 4})
	addRawReferenceItem(t, serverData.SchemaComponents, 2, map[string]any{"ID": 2, "SchemaID": 1, "ComponentItemModelID": 403, "Count": 2})
	gameWorld := world.New(1, serverData)

	objectID, ok := gameWorld.ConnectAccount(1)
	if !ok {
		t.Fatalf("account was not connected")
	}
	var materialContainer *data.EquipmentGroup
	for _, group := range serverData.EquipmentGroups.GetByCosmicObjectID(objectID) {
		if group.EquipmentItemModelID == 301 {
			materialContainer = group
			break
		}
	}
	if materialContainer == nil {
		t.Fatalf("material container was not installed")
	}
	staleMaterialContainer, err := serverData.EquipmentGroups.Add(&data.EquipmentGroup{
		CosmicObjectID:       objectID,
		Title:                "Old Material Container",
		EquipmentItemModelID: 301,
		Count:                1,
		EnabledCount:         1,
		Enabled:              true,
	})
	if err != nil {
		t.Fatal(err)
	}
	productContainer, err := serverData.EquipmentGroups.Add(&data.EquipmentGroup{
		CosmicObjectID:       objectID,
		Title:                "Product Container",
		EquipmentItemModelID: 301,
		Count:                1,
		EnabledCount:         1,
		Enabled:              true,
	})
	if err != nil {
		t.Fatal(err)
	}
	constructor, err := serverData.EquipmentGroups.Add(&data.EquipmentGroup{
		CosmicObjectID:         objectID,
		Title:                  "Constructor",
		EquipmentItemModelID:   501,
		SourceEquipmentGroupID: staleMaterialContainer.ID,
		Count:                  1,
		EnabledCount:           1,
		Enabled:                true,
	})
	if err != nil {
		t.Fatal(err)
	}
	if _, err := serverData.ItemGroups.Add(&data.ItemGroup{ContainerEquipmentGroupID: materialContainer.ID, ContentItemModelID: 303, Count: 12}); err != nil {
		t.Fatal(err)
	}
	if _, err := serverData.ItemGroups.Add(&data.ItemGroup{ContainerEquipmentGroupID: materialContainer.ID, ContentItemModelID: 403, Count: 6}); err != nil {
		t.Fatal(err)
	}

	if err := gameWorld.ApplyControlPanelConstructorProduceItem(1, "session-1", 11, world.ControlPanelConstructorProduceItem{
		ConstructorEquipmentGroupID:       constructor.ID,
		MaterialContainerEquipmentGroupID: materialContainer.ID,
		ProductContainerEquipmentGroupID:  productContainer.ID,
		SchemaID:                          1,
		Amount:                            3,
	}); err != nil {
		t.Fatalf("constructor production returned error: %v", err)
	}

	productCounts := map[int64]float64{}
	for _, item := range serverData.ItemGroups.GetByContainerEquipmentGroupID(productContainer.ID) {
		productCounts[item.ContentItemModelID] = item.Count
	}
	if productCounts[302] != 0 {
		t.Fatalf("product was created before production time passed: %+v", productCounts)
	}
	queued := gameWorld.SnapshotForAccount(1).ConstructorProductionJobs
	if len(queued) != 1 || queued[0].QueueType != "main" || queued[0].RemainingTime != 30 || queued[0].RemainingCount != 3 || queued[0].TotalCount != 3 {
		t.Fatalf("main production was not queued correctly: %+v", queued)
	}

	gameWorld.Tick(10)
	queued = gameWorld.SnapshotForAccount(1).ConstructorProductionJobs
	if len(queued) != 1 || queued[0].RemainingCount != 2 || queued[0].TotalCount != 3 {
		t.Fatalf("main production did not keep remaining and total counts: %+v", queued)
	}
	gameWorld.Tick(10)
	gameWorld.Tick(10)

	materialCounts := map[int64]float64{}
	for _, item := range serverData.ItemGroups.GetByContainerEquipmentGroupID(materialContainer.ID) {
		materialCounts[item.ContentItemModelID] = item.Count
	}
	if materialCounts[303] != 0 || materialCounts[403] != 0 {
		t.Fatalf("materials were not consumed correctly: %+v", materialCounts)
	}
	productCounts = map[int64]float64{}
	for _, item := range serverData.ItemGroups.GetByContainerEquipmentGroupID(productContainer.ID) {
		productCounts[item.ContentItemModelID] = item.Count
	}
	if productCounts[302] != 3 {
		t.Fatalf("product was not created after production time: %+v", productCounts)
	}
	if queued := gameWorld.SnapshotForAccount(1).ConstructorProductionJobs; len(queued) != 0 {
		t.Fatalf("completed production stayed in queue: %+v", queued)
	}
}

// Р В РЎСџР РЋР вЂљР В РЎвЂўР В Р вЂ Р В Р’ВµР РЋР вЂљР РЋР РЏР В Р’ВµР РЋРІР‚С™, Р РЋРІР‚РЋР РЋРІР‚С™Р В РЎвЂў Р РЋР С“Р В РЎвЂўР РЋРІР‚В¦Р РЋР вЂљР В Р’В°Р В Р вЂ¦Р В Р’ВµР В Р вЂ¦Р В РЎвЂР В Р’Вµ Р В РўвЂР В Р’В°Р В Р вЂ¦Р В Р вЂ¦Р РЋРІР‚в„–Р РЋРІР‚В¦ Р В Р’В·Р В Р’В°Р В РЎвЂ”Р В РЎвЂР РЋР С“Р РЋРІР‚в„–Р В Р вЂ Р В Р’В°Р В Р’ВµР РЋРІР‚С™ Р В РЎвЂўР В Р’В±Р В Р вЂ¦Р В РЎвЂўР В Р вЂ Р В Р’В»Р РЋРІР‚ВР В Р вЂ¦Р В Р вЂ¦Р В РЎвЂўР В Р’Вµ Р В РЎвЂ”Р В РЎвЂўР В Р’В»Р В РЎвЂўР В Р’В¶Р В Р’ВµР В Р вЂ¦Р В РЎвЂР В Р’Вµ Р В РЎвЂќР В РЎвЂўР РЋР С“Р В РЎВР В РЎвЂР РЋРІР‚РЋР В Р’ВµР РЋР С“Р В РЎвЂќР В РЎвЂўР В РЎвЂ“Р В РЎвЂў Р В РЎвЂўР В Р’В±Р РЋР вЂ°Р В Р’ВµР В РЎвЂќР РЋРІР‚С™Р В Р’В°.
// Р В РЎСџР РЋР вЂљР В РЎвЂўР В Р вЂ Р В Р’ВµР РЋР вЂљР РЋР РЏР В Р’ВµР РЋРІР‚С™, Р РЋРІР‚РЋР РЋРІР‚С™Р В РЎвЂў Р РЋР С“Р В РЎвЂўР РЋРІР‚В¦Р РЋР вЂљР В Р’В°Р В Р вЂ¦Р В Р’ВµР В Р вЂ¦Р В РЎвЂР В Р’Вµ Р В РўвЂР В Р’В°Р В Р вЂ¦Р В Р вЂ¦Р РЋРІР‚в„–Р РЋРІР‚В¦ Р В Р’В·Р В Р’В°Р В РЎвЂ”Р В РЎвЂР РЋР С“Р РЋРІР‚в„–Р В Р вЂ Р В Р’В°Р В Р’ВµР РЋРІР‚С™ Р В РЎвЂўР В Р’В±Р В Р вЂ¦Р В РЎвЂўР В Р вЂ Р В Р’В»Р РЋРІР‚ВР В Р вЂ¦Р В Р вЂ¦Р В РЎвЂўР В Р’Вµ Р В РЎвЂ”Р В РЎвЂўР В Р’В»Р В РЎвЂўР В Р’В¶Р В Р’ВµР В Р вЂ¦Р В РЎвЂР В Р’Вµ Р В РЎвЂќР В РЎвЂўР РЋР С“Р В РЎВР В РЎвЂР РЋРІР‚РЋР В Р’ВµР РЋР С“Р В РЎвЂќР В РЎвЂўР В РЎвЂ“Р В РЎвЂў Р В РЎвЂўР В Р’В±Р РЋР вЂ°Р В Р’ВµР В РЎвЂќР РЋРІР‚С™Р В Р’В°.
// Р В РЎСџР РЋР вЂљР В РЎвЂўР В Р вЂ Р В Р’ВµР РЋР вЂљР РЋР РЏР В Р’ВµР РЋРІР‚С™, Р РЋРІР‚РЋР РЋРІР‚С™Р В РЎвЂў Р РЋР С“Р В РЎвЂўР РЋРІР‚В¦Р РЋР вЂљР В Р’В°Р В Р вЂ¦Р В Р’ВµР В Р вЂ¦Р В РЎвЂР В Р’Вµ Р В РўвЂР В Р’В°Р В Р вЂ¦Р В Р вЂ¦Р РЋРІР‚в„–Р РЋРІР‚В¦ Р В Р’В·Р В Р’В°Р В РЎвЂ”Р В РЎвЂР РЋР С“Р РЋРІР‚в„–Р В Р вЂ Р В Р’В°Р В Р’ВµР РЋРІР‚С™ Р В РЎвЂўР В Р’В±Р В Р вЂ¦Р В РЎвЂўР В Р вЂ Р В Р’В»Р РЋРІР‚ВР В Р вЂ¦Р В Р вЂ¦Р В РЎвЂўР В Р’Вµ Р В РЎвЂ”Р В РЎвЂўР В Р’В»Р В РЎвЂўР В Р’В¶Р В Р’ВµР В Р вЂ¦Р В РЎвЂР В Р’Вµ Р В РЎвЂќР В РЎвЂўР РЋР С“Р В РЎВР В РЎвЂР РЋРІР‚РЋР В Р’ВµР РЋР С“Р В РЎвЂќР В РЎвЂўР В РЎвЂ“Р В РЎвЂў Р В РЎвЂўР В Р’В±Р РЋР вЂ°Р В Р’ВµР В РЎвЂќР РЋРІР‚С™Р В Р’В°.
// Р В РЎСџР РЋР вЂљР В РЎвЂўР В Р вЂ Р В Р’ВµР РЋР вЂљР РЋР РЏР В Р’ВµР РЋРІР‚С™, Р РЋРІР‚РЋР РЋРІР‚С™Р В РЎвЂў Р РЋР С“Р В РЎвЂўР РЋРІР‚В¦Р РЋР вЂљР В Р’В°Р В Р вЂ¦Р В Р’ВµР В Р вЂ¦Р В РЎвЂР В Р’Вµ Р В РўвЂР В Р’В°Р В Р вЂ¦Р В Р вЂ¦Р РЋРІР‚в„–Р РЋРІР‚В¦ Р В Р’В·Р В Р’В°Р В РЎвЂ”Р В РЎвЂР РЋР С“Р РЋРІР‚в„–Р В Р вЂ Р В Р’В°Р В Р’ВµР РЋРІР‚С™ Р В РЎвЂўР В Р’В±Р В Р вЂ¦Р В РЎвЂўР В Р вЂ Р В Р’В»Р РЋРІР‚ВР В Р вЂ¦Р В Р вЂ¦Р В РЎвЂўР В Р’Вµ Р В РЎвЂ”Р В РЎвЂўР В Р’В»Р В РЎвЂўР В Р’В¶Р В Р’ВµР В Р вЂ¦Р В РЎвЂР В Р’Вµ Р В РЎвЂќР В РЎвЂўР РЋР С“Р В РЎВР В РЎвЂР РЋРІР‚РЋР В Р’ВµР РЋР С“Р В РЎвЂќР В РЎвЂўР В РЎвЂ“Р В РЎвЂў Р В РЎвЂўР В Р’В±Р РЋР вЂ°Р В Р’ВµР В РЎвЂќР РЋРІР‚С™Р В Р’В°.
func TestApplyControlPanelConstructorProduceItemStoresAmountInSingleTask(t *testing.T) {
	serverData := testWorldData(t)
	serverData.ItemTypes.Items[19] = &data.ItemType{ID: 19, TitleRu: "Constructor", TitleEn: "Constructor", Acronym: "Constructor", CountMustBeInteger: true}
	serverData.ItemModels.Items[501] = &data.ItemModel{ID: 501, TitleRu: "Constructor", TitleEn: "Constructor", Acronym: "Constructor", ItemTypeID: 19}
	if err := serverData.ItemTypes.RebuildIndexes(); err != nil {
		t.Fatal(err)
	}
	if err := serverData.ItemModels.RebuildIndexes(); err != nil {
		t.Fatal(err)
	}
	serverData.Schemas = storage.NewRawReferenceTable()
	serverData.SchemaComponents = storage.NewRawReferenceTable()
	addRawReferenceItem(t, serverData.Schemas, 1, map[string]any{"ID": 1, "ItemModelID": 302, "Count": 1, "ProductionEnergy": 10})
	addRawReferenceItem(t, serverData.SchemaComponents, 1, map[string]any{"ID": 1, "SchemaID": 1, "ComponentItemModelID": 303, "Count": 4})
	gameWorld := world.New(1, serverData)

	objectID, ok := gameWorld.ConnectAccount(1)
	if !ok {
		t.Fatalf("account was not connected")
	}
	var materialContainer *data.EquipmentGroup
	for _, group := range serverData.EquipmentGroups.GetByCosmicObjectID(objectID) {
		if group.EquipmentItemModelID == 301 {
			materialContainer = group
			break
		}
	}
	if materialContainer == nil {
		t.Fatalf("material container was not installed")
	}
	productContainer, err := serverData.EquipmentGroups.Add(&data.EquipmentGroup{
		CosmicObjectID:       objectID,
		Title:                "Product Container",
		EquipmentItemModelID: 301,
		Count:                1,
		EnabledCount:         1,
		Enabled:              true,
	})
	if err != nil {
		t.Fatal(err)
	}
	constructor, err := serverData.EquipmentGroups.Add(&data.EquipmentGroup{
		CosmicObjectID:       objectID,
		Title:                "Constructor",
		EquipmentItemModelID: 501,
		Count:                1,
		EnabledCount:         1,
		Enabled:              true,
	})
	if err != nil {
		t.Fatal(err)
	}
	if _, err := serverData.ItemGroups.Add(&data.ItemGroup{ContainerEquipmentGroupID: materialContainer.ID, ContentItemModelID: 303, Count: 12}); err != nil {
		t.Fatal(err)
	}

	if err := gameWorld.ApplyControlPanelConstructorProduceItem(1, "session-1", 21, world.ControlPanelConstructorProduceItem{
		ConstructorEquipmentGroupID:       constructor.ID,
		MaterialContainerEquipmentGroupID: materialContainer.ID,
		ProductContainerEquipmentGroupID:  productContainer.ID,
		SchemaID:                          1,
		Amount:                            3,
	}); err != nil {
		t.Fatalf("constructor production returned error: %v", err)
	}

	tasks := serverData.Tasks.GetByControllerEquipmentGroupID(constructor.ID)
	if len(tasks) != 1 {
		t.Fatalf("production amount was split into separate tasks: %+v", tasks)
	}
	if tasks[0].BatchCount != 3 {
		t.Fatalf("production batch count = %d, want 3", tasks[0].BatchCount)
	}
}

func TestTickActivatesConstructorEquipmentWhileProductionRuns(t *testing.T) {
	serverData := testWorldData(t)
	serverData.ItemTypes.Items[19] = &data.ItemType{ID: 19, TitleRu: "Constructor", TitleEn: "Constructor", Acronym: "Constructor", CountMustBeInteger: true}
	serverData.ItemModels.Items[501] = &data.ItemModel{ID: 501, TitleRu: "Constructor", TitleEn: "Constructor", Acronym: "Constructor", ItemTypeID: 19, ConsumingPower: 7000}
	if err := serverData.ItemTypes.RebuildIndexes(); err != nil {
		t.Fatal(err)
	}
	if err := serverData.ItemModels.RebuildIndexes(); err != nil {
		t.Fatal(err)
	}
	serverData.Schemas = storage.NewRawReferenceTable()
	serverData.SchemaComponents = storage.NewRawReferenceTable()
	addRawReferenceItem(t, serverData.Schemas, 1, map[string]any{"ID": 1, "ItemModelID": 302, "Count": 1, "ProductionEnergy": 70000})
	addRawReferenceItem(t, serverData.SchemaComponents, 1, map[string]any{"ID": 1, "SchemaID": 1, "ComponentItemModelID": 303, "Count": 4})
	gameWorld := world.New(1, serverData)

	objectID, ok := gameWorld.ConnectAccount(1)
	if !ok {
		t.Fatalf("account was not connected")
	}
	addTestPowerProducer(t, serverData, objectID)
	var materialContainer *data.EquipmentGroup
	for _, group := range serverData.EquipmentGroups.GetByCosmicObjectID(objectID) {
		if group.EquipmentItemModelID == 301 {
			materialContainer = group
			break
		}
	}
	if materialContainer == nil {
		t.Fatalf("material container was not installed")
	}
	productContainer, err := serverData.EquipmentGroups.Add(&data.EquipmentGroup{CosmicObjectID: objectID, Title: "Product Container", EquipmentItemModelID: 301, Count: 1, EnabledCount: 1, Enabled: true})
	if err != nil {
		t.Fatal(err)
	}
	constructor, err := serverData.EquipmentGroups.Add(&data.EquipmentGroup{CosmicObjectID: objectID, Title: "Constructor", EquipmentItemModelID: 501, Count: 1, EnabledCount: 1, Enabled: true})
	if err != nil {
		t.Fatal(err)
	}
	if _, err := serverData.ItemGroups.Add(&data.ItemGroup{ContainerEquipmentGroupID: materialContainer.ID, ContentItemModelID: 303, Count: 4}); err != nil {
		t.Fatal(err)
	}
	if err := gameWorld.ApplyControlPanelConstructorProduceItem(1, "session-1", 15, world.ControlPanelConstructorProduceItem{
		ConstructorEquipmentGroupID:       constructor.ID,
		MaterialContainerEquipmentGroupID: materialContainer.ID,
		ProductContainerEquipmentGroupID:  productContainer.ID,
		SchemaID:                          1,
		Amount:                            1,
	}); err != nil {
		t.Fatalf("constructor production returned error: %v", err)
	}

	gameWorld.Tick(1)

	cosmicObject := serverData.CosmicObjects.Items[objectID]
	if !constructor.Active {
		t.Fatalf("constructor must be active while production is running: %+v", constructor)
	}
	if cosmicObject.ConsumingPower != 7000 {
		t.Fatalf("consuming power = %v, want 7000", cosmicObject.ConsumingPower)
	}
}

// Р В РЎСџР РЋР вЂљР В РЎвЂўР В Р вЂ Р В Р’ВµР РЋР вЂљР РЋР РЏР В Р’ВµР РЋРІР‚С™, Р РЋРІР‚РЋР РЋРІР‚С™Р В РЎвЂў Р РЋР С“Р В РЎвЂўР РЋРІР‚В¦Р РЋР вЂљР В Р’В°Р В Р вЂ¦Р В Р’ВµР В Р вЂ¦Р В РЎвЂР В Р’Вµ Р В РўвЂР В Р’В°Р В Р вЂ¦Р В Р вЂ¦Р РЋРІР‚в„–Р РЋРІР‚В¦ Р В Р’В·Р В Р’В°Р В РЎвЂ”Р В РЎвЂР РЋР С“Р РЋРІР‚в„–Р В Р вЂ Р В Р’В°Р В Р’ВµР РЋРІР‚С™ Р В РЎвЂўР В Р’В±Р В Р вЂ¦Р В РЎвЂўР В Р вЂ Р В Р’В»Р РЋРІР‚ВР В Р вЂ¦Р В Р вЂ¦Р В РЎвЂўР В Р’Вµ Р В РЎвЂ”Р В РЎвЂўР В Р’В»Р В РЎвЂўР В Р’В¶Р В Р’ВµР В Р вЂ¦Р В РЎвЂР В Р’Вµ Р В РЎвЂќР В РЎвЂўР РЋР С“Р В РЎВР В РЎвЂР РЋРІР‚РЋР В Р’ВµР РЋР С“Р В РЎвЂќР В РЎвЂўР В РЎвЂ“Р В РЎвЂў Р В РЎвЂўР В Р’В±Р РЋР вЂ°Р В Р’ВµР В РЎвЂќР РЋРІР‚С™Р В Р’В°.
func TestApplyControlPanelConstructorQueueCommandSkipsNotStartedBatches(t *testing.T) {
	serverData := testWorldData(t)
	serverData.ItemTypes.Items[19] = &data.ItemType{ID: 19, TitleRu: "Constructor", TitleEn: "Constructor", Acronym: "Constructor", CountMustBeInteger: true}
	serverData.ItemModels.Items[501] = &data.ItemModel{ID: 501, TitleRu: "Constructor", TitleEn: "Constructor", Acronym: "Constructor", ItemTypeID: 19}
	if err := serverData.ItemTypes.RebuildIndexes(); err != nil {
		t.Fatal(err)
	}
	if err := serverData.ItemModels.RebuildIndexes(); err != nil {
		t.Fatal(err)
	}
	serverData.Schemas = storage.NewRawReferenceTable()
	serverData.SchemaComponents = storage.NewRawReferenceTable()
	addRawReferenceItem(t, serverData.Schemas, 1, map[string]any{"ID": 1, "ItemModelID": 302, "Count": 1, "ProductionEnergy": 10})
	addRawReferenceItem(t, serverData.SchemaComponents, 1, map[string]any{"ID": 1, "SchemaID": 1, "ComponentItemModelID": 303, "Count": 4})
	gameWorld := world.New(1, serverData)

	objectID, ok := gameWorld.ConnectAccount(1)
	if !ok {
		t.Fatalf("account was not connected")
	}
	var materialContainer *data.EquipmentGroup
	for _, group := range serverData.EquipmentGroups.GetByCosmicObjectID(objectID) {
		if group.EquipmentItemModelID == 301 {
			materialContainer = group
			break
		}
	}
	if materialContainer == nil {
		t.Fatalf("material container was not installed")
	}
	productContainer, err := serverData.EquipmentGroups.Add(&data.EquipmentGroup{CosmicObjectID: objectID, Title: "Product Container", EquipmentItemModelID: 301, Count: 1, EnabledCount: 1, Enabled: true})
	if err != nil {
		t.Fatal(err)
	}
	constructor, err := serverData.EquipmentGroups.Add(&data.EquipmentGroup{CosmicObjectID: objectID, Title: "Constructor", EquipmentItemModelID: 501, Count: 1, EnabledCount: 1, Enabled: true})
	if err != nil {
		t.Fatal(err)
	}
	if _, err := serverData.ItemGroups.Add(&data.ItemGroup{ContainerEquipmentGroupID: materialContainer.ID, ContentItemModelID: 303, Count: 12}); err != nil {
		t.Fatal(err)
	}
	if err := gameWorld.ApplyControlPanelConstructorProduceItem(1, "session-1", 16, world.ControlPanelConstructorProduceItem{
		ConstructorEquipmentGroupID:       constructor.ID,
		MaterialContainerEquipmentGroupID: materialContainer.ID,
		ProductContainerEquipmentGroupID:  productContainer.ID,
		SchemaID:                          1,
		Amount:                            3,
	}); err != nil {
		t.Fatalf("constructor production returned error: %v", err)
	}
	gameWorld.Tick(1)
	queued := gameWorld.SnapshotForAccount(1).ConstructorProductionJobs
	if len(queued) != 1 {
		t.Fatalf("production was not queued: %+v", queued)
	}

	if err := gameWorld.ApplyControlPanelConstructorQueueCommand(1, "session-1", 17, world.ControlPanelConstructorQueueCommand{
		ConstructorEquipmentGroupID: constructor.ID,
		JobID:                       queued[0].ID,
		Command:                     "skipNext",
	}); err != nil {
		t.Fatalf("constructor queue command returned error: %v", err)
	}

	queued = gameWorld.SnapshotForAccount(1).ConstructorProductionJobs
	if len(queued) != 1 || queued[0].RemainingCount != 1 || queued[0].TotalCount != 1 {
		t.Fatalf("not started batches were not removed: %+v", queued)
	}
	gameWorld.Tick(9)
	if queued := gameWorld.SnapshotForAccount(1).ConstructorProductionJobs; len(queued) != 0 {
		t.Fatalf("finished kept batch stayed in queue: %+v", queued)
	}
	productCounts := map[int64]float64{}
	for _, item := range serverData.ItemGroups.GetByContainerEquipmentGroupID(productContainer.ID) {
		productCounts[item.ContentItemModelID] = item.Count
	}
	if productCounts[302] != 1 {
		t.Fatalf("unexpected product count after skip next: %+v", productCounts)
	}
}

// Р В РЎСџР РЋР вЂљР В РЎвЂўР В Р вЂ Р В Р’ВµР РЋР вЂљР РЋР РЏР В Р’ВµР РЋРІР‚С™, Р РЋРІР‚РЋР РЋРІР‚С™Р В РЎвЂў Р РЋР С“Р В РЎвЂўР РЋРІР‚В¦Р РЋР вЂљР В Р’В°Р В Р вЂ¦Р В Р’ВµР В Р вЂ¦Р В РЎвЂР В Р’Вµ Р В РўвЂР В Р’В°Р В Р вЂ¦Р В Р вЂ¦Р РЋРІР‚в„–Р РЋРІР‚В¦ Р В Р’В·Р В Р’В°Р В РЎвЂ”Р В РЎвЂР РЋР С“Р РЋРІР‚в„–Р В Р вЂ Р В Р’В°Р В Р’ВµР РЋРІР‚С™ Р В РЎвЂўР В Р’В±Р В Р вЂ¦Р В РЎвЂўР В Р вЂ Р В Р’В»Р РЋРІР‚ВР В Р вЂ¦Р В Р вЂ¦Р В РЎвЂўР В Р’Вµ Р В РЎвЂ”Р В РЎвЂўР В Р’В»Р В РЎвЂўР В Р’В¶Р В Р’ВµР В Р вЂ¦Р В РЎвЂР В Р’Вµ Р В РЎвЂќР В РЎвЂўР РЋР С“Р В РЎВР В РЎвЂР РЋРІР‚РЋР В Р’ВµР РЋР С“Р В РЎвЂќР В РЎвЂўР В РЎвЂ“Р В РЎвЂў Р В РЎвЂўР В Р’В±Р РЋР вЂ°Р В Р’ВµР В РЎвЂќР РЋРІР‚С™Р В Р’В°.
func TestApplyControlPanelConstructorQueueCommandCancelsSelectedAndFollowingMainJobs(t *testing.T) {
	serverData := testWorldData(t)
	serverData.ItemTypes.Items[19] = &data.ItemType{ID: 19, TitleRu: "Constructor", TitleEn: "Constructor", Acronym: "Constructor", CountMustBeInteger: true}
	serverData.ItemModels.Items[501] = &data.ItemModel{ID: 501, TitleRu: "Constructor", TitleEn: "Constructor", Acronym: "Constructor", ItemTypeID: 19}
	if err := serverData.ItemTypes.RebuildIndexes(); err != nil {
		t.Fatal(err)
	}
	if err := serverData.ItemModels.RebuildIndexes(); err != nil {
		t.Fatal(err)
	}
	serverData.Schemas = storage.NewRawReferenceTable()
	serverData.SchemaComponents = storage.NewRawReferenceTable()
	addRawReferenceItem(t, serverData.Schemas, 1, map[string]any{"ID": 1, "ItemModelID": 302, "Count": 1, "ProductionEnergy": 10})
	addRawReferenceItem(t, serverData.SchemaComponents, 1, map[string]any{"ID": 1, "SchemaID": 1, "ComponentItemModelID": 303, "Count": 1})
	gameWorld := world.New(1, serverData)

	objectID, ok := gameWorld.ConnectAccount(1)
	if !ok {
		t.Fatalf("account was not connected")
	}
	var materialContainer *data.EquipmentGroup
	for _, group := range serverData.EquipmentGroups.GetByCosmicObjectID(objectID) {
		if group.EquipmentItemModelID == 301 {
			materialContainer = group
			break
		}
	}
	if materialContainer == nil {
		t.Fatalf("material container was not installed")
	}
	productContainer, err := serverData.EquipmentGroups.Add(&data.EquipmentGroup{CosmicObjectID: objectID, Title: "Product Container", EquipmentItemModelID: 301, Count: 1, EnabledCount: 1, Enabled: true})
	if err != nil {
		t.Fatal(err)
	}
	constructor, err := serverData.EquipmentGroups.Add(&data.EquipmentGroup{CosmicObjectID: objectID, Title: "Constructor", EquipmentItemModelID: 501, Count: 1, EnabledCount: 1, Enabled: true})
	if err != nil {
		t.Fatal(err)
	}
	if _, err := serverData.ItemGroups.Add(&data.ItemGroup{ContainerEquipmentGroupID: materialContainer.ID, ContentItemModelID: 303, Count: 4}); err != nil {
		t.Fatal(err)
	}
	for mutationSeq := int64(18); mutationSeq <= 19; mutationSeq++ {
		if err := gameWorld.ApplyControlPanelConstructorProduceItem(1, "session-1", mutationSeq, world.ControlPanelConstructorProduceItem{
			ConstructorEquipmentGroupID:       constructor.ID,
			MaterialContainerEquipmentGroupID: materialContainer.ID,
			ProductContainerEquipmentGroupID:  productContainer.ID,
			SchemaID:                          1,
			Amount:                            1,
		}); err != nil {
			t.Fatalf("constructor production returned error: %v", err)
		}
	}
	queued := gameWorld.SnapshotForAccount(1).ConstructorProductionJobs
	if len(queued) != 1 {
		t.Fatalf("productions were not queued: %+v", queued)
	}

	if err := gameWorld.ApplyControlPanelConstructorQueueCommand(1, "session-1", 20, world.ControlPanelConstructorQueueCommand{
		ConstructorEquipmentGroupID: constructor.ID,
		JobID:                       queued[0].ID,
		Command:                     "cancelAll",
	}); err != nil {
		t.Fatalf("constructor queue command returned error: %v", err)
	}
	if queued := gameWorld.SnapshotForAccount(1).ConstructorProductionJobs; len(queued) != 0 {
		t.Fatalf("selected and following main jobs stayed in queue: %+v", queued)
	}
}

func TestApplyControlPanelConstructorProduceItemQueuesOnlyMissingAuxiliaryComponents(t *testing.T) {
	serverData := testWorldData(t)
	serverData.ItemTypes.Items[19] = &data.ItemType{ID: 19, TitleRu: "Constructor", TitleEn: "Constructor", Acronym: "Constructor", CountMustBeInteger: true}
	serverData.ItemModels.Items[501] = &data.ItemModel{ID: 501, TitleRu: "Constructor", TitleEn: "Constructor", Acronym: "Constructor", ItemTypeID: 19}
	if err := serverData.ItemTypes.RebuildIndexes(); err != nil {
		t.Fatal(err)
	}
	if err := serverData.ItemModels.RebuildIndexes(); err != nil {
		t.Fatal(err)
	}
	serverData.Schemas = storage.NewRawReferenceTable()
	serverData.SchemaComponents = storage.NewRawReferenceTable()
	addRawReferenceItem(t, serverData.Schemas, 1, map[string]any{"ID": 1, "ItemModelID": 302, "Count": 1, "ProductionEnergy": 10})
	addRawReferenceItem(t, serverData.Schemas, 2, map[string]any{"ID": 2, "ItemModelID": 303, "Count": 3, "ProductionEnergy": 4})
	addRawReferenceItem(t, serverData.SchemaComponents, 1, map[string]any{"ID": 1, "SchemaID": 1, "ComponentItemModelID": 303, "Count": 5})
	addRawReferenceItem(t, serverData.SchemaComponents, 2, map[string]any{"ID": 2, "SchemaID": 2, "ComponentItemModelID": 403, "Count": 2})
	gameWorld := world.New(1, serverData)

	objectID, ok := gameWorld.ConnectAccount(1)
	if !ok {
		t.Fatalf("account was not connected")
	}
	var materialContainer *data.EquipmentGroup
	for _, group := range serverData.EquipmentGroups.GetByCosmicObjectID(objectID) {
		if group.EquipmentItemModelID == 301 {
			materialContainer = group
			break
		}
	}
	if materialContainer == nil {
		t.Fatalf("material container was not installed")
	}
	productContainer, err := serverData.EquipmentGroups.Add(&data.EquipmentGroup{CosmicObjectID: objectID, Title: "Product Container", EquipmentItemModelID: 301, Count: 1, EnabledCount: 1, Enabled: true})
	if err != nil {
		t.Fatal(err)
	}
	constructor, err := serverData.EquipmentGroups.Add(&data.EquipmentGroup{CosmicObjectID: objectID, Title: "Constructor", EquipmentItemModelID: 501, Count: 1, EnabledCount: 1, Enabled: true})
	if err != nil {
		t.Fatal(err)
	}
	if _, err := serverData.ItemGroups.Add(&data.ItemGroup{ContainerEquipmentGroupID: materialContainer.ID, ContentItemModelID: 303, Count: 4}); err != nil {
		t.Fatal(err)
	}
	if _, err := serverData.ItemGroups.Add(&data.ItemGroup{ContainerEquipmentGroupID: materialContainer.ID, ContentItemModelID: 403, Count: 4}); err != nil {
		t.Fatal(err)
	}

	if err := gameWorld.ApplyControlPanelConstructorProduceItem(1, "session-1", 12, world.ControlPanelConstructorProduceItem{
		ConstructorEquipmentGroupID:       constructor.ID,
		MaterialContainerEquipmentGroupID: materialContainer.ID,
		ProductContainerEquipmentGroupID:  productContainer.ID,
		SchemaID:                          1,
		Amount:                            2,
	}); err != nil {
		t.Fatalf("constructor production returned error: %v", err)
	}

	queued := gameWorld.SnapshotForAccount(1).ConstructorProductionJobs
	if len(queued) != 2 || queued[0].QueueType != "auxiliary" || queued[0].SchemaID != 2 || queued[0].RemainingCount != 6 || queued[0].TotalCount != 6 || queued[0].ParentJobID != queued[1].ID || queued[1].QueueType != "main" || queued[1].RemainingCount != 2 || queued[1].TotalCount != 2 {
		t.Fatalf("auxiliary and main queues were not planned correctly: %+v", queued)
	}
	gameWorld.Tick(4)
	gameWorld.Tick(4)
	gameWorld.Tick(10)
	gameWorld.Tick(10)

	materialCounts := map[int64]float64{}
	for _, item := range serverData.ItemGroups.GetByContainerEquipmentGroupID(materialContainer.ID) {
		materialCounts[item.ContentItemModelID] = item.Count
	}
	if materialCounts[303] != 0 || materialCounts[403] != 0 {
		t.Fatalf("auxiliary production did not leave expected material remainder: %+v", materialCounts)
	}
	productCounts := map[int64]float64{}
	for _, item := range serverData.ItemGroups.GetByContainerEquipmentGroupID(productContainer.ID) {
		productCounts[item.ContentItemModelID] = item.Count
	}
	if productCounts[302] != 2 {
		t.Fatalf("main production did not complete after auxiliary production: %+v", productCounts)
	}
}

// Р В РЎСџР РЋР вЂљР В РЎвЂўР В Р вЂ Р В Р’ВµР РЋР вЂљР РЋР РЏР В Р’ВµР РЋРІР‚С™, Р РЋРІР‚РЋР РЋРІР‚С™Р В РЎвЂў Р РЋР С“Р В РЎвЂўР РЋРІР‚В¦Р РЋР вЂљР В Р’В°Р В Р вЂ¦Р В Р’ВµР В Р вЂ¦Р В РЎвЂР В Р’Вµ Р В РўвЂР В Р’В°Р В Р вЂ¦Р В Р вЂ¦Р РЋРІР‚в„–Р РЋРІР‚В¦ Р В Р’В·Р В Р’В°Р В РЎвЂ”Р В РЎвЂР РЋР С“Р РЋРІР‚в„–Р В Р вЂ Р В Р’В°Р В Р’ВµР РЋРІР‚С™ Р В РЎвЂўР В Р’В±Р В Р вЂ¦Р В РЎвЂўР В Р вЂ Р В Р’В»Р РЋРІР‚ВР В Р вЂ¦Р В Р вЂ¦Р В РЎвЂўР В Р’Вµ Р В РЎвЂ”Р В РЎвЂўР В Р’В»Р В РЎвЂўР В Р’В¶Р В Р’ВµР В Р вЂ¦Р В РЎвЂР В Р’Вµ Р В РЎвЂќР В РЎвЂўР РЋР С“Р В РЎВР В РЎвЂР РЋРІР‚РЋР В Р’ВµР РЋР С“Р В РЎвЂќР В РЎвЂўР В РЎвЂ“Р В РЎвЂў Р В РЎвЂўР В Р’В±Р РЋР вЂ°Р В Р’ВµР В РЎвЂќР РЋРІР‚С™Р В Р’В°.
func TestApplyControlPanelConstructorProduceItemCreatesBlueprintObjectInFrontOfBuilder(t *testing.T) {
	serverData := testWorldData(t)
	serverData.ItemTypes.Items[19] = &data.ItemType{ID: 19, TitleRu: "Constructor", TitleEn: "Constructor", Acronym: "Constructor", CountMustBeInteger: true}
	serverData.ItemModels.Items[501] = &data.ItemModel{ID: 501, TitleRu: "Constructor", TitleEn: "Constructor", Acronym: "Constructor", ItemTypeID: 19}
	if err := serverData.ItemTypes.RebuildIndexes(); err != nil {
		t.Fatal(err)
	}
	if err := serverData.ItemModels.RebuildIndexes(); err != nil {
		t.Fatal(err)
	}
	serverData.Blueprints = storage.NewRawReferenceTable()
	serverData.BlueprintComponents = storage.NewRawReferenceTable()
	addRawReferenceItem(t, serverData.Blueprints, 1, map[string]any{"ID": 1, "CosmicObjectModelID": 4, "ProductionEnergy": 10})
	addRawReferenceItem(t, serverData.BlueprintComponents, 1, map[string]any{"ID": 1, "BlueprintID": 1, "ComponentItemModelID": 303, "Count": 4})
	gameWorld := world.New(1, serverData)

	objectID, ok := gameWorld.ConnectAccount(1)
	if !ok {
		t.Fatalf("account was not connected")
	}
	builder := serverData.CosmicObjects.Items[objectID]
	builder.X = 10
	builder.Y = 20
	builder.Rotation = 0
	builder.TargetRotation = 0
	var materialContainer *data.EquipmentGroup
	for _, group := range serverData.EquipmentGroups.GetByCosmicObjectID(objectID) {
		if group.EquipmentItemModelID == 301 {
			materialContainer = group
			break
		}
	}
	if materialContainer == nil {
		t.Fatalf("material container was not installed")
	}
	constructor, err := serverData.EquipmentGroups.Add(&data.EquipmentGroup{CosmicObjectID: objectID, Title: "Constructor", EquipmentItemModelID: 501, Count: 1, EnabledCount: 1, Enabled: true})
	if err != nil {
		t.Fatal(err)
	}
	if _, err := serverData.ItemGroups.Add(&data.ItemGroup{ContainerEquipmentGroupID: materialContainer.ID, ContentItemModelID: 303, Count: 4}); err != nil {
		t.Fatal(err)
	}

	if err := gameWorld.ApplyControlPanelConstructorProduceItem(1, "session-1", 13, world.ControlPanelConstructorProduceItem{
		ConstructorEquipmentGroupID:       constructor.ID,
		MaterialContainerEquipmentGroupID: materialContainer.ID,
		BlueprintID:                       1,
		Amount:                            1,
	}); err != nil {
		t.Fatalf("constructor production returned error: %v", err)
	}
	if len(serverData.CosmicObjects.GetByCosmicObjectModelID(4)) != 0 {
		t.Fatalf("object was created before production time passed")
	}

	gameWorld.Tick(10)

	createdObjects := serverData.CosmicObjects.GetByCosmicObjectModelID(4)
	if len(createdObjects) != 1 {
		t.Fatalf("created object count = %d, want 1", len(createdObjects))
	}
	created := createdObjects[0]
	if created.X != 10 || created.Y <= 20 || created.Rotation != builder.Rotation || created.TargetRotation != builder.Rotation {
		t.Fatalf("created object was not placed in front of builder: %+v", created)
	}
	if created.OwnerCharacterID != builder.OwnerCharacterID || created.CreatorCharacterID != builder.OwnerCharacterID {
		t.Fatalf("created object did not inherit character ownership: %+v", created)
	}
	if len(serverData.ItemGroups.GetByContainerEquipmentGroupID(materialContainer.ID)) != 0 {
		t.Fatalf("blueprint components were not consumed: %+v", serverData.ItemGroups.GetByContainerEquipmentGroupID(materialContainer.ID))
	}
}

// Р В РЎСџР РЋР вЂљР В РЎвЂўР В Р вЂ Р В Р’ВµР РЋР вЂљР РЋР РЏР В Р’ВµР РЋРІР‚С™, Р РЋРІР‚РЋР РЋРІР‚С™Р В РЎвЂў Р РЋР С“Р В РЎвЂўР РЋРІР‚В¦Р РЋР вЂљР В Р’В°Р В Р вЂ¦Р В Р’ВµР В Р вЂ¦Р В РЎвЂР В Р’Вµ Р В РўвЂР В Р’В°Р В Р вЂ¦Р В Р вЂ¦Р РЋРІР‚в„–Р РЋРІР‚В¦ Р В Р’В·Р В Р’В°Р В РЎвЂ”Р В РЎвЂР РЋР С“Р РЋРІР‚в„–Р В Р вЂ Р В Р’В°Р В Р’ВµР РЋРІР‚С™ Р В РЎвЂўР В Р’В±Р В Р вЂ¦Р В РЎвЂўР В Р вЂ Р В Р’В»Р РЋРІР‚ВР В Р вЂ¦Р В Р вЂ¦Р В РЎвЂўР В Р’Вµ Р В РЎвЂ”Р В РЎвЂўР В Р’В»Р В РЎвЂўР В Р’В¶Р В Р’ВµР В Р вЂ¦Р В РЎвЂР В Р’Вµ Р В РЎвЂќР В РЎвЂўР РЋР С“Р В РЎВР В РЎвЂР РЋРІР‚РЋР В Р’ВµР РЋР С“Р В РЎвЂќР В РЎвЂўР В РЎвЂ“Р В РЎвЂў Р В РЎвЂўР В Р’В±Р РЋР вЂ°Р В Р’ВµР В РЎвЂќР РЋРІР‚С™Р В Р’В°.
func TestApplyControlPanelConstructorProduceItemPlacesBlueprintObjectPastOccupiedSpot(t *testing.T) {
	serverData := testWorldData(t)
	serverData.ItemTypes.Items[19] = &data.ItemType{ID: 19, TitleRu: "Constructor", TitleEn: "Constructor", Acronym: "Constructor", CountMustBeInteger: true}
	serverData.ItemModels.Items[501] = &data.ItemModel{ID: 501, TitleRu: "Constructor", TitleEn: "Constructor", Acronym: "Constructor", ItemTypeID: 19}
	if err := serverData.ItemTypes.RebuildIndexes(); err != nil {
		t.Fatal(err)
	}
	if err := serverData.ItemModels.RebuildIndexes(); err != nil {
		t.Fatal(err)
	}
	serverData.Blueprints = storage.NewRawReferenceTable()
	serverData.BlueprintComponents = storage.NewRawReferenceTable()
	addRawReferenceItem(t, serverData.Blueprints, 1, map[string]any{"ID": 1, "CosmicObjectModelID": 4, "ProductionEnergy": 10})
	addRawReferenceItem(t, serverData.BlueprintComponents, 1, map[string]any{"ID": 1, "BlueprintID": 1, "ComponentItemModelID": 303, "Count": 4})
	gameWorld := world.New(1, serverData)

	objectID, ok := gameWorld.ConnectAccount(1)
	if !ok {
		t.Fatalf("account was not connected")
	}
	builder := serverData.CosmicObjects.Items[objectID]
	builder.X = 0
	builder.Y = 0
	builder.Rotation = 0
	builder.TargetRotation = 0
	var materialContainer *data.EquipmentGroup
	for _, group := range serverData.EquipmentGroups.GetByCosmicObjectID(objectID) {
		if group.EquipmentItemModelID == 301 {
			materialContainer = group
			break
		}
	}
	if materialContainer == nil {
		t.Fatalf("material container was not installed")
	}
	constructor, err := serverData.EquipmentGroups.Add(&data.EquipmentGroup{CosmicObjectID: objectID, Title: "Constructor", EquipmentItemModelID: 501, Count: 1, EnabledCount: 1, Enabled: true})
	if err != nil {
		t.Fatal(err)
	}
	if _, err := serverData.ItemGroups.Add(&data.ItemGroup{ContainerEquipmentGroupID: materialContainer.ID, ContentItemModelID: 303, Count: 4}); err != nil {
		t.Fatal(err)
	}
	blocker, err := serverData.CosmicObjects.Add(&data.CosmicObject{Title: "Blocker", CosmicObjectModelID: 4, Mass: 12, MaxArmor: 150, Armor: 150, X: 0, Y: 18.875, Rotation: 0, TargetRotation: 0, Enabled: true, Anchored: true})
	if err != nil {
		t.Fatal(err)
	}

	if err := gameWorld.ApplyControlPanelConstructorProduceItem(1, "session-1", 14, world.ControlPanelConstructorProduceItem{
		ConstructorEquipmentGroupID:       constructor.ID,
		MaterialContainerEquipmentGroupID: materialContainer.ID,
		BlueprintID:                       1,
		Amount:                            1,
	}); err != nil {
		t.Fatalf("constructor production returned error: %v", err)
	}
	gameWorld.Tick(10)

	createdObjects := serverData.CosmicObjects.GetByCosmicObjectModelID(4)
	if len(createdObjects) != 2 {
		t.Fatalf("object count with blocker = %d, want 2", len(createdObjects))
	}
	var created *data.CosmicObject
	for _, object := range createdObjects {
		if object.ID != blocker.ID {
			created = object
		}
	}
	if created == nil || created.Y <= blocker.Y {
		t.Fatalf("created object was not moved past occupied spot: blocker=%+v created=%+v", blocker, created)
	}
	if _, collided := physics.CollisionInfo(*created, *serverData.CosmicObjectModels.Items[4], *blocker, *serverData.CosmicObjectModels.Items[4]); collided {
		t.Fatalf("created object still intersects blocker: blocker=%+v created=%+v", blocker, created)
	}
}

func TestApplyControlPanelFuelTransferFillsObjectFuelFromContainer(t *testing.T) {
	serverData := testWorldData(t)
	serverData.ItemTypes.Items[10] = &data.ItemType{ID: 10, TitleRu: "Fuel Tank", TitleEn: "Fuel Tank", Acronym: "FuelTank", IsInternalUsable: true, CountMustBeInteger: true}
	serverData.ItemModels.Items[7] = &data.ItemModel{ID: 7, TitleRu: "Fuel", TitleEn: "Fuel", Acronym: "Fuel", ItemTypeID: 7, Mass: 2}
	serverData.ItemModels.Items[304] = &data.ItemModel{ID: 304, TitleRu: "Fuel Tank", TitleEn: "Fuel Tank", Acronym: "FuelTank", ItemTypeID: 10, Capacity: 100, ConsumingItemModelID: 7}
	if err := serverData.ItemTypes.RebuildIndexes(); err != nil {
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
	addTestRobots(t, serverData, objectID)
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

	if cosmicObject.Fuel != 20 {
		t.Fatalf("object fuel changed before fueling task completed: %v", cosmicObject.Fuel)
	}
	items := serverData.ItemGroups.GetByContainerEquipmentGroupID(container.ID)
	if len(items) != 1 || items[0].Count != 40 {
		t.Fatalf("container fuel changed before fueling task started: %+v", items)
	}
	tasks := serverData.Tasks.GetByControllerEquipmentGroupID(fuelTank.ID)
	if len(tasks) != 1 || !tasks[0].LeftToRightDirection || tasks[0].BatchCount != 1 {
		t.Fatalf("fueling task was not queued correctly: %+v", tasks)
	}

	gameWorld.Tick(1)
	if reserves := serverData.TaskItemGroups.GetByTaskID(tasks[0].ID); len(reserves) != 1 || !reserves[0].IsStored || reserves[0].Count != 30 {
		t.Fatalf("fueling task did not store fuel: %+v", reserves)
	}
	items = serverData.ItemGroups.GetByContainerEquipmentGroupID(container.ID)
	if len(items) != 1 || items[0].Count != 10 {
		t.Fatalf("container fuel item was not reduced after task start: %+v", items)
	}

	gameWorld.Tick(10)
	if cosmicObject.Fuel != 50 {
		t.Fatalf("object fuel = %v, want 50", cosmicObject.Fuel)
	}
	if tasks := serverData.Tasks.GetByControllerEquipmentGroupID(fuelTank.ID); len(tasks) != 0 {
		t.Fatalf("completed fueling task stayed in queue: %+v", tasks)
	}
}

// Р В РЎСџР РЋР вЂљР В РЎвЂўР В Р вЂ Р В Р’ВµР РЋР вЂљР РЋР РЏР В Р’ВµР РЋРІР‚С™, Р РЋРІР‚РЋР РЋРІР‚С™Р В РЎвЂў Р РЋР С“Р В РЎвЂўР РЋРІР‚В¦Р РЋР вЂљР В Р’В°Р В Р вЂ¦Р В Р’ВµР В Р вЂ¦Р В РЎвЂР В Р’Вµ Р В РўвЂР В Р’В°Р В Р вЂ¦Р В Р вЂ¦Р РЋРІР‚в„–Р РЋРІР‚В¦ Р В Р’В·Р В Р’В°Р В РЎвЂ”Р В РЎвЂР РЋР С“Р РЋРІР‚в„–Р В Р вЂ Р В Р’В°Р В Р’ВµР РЋРІР‚С™ Р В РЎвЂўР В Р’В±Р В Р вЂ¦Р В РЎвЂўР В Р вЂ Р В Р’В»Р РЋРІР‚ВР В Р вЂ¦Р В Р вЂ¦Р В РЎвЂўР В Р’Вµ Р В РЎвЂ”Р В РЎвЂўР В Р’В»Р В РЎвЂўР В Р’В¶Р В Р’ВµР В Р вЂ¦Р В РЎвЂР В Р’Вµ Р В РЎвЂќР В РЎвЂўР РЋР С“Р В РЎВР В РЎвЂР РЋРІР‚РЋР В Р’ВµР РЋР С“Р В РЎвЂќР В РЎвЂўР В РЎвЂ“Р В РЎвЂў Р В РЎвЂўР В Р’В±Р РЋР вЂ°Р В Р’ВµР В РЎвЂќР РЋРІР‚С™Р В Р’В°.
// Р В РЎСџР РЋР вЂљР В РЎвЂўР В Р вЂ Р В Р’ВµР РЋР вЂљР РЋР РЏР В Р’ВµР РЋРІР‚С™, Р РЋРІР‚РЋР РЋРІР‚С™Р В РЎвЂў Р РЋР С“Р В РЎвЂўР РЋРІР‚В¦Р РЋР вЂљР В Р’В°Р В Р вЂ¦Р В Р’ВµР В Р вЂ¦Р В РЎвЂР В Р’Вµ Р В РўвЂР В Р’В°Р В Р вЂ¦Р В Р вЂ¦Р РЋРІР‚в„–Р РЋРІР‚В¦ Р В Р’В·Р В Р’В°Р В РЎвЂ”Р В РЎвЂР РЋР С“Р РЋРІР‚в„–Р В Р вЂ Р В Р’В°Р В Р’ВµР РЋРІР‚С™ Р В РЎвЂўР В Р’В±Р В Р вЂ¦Р В РЎвЂўР В Р вЂ Р В Р’В»Р РЋРІР‚ВР В Р вЂ¦Р В Р вЂ¦Р В РЎвЂўР В Р’Вµ Р В РЎвЂ”Р В РЎвЂўР В Р’В»Р В РЎвЂўР В Р’В¶Р В Р’ВµР В Р вЂ¦Р В РЎвЂР В Р’Вµ Р В РЎвЂќР В РЎвЂўР РЋР С“Р В РЎВР В РЎвЂР РЋРІР‚РЋР В Р’ВµР РЋР С“Р В РЎвЂќР В РЎвЂўР В РЎвЂ“Р В РЎвЂў Р В РЎвЂўР В Р’В±Р РЋР вЂ°Р В Р’ВµР В РЎвЂќР РЋРІР‚С™Р В Р’В°.
// Р В РЎСџР РЋР вЂљР В РЎвЂўР В Р вЂ Р В Р’ВµР РЋР вЂљР РЋР РЏР В Р’ВµР РЋРІР‚С™, Р РЋРІР‚РЋР РЋРІР‚С™Р В РЎвЂў Р РЋР С“Р В РЎвЂўР РЋРІР‚В¦Р РЋР вЂљР В Р’В°Р В Р вЂ¦Р В Р’ВµР В Р вЂ¦Р В РЎвЂР В Р’Вµ Р В РўвЂР В Р’В°Р В Р вЂ¦Р В Р вЂ¦Р РЋРІР‚в„–Р РЋРІР‚В¦ Р В Р’В·Р В Р’В°Р В РЎвЂ”Р В РЎвЂР РЋР С“Р РЋРІР‚в„–Р В Р вЂ Р В Р’В°Р В Р’ВµР РЋРІР‚С™ Р В РЎвЂўР В Р’В±Р В Р вЂ¦Р В РЎвЂўР В Р вЂ Р В Р’В»Р РЋРІР‚ВР В Р вЂ¦Р В Р вЂ¦Р В РЎвЂўР В Р’Вµ Р В РЎвЂ”Р В РЎвЂўР В Р’В»Р В РЎвЂўР В Р’В¶Р В Р’ВµР В Р вЂ¦Р В РЎвЂР В Р’Вµ Р В РЎвЂќР В РЎвЂўР РЋР С“Р В РЎВР В РЎвЂР РЋРІР‚РЋР В Р’ВµР РЋР С“Р В РЎвЂќР В РЎвЂўР В РЎвЂ“Р В РЎвЂў Р В РЎвЂўР В Р’В±Р РЋР вЂ°Р В Р’ВµР В РЎвЂќР РЋРІР‚С™Р В Р’В°.
func TestApplyControlPanelFuelTransferFillsOwnedClusterFuelTankObject(t *testing.T) {
	serverData := testWorldData(t)
	serverData.ItemTypes.Items[10] = &data.ItemType{ID: 10, TitleRu: "Fuel Tank", TitleEn: "Fuel Tank", Acronym: "FuelTank", IsInternalUsable: true, CountMustBeInteger: true}
	serverData.ItemModels.Items[7] = &data.ItemModel{ID: 7, TitleRu: "Fuel", TitleEn: "Fuel", Acronym: "Fuel", ItemTypeID: 7, Mass: 2}
	serverData.ItemModels.Items[304] = &data.ItemModel{ID: 304, TitleRu: "Fuel Tank", TitleEn: "Fuel Tank", Acronym: "FuelTank", ItemTypeID: 10, Capacity: 100, ConsumingItemModelID: 7}
	if err := serverData.ItemTypes.RebuildIndexes(); err != nil {
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
	mainObject := serverData.CosmicObjects.Items[objectID]
	mainObject.ClusterMainCosmicObjectID = mainObject.ID
	targetObject, err := serverData.CosmicObjects.Add(&data.CosmicObject{
		Title:                     "Cluster Fuel Ship",
		CosmicObjectModelID:       mainObject.CosmicObjectModelID,
		OwnerCharacterID:          mainObject.OwnerCharacterID,
		CreatorCharacterID:        mainObject.OwnerCharacterID,
		MaxFuel:                   50,
		Fuel:                      20,
		Enabled:                   true,
		ClusterMainCosmicObjectID: mainObject.ID,
	})
	if err != nil {
		t.Fatal(err)
	}
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
		CosmicObjectID:       targetObject.ID,
		Title:                "Cluster Fuel Tank",
		EquipmentItemModelID: 304,
		Count:                1,
		EnabledCount:         1,
		Enabled:              true,
	})
	if err != nil {
		t.Fatal(err)
	}
	addTestRobots(t, serverData, targetObject.ID)
	fuelItem, err := serverData.ItemGroups.Add(&data.ItemGroup{ContainerEquipmentGroupID: container.ID, ContentItemModelID: 7, Count: 40})
	if err != nil {
		t.Fatal(err)
	}

	if err := gameWorld.ApplyControlPanelFuelTransfer(1, "session-1", 17, world.ControlPanelFuelTransfer{
		ContainerEquipmentGroupID: container.ID,
		FuelTankEquipmentGroupID:  fuelTank.ID,
		ItemGroupIDs:              []int64{fuelItem.ID},
	}); err != nil {
		t.Fatalf("cluster fuel transfer returned error: %v", err)
	}

	gameWorld.Tick(10)
	if targetObject.Fuel != 50 || mainObject.Fuel == 50 {
		t.Fatalf("fuel after cluster transfer: target=%v main=%v", targetObject.Fuel, mainObject.Fuel)
	}
}

func TestApplyControlPanelFuelTransferFillsOnlyRequestedAmount(t *testing.T) {
	serverData := testWorldData(t)
	serverData.ItemTypes.Items[10] = &data.ItemType{ID: 10, TitleRu: "Fuel Tank", TitleEn: "Fuel Tank", Acronym: "FuelTank", IsInternalUsable: true, CountMustBeInteger: true}
	serverData.ItemModels.Items[7] = &data.ItemModel{ID: 7, TitleRu: "Fuel", TitleEn: "Fuel", Acronym: "Fuel", ItemTypeID: 7, Mass: 2}
	serverData.ItemModels.Items[304] = &data.ItemModel{ID: 304, TitleRu: "Fuel Tank", TitleEn: "Fuel Tank", Acronym: "FuelTank", ItemTypeID: 10, Capacity: 100, ConsumingItemModelID: 7}
	if err := serverData.ItemTypes.RebuildIndexes(); err != nil {
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
	addTestRobots(t, serverData, objectID)
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

	gameWorld.Tick(10)
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
	serverData.ItemTypes.Items[10] = &data.ItemType{ID: 10, TitleRu: "Fuel Tank", TitleEn: "Fuel Tank", Acronym: "FuelTank", IsInternalUsable: true, CountMustBeInteger: true}
	serverData.ItemModels.Items[7] = &data.ItemModel{ID: 7, TitleRu: "Fuel", TitleEn: "Fuel", Acronym: "Fuel", ItemTypeID: 7, Mass: 2}
	serverData.ItemModels.Items[304] = &data.ItemModel{ID: 304, TitleRu: "Fuel Tank", TitleEn: "Fuel Tank", Acronym: "FuelTank", ItemTypeID: 10, Capacity: 100, ConsumingItemModelID: 7}
	if err := serverData.ItemTypes.RebuildIndexes(); err != nil {
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
	addTestRobots(t, serverData, objectID)

	if err := gameWorld.ApplyControlPanelFuelTransfer(1, "session-1", 7, world.ControlPanelFuelTransfer{
		ContainerEquipmentGroupID: container.ID,
		FuelTankEquipmentGroupID:  fuelTank.ID,
		Amount:                    12,
	}); err != nil {
		t.Fatalf("fuel drain returned error: %v", err)
	}

	if cosmicObject.Fuel != 20 {
		t.Fatalf("object fuel changed before drain task started: %v", cosmicObject.Fuel)
	}
	tasks := serverData.Tasks.GetByControllerEquipmentGroupID(fuelTank.ID)
	if len(tasks) != 1 || tasks[0].LeftToRightDirection || tasks[0].BatchCount != 1 {
		t.Fatalf("fuel drain task was not queued correctly: %+v", tasks)
	}

	gameWorld.Tick(1)
	if reserves := serverData.TaskItemGroups.GetByTaskID(tasks[0].ID); len(reserves) != 1 || !reserves[0].IsStored || reserves[0].Count != 12 {
		t.Fatalf("fuel drain task did not store fuel: %+v", reserves)
	}
	if cosmicObject.Fuel != 8 {
		t.Fatalf("object fuel after drain start = %v, want 8", cosmicObject.Fuel)
	}

	gameWorld.Tick(10)
	if cosmicObject.Fuel != 8 {
		t.Fatalf("object fuel = %v, want 8", cosmicObject.Fuel)
	}
	items := serverData.ItemGroups.GetByContainerEquipmentGroupID(container.ID)
	if len(items) != 1 || items[0].ContentItemModelID != 7 || items[0].Count != 12 {
		t.Fatalf("container fuel item was not created: %+v", items)
	}
}

// Р В РЎСџР РЋР вЂљР В РЎвЂўР В Р вЂ Р В Р’ВµР РЋР вЂљР РЋР РЏР В Р’ВµР РЋРІР‚С™, Р РЋРІР‚РЋР РЋРІР‚С™Р В РЎвЂў Р РЋР С“Р В РЎвЂўР РЋРІР‚В¦Р РЋР вЂљР В Р’В°Р В Р вЂ¦Р В Р’ВµР В Р вЂ¦Р В РЎвЂР В Р’Вµ Р В РўвЂР В Р’В°Р В Р вЂ¦Р В Р вЂ¦Р РЋРІР‚в„–Р РЋРІР‚В¦ Р В Р’В·Р В Р’В°Р В РЎвЂ”Р В РЎвЂР РЋР С“Р РЋРІР‚в„–Р В Р вЂ Р В Р’В°Р В Р’ВµР РЋРІР‚С™ Р В РЎвЂўР В Р’В±Р В Р вЂ¦Р В РЎвЂўР В Р вЂ Р В Р’В»Р РЋРІР‚ВР В Р вЂ¦Р В Р вЂ¦Р В РЎвЂўР В Р’Вµ Р В РЎвЂ”Р В РЎвЂўР В Р’В»Р В РЎвЂўР В Р’В¶Р В Р’ВµР В Р вЂ¦Р В РЎвЂР В Р’Вµ Р В РЎвЂќР В РЎвЂўР РЋР С“Р В РЎВР В РЎвЂР РЋРІР‚РЋР В Р’ВµР РЋР С“Р В РЎвЂќР В РЎвЂўР В РЎвЂ“Р В РЎвЂў Р В РЎвЂўР В Р’В±Р РЋР вЂ°Р В Р’ВµР В РЎвЂќР РЋРІР‚С™Р В Р’В°.
func TestApplyControlPanelItemDeconstructionQueuesAndCompletesSchemaBatch(t *testing.T) {
	serverData := testWorldData(t)
	serverData.Schemas = storage.NewRawReferenceTable()
	serverData.SchemaComponents = storage.NewRawReferenceTable()
	addRawReferenceItem(t, serverData.Schemas, 1, map[string]any{"ID": 1, "ItemModelID": 302, "Count": 2, "ProductionEnergy": 300})
	addRawReferenceItem(t, serverData.Schemas, 2, map[string]any{"ID": 2, "ItemModelID": 302, "Count": 2, "ProductionEnergy": 100})
	addRawReferenceItem(t, serverData.SchemaComponents, 1, map[string]any{"ID": 1, "SchemaID": 1, "ComponentItemModelID": 303, "Count": 9})
	addRawReferenceItem(t, serverData.SchemaComponents, 2, map[string]any{"ID": 2, "SchemaID": 2, "ComponentItemModelID": 303, "Count": 4})
	gameWorld := world.New(1, serverData)

	objectID, ok := gameWorld.ConnectAccount(1)
	if !ok {
		t.Fatalf("account was not connected")
	}
	var sourceContainer *data.EquipmentGroup
	for _, group := range serverData.EquipmentGroups.GetByCosmicObjectID(objectID) {
		if group.EquipmentItemModelID == 301 && sourceContainer == nil {
			sourceContainer = group
		}
	}
	if sourceContainer == nil {
		t.Fatalf("source container was not installed")
	}
	deconstructor := addTestDeconstructor(t, serverData, objectID)
	addTestRobots(t, serverData, objectID)
	itemGroup, err := serverData.ItemGroups.Add(&data.ItemGroup{ContainerEquipmentGroupID: sourceContainer.ID, ContentItemModelID: 302, Count: 5})
	if err != nil {
		t.Fatal(err)
	}

	if err := gameWorld.ApplyControlPanelItemDeconstruction(1, "session-1", 40, world.ControlPanelItemDeconstruction{
		DeconstructorEquipmentGroupID:   deconstructor.ID,
		SourceContainerEquipmentGroupID: sourceContainer.ID,
		TargetContainerEquipmentGroupID: sourceContainer.ID,
		ItemGroupIDs:                    []int64{itemGroup.ID},
	}); err != nil {
		t.Fatalf("item deconstruction returned error: %v", err)
	}

	tasks := serverData.Tasks.GetByControllerEquipmentGroupID(deconstructor.ID)
	if len(tasks) != 1 || tasks[0].SchemaID != 2 || tasks[0].BatchCount != 2 || tasks[0].TotalEnergy != 200 {
		t.Fatalf("deconstruction task was not queued correctly: %+v", tasks)
	}
	reserves := serverData.TaskItemGroups.GetByTaskID(tasks[0].ID)
	if len(reserves) != 1 || reserves[0].ItemModelID != 302 || reserves[0].Count != 4 || reserves[0].IsStored {
		t.Fatalf("deconstruction reserve was not created correctly: %+v", reserves)
	}

	gameWorld.Tick(1)
	reserves = serverData.TaskItemGroups.GetByTaskID(tasks[0].ID)
	if len(reserves) != 1 || !reserves[0].IsStored {
		t.Fatalf("deconstruction task did not store source items: %+v", reserves)
	}
	sourceItems := serverData.ItemGroups.GetByContainerEquipmentGroupID(sourceContainer.ID)
	if len(sourceItems) != 1 || sourceItems[0].ContentItemModelID != 302 || sourceItems[0].Count != 1 {
		t.Fatalf("source container after deconstruction start: %+v", sourceItems)
	}

	gameWorld.Tick(10)
	targetItems := serverData.ItemGroups.GetByContainerEquipmentGroupID(sourceContainer.ID)
	if len(targetItems) != 2 || targetItems[0].ContentItemModelID != 302 || targetItems[0].Count != 1 || targetItems[1].ContentItemModelID != 303 || targetItems[1].Count != 8 {
		t.Fatalf("container after deconstruction completion: %+v", targetItems)
	}
	if tasks := serverData.Tasks.GetByControllerEquipmentGroupID(deconstructor.ID); len(tasks) != 0 {
		t.Fatalf("completed deconstruction task stayed in queue: %+v", tasks)
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

// Р В РЎСџР РЋР вЂљР В РЎвЂўР В Р вЂ Р В Р’ВµР РЋР вЂљР РЋР РЏР В Р’ВµР РЋРІР‚С™, Р РЋРІР‚РЋР РЋРІР‚С™Р В РЎвЂў Р РЋР С“Р В РЎвЂўР РЋРІР‚В¦Р РЋР вЂљР В Р’В°Р В Р вЂ¦Р В Р’ВµР В Р вЂ¦Р В РЎвЂР В Р’Вµ Р В РўвЂР В Р’В°Р В Р вЂ¦Р В Р вЂ¦Р РЋРІР‚в„–Р РЋРІР‚В¦ Р В Р’В·Р В Р’В°Р В РЎвЂ”Р В РЎвЂР РЋР С“Р РЋРІР‚в„–Р В Р вЂ Р В Р’В°Р В Р’ВµР РЋРІР‚С™ Р В РЎвЂўР В Р’В±Р В Р вЂ¦Р В РЎвЂўР В Р вЂ Р В Р’В»Р РЋРІР‚ВР В Р вЂ¦Р В Р вЂ¦Р В РЎвЂўР В Р’Вµ Р В РЎвЂ”Р В РЎвЂўР В Р’В»Р В РЎвЂўР В Р’В¶Р В Р’ВµР В Р вЂ¦Р В РЎвЂР В Р’Вµ Р В РЎвЂќР В РЎвЂўР РЋР С“Р В РЎВР В РЎвЂР РЋРІР‚РЋР В Р’ВµР РЋР С“Р В РЎвЂќР В РЎвЂўР В РЎвЂ“Р В РЎвЂў Р В РЎвЂўР В Р’В±Р РЋР вЂ°Р В Р’ВµР В РЎвЂќР РЋРІР‚С™Р В Р’В°.
func TestChangeControlledShipToRandomModelAssignsCurrentCharacterOwner(t *testing.T) {
	serverData := testWorldData(t)
	serverData.CosmicObjects.Items[1].OwnerCharacterID = 2
	gameWorld := world.New(1, serverData)

	if _, ok := gameWorld.ConnectAccount(1); !ok {
		t.Fatalf("account was not connected")
	}
	if !gameWorld.ChangeControlledShipToRandomModel(1) {
		t.Fatalf("ship model was not changed")
	}

	if serverData.CosmicObjects.Items[1].OwnerCharacterID != 1 {
		t.Fatalf("changed ship owner = %d, want 1", serverData.CosmicObjects.Items[1].OwnerCharacterID)
	}
}

// Р В РЎСџР РЋР вЂљР В РЎвЂўР В Р вЂ Р В Р’ВµР РЋР вЂљР РЋР РЏР В Р’ВµР РЋРІР‚С™, Р РЋРІР‚РЋР РЋРІР‚С™Р В РЎвЂў Р РЋР С“Р В РЎвЂўР РЋРІР‚В¦Р РЋР вЂљР В Р’В°Р В Р вЂ¦Р В Р’ВµР В Р вЂ¦Р В РЎвЂР В Р’Вµ Р В РўвЂР В Р’В°Р В Р вЂ¦Р В Р вЂ¦Р РЋРІР‚в„–Р РЋРІР‚В¦ Р В Р’В·Р В Р’В°Р В РЎвЂ”Р В РЎвЂР РЋР С“Р РЋРІР‚в„–Р В Р вЂ Р В Р’В°Р В Р’ВµР РЋРІР‚С™ Р В РЎвЂўР В Р’В±Р В Р вЂ¦Р В РЎвЂўР В Р вЂ Р В Р’В»Р РЋРІР‚ВР В Р вЂ¦Р В Р вЂ¦Р В РЎвЂўР В Р’Вµ Р В РЎвЂ”Р В РЎвЂўР В Р’В»Р В РЎвЂўР В Р’В¶Р В Р’ВµР В Р вЂ¦Р В РЎвЂР В Р’Вµ Р В РЎвЂќР В РЎвЂўР РЋР С“Р В РЎВР В РЎвЂР РЋРІР‚РЋР В Р’ВµР РЋР С“Р В РЎвЂќР В РЎвЂўР В РЎвЂ“Р В РЎвЂў Р В РЎвЂўР В Р’В±Р РЋР вЂ°Р В Р’ВµР В РЎвЂќР РЋРІР‚С™Р В Р’В°.
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
	for _, itemModelID := range []int64{101, 301} {
		if counts[itemModelID] != 10 {
			t.Fatalf("changed container item model %d count = %v, want 10; all counts: %+v", itemModelID, counts[itemModelID], counts)
		}
	}
	if counts[403] != 10000 {
		t.Fatalf("changed container ammo count = %v, want 10000; all counts: %+v", counts[403], counts)
	}
}

// Р В РЎСџР РЋР вЂљР В РЎвЂўР В Р вЂ Р В Р’ВµР РЋР вЂљР РЋР РЏР В Р’ВµР РЋРІР‚С™, Р РЋРІР‚РЋР РЋРІР‚С™Р В РЎвЂў Р РЋР С“Р В РЎвЂўР РЋРІР‚В¦Р РЋР вЂљР В Р’В°Р В Р вЂ¦Р В Р’ВµР В Р вЂ¦Р В РЎвЂР В Р’Вµ Р В РўвЂР В Р’В°Р В Р вЂ¦Р В Р вЂ¦Р РЋРІР‚в„–Р РЋРІР‚В¦ Р В Р’В·Р В Р’В°Р В РЎвЂ”Р В РЎвЂР РЋР С“Р РЋРІР‚в„–Р В Р вЂ Р В Р’В°Р В Р’ВµР РЋРІР‚С™ Р В РЎвЂўР В Р’В±Р В Р вЂ¦Р В РЎвЂўР В Р вЂ Р В Р’В»Р РЋРІР‚ВР В Р вЂ¦Р В Р вЂ¦Р В РЎвЂўР В Р’Вµ Р В РЎвЂ”Р В РЎвЂўР В Р’В»Р В РЎвЂўР В Р’В¶Р В Р’ВµР В Р вЂ¦Р В РЎвЂР В Р’Вµ Р В РЎвЂќР В РЎвЂўР РЋР С“Р В РЎВР В РЎвЂР РЋРІР‚РЋР В Р’ВµР РЋР С“Р В РЎвЂќР В РЎвЂўР В РЎвЂ“Р В РЎвЂў Р В РЎвЂўР В Р’В±Р РЋР вЂ°Р В Р’ВµР В РЎвЂќР РЋРІР‚С™Р В Р’В°.
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

// Р В РЎСџР РЋР вЂљР В РЎвЂўР В Р вЂ Р В Р’ВµР РЋР вЂљР РЋР РЏР В Р’ВµР РЋРІР‚С™, Р РЋРІР‚РЋР РЋРІР‚С™Р В РЎвЂў Р РЋР С“Р В РЎвЂўР РЋРІР‚В¦Р РЋР вЂљР В Р’В°Р В Р вЂ¦Р В Р’ВµР В Р вЂ¦Р В РЎвЂР В Р’Вµ Р В РўвЂР В Р’В°Р В Р вЂ¦Р В Р вЂ¦Р РЋРІР‚в„–Р РЋРІР‚В¦ Р В Р’В·Р В Р’В°Р В РЎвЂ”Р В РЎвЂР РЋР С“Р РЋРІР‚в„–Р В Р вЂ Р В Р’В°Р В Р’ВµР РЋРІР‚С™ Р В РЎвЂўР В Р’В±Р В Р вЂ¦Р В РЎвЂўР В Р вЂ Р В Р’В»Р РЋРІР‚ВР В Р вЂ¦Р В Р вЂ¦Р В РЎвЂўР В Р’Вµ Р В РЎвЂ”Р В РЎвЂўР В Р’В»Р В РЎвЂўР В Р’В¶Р В Р’ВµР В Р вЂ¦Р В РЎвЂР В Р’Вµ Р В РЎвЂќР В РЎвЂўР РЋР С“Р В РЎВР В РЎвЂР РЋРІР‚РЋР В Р’ВµР РЋР С“Р В РЎвЂќР В РЎвЂўР В РЎвЂ“Р В РЎвЂў Р В РЎвЂўР В Р’В±Р РЋР вЂ°Р В Р’ВµР В РЎвЂќР РЋРІР‚С™Р В Р’В°.
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

// Р В РЎСџР РЋР вЂљР В РЎвЂўР В Р вЂ Р В Р’ВµР РЋР вЂљР РЋР РЏР В Р’ВµР РЋРІР‚С™, Р РЋРІР‚РЋР РЋРІР‚С™Р В РЎвЂў Р РЋР С“Р В РЎвЂўР РЋРІР‚В¦Р РЋР вЂљР В Р’В°Р В Р вЂ¦Р В Р’ВµР В Р вЂ¦Р В РЎвЂР В Р’Вµ Р В РўвЂР В Р’В°Р В Р вЂ¦Р В Р вЂ¦Р РЋРІР‚в„–Р РЋРІР‚В¦ Р В Р’В·Р В Р’В°Р В РЎвЂ”Р В РЎвЂР РЋР С“Р РЋРІР‚в„–Р В Р вЂ Р В Р’В°Р В Р’ВµР РЋРІР‚С™ Р В РЎвЂўР В Р’В±Р В Р вЂ¦Р В РЎвЂўР В Р вЂ Р В Р’В»Р РЋРІР‚ВР В Р вЂ¦Р В Р вЂ¦Р В РЎвЂўР В Р’Вµ Р В РЎвЂ”Р В РЎвЂўР В Р’В»Р В РЎвЂўР В Р’В¶Р В Р’ВµР В Р вЂ¦Р В РЎвЂР В Р’Вµ Р В РЎвЂќР В РЎвЂўР РЋР С“Р В РЎВР В РЎвЂР РЋРІР‚РЋР В Р’ВµР РЋР С“Р В РЎвЂќР В РЎвЂўР В РЎвЂ“Р В РЎвЂў Р В РЎвЂўР В Р’В±Р РЋР вЂ°Р В Р’ВµР В РЎвЂќР РЋРІР‚С™Р В Р’В°.
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

// Р В РЎСџР РЋР вЂљР В РЎвЂўР В Р вЂ Р В Р’ВµР РЋР вЂљР РЋР РЏР В Р’ВµР РЋРІР‚С™, Р РЋРІР‚РЋР РЋРІР‚С™Р В РЎвЂў Р РЋР С“Р В РЎвЂўР РЋРІР‚В¦Р РЋР вЂљР В Р’В°Р В Р вЂ¦Р В Р’ВµР В Р вЂ¦Р В РЎвЂР В Р’Вµ Р В РўвЂР В Р’В°Р В Р вЂ¦Р В Р вЂ¦Р РЋРІР‚в„–Р РЋРІР‚В¦ Р В Р’В·Р В Р’В°Р В РЎвЂ”Р В РЎвЂР РЋР С“Р РЋРІР‚в„–Р В Р вЂ Р В Р’В°Р В Р’ВµР РЋРІР‚С™ Р В РЎвЂўР В Р’В±Р В Р вЂ¦Р В РЎвЂўР В Р вЂ Р В Р’В»Р РЋРІР‚ВР В Р вЂ¦Р В Р вЂ¦Р В РЎвЂўР В Р’Вµ Р В РЎвЂ”Р В РЎвЂўР В Р’В»Р В РЎвЂўР В Р’В¶Р В Р’ВµР В Р вЂ¦Р В РЎвЂР В Р’Вµ Р В РЎвЂќР В РЎвЂўР РЋР С“Р В РЎВР В РЎвЂР РЋРІР‚РЋР В Р’ВµР РЋР С“Р В РЎвЂќР В РЎвЂўР В РЎвЂ“Р В РЎвЂў Р В РЎвЂўР В Р’В±Р РЋР вЂ°Р В Р’ВµР В РЎвЂќР РЋРІР‚С™Р В Р’В°.
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
