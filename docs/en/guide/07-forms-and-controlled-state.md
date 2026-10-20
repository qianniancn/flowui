# 07 - Forms and controlled state

Form controls display a value from the model and report user intent through a
callback. The model remains the source of truth.

## Minimal form

```go
type Model struct {
	Name string
}

type NameChanged struct{ Value string }

func View(_ *ui.Context, m Model, send ui.Send[Msg]) ui.Widget {
	return layout.Column(
		input.New(m.Name).
			Hint("Name").
			OnChange(func(value string) {
				send(NameChanged{Value: value})
			}),
		text.New("Hello, "+m.Name),
	).Gap(12)
}
```

Run `go run ./examples/inputs` from the repository root for controlled inputs.
Other examples in this chapter are usage snippets.

## Common fields

```go
input.New(m.Email).
	Hint("you@example.com").
	OnChange(func(value string) { send(EmailChanged{Value: value}) })

input.NewTextArea(m.Bio).
	OnChange(func(value string) { send(BioChanged{Value: value}) })

checkbox.New("I agree to the terms").Key("agree").Checked(m.Agree).
	OnChange(func(value bool) { send(AgreeChanged{Value: value}) })
```

### Right-click edit menu

Input and TextArea ship with a built-in right-click edit menu (Cut / Copy /
Paste / Select All). Entries enable and disable themselves based on the
read-only, password, and selection state: password fields never offer cut or
copy, and read-only fields never offer cut or paste. Shortcut hints follow the
platform (Ctrl on Windows / Linux, Command on macOS), and the menu labels
follow the interface language (Chinese / English).

Fields inside `input.Group` / `GroupTextArea` and the `combobox.New` input
share this built-in menu, with the same `.DefaultContextMenu(false)` and
`.ContextMenu(...)` options; `text.Selectable` ships a right-click menu with
Copy / Select All. While a menu is open, the selection highlight stays visible
instead of disappearing when the menu takes focus.

Call `.DefaultContextMenu(false)` to turn the built-in menu off, or
`.ContextMenu(...)` to replace it entirely with a custom menu:

```go
input.New(m.Email).DefaultContextMenu(false)

input.New(m.Email).
	ContextMenu(menu.New("email-menu", []menu.Item{
		{Key: "copy", Label: "Copy address"},
	}))
```

### input.Group actions

Use `Prefix` and `Suffix` for static content. Use `PrefixAction` and
`SuffixAction` for clickable affixes. `input.GroupAction` provides a stable
24dp target, pointer cursor, and accessible label:

```go
input.Group(input.New(m.Website)).
	SuffixAction(
		input.GroupAction("clear-website", "Clear website").Content(icon.New(lucide.X).Size(16)).
			OnClick(func() { send(ClearWebsite{}) }),
	)
```

Action slots do not focus the editor by default. Call
`.FocusOnActionPress(true)` when the editor should regain focus after an
action. Override the default slot spacing with `PrefixPadding` or
`SuffixPadding` when a custom layout needs it. The editor face is
`PartContent`; use `style.New().Part(style.PartContent, ...)` to customize its cursor,
text, or typography without changing the outer group shell.

Selection controls use model keys:

```go
selectfield.New(items).Key("state").SelectedKey(m.State).
	Label("State").
	Placeholder("Select one").
	OnChange(func(key string) { send(SetState(key)) })
```

Use `Invalid` and `ErrorMessage` for validation feedback. Validate and update
the error fields in `Update`, usually when the user submits.

## Open state

Select, dropdown, context-menu, popover, and modal controls support two modes:

```text
No Open(...)       -> uncontrolled; the widget owns open state
Open(bool)         -> controlled; the model owns open state
DefaultOpen(bool)  -> initial value for uncontrolled state
```

Controlled usage must update the model from the component's open-state callback:

```go
selectfield.New(items).Key("city").SelectedKey(m.City).
	Open(m.CityOpen).
	OnOpenChange(func(open bool) { send(SetCityOpen(open)) }).
	OnChange(func(key string) { send(SetCity(key)) })
```

Calling `Open(...)` without handling the callback leaves the panel stuck at the
value supplied by the model.

`Dropdown` uses `OnOpenChangeEvent` and reads the new state from `event.Open`.
`Select`, `ContextMenu`, `Popover`, and `Modal` continue to use `OnOpenChange`.

## Labels and descriptions

Prefer a control's `Label` and `Description` methods. For custom compositions,
associate a `Label` with the field key so keyboard focus and accessibility
semantics remain correct.

Continue with [Commands and subscriptions](08-commands-and-subscriptions.md).

