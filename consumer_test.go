package nsqhandler

import (
	"encoding/json"
	"testing"

	alog "github.com/apex/log"
	"github.com/apex/log/handlers/memory"
	nsq "github.com/nsqio/go-nsq"
)

func TestNewNSQApexLogHandler(t *testing.T) {
	handler := NewNSQApexLogHandler(nil, nil)
	if handler == nil {
		t.Fatalf("Expected *NSQApexLogHandler, but got nil")
	}
}

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
		"scoops":  2,
	}).Info("it's ice cream time!")
	sourceEntry := memoryHandler.Entries[0]
	marshalledError, err := json.Marshal(sourceEntry)
	if err != nil {
		t.Fatalf("Couldn't marshal log entry: %s", err.Error())
	}

	handler := NewNSQApexLogHandler(alog.HandlerFunc(fakeApexHandler), json.Unmarshal)
	msg := nsq.NewMessage(nsq.MessageID{'a', 'b', 'c'}, []byte(marshalledError))

	handler.HandleMessage(msg)
	if entry.Message == "" {
		t.Fatalf("Handle message didn't call the apex log handler")
	}
}
