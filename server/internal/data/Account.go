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

// РҐСЂР°РЅРёС‚ РґР°РЅРЅС‹Рµ РѕРґРЅРѕРіРѕ Р°РєРєР°СѓРЅС‚Р° РёРіСЂРѕРІРѕРіРѕ РјРёСЂР°.
type Account struct {
	ID                 int64     `json:"ID"`                 // РЈРЅРёРєР°Р»СЊРЅС‹Р№ С‡РёСЃР»РѕРІРѕР№ РёРґРµРЅС‚РёС„РёРєР°С‚РѕСЂ Р·Р°РїРёСЃРё.
	Email              string    `json:"Email"`              // РђРґСЂРµСЃ СЌР»РµРєС‚СЂРѕРЅРЅРѕР№ РїРѕС‡С‚С‹ РґР»СЏ РІС…РѕРґР° Рё РІРѕСЃСЃС‚Р°РЅРѕРІР»РµРЅРёСЏ РґРѕСЃС‚СѓРїР°.
	Nickname           string    `json:"Nickname"`           // РћС‚РѕР±СЂР°Р¶Р°РµРјРѕРµ РёРјСЏ РёРіСЂРѕРєР° РІ РёРіСЂРѕРІРѕРј РјРёСЂРµ.
	PasswordHash       string    `json:"PasswordHash"`       // РҐРµС€ РїР°СЂРѕР»СЏ Р±РµР· С…СЂР°РЅРµРЅРёСЏ РёСЃС…РѕРґРЅРѕРіРѕ СЃРµРєСЂРµС‚Р°.
	Token              string    `json:"Token"`              // РЎРµРєСЂРµС‚ РґР»СЏ Р°РІС‚РѕРјР°С‚РёС‡РµСЃРєРѕР№ Р°РІС‚РѕСЂРёР·Р°С†РёРё РєР»РёРµРЅС‚Р°.
	RegistrationTime   time.Time `json:"RegistrationTime"`   // РњРѕРјРµРЅС‚ СЃРѕР·РґР°РЅРёСЏ СѓС‡РµС‚РЅРѕР№ Р·Р°РїРёСЃРё.
	CurrentCharacterID int64     `json:"CurrentCharacterID"` // РђРєС‚РёРІРЅС‹Р№ РїРµСЂСЃРѕРЅР°Р¶, РєРѕС‚РѕСЂС‹Рј СЃРµР№С‡Р°СЃ РёРіСЂР°РµС‚ Р°РєРєР°СѓРЅС‚.
}

// РҐСЂР°РЅРёС‚ Р°РєРєР°СѓРЅС‚С‹ Рё Р±С‹СЃС‚СЂС‹Рµ РёРЅРґРµРєСЃС‹ РґР»СЏ РїРѕРёСЃРєР° РїРѕ СѓРЅРёРєР°Р»СЊРЅС‹Рј РїРѕР»СЏРј.
type Accounts struct {
	MaxID int64              `json:"MaxID"` // РџРѕСЃР»РµРґРЅРёР№ РІС‹РґР°РЅРЅС‹Р№ РёРґРµРЅС‚РёС„РёРєР°С‚РѕСЂ РґР»СЏ РЅРѕРІС‹С… Р·Р°РїРёСЃРµР№.
	Items map[int64]*Account `json:"Items"` // РћСЃРЅРѕРІРЅРѕРµ С…СЂР°РЅРёР»РёС‰Рµ Р·Р°РїРёСЃРµР№ РїРѕ С‡РёСЃР»РѕРІРѕРјСѓ РёРґРµРЅС‚РёС„РёРєР°С‚РѕСЂСѓ.

	ByEmail              map[string]*Account `json:"-"` // Р‘С‹СЃС‚СЂС‹Р№ РїРѕРёСЃРє Р·Р°РїРёСЃРё РїРѕ Р°РґСЂРµСЃСѓ СЌР»РµРєС‚СЂРѕРЅРЅРѕР№ РїРѕС‡С‚С‹.
	ByNickname           map[string]*Account `json:"-"` // Р‘С‹СЃС‚СЂС‹Р№ РїРѕРёСЃРє Р·Р°РїРёСЃРё РїРѕ РёРјРµРЅРё РёРіСЂРѕРєР°.
	ByToken              map[string]*Account `json:"-"` // Р‘С‹СЃС‚СЂС‹Р№ РїРѕРёСЃРє Р·Р°РїРёСЃРё РїРѕ СЃРµРєСЂРµС‚Сѓ Р°РІС‚РѕСЂРёР·Р°С†РёРё.
	ByCurrentCharacterID map[int64]*Account  `json:"-"` // Р‘С‹СЃС‚СЂС‹Р№ РїРѕРёСЃРє Р·Р°РїРёСЃРё РїРѕ Р°РєС‚РёРІРЅРѕРјСѓ РїРµСЂСЃРѕРЅР°Р¶Сѓ.
}

// РЎРѕР·РґР°С‘С‚ РїСѓСЃС‚РѕРµ С…СЂР°РЅРёР»РёС‰Рµ Р°РєРєР°СѓРЅС‚РѕРІ СЃ РїРѕРґРіРѕС‚РѕРІР»РµРЅРЅС‹РјРё РёРЅРґРµРєСЃР°РјРё.
func NewAccounts() *Accounts {
	accounts := &Accounts{}
	accounts.ensureMaps()
	return accounts
}

// Р”РѕР±Р°РІР»СЏРµС‚ РЅРѕРІС‹Р№ Р°РєРєР°СѓРЅС‚, РЅР°Р·РЅР°С‡Р°РµС‚ РЅРѕРІС‹Р№ ID Рё РіРµРЅРµСЂРёСЂСѓРµС‚ СѓРЅРёРєР°Р»СЊРЅС‹Р№ С‚РѕРєРµРЅ.
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

// Р’РѕР·РІСЂР°С‰Р°РµС‚ Р°РєРєР°СѓРЅС‚ РїРѕ ID.
func (accounts *Accounts) Get(id int64) (*Account, bool) {
	accounts.ensureMaps()
	account, ok := accounts.Items[id]
	return account, ok
}

// РЈРґР°Р»СЏРµС‚ Р°РєРєР°СѓРЅС‚ Рё РІСЃРµ РµРіРѕ Р±С‹СЃС‚СЂС‹Рµ РёРЅРґРµРєСЃС‹.
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

// РњРµРЅСЏРµС‚ e-mail Р°РєРєР°СѓРЅС‚Р° Рё РѕР±РЅРѕРІР»СЏРµС‚ РёРЅРґРµРєСЃ СѓРЅРёРєР°Р»СЊРЅРѕСЃС‚Рё.
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

// РњРµРЅСЏРµС‚ РЅРёРєРЅРµР№Рј Р°РєРєР°СѓРЅС‚Р° Рё РѕР±РЅРѕРІР»СЏРµС‚ РёРЅРґРµРєСЃ СѓРЅРёРєР°Р»СЊРЅРѕСЃС‚Рё.
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

// РҐРµС€РёСЂСѓРµС‚ РїР°СЂРѕР»СЊ Рё СЃРѕС…СЂР°РЅСЏРµС‚ С‚РѕР»СЊРєРѕ С…РµС€.
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

// РЎРѕР·РґР°С‘С‚ РЅРѕРІС‹Р№ СѓРЅРёРєР°Р»СЊРЅС‹Р№ С‚РѕРєРµРЅ Р°РєРєР°СѓРЅС‚Р° Рё РѕР±РЅРѕРІР»СЏРµС‚ РёРЅРґРµРєСЃ С‚РѕРєРµРЅРѕРІ.
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

// РњРµРЅСЏРµС‚ Р°РєС‚РёРІРЅРѕРіРѕ РїРµСЂСЃРѕРЅР°Р¶Р° Рё РїРѕРґРґРµСЂР¶РёРІР°РµС‚ РёРЅРґРµРєСЃ СѓРЅРёРєР°Р»СЊРЅРѕСЃС‚Рё.
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

// Р’РѕР·РІСЂР°С‰Р°РµС‚ Р°РєРєР°СѓРЅС‚ РїРѕ СѓРЅРёРєР°Р»СЊРЅРѕРјСѓ e-mail.
func (accounts *Accounts) GetByEmail(email string) (*Account, bool) {
	accounts.ensureMaps()
	account, ok := accounts.ByEmail[email]
	return account, ok
}

// Р’РѕР·РІСЂР°С‰Р°РµС‚ Р°РєРєР°СѓРЅС‚ РїРѕ СѓРЅРёРєР°Р»СЊРЅРѕРјСѓ РЅРёРєРЅРµР№РјСѓ.
func (accounts *Accounts) GetByNickname(nickname string) (*Account, bool) {
	accounts.ensureMaps()
	account, ok := accounts.ByNickname[nickname]
	return account, ok
}

// Р’РѕР·РІСЂР°С‰Р°РµС‚ Р°РєРєР°СѓРЅС‚ РїРѕ СѓРЅРёРєР°Р»СЊРЅРѕРјСѓ С‚РѕРєРµРЅСѓ.
func (accounts *Accounts) GetByToken(token string) (*Account, bool) {
	accounts.ensureMaps()
	account, ok := accounts.ByToken[token]
	return account, ok
}

// Р’РѕР·РІСЂР°С‰Р°РµС‚ Р·Р°РїРёСЃСЊ РїРѕ Р°РєС‚РёРІРЅРѕРјСѓ РїРµСЂСЃРѕРЅР°Р¶Сѓ.
func (accounts *Accounts) GetByCurrentCharacterID(characterID int64) (*Account, bool) {
	accounts.ensureMaps()
	account, ok := accounts.ByCurrentCharacterID[characterID]
	return account, ok
}

// РџРµСЂРµСЃРѕР±РёСЂР°РµС‚ Р±С‹СЃС‚СЂС‹Рµ РёРЅРґРµРєСЃС‹ РїРѕСЃР»Рµ Р·Р°РіСЂСѓР·РєРё РёР· JSON.
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

// Р—Р°РіСЂСѓР¶Р°РµС‚ Р°РєРєР°СѓРЅС‚С‹ РёР· JSON-С„Р°Р№Р»Р° Рё РїРµСЂРµСЃРѕР±РёСЂР°РµС‚ Р±С‹СЃС‚СЂС‹Рµ РёРЅРґРµРєСЃС‹.
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

// РЎРѕС…СЂР°РЅСЏРµС‚ Р°РєРєР°СѓРЅС‚С‹ РІ JSON-С„Р°Р№Р» Р±РµР· РІСЃРїРѕРјРѕРіР°С‚РµР»СЊРЅС‹С… РёРЅРґРµРєСЃРѕРІ.
func (accounts *Accounts) SaveToFile(path string) error {
	accounts.ensureMaps()
	return saveTableWithOrderedItems(path, accounts.MaxID, accounts.Items)
}

// РџРѕРґРіРѕС‚Р°РІР»РёРІР°РµС‚ РѕСЃРЅРѕРІРЅРѕРµ С…СЂР°РЅРёР»РёС‰Рµ Рё РІСЃРµ РёРЅРґРµРєСЃС‹.
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

// РџРѕРґРіРѕС‚Р°РІР»РёРІР°РµС‚ РѕСЃРЅРѕРІРЅСѓСЋ map Р°РєРєР°СѓРЅС‚РѕРІ.
func (accounts *Accounts) ensureItems() {
	if accounts.Items == nil {
		accounts.Items = make(map[int64]*Account)
	}
}

// РџСЂРѕРІРµСЂСЏРµС‚ РѕР±СЏР·Р°С‚РµР»СЊРЅС‹Рµ РїРѕР»СЏ Р°РєРєР°СѓРЅС‚Р°.
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

// РџСЂРѕРІРµСЂСЏРµС‚ РѕР±СЏР·Р°С‚РµР»СЊРЅС‹Рµ РїРѕР»СЏ СѓР¶Рµ СЃРѕС…СЂР°РЅС‘РЅРЅРѕРіРѕ Р°РєРєР°СѓРЅС‚Р°.
func (accounts *Accounts) validateStoredAccount(account *Account) error {
	if err := accounts.validateRequiredFields(account); err != nil {
		return err
	}
	if account.Token == "" {
		return errors.New("token is empty")
	}
	return nil
}

// РџСЂРѕРІРµСЂСЏРµС‚ СѓРЅРёРєР°Р»СЊРЅС‹Рµ РїРѕР»СЏ РїРµСЂРµРґ РґРѕР±Р°РІР»РµРЅРёРµРј РІ РёРЅРґРµРєСЃС‹.
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

// Р”РѕР±Р°РІР»СЏРµС‚ Р°РєРєР°СѓРЅС‚ РІРѕ РІСЃРµ Р±С‹СЃС‚СЂС‹Рµ РёРЅРґРµРєСЃС‹.
func (accounts *Accounts) addIndexes(account *Account) {
	accounts.ByEmail[account.Email] = account
	accounts.ByNickname[account.Nickname] = account
	accounts.ByToken[account.Token] = account
	if account.CurrentCharacterID > 0 {
		accounts.ByCurrentCharacterID[account.CurrentCharacterID] = account
	}
}

// РЈРґР°Р»СЏРµС‚ Р°РєРєР°СѓРЅС‚ РёР· РІСЃРµС… Р±С‹СЃС‚СЂС‹С… РёРЅРґРµРєСЃРѕРІ.
func (accounts *Accounts) deleteIndexes(account *Account) {
	delete(accounts.ByEmail, account.Email)
	delete(accounts.ByNickname, account.Nickname)
	delete(accounts.ByToken, account.Token)
	if account.CurrentCharacterID > 0 {
		delete(accounts.ByCurrentCharacterID, account.CurrentCharacterID)
	}
}

// РЎРѕР·РґР°С‘С‚ РєСЂРёРїС‚РѕСЃС‚РѕР№РєРёР№ С‚РѕРєРµРЅ, РєРѕС‚РѕСЂРѕРіРѕ РµС‰С‘ РЅРµС‚ РІ РёРЅРґРµРєСЃРµ.
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

// РЎРѕР·РґР°С‘С‚ salted SHA-256 С…РµС€ РїР°СЂРѕР»СЏ.
func hashAccountPassword(password string) (string, error) {
	salt, err := randomHex(accountPasswordSaltByteCount)
	if err != nil {
		return "", err
	}
	sum := sha256.Sum256([]byte(salt + password))
	return "sha256$" + salt + "$" + hex.EncodeToString(sum[:]), nil
}

// Р’РѕР·РІСЂР°С‰Р°РµС‚ РєСЂРёРїС‚РѕСЃС‚РѕР№РєСѓСЋ СЃР»СѓС‡Р°Р№РЅСѓСЋ СЃС‚СЂРѕРєСѓ РІ hex-РїСЂРµРґСЃС‚Р°РІР»РµРЅРёРё.
func randomHex(byteCount int) (string, error) {
	buffer := make([]byte, byteCount)
	if _, err := rand.Read(buffer); err != nil {
		return "", err
	}
	return hex.EncodeToString(buffer), nil
}
