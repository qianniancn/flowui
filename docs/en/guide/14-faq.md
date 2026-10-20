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

Use `resolve.Color` or `resolve.Brush` for theme-backed values, `draw.Fill` for
gradient fills, and `MeasureText` for a measurement pass. Low-level pointer
controls can use `AddPointerArea`, `NextPointerEvent`,
`IsPrimaryPointerPress`, and `GrabPointer`. `layout.VisualOverflow` is for a
local ripple or focus decoration; use `Popover`, `Portal`, or another overlay
component for content that is not part of the local layout. Use
`layout.VisualOutset` or `style.VisualOutset` when local paint also needs safe space
inside an ancestor scroll or split viewport.

## Minimum Go version

FlowUI currently requires Go 1.26.2 or newer, as declared by `go.mod`.

## Where are the examples?

Default-theme examples live in this module. From the repository root, run
`go run ./examples/counter`, `go run ./examples/buttons`,
`go run ./examples/inputs`, or `go run ./examples/style-gallery`. Other design
systems and their examples live in the external themes module; see
[Default theme and external styles](18-design-styles.md).
