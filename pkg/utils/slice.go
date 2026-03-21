package utils

func Map[T any, K any](slice []T, f func(v T) K) []K {
	newSlice := make([]K, 0)
	for _, value := range slice {
		newSlice = append(newSlice, f(value))
	}

	return newSlice
}

func SliceContains[T comparable](slice []T, target T) bool {
	for _, value := range slice {
		if value == target {
			return true
		}
	}

	return false
}

func ProcessBatches[T any](items []T, batchSize int, fn func(batch []T) error) error {
	for i := 0; i < len(items); i += batchSize {
		end := i + batchSize
		if end > len(items) {
			end = len(items)
		}

		err := fn(items[i:end])
		if err != nil {
			return err
		}
	}

	return nil
}
