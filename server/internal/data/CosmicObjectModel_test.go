package data

import (
	"encoding/json"
	"math"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// РџСЂРѕРІРµСЂСЏРµС‚, С‡С‚Рѕ РґРѕР±Р°РІР»РµРЅРёРµ РјРѕРґРµР»Рё РЅР°Р·РЅР°С‡Р°РµС‚ РёРґРµРЅС‚РёС„РёРєР°С‚РѕСЂ, РІС‹С‡РёСЃР»СЏРµС‚ СЂР°Р·РјРµСЂС‹ С‚РµР»Р° Рё РёРЅРґРµРєСЃРёСЂСѓРµС‚ Р°РєСЂРѕРЅРёРј.
func TestCosmicObjectModelsAddAssignsIDCalculatesBodySizeAndIndexesModel(t *testing.T) {
	cosmicObjectModels := NewCosmicObjectModels()

	cosmicObjectModel, err := cosmicObjectModels.Add(&CosmicObjectModel{
		TitleRu:            "РђСЃС‚РµСЂРѕРёРґ 1",
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
	if cosmicObjectModel.BodyLength != 206.625 {
		t.Fatalf("BodyLength = %v, want 206.625", cosmicObjectModel.BodyLength)
	}
	if cosmicObjectModel.BodyWidth != 209.475 {
		t.Fatalf("BodyWidth = %v, want 209.475", cosmicObjectModel.BodyWidth)
	}

	byAcronym, ok := cosmicObjectModels.GetByAcronym("asteroid_0001")
	if !ok || byAcronym != cosmicObjectModel {
		t.Fatal("GetByAcronym did not return added model")
	}
}

// РџСЂРѕРІРµСЂСЏРµС‚, С‡С‚Рѕ С„РёР·РёС‡РµСЃРєРѕРµ С‚РµР»Рѕ РјРѕРґРµР»Рё РѕРїРёСЃС‹РІР°РµС‚СЃСЏ С€РµСЃС‚РЅР°РґС†Р°С‚РёС‚РѕС‡РµС‡РЅС‹Рј РјРЅРѕРіРѕСѓРіРѕР»СЊРЅРёРєРѕРј.
func TestCosmicObjectModelsBuildsSixteenPointBodyPolygon(t *testing.T) {
	cosmicObjectModels := NewCosmicObjectModels()

	cosmicObjectModel, err := cosmicObjectModels.Add(&CosmicObjectModel{
		TitleRu:            "Р С™Р С•РЎР‚Р В°Р В±Р В»РЎРЉ",
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
	if cosmicObjectModel.BodyPolygon[0].X != 0 || cosmicObjectModel.BodyPolygon[0].Y != 9.5 {
		t.Fatalf("BodyPolygon[0] = %+v, want X=0 Y=9.5", cosmicObjectModel.BodyPolygon[0])
	}
	if cosmicObjectModel.BodyPolygon[4].X != 4.75 || cosmicObjectModel.BodyPolygon[4].Y != 0 {
		t.Fatalf("BodyPolygon[4] = %+v, want X=4.75 Y=0", cosmicObjectModel.BodyPolygon[4])
	}
	if cosmicObjectModel.BodyPolygon[8].X != 0 || cosmicObjectModel.BodyPolygon[8].Y != -9.5 {
		t.Fatalf("BodyPolygon[8] = %+v, want X=0 Y=-9.5", cosmicObjectModel.BodyPolygon[8])
	}
	if cosmicObjectModel.BodyPolygon[12].X != -4.75 || cosmicObjectModel.BodyPolygon[12].Y != 0 {
		t.Fatalf("BodyPolygon[12] = %+v, want X=-4.75 Y=0", cosmicObjectModel.BodyPolygon[12])
	}
}

// РџСЂРѕРІРµСЂСЏРµС‚, С‡С‚Рѕ РјРЅРѕРіРѕСѓРіРѕР»СЊРЅРёРє С‚РµР»Р° СЃРјРµС‰Р°РµС‚СЃСЏ РѕС‚РЅРѕСЃРёС‚РµР»СЊРЅРѕ С†РµРЅС‚СЂР° С‚РµРєСЃС‚СѓСЂС‹.
func TestCosmicObjectModelsOffsetsBodyPolygonFromTextureBodyOrigin(t *testing.T) {
	cosmicObjectModels := NewCosmicObjectModels()

	cosmicObjectModel, err := cosmicObjectModels.Add(&CosmicObjectModel{
		TitleRu:            "Р С™Р С•РЎР‚Р В°Р В±Р В»РЎРЉ РЎРѓР С• РЎРѓР СР ВµРЎвЂ°Р ВµР Р…Р С‘Р ВµР С",
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

	if cosmicObjectModel.BodyPolygon[0].X != 1 || cosmicObjectModel.BodyPolygon[0].Y != 5.8 {
		t.Fatalf("BodyPolygon[0] = %+v, want X=1 Y=5.8", cosmicObjectModel.BodyPolygon[0])
	}
	if cosmicObjectModel.BodyPolygon[4].X != 2.9 || cosmicObjectModel.BodyPolygon[4].Y != 2 {
		t.Fatalf("BodyPolygon[4] = %+v, want X=2.9 Y=2", cosmicObjectModel.BodyPolygon[4])
	}
	if cosmicObjectModel.BodyPolygon[8].X != 1 || math.Abs(cosmicObjectModel.BodyPolygon[8].Y-(-1.8)) > 1e-9 {
		t.Fatalf("BodyPolygon[8] = %+v, want X=1 Y=-1.8", cosmicObjectModel.BodyPolygon[8])
	}
	if math.Abs(cosmicObjectModel.BodyPolygon[12].X-(-0.9)) > 1e-9 || cosmicObjectModel.BodyPolygon[12].Y != 2 {
		t.Fatalf("BodyPolygon[12] = %+v, want X=-0.9 Y=2", cosmicObjectModel.BodyPolygon[12])
	}
}

// РџСЂРѕРІРµСЂСЏРµС‚, С‡С‚Рѕ РїРѕРІС‚РѕСЂСЏСЋС‰РёРµСЃСЏ СѓРЅРёРєР°Р»СЊРЅС‹Рµ РЅР°Р·РІР°РЅРёСЏ Рё Р°РєСЂРѕРЅРёРј РјРѕРґРµР»Рё РЅРµ РґРѕРїСѓСЃРєР°СЋС‚СЃСЏ.
func TestCosmicObjectModelsAddRejectsDuplicateUniqueFields(t *testing.T) {
	cosmicObjectModels := NewCosmicObjectModels()

	if _, err := cosmicObjectModels.Add(&CosmicObjectModel{TitleRu: "РђСЃС‚РµСЂРѕРёРґ 1", TitleEn: "Asteroid 1", Acronym: "asteroid_0001", TextureScale: 4, CosmicObjectTypeID: 3}); err != nil {
		t.Fatalf("first Add returned error: %v", err)
	}

	if _, err := cosmicObjectModels.Add(&CosmicObjectModel{TitleRu: "РђСЃС‚РµСЂРѕРёРґ 1", TitleEn: "Asteroid 2", Acronym: "asteroid_0002", TextureScale: 4, CosmicObjectTypeID: 3}); err == nil {
		t.Fatal("Add accepted duplicate TitleRu")
	}
	if _, err := cosmicObjectModels.Add(&CosmicObjectModel{TitleRu: "РђСЃС‚РµСЂРѕРёРґ 2", TitleEn: "Asteroid 1", Acronym: "asteroid_0002", TextureScale: 4, CosmicObjectTypeID: 3}); err == nil {
		t.Fatal("Add accepted duplicate TitleEn")
	}
	if _, err := cosmicObjectModels.Add(&CosmicObjectModel{TitleRu: "РђСЃС‚РµСЂРѕРёРґ 2", TitleEn: "Asteroid 2", Acronym: "asteroid_0001", TextureScale: 4, CosmicObjectTypeID: 3}); err == nil {
		t.Fatal("Add accepted duplicate Acronym")
	}
}

// РџСЂРѕРІРµСЂСЏРµС‚, С‡С‚Рѕ СЃРѕС…СЂР°РЅС‘РЅРЅС‹Рµ РјРѕРґРµР»Рё Р·Р°РіСЂСѓР¶Р°СЋС‚СЃСЏ РѕР±СЂР°С‚РЅРѕ СЃ РІРѕСЃСЃС‚Р°РЅРѕРІР»РµРЅРЅС‹РјРё РѕСЃРЅРѕРІРЅС‹РјРё РїРѕР»СЏРјРё.
func TestCosmicObjectModelsSaveLoadAndRebuildIndexes(t *testing.T) {
	path := filepath.Join(t.TempDir(), "CosmicObjectModels.json")
	cosmicObjectModels := NewCosmicObjectModels()
	cosmicObjectModel, err := cosmicObjectModels.Add(&CosmicObjectModel{
		TitleRu:            "РљРѕСЂР°Р±Р»СЊ",
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

// РџСЂРѕРІРµСЂСЏРµС‚, С‡С‚Рѕ JSON-РїСЂРµРґСЃС‚Р°РІР»РµРЅРёРµ РјРѕРґРµР»РµР№ РёСЃРїРѕР»СЊР·СѓРµС‚ РёРјРµРЅР° РїРѕР»РµР№ РёР· Go-СЃС‚СЂСѓРєС‚СѓСЂ.
func TestCosmicObjectModelsJSONKeysMatchGoFieldNames(t *testing.T) {
	cosmicObjectModels := NewCosmicObjectModels()
	if _, err := cosmicObjectModels.Add(&CosmicObjectModel{TitleRu: "РљРѕСЂР°Р±Р»СЊ", TitleEn: "Ship", Acronym: "ship_0001", TextureScale: 4, CosmicObjectTypeID: 1}); err != nil {
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

// РџСЂРѕРІРµСЂСЏРµС‚, С‡С‚Рѕ РёРјРїРѕСЂС‚ СЃС‚Р°СЂРѕРіРѕ JSON РґРѕР±Р°РІР»СЏРµС‚ РЅРѕРјРµСЂР° Рє РїРѕРІС‚РѕСЂСЏСЋС‰РёРјСЃСЏ РЅР°Р·РІР°РЅРёСЏРј.
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
    "TitleRu": "РђСЃС‚РµСЂРѕРёРґ",
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
    "TitleRu": "РђСЃС‚РµСЂРѕРёРґ",
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
	if _, err := cosmicObjectTypes.Add(&CosmicObjectType{TitleRu: "РђСЃС‚РµСЂРѕРёРґ", TitleEn: "Asteroid", Acronym: "Asteroid"}); err != nil {
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
	if first.TitleRu != "РђСЃС‚РµСЂРѕРёРґ 1" || second.TitleRu != "РђСЃС‚РµСЂРѕРёРґ 2" {
		t.Fatalf("TitleRu values = %q, %q; want РђСЃС‚РµСЂРѕРёРґ 1, РђСЃС‚РµСЂРѕРёРґ 2", first.TitleRu, second.TitleRu)
	}
	if first.TitleEn != "Asteroid 1" || second.TitleEn != "Asteroid 2" {
		t.Fatalf("TitleEn values = %q, %q; want Asteroid 1, Asteroid 2", first.TitleEn, second.TitleEn)
	}
}
