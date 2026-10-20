# 11 - Multiple windows

Use `ui.Run(Program)` for one window. Use `ui.Application` when the process
owns multiple windows, a system tray, or an explicit lifecycle.

## Basic structure

```go
func main() {
	application := ui.NewApplication()

	counter := ui.NewWindow("counter", ui.NewProgram(
		CounterModel{}, counterUpdate, counterView,
	), window.Title("Counter"), window.Size(400, 300))

	mainWindow := ui.NewWindow("main", ui.NewProgram(
		MainModel{}, mainUpdate, mainView,
	), window.Title("Main"), window.Size(640, 480))

	application.Run(mainWindow)
	_ = counter
}
```

`Application.Open` returns `true` when a new window starts. If the key is
already open, it raises the existing window and returns `false`.

```go
if application.Open(counter) {
	// A new window started.
}
application.Close("counter")
```

Each window has an independent model instance. Use a full `Program` initializer
when a model contains slices, maps, or pointers that must be freshly allocated
for every opening.

## Theme and language

Apply changes through the application so they reach the target window's event
loop:

```go
application.SetTheme("main", base.Dark())
application.SetLanguage("main", window.LanguageChinese)
```

Do not mutate a window's theme from an arbitrary goroutine.

## Window state and actions

`ctx.WindowState()` reports size, focus, decoration, top-most state, and the
current window mode. `Program.WindowStateMessage` can map native changes to an
application message. `Application.Perform` sends actions such as minimize,
maximize, restore, fullscreen, raise, or center.

## Close lifecycle

Gio reports a native window-close request as `app.ClosingEvent` on Windows,
macOS, Wayland, and X11. FlowUI handles the event on the window event loop and
applies the configured close policy before the platform closes the window.

Use `window.OnWindowCloseRequest` for a synchronous decision:

```go
window.OnWindowCloseRequest(func() window.CloseDecision {
	if dirty {
		return window.CloseCancel
	}
	return window.CloseProceed
})
```

`CloseCancel` aborts the native request, `CloseProceed` lets it happen, and
`CloseKeepAlive` lets the native window close while keeping the process alive
for a tray or background service. An unknown decision fails closed.
`Close`, `CloseAll`, and `Quit` are force-close paths and bypass the handler.
The handler runs on the window event-loop goroutine, so it should return
promptly.

The handler has no `ui.Send`, so it cannot open an in-app confirmation dialog.
For a native close that must ask the user, set `Program.CloseRequestMessage`;
FlowUI aborts the request and queues the returned message instead:

```go
var application *ui.Application

program := ui.Program[Model, Msg]{
	// ...
	CloseRequestMessage: func() Msg { return CloseRequested{} },
}

func update(m *Model, msg Msg) ui.Cmd[Msg] {
	switch msg.(type) {
	case CloseRequested:
		if m.Dirty {
			m.ConfirmClose = true
		} else {
			application.Close("main")
		}
	case DiscardAndClose:
		m.ConfirmClose = false
		application.Close("main")
	}
	return nil
}

application = ui.NewApplication()
application.Run(ui.NewWindow("main", program, window.Theme(base.Light())))
```

The message hook applies only when no `OnWindowCloseRequest` handler is
configured; an explicit handler wins. When no handler is set,
`RequestClose` requests the native close, so application-side close controls
such as the title bar reach the hook as well. `Close` is the final forced close
and does not re-enter the hook. `ui.Run` rejects a program that sets
`CloseRequestMessage`, because it keeps its `Application` private; use
`ui.NewApplication` and `Application.Run` directly so `Update` can call
`Application.Close`. A `nil` hook leaves the native close alone.

macOS application termination can destroy a window without a `ClosingEvent`, so
that path cannot be vetoed through this contract.

See the multiple-windows and system-tray programs in the separate `examples`
module.
