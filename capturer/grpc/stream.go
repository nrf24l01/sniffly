package grpc

import (
	"context"
	"log"
	"sync"
	"time"

	pb "github.com/nrf24l01/sniffly/capture_receiver/proto"
	"github.com/nrf24l01/sniffly/capturer/core"
	"github.com/nrf24l01/sniffly/capturer/snifpacket"
)

func StreamPackets(client pb.PacketGatewayClient, cfg *core.Config, packets chan *snifpacket.SnifPacket, wg *sync.WaitGroup) error {
	defer wg.Done()

	// Keep trying to maintain a stream to the server. On transient errors
	// reconnect with backoff instead of exiting — this prevents the client
	// from silently stopping consumption of `packets`.
	backoff := 1 * time.Second
	for {
		stream, err := client.StreamPackets(withAuth(context.Background(), cfg.ApiToken))
		if err != nil {
			log.Fatalf("failed to start packet stream: %v; retrying in %s", err, backoff)
			time.Sleep(backoff)
			if backoff < 30*time.Second {
				backoff *= 2
			}
			continue
		}

		// Reset backoff on successful connection
		backoff = 1 * time.Second

		// Read from the packets channel and send to the stream. If send fails
		// try to reconnect; do not return immediately so transient network
		// issues don't stop packet processing.
		for {
			pkt, ok := <-packets
			if !ok {
				// Sender closed the channel — clean up and exit normally.
				_ = stream.CloseSend()
				log.Fatalf("StreamPackets: packets channel closed, exiting stream")
				return nil
			}

			protoPacket, err := pkt.ToProto()
			if err != nil {
				log.Printf("failed to convert packet to proto: %v", err)
				// skip this packet but keep streaming
				continue
			}

			if err := stream.Send(protoPacket); err != nil {
				log.Fatalf("failed to send packet to grpc stream: %v; will reconnect", err)
				// close this stream and break to outer loop to reconnect
				_ = stream.CloseSend()
				break
			}
		}

		// small sleep before reconnecting to avoid busy reconnect loops
		time.Sleep(500 * time.Millisecond)
	}
}