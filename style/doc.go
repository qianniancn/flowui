// Package style contains FlowUI's immutable style declarations: layout, paint,
// text, and transition properties, plus the conditional and variant
// declarations that compose them.
//
// It is a pure declaration mechanism and deliberately depends on nothing else
// in the framework. That is what lets theme name style (for recipe and
// callback signatures) without creating an import cycle. Entry points that
// need the active frame context or the active theme values live in
// github.com/qianniancn/flowui/style/resolve instead.
package style
