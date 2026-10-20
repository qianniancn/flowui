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

Applications use public packages according to their responsibilities, for example:

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

	"github.com/qianniancn/flowui/layout"
	"github.com/qianniancn/flowui/theme/base"
	"github.com/qianniancn/flowui/ui"
	"github.com/qianniancn/flowui/widget/button"
	"github.com/qianniancn/flowui/widget/text"
	"github.com/qianniancn/flowui/window"
)

type Model struct {
	Count int
}

type Msg interface{ msg() }

type Inc struct{}
type Dec struct{}

func (Inc) msg() {}
func (Dec) msg() {}

func Update(model *Model, msg Msg) ui.Cmd[Msg] {
	switch msg.(type) {
	case Inc:
		model.Count++
	case Dec:
		model.Count--
	}
	return nil
}

func View(_ *ui.Context, model Model, send ui.Send[Msg]) ui.Widget {
	return layout.Center(
		layout.Column(
			text.New("FlowUI Counter").Size(24),
			text.New(fmt.Sprintf("Count: %d", model.Count)),
			layout.Row(
				button.New("-1").OnClick(func() {
					send(Dec{})
				}),
				button.New("+1").OnClick(func() {
					send(Inc{})
				}),
			).Gap(8),
		).Gap(12),
	)
}

func main() {
	ui.Run(ui.NewProgram(Model{}, Update, View),
		window.Theme(base.Light()),
		window.Title("FlowUI Counter"),
		window.Size(640, 480),
	)
}
```

Run it with:

```bash
go run .
```

This code is also maintained as `examples/counter/main.go`. From the repository
root, run `go run ./examples/counter`. The other programs under `examples/`
demonstrate buttons, inputs, and the default theme in light and dark modes.

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
