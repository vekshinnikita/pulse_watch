package response

import (
	"github.com/gin-gonic/gin"
	"github.com/vekshinnikita/pulse_watch/internal/errs"
)

type FieldErrorsResponse struct {
	Message string                     `json:"message"`
	Errors  errs.StructuredFieldErrors `json:"errors"`
}

type ErrorResponse struct {
	Message string `json:"message"`
}

func NewErrorResponse(c *gin.Context, statusCode int, message string) {
	c.AbortWithStatusJSON(statusCode, ErrorResponse{message})
}

func NewFieldsErrorResponse(
	c *gin.Context,
	statusCode int,
	errors errs.StructuredFieldErrors,
) {
	c.AbortWithStatusJSON(statusCode, FieldErrorsResponse{Errors: errors, Message: errs.InvalidInputErrorMessage})
}
