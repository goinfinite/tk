package tkValueObject

import "testing"

func TestNewX509BasicConstraints(t *testing.T) {
	t.Run("ValidBasicConstraints", func(t *testing.T) {
		maxPathZero := 0
		maxPathOne := 1
		maxPathFive := 5
		negativeMaxPath := -1

		testCaseStructs := []struct {
			isAuthority   bool
			maxPathLength *int
			expectError   bool
		}{
			{true, nil, false},
			{false, nil, false},
			{true, &maxPathZero, false},
			{true, &maxPathOne, false},
			{true, &maxPathFive, false},
			{false, &maxPathZero, true}, // RFC 5280: MaxPathLength not allowed for non-CA
			{true, &negativeMaxPath, true},
			{false, &negativeMaxPath, true},
		}

		for _, testCase := range testCaseStructs {
			actualOutput, err := NewX509BasicConstraints(
				testCase.isAuthority, testCase.maxPathLength,
			)

			if testCase.expectError && err == nil {
				t.Errorf(
					"MissingExpectedError: [isAuthority=%v, maxPath=%v]",
					testCase.isAuthority, testCase.maxPathLength,
				)
			}

			if !testCase.expectError && err != nil {
				t.Fatalf(
					"UnexpectedError: '%s' [isAuthority=%v, maxPath=%v]",
					err.Error(), testCase.isAuthority, testCase.maxPathLength,
				)
			}

			if !testCase.expectError {
				if actualOutput.IsAuthority != testCase.isAuthority {
					t.Errorf(
						"UnexpectedIsAuthority: '%v' vs '%v'",
						actualOutput.IsAuthority, testCase.isAuthority,
					)
				}

				if testCase.maxPathLength == nil &&
					actualOutput.MaxPathLength != nil {
					t.Errorf(
						"UnexpectedMaxPathLength: expected nil, got %v",
						*actualOutput.MaxPathLength,
					)
				}

				if testCase.maxPathLength != nil &&
					actualOutput.MaxPathLength == nil {
					t.Errorf(
						"UnexpectedMaxPathLength: expected %v, got nil",
						*testCase.maxPathLength,
					)
				}

				if testCase.maxPathLength != nil &&
					actualOutput.MaxPathLength != nil &&
					*actualOutput.MaxPathLength != *testCase.maxPathLength {
					t.Errorf(
						"UnexpectedMaxPathLength: '%v' vs '%v'",
						*actualOutput.MaxPathLength, *testCase.maxPathLength,
					)
				}
			}
		}
	})
}
