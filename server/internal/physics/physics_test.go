package physics_test

import (
	"math"
	"testing"

	"space-game-07-server/internal/game"
	"space-game-07-server/internal/physics"
)

func idleInput() game.ShipInput {
	return game.ShipInput{}
}

func closeTo(t *testing.T, actual float64, expected float64) {
	t.Helper()

	if math.Abs(actual-expected) > 0.000001 {
		t.Fatalf("got %v, want %v", actual, expected)
	}
}

func TestMomentOfInertiaUsesFilledEllipseApproximation(t *testing.T) {
	closeTo(t, physics.MomentOfInertia(game.NewPlayerShip(1, game.WorldVector{})), 490173.75)
}

func TestStepShipAcceleratesForwardAlongPositiveYAtZeroRotation(t *testing.T) {
	next := physics.StepShip(
		game.NewPlayerShip(1, game.WorldVector{}),
		game.ShipInput{ThrustForward: true},
		1,
	)

	closeTo(t, next.Velocity.X, 0)
	closeTo(t, next.Velocity.Y, 162.61382953137096)
}

func TestStepShipClampsLinearVelocityToModelMaximum(t *testing.T) {
	next := physics.StepShip(
		game.NewPlayerShip(1, game.WorldVector{}),
		game.ShipInput{ThrustForward: true},
		10,
	)

	closeTo(t, math.Hypot(next.Velocity.X, next.Velocity.Y), 497)
}

func TestStepShipUpdatesTargetRotationFromInput(t *testing.T) {
	next := physics.StepShip(
		game.NewPlayerShip(1, game.WorldVector{}),
		game.ShipInput{TargetRotationDelta: 0.25},
		0.016,
	)

	closeTo(t, next.TargetRotation, 0.25)
}

func TestStepShipStartsRotatingTowardTargetRotation(t *testing.T) {
	ship := game.NewPlayerShip(1, game.WorldVector{})
	ship.TargetRotation = 1

	next := physics.StepShip(ship, idleInput(), 0.05)

	if next.AngularVelocity <= 0 {
		t.Fatalf("got angular velocity %v, want positive", next.AngularVelocity)
	}
}

func TestStepShipDoesNotNormalizeAngleErrorAcrossPiBoundary(t *testing.T) {
	ship := game.NewPlayerShip(1, game.WorldVector{})
	ship.Rotation = math.Pi - 0.1
	ship.TargetRotation = -math.Pi + 0.1

	next := physics.StepShip(ship, idleInput(), 0.05)

	if next.AngularVelocity >= 0 {
		t.Fatalf("got angular velocity %v, want negative", next.AngularVelocity)
	}
}

func TestStepShipBrakesAngularVelocityNearTargetRotation(t *testing.T) {
	ship := game.NewPlayerShip(1, game.WorldVector{})
	ship.Rotation = 1
	ship.TargetRotation = 1
	ship.AngularVelocity = 0.1

	next := physics.StepShip(ship, idleInput(), 1)

	closeTo(t, next.AngularVelocity, 0)
	closeTo(t, next.Rotation, 1)
}

func TestStepShipStopsRotationAtTargetWithoutOvershoot(t *testing.T) {
	ship := game.NewPlayerShip(1, game.WorldVector{})
	ship.Rotation = 0
	ship.TargetRotation = 0.01
	ship.AngularVelocity = 1

	next := physics.StepShip(ship, idleInput(), 0.05)

	closeTo(t, next.Rotation, 0.01)
	closeTo(t, next.AngularVelocity, 0)
}

func TestStepShipReducesAngularVelocityBeforeFinalStopAtTarget(t *testing.T) {
	dtSeconds := 0.05
	ship := game.NewPlayerShip(1, game.WorldVector{})
	ship.TargetRotation = 0.5
	maxAngularVelocityChange := ship.Model.TorqueNm / physics.MomentOfInertia(ship) * dtSeconds
	current := ship
	angularVelocityBeforeStop := 0.0

	for step := 0; step < 100; step++ {
		next := physics.StepShip(current, idleInput(), dtSeconds)

		if next.AngularVelocity == 0 && next.Rotation == next.TargetRotation {
			angularVelocityBeforeStop = math.Abs(current.AngularVelocity)
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
	ship := game.NewPlayerShip(1, game.WorldVector{})
	ship.Rotation = 0
	ship.TargetRotation = 0.01
	ship.AngularVelocity = 1
	maxAngularVelocityChange := ship.Model.TorqueNm / physics.MomentOfInertia(ship) * dtSeconds

	next := physics.StepShip(ship, game.ShipInput{TargetRotationDelta: 0.005}, dtSeconds)

	closeTo(t, next.AngularVelocity, 1-maxAngularVelocityChange)
}

func TestStepShipClampsAngularVelocityToModelMaximum(t *testing.T) {
	ship := game.NewPlayerShip(1, game.WorldVector{})
	ship.TargetRotation = 100

	next := physics.StepShip(ship, idleInput(), 10)

	closeTo(t, next.AngularVelocity, 3)
}

func TestStepShipBrakesLinearAndAngularVelocityWhenTargetReached(t *testing.T) {
	ship := game.NewPlayerShip(1, game.WorldVector{})
	ship.Velocity = game.WorldVector{X: 10}
	ship.AngularVelocity = 1

	next := physics.StepShip(ship, idleInput(), 1)

	if next.Velocity.X >= 10 {
		t.Fatalf("got velocity X %v, want less than 10", next.Velocity.X)
	}
	closeTo(t, next.AngularVelocity, 0)
}

func TestStepShipKeepsLinearBrakingWhileRotatingToTarget(t *testing.T) {
	ship := game.NewPlayerShip(1, game.WorldVector{})
	ship.Velocity = game.WorldVector{X: 200}
	ship.TargetRotation = 1

	next := physics.StepShip(ship, idleInput(), 0.05)

	if next.Velocity.X >= 200 {
		t.Fatalf("got velocity X %v, want less than 200", next.Velocity.X)
	}
	if next.AngularVelocity <= 0 {
		t.Fatalf("got angular velocity %v, want positive", next.AngularVelocity)
	}
}

func TestStepShipBrakesAcrossVelocityDuringAlongThrust(t *testing.T) {
	ship := game.NewPlayerShip(1, game.WorldVector{})
	ship.Velocity = game.WorldVector{X: 100}

	next := physics.StepShip(ship, game.ShipInput{ThrustForward: true}, 0.1)

	if next.Velocity.X >= 100 {
		t.Fatalf("got velocity X %v, want less than 100", next.Velocity.X)
	}
	if next.Velocity.Y <= 0 {
		t.Fatalf("got velocity Y %v, want positive", next.Velocity.Y)
	}
}

func TestStepShipBrakesAlongVelocityDuringAcrossThrust(t *testing.T) {
	ship := game.NewPlayerShip(1, game.WorldVector{})
	ship.Velocity = game.WorldVector{Y: 100}

	next := physics.StepShip(ship, game.ShipInput{ThrustRight: true}, 0.1)

	if next.Velocity.X <= 0 {
		t.Fatalf("got velocity X %v, want positive", next.Velocity.X)
	}
	if next.Velocity.Y >= 100 {
		t.Fatalf("got velocity Y %v, want less than 100", next.Velocity.Y)
	}
}

func TestStepShipDoesNotAutobrakeAlongAxisWhenForwardAndBackwardPressed(t *testing.T) {
	ship := game.NewPlayerShip(1, game.WorldVector{})
	ship.Velocity = game.WorldVector{Y: 100}

	next := physics.StepShip(ship, game.ShipInput{ThrustForward: true, ThrustBackward: true}, 0.1)

	closeTo(t, next.Velocity.X, 0)
	closeTo(t, next.Velocity.Y, 100)
}

func TestStepShipDoesNotAutobrakeAcrossAxisWhenLeftAndRightPressed(t *testing.T) {
	ship := game.NewPlayerShip(1, game.WorldVector{})
	ship.Velocity = game.WorldVector{X: 100}

	next := physics.StepShip(ship, game.ShipInput{ThrustLeft: true, ThrustRight: true}, 0.1)

	closeTo(t, next.Velocity.X, 100)
	closeTo(t, next.Velocity.Y, 0)
}
