package data

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"time"
)

const (
	accountTokenByteCount        = 32
	accountPasswordSaltByteCount = 16
)

// Хранит данные одного аккаунта игрового мира.
type Account struct {
	ID                 int64     `json:"ID"`                 // Уникальный числовой идентификатор записи.
	Email              string    `json:"Email"`              // Адрес электронной почты для входа и восстановления доступа.
	Nickname           string    `json:"Nickname"`           // Отображаемое имя игрока в игровом мире.
	PasswordHash       string    `json:"PasswordHash"`       // Хеш пароля без хранения исходного секрета.
	Token              string    `json:"Token"`              // Секрет для автоматической авторизации клиента.
	RegistrationTime   time.Time `json:"RegistrationTime"`   // Момент создания учетной записи.
	CurrentCharacterID int64     `json:"CurrentCharacterID"` // Активный персонаж, которым сейчас играет аккаунт.
}

// Хранит аккаунты и быстрые индексы для поиска по уникальным полям.
type Accounts struct {
	MaxID int64              `json:"MaxID"` // Последний выданный идентификатор для новых записей.
	Items map[int64]*Account `json:"Items"` // Основное хранилище записей по числовому идентификатору.

	ByEmail              map[string]*Account `json:"-"` // Быстрый поиск записи по адресу электронной почты.
	ByNickname           map[string]*Account `json:"-"` // Быстрый поиск записи по имени игрока.
	ByToken              map[string]*Account `json:"-"` // Быстрый поиск записи по секрету авторизации.
	ByCurrentCharacterID map[int64]*Account  `json:"-"` // Быстрый поиск записи по активному персонажу.
}

// Создаёт пустое хранилище аккаунтов с подготовленными индексами.
func NewAccounts() *Accounts {
	accounts := &Accounts{}
	accounts.ensureMaps()
	return accounts
}

// Добавляет новый аккаунт, назначает новый ID и генерирует уникальный токен.
func (accounts *Accounts) Add(account *Account) (*Account, error) {
	if account == nil {
		return nil, errors.New("account is nil")
	}
	accounts.ensureMaps()
	if err := accounts.validateRequiredFields(account); err != nil {
		return nil, err
	}
	if err := accounts.ensureUniqueForNewAccount(account); err != nil {
		return nil, err
	}

	token, err := accounts.generateUniqueToken()
	if err != nil {
		return nil, err
	}

	accounts.MaxID++
	account.ID = accounts.MaxID
	account.Token = token
	if account.RegistrationTime.IsZero() {
		account.RegistrationTime = time.Now()
	}

	accounts.Items[account.ID] = account
	accounts.addIndexes(account)
	return account, nil
}

// Возвращает аккаунт по ID.
func (accounts *Accounts) Get(id int64) (*Account, bool) {
	accounts.ensureMaps()
	account, ok := accounts.Items[id]
	return account, ok
}

// Удаляет аккаунт и все его быстрые индексы.
func (accounts *Accounts) Delete(id int64) bool {
	accounts.ensureMaps()
	account, ok := accounts.Items[id]
	if !ok {
		return false
	}

	accounts.deleteIndexes(account)
	delete(accounts.Items, id)
	return true
}

// Меняет e-mail аккаунта и обновляет индекс уникальности.
func (accounts *Accounts) SetEmail(id int64, email string) error {
	accounts.ensureMaps()
	if email == "" {
		return errors.New("email is empty")
	}

	account, ok := accounts.Items[id]
	if !ok {
		return fmt.Errorf("account with ID %d not found", id)
	}
	if existing, ok := accounts.ByEmail[email]; ok && existing.ID != id {
		return fmt.Errorf("email %q already exists", email)
	}

	delete(accounts.ByEmail, account.Email)
	account.Email = email
	accounts.ByEmail[email] = account
	return nil
}

// Меняет никнейм аккаунта и обновляет индекс уникальности.
func (accounts *Accounts) SetNickname(id int64, nickname string) error {
	accounts.ensureMaps()
	if nickname == "" {
		return errors.New("nickname is empty")
	}

	account, ok := accounts.Items[id]
	if !ok {
		return fmt.Errorf("account with ID %d not found", id)
	}
	if existing, ok := accounts.ByNickname[nickname]; ok && existing.ID != id {
		return fmt.Errorf("nickname %q already exists", nickname)
	}

	delete(accounts.ByNickname, account.Nickname)
	account.Nickname = nickname
	accounts.ByNickname[nickname] = account
	return nil
}

// Хеширует пароль и сохраняет только хеш.
func (accounts *Accounts) SetPassword(id int64, password string) error {
	accounts.ensureMaps()
	if password == "" {
		return errors.New("password is empty")
	}

	account, ok := accounts.Items[id]
	if !ok {
		return fmt.Errorf("account with ID %d not found", id)
	}

	passwordHash, err := hashAccountPassword(password)
	if err != nil {
		return err
	}
	account.PasswordHash = passwordHash
	return nil
}

// Создаёт новый уникальный токен аккаунта и обновляет индекс токенов.
func (accounts *Accounts) GenerateToken(id int64) (string, error) {
	accounts.ensureMaps()
	account, ok := accounts.Items[id]
	if !ok {
		return "", fmt.Errorf("account with ID %d not found", id)
	}

	token, err := accounts.generateUniqueToken()
	if err != nil {
		return "", err
	}

	delete(accounts.ByToken, account.Token)
	account.Token = token
	accounts.ByToken[token] = account
	return token, nil
}

// Меняет активного персонажа и поддерживает индекс уникальности.
func (accounts *Accounts) SetCurrentCharacter(id int64, characterID int64) error {
	accounts.ensureMaps()

	account, ok := accounts.Items[id]
	if !ok {
		return fmt.Errorf("account with ID %d not found", id)
	}
	if characterID > 0 {
		if existing, ok := accounts.ByCurrentCharacterID[characterID]; ok && existing.ID != id {
			return fmt.Errorf("current character ID %d already exists", characterID)
		}
	}

	if account.CurrentCharacterID > 0 {
		delete(accounts.ByCurrentCharacterID, account.CurrentCharacterID)
	}
	account.CurrentCharacterID = characterID
	if characterID > 0 {
		accounts.ByCurrentCharacterID[characterID] = account
	}
	return nil
}

// Возвращает аккаунт по уникальному e-mail.
func (accounts *Accounts) GetByEmail(email string) (*Account, bool) {
	accounts.ensureMaps()
	account, ok := accounts.ByEmail[email]
	return account, ok
}

// Возвращает аккаунт по уникальному никнейму.
func (accounts *Accounts) GetByNickname(nickname string) (*Account, bool) {
	accounts.ensureMaps()
	account, ok := accounts.ByNickname[nickname]
	return account, ok
}

// Возвращает аккаунт по уникальному токену.
func (accounts *Accounts) GetByToken(token string) (*Account, bool) {
	accounts.ensureMaps()
	account, ok := accounts.ByToken[token]
	return account, ok
}

// Возвращает запись по активному персонажу.
func (accounts *Accounts) GetByCurrentCharacterID(characterID int64) (*Account, bool) {
	accounts.ensureMaps()
	account, ok := accounts.ByCurrentCharacterID[characterID]
	return account, ok
}

// Пересобирает быстрые индексы после загрузки из JSON.
func (accounts *Accounts) RebuildIndexes() error {
	accounts.ensureItems()
	accounts.ByEmail = make(map[string]*Account)
	accounts.ByNickname = make(map[string]*Account)
	accounts.ByToken = make(map[string]*Account)
	accounts.ByCurrentCharacterID = make(map[int64]*Account)

	var maxID int64
	for id, account := range accounts.Items {
		if account == nil {
			return fmt.Errorf("account with ID %d is nil", id)
		}
		if account.ID != id {
			return fmt.Errorf("account map key %d does not match account ID %d", id, account.ID)
		}
		if err := accounts.validateStoredAccount(account); err != nil {
			return fmt.Errorf("account with ID %d is invalid: %w", id, err)
		}
		if id > maxID {
			maxID = id
		}
		if err := accounts.ensureUniqueForNewAccount(account); err != nil {
			return err
		}
		accounts.addIndexes(account)
	}
	if accounts.MaxID < maxID {
		accounts.MaxID = maxID
	}
	return nil
}

// Загружает аккаунты из JSON-файла и пересобирает быстрые индексы.
func (accounts *Accounts) LoadFromFile(path string) error {
	content, err := os.ReadFile(path)
	if err != nil {
		return err
	}

	loaded := Accounts{}
	if err := json.Unmarshal(content, &loaded); err != nil {
		return err
	}
	if err := loaded.RebuildIndexes(); err != nil {
		return err
	}

	*accounts = loaded
	return nil
}

// Сохраняет аккаунты в JSON-файл без вспомогательных индексов.
func (accounts *Accounts) SaveToFile(path string) error {
	accounts.ensureMaps()
	return saveTableWithOrderedItems(path, accounts.MaxID, accounts.Items)
}

// Подготавливает основное хранилище и все индексы.
func (accounts *Accounts) ensureMaps() {
	accounts.ensureItems()
	if accounts.ByEmail == nil {
		accounts.ByEmail = make(map[string]*Account)
	}
	if accounts.ByNickname == nil {
		accounts.ByNickname = make(map[string]*Account)
	}
	if accounts.ByToken == nil {
		accounts.ByToken = make(map[string]*Account)
	}
	if accounts.ByCurrentCharacterID == nil {
		accounts.ByCurrentCharacterID = make(map[int64]*Account)
	}
}

// Подготавливает основную map аккаунтов.
func (accounts *Accounts) ensureItems() {
	if accounts.Items == nil {
		accounts.Items = make(map[int64]*Account)
	}
}

// Проверяет обязательные поля аккаунта.
func (accounts *Accounts) validateRequiredFields(account *Account) error {
	if account.Email == "" {
		return errors.New("email is empty")
	}
	if account.Nickname == "" {
		return errors.New("nickname is empty")
	}
	if account.PasswordHash == "" {
		return errors.New("password hash is empty")
	}
	return nil
}

// Проверяет обязательные поля уже сохранённого аккаунта.
func (accounts *Accounts) validateStoredAccount(account *Account) error {
	if err := accounts.validateRequiredFields(account); err != nil {
		return err
	}
	if account.Token == "" {
		return errors.New("token is empty")
	}
	return nil
}

// Проверяет уникальные поля перед добавлением в индексы.
func (accounts *Accounts) ensureUniqueForNewAccount(account *Account) error {
	if existing, ok := accounts.ByEmail[account.Email]; ok && existing.ID != account.ID {
		return fmt.Errorf("email %q already exists", account.Email)
	}
	if existing, ok := accounts.ByNickname[account.Nickname]; ok && existing.ID != account.ID {
		return fmt.Errorf("nickname %q already exists", account.Nickname)
	}
	if account.Token != "" {
		if existing, ok := accounts.ByToken[account.Token]; ok && existing.ID != account.ID {
			return fmt.Errorf("token %q already exists", account.Token)
		}
	}
	if account.CurrentCharacterID > 0 {
		if existing, ok := accounts.ByCurrentCharacterID[account.CurrentCharacterID]; ok && existing.ID != account.ID {
			return fmt.Errorf("current character ID %d already exists", account.CurrentCharacterID)
		}
	}
	return nil
}

// Добавляет аккаунт во все быстрые индексы.
func (accounts *Accounts) addIndexes(account *Account) {
	accounts.ByEmail[account.Email] = account
	accounts.ByNickname[account.Nickname] = account
	accounts.ByToken[account.Token] = account
	if account.CurrentCharacterID > 0 {
		accounts.ByCurrentCharacterID[account.CurrentCharacterID] = account
	}
}

// Удаляет аккаунт из всех быстрых индексов.
func (accounts *Accounts) deleteIndexes(account *Account) {
	delete(accounts.ByEmail, account.Email)
	delete(accounts.ByNickname, account.Nickname)
	delete(accounts.ByToken, account.Token)
	if account.CurrentCharacterID > 0 {
		delete(accounts.ByCurrentCharacterID, account.CurrentCharacterID)
	}
}

// Создаёт криптостойкий токен, которого ещё нет в индексе.
func (accounts *Accounts) generateUniqueToken() (string, error) {
	for {
		token, err := randomHex(accountTokenByteCount)
		if err != nil {
			return "", err
		}
		if _, exists := accounts.ByToken[token]; !exists {
			return token, nil
		}
	}
}

// Создаёт salted SHA-256 хеш пароля.
func hashAccountPassword(password string) (string, error) {
	salt, err := randomHex(accountPasswordSaltByteCount)
	if err != nil {
		return "", err
	}
	sum := sha256.Sum256([]byte(salt + password))
	return "sha256$" + salt + "$" + hex.EncodeToString(sum[:]), nil
}

// Возвращает криптостойкую случайную строку в hex-представлении.
func randomHex(byteCount int) (string, error) {
	buffer := make([]byte, byteCount)
	if _, err := rand.Read(buffer); err != nil {
		return "", err
	}
	return hex.EncodeToString(buffer), nil
}
