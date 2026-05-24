package data

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"sort"
)

// Хранит итоговые характеристики модели объекта с установленным оборудованием.
type Assembly struct {
	ID                  int64   `json:"ID"`                  // Уникальный числовой идентификатор записи.
	AuthorCharacterID   int64   `json:"AuthorCharacterID"`   // Авторская привязка, равная нулю для системных записей разработчиков.
	Title               string  `json:"Title"`               // Человекочитаемое название варианта оснащения.
	CosmicObjectModelID int64   `json:"CosmicObjectModelID"` // Базовая модель корпуса, для которой рассчитаны характеристики.
	IsPublic            bool    `json:"IsPublic"`            // Доступность варианта для общего использования.
	Mass                float64 `json:"Mass"`                // Итоговая масса корпуса и оборудования в килограммах.
	MaxArmor            float64 `json:"MaxArmor"`            // Итоговая максимальная броня.
	MaxAlongForce       float64 `json:"MaxAlongForce"`       // Итоговая доступная продольная тяга.
	MaxAcrossForce      float64 `json:"MaxAcrossForce"`      // Итоговая доступная поперечная тяга.
	MaxTorque           float64 `json:"MaxTorque"`           // Итоговый доступный крутящий момент.
	GeneratingPower     float64 `json:"GeneratingPower"`     // Итоговая вырабатываемая мощность.
	ConsumingPower      float64 `json:"ConsumingPower"`      // Итоговая потребляемая мощность.
	Complexity          float64 `json:"Complexity"`          // Итоговая сложность производства.
	OccupiedVolume      float64 `json:"OccupiedVolume"`      // Объем, занятый оборудованием.
	MaxFuel             float64 `json:"MaxFuel"`             // Итоговая максимальная вместимость топлива.
}

// Хранит варианты оснащения и быстрый поиск по модели корпуса.
type Assemblies struct {
	MaxID int64               `json:"MaxID"` // Последний выданный числовой идентификатор записей.
	Items map[int64]*Assembly `json:"Items"` // Основное хранилище записей по числовому идентификатору.

	ByCosmicObjectModelID map[int64][]*Assembly `json:"-"` // Быстрый поиск вариантов по базовой модели корпуса.
}

// Создает пустое хранилище вариантов оснащения с подготовленными индексами.
func NewAssemblies() *Assemblies {
	assemblies := &Assemblies{}
	assemblies.ensureMaps()
	return assemblies
}

// Добавляет новый вариант оснащения и назначает новый ID.
func (assemblies *Assemblies) Add(assembly *Assembly) (*Assembly, error) {
	if assembly == nil {
		return nil, errors.New("assembly is nil")
	}
	assemblies.ensureMaps()
	if err := assemblies.validateRequiredFields(assembly); err != nil {
		return nil, err
	}

	assemblies.MaxID++
	assembly.ID = assemblies.MaxID
	assemblies.Items[assembly.ID] = assembly
	assemblies.addIndexes(assembly)
	return assembly, nil
}

// Возвращает вариант оснащения по ID.
func (assemblies *Assemblies) Get(id int64) (*Assembly, bool) {
	assemblies.ensureMaps()
	assembly, ok := assemblies.Items[id]
	return assembly, ok
}

// Возвращает первую публичную системную сборку для указанной модели.
func (assemblies *Assemblies) FirstPublicDeveloperAssembly(cosmicObjectModelID int64) (*Assembly, bool) {
	assemblies.ensureMaps()
	for _, assembly := range assemblies.ByCosmicObjectModelID[cosmicObjectModelID] {
		if assembly.AuthorCharacterID == 0 && assembly.IsPublic {
			return assembly, true
		}
	}
	return nil, false
}

// Пересобирает быстрые индексы после загрузки из JSON.
func (assemblies *Assemblies) RebuildIndexes() error {
	assemblies.ensureItems()
	assemblies.ByCosmicObjectModelID = make(map[int64][]*Assembly)

	var maxID int64
	ids := make([]int64, 0, len(assemblies.Items))
	for id := range assemblies.Items {
		ids = append(ids, id)
	}
	sort.Slice(ids, func(left int, right int) bool {
		return ids[left] < ids[right]
	})
	for _, id := range ids {
		assembly := assemblies.Items[id]
		if assembly == nil {
			return fmt.Errorf("assembly with ID %d is nil", id)
		}
		if assembly.ID != id {
			return fmt.Errorf("assembly map key %d does not match assembly ID %d", id, assembly.ID)
		}
		if err := assemblies.validateRequiredFields(assembly); err != nil {
			return fmt.Errorf("assembly with ID %d is invalid: %w", id, err)
		}
		if id > maxID {
			maxID = id
		}
		assemblies.addIndexes(assembly)
	}
	if assemblies.MaxID < maxID {
		assemblies.MaxID = maxID
	}
	return nil
}

// Загружает варианты оснащения из JSON-файла.
func (assemblies *Assemblies) LoadFromFile(path string) error {
	content, err := os.ReadFile(path)
	if err != nil {
		return err
	}

	loaded := Assemblies{}
	if err := json.Unmarshal(content, &loaded); err != nil {
		return err
	}
	if err := loaded.RebuildIndexes(); err != nil {
		return err
	}

	*assemblies = loaded
	return nil
}

// Сохраняет варианты оснащения в JSON-файл.
func (assemblies *Assemblies) SaveToFile(path string) error {
	assemblies.ensureMaps()
	return saveTableWithOrderedItems(path, assemblies.MaxID, assemblies.Items)
}

// Подготавливает основное хранилище и индексы.
func (assemblies *Assemblies) ensureMaps() {
	assemblies.ensureItems()
	if assemblies.ByCosmicObjectModelID == nil {
		assemblies.ByCosmicObjectModelID = make(map[int64][]*Assembly)
	}
}

// Подготавливает основную map.
func (assemblies *Assemblies) ensureItems() {
	if assemblies.Items == nil {
		assemblies.Items = make(map[int64]*Assembly)
	}
}

// Проверяет обязательные поля варианта оснащения.
func (assemblies *Assemblies) validateRequiredFields(assembly *Assembly) error {
	if assembly.CosmicObjectModelID <= 0 {
		return errors.New("cosmic object model ID is empty")
	}
	return nil
}

// Добавляет вариант оснащения в быстрые индексы.
func (assemblies *Assemblies) addIndexes(assembly *Assembly) {
	assemblies.ByCosmicObjectModelID[assembly.CosmicObjectModelID] = append(assemblies.ByCosmicObjectModelID[assembly.CosmicObjectModelID], assembly)
}

// Хранит количество оборудования одной модели на конкретной сборке.
type AssemblyEquipmentGroup struct {
	ID                   int64  `json:"ID"`                   // Уникальный числовой идентификатор записи.
	AssemblyID           int64  `json:"AssemblyID"`           // Сборка, на которой установлена группа оборудования.
	Title                string `json:"Title"`                // Название группы установленного оборудования.
	EquipmentItemModelID int64  `json:"EquipmentItemModelID"` // Модель установленного оборудования.
	Count                int64  `json:"Count"`                // Количество установленных единиц оборудования.
}

// Хранит группы оборудования сборок.
type AssemblyEquipmentGroups struct {
	MaxID int64                             `json:"MaxID"` // Последний выданный числовой идентификатор записей.
	Items map[int64]*AssemblyEquipmentGroup `json:"Items"` // Основное хранилище записей по числовому идентификатору.

	ByAssemblyID map[int64][]*AssemblyEquipmentGroup `json:"-"` // Быстрый поиск групп оборудования по сборке.
}

// Создает пустое хранилище групп оборудования.
func NewAssemblyEquipmentGroups() *AssemblyEquipmentGroups {
	return &AssemblyEquipmentGroups{
		Items:        make(map[int64]*AssemblyEquipmentGroup),
		ByAssemblyID: make(map[int64][]*AssemblyEquipmentGroup),
	}
}

// Возвращает группы оборудования указанной сборки.
func (groups *AssemblyEquipmentGroups) GetByAssemblyID(assemblyID int64) []*AssemblyEquipmentGroup {
	if groups.ByAssemblyID == nil {
		_ = groups.RebuildIndexes()
	}
	return groups.ByAssemblyID[assemblyID]
}

// Пересобирает хранилище после загрузки из JSON.
func (groups *AssemblyEquipmentGroups) RebuildIndexes() error {
	if groups.Items == nil {
		groups.Items = make(map[int64]*AssemblyEquipmentGroup)
	}
	groups.ByAssemblyID = make(map[int64][]*AssemblyEquipmentGroup)

	var maxID int64
	ids := make([]int64, 0, len(groups.Items))
	for id := range groups.Items {
		ids = append(ids, id)
	}
	sort.Slice(ids, func(left int, right int) bool {
		return ids[left] < ids[right]
	})
	for _, id := range ids {
		group := groups.Items[id]
		if group == nil {
			return fmt.Errorf("assembly equipment group with ID %d is nil", id)
		}
		if group.ID != id {
			return fmt.Errorf("assembly equipment group map key %d does not match group ID %d", id, group.ID)
		}
		if group.AssemblyID <= 0 {
			return fmt.Errorf("assembly equipment group with ID %d has empty assembly ID", id)
		}
		if id > maxID {
			maxID = id
		}
		groups.ByAssemblyID[group.AssemblyID] = append(groups.ByAssemblyID[group.AssemblyID], group)
	}
	if groups.MaxID < maxID {
		groups.MaxID = maxID
	}
	return nil
}

// Загружает группы оборудования из JSON-файла.
func (groups *AssemblyEquipmentGroups) LoadFromFile(path string) error {
	content, err := os.ReadFile(path)
	if err != nil {
		return err
	}

	loaded := AssemblyEquipmentGroups{}
	if err := json.Unmarshal(content, &loaded); err != nil {
		return err
	}
	if err := loaded.RebuildIndexes(); err != nil {
		return err
	}

	*groups = loaded
	return nil
}

// Сохраняет группы оборудования в JSON-файл.
func (groups *AssemblyEquipmentGroups) SaveToFile(path string) error {
	if groups.Items == nil {
		groups.Items = make(map[int64]*AssemblyEquipmentGroup)
	}
	return saveTableWithOrderedItems(path, groups.MaxID, groups.Items)
}
