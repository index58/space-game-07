import type { ShipState, WorldVector } from "./types";

export type ShipInput = {
  thrustForward: boolean;
  thrustBackward: boolean;
  thrustLeft: boolean;
  thrustRight: boolean;
  turnClockwise: boolean;
  turnCounterClockwise: boolean;
};

export const EPSILON = 0.000001;

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

export const hasAnyThrust = (input: ShipInput): boolean =>
  input.thrustForward ||
  input.thrustBackward ||
  input.thrustLeft ||
  input.thrustRight ||
  input.turnClockwise ||
  input.turnCounterClockwise;

export const stepShipPhysics = (
  ship: ShipState,
  input: ShipInput,
  dtSeconds: number,
): ShipState => {
  const forward = getForwardVector(ship.rotation);
  const right = getRightVector(ship.rotation);
  const along = (input.thrustForward ? 1 : 0) - (input.thrustBackward ? 1 : 0);
  const across = (input.thrustRight ? 1 : 0) - (input.thrustLeft ? 1 : 0);
  const torqueDirection = (input.turnClockwise ? 1 : 0) - (input.turnCounterClockwise ? 1 : 0);

  let velocityX = ship.velocity.x;
  let velocityY = ship.velocity.y;
  let angularVelocity = ship.angularVelocity;

  if (hasAnyThrust(input)) {
    const forceX = forward.x * along * ship.model.thrustN + right.x * across * ship.model.thrustN;
    const forceY = forward.y * along * ship.model.thrustN + right.y * across * ship.model.thrustN;

    velocityX += (forceX / ship.model.massKg) * dtSeconds;
    velocityY += (forceY / ship.model.massKg) * dtSeconds;
    angularVelocity += (torqueDirection * ship.model.torqueNm / getMomentOfInertia(ship)) * dtSeconds;
  } else {
    const speed = Math.hypot(velocityX, velocityY);

    if (speed > EPSILON) {
      const brakeDelta = Math.min((ship.model.thrustN / ship.model.massKg) * dtSeconds, speed);

      velocityX -= (velocityX / speed) * brakeDelta;
      velocityY -= (velocityY / speed) * brakeDelta;
    }

    angularVelocity = brakeValue(
      angularVelocity,
      ship.model.torqueNm / getMomentOfInertia(ship),
      dtSeconds,
    );
  }

  return {
    ...ship,
    position: {
      x: ship.position.x + velocityX * dtSeconds,
      y: ship.position.y + velocityY * dtSeconds,
    },
    velocity: { x: velocityX, y: velocityY },
    rotation: ship.rotation + angularVelocity * dtSeconds,
    angularVelocity,
  };
};
