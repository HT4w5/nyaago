package server

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

/*
 * API handler functions
 */

const (
	errorKey = "err"
	infoKey  = "msg"
)

func (s *Server) HandlePing(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		infoKey: "pong",
	})
}
