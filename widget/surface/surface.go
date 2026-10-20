package surface

import (
	"gioui.org/layout"
	"github.com/qianniancn/flowui/internal/core"
	"github.com/qianniancn/flowui/style"
	"github.com/qianniancn/flowui/theme"
)

type Variant = theme.SurfaceVariant

const (
	Default     = theme.SurfaceDefault
	Secondary   = theme.SurfaceSecondary
	Tertiary    = theme.SurfaceTertiary
	Transparent = theme.SurfaceTransparent
)

// Widget provides a semantic content surface around a child. Its visual
// declaration is supplied by the active theme provider.
type Widget struct {
	override    *theme.SurfaceOverride
	child       core.Widget
	variant     Variant
	visualCache *surfaceVisualCache
}

func New(child core.Widget) Widget {
	return Widget{child: child, visualCache: &surfaceVisualCache{}}
}

func (s Widget) Variant(variant Variant) Widget {
	s.variant = variant
	return s
}

// Layout resolves the provider-owned surface and scopes its background and
// foreground around the child while preserving the child's behavior.
func (s Widget) Layout(ctx *core.Context, gtx layout.Context) layout.Dimensions {
	return s.layoutWithOverride(ctx, gtx, s.override)
}

// Style merges instance declarations into the widget. The declaration
// reaches the provider at layout time, so Style can sit anywhere in the chain.
func (w Widget) Style(values ...style.Style) Widget {
	declaration := style.Join(values...)
	if w.override == nil {
		w.override = &theme.SurfaceOverride{}
	}
	w.override.Declaration = style.Join(w.override.Declaration, declaration)
	return w
}
