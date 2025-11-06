package tkPresentation

import (
	"slices"
	"testing"

	tkValueObject "github.com/goinfinite/tk/src/domain/valueObject"
)

func TestTimeParamsParser(t *testing.T) {
	t.Run("SuccessWithAllValidFields", func(t *testing.T) {
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
					t.Errorf("Unexpected%s: expected '%v', got '%v'", paramName, expectedValue, actualValue)
				}
			}
		}
	})

	t.Run("SuccessWithPartialFields", func(t *testing.T) {
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
					t.Errorf("Unexpected%s: expected '%v', got '%v'", paramName, expectedValue, actualValue)
				}
			}
		}
	})

	t.Run("SuccessWithNilInput", func(t *testing.T) {
		timeParamNames := []string{"createdAt", "updatedAt"}

		actualOutput := TimeParamsParser(timeParamNames, nil)

		for _, paramName := range timeParamNames {
			if actualOutput[paramName] != nil {
				t.Errorf("Unexpected%s: ShouldBeNil, got '%v'", paramName, actualOutput[paramName])
			}
		}
	})

	t.Run("SuccessWithEmptyStringInput", func(t *testing.T) {
		timeParamNames := []string{"createdAt", "updatedAt"}
		inputMap := map[string]any{
			"createdAt": "",
			"updatedAt": "1609545600",
		}

		actualOutput := TimeParamsParser(timeParamNames, inputMap)

		if actualOutput["createdAt"] != nil {
			t.Errorf("UnexpectedCreatedAt: ShouldBeNil, got '%v'", actualOutput["createdAt"])
		}
		if actualOutput["updatedAt"] == nil {
			t.Errorf("UnexpectedUpdatedAt: ShouldNotBeNil")
		}
	})

	t.Run("InvalidTimeParamSetsToNil", func(t *testing.T) {
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
						t.Errorf("Unexpected%s: ShouldBeNil, got '%v'", paramName, actualValue)
					}
					return
				}

				if actualValue == nil {
					t.Errorf("Unexpected%s: ShouldNotBeNil", paramName)
				}
			}
		}
	})
}
