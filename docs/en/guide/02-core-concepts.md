# 02 - Core concepts

FlowUI keeps three responsibilities separate:

| Responsibility | Owner |
| --- | --- |
| Business state and side effects | `Model`, messages, and `Update` |
| Box geometry and appearance | `Style` and component variants |
| Child measurement and placement | Layout containers such as `Row`, `Column`, and `Grid` |

## Model, View, and Update

- **Model** is application data.
- **Msg** is a user intent or an asynchronous result.
- **Update** changes the model and can return a `Cmd`.
- **View** reads a model value and returns a `Widget`.

```go
type Model struct {
	Name string
}

type NameChanged struct{ Value string }

func Update(m *Model, msg Msg) ui.Cmd[Msg] {
	switch msg := msg.(type) {
	case NameChanged:
		m.Name = msg.Value
	}
	return nil
}

func View(_ *ui.Context, m Model, send ui.Send[Msg]) ui.Widget {
	return ui.Column(
		ui.Input("name", m.Name).OnChange(func(value string) {
			send(NameChanged{Value: value})
		}),
		ui.Text("Hello, "+m.Name),
	).Gap(12)
}
```

View code must not mutate the model. Event callbacks send a message; the next
Update call applies it.

## Keys

Interactive widgets retain Gio state between frames by key. Give every repeated
or stateful widget a stable key derived from the business identity, not from a
temporary position in a slice.

```go
ui.Input("email", m.Email)
ui.Button("save", ui.Text("Save"))
```

Changing a key intentionally creates a new interaction state. Reusing one key
for two widgets causes state collisions and is an error.

## Context

`*ui.Context` provides frame services such as theme snapshots, language,
focus, transient widget state, and invalidation. It is not a business data
store. Keep application data in `Model`; use `ui.UseState` only for short-lived
interaction state owned by a widget.

## Widgets and layout

Widgets are value-like declarations rebuilt on each frame. Layout containers
measure and place their children. Styles describe the child's box and visual
properties; they do not replace layout policy.

See [Layout](04-layout.md) and [Styling and themes](05-styling-and-themes.md)
for the two halves of rendering.
