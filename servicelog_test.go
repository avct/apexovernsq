package apexovernsq

import (
	"fmt"
	"os"
	"testing"

	"github.com/apex/log"
	"github.com/apex/log/handlers/logfmt"
	"github.com/apex/log/handlers/memory"
)

const expectedServiceName string = "apexovernsq.test"

func TestProcessName(t *testing.T) {
	result := processName()
	if result != expectedServiceName {
		t.Fatalf("Expected %q, got %q", expectedServiceName, result)
	}
}

func TestNewServiceLogContext(t *testing.T) {
	handler := memory.New()
	log.SetHandler(handler)
	ctx := NewApexLogServiceContext()
	ctx.Info("Hello")

	entry := handler.Entries[0]
	if len(entry.Fields) != 3 {
		t.Fatalf("Expected 3 fields, got %d", len(entry.Fields))
	}

	serviceName := entry.Fields.Get("service")
	if serviceName == nil {
		t.Fatalf("No serviceName field in Entry")
	}

	if serviceName != expectedServiceName {
		t.Errorf("Expected %q, got %s", expectedServiceName, serviceName)
	}

	pid := entry.Fields.Get("pid")
	expectedPid := fmt.Sprintf("%d", os.Getpid())
	if pid != expectedPid {
		t.Errorf("Expected %s, got %s", expectedPid, pid)
	}

	hostname := entry.Fields.Get("hostname")
	expectedHostname, _ := os.Hostname()
	if hostname != expectedHostname {
		t.Errorf("Expected %q, got %q", expectedHostname, hostname)
	}
}

// If we create a new service log context from an existing context,
// we'll use it's Logger not the default one.
func TestNewServiceLogContextWithHandler(t *testing.T) {
	mem := memory.New()
	lf := logfmt.New(os.Stdout)
	log.SetHandler(lf)
	ctx2 := NewApexLogServiceContextWithHandler(mem)
	_, ok := ctx2.Logger.Handler.(*logfmt.Handler)
	if ok {
		t.Fatal("got logfmt handler expected memory handler")
	}

}

// The ServiceFilterApexLogHandler will let all entries through when the filter is nil.
func TestServiceFilterApexLogHandlerNoFilter(t *testing.T) {
	mem := memory.New()
	handler := NewApexLogServiceFilterHandler(mem, nil)
	log.SetHandler(handler)
	log.WithFields(log.Fields{"service": "test"}).Info("Test")
	resultCount := len(mem.Entries)
	if resultCount != 1 {
		t.Errorf("Expected %d entries, got %d", 1, resultCount)
	}
}
