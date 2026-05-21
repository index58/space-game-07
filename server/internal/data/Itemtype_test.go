package data

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// РџСЂРѕРІРµСЂСЏРµС‚, С‡С‚Рѕ РґРѕР±Р°РІР»РµРЅРёРµ С‚РёРїР° РїСЂРµРґРјРµС‚Р° РЅР°Р·РЅР°С‡Р°РµС‚ РёРґРµРЅС‚РёС„РёРєР°С‚РѕСЂ Рё СЃС‚СЂРѕРёС‚ СѓРЅРёРєР°Р»СЊРЅС‹Рµ РёРЅРґРµРєСЃС‹.
func TestItemTypesAddAssignsIDAndIndexesType(t *testing.T) {
	itemTypes := NewItemTypes()

	itemType, err := itemTypes.Add(&ItemType{
		TitleRu:               "РћСЂСѓР¶РёРµ",
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

	byTitleRu, ok := itemTypes.GetByTitleRu("РћСЂСѓР¶РёРµ")
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

// РџСЂРѕРІРµСЂСЏРµС‚, С‡С‚Рѕ РїРѕРІС‚РѕСЂСЏСЋС‰РёРµСЃСЏ СѓРЅРёРєР°Р»СЊРЅС‹Рµ РЅР°Р·РІР°РЅРёСЏ Рё Р°РєСЂРѕРЅРёРј РЅРµ РґРѕРїСѓСЃРєР°СЋС‚СЃСЏ.
func TestItemTypesAddRejectsDuplicateUniqueFields(t *testing.T) {
	itemTypes := NewItemTypes()

	if _, err := itemTypes.Add(&ItemType{TitleRu: "РћСЂСѓР¶РёРµ", TitleEn: "Weapon", Acronym: "Weapon"}); err != nil {
		t.Fatalf("first Add returned error: %v", err)
	}

	if _, err := itemTypes.Add(&ItemType{TitleRu: "РћСЂСѓР¶РёРµ", TitleEn: "Radar", Acronym: "Radar"}); err == nil {
		t.Fatal("Add accepted duplicate TitleRu")
	}
	if _, err := itemTypes.Add(&ItemType{TitleRu: "Р Р°РґР°СЂ", TitleEn: "Weapon", Acronym: "Radar"}); err == nil {
		t.Fatal("Add accepted duplicate TitleEn")
	}
	if _, err := itemTypes.Add(&ItemType{TitleRu: "Р Р°РґР°СЂ", TitleEn: "Radar", Acronym: "Weapon"}); err == nil {
		t.Fatal("Add accepted duplicate Acronym")
	}
}

// РџСЂРѕРІРµСЂСЏРµС‚, С‡С‚Рѕ СѓРґР°Р»РµРЅРёРµ С‚РёРїР° РїСЂРµРґРјРµС‚Р° РѕС‡РёС‰Р°РµС‚ РѕСЃРЅРѕРІРЅРѕРµ С…СЂР°РЅРёР»РёС‰Рµ Рё РІСЃРµ СѓРЅРёРєР°Р»СЊРЅС‹Рµ РёРЅРґРµРєСЃС‹.
func TestItemTypesDeleteRemovesTypeAndIndexes(t *testing.T) {
	itemTypes := NewItemTypes()
	itemType, err := itemTypes.Add(&ItemType{TitleRu: "РћСЂСѓР¶РёРµ", TitleEn: "Weapon", Acronym: "Weapon"})
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

// РџСЂРѕРІРµСЂСЏРµС‚, С‡С‚Рѕ СЃРѕС…СЂР°РЅС‘РЅРЅС‹Рµ С‚РёРїС‹ РїСЂРµРґРјРµС‚РѕРІ Р·Р°РіСЂСѓР¶Р°СЋС‚СЃСЏ РѕР±СЂР°С‚РЅРѕ СЃ РІРѕСЃСЃС‚Р°РЅРѕРІР»РµРЅРЅС‹Рј РёРЅРґРµРєСЃРѕРј РїРѕ Р°РєСЂРѕРЅРёРјСѓ.
func TestItemTypesSaveLoadAndRebuildIndexes(t *testing.T) {
	path := filepath.Join(t.TempDir(), "ItemTypes.json")
	itemTypes := NewItemTypes()
	itemType, err := itemTypes.Add(&ItemType{
		TitleRu:            "Р РµСЃСѓСЂСЃ",
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

// РџСЂРѕРІРµСЂСЏРµС‚, С‡С‚Рѕ JSON-РїСЂРµРґСЃС‚Р°РІР»РµРЅРёРµ С‚РёРїРѕРІ РїСЂРµРґРјРµС‚РѕРІ РёСЃРїРѕР»СЊР·СѓРµС‚ РёРјРµРЅР° РїРѕР»РµР№ РёР· Go-СЃС‚СЂСѓРєС‚СѓСЂ.
func TestItemTypesJSONKeysMatchGoFieldNames(t *testing.T) {
	itemTypes := NewItemTypes()
	if _, err := itemTypes.Add(&ItemType{TitleRu: "РћСЂСѓР¶РёРµ", TitleEn: "Weapon", Acronym: "Weapon"}); err != nil {
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

// РџСЂРѕРІРµСЂСЏРµС‚, С‡С‚Рѕ РІРѕСЃСЃС‚Р°РЅРѕРІР»РµРЅРёРµ РёРЅРґРµРєСЃРѕРІ РѕС‚РєР»РѕРЅСЏРµС‚ СЃРѕС…СЂР°РЅС‘РЅРЅС‹Р№ С‚РёРї РїСЂРµРґРјРµС‚Р° Р±РµР· РѕР±СЏР·Р°С‚РµР»СЊРЅС‹С… РїРѕР»РµР№.
func TestItemTypesRebuildIndexesRejectsInvalidStoredType(t *testing.T) {
	itemTypes := NewItemTypes()
	itemTypes.Items[1] = &ItemType{ID: 1}

	if err := itemTypes.RebuildIndexes(); err == nil {
		t.Fatal("RebuildIndexes accepted itemType without required fields")
	}
}
