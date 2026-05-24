package data

import (
	"encoding/json"
	"errors"
	"fmt"
	"math"
	"os"
	"strconv"
)

const defaultCosmicObjectModelTextureScale = 4
const bodyPolygonVertexCount = 16

// Коэффициент уменьшения физического тела относительно тела на текстуре.
const cosmicObjectModelBodyScale = 0.95

// Описывает локальную точку физического тела относительно центра объекта.
type BodyPoint struct {
	X float64 `json:"x"` // Смещение по горизонтальной локальной оси в метрах.
	Y float64 `json:"y"` // Смещение по продольной локальной оси в метрах.
}

// Хранит данные одной модели космического объекта.
type CosmicObjectModel struct {
	ID                 int64       `json:"ID"`                 // Уникальный числовой идентификатор записи.
	TitleRu            string      `json:"TitleRu"`            // Русское название для интерфейса и данных.
	TitleEn            string      `json:"TitleEn"`            // Английское название для интерфейса и данных.
	Acronym            string      `json:"Acronym"`            // Неизменяемый строковый идентификатор для логики и ссылок.
	IconFilePath       string      `json:"IconFilePath"`       // Путь к маленькому изображению модели в интерфейсе.
	TextureFilePath    string      `json:"TextureFilePath"`    // Путь к основной текстуре объекта в игровом мире.
	TextureWidth       int64       `json:"TextureWidth"`       // Полная ширина текстуры в пикселях.
	TextureHeight      int64       `json:"TextureHeight"`      // Полная высота текстуры в пикселях.
	TextureBodyOriginX int64       `json:"TextureBodyOriginX"` // Горизонтальное смещение физического тела внутри текстуры.
	TextureBodyOriginY int64       `json:"TextureBodyOriginY"` // Вертикальное смещение физического тела внутри текстуры.
	TextureBodyWidth   int64       `json:"TextureBodyWidth"`   // Ширина физического тела на текстуре в пикселях.
	TextureBodyLength  int64       `json:"TextureBodyLength"`  // Длина физического тела на текстуре в пикселях.
	TextureScale       float64     `json:"TextureScale"`       // Количество пикселей текстуры на один метр мира.
	CosmicObjectTypeID int64       `json:"CosmicObjectTypeID"` // Тип объекта, к которому относится модель.
	Mass               float64     `json:"Mass"`               // Базовая масса экземпляра этой модели.
	Capacity           float64     `json:"Capacity"`           // Базовый доступный объем для оборудования или содержимого.
	MaxArmor           float64     `json:"MaxArmor"`           // Базовый максимум брони.
	MaxSpeed           float64     `json:"MaxSpeed"`           // Базовая максимальная линейная скорость.
	MaxAngularSpeed    float64     `json:"MaxAngularSpeed"`    // Базовая максимальная угловая скорость.
	Complexity         float64     `json:"Complexity"`         // Сложность производства и оценки стоимости модели.
	BodyLength         float64     `json:"BodyLength"`         // Рассчитанная длина физического тела в метрах.
	BodyWidth          float64     `json:"BodyWidth"`          // Рассчитанная ширина физического тела в метрах.
	Damage             float64     `json:"Damage"`             // Урон броне при попадании объекта этой модели в цель.
	BodyPolygon        []BodyPoint `json:"-"`                  // Локальные вершины физического тела, рассчитанные при загрузке справочника.
}

// Хранит модели космических объектов и быстрые индексы по уникальным полям.
type CosmicObjectModels struct {
	MaxID int64                        `json:"MaxID"` // Последний выданный идентификатор для новых записей.
	Items map[int64]*CosmicObjectModel `json:"Items"` // Основное хранилище записей по числовому идентификатору.

	ByTitleRu map[string]*CosmicObjectModel `json:"-"` // Быстрый поиск записи по русскому названию.
	ByTitleEn map[string]*CosmicObjectModel `json:"-"` // Быстрый поиск записи по английскому названию.
	ByAcronym map[string]*CosmicObjectModel `json:"-"` // Быстрый поиск записи по акрониму.
}

// Создаёт пустое хранилище моделей космических объектов с подготовленными индексами.
func NewCosmicObjectModels() *CosmicObjectModels {
	cosmicObjectModels := &CosmicObjectModels{}
	cosmicObjectModels.ensureMaps()
	return cosmicObjectModels
}

// Добавляет новую модель космического объекта и назначает новый ID.
func (cosmicObjectModels *CosmicObjectModels) Add(cosmicObjectModel *CosmicObjectModel) (*CosmicObjectModel, error) {
	if cosmicObjectModel == nil {
		return nil, errors.New("cosmic object model is nil")
	}
	cosmicObjectModels.ensureMaps()
	cosmicObjectModels.prepareCalculatedFields(cosmicObjectModel)
	if err := cosmicObjectModels.validateRequiredFields(cosmicObjectModel); err != nil {
		return nil, err
	}
	if err := cosmicObjectModels.ensureUniqueForNewModel(cosmicObjectModel); err != nil {
		return nil, err
	}

	cosmicObjectModels.MaxID++
	cosmicObjectModel.ID = cosmicObjectModels.MaxID
	cosmicObjectModels.Items[cosmicObjectModel.ID] = cosmicObjectModel
	cosmicObjectModels.addIndexes(cosmicObjectModel)
	return cosmicObjectModel, nil
}

// Возвращает модель космического объекта по ID.
func (cosmicObjectModels *CosmicObjectModels) Get(id int64) (*CosmicObjectModel, bool) {
	cosmicObjectModels.ensureMaps()
	cosmicObjectModel, ok := cosmicObjectModels.Items[id]
	return cosmicObjectModel, ok
}

// Удаляет модель космического объекта и все её быстрые индексы.
func (cosmicObjectModels *CosmicObjectModels) Delete(id int64) bool {
	cosmicObjectModels.ensureMaps()
	cosmicObjectModel, ok := cosmicObjectModels.Items[id]
	if !ok {
		return false
	}

	cosmicObjectModels.deleteIndexes(cosmicObjectModel)
	delete(cosmicObjectModels.Items, id)
	return true
}

// Возвращает модель космического объекта по уникальному русскому названию.
func (cosmicObjectModels *CosmicObjectModels) GetByTitleRu(titleRu string) (*CosmicObjectModel, bool) {
	cosmicObjectModels.ensureMaps()
	cosmicObjectModel, ok := cosmicObjectModels.ByTitleRu[titleRu]
	return cosmicObjectModel, ok
}

// Возвращает модель космического объекта по уникальному английскому названию.
func (cosmicObjectModels *CosmicObjectModels) GetByTitleEn(titleEn string) (*CosmicObjectModel, bool) {
	cosmicObjectModels.ensureMaps()
	cosmicObjectModel, ok := cosmicObjectModels.ByTitleEn[titleEn]
	return cosmicObjectModel, ok
}

// Возвращает модель космического объекта по уникальному акрониму.
func (cosmicObjectModels *CosmicObjectModels) GetByAcronym(acronym string) (*CosmicObjectModel, bool) {
	cosmicObjectModels.ensureMaps()
	cosmicObjectModel, ok := cosmicObjectModels.ByAcronym[acronym]
	return cosmicObjectModel, ok
}

// Пересобирает быстрые индексы после загрузки из JSON.
func (cosmicObjectModels *CosmicObjectModels) RebuildIndexes() error {
	cosmicObjectModels.ensureItems()
	cosmicObjectModels.ByTitleRu = make(map[string]*CosmicObjectModel)
	cosmicObjectModels.ByTitleEn = make(map[string]*CosmicObjectModel)
	cosmicObjectModels.ByAcronym = make(map[string]*CosmicObjectModel)

	var maxID int64
	for id, cosmicObjectModel := range cosmicObjectModels.Items {
		if cosmicObjectModel == nil {
			return fmt.Errorf("cosmic object model with ID %d is nil", id)
		}
		if cosmicObjectModel.ID != id {
			return fmt.Errorf("cosmic object model map key %d does not match model ID %d", id, cosmicObjectModel.ID)
		}
		cosmicObjectModels.prepareCalculatedFields(cosmicObjectModel)
		if err := cosmicObjectModels.validateRequiredFields(cosmicObjectModel); err != nil {
			return fmt.Errorf("cosmic object model with ID %d is invalid: %w", id, err)
		}
		if id > maxID {
			maxID = id
		}
		if err := cosmicObjectModels.ensureUniqueForNewModel(cosmicObjectModel); err != nil {
			return err
		}
		cosmicObjectModels.addIndexes(cosmicObjectModel)
	}
	if cosmicObjectModels.MaxID < maxID {
		cosmicObjectModels.MaxID = maxID
	}
	return nil
}

// Загружает модели космических объектов из JSON-файла и пересобирает быстрые индексы.
func (cosmicObjectModels *CosmicObjectModels) LoadFromFile(path string) error {
	content, err := os.ReadFile(path)
	if err != nil {
		return err
	}

	loaded := CosmicObjectModels{}
	if err := json.Unmarshal(content, &loaded); err != nil {
		return err
	}
	if err := loaded.RebuildIndexes(); err != nil {
		return err
	}

	*cosmicObjectModels = loaded
	return nil
}

// Сохраняет модели космических объектов в JSON-файл без вспомогательных индексов.
func (cosmicObjectModels *CosmicObjectModels) SaveToFile(path string) error {
	cosmicObjectModels.ensureMaps()
	return saveTableWithOrderedItems(path, cosmicObjectModels.MaxID, cosmicObjectModels.Items)
}

// Подготавливает основное хранилище и все индексы.
func (cosmicObjectModels *CosmicObjectModels) ensureMaps() {
	cosmicObjectModels.ensureItems()
	if cosmicObjectModels.ByTitleRu == nil {
		cosmicObjectModels.ByTitleRu = make(map[string]*CosmicObjectModel)
	}
	if cosmicObjectModels.ByTitleEn == nil {
		cosmicObjectModels.ByTitleEn = make(map[string]*CosmicObjectModel)
	}
	if cosmicObjectModels.ByAcronym == nil {
		cosmicObjectModels.ByAcronym = make(map[string]*CosmicObjectModel)
	}
}

// Подготавливает основную map моделей космических объектов.
func (cosmicObjectModels *CosmicObjectModels) ensureItems() {
	if cosmicObjectModels.Items == nil {
		cosmicObjectModels.Items = make(map[int64]*CosmicObjectModel)
	}
}

// Проверяет обязательные поля модели космического объекта.
func (cosmicObjectModels *CosmicObjectModels) validateRequiredFields(cosmicObjectModel *CosmicObjectModel) error {
	if cosmicObjectModel.TitleRu == "" {
		return errors.New("title ru is empty")
	}
	if cosmicObjectModel.TitleEn == "" {
		return errors.New("title en is empty")
	}
	if cosmicObjectModel.Acronym == "" {
		return errors.New("acronym is empty")
	}
	if cosmicObjectModel.TextureScale <= 0 {
		return errors.New("texture scale is empty")
	}
	if cosmicObjectModel.CosmicObjectTypeID <= 0 {
		return errors.New("cosmic object type ID is empty")
	}
	return nil
}

// Проверяет уникальные поля перед добавлением в индексы.
func (cosmicObjectModels *CosmicObjectModels) ensureUniqueForNewModel(cosmicObjectModel *CosmicObjectModel) error {
	if existing, ok := cosmicObjectModels.ByTitleRu[cosmicObjectModel.TitleRu]; ok && existing.ID != cosmicObjectModel.ID {
		return fmt.Errorf("title ru %q already exists", cosmicObjectModel.TitleRu)
	}
	if existing, ok := cosmicObjectModels.ByTitleEn[cosmicObjectModel.TitleEn]; ok && existing.ID != cosmicObjectModel.ID {
		return fmt.Errorf("title en %q already exists", cosmicObjectModel.TitleEn)
	}
	if existing, ok := cosmicObjectModels.ByAcronym[cosmicObjectModel.Acronym]; ok && existing.ID != cosmicObjectModel.ID {
		return fmt.Errorf("acronym %q already exists", cosmicObjectModel.Acronym)
	}
	return nil
}

// Добавляет модель космического объекта во все быстрые индексы.
func (cosmicObjectModels *CosmicObjectModels) addIndexes(cosmicObjectModel *CosmicObjectModel) {
	cosmicObjectModels.ByTitleRu[cosmicObjectModel.TitleRu] = cosmicObjectModel
	cosmicObjectModels.ByTitleEn[cosmicObjectModel.TitleEn] = cosmicObjectModel
	cosmicObjectModels.ByAcronym[cosmicObjectModel.Acronym] = cosmicObjectModel
}

// Удаляет модель космического объекта из всех быстрых индексов.
func (cosmicObjectModels *CosmicObjectModels) deleteIndexes(cosmicObjectModel *CosmicObjectModel) {
	delete(cosmicObjectModels.ByTitleRu, cosmicObjectModel.TitleRu)
	delete(cosmicObjectModels.ByTitleEn, cosmicObjectModel.TitleEn)
	delete(cosmicObjectModels.ByAcronym, cosmicObjectModel.Acronym)
}

// Выставляет значения по умолчанию и пересчитывает вычисляемые поля.
func (cosmicObjectModels *CosmicObjectModels) prepareCalculatedFields(cosmicObjectModel *CosmicObjectModel) {
	if cosmicObjectModel.TextureScale == 0 {
		cosmicObjectModel.TextureScale = defaultCosmicObjectModelTextureScale
	}
	cosmicObjectModel.BodyLength = float64(cosmicObjectModel.TextureBodyLength) / cosmicObjectModel.TextureScale * cosmicObjectModelBodyScale
	cosmicObjectModel.BodyWidth = float64(cosmicObjectModel.TextureBodyWidth) / cosmicObjectModel.TextureScale * cosmicObjectModelBodyScale
	cosmicObjectModel.BodyPolygon = buildBodyPolygon(*cosmicObjectModel)
}

// Строит выпуклое тело по равномерным центральным углам эллипса.
func buildBodyPolygon(cosmicObjectModel CosmicObjectModel) []BodyPoint {
	points := make([]BodyPoint, bodyPolygonVertexCount)
	offsetX, offsetY := bodyPolygonOffset(cosmicObjectModel)
	radiusX := cosmicObjectModel.BodyWidth / 2
	radiusY := cosmicObjectModel.BodyLength / 2

	for index := 0; index < bodyPolygonVertexCount; index++ {
		angle := 2 * math.Pi * float64(index) / bodyPolygonVertexCount
		x := offsetX + zeroSmallValue(radiusX*math.Sin(angle))
		y := offsetY + zeroSmallValue(radiusY*math.Cos(angle))
		points[index] = BodyPoint{
			X: zeroSmallValue(x),
			Y: zeroSmallValue(y),
		}
	}

	return points
}

// Рассчитывает смещение центра тела относительно центра текстуры.
func bodyPolygonOffset(cosmicObjectModel CosmicObjectModel) (float64, float64) {
	if cosmicObjectModel.TextureScale <= 0 || cosmicObjectModel.TextureWidth <= 0 || cosmicObjectModel.TextureHeight <= 0 {
		return 0, 0
	}

	return (float64(cosmicObjectModel.TextureBodyOriginX) - float64(cosmicObjectModel.TextureWidth)/2) / cosmicObjectModel.TextureScale,
		(float64(cosmicObjectModel.TextureBodyOriginY) - float64(cosmicObjectModel.TextureHeight)/2) / cosmicObjectModel.TextureScale
}

// Убирает микроскопические погрешности тригонометрии у осевых вершин.
func zeroSmallValue(value float64) float64 {
	if math.Abs(value) < 0.000000000001 {
		return 0
	}
	return value
}

type legacyCosmicObjectModel struct {
	TextureFilePath      string  `json:"TextureFilePath"`      // Путь к основной текстуре в старом формате данных.
	TextureWidth         int64   `json:"TextureWidth"`         // Полная ширина текстуры в пикселях в старом формате.
	TextureHeight        int64   `json:"TextureHeight"`        // Полная высота текстуры в пикселях в старом формате.
	TextureObjectOriginX int64   `json:"TextureObjectOriginX"` // Горизонтальное смещение тела в старом формате.
	TextureObjectOriginY int64   `json:"TextureObjectOriginY"` // Вертикальное смещение тела в старом формате.
	TextureObjectWidth   int64   `json:"TextureObjectWidth"`   // Ширина тела в пикселях в старом формате.
	TextureObjectLength  int64   `json:"TextureObjectLength"`  // Длина тела в пикселях в старом формате.
	CosmicObjectType     string  `json:"CosmicObjectType"`     // Строковое имя типа из старого JSON.
	TitleRu              string  `json:"TitleRu"`              // Русское название из старого JSON.
	TitleEn              string  `json:"TitleEn"`              // Английское название из старого JSON.
	Acronym              string  `json:"Acronym"`              // Акроним из старого JSON.
	Mass                 float64 `json:"Mass"`                 // Базовая масса из старого JSON.
	MaxSpeed             float64 `json:"MaxSpeed"`             // Максимальная линейная скорость из старого JSON.
	MaxAngularSpeed      float64 `json:"MaxAngularSpeed"`      // Максимальная угловая скорость из старого JSON.
}

// Конвертирует старый JSON моделей в текущую структуру данных.
func LoadCosmicObjectModelsFromLegacyFile(path string, cosmicObjectTypes *CosmicObjectTypes) (*CosmicObjectModels, error) {
	content, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	var legacyModels []legacyCosmicObjectModel
	if err := json.Unmarshal(content, &legacyModels); err != nil {
		return nil, err
	}

	titleRuCounts := make(map[string]int)
	titleEnCounts := make(map[string]int)
	for _, legacyModel := range legacyModels {
		titleRuCounts[legacyModel.TitleRu]++
		titleEnCounts[legacyModel.TitleEn]++
	}

	titleRuNumbers := make(map[string]int)
	titleEnNumbers := make(map[string]int)
	cosmicObjectModels := NewCosmicObjectModels()
	for _, legacyModel := range legacyModels {
		cosmicObjectTypeID, err := legacyCosmicObjectTypeID(legacyModel.CosmicObjectType, cosmicObjectTypes)
		if err != nil {
			return nil, err
		}
		titleRu := numberedLegacyTitle(legacyModel.TitleRu, titleRuCounts, titleRuNumbers)
		titleEn := numberedLegacyTitle(legacyModel.TitleEn, titleEnCounts, titleEnNumbers)

		if _, err := cosmicObjectModels.Add(&CosmicObjectModel{
			TitleRu:            titleRu,
			TitleEn:            titleEn,
			Acronym:            legacyModel.Acronym,
			TextureFilePath:    legacyModel.TextureFilePath,
			TextureWidth:       legacyModel.TextureWidth,
			TextureHeight:      legacyModel.TextureHeight,
			TextureBodyOriginX: legacyModel.TextureObjectOriginX,
			TextureBodyOriginY: legacyModel.TextureObjectOriginY,
			TextureBodyWidth:   legacyModel.TextureObjectWidth,
			TextureBodyLength:  legacyModel.TextureObjectLength,
			TextureScale:       defaultCosmicObjectModelTextureScale,
			CosmicObjectTypeID: cosmicObjectTypeID,
			Mass:               legacyModel.Mass,
			MaxSpeed:           legacyModel.MaxSpeed,
			MaxAngularSpeed:    legacyModel.MaxAngularSpeed,
		}); err != nil {
			return nil, err
		}
	}
	return cosmicObjectModels, nil
}

// Добавляет порядковый номер только к повторяющимся названиям.
func numberedLegacyTitle(title string, counts map[string]int, numbers map[string]int) string {
	if counts[title] <= 1 {
		return title
	}
	numbers[title]++
	return title + " " + strconv.Itoa(numbers[title])
}

// Сопоставляет строковый тип старого JSON с ID текущего справочника типов.
func legacyCosmicObjectTypeID(legacyType string, cosmicObjectTypes *CosmicObjectTypes) (int64, error) {
	acronymByLegacyType := map[string]string{
		"ship":     "Ship",
		"station":  "Station",
		"asteroid": "Asteroid",
	}
	acronym, ok := acronymByLegacyType[legacyType]
	if !ok {
		return 0, fmt.Errorf("unknown legacy cosmic object type %q", legacyType)
	}
	cosmicObjectType, ok := cosmicObjectTypes.GetByAcronym(acronym)
	if !ok {
		return 0, fmt.Errorf("cosmic object type %q not found", acronym)
	}
	return cosmicObjectType.ID, nil
}
