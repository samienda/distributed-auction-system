package kafka_services

import (
	"context"
	"fmt"

	"github.com/segmentio/kafka-go"
)

type KafkaProducer struct {
	writer *kafka.Writer
}

func NewKafkaProducer() *KafkaProducer {
	return &KafkaProducer{
		writer: &kafka.Writer{
			Addr:     kafka.TCP(brokerAddress),
			Topic:    topic,
			Balancer: &kafka.LeastBytes{},
		},
	}
}

func (p *KafkaProducer) PublishBid(auctionID, bidderID string, amount float64) error {
	message := fmt.Sprintf("AuctionID: %s, BidderID: %s, Amount: %.2f", auctionID, bidderID, amount)
	err := p.writer.WriteMessages(context.Background(), kafka.Message{
		Key:   []byte(auctionID),
		Value: []byte(message),
	})

	if err != nil {
		return fmt.Errorf("failed to publish bid: %w", err)
	}

	return nil
}