package data

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"sort"
)

// Task хранит одно задание в очереди группы оборудования.
type Task struct {
	ID                              int64   `json:"ID"`                              // Уникальный числовой идентификатор записи.
	ControllerEquipmentGroupID      int64   `json:"ControllerEquipmentGroupID"`      // Группа оборудования, в очереди которой находится работа.
	ParentTaskID                    int64   `json:"ParentTaskID"`                    // Родительская работа для вспомогательной строки.
	TaskTypeID                      int64   `json:"TaskTypeID"`                      // Тип выполняемой работы.
	RemainingEnergy                 float64 `json:"RemainingEnergy"`                 // Остаток работы в джоулях.
	TotalEnergy                     float64 `json:"TotalEnergy"`                     // Полный объем работы в джоулях.
	Count                           float64 `json:"Count"`                           // Количество единиц результата, которое должно выполнить задание.
	SchemaID                        int64   `json:"SchemaID"`                        // Схема для изготовления предмета.
	BlueprintID                     int64   `json:"BlueprintID"`                     // Чертеж для изготовления объекта.
	LeftToRightDirection            bool    `json:"LeftToRightDirection"`            // Направление работы слева направо для заданий с парным интерфейсом.
	SourceContainerEquipmentGroupID int64   `json:"SourceContainerEquipmentGroupID"` // Контейнер, из которого предметы были зарезервированы.
	TargetContainerEquipmentGroupID int64   `json:"TargetContainerEquipmentGroupID"` // Контейнер, куда нужно положить результат перемещения.
	FuelTankEquipmentGroupID        int64   `json:"FuelTankEquipmentGroupID"`        // Бак, участвующий в заправке или сливе.
}

// Tasks хранит задания и быстрые индексы по очереди оборудования.
type Tasks struct {
	MaxID int64           `json:"MaxID"` // Последний выданный числовой идентификатор записей.
	Items map[int64]*Task `json:"Items"` // Основное хранилище записей по числовому идентификатору.

	ByControllerEquipmentGroupID map[int64][]*Task `json:"-"` // Быстрый поиск заданий по контроллеру очереди.
}

// NewTasks создает пустое хранилище заданий.
func NewTasks() *Tasks {
	tasks := &Tasks{}
	tasks.ensureMaps()
	return tasks
}

// Add добавляет задание и назначает новый ID.
func (tasks *Tasks) Add(task *Task) (*Task, error) {
	if task == nil {
		return nil, errors.New("task is nil")
	}
	tasks.ensureMaps()
	if err := tasks.validateRequiredFields(task); err != nil {
		return nil, err
	}

	tasks.MaxID++
	task.ID = tasks.MaxID
	tasks.Items[task.ID] = task
	tasks.addIndexes(task)
	return task, nil
}

// Get возвращает задание по ID.
func (tasks *Tasks) Get(id int64) (*Task, bool) {
	tasks.ensureMaps()
	task, ok := tasks.Items[id]
	return task, ok
}

// GetByControllerEquipmentGroupID возвращает задания в очереди указанной группы.
func (tasks *Tasks) GetByControllerEquipmentGroupID(controllerEquipmentGroupID int64) []*Task {
	tasks.ensureMaps()
	return tasks.ByControllerEquipmentGroupID[controllerEquipmentGroupID]
}

// Delete удаляет задание из хранилища.
func (tasks *Tasks) Delete(id int64) bool {
	tasks.ensureMaps()
	if _, ok := tasks.Items[id]; !ok {
		return false
	}
	delete(tasks.Items, id)
	tasks.RebuildIndexes()
	return true
}

// RebuildIndexes пересобирает индексы после загрузки из JSON.
func (tasks *Tasks) RebuildIndexes() error {
	tasks.ensureItems()
	tasks.ByControllerEquipmentGroupID = make(map[int64][]*Task)

	var maxID int64
	ids := make([]int64, 0, len(tasks.Items))
	for id := range tasks.Items {
		ids = append(ids, id)
	}
	sort.Slice(ids, func(left int, right int) bool { return ids[left] < ids[right] })
	for _, id := range ids {
		task := tasks.Items[id]
		if task == nil {
			return fmt.Errorf("task with ID %d is nil", id)
		}
		if task.ID != id {
			return fmt.Errorf("task map key %d does not match task ID %d", id, task.ID)
		}
		if err := tasks.validateRequiredFields(task); err != nil {
			return fmt.Errorf("task with ID %d is invalid: %w", id, err)
		}
		if id > maxID {
			maxID = id
		}
		tasks.addIndexes(task)
	}
	if tasks.MaxID < maxID {
		tasks.MaxID = maxID
	}
	return nil
}

// LoadFromFile загружает задания из JSON-файла.
func (tasks *Tasks) LoadFromFile(path string) error {
	content, err := os.ReadFile(path)
	if err != nil {
		return err
	}

	loaded := Tasks{}
	if err := json.Unmarshal(content, &loaded); err != nil {
		return err
	}
	if err := loaded.RebuildIndexes(); err != nil {
		return err
	}

	*tasks = loaded
	return nil
}

// SaveToFile сохраняет задания в JSON-файл.
func (tasks *Tasks) SaveToFile(path string) error {
	tasks.ensureMaps()
	return saveTableWithOrderedItems(path, tasks.MaxID, tasks.Items)
}

// ensureMaps подготавливает хранилище и индексы.
func (tasks *Tasks) ensureMaps() {
	tasks.ensureItems()
	if tasks.ByControllerEquipmentGroupID == nil {
		tasks.ByControllerEquipmentGroupID = make(map[int64][]*Task)
	}
}

// ensureItems подготавливает основное хранилище.
func (tasks *Tasks) ensureItems() {
	if tasks.Items == nil {
		tasks.Items = make(map[int64]*Task)
	}
}

// validateRequiredFields проверяет обязательные поля задания.
func (tasks *Tasks) validateRequiredFields(task *Task) error {
	if task.ControllerEquipmentGroupID <= 0 {
		return errors.New("controller equipment group ID is empty")
	}
	if task.TaskTypeID <= 0 {
		return errors.New("task type ID is empty")
	}
	if task.RemainingEnergy < 0 {
		return errors.New("remaining energy is negative")
	}
	if task.TotalEnergy < 0 {
		return errors.New("total energy is negative")
	}
	return nil
}

// addIndexes добавляет задание в быстрые индексы.
func (tasks *Tasks) addIndexes(task *Task) {
	tasks.ByControllerEquipmentGroupID[task.ControllerEquipmentGroupID] = append(tasks.ByControllerEquipmentGroupID[task.ControllerEquipmentGroupID], task)
}
