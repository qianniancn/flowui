// Package listboxstyle contains the visual recipe for listbox providers.
// It deliberately depends on the neutral theme protocol, never on widget
// behavior or selection state.
package listboxstyle

import (
	"image"
	"image/color"
	"time"

	"gioui.org/f32"
	"gioui.org/font"
	"gioui.org/layout"
	"gioui.org/op/clip"
	"gioui.org/op/paint"
	"gioui.org/unit"
	"gioui.org/widget/material"
	"github.com/qianniancn/flowui/draw"
	"github.com/qianniancn/flowui/style"
	"github.com/qianniancn/flowui/theme"
	"github.com/qianniancn/flowui/theme/internal/textrender"
)

const (
	itemDefault uint8 = iota
	itemDanger
)

const (
	defaultColorDuration     = 100 * time.Millisecond
	defaultFocusDuration     = 100 * time.Millisecond
	defaultPressInDuration   = 80 * time.Millisecond
	defaultPressOutDuration  = 140 * time.Millisecond
	defaultSelectionDuration = 200 * time.Millisecond
)

// BuildListBox returns the complete ListBox visual protocol for the active
// theme. All fields are derived from the active theme on this call so copied
// themes and runtime replacement cannot retain stale visual values.
func BuildListBox(active *theme.Theme, request theme.ListBoxRequest) theme.ListBoxVisual {
	if active == nil {
		panic(theme.ErrMissingVisualProvider)
	}
	tokens := active.Components.ListBox
	return theme.ListBoxVisual{
		Metrics:  metrics(active),
		Colors:   colors(active),
		Motion:   active.Motion,
		Override: theme.ListBoxOverrideStyle(request),
		Text:     textRenderer(active, tokens),
		Item: func(item theme.ListBoxItemRequest) theme.ListBoxItemVisual {
			return itemVisual(active, item)
		},
		Indicator: indicatorRenderer(tokens.ItemIndicatorInset, tokens.ItemIndicatorStroke),
	}
}

func metrics(active *theme.Theme) theme.ListBoxMetrics {
	tokens := active.Components.ListBox
	return theme.ListBoxMetrics{
		Padding:               tokens.Padding,
		Gap:                   tokens.Gap,
		MaxHeight:             tokens.MaxHeight,
		SectionHeaderPaddingX: tokens.SectionHeaderPaddingX,
		SectionHeaderPaddingY: tokens.SectionHeaderPaddingY,
		ItemMinHeight:         tokens.ItemMinHeight,
		ItemContentGap:        tokens.ItemContentGap,
		ItemIndicatorSize:     tokens.ItemIndicatorSize,
		SelectionDuration:     defaultSelectionDuration,
	}
}

func colors(active *theme.Theme) theme.ListBoxColors {
	sem := active.Semantic()
	return theme.ListBoxColors{
		MutedForeground: sem.MutedForeground,
	}
}

func textRenderer(active *theme.Theme, tokens theme.ListBoxTheme) theme.ListBoxTextRenderer {
	return func(gtx layout.Context, request theme.ListBoxTextRequest) layout.Dimensions {
		size := tokens.ItemTextSize
		weight := font.Normal
		switch request.Role {
		case theme.ListBoxTextSectionHeader:
			size = tokens.SectionHeaderTextSize
		case theme.ListBoxTextDescription:
			size = tokens.ItemDescriptionSize
		case theme.ListBoxTextItem:
			weight = font.Medium
		case theme.ListBoxTextEmpty:
		default:
			panic(theme.ErrIncompleteVisualProvider)
		}
		label := material.Label(theme.MaterialOf(active), size, request.Text)
		label.Color = request.Color
		label.SelectionColor = active.Semantic().Selection
		label.Font.Weight = weight
		return label.Layout(gtx)
	}
}

func itemVisual(active *theme.Theme, request theme.ListBoxItemRequest) theme.ListBoxItemVisual {
	sem := active.Semantic()
	tokens := active.Components.ListBox
	pressed := pressedScale(tokens.PressedScale)
	foreground := sem.Foreground
	indicator := sem.Foreground
	if request.Variant == itemDanger {
		danger := sem.Tone(theme.ToneDanger).Fill
		foreground = danger
		indicator = danger
	}
	if request.State.Disabled {
		foreground = active.DisabledColor(foreground)
		indicator = active.DisabledColor(indicator)
	}
	item := style.Style{}.
		MinHeight(tokens.ItemMinHeight).
		PaddingX(tokens.ItemPaddingX).
		PaddingY(tokens.ItemPaddingY).
		Radius(tokens.ItemRadius).
		Background(style.RGBA(0x00000000)).
		TextColor(style.Solid(foreground)).
		BorderWidth(tokens.FocusRingWidth).
		BorderColor(style.WithAlpha(sem.Ring, 0)).
		Opacity(1).
		Scale(1).
		Transition(style.PropBackgroundColor, defaultColorDuration).
		Transition(style.PropBorderColor, defaultFocusDuration).
		Transition(style.PropTransform, defaultPressOutDuration)
	defaults := style.Style{}.Part(style.PartItem, item).
		Rules(
			style.Rule{
				When: style.Hovered,
				Set: style.Style{}.Part(style.PartItem,
					style.Style{}.Background(style.Solid(sem.SurfaceRaised))),
			},
			style.Rule{
				When: style.Pressed,
				Set: style.Style{}.Part(style.PartItem,
					style.Style{}.
						Background(style.Solid(sem.SurfacePressed)).
						Scale(pressed).
						Transition(style.PropTransform, defaultPressInDuration)),
			},
			style.Rule{
				When: style.FocusVisible,
				Set: style.Style{}.Part(style.PartItem,
					style.Style{}.BorderColor(style.Solid(sem.Ring))),
			},
			style.Rule{
				When: style.Disabled,
				Set: style.Style{}.Part(style.PartItem,
					style.Style{}.Opacity(active.DisabledOpacityValue())),
			},
		)
	variant := style.Style{}
	if request.Variant == itemDanger {
		danger := sem.Tone(theme.ToneDanger).Fill
		variant = style.Style{}.Part(style.PartItem, style.Style{}.
			TextColor(style.Solid(foreground))).
			Rules(style.Rule{
				When: style.FocusVisible,
				Set: style.Style{}.Part(style.PartItem,
					style.Style{}.BorderColor(style.WithAlpha(danger, float32(sem.Ring.A)/255))),
			})
	}
	description := sem.MutedForeground
	if request.State.Disabled {
		description = active.DisabledColor(description)
	}
	return theme.ListBoxItemVisual{
		Defaults: defaults,
		Variant:  variant,
		Colors: func(resolved style.ResolvedStyle) theme.ListBoxItemColors {
			if foregroundValue, ok := textrender.ColorValue(resolved); ok {
				return theme.ListBoxItemColors{
					Foreground:  foregroundValue,
					Description: foregroundValue,
					Indicator:   foregroundValue,
				}
			}
			return theme.ListBoxItemColors{
				Foreground:  foreground,
				Description: description,
				Indicator:   indicator,
			}
		},
	}
}

func pressedScale(value float32) float32 {
	if value <= 0 || value > 1 {
		return 0.98
	}
	return value
}

func indicatorRenderer(inset, stroke unit.Dp) theme.ListBoxIndicatorRenderer {
	return func(gtx layout.Context, size image.Point, progress float32, foreground color.NRGBA) {
		progress = min(max(progress, 0), 1)
		if progress == 0 {
			return
		}
		rect := image.Rectangle{Max: size}.Inset(max(gtx.Dp(inset), 0))
		if rect.Empty() {
			return
		}
		x, y := float32(rect.Min.X), float32(rect.Min.Y)
		width, height := float32(rect.Dx()), float32(rect.Dy())
		points := [3]f32.Point{
			f32.Pt(x+width*0.05, y+height*0.56),
			f32.Pt(x+width*0.40, y+height*0.86),
			f32.Pt(x+width*0.95, y+height*0.14),
		}
		path := draw.CheckPath(gtx.Ops, points, progress)
		foreground.A = byte(float32(foreground.A)*progress + 0.5)
		strokeOp := clip.Stroke{
			Path:  path,
			Width: max(draw.DpFloat(gtx, stroke), 1),
		}.Op().Push(gtx.Ops)
		paint.Fill(gtx.Ops, foreground)
		strokeOp.Pop()
	}
}
