package graph

import (
	"context"

	"github.com/JIeeiroSst/gateway-api/graph/model"
	"github.com/JIeeiroSst/gateway-api/proxy/rest"
)

func (r *mutationResolver) LotteryScrape(ctx context.Context, date string) (*model.LotteryScrape, error) {
	return restCall[*model.LotteryScrape](ctx, r.Proxy, "lottery-service", rest.Request{Method: "POST", Path: "/api/v1/scrape", Body: map[string]any{"date": date}}, "")
}

func (r *mutationResolver) LotteryBackfill(ctx context.Context, start string, end string, force *bool) (*model.LotteryBackfill, error) {
	return restCall[*model.LotteryBackfill](ctx, r.Proxy, "lottery-service", rest.Request{Method: "POST", Path: "/api/v1/backfill", Body: map[string]any{"start": start, "end": end, "force": force}}, "")
}

func (r *queryResolver) LotteryResults(ctx context.Context, date string, province string) (*model.LotteryResult, error) {
	return restCall[*model.LotteryResult](ctx, r.Proxy, "lottery-service", rest.Request{Method: "GET", Path: "/api/v1/results", Query: restQuery("date", date, "province", province)}, "")
}

func (r *queryResolver) LotteryCheck(ctx context.Context, date string, province string, ticketNumber string) (*model.LotteryCheck, error) {
	return restCall[*model.LotteryCheck](ctx, r.Proxy, "lottery-service", rest.Request{Method: "POST", Path: "/api/v1/check", Body: map[string]any{"date": date, "province": province, "ticket_number": ticketNumber}}, "")
}
