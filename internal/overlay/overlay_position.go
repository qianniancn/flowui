package overlay

import (
	"image"

	"gioui.org/f32"
)

type PopoverPlacement int

const (
	PopoverBottom PopoverPlacement = iota
	PopoverTop
	PopoverLeft
	PopoverRight
	PopoverBottomStart
	PopoverBottomEnd
	PopoverTopStart
	PopoverTopEnd
	PopoverLeftStart
	PopoverLeftEnd
	PopoverRightStart
	PopoverRightEnd
)

type Side uint8

const (
	SideBottom Side = iota
	SideTop
	SideLeft
	SideRight
)

type Align uint8

const (
	AlignCenter Align = iota
	AlignStart
	AlignEnd
)

type Placement struct {
	Side  Side
	Align Align
}

func (p PopoverPlacement) Placement() Placement {
	result := Placement{Side: SideBottom, Align: AlignCenter}
	switch p {
	case PopoverTop, PopoverTopStart, PopoverTopEnd:
		result.Side = SideTop
	case PopoverLeft, PopoverLeftStart, PopoverLeftEnd:
		result.Side = SideLeft
	case PopoverRight, PopoverRightStart, PopoverRightEnd:
		result.Side = SideRight
	}
	switch p {
	case PopoverBottomStart, PopoverTopStart, PopoverLeftStart, PopoverRightStart:
		result.Align = AlignStart
	case PopoverBottomEnd, PopoverTopEnd, PopoverLeftEnd, PopoverRightEnd:
		result.Align = AlignEnd
	}
	return result
}

func (p Placement) PopoverPlacement() PopoverPlacement {
	switch p.Side {
	case SideTop:
		switch p.Align {
		case AlignStart:
			return PopoverTopStart
		case AlignEnd:
			return PopoverTopEnd
		default:
			return PopoverTop
		}
	case SideLeft:
		switch p.Align {
		case AlignStart:
			return PopoverLeftStart
		case AlignEnd:
			return PopoverLeftEnd
		default:
			return PopoverLeft
		}
	case SideRight:
		switch p.Align {
		case AlignStart:
			return PopoverRightStart
		case AlignEnd:
			return PopoverRightEnd
		default:
			return PopoverRight
		}
	default:
		switch p.Align {
		case AlignStart:
			return PopoverBottomStart
		case AlignEnd:
			return PopoverBottomEnd
		default:
			return PopoverBottom
		}
	}
}

type PositionConfig struct {
	Trigger          image.Point
	TriggerOrigin    image.Point
	HasTriggerOrigin bool
	Panel            image.Point
	Bounds           image.Point
	Offset           int
	Placement        Placement
	Flip             bool
	AvoidOverflow    bool
}

type PositionResult struct {
	Placement Placement
	Position  image.Point
	Rect      image.Rectangle
}

func ResolvePosition(cfg PositionConfig) PositionResult {
	placement := cfg.Placement
	if cfg.Flip {
		if cfg.HasTriggerOrigin {
			placement = ResolvePlacementAt(cfg.TriggerOrigin, cfg.Trigger, cfg.Panel, cfg.Bounds, cfg.Offset, placement)
		} else {
			placement = ResolvePlacement(cfg.Trigger, cfg.Panel, cfg.Bounds, cfg.Offset, placement)
		}
	}
	pos := RawPosition(cfg.Trigger, cfg.Panel, cfg.Offset, placement)
	if cfg.HasTriggerOrigin {
		pos = RawPositionAt(cfg.TriggerOrigin, cfg.Trigger, cfg.Panel, cfg.Offset, placement)
	}
	if cfg.AvoidOverflow {
		pos = AvoidOverflow(pos, cfg.Panel, cfg.Bounds)
	}
	return PositionResult{
		Placement: placement,
		Position:  pos,
		Rect:      image.Rectangle{Min: pos, Max: pos.Add(cfg.Panel)},
	}
}

func ResolvePlacement(trigger, panel, bounds image.Point, offset int, placement Placement) Placement {
	pos := RawPosition(trigger, panel, offset, placement)
	switch placement.Side {
	case SideBottom:
		if pos.Y+panel.Y > bounds.Y && trigger.Y >= panel.Y+offset {
			return Placement{Side: SideTop, Align: placement.Align}
		}
	case SideTop:
		if pos.Y < 0 && bounds.Y-trigger.Y >= panel.Y+offset {
			return Placement{Side: SideBottom, Align: placement.Align}
		}
	case SideLeft:
		if pos.X < 0 && bounds.X-trigger.X >= panel.X+offset {
			return Placement{Side: SideRight, Align: placement.Align}
		}
	case SideRight:
		if pos.X+panel.X > bounds.X && trigger.X >= panel.X+offset {
			return Placement{Side: SideLeft, Align: placement.Align}
		}
	}
	return placement
}

func ResolvePlacementAt(origin, trigger, panel, bounds image.Point, offset int, placement Placement) Placement {
	pos := RawPositionAt(origin, trigger, panel, offset, placement)
	switch placement.Side {
	case SideBottom:
		if pos.Y+panel.Y > bounds.Y && origin.Y >= panel.Y+offset {
			return Placement{Side: SideTop, Align: placement.Align}
		}
	case SideTop:
		if pos.Y < 0 && bounds.Y-origin.Y-trigger.Y >= panel.Y+offset {
			return Placement{Side: SideBottom, Align: placement.Align}
		}
	case SideLeft:
		if pos.X < 0 && bounds.X-origin.X-trigger.X >= panel.X+offset {
			return Placement{Side: SideRight, Align: placement.Align}
		}
	case SideRight:
		if pos.X+panel.X > bounds.X && origin.X >= panel.X+offset {
			return Placement{Side: SideLeft, Align: placement.Align}
		}
	}
	return placement
}

func RawPosition(trigger, panel image.Point, offset int, placement Placement) image.Point {
	return RawPositionAt(image.Point{}, trigger, panel, offset, placement)
}

func RawPositionAt(origin, trigger, panel image.Point, offset int, placement Placement) image.Point {
	switch placement.Side {
	case SideTop:
		return origin.Add(image.Pt(crossPosition(trigger.X, panel.X, placement.Align), -panel.Y-offset))
	case SideLeft:
		return origin.Add(image.Pt(-panel.X-offset, crossPosition(trigger.Y, panel.Y, placement.Align)))
	case SideRight:
		return origin.Add(image.Pt(trigger.X+offset, crossPosition(trigger.Y, panel.Y, placement.Align)))
	default:
		return origin.Add(image.Pt(crossPosition(trigger.X, panel.X, placement.Align), trigger.Y+offset))
	}
}

func crossPosition(trigger, panel int, align Align) int {
	switch align {
	case AlignStart:
		return 0
	case AlignEnd:
		return trigger - panel
	default:
		return (trigger - panel) / 2
	}
}

func AvoidOverflow(pos, panel, bounds image.Point) image.Point {
	// Positions are relative to the trigger's local origin, so negative values
	// are valid for top and left placements. Bounds constrain positive edges.
	if bounds.X > 0 && pos.X+panel.X > bounds.X {
		pos.X = bounds.X - panel.X
	}
	if bounds.Y > 0 && pos.Y+panel.Y > bounds.Y {
		pos.Y = bounds.Y - panel.Y
	}
	return pos
}

func TransformOrigin(rect image.Rectangle, placement Placement) f32.Point {
	switch placement.Side {
	case SideTop:
		return f32.Pt(float32(rect.Min.X+rect.Dx()/2), float32(rect.Max.Y))
	case SideLeft:
		return f32.Pt(float32(rect.Max.X), float32(rect.Min.Y+rect.Dy()/2))
	case SideRight:
		return f32.Pt(float32(rect.Min.X), float32(rect.Min.Y+rect.Dy()/2))
	default:
		return f32.Pt(float32(rect.Min.X+rect.Dx()/2), float32(rect.Min.Y))
	}
}

func PanelTransformOrigin(trigger, panelPos, panel image.Point, placement Placement) f32.Point {
	return PanelTransformOriginAt(image.Rectangle{Max: trigger}, panelPos, panel, placement)
}

func PanelTransformOriginAt(trigger image.Rectangle, panelPos, panel image.Point, placement Placement) f32.Point {
	triggerCenter := f32.Pt(
		float32(trigger.Min.X)+float32(trigger.Dx())/2,
		float32(trigger.Min.Y)+float32(trigger.Dy())/2,
	)
	switch placement.Side {
	case SideTop:
		return f32.Pt(triggerCenter.X-float32(panelPos.X), float32(panel.Y))
	case SideLeft:
		return f32.Pt(float32(panel.X), triggerCenter.Y-float32(panelPos.Y))
	case SideRight:
		return f32.Pt(0, triggerCenter.Y-float32(panelPos.Y))
	default:
		return f32.Pt(triggerCenter.X-float32(panelPos.X), 0)
	}
}

func SlideOffset(delta int, progress float32, placement Placement) image.Point {
	remaining := int(float32(delta)*(1-progress) + 0.5)
	switch placement.Side {
	case SideTop:
		return image.Pt(0, remaining)
	case SideLeft:
		return image.Pt(remaining, 0)
	case SideRight:
		return image.Pt(-remaining, 0)
	default:
		return image.Pt(0, -remaining)
	}
}
