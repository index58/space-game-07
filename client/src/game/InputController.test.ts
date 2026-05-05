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

const pressKey = (code: string, key = "") => {
  window.dispatchEvent(new KeyboardEvent("keydown", { code, key }));
};

const releaseKey = (code: string) => {
  window.dispatchEvent(new KeyboardEvent("keyup", { code }));
};

const setViewportHeight = (height: number) => {
  Object.defineProperty(window, "innerHeight", {
    configurable: true,
    value: height,
  });
};

afterEach(() => {
  setPointerLockElement(null);
  setViewportHeight(768);
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

  it("focuses chat with Enter, types text and sends it with second Enter", () => {
    const canvas = document.createElement("canvas");
    const controller = new InputController(canvas);
    controller.getVisibleChatState({
      type: "chatState",
      selectedChatId: 5,
      tabs: [{ chatId: 5, title: "Server", communityTypeAcronym: "Server", duoChatKey: "", messages: [] }],
    });

    pressKey("Enter", "Enter");
    releaseKey("Enter");
    pressKey("KeyH", "h");
    pressKey("KeyI", "i");
    pressKey("Enter", "Enter");

    expect(controller.isChatInputFocused()).toBe(true);
    expect(controller.consumeChatAction()).toEqual({ chatId: 5, text: "hi" });
    expect(controller.getChatInputText()).toBe("");
  });

  it("parses addressed chat command by account nickname", () => {
    const canvas = document.createElement("canvas");
    const controller = new InputController(canvas);

    pressKey("Enter", "Enter");
    releaseKey("Enter");
    for (const character of "@Pilot2 hello") {
      pressKey(`Key${character.toUpperCase()}`, character);
    }
    pressKey("Enter", "Enter");

    expect(controller.consumeChatAction()).toEqual({ targetNickname: "Pilot2", text: "hello" });
  });

  it("keeps server tab visible when duo tab is locally closed", () => {
    const canvas = document.createElement("canvas");
    const controller = new InputController(canvas);

    controller.closeChatTab(2, "Duo");
    const visibleState = controller.getVisibleChatState({
      type: "chatState",
      selectedChatId: 2,
      tabs: [
        { chatId: 1, title: "Server", communityTypeAcronym: "Server", duoChatKey: "", messages: [] },
        { chatId: 2, title: "Pilot2", communityTypeAcronym: "Duo", duoChatKey: "1:2", messages: [] },
      ],
    });

    expect(visibleState?.tabs.map((tab) => tab.chatId)).toEqual([1]);
    expect(visibleState?.selectedChatId).toBe(1);
  });

  it("opens game context menu on duo tab and closes it with left click", () => {
    const canvas = document.createElement("canvas");
    const controller = new InputController(canvas);
    setViewportHeight(1000);
    controller.getVisibleChatState({
      type: "chatState",
      selectedChatId: 1,
      tabs: [
        { chatId: 1, title: "Server", communityTypeAcronym: "Server", duoChatKey: "", messages: [] },
        { chatId: 2, title: "Pilot2", communityTypeAcronym: "Duo", duoChatKey: "1:2", messages: [] },
      ],
    });

    window.dispatchEvent(new MouseEvent("contextmenu", { clientX: 185, clientY: 340, button: 2 }));
    window.dispatchEvent(new MouseEvent("mousedown", { clientX: 186, clientY: 341, button: 0 }));
    const visibleState = controller.getVisibleChatState({
      type: "chatState",
      selectedChatId: 2,
      tabs: [
        { chatId: 1, title: "Server", communityTypeAcronym: "Server", duoChatKey: "", messages: [] },
        { chatId: 2, title: "Pilot2", communityTypeAcronym: "Duo", duoChatKey: "1:2", messages: [] },
      ],
    });

    expect(controller.getChatContextMenu()).toBeNull();
    expect(visibleState?.tabs.map((tab) => tab.chatId)).toEqual([1]);
  });
});
