package tkInfraDb

import "errors"

var (
	ErrItemsPerPageCannotBeZero = errors.New("ItemsPerPageCannotBeZero")
	ErrPagesTotalOverflow       = errors.New("PagesTotalOverflow")
)

func PaginationPagesTotalResolver(
	itemsTotal uint64,
	itemsPerPage uint16,
) (pagesTotal uint32, err error) {
	if itemsPerPage == 0 {
		return 0, ErrItemsPerPageCannotBeZero
	}

	perPage := uint64(itemsPerPage)
	fullPages := itemsTotal / perPage
	if itemsTotal%perPage != 0 {
		fullPages++
	}

	maxUint32 := uint64(^uint32(0))
	if fullPages > maxUint32 {
		return 0, ErrPagesTotalOverflow
	}

	return uint32(fullPages), nil
}
