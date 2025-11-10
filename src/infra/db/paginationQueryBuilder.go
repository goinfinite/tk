package tkInfraDb

import (
	"errors"
	"math"
	"strings"

	tkDto "github.com/goinfinite/tk/src/domain/dto"
	tkValueObject "github.com/goinfinite/tk/src/domain/valueObject"
	"github.com/iancoleman/strcase"
	"gorm.io/gorm"
)

const (
	errItemsPerPageCannotBeZero string = "ItemsPerPageCannotBeZero"
	errCountItemsTotalError     string = "CountItemsTotalError"
)

func PaginationQueryBuilder(
	dbQuery *gorm.DB,
	requestPagination tkDto.Pagination,
) (paginatedQuery *gorm.DB, responsePagination tkDto.Pagination, err error) {
	if requestPagination.ItemsPerPage == 0 {
		return paginatedQuery, responsePagination, errors.New(errItemsPerPageCannotBeZero)
	}

	var itemsTotal int64
	err = dbQuery.Count(&itemsTotal).Error
	if err != nil {
		return paginatedQuery, responsePagination, errors.New(errCountItemsTotalError + ": " + err.Error())
	}

	paginatedQuery = dbQuery.Limit(int(requestPagination.ItemsPerPage))
	switch requestPagination.LastSeenId {
	case nil:
		if requestPagination.PageNumber > 0 {
			offset := int(requestPagination.PageNumber) * int(requestPagination.ItemsPerPage)
			paginatedQuery = paginatedQuery.Offset(offset)
		}
	default:
		paginatedQuery = paginatedQuery.Where("id > ?", requestPagination.LastSeenId.String())
	}

	orderStatement := "id " + tkValueObject.PaginationSortDirectionAsc.String()
	if requestPagination.SortBy != nil {
		orderStatement = requestPagination.SortBy.String()
		orderStatement = strings.ToLower(orderStatement)
		orderStatement = strcase.ToSnake(orderStatement)
		if requestPagination.SortDirection != nil {
			orderStatement += " " + requestPagination.SortDirection.String()
		}
	}
	paginatedQuery = paginatedQuery.Order(orderStatement)

	itemsTotalUint := uint64(itemsTotal)
	pagesTotal := uint32(
		math.Ceil(float64(itemsTotal) / float64(requestPagination.ItemsPerPage)),
	)

	return paginatedQuery, tkDto.Pagination{
		PageNumber:    requestPagination.PageNumber,
		ItemsPerPage:  requestPagination.ItemsPerPage,
		SortBy:        requestPagination.SortBy,
		SortDirection: requestPagination.SortDirection,
		PagesTotal:    &pagesTotal,
		ItemsTotal:    &itemsTotalUint,
	}, nil
}
