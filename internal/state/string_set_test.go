package state

import (
	"fmt"
	"testing"
)

func TestStringSetCacheTracksControlledKeys(t *testing.T) {
	keys := make([]string, stringSetThreshold)
	for index := range keys {
		keys[index] = fmt.Sprintf("key-%d", index)
	}
	var cache StringSetCache
	set := cache.Resolve(keys)
	if !StringSetContains(keys, set, keys[len(keys)-1]) {
		t.Fatal("large key set missed an existing key")
	}
	keys[len(keys)-1] = "changed"
	set = cache.Resolve(keys)
	if StringSetContains(keys, set, "key-63") || !StringSetContains(keys, set, "changed") {
		t.Fatal("key set cache did not observe an in-place mutation")
	}
	if set := cache.Resolve(keys[:1]); set != nil || !StringSetContains(keys[:1], set, keys[0]) {
		t.Fatal("short key list did not use linear lookup")
	}
}
