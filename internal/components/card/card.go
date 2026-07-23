package card

import (
	"gioui.org/layout"
	"gioui.org/unit"
	"github.com/qianniancn/FlowUI/internal/components/layout"
	"github.com/qianniancn/FlowUI/internal/components/surface"
	"github.com/qianniancn/FlowUI/internal/frame"
	flowstyle "github.com/qianniancn/FlowUI/internal/style"
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
	children    []frame.Widget
	customStyle flowstyle.Style
	variant     CardVariant
	padding     unit.Dp
	gap         unit.Dp
	hasPadding  bool
	hasGap      bool
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

func (c CardWidget) Style(value flowstyle.Style) CardWidget {
	c.customStyle = value
	return c
}

func (c CardWidget) Layout(ctx *frame.Context, gtx layout.Context) layout.Dimensions {
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
	content := layoutui.Box(
		layoutui.Column(nonNilWidgets(c.children)...).Gap(int(gap)),
	)
	root := cardRootDeclaration(frame.ActiveTheme(ctx), c.variant).
		Padding(padding)

	return surface.Surface(content).
		Variant(cardSurfaceVariant(c.variant)).
		Style(flowstyle.Join(root, c.customStyle)).
		Layout(ctx, gtx)
}

func cardRootDeclaration(activeTheme *theme.Theme, variant CardVariant) flowstyle.Style {
	builder := flowstyle.Style{}.Radius(activeTheme.Components.Card.Radius)
	if variant != CardTransparent {
		builder = builder.Shadow(flowstyle.ShadowSurface)
	}
	return builder
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
