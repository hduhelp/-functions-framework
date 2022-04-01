package functions

import (
	"github.com/gin-gonic/gin"
)

type Function interface {
	Handle(r *gin.RouterGroup)
}
