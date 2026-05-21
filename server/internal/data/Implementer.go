package data

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"sort"
)

// Implementer хранит тип оборудования, выполняющий часть работы задания.
type Implementer struct {
	ID                             int64   `json:"ID"`                             // Уникальный числовой идентификатор записи.
	TaskTypeID                     int64   `json:"TaskTypeID"`                     // Тип задания, для которого используется исполнитель.
	ImplementerEquipmentItemTypeID int64   `json:"ImplementerEquipmentItemTypeID"` // Тип предмета-оборудования, выполняющего работу.
	WorkPart                       float64 `json:"WorkPart"`                       // Доля работы этого типа оборудования среди исполнителей задания.
}

// Implementers хранит исполнителей заданий и индексы для поиска по типу задания.
type Implementers struct {
	MaxID int64                  `json:"MaxID"` // Последний выданный числовой идентификатор записей.
	Items map[int64]*Implementer `json:"Items"` // Основное хранилище записей по числовому идентификатору.

	ByTaskTypeID map[int64][]*Implementer        `json:"-"` // Быстрый поиск исполнителей по типу задания.
	ByUniqueKey  map[implementerKey]*Implementer `json:"-"` // Быстрый поиск уникальной пары типа задания и оборудования.
}

type implementerKey struct {
	TaskTypeID                     int64
	ImplementerEquipmentItemTypeID int64
}

// NewImplementers создает пустое хранилище исполнителей.
func NewImplementers() *Implementers {
	implementers := &Implementers{}
	implementers.ensureMaps()
	return implementers
}

// Add добавляет исполнителя и назначает новый ID.
func (implementers *Implementers) Add(implementer *Implementer) (*Implementer, error) {
	if implementer == nil {
		return nil, errors.New("implementer is nil")
	}
	implementers.ensureMaps()
	if err := implementers.validateRequiredFields(implementer); err != nil {
		return nil, err
	}
	if err := implementers.ensureUniqueForNewImplementer(implementer); err != nil {
		return nil, err
	}

	implementers.MaxID++
	implementer.ID = implementers.MaxID
	implementers.Items[implementer.ID] = implementer
	implementers.addIndexes(implementer)
	return implementer, nil
}

// GetByTaskTypeID возвращает исполнителей указанного типа задания.
func (implementers *Implementers) GetByTaskTypeID(taskTypeID int64) []*Implementer {
	implementers.ensureMaps()
	return implementers.ByTaskTypeID[taskTypeID]
}

// RebuildIndexes пересобирает индексы после загрузки из JSON.
func (implementers *Implementers) RebuildIndexes() error {
	implementers.ensureItems()
	implementers.ByTaskTypeID = make(map[int64][]*Implementer)
	implementers.ByUniqueKey = make(map[implementerKey]*Implementer)

	var maxID int64
	ids := make([]int64, 0, len(implementers.Items))
	for id := range implementers.Items {
		ids = append(ids, id)
	}
	sort.Slice(ids, func(left int, right int) bool { return ids[left] < ids[right] })
	for _, id := range ids {
		implementer := implementers.Items[id]
		if implementer == nil {
			return fmt.Errorf("implementer with ID %d is nil", id)
		}
		if implementer.ID != id {
			return fmt.Errorf("implementer map key %d does not match implementer ID %d", id, implementer.ID)
		}
		if err := implementers.validateRequiredFields(implementer); err != nil {
			return fmt.Errorf("implementer with ID %d is invalid: %w", id, err)
		}
		if id > maxID {
			maxID = id
		}
		if err := implementers.ensureUniqueForNewImplementer(implementer); err != nil {
			return err
		}
		implementers.addIndexes(implementer)
	}
	if implementers.MaxID < maxID {
		implementers.MaxID = maxID
	}
	return nil
}

// LoadFromFile загружает исполнителей из JSON-файла.
func (implementers *Implementers) LoadFromFile(path string) error {
	content, err := os.ReadFile(path)
	if err != nil {
		return err
	}

	loaded := Implementers{}
	if err := json.Unmarshal(content, &loaded); err != nil {
		return err
	}
	if err := loaded.RebuildIndexes(); err != nil {
		return err
	}

	*implementers = loaded
	return nil
}

// SaveToFile сохраняет исполнителей в JSON-файл.
func (implementers *Implementers) SaveToFile(path string) error {
	implementers.ensureMaps()
	return saveTableWithOrderedItems(path, implementers.MaxID, implementers.Items)
}

// ensureMaps подготавливает хранилище и индексы.
func (implementers *Implementers) ensureMaps() {
	implementers.ensureItems()
	if implementers.ByTaskTypeID == nil {
		implementers.ByTaskTypeID = make(map[int64][]*Implementer)
	}
	if implementers.ByUniqueKey == nil {
		implementers.ByUniqueKey = make(map[implementerKey]*Implementer)
	}
}

// ensureItems подготавливает основное хранилище.
func (implementers *Implementers) ensureItems() {
	if implementers.Items == nil {
		implementers.Items = make(map[int64]*Implementer)
	}
}

// validateRequiredFields проверяет обязательные поля исполнителя.
func (implementers *Implementers) validateRequiredFields(implementer *Implementer) error {
	if implementer.TaskTypeID <= 0 {
		return errors.New("task type ID is empty")
	}
	if implementer.ImplementerEquipmentItemTypeID <= 0 {
		return errors.New("implementer equipment item type ID is empty")
	}
	if implementer.WorkPart < 0 {
		return errors.New("work part is negative")
	}
	return nil
}

// ensureUniqueForNewImplementer проверяет уникальность пары задания и оборудования.
func (implementers *Implementers) ensureUniqueForNewImplementer(implementer *Implementer) error {
	key := implementerKey{TaskTypeID: implementer.TaskTypeID, ImplementerEquipmentItemTypeID: implementer.ImplementerEquipmentItemTypeID}
	if existing, ok := implementers.ByUniqueKey[key]; ok && existing.ID != implementer.ID {
		return fmt.Errorf("implementer for task type %d and equipment item type %d already exists", implementer.TaskTypeID, implementer.ImplementerEquipmentItemTypeID)
	}
	return nil
}

// addIndexes добавляет исполнителя в быстрые индексы.
func (implementers *Implementers) addIndexes(implementer *Implementer) {
	key := implementerKey{TaskTypeID: implementer.TaskTypeID, ImplementerEquipmentItemTypeID: implementer.ImplementerEquipmentItemTypeID}
	implementers.ByUniqueKey[key] = implementer
	implementers.ByTaskTypeID[implementer.TaskTypeID] = append(implementers.ByTaskTypeID[implementer.TaskTypeID], implementer)
}
