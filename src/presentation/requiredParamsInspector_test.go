package tkPresentation

import (
	"strings"
	"testing"
)

func TestRequiredParamsInspector(t *testing.T) {
	testCases := []struct {
		name            string
		paramsReceived  map[string]any
		paramsRequired  []string
		expectError     bool
		expectedMissing []string
	}{
		{
			name: "AllRequiredParamsPresent",
			paramsReceived: map[string]any{
				"param1": "value1",
				"param2": "value2",
			},
			paramsRequired:  []string{"param1", "param2"},
			expectError:     false,
			expectedMissing: nil,
		},
		{
			name: "SomeRequiredParamsMissing",
			paramsReceived: map[string]any{
				"param1": "value1",
			},
			paramsRequired:  []string{"param1", "param2", "param3"},
			expectError:     true,
			expectedMissing: []string{"param2", "param3"},
		},
		{
			name: "NoRequiredParams",
			paramsReceived: map[string]any{
				"param1": "value1",
			},
			paramsRequired:  []string{},
			expectError:     false,
			expectedMissing: nil,
		},
		{
			name: "ExtraParamsReceived",
			paramsReceived: map[string]any{
				"param1": "value1",
				"param2": "value2",
				"extra":  "value3",
			},
			paramsRequired:  []string{"param1", "param2"},
			expectError:     false,
			expectedMissing: nil,
		},
		{
			name:            "EmptyMaps",
			paramsReceived:  map[string]any{},
			paramsRequired:  []string{},
			expectError:     false,
			expectedMissing: nil,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			err := RequiredParamsInspector(testCase.paramsReceived, testCase.paramsRequired)

			if testCase.expectError {
				if err == nil {
					t.Errorf("RequiredParamsInspectorSucceededWhenItShouldFail")
					return
				}

				expectedErrorMsg := "RequiredParamsMissing: " + strings.Join(testCase.expectedMissing, ", ")
				if err.Error() != expectedErrorMsg {
					t.Errorf("UnexpectedErrorMessage: expected '%s', got '%s'", expectedErrorMsg, err.Error())
				}
				return
			}

			if err != nil {
				t.Errorf("RequiredParamsInspectorFailedWhenItShouldSucceed: %v", err)
				return
			}
		})
	}
}
