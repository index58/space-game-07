package data

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"sort"
)

// Хранит данные одного космического объекта игрового мира.
type CosmicObject struct {
	ID                     int64   `json:"ID"`
	Title                  string  `json:"Title"`
	CosmicObjectModelID    int64   `json:"CosmicObjectModelID"`
	OwnerCharacterID       int64   `json:"OwnerCharacterID"`
	OwnerNpcClanID         int64   `json:"OwnerNpcClanID"`
	CreatorCharacterID     int64   `json:"CreatorCharacterID"`
	Mass                   float64 `json:"Mass"`
	Capacity               float64 `json:"Capacity"`
	MaxArmor               float64 `json:"MaxArmor"`
	MaxSpeed               float64 `json:"MaxSpeed"`
	MaxAngularSpeed        float64 `json:"MaxAngularSpeed"`
	X                      float64 `json:"X"`
	Y                      float64 `json:"Y"`
	Rotation               float64 `json:"Rotation"`
	Armor                  float64 `json:"Armor"`
	MaxAlongForce          float64 `json:"MaxAlongForce"`
	MaxAcrossForce         float64 `json:"MaxAcrossForce"`
	MaxTorque              float64 `json:"MaxTorque"`
	GeneratingPower        float64 `json:"GeneratingPower"`
	ConsumingPower         float64 `json:"ConsumingPower"`
	AlongForce             float64 `json:"AlongForce"`
	AcrossForce            float64 `json:"AcrossForce"`
	Torque                 float64 `json:"Torque"`
	Enabled                bool    `json:"Enabled"`
	LastReceivedDamageTime int64   `json:"LastReceivedDamageTime"`
	Anchored               bool    `json:"Anchored"`
	Complexity             float64 `json:"Complexity"`
	OccupiedVolume         float64 `json:"OccupiedVolume"`
	MaxFuel                float64 `json:"MaxFuel"`
	Fuel                   float64 `json:"Fuel"`
	Speed                  float64 `json:"Speed"`
	AngularSpeed           float64 `json:"AngularSpeed"`
}

// Хранит космические объекты и быстрые индексы по связанным объектам.
type CosmicObjects struct {
	MaxID int64                   `json:"MaxID"`
	Items map[int64]*CosmicObject `json:"Items"`

	ByCosmicObjectModelID map[int64]map[int64]*CosmicObject `json:"-"`
	ByOwnerCharacterID    map[int64]map[int64]*CosmicObject `json:"-"`
}

// Создаёт пустое хранилище космических объектов с подготовленными индексами.
func NewCosmicObjects() *CosmicObjects {
	cosmicObjects := &CosmicObjects{}
	cosmicObjects.ensureMaps()
	return cosmicObjects
}

// Добавляет новый космический объект и назначает новый ID.
func (cosmicObjects *CosmicObjects) Add(cosmicObject *CosmicObject) (*CosmicObject, error) {
	if cosmicObject == nil {
		return nil, errors.New("cosmic object is nil")
	}
	cosmicObjects.ensureMaps()
	if err := cosmicObjects.validateRequiredFields(cosmicObject); err != nil {
		return nil, err
	}

	cosmicObjects.MaxID++
	cosmicObject.ID = cosmicObjects.MaxID
	cosmicObjects.Items[cosmicObject.ID] = cosmicObject
	cosmicObjects.addIndexes(cosmicObject)
	return cosmicObject, nil
}

// Возвращает космический объект по ID.
func (cosmicObjects *CosmicObjects) Get(id int64) (*CosmicObject, bool) {
	cosmicObjects.ensureMaps()
	cosmicObject, ok := cosmicObjects.Items[id]
	return cosmicObject, ok
}

// Удаляет космический объект и все его быстрые индексы.
func (cosmicObjects *CosmicObjects) Delete(id int64) bool {
	cosmicObjects.ensureMaps()
	cosmicObject, ok := cosmicObjects.Items[id]
	if !ok {
		return false
	}

	cosmicObjects.deleteIndexes(cosmicObject)
	delete(cosmicObjects.Items, id)
	return true
}

// Возвращает объекты указанной модели в порядке возрастания ID.
func (cosmicObjects *CosmicObjects) GetByCosmicObjectModelID(cosmicObjectModelID int64) []*CosmicObject {
	cosmicObjects.ensureMaps()
	return sortedCosmicObjects(cosmicObjects.ByCosmicObjectModelID[cosmicObjectModelID])
}

// Возвращает объекты указанного владельца-персонажа в порядке возрастания ID.
func (cosmicObjects *CosmicObjects) GetByOwnerCharacterID(ownerCharacterID int64) []*CosmicObject {
	cosmicObjects.ensureMaps()
	return sortedCosmicObjects(cosmicObjects.ByOwnerCharacterID[ownerCharacterID])
}

// Пересобирает быстрые индексы после загрузки из JSON.
func (cosmicObjects *CosmicObjects) RebuildIndexes() error {
	cosmicObjects.ensureItems()
	cosmicObjects.ByCosmicObjectModelID = make(map[int64]map[int64]*CosmicObject)
	cosmicObjects.ByOwnerCharacterID = make(map[int64]map[int64]*CosmicObject)

	var maxID int64
	for id, cosmicObject := range cosmicObjects.Items {
		if cosmicObject == nil {
			return fmt.Errorf("cosmic object with ID %d is nil", id)
		}
		if cosmicObject.ID != id {
			return fmt.Errorf("cosmic object map key %d does not match object ID %d", id, cosmicObject.ID)
		}
		if err := cosmicObjects.validateRequiredFields(cosmicObject); err != nil {
			return fmt.Errorf("cosmic object with ID %d is invalid: %w", id, err)
		}
		if id > maxID {
			maxID = id
		}
		cosmicObjects.addIndexes(cosmicObject)
	}
	if cosmicObjects.MaxID < maxID {
		cosmicObjects.MaxID = maxID
	}
	return nil
}

// Загружает космические объекты из JSON-файла и пересобирает быстрые индексы.
func (cosmicObjects *CosmicObjects) LoadFromFile(path string) error {
	content, err := os.ReadFile(path)
	if err != nil {
		return err
	}

	loaded := CosmicObjects{}
	if err := json.Unmarshal(content, &loaded); err != nil {
		return err
	}
	if err := loaded.RebuildIndexes(); err != nil {
		return err
	}

	*cosmicObjects = loaded
	return nil
}

// Сохраняет космические объекты в JSON-файл без вспомогательных индексов.
func (cosmicObjects *CosmicObjects) SaveToFile(path string) error {
	cosmicObjects.ensureMaps()
	content, err := json.MarshalIndent(cosmicObjects, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(path, content, 0o600)
}

// Подготавливает основное хранилище и все индексы.
func (cosmicObjects *CosmicObjects) ensureMaps() {
	cosmicObjects.ensureItems()
	if cosmicObjects.ByCosmicObjectModelID == nil {
		cosmicObjects.ByCosmicObjectModelID = make(map[int64]map[int64]*CosmicObject)
	}
	if cosmicObjects.ByOwnerCharacterID == nil {
		cosmicObjects.ByOwnerCharacterID = make(map[int64]map[int64]*CosmicObject)
	}
}

// Подготавливает основную map космических объектов.
func (cosmicObjects *CosmicObjects) ensureItems() {
	if cosmicObjects.Items == nil {
		cosmicObjects.Items = make(map[int64]*CosmicObject)
	}
}

// Проверяет обязательные поля космического объекта.
func (cosmicObjects *CosmicObjects) validateRequiredFields(cosmicObject *CosmicObject) error {
	if cosmicObject.CosmicObjectModelID <= 0 {
		return errors.New("cosmic object model ID is empty")
	}
	return nil
}

// Добавляет космический объект во все быстрые индексы.
func (cosmicObjects *CosmicObjects) addIndexes(cosmicObject *CosmicObject) {
	addCosmicObjectIndex(cosmicObjects.ByCosmicObjectModelID, cosmicObject.CosmicObjectModelID, cosmicObject)
	if cosmicObject.OwnerCharacterID > 0 {
		addCosmicObjectIndex(cosmicObjects.ByOwnerCharacterID, cosmicObject.OwnerCharacterID, cosmicObject)
	}
}

// Удаляет космический объект из всех быстрых индексов.
func (cosmicObjects *CosmicObjects) deleteIndexes(cosmicObject *CosmicObject) {
	deleteCosmicObjectIndex(cosmicObjects.ByCosmicObjectModelID, cosmicObject.CosmicObjectModelID, cosmicObject.ID)
	if cosmicObject.OwnerCharacterID > 0 {
		deleteCosmicObjectIndex(cosmicObjects.ByOwnerCharacterID, cosmicObject.OwnerCharacterID, cosmicObject.ID)
	}
}

// Добавляет объект в неуникальный индекс.
func addCosmicObjectIndex(index map[int64]map[int64]*CosmicObject, key int64, cosmicObject *CosmicObject) {
	if index[key] == nil {
		index[key] = make(map[int64]*CosmicObject)
	}
	index[key][cosmicObject.ID] = cosmicObject
}

// Удаляет объект из неуникального индекса.
func deleteCosmicObjectIndex(index map[int64]map[int64]*CosmicObject, key int64, id int64) {
	indexItems := index[key]
	if indexItems == nil {
		return
	}
	delete(indexItems, id)
	if len(indexItems) == 0 {
		delete(index, key)
	}
}

// Возвращает объекты индекса в стабильном порядке ID.
func sortedCosmicObjects(indexItems map[int64]*CosmicObject) []*CosmicObject {
	if len(indexItems) == 0 {
		return []*CosmicObject{}
	}

	ids := make([]int64, 0, len(indexItems))
	for id := range indexItems {
		ids = append(ids, id)
	}
	sort.Slice(ids, func(left int, right int) bool {
		return ids[left] < ids[right]
	})

	result := make([]*CosmicObject, 0, len(ids))
	for _, id := range ids {
		result = append(result, indexItems[id])
	}
	return result
}
