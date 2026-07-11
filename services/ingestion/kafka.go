package main

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/segmentio/kafka-go"
)

// Publisher wraps one kafka-go Writer per topic. In a service this small
// that's fine; if this grows to many topics, switch to a single Writer
// with per-message Topic set on kafka.Message instead.
type Publisher struct {
	priceWriter   *kafka.Writer
	weatherWriter *kafka.Writer
}

func NewPublisher(brokers []string) *Publisher {
	newWriter := func(topic string) *kafka.Writer {
		return &kafka.Writer{
			Addr:         kafka.TCP(brokers...),
			Topic:        topic,
			Balancer:     &kafka.LeastBytes{},
			RequiredAcks: kafka.RequireAll, // wait for all in-sync replicas — favor durability over latency here
			BatchTimeout: 500 * time.Millisecond,
		}
	}
	return &Publisher{
		priceWriter:   newWriter("price.raw"),
		weatherWriter: newWriter("weather.raw"),
	}
}

func (p *Publisher) PublishPrices(ctx context.Context, records []PriceRecord) error {
	msgs := make([]kafka.Message, 0, len(records))
	for _, r := range records {
		payload, err := json.Marshal(r)
		if err != nil {
			return fmt.Errorf("marshaling price record: %w", err)
		}
		// Key by mandi+commodity so Kafka partitions keep one mandi's
		// history in order relative to itself — important once you have
		// more than one partition and care about per-key ordering.
		key := fmt.Sprintf("%s|%s", r.Market, r.Commodity)
		msgs = append(msgs, kafka.Message{Key: []byte(key), Value: payload})
	}
	if len(msgs) == 0 {
		return nil
	}
	return p.priceWriter.WriteMessages(ctx, msgs...)
}

func (p *Publisher) PublishWeather(ctx context.Context, records []WeatherRecord) error {
	msgs := make([]kafka.Message, 0, len(records))
	for _, r := range records {
		payload, err := json.Marshal(r)
		if err != nil {
			return fmt.Errorf("marshaling weather record: %w", err)
		}
		msgs = append(msgs, kafka.Message{Key: []byte(r.District), Value: payload})
	}
	if len(msgs) == 0 {
		return nil
	}
	return p.weatherWriter.WriteMessages(ctx, msgs...)
}

func (p *Publisher) Close() error {
	if err := p.priceWriter.Close(); err != nil {
		return err
	}
	return p.weatherWriter.Close()
}
