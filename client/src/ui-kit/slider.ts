// Переводит позицию указателя на шкале количества в целое число включенных единиц.
export const getCountSliderValue = (position: number, totalCount: number): number => {
  const maximum = Math.max(1, totalCount);
  return Math.min(maximum, Math.max(1, Math.round(Math.min(1, Math.max(0, position)) * maximum)));
};
