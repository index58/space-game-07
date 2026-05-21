package data

import "testing"

// Проверяет, что тип задания можно найти по акрониму после добавления.
func TestTaskTypesAddIndexesByAcronym(t *testing.T) {
	types := NewTaskTypes()

	taskType, err := types.Add(&TaskType{TitleRu: "Производство предмета", TitleEn: "Item production", Acronym: "ItemProduction"})
	if err != nil {
		t.Fatalf("Add returned error: %v", err)
	}

	byAcronym, ok := types.GetByAcronym("ItemProduction")
	if !ok || byAcronym != taskType {
		t.Fatal("task type is not indexed by acronym")
	}
}

// Проверяет, что одинаковый акроним типа задания не допускается.
func TestTaskTypesRejectDuplicateAcronym(t *testing.T) {
	types := NewTaskTypes()

	if _, err := types.Add(&TaskType{TitleRu: "Производство предмета", TitleEn: "Item production", Acronym: "ItemProduction"}); err != nil {
		t.Fatalf("first Add returned error: %v", err)
	}
	if _, err := types.Add(&TaskType{TitleRu: "Другое", TitleEn: "Other", Acronym: "ItemProduction"}); err == nil {
		t.Fatal("Add accepted duplicate acronym")
	}
}
