import type {
  ClientInputMessage,
  ClientInputState,
  ConnectionStatus,
  SnapshotMessage,
} from "./protocol";

type WebSocketLike = {
  onopen: ((event?: unknown) => void) | null;
  onclose: ((event?: unknown) => void) | null;
  onerror: ((event?: unknown) => void) | null;
  onmessage: ((event: { data: string }) => void) | null;
  send(payload: string): void;
  close(): void;
};

export type GameClientOptions = {
  url?: string;
  token?: string;
  accountNickname?: string;
  reconnectDelayMs?: number;
  inputIntervalMs?: number;
  socketFactory?: (url: string) => WebSocketLike;
};

const DEFAULT_URL = "ws://127.0.0.1:8080/ws";
const DEFAULT_RECONNECT_DELAY_MS = 1000;
const DEFAULT_INPUT_INTERVAL_MS = 1000 / 30;

const emptyInput = (): ClientInputState => ({
  thrustForward: false,
  thrustBackward: false,
  thrustLeft: false,
  thrustRight: false,
  targetRotationDelta: 0,
});

const readCookie = (name: string): string | null => {
  if (typeof document === "undefined") {
    return null;
  }

  const prefix = `${name}=`;
  const cookie = document.cookie
    .split(";")
    .map((part) => part.trim())
    .find((part) => part.startsWith(prefix));

  return cookie ? decodeURIComponent(cookie.slice(prefix.length)) : null;
};

const readStoredToken = (): string | null => {
  if (typeof localStorage !== "undefined") {
    const token = localStorage.getItem("accountToken");
    if (token) {
      return token;
    }
  }

  return readCookie("Token");
};

const readStoredAccountNickname = (): string => {
  if (typeof localStorage !== "undefined") {
    const nickname = localStorage.getItem("accountNickname");
    if (nickname) {
      return nickname;
    }
  }

  return "index";
};

const withQueryParameter = (url: string, name: string, value: string | null): string => {
  if (!value) {
    return url;
  }

  const separator = url.includes("?") ? "&" : "?";
  return `${url}${separator}${name}=${encodeURIComponent(value)}`;
};

const withAccountIdentity = (url: string, token: string | null, accountNickname: string): string => {
  if (token) {
    return withQueryParameter(url, "token", token);
  }

  return withQueryParameter(url, "nickname", accountNickname);
};

const isSnapshotMessage = (message: unknown): message is SnapshotMessage => {
  if (!message || typeof message !== "object") {
    return false;
  }

  const snapshot = message as SnapshotMessage;

  return snapshot.type === "snapshot" &&
    typeof snapshot.tick === "number" &&
    typeof snapshot.selfObjectId === "number" &&
    Array.isArray(snapshot.objects);
};

export class GameClient {
  private readonly url: string;
  private readonly reconnectDelayMs: number;
  private readonly inputIntervalMs: number;
  private readonly socketFactory: (url: string) => WebSocketLike;
  private socket: WebSocketLike | null = null;
  private status: ConnectionStatus = "connecting";
  private latestSnapshot: SnapshotMessage | null = null;
  private latestInput: ClientInputState = emptyInput();
  private seq = 0;
  private reconnectTimer: ReturnType<typeof setTimeout> | null = null;
  private inputTimer: ReturnType<typeof setInterval> | null = null;
  private destroyed = false;

  constructor(options: GameClientOptions = {}) {
    this.url = withAccountIdentity(
      options.url ?? DEFAULT_URL,
      options.token ?? readStoredToken(),
      options.accountNickname ?? readStoredAccountNickname(),
    );
    this.reconnectDelayMs = options.reconnectDelayMs ?? DEFAULT_RECONNECT_DELAY_MS;
    this.inputIntervalMs = options.inputIntervalMs ?? DEFAULT_INPUT_INTERVAL_MS;
    this.socketFactory = options.socketFactory ?? ((url) => new WebSocket(url) as unknown as WebSocketLike);

    this.connect();
  }

  getStatus(): ConnectionStatus {
    return this.status;
  }

  getLatestSnapshot(): SnapshotMessage | null {
    return this.latestSnapshot;
  }

  setInput(input: ClientInputState): void {
    this.latestInput = {
      ...input,
      // Угловая дельта мыши приходит каждый кадр рендера, поэтому копим ее до сетевой отправки.
      targetRotationDelta: this.latestInput.targetRotationDelta + input.targetRotationDelta,
    };
  }

  destroy(): void {
    this.destroyed = true;
    this.clearReconnectTimer();
    this.clearInputTimer();

    const socket = this.socket;
    this.socket = null;
    socket?.close();
  }

  private connect(): void {
    if (this.destroyed) {
      return;
    }

    this.status = "connecting";
    const socket = this.socketFactory(this.url);
    this.socket = socket;

    socket.onopen = () => {
      if (this.destroyed || this.socket !== socket) {
        return;
      }

      this.status = "connected";
      this.startInputTimer();
    };

    socket.onmessage = (event) => {
      this.handleMessage(event.data);
    };

    socket.onerror = () => {
      this.handleDisconnected(socket);
    };

    socket.onclose = () => {
      this.handleDisconnected(socket);
    };
  }

  private handleMessage(data: string): void {
    let parsed: unknown;

    try {
      parsed = JSON.parse(data);
    } catch {
      return;
    }

    if (isSnapshotMessage(parsed)) {
      this.latestSnapshot = parsed;
    }
  }

  private handleDisconnected(socket: WebSocketLike): void {
    if (this.destroyed || this.socket !== socket) {
      return;
    }

    this.status = "waiting";
    this.socket = null;
    this.clearInputTimer();
    this.scheduleReconnect();
  }

  private scheduleReconnect(): void {
    this.clearReconnectTimer();
    this.reconnectTimer = setTimeout(() => {
      this.reconnectTimer = null;
      this.connect();
    }, this.reconnectDelayMs);
  }

  private startInputTimer(): void {
    this.clearInputTimer();
    this.inputTimer = setInterval(() => {
      this.sendInput();
    }, this.inputIntervalMs);
  }

  private sendInput(): void {
    if (this.status !== "connected" || !this.socket) {
      return;
    }

    const message: ClientInputMessage = {
      type: "input",
      seq: ++this.seq,
      ...this.latestInput,
    };

    this.socket.send(JSON.stringify(message));

    // После успешной отправки оставляем последние состояния клавиш, но начинаем новую угловую дельту.
    this.latestInput = {
      ...this.latestInput,
      targetRotationDelta: 0,
    };
  }

  private clearReconnectTimer(): void {
    if (this.reconnectTimer) {
      clearTimeout(this.reconnectTimer);
      this.reconnectTimer = null;
    }
  }

  private clearInputTimer(): void {
    if (this.inputTimer) {
      clearInterval(this.inputTimer);
      this.inputTimer = null;
    }
  }
}
