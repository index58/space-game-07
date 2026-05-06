import { render } from "solid-js/web";
import { afterEach, describe, expect, it } from "vitest";
import { Button, Checkbox, ContextMenu, Dropdown, EditControl, ListBox, Modal, NumericStepper, RadioGroup, Scrollbar, Slider, Splitter, Tabs, Tooltip, TreeView, VirtualList } from "./components";
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
        <ListBox id="list" selectedValue="2" items={[{ value: "1", label: "One" }, { value: "2", label: "Two" }]} />
        <TreeView id="tree" selectedValue="child" nodes={[{ value: "root", label: "Root", children: [{ value: "child", label: "Child" }] }]} />
        <VirtualList id="virtual" startIndex={10} items={Array.from({ length: 3 }, (_, index) => ({ value: String(index), label: `Item ${index}` }))} />
      </>
    ), root);

    expect(root.querySelector(".ui-kit-dropdown__menu")).not.toBeNull();
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

    expect(root.querySelector<HTMLElement>("#select-a")?.dataset.uiZIndex).toBe("1000");
    expect(root.querySelector<HTMLElement>("#select-b")?.dataset.uiZIndex).toBe("1000");
  });

  // Проверяет, что пустая подпись выпадающего списка не занимает отдельную строку.
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
        <Tabs id="tabs" selectedValue="chat" tabs={[{ value: "map", label: "Map" }, { value: "chat", label: "Chat", badge: 2 }]} />
        <Scrollbar id="scroll" thumbTopPercent={25} thumbHeightPercent={40} dragging={true} />
        <Slider id="slider" value={40} min={0} max={100} />
        <NumericStepper id="stepper" value={7} />
        <Splitter id="splitter" vertical={true} />
        <ContextMenu id="menu" items={[{ value: "close", label: "Close" }]} />
        <Modal id="modal" title="Panel">Body</Modal>
        <Tooltip id="tip" text="Hint" />
      </>
    ), root);

    expect(root.querySelector(".ui-kit-edit__selection")?.textContent).toBe("bcd");
    expect(root.querySelector(".ui-kit-tab.is-selected")?.textContent).toContain("Chat");
    expect(root.querySelector(".ui-kit-scrollbar.is-dragging")).not.toBeNull();
    expect(root.querySelector<HTMLElement>(".ui-kit-slider__fill")?.style.width).toBe("40%");
    expect(root.querySelector(".ui-kit-stepper")?.textContent).toContain("7");
    expect(root.querySelector(".ui-kit-splitter.is-vertical")).not.toBeNull();
    expect(root.querySelector(".ui-kit-context-menu__item")?.textContent).toBe("Close");
    expect(root.querySelector(".ui-kit-modal__title")?.textContent).toBe("Panel");
    expect(root.querySelector(".ui-kit-tooltip")?.textContent).toBe("Hint");
  });
});
