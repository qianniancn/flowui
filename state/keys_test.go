package state

import "testing"

func TestKeysClaimScopesKey(t *testing.T) {
	var keys Keys
	keys.BeginFrame()
	pop := keys.Push("todo:1")
	got := keys.Claim(KindClickable, "done")
	pop()

	if got != "todo:1/done" {
		t.Fatalf("claimed key = %q, want scoped key", got)
	}
	if keys.Frame()[got] != KindClickable {
		t.Fatalf("frame kind = %q, want clickable", keys.Frame()[got])
	}
}

func TestKeysRejectDuplicate(t *testing.T) {
	var keys Keys
	keys.BeginFrame()
	keys.Claim(KindClickable, "save")

	defer func() {
		if recover() == nil {
			t.Fatal("expected duplicate key panic")
		}
	}()
	keys.Claim(KindInput, "save")
}

func TestSweepRemovesUnusedState(t *testing.T) {
	states := map[string]*int{
		"keep": new(int),
		"drop": new(int),
	}
	frame := map[string]Kind{
		"keep": KindInput,
	}

	Sweep(states, frame, KindInput)

	if states["keep"] == nil {
		t.Fatal("kept state was removed")
	}
	if states["drop"] != nil {
		t.Fatal("unused state was kept")
	}
}
