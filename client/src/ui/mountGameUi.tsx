import { createComponent, render } from "solid-js/web";
import { GameUi } from "./GameUi";
import { createGameUiController, type GameUiController } from "./gameUiState";

export type MountedGameUi = {
  // Мост для передачи состояния из игрового цикла.
  controller: GameUiController;
  // Освобождает Solid root при горячей перезагрузке.
  dispose: () => void;
};

// Монтирует SolidJS UI в отдельный DOM-контейнер поверх Phaser canvas.
export const mountGameUi = (element: HTMLElement): MountedGameUi => {
  const controller = createGameUiController();
  const dispose = render(() => createComponent(GameUi, { state: controller.state }), element);

  return { controller, dispose };
};
