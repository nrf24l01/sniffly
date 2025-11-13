package main

import (
	"log"
	"net"

	"github.com/nrf24l01/sniffly/capture_receiver/core"
	"github.com/nrf24l01/sniffly/capture_receiver/handler"
	"google.golang.org/grpc"

	pb "github.com/nrf24l01/sniffly/capture_receiver/proto"
)

func StartGRPCServer(cfg *core.AppConfig, packetGatewayServer *handler.PacketGatewayServer) {
	lis, err := net.Listen("tcp", ":50051")
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}

	server := grpc.NewServer()
	pb.RegisterPacketGatewayServer(server, packetGatewayServer)

	log.Println("gRPC server listening on :50051")
	if err := server.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}