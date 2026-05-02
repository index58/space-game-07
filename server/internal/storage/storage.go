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
const equipmentGroupFileName = "EquipmentGroups.json"
const npcClanFileName = "NpcClans.json"
const itemModelFileName = "ItemModels.json"
const blueprintFileName = "Blueprints.json"
const blueprintComponentFileName = "BlueprintComponents.json"
const schemaFileName = "Schemas.json"
const schemaComponentFileName = "SchemaComponents.json"
const assemblyFileName = "Assemblies.json"
const assemblyEquipmentGroupFileName = "AssemblyEquipmentGroups.json"

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
	Accounts                *data.Accounts                // Загруженные учетные записи игроков.
	Characters              *data.Characters              // Загруженные персонажи игроков.
	CosmicObjects           *data.CosmicObjects           // Загруженные экземпляры объектов мира.
	CosmicObjectTypes       *data.CosmicObjectTypes       // Загруженный справочник типов космических объектов.
	CosmicObjectModels      *data.CosmicObjectModels      // Загруженный справочник моделей космических объектов.
	Itemtypes               *data.Itemtypes               // Загруженный справочник типов предметов.
	EquipmentGroups         *data.EquipmentGroups         // Загруженные группы оборудования космических объектов.
	Assemblies              *data.Assemblies              // Загруженный справочник сборок космических объектов.
	AssemblyEquipmentGroups *data.AssemblyEquipmentGroups // Загруженный справочник оборудования сборок.
	NpcClans                *RawReferenceTable            // Загруженный справочник NPC-кланов.
	ItemModels              *RawReferenceTable            // Загруженный справочник моделей предметов.
	Blueprints              *RawReferenceTable            // Загруженный справочник чертежей объектов.
	BlueprintComponents     *RawReferenceTable            // Загруженный справочник компонентов чертежей.
	Schemas                 *RawReferenceTable            // Загруженный справочник схем предметов.
	SchemaComponents        *RawReferenceTable            // Загруженный справочник компонентов схем.
}

// Загружает все JSON-файлы данных сервера из указанного рабочего каталога.
func LoadServerData(workingDirectory string) (*ServerData, error) {
	dataDirectory := filepath.Join(workingDirectory, "data")

	accounts := data.NewAccounts()
	if err := accounts.LoadFromFile(filepath.Join(dataDirectory, accountsFileName)); err != nil {
		return nil, err
	}

	characters := data.NewCharacters()
	if err := characters.LoadFromFile(filepath.Join(dataDirectory, charactersFileName)); err != nil {
		return nil, err
	}

	cosmicObjects := data.NewCosmicObjects()
	if err := cosmicObjects.LoadFromFile(filepath.Join(dataDirectory, cosmicObjectsFileName)); err != nil {
		return nil, err
	}

	cosmicObjectTypes := data.NewCosmicObjectTypes()
	if err := cosmicObjectTypes.LoadFromFile(filepath.Join(dataDirectory, cosmicObjectTypesFileName)); err != nil {
		return nil, err
	}

	cosmicObjectModels := data.NewCosmicObjectModels()
	if err := cosmicObjectModels.LoadFromFile(filepath.Join(dataDirectory, cosmicObjectModelsFileName)); err != nil {
		return nil, err
	}

	itemtypes := data.NewItemtypes()
	if err := itemtypes.LoadFromFile(filepath.Join(dataDirectory, itemtypesFileName)); err != nil {
		return nil, err
	}

	equipmentGroups, err := loadOptionalEquipmentGroups(filepath.Join(dataDirectory, equipmentGroupFileName))
	if err != nil {
		return nil, err
	}

	assemblies, err := loadOptionalAssemblies(filepath.Join(dataDirectory, assemblyFileName))
	if err != nil {
		return nil, err
	}

	assemblyEquipmentGroups, err := loadOptionalAssemblyEquipmentGroups(filepath.Join(dataDirectory, assemblyEquipmentGroupFileName))
	if err != nil {
		return nil, err
	}

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
		Accounts:                accounts,
		Characters:              characters,
		CosmicObjects:           cosmicObjects,
		CosmicObjectTypes:       cosmicObjectTypes,
		CosmicObjectModels:      cosmicObjectModels,
		Itemtypes:               itemtypes,
		EquipmentGroups:         equipmentGroups,
		Assemblies:              assemblies,
		AssemblyEquipmentGroups: assemblyEquipmentGroups,
		NpcClans:                npcClans,
		ItemModels:              itemModels,
		Blueprints:              blueprints,
		BlueprintComponents:     blueprintComponents,
		Schemas:                 schemas,
		SchemaComponents:        schemaComponents,
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

// Загружает необязательную таблицу оборудования объектов или возвращает пустой контейнер до появления файла.
func loadOptionalEquipmentGroups(path string) (*data.EquipmentGroups, error) {
	groups := data.NewEquipmentGroups()
	if err := groups.LoadFromFile(path); err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return data.NewEquipmentGroups(), nil
		}
		return nil, err
	}
	return groups, nil
}

// Загружает необязательную таблицу сборок или возвращает пустой контейнер до появления файла.
func loadOptionalAssemblies(path string) (*data.Assemblies, error) {
	assemblies := data.NewAssemblies()
	if err := assemblies.LoadFromFile(path); err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return data.NewAssemblies(), nil
		}
		return nil, err
	}
	return assemblies, nil
}

// Загружает необязательную таблицу оборудования сборок или возвращает пустой контейнер до появления файла.
func loadOptionalAssemblyEquipmentGroups(path string) (*data.AssemblyEquipmentGroups, error) {
	groups := data.NewAssemblyEquipmentGroups()
	if err := groups.LoadFromFile(path); err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return data.NewAssemblyEquipmentGroups(), nil
		}
		return nil, err
	}
	return groups, nil
}
