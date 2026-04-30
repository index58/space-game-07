package data

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestItemtypesAddAssignsIDAndIndexesType(t *testing.T) {
	itemtypes := NewItemtypes()

	itemtype, err := itemtypes.Add(&Itemtype{
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

	if itemtype.ID != 1 {
		t.Fatalf("itemtype ID = %d, want 1", itemtype.ID)
	}
	if itemtypes.MaxID != 1 {
		t.Fatalf("MaxID = %d, want 1", itemtypes.MaxID)
	}

	byID, ok := itemtypes.Get(itemtype.ID)
	if !ok || byID != itemtype {
		t.Fatal("Get did not return added itemtype")
	}

	byTitleRu, ok := itemtypes.GetByTitleRu("Оружие")
	if !ok || byTitleRu != itemtype {
		t.Fatal("GetByTitleRu did not return added itemtype")
	}

	byTitleEn, ok := itemtypes.GetByTitleEn("Weapon")
	if !ok || byTitleEn != itemtype {
		t.Fatal("GetByTitleEn did not return added itemtype")
	}

	byAcronym, ok := itemtypes.GetByAcronym("Weapon")
	if !ok || byAcronym != itemtype {
		t.Fatal("GetByAcronym did not return added itemtype")
	}
}

func TestItemtypesAddRejectsDuplicateUniqueFields(t *testing.T) {
	itemtypes := NewItemtypes()

	if _, err := itemtypes.Add(&Itemtype{TitleRu: "Оружие", TitleEn: "Weapon", Acronym: "Weapon"}); err != nil {
		t.Fatalf("first Add returned error: %v", err)
	}

	if _, err := itemtypes.Add(&Itemtype{TitleRu: "Оружие", TitleEn: "Radar", Acronym: "Radar"}); err == nil {
		t.Fatal("Add accepted duplicate TitleRu")
	}
	if _, err := itemtypes.Add(&Itemtype{TitleRu: "Радар", TitleEn: "Weapon", Acronym: "Radar"}); err == nil {
		t.Fatal("Add accepted duplicate TitleEn")
	}
	if _, err := itemtypes.Add(&Itemtype{TitleRu: "Радар", TitleEn: "Radar", Acronym: "Weapon"}); err == nil {
		t.Fatal("Add accepted duplicate Acronym")
	}
}

func TestItemtypesDeleteRemovesTypeAndIndexes(t *testing.T) {
	itemtypes := NewItemtypes()
	itemtype, err := itemtypes.Add(&Itemtype{TitleRu: "Оружие", TitleEn: "Weapon", Acronym: "Weapon"})
	if err != nil {
		t.Fatalf("Add returned error: %v", err)
	}

	if !itemtypes.Delete(itemtype.ID) {
		t.Fatal("Delete returned false")
	}

	if _, ok := itemtypes.Get(itemtype.ID); ok {
		t.Fatal("deleted itemtype is still stored by ID")
	}
	if _, ok := itemtypes.GetByTitleRu(itemtype.TitleRu); ok {
		t.Fatal("deleted itemtype TitleRu is still indexed")
	}
	if _, ok := itemtypes.GetByTitleEn(itemtype.TitleEn); ok {
		t.Fatal("deleted itemtype TitleEn is still indexed")
	}
	if _, ok := itemtypes.GetByAcronym(itemtype.Acronym); ok {
		t.Fatal("deleted itemtype Acronym is still indexed")
	}
}

func TestItemtypesSaveLoadAndRebuildIndexes(t *testing.T) {
	path := filepath.Join(t.TempDir(), "Itemtypes.json")
	itemtypes := NewItemtypes()
	itemtype, err := itemtypes.Add(&Itemtype{
		TitleRu:            "Ресурс",
		TitleEn:            "Resource",
		Acronym:            "Resource",
		CountMustBeInteger: false,
	})
	if err != nil {
		t.Fatalf("Add returned error: %v", err)
	}

	if err := itemtypes.SaveToFile(path); err != nil {
		t.Fatalf("SaveToFile returned error: %v", err)
	}
	if _, err := os.Stat(path); err != nil {
		t.Fatalf("saved file is not available: %v", err)
	}

	loaded := NewItemtypes()
	if err := loaded.LoadFromFile(path); err != nil {
		t.Fatalf("LoadFromFile returned error: %v", err)
	}

	loadedType, ok := loaded.Get(itemtype.ID)
	if !ok {
		t.Fatal("loaded itemtype is not available by ID")
	}
	if loadedType.TitleRu != itemtype.TitleRu || loadedType.TitleEn != itemtype.TitleEn || loadedType.Acronym != itemtype.Acronym {
		t.Fatal("loaded itemtype fields do not match saved type")
	}
	if byAcronym, ok := loaded.GetByAcronym(itemtype.Acronym); !ok || byAcronym != loadedType {
		t.Fatal("loaded Acronym index is not rebuilt")
	}
}

func TestItemtypesJSONKeysMatchGoFieldNames(t *testing.T) {
	itemtypes := NewItemtypes()
	if _, err := itemtypes.Add(&Itemtype{TitleRu: "Оружие", TitleEn: "Weapon", Acronym: "Weapon"}); err != nil {
		t.Fatalf("Add returned error: %v", err)
	}

	content, err := json.Marshal(itemtypes)
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
		`"CountMustBeInteger"`,
	}
	for _, expectedKey := range expectedKeys {
		if !strings.Contains(text, expectedKey) {
			t.Fatalf("JSON %s does not contain key %s", text, expectedKey)
		}
	}
}

func TestItemtypesRebuildIndexesRejectsInvalidStoredType(t *testing.T) {
	itemtypes := NewItemtypes()
	itemtypes.Items[1] = &Itemtype{ID: 1}

	if err := itemtypes.RebuildIndexes(); err == nil {
		t.Fatal("RebuildIndexes accepted itemtype without required fields")
	}
}
