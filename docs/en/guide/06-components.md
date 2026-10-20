# 06 - Components

Import components from `github.com/qianniancn/flowui/widget/<component>`.
Each main constructor is named `New` and returns a value-like `Widget` with
chainable methods. Most interactive constructors take a stable key first.
Related controls use descriptive names, such as `input.TextArea` and
`button.Group`. The `ui` package provides the application runtime.

## Content

`text.New`, `label.New`, `description.New`, `image.New`, `icon.New`, `avatar.New`, `badge.New`, and `chip.New`
cover common content and metadata needs.

```go
text.New("Heading").Size(20)
icon.New(lucide.Plus).Size(16)
```

`code.New` renders a selectable code block. Chroma v2 supplies the lexer, and
multi-line blocks show a line-number gutter by default; single-line blocks hide
it. Use `LineNumbers` to force either state.

```go
code.New("sample", "go", source)
code.New("inline", "go", oneLine).LineNumbers(false)
```

## Actions

`button.New`, `button.Group`, `dropdown.Button`, `togglebutton.New`, `closebutton.New`, and `chip.New` provide
action-oriented controls. Use variants and sizes instead of copying paint code.

```go
button.New("Save").
	Variant(button.Primary).
	Size(button.Medium).
	OnClick(func() { send(Save{}) })
```

Buttons, toggle buttons, tooltips, and input group actions accept plain text
directly. Use `Content` when an icon, styled text, or a layout is needed:

```go
togglebutton.New("Bold").Key("bold").Selected(model.Bold).OnChange(setBold)
tooltip.New("save-tip", saveButton, "Save changes")
input.GroupAction("clear", "Clear").OnClick(clear)
button.New("Save").Content(icon.New(lucide.Save)).IconOnly()
button.New("Save").Content(text.New("Save").Size(20))
```

For buttons, toggle buttons, and input actions, the constructor text also sets
the accessible label. `Content` preserves it; `Label` overrides it without
changing visible content. Cards, toolbars, modal bodies, and popovers remain
widget containers. Standalone text uses `text.New`.

When migrating, rename former main constructors to `New`, replace plain text
children with strings, and move custom children to `.Content(...)`.
`input.GroupAction` now takes two arguments; `toast.Toast` is `toast.NewItem`.

## Forms

The form family includes `input.New`, `input.TextArea`, `input.Group`, `input.GroupAction`,
`checkbox.New`, `switchcontrol.New`, `radiogroup.New`, `selectfield.New`, `combobox.New`, date pickers, color
pickers, and sliders. Form values are controlled by the model; callbacks send
messages.

`numericinput.New` adds controlled `float64` editing with `Range`, `Step`, and
`Precision`. Incomplete drafts remain local until Enter or focus loss; invalid
input restores the controlled value. Step buttons commit immediately.
`segmented.New` provides compact, mutually exclusive choices; `FullWidth()` gives
each choice an equal share of the container. Clicking the current choice does
not deselect it.

```go
numericinput.New("amount", model.Amount).Range(-100, 100).Step(0.25).
	Precision(2).Label("Amount").OnChange(setAmount)

segmented.New([]segmented.Item{
	{Key: "day", Label: "Day"}, {Key: "week", Label: "Week"},
}).Key("period").SelectedKey(model.Period).OnChange(setPeriod)
```

## Navigation and overlays

Use `tabs.New`, `sidebar.New`, `tree.New`, `menubar.New`, `menu.New`, `menu.ContextMenu`, `pagination.New`,
and `toolbar.New` for navigation. `dropdown.New`, `popover.New`, `tooltip.New`, `modal.New`,
`alertdialog.New`, `toast.New`, and `portal.New` handle transient surfaces.

Use `toast.MessageItem` for the lightweight top-center message pattern: a
single-line title, automatic 3-second timeout, content-driven width, and no
description, action, or close button. Use `toast.NewItem` when the feedback needs
a description or action.

`breadcrumb.New("path", items).OnChange(navigate)` displays a hierarchy: ancestors
are clickable, and the final item is the current location. It wraps in narrow containers.

`commandpalette.New("commands", model.Open, commands).OnOpenChange(setOpen)` shares
`[]command.Command` with buttons and menus. It searches labels, descriptions, and
keys, supports arrow keys, Enter, Escape, disabled commands, and `KeepOpen`.
Register application shortcuts separately through `command.Scope`.

`sidebar.New` is the controlled primary application navigation surface. Use
`sidebar.Item.Children` for nested destinations and keep the expanded groups in
application state with `OpenKeys` and `OnOpenChange`; `InlineIndent` adjusts
the nesting offset, `ExpandAction` chooses click or hover activation while the
sidebar is expanded, and each item can replace the default chevron through
`Switcher` / `ExpandedSwitcher`.
A collapsed parent opens its descendants in a right-side menu. `menu.New` remains
the component for commands and context actions, while `tree.New` is for content
hierarchies such as a file explorer.

`tabs.New` uses secondary tabs (`tabs.Secondary`) as its standard treatment and also
supports segmented tabs (`tabs.Primary`), three sizes, leading and trailing
slots, icon/content tab items, close actions, and independently configured
label, indicator, and panel motion:

```go
tabs.New("workspace-tabs", model.Active, items).
	Variant(tabs.Secondary).
	Size(tabs.Large).
	Placement(tabs.Bottom).
	Centered(true).
	Leading(icon.New(lucide.LayoutDashboard).Size(16)).
	Trailing(button.New("New tab").Content(icon.New(lucide.Plus)).IconOnly().OnClick(addTab)).
	KeepAlive(true).
	ForceRender(true).
	PanelTransition(tabs.PanelFade).
	OnClose(closeTab)
```

Use `tabs.Item.Editable` with `EditingKey` and `OnEdit` for inline editing, and
`Closable` with `OnClose` for closable tabs.

Use `Overflow(tabs.OverflowMenu)` (or `tabs.OverflowAuto`) to replace items that
do not fit with a keyboard-navigable More menu. `OnAdd` renders an add-tab action;
`tabs.Item.Editable`, `EditingKey`, and `OnEdit` provide controlled inline label
editing. `OverflowTrigger` supplies custom More content, while
`IndicatorWidth` and `IndicatorAlign` configure a narrowed selected indicator.
Closing the active item requests the next enabled tab, falling back to the
previous enabled tab.

Use `KeepAlive(true)` when hidden panels should retain transient widget state;
`ForceRender(true)` (or `Lazy(false)`) initializes inactive panels in an
isolated, non-interactive layout pass, and `DestroyOnHidden(true)` takes
precedence when a panel must be recreated after switching away. Use
`Activation(tabs.ActivationManual)` for
focus-first keyboard navigation, and set `tabs.Item.AccessibleLabel` when an icon
or custom tab content needs a label for assistive technology. Gio's current
semantic model exposes these as button-compatible tab descriptions; FlowUI also
publishes tab-list and active-panel labels/descriptions and the panel enabled state.
Inline editing commits on Enter or focus loss and cancels with Escape.

`tree.New` now supports single and multi selection, drag and drop, async loading, rename, filtering, checkbox trees (`Checkable`, `CheckedKeys`, `HalfCheckedKeys`, `CheckStrictly`), directory-style `ExpandAction` single or double click expansion, and per-item `Content` / `Switcher` / `ExpandedSwitcher` customization. Checkbox conduction treats `Disabled`, `DisabledKeys`, and `DisableCheckbox` as association boundaries without affecting parent derivation outside those boundaries. `DragDisabled` and `DropDisabled` independently restrict a node as a source or target; `OnDragStart`, `OnDragEnter`, `OnDragLeave`, `OnDragOver`, and `OnDragEnd` report valid drop-target lifecycle changes while `OnDrop` remains the final commit signal. `OnLoadChildrenEvent` reports the key, retry state, and current expansion state. `FilterFunc` supports business-specific matching while preserving ancestors; pair dynamic predicates with `FilterVersion` to invalidate the cache. Use `DataVersion` for large stable data and increase it whenever the hierarchy or row content changes. `tree.ItemsFromSimple` adapts flat records into a hierarchy, treating unknown parent keys as roots.

## Time line

`timeline.New` displays a sequence of events. It supports vertical or horizontal orientation, `start`, `end`, and `alternate` modes, per-item `Placement`, `TitleSpan` or `TitleWidth`, `Gap`, filled or outlined markers, reverse order, custom colors, icons, and pending events. `PendingLoading` controls the pending marker, and `Disabled` mutes the entire timeline.

## Data and charts

`table.New`, `listbox.New`, progress controls, `progress.Meter`, `spinner.New`, and the chart family
cover data-heavy screens. The chart family includes line, bar, pie,
candlestick, heatmap, and Gantt charts. `nodegraph.New` provides a node-and-edge
canvas with controlled panning and zooming; see [Node graph](17-node-graph.md).

Use `layout.List(count, rowAt)` or
`table.NewVirtual("table", columns, count, rowAt)` with a bounded viewport for
large datasets. Rows are built on demand; the application applies virtual-table
sorting in its row provider. Showcase includes 10,000-row list and table demos,
plus a tree page covering loading, filtering, checkboxes, rename, and drag and drop.

`skeleton.New().Width(160).Height(16)` creates a loading placeholder;
`Circle().Height(48)` creates a circular one. Width fills the container by default.
`Animated(false)` or disabled theme Motion stops its pulse and frame requests.

## Layout and shells

`layout.Box`, `layout.Row`, `layout.Column`, `layout.Grid`, `layout.Wrap`, `layout.Stack`, `layout.Scroll`, `layout.SplitPane`,
`dock.New`, `panel.New`, `surface.New`, `card.New`, `layout.Overlay`,
`titlebar.New`, and `statusbar.New` form the page shell.

`animation.Tween` interpolates values and `animation.AnimateLayout` animates
layout dimensions. Both follow the theme Motion setting; see the Animation showcase page.

`panel.New` owns the lifecycle of one visible view without
rendering navigation. `KeepAlive(true)` retains hidden widget state,
`ForceRender(true)` initializes inactive views in an isolated pass, and
`DestroyOnHidden(true)` takes precedence when views must be recreated.

`dock.New` is a declarative tree of `dock.Panel` and `dock.Split` nodes. Nested
splits provide independent resizers while `dock.Snapshot` can be kept in
the application model for restoring divider ratios, collapsed branches, and a
maximized node:

```go
root := dock.Split("workspace", dock.Horizontal,
	dock.Panel("explorer", explorer),
	dock.Split("editor-bottom", dock.Vertical,
		dock.Panel("editor", editor),
		dock.Panel("bottom", bottom),
	).Ratio(.72),
).Ratio(.24)

dock.New("workbench", root).
	OnChange(func(snapshot dock.Snapshot) { send(LayoutChanged{snapshot}) })
```

Use `MaximizedKey("editor")` (or `Snapshot.MaximizedKey`) to give one node the
full region and pass an empty key to restore the tree.

The layout primitive does not impose panel headers, tabs, or collapse buttons;
those remain regular FlowUI widgets supplied by the application.

`workbench.State` and `workbench.Controller` coordinate tab groups, dock
geometry, and shell-region visibility without owning editor buffers or view
content. Use `BindTabs`, `BindPanel`, and `BindDock` to connect ordinary
widgets to the controller. The controller's `Commands` or `CommandScope`
method installs the standard next/previous/close-tab and shell visibility
shortcuts.

`workbench.Snapshot` persists stable group/tab keys, dock geometry, and chrome
visibility. Encode it with `workbench.MarshalSnapshot` and decode it with
`workbench.UnmarshalSnapshot`. Snapshots must use the current schema version;
Restore validates the complete snapshot before changing state.

## Examples

Run `go run ./examples/showcase` from the repository root to explore 65 pages
covering every public widget package, with search, live interaction, usage snippets,
light/dark modes, and a directory view for narrow windows.

Run `go run ./examples/style-gallery` for a compact preview of the default theme.

The `examples/buttons` and `examples/inputs` programs demonstrate focused
interaction contracts. The [component screenshots](16-component-screenshots.md)
provide a visual index.

The [Forms and controlled state](07-forms-and-controlled-state.md) guide covers
the value and open-state contracts shared by many components.

For custom controls and domain drawing, see [Custom components](10-custom-components.md)
for the public style, text measurement, brush, pointer, and visual-overflow
helpers.

