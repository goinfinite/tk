package tkInfra

import (
	"regexp"
	"strings"
	"unicode"
)

var escapableCharsRegex = regexp.MustCompile(`[^\w@%+=:,./-]`)

type ShellEscape struct {
}

func (helper ShellEscape) Quote(inputStr string) string {
	if len(inputStr) == 0 {
		return "''"
	}

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
