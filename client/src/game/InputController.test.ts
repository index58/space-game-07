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

const moveMouse = (movementX: number, movementY: number) => {
  const event = new MouseEvent("mousemove");
  Object.defineProperty(event, "movementX", { configurable: true, value: movementX });
  Object.defineProperty(event, "movementY", { configurable: true, value: movementY });
  window.dispatchEvent(event);
};

const setViewportHeight = (height: number) => {
  Object.defineProperty(window, "innerHeight", {
    configurable: true,
    value: height,
  });
};

const setViewportWidth = (width: number) => {
  Object.defineProperty(window, "innerWidth", {
    configurable: true,
    value: width,
  });
};

const messages = (count: number) => Array.from({ length: count }, (_, index) => ({
  id: index + 1,
  chatId: 1,
  senderCharacterId: 1,
  senderNickname: "Pilot1",
  messageTypeAcronym: "FromCharacter",
  text: `message-${index + 1}`,
  color: "D8F3FF",
  sentTime: "",
}));

afterEach(() => {
  setPointerLockElement(null);
  setViewportHeight(768);
  setViewportWidth(1024);
});

describe("InputController", () => {
  it("uses Shift and mouse wheel for pilot tool selection instead of zoom", () => {
    const canvas = document.createElement("canvas");
    const controller = new InputController(canvas);
    setPointerLockElement(canvas);

    wheel(-100, true);
    wheel(100, true);
    wheel(100, false);

    expect(controller.consumePilotToolSelectionDelta()).toBe(0);
    expect(controller.getZoom()).toBe(INITIAL_ZOOM - 1);
  });

  it("returns pilot tool wheel delta once", () => {
    const canvas = document.createElement("canvas");
    const controller = new InputController(canvas);
    setPointerLockElement(canvas);

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
    setPointerLockElement(canvas);
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

    expect(controller.isChatInputFocused()).toBe(false);
    expect(controller.consumeChatAction()).toEqual({ chatId: 5, text: "hi" });
    expect(controller.getChatInputText()).toBe("");
  });

  it("closes empty chat input on Enter without sending message", () => {
    const canvas = document.createElement("canvas");
    const controller = new InputController(canvas);
    setPointerLockElement(canvas);

    pressKey("Enter", "Enter");
    releaseKey("Enter");
    pressKey("Enter", "Enter");

    expect(controller.isChatInputFocused()).toBe(false);
    expect(controller.consumeChatAction()).toBeNull();
  });

  it("parses addressed chat command by account nickname", () => {
    const canvas = document.createElement("canvas");
    const controller = new InputController(canvas);
    setPointerLockElement(canvas);

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
    setPointerLockElement(canvas);
    controller.getVisibleChatState({
      type: "chatState",
      selectedChatId: 1,
      tabs: [
        { chatId: 1, title: "Server", communityTypeAcronym: "Server", duoChatKey: "", messages: [] },
        { chatId: 2, title: "Pilot2", communityTypeAcronym: "Duo", duoChatKey: "1:2", messages: [] },
      ],
    });

    pressKey("Enter", "Enter");
    releaseKey("Enter");
    moveMouse(-315, -155);
    window.dispatchEvent(new MouseEvent("contextmenu", { clientX: 185, clientY: 340, button: 2 }));
    expect(controller.getChatContextMenu()?.chatId).toBe(2);
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

  it("ignores game keyboard and wheel input while system pointer is unlocked", () => {
    const canvas = document.createElement("canvas");
    const controller = new InputController(canvas);

    pressKey("Enter", "Enter");
    pressKey("Backslash", "\\");
    pressKey("KeyO", "o");
    wheel(100, false);
    wheel(100, true);

    expect(controller.isChatInputFocused()).toBe(false);
    expect(controller.consumeRandomShipChangeRequest()).toBe(false);
    expect(controller.consumeBodyPolygonDebugToggleRequest()).toBe(false);
    expect(controller.consumePilotToolSelectionDelta()).toBe(0);
    expect(controller.getZoom()).toBe(INITIAL_ZOOM);
  });

  it("shows game cursor only while chat input is focused and moves it by mouse delta", () => {
    setViewportWidth(1000);
    setViewportHeight(1000);
    const canvas = document.createElement("canvas");
    const controller = new InputController(canvas);

    expect(controller.getGameCursor().visible).toBe(false);

    setPointerLockElement(canvas);
    moveMouse(15, -20);
    expect(controller.getGameCursor()).toEqual({ visible: false, x: 515, y: 480 });

    pressKey("Enter", "Enter");
    releaseKey("Enter");

    expect(controller.getGameCursor()).toEqual({ visible: true, x: 515, y: 480 });
  });

  it("suppresses ship controls while game cursor is visible", () => {
    const canvas = document.createElement("canvas");
    const controller = new InputController(canvas);
    setPointerLockElement(canvas);

    pressKey("Enter", "Enter");
    releaseKey("Enter");
    pressKey("KeyW", "w");
    pressKey("KeyP", "p");
    moveMouse(40, 0);

    expect(controller.consumeShipInput()).toEqual({
      thrustForward: false,
      thrustBackward: false,
      thrustLeft: false,
      thrustRight: false,
      toggleAnchor: false,
      targetRotationDelta: 0,
    });
  });

  it("scrolls selected chat with wheel when game cursor is over chat panel", () => {
    setViewportWidth(1000);
    setViewportHeight(1000);
    const canvas = document.createElement("canvas");
    const controller = new InputController(canvas);
    setPointerLockElement(canvas);
    controller.getVisibleChatState({
      type: "chatState",
      selectedChatId: 1,
      tabs: [{ chatId: 1, title: "Server", communityTypeAcronym: "Server", duoChatKey: "", messages: messages(20) }],
    });

    pressKey("Enter", "Enter");
    releaseKey("Enter");
    moveMouse(-300, 0);
    wheel(-100, false);
    const visibleState = controller.getVisibleChatState({
      type: "chatState",
      selectedChatId: 1,
      tabs: [{ chatId: 1, title: "Server", communityTypeAcronym: "Server", duoChatKey: "", messages: messages(20) }],
    });

    expect(visibleState?.tabs[0].messages.map((message) => message.id)).toEqual(messages(20).map((message) => message.id));
    expect(controller.getChatScrollState().visible).toBe(true);
    expect(controller.getChatScrollState().contentOffsetPx).toBeGreaterThan(0);
  });

  // Проверяет, что перетаскивание начинается на видимой полосе и двигает ползунок с той же скоростью, что и указатель.
  it("scrolls selected chat by dragging scrollbar with game cursor", () => {
    setViewportWidth(1000);
    setViewportHeight(1000);
    const canvas = document.createElement("canvas");
    const controller = new InputController(canvas);
    setPointerLockElement(canvas);
    controller.getVisibleChatState({
      type: "chatState",
      selectedChatId: 1,
      tabs: [{ chatId: 1, title: "Server", communityTypeAcronym: "Server", duoChatKey: "", messages: messages(20) }],
    });

    pressKey("Enter", "Enter");
    releaseKey("Enter");
    const initialScroll = controller.getChatScrollState();
    moveMouse(-28, 0);
    window.dispatchEvent(new MouseEvent("mousedown", { button: 0 }));
    controller.getVisibleChatState({
      type: "chatState",
      selectedChatId: 1,
      tabs: [{ chatId: 1, title: "Server", communityTypeAcronym: "Server", duoChatKey: "", messages: messages(20) }],
    });
    expect(controller.getChatScrollState().dragging).toBe(true);
    moveMouse(0, -20);
    window.dispatchEvent(new MouseEvent("mouseup", { button: 0 }));
    const visibleState = controller.getVisibleChatState({
      type: "chatState",
      selectedChatId: 1,
      tabs: [{ chatId: 1, title: "Server", communityTypeAcronym: "Server", duoChatKey: "", messages: messages(20) }],
    });

    expect(visibleState?.tabs[0].messages.map((message) => message.id)).toEqual(messages(20).map((message) => message.id));
    expect(controller.getChatScrollState().contentOffsetPx).toBeGreaterThan(0);
    const scrollbarTrackHeightPx = 221.2;
    expect((initialScroll.thumbTopPercent - controller.getChatScrollState().thumbTopPercent) * scrollbarTrackHeightPx / 100).toBeCloseTo(20);
  });

  // Проверяет, что левая кромка видимой полосы участвует в захвате, а не отрезается рамкой блока сообщений.
  it("starts chat scrollbar drag on the visible left edge", () => {
    setViewportWidth(1000);
    setViewportHeight(1000);
    const canvas = document.createElement("canvas");
    const controller = new InputController(canvas);
    setPointerLockElement(canvas);
    controller.getVisibleChatState({
      type: "chatState",
      selectedChatId: 1,
      tabs: [{ chatId: 1, title: "Server", communityTypeAcronym: "Server", duoChatKey: "", messages: messages(20) }],
    });

    pressKey("Enter", "Enter");
    releaseKey("Enter");
    moveMouse(-34, 0);
    window.dispatchEvent(new MouseEvent("mousedown", { button: 0 }));
    controller.getVisibleChatState({
      type: "chatState",
      selectedChatId: 1,
      tabs: [{ chatId: 1, title: "Server", communityTypeAcronym: "Server", duoChatKey: "", messages: messages(20) }],
    });

    expect(controller.getChatScrollState().dragging).toBe(true);
  });

  // Проверяет, что правее видимой полосы перетаскивание не начинается.
  it("does not drag chat scrollbar outside the visible track", () => {
    setViewportWidth(1000);
    setViewportHeight(1000);
    const canvas = document.createElement("canvas");
    const controller = new InputController(canvas);
    setPointerLockElement(canvas);
    controller.getVisibleChatState({
      type: "chatState",
      selectedChatId: 1,
      tabs: [{ chatId: 1, title: "Server", communityTypeAcronym: "Server", duoChatKey: "", messages: messages(20) }],
    });

    pressKey("Enter", "Enter");
    releaseKey("Enter");
    moveMouse(-18, 0);
    window.dispatchEvent(new MouseEvent("mousedown", { button: 0 }));
    moveMouse(0, -50);
    window.dispatchEvent(new MouseEvent("mouseup", { button: 0 }));
    controller.getVisibleChatState({
      type: "chatState",
      selectedChatId: 1,
      tabs: [{ chatId: 1, title: "Server", communityTypeAcronym: "Server", duoChatKey: "", messages: messages(20) }],
    });

    expect(controller.getChatScrollState().contentOffsetPx).toBe(0);
  });
});
