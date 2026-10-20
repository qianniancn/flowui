// Package resolve exposes the themed style resolvers used by custom
// components and custom drawing.
//
// It is deliberately separate from github.com/qianniancn/flowui/style. The
// style package is a pure declaration mechanism: it depends on nothing in the
// framework, which is what lets theme depend on it. Resolving a declaration
// needs the active frame context and the active theme values, so those entry
// points live here instead — keeping the dependency direction one-way
// (resolve -> computed -> theme -> style).
//
// There are two resolution paths, and they exist for different call shapes.
// Resolve and ResolvePart resolve a declaration and apply its transitions in
// one call — use them when the call site is the render pass. A component that
// must also resolve during measurement uses the static pair instead and applies
// transition state once per frame:
//
//	root := resolve.ResolveStatic(ctx, state, base, instance)
//	if resolve.HasTransitions(root) {
//		root = resolve.ApplyTransitions(ctx, gtx, key, root)
//	}
//
// key is the same widget identity in both paths. An empty key means "no
// identity": transitions are not applied and values stay at their resolved
// targets, so a caller without a stable key does not panic on state slots.
package resolve

import (
	"image/color"

	"gioui.org/layout"
	"github.com/qianniancn/flowui/draw"
	"github.com/qianniancn/flowui/internal/computed"
	"github.com/qianniancn/flowui/internal/core"
	"github.com/qianniancn/flowui/style"
)

// Context is the per-frame context used by style resolvers.
//
// The alias keeps these helpers usable from public code without exposing the
// internal core package in their signatures.
type Context = core.Context

// Widget is a composable FlowUI node accepted by Scope.
type Widget = core.Widget

// WidgetFunc adapts a layout function to Widget.
type WidgetFunc = core.WidgetFunc

// Resolve resolves a custom component's base style, inherited scopes,
// and instance style in that order, including transitions.
func Resolve(
	ctx *Context,
	gtx layout.Context,
	key string,
	state style.State,
	base, instance style.Style,
) style.ResolvedStyle {
	return computed.Resolve(ctx, gtx, scopedKey(ctx, key), state, base, style.Style{}, style.Style{}, instance)
}

// ResolveStatic resolves the same cascade without retaining transition
// state. Custom components can use it for measurement passes.
func ResolveStatic(ctx *Context, state style.State, base, instance style.Style) style.ResolvedStyle {
	return computed.ResolveStatic(ctx, state, base, style.Style{}, style.Style{}, instance)
}

// ResolvePart resolves a named part of a custom component, including
// inherited scopes, conditions, and transitions.
func ResolvePart(
	ctx *Context,
	gtx layout.Context,
	key string,
	part style.Part,
	state style.State,
	base, instance style.Style,
) style.ResolvedStyle {
	return computed.ResolvePart(ctx, gtx, scopedKey(ctx, key), part, state, base, style.Style{}, style.Style{}, instance)
}

// ResolvePartStatic resolves the same part without transition state.
func ResolvePartStatic(ctx *Context, part style.Part, state style.State, base, instance style.Style) style.ResolvedStyle {
	return computed.ResolvePartStatic(ctx, part, state, base, style.Style{}, style.Style{}, instance)
}

// HasTransitions reports whether any resolved style needs runtime state.
// Check it before claiming a key, so stateless styles stay free of animation
// bookkeeping.
func HasTransitions(values ...style.ResolvedStyle) bool {
	return computed.HasTransitions(values...)
}

// ApplyTransitions animates a style that has already been resolved by
// ResolveStatic. Styles without transitions remain stateless. The key is the
// identity the same component passes to Resolve or ResolvePart.
//
// An empty key soft-fails: transitions are not applied, so values stay at
// their resolved targets.
func ApplyTransitions(ctx *Context, gtx layout.Context, key string, resolved style.ResolvedStyle) style.ResolvedStyle {
	return computed.ApplyTransitions(ctx, gtx, scopedKey(ctx, key), resolved)
}

// ApplyPartTransitions animates an already-resolved named part, sharing the
// identity ApplyTransitions uses for the component root.
func ApplyPartTransitions(ctx *Context, gtx layout.Context, key string, part style.Part, resolved style.ResolvedStyle) style.ResolvedStyle {
	return computed.ApplyPartTransitions(ctx, gtx, scopedKey(ctx, key), part, resolved)
}

// CloneResolvedStyle returns an independently owned copy of a resolved style.
// Use it before mutating a snapshot returned by Resolve or ResolveStatic when
// the snapshot may be cache-owned.
func CloneResolvedStyle(value style.ResolvedStyle) style.ResolvedStyle {
	return computed.CloneResolvedStyle(value)
}

// Color resolves a color source to its concrete value. Sources carry concrete
// colors since the theme token types were retired, so no context is needed.
func Color(source style.ColorSource) (color.NRGBA, bool) {
	return computed.Color(source)
}

// ApplyOpacity applies a resolved Paint opacity to a concrete color.
func ApplyOpacity(value color.NRGBA, opacity *float32) color.NRGBA {
	return computed.ApplyOpacity(value, opacity)
}

// ApplyOutlineOpacity returns a copy of value whose outline alpha is
// multiplied by opacity, leaving the rest of the paint untouched.
func ApplyOutlineOpacity(value style.ResolvedStyle, opacity float32) style.ResolvedStyle {
	return computed.ApplyOutlineOpacity(value, opacity)
}

// Brush converts a paint source into a brush for custom drawing, mirroring
// Color for solid colors. Paint sources carry concrete colors since the theme
// token types were retired, so conversion is a pure type dispatch.
func Brush(source style.PaintSource) (draw.Brush, bool) {
	return computed.Brush(source)
}

// Scope applies value to all descendant components that support the
// common style cascade. An instance Style call still has the final say.
func Scope(value style.Style, child Widget) Widget {
	return WidgetFunc(func(ctx *Context, gtx layout.Context) layout.Dimensions {
		if child == nil {
			return layout.Dimensions{}
		}
		restore := core.PushStyle(ctx, value)
		defer restore()
		return child.Layout(ctx, gtx)
	})
}

// scopedKey qualifies a caller key with the current key scope.
//
// An empty key is passed through untouched: FullKey turns it into the bare
// scope prefix inside a nested key scope, which would read as real identity and
// silently share one animation slot between unrelated widgets instead of
// soft-failing the way the static pair documents.
func scopedKey(ctx *Context, key string) string {
	if key == "" {
		return ""
	}
	return core.FullKey(ctx, key)
}
