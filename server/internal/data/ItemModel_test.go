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

// Проверяет, что оружие хранит ссылку на модель объекта-снаряда.
func TestItemModelLoadsProjectileObjectModelID(t *testing.T) {
	raw := []byte(`{"ID":1,"Acronym":"LightCannon120mm","ProjectileObjectModelID":546}`)

	var model ItemModel
	if err := json.Unmarshal(raw, &model); err != nil {
		t.Fatalf("unmarshal item model: %v", err)
	}

	field := reflect.ValueOf(model).FieldByName("ProjectileObjectModelID")
	if !field.IsValid() {
		t.Fatalf("projectile object model ID field is missing")
	}
	if field.Int() != 546 {
		t.Fatalf("projectile object model ID = %v, want 546", field.Int())
	}
}

// Проверяет, что параметры урона и скорости больше не хранятся в модели предмета.
func TestItemModelDoesNotContainProjectileDamageAndSpeed(t *testing.T) {
	modelType := reflect.TypeOf(ItemModel{})

	if _, ok := modelType.FieldByName("Damage"); ok {
		t.Fatal("item model still contains Damage")
	}
	if _, ok := modelType.FieldByName("ProjectileSpeed"); ok {
		t.Fatal("item model still contains ProjectileSpeed")
	}
}
