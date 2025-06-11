package tkInfra

import (
	"regexp"
	"strings"
	"unicode"
)

type ShellEscape struct {
}

func (helper ShellEscape) Quote(inputStr string) string {
	if len(inputStr) == 0 {
		return "''"
	}

	escapableCharsRegex := regexp.MustCompile(`[^\w@%+=:,./-]`)
	if !escapableCharsRegex.MatchString(inputStr) {
		return inputStr
	}

	return "'" + strings.ReplaceAll(inputStr, "'", "'\"'\"'") + "'"
}

func (helper ShellEscape) StripUnsafe(inputStr string) string {
	return strings.Map(func(char rune) rune {
		if unicode.IsPrint(char) {
			return char
		}

		return -1
	}, inputStr)
}
