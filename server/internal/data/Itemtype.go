package data

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
)

// РҐСЂР°РЅРёС‚ РґР°РЅРЅС‹Рµ РѕРґРЅРѕРіРѕ С‚РёРїР° РїСЂРµРґРјРµС‚Р°.
type ItemType struct {
	ID                    int64  `json:"ID"`                    // РЈРЅРёРєР°Р»СЊРЅС‹Р№ С‡РёСЃР»РѕРІРѕР№ РёРґРµРЅС‚РёС„РёРєР°С‚РѕСЂ Р·Р°РїРёСЃРё.
	TitleRu               string `json:"TitleRu"`               // Р СѓСЃСЃРєРѕРµ РЅР°Р·РІР°РЅРёРµ РґР»СЏ РёРЅС‚РµСЂС„РµР№СЃР° Рё РґР°РЅРЅС‹С….
	TitleEn               string `json:"TitleEn"`               // РђРЅРіР»РёР№СЃРєРѕРµ РЅР°Р·РІР°РЅРёРµ РґР»СЏ РёРЅС‚РµСЂС„РµР№СЃР° Рё РґР°РЅРЅС‹С….
	Acronym               string `json:"Acronym"`               // РќРµРёР·РјРµРЅСЏРµРјС‹Р№ СЃС‚СЂРѕРєРѕРІС‹Р№ РёРґРµРЅС‚РёС„РёРєР°С‚РѕСЂ РґР»СЏ Р»РѕРіРёРєРё Рё СЃСЃС‹Р»РѕРє.
	IsEquipmentForShip    bool   `json:"IsEquipmentForShip"`    // Р Р°Р·СЂРµС€Р°РµС‚ СѓСЃС‚Р°РЅР°РІР»РёРІР°С‚СЊ РїСЂРµРґРјРµС‚ СЌС‚РѕРіРѕ С‚РёРїР° РЅР° РєРѕСЂР°Р±Р»СЊ.
	IsEquipmentForStation bool   `json:"IsEquipmentForStation"` // Р Р°Р·СЂРµС€Р°РµС‚ СѓСЃС‚Р°РЅР°РІР»РёРІР°С‚СЊ РїСЂРµРґРјРµС‚ СЌС‚РѕРіРѕ С‚РёРїР° РЅР° СЃС‚Р°РЅС†РёСЋ.
	IsPilotInstrument     bool   `json:"IsPilotInstrument"`     // Р Р°Р·СЂРµС€Р°РµС‚ РЅР°Р·РЅР°С‡Р°С‚СЊ РїСЂРµРґРјРµС‚ СЌС‚РѕРіРѕ С‚РёРїР° РІ РїР°РЅРµР»СЊ РїРёР»РѕС‚Р°.
	IsInternalUsable      bool   `json:"IsInternalUsable"`      // Р Р°Р·СЂРµС€Р°РµС‚ РІРЅСѓС‚СЂРµРЅРЅРµРµ РёСЃРїРѕР»СЊР·РѕРІР°РЅРёРµ РїСЂРµРґРјРµС‚Р° СЌС‚РѕРіРѕ С‚РёРїР° РёР· РїР°РЅРµР»Рё СѓРїСЂР°РІР»РµРЅРёСЏ РѕР±РѕСЂСѓРґРѕРІР°РЅРёРµРј.
	CountMustBeInteger    bool   `json:"CountMustBeInteger"`    // РўСЂРµР±СѓРµС‚ С…СЂР°РЅРёС‚СЊ РєРѕР»РёС‡РµСЃС‚РІРѕ С‚РѕР»СЊРєРѕ С†РµР»С‹РјРё РµРґРёРЅРёС†Р°РјРё.
}

// РҐСЂР°РЅРёС‚ С‚РёРїС‹ РїСЂРµРґРјРµС‚РѕРІ Рё Р±С‹СЃС‚СЂС‹Рµ РёРЅРґРµРєСЃС‹ РїРѕ СѓРЅРёРєР°Р»СЊРЅС‹Рј РїРѕР»СЏРј.
type ItemTypes struct {
	MaxID int64               `json:"MaxID"` // РџРѕСЃР»РµРґРЅРёР№ РІС‹РґР°РЅРЅС‹Р№ РёРґРµРЅС‚РёС„РёРєР°С‚РѕСЂ РґР»СЏ РЅРѕРІС‹С… Р·Р°РїРёСЃРµР№.
	Items map[int64]*ItemType `json:"Items"` // РћСЃРЅРѕРІРЅРѕРµ С…СЂР°РЅРёР»РёС‰Рµ Р·Р°РїРёСЃРµР№ РїРѕ С‡РёСЃР»РѕРІРѕРјСѓ РёРґРµРЅС‚РёС„РёРєР°С‚РѕСЂСѓ.

	ByTitleRu map[string]*ItemType `json:"-"` // Р‘С‹СЃС‚СЂС‹Р№ РїРѕРёСЃРє Р·Р°РїРёСЃРё РїРѕ СЂСѓСЃСЃРєРѕРјСѓ РЅР°Р·РІР°РЅРёСЋ.
	ByTitleEn map[string]*ItemType `json:"-"` // Р‘С‹СЃС‚СЂС‹Р№ РїРѕРёСЃРє Р·Р°РїРёСЃРё РїРѕ Р°РЅРіР»РёР№СЃРєРѕРјСѓ РЅР°Р·РІР°РЅРёСЋ.
	ByAcronym map[string]*ItemType `json:"-"` // Р‘С‹СЃС‚СЂС‹Р№ РїРѕРёСЃРє Р·Р°РїРёСЃРё РїРѕ Р°РєСЂРѕРЅРёРјСѓ.
}

// РЎРѕР·РґР°С‘С‚ РїСѓСЃС‚РѕРµ С…СЂР°РЅРёР»РёС‰Рµ С‚РёРїРѕРІ РїСЂРµРґРјРµС‚РѕРІ СЃ РїРѕРґРіРѕС‚РѕРІР»РµРЅРЅС‹РјРё РёРЅРґРµРєСЃР°РјРё.
func NewItemTypes() *ItemTypes {
	itemTypes := &ItemTypes{}
	itemTypes.ensureMaps()
	return itemTypes
}

// Р”РѕР±Р°РІР»СЏРµС‚ РЅРѕРІС‹Р№ С‚РёРї РїСЂРµРґРјРµС‚Р° Рё РЅР°Р·РЅР°С‡Р°РµС‚ РЅРѕРІС‹Р№ ID.
func (itemTypes *ItemTypes) Add(itemType *ItemType) (*ItemType, error) {
	if itemType == nil {
		return nil, errors.New("itemType is nil")
	}
	itemTypes.ensureMaps()
	if err := itemTypes.validateRequiredFields(itemType); err != nil {
		return nil, err
	}
	if err := itemTypes.ensureUniqueForNewType(itemType); err != nil {
		return nil, err
	}

	itemTypes.MaxID++
	itemType.ID = itemTypes.MaxID
	itemTypes.Items[itemType.ID] = itemType
	itemTypes.addIndexes(itemType)
	return itemType, nil
}

// Р’РѕР·РІСЂР°С‰Р°РµС‚ С‚РёРї РїСЂРµРґРјРµС‚Р° РїРѕ ID.
func (itemTypes *ItemTypes) Get(id int64) (*ItemType, bool) {
	itemTypes.ensureMaps()
	itemType, ok := itemTypes.Items[id]
	return itemType, ok
}

// РЈРґР°Р»СЏРµС‚ С‚РёРї РїСЂРµРґРјРµС‚Р° Рё РІСЃРµ РµРіРѕ Р±С‹СЃС‚СЂС‹Рµ РёРЅРґРµРєСЃС‹.
func (itemTypes *ItemTypes) Delete(id int64) bool {
	itemTypes.ensureMaps()
	itemType, ok := itemTypes.Items[id]
	if !ok {
		return false
	}

	itemTypes.deleteIndexes(itemType)
	delete(itemTypes.Items, id)
	return true
}

// Р’РѕР·РІСЂР°С‰Р°РµС‚ С‚РёРї РїСЂРµРґРјРµС‚Р° РїРѕ СѓРЅРёРєР°Р»СЊРЅРѕРјСѓ СЂСѓСЃСЃРєРѕРјСѓ РЅР°Р·РІР°РЅРёСЋ.
func (itemTypes *ItemTypes) GetByTitleRu(titleRu string) (*ItemType, bool) {
	itemTypes.ensureMaps()
	itemType, ok := itemTypes.ByTitleRu[titleRu]
	return itemType, ok
}

// Р’РѕР·РІСЂР°С‰Р°РµС‚ С‚РёРї РїСЂРµРґРјРµС‚Р° РїРѕ СѓРЅРёРєР°Р»СЊРЅРѕРјСѓ Р°РЅРіР»РёР№СЃРєРѕРјСѓ РЅР°Р·РІР°РЅРёСЋ.
func (itemTypes *ItemTypes) GetByTitleEn(titleEn string) (*ItemType, bool) {
	itemTypes.ensureMaps()
	itemType, ok := itemTypes.ByTitleEn[titleEn]
	return itemType, ok
}

// Р’РѕР·РІСЂР°С‰Р°РµС‚ С‚РёРї РїСЂРµРґРјРµС‚Р° РїРѕ СѓРЅРёРєР°Р»СЊРЅРѕРјСѓ Р°РєСЂРѕРЅРёРјСѓ.
func (itemTypes *ItemTypes) GetByAcronym(acronym string) (*ItemType, bool) {
	itemTypes.ensureMaps()
	itemType, ok := itemTypes.ByAcronym[acronym]
	return itemType, ok
}

// РџРµСЂРµСЃРѕР±РёСЂР°РµС‚ Р±С‹СЃС‚СЂС‹Рµ РёРЅРґРµРєСЃС‹ РїРѕСЃР»Рµ Р·Р°РіСЂСѓР·РєРё РёР· JSON.
func (itemTypes *ItemTypes) RebuildIndexes() error {
	itemTypes.ensureItems()
	itemTypes.ByTitleRu = make(map[string]*ItemType)
	itemTypes.ByTitleEn = make(map[string]*ItemType)
	itemTypes.ByAcronym = make(map[string]*ItemType)

	var maxID int64
	for id, itemType := range itemTypes.Items {
		if itemType == nil {
			return fmt.Errorf("itemType with ID %d is nil", id)
		}
		if itemType.ID != id {
			return fmt.Errorf("itemType map key %d does not match type ID %d", id, itemType.ID)
		}
		if err := itemTypes.validateRequiredFields(itemType); err != nil {
			return fmt.Errorf("itemType with ID %d is invalid: %w", id, err)
		}
		if id > maxID {
			maxID = id
		}
		if err := itemTypes.ensureUniqueForNewType(itemType); err != nil {
			return err
		}
		itemTypes.addIndexes(itemType)
	}
	if itemTypes.MaxID < maxID {
		itemTypes.MaxID = maxID
	}
	return nil
}

// Р—Р°РіСЂСѓР¶Р°РµС‚ С‚РёРїС‹ РїСЂРµРґРјРµС‚РѕРІ РёР· JSON-С„Р°Р№Р»Р° Рё РїРµСЂРµСЃРѕР±РёСЂР°РµС‚ Р±С‹СЃС‚СЂС‹Рµ РёРЅРґРµРєСЃС‹.
func (itemTypes *ItemTypes) LoadFromFile(path string) error {
	content, err := os.ReadFile(path)
	if err != nil {
		return err
	}

	loaded := ItemTypes{}
	if err := json.Unmarshal(content, &loaded); err != nil {
		return err
	}
	if err := loaded.RebuildIndexes(); err != nil {
		return err
	}

	*itemTypes = loaded
	return nil
}

// РЎРѕС…СЂР°РЅСЏРµС‚ С‚РёРїС‹ РїСЂРµРґРјРµС‚РѕРІ РІ JSON-С„Р°Р№Р» Р±РµР· РІСЃРїРѕРјРѕРіР°С‚РµР»СЊРЅС‹С… РёРЅРґРµРєСЃРѕРІ.
func (itemTypes *ItemTypes) SaveToFile(path string) error {
	itemTypes.ensureMaps()
	return saveTableWithOrderedItems(path, itemTypes.MaxID, itemTypes.Items)
}

// РџРѕРґРіРѕС‚Р°РІР»РёРІР°РµС‚ РѕСЃРЅРѕРІРЅРѕРµ С…СЂР°РЅРёР»РёС‰Рµ Рё РІСЃРµ РёРЅРґРµРєСЃС‹.
func (itemTypes *ItemTypes) ensureMaps() {
	itemTypes.ensureItems()
	if itemTypes.ByTitleRu == nil {
		itemTypes.ByTitleRu = make(map[string]*ItemType)
	}
	if itemTypes.ByTitleEn == nil {
		itemTypes.ByTitleEn = make(map[string]*ItemType)
	}
	if itemTypes.ByAcronym == nil {
		itemTypes.ByAcronym = make(map[string]*ItemType)
	}
}

// РџРѕРґРіРѕС‚Р°РІР»РёРІР°РµС‚ РѕСЃРЅРѕРІРЅСѓСЋ map С‚РёРїРѕРІ РїСЂРµРґРјРµС‚РѕРІ.
func (itemTypes *ItemTypes) ensureItems() {
	if itemTypes.Items == nil {
		itemTypes.Items = make(map[int64]*ItemType)
	}
}

// РџСЂРѕРІРµСЂСЏРµС‚ РѕР±СЏР·Р°С‚РµР»СЊРЅС‹Рµ РїРѕР»СЏ С‚РёРїР° РїСЂРµРґРјРµС‚Р°.
func (itemTypes *ItemTypes) validateRequiredFields(itemType *ItemType) error {
	if itemType.TitleRu == "" {
		return errors.New("title ru is empty")
	}
	if itemType.TitleEn == "" {
		return errors.New("title en is empty")
	}
	if itemType.Acronym == "" {
		return errors.New("acronym is empty")
	}
	return nil
}

// РџСЂРѕРІРµСЂСЏРµС‚ СѓРЅРёРєР°Р»СЊРЅС‹Рµ РїРѕР»СЏ РїРµСЂРµРґ РґРѕР±Р°РІР»РµРЅРёРµРј РІ РёРЅРґРµРєСЃС‹.
func (itemTypes *ItemTypes) ensureUniqueForNewType(itemType *ItemType) error {
	if existing, ok := itemTypes.ByTitleRu[itemType.TitleRu]; ok && existing.ID != itemType.ID {
		return fmt.Errorf("title ru %q already exists", itemType.TitleRu)
	}
	if existing, ok := itemTypes.ByTitleEn[itemType.TitleEn]; ok && existing.ID != itemType.ID {
		return fmt.Errorf("title en %q already exists", itemType.TitleEn)
	}
	if existing, ok := itemTypes.ByAcronym[itemType.Acronym]; ok && existing.ID != itemType.ID {
		return fmt.Errorf("acronym %q already exists", itemType.Acronym)
	}
	return nil
}

// Р”РѕР±Р°РІР»СЏРµС‚ С‚РёРї РїСЂРµРґРјРµС‚Р° РІРѕ РІСЃРµ Р±С‹СЃС‚СЂС‹Рµ РёРЅРґРµРєСЃС‹.
func (itemTypes *ItemTypes) addIndexes(itemType *ItemType) {
	itemTypes.ByTitleRu[itemType.TitleRu] = itemType
	itemTypes.ByTitleEn[itemType.TitleEn] = itemType
	itemTypes.ByAcronym[itemType.Acronym] = itemType
}

// РЈРґР°Р»СЏРµС‚ С‚РёРї РїСЂРµРґРјРµС‚Р° РёР· РІСЃРµС… Р±С‹СЃС‚СЂС‹С… РёРЅРґРµРєСЃРѕРІ.
func (itemTypes *ItemTypes) deleteIndexes(itemType *ItemType) {
	delete(itemTypes.ByTitleRu, itemType.TitleRu)
	delete(itemTypes.ByTitleEn, itemType.TitleEn)
	delete(itemTypes.ByAcronym, itemType.Acronym)
}
