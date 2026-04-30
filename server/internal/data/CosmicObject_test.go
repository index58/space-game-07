package data

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestCosmicObjectsAddAssignsIDAndIndexesObject(t *testing.T) {
	cosmicObjects := NewCosmicObjects()

	cosmicObject, err := cosmicObjects.Add(&CosmicObject{
		Title:               "Стартовый корабль",
		CosmicObjectModelID: 23,
		OwnerCharacterID:    1,
		CreatorCharacterID:  1,
		Mass:                7.92,
		MaxSpeed:            497,
		MaxAngularSpeed:     3,
		Enabled:             true,
	})
	if err != nil {
		t.Fatalf("Add returned error: %v", err)
	}

	if cosmicObject.ID != 1 {
		t.Fatalf("cosmic object ID = %d, want 1", cosmicObject.ID)
	}
	if cosmicObjects.MaxID != 1 {
		t.Fatalf("MaxID = %d, want 1", cosmicObjects.MaxID)
	}

	byID, ok := cosmicObjects.Get(cosmicObject.ID)
	if !ok || byID != cosmicObject {
		t.Fatal("Get did not return added cosmic object")
	}

	byModelID := cosmicObjects.GetByCosmicObjectModelID(cosmicObject.CosmicObjectModelID)
	if len(byModelID) != 1 || byModelID[0] != cosmicObject {
		t.Fatal("GetByCosmicObjectModelID did not return added cosmic object")
	}

	byOwnerID := cosmicObjects.GetByOwnerCharacterID(cosmicObject.OwnerCharacterID)
	if len(byOwnerID) != 1 || byOwnerID[0] != cosmicObject {
		t.Fatal("GetByOwnerCharacterID did not return added cosmic object")
	}
}

func TestCosmicObjectsAddRejectsEmptyModelID(t *testing.T) {
	cosmicObjects := NewCosmicObjects()

	if _, err := cosmicObjects.Add(&CosmicObject{}); err == nil {
		t.Fatal("Add accepted empty CosmicObjectModelID")
	}
}

func TestCosmicObjectsDeleteRemovesObjectAndIndexes(t *testing.T) {
	cosmicObjects := NewCosmicObjects()
	cosmicObject, err := cosmicObjects.Add(&CosmicObject{CosmicObjectModelID: 23, OwnerCharacterID: 1})
	if err != nil {
		t.Fatalf("Add returned error: %v", err)
	}

	if !cosmicObjects.Delete(cosmicObject.ID) {
		t.Fatal("Delete returned false")
	}

	if _, ok := cosmicObjects.Get(cosmicObject.ID); ok {
		t.Fatal("deleted cosmic object is still stored by ID")
	}
	if byModelID := cosmicObjects.GetByCosmicObjectModelID(cosmicObject.CosmicObjectModelID); len(byModelID) != 0 {
		t.Fatal("deleted cosmic object is still indexed by model ID")
	}
	if byOwnerID := cosmicObjects.GetByOwnerCharacterID(cosmicObject.OwnerCharacterID); len(byOwnerID) != 0 {
		t.Fatal("deleted cosmic object is still indexed by owner character ID")
	}
}

func TestCosmicObjectsSaveLoadAndRebuildIndexes(t *testing.T) {
	path := filepath.Join(t.TempDir(), "CosmicObjects.json")
	cosmicObjects := NewCosmicObjects()
	cosmicObject, err := cosmicObjects.Add(&CosmicObject{
		CosmicObjectModelID: 2,
		Mass:                629.532,
		MaxSpeed:            475,
		MaxAngularSpeed:     3,
		X:                   -500,
		Y:                   800,
	})
	if err != nil {
		t.Fatalf("Add returned error: %v", err)
	}

	if err := cosmicObjects.SaveToFile(path); err != nil {
		t.Fatalf("SaveToFile returned error: %v", err)
	}
	if _, err := os.Stat(path); err != nil {
		t.Fatalf("saved file is not available: %v", err)
	}

	loaded := NewCosmicObjects()
	if err := loaded.LoadFromFile(path); err != nil {
		t.Fatalf("LoadFromFile returned error: %v", err)
	}

	loadedObject, ok := loaded.Get(cosmicObject.ID)
	if !ok {
		t.Fatal("loaded cosmic object is not available by ID")
	}
	if loadedObject.CosmicObjectModelID != cosmicObject.CosmicObjectModelID || loadedObject.X != cosmicObject.X || loadedObject.Y != cosmicObject.Y {
		t.Fatal("loaded cosmic object fields do not match saved object")
	}
}

func TestCosmicObjectsJSONKeysMatchGoFieldNames(t *testing.T) {
	cosmicObjects := NewCosmicObjects()
	if _, err := cosmicObjects.Add(&CosmicObject{CosmicObjectModelID: 23}); err != nil {
		t.Fatalf("Add returned error: %v", err)
	}

	content, err := json.Marshal(cosmicObjects)
	if err != nil {
		t.Fatalf("Marshal returned error: %v", err)
	}
	text := string(content)

	expectedKeys := []string{
		`"MaxID"`,
		`"Items"`,
		`"ID"`,
		`"Title"`,
		`"CosmicObjectModelID"`,
		`"OwnerCharacterID"`,
		`"Mass"`,
		`"MaxSpeed"`,
		`"Enabled"`,
		`"VelocityX"`,
		`"VelocityY"`,
		`"TargetRotation"`,
		`"AngularSpeed"`,
	}
	for _, expectedKey := range expectedKeys {
		if !strings.Contains(text, expectedKey) {
			t.Fatalf("JSON %s does not contain key %s", text, expectedKey)
		}
	}
}

func TestCosmicObjectsRebuildIndexesRejectsInvalidStoredObject(t *testing.T) {
	cosmicObjects := NewCosmicObjects()
	cosmicObjects.Items[1] = &CosmicObject{ID: 1}

	if err := cosmicObjects.RebuildIndexes(); err == nil {
		t.Fatal("RebuildIndexes accepted cosmic object without model ID")
	}
}
