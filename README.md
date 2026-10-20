# FlowUI

[简体中文](README.zh-CN.md)

[![Go Reference](https://pkg.go.dev/badge/github.com/qianniancn/flowui/ui.svg)](https://pkg.go.dev/github.com/qianniancn/flowui/ui)
[![Go version](https://img.shields.io/github/go-mod/go-version/qianniancn/flowui)](go.mod)
[![License](https://img.shields.io/github/license/qianniancn/flowui)](LICENSE)

[English documentation](https://qianniancn.github.io/flowui/en/) · [中文文档](https://qianniancn.github.io/flowui/)

FlowUI is a desktop UI framework for Go, built on [Gio](https://gioui.org/).
It combines a typed MVU runtime with official widgets, layout primitives,
declarative styles, window management, asynchronous effects, and deterministic
UI testing through focused public packages.

```go
import (
	"github.com/qianniancn/flowui/layout"
	"github.com/qianniancn/flowui/style"
	"github.com/qianniancn/flowui/theme/base"
	"github.com/qianniancn/flowui/ui"
	"github.com/qianniancn/flowui/widget/button"
	"github.com/qianniancn/flowui/window"
)
```

![FlowUI component showcase](docs/assets/showcase-light.png)

## Highlights

- **Typed MVU:** keep business state in a model, describe changes with messages,
  and render a read-only view.
- **Desktop components:** forms, navigation, overlays, data display, charts,
  window chrome, and layout primitives.
- **Declarative styling:** theme tokens, light and dark themes, component parts,
  runtime states, transitions, and scoped styles.
- **Application runtime:** commands, subscriptions, multiple windows, retained
  models, and localization.
- **Platform services:** optional packages provide native file dialogs, desktop
  notifications, and system tray integration.
- **Testable by design:** deterministic frame, input, time, and application
  helpers are available in `uitest`.

## Requirements

- Go 1.26.2 or newer
- A desktop platform supported by Gio

## Install

```bash
go get github.com/qianniancn/flowui/ui
```

## Quick Start

Create `main.go` inside a Go module:

```go
package main

import (
	"fmt"

	"github.com/qianniancn/flowui/layout"
	"github.com/qianniancn/flowui/theme/base"
	"github.com/qianniancn/flowui/ui"
	"github.com/qianniancn/flowui/widget/button"
	"github.com/qianniancn/flowui/widget/text"
	"github.com/qianniancn/flowui/window"
)

type Model struct {
	Count int
}

type Msg interface{ msg() }

type Inc struct{}
type Dec struct{}

func (Inc) msg() {}
func (Dec) msg() {}

func Update(model *Model, msg Msg) ui.Cmd[Msg] {
	switch msg.(type) {
	case Inc:
		model.Count++
	case Dec:
		model.Count--
	}
	return nil
}

func View(_ *ui.Context, model Model, send ui.Send[Msg]) ui.Widget {
	return layout.Center(
		layout.Column(
			text.New("FlowUI Counter").Size(24),
			text.New(fmt.Sprintf("Count: %d", model.Count)),
			layout.Row(
				button.New("-1").OnClick(func() {
					send(Dec{})
				}),
				button.New("+1").OnClick(func() {
					send(Inc{})
				}),
			).Gap(8),
		).Gap(12),
	)
}

func main() {
	ui.Run(ui.NewProgram(Model{}, Update, View),
		window.Theme(base.Light()),
		window.Title("FlowUI Counter"),
		window.Size(640, 480),
	)
}
```

Run it with:

```bash
go run .
```

For a guided walkthrough, read the [English Quick Start](https://qianniancn.github.io/flowui/en/guide/01-quick-start/).

## Core Model

FlowUI follows three ownership rules:

1. **Model, messages, and Update own business state.** View reads a model value
   and sends typed messages instead of mutating captured state.
2. **Style owns a widget's box and appearance.** Layout containers own how
   children are measured and positioned.
3. **Keys provide identity across frames.** Stable keys retain interaction and
   animation state for repeated or moving widgets.

Choose the application API that matches the required lifecycle:

| API | Use it for |
| --- | --- |
| `ui.Run(ui.Program)` | The single-window entry point for every MVU program |
| `ui.NewProgram` | A compact program declaration for a fixed initial model |
| `ui.Application` | Multiple windows, tray integration, and application-owned lifecycles |

Commands run outside the event loop and return results through `ui.Send`.
Subscriptions represent long-lived inputs such as timers or external event
streams. The [MVU guide](https://qianniancn.github.io/flowui/en/guide/03-mvu-and-messages/)
and [effects guide](https://qianniancn.github.io/flowui/en/guide/08-commands-and-subscriptions/)
cover the complete contracts.

## Components

| Area | Included APIs |
| --- | --- |
| Content | `text.New`, `label.New`, `description.New`, `image.New`, `avatar.New`, `badge.New`, `chip.New` |
| Actions | `button.New`, `button.Group`, `togglebutton.New`, `closebutton.New` |
| Forms | `input.New`, `input.NewTextArea`, `checkbox.New`, `switchcontrol.New`, `radiogroup.New`, `selectfield.New`, `combobox.New`, date and color pickers |
| Navigation | `tabs.New`, `sidebar.New`, `tree.New`, `menu.New`, `menubar.New`, `pagination.New`, `toolbar.New` |
| Overlays | `dropdown.New`, `menu.ContextMenu`, `popover.New`, `tooltip.New`, `modal.New`, `alertdialog.New`, `toast.NewItem` |
| Data and feedback | `table.New`, `progress.Bar`, `progress.Circle`, `progress.Meter`, `spinner.New`, `slider.New`, charts, `heatmap.New`, `gantt.New` |
| Layout | `layout.Box`, `surface.New`, `card.New`, `layout.Row`, `layout.Column`, `layout.Grid`, `layout.Scroll`, `layout.SplitPane`, `layout.Stack`, `layout.Overlay` |

See the [component guide](https://qianniancn.github.io/flowui/en/guide/06-components/)
for the complete widget catalog.

## Styling and Themes

Styles are immutable declarations that draw colors from the active theme and
respond to runtime state. The same declaration follows the active light or
dark theme:

```go
activeTheme := base.Light()

primary := style.New().
	Background(style.Solid(activeTheme.Palette.Accent)).
	TextColor(style.Solid(activeTheme.Palette.AccentForeground)).
	Radius(8).
	Cursor(style.CursorPointer).
	On(style.Hovered, style.New().Background(style.Solid(activeTheme.Palette.AccentHover))).
	On(style.Pressed, style.New().Background(style.Solid(activeTheme.Palette.AccentPressed)).Scale(0.96))

save := button.New("Save").Style(primary)
```

The runtime requires an explicit complete Style Provider. Pass
`window.Theme(base.Light())`, `window.Theme(base.Dark())`, or a theme from
another provider package; use `window.CustomizeTheme` after selecting one for
focused changes. Compound controls expose named
parts such as `style.PartContent`, `style.PartTrack`, `style.PartIndicator`,
and `style.PartPanel` for focused customization. FlowUI
includes English and Chinese component strings; `window.LanguageAuto` follows the
host language.

The [style and theme guide](https://qianniancn.github.io/flowui/en/guide/05-styling-and-themes/)
documents precedence, parts, colors, geometry, transitions, and bundled font setup.

FlowUI uses system font fallback by default. For reproducible rendering, parse an
embedded TTF, OTF, or TTC with `style.ParseFontCollection`, assign the faces to
`theme.Fonts.Collection`, and set `theme.Fonts.SystemFonts = false`.

The core maintains one default design system in `theme/base`, with light and dark
modes. Applications can use `base.Light()` / `base.Dark()` directly or build an
external Provider on top of the public theme APIs.

## Public Packages

| Package | Responsibility |
| --- | --- |
| `ui` | MVU program, context, widget contract, and application lifecycle |
| `widget/<component>` | Official fully-rendered components |
| `layout` | Layout containers and the common box host |
| `style` | Immutable declarations, cascade, and tokens |
| `theme/base` | The default light/dark theme and complete Provider |
| `theme` | Theme values, visual contracts, and extension APIs |
| `draw` | Low-level brushes, borders, shadows, paths, and transforms |
| `animation` | Reusable tween, spring, and timeline primitives |
| `window` | Native window options and locale configuration |
| `uitest` | Deterministic widget and application test harnesses |

Start with the [interactive component showcase](examples/showcase/README.md):
`go run ./examples/showcase`. Its 65 pages cover every public widget package,
with search, live demos, usage snippets, and light/dark modes.

Focused default-theme examples also live in this module: `go run ./examples/counter`,
`go run ./examples/buttons`, `go run ./examples/inputs`,
`go run ./examples/style-gallery`, and `go run ./examples/close_guard` for a
native close-request guard.

## Documentation

| Resource | Purpose |
| --- | --- |
| [Documentation](https://qianniancn.github.io/flowui/en/) | English task-oriented guide from first app through advanced features |
| [中文文档](https://qianniancn.github.io/flowui/) | 中文教程和组件说明 |
| [Package reference](https://pkg.go.dev/github.com/qianniancn/flowui/ui) | Public Go API |
| [`docs/architecture.md`](docs/architecture.md) | Dependency direction, state ownership, overlays, and effects |
| [`explorer/README.md`](explorer/README.md) | Per-window native open and save dialogs |
| [`notify/README.md`](notify/README.md) | Cross-platform native notifications |
| [`systray/README.md`](systray/README.md) | Cross-platform system tray lifecycle and native menus |
| [`CONTRIBUTING.md`](CONTRIBUTING.md) | Development workflow and contribution rules |

Applications use `ui` for MVU and `widget/<component>`, `layout`, `style`, `draw`, and
`window` according to responsibility. `explorer`, `notify`, and `systray` are
optional platform services.
Applications must not import packages under `internal`. The repository root
intentionally contains no Go package.

## Testing

Run the complete project checks from the repository root:

```bash
go test ./...
go vet ./...
```

`uitest` is intended for component and application tests; it is not required
by applications at runtime. See the [testing guide](https://qianniancn.github.io/flowui/en/guide/13-testing/).

## Contributing

Contributions are welcome. Read [`CONTRIBUTING.md`](CONTRIBUTING.md) before
submitting a change.

## License

FlowUI is available under the [MIT License](LICENSE).
