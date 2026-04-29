// Общий форматтер нужен, чтобы debug overlay не размазывал правила округления по UI-коду.
export const formatNumber = (value: number, digits = 2): string =>
  Number.isFinite(value) ? value.toFixed(digits) : "NaN";
