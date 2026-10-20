// Package dropdownstyle contains the theme-owned visual recipe for dropdown
// behavior. It deliberately does not import widget/dropdown.
package dropdownstyle

import (
	"image"
	"sync"

	"gioui.org/layout"
	"gioui.org/op/clip"
	"gioui.org/op/paint"
	"gioui.org/widget"
	lucide "github.com/qianniancn/flowui-icons-lucide"
	"github.com/qianniancn/flowui/theme"
)

// BuildDropdown returns the complete dropdown visual protocol for the active
// theme. Popup menu declarations continue to come from MenuProvider.
func BuildDropdown(active *theme.Theme, request theme.DropdownRequest) theme.DropdownVisual {
	if active == nil {
		panic(theme.ErrMissingVisualProvider)
	}
	tokens := active.Components.Dropdown
	return theme.DropdownVisual{
		TriggerPressedScale: pressedScale(tokens.TriggerPressedScale),
		PressDuration:       tokens.PressDuration,
		EnterDuration:       tokens.EnterDuration,
		ExitDuration:        tokens.ExitDuration,
		LongPressDuration:   tokens.LongPressDuration,
		HoverOpenDuration:   tokens.HoverOpenDuration,
		HoverCloseDuration:  tokens.HoverCloseDuration,
		PanelGap:            tokens.PanelGap,
		ArrowSize:           tokens.ArrowSize,
		Motion:              active.Motion,
		TriggerFocus:        focusRenderer(tokens, active.Semantic()),
		TriggerGlyph:        triggerGlyphRenderer(),
		Override:            theme.DropdownOverrideStyle(request),
	}
}

var triggerGlyph struct {
	sync.Once
	value *widget.Icon
}

func triggerGlyphRenderer() theme.DropdownTriggerGlyphRenderer {
	return func(gtx layout.Context, request theme.DropdownTriggerGlyphRequest) layout.Dimensions {
		const size = 16
		glyphGtx := gtx
		glyphGtx.Constraints = layout.Exact(image.Pt(gtx.Dp(size), gtx.Dp(size)))
		triggerGlyphIcon().Layout(glyphGtx, request.Color)
		return layout.Dimensions{Size: glyphGtx.Constraints.Min}
	}
}

func triggerGlyphIcon() *widget.Icon {
	triggerGlyph.Do(func() {
		value, err := widget.NewIcon(lucide.Ellipsis)
		if err == nil {
			triggerGlyph.value = value
		}
	})
	if triggerGlyph.value == nil {
		panic(theme.ErrIncompleteVisualProvider)
	}
	return triggerGlyph.value
}

func pressedScale(value float32) float32 {
	if value <= 0 || value > 1 {
		return .98
	}
	return value
}

func focusRenderer(tokens theme.DropdownTheme, semantic theme.Semantic) theme.DropdownTriggerFocusRenderer {
	width := tokens.TriggerFocusRingWidth
	offset := tokens.TriggerFocusRingOffset
	radius := tokens.TriggerFocusRadius
	colorValue := theme.ColorOr(tokens.FocusColor, semantic.Ring)
	return func(gtx layout.Context, request theme.DropdownTriggerFocusRequest) {
		if request.Opacity <= 0 || request.Size.X <= 0 || request.Size.Y <= 0 {
			return
		}
		strokeWidth := max(gtx.Dp(width), 1)
		inset := max(gtx.Dp(offset), 0) + (strokeWidth+1)/2
		rect := image.Rectangle{Max: request.Size}.Inset(inset)
		if rect.Empty() {
			return
		}
		cornerRadius := max(gtx.Dp(radius)-inset, 0)
		col := colorValue
		col.A = byte(float32(col.A)*request.Opacity + 0.5)
		stroke := clip.Stroke{
			Path:  clip.UniformRRect(rect, cornerRadius).Path(gtx.Ops),
			Width: float32(strokeWidth),
		}.Op().Push(gtx.Ops)
		paint.Fill(gtx.Ops, col)
		stroke.Pop()
	}
}
