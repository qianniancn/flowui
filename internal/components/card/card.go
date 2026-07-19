package card

import (
	"gioui.org/layout"
	"gioui.org/unit"
	"github.com/qianniancn/FlowUI/internal/components/layout"
	"github.com/qianniancn/FlowUI/internal/components/surface"
	"github.com/qianniancn/FlowUI/internal/frame"
	"github.com/qianniancn/FlowUI/internal/theme"
)

// CardVariant selects the semantic surface prominence of a card.
type CardVariant uint8

const (
	CardDefault CardVariant = iota
	CardSecondary
	CardTertiary
	CardTransparent
)

// CardWidget groups related content on a HeroUI-style semantic surface.
type CardWidget struct {
	theme      func(*theme.Theme)
	children   []frame.Widget
	variant    CardVariant
	padding    unit.Dp
	gap        unit.Dp
	radius     unit.Dp
	shadow     bool
	hasPadding bool
	hasGap     bool
	hasRadius  bool
	hasShadow  bool
}

func Card(children ...frame.Widget) CardWidget {
	return CardWidget{children: children}
}

func (c CardWidget) Variant(variant CardVariant) CardWidget {
	c.variant = variant
	return c
}

func (c CardWidget) Padding(dp int) CardWidget {
	c.padding = unit.Dp(max(dp, 0))
	c.hasPadding = true
	return c
}

func (c CardWidget) Gap(dp int) CardWidget {
	c.gap = unit.Dp(max(dp, 0))
	c.hasGap = true
	return c
}

func (c CardWidget) Radius(dp int) CardWidget {
	c.radius = unit.Dp(max(dp, 0))
	c.hasRadius = true
	return c
}

func (c CardWidget) Shadow(enabled bool) CardWidget {
	c.shadow = enabled
	c.hasShadow = true
	return c
}

func (c CardWidget) Theme(fn func(*theme.Theme)) CardWidget {
	c.theme = fn
	return c
}

func (c CardWidget) Layout(ctx *frame.Context, gtx layout.Context) layout.Dimensions {
	if restore := frame.PushInstanceTheme(ctx, c.theme); restore != nil {
		defer restore()
	}
	prepareCardFieldAssociations(ctx, c.children...)
	tokens := frame.ActiveTheme(ctx).Components.Card
	padding := tokens.Padding
	if c.hasPadding {
		padding = c.padding
	}
	gap := tokens.Gap
	if c.hasGap {
		gap = c.gap
	}
	radius := tokens.Radius
	if c.hasRadius {
		radius = c.radius
	}

	content := layoutui.Box(
		layoutui.Column(nonNilWidgets(c.children)...).Gap(int(gap)),
	).Padding(int(padding))

	return surface.Surface(content).
		Variant(cardSurfaceVariant(c.variant)).
		Radius(int(radius)).
		Shadow(c.resolvedShadow()).
		Layout(ctx, gtx)
}

func prepareCardFieldAssociations(ctx *frame.Context, widgets ...frame.Widget) {
	for _, widget := range widgets {
		switch widget := widget.(type) {
		case CardWidget:
			prepareCardFieldAssociations(ctx, widget.children...)
		case *CardWidget:
			if widget != nil {
				prepareCardFieldAssociations(ctx, widget.children...)
			}
		default:
			layoutui.PrepareFieldAssociations(ctx, widget)
		}
	}
}

func (c CardWidget) resolvedShadow() bool {
	if c.hasShadow {
		return c.shadow
	}
	return c.variant != CardTransparent
}

func cardSurfaceVariant(variant CardVariant) surface.SurfaceVariant {
	switch variant {
	case CardSecondary:
		return surface.SurfaceSecondary
	case CardTertiary:
		return surface.SurfaceTertiary
	case CardTransparent:
		return surface.SurfaceTransparent
	default:
		return surface.SurfaceDefault
	}
}

func nonNilWidgets(widgets []frame.Widget) []frame.Widget {
	result := make([]frame.Widget, 0, len(widgets))
	for _, widget := range widgets {
		if widget != nil {
			result = append(result, widget)
		}
	}
	return result
}
