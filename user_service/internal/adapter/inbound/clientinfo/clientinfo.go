package clientinfo

import (
	"context"
	"net"
	"net/http"
	"strings"

	"github.com/JIeeiroSst/user-service/internal/domain"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/peer"
)

func FromHTTP(r *http.Request) context.Context {
	ip := firstForwarded(r.Header.Get("X-Forwarded-For"))
	if ip == "" {
		ip = hostOnly(r.RemoteAddr)
	}
	return domain.WithClientInfo(r.Context(), domain.ClientInfo{IP: ip, UserAgent: r.UserAgent()})
}

func UnaryInterceptor(ctx context.Context, req any, _ *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (any, error) {
	var info domain.ClientInfo
	if md, ok := metadata.FromIncomingContext(ctx); ok {
		info.IP = firstForwarded(first(md, "x-forwarded-for"))
		info.UserAgent = first(md, "grpcgateway-user-agent")
		if info.UserAgent == "" {
			info.UserAgent = first(md, "user-agent")
		}
	}
	if info.IP == "" {
		if p, ok := peer.FromContext(ctx); ok && p.Addr != nil {
			info.IP = hostOnly(p.Addr.String())
		}
	}
	return handler(domain.WithClientInfo(ctx, info), req)
}

func first(md metadata.MD, key string) string {
	if v := md.Get(key); len(v) > 0 {
		return v[0]
	}
	return ""
}

func firstForwarded(v string) string {
	if v == "" {
		return ""
	}
	return strings.TrimSpace(strings.Split(v, ",")[0])
}

func hostOnly(addr string) string {
	if host, _, err := net.SplitHostPort(addr); err == nil {
		return host
	}
	return addr
}
