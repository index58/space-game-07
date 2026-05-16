import type {
  AuthMessage,
  ChatErrorMessage,
  ChatSelectMessage,
  ChatSendMessage,
  ChatStateMessage,
  ClientInputMessage,
  ClientInputState,
  ConnectionStatus,
  ControlPanelConstructorQueueCommandMessage,
  ControlPanelConstructorProduceItemMessage,
  ControlPanelEquipmentGroupRelationUpdateMessage,
  ControlPanelEquipmentUpdateMessage,
  ControlPanelContainerTransferMessage,
  ControlPanelErrorMessage,
  ControlPanelFuelTransferMessage,
  ControlPanelMutationRef,
  DockingCommandMessage,
  DockingEventMessage,
  ControlPanelObjectUpdateMessage,
  InputSettingsErrorMessage,
  InputSettingsMessage,
  InputSettingsRequestMessage,
  InputSettingsSaveMessage,
  InputSettingPayload,
  RandomShipMessage,
  SnapshotMessage,
  TestClaimFocusedObjectOwnerMessage,
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
  // Идентификатор клиентской сессии для тестов и восстановления pending-состояния.
  clientSessionId?: string;
};

const DEFAULT_URL = "ws://127.0.0.1:8080/ws";
const DEFAULT_RECONNECT_DELAY_MS = 1000;
const DEFAULT_INPUT_INTERVAL_MS = 1000 / 30;

// Создает достаточно уникальный идентификатор браузерной сессии без зависимости от одного API.
const createClientSessionId = (): string => {
  if (typeof crypto !== "undefined" && typeof crypto.randomUUID === "function") {
    return crypto.randomUUID();
  }

  return `${Date.now()}-${Math.random().toString(36).slice(2)}`;
};

// Возвращает нейтральное управление без тяги и поворота.
const emptyInput = (): ClientInputState => ({
  thrustForward: false,
  thrustBackward: false,
  thrustLeft: false,
  thrustRight: false,
  toggleAnchor: false,
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

// Проверяет минимальный контракт состояния чата перед передачей в HUD.
const isChatStateMessage = (message: unknown): message is ChatStateMessage => {
  if (!message || typeof message !== "object") {
    return false;
  }

  const chatState = message as ChatStateMessage;

  return chatState.type === "chatState" &&
    Array.isArray(chatState.tabs) &&
    typeof chatState.selectedChatId === "number";
};

// Проверяет пакет ошибки перед отображением в панели.
const isChatErrorMessage = (message: unknown): message is ChatErrorMessage => {
  if (!message || typeof message !== "object") {
    return false;
  }

  const chatError = message as ChatErrorMessage;

  return chatError.type === "chatError" && typeof chatError.message === "string";
};

// Проверяет пакет настроек ввода перед передачей в окно настроек.
const isInputSettingsMessage = (message: unknown): message is InputSettingsMessage => {
  if (!message || typeof message !== "object") {
    return false;
  }

  const inputSettings = message as InputSettingsMessage;

  return inputSettings.type === "inputSettings" &&
    Array.isArray(inputSettings.settings) &&
    inputSettings.settings.every((setting) => typeof setting.actionTypeId === "number" && typeof setting.inputEventTypeId === "number");
};

// Проверяет пакет ошибки настроек перед отображением в окне.
const isInputSettingsErrorMessage = (message: unknown): message is InputSettingsErrorMessage => {
  if (!message || typeof message !== "object") {
    return false;
  }

  const error = message as InputSettingsErrorMessage;

  return error.type === "inputSettingsError" && typeof error.message === "string";
};

// Проверяет пакет отказа команды панели управления.
const isControlPanelErrorMessage = (message: unknown): message is ControlPanelErrorMessage => {
  if (!message || typeof message !== "object") {
    return false;
  }

  const error = message as ControlPanelErrorMessage;

  return error.type === "controlPanelError" &&
    typeof error.clientSessionId === "string" &&
    typeof error.mutationSeq === "number" &&
    typeof error.message === "string";
};

// Проверяет пакет стыковки перед передачей в игровой HUD.
const isDockingEventMessage = (message: unknown): message is DockingEventMessage => {
  if (!message || typeof message !== "object") {
    return false;
  }

  const event = message as DockingEventMessage;

  return event.type === "dockingEvent" &&
    typeof event.kind === "string" &&
    (event.role === undefined || event.role === "sender" || event.role === "receiver") &&
    (event.message === undefined || typeof event.message === "string") &&
    (event.duration === undefined || typeof event.duration === "number") &&
    (event.targetIds === undefined || Array.isArray(event.targetIds));
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
  // Сессия клиента для связывания pending-команд с серверным ack.
  private readonly clientSessionId: string;
  // Текущее активное соединение, если оно открыто или открывается.
  private socket: WebSocketLike | null = null;
  // Состояние соединения для сцены и отладочного слоя.
  private status: ConnectionStatus = "connecting";
  // Последний валидный снимок мира от сервера.
  private latestSnapshot: SnapshotMessage | null = null;
  // Последнее валидное состояние вкладок чата.
  private latestChatState: ChatStateMessage | null = null;
  // Последняя ошибка отправки текста.
  private latestChatError: string | null = null;
  // Порядковый номер последней ошибки, чтобы UI мог перезапустить одинаковую плашку.
  private latestChatErrorSeq = 0;
  // Последние сохраненные сервером привязки ввода текущего аккаунта.
  private latestInputSettings: InputSettingPayload[] = [];
  // Порядковый номер пакета настроек для синхронизации локального окна.
  private latestInputSettingsSeq = 0;
  // Последняя ошибка сохранения настроек ввода.
  private latestInputSettingsError: string | null = null;
  // Порядковый номер ошибки настроек для повторного показа одинакового текста.
  private latestInputSettingsErrorSeq = 0;
  // Последняя ошибка команды панели управления.
  private latestControlPanelError: ControlPanelErrorMessage | null = null;
  // Порядковый номер ошибки панели для повторной обработки одинакового текста.
  private latestControlPanelErrorSeq = 0;
  // Последнее событие стыковки для игрового HUD.
  private latestDockingEvent: DockingEventMessage | null = null;
  // Порядковый номер события стыковки для обработки одинаковых сообщений.
  private latestDockingEventSeq = 0;
  // Очередь событий стыковки, чтобы быстрые события не затирали друг друга.
  private dockingEvents: DockingEventMessage[] = [];
  // Последнее состояние управления, готовое к отправке.
  private latestInput: ClientInputState = emptyInput();
  // Последний выданный порядковый номер пакета ввода.
  private seq = 0;
  // Последний выданный порядковый номер команды панели управления.
  private mutationSeq = 0;
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
    this.clientSessionId = options.clientSessionId ?? createClientSessionId();

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

  // Возвращает последнюю историю и список вкладок без копирования сообщений.
  getLatestChatState(): ChatStateMessage | null {
    return this.latestChatState;
  }

  // Возвращает последнюю ошибку панели, если сервер отказал команде.
  getLatestChatError(): string | null {
    return this.latestChatError;
  }

  // Возвращает счетчик ошибок панели чата.
  getLatestChatErrorSeq(): number {
    return this.latestChatErrorSeq;
  }

  // Возвращает последние сохраненные сервером привязки ввода.
  getLatestInputSettings(): InputSettingPayload[] {
    return this.latestInputSettings;
  }

  // Возвращает счетчик обновлений привязок ввода.
  getLatestInputSettingsSeq(): number {
    return this.latestInputSettingsSeq;
  }

  // Возвращает последнюю ошибку сохранения настроек ввода.
  getLatestInputSettingsError(): string | null {
    return this.latestInputSettingsError;
  }

  // Возвращает счетчик ошибок сохранения настроек ввода.
  getLatestInputSettingsErrorSeq(): number {
    return this.latestInputSettingsErrorSeq;
  }

  // Возвращает последнюю ошибку команды панели управления.
  getLatestControlPanelError(): ControlPanelErrorMessage | null {
    return this.latestControlPanelError;
  }

  // Возвращает счетчик ошибок панели управления.
  getLatestControlPanelErrorSeq(): number {
    return this.latestControlPanelErrorSeq;
  }

  // Возвращает последнее событие стыковки от сервера.
  getLatestDockingEvent(): DockingEventMessage | null {
    return this.latestDockingEvent;
  }

  // Возвращает счетчик событий стыковки.
  getLatestDockingEventSeq(): number {
    return this.latestDockingEventSeq;
  }

  // Возвращает накопленные события стыковки в порядке прихода.
  consumeDockingEvents(): DockingEventMessage[] {
    const events = this.dockingEvents;
    this.dockingEvents = [];
    return events;
  }

  // Обновляет состояние клавиш и накапливает относительный поворот мыши до отправки.
  setInput(input: ClientInputState): void {
    this.latestInput = {
      ...input,
      // Переключение якоря является одноразовой командой и должно дожить до ближайшей отправки.
      toggleAnchor: this.latestInput.toggleAnchor || input.toggleAnchor,
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

  // Отправляет тестовую команду присвоения объекта в фокусе информационной панели.
  requestFocusedObjectOwnerClaim(): void {
    if (this.status !== "connected" || !this.socket) {
      return;
    }

    const message: TestClaimFocusedObjectOwnerMessage = {
      type: "testClaimFocusedObjectOwner",
    };

    this.socket.send(JSON.stringify(message));
  }

  // Отправляет запрос стыковки отдельной WebSocket-командой.
  sendDockingRequest(): void {
    this.sendDockingCommand("dockingRequest");
  }

  // Отправляет принятие входящего запроса отдельной WebSocket-командой.
  sendDockingApprove(): void {
    this.sendDockingCommand("dockingApprove");
  }

  // Отправляет отказ входящего запроса отдельной WebSocket-командой.
  sendDockingReject(): void {
    this.sendDockingCommand("dockingReject");
  }

  // Отправляет отстыковку текущего объекта отдельной WebSocket-командой.
  sendDockingUndock(): void {
    this.sendDockingCommand("dockingUndock");
  }

  // Отправляет начало пересадки персонажа отдельной WebSocket-командой.
  sendLandingBegin(): void {
    this.sendDockingCommand("landingBegin");
  }

  // Отправляет принятие входящего запроса посадки отдельной WebSocket-командой.
  sendLandingApprove(): void {
    this.sendDockingCommand("landingApprove");
  }

  // Отправляет отказ входящего запроса посадки отдельной WebSocket-командой.
  sendLandingReject(): void {
    this.sendDockingCommand("landingReject");
  }

  // Отправляет запрос посадки в выбранный объект назначения.
  sendLandingRequest(targetObjectId: number): void {
    if (this.status !== "connected" || !this.socket || targetObjectId <= 0) {
      return;
    }

    this.socket.send(JSON.stringify({ type: "landingRequest", targetObjectId }));
  }

  // Отправляет текстовую команду отдельно от потокового управления кораблем.
  sendChatMessage(message: Omit<ChatSendMessage, "type">): void {
    if (this.status !== "connected" || !this.socket) {
      return;
    }

    const payload: ChatSendMessage = {
      type: "chatSend",
      ...message,
    };

    this.socket.send(JSON.stringify(payload));
  }

  // Сообщает серверу, что игрок выбрал вкладку чата.
  selectChat(chatId: number): void {
    if (this.status !== "connected" || !this.socket) {
      return;
    }

    const payload: ChatSelectMessage = {
      type: "chatSelect",
      chatId,
    };

    this.socket.send(JSON.stringify(payload));
  }

  // Отправляет полный набор выбранных привязок ввода текущего аккаунта.
  saveInputSettings(settings: InputSettingPayload[]): void {
    if (this.status !== "connected" || !this.socket) {
      return;
    }

    const payload: InputSettingsSaveMessage = {
      type: "inputSettingsSave",
      settings,
    };

    this.socket.send(JSON.stringify(payload));
  }

  // Отправляет частичное изменение объекта панели управления.
  sendControlPanelObjectUpdate(update: Omit<ControlPanelObjectUpdateMessage, "type" | "clientSessionId" | "mutationSeq">): ControlPanelMutationRef | null {
    if (this.status !== "connected" || !this.socket) {
      return null;
    }

    const mutation = this.nextControlPanelMutation();
    const payload: ControlPanelObjectUpdateMessage = {
      type: "controlPanelObjectUpdate",
      clientSessionId: this.clientSessionId,
      mutationSeq: mutation.seq,
      ...update,
    };

    this.socket.send(JSON.stringify(payload));
    return mutation;
  }

  // Отправляет частичное изменение группы оборудования панели управления.
  sendControlPanelEquipmentUpdate(update: Omit<ControlPanelEquipmentUpdateMessage, "type" | "clientSessionId" | "mutationSeq">): ControlPanelMutationRef | null {
    if (this.status !== "connected" || !this.socket) {
      return null;
    }

    const mutation = this.nextControlPanelMutation();
    const payload: ControlPanelEquipmentUpdateMessage = {
      type: "controlPanelEquipmentUpdate",
      clientSessionId: this.clientSessionId,
      mutationSeq: mutation.seq,
      ...update,
    };

    this.socket.send(JSON.stringify(payload));
    return mutation;
  }

  // Отправляет перенос содержимого между контейнерами панели управления.
  sendControlPanelContainerTransfer(transfer: Omit<ControlPanelContainerTransferMessage, "type" | "clientSessionId" | "mutationSeq">): ControlPanelMutationRef | null {
    if (this.status !== "connected" || !this.socket) {
      return null;
    }

    const mutation = this.nextControlPanelMutation();
    const payload: ControlPanelContainerTransferMessage = {
      type: "controlPanelContainerTransfer",
      clientSessionId: this.clientSessionId,
      mutationSeq: mutation.seq,
      ...transfer,
    };

    this.socket.send(JSON.stringify(payload));
    return mutation;
  }

  // Отправляет перенос топлива между контейнером и баком панели управления.
  sendControlPanelFuelTransfer(transfer: Omit<ControlPanelFuelTransferMessage, "type" | "clientSessionId" | "mutationSeq">): ControlPanelMutationRef | null {
    if (this.status !== "connected" || !this.socket) {
      return null;
    }

    const mutation = this.nextControlPanelMutation();
    const payload: ControlPanelFuelTransferMessage = {
      type: "controlPanelFuelTransfer",
      clientSessionId: this.clientSessionId,
      mutationSeq: mutation.seq,
      ...transfer,
    };

    this.socket.send(JSON.stringify(payload));
    return mutation;
  }

  // Отправляет изготовление предмета по схеме конструктора панели управления.
  sendControlPanelConstructorProduceItem(
    production: Omit<ControlPanelConstructorProduceItemMessage, "type" | "clientSessionId" | "mutationSeq">,
  ): ControlPanelMutationRef | null {
    if (this.status !== "connected" || !this.socket) {
      return null;
    }

    const mutation = this.nextControlPanelMutation();
    const payload: ControlPanelConstructorProduceItemMessage = {
      type: "controlPanelConstructorProduceItem",
      clientSessionId: this.clientSessionId,
      mutationSeq: mutation.seq,
      ...production,
    };

    this.socket.send(JSON.stringify(payload));
    return mutation;
  }

  // Отправляет изменение основной очереди конструктора панели управления.
  sendControlPanelConstructorQueueCommand(
    command: Omit<ControlPanelConstructorQueueCommandMessage, "type" | "clientSessionId" | "mutationSeq">,
  ): ControlPanelMutationRef | null {
    if (this.status !== "connected" || !this.socket) {
      return null;
    }

    const mutation = this.nextControlPanelMutation();
    const payload: ControlPanelConstructorQueueCommandMessage = {
      type: "controlPanelConstructorQueueCommand",
      clientSessionId: this.clientSessionId,
      mutationSeq: mutation.seq,
      ...command,
    };

    this.socket.send(JSON.stringify(payload));
    return mutation;
  }

  // Отправляет сохранение связанной группы оборудования панели управления.
  sendControlPanelEquipmentGroupRelationUpdate(
    update: Omit<ControlPanelEquipmentGroupRelationUpdateMessage, "type" | "clientSessionId" | "mutationSeq">,
  ): ControlPanelMutationRef | null {
    if (this.status !== "connected" || !this.socket) {
      return null;
    }

    const mutation = this.nextControlPanelMutation();
    const payload: ControlPanelEquipmentGroupRelationUpdateMessage = {
      type: "controlPanelEquipmentGroupRelationUpdate",
      clientSessionId: this.clientSessionId,
      mutationSeq: mutation.seq,
      ...update,
    };

    this.socket.send(JSON.stringify(payload));
    return mutation;
  }

  // Запрашивает у сервера последние сохраненные привязки ввода текущего аккаунта.
  requestInputSettings(): void {
    if (this.status !== "connected" || !this.socket) {
      return;
    }

    const payload: InputSettingsRequestMessage = {
      type: "inputSettingsRequest",
    };

    this.socket.send(JSON.stringify(payload));
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
      return;
    }

    if (isChatStateMessage(parsed)) {
      this.latestChatState = parsed;
      this.latestChatError = null;
      return;
    }

    if (isChatErrorMessage(parsed)) {
      this.latestChatError = parsed.message;
      this.latestChatErrorSeq++;
      return;
    }

    if (isInputSettingsMessage(parsed)) {
      this.latestInputSettings = parsed.settings;
      this.latestInputSettingsSeq++;
      this.latestInputSettingsError = null;
      return;
    }

    if (isInputSettingsErrorMessage(parsed)) {
      this.latestInputSettingsError = parsed.message;
      this.latestInputSettingsErrorSeq++;
      return;
    }

    if (isControlPanelErrorMessage(parsed)) {
      this.latestControlPanelError = parsed;
      this.latestControlPanelErrorSeq++;
      return;
    }

    if (isDockingEventMessage(parsed)) {
      this.latestDockingEvent = parsed;
      this.latestDockingEventSeq++;
      this.dockingEvents.push(parsed);
    }
  }

  // Сериализует одноразовую команду стыковки, если соединение открыто.
  private sendDockingCommand(type: DockingCommandMessage["type"]): void {
    if (this.status !== "connected" || !this.socket) {
      return;
    }

    const payload: DockingCommandMessage = { type };
    this.socket.send(JSON.stringify(payload));
  }

  // Выдает следующий номер команды панели в общей последовательности клиента.
  private nextControlPanelMutation(): ControlPanelMutationRef {
    return {
      sessionId: this.clientSessionId,
      seq: ++this.mutationSeq,
    };
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
      toggleAnchor: false,
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
