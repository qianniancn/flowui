package label

import (
	"gioui.org/layout"
	"github.com/qianniancn/flowui/internal/computed"
	"github.com/qianniancn/flowui/internal/core"
	"github.com/qianniancn/flowui/style"
	"github.com/qianniancn/flowui/theme"
)

// Widget identifies and describes a form field.
type Widget struct {
	override    *theme.LabelOverride
	text        string
	forKey      string
	required    bool
	disabled    bool
	invalid     bool
	visualCache *labelVisualCache
}

// New creates a form label.
func New(text string) Widget {
	return Widget{text: text, visualCache: &labelVisualCache{}}
}

// For associates the label with a keyed form control.
func (l Widget) For(key string) Widget {
	l.forKey = key
	return l
}

func (l Widget) Required(required bool) Widget {
	l.required = required
	return l
}

func (l Widget) Disabled(disabled bool) Widget {
	l.disabled = disabled
	return l
}

func (l Widget) Invalid(invalid bool) Widget {
	l.invalid = invalid
	return l
}

// Layout lays out the label and preserves its field focus and accessibility
// behavior while obtaining all visual declarations from the active provider.
func (l Widget) Layout(ctx *core.Context, gtx layout.Context) layout.Dimensions {
	return l.layoutWithOverride(ctx, gtx, l.override)
}

func (l Widget) resolveLayoutStyle(
	ctx *core.Context,
	gtx layout.Context,
	state style.State,
	visual theme.LabelVisual,
) labelResolvedStyle {
	resolved := l.resolveStyleStatic(ctx, state, visual)
	if len(resolved.root.Transitions) == 0 {
		return resolved
	}
	resolved.root = computed.ApplyTransitions(ctx, gtx, core.ClaimKey(ctx, core.KindStyle, "label"), resolved.root)
	return resolved
}

func (l Widget) resolveStyle(
	ctx *core.Context,
	gtx layout.Context,
	key string,
	state style.State,
	visual theme.LabelVisual,
) labelResolvedStyle {
	resolved := l.resolveStyleStatic(ctx, state, visual)
	if len(resolved.root.Transitions) == 0 {
		return resolved
	}
	resolved.root = computed.ApplyTransitions(ctx, gtx, key, resolved.root)
	return resolved
}

func (l Widget) resolveStyleStatic(
	ctx *core.Context,
	state style.State,
	visual theme.LabelVisual,
) labelResolvedStyle {
	return labelResolvedStyle{
		root: computed.ResolveStatic(
			ctx,
			state,
			visual.Defaults,
			style.Style{},
			style.Style{},
			visual.Override,
		),
		required: computed.ResolveStatic(
			ctx,
			state,
			visual.Required,
			style.Style{},
			style.Style{},
			style.Style{},
		),
	}
}

func (l Widget) visualWithCache(ctx *core.Context, override *theme.LabelOverride) theme.LabelVisual {
	active := core.ActiveTheme(ctx)
	request := theme.LabelRequest{Foreground: ctx.ForegroundColor(), Override: override}
	if override != nil || l.visualCache == nil {
		return resolveLabelVisual(active, request)
	}
	key := labelVisualKey{Foreground: request.Foreground}
	return l.visualCache.cache.Resolve(active, key, func() theme.LabelVisual {
		return resolveLabelVisual(active, request)
	})
}

func resolveLabelVisual(active *theme.Theme, request theme.LabelRequest) theme.LabelVisual {
	return theme.ResolveVisual(active, request, theme.ComponentVisualProvider.Label, theme.ValidateLabelVisual)
}

func (l Widget) registerFieldAssociation(ctx *core.Context) string {
	fieldKey := core.FullKey(ctx, l.forKey)
	core.RegisterFieldLabel(ctx, fieldKey, l.text)
	return fieldKey
}

// PrepareFieldAssociation registers the label before a container chooses its
// child layout order.
func (l Widget) PrepareFieldAssociation(ctx *core.Context) bool {
	if l.forKey == "" {
		return false
	}
	core.PrepareFieldLabel(ctx, core.FullKey(ctx, l.forKey), l.text)
	return true
}

// Style merges instance declarations into the widget. The declaration
// reaches the provider at layout time, so Style can sit anywhere in the chain.
func (w Widget) Style(values ...style.Style) Widget {
	declaration := style.Join(values...)
	if w.override == nil {
		w.override = &theme.LabelOverride{}
	}
	w.override.Declaration = style.Join(w.override.Declaration, declaration)
	return w
}
