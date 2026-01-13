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
		// Reduce ping frequency to avoid server ENHANCE_YOUR_CALM errors.
		// PermitWithoutStream=false prevents pings when there are no active RPCs.
		Time:                60 * time.Second,
		Timeout:             20 * time.Second,
		PermitWithoutStream: false,
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