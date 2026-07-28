package nav

import (
	"testing"
	"time"

	"gioui.org/io/key"
)

func list(labels []string, disabled ...int) List {
	dset := make(map[int]bool, len(disabled))
	for _, d := range disabled {
		dset[d] = true
	}
	return List{
		Count:    len(labels),
		Disabled: func(i int) bool { return dset[i] },
		Label:    func(i int) string { return labels[i] },
	}
}

func TestMoveNoWrapStopsAtEdges(t *testing.T) {
	l := list([]string{"a", "b", "c"})
	if got, ok := Move(l, 0, 1, false); !ok || got != 1 {
		t.Fatalf("down from 0 = (%d,%v), want (1,true)", got, ok)
	}
	if got, ok := Move(l, 2, 1, false); ok || got != 2 {
		t.Fatalf("down from last = (%d,%v), want (2,false)", got, ok)
	}
	if got, ok := Move(l, 2, -1, false); !ok || got != 1 {
		t.Fatalf("up from 2 = (%d,%v), want (1,true)", got, ok)
	}
}

func TestMoveSkipsDisabled(t *testing.T) {
	l := list([]string{"a", "b", "c"}, 1)
	if got, ok := Move(l, 0, 1, false); !ok || got != 2 {
		t.Fatalf("down skipping disabled = (%d,%v), want (2,true)", got, ok)
	}
}

func TestMoveOutOfRangeSeedsFromEdge(t *testing.T) {
	l := list([]string{"a", "b", "c"}, 0)
	if got, ok := Move(l, -1, 1, false); !ok || got != 1 {
		t.Fatalf("down from -1 = (%d,%v), want (1,true)", got, ok)
	}
	if got, ok := Move(l, -1, -1, false); !ok || got != 2 {
		t.Fatalf("up from -1 = (%d,%v), want (2,true)", got, ok)
	}
}

func TestMoveRecoversWhenCurrentDisabled(t *testing.T) {
	l := list([]string{"a", "b", "c"}, 1)
	if got, ok := Move(l, 1, 1, false); !ok || got != 2 {
		t.Fatalf("recover forward = (%d,%v), want (2,true)", got, ok)
	}
	last := list([]string{"a", "b", "c"}, 2)
	if got, ok := Move(last, 2, 1, false); !ok || got != 1 {
		t.Fatalf("recover reverse = (%d,%v), want (1,true)", got, ok)
	}
}

func TestMoveWrapsAround(t *testing.T) {
	l := list([]string{"a", "b", "c"})
	if got, ok := Move(l, 2, 1, true); !ok || got != 0 {
		t.Fatalf("wrap down from last = (%d,%v), want (0,true)", got, ok)
	}
	if got, ok := Move(l, 0, -1, true); !ok || got != 2 {
		t.Fatalf("wrap up from first = (%d,%v), want (2,true)", got, ok)
	}
	skip := list([]string{"a", "b", "c"}, 0)
	if got, ok := Move(skip, 2, 1, true); !ok || got != 1 {
		t.Fatalf("wrap skipping disabled = (%d,%v), want (1,true)", got, ok)
	}
	only := list([]string{"a", "b", "c"}, 0, 2)
	if got, ok := Move(only, 1, 1, true); ok || got != 1 {
		t.Fatalf("wrap with no other enabled = (%d,%v), want (1,false)", got, ok)
	}
	// Disabled current under wrap follows the forward wrap scan (not the
	// bidirectional NearestEnabled used by no-wrap), matching radio-group.
	fromDisabled := list([]string{"a", "b", "c"}, 1)
	if got, ok := Move(fromDisabled, 1, 1, true); !ok || got != 2 {
		t.Fatalf("wrap from disabled current = (%d,%v), want (2,true)", got, ok)
	}
	wrapPast := list([]string{"a", "b", "c"}, 2)
	if got, ok := Move(wrapPast, 2, 1, true); !ok || got != 0 {
		t.Fatalf("wrap from disabled last = (%d,%v), want (0,true)", got, ok)
	}
}

func TestFirstLastEnabled(t *testing.T) {
	l := list([]string{"a", "b", "c"}, 0)
	if got, ok := First(l); !ok || got != 1 {
		t.Fatalf("First = (%d,%v), want (1,true)", got, ok)
	}
	if got, ok := Last(l); !ok || got != 2 {
		t.Fatalf("Last = (%d,%v), want (2,true)", got, ok)
	}
	none := list([]string{"a", "b"}, 0, 1)
	if _, ok := First(none); ok {
		t.Fatal("First should fail when all disabled")
	}
	if _, ok := Last(none); ok {
		t.Fatal("Last should fail when all disabled")
	}
}

func TestMatchPrefixWithWraparound(t *testing.T) {
	l := list([]string{"alpha", "beta", "bravo"})
	if got, ok := Match(l, 0, "b"); !ok || got != 1 {
		t.Fatalf("match b from 0 = (%d,%v), want (1,true)", got, ok)
	}
	if got, ok := Match(l, 1, "a"); !ok || got != 0 {
		t.Fatalf("match a wrapping = (%d,%v), want (0,true)", got, ok)
	}
	if got, ok := Match(l, -1, "a"); !ok || got != 0 {
		t.Fatalf("match a from -1 = (%d,%v), want (0,true)", got, ok)
	}
	skip := list([]string{"alpha", "beta", "bravo"}, 1)
	if got, ok := Match(skip, 0, "b"); !ok || got != 2 {
		t.Fatalf("match skipping disabled = (%d,%v), want (2,true)", got, ok)
	}
	if got, ok := Match(l, 0, ""); ok || got != 0 {
		t.Fatalf("empty query = (%d,%v), want (0,false)", got, ok)
	}
	noLabel := List{Count: 3, Disabled: func(int) bool { return false }}
	if _, ok := Match(noLabel, 0, "a"); ok {
		t.Fatal("Match should fail when Label is nil")
	}
}

func TestPrintable(t *testing.T) {
	if got := Printable("a"); got != "a" {
		t.Fatalf("Printable(a) = %q, want a", got)
	}
	if got := Printable("A"); got != "a" {
		t.Fatalf("Printable(A) = %q, want a (lowered)", got)
	}
	if got := Printable(key.NameF1); got != "" {
		t.Fatalf("Printable(F1) = %q, want empty (multi-rune)", got)
	}
	if got := Printable(key.NameTab); got != "" {
		t.Fatalf("Printable(Tab) = %q, want empty (control)", got)
	}
}

func TestTypeaheadAppendAndReset(t *testing.T) {
	base := time.Unix(0, 0)
	var ta Typeahead
	if got := ta.Append(base, "n"); got != "n" {
		t.Fatalf("append n = %q, want n", got)
	}
	if got := ta.Append(base.Add(100*time.Millisecond), "e"); got != "ne" {
		t.Fatalf("append e = %q, want ne", got)
	}
	if got := ta.Append(base.Add(100*time.Millisecond+DefaultTypeaheadTimeout+time.Millisecond), "g"); got != "g" {
		t.Fatalf("append after timeout = %q, want g (reset)", got)
	}
	ta.Set("xy")
	if got := ta.Append(base.Add(2*time.Second), "z"); got != "z" {
		t.Fatalf("append after long idle = %q, want z", got)
	}
	ta.Set("ab")
	ta.Reset()
	if got := ta.Append(base.Add(3*time.Second), "c"); got != "c" {
		t.Fatalf("append after reset = %q, want c", got)
	}
}
