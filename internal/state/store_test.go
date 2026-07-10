package state

import "testing"

type frameAwareState struct {
	frames int
}

func (s *frameAwareState) BeginFrame() {
	s.frames++
}

func TestStoreRetainsUsedStateAndSweepsUnusedState(t *testing.T) {
	var store Store
	id := Identity{Key: "settings/language", Slot: "select"}

	store.BeginFrame()
	first := Use[frameAwareState](&store, id, nil)
	store.EndFrame()

	store.BeginFrame()
	second := Use[frameAwareState](&store, id, nil)
	store.EndFrame()
	if first != second {
		t.Fatal("used state was not retained")
	}
	if second.frames != 1 {
		t.Fatalf("BeginFrame calls = %d, want 1", second.frames)
	}

	store.BeginFrame()
	store.EndFrame()
	if store.Len() != 0 {
		t.Fatalf("retained states = %d, want 0", store.Len())
	}
}

func TestStoreSeparatesSlotsForOneComponent(t *testing.T) {
	var store Store
	store.BeginFrame()
	clickable := Use[int](&store, Identity{Key: "save", Slot: "clickable"}, nil)
	animation := Use[int](&store, Identity{Key: "save", Slot: "animation"}, nil)
	if clickable == animation {
		t.Fatal("different slots shared the same state")
	}
}

func TestStoreRejectsTypeMismatch(t *testing.T) {
	var store Store
	id := Identity{Key: "tabs", Slot: "interaction"}
	Use[int](&store, id, nil)
	defer func() {
		if recover() == nil {
			t.Fatal("type mismatch did not panic")
		}
	}()
	Use[string](&store, id, nil)
}

func TestStorePeekDoesNotRetainState(t *testing.T) {
	var store Store
	id := Identity{Key: "dialog", Slot: "modal"}
	store.BeginFrame()
	want := Use[int](&store, id, nil)
	store.EndFrame()

	store.BeginFrame()
	got, ok := Peek[int](&store, id)
	if !ok || got != want {
		t.Fatal("Peek did not return the existing state")
	}
	store.EndFrame()
	if store.Len() != 0 {
		t.Fatal("Peek retained an unused state")
	}
}
