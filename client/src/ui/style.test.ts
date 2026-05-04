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
});
