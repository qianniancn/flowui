package state

import "testing"

func TestExclusiveClosesPreviousMember(t *testing.T) {
	var exclusive Exclusive
	closed := ""
	exclusive.BeginFrame()
	exclusive.Register("select", "first", func() { closed = "first" })
	exclusive.Register("select", "second", func() { closed = "second" })
	exclusive.Activate("select", "first")
	exclusive.Activate("select", "second")
	if closed != "first" {
		t.Fatalf("closed member = %q, want first", closed)
	}
	if got := exclusive.Active("select"); got != "second" {
		t.Fatalf("active member = %q, want second", got)
	}
}

func TestExclusiveDropsUnmountedMember(t *testing.T) {
	var exclusive Exclusive
	exclusive.BeginFrame()
	exclusive.Register("select", "language", func() {})
	exclusive.Activate("select", "language")
	exclusive.EndFrame()

	exclusive.BeginFrame()
	exclusive.EndFrame()
	if got := exclusive.Active("select"); got != "" {
		t.Fatalf("active member = %q, want empty", got)
	}
}
