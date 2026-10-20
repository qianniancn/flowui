// Package sidebarstyle owns the shared visual recipe for sidebar providers.
// It maps theme values to declarations and renderers; navigation behavior
// remains in widget/sidebar.
package sidebarstyle

import (
	"image"
	"image/color"
	"sync"

	"gioui.org/font"
	"gioui.org/layout"
	"gioui.org/op/clip"
	"gioui.org/op/paint"
	"gioui.org/unit"
	giowidget "gioui.org/widget"
	"gioui.org/widget/material"
	lucide "github.com/qianniancn/flowui-icons-lucide"
	"github.com/qianniancn/flowui/draw"
	"github.com/qianniancn/flowui/style"
	"github.com/qianniancn/flowui/theme"
)

// BuildSidebar returns the complete visual protocol for the active theme.
func BuildSidebar(active *theme.Theme, request theme.SidebarRequest) theme.SidebarVisual {
	if active == nil {
		panic(theme.ErrMissingVisualProvider)
	}
	tokens := active.Components.Sidebar
	sem := active.Semantic()
	metrics := theme.SidebarMetrics(tokens)
	metrics = request.Apply(metrics)
	override := theme.SidebarOverrideStyle(request)
	root := style.Cascade(style.State{}, override)
	foreground, background, border, borderWidth := sem.SurfaceForeground, sem.Surface, sem.Border, metrics.BorderWidth
	applyRootOverride(&foreground, &background, &border, &borderWidth, root)
	return theme.SidebarVisual{
		Metrics:         metrics,
		Foreground:      foreground,
		Background:      background,
		MutedForeground: sem.MutedForeground,
		Motion:          active.Motion,
		ExpansionEasing: theme.EaseOut,
		CollapseEasing:  theme.EaseInOut,
		Override:        override,
		Item: func(request theme.SidebarItemRequest) theme.SidebarItemVisual {
			return itemVisual(active, metrics, override, request)
		},
		Root:       rootRenderer(background, border, borderWidth),
		RenderItem: itemRenderer(metrics.FocusRingWidth),
		Text:       textRenderer(active),
		Section:    sectionRenderer(active, metrics),
		Switcher:   switcherRenderer,
	}
}

type sidebarSwitcherGlyph struct {
	sync.Mutex
	icon *giowidget.Icon
}

var sidebarSwitchers = [2]sidebarSwitcherGlyph{}

func switcherRenderer(gtx layout.Context, request theme.SidebarSwitcherRequest) layout.Dimensions {
	index := 0
	data := lucide.ChevronRight
	if request.Expanded {
		index = 1
		data = lucide.ChevronDown
	}
	glyph := &sidebarSwitchers[index]
	glyph.Lock()
	defer glyph.Unlock()
	if glyph.icon == nil {
		icon, err := giowidget.NewIcon(data)
		if err != nil {
			panic(err)
		}
		glyph.icon = icon
	}
	size := max(gtx.Dp(request.Size), 0)
	gtx.Constraints = layout.Exact(image.Pt(size, size))
	return glyph.icon.Layout(gtx, request.Foreground)
}

func applyRootOverride(foreground, background, border *color.NRGBA, borderWidth *unit.Dp, resolved style.ResolvedStyle) {
	if resolved.Paint != nil {
		if value, ok := solidColor(resolved.Paint.Background); ok {
			*background = value
		}
		if resolved.Paint.Border != nil {
			if value, ok := solidColor(resolved.Paint.Border.Color); ok {
				*border = value
			}
			if resolved.Paint.Border.Width != nil {
				*borderWidth = *resolved.Paint.Border.Width
			}
		}
	}
	if resolved.Text != nil {
		if value, ok := solidColor(resolved.Text.Color); ok {
			*foreground = value
		}
	}
}

func itemVisual(active *theme.Theme, metrics theme.SidebarMetrics, override style.Style, request theme.SidebarItemRequest) theme.SidebarItemVisual {
	sem := active.Semantic()
	result := theme.SidebarItemVisual{
		Foreground: sem.Foreground,
		Focus:      sem.Ring,
		Opacity:    1,
		Radius:     metrics.ItemRadius,
	}
	if request.Hovered {
		result.Background = sem.SurfaceTertiary
	}
	if request.Pressed {
		result.Background = sem.SurfacePressed
	}
	if request.Selected {
		tone := sem.Tone(theme.TonePrimary)
		result.Background = tone.Soft
		result.Foreground = tone.OnSoft
		if request.Hovered || request.Pressed {
			result.Background = tone.SoftHover
		}
	}
	if request.Disabled {
		result.Opacity = active.DisabledOpacityValue()
		result.Focus = color.NRGBA{}
	}
	applyItemOverride(&result, override, requestState(request))
	return result
}

func requestState(request theme.SidebarItemRequest) style.State {
	return style.State{Selected: request.Selected, Hovered: request.Hovered, Pressed: request.Pressed,
		Focused: request.Focused, FocusVisible: request.Focused, Disabled: request.Disabled}
}

func applyItemOverride(result *theme.SidebarItemVisual, declaration style.Style, state style.State) {
	applyResolvedItem(result, style.CascadePart(state, style.PartItem, declaration))
}

func applyResolvedItem(result *theme.SidebarItemVisual, resolved style.ResolvedStyle) {
	if resolved.Paint != nil {
		if value, ok := solidColor(resolved.Paint.Background); ok {
			result.Background = value
		}
		if resolved.Paint.Radius != nil {
			result.Radius = *resolved.Paint.Radius
		}
		if resolved.Paint.Border != nil {
			if value, ok := solidColor(resolved.Paint.Border.Color); ok {
				result.Border = value
			}
			if resolved.Paint.Border.Width != nil {
				result.BorderWidth = *resolved.Paint.Border.Width
			}
		}
		if resolved.Paint.Outline != nil {
			if value, ok := solidColor(resolved.Paint.Outline.Color); ok {
				result.Outline = value
			}
			result.OutlineWidth = resolved.Paint.Outline.Width
		}
		if resolved.Paint.Opacity != nil {
			result.Opacity = *resolved.Paint.Opacity
		}
	}
	if resolved.Text != nil {
		if value, ok := solidColor(resolved.Text.Color); ok {
			result.Foreground = value
		}
	}
}

func solidColor(source style.PaintSource) (color.NRGBA, bool) {
	switch value := source.(type) {
	case style.SolidColor:
		return value.Color, true
	case *style.SolidColor:
		if value != nil {
			return value.Color, true
		}
	}
	return color.NRGBA{}, false
}

func rootRenderer(background, border color.NRGBA, borderWidth unit.Dp) theme.SidebarRootRenderer {
	return func(gtx layout.Context, size image.Point) {
		if size.X <= 0 || size.Y <= 0 {
			return
		}
		paint.FillShape(gtx.Ops, background, clip.Rect{Max: size}.Op())
		width := max(gtx.Dp(borderWidth), 0)
		if width == 0 {
			return
		}
		borderRect := image.Rect(max(size.X-width, 0), 0, size.X, size.Y)
		paint.FillShape(gtx.Ops, border, clip.Rect(borderRect).Op())
	}
}

func itemRenderer(focusRingWidth unit.Dp) theme.SidebarItemRenderer {
	return func(gtx layout.Context, size image.Point, visual theme.SidebarItemVisual, focus float32) {
		rect := image.Rectangle{Max: size}
		radius := min(max(gtx.Dp(visual.Radius), 0), min(size.X, size.Y)/2)
		if visual.Background.A != 0 {
			paint.FillShape(gtx.Ops, visual.Background, clip.UniformRRect(rect, radius).Op(gtx.Ops))
		}
		if visual.BorderWidth > 0 {
			draw.RoundedBorder(gtx, rect, radius, max(gtx.Dp(visual.BorderWidth), 1), visual.Border)
		}
		if focus <= 0 {
			return
		}
		focusWidth := focusRingWidth
		if visual.OutlineWidth > 0 {
			focusWidth = visual.OutlineWidth
			visual.Focus = visual.Outline
		}
		width := max(gtx.Dp(focusWidth), 0)
		if width == 0 {
			return
		}
		inset := max(width/2+1, 1)
		focusRect := rect.Inset(inset)
		if focusRect.Empty() {
			return
		}
		focusColor := visual.Focus
		focusColor.A = byte(float32(focusColor.A)*focus + .5)
		draw.RoundedInsetStroke(gtx, rect, max(radius-inset, 0), width, inset, focusColor)
	}
}

func textRenderer(active *theme.Theme) theme.SidebarTextRenderer {
	return func(gtx layout.Context, request theme.SidebarTextRequest) layout.Dimensions {
		value := request.Text
		if request.Initial {
			for _, runeValue := range value {
				value = string(runeValue)
				break
			}
		}
		label := material.Label(theme.MaterialOf(active), request.Size, value)
		label.Color = request.Color
		label.Font.Weight = font.Medium
		label.MaxLines = 1
		return label.Layout(gtx)
	}
}

func sectionRenderer(active *theme.Theme, metrics theme.SidebarMetrics) theme.SidebarSectionRenderer {
	return func(gtx layout.Context, request theme.SidebarSectionRequest) layout.Dimensions {
		height := min(gtx.Dp(metrics.SectionHeight), gtx.Constraints.Max.Y)
		sectionGtx := gtx
		sectionGtx.Constraints.Min.Y = height
		sectionGtx.Constraints.Max.Y = height
		if request.Collapsed {
			return layout.Center.Layout(sectionGtx, func(gtx layout.Context) layout.Dimensions {
				width := max(gtx.Constraints.Max.X-gtx.Dp(metrics.SectionSeparatorInset)*2, 0)
				lineHeight := max(gtx.Dp(metrics.BorderWidth), 1)
				rect := image.Rect(0, 0, width, lineHeight)
				paint.FillShape(gtx.Ops, active.Semantic().Border, clip.Rect(rect).Op())
				return layout.Dimensions{Size: rect.Size()}
			})
		}
		return layout.W.Layout(sectionGtx, func(gtx layout.Context) layout.Dimensions {
			label := material.Label(theme.MaterialOf(active), metrics.SectionTextSize, request.Title)
			label.Color = active.Semantic().MutedForeground
			label.Font.Weight = font.Medium
			label.MaxLines = 1
			return label.Layout(gtx)
		})
	}
}
