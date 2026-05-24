package data

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
)

// РџСЂРѕРІРµСЂСЏРµС‚ РѕР±СЏР·Р°С‚РµР»СЊРЅС‹Рµ РїРѕР»СЏ РјРѕРґРµР»Рё РїСЂРµРґРјРµС‚Р°.
type ItemModel struct {
	ID                   int64   `json:"ID"`                   // РЈРЅРёРєР°Р»СЊРЅС‹Р№ С‡РёСЃР»РѕРІРѕР№ РёРґРµРЅС‚РёС„РёРєР°С‚РѕСЂ Р·Р°РїРёСЃРё.
	TitleRu              string  `json:"TitleRu"`              // Р СѓСЃСЃРєРѕРµ РЅР°Р·РІР°РЅРёРµ РґР»СЏ РёРЅС‚РµСЂС„РµР№СЃР° Рё РґР°РЅРЅС‹С….
	TitleEn              string  `json:"TitleEn"`              // РђРЅРіР»РёР№СЃРєРѕРµ РЅР°Р·РІР°РЅРёРµ РґР»СЏ РёРЅС‚РµСЂС„РµР№СЃР° Рё РґР°РЅРЅС‹С….
	Acronym              string  `json:"Acronym"`              // РќРµРёР·РјРµРЅСЏРµРјС‹Р№ СЃС‚СЂРѕРєРѕРІС‹Р№ РёРґРµРЅС‚РёС„РёРєР°С‚РѕСЂ РґР»СЏ Р»РѕРіРёРєРё Рё СЃСЃС‹Р»РѕРє.
	IconFilePath         string  `json:"IconFilePath"`         // РџСѓС‚СЊ Рє С„Р°Р№Р»Сѓ РёРєРѕРЅРєРё РїСЂРµРґРјРµС‚Р°.
	ItemTypeID           int64   `json:"ItemTypeID"`           // РўРёРї РїСЂРµРґРјРµС‚Р° РёР· СЃРїСЂР°РІРѕС‡РЅРёРєР° С‚РёРїРѕРІ.
	Mass                 float64 `json:"Mass"`                 // РњР°СЃСЃР° РѕРґРЅРѕР№ РµРґРёРЅРёС†С‹ РїСЂРµРґРјРµС‚Р°.
	Volume               float64 `json:"Volume"`               // РћР±СЉРµРј РѕРґРЅРѕР№ РµРґРёРЅРёС†С‹ РїСЂРµРґРјРµС‚Р°.
	Capacity             float64 `json:"Capacity"`             // Р’РјРµСЃС‚РёРјРѕСЃС‚СЊ РѕР±РѕСЂСѓРґРѕРІР°РЅРёСЏ РёР»Рё РїСЂРµРґРјРµС‚Р°.
	MaxArmor             float64 `json:"MaxArmor"`             // РњР°РєСЃРёРјР°Р»СЊРЅР°СЏ РїСЂРѕС‡РЅРѕСЃС‚СЊ СѓСЃС‚Р°РЅРѕРІР»РµРЅРЅРѕРіРѕ РѕР±РѕСЂСѓРґРѕРІР°РЅРёСЏ.
	ConsumingPower       float64 `json:"ConsumingPower"`       // РџРѕС‚СЂРµР±Р»РµРЅРёРµ СЌРЅРµСЂРіРёРё РІРєР»СЋС‡РµРЅРЅС‹Рј РѕР±РѕСЂСѓРґРѕРІР°РЅРёРµРј.
	GeneratingPower      float64 `json:"GeneratingPower"`      // Р“РµРЅРµСЂР°С†РёСЏ СЌРЅРµСЂРіРёРё РІРєР»СЋС‡РµРЅРЅС‹Рј РѕР±РѕСЂСѓРґРѕРІР°РЅРёРµРј.
	AmmoItemModelID      int64   `json:"AmmoItemModelID"`      // РњРѕРґРµР»СЊ Р±РѕРµРїСЂРёРїР°СЃР° РґР»СЏ РѕСЂСѓР¶РёСЏ.
	FiringRate           float64 `json:"FiringRate"`           // РљРѕР»РёС‡РµСЃС‚РІРѕ РІС‹СЃС‚СЂРµР»РѕРІ РІ СЃРµРєСѓРЅРґСѓ.
	MagazineCapacity     int64   `json:"MagazineCapacity"`     // Р’РјРµСЃС‚РёРјРѕСЃС‚СЊ РјР°РіР°Р·РёРЅР° РѕСЂСѓР¶РёСЏ.
	RechargeTime         float64 `json:"RechargeTime"`         // Р’СЂРµРјСЏ РїРµСЂРµР·Р°СЂСЏРґРєРё РІ СЃРµРєСѓРЅРґР°С….
	Range                float64 `json:"Range"`                // Р”Р°Р»СЊРЅРѕСЃС‚СЊ РґРµР№СЃС‚РІРёСЏ РѕР±РѕСЂСѓРґРѕРІР°РЅРёСЏ.
	Damage               float64 `json:"Damage"`               // РЈСЂРѕРЅ РѕРґРЅРѕРіРѕ РїРѕРїР°РґР°РЅРёСЏ РёР»Рё РІРѕР·РґРµР№СЃС‚РІРёСЏ.
	MiningSpeed          float64 `json:"MiningSpeed"`          // Скорость добычи ресурса в килограммах за секунду для одной установленной единицы.
	ConsumingItemModelID int64   `json:"ConsumingItemModelID"` // РњРѕРґРµР»СЊ РїРѕС‚СЂРµР±Р»СЏРµРјРѕРіРѕ СЂРµСЃСѓСЂСЃР°.
	ConsumingCount       float64 `json:"ConsumingCount"`       // Р Р°СЃС…РѕРґ СЂРµСЃСѓСЂСЃР° Р·Р° СЃРµРєСѓРЅРґСѓ.
	MaxAlongForce        float64 `json:"MaxAlongForce"`        // РњР°РєСЃРёРјР°Р»СЊРЅР°СЏ РїСЂРѕРґРѕР»СЊРЅР°СЏ СЃРёР»Р° РѕР±РѕСЂСѓРґРѕРІР°РЅРёСЏ.
	MaxAcrossForce       float64 `json:"MaxAcrossForce"`       // РњР°РєСЃРёРјР°Р»СЊРЅР°СЏ РїРѕРїРµСЂРµС‡РЅР°СЏ СЃРёР»Р° РѕР±РѕСЂСѓРґРѕРІР°РЅРёСЏ.
	MaxTorque            float64 `json:"MaxTorque"`            // РњР°РєСЃРёРјР°Р»СЊРЅС‹Р№ РєСЂСѓС‚СЏС‰РёР№ РјРѕРјРµРЅС‚ РѕР±РѕСЂСѓРґРѕРІР°РЅРёСЏ.
	MaxEquipmentCount    int64   `json:"MaxEquipmentCount"`    // РњР°РєСЃРёРјР°Р»СЊРЅРѕРµ РєРѕР»РёС‡РµСЃС‚РІРѕ РµРґРёРЅРёС† РѕР±РѕСЂСѓРґРѕРІР°РЅРёСЏ РЅР° РѕР±СЉРµРєС‚Рµ.
	ArmorRepairSpeed     float64 `json:"ArmorRepairSpeed"`     // РЎРєРѕСЂРѕСЃС‚СЊ РІРѕСЃСЃС‚Р°РЅРѕРІР»РµРЅРёСЏ РїСЂРѕС‡РЅРѕСЃС‚Рё.
	ProjectileSpeed      float64 `json:"ProjectileSpeed"`      // Скорость выпущенного боеприпаса в метрах за секунду.
	Complexity           float64 `json:"Complexity"`           // РЎР»РѕР¶РЅРѕСЃС‚СЊ РёР·РіРѕС‚РѕРІР»РµРЅРёСЏ РёР»Рё РѕР±СЃР»СѓР¶РёРІР°РЅРёСЏ.
	Efficiency           float64 `json:"Efficiency"`           // КПД оборудования при выполнении работы.
}

// РџСЂРѕРІРµСЂСЏРµС‚ РѕР±СЏР·Р°С‚РµР»СЊРЅС‹Рµ РїРѕР»СЏ РјРѕРґРµР»Рё РїСЂРµРґРјРµС‚Р°.
type ItemModels struct {
	MaxID int64                `json:"MaxID"` // РџРѕСЃР»РµРґРЅРёР№ РІС‹РґР°РЅРЅС‹Р№ С‡РёСЃР»РѕРІРѕР№ РёРґРµРЅС‚РёС„РёРєР°С‚РѕСЂ РґР»СЏ РЅРѕРІС‹С… Р·Р°РїРёСЃРµР№.
	Items map[int64]*ItemModel `json:"Items"` // РћСЃРЅРѕРІРЅРѕРµ С…СЂР°РЅРёР»РёС‰Рµ Р·Р°РїРёСЃРµР№ РїРѕ С‡РёСЃР»РѕРІРѕРјСѓ РёРґРµРЅС‚РёС„РёРєР°С‚РѕСЂСѓ.

	ByAcronym map[string]*ItemModel `json:"-"` // Р‘С‹СЃС‚СЂС‹Р№ РїРѕРёСЃРє Р·Р°РїРёСЃРё РїРѕ Р°РєСЂРѕРЅРёРјСѓ.
}

// РџСЂРѕРІРµСЂСЏРµС‚ РѕР±СЏР·Р°С‚РµР»СЊРЅС‹Рµ РїРѕР»СЏ РјРѕРґРµР»Рё РїСЂРµРґРјРµС‚Р°.
func NewItemModels() *ItemModels {
	models := &ItemModels{}
	models.ensureMaps()
	return models
}

// РџСЂРѕРІРµСЂСЏРµС‚ РѕР±СЏР·Р°С‚РµР»СЊРЅС‹Рµ РїРѕР»СЏ РјРѕРґРµР»Рё РїСЂРµРґРјРµС‚Р°.
func (models *ItemModels) Get(id int64) (*ItemModel, bool) {
	models.ensureMaps()
	model, ok := models.Items[id]
	return model, ok
}

// РџСЂРѕРІРµСЂСЏРµС‚ РѕР±СЏР·Р°С‚РµР»СЊРЅС‹Рµ РїРѕР»СЏ РјРѕРґРµР»Рё РїСЂРµРґРјРµС‚Р°.
func (models *ItemModels) GetByAcronym(acronym string) (*ItemModel, bool) {
	models.ensureMaps()
	model, ok := models.ByAcronym[acronym]
	return model, ok
}

// РџСЂРѕРІРµСЂСЏРµС‚ РѕР±СЏР·Р°С‚РµР»СЊРЅС‹Рµ РїРѕР»СЏ РјРѕРґРµР»Рё РїСЂРµРґРјРµС‚Р°.
func (models *ItemModels) RebuildIndexes() error {
	models.ensureItems()
	models.ByAcronym = make(map[string]*ItemModel)

	var maxID int64
	for id, model := range models.Items {
		if model == nil {
			return fmt.Errorf("item model with ID %d is nil", id)
		}
		if model.ID != id {
			return fmt.Errorf("item model map key %d does not match model ID %d", id, model.ID)
		}
		if err := models.validateRequiredFields(model); err != nil {
			return fmt.Errorf("item model with ID %d is invalid: %w", id, err)
		}
		if id > maxID {
			maxID = id
		}
		if existing, ok := models.ByAcronym[model.Acronym]; ok && existing.ID != model.ID {
			return fmt.Errorf("acronym %q already exists", model.Acronym)
		}
		models.ByAcronym[model.Acronym] = model
	}
	if models.MaxID < maxID {
		models.MaxID = maxID
	}
	return nil
}

// РџСЂРѕРІРµСЂСЏРµС‚ РѕР±СЏР·Р°С‚РµР»СЊРЅС‹Рµ РїРѕР»СЏ РјРѕРґРµР»Рё РїСЂРµРґРјРµС‚Р°.
func (models *ItemModels) LoadFromFile(path string) error {
	content, err := os.ReadFile(path)
	if err != nil {
		return err
	}

	loaded := ItemModels{}
	if err := json.Unmarshal(content, &loaded); err != nil {
		return err
	}
	if err := loaded.RebuildIndexes(); err != nil {
		return err
	}

	*models = loaded
	return nil
}

// РџСЂРѕРІРµСЂСЏРµС‚ РѕР±СЏР·Р°С‚РµР»СЊРЅС‹Рµ РїРѕР»СЏ РјРѕРґРµР»Рё РїСЂРµРґРјРµС‚Р°.
func (models *ItemModels) SaveToFile(path string) error {
	models.ensureMaps()
	return saveTableWithOrderedItems(path, models.MaxID, models.Items)
}

// РџСЂРѕРІРµСЂСЏРµС‚ РѕР±СЏР·Р°С‚РµР»СЊРЅС‹Рµ РїРѕР»СЏ РјРѕРґРµР»Рё РїСЂРµРґРјРµС‚Р°.
func (models *ItemModels) ensureMaps() {
	models.ensureItems()
	if models.ByAcronym == nil {
		models.ByAcronym = make(map[string]*ItemModel)
	}
}

// РџСЂРѕРІРµСЂСЏРµС‚ РѕР±СЏР·Р°С‚РµР»СЊРЅС‹Рµ РїРѕР»СЏ РјРѕРґРµР»Рё РїСЂРµРґРјРµС‚Р°.
func (models *ItemModels) ensureItems() {
	if models.Items == nil {
		models.Items = make(map[int64]*ItemModel)
	}
}

// РџСЂРѕРІРµСЂСЏРµС‚ РѕР±СЏР·Р°С‚РµР»СЊРЅС‹Рµ РїРѕР»СЏ РјРѕРґРµР»Рё РїСЂРµРґРјРµС‚Р°.
func (models *ItemModels) validateRequiredFields(model *ItemModel) error {
	if model.Acronym == "" {
		return errors.New("acronym is empty")
	}
	return nil
}
