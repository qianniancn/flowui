// Package collapsiblestyle owns the visual recipe shared by collapsible
// providers. Expansion state, focus, and keyboard behavior remain in the
// widget package.
package collapsiblestyle

import (
	"image"
	"math"
	"sync"

	"gioui.org/f32"
	"gioui.org/font"
	"gioui.org/layout"
	"gioui.org/op"
	"gioui.org/widget"
	lucide "github.com/qianniancn/flowui-icons-lucide"
	"github.com/qianniancn/flowui/style"
	"github.com/qianniancn/flowui/theme"
)

// BuildCollapsible builds the provider-owned collapsible declaration and
// indicator renderer.
func BuildCollapsible(active *theme.Theme, request theme.CollapsibleRequest) theme.CollapsibleVisual {
	if active == nil {
		panic(theme.ErrMissingVisualProvider)
	}

	tokens := active.Components.Collapsible
	sem := active.Semantic()
	labelColor := sem.Foreground
	disabledLabelColor := active.DisabledColor(labelColor)
	return theme.CollapsibleVisual{
		Label: style.Style{}.
			TextColor(style.Solid(labelColor)).
			FontSize(active.Typography.ControlSize).
			FontWeight(int(font.Medium)).
			On(style.Disabled, style.Style{}.TextColor(style.Solid(disabledLabelColor))),
		Trigger: theme.CollapsibleTriggerOverrideStyle(request),
		Metrics: theme.CollapsibleMetrics{
			BodyPadding:       tokens.BodyPadding,
			ContentGap:        tokens.ContentGap,
			ContentDuration:   tokens.ContentDuration,
			IndicatorDuration: tokens.IndicatorDuration,
		},
		Motion:    active.Motion,
		Indicator: indicatorRenderer(active),
		Override:  theme.CollapsibleOverrideStyle(request),
	}
}

var indicatorIcon struct {
	sync.Once
	value *widget.Icon
}

func indicatorRenderer(active *theme.Theme) theme.CollapsibleIndicatorRenderer {
	return func(gtx layout.Context, request theme.CollapsibleIndicatorRequest) layout.Dimensions {
		diameter := max(gtx.Dp(active.Components.Collapsible.IndicatorSize), gtx.Dp(active.Components.Collapsible.IndicatorStroke*12))
		diameter = min(diameter, min(gtx.Constraints.Max.X, gtx.Constraints.Max.Y))
		if diameter <= 0 {
			return layout.Dimensions{}
		}

		size := image.Pt(diameter, diameter)
		icon := disclosureIcon()
		if icon == nil {
			return layout.Dimensions{Size: size}
		}
		color := active.Semantic().MutedForeground
		if request.Disabled {
			color = active.DisabledColor(color)
		}
		// The reported size is the icon box, so the caller's flex already places
		// it. Centering it again inside the leftover constraints moved the
		// chevron towards the middle of the trigger instead of its end.
		center := f32.Pt(float32(size.X)/2, float32(size.Y)/2)
		rotation := op.Affine(f32.AffineId().Rotate(center, -request.Progress*float32(math.Pi))).Push(gtx.Ops)
		iconGtx := gtx
		iconGtx.Constraints = layout.Exact(size)
		icon.Layout(iconGtx, color)
		rotation.Pop()
		return layout.Dimensions{Size: size}
	}
}

func disclosureIcon() *widget.Icon {
	indicatorIcon.Do(func() {
		value, err := widget.NewIcon(lucide.ChevronDown)
		if err == nil {
			indicatorIcon.value = value
		}
	})
	return indicatorIcon.value
}
