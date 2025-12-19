package server

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

const (
	infoKey  = "msg"
	errorKey = "err"
)

func (s *Server) setupRoutes() {
	s.router.GET("/ping", s.ping)
}

func (s *Server) ping(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		infoKey: "pong",
	})
}
