# 04 - Layout

FlowUI separates box styling from child arrangement:

| Owner | Controls |
| --- | --- |
| `Style` | Width, height, min/max, padding, margin, overflow, and paint |
| Layout container | Direction, gaps, alignment, flex growth, scrolling, and measurement |

Do not use a style to express sibling spacing or flex growth. Put that policy
on `Row`, `Column`, or another container.

## Common containers

```go
ui.Column(
	ui.Text("Title").Size(20),
	ui.Row(
		ui.Button("cancel", ui.Text("Cancel")),
		ui.Button("save", ui.Text("Save")),
	).Gap(8),
).Gap(12)
```

`Box` gives one child a styled shell. `Center` centers its child in the
available constraints. `Expanded` and `Flexible` distribute remaining space.

```go
ui.Row(
	ui.Sidebar("navigation", "home", items),
	ui.Expanded(ui.Scroll("content", ui.Column(/* content */))),
)
```

## Scrolling and grids

`Scroll` requires a stable key. `Grid` uses a fixed column count, while
`AutoGrid` chooses columns from a minimum column width.

```go
ui.Scroll("content", ui.Column(/* many children */).Gap(8))

ui.Grid(3, cardA, cardB, cardC).Gap(16)
ui.AutoGrid(190, cardA, cardB, cardC).ColumnGap(16).RowGap(12)
```

The runnable [`examples/grid_layout`](https://github.com/qianniancn/flowui/tree/main/examples/grid_layout) compares
fixed columns, responsive columns, and independent row/column gaps.

Other useful containers include `Wrap`, `Stack`, `SplitPane`, `Surface`, and
`Card`. See the [component guide](06-components.md) for the full list.

## Size and constraints

Declare dimensions on a style and apply it to a box or component:

```go
ui.Box(ui.Input("name", m.Name)).Style(ui.Width(320))

ui.Box(child).Style(
	ui.Width(200).
		Height(40).
		MinWidth(120).
		MaxWidth(400).
		Padding(12).
		Margin(4),
)
```

Use `AspectRatio` for a fixed width-to-height ratio:

```go
ui.AspectRatio(1, ui.Box(ui.Text("Square content")))
```

Parents provide maximum constraints. Children choose a size inside those
constraints, and containers decide how children are placed.

## A typical page shell

```go
func View(_ *ui.Context, m Model, send ui.Send[Msg]) ui.Widget {
	return ui.Column(
		ui.WindowTitleBar("title", "FlowUI", menu),
		ui.Expanded(ui.Row(
			ui.Sidebar("sidebar", m.Selected, items),
			ui.Expanded(ui.Scroll("page", content)),
		)),
		ui.StatusBar(left, right),
	)
}
```

Build widgets every frame, keep keys stable, and use business IDs for list
items whose order can change.
