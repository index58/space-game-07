import type {
  AuthMessage,
  ClientInputMessage,
  ClientInputState,
  ConnectionStatus,
  RandomShipMessage,
  SnapshotMessage,
} from "./protocol";

// Позволяет тестам подменять браузерный WebSocket простой заглушкой.
type WebSocketLike = {
  // Обработчик успешного открытия соединения.
  onopen: ((event?: unknown) => void) | null;
  // Обработчик закрытия соединения.
  onclose: ((event?: unknown) => void) | null;
  // Обработчик сетевой ошибки.
  onerror: ((event?: unknown) => void) | null;
  // Обработчик входящего текстового сообщения.
  onmessage: ((event: { data: string }) => void) | null;
  // Отправляет текстовый пакет на сервер.
  send(payload: string): void;
  // Закрывает соединение со стороны клиента.
  close(): void;
};

// Настраивает адрес сервера, учетную запись и тайминги сетевого клиента.
export type GameClientOptions = {
  // Полный WebSocket-адрес сервера.
  url?: string;
  // Секрет авторизации аккаунта.
  token?: string;
  // Задержка перед повторным подключением.
  reconnectDelayMs?: number;
  // Период отправки ввода на сервер.
  inputIntervalMs?: number;
  // Фабрика соединений для тестов и нестандартных окружений.
  socketFactory?: (url: string) => WebSocketLike;
};

const DEFAULT_URL = "ws://127.0.0.1:8080/ws";
const DEFAULT_RECONNECT_DELAY_MS = 1000;
const DEFAULT_INPUT_INTERVAL_MS = 1000 / 30;

// Возвращает нейтральное управление без тяги и поворота.
const emptyInput = (): ClientInputState => ({
  thrustForward: false,
  thrustBackward: false,
  thrustLeft: false,
  thrustRight: false,
  targetRotationDelta: 0,
});

// Достает значение cookie без зависимости от браузерных API в тестовой среде.
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

// Ищет токен сначала в localStorage, затем в cookie старого формата.
const readStoredToken = (): string | null => {
  if (typeof localStorage !== "undefined") {
    const token = localStorage.getItem("accountToken");
    if (token) {
      return token;
    }
  }

  return readCookie("Token");
};

// Добавляет query-параметр к URL, сохраняя уже существующую строку запроса.
const withQueryParameter = (url: string, name: string, value: string | null): string => {
  if (!value) {
    return url;
  }

  const separator = url.includes("?") ? "&" : "?";
  return `${url}${separator}${name}=${encodeURIComponent(value)}`;
};

// Передает серверу сохраненный секрет или оставляет запрос гостевым.
const withAccountIdentity = (url: string, token: string | null): string => {
  return withQueryParameter(url, "token", token);
};

// Проверяет минимальный контракт снимка перед тем, как сцена начнет его рисовать.
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

// Проверяет пакет с секретом перед записью в браузерное хранилище.
const isAuthMessage = (message: unknown): message is AuthMessage => {
  if (!message || typeof message !== "object") {
    return false;
  }

  const auth = message as AuthMessage;

  return auth.type === "auth" && typeof auth.token === "string" && auth.token.length > 0;
};

// Сохраняет секрет, если окружение предоставляет браузерное хранилище.
const storeAccountToken = (token: string): void => {
  if (typeof localStorage === "undefined") {
    return;
  }

  localStorage.setItem("accountToken", token);
};

// Держит WebSocket-соединение, переподключение и периодическую отправку ввода.
export class GameClient {
  // Адрес подключения с уже добавленными параметрами авторизации.
  private readonly url: string;
  // Задержка перед следующей попыткой соединения.
  private readonly reconnectDelayMs: number;
  // Период отправки последнего ввода на сервер.
  private readonly inputIntervalMs: number;
  // Фабрика сокета, позволяющая подменять транспорт в тестах.
  private readonly socketFactory: (url: string) => WebSocketLike;
  // Текущее активное соединение, если оно открыто или открывается.
  private socket: WebSocketLike | null = null;
  // Состояние соединения для сцены и отладочного слоя.
  private status: ConnectionStatus = "connecting";
  // Последний валидный снимок мира от сервера.
  private latestSnapshot: SnapshotMessage | null = null;
  // Последнее состояние управления, готовое к отправке.
  private latestInput: ClientInputState = emptyInput();
  // Последний выданный порядковый номер пакета ввода.
  private seq = 0;
  // Таймер отложенного переподключения.
  private reconnectTimer: ReturnType<typeof setTimeout> | null = null;
  // Таймер периодической отправки ввода.
  private inputTimer: ReturnType<typeof setInterval> | null = null;
  // Флаг окончательного уничтожения клиента.
  private destroyed = false;

  // Конструктор сразу начинает подключение, потому что сцена ожидает живой поток снимков.
  constructor(options: GameClientOptions = {}) {
    this.url = withAccountIdentity(
      options.url ?? DEFAULT_URL,
      options.token ?? readStoredToken(),
    );
    this.reconnectDelayMs = options.reconnectDelayMs ?? DEFAULT_RECONNECT_DELAY_MS;
    this.inputIntervalMs = options.inputIntervalMs ?? DEFAULT_INPUT_INTERVAL_MS;
    this.socketFactory = options.socketFactory ?? ((url) => new WebSocket(url) as unknown as WebSocketLike);

    this.connect();
  }

  // Нужен сцене для выбора между ожиданием и отрисовкой мира.
  getStatus(): ConnectionStatus {
    return this.status;
  }

  // Возвращает последний валидный снимок без копирования массивов объектов.
  getLatestSnapshot(): SnapshotMessage | null {
    return this.latestSnapshot;
  }

  // Обновляет состояние клавиш и накапливает относительный поворот мыши до отправки.
  setInput(input: ClientInputState): void {
    this.latestInput = {
      ...input,
      // Угловая дельта мыши приходит каждый кадр рендера, поэтому копим ее до сетевой отправки.
      targetRotationDelta: this.latestInput.targetRotationDelta + input.targetRotationDelta,
    };
  }

  // Отправляет одноразовую команду смены модели, не смешивая ее с потоковым управлением.
  requestRandomShipChange(): void {
    if (this.status !== "connected" || !this.socket) {
      return;
    }

    const message: RandomShipMessage = {
      type: "randomShip",
    };

    this.socket.send(JSON.stringify(message));
  }

  // Останавливает таймеры и закрывает сокет при уничтожении Phaser-сцены или теста.
  destroy(): void {
    this.destroyed = true;
    this.clearReconnectTimer();
    this.clearInputTimer();

    const socket = this.socket;
    this.socket = null;
    socket?.close();
  }

  // Создает новый WebSocket и привязывает обработчики к конкретному экземпляру сокета.
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

  // Принимает только валидные снимки мира, остальные сообщения игнорируются.
  private handleMessage(data: string): void {
    let parsed: unknown;

    try {
      parsed = JSON.parse(data);
    } catch {
      return;
    }

    if (isAuthMessage(parsed)) {
      storeAccountToken(parsed.token);
      return;
    }

    if (isSnapshotMessage(parsed)) {
      this.latestSnapshot = parsed;
    }
  }

  // Переводит клиента в ожидание и запускает отложенное переподключение.
  private handleDisconnected(socket: WebSocketLike): void {
    if (this.destroyed || this.socket !== socket) {
      return;
    }

    this.status = "waiting";
    this.socket = null;
    this.clearInputTimer();
    this.scheduleReconnect();
  }

  // Планирует ровно одну попытку переподключения.
  private scheduleReconnect(): void {
    this.clearReconnectTimer();
    this.reconnectTimer = setTimeout(() => {
      this.reconnectTimer = null;
      this.connect();
    }, this.reconnectDelayMs);
  }

  // Синхронизирует отправку ввода с серверной частотой симуляции.
  private startInputTimer(): void {
    this.clearInputTimer();
    this.inputTimer = setInterval(() => {
      this.sendInput();
    }, this.inputIntervalMs);
  }

  // Сериализует последний ввод и очищает только накопленную дельту мыши.
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

  // Убирает отложенное переподключение при новом состоянии клиента.
  private clearReconnectTimer(): void {
    if (this.reconnectTimer) {
      clearTimeout(this.reconnectTimer);
      this.reconnectTimer = null;
    }
  }

  // Останавливает отправку ввода, когда сокет уже не подключен.
  private clearInputTimer(): void {
    if (this.inputTimer) {
      clearInterval(this.inputTimer);
      this.inputTimer = null;
    }
  }
}
