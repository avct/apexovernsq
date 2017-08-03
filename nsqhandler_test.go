package nsqhandler

import (
	"testing"

	"github.com/apex/log"
)

func TestNew(t *testing.T) {
	called := false
	fakePublish := func(topic string, body []byte) error {
		called = true
		return nil
	}
	handler := New(fakePublish, "testing")
	if handler == nil {
		t.Fatal("Expected *Handler, got nil")
	}
	if handler.pfunc == nil {
		t.Fatal("Expected pfunc to be set, but it was not")
	}
	handler.pfunc("foo", nil)
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

	log.SetHandler(New(fakePublish, "testing"))
	log.WithField("user", "tealeg").Info("Hello")

	messageCount := len(messages)
	if messageCount != 1 {
		t.Fatal("Expected 1 message to be logged, but found %d", messageCount)
	}
}
