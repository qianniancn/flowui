package style

import "maps"

// Non-slot member of the fluent surface: part declarations. The five storage
// slots (box, paint, text, transform, transitions) live in fluent_box.go,
// fluent_paint.go, fluent_text.go, fluent_transform.go, and fluent_motion.go.

func (s Style) Part(part Part, override Style) Style {
	if part == PartRoot {
		return merge(s, override)
	}
	result := s
	result.parts = make(map[Part]Style, len(s.parts)+1)
	maps.Copy(result.parts, s.parts)
	result.parts[part] = merge(result.parts[part], override)
	return result
}
