import { describe, expect, it } from "vitest";
import { FrameRateMeter } from "./frameRateMeter";

describe("FrameRateMeter", () => {
  // Проверяет, что частота считается как среднее по кадрам за последнюю секунду без сглаживания со старыми значениями.
  it("measures average frame rate over the latest second", () => {
    const meter = new FrameRateMeter();
    let fps = 0;

    for (let timeMs = 0; timeMs <= 1000; timeMs += 100) {
      fps = meter.recordFrame(timeMs);
    }

    expect(fps).toBe(10);
  });

  // Проверяет, что старые кадры за пределами секунды перестают влиять на показания.
  it("drops frames older than one second", () => {
    const meter = new FrameRateMeter();

    for (let timeMs = 0; timeMs <= 1000; timeMs += 100) {
      meter.recordFrame(timeMs);
    }
    const fps = meter.recordFrame(2000);

    expect(fps).toBe(1);
  });
});
