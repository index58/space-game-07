import { readFileSync } from "node:fs";
import { describe, expect, it } from "vitest";

describe("technical specification UI Kit section", () => {
  // Проверяет, что ТЗ закрепляет единый UI Kit и содержит ссылки на ключевые файлы.
  it("documents the shared game UI kit entry points", () => {
    const content = readFileSync("../specifications/technical.md", "utf8");

    expect(content).toContain("Единый игровой UI Kit");
    expect(content).toContain("client/src/ui-kit/types.ts");
    expect(content).toContain("client/src/ui-kit/runtime.ts");
    expect(content).toContain("client/src/ui-kit/textEdit.ts");
    expect(content).toContain("client/src/ui-kit/components.tsx");
    expect(content).toContain("client/src/game/InputController.ts");
  });
});
