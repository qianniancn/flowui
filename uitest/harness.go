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
	"github.com/qianniancn/flowui/i18n"
	"github.com/qianniancn/flowui/internal/core"
	internalocale "github.com/qianniancn/flowui/internal/locale"
	"github.com/qianniancn/flowui/theme"
	"github.com/qianniancn/flowui/ui"
	"github.com/qianniancn/flowui/window"
)

// Config controls a Harness. Size must be positive and Theme must provide a
// complete component visual provider. LanguageAuto selects English, a zero
// Metric uses one physical pixel per dp and sp, and a zero Now starts from a
// deterministic non-zero time.
type Config struct {
	Size     image.Point
	Theme    *theme.Theme
	Language window.Language
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

// New creates a one-to-one pixel harness using an explicit theme and English.
func New(size image.Point, activeTheme theme.Theme) *Harness {
	return NewWithConfig(Config{Size: size, Theme: &activeTheme, Language: window.LanguageEnglish})
}

// NewWithConfig creates a harness with explicit test configuration.
func NewWithConfig(config Config) *Harness {
	validateSize(config.Size)
	if config.Theme == nil {
		panic(theme.ErrMissingVisualProvider)
	}
	if err := theme.ValidateVisualProvider(*config.Theme); err != nil {
		panic(err)
	}
	activeTheme := *config.Theme
	theme.DetachMaterial(&activeTheme)
	theme.SyncMaterialTheme(&activeTheme)
	now := config.Now
	if now.IsZero() {
		now = time.Unix(1, 0)
	}
	language := config.Language
	if language == window.LanguageAuto {
		language = window.LanguageEnglish
	}
	ctx, err := core.NewWithCatalogStrict(nil, &activeTheme, internalocale.Language(language), i18n.Catalog{})
	if err != nil {
		panic(err)
	}
	return &Harness{
		ctx:    (*ui.Context)(ctx),
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
	frameCtx := (*core.Context)(h.ctx)
	core.BeginFrameWithViewport(frameCtx, h.size)
	dimensions := layout.Dimensions{}
	if root != nil {
		dimensions = root.Layout(h.ctx, gtx)
	}
	core.LayoutOverlays(frameCtx, gtx)
	core.ApplyFrameCommands(frameCtx, gtx)
	core.EndFrame(frameCtx)
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

// touchPointerID keeps touch input on an identity distinct from the id Click
// uses, so a test that mixes mouse and touch cannot continue one gesture with
// the other by accident.
const touchPointerID pointer.ID = 2

// TouchDown queues a touch press. Pair it with TouchMove and TouchUp to build a
// gesture, calling Frame between events so the component observes each step.
func (h *Harness) TouchDown(position f32.Point) {
	h.Touch(touchPointerID, pointer.Press, position)
}

// TouchMove queues a touch move.
func (h *Harness) TouchMove(position f32.Point) {
	h.Touch(touchPointerID, pointer.Move, position)
}

// TouchUp queues a touch release.
func (h *Harness) TouchUp(position f32.Point) {
	h.Touch(touchPointerID, pointer.Release, position)
}

// TouchTap queues a touch press and release for the next frame.
func (h *Harness) TouchTap(position f32.Point) {
	h.Touch(touchPointerID, pointer.Press, position)
	h.Touch(touchPointerID, pointer.Release, position)
}

// Touch queues one raw touch pointer event. Pass an explicit id to drive
// concurrent pointers; single-gesture tests should prefer TouchDown, TouchMove,
// TouchUp, and TouchTap.
func (h *Harness) Touch(id pointer.ID, kind pointer.Kind, position f32.Point) {
	h.router.Queue(pointer.Event{
		Kind:      kind,
		Source:    pointer.Touch,
		PointerID: id,
		Position:  position,
	})
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
