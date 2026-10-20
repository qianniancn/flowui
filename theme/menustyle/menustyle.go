// Package menustyle contains the theme-owned visual recipe for menu behavior.
// It deliberately does not import widget/menu: menu behavior supplies only
// semantic item facts and executes the returned declarations and renderers.
package menustyle

import (
	"image"
	"image/color"
	"math"
	"time"

	"gioui.org/f32"
	"gioui.org/font"
	"gioui.org/io/system"
	"gioui.org/layout"
	"gioui.org/op"
	"gioui.org/op/clip"
	"gioui.org/op/paint"
	giotext "gioui.org/text"
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

const itemPressDuration = 250 * time.Millisecond

// BuildMenu returns the complete menu visual protocol for the active theme.
func BuildMenu(active *theme.Theme, request theme.MenuRequest) theme.MenuVisual {
	if active == nil {
		panic(theme.ErrMissingVisualProvider)
	}
	tokens := active.Components.Menu
	if request.Compact {
		tokens = compact(tokens)
	}
	// Menu item visuals only vary by the two request bits. Build the complete
	// four-cell table once with the provider bundle instead of rebuilding the
	// same Style rule tree for every measurement and visible row.
	itemVisuals := [2][2]theme.MenuItemVisual{}
	for variant := range uint8(2) {
		for disabled := range uint8(2) {
			itemVisuals[variant][disabled] = itemVisual(active, tokens, theme.MenuItemRequest{
				Variant:  variant,
				Disabled: disabled != 0,
			})
		}
	}
	return theme.MenuVisual{
		Metrics:     metrics(tokens),
		Colors:      colors(active, tokens),
		Motion:      active.Motion,
		Root:        root(active, tokens),
		Override:    theme.MenuOverrideStyle(request),
		PanelColors: panelColors(active, tokens),
		Text:        textRenderer(active, tokens),
		TextWidth:   textWidth(active, tokens),
		Item: func(item theme.MenuItemRequest) theme.MenuItemVisual {
			variant := uint8(0)
			if item.Variant == itemDanger {
				variant = 1
			}
			disabled := uint8(0)
			if item.Disabled {
				disabled = 1
			}
			return itemVisuals[variant][disabled]
		},
		Separator:        separatorRenderer,
		Indicator:        indicatorRenderer,
		SubmenuIndicator: submenuRenderer,
	}
}

func panelColors(active *theme.Theme, tokens theme.MenuTheme) theme.MenuPanelColorsResolver {
	base := colors(active, tokens)
	return func(resolved style.ResolvedStyle) theme.MenuPanelColors {
		result := theme.MenuPanelColors{Background: base.PanelBackground, Border: base.PanelBorder, Valid: true}
		if resolved.Paint == nil {
			return result
		}
		if resolved.Paint.Background != nil {
			value, ok := resolved.Paint.Background.(style.SolidColor)
			if !ok {
				result.Valid = false
			} else {
				result.Background = value.Color
			}
		}
		if resolved.Paint.Border != nil {
			value, ok := resolved.Paint.Border.Color.(style.SolidColor)
			if !ok {
				result.Valid = false
			} else {
				result.Border = value.Color
			}
		}
		return result
	}
}

func compact(tokens theme.MenuTheme) theme.MenuTheme {
	tokens.Padding = minDp(tokens.Padding, 4)
	tokens.Radius = minDp(tokens.Radius, 8)
	tokens.ItemGap = minDp(tokens.ItemGap, 1)
	tokens.ItemMinHeight = minDp(tokens.ItemMinHeight, 28)
	tokens.ItemRadius = minDp(tokens.ItemRadius, 4)
	tokens.ItemPaddingX = minDp(tokens.ItemPaddingX, 8)
	tokens.ItemPaddingY = minDp(tokens.ItemPaddingY, 3)
	tokens.ItemContentGap = minDp(tokens.ItemContentGap, 8)
	tokens.ItemTextSize = minSp(tokens.ItemTextSize, 13)
	tokens.ItemDescriptionSize = minSp(tokens.ItemDescriptionSize, 11)
	tokens.ShortcutTextSize = minSp(tokens.ShortcutTextSize, 12)
	tokens.ShortcutHeight = minDp(tokens.ShortcutHeight, 20)
	tokens.ShortcutPaddingX = minDp(tokens.ShortcutPaddingX, 4)
	tokens.IndicatorSize = minDp(tokens.IndicatorSize, 14)
	tokens.CheckmarkSize = minDp(tokens.CheckmarkSize, 9)
	tokens.RadioDotSize = minDp(tokens.RadioDotSize, 6)
	tokens.SubmenuIndicatorSize = minDp(tokens.SubmenuIndicatorSize, 12)
	tokens.SectionTextSize = minSp(tokens.SectionTextSize, 11)
	tokens.SectionPaddingX = minDp(tokens.SectionPaddingX, 6)
	tokens.SectionPaddingTop = minDp(tokens.SectionPaddingTop, 4)
	tokens.SectionPaddingBottom = minDp(tokens.SectionPaddingBottom, 3)
	return tokens
}

func minDp(value, limit unit.Dp) unit.Dp {
	return unit.Dp(math.Min(float64(value), float64(limit)))
}

func minSp(value, limit unit.Sp) unit.Sp {
	return unit.Sp(math.Min(float64(value), float64(limit)))
}

func metrics(tokens theme.MenuTheme) theme.MenuMetrics {
	return theme.MenuMetrics{
		Width: tokens.Width, MaxHeight: tokens.MaxHeight, MaxWidthFraction: tokens.MaxWidthFraction,
		Padding: tokens.Padding, Radius: tokens.Radius, BorderWidth: tokens.BorderWidth,
		ItemGap: tokens.ItemGap, ItemMinHeight: tokens.ItemMinHeight, ItemRadius: tokens.ItemRadius,
		ItemPaddingX: tokens.ItemPaddingX, ItemPaddingY: tokens.ItemPaddingY, ItemContentGap: tokens.ItemContentGap,
		ShortcutHeight:   tokens.ShortcutHeight,
		ShortcutPaddingX: tokens.ShortcutPaddingX, IndicatorSize: tokens.IndicatorSize,
		IndicatorContentGap: tokens.IndicatorContentGap, CheckmarkSize: tokens.CheckmarkSize,
		RadioDotSize: tokens.RadioDotSize, IndicatorOffsetY: tokens.IndicatorOffsetY,
		SubmenuIndicatorSize: tokens.SubmenuIndicatorSize, FocusRingWidth: tokens.FocusRingWidth,
		FocusRingOffset: tokens.FocusRingOffset, PressedScale: pressedScale(tokens.PressedScale),
		SectionPaddingX:   tokens.SectionPaddingX,
		SectionPaddingTop: tokens.SectionPaddingTop, SectionPaddingBottom: tokens.SectionPaddingBottom,
		SeparatorMarginX: tokens.SeparatorMarginX, SeparatorMarginY: tokens.SeparatorMarginY,
		SeparatorWidth: tokens.SeparatorWidth, DescriptionLeadingHeight: tokens.DescriptionLeadingHeight,
		DescriptionLeadingInsetTop: tokens.DescriptionLeadingInsetTop, SubmenuGap: tokens.SubmenuGap,
		ContextMenuOffset: tokens.ContextMenuOffset, EnterScale: tokens.EnterScale, ExitScale: tokens.ExitScale,
		EnterDuration: tokens.EnterDuration, ExitDuration: tokens.ExitDuration, SubmenuOpenDelay: tokens.SubmenuOpenDelay,
		AnimationDistance: tokens.AnimationDistance, ShadowOpacity: clamp01(tokens.ShadowOpacity),
	}
}

func colors(active *theme.Theme, tokens theme.MenuTheme) theme.MenuColors {
	sem := active.Semantic()
	muted := theme.ColorOr(tokens.MutedColor, sem.MutedForeground)
	indicator := theme.ColorOr(tokens.IndicatorColor, sem.OverlayForeground)
	shadow := tokens.ShadowColor
	if shadow.A == 0 && tokens.ShadowOpacity > 0 {
		shadow = sem.OverlayShadow
	}
	return theme.MenuColors{
		PanelBackground: theme.ColorOr(tokens.BackgroundColor, sem.Overlay),
		PanelForeground: theme.ColorOr(tokens.ForegroundColor, sem.OverlayForeground),
		PanelBorder:     theme.ColorOr(tokens.BorderColor, sem.Border),
		PanelShadow:     shadow,
		Separator:       sem.Border,
		Foreground:      theme.ColorOr(tokens.ForegroundColor, sem.OverlayForeground),
		Description:     muted,
		Shortcut:        muted,
		Indicator:       indicator,
		Danger:          theme.ColorOr(tokens.DangerColor, sem.Tone(theme.ToneDanger).OnSoft),
		Focus:           theme.ColorOr(tokens.FocusColor, sem.Ring),
	}
}

func root(active *theme.Theme, tokens theme.MenuTheme) style.Style {
	colors := colors(active, tokens)
	return style.Style{}.
		Background(style.Solid(colors.PanelBackground)).
		TextColor(style.Solid(colors.PanelForeground)).
		BorderColor(style.Solid(colors.PanelBorder)).
		BorderWidth(tokens.BorderWidth).
		Radius(tokens.Radius).
		Shadow(style.ShadowMenu).
		Overflow(style.OverflowHidden)
}

func itemVisual(active *theme.Theme, tokens theme.MenuTheme, request theme.MenuItemRequest) theme.MenuItemVisual {
	sem := active.Semantic()
	colors := colors(active, tokens)
	foreground := colors.Foreground
	indicator := colors.Indicator
	if request.Variant == itemDanger {
		foreground = colors.Danger
		indicator = colors.Danger
	}
	if request.Disabled {
		foreground = active.DisabledColor(foreground)
		indicator = active.DisabledColor(indicator)
	}
	fill, fillText := itemHighlight(active, tokens, request.Variant)
	highlighted := style.Style{}.
		Background(style.Solid(fill)).
		TextColor(style.Solid(fillText))
	base := style.Style{}.
		MinHeight(tokens.ItemMinHeight).
		PaddingX(tokens.ItemPaddingX).
		PaddingY(tokens.ItemPaddingY).
		Radius(tokens.ItemRadius).
		Background(style.RGBA(0x00000000)).
		TextColor(style.Solid(foreground)).
		BorderWidth(tokens.FocusRingWidth).
		BorderColor(style.WithAlpha(sem.Ring, 0)).
		Scale(1).
		Transition(style.PropTransform, itemPressDuration)
	// The state rules live in the variant layer: the menu's text is inherited
	// into an item after its defaults, and would otherwise keep the highlighted
	// text dark.
	states := style.Style{}.
		Part(style.PartItem, style.Style{}.TextColor(style.Solid(foreground))).
		Rules(
			style.Rule{When: style.Hovered, Set: style.Style{}.Part(style.PartItem, highlighted)},
			style.Rule{When: style.FocusVisible, Set: style.Style{}.Part(style.PartItem, highlighted)},
			style.Rule{When: style.Open, Rank: -1, Set: style.Style{}.Part(style.PartItem, style.Style{}.Background(style.Solid(sem.HoverSurface())))},
			style.Rule{When: style.Pressed, Set: style.Style{}.Part(style.PartItem, style.Style{}.Scale(pressedScale(tokens.PressedScale)))},
			style.Rule{When: style.Selected, Set: style.Style{}.Part(style.PartItem, style.Style{}.Background(selectedBackground(active, tokens)))},
			style.Rule{When: style.Selected | style.Hovered, Set: style.Style{}.Part(style.PartItem, highlighted)},
			style.Rule{When: style.Selected | style.FocusVisible, Set: style.Style{}.Part(style.PartItem, highlighted)},
		)
	defaults := style.Style{}.Part(style.PartItem, base)
	baseColors := theme.MenuItemColors{
		Foreground:  foreground,
		Description: colors.Description,
		Shortcut:    colors.Shortcut,
		Indicator:   indicator,
	}
	return theme.MenuItemVisual{
		Defaults: defaults,
		Variant:  states,
		Colors: func(resolved style.ResolvedStyle) theme.MenuItemColors {
			result := baseColors
			if resolved.Text != nil {
				value := textrender.ColorOr(resolved, result.Foreground)
				result.Foreground = value
				result.Description = value
				result.Shortcut = value
			}
			return result
		},
	}
}

// itemHighlight returns the fill and the text color of a highlighted item. The
// highlighted item is the solid accent; a danger item uses the danger tone, and
// an explicit HoverColor token wins.
func itemHighlight(active *theme.Theme, tokens theme.MenuTheme, variant uint8) (fill, text color.NRGBA) {
	sem := active.Semantic()
	tone := sem.Tone(theme.TonePrimary)
	if variant == itemDanger {
		tone = sem.Tone(theme.ToneDanger)
	}
	if tokens.HoverColor.A != 0 {
		return tokens.HoverColor, theme.ColorOr(tokens.ForegroundColor, sem.OverlayForeground)
	}
	fill, text = tone.Fill, tone.OnFill
	// Mid-dark fills pair with white text. The shared luminance rule picks black
	// for red, which reads as a tinted row rather than a destructive action, so
	// prefer white when it stays legible.
	if theme.ContrastRatio(fill, white) >= 3 {
		text = white
	}
	return fill, text
}

var white = color.NRGBA{R: 0xff, G: 0xff, B: 0xff, A: 0xff}

func textRenderer(active *theme.Theme, tokens theme.MenuTheme) theme.MenuTextRenderer {
	return func(gtx layout.Context, request theme.MenuTextRequest) layout.Dimensions {
		return menuLabel(active, tokens, request).Layout(gtx)
	}
}

func textWidth(active *theme.Theme, tokens theme.MenuTheme) theme.MenuTextWidthResolver {
	var cache menuTextWidthCache
	return func(gtx layout.Context, request theme.MenuTextRequest) int {
		if request.Text == "" {
			return 0
		}
		label := menuLabel(active, tokens, request)
		key := menuTextWidthKey{
			Text:            request.Text,
			Role:            request.Role,
			Constraints:     gtx.Constraints,
			Metric:          gtx.Metric,
			Locale:          gtx.Locale,
			Font:            label.Font,
			Alignment:       label.Alignment,
			MaxLines:        label.MaxLines,
			WrapPolicy:      label.WrapPolicy,
			Truncator:       label.Truncator,
			TextSize:        label.TextSize,
			LineHeight:      label.LineHeight,
			LineHeightScale: label.LineHeightScale,
		}
		if width, ok := cache.lookup(key); ok {
			return width
		}
		measureGtx := gtx
		measureGtx.Constraints = layout.Constraints{
			Max: image.Pt(gtx.Constraints.Max.X, max(gtx.Constraints.Max.Y, gtx.Dp(1024))),
		}
		measureGtx.Ops = new(op.Ops)
		width := label.Layout(measureGtx).Size.X
		cache.store(key, width)
		return width
	}
}

const menuTextWidthCacheCapacity = 32

type menuTextWidthKey struct {
	Text            string
	Role            theme.MenuTextRole
	Constraints     layout.Constraints
	Metric          unit.Metric
	Locale          system.Locale
	Font            font.Font
	Alignment       giotext.Alignment
	MaxLines        int
	WrapPolicy      giotext.WrapPolicy
	Truncator       string
	TextSize        unit.Sp
	LineHeight      unit.Sp
	LineHeightScale float32
}

type menuTextWidthEntry struct {
	key   menuTextWidthKey
	width int
	valid bool
}

type menuTextWidthCache struct {
	entries [menuTextWidthCacheCapacity]menuTextWidthEntry
	next    uint8
}

func (c *menuTextWidthCache) lookup(key menuTextWidthKey) (int, bool) {
	for _, entry := range c.entries {
		if entry.valid && entry.key == key {
			return entry.width, true
		}
	}
	return 0, false
}

func (c *menuTextWidthCache) store(key menuTextWidthKey, width int) {
	index := int(c.next) % len(c.entries)
	c.entries[index] = menuTextWidthEntry{key: key, width: width, valid: true}
	c.next = uint8((index + 1) % len(c.entries))
}

func menuLabel(active *theme.Theme, tokens theme.MenuTheme, request theme.MenuTextRequest) material.LabelStyle {
	size := tokens.ItemTextSize
	weight := font.Normal
	switch request.Role {
	case theme.MenuTextEmpty:
	case theme.MenuTextSection:
		size = tokens.SectionTextSize
	case theme.MenuTextItem:
		weight = font.Medium
	case theme.MenuTextDescription:
		size = tokens.ItemDescriptionSize
	case theme.MenuTextShortcut:
		size = tokens.ShortcutTextSize
		weight = font.Medium
	default:
		panic(theme.ErrIncompleteVisualProvider)
	}
	if request.Style.Text != nil {
		textStyle := request.Style.Text
		if textStyle.FontSize != nil && *textStyle.FontSize > 0 {
			size = *textStyle.FontSize
		}
		if request.Role == theme.MenuTextItem && textStyle.FontWeight != nil {
			weight = font.Weight(*textStyle.FontWeight)
		}
	}
	label := material.Label(theme.MaterialOf(active), size, request.Text)
	label.Color = textrender.ColorOr(request.Style, request.Color)
	label.SelectionColor = active.Semantic().Selection
	label.Font.Weight = weight
	if request.Style.Text != nil {
		textStyle := request.Style.Text
		if textStyle.Typeface != nil {
			label.Font.Typeface = *textStyle.Typeface
		}
		if textStyle.FontStyle != nil {
			label.Font.Style = *textStyle.FontStyle
		}
		if textStyle.LineHeight != nil {
			label.LineHeight = *textStyle.LineHeight
		}
		if textStyle.LineHeightScale != nil {
			label.LineHeightScale = *textStyle.LineHeightScale
		}
	}
	return label
}

func selectedBackground(active *theme.Theme, tokens theme.MenuTheme) style.SolidColor {
	if tokens.SelectedColor.A != 0 {
		return style.Solid(tokens.SelectedColor)
	}
	return style.Solid(active.Semantic().Tone(theme.TonePrimary).Soft)
}

func pressedScale(value float32) float32 {
	if value <= 0 || value > 1 {
		return .98
	}
	return value
}

func clamp01(value float32) float32 {
	if value < 0 {
		return 0
	}
	if value > 1 {
		return 1
	}
	return value
}

func separatorRenderer(gtx layout.Context, size image.Point, foreground color.NRGBA) {
	if size.X <= 0 || size.Y <= 0 || foreground.A == 0 {
		return
	}
	paint.FillShape(gtx.Ops, foreground, clip.Rect{Max: size}.Op())
}

func indicatorRenderer(gtx layout.Context, request theme.MenuIndicatorRequest) {
	if request.Size.X <= 0 || request.Size.Y <= 0 || request.Color.A == 0 {
		return
	}
	if request.Type == 2 {
		dot := minInt(maxInt(request.DotSize, 1), minInt(request.Size.X, request.Size.Y))
		rect := image.Rect((request.Size.X-dot)/2, (request.Size.Y-dot)/2, (request.Size.X+dot)/2, (request.Size.Y+dot)/2)
		paint.FillShape(gtx.Ops, request.Color, clip.Ellipse(rect).Op(gtx.Ops))
		return
	}
	if request.Type != 1 {
		return
	}
	check := minInt(maxInt(request.CheckmarkSize, 1), minInt(request.Size.X, request.Size.Y))
	offset := (request.Size.X - check) / 2
	points := [3]f32.Point{
		f32.Pt(float32(offset)+float32(check)*.05, float32(request.Size.Y)*.56),
		f32.Pt(float32(offset)+float32(check)*.40, float32(request.Size.Y)*.86),
		f32.Pt(float32(offset)+float32(check)*.95, float32(request.Size.Y)*.14),
	}
	path := draw.CheckPath(gtx.Ops, points, 1)
	stroke := clip.Stroke{Path: path, Width: max(draw.DpFloat(gtx, unit.Dp(1.5)), 1)}.Op().Push(gtx.Ops)
	paint.Fill(gtx.Ops, request.Color)
	stroke.Pop()
}

func submenuRenderer(gtx layout.Context, size image.Point, foreground color.NRGBA) layout.Dimensions {
	if size.X <= 0 || size.Y <= 0 || foreground.A == 0 {
		return layout.Dimensions{Size: size}
	}
	width := float32(size.X)
	height := float32(size.Y)
	path := clip.Path{}
	path.Begin(gtx.Ops)
	path.MoveTo(f32.Pt(width*.35, height*.18))
	path.LineTo(f32.Pt(width*.68, height*.5))
	path.LineTo(f32.Pt(width*.35, height*.82))
	pathSpec := path.End()
	stroke := clip.Stroke{Path: pathSpec, Width: max(draw.DpFloat(gtx, unit.Dp(1.5)), 1)}.Op().Push(gtx.Ops)
	paint.Fill(gtx.Ops, foreground)
	stroke.Pop()
	return layout.Dimensions{Size: size}
}

func maxInt(a, b int) int {
	if a > b {
		return a
	}
	return b
}

func minInt(a, b int) int {
	if a < b {
		return a
	}
	return b
}
