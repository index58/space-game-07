package data

import "testing"

// Проверяет, что добавление задания назначает идентификатор и строит индекс по группе-контроллеру.
func TestTasksAddAssignsIDAndIndexesByController(t *testing.T) {
	tasks := NewTasks()

	task, err := tasks.Add(&Task{ControllerEquipmentGroupID: 10, TaskTypeID: 1, RemainingEnergy: 50, TotalEnergy: 100})
	if err != nil {
		t.Fatalf("Add returned error: %v", err)
	}

	if task.ID != 1 {
		t.Fatalf("task ID = %d, want 1", task.ID)
	}
	if byController := tasks.GetByControllerEquipmentGroupID(10); len(byController) != 1 || byController[0] != task {
		t.Fatal("task is not indexed by controller equipment group")
	}
}

// Проверяет, что восстановление индексов отклоняет задание без обязательных ссылок.
func TestTasksRebuildIndexesRejectsInvalidStoredTask(t *testing.T) {
	tasks := NewTasks()
	tasks.Items[1] = &Task{ID: 1}

	if err := tasks.RebuildIndexes(); err == nil {
		t.Fatal("RebuildIndexes accepted task without required fields")
	}
}
