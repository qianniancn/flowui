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
	return ui.Column(
		ui.Input("name", m.Name).
			Hint("Name").
			OnChange(func(value string) {
				send(NameChanged{Value: value})
			}),
		ui.Text("Hello, "+m.Name),
	).Gap(12)
}
```

The complete runnable form is in [`examples/form`](https://github.com/qianniancn/flowui/tree/main/examples/form).

## Common fields

```go
ui.Input("email", m.Email).
	Hint("you@example.com").
	OnChange(func(value string) { send(EmailChanged{Value: value}) })

ui.TextArea("bio", m.Bio).
	OnChange(func(value string) { send(BioChanged{Value: value}) })

ui.Checkbox("agree", m.Agree).
	Label("I agree to the terms").
	OnChange(func(value bool) { send(AgreeChanged{Value: value}) })
```

### InputGroup actions

Use `Prefix` and `Suffix` for static content. Use `PrefixAction` and
`SuffixAction` for clickable affixes. `InputGroupAction` provides a stable
24dp target, pointer cursor, and accessible label:

```go
ui.InputGroup(ui.Input("website", m.Website)).
	SuffixAction(
		ui.InputGroupAction("clear-website", "Clear website", ui.Icon(lucide.X).Size(16)).
			OnClick(func() { send(ClearWebsite{}) }),
	)
```

Action slots do not focus the editor by default. Call
`.FocusOnActionPress(true)` when the editor should regain focus after an
action. Override the default slot spacing with `PrefixPadding` or
`SuffixPadding` when a custom layout needs it. The editor face is
`PartContent`; use `ui.Part(ui.PartContent, ...)` to customize its cursor,
text, or typography without changing the outer group shell.

Selection controls use model keys:

```go
ui.Select("state", m.State, items).
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
ui.Select("city", m.City, items).
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
