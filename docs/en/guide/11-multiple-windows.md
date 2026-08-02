# 11 - Multiple windows

Use `ui.Run(Program)` for one window. Use `ui.Application` when the process
owns multiple windows, a system tray, or an explicit lifecycle.

## Basic structure

```go
func main() {
	application := ui.NewApplication()

	counter := ui.NewWindow("counter", ui.NewProgram(
		CounterModel{}, counterUpdate, counterView,
	), ui.Title("Counter"), ui.Size(400, 300))

	mainWindow := ui.NewWindow("main", ui.NewProgram(
		MainModel{}, mainUpdate, mainView,
	), ui.Title("Main"), ui.Size(640, 480))

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
application.SetTheme("main", ui.DarkTheme())
application.SetLanguage("main", ui.LanguageChinese)
```

Do not mutate a window's theme from an arbitrary goroutine.

## Window state and actions

`ctx.WindowState()` reports size, focus, decoration, top-most state, and the
current window mode. `Program.WindowStateMessage` can map native changes to an
application message. `Application.Perform` sends actions such as minimize,
maximize, restore, fullscreen, raise, or center.

## Close lifecycle

`OnWindowCloseRequest` handles FlowUI close requests. `WindowCloseCancel` keeps
the window open; `WindowCloseKeepAlive` closes the native window while keeping
the process alive for a tray or background service. Native window manager close
events cannot all be intercepted on every platform.

See [`examples/multi_windows`](https://github.com/qianniancn/flowui/tree/main/examples/multi_windows) and
[`examples/systray_ui`](https://github.com/qianniancn/flowui/tree/main/examples/systray_ui).
