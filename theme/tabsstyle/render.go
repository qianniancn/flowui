package tabsstyle

import (
	"image"
	"image/color"

	"gioui.org/f32"
	"gioui.org/layout"
	"gioui.org/op"
	"gioui.org/op/clip"
	"gioui.org/op/paint"
	"gioui.org/unit"
	lucide "github.com/qianniancn/flowui-icons-lucide"
	"github.com/qianniancn/flowui/draw"
	"github.com/qianniancn/flowui/theme"
	"github.com/qianniancn/flowui/widget/icon"
)

type tabsRenderers struct {
	indicator theme.TabsIndicatorRenderer
	focus     theme.TabsFocusRenderer
	separator theme.TabsSeparatorRenderer
	list      theme.TabsListRenderer
	scroll    theme.TabsScrollRenderer
	chevron   theme.TabsChevronRenderer
	close     theme.TabsCloseRenderer
	add       theme.TabsAddRenderer
	overflow  theme.TabsOverflowRenderer
}

// tabsRendererMetrics is intentionally private to the recipe. These values
// describe how the theme paints chrome; they are not widget layout protocol.
type tabsRendererMetrics struct {
	listRadius          unit.Dp
	indicatorRadius     unit.Dp
	indicatorLineWidth  unit.Dp
	focusRingWidth      unit.Dp
	separatorWidth      unit.Dp
	scrollShadowSize    unit.Dp
	scrollChevronSize   unit.Dp
	scrollChevronStroke unit.Dp
}

func newTabsRenderers(tokens theme.TabsTheme) tabsRenderers {
	metrics := tabsRendererMetrics{
		listRadius: tokens.ListRadius, indicatorRadius: tokens.IndicatorRadius,
		indicatorLineWidth: tokens.IndicatorLineWidth, focusRingWidth: tokens.FocusRingWidth,
		separatorWidth: tokens.SeparatorWidth, scrollShadowSize: tokens.ScrollShadowSize,
		scrollChevronSize: tokens.ScrollChevronSize, scrollChevronStroke: tokens.ScrollChevronStroke,
	}
	return tabsRenderers{
		indicator: tabsIndicatorRenderer(metrics), focus: tabsFocusRenderer(metrics),
		separator: tabsSeparatorRenderer(metrics), list: tabsListRenderer(metrics),
		scroll: tabsScrollRenderer(metrics), chevron: tabsChevronRenderer(metrics),
		close: tabsCloseRenderer(), add: tabsAddRenderer(), overflow: tabsOverflowRenderer(),
	}
}

// tabsIndicatorRenderer owns the selected-tab shape. The widget computes the
// animated bounds and supplies only neutral axis/variant facts and color.
func tabsIndicatorRenderer(metrics tabsRendererMetrics) theme.TabsIndicatorRenderer {
	return func(gtx layout.Context, request theme.TabsIndicatorRenderRequest) {
		rect := request.Bounds
		if rect.Empty() || request.Color.A == 0 {
			return
		}
		if request.Variant == 1 {
			width := max(gtx.Dp(metrics.indicatorLineWidth), 1)
			if request.Orientation == 1 {
				paint.FillShape(gtx.Ops, request.Color,
					clip.Rect(image.Rect(rect.Min.X, rect.Min.Y, rect.Min.X+width, rect.Max.Y)).Op())
			} else {
				paint.FillShape(gtx.Ops, request.Color,
					clip.Rect(image.Rect(rect.Min.X, rect.Max.Y-width, rect.Max.X, rect.Max.Y)).Op())
			}
			return
		}
		radius := min(max(gtx.Dp(metrics.indicatorRadius), 1), min(rect.Dx(), rect.Dy())/2)
		paint.FillShape(gtx.Ops, request.Color, clip.UniformRRect(rect, radius).Op(gtx.Ops))
	}
}

func tabsFocusRenderer(metrics tabsRendererMetrics) theme.TabsFocusRenderer {
	return func(gtx layout.Context, request theme.TabsFocusRenderRequest) {
		if request.Opacity <= 0 || request.Color.A == 0 || request.Size.X <= 0 || request.Size.Y <= 0 {
			return
		}
		width := max(gtx.Dp(metrics.focusRingWidth), 1)
		rect := image.Rectangle{Max: request.Size}.Inset(max(width/2, 1))
		radius := 0
		if request.Variant != 1 {
			radius = min(max(gtx.Dp(metrics.indicatorRadius), 1), min(rect.Dx(), rect.Dy())/2)
		}
		color := request.Color
		color.A = byte(float32(color.A)*request.Opacity + 0.5)
		draw.RoundedInsetStroke(gtx, image.Rectangle{Max: request.Size}, radius, width, max(width/2, 1), color)
	}
}

func tabsSeparatorRenderer(metrics tabsRendererMetrics) theme.TabsSeparatorRenderer {
	return func(gtx layout.Context, request theme.TabsSeparatorRenderRequest) {
		if request.Color.A == 0 || request.Size.X <= 0 || request.Size.Y <= 0 {
			return
		}
		width := max(gtx.Dp(metrics.separatorWidth), 1)
		if request.Orientation == 1 {
			inset := max(request.Size.X/20, 1)
			paint.FillShape(gtx.Ops, request.Color,
				clip.Rect(image.Rect(inset, 0, max(request.Size.X-inset, inset), width)).Op())
			return
		}
		inset := request.Size.Y / 4
		paint.FillShape(gtx.Ops, request.Color,
			clip.Rect(image.Rect(0, inset, width, max(request.Size.Y-inset, inset))).Op())
	}
}

func tabsListRenderer(metrics tabsRendererMetrics) theme.TabsListRenderer {
	return func(gtx layout.Context, request theme.TabsListRenderRequest) {
		if request.Size.X <= 0 || request.Size.Y <= 0 {
			return
		}
		rect := image.Rectangle{Max: request.Size}
		if request.Variant == 1 {
			width := max(gtx.Dp(metrics.separatorWidth), 1)
			if request.Orientation == 1 {
				paint.FillShape(gtx.Ops, request.Border, clip.Rect(image.Rect(0, 0, width, request.Size.Y)).Op())
				return
			}
			paint.FillShape(gtx.Ops, request.Border,
				clip.Rect(image.Rect(0, request.Size.Y-width, request.Size.X, request.Size.Y)).Op())
			return
		}
		radius := min(max(gtx.Dp(metrics.listRadius), 1), min(request.Size.X, request.Size.Y)/2)
		paint.FillShape(gtx.Ops, request.Background, clip.UniformRRect(rect, radius).Op(gtx.Ops))
	}
}

type tabsScrollShadowSpec struct {
	bounds image.Rectangle
	stop1  f32.Point
	color1 color.NRGBA
	stop2  f32.Point
	color2 color.NRGBA
}

func tabsScrollShadowFor(size image.Point, orientation uint8, length, direction int, background color.NRGBA) (tabsScrollShadowSpec, bool) {
	mainSize := size.X
	if orientation == 1 {
		mainSize = size.Y
	}
	length = min(max(length, 0), mainSize/2)
	if length == 0 || direction == 0 || size.X <= 0 || size.Y <= 0 || background.A == 0 {
		return tabsScrollShadowSpec{}, false
	}
	transparent := background
	transparent.A = 0
	if orientation == 1 {
		if direction < 0 {
			return tabsScrollShadowSpec{
				bounds: image.Rect(0, 0, size.X, length),
				stop1:  f32.Pt(0, 0), color1: background,
				stop2: f32.Pt(0, float32(length)), color2: transparent,
			}, true
		}
		return tabsScrollShadowSpec{
			bounds: image.Rect(0, size.Y-length, size.X, size.Y),
			stop1:  f32.Pt(0, float32(size.Y-length)), color1: transparent,
			stop2: f32.Pt(0, float32(size.Y)), color2: background,
		}, true
	}
	if direction < 0 {
		return tabsScrollShadowSpec{
			bounds: image.Rect(0, 0, length, size.Y),
			stop1:  f32.Pt(0, 0), color1: background,
			stop2: f32.Pt(float32(length), 0), color2: transparent,
		}, true
	}
	return tabsScrollShadowSpec{
		bounds: image.Rect(size.X-length, 0, size.X, size.Y),
		stop1:  f32.Pt(float32(size.X-length), 0), color1: transparent,
		stop2: f32.Pt(float32(size.X), 0), color2: background,
	}, true
}

func tabsScrollRenderer(metrics tabsRendererMetrics) theme.TabsScrollRenderer {
	return func(gtx layout.Context, request theme.TabsScrollRenderRequest) {
		if request.Size.X <= 0 || request.Size.Y <= 0 {
			return
		}
		length := gtx.Dp(metrics.scrollShadowSize)
		var shadowClip clip.Stack
		hasShadowClip := false
		if request.Variant != 1 {
			radius := min(max(gtx.Dp(metrics.listRadius), 1), min(request.Size.X, request.Size.Y)/2)
			shadowClip = clip.UniformRRect(image.Rectangle{Max: request.Size}, radius).Push(gtx.Ops)
			hasShadowClip = true
		}
		for _, edge := range []struct {
			visible   bool
			direction int
		}{
			{visible: request.Previous, direction: -1},
			{visible: request.Next, direction: 1},
		} {
			if !edge.visible {
				continue
			}
			spec, ok := tabsScrollShadowFor(request.Size, request.Orientation, length, edge.direction, request.Background)
			if !ok {
				continue
			}
			stack := clip.Rect(spec.bounds).Push(gtx.Ops)
			draw.PaintLinearGradient(gtx, spec.stop1, spec.color1, spec.stop2, spec.color2)
			stack.Pop()
		}
		if hasShadowClip {
			shadowClip.Pop()
		}
		if request.Variant != 1 || request.Border.A == 0 {
			return
		}
		width := max(gtx.Dp(metrics.separatorWidth), 1)
		for _, edge := range []struct {
			visible   bool
			direction int
		}{
			{visible: request.Previous, direction: -1},
			{visible: request.Next, direction: 1},
		} {
			if !edge.visible {
				continue
			}
			spec, ok := tabsScrollShadowFor(request.Size, request.Orientation, length, edge.direction, request.Border)
			if !ok {
				continue
			}
			bounds := spec.bounds
			if request.Orientation == 1 {
				bounds.Max.X = min(bounds.Min.X+width, bounds.Max.X)
			} else {
				bounds.Min.Y = max(bounds.Max.Y-width, bounds.Min.Y)
			}
			paint.FillShape(gtx.Ops, request.Border, clip.Rect(bounds).Op())
		}
	}
}

func tabsChevronRenderer(metrics tabsRendererMetrics) theme.TabsChevronRenderer {
	return func(gtx layout.Context, request theme.TabsChevronRenderRequest) {
		data := lucide.ChevronRight
		if request.Orientation == 1 {
			data = lucide.ChevronDown
			if request.Direction < 0 {
				data = lucide.ChevronUp
			}
		} else if request.Direction < 0 {
			data = lucide.ChevronLeft
		}
		diameter := max(
			gtx.Dp(metrics.scrollChevronSize*2),
			icon.LucideSizeForStroke(gtx, metrics.scrollChevronStroke),
		)
		diameter = min(diameter, min(request.Size.X, request.Size.Y))
		if diameter <= 0 {
			return
		}
		offset := op.Offset(image.Pt((request.Size.X-diameter)/2, (request.Size.Y-diameter)/2)).Push(gtx.Ops)
		iconGtx := gtx
		iconGtx.Constraints = layout.Exact(image.Pt(diameter, diameter))
		icon.Layout(data, iconGtx, request.Color)
		offset.Pop()
	}
}

func tabsCloseRenderer() theme.TabsCloseRenderer {
	return func(gtx layout.Context, request theme.TabsCloseRenderRequest) {
		diameter := min(request.Size.X, request.Size.Y)
		if diameter <= 0 {
			return
		}
		offset := op.Offset(image.Pt((request.Size.X-diameter)/2, (request.Size.Y-diameter)/2)).Push(gtx.Ops)
		iconGtx := gtx
		iconGtx.Constraints = layout.Exact(image.Pt(diameter, diameter))
		icon.Layout(lucide.X, iconGtx, request.Color)
		offset.Pop()
	}
}

func tabsAddRenderer() theme.TabsAddRenderer {
	return func(gtx layout.Context, request theme.TabsAddRenderRequest) {
		if request.Size.X <= 0 || request.Size.Y <= 0 {
			return
		}
		diameter := min(request.Size.X, request.Size.Y)
		offset := op.Offset(image.Pt((request.Size.X-diameter)/2, (request.Size.Y-diameter)/2)).Push(gtx.Ops)
		iconGtx := gtx
		iconGtx.Constraints = layout.Exact(image.Pt(diameter, diameter))
		icon.Layout(lucide.Plus, iconGtx, request.Color)
		offset.Pop()
	}
}

func tabsOverflowRenderer() theme.TabsOverflowRenderer {
	return func(gtx layout.Context, request theme.TabsOverflowRenderRequest) {
		if request.Size.X <= 0 || request.Size.Y <= 0 {
			return
		}
		diameter := min(request.Size.X, request.Size.Y)
		offset := op.Offset(image.Pt((request.Size.X-diameter)/2, (request.Size.Y-diameter)/2)).Push(gtx.Ops)
		iconGtx := gtx
		iconGtx.Constraints = layout.Exact(image.Pt(diameter, diameter))
		icon.Layout(lucide.Ellipsis, iconGtx, request.Color)
		offset.Pop()
	}
}
