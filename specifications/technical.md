# 📋 ТЕХНИЧЕСКОЕ ЗАДАНИЕ

---

## 🎯 1. Общее описание

### Цель
Создать клиент-серверную браузерную игру в жанре 2D space shooter с:
- Массовыми PvP-боями (до 1000 одновременных игроков)
- Бесшовным огромным миром (координаты `float64`, без зон загрузки)
- Низкой задержкой (<100 мс от ввода до отклика)
- Сложным реактивным интерфейсом (чат, инвентарь, дерево прокачки, статистика в реальном времени)

### Ключевые ограничения
| Параметр | Значение |
|----------|----------|
| Клиент | Браузер (Chrome 120+, Firefox 115+, Safari 17+) |
| Сервер | Linux, Go 1.22+, многоядерный CPU |
| Сеть | WebTransport + WebSocket fallback |
| Синхронизация | Авторитетный сервер, 30 Гц тикрейт, клиентская интерполяция |
| Разработка | Код пишется по этому ТЗ; архитектура контролируется человеком |

---

## 🛠 2. Стек технологий и версий

### Сервер (Go)
| Компонент | Версия | Назначение |
|-----------|--------|------------|
| Go | 1.22+ | Язык сервера |
| quic-go | v0.48+ | WebTransport (QUIC/HTTP/3) сервер |
| google.golang.org/protobuf | v1.34+ | Protobuf-сериализация |
| gorilla/websocket | v1.5+ | WebSocket fallback для Safari/Firefox |
| prometheus/client_golang | v1.19+ | Метрики |

### Клиент (TypeScript)
| Компонент | Версия | Назначение |
|-----------|--------|------------|
| TypeScript | 5.4+ | Язык клиента |
| Phaser | 4.0.0 | Игровой движок, WebGL 2.0 рендер |
| SolidJS | 1.8+ | Реактивный UI (HUD, чат, инвентарь) |
| ts-proto | 1.170+ | Генерация TypeScript-кода из .proto |
| Vite | 5.3+ | Сборка и dev-сервер |
| web-transport-web | latest | WebTransport API в браузере |

### Инструменты
| Компонент | Версия | Назначение |
|-----------|--------|------------|
| protoc | 26+ | Компилятор protobuf-файлов |
| Node.js | 20 LTS | Рантайм для сборки клиента |
| ts-proto plugin | 1.170+ | Плагин protoc для генерации TS |

### Сетевой протокол
- **WebTransport** (Chrome, Edge, Яндекс.Браузер) — основной протокол, задержка ~30-80 мс
- **WebSocket** (Safari, Firefox, остальные) — fallback, задержка ~50-150 мс

### Поддерживаемые браузеры
| Браузер | Протокол | Примечание |
|---------|----------|------------|
| Chrome 97+ | WebTransport | |
| Edge 97+ | WebTransport | |
| Яндекс.Браузер 24+ | WebTransport | На базе Chromium |
| Safari 16+ | WebSocket | WebTransport не поддерживается |
| Firefox 115+ | WebSocket | WebTransport только с флагом |

---

## 🧱 3. Архитектура системы

### 3.1. Общая схема
```
[Браузер: Клиент]
  ├── Ядро игры: TypeScript + Phaser 4 (WebGL 2.0)
  ├── UI-слой: SolidJS (реактивный, отдельный DOM-контейнер)
  ├── Сеть: WebTransport (web-transport-web) + WebSocket fallback
  ├── Сериализация: protobuf (ts-proto)
  └── Синхронизация: интерполяция + клиентское предсказание

[Сервер: Go]
  ├── Транспорт: quic-go (WebTransport/HTTP3) + gorilla/websocket (fallback)
  ├── Протокол: protobuf (google.golang.org/protobuf)
  ├── Пространство: spatial hash (фиксированная сетка)
  ├── Параллелизм: worker pool + чанки по 16-32 сущности
  ├── Тик-лук: фазы Input → Simulate → Barrier → Apply → Broadcast
  ├── AOI: Area of Interest router (рассылка только видимого)
  └── Мониторинг: pprof + Prometheus metrics

[Контракт]
  └── Единый файл: shared/game.proto → генерация кода для Go и TS
```

### 3.2. Поток данных (тик)
```
1. Клиент: сбор ввода → отправка PlayerInput{tick, playerId, controls, timestamp}
2. Сервер: приём вводов → сортировка по server_receive_time → привязка к tick N
3. Сервер: spatial hash update → миграция сущностей по клеткам
4. Сервер: worker pool обрабатывает чанки (параллельно):
   • Чтение 3×3 соседей (read-only snapshot)
   • Расчёт физики/коллизий/урона
   • Запись событий в локальный []Event
5. Сервер: barrier (WaitGroup) → атомарное применение событий к миру
6. Сервер: генерация GameStateDelta{tick, entities[], events[]} по AOI
7. Сервер: отправка дельты клиентам через WebTransport/WebSocket
8. Клиент: буферизация тиков → интерполяция → рендер Phaser + обновление UI SolidJS
```

---

## 📐 4. Пространственная модель и оптимизация

### 4.1. Единицы измерения углов

**Все углы в проекте — только в радианах.**

- Сервер хранит и отправляет углы в радианах
- Клиент получает углы в радианах
- Физика, коллизии, интерполяция — всё в радианах
- Никаких градусов в коде, протоколе или конфигах

| Где | Единица | Диапазон |
|-----|---------|----------|
| Сервер (Go) | радианы | непрерывные (без нормализации) |
| Сеть (proto `angle`) | радианы | `float` |
| Клиент (TS) | радианы | непрерывные (без нормализации) |
| Рендер (Phaser) | радианы | Phaser сам обработает |

### 4.2. Система координат

**Сервер и клиент используют разные направления оси Y.**

| Где | Направление Y | Направление X |
|-----|---------------|---------------|
| Сервер (игровой мир) | **вверх** (+Y = север) | вправо |
| Клиент (Phaser 4.0.0) | **вниз** (+Y = юг, экран) | вправо |

При рендере серверные координаты преобразуются в экранные:

### 4.3. Spatial Hash (сервер)
```go
// КОНТРАКТ: функция должна быть детерминированной и без аллокаций в горячем цикле
func CellKey(x, y, cellSize float64) (int, int) {
    return int(math.Floor(x / cellSize)), int(math.Floor(y / cellSize))
}
```

**Параметры:**
- `cell_size = 1500` (единиц мира) — рассчитано как `1.5 × max_weapon_range (1000)`
- Проверка взаимодействий: **3×3 соседние клетки** всегда покрывают дальность оружия
- Пустые клетки не хранятся в памяти (`map[CellKey]*Cell`)

---

## 🌐 5. Сетевой протокол (protobuf)

### 5.1. Требования к реализации
| Компонент | Сервер (Go) | Клиент (TS) |
|-----------|-------------|-------------|
| Генерация кода | `protoc --go_out=. --go_opt=paths=source_relative game.proto` | `protoc --plugin=node_modules/.bin/protoc-gen-ts_proto --ts_proto_out=. game.proto` |
| WebTransport | `quic-go` HTTP/3 режим | `web-transport-web` |
| WebSocket fallback | `gorilla/websocket` отдельный слушатель | нативный `WebSocket` API браузера |
| Сериализация | `google.golang.org/protobuf/proto.Marshal` (zero-copy где возможно) | `ts-proto` сгенерированный код |
| Приоритизация | 3 уровня: `Critical` (позиции/урон), `Normal` (эффекты), `Bulk` (чат/инвентарь) | Отправка пакетов в соответствии с приоритетом данных |

---

## 🖥️ 6. Клиентская архитектура

### 6.1. Структура проекта (TypeScript)
```
client/
├── world/                  # Точка входа, сборка (Vite), игровой мир
│   ├── main.ts             #   Инициализация Phaser, оркестрация
│   ├── index.html          #   HTML-шаблон
│   ├── package.json        #   Зависимости (Phaser, Vite, TypeScript)
│   ├── vite.config.ts      #   Конфиг Vite + алиасы @ui, @utils
│   ├── tsconfig.json       #   Конфиг TypeScript
│   ├── public/             #   Статические файлы (копируются без обработки)
│   │   └── images/         #     Текстуры, спрайты
│   ├── src/
│   │   ├── GameLoop.ts     #   requestAnimationFrame + delta-time
│   │   ├── InputHandler.ts #   Сбор клавиатуры/мыши, троттлинг
│   │   ├── Prediction.ts   #   Клиентское предсказание движения
│   │   ├── Interpolation.ts #  Интерполяция серверных снапшотов
│   │   ├── Network/
│   │   │   ├── Transport.ts    #   WebTransport + WebSocket, авто-выбор
│   │   │   ├── Protocol.ts     #   Обёртки над protobuf-сообщениями
│   │   │   └── Reconnect.ts    #   Логика реконнекта с буферизацией
│   │   └── Render/
│   │       ├── PhaserConfig.ts #   Настройка WebGL 2.0, камеры, слоёв
│   │       ├── ShipRenderer.ts #   Отрисовка кораблей, эффекты
│   │       └── ParticlePool.ts #   Пул частиц для взрывов/следов
├── ui/
│   ├── solid-main.tsx      # Точка входа SolidJS, рендер в #ui-root
│   ├── stores/
│   │   ├── gameState.ts    # createSignal<GameState>, синхронизация с игрой
│   │   ├── chat.ts         # Логика чата, история, отправка
│   │   └── inventory.ts    # Инвентарь, активация предметов
│   ├── components/
│   │   ├── HUD.tsx         # HP, щит, патроны, пинг, компас
│   │   ├── ChatWindow.tsx  # Окно чата с авто-скроллом
│   │   ├── InventoryGrid.tsx # Сетка инвентаря с drag-and-drop
│   │   └── TechTree.tsx    # Дерево прокачки (сложная реактивность)
│   └── bridge/
│       ├── GameToUI.ts     # EventEmitter: игра → SolidJS стейт
│       └── UIToGame.ts     # События UI → команды игре (купить, активировать)
└── utils/
    ├── math.ts             # lerp, lerpAngle, distanceSq, worldToScreen — углы в радианах
    └── pool.ts             # Generic object pool для аллокаций
```

### 6.2. Мост между Phaser и SolidJS
```ts
// КОНТРАКТ: client/ui/bridge/GameToUI.ts
import { createSignal, Signal } from 'solid-js';

export interface UIState {
  hp: number;
  shield: number;
  ammo: { primary: number; secondary: number };
  ping: number;
  chat: ChatMessage[];
  // ... другие UI-поля
}

export const uiState: Signal<UIState> = createSignal({
  hp: 100, shield: 100, ammo: { primary: 100, secondary: 10 },
  ping: 0, chat: []
});

// Игра вызывает это при изменении состояния
export function pushToUI(update: Partial<UIState>) {
  uiState[1](prev => ({ ...prev, ...update }));
}

// КОНТРАКТ: client/ui/bridge/UIToGame.ts
export type UICommand =
  | { type: 'CHAT_SEND'; text: string }
  | { type: 'ITEM_USE'; itemId: string; slot: number }
  | { type: 'TECH_RESEARCH'; techId: string };

export const uiCommandQueue: UICommand[] = [];

export function queueUICommand(cmd: UICommand) {
  uiCommandQueue.push(cmd);
  // Игра опрашивает этот массив в своём цикле
}
```

### 6.3. Требования к SolidJS-компонентам
- **Только UI-логика**: никаких расчётов физики, коллизий, предсказаний
- **Реактивность через сигналы**: использовать `createSignal`, `createMemo`, `createEffect`
- **Избегать перерендеров**: `createMemo` для вычисляемых значений, `keyed` списки для чата/инвентаря
- **Типизация**: все пропсы и стейт — строгие TypeScript-интерфейсы
- **Изоляция**: компоненты не импортируют игровые модули, только через `bridge/`

---

## ⚙️ 7. Серверная архитектура (Go)

### 7.1. Структура проекта
```
server/
├── main.go             # Точка входа, инициализация, graceful shutdown
├── config.go           # Загрузка из env/JSON, валидация
├── proto/              # Сгенерированные из game.pb.go файлы
├── network.go          # WebTransport listener + WebSocket fallback, PlayerSession
├── world.go            # Главный объект: сущности, тик-лук, физика
├── entity.go           # Entity, Ship интерфейсы
├── events.go           # Event, детерминированный резолв коллизий
├── metrics.go          # Метрики: tick_duration, Prometheus, pprof
└── utils.go            # math утилиты — углы в радианах, distanceSq, пулы
```

## 🚫 10. Анти-паттерны (запрещено генерировать)

### 10.1. Сервер
```go
// ❌ НЕЛЬЗЯ: использовать math.Sqrt в горячем цикле коллизий
if math.Sqrt(dx*dx+dy*dy) < range { ... }
// ✅ НАДО: сравнивать квадрат расстояния
if dx*dx+dy*dy < range*range { ... }

// ❌ НЕЛЬЗЯ: блокировать воркеры мьютексом на чтение соседей
cell.mu.RLock() // внутри параллельного воркера
// ✅ НАДО: использовать read-only snapshot, созданный до запуска воркеров

// ❌ НЕЛЬЗЯ: отправлять полное состояние объекта каждый тик
SendFullEntityState(entity)
// ✅ НАДО: отправлять только дельту (изменённые поля)
SendEntityDelta(entity, changedFields)
```

### 10.2. Клиент
```ts
// ❌ НЕЛЬЗЯ: доверять клиенту в расчёте урона
if (distanceToTarget < weapon.range) { applyDamage(); }
// ✅ НАДО: клиент только отображает, сервер авторитетен в логике

// ❌ НЕЛЬЗЯ: создавать новые объекты в игровом цикле без пула
new Particle() // каждый кадр → давление на GC
// ✅ НАДО: использовать object pool
const p = particlePool.acquire(); p.reset(); ...

// ❌ НЕЛЬЗЯ: обновлять SolidJS стейт напрямую из игрового цикла
uiState[1](newState) // 60 раз в секунду → лишние ре-рендеры
// ✅ НАДО: батчить обновления, использовать createMemo для вычисляемых полей

// ❌ НЕЛЬЗЯ: нормализовать углы
// Углы поворота в проекте нормализовать ЗАПРЕЩЕНО!
// ✅ НАДО: хранить и передавать углы как непрерывные значения, накапливая вращение.
// Нормализация не нужна — Phaser корректно отображает любые радианы.
```

### 10.3. Сеть
```go
// ❌ НЕЛЬЗЯ: использовать JSON для реал-тайм пакетов
json.Marshal(packet) // медленно, много трафика
// ✅ НАДО: protobuf + бинарная сериализация

// ❌ НЕЛЬЗЯ: отправлять пакеты без приоритизации
stream.Send(packet) // все пакеты в одну очередь
// ✅ НАДО: 3 уровня приоритета: Critical > Normal > Bulk
```

---

## 📂 13. Структура каталогов проекта

```
space-game-04/
├── client/                     # Клиентский код (TypeScript + Phaser 4 + SolidJS)
│   ├── world/                  #   Точка входа, сборка, игровой мир
│   ├── ui/                     #   UI-слой на SolidJS (без зависимостей)
│   └── utils/                  #   Утилиты (математика, пулы объектов)
├── server/                     # Серверный код (Go)
├── shared/                     # Общие ресурсы (протокол, типы, константы)
└── specifications/             # Документация проекта
```

