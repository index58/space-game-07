package storage

import (
	"os"
	"path/filepath"
	"testing"
)

// Проверяет, что дополнительные справочные таблицы загружаются из файлов при их наличии.
func TestLoadServerDataLoadsOptionalReferenceTablesWhenFilesExist(t *testing.T) {
	workingDirectory := t.TempDir()
	dataDirectory := filepath.Join(workingDirectory, "data")
	if err := os.Mkdir(dataDirectory, 0o700); err != nil {
		t.Fatalf("Mkdir returned error: %v", err)
	}

	requiredFiles := map[string]string{
		"Accounts.json":                `{"MaxID":0,"Items":{}}`,
		"Characters.json":              `{"MaxID":0,"Items":{}}`,
		"CosmicObjects.json":           `{"MaxID":0,"Items":{}}`,
		"CosmicObjectTypes.json":       `{"MaxID":0,"Items":{}}`,
		"CosmicObjectModels.json":      `{"MaxID":0,"Items":{}}`,
		"ItemTypes.json":               `{"MaxID":0,"Items":{}}`,
		"EquipmentGroups.json":         `{"MaxID":1,"Items":{"1":{"ID":1,"CosmicObjectID":1,"Title":"Main","EquipmentItemModelID":1,"Count":2,"EnabledCount":1,"Enabled":true,"Active":true}}}`,
		"ItemGroups.json":              `{"MaxID":1,"Items":{"1":{"ID":1,"ContainerEquipmentGroupID":1,"ContentItemModelID":2,"Count":3}}}`,
		"Assemblies.json":              `{"MaxID":1,"Items":{"1":{"ID":1,"Title":"Default","CosmicObjectModelID":1,"IsPublic":true}}}`,
		"AssemblyEquipmentGroups.json": `{"MaxID":1,"Items":{"1":{"ID":1,"AssemblyID":1,"Title":"Thrusters","EquipmentItemModelID":1,"Count":2}}}`,
		"NpcClans.json":                `{"MaxID":1,"Items":{"1":{"ID":1,"TitleRu":"Клан","TitleEn":"Clan","Acronym":"Clan"}}}`,
		"ItemModels.json":              `{"MaxID":1,"Items":{"1":{"ID":1,"Acronym":"MagicThruster"}}}`,
		"Blueprints.json":              `{"MaxID":1,"Items":{"1":{"ID":1,"TitleEn":"Ship"}}}`,
		"BlueprintComponents.json":     `{"MaxID":1,"Items":{"1":{"ID":1,"BlueprintID":1}}}`,
		"Schemas.json":                 `{"MaxID":1,"Items":{"1":{"ID":1,"TitleEn":"Thruster"}}}`,
		"SchemaComponents.json":        `{"MaxID":1,"Items":{"1":{"ID":1,"SchemaID":1}}}`,
	}
	for fileName, content := range requiredFiles {
		if err := os.WriteFile(filepath.Join(dataDirectory, fileName), []byte(content), 0o600); err != nil {
			t.Fatalf("WriteFile %s returned error: %v", fileName, err)
		}
	}

	serverData, err := LoadServerData(workingDirectory)
	if err != nil {
		t.Fatalf("LoadServerData returned error: %v", err)
	}

	if serverData.NpcClans.MaxID != 1 {
		t.Fatalf("NpcClans MaxID = %d, want 1", serverData.NpcClans.MaxID)
	}
	if _, ok := serverData.NpcClans.Items["1"]; !ok {
		t.Fatal("NpcClan item 1 was not loaded")
	}
	if _, ok := serverData.Assemblies.Get(1); !ok {
		t.Fatal("Assemblies item 1 was not loaded")
	}
	if _, ok := serverData.AssemblyEquipmentGroups.Items[1]; !ok {
		t.Fatal("AssemblyEquipmentGroups item 1 was not loaded")
	}
	equipmentGroup, ok := serverData.EquipmentGroups.Get(1)
	if !ok {
		t.Fatal("EquipmentGroups item 1 was not loaded")
	}
	if equipmentGroup.Count != 2 || equipmentGroup.EnabledCount != 1 {
		t.Fatalf("EquipmentGroups counts were not loaded as integers: %+v", equipmentGroup)
	}
	itemGroup, ok := serverData.ItemGroups.Get(1)
	if !ok {
		t.Fatal("ItemGroups item 1 was not loaded")
	}
	if itemGroup.ContainerEquipmentGroupID != 1 || itemGroup.ContentItemModelID != 2 || itemGroup.Count != 3 {
		t.Fatalf("ItemGroups item 1 was loaded incorrectly: %+v", itemGroup)
	}
	if _, ok := serverData.ItemModels.Get(1); !ok {
		t.Fatal("ItemModels item 1 was not loaded")
	}
	if _, ok := serverData.Blueprints.Items["1"]; !ok {
		t.Fatal("Blueprints item 1 was not loaded")
	}
	if _, ok := serverData.BlueprintComponents.Items["1"]; !ok {
		t.Fatal("BlueprintComponents item 1 was not loaded")
	}
	if _, ok := serverData.Schemas.Items["1"]; !ok {
		t.Fatal("Schemas item 1 was not loaded")
	}
	if _, ok := serverData.SchemaComponents.Items["1"]; !ok {
		t.Fatal("SchemaComponents item 1 was not loaded")
	}
}
