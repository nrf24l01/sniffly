package grpc

import (
	"context"
	"io"
	"log"
	"sync"
	"time"

	pb "github.com/nrf24l01/sniffly/capture_receiver/proto"
	"github.com/nrf24l01/sniffly/capturer/core"
	"github.com/nrf24l01/sniffly/capturer/snifpacket"
)

func StreamPackets(client pb.PacketGatewayClient, cfg *core.Config, packets chan *snifpacket.SnifPacket, wg *sync.WaitGroup) error {
	defer wg.Done()

	backoff := 1 * time.Second
	for {
		stream, err := client.StreamPackets(withAuth(context.Background(), cfg.ApiToken))
		if err != nil {
			log.Printf("failed to start packet stream: %v; retrying in %s", err, backoff)
			time.Sleep(backoff)
			if backoff < 30*time.Second {
				backoff *= 2
			}
			continue
		}

		backoff = 1 * time.Second

		go func() {
			for {
				_, err := stream.Recv()
				if err == io.EOF {
					log.Printf("grpc: server closed response stream")
					return
				}
				if err != nil {
					log.Printf("grpc: error receiving ack: %v", err)
					return
				}
			}
		}()

		for {
			pkt, ok := <-packets
			if !ok {
				_ = stream.CloseSend()
				log.Printf("StreamPackets: packets channel closed, exiting stream")
				return nil
			}

			protoPacket, err := pkt.ToProto()
			if err != nil {
				log.Printf("failed to convert packet to proto: %v", err)
				continue
			}

			if err := stream.Send(protoPacket); err != nil {
				log.Printf("failed to send packet to grpc stream: %v; will reconnect", err)
				_ = stream.CloseSend()
				break
			}
		}

		time.Sleep(500 * time.Millisecond)
	}
}