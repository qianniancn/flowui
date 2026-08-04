# 14 - FAQ

## Which packages should an application import?

Use `github.com/qianniancn/flowui/ui` for the framework facade. Use
`uitest` only in tests. Do not import `internal/...` packages.

## Run or Application?

| Scenario | API |
| --- | --- |
| One window, with or without asynchronous work | `ui.Run(Program)` |
| Multiple windows, tray, or lifecycle control | `ui.Application` |

## Why did a button not respond?

Check that the callback sends a message, `Update` handles it, the key is stable,
and the widget is not disabled or covered by an overlay.

## Can View mutate the model?

No. View is read-only. Send a message and change the model in `Update`.

## Why is a controlled select stuck?

When `Open(m.Open)` is present, keep `Open` in the model, handle
`OnOpenChange`, and update the model from that callback. Otherwise the widget
cannot reflect its close request.

## Why did a theme change not affect a widget?

An instance style or variant may override a token. Check the cascade order and
avoid hard-coded colors when a semantic token is available.

## How should a custom widget handle drawing and pointer input?

Use `ResolveColor` or `ResolveBrush` for theme-backed values, `DrawBrush` for
gradient fills, and `MeasureText` for a measurement pass. Low-level pointer
controls can use `AddPointerArea`, `NextPointerEvent`,
`IsPrimaryPointerPress`, and `GrabPointer`. `LayoutVisualOverflow` is for a
local ripple or focus decoration; use `Popover`, `Portal`, or another overlay
component for content that is not part of the local layout.

## Minimum Go version

FlowUI currently requires Go 1.26.2 or newer, as declared by `go.mod`.

## Where are the examples?

```text
examples/counter          Minimal MVU
examples/form             Controlled fields
examples/async            Commands and subscriptions
examples/components       Component gallery
examples/fonts            System and bundled fonts
examples/grid_layout      Fixed and responsive grids
examples/multi_windows    Application lifecycle
examples/systray_ui       System tray integration
examples/profiling        Rendering and memory profiles
```

Run one from the repository root, for example:

```bash
go run ./examples/components
```
