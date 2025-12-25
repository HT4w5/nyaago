package server

import (
	"encoding/base64"
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

func (s *Server) HandleGetRule(c *gin.Context) {
	decoded, err := base64.URLEncoding.DecodeString(c.Param("addr"))
	if err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorJSON{Error: "invalid base64url"})
		return
	}
	addr, err := netip.ParseAddr(string(decoded))
	if err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorJSON{Error: "invalid addr"})
		return
	}

	rule, err := s.denylist.GetRule()
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.MakeErrorJSON(err))
		return
	}
	if !rule.Prefix.IsValid() {
		c.JSON(http.StatusNotFound, dto.ErrorJSON{Error: "rule not found"})
		return
	}

	c.JSON(http.StatusOK, rule.JSON())
}

func (s *Server) HandleDeleteRule(c *gin.Context) {
	decoded, err := base64.URLEncoding.DecodeString(c.Param("prefix"))
	if err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorJSON{Error: "invalid base64url"})
		return
	}
	prefix, err := netip.ParsePrefix(string(decoded))
	if err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorJSON{Error: "invalid prefix"})
		return
	}

	tx, err := s.db.Begin()
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.MakeErrorJSON(err))
		return
	}

	err = tx.DelRule(prefix)
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.MakeErrorJSON(err))
		return
	}

	err = tx.Commit()
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.MakeErrorJSON(err))
		return
	}

	c.Status(http.StatusNoContent)
}

func (s *Server) HandlePutRule(c *gin.Context) {
	decoded, err := base64.URLEncoding.DecodeString(c.Param("prefix"))
	if err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorJSON{Error: "invalid base64url"})
		return
	}
	prefix, err := netip.ParsePrefix(string(decoded))
	if err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorJSON{Error: "invalid prefix"})
		return
	}

	var ruleJSON dto.RuleJSON
	if err := c.ShouldBindJSON(&ruleJSON); err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorJSON{Error: "invalid rule"})
		return
	}

	rule, err := ruleJSON.ToObject()
	if err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorJSON{Error: "invalid rule"})
		return
	}

	rule.Prefix = prefix

	tx, err := s.db.Begin()
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.MakeErrorJSON(err))
		return
	}

	err = tx.PutRule(rule)
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.MakeErrorJSON(err))
		return
	}

	err = tx.Commit()
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.MakeErrorJSON(err))
		return
	}

	c.Status(http.StatusNoContent)
}
