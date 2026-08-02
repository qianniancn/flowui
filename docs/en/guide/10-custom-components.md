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

The [`examples/custom_widgets`](https://github.com/qianniancn/flowui/tree/main/examples/custom_widgets) program shows a
custom control and a canvas-backed widget. `ui.WidgetFunc` is useful when only a
small layout callback is needed.

## Path C: custom drawing

Use Gio operations for the content of a chart or illustration, but keep the
outer box, hit target, focus behavior, and theme tokens in FlowUI components.
Canvas code should not duplicate button or field semantics.

## Rules of thumb

1. Use a public component before copying its implementation.
2. Keep business state in `Model` and interaction state under a stable key.
3. Resolve inherited and state styles rather than hard-coding colors.
4. Give custom controls keyboard focus and semantic labels where appropriate.
5. Test layout and pointer behavior with `uitest.Harness`.

Continue with [Multiple windows](11-multiple-windows.md).
