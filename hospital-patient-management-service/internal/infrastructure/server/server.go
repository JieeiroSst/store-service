package server

import (
	"context"
	"errors"
	"net"
	"net/http"
	"runtime/debug"
	"time"

	pb "github.com/JIeeiroSst/lib-gateway/hospital-patientm-anagement-service/gateway/hospital-patientm-anagement-service"
	"github.com/JIeeroSst/hospital-patient-management-service/config"
	"github.com/JIeeroSst/hospital-patient-management-service/internal/adapter/primary/auth"
	"github.com/JIeeroSst/hospital-patient-management-service/internal/adapter/primary/grpcapi"
	"github.com/JIeeroSst/hospital-patient-management-service/internal/adapter/primary/httpapi"
	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"github.com/sirupsen/logrus"
	"go.uber.org/fx"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/reflection"
	"google.golang.org/grpc/status"
	"gorm.io/gorm"
)

type Params struct {
	fx.In

	Lifecycle   fx.Lifecycle
	Config      *config.Config
	DB          *gorm.DB
	Auth        *auth.Authenticator
	Hospital    *grpcapi.HospitalHandler
	Appointment *grpcapi.AppointmentHandler
	Identity    *httpapi.IdentityHandler
	Documents   *httpapi.DocumentHandler
}

func New(p Params) error {
	grpcSrv := grpc.NewServer(grpc.ChainUnaryInterceptor(recoverInterceptor, logInterceptor, p.Auth.Unary()))
	pb.RegisterHospotalServiceServer(grpcSrv, p.Hospital)
	pb.RegisterAppointmentServiceServer(grpcSrv, p.Appointment)
	reflection.Register(grpcSrv)

	mux := runtime.NewServeMux()
	ctx, cancel := context.WithCancel(context.Background())
	if err := pb.RegisterHospotalServiceHandlerServer(ctx, mux, p.Hospital); err != nil {
		cancel()
		return err
	}
	if err := pb.RegisterAppointmentServiceHandlerServer(ctx, mux, p.Appointment); err != nil {
		cancel()
		return err
	}

	root := http.NewServeMux()
	root.HandleFunc("GET /health", health(p.DB))
	p.Documents.Register(root)
	p.Identity.Register(root) 
	root.Handle("/", mux)
	httpSrv := &http.Server{
		Addr:              ":" + p.Config.Server.HTTPPort,
		Handler:           recoverHTTP(p.Auth.HTTP(root)),
		ReadHeaderTimeout: 10 * time.Second,
	}

	shutdownTimeout := p.Config.Server.ShutdownTimeout
	p.Lifecycle.Append(fx.Hook{
		OnStart: func(context.Context) error {
			grpcLn, err := net.Listen("tcp", ":"+p.Config.Server.GRPCPort)
			if err != nil {
				return err
			}
			httpLn, err := net.Listen("tcp", httpSrv.Addr)
			if err != nil {
				grpcLn.Close()
				return err
			}
			go func() {
				if err := grpcSrv.Serve(grpcLn); err != nil && !errors.Is(err, grpc.ErrServerStopped) {
					logrus.WithError(err).Error("grpc server stopped")
				}
			}()
			go func() {
				if err := httpSrv.Serve(httpLn); err != nil && !errors.Is(err, http.ErrServerClosed) {
					logrus.WithError(err).Error("http server stopped")
				}
			}()
			logrus.Infof("grpc listening on %s, http on %s", grpcLn.Addr(), httpLn.Addr())
			return nil
		},
		OnStop: func(stopCtx context.Context) error {
			cancel()
			ctx, done := context.WithTimeout(stopCtx, shutdownTimeout)
			defer done()

			stopped := make(chan struct{})
			go func() { grpcSrv.GracefulStop(); close(stopped) }()
			err := httpSrv.Shutdown(ctx)
			select {
			case <-stopped:
			case <-ctx.Done():
				grpcSrv.Stop()
			}
			return err
		},
	})
	return nil
}

func health(db *gorm.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx, cancel := context.WithTimeout(r.Context(), 2*time.Second)
		defer cancel()

		sqlDB, err := db.DB()
		if err == nil {
			err = sqlDB.PingContext(ctx)
		}
		w.Header().Set("Content-Type", "application/json")
		if err != nil {
			w.WriteHeader(http.StatusServiceUnavailable)
			_, _ = w.Write([]byte(`{"status":"unavailable"}`))
			return
		}
		_, _ = w.Write([]byte(`{"status":"ok"}`))
	}
}

func recoverHTTP(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if v := recover(); v != nil {
				logrus.Errorf("panic serving %s %s: %v\n%s", r.Method, r.URL.Path, v, debug.Stack())
				http.Error(w, "internal error", http.StatusInternalServerError)
			}
		}()
		next.ServeHTTP(w, r)
	})
}

func recoverInterceptor(ctx context.Context, req any, info *grpc.UnaryServerInfo, h grpc.UnaryHandler) (resp any, err error) {
	defer func() {
		if v := recover(); v != nil {
			logrus.Errorf("panic in %s: %v\n%s", info.FullMethod, v, debug.Stack())
			err = status.Error(codes.Internal, "internal error")
		}
	}()
	return h(ctx, req)
}

func logInterceptor(ctx context.Context, req any, info *grpc.UnaryServerInfo, h grpc.UnaryHandler) (any, error) {
	start := time.Now()
	resp, err := h(ctx, req)
	logrus.WithFields(logrus.Fields{
		"method":   info.FullMethod,
		"code":     status.Code(err).String(),
		"duration": time.Since(start).String(),
	}).Info("grpc request")
	return resp, err
}
