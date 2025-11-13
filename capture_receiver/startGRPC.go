package main

import (
	"log"
	"net"

	"github.com/nrf24l01/sniffly/capture_receiver/core"
	"github.com/nrf24l01/sniffly/capture_receiver/handler"
	"github.com/nrf24l01/sniffly/capture_receiver/interceptors"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"

	pb "github.com/nrf24l01/sniffly/capture_receiver/proto"
)

func StartGRPCServer(cfg *core.AppConfig, packetGatewayServer *handler.PacketGatewayServer) {
	lis, err := net.Listen("tcp", cfg.CaptureConfig.AppHost)
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}

	unaryInt, streamInt := interceptors.NewAuthInterceptors(packetGatewayServer.DB)
	server := grpc.NewServer(
		grpc.UnaryInterceptor(unaryInt),
		grpc.StreamInterceptor(streamInt),
	)
	pb.RegisterPacketGatewayServer(server, packetGatewayServer)

	log.Printf("gRPC server listening on %s, reflection enabled: %v", cfg.CaptureConfig.AppHost, cfg.CaptureConfig.ReflectionEnabled)
	if cfg.CaptureConfig.ReflectionEnabled {
		reflection.Register(server)
	}
	if err := server.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}