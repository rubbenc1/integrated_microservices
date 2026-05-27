package main

import (
	"gobr/internal/notification/config"

	"github.com/IBM/sarama"
)

func main() {
	cfg := config.LoadConfig()
	cfgKafka:=sarama.NewConfig()
	cfgKafka.Consumer.Return.Errors=true
	cfgKafka.Consumer.Offsets.Initial=sarama.OffsetNewest
	topic:="user_created"
	group:="notification-group"
	consumerGroup,err:=sarama.NewConsumerGroup([]string{cfg.KAFKA_BROKER},group,cfgKafka)
}