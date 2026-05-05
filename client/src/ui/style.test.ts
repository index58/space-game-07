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
    const panel = readCssBlock(".hud-panel");
    const indicatorIcon = readCssBlock(".object-indicator__icon");
    const indicatorSvg = readCssBlock(".object-indicator__icon svg");
    const statusItem = readCssBlock(".minimap-status__item");
    const anchorIcon = readCssBlock(".minimap-status__anchor-icon");

    expect(panel).toContain("--hud-icon-muted: rgba(216, 243, 255, 0.64);");
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
    expect(readCssBlock(".hud-panel--right-bottom")).toContain("right: 1vh;");
    expect(readCssBlock(".hud-panel--right-bottom")).toContain("bottom: 1vh;");
    expect(readCssBlock(".hud-panel--left-top")).toContain("left: 1vh;");
    expect(readCssBlock(".hud-panel--left-top")).toContain("top: 1vh;");
  });

  it("uses smooth chat history movement and a dark solid game cursor", () => {
    const messageContent = readCssBlock(".chat-messages__content");
    const scrollbar = readCssBlock(".chat-scrollbar");
    const scrollbarThumb = readCssBlock(".chat-scrollbar__thumb");
    const draggingScrollbarThumb = readCssBlock(".chat-scrollbar.is-dragging .chat-scrollbar__thumb");
    const message = readCssBlock(".chat-message");
    const cursor = readCssBlock(".game-cursor");
    const cursorHighlight = readCssBlock(".game-cursor::after");

    expect(messageContent).toContain("transition: transform 120ms ease-out;");
    expect(scrollbar).toContain("right: 0.25vh;");
    expect(scrollbar).toContain("width: 1.1vh;");
    expect(scrollbarThumb).toContain("transition: top 120ms ease-out, height 120ms ease-out;");
    expect(draggingScrollbarThumb).toContain("transition: none;");
    expect(message).toContain("font: 1.05vh/1.2 Consolas, monospace;");
    expect(cursor).toContain("#252c32");
    expect(cursor).toContain("clip-path: polygon(0 0, 0 82%, 31% 58%, 50% 100%, 64% 94%, 45% 54%, 76% 54%);");
    expect(cursorHighlight).not.toContain("border-left");
  });
});
