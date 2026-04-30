package data

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"sort"
	"time"
)

// Хранит данные одного персонажа игрового мира.
type Character struct {
	ID                     int64     `json:"ID"`
	AccountID              int64     `json:"AccountID"`
	CreationTime           time.Time `json:"CreationTime"`
	Balance                int64     `json:"Balance"`
	LocationCosmicObjectID int64     `json:"LocationCosmicObjectID"`
}

// Хранит персонажей и быстрые индексы для поиска по связанным объектам.
type Characters struct {
	MaxID int64                `json:"MaxID"`
	Items map[int64]*Character `json:"Items"`

	ByAccountID map[int64]map[int64]*Character `json:"-"`
}

// Создаёт пустое хранилище персонажей с подготовленными индексами.
func NewCharacters() *Characters {
	characters := &Characters{}
	characters.ensureMaps()
	return characters
}

// Добавляет нового персонажа, назначает новый ID и время создания.
func (characters *Characters) Add(character *Character) (*Character, error) {
	if character == nil {
		return nil, errors.New("character is nil")
	}
	characters.ensureMaps()
	if err := characters.validateRequiredFields(character); err != nil {
		return nil, err
	}

	characters.MaxID++
	character.ID = characters.MaxID
	if character.CreationTime.IsZero() {
		character.CreationTime = time.Now()
	}

	characters.Items[character.ID] = character
	characters.addIndexes(character)
	return character, nil
}

// Возвращает персонажа по ID.
func (characters *Characters) Get(id int64) (*Character, bool) {
	characters.ensureMaps()
	character, ok := characters.Items[id]
	return character, ok
}

// Удаляет персонажа и все его быстрые индексы.
func (characters *Characters) Delete(id int64) bool {
	characters.ensureMaps()
	character, ok := characters.Items[id]
	if !ok {
		return false
	}

	characters.deleteIndexes(character)
	delete(characters.Items, id)
	return true
}

// Возвращает персонажей указанного аккаунта в порядке возрастания ID.
func (characters *Characters) GetByAccountID(accountID int64) []*Character {
	characters.ensureMaps()
	indexItems := characters.ByAccountID[accountID]
	if len(indexItems) == 0 {
		return []*Character{}
	}

	ids := make([]int64, 0, len(indexItems))
	for id := range indexItems {
		ids = append(ids, id)
	}
	sort.Slice(ids, func(left int, right int) bool {
		return ids[left] < ids[right]
	})

	result := make([]*Character, 0, len(ids))
	for _, id := range ids {
		result = append(result, indexItems[id])
	}
	return result
}

// Пересобирает быстрые индексы после загрузки из JSON.
func (characters *Characters) RebuildIndexes() error {
	characters.ensureItems()
	characters.ByAccountID = make(map[int64]map[int64]*Character)

	var maxID int64
	for id, character := range characters.Items {
		if character == nil {
			return fmt.Errorf("character with ID %d is nil", id)
		}
		if character.ID != id {
			return fmt.Errorf("character map key %d does not match character ID %d", id, character.ID)
		}
		if err := characters.validateRequiredFields(character); err != nil {
			return fmt.Errorf("character with ID %d is invalid: %w", id, err)
		}
		if id > maxID {
			maxID = id
		}
		characters.addIndexes(character)
	}
	if characters.MaxID < maxID {
		characters.MaxID = maxID
	}
	return nil
}

// Загружает персонажей из JSON-файла и пересобирает быстрые индексы.
func (characters *Characters) LoadFromFile(path string) error {
	content, err := os.ReadFile(path)
	if err != nil {
		return err
	}

	loaded := Characters{}
	if err := json.Unmarshal(content, &loaded); err != nil {
		return err
	}
	if err := loaded.RebuildIndexes(); err != nil {
		return err
	}

	*characters = loaded
	return nil
}

// Сохраняет персонажей в JSON-файл без вспомогательных индексов.
func (characters *Characters) SaveToFile(path string) error {
	characters.ensureMaps()
	content, err := json.MarshalIndent(characters, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(path, content, 0o600)
}

// Подготавливает основное хранилище и все индексы.
func (characters *Characters) ensureMaps() {
	characters.ensureItems()
	if characters.ByAccountID == nil {
		characters.ByAccountID = make(map[int64]map[int64]*Character)
	}
}

// Подготавливает основную map персонажей.
func (characters *Characters) ensureItems() {
	if characters.Items == nil {
		characters.Items = make(map[int64]*Character)
	}
}

// Проверяет обязательные поля персонажа.
func (characters *Characters) validateRequiredFields(character *Character) error {
	if character.AccountID <= 0 {
		return errors.New("account ID is empty")
	}
	return nil
}

// Добавляет персонажа во все быстрые индексы.
func (characters *Characters) addIndexes(character *Character) {
	if characters.ByAccountID[character.AccountID] == nil {
		characters.ByAccountID[character.AccountID] = make(map[int64]*Character)
	}
	characters.ByAccountID[character.AccountID][character.ID] = character
}

// Удаляет персонажа из всех быстрых индексов.
func (characters *Characters) deleteIndexes(character *Character) {
	accountCharacters := characters.ByAccountID[character.AccountID]
	if accountCharacters == nil {
		return
	}

	delete(accountCharacters, character.ID)
	if len(accountCharacters) == 0 {
		delete(characters.ByAccountID, character.AccountID)
	}
}
