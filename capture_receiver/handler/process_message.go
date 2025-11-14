package handler

import (
	"context"
	"fmt"
	"io"
	"log"

	pb "github.com/nrf24l01/sniffly/capture_receiver/proto"
	"github.com/nrf24l01/sniffly/capture_receiver/rabbit"
)

func (s *PacketGatewayServer) PublishPacket(ctx context.Context, pkt *pb.Packet) (*pb.PublishResponse, error) {
	msg := rabbit.NewMessage(pkt.Payload, pkt.Timestamp, pkt.SourceId, s.RMQTopic)
	if err := msg.ToRabbitMQMessage(s.RMQ, ctx, false); err != nil {
		return nil, fmt.Errorf("failed to publish message to RabbitMQ: %w", err)
	}
	log.Printf("[Unary] Packet from %s saved.", pkt.SourceId)
	return &pb.PublishResponse{Success: true, MessageId: "uuid-1234"}, nil
}

func (s *PacketGatewayServer) StreamPackets(stream pb.PacketGateway_StreamPacketsServer) error {
	ctx := stream.Context()
	log.Printf("[Stream] Started receiving packets...")
	for {
		pkt, err := stream.Recv()
		if err == io.EOF {
			log.Println("[Stream] End of stream")
			return nil
		}
		if err != nil {
			return err
		}
		msg := rabbit.NewMessage(pkt.Payload, pkt.Timestamp, pkt.SourceId, s.RMQTopic)
		if err := msg.ToRabbitMQMessage(s.RMQ, ctx, false); err != nil {
			return fmt.Errorf("failed to publish message to RabbitMQ: %w", err)
		}

		resp := &pb.PublishResponse{Success: true, MessageId: fmt.Sprintf("stream-%d", pkt.Timestamp)}
		if err := stream.Send(resp); err != nil {
			return err
		}
	}
}