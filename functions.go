package functions

import (
	"github.com/gin-gonic/gin"
)

type GroupHandler interface {
	Handle(r *gin.RouterGroup)
}
