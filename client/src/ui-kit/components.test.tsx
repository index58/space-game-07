import { render } from "solid-js/web";
import { afterEach, describe, expect, it } from "vitest";
import { Button, Checkbox, ContextMenu, Dropdown, EditControl, ListBox, Modal, NumericStepper, RadioGroup, Scrollbar, Slider, Splitter, Tabs, TextInput, Tooltip, TreeView, VirtualList } from "./components";
import type { DropdownProps, UiKitOption } from "./components";

let dispose: (() => void) | null = null;

afterEach(() => {
  dispose?.();
  dispose = null;
  document.body.innerHTML = "";
});

describe("ui-kit components", () => {
  // Проверяет, что базовые контролы получают единый класс и состояния UI Kit.
  it("renders buttons and toggles in shared ui kit style", () => {
    const root = document.createElement("div");
    document.body.append(root);

    dispose = render(() => (
      <>
        <Button id="save" label="Save" state="hovered" />
        <Checkbox id="flag" label="Flag" checked={true} />
        <RadioGroup id="mode" value="a" options={[{ value: "a", label: "A" }, { value: "b", label: "B" }]} />
      </>
    ), root);

    expect(root.querySelector(".ui-kit-button.is-hovered")?.textContent).toBe("Save");
    expect(root.querySelector(".ui-kit-checkbox.is-checked")?.textContent).toContain("Flag");
    expect(root.querySelectorAll(".ui-kit-radio__option").length).toBe(2);
  });

  // Проверяет, что списочные контролы и dropdown используют общий визуальный слой.
  it("renders selection controls with selected state", () => {
    const root = document.createElement("div");
    document.body.append(root);

    dispose = render(() => (
      <>
        <Dropdown id="select" label="Mode" open={true} selectedValue="b" options={[{ value: "a", label: "A" }, { value: "b", label: "B" }]} />
        <ListBox id="list" selectedValue="2" scrollOffsetPx={12} items={[{ value: "1", label: "One", labelPrefix: "✓" }, { value: "2", label: "Two" }]} />
        <TreeView id="tree" selectedValue="child" nodes={[{ value: "root", label: "Root", children: [{ value: "child", label: "Child" }] }]} />
        <VirtualList id="virtual" startIndex={10} items={Array.from({ length: 3 }, (_, index) => ({ value: String(index), label: `Item ${index}` }))} />
      </>
    ), root);

    expect(document.body.querySelector(".ui-kit-dropdown__menu")).not.toBeNull();
    expect(root.querySelector<HTMLElement>("#list .ui-kit-list__content")?.style.transform).toBe("translateY(-12px)");
    expect(root.querySelector("#list-1 .ui-kit-list__item-label-prefix")?.textContent).toBe("✓");
    expect(root.querySelector(".ui-kit-list__item.is-selected")?.textContent).toBe("Two");
    expect(root.querySelector(".ui-kit-tree__item.is-selected")?.textContent).toBe("Child");
    expect(root.querySelector(".ui-kit-virtual-list__index")?.textContent).toBe("10");
  });

  // Проверяет, что пункты раскрытого списка получают явный слой для игрового hit-test поверх соседних контролов.
  it("marks dropdown options as overlay controls", () => {
    const root = document.createElement("div");
    document.body.append(root);

    dispose = render(() => (
      <Dropdown id="select" label="Mode" open={true} selectedValue="b" options={[{ value: "a", label: "A" }, { value: "b", label: "B" }]} />
    ), root);

    expect(document.body.querySelector<HTMLElement>("#select-a")?.dataset.uiZIndex).toBe("1000");
    expect(document.body.querySelector<HTMLElement>("#select-b")?.dataset.uiZIndex).toBe("1000");
  });

  // Проверяет, что раскрытый список получает экранный перехватчик для внешнего клика.
  it("renders dropdown outside click blocker under options", () => {
    const root = document.createElement("div");
    document.body.append(root);

    dispose = render(() => (
      <Dropdown id="select" label="Mode" open={true} selectedValue="b" options={[{ value: "a", label: "A" }, { value: "b", label: "B" }]} />
    ), root);

    const blocker = document.body.querySelector<HTMLElement>("#select-outside-blocker");
    expect(blocker?.dataset.uiKind).toBe("modal");
    expect(blocker?.dataset.uiZIndex).toBe("900");
    expect(document.body.querySelector<HTMLElement>("#select-a")?.dataset.uiZIndex).toBe("1000");
  });

  // Проверяет, что экранный перехватчик не ограничивается таблицей настроек.
  it("portals dropdown outside click blocker out of settings clipping containers", () => {
    const root = document.createElement("div");
    document.body.append(root);

    dispose = render(() => (
      <div class="settings-input-table">
        <Dropdown id="select" label="Mode" open={true} selectedValue="b" options={[{ value: "a", label: "A" }, { value: "b", label: "B" }]} />
      </div>
    ), root);

    const blocker = document.body.querySelector<HTMLElement>("#select-outside-blocker");
    expect(blocker?.closest(".settings-input-table")).toBeNull();
    expect(blocker ? document.body.contains(blocker) : false).toBe(true);
  });

  // Проверяет, что раскрытое меню не обрезается контейнерами модального окна настроек.
  it("portals dropdown menu out of settings clipping containers", () => {
    const root = document.createElement("div");
    document.body.append(root);

    dispose = render(() => (
      <div class="settings-input-table">
        <Dropdown id="select" label="Mode" open={true} selectedValue="b" options={[{ value: "a", label: "A" }, { value: "b", label: "B" }]} />
      </div>
    ), root);

    const menu = document.body.querySelector<HTMLElement>("#select-menu");
    expect(menu?.closest(".settings-input-table")).toBeNull();
    expect(menu ? document.body.contains(menu) : false).toBe(true);
  });

  // Проверяет, что выпавший список живёт вне корня HUD и наследует общие стили только со страницы.
  it("portals dropdown menu outside ui root", () => {
    const root = document.createElement("div");
    root.id = "ui-root";
    document.body.append(root);

    dispose = render(() => (
      <Dropdown id="select" label="Mode" open={true} selectedValue="b" options={[{ value: "a", label: "A" }, { value: "b", label: "B" }]} />
    ), root);

    const menu = document.body.querySelector<HTMLElement>("#select-menu");
    expect(menu).not.toBeNull();
    expect(menu?.closest("#ui-root")).toBeNull();
  });

  // Проверяет, что вынесенное меню привязано к экранным границам исходного поля.
  it("positions portaled dropdown menu at trigger screen rect", async () => {
    const originalGetBoundingClientRect = HTMLElement.prototype.getBoundingClientRect;
    HTMLElement.prototype.getBoundingClientRect = function getBoundingClientRect() {
      if (this.id === "select") {
        return { x: 120, y: 300, left: 120, top: 300, right: 360, bottom: 332, width: 240, height: 32, toJSON: () => ({}) } as DOMRect;
      }
      return originalGetBoundingClientRect.call(this);
    };
    const root = document.createElement("div");
    document.body.append(root);

    try {
      dispose = render(() => (
        <Dropdown id="select" label="Mode" open={true} selectedValue="b" options={[{ value: "a", label: "A" }, { value: "b", label: "B" }]} />
      ), root);
      await Promise.resolve();

      const menu = document.body.querySelector<HTMLElement>("#select-menu");
      expect(menu?.style.left).toBe("120px");
      expect(menu?.style.top).toBe("332px");
      expect(menu?.style.width).toBe("240px");
    } finally {
      HTMLElement.prototype.getBoundingClientRect = originalGetBoundingClientRect;
    }
  });

  // Проверяет, что пустая подпись выпадающего списка не занимает отдельную строку.
  // Проверяет, что меню у нижней границы экрана поднимается над полем и не уходит за экран.
  it("keeps portaled dropdown menu inside bottom viewport edge", async () => {
    const originalGetBoundingClientRect = HTMLElement.prototype.getBoundingClientRect;
    const originalInnerHeight = window.innerHeight;
    Object.defineProperty(window, "innerHeight", { configurable: true, value: 700 });
    HTMLElement.prototype.getBoundingClientRect = function getBoundingClientRect() {
      if (this.id === "select") {
        return { x: 120, y: 560, left: 120, top: 560, right: 360, bottom: 592, width: 240, height: 32, toJSON: () => ({}) } as DOMRect;
      }
      if (this.id === "select-menu") {
        return { x: 120, y: 592, left: 120, top: 592, right: 360, bottom: 812, width: 240, height: 220, toJSON: () => ({}) } as DOMRect;
      }
      return originalGetBoundingClientRect.call(this);
    };
    const root = document.createElement("div");
    document.body.append(root);

    try {
      dispose = render(() => (
        <Dropdown id="select" label="Mode" open={true} selectedValue="b" options={[{ value: "a", label: "A" }, { value: "b", label: "B" }]} />
      ), root);
      await Promise.resolve();

      const menu = document.body.querySelector<HTMLElement>("#select-menu");
      expect(menu?.style.top).toBe("340px");
    } finally {
      HTMLElement.prototype.getBoundingClientRect = originalGetBoundingClientRect;
      Object.defineProperty(window, "innerHeight", { configurable: true, value: originalInnerHeight });
    }
  });

  it("omits dropdown label when it is empty", () => {
    const root = document.createElement("div");
    document.body.append(root);

    dispose = render(() => (
      <Dropdown id="select" open={false} selectedValue="b" options={[{ value: "a", label: "A" }, { value: "b", label: "B" }]} />
    ), root);

    expect(root.querySelector(".ui-kit-dropdown__label")).toBeNull();
    expect(root.querySelector(".ui-kit-dropdown__value")?.textContent).toBe("B");
  });

  // Проверяет, что внешний код может переиспользовать публичные типы выпадающего списка.
  it("exposes reusable dropdown props and option types", () => {
    const options: UiKitOption[] = [{ value: "manual", label: "Manual" }];
    const props: DropdownProps = { id: "drive-mode", open: false, selectedValue: "manual", options };

    expect(props.options[0].value).toBe("manual");
  });

  // Проверяет, что overlay и drag-подобные controls имеют стабильные элементы для hit-test.
  it("renders overlays, sliders and text edit primitives", () => {
    const root = document.createElement("div");
    document.body.append(root);

    dispose = render(() => (
      <>
        <EditControl id="edit" text="abcdef" selectionStart={1} selectionEnd={4} focused={true} />
        <Tabs id="tabs" align="center" selectedValue="chat" tabs={[{ value: "map", label: "Map" }, { value: "chat", label: "Chat", badge: 2 }]} />
        <Scrollbar id="scroll" thumbTopPercent={25} thumbHeightPercent={40} dragging={true} />
        <Slider id="slider" value={40} min={0} max={100} label="40 / 100" />
        <NumericStepper id="stepper" value={7} />
        <Splitter id="splitter" vertical={true} />
        <ContextMenu id="menu" items={[{ value: "close", label: "Close" }]} />
        <Modal id="modal" title="Panel">Body</Modal>
        <Tooltip id="tip" text="Hint" />
      </>
    ), root);

    expect(root.querySelector(".ui-kit-edit__selection")?.textContent).toBe("bcd");
    expect(root.querySelector("#tabs")?.classList.contains("ui-kit-tabs--center")).toBe(true);
    expect(root.querySelector(".ui-kit-tab.is-selected")?.textContent).toContain("Chat");
    expect(root.querySelector(".ui-kit-scrollbar.is-dragging")).not.toBeNull();
    expect(root.querySelector<HTMLElement>(".ui-kit-slider__fill")?.style.width).toBe("40%");
    expect(root.querySelector(".ui-kit-slider__label")?.textContent).toBe("40 / 100");
    expect(root.querySelector(".ui-kit-stepper")?.textContent).toContain("7");
    expect(root.querySelector(".ui-kit-splitter.is-vertical")).not.toBeNull();
    expect(root.querySelector(".ui-kit-context-menu__item")?.textContent).toBe("Close");
    expect(root.querySelector(".ui-kit-modal__title")?.textContent).toBe("Panel");
    expect(root.querySelector("#modal-close-button")?.textContent).toBe("");
    expect(root.querySelector(".ui-kit-tooltip")?.textContent).toBe("Hint");
  });

  // Проверяет, что общий текстовый ввод использует чатовый шаблон с измерителем, выделением и кареткой.
  it("renders reusable text input from shared ui kit source", () => {
    const root = document.createElement("div");
    document.body.append(root);

    dispose = render(() => <TextInput id="name-input" text="abcdef" selectionStart={1} selectionEnd={4} focused={true} className="compact-input" />, root);

    expect(root.querySelector("#name-input")?.className).toBe("ui-kit-control ui-kit-text-input compact-input is-focused");
    expect(root.querySelector(".ui-kit-text-input__text")?.textContent).toBe("abcdef");
    expect(root.querySelector(".ui-kit-text-input__selection")?.textContent).toBe("bcd");
    expect(root.querySelectorAll(".ui-kit-text-input__measure").length).toBe(2);
    expect(root.querySelector(".ui-kit-text-input__caret")).not.toBeNull();
  });

  // Проверяет, что каждое окно получает кнопку закрытия из общего шаблона модального окна.
  it("renders shared modal close button", () => {
    const root = document.createElement("div");
    document.body.append(root);

    dispose = render(() => <Modal id="modal" title="Panel">Body</Modal>, root);

    const closeButton = root.querySelector<HTMLElement>("#modal-close-button");
    expect(closeButton?.dataset.uiKind).toBe("button");
    expect(closeButton?.className).toBe("ui-kit-control ui-kit-button ui-kit-modal__close");
    expect(closeButton?.getAttribute("aria-label")).toBe("Закрыть окно");
    expect(closeButton?.textContent).toBe("");
  });

  // Проверяет, что общий шаблон окна может скрывать кнопку закрытия для окон с собственными действиями.
  it("can hide shared modal close button", () => {
    const root = document.createElement("div");
    document.body.append(root);

    dispose = render(() => <Modal id="modal" title="Panel" closeButton={false}>Body</Modal>, root);

    expect(root.querySelector("#modal-close-button")).toBeNull();
    expect(root.querySelector(".ui-kit-modal__title")?.textContent).toBe("Panel");
  });

  // Проверяет, что обычная полоса прокрутки не получает overlay-приоритет раскрытого списка.
  it("keeps ordinary scrollbar below dropdown overlay blocker", () => {
    const root = document.createElement("div");
    document.body.append(root);

    dispose = render(() => <Scrollbar id="scroll" thumbTopPercent={25} thumbHeightPercent={40} dragging={false} />, root);

    expect(root.querySelector("#scroll")?.getAttribute("data-ui-z-index")).toBeNull();
  });

  // Проверяет, что полоса прокрутки раскрытого списка остается доступной поверх его блокирующего слоя.
  it("keeps dropdown scrollbar above its overlay blocker", () => {
    const root = document.createElement("div");
    document.body.append(root);

    dispose = render(() => (
      <Dropdown
        id="select"
        open={true}
        selectedValue="b"
        options={[{ value: "a", label: "A" }, { value: "b", label: "B" }]}
        menuScroll={{ visible: true, thumbTopPercent: 10, thumbHeightPercent: 50, contentOffsetPx: 0, dragging: false }}
      />
    ), root);

    expect(document.body.querySelector("#select-outside-blocker")?.getAttribute("data-ui-z-index")).toBe("900");
    expect(document.body.querySelector("#select-scrollbar")?.getAttribute("data-ui-z-index")).toBe("1100");
  });
});
