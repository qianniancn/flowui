# FlowUI

[中文文档](README.zh-CN.md)

FlowUI is a desktop UI framework for Go, built on [Gio](https://gioui.org/).
It provides a public `ui` package with an MVU application loop,
message-driven updates, controlled components, theme settings, layout
primitives, popups, and test tools.

The public API is:

```go
import "github.com/qianniancn/FlowUI/ui"
```

Application state belongs in your model. Components own interaction state and
derived rendering state such as focus, hover, selection, and animation progress.

## Requirements

- Go 1.26.2 or newer
- A desktop platform supported by Gio

## Install

```bash
go get github.com/qianniancn/FlowUI/ui
```

## Quick Start

Save this as `main.go` in a Go module:

```go
package main

import (
	"fmt"

	"github.com/qianniancn/FlowUI/ui"
)

type Model struct {
	Count int
}

type Msg struct {
	Delta int
}

func Update(model *Model, msg Msg) {
	model.Count += msg.Delta
}

func View(_ *ui.Context, model Model, send ui.Send[Msg]) ui.Widget {
	return ui.Column(
		ui.Text(fmt.Sprintf("Count: %d", model.Count)).Size(20),
		ui.Row(
			ui.Button("decrease", ui.Text("-1")).OnClick(func() {
				send(Msg{Delta: -1})
			}),
			ui.Button("increase", ui.Text("+1")).OnClick(func() {
				send(Msg{Delta: 1})
			}),
		).Gap(8),
	).Gap(12)
}

func main() {
	ui.Run(Model{}, Update, View,
		ui.Title("FlowUI Counter"),
		ui.Size(640, 480),
	)
}
```

Run it with:

```bash
go run .
```

## Application Model

Use the simplest entry point that fits the application:

- `ui.Run` for synchronous updates.
- `ui.RunCmd` when updates return asynchronous commands.
- `ui.RunWithSubscriptions` for commands plus long-lived subscriptions.
- `ui.RunProgram` when startup work or native window-state messages are part of
  the program.

Commands run outside the event loop. Capture immutable values in a command and
send results back through `ui.Send`; do not retain a model pointer or a
`ui.Context`.

Use `ui.LatestCmd` and `ui.CancelLatestCmd` for keyed work such as search or
preview, where an older result should be canceled and ignored.

For multiple windows, create `ui.WindowSpec` values with `ui.NewWindow`,
`ui.NewWindowCmd`, or `ui.NewProgramWindow`, then pass them to
`ui.RunWindows` or `ui.Application.Run`. The `NewWindow` initializers run for
each window instance and must return independent model state.

## Components

The public component API covers common desktop application work:

- **Content:** `Text`, `Label`, `Description`, `Image`, `Avatar`, `Badge`.
- **Forms:** `Input`, `TextArea`, `Checkbox`, `Switch`, `RadioGroup`, `Select`,
  `ComboBox`, date pickers, and color pickers.
- **Navigation and popups:** `Tabs`, `Sidebar`, `Tree`, `Menu`, `Menubar`,
  `Dropdown`, `Popover`, `Tooltip`, `Modal`, `AlertDialog`, and `Toast`.
- **Data and feedback:** `Table`, `Pagination`, `Slider`, `ProgressBar`,
  `ProgressCircle`, `Meter`, `Spinner`, and line, bar, pie, and candlestick
  charts.
- **Layout:** `Surface`, `Card`, `Box`, `Row`, `Column`, `Grid`, `Scroll`,
  `SplitPane`, `Stack`, `Overlay`, and related primitives.

Component styling follows FlowUI theme settings and uses HeroUI conventions for
variants, sizes, states, and interaction feedback where they fit a desktop
control.

## Theme and Locale

Start with `ui.DefaultTheme` or `ui.DarkTheme`. Use `WithTheme` to replace the
theme, or `CustomizeTheme` to change selected settings. In the application from
the previous section:

```go
package main

import (
	"image/color"

	"github.com/qianniancn/FlowUI/ui"
)

func main() {
	ui.Run(Model{}, Update, View,
		ui.CustomizeTheme(func(theme *ui.Theme) {
			theme.Palette.Accent = color.NRGBA{R: 0x00, G: 0x78, B: 0xd4, A: 0xff}
			theme.Components.Button.Radius = 8
			theme.Components.Button.BorderWidth = 2
		}),
		ui.Locale(ui.LanguageEnglish),
	)
}
```

Application theme customization applies to the application or window. Component
instances use reusable `Style` snapshots. Defaults, variants, sizes, inherited
`StyleScope` values, and the instance style are cascaded in that order:

```go
primary := ui.Background(ui.TokenAccent).
	TextColor(ui.TokenAccentForeground).
	Radius(8).
	Cursor(ui.CursorPointer).
	BoxShadow(0, 6, 18, 0, ui.RGBA(0x00000030)).
	When(ui.Hovered, ui.Background(ui.TokenAccentHover)).
	When(ui.Pressed, ui.Background(ui.TokenAccentPressed).
		Scale(0.96, 0.96))

ui.StyleScope(
	ui.FontSize(14),
	ui.Button("save", ui.Text("Save")).Style(primary),
)
```

`When` receives runtime states such as `Hovered`, `Pressed`, `Focused`,
`Disabled`, `Selected`, and `Invalid`. `StyleScope` applies to descendants, and
the instance `Style` has the final precedence. Theme color tokens are resolved
at layout time, so the same declaration follows light or dark application themes.
MVU model values use the same path with `When(ui.If(model.Highlighted), ...)`;
the View rebuilds the declaration when the model changes.

`RGB` uses packed `0xRRGGBB`, `RGBA` uses packed `0xRRGGBBAA`, `Color` accepts
any standard-library `color.Color`, and `WithAlpha` changes the alpha of either
a concrete color or a theme token. Geometry includes fixed/min/max sizes,
fill, margin, padding, overflow clipping, cursor, and aspect ratio:

```go
square := ui.Width(40).AspectRatio(1).Background(ui.RGBA(0x9333eacc))
```

Root properties always style the component's outer box. Compound internals use
named parts. The built-in parts are `PartContent`, `PartLabel`,
`PartDescription`, `PartIcon`, `PartTrack`, `PartFill`, `PartThumb`,
`PartIndicator`, `PartPanel`, `PartItem`, `PartBackdrop`, `PartPlaceholder`,
`PartSelection`, `PartPrefix`, and `PartSuffix`:

```go
barStyle := ui.Background(ui.RGBA(0x111827cc)). // outer component
	Part(ui.PartTrack, ui.Height(6).Background(ui.TokenSurfaceRaised)).
	Part(ui.PartFill, ui.Background(ui.TokenAccent)).
	Part(ui.PartLabel, ui.TextColor(ui.TokenMutedForeground))

ui.ProgressBar("upload", 42).Label("Upload").Style(barStyle)
```

For compound field controls such as Select, ComboBox, and the date controls,
`PartContent` is the styled field face; the root remains the outer component.

Custom components resolve the same cascade and reuse the same renderer with
`ResolveStyle`, `ResolveStylePart`, `LayoutResolvedStyle`, and
`LayoutInteractiveResolvedStyle`; see `examples/custom_widgets`.

`Surface` uses the same style API for per-instance geometry and paint:

```go
ui.Surface(content).Style(ui.Radius(12).
	BorderWidth(1).
	BorderColor(ui.RGB(0x9333ea)))
```

Transitions need stable identity. Interactive components already have one;
wrap each non-interactive transitioning sibling in a distinct `ui.Key` scope.
A `Box` may use its own `.Key(...)` instead.

Shadow geometry is configured at the application theme level. Profiles contain
three layers ordered from tightest to broadest; a layer with zero opacity is
disabled:

```go
ui.Run(Model{}, Update, View, ui.CustomizeTheme(func(theme *ui.Theme) {
	theme.Palette.SurfaceShadow = color.NRGBA{R: 0x93, G: 0x33, B: 0xea, A: 0xff}
	theme.Shadows.Surface.Layers = [ui.ShadowLayerCount]ui.ShadowLayerTheme{
		{OffsetY: 2, Blur: 4, Opacity: 0.65},
		{OffsetY: 7, Blur: 16, Spread: 2, Opacity: 0.4},
		{OffsetY: 16, Blur: 36, Spread: 6, Opacity: 0.3},
	}
}))
```

`Layers[0]`, `Layers[1]`, and `Layers[2]` are the near, middle, and far
layers respectively. Each layer controls its offset, blur, spread, and opacity.
The available profiles are `Surface`, `Overlay`, `Menu`, `Control`,
`Checkbox`, and `SwitchThumb`. Profiles control geometry and per-layer opacity;
components still provide the base color through settings such as
`Palette.SurfaceShadow` and `Components.Menu.ShadowColor`.

FlowUI includes English and Chinese component strings. `ui.LanguageAuto`
selects the host language. An application that owns its windows can change a
window at runtime with `Application.SetTheme` and `Application.SetLanguage`.

## Examples

Runnable programs are in [`examples/`](examples/). Start with:

```bash
go run ./examples/components
go run ./examples/counter
go run ./examples/async
go run ./examples/tables
go run ./examples/multi_windows
```

The directory also contains focused examples for forms, navigation, popups,
charts, custom widgets, window title bars, and layout primitives.

## Testing

`uitest` provides deterministic helpers for frame, input, time, and application
tests. It is a test helper and is not required by applications at runtime.

```bash
go test ./...
go vet ./...
```

## Project Structure

- `ui/` is the application-facing package.
- `internal/components/` contains component implementations.
- `internal/frame/`, `internal/state/`, `internal/theme/`, and related packages
  provide shared implementation services.
- `uitest/` contains deterministic test harnesses.
- [`docs/architecture.md`](docs/architecture.md) describes dependency direction,
  state ownership, and overlay behavior.

Applications should import `ui`, not `internal` packages. The API is still
evolving; the examples and public package documentation are the current usage
reference.

## License

FlowUI is licensed under the [MIT License](LICENSE).

## Contributing

See [`CONTRIBUTING.md`](CONTRIBUTING.md) for development, testing, and commit
guidelines.
