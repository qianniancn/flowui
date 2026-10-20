package disclosure

import "testing"

func BenchmarkDisclosureBindingSteadyState(b *testing.B) {
	var binding Binding[bool]
	cfg := Config[bool]{OnChange: func(bool) {}}
	binding.Current(cfg)
	b.ReportAllocs()
	b.ResetTimer()
	for index := 0; b.Loop(); index++ {
		binding.Request(cfg, index&1 == 0)
	}
}

func BenchmarkDisclosureControlledRequest(b *testing.B) {
	var binding Binding[bool]
	cfg := Config[bool]{Controlled: true, OnChange: func(bool) {}}
	b.ReportAllocs()
	b.ResetTimer()
	for index := 0; b.Loop(); index++ {
		binding.Request(cfg, index&1 == 0)
	}
}
