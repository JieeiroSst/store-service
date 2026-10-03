package medical_service

import (
	"context"
	"net/http"
	"net/url"
	"strconv"

	"github.com/JIeeiroSst/filter-service/adapter/rest"
	"github.com/JIeeiroSst/filter-service/gql/model"
)

var _ = strconv.Itoa

const Service = "medical-service"

const DefaultBaseURL = "http://medical-service-svc"

type Client struct{ rest *rest.Client }

func New(c *rest.Client) *Client { return &Client{rest: c} }

func (c *Client) Medicines(ctx context.Context, qArg *string, limit *int, offset *int) (*model.MedicalMedicineList, error) {
	path := "/api/v1/medicines"
	q := url.Values{}
	if qArg != nil {
		q.Set("q", *qArg)
	}
	if limit != nil {
		q.Set("limit", strconv.Itoa(*limit))
	}
	if offset != nil {
		q.Set("offset", strconv.Itoa(*offset))
	}
	h := http.Header{}
	var out *model.MedicalMedicineList
	err := c.rest.Get(ctx, path, q, h, "", &out)
	return out, err
}

func (c *Client) Medicine(ctx context.Context, id int) (*model.MedicalMedicine, error) {
	path := "/api/v1/medicines/" + url.PathEscape(strconv.Itoa(id))
	q := url.Values{}
	h := http.Header{}
	var out *model.MedicalMedicine
	err := c.rest.Get(ctx, path, q, h, "", &out)
	return out, err
}

func (c *Client) Stock(ctx context.Context, id int) (*model.MedicalStockSummary, error) {
	path := "/api/v1/medicines/" + url.PathEscape(strconv.Itoa(id)) + "/stock"
	q := url.Values{}
	h := http.Header{}
	var out *model.MedicalStockSummary
	err := c.rest.Get(ctx, path, q, h, "", &out)
	return out, err
}

func (c *Client) Movements(ctx context.Context, id int, limit *int, offset *int) ([]*model.MedicalStockMovement, error) {
	path := "/api/v1/medicines/" + url.PathEscape(strconv.Itoa(id)) + "/movements"
	q := url.Values{}
	if limit != nil {
		q.Set("limit", strconv.Itoa(*limit))
	}
	if offset != nil {
		q.Set("offset", strconv.Itoa(*offset))
	}
	h := http.Header{}
	var out []*model.MedicalStockMovement
	err := c.rest.Get(ctx, path, q, h, "items", &out)
	return out, err
}

func (c *Client) Expiring(ctx context.Context, days *int) ([]*model.MedicalBatch, error) {
	path := "/api/v1/inventory/expiring"
	q := url.Values{}
	if days != nil {
		q.Set("days", strconv.Itoa(*days))
	}
	h := http.Header{}
	var out []*model.MedicalBatch
	err := c.rest.Get(ctx, path, q, h, "items", &out)
	return out, err
}

func (c *Client) LowStock(ctx context.Context) ([]*model.MedicalLowStockItem, error) {
	path := "/api/v1/inventory/low-stock"
	q := url.Values{}
	h := http.Header{}
	var out []*model.MedicalLowStockItem
	err := c.rest.Get(ctx, path, q, h, "items", &out)
	return out, err
}

func (c *Client) Interactions(ctx context.Context, limit *int, offset *int) ([]*model.MedicalInteractionRule, error) {
	path := "/api/v1/interactions"
	q := url.Values{}
	if limit != nil {
		q.Set("limit", strconv.Itoa(*limit))
	}
	if offset != nil {
		q.Set("offset", strconv.Itoa(*offset))
	}
	h := http.Header{}
	var out []*model.MedicalInteractionRule
	err := c.rest.Get(ctx, path, q, h, "items", &out)
	return out, err
}

func (c *Client) Terminology(ctx context.Context) (*model.MedicalTerminologyList, error) {
	path := "/api/v1/terminology"
	q := url.Values{}
	h := http.Header{}
	var out *model.MedicalTerminologyList
	err := c.rest.Get(ctx, path, q, h, "", &out)
	return out, err
}

func (c *Client) Dispenses(ctx context.Context, patientRef *string, limit *int, offset *int) ([]*model.MedicalDispense, error) {
	path := "/api/v1/dispenses"
	q := url.Values{}
	if patientRef != nil {
		q.Set("patient_ref", *patientRef)
	}
	if limit != nil {
		q.Set("limit", strconv.Itoa(*limit))
	}
	if offset != nil {
		q.Set("offset", strconv.Itoa(*offset))
	}
	h := http.Header{}
	var out []*model.MedicalDispense
	err := c.rest.Get(ctx, path, q, h, "items", &out)
	return out, err
}

func (c *Client) Dispense(ctx context.Context, id int) (*model.MedicalDispense, error) {
	path := "/api/v1/dispenses/" + url.PathEscape(strconv.Itoa(id))
	q := url.Values{}
	h := http.Header{}
	var out *model.MedicalDispense
	err := c.rest.Get(ctx, path, q, h, "", &out)
	return out, err
}
