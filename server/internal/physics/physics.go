package physics

import (
	"math"

	"space-game-07-server/internal/data"
	"space-game-07-server/internal/game"
)

const (
	// Защищает вычисления от дрожания около нуля при работе с float64.
	Epsilon = 0.000001
	// Задает допуск, при котором корабль считается почти повернутым к цели.
	angleEpsilon = 0.0001
)

// Переводит пиксельный размер физического тела модели в метры мира.
func BodySizeMeters(model data.CosmicObjectModel) game.WorldVector {
	return game.WorldVector{
		X: float64(model.TextureBodyWidth) / model.TextureScale,
		Y: float64(model.TextureBodyLength) / model.TextureScale,
	}
}

// Оценивает момент инерции корабля как прямоугольного тела.
func MomentOfInertia(cosmicObject data.CosmicObject, model data.CosmicObjectModel) float64 {
	body := BodySizeMeters(model)
	return cosmicObject.Mass * (body.X*body.X + body.Y*body.Y) / 16
}

// Возвращает локальную ось "вперед" для текущего угла корабля.
func ForwardVector(rotation float64) game.WorldVector {
	return game.WorldVector{
		X: math.Sin(rotation),
		Y: math.Cos(rotation),
	}
}

// Возвращает локальную ось "вправо" для текущего угла корабля.
func RightVector(rotation float64) game.WorldVector {
	return game.WorldVector{
		X: math.Cos(rotation),
		Y: -math.Sin(rotation),
	}
}

// Уменьшает значение к нулю с заданным ускорением без смены знака.
func brakeValue(value float64, acceleration float64, dtSeconds float64) float64 {
	delta := acceleration * dtSeconds

	if math.Abs(value) <= delta {
		return 0
	}

	return value - math.Copysign(delta, value)
}

// Ограничивает длину вектора максимальной скоростью.
func clampVectorLength(x float64, y float64, maxLength float64) game.WorldVector {
	length := math.Hypot(x, y)

	if length <= maxLength || length <= Epsilon {
		return game.WorldVector{X: x, Y: y}
	}

	scale := maxLength / length

	return game.WorldVector{X: x * scale, Y: y * scale}
}

// Ограничивает модуль скалярного значения и зануляет микродрожание.
func clampAbsoluteValue(value float64, maxAbsoluteValue float64) float64 {
	if math.Abs(value) <= Epsilon {
		return 0
	}

	return math.Copysign(math.Min(math.Abs(value), maxAbsoluteValue), value)
}

// Двигает значение к цели не дальше разрешенной дельты за шаг.
func moveToward(value float64, target float64, maxDelta float64) float64 {
	delta := target - value

	if math.Abs(delta) <= maxDelta {
		return target
	}

	return value + math.Copysign(maxDelta, delta)
}

// Возвращает скалярную проекцию вектора на заданную ось.
func projection(vector game.WorldVector, axis game.WorldVector) float64 {
	return vector.X*axis.X + vector.Y*axis.Y
}

// Считает доступное угловое ускорение от крутящего момента объекта.
func angularAcceleration(cosmicObject data.CosmicObject, model data.CosmicObjectModel) float64 {
	return cosmicObject.MaxTorque / MomentOfInertia(cosmicObject, model)
}

// Ведет угол к целевому так, чтобы успеть затормозить без перелета.
func stepAngularVelocityToTarget(cosmicObject data.CosmicObject, model data.CosmicObjectModel, targetRotation float64, targetRotationSpeed float64, dtSeconds float64) (float64, float64) {
	acceleration := angularAcceleration(cosmicObject, model)
	angleError := targetRotation - cosmicObject.Rotation
	isTargetRotationStopped := math.Abs(targetRotationSpeed) <= Epsilon

	if math.Abs(angleError) <= angleEpsilon {
		angularVelocity := brakeValue(cosmicObject.AngularSpeed, acceleration, dtSeconds)

		if angularVelocity == 0 {
			return targetRotation, angularVelocity
		}

		return cosmicObject.Rotation + angularVelocity*dtSeconds, angularVelocity
	}

	directionToTarget := math.Copysign(1, angleError)
	distanceToTarget := math.Abs(angleError)
	maxVelocityDelta := acceleration * dtSeconds
	velocityTowardTarget := cosmicObject.AngularSpeed * directionToTarget

	safeSpeed := math.Max(
		0,
		math.Sqrt(maxVelocityDelta*maxVelocityDelta+2*acceleration*distanceToTarget)-maxVelocityDelta,
	)
	desiredSpeed := math.Min(safeSpeed, cosmicObject.MaxAngularSpeed)
	if velocityTowardTarget < 0 {
		desiredSpeed = 0
	}
	desiredAngularVelocity := directionToTarget * desiredSpeed
	rawAngularVelocity := moveToward(cosmicObject.AngularSpeed, desiredAngularVelocity, maxVelocityDelta)
	angularVelocity := clampAbsoluteValue(rawAngularVelocity, cosmicObject.MaxAngularSpeed)
	rotation := cosmicObject.Rotation + angularVelocity*dtSeconds
	remainingAngleError := targetRotation - rotation
	crossedTarget := math.Copysign(1, remainingAngleError) != directionToTarget || remainingAngleError == 0

	// Обнуляем угловую скорость у цели только когда целевой угол сейчас не двигается.
	if isTargetRotationStopped && crossedTarget {
		return targetRotation, 0
	}

	return rotation, angularVelocity
}

// Применяет ввод пилота к объекту и возвращает новое состояние после одного физического шага.
func StepShip(cosmicObject data.CosmicObject, model data.CosmicObjectModel, input game.ShipInput, dtSeconds float64) data.CosmicObject {
	forward := ForwardVector(cosmicObject.Rotation)
	right := RightVector(cosmicObject.Rotation)
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

	linearAcceleration := cosmicObject.MaxAlongForce / cosmicObject.Mass
	currentVelocity := game.WorldVector{X: cosmicObject.VelocityX, Y: cosmicObject.VelocityY}
	alongVelocity := 0.0
	if hasAlongControl {
		alongVelocity = projection(currentVelocity, forward) + along*linearAcceleration*dtSeconds
	} else {
		alongVelocity = brakeValue(projection(currentVelocity, forward), linearAcceleration, dtSeconds)
	}

	acrossVelocity := 0.0
	if hasAcrossControl {
		acrossVelocity = projection(currentVelocity, right) + across*linearAcceleration*dtSeconds
	} else {
		acrossVelocity = brakeValue(projection(currentVelocity, right), linearAcceleration, dtSeconds)
	}

	velocityX := forward.X*alongVelocity + right.X*acrossVelocity
	velocityY := forward.Y*alongVelocity + right.Y*acrossVelocity
	limitedVelocity := clampVectorLength(velocityX, velocityY, cosmicObject.MaxSpeed)
	targetRotation := cosmicObject.TargetRotation + input.TargetRotationDelta
	targetRotationSpeed := 0.0
	if dtSeconds > 0 {
		targetRotationSpeed = input.TargetRotationDelta / dtSeconds
	}
	rotation, angularVelocity := stepAngularVelocityToTarget(cosmicObject, model, targetRotation, targetRotationSpeed, dtSeconds)

	cosmicObject.X += limitedVelocity.X * dtSeconds
	cosmicObject.Y += limitedVelocity.Y * dtSeconds
	cosmicObject.VelocityX = limitedVelocity.X
	cosmicObject.VelocityY = limitedVelocity.Y
	cosmicObject.Speed = math.Hypot(limitedVelocity.X, limitedVelocity.Y)
	cosmicObject.Rotation = rotation
	cosmicObject.AngularSpeed = angularVelocity
	cosmicObject.TargetRotation = targetRotation
	cosmicObject.AlongForce = axisForce(input.ThrustForward, input.ThrustBackward, cosmicObject.MaxAlongForce)
	cosmicObject.AcrossForce = axisForce(input.ThrustRight, input.ThrustLeft, cosmicObject.MaxAcrossForce)
	if input.TargetRotationDelta == 0 {
		cosmicObject.Torque = 0
	} else {
		cosmicObject.Torque = cosmicObject.MaxTorque
	}

	return cosmicObject
}

// Двигает свободное тело без управляющей тяги и автопилотного торможения.
func StepFreeBody(cosmicObject data.CosmicObject, dtSeconds float64) data.CosmicObject {
	cosmicObject.X += cosmicObject.VelocityX * dtSeconds
	cosmicObject.Y += cosmicObject.VelocityY * dtSeconds
	cosmicObject.Rotation += cosmicObject.AngularSpeed * dtSeconds
	cosmicObject.Speed = math.Hypot(cosmicObject.VelocityX, cosmicObject.VelocityY)
	cosmicObject.AlongForce = 0
	cosmicObject.AcrossForce = 0
	cosmicObject.Torque = 0
	return cosmicObject
}

// Переводит пару противоположных кнопок в силу по одной локальной оси.
func axisForce(positive bool, negative bool, maxForce float64) float64 {
	if positive == negative {
		return 0
	}
	if positive {
		return maxForce
	}
	return -maxForce
}
