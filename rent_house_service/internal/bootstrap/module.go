// Package bootstrap wires the hexagon together with uber-go/fx: adapters are
// provided as their port interfaces so services never see a concrete adapter.
package bootstrap

import (
	"context"
	"errors"
	"log/slog"
	"net"
	nethttp "net/http"
	"os"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"go.uber.org/fx"

	"github.com/JIeerioSst/rent-house-service/config"
	httpadapter "github.com/JIeerioSst/rent-house-service/internal/adapter/inbound/http"
	"github.com/JIeerioSst/rent-house-service/internal/adapter/outbound/paymentservice"
	"github.com/JIeerioSst/rent-house-service/internal/adapter/outbound/postgres"
	"github.com/JIeerioSst/rent-house-service/internal/adapter/outbound/recompenseservice"
	"github.com/JIeerioSst/rent-house-service/internal/adapter/outbound/userservice"
	"github.com/JIeerioSst/rent-house-service/internal/adapter/outbound/walletservice"
	"github.com/JIeerioSst/rent-house-service/internal/application/port/inbound"
	"github.com/JIeerioSst/rent-house-service/internal/application/port/outbound"
	"github.com/JIeerioSst/rent-house-service/internal/application/service"
)

var Module = fx.Options(
	fx.Provide(
		config.Load,
		newLogger,
		newPool,

		// outbound adapters -> outbound ports
		fx.Annotate(postgres.NewHomestayRepository, fx.As(new(outbound.HomestayRepository))),
		fx.Annotate(postgres.NewBookingRepository, fx.As(new(outbound.BookingRepository))),
		fx.Annotate(userservice.NewClient, fx.As(new(outbound.IdentityProvider))),
		newWalletGateway,
		newPaymentGateway,
		newLoyaltyGateway,
		newPaymentRail,
		newHostTiers,
		fx.Annotate(postgres.NewLeaseRepository, fx.As(new(outbound.LeaseRepository))),
		fx.Annotate(postgres.NewReviewRepository, fx.As(new(outbound.ReviewRepository))),
		fx.Annotate(postgres.NewWishlistRepository, fx.As(new(outbound.WishlistRepository))),

		// application services -> inbound ports
		fx.Annotate(newAuthService, fx.As(new(inbound.AuthUseCase))),
		fx.Annotate(service.NewHomestayService, fx.As(new(inbound.HomestayUseCase))),
		bookingOptions,
		leaseOptions,
		fx.Annotate(service.NewLeaseService, fx.As(new(inbound.LeaseUseCase))),
		fx.Annotate(service.NewReviewService, fx.As(new(inbound.ReviewUseCase))),
		fx.Annotate(service.NewWishlistService, fx.As(new(inbound.WishlistUseCase))),
		fx.Annotate(service.NewLoyaltyService, fx.As(new(inbound.LoyaltyUseCase))),
		fx.Annotate(service.NewBookingService, fx.As(new(inbound.BookingUseCase))),
		fx.Annotate(service.NewWalletService, fx.As(new(inbound.WalletUseCase))),

		// inbound adapter
		httpadapter.NewHandler,
		newPinger,
		httpadapter.NewRouter,
		newServer,
	),
	fx.Invoke(startServer, startHoldSweeper),
)

func newLogger() *slog.Logger { return slog.New(slog.NewJSONHandler(os.Stdout, nil)) }

func newPool(lc fx.Lifecycle, cfg config.Config) (*pgxpool.Pool, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	pool, err := postgres.NewPool(ctx, cfg)
	if err != nil {
		return nil, err
	}
	lc.Append(fx.Hook{OnStop: func(context.Context) error { pool.Close(); return nil }})
	return pool, nil
}

func newPinger(pool *pgxpool.Pool) httpadapter.Pinger { return pool.Ping }

func newWalletGateway(cfg config.Config) outbound.WalletGateway {
	if cfg.WalletServiceURL == "" {
		return nil
	}
	return walletservice.NewClient(cfg)
}

func newPaymentGateway(cfg config.Config) outbound.PaymentGateway {
	if cfg.PaymentServiceURL == "" {
		return nil
	}
	return paymentservice.NewClient(cfg)
}

func newLoyaltyGateway(cfg config.Config) outbound.LoyaltyGateway {
	if cfg.RecompenseServiceURL == "" {
		return nil
	}
	return recompenseservice.NewClient(cfg)
}

func newPaymentRail(cfg config.Config, w outbound.WalletGateway, g outbound.PaymentGateway, l outbound.LoyaltyGateway, log *slog.Logger) *service.PaymentRail {
	return service.NewPaymentRail(w, g, l, log)
}

func newHostTiers(cfg config.Config, l outbound.LoyaltyGateway) *service.HostTiers {
	return service.NewHostTiers(l, cfg.RecompenseCacheTTL)
}

func bookingOptions(cfg config.Config, log *slog.Logger) service.BookingOptions {
	return service.BookingOptions{HoldTTL: cfg.BookingHoldTTL, Log: log}
}

func leaseOptions(cfg config.Config, log *slog.Logger) service.LeaseOptions {
	return service.LeaseOptions{HoldTTL: cfg.BookingHoldTTL, Log: log}
}

func startHoldSweeper(lc fx.Lifecycle, cfg config.Config, bookings inbound.BookingUseCase, leases inbound.LeaseUseCase, log *slog.Logger) {
	ctx, cancel := context.WithCancel(context.Background())
	done := make(chan struct{})
	lc.Append(fx.Hook{
		OnStart: func(context.Context) error {
			go func() {
				defer close(done)
				t := time.NewTicker(cfg.SweepInterval)
				defer t.Stop()
				for {
					select {
					case <-ctx.Done():
						return
					case <-t.C:
						nb, err := bookings.ReleaseExpired(ctx)
						if err != nil {
							log.Error("release expired booking holds", "err", err)
						}
						nl, err := leases.ReleaseExpired(ctx)
						if err != nil {
							log.Error("release expired lease holds", "err", err)
						}
						if nb+nl > 0 {
							log.Info("released expired holds", "bookings", nb, "leases", nl)
						}
					}
				}
			}()
			return nil
		},
		OnStop: func(stopCtx context.Context) error {
			cancel()
			select {
			case <-done:
			case <-stopCtx.Done():
			}
			return nil
		},
	})
}

func newAuthService(cfg config.Config, identity outbound.IdentityProvider) *service.AuthService {
	return service.NewAuthService(identity, cfg.AdminRole)
}

func newServer(cfg config.Config, handler nethttp.Handler) *nethttp.Server {
	return &nethttp.Server{
		Addr:              net.JoinHostPort("", cfg.Port),
		Handler:           handler,
		ReadHeaderTimeout: 5 * time.Second,
		ReadTimeout:       15 * time.Second,
		WriteTimeout:      30 * time.Second,
		IdleTimeout:       60 * time.Second,
	}
}

func startServer(lc fx.Lifecycle, srv *nethttp.Server, log *slog.Logger) {
	lc.Append(fx.Hook{
		OnStart: func(context.Context) error {
			ln, err := net.Listen("tcp", srv.Addr)
			if err != nil {
				return err
			}
			log.Info("listening", "addr", srv.Addr)
			go func() {
				if err := srv.Serve(ln); err != nil && !errors.Is(err, nethttp.ErrServerClosed) {
					log.Error("server", "err", err)
				}
			}()
			return nil
		},
		OnStop: func(ctx context.Context) error { return srv.Shutdown(ctx) },
	})
}
