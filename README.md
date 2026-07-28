# FlowUI

[中文文档](README.zh-CN.md)

FlowUI is a desktop UI framework for Go, built on [Gio](https://gioui.org/).
It provides a public `ui` package with an MVU application loop,
message-driven updates, controlled components, theme settings, layout
primitives, popups, and test tools.

The public API is:

```go
import "github.com/qianniancn/flowui/ui"
```

Three teaching rules:

1. **Model + messages + Update** own business data; View is read-only and
   expresses intent with `send`.
2. **Style** owns a box's size and appearance; **layout containers** own how
   children are arranged.
3. **Keys** are explicit identity across frames; without a Key, interaction and
   animation slots are not retained.

### Documentation

| Audience | Start here |
|----------|------------|
| **App developers** | [User tutorial (Chinese)](https://github.com/qianniancn/flowui/wiki) — getting started through custom widgets, multi-window, animation, and FAQ |
| **Architecture notes** | [`docs/architecture.md`](docs/architecture.md) |

The user tutorial is maintained in the project Wiki.

## Requirements

- Go 1.26.2 or newer
- A desktop platform supported by Gio

## Install

```bash
go get github.com/qianniancn/flowui/ui
```

## Quick Start

Save this as `main.go` in a Go module:

```go
package main

import (
	"fmt"

	"github.com/qianniancn/flowui/ui"
)

type Model struct {
	Count int
}

// Prefer a closed message set (not `type Msg any`).
type Msg interface{ msg() }

type Inc struct{}
type Dec struct{}

func (Inc) msg() {}
func (Dec) msg() {}

func Update(m *Model, msg Msg) {
	switch msg.(type) {
	case Inc:
		m.Count++
	case Dec:
		m.Count--
	}
}

func View(_ *ui.Context, m Model, send ui.Send[Msg]) ui.Widget {
	return ui.Center(
		ui.Box(
			ui.Column(
				ui.Text("FlowUI Counter").Size(24),
				ui.Text(fmt.Sprintf("count: %d", m.Count)).Size(18),
				ui.Row(
					ui.Button("inc", ui.Text("+1")).OnClick(func() {
						send(Inc{})
					}),
					ui.Button("dec", ui.Text("-1")).
						Variant(ui.ButtonSecondary).
						OnClick(func() {
							send(Dec{})
						}),
				).Gap(12),
			).Gap(12),
		).Style(ui.Padding(24)),
	)
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

Same program in the repo: `go run ./examples/counter`.
Step-by-step guide (Chinese): [Quick Start](https://github.com/qianniancn/flowui/wiki/01-快速开始).

## Application Model

Primary entry points:

- `ui.Run` — synchronous `Update` (most apps start here).
- `ui.RunProgram` — full `Program` (`Init`, `UpdateCmd`, subscriptions,
  window-state messages).
- `ui.Application` — multi-window host.

For commands or subscriptions, use `RunProgram` with `Program{Update, Subscriptions, View, ...}`.

### Messages

Prefer a **closed message set**, not `type Msg any`:

```go
type Msg interface{ msg() }

type Inc struct{}
type Dec struct{}

func (Inc) msg() {}
func (Dec) msg() {}

func Update(m *Model, msg Msg) {
	switch msg.(type) {
	case Inc:
		// ...
	case Dec:
		// ...
	}
}
```

Nested modules map child messages into the parent set with `MapCmd` /
`MapSubscription` (see `examples/modules`).

Reusable menus/toolbars/shortcuts use **`Action`** (`NewAction`, `ActionScope`,
`ActionMenuItem`, `ActionButton`). Asynchronous MVU side effects remain **`Cmd`**.

### Open state

Two families of open/close controls with different contracts:

**Method family** — `Select`, `Dropdown`, `ContextMenu`, `Menubar`. Uncontrolled
by default; opt into control with methods:

```text
no Open(...)     → uncontrolled; widget keeps open state
Open(bool)       → controlled; handle OnOpenChange and store open in Model
DefaultOpen(...) → seeds uncontrolled open only
```

(`Menubar` keys the open menu by string: `OpenKey` / `DefaultOpenKey`.)

**Constructor family** — `Popover`, `Modal`, `AlertDialog`, `Collapsible`. Always
controlled: pass the open/expanded value to the constructor and store it in the
Model; dismiss requests arrive through `OnOpenChange` (`OnExpandedChange` for
`Collapsible`). There is no uncontrolled mode and no `Open`/`DefaultOpen` method.

```go
Modal("confirm", m.ConfirmOpen, "Delete?", body).
    OnOpenChange(func(open bool) { send(SetConfirm{Open: open}) })
```

Commands run outside the event loop. Capture immutable values in a command and
send results back through `ui.Send`; do not retain a model pointer or a
`ui.Context`.

Use `ui.Batch` when one update must start several independent commands; do not
spawn ad-hoc goroutine farms inside a command. Use `ui.LatestCmd` and
`ui.CancelLatestCmd` for keyed work such as search or preview, where an older
result should be canceled and ignored. Use a stable, bounded workflow key (not
a per-request value), because keys are retained for the window lifetime.

For multiple windows, build `ui.WindowSpec` values (for example with
`ui.NewProgramWindow`) and pass them to `ui.Application.Run` or
`ui.RunWindows`. Initializers run per window instance and must return
independent model state. `Application.Open` returns `true` only when a new
window starts. If the same key is already open, FlowUI raises that window and
returns `false`; rejected opens also return `false`.

Deeper walkthroughs: [MVU & messages](https://github.com/qianniancn/flowui/wiki/03-MVU与消息),
[commands & subscriptions](https://github.com/qianniancn/flowui/wiki/08-命令与订阅),
[multi-window](https://github.com/qianniancn/flowui/wiki/11-多窗口) (Chinese tutorial).

## Components

The public component API covers common desktop application work:

- **Content:** `Text`, `Label`, `Description`, `Image`, `Avatar`, `Badge`.
- **Forms:** `Input`, `TextArea`, `Checkbox`, `Switch`, `RadioGroup`, `Select`,
  `ComboBox`, date pickers, and color pickers.
- **Navigation and popups:** `Tabs`, `Sidebar`, `Tree`, `Menu`, `Menubar`,
  `Dropdown`, `Popover`, `Tooltip`, `Modal`, `AlertDialog`, and `Toast`.
- **Data and feedback:** `Table`, `Pagination`, `Slider`, `ProgressBar`,
  `ProgressCircle`, `Meter`, `Spinner`, and line, bar, pie, and candlestick
  charts, `Heatmap`, and `GanttChart`.
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

	"github.com/qianniancn/flowui/ui"
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
instances use reusable `Style` snapshots. Layers resolve in this order
(defaults, inherited text, variant, size, `StyleScope`, then instance):

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

Custom composites follow path B: `BeginInteract` → `Resolve` / `ResolvePart`
→ `LayoutInteractiveBox` (or `LayoutBox`). Focus helpers on `Context`
(`RequestFocus`, `RequestFocusVisible`, `FocusVisible`) use the
`Interact.Clickable` tag so custom rings match official controls. Domain
pixels use path C: `ui.Canvas` as the host child (no chrome painting). See
`examples/custom_widgets` and the [custom widgets tutorial](https://github.com/qianniancn/flowui/wiki/10-自定义组件).

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
go run ./examples/counter
go run ./examples/form
go run ./examples/async
go run ./examples/custom_widgets
go run ./examples/multi_windows
go run ./examples/components
```

The directory also contains focused examples for forms, navigation, popups,
charts, animations, window chrome, and layout primitives.

## Testing

`uitest` provides deterministic helpers for frame, input, time, and application
tests. It is a test helper and is not required by applications at runtime.
Tutorial: [Testing](https://github.com/qianniancn/flowui/wiki/13-测试).

```bash
go test ./...
go vet ./...
```

## Project Structure

- `ui/` — application-facing public package (import this only).
- `internal/` — runtime, style, host, interact, components, theme, …
- `uitest/` — deterministic test harnesses.
- `examples/` — runnable demos aligned with the final API.
- [Project Wiki](https://github.com/qianniancn/flowui/wiki) — **user tutorial** (Chinese).
- [`docs/architecture.md`](docs/architecture.md) — dependency direction, state
  ownership, overlay behavior.

Applications should import `ui`, not `internal` packages. For day-to-day usage,
prefer the [user tutorial](https://github.com/qianniancn/flowui/wiki), `examples/`, and
`go doc github.com/qianniancn/flowui/ui`.

## License

FlowUI is licensed under the [MIT License](LICENSE).

## Contributing

See [`CONTRIBUTING.md`](CONTRIBUTING.md) for development, testing, and commit
guidelines.
