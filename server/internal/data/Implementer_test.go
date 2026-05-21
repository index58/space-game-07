package data

import "testing"

// Проверяет, что исполнитель задания уникален по типу задания и типу оборудования.
func TestImplementersRejectDuplicateTaskTypeAndEquipmentType(t *testing.T) {
	implementers := NewImplementers()

	if _, err := implementers.Add(&Implementer{TaskTypeID: 1, ImplementerEquipmentItemTypeID: 10, WorkPart: 0.5}); err != nil {
		t.Fatalf("first Add returned error: %v", err)
	}
	if _, err := implementers.Add(&Implementer{TaskTypeID: 1, ImplementerEquipmentItemTypeID: 10, WorkPart: 1}); err == nil {
		t.Fatal("Add accepted duplicate implementer")
	}
}

// Проверяет, что исполнителей можно получить по типу задания.
func TestImplementersIndexByTaskType(t *testing.T) {
	implementers := NewImplementers()

	implementer, err := implementers.Add(&Implementer{TaskTypeID: 2, ImplementerEquipmentItemTypeID: 20, WorkPart: 1})
	if err != nil {
		t.Fatalf("Add returned error: %v", err)
	}

	byTaskType := implementers.GetByTaskTypeID(2)
	if len(byTaskType) != 1 || byTaskType[0] != implementer {
		t.Fatal("implementer is not indexed by task type")
	}
}
