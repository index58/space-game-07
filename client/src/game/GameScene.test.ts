import { describe, expect, it, vi } from "vitest";
import type { CosmicObject, CosmicObjectModelReference, DockingEventMessage, ExchangeEventMessage } from "../network/protocol";
import type { DrillBeamGeometry } from "./drillBeam";

vi.mock("phaser", () => ({
  Scene: class {
    // Тесту достаточно базового конструктора без запуска движка.
    constructor(_key?: string) {}
  },
  Loader: { Events: { COMPLETE: "complete" } },
}));

describe("GameScene", () => {
  // Проверяет, что текстура привязывается по центру, потому что физическое тело уже смещено внутри полигона.
  it("uses texture center origin for shifted body polygon", async () => {
    const { GameScene } = await import("./GameScene");
    const model = {
      ID: 350,
      CosmicObjectTypeID: 1,
      TextureFilePath: "assets/ships/1024x2048/ship_1024x2048_0042.png",
      TextureWidth: 1024,
      TextureHeight: 2048,
      TextureBodyOriginX: 513,
      TextureBodyOriginY: 920,
      TextureScale: 4,
      BodyWidth: 183.825,
      BodyLength: 348.65,
    } satisfies CosmicObjectModelReference;
    const origins: Array<{ x: number; y: number }> = [];
    type TestSprite = {
      setOrigin: (x: number, y: number) => void;
    };
    const sprite: TestSprite = {
      setOrigin: (x, y) => {
        origins.push({ x, y });
      },
    };
    const scene = Object.create(GameScene.prototype) as {
      referenceData: unknown;
      updateObjectSpriteOrigin: (sprite: TestSprite, object: { CosmicObjectModelID: number }) => void;
    };

    scene.referenceData = {
      CosmicObjectModel: {
        Items: {
          "350": model,
        },
      },
    };

    scene.updateObjectSpriteOrigin(sprite, { CosmicObjectModelID: 350 });

    expect(origins[0]).toEqual({
      x: 0.5,
      y: 0.5,
    });
  });

  // Проверяет, что цвет экранной полоски меняется от зелёного к красному по остаточной броне.
  it("maps armor bar color from green to red", async () => {
    const { GameScene } = await import("./GameScene");
    const scene = Object.create(GameScene.prototype) as {
      armorBarColor: (object: CosmicObject) => number;
    };

    expect(scene.armorBarColor(testCosmicObject({ Armor: 100, MaxArmor: 100 }))).toBe(0x00ff00);
    expect(scene.armorBarColor(testCosmicObject({ Armor: 50, MaxArmor: 100 }))).toBe(0x808000);
    expect(scene.armorBarColor(testCosmicObject({ Armor: 0, MaxArmor: 100 }))).toBe(0xff0000);
  });

  // Проверяет, что экранная полоска показывается только у чужих кораблей и станций.
  it("renders armor bars only for foreign ships and stations", async () => {
    const { GameScene } = await import("./GameScene");
    const scene = Object.create(GameScene.prototype) as {
      referenceData: unknown;
      shouldRenderArmorBar: (object: CosmicObject, selfObject: CosmicObject, model: CosmicObjectModelReference) => boolean;
    };
    scene.referenceData = {
      CosmicObjectType: {
        Items: {
          "1": { Acronym: "Ship" },
          "2": { Acronym: "Station" },
          "3": { Acronym: "Asteroid" },
        },
      },
    };
    const selfObject = testCosmicObject({ ID: 1, OwnerCharacterID: 10 });

    expect(scene.shouldRenderArmorBar(testCosmicObject({ ID: 2, OwnerCharacterID: 20 }), selfObject, testModel({ CosmicObjectTypeID: 1 }))).toBe(true);
    expect(scene.shouldRenderArmorBar(testCosmicObject({ ID: 3, OwnerCharacterID: 20 }), selfObject, testModel({ CosmicObjectTypeID: 2 }))).toBe(true);
    expect(scene.shouldRenderArmorBar(testCosmicObject({ ID: 4, OwnerCharacterID: 20 }), selfObject, testModel({ CosmicObjectTypeID: 3 }))).toBe(false);
    expect(scene.shouldRenderArmorBar(testCosmicObject({ ID: 5, OwnerCharacterID: 10 }), selfObject, testModel({ CosmicObjectTypeID: 1 }))).toBe(false);
    expect(scene.shouldRenderArmorBar(testCosmicObject({ ID: 1, OwnerCharacterID: 10 }), selfObject, testModel({ CosmicObjectTypeID: 1 }))).toBe(false);
  });

  // Проверяет, что буровой луч распознаётся по новой модели космического объекта.
  it("detects drill ray by cosmic object model", async () => {
    const { GameScene } = await import("./GameScene");
    const scene = Object.create(GameScene.prototype) as {
      referenceData: unknown;
      isDrillRayObject: (object: CosmicObject) => boolean;
    };
    scene.referenceData = {
      CosmicObjectModel: {
        Items: {
          "900": testModel({ ID: 900, Acronym: "DrillRay", CosmicObjectTypeID: 6 }),
        },
      },
      CosmicObjectType: {
        Items: {
          "6": { Acronym: "Ray" },
        },
      },
    };

    expect(scene.isDrillRayObject(testCosmicObject({ CosmicObjectModelID: 900 }))).toBe(true);
  });

  // Проверяет, что снаряд распознаётся по типу модели космического объекта.
  it("detects projectile by cosmic object type", async () => {
    const { GameScene } = await import("./GameScene");
    const scene = Object.create(GameScene.prototype) as {
      referenceData: unknown;
      isProjectileObject: (object: CosmicObject) => boolean;
    };
    scene.referenceData = {
      CosmicObjectModel: {
        Items: {
          "901": testModel({ ID: 901, Acronym: "BallisticProjectile", CosmicObjectTypeID: 5 }),
        },
      },
      CosmicObjectType: {
        Items: {
          "5": { Acronym: "Projectile" },
        },
      },
    };

    expect(scene.isProjectileObject(testCosmicObject({ CosmicObjectModelID: 901 }))).toBe(true);
  });

  // Проверяет, что снаряд без текстуры рисуется в векторном слое.
  it("renders projectile objects with graphics layer", async () => {
    const { GameScene } = await import("./GameScene");
    const graphics = createProjectileGraphics();
    const scene = Object.create(GameScene.prototype) as {
      referenceData: unknown;
      projectileGraphics: ProjectileGraphics;
      zoomScale: number;
      renderProjectiles: (objects: CosmicObject[], camera: TestCamera, selfRotation: number) => void;
    };
    scene.referenceData = {
      CosmicObjectModel: {
        Items: {
          "901": testModel({ ID: 901, Acronym: "BallisticProjectile", CosmicObjectTypeID: 5, BodyWidth: 12, BodyLength: 32 }),
        },
      },
      CosmicObjectType: {
        Items: {
          "5": { Acronym: "Projectile" },
        },
      },
    };
    scene.projectileGraphics = graphics;
    scene.zoomScale = 1;

    scene.renderProjectiles([testCosmicObject({ ID: -1, CosmicObjectModelID: 901 })], testCamera({}), 0);

    expect(graphics.clear).toHaveBeenCalled();
    expect(graphics.lineStyle).toHaveBeenCalledTimes(1);
    expect(graphics.strokePath).toHaveBeenCalled();
  });

  // Проверяет, что свободный луч получает полукруглое окончание слоёв без отдельного светового пятна.
  it("does not render drill beam end cap without object hit", async () => {
    const { GameScene } = await import("./GameScene");
    const fillCircle = vi.fn();
    const strokeCircle = vi.fn();
    const graphics = createDrillBeamGraphics({ fillCircle, strokeCircle });
    const scene = Object.create(GameScene.prototype) as {
      pilotToolEffectGraphics: typeof graphics;
      renderDrillBeamGeometry: (geometry: DrillBeamGeometry, timeMs: number) => void;
    };
    scene.pilotToolEffectGraphics = graphics;

    scene.renderDrillBeamGeometry(testDrillBeamGeometry({ hitObject: false }), 1000);

    expect(fillCircle).not.toHaveBeenCalled();
    expect(strokeCircle).not.toHaveBeenCalled();
  });

  // Проверяет, что все видимые слои луча рисуются цельными фигурами с дугой на конце.
  it("renders drill beam visual layers as single shapes with end arcs", async () => {
    const { GameScene } = await import("./GameScene");
    const fillStyle = vi.fn();
    const arc = vi.fn();
    const slice = vi.fn();
    const graphics = createDrillBeamGraphics({ arc, fillStyle, slice });
    const scene = Object.create(GameScene.prototype) as {
      pilotToolEffectGraphics: typeof graphics;
      renderDrillBeamGeometry: (geometry: DrillBeamGeometry, timeMs: number) => void;
    };
    scene.pilotToolEffectGraphics = graphics;

    scene.renderDrillBeamGeometry(testDrillBeamGeometry({ hitObject: false }), 1000);

    expect(fillStyle).toHaveBeenCalledWith(0x1fdcff, expect.any(Number));
    expect(fillStyle).toHaveBeenCalledWith(0x0b6f9e, expect.any(Number));
    expect(fillStyle).toHaveBeenCalledWith(0x20d8ff, expect.any(Number));
    expect(fillStyle).toHaveBeenCalledWith(0x8cf8ff, 0.58);
    expect(fillStyle).toHaveBeenCalledWith(0xffffff, 0.95);
    expect(slice).not.toHaveBeenCalled();
    expect(arc).toHaveBeenCalledWith(0, 0, 7.199999999999999, -Math.PI, 0);
    expect(arc).toHaveBeenCalledWith(0, 0, 13.5, -Math.PI, 0);
    expect(arc).toHaveBeenCalledWith(0, 0, 8.25, -Math.PI, 0);
    expect(arc).toHaveBeenCalledWith(0, 0, 3.3000000000000003, -Math.PI, 0);
    expect(arc).toHaveBeenCalledWith(0, 0, 0.9750000000000001, -Math.PI, 0);
  });

  // Проверяет, что конец луча получает световое пятно только при попадании в объект.
  it("renders drill beam end cap on object hit", async () => {
    const { GameScene } = await import("./GameScene");
    const fillCircle = vi.fn();
    const strokeCircle = vi.fn();
    const graphics = createDrillBeamGraphics({ fillCircle, strokeCircle });
    const scene = Object.create(GameScene.prototype) as {
      pilotToolEffectGraphics: typeof graphics;
      renderDrillBeamGeometry: (geometry: DrillBeamGeometry, timeMs: number) => void;
    };
    scene.pilotToolEffectGraphics = graphics;

    scene.renderDrillBeamGeometry(testDrillBeamGeometry({ hitObject: true }), 1000);

    expect(fillCircle).toHaveBeenCalledTimes(2);
    expect(fillCircle).toHaveBeenCalledWith(0, 0, 5.4);
    expect(strokeCircle).toHaveBeenCalledTimes(1);
  });

  // Проверяет, что маленькие уведомления стыковки убираются через пять секунд.
  it("keeps docking notifications for five seconds", async () => {
    const { GameScene } = await import("./GameScene");
    const scene = Object.create(GameScene.prototype) as {
      dockingNotifications: Array<{ expiresAtMs: number }>;
      nextDockingNotificationID: number;
      applyDockingEvent: (event: DockingEventMessage, nowMs: number) => void;
    };

    scene.dockingNotifications = [];
    scene.nextDockingNotificationID = 1;
    scene.applyDockingEvent({
      type: "dockingEvent",
      kind: "dockingNotification",
      message: "Объект пристыкован",
    }, 1000);

    expect(scene.dockingNotifications[0].expiresAtMs).toBe(6000);
  });

  // Проверяет, что запрос обмена без длительности от сервера всё равно показывает обратную полоску.
  it("uses default exchange request duration when server omits it", async () => {
    const { GameScene } = await import("./GameScene");
    const scene = Object.create(GameScene.prototype) as {
      dockingWindow: { kind: string; role: string; startedAtMs: number; durationMs: number } | null;
      exchangeRequestRole: "sender" | "receiver" | null;
      applyExchangeEvent: (event: ExchangeEventMessage, nowMs: number) => void;
    };

    scene.dockingWindow = null;
    scene.exchangeRequestRole = null;

    scene.applyExchangeEvent({
      type: "exchangeEvent",
      kind: "exchangeRequestStarted",
      role: "receiver",
    }, 1234);

    expect(scene.dockingWindow).toEqual({
      kind: "exchangeRequest",
      role: "receiver",
      startedAtMs: 1234,
      durationMs: 10000,
    });
  });

  // Проверяет, что выбранная строка очереди деконструкции не сбрасывается как неподходящее задание конструктора.
  it("keeps selected deconstruction queue task", async () => {
    const { GameScene } = await import("./GameScene");
    const scene = Object.create(GameScene.prototype) as {
      referenceData: unknown;
      selectedControlPanelConstructorMainJobId: number | null;
      selectedControlPanelUsageRightEquipmentGroupId: number | null;
      syncControlPanelConstructorMainJobSelection: (jobs: unknown[], tasks: unknown[]) => void;
    };

    scene.referenceData = {
      TaskType: {
        Items: {
          "3": { ID: 3, Acronym: "ItemDeconstruction" },
        },
      },
    };
    scene.selectedControlPanelConstructorMainJobId = 42;
    scene.selectedControlPanelUsageRightEquipmentGroupId = 9;

    scene.syncControlPanelConstructorMainJobSelection([], [{
      ID: 42,
      ControllerEquipmentGroupID: 9,
      ParentTaskID: 0,
      TaskTypeID: 3,
    }]);

    expect(scene.selectedControlPanelConstructorMainJobId).toBe(42);
  });

  // Проверяет, что одна выбранная строка деконструкции открывает окно количества без немедленной отправки.
  it("opens amount dialog for one selected deconstruction item group", async () => {
    const { GameScene } = await import("./GameScene");
    const sendControlPanelItemDeconstruction = vi.fn();
    const setControlPanelFuelDrainAmount = vi.fn();
    const scene = Object.create(GameScene.prototype) as {
      selectedControlPanelUsageRightEquipmentGroupId: number | null;
      selectedControlPanelConstructorMaterialContainerGroupId: number | null;
      selectedControlPanelUsageLeftContainerGroupId: number | null;
      selectedControlPanelUsageRightItemGroupIds: number[];
      controlPanelItemDeconstructionDialogOpen: boolean;
      controlPanelItemDeconstructionMaxAmount: number;
      controlPanelFuelDrainAmount: number;
      gameUi: { state: () => { itemGroups: Array<{ ID: number; ContainerEquipmentGroupID: number; Count: number }> } };
      inputController: { setControlPanelFuelDrainAmount: (value: number) => void };
      gameClient: { sendControlPanelItemDeconstruction: typeof sendControlPanelItemDeconstruction };
      startControlPanelItemDeconstruction: () => void;
    };

    scene.selectedControlPanelUsageRightEquipmentGroupId = 11;
    scene.selectedControlPanelConstructorMaterialContainerGroupId = 22;
    scene.selectedControlPanelUsageLeftContainerGroupId = 33;
    scene.selectedControlPanelUsageRightItemGroupIds = [44];
    scene.controlPanelItemDeconstructionDialogOpen = false;
    scene.controlPanelItemDeconstructionMaxAmount = 0;
    scene.controlPanelFuelDrainAmount = 0;
    scene.gameUi = { state: () => ({ itemGroups: [{ ID: 44, ContainerEquipmentGroupID: 22, Count: 7 }] }) };
    scene.inputController = { setControlPanelFuelDrainAmount };
    scene.gameClient = { sendControlPanelItemDeconstruction };

    scene.startControlPanelItemDeconstruction();

    expect(scene.controlPanelItemDeconstructionDialogOpen).toBe(true);
    expect(scene.controlPanelItemDeconstructionMaxAmount).toBe(7);
    expect(scene.controlPanelFuelDrainAmount).toBe(7);
    expect(setControlPanelFuelDrainAmount).toHaveBeenCalledWith(7);
    expect(sendControlPanelItemDeconstruction).not.toHaveBeenCalled();
  });

  // Проверяет, что подтверждение окна количества отправляет деконструкцию с выбранным количеством.
  it("sends selected amount for item deconstruction dialog", async () => {
    const { GameScene } = await import("./GameScene");
    const sendControlPanelItemDeconstruction = vi.fn();
    const scene = Object.create(GameScene.prototype) as {
      controlPanelItemDeconstructionDialogOpen: boolean;
      controlPanelItemDeconstructionMaxAmount: number;
      controlPanelItemDeconstructionDeconstructorGroupId: number | null;
      controlPanelItemDeconstructionSourceGroupId: number | null;
      controlPanelItemDeconstructionTargetGroupId: number | null;
      controlPanelItemDeconstructionItemGroupIds: number[];
      controlPanelFuelDrainAmount: number;
      inputController: {
        getControlPanelFuelDrainAmount: (fallback: number) => number;
        blurControlPanelFuelDrainAmount: () => void;
      };
      gameClient: { sendControlPanelItemDeconstruction: typeof sendControlPanelItemDeconstruction };
      consumeControlPanelUiAction: (action: unknown) => boolean;
    };

    scene.controlPanelItemDeconstructionDialogOpen = true;
    scene.controlPanelItemDeconstructionMaxAmount = 9;
    scene.controlPanelItemDeconstructionDeconstructorGroupId = 11;
    scene.controlPanelItemDeconstructionSourceGroupId = 22;
    scene.controlPanelItemDeconstructionTargetGroupId = 33;
    scene.controlPanelItemDeconstructionItemGroupIds = [44];
    scene.controlPanelFuelDrainAmount = 9;
    scene.inputController = {
      getControlPanelFuelDrainAmount: () => 3,
      blurControlPanelFuelDrainAmount: vi.fn(),
    };
    scene.gameClient = { sendControlPanelItemDeconstruction };

    const handled = scene.consumeControlPanelUiAction({ type: "click", kind: "button", controlId: "control-panel-item-deconstruction-ok" });

    expect(handled).toBe(true);
    expect(scene.controlPanelItemDeconstructionDialogOpen).toBe(false);
    expect(sendControlPanelItemDeconstruction).toHaveBeenCalledWith({
      deconstructorEquipmentGroupId: 11,
      sourceContainerEquipmentGroupId: 22,
      targetContainerEquipmentGroupId: 33,
      itemGroupIds: [44],
      amount: 3,
    });
  });

  // Проверяет, что кнопка обмена открывает окно количества для одной выбранной строки.
  it("opens amount dialog for one selected exchange source item", async () => {
    const { GameScene } = await import("./GameScene");
    const sendExchangeAddItems = vi.fn();
    const setControlPanelFuelDrainAmount = vi.fn();
    const scene = Object.create(GameScene.prototype) as {
      exchangeState: {
        selfObjectId: number;
        selfSourceContainerEquipmentGroupId: number;
        otherConfirmed: boolean;
      } | null;
      selectedExchangeSourceObjectId: number | null;
      selectedExchangeSourceItemGroupIds: number[];
      controlPanelContainerTransferDialogOpen: boolean;
      controlPanelContainerTransferMaxAmount: number;
      controlPanelFuelDrainAmount: number;
      controlPanelContainerTransferItemGroupIds: number[];
      controlPanelContainerTransferSourceGroupId: number | null;
      gameUi: { state: () => {
        equipmentGroups: Array<{ ID: number; CosmicObjectID: number; EquipmentItemModelID: number }>;
        itemGroups: Array<{ ID: number; ContainerEquipmentGroupID: number; Count: number }>;
        referenceData: { ItemModel: { Items: Record<string, { ItemTypeID: number }> }; ItemType: { Items: Record<string, { Acronym: string }> } };
      } };
      inputController: { setControlPanelFuelDrainAmount: (value: number) => void };
      gameClient: { sendExchangeAddItems: typeof sendExchangeAddItems };
      consumeExchangeUiAction: (action: unknown) => boolean;
    };

    scene.exchangeState = { selfObjectId: 1, selfSourceContainerEquipmentGroupId: 0, otherConfirmed: false };
    scene.selectedExchangeSourceObjectId = 1;
    scene.selectedExchangeSourceItemGroupIds = [44];
    scene.controlPanelContainerTransferDialogOpen = false;
    scene.controlPanelContainerTransferMaxAmount = 0;
    scene.controlPanelFuelDrainAmount = 0;
    scene.controlPanelContainerTransferItemGroupIds = [];
    scene.controlPanelContainerTransferSourceGroupId = null;
    scene.gameUi = { state: () => ({
      equipmentGroups: [{ ID: 22, CosmicObjectID: 1, EquipmentItemModelID: 301 }],
      itemGroups: [{ ID: 44, ContainerEquipmentGroupID: 22, Count: 7 }],
      referenceData: {
        ItemModel: { Items: { "301": { ItemTypeID: 1 } } },
        ItemType: { Items: { "1": { Acronym: "Container" } } },
      },
    }) };
    scene.inputController = { setControlPanelFuelDrainAmount };
    scene.gameClient = { sendExchangeAddItems };

    const handled = scene.consumeExchangeUiAction({ type: "click", kind: "button", controlId: "exchange-move-to-queue-button" });

    expect(handled).toBe(true);
    expect(scene.controlPanelContainerTransferDialogOpen).toBe(true);
    expect(scene.controlPanelContainerTransferMaxAmount).toBe(7);
    expect(scene.controlPanelFuelDrainAmount).toBe(7);
    expect(scene.controlPanelContainerTransferItemGroupIds).toEqual([44]);
    expect(scene.controlPanelContainerTransferSourceGroupId).toBe(22);
    expect(setControlPanelFuelDrainAmount).toHaveBeenCalledWith(7);
    expect(sendExchangeAddItems).not.toHaveBeenCalled();
  });

  // Проверяет, что список источника обмена поддерживает добавление строк через Ctrl.
  it("adds exchange source rows to selection with ctrl", async () => {
    const { GameScene } = await import("./GameScene");
    const scene = Object.create(GameScene.prototype) as {
      exchangeState: {
        selfObjectId: number;
        selfSourceContainerEquipmentGroupId: number;
        otherConfirmed: boolean;
      } | null;
      selectedExchangeSourceItemGroupIds: number[];
      selectedExchangeSourceAnchorItemGroupId: number | null;
      gameUi: { state: () => {
        itemGroups: Array<{ ID: number; ContainerEquipmentGroupID: number }>;
      } };
      consumeExchangeUiAction: (action: unknown) => boolean;
    };

    scene.exchangeState = { selfObjectId: 1, selfSourceContainerEquipmentGroupId: 22, otherConfirmed: false };
    scene.selectedExchangeSourceItemGroupIds = [44];
    scene.selectedExchangeSourceAnchorItemGroupId = 44;
    scene.gameUi = { state: () => ({
      itemGroups: [
        { ID: 44, ContainerEquipmentGroupID: 22 },
        { ID: 45, ContainerEquipmentGroupID: 22 },
      ],
    }) };

    const handled = scene.consumeExchangeUiAction({
      type: "click",
      kind: "list",
      controlId: "exchange-source-list-45",
      value: "45",
      ctrlKey: true,
    });

    expect(handled).toBe(true);
    expect(scene.selectedExchangeSourceItemGroupIds).toEqual([44, 45]);
    expect(scene.selectedExchangeSourceAnchorItemGroupId).toBe(45);
  });

  // Проверяет, что после двух подтверждений кнопка отмены не отправляет отмену обмена на сервер.
  it("does not send exchange cancel while items are moving", async () => {
    const { GameScene } = await import("./GameScene");
    const sendExchangeCancel = vi.fn();
    const scene = Object.create(GameScene.prototype) as {
      exchangeState: {
        selfConfirmed: boolean;
        otherConfirmed: boolean;
      } | null;
      gameClient: { sendExchangeCancel: typeof sendExchangeCancel };
      consumeExchangeUiAction: (action: unknown) => boolean;
    };

    scene.exchangeState = { selfConfirmed: true, otherConfirmed: true };
    scene.gameClient = { sendExchangeCancel };

    const handled = scene.consumeExchangeUiAction({ type: "click", kind: "button", controlId: "exchange-cancel-button" });

    expect(handled).toBe(true);
    expect(sendExchangeCancel).not.toHaveBeenCalled();
  });

  // Проверяет, что списки предметов и очереди обмена подключены к общей прокрутке списков.
  it("adds exchange item lists to shared scroll states", async () => {
    const { GameScene } = await import("./GameScene");
    const receiverElement = document.createElement("div");
    const sourceElement = document.createElement("div");
    const otherQueueElement = document.createElement("div");
    const selfQueueElement = document.createElement("div");
    receiverElement.id = "exchange-receiver-list";
    sourceElement.id = "exchange-source-list";
    otherQueueElement.id = "exchange-other-queue";
    selfQueueElement.id = "exchange-self-queue";
    receiverElement.getBoundingClientRect = () => ({ width: 200, height: 80, left: 0, top: 0, right: 200, bottom: 80, x: 0, y: 0, toJSON: () => ({}) });
    sourceElement.getBoundingClientRect = () => ({ width: 200, height: 80, left: 0, top: 0, right: 200, bottom: 80, x: 0, y: 0, toJSON: () => ({}) });
    otherQueueElement.getBoundingClientRect = () => ({ width: 200, height: 80, left: 0, top: 0, right: 200, bottom: 80, x: 0, y: 0, toJSON: () => ({}) });
    selfQueueElement.getBoundingClientRect = () => ({ width: 200, height: 80, left: 0, top: 0, right: 200, bottom: 80, x: 0, y: 0, toJSON: () => ({}) });
    document.body.append(receiverElement, sourceElement, otherQueueElement, selfQueueElement);
    const scene = Object.create(GameScene.prototype) as {
      gameUi: { state: () => {
        exchangeState: {
          selfReceiverContainerEquipmentGroupId: number;
          selfSourceContainerEquipmentGroupId: number;
        };
        itemGroups: Array<{ ID: number; ContainerEquipmentGroupID: number }>;
      } };
      exchangeState: {
        selfReceiverContainerEquipmentGroupId: number;
        selfSourceContainerEquipmentGroupId: number;
        selfQueue: unknown[];
        otherQueue: unknown[];
      };
      listScrollOffsets: Map<string, number>;
      listScrollbarDrags: Map<string, unknown>;
      getListScrollStates: () => Record<string, { visible: boolean; contentOffsetPx: number }>;
    };

    scene.gameUi = { state: () => ({
      exchangeState: {
        selfReceiverContainerEquipmentGroupId: 10,
        selfSourceContainerEquipmentGroupId: 11,
      },
      itemGroups: [
        ...Array.from({ length: 20 }, (_, index) => ({ ID: index + 1, ContainerEquipmentGroupID: 10 })),
        ...Array.from({ length: 20 }, (_, index) => ({ ID: index + 101, ContainerEquipmentGroupID: 11 })),
      ],
    }) };
    scene.listScrollOffsets = new Map([
      ["exchange-receiver-list", 10],
      ["exchange-source-list", 20],
      ["exchange-other-queue", 30],
      ["exchange-self-queue", 40],
    ]);
    scene.listScrollbarDrags = new Map();
    scene.exchangeState = {
      selfReceiverContainerEquipmentGroupId: 10,
      selfSourceContainerEquipmentGroupId: 11,
      otherQueue: Array.from({ length: 20 }),
      selfQueue: Array.from({ length: 20 }),
    };

    const scrollStates = scene.getListScrollStates();

    expect(scrollStates["exchange-receiver-list"].visible).toBe(true);
    expect(scrollStates["exchange-receiver-list"].contentOffsetPx).toBe(10);
    expect(scrollStates["exchange-source-list"].visible).toBe(true);
    expect(scrollStates["exchange-source-list"].contentOffsetPx).toBe(20);
    expect(scrollStates["exchange-other-queue"].visible).toBe(true);
    expect(scrollStates["exchange-other-queue"].contentOffsetPx).toBe(30);
    expect(scrollStates["exchange-self-queue"].visible).toBe(true);
    expect(scrollStates["exchange-self-queue"].contentOffsetPx).toBe(40);
  });

  // Проверяет, что подтверждение количества отправляет выбранное число предметов в очередь обмена.
  it("sends selected exchange amount after dialog confirmation", async () => {
    const { GameScene } = await import("./GameScene");
    const sendExchangeSelectSource = vi.fn();
    const sendExchangeAddItems = vi.fn();
    const blurControlPanelFuelDrainAmount = vi.fn();
    const scene = Object.create(GameScene.prototype) as {
      exchangeState: {
        selfObjectId: number;
        selfSourceContainerEquipmentGroupId: number;
        otherConfirmed: boolean;
      } | null;
      controlPanelContainerTransferDialogOpen: boolean;
      controlPanelContainerTransferMaxAmount: number;
      controlPanelFuelDrainAmount: number;
      controlPanelContainerTransferItemGroupIds: number[];
      controlPanelContainerTransferSourceGroupId: number | null;
      inputController: {
        getControlPanelFuelDrainAmount: (fallback: number) => number;
        blurControlPanelFuelDrainAmount: () => void;
      };
      gameClient: {
        sendExchangeSelectSource: typeof sendExchangeSelectSource;
        sendExchangeAddItems: typeof sendExchangeAddItems;
      };
      consumeExchangeUiAction: (action: unknown) => boolean;
    };

    scene.exchangeState = { selfObjectId: 1, selfSourceContainerEquipmentGroupId: 0, otherConfirmed: false };
    scene.controlPanelContainerTransferDialogOpen = true;
    scene.controlPanelContainerTransferMaxAmount = 7;
    scene.controlPanelFuelDrainAmount = 7;
    scene.controlPanelContainerTransferItemGroupIds = [44];
    scene.controlPanelContainerTransferSourceGroupId = 22;
    scene.inputController = {
      getControlPanelFuelDrainAmount: () => 3,
      blurControlPanelFuelDrainAmount,
    };
    scene.gameClient = { sendExchangeSelectSource, sendExchangeAddItems };

    const handled = scene.consumeExchangeUiAction({ type: "click", kind: "button", controlId: "exchange-add-items-ok" });

    expect(handled).toBe(true);
    expect(scene.controlPanelContainerTransferDialogOpen).toBe(false);
    expect(sendExchangeSelectSource).toHaveBeenCalledWith(22);
    expect(sendExchangeAddItems).toHaveBeenCalledWith([44], 3);
    expect(blurControlPanelFuelDrainAmount).toHaveBeenCalled();
  });
});

type DrillBeamGraphics = {
  // Выбирает цвет и прозрачность следующей заливки.
  fillStyle: ReturnType<typeof vi.fn>;
  // Начинает новый векторный путь.
  beginPath: ReturnType<typeof vi.fn>;
  // Переносит текущую точку пути.
  moveTo: ReturnType<typeof vi.fn>;
  // Добавляет прямой участок пути.
  lineTo: ReturnType<typeof vi.fn>;
  // Замыкает текущий путь.
  closePath: ReturnType<typeof vi.fn>;
  // Заливает текущий путь.
  fillPath: ReturnType<typeof vi.fn>;
  // Выбирает цвет, прозрачность и толщину обводки.
  lineStyle: ReturnType<typeof vi.fn>;
  // Обводит текущий путь.
  strokePath: ReturnType<typeof vi.fn>;
  // Рисует залитый круг.
  fillCircle: ReturnType<typeof vi.fn>;
  // Рисует обводку круга.
  strokeCircle: ReturnType<typeof vi.fn>;
  // Рисует залитый сектор круга.
  slice: ReturnType<typeof vi.fn>;
  // Добавляет дугу к текущему пути.
  arc: ReturnType<typeof vi.fn>;
};

type ProjectileGraphics = {
  // Очищает предыдущий кадр векторного слоя.
  clear: ReturnType<typeof vi.fn>;
  // Выбирает цвет, прозрачность и толщину линии.
  lineStyle: ReturnType<typeof vi.fn>;
  // Начинает новый векторный путь.
  beginPath: ReturnType<typeof vi.fn>;
  // Переносит текущую точку пути.
  moveTo: ReturnType<typeof vi.fn>;
  // Добавляет прямой участок пути.
  lineTo: ReturnType<typeof vi.fn>;
  // Обводит текущий путь.
  strokePath: ReturnType<typeof vi.fn>;
};

type TestCamera = ReturnType<typeof testCamera>;

const createDrillBeamGraphics = (overrides: Partial<DrillBeamGraphics> = {}): DrillBeamGraphics => ({
  fillStyle: vi.fn(),
  beginPath: vi.fn(),
  moveTo: vi.fn(),
  lineTo: vi.fn(),
  closePath: vi.fn(),
  fillPath: vi.fn(),
  lineStyle: vi.fn(),
  strokePath: vi.fn(),
  fillCircle: vi.fn(),
  strokeCircle: vi.fn(),
  slice: vi.fn(),
  arc: vi.fn(),
  ...overrides,
});

const createProjectileGraphics = (overrides: Partial<ProjectileGraphics> = {}): ProjectileGraphics => ({
  clear: vi.fn(),
  lineStyle: vi.fn(),
  beginPath: vi.fn(),
  moveTo: vi.fn(),
  lineTo: vi.fn(),
  strokePath: vi.fn(),
  ...overrides,
});

const testCamera = (overrides: Partial<{
  shipPosition: { x: number; y: number };
  shipRotation: number;
  zoom: number;
  viewportWidth: number;
  viewportHeight: number;
}>): {
  shipPosition: { x: number; y: number };
  shipRotation: number;
  zoom: number;
  viewportWidth: number;
  viewportHeight: number;
} => ({
  shipPosition: { x: 0, y: 0 },
  shipRotation: 0,
  zoom: 1,
  viewportWidth: 800,
  viewportHeight: 600,
  ...overrides,
});

const testDrillBeamGeometry = (input: { hitObject: boolean }): DrillBeamGeometry => ({
  start: { x: 0, y: 100 },
  end: { x: 0, y: 0 },
  lengthPx: 100,
  widthPx: 3,
  hitObject: input.hitObject,
});

const testCosmicObject = (overrides: Partial<CosmicObject>): CosmicObject => ({
  ID: 1,
  Title: "Object",
  CosmicObjectModelID: 1,
  OwnerCharacterID: 0,
  OwnerNpcClanID: 0,
  CreatorCharacterID: 0,
  X: 0,
  Y: 0,
  Rotation: 0,
  TargetRotation: 0,
  Speed: 0,
  VelocityX: 0,
  VelocityY: 0,
  AngularSpeed: 0,
  Mass: 1,
  Capacity: 0,
  Fuel: 0,
  MaxFuel: 0,
  MaxSpeed: 0,
  MaxAngularSpeed: 0,
  MaxAlongForce: 0,
  MaxAcrossForce: 0,
  MaxTorque: 0,
  GeneratingPower: 0,
  ConsumingPower: 0,
  AlongForce: 0,
  AcrossForce: 0,
  Torque: 0,
  MaxArmor: 100,
  Armor: 100,
  LastReceivedDamageTime: 0,
  Enabled: true,
  Anchored: false,
  Complexity: 0,
  OccupiedVolume: 0,
  OwnerName: "",
  ...overrides,
});

const testModel = (overrides: Partial<CosmicObjectModelReference>): CosmicObjectModelReference => ({
  ID: 1,
  TitleRu: "Object",
  TitleEn: "Object",
  Acronym: "Object",
  CosmicObjectTypeID: 1,
  TextureFilePath: "",
  TextureWidth: 1,
  TextureHeight: 1,
  TextureBodyOriginX: 0,
  TextureBodyOriginY: 0,
  TextureBodyWidth: 30,
  TextureBodyLength: 30,
  TextureScale: 1,
  BodyWidth: 30,
  BodyLength: 30,
  MaxArmor: 100,
  ...overrides,
});
