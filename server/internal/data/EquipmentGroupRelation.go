package data

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"sort"
)

// Хранит вид связи между группами оборудования.
type RelationType struct {
	ID      int64  `json:"ID"`      // Уникальный числовой идентификатор записи.
	TitleRu string `json:"TitleRu"` // Русское название для интерфейса и данных.
	TitleEn string `json:"TitleEn"` // Английское название для интерфейса и данных.
	Acronym string `json:"Acronym"` // Неизменяемый строковый идентификатор для логики.
}

// Хранит справочник видов связей с быстрым поиском по акрониму.
type RelationTypes struct {
	MaxID int64                   `json:"MaxID"` // Последний выданный числовой идентификатор записей.
	Items map[int64]*RelationType `json:"Items"` // Основное хранилище записей по числовому идентификатору.

	ByAcronym map[string]*RelationType `json:"-"` // Быстрый поиск записи по неизменяемому строковому идентификатору.
}

// Создаёт пустой справочник видов связей.
func NewRelationTypes() *RelationTypes {
	types := &RelationTypes{}
	types.ensureMaps()
	return types
}

// Добавляет новый вид связи и назначает ему новый ID.
func (types *RelationTypes) Add(relationType *RelationType) (*RelationType, error) {
	if relationType == nil {
		return nil, errors.New("relation type is nil")
	}
	types.ensureMaps()
	if err := types.validateRequiredFields(relationType); err != nil {
		return nil, err
	}
	if err := types.ensureUniqueForNewType(relationType); err != nil {
		return nil, err
	}

	types.MaxID++
	relationType.ID = types.MaxID
	types.Items[relationType.ID] = relationType
	types.addIndexes(relationType)
	return relationType, nil
}

// Возвращает вид связи по ID.
func (types *RelationTypes) Get(id int64) (*RelationType, bool) {
	types.ensureMaps()
	relationType, ok := types.Items[id]
	return relationType, ok
}

// Возвращает вид связи по акрониму.
func (types *RelationTypes) GetByAcronym(acronym string) (*RelationType, bool) {
	types.ensureMaps()
	relationType, ok := types.ByAcronym[acronym]
	return relationType, ok
}

// Пересобирает быстрые индексы после загрузки из JSON.
func (types *RelationTypes) RebuildIndexes() error {
	types.ensureItems()
	types.ByAcronym = make(map[string]*RelationType)

	var maxID int64
	ids := make([]int64, 0, len(types.Items))
	for id := range types.Items {
		ids = append(ids, id)
	}
	sort.Slice(ids, func(left int, right int) bool {
		return ids[left] < ids[right]
	})
	for _, id := range ids {
		relationType := types.Items[id]
		if relationType == nil {
			return fmt.Errorf("relation type with ID %d is nil", id)
		}
		if relationType.ID != id {
			return fmt.Errorf("relation type map key %d does not match type ID %d", id, relationType.ID)
		}
		if err := types.validateRequiredFields(relationType); err != nil {
			return fmt.Errorf("relation type with ID %d is invalid: %w", id, err)
		}
		if id > maxID {
			maxID = id
		}
		if err := types.ensureUniqueForNewType(relationType); err != nil {
			return err
		}
		types.addIndexes(relationType)
	}
	if types.MaxID < maxID {
		types.MaxID = maxID
	}
	return nil
}

// Загружает справочник видов связей из JSON-файла.
func (types *RelationTypes) LoadFromFile(path string) error {
	content, err := os.ReadFile(path)
	if err != nil {
		return err
	}

	loaded := RelationTypes{}
	if err := json.Unmarshal(content, &loaded); err != nil {
		return err
	}
	if err := loaded.RebuildIndexes(); err != nil {
		return err
	}

	*types = loaded
	return nil
}

// Сохраняет справочник видов связей в JSON-файл.
func (types *RelationTypes) SaveToFile(path string) error {
	types.ensureMaps()
	return saveTableWithOrderedItems(path, types.MaxID, types.Items)
}

// Подготавливает основное хранилище и быстрые индексы.
func (types *RelationTypes) ensureMaps() {
	types.ensureItems()
	if types.ByAcronym == nil {
		types.ByAcronym = make(map[string]*RelationType)
	}
}

// Подготавливает основную map.
func (types *RelationTypes) ensureItems() {
	if types.Items == nil {
		types.Items = make(map[int64]*RelationType)
	}
}

// Проверяет обязательные поля вида связи.
func (types *RelationTypes) validateRequiredFields(relationType *RelationType) error {
	if relationType.TitleRu == "" {
		return errors.New("title ru is empty")
	}
	if relationType.TitleEn == "" {
		return errors.New("title en is empty")
	}
	if relationType.Acronym == "" {
		return errors.New("acronym is empty")
	}
	return nil
}

// Проверяет уникальность акронима перед добавлением в индексы.
func (types *RelationTypes) ensureUniqueForNewType(relationType *RelationType) error {
	if existing, ok := types.ByAcronym[relationType.Acronym]; ok && existing.ID != relationType.ID {
		return fmt.Errorf("acronym %q already exists", relationType.Acronym)
	}
	return nil
}

// Добавляет вид связи в быстрые индексы.
func (types *RelationTypes) addIndexes(relationType *RelationType) {
	types.ByAcronym[relationType.Acronym] = relationType
}

// Хранит сохранённую связь одной группы оборудования с другой.
type EquipmentGroupRelation struct {
	ID                      int64 `json:"ID"`                      // Уникальный числовой идентификатор записи.
	EquipmentGroupID        int64 `json:"EquipmentGroupID"`        // Группа оборудования, для которой сохранена связь.
	RelationTypeID          int64 `json:"RelationTypeID"`          // Вид связи между группами оборудования.
	RelatedEquipmentGroupID int64 `json:"RelatedEquipmentGroupID"` // Группа оборудования, выбранная для указанного вида связи.
}

// Хранит связи групп оборудования с быстрым поиском по группе и виду связи.
type EquipmentGroupRelations struct {
	MaxID int64                             `json:"MaxID"` // Последний выданный числовой идентификатор записей.
	Items map[int64]*EquipmentGroupRelation `json:"Items"` // Основное хранилище записей по числовому идентификатору.

	ByEquipmentGroupAndType map[equipmentGroupRelationKey]*EquipmentGroupRelation `json:"-"` // Быстрый поиск текущей связи по группе и виду связи.
}

type equipmentGroupRelationKey struct {
	EquipmentGroupID int64
	RelationTypeID   int64
}

// Создаёт пустое хранилище связей групп оборудования.
func NewEquipmentGroupRelations() *EquipmentGroupRelations {
	relations := &EquipmentGroupRelations{}
	relations.ensureMaps()
	return relations
}

// Добавляет или заменяет текущую связь для пары группа + вид связи.
func (relations *EquipmentGroupRelations) Upsert(relation *EquipmentGroupRelation) (*EquipmentGroupRelation, error) {
	if relation == nil {
		return nil, errors.New("equipment group relation is nil")
	}
	relations.ensureMaps()
	if err := relations.validateRequiredFields(relation); err != nil {
		return nil, err
	}

	key := equipmentGroupRelationKey{EquipmentGroupID: relation.EquipmentGroupID, RelationTypeID: relation.RelationTypeID}
	if existing, ok := relations.ByEquipmentGroupAndType[key]; ok {
		existing.RelatedEquipmentGroupID = relation.RelatedEquipmentGroupID
		return existing, nil
	}
	relations.MaxID++
	relation.ID = relations.MaxID
	relations.Items[relation.ID] = relation
	relations.addIndexes(relation)
	return relation, nil
}

// Возвращает связь по группе и виду связи.
func (relations *EquipmentGroupRelations) GetByEquipmentGroupAndType(equipmentGroupID int64, relationTypeID int64) (*EquipmentGroupRelation, bool) {
	relations.ensureMaps()
	relation, ok := relations.ByEquipmentGroupAndType[equipmentGroupRelationKey{EquipmentGroupID: equipmentGroupID, RelationTypeID: relationTypeID}]
	return relation, ok
}

// Пересобирает быстрые индексы после загрузки из JSON.
func (relations *EquipmentGroupRelations) RebuildIndexes() error {
	relations.ensureItems()
	relations.ByEquipmentGroupAndType = make(map[equipmentGroupRelationKey]*EquipmentGroupRelation)

	var maxID int64
	ids := make([]int64, 0, len(relations.Items))
	for id := range relations.Items {
		ids = append(ids, id)
	}
	sort.Slice(ids, func(left int, right int) bool {
		return ids[left] < ids[right]
	})
	for _, id := range ids {
		relation := relations.Items[id]
		if relation == nil {
			return fmt.Errorf("equipment group relation with ID %d is nil", id)
		}
		if relation.ID != id {
			return fmt.Errorf("equipment group relation map key %d does not match relation ID %d", id, relation.ID)
		}
		if err := relations.validateRequiredFields(relation); err != nil {
			return fmt.Errorf("equipment group relation with ID %d is invalid: %w", id, err)
		}
		if id > maxID {
			maxID = id
		}
		key := equipmentGroupRelationKey{EquipmentGroupID: relation.EquipmentGroupID, RelationTypeID: relation.RelationTypeID}
		if existing, ok := relations.ByEquipmentGroupAndType[key]; ok && existing.ID != relation.ID {
			return fmt.Errorf("equipment group relation for group %d and type %d already exists", relation.EquipmentGroupID, relation.RelationTypeID)
		}
		relations.addIndexes(relation)
	}
	if relations.MaxID < maxID {
		relations.MaxID = maxID
	}
	return nil
}

// Загружает связи групп оборудования из JSON-файла.
func (relations *EquipmentGroupRelations) LoadFromFile(path string) error {
	content, err := os.ReadFile(path)
	if err != nil {
		return err
	}

	loaded := EquipmentGroupRelations{}
	if err := json.Unmarshal(content, &loaded); err != nil {
		return err
	}
	if err := loaded.RebuildIndexes(); err != nil {
		return err
	}

	*relations = loaded
	return nil
}

// Сохраняет связи групп оборудования в JSON-файл.
func (relations *EquipmentGroupRelations) SaveToFile(path string) error {
	relations.ensureMaps()
	return saveTableWithOrderedItems(path, relations.MaxID, relations.Items)
}

// Подготавливает основное хранилище и быстрые индексы.
func (relations *EquipmentGroupRelations) ensureMaps() {
	relations.ensureItems()
	if relations.ByEquipmentGroupAndType == nil {
		relations.ByEquipmentGroupAndType = make(map[equipmentGroupRelationKey]*EquipmentGroupRelation)
	}
}

// Подготавливает основную map.
func (relations *EquipmentGroupRelations) ensureItems() {
	if relations.Items == nil {
		relations.Items = make(map[int64]*EquipmentGroupRelation)
	}
}

// Проверяет обязательные поля связи.
func (relations *EquipmentGroupRelations) validateRequiredFields(relation *EquipmentGroupRelation) error {
	if relation.EquipmentGroupID <= 0 {
		return errors.New("equipment group ID is empty")
	}
	if relation.RelationTypeID <= 0 {
		return errors.New("relation type ID is empty")
	}
	if relation.RelatedEquipmentGroupID <= 0 {
		return errors.New("related equipment group ID is empty")
	}
	return nil
}

// Добавляет связь в быстрые индексы.
func (relations *EquipmentGroupRelations) addIndexes(relation *EquipmentGroupRelation) {
	relations.ByEquipmentGroupAndType[equipmentGroupRelationKey{EquipmentGroupID: relation.EquipmentGroupID, RelationTypeID: relation.RelationTypeID}] = relation
}
