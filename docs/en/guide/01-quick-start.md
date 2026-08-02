# 01 - Getting started

This page gets a small FlowUI window running in a few minutes.

## Requirements

- Go 1.26.2 or newer
- A desktop platform supported by Gio: Windows, macOS, or Linux

## Install

Inside your Go module:

```bash
go get github.com/qianniancn/flowui/ui
```

Applications should import the public facade only:

```go
import "github.com/qianniancn/flowui/ui"
```

Do not import `github.com/qianniancn/flowui/internal/...`.

## First application

Create `main.go`:

```go
package main

import (
	"fmt"

	"github.com/qianniancn/flowui/ui"
)

type Model struct {
	Count int
}

type Msg interface{ msg() }

type Inc struct{}
type Dec struct{}

func (Inc) msg() {}
func (Dec) msg() {}

func Update(m *Model, msg Msg) ui.Cmd[Msg] {
	switch msg.(type) {
	case Inc:
		m.Count++
	case Dec:
		m.Count--
	}
	return nil
}

func View(_ *ui.Context, m Model, send ui.Send[Msg]) ui.Widget {
	return ui.Center(
		ui.Column(
			ui.Text("FlowUI Counter").Size(24),
			ui.Text(fmt.Sprintf("Count: %d", m.Count)),
			ui.Row(
				ui.Button("decrement", ui.Text("-1")).OnClick(func() {
					send(Dec{})
				}),
				ui.Button("increment", ui.Text("+1")).OnClick(func() {
					send(Inc{})
				}),
			).Gap(8),
		).Gap(12),
	)
}

func main() {
	ui.Run(ui.NewProgram(Model{}, Update, View),
		ui.Title("FlowUI Counter"),
		ui.Size(640, 480),
	)
}
```

Run it with:

```bash
go run .
```

The same program is available in [`examples/counter`](https://github.com/qianniancn/flowui/tree/main/examples/counter).

## What happened?

1. `Model` owns the application state.
2. `Msg` describes user intent.
3. `Update` is the only place that changes the model.
4. `View` reads the model and returns a widget tree.
5. `ui.Run` creates the window and starts the MVU loop.

## Choose an application level

| API | Use it for |
| --- | --- |
| `ui.Run(ui.Program)` | A single window |
| `ui.Application` | Multiple windows, tray integration, or explicit lifecycle control |

Continue with [Core concepts](02-core-concepts.md), or read the
[MVU and messages](03-mvu-and-messages.md) guide.
