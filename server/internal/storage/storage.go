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
const itemGroupFileName = "ItemGroups.json"
const npcClanFileName = "NpcClans.json"
const itemModelFileName = "ItemModels.json"
const blueprintFileName = "Blueprints.json"
const blueprintComponentFileName = "BlueprintComponents.json"
const schemaFileName = "Schemas.json"
const schemaComponentFileName = "SchemaComponents.json"
const assemblyFileName = "Assemblies.json"
const assemblyEquipmentGroupFileName = "AssemblyEquipmentGroups.json"
const chatsFileName = "Chats.json"
const chatMembersFileName = "ChatMembers.json"
const communityTypesFileName = "CommunityTypes.json"
const communityChatRolesFileName = "CommunityChatRoles.json"
const messagesFileName = "Messages.json"
const messageReadsFileName = "MessageReads.json"
const messageTypesFileName = "MessageTypes.json"

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
	ItemGroups              *data.ItemGroups              // Загруженные группы предметов внутри контейнеров.
	Assemblies              *data.Assemblies              // Загруженный справочник сборок космических объектов.
	AssemblyEquipmentGroups *data.AssemblyEquipmentGroups // Загруженный справочник оборудования сборок.
	NpcClans                *RawReferenceTable            // Загруженный справочник NPC-кланов.
	ItemModels              *data.ItemModels              // Загруженный справочник моделей предметов.
	Blueprints              *RawReferenceTable            // Загруженный справочник чертежей объектов.
	BlueprintComponents     *RawReferenceTable            // Загруженный справочник компонентов чертежей.
	Schemas                 *RawReferenceTable            // Загруженный справочник схем предметов.
	SchemaComponents        *RawReferenceTable            // Загруженный справочник компонентов схем.
	Chats                   *data.Chats                   // Загруженные чаты игрового мира.
	ChatMembers             *data.ChatMembers             // Загруженные участники чатов.
	CommunityTypes          *data.CommunityTypes          // Загруженный справочник типов сообществ.
	CommunityChatRoles      *data.CommunityChatRoles      // Загруженный справочник ролей в чатах сообществ.
	Messages                *data.Messages                // Загруженные сообщения чатов.
	MessageReads            *data.MessageReads            // Загруженные позиции чтения сообщений персонажами.
	MessageTypes            *data.MessageTypes            // Загруженный справочник типов сообщений.
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

	itemGroups, err := loadOptionalItemGroups(filepath.Join(dataDirectory, itemGroupFileName))
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
	itemModels, err := loadOptionalItemModels(filepath.Join(dataDirectory, itemModelFileName))
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
	chats, err := loadOptionalChats(filepath.Join(dataDirectory, chatsFileName))
	if err != nil {
		return nil, err
	}
	chatMembers, err := loadOptionalChatMembers(filepath.Join(dataDirectory, chatMembersFileName))
	if err != nil {
		return nil, err
	}
	communityTypes, err := loadOptionalCommunityTypes(filepath.Join(dataDirectory, communityTypesFileName))
	if err != nil {
		return nil, err
	}
	communityChatRoles, err := loadOptionalCommunityChatRoles(filepath.Join(dataDirectory, communityChatRolesFileName))
	if err != nil {
		return nil, err
	}
	messages, err := loadOptionalMessages(filepath.Join(dataDirectory, messagesFileName))
	if err != nil {
		return nil, err
	}
	messageReads, err := loadOptionalMessageReads(filepath.Join(dataDirectory, messageReadsFileName))
	if err != nil {
		return nil, err
	}
	messageTypes, err := loadOptionalMessageTypes(filepath.Join(dataDirectory, messageTypesFileName))
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
		ItemGroups:              itemGroups,
		Assemblies:              assemblies,
		AssemblyEquipmentGroups: assemblyEquipmentGroups,
		NpcClans:                npcClans,
		ItemModels:              itemModels,
		Blueprints:              blueprints,
		BlueprintComponents:     blueprintComponents,
		Schemas:                 schemas,
		SchemaComponents:        schemaComponents,
		Chats:                   chats,
		ChatMembers:             chatMembers,
		CommunityTypes:          communityTypes,
		CommunityChatRoles:      communityChatRoles,
		Messages:                messages,
		MessageReads:            messageReads,
		MessageTypes:            messageTypes,
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

// Загружает необязательную таблицу групп предметов или возвращает пустой контейнер до появления файла.
func loadOptionalItemGroups(path string) (*data.ItemGroups, error) {
	groups := data.NewItemGroups()
	if err := groups.LoadFromFile(path); err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return data.NewItemGroups(), nil
		}
		return nil, err
	}
	return groups, nil
}

// Загружает необязательную таблицу моделей предметов или возвращает пустой контейнер до появления файла.
func loadOptionalItemModels(path string) (*data.ItemModels, error) {
	models := data.NewItemModels()
	if err := models.LoadFromFile(path); err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return data.NewItemModels(), nil
		}
		return nil, err
	}
	return models, nil
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

// Загружает необязательную таблицу чатов или возвращает пустое хранилище до появления файла.
func loadOptionalChats(path string) (*data.Chats, error) {
	chats := data.NewChats()
	if err := chats.LoadFromFile(path); err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return data.NewChats(), nil
		}
		return nil, err
	}
	return chats, nil
}

// Загружает необязательную таблицу участников чатов или возвращает пустое хранилище до появления файла.
func loadOptionalChatMembers(path string) (*data.ChatMembers, error) {
	members := data.NewChatMembers()
	if err := members.LoadFromFile(path); err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return data.NewChatMembers(), nil
		}
		return nil, err
	}
	return members, nil
}

// Загружает необязательный справочник типов сообществ или возвращает пустое хранилище до появления файла.
func loadOptionalCommunityTypes(path string) (*data.CommunityTypes, error) {
	types := data.NewCommunityTypes()
	if err := types.LoadFromFile(path); err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return data.NewCommunityTypes(), nil
		}
		return nil, err
	}
	return types, nil
}

// Загружает необязательный справочник ролей чатов или возвращает пустое хранилище до появления файла.
func loadOptionalCommunityChatRoles(path string) (*data.CommunityChatRoles, error) {
	roles := data.NewCommunityChatRoles()
	if err := roles.LoadFromFile(path); err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return data.NewCommunityChatRoles(), nil
		}
		return nil, err
	}
	return roles, nil
}

// Загружает необязательную таблицу сообщений или возвращает пустое хранилище до появления файла.
func loadOptionalMessages(path string) (*data.Messages, error) {
	messages := data.NewMessages()
	if err := messages.LoadFromFile(path); err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return data.NewMessages(), nil
		}
		return nil, err
	}
	return messages, nil
}

// Загружает необязательную таблицу прочтений сообщений или возвращает пустое хранилище до появления файла.
func loadOptionalMessageReads(path string) (*data.MessageReads, error) {
	reads := data.NewMessageReads()
	if err := reads.LoadFromFile(path); err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return data.NewMessageReads(), nil
		}
		return nil, err
	}
	return reads, nil
}

// Загружает необязательный справочник типов сообщений или возвращает пустое хранилище до появления файла.
func loadOptionalMessageTypes(path string) (*data.MessageTypes, error) {
	types := data.NewMessageTypes()
	if err := types.LoadFromFile(path); err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return data.NewMessageTypes(), nil
		}
		return nil, err
	}
	return types, nil
}
