package handlers

import "github.com/gin-gonic/gin"

type SubHandler interface {
	InitRoutes(r *gin.RouterGroup)
}
