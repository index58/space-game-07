import { beforeEach, describe, expect, it } from "vitest";
import { TextEditController } from "./textEdit";

beforeEach(() => {
  document.body.innerHTML = "";
});

describe("TextEditController", () => {
  // Проверяет, что однострочный режим убирает переносы и оставляет браузерное выделение источником правды.
  it("keeps single line text and selection in a hidden textarea", () => {
    const edit = new TextEditController({ id: "chat", mode: "singleLine" });

    edit.focus("hello\nworld", 5, 5);
    edit.replaceSelection("\n!");

    expect(edit.snapshot()).toMatchObject({
      text: "hello!world",
      selectionStart: 6,
      selectionEnd: 6,
      focused: true,
    });
  });

  // Проверяет, что браузерные Home и End видят одно поле как одну строку, а не как узкую колонку символов.
  it("disables native wrapping for single line editing", () => {
    const edit = new TextEditController({ id: "chat", mode: "singleLine" });

    edit.focus("hello world", 5, 5);

    expect(edit.element().wrap).toBe("off");
    expect(getComputedStyle(edit.element()).whiteSpace).toBe("pre");
    expect(getComputedStyle(edit.element()).width).toBe("1000px");
  });

  // Проверяет, что многострочный режим сохраняет переносы строк.
  it("keeps new lines in multiline mode", () => {
    const edit = new TextEditController({ id: "notes", mode: "multiLine" });

    edit.focus("alpha", 5, 5);
    edit.replaceSelection("\nbeta");

    expect(edit.snapshot().text).toBe("alpha\nbeta");
  });

  // Проверяет, что обычный ввод заменяет выделенный диапазон.
  it("replaces selected range with inserted text", () => {
    const edit = new TextEditController({ id: "field", mode: "singleLine" });

    edit.focus("abcdef", 2, 5);
    edit.replaceSelection("XY");

    expect(edit.snapshot()).toMatchObject({
      text: "abXYf",
      selectionStart: 4,
      selectionEnd: 4,
    });
  });

  // Проверяет, что выделение слова использует границы вокруг позиции игрового курсора.
  it("selects word around requested index", () => {
    const edit = new TextEditController({ id: "field", mode: "singleLine" });

    edit.focus("hello brave world", 8, 8);
    edit.selectWordAt(8);

    expect(edit.snapshot()).toMatchObject({
      selectionStart: 6,
      selectionEnd: 11,
    });
  });

  // Проверяет, что состояние буфера прокрутки можно синхронизировать с визуальным HUD.
  it("keeps visual scroll offsets in state", () => {
    const edit = new TextEditController({ id: "field", mode: "multiLine" });

    edit.focus("text", 0, 0);
    edit.setScroll(12, 34);

    expect(edit.snapshot()).toMatchObject({ scrollX: 12, scrollY: 34 });
  });
});
