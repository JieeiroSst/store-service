// Package bootstrap wires the service together with fx: configuration, the database pool, the outbound adapters,
// the application services, the HTTP adapter and the background sweeper.
package bootstrap

import (
	"context"
	"errors"
	"log/slog"
	"net"
	nethttp "net/http"
	"os"
	"sync/atomic"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"go.temporal.io/sdk/client"
	"go.uber.org/fx"

	"github.com/JIeeiroSst/ticket-service/config"
	httpadapter "github.com/JIeeiroSst/ticket-service/internal/adapter/inbound/http"
	"github.com/JIeeiroSst/ticket-service/internal/adapter/inbound/temporalworker"
	"github.com/JIeeiroSst/ticket-service/internal/adapter/outbound/notificationservice"
	"github.com/JIeeiroSst/ticket-service/internal/adapter/outbound/paymentservice"
	"github.com/JIeeiroSst/ticket-service/internal/adapter/outbound/pdf"
	"github.com/JIeeiroSst/ticket-service/internal/adapter/outbound/postgres"
	"github.com/JIeeiroSst/ticket-service/internal/adapter/outbound/temporal"
	"github.com/JIeeiroSst/ticket-service/internal/adapter/outbound/uploadservice"
	"github.com/JIeeiroSst/ticket-service/internal/adapter/outbound/userservice"
	"github.com/JIeeiroSst/ticket-service/internal/adapter/outbound/walletservice"
	"github.com/JIeeiroSst/ticket-service/internal/application/port/inbound"
	"github.com/JIeeiroSst/ticket-service/internal/application/port/outbound"
	"github.com/JIeeiroSst/ticket-service/internal/application/service"
	"github.com/JIeeiroSst/ticket-service/internal/domain"
)

var Module = fx.Options(
	fx.Provide(
		config.Load,
		newLogger,
		newPool,

		// outbound adapters -> outbound ports
		fx.Annotate(postgres.NewEventRepository, fx.As(new(outbound.EventRepository))),
		fx.Annotate(postgres.NewOrderRepository, fx.As(new(outbound.OrderRepository))),
		fx.Annotate(postgres.NewTicketRepository, fx.As(new(outbound.TicketRepository))),
		fx.Annotate(postgres.NewStaffRepository, fx.As(new(outbound.StaffRepository))),
		fx.Annotate(postgres.NewWaitlistRepository, fx.As(new(outbound.WaitlistRepository))),
		fx.Annotate(postgres.NewResaleRepository, fx.As(new(outbound.ResaleRepository))),
		fx.Annotate(postgres.NewVenueRepository, fx.As(new(outbound.VenueRepository))),
		fx.Annotate(postgres.NewWishlistRepository, fx.As(new(outbound.WishlistRepository))),
		fx.Annotate(postgres.NewNotificationRepository, fx.As(new(outbound.NotificationRepository))),
		fx.Annotate(userservice.NewClient, fx.As(new(outbound.IdentityProvider)), fx.As(new(outbound.UserDirectory))),
		newWalletGateway,
		newPaymentGateway,
		newPushGateway,
		newLocation,
		newTemporalClient,
		newOrderLifecycle,
		newDocumentStore,
		fx.Annotate(pdf.NewRenderer, fx.As(new(outbound.DocumentRenderer))),
		fx.Annotate(postgres.NewDocumentRepository, fx.As(new(outbound.DocumentRepository))),
		fx.Annotate(newDocumentService, fx.As(new(inbound.DocumentUseCase))),
		newEmailGateway,

		// application services -> inbound ports
		newNotifier,
		newPaymentRail,
		newSoldOutCache,
		orderOptions,
		fx.Annotate(newAuthService, fx.As(new(inbound.AuthUseCase))),
		fx.Annotate(newEventService, fx.As(new(inbound.EventUseCase))),
		fx.Annotate(newCatalogService, fx.As(new(inbound.CatalogUseCase))),
		fx.Annotate(service.NewOrderService, fx.As(new(inbound.OrderUseCase))),
		fx.Annotate(service.NewGateService, fx.As(new(inbound.GateUseCase))),
		fx.Annotate(newTicketService, fx.As(new(inbound.TicketUseCase))),
		fx.Annotate(service.NewStaffService, fx.As(new(inbound.StaffUseCase))),
		fx.Annotate(newResaleService, fx.As(new(inbound.ResaleUseCase))),
		fx.Annotate(service.NewVenueService, fx.As(new(inbound.VenueUseCase))),
		newWaitlistService,
		fx.Annotate(func(w *service.WaitlistService) *service.WaitlistService { return w }, fx.As(new(inbound.WaitlistUseCase))),
		fx.Annotate(service.NewWishlistService, fx.As(new(inbound.WishlistUseCase))),
		fx.Annotate(service.NewNotificationService, fx.As(new(inbound.NotificationUseCase))),

		// inbound adapter
		httpadapter.NewHandler,
		newPinger,
		newRouter,
		newServer,
	),
	fx.Invoke(startServer, startSweeper, startTemporalWorker),
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

func newPushGateway(cfg config.Config) outbound.PushGateway {
	if cfg.NotificationServiceURL == "" {
		return nil
	}
	return notificationservice.NewClient(cfg)
}

func newEmailGateway(cfg config.Config) outbound.EmailGateway {
	if cfg.NotificationServiceURL == "" {
		return nil
	}
	return notificationservice.NewClient(cfg)
}

func newLocation(cfg config.Config, log *slog.Logger) *time.Location {
	loc, err := time.LoadLocation(cfg.TimeZone)
	if err != nil {
		log.Warn("unknown TimeZone, using UTC+7", "zone", cfg.TimeZone)
		loc = time.FixedZone("UTC+7", 7*3600)
	}
	return loc
}

func newNotifier(repo outbound.NotificationRepository, email outbound.EmailGateway, loc *time.Location, log *slog.Logger) *service.Notifier {
	return service.NewNotifier(repo, log).WithEmail(email != nil).WithLocation(loc)
}

func newDocumentStore(cfg config.Config) outbound.DocumentStore {
	if cfg.UploadServiceURL == "" {
		return nil
	}
	return uploadservice.NewClient(cfg)
}

func newDocumentService(orders inbound.OrderUseCase, repo outbound.OrderRepository, events outbound.EventRepository, tickets outbound.TicketRepository,
	docs outbound.DocumentRepository, store outbound.DocumentStore, render outbound.DocumentRenderer, loc *time.Location, log *slog.Logger) *service.DocumentService {
	return service.NewDocumentService(orders, repo, events, tickets, docs, store, render, loc, log)
}

func newPaymentRail(w outbound.WalletGateway, g outbound.PaymentGateway, log *slog.Logger) *service.PaymentRail {
	return service.NewPaymentRail(w, g, log)
}

func newTemporalClient(lc fx.Lifecycle, cfg config.Config) (client.Client, error) {
	if cfg.TemporalAddress == "" {
		return nil, nil
	}
	c, err := temporal.NewClient(cfg)
	if err != nil {
		return nil, err
	}
	lc.Append(fx.Hook{OnStop: func(context.Context) error { c.Close(); return nil }})
	return c, nil
}

func newOrderLifecycle(c client.Client, cfg config.Config) outbound.OrderLifecycle {
	if c == nil {
		return nil
	}
	return temporal.NewScheduler(c, cfg.TemporalTaskQueue)
}

func startTemporalWorker(lc fx.Lifecycle, cfg config.Config, c client.Client, orders inbound.OrderUseCase, documents inbound.DocumentUseCase, log *slog.Logger) {
	if c == nil {
		return
	}
	w := temporalworker.New(c, cfg.TemporalTaskQueue, &temporalworker.Activities{Orders: orders, Documents: documents})
	ctx, cancel := context.WithCancel(context.Background())
	done := make(chan struct{})
	var started atomic.Bool
	lc.Append(fx.Hook{
		OnStart: func(context.Context) error {
			go func() {
				defer close(done)
				startWithRetry(ctx, func() error {
					if err := w.Start(); err != nil {
						return err
					}
					started.Store(true)
					return nil
				}, 10*time.Second, log)
				if started.Load() {
					log.Info("order workflows on", "address", cfg.TemporalAddress, "namespace", cfg.TemporalNamespace, "queue", cfg.TemporalTaskQueue)
				}
			}()
			return nil
		},
		OnStop: func(context.Context) error {
			cancel()
			<-done
			if started.Load() {
				w.Stop()
			}
			return nil
		},
	})
}

func startWithRetry(ctx context.Context, start func() error, wait time.Duration, log *slog.Logger) {
	for {
		err := start()
		if err == nil {
			return
		}
		log.Warn("temporal worker not started, will retry", "err", err, "in", wait)
		select {
		case <-ctx.Done():
			return
		case <-time.After(wait):
		}
	}
}

func newSoldOutCache(cfg config.Config) *service.SoldOutCache {
	if cfg.SoldOutTTL <= 0 {
		return nil
	}
	return service.NewSoldOutCache(cfg.SoldOutTTL)
}

func orderOptions(cfg config.Config, log *slog.Logger, soldOut *service.SoldOutCache, n *service.Notifier, w *service.WaitlistService, lc outbound.OrderLifecycle) service.OrderOptions {
	return service.OrderOptions{
		HoldTTL:    cfg.HoldTTL,
		Log:        log,
		SoldOut:    soldOut,
		Bulkhead:   service.NewBulkhead(cfg.ReserveConcurrency, cfg.ReserveQueueWait),
		Limiter:    service.NewUserLimiter(cfg.UserRatePerSecond, cfg.UserBurst),
		MaxPending: cfg.MaxPendingPerUser,
		Notifier:   n,
		Waitlist:   w,
		Lifecycle:  lc,
		Invoice: service.InvoiceOptions{VATPercent: cfg.VATPercent,
			Seller: domain.Seller{Name: cfg.InvoiceSellerName, TaxID: cfg.InvoiceSellerTaxID, Address: cfg.InvoiceSellerAddress}},
	}
}

func newAuthService(cfg config.Config, identity outbound.IdentityProvider) *service.AuthService {
	return service.NewAuthService(identity, cfg.AdminRole)
}

func newEventService(cfg config.Config, events outbound.EventRepository, tickets outbound.TicketRepository, venues outbound.VenueRepository, orders outbound.OrderRepository, n *service.Notifier, w *service.WaitlistService) *service.EventService {
	return service.NewEventService(events, tickets, cfg.PublicCacheTTL).WithWaitlist(w).WithVenues(venues).WithNotices(orders, n)
}

func newTicketService(cfg config.Config, t outbound.TicketRepository, e outbound.EventRepository, st outbound.StaffRepository, dir outbound.UserDirectory, n *service.Notifier, log *slog.Logger) *service.TicketService {
	return service.NewTicketService(t, e, st, service.TicketOptions{TransferTTL: cfg.TransferTTL, MaxTransfers: cfg.MaxTransfers, Notifier: n, Log: log}).WithDirectory(dir)
}

func newResaleService(cfg config.Config, repo outbound.ResaleRepository, events outbound.EventRepository, rail *service.PaymentRail, n *service.Notifier, log *slog.Logger) *service.ResaleService {
	return service.NewResaleService(repo, events, rail, service.ResaleOptions{FeePercent: cfg.ResaleFeePercent, PlatformWalletID: cfg.PlatformWalletID,
		MaxTransfers: cfg.MaxTransfers, Notifier: n, Log: log})
}

func newWaitlistService(repo outbound.WaitlistRepository, events outbound.EventRepository, n *service.Notifier, log *slog.Logger) *service.WaitlistService {
	return service.NewWaitlistService(repo, events, n, log)
}

func newCatalogService(cfg config.Config, events outbound.EventRepository) *service.CatalogService {
	return service.NewCatalogService(events, cfg.PublicCacheTTL)
}

func newRouter(cfg config.Config, h *httpadapter.Handler, ping httpadapter.Pinger, log *slog.Logger) nethttp.Handler {
	return httpadapter.NewRouter(h, ping, cfg.MaxInFlight, cfg.InternalAPIKey, log)
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

func startSweeper(lc fx.Lifecycle, cfg config.Config, orders inbound.OrderUseCase, tickets inbound.TicketUseCase, resale inbound.ResaleUseCase, documents inbound.DocumentUseCase, notifications inbound.NotificationUseCase, log *slog.Logger) {
	ctx, cancel := context.WithCancel(context.Background())
	done := make(chan struct{})
	step := func(name string, fn func(context.Context) (int, error)) {
		if n, err := fn(ctx); err != nil {
			log.Error(name, "err", err)
		} else if n > 0 {
			log.Info(name, "count", n)
		}
	}
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
						step("released expired holds", orders.ReleaseExpired)
						step("refunded orders of cancelled events", orders.SettleCancelledEvents)
						step("closed unanswered ticket transfers", tickets.ExpireTransfers)
						step("expired unused tickets of ended events", tickets.ExpireTickets)
						step("settled or reopened resale listings", resale.Recover)
						step("made order documents", documents.GenerateMissing)
						step("sent event reminders", orders.SendReminders)
						step("pushed notifications", notifications.PushPending)
						step("sent e-mails", notifications.SendEmails)
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
