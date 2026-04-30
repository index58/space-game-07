package physics

import (
	"math"

	"space-game-07-server/internal/game"
)

const (
	// защищает вычисления от дрожания около нуля при работе с float64.
	Epsilon = 0.000001
	// задает допуск, при котором корабль считается почти повернутым к цели.
	angleEpsilon = 0.0001
)

// переводит пиксельный размер физического тела модели в метры мира.
func BodySizeMeters(model game.CosmicObjectModel) game.WorldVector {
	return game.WorldVector{
		X: float64(model.TextureBodyWidth) / model.TextureScale,
		Y: float64(model.TextureBodyLength) / model.TextureScale,
	}
}

// оценивает момент инерции корабля как прямоугольного тела.
func MomentOfInertia(ship game.WorldObject) float64 {
	body := BodySizeMeters(ship.Model)

	return ship.Model.MassKg * (body.X*body.X + body.Y*body.Y) / 16
}

// возвращает локальную ось "вперед" для текущего угла корабля.
func ForwardVector(rotation float64) game.WorldVector {
	return game.WorldVector{
		X: math.Sin(rotation),
		Y: math.Cos(rotation),
	}
}

// возвращает локальную ось "вправо" для текущего угла корабля.
func RightVector(rotation float64) game.WorldVector {
	return game.WorldVector{
		X: math.Cos(rotation),
		Y: -math.Sin(rotation),
	}
}

// уменьшает значение к нулю с заданным ускорением без смены знака.
func brakeValue(value float64, acceleration float64, dtSeconds float64) float64 {
	delta := acceleration * dtSeconds

	if math.Abs(value) <= delta {
		return 0
	}

	return value - math.Copysign(delta, value)
}

// ограничивает длину вектора максимальной скоростью.
func clampVectorLength(x float64, y float64, maxLength float64) game.WorldVector {
	length := math.Hypot(x, y)

	if length <= maxLength || length <= Epsilon {
		return game.WorldVector{X: x, Y: y}
	}

	scale := maxLength / length

	return game.WorldVector{X: x * scale, Y: y * scale}
}

// ограничивает модуль скалярного значения и зануляет микродрожание.
func clampAbsoluteValue(value float64, maxAbsoluteValue float64) float64 {
	if math.Abs(value) <= Epsilon {
		return 0
	}

	return math.Copysign(math.Min(math.Abs(value), maxAbsoluteValue), value)
}

// двигает значение к цели не дальше разрешенной дельты за шаг.
func moveToward(value float64, target float64, maxDelta float64) float64 {
	delta := target - value

	if math.Abs(delta) <= maxDelta {
		return target
	}

	return value + math.Copysign(maxDelta, delta)
}

// возвращает скалярную проекцию вектора на заданную ось.
func projection(vector game.WorldVector, axis game.WorldVector) float64 {
	return vector.X*axis.X + vector.Y*axis.Y
}

// считает доступное угловое ускорение от крутящего момента модели.
func angularAcceleration(ship game.WorldObject) float64 {
	return ship.Model.TorqueNm / MomentOfInertia(ship)
}

// ведет угол к целевому так, чтобы успеть затормозить без перелета.
func stepAngularVelocityToTarget(ship game.WorldObject, targetRotation float64, targetRotationSpeed float64, dtSeconds float64) (float64, float64) {
	acceleration := angularAcceleration(ship)
	angleError := targetRotation - ship.Rotation
	isTargetRotationStopped := math.Abs(targetRotationSpeed) <= Epsilon

	if math.Abs(angleError) <= angleEpsilon {
		angularVelocity := brakeValue(ship.AngularVelocity, acceleration, dtSeconds)

		if angularVelocity == 0 {
			return targetRotation, angularVelocity
		}

		return ship.Rotation + angularVelocity*dtSeconds, angularVelocity
	}

	directionToTarget := math.Copysign(1, angleError)
	distanceToTarget := math.Abs(angleError)
	maxVelocityDelta := acceleration * dtSeconds
	velocityTowardTarget := ship.AngularVelocity * directionToTarget

	safeSpeed := math.Max(
		0,
		math.Sqrt(maxVelocityDelta*maxVelocityDelta+2*acceleration*distanceToTarget)-maxVelocityDelta,
	)
	desiredSpeed := math.Min(safeSpeed, ship.Model.MaxAngularSpeedRadPerSecond)
	if velocityTowardTarget < 0 {
		desiredSpeed = 0
	}
	desiredAngularVelocity := directionToTarget * desiredSpeed
	rawAngularVelocity := moveToward(ship.AngularVelocity, desiredAngularVelocity, maxVelocityDelta)
	angularVelocity := clampAbsoluteValue(rawAngularVelocity, ship.Model.MaxAngularSpeedRadPerSecond)
	rotation := ship.Rotation + angularVelocity*dtSeconds
	remainingAngleError := targetRotation - rotation
	crossedTarget := math.Copysign(1, remainingAngleError) != directionToTarget || remainingAngleError == 0

	// Обнуляем угловую скорость у цели только когда целевой угол сейчас не двигается.
	if isTargetRotationStopped && crossedTarget {
		return targetRotation, 0
	}

	return rotation, angularVelocity
}

// применяет вход пилота к кораблю и возвращает новое состояние после одного физического шага.
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
	targetRotationSpeed := 0.0
	if dtSeconds > 0 {
		targetRotationSpeed = input.TargetRotationDelta / dtSeconds
	}
	rotation, angularVelocity := stepAngularVelocityToTarget(ship, targetRotation, targetRotationSpeed, dtSeconds)

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
