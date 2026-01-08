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
