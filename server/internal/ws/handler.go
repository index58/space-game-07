package ws

import (
	"errors"
	"net/http"
	"strings"

	"github.com/gorilla/websocket"
	"space-game-07-server/internal/data"
)

var errUnauthorized = errors.New("unauthorized")

// РҐСЂР°РЅРёС‚ РёС‚РѕРі РїСЂРѕРІРµСЂРєРё РІС…РѕРґСЏС‰РµРіРѕ Р·Р°РїСЂРѕСЃР°.
type AuthResult struct {
	Account  *data.Account // РЈС‡РµС‚РЅР°СЏ Р·Р°РїРёСЃСЊ, СЂР°Р·СЂРµС€РµРЅРЅР°СЏ РґР»СЏ РїРѕРґРєР»СЋС‡РµРЅРёСЏ.
	NewToken string        // РќРѕРІС‹Р№ СЃРµРєСЂРµС‚, РєРѕС‚РѕСЂС‹Р№ РєР»РёРµРЅС‚ РґРѕР»Р¶РµРЅ СЃРѕС…СЂР°РЅРёС‚СЊ.
}

// РђРІС‚РѕСЂРёР·СѓРµС‚ WebSocket-Р·Р°РїСЂРѕСЃС‹ Рё РїРµСЂРµРґР°РµС‚ СѓСЃРїРµС€РЅС‹Рµ РїРѕРґРєР»СЋС‡РµРЅРёСЏ РІ Hub.
type Handler struct {
	hub      *Hub               // Р”РёСЃРїРµС‚С‡РµСЂ, РєРѕС‚РѕСЂРѕРјСѓ РїРµСЂРµРґР°СЋС‚СЃСЏ СѓСЃРїРµС€РЅС‹Рµ РїРѕРґРєР»СЋС‡РµРЅРёСЏ.
	accounts *data.Accounts     // РҐСЂР°РЅРёР»РёС‰Рµ Р°РєРєР°СѓРЅС‚РѕРІ РґР»СЏ РїСЂРѕРІРµСЂРєРё Р°РІС‚РѕСЂРёР·Р°С†РёРё.
	upgrader websocket.Upgrader // РќР°СЃС‚СЂРѕР№РєРё РїРѕРІС‹С€РµРЅРёСЏ HTTP-Р·Р°РїСЂРѕСЃР° РґРѕ WebSocket.
}

// РќР°СЃС‚СЂР°РёРІР°РµС‚ РѕР±СЂР°Р±РѕС‚С‡РёРє СЃ Р»РѕРєР°Р»СЊРЅС‹РјРё origin-РїСЂР°РІРёР»Р°РјРё РґР»СЏ Р±СЂР°СѓР·РµСЂРЅРѕРіРѕ РєР»РёРµРЅС‚Р°.
func NewHandler(hub *Hub, accounts *data.Accounts) *Handler {
	return &Handler{
		hub:      hub,
		accounts: accounts,
		upgrader: websocket.Upgrader{
			CheckOrigin: func(request *http.Request) bool {
				origin := request.Header.Get("Origin")
				return origin == "" ||
					strings.HasPrefix(origin, "http://127.0.0.1:") ||
					strings.HasPrefix(origin, "http://localhost:")
			},
		},
	}
}

// РџСЂРѕРІРµСЂСЏРµС‚ Р°РєРєР°СѓРЅС‚, РїРѕРІС‹С€Р°РµС‚ HTTP-Р·Р°РїСЂРѕСЃ РґРѕ WebSocket Рё СЂРµРіРёСЃС‚СЂРёСЂСѓРµС‚ СЃРѕРµРґРёРЅРµРЅРёРµ.
func (handler *Handler) ServeHTTP(writer http.ResponseWriter, request *http.Request) {
	result, err := handler.authorizeRequest(request)
	if err != nil {
		http.Error(writer, "unauthorized", http.StatusUnauthorized)
		return
	}

	connection, err := handler.upgrader.Upgrade(writer, request, nil)
	if err != nil {
		return
	}

	var initialMessages [][]byte
	if result.NewToken != "" {
		payload, err := EncodeAuthMessage(result.NewToken)
		if err != nil {
			_ = connection.Close()
			return
		}
		initialMessages = append(initialMessages, payload)
	}

	handler.hub.AddConnection(connection, result.Account.ID, initialMessages...)
}

// РџСЂРѕРІРµСЂСЏРµС‚ СЃРµРєСЂРµС‚ РёР»Рё СЃРѕР·РґР°РµС‚ РїРѕР»РЅС‹Р№ СЃС‚Р°СЂС‚РѕРІС‹Р№ РЅР°Р±РѕСЂ РґР»СЏ РЅРѕРІРѕРіРѕ РєР»РёРµРЅС‚Р°.
func (handler *Handler) authorizeRequest(request *http.Request) (AuthResult, error) {
	token := request.URL.Query().Get("token")
	if token == "" {
		if cookie, err := request.Cookie("Token"); err == nil {
			token = cookie.Value
		}
	}
	if token != "" {
		account, ok := handler.accounts.GetByToken(token)
		if !ok {
			return AuthResult{}, errUnauthorized
		}
		return AuthResult{Account: account}, nil
	}

	if handler.hub == nil {
		return AuthResult{}, errUnauthorized
	}
	account, err := handler.hub.world.CreateStarterAccount()
	if err != nil {
		return AuthResult{}, err
	}
	return AuthResult{Account: account, NewToken: account.Token}, nil
}

// РС‰РµС‚ СѓС‡РµС‚РЅСѓСЋ Р·Р°РїРёСЃСЊ РїРѕ СЃРµРєСЂРµС‚Сѓ РёР· Р·Р°РїСЂРѕСЃР° РёР»Рё cookie.
func (handler *Handler) accountByRequestToken(request *http.Request) (*data.Account, bool) {
	token := request.URL.Query().Get("token")
	if token == "" {
		if cookie, err := request.Cookie("Token"); err == nil {
			token = cookie.Value
		}
	}
	if token != "" {
		return handler.accounts.GetByToken(token)
	}
	return nil, false
}
