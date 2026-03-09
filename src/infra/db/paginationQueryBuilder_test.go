package tkInfraDb

import (
	"fmt"
	"testing"
	"time"

	"github.com/glebarez/sqlite"
	tkDto "github.com/goinfinite/tk/src/domain/dto"
	tkValueObject "github.com/goinfinite/tk/src/domain/valueObject"
	"gorm.io/gorm"
)

type testPaginationModel struct {
	ID   uint64 `gorm:"primaryKey"`
	Name string
}

type testPaginationCustomPkModel struct {
	CustomID uint64 `gorm:"column:custom_id;primaryKey"`
	Name     string
}

func TestPaginationQueryBuilder(t *testing.T) {
	sortByName, _ := tkValueObject.NewPaginationSortBy("name")
	lastSeenId5, _ := tkValueObject.NewPaginationLastSeenId("5")

	testCases := []struct {
		name                  string
		primaryKeyColumn      string
		requestPagination     tkDto.Pagination
		expectedError         string
		expectedItemsTotal    uint64
		expectedPagesTotal    uint32
		expectedResultCount   int
		expectedFirstItemName string
	}{
		{
			name:             "ItemsPerPageZero",
			primaryKeyColumn: "id",
			requestPagination: tkDto.Pagination{
				PageNumber:   0,
				ItemsPerPage: 0,
			},
			expectedError: errItemsPerPageCannotBeZero,
		},
		{
			name:             "FirstPage",
			primaryKeyColumn: "id",
			requestPagination: tkDto.Pagination{
				PageNumber:   0,
				ItemsPerPage: 3,
			},
			expectedItemsTotal:    10,
			expectedPagesTotal:    4,
			expectedResultCount:   3,
			expectedFirstItemName: "item1",
		},
		{
			name:             "SecondPage",
			primaryKeyColumn: "id",
			requestPagination: tkDto.Pagination{
				PageNumber:   1,
				ItemsPerPage: 3,
			},
			expectedItemsTotal:    10,
			expectedPagesTotal:    4,
			expectedResultCount:   3,
			expectedFirstItemName: "item4",
		},
		{
			name:             "LastSeenId",
			primaryKeyColumn: "id",
			requestPagination: tkDto.Pagination{
				ItemsPerPage: 3,
				LastSeenId:   &lastSeenId5,
			},
			expectedItemsTotal:    10,
			expectedPagesTotal:    4,
			expectedResultCount:   3,
			expectedFirstItemName: "item6",
		},
		{
			name:             "SortByNameAsc",
			primaryKeyColumn: "id",
			requestPagination: tkDto.Pagination{
				PageNumber:   0,
				ItemsPerPage: 3,
				SortBy:       &sortByName,
				SortDirection: &tkValueObject.PaginationSortDirectionAsc,
			},
			expectedItemsTotal:    10,
			expectedPagesTotal:    4,
			expectedResultCount:   3,
			expectedFirstItemName: "item1",
		},
		{
			name:             "SortByNameDesc",
			primaryKeyColumn: "id",
			requestPagination: tkDto.Pagination{
				PageNumber:   0,
				ItemsPerPage: 3,
				SortBy:       &sortByName,
				SortDirection: &tkValueObject.PaginationSortDirectionDesc,
			},
			expectedItemsTotal:    10,
			expectedPagesTotal:    4,
			expectedResultCount:   3,
			expectedFirstItemName: "item9",
		},
		{
			name:             "NoSortByNoOrderBy",
			primaryKeyColumn: "",
			requestPagination: tkDto.Pagination{
				PageNumber:   0,
				ItemsPerPage: 5,
			},
			expectedItemsTotal:  10,
			expectedPagesTotal:  2,
			expectedResultCount: 5,
		},
		{
			name:             "LastSeenIdWithoutPrimaryKeyColumn",
			primaryKeyColumn: "",
			requestPagination: tkDto.Pagination{
				ItemsPerPage: 3,
				LastSeenId:   &lastSeenId5,
			},
			expectedItemsTotal:    10,
			expectedPagesTotal:    4,
			expectedResultCount:   3,
			expectedFirstItemName: "item6",
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			dbSvc := setupTestDb(t)
			dbQuery := dbSvc.Model(&testPaginationModel{})

			paginatedQuery, responsePagination, err := PaginationQueryBuilder(
				dbQuery,
				testCase.requestPagination,
				testCase.primaryKeyColumn,
			)

			if testCase.expectedError != "" {
				if err == nil {
					t.Errorf(
						"MissingExpectedError: %s",
						testCase.expectedError,
					)
					return
				}
				if err.Error() != testCase.expectedError {
					t.Errorf(
						"UnexpectedErrorMessage: '%s' vs '%s'",
						err.Error(), testCase.expectedError,
					)
				}
				return
			}

			if err != nil {
				t.Errorf("UnexpectedError: %v", err)
				return
			}

			if responsePagination.ItemsTotal == nil ||
				*responsePagination.ItemsTotal != testCase.expectedItemsTotal {
				t.Errorf(
					"ItemsTotalMismatch: expected %d, got %v",
					testCase.expectedItemsTotal,
					responsePagination.ItemsTotal,
				)
			}

			if responsePagination.PagesTotal == nil ||
				*responsePagination.PagesTotal != testCase.expectedPagesTotal {
				t.Errorf(
					"PagesTotalMismatch: expected %d, got %v",
					testCase.expectedPagesTotal,
					responsePagination.PagesTotal,
				)
			}

			var queryResults []testPaginationModel
			err = paginatedQuery.Find(&queryResults).Error
			if err != nil {
				t.Errorf("ExecuteQueryFailed: %v", err)
				return
			}

			if len(queryResults) != testCase.expectedResultCount {
				t.Errorf(
					"ResultCountMismatch: expected %d, got %d",
					testCase.expectedResultCount,
					len(queryResults),
				)
			}

			if testCase.expectedFirstItemName != "" &&
				len(queryResults) > 0 &&
				queryResults[0].Name != testCase.expectedFirstItemName {
				t.Errorf(
					"FirstItemNameMismatch: expected %s, got %s",
					testCase.expectedFirstItemName,
					queryResults[0].Name,
				)
			}
		})
	}

	t.Run("LastSeenIdWithCustomPrimaryKey", func(t *testing.T) {
		dbSvc := setupTestDb(t)
		dbQuery := dbSvc.Model(&testPaginationCustomPkModel{})

		paginatedQuery, responsePagination, err := PaginationQueryBuilder(
			dbQuery,
			tkDto.Pagination{
				ItemsPerPage: 3,
				LastSeenId:   &lastSeenId5,
			},
			"custom_id",
		)
		if err != nil {
			t.Fatalf("UnexpectedError: %v", err)
		}

		if responsePagination.ItemsTotal == nil ||
			*responsePagination.ItemsTotal != 10 {
			t.Errorf(
				"ItemsTotalMismatch: expected 10, got %v",
				responsePagination.ItemsTotal,
			)
		}

		var queryResults []testPaginationCustomPkModel
		err = paginatedQuery.Find(&queryResults).Error
		if err != nil {
			t.Fatalf("ExecuteQueryFailed: %v", err)
		}

		if len(queryResults) != 3 {
			t.Errorf(
				"ResultCountMismatch: expected 3, got %d",
				len(queryResults),
			)
		}

		if len(queryResults) > 0 && queryResults[0].Name != "item6" {
			t.Errorf(
				"FirstItemNameMismatch: expected item6, got %s",
				queryResults[0].Name,
			)
		}
	})
}

func setupTestDb(t *testing.T) *gorm.DB {
	t.Helper()
	dbSvc, err := gorm.Open(
		sqlite.Open("file::memory:"), &gorm.Config{
			NowFunc: func() time.Time { return time.Now().UTC() },
		},
	)
	if err != nil {
		t.Fatalf("OpenTestDbFailed: %v", err)
	}

	err = dbSvc.AutoMigrate(
		&testPaginationModel{},
		&testPaginationCustomPkModel{},
	)
	if err != nil {
		t.Fatalf("MigrateTestDbFailed: %v", err)
	}

	for itemIndex := 1; itemIndex <= 10; itemIndex++ {
		itemName := fmt.Sprintf("item%d", itemIndex)
		err = dbSvc.Create(
			&testPaginationModel{
				ID: uint64(itemIndex), Name: itemName,
			},
		).Error
		if err != nil {
			t.Fatalf("InsertTestDataFailed: %v", err)
		}
		err = dbSvc.Create(
			&testPaginationCustomPkModel{
				CustomID: uint64(itemIndex), Name: itemName,
			},
		).Error
		if err != nil {
			t.Fatalf("InsertTestDataFailed: %v", err)
		}
	}

	return dbSvc
}
