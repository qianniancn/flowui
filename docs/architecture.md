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
such as clickables, draggables, editors, focus, animation progress, and overlay coordination.
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

## Overlay Host

Popup and modal components register work with the per-window root Overlay Host.
The frame order is:

```text
BeginFrame
    -> layout the main widget tree and register overlays
    -> layout root overlays in viewport coordinates
    -> apply focus commands
EndFrame
```

The host keeps root popup branches below modal branches, preserves an overlay's
opening order across frames, and processes overlays registered by another
overlay in the same frame. Within a modal branch, a descendant popup is painted
above the modal that registered it but below any pending nested modal. Panel
content is still laid out before `EndFrame`, so keyed state, focus requests,
animations, and MVU callbacks retain normal same-frame semantics.

Input ownership is intentionally one frame behind painting. `BeginFrame`
freezes the preceding frame's top overlay as the only overlay allowed to
process overlay-level queued events such as dismiss, Escape, and selection.
Geometric routing can still use deliberate pass-through behavior. After
dynamic overlay layout completes, the host
records the new visual top and the nearest modal focus scope. Focus work runs
against that completed state, while pointer and key ownership transfers on the
next frame. This prevents one event from being consumed by two overlays when a
popup opens, closes, or changes its nesting in the same frame.

Overlay blockers use pointer-only click regions. They keep panel padding,
arrows, backdrops, and exit animations from passing clicks to the main tree,
but never enter Gio's keyboard focus order. A top modal contributes a tail
focus boundary after all of its dynamic descendants, so focus cannot escape
through a nested Select or Popover. Only the nearest focus-scope ancestor of
the visual top contributes that boundary.

Gio does not expose the transform at a widget's current layout position.
FlowUI's layout containers therefore attach a transform node while measuring a
child and fill in that node after the child's final position is known. An
overlay anchor follows this transform chain to the viewport, including nested
Box, Flex, Grid, Stack, Wrap, List, Scroll, and Tabs layouts. Position flipping
and overflow avoidance then use the real viewport-relative anchor.

The same chain carries local clipping and animated opacity. Scroll/List clips
decide whether an anchor is still visible without replacing the full anchor
used for placement. A child overlay inherits the opacity of every animated
overlay surface that owns it, so nested content exits with its parent instead
of remaining fully opaque. Transform nodes live in a frame-local index arena;
layout containers can track every measured child without one heap allocation
per node. Containers that accept repeated items must use their tracked item
layout path so each item's final offset and clip are recorded.

A custom Widget that places a FlowUI child with raw Gio layout, transform,
clip, or `paint.PushOpacity` operations is outside this tracking contract. Use
FlowUI layout containers and component animation paths for subtrees that can
open overlays; otherwise the custom widget must keep those overlay-opening
children at its local origin and full opacity. This limitation follows from
Gio's public API and must not be worked around with pointer coordinates or
private operation decoding, because those approaches fail for keyboard and
programmatic opening.

## Effects

`UpdateCmd` is serialized on the event loop and is the only place a command
workflow may mutate the model. Each returned `Cmd` starts in its own goroutine
after `UpdateCmd` returns, so it can overlap later updates and views. Every
command receives a root `context.Context` that is canceled when the window is
destroyed. Blocking and fallible work should use `ui.DoContext`; `ui.Do` remains
a convenience for short context-free work.

A command must capture only immutable value snapshots prepared during
`UpdateCmd`. It must not retain or access the model pointer or a `ui.Context`.
Reference-backed values such as slices and maps need an explicit copy before
capture. A command returns data to the application only by calling its `Send`
argument; concurrent sends are queued safely and applied by `UpdateCmd` in a
later frame. Sends made after cancellation are discarded.

`Subscription` represents long-lived asynchronous input such as a timer,
filesystem watcher, or server event stream. `RunWithSubscriptions` derives the
desired subscription set from the updated model before each view. Keys are
lifecycle identities: a stable key retains the running subscription, removal
cancels it, and a changed key starts a replacement. Duplicate or empty keys are
programming errors. A subscription should remain running until its context is
canceled. After normal or failed completion, its key remains retained and is
not restarted every frame; remove and re-add it or change the key to retry.
When keys change, FlowUI cancels removed runners and waits asynchronously for
them to stop before starting replacements. A 250ms grace bound prevents an
uncooperative runner from blocking replacement indefinitely. Messages carry a
subscription generation and are validated on the event thread, so messages
queued by a stopped generation cannot reach `UpdateCmd`.

Commands and subscriptions may return errors. The runtime recovers their
panics, attaches a stack trace, and reports both failures as `EffectError`
values to the `OnError` handler on the application event thread. Cancellation
errors are ignored. Without an `OnError` option, FlowUI writes the error and any
panic stack to standard error. The handler is for logging and unexpected
infrastructure failures; expected domain failures should normally be converted
into typed messages so `UpdateCmd` can represent them in the model.

Window shutdown cancels every command and subscription and waits up to two
seconds for their cleanup before returning. A timeout is reported as
`ErrEffectShutdownTimeout`; the process is still allowed to exit after that
bound, so effects must honor their context promptly.

`View` and a `Subscriptions` function follow the same ownership boundary: the
model and any reference-backed fields it contains are read-only. Event
callbacks and effects send messages instead of mutating captured model data.

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
