package ws

import (
	"encoding/json"

	"space-game-07-server/internal/data"
	"space-game-07-server/internal/game"
)

type clientMessageType struct {
	Type string `json:"type"` // Вид входящего пакета для выбора серверного обработчика.
}

// Передает клиенту секрет, который нужно сохранить для следующих подключений.
type AuthMessage struct {
	Type  string `json:"type"`  // Вид пакета для отличия от снимков мира.
	Token string `json:"token"` // Секрет созданной учетной записи.
}

// Передает текст в текущий чат или в личный дуэт по нику.
type ChatSendMessage struct {
	Type           string `json:"type"`                     // Вид пакета для маршрутизации входящей команды.
	ChatID         int64  `json:"chatId,omitempty"`         // Чат, в который нужно отправить обычный текст.
	TargetNickname string `json:"targetNickname,omitempty"` // Ник аккаунта для адресной команды.
	Text           string `json:"text"`                     // Содержимое, которое будет записано в историю.
}

// Передает выбор вкладки чата и тем самым подтверждает прочтение ее истории.
type ChatSelectMessage struct {
	Type   string `json:"type"`   // Вид пакета для маршрутизации входящей команды.
	ChatID int64  `json:"chatId"` // Чат, который игрок выбрал в панели.
}

// Передает ошибку чата отдельным сетевым пакетом.
type ChatErrorMessage struct {
	Type    string `json:"type"`    // Вид пакета для отдельной обработки на клиенте.
	Message string `json:"message"` // Текст, который можно показать игроку в панели.
}

// InputSettingPayload передает одну аккаунтную привязку действия к событию.
type InputSettingPayload struct {
	ActionTypeID     int64 `json:"actionTypeId"`     // Игровое действие, для которого задан ввод.
	InputEventTypeID int64 `json:"inputEventTypeId"` // Событие ввода, выбранное для действия.
}

// InputSettingsMessage передает текущие сохраненные привязки аккаунта.
type InputSettingsMessage struct {
	Type     string                `json:"type"`     // Вид пакета для клиентского маршрутизатора.
	Settings []InputSettingPayload `json:"settings"` // Список привязок текущего аккаунта.
}

// InputSettingsSaveMessage передает новые привязки аккаунта на сервер.
type InputSettingsSaveMessage struct {
	Type     string                `json:"type"`     // Вид команды для сохранения настроек ввода.
	Settings []InputSettingPayload `json:"settings"` // Полный список выбранных привязок.
}

// InputSettingsErrorMessage передает причину отказа сохранения настроек.
type InputSettingsErrorMessage struct {
	Type    string `json:"type"`    // Вид пакета для клиентского маршрутизатора.
	Message string `json:"message"` // Текст ошибки для окна настроек.
}

// ControlPanelMutationMessage хранит общие поля команды изменения мира из панели.
type ControlPanelMutationMessage struct {
	ClientSessionID string `json:"clientSessionId"` // Сессия браузерного клиента, отправившая команду.
	MutationSeq     int64  `json:"mutationSeq"`     // Порядковый номер команды внутри клиентской сессии.
}

// ControlPanelObjectUpdateMessage передает частичное изменение управляемого объекта.
type ControlPanelObjectUpdateMessage struct {
	Type string `json:"type"` // Вид команды для маршрутизации изменения объекта.
	ControlPanelMutationMessage
	Enabled *bool   `json:"enabled,omitempty"` // Новое состояние включения объекта, если оно меняется.
	Title   *string `json:"title,omitempty"`   // Новое пользовательское название объекта, если оно меняется.
}

// ControlPanelEquipmentUpdateMessage передает частичное изменение группы оборудования.
type ControlPanelEquipmentUpdateMessage struct {
	Type string `json:"type"` // Вид команды для маршрутизации изменения оборудования.
	ControlPanelMutationMessage
	EquipmentGroupID int64  `json:"equipmentGroupId"`       // Группа оборудования, которую нужно изменить.
	Enabled          *bool  `json:"enabled,omitempty"`      // Новое состояние включения группы, если оно меняется.
	EnabledCount     *int64 `json:"enabledCount,omitempty"` // Новое количество включенных единиц, если оно меняется.
}

// ControlPanelErrorMessage передает отказ команды панели управления.
type ControlPanelErrorMessage struct {
	Type            string `json:"type"`            // Вид пакета для клиентского маршрутизатора.
	ClientSessionID string `json:"clientSessionId"` // Сессия, команда которой была отклонена.
	MutationSeq     int64  `json:"mutationSeq"`     // Номер отклоненной команды.
	Message         string `json:"message"`         // Текст ошибки для диагностики клиента.
}

// Разбирает клиентский JSON и пропускает только сообщения управления кораблем.
func DecodeInputMessage(payload []byte) (game.ShipInput, bool) {
	var input game.ShipInput
	if err := json.Unmarshal(payload, &input); err != nil {
		return game.ShipInput{}, false
	}

	if input.Type != "input" {
		return game.ShipInput{}, false
	}

	return input, true
}

// Проверяет, что клиент прислал команду смены корабля.
func DecodeRandomShipMessage(payload []byte) bool {
	var message clientMessageType
	if err := json.Unmarshal(payload, &message); err != nil {
		return false
	}

	return message.Type == "randomShip"
}

// DecodeInputSettingsRequestMessage проверяет, что клиент запросил свежие настройки ввода.
func DecodeInputSettingsRequestMessage(payload []byte) bool {
	var message clientMessageType
	if err := json.Unmarshal(payload, &message); err != nil {
		return false
	}

	return message.Type == "inputSettingsRequest"
}

// Разбирает клиентский JSON и пропускает только команды чата.
func DecodeChatSendMessage(payload []byte) (ChatSendMessage, bool) {
	var message ChatSendMessage
	if err := json.Unmarshal(payload, &message); err != nil {
		return ChatSendMessage{}, false
	}

	if message.Type != "chatSend" {
		return ChatSendMessage{}, false
	}

	return message, true
}

// Разбирает клиентский JSON и пропускает только выбор вкладки чата.
func DecodeChatSelectMessage(payload []byte) (ChatSelectMessage, bool) {
	var message ChatSelectMessage
	if err := json.Unmarshal(payload, &message); err != nil {
		return ChatSelectMessage{}, false
	}

	if message.Type != "chatSelect" || message.ChatID <= 0 {
		return ChatSelectMessage{}, false
	}

	return message, true
}

// Разбирает клиентский JSON и пропускает только сохранение настроек ввода.
func DecodeInputSettingsSaveMessage(payload []byte) (InputSettingsSaveMessage, bool) {
	var message InputSettingsSaveMessage
	if err := json.Unmarshal(payload, &message); err != nil {
		return InputSettingsSaveMessage{}, false
	}

	if message.Type != "inputSettingsSave" {
		return InputSettingsSaveMessage{}, false
	}

	return message, true
}

// Разбирает клиентский JSON и пропускает только команды изменения объекта панели управления.
func DecodeControlPanelObjectUpdateMessage(payload []byte) (ControlPanelObjectUpdateMessage, bool) {
	var message ControlPanelObjectUpdateMessage
	if err := json.Unmarshal(payload, &message); err != nil {
		return ControlPanelObjectUpdateMessage{}, false
	}

	if message.Type != "controlPanelObjectUpdate" || !validControlPanelMutation(message.ControlPanelMutationMessage) || (message.Enabled == nil && message.Title == nil) {
		return ControlPanelObjectUpdateMessage{}, false
	}

	return message, true
}

// Разбирает клиентский JSON и пропускает только команды изменения оборудования панели управления.
func DecodeControlPanelEquipmentUpdateMessage(payload []byte) (ControlPanelEquipmentUpdateMessage, bool) {
	var message ControlPanelEquipmentUpdateMessage
	if err := json.Unmarshal(payload, &message); err != nil {
		return ControlPanelEquipmentUpdateMessage{}, false
	}

	if message.Type != "controlPanelEquipmentUpdate" || !validControlPanelMutation(message.ControlPanelMutationMessage) || message.EquipmentGroupID <= 0 || (message.Enabled == nil && message.EnabledCount == nil) {
		return ControlPanelEquipmentUpdateMessage{}, false
	}

	return message, true
}

// Сериализует снимок мира в формат WebSocket-сообщения.
func EncodeSnapshotMessage(snapshot game.Snapshot) ([]byte, error) {
	return json.Marshal(snapshot)
}

// Сериализует состояние вкладок и истории чата.
func EncodeChatStateMessage(chatState game.ChatState) ([]byte, error) {
	return json.Marshal(chatState)
}

// Сериализует причину отказа команды чата.
func EncodeChatErrorMessage(message string) ([]byte, error) {
	return json.Marshal(ChatErrorMessage{
		Type:    "chatError",
		Message: message,
	})
}

// Сериализует текущие настройки ввода аккаунта.
func EncodeInputSettingsMessage(settings []data.AccountActionInputSetting) ([]byte, error) {
	payload := InputSettingsMessage{
		Type:     "inputSettings",
		Settings: make([]InputSettingPayload, 0, len(settings)),
	}
	for _, setting := range settings {
		payload.Settings = append(payload.Settings, InputSettingPayload{
			ActionTypeID:     setting.ActionTypeID,
			InputEventTypeID: setting.InputEventTypeID,
		})
	}
	return json.Marshal(payload)
}

// Сериализует причину отказа сохранения настроек ввода.
func EncodeInputSettingsErrorMessage(message string) ([]byte, error) {
	return json.Marshal(InputSettingsErrorMessage{
		Type:    "inputSettingsError",
		Message: message,
	})
}

// Сериализует отказ команды панели управления с номером мутации.
func EncodeControlPanelErrorMessage(sessionID string, mutationSeq int64, message string) ([]byte, error) {
	return json.Marshal(ControlPanelErrorMessage{
		Type:            "controlPanelError",
		ClientSessionID: sessionID,
		MutationSeq:     mutationSeq,
		Message:         message,
	})
}

// Сериализует результат автоматической авторизации.
func EncodeAuthMessage(token string) ([]byte, error) {
	return json.Marshal(AuthMessage{
		Type:  "auth",
		Token: token,
	})
}

// Проверяет общие поля команды панели управления.
func validControlPanelMutation(message ControlPanelMutationMessage) bool {
	return message.ClientSessionID != "" && message.MutationSeq > 0
}
