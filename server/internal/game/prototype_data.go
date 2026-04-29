package game

const textureScale = 4

var ShipBat = CosmicObjectModel{
	Acronym:                     "ship_bat",
	TitleRu:                     "Летучая мышь",
	Kind:                        ObjectKindShip,
	TextureKey:                  "world.ship.ship_bat",
	TexturePath:                 "/assets/world/cosmic-objects/ships/256x512/ship_256x512_0008.png",
	TextureWidth:                256,
	TextureHeight:               512,
	TextureBodyOriginX:          126,
	TextureBodyOriginY:          259,
	TextureBodyWidth:            88,
	TextureBodyLength:           90,
	TextureScale:                textureScale,
	MassKg:                      7.92 * 1000,
	ThrustN:                     0.006439507649442245 * 200000000,
	MaxSpeedMps:                 497,
	TorqueNm:                    653.5649999999999 * 1000,
	MaxAngularSpeedRadPerSecond: 3,
}

var Asteroid0002 = CosmicObjectModel{
	Acronym:                     "asteroid_0002",
	TitleRu:                     "Астероид",
	Kind:                        ObjectKindAsteroid,
	TextureKey:                  "world.asteroid.asteroid_0002",
	TexturePath:                 "/assets/world/cosmic-objects/asteroids/asteroid_0002.png",
	TextureWidth:                2048,
	TextureHeight:               2048,
	TextureBodyOriginX:          988,
	TextureBodyOriginY:          1289,
	TextureBodyWidth:            804,
	TextureBodyLength:           783,
	TextureScale:                textureScale,
	MassKg:                      629.532 * 1000,
	ThrustN:                     0,
	MaxSpeedMps:                 475,
	TorqueNm:                    0,
	MaxAngularSpeedRadPerSecond: 3,
}

var StationTinyCrumb = CosmicObjectModel{
	Acronym:                     "station_tiny_crumb",
	TitleRu:                     "Крошка",
	Kind:                        ObjectKindStation,
	TextureKey:                  "world.station.station_tiny_crumb",
	TexturePath:                 "/assets/world/cosmic-objects/stations/station_0064.png",
	TextureWidth:                2048,
	TextureHeight:               2048,
	TextureBodyOriginX:          996,
	TextureBodyOriginY:          738,
	TextureBodyWidth:            225,
	TextureBodyLength:           825,
	TextureScale:                textureScale,
	MassKg:                      185.625 * 1000,
	ThrustN:                     0,
	MaxSpeedMps:                 486,
	TorqueNm:                    0,
	MaxAngularSpeedRadPerSecond: 3,
}

func NewPlayerShip(id int64, position WorldVector) WorldObject {
	return WorldObject{
		ID:       id,
		Model:    ShipBat,
		Position: position,
	}
}

func StaticObjects() []WorldObject {
	return []WorldObject{
		{ID: 1, Model: Asteroid0002, Position: WorldVector{X: -500, Y: 800}},
		{ID: 2, Model: StationTinyCrumb, Position: WorldVector{X: 500, Y: 500}},
	}
}
