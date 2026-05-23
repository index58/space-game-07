package world

import (
	"encoding/json"
	"errors"
	"fmt"
	"math"
	"math/rand"
	"path/filepath"
	"sort"
	"sync"

	"space-game-07-server/internal/data"
	"space-game-07-server/internal/game"
	"space-game-07-server/internal/physics"
	"space-game-07-server/internal/storage"
)

const (
	defaultAccountEmailDomain = "auto.local"
	defaultAccountPassword    = "auto"
	defaultStarterShipAcronym = "ship_bat"
	dockingDurationSeconds    = 10
	dockingProbeDistance      = 10
	simpleDrillAcronym        = "SimpleDrill"
	simpleDrillRayAcronym     = "SimpleDrillRay"
)

// Р РЋР С•Р В±Р С‘РЎР‚Р В°Р ВµРЎвЂљ РЎРѓР С—РЎР‚Р В°Р Р†Р С•РЎвЂЎР Р…Р С‘Р С”Р С‘ Р С‘ Р С‘Р С–РЎР‚Р С•Р Р†РЎвЂ№Р Вµ РЎРѓРЎС“РЎвЂ°Р Р…Р С•РЎРѓРЎвЂљР С‘, Р Р…РЎС“Р В¶Р Р…РЎвЂ№Р Вµ РЎРѓР С‘Р СРЎС“Р В»РЎРЏРЎвЂ Р С‘Р С‘ Р СР С‘РЎР‚Р В°.
type Data struct {
	Accounts                   *data.Accounts                   // Р Р€РЎвЂЎР ВµРЎвЂљР Р…РЎвЂ№Р Вµ Р В·Р В°Р С—Р С‘РЎРѓР С‘, Р Т‘Р С•РЎРѓРЎвЂљРЎС“Р С—Р Р…РЎвЂ№Р Вµ Р С‘Р С–РЎР‚Р С•Р Р†Р С•Р в„– РЎРѓР С‘Р СРЎС“Р В»РЎРЏРЎвЂ Р С‘Р С‘.
	Characters                 *data.Characters                 // Р СџР ВµРЎР‚РЎРѓР С•Р Р…Р В°Р В¶Р С‘, Р Т‘Р С•РЎРѓРЎвЂљРЎС“Р С—Р Р…РЎвЂ№Р Вµ Р С‘Р С–РЎР‚Р С•Р Р†Р С•Р в„– РЎРѓР С‘Р СРЎС“Р В»РЎРЏРЎвЂ Р С‘Р С‘.
	CosmicObjects              *data.CosmicObjects              // Р В­Р С”Р В·Р ВµР СР С—Р В»РЎРЏРЎР‚РЎвЂ№ Р С•Р В±РЎР‰Р ВµР С”РЎвЂљР С•Р Р†, РЎС“РЎвЂЎР В°РЎРѓРЎвЂљР Р†РЎС“РЎР‹РЎвЂ°Р С‘Р Вµ Р Р† Р СР С‘РЎР‚Р Вµ.
	CosmicObjectTypes          *data.CosmicObjectTypes          // Р РЋР С—РЎР‚Р В°Р Р†Р С•РЎвЂЎР Р…Р С‘Р С” РЎвЂљР С‘Р С—Р С•Р Р† Р С•Р В±РЎР‰Р ВµР С”РЎвЂљР С•Р Р† Р Т‘Р В»РЎРЏ Р С—РЎР‚Р В°Р Р†Р С‘Р В» Р СР С‘РЎР‚Р В°.
	CosmicObjectModels         *data.CosmicObjectModels         // Р РЋР С—РЎР‚Р В°Р Р†Р С•РЎвЂЎР Р…Р С‘Р С” Р СР С•Р Т‘Р ВµР В»Р ВµР в„– Р С•Р В±РЎР‰Р ВµР С”РЎвЂљР С•Р Р† Р Т‘Р В»РЎРЏ РЎвЂћР С‘Р В·Р С‘Р С”Р С‘ Р С‘ Р С•РЎвЂљР С•Р В±РЎР‚Р В°Р В¶Р ВµР Р…Р С‘РЎРЏ.
	ItemTypes                  *data.ItemTypes                  // Р РЋР С—РЎР‚Р В°Р Р†Р С•РЎвЂЎР Р…Р С‘Р С” РЎвЂљР С‘Р С—Р С•Р Р† Р С—РЎР‚Р ВµР Т‘Р СР ВµРЎвЂљР С•Р Р† Р Т‘Р В»РЎРЏ РЎРѓР ВµРЎР‚Р Р†Р ВµРЎР‚Р Р…Р С•Р в„– Р В»Р С•Р С–Р С‘Р С”Р С‘.
	ItemModels                 *data.ItemModels                 // Р РЋР С—РЎР‚Р В°Р Р†Р С•РЎвЂЎР Р…Р С‘Р С” Р СР С•Р Т‘Р ВµР В»Р ВµР в„– Р С—РЎР‚Р ВµР Т‘Р СР ВµРЎвЂљР С•Р Р† Р Т‘Р В»РЎРЏ Р С•Р В±Р С•РЎР‚РЎС“Р Т‘Р С•Р Р†Р В°Р Р…Р С‘РЎРЏ Р С‘ РЎРѓР С•Р Т‘Р ВµРЎР‚Р В¶Р С‘Р СР С•Р С–Р С• Р С”Р С•Р Р…РЎвЂљР ВµР в„–Р Р…Р ВµРЎР‚Р С•Р Р†.
	Blueprints                 *storage.RawReferenceTable       // Р РЋР С—РЎР‚Р В°Р Р†Р С•РЎвЂЎР Р…Р С‘Р С” РЎвЂЎР ВµРЎР‚РЎвЂљР ВµР В¶Р ВµР в„– Р С•Р В±РЎР‰Р ВµР С”РЎвЂљР С•Р Р† Р Т‘Р В»РЎРЏ Р С‘Р В·Р С–Р С•РЎвЂљР С•Р Р†Р В»Р ВµР Р…Р С‘РЎРЏ Р Р† Р С”Р С•Р Р…РЎРѓРЎвЂљРЎР‚РЎС“Р С”РЎвЂљР С•РЎР‚Р Вµ.
	BlueprintComponents        *storage.RawReferenceTable       // Р РЋР С—РЎР‚Р В°Р Р†Р С•РЎвЂЎР Р…Р С‘Р С” Р С”Р С•Р СР С—Р С•Р Р…Р ВµР Р…РЎвЂљР С•Р Р† РЎвЂЎР ВµРЎР‚РЎвЂљР ВµР В¶Р ВµР в„– Р С•Р В±РЎР‰Р ВµР С”РЎвЂљР С•Р Р† Р Т‘Р В»РЎРЏ РЎРѓР С—Р С‘РЎРѓР В°Р Р…Р С‘РЎРЏ Р СР В°РЎвЂљР ВµРЎР‚Р С‘Р В°Р В»Р С•Р Р†.
	Schemas                    *storage.RawReferenceTable       // Р РЋР С—РЎР‚Р В°Р Р†Р С•РЎвЂЎР Р…Р С‘Р С” РЎРѓРЎвЂ¦Р ВµР С Р С—РЎР‚Р ВµР Т‘Р СР ВµРЎвЂљР С•Р Р† Р Т‘Р В»РЎРЏ Р С‘Р В·Р С–Р С•РЎвЂљР С•Р Р†Р В»Р ВµР Р…Р С‘РЎРЏ Р Р† Р С”Р С•Р Р…РЎРѓРЎвЂљРЎР‚РЎС“Р С”РЎвЂљР С•РЎР‚Р Вµ.
	SchemaComponents           *storage.RawReferenceTable       // Р РЋР С—РЎР‚Р В°Р Р†Р С•РЎвЂЎР Р…Р С‘Р С” Р С”Р С•Р СР С—Р С•Р Р…Р ВµР Р…РЎвЂљР С•Р Р† РЎРѓРЎвЂ¦Р ВµР С Р С—РЎР‚Р ВµР Т‘Р СР ВµРЎвЂљР С•Р Р† Р Т‘Р В»РЎРЏ РЎРѓР С—Р С‘РЎРѓР В°Р Р…Р С‘РЎРЏ Р СР В°РЎвЂљР ВµРЎР‚Р С‘Р В°Р В»Р С•Р Р†.
	TaskTypes                  *data.TaskTypes                  // Справочник типов заданий.
	Tasks                      *data.Tasks                      // Сохраненные задания оборудования.
	TaskItemGroups             *data.TaskItemGroups             // Зарезервированные предметы заданий.
	Implementers               *data.Implementers               // Исполнители типов заданий.
	EquipmentGroups            *data.EquipmentGroups            // Р вЂњРЎР‚РЎС“Р С—Р С—РЎвЂ№ Р С•Р В±Р С•РЎР‚РЎС“Р Т‘Р С•Р Р†Р В°Р Р…Р С‘РЎРЏ, РЎС“РЎРѓРЎвЂљР В°Р Р…Р С•Р Р†Р В»Р ВµР Р…Р Р…РЎвЂ№Р Вµ Р Р…Р В° Р С•Р В±РЎР‰Р ВµР С”РЎвЂљР В°РЎвЂ¦ Р СР С‘РЎР‚Р В°.
	ItemGroups                 *data.ItemGroups                 // Р вЂњРЎР‚РЎС“Р С—Р С—РЎвЂ№ Р С—РЎР‚Р ВµР Т‘Р СР ВµРЎвЂљР С•Р Р† Р Р†Р Р…РЎС“РЎвЂљРЎР‚Р С‘ Р С”Р С•Р Р…РЎвЂљР ВµР в„–Р Р…Р ВµРЎР‚Р Р…Р С•Р С–Р С• Р С•Р В±Р С•РЎР‚РЎС“Р Т‘Р С•Р Р†Р В°Р Р…Р С‘РЎРЏ.
	Assemblies                 *data.Assemblies                 // Р РЋР С—РЎР‚Р В°Р Р†Р С•РЎвЂЎР Р…Р С‘Р С” РЎРѓР В±Р С•РЎР‚Р С•Р С” Р Т‘Р В»РЎРЏ РЎР‚Р В°РЎРѓРЎвЂЎР ВµРЎвЂљР В° РЎвЂ¦Р В°РЎР‚Р В°Р С”РЎвЂљР ВµРЎР‚Р С‘РЎРѓРЎвЂљР С‘Р С” Р С”Р С•РЎР‚Р В°Р В±Р В»Р ВµР в„–.
	AssemblyEquipmentGroups    *data.AssemblyEquipmentGroups    // Р вЂњРЎР‚РЎС“Р С—Р С—РЎвЂ№ Р С•Р В±Р С•РЎР‚РЎС“Р Т‘Р С•Р Р†Р В°Р Р…Р С‘РЎРЏ, Р В·Р В°Р Т‘Р В°Р Р…Р Р…РЎвЂ№Р Вµ Р Р† РЎРѓР В±Р С•РЎР‚Р С”Р В°РЎвЂ¦.
	Chats                      *data.Chats                      // Р В§Р В°РЎвЂљРЎвЂ№ Р С‘Р С–РЎР‚Р С•Р Р†Р С•Р С–Р С• Р СР С‘РЎР‚Р В°.
	ChatMembers                *data.ChatMembers                // Р Р€РЎвЂЎР В°РЎРѓРЎвЂљР Р…Р С‘Р С”Р С‘ РЎвЂЎР В°РЎвЂљР С•Р Р†.
	CommunityTypes             *data.CommunityTypes             // Р РЋР С—РЎР‚Р В°Р Р†Р С•РЎвЂЎР Р…Р С‘Р С” РЎвЂљР С‘Р С—Р С•Р Р† РЎРѓР С•Р С•Р В±РЎвЂ°Р ВµРЎРѓРЎвЂљР Р†.
	CommunityChatRoles         *data.CommunityChatRoles         // Р РЋР С—РЎР‚Р В°Р Р†Р С•РЎвЂЎР Р…Р С‘Р С” РЎР‚Р С•Р В»Р ВµР в„– Р Р† РЎвЂЎР В°РЎвЂљР В°РЎвЂ¦ РЎРѓР С•Р С•Р В±РЎвЂ°Р ВµРЎРѓРЎвЂљР Р†.
	Messages                   *data.Messages                   // Р РЋР С•Р С•Р В±РЎвЂ°Р ВµР Р…Р С‘РЎРЏ РЎвЂЎР В°РЎвЂљР С•Р Р†.
	MessageReads               *data.MessageReads               // Р СџР С•Р В·Р С‘РЎвЂ Р С‘Р С‘ РЎвЂЎРЎвЂљР ВµР Р…Р С‘РЎРЏ РЎРѓР С•Р С•Р В±РЎвЂ°Р ВµР Р…Р С‘Р в„– Р С—Р ВµРЎР‚РЎРѓР С•Р Р…Р В°Р В¶Р В°Р СР С‘.
	MessageTypes               *data.MessageTypes               // Р РЋР С—РЎР‚Р В°Р Р†Р С•РЎвЂЎР Р…Р С‘Р С” РЎвЂљР С‘Р С—Р С•Р Р† РЎРѓР С•Р С•Р В±РЎвЂ°Р ВµР Р…Р С‘Р в„–.
	ActionTypes                *data.ActionTypes                // Р РЋР С—РЎР‚Р В°Р Р†Р С•РЎвЂЎР Р…Р С‘Р С” Р С‘Р С–РЎР‚Р С•Р Р†РЎвЂ№РЎвЂ¦ Р Т‘Р ВµР в„–РЎРѓРЎвЂљР Р†Р С‘Р в„– Р Т‘Р В»РЎРЏ Р Р…Р В°РЎРѓРЎвЂљРЎР‚Р С•Р ВµР С” Р Р†Р Р†Р С•Р Т‘Р В°.
	InputEventTypes            *data.InputEventTypes            // Р РЋР С—РЎР‚Р В°Р Р†Р С•РЎвЂЎР Р…Р С‘Р С” РЎРѓР С•Р В±РЎвЂ№РЎвЂљР С‘Р в„– Р Р†Р Р†Р С•Р Т‘Р В° Р Т‘Р В»РЎРЏ Р Р…Р В°РЎРѓРЎвЂљРЎР‚Р С•Р ВµР С”.
	DefaultActionInputSettings *data.DefaultActionInputSettings // Р СџРЎР‚Р С‘Р Р†РЎРЏР В·Р С”Р С‘ Р Р†Р Р†Р С•Р Т‘Р В° Р С—Р С• РЎС“Р СР С•Р В»РЎвЂЎР В°Р Р…Р С‘РЎР‹.
	AccountActionInputSettings *data.AccountActionInputSettings // Р СџРЎР‚Р С‘Р Р†РЎРЏР В·Р С”Р С‘ Р Р†Р Р†Р С•Р Т‘Р В°, Р Р†РЎвЂ№Р В±РЎР‚Р В°Р Р…Р Р…РЎвЂ№Р Вµ Р В°Р С”Р С”Р В°РЎС“Р Р…РЎвЂљР В°Р СР С‘.
}

// Р Р€Р С—РЎР‚Р В°Р Р†Р В»РЎРЏР ВµРЎвЂљ Р С—Р С•Р Т‘Р С”Р В»РЎР‹РЎвЂЎР ВµР Р…Р Р…РЎвЂ№Р СР С‘ Р В°Р С”Р С”Р В°РЎС“Р Р…РЎвЂљР В°Р СР С‘, Р Р†Р Р†Р С•Р Т‘Р С•Р С Р С‘Р С–РЎР‚Р С•Р С”Р С•Р Р† Р С‘ Р С—Р С•РЎв‚¬Р В°Р С–Р С•Р Р†Р С•Р в„– РЎРѓР С‘Р СРЎС“Р В»РЎРЏРЎвЂ Р С‘Р ВµР в„– Р С•Р В±РЎР‰Р ВµР С”РЎвЂљР С•Р Р†.
type World struct {
	mu                             sync.Mutex                 // Р вЂ”Р В°РЎвЂ°Р С‘РЎвЂ°Р В°Р ВµРЎвЂљ Р С‘Р В·Р СР ВµР Р…РЎРЏР ВµР СР С•Р Вµ РЎРѓР С•РЎРѓРЎвЂљР С•РЎРЏР Р…Р С‘Р Вµ Р СР С‘РЎР‚Р В° Р С•РЎвЂљ Р С—Р В°РЎР‚Р В°Р В»Р В»Р ВµР В»РЎРЉР Р…РЎвЂ№РЎвЂ¦ Р С–Р С•РЎР‚РЎС“РЎвЂљР С‘Р Р….
	tick                           int64                      // Р СњР С•Р СР ВµРЎР‚ Р С—Р С•РЎРѓР В»Р ВµР Т‘Р Р…Р ВµР С–Р С• Р Р†РЎвЂ№Р С—Р С•Р В»Р Р…Р ВµР Р…Р Р…Р С•Р С–Р С• РЎв‚¬Р В°Р С–Р В° РЎРѓР С‘Р СРЎС“Р В»РЎРЏРЎвЂ Р С‘Р С‘.
	data                           Data                       // Р РЋР С—РЎР‚Р В°Р Р†Р С•РЎвЂЎР Р…Р С‘Р С”Р С‘ Р С‘ Р С‘Р С–РЎР‚Р С•Р Р†РЎвЂ№Р Вµ РЎРѓРЎС“РЎвЂ°Р Р…Р С•РЎРѓРЎвЂљР С‘, Р С”Р С•РЎвЂљР С•РЎР‚РЎвЂ№Р СР С‘ РЎС“Р С—РЎР‚Р В°Р Р†Р В»РЎРЏР ВµРЎвЂљ Р СР С‘РЎР‚.
	accountObjectIDs               map[int64]int64            // Р РЋР Р†РЎРЏР В·РЎРЉ Р С—Р С•Р Т‘Р С”Р В»РЎР‹РЎвЂЎР ВµР Р…Р Р…РЎвЂ№РЎвЂ¦ Р В°Р С”Р С”Р В°РЎС“Р Р…РЎвЂљР С•Р Р† РЎРѓ РЎС“Р С—РЎР‚Р В°Р Р†Р В»РЎРЏР ВµР СРЎвЂ№Р СР С‘ Р С•Р В±РЎР‰Р ВµР С”РЎвЂљР В°Р СР С‘.
	inputs                         map[int64]game.ShipInput   // Р СџР С•РЎРѓР В»Р ВµР Т‘Р Р…Р С‘Р в„– Р С—РЎР‚Р С‘Р Р…РЎРЏРЎвЂљРЎвЂ№Р в„– Р Р†Р Р†Р С•Р Т‘ Р Т‘Р В»РЎРЏ Р С”Р В°Р В¶Р Т‘Р С•Р С–Р С• Р С—Р С•Р Т‘Р С”Р В»РЎР‹РЎвЂЎР ВµР Р…Р Р…Р С•Р С–Р С• Р В°Р С”Р С”Р В°РЎС“Р Р…РЎвЂљР В°.
	mutationAcks                   map[string]int64           // Р СџР С•РЎРѓР В»Р ВµР Т‘Р Р…Р С‘Р в„– Р С•Р В±РЎР‚Р В°Р В±Р С•РЎвЂљР В°Р Р…Р Р…РЎвЂ№Р в„– Р Р…Р С•Р СР ВµРЎР‚ Р С”Р С•Р СР В°Р Р…Р Т‘РЎвЂ№ Р С—Р В°Р Р…Р ВµР В»Р С‘ Р С—Р С• Р В°Р С”Р С”Р В°РЎС“Р Р…РЎвЂљРЎС“ Р С‘ РЎРѓР ВµРЎРѓРЎРѓР С‘Р С‘.
	random                         *rand.Rand                 // Р ВРЎРѓРЎвЂљР С•РЎвЂЎР Р…Р С‘Р С” РЎРѓР В»РЎС“РЎвЂЎР В°Р в„–Р Р…Р С•РЎРѓРЎвЂљР С‘ Р Т‘Р В»РЎРЏ Р Р†Р С•РЎРѓР С—РЎР‚Р С•Р С‘Р В·Р Р†Р С•Р Т‘Р С‘Р СРЎвЂ№РЎвЂ¦ Р С”Р С•Р СР В°Р Р…Р Т‘.
	nextConstructorProductionJobID int64                      // Р РЋР В»Р ВµР Т‘РЎС“РЎР‹РЎвЂ°Р С‘Р в„– Р С‘Р Т‘Р ВµР Р…РЎвЂљР С‘РЎвЂћР С‘Р С”Р В°РЎвЂљР С•РЎР‚ Р В·Р В°Р Т‘Р В°Р Р…Р С‘РЎРЏ Р С‘Р В·Р С–Р С•РЎвЂљР С•Р Р†Р В»Р ВµР Р…Р С‘РЎРЏ.
	constructorProductionJobs      []constructorProductionJob // Р вЂ”Р В°Р Т‘Р В°Р Р…Р С‘РЎРЏ Р С‘Р В·Р С–Р С•РЎвЂљР С•Р Р†Р В»Р ВµР Р…Р С‘РЎРЏ, Р С•Р В¶Р С‘Р Т‘Р В°РЎР‹РЎвЂ°Р С‘Р Вµ Р С‘Р В»Р С‘ Р Р†РЎвЂ№Р С—Р С•Р В»Р Р…РЎРЏРЎР‹РЎвЂ°Р С‘Р ВµРЎРѓРЎРЏ Р Р† Р С”Р С•Р Р…РЎРѓРЎвЂљРЎР‚РЎС“Р С”РЎвЂљР С•РЎР‚Р В°РЎвЂ¦.
	dockingRequests                []dockingRequest           // Р С’Р С”РЎвЂљР С‘Р Р†Р Р…РЎвЂ№Р Вµ Р В·Р В°Р С—РЎР‚Р С•РЎРѓРЎвЂ№ РЎРѓРЎвЂљРЎвЂ№Р С”Р С•Р Р†Р С”Р С‘, Р С•Р В¶Р С‘Р Т‘Р В°РЎР‹РЎвЂ°Р С‘Р Вµ Р С•РЎвЂљР Р†Р ВµРЎвЂљР В°.
	dockingProcesses               []dockingProcess           // Р С’Р С”РЎвЂљР С‘Р Р†Р Р…РЎвЂ№Р Вµ Р В°Р Р†РЎвЂљР С•Р СР В°РЎвЂљР С‘РЎвЂЎР ВµРЎРѓР С”Р С‘Р Вµ РЎРѓРЎвЂљРЎвЂ№Р С”Р С•Р Р†Р С”Р С‘, РЎвЂ¦РЎР‚Р В°Р Р…РЎРЏРЎвЂ°Р С‘Р ВµРЎРѓРЎРЏ РЎвЂљР С•Р В»РЎРЉР С”Р С• Р Р† Р С—Р В°Р СРЎРЏРЎвЂљР С‘.
	landingRequests                []landingRequest           // Р С’Р С”РЎвЂљР С‘Р Р†Р Р…РЎвЂ№Р Вµ Р В·Р В°Р С—РЎР‚Р С•РЎРѓРЎвЂ№ Р С—Р С•РЎРѓР В°Р Т‘Р С”Р С‘ Р С—Р ВµРЎР‚РЎРѓР С•Р Р…Р В°Р В¶Р В°, Р С•Р В¶Р С‘Р Т‘Р В°РЎР‹РЎвЂ°Р С‘Р Вµ Р С•РЎвЂљР Р†Р ВµРЎвЂљР В°.
	dockingEvents                  []game.DockingEvent        // Р СњР В°Р С”Р С•Р С—Р В»Р ВµР Р…Р Р…РЎвЂ№Р Вµ Р С”Р В»Р С‘Р ВµР Р…РЎвЂљРЎРѓР С”Р С‘Р Вµ РЎРѓР С•Р В±РЎвЂ№РЎвЂљР С‘РЎРЏ РЎРѓРЎвЂљРЎвЂ№Р С”Р С•Р Р†Р С”Р С‘ Р Т‘Р С• Р В±Р В»Р С‘Р В¶Р В°Р в„–РЎв‚¬Р ВµР в„– РЎР‚Р В°РЎРѓРЎРѓРЎвЂ№Р В»Р С”Р С‘.
	exchangeRequests               []exchangeRequest          // Ожидающие ответы на запросы обмена.
	exchangeSessions               []exchangeSession          // Открытые и выполняющиеся обмены.
	exchangeEvents                 []game.ExchangeEvent       // Накопленные клиентские события обмена.
}

type constructorProductionJob struct {
	ID                                int64   // Р Р€Р Р…Р С‘Р С”Р В°Р В»РЎРЉР Р…РЎвЂ№Р в„– РЎвЂЎР С‘РЎРѓР В»Р С•Р Р†Р С•Р в„– Р С‘Р Т‘Р ВµР Р…РЎвЂљР С‘РЎвЂћР С‘Р С”Р В°РЎвЂљР С•РЎР‚ Р В·Р В°Р Т‘Р В°Р Р…Р С‘РЎРЏ.
	ConstructorEquipmentGroupID       int64   // Р С™Р С•Р Р…РЎРѓРЎвЂљРЎР‚РЎС“Р С”РЎвЂљР С•РЎР‚, Р С” Р С•РЎвЂЎР ВµРЎР‚Р ВµР Т‘Р С‘ Р С”Р С•РЎвЂљР С•РЎР‚Р С•Р С–Р С• Р С•РЎвЂљР Р…Р С•РЎРѓР С‘РЎвЂљРЎРѓРЎРЏ Р В·Р В°Р Т‘Р В°Р Р…Р С‘Р Вµ.
	MaterialContainerEquipmentGroupID int64   // Р С™Р С•Р Р…РЎвЂљР ВµР в„–Р Р…Р ВµРЎР‚, Р С‘Р В· Р С”Р С•РЎвЂљР С•РЎР‚Р С•Р С–Р С• РЎРѓР С—Р С‘РЎРѓРЎвЂ№Р Р†Р В°РЎР‹РЎвЂљРЎРѓРЎРЏ Р С”Р С•Р СР С—Р С•Р Р…Р ВµР Р…РЎвЂљРЎвЂ№.
	ProductContainerEquipmentGroupID  int64   // Р С™Р С•Р Р…РЎвЂљР ВµР в„–Р Р…Р ВµРЎР‚, Р Р† Р С”Р С•РЎвЂљР С•РЎР‚РЎвЂ№Р в„– Р С”Р В»Р В°Р Т‘Р ВµРЎвЂљРЎРѓРЎРЏ РЎР‚Р ВµР В·РЎС“Р В»РЎРЉРЎвЂљР В°РЎвЂљ.
	QueueType                         string  // Р С›РЎвЂЎР ВµРЎР‚Р ВµР Т‘РЎРЉ Р В·Р В°Р Т‘Р В°Р Р…Р С‘РЎРЏ: Р С•РЎРѓР Р…Р С•Р Р†Р Р…Р В°РЎРЏ Р С‘Р В»Р С‘ Р Р†РЎРѓР С—Р С•Р СР С•Р С–Р В°РЎвЂљР ВµР В»РЎРЉР Р…Р В°РЎРЏ.
	SchemaID                          int64   // Р РЋРЎвЂ¦Р ВµР СР В°, Р С—Р С• Р С”Р С•РЎвЂљР С•РЎР‚Р С•Р в„– Р С‘Р В·Р С–Р С•РЎвЂљР В°Р Р†Р В»Р С‘Р Р†Р В°Р ВµРЎвЂљРЎРѓРЎРЏ Р С—РЎР‚Р ВµР Т‘Р СР ВµРЎвЂљ.
	BlueprintID                       int64   // Р В§Р ВµРЎР‚РЎвЂљРЎвЂР В¶, Р С—Р С• Р С”Р С•РЎвЂљР С•РЎР‚Р С•Р СРЎС“ Р С‘Р В·Р С–Р С•РЎвЂљР В°Р Р†Р В»Р С‘Р Р†Р В°Р ВµРЎвЂљРЎРѓРЎРЏ Р С”Р С•РЎРѓР СР С‘РЎвЂЎР ВµРЎРѓР С”Р С‘Р в„– Р С•Р В±РЎР‰Р ВµР С”РЎвЂљ.
	ProductItemModelID                int64   // Р СљР С•Р Т‘Р ВµР В»РЎРЉ Р С—РЎР‚Р ВµР Т‘Р СР ВµРЎвЂљР В°, Р С”Р С•РЎвЂљР С•РЎР‚РЎвЂ№Р в„– Р С—Р С•Р В»РЎС“РЎвЂЎР С‘РЎвЂљРЎРѓРЎРЏ Р С—Р С•РЎРѓР В»Р Вµ Р В·Р В°Р Р†Р ВµРЎР‚РЎв‚¬Р ВµР Р…Р С‘РЎРЏ.
	ProductCosmicObjectModelID        int64   // Р СљР С•Р Т‘Р ВµР В»РЎРЉ Р С”Р С•РЎРѓР СР С‘РЎвЂЎР ВµРЎРѓР С”Р С•Р С–Р С• Р С•Р В±РЎР‰Р ВµР С”РЎвЂљР В°, Р С”Р С•РЎвЂљР С•РЎР‚РЎвЂ№Р в„– Р С—Р С•РЎРЏР Р†Р С‘РЎвЂљРЎРѓРЎРЏ Р С—Р С•РЎРѓР В»Р Вµ Р В·Р В°Р Р†Р ВµРЎР‚РЎв‚¬Р ВµР Р…Р С‘РЎРЏ.
	ProductCount                      float64 // Р С™Р С•Р В»Р С‘РЎвЂЎР ВµРЎРѓРЎвЂљР Р†Р С• Р С—РЎР‚Р ВµР Т‘Р СР ВµРЎвЂљР С•Р Р†, Р С”Р С•РЎвЂљР С•РЎР‚Р С•Р Вµ Р С—Р С•Р В»РЎС“РЎвЂЎР С‘РЎвЂљРЎРѓРЎРЏ Р С—Р С•РЎРѓР В»Р Вµ Р В·Р В°Р Р†Р ВµРЎР‚РЎв‚¬Р ВµР Р…Р С‘РЎРЏ.
	RemainingBatches                  int64   // РїС—Р…РїС—Р…РїС—Р…РїС—Р…РїС—Р…РїС—Р…РїС—Р…РїС—Р…РїС—Р…РїС—Р… РїС—Р…РїС—Р…РїС—Р…РїС—Р…РїС—Р…РїС—Р…РїС—Р…РїС—Р…, РїС—Р…РїС—Р…РїС—Р…РїС—Р…РїС—Р…РїС—Р…РїС—Р… РїС—Р…РїС—Р…РїС—Р… РїС—Р…РїС—Р…РїС—Р…РїС—Р…РїС—Р…РїС—Р…РїС—Р…РїС—Р… РїС—Р…РїС—Р…РїС—Р…РїС—Р…РїС—Р…РїС—Р…РїС—Р…РїС—Р…РїС—Р… РїС—Р…РїС—Р… РїС—Р…РїС—Р…РїС—Р…РїС—Р…РїС—Р…РїС—Р….
	TotalBatches                      int64   // РїС—Р…РїС—Р…РїС—Р…РїС—Р…РїС—Р… РїС—Р…РїС—Р…РїС—Р…РїС—Р…РїС—Р…РїС—Р…РїС—Р…РїС—Р…РїС—Р…РїС—Р… РїС—Р…РїС—Р…РїС—Р…РїС—Р…РїС—Р…РїС—Р…РїС—Р…РїС—Р…, РїС—Р…РїС—Р…РїС—Р…РїС—Р…РїС—Р…РїС—Р…РїС—Р…РїС—Р…РїС—Р…РїС—Р…РїС—Р…РїС—Р…РїС—Р…РїС—Р…РїС—Р… РїС—Р…РїС—Р… РїС—Р…РїС—Р…РїС—Р…РїС—Р…РїС—Р…РїС—Р….
	RemainingTime                     float64 // Р С›РЎРѓРЎвЂљР В°Р Р†РЎв‚¬Р ВµР ВµРЎРѓРЎРЏ Р Р†РЎР‚Р ВµР СРЎРЏ Р С‘Р В·Р С–Р С•РЎвЂљР С•Р Р†Р В»Р ВµР Р…Р С‘РЎРЏ Р Р† РЎРѓР ВµР С”РЎС“Р Р…Р Т‘Р В°РЎвЂ¦.
	TotalTime                         float64 // Р СџР С•Р В»Р Р…Р С•Р Вµ Р Р†РЎР‚Р ВµР СРЎРЏ Р С‘Р В·Р С–Р С•РЎвЂљР С•Р Р†Р В»Р ВµР Р…Р С‘РЎРЏ Р Р† РЎРѓР ВµР С”РЎС“Р Р…Р Т‘Р В°РЎвЂ¦.
	Running                           bool    // Р СџР С•Р С”Р В°Р В·РЎвЂ№Р Р†Р В°Р ВµРЎвЂљ, РЎвЂЎРЎвЂљР С• Р В·Р В°Р Т‘Р В°Р Р…Р С‘Р Вµ РЎРѓР ВµР в„–РЎвЂЎР В°РЎРѓ Р Р†РЎвЂ№Р С—Р С•Р В»Р Р…РЎРЏР ВµРЎвЂљРЎРѓРЎРЏ.
	ParentJobID                       int64   // РїС—Р…РїС—Р…РїС—Р…РїС—Р…РїС—Р…РїС—Р…РїС—Р…РїС—Р…РїС—Р…РїС—Р…РїС—Р…РїС—Р… РїС—Р…РїС—Р…РїС—Р…РїС—Р…РїС—Р…РїС—Р…, РїС—Р…РїС—Р…РїС—Р… РїС—Р…РїС—Р…РїС—Р…РїС—Р…РїС—Р…РїС—Р…РїС—Р…РїС—Р…РїС—Р…РїС—Р… РїС—Р…РїС—Р…РїС—Р…РїС—Р…РїС—Р…РїС—Р…РїС—Р… РїС—Р…РїС—Р…РїС—Р…РїС—Р…РїС—Р… РїС—Р…РїС—Р…РїС—Р… РїС—Р…РїС—Р…РїС—Р…РїС—Р…РїС—Р…РїС—Р…РїС—Р…РїС—Р…РїС—Р…РїС—Р…РїС—Р…РїС—Р…РїС—Р…РїС—Р…РїС—Р… РїС—Р…РїС—Р…РїС—Р…РїС—Р…РїС—Р…РїС—Р….
}

type dockingProcess struct {
	SenderCosmicObjectID   int64   // Р С›Р В±РЎР‰Р ВµР С”РЎвЂљ, Р С”Р С•РЎвЂљР С•РЎР‚РЎвЂ№Р в„– РЎРѓРЎвЂљР В°Р Р…Р ВµРЎвЂљ Р Р†РЎвЂљР С•РЎР‚Р С•РЎРѓРЎвЂљР ВµР С—Р ВµР Р…Р Р…РЎвЂ№Р С Р С—Р С•РЎРѓР В»Р Вµ Р В·Р В°Р Р†Р ВµРЎР‚РЎв‚¬Р ВµР Р…Р С‘РЎРЏ.
	ReceiverCosmicObjectID int64   // Р С›Р В±РЎР‰Р ВµР С”РЎвЂљ, Р С”Р С•РЎвЂљР С•РЎР‚РЎвЂ№Р в„– РЎРѓРЎвЂљР В°Р Р…Р ВµРЎвЂљ Р С–Р В»Р В°Р Р†Р Р…РЎвЂ№Р С Р С‘Р В»Р С‘ РЎС“Р В¶Р Вµ РЎРЏР Р†Р В»РЎРЏР ВµРЎвЂљРЎРѓРЎРЏ Р С–Р В»Р В°Р Р†Р Р…РЎвЂ№Р С.
	RemainingSeconds       float64 // Р С›РЎРѓРЎвЂљР В°Р Р†РЎв‚¬Р ВµР ВµРЎРѓРЎРЏ Р Р†РЎР‚Р ВµР СРЎРЏ Р В°Р Р†РЎвЂљР С•Р СР В°РЎвЂљР С‘РЎвЂЎР ВµРЎРѓР С”Р С•Р в„– РЎРѓРЎвЂљРЎвЂ№Р С”Р С•Р Р†Р С”Р С‘.
}

type dockingRequest struct {
	SenderCosmicObjectID   int64   // Р С›Р В±РЎР‰Р ВµР С”РЎвЂљ, Р С”Р С•РЎвЂљР С•РЎР‚РЎвЂ№Р в„– Р С•РЎвЂљР С—РЎР‚Р В°Р Р†Р С‘Р В» Р В·Р В°Р С—РЎР‚Р С•РЎРѓ.
	ReceiverCosmicObjectID int64   // Р С›Р В±РЎР‰Р ВµР С”РЎвЂљ, Р С”Р С•РЎвЂљР С•РЎР‚РЎвЂ№Р в„– Р Т‘Р С•Р В»Р В¶Р ВµР Р… Р С—РЎР‚Р С‘Р Р…РЎРЏРЎвЂљРЎРЉ РЎР‚Р ВµРЎв‚¬Р ВµР Р…Р С‘Р Вµ.
	RemainingSeconds       float64 // Р С›РЎРѓРЎвЂљР В°Р Р†РЎв‚¬Р ВµР ВµРЎРѓРЎРЏ Р Р†РЎР‚Р ВµР СРЎРЏ Р С•Р В¶Р С‘Р Т‘Р В°Р Р…Р С‘РЎРЏ Р С•РЎвЂљР Р†Р ВµРЎвЂљР В°.
}

type landingRequest struct {
	CharacterID            int64 // Р СџР ВµРЎР‚РЎРѓР С•Р Р…Р В°Р В¶, Р С”Р С•РЎвЂљР С•РЎР‚РЎвЂ№Р в„– Р С—Р ВµРЎР‚Р ВµРЎРѓР В°Р В¶Р С‘Р Р†Р В°Р ВµРЎвЂљРЎРѓРЎРЏ.
	SenderCosmicObjectID   int64 // Р С›Р В±РЎР‰Р ВµР С”РЎвЂљ Р С•РЎвЂљР С—РЎР‚Р В°Р Р†Р В»Р ВµР Р…Р С‘РЎРЏ, Р С–Р Т‘Р Вµ Р С—Р ВµРЎР‚РЎРѓР С•Р Р…Р В°Р В¶ Р С•РЎРѓРЎвЂљР В°Р ВµРЎвЂљРЎРѓРЎРЏ Р Т‘Р С• РЎР‚Р ВµРЎв‚¬Р ВµР Р…Р С‘РЎРЏ.
	ReceiverCosmicObjectID int64 // Р С›Р В±РЎР‰Р ВµР С”РЎвЂљ Р Р…Р В°Р В·Р Р…Р В°РЎвЂЎР ВµР Р…Р С‘РЎРЏ, Р С”Р С•РЎвЂљР С•РЎР‚РЎвЂ№Р в„– Р Т‘Р С•Р В»Р В¶Р ВµР Р… Р С—РЎР‚Р С‘Р Р…РЎРЏРЎвЂљРЎРЉ РЎР‚Р ВµРЎв‚¬Р ВµР Р…Р С‘Р Вµ.
}

// ControlPanelObjectUpdate Р С•Р С—Р С‘РЎРѓРЎвЂ№Р Р†Р В°Р ВµРЎвЂљ РЎвЂЎР В°РЎРѓРЎвЂљР С‘РЎвЂЎР Р…Р С•Р Вµ Р С‘Р В·Р СР ВµР Р…Р ВµР Р…Р С‘Р Вµ РЎС“Р С—РЎР‚Р В°Р Р†Р В»РЎРЏР ВµР СР С•Р С–Р С• Р С•Р В±РЎР‰Р ВµР С”РЎвЂљР В°.
type ControlPanelObjectUpdate struct {
	Enabled *bool   // Р СњР С•Р Р†Р С•Р Вµ РЎРѓР С•РЎРѓРЎвЂљР С•РЎРЏР Р…Р С‘Р Вµ Р Р†Р С”Р В»РЎР‹РЎвЂЎР ВµР Р…Р С‘РЎРЏ Р С•Р В±РЎР‰Р ВµР С”РЎвЂљР В°, Р ВµРЎРѓР В»Р С‘ Р С•Р Р…Р С• Р СР ВµР Р…РЎРЏР ВµРЎвЂљРЎРѓРЎРЏ.
	Title   *string // Р СњР С•Р Р†Р С•Р Вµ Р С—Р С•Р В»РЎРЉР В·Р С•Р Р†Р В°РЎвЂљР ВµР В»РЎРЉРЎРѓР С”Р С•Р Вµ Р Р…Р В°Р В·Р Р†Р В°Р Р…Р С‘Р Вµ Р С•Р В±РЎР‰Р ВµР С”РЎвЂљР В°, Р ВµРЎРѓР В»Р С‘ Р С•Р Р…Р С• Р СР ВµР Р…РЎРЏР ВµРЎвЂљРЎРѓРЎРЏ.
}

// ControlPanelEquipmentUpdate Р С•Р С—Р С‘РЎРѓРЎвЂ№Р Р†Р В°Р ВµРЎвЂљ РЎвЂЎР В°РЎРѓРЎвЂљР С‘РЎвЂЎР Р…Р С•Р Вµ Р С‘Р В·Р СР ВµР Р…Р ВµР Р…Р С‘Р Вµ Р С–РЎР‚РЎС“Р С—Р С—РЎвЂ№ Р С•Р В±Р С•РЎР‚РЎС“Р Т‘Р С•Р Р†Р В°Р Р…Р С‘РЎРЏ.
type ControlPanelEquipmentUpdate struct {
	Title            *string // Р СњР С•Р Р†Р С•Р Вµ Р С—Р С•Р В»РЎРЉР В·Р С•Р Р†Р В°РЎвЂљР ВµР В»РЎРЉРЎРѓР С”Р С•Р Вµ Р Р…Р В°Р В·Р Р†Р В°Р Р…Р С‘Р Вµ Р С–РЎР‚РЎС“Р С—Р С—РЎвЂ№, Р ВµРЎРѓР В»Р С‘ Р С•Р Р…Р С• Р СР ВµР Р…РЎРЏР ВµРЎвЂљРЎРѓРЎРЏ.
	EquipmentGroupID int64   // Р вЂњРЎР‚РЎС“Р С—Р С—Р В° Р С•Р В±Р С•РЎР‚РЎС“Р Т‘Р С•Р Р†Р В°Р Р…Р С‘РЎРЏ, Р С”Р С•РЎвЂљР С•РЎР‚РЎС“РЎР‹ Р Р…РЎС“Р В¶Р Р…Р С• Р С‘Р В·Р СР ВµР Р…Р С‘РЎвЂљРЎРЉ.
	Enabled          *bool   // Р СњР С•Р Р†Р С•Р Вµ РЎРѓР С•РЎРѓРЎвЂљР С•РЎРЏР Р…Р С‘Р Вµ Р Р†Р С”Р В»РЎР‹РЎвЂЎР ВµР Р…Р С‘РЎРЏ Р С–РЎР‚РЎС“Р С—Р С—РЎвЂ№, Р ВµРЎРѓР В»Р С‘ Р С•Р Р…Р С• Р СР ВµР Р…РЎРЏР ВµРЎвЂљРЎРѓРЎРЏ.
	EnabledCount     *int64  // Р СњР С•Р Р†Р С•Р Вµ Р С”Р С•Р В»Р С‘РЎвЂЎР ВµРЎРѓРЎвЂљР Р†Р С• Р Р†Р С”Р В»РЎР‹РЎвЂЎР ВµР Р…Р Р…РЎвЂ№РЎвЂ¦ Р ВµР Т‘Р С‘Р Р…Р С‘РЎвЂ , Р ВµРЎРѓР В»Р С‘ Р С•Р Р…Р С• Р СР ВµР Р…РЎРЏР ВµРЎвЂљРЎРѓРЎРЏ.
}

// ControlPanelContainerTransfer Р С•Р С—Р С‘РЎРѓРЎвЂ№Р Р†Р В°Р ВµРЎвЂљ Р С—Р ВµРЎР‚Р ВµР Р…Р С•РЎРѓ Р С—РЎР‚Р ВµР Т‘Р СР ВµРЎвЂљР С•Р Р† Р СР ВµР В¶Р Т‘РЎС“ Р С”Р С•Р Р…РЎвЂљР ВµР в„–Р Р…Р ВµРЎР‚Р В°Р СР С‘ РЎвЂљР ВµР С”РЎС“РЎвЂ°Р ВµР С–Р С• Р С•Р В±РЎР‰Р ВµР С”РЎвЂљР В°.
// ControlPanelEquipmentGroupRelationUpdate Р С•Р С—Р С‘РЎРѓРЎвЂ№Р Р†Р В°Р ВµРЎвЂљ РЎРѓР С•РЎвЂ¦РЎР‚Р В°Р Р…Р ВµР Р…Р С‘Р Вµ Р Р†РЎвЂ№Р В±РЎР‚Р В°Р Р…Р Р…Р С•Р в„– РЎРѓР Р†РЎРЏР В·Р В°Р Р…Р Р…Р С•Р в„– Р С–РЎР‚РЎС“Р С—Р С—РЎвЂ№ Р С•Р В±Р С•РЎР‚РЎС“Р Т‘Р С•Р Р†Р В°Р Р…Р С‘РЎРЏ.
type ControlPanelEquipmentGroupRelationUpdate struct {
	EquipmentGroupID        int64  // Р вЂњРЎР‚РЎС“Р С—Р С—Р В° Р С•Р В±Р С•РЎР‚РЎС“Р Т‘Р С•Р Р†Р В°Р Р…Р С‘РЎРЏ, Р Т‘Р В»РЎРЏ Р С”Р С•РЎвЂљР С•РЎР‚Р С•Р в„– РЎРѓР С•РЎвЂ¦РЎР‚Р В°Р Р…РЎРЏР ВµРЎвЂљРЎРѓРЎРЏ Р Р†РЎвЂ№Р В±Р С•РЎР‚.
	RelationTypeAcronym     string // Р вЂ™Р С‘Р Т‘ РЎРѓР Р†РЎРЏР В·Р С‘, Р Р†РЎвЂ№Р В±РЎР‚Р В°Р Р…Р Р…РЎвЂ№Р в„– Р С—Р С• Р Р…Р ВµР С‘Р В·Р СР ВµР Р…РЎРЏР ВµР СР С•Р СРЎС“ РЎРѓРЎвЂљРЎР‚Р С•Р С”Р С•Р Р†Р С•Р СРЎС“ Р С‘Р Т‘Р ВµР Р…РЎвЂљР С‘РЎвЂћР С‘Р С”Р В°РЎвЂљР С•РЎР‚РЎС“.
	RelatedEquipmentGroupID int64  // Р вЂњРЎР‚РЎС“Р С—Р С—Р В° Р С•Р В±Р С•РЎР‚РЎС“Р Т‘Р С•Р Р†Р В°Р Р…Р С‘РЎРЏ, Р Р†РЎвЂ№Р В±РЎР‚Р В°Р Р…Р Р…Р В°РЎРЏ Р С‘Р С–РЎР‚Р С•Р С”Р С•Р С Р Р† РЎРѓР Р†РЎРЏР В·Р В°Р Р…Р Р…Р С•Р в„– Р С—Р В°Р Р…Р ВµР В»Р С‘.
}

type ControlPanelContainerTransfer struct {
	ControllerEquipmentGroupID      int64   // Правая группа контейнеров, управляющая очередью перемещений.
	LeftToRightDirection            bool    // Перемещается ли груз из левого контейнера в правый.
	SourceContainerEquipmentGroupID int64   // Р вЂњРЎР‚РЎС“Р С—Р С—Р В° Р С”Р С•Р Р…РЎвЂљР ВµР в„–Р Р…Р ВµРЎР‚Р С•Р Р†, Р С‘Р В· Р С”Р С•РЎвЂљР С•РЎР‚Р С•Р в„– Р С—Р ВµРЎР‚Р ВµР Р…Р С•РЎРѓРЎРЏРЎвЂљРЎРѓРЎРЏ Р Р†РЎРѓР Вµ Р С—РЎР‚Р ВµР Т‘Р СР ВµРЎвЂљРЎвЂ№.
	TargetContainerEquipmentGroupID int64   // Р вЂњРЎР‚РЎС“Р С—Р С—Р В° Р С”Р С•Р Р…РЎвЂљР ВµР в„–Р Р…Р ВµРЎР‚Р С•Р Р†, Р Р† Р С”Р С•РЎвЂљР С•РЎР‚РЎС“РЎР‹ Р С—Р ВµРЎР‚Р ВµР Р…Р С•РЎРѓРЎРЏРЎвЂљРЎРѓРЎРЏ Р Р†РЎРѓР Вµ Р С—РЎР‚Р ВµР Т‘Р СР ВµРЎвЂљРЎвЂ№.
	ItemGroupIDs                    []int64 // Р вЂњРЎР‚РЎС“Р С—Р С—РЎвЂ№ Р С—РЎР‚Р ВµР Т‘Р СР ВµРЎвЂљР С•Р Р†, Р Р†РЎвЂ№Р В±РЎР‚Р В°Р Р…Р Р…РЎвЂ№Р Вµ Р Т‘Р В»РЎРЏ Р С—Р ВµРЎР‚Р ВµР Р…Р С•РЎРѓР В°.
	Amount                          float64 // Р С™Р С•Р В»Р С‘РЎвЂЎР ВµРЎРѓРЎвЂљР Р†Р С• Р С—РЎР‚Р ВµР Т‘Р СР ВµРЎвЂљР С•Р Р† Р Т‘Р В»РЎРЏ РЎвЂЎР В°РЎРѓРЎвЂљР С‘РЎвЂЎР Р…Р С•Р С–Р С• Р С—Р ВµРЎР‚Р ВµР Р…Р С•РЎРѓР В° Р С•Р Т‘Р Р…Р С•Р в„– Р Р†РЎвЂ№Р В±РЎР‚Р В°Р Р…Р Р…Р С•Р в„– Р С–РЎР‚РЎС“Р С—Р С—РЎвЂ№.
}

// Р РЋР С•Р В·Р Т‘Р В°Р ВµРЎвЂљ Р СР С‘РЎР‚ Р С—Р С•Р Р†Р ВµРЎР‚РЎвЂ¦ РЎС“Р В¶Р Вµ Р В·Р В°Р С–РЎР‚РЎС“Р В¶Р ВµР Р…Р Р…РЎвЂ№РЎвЂ¦ РЎРѓР ВµРЎР‚Р Р†Р ВµРЎР‚Р Р…РЎвЂ№РЎвЂ¦ Р Т‘Р В°Р Р…Р Р…РЎвЂ№РЎвЂ¦.
// ControlPanelFuelTransfer Р С•Р С—Р С‘РЎРѓРЎвЂ№Р Р†Р В°Р ВµРЎвЂљ Р С—Р ВµРЎР‚Р ВµР Р…Р С•РЎРѓ РЎвЂљР С•Р С—Р В»Р С‘Р Р†Р В° Р СР ВµР В¶Р Т‘РЎС“ Р С”Р С•Р Р…РЎвЂљР ВµР в„–Р Р…Р ВµРЎР‚Р С•Р С Р С‘ Р В±Р В°Р С”Р С•Р С РЎвЂљР ВµР С”РЎС“РЎвЂ°Р ВµР С–Р С• Р С•Р В±РЎР‰Р ВµР С”РЎвЂљР В°.
type ControlPanelFuelTransfer struct {
	ContainerEquipmentGroupID int64   // Р вЂњРЎР‚РЎС“Р С—Р С—Р В° Р С”Р С•Р Р…РЎвЂљР ВµР в„–Р Р…Р ВµРЎР‚Р С•Р Р†, Р С‘Р В· Р С”Р С•РЎвЂљР С•РЎР‚Р С•Р в„– Р В±Р ВµРЎР‚Р ВµРЎвЂљРЎРѓРЎРЏ Р С‘Р В»Р С‘ Р С”РЎС“Р Т‘Р В° РЎРѓР В»Р С‘Р Р†Р В°Р ВµРЎвЂљРЎРѓРЎРЏ РЎвЂљР С•Р С—Р В»Р С‘Р Р†Р С•.
	FuelTankEquipmentGroupID  int64   // Р вЂњРЎР‚РЎС“Р С—Р С—Р В° РЎвЂљР С•Р С—Р В»Р С‘Р Р†Р Р…РЎвЂ№РЎвЂ¦ Р В±Р В°Р С”Р С•Р Р†, РЎРѓ Р С”Р С•РЎвЂљР С•РЎР‚Р С•Р в„– РЎР‚Р В°Р В±Р С•РЎвЂљР В°Р ВµРЎвЂљ Р С‘Р С–РЎР‚Р С•Р С”.
	ItemGroupIDs              []int64 // Р вЂњРЎР‚РЎС“Р С—Р С—РЎвЂ№ РЎвЂљР С•Р С—Р В»Р С‘Р Р†Р В°, Р Р†РЎвЂ№Р В±РЎР‚Р В°Р Р…Р Р…РЎвЂ№Р Вµ Р Т‘Р В»РЎРЏ Р В·Р В°Р В»Р С‘Р Р†Р С”Р С‘ Р С‘Р В· Р С”Р С•Р Р…РЎвЂљР ВµР в„–Р Р…Р ВµРЎР‚Р В°.
	Amount                    float64 // Р С™Р С•Р В»Р С‘РЎвЂЎР ВµРЎРѓРЎвЂљР Р†Р С• РЎвЂљР С•Р С—Р В»Р С‘Р Р†Р В° Р Т‘Р В»РЎРЏ РЎРѓР В»Р С‘Р Р†Р В° Р С‘Р В· Р В±Р В°Р С”Р В° Р Р† Р С”Р С•Р Р…РЎвЂљР ВµР в„–Р Р…Р ВµРЎР‚.
}

// ControlPanelConstructorProduceItem Р С•Р С—Р С‘РЎРѓРЎвЂ№Р Р†Р В°Р ВµРЎвЂљ Р С‘Р В·Р С–Р С•РЎвЂљР С•Р Р†Р В»Р ВµР Р…Р С‘Р Вµ Р С—РЎР‚Р ВµР Т‘Р СР ВµРЎвЂљР В° Р С—Р С• РЎРѓРЎвЂ¦Р ВµР СР Вµ Р С”Р С•Р Р…РЎРѓРЎвЂљРЎР‚РЎС“Р С”РЎвЂљР С•РЎР‚Р В°.
// ControlPanelItemDeconstruction описывает запуск деконструкции выбранных предметов.
type ControlPanelItemDeconstruction struct {
	DeconstructorEquipmentGroupID   int64   // Группа деконструктора, которая управляет очередью заданий.
	SourceContainerEquipmentGroupID int64   // Контейнер с предметами, выбранный в правой части панели.
	TargetContainerEquipmentGroupID int64   // Контейнер результата, выбранный в левой части панели.
	ItemGroupIDs                    []int64 // Строки предметов, выбранные для деконструкции.
	Amount                          float64 // Максимальное количество предметов одной выбранной строки для деконструкции.
}

type ControlPanelConstructorProduceItem struct {
	ConstructorEquipmentGroupID       int64 // Р вЂњРЎР‚РЎС“Р С—Р С—Р В° Р С”Р С•Р Р…РЎРѓРЎвЂљРЎР‚РЎС“Р С”РЎвЂљР С•РЎР‚Р С•Р Р†, Р С”Р С•РЎвЂљР С•РЎР‚Р В°РЎРЏ Р Р†РЎвЂ№Р С—Р С•Р В»Р Р…РЎРЏР ВµРЎвЂљ Р С‘Р В·Р С–Р С•РЎвЂљР С•Р Р†Р В»Р ВµР Р…Р С‘Р Вµ.
	MaterialContainerEquipmentGroupID int64 // Р С™Р С•Р Р…РЎвЂљР ВµР в„–Р Р…Р ВµРЎР‚, Р С‘Р В· Р С”Р С•РЎвЂљР С•РЎР‚Р С•Р С–Р С• РЎРѓР С—Р С‘РЎРѓРЎвЂ№Р Р†Р В°РЎР‹РЎвЂљРЎРѓРЎРЏ Р С”Р С•Р СР С—Р С•Р Р…Р ВµР Р…РЎвЂљРЎвЂ№ РЎРѓРЎвЂ¦Р ВµР СРЎвЂ№.
	ProductContainerEquipmentGroupID  int64 // Р С™Р С•Р Р…РЎвЂљР ВµР в„–Р Р…Р ВµРЎР‚, Р Р† Р С”Р С•РЎвЂљР С•РЎР‚РЎвЂ№Р в„– Р С”Р В»Р В°Р Т‘Р ВµРЎвЂљРЎРѓРЎРЏ Р С–Р С•РЎвЂљР С•Р Р†Р В°РЎРЏ Р С—РЎР‚Р С•Р Т‘РЎС“Р С”РЎвЂ Р С‘РЎРЏ.
	SchemaID                          int64 // Р РЋРЎвЂ¦Р ВµР СР В° Р С—РЎР‚Р ВµР Т‘Р СР ВµРЎвЂљР В°, Р Р†РЎвЂ№Р В±РЎР‚Р В°Р Р…Р Р…Р В°РЎРЏ Р С‘Р С–РЎР‚Р С•Р С”Р С•Р С Р Т‘Р В»РЎРЏ Р С‘Р В·Р С–Р С•РЎвЂљР С•Р Р†Р В»Р ВµР Р…Р С‘РЎРЏ.
	BlueprintID                       int64 // Р В§Р ВµРЎР‚РЎвЂљРЎвЂР В¶ Р С•Р В±РЎР‰Р ВµР С”РЎвЂљР В°, Р Р†РЎвЂ№Р В±РЎР‚Р В°Р Р…Р Р…РЎвЂ№Р в„– Р С‘Р С–РЎР‚Р С•Р С”Р С•Р С Р Т‘Р В»РЎРЏ Р С‘Р В·Р С–Р С•РЎвЂљР С•Р Р†Р В»Р ВµР Р…Р С‘РЎРЏ.
	Amount                            int64 // Р С™Р С•Р В»Р С‘РЎвЂЎР ВµРЎРѓРЎвЂљР Р†Р С• Р В·Р В°Р С—РЎС“РЎРѓР С”Р С•Р Р† Р С‘Р В·Р С–Р С•РЎвЂљР С•Р Р†Р В»Р ВµР Р…Р С‘РЎРЏ Р С—Р С• Р Р†РЎвЂ№Р В±РЎР‚Р В°Р Р…Р Р…Р С•Р в„– РЎРѓРЎвЂ¦Р ВµР СР Вµ.
}

// controlPanelItemSchema РЎвЂ¦РЎР‚Р В°Р Р…Р С‘РЎвЂљ Р Р…РЎС“Р В¶Р Р…РЎвЂ№Р Вµ РЎРѓР ВµРЎР‚Р Р†Р ВµРЎР‚РЎС“ Р С—Р С•Р В»РЎРЏ Р С•Р Т‘Р Р…Р С•Р в„– РЎРѓРЎвЂ№РЎР‚Р С•Р в„– Р В·Р В°Р С—Р С‘РЎРѓР С‘ РЎРѓРЎвЂ¦Р ВµР СРЎвЂ№.
type ControlPanelConstructorQueueCommand struct {
	ConstructorEquipmentGroupID int64  // Р вЂњРЎР‚РЎС“Р С—Р С—Р В° Р С”Р С•Р Р…РЎРѓРЎвЂљРЎР‚РЎС“Р С”РЎвЂљР С•РЎР‚Р С•Р Р†, Р С•РЎвЂЎР ВµРЎР‚Р ВµР Т‘РЎРЉ Р С”Р С•РЎвЂљР С•РЎР‚Р С•Р в„– Р СР ВµР Р…РЎРЏР ВµРЎвЂљРЎРѓРЎРЏ.
	JobID                       int64  // Р РЋРЎвЂљРЎР‚Р С•Р С”Р В° Р С•РЎРѓР Р…Р С•Р Р†Р Р…Р С•Р в„– Р С•РЎвЂЎР ВµРЎР‚Р ВµР Т‘Р С‘, Р Р†РЎвЂ№Р В±РЎР‚Р В°Р Р…Р Р…Р В°РЎРЏ Р С‘Р С–РЎР‚Р С•Р С”Р С•Р С.
	Command                     string // Р вЂќР ВµР в„–РЎРѓРЎвЂљР Р†Р С‘Р Вµ Р Р…Р В°Р Т‘ Р Р†РЎвЂ№Р В±РЎР‚Р В°Р Р…Р Р…Р С•Р в„– РЎРѓРЎвЂљРЎР‚Р С•Р С”Р С•Р в„– Р С‘ РЎРѓР В»Р ВµР Т‘РЎС“РЎР‹РЎвЂ°Р С‘Р СР С‘ РЎРѓРЎвЂљРЎР‚Р С•Р С”Р В°Р СР С‘.
}

type controlPanelItemSchema struct {
	ID               int64   `json:"ID"`               // Р Р€Р Р…Р С‘Р С”Р В°Р В»РЎРЉР Р…РЎвЂ№Р в„– РЎвЂЎР С‘РЎРѓР В»Р С•Р Р†Р С•Р в„– Р С‘Р Т‘Р ВµР Р…РЎвЂљР С‘РЎвЂћР С‘Р С”Р В°РЎвЂљР С•РЎР‚ РЎРѓРЎвЂ¦Р ВµР СРЎвЂ№.
	ItemModelID      int64   `json:"ItemModelID"`      // Р СљР С•Р Т‘Р ВµР В»РЎРЉ Р С—РЎР‚Р ВµР Т‘Р СР ВµРЎвЂљР В°, Р С—Р С•Р В»РЎС“РЎвЂЎР В°Р ВµР СР С•Р С–Р С• Р С—Р С• РЎРѓРЎвЂ¦Р ВµР СР Вµ.
	Count            float64 `json:"Count"`            // Р С™Р С•Р В»Р С‘РЎвЂЎР ВµРЎРѓРЎвЂљР Р†Р С• Р С—РЎР‚Р ВµР Т‘Р СР ВµРЎвЂљР С•Р Р†, Р С—Р С•Р В»РЎС“РЎвЂЎР В°Р ВµР СР С•Р Вµ Р В·Р В° Р С•Р Т‘Р Р…Р С• Р С‘Р В·Р С–Р С•РЎвЂљР С•Р Р†Р В»Р ВµР Р…Р С‘Р Вµ.
	ProductionEnergy float64 `json:"ProductionEnergy"` // Р вЂР В°Р В·Р С•Р Р†Р С•Р Вµ Р Р†РЎР‚Р ВµР СРЎРЏ Р С‘Р В·Р С–Р С•РЎвЂљР С•Р Р†Р В»Р ВµР Р…Р С‘РЎРЏ, Р С—Р С•Р С”Р В° Р Р…Р Вµ Р С‘РЎРѓР С—Р С•Р В»РЎРЉР В·РЎС“Р ВµР СР С•Р Вµ Р СР С–Р Р…Р С•Р Р†Р ВµР Р…Р Р…Р С•Р в„– Р С”Р С•Р СР В°Р Р…Р Т‘Р С•Р в„–.
}

// controlPanelItemSchemaComponent РЎвЂ¦РЎР‚Р В°Р Р…Р С‘РЎвЂљ Р Р…РЎС“Р В¶Р Р…РЎвЂ№Р Вµ РЎРѓР ВµРЎР‚Р Р†Р ВµРЎР‚РЎС“ Р С—Р С•Р В»РЎРЏ Р С•Р Т‘Р Р…Р С•Р С–Р С• Р С”Р С•Р СР С—Р С•Р Р…Р ВµР Р…РЎвЂљР В° РЎРѓРЎвЂ¦Р ВµР СРЎвЂ№.
type controlPanelItemSchemaComponent struct {
	ID                   int64   `json:"ID"`                   // Р Р€Р Р…Р С‘Р С”Р В°Р В»РЎРЉР Р…РЎвЂ№Р в„– РЎвЂЎР С‘РЎРѓР В»Р С•Р Р†Р С•Р в„– Р С‘Р Т‘Р ВµР Р…РЎвЂљР С‘РЎвЂћР С‘Р С”Р В°РЎвЂљР С•РЎР‚ Р С”Р С•Р СР С—Р С•Р Р…Р ВµР Р…РЎвЂљР В° РЎРѓРЎвЂ¦Р ВµР СРЎвЂ№.
	SchemaID             int64   `json:"SchemaID"`             // Р РЋРЎвЂ¦Р ВµР СР В°, Р С” Р С”Р С•РЎвЂљР С•РЎР‚Р С•Р в„– Р С•РЎвЂљР Р…Р С•РЎРѓР С‘РЎвЂљРЎРѓРЎРЏ Р С”Р С•Р СР С—Р С•Р Р…Р ВµР Р…РЎвЂљ.
	ComponentItemModelID int64   `json:"ComponentItemModelID"` // Р СљР С•Р Т‘Р ВµР В»РЎРЉ Р С—РЎР‚Р ВµР Т‘Р СР ВµРЎвЂљР В°, Р С”Р С•РЎвЂљР С•РЎР‚РЎС“РЎР‹ Р Р…РЎС“Р В¶Р Р…Р С• РЎРѓР С—Р С‘РЎРѓР В°РЎвЂљРЎРЉ Р С”Р В°Р С” Р С”Р С•Р СР С—Р С•Р Р…Р ВµР Р…РЎвЂљ.
	Count                float64 `json:"Count"`                // Р С™Р С•Р В»Р С‘РЎвЂЎР ВµРЎРѓРЎвЂљР Р†Р С• Р С”Р С•Р СР С—Р С•Р Р…Р ВµР Р…РЎвЂљР В°, РЎвЂљРЎР‚Р ВµР В±РЎС“Р ВµР СР С•Р Вµ Р Т‘Р В»РЎРЏ Р С•Р Т‘Р Р…Р С•Р С–Р С• Р С‘Р В·Р С–Р С•РЎвЂљР С•Р Р†Р В»Р ВµР Р…Р С‘РЎРЏ.
}

type controlPanelObjectBlueprint struct {
	ID                  int64   `json:"ID"`                  // Р Р€Р Р…Р С‘Р С”Р В°Р В»РЎРЉР Р…РЎвЂ№Р в„– РЎвЂЎР С‘РЎРѓР В»Р С•Р Р†Р С•Р в„– Р С‘Р Т‘Р ВµР Р…РЎвЂљР С‘РЎвЂћР С‘Р С”Р В°РЎвЂљР С•РЎР‚ РЎвЂЎР ВµРЎР‚РЎвЂљР ВµР В¶Р В°.
	CosmicObjectModelID int64   `json:"CosmicObjectModelID"` // Р СљР С•Р Т‘Р ВµР В»РЎРЉ Р С”Р С•РЎРѓР СР С‘РЎвЂЎР ВµРЎРѓР С”Р С•Р С–Р С• Р С•Р В±РЎР‰Р ВµР С”РЎвЂљР В°, Р С—Р С•Р В»РЎС“РЎвЂЎР В°Р ВµР СР С•Р С–Р С• Р С—Р С• РЎвЂЎР ВµРЎР‚РЎвЂљР ВµР В¶РЎС“.
	ProductionEnergy    float64 `json:"ProductionEnergy"`    // Р вЂР В°Р В·Р С•Р Р†Р С•Р Вµ Р Р†РЎР‚Р ВµР СРЎРЏ Р С‘Р В·Р С–Р С•РЎвЂљР С•Р Р†Р В»Р ВµР Р…Р С‘РЎРЏ Р С•Р Т‘Р Р…Р С•Р С–Р С• Р С•Р В±РЎР‰Р ВµР С”РЎвЂљР В°.
}

type controlPanelObjectBlueprintComponent struct {
	ID                   int64   `json:"ID"`                   // Р Р€Р Р…Р С‘Р С”Р В°Р В»РЎРЉР Р…РЎвЂ№Р в„– РЎвЂЎР С‘РЎРѓР В»Р С•Р Р†Р С•Р в„– Р С‘Р Т‘Р ВµР Р…РЎвЂљР С‘РЎвЂћР С‘Р С”Р В°РЎвЂљР С•РЎР‚ Р С”Р С•Р СР С—Р С•Р Р…Р ВµР Р…РЎвЂљР В° РЎвЂЎР ВµРЎР‚РЎвЂљР ВµР В¶Р В°.
	BlueprintID          int64   `json:"BlueprintID"`          // Р В§Р ВµРЎР‚РЎвЂљРЎвЂР В¶, Р С” Р С”Р С•РЎвЂљР С•РЎР‚Р С•Р СРЎС“ Р С•РЎвЂљР Р…Р С•РЎРѓР С‘РЎвЂљРЎРѓРЎРЏ Р С”Р С•Р СР С—Р С•Р Р…Р ВµР Р…РЎвЂљ.
	ComponentItemModelID int64   `json:"ComponentItemModelID"` // Р СљР С•Р Т‘Р ВµР В»РЎРЉ Р С—РЎР‚Р ВµР Т‘Р СР ВµРЎвЂљР В°, Р С”Р С•РЎвЂљР С•РЎР‚РЎС“РЎР‹ Р Р…РЎС“Р В¶Р Р…Р С• РЎРѓР С—Р С‘РЎРѓР В°РЎвЂљРЎРЉ Р С”Р В°Р С” Р С”Р С•Р СР С—Р С•Р Р…Р ВµР Р…РЎвЂљ.
	Count                float64 `json:"Count"`                // Р С™Р С•Р В»Р С‘РЎвЂЎР ВµРЎРѓРЎвЂљР Р†Р С• Р С”Р С•Р СР С—Р С•Р Р…Р ВµР Р…РЎвЂљР В°, РЎвЂљРЎР‚Р ВµР В±РЎС“Р ВµР СР С•Р Вµ Р Т‘Р В»РЎРЏ Р С•Р Т‘Р Р…Р С•Р С–Р С• Р С‘Р В·Р С–Р С•РЎвЂљР С•Р Р†Р В»Р ВµР Р…Р С‘РЎРЏ.
}

func New(seed int64, serverData Data) *World {
	created := &World{
		data:             serverData,
		accountObjectIDs: map[int64]int64{},
		inputs:           map[int64]game.ShipInput{},
		mutationAcks:     map[string]int64{},
		random:           rand.New(rand.NewSource(seed)),
	}
	created.ensureChatData()
	created.applyAssembliesToLoadedShips()
	return created
}

// Р СџРЎР‚Р С‘Р Р†РЎРЏР В·РЎвЂ№Р Р†Р В°Р ВµРЎвЂљ Р В°Р С”Р С”Р В°РЎС“Р Р…РЎвЂљ Р С” РЎвЂљР ВµР С”РЎС“РЎвЂ°Р ВµР СРЎС“ Р С•Р В±РЎР‰Р ВµР С”РЎвЂљРЎС“ Р ВµР С–Р С• Р С—Р ВµРЎР‚РЎРѓР С•Р Р…Р В°Р В¶Р В° Р С‘ РЎР‚Р В°Р В·РЎР‚Р ВµРЎв‚¬Р В°Р ВµРЎвЂљ Р С—Р С•Р В»РЎС“РЎвЂЎР В°РЎвЂљРЎРЉ Р Р†Р Р†Р С•Р Т‘.
func (world *World) ConnectAccount(accountID int64) (int64, bool) {
	world.mu.Lock()
	defer world.mu.Unlock()

	account, ok := world.data.Accounts.Get(accountID)
	if !ok || account.CurrentCharacterID <= 0 {
		return 0, false
	}

	character, ok := world.data.Characters.Get(account.CurrentCharacterID)
	if !ok || character.AccountID != account.ID || character.LocationCosmicObjectID <= 0 {
		return 0, false
	}

	cosmicObject, ok := world.data.CosmicObjects.Get(character.LocationCosmicObjectID)
	if !ok {
		return 0, false
	}
	cosmicObject.TargetRotation = cosmicObject.Rotation

	world.accountObjectIDs[accountID] = character.LocationCosmicObjectID
	world.inputs[accountID] = game.ShipInput{}
	return character.LocationCosmicObjectID, true
}

// Р РЋР С•Р В·Р Т‘Р В°Р ВµРЎвЂљ РЎС“РЎвЂЎР ВµРЎвЂљР Р…РЎС“РЎР‹ Р В·Р В°Р С—Р С‘РЎРѓРЎРЉ, Р С—Р ВµРЎР‚РЎРѓР С•Р Р…Р В°Р В¶Р В° Р С‘ РЎРѓРЎвЂљР В°РЎР‚РЎвЂљР С•Р Р†РЎвЂ№Р в„– Р С”Р С•РЎР‚Р В°Р В±Р В»РЎРЉ Р Т‘Р В»РЎРЏ Р С—Р ВµРЎР‚Р Р†Р С•Р С–Р С• Р Р†РЎвЂ¦Р С•Р Т‘Р В°.
func (world *World) CreateStarterAccount() (*data.Account, error) {
	world.mu.Lock()
	defer world.mu.Unlock()

	model, ok := world.data.CosmicObjectModels.GetByAcronym(defaultStarterShipAcronym)
	if !ok {
		return nil, fmt.Errorf("starter ship model %q not found", defaultStarterShipAcronym)
	}
	assembly, ok := world.firstPublicDeveloperAssembly(model.ID)
	if !ok {
		return nil, fmt.Errorf("public developer assembly for starter ship model %q not found", defaultStarterShipAcronym)
	}

	nextAccountID := world.data.Accounts.MaxID + 1
	account, err := world.data.Accounts.Add(&data.Account{
		Email:        fmt.Sprintf("auto%d@%s", nextAccountID, defaultAccountEmailDomain),
		Nickname:     fmt.Sprintf("Pilot%d", nextAccountID),
		PasswordHash: defaultAccountPassword,
	})
	if err != nil {
		return nil, err
	}

	character, err := world.data.Characters.Add(&data.Character{
		AccountID: account.ID,
	})
	if err != nil {
		world.data.Accounts.Delete(account.ID)
		return nil, err
	}

	cosmicObject := world.cosmicObjectFromModelAndAssembly(model, assembly)
	cosmicObject.OwnerCharacterID = character.ID
	cosmicObject.CreatorCharacterID = character.ID
	createdObject, err := world.data.CosmicObjects.Add(cosmicObject)
	if err != nil {
		world.data.Characters.Delete(character.ID)
		world.data.Accounts.Delete(account.ID)
		return nil, err
	}
	if err := world.replaceEquipmentFromAssembly(createdObject.ID, assembly); err != nil {
		world.data.CosmicObjects.Delete(createdObject.ID)
		world.data.Characters.Delete(character.ID)
		world.data.Accounts.Delete(account.ID)
		return nil, err
	}
	world.fillShipSupplies(createdObject)

	character.LocationCosmicObjectID = createdObject.ID
	if err := world.data.Accounts.SetCurrentCharacter(account.ID, character.ID); err != nil {
		world.data.CosmicObjects.Delete(createdObject.ID)
		world.data.Characters.Delete(character.ID)
		world.data.Accounts.Delete(account.ID)
		return nil, err
	}

	return account, nil
}

// Р Р€Р В±Р С‘РЎР‚Р В°Р ВµРЎвЂљ Р В°Р С”РЎвЂљР С‘Р Р†Р Р…РЎС“РЎР‹ Р С—РЎР‚Р С‘Р Р†РЎРЏР В·Р С”РЎС“ Р В°Р С”Р С”Р В°РЎС“Р Р…РЎвЂљР В° Р С‘ Р С—Р С•РЎРѓР В»Р ВµР Т‘Р Р…Р С‘Р в„– Р Р†Р Р†Р С•Р Т‘ Р С‘Р С–РЎР‚Р С•Р С”Р В°.
func (world *World) DisconnectAccount(accountID int64) {
	world.mu.Lock()
	defer world.mu.Unlock()

	delete(world.accountObjectIDs, accountID)
	delete(world.inputs, accountID)
}

// Р РЋР С•РЎвЂ¦РЎР‚Р В°Р Р…РЎРЏР ВµРЎвЂљ Р С—Р С•РЎРѓР В»Р ВµР Т‘Р Р…Р С‘Р в„– Р С—Р В°Р С”Р ВµРЎвЂљ РЎС“Р С—РЎР‚Р В°Р Р†Р В»Р ВµР Р…Р С‘РЎРЏ РЎвЂљР С•Р В»РЎРЉР С”Р С• Р Т‘Р В»РЎРЏ РЎС“Р В¶Р Вµ Р С—Р С•Р Т‘Р С”Р В»РЎР‹РЎвЂЎР ВµР Р…Р Р…Р С•Р С–Р С• Р В°Р С”Р С”Р В°РЎС“Р Р…РЎвЂљР В°.
func (world *World) SetInput(accountID int64, input game.ShipInput) {
	world.mu.Lock()
	defer world.mu.Unlock()

	objectID, ok := world.accountObjectIDs[accountID]
	if !ok {
		return
	}

	if input.ToggleAnchor {
		if cosmicObject, ok := world.data.CosmicObjects.Get(objectID); ok {
			if cosmicObject.ClusterMainCosmicObjectID == 0 && (cosmicObject.Anchored || cosmicObjectIsFullyStopped(*cosmicObject)) {
				cosmicObject.Anchored = !cosmicObject.Anchored
			}
		}
		input.ToggleAnchor = false
	}
	world.inputs[accountID] = input
}

// SendDockingRequest Р В·Р В°Р С—РЎС“РЎРѓР С”Р В°Р ВµРЎвЂљ Р В·Р В°Р С—РЎР‚Р С•РЎРѓ РЎРѓРЎвЂљРЎвЂ№Р С”Р С•Р Р†Р С”Р С‘ РЎвЂљР ВµР С”РЎС“РЎвЂ°Р ВµР С–Р С• Р С”Р С•РЎР‚Р В°Р В±Р В»РЎРЏ Р С” Р С•Р В±РЎР‰Р ВµР С”РЎвЂљРЎС“ Р С—Р ВµРЎР‚Р ВµР Т‘ Р Р…Р С•РЎРѓР С•Р С.
func (world *World) SendDockingRequest(accountID int64) error {
	world.mu.Lock()
	defer world.mu.Unlock()

	sender, err := world.controlledCosmicObjectLocked(accountID)
	if err != nil {
		return err
	}
	if err := world.validateDockingSenderLocked(sender); err != nil {
		return err
	}
	receiver, err := world.findDockingReceiverLocked(sender)
	if err != nil {
		return err
	}
	if err := world.validateDockingReceiverLocked(receiver); err != nil {
		return err
	}
	if world.dockingObjectIsBusyLocked(sender.ID) || world.dockingObjectIsBusyLocked(receiver.ID) {
		return errors.New("object already participates in docking")
	}
	if receiver.OwnerCharacterID == sender.OwnerCharacterID {
		world.startDockingProcessLocked(sender.ID, receiver.ID)
		return nil
	}
	if !world.dockingReceiverHasDecisionMakerLocked(receiver.ID) {
		world.addDockingNotificationLocked([]int64{sender.ID}, "В Получателе нет персонажа для принятия решения")
		return nil
	}
	world.dockingRequests = append(world.dockingRequests, dockingRequest{
		SenderCosmicObjectID:   sender.ID,
		ReceiverCosmicObjectID: receiver.ID,
		RemainingSeconds:       dockingDurationSeconds,
	})
	world.addDockingWindowEventsLocked("dockingRequestStarted", sender.ID, receiver.ID, dockingDurationSeconds)
	return nil
}

// ApproveDockingRequest Р С•Р Т‘Р С•Р В±РЎР‚РЎРЏР ВµРЎвЂљ Р ВµР Т‘Р С‘Р Р…РЎРѓРЎвЂљР Р†Р ВµР Р…Р Р…РЎвЂ№Р в„– Р Р†РЎвЂ¦Р С•Р Т‘РЎРЏРЎвЂ°Р С‘Р в„– Р В·Р В°Р С—РЎР‚Р С•РЎРѓ РЎвЂљР ВµР С”РЎС“РЎвЂ°Р ВµР С–Р С• Р С•Р В±РЎР‰Р ВµР С”РЎвЂљР В°.
func (world *World) ApproveDockingRequest(accountID int64) error {
	world.mu.Lock()
	defer world.mu.Unlock()

	receiver, err := world.controlledCosmicObjectLocked(accountID)
	if err != nil {
		return err
	}
	requestIndex := world.dockingRequestIndexByReceiverLocked(receiver.ID)
	if requestIndex < 0 {
		return errors.New("docking request not found")
	}
	request := world.dockingRequests[requestIndex]
	sender, ok := world.data.CosmicObjects.Get(request.SenderCosmicObjectID)
	if !ok {
		world.removeDockingRequestLocked(requestIndex)
		return errors.New("sender object not found")
	}
	if err := world.validateDockingSenderLocked(sender); err != nil {
		world.removeDockingRequestLocked(requestIndex)
		world.closeDockingRequestWindowLocked(request)
		world.addDockingNotificationLocked([]int64{request.SenderCosmicObjectID, request.ReceiverCosmicObjectID}, "Условия стыковки больше не выполняются")
		return err
	}
	if err := world.validateDockingReceiverLocked(receiver); err != nil {
		world.removeDockingRequestLocked(requestIndex)
		world.closeDockingRequestWindowLocked(request)
		world.addDockingNotificationLocked([]int64{request.SenderCosmicObjectID, request.ReceiverCosmicObjectID}, "Условия стыковки больше не выполняются")
		return err
	}
	world.removeDockingRequestLocked(requestIndex)
	world.startDockingProcessLocked(sender.ID, receiver.ID)
	return nil
}

// RejectDockingRequest Р С•РЎвЂљР С”Р В»Р С•Р Р…РЎРЏР ВµРЎвЂљ Р ВµР Т‘Р С‘Р Р…РЎРѓРЎвЂљР Р†Р ВµР Р…Р Р…РЎвЂ№Р в„– Р Р†РЎвЂ¦Р С•Р Т‘РЎРЏРЎвЂ°Р С‘Р в„– Р В·Р В°Р С—РЎР‚Р С•РЎРѓ РЎвЂљР ВµР С”РЎС“РЎвЂ°Р ВµР С–Р С• Р С•Р В±РЎР‰Р ВµР С”РЎвЂљР В°.
func (world *World) RejectDockingRequest(accountID int64) error {
	world.mu.Lock()
	defer world.mu.Unlock()

	receiver, err := world.controlledCosmicObjectLocked(accountID)
	if err != nil {
		return err
	}
	requestIndex := world.dockingRequestIndexByReceiverLocked(receiver.ID)
	if requestIndex < 0 {
		return errors.New("docking request not found")
	}
	request := world.dockingRequests[requestIndex]
	world.removeDockingRequestLocked(requestIndex)
	world.closeDockingRequestWindowLocked(request)
	world.addDockingNotificationLocked([]int64{request.SenderCosmicObjectID, request.ReceiverCosmicObjectID}, "Отказ на запрос стыковки")
	return nil
}

// UndockControlledObject Р Р†РЎвЂ№Р Р†Р С•Р Т‘Р С‘РЎвЂљ РЎвЂљР ВµР С”РЎС“РЎвЂ°Р С‘Р в„– Р С•Р В±РЎР‰Р ВµР С”РЎвЂљ Р С‘Р В· Р С”Р В»Р В°РЎРѓРЎвЂљР ВµРЎР‚Р В° Р С‘Р В»Р С‘ РЎР‚Р В°РЎРѓР С—РЎС“РЎРѓР С”Р В°Р ВµРЎвЂљ Р С”Р В»Р В°РЎРѓРЎвЂљР ВµРЎР‚ РЎвЂ Р ВµР В»Р С‘Р С”Р С•Р С.
func (world *World) UndockControlledObject(accountID int64) error {
	world.mu.Lock()
	defer world.mu.Unlock()

	cosmicObject, err := world.controlledCosmicObjectLocked(accountID)
	if err != nil {
		return err
	}
	if world.exchangeClusterIsBusyLocked(cosmicObject.ID) {
		return errors.New("object participates in exchange")
	}
	mainID := cosmicObject.ClusterMainCosmicObjectID
	if mainID <= 0 {
		return errors.New("object is not docked")
	}
	notificationObjectIDs := world.clusterObjectIDsLocked(mainID)
	if mainID == cosmicObject.ID {
		world.disbandClusterLocked(mainID)
		world.addDockingNotificationLocked(notificationObjectIDs, "Объект отстыкован")
		return nil
	}
	cosmicObject.ClusterMainCosmicObjectID = 0
	cosmicObject.Anchored = false
	if len(world.clusterObjectIDsLocked(mainID)) <= 1 {
		world.disbandClusterLocked(mainID)
	}
	world.addDockingNotificationLocked(notificationObjectIDs, "Объект отстыкован")
	return nil
}

// BeginCharacterTransfer Р Р…Р В°РЎвЂЎР С‘Р Р…Р В°Р ВµРЎвЂљ Р С—Р ВµРЎР‚Р ВµРЎРѓР В°Р Т‘Р С”РЎС“ РЎвЂљР ВµР С”РЎС“РЎвЂ°Р ВµР С–Р С• Р С—Р ВµРЎР‚РЎРѓР С•Р Р…Р В°Р В¶Р В° Р Р† Р С•Р В±РЎР‰Р ВµР С”РЎвЂљ РЎвЂљР С•Р С–Р С• Р В¶Р Вµ Р С”Р В»Р В°РЎРѓРЎвЂљР ВµРЎР‚Р В°.
func (world *World) BeginCharacterTransfer(accountID int64) error {
	world.mu.Lock()
	defer world.mu.Unlock()

	character, err := world.currentCharacterLocked(accountID)
	if err != nil {
		return err
	}
	sender, err := world.controlledCosmicObjectLocked(accountID)
	if err != nil {
		return err
	}
	if sender.ClusterMainCosmicObjectID <= 0 {
		world.addDockingNotificationLocked([]int64{sender.ID}, "Объект не пристыкован")
		return nil
	}
	targetID, ok := world.autoLandingTargetIDLocked(sender)
	if !ok {
		world.addLandingTargetSelectionLocked(sender.ID)
		return nil
	}
	return world.requestCharacterLandingLocked(character.ID, sender.ID, targetID)
}

// RequestCharacterLanding Р С•РЎвЂљР С—РЎР‚Р В°Р Р†Р В»РЎРЏР ВµРЎвЂљ Р В·Р В°Р С—РЎР‚Р С•РЎРѓ Р Р…Р В° Р С—Р С•РЎРѓР В°Р Т‘Р С”РЎС“ Р Р† Р Р†РЎвЂ№Р В±РЎР‚Р В°Р Р…Р Р…РЎвЂ№Р в„– Р С•Р В±РЎР‰Р ВµР С”РЎвЂљ Р Р…Р В°Р В·Р Р…Р В°РЎвЂЎР ВµР Р…Р С‘РЎРЏ.
func (world *World) RequestCharacterLanding(accountID int64, targetID int64) error {
	world.mu.Lock()
	defer world.mu.Unlock()

	character, err := world.currentCharacterLocked(accountID)
	if err != nil {
		return err
	}
	sender, err := world.controlledCosmicObjectLocked(accountID)
	if err != nil {
		return err
	}
	if sender.ClusterMainCosmicObjectID <= 0 {
		world.addDockingNotificationLocked([]int64{sender.ID}, "Объект не пристыкован")
		return nil
	}
	return world.requestCharacterLandingLocked(character.ID, sender.ID, targetID)
}

// ApproveCharacterLanding Р С•Р Т‘Р С•Р В±РЎР‚РЎРЏР ВµРЎвЂљ Р ВµР Т‘Р С‘Р Р…РЎРѓРЎвЂљР Р†Р ВµР Р…Р Р…РЎвЂ№Р в„– Р Р†РЎвЂ¦Р С•Р Т‘РЎРЏРЎвЂ°Р С‘Р в„– Р В·Р В°Р С—РЎР‚Р С•РЎРѓ Р С—Р С•РЎРѓР В°Р Т‘Р С”Р С‘ Р Т‘Р В»РЎРЏ РЎвЂљР ВµР С”РЎС“РЎвЂ°Р ВµР С–Р С• Р С•Р В±РЎР‰Р ВµР С”РЎвЂљР В°.
func (world *World) ApproveCharacterLanding(accountID int64) error {
	world.mu.Lock()
	defer world.mu.Unlock()

	receiver, err := world.controlledCosmicObjectLocked(accountID)
	if err != nil {
		return err
	}
	requestIndex := world.landingRequestIndexByReceiverLocked(receiver.ID)
	if requestIndex < 0 {
		return errors.New("landing request not found")
	}
	request := world.landingRequests[requestIndex]
	world.removeLandingRequestLocked(requestIndex)
	world.moveCharacterToObjectLocked(request.CharacterID, request.ReceiverCosmicObjectID)
	world.addLandingWindowEventsLocked("landingFinished", request.SenderCosmicObjectID, request.ReceiverCosmicObjectID)
	return nil
}

// RejectCharacterLanding Р С•РЎвЂљР С”Р В»Р С•Р Р…РЎРЏР ВµРЎвЂљ Р ВµР Т‘Р С‘Р Р…РЎРѓРЎвЂљР Р†Р ВµР Р…Р Р…РЎвЂ№Р в„– Р Р†РЎвЂ¦Р С•Р Т‘РЎРЏРЎвЂ°Р С‘Р в„– Р В·Р В°Р С—РЎР‚Р С•РЎРѓ Р С—Р С•РЎРѓР В°Р Т‘Р С”Р С‘ Р Т‘Р В»РЎРЏ РЎвЂљР ВµР С”РЎС“РЎвЂ°Р ВµР С–Р С• Р С•Р В±РЎР‰Р ВµР С”РЎвЂљР В°.
func (world *World) RejectCharacterLanding(accountID int64) error {
	world.mu.Lock()
	defer world.mu.Unlock()

	receiver, err := world.controlledCosmicObjectLocked(accountID)
	if err != nil {
		return err
	}
	requestIndex := world.landingRequestIndexByReceiverLocked(receiver.ID)
	if requestIndex < 0 {
		return errors.New("landing request not found")
	}
	request := world.landingRequests[requestIndex]
	world.removeLandingRequestLocked(requestIndex)
	world.addLandingWindowEventsLocked("landingFinished", request.SenderCosmicObjectID, request.ReceiverCosmicObjectID)
	return nil
}

// ClaimFocusedObjectOwnerForTesting Р Т‘Р ВµР В»Р В°Р ВµРЎвЂљ Р С•Р В±РЎР‰Р ВµР С”РЎвЂљ Р С—Р ВµРЎР‚Р ВµР Т‘ Р Р…Р С•РЎРѓР С•Р С РЎвЂљР ВµР С”РЎС“РЎвЂ°Р ВµР С–Р С• Р С”Р С•РЎР‚Р В°Р В±Р В»РЎРЏ РЎРѓР С•Р В±РЎРѓРЎвЂљР Р†Р ВµР Р…Р Р…Р С•РЎРѓРЎвЂљРЎРЉРЎР‹ РЎвЂљР ВµР С”РЎС“РЎвЂ°Р ВµР С–Р С• Р С—Р ВµРЎР‚РЎРѓР С•Р Р…Р В°Р В¶Р В°.
func (world *World) ClaimFocusedObjectOwnerForTesting(accountID int64) error {
	world.mu.Lock()
	defer world.mu.Unlock()

	character, err := world.currentCharacterLocked(accountID)
	if err != nil {
		return err
	}
	cosmicObject, err := world.controlledCosmicObjectLocked(accountID)
	if err != nil {
		return err
	}
	target, err := world.findDockingReceiverLocked(cosmicObject)
	if err != nil {
		return err
	}
	target.OwnerCharacterID = character.ID
	target.OwnerNpcClanID = 0
	return world.data.CosmicObjects.RebuildIndexes()
}

// Р СљР ВµР Р…РЎРЏР ВµРЎвЂљ РЎС“Р С—РЎР‚Р В°Р Р†Р В»РЎРЏР ВµР СРЎвЂ№Р в„– Р С•Р В±РЎР‰Р ВµР С”РЎвЂљ Р Р…Р В° Р Т‘РЎР‚РЎС“Р С–РЎС“РЎР‹ РЎРѓР В»РЎС“РЎвЂЎР В°Р в„–Р Р…РЎС“РЎР‹ Р СР С•Р Т‘Р ВµР В»РЎРЉ Р С”Р С•РЎР‚Р В°Р В±Р В»РЎРЏ Р С‘Р В· РЎРѓР С—РЎР‚Р В°Р Р†Р С•РЎвЂЎР Р…Р С‘Р С”Р В°.
func (world *World) ChangeControlledShipToRandomModel(accountID int64) bool {
	world.mu.Lock()
	defer world.mu.Unlock()

	objectID, ok := world.accountObjectIDs[accountID]
	if !ok {
		return false
	}

	cosmicObject, ok := world.data.CosmicObjects.Get(objectID)
	if !ok {
		return false
	}
	character, err := world.currentCharacterLocked(accountID)
	if err != nil {
		return false
	}

	shipType, ok := world.data.CosmicObjectTypes.GetByAcronym("Ship")
	if !ok {
		return false
	}

	candidateIDs := make([]int64, 0)
	for modelID, model := range world.data.CosmicObjectModels.Items {
		if model.CosmicObjectTypeID == shipType.ID && modelID != cosmicObject.CosmicObjectModelID {
			if _, ok := world.firstPublicDeveloperAssembly(modelID); ok {
				candidateIDs = append(candidateIDs, modelID)
			}
		}
	}
	if len(candidateIDs) == 0 {
		return false
	}
	sort.Slice(candidateIDs, func(left int, right int) bool {
		return candidateIDs[left] < candidateIDs[right]
	})

	model, ok := world.data.CosmicObjectModels.Get(candidateIDs[world.random.Intn(len(candidateIDs))])
	if !ok {
		return false
	}
	assembly, ok := world.firstPublicDeveloperAssembly(model.ID)
	if !ok {
		return false
	}

	world.applyModelAndAssembly(cosmicObject, model, assembly)
	cosmicObject.OwnerCharacterID = character.ID
	cosmicObject.OwnerNpcClanID = 0
	cosmicObject.TargetRotation = cosmicObject.Rotation
	if err := world.replaceEquipmentFromAssembly(cosmicObject.ID, assembly); err != nil {
		return false
	}
	cosmicObject.Armor = cosmicObject.MaxArmor
	world.fillShipSupplies(cosmicObject)

	return world.data.CosmicObjects.RebuildIndexes() == nil
}

// controlledCosmicObjectLocked Р Р†Р С•Р В·Р Р†РЎР‚Р В°РЎвЂ°Р В°Р ВµРЎвЂљ Р С•Р В±РЎР‰Р ВµР С”РЎвЂљ РЎвЂљР ВµР С”РЎС“РЎвЂ°Р ВµР С–Р С• Р С—Р ВµРЎР‚РЎРѓР С•Р Р…Р В°Р В¶Р В° Р С—Р С•Р Т‘ РЎС“Р В¶Р Вµ Р Р†Р В·РЎРЏРЎвЂљРЎвЂ№Р С mutex.
func (world *World) controlledCosmicObjectLocked(accountID int64) (*data.CosmicObject, error) {
	objectID, ok := world.accountObjectIDs[accountID]
	if !ok {
		return nil, errors.New("account is not connected")
	}
	cosmicObject, ok := world.data.CosmicObjects.Get(objectID)
	if !ok {
		return nil, errors.New("controlled object not found")
	}
	return cosmicObject, nil
}

// Р вЂ™Р С•Р В·Р Р†РЎР‚Р В°РЎвЂ°Р В°Р ВµРЎвЂљ РЎвЂљР ВµР С”РЎС“РЎвЂ°Р ВµР С–Р С• Р С—Р ВµРЎР‚РЎРѓР С•Р Р…Р В°Р В¶Р В° Р С—Р С•Р Т‘Р С”Р В»РЎР‹РЎвЂЎР ВµР Р…Р Р…Р С•Р С–Р С• Р В°Р С”Р С”Р В°РЎС“Р Р…РЎвЂљР В° Р С—Р С•Р Т‘ РЎС“Р В¶Р Вµ Р Р†Р В·РЎРЏРЎвЂљРЎвЂ№Р С mutex.
func (world *World) currentCharacterLocked(accountID int64) (*data.Character, error) {
	account, ok := world.data.Accounts.Get(accountID)
	if !ok || account.CurrentCharacterID <= 0 {
		return nil, errors.New("account is not connected")
	}
	character, ok := world.data.Characters.Get(account.CurrentCharacterID)
	if !ok || character.AccountID != account.ID {
		return nil, errors.New("current character not found")
	}
	return character, nil
}

// validateDockingSenderLocked Р С—РЎР‚Р С•Р Р†Р ВµРЎР‚РЎРЏР ВµРЎвЂљ РЎС“РЎРѓР В»Р С•Р Р†Р С‘РЎРЏ Р Т‘Р В»РЎРЏ Р С•Р В±РЎР‰Р ВµР С”РЎвЂљР В°, Р С•РЎвЂљР С—РЎР‚Р В°Р Р†Р В»РЎРЏРЎР‹РЎвЂ°Р ВµР С–Р С• Р В·Р В°Р С—РЎР‚Р С•РЎРѓ.
func (world *World) validateDockingSenderLocked(sender *data.CosmicObject) error {
	if sender == nil {
		return errors.New("sender object not found")
	}
	if !world.cosmicObjectHasTypeLocked(sender, "Ship") {
		return errors.New("sender must be a ship")
	}
	if sender.ClusterMainCosmicObjectID > 0 {
		return errors.New("sender is already docked")
	}
	if !cosmicObjectIsFullyStopped(*sender) {
		return errors.New("sender is not stopped")
	}
	return nil
}

// validateDockingReceiverLocked Р С—РЎР‚Р С•Р Р†Р ВµРЎР‚РЎРЏР ВµРЎвЂљ РЎС“РЎРѓР В»Р С•Р Р†Р С‘РЎРЏ Р Т‘Р В»РЎРЏ Р С•Р В±РЎР‰Р ВµР С”РЎвЂљР В°, Р С—РЎР‚Р С‘Р Р…Р С‘Р СР В°РЎР‹РЎвЂ°Р ВµР С–Р С• Р В·Р В°Р С—РЎР‚Р С•РЎРѓ.
func (world *World) validateDockingReceiverLocked(receiver *data.CosmicObject) error {
	if receiver == nil {
		return errors.New("receiver object not found")
	}
	if !world.cosmicObjectHasTypeLocked(receiver, "Ship") && !world.cosmicObjectHasTypeLocked(receiver, "Station") {
		return errors.New("receiver must be a ship or station")
	}
	if receiver.ClusterMainCosmicObjectID > 0 && receiver.ClusterMainCosmicObjectID != receiver.ID {
		return errors.New("receiver is secondary object")
	}
	return nil
}

// cosmicObjectHasTypeLocked Р С—РЎР‚Р С•Р Р†Р ВµРЎР‚РЎРЏР ВµРЎвЂљ РЎвЂљР С‘Р С— Р С•Р В±РЎР‰Р ВµР С”РЎвЂљР В° РЎвЂЎР ВµРЎР‚Р ВµР В· Р ВµР С–Р С• Р СР С•Р Т‘Р ВµР В»РЎРЉ.
func (world *World) cosmicObjectHasTypeLocked(cosmicObject *data.CosmicObject, acronym string) bool {
	if cosmicObject == nil {
		return false
	}
	model, ok := world.data.CosmicObjectModels.Get(cosmicObject.CosmicObjectModelID)
	if !ok {
		return false
	}
	objectType, ok := world.data.CosmicObjectTypes.Get(model.CosmicObjectTypeID)
	return ok && objectType.Acronym == acronym
}

// findDockingReceiverLocked Р С‘РЎвЂ°Р ВµРЎвЂљ Р В±Р В»Р С‘Р В¶Р В°Р в„–РЎв‚¬Р С‘Р в„– Р С•Р В±РЎР‰Р ВµР С”РЎвЂљ, Р С—Р ВµРЎР‚Р ВµРЎРѓР ВµРЎвЂЎР ВµР Р…Р Р…РЎвЂ№Р в„– Р В»РЎС“РЎвЂЎР С•Р С Р С•РЎвЂљ Р Р…Р С•РЎРѓР В° Р С•РЎвЂљР С—РЎР‚Р В°Р Р†Р С‘РЎвЂљР ВµР В»РЎРЏ.
func (world *World) findDockingReceiverLocked(sender *data.CosmicObject) (*data.CosmicObject, error) {
	senderModel, ok := world.data.CosmicObjectModels.Get(sender.CosmicObjectModelID)
	if !ok {
		return nil, errors.New("sender model not found")
	}
	startX := sender.X + math.Sin(sender.Rotation)*senderModel.BodyLength/2
	startY := sender.Y + math.Cos(sender.Rotation)*senderModel.BodyLength/2
	endX := startX + math.Sin(sender.Rotation)*dockingProbeDistance
	endY := startY + math.Cos(sender.Rotation)*dockingProbeDistance

	var selected *data.CosmicObject
	selectedDistance := math.Inf(1)
	for _, candidate := range world.data.CosmicObjects.Items {
		if candidate == nil || candidate.ID == sender.ID {
			continue
		}
		candidateModel, ok := world.data.CosmicObjectModels.Get(candidate.CosmicObjectModelID)
		if !ok {
			continue
		}
		distance, ok := raySegmentPolygonDistance(startX, startY, endX, endY, *candidate, *candidateModel)
		if !ok || distance >= selectedDistance {
			continue
		}
		selected = candidate
		selectedDistance = distance
	}
	if selected == nil {
		return nil, errors.New("docking receiver not found")
	}
	return selected, nil
}

// startDockingProcessLocked РЎРѓРЎвЂљР В°Р Р†Р С‘РЎвЂљ Р С•Р В±РЎР‰Р ВµР С”РЎвЂљРЎвЂ№ Р Р…Р В° РЎРЏР С”Р С•РЎР‚РЎРЉ Р С‘ Р В·Р В°Р С—РЎС“РЎРѓР С”Р В°Р ВµРЎвЂљ РЎвЂљР В°Р в„–Р СР ВµРЎР‚ РЎРѓРЎвЂљРЎвЂ№Р С”Р С•Р Р†Р С”Р С‘.
func (world *World) startDockingProcessLocked(senderID int64, receiverID int64) {
	if sender, ok := world.data.CosmicObjects.Get(senderID); ok {
		sender.Anchored = true
	}
	if receiver, ok := world.data.CosmicObjects.Get(receiverID); ok {
		receiver.Anchored = true
	}
	world.dockingProcesses = append(world.dockingProcesses, dockingProcess{
		SenderCosmicObjectID:   senderID,
		ReceiverCosmicObjectID: receiverID,
		RemainingSeconds:       dockingDurationSeconds,
	})
	world.addDockingWindowEventsLocked("dockingProcessStarted", senderID, receiverID, dockingDurationSeconds)
}

// dockingObjectIsBusyLocked Р С—РЎР‚Р С•Р Р†Р ВµРЎР‚РЎРЏР ВµРЎвЂљ РЎС“РЎвЂЎР В°РЎРѓРЎвЂљР С‘Р Вµ Р С•Р В±РЎР‰Р ВµР С”РЎвЂљР В° Р Р† Р В°Р Р†РЎвЂљР С•Р СР В°РЎвЂљР С‘РЎвЂЎР ВµРЎРѓР С”Р С•Р в„– РЎРѓРЎвЂљРЎвЂ№Р С”Р С•Р Р†Р С”Р Вµ.
func (world *World) dockingObjectIsBusyLocked(objectID int64) bool {
	for _, request := range world.dockingRequests {
		if request.SenderCosmicObjectID == objectID || request.ReceiverCosmicObjectID == objectID {
			return true
		}
	}
	for _, process := range world.dockingProcesses {
		if process.SenderCosmicObjectID == objectID || process.ReceiverCosmicObjectID == objectID {
			return true
		}
	}
	return false
}

// dockingRequestIndexByReceiverLocked Р С‘РЎвЂ°Р ВµРЎвЂљ Р В°Р С”РЎвЂљР С‘Р Р†Р Р…РЎвЂ№Р в„– Р Р†РЎвЂ¦Р С•Р Т‘РЎРЏРЎвЂ°Р С‘Р в„– Р В·Р В°Р С—РЎР‚Р С•РЎРѓ Р Т‘Р В»РЎРЏ Р С•Р В±РЎР‰Р ВµР С”РЎвЂљР В°.
func (world *World) dockingRequestIndexByReceiverLocked(receiverID int64) int {
	for index, request := range world.dockingRequests {
		if request.ReceiverCosmicObjectID == receiverID {
			return index
		}
	}
	return -1
}

// removeDockingRequestLocked РЎС“Р Т‘Р В°Р В»РЎРЏР ВµРЎвЂљ Р В·Р В°Р С—РЎР‚Р С•РЎРѓ Р В±Р ВµР В· РЎРѓР С•РЎвЂ¦РЎР‚Р В°Р Р…Р ВµР Р…Р С‘РЎРЏ Р С—Р С•РЎР‚РЎРЏР Т‘Р С”Р В°.
func (world *World) removeDockingRequestLocked(index int) {
	world.dockingRequests = append(world.dockingRequests[:index], world.dockingRequests[index+1:]...)
}

// closeDockingRequestWindowLocked Р Т‘Р С•Р В±Р В°Р Р†Р В»РЎРЏР ВµРЎвЂљ РЎРѓР С•Р В±РЎвЂ№РЎвЂљР С‘Р Вµ Р В·Р В°Р С”РЎР‚РЎвЂ№РЎвЂљР С‘РЎРЏ Р С•Р С”Р Р…Р В° Р С•Р В¶Р С‘Р Т‘Р В°Р Р…Р С‘РЎРЏ Р С•РЎвЂљР Р†Р ВµРЎвЂљР В°.
func (world *World) closeDockingRequestWindowLocked(request dockingRequest) {
	world.addDockingWindowEventsLocked("dockingFinished", request.SenderCosmicObjectID, request.ReceiverCosmicObjectID, 0)
}

// autoLandingTargetIDLocked Р Р†РЎвЂ№Р В±Р С‘РЎР‚Р В°Р ВµРЎвЂљ Р ВµР Т‘Р С‘Р Р…РЎРѓРЎвЂљР Р†Р ВµР Р…Р Р…Р С•Р Вµ Р Р…Р В°Р В·Р Р…Р В°РЎвЂЎР ВµР Р…Р С‘Р Вµ Р С—Р ВµРЎР‚Р ВµРЎРѓР В°Р Т‘Р С”Р С‘ Р С—Р С• РЎРѓР С•РЎРѓРЎвЂљР В°Р Р†РЎС“ Р С”Р В»Р В°РЎРѓРЎвЂљР ВµРЎР‚Р В°.
func (world *World) autoLandingTargetIDLocked(sender *data.CosmicObject) (int64, bool) {
	if sender == nil || sender.ClusterMainCosmicObjectID <= 0 {
		return 0, false
	}
	mainID := sender.ClusterMainCosmicObjectID
	if sender.ID != mainID {
		return mainID, true
	}
	secondaryIDs := world.secondaryClusterObjectIDsLocked(mainID)
	if len(secondaryIDs) != 1 {
		return 0, false
	}
	return secondaryIDs[0], true
}

// requestCharacterLandingLocked Р С—РЎР‚Р С‘Р СР ВµР Р…РЎРЏР ВµРЎвЂљ Р С—РЎР‚Р В°Р Р†Р С‘Р В»Р В° Р С—Р С•РЎРѓР В°Р Т‘Р С”Р С‘ Р Р† Р Р†РЎвЂ№Р В±РЎР‚Р В°Р Р…Р Р…РЎвЂ№Р в„– Р С•Р В±РЎР‰Р ВµР С”РЎвЂљ Р Р…Р В°Р В·Р Р…Р В°РЎвЂЎР ВµР Р…Р С‘РЎРЏ.
func (world *World) requestCharacterLandingLocked(characterID int64, senderID int64, receiverID int64) error {
	character, ok := world.data.Characters.Get(characterID)
	if !ok {
		return errors.New("character not found")
	}
	sender, ok := world.data.CosmicObjects.Get(senderID)
	if !ok {
		return errors.New("sender object not found")
	}
	receiver, ok := world.data.CosmicObjects.Get(receiverID)
	if !ok {
		return errors.New("receiver object not found")
	}
	if !world.objectsInSameClusterLocked(sender, receiver) || sender.ID == receiver.ID {
		return errors.New("landing target is not in the same cluster")
	}
	if receiver.OwnerCharacterID == character.ID {
		world.moveCharacterToObjectLocked(character.ID, receiver.ID)
		return nil
	}
	if !world.cosmicObjectHasPassengerSeatLocked(receiver.ID) {
		world.addDockingNotificationLocked([]int64{sender.ID}, "В объекте назначения не установлено пассажирское кресло")
		return nil
	}
	if world.landingRequestIndexByReceiverLocked(receiver.ID) >= 0 {
		return errors.New("landing request already exists")
	}
	world.landingRequests = append(world.landingRequests, landingRequest{
		CharacterID:            character.ID,
		SenderCosmicObjectID:   sender.ID,
		ReceiverCosmicObjectID: receiver.ID,
	})
	world.addLandingWindowEventsLocked("landingRequestStarted", sender.ID, receiver.ID)
	return nil
}

// moveCharacterToObjectLocked Р С—Р ВµРЎР‚Р ВµР Р…Р С•РЎРѓР С‘РЎвЂљ Р С—Р ВµРЎР‚РЎРѓР С•Р Р…Р В°Р В¶Р В° Р С‘ Р С•Р В±Р Р…Р С•Р Р†Р В»РЎРЏР ВµРЎвЂљ РЎС“Р С—РЎР‚Р В°Р Р†Р В»РЎРЏР ВµР СРЎвЂ№Р в„– Р С•Р В±РЎР‰Р ВµР С”РЎвЂљ Р С—Р С•Р Т‘Р С”Р В»РЎР‹РЎвЂЎР ВµР Р…Р Р…Р С•Р С–Р С• Р В°Р С”Р С”Р В°РЎС“Р Р…РЎвЂљР В°.
func (world *World) moveCharacterToObjectLocked(characterID int64, objectID int64) {
	character, ok := world.data.Characters.Get(characterID)
	if !ok {
		return
	}
	character.LocationCosmicObjectID = objectID
	world.accountObjectIDs[character.AccountID] = objectID
}

// objectsInSameClusterLocked Р С—РЎР‚Р С•Р Р†Р ВµРЎР‚РЎРЏР ВµРЎвЂљ, РЎвЂЎРЎвЂљР С• Р С•Р В±Р В° Р С•Р В±РЎР‰Р ВµР С”РЎвЂљР В° Р Р†РЎвЂ¦Р С•Р Т‘РЎРЏРЎвЂљ Р Р† Р С•Р Т‘Р С‘Р Р… Р С”Р В»Р В°РЎРѓРЎвЂљР ВµРЎР‚.
func (world *World) objectsInSameClusterLocked(left *data.CosmicObject, right *data.CosmicObject) bool {
	return left != nil && right != nil && left.ClusterMainCosmicObjectID > 0 && left.ClusterMainCosmicObjectID == right.ClusterMainCosmicObjectID
}

// secondaryClusterObjectIDsLocked Р Р†Р С•Р В·Р Р†РЎР‚Р В°РЎвЂ°Р В°Р ВµРЎвЂљ Р Р†РЎвЂљР С•РЎР‚Р С•РЎРѓРЎвЂљР ВµР С—Р ВµР Р…Р Р…РЎвЂ№Р Вµ Р С•Р В±РЎР‰Р ВµР С”РЎвЂљРЎвЂ№ Р С”Р В»Р В°РЎРѓРЎвЂљР ВµРЎР‚Р В° Р Р† РЎРѓРЎвЂљР В°Р В±Р С‘Р В»РЎРЉР Р…Р С•Р С Р С—Р С•РЎР‚РЎРЏР Т‘Р С”Р Вµ.
func (world *World) secondaryClusterObjectIDsLocked(mainID int64) []int64 {
	objectIDs := make([]int64, 0)
	for _, cosmicObject := range world.data.CosmicObjects.Items {
		if cosmicObject != nil && cosmicObject.ClusterMainCosmicObjectID == mainID && cosmicObject.ID != mainID {
			objectIDs = append(objectIDs, cosmicObject.ID)
		}
	}
	sort.Slice(objectIDs, func(left int, right int) bool {
		return objectIDs[left] < objectIDs[right]
	})
	return objectIDs
}

// cosmicObjectHasPassengerSeatLocked Р С—РЎР‚Р С•Р Р†Р ВµРЎР‚РЎРЏР ВµРЎвЂљ РЎС“РЎРѓРЎвЂљР В°Р Р…Р С•Р Р†Р В»Р ВµР Р…Р Р…Р С•Р Вµ Р С—Р В°РЎРѓРЎРѓР В°Р В¶Р С‘РЎР‚РЎРѓР С”Р С•Р Вµ Р С”РЎР‚Р ВµРЎРѓР В»Р С• Р С—Р С• Р В°Р С”РЎР‚Р С•Р Р…Р С‘Р СРЎС“ РЎвЂљР С‘Р С—Р В° Р С—РЎР‚Р ВµР Т‘Р СР ВµРЎвЂљР В°.
func (world *World) cosmicObjectHasPassengerSeatLocked(objectID int64) bool {
	if world.data.EquipmentGroups == nil || world.data.ItemModels == nil || world.data.ItemTypes == nil {
		return false
	}
	for _, group := range world.data.EquipmentGroups.GetByCosmicObjectID(objectID) {
		if group == nil || group.Count <= 0 {
			continue
		}
		model, ok := world.data.ItemModels.Get(group.EquipmentItemModelID)
		if !ok {
			continue
		}
		itemType, ok := world.data.ItemTypes.Get(model.ItemTypeID)
		if ok && itemType.Acronym == "PassengerSeat" {
			return true
		}
	}
	return false
}

// landingRequestIndexByReceiverLocked Р С‘РЎвЂ°Р ВµРЎвЂљ Р В°Р С”РЎвЂљР С‘Р Р†Р Р…РЎвЂ№Р в„– Р Р†РЎвЂ¦Р С•Р Т‘РЎРЏРЎвЂ°Р С‘Р в„– Р В·Р В°Р С—РЎР‚Р С•РЎРѓ Р С—Р С•РЎРѓР В°Р Т‘Р С”Р С‘ Р Т‘Р В»РЎРЏ Р С•Р В±РЎР‰Р ВµР С”РЎвЂљР В°.
func (world *World) landingRequestIndexByReceiverLocked(receiverID int64) int {
	for index, request := range world.landingRequests {
		if request.ReceiverCosmicObjectID == receiverID {
			return index
		}
	}
	return -1
}

// removeLandingRequestLocked РЎС“Р Т‘Р В°Р В»РЎРЏР ВµРЎвЂљ Р В·Р В°Р С—РЎР‚Р С•РЎРѓ Р С—Р С•РЎРѓР В°Р Т‘Р С”Р С‘ Р В±Р ВµР В· РЎРѓР С•РЎвЂ¦РЎР‚Р В°Р Р…Р ВµР Р…Р С‘РЎРЏ Р С—Р С•РЎР‚РЎРЏР Т‘Р С”Р В°.
func (world *World) removeLandingRequestLocked(index int) {
	world.landingRequests = append(world.landingRequests[:index], world.landingRequests[index+1:]...)
}

// addLandingWindowEventsLocked Р Т‘Р С•Р В±Р В°Р Р†Р В»РЎРЏР ВµРЎвЂљ Р С—Р В°РЎР‚Р Р…РЎвЂ№Р Вµ РЎРѓР С•Р В±РЎвЂ№РЎвЂљР С‘РЎРЏ Р С•Р С”Р Р…Р В° Р Т‘Р В»РЎРЏ Р С—Р ВµРЎР‚Р ВµРЎРѓР В°Р В¶Р С‘Р Р†Р В°РЎР‹РЎвЂ°Р ВµР С–Р С•РЎРѓРЎРЏ Р С—Р ВµРЎР‚РЎРѓР С•Р Р…Р В°Р В¶Р В° Р С‘ Р С•Р В±РЎР‰Р ВµР С”РЎвЂљР В° Р Р…Р В°Р В·Р Р…Р В°РЎвЂЎР ВµР Р…Р С‘РЎРЏ.
func (world *World) addLandingWindowEventsLocked(kind string, senderID int64, receiverID int64) {
	world.dockingEvents = append(world.dockingEvents,
		game.DockingEvent{Type: "dockingEvent", Kind: kind, Role: "sender", ObjectIDs: []int64{senderID}},
		game.DockingEvent{Type: "dockingEvent", Kind: kind, Role: "receiver", ObjectIDs: []int64{receiverID}},
	)
}

// addLandingTargetSelectionLocked РЎРѓР С•Р С•Р В±РЎвЂ°Р В°Р ВµРЎвЂљ Р С”Р В»Р С‘Р ВµР Р…РЎвЂљРЎС“, РЎвЂЎРЎвЂљР С• Р Р…РЎС“Р В¶Р Р…Р С• Р Р†РЎвЂ№Р В±РЎР‚Р В°РЎвЂљРЎРЉ Р Р†РЎвЂљР С•РЎР‚Р С•РЎРѓРЎвЂљР ВµР С—Р ВµР Р…Р Р…РЎвЂ№Р в„– Р С•Р В±РЎР‰Р ВµР С”РЎвЂљ Р Р…Р В°Р В·Р Р…Р В°РЎвЂЎР ВµР Р…Р С‘РЎРЏ.
func (world *World) addLandingTargetSelectionLocked(senderID int64) {
	sender, ok := world.data.CosmicObjects.Get(senderID)
	if !ok {
		return
	}
	world.dockingEvents = append(world.dockingEvents, game.DockingEvent{
		Type:      "dockingEvent",
		Kind:      "landingTargetSelection",
		ObjectIDs: []int64{senderID},
		TargetIDs: world.secondaryClusterObjectIDsLocked(sender.ClusterMainCosmicObjectID),
	})
}

// dockingReceiverHasDecisionMakerLocked Р С—РЎР‚Р С•Р Р†Р ВµРЎР‚РЎРЏР ВµРЎвЂљ, РЎвЂЎРЎвЂљР С• Р СџР С•Р В»РЎС“РЎвЂЎР В°РЎвЂљР ВµР В»Р ВµР С РЎРѓР ВµР в„–РЎвЂЎР В°РЎРѓ Р С”РЎвЂљР С•-РЎвЂљР С• РЎС“Р С—РЎР‚Р В°Р Р†Р В»РЎРЏР ВµРЎвЂљ.
func (world *World) dockingReceiverHasDecisionMakerLocked(receiverID int64) bool {
	for _, objectID := range world.accountObjectIDs {
		if objectID == receiverID {
			return true
		}
	}
	return false
}

// addDockingWindowEventsLocked Р Т‘Р С•Р В±Р В°Р Р†Р В»РЎРЏР ВµРЎвЂљ Р С—Р В°РЎР‚Р Р…РЎвЂ№Р Вµ РЎРѓР С•Р В±РЎвЂ№РЎвЂљР С‘РЎРЏ Р С•Р С”Р Р…Р В° Р Т‘Р В»РЎРЏ Р С›РЎвЂљР С—РЎР‚Р В°Р Р†Р С‘РЎвЂљР ВµР В»РЎРЏ Р С‘ Р СџР С•Р В»РЎС“РЎвЂЎР В°РЎвЂљР ВµР В»РЎРЏ.
func (world *World) addDockingWindowEventsLocked(kind string, senderID int64, receiverID int64, duration float64) {
	world.dockingEvents = append(world.dockingEvents,
		game.DockingEvent{Type: "dockingEvent", Kind: kind, Role: "sender", Duration: duration, ObjectIDs: []int64{senderID}},
		game.DockingEvent{Type: "dockingEvent", Kind: kind, Role: "receiver", Duration: duration, ObjectIDs: []int64{receiverID}},
	)
}

// addDockingNotificationLocked Р Т‘Р С•Р В±Р В°Р Р†Р В»РЎРЏР ВµРЎвЂљ РЎС“Р Р†Р ВµР Т‘Р С•Р СР В»Р ВµР Р…Р С‘Р Вµ РЎС“Р С”Р В°Р В·Р В°Р Р…Р Р…РЎвЂ№Р С Р С•Р В±РЎР‰Р ВµР С”РЎвЂљР В°Р С Р В±Р ВµР В· Р Т‘РЎС“Р В±Р В»Р С‘РЎР‚Р С•Р Р†Р В°Р Р…Р С‘РЎРЏ ID.
func (world *World) addDockingNotificationLocked(objectIDs []int64, message string) {
	seen := map[int64]bool{}
	recipients := make([]int64, 0, len(objectIDs))
	for _, objectID := range objectIDs {
		if objectID <= 0 || seen[objectID] {
			continue
		}
		seen[objectID] = true
		recipients = append(recipients, objectID)
	}
	if len(recipients) == 0 {
		return
	}
	world.dockingEvents = append(world.dockingEvents, game.DockingEvent{
		Type:      "dockingEvent",
		Kind:      "dockingNotification",
		Message:   message,
		ObjectIDs: recipients,
	})
}

// clusterObjectIDsLocked Р Р†Р С•Р В·Р Р†РЎР‚Р В°РЎвЂ°Р В°Р ВµРЎвЂљ РЎРѓР С•РЎРѓРЎвЂљР В°Р Р† Р С”Р В»Р В°РЎРѓРЎвЂљР ВµРЎР‚Р В° Р Р† РЎРѓРЎвЂљР В°Р В±Р С‘Р В»РЎРЉР Р…Р С•Р С Р С—Р С•РЎР‚РЎРЏР Т‘Р С”Р Вµ.
func (world *World) clusterObjectIDsLocked(mainID int64) []int64 {
	objectIDs := make([]int64, 0)
	for _, cosmicObject := range world.data.CosmicObjects.Items {
		if cosmicObject != nil && cosmicObject.ClusterMainCosmicObjectID == mainID {
			objectIDs = append(objectIDs, cosmicObject.ID)
		}
	}
	sort.Slice(objectIDs, func(left int, right int) bool {
		return objectIDs[left] < objectIDs[right]
	})
	return objectIDs
}

// Р РЋР Р…Р С‘Р СР В°Р ВµРЎвЂљ РЎРѓ Р С•Р В±РЎР‰Р ВµР С”РЎвЂљР С•Р Р† РЎРѓР С•РЎРѓРЎвЂљР С•РЎРЏР Р…Р С‘Р Вµ Р ВµР Т‘Р С‘Р Р…Р С•Р С–Р С• Р С”Р В»Р В°РЎРѓРЎвЂљР ВµРЎР‚Р В° Р С‘ Р ВµР С–Р С• Р С•Р В±РЎРЏР В·Р В°РЎвЂљР ВµР В»РЎРЉР Р…РЎС“РЎР‹ Р Р…Р ВµР С—Р С•Р Т‘Р Р†Р С‘Р В¶Р Р…Р С•РЎРѓРЎвЂљРЎРЉ.
func (world *World) disbandClusterLocked(mainID int64) {
	for _, clusterObject := range world.data.CosmicObjects.Items {
		if clusterObject != nil && clusterObject.ClusterMainCosmicObjectID == mainID {
			clusterObject.ClusterMainCosmicObjectID = 0
			clusterObject.Anchored = false
		}
	}
}

// Р вЂ™Р С•Р В·Р Р†РЎР‚Р В°РЎвЂ°Р В°Р ВµРЎвЂљ Р Р†РЎР‚Р ВµР СР ВµР Р…Р Р…Р С•Р Вµ РЎвЂљР ВµРЎРѓРЎвЂљР С•Р Р†Р С•Р Вµ Р С‘Р СРЎРЏ Р Р†Р В»Р В°Р Т‘Р ВµР В»РЎРЉРЎвЂ Р В° Р С•Р В±РЎР‰Р ВµР С”РЎвЂљР В° Р Т‘Р В»РЎРЏ Р С”Р В»Р С‘Р ВµР Р…РЎвЂљРЎРѓР С”Р С•Р в„– Р С‘Р Р…РЎвЂћР С•РЎР‚Р СР В°РЎвЂ Р С‘Р С•Р Р…Р Р…Р С•Р в„– Р С—Р В°Р Р…Р ВµР В»Р С‘.
func (world *World) ownerNameForTestingLocked(ownerCharacterID int64) string {
	if ownerCharacterID <= 0 {
		return ""
	}
	character, ok := world.data.Characters.Get(ownerCharacterID)
	if !ok {
		return ""
	}
	account, ok := world.data.Accounts.Get(character.AccountID)
	if !ok {
		return ""
	}
	return account.Nickname
}

// ApplyControlPanelObjectUpdate Р С—РЎР‚Р С‘Р СР ВµР Р…РЎРЏР ВµРЎвЂљ Р С—Р С•Р Т‘РЎвЂљР Р†Р ВµРЎР‚Р В¶Р Т‘Р ВµР Р…Р Р…Р С•Р Вµ Р С‘Р В·Р СР ВµР Р…Р ВµР Р…Р С‘Р Вµ Р С—Р В°Р Р…Р ВµР В»Р С‘ Р С” Р С•Р В±РЎР‰Р ВµР С”РЎвЂљРЎС“ РЎвЂљР ВµР С”РЎС“РЎвЂ°Р ВµР С–Р С• Р В°Р С”Р С”Р В°РЎС“Р Р…РЎвЂљР В°.
func (world *World) ApplyControlPanelObjectUpdate(accountID int64, sessionID string, mutationSeq int64, update ControlPanelObjectUpdate) error {
	world.mu.Lock()
	defer world.mu.Unlock()

	objectID, ok := world.accountObjectIDs[accountID]
	if !ok {
		return errors.New("account is not connected")
	}
	cosmicObject, ok := world.data.CosmicObjects.Get(objectID)
	if !ok {
		return errors.New("controlled object not found")
	}

	if update.Enabled != nil {
		cosmicObject.Enabled = *update.Enabled
	}
	if update.Title != nil {
		cosmicObject.Title = *update.Title
	}
	world.ackMutationLocked(accountID, sessionID, mutationSeq)
	return nil
}

// ApplyControlPanelEquipmentUpdate Р С—РЎР‚Р С‘Р СР ВµР Р…РЎРЏР ВµРЎвЂљ Р С—Р С•Р Т‘РЎвЂљР Р†Р ВµРЎР‚Р В¶Р Т‘Р ВµР Р…Р Р…Р С•Р Вµ Р С‘Р В·Р СР ВµР Р…Р ВµР Р…Р С‘Р Вµ Р С—Р В°Р Р…Р ВµР В»Р С‘ Р С” Р С•Р В±Р С•РЎР‚РЎС“Р Т‘Р С•Р Р†Р В°Р Р…Р С‘РЎР‹ РЎвЂљР ВµР С”РЎС“РЎвЂ°Р ВµР С–Р С• Р С•Р В±РЎР‰Р ВµР С”РЎвЂљР В°.
func (world *World) ApplyControlPanelEquipmentUpdate(accountID int64, sessionID string, mutationSeq int64, update ControlPanelEquipmentUpdate) error {
	world.mu.Lock()
	defer world.mu.Unlock()

	objectID, ok := world.accountObjectIDs[accountID]
	if !ok {
		return errors.New("account is not connected")
	}
	if world.data.EquipmentGroups == nil {
		return errors.New("equipment groups are not loaded")
	}
	group, ok := world.data.EquipmentGroups.Get(update.EquipmentGroupID)
	if !ok {
		return errors.New("equipment group not found")
	}
	if err := world.ensureControlledClusterEquipmentLocked(objectID, group.CosmicObjectID); err != nil {
		return err
	}

	if update.EnabledCount != nil {
		if *update.EnabledCount < 1 || *update.EnabledCount > group.Count {
			return errors.New("enabled equipment count is out of range")
		}
		group.EnabledCount = *update.EnabledCount
	}
	if update.Enabled != nil {
		group.Enabled = *update.Enabled
	}
	if update.Title != nil {
		group.Title = *update.Title
	}
	world.ackMutationLocked(accountID, sessionID, mutationSeq)
	return nil
}

// ApplyControlPanelContainerTransfer Р С—Р ВµРЎР‚Р ВµР Р…Р С•РЎРѓР С‘РЎвЂљ Р Р†РЎРѓРЎвЂ РЎРѓР С•Р Т‘Р ВµРЎР‚Р В¶Р С‘Р СР С•Р Вµ Р С‘Р В· Р С•Р Т‘Р Р…Р С•Р С–Р С• Р С”Р С•Р Р…РЎвЂљР ВµР в„–Р Р…Р ВµРЎР‚Р В° РЎвЂљР ВµР С”РЎС“РЎвЂ°Р ВµР С–Р С• Р С•Р В±РЎР‰Р ВµР С”РЎвЂљР В° Р Р† Р Т‘РЎР‚РЎС“Р С–Р С•Р в„–.
// ApplyControlPanelEquipmentGroupRelationUpdate РЎРѓР С•РЎвЂ¦РЎР‚Р В°Р Р…РЎРЏР ВµРЎвЂљ Р Р†РЎвЂ№Р В±РЎР‚Р В°Р Р…Р Р…РЎС“РЎР‹ РЎРѓР Р†РЎРЏР В·Р В°Р Р…Р Р…РЎС“РЎР‹ Р С–РЎР‚РЎС“Р С—Р С—РЎС“ Р С•Р В±Р С•РЎР‚РЎС“Р Т‘Р С•Р Р†Р В°Р Р…Р С‘РЎРЏ Р Т‘Р В»РЎРЏ РЎвЂљР ВµР С”РЎС“РЎвЂ°Р ВµР С–Р С• Р С•Р В±РЎР‰Р ВµР С”РЎвЂљР В°.
func (world *World) ApplyControlPanelEquipmentGroupRelationUpdate(accountID int64, sessionID string, mutationSeq int64, update ControlPanelEquipmentGroupRelationUpdate) error {
	world.mu.Lock()
	defer world.mu.Unlock()

	objectID, ok := world.accountObjectIDs[accountID]
	if !ok {
		return errors.New("account is not connected")
	}
	if world.data.EquipmentGroups == nil {
		return errors.New("equipment group data is not loaded")
	}
	group, ok := world.data.EquipmentGroups.Get(update.EquipmentGroupID)
	if !ok {
		return errors.New("equipment group not found")
	}
	if err := world.ensureControlledClusterEquipmentLocked(objectID, group.CosmicObjectID); err != nil {
		return err
	}
	if _, err := world.controlledContainerEquipmentLocked(objectID, update.RelatedEquipmentGroupID); err != nil {
		return err
	}
	switch update.RelationTypeAcronym {
	case "Source":
		group.SourceEquipmentGroupID = update.RelatedEquipmentGroupID
	case "Destination":
		group.DestinationEquipmentGroupID = update.RelatedEquipmentGroupID
	case "Opposite":
		group.OppositeEquipmentGroupID = update.RelatedEquipmentGroupID
	default:
		return errors.New("unknown equipment group relation")
	}
	world.ackMutationLocked(accountID, sessionID, mutationSeq)
	return nil
}

func (world *World) ApplyControlPanelContainerTransfer(accountID int64, sessionID string, mutationSeq int64, transfer ControlPanelContainerTransfer) error {
	world.mu.Lock()
	defer world.mu.Unlock()

	objectID, ok := world.accountObjectIDs[accountID]
	if !ok {
		return errors.New("account is not connected")
	}
	if world.data.EquipmentGroups == nil || world.data.ItemGroups == nil || world.data.ItemModels == nil || world.data.Tasks == nil || world.data.TaskTypes == nil || world.data.TaskItemGroups == nil {
		return errors.New("cargo movement data is not loaded")
	}
	controllerID := transfer.ControllerEquipmentGroupID
	if controllerID <= 0 {
		controllerID = transfer.TargetContainerEquipmentGroupID
	}
	controller, err := world.controlledContainerEquipmentLocked(objectID, controllerID)
	if err != nil {
		return err
	}
	source, target, err := world.cargoMovementEndpointsLocked(controller, transfer.LeftToRightDirection)
	if err != nil {
		return err
	}
	if source.ID == target.ID {
		return errors.New("source and target containers must be different")
	}
	taskType, ok := world.data.TaskTypes.GetByAcronym("CargoMovement")
	if !ok {
		return errors.New("cargo movement task type not found")
	}

	movedByModel := make(map[int64]float64)
	for _, itemGroupID := range transfer.ItemGroupIDs {
		itemGroup, ok := world.data.ItemGroups.Get(itemGroupID)
		if !ok {
			return errors.New("item group not found")
		}
		if itemGroup.ContainerEquipmentGroupID != source.ID {
			return errors.New("item group does not belong to source container")
		}
		moved := itemGroup.Count
		if transfer.Amount > 0 && len(transfer.ItemGroupIDs) == 1 {
			moved = math.Min(itemGroup.Count, transfer.Amount)
		}
		if moved <= physics.Epsilon {
			continue
		}
		movedByModel[itemGroup.ContentItemModelID] += moved
	}
	if len(movedByModel) == 0 {
		return errors.New("cargo movement amount is empty")
	}
	modelIDs := make([]int64, 0, len(movedByModel))
	for itemModelID := range movedByModel {
		modelIDs = append(modelIDs, itemModelID)
	}
	sort.Slice(modelIDs, func(left int, right int) bool { return modelIDs[left] < modelIDs[right] })
	for _, itemModelID := range modelIDs {
		count := movedByModel[itemModelID]
		itemModel, ok := world.data.ItemModels.Get(itemModelID)
		if !ok {
			return errors.New("cargo item model not found")
		}
		distance, err := world.cargoMovementDistanceLocked(source.CosmicObjectID, target.CosmicObjectID)
		if err != nil {
			return err
		}
		totalEnergy := itemModel.Mass * count * distance
		if totalEnergy <= physics.Epsilon {
			totalEnergy = physics.Epsilon
		}
		task, err := world.data.Tasks.Add(&data.Task{
			ControllerEquipmentGroupID: controller.ID,
			TaskTypeID:                 taskType.ID,
			RemainingEnergy:            totalEnergy,
			TotalEnergy:                totalEnergy,
			LeftToRightDirection:       transfer.LeftToRightDirection,
			BatchCount:                 1,
		})
		if err != nil {
			return err
		}
		if _, err := world.data.TaskItemGroups.Add(&data.TaskItemGroup{TaskID: task.ID, ItemModelID: itemModelID, Count: count}); err != nil {
			return err
		}
	}
	world.ackMutationLocked(accountID, sessionID, mutationSeq)
	return nil
}

// controlledContainerEquipmentLocked Р Р†Р С•Р В·Р Р†РЎР‚Р В°РЎвЂ°Р В°Р ВµРЎвЂљ Р С”Р С•Р Р…РЎвЂљР ВµР в„–Р Р…Р ВµРЎР‚ РЎвЂљР ВµР С”РЎС“РЎвЂ°Р ВµР С–Р С• Р С•Р В±РЎР‰Р ВµР С”РЎвЂљР В°; Р Р†РЎвЂ№Р В·РЎвЂ№Р Р†Р В°Р ВµРЎвЂљРЎРѓРЎРЏ РЎвЂљР С•Р В»РЎРЉР С”Р С• Р С—Р С•Р Т‘ mutex.
// ApplyControlPanelFuelTransfer Р С—Р ВµРЎР‚Р ВµР Р…Р С•РЎРѓР С‘РЎвЂљ РЎвЂљР С•Р С—Р В»Р С‘Р Р†Р С• Р СР ВµР В¶Р Т‘РЎС“ Р С”Р С•Р Р…РЎвЂљР ВµР в„–Р Р…Р ВµРЎР‚Р С•Р С Р С‘ Р С•Р В±РЎвЂ°Р С‘Р С Р В·Р В°Р С—Р В°РЎРѓР С•Р С РЎвЂљР ВµР С”РЎС“РЎвЂ°Р ВµР С–Р С• Р С•Р В±РЎР‰Р ВµР С”РЎвЂљР В°.
// cargoMovementEndpointsLocked находит текущие контейнеры для направления складского перемещения.
func (world *World) cargoMovementEndpointsLocked(controller *data.EquipmentGroup, leftToRight bool) (*data.EquipmentGroup, *data.EquipmentGroup, error) {
	if controller == nil {
		return nil, nil, errors.New("movement controller container not found")
	}
	opposite, ok := world.data.EquipmentGroups.Get(controller.OppositeEquipmentGroupID)
	if !ok {
		return nil, nil, errors.New("opposite container equipment group not found")
	}
	if !world.equipmentGroupIsContainerLocked(opposite) {
		return nil, nil, errors.New("opposite equipment group is not a container")
	}
	if err := world.ensureControlledClusterEquipmentLocked(controller.CosmicObjectID, opposite.CosmicObjectID); err != nil {
		return nil, nil, err
	}
	if leftToRight {
		return opposite, controller, nil
	}
	return controller, opposite, nil
}

func (world *World) ApplyControlPanelFuelTransfer(accountID int64, sessionID string, mutationSeq int64, transfer ControlPanelFuelTransfer) error {
	world.mu.Lock()
	defer world.mu.Unlock()

	objectID, ok := world.accountObjectIDs[accountID]
	if !ok {
		return errors.New("account is not connected")
	}
	if world.data.EquipmentGroups == nil || world.data.ItemGroups == nil || world.data.Tasks == nil || world.data.TaskTypes == nil || world.data.TaskItemGroups == nil || world.data.ItemModels == nil {
		return errors.New("equipment or item groups are not loaded")
	}
	container, err := world.controlledContainerEquipmentLocked(objectID, transfer.ContainerEquipmentGroupID)
	if err != nil {
		return err
	}
	fuelTankGroup, ok := world.data.EquipmentGroups.Get(transfer.FuelTankEquipmentGroupID)
	if !ok {
		return errors.New("equipment group not found")
	}
	cosmicObject, ok := world.data.CosmicObjects.Get(fuelTankGroup.CosmicObjectID)
	if !ok {
		return errors.New("fuel tank object not found")
	}
	fuelModelID, err := world.controlledFuelTankFuelModelIDLocked(objectID, transfer.FuelTankEquipmentGroupID)
	if err != nil {
		return err
	}
	taskType, ok := world.data.TaskTypes.GetByAcronym("Fueling")
	if !ok {
		return errors.New("fueling task type not found")
	}
	if len(transfer.ItemGroupIDs) > 0 {
		amount, err := world.fuelFillAmountLocked(cosmicObject, container.ID, fuelModelID, transfer.ItemGroupIDs, transfer.Amount)
		if err != nil {
			return err
		}
		if amount > physics.Epsilon {
			totalEnergy, err := world.fuelingTaskEnergyLocked(container.ID, fuelTankGroup.ID, fuelModelID, amount)
			if err != nil {
				return err
			}
			task, err := world.data.Tasks.Add(&data.Task{
				ControllerEquipmentGroupID:      fuelTankGroup.ID,
				TaskTypeID:                      taskType.ID,
				RemainingEnergy:                 totalEnergy,
				TotalEnergy:                     totalEnergy,
				BatchCount:                      1,
				LeftToRightDirection:            true,
				SourceContainerEquipmentGroupID: container.ID,
				FuelTankEquipmentGroupID:        fuelTankGroup.ID,
			})
			if err != nil {
				return err
			}
			if _, err := world.data.TaskItemGroups.Add(&data.TaskItemGroup{TaskID: task.ID, ItemModelID: fuelModelID, Count: amount}); err != nil {
				return err
			}
		}
	} else if transfer.Amount > 0 {
		amount := math.Min(cosmicObject.Fuel, transfer.Amount)
		if amount <= physics.Epsilon {
			world.ackMutationLocked(accountID, sessionID, mutationSeq)
			return nil
		}
		totalEnergy, err := world.fuelingTaskEnergyLocked(container.ID, fuelTankGroup.ID, fuelModelID, amount)
		if err != nil {
			return err
		}
		task, err := world.data.Tasks.Add(&data.Task{
			ControllerEquipmentGroupID:      fuelTankGroup.ID,
			TaskTypeID:                      taskType.ID,
			RemainingEnergy:                 totalEnergy,
			TotalEnergy:                     totalEnergy,
			BatchCount:                      1,
			LeftToRightDirection:            false,
			SourceContainerEquipmentGroupID: container.ID,
			FuelTankEquipmentGroupID:        fuelTankGroup.ID,
		})
		if err != nil {
			return err
		}
		if _, err := world.data.TaskItemGroups.Add(&data.TaskItemGroup{TaskID: task.ID, ItemModelID: fuelModelID, Count: amount}); err != nil {
			return err
		}
	}
	world.ackMutationLocked(accountID, sessionID, mutationSeq)
	return nil
}

// ApplyControlPanelConstructorProduceItem Р С‘Р В·Р С–Р С•РЎвЂљР В°Р Р†Р В»Р С‘Р Р†Р В°Р ВµРЎвЂљ Р С•Р Т‘Р Р…РЎС“ Р С—Р В°РЎР‚РЎвЂљР С‘РЎР‹ Р С—РЎР‚Р ВµР Т‘Р СР ВµРЎвЂљР С•Р Р† Р С—Р С• Р Р†РЎвЂ№Р В±РЎР‚Р В°Р Р…Р Р…Р С•Р в„– РЎРѓРЎвЂ¦Р ВµР СР Вµ.
// ApplyControlPanelItemDeconstruction ставит выбранные предметы в очередь деконструктора.
func (world *World) ApplyControlPanelItemDeconstruction(accountID int64, sessionID string, mutationSeq int64, deconstruction ControlPanelItemDeconstruction) error {
	world.mu.Lock()
	defer world.mu.Unlock()

	objectID, ok := world.accountObjectIDs[accountID]
	if !ok {
		return errors.New("account is not connected")
	}
	if world.data.EquipmentGroups == nil || world.data.ItemGroups == nil || world.data.ItemModels == nil || world.data.Tasks == nil || world.data.TaskTypes == nil || world.data.TaskItemGroups == nil || world.data.Schemas == nil || world.data.SchemaComponents == nil {
		return errors.New("deconstruction data is not loaded")
	}
	if _, err := world.controlledEquipmentitemTypeLocked(objectID, deconstruction.DeconstructorEquipmentGroupID, "Deconstructor"); err != nil {
		return err
	}
	sourceContainer, err := world.controlledContainerEquipmentLocked(objectID, deconstruction.SourceContainerEquipmentGroupID)
	if err != nil {
		return err
	}
	targetContainer, err := world.controlledContainerEquipmentLocked(objectID, deconstruction.TargetContainerEquipmentGroupID)
	if err != nil {
		return err
	}
	taskType, ok := world.data.TaskTypes.GetByAcronym("ItemDeconstruction")
	if !ok {
		return errors.New("item deconstruction task type not found")
	}
	queued := 0
	for _, itemGroupID := range deconstruction.ItemGroupIDs {
		itemGroup, ok := world.data.ItemGroups.Get(itemGroupID)
		if !ok || itemGroup.ContainerEquipmentGroupID != sourceContainer.ID {
			continue
		}
		schema, err := world.cheapestItemSchemaByProductModelLocked(itemGroup.ContentItemModelID)
		if err != nil || schema.Count <= 0 {
			continue
		}
		selectedCount := itemGroup.Count
		if deconstruction.Amount > 0 && selectedCount > deconstruction.Amount {
			selectedCount = deconstruction.Amount
		}
		batches := math.Floor(selectedCount/schema.Count + physics.Epsilon)
		if batches <= 0 {
			continue
		}
		reservedCount := batches * schema.Count
		task, err := world.data.Tasks.Add(&data.Task{
			ControllerEquipmentGroupID:      deconstruction.DeconstructorEquipmentGroupID,
			TaskTypeID:                      taskType.ID,
			RemainingEnergy:                 schema.ProductionEnergy * batches,
			TotalEnergy:                     schema.ProductionEnergy * batches,
			BatchCount:                      int64(batches),
			SchemaID:                        schema.ID,
			SourceContainerEquipmentGroupID: sourceContainer.ID,
			TargetContainerEquipmentGroupID: targetContainer.ID,
		})
		if err != nil {
			return err
		}
		if _, err := world.data.TaskItemGroups.Add(&data.TaskItemGroup{TaskID: task.ID, ItemModelID: itemGroup.ContentItemModelID, Count: reservedCount}); err != nil {
			return err
		}
		queued++
	}
	if queued == 0 {
		return errors.New("deconstruction items not found")
	}
	world.ackMutationLocked(accountID, sessionID, mutationSeq)
	return nil
}

func (world *World) ApplyControlPanelConstructorProduceItem(accountID int64, sessionID string, mutationSeq int64, production ControlPanelConstructorProduceItem) error {
	world.mu.Lock()
	defer world.mu.Unlock()

	objectID, ok := world.accountObjectIDs[accountID]
	if !ok {
		return errors.New("account is not connected")
	}
	if world.data.EquipmentGroups == nil || world.data.ItemGroups == nil || world.data.ItemModels == nil || world.data.ItemTypes == nil || world.data.Tasks == nil || world.data.TaskTypes == nil || world.data.TaskItemGroups == nil {
		return errors.New("constructor data is not loaded")
	}
	if _, err := world.controlledEquipmentitemTypeLocked(objectID, production.ConstructorEquipmentGroupID, "Constructor"); err != nil {
		return err
	}
	materialContainer, err := world.constructorRelatedContainerOrFallbackLocked(objectID, production.ConstructorEquipmentGroupID, "Source", production.MaterialContainerEquipmentGroupID)
	if err != nil {
		return err
	}
	constructorGroup, ok := world.data.EquipmentGroups.Get(production.ConstructorEquipmentGroupID)
	if !ok {
		return errors.New("constructor equipment group not found")
	}
	constructorGroup.SourceEquipmentGroupID = materialContainer.ID
	if (production.SchemaID <= 0) == (production.BlueprintID <= 0) {
		return errors.New("constructor production must select schema or blueprint")
	}
	mainJob, components, amount, err := world.newMainConstructorProductionJobLocked(objectID, production, materialContainer.ID)
	if err != nil {
		return err
	}
	if mainJob.ProductContainerEquipmentGroupID > 0 {
		constructorGroup.DestinationEquipmentGroupID = mainJob.ProductContainerEquipmentGroupID
	}

	requiredByModel := map[int64]float64{}
	for _, component := range components {
		if component.Count <= 0 {
			return errors.New("constructor component count is invalid")
		}
		if _, ok := world.data.ItemModels.Get(component.ComponentItemModelID); !ok {
			return errors.New("constructor component item model not found")
		}
		requiredByModel[component.ComponentItemModelID] += component.Count
	}
	availableByModel := map[int64]float64{}
	for _, itemGroup := range world.data.ItemGroups.GetByContainerEquipmentGroupID(materialContainer.ID) {
		availableByModel[itemGroup.ContentItemModelID] += itemGroup.Count
	}
	plannedJobs := make([]constructorProductionJob, 0)
	for itemModelID, required := range requiredByModel {
		if err := world.planMissingConstructorComponentsLocked(&plannedJobs, production.ConstructorEquipmentGroupID, materialContainer.ID, availableByModel, itemModelID, required*float64(amount), mainJob.ID, map[int64]bool{}); err != nil {
			return err
		}
	}
	plannedJobs = append(plannedJobs, mainJob)
	if err := world.addConstructorTasksLocked(plannedJobs); err != nil {
		return err
	}
	world.ackMutationLocked(accountID, sessionID, mutationSeq)
	return nil
}

// ApplyControlPanelConstructorQueueCommand Р СР ВµР Р…РЎРЏР ВµРЎвЂљ Р С•РЎРѓР Р…Р С•Р Р†Р Р…РЎС“РЎР‹ Р С•РЎвЂЎР ВµРЎР‚Р ВµР Т‘РЎРЉ Р С”Р С•Р Р…РЎРѓРЎвЂљРЎР‚РЎС“Р С”РЎвЂљР С•РЎР‚Р В° Р С—Р С• Р Р†РЎвЂ№Р В±РЎР‚Р В°Р Р…Р Р…Р С•Р в„– РЎРѓРЎвЂљРЎР‚Р С•Р С”Р Вµ.
func (world *World) ApplyControlPanelConstructorQueueCommand(accountID int64, sessionID string, mutationSeq int64, command ControlPanelConstructorQueueCommand) error {
	world.mu.Lock()
	defer world.mu.Unlock()

	objectID, ok := world.accountObjectIDs[accountID]
	if !ok {
		return errors.New("account is not connected")
	}
	constructorController := true
	if _, err := world.controlledEquipmentitemTypeLocked(objectID, command.ConstructorEquipmentGroupID, "Constructor"); err != nil {
		constructorController = false
		if _, err := world.controlledContainerEquipmentLocked(objectID, command.ConstructorEquipmentGroupID); err != nil {
			if _, fuelTankErr := world.controlledFuelTankFuelModelIDLocked(objectID, command.ConstructorEquipmentGroupID); fuelTankErr != nil {
				if _, deconstructorErr := world.controlledEquipmentitemTypeLocked(objectID, command.ConstructorEquipmentGroupID, "Deconstructor"); deconstructorErr != nil {
					return err
				}
			}
		}
	}
	task := world.queueTaskLocked(command.ConstructorEquipmentGroupID, command.JobID)
	if constructorController {
		task = world.constructorMainTaskLocked(command.ConstructorEquipmentGroupID, command.JobID)
	}
	if task == nil {
		return errors.New("queue job not found")
	}
	switch command.Command {
	case "skipNext":
		if task.RemainingEnergy < task.TotalEnergy || world.taskHasReserveLocked(task.ID) {
			world.trimTaskToStartedCountLocked(task)
			world.removeConstructorMainTasksAfterLocked(command.ConstructorEquipmentGroupID, command.JobID)
		} else {
			world.removeConstructorTaskLocked(task.ID)
		}
	case "skipAllNext":
		if task.RemainingEnergy < task.TotalEnergy || world.taskHasReserveLocked(task.ID) {
			world.trimTaskToStartedCountLocked(task)
		}
		world.removeConstructorMainTasksAfterLocked(command.ConstructorEquipmentGroupID, command.JobID)
		world.removeConstructorTaskLocked(task.ID)
	case "cancel":
		world.removeConstructorTaskLocked(task.ID)
	case "cancelAll":
		world.removeConstructorMainTasksFromLocked(command.ConstructorEquipmentGroupID, command.JobID)
	default:
		return errors.New("constructor queue command is invalid")
	}
	world.ackMutationLocked(accountID, sessionID, mutationSeq)
	return nil
}

// trimTaskToStartedCountLocked оставляет в задании только уже начатые единицы результата.
func (world *World) trimTaskToStartedCountLocked(task *data.Task) {
	amount := taskCount(task)
	if amount <= 1 || task.TotalEnergy <= physics.Epsilon {
		return
	}
	perUnitEnergy := task.TotalEnergy / amount
	if perUnitEnergy <= physics.Epsilon {
		return
	}
	elapsedEnergy := math.Max(0, task.TotalEnergy-task.RemainingEnergy)
	startedCount := math.Floor(elapsedEnergy/perUnitEnergy) + 1
	if startedCount < 1 {
		startedCount = 1
	}
	if startedCount >= amount {
		return
	}
	if container, err := world.taskRelatedContainerLocked(task, "Source"); err == nil {
		components, err := world.taskComponentsLocked(task)
		if err == nil {
			keptByModel := map[int64]float64{}
			for _, component := range components {
				keptByModel[component.ComponentItemModelID] += component.Count * startedCount
			}
			for _, group := range world.data.TaskItemGroups.GetByTaskID(task.ID) {
				if group == nil {
					continue
				}
				keptCount := keptByModel[group.ItemModelID]
				if group.Count > keptCount {
					_ = world.addItemModelToContainerLocked(container.ID, group.ItemModelID, group.Count-keptCount)
					group.Count = keptCount
				}
			}
			_ = world.data.ItemGroups.RebuildIndexes()
		}
	}
	task.BatchCount = int64(startedCount)
	task.TotalEnergy = perUnitEnergy * startedCount
	task.RemainingEnergy = math.Max(0, task.TotalEnergy-elapsedEnergy)
}

// planMissingConstructorComponentsLocked Р Т‘Р С•Р В±Р В°Р Р†Р В»РЎРЏР ВµРЎвЂљ Р Р†Р С• Р Р†РЎРѓР С—Р С•Р СР С•Р С–Р В°РЎвЂљР ВµР В»РЎРЉР Р…РЎС“РЎР‹ Р С•РЎвЂЎР ВµРЎР‚Р ВµР Т‘РЎРЉ РЎвЂљР С•Р В»РЎРЉР С”Р С• Р Р…Р ВµР Т‘Р С•РЎРѓРЎвЂљР В°РЎР‹РЎвЂ°Р ВµР Вµ Р С”Р С•Р В»Р С‘РЎвЂЎР ВµРЎРѓРЎвЂљР Р†Р С• Р С”Р С•Р СР С—Р С•Р Р…Р ВµР Р…РЎвЂљР С•Р Р†.
// newMainConstructorProductionJobLocked РЎРѓР С•Р В±Р С‘РЎР‚Р В°Р ВµРЎвЂљ Р С•РЎРѓР Р…Р С•Р Р†Р Р…РЎС“РЎР‹ РЎРѓРЎвЂљРЎР‚Р С•Р С”РЎС“ Р С•РЎвЂЎР ВµРЎР‚Р ВµР Т‘Р С‘ Р С—Р С• РЎРѓРЎвЂ¦Р ВµР СР Вµ Р С‘Р В»Р С‘ РЎвЂЎР ВµРЎР‚РЎвЂљР ВµР В¶РЎС“.
func (world *World) newMainConstructorProductionJobLocked(objectID int64, production ControlPanelConstructorProduceItem, materialContainerID int64) (constructorProductionJob, []controlPanelItemSchemaComponent, int64, error) {
	amount := production.Amount
	if amount <= 0 {
		amount = 1
	}
	if production.SchemaID > 0 {
		productContainer, err := world.constructorRelatedContainerOrFallbackLocked(objectID, production.ConstructorEquipmentGroupID, "Destination", production.ProductContainerEquipmentGroupID)
		if err != nil {
			return constructorProductionJob{}, nil, 0, err
		}
		schema, err := world.itemSchemaLocked(production.SchemaID)
		if err != nil {
			return constructorProductionJob{}, nil, 0, err
		}
		if schema.Count <= 0 {
			return constructorProductionJob{}, nil, 0, errors.New("schema product count is invalid")
		}
		if _, ok := world.data.ItemModels.Get(schema.ItemModelID); !ok {
			return constructorProductionJob{}, nil, 0, errors.New("schema product item model not found")
		}
		components, err := world.itemSchemaComponentsLocked(schema.ID)
		if err != nil {
			return constructorProductionJob{}, nil, 0, err
		}
		if len(components) == 0 {
			return constructorProductionJob{}, nil, 0, errors.New("schema components not found")
		}
		return world.newConstructorProductionJobLocked(production.ConstructorEquipmentGroupID, materialContainerID, productContainer.ID, "main", schema, amount, 0), components, amount, nil
	}
	if world.data.Blueprints == nil || world.data.BlueprintComponents == nil || world.data.CosmicObjectModels == nil {
		return constructorProductionJob{}, nil, 0, errors.New("blueprint data is not loaded")
	}
	blueprint, err := world.objectBlueprintLocked(production.BlueprintID)
	if err != nil {
		return constructorProductionJob{}, nil, 0, err
	}
	if _, ok := world.data.CosmicObjectModels.Get(blueprint.CosmicObjectModelID); !ok {
		return constructorProductionJob{}, nil, 0, errors.New("blueprint object model not found")
	}
	components, err := world.objectBlueprintComponentsLocked(blueprint.ID)
	if err != nil {
		return constructorProductionJob{}, nil, 0, err
	}
	if len(components) == 0 {
		return constructorProductionJob{}, nil, 0, errors.New("blueprint components not found")
	}
	return world.newConstructorObjectProductionJobLocked(production.ConstructorEquipmentGroupID, materialContainerID, "main", blueprint, 1, 0), components, 1, nil
}

// constructorMainTaskLocked находит основное задание конструктора по номеру.
func (world *World) constructorMainTaskLocked(constructorID int64, taskID int64) *data.Task {
	if world.data.Tasks == nil || world.data.TaskTypes == nil {
		return nil
	}
	itemProductionType, _ := world.data.TaskTypes.GetByAcronym("ItemProduction")
	objectProductionType, _ := world.data.TaskTypes.GetByAcronym("ObjectProduction")
	for _, task := range world.data.Tasks.GetByControllerEquipmentGroupID(constructorID) {
		if task.ID != taskID || task.ParentTaskID != 0 {
			continue
		}
		if itemProductionType != nil && task.TaskTypeID == itemProductionType.ID {
			return task
		}
		if objectProductionType != nil && task.TaskTypeID == objectProductionType.ID {
			return task
		}
	}
	return nil
}

// removeConstructorMainTasksAfterLocked удаляет основные задания после выбранного.
// queueTaskLocked находит выбранное основное задание контроллера.
func (world *World) queueTaskLocked(controllerID int64, taskID int64) *data.Task {
	if world.data.Tasks == nil {
		return nil
	}
	for _, task := range world.data.Tasks.GetByControllerEquipmentGroupID(controllerID) {
		if task.ID == taskID && task.ParentTaskID == 0 {
			return task
		}
	}
	return nil
}

func (world *World) removeConstructorMainTasksAfterLocked(constructorID int64, taskID int64) {
	seenSelected := false
	for _, task := range append([]*data.Task(nil), world.data.Tasks.GetByControllerEquipmentGroupID(constructorID)...) {
		if task.ParentTaskID != 0 {
			continue
		}
		if seenSelected {
			world.removeConstructorTaskLocked(task.ID)
			continue
		}
		if task.ID == taskID {
			seenSelected = true
		}
	}
}

// removeConstructorMainTasksFromLocked удаляет выбранное основное задание и следующие за ним.
func (world *World) removeConstructorMainTasksFromLocked(constructorID int64, taskID int64) {
	seenSelected := false
	for _, task := range append([]*data.Task(nil), world.data.Tasks.GetByControllerEquipmentGroupID(constructorID)...) {
		if task.ParentTaskID != 0 {
			continue
		}
		if seenSelected || task.ID == taskID {
			seenSelected = true
			world.removeConstructorTaskLocked(task.ID)
		}
	}
}

// removeConstructorTaskLocked удаляет задание и его вспомогательные строки.
func (world *World) removeConstructorTaskLocked(taskID int64) {
	world.returnTaskReserveLocked(taskID)
	world.data.TaskItemGroups.DeleteByTaskID(taskID)
	world.data.Tasks.Delete(taskID)
	dependentTasks := make([]*data.Task, 0)
	for _, task := range world.data.Tasks.Items {
		dependentTasks = append(dependentTasks, task)
	}
	for _, task := range dependentTasks {
		if task != nil && task.ParentTaskID == taskID {
			world.removeConstructorTaskLocked(task.ID)
		}
	}
}

// returnTaskReserveLocked возвращает резерв в доступное место хранения или создает отдельное перемещение.
func (world *World) returnTaskReserveLocked(taskID int64) {
	task, ok := world.data.Tasks.Get(taskID)
	if !ok {
		return
	}
	if world.taskTypeAcronymLocked(task) == "CargoMovement" {
		sourceContainerID := task.SourceContainerEquipmentGroupID
		if sourceContainerID <= 0 {
			return
		}
		for _, group := range world.data.TaskItemGroups.GetByTaskID(taskID) {
			if group == nil || !group.IsStored {
				continue
			}
			_ = world.addItemModelToContainerLocked(sourceContainerID, group.ItemModelID, group.Count)
			group.IsStored = false
		}
		_ = world.data.ItemGroups.RebuildIndexes()
		return
	}
	if world.taskTypeAcronymLocked(task) == "Fueling" {
		world.returnFuelingReserveLocked(task)
		return
	}
	if world.taskTypeAcronymLocked(task) == "ItemDeconstruction" {
		world.returnStoredTaskReserveToContainerLocked(task, task.SourceContainerEquipmentGroupID)
		return
	}
	reservedGroups := world.data.TaskItemGroups.GetByTaskID(taskID)
	if len(reservedGroups) == 0 {
		return
	}
	container, err := world.taskRelatedContainerLocked(task, "Source")
	if err != nil {
		world.moveTaskReserveThroughCargoTaskLocked(task, reservedGroups)
		return
	}
	for _, group := range world.data.TaskItemGroups.GetByTaskID(taskID) {
		_ = world.addItemModelToContainerLocked(container.ID, group.ItemModelID, group.Count)
	}
	_ = world.data.ItemGroups.RebuildIndexes()
}

// returnFuelingReserveLocked возвращает уже взятое топливо при отмене заправки или слива.
// returnStoredTaskReserveToContainerLocked возвращает временно хранимые предметы в указанный контейнер.
func (world *World) returnStoredTaskReserveToContainerLocked(task *data.Task, containerID int64) {
	if task == nil || containerID <= 0 {
		return
	}
	for _, group := range world.data.TaskItemGroups.GetByTaskID(task.ID) {
		if group == nil || !group.IsStored {
			continue
		}
		_ = world.addItemModelToContainerLocked(containerID, group.ItemModelID, group.Count)
		group.IsStored = false
	}
	_ = world.data.ItemGroups.RebuildIndexes()
}

func (world *World) returnFuelingReserveLocked(task *data.Task) {
	if task == nil || task.SourceContainerEquipmentGroupID <= 0 || task.FuelTankEquipmentGroupID <= 0 {
		return
	}
	fuelTank, ok := world.data.EquipmentGroups.Get(task.FuelTankEquipmentGroupID)
	if !ok {
		return
	}
	cosmicObject, ok := world.data.CosmicObjects.Get(fuelTank.CosmicObjectID)
	if !ok {
		return
	}
	for _, group := range world.data.TaskItemGroups.GetByTaskID(task.ID) {
		if group == nil || !group.IsStored {
			continue
		}
		if task.LeftToRightDirection {
			_ = world.addItemModelToContainerLocked(task.SourceContainerEquipmentGroupID, group.ItemModelID, group.Count)
		} else {
			cosmicObject.Fuel = math.Min(cosmicObject.MaxFuel, cosmicObject.Fuel+group.Count)
		}
		group.IsStored = false
	}
	_ = world.data.ItemGroups.RebuildIndexes()
}

// moveTaskReserveThroughCargoTaskLocked передает резерв роботам для возврата в выбранный контейнер.
func (world *World) moveTaskReserveThroughCargoTaskLocked(task *data.Task, reservedGroups []*data.TaskItemGroup) {
	targetContainer := world.firstAvailableContainerForTaskLocked(task, reservedGroups)
	if targetContainer == nil {
		for _, group := range reservedGroups {
			_ = world.addItemModelToContainerLocked(task.ControllerEquipmentGroupID, group.ItemModelID, group.Count)
		}
		_ = world.data.ItemGroups.RebuildIndexes()
		return
	}
	cargoTaskType, ok := world.data.TaskTypes.GetByAcronym("CargoMovement")
	if !ok {
		for _, group := range reservedGroups {
			_ = world.addItemModelToContainerLocked(targetContainer.ID, group.ItemModelID, group.Count)
		}
		_ = world.data.ItemGroups.RebuildIndexes()
		return
	}
	totalEnergy := world.cargoMovementEnergyLocked(reservedGroups)
	cargoTask, err := world.data.Tasks.Add(&data.Task{
		ControllerEquipmentGroupID: targetContainer.ID,
		TaskTypeID:                 cargoTaskType.ID,
		RemainingEnergy:            totalEnergy,
		TotalEnergy:                totalEnergy,
	})
	if err != nil {
		return
	}
	for _, group := range reservedGroups {
		_, _ = world.data.TaskItemGroups.Add(&data.TaskItemGroup{TaskID: cargoTask.ID, ItemModelID: group.ItemModelID, Count: group.Count, IsStored: true})
	}
}

// firstAvailableContainerForTaskLocked выбирает контейнер того же управляемого кластера.
func (world *World) firstAvailableContainerForTaskLocked(task *data.Task, reservedGroups []*data.TaskItemGroup) *data.EquipmentGroup {
	controller, ok := world.data.EquipmentGroups.Get(task.ControllerEquipmentGroupID)
	if !ok {
		return nil
	}
	containers := make([]*data.EquipmentGroup, 0)
	for _, group := range world.data.EquipmentGroups.Items {
		if group == nil || !world.equipmentGroupIsContainerLocked(group) {
			continue
		}
		if err := world.ensureControlledClusterEquipmentLocked(controller.CosmicObjectID, group.CosmicObjectID); err != nil {
			continue
		}
		containers = append(containers, group)
	}
	sort.Slice(containers, func(left int, right int) bool { return containers[left].ID < containers[right].ID })
	for _, container := range containers {
		if world.containerCanAcceptTaskReserveLocked(container.ID, reservedGroups) {
			return container
		}
	}
	return nil
}

// containerCanAcceptTaskReserveLocked оставляет точку для проверки вместимости, когда она появится в модели.
func (world *World) containerCanAcceptTaskReserveLocked(containerID int64, reservedGroups []*data.TaskItemGroup) bool {
	return containerID > 0 && len(reservedGroups) > 0
}

// cargoMovementEnergyLocked задает минимальную работу, чтобы перемещение не завершалось мгновенно.
func (world *World) cargoMovementEnergyLocked(reservedGroups []*data.TaskItemGroup) float64 {
	totalMass := 0.0
	for _, group := range reservedGroups {
		if group == nil {
			continue
		}
		itemModel, ok := world.data.ItemModels.Get(group.ItemModelID)
		if !ok {
			continue
		}
		totalMass += itemModel.Mass * group.Count
	}
	if totalMass <= physics.Epsilon {
		return 1
	}
	return totalMass
}

// cargoMovementDistanceLocked вычисляет путь между контейнерами по размерам объектов.
// cargoMovementTaskEnergyLocked вычисляет работу перемещения по массе требования и расстоянию между контейнерами.
func (world *World) cargoMovementTaskEnergyLocked(sourceContainerID int64, targetContainerID int64, groups []*data.TaskItemGroup) (float64, error) {
	source, ok := world.data.EquipmentGroups.Get(sourceContainerID)
	if !ok {
		return 0, errors.New("movement source container not found")
	}
	target, ok := world.data.EquipmentGroups.Get(targetContainerID)
	if !ok {
		return 0, errors.New("movement target container not found")
	}
	distance, err := world.cargoMovementDistanceLocked(source.CosmicObjectID, target.CosmicObjectID)
	if err != nil {
		return 0, err
	}
	totalMass := 0.0
	for _, group := range groups {
		if group == nil {
			continue
		}
		itemModel, ok := world.data.ItemModels.Get(group.ItemModelID)
		if !ok {
			return 0, errors.New("cargo item model not found")
		}
		totalMass += itemModel.Mass * group.Count
	}
	return math.Max(totalMass*distance, physics.Epsilon), nil
}

func (world *World) cargoMovementDistanceLocked(sourceObjectID int64, targetObjectID int64) (float64, error) {
	if sourceObjectID == targetObjectID {
		return world.cosmicObjectHalfSizeLocked(sourceObjectID)
	}
	source, ok := world.data.CosmicObjects.Get(sourceObjectID)
	if !ok {
		return 0, errors.New("source object not found")
	}
	target, ok := world.data.CosmicObjects.Get(targetObjectID)
	if !ok {
		return 0, errors.New("target object not found")
	}
	sourceMainID := world.clusterMainObjectIDLocked(source)
	targetMainID := world.clusterMainObjectIDLocked(target)
	if sourceMainID <= 0 || sourceMainID != targetMainID {
		return 0, errors.New("cargo movement objects are not in one cluster")
	}
	sourceHalfSize, err := world.cosmicObjectHalfSizeLocked(sourceObjectID)
	if err != nil {
		return 0, err
	}
	targetHalfSize, err := world.cosmicObjectHalfSizeLocked(targetObjectID)
	if err != nil {
		return 0, err
	}
	if sourceObjectID == sourceMainID || targetObjectID == sourceMainID {
		return sourceHalfSize + targetHalfSize, nil
	}
	mainHalfSize, err := world.cosmicObjectHalfSizeLocked(sourceMainID)
	if err != nil {
		return 0, err
	}
	return sourceHalfSize + mainHalfSize*2 + targetHalfSize, nil
}

// clusterMainObjectIDLocked выбирает главный объект кластера.
func (world *World) clusterMainObjectIDLocked(cosmicObject *data.CosmicObject) int64 {
	if cosmicObject == nil {
		return 0
	}
	if cosmicObject.ClusterMainCosmicObjectID > 0 {
		return cosmicObject.ClusterMainCosmicObjectID
	}
	return cosmicObject.ID
}

// cosmicObjectHalfSizeLocked вычисляет полуразмер объекта.
func (world *World) cosmicObjectHalfSizeLocked(cosmicObjectID int64) (float64, error) {
	cosmicObject, ok := world.data.CosmicObjects.Get(cosmicObjectID)
	if !ok {
		return 0, errors.New("cosmic object not found")
	}
	model, ok := world.data.CosmicObjectModels.Get(cosmicObject.CosmicObjectModelID)
	if !ok {
		return 0, errors.New("cosmic object model not found")
	}
	halfSize := (model.BodyLength + model.BodyWidth) / 4
	if halfSize <= physics.Epsilon {
		return 1, nil
	}
	return halfSize, nil
}
func (world *World) planMissingConstructorComponentsLocked(plannedJobs *[]constructorProductionJob, constructorID int64, materialContainerID int64, availableByModel map[int64]float64, itemModelID int64, required float64, parentJobID int64, visiting map[int64]bool) error {
	shortage := required - availableByModel[itemModelID]
	if shortage <= physics.Epsilon {
		availableByModel[itemModelID] -= required
		return nil
	}
	if world.data.Schemas == nil || world.data.SchemaComponents == nil {
		return errors.New("item schema data is not loaded")
	}
	schema, err := world.itemSchemaByProductModelLocked(itemModelID)
	if err != nil {
		return errors.New("not enough schema components")
	}
	if schema.Count <= 0 {
		return errors.New("schema product count is invalid")
	}
	if visiting[schema.ID] {
		return errors.New("schema dependency cycle")
	}
	availableByModel[itemModelID] = 0
	batchCount := int64(math.Ceil(shortage / schema.Count))
	visiting[schema.ID] = true
	job := world.newConstructorProductionJobLocked(constructorID, materialContainerID, materialContainerID, "auxiliary", schema, batchCount, parentJobID)
	components, err := world.itemSchemaComponentsLocked(schema.ID)
	if err != nil {
		return err
	}
	if len(components) == 0 {
		return errors.New("schema components not found")
	}
	for _, component := range components {
		if component.Count <= 0 {
			return errors.New("schema component count is invalid")
		}
		if err := world.planMissingConstructorComponentsLocked(plannedJobs, constructorID, materialContainerID, availableByModel, component.ComponentItemModelID, component.Count*float64(batchCount), job.ID, visiting); err != nil {
			return err
		}
	}
	*plannedJobs = append(*plannedJobs, job)
	availableByModel[itemModelID] += schema.Count * float64(batchCount)
	delete(visiting, schema.ID)
	availableByModel[itemModelID] -= shortage
	return nil
}

// newConstructorProductionJobLocked РЎРѓР С•Р В·Р Т‘Р В°Р ВµРЎвЂљ РЎРѓРЎвЂљРЎР‚Р С•Р С”РЎС“ Р С•РЎвЂЎР ВµРЎР‚Р ВµР Т‘Р С‘ Р С—Р С• РЎРѓРЎвЂ¦Р ВµР СР Вµ; Р Р†РЎвЂ№Р В·РЎвЂ№Р Р†Р В°Р ВµРЎвЂљРЎРѓРЎРЏ РЎвЂљР С•Р В»РЎРЉР С”Р С• Р С—Р С•Р Т‘ mutex.
// addConstructorTasksLocked переносит рассчитанные строки производства в сохраненную очередь заданий.
func (world *World) addConstructorTasksLocked(plannedJobs []constructorProductionJob) error {
	taskIDByJobID := map[int64]int64{}
	orderedJobs := append([]constructorProductionJob(nil), plannedJobs...)
	sort.SliceStable(orderedJobs, func(left int, right int) bool {
		if (orderedJobs[left].ParentJobID == 0) != (orderedJobs[right].ParentJobID == 0) {
			return orderedJobs[left].ParentJobID == 0
		}
		return orderedJobs[left].ID < orderedJobs[right].ID
	})
	for _, job := range orderedJobs {
		taskTypeAcronym := "ItemProduction"
		totalEnergy := 0.0
		if job.BlueprintID > 0 {
			taskTypeAcronym = "ObjectProduction"
			blueprint, err := world.objectBlueprintLocked(job.BlueprintID)
			if err != nil {
				return err
			}
			totalEnergy = math.Max(0, blueprint.ProductionEnergy)
		} else {
			schema, err := world.itemSchemaLocked(job.SchemaID)
			if err != nil {
				return err
			}
			totalEnergy = math.Max(0, schema.ProductionEnergy)
		}
		taskType, ok := world.data.TaskTypes.GetByAcronym(taskTypeAcronym)
		if !ok {
			return errors.New("task type not found")
		}
		batches := job.TotalBatches
		if batches <= 0 {
			batches = 1
		}
		parentTaskID := taskIDByJobID[job.ParentJobID]
		task, err := world.data.Tasks.Add(&data.Task{
			ControllerEquipmentGroupID: job.ConstructorEquipmentGroupID,
			ParentTaskID:               parentTaskID,
			TaskTypeID:                 taskType.ID,
			RemainingEnergy:            totalEnergy * float64(batches),
			TotalEnergy:                totalEnergy * float64(batches),
			BatchCount:                 batches,
			SchemaID:                   job.SchemaID,
			BlueprintID:                job.BlueprintID,
		})
		if err != nil {
			return err
		}
		taskIDByJobID[job.ID] = task.ID
	}
	return nil
}
func (world *World) newConstructorProductionJobLocked(constructorID int64, materialContainerID int64, productContainerID int64, queueType string, schema controlPanelItemSchema, batches int64, parentJobID int64) constructorProductionJob {
	world.nextConstructorProductionJobID++
	totalTime := math.Max(0, schema.ProductionEnergy)
	if batches <= 0 {
		batches = 1
	}
	return constructorProductionJob{
		ID:                                world.nextConstructorProductionJobID,
		ConstructorEquipmentGroupID:       constructorID,
		MaterialContainerEquipmentGroupID: materialContainerID,
		ProductContainerEquipmentGroupID:  productContainerID,
		QueueType:                         queueType,
		SchemaID:                          schema.ID,
		ProductItemModelID:                schema.ItemModelID,
		ProductCount:                      schema.Count,
		RemainingBatches:                  batches,
		TotalBatches:                      batches,
		RemainingTime:                     totalTime,
		TotalTime:                         totalTime,
		ParentJobID:                       parentJobID,
	}
}

// newConstructorObjectProductionJobLocked РЎРѓР С•Р В·Р Т‘Р В°РЎвЂРЎвЂљ РЎРѓРЎвЂљРЎР‚Р С•Р С”РЎС“ Р С•РЎвЂЎР ВµРЎР‚Р ВµР Т‘Р С‘ Р С—Р С• РЎвЂЎР ВµРЎР‚РЎвЂљР ВµР В¶РЎС“ Р С•Р В±РЎР‰Р ВµР С”РЎвЂљР В°; Р Р†РЎвЂ№Р В·РЎвЂ№Р Р†Р В°Р ВµРЎвЂљРЎРѓРЎРЏ РЎвЂљР С•Р В»РЎРЉР С”Р С• Р С—Р С•Р Т‘ mutex.
func (world *World) newConstructorObjectProductionJobLocked(constructorID int64, materialContainerID int64, queueType string, blueprint controlPanelObjectBlueprint, batches int64, parentJobID int64) constructorProductionJob {
	world.nextConstructorProductionJobID++
	totalTime := math.Max(0, blueprint.ProductionEnergy)
	if batches <= 0 {
		batches = 1
	}
	return constructorProductionJob{
		ID:                                world.nextConstructorProductionJobID,
		ConstructorEquipmentGroupID:       constructorID,
		MaterialContainerEquipmentGroupID: materialContainerID,
		QueueType:                         queueType,
		BlueprintID:                       blueprint.ID,
		ProductCosmicObjectModelID:        blueprint.CosmicObjectModelID,
		ProductCount:                      1,
		RemainingBatches:                  batches,
		TotalBatches:                      batches,
		RemainingTime:                     totalTime,
		TotalTime:                         totalTime,
		ParentJobID:                       parentJobID,
	}
}

func (world *World) controlledContainerEquipmentLocked(objectID int64, groupID int64) (*data.EquipmentGroup, error) {
	group, ok := world.data.EquipmentGroups.Get(groupID)
	if !ok {
		return nil, errors.New("equipment group not found")
	}
	if err := world.ensureControlledClusterEquipmentLocked(objectID, group.CosmicObjectID); err != nil {
		return nil, err
	}
	if !world.equipmentGroupIsContainerLocked(group) {
		return nil, errors.New("equipment group is not a container")
	}
	return group, nil
}

// Р СџРЎР‚Р С•Р Р†Р ВµРЎР‚РЎРЏР ВµРЎвЂљ, РЎвЂЎРЎвЂљР С• Р С•Р В±Р С•РЎР‚РЎС“Р Т‘Р С•Р Р†Р В°Р Р…Р С‘Р ВµР С Р С•Р В±РЎР‰Р ВµР С”РЎвЂљР В° Р СР С•Р В¶Р Р…Р С• РЎС“Р С—РЎР‚Р В°Р Р†Р В»РЎРЏРЎвЂљРЎРЉ Р С‘Р В· РЎвЂљР ВµР С”РЎС“РЎвЂ°Р ВµР С–Р С• Р С•Р В±РЎР‰Р ВµР С”РЎвЂљР В°.
func (world *World) ensureControlledClusterEquipmentLocked(controlledObjectID int64, equipmentObjectID int64) error {
	if equipmentObjectID == controlledObjectID {
		return nil
	}
	controlledObject, ok := world.data.CosmicObjects.Get(controlledObjectID)
	if !ok {
		return errors.New("controlled object not found")
	}
	equipmentObject, ok := world.data.CosmicObjects.Get(equipmentObjectID)
	if !ok {
		return errors.New("equipment object not found")
	}
	mainID := controlledObject.ClusterMainCosmicObjectID
	if mainID <= 0 || equipmentObject.ClusterMainCosmicObjectID != mainID {
		return errors.New("equipment group does not belong to controlled object")
	}
	mainObject, ok := world.data.CosmicObjects.Get(mainID)
	if !ok || mainObject.OwnerCharacterID != controlledObject.OwnerCharacterID || controlledObject.OwnerCharacterID <= 0 {
		return errors.New("equipment group does not belong to controlled object")
	}
	if equipmentObject.OwnerCharacterID != controlledObject.OwnerCharacterID {
		return errors.New("equipment object does not belong to character")
	}
	return nil
}

// controlledEquipmentitemTypeLocked Р Р†Р С•Р В·Р Р†РЎР‚Р В°РЎвЂ°Р В°Р ВµРЎвЂљ Р С•Р В±Р С•РЎР‚РЎС“Р Т‘Р С•Р Р†Р В°Р Р…Р С‘Р Вµ РЎвЂљР ВµР С”РЎС“РЎвЂ°Р ВµР С–Р С• Р С•Р В±РЎР‰Р ВµР С”РЎвЂљР В° РЎРѓ Р С•Р В¶Р С‘Р Т‘Р В°Р ВµР СРЎвЂ№Р С РЎвЂљР С‘Р С—Р С•Р С Р С—РЎР‚Р ВµР Т‘Р СР ВµРЎвЂљР В°; Р Р†РЎвЂ№Р В·РЎвЂ№Р Р†Р В°Р ВµРЎвЂљРЎРѓРЎРЏ РЎвЂљР С•Р В»РЎРЉР С”Р С• Р С—Р С•Р Т‘ mutex.
// relatedContainerEquipmentLocked Р Р†Р С•Р В·Р Р†РЎР‚Р В°РЎвЂ°Р В°Р ВµРЎвЂљ РЎРѓР С•РЎвЂ¦РЎР‚Р В°Р Р…РЎвЂР Р…Р Р…РЎвЂ№Р в„– Р С”Р С•Р Р…РЎвЂљР ВµР в„–Р Р…Р ВµРЎР‚ Р Т‘Р В»РЎРЏ РЎС“Р С”Р В°Р В·Р В°Р Р…Р Р…Р С•Р в„– Р С–РЎР‚РЎС“Р С—Р С—РЎвЂ№ Р С‘ Р Р†Р С‘Р Т‘Р В° РЎРѓР Р†РЎРЏР В·Р С‘.
func (world *World) relatedContainerEquipmentLocked(objectID int64, equipmentGroupID int64, relationTypeAcronym string) (*data.EquipmentGroup, error) {
	group, ok := world.data.EquipmentGroups.Get(equipmentGroupID)
	if !ok {
		return nil, errors.New("equipment group not found")
	}
	var relatedEquipmentGroupID int64
	switch relationTypeAcronym {
	case "Source":
		relatedEquipmentGroupID = group.SourceEquipmentGroupID
	case "Destination":
		relatedEquipmentGroupID = group.DestinationEquipmentGroupID
	case "Opposite":
		relatedEquipmentGroupID = group.OppositeEquipmentGroupID
	default:
		return nil, errors.New("unknown equipment group relation")
	}
	if relatedEquipmentGroupID <= 0 {
		return nil, errors.New("equipment group relation not found")
	}
	return world.controlledContainerEquipmentLocked(objectID, relatedEquipmentGroupID)
}

// constructorRelatedContainerOrFallbackLocked возвращает явно выбранный контейнер или сохраненную связь оборудования.
func (world *World) constructorRelatedContainerOrFallbackLocked(objectID int64, constructorID int64, relationTypeAcronym string, fallbackContainerID int64) (*data.EquipmentGroup, error) {
	if fallbackContainerID > 0 {
		return world.controlledContainerEquipmentLocked(objectID, fallbackContainerID)
	}
	return world.relatedContainerEquipmentLocked(objectID, constructorID, relationTypeAcronym)
}

func (world *World) controlledEquipmentitemTypeLocked(objectID int64, groupID int64, itemTypeAcronym string) (*data.EquipmentGroup, error) {
	group, ok := world.data.EquipmentGroups.Get(groupID)
	if !ok {
		return nil, errors.New("equipment group not found")
	}
	if err := world.ensureControlledClusterEquipmentLocked(objectID, group.CosmicObjectID); err != nil {
		return nil, err
	}
	model, ok := world.data.ItemModels.Get(group.EquipmentItemModelID)
	if !ok {
		return nil, errors.New("equipment model not found")
	}
	itemType, ok := world.data.ItemTypes.Get(model.ItemTypeID)
	if !ok || itemType.Acronym != itemTypeAcronym {
		return nil, errors.New("equipment group has unexpected item type")
	}
	return group, nil
}

// itemSchemaLocked РЎР‚Р В°Р В·Р В±Р С‘РЎР‚Р В°Р ВµРЎвЂљ Р С•Р Т‘Р Р…РЎС“ РЎРѓРЎвЂ№РЎР‚РЎС“РЎР‹ Р В·Р В°Р С—Р С‘РЎРѓРЎРЉ РЎРѓРЎвЂ¦Р ВµР СРЎвЂ№ Р С—РЎР‚Р ВµР Т‘Р СР ВµРЎвЂљР В°; Р Р†РЎвЂ№Р В·РЎвЂ№Р Р†Р В°Р ВµРЎвЂљРЎРѓРЎРЏ РЎвЂљР С•Р В»РЎРЉР С”Р С• Р С—Р С•Р Т‘ mutex.
func (world *World) itemSchemaLocked(schemaID int64) (controlPanelItemSchema, error) {
	raw, ok := world.data.Schemas.Items[fmt.Sprintf("%d", schemaID)]
	if !ok {
		return controlPanelItemSchema{}, errors.New("item schema not found")
	}
	var schema controlPanelItemSchema
	if err := json.Unmarshal(raw, &schema); err != nil {
		return controlPanelItemSchema{}, err
	}
	if schema.ID != schemaID || schema.ItemModelID <= 0 {
		return controlPanelItemSchema{}, errors.New("item schema is invalid")
	}
	return schema, nil
}

// itemSchemaComponentsLocked Р Р†Р С•Р В·Р Р†РЎР‚Р В°РЎвЂ°Р В°Р ВµРЎвЂљ Р С”Р С•Р СР С—Р С•Р Р…Р ВµР Р…РЎвЂљРЎвЂ№ Р Р†РЎвЂ№Р В±РЎР‚Р В°Р Р…Р Р…Р С•Р в„– РЎРѓРЎвЂ¦Р ВµР СРЎвЂ№ Р С‘Р В· РЎРѓРЎвЂ№РЎР‚Р С•Р С–Р С• РЎРѓР С—РЎР‚Р В°Р Р†Р С•РЎвЂЎР Р…Р С‘Р С”Р В°; Р Р†РЎвЂ№Р В·РЎвЂ№Р Р†Р В°Р ВµРЎвЂљРЎРѓРЎРЏ РЎвЂљР С•Р В»РЎРЉР С”Р С• Р С—Р С•Р Т‘ mutex.
// itemSchemaByProductModelLocked Р Р…Р В°РЎвЂ¦Р С•Р Т‘Р С‘РЎвЂљ РЎРѓРЎвЂ¦Р ВµР СРЎС“, Р С—РЎР‚Р С•Р С‘Р В·Р Р†Р С•Р Т‘РЎРЏРЎвЂ°РЎС“РЎР‹ РЎС“Р С”Р В°Р В·Р В°Р Р…Р Р…РЎС“РЎР‹ Р СР С•Р Т‘Р ВµР В»РЎРЉ Р С—РЎР‚Р ВµР Т‘Р СР ВµРЎвЂљР В°; Р Р†РЎвЂ№Р В·РЎвЂ№Р Р†Р В°Р ВµРЎвЂљРЎРѓРЎРЏ РЎвЂљР С•Р В»РЎРЉР С”Р С• Р С—Р С•Р Т‘ mutex.
func (world *World) itemSchemaByProductModelLocked(itemModelID int64) (controlPanelItemSchema, error) {
	schemaIDs := make([]int64, 0, len(world.data.Schemas.Items))
	for key := range world.data.Schemas.Items {
		var schemaID int64
		if _, err := fmt.Sscanf(key, "%d", &schemaID); err == nil {
			schemaIDs = append(schemaIDs, schemaID)
		}
	}
	sort.Slice(schemaIDs, func(left int, right int) bool {
		return schemaIDs[left] < schemaIDs[right]
	})
	for _, schemaID := range schemaIDs {
		schema, err := world.itemSchemaLocked(schemaID)
		if err != nil {
			return controlPanelItemSchema{}, err
		}
		if schema.ItemModelID == itemModelID {
			return schema, nil
		}
	}
	return controlPanelItemSchema{}, errors.New("item schema not found")
}

// cheapestItemSchemaByProductModelLocked находит схему с минимальной работой для указанной модели.
func (world *World) cheapestItemSchemaByProductModelLocked(itemModelID int64) (controlPanelItemSchema, error) {
	schemaIDs := make([]int64, 0, len(world.data.Schemas.Items))
	for key := range world.data.Schemas.Items {
		var schemaID int64
		if _, err := fmt.Sscanf(key, "%d", &schemaID); err == nil {
			schemaIDs = append(schemaIDs, schemaID)
		}
	}
	sort.Slice(schemaIDs, func(left int, right int) bool {
		return schemaIDs[left] < schemaIDs[right]
	})
	var best controlPanelItemSchema
	found := false
	for _, schemaID := range schemaIDs {
		schema, err := world.itemSchemaLocked(schemaID)
		if err != nil {
			return controlPanelItemSchema{}, err
		}
		if schema.ItemModelID != itemModelID {
			continue
		}
		if !found || schema.ProductionEnergy < best.ProductionEnergy {
			best = schema
			found = true
		}
	}
	if !found {
		return controlPanelItemSchema{}, errors.New("item schema not found")
	}
	return best, nil
}

func (world *World) itemSchemaComponentsLocked(schemaID int64) ([]controlPanelItemSchemaComponent, error) {
	components := make([]controlPanelItemSchemaComponent, 0)
	for _, raw := range world.data.SchemaComponents.Items {
		var component controlPanelItemSchemaComponent
		if err := json.Unmarshal(raw, &component); err != nil {
			return nil, err
		}
		if component.SchemaID == schemaID {
			components = append(components, component)
		}
	}
	sort.Slice(components, func(left int, right int) bool {
		return components[left].ID < components[right].ID
	})
	return components, nil
}

// consumeItemModelFromContainerLocked РЎРѓР С—Р С‘РЎРѓРЎвЂ№Р Р†Р В°Р ВµРЎвЂљ РЎвЂљРЎР‚Р ВµР В±РЎС“Р ВµР СР С•Р Вµ Р С”Р С•Р В»Р С‘РЎвЂЎР ВµРЎРѓРЎвЂљР Р†Р С• Р С—РЎР‚Р ВµР Т‘Р СР ВµРЎвЂљР С•Р Р† Р С•Р Т‘Р Р…Р С•Р в„– Р СР С•Р Т‘Р ВµР В»Р С‘ Р С‘Р В· Р С”Р С•Р Р…РЎвЂљР ВµР в„–Р Р…Р ВµРЎР‚Р В°; Р Р†РЎвЂ№Р В·РЎвЂ№Р Р†Р В°Р ВµРЎвЂљРЎРѓРЎРЏ РЎвЂљР С•Р В»РЎРЉР С”Р С• Р С—Р С•Р Т‘ mutex.
// objectBlueprintLocked РЎР‚Р В°Р В·Р В±Р С‘РЎР‚Р В°Р ВµРЎвЂљ Р С•Р Т‘Р Р…РЎС“ РЎРѓРЎвЂ№РЎР‚РЎС“РЎР‹ Р В·Р В°Р С—Р С‘РЎРѓРЎРЉ РЎвЂЎР ВµРЎР‚РЎвЂљР ВµР В¶Р В° Р С•Р В±РЎР‰Р ВµР С”РЎвЂљР В°; Р Р†РЎвЂ№Р В·РЎвЂ№Р Р†Р В°Р ВµРЎвЂљРЎРѓРЎРЏ РЎвЂљР С•Р В»РЎРЉР С”Р С• Р С—Р С•Р Т‘ mutex.
func (world *World) objectBlueprintLocked(blueprintID int64) (controlPanelObjectBlueprint, error) {
	raw, ok := world.data.Blueprints.Items[fmt.Sprintf("%d", blueprintID)]
	if !ok {
		return controlPanelObjectBlueprint{}, errors.New("object blueprint not found")
	}
	var blueprint controlPanelObjectBlueprint
	if err := json.Unmarshal(raw, &blueprint); err != nil {
		return controlPanelObjectBlueprint{}, err
	}
	if blueprint.ID != blueprintID || blueprint.CosmicObjectModelID <= 0 {
		return controlPanelObjectBlueprint{}, errors.New("object blueprint is invalid")
	}
	return blueprint, nil
}

// objectBlueprintComponentsLocked Р Р†Р С•Р В·Р Р†РЎР‚Р В°РЎвЂ°Р В°Р ВµРЎвЂљ Р С”Р С•Р СР С—Р С•Р Р…Р ВµР Р…РЎвЂљРЎвЂ№ Р Р†РЎвЂ№Р В±РЎР‚Р В°Р Р…Р Р…Р С•Р С–Р С• РЎвЂЎР ВµРЎР‚РЎвЂљР ВµР В¶Р В°; Р Р†РЎвЂ№Р В·РЎвЂ№Р Р†Р В°Р ВµРЎвЂљРЎРѓРЎРЏ РЎвЂљР С•Р В»РЎРЉР С”Р С• Р С—Р С•Р Т‘ mutex.
func (world *World) objectBlueprintComponentsLocked(blueprintID int64) ([]controlPanelItemSchemaComponent, error) {
	components := make([]controlPanelItemSchemaComponent, 0)
	for _, raw := range world.data.BlueprintComponents.Items {
		var component controlPanelObjectBlueprintComponent
		if err := json.Unmarshal(raw, &component); err != nil {
			return nil, err
		}
		if component.BlueprintID == blueprintID {
			components = append(components, controlPanelItemSchemaComponent{
				ID:                   component.ID,
				ComponentItemModelID: component.ComponentItemModelID,
				Count:                component.Count,
			})
		}
	}
	sort.Slice(components, func(left int, right int) bool {
		return components[left].ID < components[right].ID
	})
	return components, nil
}

func (world *World) consumeItemModelFromContainerLocked(containerID int64, itemModelID int64, amount float64) {
	remaining := amount
	for _, itemGroup := range world.data.ItemGroups.GetByContainerEquipmentGroupID(containerID) {
		if remaining <= physics.Epsilon {
			return
		}
		if itemGroup.ContentItemModelID != itemModelID {
			continue
		}
		consumed := math.Min(itemGroup.Count, remaining)
		itemGroup.Count -= consumed
		remaining -= consumed
		if itemGroup.Count <= physics.Epsilon {
			world.deleteItemGroupLocked(itemGroup)
		}
	}
}

// addItemModelToContainerLocked Р Т‘Р С•Р В±Р В°Р Р†Р В»РЎРЏР ВµРЎвЂљ Р С—РЎР‚Р С•Р Т‘РЎС“Р С”РЎвЂ Р С‘РЎР‹ Р Р† РЎРѓРЎС“РЎвЂ°Р ВµРЎРѓРЎвЂљР Р†РЎС“РЎР‹РЎвЂ°РЎС“РЎР‹ Р С–РЎР‚РЎС“Р С—Р С—РЎС“ Р С‘Р В»Р С‘ РЎРѓР С•Р В·Р Т‘Р В°Р ВµРЎвЂљ Р Р…Р С•Р Р†РЎС“РЎР‹; Р Р†РЎвЂ№Р В·РЎвЂ№Р Р†Р В°Р ВµРЎвЂљРЎРѓРЎРЏ РЎвЂљР С•Р В»РЎРЉР С”Р С• Р С—Р С•Р Т‘ mutex.
// Удаляет группу предметов и сразу очищает быстрый индекс контейнера.
func (world *World) deleteItemGroupLocked(itemGroup *data.ItemGroup) {
	if itemGroup == nil || world.data.ItemGroups == nil {
		return
	}
	delete(world.data.ItemGroups.Items, itemGroup.ID)
	groups := world.data.ItemGroups.ByContainerEquipmentGroupID[itemGroup.ContainerEquipmentGroupID]
	for index, indexedGroup := range groups {
		if indexedGroup == nil || indexedGroup.ID != itemGroup.ID {
			continue
		}
		groups = append(groups[:index], groups[index+1:]...)
		break
	}
	if len(groups) == 0 {
		delete(world.data.ItemGroups.ByContainerEquipmentGroupID, itemGroup.ContainerEquipmentGroupID)
		return
	}
	world.data.ItemGroups.ByContainerEquipmentGroupID[itemGroup.ContainerEquipmentGroupID] = groups
}

// Добавляет предметы в существующую группу контейнера или создает новую.
func (world *World) addItemModelToContainerLocked(containerID int64, itemModelID int64, amount float64) error {
	for _, itemGroup := range world.data.ItemGroups.GetByContainerEquipmentGroupID(containerID) {
		if itemGroup.ContentItemModelID == itemModelID {
			itemGroup.Count += amount
			return nil
		}
	}
	_, err := world.data.ItemGroups.Add(&data.ItemGroup{
		ContainerEquipmentGroupID: containerID,
		ContentItemModelID:        itemModelID,
		Count:                     amount,
	})
	return err
}

// controlledFuelTankFuelModelIDLocked Р Р†Р С•Р В·Р Р†РЎР‚Р В°РЎвЂ°Р В°Р ВµРЎвЂљ Р СР С•Р Т‘Р ВµР В»РЎРЉ РЎвЂљР С•Р С—Р В»Р С‘Р Р†Р В° Р Т‘Р В»РЎРЏ Р Р†РЎвЂ№Р В±РЎР‚Р В°Р Р…Р Р…Р С•Р С–Р С• Р В±Р В°Р С”Р В° РЎвЂљР ВµР С”РЎС“РЎвЂ°Р ВµР С–Р С• Р С•Р В±РЎР‰Р ВµР С”РЎвЂљР В°.
func (world *World) controlledFuelTankFuelModelIDLocked(objectID int64, groupID int64) (int64, error) {
	group, ok := world.data.EquipmentGroups.Get(groupID)
	if !ok {
		return 0, errors.New("equipment group not found")
	}
	if err := world.ensureControlledClusterEquipmentLocked(objectID, group.CosmicObjectID); err != nil {
		return 0, err
	}
	model, ok := world.data.ItemModels.Get(group.EquipmentItemModelID)
	if !ok {
		return 0, errors.New("fuel tank model not found")
	}
	itemType, ok := world.data.ItemTypes.Get(model.ItemTypeID)
	if !ok || itemType.Acronym != "FuelTank" {
		return 0, errors.New("equipment group is not a fuel tank")
	}
	if model.ConsumingItemModelID <= 0 {
		return 0, errors.New("fuel tank fuel model is not set")
	}
	return model.ConsumingItemModelID, nil
}

// fillFuelFromContainerLocked Р В·Р В°Р В±Р С‘РЎР‚Р В°Р ВµРЎвЂљ Р Р†РЎвЂ№Р В±РЎР‚Р В°Р Р…Р Р…Р С•Р Вµ РЎвЂљР С•Р С—Р В»Р С‘Р Р†Р С• Р С‘Р В· Р С”Р С•Р Р…РЎвЂљР ВµР в„–Р Р…Р ВµРЎР‚Р В° Р Т‘Р С• Р В·Р В°Р С—Р С•Р В»Р Р…Р ВµР Р…Р С‘РЎРЏ Р С•Р В±РЎвЂ°Р ВµР С–Р С• Р В·Р В°Р С—Р В°РЎРѓР В°.
func (world *World) fillFuelFromContainerLocked(cosmicObject *data.CosmicObject, containerID int64, fuelModelID int64, itemGroupIDs []int64, amount float64) error {
	freeFuel := math.Max(0, cosmicObject.MaxFuel-cosmicObject.Fuel)
	if freeFuel <= 0 {
		return nil
	}
	remainingAmount := freeFuel
	if amount > 0 {
		remainingAmount = math.Min(freeFuel, amount)
	}
	for _, itemGroupID := range itemGroupIDs {
		if remainingAmount <= 0 {
			break
		}
		itemGroup, ok := world.data.ItemGroups.Get(itemGroupID)
		if !ok {
			return errors.New("item group not found")
		}
		if itemGroup.ContainerEquipmentGroupID != containerID {
			return errors.New("item group does not belong to source container")
		}
		if itemGroup.ContentItemModelID != fuelModelID {
			return errors.New("item group is not fuel for selected tank")
		}
		moved := math.Min(itemGroup.Count, remainingAmount)
		cosmicObject.Fuel += moved
		remainingAmount -= moved
		itemGroup.Count -= moved
		if itemGroup.Count <= physics.Epsilon {
			world.deleteItemGroupLocked(itemGroup)
		}
	}
	return nil
}

// fuelFillAmountLocked считает количество топлива, которое можно поставить в очередь заправки.
func (world *World) fuelFillAmountLocked(cosmicObject *data.CosmicObject, containerID int64, fuelModelID int64, itemGroupIDs []int64, amount float64) (float64, error) {
	freeFuel := math.Max(0, cosmicObject.MaxFuel-cosmicObject.Fuel)
	if freeFuel <= 0 {
		return 0, nil
	}
	selectedFuel := 0.0
	for _, itemGroupID := range itemGroupIDs {
		itemGroup, ok := world.data.ItemGroups.Get(itemGroupID)
		if !ok {
			return 0, errors.New("item group not found")
		}
		if itemGroup.ContainerEquipmentGroupID != containerID {
			return 0, errors.New("item group does not belong to source container")
		}
		if itemGroup.ContentItemModelID != fuelModelID {
			return 0, errors.New("item group is not fuel for selected tank")
		}
		selectedFuel += itemGroup.Count
	}
	if amount > 0 {
		selectedFuel = math.Min(selectedFuel, amount)
	}
	return math.Min(freeFuel, selectedFuel), nil
}

// fuelingTaskEnergyLocked считает работу заправки или слива топлива по массе и расстоянию.
func (world *World) fuelingTaskEnergyLocked(containerID int64, fuelTankID int64, fuelModelID int64, amount float64) (float64, error) {
	container, ok := world.data.EquipmentGroups.Get(containerID)
	if !ok {
		return 0, errors.New("fueling container not found")
	}
	fuelTank, ok := world.data.EquipmentGroups.Get(fuelTankID)
	if !ok {
		return 0, errors.New("fueling tank not found")
	}
	itemModel, ok := world.data.ItemModels.Get(fuelModelID)
	if !ok {
		return 0, errors.New("fuel item model not found")
	}
	distance, err := world.cargoMovementDistanceLocked(container.CosmicObjectID, fuelTank.CosmicObjectID)
	if err != nil {
		return 0, err
	}
	totalEnergy := itemModel.Mass * amount * distance
	if totalEnergy <= physics.Epsilon {
		totalEnergy = physics.Epsilon
	}
	return totalEnergy, nil
}

// drainFuelToContainerLocked РЎРѓР В»Р С‘Р Р†Р В°Р ВµРЎвЂљ РЎС“Р С”Р В°Р В·Р В°Р Р…Р Р…Р С•Р Вµ РЎвЂљР С•Р С—Р В»Р С‘Р Р†Р С• Р С‘Р В· Р С•Р В±РЎвЂ°Р ВµР С–Р С• Р В·Р В°Р С—Р В°РЎРѓР В° Р Р† Р Р†РЎвЂ№Р В±РЎР‚Р В°Р Р…Р Р…РЎвЂ№Р в„– Р С”Р С•Р Р…РЎвЂљР ВµР в„–Р Р…Р ВµРЎР‚.
func (world *World) drainFuelToContainerLocked(cosmicObject *data.CosmicObject, containerID int64, fuelModelID int64, amount float64) error {
	moved := math.Min(math.Max(0, amount), cosmicObject.Fuel)
	if moved <= 0 {
		return nil
	}
	for _, itemGroup := range world.data.ItemGroups.GetByContainerEquipmentGroupID(containerID) {
		if itemGroup.ContentItemModelID == fuelModelID {
			itemGroup.Count += moved
			cosmicObject.Fuel -= moved
			return nil
		}
	}
	if _, err := world.data.ItemGroups.Add(&data.ItemGroup{ContainerEquipmentGroupID: containerID, ContentItemModelID: fuelModelID, Count: moved}); err != nil {
		return err
	}
	cosmicObject.Fuel -= moved
	return nil
}

// equipmentGroupIsContainerLocked Р С—РЎР‚Р С•Р Р†Р ВµРЎР‚РЎРЏР ВµРЎвЂљ РЎвЂљР С‘Р С— РЎС“РЎРѓРЎвЂљР В°Р Р…Р С•Р Р†Р В»Р ВµР Р…Р Р…Р С•Р С–Р С• Р С—РЎР‚Р ВµР Т‘Р СР ВµРЎвЂљР В°; Р Р†РЎвЂ№Р В·РЎвЂ№Р Р†Р В°Р ВµРЎвЂљРЎРѓРЎРЏ РЎвЂљР С•Р В»РЎРЉР С”Р С• Р С—Р С•Р Т‘ mutex.
func (world *World) equipmentGroupIsContainerLocked(group *data.EquipmentGroup) bool {
	if group == nil || world.data.ItemModels == nil || world.data.ItemTypes == nil {
		return false
	}
	model, ok := world.data.ItemModels.Get(group.EquipmentItemModelID)
	if !ok {
		return false
	}
	itemType, ok := world.data.ItemTypes.Get(model.ItemTypeID)
	return ok && itemType.Acronym == "Container"
}

// Р вЂ™РЎвЂ№Р С—Р С•Р В»Р Р…РЎРЏР ВµРЎвЂљ Р С•Р Т‘Р С‘Р Р… РЎв‚¬Р В°Р С– РЎРѓР С‘Р СРЎС“Р В»РЎРЏРЎвЂ Р С‘Р С‘ Р С‘ Р Р†Р С•Р В·Р Р†РЎР‚Р В°РЎвЂ°Р В°Р ВµРЎвЂљ Р С•Р В±РЎвЂ°Р С‘Р в„– РЎРѓР Р…Р С‘Р СР С•Р С” Р СР С‘РЎР‚Р В°.
func (world *World) Tick(dtSeconds float64) game.Snapshot {
	world.mu.Lock()
	defer world.mu.Unlock()

	world.stepTasksLocked(dtSeconds)
	world.stepExchangeSessionsLocked(dtSeconds)
	world.stepMovableObjects(dtSeconds, world.inputsByObjectID())
	world.resolveAllCollisions()
	world.stepExchangeRequestsLocked(dtSeconds)
	world.stepDockingRequestsLocked(dtSeconds)
	world.stepDockingProcessesLocked(dtSeconds)

	world.tick++
	return world.snapshotLocked(0)
}

// stepDockingRequestsLocked РЎС“Р В±Р С‘РЎР‚Р В°Р ВµРЎвЂљ Р В·Р В°Р С—РЎР‚Р С•РЎРѓРЎвЂ№, Р С—Р С• Р С”Р С•РЎвЂљР С•РЎР‚РЎвЂ№Р С Р С‘РЎРѓРЎвЂљР ВµР С”Р В»Р С• Р Р†РЎР‚Р ВµР СРЎРЏ Р С•Р В¶Р С‘Р Т‘Р В°Р Р…Р С‘РЎРЏ Р С•РЎвЂљР Р†Р ВµРЎвЂљР В°.
func (world *World) stepDockingRequestsLocked(dtSeconds float64) {
	if dtSeconds <= 0 || len(world.dockingRequests) == 0 {
		return
	}
	remaining := world.dockingRequests[:0]
	for _, request := range world.dockingRequests {
		request.RemainingSeconds -= dtSeconds
		if request.RemainingSeconds > physics.Epsilon {
			remaining = append(remaining, request)
			continue
		}
		world.closeDockingRequestWindowLocked(request)
		world.addDockingNotificationLocked([]int64{request.SenderCosmicObjectID, request.ReceiverCosmicObjectID}, "Истекло время ожидания ответа на запрос стыковки")
	}
	world.dockingRequests = remaining
}

// stepDockingProcessesLocked Р В·Р В°Р Р†Р ВµРЎР‚РЎв‚¬Р В°Р ВµРЎвЂљ Р В°Р Р†РЎвЂљР С•Р СР В°РЎвЂљР С‘РЎвЂЎР ВµРЎРѓР С”Р С‘Р Вµ РЎРѓРЎвЂљРЎвЂ№Р С”Р С•Р Р†Р С”Р С‘ Р С—Р С• РЎвЂљР В°Р в„–Р СР ВµРЎР‚РЎС“.
func (world *World) stepDockingProcessesLocked(dtSeconds float64) {
	if dtSeconds <= 0 || len(world.dockingProcesses) == 0 {
		return
	}
	remaining := world.dockingProcesses[:0]
	for _, process := range world.dockingProcesses {
		process.RemainingSeconds -= dtSeconds
		if process.RemainingSeconds > physics.Epsilon {
			remaining = append(remaining, process)
			continue
		}
		world.completeDockingProcessLocked(process)
	}
	world.dockingProcesses = remaining
}

// completeDockingProcessLocked Р В·Р В°Р С—Р С‘РЎРѓРЎвЂ№Р Р†Р В°Р ВµРЎвЂљ РЎРѓР Р†РЎРЏР В·РЎРЉ Р С”Р В»Р В°РЎРѓРЎвЂљР ВµРЎР‚Р В° Р С—Р С•РЎРѓР В»Р Вµ Р В·Р В°Р Р†Р ВµРЎР‚РЎв‚¬Р ВµР Р…Р С‘РЎРЏ РЎвЂљР В°Р в„–Р СР ВµРЎР‚Р В°.
func (world *World) completeDockingProcessLocked(process dockingProcess) {
	sender, senderOK := world.data.CosmicObjects.Get(process.SenderCosmicObjectID)
	receiver, receiverOK := world.data.CosmicObjects.Get(process.ReceiverCosmicObjectID)
	if !senderOK || !receiverOK {
		return
	}
	mainID := receiver.ID
	if receiver.ClusterMainCosmicObjectID == receiver.ID {
		mainID = receiver.ClusterMainCosmicObjectID
	}
	receiver.ClusterMainCosmicObjectID = mainID
	sender.ClusterMainCosmicObjectID = mainID
	for _, cosmicObject := range world.data.CosmicObjects.Items {
		if cosmicObject != nil && cosmicObject.ClusterMainCosmicObjectID == mainID {
			cosmicObject.Anchored = true
		}
	}
	world.addDockingWindowEventsLocked("dockingFinished", sender.ID, receiver.ID, 0)
	world.openExchangeAfterDockingLocked(sender.ID, receiver.ID)
	world.addDockingNotificationLocked(world.clusterObjectIDsLocked(mainID), "Объект пристыкован")
}

// stepConstructorProductionJobsLocked Р С—РЎР‚Р С•Р Т‘Р Р†Р С‘Р С–Р В°Р ВµРЎвЂљ Р С—Р С• Р С•Р Т‘Р Р…Р С•Р СРЎС“ Р В·Р В°Р Т‘Р В°Р Р…Р С‘РЎР‹ Р Р…Р В° Р С”Р В°Р В¶Р Т‘РЎвЂ№Р в„– Р С”Р С•Р Р…РЎРѓРЎвЂљРЎР‚РЎС“Р С”РЎвЂљР С•РЎР‚ Р В·Р В° РЎвЂљР ВµР С”РЎС“РЎвЂ°Р С‘Р в„– РЎв‚¬Р В°Р С– Р СР С‘РЎР‚Р В°.
// stepTasksLocked продвигает по одному заданию каждого контроллера очереди.
func (world *World) stepTasksLocked(dtSeconds float64) {
	if dtSeconds <= 0 || world.data.Tasks == nil || world.data.TaskTypes == nil || world.data.TaskItemGroups == nil {
		return
	}
	controllerIDs := world.controllerIDsWithTasksLocked()
	for _, controllerID := range controllerIDs {
		if world.exchangePausesControllerLocked(controllerID) {
			continue
		}
		task := world.activeTaskLocked(controllerID)
		if task == nil {
			continue
		}
		if !world.taskHasReserveLocked(task.ID) {
			if !world.reserveTaskItemsLocked(task) {
				continue
			}
		}
		workPower := world.taskWorkPowerLocked(task)
		if workPower <= physics.Epsilon {
			continue
		}
		if controller, ok := world.data.EquipmentGroups.Get(task.ControllerEquipmentGroupID); ok {
			controller.Active = true
		}
		task.RemainingEnergy = math.Max(0, task.RemainingEnergy-workPower*dtSeconds)
		if task.RemainingEnergy > physics.Epsilon {
			continue
		}
		if err := world.completeTaskLocked(task); err != nil {
			continue
		}
		world.data.TaskItemGroups.DeleteByTaskID(task.ID)
		world.data.Tasks.Delete(task.ID)
		_ = world.data.ItemGroups.RebuildIndexes()
	}
}

// controllerIDsWithTasksLocked собирает контроллеры, у которых есть сохраненные задания.
func (world *World) controllerIDsWithTasksLocked() []int64 {
	seen := map[int64]bool{}
	controllerIDs := make([]int64, 0)
	for _, task := range world.data.Tasks.Items {
		if task == nil || seen[task.ControllerEquipmentGroupID] {
			continue
		}
		seen[task.ControllerEquipmentGroupID] = true
		controllerIDs = append(controllerIDs, task.ControllerEquipmentGroupID)
	}
	sort.Slice(controllerIDs, func(left int, right int) bool { return controllerIDs[left] < controllerIDs[right] })
	return controllerIDs
}

// activeTaskLocked выбирает выполняемую или ближайшую вспомогательную строку очереди.
func (world *World) activeTaskLocked(controllerID int64) *data.Task {
	tasks := append([]*data.Task(nil), world.data.Tasks.GetByControllerEquipmentGroupID(controllerID)...)
	sort.Slice(tasks, func(left int, right int) bool { return tasks[left].ID < tasks[right].ID })
	for _, task := range tasks {
		if task.RemainingEnergy < task.TotalEnergy || world.taskHasReserveLocked(task.ID) {
			return task
		}
	}
	for _, task := range tasks {
		if task.ParentTaskID > 0 {
			return task
		}
	}
	if len(tasks) == 0 {
		return nil
	}
	return tasks[0]
}

// taskHasReserveLocked проверяет, что расходники уже вынесены в отдельное хранилище задания.
func (world *World) taskHasReserveLocked(taskID int64) bool {
	task, ok := world.data.Tasks.Get(taskID)
	if ok && (world.taskTypeAcronymLocked(task) == "CargoMovement" || world.taskTypeAcronymLocked(task) == "Fueling" || world.taskTypeAcronymLocked(task) == "ItemDeconstruction") {
		return taskItemGroupsAreStored(world.data.TaskItemGroups.GetByTaskID(taskID))
	}
	return len(world.data.TaskItemGroups.GetByTaskID(taskID)) > 0
}

// taskItemGroupsAreStored проверяет, что все предметы задания уже лежат во временном хранилище.
func taskItemGroupsAreStored(groups []*data.TaskItemGroup) bool {
	if len(groups) == 0 {
		return false
	}
	for _, group := range groups {
		if group == nil || !group.IsStored {
			return false
		}
	}
	return true
}

// reserveTaskItemsLocked переносит расходники из контейнера в резерв задания.
func (world *World) reserveTaskItemsLocked(task *data.Task) bool {
	if world.taskTypeAcronymLocked(task) == "CargoMovement" {
		return world.reserveCargoMovementItemsLocked(task)
	}
	if world.taskTypeAcronymLocked(task) == "Fueling" {
		return world.reserveFuelingItemsLocked(task)
	}
	if world.taskTypeAcronymLocked(task) == "ItemDeconstruction" {
		return world.reserveItemDeconstructionItemsLocked(task)
	}
	requiredByModel, ok := world.taskRequirementsLocked(task)
	if !ok {
		return false
	}
	if len(requiredByModel) == 0 {
		return true
	}
	materialContainer, err := world.taskRelatedContainerLocked(task, "Source")
	if err != nil {
		return false
	}
	availableByModel := map[int64]float64{}
	for _, itemGroup := range world.data.ItemGroups.GetByContainerEquipmentGroupID(materialContainer.ID) {
		availableByModel[itemGroup.ContentItemModelID] += itemGroup.Count
	}
	for itemModelID, required := range requiredByModel {
		if availableByModel[itemModelID]+physics.Epsilon < required {
			return false
		}
	}
	for itemModelID, required := range requiredByModel {
		world.consumeItemModelFromContainerLocked(materialContainer.ID, itemModelID, required)
		if _, err := world.data.TaskItemGroups.Add(&data.TaskItemGroup{TaskID: task.ID, ItemModelID: itemModelID, Count: required}); err != nil {
			return false
		}
	}
	_ = world.data.ItemGroups.RebuildIndexes()
	return true
}

// reserveFuelingItemsLocked переносит топливо во временное хранилище задания заправки или слива.
// reserveItemDeconstructionItemsLocked переносит разбираемые предметы во временное хранилище задания.
func (world *World) reserveItemDeconstructionItemsLocked(task *data.Task) bool {
	requiredGroups := world.data.TaskItemGroups.GetByTaskID(task.ID)
	if taskItemGroupsAreStored(requiredGroups) {
		return true
	}
	if len(requiredGroups) == 0 || task.SourceContainerEquipmentGroupID <= 0 {
		return false
	}
	availableByModel := map[int64]float64{}
	for _, itemGroup := range world.data.ItemGroups.GetByContainerEquipmentGroupID(task.SourceContainerEquipmentGroupID) {
		availableByModel[itemGroup.ContentItemModelID] += itemGroup.Count
	}
	for _, group := range requiredGroups {
		if group != nil && group.IsStored {
			continue
		}
		if group == nil || availableByModel[group.ItemModelID]+physics.Epsilon < group.Count {
			return false
		}
	}
	for _, group := range requiredGroups {
		if group == nil || group.IsStored {
			continue
		}
		world.consumeItemModelFromContainerLocked(task.SourceContainerEquipmentGroupID, group.ItemModelID, group.Count)
		group.IsStored = true
	}
	_ = world.data.ItemGroups.RebuildIndexes()
	return true
}

func (world *World) reserveFuelingItemsLocked(task *data.Task) bool {
	requiredGroups := world.data.TaskItemGroups.GetByTaskID(task.ID)
	if taskItemGroupsAreStored(requiredGroups) {
		return true
	}
	if len(requiredGroups) == 0 || task.SourceContainerEquipmentGroupID <= 0 || task.FuelTankEquipmentGroupID <= 0 {
		return false
	}
	fuelTank, ok := world.data.EquipmentGroups.Get(task.FuelTankEquipmentGroupID)
	if !ok {
		return false
	}
	cosmicObject, ok := world.data.CosmicObjects.Get(fuelTank.CosmicObjectID)
	if !ok {
		return false
	}
	for _, group := range requiredGroups {
		if group == nil || group.IsStored {
			continue
		}
		if task.LeftToRightDirection {
			freeFuel := math.Max(0, cosmicObject.MaxFuel-cosmicObject.Fuel)
			if freeFuel+physics.Epsilon < group.Count {
				return false
			}
			available := 0.0
			for _, itemGroup := range world.data.ItemGroups.GetByContainerEquipmentGroupID(task.SourceContainerEquipmentGroupID) {
				if itemGroup.ContentItemModelID == group.ItemModelID {
					available += itemGroup.Count
				}
			}
			if available+physics.Epsilon < group.Count {
				return false
			}
			world.consumeItemModelFromContainerLocked(task.SourceContainerEquipmentGroupID, group.ItemModelID, group.Count)
		} else {
			if cosmicObject.Fuel+physics.Epsilon < group.Count {
				return false
			}
			cosmicObject.Fuel = math.Max(0, cosmicObject.Fuel-group.Count)
		}
		group.IsStored = true
	}
	_ = world.data.ItemGroups.RebuildIndexes()
	return true
}

// reserveCargoMovementItemsLocked переносит груз из текущего источника во временное хранилище начавшейся задачи.
func (world *World) reserveCargoMovementItemsLocked(task *data.Task) bool {
	requiredGroups := world.data.TaskItemGroups.GetByTaskID(task.ID)
	if taskItemGroupsAreStored(requiredGroups) {
		return true
	}
	controller, ok := world.data.EquipmentGroups.Get(task.ControllerEquipmentGroupID)
	if !ok {
		return false
	}
	source, target, err := world.cargoMovementEndpointsLocked(controller, task.LeftToRightDirection)
	if err != nil {
		return false
	}
	if len(requiredGroups) == 0 {
		return false
	}
	availableByModel := map[int64]float64{}
	for _, itemGroup := range world.data.ItemGroups.GetByContainerEquipmentGroupID(source.ID) {
		availableByModel[itemGroup.ContentItemModelID] += itemGroup.Count
	}
	for _, group := range requiredGroups {
		if group != nil && group.IsStored {
			continue
		}
		if group == nil || availableByModel[group.ItemModelID]+physics.Epsilon < group.Count {
			return false
		}
	}
	for _, group := range requiredGroups {
		if group == nil || group.IsStored {
			continue
		}
		world.consumeItemModelFromContainerLocked(source.ID, group.ItemModelID, group.Count)
		group.IsStored = true
	}
	totalEnergy, err := world.cargoMovementTaskEnergyLocked(source.ID, target.ID, requiredGroups)
	if err != nil {
		return false
	}
	task.SourceContainerEquipmentGroupID = source.ID
	task.TargetContainerEquipmentGroupID = target.ID
	task.TotalEnergy = totalEnergy
	task.RemainingEnergy = totalEnergy
	_ = world.data.ItemGroups.RebuildIndexes()
	return true
}

// taskCanReserveItemsLocked проверяет, может ли задание начать работу без изменения хранилищ.
func (world *World) taskCanReserveItemsLocked(task *data.Task) bool {
	if world.taskTypeAcronymLocked(task) == "CargoMovement" {
		if world.taskHasReserveLocked(task.ID) {
			return true
		}
		controller, ok := world.data.EquipmentGroups.Get(task.ControllerEquipmentGroupID)
		if !ok {
			return false
		}
		source, _, err := world.cargoMovementEndpointsLocked(controller, task.LeftToRightDirection)
		if err != nil {
			return false
		}
		availableByModel := map[int64]float64{}
		for _, itemGroup := range world.data.ItemGroups.GetByContainerEquipmentGroupID(source.ID) {
			availableByModel[itemGroup.ContentItemModelID] += itemGroup.Count
		}
		for _, group := range world.data.TaskItemGroups.GetByTaskID(task.ID) {
			if group != nil && group.IsStored {
				continue
			}
			if group == nil || availableByModel[group.ItemModelID]+physics.Epsilon < group.Count {
				return false
			}
		}
		return true
	}
	if world.taskTypeAcronymLocked(task) == "Fueling" {
		if world.taskHasReserveLocked(task.ID) {
			return true
		}
		if task.SourceContainerEquipmentGroupID <= 0 || task.FuelTankEquipmentGroupID <= 0 {
			return false
		}
		fuelTank, ok := world.data.EquipmentGroups.Get(task.FuelTankEquipmentGroupID)
		if !ok {
			return false
		}
		cosmicObject, ok := world.data.CosmicObjects.Get(fuelTank.CosmicObjectID)
		if !ok {
			return false
		}
		for _, group := range world.data.TaskItemGroups.GetByTaskID(task.ID) {
			if group != nil && group.IsStored {
				continue
			}
			if group == nil {
				return false
			}
			if task.LeftToRightDirection {
				if math.Max(0, cosmicObject.MaxFuel-cosmicObject.Fuel)+physics.Epsilon < group.Count {
					return false
				}
				available := 0.0
				for _, itemGroup := range world.data.ItemGroups.GetByContainerEquipmentGroupID(task.SourceContainerEquipmentGroupID) {
					if itemGroup.ContentItemModelID == group.ItemModelID {
						available += itemGroup.Count
					}
				}
				if available+physics.Epsilon < group.Count {
					return false
				}
			} else if cosmicObject.Fuel+physics.Epsilon < group.Count {
				return false
			}
		}
		return true
	}
	requiredByModel, ok := world.taskRequirementsLocked(task)
	if !ok {
		return false
	}
	if len(requiredByModel) == 0 {
		return true
	}
	materialContainer, err := world.taskRelatedContainerLocked(task, "Source")
	if err != nil {
		return false
	}
	availableByModel := map[int64]float64{}
	for _, itemGroup := range world.data.ItemGroups.GetByContainerEquipmentGroupID(materialContainer.ID) {
		availableByModel[itemGroup.ContentItemModelID] += itemGroup.Count
	}
	for itemModelID, required := range requiredByModel {
		if availableByModel[itemModelID]+physics.Epsilon < required {
			return false
		}
	}
	return true
}

// taskRequirementsLocked собирает расходники задания по схеме или чертежу.
func (world *World) taskRequirementsLocked(task *data.Task) (map[int64]float64, bool) {
	components, err := world.taskComponentsLocked(task)
	if err != nil {
		return nil, false
	}
	requiredByModel := map[int64]float64{}
	amount := taskCount(task)
	for _, component := range components {
		requiredByModel[component.ComponentItemModelID] += component.Count * amount
	}
	return requiredByModel, true
}

// taskCount возвращает количество партий задания с учетом сохранений без этого поля.
func taskCount(task *data.Task) float64 {
	if task == nil || task.BatchCount <= 0 {
		return 1
	}
	return float64(task.BatchCount)
}

// taskComponentsLocked возвращает список расходников задания.
func (world *World) taskComponentsLocked(task *data.Task) ([]controlPanelItemSchemaComponent, error) {
	if world.taskTypeAcronymLocked(task) == "CargoMovement" || world.taskTypeAcronymLocked(task) == "Fueling" {
		return nil, nil
	}
	if task.BlueprintID > 0 {
		return world.objectBlueprintComponentsLocked(task.BlueprintID)
	}
	return world.itemSchemaComponentsLocked(task.SchemaID)
}

// taskTypeAcronymLocked возвращает неизменяемое имя типа задания.
func (world *World) taskTypeAcronymLocked(task *data.Task) string {
	if task == nil || world.data.TaskTypes == nil {
		return ""
	}
	taskType, ok := world.data.TaskTypes.Get(task.TaskTypeID)
	if !ok {
		return ""
	}
	return taskType.Acronym
}

// taskRelatedContainerLocked находит связанный контейнер для задания.
func (world *World) taskRelatedContainerLocked(task *data.Task, relationTypeAcronym string) (*data.EquipmentGroup, error) {
	controller, ok := world.data.EquipmentGroups.Get(task.ControllerEquipmentGroupID)
	if !ok {
		return nil, errors.New("task controller equipment group not found")
	}
	return world.constructorRelatedContainerOrFallbackLocked(controller.CosmicObjectID, task.ControllerEquipmentGroupID, relationTypeAcronym, 0)
}

// taskWorkPowerLocked вычисляет рабочую мощность по исполнителям типа задания.
func (world *World) taskWorkPowerLocked(task *data.Task) float64 {
	if world.data.Implementers == nil || world.data.ItemTypes == nil || world.data.ItemModels == nil || world.data.EquipmentGroups == nil {
		return 0
	}
	controller, ok := world.data.EquipmentGroups.Get(task.ControllerEquipmentGroupID)
	if !ok {
		return 0
	}
	power := 0.0
	for _, implementer := range world.data.Implementers.GetByTaskTypeID(task.TaskTypeID) {
		for _, group := range world.data.EquipmentGroups.GetByCosmicObjectID(controller.CosmicObjectID) {
			model, ok := world.data.ItemModels.Get(group.EquipmentItemModelID)
			if !ok || model.ItemTypeID != implementer.ImplementerEquipmentItemTypeID {
				continue
			}
			efficiency := model.Efficiency
			if efficiency <= 0 {
				efficiency = 1
			}
			modelPower := model.ConsumingPower
			if modelPower <= 0 {
				modelPower = 1
			}
			power += modelPower * float64(enabledEquipmentCount(group)) * implementer.WorkPart * efficiency
		}
	}
	return power
}

// completeTaskLocked кладет результат задания в игровой мир.
func (world *World) completeTaskLocked(task *data.Task) error {
	if world.taskTypeAcronymLocked(task) == "CargoMovement" {
		targetContainerID := task.TargetContainerEquipmentGroupID
		if targetContainerID <= 0 {
			targetContainerID = task.ControllerEquipmentGroupID
		}
		for _, group := range world.data.TaskItemGroups.GetByTaskID(task.ID) {
			if group == nil || !group.IsStored {
				continue
			}
			if err := world.addItemModelToContainerLocked(targetContainerID, group.ItemModelID, group.Count); err != nil {
				return err
			}
		}
		return nil
	}
	if world.taskTypeAcronymLocked(task) == "Fueling" {
		return world.completeFuelingTaskLocked(task)
	}
	if world.taskTypeAcronymLocked(task) == "ItemDeconstruction" {
		return world.completeItemDeconstructionTaskLocked(task)
	}
	job := constructorProductionJob{ConstructorEquipmentGroupID: task.ControllerEquipmentGroupID, SchemaID: task.SchemaID, BlueprintID: task.BlueprintID}
	if task.BlueprintID > 0 {
		blueprint, err := world.objectBlueprintLocked(task.BlueprintID)
		if err != nil {
			return err
		}
		job.ProductCosmicObjectModelID = blueprint.CosmicObjectModelID
		amount := int(math.Ceil(taskCount(task) - physics.Epsilon))
		if amount < 1 {
			amount = 1
		}
		for index := 0; index < amount; index++ {
			if err := world.createConstructedCosmicObjectLocked(&job); err != nil {
				return err
			}
		}
		return nil
	}
	schema, err := world.itemSchemaLocked(task.SchemaID)
	if err != nil {
		return err
	}
	job.ProductItemModelID = schema.ItemModelID
	job.ProductCount = schema.Count * taskCount(task)
	relationTypeAcronym := "Destination"
	if task.ParentTaskID > 0 {
		relationTypeAcronym = "Source"
	}
	productContainer, err := world.taskRelatedContainerLocked(task, relationTypeAcronym)
	if err != nil {
		return err
	}
	return world.addItemModelToContainerLocked(productContainer.ID, job.ProductItemModelID, job.ProductCount)
}

// completeFuelingTaskLocked применяет результат завершенной заправки или слива.
// completeItemDeconstructionTaskLocked кладет компоненты разобранной партии в контейнер результата.
func (world *World) completeItemDeconstructionTaskLocked(task *data.Task) error {
	if task == nil || task.SchemaID <= 0 || task.TargetContainerEquipmentGroupID <= 0 {
		return errors.New("item deconstruction task is invalid")
	}
	components, err := world.itemSchemaComponentsLocked(task.SchemaID)
	if err != nil {
		return err
	}
	batches := taskCount(task)
	for _, component := range components {
		if component.Count <= 0 {
			continue
		}
		if err := world.addItemModelToContainerLocked(task.TargetContainerEquipmentGroupID, component.ComponentItemModelID, component.Count*batches); err != nil {
			return err
		}
	}
	return nil
}

func (world *World) completeFuelingTaskLocked(task *data.Task) error {
	if task == nil || task.FuelTankEquipmentGroupID <= 0 || task.SourceContainerEquipmentGroupID <= 0 {
		return errors.New("fueling task is invalid")
	}
	fuelTank, ok := world.data.EquipmentGroups.Get(task.FuelTankEquipmentGroupID)
	if !ok {
		return errors.New("fuel tank equipment group not found")
	}
	cosmicObject, ok := world.data.CosmicObjects.Get(fuelTank.CosmicObjectID)
	if !ok {
		return errors.New("fuel tank object not found")
	}
	for _, group := range world.data.TaskItemGroups.GetByTaskID(task.ID) {
		if group == nil || !group.IsStored {
			continue
		}
		if task.LeftToRightDirection {
			cosmicObject.Fuel = math.Min(cosmicObject.MaxFuel, cosmicObject.Fuel+group.Count)
			continue
		}
		if err := world.addItemModelToContainerLocked(task.SourceContainerEquipmentGroupID, group.ItemModelID, group.Count); err != nil {
			return err
		}
	}
	return nil
}
func (world *World) stepConstructorProductionJobsLocked(dtSeconds float64) {
	if dtSeconds <= 0 || len(world.constructorProductionJobs) == 0 {
		return
	}
	constructorIDs := world.constructorIDsWithProductionJobsLocked()
	for _, constructorID := range constructorIDs {
		jobIndex := world.activeConstructorProductionJobIndexLocked(constructorID)
		if jobIndex < 0 {
			continue
		}
		job := &world.constructorProductionJobs[jobIndex]
		if !job.Running {
			if !world.startConstructorProductionJobLocked(job) {
				continue
			}
		}
		job.RemainingTime = math.Max(0, job.RemainingTime-dtSeconds)
		if job.RemainingTime > physics.Epsilon {
			continue
		}
		if err := world.completeConstructorProductionJobLocked(job); err != nil {
			continue
		}
		job.RemainingBatches--
		if job.RemainingBatches > 0 {
			job.Running = false
			job.RemainingTime = job.TotalTime
			_ = world.data.ItemGroups.RebuildIndexes()
			continue
		}
		world.constructorProductionJobs = append(world.constructorProductionJobs[:jobIndex], world.constructorProductionJobs[jobIndex+1:]...)
		_ = world.data.ItemGroups.RebuildIndexes()
	}
}

// completeConstructorProductionJobLocked Р С”Р В»Р В°Р Т‘РЎвЂРЎвЂљ РЎР‚Р ВµР В·РЎС“Р В»РЎРЉРЎвЂљР В°РЎвЂљ Р В·Р В°Р Т‘Р В°Р Р…Р С‘РЎРЏ Р Р† Р С”Р С•Р Р…РЎвЂљР ВµР в„–Р Р…Р ВµРЎР‚ Р С‘Р В»Р С‘ РЎРѓР С•Р В·Р Т‘Р В°РЎвЂРЎвЂљ Р С•Р В±РЎР‰Р ВµР С”РЎвЂљ Р Р† Р С”Р С•РЎРѓР СР С•РЎРѓР Вµ.
func (world *World) completeConstructorProductionJobLocked(job *constructorProductionJob) error {
	if job.ProductCosmicObjectModelID > 0 {
		return world.createConstructedCosmicObjectLocked(job)
	}
	productContainer, err := world.currentConstructorJobContainerLocked(job, "Destination", job.ProductContainerEquipmentGroupID)
	if err != nil {
		return err
	}
	return world.addItemModelToContainerLocked(productContainer.ID, job.ProductItemModelID, job.ProductCount)
}

// constructorIDsWithProductionJobsLocked Р Р†Р С•Р В·Р Р†РЎР‚Р В°РЎвЂ°Р В°Р ВµРЎвЂљ Р С”Р С•Р Р…РЎРѓРЎвЂљРЎР‚РЎС“Р С”РЎвЂљР С•РЎР‚РЎвЂ№ РЎРѓ Р С•РЎвЂЎР ВµРЎР‚Р ВµР Т‘РЎРЏР СР С‘ Р Р† РЎРѓРЎвЂљР В°Р В±Р С‘Р В»РЎРЉР Р…Р С•Р С Р С—Р С•РЎР‚РЎРЏР Т‘Р С”Р Вµ.
// createConstructedCosmicObjectLocked РЎРѓР С•Р В·Р Т‘Р В°РЎвЂРЎвЂљ РЎР‚Р ВµР В·РЎС“Р В»РЎРЉРЎвЂљР В°РЎвЂљ РЎвЂЎР ВµРЎР‚РЎвЂљР ВµР В¶Р В° Р С—Р ВµРЎР‚Р ВµР Т‘ Р С•Р В±РЎР‰Р ВµР С”РЎвЂљР С•Р С-Р С‘Р В·Р С–Р С•РЎвЂљР С•Р Р†Р С‘РЎвЂљР ВµР В»Р ВµР С.
func (world *World) createConstructedCosmicObjectLocked(job *constructorProductionJob) error {
	constructor, ok := world.data.EquipmentGroups.Get(job.ConstructorEquipmentGroupID)
	if !ok {
		return errors.New("constructor equipment group not found")
	}
	builder, ok := world.data.CosmicObjects.Get(constructor.CosmicObjectID)
	if !ok {
		return errors.New("builder object not found")
	}
	model, ok := world.data.CosmicObjectModels.Get(job.ProductCosmicObjectModelID)
	if !ok {
		return errors.New("blueprint object model not found")
	}
	assembly, ok := world.firstPublicDeveloperAssembly(model.ID)
	if !ok {
		return errors.New("blueprint object assembly not found")
	}
	cosmicObject := world.cosmicObjectFromModelAndAssembly(model, assembly)
	cosmicObject.OwnerCharacterID = builder.OwnerCharacterID
	cosmicObject.CreatorCharacterID = builder.OwnerCharacterID
	cosmicObject.Rotation = builder.Rotation
	cosmicObject.TargetRotation = builder.Rotation
	world.placeConstructedCosmicObjectLocked(cosmicObject, *model, *builder)
	createdObject, err := world.data.CosmicObjects.Add(cosmicObject)
	if err != nil {
		return err
	}
	return world.ensureEquipmentFromAssembly(createdObject.ID, assembly)
}

// placeConstructedCosmicObjectLocked Р С‘РЎвЂ°Р ВµРЎвЂљ РЎРѓР Р†Р С•Р В±Р С•Р Т‘Р Р…РЎС“РЎР‹ РЎвЂљР С•РЎвЂЎР С”РЎС“ Р С—РЎР‚РЎРЏР СР С• Р С—Р С• РЎвЂ¦Р С•Р Т‘РЎС“ Р С•Р В±РЎР‰Р ВµР С”РЎвЂљР В°-Р С‘Р В·Р С–Р С•РЎвЂљР С•Р Р†Р С‘РЎвЂљР ВµР В»РЎРЏ.
func (world *World) placeConstructedCosmicObjectLocked(created *data.CosmicObject, createdModel data.CosmicObjectModel, builder data.CosmicObject) {
	builderModel, ok := world.data.CosmicObjectModels.Get(builder.CosmicObjectModelID)
	if !ok {
		created.X = builder.X
		created.Y = builder.Y
		return
	}
	forward := physics.ForwardVector(builder.Rotation)
	gap := 1.0
	baseDistance := builderModel.BodyLength/2 + createdModel.BodyLength/2 + gap
	stepDistance := math.Max(1, createdModel.BodyLength/2)
	for index := 0; index < 1000; index++ {
		distance := baseDistance + float64(index)*stepDistance
		created.X = builder.X + forward.X*distance
		created.Y = builder.Y + forward.Y*distance
		if !world.cosmicObjectIntersectsExistingLocked(*created, createdModel) {
			return
		}
	}
}

// cosmicObjectIntersectsExistingLocked Р С—РЎР‚Р С•Р Р†Р ВµРЎР‚РЎРЏР ВµРЎвЂљ Р С—Р ВµРЎР‚Р ВµРЎРѓР ВµРЎвЂЎР ВµР Р…Р С‘Р Вµ Р С”Р В°Р Р…Р Т‘Р С‘Р Т‘Р В°РЎвЂљР В° РЎРѓ РЎС“Р В¶Р Вµ РЎРѓРЎС“РЎвЂ°Р ВµРЎРѓРЎвЂљР Р†РЎС“РЎР‹РЎвЂ°Р С‘Р СР С‘ Р С•Р В±РЎР‰Р ВµР С”РЎвЂљР В°Р СР С‘.
func (world *World) cosmicObjectIntersectsExistingLocked(candidate data.CosmicObject, candidateModel data.CosmicObjectModel) bool {
	for _, existing := range world.data.CosmicObjects.Items {
		if existing == nil {
			continue
		}
		existingModel, ok := world.data.CosmicObjectModels.Get(existing.CosmicObjectModelID)
		if !ok {
			continue
		}
		if _, collided := physics.CollisionInfo(candidate, candidateModel, *existing, *existingModel); collided {
			return true
		}
	}
	return false
}

func (world *World) constructorIDsWithProductionJobsLocked() []int64 {
	seen := map[int64]bool{}
	constructorIDs := make([]int64, 0)
	for _, job := range world.constructorProductionJobs {
		if seen[job.ConstructorEquipmentGroupID] {
			continue
		}
		seen[job.ConstructorEquipmentGroupID] = true
		constructorIDs = append(constructorIDs, job.ConstructorEquipmentGroupID)
	}
	sort.Slice(constructorIDs, func(left int, right int) bool {
		return constructorIDs[left] < constructorIDs[right]
	})
	return constructorIDs
}

// constructorMainJobIndexLocked Р С‘РЎвЂ°Р ВµРЎвЂљ Р Р†РЎвЂ№Р В±РЎР‚Р В°Р Р…Р Р…РЎС“РЎР‹ РЎРѓРЎвЂљРЎР‚Р С•Р С”РЎС“ Р С•РЎРѓР Р…Р С•Р Р†Р Р…Р С•Р в„– Р С•РЎвЂЎР ВµРЎР‚Р ВµР Т‘Р С‘ Р С”Р С•Р Р…РЎРѓРЎвЂљРЎР‚РЎС“Р С”РЎвЂљР С•РЎР‚Р В°.
func (world *World) constructorMainJobIndexLocked(constructorID int64, jobID int64) int {
	for index, job := range world.constructorProductionJobs {
		if job.ID == jobID && job.ConstructorEquipmentGroupID == constructorID && job.QueueType == "main" {
			return index
		}
	}
	return -1
}

// skipConstructorMainJobNextLocked Р С•РЎРѓРЎвЂљР В°Р Р†Р В»РЎРЏР ВµРЎвЂљ РЎвЂљР С•Р В»РЎРЉР С”Р С• Р Р…Р В°РЎвЂЎР В°РЎвЂљРЎС“РЎР‹ Р ВµР Т‘Р С‘Р Р…Р С‘РЎвЂ РЎС“ Р С‘Р В»Р С‘ РЎС“Р В±Р С‘РЎР‚Р В°Р ВµРЎвЂљ Р Р…Р Вµ Р Р…Р В°РЎвЂЎР В°РЎвЂљРЎС“РЎР‹ РЎРѓРЎвЂљРЎР‚Р С•Р С”РЎС“.
func (world *World) skipConstructorMainJobNextLocked(jobIndex int) {
	if jobIndex < 0 || jobIndex >= len(world.constructorProductionJobs) {
		return
	}
	job := &world.constructorProductionJobs[jobIndex]
	if !job.Running {
		world.removeConstructorMainJobAtLocked(jobIndex)
		return
	}
	job.RemainingBatches = 1
	job.TotalBatches = 1
}

// removeConstructorMainJobsAfterLocked РЎС“Р В±Р С‘РЎР‚Р В°Р ВµРЎвЂљ Р С•РЎРѓР Р…Р С•Р Р†Р Р…РЎвЂ№Р Вµ РЎРѓРЎвЂљРЎР‚Р С•Р С”Р С‘, РЎРѓР В»Р ВµР Т‘РЎС“РЎР‹РЎвЂ°Р С‘Р Вµ Р В·Р В° Р Р†РЎвЂ№Р В±РЎР‚Р В°Р Р…Р Р…Р С•Р в„–.
func (world *World) removeConstructorMainJobsAfterLocked(constructorID int64, jobID int64) {
	seenSelected := false
	for index := 0; index < len(world.constructorProductionJobs); {
		job := world.constructorProductionJobs[index]
		if job.ConstructorEquipmentGroupID == constructorID && job.QueueType == "main" {
			if seenSelected {
				world.removeConstructorMainJobAtLocked(index)
				continue
			}
			if job.ID == jobID {
				seenSelected = true
			}
		}
		index++
	}
}

// removeConstructorMainJobsFromLocked РЎС“Р В±Р С‘РЎР‚Р В°Р ВµРЎвЂљ Р Р†РЎвЂ№Р В±РЎР‚Р В°Р Р…Р Р…РЎС“РЎР‹ Р С•РЎРѓР Р…Р С•Р Р†Р Р…РЎС“РЎР‹ РЎРѓРЎвЂљРЎР‚Р С•Р С”РЎС“ Р С‘ Р Р†РЎРѓР Вµ РЎРѓР В»Р ВµР Т‘РЎС“РЎР‹РЎвЂ°Р С‘Р Вµ Р С•РЎРѓР Р…Р С•Р Р†Р Р…РЎвЂ№Р Вµ РЎРѓРЎвЂљРЎР‚Р С•Р С”Р С‘.
func (world *World) removeConstructorMainJobsFromLocked(constructorID int64, jobID int64) {
	seenSelected := false
	for index := 0; index < len(world.constructorProductionJobs); {
		job := world.constructorProductionJobs[index]
		if job.ConstructorEquipmentGroupID == constructorID && job.QueueType == "main" && (seenSelected || job.ID == jobID) {
			seenSelected = true
			world.removeConstructorMainJobAtLocked(index)
			continue
		}
		index++
	}
}

// removeConstructorMainJobAtLocked РЎС“Р В±Р С‘РЎР‚Р В°Р ВµРЎвЂљ Р С•РЎРѓР Р…Р С•Р Р†Р Р…РЎС“РЎР‹ РЎРѓРЎвЂљРЎР‚Р С•Р С”РЎС“ Р С‘ Р ВµР Вµ Р Р†РЎРѓР С—Р С•Р СР С•Р С–Р В°РЎвЂљР ВµР В»РЎРЉР Р…РЎвЂ№Р Вµ РЎРѓРЎвЂљРЎР‚Р С•Р С”Р С‘.
func (world *World) removeConstructorMainJobAtLocked(jobIndex int) {
	if jobIndex < 0 || jobIndex >= len(world.constructorProductionJobs) {
		return
	}
	jobID := world.constructorProductionJobs[jobIndex].ID
	world.constructorProductionJobs = append(world.constructorProductionJobs[:jobIndex], world.constructorProductionJobs[jobIndex+1:]...)
	for index := 0; index < len(world.constructorProductionJobs); {
		if world.constructorProductionJobs[index].ParentJobID == jobID {
			world.constructorProductionJobs = append(world.constructorProductionJobs[:index], world.constructorProductionJobs[index+1:]...)
			continue
		}
		index++
	}
}

// activeConstructorProductionJobIndexLocked Р Р†РЎвЂ№Р В±Р С‘РЎР‚Р В°Р ВµРЎвЂљ РЎвЂљР ВµР С”РЎС“РЎвЂ°Р ВµР Вµ Р В·Р В°Р Т‘Р В°Р Р…Р С‘Р Вµ Р С”Р С•Р Р…РЎРѓРЎвЂљРЎР‚РЎС“Р С”РЎвЂљР С•РЎР‚Р В° РЎРѓ Р С—РЎР‚Р С‘Р С•РЎР‚Р С‘РЎвЂљР ВµРЎвЂљР С•Р С Р Р†РЎРѓР С—Р С•Р СР С•Р С–Р В°РЎвЂљР ВµР В»РЎРЉР Р…Р С•Р в„– Р С•РЎвЂЎР ВµРЎР‚Р ВµР Т‘Р С‘.
func (world *World) activeConstructorProductionJobIndexLocked(constructorID int64) int {
	for index, job := range world.constructorProductionJobs {
		if job.ConstructorEquipmentGroupID == constructorID && job.Running {
			return index
		}
	}
	for index, job := range world.constructorProductionJobs {
		if job.ConstructorEquipmentGroupID == constructorID && job.QueueType == "auxiliary" {
			return index
		}
	}
	for index, job := range world.constructorProductionJobs {
		if job.ConstructorEquipmentGroupID == constructorID && job.QueueType == "main" {
			return index
		}
	}
	return -1
}

// startConstructorProductionJobLocked РЎРѓР С—Р С‘РЎРѓРЎвЂ№Р Р†Р В°Р ВµРЎвЂљ Р С”Р С•Р СР С—Р С•Р Р…Р ВµР Р…РЎвЂљРЎвЂ№ Р В·Р В°Р Т‘Р В°Р Р…Р С‘РЎРЏ Р С‘ Р С—Р ВµРЎР‚Р ВµР Р†Р С•Р Т‘Р С‘РЎвЂљ Р ВµР С–Р С• Р Р† Р Р†РЎвЂ№Р С—Р С•Р В»Р Р…Р ВµР Р…Р С‘Р Вµ.
func (world *World) startConstructorProductionJobLocked(job *constructorProductionJob) bool {
	requiredByModel, ok := world.constructorProductionRequirementsLocked(job)
	if !ok {
		return false
	}
	materialContainer, err := world.currentConstructorJobContainerLocked(job, "Source", job.MaterialContainerEquipmentGroupID)
	if err != nil {
		return false
	}
	for itemModelID, required := range requiredByModel {
		world.consumeItemModelFromContainerLocked(materialContainer.ID, itemModelID, required)
	}
	job.Running = true
	job.RemainingTime = job.TotalTime
	_ = world.data.ItemGroups.RebuildIndexes()
	return true
}

// Р РЋР С•Р В±Р С‘РЎР‚Р В°Р ВµРЎвЂљ Р Р†Р Р†Р С•Р Т‘ Р С—Р С•Р Т‘Р С”Р В»РЎР‹РЎвЂЎР ВµР Р…Р Р…РЎвЂ№РЎвЂ¦ Р В°Р С”Р С”Р В°РЎС“Р Р…РЎвЂљР С•Р Р† Р С—Р С• РЎС“Р С—РЎР‚Р В°Р Р†Р В»РЎРЏР ВµР СРЎвЂ№Р С Р С•Р В±РЎР‰Р ВµР С”РЎвЂљР В°Р С.
// constructorJobComponentsLocked Р Р†Р С•Р В·Р Р†РЎР‚Р В°РЎвЂ°Р В°Р ВµРЎвЂљ Р С”Р С•Р СР С—Р С•Р Р…Р ВµР Р…РЎвЂљРЎвЂ№ Р В·Р В°Р Т‘Р В°Р Р…Р С‘РЎРЏ Р С—Р С• РЎРѓРЎвЂ¦Р ВµР СР Вµ Р С‘Р В»Р С‘ РЎвЂЎР ВµРЎР‚РЎвЂљР ВµР В¶РЎС“.
// constructorEquipmentIsWorkingLocked Р С—РЎР‚Р С•Р Р†Р ВµРЎР‚РЎРЏР ВµРЎвЂљ, Р Р†РЎвЂ№Р С—Р С•Р В»Р Р…РЎРЏР ВµРЎвЂљ Р В»Р С‘ Р С–РЎР‚РЎС“Р С—Р С—Р В° Р С”Р С•Р Р…РЎРѓРЎвЂљРЎР‚РЎС“Р С”РЎвЂљР С•РЎР‚Р В° РЎвЂљР ВµР С”РЎС“РЎвЂ°Р ВµР Вµ Р С‘Р В»Р С‘ Р С–Р С•РЎвЂљР С•Р Р†Р С•Р Вµ Р С” РЎРѓРЎвЂљР В°РЎР‚РЎвЂљРЎС“ Р В·Р В°Р Т‘Р В°Р Р…Р С‘Р Вµ.
func (world *World) constructorEquipmentIsWorkingLocked(groupID int64) bool {
	if world.data.Tasks != nil {
		task := world.activeTaskLocked(groupID)
		if task == nil {
			group, ok := world.data.EquipmentGroups.Get(groupID)
			if ok && world.equipmentGroupHasItemTypeLocked(group, "Constructor") {
				for _, candidate := range world.data.Tasks.Items {
					controller, controllerOk := world.data.EquipmentGroups.Get(candidate.ControllerEquipmentGroupID)
					if candidate != nil && controllerOk && controller.CosmicObjectID == group.CosmicObjectID {
						task = candidate
						break
					}
				}
			}
		}
		if task == nil {
			return false
		}
		return task.RemainingEnergy < task.TotalEnergy || world.taskHasReserveLocked(task.ID) || (world.taskWorkPowerLocked(task) > physics.Epsilon && world.taskCanReserveItemsLocked(task))
	}
	jobIndex := world.activeConstructorProductionJobIndexLocked(groupID)
	if jobIndex < 0 {
		return false
	}
	job := &world.constructorProductionJobs[jobIndex]
	if job.Running {
		return true
	}
	_, ok := world.constructorProductionRequirementsLocked(job)
	return ok
}

// equipmentGroupHasItemTypeLocked проверяет тип установленного предмета оборудования.
func (world *World) equipmentGroupHasItemTypeLocked(group *data.EquipmentGroup, itemTypeAcronym string) bool {
	if group == nil || world.data.ItemModels == nil || world.data.ItemTypes == nil {
		return false
	}
	model, ok := world.data.ItemModels.Get(group.EquipmentItemModelID)
	if !ok {
		return false
	}
	itemType, ok := world.data.ItemTypes.Get(model.ItemTypeID)
	return ok && itemType.Acronym == itemTypeAcronym
}

// constructorProductionRequirementsLocked РЎРѓР С•Р В±Р С‘РЎР‚Р В°Р ВµРЎвЂљ Р Т‘Р С•РЎРѓРЎвЂљРЎС“Р С—Р Р…РЎвЂ№Р Вµ Р С” РЎРѓР С—Р С‘РЎРѓР В°Р Р…Р С‘РЎР‹ Р С”Р С•Р СР С—Р С•Р Р…Р ВµР Р…РЎвЂљРЎвЂ№ Р Т‘Р В»РЎРЏ РЎРѓРЎвЂљР В°РЎР‚РЎвЂљР В° РЎРѓРЎвЂљРЎР‚Р С•Р С”Р С‘ Р С•РЎвЂЎР ВµРЎР‚Р ВµР Т‘Р С‘.
func (world *World) constructorProductionRequirementsLocked(job *constructorProductionJob) (map[int64]float64, bool) {
	if world.data.ItemGroups == nil {
		return nil, false
	}
	components, err := world.constructorJobComponentsLocked(job)
	if err != nil || len(components) == 0 {
		return nil, false
	}
	requiredByModel := map[int64]float64{}
	for _, component := range components {
		requiredByModel[component.ComponentItemModelID] += component.Count
	}
	availableByModel := map[int64]float64{}
	materialContainer, err := world.currentConstructorJobContainerLocked(job, "Source", job.MaterialContainerEquipmentGroupID)
	if err != nil {
		return nil, false
	}
	for _, itemGroup := range world.data.ItemGroups.GetByContainerEquipmentGroupID(materialContainer.ID) {
		availableByModel[itemGroup.ContentItemModelID] += itemGroup.Count
	}
	for itemModelID, required := range requiredByModel {
		if availableByModel[itemModelID]+physics.Epsilon < required {
			return nil, false
		}
	}
	return requiredByModel, true
}

func (world *World) constructorJobComponentsLocked(job *constructorProductionJob) ([]controlPanelItemSchemaComponent, error) {
	if job.BlueprintID > 0 {
		return world.objectBlueprintComponentsLocked(job.BlueprintID)
	}
	return world.itemSchemaComponentsLocked(job.SchemaID)
}

// currentConstructorJobContainerLocked Р Р…Р В°РЎвЂ¦Р С•Р Т‘Р С‘РЎвЂљ Р В°Р С”РЎвЂљРЎС“Р В°Р В»РЎРЉР Р…РЎвЂ№Р в„– Р С”Р С•Р Р…РЎвЂљР ВµР в„–Р Р…Р ВµРЎР‚ Р В·Р В°Р Т‘Р В°Р Р…Р С‘РЎРЏ Р С—Р С• РЎвЂљР ВµР С”РЎС“РЎвЂ°Р С‘Р С РЎРѓР С•РЎвЂ¦РЎР‚Р В°Р Р…РЎвЂР Р…Р Р…РЎвЂ№Р С РЎРѓР Р†РЎРЏР В·РЎРЏР С Р С”Р С•Р Р…РЎРѓРЎвЂљРЎР‚РЎС“Р С”РЎвЂљР С•РЎР‚Р В°.
func (world *World) currentConstructorJobContainerLocked(job *constructorProductionJob, relationTypeAcronym string, fallbackContainerID int64) (*data.EquipmentGroup, error) {
	constructor, ok := world.data.EquipmentGroups.Get(job.ConstructorEquipmentGroupID)
	if !ok {
		return nil, errors.New("constructor equipment group not found")
	}
	return world.constructorRelatedContainerOrFallbackLocked(constructor.CosmicObjectID, job.ConstructorEquipmentGroupID, relationTypeAcronym, fallbackContainerID)
}

func (world *World) inputsByObjectID() map[int64]game.ShipInput {
	accountIDs := make([]int64, 0, len(world.accountObjectIDs))
	for accountID := range world.accountObjectIDs {
		accountIDs = append(accountIDs, accountID)
	}
	sort.Slice(accountIDs, func(left int, right int) bool {
		return accountIDs[left] < accountIDs[right]
	})

	result := make(map[int64]game.ShipInput, len(accountIDs))
	for _, accountID := range accountIDs {
		result[world.accountObjectIDs[accountID]] = world.inputs[accountID]
	}
	return result
}

// Р СџРЎР‚Р С•Р Р†Р ВµРЎР‚РЎРЏР ВµРЎвЂљ, РЎвЂЎРЎвЂљР С• Р С•Р В±РЎР‰Р ВµР С”РЎвЂљ Р Р…Р Вµ Р С‘Р СР ВµР ВµРЎвЂљ Р Р…Р С‘ Р В»Р С‘Р Р…Р ВµР в„–Р Р…Р С•Р С–Р С•, Р Р…Р С‘ РЎС“Р С–Р В»Р С•Р Р†Р С•Р С–Р С• Р Т‘Р Р†Р С‘Р В¶Р ВµР Р…Р С‘РЎРЏ.
func cosmicObjectIsFullyStopped(cosmicObject data.CosmicObject) bool {
	return math.Abs(cosmicObject.VelocityX) <= physics.Epsilon &&
		math.Abs(cosmicObject.VelocityY) <= physics.Epsilon &&
		math.Abs(cosmicObject.Speed) <= physics.Epsilon &&
		math.Abs(cosmicObject.AngularSpeed) <= physics.Epsilon
}

// Р вЂќР Р†Р С‘Р С–Р В°Р ВµРЎвЂљ Р Р†РЎРѓР Вµ Р С—Р С•Р Т‘Р Р†Р С‘Р В¶Р Р…РЎвЂ№Р Вµ Р С•Р В±РЎР‰Р ВµР С”РЎвЂљРЎвЂ№ Р СР С‘РЎР‚Р В° Р Т‘Р С• Р С•Р В±РЎвЂ°Р ВµР С–Р С• РЎР‚Р ВµРЎв‚¬Р ВµР Р…Р С‘РЎРЏ РЎРѓРЎвЂљР С•Р В»Р С”Р Р…Р С•Р Р†Р ВµР Р…Р С‘Р в„–.
func (world *World) stepMovableObjects(dtSeconds float64, inputsByObjectID map[int64]game.ShipInput) {
	objectIDs := make([]int64, 0, len(world.data.CosmicObjects.Items))
	for objectID := range world.data.CosmicObjects.Items {
		objectIDs = append(objectIDs, objectID)
	}
	sort.Slice(objectIDs, func(left int, right int) bool {
		return objectIDs[left] < objectIDs[right]
	})

	for _, objectID := range objectIDs {
		cosmicObject, ok := world.data.CosmicObjects.Get(objectID)
		if !ok {
			continue
		}
		if !cosmicObject.Enabled {
			world.updateEquipmentUsage(cosmicObject, dtSeconds)
			continue
		}
		model, ok := world.data.CosmicObjectModels.Get(cosmicObject.CosmicObjectModelID)
		if !ok {
			world.updateEquipmentUsage(cosmicObject, dtSeconds)
			continue
		}
		if cosmicObject.Anchored {
			world.updateEquipmentUsage(cosmicObject, dtSeconds)
			continue
		}

		input, controlled := inputsByObjectID[objectID]
		isShip := world.isShipModel(model)
		if controlled && (!isShip || shipHasFuel(*cosmicObject)) {
			*cosmicObject = world.stepControlledObject(*cosmicObject, *model, input, dtSeconds)
		} else if isShip {
			*cosmicObject = physics.StepUnpilotedShip(*cosmicObject, dtSeconds)
		} else {
			*cosmicObject = physics.StepFreeBody(*cosmicObject, dtSeconds)
		}
		world.updateEquipmentUsage(cosmicObject, dtSeconds)
	}
}

// Р вЂ™РЎвЂ№Р С—Р С•Р В»Р Р…РЎРЏР ВµРЎвЂљ РЎвЂћР С‘Р В·Р С‘РЎвЂЎР ВµРЎРѓР С”Р С‘Р в„– РЎв‚¬Р В°Р С– РЎРѓ РЎРѓР С‘Р В»Р В°Р СР С‘, Р Т‘Р С•РЎРѓРЎвЂљРЎС“Р С—Р Р…РЎвЂ№Р СР С‘ РЎвЂљР С•Р В»РЎРЉР С”Р С• Р С•РЎвЂљ Р Р†Р С”Р В»РЎР‹РЎвЂЎР ВµР Р…Р Р…Р С•Р С–Р С• Р С•Р В±Р С•РЎР‚РЎС“Р Т‘Р С•Р Р†Р В°Р Р…Р С‘РЎРЏ.
func (world *World) stepControlledObject(cosmicObject data.CosmicObject, model data.CosmicObjectModel, input game.ShipInput, dtSeconds float64) data.CosmicObject {
	effectiveObject := world.objectWithEnabledEquipmentForces(cosmicObject, input)
	next := physics.StepShip(effectiveObject, model, input, dtSeconds)
	next.MaxAlongForce = cosmicObject.MaxAlongForce
	next.MaxAcrossForce = cosmicObject.MaxAcrossForce
	next.MaxTorque = cosmicObject.MaxTorque
	return next
}

// objectWithEnabledEquipmentForces РЎР‚Р В°РЎРѓРЎРѓРЎвЂЎР С‘РЎвЂљРЎвЂ№Р Р†Р В°Р ВµРЎвЂљ Р Т‘Р С•РЎРѓРЎвЂљРЎС“Р С—Р Р…РЎС“РЎР‹ РЎвЂљРЎРЏР С–РЎС“ Р С‘ Р СР С•Р СР ВµР Р…РЎвЂљ Р С—Р С• Р Р†Р С”Р В»РЎР‹РЎвЂЎР ВµР Р…Р Р…РЎвЂ№Р С Р С–РЎР‚РЎС“Р С—Р С—Р В°Р С Р С•Р В±Р С•РЎР‚РЎС“Р Т‘Р С•Р Р†Р В°Р Р…Р С‘РЎРЏ.
func (world *World) objectWithEnabledEquipmentForces(cosmicObject data.CosmicObject, input game.ShipInput) data.CosmicObject {
	if world.data.EquipmentGroups == nil || world.data.ItemModels == nil {
		return cosmicObject
	}

	electricShare := world.electricShareForInput(cosmicObject, input)
	cosmicObject.MaxAlongForce = 0
	cosmicObject.MaxAcrossForce = 0
	cosmicObject.MaxTorque = 0
	for _, group := range sortedEquipmentGroups(world.data.EquipmentGroups.GetByCosmicObjectID(cosmicObject.ID)) {
		enabledCount := enabledEquipmentCount(group)
		if enabledCount <= 0 {
			continue
		}
		model, ok := world.data.ItemModels.Get(group.EquipmentItemModelID)
		if !ok {
			continue
		}
		cosmicObject.MaxAlongForce += model.MaxAlongForce * float64(enabledCount) * electricShare
		cosmicObject.MaxAcrossForce += model.MaxAcrossForce * float64(enabledCount) * electricShare
		cosmicObject.MaxTorque += model.MaxTorque * float64(enabledCount) * electricShare
	}
	return cosmicObject
}

// electricShareForInput Р Р†Р С•Р В·Р Р†РЎР‚Р В°РЎвЂ°Р В°Р ВµРЎвЂљ Р Т‘Р С•Р В»РЎР‹ РЎРЊР В»Р ВµР С”РЎвЂљРЎР‚Р С‘РЎвЂЎР ВµРЎРѓРЎвЂљР Р†Р В°, Р Т‘Р С•РЎРѓРЎвЂљРЎС“Р С—Р Р…РЎС“РЎР‹ Р Р†РЎРѓР ВµР С Р С—Р С•РЎвЂљРЎР‚Р ВµР В±Р С‘РЎвЂљР ВµР В»РЎРЏР С Р С—РЎР‚Р С‘ РЎвЂљР ВµР С”РЎС“РЎвЂ°Р ВµР С Р Р†Р Р†Р С•Р Т‘Р Вµ.
func (world *World) electricShareForInput(cosmicObject data.CosmicObject, input game.ShipInput) float64 {
	if world.data.EquipmentGroups == nil || world.data.ItemModels == nil {
		return 1
	}

	generatedPower := 0.0
	neededPower := 0.0
	for _, group := range sortedEquipmentGroups(world.data.EquipmentGroups.GetByCosmicObjectID(cosmicObject.ID)) {
		enabledCount := enabledEquipmentCount(group)
		if enabledCount <= 0 {
			continue
		}
		model, ok := world.data.ItemModels.Get(group.EquipmentItemModelID)
		if !ok {
			continue
		}
		count := float64(enabledCount)
		generatedPower += model.GeneratingPower * count
		if equipmentNeedsElectricityForInput(input, *model) {
			neededPower += model.ConsumingPower * count
		}
	}

	return electricWorkShare(generatedPower, neededPower)
}

// Р С›Р В±Р Р…Р С•Р Р†Р В»РЎРЏР ВµРЎвЂљ Р В°Р С”РЎвЂљР С‘Р Р†Р Р…Р С•РЎРѓРЎвЂљРЎРЉ Р С•Р В±Р С•РЎР‚РЎС“Р Т‘Р С•Р Р†Р В°Р Р…Р С‘РЎРЏ, Р СР С•РЎвЂ°Р Р…Р С•РЎРѓРЎвЂљРЎРЉ Р С‘ Р В·Р В°Р С—Р В°РЎРѓ РЎвЂљР С•Р С—Р В»Р С‘Р Р†Р В° Р С—Р С•РЎРѓР В»Р Вµ РЎв‚¬Р В°Р С–Р В° Р С•Р В±РЎР‰Р ВµР С”РЎвЂљР В°.
func (world *World) updateEquipmentUsage(cosmicObject *data.CosmicObject, dtSeconds float64) {
	cosmicObject.ConsumingPower = 0
	cosmicObject.GeneratingPower = 0
	if world.data.EquipmentGroups == nil || world.data.ItemModels == nil {
		return
	}

	consumerFuelConsumptionPerSecond := 0.0
	generatorFuelConsumptionPerSecond := 0.0
	neededPower := 0.0
	generatorGroups := make([]*data.EquipmentGroup, 0)
	for _, group := range sortedEquipmentGroups(world.data.EquipmentGroups.GetByCosmicObjectID(cosmicObject.ID)) {
		model, ok := world.data.ItemModels.Get(group.EquipmentItemModelID)
		if !ok {
			group.Active = false
			continue
		}

		enabledCount := enabledEquipmentCount(group)
		if enabledCount <= 0 {
			group.Active = false
			continue
		}
		if !shipHasFuel(*cosmicObject) && equipmentConsumesStoredFuel(*model) {
			group.Active = false
			continue
		}

		count := float64(enabledCount)
		cosmicObject.GeneratingPower += model.GeneratingPower * count
		if model.GeneratingPower != 0 {
			group.Active = false
			generatorGroups = append(generatorGroups, group)
			if model.ConsumingItemModelID > 0 && model.ConsumingCount > 0 {
				generatorFuelConsumptionPerSecond += model.ConsumingCount * count
			}
		}

		group.Active = equipmentIsActive(*cosmicObject, *model) || world.constructorEquipmentIsWorkingLocked(group.ID)
		if !group.Active {
			continue
		}

		neededPower += model.ConsumingPower * count
		if model.ConsumingItemModelID > 0 && model.ConsumingCount > 0 {
			consumerFuelConsumptionPerSecond += model.ConsumingCount * count
		}
	}

	electricShare := electricWorkShare(cosmicObject.GeneratingPower, neededPower)
	cosmicObject.ConsumingPower = neededPower * electricShare

	fuelConsumptionPerSecond := consumerFuelConsumptionPerSecond * electricShare
	if math.Abs(cosmicObject.ConsumingPower) > physics.Epsilon && math.Abs(cosmicObject.GeneratingPower) > physics.Epsilon {
		generatorLoad := math.Abs(cosmicObject.ConsumingPower / cosmicObject.GeneratingPower)
		fuelConsumptionPerSecond += generatorFuelConsumptionPerSecond * generatorLoad
		for _, group := range generatorGroups {
			group.Active = true
		}
	}

	if dtSeconds > 0 && fuelConsumptionPerSecond > 0 {
		cosmicObject.Fuel = math.Max(0, cosmicObject.Fuel-fuelConsumptionPerSecond*dtSeconds)
	}
}

// Р вЂ™Р С•Р В·Р Р†РЎР‚Р В°РЎвЂ°Р В°Р ВµРЎвЂљ РЎвЂћР В°Р С”РЎвЂљР С‘РЎвЂЎР ВµРЎРѓР С”Р С‘ Р Р†Р С”Р В»РЎР‹РЎвЂЎР ВµР Р…Р Р…Р С•Р Вµ Р С”Р С•Р В»Р С‘РЎвЂЎР ВµРЎРѓРЎвЂљР Р†Р С• Р ВµР Т‘Р С‘Р Р…Р С‘РЎвЂ  Р С•Р В±Р С•РЎР‚РЎС“Р Т‘Р С•Р Р†Р В°Р Р…Р С‘РЎРЏ.
func enabledEquipmentCount(group *data.EquipmentGroup) int64 {
	if group == nil || !group.Enabled {
		return 0
	}
	return group.EnabledCount
}

// Р СџРЎР‚Р С•Р Р†Р ВµРЎР‚РЎРЏР ВµРЎвЂљ, РЎвЂЎРЎвЂљР С• Р Р† Р В±Р В°Р С”Р Вµ Р ВµРЎРѓРЎвЂљРЎРЉ РЎР‚Р ВµРЎРѓРЎС“РЎР‚РЎРѓ Р Т‘Р В»РЎРЏ РЎвЂљР С•Р С—Р В»Р С‘Р Р†Р С•Р В·Р В°Р Р†Р С‘РЎРѓР С‘Р СР С•Р С–Р С• Р С•Р В±Р С•РЎР‚РЎС“Р Т‘Р С•Р Р†Р В°Р Р…Р С‘РЎРЏ.
func shipHasFuel(cosmicObject data.CosmicObject) bool {
	return cosmicObject.Fuel > physics.Epsilon
}

// Р СџРЎР‚Р С•Р Р†Р ВµРЎР‚РЎРЏР ВµРЎвЂљ, РЎвЂЎРЎвЂљР С• Р СР С•Р Т‘Р ВµР В»РЎРЉ РЎвЂљРЎР‚Р В°РЎвЂљР С‘РЎвЂљ РЎвЂ¦РЎР‚Р В°Р Р…Р С‘Р СРЎвЂ№Р в„– РЎР‚Р ВµРЎРѓРЎС“РЎР‚РЎРѓ Р С”Р С•РЎР‚Р В°Р В±Р В»РЎРЏ.
func equipmentConsumesStoredFuel(model data.ItemModel) bool {
	return model.ConsumingItemModelID > 0 && model.ConsumingCount > 0
}

// Р С›Р С—РЎР‚Р ВµР Т‘Р ВµР В»РЎРЏР ВµРЎвЂљ Р Т‘Р С•Р В»РЎР‹ РЎР‚Р В°Р В±Р С•РЎвЂљРЎвЂ№, Р С”Р С•РЎвЂљР С•РЎР‚РЎС“РЎР‹ Р СР С•Р В¶Р Р…Р С• Р С•Р В±Р ВµРЎРѓР С—Р ВµРЎвЂЎР С‘РЎвЂљРЎРЉ Р С‘Р СР ВµРЎР‹РЎвЂ°Р ВµР в„–РЎРѓРЎРЏ РЎРЊР В»Р ВµР С”РЎвЂљРЎР‚Р С•РЎРЊР Р…Р ВµРЎР‚Р С–Р С‘Р ВµР в„–.
func electricWorkShare(generatedPower float64, neededPower float64) float64 {
	if neededPower <= 0 || generatedPower >= neededPower {
		return 1
	}
	if generatedPower <= 0 {
		return 0
	}
	return generatedPower / neededPower
}

// Р СџРЎР‚Р С•Р Р†Р ВµРЎР‚РЎРЏР ВµРЎвЂљ, Р В±РЎС“Р Т‘Р ВµРЎвЂљ Р В»Р С‘ Р СР С•Р Т‘Р ВµР В»РЎРЉ РЎвЂљРЎР‚Р В°РЎвЂљР С‘РЎвЂљРЎРЉ РЎРЊР В»Р ВµР С”РЎвЂљРЎР‚Р С‘РЎвЂЎР ВµРЎРѓРЎвЂљР Р†Р С• Р С—РЎР‚Р С‘ РЎвЂљР ВµР С”РЎС“РЎвЂ°Р ВµР С РЎС“Р С—РЎР‚Р В°Р Р†Р В»Р ВµР Р…Р С‘Р С‘.
func equipmentNeedsElectricityForInput(input game.ShipInput, model data.ItemModel) bool {
	usesAlongForce := model.MaxAlongForce != 0 && (input.ThrustForward || input.ThrustBackward)
	usesAcrossForce := model.MaxAcrossForce != 0 && (input.ThrustLeft || input.ThrustRight)
	usesTorque := model.MaxTorque != 0 && input.TargetRotationDelta != 0
	return usesAlongForce || usesAcrossForce || usesTorque
}

// Р С›Р С—РЎР‚Р ВµР Т‘Р ВµР В»РЎРЏР ВµРЎвЂљ, Р Р†РЎвЂ№Р С—Р С•Р В»Р Р…РЎРЏР ВµРЎвЂљ Р В»Р С‘ Р С•Р В±Р С•РЎР‚РЎС“Р Т‘Р С•Р Р†Р В°Р Р…Р С‘Р Вµ РЎР‚Р В°Р В±Р С•РЎвЂљРЎС“ Р Р† РЎвЂљР ВµР С”РЎС“РЎвЂ°Р ВµР С РЎвЂљР С‘Р С”Р Вµ.
func equipmentIsActive(cosmicObject data.CosmicObject, model data.ItemModel) bool {
	usesLinearForce := model.MaxAlongForce != 0 || model.MaxAcrossForce != 0
	usesTorque := model.MaxTorque != 0
	if usesLinearForce || usesTorque {
		return (usesLinearForce && (math.Abs(cosmicObject.AlongForce) > physics.Epsilon || math.Abs(cosmicObject.AcrossForce) > physics.Epsilon)) ||
			(usesTorque && math.Abs(cosmicObject.Torque) > physics.Epsilon)
	}

	return false
}

// Р вЂ™Р С•Р В·Р Р†РЎР‚Р В°РЎвЂ°Р В°Р ВµРЎвЂљ Р С–РЎР‚РЎС“Р С—Р С—РЎвЂ№ Р С•Р В±Р С•РЎР‚РЎС“Р Т‘Р С•Р Р†Р В°Р Р…Р С‘РЎРЏ Р Р† РЎРѓРЎвЂљР В°Р В±Р С‘Р В»РЎРЉР Р…Р С•Р С Р С—Р С•РЎР‚РЎРЏР Т‘Р С”Р Вµ ID.
func sortedEquipmentGroups(groups []*data.EquipmentGroup) []*data.EquipmentGroup {
	result := append([]*data.EquipmentGroup(nil), groups...)
	sort.Slice(result, func(left int, right int) bool {
		return result[left].ID < result[right].ID
	})
	return result
}

// ackMutationLocked Р В·Р В°Р С—Р С•Р СР С‘Р Р…Р В°Р ВµРЎвЂљ Р С—Р С•РЎРѓР В»Р ВµР Т‘Р Р…Р С‘Р в„– Р С•Р В±РЎР‚Р В°Р В±Р С•РЎвЂљР В°Р Р…Р Р…РЎвЂ№Р в„– Р Р…Р С•Р СР ВµРЎР‚ Р С”Р С•Р СР В°Р Р…Р Т‘РЎвЂ№ Р С—Р В°Р Р…Р ВµР В»Р С‘ Р С—Р С•Р Т‘ РЎС“Р В¶Р Вµ Р Р†Р В·РЎРЏРЎвЂљРЎвЂ№Р С mutex.
func (world *World) ackMutationLocked(accountID int64, sessionID string, mutationSeq int64) {
	if sessionID == "" || mutationSeq <= 0 {
		return
	}
	key := mutationAckKey(accountID, sessionID)
	if mutationSeq > world.mutationAcks[key] {
		world.mutationAcks[key] = mutationSeq
	}
}

// mutationAckKey РЎРѓР С•Р В±Р С‘РЎР‚Р В°Р ВµРЎвЂљ Р С”Р В»РЎР‹РЎвЂЎ Р С—Р С•Р Т‘РЎвЂљР Р†Р ВµРЎР‚Р В¶Р Т‘Р ВµР Р…Р С‘РЎРЏ Р С”Р С•Р СР В°Р Р…Р Т‘ Р С—Р В°Р Р…Р ВµР В»Р С‘ Р Т‘Р В»РЎРЏ Р В°Р С”Р С”Р В°РЎС“Р Р…РЎвЂљР В° Р С‘ РЎРѓР ВµРЎРѓРЎРѓР С‘Р С‘.
func mutationAckKey(accountID int64, sessionID string) string {
	return fmt.Sprintf("%d:%s", accountID, sessionID)
}

// Р СџРЎР‚Р С•Р Р†Р ВµРЎР‚РЎРЏР ВµРЎвЂљ, РЎвЂЎРЎвЂљР С• Р СР С•Р Т‘Р ВµР В»РЎРЉ Р С•РЎвЂљР Р…Р С•РЎРѓР С‘РЎвЂљРЎРѓРЎРЏ Р С” Р С”Р С•РЎР‚Р В°Р В±Р В»РЎРЏР С.
func (world *World) isShipModel(model *data.CosmicObjectModel) bool {
	cosmicObjectType, ok := world.data.CosmicObjectTypes.Get(model.CosmicObjectTypeID)
	return ok && cosmicObjectType.Acronym == "Ship"
}

// Р В Р ВµРЎв‚¬Р В°Р ВµРЎвЂљ РЎРѓРЎвЂљР С•Р В»Р С”Р Р…Р С•Р Р†Р ВµР Р…Р С‘РЎРЏ Р Р†РЎРѓР ВµРЎвЂ¦ Р С—Р В°РЎР‚ РЎвЂљР ВµР В» Р С—Р С•РЎРѓР В»Р Вµ Р Т‘Р Р†Р С‘Р В¶Р ВµР Р…Р С‘РЎРЏ Р С•Р В±РЎР‰Р ВµР С”РЎвЂљР С•Р Р†.
func (world *World) resolveAllCollisions() {
	objectIDs := make([]int64, 0, len(world.data.CosmicObjects.Items))
	for objectID := range world.data.CosmicObjects.Items {
		objectIDs = append(objectIDs, objectID)
	}
	sort.Slice(objectIDs, func(left int, right int) bool {
		return objectIDs[left] < objectIDs[right]
	})

	for leftIndex := 0; leftIndex < len(objectIDs); leftIndex++ {
		first, ok := world.data.CosmicObjects.Get(objectIDs[leftIndex])
		if !ok {
			continue
		}
		firstModel, ok := world.data.CosmicObjectModels.Get(first.CosmicObjectModelID)
		if !ok {
			continue
		}

		for rightIndex := leftIndex + 1; rightIndex < len(objectIDs); rightIndex++ {
			second, ok := world.data.CosmicObjects.Get(objectIDs[rightIndex])
			if !ok {
				continue
			}
			secondModel, ok := world.data.CosmicObjectModels.Get(second.CosmicObjectModelID)
			if !ok {
				continue
			}
			collision, collided := physics.CollisionInfo(*first, *firstModel, *second, *secondModel)
			if !collided {
				continue
			}
			nextFirst, nextSecond := physics.ApplyCollisionResponse(*first, *firstModel, *second, *secondModel, collision)
			*first = nextFirst
			*second = nextSecond
		}
	}
}

// Р вЂ™Р С•Р В·Р Р†РЎР‚Р В°РЎвЂ°Р В°Р ВµРЎвЂљ РЎРѓР Р…Р С‘Р СР С•Р С” Р СР С‘РЎР‚Р В° РЎРѓ Р В·Р В°Р С—Р С•Р В»Р Р…Р ВµР Р…Р Р…РЎвЂ№Р С ID Р С•Р В±РЎР‰Р ВµР С”РЎвЂљР В° РЎвЂљР ВµР С”РЎС“РЎвЂ°Р ВµР С–Р С• Р С‘Р С–РЎР‚Р С•Р С”Р В°.
func (world *World) SnapshotForAccount(accountID int64) game.Snapshot {
	world.mu.Lock()
	defer world.mu.Unlock()

	objectID := world.accountObjectIDs[accountID]
	return world.snapshotLocked(objectID)
}

// Р вЂ™Р С•Р В·Р Р†РЎР‚Р В°РЎвЂ°Р В°Р ВµРЎвЂљ Р С•Р В±РЎР‰Р ВµР С”РЎвЂљ, Р С”Р С•РЎвЂљР С•РЎР‚РЎвЂ№Р С РЎРѓР ВµР в„–РЎвЂЎР В°РЎРѓ РЎС“Р С—РЎР‚Р В°Р Р†Р В»РЎРЏР ВµРЎвЂљ Р С—Р С•Р Т‘Р С”Р В»РЎР‹РЎвЂЎР ВµР Р…Р Р…РЎвЂ№Р в„– Р В°Р С”Р С”Р В°РЎС“Р Р…РЎвЂљ.
func (world *World) ObjectIDForAccount(accountID int64) (int64, bool) {
	world.mu.Lock()
	defer world.mu.Unlock()

	objectID, ok := world.accountObjectIDs[accountID]
	return objectID, ok
}

// DrainDockingEvents Р В·Р В°Р В±Р С‘РЎР‚Р В°Р ВµРЎвЂљ Р Р…Р В°Р С”Р С•Р С—Р В»Р ВµР Р…Р Р…РЎвЂ№Р Вµ РЎРѓР С•Р В±РЎвЂ№РЎвЂљР С‘РЎРЏ РЎРѓРЎвЂљРЎвЂ№Р С”Р С•Р Р†Р С”Р С‘ Р Т‘Р В»РЎРЏ РЎРѓР ВµРЎвЂљР ВµР Р†Р С•Р в„– РЎР‚Р В°РЎРѓРЎРѓРЎвЂ№Р В»Р С”Р С‘.
func (world *World) DrainDockingEvents() []game.DockingEvent {
	world.mu.Lock()
	defer world.mu.Unlock()

	events := append([]game.DockingEvent(nil), world.dockingEvents...)
	world.dockingEvents = nil
	return events
}

// ClientMutationAck Р Р†Р С•Р В·Р Р†РЎР‚Р В°РЎвЂ°Р В°Р ВµРЎвЂљ Р С—Р С•РЎРѓР В»Р ВµР Т‘Р Р…Р С‘Р в„– Р С•Р В±РЎР‚Р В°Р В±Р С•РЎвЂљР В°Р Р…Р Р…РЎвЂ№Р в„– Р Р…Р С•Р СР ВµРЎР‚ Р С”Р С•Р СР В°Р Р…Р Т‘РЎвЂ№ Р С—Р В°Р Р…Р ВµР В»Р С‘ Р Т‘Р В»РЎРЏ Р С”Р В»Р С‘Р ВµР Р…РЎвЂљРЎРѓР С”Р С•Р в„– РЎРѓР ВµРЎРѓРЎРѓР С‘Р С‘.
func (world *World) ClientMutationAck(accountID int64, sessionID string) game.ClientMutationAck {
	world.mu.Lock()
	defer world.mu.Unlock()

	return game.ClientMutationAck{
		SessionID:      sessionID,
		LastAppliedSeq: world.mutationAcks[mutationAckKey(accountID, sessionID)],
	}
}

// Р вЂ™Р С•Р В·Р Р†РЎР‚Р В°РЎвЂ°Р В°Р ВµРЎвЂљ Р С—Р ВµРЎР‚РЎРѓР С•Р Р…Р В°Р В¶Р В° Р С—Р С• Р С‘Р Т‘Р ВµР Р…РЎвЂљР С‘РЎвЂћР С‘Р С”Р В°РЎвЂљР С•РЎР‚РЎС“ Р С‘Р В· Р В·Р В°РЎвЂ°Р С‘РЎвЂ°Р ВµР Р…Р Р…Р С•Р С–Р С• РЎРѓР С•РЎРѓРЎвЂљР С•РЎРЏР Р…Р С‘РЎРЏ.
func (world *World) CharacterByID(id int64) (*data.Character, bool) {
	world.mu.Lock()
	defer world.mu.Unlock()

	return world.data.Characters.Get(id)
}

// Р вЂ™Р С•Р В·Р Р†РЎР‚Р В°РЎвЂ°Р В°Р ВµРЎвЂљ Р С”Р С•РЎРѓР СР С‘РЎвЂЎР ВµРЎРѓР С”Р С‘Р в„– Р С•Р В±РЎР‰Р ВµР С”РЎвЂљ Р С—Р С• Р С‘Р Т‘Р ВµР Р…РЎвЂљР С‘РЎвЂћР С‘Р С”Р В°РЎвЂљР С•РЎР‚РЎС“ Р С‘Р В· Р В·Р В°РЎвЂ°Р С‘РЎвЂ°Р ВµР Р…Р Р…Р С•Р С–Р С• РЎРѓР С•РЎРѓРЎвЂљР С•РЎРЏР Р…Р С‘РЎРЏ.
func (world *World) CosmicObjectByID(id int64) (*data.CosmicObject, bool) {
	world.mu.Lock()
	defer world.mu.Unlock()

	return world.data.CosmicObjects.Get(id)
}

// Р РЋР С•РЎвЂ¦РЎР‚Р В°Р Р…РЎРЏР ВµРЎвЂљ Р С‘Р В·Р СР ВµР Р…РЎРЏР ВµР СР С•Р Вµ РЎРѓР С•РЎРѓРЎвЂљР С•РЎРЏР Р…Р С‘Р Вµ Р СР С‘РЎР‚Р В° Р С•Р В±РЎР‚Р В°РЎвЂљР Р…Р С• Р Р† JSON-РЎвЂћР В°Р в„–Р В»РЎвЂ№ РЎРѓР ВµРЎР‚Р Р†Р ВµРЎР‚Р В°.
func (world *World) SaveData(workingDirectory string) error {
	world.mu.Lock()
	defer world.mu.Unlock()

	dataDirectory := filepath.Join(workingDirectory, "data")
	if err := world.data.Accounts.SaveToFile(filepath.Join(dataDirectory, "Accounts.json")); err != nil {
		return err
	}
	if err := world.data.Characters.SaveToFile(filepath.Join(dataDirectory, "Characters.json")); err != nil {
		return err
	}
	if err := world.data.CosmicObjects.SaveToFile(filepath.Join(dataDirectory, "CosmicObjects.json")); err != nil {
		return err
	}
	if err := world.data.CosmicObjectTypes.SaveToFile(filepath.Join(dataDirectory, "CosmicObjectTypes.json")); err != nil {
		return err
	}
	if err := world.data.CosmicObjectModels.SaveToFile(filepath.Join(dataDirectory, "CosmicObjectModels.json")); err != nil {
		return err
	}
	if world.data.Assemblies != nil {
		if err := world.data.Assemblies.SaveToFile(filepath.Join(dataDirectory, "Assemblies.json")); err != nil {
			return err
		}
	}
	if world.data.EquipmentGroups != nil {
		if err := world.data.EquipmentGroups.SaveToFile(filepath.Join(dataDirectory, "EquipmentGroups.json")); err != nil {
			return err
		}
	}
	if world.data.ItemGroups != nil {
		if err := world.data.ItemGroups.SaveToFile(filepath.Join(dataDirectory, "ItemGroups.json")); err != nil {
			return err
		}
	}
	if world.data.Tasks != nil {
		if err := world.data.Tasks.SaveToFile(filepath.Join(dataDirectory, "Tasks.json")); err != nil {
			return err
		}
	}
	if world.data.TaskItemGroups != nil {
		if err := world.data.TaskItemGroups.SaveToFile(filepath.Join(dataDirectory, "TaskItemGroups.json")); err != nil {
			return err
		}
	}
	if world.data.ItemTypes != nil {
		if err := world.data.ItemTypes.SaveToFile(filepath.Join(dataDirectory, "ItemTypes.json")); err != nil {
			return err
		}
	}
	if world.data.Chats != nil {
		if err := world.data.Chats.SaveToFile(filepath.Join(dataDirectory, "Chats.json")); err != nil {
			return err
		}
	}
	if world.data.ChatMembers != nil {
		if err := world.data.ChatMembers.SaveToFile(filepath.Join(dataDirectory, "ChatMembers.json")); err != nil {
			return err
		}
	}
	if world.data.CommunityTypes != nil {
		if err := world.data.CommunityTypes.SaveToFile(filepath.Join(dataDirectory, "CommunityTypes.json")); err != nil {
			return err
		}
	}
	if world.data.CommunityChatRoles != nil {
		if err := world.data.CommunityChatRoles.SaveToFile(filepath.Join(dataDirectory, "CommunityChatRoles.json")); err != nil {
			return err
		}
	}
	if world.data.Messages != nil {
		if err := world.data.Messages.SaveToFile(filepath.Join(dataDirectory, "Messages.json")); err != nil {
			return err
		}
	}
	if world.data.MessageReads != nil {
		if err := world.data.MessageReads.SaveToFile(filepath.Join(dataDirectory, "MessageReads.json")); err != nil {
			return err
		}
	}
	if world.data.MessageTypes != nil {
		if err := world.data.MessageTypes.SaveToFile(filepath.Join(dataDirectory, "MessageTypes.json")); err != nil {
			return err
		}
	}
	if world.data.ActionTypes != nil {
		if err := world.data.ActionTypes.SaveToFile(filepath.Join(dataDirectory, "ActionTypes.json")); err != nil {
			return err
		}
	}
	if world.data.InputEventTypes != nil {
		if err := world.data.InputEventTypes.SaveToFile(filepath.Join(dataDirectory, "InputEventTypes.json")); err != nil {
			return err
		}
	}
	if world.data.DefaultActionInputSettings != nil {
		if err := world.data.DefaultActionInputSettings.SaveToFile(filepath.Join(dataDirectory, "DefaultActionInputSettings.json")); err != nil {
			return err
		}
	}
	if world.data.AccountActionInputSettings != nil {
		if err := world.data.AccountActionInputSettings.SaveToFile(filepath.Join(dataDirectory, "AccountActionInputSettings.json")); err != nil {
			return err
		}
	}
	return nil
}

// Р РЋР С•Р В±Р С‘РЎР‚Р В°Р ВµРЎвЂљ Р Т‘Р ВµРЎвЂљР ВµРЎР‚Р СР С‘Р Р…Р С‘РЎР‚Р С•Р Р†Р В°Р Р…Р Р…Р С• Р С•РЎвЂљРЎРѓР С•РЎР‚РЎвЂљР С‘РЎР‚Р С•Р Р†Р В°Р Р…Р Р…РЎвЂ№Р в„– РЎРѓР Р…Р С‘Р СР С•Р С”; Р Р†РЎвЂ№Р В·РЎвЂ№Р Р†Р В°Р ВµРЎвЂљРЎРѓРЎРЏ РЎвЂљР С•Р В»РЎРЉР С”Р С• Р С—Р С•Р Т‘ mutex.
// Создает временные лучи активных буров только для текущего сетевого снимка.
func (world *World) activeDrillRayObjectsLocked() []game.SnapshotCosmicObject {
	if world.data.CosmicObjects == nil || world.data.CosmicObjectModels == nil {
		return nil
	}
	rayModel, ok := world.data.CosmicObjectModels.GetByAcronym(simpleDrillRayAcronym)
	if !ok {
		return nil
	}

	inputs := world.inputsByObjectID()
	objectIDs := make([]int64, 0, len(inputs))
	for objectID, input := range inputs {
		if input.PrimaryPointerAction {
			objectIDs = append(objectIDs, objectID)
		}
	}
	sort.Slice(objectIDs, func(left int, right int) bool {
		return objectIDs[left] < objectIDs[right]
	})

	rays := make([]game.SnapshotCosmicObject, 0, len(objectIDs))
	for _, objectID := range objectIDs {
		cosmicObject, ok := world.data.CosmicObjects.Get(objectID)
		if !ok || !cosmicObject.Enabled {
			continue
		}
		model, ok := world.data.CosmicObjectModels.Get(cosmicObject.CosmicObjectModelID)
		if !ok {
			continue
		}
		if _, ok := world.enabledSimpleDrillRangeLocked(objectID); !ok {
			continue
		}

		forward := physics.ForwardVector(cosmicObject.Rotation)
		centerDistance := modelVisualForwardOffsetMeters(*model) + rayModel.BodyLength/2
		ray := data.CosmicObject{
			ID:                  -objectID,
			Title:               simpleDrillRayAcronym,
			CosmicObjectModelID: rayModel.ID,
			X:                   cosmicObject.X + forward.X*centerDistance,
			Y:                   cosmicObject.Y + forward.Y*centerDistance,
			Rotation:            cosmicObject.Rotation,
			TargetRotation:      cosmicObject.Rotation,
			Enabled:             true,
		}
		rays = append(rays, game.SnapshotCosmicObject{CosmicObject: ray})
	}

	return rays
}

// Возвращает расстояние от центра модели до видимого носа по направлению вперед.
func modelVisualForwardOffsetMeters(model data.CosmicObjectModel) float64 {
	if model.TextureScale > 0 && model.TextureVisibleTopY > 0 && model.TextureBodyOriginY > model.TextureVisibleTopY {
		return float64(model.TextureBodyOriginY-model.TextureVisibleTopY) / model.TextureScale
	}

	return model.BodyLength / 2
}

// Ищет включенный простой бур на объекте и возвращает дальность его действия.
func (world *World) enabledSimpleDrillRangeLocked(objectID int64) (float64, bool) {
	if world.data.EquipmentGroups == nil || world.data.ItemModels == nil {
		return 0, false
	}

	var selectedRange float64
	for _, group := range sortedEquipmentGroups(world.data.EquipmentGroups.GetByCosmicObjectID(objectID)) {
		if enabledEquipmentCount(group) <= 0 {
			continue
		}
		model, ok := world.data.ItemModels.Get(group.EquipmentItemModelID)
		if !ok || model.Acronym != simpleDrillAcronym || model.Range <= 0 {
			continue
		}
		if model.Range > selectedRange {
			selectedRange = model.Range
		}
	}

	return selectedRange, selectedRange > 0
}

func (world *World) snapshotLocked(selfObjectID int64) game.Snapshot {
	objectIDs := make([]int64, 0, len(world.data.CosmicObjects.Items))
	for objectID := range world.data.CosmicObjects.Items {
		objectIDs = append(objectIDs, objectID)
	}
	sort.Slice(objectIDs, func(left int, right int) bool {
		return objectIDs[left] < objectIDs[right]
	})

	objects := make([]game.SnapshotCosmicObject, 0, len(objectIDs))
	for _, objectID := range objectIDs {
		cosmicObject, ok := world.data.CosmicObjects.Get(objectID)
		if !ok {
			continue
		}
		objects = append(objects, game.SnapshotCosmicObject{
			CosmicObject: *cosmicObject,
			OwnerName:    world.ownerNameForTestingLocked(cosmicObject.OwnerCharacterID),
		})
	}
	objects = append(objects, world.activeDrillRayObjectsLocked()...)

	equipmentGroups := make([]data.EquipmentGroup, 0)
	if world.data.EquipmentGroups != nil {
		groupIDs := make([]int64, 0, len(world.data.EquipmentGroups.Items))
		for groupID := range world.data.EquipmentGroups.Items {
			groupIDs = append(groupIDs, groupID)
		}
		sort.Slice(groupIDs, func(left int, right int) bool {
			return groupIDs[left] < groupIDs[right]
		})
		for _, groupID := range groupIDs {
			group, ok := world.data.EquipmentGroups.Get(groupID)
			if !ok {
				continue
			}
			equipmentGroups = append(equipmentGroups, *group)
		}
	}

	itemGroups := make([]data.ItemGroup, 0)
	if world.data.ItemGroups != nil {
		itemGroupIDs := make([]int64, 0, len(world.data.ItemGroups.Items))
		for itemGroupID := range world.data.ItemGroups.Items {
			itemGroupIDs = append(itemGroupIDs, itemGroupID)
		}
		sort.Slice(itemGroupIDs, func(left int, right int) bool {
			return itemGroupIDs[left] < itemGroupIDs[right]
		})
		for _, itemGroupID := range itemGroupIDs {
			itemGroup, ok := world.data.ItemGroups.Get(itemGroupID)
			if !ok {
				continue
			}
			itemGroups = append(itemGroups, *itemGroup)
		}
	}

	tasks := make([]data.Task, 0)
	if world.data.Tasks != nil {
		taskIDs := make([]int64, 0, len(world.data.Tasks.Items))
		for taskID := range world.data.Tasks.Items {
			taskIDs = append(taskIDs, taskID)
		}
		sort.Slice(taskIDs, func(left int, right int) bool { return taskIDs[left] < taskIDs[right] })
		for _, taskID := range taskIDs {
			task, ok := world.data.Tasks.Get(taskID)
			if !ok {
				continue
			}
			tasks = append(tasks, *task)
		}
	}

	taskItemGroups := make([]data.TaskItemGroup, 0)
	if world.data.TaskItemGroups != nil {
		groupIDs := make([]int64, 0, len(world.data.TaskItemGroups.Items))
		for groupID := range world.data.TaskItemGroups.Items {
			groupIDs = append(groupIDs, groupID)
		}
		sort.Slice(groupIDs, func(left int, right int) bool { return groupIDs[left] < groupIDs[right] })
		for _, groupID := range groupIDs {
			group, ok := world.data.TaskItemGroups.Items[groupID]
			if !ok || group == nil {
				continue
			}
			taskItemGroups = append(taskItemGroups, *group)
		}
	}
	constructorProductionJobs := world.constructorProductionJobsForTestsLocked(tasks)

	return game.Snapshot{
		Type:                      "snapshot",
		Tick:                      world.tick,
		SelfObjectID:              selfObjectID,
		Objects:                   objects,
		EquipmentGroups:           equipmentGroups,
		ItemGroups:                itemGroups,
		Tasks:                     tasks,
		TaskItemGroups:            taskItemGroups,
		ConstructorProductionJobs: constructorProductionJobs,
	}
}

// Р ВРЎвЂ°Р ВµРЎвЂљ Р С—Р ВµРЎР‚Р Р†РЎС“РЎР‹ Р С—РЎС“Р В±Р В»Р С‘РЎвЂЎР Р…РЎС“РЎР‹ РЎРѓР С‘РЎРѓРЎвЂљР ВµР СР Р…РЎС“РЎР‹ РЎРѓР В±Р С•РЎР‚Р С”РЎС“ Р Т‘Р В»РЎРЏ Р СР С•Р Т‘Р ВµР В»Р С‘ Р С”Р С•РЎР‚Р С—РЎС“РЎРѓР В°.
// constructorProductionJobsForTestsLocked собирает старое представление очереди для существующих серверных тестов.
func (world *World) constructorProductionJobsForTestsLocked(tasks []data.Task) []game.ConstructorProductionJob {
	byKey := map[string]*game.ConstructorProductionJob{}
	keys := make([]string, 0)
	for _, task := range tasks {
		queueType := "main"
		if task.ParentTaskID > 0 {
			queueType = "auxiliary"
		}
		key := fmt.Sprintf("%d:%s:%d:%d:%d", task.ControllerEquipmentGroupID, queueType, task.ParentTaskID, task.SchemaID, task.BlueprintID)
		job := byKey[key]
		if job == nil {
			job = &game.ConstructorProductionJob{
				QueueType:      queueType,
				RemainingCount: 0,
				TotalCount:     0,
				RemainingTime:  task.RemainingEnergy,
				TotalTime:      task.TotalEnergy,
				Running:        false,
				ParentJobID:    task.ParentTaskID,
			}
			byKey[key] = job
			keys = append(keys, key)
		}
		if job.ID == 0 || task.ID < job.ID {
			job.ID = task.ID
			job.ConstructorEquipmentGroupID = task.ControllerEquipmentGroupID
			job.SchemaID = task.SchemaID
			job.BlueprintID = task.BlueprintID
		}
		if task.RemainingEnergy < task.TotalEnergy || world.taskHasReserveLocked(task.ID) {
			job.Running = true
			job.RemainingTime = task.RemainingEnergy
		}
		if task.SchemaID > 0 {
			if schema, err := world.itemSchemaLocked(task.SchemaID); err == nil {
				amount := taskCount(&task)
				job.ProductItemModelID = schema.ItemModelID
				job.ProductCount = schema.Count
				job.RemainingCount += schema.Count * remainingTaskCount(task.RemainingEnergy, task.TotalEnergy, amount)
				job.TotalCount += schema.Count * amount
			}
		}
		if task.BlueprintID > 0 {
			if blueprint, err := world.objectBlueprintLocked(task.BlueprintID); err == nil {
				amount := taskCount(&task)
				job.ProductCosmicObjectModelID = blueprint.CosmicObjectModelID
				job.ProductCount = 1
				job.RemainingCount += remainingTaskCount(task.RemainingEnergy, task.TotalEnergy, amount)
				job.TotalCount += amount
			}
		}
	}
	result := make([]game.ConstructorProductionJob, 0, len(keys))
	for _, key := range keys {
		result = append(result, *byKey[key])
	}
	sort.Slice(result, func(left int, right int) bool {
		if result[left].QueueType != result[right].QueueType {
			return result[left].QueueType == "auxiliary"
		}
		return result[left].ID < result[right].ID
	})
	return result
}

// remainingTaskCount оценивает оставшееся количество результата по доле невыполненной работы.
func remainingTaskCount(remainingEnergy float64, totalEnergy float64, amount float64) float64 {
	if amount <= 0 {
		return 1
	}
	if totalEnergy <= physics.Epsilon {
		return amount
	}
	if remainingEnergy <= physics.Epsilon {
		return 0
	}
	completed := math.Floor(((totalEnergy - remainingEnergy) / totalEnergy * amount) + physics.Epsilon)
	remaining := amount - completed
	if remaining < 0 {
		return 0
	}
	return remaining
}

func (world *World) firstPublicDeveloperAssembly(cosmicObjectModelID int64) (*data.Assembly, bool) {
	if world.data.Assemblies == nil {
		return nil, false
	}
	return world.data.Assemblies.FirstPublicDeveloperAssembly(cosmicObjectModelID)
}

// Р С›Р В±Р Р…Р С•Р Р†Р В»РЎРЏР ВµРЎвЂљ РЎРѓР С•РЎвЂ¦РЎР‚Р В°Р Р…Р ВµР Р…Р Р…РЎвЂ№Р Вµ Р С”Р С•РЎР‚Р В°Р В±Р В»Р С‘ Р С—Р С• РЎРѓР С‘РЎРѓРЎвЂљР ВµР СР Р…РЎвЂ№Р С Р С—РЎС“Р В±Р В»Р С‘РЎвЂЎР Р…РЎвЂ№Р С РЎРѓР В±Р С•РЎР‚Р С”Р В°Р С, РЎРѓР С•РЎвЂ¦РЎР‚Р В°Р Р…РЎРЏРЎРЏ Р С‘РЎвЂ¦ Р Т‘Р Р†Р С‘Р В¶Р ВµР Р…Р С‘Р Вµ Р С‘ Р Р†Р В»Р В°Р Т‘Р ВµР В»РЎРЉРЎвЂ Р ВµР Р†.
func (world *World) applyAssembliesToLoadedShips() {
	if world.data.CosmicObjects == nil || world.data.CosmicObjectModels == nil || world.data.CosmicObjectTypes == nil {
		return
	}

	shipType, ok := world.data.CosmicObjectTypes.GetByAcronym("Ship")
	if !ok {
		return
	}

	for _, cosmicObject := range world.data.CosmicObjects.Items {
		model, ok := world.data.CosmicObjectModels.Get(cosmicObject.CosmicObjectModelID)
		if !ok || model.CosmicObjectTypeID != shipType.ID {
			continue
		}
		assembly, ok := world.firstPublicDeveloperAssembly(model.ID)
		if !ok {
			continue
		}
		world.applyModelAndAssembly(cosmicObject, model, assembly)
		_ = world.ensureEquipmentFromAssembly(cosmicObject.ID, assembly)
	}
}

// Р РЋР С•Р В±Р С‘РЎР‚Р В°Р ВµРЎвЂљ Р Р…Р С•Р Р†РЎвЂ№Р в„– Р С•Р В±РЎР‰Р ВµР С”РЎвЂљ Р С‘Р В· Р СР С•Р Т‘Р ВµР В»Р С‘ Р С”Р С•РЎР‚Р С—РЎС“РЎРѓР В° Р С‘ РЎР‚Р В°РЎРѓРЎРѓРЎвЂЎР С‘РЎвЂљР В°Р Р…Р Р…Р С•Р в„– РЎРѓР В±Р С•РЎР‚Р С”Р С‘.
func (world *World) cosmicObjectFromModelAndAssembly(model *data.CosmicObjectModel, assembly *data.Assembly) *data.CosmicObject {
	cosmicObject := &data.CosmicObject{Enabled: true}
	world.applyModelAndAssembly(cosmicObject, model, assembly)
	cosmicObject.Armor = assembly.MaxArmor
	cosmicObject.Fuel = assembly.MaxFuel
	return cosmicObject
}

// Р СџРЎР‚Р С‘Р СР ВµР Р…РЎРЏР ВµРЎвЂљ РЎР‚Р В°РЎРѓРЎРѓРЎвЂЎР С‘РЎвЂљР В°Р Р…Р Р…РЎвЂ№Р Вµ РЎвЂ¦Р В°РЎР‚Р В°Р С”РЎвЂљР ВµРЎР‚Р С‘РЎРѓРЎвЂљР С‘Р С”Р С‘ РЎРѓР В±Р С•РЎР‚Р С”Р С‘, Р Р…Р Вµ РЎвЂљРЎР‚Р С•Р С–Р В°РЎРЏ Р Т‘Р Р†Р С‘Р В¶Р ВµР Р…Р С‘Р Вµ Р С‘ Р Р†Р В»Р В°Р Т‘Р ВµР Р…Р С‘Р Вµ Р С•Р В±РЎР‰Р ВµР С”РЎвЂљР С•Р С.
func (world *World) applyModelAndAssembly(cosmicObject *data.CosmicObject, model *data.CosmicObjectModel, assembly *data.Assembly) {
	cosmicObject.Title = model.TitleRu
	cosmicObject.CosmicObjectModelID = model.ID
	cosmicObject.Mass = assembly.Mass
	cosmicObject.Capacity = model.Capacity
	cosmicObject.MaxArmor = assembly.MaxArmor
	cosmicObject.MaxSpeed = model.MaxSpeed
	cosmicObject.MaxAngularSpeed = model.MaxAngularSpeed
	cosmicObject.MaxAlongForce = assembly.MaxAlongForce
	cosmicObject.MaxAcrossForce = assembly.MaxAcrossForce
	cosmicObject.MaxTorque = assembly.MaxTorque
	cosmicObject.GeneratingPower = assembly.GeneratingPower
	cosmicObject.ConsumingPower = assembly.ConsumingPower
	cosmicObject.Complexity = assembly.Complexity
	cosmicObject.OccupiedVolume = assembly.OccupiedVolume
	cosmicObject.MaxFuel = assembly.MaxFuel
	if cosmicObject.Armor > assembly.MaxArmor {
		cosmicObject.Armor = assembly.MaxArmor
	}
	if cosmicObject.Fuel > assembly.MaxFuel {
		cosmicObject.Fuel = assembly.MaxFuel
	}
}

// Р Р€РЎРѓРЎвЂљР В°Р Р…Р В°Р Р†Р В»Р С‘Р Р†Р В°Р ВµРЎвЂљ Р С•Р В±Р С•РЎР‚РЎС“Р Т‘Р С•Р Р†Р В°Р Р…Р С‘Р Вµ Р С‘Р В· РЎРѓР В±Р С•РЎР‚Р С”Р С‘, Р ВµРЎРѓР В»Р С‘ РЎС“ Р С•Р В±РЎР‰Р ВµР С”РЎвЂљР В° Р ВµРЎвЂ°Р Вµ Р Р…Р ВµРЎвЂљ Р С•Р В±Р С•РЎР‚РЎС“Р Т‘Р С•Р Р†Р В°Р Р…Р С‘РЎРЏ.
func (world *World) ensureEquipmentFromAssembly(cosmicObjectID int64, assembly *data.Assembly) error {
	if world.data.EquipmentGroups == nil || len(world.data.EquipmentGroups.GetByCosmicObjectID(cosmicObjectID)) > 0 {
		return nil
	}
	return world.installEquipmentFromAssembly(cosmicObjectID, assembly)
}

// Р вЂ”Р В°Р СР ВµР Р…РЎРЏР ВµРЎвЂљ Р С•Р В±Р С•РЎР‚РЎС“Р Т‘Р С•Р Р†Р В°Р Р…Р С‘Р Вµ Р С•Р В±РЎР‰Р ВµР С”РЎвЂљР В° Р Р…Р В° Р С•Р В±Р С•РЎР‚РЎС“Р Т‘Р С•Р Р†Р В°Р Р…Р С‘Р Вµ Р Р†РЎвЂ№Р В±РЎР‚Р В°Р Р…Р Р…Р С•Р в„– РЎРѓР В±Р С•РЎР‚Р С”Р С‘.
func (world *World) replaceEquipmentFromAssembly(cosmicObjectID int64, assembly *data.Assembly) error {
	if world.data.EquipmentGroups == nil {
		return nil
	}
	if world.data.ItemGroups != nil {
		equipmentGroupIDs := make([]int64, 0)
		for _, group := range world.data.EquipmentGroups.GetByCosmicObjectID(cosmicObjectID) {
			equipmentGroupIDs = append(equipmentGroupIDs, group.ID)
		}
		world.data.ItemGroups.DeleteByContainerEquipmentGroupIDs(equipmentGroupIDs)
	}
	world.data.EquipmentGroups.DeleteByCosmicObjectID(cosmicObjectID)
	return world.installEquipmentFromAssembly(cosmicObjectID, assembly)
}

// Р С™Р С•Р С—Р С‘РЎР‚РЎС“Р ВµРЎвЂљ Р С–РЎР‚РЎС“Р С—Р С—РЎвЂ№ Р С•Р В±Р С•РЎР‚РЎС“Р Т‘Р С•Р Р†Р В°Р Р…Р С‘РЎРЏ РЎРѓР В±Р С•РЎР‚Р С”Р С‘ Р Р…Р В° Р С”Р С•Р Р…Р С”РЎР‚Р ВµРЎвЂљР Р…РЎвЂ№Р в„– Р С•Р В±РЎР‰Р ВµР С”РЎвЂљ.
func (world *World) installEquipmentFromAssembly(cosmicObjectID int64, assembly *data.Assembly) error {
	if world.data.EquipmentGroups == nil || world.data.AssemblyEquipmentGroups == nil {
		return nil
	}

	for _, group := range world.data.AssemblyEquipmentGroups.GetByAssemblyID(assembly.ID) {
		if _, err := world.data.EquipmentGroups.Add(&data.EquipmentGroup{
			CosmicObjectID:       cosmicObjectID,
			Title:                group.Title,
			EquipmentItemModelID: group.EquipmentItemModelID,
			Count:                group.Count,
			EnabledCount:         group.Count,
			Enabled:              true,
			Active:               true,
		}); err != nil {
			return err
		}
	}
	return nil
}

// Р вЂ”Р В°Р С—Р С•Р В»Р Р…РЎРЏР ВµРЎвЂљ Р Р…Р С•Р Р†РЎвЂ№Р в„– Р С”Р С•РЎР‚Р В°Р В±Р В»РЎРЉ РЎвЂљР С•Р С—Р В»Р С‘Р Р†Р С•Р С Р С‘ Р С”Р В»Р В°Р Т‘Р ВµРЎвЂљ Р С•Р В±РЎР‚Р В°Р В·РЎвЂ РЎвЂ№ Р С—РЎР‚Р ВµР Т‘Р СР ВµРЎвЂљР С•Р Р† Р Р† Р С—Р ВµРЎР‚Р Р†РЎвЂ№Р в„– Р С”Р С•Р Р…РЎвЂљР ВµР в„–Р Р…Р ВµРЎР‚.
func (world *World) fillShipSupplies(cosmicObject *data.CosmicObject) {
	if cosmicObject == nil {
		return
	}
	cosmicObject.Fuel = cosmicObject.MaxFuel
	if world.data.EquipmentGroups == nil || world.data.ItemGroups == nil || world.data.ItemModels == nil || world.data.ItemTypes == nil {
		return
	}

	containerType, ok := world.data.ItemTypes.GetByAcronym("Container")
	if !ok {
		return
	}

	containerIDs := make([]int64, 0)
	for _, group := range world.data.EquipmentGroups.GetByCosmicObjectID(cosmicObject.ID) {
		model, ok := world.data.ItemModels.Get(group.EquipmentItemModelID)
		if !ok {
			continue
		}
		if model.ItemTypeID == containerType.ID {
			containerIDs = append(containerIDs, group.ID)
		}
	}
	if len(containerIDs) == 0 {
		return
	}
	sort.Slice(containerIDs, func(left int, right int) bool {
		return containerIDs[left] < containerIDs[right]
	})
	world.data.ItemGroups.DeleteByContainerEquipmentGroupIDs(containerIDs)

	sampleModelIDs := world.firstItemModelIDsByType()
	for _, itemModelID := range sampleModelIDs {
		_, _ = world.data.ItemGroups.Add(&data.ItemGroup{
			ContainerEquipmentGroupID: containerIDs[0],
			ContentItemModelID:        itemModelID,
			Count:                     10,
		})
	}
	resourceModelIDs := world.resourceItemModelIDs()
	for _, itemModelID := range resourceModelIDs {
		_, _ = world.data.ItemGroups.Add(&data.ItemGroup{
			ContainerEquipmentGroupID: containerIDs[0],
			ContentItemModelID:        itemModelID,
			Count:                     1000,
		})
	}
}

// Р вЂ™Р С•Р В·Р Р†РЎР‚Р В°РЎвЂ°Р В°Р ВµРЎвЂљ Р С—Р ВµРЎР‚Р Р†РЎС“РЎР‹ Р СР С•Р Т‘Р ВµР В»РЎРЉ Р С—РЎР‚Р ВµР Т‘Р СР ВµРЎвЂљР В° Р С”Р В°Р В¶Р Т‘Р С•Р С–Р С• РЎвЂљР С‘Р С—Р В° Р Р† РЎРѓРЎвЂљР В°Р В±Р С‘Р В»РЎРЉР Р…Р С•Р С Р С—Р С•РЎР‚РЎРЏР Т‘Р С”Р Вµ РЎвЂљР С‘Р С—Р С•Р Р†.
func (world *World) firstItemModelIDsByType() []int64 {
	firstByType := make(map[int64]int64)
	resourceTypeID := int64(0)
	if resourceType, ok := world.data.ItemTypes.GetByAcronym("Resource"); ok {
		resourceTypeID = resourceType.ID
	}
	for itemModelID, model := range world.data.ItemModels.Items {
		if model == nil || model.ItemTypeID <= 0 || model.ItemTypeID == resourceTypeID {
			continue
		}
		current, ok := firstByType[model.ItemTypeID]
		if !ok || itemModelID < current {
			firstByType[model.ItemTypeID] = itemModelID
		}
	}

	typeIDs := make([]int64, 0, len(firstByType))
	for ItemTypeID := range firstByType {
		typeIDs = append(typeIDs, ItemTypeID)
	}
	sort.Slice(typeIDs, func(left int, right int) bool {
		return typeIDs[left] < typeIDs[right]
	})

	result := make([]int64, 0, len(typeIDs))
	for _, ItemTypeID := range typeIDs {
		result = append(result, firstByType[ItemTypeID])
	}
	return result
}

// resourceItemModelIDs Р Р†Р С•Р В·Р Р†РЎР‚Р В°РЎвЂ°Р В°Р ВµРЎвЂљ Р Р†РЎРѓР Вµ Р СР С•Р Т‘Р ВµР В»Р С‘ РЎР‚Р ВµРЎРѓРЎС“РЎР‚РЎРѓР С•Р Р† Р Р† РЎРѓРЎвЂљР В°Р В±Р С‘Р В»РЎРЉР Р…Р С•Р С Р С—Р С•РЎР‚РЎРЏР Т‘Р С”Р Вµ Р Т‘Р В»РЎРЏ РЎвЂљР ВµРЎРѓРЎвЂљР С•Р Р†Р С•Р С–Р С• Р В·Р В°Р С—Р В°РЎРѓР В° Р Р…Р С•Р Р†Р С•Р С–Р С• Р С”Р С•РЎР‚Р В°Р В±Р В»РЎРЏ.
func (world *World) resourceItemModelIDs() []int64 {
	resourceType, ok := world.data.ItemTypes.GetByAcronym("Resource")
	if !ok || world.data.ItemModels == nil {
		return nil
	}
	result := make([]int64, 0)
	for itemModelID, model := range world.data.ItemModels.Items {
		if model != nil && model.ItemTypeID == resourceType.ID {
			result = append(result, itemModelID)
		}
	}
	sort.Slice(result, func(left int, right int) bool {
		return result[left] < result[right]
	})
	return result
}

// raySegmentPolygonDistance Р Р†Р С•Р В·Р Р†РЎР‚Р В°РЎвЂ°Р В°Р ВµРЎвЂљ РЎР‚Р В°РЎРѓРЎРѓРЎвЂљР С•РЎРЏР Р…Р С‘Р Вµ Р С•РЎвЂљ Р Р…Р В°РЎвЂЎР В°Р В»Р В° Р В»РЎС“РЎвЂЎР В° Р Т‘Р С• Р С—Р ВµРЎР‚Р Р†Р С•Р С–Р С• Р С—Р ВµРЎР‚Р ВµРЎРѓР ВµРЎвЂЎР ВµР Р…Р С‘РЎРЏ РЎРѓ РЎвЂљР ВµР В»Р С•Р С Р С•Р В±РЎР‰Р ВµР С”РЎвЂљР В°.
func raySegmentPolygonDistance(startX float64, startY float64, endX float64, endY float64, object data.CosmicObject, model data.CosmicObjectModel) (float64, bool) {
	points := transformedBodyPolygon(object, model)
	if len(points) < 3 {
		return 0, false
	}
	if pointInsidePolygon(startX, startY, points) {
		return 0, true
	}
	best := math.Inf(1)
	for index := range points {
		nextIndex := (index + 1) % len(points)
		distance, ok := segmentIntersectionDistance(startX, startY, endX, endY, points[index].X, points[index].Y, points[nextIndex].X, points[nextIndex].Y)
		if ok && distance < best {
			best = distance
		}
	}
	if math.IsInf(best, 1) {
		return 0, false
	}
	return best, true
}

// transformedBodyPolygon Р С—Р ВµРЎР‚Р ВµР Р†Р С•Р Т‘Р С‘РЎвЂљ Р В»Р С•Р С”Р В°Р В»РЎРЉР Р…Р С•Р Вµ РЎвЂљР ВµР В»Р С• Р СР С•Р Т‘Р ВµР В»Р С‘ Р Р† Р СР С‘РЎР‚Р С•Р Р†РЎвЂ№Р Вµ Р С”Р С•Р С•РЎР‚Р Т‘Р С‘Р Р…Р В°РЎвЂљРЎвЂ№ Р С•Р В±РЎР‰Р ВµР С”РЎвЂљР В°.
func transformedBodyPolygon(object data.CosmicObject, model data.CosmicObjectModel) []data.BodyPoint {
	points := model.BodyPolygon
	if len(points) == 0 {
		points = fallbackBodyPolygon(model)
	}
	result := make([]data.BodyPoint, 0, len(points))
	cosRotation := math.Cos(object.Rotation)
	sinRotation := math.Sin(object.Rotation)
	for _, point := range points {
		result = append(result, data.BodyPoint{
			X: object.X + point.X*cosRotation + point.Y*sinRotation,
			Y: object.Y - point.X*sinRotation + point.Y*cosRotation,
		})
	}
	return result
}

// fallbackBodyPolygon РЎРѓРЎвЂљРЎР‚Р С•Р С‘РЎвЂљ Р С—РЎР‚РЎРЏР СР С•РЎС“Р С–Р С•Р В»РЎРЉР Р…Р С•Р Вµ РЎвЂљР ВµР В»Р С•, Р ВµРЎРѓР В»Р С‘ РЎРѓР С—РЎР‚Р В°Р Р†Р С•РЎвЂЎР Р…Р С‘Р С” Р ВµРЎвЂ°РЎвЂ Р Р…Р Вµ РЎР‚Р В°РЎРѓРЎРѓРЎвЂЎР С‘РЎвЂљР В°Р В» Р СР Р…Р С•Р С–Р С•РЎС“Р С–Р С•Р В»РЎРЉР Р…Р С‘Р С”.
func fallbackBodyPolygon(model data.CosmicObjectModel) []data.BodyPoint {
	halfWidth := model.BodyWidth / 2
	halfLength := model.BodyLength / 2
	if halfWidth <= 0 && model.TextureScale > 0 {
		halfWidth = float64(model.TextureBodyWidth) / model.TextureScale / 2
	}
	if halfLength <= 0 && model.TextureScale > 0 {
		halfLength = float64(model.TextureBodyLength) / model.TextureScale / 2
	}
	return []data.BodyPoint{
		{X: -halfWidth, Y: -halfLength},
		{X: halfWidth, Y: -halfLength},
		{X: halfWidth, Y: halfLength},
		{X: -halfWidth, Y: halfLength},
	}
}

// pointInsidePolygon Р С—РЎР‚Р С•Р Р†Р ВµРЎР‚РЎРЏР ВµРЎвЂљ Р С—Р С•Р С—Р В°Р Т‘Р В°Р Р…Р С‘Р Вµ РЎвЂљР С•РЎвЂЎР С”Р С‘ Р Р†Р Р…РЎС“РЎвЂљРЎР‚РЎРЉ Р С—РЎР‚Р С•РЎРѓРЎвЂљР С•Р С–Р С• Р СР Р…Р С•Р С–Р С•РЎС“Р С–Р С•Р В»РЎРЉР Р…Р С‘Р С”Р В°.
func pointInsidePolygon(x float64, y float64, points []data.BodyPoint) bool {
	inside := false
	for index, point := range points {
		previous := points[(index+len(points)-1)%len(points)]
		if ((point.Y > y) != (previous.Y > y)) &&
			(x < (previous.X-point.X)*(y-point.Y)/(previous.Y-point.Y)+point.X) {
			inside = !inside
		}
	}
	return inside
}

// segmentIntersectionDistance Р Р†Р С•Р В·Р Р†РЎР‚Р В°РЎвЂ°Р В°Р ВµРЎвЂљ РЎР‚Р В°РЎРѓРЎРѓРЎвЂљР С•РЎРЏР Р…Р С‘Р Вµ Р Т‘Р С• Р С—Р ВµРЎР‚Р ВµРЎРѓР ВµРЎвЂЎР ВµР Р…Р С‘РЎРЏ Р Т‘Р Р†РЎС“РЎвЂ¦ Р С•РЎвЂљРЎР‚Р ВµР В·Р С”Р С•Р Р†.
func segmentIntersectionDistance(ax float64, ay float64, bx float64, by float64, cx float64, cy float64, dx float64, dy float64) (float64, bool) {
	rx := bx - ax
	ry := by - ay
	sx := dx - cx
	sy := dy - cy
	denominator := cross2D(rx, ry, sx, sy)
	if math.Abs(denominator) <= physics.Epsilon {
		return 0, false
	}
	qpx := cx - ax
	qpy := cy - ay
	t := cross2D(qpx, qpy, sx, sy) / denominator
	u := cross2D(qpx, qpy, rx, ry) / denominator
	if t < -physics.Epsilon || t > 1+physics.Epsilon || u < -physics.Epsilon || u > 1+physics.Epsilon {
		return 0, false
	}
	return math.Hypot(rx, ry) * math.Max(0, math.Min(1, t)), true
}

// cross2D РЎРѓРЎвЂЎР С‘РЎвЂљР В°Р ВµРЎвЂљ Р С—РЎРѓР ВµР Р†Р Т‘Р С•РЎРѓР С”Р В°Р В»РЎРЏРЎР‚Р Р…Р С•Р Вµ Р С—РЎР‚Р С•Р С‘Р В·Р Р†Р ВµР Т‘Р ВµР Р…Р С‘Р Вµ Р Т‘Р Р†РЎС“РЎвЂ¦ Р С—Р В»Р С•РЎРѓР С”Р С‘РЎвЂ¦ Р Р†Р ВµР С”РЎвЂљР С•РЎР‚Р С•Р Р†.
func cross2D(ax float64, ay float64, bx float64, by float64) float64 {
	return ax*by - ay*bx
}
