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

func TestLoadServerDataLoadsRepositoryAccountsFile(t *testing.T) {
	serverData, err := LoadServerData(filepath.Join("..", ".."))
	if err != nil {
		t.Fatalf("LoadServerData returned error: %v", err)
	}

	if _, ok := serverData.Accounts.GetByEmail("index@email.net"); !ok {
		t.Fatal("repository Accounts.json does not contain index@email.net")
	}
}
