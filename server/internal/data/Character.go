package data

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"sort"
	"time"
)

// РҐСЂР°РЅРёС‚ РґР°РЅРЅС‹Рµ РѕРґРЅРѕРіРѕ РїРµСЂСЃРѕРЅР°Р¶Р° РёРіСЂРѕРІРѕРіРѕ РјРёСЂР°.
type Character struct {
	ID                     int64     `json:"ID"`                     // РЈРЅРёРєР°Р»СЊРЅС‹Р№ С‡РёСЃР»РѕРІРѕР№ РёРґРµРЅС‚РёС„РёРєР°С‚РѕСЂ Р·Р°РїРёСЃРё.
	AccountID              int64     `json:"AccountID"`              // РЈС‡РµС‚РЅР°СЏ Р·Р°РїРёСЃСЊ, РєРѕС‚РѕСЂРѕР№ РїСЂРёРЅР°РґР»РµР¶РёС‚ РїРµСЂСЃРѕРЅР°Р¶.
	CreationTime           time.Time `json:"CreationTime"`           // РњРѕРјРµРЅС‚ СЃРѕР·РґР°РЅРёСЏ РїРµСЂСЃРѕРЅР°Р¶Р°.
	Balance                int64     `json:"Balance"`                // РљРѕР»РёС‡РµСЃС‚РІРѕ РґРµРЅРµР¶РЅС‹С… РµРґРёРЅРёС† РЅР° СЃС‡РµС‚Рµ РїРµСЂСЃРѕРЅР°Р¶Р°.
	LocationCosmicObjectID int64     `json:"LocationCosmicObjectID"` // РљРѕСЃРјРёС‡РµСЃРєРёР№ РѕР±СЉРµРєС‚, РЅР° РєРѕС‚РѕСЂРѕРј РїРµСЂСЃРѕРЅР°Р¶ РЅР°С…РѕРґРёС‚СЃСЏ СЃРµР№С‡Р°СЃ.
}

// РҐСЂР°РЅРёС‚ РїРµСЂСЃРѕРЅР°Р¶РµР№ Рё Р±С‹СЃС‚СЂС‹Рµ РёРЅРґРµРєСЃС‹ РґР»СЏ РїРѕРёСЃРєР° РїРѕ СЃРІСЏР·Р°РЅРЅС‹Рј РѕР±СЉРµРєС‚Р°Рј.
type Characters struct {
	MaxID int64                `json:"MaxID"` // РџРѕСЃР»РµРґРЅРёР№ РІС‹РґР°РЅРЅС‹Р№ РёРґРµРЅС‚РёС„РёРєР°С‚РѕСЂ РґР»СЏ РЅРѕРІС‹С… Р·Р°РїРёСЃРµР№.
	Items map[int64]*Character `json:"Items"` // РћСЃРЅРѕРІРЅРѕРµ С…СЂР°РЅРёР»РёС‰Рµ Р·Р°РїРёСЃРµР№ РїРѕ С‡РёСЃР»РѕРІРѕРјСѓ РёРґРµРЅС‚РёС„РёРєР°С‚РѕСЂСѓ.

	ByAccountID map[int64]map[int64]*Character `json:"-"` // Р‘С‹СЃС‚СЂС‹Р№ РїРѕРёСЃРє РІСЃРµС… РїРµСЂСЃРѕРЅР°Р¶РµР№ СѓРєР°Р·Р°РЅРЅРѕР№ СѓС‡РµС‚РЅРѕР№ Р·Р°РїРёСЃРё.
}

// РЎРѕР·РґР°С‘С‚ РїСѓСЃС‚РѕРµ С…СЂР°РЅРёР»РёС‰Рµ РїРµСЂСЃРѕРЅР°Р¶РµР№ СЃ РїРѕРґРіРѕС‚РѕРІР»РµРЅРЅС‹РјРё РёРЅРґРµРєСЃР°РјРё.
func NewCharacters() *Characters {
	characters := &Characters{}
	characters.ensureMaps()
	return characters
}

// Р”РѕР±Р°РІР»СЏРµС‚ РЅРѕРІРѕРіРѕ РїРµСЂСЃРѕРЅР°Р¶Р°, РЅР°Р·РЅР°С‡Р°РµС‚ РЅРѕРІС‹Р№ ID Рё РІСЂРµРјСЏ СЃРѕР·РґР°РЅРёСЏ.
func (characters *Characters) Add(character *Character) (*Character, error) {
	if character == nil {
		return nil, errors.New("character is nil")
	}
	characters.ensureMaps()
	if err := characters.validateRequiredFields(character); err != nil {
		return nil, err
	}

	characters.MaxID++
	character.ID = characters.MaxID
	if character.CreationTime.IsZero() {
		character.CreationTime = time.Now()
	}

	characters.Items[character.ID] = character
	characters.addIndexes(character)
	return character, nil
}

// Р’РѕР·РІСЂР°С‰Р°РµС‚ РїРµСЂСЃРѕРЅР°Р¶Р° РїРѕ ID.
func (characters *Characters) Get(id int64) (*Character, bool) {
	characters.ensureMaps()
	character, ok := characters.Items[id]
	return character, ok
}

// РЈРґР°Р»СЏРµС‚ РїРµСЂСЃРѕРЅР°Р¶Р° Рё РІСЃРµ РµРіРѕ Р±С‹СЃС‚СЂС‹Рµ РёРЅРґРµРєСЃС‹.
func (characters *Characters) Delete(id int64) bool {
	characters.ensureMaps()
	character, ok := characters.Items[id]
	if !ok {
		return false
	}

	characters.deleteIndexes(character)
	delete(characters.Items, id)
	return true
}

// Р’РѕР·РІСЂР°С‰Р°РµС‚ РїРµСЂСЃРѕРЅР°Р¶РµР№ СѓРєР°Р·Р°РЅРЅРѕРіРѕ Р°РєРєР°СѓРЅС‚Р° РІ РїРѕСЂСЏРґРєРµ РІРѕР·СЂР°СЃС‚Р°РЅРёСЏ ID.
func (characters *Characters) GetByAccountID(accountID int64) []*Character {
	characters.ensureMaps()
	indexItems := characters.ByAccountID[accountID]
	if len(indexItems) == 0 {
		return []*Character{}
	}

	ids := make([]int64, 0, len(indexItems))
	for id := range indexItems {
		ids = append(ids, id)
	}
	sort.Slice(ids, func(left int, right int) bool {
		return ids[left] < ids[right]
	})

	result := make([]*Character, 0, len(ids))
	for _, id := range ids {
		result = append(result, indexItems[id])
	}
	return result
}

// РџРµСЂРµСЃРѕР±РёСЂР°РµС‚ Р±С‹СЃС‚СЂС‹Рµ РёРЅРґРµРєСЃС‹ РїРѕСЃР»Рµ Р·Р°РіСЂСѓР·РєРё РёР· JSON.
func (characters *Characters) RebuildIndexes() error {
	characters.ensureItems()
	characters.ByAccountID = make(map[int64]map[int64]*Character)

	var maxID int64
	for id, character := range characters.Items {
		if character == nil {
			return fmt.Errorf("character with ID %d is nil", id)
		}
		if character.ID != id {
			return fmt.Errorf("character map key %d does not match character ID %d", id, character.ID)
		}
		if err := characters.validateRequiredFields(character); err != nil {
			return fmt.Errorf("character with ID %d is invalid: %w", id, err)
		}
		if id > maxID {
			maxID = id
		}
		characters.addIndexes(character)
	}
	if characters.MaxID < maxID {
		characters.MaxID = maxID
	}
	return nil
}

// Р—Р°РіСЂСѓР¶Р°РµС‚ РїРµСЂСЃРѕРЅР°Р¶РµР№ РёР· JSON-С„Р°Р№Р»Р° Рё РїРµСЂРµСЃРѕР±РёСЂР°РµС‚ Р±С‹СЃС‚СЂС‹Рµ РёРЅРґРµРєСЃС‹.
func (characters *Characters) LoadFromFile(path string) error {
	content, err := os.ReadFile(path)
	if err != nil {
		return err
	}

	loaded := Characters{}
	if err := json.Unmarshal(content, &loaded); err != nil {
		return err
	}
	if err := loaded.RebuildIndexes(); err != nil {
		return err
	}

	*characters = loaded
	return nil
}

// РЎРѕС…СЂР°РЅСЏРµС‚ РїРµСЂСЃРѕРЅР°Р¶РµР№ РІ JSON-С„Р°Р№Р» Р±РµР· РІСЃРїРѕРјРѕРіР°С‚РµР»СЊРЅС‹С… РёРЅРґРµРєСЃРѕРІ.
func (characters *Characters) SaveToFile(path string) error {
	characters.ensureMaps()
	return saveTableWithOrderedItems(path, characters.MaxID, characters.Items)
}

// РџРѕРґРіРѕС‚Р°РІР»РёРІР°РµС‚ РѕСЃРЅРѕРІРЅРѕРµ С…СЂР°РЅРёР»РёС‰Рµ Рё РІСЃРµ РёРЅРґРµРєСЃС‹.
func (characters *Characters) ensureMaps() {
	characters.ensureItems()
	if characters.ByAccountID == nil {
		characters.ByAccountID = make(map[int64]map[int64]*Character)
	}
}

// РџРѕРґРіРѕС‚Р°РІР»РёРІР°РµС‚ РѕСЃРЅРѕРІРЅСѓСЋ map РїРµСЂСЃРѕРЅР°Р¶РµР№.
func (characters *Characters) ensureItems() {
	if characters.Items == nil {
		characters.Items = make(map[int64]*Character)
	}
}

// РџСЂРѕРІРµСЂСЏРµС‚ РѕР±СЏР·Р°С‚РµР»СЊРЅС‹Рµ РїРѕР»СЏ РїРµСЂСЃРѕРЅР°Р¶Р°.
func (characters *Characters) validateRequiredFields(character *Character) error {
	if character.AccountID <= 0 {
		return errors.New("account ID is empty")
	}
	return nil
}

// Р”РѕР±Р°РІР»СЏРµС‚ РїРµСЂСЃРѕРЅР°Р¶Р° РІРѕ РІСЃРµ Р±С‹СЃС‚СЂС‹Рµ РёРЅРґРµРєСЃС‹.
func (characters *Characters) addIndexes(character *Character) {
	if characters.ByAccountID[character.AccountID] == nil {
		characters.ByAccountID[character.AccountID] = make(map[int64]*Character)
	}
	characters.ByAccountID[character.AccountID][character.ID] = character
}

// РЈРґР°Р»СЏРµС‚ РїРµСЂСЃРѕРЅР°Р¶Р° РёР· РІСЃРµС… Р±С‹СЃС‚СЂС‹С… РёРЅРґРµРєСЃРѕРІ.
func (characters *Characters) deleteIndexes(character *Character) {
	accountCharacters := characters.ByAccountID[character.AccountID]
	if accountCharacters == nil {
		return
	}

	delete(accountCharacters, character.ID)
	if len(accountCharacters) == 0 {
		delete(characters.ByAccountID, character.AccountID)
	}
}
