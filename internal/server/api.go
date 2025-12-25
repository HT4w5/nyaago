package server

import (
	"net/http"

	"github.com/HT4w5/nyaago/pkg/dto"
	"github.com/gin-gonic/gin"
)

/*
 * API handler functions
 */

func (s *Server) HandlePing(c *gin.Context) {
	c.JSON(http.StatusOK, dto.MakePingJSON())
}

// -- Rule handlers --

func (s *Server) HandleGetRules(c *gin.Context) {
	rules, err := s.denylist.ListRules()
	if err != nil {
		c.JSON(
			http.StatusInternalServerError,
			dto.MakeErrorJSON(err),
		)
		return
	}

	resp := make([]dto.RuleJSON, 0, len(rules))

	for _, v := range rules {
		resp = append(resp, v.JSON())
	}

	c.JSON(http.StatusOK, resp)
}
