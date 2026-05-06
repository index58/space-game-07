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

const addHudPanel = (rect: Partial<DOMRect> = {}) => {
  const panel = document.createElement("section");
  panel.className = "hud-panel";
  panel.getBoundingClientRect = () => ({
    x: rect.x ?? rect.left ?? 0,
    y: rect.y ?? rect.top ?? 0,
    left: rect.left ?? rect.x ?? 0,
    top: rect.top ?? rect.y ?? 0,
    right: rect.right ?? 0,
    bottom: rect.bottom ?? 0,
    width: rect.width ?? 1,
    height: rect.height ?? 1,
    toJSON: () => ({}),
  } as DOMRect);
  document.body.append(panel);
  return panel;
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
  document.body.innerHTML = "";
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
    addHudPanel({ left: 10, top: 840, right: 410, bottom: 990, width: 400, height: 150 });

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
    addHudPanel({ left: 225, top: 910, right: 775, bottom: 990, width: 550, height: 80 });

    pressKey("Enter", "Enter");
    releaseKey("Enter");
    for (const character of "@Pilot2 hello") {
      pressKey(`Key${character.toUpperCase()}`, character);
    }
    pressKey("Enter", "Enter");

    expect(controller.consumeChatAction()).toEqual({ targetNickname: "Pilot2", text: "hello" });
  });

  it("parses addressed chat command with quoted account nickname", () => {
    const canvas = document.createElement("canvas");
    const controller = new InputController(canvas);
    setPointerLockElement(canvas);
    addHudPanel({ left: 650, top: 720, right: 990, bottom: 990, width: 340, height: 270 });

    pressKey("Enter", "Enter");
    releaseKey("Enter");
    for (const character of '@"Mister Changle" Привет') {
      pressKey(`Key${character.toUpperCase()}`, character);
    }
    pressKey("Enter", "Enter");

    expect(controller.consumeChatAction()).toEqual({ targetNickname: "Mister Changle", text: "Привет" });
  });

  // Проверяет, что пробелы после адресного ника остаются в локальной строке до отправки.
  it("keeps spaces after addressed nickname while typing", () => {
    const canvas = document.createElement("canvas");
    const controller = new InputController(canvas);
    setPointerLockElement(canvas);
    addHudPanel({ left: 10, top: 10, right: 360, bottom: 190, width: 350, height: 180 });

    pressKey("Enter", "Enter");
    releaseKey("Enter");
    for (const character of "@Pilot3   ") {
      pressKey(character === " " ? "Space" : `Key${character.toUpperCase()}`, character);
    }

    expect(controller.getChatInputText()).toBe("@Pilot3   ");
  });

  // Проверяет, что стрелки двигают позицию каретки, которую HUD использует для отображения.
  it("moves visible chat cursor with arrow keys", () => {
    const canvas = document.createElement("canvas");
    const controller = new InputController(canvas);
    setPointerLockElement(canvas);
    addHudPanel({ left: 10, top: 840, right: 410, bottom: 990, width: 400, height: 150 });

    pressKey("Enter", "Enter");
    releaseKey("Enter");
    for (const character of "abc") {
      pressKey(`Key${character.toUpperCase()}`, character);
    }
    pressKey("ArrowLeft", "ArrowLeft");
    pressKey("ArrowLeft", "ArrowLeft");
    pressKey("ArrowRight", "ArrowRight");

    expect(controller.getChatCursorIndex()).toBe(2);
  });

  // Проверяет, что Home и End ставят каретку в начало и конец строки.
  it("moves chat cursor to text bounds with Home and End", () => {
    const canvas = document.createElement("canvas");
    const controller = new InputController(canvas);
    setPointerLockElement(canvas);
    addHudPanel({ left: 225, top: 910, right: 775, bottom: 990, width: 550, height: 80 });

    pressKey("Enter", "Enter");
    releaseKey("Enter");
    for (const character of "abc") {
      pressKey(`Key${character.toUpperCase()}`, character);
    }
    pressKey("Home", "Home");
    expect(controller.getChatCursorIndex()).toBe(0);

    pressKey("End", "End");

    expect(controller.getChatCursorIndex()).toBe(3);
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
    setViewportHeight(1000);
    const canvas = document.createElement("canvas");
    const controller = new InputController(canvas);
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
    moveMouse(-315, 105);
    window.dispatchEvent(new MouseEvent("contextmenu", { clientX: 185, clientY: 605, button: 2 }));
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

  // Проверяет, что левая кнопка выбирает вкладку под игровым указателем.
  it("selects chat tab with game cursor click", () => {
    setViewportHeight(1000);
    const canvas = document.createElement("canvas");
    const controller = new InputController(canvas);
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
    moveMouse(-315, 105);
    window.dispatchEvent(new MouseEvent("mousedown", { button: 0 }));
    const visibleState = controller.getVisibleChatState({
      type: "chatState",
      selectedChatId: 1,
      tabs: [
        { chatId: 1, title: "Server", communityTypeAcronym: "Server", duoChatKey: "", messages: [] },
        { chatId: 2, title: "Pilot2", communityTypeAcronym: "Duo", duoChatKey: "1:2", messages: [] },
      ],
    });

    expect(visibleState?.selectedChatId).toBe(2);
    expect(controller.consumeChatSelectAction()).toEqual({ chatId: 2 });
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
    addHudPanel({ left: 650, top: 720, right: 990, bottom: 990, width: 340, height: 270 });

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

  // Проверяет, что клик по космосу закрывает чатовый UI-режим и возвращает управление кораблем.
  it("leaves chat UI mode by clicking space outside HUD", () => {
    setViewportWidth(1000);
    setViewportHeight(1000);
    const canvas = document.createElement("canvas");
    const controller = new InputController(canvas);
    setPointerLockElement(canvas);
    addHudPanel({ left: 10, top: 10, right: 360, bottom: 190, width: 350, height: 180 });

    pressKey("Enter", "Enter");
    releaseKey("Enter");
    pressKey("KeyH", "h");
    moveMouse(0, -250);
    window.dispatchEvent(new MouseEvent("mousedown", { button: 0 }));
    pressKey("KeyW", "w");

    expect(controller.isChatInputFocused()).toBe(false);
    expect(controller.getChatInputText()).toBe("");
    expect(controller.getGameCursor().visible).toBe(false);
    expect(controller.consumeShipInput().thrustForward).toBe(true);
  });

  // Проверяет, что фактический DOM-прямоугольник HUD защищает от клика насквозь у краев.
  it("keeps chat UI mode when clicking actual HUD panel edge", () => {
    setViewportWidth(1000);
    setViewportHeight(1000);
    const canvas = document.createElement("canvas");
    const controller = new InputController(canvas);
    setPointerLockElement(canvas);
    addHudPanel({ left: 380, top: 238, right: 610, bottom: 260, width: 230, height: 22 });

    pressKey("Enter", "Enter");
    releaseKey("Enter");
    moveMouse(-118, -250);
    window.dispatchEvent(new MouseEvent("mousedown", { button: 0 }));

    expect(controller.isChatInputFocused()).toBe(true);
    expect(controller.getGameCursor().visible).toBe(true);
  });

  // Проверяет, что клик по панели основных показателей не возвращает управление кораблем.
  it("keeps chat UI mode when clicking object indicators panel", () => {
    setViewportWidth(1000);
    setViewportHeight(1000);
    const canvas = document.createElement("canvas");
    const controller = new InputController(canvas);
    setPointerLockElement(canvas);
    addHudPanel({ left: 10, top: 840, right: 410, bottom: 990, width: 400, height: 150 });

    pressKey("Enter", "Enter");
    releaseKey("Enter");
    moveMouse(-450, 450);
    window.dispatchEvent(new MouseEvent("mousedown", { button: 0 }));

    expect(controller.isChatInputFocused()).toBe(true);
    expect(controller.getGameCursor().visible).toBe(true);
  });

  // Проверяет, что клик по панели инструментов пилота не возвращает управление кораблем.
  it("keeps chat UI mode when clicking pilot toolbar panel", () => {
    setViewportWidth(1000);
    setViewportHeight(1000);
    const canvas = document.createElement("canvas");
    const controller = new InputController(canvas);
    setPointerLockElement(canvas);
    addHudPanel({ left: 225, top: 910, right: 775, bottom: 990, width: 550, height: 80 });

    pressKey("Enter", "Enter");
    releaseKey("Enter");
    moveMouse(0, 450);
    window.dispatchEvent(new MouseEvent("mousedown", { button: 0 }));

    expect(controller.isChatInputFocused()).toBe(true);
    expect(controller.getGameCursor().visible).toBe(true);
  });

  // Проверяет, что клик по мини-карте не возвращает управление кораблем.
  it("keeps chat UI mode when clicking minimap panel", () => {
    setViewportWidth(1000);
    setViewportHeight(1000);
    const canvas = document.createElement("canvas");
    const controller = new InputController(canvas);
    setPointerLockElement(canvas);
    addHudPanel({ left: 650, top: 720, right: 990, bottom: 990, width: 340, height: 270 });

    pressKey("Enter", "Enter");
    releaseKey("Enter");
    moveMouse(450, 450);
    window.dispatchEvent(new MouseEvent("mousedown", { button: 0 }));

    expect(controller.isChatInputFocused()).toBe(true);
    expect(controller.getGameCursor().visible).toBe(true);
  });

  // Проверяет, что клик по отладочной панели не возвращает управление кораблем.
  it("keeps chat UI mode when clicking debug overlay panel", () => {
    setViewportWidth(1000);
    setViewportHeight(1000);
    const canvas = document.createElement("canvas");
    const controller = new InputController(canvas);
    setPointerLockElement(canvas);
    addHudPanel({ left: 10, top: 10, right: 360, bottom: 190, width: 350, height: 180 });

    pressKey("Enter", "Enter");
    releaseKey("Enter");
    moveMouse(-450, -470);
    window.dispatchEvent(new MouseEvent("mousedown", { button: 0 }));

    expect(controller.isChatInputFocused()).toBe(true);
    expect(controller.getGameCursor().visible).toBe(true);
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

  // Проверяет, что новое сообщение в выбранной вкладке возвращает историю к последним строкам.
  it("scrolls selected chat to bottom when a new message arrives", () => {
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
    controller.getVisibleChatState({
      type: "chatState",
      selectedChatId: 1,
      tabs: [{ chatId: 1, title: "Server", communityTypeAcronym: "Server", duoChatKey: "", messages: messages(20) }],
    });
    expect(controller.getChatScrollState().contentOffsetPx).toBeGreaterThan(0);

    controller.getVisibleChatState({
      type: "chatState",
      selectedChatId: 1,
      tabs: [{ chatId: 1, title: "Server", communityTypeAcronym: "Server", duoChatKey: "", messages: messages(21) }],
    });

    expect(controller.getChatScrollState().contentOffsetPx).toBe(0);
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
