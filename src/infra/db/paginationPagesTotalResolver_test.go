package tkInfraDb

import (
	"errors"
	"testing"
)

func TestPaginationPagesTotalResolver(test *testing.T) {
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
		{
			name:          "PagesTotalOverflow",
			itemsTotal:    1 << 32,
			itemsPerPage:  1,
			expectedError: ErrPagesTotalOverflow,
		},
		{
			name:          "PagesTotalOverflowAtMaxItemsTotal",
			itemsTotal:    ^uint64(0),
			itemsPerPage:  2,
			expectedError: ErrPagesTotalOverflow,
		},
	}

	for _, testCase := range testCases {
		test.Run(testCase.name, func(subtest *testing.T) {
			pagesTotal, err := PaginationPagesTotalResolver(
				testCase.itemsTotal,
				testCase.itemsPerPage,
			)

			if testCase.expectedError != nil {
				if err == nil {
					subtest.Fatalf(
						"MissingExpectedError: %v", testCase.expectedError,
					)
				}
				if !errors.Is(err, testCase.expectedError) {
					subtest.Errorf(
						"UnexpectedError: expected %v, got %v",
						testCase.expectedError, err,
					)
				}
				if err.Error() != testCase.expectedError.Error() {
					subtest.Errorf(
						"UnexpectedErrorMessage: expected %s, got %v",
						testCase.expectedError.Error(), err,
					)
				}
				return
			}

			if err != nil {
				subtest.Fatalf("UnexpectedError: %v", err)
			}
			if pagesTotal != testCase.expectedPagesTotal {
				subtest.Errorf(
					"PagesTotalMismatch: expected %d, got %d",
					testCase.expectedPagesTotal, pagesTotal,
				)
			}
		})
	}
}
