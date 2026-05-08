import * as Phaser from "phaser";
import "./style.css";
import { GameScene } from "./game/GameScene";
import { installLocalVisualTestAccount } from "./network/localVisualTestAccount";
import { mountGameUi } from "./ui/mountGameUi";

const startGame = async (): Promise<void> => {
  // Локальный визуальный аккаунт должен попасть в хранилище раньше сетевого клиента.
  await installLocalVisualTestAccount();

  const uiRoot = document.getElementById("ui-root");

  if (!uiRoot) {
    throw new Error("ui-root element not found");
  }

  // SolidJS владеет всем текущим UI и получает данные из Phaser-сцены через контроллер.
  const gameUi = mountGameUi(uiRoot);

  // Phaser.Game является корневым объектом клиента и владеет canvas внутри #game-root.
  const game = new Phaser.Game({
    type: Phaser.AUTO,
    parent: "game-root",
    backgroundColor: "#000000",
    scale: {
      mode: Phaser.Scale.RESIZE,
      width: window.innerWidth,
      height: window.innerHeight,
    },
    render: {
      pixelArt: false,
      antialias: true,
    },
    scene: [new GameScene(gameUi.controller)],
  });

  // Глобальная ссылка нужна только для корректного уничтожения Phaser при горячей перезагрузке Vite.
  if (import.meta.hot) {
    import.meta.hot.dispose(() => {
      game.destroy(true);
      gameUi.dispose();
    });
  }
};

void startGame();
