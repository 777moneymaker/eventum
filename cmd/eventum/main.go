package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"time"

	kafka "eventum/internal/kafkaclient"
	mongo "eventum/internal/mongoclient"
	redis "eventum/internal/redisclient"

	"github.com/joho/godotenv"
)

func init() {
	log.SetFlags(log.LstdFlags | log.Lshortfile)
	err := godotenv.Load()
	if err != nil {
		log.Fatalf("Failed to load .env file: %s", err)
	}
}

func main() {
	kafkaBroker := os.Getenv("KAFKA_BROKER")
	kafkaTopic := os.Getenv("KAFKA_TOPIC")
	redisAddr := os.Getenv("REDIS_ADDR")
	redisDB, err := strconv.Atoi(os.Getenv("REDIS_DB"))
	if err != nil {
		log.Fatalf("Could not get db for redis from env")
	}
	mongoURI := os.Getenv("MONGO_URI")
	mongoDBName := os.Getenv("MONGO_DB_NAME")
	mongoDBUser := os.Getenv("MONGO_USER")
	mongoDBPasswd := os.Getenv("MONGO_PASSWD")
	mongoFileCollectionName := os.Getenv("MONGO_FILE_COLLECTION_NAME")

	log.Printf("Loaded all env vars...")

	if kafkaBroker == "" || kafkaTopic == "" || redisAddr == "" || mongoURI == "" || mongoDBName == "" {
		log.Fatal("Environment variables KAFKA_BROKER, KAFKA_TOPIC, REDIS_ADDR, MONGO_URI, and MONGO_DB_NAME must be set")
	}

	log.Printf("Starting Mongo Client...")
	ctx, timeoutCancel := context.WithTimeout(context.Background(), time.Second*5)
	defer timeoutCancel()

	mongoClient, err := mongo.New(ctx, mongoURI, mongoDBName, mongoDBUser, mongoDBPasswd, mongoFileCollectionName)
	if err != nil {
		log.Fatal("Error connecting to MongoDB:", err)
	}
	defer mongoClient.Close(context.Background())

	log.Printf("Starting Redis Client...")

	redisClient := redis.New(context.Background(), redisAddr, redisDB)
	defer redisClient.Close()

	log.Printf("Starting Kafka Client...")
	kafkaClient, err := kafka.New(kafkaBroker, kafkaTopic)
	if err != nil {
		log.Fatalf("Failed to start kafka client: %s", err)
	}

	log.Printf("Consuming...")
	closeChan := kafkaClient.StartConsuming(context.Background(), redisClient, mongoClient)
	defer close(closeChan)

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	_, ok := <-sigChan
	if ok {
		log.Printf("Received shutdown signal, shutting down gracefully...")
	}

	log.Printf("Shutting down application")
}
