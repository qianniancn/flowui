package card

import (
	"gioui.org/layout"
	"gioui.org/unit"
	"github.com/qianniancn/flowui/internal/computed"
	"github.com/qianniancn/flowui/internal/core"
	layoutui "github.com/qianniancn/flowui/layout"
	"github.com/qianniancn/flowui/style"
	"github.com/qianniancn/flowui/theme"
)

// Variant selects the semantic surface prominence of a card.
type Variant = theme.CardVariant

const (
	Default     = theme.CardDefault
	Secondary   = theme.CardSecondary
	Tertiary    = theme.CardTertiary
	Transparent = theme.CardTransparent
)

// Widget groups related content on a raised semantic surface. The look comes
// from the active theme, not from this package.
type Widget struct {
	children    []core.Widget
	variant     Variant
	visualCache *cardVisualCache
	override    *theme.CardOverride
}

func New(children ...core.Widget) Widget {
	return Widget{children: children, visualCache: &cardVisualCache{}}
}

func (c Widget) Variant(variant Variant) Widget {
	c.variant = variant
	return c
}

func (c Widget) Layout(ctx *core.Context, gtx layout.Context) layout.Dimensions {
	return c.layoutWithOverride(ctx, gtx, c.override)
}

// Style merges instance declarations into the card. The declaration reaches
// the provider at layout time, so Style can sit anywhere in the chain.
func (c Widget) Style(values ...style.Style) Widget {
	declaration := style.Join(values...)
	if c.override == nil {
		c.override = &theme.CardOverride{}
	}
	c.override.Declaration = style.Join(c.override.Declaration, declaration)
	return c
}

// StyleWithMetrics sets the instance declaration together with explicit
// internal padding and gap.
func (c Widget) StyleWithMetrics(declaration style.Style, padding, gap unit.Dp) Widget {
	padding = max(padding, 0)
	gap = max(gap, 0)
	if c.override == nil {
		c.override = &theme.CardOverride{}
	}
	c.override.Declaration = style.Join(c.override.Declaration, declaration)
	c.override.Padding = &padding
	c.override.Gap = &gap
	return c
}

func (c Widget) layoutWithOverride(ctx *core.Context, gtx layout.Context, override *theme.CardOverride) layout.Dimensions {
	visual := c.visualWithCache(ctx, override)
	prepareCardFieldAssociations(ctx, c.children...)
	content := layoutui.Box(
		layoutui.Column(nonNilWidgets(c.children)...).Gap(int(visual.Metrics.Gap)),
	)
	resolved := computed.ResolveStatic(
		ctx,
		style.State{},
		visual.Defaults,
		visual.Variant,
		style.Style{},
		visual.Override,
	)
	if len(resolved.Transitions) != 0 {
		resolved = computed.ApplyTransitions(ctx, gtx, core.ClaimKey(ctx, core.KindStyle, "card"), resolved)
	}
	return layoutui.Resolved(ctx, gtx, resolved, content)
}

func prepareCardFieldAssociations(ctx *core.Context, widgets ...core.Widget) {
	for _, widget := range widgets {
		switch widget := widget.(type) {
		case Widget:
			prepareCardFieldAssociations(ctx, widget.children...)
		case *Widget:
			if widget != nil {
				prepareCardFieldAssociations(ctx, widget.children...)
			}
		default:
			layoutui.PrepareFieldAssociations(ctx, widget)
		}
	}
}

func nonNilWidgets(widgets []core.Widget) []core.Widget {
	result := make([]core.Widget, 0, len(widgets))
	for _, widget := range widgets {
		if widget != nil {
			result = append(result, widget)
		}
	}
	return result
}
