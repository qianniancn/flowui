package ui

import "github.com/qianniancn/FlowUI/internal/components/pagination"

type PaginationWidget = pagination.Widget
type PaginationSize = pagination.Size

const (
	PaginationMedium = pagination.SizeMedium
	PaginationSmall  = pagination.SizeSmall
	PaginationLarge  = pagination.SizeLarge
)

func Pagination(key string, page, total int) PaginationWidget {
	return pagination.New(key, page, total)
}
