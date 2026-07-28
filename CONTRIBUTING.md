# Contributing to FlowUI

FlowUI is a Go desktop UI framework built on Gio. Keep changes focused on the
public `ui` API, optional platform services, component behavior, rendering,
tests, and examples.

Before changing code, read:

- [`README.md`](README.md) for installation and usage.
- [Project Wiki](https://github.com/qianniancn/flowui/wiki) — user-facing tutorial (final API).
- [`docs/architecture.md`](docs/architecture.md) — dependency direction, state
  ownership, and overlay behavior for the current tree.
- `examples/` and `go doc github.com/qianniancn/flowui/ui` for current usage.

## Development Environment

- Go 1.26.2 or newer
- A desktop platform supported by Gio

The repository root intentionally contains no Go package. Run commands from the
repository root, or run an individual example with:

```bash
go run ./examples/counter
```

## Code Changes

- Applications import `github.com/qianniancn/flowui/ui` for UI and MVU APIs;
  `explorer`, `notify`, and `systray` are optional platform services. Shared
  implementation code belongs below `internal/`.
- Keep business state in the application's model. Component state is for
  interaction and derived rendering state.
- Use explicit keys for retained component state. Reuse existing state,
  animation, layout, theme, and rendering helpers before adding new ones.
- Keep controlled components controlled: report user intent through callbacks
  or messages instead of mutating application data inside a widget.
- Avoid new dependencies and abstractions unless the existing code cannot meet
  the requirement.

When adding or changing a component:

1. Put the implementation and focused tests in `internal/components/<domain>`.
2. Expose the public API through a focused file under `ui/`.
3. Add or update a runnable example under `examples/`.
4. Preserve the dependency direction described in the architecture document.

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
clean-break refactor: use the current immutable `ui.Style`/`StyleScope`/`When` APIs
and do not reintroduce removed component-level theme or compatibility shims.

For visual or interaction changes, test the affected example and include the
platform used for verification. For behavior changes, add a focused regression
test when practical.

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
