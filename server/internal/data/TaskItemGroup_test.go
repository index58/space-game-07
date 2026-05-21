package data

import "testing"

// Проверяет, что резерв задания уникален по паре задания и модели предмета.
func TestTaskItemGroupsRejectDuplicateTaskAndItemModel(t *testing.T) {
	groups := NewTaskItemGroups()

	if _, err := groups.Add(&TaskItemGroup{TaskID: 1, ItemModelID: 10, Count: 2}); err != nil {
		t.Fatalf("first Add returned error: %v", err)
	}
	if _, err := groups.Add(&TaskItemGroup{TaskID: 1, ItemModelID: 10, Count: 3}); err == nil {
		t.Fatal("Add accepted duplicate task item group")
	}
}

// Проверяет, что резерв можно найти по заданию после добавления.
func TestTaskItemGroupsIndexByTask(t *testing.T) {
	groups := NewTaskItemGroups()

	group, err := groups.Add(&TaskItemGroup{TaskID: 5, ItemModelID: 20, Count: 4})
	if err != nil {
		t.Fatalf("Add returned error: %v", err)
	}

	byTask := groups.GetByTaskID(5)
	if len(byTask) != 1 || byTask[0] != group {
		t.Fatal("task item group is not indexed by task")
	}
}
