package graph

import (
	"context"

	"github.com/99designs/gqlgen/graphql"
	"github.com/JIeeiroSst/gateway-api/graph/model"
	"github.com/JIeeiroSst/gateway-api/proxy/rest"
)

func (r *mutationResolver) EkycSubmitCitizenCard(ctx context.Context, front graphql.Upload, back graphql.Upload) (*model.EkycCitizenIdentity, error) {
	return restMultipart[*model.EkycCitizenIdentity](ctx, r.Proxy, "ekyc-service", rest.Request{Method: "POST", Path: rest.P("/api/v1/ekyc/%v/citizen-card", userIDOf(ctx))}, map[string]any{}, uploadFiles("front", front, "back", back), "")
}

func (r *mutationResolver) EkycSubmitNfcChip(ctx context.Context, efCom *graphql.Upload, dg1 *graphql.Upload, dg2 *graphql.Upload, efSod *graphql.Upload) (*model.EkycCitizenIdentity, error) {
	return restMultipart[*model.EkycCitizenIdentity](ctx, r.Proxy, "ekyc-service", rest.Request{Method: "POST", Path: rest.P("/api/v1/ekyc/%v/nfc-chip", userIDOf(ctx))}, map[string]any{}, uploadFiles("ef_com", efCom, "dg1", dg1, "dg2", dg2, "ef_sod", efSod), "")
}

func (r *mutationResolver) EkycSubmitFaceScan(ctx context.Context, frames []*graphql.Upload) (*model.EkycFaceBiometric, error) {
	return restMultipart[*model.EkycFaceBiometric](ctx, r.Proxy, "ekyc-service", rest.Request{Method: "POST", Path: rest.P("/api/v1/ekyc/%v/face-scan", userIDOf(ctx))}, map[string]any{}, uploadFiles("frames", frames), "")
}

func (r *mutationResolver) EkycVerify(ctx context.Context) (*model.EkycVerification, error) {
	return restCall[*model.EkycVerification](ctx, r.Proxy, "ekyc-service", rest.Request{Method: "POST", Path: rest.P("/api/v1/ekyc/%v/verify", userIDOf(ctx)), Body: map[string]any{}}, "")
}

func (r *queryResolver) MyEkycStatus(ctx context.Context) (*model.EkycStatus, error) {
	return restCall[*model.EkycStatus](ctx, r.Proxy, "ekyc-service", rest.Request{Method: "GET", Path: rest.P("/api/v1/ekyc/%v", userIDOf(ctx))}, "")
}
