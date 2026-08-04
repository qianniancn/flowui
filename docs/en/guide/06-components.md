# 06 - Components

The public `ui` package exposes components as value-like widgets. Most
interactive constructors take a stable key as their first argument and return
a fluent `XxxWidget` value.

## Content

`Text`, `Label`, `Description`, `Image`, `Icon`, `Avatar`, `Badge`, and `Chip`
cover common content and metadata needs.

```go
ui.Text("Heading").Size(20)
ui.Icon(lucide.Plus).Size(16)
```

## Actions

`Button`, `ButtonGroup`, `ToggleButton`, `CloseButton`, and `Chip` provide
action-oriented controls. Use variants and sizes instead of copying paint code.

```go
ui.Button("save", ui.Text("Save")).
	Variant(ui.ButtonPrimary).
	Size(ui.ButtonMedium).
	OnClick(func() { send(Save{}) })
```

## Forms

The form family includes `Input`, `TextArea`, `InputGroup`, `InputGroupAction`,
`Checkbox`, `Switch`, `RadioGroup`, `Select`, `ComboBox`, date pickers, color
pickers, and sliders. Form values are controlled by the model; callbacks send
messages.

## Navigation and overlays

Use `Tabs`, `Sidebar`, `Tree`, `Menubar`, `Menu`, `ContextMenu`, `Pagination`,
and `Toolbar` for navigation. `Dropdown`, `Popover`, `Tooltip`, `Modal`,
`AlertDialog`, `Toast`, and `Portal` handle transient surfaces.

## Data and charts

`Table`, `ListBox`, progress controls, `Meter`, `Spinner`, and the chart family
cover data-heavy screens. The chart family includes line, bar, pie,
candlestick, heatmap, and Gantt charts.

## Layout and shells

`Box`, `Row`, `Column`, `Grid`, `Wrap`, `Stack`, `Scroll`, `SplitPane`,
`Surface`, `Card`, `Overlay`, `WindowTitleBar`, and `StatusBar` form the page
shell.

## Examples

Run the component gallery to see the controls together:

```bash
go run ./examples/components
```

The focused programs under [`examples/`](https://github.com/qianniancn/flowui/tree/main/examples) show individual
interaction contracts. The [component screenshots](16-component-screenshots.md)
provide a visual index.

The [Forms and controlled state](07-forms-and-controlled-state.md) guide covers
the value and open-state contracts shared by many components.

For custom controls and domain drawing, see [Custom components](10-custom-components.md)
for the public style, text measurement, brush, pointer, and visual-overflow
helpers.
