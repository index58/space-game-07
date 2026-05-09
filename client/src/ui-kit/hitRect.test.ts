import { describe, expect, it } from "vitest";
import { getUiKitControlHitRect } from "./hitRect";

const rect = (left: number, top: number, width: number, height: number): DOMRect => ({
  x: left,
  y: top,
  left,
  top,
  right: left + width,
  bottom: top + height,
  width,
  height,
  toJSON: () => ({}),
} as DOMRect);

describe("getUiKitControlHitRect", () => {
  // Проверяет, что слайдер в строке формы получает всю правую ячейку для hit-test.
  it("uses full form control cell for slider hit test", () => {
    const cell = document.createElement("div");
    cell.className = "control-panel-object-row__value--control";
    const slider = document.createElement("div");
    slider.dataset.uiKind = "slider";
    cell.append(slider);
    document.body.append(cell);
    cell.getBoundingClientRect = () => rect(10, 20, 200, 30);
    slider.getBoundingClientRect = () => rect(50, 24, 80, 22);

    expect(getUiKitControlHitRect(slider)).toMatchObject({ left: 10, top: 20, width: 200, height: 30 });
  });

  // Проверяет, что обычные контролы используют собственные реальные границы.
  it("keeps own rect for non-slider controls", () => {
    const button = document.createElement("div");
    button.dataset.uiKind = "button";
    button.getBoundingClientRect = () => rect(30, 40, 70, 20);

    expect(getUiKitControlHitRect(button)).toMatchObject({ left: 30, top: 40, width: 70, height: 20 });
  });
});
