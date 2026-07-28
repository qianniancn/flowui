package tooltip

import (
	"time"

	"gioui.org/layout"
	"gioui.org/op"
	"github.com/qianniancn/flowui/internal/frame"
)

const (
	tooltipCoordinatorKey  = "\x00tooltip-coordinator"
	tooltipCoordinatorSlot = "tooltip-coordinator"
	tooltipCooldown        = 500 * time.Millisecond
)

type tooltipCoordinator struct {
	warmed           bool
	activeKey        string
	activeCloseDelay time.Duration
	cooldownAt       time.Time
	seen             map[string]struct{}
	afterScheduled   bool
}

func tooltipCoordinatorFor(ctx *frame.Context) *tooltipCoordinator {
	return frame.UseState[tooltipCoordinator](ctx, tooltipCoordinatorKey, tooltipCoordinatorSlot)
}

func (c *tooltipCoordinator) BeginFrame() {
	clear(c.seen)
	c.afterScheduled = false
}

func (c *tooltipCoordinator) register(ctx *frame.Context, gtx layout.Context, key string) {
	if c.seen == nil {
		c.seen = make(map[string]struct{})
	}
	c.seen[key] = struct{}{}
	if c.afterScheduled {
		return
	}
	c.afterScheduled = true
	frame.AfterOverlays(ctx, func() {
		c.finishFrame(gtx)
	})
}

func (c *tooltipCoordinator) open(key string, closeDelay time.Duration) {
	c.warmed = true
	c.activeKey = key
	c.activeCloseDelay = closeDelay
	c.cooldownAt = time.Time{}
}

func (c *tooltipCoordinator) beginCooldown(gtx layout.Context, key string, closeDelay time.Duration) {
	if !c.warmed || c.activeKey != key {
		return
	}
	duration := max(tooltipCooldown, closeDelay)
	c.cooldownAt = gtx.Now.Add(duration)
	gtx.Execute(op.InvalidateCmd{At: c.cooldownAt})
}

func (c *tooltipCoordinator) update(gtx layout.Context) {
	if c.cooldownAt.IsZero() {
		return
	}
	if !gtx.Now.Before(c.cooldownAt) {
		c.warmed = false
		c.activeKey = ""
		c.activeCloseDelay = 0
		c.cooldownAt = time.Time{}
		return
	}
	gtx.Execute(op.InvalidateCmd{At: c.cooldownAt})
}

func (c *tooltipCoordinator) finishFrame(gtx layout.Context) {
	if !c.warmed || c.activeKey == "" || !c.cooldownAt.IsZero() {
		return
	}
	if _, ok := c.seen[c.activeKey]; ok {
		return
	}
	c.beginCooldown(gtx, c.activeKey, c.activeCloseDelay)
}
