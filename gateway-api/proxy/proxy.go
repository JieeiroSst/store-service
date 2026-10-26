package proxy

import (
	"context"
	"time"

	"github.com/JIeeiroSst/gateway-api/config"
	"github.com/JIeeiroSst/gateway-api/proxy/authorize_proxy"
	"github.com/JIeeiroSst/gateway-api/proxy/breaker"
	"github.com/JIeeiroSst/gateway-api/proxy/rest"
	"github.com/JIeeiroSst/gateway-api/proxy/user_proxy"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

const (
	UserService               = "user-service"
	AuthorizeService          = "authorize-service"
	AccountTransactionService = "account-transaction"
	CarRentalService          = "car-rental-service"
	HospitalService           = "hospital-patient-management-service"
)

type Proxy struct {
	User      user_proxy.UserProxy
	Authorize authorize_proxy.AuthorizeProxy
	rest      map[string]*rest.Client
	grpc      map[string]*grpc.ClientConn
	timeouts  map[string]time.Duration
}

func NewProxy(cfg config.Config) (*Proxy, error) {
	breakers := map[string]*breaker.Breaker{}
	for name, svc := range cfg.Services {
		breakers[name] = newBreaker(name, svc.CircuitBreaker)
	}
	userCfg := cfg.Services[UserService]
	user, err := user_proxy.NewUserProxy(userCfg.GrpcAddress, timeout(userCfg.TimeoutSeconds), breaker.DialOptions(breakers[UserService])...)
	if err != nil {
		return nil, err
	}
	authorizeCfg := cfg.Services[AuthorizeService]
	authorize, err := authorize_proxy.NewAuthorizeProxy(authorizeCfg.GrpcAddress, timeout(authorizeCfg.TimeoutSeconds), breaker.DialOptions(breakers[AuthorizeService])...)
	if err != nil {
		_ = user.Close()
		return nil, err
	}
	p := &Proxy{
		User:      user,
		Authorize: authorize,
		rest:      map[string]*rest.Client{},
		grpc:      map[string]*grpc.ClientConn{},
		timeouts:  map[string]time.Duration{},
	}
	for name, svc := range cfg.Services {
		p.timeouts[name] = timeout(svc.TimeoutSeconds)
		if svc.HttpAddress != "" {
			p.rest[name] = rest.New(name, svc.HttpAddress, p.timeouts[name], rest.Options{
				DropUserToken:     svc.ForwardUserToken != nil && !*svc.ForwardUserToken,
				BasicAuthUser:     svc.BasicAuthUser,
				BasicAuthPassword: svc.BasicAuthPassword,
				Headers:           svc.Headers,
				Breaker:           breakers[name],
			})
		}
		if svc.GrpcAddress != "" {
			opts := append([]grpc.DialOption{grpc.WithTransportCredentials(insecure.NewCredentials())}, breaker.DialOptions(breakers[name])...)
			conn, err := grpc.NewClient(svc.GrpcAddress, opts...)
			if err != nil {
				p.Close()
				return nil, err
			}
			p.grpc[name] = conn
		}
	}
	return p, nil
}

func (p *Proxy) REST(name string) *rest.Client {
	if c, ok := p.rest[name]; ok {
		return c
	}
	return rest.New(name, "", 0)
}

func (p *Proxy) GRPC(ctx context.Context, name string) (*grpc.ClientConn, context.Context, context.CancelFunc) {
	t, ok := p.timeouts[name]
	if !ok {
		t = timeout(0)
	}
	ctx, cancel := context.WithTimeout(outgoing(ctx), t)
	conn, ok := p.grpc[name]
	if !ok {
		conn = unconfigured
	}
	return conn, ctx, cancel
}

func (p *Proxy) Close() {
	if p.User != nil {
		_ = p.User.Close()
	}
	if p.Authorize != nil {
		_ = p.Authorize.Close()
	}
	for _, c := range p.grpc {
		_ = c.Close()
	}
}

func newBreaker(name string, c config.CircuitBreaker) *breaker.Breaker {
	return breaker.New(name, breaker.Settings{
		Disabled:         c.Disabled,
		FailureThreshold: c.FailureThreshold,
		OpenTimeout:      time.Duration(c.OpenSeconds) * time.Second,
		HalfOpenRequests: c.HalfOpenRequests,
	})
}

func timeout(seconds int) time.Duration {
	if seconds <= 0 {
		return 10 * time.Second
	}
	return time.Duration(seconds) * time.Second
}
