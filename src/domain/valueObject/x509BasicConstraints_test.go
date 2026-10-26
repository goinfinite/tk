package tkValueObject

import "testing"

func TestNewX509BasicConstraints(test *testing.T) {
	test.Run("ValidBasicConstraints", func(subtest *testing.T) {
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
				subtest.Errorf(
					"MissingExpectedError: [isAuthority=%v, maxPath=%v]",
					testCase.isAuthority, testCase.maxPathLength,
				)
			}

			if !testCase.expectError && err != nil {
				subtest.Fatalf(
					"UnexpectedError: '%s' [isAuthority=%v, maxPath=%v]",
					err.Error(), testCase.isAuthority, testCase.maxPathLength,
				)
			}

			if !testCase.expectError {
				if actualOutput.IsAuthority != testCase.isAuthority {
					subtest.Errorf(
						"UnexpectedIsAuthority: '%v' vs '%v'",
						actualOutput.IsAuthority, testCase.isAuthority,
					)
				}

				if testCase.maxPathLength == nil &&
					actualOutput.MaxPathLength != nil {
					subtest.Errorf(
						"UnexpectedMaxPathLength: expected nil, got %v",
						*actualOutput.MaxPathLength,
					)
				}

				if testCase.maxPathLength != nil &&
					actualOutput.MaxPathLength == nil {
					subtest.Errorf(
						"UnexpectedMaxPathLength: expected %v, got nil",
						*testCase.maxPathLength,
					)
				}

				if testCase.maxPathLength != nil &&
					actualOutput.MaxPathLength != nil &&
					*actualOutput.MaxPathLength != *testCase.maxPathLength {
					subtest.Errorf(
						"UnexpectedMaxPathLength: '%v' vs '%v'",
						*actualOutput.MaxPathLength, *testCase.maxPathLength,
					)
				}
			}
		}
	})

	test.Run("StringMethod", func(subtest *testing.T) {
		maxPathZero := 0
		maxPathFive := 5

		testCaseStructs := []struct {
			inputValue     X509BasicConstraints
			expectedOutput string
		}{
			{X509BasicConstraints{IsAuthority: true}, "CA:true"},
			{X509BasicConstraints{IsAuthority: false}, "CA:false"},
			{
				X509BasicConstraints{IsAuthority: true, MaxPathLength: &maxPathZero},
				"CA:true, pathlen:0",
			},
			{
				X509BasicConstraints{IsAuthority: true, MaxPathLength: &maxPathFive},
				"CA:true, pathlen:5",
			},
		}

		for _, testCase := range testCaseStructs {
			actualOutput := testCase.inputValue.String()
			if actualOutput != testCase.expectedOutput {
				subtest.Errorf(
					"UnexpectedOutputValue: '%v' vs '%v'",
					actualOutput, testCase.expectedOutput,
				)
			}
		}
	})
}
