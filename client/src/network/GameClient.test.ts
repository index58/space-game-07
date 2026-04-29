import { describe, expect, it, vi } from "vitest";
import { GameClient } from "./GameClient";
import type { ClientInputState } from "./protocol";

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
