package data

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"sort"
)

// TaskItemGroup хранит предметы, зарезервированные для задания.
type TaskItemGroup struct {
	ID          int64   `json:"ID"`          // Уникальный числовой идентификатор записи.
	TaskID      int64   `json:"TaskID"`      // Задание, для которого зарезервированы предметы.
	ItemModelID int64   `json:"ItemModelID"` // Модель зарезервированных предметов.
	Count       float64 `json:"Count"`       // Количество зарезервированных предметов.
	IsStored    bool    `json:"IsStored"`    // Хранится ли указанное количество предметов во временном хранилище задания.
}

// TaskItemGroups хранит резервы заданий и быстрые индексы.
type TaskItemGroups struct {
	MaxID int64                    `json:"MaxID"` // Последний выданный числовой идентификатор записей.
	Items map[int64]*TaskItemGroup `json:"Items"` // Основное хранилище записей по числовому идентификатору.

	ByTaskID    map[int64][]*TaskItemGroup          `json:"-"` // Быстрый поиск резервов по заданию.
	ByUniqueKey map[taskItemGroupKey]*TaskItemGroup `json:"-"` // Быстрый поиск уникальной пары задания и модели.
}

type taskItemGroupKey struct {
	TaskID      int64
	ItemModelID int64
}

// NewTaskItemGroups создает пустое хранилище резервов заданий.
func NewTaskItemGroups() *TaskItemGroups {
	groups := &TaskItemGroups{}
	groups.ensureMaps()
	return groups
}

// Add добавляет резерв задания и назначает новый ID.
func (groups *TaskItemGroups) Add(group *TaskItemGroup) (*TaskItemGroup, error) {
	if group == nil {
		return nil, errors.New("task item group is nil")
	}
	groups.ensureMaps()
	if err := groups.validateRequiredFields(group); err != nil {
		return nil, err
	}
	if err := groups.ensureUniqueForNewGroup(group); err != nil {
		return nil, err
	}

	groups.MaxID++
	group.ID = groups.MaxID
	groups.Items[group.ID] = group
	groups.addIndexes(group)
	return group, nil
}

// GetByTaskID возвращает резервы указанного задания.
func (groups *TaskItemGroups) GetByTaskID(taskID int64) []*TaskItemGroup {
	groups.ensureMaps()
	return groups.ByTaskID[taskID]
}

// DeleteByTaskID удаляет все резервы указанного задания.
func (groups *TaskItemGroups) DeleteByTaskID(taskID int64) {
	groups.ensureMaps()
	for _, group := range groups.ByTaskID[taskID] {
		delete(groups.Items, group.ID)
	}
	groups.RebuildIndexes()
}

// RebuildIndexes пересобирает индексы после загрузки из JSON.
func (groups *TaskItemGroups) RebuildIndexes() error {
	groups.ensureItems()
	groups.ByTaskID = make(map[int64][]*TaskItemGroup)
	groups.ByUniqueKey = make(map[taskItemGroupKey]*TaskItemGroup)

	var maxID int64
	ids := make([]int64, 0, len(groups.Items))
	for id := range groups.Items {
		ids = append(ids, id)
	}
	sort.Slice(ids, func(left int, right int) bool { return ids[left] < ids[right] })
	for _, id := range ids {
		group := groups.Items[id]
		if group == nil {
			return fmt.Errorf("task item group with ID %d is nil", id)
		}
		if group.ID != id {
			return fmt.Errorf("task item group map key %d does not match group ID %d", id, group.ID)
		}
		if err := groups.validateRequiredFields(group); err != nil {
			return fmt.Errorf("task item group with ID %d is invalid: %w", id, err)
		}
		if id > maxID {
			maxID = id
		}
		if err := groups.ensureUniqueForNewGroup(group); err != nil {
			return err
		}
		groups.addIndexes(group)
	}
	if groups.MaxID < maxID {
		groups.MaxID = maxID
	}
	return nil
}

// LoadFromFile загружает резервы заданий из JSON-файла.
func (groups *TaskItemGroups) LoadFromFile(path string) error {
	content, err := os.ReadFile(path)
	if err != nil {
		return err
	}

	loaded := TaskItemGroups{}
	if err := json.Unmarshal(content, &loaded); err != nil {
		return err
	}
	if err := loaded.RebuildIndexes(); err != nil {
		return err
	}

	*groups = loaded
	return nil
}

// SaveToFile сохраняет резервы заданий в JSON-файл.
func (groups *TaskItemGroups) SaveToFile(path string) error {
	groups.ensureMaps()
	return saveTableWithOrderedItems(path, groups.MaxID, groups.Items)
}

// ensureMaps подготавливает хранилище и индексы.
func (groups *TaskItemGroups) ensureMaps() {
	groups.ensureItems()
	if groups.ByTaskID == nil {
		groups.ByTaskID = make(map[int64][]*TaskItemGroup)
	}
	if groups.ByUniqueKey == nil {
		groups.ByUniqueKey = make(map[taskItemGroupKey]*TaskItemGroup)
	}
}

// ensureItems подготавливает основное хранилище.
func (groups *TaskItemGroups) ensureItems() {
	if groups.Items == nil {
		groups.Items = make(map[int64]*TaskItemGroup)
	}
}

// validateRequiredFields проверяет обязательные поля резерва.
func (groups *TaskItemGroups) validateRequiredFields(group *TaskItemGroup) error {
	if group.TaskID <= 0 {
		return errors.New("task ID is empty")
	}
	if group.ItemModelID <= 0 {
		return errors.New("item model ID is empty")
	}
	if group.Count <= 0 {
		return errors.New("count is empty")
	}
	return nil
}

// ensureUniqueForNewGroup проверяет уникальность пары задания и модели.
func (groups *TaskItemGroups) ensureUniqueForNewGroup(group *TaskItemGroup) error {
	key := taskItemGroupKey{TaskID: group.TaskID, ItemModelID: group.ItemModelID}
	if existing, ok := groups.ByUniqueKey[key]; ok && existing.ID != group.ID {
		return fmt.Errorf("task item group for task %d and item model %d already exists", group.TaskID, group.ItemModelID)
	}
	return nil
}

// addIndexes добавляет резерв в быстрые индексы.
func (groups *TaskItemGroups) addIndexes(group *TaskItemGroup) {
	key := taskItemGroupKey{TaskID: group.TaskID, ItemModelID: group.ItemModelID}
	groups.ByUniqueKey[key] = group
	groups.ByTaskID[group.TaskID] = append(groups.ByTaskID[group.TaskID], group)
}
