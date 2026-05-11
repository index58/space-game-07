export type GameUiControlKind =
  | "edit"
  | "button"
  | "checkbox"
  | "radio"
  | "select"
  | "list"
  | "tree"
  | "tabs"
  | "menu"
  | "modal"
  | "tooltip"
  | "scrollbar"
  | "slider"
  | "stepper"
  | "hotkey"
  | "splitter"
  | "dragItem";

export type GameUiRect = {
  // Левая граница контрола в пикселях окна.
  left: number;
  // Верхняя граница контрола в пикселях окна.
  top: number;
  // Ширина контрола в пикселях окна.
  width: number;
  // Высота контрола в пикселях окна.
  height: number;
};

export type GameUiControlState = {
  // Стабильный идентификатор контрола внутри игрового UI.
  id: string;
  // Тип поведения, по которому runtime выбирает действие.
  kind: GameUiControlKind;
  // Экранная область для hit-test игровым курсором.
  rect: GameUiRect;
  // Порядок перекрытия, где большее значение находится выше.
  zIndex: number;
  // Признак недоступности для фокуса и действий.
  disabled: boolean;
  // Признак участия в отрисовке и hit-test.
  visible: boolean;
  // Признак возможности получить клавиатурный фокус.
  focusable: boolean;
  // Текущее значение контрола, если оно есть.
  value: unknown;
};

export type GameUiActionType =
  | "click"
  | "change"
  | "submit"
  | "cancel"
  | "select"
  | "open"
  | "close"
  | "dragStart"
  | "dragMove"
  | "dragEnd";

export type GameUiAction = {
  // Тип пользовательского действия, полученного игровым UI.
  type: GameUiActionType;
  // Контрол, к которому относится действие.
  controlId: string;
  // Тип контрола на момент генерации действия.
  kind: GameUiControlKind;
  // Горизонтальная координата игрового курсора.
  x: number;
  // Вертикальная координата игрового курсора.
  y: number;
  // Значение контрола на момент действия, если оно задано.
  value?: unknown;
  // Признак зажатой клавиши Ctrl во время действия мышью.
  ctrlKey?: boolean;
  // Признак зажатой клавиши Meta во время действия мышью.
  metaKey?: boolean;
  // Признак зажатой клавиши Shift во время действия мышью.
  shiftKey?: boolean;
  // Экранная область контрола на момент действия, если runtime её знает.
  controlRect?: GameUiRect;
};

export type TextEditState = {
  // Текст, который хранит скрытый браузерный редактор.
  text: string;
  // Начало выделения в UTF-16 индексах строки.
  selectionStart: number;
  // Конец выделения в UTF-16 индексах строки.
  selectionEnd: number;
  // Направление выделения, как у браузерного поля ввода.
  selectionDirection: "forward" | "backward" | "none";
  // Горизонтальный сдвиг визуального viewport.
  scrollX: number;
  // Вертикальный сдвиг визуального viewport.
  scrollY: number;
  // Признак активного текстового ввода.
  focused: boolean;
};
