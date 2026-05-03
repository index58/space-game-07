package game

import "space-game-07-server/internal/data"

// Хранит двумерную координату или скорость в метрах игрового мира.
type WorldVector struct {
	X float64 `json:"x"` // Горизонтальная координата или компонента вектора.
	Y float64 `json:"y"` // Вертикальная координата или компонента вектора.
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
