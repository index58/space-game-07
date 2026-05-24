package data

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
)

// Проверяет обязательные поля модели предмета.
type ItemModel struct {
	ID                      int64   `json:"ID"`                      // Уникальный числовой идентификатор записи.
	TitleRu                 string  `json:"TitleRu"`                 // Русское название для интерфейса и данных.
	TitleEn                 string  `json:"TitleEn"`                 // Английское название для интерфейса и данных.
	Acronym                 string  `json:"Acronym"`                 // Неизменяемый строковый идентификатор для логики и ссылок.
	IconFilePath            string  `json:"IconFilePath"`            // Путь к файлу иконки предмета.
	ItemTypeID              int64   `json:"ItemTypeID"`              // Тип предмета из справочника типов.
	Mass                    float64 `json:"Mass"`                    // Масса одной единицы предмета.
	Volume                  float64 `json:"Volume"`                  // Объем одной единицы предмета.
	Capacity                float64 `json:"Capacity"`                // Вместимость оборудования или предмета.
	MaxArmor                float64 `json:"MaxArmor"`                // Максимальная прочность установленного оборудования.
	ConsumingPower          float64 `json:"ConsumingPower"`          // Потребление энергии включенным оборудованием.
	GeneratingPower         float64 `json:"GeneratingPower"`         // Генерация энергии включенным оборудованием.
	AmmoItemModelID         int64   `json:"AmmoItemModelID"`         // Модель боеприпаса для оружия.
	FiringRate              float64 `json:"FiringRate"`              // Количество выстрелов в секунду.
	MagazineCapacity        int64   `json:"MagazineCapacity"`        // Вместимость магазина оружия.
	RechargeTime            float64 `json:"RechargeTime"`            // Время перезарядки в секундах.
	Range                   float64 `json:"Range"`                   // Дальность действия оборудования.
	MiningSpeed             float64 `json:"MiningSpeed"`             // Скорость добычи ресурса в килограммах за секунду для одной установленной единицы.
	ProjectileObjectModelID int64   `json:"ProjectileObjectModelID"` // Модель космического объекта, создаваемого оружием при выстреле.
	ConsumingItemModelID    int64   `json:"ConsumingItemModelID"`    // Модель потребляемого ресурса.
	ConsumingCount          float64 `json:"ConsumingCount"`          // Расход ресурса за секунду.
	MaxAlongForce           float64 `json:"MaxAlongForce"`           // Максимальная продольная сила оборудования.
	MaxAcrossForce          float64 `json:"MaxAcrossForce"`          // Максимальная поперечная сила оборудования.
	MaxTorque               float64 `json:"MaxTorque"`               // Максимальный крутящий момент оборудования.
	MaxEquipmentCount       int64   `json:"MaxEquipmentCount"`       // Максимальное количество единиц оборудования на объекте.
	ArmorRepairSpeed        float64 `json:"ArmorRepairSpeed"`        // Скорость восстановления прочности.
	Complexity              float64 `json:"Complexity"`              // Сложность изготовления или обслуживания.
	Efficiency              float64 `json:"Efficiency"`              // КПД оборудования при выполнении работы.
}

// Проверяет обязательные поля модели предмета.
type ItemModels struct {
	MaxID int64                `json:"MaxID"` // Последний выданный числовой идентификатор для новых записей.
	Items map[int64]*ItemModel `json:"Items"` // Основное хранилище записей по числовому идентификатору.

	ByAcronym map[string]*ItemModel `json:"-"` // Быстрый поиск записи по акрониму.
}

// Проверяет обязательные поля модели предмета.
func NewItemModels() *ItemModels {
	models := &ItemModels{}
	models.ensureMaps()
	return models
}

// Проверяет обязательные поля модели предмета.
func (models *ItemModels) Get(id int64) (*ItemModel, bool) {
	models.ensureMaps()
	model, ok := models.Items[id]
	return model, ok
}

// Проверяет обязательные поля модели предмета.
func (models *ItemModels) GetByAcronym(acronym string) (*ItemModel, bool) {
	models.ensureMaps()
	model, ok := models.ByAcronym[acronym]
	return model, ok
}

// Проверяет обязательные поля модели предмета.
func (models *ItemModels) RebuildIndexes() error {
	models.ensureItems()
	models.ByAcronym = make(map[string]*ItemModel)

	var maxID int64
	for id, model := range models.Items {
		if model == nil {
			return fmt.Errorf("item model with ID %d is nil", id)
		}
		if model.ID != id {
			return fmt.Errorf("item model map key %d does not match model ID %d", id, model.ID)
		}
		if err := models.validateRequiredFields(model); err != nil {
			return fmt.Errorf("item model with ID %d is invalid: %w", id, err)
		}
		if id > maxID {
			maxID = id
		}
		if existing, ok := models.ByAcronym[model.Acronym]; ok && existing.ID != model.ID {
			return fmt.Errorf("acronym %q already exists", model.Acronym)
		}
		models.ByAcronym[model.Acronym] = model
	}
	if models.MaxID < maxID {
		models.MaxID = maxID
	}
	return nil
}

// Проверяет обязательные поля модели предмета.
func (models *ItemModels) LoadFromFile(path string) error {
	content, err := os.ReadFile(path)
	if err != nil {
		return err
	}

	loaded := ItemModels{}
	if err := json.Unmarshal(content, &loaded); err != nil {
		return err
	}
	if err := loaded.RebuildIndexes(); err != nil {
		return err
	}

	*models = loaded
	return nil
}

// Проверяет обязательные поля модели предмета.
func (models *ItemModels) SaveToFile(path string) error {
	models.ensureMaps()
	return saveTableWithOrderedItems(path, models.MaxID, models.Items)
}

// Проверяет обязательные поля модели предмета.
func (models *ItemModels) ensureMaps() {
	models.ensureItems()
	if models.ByAcronym == nil {
		models.ByAcronym = make(map[string]*ItemModel)
	}
}

// Проверяет обязательные поля модели предмета.
func (models *ItemModels) ensureItems() {
	if models.Items == nil {
		models.Items = make(map[int64]*ItemModel)
	}
}

// Проверяет обязательные поля модели предмета.
func (models *ItemModels) validateRequiredFields(model *ItemModel) error {
	if model.Acronym == "" {
		return errors.New("acronym is empty")
	}
	return nil
}
