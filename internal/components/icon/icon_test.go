package icon

import (
	"image"
	"image/color"
	"sync"
	"testing"

	"gioui.org/layout"
	"gioui.org/op"
	"github.com/qianniancn/flowui/internal/frame"
	"github.com/qianniancn/flowui/internal/locale"
)

var testIconVG = []byte("\x89IVG\x02\n\x00PP\xb0\xb0\xc0d\u014b\x125r\x15\x8b\x95r\x95\x8a\x95\x8a\x95r\x15\x8b5r\u014bd}\x8c\rr%\x8dYr\xa9\x8d\xddr\xf5\x8d\x85s\x9c=t\u034d\xedtm\x8dmumum\x8d\xedt\u034d=t\x9c\x85s\xf5\x8d\xddr\xa9\x8dYr%\x8d\rr}\x8c\xe2d=t\x12\rr\x85sYr\xddr\xddrYr\x85s\rr=td\xedt5rmu\x95rm\x8d\x95\x8a\u034d\x15\x8b\x9c\u014b\xf5\x8d}\x8c\xa9\x8d%\x8d%\x8d\xa9\x8d}\x8c\xf5\x8d\u014b\x9c\x15\x8b\u034d\x95\x8am\x8d\x95rmu5r\xedt\xe1")

var testIcon = testIconVG

func TestIconUsesDefaultAndConfiguredSizes(t *testing.T) {
	ctx := frame.New(nil, nil, locale.LanguageEnglish)
	if dims := New(testIcon).Layout(ctx, iconTestContext()); dims.Size != image.Pt(24, 24) {
		t.Fatalf("default size = %v, want (24,24)", dims.Size)
	}
	if dims := New(testIcon).Size(16).Layout(ctx, iconTestContext()); dims.Size != image.Pt(16, 16) {
		t.Fatalf("configured size = %v, want (16,16)", dims.Size)
	}
}

func TestIconBaselineUsesConfiguredDistanceFromBottom(t *testing.T) {
	ctx := frame.New(nil, nil, locale.LanguageEnglish)
	gtx := iconTestContext()
	gtx.Constraints = layout.Exact(image.Pt(13, 13))

	dims := New(testIcon).Baseline(1).Layout(ctx, gtx)
	if dims.Baseline != 1 {
		t.Fatalf("icon baseline = %d, want 1", dims.Baseline)
	}

	if measured := New(testIcon).Baseline(1).Measure(ctx, gtx); measured.Baseline != 1 {
		t.Fatalf("measured icon baseline = %d, want 1", measured.Baseline)
	}
}

func TestIconBaselineClampsToHeight(t *testing.T) {
	ctx := frame.New(nil, nil, locale.LanguageEnglish)
	gtx := iconTestContext()
	gtx.Constraints = layout.Exact(image.Pt(13, 13))

	dims := New(testIcon).Baseline(20).Layout(ctx, gtx)
	if dims.Baseline != 13 {
		t.Fatalf("icon baseline = %d, want 13", dims.Baseline)
	}
}

func TestNilIconHasNoLayout(t *testing.T) {
	dims := New(nil).Layout(frame.New(nil, nil, locale.LanguageEnglish), iconTestContext())
	if dims.Size != (image.Point{}) {
		t.Fatalf("nil icon size = %v, want zero", dims.Size)
	}
}

func TestIconHandlesNonSquareParentConstraints(t *testing.T) {
	ctx := frame.New(nil, nil, locale.LanguageEnglish)
	gtx := iconTestContext()
	gtx.Constraints = layout.Exact(image.Pt(40, 24))
	dims := New(testIcon).Layout(ctx, gtx)
	if dims.Size != image.Pt(40, 24) {
		t.Fatalf("non-square layout size = %v, want (40,24)", dims.Size)
	}
}

func TestIconLayoutConcurrent(t *testing.T) {
	const workers = 8
	var wg sync.WaitGroup
	wg.Add(workers)
	for worker := range workers {
		go func() {
			defer wg.Done()
			for iteration := range 25 {
				size := 16 + (worker+iteration)%3*4
				gtx := iconTestContext()
				gtx.Constraints = layout.Exact(image.Pt(size, size))
				col := color.NRGBA{R: uint8(worker), G: uint8(iteration), A: 0xff}
				if dims := New(testIcon).Color(col).Layout(frame.New(nil, nil, locale.LanguageEnglish), gtx); dims.Size != image.Pt(size, size) {
					t.Errorf("icon size = %v, want (%d,%d)", dims.Size, size, size)
				}
			}
		}()
	}
	wg.Wait()
}

func TestIconCacheEvictsLeastRecentlyUsedEntries(t *testing.T) {
	resetIconCacheForTest()
	defer resetIconCacheForTest()

	data := make([][]byte, iconCacheMaxEntries+1)
	for i := range data[:iconCacheMaxEntries] {
		data[i] = append([]byte(nil), testIconVG...)
		gtx := iconTestContext()
		New(data[i]).Layout(frame.New(nil, nil, locale.LanguageEnglish), gtx)
	}
	New(data[0]).Layout(frame.New(nil, nil, locale.LanguageEnglish), iconTestContext())
	data[iconCacheMaxEntries] = append([]byte(nil), testIconVG...)
	New(data[iconCacheMaxEntries]).Layout(frame.New(nil, nil, locale.LanguageEnglish), iconTestContext())

	renderers.Lock()
	if got := len(renderers.entries); got != iconCacheMaxEntries {
		renderers.Unlock()
		t.Fatalf("cache entries = %d, want %d", got, iconCacheMaxEntries)
	}
	if renderers.bytes > iconCacheMaxBytes {
		renderers.Unlock()
		t.Fatalf("cache bytes = %d, want <= %d", renderers.bytes, iconCacheMaxBytes)
	}
	if _, ok := renderers.entries[cacheKey{first: &data[0][0], length: len(data[0])}]; !ok {
		renderers.Unlock()
		t.Fatal("recently used icon was evicted")
	}
	if _, ok := renderers.entries[cacheKey{first: &data[1][0], length: len(data[1])}]; ok {
		renderers.Unlock()
		t.Fatal("least recently used icon was retained")
	}
	renderers.Unlock()
}

func iconTestContext() layout.Context {
	return layout.Context{
		Constraints: layout.Constraints{Max: image.Pt(100, 100)},
		Ops:         new(op.Ops),
	}
}

func resetIconCacheForTest() {
	renderers.Lock()
	renderers.entries = make(map[cacheKey]*cacheEntry)
	renderers.lru.Init()
	renderers.bytes = 0
	renderers.Unlock()
}
