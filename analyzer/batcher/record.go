package batcher

import (
	"fmt"
	"log"
	"time"
)

func (b *Batcher) RecordBatch(duration time.Duration) (Batch, error) {
	consumerTag := fmt.Sprintf("batcher-record-%d", time.Now().UnixNano())
	msgs, err := b.RMQ.Channel.Consume(b.CFG.AppConfig.CapturePacketsTopic, consumerTag, false, false, false, false, nil)
	if err != nil {
		return Batch{}, err
	}

	defer func() {
		if err := b.RMQ.Channel.Cancel(consumerTag, false); err != nil {
			log.Printf("failed to cancel consumer %s: %v", consumerTag, err)
		}
	}()

	var batch Batch

	deadline := time.After(duration)

	for {
		select {
		case msg := <-msgs:
			if err := batch.AddMessage(msg.Body); err != nil {
				log.Printf("Error getting message: %v", err)
				if nackErr := msg.Nack(false, true); nackErr != nil {
					log.Printf("failed to Nack message: %v", nackErr)
				}
			} else {
				if ackErr := msg.Ack(false); ackErr != nil {
					log.Printf("failed to Ack message: %v", ackErr)
				}
			}
		case <-deadline:
			return batch, nil
		}
	}
}

func (b *Batcher) LoadAllRecords() (Batch, error) {
	consumerTag := fmt.Sprintf("batcher-loadall-%d", time.Now().UnixNano())
	msgs, err := b.RMQ.Channel.Consume(b.CFG.AppConfig.CapturePacketsTopic, consumerTag, false, false, false, false, nil)
	if err != nil {
		return Batch{}, fmt.Errorf("failed to start consume: %w", err)
	}

	defer func() {
		if err := b.RMQ.Channel.Cancel(consumerTag, false); err != nil {
			log.Printf("failed to cancel consumer %s: %v", consumerTag, err)
		}
	}()

	var batch Batch

	timeout := time.After(200 * time.Millisecond)
	for {
		if len(batch.Packets) >= 10000 {
			return batch, nil
		}
		select {
		case msg, ok := <-msgs:
			if !ok {
				return batch, nil
			}
			if err := batch.AddMessage(msg.Body); err != nil {
				log.Printf("Error getting message: %v", err)
				if nackErr := msg.Nack(false, true); nackErr != nil {
					log.Printf("failed to Nack message: %v", nackErr)
				}
			} else {
				if ackErr := msg.Ack(false); ackErr != nil {
					log.Printf("failed to Ack message: %v", ackErr)
				}
			}
			timeout = time.After(200 * time.Millisecond)
		case <-timeout:
			return batch, nil
		}
	}
}