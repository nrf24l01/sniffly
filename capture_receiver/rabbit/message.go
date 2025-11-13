package rabbit

import (
	"context"
	"encoding/json"

	"github.com/nrf24l01/go-web-utils/rabbitMQ"
	amqp "github.com/rabbitmq/amqp091-go"
)

type Message struct {
	Payload []byte     `json:"payload"`
	Timestamp int64    `json:"timestamp"`
	SenderUUID string  `json:"sender_uuid"`
	topic *Topic       `json:"-"`
}

func NewMessage(payload []byte, timestamp int64, senderUUID string, topic_name string) *Message {
	topic := &Topic{
		Name: topic_name,
		created: false,
	}

	return &Message{
		Payload:    payload,
		Timestamp:  timestamp,
		SenderUUID: senderUUID,
		topic:     topic,
	}
}

func (m *Message) ToRabbitMQMessage(rmq *rabbitMQ.RabbitMQ, ctx context.Context) error {
	if err := m.topic.CreateIfNotExists(rmq); err != nil {
		return err
	}

	data, err := json.Marshal(m)
	if err != nil {
		return err
	}

	pub := amqp.Publishing{
		ContentType: "application/json",
		Body:        data,
	}

	if err = rmq.Channel.PublishWithContext(ctx, "", m.topic.Name, false, false, pub); err != nil {
		return err
	}

	return nil
}