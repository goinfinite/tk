package tkDto

import tkValueObject "github.com/goinfinite/tk/src/domain/valueObject"

var (
	PaginationSingleItem  = Pagination{PageNumber: 0, ItemsPerPage: 1}
	PaginationUnpaginated = Pagination{PageNumber: 0, ItemsPerPage: 1000}
)

type Pagination struct {
	PageNumber    uint32                                 `json:"pageNumber"`
	ItemsPerPage  uint16                                 `json:"itemsPerPage"`
	SortBy        *tkValueObject.PaginationSortBy        `json:"sortBy"`
	SortDirection *tkValueObject.PaginationSortDirection `json:"sortDirection"`
	LastSeenId    *tkValueObject.PaginationLastSeenId    `json:"lastSeenId"`
	PagesTotal    *uint32                                `json:"pagesTotal"`
	ItemsTotal    *uint64                                `json:"itemsTotal"`
}
