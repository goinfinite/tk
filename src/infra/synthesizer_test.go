package tkInfra

import (
	"regexp"
	"strings"
	"testing"

	tkValueObject "github.com/goinfinite/tk/src/domain/valueObject"
)

func TestCharsetPresenceGuarantor(t *testing.T) {
	synth := &Synthesizer{}

	t.Run("CharsetAlreadyPresent", func(t *testing.T) {
		testCaseStructs := []struct {
			inputString    []byte
			charset        string
			expectedOutput []byte
		}{
			{[]byte("abc123"), "123", []byte("abc123")},
			{[]byte("ABCDEF"), "ABCDEF", []byte("ABCDEF")},
			{[]byte("abc!@#"), "!@#", []byte("abc!@#")},
			{[]byte("a"), "a", []byte("a")},
		}

		for _, testCase := range testCaseStructs {
			actualOutput := synth.CharsetPresenceGuarantor(testCase.inputString, testCase.charset)
			if string(actualOutput) != string(testCase.expectedOutput) {
				t.Errorf(
					"UnexpectedOutputValue: '%s' vs '%s' [%s, %s]",
					string(actualOutput), string(testCase.expectedOutput), string(testCase.inputString),
					testCase.charset,
				)
			}
		}
	})

	t.Run("CharsetNotPresent", func(t *testing.T) {
		testCaseStructs := []struct {
			inputString []byte
			charset     string
		}{
			{[]byte("abcdef"), "123"},
			{[]byte("123456"), "abc"},
			{[]byte("abcABC"), "!@#"},
		}

		for _, testCase := range testCaseStructs {
			inputCopy := make([]byte, len(testCase.inputString))
			copy(inputCopy, testCase.inputString)

			actualOutput := synth.CharsetPresenceGuarantor(inputCopy, testCase.charset)

			if string(actualOutput) == string(testCase.inputString) {
				t.Errorf(
					"OutputUnchanged: Expected modification to include charset '%s' in '%s'",
					testCase.charset, string(testCase.inputString),
				)
			}

			if !strings.ContainsAny(string(actualOutput), testCase.charset) {
				t.Errorf(
					"CharsetNotAdded: Output '%s' does not contain any character from charset '%s'",
					string(actualOutput), testCase.charset,
				)
			}

			differentChars := 0
			for i := 0; i < len(testCase.inputString); i++ {
				if actualOutput[i] != testCase.inputString[i] {
					differentChars++
				}
			}
			if differentChars != 1 {
				t.Errorf(
					"UnexpectedModificationCount: %d characters modified instead of 1 [%s, %s]",
					differentChars, string(testCase.inputString), testCase.charset,
				)
			}
		}
	})

	t.Run("EdgeCases", func(t *testing.T) {
		singleChar := []byte("a")
		charset := "b"
		actualOutput := synth.CharsetPresenceGuarantor(singleChar, charset)
		if string(actualOutput) != "b" {
			t.Errorf(
				"UnexpectedOutputValue: '%s' vs 'b' [single character case]",
				string(actualOutput),
			)
		}

		twoChars := []byte("ab")
		charset = "c"
		actualOutput = synth.CharsetPresenceGuarantor(twoChars, charset)
		if !strings.ContainsAny(string(actualOutput), charset) {
			t.Errorf(
				"CharsetNotAdded: Output '%s' does not contain any character from charset '%s'",
				string(actualOutput), charset,
			)
		}
	})
}

func TestPasswordFactory(t *testing.T) {
	synth := &Synthesizer{}

	t.Run("PasswordLength", func(t *testing.T) {
		testCaseStructs := []struct {
			desiredLength        int
			shouldIncludeSymbols bool
		}{
			{8, false},
			{12, false},
			{16, true},
			{20, true},
			{4, false}, // Edge case: small length
			{3, true},  // Edge case: very small length
			{0, false}, // Edge case: zero length
		}

		for _, testCase := range testCaseStructs {
			password := synth.PasswordFactory(
				testCase.desiredLength, testCase.shouldIncludeSymbols,
			)
			if len(password) != testCase.desiredLength {
				t.Errorf(
					"UnexpectedPasswordLength: '%d' vs '%d' [desired: %d, includeSymbols: %t]",
					len(password), testCase.desiredLength, testCase.desiredLength,
					testCase.shouldIncludeSymbols,
				)
			}
		}
	})

	t.Run("PasswordCharacteristics", func(t *testing.T) {
		password := synth.PasswordFactory(12, false)
		if !strings.ContainsAny(password, CharsetLowercaseLetters) {
			t.Errorf("MissingLowercaseLetters: '%s'", password)
		}

		if !strings.ContainsAny(password, CharsetUppercaseLetters) {
			t.Errorf("MissingUppercaseLetters: '%s'", password)
		}

		if !strings.ContainsAny(password, CharsetNumbers) {
			t.Errorf("MissingNumbers: '%s'", password)
		}

		if len(password) != 12 {
			t.Errorf("UnexpectedPasswordLength: '%d' vs '12'", len(password))
		}

		passwordWithSymbols := synth.PasswordFactory(12, true)
		if !strings.ContainsAny(passwordWithSymbols, CharsetSymbols) {
			t.Errorf("MissingSymbols: '%s'", passwordWithSymbols)
		}

		if len(passwordWithSymbols) != 12 {
			t.Errorf("UnexpectedPasswordLength: '%d' vs '12'", len(passwordWithSymbols))
		}
	})

	t.Run("ShortPasswordCharacteristics", func(t *testing.T) {
		shortPassword := synth.PasswordFactory(4, false)
		if len(shortPassword) != 4 {
			t.Errorf("UnexpectedPasswordLength: '%d' vs '4'", len(shortPassword))
		}

		veryShortPassword := synth.PasswordFactory(3, true)
		if len(veryShortPassword) != 3 {
			t.Errorf("UnexpectedPasswordLength: '%d' vs '3'", len(veryShortPassword))
		}
	})
}

func TestUsernameFactory(t *testing.T) {
	synth := &Synthesizer{}

	t.Run("UsernameGeneration", func(t *testing.T) {
		usernamesRegex := `^\w{1,256}}$`
		re := regexp.MustCompile(usernamesRegex)

		for range 5 {
			username := synth.UsernameFactory()
			if !re.MatchString(username) {
				t.Errorf(
					"InvalidUsernameFormat: '%s' does not match regex '%s'",
					username, usernamesRegex,
				)
			}
		}
	})
}

func TestMailAddressFactory(t *testing.T) {
	synth := &Synthesizer{}

	t.Run("WithNilUsername", func(t *testing.T) {
		rawMailAddress := synth.MailAddressFactory(nil)
		_, err := tkValueObject.NewMailAddress(rawMailAddress)
		if err != nil {
			t.Errorf("InvalidMailAddress: '%s' is not a valid email address", rawMailAddress)
		}
	})

	t.Run("WithProvidedUsername", func(t *testing.T) {
		testCaseStructs := []struct {
			username string
		}{
			{"testuser"},
			{"admin"},
			{"user123"},
		}

		for _, testCase := range testCaseStructs {
			username := testCase.username
			rawMailAddress := synth.MailAddressFactory(&username)
			mailAddress, err := tkValueObject.NewMailAddress(rawMailAddress)
			if err != nil {
				t.Errorf("InvalidMailAddress: '%s' is not a valid email address", rawMailAddress)
			}

			if !strings.HasPrefix(mailAddress.String(), username) {
				t.Errorf("MissingUsername: '%s' does not start with '%s'", mailAddress, username)
			}
		}
	})
}
