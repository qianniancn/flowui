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

func TestKeysFullKeyPreservesRootKey(t *testing.T) {
	var keys Keys
	const key = "root-key"

	if got := keys.FullKey(key); got != key {
		t.Fatalf("full key = %q, want unchanged root key %q", got, key)
	}
}

func TestKeysFullKeyEncodesAmbiguousRootKey(t *testing.T) {
	tests := []struct {
		key  string
		want string
	}{
		{key: "root/with%separators", want: "~rroot%2Fwith%25separators"},
		{key: "~rreserved", want: "~r~rreserved"},
		{key: "\x00reserved", want: "~r%00reserved"},
	}

	var keys Keys
	for _, tt := range tests {
		if got := keys.FullKey(tt.key); got != tt.want {
			t.Errorf("FullKey(%q) = %q, want %q", tt.key, got, tt.want)
		}
	}
}

func TestKeysDerivedClaimsCannotCollideWithUserKeys(t *testing.T) {
	var keys Keys
	keys.BeginFrame()

	derived := keys.ClaimDerived(KindClickable, "email", "label")
	user := keys.Claim(KindClickable, "\x00email/label")
	if derived == user {
		t.Fatalf("derived key collided with user key %q", derived)
	}
	if derived[0] != 0 {
		t.Fatalf("derived key = %q, want internal namespace", derived)
	}
}

func TestKeysDerivedIdentityEncodesOwnerAndRoleSeparately(t *testing.T) {
	var keys Keys
	ownerSlash := keys.Derived("a/b", "c")
	roleSlash := keys.Derived("a", "b/c")
	ownerNUL := keys.Derived("a\x00b", "c")
	roleNUL := keys.Derived("a", "b\x00c")

	identities := []string{ownerSlash, roleSlash, ownerNUL, roleNUL}
	for i := range identities {
		for j := i + 1; j < len(identities); j++ {
			if identities[i] == identities[j] {
				t.Fatalf("distinct owner/role pairs produced derived key %q", identities[i])
			}
		}
	}
}

func TestKeysDerivedClaimsIncludeOwnerScope(t *testing.T) {
	var keys Keys
	keys.BeginFrame()

	popFirst := keys.Push("first")
	first := keys.ClaimDerived(KindClickable, "email", "label")
	popFirst()

	popSecond := keys.Push("second")
	second := keys.ClaimDerived(KindClickable, "email", "label")
	popSecond()

	if first == second {
		t.Fatalf("different owner scopes produced derived key %q", first)
	}
}

func TestKeysClaimDerivedResolvedMatchesScopedOwner(t *testing.T) {
	var scoped Keys
	scoped.BeginFrame()
	pop := scoped.Push("profile")
	fromScope := scoped.ClaimDerived(KindClickable, "email", "label")
	pop()

	var resolved Keys
	resolved.BeginFrame()
	fromResolved := resolved.ClaimDerivedResolved(KindClickable, "profile/email", "label")
	if fromResolved != fromScope {
		t.Fatalf("resolved derived key = %q, want scoped key %q", fromResolved, fromScope)
	}
}

func TestKeysRejectEmptyDerivedOwnerWithinScope(t *testing.T) {
	var keys Keys
	keys.BeginFrame()
	pop := keys.Push("profile")
	defer pop()

	defer func() {
		if recover() == nil {
			t.Fatal("expected empty derived owner to panic")
		}
	}()
	keys.ClaimDerived(KindClickable, "", "label")
}

func TestKeysFullKeyEncodesScopedSegments(t *testing.T) {
	got := fullKeyForPath([]string{"a/b", "c%d"}, "e/f%g")
	const want = "a%2Fb/c%25d/e%2Ff%25g"
	if got != want {
		t.Fatalf("full key = %q, want %q", got, want)
	}
}

func TestKeysFullKeyDistinguishesPathSegments(t *testing.T) {
	tests := []struct {
		name      string
		firstPath []string
		firstKey  string
		otherPath []string
		otherKey  string
	}{
		{
			name:      "slash in parent",
			firstPath: []string{"a/b"},
			firstKey:  "c",
			otherPath: []string{"a", "b"},
			otherKey:  "c",
		},
		{
			name:      "slash in leaf",
			firstPath: []string{"a"},
			firstKey:  "b/c",
			otherPath: []string{"a", "b"},
			otherKey:  "c",
		},
		{
			name:      "encoded-looking parent",
			firstPath: []string{"a/b"},
			firstKey:  "c",
			otherPath: []string{"a%2Fb"},
			otherKey:  "c",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			first := fullKeyForPath(tt.firstPath, tt.firstKey)
			other := fullKeyForPath(tt.otherPath, tt.otherKey)
			if first == other {
				t.Fatalf("distinct paths produced the same key %q", first)
			}
		})
	}
}

func TestKeysClaimAllowsPreviouslyCollidingPaths(t *testing.T) {
	var keys Keys
	keys.BeginFrame()

	popAB := keys.Push("a/b")
	first := keys.Claim(KindClickable, "c")
	popAB()

	popA := keys.Push("a")
	popB := keys.Push("b")
	second := keys.Claim(KindClickable, "c")
	popB()
	popA()

	if first == second {
		t.Fatalf("distinct paths produced the same claimed key %q", first)
	}
}

func TestKeysFullKeyDistinguishesRootFromScopedKey(t *testing.T) {
	var root Keys
	rootKey := root.FullKey("a/b")
	scopedKey := fullKeyForPath([]string{"a"}, "b")

	if rootKey == scopedKey {
		t.Fatalf("root and scoped paths produced the same key %q", rootKey)
	}
}

func fullKeyForPath(path []string, key string) string {
	var keys Keys
	for _, part := range path {
		keys.Push(part)
	}
	return keys.FullKey(key)
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
