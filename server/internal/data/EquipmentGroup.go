package data

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"sort"
)

// Добавляет группу оборудования в быстрые индексы.
type EquipmentGroup struct {
	ID                          int64  `json:"ID"`                          // Уникальный числовой идентификатор записи.
	CosmicObjectID              int64  `json:"CosmicObjectID"`              // Космический объект, на котором установлено оборудование.
	Title                       string `json:"Title"`                       // Название группы установленного оборудования.
	EquipmentItemModelID        int64  `json:"EquipmentItemModelID"`        // Модель установленного оборудования.
	Count                       int64  `json:"Count"`                       // Количество установленных единиц оборудования.
	EnabledCount                int64  `json:"EnabledCount"`                // Количество включенных единиц оборудования.
	Enabled                     bool   `json:"Enabled"`                     // Признак включения группы оборудования.
	Active                      bool   `json:"Active"`                      // Признак выполнения работы включенным оборудованием.
	MagazineCount               int64  `json:"MagazineCount"`               // Количество боеприпасов, уже заряженных для ближайших выстрелов.
	LastRechargeStartTime       int64  `json:"LastRechargeStartTime"`       // Время начала последней перезарядки в миллисекундах Unix.
	SourceEquipmentGroupID      int64  `json:"SourceEquipmentGroupID"`      // Источник материалов или груза для работы оборудования.
	DestinationEquipmentGroupID int64  `json:"DestinationEquipmentGroupID"` // Приемник результата или груза после работы оборудования.
	OppositeEquipmentGroupID    int64  `json:"OppositeEquipmentGroupID"`    // Противоположная группа оборудования в парном интерфейсе использования.
}

// Добавляет группу оборудования в быстрые индексы.
type EquipmentGroups struct {
	MaxID int64                     `json:"MaxID"` // Последний выданный числовой идентификатор записей.
	Items map[int64]*EquipmentGroup `json:"Items"` // Основное хранилище записей по числовому идентификатору.

	ByCosmicObjectID map[int64][]*EquipmentGroup `json:"-"` // Быстрый поиск групп оборудования по объекту.
}

// Добавляет группу оборудования в быстрые индексы.
func NewEquipmentGroups() *EquipmentGroups {
	groups := &EquipmentGroups{}
	groups.ensureMaps()
	return groups
}

// Добавляет группу оборудования в быстрые индексы.
func (groups *EquipmentGroups) Add(group *EquipmentGroup) (*EquipmentGroup, error) {
	if group == nil {
		return nil, errors.New("equipment group is nil")
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

// Добавляет группу оборудования в быстрые индексы.
func (groups *EquipmentGroups) Get(id int64) (*EquipmentGroup, bool) {
	groups.ensureMaps()
	group, ok := groups.Items[id]
	return group, ok
}

// Добавляет группу оборудования в быстрые индексы.
func (groups *EquipmentGroups) GetByCosmicObjectID(cosmicObjectID int64) []*EquipmentGroup {
	groups.ensureMaps()
	return groups.ByCosmicObjectID[cosmicObjectID]
}

// Добавляет группу оборудования в быстрые индексы.
func (groups *EquipmentGroups) DeleteByCosmicObjectID(cosmicObjectID int64) {
	groups.ensureMaps()
	for _, group := range groups.ByCosmicObjectID[cosmicObjectID] {
		delete(groups.Items, group.ID)
	}
	delete(groups.ByCosmicObjectID, cosmicObjectID)
}

// Добавляет группу оборудования в быстрые индексы.
func (groups *EquipmentGroups) RebuildIndexes() error {
	groups.ensureItems()
	groups.ByCosmicObjectID = make(map[int64][]*EquipmentGroup)

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
			return fmt.Errorf("equipment group with ID %d is nil", id)
		}
		if group.ID != id {
			return fmt.Errorf("equipment group map key %d does not match group ID %d", id, group.ID)
		}
		if err := groups.validateRequiredFields(group); err != nil {
			return fmt.Errorf("equipment group with ID %d is invalid: %w", id, err)
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

// Добавляет группу оборудования в быстрые индексы.
func (groups *EquipmentGroups) LoadFromFile(path string) error {
	content, err := os.ReadFile(path)
	if err != nil {
		return err
	}

	loaded := EquipmentGroups{}
	if err := json.Unmarshal(content, &loaded); err != nil {
		return err
	}
	if err := loaded.RebuildIndexes(); err != nil {
		return err
	}

	*groups = loaded
	return nil
}

// Добавляет группу оборудования в быстрые индексы.
func (groups *EquipmentGroups) SaveToFile(path string) error {
	groups.ensureMaps()
	return saveTableWithOrderedItems(path, groups.MaxID, groups.Items)
}

// Добавляет группу оборудования в быстрые индексы.
func (groups *EquipmentGroups) ensureMaps() {
	groups.ensureItems()
	if groups.ByCosmicObjectID == nil {
		groups.ByCosmicObjectID = make(map[int64][]*EquipmentGroup)
	}
}

// Добавляет группу оборудования в быстрые индексы.
func (groups *EquipmentGroups) ensureItems() {
	if groups.Items == nil {
		groups.Items = make(map[int64]*EquipmentGroup)
	}
}

// Добавляет группу оборудования в быстрые индексы.
func (groups *EquipmentGroups) validateRequiredFields(group *EquipmentGroup) error {
	if group.CosmicObjectID <= 0 {
		return errors.New("cosmic object ID is empty")
	}
	if group.EquipmentItemModelID <= 0 {
		return errors.New("equipment item model ID is empty")
	}
	if group.Count < 0 {
		return errors.New("count is negative")
	}
	if group.EnabledCount < 0 {
		return errors.New("enabled count is negative")
	}
	if group.EnabledCount > group.Count {
		return errors.New("enabled count is greater than count")
	}
	return nil
}

// Добавляет группу оборудования в быстрые индексы.
func (groups *EquipmentGroups) addIndexes(group *EquipmentGroup) {
	groups.ByCosmicObjectID[group.CosmicObjectID] = append(groups.ByCosmicObjectID[group.CosmicObjectID], group)
}
