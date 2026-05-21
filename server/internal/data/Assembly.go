package data

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"sort"
)

// РҐСЂР°РЅРёС‚ РёС‚РѕРіРѕРІС‹Рµ С…Р°СЂР°РєС‚РµСЂРёСЃС‚РёРєРё РјРѕРґРµР»Рё РѕР±СЉРµРєС‚Р° СЃ СѓСЃС‚Р°РЅРѕРІР»РµРЅРЅС‹Рј РѕР±РѕСЂСѓРґРѕРІР°РЅРёРµРј.
type Assembly struct {
	ID                  int64   `json:"ID"`                  // РЈРЅРёРєР°Р»СЊРЅС‹Р№ С‡РёСЃР»РѕРІРѕР№ РёРґРµРЅС‚РёС„РёРєР°С‚РѕСЂ Р·Р°РїРёСЃРё.
	AuthorCharacterID   int64   `json:"AuthorCharacterID"`   // РђРІС‚РѕСЂСЃРєР°СЏ РїСЂРёРІСЏР·РєР°, СЂР°РІРЅР°СЏ РЅСѓР»СЋ РґР»СЏ СЃРёСЃС‚РµРјРЅС‹С… Р·Р°РїРёСЃРµР№ СЂР°Р·СЂР°Р±РѕС‚С‡РёРєРѕРІ.
	Title               string  `json:"Title"`               // Р§РµР»РѕРІРµРєРѕС‡РёС‚Р°РµРјРѕРµ РЅР°Р·РІР°РЅРёРµ РІР°СЂРёР°РЅС‚Р° РѕСЃРЅР°С‰РµРЅРёСЏ.
	CosmicObjectModelID int64   `json:"CosmicObjectModelID"` // Р‘Р°Р·РѕРІР°СЏ РјРѕРґРµР»СЊ РєРѕСЂРїСѓСЃР°, РґР»СЏ РєРѕС‚РѕСЂРѕР№ СЂР°СЃСЃС‡РёС‚Р°РЅС‹ С…Р°СЂР°РєС‚РµСЂРёСЃС‚РёРєРё.
	IsPublic            bool    `json:"IsPublic"`            // Р”РѕСЃС‚СѓРїРЅРѕСЃС‚СЊ РІР°СЂРёР°РЅС‚Р° РґР»СЏ РѕР±С‰РµРіРѕ РёСЃРїРѕР»СЊР·РѕРІР°РЅРёСЏ.
	Mass                float64 `json:"Mass"`                // РС‚РѕРіРѕРІР°СЏ РјР°СЃСЃР° РєРѕСЂРїСѓСЃР° Рё РѕР±РѕСЂСѓРґРѕРІР°РЅРёСЏ РІ РєРёР»РѕРіСЂР°РјРјР°С….
	MaxArmor            float64 `json:"MaxArmor"`            // РС‚РѕРіРѕРІР°СЏ РјР°РєСЃРёРјР°Р»СЊРЅР°СЏ Р±СЂРѕРЅСЏ.
	MaxAlongForce       float64 `json:"MaxAlongForce"`       // РС‚РѕРіРѕРІР°СЏ РґРѕСЃС‚СѓРїРЅР°СЏ РїСЂРѕРґРѕР»СЊРЅР°СЏ С‚СЏРіР°.
	MaxAcrossForce      float64 `json:"MaxAcrossForce"`      // РС‚РѕРіРѕРІР°СЏ РґРѕСЃС‚СѓРїРЅР°СЏ РїРѕРїРµСЂРµС‡РЅР°СЏ С‚СЏРіР°.
	MaxTorque           float64 `json:"MaxTorque"`           // РС‚РѕРіРѕРІС‹Р№ РґРѕСЃС‚СѓРїРЅС‹Р№ РєСЂСѓС‚СЏС‰РёР№ РјРѕРјРµРЅС‚.
	GeneratingPower     float64 `json:"GeneratingPower"`     // РС‚РѕРіРѕРІР°СЏ РІС‹СЂР°Р±Р°С‚С‹РІР°РµРјР°СЏ РјРѕС‰РЅРѕСЃС‚СЊ.
	ConsumingPower      float64 `json:"ConsumingPower"`      // РС‚РѕРіРѕРІР°СЏ РїРѕС‚СЂРµР±Р»СЏРµРјР°СЏ РјРѕС‰РЅРѕСЃС‚СЊ.
	Complexity          float64 `json:"Complexity"`          // РС‚РѕРіРѕРІР°СЏ СЃР»РѕР¶РЅРѕСЃС‚СЊ РїСЂРѕРёР·РІРѕРґСЃС‚РІР°.
	OccupiedVolume      float64 `json:"OccupiedVolume"`      // РћР±СЉРµРј, Р·Р°РЅСЏС‚С‹Р№ РѕР±РѕСЂСѓРґРѕРІР°РЅРёРµРј.
	MaxFuel             float64 `json:"MaxFuel"`             // РС‚РѕРіРѕРІР°СЏ РјР°РєСЃРёРјР°Р»СЊРЅР°СЏ РІРјРµСЃС‚РёРјРѕСЃС‚СЊ С‚РѕРїР»РёРІР°.
}

// РҐСЂР°РЅРёС‚ РІР°СЂРёР°РЅС‚С‹ РѕСЃРЅР°С‰РµРЅРёСЏ Рё Р±С‹СЃС‚СЂС‹Р№ РїРѕРёСЃРє РїРѕ РјРѕРґРµР»Рё РєРѕСЂРїСѓСЃР°.
type Assemblies struct {
	MaxID int64               `json:"MaxID"` // РџРѕСЃР»РµРґРЅРёР№ РІС‹РґР°РЅРЅС‹Р№ С‡РёСЃР»РѕРІРѕР№ РёРґРµРЅС‚РёС„РёРєР°С‚РѕСЂ Р·Р°РїРёСЃРµР№.
	Items map[int64]*Assembly `json:"Items"` // РћСЃРЅРѕРІРЅРѕРµ С…СЂР°РЅРёР»РёС‰Рµ Р·Р°РїРёСЃРµР№ РїРѕ С‡РёСЃР»РѕРІРѕРјСѓ РёРґРµРЅС‚РёС„РёРєР°С‚РѕСЂСѓ.

	ByCosmicObjectModelID map[int64][]*Assembly `json:"-"` // Р‘С‹СЃС‚СЂС‹Р№ РїРѕРёСЃРє РІР°СЂРёР°РЅС‚РѕРІ РїРѕ Р±Р°Р·РѕРІРѕР№ РјРѕРґРµР»Рё РєРѕСЂРїСѓСЃР°.
}

// РЎРѕР·РґР°РµС‚ РїСѓСЃС‚РѕРµ С…СЂР°РЅРёР»РёС‰Рµ РІР°СЂРёР°РЅС‚РѕРІ РѕСЃРЅР°С‰РµРЅРёСЏ СЃ РїРѕРґРіРѕС‚РѕРІР»РµРЅРЅС‹РјРё РёРЅРґРµРєСЃР°РјРё.
func NewAssemblies() *Assemblies {
	assemblies := &Assemblies{}
	assemblies.ensureMaps()
	return assemblies
}

// Р”РѕР±Р°РІР»СЏРµС‚ РЅРѕРІС‹Р№ РІР°СЂРёР°РЅС‚ РѕСЃРЅР°С‰РµРЅРёСЏ Рё РЅР°Р·РЅР°С‡Р°РµС‚ РЅРѕРІС‹Р№ ID.
func (assemblies *Assemblies) Add(assembly *Assembly) (*Assembly, error) {
	if assembly == nil {
		return nil, errors.New("assembly is nil")
	}
	assemblies.ensureMaps()
	if err := assemblies.validateRequiredFields(assembly); err != nil {
		return nil, err
	}

	assemblies.MaxID++
	assembly.ID = assemblies.MaxID
	assemblies.Items[assembly.ID] = assembly
	assemblies.addIndexes(assembly)
	return assembly, nil
}

// Р’РѕР·РІСЂР°С‰Р°РµС‚ РІР°СЂРёР°РЅС‚ РѕСЃРЅР°С‰РµРЅРёСЏ РїРѕ ID.
func (assemblies *Assemblies) Get(id int64) (*Assembly, bool) {
	assemblies.ensureMaps()
	assembly, ok := assemblies.Items[id]
	return assembly, ok
}

// Р’РѕР·РІСЂР°С‰Р°РµС‚ РїРµСЂРІСѓСЋ РїСѓР±Р»РёС‡РЅСѓСЋ СЃРёСЃС‚РµРјРЅСѓСЋ СЃР±РѕСЂРєСѓ РґР»СЏ СѓРєР°Р·Р°РЅРЅРѕР№ РјРѕРґРµР»Рё.
func (assemblies *Assemblies) FirstPublicDeveloperAssembly(cosmicObjectModelID int64) (*Assembly, bool) {
	assemblies.ensureMaps()
	for _, assembly := range assemblies.ByCosmicObjectModelID[cosmicObjectModelID] {
		if assembly.AuthorCharacterID == 0 && assembly.IsPublic {
			return assembly, true
		}
	}
	return nil, false
}

// РџРµСЂРµСЃРѕР±РёСЂР°РµС‚ Р±С‹СЃС‚СЂС‹Рµ РёРЅРґРµРєСЃС‹ РїРѕСЃР»Рµ Р·Р°РіСЂСѓР·РєРё РёР· JSON.
func (assemblies *Assemblies) RebuildIndexes() error {
	assemblies.ensureItems()
	assemblies.ByCosmicObjectModelID = make(map[int64][]*Assembly)

	var maxID int64
	ids := make([]int64, 0, len(assemblies.Items))
	for id := range assemblies.Items {
		ids = append(ids, id)
	}
	sort.Slice(ids, func(left int, right int) bool {
		return ids[left] < ids[right]
	})
	for _, id := range ids {
		assembly := assemblies.Items[id]
		if assembly == nil {
			return fmt.Errorf("assembly with ID %d is nil", id)
		}
		if assembly.ID != id {
			return fmt.Errorf("assembly map key %d does not match assembly ID %d", id, assembly.ID)
		}
		if err := assemblies.validateRequiredFields(assembly); err != nil {
			return fmt.Errorf("assembly with ID %d is invalid: %w", id, err)
		}
		if id > maxID {
			maxID = id
		}
		assemblies.addIndexes(assembly)
	}
	if assemblies.MaxID < maxID {
		assemblies.MaxID = maxID
	}
	return nil
}

// Р—Р°РіСЂСѓР¶Р°РµС‚ РІР°СЂРёР°РЅС‚С‹ РѕСЃРЅР°С‰РµРЅРёСЏ РёР· JSON-С„Р°Р№Р»Р°.
func (assemblies *Assemblies) LoadFromFile(path string) error {
	content, err := os.ReadFile(path)
	if err != nil {
		return err
	}

	loaded := Assemblies{}
	if err := json.Unmarshal(content, &loaded); err != nil {
		return err
	}
	if err := loaded.RebuildIndexes(); err != nil {
		return err
	}

	*assemblies = loaded
	return nil
}

// РЎРѕС…СЂР°РЅСЏРµС‚ РІР°СЂРёР°РЅС‚С‹ РѕСЃРЅР°С‰РµРЅРёСЏ РІ JSON-С„Р°Р№Р».
func (assemblies *Assemblies) SaveToFile(path string) error {
	assemblies.ensureMaps()
	return saveTableWithOrderedItems(path, assemblies.MaxID, assemblies.Items)
}

// РџРѕРґРіРѕС‚Р°РІР»РёРІР°РµС‚ РѕСЃРЅРѕРІРЅРѕРµ С…СЂР°РЅРёР»РёС‰Рµ Рё РёРЅРґРµРєСЃС‹.
func (assemblies *Assemblies) ensureMaps() {
	assemblies.ensureItems()
	if assemblies.ByCosmicObjectModelID == nil {
		assemblies.ByCosmicObjectModelID = make(map[int64][]*Assembly)
	}
}

// РџРѕРґРіРѕС‚Р°РІР»РёРІР°РµС‚ РѕСЃРЅРѕРІРЅСѓСЋ map.
func (assemblies *Assemblies) ensureItems() {
	if assemblies.Items == nil {
		assemblies.Items = make(map[int64]*Assembly)
	}
}

// РџСЂРѕРІРµСЂСЏРµС‚ РѕР±СЏР·Р°С‚РµР»СЊРЅС‹Рµ РїРѕР»СЏ РІР°СЂРёР°РЅС‚Р° РѕСЃРЅР°С‰РµРЅРёСЏ.
func (assemblies *Assemblies) validateRequiredFields(assembly *Assembly) error {
	if assembly.CosmicObjectModelID <= 0 {
		return errors.New("cosmic object model ID is empty")
	}
	return nil
}

// Р”РѕР±Р°РІР»СЏРµС‚ РІР°СЂРёР°РЅС‚ РѕСЃРЅР°С‰РµРЅРёСЏ РІ Р±С‹СЃС‚СЂС‹Рµ РёРЅРґРµРєСЃС‹.
func (assemblies *Assemblies) addIndexes(assembly *Assembly) {
	assemblies.ByCosmicObjectModelID[assembly.CosmicObjectModelID] = append(assemblies.ByCosmicObjectModelID[assembly.CosmicObjectModelID], assembly)
}

// РҐСЂР°РЅРёС‚ РєРѕР»РёС‡РµСЃС‚РІРѕ РѕР±РѕСЂСѓРґРѕРІР°РЅРёСЏ РѕРґРЅРѕР№ РјРѕРґРµР»Рё РЅР° РєРѕРЅРєСЂРµС‚РЅРѕР№ СЃР±РѕСЂРєРµ.
type AssemblyEquipmentGroup struct {
	ID                   int64  `json:"ID"`                   // РЈРЅРёРєР°Р»СЊРЅС‹Р№ С‡РёСЃР»РѕРІРѕР№ РёРґРµРЅС‚РёС„РёРєР°С‚РѕСЂ Р·Р°РїРёСЃРё.
	AssemblyID           int64  `json:"AssemblyID"`           // РЎР±РѕСЂРєР°, РЅР° РєРѕС‚РѕСЂРѕР№ СѓСЃС‚Р°РЅРѕРІР»РµРЅР° РіСЂСѓРїРїР° РѕР±РѕСЂСѓРґРѕРІР°РЅРёСЏ.
	Title                string `json:"Title"`                // РќР°Р·РІР°РЅРёРµ РіСЂСѓРїРїС‹ СѓСЃС‚Р°РЅРѕРІР»РµРЅРЅРѕРіРѕ РѕР±РѕСЂСѓРґРѕРІР°РЅРёСЏ.
	EquipmentItemModelID int64  `json:"EquipmentItemModelID"` // РњРѕРґРµР»СЊ СѓСЃС‚Р°РЅРѕРІР»РµРЅРЅРѕРіРѕ РѕР±РѕСЂСѓРґРѕРІР°РЅРёСЏ.
	Count                int64  `json:"Count"`                // РљРѕР»РёС‡РµСЃС‚РІРѕ СѓСЃС‚Р°РЅРѕРІР»РµРЅРЅС‹С… РµРґРёРЅРёС† РѕР±РѕСЂСѓРґРѕРІР°РЅРёСЏ.
}

// РҐСЂР°РЅРёС‚ РіСЂСѓРїРїС‹ РѕР±РѕСЂСѓРґРѕРІР°РЅРёСЏ СЃР±РѕСЂРѕРє.
type AssemblyEquipmentGroups struct {
	MaxID int64                             `json:"MaxID"` // РџРѕСЃР»РµРґРЅРёР№ РІС‹РґР°РЅРЅС‹Р№ С‡РёСЃР»РѕРІРѕР№ РёРґРµРЅС‚РёС„РёРєР°С‚РѕСЂ Р·Р°РїРёСЃРµР№.
	Items map[int64]*AssemblyEquipmentGroup `json:"Items"` // РћСЃРЅРѕРІРЅРѕРµ С…СЂР°РЅРёР»РёС‰Рµ Р·Р°РїРёСЃРµР№ РїРѕ С‡РёСЃР»РѕРІРѕРјСѓ РёРґРµРЅС‚РёС„РёРєР°С‚РѕСЂСѓ.

	ByAssemblyID map[int64][]*AssemblyEquipmentGroup `json:"-"` // Р‘С‹СЃС‚СЂС‹Р№ РїРѕРёСЃРє РіСЂСѓРїРї РѕР±РѕСЂСѓРґРѕРІР°РЅРёСЏ РїРѕ СЃР±РѕСЂРєРµ.
}

// РЎРѕР·РґР°РµС‚ РїСѓСЃС‚РѕРµ С…СЂР°РЅРёР»РёС‰Рµ РіСЂСѓРїРї РѕР±РѕСЂСѓРґРѕРІР°РЅРёСЏ.
func NewAssemblyEquipmentGroups() *AssemblyEquipmentGroups {
	return &AssemblyEquipmentGroups{
		Items:        make(map[int64]*AssemblyEquipmentGroup),
		ByAssemblyID: make(map[int64][]*AssemblyEquipmentGroup),
	}
}

// Р’РѕР·РІСЂР°С‰Р°РµС‚ РіСЂСѓРїРїС‹ РѕР±РѕСЂСѓРґРѕРІР°РЅРёСЏ СѓРєР°Р·Р°РЅРЅРѕР№ СЃР±РѕСЂРєРё.
func (groups *AssemblyEquipmentGroups) GetByAssemblyID(assemblyID int64) []*AssemblyEquipmentGroup {
	if groups.ByAssemblyID == nil {
		_ = groups.RebuildIndexes()
	}
	return groups.ByAssemblyID[assemblyID]
}

// РџРµСЂРµСЃРѕР±РёСЂР°РµС‚ С…СЂР°РЅРёР»РёС‰Рµ РїРѕСЃР»Рµ Р·Р°РіСЂСѓР·РєРё РёР· JSON.
func (groups *AssemblyEquipmentGroups) RebuildIndexes() error {
	if groups.Items == nil {
		groups.Items = make(map[int64]*AssemblyEquipmentGroup)
	}
	groups.ByAssemblyID = make(map[int64][]*AssemblyEquipmentGroup)

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
			return fmt.Errorf("assembly equipment group with ID %d is nil", id)
		}
		if group.ID != id {
			return fmt.Errorf("assembly equipment group map key %d does not match group ID %d", id, group.ID)
		}
		if group.AssemblyID <= 0 {
			return fmt.Errorf("assembly equipment group with ID %d has empty assembly ID", id)
		}
		if id > maxID {
			maxID = id
		}
		groups.ByAssemblyID[group.AssemblyID] = append(groups.ByAssemblyID[group.AssemblyID], group)
	}
	if groups.MaxID < maxID {
		groups.MaxID = maxID
	}
	return nil
}

// Р—Р°РіСЂСѓР¶Р°РµС‚ РіСЂСѓРїРїС‹ РѕР±РѕСЂСѓРґРѕРІР°РЅРёСЏ РёР· JSON-С„Р°Р№Р»Р°.
func (groups *AssemblyEquipmentGroups) LoadFromFile(path string) error {
	content, err := os.ReadFile(path)
	if err != nil {
		return err
	}

	loaded := AssemblyEquipmentGroups{}
	if err := json.Unmarshal(content, &loaded); err != nil {
		return err
	}
	if err := loaded.RebuildIndexes(); err != nil {
		return err
	}

	*groups = loaded
	return nil
}

// РЎРѕС…СЂР°РЅСЏРµС‚ РіСЂСѓРїРїС‹ РѕР±РѕСЂСѓРґРѕРІР°РЅРёСЏ РІ JSON-С„Р°Р№Р».
func (groups *AssemblyEquipmentGroups) SaveToFile(path string) error {
	if groups.Items == nil {
		groups.Items = make(map[int64]*AssemblyEquipmentGroup)
	}
	return saveTableWithOrderedItems(path, groups.MaxID, groups.Items)
}
