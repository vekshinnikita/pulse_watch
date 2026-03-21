package utils

func SetToSlice[T comparable](set map[T]struct{}) []T {
	slice := make([]T, 0)
	for key, _ := range set {
		slice = append(slice, key)
	}

	return slice
}

func SetFromSlice[T any, V comparable](slice []T, getKeyFn func(v T) V) map[V]struct{} {
	set := make(map[V]struct{})

	for _, value := range slice {
		key := getKeyFn(value)
		set[key] = struct{}{}
	}

	return set
}

func SetFromSliceDefault[T comparable](slice []T) map[T]struct{} {
	set := make(map[T]struct{})

	for _, value := range slice {
		set[value] = struct{}{}
	}

	return set
}

func Set[T comparable](args ...T) map[T]struct{} {
	fn := func(v T) T {
		return v
	}
	return SetFromSlice(args, fn)
}
