package data

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
)

// Хранит данные одного типа предмета.
type ItemType struct {
	ID                    int64  `json:"ID"`                    // Уникальный числовой идентификатор записи.
	TitleRu               string `json:"TitleRu"`               // Русское название для интерфейса и данных.
	TitleEn               string `json:"TitleEn"`               // Английское название для интерфейса и данных.
	Acronym               string `json:"Acronym"`               // Неизменяемый строковый идентификатор для логики и ссылок.
	IsEquipmentForShip    bool   `json:"IsEquipmentForShip"`    // Разрешает устанавливать предмет этого типа на корабль.
	IsEquipmentForStation bool   `json:"IsEquipmentForStation"` // Разрешает устанавливать предмет этого типа на станцию.
	IsPilotInstrument     bool   `json:"IsPilotInstrument"`     // Разрешает назначать предмет этого типа в панель пилота.
	IsInternalUsable      bool   `json:"IsInternalUsable"`      // Разрешает внутреннее использование предмета этого типа из панели управления оборудованием.
	CountMustBeInteger    bool   `json:"CountMustBeInteger"`    // Требует хранить количество только целыми единицами.
}

// Хранит типы предметов и быстрые индексы по уникальным полям.
type ItemTypes struct {
	MaxID int64               `json:"MaxID"` // Последний выданный идентификатор для новых записей.
	Items map[int64]*ItemType `json:"Items"` // Основное хранилище записей по числовому идентификатору.

	ByTitleRu map[string]*ItemType `json:"-"` // Быстрый поиск записи по русскому названию.
	ByTitleEn map[string]*ItemType `json:"-"` // Быстрый поиск записи по английскому названию.
	ByAcronym map[string]*ItemType `json:"-"` // Быстрый поиск записи по акрониму.
}

// Создаёт пустое хранилище типов предметов с подготовленными индексами.
func NewItemTypes() *ItemTypes {
	itemTypes := &ItemTypes{}
	itemTypes.ensureMaps()
	return itemTypes
}

// Добавляет новый тип предмета и назначает новый ID.
func (itemTypes *ItemTypes) Add(itemType *ItemType) (*ItemType, error) {
	if itemType == nil {
		return nil, errors.New("itemType is nil")
	}
	itemTypes.ensureMaps()
	if err := itemTypes.validateRequiredFields(itemType); err != nil {
		return nil, err
	}
	if err := itemTypes.ensureUniqueForNewType(itemType); err != nil {
		return nil, err
	}

	itemTypes.MaxID++
	itemType.ID = itemTypes.MaxID
	itemTypes.Items[itemType.ID] = itemType
	itemTypes.addIndexes(itemType)
	return itemType, nil
}

// Возвращает тип предмета по ID.
func (itemTypes *ItemTypes) Get(id int64) (*ItemType, bool) {
	itemTypes.ensureMaps()
	itemType, ok := itemTypes.Items[id]
	return itemType, ok
}

// Удаляет тип предмета и все его быстрые индексы.
func (itemTypes *ItemTypes) Delete(id int64) bool {
	itemTypes.ensureMaps()
	itemType, ok := itemTypes.Items[id]
	if !ok {
		return false
	}

	itemTypes.deleteIndexes(itemType)
	delete(itemTypes.Items, id)
	return true
}

// Возвращает тип предмета по уникальному русскому названию.
func (itemTypes *ItemTypes) GetByTitleRu(titleRu string) (*ItemType, bool) {
	itemTypes.ensureMaps()
	itemType, ok := itemTypes.ByTitleRu[titleRu]
	return itemType, ok
}

// Возвращает тип предмета по уникальному английскому названию.
func (itemTypes *ItemTypes) GetByTitleEn(titleEn string) (*ItemType, bool) {
	itemTypes.ensureMaps()
	itemType, ok := itemTypes.ByTitleEn[titleEn]
	return itemType, ok
}

// Возвращает тип предмета по уникальному акрониму.
func (itemTypes *ItemTypes) GetByAcronym(acronym string) (*ItemType, bool) {
	itemTypes.ensureMaps()
	itemType, ok := itemTypes.ByAcronym[acronym]
	return itemType, ok
}

// Пересобирает быстрые индексы после загрузки из JSON.
func (itemTypes *ItemTypes) RebuildIndexes() error {
	itemTypes.ensureItems()
	itemTypes.ByTitleRu = make(map[string]*ItemType)
	itemTypes.ByTitleEn = make(map[string]*ItemType)
	itemTypes.ByAcronym = make(map[string]*ItemType)

	var maxID int64
	for id, itemType := range itemTypes.Items {
		if itemType == nil {
			return fmt.Errorf("itemType with ID %d is nil", id)
		}
		if itemType.ID != id {
			return fmt.Errorf("itemType map key %d does not match type ID %d", id, itemType.ID)
		}
		if err := itemTypes.validateRequiredFields(itemType); err != nil {
			return fmt.Errorf("itemType with ID %d is invalid: %w", id, err)
		}
		if id > maxID {
			maxID = id
		}
		if err := itemTypes.ensureUniqueForNewType(itemType); err != nil {
			return err
		}
		itemTypes.addIndexes(itemType)
	}
	if itemTypes.MaxID < maxID {
		itemTypes.MaxID = maxID
	}
	return nil
}

// Загружает типы предметов из JSON-файла и пересобирает быстрые индексы.
func (itemTypes *ItemTypes) LoadFromFile(path string) error {
	content, err := os.ReadFile(path)
	if err != nil {
		return err
	}

	loaded := ItemTypes{}
	if err := json.Unmarshal(content, &loaded); err != nil {
		return err
	}
	if err := loaded.RebuildIndexes(); err != nil {
		return err
	}

	*itemTypes = loaded
	return nil
}

// Сохраняет типы предметов в JSON-файл без вспомогательных индексов.
func (itemTypes *ItemTypes) SaveToFile(path string) error {
	itemTypes.ensureMaps()
	return saveTableWithOrderedItems(path, itemTypes.MaxID, itemTypes.Items)
}

// Подготавливает основное хранилище и все индексы.
func (itemTypes *ItemTypes) ensureMaps() {
	itemTypes.ensureItems()
	if itemTypes.ByTitleRu == nil {
		itemTypes.ByTitleRu = make(map[string]*ItemType)
	}
	if itemTypes.ByTitleEn == nil {
		itemTypes.ByTitleEn = make(map[string]*ItemType)
	}
	if itemTypes.ByAcronym == nil {
		itemTypes.ByAcronym = make(map[string]*ItemType)
	}
}

// Подготавливает основную map типов предметов.
func (itemTypes *ItemTypes) ensureItems() {
	if itemTypes.Items == nil {
		itemTypes.Items = make(map[int64]*ItemType)
	}
}

// Проверяет обязательные поля типа предмета.
func (itemTypes *ItemTypes) validateRequiredFields(itemType *ItemType) error {
	if itemType.TitleRu == "" {
		return errors.New("title ru is empty")
	}
	if itemType.TitleEn == "" {
		return errors.New("title en is empty")
	}
	if itemType.Acronym == "" {
		return errors.New("acronym is empty")
	}
	return nil
}

// Проверяет уникальные поля перед добавлением в индексы.
func (itemTypes *ItemTypes) ensureUniqueForNewType(itemType *ItemType) error {
	if existing, ok := itemTypes.ByTitleRu[itemType.TitleRu]; ok && existing.ID != itemType.ID {
		return fmt.Errorf("title ru %q already exists", itemType.TitleRu)
	}
	if existing, ok := itemTypes.ByTitleEn[itemType.TitleEn]; ok && existing.ID != itemType.ID {
		return fmt.Errorf("title en %q already exists", itemType.TitleEn)
	}
	if existing, ok := itemTypes.ByAcronym[itemType.Acronym]; ok && existing.ID != itemType.ID {
		return fmt.Errorf("acronym %q already exists", itemType.Acronym)
	}
	return nil
}

// Добавляет тип предмета во все быстрые индексы.
func (itemTypes *ItemTypes) addIndexes(itemType *ItemType) {
	itemTypes.ByTitleRu[itemType.TitleRu] = itemType
	itemTypes.ByTitleEn[itemType.TitleEn] = itemType
	itemTypes.ByAcronym[itemType.Acronym] = itemType
}

// Удаляет тип предмета из всех быстрых индексов.
func (itemTypes *ItemTypes) deleteIndexes(itemType *ItemType) {
	delete(itemTypes.ByTitleRu, itemType.TitleRu)
	delete(itemTypes.ByTitleEn, itemType.TitleEn)
	delete(itemTypes.ByAcronym, itemType.Acronym)
}
