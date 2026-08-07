# 10 - Custom components

Start with composition. Only drop to custom layout or drawing when the public
components and styles cannot express the behavior.

## Path A: composition

Build a reusable function from existing widgets and return a value-like widget:

```go
func StatusPill(label string, selected bool) ui.Widget {
	style := ui.PaddingX(12).PaddingY(6).Radius(999).
		When(ui.Hovered, ui.Background(ui.TokenSurfaceTertiary)).
		When(ui.If(selected), ui.Background(ui.TokenAccent))
	return ui.Box(ui.Text(label)).Style(style)
}
```

Composition keeps interaction, semantics, and theme behavior in the existing
components.

## Path B: custom interactive layout

For a new interactive control, retain state with a stable key, resolve a style,
and use the public layout helpers. Keep the widget's event host around the
whole interactive area and leave painting to a small child widget.

The public sequence is:

```text
UseState -> ResolveStyle / ResolveStylePart ->
LayoutInteractiveResolvedStyle (or LayoutResolvedStyle)
```

`examples/custom_widgets` contains a complete implementation. Do not use
internal component packages or old names such as `BeginInteract`, `Resolve`,
or `LayoutInteractiveBox`; those are not part of the public API.

The [`examples/custom_widgets`](https://github.com/qianniancn/flowui/tree/main/examples/custom_widgets) program shows a
custom control and a canvas-backed widget. `ui.WidgetFunc` is useful when only a
small layout callback is needed.

## Path C: custom drawing

Use Gio operations for the content of a chart or illustration, but keep the
outer box, hit target, focus behavior, and theme tokens in FlowUI components.
Canvas code should not duplicate button or field semantics.

For common drawing and measurement tasks, reuse the public helpers instead of
duplicating theme resolution or Gio event boilerplate:

```go
brush, ok := ui.ResolveBrush(ctx, ui.LinearGradient(
	ui.ColorStop(0, ui.TokenAccent),
	ui.ColorStop(1, ui.TokenDanger),
))
if ok {
	ui.DrawBrush(gtx, image.Rect(0, 0, 160, 32), 6, brush)
}

size := ui.MeasureText(ctx, gtx, ui.Text("Preview").Size(14).MaxLines(1))
_ = size
```

For intrinsic sizing of an arbitrary child, implement the optional
`ui.Measurable` interface or call `ui.MeasureWidget`. The measurement pass
uses a separate ops list and an empty input source, so it cannot consume
pointer or keyboard input:

An implementation of `Measure` must only report dimensions; it must not read or
consume input from the current frame.

```go
type measuredBadge struct{}

func (measuredBadge) Measure(_ *ui.Context, gtx layout.Context) layout.Dimensions {
	return layout.Dimensions{Size: image.Pt(gtx.Dp(48), gtx.Dp(20))}
}
```

Low-level pointer widgets can use `AddPointerArea`, `NextPointerEvent`,
`IsPrimaryPointerPress`, and `GrabPointer`. Use `LayoutVisualOverflow` for a
local ripple or focus decoration that must cross the child's own clip. It is
not a replacement for `Popover`, `Portal`, or another overlay host. When that
decoration also needs room inside an ancestor scroll or split viewport, report
its top/right/bottom/left extent with `LayoutVisualOutset`; a component with a
Style shell can use `VisualOutset` instead.

## Rules of thumb

1. Use a public component before copying its implementation.
2. Keep business state in `Model` and interaction state under a stable key.
3. Resolve inherited and state styles rather than hard-coding colors.
4. Give custom controls keyboard focus and semantic labels where appropriate.
5. Test layout and pointer behavior with `uitest.Harness`.

Continue with [Multiple windows](11-multiple-windows.md).
