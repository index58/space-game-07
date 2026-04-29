import type { ShipState, WorldVector } from "./types";

export type ShipInput = {
  thrustForward: boolean;
  thrustBackward: boolean;
  thrustLeft: boolean;
  thrustRight: boolean;
  targetRotationDelta: number;
};

export const EPSILON = 0.000001;
const ANGLE_EPSILON = 0.0001;

export const getBodySizeMeters = (model: ShipState["model"]) => ({
  width: model.textureBodyWidth / model.textureScale,
  length: model.textureBodyLength / model.textureScale,
});

export const getMomentOfInertia = (ship: ShipState): number => {
  const body = getBodySizeMeters(ship.model);

  // Используем приближение заполненного эллипса через ширину и длину тела.
  return (ship.model.massKg * (body.width ** 2 + body.length ** 2)) / 16;
};

export const getForwardVector = (rotation: number): WorldVector => ({
  x: Math.sin(rotation),
  y: Math.cos(rotation),
});

export const getRightVector = (rotation: number): WorldVector => ({
  x: Math.cos(rotation),
  y: -Math.sin(rotation),
});

const brakeValue = (value: number, acceleration: number, dtSeconds: number): number => {
  const delta = acceleration * dtSeconds;

  if (Math.abs(value) <= delta) {
    return 0;
  }

  return value - Math.sign(value) * delta;
};

const clampVectorLength = (x: number, y: number, maxLength: number): WorldVector => {
  const length = Math.hypot(x, y);

  // Ограничиваем только превышение, чтобы не менять направление и малые скорости.
  if (length <= maxLength || length <= EPSILON) {
    return { x, y };
  }

  const scale = maxLength / length;

  return { x: x * scale, y: y * scale };
};

const clampAbsoluteValue = (value: number, maxAbsoluteValue: number): number => {
  // Сохраняем знак скорости, но не даем модулю превысить максимум модели.
  return Math.sign(value) * Math.min(Math.abs(value), maxAbsoluteValue);
};

export const hasLinearThrust = (input: ShipInput): boolean =>
  input.thrustForward ||
  input.thrustBackward ||
  input.thrustLeft ||
  input.thrustRight;

const getProjection = (vector: WorldVector, axis: WorldVector): number =>
  vector.x * axis.x + vector.y * axis.y;

const getAngularAcceleration = (ship: ShipState): number =>
  ship.model.torqueNm / getMomentOfInertia(ship);

const stepAngularVelocityToTarget = (
  ship: ShipState,
  targetRotation: number,
  dtSeconds: number,
): { rotation: number; angularVelocity: number } => {
  const angularAcceleration = getAngularAcceleration(ship);
  const angleError = targetRotation - ship.rotation;

  if (Math.abs(angleError) <= ANGLE_EPSILON) {
    const angularVelocity = brakeValue(
      ship.angularVelocity,
      angularAcceleration,
      dtSeconds,
    );

    return {
      rotation: angularVelocity === 0 ? targetRotation : ship.rotation + angularVelocity * dtSeconds,
      angularVelocity,
    };
  }

  const directionToTarget = Math.sign(angleError);
  const currentDirection = Math.sign(ship.angularVelocity);
  const stoppingDistance = ship.angularVelocity ** 2 / (2 * angularAcceleration);
  const shouldBrake =
    currentDirection !== 0 &&
    currentDirection === directionToTarget &&
    stoppingDistance >= Math.abs(angleError);
  const torqueDirection = shouldBrake ? -currentDirection : directionToTarget;
  const rawAngularVelocity = ship.angularVelocity + torqueDirection * angularAcceleration * dtSeconds;
  const isAngularVelocityClamped = Math.abs(rawAngularVelocity) > ship.model.maxAngularSpeedRadPerSecond;
  const angularVelocity = clampAbsoluteValue(rawAngularVelocity, ship.model.maxAngularSpeedRadPerSecond);
  const rotation = ship.rotation + angularVelocity * dtSeconds;
  const remainingAngleError = targetRotation - rotation;
  const crossedTarget = Math.sign(remainingAngleError) !== directionToTarget || remainingAngleError === 0;

  // Если доступного момента хватает дойти до цели за текущий шаг, фиксируем угол без обратного разгона.
  if (!isAngularVelocityClamped && crossedTarget) {
    return {
      rotation: targetRotation,
      angularVelocity: 0,
    };
  }

  return {
    rotation,
    angularVelocity,
  };
};

export const stepShipPhysics = (
  ship: ShipState,
  input: ShipInput,
  dtSeconds: number,
): ShipState => {
  const forward = getForwardVector(ship.rotation);
  const right = getRightVector(ship.rotation);
  const hasAlongControl = input.thrustForward || input.thrustBackward;
  const hasAcrossControl = input.thrustLeft || input.thrustRight;
  const along = (input.thrustForward ? 1 : 0) - (input.thrustBackward ? 1 : 0);
  const across = (input.thrustRight ? 1 : 0) - (input.thrustLeft ? 1 : 0);
  const linearAcceleration = ship.model.thrustN / ship.model.massKg;

  // Скорость раскладывается по локальным осям, чтобы автоторможение одной оси не мешало ручной тяге другой оси.
  const alongVelocity = hasAlongControl
    ? getProjection(ship.velocity, forward) + along * linearAcceleration * dtSeconds
    : brakeValue(getProjection(ship.velocity, forward), linearAcceleration, dtSeconds);
  const acrossVelocity = hasAcrossControl
    ? getProjection(ship.velocity, right) + across * linearAcceleration * dtSeconds
    : brakeValue(getProjection(ship.velocity, right), linearAcceleration, dtSeconds);

  let velocityX = forward.x * alongVelocity + right.x * acrossVelocity;
  let velocityY = forward.y * alongVelocity + right.y * acrossVelocity;

  const limitedVelocity = clampVectorLength(velocityX, velocityY, ship.model.maxSpeedMps);
  velocityX = limitedVelocity.x;
  velocityY = limitedVelocity.y;

  const targetRotation = ship.targetRotation + input.targetRotationDelta;
  const angularStep = stepAngularVelocityToTarget(ship, targetRotation, dtSeconds);

  return {
    ...ship,
    position: {
      x: ship.position.x + velocityX * dtSeconds,
      y: ship.position.y + velocityY * dtSeconds,
    },
    velocity: { x: velocityX, y: velocityY },
    rotation: angularStep.rotation,
    angularVelocity: angularStep.angularVelocity,
    targetRotation,
  };
};
