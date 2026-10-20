package tooltip

import "github.com/qianniancn/flowui/internal/core"

// Placement identifies the side and alignment used for a tooltip panel.
// It is the same value type the overlay engine consumes, so a placement may be
// passed straight to Widget.Placement.
type Placement = core.PopoverPlacement

// Side placements ask for a side without pinning the cross-axis alignment; the
// engine flips the panel when the requested side runs out of space.
const (
	Bottom = core.PopoverBottom
	Top    = core.PopoverTop
	Left   = core.PopoverLeft
	Right  = core.PopoverRight
)

// Aligned placements pin both the side and the cross-axis alignment, so the
// panel does not slide along the trigger edge when space is tight.
const (
	BottomStart = core.PopoverBottomStart
	BottomEnd   = core.PopoverBottomEnd
	TopStart    = core.PopoverTopStart
	TopEnd      = core.PopoverTopEnd
	LeftStart   = core.PopoverLeftStart
	LeftEnd     = core.PopoverLeftEnd
	RightStart  = core.PopoverRightStart
	RightEnd    = core.PopoverRightEnd
)
