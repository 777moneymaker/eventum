package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"os"

	"eventum/internal/event"
	mongoclient "eventum/internal/mongoclient"

	"github.com/joho/godotenv"
	"go.mongodb.org/mongo-driver/bson"
)

func init() {
	log.SetFlags(log.LstdFlags | log.Lshortfile)
	err := godotenv.Load()
	if err != nil {
		log.Fatalf("Failed to load .env file: %s", err)
	}
}

func main() {
	checksum := flag.String("checksum", "", "Checksum of the file")
	filename := flag.String("filename", "", "Filename of the file")
	flag.Parse()

	mongoURI := os.Getenv("MONGO_URI")
	mongoDBName := os.Getenv("MONGO_DB_NAME")
	mongoDBUser := os.Getenv("MONGO_USER")
	mongoDBPasswd := os.Getenv("MONGO_PASSWD")
	mongoFileCollectionName := os.Getenv("MONGO_FILE_COLLECTION_NAME")

	mongoClient, err := mongoclient.New(context.Background(), mongoURI, mongoDBName, mongoDBUser, mongoDBPasswd, mongoFileCollectionName)
	if err != nil {
		log.Fatal("Error connecting to MongoDB:", err)
	}
	defer mongoClient.Close(context.Background())

	if *checksum == "" && *filename == "" {
		log.Fatal("Either checksum or filename must be provided")
	}

	collection := mongoClient.Client.Database(mongoDBName).Collection(mongoFileCollectionName)

	filter := bson.M{}
	if *filename != "" {
		filter["file_name"] = *filename
	}

	if *checksum != "" {
		filter["checksum"] = *checksum
	}

	// Fetch events from the collection
	ctx := context.Background()
	cursor, err := collection.Find(ctx, filter)
	if err != nil {
		log.Fatal("Error fetching events from MongoDB:", err)
	}
	defer cursor.Close(ctx)

	var files []event.File
	for cursor.Next(ctx) {
		var file event.File
		if err := cursor.Decode(&file); err != nil {
			log.Fatal("Error decoding event:", err)
		}
		files = append(files, file)
	}
	output, err := json.MarshalIndent(files, "", "  ")
	if err != nil {
		log.Fatalf("Could not display files: %s", err)
	}
	fmt.Printf("%s\n", string(output))

	if err := cursor.Err(); err != nil {
		log.Fatal("Error iterating over events:", err)
	}
}
