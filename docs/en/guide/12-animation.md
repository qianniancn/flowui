# 12 - Animation

FlowUI provides two animation tracks that share easing curves and
`Theme.Motion` settings:

| Track | API | Use it for |
| --- | --- | --- |
| Declarative | `Style.Transition` | Colors, opacity, and transforms |
| Value-driven | `Tween`, `Spring`, `Timeline` | Values, layout, and keyframes |

Both tracks require a stable key.

## Style transitions

```go
ui.Background(ui.TokenSurface).
	Transition(ui.PropBackgroundColor, 120*time.Millisecond).
	When(ui.Hovered, ui.Background(ui.TokenSurfaceRaised))
```

Add an easing option when the default curve is not appropriate:

```go
.Transition(ui.PropOpacity, 200*time.Millisecond,
	ui.TransitionEase(ui.EaseCubicOut))
```

## Tween

Read a keyed value in a custom layout or widget function:

```go
target := float32(0)
if expanded {
	target = 1
}
progress := ui.Tween("panel-open", target).
	Duration(200 * time.Millisecond).
	Easing(ui.EaseCubicOut).
	Value(ctx, gtx)
```

`Sample` also reports whether the value is still moving. Springs are useful
for physical-feeling layout changes:

```go
progress := ui.Tween("panel-open", target).
	Spring(ui.SpringSnappy()).
	Value(ctx, gtx)
```

`Timeline` provides multiple keyframes. `AnimateRect` interpolates rectangle
bounds, and `AnimateLayout` interpolates child size changes.

## Reduced motion

Respect `Theme.Motion` and avoid hard-coding large, uninterruptible effects.
Applications can disable motion or change the duration scale in a customized
theme. Use tokens and default transitions so this policy applies consistently.

The [`examples/animations`](https://github.com/qianniancn/flowui/tree/main/examples/animations) program demonstrates
transitions, tweens, springs, timelines, and layout movement.
