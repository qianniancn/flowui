# Behavior test themes

This package supplies neutral visual fixtures for repository behavior, cache,
and Provider-contract tests. Import it from `_test.go` files only; applications
and production framework code must not depend on it.

The fixtures provide stable geometry and valid render hooks independently of
the shipped component recipes. A production theme's padding or renderer change
should not silently redefine the setup for a behavior test.

- `testtheme.go` owns `Behavior`, `BehaviorProvider`, the Provider declarations,
  and the shared neutral palette.
- Component files such as `button.go`, `input.go`, and `tree.go` own their
  corresponding visual fixtures and private helpers.
- `field.go` contains the declaration readers shared by field fixtures.
- `sidebar.go` also owns `WithSidebarProvider`, used by focused sidebar tests.

Tests can embed `BehaviorProvider()` and override the capability being tested.
Keep component-specific fixtures explicit rather than delegating to
`theme/base` or other production recipes. Shared text and icon rendering
primitives may be reused when needed by interaction tests.

For acceptance checks against real built-in or custom theme Providers, use the
public `theme/themetest` package instead.
