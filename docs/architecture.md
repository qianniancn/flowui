# FlowUI Architecture

FlowUI exposes one application-facing package:

```go
import "github.com/qianniancn/FlowUI/ui"
```

Component tests may additionally import `github.com/qianniancn/FlowUI/uitest`.
The package owns a deterministic FlowUI context, Gio input router, viewport,
and clock. Its `Harness.Frame` method follows the same main layout, root
overlay, focus-command, state-cleanup, and router order as a real window.
Applications do not depend on `uitest` at runtime.

`uitest.AppHarness` drives the production message queue, `UpdateCmd`, command
execution, error delivery, and shutdown cancellation without opening a window.
It complements the widget Harness rather than maintaining a second MVU runtime.

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

component tests
    |
    +--> uitest
            |
            +--> ui
            +--> internal/frame
            +--> internal/theme
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

## Style Cascade

`ui.Style` is the public immutable declaration used by components and custom
widgets. Public declarations start from an attribute function such as
`ui.Background`, `ui.Padding`, or `ui.Part`, then continue through value-returning
fluent methods. Components use the same immutable chain from a zero-value
`style.Style`; there is no separate builder or build step. The runtime resolves
declarations in this order:

```text
component defaults -> variant -> size -> StyleScope ancestors -> instance Style
```

`When` stores predicates for interaction state (`Hovered`, `Pressed`,
`FocusVisible`, `Dragging`), semantic state (`Disabled`, `Checked`, `Invalid`,
`Loading`), or an MVU model value adapted with `If(bool)`. `Cascade` is the pure
merge step: matching rules are applied and later properties take precedence.
`Resolve` adds the active scopes, expands theme metric/color tokens, and retains
transition state under the component key. `StyleScope` carries declarations
down the widget tree; it never mutates the application theme.

```text
Style declaration -> Cascade(state) -> Resolve(theme, key) -> LayoutResolvedStyle
                            |                    |
                            |                    +-> transition state
                            +-> root + named Parts
```

## Interaction Core

Button-like controls share one internal interaction core for clicks, keyboard
activation, focus, disabled state, and accessibility semantics. `ui.Button`
adds FlowUI defaults, variants, sizes, loading, and compound content. A custom
control uses `ui.Box(...).Key(...).OnClick(...)`; its interaction state is
available to `Style.When` without inheriting Button visuals. Raw Gio events
remain available for interactions such as dragging that do not share button
semantics.

Root `Box`, `Paint`, `Text`, and `Transform` properties always describe the
component's outer box. A compound component exposes internal elements through
`Part`: content, labels, descriptions, icons, tracks, fills, thumbs,
indicators, panels, items, backdrops, placeholders, selections, prefixes, and
suffixes. Compound field controls use `PartContent` for the field face. Part
text inherits root text properties; box, paint, and transform properties do
not. Applications may use a custom `StylePart` string for their own components.

The common renderer owns margin, constraints, aspect ratio, padding, overflow
clipping, cursor, background/gradient, per-corner radius, shadows, border,
outline, opacity, and transforms. Components must not duplicate a partial
adapter for these properties. They retain only domain-specific geometry such as
progress ratios, slider thumb positions, or chart paths.

Components use the same runtime for root layout and interaction state. A custom
widget calls `ui.ResolveStyle` for its root, `ui.ResolveStylePart` for internals,
then renders with `ui.LayoutResolvedStyle` or
`ui.LayoutInteractiveResolvedStyle`. The latter keeps margin outside the input
host while making padding and the remaining visual box part of the hit area.
`ui.StyleScope` styles descendants. Component-level theme mutation is
intentionally not part of the API.

Transition state needs stable widget identity. Stateful components claim their
component key; each non-interactive transitioning sibling uses a distinct
`ui.Key` scope. A `Box` may use its own key directly.

Parent-owned layout policy stays outside Style: flex growth, sibling gap,
alignment, absolute/portal placement, z-order, and scroll state belong to their
layout containers. This prevents CSS-like syntax from hiding MVU state or
changing ownership boundaries.

## State Ownership

Application and business state belongs in the user's `Model`. A controlled
component receives values from the model and reports intent through callbacks;
the callback sends a typed message, and `Update` mutates the model.

`ui.Context` owns only per-window Gio services and transient interaction state,
such as clickables, draggables, editors, focus, animation progress, and overlay coordination.
It must not become a second application model.

The public context exposes read-only theme and language snapshots plus the small
set of Gio state and focus helpers needed by custom widgets. Runtime theme and
language changes go through `Application.SetTheme` and `Application.SetLanguage`
so they are applied on the target window's event loop. Frame lifecycle and
component registration remain internal; root overlays are exposed only through
the constrained Portal API.

Custom widgets implement `ui.Widget` directly or use `ui.WidgetFunc`. Transient
Gio interaction state can use `ui.UseState` or `ui.UseStateWith`; it is retained
while its explicit key is rendered and released automatically after removal.
Business values still belong in the Model. Public focus methods on `ui.Context`
keep custom focus rings consistent with pointer and keyboard modality.

`ui.Portal` is the low-level escape hatch for custom root-level content. It
provides a resolved viewport anchor, stacking group, and preceding-frame input
ownership. It intentionally does not add positioning, dismissal, animation,
backdrops, or focus trapping. Applications should prefer Popover, Modal,
Tooltip, and Menu when those policies fit. Portal content may register nested
FlowUI overlays normally.

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
clip, or `paint.PushOpacity` operations must wrap that child with
`ui.TrackOverlayPlacement` and record the matching offset, transform, clip, and
opacity on the returned frame-local placement. FlowUI layout containers already
do this automatically. Pointer coordinates and private operation decoding must
not be used as substitutes because they fail for keyboard and programmatic
opening. Custom layouts must call `PlaceOffset` or `PlaceTransform` after the
tracked child returns, including when the final transform is the identity.

## Effects

`ui.Program` groups `Init`, `Update`, `Subscriptions`, and `View` when an
application needs the complete lifecycle. `Init` runs once per window instance
and its optional command is managed by the same effect group as commands
returned by `UpdateCmd`. `WindowStateMessage`, when present, maps Gio
configuration changes into queued messages for the next update.

`UpdateCmd` is serialized on the event loop and is the only place a command
workflow may mutate the model. Each returned `Cmd` starts in its own goroutine
after `UpdateCmd` returns, so it can overlap later updates and views. Every
command receives a root `context.Context` that is canceled when the window is
destroyed. Blocking and fallible work should use `ui.DoContext`; `ui.Do` remains
a convenience for short context-free work. `ui.LatestCmd` cancels an older
same-key command and rejects messages from older generations; use it for
search, preview, and autocomplete work. `ui.CancelLatestCmd` cancels a keyed
workflow when the model no longer needs its result.

A command must capture only immutable value snapshots prepared during
`UpdateCmd`. It must not retain or access the model pointer or a `ui.Context`.
Reference-backed values such as slices and maps need an explicit copy before
capture. A command returns data to the application only by calling its `Send`
argument; concurrent sends are queued safely and applied by `UpdateCmd` in a
later frame. The runtime bounds the message queue to 256 entries. Ordinary
messages beyond the bound are dropped and reported as `QueueOverflowError`; if
a subscription stream is full, a queued value from the same generation is
replaced by the latest value. The error queue is bounded to 64 entries as well.
Sends made after cancellation are discarded.

`ui.MapCmd` lets a parent update reuse a child module's command by mapping each
child message to the parent message type. It preserves the command context and
error unchanged; it does not introduce another goroutine or effect lifecycle.

`Subscription` represents long-lived asynchronous input such as a timer,
filesystem watcher, or server event stream. `RunWithSubscriptions` derives the
desired subscription set on the initial frame and after an updated model;
stable desired data is reused on animation-only frames. Keys are
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
Panics from `UpdateCmd`, `Subscriptions`, or `View` are recovered as
`RuntimePanicError` values with a phase and stack trace; the window loop stops
rather than continuing with a potentially partially-mutated model.

Window shutdown cancels every command and subscription and waits up to two
seconds for their cleanup before returning. A timeout is reported as
`ErrEffectShutdownTimeout`; the process is still allowed to exit after that
bound, so effects must honor their context promptly.

`View` and a `Subscriptions` function follow the same ownership boundary: the
model and any reference-backed fields it contains are read-only. Event
callbacks and effects send messages instead of mutating captured model data.

Gio v0.10.1 reports a native window close only as the final `DestroyEvent`; it
does not expose a cancellable close-request event. FlowUI therefore cannot
reliably veto operating-system close actions. An application that requires an
unsaved-changes guard must keep the close action in application-controlled UI
and perform the native close only after confirmation.

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
