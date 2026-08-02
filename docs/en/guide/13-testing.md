# 13 - Testing

The `uitest` package tests FlowUI without opening a real window:

```go
import "github.com/qianniancn/flowui/uitest"
```

Applications do not need `uitest` at runtime.

## Widget harness

`uitest.Harness` drives layout frames, a fixed viewport, pointer and keyboard
events, a controllable clock, and state cleanup:

```go
func TestButtonClick(t *testing.T) {
	h := uitest.New(image.Pt(400, 300))
	clicked := false
	root := ui.Button("ok", ui.Text("OK")).OnClick(func() {
		clicked = true
	})

	// Layout a frame, inject a pointer event, and layout again.
	_ = h
	_ = root
	_ = clicked
}
```

Use stable keys and advance the harness clock for animation tests instead of
sleeping. Prefer behavior assertions to screenshot-only tests.

## Application harness

`uitest.AppHarness` exercises the production message queue, `Update`, command
execution, error delivery, and close cancellation without creating a window.
It is suited to testing loading success/failure and lifecycle behavior.

## Test guidance

1. Test pure `Update` logic directly.
2. Use `Harness` for interaction, layout, and style states.
3. Use `AppHarness` for commands, subscriptions, and errors.
4. Use business IDs as keys so repeated controls retain the right state.
5. Keep external IO behind command boundaries and inject test doubles.

See the tests under [`uitest`](https://github.com/qianniancn/flowui/tree/main/uitest) and the component tests under
`internal/components/*`.
