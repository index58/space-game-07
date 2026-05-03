import { formatNumber } from "../domain/format";
import type { ConnectionStatus, CosmicObject } from "../network/protocol";

export type DebugOverlayView = {
  // Состояние сетевого подключения.
  status: ConnectionStatus;
  // Посещаемый объект игрока, если он уже получен.
  selfObject: CosmicObject | null;
  // Путь к текстуре текущего объекта.
  textureFilePath: string | null;
  // Текущая частота кадров.
  fps: number;
  // Рассчитанный масштаб камеры.
  zoom: number;
};

// Готовит строки отладочной панели для Solid-компонента.
export const getDebugOverlayLines = (view: DebugOverlayView): string[] => {
  if (!view.selfObject) {
    return [
      `Статус: ${view.status}`,
      "Ожидание подключения к серверу",
      `Зум: ${formatNumber(view.zoom, 2)}`,
      `FPS: ${formatNumber(view.fps, 0)}`,
    ];
  }

  const speed = Math.hypot(view.selfObject.VelocityX, view.selfObject.VelocityY);

  return [
    `Статус: ${view.status}`,
    `ID своего объекта: ${view.selfObject.ID}`,
    `Файл объекта: ${view.textureFilePath ?? "неизвестно"}`,
    `X: ${formatNumber(view.selfObject.X)} м`,
    `Y: ${formatNumber(view.selfObject.Y)} м`,
    `Скорость: ${formatNumber(speed)} м/с`,
    `Угол: ${formatNumber(view.selfObject.Rotation, 4)} рад`,
    `Угл. скорость: ${formatNumber(view.selfObject.AngularSpeed, 4)} рад/с`,
    `Зум: ${formatNumber(view.zoom, 2)}`,
    `FPS: ${formatNumber(view.fps, 0)}`,
  ];
};
