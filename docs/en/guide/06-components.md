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

`Button`, `ButtonGroup`, `DropdownButton`, `ToggleButton`, `CloseButton`, and `Chip` provide
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

`Tabs` uses secondary tabs (`TabsSecondary`) as its standard treatment and also
supports segmented tabs (`TabsPrimary`), three sizes, leading and trailing
slots, icon/content tab items, close actions, and independently configured
label, indicator, and panel motion:

```go
ui.Tabs("workspace-tabs", model.Active, items).
	Variant(ui.TabsSecondary).
	Size(ui.TabsLarge).
	Placement(ui.TabsBottom).
	Centered(true).
	Leading(ui.Icon(lucide.LayoutDashboard).Size(16)).
	Trailing(ui.Button("new-tab", ui.Icon(lucide.Plus)).OnClick(addTab)).
	KeepAlive(true).
	ForceRender(true).
	PanelTransition(ui.TabsPanelFade).
	OnClose(closeTab)
```

Use `TabItem.Editable` with `EditingKey` and `OnEdit` for inline editing, and
`Closable` with `OnClose` for closable tabs.

Use `Overflow(ui.TabsOverflowMenu)` (or `TabsOverflowAuto`) to replace items that
do not fit with a keyboard-navigable More menu. `OnAdd` renders an add-tab action;
`TabItem.Editable`, `EditingKey`, and `OnEdit` provide controlled inline label
editing. `OverflowTrigger` supplies custom More content, while
`IndicatorWidth` and `IndicatorAlign` configure a narrowed selected indicator.
Closing the active item requests the next enabled tab, falling back to the
previous enabled tab.

Use `KeepAlive(true)` when hidden panels should retain transient widget state;
`ForceRender(true)` (or `Lazy(false)`) initializes inactive panels in an
isolated, non-interactive layout pass, and `DestroyOnHidden(true)` takes
precedence when a panel must be recreated after switching away. Use
`Activation(ui.TabsActivationManual)` for
focus-first keyboard navigation, and set `TabItem.AccessibleLabel` when an icon
or custom tab content needs a label for assistive technology. Gio's current
semantic model exposes these as button-compatible tab descriptions; FlowUI also
publishes tab-list and active-panel labels/descriptions and the panel enabled state.
Inline editing commits on Enter or focus loss and cancels with Escape.

## Data and charts

`Table`, `ListBox`, progress controls, `Meter`, `Spinner`, and the chart family
cover data-heavy screens. The chart family includes line, bar, pie,
candlestick, heatmap, and Gantt charts.

## Layout and shells

`Box`, `Row`, `Column`, `Grid`, `Wrap`, `Stack`, `Scroll`, `SplitPane`,
`DockLayout`, `PanelHost`, `ViewStack`, `Surface`, `Card`, `Overlay`,
`WindowTitleBar`, and `StatusBar` form the page shell.

`PanelHost`/`ViewStack` owns the lifecycle of one visible view without
rendering navigation. `KeepAlive(true)` retains hidden widget state,
`ForceRender(true)` initializes inactive views in an isolated pass, and
`DestroyOnHidden(true)` takes precedence when views must be recreated.

`DockLayout` is a declarative tree of `DockPanel` and `DockSplit` nodes. Nested
splits provide independent resizers while `DockLayoutSnapshot` can be kept in
the application model for restoring divider ratios, collapsed branches, and a
maximized node:

```go
root := ui.DockSplit("workspace", ui.DockHorizontal,
	ui.DockPanel("explorer", explorer),
	ui.DockSplit("editor-bottom", ui.DockVertical,
		ui.DockPanel("editor", editor),
		ui.DockPanel("bottom", bottom),
	).Ratio(.72),
).Ratio(.24)

ui.DockLayout("workbench", root).
	OnChange(func(snapshot ui.DockLayoutSnapshot) { send(LayoutChanged{snapshot}) })
```

Use `MaximizedKey("editor")` (or `Snapshot.MaximizedKey`) to give one node the
full region and pass an empty key to restore the tree.

The layout primitive does not impose panel headers, tabs, or collapse buttons;
those remain regular FlowUI widgets supplied by the application.

`WorkbenchState` and `WorkbenchController` coordinate tab groups, dock
geometry, and shell-region visibility without owning editor buffers or view
content. Use `BindTabs`, `BindPanel`, and `BindDock` to connect ordinary
widgets to the controller. The controller's `Commands` or `CommandScope`
method installs the standard next/previous/close-tab and shell visibility
shortcuts.

`WorkbenchSnapshot` persists stable group/tab keys, dock geometry, and chrome
visibility. Encode it with `MarshalWorkbenchSnapshot`, decode it with
`UnmarshalWorkbenchSnapshot`, and pass `WorkbenchMigration` to `Restore`
when keys have been renamed. Restore validates the complete snapshot before
changing state; pre-versioned snapshots keep the current chrome visibility
because that field did not exist in the legacy format.

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
