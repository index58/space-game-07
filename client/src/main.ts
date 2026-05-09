import * as Phaser from "phaser";
import "./style.css";
import { GameScene } from "./game/GameScene";
import { mountGameUi } from "./ui/mountGameUi";
import { syncGameUiViewportScale } from "./ui/gameViewportScale";

const startGame = async (): Promise<void> => {
  const gameRoot = document.getElementById("game-root");
  const uiRoot = document.getElementById("ui-root");

  if (!gameRoot) {
    throw new Error("game-root element not found");
  }
  if (!uiRoot) {
    throw new Error("ui-root element not found");
  }

  // SolidJS владеет всем текущим UI и получает данные из Phaser-сцены через контроллер.
  const gameUi = mountGameUi(uiRoot);
  const updateGameUiViewportScale = (): void => {
    syncGameUiViewportScale(gameRoot, uiRoot);
  };
  updateGameUiViewportScale();
  window.addEventListener("resize", updateGameUiViewportScale);

  // Phaser.Game является корневым объектом клиента и владеет canvas внутри #game-root.
  const gameViewport = gameRoot.getBoundingClientRect();
  const game = new Phaser.Game({
    type: Phaser.AUTO,
    parent: "game-root",
    backgroundColor: "#000000",
    scale: {
      mode: Phaser.Scale.RESIZE,
      width: Math.max(1, gameViewport.width),
      height: Math.max(1, gameViewport.height),
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
      window.removeEventListener("resize", updateGameUiViewportScale);
      game.destroy(true);
      gameUi.dispose();
    });
  }
};

void startGame();
