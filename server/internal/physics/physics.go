package physics

import (
	"math"

	"space-game-07-server/internal/game"
)

const (
	Epsilon      = 0.000001
	angleEpsilon = 0.0001
)

func BodySizeMeters(model game.CosmicObjectModel) game.WorldVector {
	return game.WorldVector{
		X: float64(model.TextureBodyWidth) / model.TextureScale,
		Y: float64(model.TextureBodyLength) / model.TextureScale,
	}
}

func MomentOfInertia(ship game.WorldObject) float64 {
	body := BodySizeMeters(ship.Model)

	return ship.Model.MassKg * (body.X*body.X + body.Y*body.Y) / 16
}

func ForwardVector(rotation float64) game.WorldVector {
	return game.WorldVector{
		X: math.Sin(rotation),
		Y: math.Cos(rotation),
	}
}

func RightVector(rotation float64) game.WorldVector {
	return game.WorldVector{
		X: math.Cos(rotation),
		Y: -math.Sin(rotation),
	}
}

func brakeValue(value float64, acceleration float64, dtSeconds float64) float64 {
	delta := acceleration * dtSeconds

	if math.Abs(value) <= delta {
		return 0
	}

	return value - math.Copysign(delta, value)
}

func clampVectorLength(x float64, y float64, maxLength float64) game.WorldVector {
	length := math.Hypot(x, y)

	if length <= maxLength || length <= Epsilon {
		return game.WorldVector{X: x, Y: y}
	}

	scale := maxLength / length

	return game.WorldVector{X: x * scale, Y: y * scale}
}

func clampAbsoluteValue(value float64, maxAbsoluteValue float64) float64 {
	return math.Copysign(math.Min(math.Abs(value), maxAbsoluteValue), value)
}

func projection(vector game.WorldVector, axis game.WorldVector) float64 {
	return vector.X*axis.X + vector.Y*axis.Y
}

func angularAcceleration(ship game.WorldObject) float64 {
	return ship.Model.TorqueNm / MomentOfInertia(ship)
}

func stepAngularVelocityToTarget(ship game.WorldObject, targetRotation float64, dtSeconds float64) (float64, float64) {
	acceleration := angularAcceleration(ship)
	angleError := targetRotation - ship.Rotation

	if math.Abs(angleError) <= angleEpsilon {
		angularVelocity := brakeValue(ship.AngularVelocity, acceleration, dtSeconds)

		if angularVelocity == 0 {
			return targetRotation, angularVelocity
		}

		return ship.Rotation + angularVelocity*dtSeconds, angularVelocity
	}

	directionToTarget := math.Copysign(1, angleError)
	currentDirection := math.Copysign(1, ship.AngularVelocity)
	if ship.AngularVelocity == 0 {
		currentDirection = 0
	}

	stoppingDistance := ship.AngularVelocity * ship.AngularVelocity / (2 * acceleration)
	shouldBrake := currentDirection != 0 &&
		currentDirection == directionToTarget &&
		stoppingDistance >= math.Abs(angleError)
	torqueDirection := directionToTarget
	if shouldBrake {
		torqueDirection = -currentDirection
	}

	angularVelocity := clampAbsoluteValue(
		ship.AngularVelocity+torqueDirection*acceleration*dtSeconds,
		ship.Model.MaxAngularSpeedRadPerSecond,
	)

	return ship.Rotation + angularVelocity*dtSeconds, angularVelocity
}

func StepShip(ship game.WorldObject, input game.ShipInput, dtSeconds float64) game.WorldObject {
	forward := ForwardVector(ship.Rotation)
	right := RightVector(ship.Rotation)
	hasAlongControl := input.ThrustForward || input.ThrustBackward
	hasAcrossControl := input.ThrustLeft || input.ThrustRight
	along := 0.0
	across := 0.0
	if input.ThrustForward {
		along++
	}
	if input.ThrustBackward {
		along--
	}
	if input.ThrustRight {
		across++
	}
	if input.ThrustLeft {
		across--
	}

	linearAcceleration := ship.Model.ThrustN / ship.Model.MassKg
	alongVelocity := 0.0
	if hasAlongControl {
		alongVelocity = projection(ship.Velocity, forward) + along*linearAcceleration*dtSeconds
	} else {
		alongVelocity = brakeValue(projection(ship.Velocity, forward), linearAcceleration, dtSeconds)
	}

	acrossVelocity := 0.0
	if hasAcrossControl {
		acrossVelocity = projection(ship.Velocity, right) + across*linearAcceleration*dtSeconds
	} else {
		acrossVelocity = brakeValue(projection(ship.Velocity, right), linearAcceleration, dtSeconds)
	}

	velocityX := forward.X*alongVelocity + right.X*acrossVelocity
	velocityY := forward.Y*alongVelocity + right.Y*acrossVelocity
	limitedVelocity := clampVectorLength(velocityX, velocityY, ship.Model.MaxSpeedMps)
	targetRotation := ship.TargetRotation + input.TargetRotationDelta
	rotation, angularVelocity := stepAngularVelocityToTarget(ship, targetRotation, dtSeconds)

	ship.Position = game.WorldVector{
		X: ship.Position.X + limitedVelocity.X*dtSeconds,
		Y: ship.Position.Y + limitedVelocity.Y*dtSeconds,
	}
	ship.Velocity = limitedVelocity
	ship.Rotation = rotation
	ship.AngularVelocity = angularVelocity
	ship.TargetRotation = targetRotation

	return ship
}
