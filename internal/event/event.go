package event

import (
	"fmt"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// EventType represents the type of the event (INFO, WARNING, ERROR, CRITICAL)
type EventType string

const (
	INFO     EventType = "INFO"
	WARNING  EventType = "WARNING"
	ERROR    EventType = "ERROR"
	CRITICAL EventType = "CRITICAL"
)

// Event represents an event that can be processed, stored in Redis temporarily, and persisted in MongoDB.
type Event struct {
	ID           primitive.ObjectID `json:"_id,omitempty" bson:"_id,omitempty"` // MongoDB ID (optional)
	EventName    string             `json:"event_name" bson:"event_name"`
	Description  string             `json:"description" bson:"description"`
	FileName     string             `json:"file_name" bson:"file_name"`
	Checksum     string             `json:"checksum" bson:"checksum"`
	EventType    EventType          `json:"event_type" bson:"event_type"`
	EmitDateTime time.Time          `json:"emit_datetime" bson:"emit_datetime"`
	SaveDateTime time.Time          `json:"save_datetime" bson:"save_datetime"`
	UUID         string             `json:"uuid" bson:"uuid"`
}

// File represents a collection of events associated with a specific file (file name and checksum)
type File struct {
	ID        primitive.ObjectID `json:"_id,omitempty" bson:"_id,omitempty"` // MongoDB ID (optional)
	FileName  string             `json:"file_name" bson:"file_name"`
	Checksum  string             `json:"checksum" bson:"checksum"`
	UUID      string             `json:"uuid" bson:"uuid"`
	Events    []Event            `json:"events" bson:"events"`
	CreatedAt time.Time          `json:"created_at" bson:"created_at"`
}

// MarshalBSON implements custom marshalling for the Event struct to handle time formats for BSON.
func (e *Event) MarshalBSON() ([]byte, error) {
	type Alias Event
	return bson.Marshal(&struct {
		EmitDateTime string `bson:"emit_datetime"`
		*Alias
	}{
		EmitDateTime: e.EmitDateTime.Format(time.RFC3339),
		Alias:        (*Alias)(e),
	})
}

// UnmarshalBSON implements custom unmarshalling for the Event struct to handle time formats for BSON.
func (e *Event) UnmarshalBSON(data []byte) error {
	type Alias Event
	aux := &struct {
		EmitDateTime string `bson:"emit_datetime"`
		*Alias
	}{
		Alias: (*Alias)(e),
	}
	if err := bson.Unmarshal(data, &aux); err != nil {
		return err
	}

	// Parse EmitDateTime from string (e.g., "2024-12-30T12:34:56Z")
	parsedTime, err := time.Parse(time.RFC3339, aux.EmitDateTime)
	if err != nil {
		return fmt.Errorf("error parsing emit_datetime: %v", err)
	}
	e.EmitDateTime = parsedTime
	e.SaveDateTime = time.Now().UTC() // Default to the current UTC time when saved
	return nil
}
