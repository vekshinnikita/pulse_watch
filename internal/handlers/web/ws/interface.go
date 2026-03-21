package ws_handler

import "github.com/gin-gonic/gin"

type WSSubHandler interface {
	Handle(r *gin.Context)
	HandleLive(r *gin.Context)
}
