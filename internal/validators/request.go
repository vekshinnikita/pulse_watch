package validators

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/vekshinnikita/pulse_watch/internal/errs"
	"github.com/vekshinnikita/pulse_watch/pkg/response"
)

func ValidateRequestFields(c *gin.Context, obj interface{}) error {
	if err := c.ShouldBindJSON(&obj); err != nil {

		if errs.IsValidationError(err) {
			structuredErrors := errs.MakeFieldsErrorMap(obj, err)
			response.NewFieldsErrorResponse(c, http.StatusBadRequest, structuredErrors)
			return err
		}

		response.NewErrorResponse(c, http.StatusBadRequest, "bad request input")
		return err
	}

	return nil
}

func ValidateQueryFields(c *gin.Context, obj interface{}) error {
	if err := c.ShouldBindQuery(&obj); err != nil {

		if errs.IsValidationError(err) {
			structuredErrors := errs.MakeFieldsErrorMap(obj, err)
			response.NewFieldsErrorResponse(c, http.StatusBadRequest, structuredErrors)
			return err
		}

		response.NewErrorResponse(c, http.StatusBadRequest, "bad request input")
		return err
	}

	return nil
}
