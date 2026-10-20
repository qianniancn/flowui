# Components and shared primitives

Applications import components directly, for example
`github.com/qianniancn/flowui/widget/button`. There is no Go package at the
`widget` directory itself. The `ui` package supplies the MVU/runtime contract.

The main constructor in each component package is `New`, returning `Widget`.
Related controls keep descriptive names, such as `button.Group`,
`input.TextArea`, and `progress.Bar`.

Text controls accept strings directly:

```go
button.New("save", "Save").OnClick(save)
togglebutton.New("bold", selected, "Bold").OnChange(setBold)
tooltip.New("save-tip", saveButton, "Save changes")
input.GroupAction("clear", "Clear").OnClick(clear)
```

Use `Content` for icons, styled text, or composite widgets. For buttons, toggle
buttons, and input actions, the constructor text remains the accessible label;
`Label` overrides that label without changing visible content.

```go
button.New("save", "Save").Content(icon.New(lucide.Save)).IconOnly()
button.New("save-large", "Save").Content(text.New("Save").Size(20))
tooltip.New("details", trigger, "").Content(details)
```

Containers such as cards, toolbars, and modal bodies accept widgets so callers
can compose their contents. Standalone text still uses `text.New`.

This is a breaking API change: replace the former main constructors (for
example `button.Button`, `input.Input`, and `toast.Provider`) with `New`.
`toast.Toast` becomes `toast.NewItem`. Replace text children with strings,
move custom children to `Content`, and remove the third argument of
`input.GroupAction` (or move it to `Content` when custom).

Each component owns its behavior, interaction state, accessibility, and domain
layout. Visual contracts live in `theme`; shared appearance recipes live in
`theme/<component>style`. See the [theme package map](../theme/README.md).

Some public packages are building blocks for components rather than complete
controls:

| Package | Purpose |
| --- | --- |
| `disclosure` | Controlled and uncontrolled value bindings |
| `navigation` | Keyboard movement, disabled-item skipping, and typeahead |
| `optionrow` | Reusable option-row state, animation, and drawing helpers |
| `chart` | Shared chart data, annotations, scales, caching, and interaction helpers |

These packages retain their public import paths so custom components can reuse
them. In particular, types such as `chart.Selection` and `chart.DataWindow`
appear in chart components' public APIs.

Put helpers that are private to the widget tree under `widget/internal/` when
needed. Helpers shared across framework layers belong in root `internal/`.
Keep neutral fixtures for repository behavior tests in `internal/testtheme`,
separate from the production theme recipes.
