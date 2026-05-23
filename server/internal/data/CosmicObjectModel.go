package data

import (
	"encoding/json"
	"errors"
	"fmt"
	"math"
	"os"
	"strconv"
)

const defaultCosmicObjectModelTextureScale = 4
const bodyPolygonVertexCount = 16

// Коэффициент уменьшения физического тела относительно тела на текстуре.
const cosmicObjectModelBodyScale = 0.95

// РћРїРёСЃС‹РІР°РµС‚ Р»РѕРєР°Р»СЊРЅСѓСЋ С‚РѕС‡РєСѓ С„РёР·РёС‡РµСЃРєРѕРіРѕ С‚РµР»Р° РѕС‚РЅРѕСЃРёС‚РµР»СЊРЅРѕ С†РµРЅС‚СЂР° РѕР±СЉРµРєС‚Р°.
type BodyPoint struct {
	X float64 `json:"x"` // РЎРјРµС‰РµРЅРёРµ РїРѕ РіРѕСЂРёР·РѕРЅС‚Р°Р»СЊРЅРѕР№ Р»РѕРєР°Р»СЊРЅРѕР№ РѕСЃРё РІ РјРµС‚СЂР°С….
	Y float64 `json:"y"` // РЎРјРµС‰РµРЅРёРµ РїРѕ РїСЂРѕРґРѕР»СЊРЅРѕР№ Р»РѕРєР°Р»СЊРЅРѕР№ РѕСЃРё РІ РјРµС‚СЂР°С….
}

// РҐСЂР°РЅРёС‚ РґР°РЅРЅС‹Рµ РѕРґРЅРѕР№ РјРѕРґРµР»Рё РєРѕСЃРјРёС‡РµСЃРєРѕРіРѕ РѕР±СЉРµРєС‚Р°.
type CosmicObjectModel struct {
	ID                 int64       `json:"ID"`                 // РЈРЅРёРєР°Р»СЊРЅС‹Р№ С‡РёСЃР»РѕРІРѕР№ РёРґРµРЅС‚РёС„РёРєР°С‚РѕСЂ Р·Р°РїРёСЃРё.
	TitleRu            string      `json:"TitleRu"`            // Р СѓСЃСЃРєРѕРµ РЅР°Р·РІР°РЅРёРµ РґР»СЏ РёРЅС‚РµСЂС„РµР№СЃР° Рё РґР°РЅРЅС‹С….
	TitleEn            string      `json:"TitleEn"`            // РђРЅРіР»РёР№СЃРєРѕРµ РЅР°Р·РІР°РЅРёРµ РґР»СЏ РёРЅС‚РµСЂС„РµР№СЃР° Рё РґР°РЅРЅС‹С….
	Acronym            string      `json:"Acronym"`            // РќРµРёР·РјРµРЅСЏРµРјС‹Р№ СЃС‚СЂРѕРєРѕРІС‹Р№ РёРґРµРЅС‚РёС„РёРєР°С‚РѕСЂ РґР»СЏ Р»РѕРіРёРєРё Рё СЃСЃС‹Р»РѕРє.
	IconFilePath       string      `json:"IconFilePath"`       // РџСѓС‚СЊ Рє РјР°Р»РµРЅСЊРєРѕРјСѓ РёР·РѕР±СЂР°Р¶РµРЅРёСЋ РјРѕРґРµР»Рё РІ РёРЅС‚РµСЂС„РµР№СЃРµ.
	TextureFilePath    string      `json:"TextureFilePath"`    // РџСѓС‚СЊ Рє РѕСЃРЅРѕРІРЅРѕР№ С‚РµРєСЃС‚СѓСЂРµ РѕР±СЉРµРєС‚Р° РІ РёРіСЂРѕРІРѕРј РјРёСЂРµ.
	TextureWidth       int64       `json:"TextureWidth"`       // РџРѕР»РЅР°СЏ С€РёСЂРёРЅР° С‚РµРєСЃС‚СѓСЂС‹ РІ РїРёРєСЃРµР»СЏС….
	TextureHeight      int64       `json:"TextureHeight"`      // РџРѕР»РЅР°СЏ РІС‹СЃРѕС‚Р° С‚РµРєСЃС‚СѓСЂС‹ РІ РїРёРєСЃРµР»СЏС….
	TextureBodyOriginX int64       `json:"TextureBodyOriginX"` // Р“РѕСЂРёР·РѕРЅС‚Р°Р»СЊРЅРѕРµ СЃРјРµС‰РµРЅРёРµ С„РёР·РёС‡РµСЃРєРѕРіРѕ С‚РµР»Р° РІРЅСѓС‚СЂРё С‚РµРєСЃС‚СѓСЂС‹.
	TextureBodyOriginY int64       `json:"TextureBodyOriginY"` // Р’РµСЂС‚РёРєР°Р»СЊРЅРѕРµ СЃРјРµС‰РµРЅРёРµ С„РёР·РёС‡РµСЃРєРѕРіРѕ С‚РµР»Р° РІРЅСѓС‚СЂРё С‚РµРєСЃС‚СѓСЂС‹.
	TextureBodyWidth   int64       `json:"TextureBodyWidth"`   // РЁРёСЂРёРЅР° С„РёР·РёС‡РµСЃРєРѕРіРѕ С‚РµР»Р° РЅР° С‚РµРєСЃС‚СѓСЂРµ РІ РїРёРєСЃРµР»СЏС….
	TextureBodyLength  int64       `json:"TextureBodyLength"`  // Р”Р»РёРЅР° С„РёР·РёС‡РµСЃРєРѕРіРѕ С‚РµР»Р° РЅР° С‚РµРєСЃС‚СѓСЂРµ РІ РїРёРєСЃРµР»СЏС….
	TextureScale       float64     `json:"TextureScale"`       // РљРѕР»РёС‡РµСЃС‚РІРѕ РїРёРєСЃРµР»РµР№ С‚РµРєСЃС‚СѓСЂС‹ РЅР° РѕРґРёРЅ РјРµС‚СЂ РјРёСЂР°.
	CosmicObjectTypeID int64       `json:"CosmicObjectTypeID"` // РўРёРї РѕР±СЉРµРєС‚Р°, Рє РєРѕС‚РѕСЂРѕРјСѓ РѕС‚РЅРѕСЃРёС‚СЃСЏ РјРѕРґРµР»СЊ.
	Mass               float64     `json:"Mass"`               // Р‘Р°Р·РѕРІР°СЏ РјР°СЃСЃР° СЌРєР·РµРјРїР»СЏСЂР° СЌС‚РѕР№ РјРѕРґРµР»Рё.
	Capacity           float64     `json:"Capacity"`           // Р‘Р°Р·РѕРІС‹Р№ РґРѕСЃС‚СѓРїРЅС‹Р№ РѕР±СЉРµРј РґР»СЏ РѕР±РѕСЂСѓРґРѕРІР°РЅРёСЏ РёР»Рё СЃРѕРґРµСЂР¶РёРјРѕРіРѕ.
	MaxArmor           float64     `json:"MaxArmor"`           // Р‘Р°Р·РѕРІС‹Р№ РјР°РєСЃРёРјСѓРј Р±СЂРѕРЅРё.
	MaxSpeed           float64     `json:"MaxSpeed"`           // Р‘Р°Р·РѕРІР°СЏ РјР°РєСЃРёРјР°Р»СЊРЅР°СЏ Р»РёРЅРµР№РЅР°СЏ СЃРєРѕСЂРѕСЃС‚СЊ.
	MaxAngularSpeed    float64     `json:"MaxAngularSpeed"`    // Р‘Р°Р·РѕРІР°СЏ РјР°РєСЃРёРјР°Р»СЊРЅР°СЏ СѓРіР»РѕРІР°СЏ СЃРєРѕСЂРѕСЃС‚СЊ.
	Complexity         float64     `json:"Complexity"`         // РЎР»РѕР¶РЅРѕСЃС‚СЊ РїСЂРѕРёР·РІРѕРґСЃС‚РІР° Рё РѕС†РµРЅРєРё СЃС‚РѕРёРјРѕСЃС‚Рё РјРѕРґРµР»Рё.
	BodyLength         float64     `json:"BodyLength"`         // Р Р°СЃСЃС‡РёС‚Р°РЅРЅР°СЏ РґР»РёРЅР° С„РёР·РёС‡РµСЃРєРѕРіРѕ С‚РµР»Р° РІ РјРµС‚СЂР°С….
	BodyWidth          float64     `json:"BodyWidth"`          // Р Р°СЃСЃС‡РёС‚Р°РЅРЅР°СЏ С€РёСЂРёРЅР° С„РёР·РёС‡РµСЃРєРѕРіРѕ С‚РµР»Р° РІ РјРµС‚СЂР°С….
	BodyPolygon        []BodyPoint `json:"-"`                  // Р›РѕРєР°Р»СЊРЅС‹Рµ РІРµСЂС€РёРЅС‹ С„РёР·РёС‡РµСЃРєРѕРіРѕ С‚РµР»Р°, СЂР°СЃСЃС‡РёС‚Р°РЅРЅС‹Рµ РїСЂРё Р·Р°РіСЂСѓР·РєРµ СЃРїСЂР°РІРѕС‡РЅРёРєР°.
}

// РҐСЂР°РЅРёС‚ РјРѕРґРµР»Рё РєРѕСЃРјРёС‡РµСЃРєРёС… РѕР±СЉРµРєС‚РѕРІ Рё Р±С‹СЃС‚СЂС‹Рµ РёРЅРґРµРєСЃС‹ РїРѕ СѓРЅРёРєР°Р»СЊРЅС‹Рј РїРѕР»СЏРј.
type CosmicObjectModels struct {
	MaxID int64                        `json:"MaxID"` // РџРѕСЃР»РµРґРЅРёР№ РІС‹РґР°РЅРЅС‹Р№ РёРґРµРЅС‚РёС„РёРєР°С‚РѕСЂ РґР»СЏ РЅРѕРІС‹С… Р·Р°РїРёСЃРµР№.
	Items map[int64]*CosmicObjectModel `json:"Items"` // РћСЃРЅРѕРІРЅРѕРµ С…СЂР°РЅРёР»РёС‰Рµ Р·Р°РїРёСЃРµР№ РїРѕ С‡РёСЃР»РѕРІРѕРјСѓ РёРґРµРЅС‚РёС„РёРєР°С‚РѕСЂСѓ.

	ByTitleRu map[string]*CosmicObjectModel `json:"-"` // Р‘С‹СЃС‚СЂС‹Р№ РїРѕРёСЃРє Р·Р°РїРёСЃРё РїРѕ СЂСѓСЃСЃРєРѕРјСѓ РЅР°Р·РІР°РЅРёСЋ.
	ByTitleEn map[string]*CosmicObjectModel `json:"-"` // Р‘С‹СЃС‚СЂС‹Р№ РїРѕРёСЃРє Р·Р°РїРёСЃРё РїРѕ Р°РЅРіР»РёР№СЃРєРѕРјСѓ РЅР°Р·РІР°РЅРёСЋ.
	ByAcronym map[string]*CosmicObjectModel `json:"-"` // Р‘С‹СЃС‚СЂС‹Р№ РїРѕРёСЃРє Р·Р°РїРёСЃРё РїРѕ Р°РєСЂРѕРЅРёРјСѓ.
}

// РЎРѕР·РґР°С‘С‚ РїСѓСЃС‚РѕРµ С…СЂР°РЅРёР»РёС‰Рµ РјРѕРґРµР»РµР№ РєРѕСЃРјРёС‡РµСЃРєРёС… РѕР±СЉРµРєС‚РѕРІ СЃ РїРѕРґРіРѕС‚РѕРІР»РµРЅРЅС‹РјРё РёРЅРґРµРєСЃР°РјРё.
func NewCosmicObjectModels() *CosmicObjectModels {
	cosmicObjectModels := &CosmicObjectModels{}
	cosmicObjectModels.ensureMaps()
	return cosmicObjectModels
}

// Р”РѕР±Р°РІР»СЏРµС‚ РЅРѕРІСѓСЋ РјРѕРґРµР»СЊ РєРѕСЃРјРёС‡РµСЃРєРѕРіРѕ РѕР±СЉРµРєС‚Р° Рё РЅР°Р·РЅР°С‡Р°РµС‚ РЅРѕРІС‹Р№ ID.
func (cosmicObjectModels *CosmicObjectModels) Add(cosmicObjectModel *CosmicObjectModel) (*CosmicObjectModel, error) {
	if cosmicObjectModel == nil {
		return nil, errors.New("cosmic object model is nil")
	}
	cosmicObjectModels.ensureMaps()
	cosmicObjectModels.prepareCalculatedFields(cosmicObjectModel)
	if err := cosmicObjectModels.validateRequiredFields(cosmicObjectModel); err != nil {
		return nil, err
	}
	if err := cosmicObjectModels.ensureUniqueForNewModel(cosmicObjectModel); err != nil {
		return nil, err
	}

	cosmicObjectModels.MaxID++
	cosmicObjectModel.ID = cosmicObjectModels.MaxID
	cosmicObjectModels.Items[cosmicObjectModel.ID] = cosmicObjectModel
	cosmicObjectModels.addIndexes(cosmicObjectModel)
	return cosmicObjectModel, nil
}

// Р’РѕР·РІСЂР°С‰Р°РµС‚ РјРѕРґРµР»СЊ РєРѕСЃРјРёС‡РµСЃРєРѕРіРѕ РѕР±СЉРµРєС‚Р° РїРѕ ID.
func (cosmicObjectModels *CosmicObjectModels) Get(id int64) (*CosmicObjectModel, bool) {
	cosmicObjectModels.ensureMaps()
	cosmicObjectModel, ok := cosmicObjectModels.Items[id]
	return cosmicObjectModel, ok
}

// РЈРґР°Р»СЏРµС‚ РјРѕРґРµР»СЊ РєРѕСЃРјРёС‡РµСЃРєРѕРіРѕ РѕР±СЉРµРєС‚Р° Рё РІСЃРµ РµС‘ Р±С‹СЃС‚СЂС‹Рµ РёРЅРґРµРєСЃС‹.
func (cosmicObjectModels *CosmicObjectModels) Delete(id int64) bool {
	cosmicObjectModels.ensureMaps()
	cosmicObjectModel, ok := cosmicObjectModels.Items[id]
	if !ok {
		return false
	}

	cosmicObjectModels.deleteIndexes(cosmicObjectModel)
	delete(cosmicObjectModels.Items, id)
	return true
}

// Р’РѕР·РІСЂР°С‰Р°РµС‚ РјРѕРґРµР»СЊ РєРѕСЃРјРёС‡РµСЃРєРѕРіРѕ РѕР±СЉРµРєС‚Р° РїРѕ СѓРЅРёРєР°Р»СЊРЅРѕРјСѓ СЂСѓСЃСЃРєРѕРјСѓ РЅР°Р·РІР°РЅРёСЋ.
func (cosmicObjectModels *CosmicObjectModels) GetByTitleRu(titleRu string) (*CosmicObjectModel, bool) {
	cosmicObjectModels.ensureMaps()
	cosmicObjectModel, ok := cosmicObjectModels.ByTitleRu[titleRu]
	return cosmicObjectModel, ok
}

// Р’РѕР·РІСЂР°С‰Р°РµС‚ РјРѕРґРµР»СЊ РєРѕСЃРјРёС‡РµСЃРєРѕРіРѕ РѕР±СЉРµРєС‚Р° РїРѕ СѓРЅРёРєР°Р»СЊРЅРѕРјСѓ Р°РЅРіР»РёР№СЃРєРѕРјСѓ РЅР°Р·РІР°РЅРёСЋ.
func (cosmicObjectModels *CosmicObjectModels) GetByTitleEn(titleEn string) (*CosmicObjectModel, bool) {
	cosmicObjectModels.ensureMaps()
	cosmicObjectModel, ok := cosmicObjectModels.ByTitleEn[titleEn]
	return cosmicObjectModel, ok
}

// Р’РѕР·РІСЂР°С‰Р°РµС‚ РјРѕРґРµР»СЊ РєРѕСЃРјРёС‡РµСЃРєРѕРіРѕ РѕР±СЉРµРєС‚Р° РїРѕ СѓРЅРёРєР°Р»СЊРЅРѕРјСѓ Р°РєСЂРѕРЅРёРјСѓ.
func (cosmicObjectModels *CosmicObjectModels) GetByAcronym(acronym string) (*CosmicObjectModel, bool) {
	cosmicObjectModels.ensureMaps()
	cosmicObjectModel, ok := cosmicObjectModels.ByAcronym[acronym]
	return cosmicObjectModel, ok
}

// РџРµСЂРµСЃРѕР±РёСЂР°РµС‚ Р±С‹СЃС‚СЂС‹Рµ РёРЅРґРµРєСЃС‹ РїРѕСЃР»Рµ Р·Р°РіСЂСѓР·РєРё РёР· JSON.
func (cosmicObjectModels *CosmicObjectModels) RebuildIndexes() error {
	cosmicObjectModels.ensureItems()
	cosmicObjectModels.ByTitleRu = make(map[string]*CosmicObjectModel)
	cosmicObjectModels.ByTitleEn = make(map[string]*CosmicObjectModel)
	cosmicObjectModels.ByAcronym = make(map[string]*CosmicObjectModel)

	var maxID int64
	for id, cosmicObjectModel := range cosmicObjectModels.Items {
		if cosmicObjectModel == nil {
			return fmt.Errorf("cosmic object model with ID %d is nil", id)
		}
		if cosmicObjectModel.ID != id {
			return fmt.Errorf("cosmic object model map key %d does not match model ID %d", id, cosmicObjectModel.ID)
		}
		cosmicObjectModels.prepareCalculatedFields(cosmicObjectModel)
		if err := cosmicObjectModels.validateRequiredFields(cosmicObjectModel); err != nil {
			return fmt.Errorf("cosmic object model with ID %d is invalid: %w", id, err)
		}
		if id > maxID {
			maxID = id
		}
		if err := cosmicObjectModels.ensureUniqueForNewModel(cosmicObjectModel); err != nil {
			return err
		}
		cosmicObjectModels.addIndexes(cosmicObjectModel)
	}
	if cosmicObjectModels.MaxID < maxID {
		cosmicObjectModels.MaxID = maxID
	}
	return nil
}

// Р—Р°РіСЂСѓР¶Р°РµС‚ РјРѕРґРµР»Рё РєРѕСЃРјРёС‡РµСЃРєРёС… РѕР±СЉРµРєС‚РѕРІ РёР· JSON-С„Р°Р№Р»Р° Рё РїРµСЂРµСЃРѕР±РёСЂР°РµС‚ Р±С‹СЃС‚СЂС‹Рµ РёРЅРґРµРєСЃС‹.
func (cosmicObjectModels *CosmicObjectModels) LoadFromFile(path string) error {
	content, err := os.ReadFile(path)
	if err != nil {
		return err
	}

	loaded := CosmicObjectModels{}
	if err := json.Unmarshal(content, &loaded); err != nil {
		return err
	}
	if err := loaded.RebuildIndexes(); err != nil {
		return err
	}

	*cosmicObjectModels = loaded
	return nil
}

// РЎРѕС…СЂР°РЅСЏРµС‚ РјРѕРґРµР»Рё РєРѕСЃРјРёС‡РµСЃРєРёС… РѕР±СЉРµРєС‚РѕРІ РІ JSON-С„Р°Р№Р» Р±РµР· РІСЃРїРѕРјРѕРіР°С‚РµР»СЊРЅС‹С… РёРЅРґРµРєСЃРѕРІ.
func (cosmicObjectModels *CosmicObjectModels) SaveToFile(path string) error {
	cosmicObjectModels.ensureMaps()
	return saveTableWithOrderedItems(path, cosmicObjectModels.MaxID, cosmicObjectModels.Items)
}

// РџРѕРґРіРѕС‚Р°РІР»РёРІР°РµС‚ РѕСЃРЅРѕРІРЅРѕРµ С…СЂР°РЅРёР»РёС‰Рµ Рё РІСЃРµ РёРЅРґРµРєСЃС‹.
func (cosmicObjectModels *CosmicObjectModels) ensureMaps() {
	cosmicObjectModels.ensureItems()
	if cosmicObjectModels.ByTitleRu == nil {
		cosmicObjectModels.ByTitleRu = make(map[string]*CosmicObjectModel)
	}
	if cosmicObjectModels.ByTitleEn == nil {
		cosmicObjectModels.ByTitleEn = make(map[string]*CosmicObjectModel)
	}
	if cosmicObjectModels.ByAcronym == nil {
		cosmicObjectModels.ByAcronym = make(map[string]*CosmicObjectModel)
	}
}

// РџРѕРґРіРѕС‚Р°РІР»РёРІР°РµС‚ РѕСЃРЅРѕРІРЅСѓСЋ map РјРѕРґРµР»РµР№ РєРѕСЃРјРёС‡РµСЃРєРёС… РѕР±СЉРµРєС‚РѕРІ.
func (cosmicObjectModels *CosmicObjectModels) ensureItems() {
	if cosmicObjectModels.Items == nil {
		cosmicObjectModels.Items = make(map[int64]*CosmicObjectModel)
	}
}

// РџСЂРѕРІРµСЂСЏРµС‚ РѕР±СЏР·Р°С‚РµР»СЊРЅС‹Рµ РїРѕР»СЏ РјРѕРґРµР»Рё РєРѕСЃРјРёС‡РµСЃРєРѕРіРѕ РѕР±СЉРµРєС‚Р°.
func (cosmicObjectModels *CosmicObjectModels) validateRequiredFields(cosmicObjectModel *CosmicObjectModel) error {
	if cosmicObjectModel.TitleRu == "" {
		return errors.New("title ru is empty")
	}
	if cosmicObjectModel.TitleEn == "" {
		return errors.New("title en is empty")
	}
	if cosmicObjectModel.Acronym == "" {
		return errors.New("acronym is empty")
	}
	if cosmicObjectModel.TextureScale <= 0 {
		return errors.New("texture scale is empty")
	}
	if cosmicObjectModel.CosmicObjectTypeID <= 0 {
		return errors.New("cosmic object type ID is empty")
	}
	return nil
}

// РџСЂРѕРІРµСЂСЏРµС‚ СѓРЅРёРєР°Р»СЊРЅС‹Рµ РїРѕР»СЏ РїРµСЂРµРґ РґРѕР±Р°РІР»РµРЅРёРµРј РІ РёРЅРґРµРєСЃС‹.
func (cosmicObjectModels *CosmicObjectModels) ensureUniqueForNewModel(cosmicObjectModel *CosmicObjectModel) error {
	if existing, ok := cosmicObjectModels.ByTitleRu[cosmicObjectModel.TitleRu]; ok && existing.ID != cosmicObjectModel.ID {
		return fmt.Errorf("title ru %q already exists", cosmicObjectModel.TitleRu)
	}
	if existing, ok := cosmicObjectModels.ByTitleEn[cosmicObjectModel.TitleEn]; ok && existing.ID != cosmicObjectModel.ID {
		return fmt.Errorf("title en %q already exists", cosmicObjectModel.TitleEn)
	}
	if existing, ok := cosmicObjectModels.ByAcronym[cosmicObjectModel.Acronym]; ok && existing.ID != cosmicObjectModel.ID {
		return fmt.Errorf("acronym %q already exists", cosmicObjectModel.Acronym)
	}
	return nil
}

// Р”РѕР±Р°РІР»СЏРµС‚ РјРѕРґРµР»СЊ РєРѕСЃРјРёС‡РµСЃРєРѕРіРѕ РѕР±СЉРµРєС‚Р° РІРѕ РІСЃРµ Р±С‹СЃС‚СЂС‹Рµ РёРЅРґРµРєСЃС‹.
func (cosmicObjectModels *CosmicObjectModels) addIndexes(cosmicObjectModel *CosmicObjectModel) {
	cosmicObjectModels.ByTitleRu[cosmicObjectModel.TitleRu] = cosmicObjectModel
	cosmicObjectModels.ByTitleEn[cosmicObjectModel.TitleEn] = cosmicObjectModel
	cosmicObjectModels.ByAcronym[cosmicObjectModel.Acronym] = cosmicObjectModel
}

// РЈРґР°Р»СЏРµС‚ РјРѕРґРµР»СЊ РєРѕСЃРјРёС‡РµСЃРєРѕРіРѕ РѕР±СЉРµРєС‚Р° РёР· РІСЃРµС… Р±С‹СЃС‚СЂС‹С… РёРЅРґРµРєСЃРѕРІ.
func (cosmicObjectModels *CosmicObjectModels) deleteIndexes(cosmicObjectModel *CosmicObjectModel) {
	delete(cosmicObjectModels.ByTitleRu, cosmicObjectModel.TitleRu)
	delete(cosmicObjectModels.ByTitleEn, cosmicObjectModel.TitleEn)
	delete(cosmicObjectModels.ByAcronym, cosmicObjectModel.Acronym)
}

// Р’С‹СЃС‚Р°РІР»СЏРµС‚ Р·РЅР°С‡РµРЅРёСЏ РїРѕ СѓРјРѕР»С‡Р°РЅРёСЋ Рё РїРµСЂРµСЃС‡РёС‚С‹РІР°РµС‚ РІС‹С‡РёСЃР»СЏРµРјС‹Рµ РїРѕР»СЏ.
func (cosmicObjectModels *CosmicObjectModels) prepareCalculatedFields(cosmicObjectModel *CosmicObjectModel) {
	if cosmicObjectModel.TextureScale == 0 {
		cosmicObjectModel.TextureScale = defaultCosmicObjectModelTextureScale
	}
	cosmicObjectModel.BodyLength = float64(cosmicObjectModel.TextureBodyLength) / cosmicObjectModel.TextureScale * cosmicObjectModelBodyScale
	cosmicObjectModel.BodyWidth = float64(cosmicObjectModel.TextureBodyWidth) / cosmicObjectModel.TextureScale * cosmicObjectModelBodyScale
	cosmicObjectModel.BodyPolygon = buildBodyPolygon(*cosmicObjectModel)
}

// РЎС‚СЂРѕРёС‚ РІС‹РїСѓРєР»РѕРµ С‚РµР»Рѕ РїРѕ СЂР°РІРЅРѕРјРµСЂРЅС‹Рј С†РµРЅС‚СЂР°Р»СЊРЅС‹Рј СѓРіР»Р°Рј СЌР»Р»РёРїСЃР°.
func buildBodyPolygon(cosmicObjectModel CosmicObjectModel) []BodyPoint {
	points := make([]BodyPoint, bodyPolygonVertexCount)
	offsetX, offsetY := bodyPolygonOffset(cosmicObjectModel)
	radiusX := cosmicObjectModel.BodyWidth / 2
	radiusY := cosmicObjectModel.BodyLength / 2

	for index := 0; index < bodyPolygonVertexCount; index++ {
		angle := 2 * math.Pi * float64(index) / bodyPolygonVertexCount
		x := offsetX + zeroSmallValue(radiusX*math.Sin(angle))
		y := offsetY + zeroSmallValue(radiusY*math.Cos(angle))
		points[index] = BodyPoint{
			X: zeroSmallValue(x),
			Y: zeroSmallValue(y),
		}
	}

	return points
}

// Р Р°СЃСЃС‡РёС‚С‹РІР°РµС‚ СЃРјРµС‰РµРЅРёРµ С†РµРЅС‚СЂР° С‚РµР»Р° РѕС‚РЅРѕСЃРёС‚РµР»СЊРЅРѕ С†РµРЅС‚СЂР° С‚РµРєСЃС‚СѓСЂС‹.
func bodyPolygonOffset(cosmicObjectModel CosmicObjectModel) (float64, float64) {
	if cosmicObjectModel.TextureScale <= 0 || cosmicObjectModel.TextureWidth <= 0 || cosmicObjectModel.TextureHeight <= 0 {
		return 0, 0
	}

	return (float64(cosmicObjectModel.TextureBodyOriginX) - float64(cosmicObjectModel.TextureWidth)/2) / cosmicObjectModel.TextureScale,
		(float64(cosmicObjectModel.TextureBodyOriginY) - float64(cosmicObjectModel.TextureHeight)/2) / cosmicObjectModel.TextureScale
}

// РЈР±РёСЂР°РµС‚ РјРёРєСЂРѕСЃРєРѕРїРёС‡РµСЃРєРёРµ РїРѕРіСЂРµС€РЅРѕСЃС‚Рё С‚СЂРёРіРѕРЅРѕРјРµС‚СЂРёРё Сѓ РѕСЃРµРІС‹С… РІРµСЂС€РёРЅ.
func zeroSmallValue(value float64) float64 {
	if math.Abs(value) < 0.000000000001 {
		return 0
	}
	return value
}

type legacyCosmicObjectModel struct {
	TextureFilePath      string  `json:"TextureFilePath"`      // РџСѓС‚СЊ Рє РѕСЃРЅРѕРІРЅРѕР№ С‚РµРєСЃС‚СѓСЂРµ РІ СЃС‚Р°СЂРѕРј С„РѕСЂРјР°С‚Рµ РґР°РЅРЅС‹С….
	TextureWidth         int64   `json:"TextureWidth"`         // РџРѕР»РЅР°СЏ С€РёСЂРёРЅР° С‚РµРєСЃС‚СѓСЂС‹ РІ РїРёРєСЃРµР»СЏС… РІ СЃС‚Р°СЂРѕРј С„РѕСЂРјР°С‚Рµ.
	TextureHeight        int64   `json:"TextureHeight"`        // РџРѕР»РЅР°СЏ РІС‹СЃРѕС‚Р° С‚РµРєСЃС‚СѓСЂС‹ РІ РїРёРєСЃРµР»СЏС… РІ СЃС‚Р°СЂРѕРј С„РѕСЂРјР°С‚Рµ.
	TextureObjectOriginX int64   `json:"TextureObjectOriginX"` // Р“РѕСЂРёР·РѕРЅС‚Р°Р»СЊРЅРѕРµ СЃРјРµС‰РµРЅРёРµ С‚РµР»Р° РІ СЃС‚Р°СЂРѕРј С„РѕСЂРјР°С‚Рµ.
	TextureObjectOriginY int64   `json:"TextureObjectOriginY"` // Р’РµСЂС‚РёРєР°Р»СЊРЅРѕРµ СЃРјРµС‰РµРЅРёРµ С‚РµР»Р° РІ СЃС‚Р°СЂРѕРј С„РѕСЂРјР°С‚Рµ.
	TextureObjectWidth   int64   `json:"TextureObjectWidth"`   // РЁРёСЂРёРЅР° С‚РµР»Р° РІ РїРёРєСЃРµР»СЏС… РІ СЃС‚Р°СЂРѕРј С„РѕСЂРјР°С‚Рµ.
	TextureObjectLength  int64   `json:"TextureObjectLength"`  // Р”Р»РёРЅР° С‚РµР»Р° РІ РїРёРєСЃРµР»СЏС… РІ СЃС‚Р°СЂРѕРј С„РѕСЂРјР°С‚Рµ.
	CosmicObjectType     string  `json:"CosmicObjectType"`     // РЎС‚СЂРѕРєРѕРІРѕРµ РёРјСЏ С‚РёРїР° РёР· СЃС‚Р°СЂРѕРіРѕ JSON.
	TitleRu              string  `json:"TitleRu"`              // Р СѓСЃСЃРєРѕРµ РЅР°Р·РІР°РЅРёРµ РёР· СЃС‚Р°СЂРѕРіРѕ JSON.
	TitleEn              string  `json:"TitleEn"`              // РђРЅРіР»РёР№СЃРєРѕРµ РЅР°Р·РІР°РЅРёРµ РёР· СЃС‚Р°СЂРѕРіРѕ JSON.
	Acronym              string  `json:"Acronym"`              // РђРєСЂРѕРЅРёРј РёР· СЃС‚Р°СЂРѕРіРѕ JSON.
	Mass                 float64 `json:"Mass"`                 // Р‘Р°Р·РѕРІР°СЏ РјР°СЃСЃР° РёР· СЃС‚Р°СЂРѕРіРѕ JSON.
	MaxSpeed             float64 `json:"MaxSpeed"`             // РњР°РєСЃРёРјР°Р»СЊРЅР°СЏ Р»РёРЅРµР№РЅР°СЏ СЃРєРѕСЂРѕСЃС‚СЊ РёР· СЃС‚Р°СЂРѕРіРѕ JSON.
	MaxAngularSpeed      float64 `json:"MaxAngularSpeed"`      // РњР°РєСЃРёРјР°Р»СЊРЅР°СЏ СѓРіР»РѕРІР°СЏ СЃРєРѕСЂРѕСЃС‚СЊ РёР· СЃС‚Р°СЂРѕРіРѕ JSON.
}

// РљРѕРЅРІРµСЂС‚РёСЂСѓРµС‚ СЃС‚Р°СЂС‹Р№ JSON РјРѕРґРµР»РµР№ РІ С‚РµРєСѓС‰СѓСЋ СЃС‚СЂСѓРєС‚СѓСЂСѓ РґР°РЅРЅС‹С….
func LoadCosmicObjectModelsFromLegacyFile(path string, cosmicObjectTypes *CosmicObjectTypes) (*CosmicObjectModels, error) {
	content, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	var legacyModels []legacyCosmicObjectModel
	if err := json.Unmarshal(content, &legacyModels); err != nil {
		return nil, err
	}

	titleRuCounts := make(map[string]int)
	titleEnCounts := make(map[string]int)
	for _, legacyModel := range legacyModels {
		titleRuCounts[legacyModel.TitleRu]++
		titleEnCounts[legacyModel.TitleEn]++
	}

	titleRuNumbers := make(map[string]int)
	titleEnNumbers := make(map[string]int)
	cosmicObjectModels := NewCosmicObjectModels()
	for _, legacyModel := range legacyModels {
		cosmicObjectTypeID, err := legacyCosmicObjectTypeID(legacyModel.CosmicObjectType, cosmicObjectTypes)
		if err != nil {
			return nil, err
		}
		titleRu := numberedLegacyTitle(legacyModel.TitleRu, titleRuCounts, titleRuNumbers)
		titleEn := numberedLegacyTitle(legacyModel.TitleEn, titleEnCounts, titleEnNumbers)

		if _, err := cosmicObjectModels.Add(&CosmicObjectModel{
			TitleRu:            titleRu,
			TitleEn:            titleEn,
			Acronym:            legacyModel.Acronym,
			TextureFilePath:    legacyModel.TextureFilePath,
			TextureWidth:       legacyModel.TextureWidth,
			TextureHeight:      legacyModel.TextureHeight,
			TextureBodyOriginX: legacyModel.TextureObjectOriginX,
			TextureBodyOriginY: legacyModel.TextureObjectOriginY,
			TextureBodyWidth:   legacyModel.TextureObjectWidth,
			TextureBodyLength:  legacyModel.TextureObjectLength,
			TextureScale:       defaultCosmicObjectModelTextureScale,
			CosmicObjectTypeID: cosmicObjectTypeID,
			Mass:               legacyModel.Mass,
			MaxSpeed:           legacyModel.MaxSpeed,
			MaxAngularSpeed:    legacyModel.MaxAngularSpeed,
		}); err != nil {
			return nil, err
		}
	}
	return cosmicObjectModels, nil
}

// Р”РѕР±Р°РІР»СЏРµС‚ РїРѕСЂСЏРґРєРѕРІС‹Р№ РЅРѕРјРµСЂ С‚РѕР»СЊРєРѕ Рє РїРѕРІС‚РѕСЂСЏСЋС‰РёРјСЃСЏ РЅР°Р·РІР°РЅРёСЏРј.
func numberedLegacyTitle(title string, counts map[string]int, numbers map[string]int) string {
	if counts[title] <= 1 {
		return title
	}
	numbers[title]++
	return title + " " + strconv.Itoa(numbers[title])
}

// РЎРѕРїРѕСЃС‚Р°РІР»СЏРµС‚ СЃС‚СЂРѕРєРѕРІС‹Р№ С‚РёРї СЃС‚Р°СЂРѕРіРѕ JSON СЃ ID С‚РµРєСѓС‰РµРіРѕ СЃРїСЂР°РІРѕС‡РЅРёРєР° С‚РёРїРѕРІ.
func legacyCosmicObjectTypeID(legacyType string, cosmicObjectTypes *CosmicObjectTypes) (int64, error) {
	acronymByLegacyType := map[string]string{
		"ship":     "Ship",
		"station":  "Station",
		"asteroid": "Asteroid",
	}
	acronym, ok := acronymByLegacyType[legacyType]
	if !ok {
		return 0, fmt.Errorf("unknown legacy cosmic object type %q", legacyType)
	}
	cosmicObjectType, ok := cosmicObjectTypes.GetByAcronym(acronym)
	if !ok {
		return 0, fmt.Errorf("cosmic object type %q not found", acronym)
	}
	return cosmicObjectType.ID, nil
}
