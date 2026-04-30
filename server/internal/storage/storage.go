package storage

import (
	"path/filepath"

	"space-game-07-server/internal/data"
)

const accountsFileName = "Accounts.json"
const charactersFileName = "Characters.json"
const cosmicObjectsFileName = "CosmicObjects.json"
const cosmicObjectTypesFileName = "CosmicObjectTypes.json"
const cosmicObjectModelsFileName = "CosmicObjectModels.json"
const itemtypesFileName = "Itemtypes.json"

// Объединяет данные сервера, загружаемые из JSON-файлов при старте.
type ServerData struct {
	Accounts           *data.Accounts
	Characters         *data.Characters
	CosmicObjects      *data.CosmicObjects
	CosmicObjectTypes  *data.CosmicObjectTypes
	CosmicObjectModels *data.CosmicObjectModels
	Itemtypes          *data.Itemtypes
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

	return &ServerData{
		Accounts:           accounts,
		Characters:         characters,
		CosmicObjects:      cosmicObjects,
		CosmicObjectTypes:  cosmicObjectTypes,
		CosmicObjectModels: cosmicObjectModels,
		Itemtypes:          itemtypes,
	}, nil
}
