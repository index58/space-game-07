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
    const theme = readCssBlock(":root");
    const indicatorIcon = readCssBlock(".object-indicator__icon");
    const indicatorSvg = readCssBlock(".object-indicator__icon svg");
    const statusItem = readCssBlock(".minimap-status__item");
    const anchorIcon = readCssBlock(".minimap-status__anchor-icon");

    expect(theme).toContain("--hud-icon-muted: rgba(216, 243, 255, 0.64);");
    expect(theme).toContain("--hud-tab-height: 2.45vh;");
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

  // Проверяет, что обычные кнопки заметнее, а выбранная вкладка отличается только более ярким фоном.
  it("makes buttons prominent and selected tabs visually distinct", () => {
    const button = readCssBlock(".ui-kit-button");
    const tabs = readCssBlock(".ui-kit-tabs");
    const centeredTabs = readCssBlock(".ui-kit-tabs--center");
    const tab = readCssBlock(".ui-kit-tab");
    const selectedTab = readCssBlock(".ui-kit-tab.is-selected");

    expect(button).toContain("width: max-content;");
    expect(button).toContain("white-space: nowrap;");
    expect(button).toContain("padding: 0 1.05vh;");
    expect(button).toContain("background: rgba(126, 212, 255, 0.16);");
    expect(button).not.toMatch(/\bborder(?:-color)?\s*:/);
    expect(tabs).toContain("background: var(--hud-tab-panel-bg);");
    expect(tabs).toContain("padding: var(--hud-tab-panel-padding);");
    expect(centeredTabs).toContain("justify-content: center;");
    expect(tab).toContain("color: var(--hud-tab-inactive-text);");
    expect(tab).toContain("background: var(--hud-tab-panel-bg);");
    expect(selectedTab).toContain("color: #eefaff;");
    expect(selectedTab).toContain("background: var(--hud-surface-solid-bg);");
    expect(selectedTab).not.toMatch(/\bborder(?:-color)?\s*:/);
    expect(selectedTab).not.toContain("box-shadow");
  });

  // Проверяет, что новый общий стиль убирает декоративные рамки и оставляет чёрный фон у полей и списков.
  // Проверяет, что основные интерактивные контролы получают один тип и размер шрифта из общего правила.
  it("uses one shared font for core ui kit controls", () => {
    const theme = readCssBlock(":root");
    const controlFont = readCssBlock(
      ".ui-kit-button,\n.ui-kit-checkbox,\n.ui-kit-radio,\n.ui-kit-radio__option,\n.ui-kit-dropdown,\n.ui-kit-dropdown__item,\n.ui-kit-list,\n.ui-kit-list__item,\n.ui-kit-tree,\n.ui-kit-tree__item,\n.ui-kit-virtual-list,\n.ui-kit-virtual-list__item,\n.ui-kit-edit,\n.ui-kit-tabs,\n.ui-kit-tab,\n.ui-kit-stepper,\n.ui-kit-stepper .ui-kit-control,\n.ui-kit-context-menu,\n.ui-kit-context-menu__item,\n.ui-kit-tooltip",
    );
    const dropdown = readCssBlock(".ui-kit-dropdown");
    const listItems = readCssBlock(".ui-kit-dropdown__item,\n.ui-kit-context-menu__item,\n.ui-kit-list__item,\n.ui-kit-tree__item,\n.ui-kit-virtual-list__item");
    const tab = readCssBlock(".ui-kit-tab");
    const buttonGroup = readCssBlock(".ui-kit-button,\n.ui-kit-checkbox,\n.ui-kit-radio__option,\n.ui-kit-stepper");

    expect(theme).toContain("--hud-control-font-family: Consolas, monospace;");
    expect(theme).toContain("--hud-control-font-size: 1.05vh;");
    expect(controlFont).toContain("font-family: var(--hud-control-font-family);");
    expect(controlFont).toContain("font-size: var(--hud-control-font-size);");
    expect(dropdown).not.toMatch(/\bfont(?:-size|-family)?\s*:/);
    expect(listItems).not.toMatch(/\bfont(?:-size|-family)?\s*:/);
    expect(tab).not.toMatch(/\bfont(?:-size|-family)?\s*:/);
    expect(buttonGroup).not.toMatch(/\bfont(?:-size|-family)?\s*:/);
  });

  it("removes borders from hud surfaces and uses black backgrounds for input controls", () => {
    const borderlessBlocks = [
      ".hud-panel",
      ".object-indicator__bar",
      ".chat-messages",
      ".chat-error",
      ".chat-input",
      ".ui-kit-button,\n.ui-kit-checkbox,\n.ui-kit-radio__option,\n.ui-kit-dropdown,\n.ui-kit-dropdown__item,\n.ui-kit-list,\n.ui-kit-tree,\n.ui-kit-virtual-list,\n.ui-kit-edit,\n.ui-kit-stepper,\n.ui-kit-context-menu,\n.ui-kit-modal,\n.ui-kit-tooltip",
      ".ui-kit-button",
      ".ui-kit-button.is-hovered,\n.ui-kit-button.is-focused,\n.ui-kit-list__item.is-selected,\n.ui-kit-tree__item.is-selected,\n.ui-kit-dropdown__item.is-selected",
      ".ui-kit-tab.is-selected",
      ".ui-kit-checkbox__mark",
      ".ui-kit-tab",
      ".ui-kit-slider",
      ".chat-context-menu",
      ".pilot-toolbar__magazine",
      ".pilot-tool-slot",
      ".pilot-tool-slot.is-selected",
      ".minimap-compass",
      ".minimap-status__item",
      ".minimap-map",
      ".minimap-map__crosshair",
    ];

    for (const selector of borderlessBlocks) {
      expect(readCssBlock(selector)).not.toMatch(/\bborder(?:-color)?\s*:/);
    }

    expect(readCssBlock(".chat-input")).toContain("background: rgb(0, 0, 0);");
    expect(readCssBlock(".ui-kit-edit")).toContain("background: rgb(0, 0, 0);");
    expect(readCssBlock(".ui-kit-dropdown")).toContain("background: rgb(0, 0, 0);");
    expect(readCssBlock(".ui-kit-dropdown__menu")).toContain("background: rgb(0, 0, 0);");
    expect(readCssBlock(".ui-kit-list")).toContain("background: rgb(0, 0, 0);");
    expect(readCssBlock(".ui-kit-checkbox__mark")).toContain("background: rgb(0, 0, 0);");
    expect(readCssBlock(".ui-kit-radio__option")).toContain("background: rgb(0, 0, 0);");
  });

  // Проверяет, что окна и выпадающие списки имеют светлую обводку как явное исключение из общего правила.
  it("uses light borders only for windows and opened dropdown menus", () => {
    const theme = readCssBlock(":root");
    const dropdown = readCssBlock(".ui-kit-dropdown");

    expect(theme).toContain("--hud-surface-border: rgba(238, 250, 255, 0.48);");
    expect(readCssBlock(".ui-kit-modal")).toContain("border: 0.14vh solid var(--hud-surface-border);");
    expect(dropdown).not.toMatch(/\bborder(?:-color)?\s*:/);
    expect(readCssBlock(".ui-kit-dropdown__menu")).toContain("border: 0.14vh solid var(--hud-surface-border);");
  });

  // Проверяет, что заголовки всех модальных окон выравниваются по центру общим стилем.
  it("centers modal titles from the shared source", () => {
    const title = readCssBlock(".ui-kit-modal__title");

    expect(title).toContain("text-align: center;");
  });

  // Проверяет, что размеры выпадающего списка берутся из одного места, а не переопределяются отдельно в настройках.
  it("shares compact dropdown sizing between ui kit showcase and settings", () => {
    const theme = readCssBlock(":root");
    const dropdown = readCssBlock(".ui-kit-dropdown");
    const dropdownValue = readCssBlock(".ui-kit-dropdown__value");
    const settingsDropdown = readCssBlock(".settings-input-row .ui-kit-dropdown");

    expect(theme).toContain("--hud-dropdown-min-height: 2.35vh;");
    expect(theme).toContain("--hud-dropdown-padding-y: 0.35vh;");
    expect(theme).toContain("--hud-control-font-size: 1.05vh;");
    expect(theme).toContain("--hud-control-padding-x: 0.9vh;");
    expect(theme).toContain("--hud-settings-action-bg: rgb(14, 28, 39);");
    expect(dropdown).toContain("min-height: var(--hud-dropdown-min-height);");
    expect(dropdown).toContain("padding: 0 var(--hud-control-padding-x);");
    expect(dropdown).toContain("padding-top: var(--hud-dropdown-padding-y);");
    expect(dropdown).toContain("padding-bottom: var(--hud-dropdown-padding-y);");
    expect(dropdown).toContain("padding-right: calc(var(--hud-control-padding-x) + 1.45vh);");
    expect(dropdown).not.toMatch(/\bfont(?:-size|-family)?\s*:/);
    expect(dropdownValue).toContain("display: flex;");
    expect(dropdownValue).toContain("align-items: center;");
    expect(dropdownValue).toContain("min-height: calc(var(--hud-dropdown-min-height) - (var(--hud-dropdown-padding-y) * 2));");
    expect(settingsDropdown).toContain("width: 100%;");
    expect(settingsDropdown).toContain("min-width: 0;");
    expect(settingsDropdown).not.toContain("min-height:");
    expect(settingsDropdown).not.toContain("padding-top:");
    expect(settingsDropdown).not.toContain("padding-bottom:");
    expect(settingsDropdown).not.toContain("font-size:");
  });

  // Проверяет, что обычный список получает такой же чёрный внутренний отступ по краю, как выпавший список.
  it("uses the same black perimeter padding for list and dropdown menu", () => {
    const theme = readCssBlock(":root");
    const menu = readCssBlock(".ui-kit-dropdown__menu");
    const menuViewport = readCssBlock(".ui-kit-dropdown__menu-viewport");
    const menuClip = readCssBlock(".ui-kit-dropdown__menu-clip");
    const menuContent = readCssBlock(".ui-kit-dropdown__menu-content");
    const list = readCssBlock(".ui-kit-list");

    expect(theme).toContain("--hud-list-padding: 0.35vh;");
    expect(menu).toContain("padding: 0;");
    expect(menu).toContain("background: rgb(0, 0, 0);");
    expect(menuViewport).toContain("box-sizing: border-box;");
    expect(menuViewport).toContain("display: grid;");
    expect(menuViewport).toContain("background: rgb(0, 0, 0);");
    expect(menuClip).toContain("overflow: hidden;");
    expect(menuClip).toContain("min-height: 0;");
    expect(menuContent).toContain("padding: var(--hud-list-padding);");
    expect(menuContent).toContain("box-sizing: border-box;");
    expect(list).toContain("display: grid;");
    expect(list).toContain("gap: 0;");
    expect(list).toContain("padding: var(--hud-list-padding);");
    expect(list).toContain("background: rgb(0, 0, 0);");
  });

  // Проверяет, что окна и панели получают более светлый общий фон, а чёрные поля ввода не меняются.
  it("uses lighter backgrounds for windows and panels while preserving black input controls", () => {
    const theme = readCssBlock(":root");

    expect(theme).toContain("--hud-surface-bg: rgba(18, 34, 46, 0.9);");
    expect(theme).toContain("--hud-surface-soft-bg: rgba(18, 34, 46, 0.78);");
    expect(theme).toContain("--hud-surface-solid-bg: rgb(18, 34, 46);");
    expect(readCssBlock(".hud-panel")).toContain("background: var(--hud-surface-soft-bg);");
    expect(readCssBlock(".object-indicator__bar")).toContain("background: var(--hud-surface-bg);");
    expect(readCssBlock(".chat-messages")).toContain("background: var(--hud-surface-soft-bg);");
    expect(readCssBlock(".ui-kit-button,\n.ui-kit-checkbox,\n.ui-kit-radio__option,\n.ui-kit-dropdown,\n.ui-kit-dropdown__item,\n.ui-kit-list,\n.ui-kit-tree,\n.ui-kit-virtual-list,\n.ui-kit-edit,\n.ui-kit-stepper,\n.ui-kit-context-menu,\n.ui-kit-modal,\n.ui-kit-tooltip")).toContain("background: var(--hud-surface-bg);");
    expect(readCssBlock(".game-window-layer .ui-kit-modal")).toContain("background: var(--hud-surface-solid-bg);");
    expect(readCssBlock(".chat-context-menu")).toContain("background: var(--hud-surface-solid-bg);");
    expect(readCssBlock(".pilot-toolbar__magazine")).toContain("background: var(--hud-surface-bg);");
    expect(readCssBlock(".pilot-tool-slot")).toContain("background: var(--hud-surface-soft-bg);");
    expect(readCssBlock(".minimap-compass")).toContain("background: var(--hud-surface-bg);");
    expect(readCssBlock(".minimap-map")).toContain("var(--hud-surface-bg)");
    expect(readCssBlock(".chat-input")).toContain("background: rgb(0, 0, 0);");
    expect(readCssBlock(".ui-kit-dropdown")).toContain("background: rgb(0, 0, 0);");
  });

  // Проверяет, что раскрытый список UI Kit рисуется поверх панели и не участвует в её раскладке.
  it("keeps ui kit dropdown menu out of panel flow", () => {
    const dropdown = readCssBlock(".ui-kit-dropdown");
    const dropdownMarker = readCssBlock(".ui-kit-dropdown::after");
    const menu = readCssBlock(".ui-kit-dropdown__menu");
    const menuContent = readCssBlock(".ui-kit-dropdown__menu-content");
    const dropdownItem = readCssBlock(".ui-kit-dropdown__item");
    const scrollbar = readCssBlock(".ui-kit-scrollbar");

    expect(dropdown).toContain("position: relative;");
    expect(dropdown).toContain("padding-right: calc(var(--hud-control-padding-x) + 1.45vh);");
    expect(dropdownMarker).toContain("border-left: 0.35vh solid transparent;");
    expect(dropdownMarker).toContain("border-right: 0.35vh solid transparent;");
    expect(dropdownMarker).toContain("border-top: 0.45vh solid rgba(216, 243, 255, 0.82);");
    expect(dropdownMarker).toContain("right: var(--hud-control-padding-x);");
    expect(menu).toContain("position: fixed;");
    expect(menu).toContain("z-index: 19;");
    expect(menu).toContain("grid-template-columns: minmax(0, 1fr);");
    expect(menu).toContain("background: rgb(0, 0, 0);");
    expect(menu).toContain("box-shadow: 0 0.5vh 1.4vh rgba(0, 4, 8, 0.72);");
    expect(menuContent).toContain("gap: 0;");
    expect(dropdownItem).toContain("border: 0;");
    expect(dropdownItem).toContain("background: transparent;");
    expect(dropdownItem).toContain("padding-left: var(--hud-control-padding-x);");
    expect(dropdownItem).toContain("padding-right: var(--hud-control-padding-x);");
    expect(css).toContain(".ui-kit-list__item.is-selected,\n.ui-kit-tree__item.is-selected,\n.ui-kit-dropdown__item.is-selected");
    expect(scrollbar).toContain("z-index: 30;");
  });

  // Проверяет, что раскрытый список настроек имеет фиксированное окно и использует единую полосу прокрутки.
  it("keeps settings dropdown as a single scrollable column", () => {
    const layer = readCssBlock(".game-window-layer");
    const sharedLayerSize = readCssBlock(".game-window-layer--settings,\n.game-window-layer--showcase");
    const settingsLayer = readCssBlock(".game-window-layer--settings");
    const modal = readCssBlock(".game-window-layer .ui-kit-modal");
    const body = readCssBlock(".game-window-layer .ui-kit-modal__body");
    const settings = readCssBlock(".settings-modal");
    const table = readCssBlock(".settings-input-table");
    const footer = readCssBlock(".settings-modal__footer");
    const actions = readCssBlock(".settings-modal__actions");
    const row = readCssBlock(".settings-input-row");
    const menu = readCssBlock(".ui-kit-dropdown__menu");
    const viewport = readCssBlock(".ui-kit-dropdown__menu[id^=\"settings-input-select-\"] .ui-kit-dropdown__menu-viewport");
    const action = readCssBlock(".settings-input-row__action");
    const sharedScrollbars = readCssBlock(".ui-kit-scrollbar.ui-kit-dropdown-scrollbar,\n.ui-kit-scrollbar.settings-input-scrollbar");

    expect(layer).toContain("position: fixed;");
    expect(layer).toContain("transform: translate(-50%, -50%);");
    expect(sharedLayerSize).toContain("width: min(118vh, calc(100vw - 4vh));");
    expect(sharedLayerSize).toContain("height: min(62vh, calc(100vh - 4vh));");
    expect(settingsLayer).toContain("z-index: 18;");
    expect(css).toContain(".game-window-layer--showcase {\n  z-index: 17;\n}");
    expect(modal).toContain("grid-template-rows: auto minmax(0, 1fr);");
    expect(modal).toContain("overflow: hidden;");
    expect(body).toContain("min-height: 0;");
    expect(body).toContain("overflow: hidden;");
    expect(settings).toContain("height: 100%;");
    expect(settings).toContain("min-height: 0;");
    expect(table).toContain("overflow: hidden;");
    expect(footer).toContain("grid-template-columns: minmax(0, 1fr) auto minmax(0, 1fr);");
    expect(actions).toContain("grid-column: 2;");
    expect(actions).toContain("justify-content: center;");
    expect(row).toContain("grid-template-columns: minmax(22vh, 1fr) minmax(36vh, 1fr);");
    expect(menu).toContain("grid-template-columns: minmax(0, 1fr);");
    expect(menu).toContain("position: fixed;");
    expect(menu).not.toContain("max-height: 22vh;");
    expect(viewport).toContain("height: 22vh;");
    expect(viewport).toContain("max-height: 22vh;");
    expect(action).toContain("padding: var(--hud-dropdown-padding-y) var(--hud-control-padding-x);");
    expect(action).toContain("background: var(--hud-settings-action-bg);");
    expect(action).toContain("border: 0;");
    expect(action).toContain("font-family: var(--hud-control-font-family);");
    expect(action).toContain("font-size: var(--hud-control-font-size);");
    expect(action).not.toContain("text-shadow");
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
    const tabWithMarker = readCssBlock(".ui-kit-tab--with-marker");
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
    expect(tab).toContain("flex: 0 0 auto;");
    expect(tab).toContain("width: max-content;");
    expect(tab).toContain("min-width: 0;");
    expect(tab).not.toContain("14vh");
    expect(uiKitTabs).toContain("height: calc(var(--hud-tab-height) + var(--hud-tab-panel-padding) * 2);");
    expect(uiKitTabs).toContain("box-sizing: border-box;");
    expect(uiKitTabs).toContain("overflow: visible;");
    expect(uiKitTab).toContain("flex: 0 0 auto;");
    expect(uiKitTab).toContain("width: max-content;");
    expect(uiKitTab).toContain("min-width: 0;");
    expect(uiKitTab).toContain("padding: 0 1.05vh;");
    expect(uiKitTab).toContain("overflow: visible;");
    expect(tabWithMarker).toContain("padding-left: calc((var(--hud-tab-height) - 1.6vh) / 2);");
    expect(uiKitTab).toContain("height: var(--hud-tab-height);");
    expect(uiKitTab).toContain("max-height: var(--hud-tab-height);");
    expect(readCssBlock(".ui-kit-tab__marker")).toContain("flex: 0 0 1.6vh;");
    expect(readCssBlock(".ui-kit-tab__marker")).toContain("width: 1.6vh;");
    expect(readCssBlock(".ui-kit-tab__marker")).toContain("height: 1.6vh;");
    expect(readCssBlock(".ui-kit-tab__label")).toContain("align-items: center;");
    expect(readCssBlock(".ui-kit-tab__label")).toContain("height: 100%;");
    expect(uiKitBadge).toContain("position: absolute;");
    expect(uiKitBadge).toContain("right: -0.35vh;");
    expect(uiKitBadge).toContain("top: -0.35vh;");
    expect(uiKitBadge).toContain("border: 0;");
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
