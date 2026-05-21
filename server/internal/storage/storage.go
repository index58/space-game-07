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
const itemTypesFileName = "ItemTypes.json"
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
const actionTypesFileName = "ActionTypes.json"
const inputEventTypesFileName = "InputEventTypes.json"
const defaultActionInputSettingsFileName = "DefaultActionInputSettings.json"
const accountActionInputSettingsFileName = "AccountActionInputSettings.json"
const taskTypesFileName = "TaskTypes.json"
const tasksFileName = "Tasks.json"
const taskItemGroupsFileName = "TaskItemGroups.json"
const implementersFileName = "Implementers.json"

// Р ТђРЎР‚Р В°Р Р…Р С‘РЎвЂљ РЎвЂљР В°Р В±Р В»Р С‘РЎвЂ РЎС“, Р Т‘Р В»РЎРЏ Р С”Р С•РЎвЂљР С•РЎР‚Р С•Р в„– Р Р…Р В° РЎРѓР ВµРЎР‚Р Р†Р ВµРЎР‚Р Вµ Р С—Р С•Р С”Р В° Р Р…Р ВµРЎвЂљ Р С•РЎвЂљР Т‘Р ВµР В»РЎРЉР Р…Р С•Р в„– Р С—РЎР‚Р ВµР Т‘Р СР ВµРЎвЂљР Р…Р С•Р в„– Р СР С•Р Т‘Р ВµР В»Р С‘.
type RawReferenceTable struct {
	MaxID int64                      `json:"MaxID"` // Р СџР С•РЎРѓР В»Р ВµР Т‘Р Р…Р С‘Р в„– Р Р†РЎвЂ№Р Т‘Р В°Р Р…Р Р…РЎвЂ№Р в„– РЎвЂЎР С‘РЎРѓР В»Р С•Р Р†Р С•Р в„– Р С‘Р Т‘Р ВµР Р…РЎвЂљР С‘РЎвЂћР С‘Р С”Р В°РЎвЂљР С•РЎР‚ Р В·Р В°Р С—Р С‘РЎРѓР ВµР в„–.
	Items map[string]json.RawMessage `json:"Items"` // Р вЂ”Р В°Р С—Р С‘РЎРѓР С‘ РЎвЂљР В°Р В±Р В»Р С‘РЎвЂ РЎвЂ№ Р С—Р С• РЎРѓРЎвЂљРЎР‚Р С•Р С”Р С•Р Р†Р С•Р СРЎС“ Р С—РЎР‚Р ВµР Т‘РЎРѓРЎвЂљР В°Р Р†Р В»Р ВµР Р…Р С‘РЎР‹ РЎвЂЎР С‘РЎРѓР В»Р С•Р Р†Р С•Р С–Р С• Р С‘Р Т‘Р ВµР Р…РЎвЂљР С‘РЎвЂћР С‘Р С”Р В°РЎвЂљР С•РЎР‚Р В°.
}

// Р РЋР С•Р В·Р Т‘Р В°Р ВµРЎвЂљ Р С—РЎС“РЎРѓРЎвЂљР С•Р в„– Р С”Р С•Р Р…РЎвЂљР ВµР в„–Р Р…Р ВµРЎР‚ РЎРѓР С—РЎР‚Р В°Р Р†Р С•РЎвЂЎР Р…Р С‘Р С”Р В° РЎРѓ РЎвЂљР ВµР С Р В¶Р Вµ JSON-Р С”Р С•Р Р…РЎвЂљРЎР‚Р В°Р С”РЎвЂљР С•Р С, РЎвЂЎРЎвЂљР С• РЎС“ РЎвЂљР С‘Р С—Р С‘Р В·Р С‘РЎР‚Р С•Р Р†Р В°Р Р…Р Р…РЎвЂ№РЎвЂ¦ РЎвЂљР В°Р В±Р В»Р С‘РЎвЂ .
func NewRawReferenceTable() *RawReferenceTable {
	return &RawReferenceTable{
		Items: make(map[string]json.RawMessage),
	}
}

// Р С›Р В±РЎР‰Р ВµР Т‘Р С‘Р Р…РЎРЏР ВµРЎвЂљ Р Т‘Р В°Р Р…Р Р…РЎвЂ№Р Вµ РЎРѓР ВµРЎР‚Р Р†Р ВµРЎР‚Р В°, Р В·Р В°Р С–РЎР‚РЎС“Р В¶Р В°Р ВµР СРЎвЂ№Р Вµ Р С‘Р В· JSON-РЎвЂћР В°Р в„–Р В»Р С•Р Р† Р С—РЎР‚Р С‘ РЎРѓРЎвЂљР В°РЎР‚РЎвЂљР Вµ.
type ServerData struct {
	Accounts                   *data.Accounts                   // Р вЂ”Р В°Р С–РЎР‚РЎС“Р В¶Р ВµР Р…Р Р…РЎвЂ№Р Вµ РЎС“РЎвЂЎР ВµРЎвЂљР Р…РЎвЂ№Р Вµ Р В·Р В°Р С—Р С‘РЎРѓР С‘ Р С‘Р С–РЎР‚Р С•Р С”Р С•Р Р†.
	Characters                 *data.Characters                 // Р вЂ”Р В°Р С–РЎР‚РЎС“Р В¶Р ВµР Р…Р Р…РЎвЂ№Р Вµ Р С—Р ВµРЎР‚РЎРѓР С•Р Р…Р В°Р В¶Р С‘ Р С‘Р С–РЎР‚Р С•Р С”Р С•Р Р†.
	CosmicObjects              *data.CosmicObjects              // Р вЂ”Р В°Р С–РЎР‚РЎС“Р В¶Р ВµР Р…Р Р…РЎвЂ№Р Вµ РЎРЊР С”Р В·Р ВµР СР С—Р В»РЎРЏРЎР‚РЎвЂ№ Р С•Р В±РЎР‰Р ВµР С”РЎвЂљР С•Р Р† Р СР С‘РЎР‚Р В°.
	CosmicObjectTypes          *data.CosmicObjectTypes          // Р вЂ”Р В°Р С–РЎР‚РЎС“Р В¶Р ВµР Р…Р Р…РЎвЂ№Р в„– РЎРѓР С—РЎР‚Р В°Р Р†Р С•РЎвЂЎР Р…Р С‘Р С” РЎвЂљР С‘Р С—Р С•Р Р† Р С”Р С•РЎРѓР СР С‘РЎвЂЎР ВµРЎРѓР С”Р С‘РЎвЂ¦ Р С•Р В±РЎР‰Р ВµР С”РЎвЂљР С•Р Р†.
	CosmicObjectModels         *data.CosmicObjectModels         // Р вЂ”Р В°Р С–РЎР‚РЎС“Р В¶Р ВµР Р…Р Р…РЎвЂ№Р в„– РЎРѓР С—РЎР‚Р В°Р Р†Р С•РЎвЂЎР Р…Р С‘Р С” Р СР С•Р Т‘Р ВµР В»Р ВµР в„– Р С”Р С•РЎРѓР СР С‘РЎвЂЎР ВµРЎРѓР С”Р С‘РЎвЂ¦ Р С•Р В±РЎР‰Р ВµР С”РЎвЂљР С•Р Р†.
	ItemTypes                  *data.ItemTypes                  // Р вЂ”Р В°Р С–РЎР‚РЎС“Р В¶Р ВµР Р…Р Р…РЎвЂ№Р в„– РЎРѓР С—РЎР‚Р В°Р Р†Р С•РЎвЂЎР Р…Р С‘Р С” РЎвЂљР С‘Р С—Р С•Р Р† Р С—РЎР‚Р ВµР Т‘Р СР ВµРЎвЂљР С•Р Р†.
	EquipmentGroups            *data.EquipmentGroups            // Р вЂ”Р В°Р С–РЎР‚РЎС“Р В¶Р ВµР Р…Р Р…РЎвЂ№Р Вµ Р С–РЎР‚РЎС“Р С—Р С—РЎвЂ№ Р С•Р В±Р С•РЎР‚РЎС“Р Т‘Р С•Р Р†Р В°Р Р…Р С‘РЎРЏ Р С”Р С•РЎРѓР СР С‘РЎвЂЎР ВµРЎРѓР С”Р С‘РЎвЂ¦ Р С•Р В±РЎР‰Р ВµР С”РЎвЂљР С•Р Р†.
	ItemGroups                 *data.ItemGroups                 // Р вЂ”Р В°Р С–РЎР‚РЎС“Р В¶Р ВµР Р…Р Р…РЎвЂ№Р Вµ Р С–РЎР‚РЎС“Р С—Р С—РЎвЂ№ Р С—РЎР‚Р ВµР Т‘Р СР ВµРЎвЂљР С•Р Р† Р Р†Р Р…РЎС“РЎвЂљРЎР‚Р С‘ Р С”Р С•Р Р…РЎвЂљР ВµР в„–Р Р…Р ВµРЎР‚Р С•Р Р†.
	Assemblies                 *data.Assemblies                 // Р вЂ”Р В°Р С–РЎР‚РЎС“Р В¶Р ВµР Р…Р Р…РЎвЂ№Р в„– РЎРѓР С—РЎР‚Р В°Р Р†Р С•РЎвЂЎР Р…Р С‘Р С” РЎРѓР В±Р С•РЎР‚Р С•Р С” Р С”Р С•РЎРѓР СР С‘РЎвЂЎР ВµРЎРѓР С”Р С‘РЎвЂ¦ Р С•Р В±РЎР‰Р ВµР С”РЎвЂљР С•Р Р†.
	AssemblyEquipmentGroups    *data.AssemblyEquipmentGroups    // Р вЂ”Р В°Р С–РЎР‚РЎС“Р В¶Р ВµР Р…Р Р…РЎвЂ№Р в„– РЎРѓР С—РЎР‚Р В°Р Р†Р С•РЎвЂЎР Р…Р С‘Р С” Р С•Р В±Р С•РЎР‚РЎС“Р Т‘Р С•Р Р†Р В°Р Р…Р С‘РЎРЏ РЎРѓР В±Р С•РЎР‚Р С•Р С”.
	NpcClans                   *RawReferenceTable               // Р вЂ”Р В°Р С–РЎР‚РЎС“Р В¶Р ВµР Р…Р Р…РЎвЂ№Р в„– РЎРѓР С—РЎР‚Р В°Р Р†Р С•РЎвЂЎР Р…Р С‘Р С” NPC-Р С”Р В»Р В°Р Р…Р С•Р Р†.
	ItemModels                 *data.ItemModels                 // Р вЂ”Р В°Р С–РЎР‚РЎС“Р В¶Р ВµР Р…Р Р…РЎвЂ№Р в„– РЎРѓР С—РЎР‚Р В°Р Р†Р С•РЎвЂЎР Р…Р С‘Р С” Р СР С•Р Т‘Р ВµР В»Р ВµР в„– Р С—РЎР‚Р ВµР Т‘Р СР ВµРЎвЂљР С•Р Р†.
	Blueprints                 *RawReferenceTable               // Р вЂ”Р В°Р С–РЎР‚РЎС“Р В¶Р ВµР Р…Р Р…РЎвЂ№Р в„– РЎРѓР С—РЎР‚Р В°Р Р†Р С•РЎвЂЎР Р…Р С‘Р С” РЎвЂЎР ВµРЎР‚РЎвЂљР ВµР В¶Р ВµР в„– Р С•Р В±РЎР‰Р ВµР С”РЎвЂљР С•Р Р†.
	BlueprintComponents        *RawReferenceTable               // Р вЂ”Р В°Р С–РЎР‚РЎС“Р В¶Р ВµР Р…Р Р…РЎвЂ№Р в„– РЎРѓР С—РЎР‚Р В°Р Р†Р С•РЎвЂЎР Р…Р С‘Р С” Р С”Р С•Р СР С—Р С•Р Р…Р ВµР Р…РЎвЂљР С•Р Р† РЎвЂЎР ВµРЎР‚РЎвЂљР ВµР В¶Р ВµР в„–.
	Schemas                    *RawReferenceTable               // Р вЂ”Р В°Р С–РЎР‚РЎС“Р В¶Р ВµР Р…Р Р…РЎвЂ№Р в„– РЎРѓР С—РЎР‚Р В°Р Р†Р С•РЎвЂЎР Р…Р С‘Р С” РЎРѓРЎвЂ¦Р ВµР С Р С—РЎР‚Р ВµР Т‘Р СР ВµРЎвЂљР С•Р Р†.
	SchemaComponents           *RawReferenceTable               // Р вЂ”Р В°Р С–РЎР‚РЎС“Р В¶Р ВµР Р…Р Р…РЎвЂ№Р в„– РЎРѓР С—РЎР‚Р В°Р Р†Р С•РЎвЂЎР Р…Р С‘Р С” Р С”Р С•Р СР С—Р С•Р Р…Р ВµР Р…РЎвЂљР С•Р Р† РЎРѓРЎвЂ¦Р ВµР С.
	Chats                      *data.Chats                      // Р вЂ”Р В°Р С–РЎР‚РЎС“Р В¶Р ВµР Р…Р Р…РЎвЂ№Р Вµ РЎвЂЎР В°РЎвЂљРЎвЂ№ Р С‘Р С–РЎР‚Р С•Р Р†Р С•Р С–Р С• Р СР С‘РЎР‚Р В°.
	ChatMembers                *data.ChatMembers                // Р вЂ”Р В°Р С–РЎР‚РЎС“Р В¶Р ВµР Р…Р Р…РЎвЂ№Р Вµ РЎС“РЎвЂЎР В°РЎРѓРЎвЂљР Р…Р С‘Р С”Р С‘ РЎвЂЎР В°РЎвЂљР С•Р Р†.
	CommunityTypes             *data.CommunityTypes             // Р вЂ”Р В°Р С–РЎР‚РЎС“Р В¶Р ВµР Р…Р Р…РЎвЂ№Р в„– РЎРѓР С—РЎР‚Р В°Р Р†Р С•РЎвЂЎР Р…Р С‘Р С” РЎвЂљР С‘Р С—Р С•Р Р† РЎРѓР С•Р С•Р В±РЎвЂ°Р ВµРЎРѓРЎвЂљР Р†.
	CommunityChatRoles         *data.CommunityChatRoles         // Р вЂ”Р В°Р С–РЎР‚РЎС“Р В¶Р ВµР Р…Р Р…РЎвЂ№Р в„– РЎРѓР С—РЎР‚Р В°Р Р†Р С•РЎвЂЎР Р…Р С‘Р С” РЎР‚Р С•Р В»Р ВµР в„– Р Р† РЎвЂЎР В°РЎвЂљР В°РЎвЂ¦ РЎРѓР С•Р С•Р В±РЎвЂ°Р ВµРЎРѓРЎвЂљР Р†.
	Messages                   *data.Messages                   // Р вЂ”Р В°Р С–РЎР‚РЎС“Р В¶Р ВµР Р…Р Р…РЎвЂ№Р Вµ РЎРѓР С•Р С•Р В±РЎвЂ°Р ВµР Р…Р С‘РЎРЏ РЎвЂЎР В°РЎвЂљР С•Р Р†.
	MessageReads               *data.MessageReads               // Р вЂ”Р В°Р С–РЎР‚РЎС“Р В¶Р ВµР Р…Р Р…РЎвЂ№Р Вµ Р С—Р С•Р В·Р С‘РЎвЂ Р С‘Р С‘ РЎвЂЎРЎвЂљР ВµР Р…Р С‘РЎРЏ РЎРѓР С•Р С•Р В±РЎвЂ°Р ВµР Р…Р С‘Р в„– Р С—Р ВµРЎР‚РЎРѓР С•Р Р…Р В°Р В¶Р В°Р СР С‘.
	MessageTypes               *data.MessageTypes               // Р вЂ”Р В°Р С–РЎР‚РЎС“Р В¶Р ВµР Р…Р Р…РЎвЂ№Р в„– РЎРѓР С—РЎР‚Р В°Р Р†Р С•РЎвЂЎР Р…Р С‘Р С” РЎвЂљР С‘Р С—Р С•Р Р† РЎРѓР С•Р С•Р В±РЎвЂ°Р ВµР Р…Р С‘Р в„–.
	ActionTypes                *data.ActionTypes                // Р вЂ”Р В°Р С–РЎР‚РЎС“Р В¶Р ВµР Р…Р Р…РЎвЂ№Р в„– РЎРѓР С—РЎР‚Р В°Р Р†Р С•РЎвЂЎР Р…Р С‘Р С” Р С‘Р С–РЎР‚Р С•Р Р†РЎвЂ№РЎвЂ¦ Р Т‘Р ВµР в„–РЎРѓРЎвЂљР Р†Р С‘Р в„–.
	InputEventTypes            *data.InputEventTypes            // Р вЂ”Р В°Р С–РЎР‚РЎС“Р В¶Р ВµР Р…Р Р…РЎвЂ№Р в„– РЎРѓР С—РЎР‚Р В°Р Р†Р С•РЎвЂЎР Р…Р С‘Р С” РЎРѓР С•Р В±РЎвЂ№РЎвЂљР С‘Р в„– Р Р†Р Р†Р С•Р Т‘Р В°.
	DefaultActionInputSettings *data.DefaultActionInputSettings // Р вЂ”Р В°Р С–РЎР‚РЎС“Р В¶Р ВµР Р…Р Р…РЎвЂ№Р Вµ Р С—РЎР‚Р С‘Р Р†РЎРЏР В·Р С”Р С‘ Р Р†Р Р†Р С•Р Т‘Р В° Р С—Р С• РЎС“Р СР С•Р В»РЎвЂЎР В°Р Р…Р С‘РЎР‹.
	AccountActionInputSettings *data.AccountActionInputSettings // Р вЂ”Р В°Р С–РЎР‚РЎС“Р В¶Р ВµР Р…Р Р…РЎвЂ№Р Вµ Р В°Р С”Р С”Р В°РЎС“Р Р…РЎвЂљР Р…РЎвЂ№Р Вµ Р С—РЎР‚Р С‘Р Р†РЎРЏР В·Р С”Р С‘ Р Р†Р Р†Р С•Р Т‘Р В°.
	TaskTypes                  *data.TaskTypes                  // РЎРїСЂР°РІРѕС‡РЅРёРє С‚РёРїРѕРІ Р·Р°РґР°РЅРёР№.
	Tasks                      *data.Tasks                      // РЎРѕС…СЂР°РЅРµРЅРЅС‹Рµ Р·Р°РґР°РЅРёСЏ РѕР±РѕСЂСѓРґРѕРІР°РЅРёСЏ.
	TaskItemGroups             *data.TaskItemGroups             // РџСЂРµРґРјРµС‚С‹, Р·Р°СЂРµР·РµСЂРІРёСЂРѕРІР°РЅРЅС‹Рµ Р·Р°РґР°РЅРёСЏРјРё.
	Implementers               *data.Implementers               // РСЃРїРѕР»РЅРёС‚РµР»Рё С‚РёРїРѕРІ Р·Р°РґР°РЅРёР№.
}

// Р вЂ”Р В°Р С–РЎР‚РЎС“Р В¶Р В°Р ВµРЎвЂљ Р Р†РЎРѓР Вµ JSON-РЎвЂћР В°Р в„–Р В»РЎвЂ№ Р Т‘Р В°Р Р…Р Р…РЎвЂ№РЎвЂ¦ РЎРѓР ВµРЎР‚Р Р†Р ВµРЎР‚Р В° Р С‘Р В· РЎС“Р С”Р В°Р В·Р В°Р Р…Р Р…Р С•Р С–Р С• РЎР‚Р В°Р В±Р С•РЎвЂЎР ВµР С–Р С• Р С”Р В°РЎвЂљР В°Р В»Р С•Р С–Р В°.
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

	itemTypes := data.NewItemTypes()
	if err := itemTypes.LoadFromFile(filepath.Join(dataDirectory, itemTypesFileName)); err != nil {
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
	actionTypes, err := loadOptionalActionTypes(filepath.Join(dataDirectory, actionTypesFileName))
	if err != nil {
		return nil, err
	}
	inputEventTypes, err := loadOptionalInputEventTypes(filepath.Join(dataDirectory, inputEventTypesFileName))
	if err != nil {
		return nil, err
	}
	defaultActionInputSettings, err := loadOptionalDefaultActionInputSettings(filepath.Join(dataDirectory, defaultActionInputSettingsFileName))
	if err != nil {
		return nil, err
	}
	accountActionInputSettings, err := loadOptionalAccountActionInputSettings(filepath.Join(dataDirectory, accountActionInputSettingsFileName))
	if err != nil {
		return nil, err
	}
	taskTypes, err := loadOptionalTaskTypes(filepath.Join(dataDirectory, taskTypesFileName))
	if err != nil {
		return nil, err
	}
	tasks, err := loadOptionalTasks(filepath.Join(dataDirectory, tasksFileName))
	if err != nil {
		return nil, err
	}
	taskItemGroups, err := loadOptionalTaskItemGroups(filepath.Join(dataDirectory, taskItemGroupsFileName))
	if err != nil {
		return nil, err
	}
	implementers, err := loadOptionalImplementers(filepath.Join(dataDirectory, implementersFileName))
	if err != nil {
		return nil, err
	}

	return &ServerData{
		Accounts:                   accounts,
		Characters:                 characters,
		CosmicObjects:              cosmicObjects,
		CosmicObjectTypes:          cosmicObjectTypes,
		CosmicObjectModels:         cosmicObjectModels,
		ItemTypes:                  itemTypes,
		EquipmentGroups:            equipmentGroups,
		ItemGroups:                 itemGroups,
		Assemblies:                 assemblies,
		AssemblyEquipmentGroups:    assemblyEquipmentGroups,
		NpcClans:                   npcClans,
		ItemModels:                 itemModels,
		Blueprints:                 blueprints,
		BlueprintComponents:        blueprintComponents,
		Schemas:                    schemas,
		SchemaComponents:           schemaComponents,
		Chats:                      chats,
		ChatMembers:                chatMembers,
		CommunityTypes:             communityTypes,
		CommunityChatRoles:         communityChatRoles,
		Messages:                   messages,
		MessageReads:               messageReads,
		MessageTypes:               messageTypes,
		ActionTypes:                actionTypes,
		InputEventTypes:            inputEventTypes,
		DefaultActionInputSettings: defaultActionInputSettings,
		AccountActionInputSettings: accountActionInputSettings,
		TaskTypes:                  taskTypes,
		Tasks:                      tasks,
		TaskItemGroups:             taskItemGroups,
		Implementers:               implementers,
	}, nil
}

// Р вЂ”Р В°Р С–РЎР‚РЎС“Р В¶Р В°Р ВµРЎвЂљ Р Р…Р ВµР С•Р В±РЎРЏР В·Р В°РЎвЂљР ВµР В»РЎРЉР Р…РЎС“РЎР‹ РЎвЂљР В°Р В±Р В»Р С‘РЎвЂ РЎС“ Р С‘Р В»Р С‘ Р Р†Р С•Р В·Р Р†РЎР‚Р В°РЎвЂ°Р В°Р ВµРЎвЂљ Р С—РЎС“РЎРѓРЎвЂљР С•Р в„– Р С”Р С•Р Р…РЎвЂљР ВµР в„–Р Р…Р ВµРЎР‚ Р Т‘Р С• Р С—Р С•РЎРЏР Р†Р В»Р ВµР Р…Р С‘РЎРЏ РЎвЂћР В°Р в„–Р В»Р В°.
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

// Р вЂ”Р В°Р С–РЎР‚РЎС“Р В¶Р В°Р ВµРЎвЂљ Р Р…Р ВµР С•Р В±РЎРЏР В·Р В°РЎвЂљР ВµР В»РЎРЉР Р…РЎС“РЎР‹ РЎвЂљР В°Р В±Р В»Р С‘РЎвЂ РЎС“ Р С•Р В±Р С•РЎР‚РЎС“Р Т‘Р С•Р Р†Р В°Р Р…Р С‘РЎРЏ Р С•Р В±РЎР‰Р ВµР С”РЎвЂљР С•Р Р† Р С‘Р В»Р С‘ Р Р†Р С•Р В·Р Р†РЎР‚Р В°РЎвЂ°Р В°Р ВµРЎвЂљ Р С—РЎС“РЎРѓРЎвЂљР С•Р в„– Р С”Р С•Р Р…РЎвЂљР ВµР в„–Р Р…Р ВµРЎР‚ Р Т‘Р С• Р С—Р С•РЎРЏР Р†Р В»Р ВµР Р…Р С‘РЎРЏ РЎвЂћР В°Р в„–Р В»Р В°.
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

// Р вЂ”Р В°Р С–РЎР‚РЎС“Р В¶Р В°Р ВµРЎвЂљ Р Р…Р ВµР С•Р В±РЎРЏР В·Р В°РЎвЂљР ВµР В»РЎРЉР Р…РЎС“РЎР‹ РЎвЂљР В°Р В±Р В»Р С‘РЎвЂ РЎС“ Р С–РЎР‚РЎС“Р С—Р С— Р С—РЎР‚Р ВµР Т‘Р СР ВµРЎвЂљР С•Р Р† Р С‘Р В»Р С‘ Р Р†Р С•Р В·Р Р†РЎР‚Р В°РЎвЂ°Р В°Р ВµРЎвЂљ Р С—РЎС“РЎРѓРЎвЂљР С•Р в„– Р С”Р С•Р Р…РЎвЂљР ВµР в„–Р Р…Р ВµРЎР‚ Р Т‘Р С• Р С—Р С•РЎРЏР Р†Р В»Р ВµР Р…Р С‘РЎРЏ РЎвЂћР В°Р в„–Р В»Р В°.
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

// Р вЂ”Р В°Р С–РЎР‚РЎС“Р В¶Р В°Р ВµРЎвЂљ Р Р…Р ВµР С•Р В±РЎРЏР В·Р В°РЎвЂљР ВµР В»РЎРЉР Р…РЎС“РЎР‹ РЎвЂљР В°Р В±Р В»Р С‘РЎвЂ РЎС“ Р СР С•Р Т‘Р ВµР В»Р ВµР в„– Р С—РЎР‚Р ВµР Т‘Р СР ВµРЎвЂљР С•Р Р† Р С‘Р В»Р С‘ Р Р†Р С•Р В·Р Р†РЎР‚Р В°РЎвЂ°Р В°Р ВµРЎвЂљ Р С—РЎС“РЎРѓРЎвЂљР С•Р в„– Р С”Р С•Р Р…РЎвЂљР ВµР в„–Р Р…Р ВµРЎР‚ Р Т‘Р С• Р С—Р С•РЎРЏР Р†Р В»Р ВµР Р…Р С‘РЎРЏ РЎвЂћР В°Р в„–Р В»Р В°.
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

// Р вЂ”Р В°Р С–РЎР‚РЎС“Р В¶Р В°Р ВµРЎвЂљ Р Р…Р ВµР С•Р В±РЎРЏР В·Р В°РЎвЂљР ВµР В»РЎРЉР Р…РЎС“РЎР‹ РЎвЂљР В°Р В±Р В»Р С‘РЎвЂ РЎС“ РЎРѓР В±Р С•РЎР‚Р С•Р С” Р С‘Р В»Р С‘ Р Р†Р С•Р В·Р Р†РЎР‚Р В°РЎвЂ°Р В°Р ВµРЎвЂљ Р С—РЎС“РЎРѓРЎвЂљР С•Р в„– Р С”Р С•Р Р…РЎвЂљР ВµР в„–Р Р…Р ВµРЎР‚ Р Т‘Р С• Р С—Р С•РЎРЏР Р†Р В»Р ВµР Р…Р С‘РЎРЏ РЎвЂћР В°Р в„–Р В»Р В°.
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

// Р вЂ”Р В°Р С–РЎР‚РЎС“Р В¶Р В°Р ВµРЎвЂљ Р Р…Р ВµР С•Р В±РЎРЏР В·Р В°РЎвЂљР ВµР В»РЎРЉР Р…РЎС“РЎР‹ РЎвЂљР В°Р В±Р В»Р С‘РЎвЂ РЎС“ Р С•Р В±Р С•РЎР‚РЎС“Р Т‘Р С•Р Р†Р В°Р Р…Р С‘РЎРЏ РЎРѓР В±Р С•РЎР‚Р С•Р С” Р С‘Р В»Р С‘ Р Р†Р С•Р В·Р Р†РЎР‚Р В°РЎвЂ°Р В°Р ВµРЎвЂљ Р С—РЎС“РЎРѓРЎвЂљР С•Р в„– Р С”Р С•Р Р…РЎвЂљР ВµР в„–Р Р…Р ВµРЎР‚ Р Т‘Р С• Р С—Р С•РЎРЏР Р†Р В»Р ВµР Р…Р С‘РЎРЏ РЎвЂћР В°Р в„–Р В»Р В°.
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

// Р вЂ”Р В°Р С–РЎР‚РЎС“Р В¶Р В°Р ВµРЎвЂљ Р Р…Р ВµР С•Р В±РЎРЏР В·Р В°РЎвЂљР ВµР В»РЎРЉР Р…РЎС“РЎР‹ РЎвЂљР В°Р В±Р В»Р С‘РЎвЂ РЎС“ РЎвЂЎР В°РЎвЂљР С•Р Р† Р С‘Р В»Р С‘ Р Р†Р С•Р В·Р Р†РЎР‚Р В°РЎвЂ°Р В°Р ВµРЎвЂљ Р С—РЎС“РЎРѓРЎвЂљР С•Р Вµ РЎвЂ¦РЎР‚Р В°Р Р…Р С‘Р В»Р С‘РЎвЂ°Р Вµ Р Т‘Р С• Р С—Р С•РЎРЏР Р†Р В»Р ВµР Р…Р С‘РЎРЏ РЎвЂћР В°Р в„–Р В»Р В°.
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

// Р вЂ”Р В°Р С–РЎР‚РЎС“Р В¶Р В°Р ВµРЎвЂљ Р Р…Р ВµР С•Р В±РЎРЏР В·Р В°РЎвЂљР ВµР В»РЎРЉР Р…РЎС“РЎР‹ РЎвЂљР В°Р В±Р В»Р С‘РЎвЂ РЎС“ РЎС“РЎвЂЎР В°РЎРѓРЎвЂљР Р…Р С‘Р С”Р С•Р Р† РЎвЂЎР В°РЎвЂљР С•Р Р† Р С‘Р В»Р С‘ Р Р†Р С•Р В·Р Р†РЎР‚Р В°РЎвЂ°Р В°Р ВµРЎвЂљ Р С—РЎС“РЎРѓРЎвЂљР С•Р Вµ РЎвЂ¦РЎР‚Р В°Р Р…Р С‘Р В»Р С‘РЎвЂ°Р Вµ Р Т‘Р С• Р С—Р С•РЎРЏР Р†Р В»Р ВµР Р…Р С‘РЎРЏ РЎвЂћР В°Р в„–Р В»Р В°.
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

// Р вЂ”Р В°Р С–РЎР‚РЎС“Р В¶Р В°Р ВµРЎвЂљ Р Р…Р ВµР С•Р В±РЎРЏР В·Р В°РЎвЂљР ВµР В»РЎРЉР Р…РЎвЂ№Р в„– РЎРѓР С—РЎР‚Р В°Р Р†Р С•РЎвЂЎР Р…Р С‘Р С” РЎвЂљР С‘Р С—Р С•Р Р† РЎРѓР С•Р С•Р В±РЎвЂ°Р ВµРЎРѓРЎвЂљР Р† Р С‘Р В»Р С‘ Р Р†Р С•Р В·Р Р†РЎР‚Р В°РЎвЂ°Р В°Р ВµРЎвЂљ Р С—РЎС“РЎРѓРЎвЂљР С•Р Вµ РЎвЂ¦РЎР‚Р В°Р Р…Р С‘Р В»Р С‘РЎвЂ°Р Вµ Р Т‘Р С• Р С—Р С•РЎРЏР Р†Р В»Р ВµР Р…Р С‘РЎРЏ РЎвЂћР В°Р в„–Р В»Р В°.
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

// Р вЂ”Р В°Р С–РЎР‚РЎС“Р В¶Р В°Р ВµРЎвЂљ Р Р…Р ВµР С•Р В±РЎРЏР В·Р В°РЎвЂљР ВµР В»РЎРЉР Р…РЎвЂ№Р в„– РЎРѓР С—РЎР‚Р В°Р Р†Р С•РЎвЂЎР Р…Р С‘Р С” РЎР‚Р С•Р В»Р ВµР в„– РЎвЂЎР В°РЎвЂљР С•Р Р† Р С‘Р В»Р С‘ Р Р†Р С•Р В·Р Р†РЎР‚Р В°РЎвЂ°Р В°Р ВµРЎвЂљ Р С—РЎС“РЎРѓРЎвЂљР С•Р Вµ РЎвЂ¦РЎР‚Р В°Р Р…Р С‘Р В»Р С‘РЎвЂ°Р Вµ Р Т‘Р С• Р С—Р С•РЎРЏР Р†Р В»Р ВµР Р…Р С‘РЎРЏ РЎвЂћР В°Р в„–Р В»Р В°.
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

// Р вЂ”Р В°Р С–РЎР‚РЎС“Р В¶Р В°Р ВµРЎвЂљ Р Р…Р ВµР С•Р В±РЎРЏР В·Р В°РЎвЂљР ВµР В»РЎРЉР Р…РЎС“РЎР‹ РЎвЂљР В°Р В±Р В»Р С‘РЎвЂ РЎС“ РЎРѓР С•Р С•Р В±РЎвЂ°Р ВµР Р…Р С‘Р в„– Р С‘Р В»Р С‘ Р Р†Р С•Р В·Р Р†РЎР‚Р В°РЎвЂ°Р В°Р ВµРЎвЂљ Р С—РЎС“РЎРѓРЎвЂљР С•Р Вµ РЎвЂ¦РЎР‚Р В°Р Р…Р С‘Р В»Р С‘РЎвЂ°Р Вµ Р Т‘Р С• Р С—Р С•РЎРЏР Р†Р В»Р ВµР Р…Р С‘РЎРЏ РЎвЂћР В°Р в„–Р В»Р В°.
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

// Р вЂ”Р В°Р С–РЎР‚РЎС“Р В¶Р В°Р ВµРЎвЂљ Р Р…Р ВµР С•Р В±РЎРЏР В·Р В°РЎвЂљР ВµР В»РЎРЉР Р…РЎС“РЎР‹ РЎвЂљР В°Р В±Р В»Р С‘РЎвЂ РЎС“ Р С—РЎР‚Р С•РЎвЂЎРЎвЂљР ВµР Р…Р С‘Р в„– РЎРѓР С•Р С•Р В±РЎвЂ°Р ВµР Р…Р С‘Р в„– Р С‘Р В»Р С‘ Р Р†Р С•Р В·Р Р†РЎР‚Р В°РЎвЂ°Р В°Р ВµРЎвЂљ Р С—РЎС“РЎРѓРЎвЂљР С•Р Вµ РЎвЂ¦РЎР‚Р В°Р Р…Р С‘Р В»Р С‘РЎвЂ°Р Вµ Р Т‘Р С• Р С—Р С•РЎРЏР Р†Р В»Р ВµР Р…Р С‘РЎРЏ РЎвЂћР В°Р в„–Р В»Р В°.
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

// Р вЂ”Р В°Р С–РЎР‚РЎС“Р В¶Р В°Р ВµРЎвЂљ Р Р…Р ВµР С•Р В±РЎРЏР В·Р В°РЎвЂљР ВµР В»РЎРЉР Р…РЎвЂ№Р в„– РЎРѓР С—РЎР‚Р В°Р Р†Р С•РЎвЂЎР Р…Р С‘Р С” РЎвЂљР С‘Р С—Р С•Р Р† РЎРѓР С•Р С•Р В±РЎвЂ°Р ВµР Р…Р С‘Р в„– Р С‘Р В»Р С‘ Р Р†Р С•Р В·Р Р†РЎР‚Р В°РЎвЂ°Р В°Р ВµРЎвЂљ Р С—РЎС“РЎРѓРЎвЂљР С•Р Вµ РЎвЂ¦РЎР‚Р В°Р Р…Р С‘Р В»Р С‘РЎвЂ°Р Вµ Р Т‘Р С• Р С—Р С•РЎРЏР Р†Р В»Р ВµР Р…Р С‘РЎРЏ РЎвЂћР В°Р в„–Р В»Р В°.
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

// Р вЂ”Р В°Р С–РЎР‚РЎС“Р В¶Р В°Р ВµРЎвЂљ Р Р…Р ВµР С•Р В±РЎРЏР В·Р В°РЎвЂљР ВµР В»РЎРЉР Р…РЎвЂ№Р в„– РЎРѓР С—РЎР‚Р В°Р Р†Р С•РЎвЂЎР Р…Р С‘Р С” Р С‘Р С–РЎР‚Р С•Р Р†РЎвЂ№РЎвЂ¦ Р Т‘Р ВµР в„–РЎРѓРЎвЂљР Р†Р С‘Р в„– Р С‘Р В»Р С‘ Р Р†Р С•Р В·Р Р†РЎР‚Р В°РЎвЂ°Р В°Р ВµРЎвЂљ Р С—РЎС“РЎРѓРЎвЂљР С•Р Вµ РЎвЂ¦РЎР‚Р В°Р Р…Р С‘Р В»Р С‘РЎвЂ°Р Вµ Р Т‘Р С• Р С—Р С•РЎРЏР Р†Р В»Р ВµР Р…Р С‘РЎРЏ РЎвЂћР В°Р в„–Р В»Р В°.
func loadOptionalActionTypes(path string) (*data.ActionTypes, error) {
	types := data.NewActionTypes()
	if err := types.LoadFromFile(path); err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return data.NewActionTypes(), nil
		}
		return nil, err
	}
	return types, nil
}

// Р вЂ”Р В°Р С–РЎР‚РЎС“Р В¶Р В°Р ВµРЎвЂљ Р Р…Р ВµР С•Р В±РЎРЏР В·Р В°РЎвЂљР ВµР В»РЎРЉР Р…РЎвЂ№Р в„– РЎРѓР С—РЎР‚Р В°Р Р†Р С•РЎвЂЎР Р…Р С‘Р С” РЎРѓР С•Р В±РЎвЂ№РЎвЂљР С‘Р в„– Р Р†Р Р†Р С•Р Т‘Р В° Р С‘Р В»Р С‘ Р Р†Р С•Р В·Р Р†РЎР‚Р В°РЎвЂ°Р В°Р ВµРЎвЂљ Р С—РЎС“РЎРѓРЎвЂљР С•Р Вµ РЎвЂ¦РЎР‚Р В°Р Р…Р С‘Р В»Р С‘РЎвЂ°Р Вµ Р Т‘Р С• Р С—Р С•РЎРЏР Р†Р В»Р ВµР Р…Р С‘РЎРЏ РЎвЂћР В°Р в„–Р В»Р В°.
func loadOptionalInputEventTypes(path string) (*data.InputEventTypes, error) {
	types := data.NewInputEventTypes()
	if err := types.LoadFromFile(path); err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return data.NewInputEventTypes(), nil
		}
		return nil, err
	}
	return types, nil
}

// Р вЂ”Р В°Р С–РЎР‚РЎС“Р В¶Р В°Р ВµРЎвЂљ Р Р…Р ВµР С•Р В±РЎРЏР В·Р В°РЎвЂљР ВµР В»РЎРЉР Р…РЎвЂ№Р Вµ Р С—РЎР‚Р С‘Р Р†РЎРЏР В·Р С”Р С‘ Р Р†Р Р†Р С•Р Т‘Р В° Р С—Р С• РЎС“Р СР С•Р В»РЎвЂЎР В°Р Р…Р С‘РЎР‹ Р С‘Р В»Р С‘ Р Р†Р С•Р В·Р Р†РЎР‚Р В°РЎвЂ°Р В°Р ВµРЎвЂљ Р С—РЎС“РЎРѓРЎвЂљР С•Р Вµ РЎвЂ¦РЎР‚Р В°Р Р…Р С‘Р В»Р С‘РЎвЂ°Р Вµ Р Т‘Р С• Р С—Р С•РЎРЏР Р†Р В»Р ВµР Р…Р С‘РЎРЏ РЎвЂћР В°Р в„–Р В»Р В°.
func loadOptionalDefaultActionInputSettings(path string) (*data.DefaultActionInputSettings, error) {
	settings := data.NewDefaultActionInputSettings()
	if err := settings.LoadFromFile(path); err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return data.NewDefaultActionInputSettings(), nil
		}
		return nil, err
	}
	return settings, nil
}

// Р вЂ”Р В°Р С–РЎР‚РЎС“Р В¶Р В°Р ВµРЎвЂљ Р Р…Р ВµР С•Р В±РЎРЏР В·Р В°РЎвЂљР ВµР В»РЎРЉР Р…РЎвЂ№Р Вµ Р В°Р С”Р С”Р В°РЎС“Р Р…РЎвЂљР Р…РЎвЂ№Р Вµ Р С—РЎР‚Р С‘Р Р†РЎРЏР В·Р С”Р С‘ Р Р†Р Р†Р С•Р Т‘Р В° Р С‘Р В»Р С‘ Р Р†Р С•Р В·Р Р†РЎР‚Р В°РЎвЂ°Р В°Р ВµРЎвЂљ Р С—РЎС“РЎРѓРЎвЂљР С•Р Вµ РЎвЂ¦РЎР‚Р В°Р Р…Р С‘Р В»Р С‘РЎвЂ°Р Вµ Р Т‘Р С• Р С—Р С•РЎРЏР Р†Р В»Р ВµР Р…Р С‘РЎРЏ РЎвЂћР В°Р в„–Р В»Р В°.
func loadOptionalAccountActionInputSettings(path string) (*data.AccountActionInputSettings, error) {
	settings := data.NewAccountActionInputSettings()
	if err := settings.LoadFromFile(path); err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return data.NewAccountActionInputSettings(), nil
		}
		return nil, err
	}
	return settings, nil
}

// loadOptionalTaskTypes Р·Р°РіСЂСѓР¶Р°РµС‚ РЅРµРѕР±СЏР·Р°С‚РµР»СЊРЅС‹Р№ СЃРїСЂР°РІРѕС‡РЅРёРє С‚РёРїРѕРІ Р·Р°РґР°РЅРёР№ РёР»Рё РІРѕР·РІСЂР°С‰Р°РµС‚ РїСѓСЃС‚РѕРµ С…СЂР°РЅРёР»РёС‰Рµ.
func loadOptionalTaskTypes(path string) (*data.TaskTypes, error) {
	types := data.NewTaskTypes()
	if err := types.LoadFromFile(path); err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return data.NewTaskTypes(), nil
		}
		return nil, err
	}
	return types, nil
}

// loadOptionalTasks Р·Р°РіСЂСѓР¶Р°РµС‚ РЅРµРѕР±СЏР·Р°С‚РµР»СЊРЅСѓСЋ С‚Р°Р±Р»РёС†Сѓ Р·Р°РґР°РЅРёР№ РёР»Рё РІРѕР·РІСЂР°С‰Р°РµС‚ РїСѓСЃС‚РѕРµ С…СЂР°РЅРёР»РёС‰Рµ.
func loadOptionalTasks(path string) (*data.Tasks, error) {
	tasks := data.NewTasks()
	if err := tasks.LoadFromFile(path); err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return data.NewTasks(), nil
		}
		return nil, err
	}
	return tasks, nil
}

// loadOptionalTaskItemGroups Р·Р°РіСЂСѓР¶Р°РµС‚ РЅРµРѕР±СЏР·Р°С‚РµР»СЊРЅСѓСЋ С‚Р°Р±Р»РёС†Сѓ СЂРµР·РµСЂРІРѕРІ Р·Р°РґР°РЅРёР№ РёР»Рё РІРѕР·РІСЂР°С‰Р°РµС‚ РїСѓСЃС‚РѕРµ С…СЂР°РЅРёР»РёС‰Рµ.
func loadOptionalTaskItemGroups(path string) (*data.TaskItemGroups, error) {
	groups := data.NewTaskItemGroups()
	if err := groups.LoadFromFile(path); err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return data.NewTaskItemGroups(), nil
		}
		return nil, err
	}
	return groups, nil
}

// loadOptionalImplementers Р·Р°РіСЂСѓР¶Р°РµС‚ РЅРµРѕР±СЏР·Р°С‚РµР»СЊРЅСѓСЋ С‚Р°Р±Р»РёС†Сѓ РёСЃРїРѕР»РЅРёС‚РµР»РµР№ Р·Р°РґР°РЅРёР№ РёР»Рё РІРѕР·РІСЂР°С‰Р°РµС‚ РїСѓСЃС‚РѕРµ С…СЂР°РЅРёР»РёС‰Рµ.
func loadOptionalImplementers(path string) (*data.Implementers, error) {
	implementers := data.NewImplementers()
	if err := implementers.LoadFromFile(path); err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return data.NewImplementers(), nil
		}
		return nil, err
	}
	return implementers, nil
}
