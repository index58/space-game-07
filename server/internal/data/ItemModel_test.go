package data

import (
	"encoding/json"
	"reflect"
	"testing"
)

// Проверяет, что скорость добычи из JSON доступна в модели предмета.
func TestItemModelLoadsMiningSpeed(t *testing.T) {
	raw := []byte(`{"ID":1,"Acronym":"SimpleDrill","MiningSpeed":20}`)

	var model ItemModel
	if err := json.Unmarshal(raw, &model); err != nil {
		t.Fatalf("unmarshal item model: %v", err)
	}

	field := reflect.ValueOf(model).FieldByName("MiningSpeed")
	if !field.IsValid() {
		t.Fatalf("mining speed field is missing")
	}
	if field.Float() != 20 {
		t.Fatalf("mining speed = %v, want 20", field.Float())
	}
}
