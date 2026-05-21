package ws

import (
	"net/http/httptest"
	"testing"

	"space-game-07-server/internal/data"
	"space-game-07-server/internal/world"
)

// Р“РѕС‚РѕРІРёС‚ РёРЅРґРµРєСЃРёСЂРѕРІР°РЅРЅС‹Р№ РЅР°Р±РѕСЂ Р°РєРєР°СѓРЅС‚РѕРІ РґР»СЏ РїСЂРѕРІРµСЂРєРё Р°РІС‚РѕСЂРёР·Р°С†РёРё РѕР±СЂР°Р±РѕС‚С‡РёРєР°.
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

// РџСЂРѕРІРµСЂСЏРµС‚, С‡С‚Рѕ Р·Р°РїСЂРѕСЃ СЃ РёР·РІРµСЃС‚РЅС‹Рј С‚РѕРєРµРЅРѕРј СЃРѕРїРѕСЃС‚Р°РІР»СЏРµС‚СЃСЏ СЃ СЃСѓС‰РµСЃС‚РІСѓСЋС‰РёРј Р°РєРєР°СѓРЅС‚РѕРј.
func TestHandlerFindsAccountByToken(t *testing.T) {
	handler := NewHandler(nil, testAccounts(t))
	request := httptest.NewRequest("GET", "/ws?token=token", nil)

	account, ok := handler.accountByRequestToken(request)
	if !ok || account.Nickname != "index" {
		t.Fatalf("account was not found by token")
	}
}

// РџСЂРѕРІРµСЂСЏРµС‚, С‡С‚Рѕ РЅРµРёР·РІРµСЃС‚РЅС‹Р№ С‚РѕРєРµРЅ РѕС‚РєР»РѕРЅСЏРµС‚СЃСЏ Р±РµР· СЃРѕР·РґР°РЅРёСЏ Р°РєРєР°СѓРЅС‚Р° Рё РЅРѕРІРѕРіРѕ С‚РѕРєРµРЅР°.
func TestHandlerRejectsUnknownToken(t *testing.T) {
	handler := NewHandler(nil, testAccounts(t))
	request := httptest.NewRequest("GET", "/ws?token=unknown", nil)

	result, err := handler.authorizeRequest(request)
	if err == nil {
		t.Fatal("unknown token was accepted")
	}
	if result.Account != nil || result.NewToken != "" {
		t.Fatal("unknown token created or returned an account")
	}
}

// РџСЂРѕРІРµСЂСЏРµС‚, С‡С‚Рѕ Р·Р°РїСЂРѕСЃ Р±РµР· С‚РѕРєРµРЅР° СЃРѕР·РґР°С‘С‚ СЃС‚Р°СЂС‚РѕРІС‹Р№ Р°РєРєР°СѓРЅС‚, РїРµСЂСЃРѕРЅР°Р¶Р° Рё РєРѕСЂР°Р±Р»СЊ.
func TestHandlerCreatesStarterAccountWhenTokenIsMissing(t *testing.T) {
	accounts := data.NewAccounts()
	handler := NewHandler(NewHub(testWorld(t, accounts)), accounts)
	request := httptest.NewRequest("GET", "/ws", nil)

	result, err := handler.authorizeRequest(request)
	if err != nil {
		t.Fatalf("authorizeRequest returned error: %v", err)
	}

	if result.Account == nil {
		t.Fatal("account was not created")
	}
	if result.NewToken == "" || result.NewToken != result.Account.Token {
		t.Fatalf("new token = %q, account token = %q", result.NewToken, result.Account.Token)
	}
	if result.Account.CurrentCharacterID <= 0 {
		t.Fatal("current character was not selected")
	}

	character, ok := handler.hub.world.CharacterByID(result.Account.CurrentCharacterID)
	if !ok || character.AccountID != result.Account.ID {
		t.Fatal("starter character does not belong to created account")
	}
	if character.LocationCosmicObjectID <= 0 {
		t.Fatal("starter character has no location")
	}

	cosmicObject, ok := handler.hub.world.CosmicObjectByID(character.LocationCosmicObjectID)
	if !ok || cosmicObject.OwnerCharacterID != character.ID {
		t.Fatal("starter ship does not belong to created character")
	}
}

func testWorld(t *testing.T, accounts *data.Accounts) *world.World {
	t.Helper()

	models := data.NewCosmicObjectModels()
	if _, err := models.Add(&data.CosmicObjectModel{
		TitleRu:            "РЎС‚Р°СЂС‚РѕРІС‹Р№ РєРѕСЂР°Р±Р»СЊ",
		TitleEn:            "Starter Ship",
		Acronym:            "ship_bat",
		TextureScale:       4,
		CosmicObjectTypeID: 1,
		Mass:               7920,
		MaxArmor:           100,
		MaxSpeed:           497,
		MaxAngularSpeed:    3,
	}); err != nil {
		t.Fatalf("Add model returned error: %v", err)
	}

	assemblies := data.NewAssemblies()
	if _, err := assemblies.Add(&data.Assembly{
		Title:               "Starter Assembly",
		CosmicObjectModelID: 1,
		IsPublic:            true,
		Mass:                9000,
		MaxArmor:            100,
		MaxAlongForce:       900000,
		MaxAcrossForce:      900000,
		MaxTorque:           900000,
	}); err != nil {
		t.Fatalf("Add assembly returned error: %v", err)
	}

	return world.New(1, world.Data{
		Accounts:           accounts,
		Characters:         data.NewCharacters(),
		CosmicObjects:      data.NewCosmicObjects(),
		CosmicObjectModels: models,
		Assemblies:         assemblies,
	})
}
