package data

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"sort"
)

// РҐСЂР°РЅРёС‚ РіСЂСѓРїРїСѓ РѕРґРёРЅР°РєРѕРІС‹С… РїСЂРµРґРјРµС‚РѕРІ РІРЅСѓС‚СЂРё РєРѕРЅС‚РµР№РЅРµСЂР°.
type ItemGroup struct {
	ID                        int64   `json:"ID"`                        // РЈРЅРёРєР°Р»СЊРЅС‹Р№ С‡РёСЃР»РѕРІРѕР№ РёРґРµРЅС‚РёС„РёРєР°С‚РѕСЂ Р·Р°РїРёСЃРё.
	ContainerEquipmentGroupID int64   `json:"ContainerEquipmentGroupID"` // Р“СЂСѓРїРїР° РѕР±РѕСЂСѓРґРѕРІР°РЅРёСЏ, РІРЅСѓС‚СЂРё РєРѕС‚РѕСЂРѕР№ РЅР°С…РѕРґРёС‚СЃСЏ СЃРѕРґРµСЂР¶РёРјРѕРµ.
	ContentItemModelID        int64   `json:"ContentItemModelID"`        // РњРѕРґРµР»СЊ РїСЂРµРґРјРµС‚Р°, Р»РµР¶Р°С‰РµРіРѕ РІРЅСѓС‚СЂРё РєРѕРЅС‚РµР№РЅРµСЂР°.
	Count                     float64 `json:"Count"`                     // РљРѕР»РёС‡РµСЃС‚РІРѕ РїСЂРµРґРјРµС‚РѕРІ СѓРєР°Р·Р°РЅРЅРѕР№ РјРѕРґРµР»Рё.
}

// РҐСЂР°РЅРёС‚ РіСЂСѓРїРїС‹ РїСЂРµРґРјРµС‚РѕРІ РІРЅСѓС‚СЂРё СѓСЃС‚Р°РЅРѕРІР»РµРЅРЅРѕРіРѕ РєРѕРЅС‚РµР№РЅРµСЂРЅРѕРіРѕ РѕР±РѕСЂСѓРґРѕРІР°РЅРёСЏ.
type ItemGroups struct {
	MaxID int64                `json:"MaxID"` // РџРѕСЃР»РµРґРЅРёР№ РІС‹РґР°РЅРЅС‹Р№ С‡РёСЃР»РѕРІРѕР№ РёРґРµРЅС‚РёС„РёРєР°С‚РѕСЂ РґР»СЏ РЅРѕРІС‹С… Р·Р°РїРёСЃРµР№.
	Items map[int64]*ItemGroup `json:"Items"` // РћСЃРЅРѕРІРЅРѕРµ С…СЂР°РЅРёР»РёС‰Рµ Р·Р°РїРёСЃРµР№ РїРѕ С‡РёСЃР»РѕРІРѕРјСѓ РёРґРµРЅС‚РёС„РёРєР°С‚РѕСЂСѓ.

	ByContainerEquipmentGroupID map[int64][]*ItemGroup `json:"-"` // Р‘С‹СЃС‚СЂС‹Р№ РїРѕРёСЃРє СЃРѕРґРµСЂР¶РёРјРѕРіРѕ РїРѕ РіСЂСѓРїРїРµ РєРѕРЅС‚РµР№РЅРµСЂРЅРѕРіРѕ РѕР±РѕСЂСѓРґРѕРІР°РЅРёСЏ.
}

// РЎРѕР·РґР°РµС‚ РїСѓСЃС‚РѕРµ С…СЂР°РЅРёР»РёС‰Рµ РіСЂСѓРїРї РїСЂРµРґРјРµС‚РѕРІ СЃ РїРѕРґРіРѕС‚РѕРІР»РµРЅРЅС‹РјРё РёРЅРґРµРєСЃР°РјРё.
func NewItemGroups() *ItemGroups {
	groups := &ItemGroups{}
	groups.ensureMaps()
	return groups
}

// Р”РѕР±Р°РІР»СЏРµС‚ РЅРѕРІСѓСЋ РіСЂСѓРїРїСѓ РїСЂРµРґРјРµС‚РѕРІ Рё РЅР°Р·РЅР°С‡Р°РµС‚ РЅРѕРІС‹Р№ ID.
func (groups *ItemGroups) Add(group *ItemGroup) (*ItemGroup, error) {
	if group == nil {
		return nil, errors.New("item group is nil")
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

// Р’РѕР·РІСЂР°С‰Р°РµС‚ РіСЂСѓРїРїСѓ РїСЂРµРґРјРµС‚РѕРІ РїРѕ ID.
func (groups *ItemGroups) Get(id int64) (*ItemGroup, bool) {
	groups.ensureMaps()
	group, ok := groups.Items[id]
	return group, ok
}

// Р’РѕР·РІСЂР°С‰Р°РµС‚ СЃРѕРґРµСЂР¶РёРјРѕРµ СѓРєР°Р·Р°РЅРЅРѕР№ РіСЂСѓРїРїС‹ РєРѕРЅС‚РµР№РЅРµСЂРЅРѕРіРѕ РѕР±РѕСЂСѓРґРѕРІР°РЅРёСЏ.
func (groups *ItemGroups) GetByContainerEquipmentGroupID(containerEquipmentGroupID int64) []*ItemGroup {
	groups.ensureMaps()
	return groups.ByContainerEquipmentGroupID[containerEquipmentGroupID]
}

// РЈРґР°Р»СЏРµС‚ РІСЃРµ СЃРѕРґРµСЂР¶РёРјРѕРµ СѓРєР°Р·Р°РЅРЅРѕР№ РіСЂСѓРїРїС‹ РєРѕРЅС‚РµР№РЅРµСЂРЅРѕРіРѕ РѕР±РѕСЂСѓРґРѕРІР°РЅРёСЏ.
func (groups *ItemGroups) DeleteByContainerEquipmentGroupID(containerEquipmentGroupID int64) {
	groups.ensureMaps()
	for _, group := range groups.ByContainerEquipmentGroupID[containerEquipmentGroupID] {
		delete(groups.Items, group.ID)
	}
	delete(groups.ByContainerEquipmentGroupID, containerEquipmentGroupID)
}

// РЈРґР°Р»СЏРµС‚ СЃРѕРґРµСЂР¶РёРјРѕРµ РЅРµСЃРєРѕР»СЊРєРёС… РіСЂСѓРїРї РєРѕРЅС‚РµР№РЅРµСЂРЅРѕРіРѕ РѕР±РѕСЂСѓРґРѕРІР°РЅРёСЏ.
func (groups *ItemGroups) DeleteByContainerEquipmentGroupIDs(containerEquipmentGroupIDs []int64) {
	for _, containerEquipmentGroupID := range containerEquipmentGroupIDs {
		groups.DeleteByContainerEquipmentGroupID(containerEquipmentGroupID)
	}
}

// РџРµСЂРµСЃРѕР±РёСЂР°РµС‚ Р±С‹СЃС‚СЂС‹Рµ РёРЅРґРµРєСЃС‹ РїРѕСЃР»Рµ Р·Р°РіСЂСѓР·РєРё РёР· JSON.
func (groups *ItemGroups) RebuildIndexes() error {
	groups.ensureItems()
	groups.ByContainerEquipmentGroupID = make(map[int64][]*ItemGroup)

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
			return fmt.Errorf("item group with ID %d is nil", id)
		}
		if group.ID != id {
			return fmt.Errorf("item group map key %d does not match group ID %d", id, group.ID)
		}
		if err := groups.validateRequiredFields(group); err != nil {
			return fmt.Errorf("item group with ID %d is invalid: %w", id, err)
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

// Р—Р°РіСЂСѓР¶Р°РµС‚ РіСЂСѓРїРїС‹ РїСЂРµРґРјРµС‚РѕРІ РёР· JSON-С„Р°Р№Р»Р° Рё РїРµСЂРµСЃРѕР±РёСЂР°РµС‚ Р±С‹СЃС‚СЂС‹Рµ РёРЅРґРµРєСЃС‹.
func (groups *ItemGroups) LoadFromFile(path string) error {
	content, err := os.ReadFile(path)
	if err != nil {
		return err
	}

	loaded := ItemGroups{}
	if err := json.Unmarshal(content, &loaded); err != nil {
		return err
	}
	if err := loaded.RebuildIndexes(); err != nil {
		return err
	}

	*groups = loaded
	return nil
}

// РЎРѕС…СЂР°РЅСЏРµС‚ РіСЂСѓРїРїС‹ РїСЂРµРґРјРµС‚РѕРІ РІ JSON-С„Р°Р№Р» Р±РµР· РІСЃРїРѕРјРѕРіР°С‚РµР»СЊРЅС‹С… РёРЅРґРµРєСЃРѕРІ.
func (groups *ItemGroups) SaveToFile(path string) error {
	groups.ensureMaps()
	return saveTableWithOrderedItems(path, groups.MaxID, groups.Items)
}

// РџРѕРґРіРѕС‚Р°РІР»РёРІР°РµС‚ РѕСЃРЅРѕРІРЅРѕРµ С…СЂР°РЅРёР»РёС‰Рµ Рё Р±С‹СЃС‚СЂС‹Рµ РёРЅРґРµРєСЃС‹.
func (groups *ItemGroups) ensureMaps() {
	groups.ensureItems()
	if groups.ByContainerEquipmentGroupID == nil {
		groups.ByContainerEquipmentGroupID = make(map[int64][]*ItemGroup)
	}
}

// РџРѕРґРіРѕС‚Р°РІР»РёРІР°РµС‚ РѕСЃРЅРѕРІРЅСѓСЋ map.
func (groups *ItemGroups) ensureItems() {
	if groups.Items == nil {
		groups.Items = make(map[int64]*ItemGroup)
	}
}

// РџСЂРѕРІРµСЂСЏРµС‚ РѕР±СЏР·Р°С‚РµР»СЊРЅС‹Рµ РїРѕР»СЏ РіСЂСѓРїРїС‹ РїСЂРµРґРјРµС‚РѕРІ.
func (groups *ItemGroups) validateRequiredFields(group *ItemGroup) error {
	if group.ContainerEquipmentGroupID <= 0 {
		return errors.New("container equipment group ID is empty")
	}
	if group.ContentItemModelID <= 0 {
		return errors.New("content item model ID is empty")
	}
	if group.Count <= 0 {
		return errors.New("count is empty")
	}
	return nil
}

// Р”РѕР±Р°РІР»СЏРµС‚ РіСЂСѓРїРїСѓ РїСЂРµРґРјРµС‚РѕРІ РІ Р±С‹СЃС‚СЂС‹Рµ РёРЅРґРµРєСЃС‹.
func (groups *ItemGroups) addIndexes(group *ItemGroup) {
	groups.ByContainerEquipmentGroupID[group.ContainerEquipmentGroupID] = append(groups.ByContainerEquipmentGroupID[group.ContainerEquipmentGroupID], group)
}
