import { describe, expect, it } from "vitest";
import { GameUiRuntime } from "./runtime";
import type { GameUiControlState } from "./types";

const control = (partial: Partial<GameUiControlState> = {}): GameUiControlState => ({
  // Создаёт минимальный видимый контрол для проверки маршрутизации игрового UI.
  id: "control",
  kind: "button",
  rect: { left: 0, top: 0, width: 100, height: 40 },
  zIndex: 0,
  disabled: false,
  visible: true,
  focusable: true,
  value: null,
  ...partial,
});

describe("GameUiRuntime", () => {
  // Проверяет, что попадание выбирает верхний видимый контрол, а не первый зарегистрированный.
  it("hits the top visible control by z index", () => {
    const runtime = new GameUiRuntime();

    runtime.updateControls([
      control({ id: "bottom", zIndex: 1 }),
      control({ id: "top", zIndex: 5 }),
    ]);

    expect(runtime.hitTest(10, 10)?.id).toBe("top");
  });

  // Проверяет, что недоступные и скрытые элементы не получают наведение и клики.
  it("ignores disabled and hidden controls during hit test", () => {
    const runtime = new GameUiRuntime();

    runtime.updateControls([
      control({ id: "disabled", zIndex: 5, disabled: true }),
      control({ id: "hidden", zIndex: 4, visible: false }),
      control({ id: "active", zIndex: 1 }),
    ]);

    expect(runtime.hitTest(10, 10)?.id).toBe("active");
  });

  // Проверяет, что модальное окно перекрывает элементы за пределами своего слоя.
  it("lets modal controls block lower controls outside modal bounds", () => {
    const runtime = new GameUiRuntime();

    runtime.updateControls([
      control({ id: "behind", zIndex: 1, rect: { left: 0, top: 0, width: 200, height: 200 } }),
      control({ id: "modal", kind: "modal", zIndex: 10, rect: { left: 40, top: 40, width: 80, height: 80 }, focusable: false }),
    ]);

    expect(runtime.hitTest(10, 10)).toBeNull();
    expect(runtime.hitTest(50, 50)?.id).toBe("modal");
  });

  // Проверяет, что перетаскивание остаётся у исходного контрола до отпускания кнопки.
  it("keeps drag capture until mouse up", () => {
    const runtime = new GameUiRuntime();

    runtime.updateControls([
      control({ id: "drag", kind: "slider", rect: { left: 0, top: 0, width: 100, height: 20 } }),
    ]);

    const rect = { left: 0, top: 0, width: 100, height: 20 };

    expect(runtime.pointerDown(10, 10, 0)).toEqual({ type: "dragStart", controlId: "drag", kind: "slider", x: 10, y: 10, controlRect: rect });
    expect(runtime.pointerMove(300, 300)).toEqual({ type: "dragMove", controlId: "drag", kind: "slider", x: 300, y: 300, controlRect: rect });
    expect(runtime.pointerUp(300, 300, 0)).toEqual({ type: "dragEnd", controlId: "drag", kind: "slider", x: 300, y: 300, controlRect: rect });
  });

  // Проверяет, что обычная кнопка создаёт действие клика только при отпускании над тем же элементом.
  it("emits click for button press and release on same control", () => {
    const runtime = new GameUiRuntime();

    runtime.updateControls([control({ id: "ok" })]);

    expect(runtime.pointerDown(10, 10, 0)).toBeNull();
    expect(runtime.pointerUp(10, 10, 0)).toEqual({ type: "click", controlId: "ok", kind: "button", x: 10, y: 10, controlRect: { left: 0, top: 0, width: 100, height: 40 } });
  });

  // Проверяет, что зажатая кнопка получает общий визуальный класс до отпускания мыши.
  it("toggles pressed class for button while mouse is held", () => {
    const runtime = new GameUiRuntime();
    const element = document.createElement("div");
    element.id = "ok";
    element.className = "ui-kit-button";
    document.body.append(element);

    runtime.updateControls([control({ id: "ok" })]);

    runtime.pointerDown(10, 10, 0);
    expect(element.classList.contains("is-pressed")).toBe(true);

    runtime.pointerUp(10, 10, 0);
    expect(element.classList.contains("is-pressed")).toBe(false);
  });

  // Проверяет, что слайдер внутри модального окна получает перетаскивание поверх фонового слоя.
  it("emits drag start for slider inside modal bounds", () => {
    const runtime = new GameUiRuntime();

    runtime.updateControls([
      control({ id: "behind", zIndex: 1, rect: { left: 0, top: 0, width: 240, height: 180 } }),
      control({ id: "modal", kind: "modal", zIndex: 10, rect: { left: 40, top: 40, width: 160, height: 100 }, focusable: false }),
      control({ id: "modal-slider", kind: "slider", zIndex: 11, rect: { left: 60, top: 90, width: 100, height: 20 } }),
    ]);

    expect(runtime.pointerDown(110, 100, 0)).toEqual({
      type: "dragStart",
      controlId: "modal-slider",
      kind: "slider",
      x: 110,
      y: 100,
      controlRect: { left: 60, top: 90, width: 100, height: 20 },
    });
  });

  // Проверяет, что клик мимо контролов отдаёт действие закрытия для раскрытых overlay-контролов.
  it("emits cancel when pressing outside controls", () => {
    const runtime = new GameUiRuntime();

    runtime.updateControls([control({ id: "ok" })]);

    expect(runtime.pointerDown(150, 80, 0)).toEqual({ type: "cancel", controlId: "", kind: "button", x: 150, y: 80 });
  });

  // Проверяет, что пункт раскрытого списка получает клик выше корневого поля.
  it("emits click for dropdown item above parent control", () => {
    const runtime = new GameUiRuntime();

    runtime.updateControls([
      control({ id: "select", kind: "select", rect: { left: 10, top: 10, width: 120, height: 40 }, zIndex: 1 }),
      control({ id: "select-one", kind: "select", rect: { left: 10, top: 54, width: 120, height: 30 }, zIndex: 5, value: "one" }),
    ]);

    expect(runtime.pointerDown(20, 60, 0)).toBeNull();
    expect(runtime.pointerUp(20, 60, 0)).toEqual({
      type: "click",
      controlId: "select-one",
      kind: "select",
      x: 20,
      y: 60,
      value: "one",
      controlRect: { left: 10, top: 54, width: 120, height: 30 },
    });
  });

  // Проверяет, что перехватчик раскрытого списка забирает внешний клик у нижнего элемента.
  it("uses dropdown outside blocker above lower controls", () => {
    const runtime = new GameUiRuntime();

    runtime.updateControls([
      control({ id: "button", kind: "button", rect: { left: 0, top: 0, width: 200, height: 200 }, zIndex: 1 }),
      control({ id: "select-outside-blocker", kind: "modal", rect: { left: 0, top: 0, width: 300, height: 300 }, zIndex: 900, focusable: false }),
      control({ id: "select-one", kind: "select", rect: { left: 10, top: 54, width: 120, height: 30 }, zIndex: 1000, value: "one" }),
    ]);

    expect(runtime.hitTest(150, 150)?.id).toBe("select-outside-blocker");
    expect(runtime.hitTest(20, 60)?.id).toBe("select-one");
  });
});
