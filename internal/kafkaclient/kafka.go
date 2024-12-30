package kafka

import (
	"context"
	"encoding/json"
	"log"

	"eventum/internal/event"
	mongo "eventum/internal/mongoclient"
	redis "eventum/internal/redisclient"

	"github.com/confluentinc/confluent-kafka-go/kafka"
)

type KafkaClient struct {
	Reader *kafka.Consumer
}

func New(broker, topic string) (*KafkaClient, error) {
	c, err := kafka.NewConsumer(&kafka.ConfigMap{
		"bootstrap.servers": broker,
		"group.id":          "eventum-group",
		"auto.offset.reset": "earliest",
	})
	if err != nil {
		log.Fatalf("Failed to create kafka consumer %v", err)
	}
	c.SubscribeTopics([]string{topic}, nil)

	return &KafkaClient{
		Reader: c,
	}, nil
}

func (k *KafkaClient) StartConsuming(ctx context.Context, redisClient *redis.RedisClient, mongoClient *mongo.MongoClient) chan struct{} {
	stop := make(chan struct{})

	go func() {
		defer close(stop)
		for {

			message, err := k.Reader.ReadMessage(-1)
			if err != nil {
				log.Printf("Error reading Kafka message: %v", err)
				return
			}

			log.Printf("Got message: %s", message.Value)
			var e event.Event

			if err := json.Unmarshal(message.Value, &e); err != nil {
				log.Printf("Error unmarshalling Kafka message %s: %s", message.Value, err)
				continue
			}

			key := e.Checksum + e.FileName + e.UUID
			if err := redisClient.SaveEvent(key, &e); err != nil {
				log.Printf("Error saving event to Redis: %v", err)
				continue
			}

			if err := mongoClient.ProcessEvent(ctx, &e); err != nil {
				log.Printf("Error processing event: %v", err)
				continue
			}

			if err := redisClient.DeleteEvent(key); err != nil {
				log.Printf("Error deleting event from Redis: %v", err)
			}

			log.Printf("Successfully processed event: %v", e.EventName)
		}
	}()

	return stop
}
