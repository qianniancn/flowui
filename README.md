# FlowUI

[中文](README.zh-CN.md)

FlowUI is a small MVU UI framework built on top of Gio.

Public applications import `github.com/qianniancn/FlowUI/ui`. See the
[architecture guide](docs/architecture.md) for package and state ownership
rules.

It keeps application state in a `Model`, sends typed messages from the view, and
makes `Update` the only place that mutates state. The API is intentionally
Go-first: widgets are regular values, options are chainable methods, and local
widget state is managed by `Context` with explicit keys.

## Features

- MVU architecture: `Model`, `Msg`, `Update`, and `View`.
- Typed message dispatch through `ui.Send[Msg]`.
- Optional command effects with `RunCmd`, `Cmd`, and `Do`.
- Parent-child MVU command composition with `MapCmd`.
- Complete programs with startup commands, subscriptions, and window-state messages.
- Reusable application commands shared by shortcuts, menus, dropdowns, and toolbars.
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

import "github.com/qianniancn/FlowUI/ui"

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
asynchronous work and send another message later. Use `RunProgram` when startup
work, subscriptions, or native window-state changes must also enter the MVU loop.

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

An application that owns its windows can switch one active window at runtime.
`SetTheme` and `SetLanguage` queue the change on that window's event loop:

```go
application.SetTheme("main", ui.DarkTheme())
application.SetLanguage("main", ui.LanguageChinese)
```

## Components

### Core

- `Run`, `RunCmd`, `RunProgram`, `Program`
- `Send`, `Update`, `UpdateCmd`, `View`
- `Cmd`, `Do`, `MapCmd`
- `Command`, `Shortcut`, `CommandScope`
- `Context`
- `Widget`, `WidgetFunc`
- `UseState`, `UseStateWith`
- `Key`
- `Portal`
- `TrackOverlayPlacement`

### App Options

- `Title`
- `Size`
- `WithTheme`
- `CustomizeTheme`
- `MaterialTheme`
- `Locale`
- `Application.SetTheme`, `Application.SetLanguage`

### Testing

- `uitest.New`, `uitest.NewWithConfig`
- `Harness.Frame`, `Context`, `Router`, `Click`, `Key`, `Advance`, `Resize`
- `uitest.NewApp`, `uitest.NewAppWithConfig`
- `AppHarness.Send`, `Frame`, `Wait`, `Model`, `Errors`, `Close`

### Controls

- `Text`
- `SelectableText`
- `Label`
- `Description`
- `Image`
- `Avatar`
- `Badge`
- `Button`, `ButtonGroup`
- `ToggleButton`
- `Input`
- `TextArea`
- `Checkbox`
- `Switch`
- `SwitchGroup`
- `RadioGroup`
- `ProgressBar`
- `ProgressCircle`
- `Meter`
- `LineChart`
- `BarChart`
- `PieChart`
- `CandlestickChart`
- `Spinner`
- `Slider`, `RangeSlider`
- `ListBox`
- `Tree`
- `Sidebar`
- `WindowTitleBar`
- `StatusBar`
- `Table`
- `Pagination`
- `Tabs`
- `Select`
- `ComboBox`
- `DateField`, `DatePicker`, `DateRangePicker`
- `ColorArea`, `ColorField`, `ColorPicker`, `ColorSlider`, `ColorSwatch`, `ColorSwatchPicker`
- `Popover`
- `Tooltip`
- `Menu`, `ContextMenu`
- `Menubar`
- `Dropdown`
- `Modal`

### Layout

- `Surface`
- `Card`, `CardHeader`, `CardTitle`, `CardDescription`, `CardContent`, `CardFooter`
- `Box`, `Spacer`
- `Center`
- `Row`, `Column`
- `Expanded`, `Flexible`
- `Adaptive`
- `Wrap`
- `Scroll`, `Scrollbar`, `List`
- `SplitPane`
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

### Animation

- `Tween`
- ECharts/zrender-compatible linear, quadratic, cubic, quartic, quintic, sinusoidal,
  exponential, circular, elastic, back, and bounce easing families
- `LerpFloat`, `LerpFloat64`, `LerpColor`, `LerpPoint`, `LerpRect`

## Examples

Examples live in `examples/`:

- `animations`: reusable easing, tween state, value interpolation, direction changes, and custom Gio drawing.
- `counter`: basic MVU state updates.
- `async`: asynchronous messages with commands.
- `modules`: parent-child MVU composition with a child package, mapped messages, and `MapCmd`.
- `commands`: reusable actions shared by global shortcuts, menubars, dropdowns, and toolbars.
- `texts`: text typography, alignment, truncation, wrapping, and selectable clipboard content.
- `title_bars`: VS Code-style menu and window controls for undecorated desktop windows.
- `multi_windows`: independent windows, native window controls, and per-window runtime theme and language switching.
- `custom_widgets`: custom drawing, transient keyed state, focus handling, root Portal content, Gio transforms, and FlowUI easing.
- `buttons`: button variants, loading, disabled, and interaction states.
- `button_groups`: grouped button variants, sizes, orientations, separators, disabled states, and full-width layout.
- `toggle_buttons`: HeroUI-aligned controlled toggle buttons with variants, sizes, icon-only, selected, and disabled states.
- `close_buttons`: HeroUI-aligned close buttons with disabled, custom icon, and interactive states.
- `chips`: HeroUI-aligned compact labels with colors, variants, sizes, icons, and semantic statuses.
- `images`: reusable images with scaling, positioning, dimensions, rounded clipping, opacity, and accessible labels.
- `avatars`: HeroUI-aligned profile images with sizes, semantic colors, soft variants, and text or icon fallbacks.
- `badges`: HeroUI-aligned anchored indicators with colors, variants, sizes, placements, text, icons, and status dots.
- `inputs`: HeroUI-aligned input variants, types, states, and controlled events.
- `input_groups`: HeroUI-aligned grouped fields with single-line or multiline editors, prefixes, suffixes, variants, validation states, and interactive actions.
- `textareas`: HeroUI-aligned multiline fields with variants, controlled values, row sizing, states, InputGroup integration, and surface usage.
- `labels`: form label states and field association for input, combo box, and select controls.
- `descriptions`: supporting text states, field association, wrapping, and component compatibility.
- `checkboxes`: HeroUI-aligned variants, indeterminate and read-only states, descriptions, validation, and custom indicators.
- `switches`: switch states, sizes, descriptions, label position, thumb content, and switch groups.
- `radio_groups`: mutually exclusive option selection.
- `progress_bars`: determinate and indeterminate progress indicators.
- `progress_circles`: HeroUI-aligned circular determinate and indeterminate progress indicators with sizes and semantic colors.
- `meters`: HeroUI-aligned known-range measurements with labels, formatted values, colors, sizes, and accessible label-free usage.
- `line_charts`: native Gio straight, smooth, stepped, dashed, dotted, and area line charts with animations, interactive legends, data windows, annotations, data activation, custom tooltips, gaps, and crosshairs.
- `bar_charts`: native Gio vertical and horizontal grouped or stacked bar charts with animations, value labels, interactive legends, data windows, annotations, data activation, custom tooltips, positive and negative values, and category highlights.
- `pie_charts`: native Gio pie, donut, and Nightingale rose charts with ECharts-aligned angles, labels, animations, interactive legends, data activation, and custom tooltips.
- `spinners`: animated loading indicators with HeroUI-aligned colors and sizes.
- `sliders`: controlled single-value, range, vertical, disabled, stepped, and formatted sliders.
- `list_boxes`: single-select, multi-select, sections, disabled keys, custom indicators, and action-oriented list boxes.
- `trees`: controlled hierarchical navigation with expansion, selection, custom content, disabled nodes, and scrolling.
- `sidebars`: controlled desktop navigation with sections, collapsed rail, custom content, disabled destinations, and keyboard navigation.
- `status_bars`: compact desktop application status bars with left and right content, semantic surfaces, and optional accent styling.
- `scrollbars`: HeroUI-aligned thin vertical and horizontal scrollbars backed by Gio track and thumb interactions.
- `split_panes`: resizable horizontal, vertical, and nested desktop panes with minimum sizes and keyboard control.
- `tables`: HeroUI-aligned data tables with optional Excel-style grid lines and borders, interactive custom cells, complete-row context menus, resizing, range selection, pagination, asynchronous loading, virtual rows, and scrolling.
- `paginations`: HeroUI-aligned controlled page navigation with summaries, ellipses, sizes, and disabled states.
- `tabs`: primary and secondary variants, horizontal and vertical layouts, disabled tabs, separators, compact accent styling, and overflow scrolling.
- `modals`: controlled modal dialogs with sizes, placements, backdrop variants, and dismiss behavior.
- `alert_dialogs`: HeroUI-aligned confirmation dialogs with semantic statuses, controlled dismissal, sizes, placements, and backdrop variants.
- `popovers`: controlled popovers with placement, arrows, dismiss behavior, and interactive content.
- `tooltips`: HeroUI-aligned hover and focus hints with delays, arrows, placement, and viewport flipping.
- `context_menus`: right-click and long-press menus for table rows, including checkbox, radio, disabled, danger, and submenu items.
- `menubars`: application menu bars with coordinated menus, hover switching, keyboard navigation, and nested commands.
- `dropdowns`: HeroUI-aligned action and selection dropdowns with sections, rich items, custom content, long press, and submenus.
- `toasts`: HeroUI-aligned controlled notifications with variants, actions, timeouts, expandable stacking, and six placements.
- `surfaces`: semantic surface variants, foreground context, rounded corners, and surface elevation.
- `cards`: HeroUI-aligned card variants, semantic sections, and composed actions.
- `selects`: single and multiple selection, sections, disabled options, validation, controlled open state, and Surface styling.
- `comboboxes`: selectable and filterable options.
- `datepickers`: segmented date entry, date range selection, calendars, and constraints.
- `color_pickers`: HeroUI-aligned controlled color selection with HSB area, hue and alpha sliders, presets, and hex input.
- `form`: composing controls into a form.
- `layout`: layout primitives.
- `todo`: keyed repeated UI and list interactions.
- `popup_shadow_example`: shadow rendering demo.

## Status

FlowUI is still evolving. The public API is kept small on purpose, and new
widgets are added when they fit the typed MVU style of the framework.
