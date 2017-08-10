package nsqhandler

import (
	"testing"

	"github.com/avct/nsqhandler/protobuf"

	alog "github.com/apex/log"
	"github.com/apex/log/handlers/memory"
	nsq "github.com/nsqio/go-nsq"
)

func TestHandleMessage(t *testing.T) {
	var (
		entry *alog.Entry
	)

	fakeApexHandler := func(e *alog.Entry) error {
		entry = e
		return nil
	}

	memoryHandler := memory.New()
	alog.SetHandler(memoryHandler)
	alog.WithFields(alog.Fields{
		"flavour": "pistachio",
		"scoops":  "2",
	}).Info("it's ice cream time!")
	sourceEntry := memoryHandler.Entries[0]
	marshalledError, err := protobuf.Marshal(sourceEntry)
	if err != nil {
		t.Fatalf("Couldn't marshal log entry: %s", err.Error())
	}

	handler := NewNSQApexLogHandler(alog.HandlerFunc(fakeApexHandler), protobuf.Unmarshal)
	msg := nsq.NewMessage(nsq.MessageID{'a', 'b', 'c'}, []byte(marshalledError))

	handler.HandleMessage(msg)
	if entry.Message != sourceEntry.Message {
		t.Errorf("Expected %q, got %q", sourceEntry.Message, entry.Message)
	}
	if entry.Level != sourceEntry.Level {
		t.Errorf("Expected %s, got %s", sourceEntry.Level, entry.Level)
	}
	if entry.Timestamp != sourceEntry.Timestamp {
		t.Errorf("Expected %q, got %q", sourceEntry.Timestamp, entry.Timestamp)
	}
	expectedFieldCount := len(sourceEntry.Fields)
	actualFieldCount := len(entry.Fields)
	if expectedFieldCount != actualFieldCount {
		t.Errorf("Expected %d fields, but got %d fields", expectedFieldCount, actualFieldCount)
	}
	for fieldName, value := range sourceEntry.Fields {

	}
}
