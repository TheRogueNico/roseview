// Package normalize unifies Persian/Arabic character variants and whitespace.
package normalize

import "strings"

// charVariants maps Arabic presentation-form characters to their Persian
// equivalents, and collapses zero-width/non-breaking spaces to plain spaces.
var charVariants = strings.NewReplacer(
	// Arabic/Persian letter variants
	"ك", "ک",
	"ي", "ی",
	"ى", "ی",
	"ۀ", "هٔ",
	"ة", "ه",

	// Arabic diacritical marks (harakat)
	"\u064b", "", // fathatan ً
	"\u064c", "", // dammatan ٌ
	"\u064d", "", // kasratan ٍ
	"\u064e", "", // fatha َ
	"\u064f", "", // damma ُ
	"\u0650", "", // kasra ِ
	"\u0651", "", // shadda ّ
	"\u0652", "", // sukun ْ
	"\u0670", "", // superscript alef ٰ

	// Spacing characters
	"\u200b", "", // zero-width space
	"\u200d", "", // zero-width joiner
	"\ufeff", "", // zero-width no-break space / BOM
	"\u200c", " ", // zero-width non-joiner
	"\u00a0", " ", // non-breaking space
)

// Normalize unifies character variants and collapses whitespace, so header
// and cell text can be compared and displayed consistently regardless of
// formatting quirks in the source HTML.
func Normalize(s string) string {
	s = charVariants.Replace(s)
	return strings.Join(strings.Fields(s), " ")
}
