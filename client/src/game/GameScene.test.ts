import { describe, expect, it, vi } from "vitest";
import type { DockingEventMessage } from "../network/protocol";

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
});
