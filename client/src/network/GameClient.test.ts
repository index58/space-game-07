import { describe, expect, it, vi } from "vitest";
import { GameClient } from "./GameClient";
import type { ClientInputState } from "./protocol";

// Имитирует браузерный сокет и дает тестам вручную дергать события соединения.
class FakeWebSocket {
  static instances: FakeWebSocket[] = [];

  onopen: (() => void) | null = null;
  onclose: (() => void) | null = null;
  onerror: (() => void) | null = null;
  onmessage: ((event: { data: string }) => void) | null = null;
  readonly sent: string[] = [];
  closed = false;

  constructor(readonly url: string) {
    FakeWebSocket.instances.push(this);
  }

  send(payload: string): void {
    this.sent.push(payload);
  }

  close(): void {
    this.closed = true;
    this.onclose?.();
  }
}

// Собирает нейтральный ввод, чтобы каждый тест явно менял только проверяемые поля.
const emptyInput = (): ClientInputState => ({
  thrustForward: false,
  thrustBackward: false,
  thrustLeft: false,
  thrustRight: false,
  toggleAnchor: false,
  targetRotationDelta: 0,
});

// Подменяет браузерное хранилище в среде Vitest.
const installLocalStorage = (): Storage => {
  const values = new Map<string, string>();
  const storage = {
    get length() {
      return values.size;
    },
    clear: () => values.clear(),
    getItem: (key: string) => values.get(key) ?? null,
    key: (index: number) => Array.from(values.keys())[index] ?? null,
    removeItem: (key: string) => values.delete(key),
    setItem: (key: string, value: string) => values.set(key, value),
  } as Storage;

  vi.stubGlobal("localStorage", storage);
  return storage;
};

describe("GameClient", () => {
  it("updates latest snapshot from a valid snapshot message", () => {
    FakeWebSocket.instances = [];
    const client = new GameClient({
      socketFactory: (url) => new FakeWebSocket(url),
      reconnectDelayMs: 1000,
      inputIntervalMs: 1000,
    });
    const socket = FakeWebSocket.instances[0];

    socket.onopen?.();
    socket.onmessage?.({
      data: JSON.stringify({
        type: "snapshot",
        tick: 1,
        selfObjectId: 7,
        objects: [],
      }),
    });

    expect(client.getStatus()).toBe("connected");
    expect(client.getLatestSnapshot()?.selfObjectId).toBe(7);

    client.destroy();
  });

  it("passes account token in websocket URL", () => {
    FakeWebSocket.instances = [];
    const client = new GameClient({
      token: "test token",
      socketFactory: (url) => new FakeWebSocket(url),
      reconnectDelayMs: 1000,
      inputIntervalMs: 1000,
    });

    expect(FakeWebSocket.instances[0].url).toBe("ws://127.0.0.1:8080/ws?token=test%20token");

    client.destroy();
  });

  it("connects without identity when token is empty", () => {
    FakeWebSocket.instances = [];
    const client = new GameClient({
      token: "",
      socketFactory: (url) => new FakeWebSocket(url),
      reconnectDelayMs: 1000,
      inputIntervalMs: 1000,
    });

    expect(FakeWebSocket.instances[0].url).toBe("ws://127.0.0.1:8080/ws");

    client.destroy();
  });

  it("stores account token from auth message", () => {
    const storage = installLocalStorage();
    FakeWebSocket.instances = [];
    const client = new GameClient({
      token: "",
      socketFactory: (url) => new FakeWebSocket(url),
      reconnectDelayMs: 1000,
      inputIntervalMs: 1000,
    });
    const socket = FakeWebSocket.instances[0];

    socket.onmessage?.({
      data: JSON.stringify({
        type: "auth",
        token: "new-token",
      }),
    });

    expect(storage.getItem("accountToken")).toBe("new-token");

    client.destroy();
    vi.unstubAllGlobals();
  });

  it("appends account token to websocket URL with existing query", () => {
    FakeWebSocket.instances = [];
    const client = new GameClient({
      url: "ws://127.0.0.1:8080/ws?debug=1",
      token: "token",
      socketFactory: (url) => new FakeWebSocket(url),
      reconnectDelayMs: 1000,
      inputIntervalMs: 1000,
    });

    expect(FakeWebSocket.instances[0].url).toBe("ws://127.0.0.1:8080/ws?debug=1&token=token");

    client.destroy();
  });

  it("does not replace latest snapshot when JSON is invalid", () => {
    FakeWebSocket.instances = [];
    const client = new GameClient({
      socketFactory: (url) => new FakeWebSocket(url),
      reconnectDelayMs: 1000,
      inputIntervalMs: 1000,
    });
    const socket = FakeWebSocket.instances[0];

    socket.onmessage?.({
      data: JSON.stringify({
        type: "snapshot",
        tick: 1,
        selfObjectId: 7,
        objects: [],
      }),
    });
    socket.onmessage?.({ data: "not-json" });

    expect(client.getLatestSnapshot()?.selfObjectId).toBe(7);

    client.destroy();
  });

  it("sends input message with agreed field names while connected", () => {
    vi.useFakeTimers();
    FakeWebSocket.instances = [];
    const client = new GameClient({
      socketFactory: (url) => new FakeWebSocket(url),
      reconnectDelayMs: 1000,
      inputIntervalMs: 1000,
    });
    const socket = FakeWebSocket.instances[0];

    client.setInput({ ...emptyInput(), thrustForward: true, targetRotationDelta: 0.25 });
    socket.onopen?.();
    vi.advanceTimersByTime(1000);

    expect(JSON.parse(socket.sent[0])).toEqual({
      type: "input",
      seq: 1,
      thrustForward: true,
      thrustBackward: false,
      thrustLeft: false,
      thrustRight: false,
      toggleAnchor: false,
      targetRotationDelta: 0.25,
    });

    client.destroy();
    vi.useRealTimers();
  });

  it("sends random ship command while connected", () => {
    FakeWebSocket.instances = [];
    const client = new GameClient({
      socketFactory: (url) => new FakeWebSocket(url),
      reconnectDelayMs: 1000,
      inputIntervalMs: 1000,
    });
    const socket = FakeWebSocket.instances[0];

    socket.onopen?.();
    client.requestRandomShipChange();

    expect(JSON.parse(socket.sent[0])).toEqual({
      type: "randomShip",
    });

    client.destroy();
  });

  it("keeps docking events in arrival order until scene consumes them", () => {
    FakeWebSocket.instances = [];
    const client = new GameClient({
      socketFactory: (url) => new FakeWebSocket(url),
      reconnectDelayMs: 1000,
      inputIntervalMs: 1000,
    });
    const socket = FakeWebSocket.instances[0];

    socket.onmessage?.({ data: JSON.stringify({ type: "dockingEvent", kind: "dockingFinished" }) });
    socket.onmessage?.({ data: JSON.stringify({ type: "dockingEvent", kind: "dockingNotification", message: "Объект пристыкован" }) });

    expect(client.consumeDockingEvents().map((event) => event.kind)).toEqual(["dockingFinished", "dockingNotification"]);
    expect(client.consumeDockingEvents()).toEqual([]);

    client.destroy();
  });

  it("accumulates target rotation delta between input sends", () => {
    vi.useFakeTimers();
    FakeWebSocket.instances = [];
    const client = new GameClient({
      socketFactory: (url) => new FakeWebSocket(url),
      reconnectDelayMs: 1000,
      inputIntervalMs: 1000,
    });
    const socket = FakeWebSocket.instances[0];

    socket.onopen?.();
    client.setInput({ ...emptyInput(), thrustForward: true, targetRotationDelta: 0.1 });
    client.setInput({ ...emptyInput(), thrustRight: true, targetRotationDelta: 0.2 });
    client.setInput({ ...emptyInput(), thrustRight: true, targetRotationDelta: -0.05 });
    vi.advanceTimersByTime(1000);
    vi.advanceTimersByTime(1000);

    const firstMessage = JSON.parse(socket.sent[0]) as ClientInputState & { type: "input"; seq: number };
    expect(firstMessage).toEqual({
      type: "input",
      seq: 1,
      thrustForward: false,
      thrustBackward: false,
      thrustLeft: false,
      thrustRight: true,
      toggleAnchor: false,
      targetRotationDelta: firstMessage.targetRotationDelta,
    });
    expect(firstMessage.targetRotationDelta).toBeCloseTo(0.25);

    expect(JSON.parse(socket.sent[1])).toEqual({
      type: "input",
      seq: 2,
      thrustForward: false,
      thrustBackward: false,
      thrustLeft: false,
      thrustRight: true,
      toggleAnchor: false,
      targetRotationDelta: 0,
    });

    client.destroy();
    vi.useRealTimers();
  });

  it("moves to waiting status after close", () => {
    vi.useFakeTimers();
    FakeWebSocket.instances = [];
    const client = new GameClient({
      socketFactory: (url) => new FakeWebSocket(url),
      reconnectDelayMs: 1000,
      inputIntervalMs: 1000,
    });
    const socket = FakeWebSocket.instances[0];

    socket.onopen?.();
    socket.onclose?.();

    expect(client.getStatus()).toBe("waiting");

    client.destroy();
    vi.useRealTimers();
  });

  it("sends anchor toggle only once after P input", () => {
    vi.useFakeTimers();
    FakeWebSocket.instances = [];
    const client = new GameClient({
      socketFactory: (url) => new FakeWebSocket(url),
      reconnectDelayMs: 1000,
      inputIntervalMs: 1000,
    });
    const socket = FakeWebSocket.instances[0];

    socket.onopen?.();
    client.setInput({ ...emptyInput(), toggleAnchor: true });
    vi.advanceTimersByTime(1000);
    vi.advanceTimersByTime(1000);

    expect(JSON.parse(socket.sent[0])).toMatchObject({ toggleAnchor: true });
    expect(JSON.parse(socket.sent[1])).toMatchObject({ toggleAnchor: false });

    client.destroy();
    vi.useRealTimers();
  });

  it("updates latest chat state and clears previous chat error", () => {
    FakeWebSocket.instances = [];
    const client = new GameClient({
      socketFactory: (url) => new FakeWebSocket(url),
      reconnectDelayMs: 1000,
      inputIntervalMs: 1000,
    });
    const socket = FakeWebSocket.instances[0];

    socket.onmessage?.({ data: JSON.stringify({ type: "chatError", message: "Ошибка" }) });
    socket.onmessage?.({
      data: JSON.stringify({
        type: "chatState",
        selectedChatId: 1,
        tabs: [{ chatId: 1, title: "Server", communityTypeAcronym: "Server", duoChatKey: "", messages: [] }],
      }),
    });

    expect(client.getLatestChatState()?.selectedChatId).toBe(1);
    expect(client.getLatestChatError()).toBeNull();

    client.destroy();
  });

  it("stores latest chat error from server refusal", () => {
    FakeWebSocket.instances = [];
    const client = new GameClient({
      socketFactory: (url) => new FakeWebSocket(url),
      reconnectDelayMs: 1000,
      inputIntervalMs: 1000,
    });
    const socket = FakeWebSocket.instances[0];

    socket.onmessage?.({ data: JSON.stringify({ type: "chatError", message: "Адресат не найден" }) });

    expect(client.getLatestChatError()).toBe("Адресат не найден");
    expect(client.getLatestChatErrorSeq()).toBe(1);

    client.destroy();
  });

  it("increments chat error sequence for repeated same refusal", () => {
    FakeWebSocket.instances = [];
    const client = new GameClient({
      socketFactory: (url) => new FakeWebSocket(url),
      reconnectDelayMs: 1000,
      inputIntervalMs: 1000,
    });
    const socket = FakeWebSocket.instances[0];

    socket.onmessage?.({ data: JSON.stringify({ type: "chatError", message: "Адресат не найден" }) });
    socket.onmessage?.({ data: JSON.stringify({ type: "chatError", message: "Адресат не найден" }) });

    expect(client.getLatestChatError()).toBe("Адресат не найден");
    expect(client.getLatestChatErrorSeq()).toBe(2);

    client.destroy();
  });

  it("sends chat command while connected", () => {
    FakeWebSocket.instances = [];
    const client = new GameClient({
      socketFactory: (url) => new FakeWebSocket(url),
      reconnectDelayMs: 1000,
      inputIntervalMs: 1000,
    });
    const socket = FakeWebSocket.instances[0];

    socket.onopen?.();
    client.sendChatMessage({ chatId: 3, text: "hello" });
    client.sendChatMessage({ targetNickname: "Pilot2", text: "private" });

    expect(JSON.parse(socket.sent[0])).toEqual({ type: "chatSend", chatId: 3, text: "hello" });
    expect(JSON.parse(socket.sent[1])).toEqual({ type: "chatSend", targetNickname: "Pilot2", text: "private" });

    client.destroy();
  });

  it("stores and sends input settings messages", () => {
    FakeWebSocket.instances = [];
    const client = new GameClient({
      socketFactory: (url) => new FakeWebSocket(url),
      reconnectDelayMs: 1000,
      inputIntervalMs: 1000,
    });
    const socket = FakeWebSocket.instances[0];

    socket.onmessage?.({ data: JSON.stringify({ type: "inputSettings", settings: [{ actionTypeId: 1, inputEventTypeId: 2 }] }) });
    socket.onopen?.();
    client.saveInputSettings([{ actionTypeId: 1, inputEventTypeId: 3 }]);

    expect(client.getLatestInputSettings()).toEqual([{ actionTypeId: 1, inputEventTypeId: 2 }]);
    expect(client.getLatestInputSettingsSeq()).toBe(1);
    expect(JSON.parse(socket.sent[0])).toEqual({
      type: "inputSettingsSave",
      settings: [{ actionTypeId: 1, inputEventTypeId: 3 }],
    });

    client.destroy();
  });

  // Проверяет, что команды панели управления получают общий ID сессии и возрастающие номера мутаций.
  it("sends control panel mutations with shared session and increasing sequence", () => {
    FakeWebSocket.instances = [];
    const client = new GameClient({
      socketFactory: (url) => new FakeWebSocket(url),
      reconnectDelayMs: 1000,
      inputIntervalMs: 1000,
      clientSessionId: "session-1",
    });
    const socket = FakeWebSocket.instances[0];

    socket.onopen?.();
    const objectMutation = client.sendControlPanelObjectUpdate({ enabled: false });
    const equipmentMutation = client.sendControlPanelEquipmentUpdate({ equipmentGroupId: 12, enabledCount: 3, title: "Renamed equipment" });
    const containerMutation = client.sendControlPanelContainerTransfer({ sourceContainerEquipmentGroupId: 21, targetContainerEquipmentGroupId: 22, itemGroupIds: [31], amount: 4 });
    const fuelMutation = client.sendControlPanelFuelTransfer({ containerEquipmentGroupId: 21, fuelTankEquipmentGroupId: 23, itemGroupIds: [32], amount: 12 });
    const constructorMutation = client.sendControlPanelConstructorProduceItem({ constructorEquipmentGroupId: 24, materialContainerEquipmentGroupId: 21, productContainerEquipmentGroupId: 22, schemaId: 41, amount: 3 });
    const blueprintMutation = client.sendControlPanelConstructorProduceItem({ constructorEquipmentGroupId: 24, materialContainerEquipmentGroupId: 21, blueprintId: 51, amount: 1 });
    const queueMutation = client.sendControlPanelConstructorQueueCommand({ constructorEquipmentGroupId: 24, jobId: 61, command: "cancelAll" });

    expect(objectMutation).toEqual({ sessionId: "session-1", seq: 1 });
    expect(equipmentMutation).toEqual({ sessionId: "session-1", seq: 2 });
    expect(containerMutation).toEqual({ sessionId: "session-1", seq: 3 });
    expect(fuelMutation).toEqual({ sessionId: "session-1", seq: 4 });
    expect(constructorMutation).toEqual({ sessionId: "session-1", seq: 5 });
    expect(blueprintMutation).toEqual({ sessionId: "session-1", seq: 6 });
    expect(queueMutation).toEqual({ sessionId: "session-1", seq: 7 });
    expect(JSON.parse(socket.sent[0])).toEqual({
      type: "controlPanelObjectUpdate",
      clientSessionId: "session-1",
      mutationSeq: 1,
      enabled: false,
    });
    expect(JSON.parse(socket.sent[1])).toEqual({
      type: "controlPanelEquipmentUpdate",
      clientSessionId: "session-1",
      mutationSeq: 2,
      equipmentGroupId: 12,
      enabledCount: 3,
      title: "Renamed equipment",
    });
    expect(JSON.parse(socket.sent[2])).toEqual({
      type: "controlPanelContainerTransfer",
      clientSessionId: "session-1",
      mutationSeq: 3,
      sourceContainerEquipmentGroupId: 21,
      targetContainerEquipmentGroupId: 22,
      itemGroupIds: [31],
      amount: 4,
    });
    expect(JSON.parse(socket.sent[3])).toEqual({
      type: "controlPanelFuelTransfer",
      clientSessionId: "session-1",
      mutationSeq: 4,
      containerEquipmentGroupId: 21,
      fuelTankEquipmentGroupId: 23,
      itemGroupIds: [32],
      amount: 12,
    });
    expect(JSON.parse(socket.sent[4])).toEqual({
      type: "controlPanelConstructorProduceItem",
      clientSessionId: "session-1",
      mutationSeq: 5,
      constructorEquipmentGroupId: 24,
      materialContainerEquipmentGroupId: 21,
      productContainerEquipmentGroupId: 22,
      schemaId: 41,
      amount: 3,
    });
    expect(JSON.parse(socket.sent[5])).toEqual({
      type: "controlPanelConstructorProduceItem",
      clientSessionId: "session-1",
      mutationSeq: 6,
      constructorEquipmentGroupId: 24,
      materialContainerEquipmentGroupId: 21,
      blueprintId: 51,
      amount: 1,
    });
    expect(JSON.parse(socket.sent[6])).toEqual({
      type: "controlPanelConstructorQueueCommand",
      clientSessionId: "session-1",
      mutationSeq: 7,
      constructorEquipmentGroupId: 24,
      jobId: 61,
      command: "cancelAll",
    });

    client.destroy();
  });

  // Проверяет, что отказ панели управления сохраняет номер отклоненной мутации.
  it("stores control panel mutation error from server refusal", () => {
    FakeWebSocket.instances = [];
    const client = new GameClient({
      socketFactory: (url) => new FakeWebSocket(url),
      reconnectDelayMs: 1000,
      inputIntervalMs: 1000,
    });
    const socket = FakeWebSocket.instances[0];

    socket.onmessage?.({ data: JSON.stringify({ type: "controlPanelError", clientSessionId: "session-1", mutationSeq: 4, message: "bad command" }) });

    expect(client.getLatestControlPanelError()).toEqual({
      type: "controlPanelError",
      clientSessionId: "session-1",
      mutationSeq: 4,
      message: "bad command",
    });
    expect(client.getLatestControlPanelErrorSeq()).toBe(1);

    client.destroy();
  });

  // Проверяет, что клиент умеет явно запросить свежие настройки ввода при открытии окна.
  it("requests latest input settings while connected", () => {
    FakeWebSocket.instances = [];
    const client = new GameClient({
      socketFactory: (url) => new FakeWebSocket(url),
      reconnectDelayMs: 1000,
      inputIntervalMs: 1000,
    });
    const socket = FakeWebSocket.instances[0];

    socket.onopen?.();
    client.requestInputSettings();

    expect(JSON.parse(socket.sent[0])).toEqual({ type: "inputSettingsRequest" });

    client.destroy();
  });

  it("stores latest input settings error from server refusal", () => {
    FakeWebSocket.instances = [];
    const client = new GameClient({
      socketFactory: (url) => new FakeWebSocket(url),
      reconnectDelayMs: 1000,
      inputIntervalMs: 1000,
    });
    const socket = FakeWebSocket.instances[0];

    socket.onmessage?.({ data: JSON.stringify({ type: "inputSettingsError", message: "bad settings" }) });

    expect(client.getLatestInputSettingsError()).toBe("bad settings");
    expect(client.getLatestInputSettingsErrorSeq()).toBe(1);

    client.destroy();
  });

  it("sends chat selection while connected", () => {
    FakeWebSocket.instances = [];
    const client = new GameClient({
      socketFactory: (url) => new FakeWebSocket(url),
      reconnectDelayMs: 1000,
      inputIntervalMs: 1000,
    });
    const socket = FakeWebSocket.instances[0];

    socket.onopen?.();
    client.selectChat(7);

    expect(JSON.parse(socket.sent[0])).toEqual({ type: "chatSelect", chatId: 7 });

    client.destroy();
  });
});
