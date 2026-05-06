import { describe, expect, it } from "vitest";
import { applyUiKitDemoAction, createInitialUiKitDemoState } from "./showcaseState";
import type { GameUiAction, GameUiControlKind } from "./types";

const action = (controlId: string, kind: GameUiControlKind, value?: unknown): GameUiAction => ({
  // Создает минимальное действие игрового UI для проверки изменения витрины.
  type: "click",
  controlId,
  kind,
  x: 0,
  y: 0,
  value,
});

describe("ui kit showcase state", () => {
  // Проверяет, что витрина открывается со свернутым выпадающим списком.
  it("starts with closed dropdown", () => {
    expect(createInitialUiKitDemoState().dropdownOpen).toBe(false);
  });

  // Проверяет, что кнопки и переключатели изменяют состояние демонстрационной панели.
  it("applies button and toggle actions", () => {
    const first = createInitialUiKitDemoState();
    const afterButton = applyUiKitDemoAction(first, action("ui-kit-demo-button", "button"));
    const afterCheckbox = applyUiKitDemoAction(afterButton, action("ui-kit-demo-checkbox", "checkbox"));

    expect(afterButton.buttonClicks).toBe(1);
    expect(afterCheckbox.checkboxChecked).toBe(false);
  });

  // Проверяет, что списочные контролы выбирают значение из hit-test состояния.
  it("applies option selection actions", () => {
    const first = createInitialUiKitDemoState();
    const afterRadio = applyUiKitDemoAction(first, action("ui-kit-demo-radio-a", "radio", "a"));
    const afterDropdown = applyUiKitDemoAction(afterRadio, action("ui-kit-demo-select-one", "select", "one"));
    const afterTabs = applyUiKitDemoAction(afterDropdown, action("ui-kit-demo-tabs-one", "tabs", "one"));

    expect(afterRadio.radioValue).toBe("a");
    expect(afterDropdown.dropdownValue).toBe("one");
    expect(afterDropdown.dropdownOpen).toBe(false);
    expect(afterTabs.tabValue).toBe("one");
  });

  // Проверяет, что внешний клик закрывает раскрытый список без выбора нового пункта.
  it("closes dropdown from outside click actions", () => {
    const first = { ...createInitialUiKitDemoState(), dropdownOpen: true, dropdownValue: "two" };

    const afterButton = applyUiKitDemoAction(first, action("ui-kit-demo-button", "button"));
    const afterEmptyPanel = applyUiKitDemoAction(first, { ...action("", "button"), type: "cancel" });
    const afterModalPanel = applyUiKitDemoAction(first, action("ui-kit-showcase-modal", "modal"));

    expect(afterButton.dropdownOpen).toBe(false);
    expect(afterButton.dropdownValue).toBe("two");
    expect(afterEmptyPanel.dropdownOpen).toBe(false);
    expect(afterEmptyPanel.dropdownValue).toBe("two");
    expect(afterModalPanel.dropdownOpen).toBe(false);
    expect(afterModalPanel.dropdownValue).toBe("two");
  });

  // Проверяет, что drag-подобные контролы дают видимое числовое изменение.
  it("applies range and stepper actions", () => {
    const first = createInitialUiKitDemoState();
    const afterSlider = applyUiKitDemoAction(first, { ...action("ui-kit-demo-slider", "slider"), type: "dragMove" });
    const afterStepper = applyUiKitDemoAction(afterSlider, action("ui-kit-demo-stepper-increment", "stepper", "increment"));

    expect(afterSlider.sliderValue).toBe(55);
    expect(afterStepper.stepperValue).toBe(8);
  });

  // Проверяет, что демонстрационный ползунок идет за курсором без скачков по той же математике, что и история чата.
  it("drags scrollbar by captured cursor offset", () => {
    const first = createInitialUiKitDemoState();
    const controlRect = { left: 0, top: 0, width: 10, height: 100 };
    const afterStart = applyUiKitDemoAction(first, { ...action("ui-kit-demo-scrollbar", "scrollbar"), type: "dragStart", y: 36, controlRect });
    const afterMove = applyUiKitDemoAction(afterStart, { ...action("ui-kit-demo-scrollbar", "scrollbar"), type: "dragMove", y: 56, controlRect });
    const afterEnd = applyUiKitDemoAction(afterMove, { ...action("ui-kit-demo-scrollbar", "scrollbar"), type: "dragEnd", y: 56, controlRect });

    expect(afterStart.scrollbarDrag).toEqual({ grabOffsetPx: 16 });
    expect(afterMove.scrollbarTopPercent).toBeCloseTo(40);
    expect(afterEnd.scrollbarDrag).toBeNull();
  });
});
