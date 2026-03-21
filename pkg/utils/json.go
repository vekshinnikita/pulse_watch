package utils

import (
	"encoding/json"
	"fmt"
)

func JSONScan(value any, dest any) error {
	if value == nil {
		return nil
	}

	bytes, ok := value.([]byte)
	if !ok {
		structName := GetStructName(dest)
		return fmt.Errorf("%s.Scan: expected []byte, got %T", structName, value)
	}

	return json.Unmarshal(bytes, dest)
}
