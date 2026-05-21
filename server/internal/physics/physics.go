package physics

import (
	"math"

	"space-game-07-server/internal/data"
	"space-game-07-server/internal/game"
)

const (
	// Р—Р°С‰РёС‰Р°РµС‚ РІС‹С‡РёСЃР»РµРЅРёСЏ РѕС‚ РґСЂРѕР¶Р°РЅРёСЏ РѕРєРѕР»Рѕ РЅСѓР»СЏ РїСЂРё СЂР°Р±РѕС‚Рµ СЃ float64.
	Epsilon = 0.000001
	// Р—Р°РґР°РµС‚ РґРѕРїСѓСЃРє, РїСЂРё РєРѕС‚РѕСЂРѕРј РєРѕСЂР°Р±Р»СЊ СЃС‡РёС‚Р°РµС‚СЃСЏ РїРѕС‡С‚Рё РїРѕРІРµСЂРЅСѓС‚С‹Рј Рє С†РµР»Рё.
	angleEpsilon = 0.0001
	// Р—Р°РґР°РµС‚ С‚РѕСЂРјРѕР¶РµРЅРёРµ РІСЂР°С‰РµРЅРёСЏ РґР»СЏ РєРѕСЂР°Р±Р»СЏ Р±РµР· РїРѕРґРєР»СЋС‡РµРЅРЅРѕРіРѕ РїРёР»РѕС‚Р°.
	unpilotedShipAngularBrake = 1.0
)

// РџРµСЂРµРІРѕРґРёС‚ РїРёРєСЃРµР»СЊРЅС‹Р№ СЂР°Р·РјРµСЂ С„РёР·РёС‡РµСЃРєРѕРіРѕ С‚РµР»Р° РјРѕРґРµР»Рё РІ РјРµС‚СЂС‹ РјРёСЂР°.
func BodySizeMeters(model data.CosmicObjectModel) game.WorldVector {
	return game.WorldVector{
		X: float64(model.TextureBodyWidth) / model.TextureScale,
		Y: float64(model.TextureBodyLength) / model.TextureScale,
	}
}

// РћС†РµРЅРёРІР°РµС‚ РјРѕРјРµРЅС‚ РёРЅРµСЂС†РёРё РєРѕСЂР°Р±Р»СЏ РєР°Рє РїСЂСЏРјРѕСѓРіРѕР»СЊРЅРѕРіРѕ С‚РµР»Р°.
func MomentOfInertia(cosmicObject data.CosmicObject, model data.CosmicObjectModel) float64 {
	body := BodySizeMeters(model)
	return cosmicObject.Mass * (body.X*body.X + body.Y*body.Y) / 16
}

// Р’РѕР·РІСЂР°С‰Р°РµС‚ Р»РѕРєР°Р»СЊРЅСѓСЋ РѕСЃСЊ "РІРїРµСЂРµРґ" РґР»СЏ С‚РµРєСѓС‰РµРіРѕ СѓРіР»Р° РєРѕСЂР°Р±Р»СЏ.
func ForwardVector(rotation float64) game.WorldVector {
	return game.WorldVector{
		X: math.Sin(rotation),
		Y: math.Cos(rotation),
	}
}

// Р’РѕР·РІСЂР°С‰Р°РµС‚ Р»РѕРєР°Р»СЊРЅСѓСЋ РѕСЃСЊ "РІРїСЂР°РІРѕ" РґР»СЏ С‚РµРєСѓС‰РµРіРѕ СѓРіР»Р° РєРѕСЂР°Р±Р»СЏ.
func RightVector(rotation float64) game.WorldVector {
	return game.WorldVector{
		X: math.Cos(rotation),
		Y: -math.Sin(rotation),
	}
}

// РЈРјРµРЅСЊС€Р°РµС‚ Р·РЅР°С‡РµРЅРёРµ Рє РЅСѓР»СЋ СЃ Р·Р°РґР°РЅРЅС‹Рј СѓСЃРєРѕСЂРµРЅРёРµРј Р±РµР· СЃРјРµРЅС‹ Р·РЅР°РєР°.
func brakeValue(value float64, acceleration float64, dtSeconds float64) float64 {
	delta := acceleration * dtSeconds

	if math.Abs(value) <= delta {
		return 0
	}

	return value - math.Copysign(delta, value)
}

// РћРіСЂР°РЅРёС‡РёРІР°РµС‚ РґР»РёРЅСѓ РІРµРєС‚РѕСЂР° РјР°РєСЃРёРјР°Р»СЊРЅРѕР№ СЃРєРѕСЂРѕСЃС‚СЊСЋ.
func clampVectorLength(x float64, y float64, maxLength float64) game.WorldVector {
	length := math.Hypot(x, y)

	if length <= maxLength || length <= Epsilon {
		return game.WorldVector{X: x, Y: y}
	}

	scale := maxLength / length

	return game.WorldVector{X: x * scale, Y: y * scale}
}

// РћРіСЂР°РЅРёС‡РёРІР°РµС‚ РјРѕРґСѓР»СЊ СЃРєР°Р»СЏСЂРЅРѕРіРѕ Р·РЅР°С‡РµРЅРёСЏ Рё Р·Р°РЅСѓР»СЏРµС‚ РјРёРєСЂРѕРґСЂРѕР¶Р°РЅРёРµ.
func clampAbsoluteValue(value float64, maxAbsoluteValue float64) float64 {
	if math.Abs(value) <= Epsilon {
		return 0
	}

	return math.Copysign(math.Min(math.Abs(value), maxAbsoluteValue), value)
}

// РЈРјРµРЅСЊС€Р°РµС‚ Р»РёРЅРµР№РЅСѓСЋ СЃРєРѕСЂРѕСЃС‚СЊ РїРѕСЃС‚РѕСЏРЅРЅС‹Рј СѓСЃРєРѕСЂРµРЅРёРµРј Р±РµР· СЂР°Р·РІРѕСЂРѕС‚Р° РЅР°РїСЂР°РІР»РµРЅРёСЏ.
func applyConstantBrake(cosmicObject data.CosmicObject, dtSeconds float64) data.CosmicObject {
	speed := math.Hypot(cosmicObject.VelocityX, cosmicObject.VelocityY)
	if speed <= Epsilon {
		cosmicObject.VelocityX = 0
		cosmicObject.VelocityY = 0
		cosmicObject.Speed = 0
		return cosmicObject
	}
	if dtSeconds <= 0 {
		cosmicObject.Speed = speed
		return cosmicObject
	}
	if cosmicObject.Mass <= 0 || cosmicObject.MaxAlongForce <= 0 {
		cosmicObject.Speed = speed
		return cosmicObject
	}

	speedDelta := cosmicObject.MaxAlongForce / cosmicObject.Mass * dtSeconds
	if speedDelta >= speed {
		cosmicObject.VelocityX = 0
		cosmicObject.VelocityY = 0
		cosmicObject.Speed = 0
		return cosmicObject
	}

	nextSpeed := speed - speedDelta
	scale := nextSpeed / speed
	cosmicObject.VelocityX *= scale
	cosmicObject.VelocityY *= scale
	cosmicObject.Speed = nextSpeed
	return cosmicObject
}

// Р”РІРёРіР°РµС‚ Р·РЅР°С‡РµРЅРёРµ Рє С†РµР»Рё РЅРµ РґР°Р»СЊС€Рµ СЂР°Р·СЂРµС€РµРЅРЅРѕР№ РґРµР»СЊС‚С‹ Р·Р° С€Р°Рі.
func moveToward(value float64, target float64, maxDelta float64) float64 {
	delta := target - value

	if math.Abs(delta) <= maxDelta {
		return target
	}

	return value + math.Copysign(maxDelta, delta)
}

// Р’РѕР·РІСЂР°С‰Р°РµС‚ СЃРєР°Р»СЏСЂРЅСѓСЋ РїСЂРѕРµРєС†РёСЋ РІРµРєС‚РѕСЂР° РЅР° Р·Р°РґР°РЅРЅСѓСЋ РѕСЃСЊ.
func projection(vector game.WorldVector, axis game.WorldVector) float64 {
	return vector.X*axis.X + vector.Y*axis.Y
}

// РЎС‡РёС‚Р°РµС‚ РґРѕСЃС‚СѓРїРЅРѕРµ СѓРіР»РѕРІРѕРµ СѓСЃРєРѕСЂРµРЅРёРµ РѕС‚ РєСЂСѓС‚СЏС‰РµРіРѕ РјРѕРјРµРЅС‚Р° РѕР±СЉРµРєС‚Р°.
func angularAcceleration(cosmicObject data.CosmicObject, model data.CosmicObjectModel) float64 {
	return cosmicObject.MaxTorque / MomentOfInertia(cosmicObject, model)
}

// Р’РµРґРµС‚ СѓРіРѕР» Рє С†РµР»РµРІРѕРјСѓ С‚Р°Рє, С‡С‚РѕР±С‹ СѓСЃРїРµС‚СЊ Р·Р°С‚РѕСЂРјРѕР·РёС‚СЊ Р±РµР· РїРµСЂРµР»РµС‚Р°.
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

	// РћР±РЅСѓР»СЏРµРј СѓРіР»РѕРІСѓСЋ СЃРєРѕСЂРѕСЃС‚СЊ Сѓ С†РµР»Рё С‚РѕР»СЊРєРѕ РєРѕРіРґР° С†РµР»РµРІРѕР№ СѓРіРѕР» СЃРµР№С‡Р°СЃ РЅРµ РґРІРёРіР°РµС‚СЃСЏ.
	if isTargetRotationStopped && crossedTarget {
		return targetRotation, 0
	}

	return rotation, angularVelocity
}

// РџСЂРёРјРµРЅСЏРµС‚ РІРІРѕРґ РїРёР»РѕС‚Р° Рє РѕР±СЉРµРєС‚Сѓ Рё РІРѕР·РІСЂР°С‰Р°РµС‚ РЅРѕРІРѕРµ СЃРѕСЃС‚РѕСЏРЅРёРµ РїРѕСЃР»Рµ РѕРґРЅРѕРіРѕ С„РёР·РёС‡РµСЃРєРѕРіРѕ С€Р°РіР°.
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

// Р”РІРёРіР°РµС‚ СЃРІРѕР±РѕРґРЅРѕРµ С‚РµР»Рѕ Р±РµР· СѓРїСЂР°РІР»СЏСЋС‰РµР№ С‚СЏРіРё Рё Р°РІС‚РѕРїРёР»РѕС‚РЅРѕРіРѕ С‚РѕСЂРјРѕР¶РµРЅРёСЏ.
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

// Р”РІРёРіР°РµС‚ РєРѕСЂР°Р±Р»СЊ Р±РµР· РїРѕРґРєР»СЋС‡РµРЅРЅРѕРіРѕ РїРёР»РѕС‚Р° СЃ РґРѕРїРѕР»РЅРёС‚РµР»СЊРЅС‹Рј СЃРѕРїСЂРѕС‚РёРІР»РµРЅРёРµРј.
func StepUnpilotedShip(cosmicObject data.CosmicObject, dtSeconds float64) data.CosmicObject {
	cosmicObject = applyConstantBrake(cosmicObject, dtSeconds)
	cosmicObject.AngularSpeed = brakeValue(cosmicObject.AngularSpeed, unpilotedShipAngularBrake, dtSeconds)
	return StepFreeBody(cosmicObject, dtSeconds)
}

// РџРµСЂРµРІРѕРґРёС‚ РїР°СЂСѓ РїСЂРѕС‚РёРІРѕРїРѕР»РѕР¶РЅС‹С… РєРЅРѕРїРѕРє РІ СЃРёР»Сѓ РїРѕ РѕРґРЅРѕР№ Р»РѕРєР°Р»СЊРЅРѕР№ РѕСЃРё.
func axisForce(positive bool, negative bool, maxForce float64) float64 {
	if positive == negative {
		return 0
	}
	if positive {
		return maxForce
	}
	return -maxForce
}
