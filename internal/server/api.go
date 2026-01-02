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
	rules, err := s.iplist.ListRules()
	if err != nil {
		c.JSON(
			http.StatusInternalServerError,
			dto.MakeErrorJSON(err),
		)
		return
	}

	c.JSON(http.StatusOK, rules)
}

// -- Record handlers --

func (s *Server) HandleGetRecords(c *gin.Context) {
	records := make([]dto.Record, 0, s.analyzer.Len())
	for v := range s.analyzer.Iterator() {
		records = append(records, v)
	}

	c.JSON(http.StatusOK, records)
}
