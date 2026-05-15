// Отражает текущее состояние WebSocket-клиента для сцены и отладочного слоя.
export type ConnectionStatus = "connecting" | "connected" | "waiting";

// Хранит последний пользовательский ввод, еще не упакованный в сетевое сообщение.
export type ClientInputState = {
  // Запрос продольной тяги вперед.
  thrustForward: boolean;
  // Запрос продольной тяги назад.
  thrustBackward: boolean;
  // Запрос поперечной тяги влево.
  thrustLeft: boolean;
  // Запрос поперечной тяги вправо.
  thrustRight: boolean;
  // Одноразовый запрос переключения якоря.
  toggleAnchor: boolean;
  // Изменение целевого угла поворота с прошлого пакета.
  targetRotationDelta: number;
};

// Добавляет к вводу тип сообщения и порядковый номер для серверного протокола.
export type ClientInputMessage = ClientInputState & {
  // Вид сообщения, по которому сервер отличает ввод от других пакетов.
  type: "input";
  // Порядковый номер отправленного пакета.
  seq: number;
};

export type RandomShipMessage = {
  // Вид команды, по которому сервер запускает смену модели корабля.
  type: "randomShip";
};

export type TestClaimFocusedObjectOwnerMessage = {
  // Вид тестовой команды, по которому сервер присваивает объект в фокусе текущему персонажу.
  type: "testClaimFocusedObjectOwner";
};

export type DockingCommandMessage = {
  // Вид отдельной команды стыковки для серверного маршрутизатора.
  type: "dockingRequest" | "dockingApprove" | "dockingReject" | "dockingUndock";
};

export type DockingEventKind = "dockingRequestStarted" | "dockingProcessStarted" | "dockingFinished" | "dockingNotification";

export type DockingEventMessage = {
  // Вид сообщения, по которому клиент отличает события стыковки от снимков мира.
  type: "dockingEvent";
  // Событие, определяющее окно или уведомление.
  kind: DockingEventKind;
  // Роль текущего клиента в парном окне стыковки.
  role?: "sender" | "receiver";
  // Текст всплывающего уведомления.
  message?: string;
  // Длительность окна с прогрессом в секундах.
  duration?: number;
};

export type DockingWindowState = {
  // Фаза, от которой зависит текст маленького окна.
  kind: "request" | "process";
  // Роль текущего клиента в парном окне.
  role: "sender" | "receiver";
  // Время появления окна в миллисекундах игрового кадра.
  startedAtMs: number;
  // Длительность прогресса в миллисекундах.
  durationMs: number;
};

export type DockingNotification = {
  // Локальный идентификатор для стабильной отрисовки списка.
  id: number;
  // Текст отдельного всплывающего уведомления.
  message: string;
  // Время автоматического скрытия в миллисекундах игрового кадра.
  expiresAtMs: number;
};

// Передает серверу текст для выбранной вкладки или адресного дуэта.
export type ChatSendMessage = {
  // Вид команды, по которому сервер отличает чат от игрового ввода.
  type: "chatSend";
  // Чат, куда уйдет обычный текст без адресного ника.
  chatId?: number;
  // Ник аккаунта, которому адресована личная команда.
  targetNickname?: string;
  // Содержимое сообщения после разбора локальной строки ввода.
  text: string;
};

// Сообщает серверу, какую вкладку игрок выбрал и прочитал.
export type ChatSelectMessage = {
  // Вид команды, по которому сервер отличает выбор вкладки.
  type: "chatSelect";
  // Выбранный игроком чат.
  chatId: number;
};

export type InputSettingPayload = {
  // Игровое действие, для которого задан ввод.
  actionTypeId: number;
  // Событие ввода, выбранное для действия.
  inputEventTypeId: number;
};

export type InputSettingsMessage = {
  // Вид сообщения с текущими настройками аккаунта.
  type: "inputSettings";
  // Сохраненные привязки текущего аккаунта.
  settings: InputSettingPayload[];
};

export type InputSettingsSaveMessage = {
  // Вид команды сохранения настроек ввода.
  type: "inputSettingsSave";
  // Полный список выбранных привязок.
  settings: InputSettingPayload[];
};

export type InputSettingsRequestMessage = {
  // Вид команды запроса актуальных настроек ввода.
  type: "inputSettingsRequest";
};

export type InputSettingsErrorMessage = {
  // Вид сообщения с ошибкой сохранения настроек.
  type: "inputSettingsError";
  // Текст ошибки для окна настроек.
  message: string;
};

export type ControlPanelMutationAck = {
  // Сессия клиента, к которой относится подтверждение.
  sessionId: string;
  // Последний обработанный сервером номер мутации этой сессии.
  lastAppliedSeq: number;
};

export type ControlPanelMutationRef = {
  // Сессия клиента, отправившая команду.
  sessionId: string;
  // Порядковый номер команды внутри сессии.
  seq: number;
};

export type ControlPanelObjectUpdateMessage = {
  // Вид команды изменения управляемого объекта.
  type: "controlPanelObjectUpdate";
  // Сессия клиента, отправившая команду.
  clientSessionId: string;
  // Порядковый номер команды внутри сессии.
  mutationSeq: number;
  // Новое состояние включения объекта.
  enabled?: boolean;
  // Новое пользовательское название объекта.
  title?: string;
};

export type ControlPanelEquipmentUpdateMessage = {
  // Вид команды изменения группы оборудования.
  type: "controlPanelEquipmentUpdate";
  // Сессия клиента, отправившая команду.
  clientSessionId: string;
  // Порядковый номер команды внутри сессии.
  mutationSeq: number;
  // Группа оборудования, которую нужно изменить.
  equipmentGroupId: number;
  // Новое состояние включения группы.
  enabled?: boolean;
  // Новое количество включенных единиц.
  enabledCount?: number;
  // Новое пользовательское название группы.
  title?: string;
};

export type ControlPanelEquipmentGroupRelationUpdateMessage = {
  // Вид команды сохранения связи групп оборудования.
  type: "controlPanelEquipmentGroupRelationUpdate";
  // Сессия клиента, отправившая команду.
  clientSessionId: string;
  // Порядковый номер команды внутри сессии.
  mutationSeq: number;
  // Группа оборудования, для которой сохраняется выбор.
  equipmentGroupId: number;
  // Вид связи по неизменяемому строковому идентификатору.
  relationTypeAcronym: "Source" | "Destination" | "Opposite";
  // Группа оборудования, выбранная игроком в связанной панели.
  relatedEquipmentGroupId: number;
};

export type ControlPanelContainerTransferMessage = {
  // Вид команды переноса содержимого между контейнерами.
  type: "controlPanelContainerTransfer";
  // Сессия клиента, отправившая команду.
  clientSessionId: string;
  // Порядковый номер команды внутри сессии.
  mutationSeq: number;
  // Контейнер, из которого переносятся предметы.
  sourceContainerEquipmentGroupId: number;
  // Контейнер, в который переносятся предметы.
  targetContainerEquipmentGroupId: number;
  // Группы предметов, выбранные для переноса.
  itemGroupIds: number[];
  // Количество предметов для частичного переноса одной выбранной группы.
  amount?: number;
};

export type ControlPanelFuelTransferMessage = {
  // Вид команды переноса топлива между контейнером и баком.
  type: "controlPanelFuelTransfer";
  // Сессия клиента, отправившая команду.
  clientSessionId: string;
  // Порядковый номер команды внутри сессии.
  mutationSeq: number;
  // Контейнер, из которого берётся или куда сливается топливо.
  containerEquipmentGroupId: number;
  // Топливный бак, с которым работает игрок.
  fuelTankEquipmentGroupId: number;
  // Группы топлива, выбранные в контейнере для заливки.
  itemGroupIds: number[];
  // Количество топлива для залива в бак или слива из бака.
  amount?: number;
};

export type ControlPanelConstructorProduceItemMessage = {
  // Вид команды изготовления предмета в конструкторе.
  type: "controlPanelConstructorProduceItem";
  // Сессия клиента, отправившая команду.
  clientSessionId: string;
  // Порядковый номер команды внутри сессии.
  mutationSeq: number;
  // Группа конструкторов, которая выполняет изготовление.
  constructorEquipmentGroupId: number;
  // Контейнер, из которого списываются компоненты.
  materialContainerEquipmentGroupId: number;
  // Контейнер, в который кладется готовая продукция.
  productContainerEquipmentGroupId?: number;
  // Схема предмета, выбранная для изготовления.
  schemaId?: number;
  // Чертёж объекта, выбранный для изготовления.
  blueprintId?: number;
  // Количество запусков изготовления по выбранной схеме.
  amount: number;
};

export type ControlPanelConstructorQueueCommand = "skipNext" | "skipAllNext" | "cancel" | "cancelAll";

export type ControlPanelConstructorQueueCommandMessage = {
  // Вид команды изменения очереди конструктора.
  type: "controlPanelConstructorQueueCommand";
  // Сессия клиента, отправившая команду.
  clientSessionId: string;
  // Порядковый номер команды внутри сессии.
  mutationSeq: number;
  // Группа конструкторов, очередь которой меняется.
  constructorEquipmentGroupId: number;
  // Строка основной очереди, выбранная игроком.
  jobId: number;
  // Действие над выбранной строкой и следующими строками.
  command: ControlPanelConstructorQueueCommand;
};

export type ControlPanelErrorMessage = {
  // Вид сообщения с отказом команды панели управления.
  type: "controlPanelError";
  // Сессия клиента, команда которой была отклонена.
  clientSessionId: string;
  // Номер отклоненной команды.
  mutationSeq: number;
  // Текст ошибки для диагностики клиента.
  message: string;
};

// Описывает одну строку истории в панели чата.
export type ChatMessage = {
  // Уникальный числовой идентификатор записи.
  id: number;
  // Чат, которому принадлежит строка.
  chatId: number;
  // Персонаж, от имени которого сохранено сообщение.
  senderCharacterId: number;
  // Временное отображаемое имя из учетной записи отправителя.
  senderNickname: string;
  // Тип сообщения для выбора оформления.
  messageTypeAcronym: string;
  // Текст, видимый игрокам.
  text: string;
  // Цвет строки в RGB-HEX без решетки.
  color: string;
  // Время записи в серверном формате.
  sentTime: string;
};

// Описывает одну доступную вкладку с последними сообщениями.
export type ChatTab = {
  // Уникальный числовой идентификатор чата.
  chatId: number;
  // Подпись вкладки в HUD.
  title: string;
  // Тип сообщества, которому принадлежит чат.
  communityTypeAcronym: string;
  // Ключ личной переписки двух персонажей.
  duoChatKey: string;
  // Количество сообщений, которые появились после последнего чтения.
  unreadCount?: number;
  // Последние строки истории в порядке от старых к новым.
  messages: ChatMessage[];
};

// Передает клиенту доступные вкладки и выбранный сервером чат.
export type ChatStateMessage = {
  // Вид сообщения для клиентского маршрутизатора.
  type: "chatState";
  // Вкладки, доступные текущему персонажу.
  tabs: ChatTab[];
  // Чат, который нужно выбрать после серверного действия.
  selectedChatId: number;
};

// Передает причину отказа команды чата.
export type ChatErrorMessage = {
  // Вид сообщения для клиентского маршрутизатора.
  type: "chatError";
  // Текст ошибки для HUD.
  message: string;
};

// Передает клиенту секрет автоматически созданной учетной записи.
export type AuthMessage = {
  // Вид сообщения, по которому клиент отличает авторизацию от снимка мира.
  type: "auth";
  // Секрет для следующих подключений к серверу.
  token: string;
};

// Повторяет серверный формат хранения одного космического объекта.
export type CosmicObject = {
  // Уникальный числовой идентификатор записи.
  ID: number;
  // Пользовательское название объекта в игровом мире.
  Title: string;
  // Модель, от которой взяты базовые характеристики и графика.
  CosmicObjectModelID: number;
  // Персонаж-владелец, если объект принадлежит игроку.
  OwnerCharacterID: number;
  // Временное тестовое имя владельца для информационной панели.
  OwnerName?: string;
  // NPC-клан-владелец, если объект не принадлежит персонажу.
  OwnerNpcClanID: number;
  // Персонаж, создавший объект.
  CreatorCharacterID: number;
  // Текущая суммарная масса объекта и содержимого.
  Mass: number;
  // Максимальный объем оборудования или содержимого.
  Capacity: number;
  // Верхняя граница прочности брони.
  MaxArmor: number;
  // Максимальная линейная скорость в метрах за секунду.
  MaxSpeed: number;
  // Максимальная угловая скорость в радианах за секунду.
  MaxAngularSpeed: number;
  // Горизонтальная координата положения в мире.
  X: number;
  // Вертикальная координата положения в мире.
  Y: number;
  // Текущий угол поворота в радианах без нормализации.
  Rotation: number;
  // Текущее количество единиц брони.
  Armor: number;
  // Доступная продольная сила тяги.
  MaxAlongForce: number;
  // Доступная поперечная сила тяги.
  MaxAcrossForce: number;
  // Доступный крутящий момент.
  MaxTorque: number;
  // Суммарная вырабатываемая мощность оборудования.
  GeneratingPower: number;
  // Суммарная потребляемая мощность оборудования.
  ConsumingPower: number;
  // Фактически примененная продольная тяга на текущем шаге.
  AlongForce: number;
  // Фактически примененная поперечная тяга на текущем шаге.
  AcrossForce: number;
  // Фактически примененный крутящий момент на текущем шаге.
  Torque: number;
  // Разрешает объекту работать и участвовать в симуляции.
  Enabled: boolean;
  // Время последнего получения урона для боевых и ремонтных правил.
  LastReceivedDamageTime: number;
  // Запрещает физическое перемещение объекта.
  Anchored: boolean;
  // Главный объект кластера, если объект пристыкован.
  ClusterMainCosmicObjectID?: number;
  // Сложность устройства для производства и оценки стоимости.
  Complexity: number;
  // Объем, уже занятый содержимым или оборудованием.
  OccupiedVolume: number;
  // Максимальный запас топлива.
  MaxFuel: number;
  // Текущий запас топлива.
  Fuel: number;
  // Текущая длина вектора скорости.
  Speed: number;
  // Горизонтальная компонента текущей скорости.
  VelocityX: number;
  // Вертикальная компонента текущей скорости.
  VelocityY: number;
  // Текущая угловая скорость в радианах за секунду.
  AngularSpeed: number;
  // Угол, к которому автоматика поворота ведет объект.
  TargetRotation: number;
};

// Повторяет серверный формат группы установленного оборудования.
export type EquipmentGroup = {
  // Уникальный числовой идентификатор записи.
  ID: number;
  // Космический объект, на котором установлено оборудование.
  CosmicObjectID: number;
  // Пользовательское название группы оборудования.
  Title: string;
  // Модель установленного оборудования.
  EquipmentItemModelID: number;
  // Количество единиц оборудования в группе.
  Count: number;
  // Количество включенных единиц оборудования в группе.
  EnabledCount: number;
  // Разрешает группе оборудования получать питание.
  Enabled: boolean;
  // Показывает, выполняет ли оборудование работу в текущем тике.
  Active: boolean;
  // Время начала последней перезарядки в миллисекундах Unix.
  LastRechargeStartTime: number;
};

// Повторяет серверный формат группы предметов внутри контейнера.
export type ItemGroup = {
  // Уникальный числовой идентификатор записи.
  ID: number;
  // Группа контейнерного оборудования, внутри которой лежат предметы.
  ContainerEquipmentGroupID: number;
  // Модель предмета внутри контейнера.
  ContentItemModelID: number;
  // Количество предметов указанной модели.
  Count: number;
};

export type EquipmentGroupRelation = {
  // Уникальный числовой идентификатор записи.
  ID: number;
  // Группа оборудования, для которой сохранена связь.
  EquipmentGroupID: number;
  // Вид связи между группами оборудования.
  RelationTypeID: number;
  // Группа оборудования, выбранная для указанного вида связи.
  RelatedEquipmentGroupID: number;
};

export type ConstructorProductionJob = {
  // Уникальный числовой идентификатор задания.
  id: number;
  // Конструктор, к очереди которого относится задание.
  constructorEquipmentGroupId: number;
  // Очередь задания: основная или вспомогательная.
  queueType: "main" | "auxiliary";
  // Схема, по которой изготавливается предмет.
  schemaId: number;
  // Чертёж, по которому изготавливается объект.
  blueprintId: number;
  // Модель предмета, который получится после завершения.
  productItemModelId: number;
  // Модель объекта, который появится после завершения.
  productCosmicObjectModelId: number;
  // Количество предметов, которое получится после завершения.
  productCount: number;
  // Количество предметов, которое еще осталось изготовить по строке.
  remainingCount: number;
  // Общее количество предметов, запланированное по строке.
  totalCount: number;
  // Оставшееся время изготовления в секундах.
  remainingTime: number;
  // Полное время изготовления в секундах.
  totalTime: number;
  // Показывает, что задание сейчас выполняется.
  running: boolean;
  // Родительская строка, для компонента которой нужна эта вспомогательная строка.
  parentJobId: number;
};

// Является полным состоянием мира, которое сервер регулярно отправляет клиенту.
export type SnapshotMessage = {
  // Вид сообщения, по которому клиент отличает снимок от других пакетов.
  type: "snapshot";
  // Номер шага симуляции, на котором сделан снимок.
  tick: number;
  // Управляемый объект получателя снимка.
  selfObjectId: number;
  // Объекты мира, видимые клиенту в текущем снимке.
  objects: CosmicObject[];
  // Группы оборудования, нужные UI для панели пилота.
  equipmentGroups: EquipmentGroup[];
  // Сохранённые связи выбора групп оборудования.
  equipmentGroupRelations: EquipmentGroupRelation[];
  // Группы предметов внутри контейнерного оборудования.
  itemGroups: ItemGroup[];
  // Задания изготовления в очередях конструкторов.
  constructorProductionJobs: ConstructorProductionJob[];
  // Подтверждение обработанных команд панели для текущей сессии.
  clientMutationAck?: ControlPanelMutationAck;
};

// Хранит одну JSON-таблицу справочника в серверном формате.
export type ReferenceTable<TItem = Record<string, unknown>> = {
  // Последний выданный числовой идентификатор записей.
  MaxID: number;
  // Записи таблицы по строковому представлению числового идентификатора.
  Items: Record<string, TItem>;
};

// Описывает только поля модели, нужные клиенту для выбора и масштабирования текстуры.
export type CosmicObjectModelReference = Record<string, unknown> & {
  // Уникальный числовой идентификатор записи.
  ID: number;
  // Тип космического объекта, к которому относится модель.
  CosmicObjectTypeID: number;
  // Путь к основной текстуре объекта в игровом мире.
  TextureFilePath: string;
  // Полная ширина текстуры в пикселях.
  TextureWidth: number;
  // Полная высота текстуры в пикселях.
  TextureHeight: number;
  // Горизонтальная координата центра тела на текстуре.
  TextureBodyOriginX: number;
  // Вертикальная координата центра тела на текстуре.
  TextureBodyOriginY: number;
  // Количество пикселей текстуры на один метр мира.
  TextureScale: number;
  // Рассчитанная ширина физического тела в метрах.
  BodyWidth: number;
  // Рассчитанная длина физического тела в метрах.
  BodyLength: number;
};

// Описывает поля типа предмета, нужные клиентскому UI.
export type ItemtypeReference = Record<string, unknown> & {
  // Уникальный числовой идентификатор записи.
  ID: number;
  // Неизменяемый строковый идентификатор типа.
  Acronym: string;
  // Разрешает назначать предметы этого типа в панель пилота.
  IsPilotInstrument: boolean;
  // Разрешает внутреннее использование предметов этого типа из панели управления.
  IsInternalUsable: boolean;
};

export type RelationTypeReference = Record<string, unknown> & {
  // Уникальный числовой идентификатор записи.
  ID: number;
  // Неизменяемый строковый идентификатор вида связи.
  Acronym: string;
};

// Описывает поля модели предмета, нужные клиентскому UI.
export type ItemModelReference = Record<string, unknown> & {
  // Уникальный числовой идентификатор записи.
  ID: number;
  // Тип предмета из справочника типов.
  ItemtypeID: number;
  // Неизменяемый строковый идентификатор модели.
  Acronym: string;
  // Русское название модели для интерфейса.
  TitleRu?: string;
  // Английское название модели для интерфейса.
  TitleEn?: string;
  // Путь к файлу иконки предмета.
  IconFilePath?: string;
  // Вместимость магазина, если у инструмента есть магазин.
  MagazineCapacity?: number;
};

export type ActionTypeReference = Record<string, unknown> & {
  // Уникальный числовой идентификатор записи.
  ID: number;
  // Русское название действия.
  TitleRu: string;
  // Английское название действия.
  TitleEn: string;
  // Неизменяемый строковый идентификатор действия.
  Acronym: string;
  // Пояснение действия.
  Description?: string;
};

export type InputEventTypeReference = Record<string, unknown> & {
  // Уникальный числовой идентификатор записи.
  ID: number;
  // Русское название события.
  TitleRu: string;
  // Английское название события.
  TitleEn: string;
  // Неизменяемый строковый идентификатор события.
  Acronym: string;
  // Системное строковое значение браузерного события.
  SystemStringValue: string;
  // Системное числовое значение, если оно есть.
  SystemIntegerValue: number;
  // Пояснение события.
  Description?: string;
};

export type DefaultActionInputSettingReference = Record<string, unknown> & {
  // Уникальный числовой идентификатор записи.
  ID: number;
  // Действие, выполняемое по умолчанию.
  ActionTypeID: number;
  // Событие ввода для действия по умолчанию.
  InputEventTypeID: number;
};

export type BlueprintReference = Record<string, unknown> & {
  // Уникальный числовой идентификатор записи.
  ID: number;
  // Русское название чертежа.
  TitleRu: string;
  // Английское название чертежа.
  TitleEn: string;
  // Модель космического объекта, получаемого по чертежу.
  CosmicObjectModelID: number;
  // Базовое время изготовления объекта в секундах.
  ProductionBaseTime: number;
};

export type BlueprintComponentReference = Record<string, unknown> & {
  // Уникальный числовой идентификатор записи.
  ID: number;
  // Чертёж, которому принадлежит компонент.
  BlueprintID: number;
  // Модель предмета, необходимого для изготовления.
  ComponentItemModelID: number;
  // Количество предметов этой модели для изготовления.
  Count: number;
};

export type SchemaReference = Record<string, unknown> & {
  // Уникальный числовой идентификатор записи.
  ID: number;
  // Русское название схемы.
  TitleRu: string;
  // Английское название схемы.
  TitleEn: string;
  // Модель предмета, получаемого по схеме.
  ItemModelID: number;
  // Количество предметов, получаемых за одно изготовление.
  Count: number;
  // Базовое время изготовления предметов в секундах.
  ProductionBaseTime: number;
};

export type SchemaComponentReference = Record<string, unknown> & {
  // Уникальный числовой идентификатор записи.
  ID: number;
  // Схема, которой принадлежит компонент.
  SchemaID: number;
  // Модель предмета, необходимого для изготовления.
  ComponentItemModelID: number;
  // Количество предметов этой модели для изготовления.
  Count: number;
};

// Является полным пакетом справочников, загружаемым перед подключением к миру.
export type ReferenceDataMessage = {
  // Вид сообщения, по которому клиент проверяет назначение ответа.
  type: "referenceData";
  // Справочник NPC-кланов.
  NpcClan: ReferenceTable;
  // Справочник типов космических объектов.
  CosmicObjectType: ReferenceTable;
  // Справочник типов предметов.
  Itemtype: ReferenceTable<ItemtypeReference>;
  // Справочник видов связей групп оборудования.
  RelationType?: ReferenceTable<RelationTypeReference>;
  // Справочник моделей космических объектов.
  CosmicObjectModel: ReferenceTable<CosmicObjectModelReference>;
  // Справочник моделей предметов.
  ItemModel: ReferenceTable<ItemModelReference>;
  // Справочник чертежей объектов.
  Blueprint: ReferenceTable<BlueprintReference>;
  // Справочник компонентов чертежей.
  BlueprintComponent: ReferenceTable<BlueprintComponentReference>;
  // Справочник схем предметов.
  Schema: ReferenceTable<SchemaReference>;
  // Справочник компонентов схем.
  SchemaComponent: ReferenceTable<SchemaComponentReference>;
  // Справочник игровых действий.
  ActionType: ReferenceTable<ActionTypeReference>;
  // Справочник событий ввода.
  InputEventType: ReferenceTable<InputEventTypeReference>;
  // Привязки ввода по умолчанию.
  DefaultActionInputSetting: ReferenceTable<DefaultActionInputSettingReference>;
};
