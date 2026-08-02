# 03 - MVU and messages

FlowUI uses a small Model-View-Update loop:

| Role | Signature | Responsibility |
| --- | --- | --- |
| Model | Any application value | Business state |
| Msg | A closed set of message types | What happened |
| Update | `func(*M, Msg) Cmd[Msg]` | Apply state changes and effects |
| View | `func(*Context, M, Send[Msg]) Widget` | Render the current state |

## Closed message sets

A private marker method keeps the message set explicit and makes type switches
easy to review:

```go
type Msg interface{ msg() }

type NameChanged struct{ Value string }
type Submitted struct{}

func (NameChanged) msg() {}
func (Submitted) msg() {}
```

Update handles messages and returns an optional command:

```go
func Update(m *Model, msg Msg) ui.Cmd[Msg] {
	switch msg := msg.(type) {
	case NameChanged:
		m.Name = msg.Value
	case Submitted:
		return submitCmd(m.Name)
	}
	return nil
}
```

## Single-window programs

Use `NewProgram` when the initial model is a fixed value:

```go
ui.Run(ui.NewProgram(Model{}, Update, View),
	ui.Title("FlowUI app"),
	ui.Size(800, 600),
)
```

Use a full `Program` value when initialization returns a command or must create
fresh reference-backed state:

```go
ui.Run(ui.Program[Model, Msg]{
	Init: func() (Model, ui.Cmd[Msg]) {
		return Model{}, loadInitial()
	},
	Update: Update,
	Subscriptions: func(Model) []ui.Subscription[Msg] { return nil },
	View: View,
})
```

## Nested modules

Child modules can keep their own model and messages. The parent wraps child
messages with `ui.MapCmd` and maps the child's `Send` callback into a parent
message. The complete pattern is shown in [`examples/modules`](https://github.com/qianniancn/flowui/tree/main/examples/modules).

## Actions and commands

`Cmd` represents asynchronous effects such as network requests, file IO, and
timers. `Action` and `Command` describe reusable menu, toolbar, and shortcut
actions. They are related but serve different purposes.

Continue with [Commands and subscriptions](08-commands-and-subscriptions.md).
