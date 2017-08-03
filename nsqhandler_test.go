package nsqhandler

import (
	"testing"
)

func TestNew(t *testing.T) {
	called := false
	fakePublish := func(topic string, body []byte) error {
		called = true
		return nil
	}
	handler := New(fakePublish)
	if handler == nil {
		t.Fatal("Expected *Handler, got nil")
	}
	if handler.pfunc == nil {
		t.Fatal("Expected pfunc to be set, but it was not")
	}
	handler.pfunc("foo", nil)
	if !called {
		t.Fatal("Expect fakePublish to be called, but it was not.")
	}
}
