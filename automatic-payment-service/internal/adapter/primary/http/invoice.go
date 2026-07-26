package http

import (
	"github.com/google/uuid"
	"gofr.dev/pkg/gofr"
	gofrhttp "gofr.dev/pkg/gofr/http"
)

func (h *Handler) GetInvoice(c *gofr.Context) (interface{}, error) {
	id, err := uuid.Parse(c.PathParam("id"))
	if err != nil {
		return nil, gofrhttp.ErrorInvalidParam{Params: []string{"id"}}
	}

	invoice, err := h.invoice.GetInvoice(c, id)
	if err != nil {
		return nil, mapError(err)
	}
	return invoice, nil
}

func (h *Handler) ListInvoices(c *gofr.Context) (interface{}, error) {
	userID, err := uuid.Parse(c.Param("userId"))
	if err != nil {
		return nil, gofrhttp.ErrorInvalidParam{Params: []string{"userId"}}
	}

	invoices, err := h.invoice.ListInvoicesByUser(c, userID)
	if err != nil {
		return nil, mapError(err)
	}
	return invoices, nil
}
