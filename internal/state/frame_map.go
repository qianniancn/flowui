package state

func BeginFrameMap(seen *map[string]struct{}) {
	if *seen == nil {
		*seen = make(map[string]struct{})
		return
	}
	clear(*seen)
}

func UseFrameMap[T any](values *map[string]*T, seen *map[string]struct{}, key string) *T {
	beginFrameMapIfNeeded(seen)
	(*seen)[key] = struct{}{}
	return EnsureFrameMap(values, key)
}

func EnsureFrameMap[T any](values *map[string]*T, key string) *T {
	if *values == nil {
		*values = make(map[string]*T)
	}
	if value := (*values)[key]; value != nil {
		return value
	}
	value := new(T)
	(*values)[key] = value
	return value
}

func SweepFrameMap[T any](values map[string]*T, seen map[string]struct{}) {
	for key := range values {
		if _, ok := seen[key]; !ok {
			delete(values, key)
		}
	}
}

func beginFrameMapIfNeeded(seen *map[string]struct{}) {
	if *seen == nil {
		*seen = make(map[string]struct{})
	}
}
