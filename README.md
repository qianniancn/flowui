# FlowUI

[简体中文](README.zh-CN.md)

[![Go Reference](https://pkg.go.dev/badge/github.com/qianniancn/flowui/ui.svg)](https://pkg.go.dev/github.com/qianniancn/flowui/ui)
[![Go version](https://img.shields.io/github/go-mod/go-version/qianniancn/flowui)](go.mod)
[![License](https://img.shields.io/github/license/qianniancn/flowui)](LICENSE)

FlowUI is a desktop UI framework for Go, built on [Gio](https://gioui.org/).
It combines a typed MVU application model with a broad component set,
declarative styling, window management, asynchronous effects, and deterministic
UI testing behind a focused public API.

```go
import "github.com/qianniancn/flowui/ui"
```

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

	"github.com/qianniancn/flowui/ui"
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
	return ui.Center(
		ui.Column(
			ui.Text("FlowUI Counter").Size(24),
			ui.Text(fmt.Sprintf("Count: %d", model.Count)),
			ui.Row(
				ui.Button("decrement", ui.Text("-1")).OnClick(func() {
					send(Dec{})
				}),
				ui.Button("increment", ui.Text("+1")).OnClick(func() {
					send(Inc{})
				}),
			).Gap(8),
		).Gap(12),
	)
}

func main() {
	ui.Run(ui.NewProgram(Model{}, Update, View),
		ui.Title("FlowUI Counter"),
		ui.Size(640, 480),
	)
}
```

Run it with:

```bash
go run .
```

The same application is available at [`examples/counter`](examples/counter).
For a guided walkthrough, read [Quick Start](https://qianniancn.github.io/flowui/guide/01-快速开始/).

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
streams. The [MVU guide](https://qianniancn.github.io/flowui/guide/03-MVU与消息/)
and [effects guide](https://qianniancn.github.io/flowui/guide/08-命令与订阅/)
cover the complete contracts.

## Components

| Area | Included APIs |
| --- | --- |
| Content | `Text`, `Label`, `Description`, `Image`, `Avatar`, `Badge`, `Chip` |
| Actions | `Button`, `ButtonGroup`, `ToggleButton`, `CloseButton`, `Action` |
| Forms | `Input`, `TextArea`, `Checkbox`, `Switch`, `RadioGroup`, `Select`, `ComboBox`, date and color pickers |
| Navigation | `Tabs`, `Sidebar`, `Tree`, `Menu`, `Menubar`, `Pagination`, `Toolbar` |
| Overlays | `Dropdown`, `ContextMenu`, `Popover`, `Tooltip`, `Modal`, `AlertDialog`, `Toast` |
| Data and feedback | `Table`, `ProgressBar`, `ProgressCircle`, `Meter`, `Spinner`, `Slider`, charts, `Heatmap`, `GanttChart` |
| Layout | `Box`, `Surface`, `Card`, `Row`, `Column`, `Grid`, `Scroll`, `SplitPane`, `Stack`, `Overlay` |

See the [component guide](https://qianniancn.github.io/flowui/guide/06-组件一览/)
or run the component gallery:

```bash
go run ./examples/components
```

## Styling and Themes

Styles are immutable declarations that can use theme tokens and respond to
runtime state. The same declaration follows the active light or dark theme:

```go
primary := ui.Background(ui.TokenAccent).
	TextColor(ui.TokenAccentForeground).
	Radius(8).
	Cursor(ui.CursorPointer).
	When(ui.Hovered, ui.Background(ui.TokenAccentHover)).
	When(ui.Pressed, ui.Background(ui.TokenAccentPressed).Scale(0.96, 0.96))

save := ui.Button("save", ui.Text("Save")).Style(primary)
```

The runtime starts with `ui.DefaultTheme()`. Pass
`ui.WithTheme(ui.DarkTheme())` to replace it, or use `ui.CustomizeTheme` for
focused changes. Compound controls expose named parts such as `PartContent`,
`PartTrack`, `PartIndicator`, and `PartPanel` for focused customization. FlowUI
includes English and Chinese component strings; `ui.LanguageAuto` follows the
host language.

The [style and theme guide](https://qianniancn.github.io/flowui/guide/05-样式与主题/)
documents precedence, parts, colors, geometry, and transitions.

## Examples

Every directory below is a runnable program:

| Example | Demonstrates |
| --- | --- |
| [`examples/counter`](examples/counter) | Minimal typed MVU application |
| [`examples/form`](examples/form) | Form controls and validation |
| [`examples/async`](examples/async) | Commands and asynchronous results |
| [`examples/components`](examples/components) | Component gallery |
| [`examples/custom_widgets`](examples/custom_widgets) | Custom composites and canvas widgets |
| [`examples/multi_windows`](examples/multi_windows) | Application-owned multiple windows |
| [`examples/file_dialogs`](examples/file_dialogs) | Native open and save file dialogs |
| [`examples/notifications`](examples/notifications) | Native desktop notifications |
| [`examples/systray_ui`](examples/systray_ui) | FlowUI window with a native system tray |

Additional focused examples under [`examples/`](examples/) cover charts,
animations, menus, overlays, layout, window chrome, and individual controls.

## Documentation

| Resource | Purpose |
| --- | --- |
| [Documentation](https://qianniancn.github.io/flowui/) | Task-oriented guide from first app through advanced features (Chinese) |
| [Package reference](https://pkg.go.dev/github.com/qianniancn/flowui/ui) | Public Go API |
| [`docs/architecture.md`](docs/architecture.md) | Dependency direction, state ownership, overlays, and effects |
| [`explorer/README.md`](explorer/README.md) | Per-window native open and save dialogs |
| [`notify/README.md`](notify/README.md) | Cross-platform native notifications |
| [`systray/README.md`](systray/README.md) | Cross-platform system tray lifecycle and native menus |
| [`CONTRIBUTING.md`](CONTRIBUTING.md) | Development workflow and contribution rules |

Applications use `github.com/qianniancn/flowui/ui` for the interface and MVU
runtime. `explorer`, `notify`, and `systray` are optional platform services.
Applications must not import packages under `internal`. The repository root
intentionally contains no Go package.

## Testing

Run the complete project checks from the repository root:

```bash
go test ./...
go vet ./...
```

`uitest` is intended for component and application tests; it is not required
by applications at runtime. See the [testing guide](https://qianniancn.github.io/flowui/guide/13-测试/).

## Contributing

Contributions are welcome. Read [`CONTRIBUTING.md`](CONTRIBUTING.md) before
submitting a change.

## License

FlowUI is available under the [MIT License](LICENSE).
