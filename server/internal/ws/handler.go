package ws

import (
	"net/http"
	"strings"

	"github.com/gorilla/websocket"
	"space-game-07-server/internal/data"
)

// Авторизует WebSocket-запросы и передает успешные подключения в Hub.
type Handler struct {
	hub      *Hub               // Диспетчер, которому передаются успешные подключения.
	accounts *data.Accounts     // Хранилище аккаунтов для проверки авторизации.
	upgrader websocket.Upgrader // Настройки повышения HTTP-запроса до WebSocket.
}

// Настраивает обработчик с локальными origin-правилами для браузерного клиента.
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

// Проверяет аккаунт, повышает HTTP-запрос до WebSocket и регистрирует соединение.
func (handler *Handler) ServeHTTP(writer http.ResponseWriter, request *http.Request) {
	account, ok := handler.accountByRequestToken(request)
	if !ok {
		http.Error(writer, "unauthorized", http.StatusUnauthorized)
		return
	}

	connection, err := handler.upgrader.Upgrade(writer, request, nil)
	if err != nil {
		return
	}

	handler.hub.AddConnection(connection, account.ID)
}

// Ищет аккаунт по токену, cookie или никнейму для локальной разработки.
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

	nickname := request.URL.Query().Get("nickname")
	if nickname == "" {
		return nil, false
	}
	return handler.accounts.GetByNickname(nickname)
}
