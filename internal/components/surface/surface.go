package surface

import (
	"gioui.org/layout"
	layoutui "github.com/qianniancn/flowui/internal/components/layout"
	"github.com/qianniancn/flowui/internal/frame"
	stateutil "github.com/qianniancn/flowui/internal/state"
	flowstyle "github.com/qianniancn/flowui/internal/style"
	styleruntime "github.com/qianniancn/flowui/internal/style/runtime"
)

type SurfaceVariant uint8

const (
	SurfaceDefault SurfaceVariant = iota
	SurfaceSecondary
	SurfaceTertiary
	SurfaceTransparent
)

// SurfaceWidget provides a semantic, theme-aware background for non-overlay content.
type SurfaceWidget struct {
	child       frame.Widget
	customStyle flowstyle.Style
	variant     SurfaceVariant
}

func Surface(child frame.Widget) SurfaceWidget {
	return SurfaceWidget{child: child}
}

func (s SurfaceWidget) Variant(variant SurfaceVariant) SurfaceWidget {
	s.variant = variant
	return s
}

func (s SurfaceWidget) Style(value flowstyle.Style) SurfaceWidget {
	s.customStyle = value
	return s
}

func (s SurfaceWidget) Layout(ctx *frame.Context, gtx layout.Context) layout.Dimensions {
	state := flowstyle.StyleState{}
	defaults := surfaceDefaultDeclaration(frame.ActiveTheme(ctx))
	variant := surfaceVariantDeclaration(frame.ActiveTheme(ctx), s.variant)
	resolved := styleruntime.ResolveStatic(
		ctx,
		state,
		defaults,
		variant,
		flowstyle.Style{},
		s.customStyle,
	)
	if len(resolved.Transitions) != 0 {
		key := frame.ClaimKey(ctx, stateutil.KindStyle, "surface")
		resolved = styleruntime.ApplyTransitions(ctx, gtx, key, resolved)
	}
	return layoutui.LayoutResolved(ctx, gtx, resolved, s.child)
}
