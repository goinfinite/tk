package tkInfraDb

import (
	"errors"

	tkDto "github.com/goinfinite/tk/src/domain/dto"
	tkValueObject "github.com/goinfinite/tk/src/domain/valueObject"
	"gorm.io/gorm"
)

const (
	errCountItemsTotalError            string = "CountItemsTotalError"
	errParseStatementSchemaError       string = "ParseStatementSchemaError"
	errUnknownPaginationSortFieldError string = "UnknownPaginationSortFieldError"
	errModelWithoutPrimaryKeyError     string = "ModelWithoutPrimaryKeyError"
)

func resolveOrderStatement(
	statement *gorm.Statement,
	sortBy tkValueObject.PaginationSortBy,
	sortDirectionPtr *tkValueObject.PaginationSortDirection,
) (orderStatement string, err error) {
	sortField := statement.Schema.LookUpField(sortBy.String())
	if sortField == nil {
		return "", errors.New(
			errUnknownPaginationSortFieldError + ": " + sortBy.String(),
		)
	}

	orderStatement = statement.Quote(sortField.DBName)
	if sortDirectionPtr != nil {
		orderStatement += " " + sortDirectionPtr.String()
	}
	return orderStatement, nil
}

func PaginationQueryBuilder(
	dbQuery *gorm.DB,
	requestPagination tkDto.Pagination,
	primaryKeyColumn string,
) (paginatedQuery *gorm.DB, responsePagination tkDto.Pagination, err error) {
	err = dbQuery.Statement.Parse(dbQuery.Statement.Model)
	if err != nil {
		return paginatedQuery, responsePagination,
			errors.New(errParseStatementSchemaError + ": " + err.Error())
	}

	var itemsTotal int64
	err = dbQuery.Count(&itemsTotal).Error
	if err != nil {
		return paginatedQuery, responsePagination,
			errors.New(errCountItemsTotalError + ": " + err.Error())
	}

	paginatedQuery = dbQuery.Limit(int(requestPagination.ItemsPerPage))
	switch requestPagination.LastSeenId {
	case nil:
		if requestPagination.PageNumber > 0 {
			offset := int(requestPagination.PageNumber) * int(requestPagination.ItemsPerPage)
			paginatedQuery = paginatedQuery.Offset(offset)
		}
	default:
		cursorColumn := primaryKeyColumn
		if cursorColumn == "" {
			primaryField := dbQuery.Statement.Schema.PrioritizedPrimaryField
			if primaryField == nil {
				return paginatedQuery, responsePagination,
					errors.New(errModelWithoutPrimaryKeyError)
			}
			cursorColumn = primaryField.DBName
		}
		paginatedQuery = paginatedQuery.Where(
			dbQuery.Statement.Quote(cursorColumn)+" > ?",
			requestPagination.LastSeenId.String(),
		)
	}

	if requestPagination.SortBy != nil {
		orderStatement, resolveErr := resolveOrderStatement(
			dbQuery.Statement,
			*requestPagination.SortBy,
			requestPagination.SortDirection,
		)
		if resolveErr != nil {
			return paginatedQuery, responsePagination, resolveErr
		}
		paginatedQuery = paginatedQuery.Order(orderStatement)
	}

	itemsTotalUint := uint64(itemsTotal)
	pagesTotal, err := PaginationPagesTotalResolver(
		itemsTotalUint, requestPagination.ItemsPerPage,
	)
	if err != nil {
		return paginatedQuery, responsePagination, err
	}

	return paginatedQuery, tkDto.Pagination{
		PageNumber:    requestPagination.PageNumber,
		ItemsPerPage:  requestPagination.ItemsPerPage,
		SortBy:        requestPagination.SortBy,
		SortDirection: requestPagination.SortDirection,
		PagesTotal:    &pagesTotal,
		ItemsTotal:    &itemsTotalUint,
	}, nil
}
