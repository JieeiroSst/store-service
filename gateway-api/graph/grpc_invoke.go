package graph

import (
	"context"
	"encoding/json"

	"github.com/JIeeiroSst/gateway-api/auth"
	"github.com/JIeeiroSst/gateway-api/proxy"
	"github.com/JIeeiroSst/gateway-api/proxy/protomap"
	"github.com/JIeeiroSst/utils/logger"
	"google.golang.org/grpc"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/reflect/protoreflect"
	"google.golang.org/protobuf/reflect/protoregistry"

	_ "github.com/JIeeiroSst/gateway-api/proxy/pb/accounttxpb"
	_ "github.com/JIeeiroSst/gateway-api/proxy/pb/carrentalpb"
	_ "github.com/JIeeiroSst/gateway-api/proxy/pb/hospitalpb"
)

var protoIn = protojson.UnmarshalOptions{DiscardUnknown: true}

func grpcInvoke[Out any](ctx context.Context, p *proxy.Proxy, service, method, reqType, resType string, input any) (*Out, error) {
	res, err := grpcInvokeRaw(ctx, p, service, method, reqType, resType, input)
	if err != nil {
		return nil, err
	}
	payload, err := json.Marshal(protomap.ToMap(res))
	if err != nil {
		return nil, internalError(ctx, service, method, err)
	}
	out := new(Out)
	if err := json.Unmarshal(payload, out); err != nil {
		return nil, internalError(ctx, service, method, err)
	}
	return out, nil
}

func grpcInvokeEmpty(ctx context.Context, p *proxy.Proxy, service, method, reqType, resType string, input any) (bool, error) {
	if _, err := grpcInvokeRaw(ctx, p, service, method, reqType, resType, input); err != nil {
		return false, err
	}
	return true, nil
}

func grpcInvokeRaw(ctx context.Context, p *proxy.Proxy, service, method, reqType, resType string, input any) (proto.Message, error) {
	req, err := newMessage(reqType)
	if err != nil {
		return nil, internalError(ctx, service, method, err)
	}
	if input != nil {
		payload, err := json.Marshal(input)
		if err != nil {
			return nil, auth.NewError(ctx, auth.CodeBadRequest, "invalid input")
		}
		if err := protoIn.Unmarshal(payload, req); err != nil {
			return nil, auth.NewError(ctx, auth.CodeBadRequest, "invalid input: "+err.Error())
		}
	}
	res, err := newMessage(resType)
	if err != nil {
		return nil, internalError(ctx, service, method, err)
	}
	return grpcCall(ctx, p, service, func(ctx context.Context, conn *grpc.ClientConn) (proto.Message, error) {
		return res, conn.Invoke(ctx, method, req, res)
	})
}

func newMessage(name string) (proto.Message, error) {
	mt, err := protoregistry.GlobalTypes.FindMessageByName(protoreflect.FullName(name))
	if err != nil {
		return nil, err
	}
	return mt.New().Interface(), nil
}

func internalError(ctx context.Context, service, method string, err error) error {
	logger.Error(ctx, "%s %s: %v", service, method, err)
	return auth.NewError(ctx, auth.CodeServiceUnavailable, "unexpected service response")
}
