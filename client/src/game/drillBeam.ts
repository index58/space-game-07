export const SIMPLE_DRILL_RAY_ACRONYM = "SimpleDrillRay";

export type DrillBeamPoint = {
  // Горизонтальная координата в пикселях экрана.
  x: number;
  // Вертикальная координата в пикселях экрана.
  y: number;
};

export type DrillBeamGeometry = {
  // Точка выхода луча из носа корабля.
  start: DrillBeamPoint;
  // Точка контакта на предельной дальности инструмента.
  end: DrillBeamPoint;
  // Длина видимой части в пикселях экрана.
  lengthPx: number;
  // Базовая толщина яркого ядра в пикселях экрана.
  widthPx: number;
};

export type DrillBeamInput = {
  // Экранный центр объекта луча из серверного снимка.
  center: DrillBeamPoint;
  // Экранный поворот объекта луча относительно камеры пилота.
  rotation: number;
  // Длина луча в метрах мира.
  lengthMeters: number;
  // Текущий масштаб камеры в пикселях на метр.
  zoomScale: number;
};

// Считает экранные точки луча, не привязывая эффект к Phaser.
export const getDrillBeamGeometry = (input: DrillBeamInput): DrillBeamGeometry | null => {
  const lengthPx = input.lengthMeters * input.zoomScale;
  if (lengthPx <= 0) {
    return null;
  }

  const halfLength = lengthPx / 2;
  const forward = {
    x: Math.sin(input.rotation),
    y: -Math.cos(input.rotation),
  };

  return {
    start: {
      x: input.center.x - forward.x * halfLength,
      y: input.center.y - forward.y * halfLength,
    },
    end: {
      x: input.center.x + forward.x * halfLength,
      y: input.center.y + forward.y * halfLength,
    },
    lengthPx,
    widthPx: input.zoomScale * 3,
  };
};

// Возвращает положение внутренних штрихов так, чтобы движение шло от цели к кораблю.
export const getDrillBeamIntakeProgress = (timeMs: number, index: number): number => {
  const progress = (timeMs * 0.00042 + index * 0.173) % 1;
  return 1 - progress;
};
