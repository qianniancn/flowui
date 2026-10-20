# 05 - Styling and themes

`style.Style` is an immutable declaration of box geometry and appearance. Start
with its zero value and chain the properties you need:

```go
primary := style.New().Background(style.Solid(activeTheme.Palette.Accent)).
	TextColor(style.Solid(activeTheme.Palette.AccentForeground)).
	Radius(8).
	PaddingX(12).
	PaddingY(6).
	Cursor(style.CursorPointer)
```

Apply it to a widget:

```go
button.New("Save").Style(primary)
layout.Box(child).Style(style.New().Padding(24))
```

There is no separate builder step. Each method returns a new style value.

## Cascade order

The effective style is resolved in this order:

```text
component defaults
  -> inherited text
  -> variant
  -> size
  -> resolve.Scope ancestors
  -> instance Style
  -> state conditions and transitions
```

Later declarations at the same layer replace earlier values. Use `On` with
variant masks for runtime states:

```go
primary := style.New().Background(style.Solid(activeTheme.Palette.Accent)).
	TextColor(style.Solid(activeTheme.Palette.AccentForeground)).
	On(style.Hovered, style.New().Background(style.Solid(activeTheme.Palette.AccentHover))).
	On(style.Pressed, style.New().Background(style.Solid(activeTheme.Palette.AccentPressed))).
	On(style.Disabled, style.New().Opacity(0.5))
```

`resolve.Scope` supplies defaults to a subtree without changing the global theme.
It lives in `github.com/qianniancn/flowui/style/resolve`, together with the other
context-bound resolvers — `style` itself is a pure declaration mechanism, so
anything that needs the active context or theme values sits beside it.
`Part` targets a named part inside a compound control, such as
`PartContent`, `PartTrack`, or `PartPanel`.

## Action presets

The theme ships four action presets — variant colors plus hover/pressed/disabled
rows in one call, geometry left to the adopting component:

```go
// One call for the primary action look (accent colors, state rows, color transition).
button.New("Save").Style(activeTheme.PrimaryAction())

// Presets are values: chain overrides or extra rows freely.
activeTheme.PrimaryAction().On(style.Loading, style.New().Opacity(0.7))
```

The four are `PrimaryAction` (accent filled), `SecondaryAction` (soft surface),
`DangerAction` (destructive filled), and `GhostAction` (transparent quiet).
They use the theme's semantic palette; individual Providers can give button
variants a different treatment. Presets carry no radius, height, padding, or
focus ring — those belong to the component that adopts them.

## Tokens and colors

Prefer semantic theme tokens so the same style works in light and dark themes:

```go
style.New().Background(style.Solid(activeTheme.Palette.Surface))
style.New().TextColor(style.Solid(activeTheme.Palette.Foreground))
style.New().Background(style.Solid(activeTheme.Palette.Accent))
```

Use literal colors when a design requires them:

```go
style.RGB(0x0078d4)          // opaque 0xRRGGBB
style.RGBA(0x00000030)       // 0xRRGGBBAA
style.WithAlpha(activeTheme.Palette.Focus, 0.5)
```

## Gradients and custom drawing

When a custom widget needs a gradient, resolve it against the active theme and
reuse the public drawing helpers:

```go
brush, ok := resolve.Brush(style.LinearGradient(
	style.ColorStop(0, style.Solid(activeTheme.Palette.Accent)),
	style.ColorStop(1, style.Solid(activeTheme.Palette.Danger)),
))
if ok {
	draw.Fill(gtx, image.Rect(0, 0, 160, 32), 6, brush)
}
```

Use `resolve.Color` for a single theme-resolved color and `draw.FillRRect` when
you already have a `clip.RRect`. This keeps theme resolution in one place and
does not turn custom drawing into an overlay system.

## Themes

Import `github.com/qianniancn/flowui/theme/base` for the default theme.
Use `base.Light()` / `base.Dark()` directly, or customize its seed and metrics:

The default design uses Slate greys, Indigo accents, 1dp borders and 32dp
standard controls. See [Default theme and external styles](18-design-styles.md)
for screenshots and geometry. `Light` / `Dark` supply the default Indigo accent;
`New` / `NewDark` use the supplied seed colours.

```go
seed := theme.DefaultSeed()
seed.Primary = color.NRGBA{R: 0x00, G: 0x78, B: 0xd4, A: 0xff}

ui.Run(ui.NewProgram(Model{}, Update, View),
	window.Theme(base.NewDark(seed)),
	window.CustomizeTheme(func(t *theme.Theme) {
		t.Components.Input.Radius = 8
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

faces, err := style.ParseFontCollection(interRegular)
if err != nil {
	panic(err)
}

t := base.Light()
t.Typography.Typeface = "Inter"
t.Fonts.Collection = faces
t.Fonts.SystemFonts = false
```

Prepare the collection before opening windows. Shared face data can be reused,
while each window receives its own text shaper. If system fallback is disabled,
include every language and symbol the application needs. Supply the font file
referenced by the snippet in your own application.

## Geometry and transitions

```go
style.New().Width(200).Height(40)
style.New().MinWidth(100).MaxWidth(400)
style.New().Padding(12).Margin(8).Radius(8)
style.New().BorderWidth(1).BorderColor(style.Solid(activeTheme.Palette.Border))
style.New().BorderBottom(1, style.Solid(activeTheme.Palette.Border))
style.New().BoxShadow(0, 6, 18, 0, style.RGBA(0x00000030))
style.New().Opacity(0.9)
style.New().Overflow(style.OverflowHidden)
```

Use `Transition` for state-driven properties and give the interactive widget a
stable key:

```go
style.New().Background(style.Solid(activeTheme.Palette.Surface)).
	Transition(style.PropBackgroundColor, 120*time.Millisecond).
	On(style.Hovered, style.New().Background(style.Solid(activeTheme.Palette.SurfaceRaised)))
```

Continue with [Animation](12-animation.md).

## Instance styles vs. theme appearance

`.Style(...)` is an **instance-level** override. Without it, a widget's default look comes from the theme provider (`theme/<x>style` recipes plus `Components.*` tokens); behavior and layout live in `widget/*` and never own concrete colors.

```go
// Instance override: affects this button only
button.New("Save").Style(primary).Variant(button.Primary)

// Theme appearance: swap the provider to reskin everything; widget code is untouched
type Brand struct{ base.Provider }

func (Brand) Button(active *theme.Theme, request theme.ButtonRequest) theme.ButtonVisual {
	visual := buttonstyle.BuildButton(active, request, buttonstyle.DefaultConfig())
	visual.Variant = visual.Variant.Radius(10)
	return visual
}

active := base.Light()
active.VisualProvider = Brand{}
```

Rule of thumb: values that vary per instance go through `.Style(...)`; values that one theme applies to every instance go through the provider recipe or `Components.*` tokens.
