package batcher

import (
	"log"
	"time"
)

func (b *Batcher) RecordBatch(duration time.Duration) (Batch, error) {
	msgs, err := b.RMQ.Channel.Consume(b.CFG.AppConfig.CapturePacketsTopic, "", false, false, false, false, nil)
	if err != nil {
		return Batch{}, err
	}

	var batch Batch

	deadline := time.After(duration)

	for {
		select {
		case msg := <-msgs:
			if err := batch.AddMessage(msg.Body); err != nil {
				log.Printf("Error getting message: %v", err)
			}
			_ = msg.Ack(false)
		case <-deadline:
			return batch, nil
		}
	}
}

func (b *Batcher) LoadAllRecords() (Batch, error) {
	msgs, err := b.RMQ.Channel.Consume(b.CFG.AppConfig.CapturePacketsTopic, "", false, false, false, false, nil)
	if err != nil {
		return Batch{}, err
	}

	var batch Batch

	for {
		select {
		case msg := <-msgs:
			if err := batch.AddMessage(msg.Body); err != nil {
				log.Printf("Error getting message: %v", err)
			}
			_ = msg.Ack(false)
		default:
			return batch, nil
		}
	}
}