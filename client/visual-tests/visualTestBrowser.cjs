const { execFileSync } = require("node:child_process");
const path = require("node:path");

// Берет Playwright из проекта или из глобальной установки, которую использует Codex для визуальных проверок.
const loadPlaywright = () => {
  try {
    return require("@playwright/test");
  } catch (localError) {
    try {
      const globalNodeModules = execFileSync("npm.cmd", ["root", "-g"], { encoding: "utf8", shell: true }).trim();
      return require(path.join(globalNodeModules, "@playwright", "test"));
    } catch {
      throw localError;
    }
  }
};

const { chromium } = loadPlaywright();

const VISUAL_TEST_URL = "http://127.0.0.1:5173";
const VISUAL_TEST_ACCOUNT_TOKEN = "f17dd5569343eb59654629c9a6607a41daa2d09ce989f944a9cac85158232122";
const DEFAULT_VIEWPORT = { width: 1365, height: 768 };
const DEFAULT_TIMEOUT_MS = 30000;

// Создает изолированный браузерный контекст, чтобы тестовый аккаунт не попадал в обычный браузер игрока.
const createVisualTestContext = async (options = {}) => {
  const browser = await chromium.launch({
    channel: options.channel ?? "chrome",
    headless: options.headless ?? true,
  });
  const context = await browser.newContext({
    viewport: options.viewport ?? DEFAULT_VIEWPORT,
  });

  await context.addInitScript((token) => {
    localStorage.setItem("accountToken", token);
  }, VISUAL_TEST_ACCOUNT_TOKEN);

  return { browser, context };
};

// Открывает работающий клиент игры только после установки тестового токена в изолированное хранилище.
const openVisualTestPage = async (options = {}) => {
  const { browser, context } = await createVisualTestContext(options);
  const page = await context.newPage();
  const url = options.url ?? VISUAL_TEST_URL;
  const timeout = options.timeoutMs ?? DEFAULT_TIMEOUT_MS;

  await page.goto(url, { waitUntil: "domcontentloaded", timeout });
  await page.waitForFunction(
    (token) => localStorage.getItem("accountToken") === token,
    VISUAL_TEST_ACCOUNT_TOKEN,
    { timeout },
  );

  return {
    browser,
    context,
    page,
    token: VISUAL_TEST_ACCOUNT_TOKEN,
    url,
    // Закрывает весь тестовый профиль целиком, не трогая обычный браузер пользователя.
    close: async () => {
      await browser.close();
    },
  };
};

module.exports = {
  DEFAULT_VIEWPORT,
  VISUAL_TEST_ACCOUNT_TOKEN,
  VISUAL_TEST_URL,
  createVisualTestContext,
  openVisualTestPage,
};
