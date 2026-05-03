package telegram

import (
	"strings"
	"testing"
)

func TestTelegramEscapeSpecialSymbols(t *testing.T) {
	input := "boil_1.t_(out)!"
	escaped := TelegramEscape(input)

	for _, fragment := range []string{`\_`, `\.`, `\(`, `\)`, `\!`} {
		if !strings.Contains(escaped, fragment) {
			t.Fatalf("expected escaped fragment %q in %q", fragment, escaped)
		}
	}
}

func TestTelegramEscapeExactSingleBackslash(t *testing.T) {
	input := "a.b_c"
	escaped := TelegramEscape(input)
	expected := "a\\.b\\_c"

	if escaped != expected {
		t.Fatalf("expected %q, got %q", expected, escaped)
	}
}
