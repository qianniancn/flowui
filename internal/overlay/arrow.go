package overlay

import (
	"image"
	"image/color"

	"gioui.org/f32"
	"gioui.org/layout"
	"gioui.org/op/clip"
	"gioui.org/op/paint"
)

// ArrowAnchor returns the cross-axis position, in panel coordinates, where an
// anchored popup arrow should point. It keeps the arrow away from rounded
// corners and clamps it when the panel is too small.
func ArrowAnchor(trigger image.Rectangle, panelPos, panelSize image.Point, placement PopoverPlacement, radius, arrowSize int) float32 {
	var anchor float32
	var crossSize int
	if placement.Placement().Side == SideTop || placement.Placement().Side == SideBottom {
		anchor = float32(trigger.Min.X + trigger.Dx()/2 - panelPos.X)
		crossSize = panelSize.X
	} else {
		anchor = float32(trigger.Min.Y + trigger.Dy()/2 - panelPos.Y)
		crossSize = panelSize.Y
	}
	halfArrow := float32(max(arrowSize, 0)) / 2
	margin := float32(max(radius, 0)) + halfArrow
	if float32(crossSize) < margin*2 {
		return float32(crossSize) / 2
	}
	return min(max(anchor, margin), float32(crossSize)-margin)
}

// ArrowRect returns a conservative local bounds rectangle for an arrow drawn
// around a panel. It is used to keep dismissal blockers away from the arrow.
func ArrowRect(panel image.Point, placement PopoverPlacement, anchor float32, size int) image.Rectangle {
	if size <= 0 || panel.X <= 0 || panel.Y <= 0 {
		return image.Rectangle{}
	}
	span := float32(size)
	depth := int(span*2/3 + 0.5)
	half := int(span/2 + 0.5)
	center := int(anchor + 0.5)
	switch placement.Placement().Side {
	case SideTop:
		return image.Rect(center-half, panel.Y, center+half+1, panel.Y+depth+1)
	case SideBottom:
		return image.Rect(center-half, -depth, center+half+1, 1)
	case SideLeft:
		return image.Rect(panel.X, center-half, panel.X+depth+1, center+half+1)
	case SideRight:
		return image.Rect(-depth, center-half, 1, center+half+1)
	default:
		return image.Rectangle{}
	}
}

// DrawArrow draws a rounded popup arrow using the same surface and border as
// its panel. anchor is measured along the panel's cross axis.
func DrawArrow(gtx layout.Context, placement PopoverPlacement, panel image.Point, anchor float32, size, borderWidth int, surface, border color.NRGBA) {
	if size <= 0 || panel.X <= 0 || panel.Y <= 0 {
		return
	}
	span := float32(size)
	depth := span * 2 / 3
	half := span / 2
	centerX := anchor
	centerY := anchor
	inset := min(float32(max(borderWidth, 0)), depth)
	var from, control0, control1, to f32.Point
	switch placement.Placement().Side {
	case SideTop:
		y := float32(panel.Y) - inset
		from = f32.Pt(centerX-half, y)
		control0 = f32.Pt(centerX-half+span*0.457069, y+depth)
		control1 = f32.Pt(centerX+span*0.041667, y+depth)
		to = f32.Pt(centerX+half, y)
	case SideBottom:
		y := inset
		from = f32.Pt(centerX-half, y)
		control0 = f32.Pt(centerX-half+span*0.457069, y-depth)
		control1 = f32.Pt(centerX+span*0.041667, y-depth)
		to = f32.Pt(centerX+half, y)
	case SideLeft:
		x := float32(panel.X) - inset
		from = f32.Pt(x, centerY-half)
		control0 = f32.Pt(x+depth, centerY-half+span*0.457069)
		control1 = f32.Pt(x+depth, centerY+span*0.041667)
		to = f32.Pt(x, centerY+half)
	case SideRight:
		x := inset
		from = f32.Pt(x, centerY-half)
		control0 = f32.Pt(x-depth, centerY-half+span*0.457069)
		control1 = f32.Pt(x-depth, centerY+span*0.041667)
		to = f32.Pt(x, centerY+half)
	}

	var fillPath clip.Path
	fillPath.Begin(gtx.Ops)
	fillPath.MoveTo(from)
	fillPath.CubeTo(control0, control1, to)
	fillPath.Close()
	paint.FillShape(gtx.Ops, surface, clip.Outline{Path: fillPath.End()}.Op())
	if borderWidth <= 0 || border.A == 0 {
		return
	}
	var curve clip.Path
	curve.Begin(gtx.Ops)
	curve.MoveTo(from)
	curve.CubeTo(control0, control1, to)
	stroke := clip.Stroke{Path: curve.End(), Width: float32(borderWidth)}.Op().Push(gtx.Ops)
	paint.Fill(gtx.Ops, border)
	stroke.Pop()
}
