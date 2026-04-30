package storage

import (
	"os"
	"path/filepath"
	"testing"
)

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
	if err := os.WriteFile(filepath.Join(dataDirectory, "CosmicObjectTypes.json"), []byte(`{"MaxID":0,"Items":{}}`), 0o600); err != nil {
		t.Fatalf("WriteFile returned error: %v", err)
	}
	if err := os.WriteFile(filepath.Join(dataDirectory, "CosmicObjectModels.json"), []byte(`{"MaxID":0,"Items":{}}`), 0o600); err != nil {
		t.Fatalf("WriteFile returned error: %v", err)
	}
	if err := os.WriteFile(filepath.Join(dataDirectory, "Itemtypes.json"), []byte(`{"MaxID":0,"Items":{}}`), 0o600); err != nil {
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

func TestLoadServerDataReturnsErrorWhenAccountsFileIsMissing(t *testing.T) {
	_, err := LoadServerData(t.TempDir())
	if err == nil {
		t.Fatal("LoadServerData accepted missing Accounts.json")
	}
}

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
	if err := os.WriteFile(filepath.Join(dataDirectory, "Itemtypes.json"), []byte(`{"MaxID":0,"Items":{}}`), 0o600); err != nil {
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
	if err := os.WriteFile(filepath.Join(dataDirectory, "Itemtypes.json"), []byte(`{"MaxID":0,"Items":{}}`), 0o600); err != nil {
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
      "TitleRu": "Корабль",
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
	if cosmicObjectType.TitleRu != "Корабль" {
		t.Fatalf("cosmic object type TitleRu = %q, want Корабль", cosmicObjectType.TitleRu)
	}
}

func TestLoadServerDataLoadsItemtypesFromDefaultFile(t *testing.T) {
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

	content := []byte(`{
  "MaxID": 1,
  "Items": {
    "1": {
      "ID": 1,
      "TitleRu": "Оружие",
      "TitleEn": "Weapon",
      "Acronym": "Weapon",
      "IsEquipmentForShip": true,
      "IsEquipmentForStation": true,
      "IsPilotInstrument": true,
      "CountMustBeInteger": true
    }
  }
}`)
	if err := os.WriteFile(filepath.Join(dataDirectory, "Itemtypes.json"), content, 0o600); err != nil {
		t.Fatalf("WriteFile returned error: %v", err)
	}

	serverData, err := LoadServerData(workingDirectory)
	if err != nil {
		t.Fatalf("LoadServerData returned error: %v", err)
	}

	itemtype, ok := serverData.Itemtypes.GetByAcronym("Weapon")
	if !ok {
		t.Fatal("itemtype is not available by acronym")
	}
	if itemtype.TitleRu != "Оружие" {
		t.Fatalf("itemtype TitleRu = %q, want Оружие", itemtype.TitleRu)
	}
}

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
	if err := os.WriteFile(filepath.Join(dataDirectory, "CosmicObjectTypes.json"), []byte(`{"MaxID":0,"Items":{}}`), 0o600); err != nil {
		t.Fatalf("WriteFile returned error: %v", err)
	}
	if err := os.WriteFile(filepath.Join(dataDirectory, "Itemtypes.json"), []byte(`{"MaxID":0,"Items":{}}`), 0o600); err != nil {
		t.Fatalf("WriteFile returned error: %v", err)
	}

	content := []byte(`{
  "MaxID": 1,
  "Items": {
    "1": {
      "ID": 1,
      "TitleRu": "Астероид 1",
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
	if _, ok := serverData.CosmicObjectTypes.GetByAcronym("Ship"); !ok {
		t.Fatal("repository CosmicObjectTypes.json does not contain Ship")
	}
	if _, ok := serverData.CosmicObjectModels.GetByAcronym("asteroid_0001"); !ok {
		t.Fatal("repository CosmicObjectModels.json does not contain asteroid_0001")
	}
	if _, ok := serverData.Itemtypes.GetByAcronym("Weapon"); !ok {
		t.Fatal("repository Itemtypes.json does not contain Weapon")
	}
}
