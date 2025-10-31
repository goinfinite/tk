package tkValueObject

import (
	"testing"
)

func TestNewHttpStatusCode(t *testing.T) {
	t.Run("StringInput", func(t *testing.T) {
		testCaseStructs := []struct {
			inputValue     any
			expectedOutput HttpStatusCode
			expectError    bool
		}{
			{200, HttpStatusCode(200), false},
			{"200", HttpStatusCode(200), false},
			{201, HttpStatusCode(201), false},
			{"207", HttpStatusCode(207), false},
			{401, HttpStatusCode(401), false},
			{"404", HttpStatusCode(404), false},
			{"500", HttpStatusCode(500), false},
			{300, HttpStatusCode(300), false},
			{"100", HttpStatusCode(100), false},
			{599, HttpStatusCode(599), false},
			// Invalid status codes
			{"600", HttpStatusCode(0), true},
			{666, HttpStatusCode(0), true},
			{"99", HttpStatusCode(0), true},
			{"abc", HttpStatusCode(0), true},
			{"", HttpStatusCode(0), true},
			{nil, HttpStatusCode(0), true},
			{700, HttpStatusCode(0), true},
			{true, HttpStatusCode(0), true},
			{"<script>alert('xss')</script>", HttpStatusCode(0), true},
			{"rm -rf /", HttpStatusCode(0), true},
			{"asds12388,jpg", HttpStatusCode(0), true},
			{"@nDr3A5_", HttpStatusCode(0), true},
			{[]string{"200"}, HttpStatusCode(0), true},
		}

		for _, testCase := range testCaseStructs {
			actualOutput, conversionErr := NewHttpStatusCode(testCase.inputValue)
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
			inputValue     HttpStatusCode
			expectedOutput string
		}{
			{HttpStatusCode(200), "200"},
			{HttpStatusCode(404), "404"},
			{HttpStatusCode(500), "500"},
		}

		for _, testCase := range testCaseStructs {
			actualOutput := testCase.inputValue.String()
			if actualOutput != testCase.expectedOutput {
				t.Errorf("UnexpectedOutputValue: '%v' vs '%v' [%v]", actualOutput, testCase.expectedOutput, testCase.inputValue)
			}
		}
	})

	t.Run("Uint16Method", func(t *testing.T) {
		testCaseStructs := []struct {
			inputValue     HttpStatusCode
			expectedOutput uint16
		}{
			{HttpStatusCode(200), 200},
			{HttpStatusCode(404), 404},
			{HttpStatusCode(500), 500},
		}

		for _, testCase := range testCaseStructs {
			actualOutput := testCase.inputValue.Uint16()
			if actualOutput != testCase.expectedOutput {
				t.Errorf("UnexpectedOutputValue: '%v' vs '%v' [%v]", actualOutput, testCase.expectedOutput, testCase.inputValue)
			}
		}
	})
}
