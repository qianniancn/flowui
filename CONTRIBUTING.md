# Contributing to FlowUI

FlowUI is a Go desktop UI framework built on Gio. Keep changes focused on the
public `ui` API, component behavior, rendering, tests, and examples.

Before changing code, read:

- [`README.md`](README.md) for installation and usage.
- [`docs/architecture.md`](docs/architecture.md) for package boundaries,
  state ownership, and overlay behavior.

## Development Environment

- Go 1.26.2 or newer
- A desktop platform supported by Gio

The repository root intentionally contains no Go package. Run commands from the
repository root, or run an individual example with:

```bash
go run ./examples/counter
```

## Code Changes

- Applications import `github.com/qianniancn/FlowUI/ui`; implementation code
  belongs below `internal/`.
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
