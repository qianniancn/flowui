package core

import (
	"path/filepath"
	"runtime"
	"strconv"
	"sync"
)

// autoKeys names anonymous widgets by their construction site. A widget keeps
// the name it was built with, so a measure pass and a paint pass resolve the
// same identity, while repeated construction at one site (a helper called in a
// loop) still gets distinct names. The per-site counters start over every frame,
// so a stable view tree rebuilds the same names.
var autoKeys struct {
	mu    sync.Mutex
	sites map[string]int
}

func beginAutoKeyFrame() {
	autoKeys.mu.Lock()
	clear(autoKeys.sites)
	autoKeys.mu.Unlock()
}

// AutoKey returns the identity for a widget created at the caller's call site.
// Widget constructors call it; user code only needs Key for reorderable content.
func AutoKey() string {
	_, file, line, ok := runtime.Caller(2)
	if !ok {
		file, line = "unknown", 0
	}
	site := filepath.Base(file) + ":" + strconv.Itoa(line)
	autoKeys.mu.Lock()
	if autoKeys.sites == nil {
		autoKeys.sites = make(map[string]int)
	}
	index := autoKeys.sites[site]
	autoKeys.sites[site] = index + 1
	autoKeys.mu.Unlock()
	return "\x00auto:" + site + ":" + strconv.Itoa(index)
}
