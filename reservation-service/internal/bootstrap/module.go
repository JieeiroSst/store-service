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

	"github.com/JIeeiroSSt/reservation-service/config"
	httpadapter "github.com/JIeeiroSSt/reservation-service/internal/adapter/inbound/http"
	"github.com/JIeeiroSSt/reservation-service/internal/adapter/outbound/notificationservice"
	"github.com/JIeeiroSSt/reservation-service/internal/adapter/outbound/paymentservice"
	"github.com/JIeeiroSSt/reservation-service/internal/adapter/outbound/postgres"
	"github.com/JIeeiroSSt/reservation-service/internal/adapter/outbound/recompenseservice"
	"github.com/JIeeiroSSt/reservation-service/internal/adapter/outbound/userservice"
	"github.com/JIeeiroSSt/reservation-service/internal/adapter/outbound/walletservice"
	"github.com/JIeeiroSSt/reservation-service/internal/application/port/inbound"
	"github.com/JIeeiroSSt/reservation-service/internal/application/port/outbound"
	"github.com/JIeeiroSSt/reservation-service/internal/application/service"
)

var Module = fx.Options(
	fx.Provide(
		config.Load,
		newLogger,
		newPool,

		// outbound adapters -> outbound ports
		fx.Annotate(postgres.NewHotelRepository, fx.As(new(outbound.HotelRepository))),
		fx.Annotate(postgres.NewReservationRepository, fx.As(new(outbound.ReservationRepository))),
		fx.Annotate(postgres.NewReviewRepository, fx.As(new(outbound.ReviewRepository))),
		fx.Annotate(postgres.NewNotificationRepository, fx.As(new(outbound.NotificationRepository))),
		fx.Annotate(postgres.NewWishlistRepository, fx.As(new(outbound.WishlistRepository))),
		fx.Annotate(postgres.NewWaitlistRepository, fx.As(new(outbound.WaitlistRepository))),
		newPushGateway,
		newNotifier,
		fx.Annotate(userservice.NewClient, fx.As(new(outbound.IdentityProvider))),
		newWalletGateway,
		newPaymentGateway,
		newLoyaltyGateway,
		newPaymentRail,
		newManagerTiers,

		// application services -> inbound ports
		fx.Annotate(newAuthService, fx.As(new(inbound.AuthUseCase))),
		fx.Annotate(newHotelService, fx.As(new(inbound.HotelUseCase))),
		fx.Annotate(service.NewNotificationService, fx.As(new(inbound.NotificationUseCase))),
		fx.Annotate(service.NewWishlistService, fx.As(new(inbound.WishlistUseCase))),
		fx.Annotate(newCatalogService, fx.As(new(inbound.CatalogUseCase))),
		waitlistOptions,
		service.NewWaitlistService,
		fx.Annotate(func(w *service.WaitlistService) *service.WaitlistService { return w }, fx.As(new(inbound.WaitlistUseCase))),
		newSoldOutCache,
		reservationOptions,
		fx.Annotate(service.NewReservationService, fx.As(new(inbound.ReservationUseCase))),
		fx.Annotate(service.NewReviewService, fx.As(new(inbound.ReviewUseCase))),
		fx.Annotate(service.NewLoyaltyService, fx.As(new(inbound.LoyaltyUseCase))),
		fx.Annotate(service.NewWalletService, fx.As(new(inbound.WalletUseCase))),

		// inbound adapter
		httpadapter.NewHandler,
		newPinger,
		newRouter,
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

func newPushGateway(cfg config.Config) outbound.PushGateway {
	if cfg.NotificationServiceURL == "" {
		return nil
	}
	return notificationservice.NewClient(cfg)
}

func newNotifier(repo outbound.NotificationRepository, log *slog.Logger) *service.Notifier {
	return service.NewNotifier(repo, log)
}

func newHotelService(repo outbound.HotelRepository, w outbound.WalletGateway, t *service.ManagerTiers, n *service.Notifier) *service.HotelService {
	return service.NewHotelService(repo, w, t).WithNotifier(n)
}

func newPaymentRail(w outbound.WalletGateway, g outbound.PaymentGateway, l outbound.LoyaltyGateway, log *slog.Logger) *service.PaymentRail {
	return service.NewPaymentRail(w, g, l, log)
}

func newManagerTiers(cfg config.Config, l outbound.LoyaltyGateway) *service.ManagerTiers {
	return service.NewManagerTiers(l, cfg.RecompenseCacheTTL)
}

func newSoldOutCache(cfg config.Config) *service.SoldOutCache {
	if cfg.SoldOutTTL <= 0 {
		return nil
	}
	return service.NewSoldOutCache(cfg.SoldOutTTL)
}

func newCatalogService(repo outbound.HotelRepository, soldOut *service.SoldOutCache, w *service.WaitlistService) *service.CatalogService {
	return service.NewCatalogService(repo, soldOut).WithPromoter(w)
}

func waitlistOptions(cfg config.Config, log *slog.Logger, soldOut *service.SoldOutCache, n *service.Notifier) service.WaitlistOptions {
	return service.WaitlistOptions{OfferTTL: cfg.HoldTTL, MaxPending: cfg.MaxPendingPerGuest, Log: log, Notifier: n, SoldOut: soldOut}
}

func reservationOptions(cfg config.Config, log *slog.Logger, soldOut *service.SoldOutCache, n *service.Notifier, w *service.WaitlistService) service.ReservationOptions {
	return service.ReservationOptions{
		HoldTTL: cfg.HoldTTL, Log: log,
		SoldOut:    soldOut,
		Bulkhead:   service.NewBulkhead(cfg.ReserveConcurrency, cfg.ReserveQueueWait),
		Limiter:    service.NewUserLimiter(cfg.UserRatePerSecond, cfg.UserBurst),
		MaxPending: cfg.MaxPendingPerGuest,
		Notifier:   n,
		Promoter:   w,
	}
}

func newRouter(cfg config.Config, h *httpadapter.Handler, ping httpadapter.Pinger, log *slog.Logger) nethttp.Handler {
	return httpadapter.NewRouter(h, ping, cfg.MaxInFlight, log)
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

func startHoldSweeper(lc fx.Lifecycle, cfg config.Config, reservations inbound.ReservationUseCase, notifications inbound.NotificationUseCase, waitlist inbound.WaitlistUseCase, log *slog.Logger) {
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
						if n, err := reservations.ReleaseExpired(ctx); err != nil {
							log.Error("release expired holds", "err", err)
						} else if n > 0 {
							log.Info("released expired holds", "reservations", n)
						}
						if n, err := reservations.SendReminders(ctx); err != nil {
							log.Error("send reminders", "err", err)
						} else if n > 0 {
							log.Info("sent check-in reminders", "reservations", n)
						}
						if n, err := waitlist.PromoteAll(ctx); err != nil {
							log.Error("waiting list round", "err", err)
						} else if n > 0 {
							log.Info("offered rooms from the waiting list", "offers", n)
						}
						if n, err := notifications.PushPending(ctx); err != nil {
							log.Error("push notifications", "err", err)
						} else if n > 0 {
							log.Info("pushed notifications", "count", n)
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
