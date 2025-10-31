package tkValueObject

import "testing"

func TestNewNetworkProtocol(t *testing.T) {
	t.Run("StringInput", func(t *testing.T) {
		testCaseStructs := []struct {
			inputValue     any
			expectedOutput NetworkProtocol
			expectError    bool
		}{
			{"http", NetworkProtocol("http"), false},
			{"https", NetworkProtocol("https"), false},
			{"ws", NetworkProtocol("ws"), false},
			{"wss", NetworkProtocol("wss"), false},
			{"grpc", NetworkProtocol("grpc"), false},
			{"grpcs", NetworkProtocol("grpcs"), false},
			{"tcp", NetworkProtocol("tcp"), false},
			{"udp", NetworkProtocol("udp"), false},
			{"HTTP", NetworkProtocol("http"), false}, // Case insensitive
			{"HTTPS", NetworkProtocol("https"), false},
			// Invalid network protocols
			{"", NetworkProtocol(""), true},
			{"ftp", NetworkProtocol(""), true},
			{"dhcp", NetworkProtocol(""), true},
			{"smtp", NetworkProtocol(""), true},
			{"invalid", NetworkProtocol(""), true},
			{123, NetworkProtocol(""), true},
			{true, NetworkProtocol(""), true},
			{[]string{"http"}, NetworkProtocol(""), true},
		}

		for _, testCase := range testCaseStructs {
			actualOutput, conversionErr := NewNetworkProtocol(testCase.inputValue)
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
			inputValue     NetworkProtocol
			expectedOutput string
		}{
			{NetworkProtocol("http"), "http"},
			{NetworkProtocol("https"), "https"},
			{NetworkProtocol("tcp"), "tcp"},
			{NetworkProtocol("udp"), "udp"},
		}

		for _, testCase := range testCaseStructs {
			actualOutput := testCase.inputValue.String()
			if actualOutput != testCase.expectedOutput {
				t.Errorf("UnexpectedOutputValue: '%v' vs '%v' [%v]", actualOutput, testCase.expectedOutput, testCase.inputValue)
			}
		}
	})
}
