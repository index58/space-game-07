package data

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"sort"
	"time"
)

// Chat С…СЂР°РЅРёС‚ РїСЂРёРЅР°РґР»РµР¶РЅРѕСЃС‚СЊ РїРµСЂРµРїРёСЃРєРё Рє РёРіСЂРѕРІРѕРјСѓ СЃРѕРѕР±С‰РµСЃС‚РІСѓ.
type Chat struct {
	ID              int64  `json:"ID"`              // РЈРЅРёРєР°Р»СЊРЅС‹Р№ С‡РёСЃР»РѕРІРѕР№ РёРґРµРЅС‚РёС„РёРєР°С‚РѕСЂ Р·Р°РїРёСЃРё.
	CommunityTypeID int64  `json:"CommunityTypeID"` // РўРёРї СЃРѕРѕР±С‰РµСЃС‚РІР°, РІРЅСѓС‚СЂРё РєРѕС‚РѕСЂРѕРіРѕ СЃСѓС‰РµСЃС‚РІСѓРµС‚ РїРµСЂРµРїРёСЃРєР°.
	CommunityID     int64  `json:"CommunityID"`     // РРґРµРЅС‚РёС„РёРєР°С‚РѕСЂ СЃРѕРѕР±С‰РµСЃС‚РІР° СѓРєР°Р·Р°РЅРЅРѕРіРѕ С‚РёРїР°.
	DuoChatKey      string `json:"DuoChatKey"`      // РЎС‚Р°Р±РёР»СЊРЅС‹Р№ РєР»СЋС‡ Р»РёС‡РЅРѕР№ РїРµСЂРµРїРёСЃРєРё РґРІСѓС… РїРµСЂСЃРѕРЅР°Р¶РµР№.
}

// Chats С…СЂР°РЅРёС‚ С‡Р°С‚С‹ Рё Р±С‹СЃС‚СЂС‹Рµ РёРЅРґРµРєСЃС‹ РґР»СЏ РїРѕРёСЃРєР° РїРѕ СЃРѕРѕР±С‰РµСЃС‚РІСѓ Рё РґСѓСЌС‚РЅРѕРјСѓ РєР»СЋС‡Сѓ.
type Chats struct {
	MaxID int64           `json:"MaxID"` // РџРѕСЃР»РµРґРЅРёР№ РІС‹РґР°РЅРЅС‹Р№ С‡РёСЃР»РѕРІРѕР№ РёРґРµРЅС‚РёС„РёРєР°С‚РѕСЂ Р·Р°РїРёСЃРµР№.
	Items map[int64]*Chat `json:"Items"` // РћСЃРЅРѕРІРЅРѕРµ С…СЂР°РЅРёР»РёС‰Рµ Р·Р°РїРёСЃРµР№ РїРѕ С‡РёСЃР»РѕРІРѕРјСѓ РёРґРµРЅС‚РёС„РёРєР°С‚РѕСЂСѓ.

	ByCommunity map[string]*Chat `json:"-"` // Р‘С‹СЃС‚СЂС‹Р№ РїРѕРёСЃРє С‡Р°С‚Р° РїРѕ С‚РёРїСѓ Рё РёРґРµРЅС‚РёС„РёРєР°С‚РѕСЂСѓ СЃРѕРѕР±С‰РµСЃС‚РІР°.
	ByDuoKey    map[string]*Chat `json:"-"` // Р‘С‹СЃС‚СЂС‹Р№ РїРѕРёСЃРє Р»РёС‡РЅРѕР№ РїРµСЂРµРїРёСЃРєРё РїРѕ РєР»СЋС‡Сѓ РґРІСѓС… РїРµСЂСЃРѕРЅР°Р¶РµР№.
}

// ChatMember С…СЂР°РЅРёС‚ СѓС‡Р°СЃС‚РёРµ РїРµСЂСЃРѕРЅР°Р¶Р° РІ РєРѕРЅРєСЂРµС‚РЅРѕР№ РїРµСЂРµРїРёСЃРєРµ.
type ChatMember struct {
	ID                  int64 `json:"ID"`                  // РЈРЅРёРєР°Р»СЊРЅС‹Р№ С‡РёСЃР»РѕРІРѕР№ РёРґРµРЅС‚РёС„РёРєР°С‚РѕСЂ Р·Р°РїРёСЃРё.
	ChatID              int64 `json:"ChatID"`              // Р§Р°С‚, РІ РєРѕС‚РѕСЂРѕРј РїСЂРёСЃСѓС‚СЃС‚РІСѓРµС‚ РїРµСЂСЃРѕРЅР°Р¶.
	MemberCharacterID   int64 `json:"MemberCharacterID"`   // РџРµСЂСЃРѕРЅР°Р¶, РїСЂРёСЃСѓС‚СЃС‚РІСѓСЋС‰РёР№ РІ С‡Р°С‚Рµ.
	CommunityChatRoleID int64 `json:"CommunityChatRoleID"` // Р РѕР»СЊ РїРµСЂСЃРѕРЅР°Р¶Р° РІ С‡Р°С‚Рµ.
}

// ChatMembers С…СЂР°РЅРёС‚ СѓС‡Р°СЃС‚РЅРёРєРѕРІ С‡Р°С‚РѕРІ Рё Р±С‹СЃС‚СЂС‹Рµ РёРЅРґРµРєСЃС‹ С‡Р»РµРЅСЃС‚РІР°.
type ChatMembers struct {
	MaxID int64                 `json:"MaxID"` // РџРѕСЃР»РµРґРЅРёР№ РІС‹РґР°РЅРЅС‹Р№ С‡РёСЃР»РѕРІРѕР№ РёРґРµРЅС‚РёС„РёРєР°С‚РѕСЂ Р·Р°РїРёСЃРµР№.
	Items map[int64]*ChatMember `json:"Items"` // РћСЃРЅРѕРІРЅРѕРµ С…СЂР°РЅРёР»РёС‰Рµ Р·Р°РїРёСЃРµР№ РїРѕ С‡РёСЃР»РѕРІРѕРјСѓ РёРґРµРЅС‚РёС„РёРєР°С‚РѕСЂСѓ.

	ByChatID      map[int64][]*ChatMember `json:"-"` // Р‘С‹СЃС‚СЂС‹Р№ РїРѕРёСЃРє СѓС‡Р°СЃС‚РЅРёРєРѕРІ СѓРєР°Р·Р°РЅРЅРѕРіРѕ С‡Р°С‚Р°.
	ByCharacterID map[int64][]*ChatMember `json:"-"` // Р‘С‹СЃС‚СЂС‹Р№ РїРѕРёСЃРє С‡Р°С‚РѕРІ СѓРєР°Р·Р°РЅРЅРѕРіРѕ РїРµСЂСЃРѕРЅР°Р¶Р°.
	ByPair        map[string]*ChatMember  `json:"-"` // Р‘С‹СЃС‚СЂР°СЏ РїСЂРѕРІРµСЂРєР° СѓРЅРёРєР°Р»СЊРЅРѕСЃС‚Рё РїР°СЂС‹ С‡Р°С‚Р° Рё РїРµСЂСЃРѕРЅР°Р¶Р°.
}

// CommunityType С…СЂР°РЅРёС‚ С‚РёРї СЃРѕРѕР±С‰РµСЃС‚РІР°, Рє РєРѕС‚РѕСЂРѕРјСѓ РѕС‚РЅРѕСЃРёС‚СЃСЏ С‡Р°С‚.
type CommunityType struct {
	ID      int64  `json:"ID"`      // РЈРЅРёРєР°Р»СЊРЅС‹Р№ С‡РёСЃР»РѕРІРѕР№ РёРґРµРЅС‚РёС„РёРєР°С‚РѕСЂ Р·Р°РїРёСЃРё.
	TitleRu string `json:"TitleRu"` // Р СѓСЃСЃРєРѕРµ РЅР°Р·РІР°РЅРёРµ С‚РёРїР° СЃРѕРѕР±С‰РµСЃС‚РІР°.
	TitleEn string `json:"TitleEn"` // РђРЅРіР»РёР№СЃРєРѕРµ РЅР°Р·РІР°РЅРёРµ С‚РёРїР° СЃРѕРѕР±С‰РµСЃС‚РІР°.
	Acronym string `json:"Acronym"` // РќРµРёР·РјРµРЅСЏРµРјС‹Р№ СЃС‚СЂРѕРєРѕРІС‹Р№ РёРґРµРЅС‚РёС„РёРєР°С‚РѕСЂ С‚РёРїР° СЃРѕРѕР±С‰РµСЃС‚РІР°.
}

// CommunityTypes С…СЂР°РЅРёС‚ С‚РёРїС‹ СЃРѕРѕР±С‰РµСЃС‚РІ Рё РёРЅРґРµРєСЃС‹ СѓРЅРёРєР°Р»СЊРЅС‹С… СЃС‚СЂРѕРє.
type CommunityTypes struct {
	MaxID int64                    `json:"MaxID"` // РџРѕСЃР»РµРґРЅРёР№ РІС‹РґР°РЅРЅС‹Р№ С‡РёСЃР»РѕРІРѕР№ РёРґРµРЅС‚РёС„РёРєР°С‚РѕСЂ Р·Р°РїРёСЃРµР№.
	Items map[int64]*CommunityType `json:"Items"` // РћСЃРЅРѕРІРЅРѕРµ С…СЂР°РЅРёР»РёС‰Рµ Р·Р°РїРёСЃРµР№ РїРѕ С‡РёСЃР»РѕРІРѕРјСѓ РёРґРµРЅС‚РёС„РёРєР°С‚РѕСЂСѓ.

	ByAcronym map[string]*CommunityType `json:"-"` // Р‘С‹СЃС‚СЂС‹Р№ РїРѕРёСЃРє Р·Р°РїРёСЃРё РїРѕ Р°РєСЂРѕРЅРёРјСѓ.
}

// CommunityChatRole С…СЂР°РЅРёС‚ СЂРѕР»СЊ РїРµСЂСЃРѕРЅР°Р¶Р° РІРЅСѓС‚СЂРё С‡Р°С‚Р° СЃРѕРѕР±С‰РµСЃС‚РІР°.
type CommunityChatRole struct {
	ID              int64  `json:"ID"`              // РЈРЅРёРєР°Р»СЊРЅС‹Р№ С‡РёСЃР»РѕРІРѕР№ РёРґРµРЅС‚РёС„РёРєР°С‚РѕСЂ Р·Р°РїРёСЃРё.
	CommunityTypeID int64  `json:"CommunityTypeID"` // РўРёРї СЃРѕРѕР±С‰РµСЃС‚РІР°, РґР»СЏ РєРѕС‚РѕСЂРѕРіРѕ Р·Р°РґР°РЅР° СЂРѕР»СЊ.
	TitleRu         string `json:"TitleRu"`         // Р СѓСЃСЃРєРѕРµ РЅР°Р·РІР°РЅРёРµ СЂРѕР»Рё.
	TitleEn         string `json:"TitleEn"`         // РђРЅРіР»РёР№СЃРєРѕРµ РЅР°Р·РІР°РЅРёРµ СЂРѕР»Рё.
	Acronym         string `json:"Acronym"`         // РќРµРёР·РјРµРЅСЏРµРјС‹Р№ СЃС‚СЂРѕРєРѕРІС‹Р№ РёРґРµРЅС‚РёС„РёРєР°С‚РѕСЂ СЂРѕР»Рё.
}

// CommunityChatRoles С…СЂР°РЅРёС‚ СЂРѕР»Рё С‡Р°С‚РѕРІ Рё РёРЅРґРµРєСЃ РїРѕ С‚РёРїСѓ СЃРѕРѕР±С‰РµСЃС‚РІР° СЃ Р°РєСЂРѕРЅРёРјРѕРј.
type CommunityChatRoles struct {
	MaxID int64                        `json:"MaxID"` // РџРѕСЃР»РµРґРЅРёР№ РІС‹РґР°РЅРЅС‹Р№ С‡РёСЃР»РѕРІРѕР№ РёРґРµРЅС‚РёС„РёРєР°С‚РѕСЂ Р·Р°РїРёСЃРµР№.
	Items map[int64]*CommunityChatRole `json:"Items"` // РћСЃРЅРѕРІРЅРѕРµ С…СЂР°РЅРёР»РёС‰Рµ Р·Р°РїРёСЃРµР№ РїРѕ С‡РёСЃР»РѕРІРѕРјСѓ РёРґРµРЅС‚РёС„РёРєР°С‚РѕСЂСѓ.

	ByTypeAndAcronym map[string]*CommunityChatRole `json:"-"` // Р‘С‹СЃС‚СЂС‹Р№ РїРѕРёСЃРє СЂРѕР»Рё РїРѕ С‚РёРїСѓ СЃРѕРѕР±С‰РµСЃС‚РІР° Рё Р°РєСЂРѕРЅРёРјСѓ.
}

// Message С…СЂР°РЅРёС‚ РѕРґРЅРѕ СЃРѕРѕР±С‰РµРЅРёРµ РёРіСЂРѕРІРѕРіРѕ С‡Р°С‚Р°.
type Message struct {
	ID                int64     `json:"ID"`                // РЈРЅРёРєР°Р»СЊРЅС‹Р№ С‡РёСЃР»РѕРІРѕР№ РёРґРµРЅС‚РёС„РёРєР°С‚РѕСЂ Р·Р°РїРёСЃРё.
	ChatID            int64     `json:"ChatID"`            // Р§Р°С‚, РІ РєРѕС‚РѕСЂРѕРј РїРѕСЏРІРёР»РѕСЃСЊ СЃРѕРѕР±С‰РµРЅРёРµ.
	SenderCharacterID int64     `json:"SenderCharacterID"` // РџРµСЂСЃРѕРЅР°Р¶, РѕС‚ РёРјРµРЅРё РєРѕС‚РѕСЂРѕРіРѕ РѕС‚РїСЂР°РІР»РµРЅРѕ СЃРѕРѕР±С‰РµРЅРёРµ.
	MessageTypeID     int64     `json:"MessageTypeID"`     // РўРёРї СЃРѕРѕР±С‰РµРЅРёСЏ, РІР»РёСЏСЋС‰РёР№ РЅР° РІРЅРµС€РЅРёР№ РІРёРґ.
	Text              string    `json:"Text"`              // РўРµРєСЃС‚ СЃРѕРѕР±С‰РµРЅРёСЏ.
	Color             string    `json:"Color"`             // Р¦РІРµС‚ СЃРѕРѕР±С‰РµРЅРёСЏ РІ RGB-HEX С„РѕСЂРјР°С‚Рµ Р±РµР· СЂРµС€РµС‚РєРё.
	SentTime          time.Time `json:"SentTime"`          // РњРѕРјРµРЅС‚ РѕС‚РїСЂР°РІРєРё СЃРѕРѕР±С‰РµРЅРёСЏ.
}

// Messages С…СЂР°РЅРёС‚ СЃРѕРѕР±С‰РµРЅРёСЏ Рё РёРЅРґРµРєСЃ РїРѕ С‡Р°С‚Сѓ.
type Messages struct {
	MaxID int64              `json:"MaxID"` // РџРѕСЃР»РµРґРЅРёР№ РІС‹РґР°РЅРЅС‹Р№ С‡РёСЃР»РѕРІРѕР№ РёРґРµРЅС‚РёС„РёРєР°С‚РѕСЂ Р·Р°РїРёСЃРµР№.
	Items map[int64]*Message `json:"Items"` // РћСЃРЅРѕРІРЅРѕРµ С…СЂР°РЅРёР»РёС‰Рµ Р·Р°РїРёСЃРµР№ РїРѕ С‡РёСЃР»РѕРІРѕРјСѓ РёРґРµРЅС‚РёС„РёРєР°С‚РѕСЂСѓ.

	ByChatID map[int64][]*Message `json:"-"` // Р‘С‹СЃС‚СЂС‹Р№ РїРѕРёСЃРє СЃРѕРѕР±С‰РµРЅРёР№ СѓРєР°Р·Р°РЅРЅРѕРіРѕ С‡Р°С‚Р°.
}

// MessageRead С…СЂР°РЅРёС‚ РїРѕСЃР»РµРґРЅСЋСЋ РїСЂРѕС‡РёС‚Р°РЅРЅСѓСЋ СЃС‚СЂРѕРєСѓ РєРѕРЅРєСЂРµС‚РЅРѕРіРѕ РїРµСЂСЃРѕРЅР°Р¶Р° РІ РєРѕРЅРєСЂРµС‚РЅРѕРј С‡Р°С‚Рµ.
type MessageRead struct {
	ID                int64     `json:"ID"`                // РЈРЅРёРєР°Р»СЊРЅС‹Р№ С‡РёСЃР»РѕРІРѕР№ РёРґРµРЅС‚РёС„РёРєР°С‚РѕСЂ Р·Р°РїРёСЃРё.
	CharacterID       int64     `json:"CharacterID"`       // РџРµСЂСЃРѕРЅР°Р¶, РґР»СЏ РєРѕС‚РѕСЂРѕРіРѕ СЃРѕС…СЂР°РЅРµРЅР° РїРѕР·РёС†РёСЏ С‡С‚РµРЅРёСЏ.
	ChatID            int64     `json:"ChatID"`            // Р§Р°С‚, РІРЅСѓС‚СЂРё РєРѕС‚РѕСЂРѕРіРѕ СЃРѕС…СЂР°РЅРµРЅР° РїРѕР·РёС†РёСЏ С‡С‚РµРЅРёСЏ.
	LastReadMessageID int64     `json:"LastReadMessageID"` // РџРѕСЃР»РµРґРЅРµРµ РїСЂРѕС‡РёС‚Р°РЅРЅРѕРµ СЃРѕРѕР±С‰РµРЅРёРµ СЌС‚РѕРіРѕ С‡Р°С‚Р°.
	ReadTime          time.Time `json:"ReadTime"`          // РњРѕРјРµРЅС‚ РѕР±РЅРѕРІР»РµРЅРёСЏ РїРѕР·РёС†РёРё С‡С‚РµРЅРёСЏ.
}

// MessageReads С…СЂР°РЅРёС‚ РїРѕР·РёС†РёРё С‡С‚РµРЅРёСЏ Рё РёРЅРґРµРєСЃРёСЂСѓРµС‚ РёС… РїРѕ РїР°СЂРµ РїРµСЂСЃРѕРЅР°Р¶Р° Рё С‡Р°С‚Р°.
type MessageReads struct {
	MaxID int64                  `json:"MaxID"` // РџРѕСЃР»РµРґРЅРёР№ РІС‹РґР°РЅРЅС‹Р№ С‡РёСЃР»РѕРІРѕР№ РёРґРµРЅС‚РёС„РёРєР°С‚РѕСЂ Р·Р°РїРёСЃРµР№.
	Items map[int64]*MessageRead `json:"Items"` // РћСЃРЅРѕРІРЅРѕРµ С…СЂР°РЅРёР»РёС‰Рµ Р·Р°РїРёСЃРµР№ РїРѕ С‡РёСЃР»РѕРІРѕРјСѓ РёРґРµРЅС‚РёС„РёРєР°С‚РѕСЂСѓ.

	ByCharacterAndChat map[string]*MessageRead `json:"-"` // Р‘С‹СЃС‚СЂС‹Р№ РїРѕРёСЃРє РїРѕР·РёС†РёРё С‡С‚РµРЅРёСЏ РїРѕ РїРµСЂСЃРѕРЅР°Р¶Сѓ Рё С‡Р°С‚Сѓ.
}

// MessageType С…СЂР°РЅРёС‚ С‚РёРї СЃРѕРѕР±С‰РµРЅРёСЏ РґР»СЏ РІРёР·СѓР°Р»СЊРЅРѕРіРѕ РѕС„РѕСЂРјР»РµРЅРёСЏ.
type MessageType struct {
	ID      int64  `json:"ID"`      // РЈРЅРёРєР°Р»СЊРЅС‹Р№ С‡РёСЃР»РѕРІРѕР№ РёРґРµРЅС‚РёС„РёРєР°С‚РѕСЂ Р·Р°РїРёСЃРё.
	TitleRu string `json:"TitleRu"` // Р СѓСЃСЃРєРѕРµ РЅР°Р·РІР°РЅРёРµ С‚РёРїР° СЃРѕРѕР±С‰РµРЅРёСЏ.
	TitleEn string `json:"TitleEn"` // РђРЅРіР»РёР№СЃРєРѕРµ РЅР°Р·РІР°РЅРёРµ С‚РёРїР° СЃРѕРѕР±С‰РµРЅРёСЏ.
	Acronym string `json:"Acronym"` // РќРµРёР·РјРµРЅСЏРµРјС‹Р№ СЃС‚СЂРѕРєРѕРІС‹Р№ РёРґРµРЅС‚РёС„РёРєР°С‚РѕСЂ С‚РёРїР° СЃРѕРѕР±С‰РµРЅРёСЏ.
}

// MessageTypes С…СЂР°РЅРёС‚ С‚РёРїС‹ СЃРѕРѕР±С‰РµРЅРёР№ Рё РёРЅРґРµРєСЃ РїРѕ Р°РєСЂРѕРЅРёРјСѓ.
type MessageTypes struct {
	MaxID int64                  `json:"MaxID"` // РџРѕСЃР»РµРґРЅРёР№ РІС‹РґР°РЅРЅС‹Р№ С‡РёСЃР»РѕРІРѕР№ РёРґРµРЅС‚РёС„РёРєР°С‚РѕСЂ Р·Р°РїРёСЃРµР№.
	Items map[int64]*MessageType `json:"Items"` // РћСЃРЅРѕРІРЅРѕРµ С…СЂР°РЅРёР»РёС‰Рµ Р·Р°РїРёСЃРµР№ РїРѕ С‡РёСЃР»РѕРІРѕРјСѓ РёРґРµРЅС‚РёС„РёРєР°С‚РѕСЂСѓ.

	ByAcronym map[string]*MessageType `json:"-"` // Р‘С‹СЃС‚СЂС‹Р№ РїРѕРёСЃРє Р·Р°РїРёСЃРё РїРѕ Р°РєСЂРѕРЅРёРјСѓ.
}

func NewChats() *Chats {
	chats := &Chats{}
	chats.ensureMaps()
	return chats
}

func NewChatMembers() *ChatMembers {
	members := &ChatMembers{}
	members.ensureMaps()
	return members
}

func NewCommunityTypes() *CommunityTypes {
	types := &CommunityTypes{}
	types.ensureMaps()
	return types
}

func NewCommunityChatRoles() *CommunityChatRoles {
	roles := &CommunityChatRoles{}
	roles.ensureMaps()
	return roles
}

func NewMessages() *Messages {
	messages := &Messages{}
	messages.ensureMaps()
	return messages
}

func NewMessageReads() *MessageReads {
	reads := &MessageReads{}
	reads.ensureMaps()
	return reads
}

func NewMessageTypes() *MessageTypes {
	types := &MessageTypes{}
	types.ensureMaps()
	return types
}

func (chats *Chats) Add(chat *Chat) (*Chat, error) {
	if chat == nil {
		return nil, errors.New("chat is nil")
	}
	chats.ensureMaps()
	if chat.CommunityTypeID <= 0 {
		return nil, errors.New("community type ID is empty")
	}
	if existing := chats.ByCommunity[chatCommunityKey(chat.CommunityTypeID, chat.CommunityID)]; existing != nil && existing.ID != chat.ID {
		return nil, errors.New("chat community is not unique")
	}
	if chat.DuoChatKey != "" {
		if existing := chats.ByDuoKey[chat.DuoChatKey]; existing != nil && existing.ID != chat.ID {
			return nil, errors.New("duo chat key is not unique")
		}
	}
	chats.MaxID++
	chat.ID = chats.MaxID
	chats.Items[chat.ID] = chat
	chats.addIndexes(chat)
	return chat, nil
}

func (chats *Chats) Get(id int64) (*Chat, bool) {
	chats.ensureMaps()
	chat, ok := chats.Items[id]
	return chat, ok
}

func (chats *Chats) GetByCommunity(communityTypeID int64, communityID int64) (*Chat, bool) {
	chats.ensureMaps()
	chat, ok := chats.ByCommunity[chatCommunityKey(communityTypeID, communityID)]
	return chat, ok
}

func (chats *Chats) GetByDuoKey(duoChatKey string) (*Chat, bool) {
	chats.ensureMaps()
	chat, ok := chats.ByDuoKey[duoChatKey]
	return chat, ok
}

func (chats *Chats) RebuildIndexes() error {
	chats.ensureItems()
	chats.ByCommunity = map[string]*Chat{}
	chats.ByDuoKey = map[string]*Chat{}
	var maxID int64
	for _, id := range sortedTableItemIDs(chats.Items) {
		chat := chats.Items[id]
		if chat == nil {
			return fmt.Errorf("chat with ID %d is nil", id)
		}
		if chat.ID != id {
			return fmt.Errorf("chat map key %d does not match chat ID %d", id, chat.ID)
		}
		if chat.CommunityTypeID <= 0 {
			return fmt.Errorf("chat with ID %d has empty community type ID", id)
		}
		if existing := chats.ByCommunity[chatCommunityKey(chat.CommunityTypeID, chat.CommunityID)]; existing != nil && existing.ID != chat.ID {
			return errors.New("chat community is not unique")
		}
		if chat.DuoChatKey != "" {
			if existing := chats.ByDuoKey[chat.DuoChatKey]; existing != nil && existing.ID != chat.ID {
				return errors.New("duo chat key is not unique")
			}
		}
		if id > maxID {
			maxID = id
		}
		chats.addIndexes(chat)
	}
	if chats.MaxID < maxID {
		chats.MaxID = maxID
	}
	return nil
}

func (chats *Chats) LoadFromFile(path string) error {
	content, err := os.ReadFile(path)
	if err != nil {
		return err
	}
	loaded := Chats{}
	if err := json.Unmarshal(content, &loaded); err != nil {
		return err
	}
	if err := loaded.RebuildIndexes(); err != nil {
		return err
	}
	*chats = loaded
	return nil
}

func (chats *Chats) SaveToFile(path string) error {
	chats.ensureMaps()
	return saveTableWithOrderedItems(path, chats.MaxID, chats.Items)
}

func (chats *Chats) ensureMaps() {
	chats.ensureItems()
	if chats.ByCommunity == nil {
		chats.ByCommunity = map[string]*Chat{}
	}
	if chats.ByDuoKey == nil {
		chats.ByDuoKey = map[string]*Chat{}
	}
}

func (chats *Chats) ensureItems() {
	if chats.Items == nil {
		chats.Items = map[int64]*Chat{}
	}
}

func (chats *Chats) addIndexes(chat *Chat) {
	chats.ByCommunity[chatCommunityKey(chat.CommunityTypeID, chat.CommunityID)] = chat
	if chat.DuoChatKey != "" {
		chats.ByDuoKey[chat.DuoChatKey] = chat
	}
}

func (members *ChatMembers) Add(member *ChatMember) (*ChatMember, error) {
	if member == nil {
		return nil, errors.New("chat member is nil")
	}
	members.ensureMaps()
	if member.ChatID <= 0 {
		return nil, errors.New("chat ID is empty")
	}
	if member.MemberCharacterID <= 0 {
		return nil, errors.New("member character ID is empty")
	}
	pairKey := chatMemberPairKey(member.ChatID, member.MemberCharacterID)
	if existing := members.ByPair[pairKey]; existing != nil && existing.ID != member.ID {
		return nil, errors.New("chat member pair is not unique")
	}
	members.MaxID++
	member.ID = members.MaxID
	members.Items[member.ID] = member
	members.addIndexes(member)
	return member, nil
}

func (members *ChatMembers) GetByChatID(chatID int64) []*ChatMember {
	members.ensureMaps()
	return append([]*ChatMember(nil), members.ByChatID[chatID]...)
}

func (members *ChatMembers) GetByCharacterID(characterID int64) []*ChatMember {
	members.ensureMaps()
	return append([]*ChatMember(nil), members.ByCharacterID[characterID]...)
}

func (members *ChatMembers) HasMember(chatID int64, characterID int64) bool {
	members.ensureMaps()
	return members.ByPair[chatMemberPairKey(chatID, characterID)] != nil
}

func (members *ChatMembers) RebuildIndexes() error {
	members.ensureItems()
	members.ByChatID = map[int64][]*ChatMember{}
	members.ByCharacterID = map[int64][]*ChatMember{}
	members.ByPair = map[string]*ChatMember{}
	var maxID int64
	for _, id := range sortedTableItemIDs(members.Items) {
		member := members.Items[id]
		if member == nil {
			return fmt.Errorf("chat member with ID %d is nil", id)
		}
		if member.ID != id {
			return fmt.Errorf("chat member map key %d does not match member ID %d", id, member.ID)
		}
		if member.ChatID <= 0 || member.MemberCharacterID <= 0 {
			return fmt.Errorf("chat member with ID %d has empty required fields", id)
		}
		pairKey := chatMemberPairKey(member.ChatID, member.MemberCharacterID)
		if existing := members.ByPair[pairKey]; existing != nil && existing.ID != member.ID {
			return errors.New("chat member pair is not unique")
		}
		if id > maxID {
			maxID = id
		}
		members.addIndexes(member)
	}
	if members.MaxID < maxID {
		members.MaxID = maxID
	}
	return nil
}

func (members *ChatMembers) LoadFromFile(path string) error {
	content, err := os.ReadFile(path)
	if err != nil {
		return err
	}
	loaded := ChatMembers{}
	if err := json.Unmarshal(content, &loaded); err != nil {
		return err
	}
	if err := loaded.RebuildIndexes(); err != nil {
		return err
	}
	*members = loaded
	return nil
}

func (members *ChatMembers) SaveToFile(path string) error {
	members.ensureMaps()
	return saveTableWithOrderedItems(path, members.MaxID, members.Items)
}

func (members *ChatMembers) ensureMaps() {
	members.ensureItems()
	if members.ByChatID == nil {
		members.ByChatID = map[int64][]*ChatMember{}
	}
	if members.ByCharacterID == nil {
		members.ByCharacterID = map[int64][]*ChatMember{}
	}
	if members.ByPair == nil {
		members.ByPair = map[string]*ChatMember{}
	}
}

func (members *ChatMembers) ensureItems() {
	if members.Items == nil {
		members.Items = map[int64]*ChatMember{}
	}
}

func (members *ChatMembers) addIndexes(member *ChatMember) {
	members.ByChatID[member.ChatID] = append(members.ByChatID[member.ChatID], member)
	members.ByCharacterID[member.MemberCharacterID] = append(members.ByCharacterID[member.MemberCharacterID], member)
	members.ByPair[chatMemberPairKey(member.ChatID, member.MemberCharacterID)] = member
}

func (types *CommunityTypes) Add(communityType *CommunityType) (*CommunityType, error) {
	if communityType == nil {
		return nil, errors.New("community type is nil")
	}
	types.ensureMaps()
	if communityType.TitleRu == "" || communityType.TitleEn == "" || communityType.Acronym == "" {
		return nil, errors.New("community type has empty required fields")
	}
	if existing := types.ByAcronym[communityType.Acronym]; existing != nil && existing.ID != communityType.ID {
		return nil, errors.New("community type acronym is not unique")
	}
	types.MaxID++
	communityType.ID = types.MaxID
	types.Items[communityType.ID] = communityType
	types.addIndexes(communityType)
	return communityType, nil
}

func (types *CommunityTypes) Get(id int64) (*CommunityType, bool) {
	types.ensureMaps()
	communityType, ok := types.Items[id]
	return communityType, ok
}

func (types *CommunityTypes) GetByAcronym(acronym string) (*CommunityType, bool) {
	types.ensureMaps()
	communityType, ok := types.ByAcronym[acronym]
	return communityType, ok
}

func (types *CommunityTypes) RebuildIndexes() error {
	types.ensureItems()
	types.ByAcronym = map[string]*CommunityType{}
	var maxID int64
	for _, id := range sortedTableItemIDs(types.Items) {
		item := types.Items[id]
		if item == nil {
			return fmt.Errorf("community type with ID %d is nil", id)
		}
		if item.ID != id || item.TitleRu == "" || item.TitleEn == "" || item.Acronym == "" {
			return fmt.Errorf("community type with ID %d is invalid", id)
		}
		if existing := types.ByAcronym[item.Acronym]; existing != nil && existing.ID != item.ID {
			return errors.New("community type acronym is not unique")
		}
		if id > maxID {
			maxID = id
		}
		types.addIndexes(item)
	}
	if types.MaxID < maxID {
		types.MaxID = maxID
	}
	return nil
}

func (types *CommunityTypes) LoadFromFile(path string) error {
	content, err := os.ReadFile(path)
	if err != nil {
		return err
	}
	loaded := CommunityTypes{}
	if err := json.Unmarshal(content, &loaded); err != nil {
		return err
	}
	if err := loaded.RebuildIndexes(); err != nil {
		return err
	}
	*types = loaded
	return nil
}

func (types *CommunityTypes) SaveToFile(path string) error {
	types.ensureMaps()
	return saveTableWithOrderedItems(path, types.MaxID, types.Items)
}

func (types *CommunityTypes) ensureMaps() {
	types.ensureItems()
	if types.ByAcronym == nil {
		types.ByAcronym = map[string]*CommunityType{}
	}
}

func (types *CommunityTypes) ensureItems() {
	if types.Items == nil {
		types.Items = map[int64]*CommunityType{}
	}
}

func (types *CommunityTypes) addIndexes(communityType *CommunityType) {
	types.ByAcronym[communityType.Acronym] = communityType
}

func (roles *CommunityChatRoles) Add(role *CommunityChatRole) (*CommunityChatRole, error) {
	if role == nil {
		return nil, errors.New("community chat role is nil")
	}
	roles.ensureMaps()
	if role.CommunityTypeID <= 0 || role.TitleRu == "" || role.TitleEn == "" || role.Acronym == "" {
		return nil, errors.New("community chat role has empty required fields")
	}
	key := communityRoleKey(role.CommunityTypeID, role.Acronym)
	if existing := roles.ByTypeAndAcronym[key]; existing != nil && existing.ID != role.ID {
		return nil, errors.New("community chat role is not unique")
	}
	roles.MaxID++
	role.ID = roles.MaxID
	roles.Items[role.ID] = role
	roles.addIndexes(role)
	return role, nil
}

func (roles *CommunityChatRoles) GetByTypeAndAcronym(communityTypeID int64, acronym string) (*CommunityChatRole, bool) {
	roles.ensureMaps()
	role, ok := roles.ByTypeAndAcronym[communityRoleKey(communityTypeID, acronym)]
	return role, ok
}

func (roles *CommunityChatRoles) RebuildIndexes() error {
	roles.ensureItems()
	roles.ByTypeAndAcronym = map[string]*CommunityChatRole{}
	var maxID int64
	for _, id := range sortedTableItemIDs(roles.Items) {
		role := roles.Items[id]
		if role == nil {
			return fmt.Errorf("community chat role with ID %d is nil", id)
		}
		if role.ID != id || role.CommunityTypeID <= 0 || role.TitleRu == "" || role.TitleEn == "" || role.Acronym == "" {
			return fmt.Errorf("community chat role with ID %d is invalid", id)
		}
		key := communityRoleKey(role.CommunityTypeID, role.Acronym)
		if existing := roles.ByTypeAndAcronym[key]; existing != nil && existing.ID != role.ID {
			return errors.New("community chat role is not unique")
		}
		if id > maxID {
			maxID = id
		}
		roles.addIndexes(role)
	}
	if roles.MaxID < maxID {
		roles.MaxID = maxID
	}
	return nil
}

func (roles *CommunityChatRoles) LoadFromFile(path string) error {
	content, err := os.ReadFile(path)
	if err != nil {
		return err
	}
	loaded := CommunityChatRoles{}
	if err := json.Unmarshal(content, &loaded); err != nil {
		return err
	}
	if err := loaded.RebuildIndexes(); err != nil {
		return err
	}
	*roles = loaded
	return nil
}

func (roles *CommunityChatRoles) SaveToFile(path string) error {
	roles.ensureMaps()
	return saveTableWithOrderedItems(path, roles.MaxID, roles.Items)
}

func (roles *CommunityChatRoles) ensureMaps() {
	roles.ensureItems()
	if roles.ByTypeAndAcronym == nil {
		roles.ByTypeAndAcronym = map[string]*CommunityChatRole{}
	}
}

func (roles *CommunityChatRoles) ensureItems() {
	if roles.Items == nil {
		roles.Items = map[int64]*CommunityChatRole{}
	}
}

func (roles *CommunityChatRoles) addIndexes(role *CommunityChatRole) {
	roles.ByTypeAndAcronym[communityRoleKey(role.CommunityTypeID, role.Acronym)] = role
}

func (messages *Messages) Add(message *Message) (*Message, error) {
	if message == nil {
		return nil, errors.New("message is nil")
	}
	messages.ensureMaps()
	if message.ChatID <= 0 {
		return nil, errors.New("chat ID is empty")
	}
	if message.MessageTypeID <= 0 {
		return nil, errors.New("message type ID is empty")
	}
	if message.Text == "" {
		return nil, errors.New("text is empty")
	}
	messages.MaxID++
	message.ID = messages.MaxID
	if message.SentTime.IsZero() {
		message.SentTime = time.Now()
	}
	messages.Items[message.ID] = message
	messages.addIndexes(message)
	return message, nil
}

func (messages *Messages) GetByChatID(chatID int64) []*Message {
	messages.ensureMaps()
	result := append([]*Message(nil), messages.ByChatID[chatID]...)
	sort.Slice(result, func(left int, right int) bool {
		return result[left].ID < result[right].ID
	})
	return result
}

func (messages *Messages) RebuildIndexes() error {
	messages.ensureItems()
	messages.ByChatID = map[int64][]*Message{}
	var maxID int64
	for _, id := range sortedTableItemIDs(messages.Items) {
		message := messages.Items[id]
		if message == nil {
			return fmt.Errorf("message with ID %d is nil", id)
		}
		if message.ID != id || message.ChatID <= 0 || message.MessageTypeID <= 0 || message.Text == "" {
			return fmt.Errorf("message with ID %d is invalid", id)
		}
		if id > maxID {
			maxID = id
		}
		messages.addIndexes(message)
	}
	if messages.MaxID < maxID {
		messages.MaxID = maxID
	}
	return nil
}

func (messages *Messages) LoadFromFile(path string) error {
	content, err := os.ReadFile(path)
	if err != nil {
		return err
	}
	loaded := Messages{}
	if err := json.Unmarshal(content, &loaded); err != nil {
		return err
	}
	if err := loaded.RebuildIndexes(); err != nil {
		return err
	}
	*messages = loaded
	return nil
}

func (messages *Messages) SaveToFile(path string) error {
	messages.ensureMaps()
	return saveTableWithOrderedItems(path, messages.MaxID, messages.Items)
}

func (messages *Messages) ensureMaps() {
	messages.ensureItems()
	if messages.ByChatID == nil {
		messages.ByChatID = map[int64][]*Message{}
	}
}

func (messages *Messages) ensureItems() {
	if messages.Items == nil {
		messages.Items = map[int64]*Message{}
	}
}

func (messages *Messages) addIndexes(message *Message) {
	messages.ByChatID[message.ChatID] = append(messages.ByChatID[message.ChatID], message)
}

func (reads *MessageReads) SetLastRead(characterID int64, chatID int64, messageID int64) (*MessageRead, error) {
	if characterID <= 0 {
		return nil, errors.New("character ID is empty")
	}
	if chatID <= 0 {
		return nil, errors.New("chat ID is empty")
	}
	reads.ensureMaps()
	if existing := reads.ByCharacterAndChat[messageReadKey(characterID, chatID)]; existing != nil {
		if messageID > existing.LastReadMessageID {
			existing.LastReadMessageID = messageID
		}
		existing.ReadTime = time.Now()
		return existing, nil
	}
	reads.MaxID++
	messageRead := &MessageRead{
		ID:                reads.MaxID,
		CharacterID:       characterID,
		ChatID:            chatID,
		LastReadMessageID: messageID,
		ReadTime:          time.Now(),
	}
	reads.Items[messageRead.ID] = messageRead
	reads.addIndexes(messageRead)
	return messageRead, nil
}

func (reads *MessageReads) GetByCharacterAndChat(characterID int64, chatID int64) (*MessageRead, bool) {
	reads.ensureMaps()
	messageRead, ok := reads.ByCharacterAndChat[messageReadKey(characterID, chatID)]
	return messageRead, ok
}

func (reads *MessageReads) RebuildIndexes() error {
	reads.ensureItems()
	reads.ByCharacterAndChat = map[string]*MessageRead{}
	var maxID int64
	for _, id := range sortedTableItemIDs(reads.Items) {
		messageRead := reads.Items[id]
		if messageRead == nil {
			return fmt.Errorf("message read with ID %d is nil", id)
		}
		if messageRead.ID != id || messageRead.CharacterID <= 0 || messageRead.ChatID <= 0 {
			return fmt.Errorf("message read with ID %d is invalid", id)
		}
		if existing := reads.ByCharacterAndChat[messageReadKey(messageRead.CharacterID, messageRead.ChatID)]; existing != nil && existing.ID != messageRead.ID {
			return fmt.Errorf("message read with ID %d is not unique", id)
		}
		if id > maxID {
			maxID = id
		}
		reads.addIndexes(messageRead)
	}
	if reads.MaxID < maxID {
		reads.MaxID = maxID
	}
	return nil
}

func (reads *MessageReads) LoadFromFile(path string) error {
	content, err := os.ReadFile(path)
	if err != nil {
		return err
	}
	loaded := MessageReads{}
	if err := json.Unmarshal(content, &loaded); err != nil {
		return err
	}
	if err := loaded.RebuildIndexes(); err != nil {
		return err
	}
	*reads = loaded
	return nil
}

func (reads *MessageReads) SaveToFile(path string) error {
	reads.ensureMaps()
	return saveTableWithOrderedItems(path, reads.MaxID, reads.Items)
}

func (reads *MessageReads) ensureMaps() {
	reads.ensureItems()
	if reads.ByCharacterAndChat == nil {
		reads.ByCharacterAndChat = map[string]*MessageRead{}
	}
}

func (reads *MessageReads) ensureItems() {
	if reads.Items == nil {
		reads.Items = map[int64]*MessageRead{}
	}
}

func (reads *MessageReads) addIndexes(messageRead *MessageRead) {
	reads.ByCharacterAndChat[messageReadKey(messageRead.CharacterID, messageRead.ChatID)] = messageRead
}

func (types *MessageTypes) Add(messageType *MessageType) (*MessageType, error) {
	if messageType == nil {
		return nil, errors.New("message type is nil")
	}
	types.ensureMaps()
	if messageType.TitleRu == "" || messageType.TitleEn == "" || messageType.Acronym == "" {
		return nil, errors.New("message type has empty required fields")
	}
	if existing := types.ByAcronym[messageType.Acronym]; existing != nil && existing.ID != messageType.ID {
		return nil, errors.New("message type acronym is not unique")
	}
	types.MaxID++
	messageType.ID = types.MaxID
	types.Items[messageType.ID] = messageType
	types.addIndexes(messageType)
	return messageType, nil
}

func (types *MessageTypes) Get(id int64) (*MessageType, bool) {
	types.ensureMaps()
	messageType, ok := types.Items[id]
	return messageType, ok
}

func (types *MessageTypes) GetByAcronym(acronym string) (*MessageType, bool) {
	types.ensureMaps()
	messageType, ok := types.ByAcronym[acronym]
	return messageType, ok
}

func (types *MessageTypes) RebuildIndexes() error {
	types.ensureItems()
	types.ByAcronym = map[string]*MessageType{}
	var maxID int64
	for _, id := range sortedTableItemIDs(types.Items) {
		item := types.Items[id]
		if item == nil {
			return fmt.Errorf("message type with ID %d is nil", id)
		}
		if item.ID != id || item.TitleRu == "" || item.TitleEn == "" || item.Acronym == "" {
			return fmt.Errorf("message type with ID %d is invalid", id)
		}
		if existing := types.ByAcronym[item.Acronym]; existing != nil && existing.ID != item.ID {
			return errors.New("message type acronym is not unique")
		}
		if id > maxID {
			maxID = id
		}
		types.addIndexes(item)
	}
	if types.MaxID < maxID {
		types.MaxID = maxID
	}
	return nil
}

func (types *MessageTypes) LoadFromFile(path string) error {
	content, err := os.ReadFile(path)
	if err != nil {
		return err
	}
	loaded := MessageTypes{}
	if err := json.Unmarshal(content, &loaded); err != nil {
		return err
	}
	if err := loaded.RebuildIndexes(); err != nil {
		return err
	}
	*types = loaded
	return nil
}

func (types *MessageTypes) SaveToFile(path string) error {
	types.ensureMaps()
	return saveTableWithOrderedItems(path, types.MaxID, types.Items)
}

func (types *MessageTypes) ensureMaps() {
	types.ensureItems()
	if types.ByAcronym == nil {
		types.ByAcronym = map[string]*MessageType{}
	}
}

func (types *MessageTypes) ensureItems() {
	if types.Items == nil {
		types.Items = map[int64]*MessageType{}
	}
}

func (types *MessageTypes) addIndexes(messageType *MessageType) {
	types.ByAcronym[messageType.Acronym] = messageType
}

func chatCommunityKey(communityTypeID int64, communityID int64) string {
	return fmt.Sprintf("%d:%d", communityTypeID, communityID)
}

func chatMemberPairKey(chatID int64, characterID int64) string {
	return fmt.Sprintf("%d:%d", chatID, characterID)
}

func messageReadKey(characterID int64, chatID int64) string {
	return fmt.Sprintf("%d:%d", characterID, chatID)
}

func communityRoleKey(communityTypeID int64, acronym string) string {
	return fmt.Sprintf("%d:%s", communityTypeID, acronym)
}
