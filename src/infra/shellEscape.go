package tkInfra

import (
	"regexp"
	"strings"
	"unicode"
)

var shellEscapeEscapableCharsRegex = regexp.MustCompile(`[^\w@%+=:,./-]`)

type ShellEscape struct {
}

func (ShellEscape) quoteAlways(inputStr string) string {
	return "'" + strings.ReplaceAll(inputStr, "'", "'\"'\"'") + "'"
}

func (shellEscape ShellEscape) Quote(inputStr string) string {
	if len(inputStr) > 0 && !shellEscapeEscapableCharsRegex.MatchString(inputStr) {
		return inputStr
	}

	return shellEscape.quoteAlways(inputStr)
}

func (shellEscape ShellEscape) Join(arguments []string) string {
	quotedArguments := make([]string, 0, len(arguments))
	for _, argument := range arguments {
		quotedArguments = append(quotedArguments, shellEscape.quoteAlways(argument))
	}

	return strings.Join(quotedArguments, " ")
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
