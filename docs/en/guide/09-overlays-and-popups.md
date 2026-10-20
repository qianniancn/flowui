# 09 - Overlays and popups

FlowUI has several overlay primitives. Choose the smallest one that matches
the interaction:

| Component | Use it for |
| --- | --- |
| `Modal` | Blocking work that requires a decision |
| `AlertDialog` | A focused confirmation or warning |
| `Popover` | Content anchored to a trigger |
| `Dropdown` / `Menu` | Selection and commands |
| `ContextMenu` | Pointer-anchored commands |
| `Tooltip` | Short explanatory text |
| `Toast` | Non-blocking feedback |
| `Portal` | Advanced custom placement |

## Modal and dialogs

```go
modal.New("Settings", body).Key("settings").Open(m.SettingsOpen).
	OnOpenChange(func(open bool) { send(SetSettingsOpen(open)) })
```

Keep the open value in the model when other controls need to close or replace
the surface. Use `AlertDialog` for a decision with a clear confirm/cancel path.

## Anchored content

`Popover`, `Dropdown`, and `Tooltip` are anchored to a trigger and manage their
placement around that widget.

`Popover` does not own its trigger's click: the trigger is an ordinary widget, so
opening on click is the application's wiring inside that widget's own `OnClick`.
`Dropdown` and `Menu` are the opposite — they take over the trigger press.

```go
trigger := button.New("Help").
	OnClick(func() { send(SetHelpOpen(!m.HelpOpen)) })
popover.New(trigger, helpBody).Key("help").Open(m.HelpOpen).
	OnOpenChange(func(open bool) { send(SetHelpOpen(open)) })
```

`Dropdown` opens on press by default. Use long press or hover when the
interaction calls for it:

```go
dropdown.New("actions", trigger, items).
	TriggerMode(dropdown.TriggerHover)
```

Hover mode uses short enter and leave delays so the pointer can move from the
trigger into the menu without closing the panel.

Dropdowns can also open from a secondary click, draw an anchor arrow, and size
their panel to its content:

```go
dropdown.New("actions", trigger, items).
	TriggerMode(dropdown.TriggerContextMenu).
	AutoWidth().
	MinWidth(160).
	MaxWidth(320).
	Arrow(true)
```

Use `MatchTriggerWidth(true)` when the panel should be at least as wide as the
trigger. Hover and long-press delays are configurable:

```go
dropdown.New("actions", trigger, items).
	TriggerMode(dropdown.TriggerHover).
	HoverOpenDelay(200 * time.Millisecond).
	HoverCloseDelay(120 * time.Millisecond)
```

Dropdown and menu callbacks use event values. `DropdownOpenChangeEvent` carries
the interaction source, while `MenuActionEvent` and `DropdownActionEvent` carry
the complete item and submenu path:

```go
dropdown.New("actions", trigger, items).
	OnOpenChangeEvent(func(event dropdown.OpenChangeEvent) {
		// event.Source identifies trigger, context menu, menu, outside, and more.
	}).
	OnActionEvent(func(event dropdown.ActionEvent) {
		// event.Item is the full item; event.Path contains its submenu keys.
	})
```

For a split button, use the reusable `DropdownButton` component:

```go
dropdown.Button("create", button.New("Create"), items).
	OnClick(func() { send(Create{}) })
```

`Dropdown` forwards the menu configuration surface as well, including
`OnCheckedChange`, `OnRadioChange`, `AutoSeparateSections`, `Compact`, and
`DataVersion`. Increment `DataVersion` when menu content, grouping, or a
width-affecting child changes so flattened entries and `AutoWidth` measurements
can be reused safely. Use `BeforeContent` and `AfterContent` for custom content
around the menu items.
Use `DropdownItemAction`, `DropdownItemCheckbox`, `DropdownItemRadio`, and
`DropdownItemSubmenu` for explicit item behavior; `DropdownGroupLabel` adds a
non-selectable section heading.

Menus use `MenuItem` values, separators, groups, and `OnChange` messages. A
context menu wraps a trigger area:

```go
menu.ContextMenu("file-menu", fileRow, menu.New("file-actions", items))
```

## Portal

Use `Portal` when a custom overlay needs to escape a local clip or coordinate
system. The content callback receives the anchor rectangle and whether the
overlay is interactive:

```go
portal.New("custom", m.Open, trigger,
	func(anchor image.Rectangle, interactive bool) ui.Widget {
		return layout.Box(content)
	})
```

Most application code should prefer the higher-level overlay components. The
component tests under `widget/modal`, `widget/popover` and `widget/menu` cover
their interaction contracts. The code in this chapter is intended for use in an
application with its own model and message loop.

## Controlled state

Openable surfaces support the same controlled/uncontrolled contract described
in [Forms and controlled state](07-forms-and-controlled-state.md). `Select`,
`ContextMenu`, `Popover`, and `Modal` report changes through `OnOpenChange`;
`Dropdown` reports them through `OnOpenChangeEvent`.

