import type {
  ClientInputMessage,
  ClientInputState,
  ConnectionStatus,
  SnapshotMessage,
} from "./protocol";

// позволяет тестам подменять браузерный WebSocket простой заглушкой.
type WebSocketLike = {
  onopen: ((event?: unknown) => void) | null;
  onclose: ((event?: unknown) => void) | null;
  onerror: ((event?: unknown) => void) | null;
  onmessage: ((event: { data: string }) => void) | null;
  send(payload: string): void;
  close(): void;
};

// настраивает адрес сервера, учетную запись и тайминги сетевого клиента.
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

// возвращает нейтральное управление без тяги и поворота.
const emptyInput = (): ClientInputState => ({
  thrustForward: false,
  thrustBackward: false,
  thrustLeft: false,
  thrustRight: false,
  targetRotationDelta: 0,
});

// достает значение cookie без зависимости от браузерных API в тестовой среде.
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

// ищет токен сначала в localStorage, затем в cookie старого формата.
const readStoredToken = (): string | null => {
  if (typeof localStorage !== "undefined") {
    const token = localStorage.getItem("accountToken");
    if (token) {
      return token;
    }
  }

  return readCookie("Token");
};

// дает локальный никнейм по умолчанию, чтобы прототип подключался без формы входа.
const readStoredAccountNickname = (): string => {
  if (typeof localStorage !== "undefined") {
    const nickname = localStorage.getItem("accountNickname");
    if (nickname) {
      return nickname;
    }
  }

  return "index";
};

// добавляет query-параметр к URL, сохраняя уже существующую строку запроса.
const withQueryParameter = (url: string, name: string, value: string | null): string => {
  if (!value) {
    return url;
  }

  const separator = url.includes("?") ? "&" : "?";
  return `${url}${separator}${name}=${encodeURIComponent(value)}`;
};

// передает серверу токен или запасной никнейм локального аккаунта.
const withAccountIdentity = (url: string, token: string | null, accountNickname: string): string => {
  if (token) {
    return withQueryParameter(url, "token", token);
  }

  return withQueryParameter(url, "nickname", accountNickname);
};

// проверяет минимальный контракт снимка перед тем, как сцена начнет его рисовать.
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

// держит WebSocket-соединение, переподключение и периодическую отправку ввода.
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

  // Конструктор сразу начинает подключение, потому что сцена ожидает живой поток снимков.
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

  // нужен сцене для выбора между ожиданием и отрисовкой мира.
  getStatus(): ConnectionStatus {
    return this.status;
  }

  // возвращает последний валидный снимок без копирования массивов объектов.
  getLatestSnapshot(): SnapshotMessage | null {
    return this.latestSnapshot;
  }

  // обновляет состояние клавиш и накапливает относительный поворот мыши до отправки.
  setInput(input: ClientInputState): void {
    this.latestInput = {
      ...input,
      // Угловая дельта мыши приходит каждый кадр рендера, поэтому копим ее до сетевой отправки.
      targetRotationDelta: this.latestInput.targetRotationDelta + input.targetRotationDelta,
    };
  }

  // останавливает таймеры и закрывает сокет при уничтожении Phaser-сцены или теста.
  destroy(): void {
    this.destroyed = true;
    this.clearReconnectTimer();
    this.clearInputTimer();

    const socket = this.socket;
    this.socket = null;
    socket?.close();
  }

  // создает новый WebSocket и привязывает обработчики к конкретному экземпляру сокета.
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

  // принимает только валидные снимки мира, остальные сообщения игнорируются.
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

  // переводит клиента в ожидание и запускает отложенное переподключение.
  private handleDisconnected(socket: WebSocketLike): void {
    if (this.destroyed || this.socket !== socket) {
      return;
    }

    this.status = "waiting";
    this.socket = null;
    this.clearInputTimer();
    this.scheduleReconnect();
  }

  // планирует ровно одну попытку переподключения.
  private scheduleReconnect(): void {
    this.clearReconnectTimer();
    this.reconnectTimer = setTimeout(() => {
      this.reconnectTimer = null;
      this.connect();
    }, this.reconnectDelayMs);
  }

  // синхронизирует отправку ввода с серверной частотой симуляции.
  private startInputTimer(): void {
    this.clearInputTimer();
    this.inputTimer = setInterval(() => {
      this.sendInput();
    }, this.inputIntervalMs);
  }

  // сериализует последний ввод и очищает только накопленную дельту мыши.
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

  // убирает отложенное переподключение при новом состоянии клиента.
  private clearReconnectTimer(): void {
    if (this.reconnectTimer) {
      clearTimeout(this.reconnectTimer);
      this.reconnectTimer = null;
    }
  }

  // останавливает отправку ввода, когда сокет уже не подключен.
  private clearInputTimer(): void {
    if (this.inputTimer) {
      clearInterval(this.inputTimer);
      this.inputTimer = null;
    }
  }
}
