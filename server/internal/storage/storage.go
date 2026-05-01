package storage

import (
	"encoding/json"
	"errors"
	"os"
	"path/filepath"

	"space-game-07-server/internal/data"
)

const accountsFileName = "Accounts.json"
const charactersFileName = "Characters.json"
const cosmicObjectsFileName = "CosmicObjects.json"
const cosmicObjectTypesFileName = "CosmicObjectTypes.json"
const cosmicObjectModelsFileName = "CosmicObjectModels.json"
const itemtypesFileName = "Itemtypes.json"
const npcClanFileName = "NpcClan.json"
const itemModelFileName = "ItemModel.json"
const blueprintFileName = "Blueprint.json"
const blueprintComponentFileName = "BlueprintComponent.json"
const schemaFileName = "Schema.json"
const schemaComponentFileName = "SchemaComponent.json"

// Хранит таблицу, для которой на сервере пока нет отдельной предметной модели.
type RawReferenceTable struct {
	MaxID int64                      `json:"MaxID"` // Последний выданный числовой идентификатор записей.
	Items map[string]json.RawMessage `json:"Items"` // Записи таблицы по строковому представлению числового идентификатора.
}

// Создает пустой контейнер справочника с тем же JSON-контрактом, что у типизированных таблиц.
func NewRawReferenceTable() *RawReferenceTable {
	return &RawReferenceTable{
		Items: make(map[string]json.RawMessage),
	}
}

// Объединяет данные сервера, загружаемые из JSON-файлов при старте.
type ServerData struct {
	Accounts            *data.Accounts           // Загруженные учетные записи игроков.
	Characters          *data.Characters         // Загруженные персонажи игроков.
	CosmicObjects       *data.CosmicObjects      // Загруженные экземпляры объектов мира.
	CosmicObjectTypes   *data.CosmicObjectTypes  // Загруженный справочник типов космических объектов.
	CosmicObjectModels  *data.CosmicObjectModels // Загруженный справочник моделей космических объектов.
	Itemtypes           *data.Itemtypes          // Загруженный справочник типов предметов.
	NpcClans            *RawReferenceTable       // Загруженный справочник NPC-кланов.
	ItemModels          *RawReferenceTable       // Загруженный справочник моделей предметов.
	Blueprints          *RawReferenceTable       // Загруженный справочник чертежей объектов.
	BlueprintComponents *RawReferenceTable       // Загруженный справочник компонентов чертежей.
	Schemas             *RawReferenceTable       // Загруженный справочник схем предметов.
	SchemaComponents    *RawReferenceTable       // Загруженный справочник компонентов схем.
}

// Загружает все JSON-файлы данных сервера из указанного рабочего каталога.
func LoadServerData(workingDirectory string) (*ServerData, error) {
	accounts := data.NewAccounts()
	if err := accounts.LoadFromFile(filepath.Join(workingDirectory, "data", accountsFileName)); err != nil {
		return nil, err
	}

	characters := data.NewCharacters()
	if err := characters.LoadFromFile(filepath.Join(workingDirectory, "data", charactersFileName)); err != nil {
		return nil, err
	}

	cosmicObjects := data.NewCosmicObjects()
	if err := cosmicObjects.LoadFromFile(filepath.Join(workingDirectory, "data", cosmicObjectsFileName)); err != nil {
		return nil, err
	}

	cosmicObjectTypes := data.NewCosmicObjectTypes()
	if err := cosmicObjectTypes.LoadFromFile(filepath.Join(workingDirectory, "data", cosmicObjectTypesFileName)); err != nil {
		return nil, err
	}

	cosmicObjectModels := data.NewCosmicObjectModels()
	if err := cosmicObjectModels.LoadFromFile(filepath.Join(workingDirectory, "data", cosmicObjectModelsFileName)); err != nil {
		return nil, err
	}

	itemtypes := data.NewItemtypes()
	if err := itemtypes.LoadFromFile(filepath.Join(workingDirectory, "data", itemtypesFileName)); err != nil {
		return nil, err
	}

	dataDirectory := filepath.Join(workingDirectory, "data")
	npcClans, err := loadOptionalRawReferenceTable(filepath.Join(dataDirectory, npcClanFileName))
	if err != nil {
		return nil, err
	}
	itemModels, err := loadOptionalRawReferenceTable(filepath.Join(dataDirectory, itemModelFileName))
	if err != nil {
		return nil, err
	}
	blueprints, err := loadOptionalRawReferenceTable(filepath.Join(dataDirectory, blueprintFileName))
	if err != nil {
		return nil, err
	}
	blueprintComponents, err := loadOptionalRawReferenceTable(filepath.Join(dataDirectory, blueprintComponentFileName))
	if err != nil {
		return nil, err
	}
	schemas, err := loadOptionalRawReferenceTable(filepath.Join(dataDirectory, schemaFileName))
	if err != nil {
		return nil, err
	}
	schemaComponents, err := loadOptionalRawReferenceTable(filepath.Join(dataDirectory, schemaComponentFileName))
	if err != nil {
		return nil, err
	}

	return &ServerData{
		Accounts:            accounts,
		Characters:          characters,
		CosmicObjects:       cosmicObjects,
		CosmicObjectTypes:   cosmicObjectTypes,
		CosmicObjectModels:  cosmicObjectModels,
		Itemtypes:           itemtypes,
		NpcClans:            npcClans,
		ItemModels:          itemModels,
		Blueprints:          blueprints,
		BlueprintComponents: blueprintComponents,
		Schemas:             schemas,
		SchemaComponents:    schemaComponents,
	}, nil
}

// Загружает необязательную таблицу или возвращает пустой контейнер до появления файла.
func loadOptionalRawReferenceTable(path string) (*RawReferenceTable, error) {
	content, err := os.ReadFile(path)
	if errors.Is(err, os.ErrNotExist) {
		return NewRawReferenceTable(), nil
	}
	if err != nil {
		return nil, err
	}

	loaded := NewRawReferenceTable()
	if err := json.Unmarshal(content, loaded); err != nil {
		return nil, err
	}
	if loaded.Items == nil {
		loaded.Items = make(map[string]json.RawMessage)
	}
	return loaded, nil
}
