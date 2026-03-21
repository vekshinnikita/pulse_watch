package entities

import "time"

type CreateAppData struct {
	Name        string  `json:"name" binding:"required"`
	Code        string  `json:"code" binding:"required"`
	Description *string `json:"description"`
}

type CreateApiKeyData struct {
	Name      string     `json:"name" binding:"required"`
	AppId     int        `json:"app_id" binding:"required"`
	ExpiresAt *time.Time `json:"expires_at" binding:"omitempty,future"`
}

type PaginationDataWithAppId struct {
	PaginationData
	AppId *int `form:"app_id" json:"app_id" binding:"omitempty"`
}
