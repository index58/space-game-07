package physics_test

import (
	"math"
	"testing"

	"space-game-07-server/internal/data"
	"space-game-07-server/internal/game"
	"space-game-07-server/internal/physics"
)

// Возвращает пустой ввод для сценариев, где проверяется инерция или торможение.
func idleInput() game.ShipInput {
	return game.ShipInput{}
}

// Создаёт объект с устойчивыми параметрами, чтобы физические ожидания были воспроизводимыми.
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

// Создаёт модель с размерами физического тела в пикселях текстуры.
func testModel() data.CosmicObjectModel {
	return data.CosmicObjectModel{
		ID:                1,
		TextureBodyWidth:  88,
		TextureBodyLength: 90,
		TextureScale:      4,
	}
}

// Сравнивает float64 с малым допуском, потому что физика работает с дробными величинами.
func closeTo(t *testing.T, actual float64, expected float64) {
	t.Helper()

	if math.Abs(actual-expected) > 0.000001 {
		t.Fatalf("got %v, want %v", actual, expected)
	}
}

func TestMomentOfInertiaUsesFilledEllipseApproximation(t *testing.T) {
	closeTo(t, physics.MomentOfInertia(testShip(1, 0, 0), testModel()), 490173.75)
}

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

func TestStepShipClampsLinearVelocityToModelMaximum(t *testing.T) {
	next := physics.StepShip(
		testShip(1, 0, 0),
		testModel(),
		game.ShipInput{ThrustForward: true},
		10,
	)

	closeTo(t, math.Hypot(next.VelocityX, next.VelocityY), 497)
}

func TestStepShipUpdatesTargetRotationFromInput(t *testing.T) {
	next := physics.StepShip(
		testShip(1, 0, 0),
		testModel(),
		game.ShipInput{TargetRotationDelta: 0.25},
		0.016,
	)

	closeTo(t, next.TargetRotation, 0.25)
}

func TestStepShipStartsRotatingTowardTargetRotation(t *testing.T) {
	ship := testShip(1, 0, 0)
	ship.TargetRotation = 1

	next := physics.StepShip(ship, testModel(), idleInput(), 0.05)

	if next.AngularSpeed <= 0 {
		t.Fatalf("got angular velocity %v, want positive", next.AngularSpeed)
	}
}

func TestStepShipDoesNotNormalizeAngleErrorAcrossPiBoundary(t *testing.T) {
	ship := testShip(1, 0, 0)
	ship.Rotation = math.Pi - 0.1
	ship.TargetRotation = -math.Pi + 0.1

	next := physics.StepShip(ship, testModel(), idleInput(), 0.05)

	if next.AngularSpeed >= 0 {
		t.Fatalf("got angular velocity %v, want negative", next.AngularSpeed)
	}
}

func TestStepShipBrakesAngularVelocityNearTargetRotation(t *testing.T) {
	ship := testShip(1, 0, 0)
	ship.Rotation = 1
	ship.TargetRotation = 1
	ship.AngularSpeed = 0.1

	next := physics.StepShip(ship, testModel(), idleInput(), 1)

	closeTo(t, next.AngularSpeed, 0)
	closeTo(t, next.Rotation, 1)
}

func TestStepShipStopsRotationAtTargetWithoutOvershoot(t *testing.T) {
	ship := testShip(1, 0, 0)
	ship.Rotation = 0
	ship.TargetRotation = 0.01
	ship.AngularSpeed = 1

	next := physics.StepShip(ship, testModel(), idleInput(), 0.05)

	closeTo(t, next.Rotation, 0.01)
	closeTo(t, next.AngularSpeed, 0)
}

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

func TestStepShipClampsAngularVelocityToModelMaximum(t *testing.T) {
	ship := testShip(1, 0, 0)
	ship.TargetRotation = 100

	next := physics.StepShip(ship, testModel(), idleInput(), 10)

	closeTo(t, next.AngularSpeed, 3)
}

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

func TestStepShipDoesNotAutobrakeAlongAxisWhenForwardAndBackwardPressed(t *testing.T) {
	ship := testShip(1, 0, 0)
	ship.VelocityY = 100

	next := physics.StepShip(ship, testModel(), game.ShipInput{ThrustForward: true, ThrustBackward: true}, 0.1)

	closeTo(t, next.VelocityX, 0)
	closeTo(t, next.VelocityY, 100)
}

func TestStepShipDoesNotAutobrakeAcrossAxisWhenLeftAndRightPressed(t *testing.T) {
	ship := testShip(1, 0, 0)
	ship.VelocityX = 100

	next := physics.StepShip(ship, testModel(), game.ShipInput{ThrustLeft: true, ThrustRight: true}, 0.1)

	closeTo(t, next.VelocityX, 100)
	closeTo(t, next.VelocityY, 0)
}

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
