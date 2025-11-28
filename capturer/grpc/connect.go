package grpc

import (
	"time"

	pb "github.com/nrf24l01/sniffly/capture_receiver/proto"
	"github.com/nrf24l01/sniffly/capturer/core"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/keepalive"
)

func ConnectPacketGatewayClient(cfg *core.Config) (pb.PacketGatewayClient, error) {
	kp := keepalive.ClientParameters{
		Time:                10 * time.Second,
		Timeout:             20 * time.Second,
		PermitWithoutStream: true,
	}

	conn, err := grpc.NewClient(cfg.ServerAddress,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithKeepaliveParams(kp),
	)
	if err != nil {
		return nil, err
	}

	client := pb.NewPacketGatewayClient(conn)
	return client, nil
}