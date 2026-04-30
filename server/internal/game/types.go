package game

// Хранит двумерную координату или скорость в метрах игрового мира.
type WorldVector struct {
	X float64 `json:"x"`
	Y float64 `json:"y"`
}

// Описывает один сетевой пакет управления кораблем от клиента.
type ShipInput struct {
	Type                string  `json:"type,omitempty"`
	Seq                 int64   `json:"seq,omitempty"`
	ThrustForward       bool    `json:"thrustForward"`
	ThrustBackward      bool    `json:"thrustBackward"`
	ThrustLeft          bool    `json:"thrustLeft"`
	ThrustRight         bool    `json:"thrustRight"`
	TargetRotationDelta float64 `json:"targetRotationDelta"`
}

// Содержит состояние одного объекта в снимке мира, отправляемом клиенту.
type SnapshotObject struct {
	ID              int64   `json:"id"`
	ModelAcronym    string  `json:"modelAcronym"`
	Kind            string  `json:"kind"`
	TextureScale    float64 `json:"textureScale"`
	X               float64 `json:"x"`
	Y               float64 `json:"y"`
	VelocityX       float64 `json:"velocityX"`
	VelocityY       float64 `json:"velocityY"`
	Rotation        float64 `json:"rotation"`
	AngularVelocity float64 `json:"angularVelocity"`
	TargetRotation  float64 `json:"targetRotation"`
}

// Содержит полный серверный снимок мира на конкретном тике.
type Snapshot struct {
	Type         string           `json:"type"`
	Tick         int64            `json:"tick"`
	SelfObjectID int64            `json:"selfObjectId"`
	Objects      []SnapshotObject `json:"objects"`
}
