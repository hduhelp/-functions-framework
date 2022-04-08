package main

import (
	"github.com/gin-gonic/gin"
)

var Instance Server

type Server struct{}

func (s *Server) Handle(r *gin.RouterGroup) {
	r.GET("/", func(c *gin.Context) {
		c.String(200, "Hello World")
	})
}
