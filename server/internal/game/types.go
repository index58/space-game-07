package game

// задает клиентскую категорию объекта, по которой выбираются ассеты и правила отображения.
type ObjectKind string

const (
	// обозначает управляемые и неуправляемые корабли.
	ObjectKindShip ObjectKind = "ship"
	// обозначает астероиды и другие малые природные тела.
	ObjectKindAsteroid ObjectKind = "asteroid"
	// обозначает станции и прочие неподвижные инфраструктурные объекты.
	ObjectKindStation ObjectKind = "station"
)

// хранит двумерную координату или скорость в метрах игрового мира.
type WorldVector struct {
	X float64 `json:"x"`
	Y float64 `json:"y"`
}

// содержит нормализованные параметры модели, нужные симуляции и клиенту.
type CosmicObjectModel struct {
	Acronym                     string
	TitleRu                     string
	Kind                        ObjectKind
	TextureKey                  string
	TexturePath                 string
	TextureWidth                int64
	TextureHeight               int64
	TextureBodyOriginX          int64
	TextureBodyOriginY          int64
	TextureBodyWidth            int64
	TextureBodyLength           int64
	TextureScale                float64
	MassKg                      float64
	ThrustN                     float64
	MaxSpeedMps                 float64
	TorqueNm                    float64
	MaxAngularSpeedRadPerSecond float64
}

// описывает один сетевой пакет управления кораблем от клиента.
type ShipInput struct {
	Type                string  `json:"type,omitempty"`
	Seq                 int64   `json:"seq,omitempty"`
	ThrustForward       bool    `json:"thrustForward"`
	ThrustBackward      bool    `json:"thrustBackward"`
	ThrustLeft          bool    `json:"thrustLeft"`
	ThrustRight         bool    `json:"thrustRight"`
	TargetRotationDelta float64 `json:"targetRotationDelta"`
}

// является рабочим представлением объекта внутри физической симуляции.
type WorldObject struct {
	ID              int64
	Model           CosmicObjectModel
	Position        WorldVector
	Velocity        WorldVector
	Rotation        float64
	AngularVelocity float64
	TargetRotation  float64
}

// содержит состояние одного объекта в снимке мира, отправляемом клиенту.
type SnapshotObject struct {
	ID              int64      `json:"id"`
	ModelAcronym    string     `json:"modelAcronym"`
	Kind            ObjectKind `json:"kind"`
	TextureScale    float64    `json:"textureScale"`
	X               float64    `json:"x"`
	Y               float64    `json:"y"`
	VelocityX       float64    `json:"velocityX"`
	VelocityY       float64    `json:"velocityY"`
	Rotation        float64    `json:"rotation"`
	AngularVelocity float64    `json:"angularVelocity"`
	TargetRotation  float64    `json:"targetRotation"`
}

// содержит полный серверный снимок мира на конкретном тике.
type Snapshot struct {
	Type         string           `json:"type"`
	Tick         int64            `json:"tick"`
	SelfObjectID int64            `json:"selfObjectId"`
	Objects      []SnapshotObject `json:"objects"`
}

// преобразует рабочий объект симуляции в компактный сетевой DTO.
func NewSnapshotObject(object WorldObject) SnapshotObject {
	return SnapshotObject{
		ID:              object.ID,
		ModelAcronym:    object.Model.Acronym,
		Kind:            object.Model.Kind,
		TextureScale:    object.Model.TextureScale,
		X:               object.Position.X,
		Y:               object.Position.Y,
		VelocityX:       object.Velocity.X,
		VelocityY:       object.Velocity.Y,
		Rotation:        object.Rotation,
		AngularVelocity: object.AngularVelocity,
		TargetRotation:  object.TargetRotation,
	}
}
