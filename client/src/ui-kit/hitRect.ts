// Возвращает область, по которой игровой курсор должен попадать в UI Kit-контрол.
export const getUiKitControlHitRect = (element: HTMLElement): DOMRect => {
  const rect = element.getBoundingClientRect();
  if (element.dataset.uiKind !== "slider") {
    return rect;
  }

  const controlCell = element.closest<HTMLElement>(".control-panel-object-row__value--control");
  const controlCellRect = controlCell?.getBoundingClientRect();
  return controlCellRect && controlCellRect.width > rect.width ? controlCellRect : rect;
};
