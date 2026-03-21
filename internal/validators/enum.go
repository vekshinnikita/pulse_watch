package validators

import (
	"fmt"
	"strings"
)

func ValidateEnumValue[T ~string](enum map[T]struct{}, value string) (T, error) {
	typedValue := T(value)

	_, ok := enum[typedValue]
	if !ok {
		var empty T

		values := make([]string, 0, len(enum))
		for k := range enum {
			values = append(values, fmt.Sprintf("'%s'", string(k)))
		}

		return empty, fmt.Errorf("must be one of values: %s", strings.Join(values, ", "))
	}

	return typedValue, nil
}
