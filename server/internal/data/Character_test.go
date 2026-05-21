package data

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// РџСЂРѕРІРµСЂСЏРµС‚, С‡С‚Рѕ РґРѕР±Р°РІР»РµРЅРёРµ РїРµСЂСЃРѕРЅР°Р¶Р° РЅР°Р·РЅР°С‡Р°РµС‚ РёРґРµРЅС‚РёС„РёРєР°С‚РѕСЂ, РІСЂРµРјСЏ СЃРѕР·РґР°РЅРёСЏ Рё РёРЅРґРµРєСЃ РїРѕ Р°РєРєР°СѓРЅС‚Сѓ.
func TestCharactersAddAssignsIDCreationTimeAndIndexesCharacter(t *testing.T) {
	characters := NewCharacters()

	character, err := characters.Add(&Character{
		AccountID:              7,
		Balance:                100,
		LocationCosmicObjectID: 20,
	})
	if err != nil {
		t.Fatalf("Add returned error: %v", err)
	}

	if character.ID != 1 {
		t.Fatalf("character ID = %d, want 1", character.ID)
	}
	if characters.MaxID != 1 {
		t.Fatalf("MaxID = %d, want 1", characters.MaxID)
	}
	if character.CreationTime.IsZero() {
		t.Fatal("CreationTime is zero")
	}

	byID, ok := characters.Get(character.ID)
	if !ok || byID != character {
		t.Fatal("Get did not return added character")
	}

	byAccountID := characters.GetByAccountID(character.AccountID)
	if len(byAccountID) != 1 || byAccountID[0] != character {
		t.Fatal("GetByAccountID did not return added character")
	}
}

// РџСЂРѕРІРµСЂСЏРµС‚, С‡С‚Рѕ РїРµСЂСЃРѕРЅР°Р¶ Р±РµР· РїСЂРёРІСЏР·РєРё Рє Р°РєРєР°СѓРЅС‚Сѓ РЅРµ РґРѕР±Р°РІР»СЏРµС‚СЃСЏ.
func TestCharactersAddRejectsEmptyAccountID(t *testing.T) {
	characters := NewCharacters()

	if _, err := characters.Add(&Character{}); err == nil {
		t.Fatal("Add accepted empty AccountID")
	}
}

// РџСЂРѕРІРµСЂСЏРµС‚, С‡С‚Рѕ СѓРґР°Р»РµРЅРёРµ РїРµСЂСЃРѕРЅР°Р¶Р° РѕС‡РёС‰Р°РµС‚ РѕСЃРЅРѕРІРЅРѕРµ С…СЂР°РЅРёР»РёС‰Рµ Рё РёРЅРґРµРєСЃ РїРѕ Р°РєРєР°СѓРЅС‚Сѓ.
func TestCharactersDeleteRemovesCharacterAndIndexes(t *testing.T) {
	characters := NewCharacters()
	character, err := characters.Add(&Character{AccountID: 7})
	if err != nil {
		t.Fatalf("Add returned error: %v", err)
	}

	if !characters.Delete(character.ID) {
		t.Fatal("Delete returned false")
	}

	if _, ok := characters.Get(character.ID); ok {
		t.Fatal("deleted character is still stored by ID")
	}
	if byAccountID := characters.GetByAccountID(character.AccountID); len(byAccountID) != 0 {
		t.Fatal("deleted character is still indexed by AccountID")
	}
}

// РџСЂРѕРІРµСЂСЏРµС‚, С‡С‚Рѕ СЃРѕС…СЂР°РЅС‘РЅРЅС‹Рµ РїРµСЂСЃРѕРЅР°Р¶Рё Р·Р°РіСЂСѓР¶Р°СЋС‚СЃСЏ РѕР±СЂР°С‚РЅРѕ СЃ РІРѕСЃСЃС‚Р°РЅРѕРІР»РµРЅРЅС‹Рј РёРЅРґРµРєСЃРѕРј РїРѕ Р°РєРєР°СѓРЅС‚Сѓ.
func TestCharactersSaveLoadAndRebuildIndexes(t *testing.T) {
	path := filepath.Join(t.TempDir(), "Characters.json")
	characters := NewCharacters()
	character, err := characters.Add(&Character{AccountID: 7, Balance: 100, LocationCosmicObjectID: 20})
	if err != nil {
		t.Fatalf("Add returned error: %v", err)
	}

	if err := characters.SaveToFile(path); err != nil {
		t.Fatalf("SaveToFile returned error: %v", err)
	}
	if _, err := os.Stat(path); err != nil {
		t.Fatalf("saved file is not available: %v", err)
	}

	loaded := NewCharacters()
	if err := loaded.LoadFromFile(path); err != nil {
		t.Fatalf("LoadFromFile returned error: %v", err)
	}

	loadedCharacter, ok := loaded.Get(character.ID)
	if !ok {
		t.Fatal("loaded character is not available by ID")
	}
	if loadedCharacter.AccountID != character.AccountID || loadedCharacter.Balance != character.Balance {
		t.Fatal("loaded character fields do not match saved character")
	}
	if byAccountID := loaded.GetByAccountID(character.AccountID); len(byAccountID) != 1 || byAccountID[0] != loadedCharacter {
		t.Fatal("loaded AccountID index is not rebuilt")
	}
}

// РџСЂРѕРІРµСЂСЏРµС‚, С‡С‚Рѕ JSON-РїСЂРµРґСЃС‚Р°РІР»РµРЅРёРµ РїРµСЂСЃРѕРЅР°Р¶РµР№ РёСЃРїРѕР»СЊР·СѓРµС‚ РёРјРµРЅР° РїРѕР»РµР№ РёР· Go-СЃС‚СЂСѓРєС‚СѓСЂ.
func TestCharactersJSONKeysMatchGoFieldNames(t *testing.T) {
	characters := NewCharacters()
	if _, err := characters.Add(&Character{AccountID: 7}); err != nil {
		t.Fatalf("Add returned error: %v", err)
	}

	content, err := json.Marshal(characters)
	if err != nil {
		t.Fatalf("Marshal returned error: %v", err)
	}
	text := string(content)

	expectedKeys := []string{
		`"MaxID"`,
		`"Items"`,
		`"ID"`,
		`"AccountID"`,
		`"CreationTime"`,
		`"Balance"`,
		`"LocationCosmicObjectID"`,
	}
	for _, expectedKey := range expectedKeys {
		if !strings.Contains(text, expectedKey) {
			t.Fatalf("JSON %s does not contain key %s", text, expectedKey)
		}
	}
}

// РџСЂРѕРІРµСЂСЏРµС‚, С‡С‚Рѕ РІРѕСЃСЃС‚Р°РЅРѕРІР»РµРЅРёРµ РёРЅРґРµРєСЃРѕРІ РѕС‚РєР»РѕРЅСЏРµС‚ СЃРѕС…СЂР°РЅС‘РЅРЅРѕРіРѕ РїРµСЂСЃРѕРЅР°Р¶Р° Р±РµР· Р°РєРєР°СѓРЅС‚Р°.
func TestCharactersRebuildIndexesRejectsInvalidStoredCharacter(t *testing.T) {
	characters := NewCharacters()
	characters.Items[1] = &Character{ID: 1}

	if err := characters.RebuildIndexes(); err == nil {
		t.Fatal("RebuildIndexes accepted character without AccountID")
	}
}
