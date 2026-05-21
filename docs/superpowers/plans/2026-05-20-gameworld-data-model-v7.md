# Gameworld Data Model V7 Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Переписать сервер, клиент и JSON-данные под новую модель из `specifications/gameworld_data_model.go`.

**Architecture:** Старые отдельные связи групп оборудования удаляются и становятся обычными полями `EquipmentGroup`. Производство перестает храниться как время и становится заданием с энергией работы. Клиентский контракт переходит на новые имена без слоя совместимости.

**Tech Stack:** Go, JSON-файлы данных сервера, WebSocket/HTTP-протокол сервера, TypeScript-клиент.

---

## Принятые решения

- `ProductionEnergy` полностью заменяет `ProductionBaseTime`.
- Время выполнения задания считается из оставшейся энергии, доступной мощности исполнителей и `ItemModel.Efficiency`.
- Текущая очередь `ConstructorProductionJob` заменяется новой таблицей `Task` сразу, без временного старого формата.
- `Itemtype` переименовывается в `ItemType` в Go-коде, JSON-файлах, HTTP-ответе справочников и TypeScript-типах.
- Начальные акронимы `TaskType`: `ItemProduction`, `ObjectProduction`, `CargoMovement`, `Fueling`, `ItemDeconstruction`, `ObjectDeconstruction`.
- `CargoMovement` используется только для перемещения груза между контейнерами.
- `Fueling` используется для заправки и слива топлива.
- Все задания выполняются оборудованием, указанным в таблице `Implementer`.
- Зарезервированные для задания предметы хранятся в `TaskItemGroup`, чтобы при аварийном завершении сервера предметы не пропадали и не дублировались.
- Если задание отменено или не может продолжаться, резерв из `TaskItemGroup` не возвращается мгновенно. Для возврата создается задание `CargoMovement` в любой свободный контейнер.
- Значения `ProductionEnergy` в существующих данных вычисляются из старого `ProductionBaseTime` и мощности одного малого конструктора. Малый конструктор один в JSON-базе.
- Модель малого конструктора в текущих данных имеет акроним `SmallConstructor`.
- Акронимы типов предметов для исполнителей в текущих данных: `Constructor`, `Deconstructor`, `Robot`, `Container`, `FuelTank`.
- Клиент не показывает энергию задания. Клиент показывает только расчетное время.

## Файлы

- Modify: `server/internal/data/Itemtype.go`
- Move/Rename: `server/internal/data/Itemtype.go` -> `server/internal/data/ItemType.go`
- Modify/Rename: `server/internal/data/Itemtype_test.go`
- Modify: `server/internal/data/EquipmentGroup.go`
- Delete: `server/internal/data/EquipmentGroupRelation.go`
- Delete: `server/internal/data/EquipmentGroupRelation_test.go`
- Create: `server/internal/data/TaskType.go`
- Create: `server/internal/data/TaskType_test.go`
- Create: `server/internal/data/Task.go`
- Create: `server/internal/data/Task_test.go`
- Create: `server/internal/data/TaskItemGroup.go`
- Create: `server/internal/data/TaskItemGroup_test.go`
- Create: `server/internal/data/Implementer.go`
- Create: `server/internal/data/Implementer_test.go`
- Modify: `server/internal/data/ItemModel.go`
- Modify: `server/internal/storage/storage.go`
- Modify: `server/internal/storage/storage_test.go`
- Modify: `server/internal/storage/reference_tables_test.go`
- Modify: `server/internal/ws/reference_data.go`
- Modify: `server/internal/ws/reference_data_test.go`
- Modify: `server/internal/ws/protocol.go`
- Modify: `server/internal/ws/protocol_test.go`
- Modify: `server/internal/ws/hub.go`
- Modify: `server/internal/world/world.go`
- Modify: `server/internal/world/world_test.go`
- Modify: `server/internal/game/types.go`
- Modify: `server/app/space-game-server/main.go`
- Modify/Rename: `server/data/Itemtypes.json` -> `server/data/ItemTypes.json`
- Modify: `server/data/EquipmentGroups.json`
- Delete or stop using: `server/data/EquipmentGroupRelations.json`
- Delete or stop using: `server/data/RelationTypes.json`
- Modify: `server/data/Schemas.json`
- Modify: `server/data/Blueprints.json`
- Modify: `server/data/ItemModels.json`
- Create: `server/data/TaskTypes.json`
- Create: `server/data/Tasks.json`
- Create: `server/data/TaskItemGroups.json`
- Create: `server/data/Implementers.json`
- Modify: `client/src/network/protocol.ts`
- Modify: `client/src/network/referenceData.ts`
- Modify: `client/src/network/referenceData.test.ts`
- Modify: `client/src/network/GameClient.ts`
- Modify: `client/src/network/GameClient.test.ts`
- Modify: `client/src/ui/gameUiState.ts`
- Modify: `client/src/game/GameScene.ts`
- Modify: `client/src/game/controlPanelUsageSelection.ts`
- Modify: `client/src/game/gameUiControlSignature.ts`
- Modify: `client/src/ui/GameUi.tsx`
- Modify: `client/src/**/*.test.ts`

## Task 1: Переименовать типы предметов

- [ ] **Step 1: Обновить серверный тип**

Переименовать `Itemtype` в `ItemType`, `Itemtypes` в `ItemTypes`, конструктор в `NewItemTypes`. JSON-файл должен стать `ItemTypes.json`.

- [ ] **Step 2: Обновить все серверные ссылки**

Заменить `Itemtypes`, `Itemtype`, `NewItemtypes`, `Itemtypes.json` во всех Go-файлах сервера.

- [ ] **Step 3: Обновить клиентский контракт**

В `client/src/network/protocol.ts` заменить `ItemtypeReference` на `ItemTypeReference`, поле `ReferenceDataMessage.Itemtype` на `ItemType`.

- [ ] **Step 4: Обновить клиентские обращения**

Во всех файлах клиента заменить `referenceData.Itemtype` на `referenceData.ItemType`.

- [ ] **Step 5: Проверить тесты**

Run: `go test ./...` из `server` с повышенным разрешением Codex.

Run: `npm test -- --run` из `client`.

Expected: ошибки только по еще не переписанным частям модели.

## Task 2: Встроить связи в EquipmentGroup

- [ ] **Step 1: Расширить тип**

Добавить в `server/internal/data/EquipmentGroup.go` поля `SourceEquipmentGroupID`, `DestinationEquipmentGroupID`, `OppositeEquipmentGroupID`.

- [ ] **Step 2: Удалить отдельную таблицу связей**

Убрать `RelationTypes` и `EquipmentGroupRelations` из `storage.ServerData`, `world.Data`, `game.Snapshot`, `ReferenceDataResponse`.

- [ ] **Step 3: Переписать сохранение выбора связанной группы**

Вместо `EquipmentGroupRelations.Upsert` обновлять одно поле выбранной `EquipmentGroup`:

```go
switch update.RelationTypeAcronym {
case "Source":
	group.SourceEquipmentGroupID = update.RelatedEquipmentGroupID
case "Destination":
	group.DestinationEquipmentGroupID = update.RelatedEquipmentGroupID
case "Opposite":
	group.OppositeEquipmentGroupID = update.RelatedEquipmentGroupID
default:
	return errors.New("unknown equipment group relation")
}
```

- [ ] **Step 4: Переписать чтение связей**

В `controlledEquipmentItemtypeLocked` и связанных методах читать нужный ID прямо из группы оборудования.

- [ ] **Step 5: Обновить клиент**

Удалить `EquipmentGroupRelation` из `SnapshotMessage`, `GameUiState` и `GameScene`. Метод `relatedEquipmentGroupId` должен читать поле группы:

```ts
if (relationTypeAcronym === "Source") return group.SourceEquipmentGroupID || null;
if (relationTypeAcronym === "Destination") return group.DestinationEquipmentGroupID || null;
return group.OppositeEquipmentGroupID || null;
```

## Task 3: Перевести производство на энергию

- [ ] **Step 1: Обновить справочники**

В `BlueprintReference` и `SchemaReference` заменить `ProductionBaseTime` на `ProductionEnergy`. То же сделать в серверных временных структурах `world.go`.

- [ ] **Step 2: Обновить JSON**

В `server/data/Schemas.json` и `server/data/Blueprints.json` заменить ключи на `ProductionEnergy`. Каждое значение вычислить по формуле:

```text
ProductionEnergy = ProductionBaseTime * мощность одного конструктора
```

Мощность одного конструктора взять из `server/data/ItemModels.json` у модели с акронимом `SmallConstructor`.

- [ ] **Step 3: Добавить эффективность модели предмета**

Добавить `Efficiency float64` в `server/internal/data/ItemModel.go`. При нулевом значении в расчетах использовать `1`, чтобы старые записи без поля не давали нулевую работу.

- [ ] **Step 4: Заменить расчет прогресса**

Текущий источник состояния заменить на `Task.RemainingEnergy` и `Task.TotalEnergy`. Для клиента сервер должен передавать или вычислимые поля времени, или данные, из которых клиент без показа энергии рассчитает время. В интерфейсе показывать только время.

## Task 4: Ввести TaskType, Task, TaskItemGroup и Implementer

- [ ] **Step 1: Добавить таблицу TaskType**

Создать тип с полями `ID`, `TitleRu`, `TitleEn`, `Acronym` и индексами по уникальным полям.

- [ ] **Step 2: Добавить таблицу Implementer**

Создать тип с полями `ID`, `TaskTypeID`, `ImplementerEquipmentItemTypeID`, `WorkPart`. Индекс уникальности: `(TaskTypeID, ImplementerEquipmentItemTypeID)`.

- [ ] **Step 3: Добавить таблицу Task**

Создать тип с полями `ID`, `ControllerEquipmentGroupID`, `ParentTaskID`, `TaskTypeID`, `RemainingEnergy`, `TotalEnergy`, `SchemaID`, `BlueprintID`.

- [ ] **Step 4: Добавить таблицу TaskItemGroup**

Создать тип с полями `ID`, `TaskID`, `ItemModelID`, `Count`. Индекс уникальности: `(TaskID, ItemModelID)`. Эта таблица хранит предметы, изъятые из обычных контейнеров и зарезервированные для выполнения задания.

- [ ] **Step 5: Подключить таблицы к storage**

Загружать и сохранять `TaskTypes.json`, `Tasks.json`, `TaskItemGroups.json`, `Implementers.json`.

- [ ] **Step 6: Заменить внутреннюю очередь конструктора**

Удалить серверную модель `ConstructorProductionJob` как источник истины. Создание предметов и объектов должно создавать записи `Task`, а тик мира должен уменьшать `RemainingEnergy`.

- [ ] **Step 7: Использовать Implementer как источник исполнителей**

При выполнении задания находить записи `Implementer` по `TaskTypeID`. Работу выполняют только активные группы оборудования, чей тип предмета совпадает с `ImplementerEquipmentItemTypeID`. Вклад каждой группы считать через `WorkPart`, доступную мощность и `ItemModel.Efficiency`.

- [ ] **Step 8: Резервировать предметы через TaskItemGroup**

При создании задания компоненты, груз или топливо списывать из обычных `ItemGroup` и записывать в `TaskItemGroup`. При успешном завершении задания использовать резерв. При отмене или невозможности продолжить задание не перекладывать резерв мгновенно: создать новое задание `CargoMovement`, которое перемещает зарезервированные предметы в любой свободный контейнер. При загрузке сервера после аварии резерв должен оставаться связанным с незавершенной `Task`.

- [ ] **Step 9: Возвращать резерв через CargoMovement**

Если исходное задание отменено, а исходный контейнер недоступен или не выбран, найти любой контейнер со свободным объемом для всех предметов резерва. Создать `Task` с типом `CargoMovement`, перенести к нему соответствующие `TaskItemGroup` и завершить возврат только после выполнения этого задания роботами.

## Task 5: Обновить протокол и UI

- [ ] **Step 1: Снимок мира**

В `game.Snapshot` удалить `EquipmentGroupRelations` и `ConstructorProductionJobs`, добавить `Tasks []data.Task`.

- [ ] **Step 2: Команды очереди**

Команды конструктора должны принимать `taskId` вместо `jobId`, если клиент работает с задачами напрямую.

- [ ] **Step 3: Отображение очереди**

В `GameUi.tsx` заменить список заданий производства на список `Task`: показывать тип задания, схему или чертеж, расчетное оставшееся время и расчетное полное время. Энергию задания не показывать.

- [ ] **Step 4: Подсказки схем и чертежей**

Оставить текст `"Время"` и показывать расчетное время на основе `ProductionEnergy`, мощности одного конструктора и эффективности исполнителя. Значение `ProductionEnergy` не выводить в интерфейсе.

## Task 6: Миграция данных

- [ ] **Step 1: Проверить сервер перед ручной правкой JSON**

Перед правкой `server/data/*.json` проверить, не запущен ли сервер пользователем. Если запущен, попросить остановить сервер.

- [ ] **Step 2: Перенести связи**

Для каждой записи из `EquipmentGroupRelations.json` найти `EquipmentGroupID` и записать `RelatedEquipmentGroupID` в одно из полей:

- `Source` -> `SourceEquipmentGroupID`
- `Destination` -> `DestinationEquipmentGroupID`
- `Opposite` -> `OppositeEquipmentGroupID`

- [ ] **Step 3: Удалить старые файлы из загрузки**

После переноса сервер больше не должен читать `RelationTypes.json` и `EquipmentGroupRelations.json`.

- [ ] **Step 4: Создать начальные таблицы заданий**

`TaskTypes.json` должен содержать типы:

- `ItemProduction` — изготовление предмета.
- `ObjectProduction` — изготовление космического объекта.
- `CargoMovement` — перемещение груза между контейнерами.
- `Fueling` — заправка и слив топлива.
- `ItemDeconstruction` — деконструкция предметов.
- `ObjectDeconstruction` — деконструкция космических объектов.

`Implementers.json` должен содержать исполнителей:

- `ItemProduction`: `Constructor` с `WorkPart = 0.5`, `Robot` с `WorkPart = 0.5`.
- `ObjectProduction`: `Constructor` с `WorkPart = 0.0`, `Robot` с `WorkPart = 1.0`.
- `ItemDeconstruction`: `Deconstructor` с `WorkPart = 0.5`, `Robot` с `WorkPart = 0.5`.
- `ObjectDeconstruction`: `Deconstructor` с `WorkPart = 0.0`, `Robot` с `WorkPart = 1.0`.
- `CargoMovement`: `Container` с `WorkPart = 0.0`, `Robot` с `WorkPart = 1.0`.
- `Fueling`: `FuelTank` с `WorkPart = 0.0`, `Robot` с `WorkPart = 1.0`.

Значения `WorkPart = 0.0` означают, что тип оборудования участвует как обязательный контроллер или место операции, но сам не добавляет работу к прогрессу.

## Task 7: Проверка

- [ ] **Step 1: Серверные тесты**

Run: `go test ./...` из `server` с повышенным разрешением Codex.

Expected: PASS.

- [ ] **Step 2: Клиентские тесты**

Run: `npm test -- --run` из `client`.

Expected: PASS.

- [ ] **Step 3: Поиск старых имен**

Run: `rg -n "Itemtype|Itemtypes|RelationType|EquipmentGroupRelation|ProductionBaseTime|ConstructorProductionJob" server client --glob '!client/node_modules/**'`.

Expected: нет старых имен, кроме явно объясненных исторических комментариев или тестовых строк, если они нужны.

## Открытые вопросы перед реализацией

- Нет открытых вопросов по плану модели. При реализации уточнять только конкретные случаи, если данные JSON противоречат указанным акронимам или отсутствует свободный контейнер для возврата резерва.
