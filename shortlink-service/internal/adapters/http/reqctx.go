package http

import (
	"net"

	"github.com/JIeeiroSst/shortlink-service/internal/domain"
	"github.com/gin-gonic/gin"
)

func ClientIP(c *gin.Context, tp domain.TrustProxy) string {
	return domain.ResolveClientIP(tp, remoteAddrIP(c.Request.RemoteAddr), c.GetHeader("X-Forwarded-For"))
}

func remoteAddrIP(addr string) string {
	host, _, err := net.SplitHostPort(addr)
	if err != nil {
		return addr
	}
	return host
}
