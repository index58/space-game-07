package storage

import (
	"os"
	"path/filepath"
	"testing"
)

// РџСЂРѕРІРµСЂСЏРµС‚, С‡С‚Рѕ РґРµР№СЃС‚РІРёСЏ СЃС‚С‹РєРѕРІРєРё РґРѕСЃС‚СѓРїРЅС‹ РІ СЃРїСЂР°РІРѕС‡РЅРёРєРµ РЅР°СЃС‚СЂРѕРµРє РІРІРѕРґР°.
func TestServerDataContainsDockingInputActions(t *testing.T) {
	workingDirectory := serverDataWorkingDirectory(t)
	serverData, err := LoadServerData(workingDirectory)
	if err != nil {
		t.Fatalf("LoadServerData returned error: %v", err)
	}

	expected := map[string]string{
		"DockingRequest": "KeyboardEvent.altKey&&KeyboardEvent.code:Equal",
		"ApproveRequest": "KeyboardEvent.altKey&&KeyboardEvent.code:Digit1",
		"RejectRequest":  "KeyboardEvent.altKey&&KeyboardEvent.code:Digit2",
		"DockingUndock":  "KeyboardEvent.altKey&&KeyboardEvent.code:Minus",
	}
	for acronym, systemStringValue := range expected {
		action := serverData.ActionTypes.ByAcronym[acronym]
		if action == nil {
			t.Fatalf("action %s is missing", acronym)
		}
		setting := serverData.DefaultActionInputSettings.ByActionTypeID[action.ID]
		if setting == nil {
			t.Fatalf("default setting for %s is missing", acronym)
		}
		eventType, ok := serverData.InputEventTypes.Get(setting.InputEventTypeID)
		if !ok {
			t.Fatalf("input event for %s is missing", acronym)
		}
		if eventType.SystemStringValue != systemStringValue {
			t.Fatalf("%s default = %q, want %q", acronym, eventType.SystemStringValue, systemStringValue)
		}
	}
}

// РќР°С…РѕРґРёС‚ РєРѕСЂРµРЅСЊ СЃРµСЂРІРµСЂРЅРѕР№ С‡Р°СЃС‚Рё РЅРµР·Р°РІРёСЃРёРјРѕ РѕС‚ С‚РµРєСѓС‰РµРіРѕ РєР°С‚Р°Р»РѕРіР° С‚РµСЃС‚РѕРІРѕРіРѕ РїСЂРѕС†РµСЃСЃР°.
func serverDataWorkingDirectory(t *testing.T) string {
	t.Helper()
	workingDirectory, err := os.Getwd()
	if err != nil {
		t.Fatalf("Getwd returned error: %v", err)
	}
	for {
		if _, err := os.Stat(filepath.Join(workingDirectory, "data", "ActionTypes.json")); err == nil {
			return workingDirectory
		}
		next := filepath.Dir(workingDirectory)
		if next == workingDirectory {
			t.Fatal("server data directory is not found")
		}
		workingDirectory = next
	}
}
