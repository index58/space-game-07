const FRAME_RATE_WINDOW_MS = 1000;

// Считает простую среднюю частоту кадров по отметкам времени за последнюю секунду.
export class FrameRateMeter {
  // Временные отметки кадров, которые еще попадают в измерительное окно.
  private frameTimesMs: number[] = [];

  // Регистрирует новый кадр и возвращает среднюю частоту за последнюю секунду.
  recordFrame(timeMs: number): number {
    this.frameTimesMs.push(timeMs);
    const minTimeMs = timeMs - FRAME_RATE_WINDOW_MS;
    while (this.frameTimesMs.length > 0 && this.frameTimesMs[0] < minTimeMs) {
      this.frameTimesMs.shift();
    }

    if (this.frameTimesMs.length < 2) {
      return this.frameTimesMs.length;
    }

    const elapsedMs = this.frameTimesMs[this.frameTimesMs.length - 1] - this.frameTimesMs[0];
    return elapsedMs > 0 ? (this.frameTimesMs.length - 1) * 1000 / elapsedMs : 0;
  }
}
