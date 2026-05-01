package storage

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoadServerDataLoadsOptionalReferenceTablesWhenFilesExist(t *testing.T) {
	workingDirectory := t.TempDir()
	dataDirectory := filepath.Join(workingDirectory, "data")
	if err := os.Mkdir(dataDirectory, 0o700); err != nil {
		t.Fatalf("Mkdir returned error: %v", err)
	}

	requiredFiles := map[string]string{
		"Accounts.json":           `{"MaxID":0,"Items":{}}`,
		"Characters.json":         `{"MaxID":0,"Items":{}}`,
		"CosmicObjects.json":      `{"MaxID":0,"Items":{}}`,
		"CosmicObjectTypes.json":  `{"MaxID":0,"Items":{}}`,
		"CosmicObjectModels.json": `{"MaxID":0,"Items":{}}`,
		"Itemtypes.json":          `{"MaxID":0,"Items":{}}`,
		"NpcClan.json":            `{"MaxID":1,"Items":{"1":{"ID":1,"TitleRu":"Клан","TitleEn":"Clan","Acronym":"Clan"}}}`,
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
	if len(serverData.ItemModels.Items) != 0 {
		t.Fatal("missing optional ItemModel.json should produce an empty table")
	}
}
