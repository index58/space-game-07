package data

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"sort"
)

// TaskType хранит справочник типов заданий.
type TaskType struct {
	ID      int64  `json:"ID"`      // Уникальный числовой идентификатор записи.
	TitleRu string `json:"TitleRu"` // Русское название для интерфейса и данных.
	TitleEn string `json:"TitleEn"` // Английское название для интерфейса и данных.
	Acronym string `json:"Acronym"` // Неизменяемый строковый идентификатор для логики.
}

// TaskTypes хранит типы заданий и быстрые индексы по уникальным полям.
type TaskTypes struct {
	MaxID int64               `json:"MaxID"` // Последний выданный числовой идентификатор записей.
	Items map[int64]*TaskType `json:"Items"` // Основное хранилище записей по числовому идентификатору.

	ByAcronym map[string]*TaskType `json:"-"` // Быстрый поиск записи по неизменяемому строковому идентификатору.
}

// NewTaskTypes создает пустой справочник типов заданий.
func NewTaskTypes() *TaskTypes {
	types := &TaskTypes{}
	types.ensureMaps()
	return types
}

// Add добавляет новый тип задания и назначает новый ID.
func (types *TaskTypes) Add(taskType *TaskType) (*TaskType, error) {
	if taskType == nil {
		return nil, errors.New("task type is nil")
	}
	types.ensureMaps()
	if err := types.validateRequiredFields(taskType); err != nil {
		return nil, err
	}
	if err := types.ensureUniqueForNewType(taskType); err != nil {
		return nil, err
	}

	types.MaxID++
	taskType.ID = types.MaxID
	types.Items[taskType.ID] = taskType
	types.addIndexes(taskType)
	return taskType, nil
}

// Get возвращает тип задания по ID.
func (types *TaskTypes) Get(id int64) (*TaskType, bool) {
	types.ensureMaps()
	taskType, ok := types.Items[id]
	return taskType, ok
}

// GetByAcronym возвращает тип задания по акрониму.
func (types *TaskTypes) GetByAcronym(acronym string) (*TaskType, bool) {
	types.ensureMaps()
	taskType, ok := types.ByAcronym[acronym]
	return taskType, ok
}

// RebuildIndexes пересобирает индексы после загрузки из JSON.
func (types *TaskTypes) RebuildIndexes() error {
	types.ensureItems()
	types.ByAcronym = make(map[string]*TaskType)

	var maxID int64
	ids := make([]int64, 0, len(types.Items))
	for id := range types.Items {
		ids = append(ids, id)
	}
	sort.Slice(ids, func(left int, right int) bool { return ids[left] < ids[right] })
	for _, id := range ids {
		taskType := types.Items[id]
		if taskType == nil {
			return fmt.Errorf("task type with ID %d is nil", id)
		}
		if taskType.ID != id {
			return fmt.Errorf("task type map key %d does not match type ID %d", id, taskType.ID)
		}
		if err := types.validateRequiredFields(taskType); err != nil {
			return fmt.Errorf("task type with ID %d is invalid: %w", id, err)
		}
		if id > maxID {
			maxID = id
		}
		if err := types.ensureUniqueForNewType(taskType); err != nil {
			return err
		}
		types.addIndexes(taskType)
	}
	if types.MaxID < maxID {
		types.MaxID = maxID
	}
	return nil
}

// LoadFromFile загружает справочник типов заданий из JSON-файла.
func (types *TaskTypes) LoadFromFile(path string) error {
	content, err := os.ReadFile(path)
	if err != nil {
		return err
	}

	loaded := TaskTypes{}
	if err := json.Unmarshal(content, &loaded); err != nil {
		return err
	}
	if err := loaded.RebuildIndexes(); err != nil {
		return err
	}

	*types = loaded
	return nil
}

// SaveToFile сохраняет справочник типов заданий в JSON-файл.
func (types *TaskTypes) SaveToFile(path string) error {
	types.ensureMaps()
	return saveTableWithOrderedItems(path, types.MaxID, types.Items)
}

// ensureMaps подготавливает основное хранилище и индексы.
func (types *TaskTypes) ensureMaps() {
	types.ensureItems()
	if types.ByAcronym == nil {
		types.ByAcronym = make(map[string]*TaskType)
	}
}

// ensureItems подготавливает основное хранилище.
func (types *TaskTypes) ensureItems() {
	if types.Items == nil {
		types.Items = make(map[int64]*TaskType)
	}
}

// validateRequiredFields проверяет обязательные поля типа задания.
func (types *TaskTypes) validateRequiredFields(taskType *TaskType) error {
	if taskType.TitleRu == "" {
		return errors.New("title ru is empty")
	}
	if taskType.TitleEn == "" {
		return errors.New("title en is empty")
	}
	if taskType.Acronym == "" {
		return errors.New("acronym is empty")
	}
	return nil
}

// ensureUniqueForNewType проверяет уникальность акронима перед добавлением в индексы.
func (types *TaskTypes) ensureUniqueForNewType(taskType *TaskType) error {
	if existing, ok := types.ByAcronym[taskType.Acronym]; ok && existing.ID != taskType.ID {
		return fmt.Errorf("acronym %q already exists", taskType.Acronym)
	}
	return nil
}

// addIndexes добавляет тип задания в быстрые индексы.
func (types *TaskTypes) addIndexes(taskType *TaskType) {
	types.ByAcronym[taskType.Acronym] = taskType
}
