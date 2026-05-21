package physics_test

import (
	"math"
	"testing"

	"space-game-07-server/internal/data"
	"space-game-07-server/internal/game"
	"space-game-07-server/internal/physics"
)

// Р’РѕР·РІСЂР°С‰Р°РµС‚ РїСѓСЃС‚РѕР№ РІРІРѕРґ РґР»СЏ СЃС†РµРЅР°СЂРёРµРІ, РіРґРµ РїСЂРѕРІРµСЂСЏРµС‚СЃСЏ РёРЅРµСЂС†РёСЏ РёР»Рё С‚РѕСЂРјРѕР¶РµРЅРёРµ.
func idleInput() game.ShipInput {
	return game.ShipInput{}
}

// РЎРѕР·РґР°С‘С‚ РѕР±СЉРµРєС‚ СЃ СѓСЃС‚РѕР№С‡РёРІС‹РјРё РїР°СЂР°РјРµС‚СЂР°РјРё, С‡С‚РѕР±С‹ С„РёР·РёС‡РµСЃРєРёРµ РѕР¶РёРґР°РЅРёСЏ Р±С‹Р»Рё РІРѕСЃРїСЂРѕРёР·РІРѕРґРёРјС‹РјРё.
func testShip(id int64, x float64, y float64) data.CosmicObject {
	return data.CosmicObject{
		ID:                  id,
		CosmicObjectModelID: 1,
		X:                   x,
		Y:                   y,
		Mass:                7920,
		MaxSpeed:            497,
		MaxAngularSpeed:     3,
		MaxAlongForce:       1287901.529888449,
		MaxTorque:           653565,
	}
}

// РЎРѕР·РґР°С‘С‚ РјРѕРґРµР»СЊ СЃ СЂР°Р·РјРµСЂР°РјРё С„РёР·РёС‡РµСЃРєРѕРіРѕ С‚РµР»Р° РІ РїРёРєСЃРµР»СЏС… С‚РµРєСЃС‚СѓСЂС‹.
func testModel() data.CosmicObjectModel {
	return data.CosmicObjectModel{
		ID:                1,
		TextureBodyWidth:  88,
		TextureBodyLength: 90,
		TextureScale:      4,
	}
}

// РЎСЂР°РІРЅРёРІР°РµС‚ float64 СЃ РјР°Р»С‹Рј РґРѕРїСѓСЃРєРѕРј, РїРѕС‚РѕРјСѓ С‡С‚Рѕ С„РёР·РёРєР° СЂР°Р±РѕС‚Р°РµС‚ СЃ РґСЂРѕР±РЅС‹РјРё РІРµР»РёС‡РёРЅР°РјРё.
func closeTo(t *testing.T, actual float64, expected float64) {
	t.Helper()

	if math.Abs(actual-expected) > 0.000001 {
		t.Fatalf("got %v, want %v", actual, expected)
	}
}

// РџСЂРѕРІРµСЂСЏРµС‚, С‡С‚Рѕ РјРѕРјРµРЅС‚ РёРЅРµСЂС†РёРё СЃС‡РёС‚Р°РµС‚СЃСЏ РїРѕ РїСЂРёР±Р»РёР¶РµРЅРёСЋ Р·Р°РїРѕР»РЅРµРЅРЅРѕРіРѕ СЌР»Р»РёРїСЃР°.
func TestMomentOfInertiaUsesFilledEllipseApproximation(t *testing.T) {
	closeTo(t, physics.MomentOfInertia(testShip(1, 0, 0), testModel()), 490173.75)
}

// РџСЂРѕРІРµСЂСЏРµС‚, С‡С‚Рѕ РїСЂРё РЅСѓР»РµРІРѕРј РїРѕРІРѕСЂРѕС‚Рµ С‚СЏРіР° РІРїРµСЂС‘Рґ СѓСЃРєРѕСЂСЏРµС‚ РєРѕСЂР°Р±Р»СЊ РїРѕ РїРѕР»РѕР¶РёС‚РµР»СЊРЅРѕР№ РѕСЃРё Y.
func TestStepShipAcceleratesForwardAlongPositiveYAtZeroRotation(t *testing.T) {
	next := physics.StepShip(
		testShip(1, 0, 0),
		testModel(),
		game.ShipInput{ThrustForward: true},
		1,
	)

	closeTo(t, next.VelocityX, 0)
	closeTo(t, next.VelocityY, 162.61382953137096)
}

// РџСЂРѕРІРµСЂСЏРµС‚, С‡С‚Рѕ РјР°СЃСЃР° РѕР±СЉРµРєС‚Р° РЅР°РїСЂСЏРјСѓСЋ СѓС‡Р°СЃС‚РІСѓРµС‚ РІ СЂР°СЃС‡С‘С‚Рµ СѓСЃРєРѕСЂРµРЅРёСЏ РєР°Рє РєРёР»РѕРіСЂР°РјРјС‹.
func TestStepShipUsesObjectMassAsKilograms(t *testing.T) {
	ship := testShip(1, 0, 0)
	ship.Mass = 1000
	ship.MaxAlongForce = 100000
	ship.MaxSpeed = 1000

	next := physics.StepShip(
		ship,
		testModel(),
		game.ShipInput{ThrustForward: true},
		1,
	)

	closeTo(t, next.VelocityY, 100)
}

// РџСЂРѕРІРµСЂСЏРµС‚, С‡С‚Рѕ Р»РёРЅРµР№РЅР°СЏ СЃРєРѕСЂРѕСЃС‚СЊ РѕРіСЂР°РЅРёС‡РёРІР°РµС‚СЃСЏ РјР°РєСЃРёРјР°Р»СЊРЅС‹Рј Р·РЅР°С‡РµРЅРёРµРј РјРѕРґРµР»Рё.
func TestStepShipClampsLinearVelocityToModelMaximum(t *testing.T) {
	next := physics.StepShip(
		testShip(1, 0, 0),
		testModel(),
		game.ShipInput{ThrustForward: true},
		10,
	)

	closeTo(t, math.Hypot(next.VelocityX, next.VelocityY), 497)
}

// РџСЂРѕРІРµСЂСЏРµС‚, С‡С‚Рѕ РІРІРѕРґ РёРіСЂРѕРєР° РёР·РјРµРЅСЏРµС‚ С†РµР»РµРІРѕР№ СѓРіРѕР» РїРѕРІРѕСЂРѕС‚Р°.
func TestStepShipUpdatesTargetRotationFromInput(t *testing.T) {
	next := physics.StepShip(
		testShip(1, 0, 0),
		testModel(),
		game.ShipInput{TargetRotationDelta: 0.25},
		0.016,
	)

	closeTo(t, next.TargetRotation, 0.25)
}

// РџСЂРѕРІРµСЂСЏРµС‚, С‡С‚Рѕ РєРѕСЂР°Р±Р»СЊ РЅР°С‡РёРЅР°РµС‚ РІСЂР°С‰Р°С‚СЊСЃСЏ РІ СЃС‚РѕСЂРѕРЅСѓ Р·Р°РґР°РЅРЅРѕРіРѕ С†РµР»РµРІРѕРіРѕ СѓРіР»Р°.
func TestStepShipStartsRotatingTowardTargetRotation(t *testing.T) {
	ship := testShip(1, 0, 0)
	ship.TargetRotation = 1

	next := physics.StepShip(ship, testModel(), idleInput(), 0.05)

	if next.AngularSpeed <= 0 {
		t.Fatalf("got angular velocity %v, want positive", next.AngularSpeed)
	}
}

// РџСЂРѕРІРµСЂСЏРµС‚, С‡С‚Рѕ РѕС€РёР±РєР° СѓРіР»Р° С‡РµСЂРµР· РіСЂР°РЅРёС†Сѓ РїРё РЅРµ РЅРѕСЂРјР°Р»РёР·СѓРµС‚СЃСЏ РІ РєРѕСЂРѕС‚РєРёР№ РїСѓС‚СЊ.
func TestStepShipDoesNotNormalizeAngleErrorAcrossPiBoundary(t *testing.T) {
	ship := testShip(1, 0, 0)
	ship.Rotation = math.Pi - 0.1
	ship.TargetRotation = -math.Pi + 0.1

	next := physics.StepShip(ship, testModel(), idleInput(), 0.05)

	if next.AngularSpeed >= 0 {
		t.Fatalf("got angular velocity %v, want negative", next.AngularSpeed)
	}
}

// РџСЂРѕРІРµСЂСЏРµС‚, С‡С‚Рѕ СЂСЏРґРѕРј СЃ С†РµР»РµРІС‹Рј СѓРіР»РѕРј РІСЂР°С‰РµРЅРёРµ РіР°СЃРёС‚СЃСЏ РґРѕ РѕСЃС‚Р°РЅРѕРІРєРё.
func TestStepShipBrakesAngularVelocityNearTargetRotation(t *testing.T) {
	ship := testShip(1, 0, 0)
	ship.Rotation = 1
	ship.TargetRotation = 1
	ship.AngularSpeed = 0.1

	next := physics.StepShip(ship, testModel(), idleInput(), 1)

	closeTo(t, next.AngularSpeed, 0)
	closeTo(t, next.Rotation, 1)
}

// РџСЂРѕРІРµСЂСЏРµС‚, С‡С‚Рѕ РїРѕРІРѕСЂРѕС‚ РѕСЃС‚Р°РЅР°РІР»РёРІР°РµС‚СЃСЏ СЂРѕРІРЅРѕ РЅР° С†РµР»Рё Р±РµР· РїРµСЂРµР»С‘С‚Р°.
func TestStepShipStopsRotationAtTargetWithoutOvershoot(t *testing.T) {
	ship := testShip(1, 0, 0)
	ship.Rotation = 0
	ship.TargetRotation = 0.01
	ship.AngularSpeed = 1

	next := physics.StepShip(ship, testModel(), idleInput(), 0.05)

	closeTo(t, next.Rotation, 0.01)
	closeTo(t, next.AngularSpeed, 0)
}

// РџСЂРѕРІРµСЂСЏРµС‚, С‡С‚Рѕ РїРµСЂРµРґ С„РёРЅР°Р»СЊРЅРѕР№ РѕСЃС‚Р°РЅРѕРІРєРѕР№ СѓРіР»РѕРІР°СЏ СЃРєРѕСЂРѕСЃС‚СЊ СѓР¶Рµ СѓРјРµРЅСЊС€РµРЅР° РґРѕ РґРѕСЃС‚РёР¶РёРјРѕРіРѕ С€Р°РіР° С‚РѕСЂРјРѕР¶РµРЅРёСЏ.
func TestStepShipReducesAngularVelocityBeforeFinalStopAtTarget(t *testing.T) {
	dtSeconds := 0.05
	ship := testShip(1, 0, 0)
	ship.TargetRotation = 0.5
	maxAngularVelocityChange := ship.MaxTorque / physics.MomentOfInertia(ship, testModel()) * dtSeconds
	current := ship
	angularVelocityBeforeStop := 0.0

	for step := 0; step < 100; step++ {
		next := physics.StepShip(current, testModel(), idleInput(), dtSeconds)

		if next.AngularSpeed == 0 && next.Rotation == next.TargetRotation {
			angularVelocityBeforeStop = math.Abs(current.AngularSpeed)
			break
		}

		current = next
	}

	if angularVelocityBeforeStop <= 0 {
		t.Fatalf("got angular velocity before stop %v, want positive", angularVelocityBeforeStop)
	}
	if angularVelocityBeforeStop > maxAngularVelocityChange+0.000001 {
		t.Fatalf("got angular velocity before stop %v, want at most %v", angularVelocityBeforeStop, maxAngularVelocityChange)
	}
}

// РџСЂРѕРІРµСЂСЏРµС‚, С‡С‚Рѕ РґРІРёР¶РµРЅРёРµ С†РµР»Рё РїРѕРІРѕСЂРѕС‚Р° РЅРµ РѕР±РЅСѓР»СЏРµС‚ С‚РµРєСѓС‰СѓСЋ СѓРіР»РѕРІСѓСЋ СЃРєРѕСЂРѕСЃС‚СЊ РїСЂРµР¶РґРµРІСЂРµРјРµРЅРЅРѕ.
func TestStepShipDoesNotZeroAngularVelocityWhileTargetRotationIsMoving(t *testing.T) {
	dtSeconds := 0.05
	ship := testShip(1, 0, 0)
	ship.Rotation = 0
	ship.TargetRotation = 0.01
	ship.AngularSpeed = 1
	maxAngularVelocityChange := ship.MaxTorque / physics.MomentOfInertia(ship, testModel()) * dtSeconds

	next := physics.StepShip(ship, testModel(), game.ShipInput{TargetRotationDelta: 0.005}, dtSeconds)

	closeTo(t, next.AngularSpeed, 1-maxAngularVelocityChange)
}

// РџСЂРѕРІРµСЂСЏРµС‚, С‡С‚Рѕ СѓРіР»РѕРІР°СЏ СЃРєРѕСЂРѕСЃС‚СЊ РѕРіСЂР°РЅРёС‡РёРІР°РµС‚СЃСЏ РјР°РєСЃРёРјР°Р»СЊРЅС‹Рј Р·РЅР°С‡РµРЅРёРµРј РјРѕРґРµР»Рё.
func TestStepShipClampsAngularVelocityToModelMaximum(t *testing.T) {
	ship := testShip(1, 0, 0)
	ship.TargetRotation = 100

	next := physics.StepShip(ship, testModel(), idleInput(), 10)

	closeTo(t, next.AngularSpeed, 3)
}

// РџСЂРѕРІРµСЂСЏРµС‚, С‡С‚Рѕ РїСЂРё РґРѕСЃС‚РёРіРЅСѓС‚РѕРј С†РµР»РµРІРѕРј СѓРіР»Рµ РіР°СЃРёС‚СЃСЏ Р»РёРЅРµР№РЅРѕРµ РґРІРёР¶РµРЅРёРµ Рё РІСЂР°С‰РµРЅРёРµ.
func TestStepShipBrakesLinearAndAngularVelocityWhenTargetReached(t *testing.T) {
	ship := testShip(1, 0, 0)
	ship.VelocityX = 10
	ship.AngularSpeed = 1

	next := physics.StepShip(ship, testModel(), idleInput(), 1)

	if next.VelocityX >= 10 {
		t.Fatalf("got velocity X %v, want less than 10", next.VelocityX)
	}
	closeTo(t, next.AngularSpeed, 0)
}

// РџСЂРѕРІРµСЂСЏРµС‚, С‡С‚Рѕ Р»РёРЅРµР№РЅРѕРµ С‚РѕСЂРјРѕР¶РµРЅРёРµ РїСЂРѕРґРѕР»Р¶Р°РµС‚СЃСЏ РІРѕ РІСЂРµРјСЏ СЂР°Р·РІРѕСЂРѕС‚Р° Рє С†РµР»Рё.
func TestStepShipKeepsLinearBrakingWhileRotatingToTarget(t *testing.T) {
	ship := testShip(1, 0, 0)
	ship.VelocityX = 200
	ship.TargetRotation = 1

	next := physics.StepShip(ship, testModel(), idleInput(), 0.05)

	if next.VelocityX >= 200 {
		t.Fatalf("got velocity X %v, want less than 200", next.VelocityX)
	}
	if next.AngularSpeed <= 0 {
		t.Fatalf("got angular velocity %v, want positive", next.AngularSpeed)
	}
}

// РџСЂРѕРІРµСЂСЏРµС‚, С‡С‚Рѕ РїСЂРѕРґРѕР»СЊРЅР°СЏ С‚СЏРіР° РіР°СЃРёС‚ РїРѕРїРµСЂРµС‡РЅСѓСЋ СЃРєРѕСЂРѕСЃС‚СЊ Рё СЃРѕР·РґР°С‘С‚ РґРІРёР¶РµРЅРёРµ РІРїРµСЂС‘Рґ.
func TestStepShipBrakesAcrossVelocityDuringAlongThrust(t *testing.T) {
	ship := testShip(1, 0, 0)
	ship.VelocityX = 100

	next := physics.StepShip(ship, testModel(), game.ShipInput{ThrustForward: true}, 0.1)

	if next.VelocityX >= 100 {
		t.Fatalf("got velocity X %v, want less than 100", next.VelocityX)
	}
	if next.VelocityY <= 0 {
		t.Fatalf("got velocity Y %v, want positive", next.VelocityY)
	}
}

// РџСЂРѕРІРµСЂСЏРµС‚, С‡С‚Рѕ РїРѕРїРµСЂРµС‡РЅР°СЏ С‚СЏРіР° РіР°СЃРёС‚ РїСЂРѕРґРѕР»СЊРЅСѓСЋ СЃРєРѕСЂРѕСЃС‚СЊ Рё СЃРѕР·РґР°С‘С‚ Р±РѕРєРѕРІРѕРµ РґРІРёР¶РµРЅРёРµ.
func TestStepShipBrakesAlongVelocityDuringAcrossThrust(t *testing.T) {
	ship := testShip(1, 0, 0)
	ship.VelocityY = 100

	next := physics.StepShip(ship, testModel(), game.ShipInput{ThrustRight: true}, 0.1)

	if next.VelocityX <= 0 {
		t.Fatalf("got velocity X %v, want positive", next.VelocityX)
	}
	if next.VelocityY >= 100 {
		t.Fatalf("got velocity Y %v, want less than 100", next.VelocityY)
	}
}

// РџСЂРѕРІРµСЂСЏРµС‚, С‡С‚Рѕ РѕРґРЅРѕРІСЂРµРјРµРЅРЅР°СЏ РІСЃС‚СЂРµС‡РЅР°СЏ РїСЂРѕРґРѕР»СЊРЅР°СЏ С‚СЏРіР° РЅРµ РІРєР»СЋС‡Р°РµС‚ Р°РІС‚РѕС‚РѕСЂРјРѕР¶РµРЅРёРµ РїРѕ СЌС‚РѕР№ РѕСЃРё.
func TestStepShipDoesNotAutobrakeAlongAxisWhenForwardAndBackwardPressed(t *testing.T) {
	ship := testShip(1, 0, 0)
	ship.VelocityY = 100

	next := physics.StepShip(ship, testModel(), game.ShipInput{ThrustForward: true, ThrustBackward: true}, 0.1)

	closeTo(t, next.VelocityX, 0)
	closeTo(t, next.VelocityY, 100)
}

// РџСЂРѕРІРµСЂСЏРµС‚, С‡С‚Рѕ РѕРґРЅРѕРІСЂРµРјРµРЅРЅР°СЏ РІСЃС‚СЂРµС‡РЅР°СЏ РїРѕРїРµСЂРµС‡РЅР°СЏ С‚СЏРіР° РЅРµ РІРєР»СЋС‡Р°РµС‚ Р°РІС‚РѕС‚РѕСЂРјРѕР¶РµРЅРёРµ РїРѕ СЌС‚РѕР№ РѕСЃРё.
func TestStepShipDoesNotAutobrakeAcrossAxisWhenLeftAndRightPressed(t *testing.T) {
	ship := testShip(1, 0, 0)
	ship.VelocityX = 100

	next := physics.StepShip(ship, testModel(), game.ShipInput{ThrustLeft: true, ThrustRight: true}, 0.1)

	closeTo(t, next.VelocityX, 100)
	closeTo(t, next.VelocityY, 0)
}

// РџСЂРѕРІРµСЂСЏРµС‚, С‡С‚Рѕ РєРѕСЂР°Р±Р»СЊ Р±РµР· РїРёР»РѕС‚Р° С‚РѕСЂРјРѕР·РёС‚ СЃ РїРѕСЃС‚РѕСЏРЅРЅС‹Рј Р·Р°РјРµРґР»РµРЅРёРµРј Рё РѕР±РЅРѕРІР»СЏРµС‚ РїРѕР»РѕР¶РµРЅРёРµ.
func TestStepUnpilotedShipAppliesConstantBrake(t *testing.T) {
	ship := testShip(1, 0, 0)
	ship.Mass = 1000
	ship.MaxAlongForce = 100000
	ship.VelocityX = 100

	next := physics.StepUnpilotedShip(ship, 0.1)

	closeTo(t, next.VelocityX, 90)
	closeTo(t, next.VelocityY, 0)
	closeTo(t, next.Speed, 90)
	closeTo(t, next.X, 9)
}

// РџСЂРѕРІРµСЂСЏРµС‚, С‡С‚Рѕ С‚РѕСЂРјРѕР¶РµРЅРёРµ РєРѕСЂР°Р±Р»СЏ Р±РµР· РїРёР»РѕС‚Р° РЅРµ Р·Р°РІРёСЃРёС‚ РѕС‚ С‚РµРєСѓС‰РµР№ СЃРєРѕСЂРѕСЃС‚Рё.
func TestStepUnpilotedShipBrakeDoesNotDependOnSpeed(t *testing.T) {
	slowShip := testShip(1, 0, 0)
	slowShip.Mass = 1000
	slowShip.MaxAlongForce = 100000
	slowShip.VelocityX = 100
	fastShip := slowShip
	fastShip.ID = 2
	fastShip.VelocityX = 200

	nextSlow := physics.StepUnpilotedShip(slowShip, 0.1)
	nextFast := physics.StepUnpilotedShip(fastShip, 0.1)

	closeTo(t, slowShip.VelocityX-nextSlow.VelocityX, 10)
	closeTo(t, fastShip.VelocityX-nextFast.VelocityX, 10)
}

// РџСЂРѕРІРµСЂСЏРµС‚, С‡С‚Рѕ РїРѕСЃС‚РѕСЏРЅРЅРѕРµ С‚РѕСЂРјРѕР¶РµРЅРёРµ СЃРѕС…СЂР°РЅСЏРµС‚ РЅР°РїСЂР°РІР»РµРЅРёРµ РІРµРєС‚РѕСЂР° СЃРєРѕСЂРѕСЃС‚Рё.
func TestStepUnpilotedShipKeepsVelocityDirectionDuringConstantBrake(t *testing.T) {
	ship := testShip(1, 0, 0)
	ship.Mass = 1000
	ship.MaxAlongForce = 100000
	ship.VelocityX = 60
	ship.VelocityY = 80

	next := physics.StepUnpilotedShip(ship, 0.1)

	closeTo(t, next.VelocityX, 54)
	closeTo(t, next.VelocityY, 72)
	closeTo(t, next.Speed, 90)
}

// РџСЂРѕРІРµСЂСЏРµС‚, С‡С‚Рѕ РїРѕСЃС‚РѕСЏРЅРЅРѕРµ С‚РѕСЂРјРѕР¶РµРЅРёРµ РЅРµ СЂР°Р·РІРѕСЂР°С‡РёРІР°РµС‚ СЃРєРѕСЂРѕСЃС‚СЊ РІ РїСЂРѕС‚РёРІРѕРїРѕР»РѕР¶РЅСѓСЋ СЃС‚РѕСЂРѕРЅСѓ.
func TestStepUnpilotedShipDoesNotReverseVelocityDuringConstantBrake(t *testing.T) {
	ship := testShip(1, 0, 0)
	ship.Mass = 1000
	ship.MaxAlongForce = 100000
	ship.VelocityX = 10

	next := physics.StepUnpilotedShip(ship, 10)

	closeTo(t, next.VelocityX, 0)
	closeTo(t, next.VelocityY, 0)
	closeTo(t, next.Speed, 0)
	closeTo(t, next.X, 0)
}

// РџСЂРѕРІРµСЂСЏРµС‚, С‡С‚Рѕ РєРѕСЂР°Р±Р»СЊ Р±РµР· РїРёР»РѕС‚Р° РїРѕР»СѓС‡Р°РµС‚ СѓРіР»РѕРІРѕРµ С‚РѕСЂРјРѕР¶РµРЅРёРµ.
func TestStepUnpilotedShipAppliesAngularBrake(t *testing.T) {
	ship := testShip(1, 0, 0)
	ship.AngularSpeed = 3

	next := physics.StepUnpilotedShip(ship, 0.5)

	closeTo(t, next.AngularSpeed, 2.5)
}

// РџСЂРѕРІРµСЂСЏРµС‚, С‡С‚Рѕ СѓРіР»РѕРІРѕРµ С‚РѕСЂРјРѕР¶РµРЅРёРµ РЅРµ СЂР°Р·РІРѕСЂР°С‡РёРІР°РµС‚ РІСЂР°С‰РµРЅРёРµ РІ РїСЂРѕС‚РёРІРѕРїРѕР»РѕР¶РЅСѓСЋ СЃС‚РѕСЂРѕРЅСѓ.
func TestStepUnpilotedShipDoesNotReverseAngularVelocityDuringBrake(t *testing.T) {
	ship := testShip(1, 0, 0)
	ship.AngularSpeed = -0.25

	next := physics.StepUnpilotedShip(ship, 1)

	closeTo(t, next.AngularSpeed, 0)
}

// РџСЂРѕРІРµСЂСЏРµС‚, С‡С‚Рѕ СЃС‚РѕР»РєРЅРѕРІРµРЅРёРµ СЃ Р·Р°РєСЂРµРїР»С‘РЅРЅС‹Рј С‚РµР»РѕРј РѕС‚СЂР°Р¶Р°РµС‚ СЃРєРѕСЂРѕСЃС‚СЊ РІРґРѕР»СЊ РЅРѕСЂРјР°Р»Рё СЃ СѓРїСЂСѓРіРѕСЃС‚СЊСЋ.
func TestApplyCollisionResponseBouncesAlongNormalWithRestitution(t *testing.T) {
	ship := testShip(1, 0, 0)
	ship.Enabled = true
	ship.VelocityX = -10
	ship.VelocityY = 5
	obstacle := testShip(2, 0, 0)
	obstacle.Enabled = true
	obstacle.Anchored = true

	next, nextObstacle := physics.ApplyCollisionResponse(ship, testModel(), obstacle, testModel(), physics.Collision{
		Correction:   game.WorldVector{X: 2, Y: 0},
		ContactPoint: game.WorldVector{},
	})

	closeTo(t, next.X, 2)
	closeTo(t, nextObstacle.X, 0)
	closeTo(t, next.VelocityX, 5)
	closeTo(t, next.VelocityY, 5)
	closeTo(t, next.Speed, math.Hypot(5, 5))
}

// РџСЂРѕРІРµСЂСЏРµС‚, С‡С‚Рѕ СЃС‚РѕР»РєРЅРѕРІРµРЅРёРµ РґРІСѓС… РїРѕРґРІРёР¶РЅС‹С… С‚РµР» СЂР°Р·РґРµР»СЏРµС‚ РєРѕСЂСЂРµРєС†РёСЋ Рё РёРјРїСѓР»СЊСЃ СЃ СѓС‡С‘С‚РѕРј РјР°СЃСЃ.
func TestApplyCollisionResponseUsesMassesForMovableBodies(t *testing.T) {
	ship := testShip(1, 0, 0)
	ship.Enabled = true
	ship.Mass = 1
	ship.VelocityX = -10
	obstacle := testShip(2, 0, 0)
	obstacle.Enabled = true
	obstacle.Mass = 1

	next, nextObstacle := physics.ApplyCollisionResponse(ship, testModel(), obstacle, testModel(), physics.Collision{
		Correction:   game.WorldVector{X: 2, Y: 0},
		ContactPoint: game.WorldVector{},
	})

	closeTo(t, next.X, 1)
	closeTo(t, nextObstacle.X, -1)
	closeTo(t, next.VelocityX, -2.5)
	closeTo(t, nextObstacle.VelocityX, -7.5)
}

// РџСЂРѕРІРµСЂСЏРµС‚, С‡С‚Рѕ РЅРµС†РµРЅС‚СЂР°Р»СЊРЅС‹Р№ РєРѕРЅС‚Р°РєС‚ РґРѕР±Р°РІР»СЏРµС‚ СѓРіР»РѕРІСѓСЋ СЃРєРѕСЂРѕСЃС‚СЊ СѓРґР°СЂРёРІС€РµРјСѓСЃСЏ С‚РµР»Сѓ.
func TestApplyCollisionResponseAddsAngularSpeedFromOffCenterContact(t *testing.T) {
	ship := testShip(1, 0, 0)
	ship.Enabled = true
	ship.Mass = 1000
	ship.MaxAngularSpeed = 10
	ship.VelocityX = -10
	obstacle := testShip(2, 0, 0)
	obstacle.Enabled = true
	obstacle.Anchored = true
	collision := physics.Collision{
		Correction:   game.WorldVector{X: 2, Y: 0},
		ContactPoint: game.WorldVector{X: 0, Y: 10},
	}

	next, nextObstacle := physics.ApplyCollisionResponse(ship, testModel(), obstacle, testModel(), collision)

	if next.AngularSpeed <= 0 {
		t.Fatalf("got angular velocity %v, want positive", next.AngularSpeed)
	}
	closeTo(t, nextObstacle.AngularSpeed, 0)
}

// РџСЂРѕРІРµСЂСЏРµС‚, С‡С‚Рѕ РїРѕСЃР»Рµ РІСЂР°С‰Р°С‚РµР»СЊРЅРѕРіРѕ РѕС‚СЃРєРѕРєР° С†РµР»РµРІРѕР№ СѓРіРѕР» РїРµСЂРµРЅРѕСЃРёС‚СЃСЏ РЅР° СѓРіРѕР» РѕСЃС‚Р°РЅРѕРІРєРё.
func TestApplyCollisionResponseMovesTargetRotationByAngularBounceStopAngle(t *testing.T) {
	ship := testShip(1, 0, 0)
	ship.Enabled = true
	ship.Mass = 1000
	ship.MaxAngularSpeed = 10
	ship.MaxTorque = 10000
	ship.VelocityX = -10
	obstacle := testShip(2, 0, 0)
	obstacle.Enabled = true
	obstacle.Anchored = true
	collision := physics.Collision{
		Correction:   game.WorldVector{X: 2, Y: 0},
		ContactPoint: game.WorldVector{X: 0, Y: 10},
	}

	next, _ := physics.ApplyCollisionResponse(ship, testModel(), obstacle, testModel(), collision)
	angularStopAngle := next.AngularSpeed * next.AngularSpeed / (2 * ship.MaxTorque / physics.MomentOfInertia(ship, testModel()))

	closeTo(t, next.TargetRotation, angularStopAngle)
}
