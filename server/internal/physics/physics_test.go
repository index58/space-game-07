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

// Проверяет, что момент инерции считается по приближению заполненного эллипса.
func TestMomentOfInertiaUsesFilledEllipseApproximation(t *testing.T) {
	closeTo(t, physics.MomentOfInertia(testShip(1, 0, 0), testModel()), 490173.75)
}

// Проверяет, что при нулевом повороте тяга вперёд ускоряет корабль по положительной оси Y.
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

// Проверяет, что масса объекта напрямую участвует в расчёте ускорения как килограммы.
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

// Проверяет, что линейная скорость ограничивается максимальным значением модели.
func TestStepShipClampsLinearVelocityToModelMaximum(t *testing.T) {
	next := physics.StepShip(
		testShip(1, 0, 0),
		testModel(),
		game.ShipInput{ThrustForward: true},
		10,
	)

	closeTo(t, math.Hypot(next.VelocityX, next.VelocityY), 497)
}

// Проверяет, что ввод игрока изменяет целевой угол поворота.
func TestStepShipUpdatesTargetRotationFromInput(t *testing.T) {
	next := physics.StepShip(
		testShip(1, 0, 0),
		testModel(),
		game.ShipInput{TargetRotationDelta: 0.25},
		0.016,
	)

	closeTo(t, next.TargetRotation, 0.25)
}

// Проверяет, что корабль начинает вращаться в сторону заданного целевого угла.
func TestStepShipStartsRotatingTowardTargetRotation(t *testing.T) {
	ship := testShip(1, 0, 0)
	ship.TargetRotation = 1

	next := physics.StepShip(ship, testModel(), idleInput(), 0.05)

	if next.AngularSpeed <= 0 {
		t.Fatalf("got angular velocity %v, want positive", next.AngularSpeed)
	}
}

// Проверяет, что ошибка угла через границу пи не нормализуется в короткий путь.
func TestStepShipDoesNotNormalizeAngleErrorAcrossPiBoundary(t *testing.T) {
	ship := testShip(1, 0, 0)
	ship.Rotation = math.Pi - 0.1
	ship.TargetRotation = -math.Pi + 0.1

	next := physics.StepShip(ship, testModel(), idleInput(), 0.05)

	if next.AngularSpeed >= 0 {
		t.Fatalf("got angular velocity %v, want negative", next.AngularSpeed)
	}
}

// Проверяет, что рядом с целевым углом вращение гасится до остановки.
func TestStepShipBrakesAngularVelocityNearTargetRotation(t *testing.T) {
	ship := testShip(1, 0, 0)
	ship.Rotation = 1
	ship.TargetRotation = 1
	ship.AngularSpeed = 0.1

	next := physics.StepShip(ship, testModel(), idleInput(), 1)

	closeTo(t, next.AngularSpeed, 0)
	closeTo(t, next.Rotation, 1)
}

// Проверяет, что поворот останавливается ровно на цели без перелёта.
func TestStepShipStopsRotationAtTargetWithoutOvershoot(t *testing.T) {
	ship := testShip(1, 0, 0)
	ship.Rotation = 0
	ship.TargetRotation = 0.01
	ship.AngularSpeed = 1

	next := physics.StepShip(ship, testModel(), idleInput(), 0.05)

	closeTo(t, next.Rotation, 0.01)
	closeTo(t, next.AngularSpeed, 0)
}

// Проверяет, что перед финальной остановкой угловая скорость уже уменьшена до достижимого шага торможения.
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

// Проверяет, что движение цели поворота не обнуляет текущую угловую скорость преждевременно.
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

// Проверяет, что угловая скорость ограничивается максимальным значением модели.
func TestStepShipClampsAngularVelocityToModelMaximum(t *testing.T) {
	ship := testShip(1, 0, 0)
	ship.TargetRotation = 100

	next := physics.StepShip(ship, testModel(), idleInput(), 10)

	closeTo(t, next.AngularSpeed, 3)
}

// Проверяет, что при достигнутом целевом угле гасится линейное движение и вращение.
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

// Проверяет, что линейное торможение продолжается во время разворота к цели.
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

// Проверяет, что продольная тяга гасит поперечную скорость и создаёт движение вперёд.
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

// Проверяет, что поперечная тяга гасит продольную скорость и создаёт боковое движение.
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

// Проверяет, что одновременная встречная продольная тяга не включает автоторможение по этой оси.
func TestStepShipDoesNotAutobrakeAlongAxisWhenForwardAndBackwardPressed(t *testing.T) {
	ship := testShip(1, 0, 0)
	ship.VelocityY = 100

	next := physics.StepShip(ship, testModel(), game.ShipInput{ThrustForward: true, ThrustBackward: true}, 0.1)

	closeTo(t, next.VelocityX, 0)
	closeTo(t, next.VelocityY, 100)
}

// Проверяет, что одновременная встречная поперечная тяга не включает автоторможение по этой оси.
func TestStepShipDoesNotAutobrakeAcrossAxisWhenLeftAndRightPressed(t *testing.T) {
	ship := testShip(1, 0, 0)
	ship.VelocityX = 100

	next := physics.StepShip(ship, testModel(), game.ShipInput{ThrustLeft: true, ThrustRight: true}, 0.1)

	closeTo(t, next.VelocityX, 100)
	closeTo(t, next.VelocityY, 0)
}

// Проверяет, что корабль без пилота тормозит с постоянным замедлением и обновляет положение.
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

// Проверяет, что торможение корабля без пилота не зависит от текущей скорости.
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

// Проверяет, что постоянное торможение сохраняет направление вектора скорости.
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

// Проверяет, что постоянное торможение не разворачивает скорость в противоположную сторону.
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

// Проверяет, что корабль без пилота получает угловое торможение.
func TestStepUnpilotedShipAppliesAngularBrake(t *testing.T) {
	ship := testShip(1, 0, 0)
	ship.AngularSpeed = 3

	next := physics.StepUnpilotedShip(ship, 0.5)

	closeTo(t, next.AngularSpeed, 2.5)
}

// Проверяет, что угловое торможение не разворачивает вращение в противоположную сторону.
func TestStepUnpilotedShipDoesNotReverseAngularVelocityDuringBrake(t *testing.T) {
	ship := testShip(1, 0, 0)
	ship.AngularSpeed = -0.25

	next := physics.StepUnpilotedShip(ship, 1)

	closeTo(t, next.AngularSpeed, 0)
}

// Проверяет, что столкновение с закреплённым телом отражает скорость вдоль нормали с упругостью.
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

// Проверяет, что столкновение двух подвижных тел разделяет коррекцию и импульс с учётом масс.
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

// Проверяет, что нецентральный контакт добавляет угловую скорость ударившемуся телу.
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

// Проверяет, что после вращательного отскока целевой угол переносится на угол остановки.
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
