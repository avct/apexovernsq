package nsqhandler

import (
	"encoding/json"
	"testing"

	"github.com/apex/log"
)

func TestNew(t *testing.T) {
	called := false
	fakePublish := func(topic string, body []byte) error {
		called = true
		return nil
	}
	fakeMarshal := func(x interface{}) ([]byte, error) {
		return nil, nil
	}
	handler := New(fakeMarshal, fakePublish, "testing")
	if handler == nil {
		t.Fatal("Expected *Handler, got nil")
	}
	if handler.publishFunc == nil {
		t.Fatal("Expected publishFunc to be set, but it was not")
	}
	handler.publishFunc("foo", nil)
	if !called {
		t.Fatal("Expected fakePublish to be called, but it was not.")
	}
	if handler.topic != "testing" {
		t.Fatalf("Expected topic to be \"testing\", but got %q",
			handler.topic)
	}
}

func TestHandler(t *testing.T) {
	var messages []*[]byte
	var loggedTopic string
	fakePublish := func(topic string, body []byte) error {
		messages = append(messages, &body)
		loggedTopic = topic
		return nil
	}

	log.SetHandler(New(json.Marshal, fakePublish, "testing"))
	log.WithField("user", "tealeg").Info("Hello")

	messageCount := len(messages)
	if messageCount != 1 {
		t.Fatalf("Expected 1 message to be logged, but found %d", messageCount)
	}
	message := messages[0]
	entry := &log.Entry{}
	err := json.Unmarshal(*message, entry)
	if err != nil {
		t.Fatalf("Error unmarshalling log message: %s", err.Error())
	}
	if entry.Message != "Hello" {
		t.Errorf("Expected unmarshalled message to be \"Hello\", got %q", entry.Message)
	}
	if entry.Level != log.InfoLevel {
		t.Error("Incorrect log level in unmarhalled entry")
	}
}
