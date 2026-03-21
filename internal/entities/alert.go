package entities

import "github.com/vekshinnikita/pulse_watch/internal/constants"

type CreateAlertRule struct {
	AppId     *int               `json:"app_id" binding:"omitempty"`
	UserId    int                `json:"user_id" binding:"required"`
	Name      string             `json:"name" binding:"required"`
	Level     constants.LogLevel `json:"level" binding:"required,oneof=CRITICAL ERROR WARNING"`
	Threshold int                `json:"threshold" binding:"required,min=1"`
	Interval  int                `json:"interval" binding:"required,min=1"`
	Message   string             `json:"message" binding:"required"`
}

type UpdateAlertRule struct {
	UserId    *int                `json:"user_id" binding:"omitempty"`
	Name      *string             `json:"name" binding:"omitempty"`
	Level     *constants.LogLevel `json:"level" binding:"omitempty,oneof=CRITICAL ERROR WARNING"`
	Threshold *int                `json:"threshold" binding:"omitempty,min=1"`
	Interval  *int                `json:"interval" binding:"omitempty,min=1"`
	Message   *string             `json:"message" binding:"omitempty"`
}
