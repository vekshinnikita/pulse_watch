package utils

func MapValues[K comparable, T any](data map[K]T) []T {
	values := make([]T, len(data))

	i := 0
	for _, value := range data {
		values[i] = value
		i++
	}

	return values
}

func MapKeys[K comparable, T any](data map[K]T) []K {
	keys := make([]K, len(data))

	i := 0
	for key, _ := range data {
		keys[i] = key
		i++
	}

	return keys
}

func CopyMap[K comparable, T any](data map[K]T) map[K]T {
	copy := make(map[K]T)

	for key, value := range data {
		copy[key] = value
	}

	return copy
}

func SwapMap[K comparable, T comparable](data map[K]T) map[T]K {
	swapped := make(map[T]K)

	for key, value := range data {
		swapped[value] = key
	}

	return swapped
}
