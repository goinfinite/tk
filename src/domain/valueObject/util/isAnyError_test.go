package tkVoUtil

import (
	"errors"
	"fmt"
	"testing"
)

func TestIsAnyError(t *testing.T) {
	errA := errors.New("ErrA")
	errB := errors.New("ErrB")
	wrappedA := fmt.Errorf("context: %w", errA)
	unrelatedErr := errors.New("ErrC")

	testCases := []struct {
		name     string
		err      error
		targets  []error
		expected bool
	}{
		{"ExactMatch", errA, []error{errA}, true},
		{"SecondTargetMatches", errA, []error{errB, errA}, true},
		{"WrappedErrorMatches", wrappedA, []error{errA}, true},
		{"NoTargetMatches", errA, []error{errB, unrelatedErr}, false},
		{"NoTargets", errA, nil, false},
		{"NilError", nil, []error{errA}, false},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			actual := IsAnyError(testCase.err, testCase.targets...)

			if actual != testCase.expected {
				t.Errorf("UnexpectedResult: %t vs %t", actual, testCase.expected)
			}
		})
	}
}
