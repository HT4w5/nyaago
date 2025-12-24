package server

import (
	"net/http"
	"net/netip"

	"github.com/HT4w5/nyaago/pkg/dto"
	"github.com/gin-gonic/gin"
)

/*
 * API handler functions
 */

func (s *Server) HandlePing(c *gin.Context) {
	c.JSON(http.StatusOK, dto.MakePingJSON())
}

// TODO: implement paging (requires modifying db layer)
func (s *Server) HandleGetRules(c *gin.Context) {
	rules, err := s.db.ListRules()
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

func (s *Server) HandleGetRule(c *gin.Context) {
	prefix, err := netip.ParsePrefix(c.Param("prefix"))
	if err != nil {
		c.JSON(http.StatusBadRequest, dto.MakeErrorJSON(err))
		return
	}

	rule, err := s.db.GetRule(prefix)
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.MakeErrorJSON(err))
		return
	}
	if !rule.Prefix.IsValid() {
		c.JSON(http.StatusNotFound, dto.ErrorJSON{Error: "rule not found"})
	}
}
