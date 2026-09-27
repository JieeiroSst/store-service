package handler

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"context"
	"github.com/gin-gonic/gin"

	"github.com/JIeeiroSst/tool-service/internal/apitest"
)

func (h *Handler) registerSuites(v1 *gin.RouterGroup) {
	s := v1.Group("/suites")
	s.GET("/:name", h.getSuite)
	s.POST("/:name/approve", h.approveCases)
	s.DELETE("/:name/cases/:hash", h.removeCase)
	s.POST("/:name/run", h.runSuite)
}

func (h *Handler) suiteOr503(c *gin.Context) bool {
	if h.suites == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "suites are not enabled"})
		return false
	}
	return true
}

func (h *Handler) getSuite(c *gin.Context) {
	if !h.suiteOr503(c) {
		return
	}
	su, err := h.suites.Get(c.Param("name"))
	if err != nil {
		badRequest(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"suite": su, "precision": su.Precision()})
}

type approveRequest struct {
	ExperienceID string   `json:"experience_id"`
	Approve      []string `json:"approve"`
	Reject       []string `json:"reject"`

	Cases []apitest.TestCase `json:"cases"`
}

func (h *Handler) approveCases(c *gin.Context) {
	if !h.suiteOr503(c) {
		return
	}
	var req approveRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		badRequest(c, err)
		return
	}
	name := c.Param("name")

	var chosen []apitest.TestCase
	fromAI := false
	if req.ExperienceID != "" {
		exp, ok := h.mem.Get(req.ExperienceID)
		if !ok {
			badRequest(c, fmt.Errorf("experience not found"))
			return
		}
		var tr taskResult
		if err := json.Unmarshal(exp.Output, &tr); err != nil {
			badRequest(c, fmt.Errorf("experience has no test cases: %w", err))
			return
		}
		byName := map[string]apitest.TestCase{}
		for _, cs := range tr.Cases {
			byName[cs.Name] = cs
		}
		for _, n := range req.Approve {
			cs, ok := byName[n]
			if !ok {
				badRequest(c, fmt.Errorf("case %q not in that result", n))
				return
			}
			chosen = append(chosen, cs)
		}
		fromAI = true
	}
	chosen = append(chosen, req.Cases...)
	handWritten := len(req.Cases)

	for _, cs := range chosen {
		if err := validateApproved(cs); err != nil {
			badRequest(c, fmt.Errorf("case %q: %w", cs.Name, err))
			return
		}
	}
	su, err := h.suites.Approve(name, chosen[:len(chosen)-handWritten], fromAI, len(req.Reject))
	if err == nil && handWritten > 0 {
		su, err = h.suites.Approve(name, chosen[len(chosen)-handWritten:], false, 0)
	}
	if err != nil {
		badRequest(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"suite": su, "precision": su.Precision()})
}

func validateApproved(c apitest.TestCase) error {
	if c.Path == "" {
		return fmt.Errorf("path is required")
	}
	if c.ExpectStatus == 0 && len(c.ExpectJSON) == 0 && len(c.ExpectBodyContains) == 0 {
		return fmt.Errorf("case asserts nothing: set expect_status, expect_json or expect_body_contains")
	}
	return checkCalc(c)
}

func (h *Handler) removeCase(c *gin.Context) {
	if !h.suiteOr503(c) {
		return
	}
	if err := h.suites.Remove(c.Param("name"), c.Param("hash")); err != nil {
		badRequest(c, err)
		return
	}
	c.Status(http.StatusNoContent)
}

func (h *Handler) runSuite(c *gin.Context) {
	if !h.suiteOr503(c) {
		return
	}
	var req struct {
		BaseURL string `json:"base_url" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		badRequest(c, err)
		return
	}
	if err := h.checkTarget(req.BaseURL); err != nil {
		badRequest(c, err)
		return
	}
	su, err := h.suites.Get(c.Param("name"))
	if err != nil {
		badRequest(c, err)
		return
	}
	if len(su.Cases) == 0 {
		badRequest(c, fmt.Errorf("suite %q has no approved cases", su.Name))
		return
	}
	cases := make([]apitest.TestCase, len(su.Cases))
	trust := make([]caseTrust, len(su.Cases))
	for i, a := range su.Cases {
		cases[i] = a.Case
		trust[i] = caseTrust{Name: a.Case.Name, Hash: a.Hash, Trusted: true, Approved: true, Critical: isMoneyRelated(a.Case)}
	}
	ctx, cancel := context.WithTimeout(c.Request.Context(), 10*time.Minute)
	defer cancel()

	report := apitest.Run(ctx, h.httpClient, req.BaseURL, cases)
	res := taskResult{Model: "none (deterministic)", Source: "suite:" + su.Name, Suite: su.Name, Cases: cases, Trust: trust, Report: &report}
	res.Gate = evaluateGate(res, 1.0, true)

	status := http.StatusOK
	if !res.Gate.Passed && c.Query("fail_http") == "true" {
		status = http.StatusUnprocessableEntity
	}
	if c.Query("format") == "junit" {
		c.Data(status, "application/xml", toJUnit(res))
		return
	}
	c.JSON(status, res)
}
