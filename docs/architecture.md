# FlowUI Architecture

Implementation notes for the current tree: dependency direction, style cascade,
interaction host, state ownership, overlay host, and effects. For application
usage, see the [project Wiki](https://github.com/qianniancn/FlowUI/wiki) and the public `ui` package docs.

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

Dependencies only point downward. High-traffic public types are **facade-owned
true types, defined types, or thin wrappers** (`Style`, `Context`, `Theme`,
`ButtonWidget`, `MenuItem`, enums, …) so `go doc` lists methods and DTO fields
use `ui.Widget` rather than internal `frame.Widget`.

The remaining `=` aliases fall into three groups: predicate/snapshot helpers
(`StyleState`, `Condition`) and error types that must keep their method sets;
and — the largest group — structural **style and render data types** that
applications read as plain data (`ResolvedStyle`, `PaintSource`, `ColorSource`,
`SolidColor`, `ThemeColor`, `Gradient`, `Insets`, `BoxStyle`, `PaintStyle`,
`CornerRadii`, `StyleShadow`, `OutlineStyle`, `BorderStyle`, `TextStyle`,
`TransformStyle`, `StyleTransition` in `ui/style.go`; `ShadowCornerRadii`,
`ShadowShape`, `ShadowLayer`, `RenderBoxShadow` in `ui/shadow.go`). These
data types are aliased so their fields stay directly usable through the facade;
`go doc ui.BoxStyle` therefore resolves to `internal/style`. Behavioral types
(widgets, context, style builder) remain facade-owned; only field-bearing data
types are aliased.

All implementation packages live below `internal/`. Theme, locale, layout, and
drawing types needed by applications are re-exported through `ui`; their
implementation paths are not additional public entry points.

## Style Cascade

`ui.Style` is the public immutable declaration owned by package `ui` (true type,
not an alias). Public declarations start from an attribute function such as
`ui.Background`, `ui.Padding`, or `ui.Part`, then continue through value-returning
fluent methods. Components use the same immutable chain from a zero-value
`Style`; there is no separate builder or build step. The runtime resolves
declarations in this order (`ResolveStatic` / `Resolve`):

```text
component defaults
  -> inherited text (parent TextDeclaration stack)
  -> variant
  -> size
  -> StyleScope ancestors
  -> instance Style
```

Then metric tokens expand, `Cascade` applies matching `When` rules (later
properties win), color tokens resolve, and optional transitions run under the
component key.

`When` stores predicates for interaction state (`Hovered`, `Pressed`,
`FocusVisible`, `Dragging`), semantic state (`Disabled`, `Checked`, `Invalid`,
`Loading`), or an MVU model value adapted with `If(bool)`. `StyleScope` carries
declarations down the widget tree; it never mutates the application theme.

```text
layers -> ExpandTokens -> Cascade(state) -> ResolveColors -> Transitions(key)
                              |
                              +-> root + named Parts (CascadePart)
```

Animation is dual-track: Style `Transition` (property whitelist under a stable
key) and imperative `Tween` for floats. Both share `Ease*` curves and honor
`Theme.Motion`. Application-facing animation notes: [动画](https://github.com/qianniancn/FlowUI/wiki/12-动画).

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
widget calls `ui.BeginInteract`, then `ui.Resolve` for its root and
`ui.ResolvePart` for internals, then renders with `ui.LayoutBox` or
`ui.LayoutInteractiveBox`. The latter keeps margin outside the input
host while making padding and the remaining visual box part of the hit area.
`Interact.Clickable` is the focus tag for `RequestFocus` / `FocusVisible`.
`ui.StyleScope` styles descendants. Component-level theme mutation is
intentionally not part of the API.

Transition state needs stable widget identity. Stateful components claim their
component key; each non-interactive transitioning sibling uses a distinct
`ui.Key` scope. A `Box` may use its own key directly.

Parent-owned layout policy stays outside Style: flex growth, sibling gap,
alignment, absolute/portal placement, z-order, and scroll state belong to their
layout containers. This prevents CSS-like syntax from hiding MVU state or
changing ownership boundaries.

`ui.Box` follows the same split. Box-model geometry (width, height, min/max,
fill, padding, margin, overflow) and paint live only in `Style` (including
`StyleScope`). The widget itself keeps identity (`Key`), interaction
(`OnClick` / `Disabled` / `Label`), and content alignment (`Align`).

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

For application-facing overlay behavior, see [浮层与弹出](https://github.com/qianniancn/FlowUI/wiki/09-浮层与弹出).

Input ownership for overlay *policy* events (dismiss, Escape, exclusive
open/close) is owned by at most one non-passive overlay:

- When the preceding frame had an owner, `BeginFrame` freezes that identity as
  the only overlay that receives `interactive=true` during the paint pass
  (ownership lags the visual top by one frame).
- When there is no preceding owner (first open, or after the stack clears),
  every overlay paints with `interactive=false`, then only the final visual top
  is re-entered with `interactive=true` on a throwaway ops list. Key claims from
  the paint pass are reused; nested `RegisterOverlay` is ignored during that
  policy pass. This prevents a first-frame modal and its nested popup from both
  consuming Escape.

Geometric pointer routing can still use deliberate pass-through behavior. After
dynamic overlay layout completes, the host records the new visual top and the
nearest modal focus scope. Focus work runs against that completed state, while
policy ownership transfers on the next frame when an owner already exists so
open/close races cannot double-consume one event.

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
workflow when the model no longer needs its result. Keys are retained for a
window lifetime so delayed older closures remain rejectable; use bounded,
stable workflow keys rather than request-specific values.

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
filesystem watcher, or server event stream. `RunProgram` with
`Program.Subscriptions` derives the
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
rather than continuing with a potentially partially-mutated model. When the
window can request a native close, FlowUI cancels effects, issues that close,
and drains events until `DestroyEvent` before returning the panic error so the
OS shell is not left as a live zombie. Event sources without a close path
(such as headless harnesses) return the panic error immediately.

Within a single frame's message batch, a panic in message *k* leaves earlier
mutations applied, discards later messages in that batch, and does not start
commands for the panicking or later messages. Commands already started from
earlier messages are canceled when the window session ends.

`ui.Batch` starts several independent commands from one `UpdateCmd` return.
Each command runs concurrently under the same effect context and capture rules
as a single command.

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

## Internal Interface Design

Internal packages serve as building blocks for components and must be robust
against misuse. The following principles guide internal API design to prevent
subtle bugs from incorrect usage patterns.

### Type Safety Over Documentation

Encode invariants in types and signatures rather than relying on documentation
or runtime checks:

- **Bundle synchronized values**: If two fields must stay in sync (like a cache
  and its generation), put them in one struct rather than separate variables
- **Use the type system**: Prevent invalid states through types, not runtime
  validation
- **Make dependencies explicit**: If function B needs the result of function A,
  make A's result a parameter to B

Example: `KeyFilterCache` bundles `target`, `names`, and `filters` together.
The cache key includes both `target` and `names`; omitting either would cause
silent routing errors.

### Eliminate Call-Order Dependencies

Methods should work correctly regardless of invocation order unless the
dependency is explicit in the signature:

- **Idempotent by default**: Calling a method twice with the same arguments
  should produce the same result
- **Order independence**: If two methods don't have a data dependency, they
  should work in any order
- **Explicit sequencing**: When order matters, use:
  - A builder pattern with a terminal operation
  - Method chaining where each step returns the next valid state
  - Parameters that carry results from required prior steps

Example: `ClickArea.Layout` previously depended on being called after
`Clicked`, which drained events. Now both methods process events identically,
so call order doesn't matter.

### Separate Queries from Commands

Methods that return values should not have side effects unless the side effect
is the primary purpose:

- **Query methods are pure**: Methods named like `Get*`, `Is*`, `Visible`,
  `Contains`, or returning a value without changing receiver state should have
  no side effects
- **Command methods change state**: Methods named like `Set*`, `Add*`,
  `Commit*`, or clearly mutating should have side effects as their primary job
- **Split when necessary**: If a method needs both, split it into a query phase
  and a command phase

Example: `Focus.Observe` records frame-local focus input while
`Focus.CommitObservations` applies persistent modality changes at the frame
boundary. The old `Visible` method remains for internal compatibility.

### Complete Cache Keys

When designing caches, ensure the cache key includes every input that affects
the result:

- **All dependencies**: If the result depends on `(A, B, C)`, the cache key
  must include all three
- **Explicit versioning**: Use version tokens (`DataVersion`, theme generation)
  to invalidate when indirect dependencies change
- **Test with variation**: Test caches by varying each input independently, not
  just the common case

Example: `ComboBoxState.visibleItems` caches based on `(query, selectedLabel,
DataVersion)`. Omitting any of these would cause stale results when that input
changes.

### Fail Loudly, Not Silently

When invariants are violated, fail immediately with a clear error rather than
continuing with corrupt state:

- **Panic on impossible states**: Use `panic` for internal invariant violations
  that indicate programming errors
- **Return errors for expected failures**: Use `error` returns for failures
  that callers should handle (I/O, validation, user input)
- **Validate early**: Check preconditions at the start of a method, not after
  partial state changes

Example: `state.Keys.claim` panics on duplicate keys with a clear message
rather than silently overwriting the previous registration.

### Design Review Questions

When reviewing or designing internal APIs, ask:

- Can this be misused by changing call order?
- Does this method name promise a query but perform writes?
- Are all cache dependencies reflected in the cache key?
- Will incorrect usage fail loudly or silently corrupt state?
- Could two callers interfere with each other's state?
- Are there implicit assumptions that should be explicit parameters?

These principles emerged from real issues encountered during development:
- `KeyFilterCache` ignored `target`, causing potential routing errors (#12)
- `ClickArea.Layout` had hidden call-order requirements (#13)
- `Focus.Visible` mixed query and command concerns (#14)

Internal APIs built with these principles are easier to use correctly, harder
to misuse, and fail early when misused.
