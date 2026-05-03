package data

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// Проверяет, что добавление модели назначает идентификатор, вычисляет размеры тела и индексирует акроним.
func TestCosmicObjectModelsAddAssignsIDCalculatesBodySizeAndIndexesModel(t *testing.T) {
	cosmicObjectModels := NewCosmicObjectModels()

	cosmicObjectModel, err := cosmicObjectModels.Add(&CosmicObjectModel{
		TitleRu:            "Астероид 1",
		TitleEn:            "Asteroid 1",
		Acronym:            "asteroid_0001",
		TextureFilePath:    "assets/asteroids/asteroid_0001.png",
		TextureWidth:       2048,
		TextureHeight:      2048,
		TextureBodyOriginX: 1055,
		TextureBodyOriginY: 1107,
		TextureBodyWidth:   882,
		TextureBodyLength:  870,
		TextureScale:       4,
		CosmicObjectTypeID: 3,
		Mass:               767.34,
		MaxSpeed:           472,
		MaxAngularSpeed:    3,
	})
	if err != nil {
		t.Fatalf("Add returned error: %v", err)
	}

	if cosmicObjectModel.ID != 1 {
		t.Fatalf("model ID = %d, want 1", cosmicObjectModel.ID)
	}
	if cosmicObjectModel.BodyLength != 217.5 {
		t.Fatalf("BodyLength = %v, want 217.5", cosmicObjectModel.BodyLength)
	}
	if cosmicObjectModel.BodyWidth != 220.5 {
		t.Fatalf("BodyWidth = %v, want 220.5", cosmicObjectModel.BodyWidth)
	}

	byAcronym, ok := cosmicObjectModels.GetByAcronym("asteroid_0001")
	if !ok || byAcronym != cosmicObjectModel {
		t.Fatal("GetByAcronym did not return added model")
	}
}

// Проверяет, что физическое тело модели описывается шестнадцатиточечным многоугольником.
func TestCosmicObjectModelsBuildsSixteenPointBodyPolygon(t *testing.T) {
	cosmicObjectModels := NewCosmicObjectModels()

	cosmicObjectModel, err := cosmicObjectModels.Add(&CosmicObjectModel{
		TitleRu:            "РљРѕСЂР°Р±Р»СЊ",
		TitleEn:            "Ship",
		Acronym:            "ship_test",
		TextureScale:       4,
		TextureBodyWidth:   40,
		TextureBodyLength:  80,
		CosmicObjectTypeID: 1,
	})
	if err != nil {
		t.Fatalf("Add returned error: %v", err)
	}

	if len(cosmicObjectModel.BodyPolygon) != 16 {
		t.Fatalf("BodyPolygon vertex count = %d, want 16", len(cosmicObjectModel.BodyPolygon))
	}
	if cosmicObjectModel.BodyPolygon[0].X != 0 || cosmicObjectModel.BodyPolygon[0].Y != 10 {
		t.Fatalf("BodyPolygon[0] = %+v, want X=0 Y=10", cosmicObjectModel.BodyPolygon[0])
	}
	if cosmicObjectModel.BodyPolygon[4].X != 5 || cosmicObjectModel.BodyPolygon[4].Y != 0 {
		t.Fatalf("BodyPolygon[4] = %+v, want X=5 Y=0", cosmicObjectModel.BodyPolygon[4])
	}
	if cosmicObjectModel.BodyPolygon[8].X != 0 || cosmicObjectModel.BodyPolygon[8].Y != -10 {
		t.Fatalf("BodyPolygon[8] = %+v, want X=0 Y=-10", cosmicObjectModel.BodyPolygon[8])
	}
	if cosmicObjectModel.BodyPolygon[12].X != -5 || cosmicObjectModel.BodyPolygon[12].Y != 0 {
		t.Fatalf("BodyPolygon[12] = %+v, want X=-5 Y=0", cosmicObjectModel.BodyPolygon[12])
	}
}

// Проверяет, что многоугольник тела смещается относительно центра текстуры.
func TestCosmicObjectModelsOffsetsBodyPolygonFromTextureBodyOrigin(t *testing.T) {
	cosmicObjectModels := NewCosmicObjectModels()

	cosmicObjectModel, err := cosmicObjectModels.Add(&CosmicObjectModel{
		TitleRu:            "РљРѕСЂР°Р±Р»СЊ СЃРѕ СЃРјРµС‰РµРЅРёРµРј",
		TitleEn:            "Offset Ship",
		Acronym:            "ship_offset",
		TextureWidth:       100,
		TextureHeight:      200,
		TextureBodyOriginX: 60,
		TextureBodyOriginY: 120,
		TextureScale:       10,
		TextureBodyWidth:   40,
		TextureBodyLength:  80,
		CosmicObjectTypeID: 1,
	})
	if err != nil {
		t.Fatalf("Add returned error: %v", err)
	}

	if cosmicObjectModel.BodyPolygon[0].X != 1 || cosmicObjectModel.BodyPolygon[0].Y != 6 {
		t.Fatalf("BodyPolygon[0] = %+v, want X=1 Y=6", cosmicObjectModel.BodyPolygon[0])
	}
	if cosmicObjectModel.BodyPolygon[4].X != 3 || cosmicObjectModel.BodyPolygon[4].Y != 2 {
		t.Fatalf("BodyPolygon[4] = %+v, want X=3 Y=2", cosmicObjectModel.BodyPolygon[4])
	}
	if cosmicObjectModel.BodyPolygon[8].X != 1 || cosmicObjectModel.BodyPolygon[8].Y != -2 {
		t.Fatalf("BodyPolygon[8] = %+v, want X=1 Y=-2", cosmicObjectModel.BodyPolygon[8])
	}
	if cosmicObjectModel.BodyPolygon[12].X != -1 || cosmicObjectModel.BodyPolygon[12].Y != 2 {
		t.Fatalf("BodyPolygon[12] = %+v, want X=-1 Y=2", cosmicObjectModel.BodyPolygon[12])
	}
}

// Проверяет, что повторяющиеся уникальные названия и акроним модели не допускаются.
func TestCosmicObjectModelsAddRejectsDuplicateUniqueFields(t *testing.T) {
	cosmicObjectModels := NewCosmicObjectModels()

	if _, err := cosmicObjectModels.Add(&CosmicObjectModel{TitleRu: "Астероид 1", TitleEn: "Asteroid 1", Acronym: "asteroid_0001", TextureScale: 4, CosmicObjectTypeID: 3}); err != nil {
		t.Fatalf("first Add returned error: %v", err)
	}

	if _, err := cosmicObjectModels.Add(&CosmicObjectModel{TitleRu: "Астероид 1", TitleEn: "Asteroid 2", Acronym: "asteroid_0002", TextureScale: 4, CosmicObjectTypeID: 3}); err == nil {
		t.Fatal("Add accepted duplicate TitleRu")
	}
	if _, err := cosmicObjectModels.Add(&CosmicObjectModel{TitleRu: "Астероид 2", TitleEn: "Asteroid 1", Acronym: "asteroid_0002", TextureScale: 4, CosmicObjectTypeID: 3}); err == nil {
		t.Fatal("Add accepted duplicate TitleEn")
	}
	if _, err := cosmicObjectModels.Add(&CosmicObjectModel{TitleRu: "Астероид 2", TitleEn: "Asteroid 2", Acronym: "asteroid_0001", TextureScale: 4, CosmicObjectTypeID: 3}); err == nil {
		t.Fatal("Add accepted duplicate Acronym")
	}
}

// Проверяет, что сохранённые модели загружаются обратно с восстановленными основными полями.
func TestCosmicObjectModelsSaveLoadAndRebuildIndexes(t *testing.T) {
	path := filepath.Join(t.TempDir(), "CosmicObjectModels.json")
	cosmicObjectModels := NewCosmicObjectModels()
	cosmicObjectModel, err := cosmicObjectModels.Add(&CosmicObjectModel{
		TitleRu:            "Корабль",
		TitleEn:            "Ship",
		Acronym:            "ship_0001",
		TextureBodyWidth:   88,
		TextureBodyLength:  90,
		TextureScale:       4,
		CosmicObjectTypeID: 1,
	})
	if err != nil {
		t.Fatalf("Add returned error: %v", err)
	}

	if err := cosmicObjectModels.SaveToFile(path); err != nil {
		t.Fatalf("SaveToFile returned error: %v", err)
	}
	if _, err := os.Stat(path); err != nil {
		t.Fatalf("saved file is not available: %v", err)
	}

	loaded := NewCosmicObjectModels()
	if err := loaded.LoadFromFile(path); err != nil {
		t.Fatalf("LoadFromFile returned error: %v", err)
	}

	loadedModel, ok := loaded.Get(cosmicObjectModel.ID)
	if !ok {
		t.Fatal("loaded model is not available by ID")
	}
	if loadedModel.TitleRu != cosmicObjectModel.TitleRu || loadedModel.TitleEn != cosmicObjectModel.TitleEn || loadedModel.Acronym != cosmicObjectModel.Acronym {
		t.Fatal("loaded model fields do not match saved model")
	}
}

// Проверяет, что JSON-представление моделей использует имена полей из Go-структур.
func TestCosmicObjectModelsJSONKeysMatchGoFieldNames(t *testing.T) {
	cosmicObjectModels := NewCosmicObjectModels()
	if _, err := cosmicObjectModels.Add(&CosmicObjectModel{TitleRu: "Корабль", TitleEn: "Ship", Acronym: "ship_0001", TextureScale: 4, CosmicObjectTypeID: 1}); err != nil {
		t.Fatalf("Add returned error: %v", err)
	}

	content, err := json.Marshal(cosmicObjectModels)
	if err != nil {
		t.Fatalf("Marshal returned error: %v", err)
	}
	text := string(content)

	expectedKeys := []string{
		`"MaxID"`,
		`"Items"`,
		`"ID"`,
		`"TitleRu"`,
		`"TitleEn"`,
		`"Acronym"`,
		`"IconFilePath"`,
		`"TextureFilePath"`,
		`"TextureBodyOriginX"`,
		`"TextureScale"`,
		`"CosmicObjectTypeID"`,
		`"BodyLength"`,
		`"BodyWidth"`,
	}
	for _, expectedKey := range expectedKeys {
		if !strings.Contains(text, expectedKey) {
			t.Fatalf("JSON %s does not contain key %s", text, expectedKey)
		}
	}
}

// Проверяет, что импорт старого JSON добавляет номера к повторяющимся названиям.
func TestCosmicObjectModelsLoadFromLegacyJSONNumbersDuplicateTitles(t *testing.T) {
	path := filepath.Join(t.TempDir(), "legacy.json")
	content := []byte(`[
  {
    "TextureFilePath": "assets/asteroids/asteroid_0001.png",
    "TextureWidth": 2048,
    "TextureHeight": 2048,
    "TextureObjectOriginX": 1055,
    "TextureObjectOriginY": 1107,
    "TextureObjectWidth": 882,
    "TextureObjectLength": 870,
    "CosmicObjectType": "asteroid",
    "TitleRu": "Астероид",
    "TitleEn": "Asteroid",
    "Acronym": "asteroid_0001",
    "Mass": 767.34,
    "MaxSpeed": 472,
    "MaxAngularSpeed": 3
  },
  {
    "TextureFilePath": "assets/asteroids/asteroid_0002.png",
    "TextureWidth": 2048,
    "TextureHeight": 2048,
    "TextureObjectOriginX": 988,
    "TextureObjectOriginY": 1289,
    "TextureObjectWidth": 804,
    "TextureObjectLength": 783,
    "CosmicObjectType": "asteroid",
    "TitleRu": "Астероид",
    "TitleEn": "Asteroid",
    "Acronym": "asteroid_0002",
    "Mass": 629.532,
    "MaxSpeed": 475,
    "MaxAngularSpeed": 3
  }
]`)
	if err := os.WriteFile(path, content, 0o600); err != nil {
		t.Fatalf("WriteFile returned error: %v", err)
	}

	cosmicObjectTypes := NewCosmicObjectTypes()
	if _, err := cosmicObjectTypes.Add(&CosmicObjectType{TitleRu: "Астероид", TitleEn: "Asteroid", Acronym: "Asteroid"}); err != nil {
		t.Fatalf("Add returned error: %v", err)
	}

	cosmicObjectModels, err := LoadCosmicObjectModelsFromLegacyFile(path, cosmicObjectTypes)
	if err != nil {
		t.Fatalf("LoadCosmicObjectModelsFromLegacyFile returned error: %v", err)
	}

	first, ok := cosmicObjectModels.GetByAcronym("asteroid_0001")
	if !ok {
		t.Fatal("first model is not available by acronym")
	}
	second, ok := cosmicObjectModels.GetByAcronym("asteroid_0002")
	if !ok {
		t.Fatal("second model is not available by acronym")
	}
	if first.TitleRu != "Астероид 1" || second.TitleRu != "Астероид 2" {
		t.Fatalf("TitleRu values = %q, %q; want Астероид 1, Астероид 2", first.TitleRu, second.TitleRu)
	}
	if first.TitleEn != "Asteroid 1" || second.TitleEn != "Asteroid 2" {
		t.Fatalf("TitleEn values = %q, %q; want Asteroid 1, Asteroid 2", first.TitleEn, second.TitleEn)
	}
}
