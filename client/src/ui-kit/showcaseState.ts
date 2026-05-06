import type { GameUiAction } from "./types";
import { getScrollbarThumbTopPercentFromCursor, startScrollbarDrag, type ScrollbarDragState } from "./scrollbar";

export type UiKitDemoState = {
  // Количество подтвержденных кликов по обычной кнопке.
  buttonClicks: number;
  // Состояние демонстрационной галочки.
  checkboxChecked: boolean;
  // Выбранный пункт демонстрационной группы переключателей.
  radioValue: string;
  // Показывает, раскрыт ли демонстрационный выпадающий список.
  dropdownOpen: boolean;
  // Выбранный пункт демонстрационного выпадающего списка.
  dropdownValue: string;
  // Выбранный пункт демонстрационного обычного списка.
  listValue: string;
  // Выбранный узел демонстрационного дерева.
  treeValue: string;
  // Индекс первого пункта демонстрационного виртуального списка.
  virtualStartIndex: number;
  // Выбранная демонстрационная вкладка.
  tabValue: string;
  // Текст демонстрационного поля ввода.
  editText: string;
  // Начало выделения в демонстрационном поле ввода.
  editSelectionStart: number;
  // Конец выделения в демонстрационном поле ввода.
  editSelectionEnd: number;
  // Верх демонстрационного ползунка прокрутки в процентах.
  scrollbarTopPercent: number;
  // Состояние захвата демонстрационного ползунка прокрутки.
  scrollbarDrag: ScrollbarDragState | null;
  // Значение демонстрационного слайдера.
  sliderValue: number;
  // Значение демонстрационного числового поля.
  stepperValue: number;
  // Ориентация демонстрационного разделителя.
  splitterVertical: boolean;
  // Показывает, раскрыто ли демонстрационное меню.
  menuOpen: boolean;
  // Показывает, активна ли демонстрационная подсказка.
  tooltipVisible: boolean;
};

// Создает независимый снимок начального состояния витрины.
export const createInitialUiKitDemoState = (): UiKitDemoState => ({
  buttonClicks: 0,
  checkboxChecked: true,
  radioValue: "b",
  dropdownOpen: true,
  dropdownValue: "two",
  listValue: "2",
  treeValue: "child",
  virtualStartIndex: 20,
  tabValue: "two",
  editText: "Selected text",
  editSelectionStart: 0,
  editSelectionEnd: 8,
  scrollbarTopPercent: 20,
  scrollbarDrag: null,
  sliderValue: 45,
  stepperValue: 7,
  splitterVertical: true,
  menuOpen: true,
  tooltipVisible: true,
});

// Применяет игровое действие к витрине без зависимости от Phaser или SolidJS.
export const applyUiKitDemoAction = (state: UiKitDemoState, action: GameUiAction): UiKitDemoState => {
  if (action.type !== "click" && action.type !== "dragStart" && action.type !== "dragMove" && action.type !== "dragEnd") {
    return state;
  }

  if (action.controlId === "ui-kit-demo-button") {
    return { ...state, buttonClicks: state.buttonClicks + 1 };
  }
  if (action.controlId === "ui-kit-demo-icon-button") {
    return { ...state, tooltipVisible: !state.tooltipVisible };
  }
  if (action.controlId === "ui-kit-demo-checkbox") {
    return { ...state, checkboxChecked: !state.checkboxChecked };
  }
  if (action.controlId.startsWith("ui-kit-demo-radio-") && typeof action.value === "string") {
    return { ...state, radioValue: action.value };
  }
  if (action.controlId === "ui-kit-demo-select") {
    return { ...state, dropdownOpen: !state.dropdownOpen };
  }
  if (action.controlId.startsWith("ui-kit-demo-select-") && typeof action.value === "string") {
    return { ...state, dropdownValue: action.value, dropdownOpen: false };
  }
  if (action.controlId.startsWith("ui-kit-demo-list-") && typeof action.value === "string") {
    return { ...state, listValue: action.value };
  }
  if (action.controlId.startsWith("ui-kit-demo-tree-") && typeof action.value === "string") {
    return { ...state, treeValue: action.value };
  }
  if (action.controlId === "ui-kit-demo-virtual-list") {
    return { ...state, virtualStartIndex: state.virtualStartIndex >= 40 ? 0 : state.virtualStartIndex + 10 };
  }
  if (action.controlId.startsWith("ui-kit-demo-virtual-list-") && typeof action.value === "string") {
    return { ...state, listValue: action.value };
  }
  if (action.controlId.startsWith("ui-kit-demo-tabs-") && typeof action.value === "string") {
    return { ...state, tabValue: action.value };
  }
  if (action.controlId === "ui-kit-demo-edit") {
    const selectedFirstWord = state.editSelectionStart === 0 && state.editSelectionEnd === 8;
    return { ...state, editSelectionStart: selectedFirstWord ? 9 : 0, editSelectionEnd: selectedFirstWord ? state.editText.length : 8 };
  }
  if (action.controlId === "ui-kit-demo-scrollbar" || action.kind === "scrollbar") {
    if (!action.controlRect) {
      return state;
    }
    const track = { top: action.controlRect.top, height: action.controlRect.height };
    if (action.type === "dragStart") {
      return {
        ...state,
        scrollbarDrag: startScrollbarDrag({ ...track, thumbTopPercent: state.scrollbarTopPercent, thumbHeightPercent: 45 }, action.y),
      };
    }
    if (action.type === "dragEnd") {
      return { ...state, scrollbarDrag: null };
    }
    if (action.type === "dragMove" && state.scrollbarDrag) {
      return {
        ...state,
        scrollbarTopPercent: getScrollbarThumbTopPercentFromCursor({ ...track, thumbHeightPercent: 45, drag: state.scrollbarDrag }, action.y),
      };
    }
    return state;
  }
  if (action.controlId === "ui-kit-demo-slider" || action.kind === "slider") {
    return { ...state, sliderValue: nextPercent(state.sliderValue, 10) };
  }
  if (action.controlId === "ui-kit-demo-stepper-decrement") {
    return { ...state, stepperValue: state.stepperValue - 1 };
  }
  if (action.controlId === "ui-kit-demo-stepper" || action.controlId === "ui-kit-demo-stepper-increment") {
    return { ...state, stepperValue: state.stepperValue + 1 };
  }
  if (action.controlId === "ui-kit-demo-splitter") {
    return { ...state, splitterVertical: !state.splitterVertical };
  }
  if (action.controlId === "ui-kit-demo-menu" || action.controlId === "ui-kit-demo-menu-close") {
    return { ...state, menuOpen: !state.menuOpen };
  }
  if (action.controlId === "ui-kit-demo-tooltip") {
    return { ...state, tooltipVisible: !state.tooltipVisible };
  }

  return state;
};

const nextPercent = (value: number, delta: number): number => (value + delta > 100 ? 0 : value + delta);
