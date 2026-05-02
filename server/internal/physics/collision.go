package physics

import (
	"math"

	"space-game-07-server/internal/data"
	"space-game-07-server/internal/game"
)

// Задает долю скорости, сохраняемую после отскока по нормали удара.
const collisionRestitution = 0.5

// Хранит геометрию найденного столкновения для разделения и импульса.
type Collision struct {
	Correction   game.WorldVector // Минимальный сдвиг первого тела из пересечения.
	ContactPoint game.WorldVector // Точка приложения импульса в координатах мира.
}

// Возвращает минимальный сдвиг первого тела, если два выпуклых тела пересекаются.
func CollisionCorrection(moving data.CosmicObject, movingModel data.CosmicObjectModel, obstacle data.CosmicObject, obstacleModel data.CosmicObjectModel) (game.WorldVector, bool) {
	collision, collided := CollisionInfo(moving, movingModel, obstacle, obstacleModel)
	return collision.Correction, collided
}

// Возвращает сдвиг и точку контакта для двух пересекающихся выпуклых тел.
func CollisionInfo(moving data.CosmicObject, movingModel data.CosmicObjectModel, obstacle data.CosmicObject, obstacleModel data.CosmicObjectModel) (Collision, bool) {
	movingPolygon := transformedBodyPolygon(moving, movingModel)
	obstaclePolygon := transformedBodyPolygon(obstacle, obstacleModel)
	centerDelta := game.WorldVector{
		X: moving.X - obstacle.X,
		Y: moving.Y - obstacle.Y,
	}

	smallestOverlap := math.MaxFloat64
	smallestAxis := game.WorldVector{}
	if separated, overlap, axis := smallestSeparatingAxis(movingPolygon, obstaclePolygon, centerDelta); separated {
		return Collision{}, false
	} else if overlap < smallestOverlap {
		smallestOverlap = overlap
		smallestAxis = axis
	}
	if separated, overlap, axis := smallestSeparatingAxis(obstaclePolygon, movingPolygon, centerDelta); separated {
		return Collision{}, false
	} else if overlap < smallestOverlap {
		smallestOverlap = overlap
		smallestAxis = axis
	}

	correction := game.WorldVector{
		X: smallestAxis.X * smallestOverlap,
		Y: smallestAxis.Y * smallestOverlap,
	}

	return Collision{
		Correction:   correction,
		ContactPoint: collisionContactPoint(movingPolygon, obstaclePolygon, correction),
	}, true
}

// Применяет разделение тел и импульс столкновения с учетом массы и вращения.
func ApplyCollisionResponse(moving data.CosmicObject, movingModel data.CosmicObjectModel, obstacle data.CosmicObject, obstacleModel data.CosmicObjectModel, collision Collision) (data.CosmicObject, data.CosmicObject) {
	correction := collision.Correction
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

	movingRadius := game.WorldVector{
		X: collision.ContactPoint.X - moving.X,
		Y: collision.ContactPoint.Y - moving.Y,
	}
	obstacleRadius := game.WorldVector{
		X: collision.ContactPoint.X - obstacle.X,
		Y: collision.ContactPoint.Y - obstacle.Y,
	}
	movingContactVelocity := contactVelocity(moving, movingRadius)
	obstacleContactVelocity := contactVelocity(obstacle, obstacleRadius)
	relativeVelocityX := movingContactVelocity.X - obstacleContactVelocity.X
	relativeVelocityY := movingContactVelocity.Y - obstacleContactVelocity.Y
	velocityAlongNormal := relativeVelocityX*normal.X + relativeVelocityY*normal.Y
	if velocityAlongNormal >= 0 {
		return moving, obstacle
	}

	movingInverseInertia := collisionInverseInertia(moving, movingModel)
	obstacleInverseInertia := collisionInverseInertia(obstacle, obstacleModel)
	movingAngularNormal := clockwiseCross(movingRadius, normal)
	obstacleAngularNormal := clockwiseCross(obstacleRadius, normal)
	impulseDenominator := totalInverseMass +
		movingAngularNormal*movingAngularNormal*movingInverseInertia +
		obstacleAngularNormal*obstacleAngularNormal*obstacleInverseInertia
	if impulseDenominator <= 0 {
		return moving, obstacle
	}

	normalImpulse := -(1 + collisionRestitution) * velocityAlongNormal / impulseDenominator
	moving.VelocityX += normal.X * normalImpulse * movingInverseMass
	moving.VelocityY += normal.Y * normalImpulse * movingInverseMass
	obstacle.VelocityX -= normal.X * normalImpulse * obstacleInverseMass
	obstacle.VelocityY -= normal.Y * normalImpulse * obstacleInverseMass
	moving.AngularSpeed = clampAbsoluteValue(moving.AngularSpeed+movingAngularNormal*normalImpulse*movingInverseInertia, moving.MaxAngularSpeed)
	obstacle.AngularSpeed = clampAbsoluteValue(obstacle.AngularSpeed-obstacleAngularNormal*normalImpulse*obstacleInverseInertia, obstacle.MaxAngularSpeed)
	moving.Speed = math.Hypot(moving.VelocityX, moving.VelocityY)
	obstacle.Speed = math.Hypot(obstacle.VelocityX, obstacle.VelocityY)

	return moving, obstacle
}

// Возвращает подвижность тела для импульсного расчета.
func collisionInverseMass(cosmicObject data.CosmicObject) float64 {
	if cosmicObject.Anchored || !cosmicObject.Enabled || cosmicObject.Mass <= 0 {
		return 0
	}

	return 1 / cosmicObject.Mass
}

// Возвращает обратный момент инерции для импульсного расчета.
func collisionInverseInertia(cosmicObject data.CosmicObject, model data.CosmicObjectModel) float64 {
	if cosmicObject.Anchored || !cosmicObject.Enabled || cosmicObject.Mass <= 0 {
		return 0
	}

	moment := MomentOfInertia(cosmicObject, model)
	if moment <= 0 {
		return 0
	}

	return 1 / moment
}

// Считает скорость точки тела с учетом поступательного и вращательного движения.
func contactVelocity(cosmicObject data.CosmicObject, radius game.WorldVector) game.WorldVector {
	return game.WorldVector{
		X: cosmicObject.VelocityX + cosmicObject.AngularSpeed*radius.Y,
		Y: cosmicObject.VelocityY - cosmicObject.AngularSpeed*radius.X,
	}
}

// Возвращает двумерное векторное произведение в знаке часовой стрелки.
func clockwiseCross(first game.WorldVector, second game.WorldVector) float64 {
	return first.Y*second.X - first.X*second.Y
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

// Оценивает точку контакта как центр пересечения двух выпуклых многоугольников.
func collisionContactPoint(first []game.WorldVector, second []game.WorldVector, correction game.WorldVector) game.WorldVector {
	intersection := convexIntersection(first, second)
	if len(intersection) > 0 {
		return polygonCentroid(intersection)
	}

	return supportContactPoint(first, second, correction)
}

// Обрезает один выпуклый многоугольник границами другого.
func convexIntersection(subject []game.WorldVector, clip []game.WorldVector) []game.WorldVector {
	result := append([]game.WorldVector{}, subject...)
	orientation := polygonSignedArea(clip)

	for index := range clip {
		if len(result) == 0 {
			return result
		}

		nextIndex := (index + 1) % len(clip)
		edgeStart := clip[index]
		edgeEnd := clip[nextIndex]
		input := result
		result = []game.WorldVector{}
		previous := input[len(input)-1]
		previousInside := insideClipEdge(previous, edgeStart, edgeEnd, orientation)

		for _, current := range input {
			currentInside := insideClipEdge(current, edgeStart, edgeEnd, orientation)
			if currentInside {
				if !previousInside {
					result = append(result, lineIntersection(previous, current, edgeStart, edgeEnd))
				}
				result = append(result, current)
			} else if previousInside {
				result = append(result, lineIntersection(previous, current, edgeStart, edgeEnd))
			}
			previous = current
			previousInside = currentInside
		}
	}

	return result
}

// Проверяет, лежит ли точка с внутренней стороны ребра.
func insideClipEdge(point game.WorldVector, edgeStart game.WorldVector, edgeEnd game.WorldVector, orientation float64) bool {
	cross := (edgeEnd.X-edgeStart.X)*(point.Y-edgeStart.Y) - (edgeEnd.Y-edgeStart.Y)*(point.X-edgeStart.X)
	if orientation >= 0 {
		return cross >= -Epsilon
	}
	return cross <= Epsilon
}

// Возвращает пересечение отрезка с прямой ребра отсечения.
func lineIntersection(segmentStart game.WorldVector, segmentEnd game.WorldVector, edgeStart game.WorldVector, edgeEnd game.WorldVector) game.WorldVector {
	segment := game.WorldVector{
		X: segmentEnd.X - segmentStart.X,
		Y: segmentEnd.Y - segmentStart.Y,
	}
	edge := game.WorldVector{
		X: edgeEnd.X - edgeStart.X,
		Y: edgeEnd.Y - edgeStart.Y,
	}
	denominator := segment.X*edge.Y - segment.Y*edge.X
	if math.Abs(denominator) <= Epsilon {
		return segmentEnd
	}

	offset := game.WorldVector{
		X: edgeStart.X - segmentStart.X,
		Y: edgeStart.Y - segmentStart.Y,
	}
	t := (offset.X*edge.Y - offset.Y*edge.X) / denominator
	return game.WorldVector{
		X: segmentStart.X + segment.X*t,
		Y: segmentStart.Y + segment.Y*t,
	}
}

// Считает ориентированную площадь многоугольника.
func polygonSignedArea(points []game.WorldVector) float64 {
	area := 0.0
	for index := range points {
		nextIndex := (index + 1) % len(points)
		area += points[index].X*points[nextIndex].Y - points[nextIndex].X*points[index].Y
	}
	return area / 2
}

// Считает центр площади многоугольника или среднюю точку вырожденного набора.
func polygonCentroid(points []game.WorldVector) game.WorldVector {
	areaFactor := 0.0
	centroid := game.WorldVector{}
	for index := range points {
		nextIndex := (index + 1) % len(points)
		cross := points[index].X*points[nextIndex].Y - points[nextIndex].X*points[index].Y
		areaFactor += cross
		centroid.X += (points[index].X + points[nextIndex].X) * cross
		centroid.Y += (points[index].Y + points[nextIndex].Y) * cross
	}
	if math.Abs(areaFactor) <= Epsilon {
		return averagePoint(points)
	}

	return game.WorldVector{
		X: centroid.X / (3 * areaFactor),
		Y: centroid.Y / (3 * areaFactor),
	}
}

// Возвращает среднее положение набора точек.
func averagePoint(points []game.WorldVector) game.WorldVector {
	result := game.WorldVector{}
	for _, point := range points {
		result.X += point.X
		result.Y += point.Y
	}
	if len(points) == 0 {
		return result
	}

	return game.WorldVector{
		X: result.X / float64(len(points)),
		Y: result.Y / float64(len(points)),
	}
}

// Оценивает контакт через опорные точки, если область пересечения выродилась.
func supportContactPoint(first []game.WorldVector, second []game.WorldVector, correction game.WorldVector) game.WorldVector {
	length := math.Hypot(correction.X, correction.Y)
	if length <= Epsilon {
		allPoints := append(append([]game.WorldVector{}, first...), second...)
		return averagePoint(allPoints)
	}

	normal := game.WorldVector{
		X: correction.X / length,
		Y: correction.Y / length,
	}
	firstPoint := supportAverage(first, normal, false)
	secondPoint := supportAverage(second, normal, true)
	return game.WorldVector{
		X: (firstPoint.X + secondPoint.X) / 2,
		Y: (firstPoint.Y + secondPoint.Y) / 2,
	}
}

// Возвращает среднюю опорную точку по выбранной стороне оси.
func supportAverage(points []game.WorldVector, axis game.WorldVector, maximum bool) game.WorldVector {
	target := projection(points[0], axis)
	for _, point := range points[1:] {
		value := projection(point, axis)
		if maximum {
			target = math.Max(target, value)
		} else {
			target = math.Min(target, value)
		}
	}

	selected := make([]game.WorldVector, 0)
	for _, point := range points {
		if math.Abs(projection(point, axis)-target) <= Epsilon {
			selected = append(selected, point)
		}
	}

	return averagePoint(selected)
}
