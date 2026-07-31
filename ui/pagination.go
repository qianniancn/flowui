package ui

import "github.com/qianniancn/flowui/internal/components/pagination"

type PaginationWidget = pagination.Widget

// PaginationSize selects the size of a pagination control.
type PaginationSize = pagination.Size

const (
	PaginationMedium = pagination.SizeMedium
	PaginationSmall  = pagination.SizeSmall
	PaginationLarge  = pagination.SizeLarge
)

// Pagination creates a page selector for page out of total pages.
func Pagination(key string, page, total int) PaginationWidget {
	return pagination.New(key, page, total)
}
