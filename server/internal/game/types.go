package game

import "space-game-07-server/internal/data"

// Хранит двумерную координату или скорость в метрах игрового мира.
type WorldVector struct {
	X float64 `json:"x"` // Горизонтальная координата или компонента вектора.
	Y float64 `json:"y"` // Вертикальная координата или компонента вектора.
}

// ChatState описывает доступные вкладки чата для одного подключенного аккаунта.
type ChatState struct {
	Type           string    `json:"type"`           // Вид сетевого сообщения для клиентского маршрутизатора.
	Tabs           []ChatTab `json:"tabs"`           // Доступные вкладки чатов с последними сообщениями.
	SelectedChatID int64     `json:"selectedChatId"` // Чат, который должен быть выбран на клиенте.
}

// ChatTab описывает одну вкладку чата и историю, видимую получателю.
type ChatTab struct {
	ChatID               int64         `json:"chatId"`               // Уникальный идентификатор чата.
	Title                string        `json:"title"`                // Подпись вкладки для интерфейса.
	CommunityTypeAcronym string        `json:"communityTypeAcronym"` // Тип сообщества, к которому относится чат.
	DuoChatKey           string        `json:"duoChatKey"`           // Ключ личной переписки двух персонажей.
	UnreadCount          int64         `json:"unreadCount"`          // Количество сообщений после последней прочитанной строки.
	Messages             []ChatMessage `json:"messages"`             // Полная доступная история выбранного чата.
}

// ChatMessage описывает одно сообщение в клиентском представлении чата.
type ChatMessage struct {
	ID                 int64  `json:"id"`                 // Уникальный идентификатор сообщения.
	ChatID             int64  `json:"chatId"`             // Чат, в котором находится сообщение.
	SenderCharacterID  int64  `json:"senderCharacterId"`  // Персонаж, от имени которого отправлено сообщение.
	SenderNickname     string `json:"senderNickname"`     // Временное отображаемое имя из аккаунта отправителя.
	MessageTypeAcronym string `json:"messageTypeAcronym"` // Тип сообщения для визуального оформления.
	Text               string `json:"text"`               // Текст сообщения.
	Color              string `json:"color"`              // Цвет сообщения в RGB-HEX формате без решетки.
	SentTime           string `json:"sentTime"`           // Время отправки в RFC3339Nano для клиента.
}

// Описывает один сетевой пакет управления кораблем от клиента.
type ShipInput struct {
	Type                string  `json:"type,omitempty"`      // Вид сетевого сообщения от клиента.
	Seq                 int64   `json:"seq,omitempty"`       // Порядковый номер пакета управления.
	ThrustForward       bool    `json:"thrustForward"`       // Запрос продольной тяги вперед.
	ThrustBackward      bool    `json:"thrustBackward"`      // Запрос продольной тяги назад.
	ThrustLeft          bool    `json:"thrustLeft"`          // Запрос поперечной тяги влево.
	ThrustRight         bool    `json:"thrustRight"`         // Запрос поперечной тяги вправо.
	ToggleAnchor        bool    `json:"toggleAnchor"`        // Одноразовый запрос переключения якоря.
	TargetRotationDelta float64 `json:"targetRotationDelta"` // Изменение целевого угла поворота за пакет.
}

// Содержит полный серверный снимок мира на конкретном тике.
type Snapshot struct {
	Type            string                `json:"type"`            // Вид сетевого сообщения со снимком мира.
	Tick            int64                 `json:"tick"`            // Номер шага симуляции, на котором сделан снимок.
	SelfObjectID    int64                 `json:"selfObjectId"`    // Управляемый объект получателя снимка.
	Objects         []data.CosmicObject   `json:"objects"`         // Объекты мира, видимые клиенту в текущем снимке.
	EquipmentGroups []data.EquipmentGroup `json:"equipmentGroups"` // Группы оборудования, нужные UI для панели пилота.
}
