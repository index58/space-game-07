import { For, Show, type JSX } from "solid-js";

export type UiKitOption = {
  // Значение, которое возвращает контрол при выборе.
  value: string;
  // Видимый текст пункта.
  label: string;
};

type TreeNode = UiKitOption & {
  // Дочерние пункты дерева.
  children?: TreeNode[];
};

type ButtonProps = {
  // Стабильный идентификатор контрола.
  id: string;
  // Видимая подпись.
  label: string;
  // Визуальное состояние для отладки и HUD.
  state?: "normal" | "hovered" | "pressed" | "focused" | "disabled" | "error";
};

type CheckboxProps = {
  // Стабильный идентификатор контрола.
  id: string;
  // Видимая подпись.
  label: string;
  // Текущее логическое значение.
  checked: boolean;
};

type RadioGroupProps = {
  // Стабильный идентификатор группы.
  id: string;
  // Выбранное значение.
  value: string;
  // Доступные варианты.
  options: UiKitOption[];
};

export type DropdownProps = {
  // Стабильный идентификатор выпадающего списка.
  id: string;
  // Видимая подпись поля. Если не задана, строка подписи не рисуется.
  label?: string;
  // Показывает раскрытый список.
  open: boolean;
  // Выбранное значение.
  selectedValue: string;
  // Доступные варианты.
  options: UiKitOption[];
  // Состояние прокрутки раскрытого меню.
  menuScroll?: {
    // Показывает, что пункты не помещаются по высоте.
    visible: boolean;
    // Верх ползунка в процентах высоты полосы.
    thumbTopPercent: number;
    // Высота ползунка в процентах высоты полосы.
    thumbHeightPercent: number;
    // Вертикальный сдвиг списка пунктов.
    contentOffsetPx: number;
    // Показывает активное перетаскивание ползунка.
    dragging: boolean;
  };
};

type ListBoxProps = {
  // Стабильный идентификатор списка.
  id: string;
  // Выбранное значение.
  selectedValue: string;
  // Пункты списка.
  items: UiKitOption[];
};

type TreeViewProps = {
  // Стабильный идентификатор дерева.
  id: string;
  // Выбранное значение.
  selectedValue: string;
  // Корневые узлы.
  nodes: TreeNode[];
};

type VirtualListProps = {
  // Стабильный идентификатор виртуального списка.
  id: string;
  // Индекс первого видимого пункта в полном наборе.
  startIndex: number;
  // Видимые пункты окна.
  items: UiKitOption[];
};

type EditControlProps = {
  // Стабильный идентификатор поля.
  id: string;
  // Видимый текст.
  text: string;
  // Начало выделения.
  selectionStart: number;
  // Конец выделения.
  selectionEnd: number;
  // Признак активного фокуса.
  focused: boolean;
};

type TabsProps = {
  // Стабильный идентификатор набора вкладок.
  id: string;
  // Дополнительный CSS-класс корня для конкретной панели.
  className?: string;
  // Префикс DOM-идентификатора отдельной вкладки.
  itemIdPrefix?: string;
  // Дополнительный CSS-класс отдельной вкладки.
  itemClassName?: string;
  // Выбранная вкладка.
  selectedValue: string;
  // Доступные вкладки.
  tabs: Array<UiKitOption & {
    // Короткий маркер перед названием вкладки.
    marker?: JSX.Element;
    // Количественный индикатор в общем стиле UI Kit.
    badge?: number;
  }>;
};

type ScrollbarProps = {
  // Стабильный идентификатор полосы.
  id: string;
  // Дополнительный CSS-класс корня для конкретной панели.
  className?: string;
  // Верх ползунка в процентах.
  thumbTopPercent: number;
  // Высота ползунка в процентах.
  thumbHeightPercent: number;
  // Признак перетаскивания.
  dragging: boolean;
};

type SliderProps = {
  // Стабильный идентификатор слайдера.
  id: string;
  // Текущее значение.
  value: number;
  // Минимальное значение.
  min: number;
  // Максимальное значение.
  max: number;
};

type NumericStepperProps = {
  // Стабильный идентификатор степпера.
  id: string;
  // Текущее числовое значение.
  value: number;
};

type SplitterProps = {
  // Стабильный идентификатор разделителя.
  id: string;
  // Вертикальная ориентация разделителя.
  vertical: boolean;
};

type ContextMenuProps = {
  // Стабильный идентификатор меню.
  id: string;
  // Пункты меню.
  items: UiKitOption[];
};

type ModalProps = {
  // Стабильный идентификатор окна.
  id: string;
  // Заголовок окна.
  title: string;
  // Содержимое окна.
  children: JSX.Element;
};

type TooltipProps = {
  // Стабильный идентификатор подсказки.
  id: string;
  // Текст подсказки.
  text: string;
};

export const Button = (props: ButtonProps) => (
  <div id={props.id} data-ui-kind="button" class={`ui-kit-control ui-kit-button ${stateClass(props.state)}`}>{props.label}</div>
);

export const IconButton = Button;

export const Checkbox = (props: CheckboxProps) => (
  <div id={props.id} data-ui-kind="checkbox" class={`ui-kit-control ui-kit-checkbox ${props.checked ? "is-checked" : ""}`}>
    <span class="ui-kit-checkbox__mark">{props.checked ? "✓" : ""}</span>
    <span>{props.label}</span>
  </div>
);

export const RadioGroup = (props: RadioGroupProps) => (
  <div id={props.id} data-ui-kind="radio" class="ui-kit-control ui-kit-radio">
    <For each={props.options}>
      {(option) => <span id={`${props.id}-${option.value}`} data-ui-kind="radio" data-ui-value={option.value} class={`ui-kit-control ui-kit-radio__option ${option.value === props.value ? "is-selected" : ""}`}>{option.label}</span>}
    </For>
  </div>
);

export const Dropdown = (props: DropdownProps) => (
  <div id={props.id} data-ui-kind="select" class={`ui-kit-control ui-kit-dropdown ${props.open ? "is-open" : ""}`}>
    <Show when={(props.label ?? "").trim() !== ""}>
      <div class="ui-kit-dropdown__label">{props.label}</div>
    </Show>
    <div class="ui-kit-dropdown__value">{props.options.find((option) => option.value === props.selectedValue)?.label ?? ""}</div>
    <Show when={props.open}>
      <div class="ui-kit-dropdown__menu">
        <div class="ui-kit-dropdown__menu-viewport">
          <div class="ui-kit-dropdown__menu-content" style={{ transform: `translateY(-${props.menuScroll?.contentOffsetPx ?? 0}px)` }}>
            <For each={props.options}>
              {(option) => <div id={`${props.id}-${option.value}`} data-ui-kind="select" data-ui-value={option.value} data-ui-z-index="1000" class={`ui-kit-control ui-kit-dropdown__item ${option.value === props.selectedValue ? "is-selected" : ""}`}>{option.label}</div>}
            </For>
          </div>
        </div>
        <Show when={props.menuScroll?.visible}>
          <Scrollbar
            id={`${props.id}-scrollbar`}
            className="ui-kit-dropdown-scrollbar"
            thumbTopPercent={props.menuScroll?.thumbTopPercent ?? 0}
            thumbHeightPercent={props.menuScroll?.thumbHeightPercent ?? 100}
            dragging={props.menuScroll?.dragging ?? false}
          />
        </Show>
      </div>
    </Show>
  </div>
);

export const ListBox = (props: ListBoxProps) => (
  <div id={props.id} data-ui-kind="list" class="ui-kit-control ui-kit-list">
    <For each={props.items}>
      {(item) => <div id={`${props.id}-${item.value}`} data-ui-kind="list" data-ui-value={item.value} class={`ui-kit-control ui-kit-list__item ${item.value === props.selectedValue ? "is-selected" : ""}`}>{item.label}</div>}
    </For>
  </div>
);

export const TreeView = (props: TreeViewProps) => (
  <div id={props.id} data-ui-kind="tree" class="ui-kit-control ui-kit-tree">
    <For each={flattenTree(props.nodes)}>
      {(node) => <div id={`${props.id}-${node.value}`} data-ui-kind="tree" data-ui-value={node.value} class={`ui-kit-control ui-kit-tree__item ${node.value === props.selectedValue ? "is-selected" : ""}`}>{node.label}</div>}
    </For>
  </div>
);

export const VirtualList = (props: VirtualListProps) => (
  <div id={props.id} data-ui-kind="list" class="ui-kit-control ui-kit-virtual-list">
    <For each={props.items}>
      {(item, index) => (
        <div id={`${props.id}-${item.value}`} data-ui-kind="list" data-ui-value={item.value} class="ui-kit-control ui-kit-virtual-list__item">
          <span class="ui-kit-virtual-list__index">{props.startIndex + index()}</span>
          <span>{item.label}</span>
        </div>
      )}
    </For>
  </div>
);

export const EditControl = (props: EditControlProps) => (
  <div id={props.id} data-ui-kind="edit" class={`ui-kit-control ui-kit-edit ${props.focused ? "is-focused" : ""}`}>
    <span class="ui-kit-edit__text">{props.text.slice(0, props.selectionStart)}</span>
    <span class="ui-kit-edit__selection">{props.text.slice(props.selectionStart, props.selectionEnd)}</span>
    <span class="ui-kit-edit__text">{props.text.slice(props.selectionEnd)}</span>
    <Show when={props.focused}>
      <span class="ui-kit-edit__caret" />
    </Show>
  </div>
);

export const Tabs = (props: TabsProps) => (
  <div id={props.id} data-ui-kind="tabs" class={`ui-kit-control ui-kit-tabs ${props.className ?? ""}`}>
    <For each={props.tabs}>
      {(tab) => (
        <div id={`${props.itemIdPrefix ?? props.id}-${tab.value}`} data-ui-kind="tabs" data-ui-value={tab.value} class={`ui-kit-control ui-kit-tab ${props.itemClassName ?? ""} ${tab.value === props.selectedValue ? "is-selected" : ""}`}>
          <Show when={tab.marker}>{(marker) => <span class="ui-kit-tab__marker">{marker()}</span>}</Show>
          <span class="ui-kit-tab__label">{tab.label}</span>
          <Show when={tab.badge}>{(badge) => <span class="ui-kit-tab__badge">{badge()}</span>}</Show>
        </div>
      )}
    </For>
  </div>
);

export const Scrollbar = (props: ScrollbarProps) => (
  <div id={props.id} data-ui-kind="scrollbar" data-ui-z-index="1100" class={`ui-kit-control ui-kit-scrollbar ${props.className ?? ""} ${props.dragging ? "is-dragging" : ""}`}>
    <div class="ui-kit-scrollbar__thumb" style={{ top: `${props.thumbTopPercent}%`, height: `${props.thumbHeightPercent}%` }} />
  </div>
);

export const Slider = (props: SliderProps) => {
  const percent = () => `${Math.max(0, Math.min(100, ((props.value - props.min) / Math.max(1, props.max - props.min)) * 100))}%`;
  return (
    <div id={props.id} data-ui-kind="slider" class="ui-kit-control ui-kit-slider">
      <div class="ui-kit-slider__fill" style={{ width: percent() }} />
    </div>
  );
};

export const NumericStepper = (props: NumericStepperProps) => (
  <div id={props.id} data-ui-kind="stepper" class="ui-kit-control ui-kit-stepper"><span id={`${props.id}-decrement`} data-ui-kind="stepper" data-ui-value="decrement" class="ui-kit-control">-</span><span>{props.value}</span><span id={`${props.id}-increment`} data-ui-kind="stepper" data-ui-value="increment" class="ui-kit-control">+</span></div>
);

export const Splitter = (props: SplitterProps) => (
  <div id={props.id} data-ui-kind="splitter" class={`ui-kit-control ui-kit-splitter ${props.vertical ? "is-vertical" : "is-horizontal"}`} />
);

export const ContextMenu = (props: ContextMenuProps) => (
  <div id={props.id} data-ui-kind="menu" class="ui-kit-control ui-kit-context-menu">
    <For each={props.items}>{(item) => <div id={`${props.id}-${item.value}`} data-ui-kind="menu" data-ui-value={item.value} class="ui-kit-control ui-kit-context-menu__item">{item.label}</div>}</For>
  </div>
);

export const Modal = (props: ModalProps) => (
  <div id={props.id} data-ui-kind="modal" class="ui-kit-control ui-kit-modal">
    <div class="ui-kit-modal__title">{props.title}</div>
    <div class="ui-kit-modal__body">{props.children}</div>
  </div>
);

export const Tooltip = (props: TooltipProps) => (
  <div id={props.id} data-ui-kind="tooltip" class="ui-kit-control ui-kit-tooltip">{props.text}</div>
);

export const HotkeyCapture = EditControl;

const stateClass = (state: ButtonProps["state"]): string => state && state !== "normal" ? `is-${state}` : "";

const flattenTree = (nodes: TreeNode[]): TreeNode[] => nodes.flatMap((node) => [node, ...flattenTree(node.children ?? [])]);
