package ws

import (
	"errors"
	"net/http"
	"strings"

	"github.com/gorilla/websocket"
	"space-game-07-server/internal/data"
)

var errUnauthorized = errors.New("unauthorized")

// Хранит итог проверки входящего запроса.
type AuthResult struct {
	Account  *data.Account // Учетная запись, разрешенная для подключения.
	NewToken string        // Новый секрет, который клиент должен сохранить.
}

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

// Проверяет секрет или создает полный стартовый набор для нового клиента.
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

// Ищет учетную запись по секрету из запроса или cookie.
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
