package handler

import (
	"github.com/nrf24l01/go-web-utils/rabbitMQ"
	"github.com/nrf24l01/sniffly/capture_receiver/core"
	pb "github.com/nrf24l01/sniffly/capture_receiver/proto"
	"github.com/nrf24l01/sniffly/capture_receiver/rabbit"
	"gorm.io/gorm"
)

type PacketGatewayServer struct {
	pb.UnimplementedPacketGatewayServer
	Config   *core.AppConfig
	DB       *gorm.DB
	RMQ      *rabbitMQ.RabbitMQ
	RMQTopic *rabbit.Topic
}