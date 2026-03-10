package handler

import (
	"context"
	"fmt"
	"io"
	"log"
	"time"

	"github.com/nrf24l01/sniffly/capture_receiver/postgres"
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

func (s *PacketGatewayServer) GetBlockRules(ctx context.Context, _ *pb.GetBlockRulesRequest) (*pb.GetBlockRulesResponse, error) {
	type row struct {
		ID          string
		DeviceID    string
		DeviceMAC   string
		DeviceIP    string
		DeviceLabel string
		TargetType  string
		TargetValue string
		Enabled     bool
	}

	rows := make([]row, 0)
	if err := s.DB.Raw(`
		SELECT
			dbr.id,
			dbr.device_id,
			di.mac AS device_mac,
			di.ip AS device_ip,
			di.label AS device_label,
			dbr.target_type,
			dbr.target_value,
			dbr.enabled
		FROM device_block_rules dbr
		JOIN device_info di ON di.id = dbr.device_id
		WHERE dbr.enabled = TRUE
		ORDER BY di.label ASC, di.mac ASC, dbr.created_at ASC
	`).Scan(&rows).Error; err != nil {
		return nil, fmt.Errorf("failed to load block rules: %w", err)
	}

	rules := make([]*pb.BlockRule, 0, len(rows))
	for _, row := range rows {
		target_type := pb.BlockRuleTargetType_BLOCK_RULE_TARGET_TYPE_UNSPECIFIED
		switch row.TargetType {
		case postgres.DeviceBlockRuleTargetIP:
			target_type = pb.BlockRuleTargetType_BLOCK_RULE_TARGET_TYPE_IP
		case postgres.DeviceBlockRuleTargetCIDR:
			target_type = pb.BlockRuleTargetType_BLOCK_RULE_TARGET_TYPE_CIDR
		case postgres.DeviceBlockRuleTargetSNI:
			target_type = pb.BlockRuleTargetType_BLOCK_RULE_TARGET_TYPE_SNI
		}

		rules = append(rules, &pb.BlockRule{
			Id:          row.ID,
			DeviceId:    row.DeviceID,
			DeviceMac:   row.DeviceMAC,
			DeviceIp:    row.DeviceIP,
			DeviceLabel: row.DeviceLabel,
			TargetType:  target_type,
			TargetValue: row.TargetValue,
			Enabled:     row.Enabled,
		})
	}

	return &pb.GetBlockRulesResponse{
		Rules:       rules,
		GeneratedAt: time.Now().UTC().Unix(),
	}, nil
}
