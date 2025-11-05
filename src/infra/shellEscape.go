package tkInfra

import (
	"regexp"
	"strings"
	"unicode"
)

var shellEscapeEscapableCharsRegex = regexp.MustCompile(`[^\w@%+=:,./-]`)

type ShellEscape struct {
}

func (ShellEscape) Quote(inputStr string) string {
	if len(inputStr) == 0 {
		return "''"
	}

	if !shellEscapeEscapableCharsRegex.MatchString(inputStr) {
		return inputStr
	}

	return "'" + strings.ReplaceAll(inputStr, "'", "'\"'\"'") + "'"
}

func (ShellEscape) StripUnsafe(inputStr string) string {
	utf8ValidStr := strings.ToValidUTF8(inputStr, "")

	return strings.Map(func(char rune) rune {
		if unicode.IsPrint(char) {
			return char
		}

		return -1
	}, utf8ValidStr)
}
