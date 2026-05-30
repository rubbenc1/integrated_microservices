package main

import (
	"context"
	"encoding/json"
	"gobr/internal/notification/config"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/IBM/sarama"
)

type UserCreatedEvent struct {
	ID       string `json:"id"`
	Email    string `json:"email"`
	UserName string `json:"username"`
}

type handler struct{}

func (h *handler) Setup(sarama.ConsumerGroupSession) error   { return nil }
func (h *handler) Cleanup(sarama.ConsumerGroupSession) error { return nil }

func (h *handler) ConsumeClaim(session sarama.ConsumerGroupSession, claim sarama.ConsumerGroupClaim) error {
	for msg := range claim.Messages() {
		var event UserCreatedEvent
		if err := json.Unmarshal(msg.Value, &event); err != nil {
			log.Printf("Failed to unmarshal: %v", err)
			session.MarkMessage(msg, "")
			continue
		}
		log.Printf("New user: ID=%s, Email=%s, Username=%s", event.ID, event.Email, event.UserName)
		session.MarkMessage(msg, "")
	}
	return nil
}

func main() {
	cfg := config.LoadConfig()
	cfgKafka := sarama.NewConfig()
	cfgKafka.Consumer.Return.Errors = true
	cfgKafka.Consumer.Offsets.Initial = sarama.OffsetOldest
	consumerGroup, err := sarama.NewConsumerGroup([]string{cfg.KAFKA_BROKER}, "notification-group", cfgKafka)
	if err != nil {
		log.Fatalf("Failed to create consumer group: %v", err)
	}
	defer consumerGroup.Close()

	ctx, cancel := context.WithCancel(context.Background())
	go func() {
		for {
			if err := consumerGroup.Consume(ctx, []string{"user_created"}, &handler{}); err != nil {
				log.Printf("Consume error: %v", err)
				time.Sleep(2 * time.Second)  // ← add this
			}
			if ctx.Err() != nil {
				return
			}
		}
	}()
	go func() {
		for err := range consumerGroup.Errors() {
			log.Printf("Consumer group error: %v", err)
		}
	}()
	sigterm := make(chan os.Signal, 1)
	signal.Notify(sigterm, syscall.SIGINT, syscall.SIGTERM)
	<-sigterm
	log.Println("Shutting down...")
	cancel()
}
