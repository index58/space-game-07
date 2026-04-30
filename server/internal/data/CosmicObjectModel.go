package data

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"strconv"
)

const defaultCosmicObjectModelTextureScale = 4

// Хранит данные одной модели космического объекта.
type CosmicObjectModel struct {
	ID                 int64   `json:"ID"`
	TitleRu            string  `json:"TitleRu"`
	TitleEn            string  `json:"TitleEn"`
	Acronym            string  `json:"Acronym"`
	IconFilePath       string  `json:"IconFilePath"`
	TextureFilePath    string  `json:"TextureFilePath"`
	TextureWidth       int64   `json:"TextureWidth"`
	TextureHeight      int64   `json:"TextureHeight"`
	TextureBodyOriginX int64   `json:"TextureBodyOriginX"`
	TextureBodyOriginY int64   `json:"TextureBodyOriginY"`
	TextureBodyWidth   int64   `json:"TextureBodyWidth"`
	TextureBodyLength  int64   `json:"TextureBodyLength"`
	TextureScale       float64 `json:"TextureScale"`
	CosmicObjectTypeID int64   `json:"CosmicObjectTypeID"`
	Mass               float64 `json:"Mass"`
	Capacity           float64 `json:"Capacity"`
	MaxArmor           float64 `json:"MaxArmor"`
	MaxSpeed           float64 `json:"MaxSpeed"`
	MaxAngularSpeed    float64 `json:"MaxAngularSpeed"`
	Complexity         float64 `json:"Complexity"`
	BodyLength         float64 `json:"BodyLength"`
	BodyWidth          float64 `json:"BodyWidth"`
}

// Хранит модели космических объектов и быстрые индексы по уникальным полям.
type CosmicObjectModels struct {
	MaxID int64                        `json:"MaxID"`
	Items map[int64]*CosmicObjectModel `json:"Items"`

	ByTitleRu map[string]*CosmicObjectModel `json:"-"`
	ByTitleEn map[string]*CosmicObjectModel `json:"-"`
	ByAcronym map[string]*CosmicObjectModel `json:"-"`
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
	content, err := json.MarshalIndent(cosmicObjectModels, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(path, content, 0o600)
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
	cosmicObjectModel.BodyLength = float64(cosmicObjectModel.TextureBodyLength) / cosmicObjectModel.TextureScale
	cosmicObjectModel.BodyWidth = float64(cosmicObjectModel.TextureBodyWidth) / cosmicObjectModel.TextureScale
}

type legacyCosmicObjectModel struct {
	TextureFilePath      string  `json:"TextureFilePath"`
	TextureWidth         int64   `json:"TextureWidth"`
	TextureHeight        int64   `json:"TextureHeight"`
	TextureObjectOriginX int64   `json:"TextureObjectOriginX"`
	TextureObjectOriginY int64   `json:"TextureObjectOriginY"`
	TextureObjectWidth   int64   `json:"TextureObjectWidth"`
	TextureObjectLength  int64   `json:"TextureObjectLength"`
	CosmicObjectType     string  `json:"CosmicObjectType"`
	TitleRu              string  `json:"TitleRu"`
	TitleEn              string  `json:"TitleEn"`
	Acronym              string  `json:"Acronym"`
	Mass                 float64 `json:"Mass"`
	MaxSpeed             float64 `json:"MaxSpeed"`
	MaxAngularSpeed      float64 `json:"MaxAngularSpeed"`
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
