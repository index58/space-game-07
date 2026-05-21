package data

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"sort"
)

// РҐСЂР°РЅРёС‚ РґР°РЅРЅС‹Рµ РѕРґРЅРѕРіРѕ РєРѕСЃРјРёС‡РµСЃРєРѕРіРѕ РѕР±СЉРµРєС‚Р° РёРіСЂРѕРІРѕРіРѕ РјРёСЂР°.
type CosmicObject struct {
	ID                        int64   `json:"ID"`                        // РЈРЅРёРєР°Р»СЊРЅС‹Р№ С‡РёСЃР»РѕРІРѕР№ РёРґРµРЅС‚РёС„РёРєР°С‚РѕСЂ Р·Р°РїРёСЃРё.
	Title                     string  `json:"Title"`                     // РџРѕР»СЊР·РѕРІР°С‚РµР»СЊСЃРєРѕРµ РЅР°Р·РІР°РЅРёРµ РѕР±СЉРµРєС‚Р° РІ РёРіСЂРѕРІРѕРј РјРёСЂРµ.
	CosmicObjectModelID       int64   `json:"CosmicObjectModelID"`       // РњРѕРґРµР»СЊ, РѕС‚ РєРѕС‚РѕСЂРѕР№ РІР·СЏС‚С‹ Р±Р°Р·РѕРІС‹Рµ С…Р°СЂР°РєС‚РµСЂРёСЃС‚РёРєРё Рё РіСЂР°С„РёРєР°.
	OwnerCharacterID          int64   `json:"OwnerCharacterID"`          // РџРµСЂСЃРѕРЅР°Р¶-РІР»Р°РґРµР»РµС†, РµСЃР»Рё РѕР±СЉРµРєС‚ РїСЂРёРЅР°РґР»РµР¶РёС‚ РёРіСЂРѕРєСѓ.
	OwnerNpcClanID            int64   `json:"OwnerNpcClanID"`            // NPC-РєР»Р°РЅ-РІР»Р°РґРµР»РµС†, РµСЃР»Рё РѕР±СЉРµРєС‚ РЅРµ РїСЂРёРЅР°РґР»РµР¶РёС‚ РїРµСЂСЃРѕРЅР°Р¶Сѓ.
	CreatorCharacterID        int64   `json:"CreatorCharacterID"`        // РџРµСЂСЃРѕРЅР°Р¶, СЃРѕР·РґР°РІС€РёР№ РѕР±СЉРµРєС‚.
	Mass                      float64 `json:"Mass"`                      // РўРµРєСѓС‰Р°СЏ СЃСѓРјРјР°СЂРЅР°СЏ РјР°СЃСЃР° РѕР±СЉРµРєС‚Р° Рё СЃРѕРґРµСЂР¶РёРјРѕРіРѕ.
	Capacity                  float64 `json:"Capacity"`                  // РњР°РєСЃРёРјР°Р»СЊРЅС‹Р№ РѕР±СЉРµРј РѕР±РѕСЂСѓРґРѕРІР°РЅРёСЏ РёР»Рё СЃРѕРґРµСЂР¶РёРјРѕРіРѕ.
	MaxArmor                  float64 `json:"MaxArmor"`                  // Р’РµСЂС…РЅСЏСЏ РіСЂР°РЅРёС†Р° РїСЂРѕС‡РЅРѕСЃС‚Рё Р±СЂРѕРЅРё.
	MaxSpeed                  float64 `json:"MaxSpeed"`                  // РњР°РєСЃРёРјР°Р»СЊРЅР°СЏ Р»РёРЅРµР№РЅР°СЏ СЃРєРѕСЂРѕСЃС‚СЊ РІ РјРµС‚СЂР°С… Р·Р° СЃРµРєСѓРЅРґСѓ.
	MaxAngularSpeed           float64 `json:"MaxAngularSpeed"`           // РњР°РєСЃРёРјР°Р»СЊРЅР°СЏ СѓРіР»РѕРІР°СЏ СЃРєРѕСЂРѕСЃС‚СЊ РІ СЂР°РґРёР°РЅР°С… Р·Р° СЃРµРєСѓРЅРґСѓ.
	X                         float64 `json:"X"`                         // Р“РѕСЂРёР·РѕРЅС‚Р°Р»СЊРЅР°СЏ РєРѕРѕСЂРґРёРЅР°С‚Р° РїРѕР»РѕР¶РµРЅРёСЏ РІ РјРёСЂРµ.
	Y                         float64 `json:"Y"`                         // Р’РµСЂС‚РёРєР°Р»СЊРЅР°СЏ РєРѕРѕСЂРґРёРЅР°С‚Р° РїРѕР»РѕР¶РµРЅРёСЏ РІ РјРёСЂРµ.
	Rotation                  float64 `json:"Rotation"`                  // РўРµРєСѓС‰РёР№ СѓРіРѕР» РїРѕРІРѕСЂРѕС‚Р° РІ СЂР°РґРёР°РЅР°С… Р±РµР· РЅРѕСЂРјР°Р»РёР·Р°С†РёРё.
	Armor                     float64 `json:"Armor"`                     // РўРµРєСѓС‰РµРµ РєРѕР»РёС‡РµСЃС‚РІРѕ РµРґРёРЅРёС† Р±СЂРѕРЅРё.
	MaxAlongForce             float64 `json:"MaxAlongForce"`             // Р”РѕСЃС‚СѓРїРЅР°СЏ РїСЂРѕРґРѕР»СЊРЅР°СЏ СЃРёР»Р° С‚СЏРіРё.
	MaxAcrossForce            float64 `json:"MaxAcrossForce"`            // Р”РѕСЃС‚СѓРїРЅР°СЏ РїРѕРїРµСЂРµС‡РЅР°СЏ СЃРёР»Р° С‚СЏРіРё.
	MaxTorque                 float64 `json:"MaxTorque"`                 // Р”РѕСЃС‚СѓРїРЅС‹Р№ РєСЂСѓС‚СЏС‰РёР№ РјРѕРјРµРЅС‚.
	GeneratingPower           float64 `json:"GeneratingPower"`           // РЎСѓРјРјР°СЂРЅР°СЏ РІС‹СЂР°Р±Р°С‚С‹РІР°РµРјР°СЏ РјРѕС‰РЅРѕСЃС‚СЊ РѕР±РѕСЂСѓРґРѕРІР°РЅРёСЏ.
	ConsumingPower            float64 `json:"ConsumingPower"`            // РЎСѓРјРјР°СЂРЅР°СЏ РїРѕС‚СЂРµР±Р»СЏРµРјР°СЏ РјРѕС‰РЅРѕСЃС‚СЊ РѕР±РѕСЂСѓРґРѕРІР°РЅРёСЏ.
	AlongForce                float64 `json:"AlongForce"`                // Р¤Р°РєС‚РёС‡РµСЃРєРё РїСЂРёРјРµРЅРµРЅРЅР°СЏ РїСЂРѕРґРѕР»СЊРЅР°СЏ С‚СЏРіР° РЅР° С‚РµРєСѓС‰РµРј С€Р°РіРµ.
	AcrossForce               float64 `json:"AcrossForce"`               // Р¤Р°РєС‚РёС‡РµСЃРєРё РїСЂРёРјРµРЅРµРЅРЅР°СЏ РїРѕРїРµСЂРµС‡РЅР°СЏ С‚СЏРіР° РЅР° С‚РµРєСѓС‰РµРј С€Р°РіРµ.
	Torque                    float64 `json:"Torque"`                    // Р¤Р°РєС‚РёС‡РµСЃРєРё РїСЂРёРјРµРЅРµРЅРЅС‹Р№ РєСЂСѓС‚СЏС‰РёР№ РјРѕРјРµРЅС‚ РЅР° С‚РµРєСѓС‰РµРј С€Р°РіРµ.
	Enabled                   bool    `json:"Enabled"`                   // Р Р°Р·СЂРµС€Р°РµС‚ РѕР±СЉРµРєС‚Сѓ СЂР°Р±РѕС‚Р°С‚СЊ Рё СѓС‡Р°СЃС‚РІРѕРІР°С‚СЊ РІ СЃРёРјСѓР»СЏС†РёРё.
	LastReceivedDamageTime    int64   `json:"LastReceivedDamageTime"`    // Р’СЂРµРјСЏ РїРѕСЃР»РµРґРЅРµРіРѕ РїРѕР»СѓС‡РµРЅРёСЏ СѓСЂРѕРЅР° РґР»СЏ Р±РѕРµРІС‹С… Рё СЂРµРјРѕРЅС‚РЅС‹С… РїСЂР°РІРёР».
	Anchored                  bool    `json:"Anchored"`                  // Р—Р°РїСЂРµС‰Р°РµС‚ С„РёР·РёС‡РµСЃРєРѕРµ РїРµСЂРµРјРµС‰РµРЅРёРµ РѕР±СЉРµРєС‚Р°.
	Complexity                float64 `json:"Complexity"`                // РЎР»РѕР¶РЅРѕСЃС‚СЊ СѓСЃС‚СЂРѕР№СЃС‚РІР° РґР»СЏ РїСЂРѕРёР·РІРѕРґСЃС‚РІР° Рё РѕС†РµРЅРєРё СЃС‚РѕРёРјРѕСЃС‚Рё.
	OccupiedVolume            float64 `json:"OccupiedVolume"`            // РћР±СЉРµРј, СѓР¶Рµ Р·Р°РЅСЏС‚С‹Р№ СЃРѕРґРµСЂР¶РёРјС‹Рј РёР»Рё РѕР±РѕСЂСѓРґРѕРІР°РЅРёРµРј.
	MaxFuel                   float64 `json:"MaxFuel"`                   // РњР°РєСЃРёРјР°Р»СЊРЅС‹Р№ Р·Р°РїР°СЃ С‚РѕРїР»РёРІР°.
	Fuel                      float64 `json:"Fuel"`                      // РўРµРєСѓС‰РёР№ Р·Р°РїР°СЃ С‚РѕРїР»РёРІР°.
	Speed                     float64 `json:"Speed"`                     // РўРµРєСѓС‰Р°СЏ РґР»РёРЅР° РІРµРєС‚РѕСЂР° СЃРєРѕСЂРѕСЃС‚Рё.
	VelocityX                 float64 `json:"VelocityX"`                 // Р“РѕСЂРёР·РѕРЅС‚Р°Р»СЊРЅР°СЏ РєРѕРјРїРѕРЅРµРЅС‚Р° С‚РµРєСѓС‰РµР№ СЃРєРѕСЂРѕСЃС‚Рё.
	VelocityY                 float64 `json:"VelocityY"`                 // Р’РµСЂС‚РёРєР°Р»СЊРЅР°СЏ РєРѕРјРїРѕРЅРµРЅС‚Р° С‚РµРєСѓС‰РµР№ СЃРєРѕСЂРѕСЃС‚Рё.
	AngularSpeed              float64 `json:"AngularSpeed"`              // РўРµРєСѓС‰Р°СЏ СѓРіР»РѕРІР°СЏ СЃРєРѕСЂРѕСЃС‚СЊ РІ СЂР°РґРёР°РЅР°С… Р·Р° СЃРµРєСѓРЅРґСѓ.
	TargetRotation            float64 `json:"TargetRotation"`            // РЈРіРѕР», Рє РєРѕС‚РѕСЂРѕРјСѓ Р°РІС‚РѕРјР°С‚РёРєР° РїРѕРІРѕСЂРѕС‚Р° РІРµРґРµС‚ РѕР±СЉРµРєС‚.
	ClusterMainCosmicObjectID int64   `json:"ClusterMainCosmicObjectID"` // Р“Р»Р°РІРЅС‹Р№ РѕР±СЉРµРєС‚ РєР»Р°СЃС‚РµСЂР°, РµСЃР»Рё РѕР±СЉРµРєС‚ РїСЂРёСЃС‚С‹РєРѕРІР°РЅ.
}

// РҐСЂР°РЅРёС‚ РєРѕСЃРјРёС‡РµСЃРєРёРµ РѕР±СЉРµРєС‚С‹ Рё Р±С‹СЃС‚СЂС‹Рµ РёРЅРґРµРєСЃС‹ РїРѕ СЃРІСЏР·Р°РЅРЅС‹Рј РѕР±СЉРµРєС‚Р°Рј.
type CosmicObjects struct {
	MaxID int64                   `json:"MaxID"` // РџРѕСЃР»РµРґРЅРёР№ РІС‹РґР°РЅРЅС‹Р№ РёРґРµРЅС‚РёС„РёРєР°С‚РѕСЂ РґР»СЏ РЅРѕРІС‹С… Р·Р°РїРёСЃРµР№.
	Items map[int64]*CosmicObject `json:"Items"` // РћСЃРЅРѕРІРЅРѕРµ С…СЂР°РЅРёР»РёС‰Рµ Р·Р°РїРёСЃРµР№ РїРѕ С‡РёСЃР»РѕРІРѕРјСѓ РёРґРµРЅС‚РёС„РёРєР°С‚РѕСЂСѓ.

	ByCosmicObjectModelID map[int64]map[int64]*CosmicObject `json:"-"` // Р‘С‹СЃС‚СЂС‹Р№ РїРѕРёСЃРє РѕР±СЉРµРєС‚РѕРІ РїРѕ РјРѕРґРµР»Рё.
	ByOwnerCharacterID    map[int64]map[int64]*CosmicObject `json:"-"` // Р‘С‹СЃС‚СЂС‹Р№ РїРѕРёСЃРє РѕР±СЉРµРєС‚РѕРІ РїРѕ РїРµСЂСЃРѕРЅР°Р¶Сѓ-РІР»Р°РґРµР»СЊС†Сѓ.
}

// РЎРѕР·РґР°С‘С‚ РїСѓСЃС‚РѕРµ С…СЂР°РЅРёР»РёС‰Рµ РєРѕСЃРјРёС‡РµСЃРєРёС… РѕР±СЉРµРєС‚РѕРІ СЃ РїРѕРґРіРѕС‚РѕРІР»РµРЅРЅС‹РјРё РёРЅРґРµРєСЃР°РјРё.
func NewCosmicObjects() *CosmicObjects {
	cosmicObjects := &CosmicObjects{}
	cosmicObjects.ensureMaps()
	return cosmicObjects
}

// Р”РѕР±Р°РІР»СЏРµС‚ РЅРѕРІС‹Р№ РєРѕСЃРјРёС‡РµСЃРєРёР№ РѕР±СЉРµРєС‚ Рё РЅР°Р·РЅР°С‡Р°РµС‚ РЅРѕРІС‹Р№ ID.
func (cosmicObjects *CosmicObjects) Add(cosmicObject *CosmicObject) (*CosmicObject, error) {
	if cosmicObject == nil {
		return nil, errors.New("cosmic object is nil")
	}
	cosmicObjects.ensureMaps()
	if err := cosmicObjects.validateRequiredFields(cosmicObject); err != nil {
		return nil, err
	}

	cosmicObjects.MaxID++
	cosmicObject.ID = cosmicObjects.MaxID
	cosmicObjects.Items[cosmicObject.ID] = cosmicObject
	cosmicObjects.addIndexes(cosmicObject)
	return cosmicObject, nil
}

// Р’РѕР·РІСЂР°С‰Р°РµС‚ РєРѕСЃРјРёС‡РµСЃРєРёР№ РѕР±СЉРµРєС‚ РїРѕ ID.
func (cosmicObjects *CosmicObjects) Get(id int64) (*CosmicObject, bool) {
	cosmicObjects.ensureMaps()
	cosmicObject, ok := cosmicObjects.Items[id]
	return cosmicObject, ok
}

// РЈРґР°Р»СЏРµС‚ РєРѕСЃРјРёС‡РµСЃРєРёР№ РѕР±СЉРµРєС‚ Рё РІСЃРµ РµРіРѕ Р±С‹СЃС‚СЂС‹Рµ РёРЅРґРµРєСЃС‹.
func (cosmicObjects *CosmicObjects) Delete(id int64) bool {
	cosmicObjects.ensureMaps()
	cosmicObject, ok := cosmicObjects.Items[id]
	if !ok {
		return false
	}

	cosmicObjects.deleteIndexes(cosmicObject)
	delete(cosmicObjects.Items, id)
	return true
}

// Р’РѕР·РІСЂР°С‰Р°РµС‚ РѕР±СЉРµРєС‚С‹ СѓРєР°Р·Р°РЅРЅРѕР№ РјРѕРґРµР»Рё РІ РїРѕСЂСЏРґРєРµ РІРѕР·СЂР°СЃС‚Р°РЅРёСЏ ID.
func (cosmicObjects *CosmicObjects) GetByCosmicObjectModelID(cosmicObjectModelID int64) []*CosmicObject {
	cosmicObjects.ensureMaps()
	return sortedCosmicObjects(cosmicObjects.ByCosmicObjectModelID[cosmicObjectModelID])
}

// Р’РѕР·РІСЂР°С‰Р°РµС‚ РѕР±СЉРµРєС‚С‹ СѓРєР°Р·Р°РЅРЅРѕРіРѕ РІР»Р°РґРµР»СЊС†Р°-РїРµСЂСЃРѕРЅР°Р¶Р° РІ РїРѕСЂСЏРґРєРµ РІРѕР·СЂР°СЃС‚Р°РЅРёСЏ ID.
func (cosmicObjects *CosmicObjects) GetByOwnerCharacterID(ownerCharacterID int64) []*CosmicObject {
	cosmicObjects.ensureMaps()
	return sortedCosmicObjects(cosmicObjects.ByOwnerCharacterID[ownerCharacterID])
}

// РџРµСЂРµСЃРѕР±РёСЂР°РµС‚ Р±С‹СЃС‚СЂС‹Рµ РёРЅРґРµРєСЃС‹ РїРѕСЃР»Рµ Р·Р°РіСЂСѓР·РєРё РёР· JSON.
func (cosmicObjects *CosmicObjects) RebuildIndexes() error {
	cosmicObjects.ensureItems()
	cosmicObjects.ByCosmicObjectModelID = make(map[int64]map[int64]*CosmicObject)
	cosmicObjects.ByOwnerCharacterID = make(map[int64]map[int64]*CosmicObject)

	var maxID int64
	for id, cosmicObject := range cosmicObjects.Items {
		if cosmicObject == nil {
			return fmt.Errorf("cosmic object with ID %d is nil", id)
		}
		if cosmicObject.ID != id {
			return fmt.Errorf("cosmic object map key %d does not match object ID %d", id, cosmicObject.ID)
		}
		if err := cosmicObjects.validateRequiredFields(cosmicObject); err != nil {
			return fmt.Errorf("cosmic object with ID %d is invalid: %w", id, err)
		}
		if id > maxID {
			maxID = id
		}
		cosmicObjects.addIndexes(cosmicObject)
	}
	if cosmicObjects.MaxID < maxID {
		cosmicObjects.MaxID = maxID
	}
	return nil
}

// Р—Р°РіСЂСѓР¶Р°РµС‚ РєРѕСЃРјРёС‡РµСЃРєРёРµ РѕР±СЉРµРєС‚С‹ РёР· JSON-С„Р°Р№Р»Р° Рё РїРµСЂРµСЃРѕР±РёСЂР°РµС‚ Р±С‹СЃС‚СЂС‹Рµ РёРЅРґРµРєСЃС‹.
func (cosmicObjects *CosmicObjects) LoadFromFile(path string) error {
	content, err := os.ReadFile(path)
	if err != nil {
		return err
	}

	loaded := CosmicObjects{}
	if err := json.Unmarshal(content, &loaded); err != nil {
		return err
	}
	if err := loaded.RebuildIndexes(); err != nil {
		return err
	}

	*cosmicObjects = loaded
	return nil
}

// РЎРѕС…СЂР°РЅСЏРµС‚ РєРѕСЃРјРёС‡РµСЃРєРёРµ РѕР±СЉРµРєС‚С‹ РІ JSON-С„Р°Р№Р» Р±РµР· РІСЃРїРѕРјРѕРіР°С‚РµР»СЊРЅС‹С… РёРЅРґРµРєСЃРѕРІ.
func (cosmicObjects *CosmicObjects) SaveToFile(path string) error {
	cosmicObjects.ensureMaps()
	return saveTableWithOrderedItems(path, cosmicObjects.MaxID, cosmicObjects.Items)
}

// РџРѕРґРіРѕС‚Р°РІР»РёРІР°РµС‚ РѕСЃРЅРѕРІРЅРѕРµ С…СЂР°РЅРёР»РёС‰Рµ Рё РІСЃРµ РёРЅРґРµРєСЃС‹.
func (cosmicObjects *CosmicObjects) ensureMaps() {
	cosmicObjects.ensureItems()
	if cosmicObjects.ByCosmicObjectModelID == nil {
		cosmicObjects.ByCosmicObjectModelID = make(map[int64]map[int64]*CosmicObject)
	}
	if cosmicObjects.ByOwnerCharacterID == nil {
		cosmicObjects.ByOwnerCharacterID = make(map[int64]map[int64]*CosmicObject)
	}
}

// РџРѕРґРіРѕС‚Р°РІР»РёРІР°РµС‚ РѕСЃРЅРѕРІРЅСѓСЋ map РєРѕСЃРјРёС‡РµСЃРєРёС… РѕР±СЉРµРєС‚РѕРІ.
func (cosmicObjects *CosmicObjects) ensureItems() {
	if cosmicObjects.Items == nil {
		cosmicObjects.Items = make(map[int64]*CosmicObject)
	}
}

// РџСЂРѕРІРµСЂСЏРµС‚ РѕР±СЏР·Р°С‚РµР»СЊРЅС‹Рµ РїРѕР»СЏ РєРѕСЃРјРёС‡РµСЃРєРѕРіРѕ РѕР±СЉРµРєС‚Р°.
func (cosmicObjects *CosmicObjects) validateRequiredFields(cosmicObject *CosmicObject) error {
	if cosmicObject.CosmicObjectModelID <= 0 {
		return errors.New("cosmic object model ID is empty")
	}
	return nil
}

// Р”РѕР±Р°РІР»СЏРµС‚ РєРѕСЃРјРёС‡РµСЃРєРёР№ РѕР±СЉРµРєС‚ РІРѕ РІСЃРµ Р±С‹СЃС‚СЂС‹Рµ РёРЅРґРµРєСЃС‹.
func (cosmicObjects *CosmicObjects) addIndexes(cosmicObject *CosmicObject) {
	addCosmicObjectIndex(cosmicObjects.ByCosmicObjectModelID, cosmicObject.CosmicObjectModelID, cosmicObject)
	if cosmicObject.OwnerCharacterID > 0 {
		addCosmicObjectIndex(cosmicObjects.ByOwnerCharacterID, cosmicObject.OwnerCharacterID, cosmicObject)
	}
}

// РЈРґР°Р»СЏРµС‚ РєРѕСЃРјРёС‡РµСЃРєРёР№ РѕР±СЉРµРєС‚ РёР· РІСЃРµС… Р±С‹СЃС‚СЂС‹С… РёРЅРґРµРєСЃРѕРІ.
func (cosmicObjects *CosmicObjects) deleteIndexes(cosmicObject *CosmicObject) {
	deleteCosmicObjectIndex(cosmicObjects.ByCosmicObjectModelID, cosmicObject.CosmicObjectModelID, cosmicObject.ID)
	if cosmicObject.OwnerCharacterID > 0 {
		deleteCosmicObjectIndex(cosmicObjects.ByOwnerCharacterID, cosmicObject.OwnerCharacterID, cosmicObject.ID)
	}
}

// Р”РѕР±Р°РІР»СЏРµС‚ РѕР±СЉРµРєС‚ РІ РЅРµСѓРЅРёРєР°Р»СЊРЅС‹Р№ РёРЅРґРµРєСЃ.
func addCosmicObjectIndex(index map[int64]map[int64]*CosmicObject, key int64, cosmicObject *CosmicObject) {
	if index[key] == nil {
		index[key] = make(map[int64]*CosmicObject)
	}
	index[key][cosmicObject.ID] = cosmicObject
}

// РЈРґР°Р»СЏРµС‚ РѕР±СЉРµРєС‚ РёР· РЅРµСѓРЅРёРєР°Р»СЊРЅРѕРіРѕ РёРЅРґРµРєСЃР°.
func deleteCosmicObjectIndex(index map[int64]map[int64]*CosmicObject, key int64, id int64) {
	indexItems := index[key]
	if indexItems == nil {
		return
	}
	delete(indexItems, id)
	if len(indexItems) == 0 {
		delete(index, key)
	}
}

// Р’РѕР·РІСЂР°С‰Р°РµС‚ РѕР±СЉРµРєС‚С‹ РёРЅРґРµРєСЃР° РІ СЃС‚Р°Р±РёР»СЊРЅРѕРј РїРѕСЂСЏРґРєРµ ID.
func sortedCosmicObjects(indexItems map[int64]*CosmicObject) []*CosmicObject {
	if len(indexItems) == 0 {
		return []*CosmicObject{}
	}

	ids := make([]int64, 0, len(indexItems))
	for id := range indexItems {
		ids = append(ids, id)
	}
	sort.Slice(ids, func(left int, right int) bool {
		return ids[left] < ids[right]
	})

	result := make([]*CosmicObject, 0, len(ids))
	for _, id := range ids {
		result = append(result, indexItems[id])
	}
	return result
}
