package normalize

import "testing"

func TestNormalize(t *testing.T) {
	tests := []struct {
		name string
		in   string
		want string
	}{
		{"persian passthrough", "مبانی آنالیزعددی", "مبانی آنالیزعددی"},
		{"arabic keheh to kaf", "\u0643د درس", "\u06a9د درس"},
		{"arabic yeh to yeh", "\u064a", "\u06cc"},
		{"yeh barree to heh+hamza", "ۀ", "هٔ"},
		{"tah marbuta to heh", "ة", "ه"},
		{"zero-width non-joiner", "\u200cسلام\u200c", "سلام"},
		{"diacritics removed", "\u0645\u064f\u0628\u0627\u0646\u06cc", "مبانی"},
		{"fatha and shadda", "\u0634\u064e\u0651\u0645\u0633", "شمس"},
		{"zero-width space", "a\u200bb", "ab"},
		{"non-breaking space", "a\u00a0b", "a b"},
		{"collapsed whitespace", "  a \t b  ", "a b"},
		{"empty", "", ""},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := Normalize(tt.in); got != tt.want {
				t.Errorf("Normalize(%q) = %q, want %q", tt.in, got, tt.want)
			}
		})
	}
}
