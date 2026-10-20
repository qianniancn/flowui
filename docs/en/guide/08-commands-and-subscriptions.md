# 08 - Commands and subscriptions

`Cmd` represents one asynchronous effect. It receives a context and a typed
`Send` function, and returns an error when the effect fails.

## Commands

```go
func loadCmd() ui.Cmd[Msg] {
	return ui.DoContext(func(ctx context.Context, send ui.Send[Msg]) error {
		value, err := load(ctx)
		if err != nil {
			return err
		}
		send(Loaded{Value: value})
		return nil
	})
}
```

`ui.Do` is useful when the effect does not need cancellation details. Keep
inputs immutable: copy slices and maps before capturing them in a command.

## Update and effects

Update changes state first, then returns a command for work that happens outside
the event loop:

```go
func Update(m *Model, msg Msg) ui.Cmd[Msg] {
	switch msg := msg.(type) {
	case Submit:
		m.Busy = true
		name := m.Name
		return submitCmd(name)
	case Submitted:
		m.Busy = false
	}
	return nil
}
```

Commands may send zero, one, or several messages. They should stop when their
context is canceled.

Use `ui.Batch` when one update starts independent effects that can run in
parallel. Nil commands are ignored and the first returned error is reported
after all active commands finish:

```go
return ui.Batch(loadProfileCmd(id), loadPermissionsCmd(id))
```

## Latest command

For search or other replaceable work, use `LatestCmd` with a stable workflow
key. Starting a newer command cancels the previous one:

```go
return ui.LatestCmd("search", searchCmd(query))
```

## Subscriptions

Subscriptions describe long-lived inputs for the current model:

```go
Subscriptions: func(m Model) []ui.Subscription[Msg] {
	if !m.ClockEnabled {
		return nil
	}
	return []ui.Subscription[Msg]{
		ui.Subscribe("clock", func(ctx context.Context, send ui.Send[Msg]) error {
			ticker := time.NewTicker(time.Second)
			defer ticker.Stop()
			for {
				select {
				case <-ctx.Done():
					return ctx.Err()
				case now := <-ticker.C:
					send(Tick{At: now})
				}
			}
		}),
	}
}
```

Subscription keys are retained after a run ends; change the requested set to
restart work intentionally.

Nested modules can adapt a child subscription with `ui.MapSubscription`. It
preserves the child's lifecycle key while mapping emitted messages to the
parent type.

## Errors and shutdown

Use `window.OnError` to receive effect, runtime, and window lifecycle errors. The
window context is canceled when that window closes, which stops its commands
and subscriptions.

Combine these command snippets with your application and continue with
[Multiple windows](11-multiple-windows.md).
