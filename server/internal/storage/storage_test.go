package storage

import (
	"os"
	"path/filepath"
	"testing"
)

// РџСЂРѕРІРµСЂСЏРµС‚, С‡С‚Рѕ РѕСЃРЅРѕРІРЅРѕР№ Р·Р°РіСЂСѓР·С‡РёРє С‡РёС‚Р°РµС‚ Р°РєРєР°СѓРЅС‚С‹ РёР· СЃС‚Р°РЅРґР°СЂС‚РЅРѕРіРѕ С„Р°Р№Р»Р° Рё СЃС‚СЂРѕРёС‚ РїРѕРёСЃРє РїРѕ РїРѕС‡С‚Рµ.
func TestLoadServerDataLoadsAccountsFromDefaultFile(t *testing.T) {
	workingDirectory := t.TempDir()
	dataDirectory := filepath.Join(workingDirectory, "data")
	if err := os.Mkdir(dataDirectory, 0o700); err != nil {
		t.Fatalf("Mkdir returned error: %v", err)
	}

	content := []byte(`{
  "MaxID": 1,
  "Items": {
    "1": {
      "ID": 1,
      "Email": "index@email.net",
      "Nickname": "index",
      "PasswordHash": "hash",
      "Token": "token",
      "RegistrationTime": "2026-04-30T18:13:48.8712091+03:00",
      "CurrentCharacterID": 0
    }
  }
}`)
	if err := os.WriteFile(filepath.Join(dataDirectory, "Accounts.json"), content, 0o600); err != nil {
		t.Fatalf("WriteFile returned error: %v", err)
	}
	if err := os.WriteFile(filepath.Join(dataDirectory, "Characters.json"), []byte(`{"MaxID":0,"Items":{}}`), 0o600); err != nil {
		t.Fatalf("WriteFile returned error: %v", err)
	}
	if err := os.WriteFile(filepath.Join(dataDirectory, "CosmicObjects.json"), []byte(`{"MaxID":0,"Items":{}}`), 0o600); err != nil {
		t.Fatalf("WriteFile returned error: %v", err)
	}
	if err := os.WriteFile(filepath.Join(dataDirectory, "CosmicObjectTypes.json"), []byte(`{"MaxID":0,"Items":{}}`), 0o600); err != nil {
		t.Fatalf("WriteFile returned error: %v", err)
	}
	if err := os.WriteFile(filepath.Join(dataDirectory, "CosmicObjectModels.json"), []byte(`{"MaxID":0,"Items":{}}`), 0o600); err != nil {
		t.Fatalf("WriteFile returned error: %v", err)
	}
	if err := os.WriteFile(filepath.Join(dataDirectory, "ItemTypes.json"), []byte(`{"MaxID":0,"Items":{}}`), 0o600); err != nil {
		t.Fatalf("WriteFile returned error: %v", err)
	}

	serverData, err := LoadServerData(workingDirectory)
	if err != nil {
		t.Fatalf("LoadServerData returned error: %v", err)
	}

	account, ok := serverData.Accounts.GetByEmail("index@email.net")
	if !ok {
		t.Fatal("account is not available by email")
	}
	if account.ID != 1 {
		t.Fatalf("account ID = %d, want 1", account.ID)
	}
}

// РџСЂРѕРІРµСЂСЏРµС‚, С‡С‚Рѕ РѕС‚СЃСѓС‚СЃС‚РІРёРµ РѕР±СЏР·Р°С‚РµР»СЊРЅРѕРіРѕ С„Р°Р№Р»Р° Р°РєРєР°СѓРЅС‚РѕРІ СЃС‡РёС‚Р°РµС‚СЃСЏ РѕС€РёР±РєРѕР№ Р·Р°РіСЂСѓР·РєРё.
func TestLoadServerDataReturnsErrorWhenAccountsFileIsMissing(t *testing.T) {
	_, err := LoadServerData(t.TempDir())
	if err == nil {
		t.Fatal("LoadServerData accepted missing Accounts.json")
	}
}

// РџСЂРѕРІРµСЂСЏРµС‚, С‡С‚Рѕ РїРµСЂСЃРѕРЅР°Р¶Рё С‡РёС‚Р°СЋС‚СЃСЏ РёР· СЃС‚Р°РЅРґР°СЂС‚РЅРѕРіРѕ С„Р°Р№Р»Р° Рё РґРѕСЃС‚СѓРїРЅС‹ РїРѕ С‡РёСЃР»РѕРІРѕРјСѓ РёРґРµРЅС‚РёС„РёРєР°С‚РѕСЂСѓ.
func TestLoadServerDataLoadsCharactersFromDefaultFile(t *testing.T) {
	workingDirectory := t.TempDir()
	dataDirectory := filepath.Join(workingDirectory, "data")
	if err := os.Mkdir(dataDirectory, 0o700); err != nil {
		t.Fatalf("Mkdir returned error: %v", err)
	}
	if err := os.WriteFile(filepath.Join(dataDirectory, "Accounts.json"), []byte(`{"MaxID":0,"Items":{}}`), 0o600); err != nil {
		t.Fatalf("WriteFile returned error: %v", err)
	}
	if err := os.WriteFile(filepath.Join(dataDirectory, "CosmicObjectTypes.json"), []byte(`{"MaxID":0,"Items":{}}`), 0o600); err != nil {
		t.Fatalf("WriteFile returned error: %v", err)
	}
	if err := os.WriteFile(filepath.Join(dataDirectory, "CosmicObjectModels.json"), []byte(`{"MaxID":0,"Items":{}}`), 0o600); err != nil {
		t.Fatalf("WriteFile returned error: %v", err)
	}
	if err := os.WriteFile(filepath.Join(dataDirectory, "ItemTypes.json"), []byte(`{"MaxID":0,"Items":{}}`), 0o600); err != nil {
		t.Fatalf("WriteFile returned error: %v", err)
	}
	if err := os.WriteFile(filepath.Join(dataDirectory, "CosmicObjectModels.json"), []byte(`{"MaxID":0,"Items":{}}`), 0o600); err != nil {
		t.Fatalf("WriteFile returned error: %v", err)
	}

	content := []byte(`{
  "MaxID": 1,
  "Items": {
    "1": {
      "ID": 1,
      "AccountID": 7,
      "CreationTime": "2026-04-30T18:13:48.8712091+03:00",
      "Balance": 100,
      "LocationCosmicObjectID": 20
    }
  }
}`)
	if err := os.WriteFile(filepath.Join(dataDirectory, "Characters.json"), content, 0o600); err != nil {
		t.Fatalf("WriteFile returned error: %v", err)
	}
	if err := os.WriteFile(filepath.Join(dataDirectory, "CosmicObjects.json"), []byte(`{"MaxID":0,"Items":{}}`), 0o600); err != nil {
		t.Fatalf("WriteFile returned error: %v", err)
	}

	serverData, err := LoadServerData(workingDirectory)
	if err != nil {
		t.Fatalf("LoadServerData returned error: %v", err)
	}

	character, ok := serverData.Characters.Get(1)
	if !ok {
		t.Fatal("character is not available by ID")
	}
	if character.AccountID != 7 {
		t.Fatalf("character AccountID = %d, want 7", character.AccountID)
	}
}

// РџСЂРѕРІРµСЂСЏРµС‚, С‡С‚Рѕ С‚РёРїС‹ РєРѕСЃРјРёС‡РµСЃРєРёС… РѕР±СЉРµРєС‚РѕРІ С‡РёС‚Р°СЋС‚СЃСЏ РёР· СЃС‚Р°РЅРґР°СЂС‚РЅРѕРіРѕ С„Р°Р№Р»Р° Рё РёРЅРґРµРєСЃРёСЂСѓСЋС‚СЃСЏ РїРѕ Р°РєСЂРѕРЅРёРјСѓ.
func TestLoadServerDataLoadsCosmicObjectTypesFromDefaultFile(t *testing.T) {
	workingDirectory := t.TempDir()
	dataDirectory := filepath.Join(workingDirectory, "data")
	if err := os.Mkdir(dataDirectory, 0o700); err != nil {
		t.Fatalf("Mkdir returned error: %v", err)
	}
	if err := os.WriteFile(filepath.Join(dataDirectory, "Accounts.json"), []byte(`{"MaxID":0,"Items":{}}`), 0o600); err != nil {
		t.Fatalf("WriteFile returned error: %v", err)
	}
	if err := os.WriteFile(filepath.Join(dataDirectory, "Characters.json"), []byte(`{"MaxID":0,"Items":{}}`), 0o600); err != nil {
		t.Fatalf("WriteFile returned error: %v", err)
	}
	if err := os.WriteFile(filepath.Join(dataDirectory, "CosmicObjects.json"), []byte(`{"MaxID":0,"Items":{}}`), 0o600); err != nil {
		t.Fatalf("WriteFile returned error: %v", err)
	}
	if err := os.WriteFile(filepath.Join(dataDirectory, "ItemTypes.json"), []byte(`{"MaxID":0,"Items":{}}`), 0o600); err != nil {
		t.Fatalf("WriteFile returned error: %v", err)
	}
	if err := os.WriteFile(filepath.Join(dataDirectory, "CosmicObjectModels.json"), []byte(`{"MaxID":0,"Items":{}}`), 0o600); err != nil {
		t.Fatalf("WriteFile returned error: %v", err)
	}

	content := []byte(`{
  "MaxID": 1,
  "Items": {
    "1": {
      "ID": 1,
      "TitleRu": "РљРѕСЂР°Р±Р»СЊ",
      "TitleEn": "Ship",
      "Acronym": "Ship",
      "CharacterLocatable": true,
      "Movable": true,
      "Rotatable": true
    }
  }
}`)
	if err := os.WriteFile(filepath.Join(dataDirectory, "CosmicObjectTypes.json"), content, 0o600); err != nil {
		t.Fatalf("WriteFile returned error: %v", err)
	}

	serverData, err := LoadServerData(workingDirectory)
	if err != nil {
		t.Fatalf("LoadServerData returned error: %v", err)
	}

	cosmicObjectType, ok := serverData.CosmicObjectTypes.GetByAcronym("Ship")
	if !ok {
		t.Fatal("cosmic object type is not available by acronym")
	}
	if cosmicObjectType.TitleRu != "РљРѕСЂР°Р±Р»СЊ" {
		t.Fatalf("cosmic object type TitleRu = %q, want РљРѕСЂР°Р±Р»СЊ", cosmicObjectType.TitleRu)
	}
}

// РџСЂРѕРІРµСЂСЏРµС‚, С‡С‚Рѕ РєРѕСЃРјРёС‡РµСЃРєРёРµ РѕР±СЉРµРєС‚С‹ С‡РёС‚Р°СЋС‚СЃСЏ РёР· СЃС‚Р°РЅРґР°СЂС‚РЅРѕРіРѕ С„Р°Р№Р»Р° СЃ СЃРѕС…СЂР°РЅРµРЅРёРµРј РјРѕРґРµР»Рё.
func TestLoadServerDataLoadsCosmicObjectsFromDefaultFile(t *testing.T) {
	workingDirectory := t.TempDir()
	dataDirectory := filepath.Join(workingDirectory, "data")
	if err := os.Mkdir(dataDirectory, 0o700); err != nil {
		t.Fatalf("Mkdir returned error: %v", err)
	}
	if err := os.WriteFile(filepath.Join(dataDirectory, "Accounts.json"), []byte(`{"MaxID":0,"Items":{}}`), 0o600); err != nil {
		t.Fatalf("WriteFile returned error: %v", err)
	}
	if err := os.WriteFile(filepath.Join(dataDirectory, "Characters.json"), []byte(`{"MaxID":0,"Items":{}}`), 0o600); err != nil {
		t.Fatalf("WriteFile returned error: %v", err)
	}
	if err := os.WriteFile(filepath.Join(dataDirectory, "CosmicObjectTypes.json"), []byte(`{"MaxID":0,"Items":{}}`), 0o600); err != nil {
		t.Fatalf("WriteFile returned error: %v", err)
	}
	if err := os.WriteFile(filepath.Join(dataDirectory, "CosmicObjectModels.json"), []byte(`{"MaxID":0,"Items":{}}`), 0o600); err != nil {
		t.Fatalf("WriteFile returned error: %v", err)
	}
	if err := os.WriteFile(filepath.Join(dataDirectory, "ItemTypes.json"), []byte(`{"MaxID":0,"Items":{}}`), 0o600); err != nil {
		t.Fatalf("WriteFile returned error: %v", err)
	}

	content := []byte(`{
  "MaxID": 1,
  "Items": {
    "1": {
      "ID": 1,
      "Title": "",
      "CosmicObjectModelID": 23,
      "OwnerCharacterID": 1,
      "OwnerNpcClanID": 0,
      "CreatorCharacterID": 1,
      "Mass": 7.92,
      "Capacity": 0,
      "MaxArmor": 0,
      "MaxSpeed": 497,
      "MaxAngularSpeed": 3,
      "X": 0,
      "Y": 0,
      "Rotation": 0,
      "Armor": 0,
      "MaxAlongForce": 0,
      "MaxAcrossForce": 0,
      "MaxTorque": 0,
      "GeneratingPower": 0,
      "ConsumingPower": 0,
      "AlongForce": 0,
      "AcrossForce": 0,
      "Torque": 0,
      "Enabled": true,
      "LastReceivedDamageTime": 0,
      "Anchored": false,
      "Complexity": 0,
      "OccupiedVolume": 0,
      "MaxFuel": 0,
      "Fuel": 0,
      "Speed": 0,
      "AngularSpeed": 0
    }
  }
}`)
	if err := os.WriteFile(filepath.Join(dataDirectory, "CosmicObjects.json"), content, 0o600); err != nil {
		t.Fatalf("WriteFile returned error: %v", err)
	}

	serverData, err := LoadServerData(workingDirectory)
	if err != nil {
		t.Fatalf("LoadServerData returned error: %v", err)
	}

	cosmicObject, ok := serverData.CosmicObjects.Get(1)
	if !ok {
		t.Fatal("cosmic object is not available by ID")
	}
	if cosmicObject.CosmicObjectModelID != 23 {
		t.Fatalf("cosmic object model ID = %d, want 23", cosmicObject.CosmicObjectModelID)
	}
}

// РџСЂРѕРІРµСЂСЏРµС‚, С‡С‚Рѕ С‚РёРїС‹ РїСЂРµРґРјРµС‚РѕРІ С‡РёС‚Р°СЋС‚СЃСЏ РёР· СЃС‚Р°РЅРґР°СЂС‚РЅРѕРіРѕ С„Р°Р№Р»Р° Рё РёРЅРґРµРєСЃРёСЂСѓСЋС‚СЃСЏ РїРѕ Р°РєСЂРѕРЅРёРјСѓ.
func TestLoadServerDataLoadsitemTypesFromDefaultFile(t *testing.T) {
	workingDirectory := t.TempDir()
	dataDirectory := filepath.Join(workingDirectory, "data")
	if err := os.Mkdir(dataDirectory, 0o700); err != nil {
		t.Fatalf("Mkdir returned error: %v", err)
	}
	if err := os.WriteFile(filepath.Join(dataDirectory, "Accounts.json"), []byte(`{"MaxID":0,"Items":{}}`), 0o600); err != nil {
		t.Fatalf("WriteFile returned error: %v", err)
	}
	if err := os.WriteFile(filepath.Join(dataDirectory, "Characters.json"), []byte(`{"MaxID":0,"Items":{}}`), 0o600); err != nil {
		t.Fatalf("WriteFile returned error: %v", err)
	}
	if err := os.WriteFile(filepath.Join(dataDirectory, "CosmicObjects.json"), []byte(`{"MaxID":0,"Items":{}}`), 0o600); err != nil {
		t.Fatalf("WriteFile returned error: %v", err)
	}
	if err := os.WriteFile(filepath.Join(dataDirectory, "CosmicObjectTypes.json"), []byte(`{"MaxID":0,"Items":{}}`), 0o600); err != nil {
		t.Fatalf("WriteFile returned error: %v", err)
	}
	if err := os.WriteFile(filepath.Join(dataDirectory, "CosmicObjectModels.json"), []byte(`{"MaxID":0,"Items":{}}`), 0o600); err != nil {
		t.Fatalf("WriteFile returned error: %v", err)
	}

	content := []byte(`{
  "MaxID": 1,
  "Items": {
    "1": {
      "ID": 1,
      "TitleRu": "РћСЂСѓР¶РёРµ",
      "TitleEn": "Weapon",
      "Acronym": "Weapon",
      "IsEquipmentForShip": true,
      "IsEquipmentForStation": true,
      "IsPilotInstrument": true,
      "CountMustBeInteger": true
    }
  }
}`)
	if err := os.WriteFile(filepath.Join(dataDirectory, "ItemTypes.json"), content, 0o600); err != nil {
		t.Fatalf("WriteFile returned error: %v", err)
	}

	serverData, err := LoadServerData(workingDirectory)
	if err != nil {
		t.Fatalf("LoadServerData returned error: %v", err)
	}

	itemType, ok := serverData.ItemTypes.GetByAcronym("Weapon")
	if !ok {
		t.Fatal("itemType is not available by acronym")
	}
	if itemType.TitleRu != "РћСЂСѓР¶РёРµ" {
		t.Fatalf("itemType TitleRu = %q, want РћСЂСѓР¶РёРµ", itemType.TitleRu)
	}
}

// РџСЂРѕРІРµСЂСЏРµС‚, С‡С‚Рѕ РјРѕРґРµР»Рё РєРѕСЃРјРёС‡РµСЃРєРёС… РѕР±СЉРµРєС‚РѕРІ С‡РёС‚Р°СЋС‚СЃСЏ РёР· СЃС‚Р°РЅРґР°СЂС‚РЅРѕРіРѕ С„Р°Р№Р»Р° СЃ РІС‹С‡РёСЃР»РµРЅРЅС‹РјРё СЂР°Р·РјРµСЂР°РјРё С‚РµР»Р°.
func TestLoadServerDataLoadsCosmicObjectModelsFromDefaultFile(t *testing.T) {
	workingDirectory := t.TempDir()
	dataDirectory := filepath.Join(workingDirectory, "data")
	if err := os.Mkdir(dataDirectory, 0o700); err != nil {
		t.Fatalf("Mkdir returned error: %v", err)
	}
	if err := os.WriteFile(filepath.Join(dataDirectory, "Accounts.json"), []byte(`{"MaxID":0,"Items":{}}`), 0o600); err != nil {
		t.Fatalf("WriteFile returned error: %v", err)
	}
	if err := os.WriteFile(filepath.Join(dataDirectory, "Characters.json"), []byte(`{"MaxID":0,"Items":{}}`), 0o600); err != nil {
		t.Fatalf("WriteFile returned error: %v", err)
	}
	if err := os.WriteFile(filepath.Join(dataDirectory, "CosmicObjects.json"), []byte(`{"MaxID":0,"Items":{}}`), 0o600); err != nil {
		t.Fatalf("WriteFile returned error: %v", err)
	}
	if err := os.WriteFile(filepath.Join(dataDirectory, "CosmicObjectTypes.json"), []byte(`{"MaxID":0,"Items":{}}`), 0o600); err != nil {
		t.Fatalf("WriteFile returned error: %v", err)
	}
	if err := os.WriteFile(filepath.Join(dataDirectory, "ItemTypes.json"), []byte(`{"MaxID":0,"Items":{}}`), 0o600); err != nil {
		t.Fatalf("WriteFile returned error: %v", err)
	}

	content := []byte(`{
  "MaxID": 1,
  "Items": {
    "1": {
      "ID": 1,
      "TitleRu": "РђСЃС‚РµСЂРѕРёРґ 1",
      "TitleEn": "Asteroid 1",
      "Acronym": "asteroid_0001",
      "IconFilePath": "",
      "TextureFilePath": "assets/asteroids/asteroid_0001.png",
      "TextureWidth": 2048,
      "TextureHeight": 2048,
      "TextureBodyOriginX": 1055,
      "TextureBodyOriginY": 1107,
      "TextureBodyWidth": 882,
      "TextureBodyLength": 870,
      "TextureScale": 4,
      "CosmicObjectTypeID": 3,
      "Mass": 767.34,
      "Capacity": 0,
      "MaxArmor": 0,
      "MaxSpeed": 472,
      "MaxAngularSpeed": 3,
      "Complexity": 0,
      "BodyLength": 217.5,
      "BodyWidth": 220.5
    }
  }
}`)
	if err := os.WriteFile(filepath.Join(dataDirectory, "CosmicObjectModels.json"), content, 0o600); err != nil {
		t.Fatalf("WriteFile returned error: %v", err)
	}

	serverData, err := LoadServerData(workingDirectory)
	if err != nil {
		t.Fatalf("LoadServerData returned error: %v", err)
	}

	cosmicObjectModel, ok := serverData.CosmicObjectModels.GetByAcronym("asteroid_0001")
	if !ok {
		t.Fatal("cosmic object model is not available by acronym")
	}
	if cosmicObjectModel.BodyLength != 217.5 {
		t.Fatalf("cosmic object model BodyLength = %v, want 217.5", cosmicObjectModel.BodyLength)
	}
}

// РџСЂРѕРІРµСЂСЏРµС‚, С‡С‚Рѕ СЂРµР°Р»СЊРЅС‹Рµ С„Р°Р№Р»С‹ СЂРµРїРѕР·РёС‚РѕСЂРёСЏ Р·Р°РіСЂСѓР¶Р°СЋС‚СЃСЏ РєР°Рє РµРґРёРЅС‹Р№ РЅР°Р±РѕСЂ СЃРµСЂРІРµСЂРЅС‹С… РґР°РЅРЅС‹С….
func TestLoadServerDataLoadsRepositoryAccountsFile(t *testing.T) {
	serverData, err := LoadServerData(filepath.Join("..", ".."))
	if err != nil {
		t.Fatalf("LoadServerData returned error: %v", err)
	}

	if _, ok := serverData.Accounts.GetByEmail("index@email.net"); !ok {
		t.Fatal("repository Accounts.json does not contain index@email.net")
	}
	if serverData.Characters == nil {
		t.Fatal("repository Characters.json was not loaded")
	}
	if _, ok := serverData.CosmicObjects.Get(1); !ok {
		t.Fatal("repository CosmicObjects.json does not contain object 1")
	}
	if _, ok := serverData.CosmicObjectTypes.GetByAcronym("Ship"); !ok {
		t.Fatal("repository CosmicObjectTypes.json does not contain Ship")
	}
	if _, ok := serverData.CosmicObjectModels.GetByAcronym("asteroid_0001"); !ok {
		t.Fatal("repository CosmicObjectModels.json does not contain asteroid_0001")
	}
	if _, ok := serverData.ItemTypes.GetByAcronym("Weapon"); !ok {
		t.Fatal("repository itemTypes.json does not contain Weapon")
	}
}

// РџСЂРѕРІРµСЂСЏРµС‚, С‡С‚Рѕ РѕС‚СЃСѓС‚СЃС‚РІСѓСЋС‰РёРµ С„Р°Р№Р»С‹ С‡Р°С‚-С‚Р°Р±Р»РёС† РґР°СЋС‚ РїСѓСЃС‚С‹Рµ С…СЂР°РЅРёР»РёС‰Р° РґР»СЏ РїРµСЂРІРѕР№ СЃРµСЂРІРµСЂРЅРѕР№ РёРЅРёС†РёР°Р»РёР·Р°С†РёРё.
func TestLoadServerDataCreatesEmptyChatTablesWhenFilesAreMissing(t *testing.T) {
	workingDirectory := t.TempDir()
	dataDirectory := filepath.Join(workingDirectory, "data")
	if err := os.Mkdir(dataDirectory, 0o700); err != nil {
		t.Fatalf("Mkdir returned error: %v", err)
	}
	for fileName, content := range map[string][]byte{
		"Accounts.json":           []byte(`{"MaxID":0,"Items":{}}`),
		"Characters.json":         []byte(`{"MaxID":0,"Items":{}}`),
		"CosmicObjects.json":      []byte(`{"MaxID":0,"Items":{}}`),
		"CosmicObjectTypes.json":  []byte(`{"MaxID":0,"Items":{}}`),
		"CosmicObjectModels.json": []byte(`{"MaxID":0,"Items":{}}`),
		"ItemTypes.json":          []byte(`{"MaxID":0,"Items":{}}`),
	} {
		if err := os.WriteFile(filepath.Join(dataDirectory, fileName), content, 0o600); err != nil {
			t.Fatalf("WriteFile %s returned error: %v", fileName, err)
		}
	}

	serverData, err := LoadServerData(workingDirectory)
	if err != nil {
		t.Fatalf("LoadServerData returned error: %v", err)
	}

	if serverData.Chats == nil || serverData.ChatMembers == nil || serverData.Messages == nil {
		t.Fatal("chat runtime tables were not initialized")
	}
	if serverData.CommunityTypes == nil || serverData.CommunityChatRoles == nil || serverData.MessageTypes == nil {
		t.Fatal("chat reference tables were not initialized")
	}
}
