package data

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
)

// РҐСЂР°РЅРёС‚ РґР°РЅРЅС‹Рµ РѕРґРЅРѕРіРѕ С‚РёРїР° РєРѕСЃРјРёС‡РµСЃРєРѕРіРѕ РѕР±СЉРµРєС‚Р°.
type CosmicObjectType struct {
	ID                 int64  `json:"ID"`                 // РЈРЅРёРєР°Р»СЊРЅС‹Р№ С‡РёСЃР»РѕРІРѕР№ РёРґРµРЅС‚РёС„РёРєР°С‚РѕСЂ Р·Р°РїРёСЃРё.
	TitleRu            string `json:"TitleRu"`            // Р СѓСЃСЃРєРѕРµ РЅР°Р·РІР°РЅРёРµ РґР»СЏ РёРЅС‚РµСЂС„РµР№СЃР° Рё РґР°РЅРЅС‹С….
	TitleEn            string `json:"TitleEn"`            // РђРЅРіР»РёР№СЃРєРѕРµ РЅР°Р·РІР°РЅРёРµ РґР»СЏ РёРЅС‚РµСЂС„РµР№СЃР° Рё РґР°РЅРЅС‹С….
	Acronym            string `json:"Acronym"`            // РќРµРёР·РјРµРЅСЏРµРјС‹Р№ СЃС‚СЂРѕРєРѕРІС‹Р№ РёРґРµРЅС‚РёС„РёРєР°С‚РѕСЂ РґР»СЏ Р»РѕРіРёРєРё Рё СЃСЃС‹Р»РѕРє.
	CharacterLocatable bool   `json:"CharacterLocatable"` // РњРѕР¶РµС‚ Р»Рё РїРµСЂСЃРѕРЅР°Р¶ РЅР°С…РѕРґРёС‚СЊСЃСЏ РІРЅСѓС‚СЂРё РѕР±СЉРµРєС‚Р° СЌС‚РѕРіРѕ С‚РёРїР°.
	Movable            bool   `json:"Movable"`            // РњРѕР¶РµС‚ Р»Рё РѕР±СЉРµРєС‚ СЌС‚РѕРіРѕ С‚РёРїР° РјРµРЅСЏС‚СЊ РїРѕР»РѕР¶РµРЅРёРµ РІ РјРёСЂРµ.
	Rotatable          bool   `json:"Rotatable"`          // РњРѕР¶РµС‚ Р»Рё РѕР±СЉРµРєС‚ СЌС‚РѕРіРѕ С‚РёРїР° РјРµРЅСЏС‚СЊ СѓРіРѕР» РїРѕРІРѕСЂРѕС‚Р°.
}

// РҐСЂР°РЅРёС‚ С‚РёРїС‹ РєРѕСЃРјРёС‡РµСЃРєРёС… РѕР±СЉРµРєС‚РѕРІ Рё Р±С‹СЃС‚СЂС‹Рµ РёРЅРґРµРєСЃС‹ РїРѕ СѓРЅРёРєР°Р»СЊРЅС‹Рј РїРѕР»СЏРј.
type CosmicObjectTypes struct {
	MaxID int64                       `json:"MaxID"` // РџРѕСЃР»РµРґРЅРёР№ РІС‹РґР°РЅРЅС‹Р№ РёРґРµРЅС‚РёС„РёРєР°С‚РѕСЂ РґР»СЏ РЅРѕРІС‹С… Р·Р°РїРёСЃРµР№.
	Items map[int64]*CosmicObjectType `json:"Items"` // РћСЃРЅРѕРІРЅРѕРµ С…СЂР°РЅРёР»РёС‰Рµ Р·Р°РїРёСЃРµР№ РїРѕ С‡РёСЃР»РѕРІРѕРјСѓ РёРґРµРЅС‚РёС„РёРєР°С‚РѕСЂСѓ.

	ByTitleRu map[string]*CosmicObjectType `json:"-"` // Р‘С‹СЃС‚СЂС‹Р№ РїРѕРёСЃРє Р·Р°РїРёСЃРё РїРѕ СЂСѓСЃСЃРєРѕРјСѓ РЅР°Р·РІР°РЅРёСЋ.
	ByTitleEn map[string]*CosmicObjectType `json:"-"` // Р‘С‹СЃС‚СЂС‹Р№ РїРѕРёСЃРє Р·Р°РїРёСЃРё РїРѕ Р°РЅРіР»РёР№СЃРєРѕРјСѓ РЅР°Р·РІР°РЅРёСЋ.
	ByAcronym map[string]*CosmicObjectType `json:"-"` // Р‘С‹СЃС‚СЂС‹Р№ РїРѕРёСЃРє Р·Р°РїРёСЃРё РїРѕ Р°РєСЂРѕРЅРёРјСѓ.
}

// РЎРѕР·РґР°С‘С‚ РїСѓСЃС‚РѕРµ С…СЂР°РЅРёР»РёС‰Рµ С‚РёРїРѕРІ РєРѕСЃРјРёС‡РµСЃРєРёС… РѕР±СЉРµРєС‚РѕРІ СЃ РїРѕРґРіРѕС‚РѕРІР»РµРЅРЅС‹РјРё РёРЅРґРµРєСЃР°РјРё.
func NewCosmicObjectTypes() *CosmicObjectTypes {
	cosmicObjectTypes := &CosmicObjectTypes{}
	cosmicObjectTypes.ensureMaps()
	return cosmicObjectTypes
}

// Р”РѕР±Р°РІР»СЏРµС‚ РЅРѕРІС‹Р№ С‚РёРї РєРѕСЃРјРёС‡РµСЃРєРѕРіРѕ РѕР±СЉРµРєС‚Р° Рё РЅР°Р·РЅР°С‡Р°РµС‚ РЅРѕРІС‹Р№ ID.
func (cosmicObjectTypes *CosmicObjectTypes) Add(cosmicObjectType *CosmicObjectType) (*CosmicObjectType, error) {
	if cosmicObjectType == nil {
		return nil, errors.New("cosmic object type is nil")
	}
	cosmicObjectTypes.ensureMaps()
	if err := cosmicObjectTypes.validateRequiredFields(cosmicObjectType); err != nil {
		return nil, err
	}
	if err := cosmicObjectTypes.ensureUniqueForNewType(cosmicObjectType); err != nil {
		return nil, err
	}

	cosmicObjectTypes.MaxID++
	cosmicObjectType.ID = cosmicObjectTypes.MaxID
	cosmicObjectTypes.Items[cosmicObjectType.ID] = cosmicObjectType
	cosmicObjectTypes.addIndexes(cosmicObjectType)
	return cosmicObjectType, nil
}

// Р’РѕР·РІСЂР°С‰Р°РµС‚ С‚РёРї РєРѕСЃРјРёС‡РµСЃРєРѕРіРѕ РѕР±СЉРµРєС‚Р° РїРѕ ID.
func (cosmicObjectTypes *CosmicObjectTypes) Get(id int64) (*CosmicObjectType, bool) {
	cosmicObjectTypes.ensureMaps()
	cosmicObjectType, ok := cosmicObjectTypes.Items[id]
	return cosmicObjectType, ok
}

// РЈРґР°Р»СЏРµС‚ С‚РёРї РєРѕСЃРјРёС‡РµСЃРєРѕРіРѕ РѕР±СЉРµРєС‚Р° Рё РІСЃРµ РµРіРѕ Р±С‹СЃС‚СЂС‹Рµ РёРЅРґРµРєСЃС‹.
func (cosmicObjectTypes *CosmicObjectTypes) Delete(id int64) bool {
	cosmicObjectTypes.ensureMaps()
	cosmicObjectType, ok := cosmicObjectTypes.Items[id]
	if !ok {
		return false
	}

	cosmicObjectTypes.deleteIndexes(cosmicObjectType)
	delete(cosmicObjectTypes.Items, id)
	return true
}

// Р’РѕР·РІСЂР°С‰Р°РµС‚ С‚РёРї РєРѕСЃРјРёС‡РµСЃРєРѕРіРѕ РѕР±СЉРµРєС‚Р° РїРѕ СѓРЅРёРєР°Р»СЊРЅРѕРјСѓ СЂСѓСЃСЃРєРѕРјСѓ РЅР°Р·РІР°РЅРёСЋ.
func (cosmicObjectTypes *CosmicObjectTypes) GetByTitleRu(titleRu string) (*CosmicObjectType, bool) {
	cosmicObjectTypes.ensureMaps()
	cosmicObjectType, ok := cosmicObjectTypes.ByTitleRu[titleRu]
	return cosmicObjectType, ok
}

// Р’РѕР·РІСЂР°С‰Р°РµС‚ С‚РёРї РєРѕСЃРјРёС‡РµСЃРєРѕРіРѕ РѕР±СЉРµРєС‚Р° РїРѕ СѓРЅРёРєР°Р»СЊРЅРѕРјСѓ Р°РЅРіР»РёР№СЃРєРѕРјСѓ РЅР°Р·РІР°РЅРёСЋ.
func (cosmicObjectTypes *CosmicObjectTypes) GetByTitleEn(titleEn string) (*CosmicObjectType, bool) {
	cosmicObjectTypes.ensureMaps()
	cosmicObjectType, ok := cosmicObjectTypes.ByTitleEn[titleEn]
	return cosmicObjectType, ok
}

// Р’РѕР·РІСЂР°С‰Р°РµС‚ С‚РёРї РєРѕСЃРјРёС‡РµСЃРєРѕРіРѕ РѕР±СЉРµРєС‚Р° РїРѕ СѓРЅРёРєР°Р»СЊРЅРѕРјСѓ Р°РєСЂРѕРЅРёРјСѓ.
func (cosmicObjectTypes *CosmicObjectTypes) GetByAcronym(acronym string) (*CosmicObjectType, bool) {
	cosmicObjectTypes.ensureMaps()
	cosmicObjectType, ok := cosmicObjectTypes.ByAcronym[acronym]
	return cosmicObjectType, ok
}

// РџРµСЂРµСЃРѕР±РёСЂР°РµС‚ Р±С‹СЃС‚СЂС‹Рµ РёРЅРґРµРєСЃС‹ РїРѕСЃР»Рµ Р·Р°РіСЂСѓР·РєРё РёР· JSON.
func (cosmicObjectTypes *CosmicObjectTypes) RebuildIndexes() error {
	cosmicObjectTypes.ensureItems()
	cosmicObjectTypes.ByTitleRu = make(map[string]*CosmicObjectType)
	cosmicObjectTypes.ByTitleEn = make(map[string]*CosmicObjectType)
	cosmicObjectTypes.ByAcronym = make(map[string]*CosmicObjectType)

	var maxID int64
	for id, cosmicObjectType := range cosmicObjectTypes.Items {
		if cosmicObjectType == nil {
			return fmt.Errorf("cosmic object type with ID %d is nil", id)
		}
		if cosmicObjectType.ID != id {
			return fmt.Errorf("cosmic object type map key %d does not match type ID %d", id, cosmicObjectType.ID)
		}
		if err := cosmicObjectTypes.validateRequiredFields(cosmicObjectType); err != nil {
			return fmt.Errorf("cosmic object type with ID %d is invalid: %w", id, err)
		}
		if id > maxID {
			maxID = id
		}
		if err := cosmicObjectTypes.ensureUniqueForNewType(cosmicObjectType); err != nil {
			return err
		}
		cosmicObjectTypes.addIndexes(cosmicObjectType)
	}
	if cosmicObjectTypes.MaxID < maxID {
		cosmicObjectTypes.MaxID = maxID
	}
	return nil
}

// Р—Р°РіСЂСѓР¶Р°РµС‚ С‚РёРїС‹ РєРѕСЃРјРёС‡РµСЃРєРёС… РѕР±СЉРµРєС‚РѕРІ РёР· JSON-С„Р°Р№Р»Р° Рё РїРµСЂРµСЃРѕР±РёСЂР°РµС‚ Р±С‹СЃС‚СЂС‹Рµ РёРЅРґРµРєСЃС‹.
func (cosmicObjectTypes *CosmicObjectTypes) LoadFromFile(path string) error {
	content, err := os.ReadFile(path)
	if err != nil {
		return err
	}

	loaded := CosmicObjectTypes{}
	if err := json.Unmarshal(content, &loaded); err != nil {
		return err
	}
	if err := loaded.RebuildIndexes(); err != nil {
		return err
	}

	*cosmicObjectTypes = loaded
	return nil
}

// РЎРѕС…СЂР°РЅСЏРµС‚ С‚РёРїС‹ РєРѕСЃРјРёС‡РµСЃРєРёС… РѕР±СЉРµРєС‚РѕРІ РІ JSON-С„Р°Р№Р» Р±РµР· РІСЃРїРѕРјРѕРіР°С‚РµР»СЊРЅС‹С… РёРЅРґРµРєСЃРѕРІ.
func (cosmicObjectTypes *CosmicObjectTypes) SaveToFile(path string) error {
	cosmicObjectTypes.ensureMaps()
	return saveTableWithOrderedItems(path, cosmicObjectTypes.MaxID, cosmicObjectTypes.Items)
}

// РџРѕРґРіРѕС‚Р°РІР»РёРІР°РµС‚ РѕСЃРЅРѕРІРЅРѕРµ С…СЂР°РЅРёР»РёС‰Рµ Рё РІСЃРµ РёРЅРґРµРєСЃС‹.
func (cosmicObjectTypes *CosmicObjectTypes) ensureMaps() {
	cosmicObjectTypes.ensureItems()
	if cosmicObjectTypes.ByTitleRu == nil {
		cosmicObjectTypes.ByTitleRu = make(map[string]*CosmicObjectType)
	}
	if cosmicObjectTypes.ByTitleEn == nil {
		cosmicObjectTypes.ByTitleEn = make(map[string]*CosmicObjectType)
	}
	if cosmicObjectTypes.ByAcronym == nil {
		cosmicObjectTypes.ByAcronym = make(map[string]*CosmicObjectType)
	}
}

// РџРѕРґРіРѕС‚Р°РІР»РёРІР°РµС‚ РѕСЃРЅРѕРІРЅСѓСЋ map С‚РёРїРѕРІ РєРѕСЃРјРёС‡РµСЃРєРёС… РѕР±СЉРµРєС‚РѕРІ.
func (cosmicObjectTypes *CosmicObjectTypes) ensureItems() {
	if cosmicObjectTypes.Items == nil {
		cosmicObjectTypes.Items = make(map[int64]*CosmicObjectType)
	}
}

// РџСЂРѕРІРµСЂСЏРµС‚ РѕР±СЏР·Р°С‚РµР»СЊРЅС‹Рµ РїРѕР»СЏ С‚РёРїР° РєРѕСЃРјРёС‡РµСЃРєРѕРіРѕ РѕР±СЉРµРєС‚Р°.
func (cosmicObjectTypes *CosmicObjectTypes) validateRequiredFields(cosmicObjectType *CosmicObjectType) error {
	if cosmicObjectType.TitleRu == "" {
		return errors.New("title ru is empty")
	}
	if cosmicObjectType.TitleEn == "" {
		return errors.New("title en is empty")
	}
	if cosmicObjectType.Acronym == "" {
		return errors.New("acronym is empty")
	}
	return nil
}

// РџСЂРѕРІРµСЂСЏРµС‚ СѓРЅРёРєР°Р»СЊРЅС‹Рµ РїРѕР»СЏ РїРµСЂРµРґ РґРѕР±Р°РІР»РµРЅРёРµРј РІ РёРЅРґРµРєСЃС‹.
func (cosmicObjectTypes *CosmicObjectTypes) ensureUniqueForNewType(cosmicObjectType *CosmicObjectType) error {
	if existing, ok := cosmicObjectTypes.ByTitleRu[cosmicObjectType.TitleRu]; ok && existing.ID != cosmicObjectType.ID {
		return fmt.Errorf("title ru %q already exists", cosmicObjectType.TitleRu)
	}
	if existing, ok := cosmicObjectTypes.ByTitleEn[cosmicObjectType.TitleEn]; ok && existing.ID != cosmicObjectType.ID {
		return fmt.Errorf("title en %q already exists", cosmicObjectType.TitleEn)
	}
	if existing, ok := cosmicObjectTypes.ByAcronym[cosmicObjectType.Acronym]; ok && existing.ID != cosmicObjectType.ID {
		return fmt.Errorf("acronym %q already exists", cosmicObjectType.Acronym)
	}
	return nil
}

// Р”РѕР±Р°РІР»СЏРµС‚ С‚РёРї РєРѕСЃРјРёС‡РµСЃРєРѕРіРѕ РѕР±СЉРµРєС‚Р° РІРѕ РІСЃРµ Р±С‹СЃС‚СЂС‹Рµ РёРЅРґРµРєСЃС‹.
func (cosmicObjectTypes *CosmicObjectTypes) addIndexes(cosmicObjectType *CosmicObjectType) {
	cosmicObjectTypes.ByTitleRu[cosmicObjectType.TitleRu] = cosmicObjectType
	cosmicObjectTypes.ByTitleEn[cosmicObjectType.TitleEn] = cosmicObjectType
	cosmicObjectTypes.ByAcronym[cosmicObjectType.Acronym] = cosmicObjectType
}

// РЈРґР°Р»СЏРµС‚ С‚РёРї РєРѕСЃРјРёС‡РµСЃРєРѕРіРѕ РѕР±СЉРµРєС‚Р° РёР· РІСЃРµС… Р±С‹СЃС‚СЂС‹С… РёРЅРґРµРєСЃРѕРІ.
func (cosmicObjectTypes *CosmicObjectTypes) deleteIndexes(cosmicObjectType *CosmicObjectType) {
	delete(cosmicObjectTypes.ByTitleRu, cosmicObjectType.TitleRu)
	delete(cosmicObjectTypes.ByTitleEn, cosmicObjectType.TitleEn)
	delete(cosmicObjectTypes.ByAcronym, cosmicObjectType.Acronym)
}
