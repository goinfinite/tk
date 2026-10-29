package tkPresentation

import (
	"testing"

	tkDto "github.com/goinfinite/tk/src/domain/dto"
	tkValueObject "github.com/goinfinite/tk/src/domain/valueObject"
)

func TestPaginationDtoParser(test *testing.T) {
	test.Run("SuccessWithAllValidFields", func(subtest *testing.T) {
		testCaseStructs := []struct {
			inputMap       map[string]any
			expectedOutput tkDto.Pagination
			expectError    bool
		}{
			{
				inputMap: map[string]any{
					"pageNumber":    uint32(5),
					"itemsPerPage":  uint16(20),
					"sortBy":        "name",
					"sortDirection": "asc",
					"lastSeenId":    "abc123",
				},
				expectedOutput: tkDto.Pagination{
					PageNumber:   5,
					ItemsPerPage: 20,
					SortBy: func() *tkValueObject.PaginationSortBy {
						sortBy, _ := tkValueObject.NewPaginationSortBy("name")
						return &sortBy
					}(),
					SortDirection: func() *tkValueObject.PaginationSortDirection {
						sortDir, _ := tkValueObject.NewPaginationSortDirection("asc")
						return &sortDir
					}(),
					LastSeenId: func() *tkValueObject.PaginationLastSeenId {
						lastSeenId, _ := tkValueObject.NewPaginationLastSeenId("abc123")
						return &lastSeenId
					}(),
				},
				expectError: false,
			},
			{
				inputMap: map[string]any{
					"pageNumber":   5.0,
					"itemsPerPage": 20.0,
				},
				expectedOutput: tkDto.Pagination{
					PageNumber:   5,
					ItemsPerPage: 20,
				},
				expectError: false,
			},
		}

		defaultPagination := tkDto.Pagination{
			PageNumber:   1,
			ItemsPerPage: 10,
		}

		for _, testCase := range testCaseStructs {
			actualOutput, conversionErr := PaginationDtoParser(defaultPagination, testCase.inputMap)
			if testCase.expectError && conversionErr == nil {
				subtest.Errorf("MissingExpectedError: [%v]", testCase.inputMap)
			}
			if !testCase.expectError && conversionErr != nil {
				subtest.Errorf("UnexpectedError: '%s' [%v]", conversionErr.Error(), testCase.inputMap)
			}
			if !testCase.expectError {
				if actualOutput.PageNumber != testCase.expectedOutput.PageNumber {
					subtest.Errorf("UnexpectedPageNumber: '%v' vs '%v'", actualOutput.PageNumber, testCase.expectedOutput.PageNumber)
				}
				if actualOutput.ItemsPerPage != testCase.expectedOutput.ItemsPerPage {
					subtest.Errorf("UnexpectedItemsPerPage: '%v' vs '%v'", actualOutput.ItemsPerPage, testCase.expectedOutput.ItemsPerPage)
				}
				if (actualOutput.SortBy == nil && testCase.expectedOutput.SortBy != nil) ||
					(actualOutput.SortBy != nil && testCase.expectedOutput.SortBy == nil) ||
					(actualOutput.SortBy != nil && *actualOutput.SortBy != *testCase.expectedOutput.SortBy) {
					subtest.Errorf("UnexpectedSortBy: '%v' vs '%v'", actualOutput.SortBy, testCase.expectedOutput.SortBy)
				}
				if (actualOutput.SortDirection == nil && testCase.expectedOutput.SortDirection != nil) ||
					(actualOutput.SortDirection != nil && testCase.expectedOutput.SortDirection == nil) ||
					(actualOutput.SortDirection != nil && *actualOutput.SortDirection != *testCase.expectedOutput.SortDirection) {
					subtest.Errorf("UnexpectedSortDirection: '%v' vs '%v'", actualOutput.SortDirection, testCase.expectedOutput.SortDirection)
				}
				if (actualOutput.LastSeenId == nil && testCase.expectedOutput.LastSeenId != nil) ||
					(actualOutput.LastSeenId != nil && testCase.expectedOutput.LastSeenId == nil) ||
					(actualOutput.LastSeenId != nil && *actualOutput.LastSeenId != *testCase.expectedOutput.LastSeenId) {
					subtest.Errorf("UnexpectedLastSeenId: '%v' vs '%v'", actualOutput.LastSeenId, testCase.expectedOutput.LastSeenId)
				}
			}
		}
	})

	test.Run("SuccessWithPartialFields", func(subtest *testing.T) {
		testCaseStructs := []struct {
			inputMap       map[string]any
			expectedOutput tkDto.Pagination
			expectError    bool
		}{
			{
				inputMap: map[string]any{
					"pageNumber": uint32(3),
				},
				expectedOutput: tkDto.Pagination{
					PageNumber:   3,
					ItemsPerPage: 10,
				},
				expectError: false,
			},
			{
				inputMap: map[string]any{
					"itemsPerPage": uint16(50),
				},
				expectedOutput: tkDto.Pagination{
					PageNumber:   1,
					ItemsPerPage: 50,
				},
				expectError: false,
			},
			{
				inputMap: map[string]any{
					"sortBy": "created_at",
				},
				expectedOutput: tkDto.Pagination{
					PageNumber:   1,
					ItemsPerPage: 10,
					SortBy: func() *tkValueObject.PaginationSortBy {
						sortBy, _ := tkValueObject.NewPaginationSortBy("created_at")
						return &sortBy
					}(),
				},
				expectError: false,
			},
			{
				inputMap: map[string]any{
					"sortDirection": "desc",
				},
				expectedOutput: tkDto.Pagination{
					PageNumber:   1,
					ItemsPerPage: 10,
					SortDirection: func() *tkValueObject.PaginationSortDirection {
						sortDir, _ := tkValueObject.NewPaginationSortDirection("desc")
						return &sortDir
					}(),
				},
				expectError: false,
			},
			{
				inputMap: map[string]any{
					"lastSeenId": "xyz789",
				},
				expectedOutput: tkDto.Pagination{
					PageNumber:   1,
					ItemsPerPage: 10,
					LastSeenId: func() *tkValueObject.PaginationLastSeenId {
						lastSeenId, _ := tkValueObject.NewPaginationLastSeenId("xyz789")
						return &lastSeenId
					}(),
				},
				expectError: false,
			},
		}

		defaultPagination := tkDto.Pagination{
			PageNumber:   1,
			ItemsPerPage: 10,
		}

		for _, testCase := range testCaseStructs {
			actualOutput, conversionErr := PaginationDtoParser(defaultPagination, testCase.inputMap)
			if testCase.expectError && conversionErr == nil {
				subtest.Errorf("MissingExpectedError: [%v]", testCase.inputMap)
			}
			if !testCase.expectError && conversionErr != nil {
				subtest.Errorf("UnexpectedError: '%s' [%v]", conversionErr.Error(), testCase.inputMap)
			}
			if !testCase.expectError {
				if actualOutput.PageNumber != testCase.expectedOutput.PageNumber {
					subtest.Errorf("UnexpectedPageNumber: '%v' vs '%v'", actualOutput.PageNumber, testCase.expectedOutput.PageNumber)
				}
				if actualOutput.ItemsPerPage != testCase.expectedOutput.ItemsPerPage {
					subtest.Errorf("UnexpectedItemsPerPage: '%v' vs '%v'", actualOutput.ItemsPerPage, testCase.expectedOutput.ItemsPerPage)
				}
				if (actualOutput.SortBy == nil && testCase.expectedOutput.SortBy != nil) ||
					(actualOutput.SortBy != nil && testCase.expectedOutput.SortBy == nil) ||
					(actualOutput.SortBy != nil && *actualOutput.SortBy != *testCase.expectedOutput.SortBy) {
					subtest.Errorf("UnexpectedSortBy: '%v' vs '%v'", actualOutput.SortBy, testCase.expectedOutput.SortBy)
				}
				if (actualOutput.SortDirection == nil && testCase.expectedOutput.SortDirection != nil) ||
					(actualOutput.SortDirection != nil && testCase.expectedOutput.SortDirection == nil) ||
					(actualOutput.SortDirection != nil && *actualOutput.SortDirection != *testCase.expectedOutput.SortDirection) {
					subtest.Errorf("UnexpectedSortDirection: '%v' vs '%v'", actualOutput.SortDirection, testCase.expectedOutput.SortDirection)
				}
				if (actualOutput.LastSeenId == nil && testCase.expectedOutput.LastSeenId != nil) ||
					(actualOutput.LastSeenId != nil && testCase.expectedOutput.LastSeenId == nil) ||
					(actualOutput.LastSeenId != nil && *actualOutput.LastSeenId != *testCase.expectedOutput.LastSeenId) {
					subtest.Errorf("UnexpectedLastSeenId: '%v' vs '%v'", actualOutput.LastSeenId, testCase.expectedOutput.LastSeenId)
				}
			}
		}
	})

	test.Run("ErrorWithInvalidPageNumber", func(subtest *testing.T) {
		testCaseStructs := []struct {
			inputMap    map[string]any
			expectError bool
		}{
			{
				inputMap: map[string]any{
					"pageNumber": "invalid",
				},
				expectError: true,
			},
			{
				inputMap: map[string]any{
					"pageNumber": -1,
				},
				expectError: true,
			},
			{
				inputMap: map[string]any{
					"pageNumber": []int{1},
				},
				expectError: true,
			},
			{
				inputMap: map[string]any{
					"pageNumber": 1.5,
				},
				expectError: true,
			},
		}

		defaultPagination := tkDto.Pagination{
			PageNumber:   1,
			ItemsPerPage: 10,
		}

		for _, testCase := range testCaseStructs {
			_, conversionErr := PaginationDtoParser(defaultPagination, testCase.inputMap)
			if testCase.expectError && conversionErr == nil {
				subtest.Errorf("MissingExpectedError: [%v]", testCase.inputMap)
			}
			if !testCase.expectError && conversionErr != nil {
				subtest.Errorf("UnexpectedError: '%s' [%v]", conversionErr.Error(), testCase.inputMap)
			}
			if testCase.expectError && conversionErr != nil && conversionErr.Error() != "InvalidPageNumber" {
				subtest.Errorf("UnexpectedErrorMessage: expected 'InvalidPageNumber', got '%s'", conversionErr.Error())
			}
		}
	})

	test.Run("ErrorWithInvalidItemsPerPage", func(subtest *testing.T) {
		testCaseStructs := []struct {
			inputMap    map[string]any
			expectError bool
		}{
			{
				inputMap: map[string]any{
					"itemsPerPage": "invalid",
				},
				expectError: true,
			},
			{
				inputMap: map[string]any{
					"itemsPerPage": -1,
				},
				expectError: true,
			},
			{
				inputMap: map[string]any{
					"itemsPerPage": []int{1},
				},
				expectError: true,
			},
			{
				inputMap: map[string]any{
					"itemsPerPage": 0,
				},
				expectError: true,
			},
			{
				inputMap: map[string]any{
					"itemsPerPage": uint16(0),
				},
				expectError: true,
			},
			{
				inputMap: map[string]any{
					"itemsPerPage": "0",
				},
				expectError: true,
			},
			{
				inputMap: map[string]any{
					"itemsPerPage": 10.5,
				},
				expectError: true,
			},
		}

		defaultPagination := tkDto.Pagination{
			PageNumber:   1,
			ItemsPerPage: 10,
		}

		for _, testCase := range testCaseStructs {
			_, conversionErr := PaginationDtoParser(defaultPagination, testCase.inputMap)
			if testCase.expectError && conversionErr == nil {
				subtest.Errorf("MissingExpectedError: [%v]", testCase.inputMap)
			}
			if !testCase.expectError && conversionErr != nil {
				subtest.Errorf("UnexpectedError: '%s' [%v]", conversionErr.Error(), testCase.inputMap)
			}
			if testCase.expectError && conversionErr != nil && conversionErr.Error() != "InvalidItemsPerPage" {
				subtest.Errorf("UnexpectedErrorMessage: expected 'InvalidItemsPerPage', got '%s'", conversionErr.Error())
			}
		}
	})

	test.Run("ErrorWithInvalidSortBy", func(subtest *testing.T) {
		testCaseStructs := []struct {
			inputMap    map[string]any
			expectError bool
		}{
			{
				inputMap: map[string]any{
					"sortBy": "",
				},
				expectError: true,
			},
			{
				inputMap: map[string]any{
					"sortBy": "invalid@sort",
				},
				expectError: true,
			},
			{
				inputMap: map[string]any{
					"sortBy": []string{"name"},
				},
				expectError: true,
			},
		}

		defaultPagination := tkDto.Pagination{
			PageNumber:   1,
			ItemsPerPage: 10,
		}

		for _, testCase := range testCaseStructs {
			_, conversionErr := PaginationDtoParser(defaultPagination, testCase.inputMap)
			if testCase.expectError && conversionErr == nil {
				subtest.Errorf("MissingExpectedError: [%v]", testCase.inputMap)
			}
			if !testCase.expectError && conversionErr != nil {
				subtest.Errorf("UnexpectedError: '%s' [%v]", conversionErr.Error(), testCase.inputMap)
			}
		}
	})

	test.Run("ErrorWithInvalidSortDirection", func(subtest *testing.T) {
		testCaseStructs := []struct {
			inputMap    map[string]any
			expectError bool
		}{
			{
				inputMap: map[string]any{
					"sortDirection": "",
				},
				expectError: true,
			},
			{
				inputMap: map[string]any{
					"sortDirection": "ascending",
				},
				expectError: true,
			},
			{
				inputMap: map[string]any{
					"sortDirection": 123,
				},
				expectError: true,
			},
		}

		defaultPagination := tkDto.Pagination{
			PageNumber:   1,
			ItemsPerPage: 10,
		}

		for _, testCase := range testCaseStructs {
			_, conversionErr := PaginationDtoParser(defaultPagination, testCase.inputMap)
			if testCase.expectError && conversionErr == nil {
				subtest.Errorf("MissingExpectedError: [%v]", testCase.inputMap)
			}
			if !testCase.expectError && conversionErr != nil {
				subtest.Errorf("UnexpectedError: '%s' [%v]", conversionErr.Error(), testCase.inputMap)
			}
		}
	})

	test.Run("ErrorWithInvalidLastSeenId", func(subtest *testing.T) {
		testCaseStructs := []struct {
			inputMap    map[string]any
			expectError bool
		}{
			{
				inputMap: map[string]any{
					"lastSeenId": "",
				},
				expectError: true,
			},
			{
				inputMap: map[string]any{
					"lastSeenId": "with space",
				},
				expectError: true,
			},
			{
				inputMap: map[string]any{
					"lastSeenId": "with.dot",
				},
				expectError: true,
			},
		}

		defaultPagination := tkDto.Pagination{
			PageNumber:   1,
			ItemsPerPage: 10,
		}

		for _, testCase := range testCaseStructs {
			_, conversionErr := PaginationDtoParser(defaultPagination, testCase.inputMap)
			if testCase.expectError && conversionErr == nil {
				subtest.Errorf("MissingExpectedError: [%v]", testCase.inputMap)
			}
			if !testCase.expectError && conversionErr != nil {
				subtest.Errorf("UnexpectedError: '%s' [%v]", conversionErr.Error(), testCase.inputMap)
			}
		}
	})

	test.Run("SuccessWithEmptyInput", func(subtest *testing.T) {
		defaultPagination := tkDto.Pagination{
			PageNumber:   1,
			ItemsPerPage: 10,
		}

		actualOutput, conversionErr := PaginationDtoParser(defaultPagination, nil)
		if conversionErr != nil {
			subtest.Errorf("UnexpectedError: '%s'", conversionErr.Error())
		}

		if actualOutput.PageNumber != 1 {
			subtest.Errorf("UnexpectedPageNumber: '%v' vs '1'", actualOutput.PageNumber)
		}
		if actualOutput.ItemsPerPage != 10 {
			subtest.Errorf("UnexpectedItemsPerPage: '%v' vs '10'", actualOutput.ItemsPerPage)
		}
		if actualOutput.SortBy != nil {
			subtest.Errorf("UnexpectedSortBy: ShouldBeNil")
		}
		if actualOutput.SortDirection != nil {
			subtest.Errorf("UnexpectedSortDirection: ShouldBeNil")
		}
		if actualOutput.LastSeenId != nil {
			subtest.Errorf("UnexpectedLastSeenId: ShouldBeNil")
		}
	})
}
