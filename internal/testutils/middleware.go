package testutils

import (
	"github.com/gin-gonic/gin"
)

func MiddlewaresSetup() {
	BaseSetup()

	// Убираем вывод информации от gin
	gin.SetMode(gin.ReleaseMode)
}
