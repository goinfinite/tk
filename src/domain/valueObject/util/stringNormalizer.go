package tkVoUtil

import (
	"strings"
	"unicode"

	"golang.org/x/text/runes"
	"golang.org/x/text/transform"
	"golang.org/x/text/unicode/norm"
)

func StripAccents(input string) (string, error) {
	// Create a transformation pipeline that:
	// 1. NFD: Decomposes characters into base + combining marks (é → e + ´)
	// 2. Removes unicode.Mn category (nonspacing marks like accents)
	// 3. NFC: Recomposes characters into canonical form
	accentStripTransformer := transform.Chain(
		norm.NFD, runes.Remove(runes.In(unicode.Mn)), norm.NFC,
	)

	stringWithoutAccents, _, err := transform.String(accentStripTransformer, input)
	if err != nil {
		return input, err
	}

	return strings.TrimSpace(stringWithoutAccents), nil
}
