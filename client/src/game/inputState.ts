import type { ClientInputState } from "../network/protocol";

// Хранит текущее состояние клавиш по DOM-кодам KeyboardEvent.code.
export type KeyState = Record<string, boolean>;

// Переводит движение мыши в изменение целевого угла корабля.
export const MOUSE_TARGET_ROTATION_RADIANS_PER_PIXEL = 0.0025;

// Проверяет первое событие нажатия, игнорируя автоповторы удерживаемой клавиши.
export const isFreshKeyDown = (
  eventCode: string,
  wasAlreadyPressed: boolean,
  targetCode: string,
): boolean => eventCode === targetCode && !wasAlreadyPressed;

// Возвращает безопасное управление без тяги, когда пилот не захватил мышь.
export const emptyShipInput = (): ClientInputState => ({
  thrustForward: false,
  thrustBackward: false,
  thrustLeft: false,
  thrustRight: false,
  targetRotationDelta: 0,
});

// Собирает сетевой ввод из захвата мыши, WASD-клавиш и накопленной дельты мыши.
export const toShipInput = (
  isPointerLocked: boolean,
  keys: KeyState,
  mouseDeltaX: number,
): ClientInputState => {
  if (!isPointerLocked) {
    return emptyShipInput();
  }

  // Мышь задает изменение целевого угла, а поворотные двигатели догоняют эту цель на сервере.
  return {
    thrustForward: Boolean(keys.KeyW),
    thrustBackward: Boolean(keys.KeyS),
    thrustLeft: Boolean(keys.KeyA),
    thrustRight: Boolean(keys.KeyD),
    targetRotationDelta: mouseDeltaX * MOUSE_TARGET_ROTATION_RADIANS_PER_PIXEL,
  };
};
