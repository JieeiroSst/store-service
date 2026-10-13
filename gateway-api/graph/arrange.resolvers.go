package graph

import (
	"context"

	"github.com/JIeeiroSst/gateway-api/graph/model"
	"github.com/JIeeiroSst/gateway-api/proxy/rest"
)

func (r *mutationResolver) ArrangeDispatch(ctx context.Context, input model.ArrangeDispatchInput) (*model.ArrangeDispatchAssignment, error) {
	return restCall[*model.ArrangeDispatchAssignment](ctx, r.Proxy, "arrange-service", rest.Request{Method: "POST", Path: "/dispatch/assignments", Body: input}, "")
}

func (r *mutationResolver) ArrangeMarkDelivered(ctx context.Context, id int) (bool, error) {
	return restCall[bool](ctx, r.Proxy, "arrange-service", rest.Request{Method: "PATCH", Path: rest.P("/dispatch/assignments/%v/delivered", id), Body: map[string]any{}}, "")
}

func (r *mutationResolver) ArrangeShift(ctx context.Context, input model.ArrangeShiftInput) (*model.ArrangeShift, error) {
	return restCall[*model.ArrangeShift](ctx, r.Proxy, "arrange-service", rest.Request{Method: "POST", Path: "/shifts", Body: input}, "")
}

func (r *mutationResolver) ArrangeCancelShift(ctx context.Context, id int) (bool, error) {
	return restCall[bool](ctx, r.Proxy, "arrange-service", rest.Request{Method: "DELETE", Path: rest.P("/shifts/%v", id)}, "")
}

func (r *mutationResolver) ArrangeSpace(ctx context.Context, input model.ArrangeSpaceInput) (*model.ArrangeSpaceAssignment, error) {
	return restCall[*model.ArrangeSpaceAssignment](ctx, r.Proxy, "arrange-service", rest.Request{Method: "POST", Path: "/spaces/assignments", Body: input}, "")
}

func (r *mutationResolver) ArrangeReleaseSpace(ctx context.Context, id int) (bool, error) {
	return restCall[bool](ctx, r.Proxy, "arrange-service", rest.Request{Method: "DELETE", Path: rest.P("/spaces/assignments/%v", id)}, "")
}

func (r *queryResolver) ArrangeDispatchAssignments(ctx context.Context, driverID int) ([]*model.ArrangeDispatchAssignment, error) {
	return restCall[[]*model.ArrangeDispatchAssignment](ctx, r.Proxy, "arrange-service", rest.Request{Method: "GET", Path: "/dispatch/assignments", Query: restQuery("driverId", driverID)}, "")
}

func (r *queryResolver) ArrangeShifts(ctx context.Context, storeID int) ([]*model.ArrangeShift, error) {
	return restCall[[]*model.ArrangeShift](ctx, r.Proxy, "arrange-service", rest.Request{Method: "GET", Path: "/shifts", Query: restQuery("storeId", storeID)}, "")
}

func (r *queryResolver) ArrangeSpaces(ctx context.Context, storeID int, minCapacity *int) ([]*model.ArrangeSpace, error) {
	return restCall[[]*model.ArrangeSpace](ctx, r.Proxy, "arrange-service", rest.Request{Method: "GET", Path: "/spaces", Query: restQuery("storeId", storeID, "minCapacity", minCapacity)}, "")
}

func (r *queryResolver) ArrangeSpaceAssignments(ctx context.Context, storeID int) ([]*model.ArrangeSpaceAssignment, error) {
	return restCall[[]*model.ArrangeSpaceAssignment](ctx, r.Proxy, "arrange-service", rest.Request{Method: "GET", Path: "/spaces/assignments", Query: restQuery("storeId", storeID)}, "")
}
