# FlowUI

[中文](README.zh-CN.md)

FlowUI is a small MVU UI framework built on top of Gio.

It keeps application state in a `Model`, sends typed messages from the view, and
makes `Update` the only place that mutates state. The API is intentionally
Go-first: widgets are regular values, options are chainable methods, and local
widget state is managed by `Context` with explicit keys.

## Features

- MVU architecture: `Model`, `Msg`, `Update`, and `View`.
- Typed message dispatch through `ui.Send[Msg]`.
- Optional command effects with `RunCmd`, `Cmd`, and `Do`.
- Gio widget state managed by `Context`.
- Key-scoped local state with automatic cleanup after each frame.
- HeroUI-inspired controls with variants, sizes, disabled states, validation
  states, loading states, focus behavior, and basic animations.
- Date picker labels with system-language detection, built-in English and
  Chinese locales, and day/month/year selection views.
- FlowUI-native theme tokens for palette, typography, shape, spacing, and
  component styles, with Gio material theme kept as a bridge.
- Layout primitives for fixed, flexible, responsive, scrollable, stacked, and
  grid-based interfaces.
- Overlay and popup surfaces with a cached soft-shadow renderer.
- Window title, window size, locale, and theme options.

## Basic Usage

```go
package main

import ui "github.com/qianniancn/FlowUI"

type Model struct {
	Name string
}

type Msg struct {
	Name string
}

func Update(m *Model, msg Msg) {
	m.Name = msg.Name
}

func View(_ *ui.Context, m Model, send ui.Send[Msg]) ui.Widget {
	return ui.Center(
		ui.Box(
			ui.Column(
				ui.Text("FlowUI").Size(24),
				ui.Input("name", m.Name).
					Hint("Name").
					OnChange(func(text string) {
						send(Msg{Name: text})
					}),
			).Gap(12),
		).Width(320).Padding(24),
	)
}

func main() {
	ui.Run(Model{}, Update, View, ui.Title("FlowUI"), ui.Size(900, 600))
}
```

Use `Run` for synchronous updates. Use `RunCmd` when `Update` needs to start
asynchronous work and send another message later.

## Theming

FlowUI exposes its own `Theme` instead of asking applications to style Gio
widgets directly. Use `DefaultTheme` or `DarkTheme`, then pass it with
`WithTheme`, or mutate tokens with `CustomizeTheme`:

```go
theme := ui.DarkTheme()
theme.Palette.Accent = color.NRGBA{R: 0x00, G: 0x78, B: 0xd4, A: 0xff}

ui.Run(Model{}, Update, View, ui.WithTheme(theme))
```

`MaterialTheme` is available only for the Gio material bridge used by low-level
text/editor internals.

## Components

### Core

- `Run`, `RunCmd`
- `Send`, `Update`, `UpdateCmd`, `View`
- `Cmd`, `Do`
- `Context`
- `Widget`
- `Key`

### App Options

- `Title`
- `Size`
- `WithTheme`
- `CustomizeTheme`
- `MaterialTheme`
- `Locale`

### Controls

- `Text`
- `Label`
- `Button`
- `Input`
- `Checkbox`
- `Switch`
- `SwitchGroup`
- `RadioGroup`
- `ProgressBar`
- `ListBox`
- `Select`
- `ComboBox`
- `DatePicker`
- `Popover`
- `Modal`

### Layout

- `Surface`
- `Box`, `Spacer`
- `Center`
- `Row`, `Column`
- `Expanded`, `Flexible`
- `Adaptive`
- `Wrap`
- `Scroll`, `List`
- `Stack`, `Stacked`, `Overlay`
- `AspectRatio`
- `Grid`, `AutoGrid`
- `Divider`, `Separator`

### Drawing

- `DrawShadow`
- `SurfaceShadow`
- `PopupShadow`
- `RoundedShadowCorners`
- `EllipseShadow`

## Examples

Examples live in `examples/`:

- `counter`: basic MVU state updates.
- `async`: asynchronous messages with commands.
- `buttons`: button variants, loading, disabled, and interaction states.
- `inputs`: input styles and events.
- `labels`: form label states and field association for input, combo box, and select controls.
- `checkboxes`: checkbox states and validation.
- `switches`: switch states, sizes, descriptions, label position, thumb content, and switch groups.
- `radio_groups`: mutually exclusive option selection.
- `progress_bars`: determinate and indeterminate progress indicators.
- `list_boxes`: single-select, multi-select, sections, disabled keys, custom indicators, and action-oriented list boxes.
- `modals`: controlled modal dialogs with sizes, placements, backdrop variants, and dismiss behavior.
- `popovers`: controlled popovers with placement, arrows, dismiss behavior, and interactive content.
- `surfaces`: semantic surface variants, foreground context, rounded corners, and surface elevation.
- `selects`: single and multiple selection, sections, disabled options, validation, controlled open state, and Surface styling.
- `comboboxes`: selectable and filterable options.
- `datepickers`: date selection and constraints.
- `form`: composing controls into a form.
- `layout`: layout primitives.
- `todo`: keyed repeated UI and list interactions.
- `popup_shadow_example`: shadow rendering demo.

## Status

FlowUI is still evolving. The public API is kept small on purpose, and new
widgets are added when they fit the typed MVU style of the framework.
