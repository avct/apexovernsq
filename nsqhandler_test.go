package nsqhandler

import (
	"testing"
)

func TestNew(t *testing.T) {
	fakePublish := func(topic string, body []byte) error {
		return nil
	}
	handler := New(fakePublish)
	if handler == nil {
		t.Fatal("Expected *Handler, got nil")
	}

}
