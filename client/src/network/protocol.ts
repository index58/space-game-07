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
  // Справочник моделей космических объектов.
  CosmicObjectModel: ReferenceTable<CosmicObjectModelReference>;
  // Справочник моделей предметов.
  ItemModel: ReferenceTable<ItemModelReference>;
  // Справочник чертежей объектов.
  Blueprint: ReferenceTable;
  // Справочник компонентов чертежей.
  BlueprintComponent: ReferenceTable;
  // Справочник схем предметов.
  Schema: ReferenceTable;
  // Справочник компонентов схем.
  SchemaComponent: ReferenceTable;
};
