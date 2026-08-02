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

Select, dropdown, menu, popover, and modal controls support two modes:

```text
No Open(...)       -> uncontrolled; the widget owns open state
Open(bool)         -> controlled; the model owns open state
DefaultOpen(bool)  -> initial value for uncontrolled state
```

Controlled usage must update the model from `OnOpenChange`:

```go
ui.Select("city", m.City, items).
	Open(m.CityOpen).
	OnOpenChange(func(open bool) { send(SetCityOpen(open)) }).
	OnChange(func(key string) { send(SetCity(key)) })
```

Calling `Open(...)` without handling the callback leaves the panel stuck at the
value supplied by the model.

## Labels and descriptions

Prefer a control's `Label` and `Description` methods. For custom compositions,
associate a `Label` with the field key so keyboard focus and accessibility
semantics remain correct.

Continue with [Commands and subscriptions](08-commands-and-subscriptions.md).
