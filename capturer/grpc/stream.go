package grpc

import (
	"context"
	"log"
	"sync"

	pb "github.com/nrf24l01/sniffly/capture_receiver/proto"
	"github.com/nrf24l01/sniffly/capturer/core"
	"github.com/nrf24l01/sniffly/capturer/snifpacket"
)

func StreamPackets(client pb.PacketGatewayClient, cfg *core.Config, packets chan *snifpacket.SnifPacket, wg *sync.WaitGroup) error {
	defer wg.Done()
	
	stream, err := client.StreamPackets(withAuth(context.Background(), cfg.ApiToken))
	if err != nil {
		log.Fatalf("Failed to start packet stream: %v", err)
		return err
	}
	defer stream.CloseSend()

	for packet := range packets {
		// Transform SnifPacket to protobuf Packet
		protoPacket, err := packet.ToProto()
		if err != nil {
			return err
		}

		// Send the packet to the gRPC server
		err = stream.Send(protoPacket)
		if err != nil {
			return err
		}
	}

	return nil
}