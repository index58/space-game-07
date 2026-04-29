import type { ShipInput } from "../domain/physics";

export type KeyState = Record<string, boolean>;

export const MOUSE_TARGET_ROTATION_RADIANS_PER_PIXEL = 0.0025;

export const emptyShipInput = (): ShipInput => ({
  thrustForward: false,
  thrustBackward: false,
  thrustLeft: false,
  thrustRight: false,
  targetRotationDelta: 0,
});

export const toShipInput = (
  isPointerLocked: boolean,
  keys: KeyState,
  mouseDeltaX: number,
): ShipInput => {
  if (!isPointerLocked) {
    return emptyShipInput();
  }

  // Мышь задает изменение целевого угла, а поворотные двигатели догоняют эту цель в физике.
  return {
    thrustForward: Boolean(keys.KeyW),
    thrustBackward: Boolean(keys.KeyS),
    thrustLeft: Boolean(keys.KeyA),
    thrustRight: Boolean(keys.KeyD),
    targetRotationDelta: mouseDeltaX * MOUSE_TARGET_ROTATION_RADIANS_PER_PIXEL,
  };
};
