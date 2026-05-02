package data

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"sort"
)

// Хранит группу одинаковых предметов внутри контейнера.
type ItemGroup struct {
	ID                        int64   `json:"ID"`                        // Уникальный числовой идентификатор записи.
	ContainerEquipmentGroupID int64   `json:"ContainerEquipmentGroupID"` // Группа оборудования, внутри которой находится содержимое.
	ContentItemModelID        int64   `json:"ContentItemModelID"`        // Модель предмета, лежащего внутри контейнера.
	Count                     float64 `json:"Count"`                     // Количество предметов указанной модели.
}

// Хранит группы предметов внутри установленного контейнерного оборудования.
type ItemGroups struct {
	MaxID int64                `json:"MaxID"` // Последний выданный числовой идентификатор для новых записей.
	Items map[int64]*ItemGroup `json:"Items"` // Основное хранилище записей по числовому идентификатору.

	ByContainerEquipmentGroupID map[int64][]*ItemGroup `json:"-"` // Быстрый поиск содержимого по группе контейнерного оборудования.
}

// Создает пустое хранилище групп предметов с подготовленными индексами.
func NewItemGroups() *ItemGroups {
	groups := &ItemGroups{}
	groups.ensureMaps()
	return groups
}

// Добавляет новую группу предметов и назначает новый ID.
func (groups *ItemGroups) Add(group *ItemGroup) (*ItemGroup, error) {
	if group == nil {
		return nil, errors.New("item group is nil")
	}
	groups.ensureMaps()
	if err := groups.validateRequiredFields(group); err != nil {
		return nil, err
	}

	groups.MaxID++
	group.ID = groups.MaxID
	groups.Items[group.ID] = group
	groups.addIndexes(group)
	return group, nil
}

// Возвращает группу предметов по ID.
func (groups *ItemGroups) Get(id int64) (*ItemGroup, bool) {
	groups.ensureMaps()
	group, ok := groups.Items[id]
	return group, ok
}

// Возвращает содержимое указанной группы контейнерного оборудования.
func (groups *ItemGroups) GetByContainerEquipmentGroupID(containerEquipmentGroupID int64) []*ItemGroup {
	groups.ensureMaps()
	return groups.ByContainerEquipmentGroupID[containerEquipmentGroupID]
}

// Удаляет все содержимое указанной группы контейнерного оборудования.
func (groups *ItemGroups) DeleteByContainerEquipmentGroupID(containerEquipmentGroupID int64) {
	groups.ensureMaps()
	for _, group := range groups.ByContainerEquipmentGroupID[containerEquipmentGroupID] {
		delete(groups.Items, group.ID)
	}
	delete(groups.ByContainerEquipmentGroupID, containerEquipmentGroupID)
}

// Удаляет содержимое нескольких групп контейнерного оборудования.
func (groups *ItemGroups) DeleteByContainerEquipmentGroupIDs(containerEquipmentGroupIDs []int64) {
	for _, containerEquipmentGroupID := range containerEquipmentGroupIDs {
		groups.DeleteByContainerEquipmentGroupID(containerEquipmentGroupID)
	}
}

// Пересобирает быстрые индексы после загрузки из JSON.
func (groups *ItemGroups) RebuildIndexes() error {
	groups.ensureItems()
	groups.ByContainerEquipmentGroupID = make(map[int64][]*ItemGroup)

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
			return fmt.Errorf("item group with ID %d is nil", id)
		}
		if group.ID != id {
			return fmt.Errorf("item group map key %d does not match group ID %d", id, group.ID)
		}
		if err := groups.validateRequiredFields(group); err != nil {
			return fmt.Errorf("item group with ID %d is invalid: %w", id, err)
		}
		if id > maxID {
			maxID = id
		}
		groups.addIndexes(group)
	}
	if groups.MaxID < maxID {
		groups.MaxID = maxID
	}
	return nil
}

// Загружает группы предметов из JSON-файла и пересобирает быстрые индексы.
func (groups *ItemGroups) LoadFromFile(path string) error {
	content, err := os.ReadFile(path)
	if err != nil {
		return err
	}

	loaded := ItemGroups{}
	if err := json.Unmarshal(content, &loaded); err != nil {
		return err
	}
	if err := loaded.RebuildIndexes(); err != nil {
		return err
	}

	*groups = loaded
	return nil
}

// Сохраняет группы предметов в JSON-файл без вспомогательных индексов.
func (groups *ItemGroups) SaveToFile(path string) error {
	groups.ensureMaps()
	return saveTableWithOrderedItems(path, groups.MaxID, groups.Items)
}

// Подготавливает основное хранилище и быстрые индексы.
func (groups *ItemGroups) ensureMaps() {
	groups.ensureItems()
	if groups.ByContainerEquipmentGroupID == nil {
		groups.ByContainerEquipmentGroupID = make(map[int64][]*ItemGroup)
	}
}

// Подготавливает основную map.
func (groups *ItemGroups) ensureItems() {
	if groups.Items == nil {
		groups.Items = make(map[int64]*ItemGroup)
	}
}

// Проверяет обязательные поля группы предметов.
func (groups *ItemGroups) validateRequiredFields(group *ItemGroup) error {
	if group.ContainerEquipmentGroupID <= 0 {
		return errors.New("container equipment group ID is empty")
	}
	if group.ContentItemModelID <= 0 {
		return errors.New("content item model ID is empty")
	}
	if group.Count <= 0 {
		return errors.New("count is empty")
	}
	return nil
}

// Добавляет группу предметов в быстрые индексы.
func (groups *ItemGroups) addIndexes(group *ItemGroup) {
	groups.ByContainerEquipmentGroupID[group.ContainerEquipmentGroupID] = append(groups.ByContainerEquipmentGroupID[group.ContainerEquipmentGroupID], group)
}
