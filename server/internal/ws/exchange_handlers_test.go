package ws

import "testing"

// Проверяет, что технические причины отказа обмена показываются игроку по-русски.
func TestExchangeErrorTextReturnsRussianMessage(t *testing.T) {
	cases := map[string]string{
		"source container is not selected":        "Контейнер-источник не выбран",
		"exchange amount is empty":                "Количество не выбрано",
		"object already participates in exchange": "Объект уже участвует в обмене",
	}

	for input, expected := range cases {
		if actual := exchangeErrorText(input); actual != expected {
			t.Fatalf("unexpected exchange error text for %q: %q", input, actual)
		}
	}
}

// Проверяет, что технические причины отказа стыковки показываются игроку по-русски.
func TestDockingErrorTextReturnsRussianMessage(t *testing.T) {
	cases := map[string]string{
		"object participates in exchange": "Объект участвует в обмене",
		"object is not docked":            "Объект не пристыкован",
	}

	for input, expected := range cases {
		if actual := dockingErrorText(input); actual != expected {
			t.Fatalf("unexpected docking error text for %q: %q", input, actual)
		}
	}
}
