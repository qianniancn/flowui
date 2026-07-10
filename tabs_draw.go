package flowui

import (
	"image"
	"image/color"

	"gioui.org/f32"
	"gioui.org/layout"
	"gioui.org/op/clip"
	"gioui.org/op/paint"
)

type tabsScrollShadowSpec struct {
	bounds image.Rectangle
	stop1  f32.Point
	color1 color.NRGBA
	stop2  f32.Point
	color2 color.NRGBA
}

func drawTabsList(gtx layout.Context, theme *Theme, size image.Point, orientation TabsOrientation, variant TabsVariant, style tabsListStyle) {
	if size.X <= 0 || size.Y <= 0 {
		return
	}
	rect := image.Rectangle{Max: size}
	if variant == TabsSecondary {
		width := max(gtx.Dp(theme.Components.Tabs.SeparatorWidth), 1)
		if orientation == TabsVertical {
			paint.FillShape(gtx.Ops, style.border, clip.Rect(image.Rect(0, 0, width, size.Y)).Op())
		} else {
			paint.FillShape(gtx.Ops, style.border, clip.Rect(image.Rect(0, size.Y-width, size.X, size.Y)).Op())
		}
		return
	}
	radius := min(max(gtx.Dp(theme.Components.Tabs.ListRadius), 1), min(size.X, size.Y)/2)
	paint.FillShape(gtx.Ops, style.background, clip.UniformRRect(rect, radius).Op(gtx.Ops))
}

func drawTabIndicator(gtx layout.Context, theme *Theme, rect image.Rectangle, orientation TabsOrientation, variant TabsVariant, col color.NRGBA) {
	if rect.Empty() || col.A == 0 {
		return
	}
	if variant == TabsSecondary {
		width := max(gtx.Dp(theme.Components.Tabs.IndicatorLineWidth), 1)
		if orientation == TabsVertical {
			paint.FillShape(gtx.Ops, col, clip.Rect(image.Rect(rect.Min.X, rect.Min.Y, rect.Min.X+width, rect.Max.Y)).Op())
		} else {
			paint.FillShape(gtx.Ops, col, clip.Rect(image.Rect(rect.Min.X, rect.Max.Y-width, rect.Max.X, rect.Max.Y)).Op())
		}
		return
	}
	radius := min(max(gtx.Dp(theme.Components.Tabs.IndicatorRadius), 1), min(rect.Dx(), rect.Dy())/2)
	if theme.Palette.SurfaceShadow.A != 0 {
		shapeRadius := theme.Components.Tabs.IndicatorRadius
		DrawShadow(gtx, rect, RoundedShadowCorners(shapeRadius, shapeRadius, shapeRadius, shapeRadius), SurfaceShadow(theme.Palette.SurfaceShadow))
	}
	paint.FillShape(gtx.Ops, col, clip.UniformRRect(rect, radius).Op(gtx.Ops))
}

func drawTabFocus(gtx layout.Context, theme *Theme, size image.Point, variant TabsVariant, opacity float32, col color.NRGBA) {
	if opacity <= 0 || col.A == 0 || size.X <= 0 || size.Y <= 0 {
		return
	}
	width := max(gtx.Dp(theme.Components.Tabs.FocusRingWidth), 1)
	rect := image.Rectangle{Max: size}.Inset(max(width/2, 1))
	radius := 0
	if variant != TabsSecondary {
		radius = min(max(gtx.Dp(theme.Components.Tabs.IndicatorRadius), 1), min(rect.Dx(), rect.Dy())/2)
	}
	col.A = byte(float32(col.A)*opacity + 0.5)
	stroke := clip.Stroke{
		Path:  clip.UniformRRect(rect, radius).Path(gtx.Ops),
		Width: float32(width),
	}.Op().Push(gtx.Ops)
	paint.Fill(gtx.Ops, col)
	stroke.Pop()
}

func drawTabSeparator(gtx layout.Context, theme *Theme, size image.Point, orientation TabsOrientation, col color.NRGBA) {
	if col.A == 0 || size.X <= 0 || size.Y <= 0 {
		return
	}
	width := max(gtx.Dp(theme.Components.Tabs.SeparatorWidth), 1)
	if orientation == TabsVertical {
		inset := max(size.X/20, 1)
		paint.FillShape(gtx.Ops, col, clip.Rect(image.Rect(inset, 0, max(size.X-inset, inset), width)).Op())
		return
	}
	inset := size.Y / 4
	paint.FillShape(gtx.Ops, col, clip.Rect(image.Rect(0, inset, width, max(size.Y-inset, inset))).Op())
}

func drawTabsScrollShadows(gtx layout.Context, theme *Theme, size image.Point, orientation TabsOrientation, previous, next bool, background color.NRGBA) {
	length := gtx.Dp(theme.Components.Tabs.ScrollShadowSize)
	for _, edge := range []struct {
		visible   bool
		direction int
	}{
		{visible: previous, direction: -1},
		{visible: next, direction: 1},
	} {
		if !edge.visible {
			continue
		}
		spec, ok := tabsScrollShadowFor(size, orientation, length, edge.direction, background)
		if !ok {
			continue
		}
		stack := clip.Rect(spec.bounds).Push(gtx.Ops)
		paint.LinearGradientOp{
			Stop1:  spec.stop1,
			Color1: spec.color1,
			Stop2:  spec.stop2,
			Color2: spec.color2,
		}.Add(gtx.Ops)
		paint.PaintOp{}.Add(gtx.Ops)
		stack.Pop()
	}
}

func tabsScrollShadowFor(size image.Point, orientation TabsOrientation, length, direction int, background color.NRGBA) (tabsScrollShadowSpec, bool) {
	mainSize := size.X
	if orientation == TabsVertical {
		mainSize = size.Y
	}
	length = min(max(length, 0), mainSize/2)
	if length == 0 || direction == 0 || size.X <= 0 || size.Y <= 0 || background.A == 0 {
		return tabsScrollShadowSpec{}, false
	}
	transparent := background
	transparent.A = 0
	if orientation == TabsVertical {
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

func drawTabsScrollShadowBorders(gtx layout.Context, theme *Theme, size image.Point, orientation TabsOrientation, previous, next bool, border color.NRGBA) {
	if border.A == 0 {
		return
	}
	length := gtx.Dp(theme.Components.Tabs.ScrollShadowSize)
	width := max(gtx.Dp(theme.Components.Tabs.SeparatorWidth), 1)
	for _, edge := range []struct {
		visible   bool
		direction int
	}{
		{visible: previous, direction: -1},
		{visible: next, direction: 1},
	} {
		if !edge.visible {
			continue
		}
		spec, ok := tabsScrollShadowFor(size, orientation, length, edge.direction, border)
		if !ok {
			continue
		}
		bounds := spec.bounds
		if orientation == TabsVertical {
			bounds.Max.X = min(bounds.Min.X+width, bounds.Max.X)
		} else {
			bounds.Min.Y = max(bounds.Max.Y-width, bounds.Min.Y)
		}
		paint.FillShape(gtx.Ops, border, clip.Rect(bounds).Op())
	}
}

func drawTabsChevron(gtx layout.Context, theme *Theme, size image.Point, orientation TabsOrientation, direction int, col color.NRGBA) {
	center := f32.Pt(float32(size.X)/2, float32(size.Y)/2)
	half := float32(gtx.Dp(theme.Components.Tabs.ScrollChevronSize)) / 2
	if half < 2 {
		half = 2
	}
	var path clip.Path
	path.Begin(gtx.Ops)
	if orientation == TabsVertical {
		path.MoveTo(f32.Pt(center.X-half, center.Y-float32(direction)*half/2))
		path.LineTo(f32.Pt(center.X, center.Y+float32(direction)*half/2))
		path.LineTo(f32.Pt(center.X+half, center.Y-float32(direction)*half/2))
	} else {
		path.MoveTo(f32.Pt(center.X-float32(direction)*half/2, center.Y-half))
		path.LineTo(f32.Pt(center.X+float32(direction)*half/2, center.Y))
		path.LineTo(f32.Pt(center.X-float32(direction)*half/2, center.Y+half))
	}
	stroke := clip.Stroke{
		Path:  path.End(),
		Width: max(dpFloat(gtx, theme.Components.Tabs.ScrollChevronStroke), 1),
	}.Op().Push(gtx.Ops)
	paint.Fill(gtx.Ops, col)
	stroke.Pop()
}
