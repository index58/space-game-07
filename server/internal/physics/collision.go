package physics

import (
	"math"

	"space-game-07-server/internal/data"
	"space-game-07-server/internal/game"
)

// Задаёт долю скорости, сохраняемую после отскока по нормали удара.
const collisionRestitution = 0.5

// Возвращает минимальный сдвиг первого тела, если два выпуклых тела пересекаются.
func CollisionCorrection(moving data.CosmicObject, movingModel data.CosmicObjectModel, obstacle data.CosmicObject, obstacleModel data.CosmicObjectModel) (game.WorldVector, bool) {
	movingPolygon := transformedBodyPolygon(moving, movingModel)
	obstaclePolygon := transformedBodyPolygon(obstacle, obstacleModel)
	centerDelta := game.WorldVector{
		X: moving.X - obstacle.X,
		Y: moving.Y - obstacle.Y,
	}

	smallestOverlap := math.MaxFloat64
	smallestAxis := game.WorldVector{}
	if separated, overlap, axis := smallestSeparatingAxis(movingPolygon, obstaclePolygon, centerDelta); separated {
		return game.WorldVector{}, false
	} else if overlap < smallestOverlap {
		smallestOverlap = overlap
		smallestAxis = axis
	}
	if separated, overlap, axis := smallestSeparatingAxis(obstaclePolygon, movingPolygon, centerDelta); separated {
		return game.WorldVector{}, false
	} else if overlap < smallestOverlap {
		smallestOverlap = overlap
		smallestAxis = axis
	}

	return game.WorldVector{
		X: smallestAxis.X * smallestOverlap,
		Y: smallestAxis.Y * smallestOverlap,
	}, true
}

// Применяет разделение тел и импульс столкновения с учётом масс.
func ApplyCollisionResponse(moving data.CosmicObject, obstacle data.CosmicObject, correction game.WorldVector) (data.CosmicObject, data.CosmicObject) {
	length := math.Hypot(correction.X, correction.Y)
	if length <= Epsilon {
		return moving, obstacle
	}

	normal := game.WorldVector{
		X: correction.X / length,
		Y: correction.Y / length,
	}
	movingInverseMass := collisionInverseMass(moving)
	obstacleInverseMass := collisionInverseMass(obstacle)
	totalInverseMass := movingInverseMass + obstacleInverseMass
	if totalInverseMass <= 0 {
		return moving, obstacle
	}

	moving.X += correction.X * movingInverseMass / totalInverseMass
	moving.Y += correction.Y * movingInverseMass / totalInverseMass
	obstacle.X -= correction.X * obstacleInverseMass / totalInverseMass
	obstacle.Y -= correction.Y * obstacleInverseMass / totalInverseMass

	relativeVelocityX := moving.VelocityX - obstacle.VelocityX
	relativeVelocityY := moving.VelocityY - obstacle.VelocityY
	velocityAlongNormal := relativeVelocityX*normal.X + relativeVelocityY*normal.Y
	if velocityAlongNormal >= 0 {
		return moving, obstacle
	}

	normalImpulse := -(1 + collisionRestitution) * velocityAlongNormal / totalInverseMass
	moving.VelocityX += normal.X * normalImpulse * movingInverseMass
	moving.VelocityY += normal.Y * normalImpulse * movingInverseMass
	obstacle.VelocityX -= normal.X * normalImpulse * obstacleInverseMass
	obstacle.VelocityY -= normal.Y * normalImpulse * obstacleInverseMass
	moving.Speed = math.Hypot(moving.VelocityX, moving.VelocityY)
	obstacle.Speed = math.Hypot(obstacle.VelocityX, obstacle.VelocityY)

	return moving, obstacle
}

// Возвращает подвижность тела для импульсного расчёта.
func collisionInverseMass(cosmicObject data.CosmicObject) float64 {
	if cosmicObject.Anchored || !cosmicObject.Enabled || cosmicObject.Mass <= 0 {
		return 0
	}

	return 1 / cosmicObject.Mass
}

// Переводит локальные вершины модели в координаты игрового мира.
func transformedBodyPolygon(cosmicObject data.CosmicObject, model data.CosmicObjectModel) []game.WorldVector {
	forward := ForwardVector(cosmicObject.Rotation)
	right := RightVector(cosmicObject.Rotation)
	points := make([]game.WorldVector, len(model.BodyPolygon))

	for index, point := range model.BodyPolygon {
		points[index] = game.WorldVector{
			X: cosmicObject.X + right.X*point.X + forward.X*point.Y,
			Y: cosmicObject.Y + right.Y*point.X + forward.Y*point.Y,
		}
	}

	return points
}

// Проверяет оси одного тела и возвращает наименьшее перекрытие.
func smallestSeparatingAxis(first []game.WorldVector, second []game.WorldVector, centerDelta game.WorldVector) (bool, float64, game.WorldVector) {
	smallestOverlap := math.MaxFloat64
	smallestAxis := game.WorldVector{}

	for index := range first {
		nextIndex := (index + 1) % len(first)
		edge := game.WorldVector{
			X: first[nextIndex].X - first[index].X,
			Y: first[nextIndex].Y - first[index].Y,
		}
		axisLength := math.Hypot(edge.X, edge.Y)
		if axisLength <= Epsilon {
			continue
		}

		axis := game.WorldVector{
			X: -edge.Y / axisLength,
			Y: edge.X / axisLength,
		}
		if axis.X*centerDelta.X+axis.Y*centerDelta.Y < 0 {
			axis.X = -axis.X
			axis.Y = -axis.Y
		}

		firstMin, firstMax := projectPolygon(first, axis)
		secondMin, secondMax := projectPolygon(second, axis)
		overlap := math.Min(firstMax, secondMax) - math.Max(firstMin, secondMin)
		if overlap <= Epsilon {
			return true, 0, game.WorldVector{}
		}
		if overlap < smallestOverlap {
			smallestOverlap = overlap
			smallestAxis = axis
		}
	}

	return false, smallestOverlap, smallestAxis
}

// Проецирует вершины на заданную ось.
func projectPolygon(points []game.WorldVector, axis game.WorldVector) (float64, float64) {
	minimum := points[0].X*axis.X + points[0].Y*axis.Y
	maximum := minimum

	for _, point := range points[1:] {
		value := point.X*axis.X + point.Y*axis.Y
		minimum = math.Min(minimum, value)
		maximum = math.Max(maximum, value)
	}

	return minimum, maximum
}
