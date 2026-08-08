# FlowUI Documentation

FlowUI is a Go desktop UI framework and component library built on Gio. It
combines a typed MVU application model with reusable components, declarative
styles, window management, asynchronous effects, and deterministic UI tests.

[简体中文](https://qianniancn.github.io/flowui/)

![FlowUI component gallery](https://qianniancn.github.io/flowui/assets/components-gallery.png)

[Get started](guide/01-quick-start.md){ .md-button .md-button--primary }
[View the source](https://github.com/qianniancn/flowui){ .md-button }

## Reading path

| Order | Page | What it covers |
| --- | --- | --- |
| 1 | [Getting started](guide/01-quick-start.md) | Install and run your first window |
| 2 | [Core concepts](guide/02-core-concepts.md) | Model, messages, keys, and ownership |
| 3 | [MVU and messages](guide/03-mvu-and-messages.md) | Update, View, Cmd, and subscriptions |
| 4 | [Layout](guide/04-layout.md) | Containers, constraints, and grids |
| 5 | [Styling and themes](guide/05-styling-and-themes.md) | Tokens, parts, states, and fonts |
| 6 | [Components](guide/06-components.md) | Public component families |
| 7 | [Forms and controlled state](guide/07-forms-and-controlled-state.md) | Inputs, selection, and open state |
| 8 | [Commands and subscriptions](guide/08-commands-and-subscriptions.md) | Effects, cancellation, and errors |
| 9 | [Overlays and popups](guide/09-overlays-and-popups.md) | Menus, popovers, modals, and portals |
| 10 | [Custom components](guide/10-custom-components.md) | Composition and custom layout |
| 11 | [Multiple windows](guide/11-multiple-windows.md) | Application lifecycle and window state |
| 12 | [Animation](guide/12-animation.md) | Transitions, tweens, springs, and timelines |
| 13 | [Testing](guide/13-testing.md) | Harness and application tests |
| 14 | [FAQ](guide/14-faq.md) | Common questions and troubleshooting |
| 15 | [Gantt chart](guide/15-gantt.md) | Tasks, dependencies, and editing |
| 16 | [Component screenshots](guide/16-component-screenshots.md) | Visual reference for the examples |
| 17 | [Node graph](guide/17-node-graph.md) | Nodes, ports, edges, and a controlled viewport |

## Repository map

| Purpose | Location |
| --- | --- |
| Public package | `github.com/qianniancn/flowui/ui` |
| Runnable examples | `examples/` |
| Go API reference | [pkg.go.dev](https://pkg.go.dev/github.com/qianniancn/flowui/ui) |

The guide targets application developers and uses only the public `ui` package.
The `internal/` tree is implementation detail.

## Current versions

This documentation matches Go 1.26.2 and Gio v0.10.1. The single-window entry
point is `ui.Run(Program)`; use `ui.Application` when the process owns multiple
windows or a longer lifecycle.
