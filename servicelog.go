package apexovernsq

import (
	"errors"
	"fmt"
	"os"
	"path"
	"sort"
	"sync"

	"github.com/apex/log"
)

func processName() string {
	return path.Base(os.Args[0])
}

func appendServiceFieldsToEntry(entry *log.Entry) *log.Entry {
	hostname, err := os.Hostname()
	if err != nil {
		log.WithError(err).Warn("Unable to get hostname for service logging")
	}
	return entry.WithFields(
		log.Fields{
			"service":  processName(),
			"hostname": hostname,
			"pid":      fmt.Sprintf("%d", os.Getpid()),
		})

}

// Create a new logging context with service information.
func NewApexLogServiceContext() *log.Entry {
	logger, _ := log.Log.(*log.Logger)
	entry := log.NewEntry(logger)
	return appendServiceFieldsToEntry(entry)
}

// Create a new logging context, with a handler that isn't the default handler and with service information appended.
func NewApexLogServiceContextWithHandler(handler log.Handler, level log.Level) *log.Entry {
	logger := &log.Logger{
		Handler: handler,
		Level:   level,
	}
	entry := log.NewEntry(logger)
	return appendServiceFieldsToEntry(entry)
}

type ServiceFilterApexLogHandler struct {
	mu      sync.Mutex
	filter  *[]string
	handler log.Handler
}

func NewApexLogServiceFilterHandler(handler log.Handler, filter *[]string) *ServiceFilterApexLogHandler {
	if filter != nil {
		sort.Strings(*filter)
	}

	return &ServiceFilterApexLogHandler{
		handler: handler,
		filter:  filter,
	}
}

func (h *ServiceFilterApexLogHandler) shouldLog(e *log.Entry, serviceName string) bool {
	if h.filter == nil || len(*h.filter) == 0 || serviceName == "" {
		return true
	}
	index := sort.SearchStrings(*h.filter, serviceName)
	return index < len(*h.filter) && (*h.filter)[index] == serviceName
}

func (h *ServiceFilterApexLogHandler) HandleLog(e *log.Entry) error {
	h.mu.Lock()
	defer h.mu.Unlock()

	field := e.Fields.Get("service")
	serviceName, ok := field.(string)
	if ok {
		if h.shouldLog(e, serviceName) {
			return h.handler.HandleLog(e)
		}
		return nil
	}
	return errors.New("Entry had a service name that was not a string")
}
