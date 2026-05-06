import type { TextEditState } from "./types";

export type TextEditMode = "singleLine" | "multiLine";

export type TextEditControllerOptions = {
  // Стабильный идентификатор редактора внутри игрового UI.
  id: string;
  // Режим обработки переносов строк.
  mode: TextEditMode;
};

// Оборачивает скрытый браузерный textarea, чтобы HUD получал нативное редактирование текста.
export class TextEditController {
  // Скрытое поле, которому браузер отдаёт ввод, выделение, буфер и IME.
  private readonly textarea: HTMLTextAreaElement;
  // Текущий визуальный сдвиг по горизонтали.
  private scrollX = 0;
  // Текущий визуальный сдвиг по вертикали.
  private scrollY = 0;
  // Признак активного редактора.
  private focused = false;

  constructor(
    // Настройки конкретного игрового поля.
    private readonly options: TextEditControllerOptions,
  ) {
    this.textarea = document.createElement("textarea");
    this.textarea.dataset.uiKitEditId = options.id;
    this.textarea.className = "ui-kit-hidden-textarea";
    this.textarea.wrap = options.mode === "singleLine" ? "off" : "soft";
    this.textarea.style.whiteSpace = options.mode === "singleLine" ? "pre" : "pre-wrap";
    this.textarea.style.width = options.mode === "singleLine" ? "1000px" : "1px";
    this.textarea.setAttribute("aria-hidden", "true");
    this.textarea.tabIndex = -1;
    document.body.append(this.textarea);
  }

  // Делает редактор активным и синхронизирует начальное значение.
  focus(text: string, selectionStart = text.length, selectionEnd = selectionStart): void {
    this.textarea.value = this.normalizeText(text);
    this.focused = true;
    this.textarea.focus();
    this.setSelection(selectionStart, selectionEnd);
  }

  // Снимает фокус без удаления текста.
  blur(): void {
    this.focused = false;
    this.textarea.blur();
  }

  // Заменяет текущее выделение новым текстом.
  replaceSelection(text: string): void {
    const inserted = this.normalizeText(text);
    const start = this.textarea.selectionStart;
    const end = this.textarea.selectionEnd;
    this.textarea.value = `${this.textarea.value.slice(0, start)}${inserted}${this.textarea.value.slice(end)}`;
    const next = start + inserted.length;
    this.setSelection(next, next);
  }

  // Выделяет слово вокруг индекса, используя обычные пробельные границы.
  selectWordAt(index: number): void {
    const text = this.textarea.value;
    const safeIndex = clamp(index, 0, text.length);
    let start = safeIndex;
    let end = safeIndex;
    while (start > 0 && !/\s/.test(text[start - 1])) {
      start -= 1;
    }
    while (end < text.length && !/\s/.test(text[end])) {
      end += 1;
    }
    this.setSelection(start, end);
  }

  // Сохраняет визуальный сдвиг, рассчитанный компонентом HUD.
  setScroll(scrollX: number, scrollY: number): void {
    this.scrollX = Math.max(0, scrollX);
    this.scrollY = Math.max(0, scrollY);
  }

  // Возвращает снимок, который можно отрисовать в SolidJS.
  snapshot(): TextEditState {
    return {
      text: this.textarea.value,
      selectionStart: this.textarea.selectionStart,
      selectionEnd: this.textarea.selectionEnd,
      selectionDirection: this.textarea.selectionDirection ?? "none",
      scrollX: this.scrollX,
      scrollY: this.scrollY,
      focused: this.focused,
    };
  }

  // Даёт владельцу доступ к браузерному полю для подписки на native-события.
  element(): HTMLTextAreaElement {
    return this.textarea;
  }

  // Удаляет скрытый textarea при уничтожении владельца.
  dispose(): void {
    this.textarea.remove();
  }

  private setSelection(selectionStart: number, selectionEnd: number): void {
    const start = Math.min(selectionStart, selectionEnd);
    const end = Math.max(selectionStart, selectionEnd);
    this.textarea.setSelectionRange(clamp(start, 0, this.textarea.value.length), clamp(end, 0, this.textarea.value.length));
  }

  private normalizeText(text: string): string {
    if (this.options.mode === "multiLine") {
      return text;
    }
    return text.replace(/\r?\n/g, "");
  }
}

const clamp = (value: number, min: number, max: number): number => Math.min(max, Math.max(min, value));
