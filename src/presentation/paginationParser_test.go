package tkPresentation

import (
	"testing"

	tkDto "github.com/goinfinite/tk/src/domain/dto"
	tkValueObject "github.com/goinfinite/tk/src/domain/valueObject"
)

func TestPaginationParser(t *testing.T) {
	t.Run("SuccessWithAllValidFields", func(t *testing.T) {
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
		}

		defaultPagination := tkDto.Pagination{
			PageNumber:   1,
			ItemsPerPage: 10,
		}

		for _, testCase := range testCaseStructs {
			actualOutput, conversionErr := PaginationParser(defaultPagination, testCase.inputMap)
			if testCase.expectError && conversionErr == nil {
				t.Errorf("MissingExpectedError: [%v]", testCase.inputMap)
			}
			if !testCase.expectError && conversionErr != nil {
				t.Errorf("UnexpectedError: '%s' [%v]", conversionErr.Error(), testCase.inputMap)
			}
			if !testCase.expectError {
				if actualOutput.PageNumber != testCase.expectedOutput.PageNumber {
					t.Errorf("UnexpectedPageNumber: '%v' vs '%v'", actualOutput.PageNumber, testCase.expectedOutput.PageNumber)
				}
				if actualOutput.ItemsPerPage != testCase.expectedOutput.ItemsPerPage {
					t.Errorf("UnexpectedItemsPerPage: '%v' vs '%v'", actualOutput.ItemsPerPage, testCase.expectedOutput.ItemsPerPage)
				}
				if (actualOutput.SortBy == nil && testCase.expectedOutput.SortBy != nil) ||
					(actualOutput.SortBy != nil && testCase.expectedOutput.SortBy == nil) ||
					(actualOutput.SortBy != nil && *actualOutput.SortBy != *testCase.expectedOutput.SortBy) {
					t.Errorf("UnexpectedSortBy: '%v' vs '%v'", actualOutput.SortBy, testCase.expectedOutput.SortBy)
				}
				if (actualOutput.SortDirection == nil && testCase.expectedOutput.SortDirection != nil) ||
					(actualOutput.SortDirection != nil && testCase.expectedOutput.SortDirection == nil) ||
					(actualOutput.SortDirection != nil && *actualOutput.SortDirection != *testCase.expectedOutput.SortDirection) {
					t.Errorf("UnexpectedSortDirection: '%v' vs '%v'", actualOutput.SortDirection, testCase.expectedOutput.SortDirection)
				}
				if (actualOutput.LastSeenId == nil && testCase.expectedOutput.LastSeenId != nil) ||
					(actualOutput.LastSeenId != nil && testCase.expectedOutput.LastSeenId == nil) ||
					(actualOutput.LastSeenId != nil && *actualOutput.LastSeenId != *testCase.expectedOutput.LastSeenId) {
					t.Errorf("UnexpectedLastSeenId: '%v' vs '%v'", actualOutput.LastSeenId, testCase.expectedOutput.LastSeenId)
				}
			}
		}
	})

	t.Run("SuccessWithPartialFields", func(t *testing.T) {
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
			actualOutput, conversionErr := PaginationParser(defaultPagination, testCase.inputMap)
			if testCase.expectError && conversionErr == nil {
				t.Errorf("MissingExpectedError: [%v]", testCase.inputMap)
			}
			if !testCase.expectError && conversionErr != nil {
				t.Errorf("UnexpectedError: '%s' [%v]", conversionErr.Error(), testCase.inputMap)
			}
			if !testCase.expectError {
				if actualOutput.PageNumber != testCase.expectedOutput.PageNumber {
					t.Errorf("UnexpectedPageNumber: '%v' vs '%v'", actualOutput.PageNumber, testCase.expectedOutput.PageNumber)
				}
				if actualOutput.ItemsPerPage != testCase.expectedOutput.ItemsPerPage {
					t.Errorf("UnexpectedItemsPerPage: '%v' vs '%v'", actualOutput.ItemsPerPage, testCase.expectedOutput.ItemsPerPage)
				}
				if (actualOutput.SortBy == nil && testCase.expectedOutput.SortBy != nil) ||
					(actualOutput.SortBy != nil && testCase.expectedOutput.SortBy == nil) ||
					(actualOutput.SortBy != nil && *actualOutput.SortBy != *testCase.expectedOutput.SortBy) {
					t.Errorf("UnexpectedSortBy: '%v' vs '%v'", actualOutput.SortBy, testCase.expectedOutput.SortBy)
				}
				if (actualOutput.SortDirection == nil && testCase.expectedOutput.SortDirection != nil) ||
					(actualOutput.SortDirection != nil && testCase.expectedOutput.SortDirection == nil) ||
					(actualOutput.SortDirection != nil && *actualOutput.SortDirection != *testCase.expectedOutput.SortDirection) {
					t.Errorf("UnexpectedSortDirection: '%v' vs '%v'", actualOutput.SortDirection, testCase.expectedOutput.SortDirection)
				}
				if (actualOutput.LastSeenId == nil && testCase.expectedOutput.LastSeenId != nil) ||
					(actualOutput.LastSeenId != nil && testCase.expectedOutput.LastSeenId == nil) ||
					(actualOutput.LastSeenId != nil && *actualOutput.LastSeenId != *testCase.expectedOutput.LastSeenId) {
					t.Errorf("UnexpectedLastSeenId: '%v' vs '%v'", actualOutput.LastSeenId, testCase.expectedOutput.LastSeenId)
				}
			}
		}
	})

	t.Run("ErrorWithInvalidPageNumber", func(t *testing.T) {
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
		}

		defaultPagination := tkDto.Pagination{
			PageNumber:   1,
			ItemsPerPage: 10,
		}

		for _, testCase := range testCaseStructs {
			_, conversionErr := PaginationParser(defaultPagination, testCase.inputMap)
			if testCase.expectError && conversionErr == nil {
				t.Errorf("MissingExpectedError: [%v]", testCase.inputMap)
			}
			if !testCase.expectError && conversionErr != nil {
				t.Errorf("UnexpectedError: '%s' [%v]", conversionErr.Error(), testCase.inputMap)
			}
			if testCase.expectError && conversionErr != nil && conversionErr.Error() != "InvalidPageNumber" {
				t.Errorf("UnexpectedErrorMessage: expected 'InvalidPageNumber', got '%s'", conversionErr.Error())
			}
		}
	})

	t.Run("ErrorWithInvalidItemsPerPage", func(t *testing.T) {
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
		}

		defaultPagination := tkDto.Pagination{
			PageNumber:   1,
			ItemsPerPage: 10,
		}

		for _, testCase := range testCaseStructs {
			_, conversionErr := PaginationParser(defaultPagination, testCase.inputMap)
			if testCase.expectError && conversionErr == nil {
				t.Errorf("MissingExpectedError: [%v]", testCase.inputMap)
			}
			if !testCase.expectError && conversionErr != nil {
				t.Errorf("UnexpectedError: '%s' [%v]", conversionErr.Error(), testCase.inputMap)
			}
			if testCase.expectError && conversionErr != nil && conversionErr.Error() != "InvalidItemsPerPage" {
				t.Errorf("UnexpectedErrorMessage: expected 'InvalidItemsPerPage', got '%s'", conversionErr.Error())
			}
		}
	})

	t.Run("ErrorWithInvalidSortBy", func(t *testing.T) {
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
			_, conversionErr := PaginationParser(defaultPagination, testCase.inputMap)
			if testCase.expectError && conversionErr == nil {
				t.Errorf("MissingExpectedError: [%v]", testCase.inputMap)
			}
			if !testCase.expectError && conversionErr != nil {
				t.Errorf("UnexpectedError: '%s' [%v]", conversionErr.Error(), testCase.inputMap)
			}
		}
	})

	t.Run("ErrorWithInvalidSortDirection", func(t *testing.T) {
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
			_, conversionErr := PaginationParser(defaultPagination, testCase.inputMap)
			if testCase.expectError && conversionErr == nil {
				t.Errorf("MissingExpectedError: [%v]", testCase.inputMap)
			}
			if !testCase.expectError && conversionErr != nil {
				t.Errorf("UnexpectedError: '%s' [%v]", conversionErr.Error(), testCase.inputMap)
			}
		}
	})

	t.Run("ErrorWithInvalidLastSeenId", func(t *testing.T) {
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
			_, conversionErr := PaginationParser(defaultPagination, testCase.inputMap)
			if testCase.expectError && conversionErr == nil {
				t.Errorf("MissingExpectedError: [%v]", testCase.inputMap)
			}
			if !testCase.expectError && conversionErr != nil {
				t.Errorf("UnexpectedError: '%s' [%v]", conversionErr.Error(), testCase.inputMap)
			}
		}
	})

	t.Run("SuccessWithEmptyInput", func(t *testing.T) {
		defaultPagination := tkDto.Pagination{
			PageNumber:   1,
			ItemsPerPage: 10,
		}

		actualOutput, conversionErr := PaginationParser(defaultPagination, nil)
		if conversionErr != nil {
			t.Errorf("UnexpectedError: '%s'", conversionErr.Error())
		}

		if actualOutput.PageNumber != 1 {
			t.Errorf("UnexpectedPageNumber: '%v' vs '1'", actualOutput.PageNumber)
		}
		if actualOutput.ItemsPerPage != 10 {
			t.Errorf("UnexpectedItemsPerPage: '%v' vs '10'", actualOutput.ItemsPerPage)
		}
		if actualOutput.SortBy != nil {
			t.Errorf("UnexpectedSortBy: ShouldBeNil")
		}
		if actualOutput.SortDirection != nil {
			t.Errorf("UnexpectedSortDirection: ShouldBeNil")
		}
		if actualOutput.LastSeenId != nil {
			t.Errorf("UnexpectedLastSeenId: ShouldBeNil")
		}
	})
}
