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
			log.Printf("failed to convert packet to proto: %v", err)
			return err
		}

		// Send the packet to the gRPC server
		err = stream.Send(protoPacket)
		if err != nil {
			log.Printf("failed to send packet to grpc stream: %v", err)
			return err
		}
	}

	log.Printf("StreamPackets: packets channel closed, exiting stream")
	return nil
}