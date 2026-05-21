package services

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/droskat/BH-Backend/models"
	"github.com/segmentio/kafka-go"
)

type KafkaPublisher struct {
	writer *kafka.Writer
}

func NewKafkaPublisher(writer *kafka.Writer) *KafkaPublisher {
	return &KafkaPublisher{writer: writer}
}

func (p *KafkaPublisher) PublishBatch(ctx context.Context, events []models.ImageEvent) error {
	messages := make([]kafka.Message, 0, len(events))

	for _, event := range events {
		payload, err := json.Marshal(event)
		if err != nil {
			return fmt.Errorf("kafka marshal event: %w", err)
		}
		messages = append(messages, kafka.Message{
			Key:   []byte(event.ImageID.String()),
			Value: payload,
		})
	}

	if err := p.writer.WriteMessages(ctx, messages...); err != nil {
		return fmt.Errorf("kafka publish batch: %w", err)
	}
	return nil
}

func (p *KafkaPublisher) Close() error {
	return p.writer.Close()
}
