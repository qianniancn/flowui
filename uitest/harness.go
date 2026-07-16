// Package uitest provides a deterministic frame harness for testing FlowUI widgets.
package uitest

import (
	"image"
	"time"

	"gioui.org/f32"
	"gioui.org/io/input"
	"gioui.org/io/key"
	"gioui.org/io/pointer"
	"gioui.org/layout"
	"gioui.org/op"
	"gioui.org/unit"
	"github.com/qianniancn/FlowUI/internal/frame"
	internaltheme "github.com/qianniancn/FlowUI/internal/theme"
	"github.com/qianniancn/FlowUI/ui"
)

// Config controls a Harness. Size must be positive. A nil Theme uses the
// default theme, LanguageAuto selects English, a zero Metric uses one physical
// pixel per dp and sp, and a zero Now starts from a deterministic non-zero time.
type Config struct {
	Size     image.Point
	Theme    *ui.Theme
	Language ui.Language
	Metric   unit.Metric
	Now      time.Time
}

// Harness owns the FlowUI context, Gio input router, viewport, and clock used
// by a component test. It is not safe for concurrent use.
type Harness struct {
	ctx    *ui.Context
	router input.Router
	size   image.Point
	metric unit.Metric
	now    time.Time
}

// New creates a one-to-one pixel harness using the default theme and English.
func New(size image.Point) *Harness {
	return NewWithConfig(Config{Size: size, Language: ui.LanguageEnglish})
}

// NewWithConfig creates a harness with explicit test configuration.
func NewWithConfig(config Config) *Harness {
	validateSize(config.Size)
	activeTheme := ui.DefaultTheme()
	if config.Theme != nil {
		activeTheme = *config.Theme
	}
	if activeTheme.Material != nil {
		materialTheme := *activeTheme.Material
		activeTheme.Material = &materialTheme
	}
	internaltheme.SyncMaterialTheme(&activeTheme)
	now := config.Now
	if now.IsZero() {
		now = time.Unix(1, 0)
	}
	language := config.Language
	if language == ui.LanguageAuto {
		language = ui.LanguageEnglish
	}
	return &Harness{
		ctx:    frame.New(nil, &activeTheme, language),
		size:   config.Size,
		metric: config.Metric,
		now:    now,
	}
}

// Context returns the FlowUI context owned by the harness.
func (h *Harness) Context() *ui.Context {
	return h.ctx
}

// Router returns the Gio input router owned by the harness.
func (h *Harness) Router() *input.Router {
	return &h.router
}

// Frame lays out one complete FlowUI frame with exact viewport constraints,
// including root overlays, focus commands, state cleanup, and input routing.
func (h *Harness) Frame(root ui.Widget) layout.Dimensions {
	var ops op.Ops
	gtx := layout.Context{
		Constraints: layout.Exact(h.size),
		Metric:      h.metric,
		Now:         h.now,
		Source:      h.router.Source(),
		Ops:         &ops,
	}
	frame.BeginFrameWithViewport(h.ctx, h.size)
	dimensions := layout.Dimensions{}
	if root != nil {
		dimensions = root.Layout(h.ctx, gtx)
	}
	frame.LayoutOverlays(h.ctx, gtx)
	frame.ApplyFrameCommands(h.ctx, gtx)
	frame.EndFrame(h.ctx)
	h.router.Frame(&ops)
	return dimensions
}

// Click queues a primary mouse press and release for the next frame.
func (h *Harness) Click(position f32.Point) {
	h.router.Queue(
		pointer.Event{
			Kind:      pointer.Press,
			Source:    pointer.Mouse,
			PointerID: 1,
			Buttons:   pointer.ButtonPrimary,
			Position:  position,
		},
		pointer.Event{
			Kind:      pointer.Release,
			Source:    pointer.Mouse,
			PointerID: 1,
			Position:  position,
		},
	)
}

// Key queues one keyboard press and release for the next frame.
func (h *Harness) Key(name key.Name, modifiers key.Modifiers) {
	h.router.Queue(
		key.Event{Name: name, Modifiers: modifiers, State: key.Press},
		key.Event{Name: name, Modifiers: modifiers, State: key.Release},
	)
}

// Advance moves the deterministic frame clock forward.
func (h *Harness) Advance(duration time.Duration) {
	if duration < 0 {
		panic("flowui/uitest: negative clock advance")
	}
	h.now = h.now.Add(duration)
}

// Resize changes the viewport used by subsequent frames.
func (h *Harness) Resize(size image.Point) {
	validateSize(size)
	h.size = size
}

func validateSize(size image.Point) {
	if size.X <= 0 || size.Y <= 0 {
		panic("flowui/uitest: size must be positive")
	}
}
