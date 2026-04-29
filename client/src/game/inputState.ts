import type { ShipInput } from "../domain/physics";

export type KeyState = Record<string, boolean>;

export const MOUSE_TURN_IMPULSE_SECONDS = 0.05;

export const emptyShipInput = (): ShipInput => ({
  thrustForward: false,
  thrustBackward: false,
  thrustLeft: false,
  thrustRight: false,
  turnClockwise: false,
  turnCounterClockwise: false,
});

export const toShipInput = (
  isPointerLocked: boolean,
  keys: KeyState,
  mouseDeltaX: number,
): ShipInput => {
  if (!isPointerLocked) {
    return emptyShipInput();
  }

  // Любой ненулевой сдвиг мыши считается командой поворота на текущий кадр.
  return {
    thrustForward: Boolean(keys.KeyW),
    thrustBackward: Boolean(keys.KeyS),
    thrustLeft: Boolean(keys.KeyA),
    thrustRight: Boolean(keys.KeyD),
    turnClockwise: mouseDeltaX > 0,
    turnCounterClockwise: mouseDeltaX < 0,
  };
};

export class MouseTurnImpulse {
  private direction = 0;
  private remainingSeconds = 0;

  addMouseDelta(mouseDeltaX: number): void {
    if (mouseDeltaX === 0) {
      return;
    }

    this.direction = Math.sign(mouseDeltaX);
    this.remainingSeconds = MOUSE_TURN_IMPULSE_SECONDS;
  }

  consume(dtSeconds: number): Pick<ShipInput, "turnClockwise" | "turnCounterClockwise"> {
    const direction = this.remainingSeconds > 0 ? this.direction : 0;
    this.remainingSeconds = Math.max(0, this.remainingSeconds - dtSeconds);

    return {
      turnClockwise: direction > 0,
      turnCounterClockwise: direction < 0,
    };
  }

  clear(): void {
    this.direction = 0;
    this.remainingSeconds = 0;
  }
}
