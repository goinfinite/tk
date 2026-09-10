package tkInfraDb

import (
	"errors"
	"testing"
)

func TestPaginationPagesTotalResolver(t *testing.T) {
	testCases := []struct {
		name               string
		itemsTotal         uint64
		itemsPerPage       uint16
		expectedPagesTotal uint32
		expectedError      error
	}{
		{
			name:               "ZeroItems",
			itemsTotal:         0,
			itemsPerPage:       3,
			expectedPagesTotal: 0,
		},
		{
			name:               "ExactPage",
			itemsTotal:         9,
			itemsPerPage:       3,
			expectedPagesTotal: 3,
		},
		{
			name:               "PartialPageCountsAsPage",
			itemsTotal:         10,
			itemsPerPage:       3,
			expectedPagesTotal: 4,
		},
		{
			name:               "SingleItemSinglePage",
			itemsTotal:         1,
			itemsPerPage:       1,
			expectedPagesTotal: 1,
		},
		{
			name:          "ZeroItemsPerPage",
			itemsTotal:    10,
			itemsPerPage:  0,
			expectedError: ErrItemsPerPageCannotBeZero,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			pagesTotal, err := PaginationPagesTotalResolver(
				testCase.itemsTotal,
				testCase.itemsPerPage,
			)

			if testCase.expectedError != nil {
				if err == nil {
					t.Fatalf(
						"MissingExpectedError: %v", testCase.expectedError,
					)
				}
				if !errors.Is(err, testCase.expectedError) {
					t.Errorf(
						"UnexpectedError: expected %v, got %v",
						testCase.expectedError, err,
					)
				}
				if err.Error() != ErrItemsPerPageCannotBeZero.Error() {
					t.Errorf(
						"UnexpectedErrorMessage: expected %s, got %v",
						ErrItemsPerPageCannotBeZero.Error(), err,
					)
				}
				return
			}

			if err != nil {
				t.Fatalf("UnexpectedError: %v", err)
			}
			if pagesTotal != testCase.expectedPagesTotal {
				t.Errorf(
					"PagesTotalMismatch: expected %d, got %d",
					testCase.expectedPagesTotal, pagesTotal,
				)
			}
		})
	}
}
