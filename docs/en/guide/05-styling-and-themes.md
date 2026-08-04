# 05 - Styling and themes

`ui.Style` is an immutable declaration of box geometry and appearance. Start
with a free function and chain the properties you need:

```go
primary := ui.Background(ui.TokenAccent).
	TextColor(ui.TokenAccentForeground).
	Radius(8).
	PaddingX(12).
	PaddingY(6).
	Cursor(ui.CursorPointer)
```

Apply it to a widget:

```go
ui.Button("save", ui.Text("Save")).Style(primary)
ui.Box(child).Style(ui.Padding(24))
```

There is no separate builder step. Each method returns a new style value.

## Cascade order

The effective style is resolved in this order:

```text
component defaults
  -> inherited text
  -> variant
  -> size
  -> StyleScope ancestors
  -> instance Style
  -> state conditions and transitions
```

Later declarations at the same layer replace earlier values. Use `When` for
runtime states:

```go
primary := ui.Background(ui.TokenAccent).
	TextColor(ui.TokenAccentForeground).
	When(ui.Hovered, ui.Background(ui.TokenAccentHover)).
	When(ui.Pressed, ui.Background(ui.TokenAccentPressed)).
	When(ui.Disabled, ui.Opacity(0.5))
```

`StyleScope` supplies defaults to a subtree without changing the global theme.
`Part` targets a named part inside a compound control, such as
`PartContent`, `PartTrack`, or `PartPanel`.

## Tokens and colors

Prefer semantic theme tokens so the same style works in light and dark themes:

```go
ui.Background(ui.TokenSurface)
ui.TextColor(ui.TokenForeground)
ui.Background(ui.TokenAccent)
```

Use literal colors when a design requires them:

```go
ui.RGB(0x0078d4)          // opaque 0xRRGGBB
ui.RGBA(0x00000030)       // 0xRRGGBBAA
ui.WithAlpha(ui.TokenFocus, 0.5)
```

## Gradients and custom drawing

When a custom widget needs a gradient, resolve it against the active theme and
reuse the public drawing helpers:

```go
brush, ok := ui.ResolveBrush(ctx, ui.LinearGradient(
	ui.ColorStop(0, ui.TokenAccent),
	ui.ColorStop(1, ui.TokenDanger),
))
if ok {
	ui.DrawBrush(gtx, image.Rect(0, 0, 160, 32), 6, brush)
}
```

Use `ResolveColor` for a single theme-resolved color and `DrawBrushRRect` when
you already have a `clip.RRect`. This keeps theme resolution in one place and
does not turn custom drawing into an overlay system.

## Themes

Choose a theme at startup or customize a copy of the default tokens:

```go
ui.Run(ui.NewProgram(Model{}, Update, View),
	ui.WithTheme(ui.DarkTheme()),
	ui.CustomizeTheme(func(theme *ui.Theme) {
		theme.Palette.Accent = color.NRGBA{R: 0x00, G: 0x78, B: 0xd4, A: 0xff}
		theme.Components.Button.Radius = 8
	}),
)
```

For an `Application`, use `SetTheme` and `SetLanguage` so the replacement is
applied on the target window's event loop.

## Fonts

The default theme uses `sans-serif` and `monospace` and permits Gio's system
font fallback. This is convenient for multilingual text, but the system font
index can add startup and memory cost. For deterministic rendering, parse a
bundled TTF, OTF, or TTC and disable system fonts:

```go
//go:embed fonts/Inter-Regular.ttf
var interRegular []byte

faces, err := ui.ParseFontCollection(interRegular)
if err != nil {
	panic(err)
}

theme := ui.DefaultTheme()
	theme.Typography.Typeface = "Inter"
	theme.Fonts.Collection = faces
	theme.Fonts.SystemFonts = false
```

Prepare the collection before opening windows. Shared face data can be reused,
while each window receives its own text shaper. If system fallback is disabled,
include every language and symbol the application needs. See the runnable
[`examples/fonts`](https://github.com/qianniancn/flowui/tree/main/examples/fonts) for Regular-only Source Han Sans SC,
Windows font preference, and the generic subsetting tool.

## Geometry and transitions

```go
ui.Width(200).Height(40)
ui.MinWidth(100).MaxWidth(400)
ui.Padding(12).Margin(8).Radius(8)
ui.BorderWidth(1).BorderColor(ui.TokenBorder)
ui.BorderBottomWidth(1).BorderBottomColor(ui.TokenBorder)
ui.BoxShadow(0, 6, 18, 0, ui.RGBA(0x00000030))
ui.Opacity(0.9)
ui.Overflow(ui.StyleOverflowHidden)
```

Use `Transition` for state-driven properties and give the interactive widget a
stable key:

```go
ui.Background(ui.TokenSurface).
	Transition(ui.PropBackgroundColor, 120*time.Millisecond).
	When(ui.Hovered, ui.Background(ui.TokenSurfaceRaised))
```

Continue with [Animation](12-animation.md).
