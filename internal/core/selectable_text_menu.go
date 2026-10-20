package core

import (
	"gioui.org/layout"
	"gioui.org/widget"
)

// SelectableEditAction identifies an edit operation on selectable text.
type SelectableEditAction uint8

const (
	SelectableEditCopy SelectableEditAction = iota + 1
	SelectableEditSelectAll
)

// SelectableTextMenu decorates a selectable-text field with its edit menu.
// The menu layer installs an implementation in its init; text stays decoupled
// from menu because menu renders its item labels through the text package.
// When nil, selectable text renders without a menu.
var SelectableTextMenu func(ctx *Context, gtx layout.Context, key string, field Widget, selectable *widget.Selectable, queue func(SelectableEditAction)) layout.Dimensions
