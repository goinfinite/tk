package tkValueObject

import "testing"

func TestNewNetworkPort(t *testing.T) {
	t.Run("StringInput", func(t *testing.T) {
		testCaseStructs := []struct {
			inputValue     any
			expectedOutput NetworkPort
			expectError    bool
		}{
			{"8080", NetworkPort(8080), false},
			{int(443), NetworkPort(443), false},
			{int8(80), NetworkPort(80), false},
			{int16(8000), NetworkPort(8000), false},
			{int32(8080), NetworkPort(8080), false},
			{int64(8443), NetworkPort(8443), false},
			{uint(443), NetworkPort(443), false},
			{uint8(80), NetworkPort(80), false},
			{uint16(8000), NetworkPort(8000), false},
			{uint32(8080), NetworkPort(8080), false},
			{uint64(8443), NetworkPort(8443), false},
			{float32(8080), NetworkPort(8080), false},
			{float64(8443), NetworkPort(8443), false},
			{0, NetworkPort(0), false},
			{65535, NetworkPort(65535), false},
			// Invalid network ports
			{"-1", NetworkPort(0), true},
			{int(-1), NetworkPort(0), true},
			{int8(-1), NetworkPort(0), true},
			{int16(-1), NetworkPort(0), true},
			{int32(-1), NetworkPort(0), true},
			{int64(-1), NetworkPort(0), true},
			{float32(-1), NetworkPort(0), true},
			{float64(-1), NetworkPort(0), true},
			{"abc", NetworkPort(0), true},
			{true, NetworkPort(0), true},
			{[]string{"8080"}, NetworkPort(0), true},
		}

		for _, testCase := range testCaseStructs {
			actualOutput, conversionErr := NewNetworkPort(testCase.inputValue)
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
			inputValue     NetworkPort
			expectedOutput string
		}{
			{NetworkPort(80), "80"},
			{NetworkPort(443), "443"},
			{NetworkPort(8080), "8080"},
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
			inputValue     NetworkPort
			expectedOutput uint16
		}{
			{NetworkPort(80), 80},
			{NetworkPort(443), 443},
			{NetworkPort(8080), 8080},
		}

		for _, testCase := range testCaseStructs {
			actualOutput := testCase.inputValue.Uint16()
			if actualOutput != testCase.expectedOutput {
				t.Errorf("UnexpectedOutputValue: '%v' vs '%v' [%v]", actualOutput, testCase.expectedOutput, testCase.inputValue)
			}
		}
	})
}
