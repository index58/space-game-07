package data

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
)

// Хранит данные одного типа предмета.
type Itemtype struct {
	ID                    int64  `json:"ID"`
	TitleRu               string `json:"TitleRu"`
	TitleEn               string `json:"TitleEn"`
	Acronym               string `json:"Acronym"`
	IsEquipmentForShip    bool   `json:"IsEquipmentForShip"`
	IsEquipmentForStation bool   `json:"IsEquipmentForStation"`
	IsPilotInstrument     bool   `json:"IsPilotInstrument"`
	CountMustBeInteger    bool   `json:"CountMustBeInteger"`
}

// Хранит типы предметов и быстрые индексы по уникальным полям.
type Itemtypes struct {
	MaxID int64               `json:"MaxID"`
	Items map[int64]*Itemtype `json:"Items"`

	ByTitleRu map[string]*Itemtype `json:"-"`
	ByTitleEn map[string]*Itemtype `json:"-"`
	ByAcronym map[string]*Itemtype `json:"-"`
}

// Создаёт пустое хранилище типов предметов с подготовленными индексами.
func NewItemtypes() *Itemtypes {
	itemtypes := &Itemtypes{}
	itemtypes.ensureMaps()
	return itemtypes
}

// Добавляет новый тип предмета и назначает новый ID.
func (itemtypes *Itemtypes) Add(itemtype *Itemtype) (*Itemtype, error) {
	if itemtype == nil {
		return nil, errors.New("itemtype is nil")
	}
	itemtypes.ensureMaps()
	if err := itemtypes.validateRequiredFields(itemtype); err != nil {
		return nil, err
	}
	if err := itemtypes.ensureUniqueForNewType(itemtype); err != nil {
		return nil, err
	}

	itemtypes.MaxID++
	itemtype.ID = itemtypes.MaxID
	itemtypes.Items[itemtype.ID] = itemtype
	itemtypes.addIndexes(itemtype)
	return itemtype, nil
}

// Возвращает тип предмета по ID.
func (itemtypes *Itemtypes) Get(id int64) (*Itemtype, bool) {
	itemtypes.ensureMaps()
	itemtype, ok := itemtypes.Items[id]
	return itemtype, ok
}

// Удаляет тип предмета и все его быстрые индексы.
func (itemtypes *Itemtypes) Delete(id int64) bool {
	itemtypes.ensureMaps()
	itemtype, ok := itemtypes.Items[id]
	if !ok {
		return false
	}

	itemtypes.deleteIndexes(itemtype)
	delete(itemtypes.Items, id)
	return true
}

// Возвращает тип предмета по уникальному русскому названию.
func (itemtypes *Itemtypes) GetByTitleRu(titleRu string) (*Itemtype, bool) {
	itemtypes.ensureMaps()
	itemtype, ok := itemtypes.ByTitleRu[titleRu]
	return itemtype, ok
}

// Возвращает тип предмета по уникальному английскому названию.
func (itemtypes *Itemtypes) GetByTitleEn(titleEn string) (*Itemtype, bool) {
	itemtypes.ensureMaps()
	itemtype, ok := itemtypes.ByTitleEn[titleEn]
	return itemtype, ok
}

// Возвращает тип предмета по уникальному акрониму.
func (itemtypes *Itemtypes) GetByAcronym(acronym string) (*Itemtype, bool) {
	itemtypes.ensureMaps()
	itemtype, ok := itemtypes.ByAcronym[acronym]
	return itemtype, ok
}

// Пересобирает быстрые индексы после загрузки из JSON.
func (itemtypes *Itemtypes) RebuildIndexes() error {
	itemtypes.ensureItems()
	itemtypes.ByTitleRu = make(map[string]*Itemtype)
	itemtypes.ByTitleEn = make(map[string]*Itemtype)
	itemtypes.ByAcronym = make(map[string]*Itemtype)

	var maxID int64
	for id, itemtype := range itemtypes.Items {
		if itemtype == nil {
			return fmt.Errorf("itemtype with ID %d is nil", id)
		}
		if itemtype.ID != id {
			return fmt.Errorf("itemtype map key %d does not match type ID %d", id, itemtype.ID)
		}
		if err := itemtypes.validateRequiredFields(itemtype); err != nil {
			return fmt.Errorf("itemtype with ID %d is invalid: %w", id, err)
		}
		if id > maxID {
			maxID = id
		}
		if err := itemtypes.ensureUniqueForNewType(itemtype); err != nil {
			return err
		}
		itemtypes.addIndexes(itemtype)
	}
	if itemtypes.MaxID < maxID {
		itemtypes.MaxID = maxID
	}
	return nil
}

// Загружает типы предметов из JSON-файла и пересобирает быстрые индексы.
func (itemtypes *Itemtypes) LoadFromFile(path string) error {
	content, err := os.ReadFile(path)
	if err != nil {
		return err
	}

	loaded := Itemtypes{}
	if err := json.Unmarshal(content, &loaded); err != nil {
		return err
	}
	if err := loaded.RebuildIndexes(); err != nil {
		return err
	}

	*itemtypes = loaded
	return nil
}

// Сохраняет типы предметов в JSON-файл без вспомогательных индексов.
func (itemtypes *Itemtypes) SaveToFile(path string) error {
	itemtypes.ensureMaps()
	content, err := json.MarshalIndent(itemtypes, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(path, content, 0o600)
}

// Подготавливает основное хранилище и все индексы.
func (itemtypes *Itemtypes) ensureMaps() {
	itemtypes.ensureItems()
	if itemtypes.ByTitleRu == nil {
		itemtypes.ByTitleRu = make(map[string]*Itemtype)
	}
	if itemtypes.ByTitleEn == nil {
		itemtypes.ByTitleEn = make(map[string]*Itemtype)
	}
	if itemtypes.ByAcronym == nil {
		itemtypes.ByAcronym = make(map[string]*Itemtype)
	}
}

// Подготавливает основную map типов предметов.
func (itemtypes *Itemtypes) ensureItems() {
	if itemtypes.Items == nil {
		itemtypes.Items = make(map[int64]*Itemtype)
	}
}

// Проверяет обязательные поля типа предмета.
func (itemtypes *Itemtypes) validateRequiredFields(itemtype *Itemtype) error {
	if itemtype.TitleRu == "" {
		return errors.New("title ru is empty")
	}
	if itemtype.TitleEn == "" {
		return errors.New("title en is empty")
	}
	if itemtype.Acronym == "" {
		return errors.New("acronym is empty")
	}
	return nil
}

// Проверяет уникальные поля перед добавлением в индексы.
func (itemtypes *Itemtypes) ensureUniqueForNewType(itemtype *Itemtype) error {
	if existing, ok := itemtypes.ByTitleRu[itemtype.TitleRu]; ok && existing.ID != itemtype.ID {
		return fmt.Errorf("title ru %q already exists", itemtype.TitleRu)
	}
	if existing, ok := itemtypes.ByTitleEn[itemtype.TitleEn]; ok && existing.ID != itemtype.ID {
		return fmt.Errorf("title en %q already exists", itemtype.TitleEn)
	}
	if existing, ok := itemtypes.ByAcronym[itemtype.Acronym]; ok && existing.ID != itemtype.ID {
		return fmt.Errorf("acronym %q already exists", itemtype.Acronym)
	}
	return nil
}

// Добавляет тип предмета во все быстрые индексы.
func (itemtypes *Itemtypes) addIndexes(itemtype *Itemtype) {
	itemtypes.ByTitleRu[itemtype.TitleRu] = itemtype
	itemtypes.ByTitleEn[itemtype.TitleEn] = itemtype
	itemtypes.ByAcronym[itemtype.Acronym] = itemtype
}

// Удаляет тип предмета из всех быстрых индексов.
func (itemtypes *Itemtypes) deleteIndexes(itemtype *Itemtype) {
	delete(itemtypes.ByTitleRu, itemtype.TitleRu)
	delete(itemtypes.ByTitleEn, itemtype.TitleEn)
	delete(itemtypes.ByAcronym, itemtype.Acronym)
}
