import { createEffect, createSignal, For, onCleanup, onMount, Show, type JSX } from "solid-js";
import { Portal } from "solid-js/web";

export type UiKitOption = {
  // Значение, которое возвращает контрол при выборе.
  value: string;
  // Видимый текст пункта.
  label: string;
  // Дополнительный текст справа для двухколоночных списков.
  secondaryLabel?: string;
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
  // Дополнительный CSS-класс для специализированного размещения общего контрола.
  className?: string;
  // Доступное название, если видимая подпись является только символом.
  ariaLabel?: string;
  // Overlay-приоритет для hit-test игровым курсором.
  zIndex?: number;
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

type DropdownMenuPosition = {
  // Горизонтальная координата меню в пикселях окна.
  leftPx: number;
  // Верхняя координата меню в пикселях окна.
  topPx: number;
  // Ширина меню в пикселях окна.
  widthPx: number;
};

type DropdownMenuRects = {
  // Границы исходного поля в пикселях окна.
  rootRect: DOMRect;
  // Границы вынесенного меню в пикселях окна.
  menuRect: DOMRect | null;
  // Высота браузерного окна в пикселях.
  viewportHeightPx: number;
};

type ListBoxProps = {
  // Стабильный идентификатор списка.
  id: string;
  // Выбранное значение.
  selectedValue: string;
  // Несколько выбранных значений для списков с множественным выделением.
  selectedValues?: string[];
  // Пункты списка.
  items: UiKitOption[];
  // Вертикальный сдвиг содержимого для списков с внешней полосой прокрутки.
  scrollOffsetPx?: number;
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

type TextInputProps = {
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
  // Дополнительный CSS-класс для специализированного размещения общего поля.
  className?: string;
};

type TabsProps = {
  // Стабильный идентификатор набора вкладок.
  id: string;
  // Горизонтальное выравнивание вкладок внутри общей панели.
  align?: "start" | "center";
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
  // Overlay-приоритет для hit-test, когда полоса должна лежать поверх блокирующего слоя.
  zIndex?: number;
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
  // Видимый текст поверх шкалы, если его нужно показать.
  label?: string;
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
  <div id={props.id} data-ui-kind="button" data-ui-z-index={props.zIndex} aria-label={props.ariaLabel} class={buttonClass(props)}>{props.label}</div>
);

export const IconButton = Button;

export const Checkbox = (props: CheckboxProps) => (
  <div id={props.id} data-ui-kind="checkbox" class={`ui-kit-control ui-kit-checkbox ${props.checked ? "is-checked" : ""}`}>
    <span class="ui-kit-checkbox__mark" />
    <Show when={props.label.trim() !== ""}>
      <span>{props.label}</span>
    </Show>
  </div>
);

export const RadioGroup = (props: RadioGroupProps) => (
  <div id={props.id} data-ui-kind="radio" class="ui-kit-control ui-kit-radio">
    <For each={props.options}>
      {(option) => <span id={`${props.id}-${option.value}`} data-ui-kind="radio" data-ui-value={option.value} class={`ui-kit-control ui-kit-radio__option ${option.value === props.value ? "is-selected" : ""}`}>{option.label}</span>}
    </For>
  </div>
);

export const Dropdown = (props: DropdownProps) => {
  let rootElement: HTMLDivElement | undefined;
  let menuElement: HTMLDivElement | undefined;
  const [menuPosition, setMenuPosition] = createSignal<DropdownMenuPosition>({ leftPx: 0, topPx: 0, widthPx: 0 });

  const updateMenuPosition = () => {
    const rect = rootElement?.getBoundingClientRect();
    if (!rect) {
      return;
    }
    setMenuPosition(getDropdownMenuPosition({
      rootRect: rect,
      menuRect: menuElement?.getBoundingClientRect() ?? null,
      viewportHeightPx: window.innerHeight,
    }));
  };

  createEffect(() => {
    const open = props.open;
    const selectedValue = props.selectedValue;
    const optionCount = props.options.length;
    const scrollVisible = props.menuScroll?.visible;
    void selectedValue;
    void optionCount;
    void scrollVisible;
    if (!open) {
      return;
    }

    updateMenuPosition();
    window.addEventListener("resize", updateMenuPosition);
    onCleanup(() => window.removeEventListener("resize", updateMenuPosition));
  });

  return (
    <div ref={(element) => { rootElement = element; }} id={props.id} data-ui-kind="select" class={`ui-kit-control ui-kit-dropdown ${props.open ? "is-open" : ""}`}>
      <Show when={(props.label ?? "").trim() !== ""}>
        <div class="ui-kit-dropdown__label">{props.label}</div>
      </Show>
      <div class="ui-kit-dropdown__value">{props.options.find((option) => option.value === props.selectedValue)?.label ?? ""}</div>
      <Show when={props.open}>
        <Portal>
          {/* Экранный слой забирает внешний клик у нижних контролов и оставляет активными пункты меню. */}
          <div id={`${props.id}-outside-blocker`} data-ui-kind="modal" data-ui-z-index="900" data-ui-focusable="false" class="ui-kit-control ui-kit-dropdown__outside-blocker" />
          <div ref={(element) => { menuElement = element; }} id={`${props.id}-menu`} class="ui-kit-dropdown__menu" style={dropdownMenuStyle(menuPosition())}>
            <div class="ui-kit-dropdown__menu-viewport">
              <div class="ui-kit-dropdown__menu-clip">
                <div class="ui-kit-dropdown__menu-content" style={{ transform: `translateY(-${props.menuScroll?.contentOffsetPx ?? 0}px)` }}>
                  <For each={props.options}>
                    {(option) => <div id={`${props.id}-${option.value}`} data-ui-kind="select" data-ui-value={option.value} data-ui-z-index="1000" class={`ui-kit-control ui-kit-dropdown__item ${option.value === props.selectedValue ? "is-selected" : ""}`}>{option.label}</div>}
                  </For>
                </div>
              </div>
            </div>
            <Show when={props.menuScroll?.visible}>
              <Scrollbar
                id={`${props.id}-scrollbar`}
                className="ui-kit-dropdown-scrollbar"
                zIndex={1100}
                thumbTopPercent={props.menuScroll?.thumbTopPercent ?? 0}
                thumbHeightPercent={props.menuScroll?.thumbHeightPercent ?? 100}
                dragging={props.menuScroll?.dragging ?? false}
              />
            </Show>
          </div>
        </Portal>
      </Show>
    </div>
  );
};

export const ListBox = (props: ListBoxProps) => (
  <div id={props.id} data-ui-kind="list" class="ui-kit-control ui-kit-list">
    <div class="ui-kit-list__content" style={{ transform: `translateY(-${props.scrollOffsetPx ?? 0}px)` }}>
      <For each={props.items}>
        {(item) => (
          <div id={`${props.id}-${item.value}`} data-ui-kind="list" data-ui-value={item.value} class={`ui-kit-control ui-kit-list__item ${isListBoxItemSelected(props, item.value) ? "is-selected" : ""}`}>
            <span class="ui-kit-list__item-label">{item.label}</span>
            <Show when={item.secondaryLabel !== undefined}>
              <span class="ui-kit-list__item-secondary">{item.secondaryLabel}</span>
            </Show>
          </div>
        )}
      </For>
    </div>
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

const isListBoxItemSelected = (props: ListBoxProps, value: string): boolean =>
  props.selectedValues ? props.selectedValues.includes(value) : value === props.selectedValue;

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

export const TextInput = (props: TextInputProps) => {
  let viewportElement: HTMLDivElement | undefined;
  let textMeasureElement: HTMLSpanElement | undefined;
  let caretMeasureElement: HTMLSpanElement | undefined;
  const [metrics, setMetrics] = createSignal({ textOffsetPx: 0, caretLeftPx: 0 });
  const selectionStart = () => Math.max(0, Math.min(props.text.length, Math.min(props.selectionStart, props.selectionEnd)));
  const selectionEnd = () => Math.max(0, Math.min(props.text.length, Math.max(props.selectionStart, props.selectionEnd)));

  // Держит каретку внутри видимой области без зависимости от конкретного окна.
  const updateMetrics = () => {
    const viewportWidth = viewportElement?.getBoundingClientRect().width ?? 0;
    const textWidth = textMeasureElement?.getBoundingClientRect().width ?? 0;
    const caretWidth = caretMeasureElement?.getBoundingClientRect().width ?? 0;
    if (viewportWidth <= 0) {
      setMetrics({ textOffsetPx: 0, caretLeftPx: caretWidth });
      return;
    }

    const edgePaddingPx = Math.min(12, viewportWidth * 0.08);
    const maxOffsetPx = Math.max(0, Math.max(textWidth, caretWidth) - viewportWidth + edgePaddingPx);
    const previousOffsetPx = metrics().textOffsetPx;
    let nextOffsetPx = previousOffsetPx;
    if (caretWidth - nextOffsetPx > viewportWidth - edgePaddingPx) {
      nextOffsetPx = caretWidth - viewportWidth + edgePaddingPx;
    }
    if (caretWidth - nextOffsetPx < edgePaddingPx) {
      nextOffsetPx = caretWidth - edgePaddingPx;
    }

    const textOffsetPx = Math.max(0, Math.min(maxOffsetPx, nextOffsetPx));
    setMetrics({
      textOffsetPx,
      caretLeftPx: Math.max(0, Math.min(viewportWidth, caretWidth - textOffsetPx)),
    });
  };

  createEffect(() => {
    props.text;
    props.selectionStart;
    props.selectionEnd;
    props.focused;
    queueMicrotask(updateMetrics);
  });

  onMount(() => {
    window.addEventListener("resize", updateMetrics);
  });

  onCleanup(() => {
    window.removeEventListener("resize", updateMetrics);
  });

  return (
    <div id={props.id} data-ui-kind="edit" class={textInputClass(props)}>
      <div class="ui-kit-text-input__viewport" ref={(element) => { viewportElement = element; }}>
        <span class="ui-kit-text-input__text" style={{ transform: `translateX(${-metrics().textOffsetPx}px)` }}>
          <span>{props.text.slice(0, selectionStart())}</span>
          <Show when={selectionEnd() > selectionStart()}>
            <span class="ui-kit-text-input__selection">{props.text.slice(selectionStart(), selectionEnd())}</span>
          </Show>
          <span>{props.text.slice(selectionEnd())}</span>
        </span>
        <span class="ui-kit-text-input__measure" ref={(element) => { textMeasureElement = element; }} aria-hidden="true">{props.text}</span>
        <span class="ui-kit-text-input__measure" ref={(element) => { caretMeasureElement = element; }} aria-hidden="true">{props.text.slice(0, selectionEnd())}</span>
        <Show when={props.focused}>
          <span class="ui-kit-text-input__caret" style={{ left: `${metrics().caretLeftPx}px` }} />
        </Show>
      </div>
    </div>
  );
};

export const Tabs = (props: TabsProps) => (
  <div id={props.id} data-ui-kind="tabs" class={`ui-kit-control ui-kit-tabs ${tabsAlignmentClass(props.align)} ${props.className ?? ""}`}>
    <For each={props.tabs}>
      {(tab) => (
        <div id={`${props.itemIdPrefix ?? props.id}-${tab.value}`} data-ui-kind="tabs" data-ui-value={tab.value} class={`ui-kit-control ui-kit-tab ${markerClass(tab.marker)} ${props.itemClassName ?? ""} ${tab.value === props.selectedValue ? "is-selected" : ""}`}>
          <Show when={tab.marker}>{(marker) => <span class="ui-kit-tab__marker">{marker()}</span>}</Show>
          <span class="ui-kit-tab__label">{tab.label}</span>
          <Show when={tab.badge}>{(badge) => <span class="ui-kit-tab__badge">{badge()}</span>}</Show>
        </div>
      )}
    </For>
  </div>
);

export const Scrollbar = (props: ScrollbarProps) => (
  <div id={props.id} data-ui-kind="scrollbar" data-ui-z-index={props.zIndex} class={`ui-kit-control ui-kit-scrollbar ${props.className ?? ""} ${props.dragging ? "is-dragging" : ""}`}>
    <div class="ui-kit-scrollbar__thumb" style={{ top: `${props.thumbTopPercent}%`, height: `${props.thumbHeightPercent}%` }} />
  </div>
);

export const Slider = (props: SliderProps) => {
  const percent = () => `${Math.max(0, Math.min(100, ((props.value - props.min) / Math.max(1, props.max - props.min)) * 100))}%`;
  return (
    <div id={props.id} data-ui-kind="slider" class="ui-kit-control ui-kit-slider">
      <div class="ui-kit-slider__fill" style={{ width: percent() }} />
      <Show when={(props.label ?? "").trim() !== ""}>
        <div class="ui-kit-slider__label">{props.label}</div>
      </Show>
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
    <Button id={`${props.id}-close-button`} label="" className="ui-kit-modal__close" ariaLabel="Закрыть окно" zIndex={1200} />
    <div class="ui-kit-modal__body">{props.children}</div>
  </div>
);

export const Tooltip = (props: TooltipProps) => (
  <div id={props.id} data-ui-kind="tooltip" class="ui-kit-control ui-kit-tooltip">{props.text}</div>
);

export const HotkeyCapture = EditControl;

const stateClass = (state: ButtonProps["state"]): string => state && state !== "normal" ? `is-${state}` : "";

const buttonClass = (props: ButtonProps): string => ["ui-kit-control", "ui-kit-button", props.className, stateClass(props.state)].filter(Boolean).join(" ");

const textInputClass = (props: TextInputProps): string => ["ui-kit-control", "ui-kit-text-input", props.className, props.focused ? "is-focused" : ""].filter(Boolean).join(" ");

// Дополнительный класс нужен только для вкладок, где левый край выравнивается по квадратному значку.
const markerClass = (marker: JSX.Element | undefined): string => marker ? "ui-kit-tab--with-marker" : "";

const tabsAlignmentClass = (align: TabsProps["align"]): string => align === "center" ? "ui-kit-tabs--center" : "";

const dropdownMenuStyle = (position: DropdownMenuPosition): JSX.CSSProperties => ({
  left: `${position.leftPx}px`,
  top: `${position.topPx}px`,
  width: `${position.widthPx}px`,
});

const getDropdownMenuPosition = (rects: DropdownMenuRects): DropdownMenuPosition => {
  const menuHeightPx = rects.menuRect?.height ?? 0;
  const belowTopPx = rects.rootRect.bottom;
  const aboveTopPx = rects.rootRect.top - menuHeightPx;
  const topPx = menuHeightPx > 0 && belowTopPx + menuHeightPx > rects.viewportHeightPx
    ? Math.max(0, aboveTopPx)
    : belowTopPx;

  return {
    leftPx: rects.rootRect.left,
    topPx,
    widthPx: rects.rootRect.width,
  };
};

const flattenTree = (nodes: TreeNode[]): TreeNode[] => nodes.flatMap((node) => [node, ...flattenTree(node.children ?? [])]);
