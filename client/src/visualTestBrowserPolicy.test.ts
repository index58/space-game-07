import { existsSync, readdirSync, readFileSync, statSync } from "node:fs";
import { describe, expect, it } from "vitest";

const rootAgents = readFileSync("../AGENTS.md", "utf8");
const mainSource = readFileSync("src/main.ts", "utf8");
const helperPath = "visual-tests/visualTestBrowser.cjs";
const helperSource = () => readFileSync(helperPath, "utf8");

// Собирает исходные файлы проекта, чтобы найти обходы единой браузерной обвязки.
const projectFiles = (directory: string): string[] => {
  const result: string[] = [];
  for (const entry of readdirSync(directory)) {
    if (entry === "node_modules" || entry === "dist") {
      continue;
    }
    const path = `${directory}/${entry}`;
    const stat = statSync(path);
    if (stat.isDirectory()) {
      result.push(...projectFiles(path));
      continue;
    }
    if (/\.(?:ts|tsx|js|cjs|md)$/.test(entry)) {
      result.push(path);
    }
  }
  return result;
};

describe("visual test browser policy", () => {
  // Проверяет, что обычный игровой клиент не подхватывает локальный тестовый аккаунт.
  it("does not install visual test token during regular startup", () => {
    expect(mainSource).not.toContain("localVisualTestAccount");
    expect(mainSource).not.toContain("installLocalVisualTestAccount");
    expect(existsSync("src/network/localVisualTestAccount.ts")).toBe(false);
    expect(existsSync("src/network/localVisualTestAccount.test.ts")).toBe(false);
  });

  // Проверяет, что визуальные проверки имеют единую браузерную обвязку с явной установкой тестового токена.
  it("uses a shared playwright helper that injects and verifies the visual account token", () => {
    expect(existsSync(helperPath)).toBe(true);

    const source = helperSource();
    expect(source).toContain("VISUAL_TEST_ACCOUNT_TOKEN");
    expect(source).toContain("f17dd5569343eb59654629c9a6607a41daa2d09ce989f944a9cac85158232122");
    expect(source).toContain("context.addInitScript");
    expect(source).toContain("localStorage.setItem(\"accountToken\"");
    expect(source).toContain("page.waitForFunction");
    expect(source).toContain("localStorage.getItem(\"accountToken\") === token");
  });

  // Проверяет, что инструкции агенту больше не направляют обычный клиент к старому локальному файлу.
  it("documents the shared helper instead of the old client-side local account loader", () => {
    const forbiddenCall = ["chromium", "launch"].join(".");
    expect(rootAgents).toContain("client/visual-tests/visualTestBrowser.cjs");
    expect(rootAgents).toContain(`запрещено запускать Playwright напрямую через \`${forbiddenCall}\``);
    expect(rootAgents).not.toContain("visual-test-account.local.json");
    expect(rootAgents).not.toContain("загружается клиентом");
  });

  // Проверяет, что визуальные скрипты не создают браузер в обход общего helper-а.
  it("keeps direct playwright browser launch inside the shared helper only", () => {
    const forbiddenCall = ["chromium", "launch"].join(".");
    const offenders = projectFiles(".")
      .filter((path) => path !== `./${helperPath}`)
      .filter((path) => readFileSync(path, "utf8").includes(forbiddenCall));

    expect(offenders).toEqual([]);
  });
});
