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
ui.Modal("settings", m.SettingsOpen, "Settings", body).
	OnOpenChange(func(open bool) { send(SetSettingsOpen(open)) })
```

Keep the open value in the model when other controls need to close or replace
the surface. Use `AlertDialog` for a decision with a clear confirm/cancel path.

## Anchored content

`Popover`, `Dropdown`, and `Tooltip` are anchored to a trigger and manage their
placement around that widget:

```go
ui.Popover("help", m.HelpOpen, ui.Button("help", ui.Text("Help")), helpBody).
	OnOpenChange(func(open bool) { send(SetHelpOpen(open)) })
```

Menus use `MenuItem` values, separators, groups, and `OnChange` messages. A
context menu wraps a trigger area:

```go
ui.ContextMenu("file-menu", fileRow, ui.Menu("file-actions", items))
```

## Portal

Use `Portal` when a custom overlay needs to escape a local clip or coordinate
system. The content callback receives the anchor rectangle and whether the
overlay is interactive:

```go
ui.Portal("custom", m.Open, trigger,
	func(anchor image.Rectangle, interactive bool) ui.Widget {
		return ui.Box(content)
	})
```

Most application code should prefer the higher-level overlay components. The
focused programs under [`examples/`](https://github.com/qianniancn/flowui/tree/main/examples) cover modals, popovers,
dropdowns, menus, and context menus.

## Controlled state

All openable surfaces support the same contract described in
[Forms and controlled state](07-forms-and-controlled-state.md): omit `Open` for
uncontrolled use, or keep it in the model and handle `OnOpenChange`.
