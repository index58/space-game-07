package data

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestAccountsAddAssignsIDGeneratesTokenAndIndexesAccount(t *testing.T) {
	accounts := NewAccounts()

	account, err := accounts.Add(&Account{
		Email:        "pilot@example.com",
		Nickname:     "Pilot",
		PasswordHash: "initial-hash",
	})
	if err != nil {
		t.Fatalf("Add returned error: %v", err)
	}

	if account.ID != 1 {
		t.Fatalf("account ID = %d, want 1", account.ID)
	}
	if accounts.MaxID != 1 {
		t.Fatalf("MaxID = %d, want 1", accounts.MaxID)
	}
	if account.Token == "" {
		t.Fatal("Token is empty")
	}

	byID, ok := accounts.Get(account.ID)
	if !ok || byID != account {
		t.Fatal("Get did not return added account")
	}

	byEmail, ok := accounts.GetByEmail(account.Email)
	if !ok || byEmail != account {
		t.Fatal("GetByEmail did not return added account")
	}

	byNickname, ok := accounts.GetByNickname(account.Nickname)
	if !ok || byNickname != account {
		t.Fatal("GetByNickname did not return added account")
	}

	byToken, ok := accounts.GetByToken(account.Token)
	if !ok || byToken != account {
		t.Fatal("GetByToken did not return added account")
	}
}

func TestAccountsAddRejectsDuplicateEmailAndNickname(t *testing.T) {
	accounts := NewAccounts()

	if _, err := accounts.Add(&Account{Email: "pilot@example.com", Nickname: "Pilot", PasswordHash: "hash"}); err != nil {
		t.Fatalf("first Add returned error: %v", err)
	}

	if _, err := accounts.Add(&Account{Email: "pilot@example.com", Nickname: "Other", PasswordHash: "hash"}); err == nil {
		t.Fatal("Add accepted duplicate email")
	}

	if _, err := accounts.Add(&Account{Email: "other@example.com", Nickname: "Pilot", PasswordHash: "hash"}); err == nil {
		t.Fatal("Add accepted duplicate nickname")
	}
}

func TestAccountsSetEmailAndNicknameUpdateIndexes(t *testing.T) {
	accounts := NewAccounts()
	account, err := accounts.Add(&Account{Email: "pilot@example.com", Nickname: "Pilot", PasswordHash: "hash"})
	if err != nil {
		t.Fatalf("Add returned error: %v", err)
	}

	if err := accounts.SetEmail(account.ID, "captain@example.com"); err != nil {
		t.Fatalf("SetEmail returned error: %v", err)
	}
	if _, ok := accounts.GetByEmail("pilot@example.com"); ok {
		t.Fatal("old email is still indexed")
	}
	if byEmail, ok := accounts.GetByEmail("captain@example.com"); !ok || byEmail != account {
		t.Fatal("new email is not indexed")
	}

	if err := accounts.SetNickname(account.ID, "Captain"); err != nil {
		t.Fatalf("SetNickname returned error: %v", err)
	}
	if _, ok := accounts.GetByNickname("Pilot"); ok {
		t.Fatal("old nickname is still indexed")
	}
	if byNickname, ok := accounts.GetByNickname("Captain"); !ok || byNickname != account {
		t.Fatal("new nickname is not indexed")
	}
}

func TestAccountsSetPasswordHashesPassword(t *testing.T) {
	accounts := NewAccounts()
	account, err := accounts.Add(&Account{Email: "pilot@example.com", Nickname: "Pilot", PasswordHash: "old-hash"})
	if err != nil {
		t.Fatalf("Add returned error: %v", err)
	}

	if err := accounts.SetPassword(account.ID, "secret-password"); err != nil {
		t.Fatalf("SetPassword returned error: %v", err)
	}

	if account.PasswordHash == "secret-password" {
		t.Fatal("password was stored without hashing")
	}
	if !strings.HasPrefix(account.PasswordHash, "sha256$") {
		t.Fatalf("PasswordHash = %q, want sha256 prefix", account.PasswordHash)
	}
}

func TestAccountsGenerateTokenReplacesTokenIndex(t *testing.T) {
	accounts := NewAccounts()
	account, err := accounts.Add(&Account{Email: "pilot@example.com", Nickname: "Pilot", PasswordHash: "hash"})
	if err != nil {
		t.Fatalf("Add returned error: %v", err)
	}
	oldToken := account.Token

	newToken, err := accounts.GenerateToken(account.ID)
	if err != nil {
		t.Fatalf("GenerateToken returned error: %v", err)
	}

	if newToken == "" || newToken == oldToken {
		t.Fatalf("new token = %q, old token = %q", newToken, oldToken)
	}
	if _, ok := accounts.GetByToken(oldToken); ok {
		t.Fatal("old token is still indexed")
	}
	if byToken, ok := accounts.GetByToken(newToken); !ok || byToken != account {
		t.Fatal("new token is not indexed")
	}
}

func TestAccountsSetCurrentCharacterUpdatesIndex(t *testing.T) {
	accounts := NewAccounts()
	account, err := accounts.Add(&Account{Email: "pilot@example.com", Nickname: "Pilot", PasswordHash: "hash"})
	if err != nil {
		t.Fatalf("Add returned error: %v", err)
	}

	if err := accounts.SetCurrentCharacter(account.ID, 10); err != nil {
		t.Fatalf("SetCurrentCharacter returned error: %v", err)
	}

	if account.CurrentCharacterID != 10 {
		t.Fatalf("CurrentCharacterID = %d, want 10", account.CurrentCharacterID)
	}
	if byCharacterID, ok := accounts.GetByCurrentCharacterID(10); !ok || byCharacterID != account {
		t.Fatal("account is not indexed by current character")
	}
}

func TestAccountsDeleteRemovesAccountAndIndexes(t *testing.T) {
	accounts := NewAccounts()
	account, err := accounts.Add(&Account{Email: "pilot@example.com", Nickname: "Pilot", PasswordHash: "hash"})
	if err != nil {
		t.Fatalf("Add returned error: %v", err)
	}

	if !accounts.Delete(account.ID) {
		t.Fatal("Delete returned false")
	}

	if _, ok := accounts.Get(account.ID); ok {
		t.Fatal("deleted account is still stored by ID")
	}
	if _, ok := accounts.GetByEmail(account.Email); ok {
		t.Fatal("deleted account email is still indexed")
	}
	if _, ok := accounts.GetByNickname(account.Nickname); ok {
		t.Fatal("deleted account nickname is still indexed")
	}
	if _, ok := accounts.GetByToken(account.Token); ok {
		t.Fatal("deleted account token is still indexed")
	}
}

func TestAccountsSaveLoadAndRebuildIndexes(t *testing.T) {
	path := filepath.Join(t.TempDir(), "accounts.json")
	accounts := NewAccounts()
	account, err := accounts.Add(&Account{Email: "pilot@example.com", Nickname: "Pilot", PasswordHash: "hash"})
	if err != nil {
		t.Fatalf("Add returned error: %v", err)
	}

	if err := accounts.SaveToFile(path); err != nil {
		t.Fatalf("SaveToFile returned error: %v", err)
	}
	if _, err := os.Stat(path); err != nil {
		t.Fatalf("saved file is not available: %v", err)
	}

	loaded := NewAccounts()
	if err := loaded.LoadFromFile(path); err != nil {
		t.Fatalf("LoadFromFile returned error: %v", err)
	}

	if loaded.MaxID != accounts.MaxID {
		t.Fatalf("loaded MaxID = %d, want %d", loaded.MaxID, accounts.MaxID)
	}
	loadedAccount, ok := loaded.Get(account.ID)
	if !ok {
		t.Fatal("loaded account is not available by ID")
	}
	if loadedAccount.Email != account.Email || loadedAccount.Nickname != account.Nickname || loadedAccount.Token != account.Token {
		t.Fatal("loaded account fields do not match saved account")
	}
	if byEmail, ok := loaded.GetByEmail(account.Email); !ok || byEmail != loadedAccount {
		t.Fatal("loaded email index is not rebuilt")
	}
}

func TestAccountsSaveToFileOrdersItemsByNumericID(t *testing.T) {
	path := filepath.Join(t.TempDir(), "accounts.json")
	accounts := NewAccounts()
	accounts.MaxID = 10
	accounts.Items[10] = &Account{ID: 10, Email: "ten@example.com", Nickname: "Ten", PasswordHash: "hash", Token: "token-10"}
	accounts.Items[2] = &Account{ID: 2, Email: "two@example.com", Nickname: "Two", PasswordHash: "hash", Token: "token-2"}
	accounts.Items[1] = &Account{ID: 1, Email: "one@example.com", Nickname: "One", PasswordHash: "hash", Token: "token-1"}

	if err := accounts.SaveToFile(path); err != nil {
		t.Fatalf("SaveToFile returned error: %v", err)
	}

	content, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("ReadFile returned error: %v", err)
	}
	text := string(content)
	firstIndex := strings.Index(text, `"1":`)
	secondIndex := strings.Index(text, `"2":`)
	tenthIndex := strings.Index(text, `"10":`)
	if firstIndex < 0 || secondIndex < 0 || tenthIndex < 0 {
		t.Fatalf("saved JSON does not contain all expected IDs: %s", text)
	}
	if !(firstIndex < secondIndex && secondIndex < tenthIndex) {
		t.Fatalf("saved JSON IDs are not in numeric order: %s", text)
	}
}

func TestAccountsJSONKeysMatchGoFieldNames(t *testing.T) {
	accounts := NewAccounts()
	if _, err := accounts.Add(&Account{Email: "pilot@example.com", Nickname: "Pilot", PasswordHash: "hash"}); err != nil {
		t.Fatalf("Add returned error: %v", err)
	}

	content, err := json.Marshal(accounts)
	if err != nil {
		t.Fatalf("Marshal returned error: %v", err)
	}
	text := string(content)

	expectedKeys := []string{
		`"MaxID"`,
		`"Items"`,
		`"ID"`,
		`"Email"`,
		`"Nickname"`,
		`"PasswordHash"`,
		`"Token"`,
		`"RegistrationTime"`,
		`"CurrentCharacterID"`,
	}
	for _, expectedKey := range expectedKeys {
		if !strings.Contains(text, expectedKey) {
			t.Fatalf("JSON %s does not contain key %s", text, expectedKey)
		}
	}
}

func TestAccountsRebuildIndexesRejectsStoredAccountWithoutToken(t *testing.T) {
	accounts := NewAccounts()
	accounts.Items[1] = &Account{
		ID:           1,
		Email:        "pilot@example.com",
		Nickname:     "Pilot",
		PasswordHash: "hash",
	}

	if err := accounts.RebuildIndexes(); err == nil {
		t.Fatal("RebuildIndexes accepted account without token")
	}
}

func TestAccountsRebuildIndexesRejectsDuplicateCurrentCharacterID(t *testing.T) {
	accounts := NewAccounts()
	accounts.Items[1] = &Account{
		ID:                 1,
		Email:              "pilot@example.com",
		Nickname:           "Pilot",
		PasswordHash:       "hash",
		Token:              "token-1",
		CurrentCharacterID: 10,
	}
	accounts.Items[2] = &Account{
		ID:                 2,
		Email:              "captain@example.com",
		Nickname:           "Captain",
		PasswordHash:       "hash",
		Token:              "token-2",
		CurrentCharacterID: 10,
	}

	if err := accounts.RebuildIndexes(); err == nil {
		t.Fatal("RebuildIndexes accepted duplicate CurrentCharacterID")
	}
}
