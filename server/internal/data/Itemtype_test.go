package data

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// Проверяет, что добавление типа предмета назначает идентификатор и строит уникальные индексы.
func TestItemTypesAddAssignsIDAndIndexesType(t *testing.T) {
	itemTypes := NewItemTypes()

	itemType, err := itemTypes.Add(&ItemType{
		TitleRu:               "Оружие",
		TitleEn:               "Weapon",
		Acronym:               "Weapon",
		IsEquipmentForShip:    true,
		IsEquipmentForStation: true,
		IsPilotInstrument:     true,
		CountMustBeInteger:    true,
	})
	if err != nil {
		t.Fatalf("Add returned error: %v", err)
	}

	if itemType.ID != 1 {
		t.Fatalf("itemType ID = %d, want 1", itemType.ID)
	}
	if itemTypes.MaxID != 1 {
		t.Fatalf("MaxID = %d, want 1", itemTypes.MaxID)
	}

	byID, ok := itemTypes.Get(itemType.ID)
	if !ok || byID != itemType {
		t.Fatal("Get did not return added itemType")
	}

	byTitleRu, ok := itemTypes.GetByTitleRu("Оружие")
	if !ok || byTitleRu != itemType {
		t.Fatal("GetByTitleRu did not return added itemType")
	}

	byTitleEn, ok := itemTypes.GetByTitleEn("Weapon")
	if !ok || byTitleEn != itemType {
		t.Fatal("GetByTitleEn did not return added itemType")
	}

	byAcronym, ok := itemTypes.GetByAcronym("Weapon")
	if !ok || byAcronym != itemType {
		t.Fatal("GetByAcronym did not return added itemType")
	}
}

// Проверяет, что повторяющиеся уникальные названия и акроним не допускаются.
func TestItemTypesAddRejectsDuplicateUniqueFields(t *testing.T) {
	itemTypes := NewItemTypes()

	if _, err := itemTypes.Add(&ItemType{TitleRu: "Оружие", TitleEn: "Weapon", Acronym: "Weapon"}); err != nil {
		t.Fatalf("first Add returned error: %v", err)
	}

	if _, err := itemTypes.Add(&ItemType{TitleRu: "Оружие", TitleEn: "Radar", Acronym: "Radar"}); err == nil {
		t.Fatal("Add accepted duplicate TitleRu")
	}
	if _, err := itemTypes.Add(&ItemType{TitleRu: "Радар", TitleEn: "Weapon", Acronym: "Radar"}); err == nil {
		t.Fatal("Add accepted duplicate TitleEn")
	}
	if _, err := itemTypes.Add(&ItemType{TitleRu: "Радар", TitleEn: "Radar", Acronym: "Weapon"}); err == nil {
		t.Fatal("Add accepted duplicate Acronym")
	}
}

// Проверяет, что удаление типа предмета очищает основное хранилище и все уникальные индексы.
func TestItemTypesDeleteRemovesTypeAndIndexes(t *testing.T) {
	itemTypes := NewItemTypes()
	itemType, err := itemTypes.Add(&ItemType{TitleRu: "Оружие", TitleEn: "Weapon", Acronym: "Weapon"})
	if err != nil {
		t.Fatalf("Add returned error: %v", err)
	}

	if !itemTypes.Delete(itemType.ID) {
		t.Fatal("Delete returned false")
	}

	if _, ok := itemTypes.Get(itemType.ID); ok {
		t.Fatal("deleted itemType is still stored by ID")
	}
	if _, ok := itemTypes.GetByTitleRu(itemType.TitleRu); ok {
		t.Fatal("deleted itemType TitleRu is still indexed")
	}
	if _, ok := itemTypes.GetByTitleEn(itemType.TitleEn); ok {
		t.Fatal("deleted itemType TitleEn is still indexed")
	}
	if _, ok := itemTypes.GetByAcronym(itemType.Acronym); ok {
		t.Fatal("deleted itemType Acronym is still indexed")
	}
}

// Проверяет, что сохранённые типы предметов загружаются обратно с восстановленным индексом по акрониму.
func TestItemTypesSaveLoadAndRebuildIndexes(t *testing.T) {
	path := filepath.Join(t.TempDir(), "ItemTypes.json")
	itemTypes := NewItemTypes()
	itemType, err := itemTypes.Add(&ItemType{
		TitleRu:            "Ресурс",
		TitleEn:            "Resource",
		Acronym:            "Resource",
		CountMustBeInteger: false,
	})
	if err != nil {
		t.Fatalf("Add returned error: %v", err)
	}

	if err := itemTypes.SaveToFile(path); err != nil {
		t.Fatalf("SaveToFile returned error: %v", err)
	}
	if _, err := os.Stat(path); err != nil {
		t.Fatalf("saved file is not available: %v", err)
	}

	loaded := NewItemTypes()
	if err := loaded.LoadFromFile(path); err != nil {
		t.Fatalf("LoadFromFile returned error: %v", err)
	}

	loadedType, ok := loaded.Get(itemType.ID)
	if !ok {
		t.Fatal("loaded itemType is not available by ID")
	}
	if loadedType.TitleRu != itemType.TitleRu || loadedType.TitleEn != itemType.TitleEn || loadedType.Acronym != itemType.Acronym {
		t.Fatal("loaded itemType fields do not match saved type")
	}
	if byAcronym, ok := loaded.GetByAcronym(itemType.Acronym); !ok || byAcronym != loadedType {
		t.Fatal("loaded Acronym index is not rebuilt")
	}
}

// Проверяет, что JSON-представление типов предметов использует имена полей из Go-структур.
func TestItemTypesJSONKeysMatchGoFieldNames(t *testing.T) {
	itemTypes := NewItemTypes()
	if _, err := itemTypes.Add(&ItemType{TitleRu: "Оружие", TitleEn: "Weapon", Acronym: "Weapon"}); err != nil {
		t.Fatalf("Add returned error: %v", err)
	}

	content, err := json.Marshal(itemTypes)
	if err != nil {
		t.Fatalf("Marshal returned error: %v", err)
	}
	text := string(content)

	expectedKeys := []string{
		`"MaxID"`,
		`"Items"`,
		`"ID"`,
		`"TitleRu"`,
		`"TitleEn"`,
		`"Acronym"`,
		`"IsEquipmentForShip"`,
		`"IsEquipmentForStation"`,
		`"IsPilotInstrument"`,
		`"IsInternalUsable"`,
		`"CountMustBeInteger"`,
	}
	for _, expectedKey := range expectedKeys {
		if !strings.Contains(text, expectedKey) {
			t.Fatalf("JSON %s does not contain key %s", text, expectedKey)
		}
	}
}

// Проверяет, что восстановление индексов отклоняет сохранённый тип предмета без обязательных полей.
func TestItemTypesRebuildIndexesRejectsInvalidStoredType(t *testing.T) {
	itemTypes := NewItemTypes()
	itemTypes.Items[1] = &ItemType{ID: 1}

	if err := itemTypes.RebuildIndexes(); err == nil {
		t.Fatal("RebuildIndexes accepted itemType without required fields")
	}
}
