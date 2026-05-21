package data

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"sort"
)

// РҐСЂР°РЅРёС‚ РіСЂСѓРїРїСѓ РѕР±РѕСЂСѓРґРѕРІР°РЅРёСЏ, СѓСЃС‚Р°РЅРѕРІР»РµРЅРЅРѕРіРѕ РЅР° РєРѕРЅРєСЂРµС‚РЅРѕРј РєРѕСЃРјРёС‡РµСЃРєРѕРј РѕР±СЉРµРєС‚Рµ.
type EquipmentGroup struct {
	ID                          int64  `json:"ID"`                          // РЈРЅРёРєР°Р»СЊРЅС‹Р№ С‡РёСЃР»РѕРІРѕР№ РёРґРµРЅС‚РёС„РёРєР°С‚РѕСЂ Р·Р°РїРёСЃРё.
	CosmicObjectID              int64  `json:"CosmicObjectID"`              // РљРѕСЃРјРёС‡РµСЃРєРёР№ РѕР±СЉРµРєС‚, РЅР° РєРѕС‚РѕСЂРѕРј СѓСЃС‚Р°РЅРѕРІР»РµРЅРѕ РѕР±РѕСЂСѓРґРѕРІР°РЅРёРµ.
	Title                       string `json:"Title"`                       // РќР°Р·РІР°РЅРёРµ РіСЂСѓРїРїС‹ СѓСЃС‚Р°РЅРѕРІР»РµРЅРЅРѕРіРѕ РѕР±РѕСЂСѓРґРѕРІР°РЅРёСЏ.
	EquipmentItemModelID        int64  `json:"EquipmentItemModelID"`        // РњРѕРґРµР»СЊ СѓСЃС‚Р°РЅРѕРІР»РµРЅРЅРѕРіРѕ РѕР±РѕСЂСѓРґРѕРІР°РЅРёСЏ.
	Count                       int64  `json:"Count"`                       // РљРѕР»РёС‡РµСЃС‚РІРѕ СѓСЃС‚Р°РЅРѕРІР»РµРЅРЅС‹С… РµРґРёРЅРёС† РѕР±РѕСЂСѓРґРѕРІР°РЅРёСЏ.
	EnabledCount                int64  `json:"EnabledCount"`                // РљРѕР»РёС‡РµСЃС‚РІРѕ РІРєР»СЋС‡РµРЅРЅС‹С… РµРґРёРЅРёС† РѕР±РѕСЂСѓРґРѕРІР°РЅРёСЏ.
	Enabled                     bool   `json:"Enabled"`                     // РџСЂРёР·РЅР°Рє РІРєР»СЋС‡РµРЅРёСЏ РіСЂСѓРїРїС‹ РѕР±РѕСЂСѓРґРѕРІР°РЅРёСЏ.
	Active                      bool   `json:"Active"`                      // РџСЂРёР·РЅР°Рє РІС‹РїРѕР»РЅРµРЅРёСЏ СЂР°Р±РѕС‚С‹ РІРєР»СЋС‡РµРЅРЅС‹Рј РѕР±РѕСЂСѓРґРѕРІР°РЅРёРµРј.
	LastRechargeStartTime       int64  `json:"LastRechargeStartTime"`       // Р’СЂРµРјСЏ РЅР°С‡Р°Р»Р° РїРѕСЃР»РµРґРЅРµР№ РїРµСЂРµР·Р°СЂСЏРґРєРё РІ РјРёР»Р»РёСЃРµРєСѓРЅРґР°С… Unix.
	SourceEquipmentGroupID      int64  `json:"SourceEquipmentGroupID"`      // Источник материалов или груза для работы оборудования.
	DestinationEquipmentGroupID int64  `json:"DestinationEquipmentGroupID"` // Приемник результата или груза после работы оборудования.
	OppositeEquipmentGroupID    int64  `json:"OppositeEquipmentGroupID"`    // Противоположная группа оборудования в парном интерфейсе использования.
}

// РҐСЂР°РЅРёС‚ РіСЂСѓРїРїС‹ РѕР±РѕСЂСѓРґРѕРІР°РЅРёСЏ РєРѕСЃРјРёС‡РµСЃРєРёС… РѕР±СЉРµРєС‚РѕРІ.
type EquipmentGroups struct {
	MaxID int64                     `json:"MaxID"` // РџРѕСЃР»РµРґРЅРёР№ РІС‹РґР°РЅРЅС‹Р№ С‡РёСЃР»РѕРІРѕР№ РёРґРµРЅС‚РёС„РёРєР°С‚РѕСЂ Р·Р°РїРёСЃРµР№.
	Items map[int64]*EquipmentGroup `json:"Items"` // РћСЃРЅРѕРІРЅРѕРµ С…СЂР°РЅРёР»РёС‰Рµ Р·Р°РїРёСЃРµР№ РїРѕ С‡РёСЃР»РѕРІРѕРјСѓ РёРґРµРЅС‚РёС„РёРєР°С‚РѕСЂСѓ.

	ByCosmicObjectID map[int64][]*EquipmentGroup `json:"-"` // Р‘С‹СЃС‚СЂС‹Р№ РїРѕРёСЃРє РіСЂСѓРїРї РѕР±РѕСЂСѓРґРѕРІР°РЅРёСЏ РїРѕ РѕР±СЉРµРєС‚Сѓ.
}

// РЎРѕР·РґР°РµС‚ РїСѓСЃС‚РѕРµ С…СЂР°РЅРёР»РёС‰Рµ РіСЂСѓРїРї РѕР±РѕСЂСѓРґРѕРІР°РЅРёСЏ РѕР±СЉРµРєС‚РѕРІ.
func NewEquipmentGroups() *EquipmentGroups {
	groups := &EquipmentGroups{}
	groups.ensureMaps()
	return groups
}

// Р”РѕР±Р°РІР»СЏРµС‚ РЅРѕРІСѓСЋ РіСЂСѓРїРїСѓ РѕР±РѕСЂСѓРґРѕРІР°РЅРёСЏ Рё РЅР°Р·РЅР°С‡Р°РµС‚ РЅРѕРІС‹Р№ ID.
func (groups *EquipmentGroups) Add(group *EquipmentGroup) (*EquipmentGroup, error) {
	if group == nil {
		return nil, errors.New("equipment group is nil")
	}
	groups.ensureMaps()
	if err := groups.validateRequiredFields(group); err != nil {
		return nil, err
	}

	groups.MaxID++
	group.ID = groups.MaxID
	groups.Items[group.ID] = group
	groups.addIndexes(group)
	return group, nil
}

// Р’РѕР·РІСЂР°С‰Р°РµС‚ РіСЂСѓРїРїСѓ РѕР±РѕСЂСѓРґРѕРІР°РЅРёСЏ РїРѕ ID.
func (groups *EquipmentGroups) Get(id int64) (*EquipmentGroup, bool) {
	groups.ensureMaps()
	group, ok := groups.Items[id]
	return group, ok
}

// Р’РѕР·РІСЂР°С‰Р°РµС‚ РіСЂСѓРїРїС‹ РѕР±РѕСЂСѓРґРѕРІР°РЅРёСЏ СѓРєР°Р·Р°РЅРЅРѕРіРѕ РєРѕСЃРјРёС‡РµСЃРєРѕРіРѕ РѕР±СЉРµРєС‚Р°.
func (groups *EquipmentGroups) GetByCosmicObjectID(cosmicObjectID int64) []*EquipmentGroup {
	groups.ensureMaps()
	return groups.ByCosmicObjectID[cosmicObjectID]
}

// РЈРґР°Р»СЏРµС‚ РІСЃРµ РіСЂСѓРїРїС‹ РѕР±РѕСЂСѓРґРѕРІР°РЅРёСЏ СѓРєР°Р·Р°РЅРЅРѕРіРѕ РєРѕСЃРјРёС‡РµСЃРєРѕРіРѕ РѕР±СЉРµРєС‚Р°.
func (groups *EquipmentGroups) DeleteByCosmicObjectID(cosmicObjectID int64) {
	groups.ensureMaps()
	for _, group := range groups.ByCosmicObjectID[cosmicObjectID] {
		delete(groups.Items, group.ID)
	}
	delete(groups.ByCosmicObjectID, cosmicObjectID)
}

// РџРµСЂРµСЃРѕР±РёСЂР°РµС‚ С…СЂР°РЅРёР»РёС‰Рµ РїРѕСЃР»Рµ Р·Р°РіСЂСѓР·РєРё РёР· JSON.
func (groups *EquipmentGroups) RebuildIndexes() error {
	groups.ensureItems()
	groups.ByCosmicObjectID = make(map[int64][]*EquipmentGroup)

	var maxID int64
	ids := make([]int64, 0, len(groups.Items))
	for id := range groups.Items {
		ids = append(ids, id)
	}
	sort.Slice(ids, func(left int, right int) bool {
		return ids[left] < ids[right]
	})
	for _, id := range ids {
		group := groups.Items[id]
		if group == nil {
			return fmt.Errorf("equipment group with ID %d is nil", id)
		}
		if group.ID != id {
			return fmt.Errorf("equipment group map key %d does not match group ID %d", id, group.ID)
		}
		if err := groups.validateRequiredFields(group); err != nil {
			return fmt.Errorf("equipment group with ID %d is invalid: %w", id, err)
		}
		if id > maxID {
			maxID = id
		}
		groups.addIndexes(group)
	}
	if groups.MaxID < maxID {
		groups.MaxID = maxID
	}
	return nil
}

// Р—Р°РіСЂСѓР¶Р°РµС‚ РіСЂСѓРїРїС‹ РѕР±РѕСЂСѓРґРѕРІР°РЅРёСЏ РѕР±СЉРµРєС‚РѕРІ РёР· JSON-С„Р°Р№Р»Р°.
func (groups *EquipmentGroups) LoadFromFile(path string) error {
	content, err := os.ReadFile(path)
	if err != nil {
		return err
	}

	loaded := EquipmentGroups{}
	if err := json.Unmarshal(content, &loaded); err != nil {
		return err
	}
	if err := loaded.RebuildIndexes(); err != nil {
		return err
	}

	*groups = loaded
	return nil
}

// РЎРѕС…СЂР°РЅСЏРµС‚ РіСЂСѓРїРїС‹ РѕР±РѕСЂСѓРґРѕРІР°РЅРёСЏ РѕР±СЉРµРєС‚РѕРІ РІ JSON-С„Р°Р№Р».
func (groups *EquipmentGroups) SaveToFile(path string) error {
	groups.ensureMaps()
	return saveTableWithOrderedItems(path, groups.MaxID, groups.Items)
}

// РџРѕРґРіРѕС‚Р°РІР»РёРІР°РµС‚ РѕСЃРЅРѕРІРЅРѕРµ С…СЂР°РЅРёР»РёС‰Рµ Рё Р±С‹СЃС‚СЂС‹Рµ РёРЅРґРµРєСЃС‹.
func (groups *EquipmentGroups) ensureMaps() {
	groups.ensureItems()
	if groups.ByCosmicObjectID == nil {
		groups.ByCosmicObjectID = make(map[int64][]*EquipmentGroup)
	}
}

// РџРѕРґРіРѕС‚Р°РІР»РёРІР°РµС‚ РѕСЃРЅРѕРІРЅСѓСЋ map.
func (groups *EquipmentGroups) ensureItems() {
	if groups.Items == nil {
		groups.Items = make(map[int64]*EquipmentGroup)
	}
}

// РџСЂРѕРІРµСЂСЏРµС‚ РѕР±СЏР·Р°С‚РµР»СЊРЅС‹Рµ РїРѕР»СЏ РіСЂСѓРїРїС‹ РѕР±РѕСЂСѓРґРѕРІР°РЅРёСЏ.
func (groups *EquipmentGroups) validateRequiredFields(group *EquipmentGroup) error {
	if group.CosmicObjectID <= 0 {
		return errors.New("cosmic object ID is empty")
	}
	if group.EquipmentItemModelID <= 0 {
		return errors.New("equipment item model ID is empty")
	}
	if group.Count < 0 {
		return errors.New("count is negative")
	}
	if group.EnabledCount < 0 {
		return errors.New("enabled count is negative")
	}
	if group.EnabledCount > group.Count {
		return errors.New("enabled count is greater than count")
	}
	return nil
}

// Р”РѕР±Р°РІР»СЏРµС‚ РіСЂСѓРїРїСѓ РѕР±РѕСЂСѓРґРѕРІР°РЅРёСЏ РІ Р±С‹СЃС‚СЂС‹Рµ РёРЅРґРµРєСЃС‹.
func (groups *EquipmentGroups) addIndexes(group *EquipmentGroup) {
	groups.ByCosmicObjectID[group.CosmicObjectID] = append(groups.ByCosmicObjectID[group.CosmicObjectID], group)
}
