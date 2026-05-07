package data

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
)

// ActionType хранит игровое действие, к которому можно привязать событие ввода.
type ActionType struct {
	ID          int64  `json:"ID"`          // Уникальный числовой идентификатор записи.
	TitleRu     string `json:"TitleRu"`     // Русское название для интерфейса настроек.
	TitleEn     string `json:"TitleEn"`     // Английское название для интерфейса и данных.
	Acronym     string `json:"Acronym"`     // Неизменяемый строковый идентификатор действия.
	Description string `json:"Description"` // Пояснение назначения действия.
}

// ActionTypes хранит игровые действия и быстрый поиск по акрониму.
type ActionTypes struct {
	MaxID int64                 `json:"MaxID"` // Последний выданный числовой идентификатор записей.
	Items map[int64]*ActionType `json:"Items"` // Основное хранилище записей по числовому идентификатору.

	ByAcronym map[string]*ActionType `json:"-"` // Быстрый поиск записи по акрониму.
}

// InputEventType хранит системное событие, доступное для привязки к действию.
type InputEventType struct {
	ID                 int64  `json:"ID"`                 // Уникальный числовой идентификатор записи.
	TitleRu            string `json:"TitleRu"`            // Русское название для интерфейса настроек.
	TitleEn            string `json:"TitleEn"`            // Английское название для интерфейса и данных.
	Acronym            string `json:"Acronym"`            // Неизменяемый строковый идентификатор события.
	SystemStringValue  string `json:"SystemStringValue"`  // Системное строковое значение браузерного события.
	SystemIntegerValue int64  `json:"SystemIntegerValue"` // Системное числовое значение, если оно есть.
	Description        string `json:"Description"`        // Пояснение события ввода.
}

// InputEventTypes хранит доступные события ввода и быстрый поиск по системному значению.
type InputEventTypes struct {
	MaxID int64                     `json:"MaxID"` // Последний выданный числовой идентификатор записей.
	Items map[int64]*InputEventType `json:"Items"` // Основное хранилище записей по числовому идентификатору.

	ByAcronym           map[string]*InputEventType `json:"-"` // Быстрый поиск записи по акрониму.
	BySystemStringValue map[string]*InputEventType `json:"-"` // Быстрый поиск записи по системной строке.
}

// DefaultActionInputSetting хранит исходную привязку действия к событию.
type DefaultActionInputSetting struct {
	ID               int64 `json:"ID"`               // Уникальный числовой идентификатор записи.
	ActionTypeID     int64 `json:"ActionTypeID"`     // Действие, выполняемое по умолчанию.
	InputEventTypeID int64 `json:"InputEventTypeID"` // Событие ввода для действия по умолчанию.
}

// DefaultActionInputSettings хранит исходные привязки и индексы уникальности.
type DefaultActionInputSettings struct {
	MaxID int64                                `json:"MaxID"` // Последний выданный числовой идентификатор записей.
	Items map[int64]*DefaultActionInputSetting `json:"Items"` // Основное хранилище записей по числовому идентификатору.

	ByActionTypeID     map[int64]*DefaultActionInputSetting `json:"-"` // Быстрый поиск настройки по действию.
	ByInputEventTypeID map[int64]*DefaultActionInputSetting `json:"-"` // Быстрый поиск настройки по событию.
}

// AccountActionInputSetting хранит переопределение привязки для конкретного аккаунта.
type AccountActionInputSetting struct {
	ID               int64 `json:"ID"`               // Уникальный числовой идентификатор записи.
	AccountID        int64 `json:"AccountID"`        // Аккаунт, которому принадлежит переопределение.
	ActionTypeID     int64 `json:"ActionTypeID"`     // Действие, переопределенное игроком.
	InputEventTypeID int64 `json:"InputEventTypeID"` // Событие ввода, выбранное игроком.
}

// AccountActionInputSettings хранит аккаунтные привязки и индексы уникальности.
type AccountActionInputSettings struct {
	MaxID int64                                `json:"MaxID"` // Последний выданный числовой идентификатор записей.
	Items map[int64]*AccountActionInputSetting `json:"Items"` // Основное хранилище записей по числовому идентификатору.

	ByAccountAndAction map[string]*AccountActionInputSetting `json:"-"` // Быстрый поиск настройки по аккаунту и действию.
	ByAccountAndEvent  map[string]*AccountActionInputSetting `json:"-"` // Быстрый поиск настройки по аккаунту и событию.
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
