package connectors

import (
	"github.com/droskat/BH-Backend/config"
	"github.com/segmentio/kafka-go"
)

func NewKafkaWriter(cfg config.KafkaConfig) *kafka.Writer {
	return &kafka.Writer{
		Addr:         kafka.TCP(cfg.Brokers...),
		Topic:        cfg.Topic,
		Balancer:     &kafka.LeastBytes{},
		RequiredAcks: kafka.RequireAll,
		Async:        false,
		BatchSize:    100,
		MaxAttempts:  5,
	}
}

func NewKafkaReader(cfg config.KafkaConfig, groupID string) *kafka.Reader {
	return kafka.NewReader(kafka.ReaderConfig{
		Brokers:  cfg.Brokers,
		Topic:    cfg.Topic,
		GroupID:  groupID,
		MinBytes: 1e3,
		MaxBytes: 10e6,
	})
}
