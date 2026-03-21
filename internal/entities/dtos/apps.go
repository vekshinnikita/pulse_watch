package dtos

import (
	"time"
)

type CreateApiKeyData struct {
	Name      string
	AppId     int
	ExpiresAt *time.Time
	CreatedAt time.Time
	KeyHash   string
}
