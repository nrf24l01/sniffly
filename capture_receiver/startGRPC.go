package main

import (
	"context"
	"log"
	"net"
	"strings"
	"time"

	"github.com/nrf24l01/sniffly/capture_receiver/core"
	"github.com/nrf24l01/sniffly/capture_receiver/handler"
	"github.com/nrf24l01/sniffly/capture_receiver/interceptors"
	"google.golang.org/grpc"
	"google.golang.org/grpc/health"
	healthpb "google.golang.org/grpc/health/grpc_health_v1"
	"google.golang.org/grpc/keepalive"
	"google.golang.org/grpc/reflection"

	pb "github.com/nrf24l01/sniffly/capture_receiver/proto"
)

func StartGRPCServer(cfg *core.AppConfig, packetGatewayServer *handler.PacketGatewayServer) {
	lis, err := net.Listen("tcp", cfg.CaptureConfig.AppHost)
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}

	unaryInt, streamInt := interceptors.NewAuthInterceptors(packetGatewayServer.DB)

	// Wrap interceptors to skip auth for health check/watch
	wrappedUnary := func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
		if strings.HasSuffix(info.FullMethod, "Health/Check") || strings.Contains(info.FullMethod, "ServerReflection") {
			return handler(ctx, req)
		}
		return unaryInt(ctx, req, info, handler)
	}
	wrappedStream := func(srv interface{}, ss grpc.ServerStream, info *grpc.StreamServerInfo, handler grpc.StreamHandler) error {
		if strings.HasSuffix(info.FullMethod, "Health/Watch") || strings.Contains(info.FullMethod, "ServerReflection") {
			return handler(srv, ss)
		}
		return streamInt(srv, ss, info, handler)
	}

	server := grpc.NewServer(
		grpc.UnaryInterceptor(wrappedUnary),
		grpc.StreamInterceptor(wrappedStream),
		grpc.KeepaliveParams(keepalive.ServerParameters{
			MaxConnectionIdle:     0,
			MaxConnectionAge:      0,
			MaxConnectionAgeGrace: 0,
			Time:                  15 * time.Second,
			Timeout:               30 * time.Second,
		}),
	)
	pb.RegisterPacketGatewayServer(server, packetGatewayServer)

	if cfg.CaptureConfig.PingEnabled {
		healthServer := health.NewServer()
		healthServer.SetServingStatus("", healthpb.HealthCheckResponse_SERVING)
		healthpb.RegisterHealthServer(server, healthServer)
	}

	log.Printf("gRPC server listening on %s, reflection enabled: %v", cfg.CaptureConfig.AppHost, cfg.CaptureConfig.ReflectionEnabled)
	if cfg.CaptureConfig.ReflectionEnabled {
		reflection.Register(server)
	}
	if err := server.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}