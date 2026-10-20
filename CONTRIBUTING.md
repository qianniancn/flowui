# Contributing to FlowUI

FlowUI is a Go desktop UI framework built on Gio. Keep changes focused on the
public APIs, optional platform services, component behavior, theme recipes,
rendering, tests, and examples.

Before changing code, read:

- [`README.md`](README.md) for installation and usage.
- [Project Wiki](https://github.com/qianniancn/flowui/wiki) — user-facing tutorial (final API).
- [`docs/architecture.md`](docs/architecture.md) — dependency direction, state
  ownership, and overlay behavior for the current tree.
- `go doc github.com/qianniancn/flowui/ui` and the focused public package docs
  for current usage.

## Development Environment

- Go 1.26.2 or newer
- A desktop platform supported by Gio

The repository root intentionally contains no Go package. Run checks from the
repository root. Runnable examples live under `examples/` in the same Go module.

## Code Changes

- Applications import `ui` for MVU, `widget/<component>` for official components
  (one package per component, e.g. `widget/button`), `layout`
  for composition, `style` for declarations, `theme/base` for the default theme,
  `theme` for visual contracts, `draw` for custom rendering, and `window` for
  native options. `explorer`, `notify`, and
  `systray` are optional platform services. Shared implementation code belongs
  below `internal/`.
- Keep business state in the application's model. Component state is for
  interaction and derived rendering state.
- Use explicit keys for retained component state. Reuse existing state,
  animation, layout, theme, and rendering helpers before adding new ones.
- Keep controlled components controlled: report user intent through callbacks
  or messages instead of mutating application data inside a widget.
- Avoid new dependencies and abstractions unless the existing code cannot meet
  the requirement.

When adding or changing a component:

1. Put behavior, layout, public constructors, and focused tests in
   `widget/<component>`. Applications import that package directly; `ui/` owns
   the MVU/runtime API.
2. Keep visual requests, results, Provider contracts, and validators together
   in `theme/<component>_visual.go`. Implement shared visual recipes in
   `theme/<component>style`, and wire them once through `theme/base.Provider`.
   External themes inherit the shared recipe and override it only when needed.
3. Update the matching fixture in `internal/testtheme/<component>.go` when the
   behavior tests need a new visual capability. Keep these fixtures independent
   of production recipes. Use `theme/themetest` to check production Providers.
4. Update `examples/` when a public usage change needs a runnable demonstration.
   Keep new example-only assets beside the example; use `examples/internal/`
   when several examples share them.
5. Preserve the dependency direction described in the architecture document
   and keep existing public import paths stable during file organization work.

### Directory Ownership

- `widget/<component>` owns a usable component. The public `widget/disclosure`,
  `widget/navigation`, and `widget/optionrow` packages provide reusable behavior
  and rendering primitives rather than complete controls. `widget/chart`
  includes shared public chart data and interaction contracts.
- `theme/<component>style` owns appearance recipes. Helpers used only by the
  theme tree belong in `theme/internal/`; shared behavior used across framework
  layers belongs in root `internal/`.
- `internal/testtheme` supplies neutral fixtures for repository tests only.
  Keep the entry points in `testtheme.go`, component fixtures in focused files,
  and shared field declaration helpers in `field.go`.
- `theme/base` owns the only default design system. Other styles and their
  dedicated examples belong in external modules, not in new core theme packages.
- `theme/themetest` is a public acceptance-test helper for default and external
  themes. It checks real Providers and has a different role from `testtheme`.
- `docs/assets` contains documentation images. New runnable-example assets
  belong under `examples/`, rather than extending the existing `assets/images`
  package.

### Internal API Design Principles

Internal packages must be robust against misuse. Design interfaces so incorrect
usage is **caught at compile time** or **fails immediately** rather than causing
silent corruption or order-dependent behavior.

**Type safety over conventions:**
- Encode invariants in types and signatures, not documentation
- If two values must stay synchronized, bundle them in one struct
- Use the type system to prevent invalid states

**Eliminate calling-order dependencies:**
- Methods should work regardless of call order unless the dependency is
  explicit in the signature
- If method A must be called before method B, consider:
  - Combining them into one operation
  - Making B take A's result as a parameter
  - Using a builder pattern with a terminal operation

**Query methods must not have side effects:**
- Methods that appear to be queries (return a value, named like `Get*`,
  `Is*`, `Visible`, `Contains`) should be pure reads
- Separate observation from state changes into distinct phases when needed
- Document any unavoidable side effects prominently

**Make cache keys complete:**
- If a cache depends on multiple inputs, include all of them in the key
- Missing inputs cause silent cache misses or stale data reuse
- Test caches with varying inputs, not just the happy path

**Examples of these principles in practice:**

- `KeyFilterCache` now includes the `target` tag in its cache key (#12), not
  just the filter names. Reusing the cache with a different target would route
  keys to the wrong widget.

- `ClickArea.Layout` now processes events the same way as `Clicked` (#13),
  eliminating the hidden dependency on calling `Clicked` before `Layout`.

- `Focus.Observe` records only frame-local focus input; persistent modality
  changes happen later in `CommitObservations` (#14).

When reviewing internal APIs, ask:
- Can this be misused by changing call order?
- Does this method name promise a query but perform writes?
- Are all cache dependencies reflected in the cache key?
- Will incorrect usage fail loudly or silently corrupt state?

## Checks

Run these before opening a pull request:

```bash
go fix ./...
gofmt -w path/to/changed.go
go test ./...
go vet ./...
git diff --check
```

Use Go 1.26.2 (or a newer compatible toolchain) for all changes. FlowUI is a
clean-break refactor: use the current immutable `style.Style`/`style.Scope` APIs
and do not reintroduce removed component-level theme or compatibility shims.

For visual or interaction changes, test the affected component with `uitest` and
include the platform used for verification. For behavior changes, add a focused
regression test when practical.

## Commit Messages

Use the format:

```text
type(scope): summary
```

Common types are `feat`, `fix`, `refactor`, `perf`, `docs`, `test`, and
`chore`. Keep the summary short and specific. Existing commits use scopes such
as `core`, `components`, `animation`, `render`, `focus`, `readme`, and
`license`.

Examples:

```text
fix(render): prevent duplicate shadow cache accounting
refactor(components): reuse shared animation state
docs(readme): update installation and examples
```

## Pull Requests

Describe the problem, the change, and the checks you ran. Call out public API
changes, visual changes, performance effects, and compatibility concerns. Keep
unrelated cleanup out of the same pull request.

Do not commit generated binaries, local configuration, or unrelated formatting
changes. Keep third-party code and its license notices intact.

## License

FlowUI is licensed under the [MIT License](LICENSE). Contributions must be
compatible with that license.
