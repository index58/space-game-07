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

const pressInputKey = (element: HTMLElement, code: string, key = "") => {
  element.dispatchEvent(new KeyboardEvent("keydown", { code, key, bubbles: true }));
};

const waitForNativeEditSync = () => new Promise((resolve) => {
  window.setTimeout(resolve, 0);
});

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

const addChatInputDom = () => {
  const input = document.createElement("div");
  input.id = "chat-input";
  input.getBoundingClientRect = () => ({
    x: 500,
    y: 370,
    left: 500,
    top: 370,
    right: 620,
    bottom: 400,
    width: 120,
    height: 30,
    toJSON: () => ({}),
  } as DOMRect);
  const viewport = document.createElement("div");
  viewport.className = "chat-input__viewport";
  viewport.getBoundingClientRect = () => ({
    x: 505,
    y: 370,
    left: 505,
    top: 370,
    right: 615,
    bottom: 400,
    width: 110,
    height: 30,
    toJSON: () => ({}),
  } as DOMRect);
  const text = document.createElement("span");
  text.className = "chat-input__text";
  text.style.transform = "translateX(0px)";
  const measure = document.createElement("span");
  measure.className = "chat-input__measure";
  measure.getBoundingClientRect = () => ({
    x: 0,
    y: 0,
    left: 0,
    top: 0,
    right: 70,
    bottom: 20,
    width: 70,
    height: 20,
    toJSON: () => ({}),
  } as DOMRect);
  viewport.append(text, measure);
  input.append(viewport);
  document.body.append(input);
};

const addChatMessagesDom = (contentHeight: number) => {
  const messagesPanel = document.createElement("div");
  messagesPanel.className = "chat-messages";
  messagesPanel.getBoundingClientRect = () => ({
    x: 10,
    y: 100,
    left: 10,
    top: 100,
    right: 490,
    bottom: 340,
    width: 480,
    height: 240,
    toJSON: () => ({}),
  } as DOMRect);
  const content = document.createElement("div");
  content.className = "chat-messages__content";
  content.getBoundingClientRect = () => ({
    x: 18,
    y: 0,
    left: 18,
    top: 0,
    right: 450,
    bottom: contentHeight,
    width: 432,
    height: contentHeight,
    toJSON: () => ({}),
  } as DOMRect);
  messagesPanel.append(content);
  document.body.append(messagesPanel);
};

const addChatTabDom = (id: number, rect: Partial<DOMRect>) => {
  const tab = document.createElement("div");
  tab.id = `chat-tab-${id}`;
  tab.className = "chat-tab";
  tab.getBoundingClientRect = () => ({
    x: rect.x ?? rect.left ?? 0,
    y: rect.y ?? rect.top ?? 0,
    left: rect.left ?? rect.x ?? 0,
    top: rect.top ?? rect.y ?? 0,
    right: rect.right ?? (rect.left ?? rect.x ?? 0) + (rect.width ?? 1),
    bottom: rect.bottom ?? (rect.top ?? rect.y ?? 0) + (rect.height ?? 1),
    width: rect.width ?? 1,
    height: rect.height ?? 1,
    toJSON: () => ({}),
  } as DOMRect);
  document.body.append(tab);
  return tab;
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

  // Проверяет, что выбор вкладки чата использует фактическую DOM-ширину, а не старую фиксированную сетку.
  it("selects chat tab by rendered bounds", () => {
    setViewportWidth(1000);
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
    addChatTabDom(1, { left: 20, top: 590, width: 76, height: 25 });
    addChatTabDom(2, { left: 100, top: 590, width: 64, height: 25 });

    pressKey("Enter", "Enter");
    releaseKey("Enter");
    moveMouse(-370, 100);
    window.dispatchEvent(new MouseEvent("mousedown", { button: 0 }));

    expect(controller.consumeChatSelectAction()).toEqual({ chatId: 2 });
  });

  // Проверяет, что внешний клик раскрытого списка не проходит в лежащую ниже вкладку чата.
  it("blocks chat tab selection under dropdown outside blocker", () => {
    setViewportWidth(1000);
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
    controller.updateGameUiControls([
      {
        id: "settings-input-select-1-outside-blocker",
        kind: "modal",
        rect: { left: 0, top: 0, width: 1000, height: 1000 },
        zIndex: 900,
        disabled: false,
        visible: true,
        focusable: false,
        value: null,
      },
    ]);

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

    expect(visibleState?.selectedChatId).toBe(1);
    expect(controller.consumeChatSelectAction()).toBeNull();
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

  // Проверяет, что игровой указатель появляется в точке последнего клика перед захватом мыши.
  it("places game cursor at the system pointer lock click point", () => {
    setViewportWidth(1000);
    setViewportHeight(1000);
    const canvas = document.createElement("canvas");
    canvas.requestPointerLock = () => Promise.resolve();
    const controller = new InputController(canvas);

    canvas.dispatchEvent(new MouseEvent("click", { clientX: 240, clientY: 360 }));
    setPointerLockElement(canvas);
    pressKey("Enter", "Enter");

    expect(controller.getGameCursor()).toEqual({ visible: true, x: 240, y: 360 });
  });

  // Проверяет, что клики после захвата мыши не переносят игровой указатель как системные.
  it("ignores canvas click coordinates while pointer is already locked", () => {
    setViewportWidth(1000);
    setViewportHeight(1000);
    const canvas = document.createElement("canvas");
    canvas.requestPointerLock = () => Promise.resolve();
    const controller = new InputController(canvas);

    canvas.dispatchEvent(new MouseEvent("click", { clientX: 240, clientY: 360 }));
    setPointerLockElement(canvas);
    canvas.dispatchEvent(new MouseEvent("click", { clientX: 900, clientY: 920 }));
    pressKey("Enter", "Enter");

    expect(controller.getGameCursor()).toEqual({ visible: true, x: 240, y: 360 });
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
  // Проверяет, что витрина UI Kit включает игровой курсор и блокирует управление кораблём.
  it("shows game cursor and suppresses ship controls while UI kit showcase is visible", () => {
    const canvas = document.createElement("canvas");
    const controller = new InputController(canvas);
    setPointerLockElement(canvas);

    pressKey("F9", "F9");
    pressKey("KeyW", "w");

    expect(controller.isUiKitShowcaseVisible()).toBe(true);
    expect(controller.getGameCursor().visible).toBe(true);
    expect(controller.consumeShipInput().thrustForward).toBe(false);
  });

  // Проверяет, что витрина UI Kit открывается даже когда фокус находится в строке чата.
  it("toggles UI kit showcase while chat input is focused", () => {
    const canvas = document.createElement("canvas");
    const controller = new InputController(canvas);
    setPointerLockElement(canvas);

    pressKey("Enter", "Enter");
    releaseKey("Enter");
    pressKey("F9", "F9");

    expect(controller.isUiKitShowcaseVisible()).toBe(true);
  });

  // Проверяет, что открытие витрины UI Kit закрывает другое модальное окно.
  it("closes settings when opening UI kit showcase", () => {
    const canvas = document.createElement("canvas");
    const controller = new InputController(canvas);
    setPointerLockElement(canvas);

    pressKey("F10", "F10");
    releaseKey("F10");
    pressKey("F9", "F9");

    expect(controller.isSettingsVisible()).toBe(false);
    expect(controller.isUiKitShowcaseVisible()).toBe(true);
  });

  // Проверяет, что открытие настроек закрывает другое модальное окно.
  it("closes UI kit showcase when opening settings", () => {
    const canvas = document.createElement("canvas");
    const controller = new InputController(canvas);
    setPointerLockElement(canvas);

    pressKey("F9", "F9");
    releaseKey("F9");
    pressKey("F10", "F10");

    expect(controller.isUiKitShowcaseVisible()).toBe(false);
    expect(controller.isSettingsVisible()).toBe(true);
  });

  // Проверяет, что повторная команда текущего модального окна закрывает его без открытия другого.
  it("closes current modal window when toggling the same modal again", () => {
    const canvas = document.createElement("canvas");
    const controller = new InputController(canvas);
    setPointerLockElement(canvas);

    pressKey("F10", "F10");
    releaseKey("F10");
    pressKey("F10", "F10");

    expect(controller.isSettingsVisible()).toBe(false);
    expect(controller.isUiKitShowcaseVisible()).toBe(false);
  });

  // Проверяет, что общий runtime отдаёт действие по зарегистрированному HUD-контролу.
  it("queues game UI runtime actions from registered controls", () => {
    const canvas = document.createElement("canvas");
    const controller = new InputController(canvas);
    setPointerLockElement(canvas);
    controller.updateGameUiControls([{
      id: "demo-button",
      kind: "button",
      rect: { left: 500, top: 370, width: 60, height: 40 },
      zIndex: 1,
      disabled: false,
      visible: true,
      focusable: true,
      value: null,
    }]);

    pressKey("F9", "F9");
    window.dispatchEvent(new MouseEvent("mousedown", { button: 0 }));
    window.dispatchEvent(new MouseEvent("mouseup", { button: 0 }));

    expect(controller.consumeGameUiAction()).toMatchObject({ type: "click", controlId: "demo-button" });
  });

  // Проверяет, что пункт выпадающего списка вне прямоугольника панели не считается кликом по космосу.
  it("keeps UI kit showcase visible when clicking dropdown item outside panel bounds", () => {
    const canvas = document.createElement("canvas");
    const controller = new InputController(canvas);
    setPointerLockElement(canvas);
    addHudPanel({ left: 10, top: 10, right: 130, bottom: 50, width: 120, height: 40 });
    controller.updateGameUiControls([{
      id: "ui-kit-demo-select-one",
      kind: "select",
      rect: { left: 10, top: 54, width: 120, height: 30 },
      zIndex: 5,
      disabled: false,
      visible: true,
      focusable: true,
      value: "one",
    }]);

    pressKey("F9", "F9");
    moveMouse(-492, -324);
    window.dispatchEvent(new MouseEvent("mousedown", { button: 0 }));
    window.dispatchEvent(new MouseEvent("mouseup", { button: 0 }));

    expect(controller.isUiKitShowcaseVisible()).toBe(true);
    expect(controller.consumeGameUiAction()).toMatchObject({ type: "click", controlId: "ui-kit-demo-select-one", value: "one" });
  });

  // Проверяет, что двойной клик игровым курсором по строке ввода выделяет слово.
  it("selects chat input word by game cursor double click", () => {
    const canvas = document.createElement("canvas");
    const controller = new InputController(canvas);
    setPointerLockElement(canvas);
    addChatInputDom();

    pressKey("Enter", "Enter");
    releaseKey("Enter");
    for (const character of "abc def") {
      pressKey(`Key${character.toUpperCase()}`, character);
    }
    moveMouse(28, 0);
    window.dispatchEvent(new MouseEvent("mousedown", { button: 0, detail: 2 }));

    expect(controller.getChatSelectionStart()).toBe(4);
    expect(controller.getChatSelectionEnd()).toBe(7);
  });

  // Проверяет, что навигационные клавиши native textarea обновляют видимую каретку HUD.
  it("syncs chat cursor after native arrow home and end keys", async () => {
    const canvas = document.createElement("canvas");
    const controller = new InputController(canvas);
    setPointerLockElement(canvas);

    pressKey("Enter", "Enter");
    releaseKey("Enter");
    for (const character of "abcd") {
      pressKey(`Key${character.toUpperCase()}`, character);
    }

    const textarea = document.querySelector<HTMLTextAreaElement>("[data-ui-kit-edit-id='chat-input']");
    if (!textarea) {
      throw new Error("Нативное поле чата не создано.");
    }

    textarea.setSelectionRange(4, 4);
    pressInputKey(textarea, "ArrowLeft", "ArrowLeft");
    textarea.setSelectionRange(3, 3);
    await waitForNativeEditSync();
    expect(controller.getChatCursorIndex()).toBe(3);

    textarea.setSelectionRange(3, 3);
    pressInputKey(textarea, "Home", "Home");
    textarea.setSelectionRange(0, 0);
    await waitForNativeEditSync();
    expect(controller.getChatCursorIndex()).toBe(0);

    textarea.setSelectionRange(0, 0);
    pressInputKey(textarea, "End", "End");
    textarea.setSelectionRange(4, 4);
    await waitForNativeEditSync();
    expect(controller.getChatCursorIndex()).toBe(4);
  });

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
  // Проверяет, что длинные перенесенные строки увеличивают максимальную прокрутку до фактического верха истории.
  it("uses measured wrapped chat content height for maximum scroll", () => {
    setViewportWidth(1000);
    setViewportHeight(1000);
    addChatMessagesDom(620);
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
    for (let index = 0; index < 20; index += 1) {
      wheel(-100, false);
    }
    controller.getVisibleChatState({
      type: "chatState",
      selectedChatId: 1,
      tabs: [{ chatId: 1, title: "Server", communityTypeAcronym: "Server", duoChatKey: "", messages: messages(20) }],
    });

    expect(controller.getChatScrollState().contentOffsetPx).toBe(396);
    expect(controller.getChatScrollState().thumbTopPercent).toBe(0);
  });

  // Проверяет, что левая кромка видимой полосы участвует в захвате.
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
