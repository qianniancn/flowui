package style

import "testing"

func TestUnderlineIsHashedAndResolved(t *testing.T) {
	plain := New()
	underlined := New().Underline(true)

	if plain.Hash64() == underlined.Hash64() {
		t.Fatal("Underline(true) hashed identically to an unset declaration")
	}
	if New().Underline(false).Hash64() == underlined.Hash64() {
		t.Fatal("Underline(false) hashed identically to Underline(true)")
	}

	resolved := underlined.Resolve(State{})
	if resolved.Text == nil || resolved.Text.Underline == nil || !*resolved.Text.Underline {
		t.Fatal("Underline did not survive resolution")
	}
}

// A state row that turns the underline off must survive the merge, not just
// the hash and the clone.
func TestUnderlineComposesWithStateRules(t *testing.T) {
	declaration := New().
		Underline(true).
		On(Disabled, New().Underline(false))

	resolved := declaration.Resolve(State{Disabled: true})
	if resolved.Text == nil || resolved.Text.Underline == nil {
		t.Fatal("underline missing after state resolution")
	}
	if *resolved.Text.Underline {
		t.Fatal("state row did not turn the underline off")
	}
}

func TestLineStyleIsHashedAndResolved(t *testing.T) {
	base := New().BorderWidth(1).BorderColor(RGB(0x000000))
	dashed := base.BorderLine(LineDashed)

	if base.Hash64() == dashed.Hash64() {
		t.Fatal("LineDashed hashed identically to an unset line style")
	}
	if dashed.Hash64() == base.BorderLine(LineSolid).Hash64() {
		t.Fatal("LineDashed hashed identically to LineSolid")
	}

	resolved := dashed.Resolve(State{})
	if resolved.Paint == nil || resolved.Paint.Border == nil {
		t.Fatal("border did not survive resolution")
	}
	if resolved.Paint.Border.LineStyle == nil || *resolved.Paint.Border.LineStyle != LineDashed {
		t.Fatal("LineDashed did not survive resolution")
	}
}

func TestLineStyleDashPattern(t *testing.T) {
	if on, off := LineSolid.DashPattern(); on != 0 || off != 0 {
		t.Fatalf("LineSolid pattern = (%v, %v), want (0, 0)", on, off)
	}
	if on, off := LineDashed.DashPattern(); on <= 0 || off <= 0 {
		t.Fatalf("LineDashed pattern = (%v, %v), want positive lengths", on, off)
	}
	if on, off := LineDotted.DashPattern(); on <= 0 || off <= 0 {
		t.Fatalf("LineDotted pattern = (%v, %v), want positive lengths", on, off)
	}
}

// A state row that switches the line style must keep the rest of the border.
func TestLineStyleComposesWithStateRules(t *testing.T) {
	declaration := New().
		BorderWidth(1).
		BorderColor(RGB(0x111111)).
		On(Disabled, New().BorderLine(LineDashed))

	resolved := declaration.Resolve(State{Disabled: true})
	if resolved.Paint == nil || resolved.Paint.Border == nil {
		t.Fatal("border missing after state resolution")
	}
	if resolved.Paint.Border.Width == nil || *resolved.Paint.Border.Width != 1 {
		t.Fatal("state row dropped the base border width")
	}
	if resolved.Paint.Border.LineStyle == nil || *resolved.Paint.Border.LineStyle != LineDashed {
		t.Fatal("state row did not apply LineDashed")
	}
}
