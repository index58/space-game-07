import type { ClientInputState } from "../network/protocol";

// Хранит текущее состояние клавиш по DOM-кодам KeyboardEvent.code.
export type KeyState = Record<string, boolean>;

// Переводит движение мыши в изменение целевого угла корабля.
export const MOUSE_TARGET_ROTATION_RADIANS_PER_PIXEL = 0.0025;

export type InputBindingMap = Record<string, string>;

const defaultInputBindings: InputBindingMap = {
  ThrustForward: "KeyboardEvent.code:KeyW",
  ThrustBackward: "KeyboardEvent.code:KeyS",
  ThrustLeft: "KeyboardEvent.code:KeyA",
  ThrustRight: "KeyboardEvent.code:KeyD",
  RotateClockwise: "MouseEvent.movementX>0",
  RotateCounterclockwise: "MouseEvent.movementX<0",
};

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
  toggleAnchor: false,
  targetRotationDelta: 0,
});

// Собирает сетевой ввод из захвата мыши, WASD-клавиш и накопленной дельты мыши.
export const toShipInput = (
  isPointerLocked: boolean,
  keys: KeyState,
  mouseDeltaX: number,
  bindings: InputBindingMap = defaultInputBindings,
): ClientInputState => {
  if (!isPointerLocked) {
    return emptyShipInput();
  }
  const activeBindings = { ...defaultInputBindings, ...bindings };

  // Мышь задает изменение целевого угла, а поворотные двигатели догоняют эту цель на сервере.
  return {
    thrustForward: isKeyboardBindingPressed(keys, activeBindings.ThrustForward),
    thrustBackward: isKeyboardBindingPressed(keys, activeBindings.ThrustBackward),
    thrustLeft: isKeyboardBindingPressed(keys, activeBindings.ThrustLeft),
    thrustRight: isKeyboardBindingPressed(keys, activeBindings.ThrustRight),
    toggleAnchor: false,
    targetRotationDelta: rotationDeltaFromBindings(keys, mouseDeltaX, activeBindings),
  };
};

// Проверяет удержание клавиши, если действие привязано к DOM-коду клавиатуры.
export const isKeyboardBindingPressed = (keys: KeyState, systemStringValue: string | undefined): boolean => {
  const keyCode = keyboardCodeFromSystemString(systemStringValue);
  return keyCode ? Boolean(keys[keyCode]) : false;
};

// Проверяет первое нажатие клавиши для дискретного действия.
export const isFreshKeyboardBinding = (
  eventCode: string,
  wasAlreadyPressed: boolean,
  systemStringValue: string | undefined,
): boolean => {
  const keyCode = keyboardCodeFromSystemString(systemStringValue);
  return keyCode ? isFreshKeyDown(eventCode, wasAlreadyPressed, keyCode) : false;
};

// Вычисляет поворот из мыши или клавиш, если игрок переназначил вращение.
const rotationDeltaFromBindings = (keys: KeyState, mouseDeltaX: number, bindings: InputBindingMap): number => {
  let delta = 0;
  if (bindings.RotateClockwise === "MouseEvent.movementX>0" || bindings.RotateCounterclockwise === "MouseEvent.movementX<0") {
    delta += mouseDeltaX * MOUSE_TARGET_ROTATION_RADIANS_PER_PIXEL;
  }
  if (isKeyboardBindingPressed(keys, bindings.RotateClockwise)) {
    delta += 0.05;
  }
  if (isKeyboardBindingPressed(keys, bindings.RotateCounterclockwise)) {
    delta -= 0.05;
  }
  return delta;
};

// Достает DOM-код клавиши из системной строки справочника.
const keyboardCodeFromSystemString = (systemStringValue: string | undefined): string | null => {
  const match = systemStringValue?.match(/^KeyboardEvent\.code:(.+)$/);
  return match?.[1] ?? null;
};
