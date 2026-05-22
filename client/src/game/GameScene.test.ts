import { describe, expect, it, vi } from "vitest";
import type { DockingEventMessage, ExchangeEventMessage } from "../network/protocol";

vi.mock("phaser", () => ({
  Scene: class {
    // Тесту достаточно базового конструктора без запуска движка.
    constructor(_key?: string) {}
  },
  Loader: { Events: { COMPLETE: "complete" } },
}));

describe("GameScene", () => {
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
