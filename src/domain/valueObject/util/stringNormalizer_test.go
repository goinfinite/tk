package tkVoUtil

import "testing"

func TestStripAccents(t *testing.T) {
	testCases := []struct {
		input          string
		expectedOutput string
	}{
		{"São Paulo", "Sao Paulo"},
		{"Montréal", "Montreal"},
		{"München", "Munchen"},
		{"Zürich", "Zurich"},
		{"Bogotá", "Bogota"},
		{"Florianópolis", "Florianopolis"},
		{"Região", "Regiao"},
		{"Ação", "Acao"},
		{"Curaçao", "Curacao"},
		{"Café", "Cafe"},
		{"naïve", "naive"},
		{"résumé", "resume"},
		{"NoAccents", "NoAccents"},
		{"123", "123"},
		{"Test-City", "Test-City"},
	}

	for _, testCase := range testCases {
		actualOutput, err := StripAccents(testCase.input)
		if err != nil {
			t.Fatalf(
				"UnexpectedError: '%s' for input '%s'",
				err.Error(), testCase.input,
			)
		}

		if actualOutput != testCase.expectedOutput {
			t.Errorf(
				"UnexpectedOutput: '%s' vs '%s' for input '%s'",
				actualOutput, testCase.expectedOutput, testCase.input,
			)
		}
	}
}

func TestStripHexSeparators(t *testing.T) {
	testCases := []struct {
		input          string
		expectedOutput string
	}{
		{"AA:BB:CC:DD", "AABBCCDD"},
		{"AA BB CC DD", "AABBCCDD"},
		{"AA:BB CC:DD", "AABBCCDD"},
		{"AABBCCDD", "AABBCCDD"},
		{"aa:bb:cc:dd:ee:ff", "aabbccddeeff"},
		{"12:34:56:78:90:AB:CD:EF", "1234567890ABCDEF"},
		{"12 34 56 78 90 AB CD EF", "1234567890ABCDEF"},
		{"12:34 56:78 90:AB CD:EF", "1234567890ABCDEF"},
		{"", ""},
		{"A", "A"},
		{"1:2:3:4:5:6:7:8:9:0", "1234567890"},
		{" : ", ""},
		{":::", ""},
		{"   ", ""},
		{"A:B C:D E:F", "ABCDEF"},
	}

	for _, testCase := range testCases {
		actualOutput := StripHexSeparators(testCase.input)

		if actualOutput != testCase.expectedOutput {
			t.Errorf(
				"UnexpectedOutput: '%s' vs '%s' for input '%s'",
				actualOutput, testCase.expectedOutput, testCase.input,
			)
		}
	}
}
