package data

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
)

// хранит данные одного типа космического объекта.
type CosmicObjectType struct {
	ID                 int64  `json:"ID"`
	TitleRu            string `json:"TitleRu"`
	TitleEn            string `json:"TitleEn"`
	Acronym            string `json:"Acronym"`
	CharacterLocatable bool   `json:"CharacterLocatable"`
	Movable            bool   `json:"Movable"`
	Rotatable          bool   `json:"Rotatable"`
}

// хранит типы космических объектов и быстрые индексы по уникальным полям.
type CosmicObjectTypes struct {
	MaxID int64                       `json:"MaxID"`
	Items map[int64]*CosmicObjectType `json:"Items"`

	ByTitleRu map[string]*CosmicObjectType `json:"-"`
	ByTitleEn map[string]*CosmicObjectType `json:"-"`
	ByAcronym map[string]*CosmicObjectType `json:"-"`
}

// создаёт пустое хранилище типов космических объектов с подготовленными индексами.
func NewCosmicObjectTypes() *CosmicObjectTypes {
	cosmicObjectTypes := &CosmicObjectTypes{}
	cosmicObjectTypes.ensureMaps()
	return cosmicObjectTypes
}

// добавляет новый тип космического объекта и назначает новый ID.
func (cosmicObjectTypes *CosmicObjectTypes) Add(cosmicObjectType *CosmicObjectType) (*CosmicObjectType, error) {
	if cosmicObjectType == nil {
		return nil, errors.New("cosmic object type is nil")
	}
	cosmicObjectTypes.ensureMaps()
	if err := cosmicObjectTypes.validateRequiredFields(cosmicObjectType); err != nil {
		return nil, err
	}
	if err := cosmicObjectTypes.ensureUniqueForNewType(cosmicObjectType); err != nil {
		return nil, err
	}

	cosmicObjectTypes.MaxID++
	cosmicObjectType.ID = cosmicObjectTypes.MaxID
	cosmicObjectTypes.Items[cosmicObjectType.ID] = cosmicObjectType
	cosmicObjectTypes.addIndexes(cosmicObjectType)
	return cosmicObjectType, nil
}

// возвращает тип космического объекта по ID.
func (cosmicObjectTypes *CosmicObjectTypes) Get(id int64) (*CosmicObjectType, bool) {
	cosmicObjectTypes.ensureMaps()
	cosmicObjectType, ok := cosmicObjectTypes.Items[id]
	return cosmicObjectType, ok
}

// удаляет тип космического объекта и все его быстрые индексы.
func (cosmicObjectTypes *CosmicObjectTypes) Delete(id int64) bool {
	cosmicObjectTypes.ensureMaps()
	cosmicObjectType, ok := cosmicObjectTypes.Items[id]
	if !ok {
		return false
	}

	cosmicObjectTypes.deleteIndexes(cosmicObjectType)
	delete(cosmicObjectTypes.Items, id)
	return true
}

// возвращает тип космического объекта по уникальному русскому названию.
func (cosmicObjectTypes *CosmicObjectTypes) GetByTitleRu(titleRu string) (*CosmicObjectType, bool) {
	cosmicObjectTypes.ensureMaps()
	cosmicObjectType, ok := cosmicObjectTypes.ByTitleRu[titleRu]
	return cosmicObjectType, ok
}

// возвращает тип космического объекта по уникальному английскому названию.
func (cosmicObjectTypes *CosmicObjectTypes) GetByTitleEn(titleEn string) (*CosmicObjectType, bool) {
	cosmicObjectTypes.ensureMaps()
	cosmicObjectType, ok := cosmicObjectTypes.ByTitleEn[titleEn]
	return cosmicObjectType, ok
}

// возвращает тип космического объекта по уникальному акрониму.
func (cosmicObjectTypes *CosmicObjectTypes) GetByAcronym(acronym string) (*CosmicObjectType, bool) {
	cosmicObjectTypes.ensureMaps()
	cosmicObjectType, ok := cosmicObjectTypes.ByAcronym[acronym]
	return cosmicObjectType, ok
}

// пересобирает быстрые индексы после загрузки из JSON.
func (cosmicObjectTypes *CosmicObjectTypes) RebuildIndexes() error {
	cosmicObjectTypes.ensureItems()
	cosmicObjectTypes.ByTitleRu = make(map[string]*CosmicObjectType)
	cosmicObjectTypes.ByTitleEn = make(map[string]*CosmicObjectType)
	cosmicObjectTypes.ByAcronym = make(map[string]*CosmicObjectType)

	var maxID int64
	for id, cosmicObjectType := range cosmicObjectTypes.Items {
		if cosmicObjectType == nil {
			return fmt.Errorf("cosmic object type with ID %d is nil", id)
		}
		if cosmicObjectType.ID != id {
			return fmt.Errorf("cosmic object type map key %d does not match type ID %d", id, cosmicObjectType.ID)
		}
		if err := cosmicObjectTypes.validateRequiredFields(cosmicObjectType); err != nil {
			return fmt.Errorf("cosmic object type with ID %d is invalid: %w", id, err)
		}
		if id > maxID {
			maxID = id
		}
		if err := cosmicObjectTypes.ensureUniqueForNewType(cosmicObjectType); err != nil {
			return err
		}
		cosmicObjectTypes.addIndexes(cosmicObjectType)
	}
	if cosmicObjectTypes.MaxID < maxID {
		cosmicObjectTypes.MaxID = maxID
	}
	return nil
}

// загружает типы космических объектов из JSON-файла и пересобирает быстрые индексы.
func (cosmicObjectTypes *CosmicObjectTypes) LoadFromFile(path string) error {
	content, err := os.ReadFile(path)
	if err != nil {
		return err
	}

	loaded := CosmicObjectTypes{}
	if err := json.Unmarshal(content, &loaded); err != nil {
		return err
	}
	if err := loaded.RebuildIndexes(); err != nil {
		return err
	}

	*cosmicObjectTypes = loaded
	return nil
}

// сохраняет типы космических объектов в JSON-файл без вспомогательных индексов.
func (cosmicObjectTypes *CosmicObjectTypes) SaveToFile(path string) error {
	cosmicObjectTypes.ensureMaps()
	content, err := json.MarshalIndent(cosmicObjectTypes, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(path, content, 0o600)
}

// подготавливает основное хранилище и все индексы.
func (cosmicObjectTypes *CosmicObjectTypes) ensureMaps() {
	cosmicObjectTypes.ensureItems()
	if cosmicObjectTypes.ByTitleRu == nil {
		cosmicObjectTypes.ByTitleRu = make(map[string]*CosmicObjectType)
	}
	if cosmicObjectTypes.ByTitleEn == nil {
		cosmicObjectTypes.ByTitleEn = make(map[string]*CosmicObjectType)
	}
	if cosmicObjectTypes.ByAcronym == nil {
		cosmicObjectTypes.ByAcronym = make(map[string]*CosmicObjectType)
	}
}

// подготавливает основную map типов космических объектов.
func (cosmicObjectTypes *CosmicObjectTypes) ensureItems() {
	if cosmicObjectTypes.Items == nil {
		cosmicObjectTypes.Items = make(map[int64]*CosmicObjectType)
	}
}

// проверяет обязательные поля типа космического объекта.
func (cosmicObjectTypes *CosmicObjectTypes) validateRequiredFields(cosmicObjectType *CosmicObjectType) error {
	if cosmicObjectType.TitleRu == "" {
		return errors.New("title ru is empty")
	}
	if cosmicObjectType.TitleEn == "" {
		return errors.New("title en is empty")
	}
	if cosmicObjectType.Acronym == "" {
		return errors.New("acronym is empty")
	}
	return nil
}

// проверяет уникальные поля перед добавлением в индексы.
func (cosmicObjectTypes *CosmicObjectTypes) ensureUniqueForNewType(cosmicObjectType *CosmicObjectType) error {
	if existing, ok := cosmicObjectTypes.ByTitleRu[cosmicObjectType.TitleRu]; ok && existing.ID != cosmicObjectType.ID {
		return fmt.Errorf("title ru %q already exists", cosmicObjectType.TitleRu)
	}
	if existing, ok := cosmicObjectTypes.ByTitleEn[cosmicObjectType.TitleEn]; ok && existing.ID != cosmicObjectType.ID {
		return fmt.Errorf("title en %q already exists", cosmicObjectType.TitleEn)
	}
	if existing, ok := cosmicObjectTypes.ByAcronym[cosmicObjectType.Acronym]; ok && existing.ID != cosmicObjectType.ID {
		return fmt.Errorf("acronym %q already exists", cosmicObjectType.Acronym)
	}
	return nil
}

// добавляет тип космического объекта во все быстрые индексы.
func (cosmicObjectTypes *CosmicObjectTypes) addIndexes(cosmicObjectType *CosmicObjectType) {
	cosmicObjectTypes.ByTitleRu[cosmicObjectType.TitleRu] = cosmicObjectType
	cosmicObjectTypes.ByTitleEn[cosmicObjectType.TitleEn] = cosmicObjectType
	cosmicObjectTypes.ByAcronym[cosmicObjectType.Acronym] = cosmicObjectType
}

// удаляет тип космического объекта из всех быстрых индексов.
func (cosmicObjectTypes *CosmicObjectTypes) deleteIndexes(cosmicObjectType *CosmicObjectType) {
	delete(cosmicObjectTypes.ByTitleRu, cosmicObjectType.TitleRu)
	delete(cosmicObjectTypes.ByTitleEn, cosmicObjectType.TitleEn)
	delete(cosmicObjectTypes.ByAcronym, cosmicObjectType.Acronym)
}
