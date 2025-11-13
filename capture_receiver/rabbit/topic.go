package rabbit

import "github.com/nrf24l01/go-web-utils/rabbitMQ"

type Topic struct {
	Name string
	created bool
}

func (t *Topic) CreateIfNotExists(rmq *rabbitMQ.RabbitMQ) error {
	if t.created {
		return nil
	}

	_, err := rmq.Channel.QueueDeclare(t.Name, true, false, false, false, nil)
	if err != nil {
		return err
	}

	t.created = true
	return nil
}