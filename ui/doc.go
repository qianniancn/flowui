// Package ui is FlowUI's public MVU entry point.
//
// Applications keep business state in a Model, send typed messages from View,
// and mutate the model in Update. Context owns only Gio frame and interaction
// services needed by widgets. New code can use the responsibility-oriented
// packages alongside ui:
//
//   - widget contains the official components;
//   - layout contains layout primitives;
//   - style contains style declarations and themes;
//   - draw contains low-level drawing helpers;
//   - animation contains reusable motion primitives;
//   - window contains native window options.
//
// Import this package as:
//
//	import "github.com/qianniancn/flowui/ui"
package ui
