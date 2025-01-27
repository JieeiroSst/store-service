package cmd

import (
	"context"
	"fmt"
	"log"
	"net"
	"net/http"

	pd "github.com/JIeeiroSst/coupon-service/gateway/proto"
	"github.com/JIeeiroSst/coupon-service/internal/config"
	serverHttp "github.com/JIeeiroSst/coupon-service/internal/delivery/http"
	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"google.golang.org/grpc"
)

var (
	ecosystem = ".env"
)

func runAPI() {
	config, _ := config.InitializeConfiguration(ecosystem)
	lis, err := net.Listen("tcp", fmt.Sprintf(":%v", config.Server.PortGrpcServer))
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}

	server := serverHttp.NewHandler()
	grpcServer := grpc.NewServer()
	pd.RegisterCouponServiceServer(grpcServer, server)

	go func() {
		log.Printf("Starting gRPC server on :%v", config.Server.PortGrpcServer)
		if err := grpcServer.Serve(lis); err != nil {
			log.Fatalf("failed to serve: %v", err)
		}
	}()

	ctx := context.Background()
	mux := runtime.NewServeMux()
	opts := []grpc.DialOption{grpc.WithInsecure()}
	err = pd.RegisterCouponServiceHandlerFromEndpoint(ctx, mux, fmt.Sprintf("localhost:%v", config.Server.PortGrpcServer), opts)
	if err != nil {
		log.Fatalf("Failed to register gateway: %v", err)
	}

	log.Printf("Starting HTTP server on :%v", config.Server.PortHttpServer)
	if err := http.ListenAndServe(fmt.Sprintf(":%v", config.Server.PortHttpServer), mux); err != nil {
		log.Fatalf("Failed to listen: %v", err)
	}
}
