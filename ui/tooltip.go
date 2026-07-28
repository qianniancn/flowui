package ui

import (
	"github.com/qianniancn/flowui/internal/components/tooltip"
	"github.com/qianniancn/flowui/internal/overlay"
)

type TooltipWidget = tooltip.TooltipWidget
type TooltipTrigger = tooltip.TooltipTrigger
type TooltipPlacement = overlay.PopoverPlacement

const (
	TooltipHover = tooltip.TooltipHover
	TooltipFocus = tooltip.TooltipFocus

	TooltipBottom      = overlay.PopoverBottom
	TooltipTop         = overlay.PopoverTop
	TooltipLeft        = overlay.PopoverLeft
	TooltipRight       = overlay.PopoverRight
	TooltipBottomStart = overlay.PopoverBottomStart
	TooltipBottomEnd   = overlay.PopoverBottomEnd
	TooltipTopStart    = overlay.PopoverTopStart
	TooltipTopEnd      = overlay.PopoverTopEnd
	TooltipLeftStart   = overlay.PopoverLeftStart
	TooltipLeftEnd     = overlay.PopoverLeftEnd
	TooltipRightStart  = overlay.PopoverRightStart
	TooltipRightEnd    = overlay.PopoverRightEnd
)

func Tooltip(key string, trigger Widget, content Widget) TooltipWidget {
	return tooltip.Tooltip(key, trigger, content)
}
