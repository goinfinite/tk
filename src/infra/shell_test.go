package tkInfra

import (
	"testing"
)

func TestShell(t *testing.T) {
	t.Run("BasicShell", func(t *testing.T) {
		testCaseStructs := []struct {
			command        string
			args           []string
			expectedOutput string
			expectError    bool
		}{
			{"echo", []string{"hello"}, "hello", false},
			{"echo", []string{"-n", "test"}, "test", false},
			{"true", []string{}, "", false},
			{"false", []string{}, "", true},
		}

		for _, testCase := range testCaseStructs {
			shell := NewShell(
				ShellSettings{Command: testCase.command, Args: testCase.args},
			)
			shellOutput, err := shell.Run()
			if testCase.expectError && err == nil {
				t.Errorf("MissingExpectedError: [%s %v]", testCase.command, testCase.args)
			}
			if !testCase.expectError && err != nil {
				t.Errorf("UnexpectedError: '%s' [%s %v]", err.Error(), testCase.command, testCase.args)
			}
			if !testCase.expectError && shellOutput != testCase.expectedOutput {
				t.Errorf(
					"UnexpectedOutputValue: '%s' vs '%s' [%s %v]",
					shellOutput, testCase.expectedOutput, testCase.command, testCase.args,
				)
			}
		}
	})

	t.Run("ErrorHandling", func(t *testing.T) {
		testCaseStructs := []struct {
			command     string
			args        []string
			expectError bool
		}{
			{"nonexistentcommand", []string{}, true},
			{"ls", []string{"/nonexistent/path"}, true},
			{"cat", []string{"/nonexistent/file"}, true},
		}

		for _, testCase := range testCaseStructs {
			shell := NewShell(
				ShellSettings{Command: testCase.command, Args: testCase.args},
			)

			_, err := shell.Run()
			if testCase.expectError && err == nil {
				t.Errorf("MissingExpectedError: [%s %v]", testCase.command, testCase.args)
			}

			if err != nil {
				if shellErr, assertOk := err.(*ShellError); assertOk {
					if shellErr.ExitCode == 0 {
						t.Errorf("UnexpectedZeroExitCode: [%s %v]", testCase.command, testCase.args)
					}
				}
			}
		}
	})
}
