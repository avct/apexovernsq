package apexovernsq

import (
	"encoding/json"
	"fmt"
	"strings"
	"testing"
	"time"

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
	fakePublish := func(topic string, body []byte) error {
		messages = append(messages, &body)
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

	callerFile := entry.Fields.Get("caller_file").(string)
	expected := "producer_test.go"
	if !strings.HasSuffix(callerFile, expected) {
		t.Fatalf("Expected caller file to be %q, but got %q", expected, callerFile)
	}

	callerLine := entry.Fields.Get("caller_line").(string)
	expected = "50" // Sorry, this test is going to break a lot ;-)
	if callerLine != expected {
		t.Fatalf("Expected caller line to be %q, but got %q", expected, callerLine)
	}

}

func TestNewAsyncApexLogHandler(t *testing.T) {
	fakePublish := func(topic string, body []byte) error {
		return nil
	}
	fakeMarshal := func(x interface{}) ([]byte, error) {
		return nil, nil
	}
	handler := NewAsyncApexLogNSQHandler(fakeMarshal, fakePublish, "testing", 2)
	if handler == nil {
		t.Fatal("Expected *AsyncApexLogNSQHandler, got nil")
	}
	if handler.logChan == nil {
		t.Fatal("Expected the establishment of a buffered channel for *log.Entry")
	}
	if handler.stopChan == nil {
		t.Fatal("Expected the establishment of a channel for stopping the go routine")
	}
	handler.Stop()
}

func TestAsyncApexLogHandlerSendsMessagesToBePublished(t *testing.T) {
	fakePublish := func(topic string, body []byte) error {
		return nil
	}
	fakeMarshal := func(x interface{}) ([]byte, error) {
		return nil, nil
	}
	handler := NewAsyncApexLogNSQHandler(fakeMarshal, fakePublish, "testing", 2)
	log.SetHandler(handler)
	handler.Stop() // Stop any messages getting consumed
	log.Info("Log something")
	entry := <-handler.logChan
	if entry == nil {
		t.Fatal("No log.Entry on channel")
	}
	if entry.Message != "Log something" {
		t.Fatal("Incorrect log.Entry on channel")
	}
}

func TestAsyncApexLogNSQHandlerLogs(t *testing.T) {
	var messages []*[]byte
	fakePublish := func(topic string, body []byte) error {
		messages = append(messages, &body)
		return nil
	}

	handler := NewAsyncApexLogNSQHandler(json.Marshal, fakePublish, "testing", 2)
	log.SetHandler(handler)
	log.WithField("user", "tealeg").Info("Hello")
	var messageCount int
	timeout := time.After(time.Second)
	ticker := time.NewTicker(10 * time.Millisecond)
	defer ticker.Stop()
Loop:
	for {
		select {
		case <-timeout:
			t.Fatal("Nothing logged within 1 second")
			break Loop
		case <-ticker.C:
			messageCount = len(messages)
			if messageCount > 0 {
				break Loop
			}
		}

	}

	if messageCount == 0 {
		t.Fatal("No messages logged")
	}

	handler.Stop()

}
