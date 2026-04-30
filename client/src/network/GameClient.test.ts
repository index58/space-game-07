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
  targetRotationDelta: 0,
});

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

  it("passes account nickname when token is empty", () => {
    FakeWebSocket.instances = [];
    const client = new GameClient({
      accountNickname: "index",
      token: "",
      socketFactory: (url) => new FakeWebSocket(url),
      reconnectDelayMs: 1000,
      inputIntervalMs: 1000,
    });

    expect(FakeWebSocket.instances[0].url).toBe("ws://127.0.0.1:8080/ws?nickname=index");

    client.destroy();
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
      targetRotationDelta: 0.25,
    });

    client.destroy();
    vi.useRealTimers();
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
});
