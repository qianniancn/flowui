package disclosure

import "testing"

// helpers

func boolCfg(controlled bool, value bool, hasDefault bool, defaultVal bool, onChange func(bool)) Config[bool] {
	return Config[bool]{
		Controlled: controlled,
		Value:      value,
		HasDefault: hasDefault,
		Default:    defaultVal,
		OnChange:   onChange,
	}
}

func strCfg(controlled bool, value string, hasDefault bool, defaultVal string, onChange func(string)) Config[string] {
	return Config[string]{
		Controlled: controlled,
		Value:      value,
		HasDefault: hasDefault,
		Default:    defaultVal,
		OnChange:   onChange,
	}
}

// Current

func TestCurrentUncontrolledZeroValue(t *testing.T) {
	var b Binding[bool]
	if got := b.Current(boolCfg(false, false, false, false, nil)); got != false {
		t.Fatalf("got %v, want false", got)
	}
}

func TestCurrentSeedsDefaultOnFirstCall(t *testing.T) {
	var b Binding[bool]
	cfg := boolCfg(false, false, true, true, nil)
	if got := b.Current(cfg); got != true {
		t.Fatalf("got %v, want true (seeded from default)", got)
	}
	// Second call with different default must NOT re-seed.
	cfg2 := boolCfg(false, false, true, false, nil)
	if got := b.Current(cfg2); got != true {
		t.Fatalf("got %v, want true (default not re-seeded)", got)
	}
}

func TestCurrentControlledReturnsConfigValue(t *testing.T) {
	var b Binding[bool]
	if got := b.Current(boolCfg(true, true, false, false, nil)); got != true {
		t.Fatalf("got %v, want true", got)
	}
	if got := b.Current(boolCfg(true, false, false, false, nil)); got != false {
		t.Fatalf("got %v, want false", got)
	}
}

func TestCurrentControlledIgnoresInternalValue(t *testing.T) {
	var b Binding[bool]
	// Seed internal value via uncontrolled.
	b.Current(boolCfg(false, false, true, true, nil))
	// Switch to controlled: must return Config.Value, not internal true.
	if got := b.Current(boolCfg(true, false, false, false, nil)); got != false {
		t.Fatalf("got %v, want false (controlled overrides internal)", got)
	}
}

func TestCurrentStringBinding(t *testing.T) {
	var b Binding[string]
	cfg := strCfg(false, "", true, "file", nil)
	if got := b.Current(cfg); got != "file" {
		t.Fatalf("got %q, want file", got)
	}
}

// Request

func TestRequestUncontrolledMutatesInternalValue(t *testing.T) {
	var b Binding[bool]
	b.Current(boolCfg(false, false, false, false, nil))
	eff, changed := b.Request(boolCfg(false, false, false, false, nil), true)
	if !eff || !changed {
		t.Fatalf("eff=%v changed=%v, want true/true", eff, changed)
	}
	// Idempotent: requesting same value must not fire change.
	_, changed2 := b.Request(boolCfg(false, false, false, false, nil), true)
	if changed2 {
		t.Fatal("requesting same value reported changed=true")
	}
}

func TestRequestUncontrolledFiresOnChange(t *testing.T) {
	var b Binding[bool]
	b.Current(boolCfg(false, false, false, false, nil))
	fired := false
	b.Request(boolCfg(false, false, false, false, func(v bool) { fired = v }), true)
	if !fired {
		t.Fatal("OnChange not called")
	}
}

func TestRequestControlledOnlyFiresOnChangeWhenDifferent(t *testing.T) {
	var b Binding[bool]
	calls := 0
	cfg := boolCfg(true, false, false, false, func(bool) { calls++ })
	// Requesting same value as Config.Value: no call.
	b.Request(cfg, false)
	if calls != 0 {
		t.Fatalf("calls=%d, want 0 (same value)", calls)
	}
	// Requesting different value: one call.
	b.Request(cfg, true)
	if calls != 1 {
		t.Fatalf("calls=%d, want 1", calls)
	}
}

func TestRequestControlledReturnsConfigValueNotNext(t *testing.T) {
	var b Binding[bool]
	// Controlled=true, Config.Value=false; requesting true fires OnChange but
	// effective value stays false until the model feeds it back.
	cfg := boolCfg(true, false, false, false, nil)
	eff, changed := b.Request(cfg, true)
	if eff != false {
		t.Fatalf("eff=%v, want false (model not updated yet)", eff)
	}
	if !changed {
		t.Fatal("changed should be true")
	}
}

// PeerClose

func TestPeerCloseUncontrolled(t *testing.T) {
	var b Binding[bool]
	b.Current(boolCfg(false, false, true, true, nil)) // seed true
	b.Bind(boolCfg(false, false, false, false, nil))
	if !b.PeerClose(false) {
		t.Fatal("PeerClose should return true when value changed")
	}
	if got := b.Current(boolCfg(false, false, false, false, nil)); got != false {
		t.Fatalf("value after PeerClose = %v, want false", got)
	}
	// Idempotent.
	if b.PeerClose(false) {
		t.Fatal("second PeerClose should return false (already closed)")
	}
}

func TestPeerCloseUncontrolledFiresOnChange(t *testing.T) {
	var b Binding[bool]
	b.Current(boolCfg(false, false, true, true, nil))
	fired := false
	b.Bind(boolCfg(false, false, false, false, func(v bool) { fired = !v }))
	b.PeerClose(false)
	if !fired {
		t.Fatal("PeerClose should fire OnChange")
	}
}

func TestPeerCloseControlledFiresOnChangeWhenNonZero(t *testing.T) {
	var b Binding[bool]
	calls := 0
	b.Bind(boolCfg(true, true, false, false, func(bool) { calls++ }))
	if !b.PeerClose(false) {
		t.Fatal("should return true when controlled and value != zero")
	}
	if calls != 1 {
		t.Fatalf("calls=%d, want 1", calls)
	}
}

func TestPeerCloseControlledNoCallWhenAlreadyZero(t *testing.T) {
	var b Binding[bool]
	calls := 0
	b.Bind(boolCfg(true, false, false, false, func(bool) { calls++ }))
	if b.PeerClose(false) {
		t.Fatal("should return false when controlled value already zero")
	}
	if calls != 0 {
		t.Fatalf("calls=%d, want 0", calls)
	}
}

func TestPeerCloseStringBinding(t *testing.T) {
	var b Binding[string]
	b.Current(strCfg(false, "", true, "file", nil))
	b.Bind(strCfg(false, "", false, "", nil))
	if !b.PeerClose("") {
		t.Fatal("PeerClose should return true")
	}
	if got := b.Current(strCfg(false, "", false, "", nil)); got != "" {
		t.Fatalf("value after PeerClose = %q, want empty", got)
	}
}
