package protobuf

import (
	"fmt"
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

func TestAttemptToUnmarshallToUnsupportedType(t *testing.T) {
	type foo struct{}

	handler := memory.New()
	alog.SetHandler(handler)
	alog.WithField("test", "true").Warn("it's a test")
	entry := handler.Entries[0]
	marshalled, err := Marshal(entry)
	if err != nil {
		t.Fatalf("Error marshalling: %s", err.Error())
	}

	err = Unmarshal(marshalled, &foo{})
	if err == nil {
		t.Fatal("Expected an error, but nil was returned.")
	}
	if err.Error() != "Attempted to unmarshal to a type other than apex.log.Entry" {
		t.Error("Incorrect error message when attempting to Unmarshal an invalid type.")
	}
}

func TestMarshalAndUnmarshalEntry(t *testing.T) {
	handler := memory.New()
	alog.SetHandler(handler)
	alog.WithField("test", "true").Warn("it's a test")
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
		t.Error("Failed to set Entry.Level")
	}
	if logEntry.Timestamp != entry.Timestamp {
		t.Error("Failed to set Entry.Timestamp")
	}
	if logEntry.Message != entry.Message {
		t.Error("Failed to set Entry.Message")
	}
	if len(logEntry.Fields) != len(entry.Fields) {
		t.Error("Failed to set Entry.Fields")
	}
}

func TestMarshalAndUnmarshalWithError(t *testing.T) {
	handler := memory.New()
	alog.SetHandler(handler)
	originalErr := fmt.Errorf("oops")
	alog.WithError(originalErr).Error("it done broke")
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
	errMessage := logEntry.Fields.Get("error")
	if errMessage != originalErr.Error() {
		t.Errorf("Expected %q, got %q", originalErr.Error(), errMessage)
	}

}
