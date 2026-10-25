package di

import (
	"context"
	"fmt"
	"log"
	"net"
	"net/http"

	"github.com/JIeeiroSst/user-service/config"
	"github.com/JIeeiroSst/user-service/internal/adapter/inbound/clientinfo"
	"github.com/JIeeiroSst/user-service/internal/adapter/inbound/grpcadapter"
	"github.com/JIeeiroSst/user-service/internal/adapter/inbound/internalhttp"
	"github.com/JIeeiroSst/user-service/internal/adapter/inbound/mfahttp"
	"github.com/JIeeiroSst/user-service/internal/adapter/inbound/sessionhttp"
	"github.com/JIeeiroSst/user-service/internal/adapter/inbound/ssohttp"
	"github.com/JIeeiroSst/user-service/internal/port/input"
	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"go.uber.org/fx"
	"google.golang.org/grpc"

	userServiceGrpc "github.com/JIeeiroSst/lib-gateway/user-service/gateway/user-service"
)

type SocialLogin struct {
	Name    string
	Service input.SSOService
}

type ServerParams struct {
	fx.In
	Lifecycle fx.Lifecycle
	Config    *config.Config
	Handler   *grpcadapter.Handler
	Users     input.UserService
	Auth      input.AuthService
	Sessions  input.SessionService
	MFA       input.MFAService
	SSO       input.SSOService
	Social    []SocialLogin
}

func RegisterServer(p ServerParams) error {
	lc, cfg, handler := p.Lifecycle, p.Config, p.Handler
	lis, err := net.Listen("tcp", fmt.Sprintf(":%v", cfg.Server.PortGrpcServer))
	if err != nil {
		return fmt.Errorf("failed to listen: %w", err)
	}

	grpcServer := grpc.NewServer(grpc.ChainUnaryInterceptor(clientinfo.UnaryInterceptor, grpcadapter.UnaryErrorInterceptor))
	userServiceGrpc.RegisterUserServiceServer(grpcServer, handler)

	mux := runtime.NewServeMux(runtime.WithErrorHandler(grpcadapter.HTTPErrorHandler))
	if err := internalhttp.New(p.Users, p.Auth, cfg.Internal.ApiTokens).Register(mux); err != nil {
		return fmt.Errorf("failed to register internal routes: %w", err)
	}
	if err := sessionhttp.New(p.Sessions).Register(mux); err != nil {
		return fmt.Errorf("failed to register session routes: %w", err)
	}
	if err := mfahttp.New(p.MFA).Register(mux); err != nil {
		return fmt.Errorf("failed to register mfa routes: %w", err)
	}
	if p.SSO != nil {
		if err := ssohttp.New(p.SSO, ssohttp.SSOBasePath).Register(mux); err != nil {
			return fmt.Errorf("failed to register sso routes: %w", err)
		}
	}
	for _, s := range p.Social {
		if err := ssohttp.New(s.Service, ssohttp.SocialBasePath+s.Name).Register(mux); err != nil {
			return fmt.Errorf("failed to register %s login routes: %w", s.Name, err)
		}
	}
	httpServer := &http.Server{
		Addr:    fmt.Sprintf(":%v", cfg.Server.PortHttpServer),
		Handler: mux,
	}

	lc.Append(fx.Hook{
		OnStart: func(ctx context.Context) error {
			go func() {
				log.Printf("Starting gRPC server on :%v", cfg.Server.PortGrpcServer)
				if err := grpcServer.Serve(lis); err != nil && err != grpc.ErrServerStopped {
					log.Printf("gRPC server stopped: %v", err)
				}
			}()

			opts := []grpc.DialOption{grpc.WithInsecure()}
			endpoint := fmt.Sprintf("localhost:%v", cfg.Server.PortGrpcServer)
			if err := userServiceGrpc.RegisterUserServiceHandlerFromEndpoint(ctx, mux, endpoint, opts); err != nil {
				return fmt.Errorf("failed to register gateway: %w", err)
			}

			go func() {
				log.Printf("Starting HTTP server on :%v", cfg.Server.PortHttpServer)
				if err := httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
					log.Printf("HTTP server stopped: %v", err)
				}
			}()

			return nil
		},
		OnStop: func(ctx context.Context) error {
			grpcServer.GracefulStop()
			return httpServer.Shutdown(ctx)
		},
	})

	return nil
}
