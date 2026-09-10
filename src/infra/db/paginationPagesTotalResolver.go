package tkInfraDb

import "errors"

var ErrItemsPerPageCannotBeZero = errors.New("ItemsPerPageCannotBeZero")

func PaginationPagesTotalResolver(
	itemsTotal uint64,
	itemsPerPage uint16,
) (pagesTotal uint32, err error) {
	if itemsPerPage == 0 {
		return 0, ErrItemsPerPageCannotBeZero
	}

	perPage := uint64(itemsPerPage)
	fullPagesIncludingPartialLast := (itemsTotal + perPage - 1) / perPage
	return uint32(fullPagesIncludingPartialLast), nil
}
