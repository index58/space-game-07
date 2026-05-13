package data

import "testing"

func TestEquipmentGroupRelationsReplaceRelationForGroupAndType(t *testing.T) {
	// Проверяет, что для одной группы и одного типа связи хранится только последний выбранный контейнер.
	relations := NewEquipmentGroupRelations()
	if _, err := relations.Upsert(&EquipmentGroupRelation{EquipmentGroupID: 10, RelationTypeID: 2, RelatedEquipmentGroupID: 20}); err != nil {
		t.Fatalf("first upsert returned error: %v", err)
	}
	if _, err := relations.Upsert(&EquipmentGroupRelation{EquipmentGroupID: 10, RelationTypeID: 2, RelatedEquipmentGroupID: 30}); err != nil {
		t.Fatalf("second upsert returned error: %v", err)
	}

	relation, ok := relations.GetByEquipmentGroupAndType(10, 2)
	if !ok {
		t.Fatalf("relation was not found")
	}
	if relation.RelatedEquipmentGroupID != 30 {
		t.Fatalf("related equipment group = %d, want 30", relation.RelatedEquipmentGroupID)
	}
	if len(relations.Items) != 1 {
		t.Fatalf("relation count = %d, want 1", len(relations.Items))
	}
}

func TestRelationTypesFindByAcronym(t *testing.T) {
	// Проверяет поиск типа связи по неизменяемому строковому идентификатору.
	types := NewRelationTypes()
	if _, err := types.Add(&RelationType{TitleRu: "Источник", TitleEn: "Source", Acronym: "Source"}); err != nil {
		t.Fatalf("add returned error: %v", err)
	}

	relationType, ok := types.GetByAcronym("Source")
	if !ok {
		t.Fatalf("relation type was not found")
	}
	if relationType.TitleRu != "Источник" {
		t.Fatalf("title ru = %q, want Источник", relationType.TitleRu)
	}
}
