package world

import (
	"fmt"
	"sort"
	"strings"
	"time"

	"space-game-07-server/internal/data"
	"space-game-07-server/internal/game"
)

const (
	communityTypeServer       = "Server"
	communityTypeDuo          = "Duo"
	communityChatRoleMember   = "Member"
	messageTypeFromCharacter  = "FromCharacter"
	defaultCharacterTextColor = "D8F3FF"
)

// ChatStateForAccount возвращает вкладки чата, доступные текущему персонажу аккаунта.
func (world *World) ChatStateForAccount(accountID int64, selectedChatID int64) (game.ChatState, bool) {
	world.mu.Lock()
	defer world.mu.Unlock()

	account, character, ok := world.currentAccountCharacterLocked(accountID)
	if !ok {
		return game.ChatState{}, false
	}
	return world.chatStateLocked(account, character, selectedChatID), true
}

// SendChatMessage сохраняет сообщение и возвращает обновленное состояние для отправителя.
func (world *World) SendChatMessage(accountID int64, chatID int64, targetNickname string, text string) (game.ChatState, []int64, string) {
	world.mu.Lock()
	defer world.mu.Unlock()

	account, character, ok := world.currentAccountCharacterLocked(accountID)
	if !ok {
		return game.ChatState{}, nil, "Аккаунт не подключен к персонажу"
	}

	text = strings.TrimSpace(text)
	if text == "" {
		return world.chatStateLocked(account, character, chatID), nil, "Нельзя отправить пустое сообщение"
	}

	var chat *data.Chat
	var recipients []int64
	if targetNickname != "" {
		targetAccount, ok := world.data.Accounts.GetByNickname(targetNickname)
		if !ok || targetAccount.CurrentCharacterID <= 0 {
			return world.chatStateLocked(account, character, chatID), nil, "Адресат не найден"
		}
		targetCharacter, ok := world.data.Characters.Get(targetAccount.CurrentCharacterID)
		if !ok {
			return world.chatStateLocked(account, character, chatID), nil, "Адресат не найден"
		}
		chat = world.ensureDuoChatLocked(character.ID, targetCharacter.ID)
		recipients = world.connectedAccountsByCharacterIDsLocked([]int64{character.ID, targetCharacter.ID})
	} else {
		var ok bool
		chat, ok = world.data.Chats.Get(chatID)
		if !ok || !world.characterCanReadChatLocked(character.ID, chat) {
			return world.chatStateLocked(account, character, chatID), nil, "Чат недоступен"
		}
		recipients = world.connectedAccountsForChatLocked(chat)
	}

	messageType, ok := world.data.MessageTypes.GetByAcronym(messageTypeFromCharacter)
	if !ok {
		return world.chatStateLocked(account, character, chat.ID), nil, "Тип сообщения не найден"
	}
	if _, err := world.data.Messages.Add(&data.Message{
		ChatID:            chat.ID,
		SenderCharacterID: character.ID,
		MessageTypeID:     messageType.ID,
		Text:              text,
		Color:             defaultCharacterTextColor,
		SentTime:          time.Now(),
	}); err != nil {
		return world.chatStateLocked(account, character, chat.ID), nil, err.Error()
	}

	return world.chatStateLocked(account, character, chat.ID), recipients, ""
}

// ensureChatData готовит минимальные справочники и серверный чат для первой версии.
func (world *World) ensureChatData() {
	if world.data.Chats == nil {
		world.data.Chats = data.NewChats()
	}
	if world.data.ChatMembers == nil {
		world.data.ChatMembers = data.NewChatMembers()
	}
	if world.data.CommunityTypes == nil {
		world.data.CommunityTypes = data.NewCommunityTypes()
	}
	if world.data.CommunityChatRoles == nil {
		world.data.CommunityChatRoles = data.NewCommunityChatRoles()
	}
	if world.data.Messages == nil {
		world.data.Messages = data.NewMessages()
	}
	if world.data.MessageReads == nil {
		world.data.MessageReads = data.NewMessageReads()
	}
	if world.data.MessageTypes == nil {
		world.data.MessageTypes = data.NewMessageTypes()
	}

	serverType := world.ensureCommunityType(communityTypeServer, "Сервер", "Server")
	duoType := world.ensureCommunityType(communityTypeDuo, "Дуэт", "Duo")
	world.ensureCommunityRole(serverType.ID, communityChatRoleMember, "Участник", "Member")
	world.ensureCommunityRole(duoType.ID, communityChatRoleMember, "Участник", "Member")
	world.ensureMessageType(messageTypeFromCharacter, "От персонажа", "From character")
	if _, ok := world.data.Chats.GetByCommunity(serverType.ID, 0); !ok {
		_, _ = world.data.Chats.Add(&data.Chat{
			CommunityTypeID: serverType.ID,
			CommunityID:     0,
			DuoChatKey:      "",
		})
	}
}

// ensureCommunityType возвращает существующий тип сообщества или добавляет базовую запись.
func (world *World) ensureCommunityType(acronym string, titleRu string, titleEn string) *data.CommunityType {
	if communityType, ok := world.data.CommunityTypes.GetByAcronym(acronym); ok {
		return communityType
	}
	communityType, err := world.data.CommunityTypes.Add(&data.CommunityType{
		TitleRu: titleRu,
		TitleEn: titleEn,
		Acronym: acronym,
	})
	if err != nil {
		panic(err)
	}
	return communityType
}

// ensureCommunityRole возвращает существующую роль или добавляет базовую роль участника.
func (world *World) ensureCommunityRole(communityTypeID int64, acronym string, titleRu string, titleEn string) *data.CommunityChatRole {
	if role, ok := world.data.CommunityChatRoles.GetByTypeAndAcronym(communityTypeID, acronym); ok {
		return role
	}
	role, err := world.data.CommunityChatRoles.Add(&data.CommunityChatRole{
		CommunityTypeID: communityTypeID,
		TitleRu:         titleRu,
		TitleEn:         titleEn,
		Acronym:         acronym,
	})
	if err != nil {
		panic(err)
	}
	return role
}

// ensureMessageType возвращает существующий тип сообщения или добавляет базовую запись.
func (world *World) ensureMessageType(acronym string, titleRu string, titleEn string) *data.MessageType {
	if messageType, ok := world.data.MessageTypes.GetByAcronym(acronym); ok {
		return messageType
	}
	messageType, err := world.data.MessageTypes.Add(&data.MessageType{
		TitleRu: titleRu,
		TitleEn: titleEn,
		Acronym: acronym,
	})
	if err != nil {
		panic(err)
	}
	return messageType
}

// chatStateLocked собирает клиентское представление чатов под общей блокировкой мира.
func (world *World) chatStateLocked(account *data.Account, character *data.Character, selectedChatID int64) game.ChatState {
	tabs := world.chatTabsForCharacterLocked(account, character)
	if len(tabs) == 0 {
		return game.ChatState{Type: "chatState"}
	}
	if selectedChatID == 0 || !chatTabExists(tabs, selectedChatID) {
		selectedChatID = tabs[0].ChatID
	}
	world.markChatReadLocked(character.ID, selectedChatID)
	tabs = world.chatTabsForCharacterLocked(account, character)
	return game.ChatState{
		Type:           "chatState",
		Tabs:           tabs,
		SelectedChatID: selectedChatID,
	}
}

// chatTabsForCharacterLocked собирает серверную вкладку и открытые дуэты персонажа.
func (world *World) chatTabsForCharacterLocked(account *data.Account, character *data.Character) []game.ChatTab {
	chats := make([]*data.Chat, 0)
	if serverType, ok := world.data.CommunityTypes.GetByAcronym(communityTypeServer); ok {
		if serverChat, ok := world.data.Chats.GetByCommunity(serverType.ID, 0); ok {
			chats = append(chats, serverChat)
		}
	}
	for _, member := range world.data.ChatMembers.GetByCharacterID(character.ID) {
		if chat, ok := world.data.Chats.Get(member.ChatID); ok {
			chats = append(chats, chat)
		}
	}
	sort.Slice(chats, func(left int, right int) bool {
		return chats[left].ID < chats[right].ID
	})

	tabs := make([]game.ChatTab, 0, len(chats))
	seen := map[int64]bool{}
	for _, chat := range chats {
		if seen[chat.ID] || !world.characterCanReadChatLocked(character.ID, chat) {
			continue
		}
		seen[chat.ID] = true
		tabs = append(tabs, world.chatTabLocked(character, chat))
	}
	return tabs
}

// chatTabLocked строит одну вкладку с полной доступной историей сообщений.
func (world *World) chatTabLocked(character *data.Character, chat *data.Chat) game.ChatTab {
	communityType, _ := world.data.CommunityTypes.Get(chat.CommunityTypeID)
	title := communityType.Acronym
	if communityType.Acronym == communityTypeDuo {
		title = world.duoPeerNicknameLocked(character.ID, chat)
	}
	messages := world.data.Messages.GetByChatID(chat.ID)

	return game.ChatTab{
		ChatID:               chat.ID,
		Title:                title,
		CommunityTypeAcronym: communityType.Acronym,
		DuoChatKey:           chat.DuoChatKey,
		UnreadCount:          world.unreadMessageCountLocked(character.ID, chat.ID, messages),
		Messages:             world.chatMessagesLocked(messages),
	}
}

// markChatReadLocked сохраняет, что персонаж видел последнюю строку выбранного чата.
func (world *World) markChatReadLocked(characterID int64, chatID int64) {
	if world.data.MessageReads == nil {
		return
	}
	messages := world.data.Messages.GetByChatID(chatID)
	lastMessageID := int64(0)
	if len(messages) > 0 {
		lastMessageID = messages[len(messages)-1].ID
	}
	_, _ = world.data.MessageReads.SetLastRead(characterID, chatID, lastMessageID)
}

// unreadMessageCountLocked считает строки после сохраненной позиции чтения.
func (world *World) unreadMessageCountLocked(characterID int64, chatID int64, messages []*data.Message) int64 {
	if world.data.MessageReads == nil {
		return 0
	}
	messageRead, ok := world.data.MessageReads.GetByCharacterAndChat(characterID, chatID)
	if !ok {
		return int64(len(messages))
	}
	var count int64
	for _, message := range messages {
		if message.ID > messageRead.LastReadMessageID {
			count++
		}
	}
	return count
}

// chatMessagesLocked переводит серверные записи в сетевое представление.
func (world *World) chatMessagesLocked(messages []*data.Message) []game.ChatMessage {
	result := make([]game.ChatMessage, 0, len(messages))
	for _, message := range messages {
		messageTypeAcronym := ""
		if messageType, ok := world.data.MessageTypes.Get(message.MessageTypeID); ok {
			messageTypeAcronym = messageType.Acronym
		}
		result = append(result, game.ChatMessage{
			ID:                 message.ID,
			ChatID:             message.ChatID,
			SenderCharacterID:  message.SenderCharacterID,
			SenderNickname:     world.nicknameByCharacterIDLocked(message.SenderCharacterID),
			MessageTypeAcronym: messageTypeAcronym,
			Text:               message.Text,
			Color:              message.Color,
			SentTime:           message.SentTime.Format(time.RFC3339Nano),
		})
	}
	return result
}

// currentAccountCharacterLocked возвращает аккаунт и текущего персонажа под общей блокировкой мира.
func (world *World) currentAccountCharacterLocked(accountID int64) (*data.Account, *data.Character, bool) {
	if _, ok := world.accountObjectIDs[accountID]; !ok {
		return nil, nil, false
	}
	account, ok := world.data.Accounts.Get(accountID)
	if !ok || account.CurrentCharacterID <= 0 {
		return nil, nil, false
	}
	character, ok := world.data.Characters.Get(account.CurrentCharacterID)
	if !ok || character.AccountID != account.ID {
		return nil, nil, false
	}
	return account, character, true
}

// ensureDuoChatLocked возвращает существующий дуэт или создает новый с двумя участниками.
func (world *World) ensureDuoChatLocked(firstCharacterID int64, secondCharacterID int64) *data.Chat {
	duoType, _ := world.data.CommunityTypes.GetByAcronym(communityTypeDuo)
	key := duoChatKey(firstCharacterID, secondCharacterID)
	if chat, ok := world.data.Chats.GetByDuoKey(key); ok {
		return chat
	}
	nextID := world.data.Chats.MaxID + 1
	chat, err := world.data.Chats.Add(&data.Chat{
		CommunityTypeID: duoType.ID,
		CommunityID:     nextID,
		DuoChatKey:      key,
	})
	if err != nil {
		panic(err)
	}
	role, _ := world.data.CommunityChatRoles.GetByTypeAndAcronym(duoType.ID, communityChatRoleMember)
	_, _ = world.data.ChatMembers.Add(&data.ChatMember{ChatID: chat.ID, MemberCharacterID: firstCharacterID, CommunityChatRoleID: role.ID})
	if firstCharacterID != secondCharacterID {
		_, _ = world.data.ChatMembers.Add(&data.ChatMember{ChatID: chat.ID, MemberCharacterID: secondCharacterID, CommunityChatRoleID: role.ID})
	}
	return chat
}

// characterCanReadChatLocked проверяет доступ персонажа к указанному чату.
func (world *World) characterCanReadChatLocked(characterID int64, chat *data.Chat) bool {
	communityType, ok := world.data.CommunityTypes.Get(chat.CommunityTypeID)
	if !ok {
		return false
	}
	if communityType.Acronym == communityTypeServer {
		return true
	}
	if communityType.Acronym == communityTypeDuo {
		return world.data.ChatMembers.HasMember(chat.ID, characterID)
	}
	return false
}

// connectedAccountsForChatLocked возвращает подключенных получателей указанного чата.
func (world *World) connectedAccountsForChatLocked(chat *data.Chat) []int64 {
	communityType, ok := world.data.CommunityTypes.Get(chat.CommunityTypeID)
	if !ok {
		return nil
	}
	if communityType.Acronym == communityTypeServer {
		accountIDs := make([]int64, 0, len(world.accountObjectIDs))
		for accountID := range world.accountObjectIDs {
			accountIDs = append(accountIDs, accountID)
		}
		sort.Slice(accountIDs, func(left int, right int) bool {
			return accountIDs[left] < accountIDs[right]
		})
		return accountIDs
	}

	characterIDs := make([]int64, 0)
	for _, member := range world.data.ChatMembers.GetByChatID(chat.ID) {
		characterIDs = append(characterIDs, member.MemberCharacterID)
	}
	return world.connectedAccountsByCharacterIDsLocked(characterIDs)
}

// connectedAccountsByCharacterIDsLocked возвращает подключенные аккаунты указанных персонажей.
func (world *World) connectedAccountsByCharacterIDsLocked(characterIDs []int64) []int64 {
	characterSet := map[int64]bool{}
	for _, characterID := range characterIDs {
		characterSet[characterID] = true
	}
	accountIDs := make([]int64, 0, len(characterSet))
	for accountID := range world.accountObjectIDs {
		account, ok := world.data.Accounts.Get(accountID)
		if ok && characterSet[account.CurrentCharacterID] {
			accountIDs = append(accountIDs, accountID)
		}
	}
	sort.Slice(accountIDs, func(left int, right int) bool {
		return accountIDs[left] < accountIDs[right]
	})
	return accountIDs
}

// duoPeerNicknameLocked возвращает ник второго участника личного чата.
func (world *World) duoPeerNicknameLocked(characterID int64, chat *data.Chat) string {
	for _, member := range world.data.ChatMembers.GetByChatID(chat.ID) {
		if member.MemberCharacterID != characterID {
			return world.nicknameByCharacterIDLocked(member.MemberCharacterID)
		}
	}
	return ""
}

// nicknameByCharacterIDLocked временно берет отображаемое имя из аккаунта персонажа.
func (world *World) nicknameByCharacterIDLocked(characterID int64) string {
	account, ok := world.data.Accounts.GetByCurrentCharacterID(characterID)
	if !ok {
		return ""
	}
	return account.Nickname
}

func chatTabExists(tabs []game.ChatTab, chatID int64) bool {
	for _, tab := range tabs {
		if tab.ChatID == chatID {
			return true
		}
	}
	return false
}

func duoChatKey(firstCharacterID int64, secondCharacterID int64) string {
	if firstCharacterID > secondCharacterID {
		firstCharacterID, secondCharacterID = secondCharacterID, firstCharacterID
	}
	return fmt.Sprintf("%d:%d", firstCharacterID, secondCharacterID)
}
