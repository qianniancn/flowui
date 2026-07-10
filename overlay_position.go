package flowui

import (
	"image"

	"gioui.org/f32"
)

type overlaySide uint8

const (
	overlaySideBottom overlaySide = iota
	overlaySideTop
	overlaySideLeft
	overlaySideRight
)

type overlayAlign uint8

const (
	overlayAlignCenter overlayAlign = iota
	overlayAlignStart
	overlayAlignEnd
)

type overlayPlacement struct {
	side  overlaySide
	align overlayAlign
}

type overlayPositionConfig struct {
	Trigger          image.Point
	TriggerOrigin    image.Point
	HasTriggerOrigin bool
	Panel            image.Point
	Bounds           image.Point
	Offset           int
	Placement        overlayPlacement
	Flip             bool
	AvoidOverflow    bool
}

type overlayPositionResult struct {
	Placement overlayPlacement
	Position  image.Point
	Rect      image.Rectangle
}

func overlayResolvePosition(cfg overlayPositionConfig) overlayPositionResult {
	placement := cfg.Placement
	if cfg.Flip {
		if cfg.HasTriggerOrigin {
			placement = overlayResolvePlacementAt(cfg.TriggerOrigin, cfg.Trigger, cfg.Panel, cfg.Bounds, cfg.Offset, placement)
		} else {
			placement = overlayResolvePlacement(cfg.Trigger, cfg.Panel, cfg.Bounds, cfg.Offset, placement)
		}
	}
	pos := overlayRawPosition(cfg.Trigger, cfg.Panel, cfg.Offset, placement)
	if cfg.HasTriggerOrigin {
		pos = overlayRawPositionAt(cfg.TriggerOrigin, cfg.Trigger, cfg.Panel, cfg.Offset, placement)
	}
	if cfg.AvoidOverflow {
		pos = overlayAvoidOverflow(pos, cfg.Panel, cfg.Bounds)
	}
	return overlayPositionResult{
		Placement: placement,
		Position:  pos,
		Rect:      image.Rectangle{Min: pos, Max: pos.Add(cfg.Panel)},
	}
}

func overlayResolvePlacement(trigger, panel, bounds image.Point, offset int, placement overlayPlacement) overlayPlacement {
	pos := overlayRawPosition(trigger, panel, offset, placement)
	switch placement.side {
	case overlaySideBottom:
		if pos.Y+panel.Y > bounds.Y && trigger.Y >= panel.Y+offset {
			return overlayPlacement{side: overlaySideTop, align: placement.align}
		}
	case overlaySideTop:
		if pos.Y < 0 && bounds.Y-trigger.Y >= panel.Y+offset {
			return overlayPlacement{side: overlaySideBottom, align: placement.align}
		}
	case overlaySideLeft:
		if pos.X < 0 && bounds.X-trigger.X >= panel.X+offset {
			return overlayPlacement{side: overlaySideRight, align: placement.align}
		}
	case overlaySideRight:
		if pos.X+panel.X > bounds.X && trigger.X >= panel.X+offset {
			return overlayPlacement{side: overlaySideLeft, align: placement.align}
		}
	}
	return placement
}

func overlayResolvePlacementAt(origin, trigger, panel, bounds image.Point, offset int, placement overlayPlacement) overlayPlacement {
	pos := overlayRawPositionAt(origin, trigger, panel, offset, placement)
	switch placement.side {
	case overlaySideBottom:
		if pos.Y+panel.Y > bounds.Y && origin.Y >= panel.Y+offset {
			return overlayPlacement{side: overlaySideTop, align: placement.align}
		}
	case overlaySideTop:
		if pos.Y < 0 && bounds.Y-origin.Y-trigger.Y >= panel.Y+offset {
			return overlayPlacement{side: overlaySideBottom, align: placement.align}
		}
	case overlaySideLeft:
		if pos.X < 0 && bounds.X-origin.X-trigger.X >= panel.X+offset {
			return overlayPlacement{side: overlaySideRight, align: placement.align}
		}
	case overlaySideRight:
		if pos.X+panel.X > bounds.X && origin.X >= panel.X+offset {
			return overlayPlacement{side: overlaySideLeft, align: placement.align}
		}
	}
	return placement
}

func overlayRawPosition(trigger, panel image.Point, offset int, placement overlayPlacement) image.Point {
	return overlayRawPositionAt(image.Point{}, trigger, panel, offset, placement)
}

func overlayRawPositionAt(origin, trigger, panel image.Point, offset int, placement overlayPlacement) image.Point {
	switch placement.side {
	case overlaySideTop:
		return origin.Add(image.Pt(overlayCrossPosition(trigger.X, panel.X, placement.align), -panel.Y-offset))
	case overlaySideLeft:
		return origin.Add(image.Pt(-panel.X-offset, overlayCrossPosition(trigger.Y, panel.Y, placement.align)))
	case overlaySideRight:
		return origin.Add(image.Pt(trigger.X+offset, overlayCrossPosition(trigger.Y, panel.Y, placement.align)))
	default:
		return origin.Add(image.Pt(overlayCrossPosition(trigger.X, panel.X, placement.align), trigger.Y+offset))
	}
}

func overlayCrossPosition(trigger, panel int, align overlayAlign) int {
	switch align {
	case overlayAlignStart:
		return 0
	case overlayAlignEnd:
		return trigger - panel
	default:
		return (trigger - panel) / 2
	}
}

func overlayAvoidOverflow(pos, panel, bounds image.Point) image.Point {
	if bounds.X > 0 && pos.X+panel.X > bounds.X {
		pos.X = bounds.X - panel.X
	}
	if bounds.Y > 0 && pos.Y+panel.Y > bounds.Y {
		pos.Y = bounds.Y - panel.Y
	}
	return pos
}

func overlayTransformOrigin(rect image.Rectangle, placement overlayPlacement) f32.Point {
	switch placement.side {
	case overlaySideTop:
		return f32.Pt(float32(rect.Min.X+rect.Dx()/2), float32(rect.Max.Y))
	case overlaySideLeft:
		return f32.Pt(float32(rect.Max.X), float32(rect.Min.Y+rect.Dy()/2))
	case overlaySideRight:
		return f32.Pt(float32(rect.Min.X), float32(rect.Min.Y+rect.Dy()/2))
	default:
		return f32.Pt(float32(rect.Min.X+rect.Dx()/2), float32(rect.Min.Y))
	}
}

func overlayPanelTransformOrigin(trigger, panelPos, panel image.Point, placement overlayPlacement) f32.Point {
	return overlayPanelTransformOriginAt(image.Rectangle{Max: trigger}, panelPos, panel, placement)
}

func overlayPanelTransformOriginAt(trigger image.Rectangle, panelPos, panel image.Point, placement overlayPlacement) f32.Point {
	triggerCenter := f32.Pt(
		float32(trigger.Min.X)+float32(trigger.Dx())/2,
		float32(trigger.Min.Y)+float32(trigger.Dy())/2,
	)
	switch placement.side {
	case overlaySideTop:
		return f32.Pt(triggerCenter.X-float32(panelPos.X), float32(panel.Y))
	case overlaySideLeft:
		return f32.Pt(float32(panel.X), triggerCenter.Y-float32(panelPos.Y))
	case overlaySideRight:
		return f32.Pt(0, triggerCenter.Y-float32(panelPos.Y))
	default:
		return f32.Pt(triggerCenter.X-float32(panelPos.X), 0)
	}
}

func overlaySlideOffset(delta int, progress float32, placement overlayPlacement) image.Point {
	remaining := int(float32(delta)*(1-progress) + 0.5)
	switch placement.side {
	case overlaySideTop:
		return image.Pt(0, remaining)
	case overlaySideLeft:
		return image.Pt(remaining, 0)
	case overlaySideRight:
		return image.Pt(-remaining, 0)
	default:
		return image.Pt(0, -remaining)
	}
}
