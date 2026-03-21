package models

import (
	"encoding/json"
	"fmt"
)

type JSONB map[string]any

func (j *JSONB) Scan(value any) error {
	if value == nil {
		*j = nil
		return nil
	}

	b, ok := value.([]byte)
	if !ok {
		return fmt.Errorf("cannot scan %T into JSONB", value)
	}

	return json.Unmarshal(b, j)
}
