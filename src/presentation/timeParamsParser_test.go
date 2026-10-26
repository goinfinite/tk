package tkPresentation

import (
	"slices"
	"testing"

	tkValueObject "github.com/goinfinite/tk/src/domain/valueObject"
)

func TestTimeParamsParser(test *testing.T) {
	test.Run("SuccessWithAllValidFields", func(subtest *testing.T) {
		testCaseStructs := []struct {
			timeParamNames []string
			inputMap       map[string]any
			expectedOutput map[string]*tkValueObject.UnixTime
		}{
			{
				timeParamNames: []string{"createdAt", "updatedAt"},
				inputMap: map[string]any{
					"createdAt": 1609459200,   // 2021-01-01 00:00:00 UTC
					"updatedAt": "1609545600", // 2021-01-02 00:00:00 UTC
				},
				expectedOutput: map[string]*tkValueObject.UnixTime{
					"createdAt": func() *tkValueObject.UnixTime {
						timeParam, _ := tkValueObject.NewUnixTime(1609459200)
						return &timeParam
					}(),
					"updatedAt": func() *tkValueObject.UnixTime {
						timeParam, _ := tkValueObject.NewUnixTime("1609545600")
						return &timeParam
					}(),
				},
			},
		}

		for _, testCase := range testCaseStructs {
			actualOutput := TimeParamsParser(testCase.timeParamNames, testCase.inputMap)

			for _, paramName := range testCase.timeParamNames {
				expectedValue := testCase.expectedOutput[paramName]
				actualValue := actualOutput[paramName]

				if (expectedValue == nil && actualValue != nil) ||
					(expectedValue != nil && actualValue == nil) ||
					(expectedValue != nil && actualValue != nil && *expectedValue != *actualValue) {
					subtest.Errorf("Unexpected%s: expected '%v', got '%v'", paramName, expectedValue, actualValue)
				}
			}
		}
	})

	test.Run("SuccessWithPartialFields", func(subtest *testing.T) {
		testCaseStructs := []struct {
			timeParamNames []string
			inputMap       map[string]any
			expectedOutput map[string]*tkValueObject.UnixTime
		}{
			{
				timeParamNames: []string{"createdAt", "updatedAt"},
				inputMap: map[string]any{
					"createdAt": 1609459200,
				},
				expectedOutput: map[string]*tkValueObject.UnixTime{
					"createdAt": func() *tkValueObject.UnixTime {
						timeParam, _ := tkValueObject.NewUnixTime(1609459200)
						return &timeParam
					}(),
					"updatedAt": nil,
				},
			},
			{
				timeParamNames: []string{"createdAt", "updatedAt"},
				inputMap: map[string]any{
					"updatedAt": "1609545600",
				},
				expectedOutput: map[string]*tkValueObject.UnixTime{
					"createdAt": nil,
					"updatedAt": func() *tkValueObject.UnixTime {
						timeParam, _ := tkValueObject.NewUnixTime("1609545600")
						return &timeParam
					}(),
				},
			},
		}

		for _, testCase := range testCaseStructs {
			actualOutput := TimeParamsParser(testCase.timeParamNames, testCase.inputMap)

			for _, paramName := range testCase.timeParamNames {
				expectedValue := testCase.expectedOutput[paramName]
				actualValue := actualOutput[paramName]

				if (expectedValue == nil && actualValue != nil) ||
					(expectedValue != nil && actualValue == nil) ||
					(expectedValue != nil && actualValue != nil && *expectedValue != *actualValue) {
					subtest.Errorf("Unexpected%s: expected '%v', got '%v'", paramName, expectedValue, actualValue)
				}
			}
		}
	})

	test.Run("SuccessWithNilInput", func(subtest *testing.T) {
		timeParamNames := []string{"createdAt", "updatedAt"}

		actualOutput := TimeParamsParser(timeParamNames, nil)

		for _, paramName := range timeParamNames {
			if actualOutput[paramName] != nil {
				subtest.Errorf("Unexpected%s: ShouldBeNil, got '%v'", paramName, actualOutput[paramName])
			}
		}
	})

	test.Run("SuccessWithEmptyStringInput", func(subtest *testing.T) {
		timeParamNames := []string{"createdAt", "updatedAt"}
		inputMap := map[string]any{
			"createdAt": "",
			"updatedAt": "1609545600",
		}

		actualOutput := TimeParamsParser(timeParamNames, inputMap)

		if actualOutput["createdAt"] != nil {
			subtest.Errorf("UnexpectedCreatedAt: ShouldBeNil, got '%v'", actualOutput["createdAt"])
		}
		if actualOutput["updatedAt"] == nil {
			subtest.Errorf("UnexpectedUpdatedAt: ShouldNotBeNil")
		}
	})

	test.Run("InvalidTimeParamSetsToNil", func(subtest *testing.T) {
		testCaseStructs := []struct {
			timeParamNames []string
			inputMap       map[string]any
			invalidParams  []string
		}{
			{
				timeParamNames: []string{"createdAt"},
				inputMap: map[string]any{
					"createdAt": "invalid",
				},
				invalidParams: []string{"createdAt"},
			},
			{
				timeParamNames: []string{"createdAt", "updatedAt"},
				inputMap: map[string]any{
					"createdAt": 1609459200,
					"updatedAt": []int{1, 2, 3},
				},
				invalidParams: []string{"updatedAt"},
			},
			{
				timeParamNames: []string{"createdAt", "updatedAt"},
				inputMap: map[string]any{
					"createdAt": nil,
					"updatedAt": "1609545600",
				},
				invalidParams: []string{"createdAt"},
			},
		}

		for _, testCase := range testCaseStructs {
			actualOutput := TimeParamsParser(testCase.timeParamNames, testCase.inputMap)

			for _, paramName := range testCase.timeParamNames {
				actualValue := actualOutput[paramName]

				isInvalidParam := slices.Contains(testCase.invalidParams, paramName)
				if isInvalidParam {
					if actualValue != nil {
						subtest.Errorf("Unexpected%s: ShouldBeNil, got '%v'", paramName, actualValue)
					}
					continue
				}

				if actualValue == nil {
					subtest.Errorf("Unexpected%s: ShouldNotBeNil", paramName)
				}
			}
		}
	})
}
