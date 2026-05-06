import type { GameUiAction, GameUiControlKind, GameUiControlState } from "./types";

const dragKinds: ReadonlySet<GameUiControlKind> = new Set(["scrollbar", "slider", "splitter", "dragItem"]);

// Управляет hit-test, фокусом и захватом мыши для игрового HUD без DOM pointer events.
export class GameUiRuntime {
  // Последний снимок зарегистрированных контролов.
  private controls: GameUiControlState[] = [];
  // Контрол, над которым сейчас находится игровой курсор.
  private hoveredControlId: string | null = null;
  // Контрол, удерживающий левую кнопку мыши.
  private pressedControlId: string | null = null;
  // Контрол, получивший клавиатурный фокус.
  private focusedControlId: string | null = null;
  // Контрол, который держит drag-capture до отпускания кнопки.
  private activeCaptureControlId: string | null = null;

  // Заменяет registry актуальным снимком от SolidJS-слоя.
  updateControls(controls: GameUiControlState[]): void {
    this.controls = controls;
    const visibleIds = new Set(this.interactiveControls().map((control) => control.id));
    if (this.focusedControlId && !visibleIds.has(this.focusedControlId)) {
      this.focusedControlId = null;
    }
    if (this.hoveredControlId && !visibleIds.has(this.hoveredControlId)) {
      this.hoveredControlId = null;
    }
  }

  // Возвращает верхний доступный контрол под координатой игрового курсора.
  hitTest(x: number, y: number): GameUiControlState | null {
    const modalTop = this.topModal();
    if (modalTop && !contains(modalTop, x, y)) {
      return null;
    }

    return this.interactiveControls()
      .filter((control) => !modalTop || control.zIndex >= modalTop.zIndex)
      .filter((control) => contains(control, x, y))
      .sort((left, right) => right.zIndex - left.zIndex)[0] ?? null;
  }

  // Обновляет наведение и отдаёт drag-событие, если есть активный захват.
  pointerMove(x: number, y: number): GameUiAction | null {
    if (this.activeCaptureControlId) {
      const captured = this.controlById(this.activeCaptureControlId);
      return captured ? action("dragMove", captured, x, y) : null;
    }
    this.hoveredControlId = this.hitTest(x, y)?.id ?? null;
    return null;
  }

  // Начинает нажатие или перетаскивание на верхнем контроле.
  pointerDown(x: number, y: number, button: number): GameUiAction | null {
    if (button !== 0) {
      return null;
    }
    const target = this.hitTest(x, y);
    if (!target) {
      this.pressedControlId = null;
      this.focusedControlId = null;
      return null;
    }

    this.pressedControlId = target.id;
    if (target.focusable) {
      this.focusedControlId = target.id;
    }
    if (dragKinds.has(target.kind)) {
      this.activeCaptureControlId = target.id;
      return action("dragStart", target, x, y);
    }
    return null;
  }

  // Завершает нажатие или drag-capture.
  pointerUp(x: number, y: number, button: number): GameUiAction | null {
    if (button !== 0) {
      return null;
    }
    if (this.activeCaptureControlId) {
      const captured = this.controlById(this.activeCaptureControlId);
      this.activeCaptureControlId = null;
      this.pressedControlId = null;
      return captured ? action("dragEnd", captured, x, y) : null;
    }

    const pressed = this.pressedControlId ? this.controlById(this.pressedControlId) : null;
    const released = this.hitTest(x, y);
    this.pressedControlId = null;
    if (pressed && released?.id === pressed.id) {
      return action("click", pressed, x, y);
    }
    return null;
  }

  // Возвращает идентификаторы текущих transient-состояний для отрисовки.
  snapshot(): { hoveredControlId: string | null; pressedControlId: string | null; focusedControlId: string | null; activeCaptureControlId: string | null } {
    return {
      hoveredControlId: this.hoveredControlId,
      pressedControlId: this.pressedControlId,
      focusedControlId: this.focusedControlId,
      activeCaptureControlId: this.activeCaptureControlId,
    };
  }

  private interactiveControls(): GameUiControlState[] {
    return this.controls.filter((control) => control.visible && !control.disabled);
  }

  private controlById(id: string): GameUiControlState | null {
    return this.controls.find((control) => control.id === id) ?? null;
  }

  private topModal(): GameUiControlState | null {
    return this.interactiveControls()
      .filter((control) => control.kind === "modal")
      .sort((left, right) => right.zIndex - left.zIndex)[0] ?? null;
  }
}

const contains = (control: GameUiControlState, x: number, y: number): boolean => (
  x >= control.rect.left &&
  x <= control.rect.left + control.rect.width &&
  y >= control.rect.top &&
  y <= control.rect.top + control.rect.height
);

const action = (type: GameUiAction["type"], control: GameUiControlState, x: number, y: number): GameUiAction => {
  const result: GameUiAction = {
    type,
    controlId: control.id,
    kind: control.kind,
    x,
    y,
  };
  if (control.value !== null && control.value !== undefined) {
    result.value = control.value;
  }
  result.controlRect = control.rect;
  return result;
};
