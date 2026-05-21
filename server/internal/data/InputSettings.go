package data

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
)

// ActionType С…СЂР°РЅРёС‚ РёРіСЂРѕРІРѕРµ РґРµР№СЃС‚РІРёРµ, Рє РєРѕС‚РѕСЂРѕРјСѓ РјРѕР¶РЅРѕ РїСЂРёРІСЏР·Р°С‚СЊ СЃРѕР±С‹С‚РёРµ РІРІРѕРґР°.
type ActionType struct {
	ID          int64  `json:"ID"`          // РЈРЅРёРєР°Р»СЊРЅС‹Р№ С‡РёСЃР»РѕРІРѕР№ РёРґРµРЅС‚РёС„РёРєР°С‚РѕСЂ Р·Р°РїРёСЃРё.
	TitleRu     string `json:"TitleRu"`     // Р СѓСЃСЃРєРѕРµ РЅР°Р·РІР°РЅРёРµ РґР»СЏ РёРЅС‚РµСЂС„РµР№СЃР° РЅР°СЃС‚СЂРѕРµРє.
	TitleEn     string `json:"TitleEn"`     // РђРЅРіР»РёР№СЃРєРѕРµ РЅР°Р·РІР°РЅРёРµ РґР»СЏ РёРЅС‚РµСЂС„РµР№СЃР° Рё РґР°РЅРЅС‹С….
	Acronym     string `json:"Acronym"`     // РќРµРёР·РјРµРЅСЏРµРјС‹Р№ СЃС‚СЂРѕРєРѕРІС‹Р№ РёРґРµРЅС‚РёС„РёРєР°С‚РѕСЂ РґРµР№СЃС‚РІРёСЏ.
	Description string `json:"Description"` // РџРѕСЏСЃРЅРµРЅРёРµ РЅР°Р·РЅР°С‡РµРЅРёСЏ РґРµР№СЃС‚РІРёСЏ.
}

// ActionTypes С…СЂР°РЅРёС‚ РёРіСЂРѕРІС‹Рµ РґРµР№СЃС‚РІРёСЏ Рё Р±С‹СЃС‚СЂС‹Р№ РїРѕРёСЃРє РїРѕ Р°РєСЂРѕРЅРёРјСѓ.
type ActionTypes struct {
	MaxID int64                 `json:"MaxID"` // РџРѕСЃР»РµРґРЅРёР№ РІС‹РґР°РЅРЅС‹Р№ С‡РёСЃР»РѕРІРѕР№ РёРґРµРЅС‚РёС„РёРєР°С‚РѕСЂ Р·Р°РїРёСЃРµР№.
	Items map[int64]*ActionType `json:"Items"` // РћСЃРЅРѕРІРЅРѕРµ С…СЂР°РЅРёР»РёС‰Рµ Р·Р°РїРёСЃРµР№ РїРѕ С‡РёСЃР»РѕРІРѕРјСѓ РёРґРµРЅС‚РёС„РёРєР°С‚РѕСЂСѓ.

	ByAcronym map[string]*ActionType `json:"-"` // Р‘С‹СЃС‚СЂС‹Р№ РїРѕРёСЃРє Р·Р°РїРёСЃРё РїРѕ Р°РєСЂРѕРЅРёРјСѓ.
}

// InputEventType С…СЂР°РЅРёС‚ СЃРёСЃС‚РµРјРЅРѕРµ СЃРѕР±С‹С‚РёРµ, РґРѕСЃС‚СѓРїРЅРѕРµ РґР»СЏ РїСЂРёРІСЏР·РєРё Рє РґРµР№СЃС‚РІРёСЋ.
type InputEventType struct {
	ID                 int64  `json:"ID"`                 // РЈРЅРёРєР°Р»СЊРЅС‹Р№ С‡РёСЃР»РѕРІРѕР№ РёРґРµРЅС‚РёС„РёРєР°С‚РѕСЂ Р·Р°РїРёСЃРё.
	TitleRu            string `json:"TitleRu"`            // Р СѓСЃСЃРєРѕРµ РЅР°Р·РІР°РЅРёРµ РґР»СЏ РёРЅС‚РµСЂС„РµР№СЃР° РЅР°СЃС‚СЂРѕРµРє.
	TitleEn            string `json:"TitleEn"`            // РђРЅРіР»РёР№СЃРєРѕРµ РЅР°Р·РІР°РЅРёРµ РґР»СЏ РёРЅС‚РµСЂС„РµР№СЃР° Рё РґР°РЅРЅС‹С….
	Acronym            string `json:"Acronym"`            // РќРµРёР·РјРµРЅСЏРµРјС‹Р№ СЃС‚СЂРѕРєРѕРІС‹Р№ РёРґРµРЅС‚РёС„РёРєР°С‚РѕСЂ СЃРѕР±С‹С‚РёСЏ.
	SystemStringValue  string `json:"SystemStringValue"`  // РЎРёСЃС‚РµРјРЅРѕРµ СЃС‚СЂРѕРєРѕРІРѕРµ Р·РЅР°С‡РµРЅРёРµ Р±СЂР°СѓР·РµСЂРЅРѕРіРѕ СЃРѕР±С‹С‚РёСЏ.
	SystemIntegerValue int64  `json:"SystemIntegerValue"` // РЎРёСЃС‚РµРјРЅРѕРµ С‡РёСЃР»РѕРІРѕРµ Р·РЅР°С‡РµРЅРёРµ, РµСЃР»Рё РѕРЅРѕ РµСЃС‚СЊ.
	Description        string `json:"Description"`        // РџРѕСЏСЃРЅРµРЅРёРµ СЃРѕР±С‹С‚РёСЏ РІРІРѕРґР°.
}

// InputEventTypes С…СЂР°РЅРёС‚ РґРѕСЃС‚СѓРїРЅС‹Рµ СЃРѕР±С‹С‚РёСЏ РІРІРѕРґР° Рё Р±С‹СЃС‚СЂС‹Р№ РїРѕРёСЃРє РїРѕ СЃРёСЃС‚РµРјРЅРѕРјСѓ Р·РЅР°С‡РµРЅРёСЋ.
type InputEventTypes struct {
	MaxID int64                     `json:"MaxID"` // РџРѕСЃР»РµРґРЅРёР№ РІС‹РґР°РЅРЅС‹Р№ С‡РёСЃР»РѕРІРѕР№ РёРґРµРЅС‚РёС„РёРєР°С‚РѕСЂ Р·Р°РїРёСЃРµР№.
	Items map[int64]*InputEventType `json:"Items"` // РћСЃРЅРѕРІРЅРѕРµ С…СЂР°РЅРёР»РёС‰Рµ Р·Р°РїРёСЃРµР№ РїРѕ С‡РёСЃР»РѕРІРѕРјСѓ РёРґРµРЅС‚РёС„РёРєР°С‚РѕСЂСѓ.

	ByAcronym           map[string]*InputEventType `json:"-"` // Р‘С‹СЃС‚СЂС‹Р№ РїРѕРёСЃРє Р·Р°РїРёСЃРё РїРѕ Р°РєСЂРѕРЅРёРјСѓ.
	BySystemStringValue map[string]*InputEventType `json:"-"` // Р‘С‹СЃС‚СЂС‹Р№ РїРѕРёСЃРє Р·Р°РїРёСЃРё РїРѕ СЃРёСЃС‚РµРјРЅРѕР№ СЃС‚СЂРѕРєРµ.
}

// DefaultActionInputSetting С…СЂР°РЅРёС‚ РёСЃС…РѕРґРЅСѓСЋ РїСЂРёРІСЏР·РєСѓ РґРµР№СЃС‚РІРёСЏ Рє СЃРѕР±С‹С‚РёСЋ.
type DefaultActionInputSetting struct {
	ID               int64 `json:"ID"`               // РЈРЅРёРєР°Р»СЊРЅС‹Р№ С‡РёСЃР»РѕРІРѕР№ РёРґРµРЅС‚РёС„РёРєР°С‚РѕСЂ Р·Р°РїРёСЃРё.
	ActionTypeID     int64 `json:"ActionTypeID"`     // Р”РµР№СЃС‚РІРёРµ, РІС‹РїРѕР»РЅСЏРµРјРѕРµ РїРѕ СѓРјРѕР»С‡Р°РЅРёСЋ.
	InputEventTypeID int64 `json:"InputEventTypeID"` // РЎРѕР±С‹С‚РёРµ РІРІРѕРґР° РґР»СЏ РґРµР№СЃС‚РІРёСЏ РїРѕ СѓРјРѕР»С‡Р°РЅРёСЋ.
}

// DefaultActionInputSettings С…СЂР°РЅРёС‚ РёСЃС…РѕРґРЅС‹Рµ РїСЂРёРІСЏР·РєРё Рё РёРЅРґРµРєСЃС‹ СѓРЅРёРєР°Р»СЊРЅРѕСЃС‚Рё.
type DefaultActionInputSettings struct {
	MaxID int64                                `json:"MaxID"` // РџРѕСЃР»РµРґРЅРёР№ РІС‹РґР°РЅРЅС‹Р№ С‡РёСЃР»РѕРІРѕР№ РёРґРµРЅС‚РёС„РёРєР°С‚РѕСЂ Р·Р°РїРёСЃРµР№.
	Items map[int64]*DefaultActionInputSetting `json:"Items"` // РћСЃРЅРѕРІРЅРѕРµ С…СЂР°РЅРёР»РёС‰Рµ Р·Р°РїРёСЃРµР№ РїРѕ С‡РёСЃР»РѕРІРѕРјСѓ РёРґРµРЅС‚РёС„РёРєР°С‚РѕСЂСѓ.

	ByActionTypeID     map[int64]*DefaultActionInputSetting `json:"-"` // Р‘С‹СЃС‚СЂС‹Р№ РїРѕРёСЃРє РЅР°СЃС‚СЂРѕР№РєРё РїРѕ РґРµР№СЃС‚РІРёСЋ.
	ByInputEventTypeID map[int64]*DefaultActionInputSetting `json:"-"` // Р‘С‹СЃС‚СЂС‹Р№ РїРѕРёСЃРє РЅР°СЃС‚СЂРѕР№РєРё РїРѕ СЃРѕР±С‹С‚РёСЋ.
}

// AccountActionInputSetting С…СЂР°РЅРёС‚ РїРµСЂРµРѕРїСЂРµРґРµР»РµРЅРёРµ РїСЂРёРІСЏР·РєРё РґР»СЏ РєРѕРЅРєСЂРµС‚РЅРѕРіРѕ Р°РєРєР°СѓРЅС‚Р°.
type AccountActionInputSetting struct {
	ID               int64 `json:"ID"`               // РЈРЅРёРєР°Р»СЊРЅС‹Р№ С‡РёСЃР»РѕРІРѕР№ РёРґРµРЅС‚РёС„РёРєР°С‚РѕСЂ Р·Р°РїРёСЃРё.
	AccountID        int64 `json:"AccountID"`        // РђРєРєР°СѓРЅС‚, РєРѕС‚РѕСЂРѕРјСѓ РїСЂРёРЅР°РґР»РµР¶РёС‚ РїРµСЂРµРѕРїСЂРµРґРµР»РµРЅРёРµ.
	ActionTypeID     int64 `json:"ActionTypeID"`     // Р”РµР№СЃС‚РІРёРµ, РїРµСЂРµРѕРїСЂРµРґРµР»РµРЅРЅРѕРµ РёРіСЂРѕРєРѕРј.
	InputEventTypeID int64 `json:"InputEventTypeID"` // РЎРѕР±С‹С‚РёРµ РІРІРѕРґР°, РІС‹Р±СЂР°РЅРЅРѕРµ РёРіСЂРѕРєРѕРј.
}

// AccountActionInputSettings С…СЂР°РЅРёС‚ Р°РєРєР°СѓРЅС‚РЅС‹Рµ РїСЂРёРІСЏР·РєРё Рё РёРЅРґРµРєСЃС‹ СѓРЅРёРєР°Р»СЊРЅРѕСЃС‚Рё.
type AccountActionInputSettings struct {
	MaxID int64                                `json:"MaxID"` // РџРѕСЃР»РµРґРЅРёР№ РІС‹РґР°РЅРЅС‹Р№ С‡РёСЃР»РѕРІРѕР№ РёРґРµРЅС‚РёС„РёРєР°С‚РѕСЂ Р·Р°РїРёСЃРµР№.
	Items map[int64]*AccountActionInputSetting `json:"Items"` // РћСЃРЅРѕРІРЅРѕРµ С…СЂР°РЅРёР»РёС‰Рµ Р·Р°РїРёСЃРµР№ РїРѕ С‡РёСЃР»РѕРІРѕРјСѓ РёРґРµРЅС‚РёС„РёРєР°С‚РѕСЂСѓ.

	ByAccountAndAction map[string]*AccountActionInputSetting `json:"-"` // Р‘С‹СЃС‚СЂС‹Р№ РїРѕРёСЃРє РЅР°СЃС‚СЂРѕР№РєРё РїРѕ Р°РєРєР°СѓРЅС‚Сѓ Рё РґРµР№СЃС‚РІРёСЋ.
	ByAccountAndEvent  map[string]*AccountActionInputSetting `json:"-"` // Р‘С‹СЃС‚СЂС‹Р№ РїРѕРёСЃРє РЅР°СЃС‚СЂРѕР№РєРё РїРѕ Р°РєРєР°СѓРЅС‚Сѓ Рё СЃРѕР±С‹С‚РёСЋ.
}

func NewActionTypes() *ActionTypes {
	types := &ActionTypes{}
	types.ensureMaps()
	return types
}

func NewInputEventTypes() *InputEventTypes {
	types := &InputEventTypes{}
	types.ensureMaps()
	return types
}

func NewDefaultActionInputSettings() *DefaultActionInputSettings {
	settings := &DefaultActionInputSettings{}
	settings.ensureMaps()
	return settings
}

func NewAccountActionInputSettings() *AccountActionInputSettings {
	settings := &AccountActionInputSettings{}
	settings.ensureMaps()
	return settings
}

func (types *ActionTypes) Get(id int64) (*ActionType, bool) {
	types.ensureMaps()
	item, ok := types.Items[id]
	return item, ok
}

func (types *ActionTypes) RebuildIndexes() error {
	types.ensureItems()
	types.ByAcronym = map[string]*ActionType{}
	var maxID int64
	for _, id := range sortedTableItemIDs(types.Items) {
		item := types.Items[id]
		if item == nil {
			return fmt.Errorf("action type with ID %d is nil", id)
		}
		if item.ID != id || item.TitleRu == "" || item.TitleEn == "" || item.Acronym == "" {
			return fmt.Errorf("action type with ID %d is invalid", id)
		}
		if existing := types.ByAcronym[item.Acronym]; existing != nil && existing.ID != item.ID {
			return errors.New("action type acronym is not unique")
		}
		if id > maxID {
			maxID = id
		}
		types.ByAcronym[item.Acronym] = item
	}
	if types.MaxID < maxID {
		types.MaxID = maxID
	}
	return nil
}

func (types *ActionTypes) LoadFromFile(path string) error {
	content, err := os.ReadFile(path)
	if err != nil {
		return err
	}
	loaded := ActionTypes{}
	if err := json.Unmarshal(content, &loaded); err != nil {
		return err
	}
	if err := loaded.RebuildIndexes(); err != nil {
		return err
	}
	*types = loaded
	return nil
}

func (types *ActionTypes) SaveToFile(path string) error {
	types.ensureMaps()
	return saveTableWithOrderedItems(path, types.MaxID, types.Items)
}

func (types *ActionTypes) ensureMaps() {
	types.ensureItems()
	if types.ByAcronym == nil {
		types.ByAcronym = map[string]*ActionType{}
	}
}

func (types *ActionTypes) ensureItems() {
	if types.Items == nil {
		types.Items = map[int64]*ActionType{}
	}
}

func (types *InputEventTypes) Get(id int64) (*InputEventType, bool) {
	types.ensureMaps()
	item, ok := types.Items[id]
	return item, ok
}

func (types *InputEventTypes) RebuildIndexes() error {
	types.ensureItems()
	types.ByAcronym = map[string]*InputEventType{}
	types.BySystemStringValue = map[string]*InputEventType{}
	var maxID int64
	for _, id := range sortedTableItemIDs(types.Items) {
		item := types.Items[id]
		if item == nil {
			return fmt.Errorf("input event type with ID %d is nil", id)
		}
		if item.ID != id || item.TitleRu == "" || item.TitleEn == "" || item.Acronym == "" || item.SystemStringValue == "" {
			return fmt.Errorf("input event type with ID %d is invalid", id)
		}
		if existing := types.ByAcronym[item.Acronym]; existing != nil && existing.ID != item.ID {
			return errors.New("input event type acronym is not unique")
		}
		if existing := types.BySystemStringValue[item.SystemStringValue]; existing != nil && existing.ID != item.ID {
			return errors.New("input event type system string is not unique")
		}
		if id > maxID {
			maxID = id
		}
		types.ByAcronym[item.Acronym] = item
		types.BySystemStringValue[item.SystemStringValue] = item
	}
	if types.MaxID < maxID {
		types.MaxID = maxID
	}
	return nil
}

func (types *InputEventTypes) LoadFromFile(path string) error {
	content, err := os.ReadFile(path)
	if err != nil {
		return err
	}
	loaded := InputEventTypes{}
	if err := json.Unmarshal(content, &loaded); err != nil {
		return err
	}
	if err := loaded.RebuildIndexes(); err != nil {
		return err
	}
	*types = loaded
	return nil
}

func (types *InputEventTypes) SaveToFile(path string) error {
	types.ensureMaps()
	return saveTableWithOrderedItems(path, types.MaxID, types.Items)
}

func (types *InputEventTypes) ensureMaps() {
	types.ensureItems()
	if types.ByAcronym == nil {
		types.ByAcronym = map[string]*InputEventType{}
	}
	if types.BySystemStringValue == nil {
		types.BySystemStringValue = map[string]*InputEventType{}
	}
}

func (types *InputEventTypes) ensureItems() {
	if types.Items == nil {
		types.Items = map[int64]*InputEventType{}
	}
}

func (settings *DefaultActionInputSettings) RebuildIndexes() error {
	settings.ensureItems()
	settings.ByActionTypeID = map[int64]*DefaultActionInputSetting{}
	settings.ByInputEventTypeID = map[int64]*DefaultActionInputSetting{}
	var maxID int64
	for _, id := range sortedTableItemIDs(settings.Items) {
		item := settings.Items[id]
		if item == nil {
			return fmt.Errorf("default input setting with ID %d is nil", id)
		}
		if item.ID != id || item.ActionTypeID <= 0 || item.InputEventTypeID <= 0 {
			return fmt.Errorf("default input setting with ID %d is invalid", id)
		}
		if existing := settings.ByActionTypeID[item.ActionTypeID]; existing != nil && existing.ID != item.ID {
			return errors.New("default input setting action is not unique")
		}
		if existing := settings.ByInputEventTypeID[item.InputEventTypeID]; existing != nil && existing.ID != item.ID {
			return errors.New("default input setting event is not unique")
		}
		if id > maxID {
			maxID = id
		}
		settings.ByActionTypeID[item.ActionTypeID] = item
		settings.ByInputEventTypeID[item.InputEventTypeID] = item
	}
	if settings.MaxID < maxID {
		settings.MaxID = maxID
	}
	return nil
}

func (settings *DefaultActionInputSettings) LoadFromFile(path string) error {
	content, err := os.ReadFile(path)
	if err != nil {
		return err
	}
	loaded := DefaultActionInputSettings{}
	if err := json.Unmarshal(content, &loaded); err != nil {
		return err
	}
	if err := loaded.RebuildIndexes(); err != nil {
		return err
	}
	*settings = loaded
	return nil
}

func (settings *DefaultActionInputSettings) SaveToFile(path string) error {
	settings.ensureMaps()
	return saveTableWithOrderedItems(path, settings.MaxID, settings.Items)
}

func (settings *DefaultActionInputSettings) ensureMaps() {
	settings.ensureItems()
	if settings.ByActionTypeID == nil {
		settings.ByActionTypeID = map[int64]*DefaultActionInputSetting{}
	}
	if settings.ByInputEventTypeID == nil {
		settings.ByInputEventTypeID = map[int64]*DefaultActionInputSetting{}
	}
}

func (settings *DefaultActionInputSettings) ensureItems() {
	if settings.Items == nil {
		settings.Items = map[int64]*DefaultActionInputSetting{}
	}
}

func (settings *AccountActionInputSettings) GetByAccount(accountID int64) []*AccountActionInputSetting {
	settings.ensureMaps()
	result := make([]*AccountActionInputSetting, 0)
	for _, item := range settings.Items {
		if item.AccountID == accountID {
			result = append(result, item)
		}
	}
	return result
}

func (settings *AccountActionInputSettings) ReplaceForAccount(accountID int64, items []AccountActionInputSetting) error {
	if accountID <= 0 {
		return errors.New("account ID is empty")
	}
	settings.ensureMaps()
	for _, item := range items {
		if item.ActionTypeID <= 0 || item.InputEventTypeID <= 0 {
			return errors.New("account input setting has empty required fields")
		}
	}
	for _, item := range settings.GetByAccount(accountID) {
		delete(settings.Items, item.ID)
	}
	for _, item := range items {
		settings.MaxID++
		created := item
		created.ID = settings.MaxID
		created.AccountID = accountID
		settings.Items[created.ID] = &created
	}
	return settings.RebuildIndexes()
}

func (settings *AccountActionInputSettings) RebuildIndexes() error {
	settings.ensureItems()
	settings.ByAccountAndAction = map[string]*AccountActionInputSetting{}
	settings.ByAccountAndEvent = map[string]*AccountActionInputSetting{}
	var maxID int64
	for _, id := range sortedTableItemIDs(settings.Items) {
		item := settings.Items[id]
		if item == nil {
			return fmt.Errorf("account input setting with ID %d is nil", id)
		}
		if item.ID != id || item.AccountID <= 0 || item.ActionTypeID <= 0 || item.InputEventTypeID <= 0 {
			return fmt.Errorf("account input setting with ID %d is invalid", id)
		}
		actionKey := accountInputSettingKey(item.AccountID, item.ActionTypeID)
		if existing := settings.ByAccountAndAction[actionKey]; existing != nil && existing.ID != item.ID {
			return errors.New("account input setting action is not unique")
		}
		eventKey := accountInputSettingKey(item.AccountID, item.InputEventTypeID)
		if existing := settings.ByAccountAndEvent[eventKey]; existing != nil && existing.ID != item.ID {
			return errors.New("account input setting event is not unique")
		}
		if id > maxID {
			maxID = id
		}
		settings.ByAccountAndAction[actionKey] = item
		settings.ByAccountAndEvent[eventKey] = item
	}
	if settings.MaxID < maxID {
		settings.MaxID = maxID
	}
	return nil
}

func (settings *AccountActionInputSettings) LoadFromFile(path string) error {
	content, err := os.ReadFile(path)
	if err != nil {
		return err
	}
	loaded := AccountActionInputSettings{}
	if err := json.Unmarshal(content, &loaded); err != nil {
		return err
	}
	if err := loaded.RebuildIndexes(); err != nil {
		return err
	}
	*settings = loaded
	return nil
}

func (settings *AccountActionInputSettings) SaveToFile(path string) error {
	settings.ensureMaps()
	return saveTableWithOrderedItems(path, settings.MaxID, settings.Items)
}

func (settings *AccountActionInputSettings) ensureMaps() {
	settings.ensureItems()
	if settings.ByAccountAndAction == nil {
		settings.ByAccountAndAction = map[string]*AccountActionInputSetting{}
	}
	if settings.ByAccountAndEvent == nil {
		settings.ByAccountAndEvent = map[string]*AccountActionInputSetting{}
	}
}

func (settings *AccountActionInputSettings) ensureItems() {
	if settings.Items == nil {
		settings.Items = map[int64]*AccountActionInputSetting{}
	}
}

func accountInputSettingKey(accountID int64, valueID int64) string {
	return fmt.Sprintf("%d:%d", accountID, valueID)
}
