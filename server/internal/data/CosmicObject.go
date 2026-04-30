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
	ID                     int64   `json:"ID"`                     // Уникальный числовой идентификатор записи.
	Title                  string  `json:"Title"`                  // Пользовательское название объекта в игровом мире.
	CosmicObjectModelID    int64   `json:"CosmicObjectModelID"`    // Модель, от которой взяты базовые характеристики и графика.
	OwnerCharacterID       int64   `json:"OwnerCharacterID"`       // Персонаж-владелец, если объект принадлежит игроку.
	OwnerNpcClanID         int64   `json:"OwnerNpcClanID"`         // NPC-клан-владелец, если объект не принадлежит персонажу.
	CreatorCharacterID     int64   `json:"CreatorCharacterID"`     // Персонаж, создавший объект.
	Mass                   float64 `json:"Mass"`                   // Текущая суммарная масса объекта и содержимого.
	Capacity               float64 `json:"Capacity"`               // Максимальный объем оборудования или содержимого.
	MaxArmor               float64 `json:"MaxArmor"`               // Верхняя граница прочности брони.
	MaxSpeed               float64 `json:"MaxSpeed"`               // Максимальная линейная скорость в метрах за секунду.
	MaxAngularSpeed        float64 `json:"MaxAngularSpeed"`        // Максимальная угловая скорость в радианах за секунду.
	X                      float64 `json:"X"`                      // Горизонтальная координата положения в мире.
	Y                      float64 `json:"Y"`                      // Вертикальная координата положения в мире.
	Rotation               float64 `json:"Rotation"`               // Текущий угол поворота в радианах без нормализации.
	Armor                  float64 `json:"Armor"`                  // Текущее количество единиц брони.
	MaxAlongForce          float64 `json:"MaxAlongForce"`          // Доступная продольная сила тяги.
	MaxAcrossForce         float64 `json:"MaxAcrossForce"`         // Доступная поперечная сила тяги.
	MaxTorque              float64 `json:"MaxTorque"`              // Доступный крутящий момент.
	GeneratingPower        float64 `json:"GeneratingPower"`        // Суммарная вырабатываемая мощность оборудования.
	ConsumingPower         float64 `json:"ConsumingPower"`         // Суммарная потребляемая мощность оборудования.
	AlongForce             float64 `json:"AlongForce"`             // Фактически примененная продольная тяга на текущем шаге.
	AcrossForce            float64 `json:"AcrossForce"`            // Фактически примененная поперечная тяга на текущем шаге.
	Torque                 float64 `json:"Torque"`                 // Фактически примененный крутящий момент на текущем шаге.
	Enabled                bool    `json:"Enabled"`                // Разрешает объекту работать и участвовать в симуляции.
	LastReceivedDamageTime int64   `json:"LastReceivedDamageTime"` // Время последнего получения урона для боевых и ремонтных правил.
	Anchored               bool    `json:"Anchored"`               // Запрещает физическое перемещение объекта.
	Complexity             float64 `json:"Complexity"`             // Сложность устройства для производства и оценки стоимости.
	OccupiedVolume         float64 `json:"OccupiedVolume"`         // Объем, уже занятый содержимым или оборудованием.
	MaxFuel                float64 `json:"MaxFuel"`                // Максимальный запас топлива.
	Fuel                   float64 `json:"Fuel"`                   // Текущий запас топлива.
	Speed                  float64 `json:"Speed"`                  // Текущая длина вектора скорости.
	VelocityX              float64 `json:"VelocityX"`              // Горизонтальная компонента текущей скорости.
	VelocityY              float64 `json:"VelocityY"`              // Вертикальная компонента текущей скорости.
	AngularSpeed           float64 `json:"AngularSpeed"`           // Текущая угловая скорость в радианах за секунду.
	TargetRotation         float64 `json:"TargetRotation"`         // Угол, к которому автоматика поворота ведет объект.
}

// Хранит космические объекты и быстрые индексы по связанным объектам.
type CosmicObjects struct {
	MaxID int64                   `json:"MaxID"` // Последний выданный идентификатор для новых записей.
	Items map[int64]*CosmicObject `json:"Items"` // Основное хранилище записей по числовому идентификатору.

	ByCosmicObjectModelID map[int64]map[int64]*CosmicObject `json:"-"` // Быстрый поиск объектов по модели.
	ByOwnerCharacterID    map[int64]map[int64]*CosmicObject `json:"-"` // Быстрый поиск объектов по персонажу-владельцу.
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
