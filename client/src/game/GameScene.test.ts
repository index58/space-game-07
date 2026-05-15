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
});
