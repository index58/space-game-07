package game

// ExchangeEvent описывает клиентское событие обмена или уведомление.
type ExchangeEvent struct {
	Type      string        `json:"type"`               // Вид сетевого сообщения для маршрутизации на клиенте.
	Kind      string        `json:"kind"`               // Вид события внутри сообщения.
	Role      string        `json:"role,omitempty"`     // Роль текущего клиента в окне обмена.
	Message   string        `json:"message,omitempty"`  // Текст уведомления или причины отказа.
	Duration  float64       `json:"duration,omitempty"` // Длительность окна ожидания или процесса в секундах.
	State     ExchangeState `json:"state,omitempty"`    // Текущее состояние окна, если его нужно перерисовать.
	ObjectIDs []int64       `json:"-"`                  // Объекты, подключенным игрокам которых нужно отправить событие.
}

// ExchangeState хранит состояние окна обмена для одного клиента.
type ExchangeState struct {
	SelfObjectID            int64               `json:"selfObjectId"`                          // Объект текущего игрока.
	OtherObjectID           int64               `json:"otherObjectId"`                         // Объект второго игрока.
	SelfNickname            string              `json:"selfNickname"`                          // Никнейм текущего игрока.
	OtherNickname           string              `json:"otherNickname"`                         // Никнейм второго игрока.
	SelfReceiverContainerID int64               `json:"selfReceiverContainerEquipmentGroupId"` // Контейнер-приемник текущего игрока.
	SelfSourceContainerID   int64               `json:"selfSourceContainerEquipmentGroupId"`   // Контейнер-источник текущего игрока.
	SelfConfirmed           bool                `json:"selfConfirmed"`                         // Подтвердил ли текущий игрок.
	OtherConfirmed          bool                `json:"otherConfirmed"`                        // Подтвердил ли второй игрок.
	NotEnoughSpace          bool                `json:"notEnoughSpace"`                        // Нужно ли показать нехватку места текущему игроку.
	SelfQueue               []ExchangeQueueItem `json:"selfQueue"`                             // Предметы, которые текущий игрок отдает.
	OtherQueue              []ExchangeQueueItem `json:"otherQueue"`                            // Предметы, которые второй игрок отдает.
}

// ExchangeQueueItem описывает одну строку очереди обмена.
type ExchangeQueueItem struct {
	TaskItemGroupID int64   `json:"taskItemGroupId"` // Строка задания, связанная с визуальной строкой.
	ItemModelID     int64   `json:"itemModelId"`     // Модель предметов в строке.
	Count           float64 `json:"count"`           // Количество предметов в строке.
	Progress        float64 `json:"progress"`        // Доля выполненного перемещения от 0 до 1.
	IsReady         bool    `json:"isReady"`         // Готова ли строка к финальному переносу.
}
