package data

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// Проверяет, что добавление типа космического объекта назначает идентификатор и строит уникальные индексы.
func TestCosmicObjectTypesAddAssignsIDAndIndexesType(t *testing.T) {
	cosmicObjectTypes := NewCosmicObjectTypes()

	cosmicObjectType, err := cosmicObjectTypes.Add(&CosmicObjectType{
		TitleRu:            "Корабль",
		TitleEn:            "Ship",
		Acronym:            "Ship",
		CharacterLocatable: true,
		Movable:            true,
		Rotatable:          true,
	})
	if err != nil {
		t.Fatalf("Add returned error: %v", err)
	}

	if cosmicObjectType.ID != 1 {
		t.Fatalf("cosmic object type ID = %d, want 1", cosmicObjectType.ID)
	}
	if cosmicObjectTypes.MaxID != 1 {
		t.Fatalf("MaxID = %d, want 1", cosmicObjectTypes.MaxID)
	}

	byID, ok := cosmicObjectTypes.Get(cosmicObjectType.ID)
	if !ok || byID != cosmicObjectType {
		t.Fatal("Get did not return added cosmic object type")
	}

	byTitleRu, ok := cosmicObjectTypes.GetByTitleRu("Корабль")
	if !ok || byTitleRu != cosmicObjectType {
		t.Fatal("GetByTitleRu did not return added cosmic object type")
	}

	byTitleEn, ok := cosmicObjectTypes.GetByTitleEn("Ship")
	if !ok || byTitleEn != cosmicObjectType {
		t.Fatal("GetByTitleEn did not return added cosmic object type")
	}

	byAcronym, ok := cosmicObjectTypes.GetByAcronym("Ship")
	if !ok || byAcronym != cosmicObjectType {
		t.Fatal("GetByAcronym did not return added cosmic object type")
	}
}

// Проверяет, что повторяющиеся уникальные названия и акроним не допускаются.
func TestCosmicObjectTypesAddRejectsDuplicateUniqueFields(t *testing.T) {
	cosmicObjectTypes := NewCosmicObjectTypes()

	if _, err := cosmicObjectTypes.Add(&CosmicObjectType{TitleRu: "Корабль", TitleEn: "Ship", Acronym: "Ship"}); err != nil {
		t.Fatalf("first Add returned error: %v", err)
	}

	if _, err := cosmicObjectTypes.Add(&CosmicObjectType{TitleRu: "Корабль", TitleEn: "Station", Acronym: "Station"}); err == nil {
		t.Fatal("Add accepted duplicate TitleRu")
	}
	if _, err := cosmicObjectTypes.Add(&CosmicObjectType{TitleRu: "Станция", TitleEn: "Ship", Acronym: "Station"}); err == nil {
		t.Fatal("Add accepted duplicate TitleEn")
	}
	if _, err := cosmicObjectTypes.Add(&CosmicObjectType{TitleRu: "Станция", TitleEn: "Station", Acronym: "Ship"}); err == nil {
		t.Fatal("Add accepted duplicate Acronym")
	}
}

// Проверяет, что удаление типа космического объекта очищает основное хранилище и все уникальные индексы.
func TestCosmicObjectTypesDeleteRemovesTypeAndIndexes(t *testing.T) {
	cosmicObjectTypes := NewCosmicObjectTypes()
	cosmicObjectType, err := cosmicObjectTypes.Add(&CosmicObjectType{TitleRu: "Корабль", TitleEn: "Ship", Acronym: "Ship"})
	if err != nil {
		t.Fatalf("Add returned error: %v", err)
	}

	if !cosmicObjectTypes.Delete(cosmicObjectType.ID) {
		t.Fatal("Delete returned false")
	}

	if _, ok := cosmicObjectTypes.Get(cosmicObjectType.ID); ok {
		t.Fatal("deleted cosmic object type is still stored by ID")
	}
	if _, ok := cosmicObjectTypes.GetByTitleRu(cosmicObjectType.TitleRu); ok {
		t.Fatal("deleted cosmic object type TitleRu is still indexed")
	}
	if _, ok := cosmicObjectTypes.GetByTitleEn(cosmicObjectType.TitleEn); ok {
		t.Fatal("deleted cosmic object type TitleEn is still indexed")
	}
	if _, ok := cosmicObjectTypes.GetByAcronym(cosmicObjectType.Acronym); ok {
		t.Fatal("deleted cosmic object type Acronym is still indexed")
	}
}

// Проверяет, что сохранённые типы космических объектов загружаются обратно с восстановленным индексом по акрониму.
func TestCosmicObjectTypesSaveLoadAndRebuildIndexes(t *testing.T) {
	path := filepath.Join(t.TempDir(), "CosmicObjectTypes.json")
	cosmicObjectTypes := NewCosmicObjectTypes()
	cosmicObjectType, err := cosmicObjectTypes.Add(&CosmicObjectType{
		TitleRu: "Корабль",
		TitleEn: "Ship",
		Acronym: "Ship",
		Movable: true,
	})
	if err != nil {
		t.Fatalf("Add returned error: %v", err)
	}

	if err := cosmicObjectTypes.SaveToFile(path); err != nil {
		t.Fatalf("SaveToFile returned error: %v", err)
	}
	if _, err := os.Stat(path); err != nil {
		t.Fatalf("saved file is not available: %v", err)
	}

	loaded := NewCosmicObjectTypes()
	if err := loaded.LoadFromFile(path); err != nil {
		t.Fatalf("LoadFromFile returned error: %v", err)
	}

	loadedType, ok := loaded.Get(cosmicObjectType.ID)
	if !ok {
		t.Fatal("loaded cosmic object type is not available by ID")
	}
	if loadedType.TitleRu != cosmicObjectType.TitleRu || loadedType.TitleEn != cosmicObjectType.TitleEn || loadedType.Acronym != cosmicObjectType.Acronym {
		t.Fatal("loaded cosmic object type fields do not match saved type")
	}
	if byAcronym, ok := loaded.GetByAcronym(cosmicObjectType.Acronym); !ok || byAcronym != loadedType {
		t.Fatal("loaded Acronym index is not rebuilt")
	}
}

// Проверяет, что JSON-представление типов космических объектов использует имена полей из Go-структур.
func TestCosmicObjectTypesJSONKeysMatchGoFieldNames(t *testing.T) {
	cosmicObjectTypes := NewCosmicObjectTypes()
	if _, err := cosmicObjectTypes.Add(&CosmicObjectType{TitleRu: "Корабль", TitleEn: "Ship", Acronym: "Ship"}); err != nil {
		t.Fatalf("Add returned error: %v", err)
	}

	content, err := json.Marshal(cosmicObjectTypes)
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
		`"CharacterLocatable"`,
		`"Movable"`,
		`"Rotatable"`,
	}
	for _, expectedKey := range expectedKeys {
		if !strings.Contains(text, expectedKey) {
			t.Fatalf("JSON %s does not contain key %s", text, expectedKey)
		}
	}
}

// Проверяет, что восстановление индексов отклоняет сохранённый тип космического объекта без обязательных полей.
func TestCosmicObjectTypesRebuildIndexesRejectsInvalidStoredType(t *testing.T) {
	cosmicObjectTypes := NewCosmicObjectTypes()
	cosmicObjectTypes.Items[1] = &CosmicObjectType{ID: 1}

	if err := cosmicObjectTypes.RebuildIndexes(); err == nil {
		t.Fatal("RebuildIndexes accepted cosmic object type without required fields")
	}
}
