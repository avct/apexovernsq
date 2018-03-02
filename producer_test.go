package apexovernsq

import (
	"encoding/json"
	"errors"
	"fmt"
	"sync"
	"testing"
	"time"

	"github.com/apex/log"
	"github.com/apex/log/handlers/memory"
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
	var mu sync.RWMutex
	fakePublish := func(topic string, body []byte) error {
		mu.Lock()
		messages = append(messages, &body)
		mu.Unlock()
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
			mu.RLock()
			messageCount = len(messages)
			mu.RUnlock()
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

func TestAsyncApexLogNSQHandlerBacksOff(t *testing.T) {
	backupHandler := memory.New()
	backupLogger = log.Logger{
		Handler: backupHandler,
		Level:   log.InfoLevel,
	}

	failyPublish := func(topic string, body []byte) error {
		return errors.New("oopsy")
	}

	handler := NewAsyncApexLogNSQHandler(json.Marshal, failyPublish, "testing", 5)
	log.SetHandler(handler)
	log.WithField("user", "tealeg").Info("Hello")
	log.WithField("user", "tealeg").Info("Hello")
	log.WithField("user", "tealeg").Info("Hello")
	timeout := time.After(time.Second * 5)
	<-timeout
	if len(backupHandler.Entries) != 6 {
		t.Errorf("Expected 6 backup-log messages, got %d", len(backupHandler.Entries))
	}
	expected := map[time.Duration]bool{
		0 * time.Second: false,
		1 * time.Second: false,
		2 * time.Second: false,
	}
	for _, entry := range backupHandler.Entries {
		inter, ok := entry.Fields["backoff"]
		if ok {
			backoff, _ := inter.(time.Duration)
			_, found := expected[backoff]
			if found {
				expected[backoff] = true
			}
		}
	}
	for backoff, found := range expected {
		if !found {
			t.Errorf("Expected to see a %s backoff, but it was not found", backoff)
		}
	}
}
