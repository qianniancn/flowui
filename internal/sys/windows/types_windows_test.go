//go:build windows

package windows

import "testing"

func TestSizeOfUsesUnderlyingStruct(t *testing.T) {
	if got := SizeOf(WNDCLASSEX{}); got <= 16 {
		t.Fatalf("SizeOf(WNDCLASSEX{}) = %d, want the struct size", got)
	}
	if got := SizeOf(NOTIFYICONDATA{}); got <= 16 {
		t.Fatalf("SizeOf(NOTIFYICONDATA{}) = %d, want the struct size", got)
	}
}
