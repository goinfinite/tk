package tkVoUtil

import "unicode/utf8"

func SafeTruncateString(input string, maxBytes int) string {
	if maxBytes <= 0 {
		return ""
	}

	if len(input) <= maxBytes {
		return input
	}

	cutoff := maxBytes
	for cutoff > 0 && !utf8.RuneStart(input[cutoff]) {
		cutoff--
	}

	return input[:cutoff]
}
