package tkValueObject

import (
	"strings"
	"testing"
)

func TestNewGenericNotes(t *testing.T) {
	t.Run("StringInput", func(t *testing.T) {
		testCaseStructs := []struct {
			inputValue     any
			expectedOutput GenericNotes
			expectError    bool
		}{
			// Valid notes: quotations from famous books
			{"All that is gold does not glitter, Not all those who wander are lost.", GenericNotes("All that is gold does not glitter, Not all those who wander are lost."), false},                                                                                                 // The Lord of the Rings
			{"It is a truth universally acknowledged, that a single man in possession of a good fortune, must be in want of a wife.", GenericNotes("It is a truth universally acknowledged, that a single man in possession of a good fortune, must be in want of a wife."), false}, // Pride and Prejudice
			{"In the beginning was the Word, and the Word was with God, and the Word was God.", GenericNotes("In the beginning was the Word, and the Word was with God, and the Word was God."), false},                                                                             // The Bible
			{"Call me Ishmael.", GenericNotes("Call me Ishmael."), false}, // Moby-Dick
			{"Happy families are all alike; every unhappy family is unhappy in its own way.", GenericNotes("Happy families are all alike; every unhappy family is unhappy in its own way."), false}, // Anna Karenina
			{123, GenericNotes("123"), false},
			{true, GenericNotes("true"), false},
			// Patterns unable to be blocked or regex would be too restrictive
			{"../../../etc/passwd", GenericNotes("../../../etc/passwd"), false},
			{"'; DROP TABLE users; --", GenericNotes("'; DROP TABLE users; --"), false},
			{"; rm -rf /", GenericNotes("; rm -rf /"), false},
			{"javascript:alert('JS injection')", GenericNotes("javascript:alert('JS injection')"), false},
			// Patterns for sure to be blocked
			{"<script>alert('XSS')</script>", GenericNotes("<script>alert('XSS')</script>"), true},
			{"<img src=x onerror=alert(1)>", GenericNotes("<img src=x onerror=alert(1)>"), true},
			{"<?php echo 'PHP injection'; ?>", GenericNotes("<?php echo 'PHP injection'; ?>"), true},
			{"{{7*7}}", GenericNotes("{{7*7}}"), true}, // Template injection
			{"<iframe src='javascript:alert(1)'></iframe>", GenericNotes("<iframe src='javascript:alert(1)'></iframe>"), true},
			{"test\x00", GenericNotes(""), true}, // null byte
			// Invalid inputs
			{"test 😀", GenericNotes(""), true},                  // emoji (not in allowed chars)
			{"", GenericNotes(""), true},                        // Too small
			{strings.Repeat("a", 5001), GenericNotes(""), true}, // Too big
			{[]string{"note"}, GenericNotes("note"), true},
		}

		for _, testCase := range testCaseStructs {
			actualOutput, conversionErr := NewGenericNotes(testCase.inputValue)
			if testCase.expectError && conversionErr == nil {
				t.Errorf("MissingExpectedError: [%v]", testCase.inputValue)
			}
			if !testCase.expectError && conversionErr != nil {
				t.Errorf("UnexpectedError: '%s' [%v]", conversionErr.Error(), testCase.inputValue)
			}
			if !testCase.expectError && actualOutput != testCase.expectedOutput {
				t.Errorf("UnexpectedOutputValue: '%v' vs '%v' [%v]", actualOutput, testCase.expectedOutput, testCase.inputValue)
			}
		}
	})

	t.Run("StringMethod", func(t *testing.T) {
		testCaseStructs := []struct {
			inputValue     GenericNotes
			expectedOutput string
		}{
			{GenericNotes("All that is gold does not glitter, Not all those who wander are lost."), "All that is gold does not glitter, Not all those who wander are lost."},
			{GenericNotes("It is a truth universally acknowledged, that a single man in possession of a good fortune, must be in want of a wife."), "It is a truth universally acknowledged, that a single man in possession of a good fortune, must be in want of a wife."},
		}

		for _, testCase := range testCaseStructs {
			actualOutput := testCase.inputValue.String()
			if actualOutput != testCase.expectedOutput {
				t.Errorf("UnexpectedOutputValue: '%v' vs '%v' [%v]", actualOutput, testCase.expectedOutput, testCase.inputValue)
			}
		}
	})
}
