package apexovernsq

import (
	"encoding/json"
	"fmt"
	"testing"

	"github.com/apex/log"
)

func TestNewApexLogNSQHandler(t *testing.T) {
	called := false
	fakePublish := func(topic string, body []byte) error {
		called = true
		return nil
	}
	fakeMarshal := func(x interface{}) ([]byte, error) {
		return nil, nil
	}
	handler := NewApexLogNSQHandler(fakeMarshal, fakePublish, "testing")
	if handler == nil {
		t.Fatal("Expected *ApexLogNSQHandler, got nil")
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

func TestApexLogNSQHandler(t *testing.T) {
	var messages []*[]byte
	var loggedTopic string
	fakePublish := func(topic string, body []byte) error {
		messages = append(messages, &body)
		loggedTopic = topic
		return nil
	}

	log.SetHandler(NewApexLogNSQHandler(json.Marshal, fakePublish, "testing"))
	log.WithField("user", "tealeg").Info("Hello")
	log.WithError(fmt.Errorf("Test Error")).Error("Oh dear!")

	messageCount := len(messages)
	if messageCount != 2 {
		t.Fatalf("Expected 2 message to be logged, but found %d", messageCount)
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
	user := entry.Fields.Get("user")
	if user == nil {
		t.Fatal("Expected value for field 'user', got nil")
	}
	if user != "tealeg" {
		t.Errorf("Expected user=\"tealeg\", got user=%q", user)
	}

	message = messages[1]
	err = json.Unmarshal(*message, entry)
	if err != nil {
		t.Fatalf("Error unmarshalling log message: %s", err.Error())
	}
	errorField := entry.Fields.Get("error")
	if errorField != "Test Error" {
		t.Errorf("Expected 'error' field to be \"Test Error\" got %q", errorField)
	}
}
