package tkPresentation

import (
	"errors"

	tkDto "github.com/goinfinite/tk/src/domain/dto"
	tkValueObject "github.com/goinfinite/tk/src/domain/valueObject"
	tkVoUtil "github.com/goinfinite/tk/src/domain/valueObject/util"
)

func PaginationParser(
	defaultPagination tkDto.Pagination,
	untrustedInput map[string]any,
) (parsedPagination tkDto.Pagination, err error) {
	parsedPagination = defaultPagination

	if untrustedInput["pageNumber"] != nil {
		pageNumber, err := tkVoUtil.InterfaceToUint32(untrustedInput["pageNumber"])
		if err != nil {
			return parsedPagination, errors.New("InvalidPageNumber")
		}
		parsedPagination.PageNumber = pageNumber
	}

	if untrustedInput["itemsPerPage"] != nil {
		itemsPerPage, err := tkVoUtil.InterfaceToUint16(untrustedInput["itemsPerPage"])
		if err != nil {
			return parsedPagination, errors.New("InvalidItemsPerPage")
		}
		parsedPagination.ItemsPerPage = itemsPerPage
	}

	if untrustedInput["sortBy"] != nil {
		sortBy, err := tkValueObject.NewPaginationSortBy(untrustedInput["sortBy"])
		if err != nil {
			return parsedPagination, err
		}
		parsedPagination.SortBy = &sortBy
	}

	if untrustedInput["sortDirection"] != nil {
		sortDirection, err := tkValueObject.NewPaginationSortDirection(untrustedInput["sortDirection"])
		if err != nil {
			return parsedPagination, err
		}
		parsedPagination.SortDirection = &sortDirection
	}

	if untrustedInput["lastSeenId"] != nil {
		lastSeenId, err := tkValueObject.NewPaginationLastSeenId(untrustedInput["lastSeenId"])
		if err != nil {
			return parsedPagination, err
		}
		parsedPagination.LastSeenId = &lastSeenId
	}

	return parsedPagination, nil
}
