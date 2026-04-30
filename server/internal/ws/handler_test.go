package ws

import (
	"net/http/httptest"
	"testing"

	"space-game-07-server/internal/data"
)

// Готовит индексированный набор аккаунтов для проверки авторизации обработчика.
func testAccounts(t *testing.T) *data.Accounts {
	t.Helper()

	accounts := &data.Accounts{
		MaxID: 1,
		Items: map[int64]*data.Account{
			1: {ID: 1, Email: "index@email.net", Nickname: "index", PasswordHash: "hash", Token: "token"},
		},
	}
	if err := accounts.RebuildIndexes(); err != nil {
		t.Fatal(err)
	}
	return accounts
}

func TestHandlerFindsAccountByToken(t *testing.T) {
	handler := NewHandler(nil, testAccounts(t))
	request := httptest.NewRequest("GET", "/ws?token=token", nil)

	account, ok := handler.accountByRequestToken(request)
	if !ok || account.Nickname != "index" {
		t.Fatalf("account was not found by token")
	}
}

func TestHandlerFindsAccountByNicknameWhenTokenIsEmpty(t *testing.T) {
	handler := NewHandler(nil, testAccounts(t))
	request := httptest.NewRequest("GET", "/ws?nickname=index", nil)

	account, ok := handler.accountByRequestToken(request)
	if !ok || account.Token != "token" {
		t.Fatalf("account was not found by nickname")
	}
}
