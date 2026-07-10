# FlowUI Architecture

FlowUI exposes one application-facing package:

```go
import "github.com/qianniancn/FlowUI/ui"
```

The `ui` package is the MVU entry point and public facade. Applications should
not import `internal` packages, and lower-level FlowUI packages must not import
`ui`.

## Dependency Direction

```text
application
    |
    v
ui (MVU entry point and public facade)
    |
    +--> internal/runtime
    +--> internal/components/*
             |
             +--> internal/frame
             +--> internal/overlay
             +--> internal/state
             +--> internal/theme
             +--> internal/locale
             +--> internal/render
```

Dependencies only point downward. The facade uses type aliases, constants, and
small forwarding constructors so application code keeps a compact API without
placing component implementations in one large package.

The aliases preserve the concrete fluent APIs without wrapper duplication, but
Go's `go doc` does not currently enumerate methods promoted through aliases.
Changing to facade-owned wrapper types would be a public API migration and must
be handled as a separate, versioned change rather than mixed into package moves.

All implementation packages live below `internal/`. Theme, locale, layout, and
drawing types needed by applications are re-exported through `ui`; their
implementation paths are not additional public entry points.

## State Ownership

Application and business state belongs in the user's `Model`. A controlled
component receives values from the model and reports intent through callbacks;
the callback sends a typed message, and `Update` mutates the model.

`ui.Context` owns only per-window Gio services and transient interaction state,
such as clickables, editors, focus, animation progress, and overlay coordination.
It must not become a second application model.

The public context exposes read-only theme and language snapshots plus the small
set of Gio state helpers needed by custom widgets. Frame lifecycle, component
registration, focus coordination, and overlay coordination remain internal.

Component identity is explicit. A key is scoped by `ui.Key`, claimed with a
component kind, and paired with a typed state slot. Identity must not depend on
render order or an implicit widget-tree path.

FlowUI layout containers pre-register keyed Label and Description associations,
so their semantic relationship is independent of visual layout order. A custom
composite Widget that directly lays out children after the control uses an
invalidated follow-up frame as a compatibility fallback; laying associations
before the control gives same-frame semantics.

Deferred overlays currently position panels in the trigger's local coordinate
space. Negative coordinates are therefore valid for top and left placements.
Fully viewport-aware flipping for arbitrarily nested or custom layouts requires
a future root overlay host with absolute anchor coordinates.

## Command Concurrency

`UpdateCmd` is serialized on the event loop and is the only place a command
workflow may mutate the model. Each returned `Cmd` starts in its own goroutine
after `UpdateCmd` returns, so it can overlap later updates and views.

A command must capture only immutable value snapshots prepared during
`UpdateCmd`. It must not retain or access the model pointer or a `ui.Context`.
Reference-backed values such as slices and maps need an explicit copy before
capture. A command returns data to the application only by calling its `Send`
argument; concurrent sends are queued safely and applied by `UpdateCmd` in a
later frame.

`View` follows the same ownership boundary: the model and any reference-backed
fields it contains are read-only while building and laying out widgets. Event
callbacks send messages instead of mutating captured model data.

## Adding A Component

1. Put the implementation and focused tests in `internal/components/<domain>`.
2. Depend on `internal/frame`, `internal/state`, `internal/theme`,
   `internal/locale`, and `internal/render` as needed.
3. Add aliases, exported constants, and forwarding constructors to a focused
   facade file under `ui`.
4. Keep model values controlled and report changes through callbacks.
5. Add or extend a runnable example under `examples`.
6. Run `go test ./...` and `go vet ./...`.

Component packages follow ownership boundaries, not a fixed "one directory per
widget" rule. Closely coupled controls may share a domain package when they own
the same interaction primitives. For example, input-like controls share field
state and Select reuses ListBox item behavior. Split a domain only after its
shared primitives can move into a lower package without introducing dependency
cycles. File count alone is not a package boundary.

The root module intentionally contains no Go package. This prevents two public
entry points from drifting apart and keeps `github.com/qianniancn/FlowUI/ui` as
the stable import contract.
