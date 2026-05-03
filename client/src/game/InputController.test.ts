import { afterEach, describe, expect, it } from "vitest";
import { INITIAL_ZOOM } from "../domain/camera";
import { InputController } from "./InputController";

const wheel = (deltaY: number, shiftKey: boolean) => {
  window.dispatchEvent(new WheelEvent("wheel", { deltaY, shiftKey }));
};

const setPointerLockElement = (element: Element | null) => {
  Object.defineProperty(document, "pointerLockElement", {
    configurable: true,
    value: element,
  });
};

const pressKey = (code: string) => {
  window.dispatchEvent(new KeyboardEvent("keydown", { code }));
};

afterEach(() => {
  setPointerLockElement(null);
});

describe("InputController", () => {
  it("uses Shift and mouse wheel for pilot tool selection instead of zoom", () => {
    const canvas = document.createElement("canvas");
    const controller = new InputController(canvas);

    wheel(-100, true);
    wheel(100, true);
    wheel(100, false);

    expect(controller.consumePilotToolSelectionDelta()).toBe(0);
    expect(controller.getZoom()).toBe(INITIAL_ZOOM - 1);
  });

  it("returns pilot tool wheel delta once", () => {
    const canvas = document.createElement("canvas");
    const controller = new InputController(canvas);

    wheel(-100, true);
    expect(controller.consumePilotToolSelectionDelta()).toBe(-1);
    expect(controller.consumePilotToolSelectionDelta()).toBe(0);

    wheel(100, true);
    expect(controller.consumePilotToolSelectionDelta()).toBe(1);
  });

  it("returns anchor toggle once from P while pointer is locked", () => {
    const canvas = document.createElement("canvas");
    const controller = new InputController(canvas);
    setPointerLockElement(canvas);

    pressKey("KeyP");

    expect(controller.consumeShipInput().toggleAnchor).toBe(true);
    expect(controller.consumeShipInput().toggleAnchor).toBe(false);
  });
});
