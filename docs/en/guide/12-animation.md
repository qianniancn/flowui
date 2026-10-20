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
style.New().Background(style.Solid(activeTheme.Palette.Surface)).
	Transition(style.PropBackgroundColor, 120*time.Millisecond).
	On(style.Hovered, style.New().Background(style.Solid(activeTheme.Palette.SurfaceRaised)))
```

Add an easing option when the default curve is not appropriate:

```go
.Transition(style.PropOpacity, 200*time.Millisecond,
	style.TransitionEase(animation.EaseCubicOut))
```

## Tween

Read a keyed value in a custom layout or widget function:

```go
target := float32(0)
if expanded {
	target = 1
}
progress := animation.Tween("panel-open", target).
	Duration(200 * time.Millisecond).
	Easing(animation.EaseCubicOut).
	Value(ctx, gtx)
```

`Sample` also reports whether the value is still moving. Springs are useful
for physical-feeling layout changes:

```go
progress := animation.Tween("panel-open", target).
	Spring(animation.SpringSnappy()).
	Value(ctx, gtx)
```

`Timeline` provides multiple keyframes. `AnimateRect` interpolates rectangle
bounds, and `AnimateLayout` interpolates child size changes.

## Reduced motion

Respect `Theme.Motion` and avoid hard-coding large, uninterruptible effects.
Applications can disable motion or change the duration scale in a customized
theme. Use tokens and default transitions so this policy applies consistently.

The button example (`go run ./examples/buttons`) demonstrates control feedback.
Tests for tweens, springs and timelines live under `animation/`; combine the
usage snippets above with your application for custom motion.
