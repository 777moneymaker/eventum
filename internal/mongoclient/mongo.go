package mongo

import (
	"context"
	"fmt"
	"log"

	"eventum/internal/event"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type MongoClient struct {
	Client          *mongo.Client
	Database        *mongo.Database
	FilesCollection *mongo.Collection
}

func New(ctx context.Context, uri, dbName, user, passwd, fileCollectionName string) (*MongoClient, error) {
	opts := options.Client().ApplyURI(uri)
	client, err := mongo.Connect(ctx, opts)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to MongoDB: %v", err)
	}

	if err := client.Ping(ctx, nil); err != nil {
		return nil, fmt.Errorf("failed to ping MongoDB: %v", err)
	}

	database := client.Database(dbName)

	if err := database.CreateCollection(ctx, fileCollectionName); err != nil {
		log.Fatalf("Could not create collection : %s", err)
	}

	fileCollection := database.Collection(fileCollectionName)

	indexModel := mongo.IndexModel{
		Keys: bson.D{
			{"file_name", 1},
			{"checksum", 1},
		},
		Options: options.Index().SetUnique(true),
	}
	_, err = fileCollection.Indexes().CreateOne(context.Background(), indexModel)
	if err != nil {
		log.Fatal(err)
	}

	return &MongoClient{
		Client:          client,
		Database:        database,
		FilesCollection: fileCollection,
	}, nil
}

func (m *MongoClient) ProcessEvent(ctx context.Context, e *event.Event) error {
	// Check if the event already exists
	filter := bson.M{"file_name": e.FileName, "checksum": e.Checksum}
	var existingFile event.File
	err := m.FilesCollection.FindOne(ctx, filter).Decode(&existingFile)
	if err != nil && err != mongo.ErrNoDocuments {
		return fmt.Errorf("failed to check if file exists: %v", err)
	}

	if err == mongo.ErrNoDocuments {
		newFile := event.File{
			FileName:  e.FileName,
			Checksum:  e.Checksum,
			UUID:      e.UUID,
			CreatedAt: e.EmitDateTime.UTC(),
			Events:    []event.Event{*e},
		}

		_, err := m.FilesCollection.InsertOne(ctx, newFile)
		if err != nil {
			return fmt.Errorf("failed to insert new file: %v", err)
		}

		log.Printf("New file added: %s, %s", e.FileName, e.Checksum)
		return nil
	}
	existingFile.Events = append(existingFile.Events, *e)
	update := bson.M{"$set": bson.M{
		"events": existingFile.Events,
	}}
	_, err = m.FilesCollection.UpdateOne(ctx, bson.M{"_id": existingFile.ID}, update)
	if err != nil {
		return fmt.Errorf("failed to update file with new event: %v", err)
	}

	log.Printf("Event added to existing file: %s, %s", e.FileName, e.Checksum)
	return nil
}

func (m *MongoClient) Close(ctx context.Context) error {
	return m.Client.Disconnect(ctx)
}
