import { describe, expect, it } from "vitest";
import { readFileSync } from "node:fs";

const css = readFileSync("src/style.css", "utf8");

const readCssBlock = (selector: string): string => {
  const escapedSelector = selector.replace(/[.*+?^${}()|[\]\\]/g, "\\$&");
  const match = css.match(new RegExp(`${escapedSelector}\\s*\\{([^}]*)\\}`));

  if (!match) {
    throw new Error(`Не найден CSS-блок ${selector}`);
  }

  return match[1];
};

describe("HUD styles", () => {
  it("keeps pilot toolbar slots inside the toolbar grid", () => {
    const slots = readCssBlock(".pilot-toolbar__slots");
    const slot = readCssBlock(".pilot-tool-slot");

    expect(slots).toContain("grid-template-columns: repeat(10, minmax(0, 1fr));");
    expect(slot).toContain("width: 100%;");
    expect(slot).toContain("aspect-ratio: 1;");
    expect(slot).not.toContain("width: 4.96vh;");
  });

  it("uses one muted icon style for object indicators and status icons", () => {
    const root = readCssBlock("#ui-root");
    const indicatorIcon = readCssBlock(".object-indicator__icon");
    const indicatorSvg = readCssBlock(".object-indicator__icon svg");
    const statusItem = readCssBlock(".minimap-status__item");
    const anchorIcon = readCssBlock(".minimap-status__anchor-icon");

    expect(root).toContain("--hud-icon-muted: rgba(216, 243, 255, 0.64);");
    expect(root).toContain("--hud-tab-height: 2.45vh;");
    expect(indicatorIcon).toContain("color: var(--hud-icon-muted);");
    expect(statusItem).toContain("color: var(--hud-icon-muted);");
    expect(indicatorSvg).toContain("width: 2.35vh;");
    expect(indicatorSvg).toContain("height: 2.35vh;");
    expect(anchorIcon).toContain("width: 2.35vh;");
    expect(anchorIcon).toContain("height: 2.35vh;");
  });

  it("aligns HUD panels with the same edge spacing as debug overlay", () => {
    expect(readCssBlock(".hud-panel--left-bottom")).toContain("left: 1vh;");
    expect(readCssBlock(".hud-panel--left-bottom")).toContain("bottom: 1vh;");
    expect(readCssBlock(".hud-panel--left-middle")).toContain("left: 1vh;");
    expect(readCssBlock(".hud-panel--bottom-center")).toContain("bottom: 1vh;");
    expect(readCssBlock(".hud-panel--right-middle")).toContain("right: 1vh;");
    expect(readCssBlock(".hud-panel--right-bottom")).toContain("right: 1vh;");
    expect(readCssBlock(".hud-panel--right-bottom")).toContain("bottom: 1vh;");
    expect(readCssBlock(".hud-panel--left-top")).toContain("left: 1vh;");
    expect(readCssBlock(".hud-panel--left-top")).toContain("top: 1vh;");
  });

  // Проверяет, что информационная панель наследует плотный HUD-стиль без отдельных отступов от края экрана.
  it("styles information panel as compact HUD panel", () => {
    const panel = readCssBlock(".information-panel");
    const row = readCssBlock(".information-panel__row");
    const label = readCssBlock(".information-panel__label");

    expect(panel).toContain("width: 30vh;");
    expect(panel).toContain("font: 1.05vh/1.25 Consolas, monospace;");
    expect(row).toContain("grid-template-columns: 8.2vh minmax(0, 1fr);");
    expect(label).toContain("color: rgba(216, 243, 255, 0.58);");
  });

  // Проверяет, что раскрытый список UI Kit рисуется поверх панели и не участвует в её раскладке.
  it("keeps ui kit dropdown menu out of panel flow", () => {
    const dropdown = readCssBlock(".ui-kit-dropdown");
    const dropdownMarker = readCssBlock(".ui-kit-dropdown::after");
    const menu = readCssBlock(".ui-kit-dropdown__menu");
    const scrollbar = readCssBlock(".ui-kit-scrollbar");

    expect(dropdown).toContain("position: relative;");
    expect(dropdown).toContain("padding-right: 2.1vh;");
    expect(dropdownMarker).toContain("border-left: 0.35vh solid transparent;");
    expect(dropdownMarker).toContain("border-right: 0.35vh solid transparent;");
    expect(dropdownMarker).toContain("border-top: 0.45vh solid rgba(216, 243, 255, 0.82);");
    expect(dropdownMarker).toContain("right: 0.75vh;");
    expect(menu).toContain("position: fixed;");
    expect(menu).toContain("z-index: 19;");
    expect(menu).toContain("grid-template-columns: minmax(0, 1fr);");
    expect(menu).toContain("background: rgb(18, 44, 60);");
    expect(menu).toContain("border: 0.18vh solid rgba(180, 232, 255, 0.72);");
    expect(menu).toContain("box-shadow: 0 0.5vh 1.4vh rgba(0, 4, 8, 0.72), inset 0 0 0 0.12vh rgba(238, 250, 255, 0.08);");
    expect(scrollbar).toContain("z-index: 30;");
  });

  // Проверяет, что раскрытый список настроек имеет фиксированное окно и использует единую полосу прокрутки.
  it("keeps settings dropdown as a single scrollable column", () => {
    const modal = readCssBlock(".settings-modal-layer .ui-kit-modal");
    const body = readCssBlock(".settings-modal-layer .ui-kit-modal__body");
    const settings = readCssBlock(".settings-modal");
    const table = readCssBlock(".settings-input-table");
    const menu = readCssBlock(".ui-kit-dropdown__menu");
    const viewport = readCssBlock(".ui-kit-dropdown__menu[id^=\"settings-input-select-\"] .ui-kit-dropdown__menu-viewport");
    const sharedScrollbars = readCssBlock(".ui-kit-scrollbar.ui-kit-dropdown-scrollbar,\n.ui-kit-scrollbar.settings-input-scrollbar");

    expect(modal).toContain("grid-template-rows: auto minmax(0, 1fr);");
    expect(modal).toContain("overflow: hidden;");
    expect(body).toContain("min-height: 0;");
    expect(body).toContain("overflow: hidden;");
    expect(settings).toContain("height: 100%;");
    expect(settings).toContain("min-height: 0;");
    expect(table).toContain("overflow: hidden;");
    expect(menu).toContain("grid-template-columns: minmax(0, 1fr);");
    expect(menu).toContain("position: fixed;");
    expect(menu).not.toContain("max-height: 22vh;");
    expect(viewport).toContain("height: 22vh;");
    expect(viewport).toContain("max-height: 22vh;");
    expect(sharedScrollbars).toContain("position: absolute;");
    expect(sharedScrollbars).toContain("min-height: 0;");
  });

  it("uses smooth chat history movement and a dark solid game cursor", () => {
    const messageContent = readCssBlock(".chat-messages__content");
    const scrollbar = readCssBlock(".chat-scrollbar");
    const chatScrollbarInstance = readCssBlock(".ui-kit-scrollbar.chat-scrollbar");
    const scrollbarThumb = readCssBlock(".ui-kit-scrollbar__thumb");
    const draggingScrollbarThumb = readCssBlock(".ui-kit-scrollbar.is-dragging .ui-kit-scrollbar__thumb");
    const message = readCssBlock(".chat-message");
    const tabs = readCssBlock(".chat-tabs");
    const tab = readCssBlock(".chat-tab");
    const uiKitTabs = readCssBlock(".ui-kit-tabs");
    const uiKitTab = readCssBlock(".ui-kit-tab");
    const uiKitBadge = readCssBlock(".ui-kit-tab__badge");
    const error = readCssBlock(".chat-error");
    const input = readCssBlock(".chat-input");
    const inputText = readCssBlock(".chat-input__text");
    const inputCaret = readCssBlock(".chat-input__caret");
    const cursor = readCssBlock(".game-cursor");
    const cursorHighlight = readCssBlock(".game-cursor::after");

    expect(messageContent).toContain("transition: transform 120ms ease-out;");
    expect(scrollbar).toContain("right: 0.25vh;");
    expect(scrollbar).toContain("width: 1.1vh;");
    expect(chatScrollbarInstance).toContain("position: absolute;");
    expect(chatScrollbarInstance).toContain("right: 0.25vh;");
    expect(chatScrollbarInstance).toContain("min-height: 0;");
    expect(scrollbarThumb).toContain("transition: top 120ms ease-out, height 120ms ease-out;");
    expect(draggingScrollbarThumb).toContain("transition: none;");
    expect(css).not.toContain(".chat-scrollbar__thumb");
    expect(message).toContain("font: 1.05vh/1.2 Consolas, monospace;");
    expect(tabs).toContain("margin-top: 0.8vh;");
    expect(tab).toContain("flex: 0 0 14vh;");
    expect(uiKitTabs).toContain("height: var(--hud-tab-height);");
    expect(uiKitTabs).toContain("overflow: visible;");
    expect(uiKitTab).toContain("height: var(--hud-tab-height);");
    expect(uiKitTab).toContain("max-height: var(--hud-tab-height);");
    expect(readCssBlock(".ui-kit-tab__marker")).toContain("max-height: 1.6vh;");
    expect(readCssBlock(".ui-kit-tab__label")).toContain("align-items: center;");
    expect(readCssBlock(".ui-kit-tab__label")).toContain("height: 100%;");
    expect(uiKitBadge).toContain("top: 0.18vh;");
    expect(css).not.toContain(".chat-tab__unread");
    expect(error).toContain("position: absolute;");
    expect(error).toContain("animation-duration: 4.8s;");
    expect(css).toContain("@keyframes chat-error-fade-odd");
    expect(css).toContain("@keyframes chat-error-fade-even");
    expect(input).toContain("position: relative;");
    expect(inputText).not.toContain("display: flex;");
    expect(inputCaret).toContain("position: absolute;");
    expect(inputCaret).toContain("min-width: 2px;");
    expect(cursor).toContain("#252c32");
    expect(cursor).toContain("clip-path: polygon(0 0, 0 82%, 31% 58%, 50% 100%, 64% 94%, 45% 54%, 76% 54%);");
    expect(cursorHighlight).not.toContain("border-left");
  });
});
