package style

import "gioui.org/io/pointer"

// Cursor identifies the pointer cursor shown over a Widget.
type Cursor = pointer.Cursor

// Pointer cursor values understood by the style package.
const (
	CursorDefault                  = pointer.CursorDefault
	CursorNone                     = pointer.CursorNone
	CursorText                     = pointer.CursorText
	CursorVerticalText             = pointer.CursorVerticalText
	CursorPointer                  = pointer.CursorPointer
	CursorCrosshair                = pointer.CursorCrosshair
	CursorAllScroll                = pointer.CursorAllScroll
	CursorColResize                = pointer.CursorColResize
	CursorRowResize                = pointer.CursorRowResize
	CursorGrab                     = pointer.CursorGrab
	CursorGrabbing                 = pointer.CursorGrabbing
	CursorNotAllowed               = pointer.CursorNotAllowed
	CursorWait                     = pointer.CursorWait
	CursorProgress                 = pointer.CursorProgress
	CursorNorthWestResize          = pointer.CursorNorthWestResize
	CursorNorthEastResize          = pointer.CursorNorthEastResize
	CursorSouthWestResize          = pointer.CursorSouthWestResize
	CursorSouthEastResize          = pointer.CursorSouthEastResize
	CursorNorthSouthResize         = pointer.CursorNorthSouthResize
	CursorEastWestResize           = pointer.CursorEastWestResize
	CursorWestResize               = pointer.CursorWestResize
	CursorEastResize               = pointer.CursorEastResize
	CursorNorthResize              = pointer.CursorNorthResize
	CursorSouthResize              = pointer.CursorSouthResize
	CursorNorthEastSouthWestResize = pointer.CursorNorthEastSouthWestResize
	CursorNorthWestSouthEastResize = pointer.CursorNorthWestSouthEastResize
)
