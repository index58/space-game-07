import * as Phaser from "phaser";
import "./style.css";
import { GameScene } from "./game/GameScene";

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
  scene: [GameScene],
});

// Глобальная ссылка нужна только для корректного уничтожения Phaser при горячей перезагрузке Vite.
if (import.meta.hot) {
  import.meta.hot.dispose(() => game.destroy(true));
}
