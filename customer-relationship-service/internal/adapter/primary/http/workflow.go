package http

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"net/http"
	"time"

	"github.com/JIeeiroSst/customer-relationship-service/internal/domain/port"
	"github.com/gin-gonic/gin"
)

const maxSignBody = 16 << 20

type leadHandler struct{ converter port.LeadConverter }

func NewLeadHandler(c port.LeadConverter) *leadHandler { return &leadHandler{converter: c} }

func (h *leadHandler) Register(api *gin.RouterGroup) {
	api.POST("/leads/:id/convert", func(c *gin.Context) {
		id, ok := parseID(c)
		if !ok {
			return
		}
		var in port.ConvertLeadInput
		if c.Request.ContentLength != 0 {
			if err := c.ShouldBindJSON(&in); err != nil {
				c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
				return
			}
		}
		result, err := h.converter.Convert(c.Request.Context(), id, in)
		if err != nil {
			writeError(c, err)
			return
		}
		c.JSON(http.StatusOK, result)
	})
}

type opportunityHandler struct{ workflow port.OpportunityWorkflow }

func NewOpportunityHandler(w port.OpportunityWorkflow) *opportunityHandler {
	return &opportunityHandler{workflow: w}
}

func (h *opportunityHandler) Register(api *gin.RouterGroup) {
	api.POST("/opportunities/:id/stage", func(c *gin.Context) {
		id, ok := parseID(c)
		if !ok {
			return
		}
		var body struct {
			Stage string `json:"stage" binding:"required"`
		}
		if err := c.ShouldBindJSON(&body); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		result, err := h.workflow.ChangeStage(c.Request.Context(), id, body.Stage)
		if err != nil {
			writeError(c, err)
			return
		}
		c.JSON(http.StatusOK, result)
	})
}

type contractHandler struct{ workflow port.ContractWorkflow }

func NewContractHandler(w port.ContractWorkflow) *contractHandler {
	return &contractHandler{workflow: w}
}

func (h *contractHandler) Register(api *gin.RouterGroup) {
	g := api.Group("/contracts/:id")
	g.GET("/signing-payload", func(c *gin.Context) {
		id, ok := parseID(c)
		if !ok {
			return
		}
		endDate, err := time.Parse(time.RFC3339, c.Query("end_date"))
		signedBy := c.Query("signed_by")
		if err != nil || signedBy == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "signed_by and end_date (RFC3339) are required"})
			return
		}
		signingTime := time.Now()
		if v := c.Query("signing_time"); v != "" {
			if signingTime, err = time.Parse(time.RFC3339, v); err != nil {
				c.JSON(http.StatusBadRequest, gin.H{"error": "signing_time must be RFC3339"})
				return
			}
		}
		payload, err := h.workflow.SigningPayload(c.Request.Context(), id, signedBy, endDate, signingTime)
		if err != nil {
			writeError(c, err)
			return
		}
		sum := sha256.Sum256(payload)
		c.JSON(http.StatusOK, gin.H{
			"payload":      string(payload),
			"sha256":       hex.EncodeToString(sum[:]),
			"signing_time": signingTime.UTC().Truncate(time.Second).Format(time.RFC3339),
		})
	})
	g.GET("/document", func(c *gin.Context) {
		id, ok := parseID(c)
		if !ok {
			return
		}
		endDate, err := time.Parse(time.RFC3339, c.Query("end_date"))
		signedBy := c.Query("signed_by")
		signingTime, err2 := time.Parse(time.RFC3339, c.Query("signing_time"))
		if err != nil || err2 != nil || signedBy == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "signed_by, end_date and signing_time (RFC3339) are required"})
			return
		}
		doc, err := h.workflow.Document(c.Request.Context(), id, signedBy, endDate, signingTime)
		if err != nil {
			writeError(c, err)
			return
		}
		c.Header("Content-Disposition", fmt.Sprintf(`attachment; filename="contract-%d-renewal.pdf"`, id))
		c.Data(http.StatusOK, "application/pdf", doc)
	})
	g.GET("/signed-document", func(c *gin.Context) {
		id, ok := parseID(c)
		if !ok {
			return
		}
		doc, err := h.workflow.SignedDocument(c.Request.Context(), id)
		if err != nil {
			writeError(c, err)
			return
		}
		c.Data(http.StatusOK, "application/pdf", doc)
	})
	g.POST("/sign", func(c *gin.Context) {
		id, ok := parseID(c)
		if !ok {
			return
		}
		c.Request.Body = http.MaxBytesReader(c.Writer, c.Request.Body, maxSignBody)
		var in port.SignContractInput
		if err := c.ShouldBindJSON(&in); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		result, err := h.workflow.Sign(c.Request.Context(), id, in)
		if err != nil {
			writeError(c, err)
			return
		}
		c.JSON(http.StatusOK, result)
	})
	for path, step := range map[string]func(*gin.Context, uint) (any, error){
		"/submit":  func(c *gin.Context, id uint) (any, error) { return h.workflow.Submit(c.Request.Context(), id) },
		"/approve": func(c *gin.Context, id uint) (any, error) { return h.workflow.Approve(c.Request.Context(), id) },
		"/reject":  func(c *gin.Context, id uint) (any, error) { return h.workflow.Reject(c.Request.Context(), id) },
	} {
		g.POST(path, func(c *gin.Context) {
			id, ok := parseID(c)
			if !ok {
				return
			}
			result, err := step(c, id)
			if err != nil {
				writeError(c, err)
				return
			}
			c.JSON(http.StatusOK, result)
		})
	}
}

type caseHandler struct{ workflow port.CaseWorkflow }

func NewCaseHandler(w port.CaseWorkflow) *caseHandler { return &caseHandler{workflow: w} }

func (h *caseHandler) Register(api *gin.RouterGroup) {
	api.POST("/cases/:id/close", func(c *gin.Context) {
		id, ok := parseID(c)
		if !ok {
			return
		}
		result, err := h.workflow.Close(c.Request.Context(), id)
		if err != nil {
			writeError(c, err)
			return
		}
		c.JSON(http.StatusOK, result)
	})
}

type reportHandler struct {
	overview port.AccountOverviewUsecase
	reports  port.ReportUsecase
}

func NewReportHandler(o port.AccountOverviewUsecase, r port.ReportUsecase) *reportHandler {
	return &reportHandler{overview: o, reports: r}
}

func (h *reportHandler) Register(api *gin.RouterGroup) {
	api.GET("/accounts/:id/overview", func(c *gin.Context) {
		id, ok := parseID(c)
		if !ok {
			return
		}
		result, err := h.overview.Overview(c.Request.Context(), id)
		if err != nil {
			writeError(c, err)
			return
		}
		c.JSON(http.StatusOK, result)
	})
	api.GET("/reports/summary", func(c *gin.Context) {
		result, err := h.reports.Summary(c.Request.Context())
		if err != nil {
			writeError(c, err)
			return
		}
		c.JSON(http.StatusOK, result)
	})
}
