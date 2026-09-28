package http

import (
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/JIeeiroSst/vending-machine-service/internal/domain"
	"github.com/gin-gonic/gin"
)

const defaultReportDays = 30

func (h *Handler) RegisterMachine(c *gin.Context) {
	var req machineRequest
	if !bind(c, &req) {
		return
	}
	m := &domain.Machine{Location: req.Location, Model: req.Model}
	if req.Status != "" {
		status, err := domain.ParseMachineStatus(req.Status)
		if err != nil {
			writeError(c, err)
			return
		}
		m.Status = status
	}
	m, err := h.machines.Register(c.Request.Context(), m)
	if err != nil {
		writeError(c, err)
		return
	}
	c.JSON(http.StatusCreated, toMachine(m))
}

func (h *Handler) ListMachines(c *gin.Context) {
	page, err := pageRequest(c)
	if err != nil {
		writeError(c, err)
		return
	}
	machines, err := h.machines.List(c.Request.Context(), c.Query("status"), page)
	if err != nil {
		writeError(c, err)
		return
	}
	c.JSON(http.StatusOK, toPage(machines, page, toMachine))
}

func (h *Handler) GetMachine(c *gin.Context) {
	m, err := h.machines.Get(c.Request.Context(), c.Param("id"))
	if err != nil {
		writeError(c, err)
		return
	}
	c.JSON(http.StatusOK, toMachine(m))
}

func (h *Handler) ChangeMachineStatus(c *gin.Context) {
	var req statusRequest
	if !bind(c, &req) {
		return
	}
	status, err := domain.ParseMachineStatus(req.Status)
	if err != nil {
		writeError(c, err)
		return
	}
	m, err := h.machines.ChangeStatus(c.Request.Context(), c.Param("id"), status)
	if err != nil {
		writeError(c, err)
		return
	}
	c.JSON(http.StatusOK, toMachine(m))
}

func (h *Handler) RecordMaintenance(c *gin.Context) {
	var req maintenanceRequest
	if !bind(c, &req) {
		return
	}
	l, err := h.machines.RecordMaintenance(c.Request.Context(), &domain.MaintenanceLog{
		MachineID:       c.Param("id"),
		TechnicianID:    req.TechnicianID,
		MaintenanceType: req.MaintenanceType,
		Notes:           req.Notes,
	})
	if err != nil {
		writeError(c, err)
		return
	}
	c.JSON(http.StatusCreated, toMaintenance(l))
}

func (h *Handler) ListMaintenance(c *gin.Context) {
	logs, err := h.machines.ListMaintenance(c.Request.Context(), c.Param("id"))
	if err != nil {
		writeError(c, err)
		return
	}
	c.JSON(http.StatusOK, mapSlice(logs, toMaintenance))
}

func (h *Handler) ListEvents(c *gin.Context) {
	limit, _ := strconv.Atoi(c.Query("limit"))
	events, err := h.machines.ListEvents(c.Request.Context(), c.Param("id"), limit)
	if err != nil {
		writeError(c, err)
		return
	}
	c.JSON(http.StatusOK, mapSlice(events, toEvent))
}

func (h *Handler) UpdateMachine(c *gin.Context) {
	var req machinePatchRequest
	if !bind(c, &req) {
		return
	}
	m, err := h.machines.Update(c.Request.Context(), c.Param("id"), domain.MachinePatch{
		Location: req.Location,
		Model:    req.Model,
	})
	if err != nil {
		writeError(c, err)
		return
	}
	c.JSON(http.StatusOK, toMachine(m))
}

func (h *Handler) SalesReport(c *gin.Context) {
	to := time.Now().UTC()
	from := to.AddDate(0, 0, -defaultReportDays)
	var err error
	if raw := c.Query("from"); raw != "" {
		if from, err = parseTime(raw); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "from: " + err.Error()})
			return
		}
	}
	if raw := c.Query("to"); raw != "" {
		if to, err = parseTime(raw); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "to: " + err.Error()})
			return
		}
	}
	report, err := h.machines.SalesReport(c.Request.Context(), c.Param("id"), from, to)
	if err != nil {
		writeError(c, err)
		return
	}
	c.JSON(http.StatusOK, toSalesReport(report))
}

func parseTime(raw string) (time.Time, error) {
	if t, err := time.Parse(time.RFC3339, raw); err == nil {
		return t.UTC(), nil
	}
	t, err := time.Parse(time.DateOnly, raw)
	if err != nil {
		return time.Time{}, fmt.Errorf("use RFC3339 or YYYY-MM-DD")
	}
	return t, nil
}
