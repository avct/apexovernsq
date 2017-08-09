package protobuf

import (
	"testing"

	alog "github.com/apex/log"
	"github.com/apex/log/handlers/memory"
)

func TestAttemptToMarshalUnsupportedType(t *testing.T) {
	_, err := Marshal("hello")
	if err == nil {
		t.Fatalf("Expected error, got nil")
	}
}

func TestMarshalAndUnmarshalEntry(t *testing.T) {
	handler := memory.New()
	alog.SetHandler(handler)
	alog.WithField("test", true).Warn("it's a test")
	entry := handler.Entries[0]
	marshalled, err := Marshal(entry)
	if err != nil {
		t.Fatalf("Error marshalling: %s", err.Error())
	}
	logEntry := &alog.Entry{}
	err = Unmarshal(marshalled, logEntry)
	if err != nil {
		t.Fatalf("Error unmarshalling: %s", err.Error())
	}
	if logEntry.Level != entry.Level {
		t.Fatalf("Failed to set Entry.Level")
	}
	if logEntry.Timestamp != entry.Timestamp {
		t.Fatalf("Failed to set Entry.Timestamp")
	}

}
