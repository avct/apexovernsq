/*
Package apexovernsq provides a handler for github.com/apex/log.
It's intended to act as a transport to allow log.Entry structs to pass
through nsq and be reconstructed and passed to another handler on the
other side.
*/
package apexovernsq

import (
	"sync"

	"github.com/apex/log"
)

// PublishFunc is a function signature for any function that publishes
// a message on a provided nsq topic.  Typically this is
// github.com/nsqio/go-nsq.Producer.Publish, or something that wraps
// it.
type PublishFunc func(topic string, body []byte) error

// MarshalFunc is a function signature for any function that can
// marshal an arbitrary struct to a slice of bytes.
type MarshalFunc func(x interface{}) ([]byte, error)

// ApexLogNSQHandler is a handler that can be passed to github.com/apex/log.SetHandler.
type ApexLogNSQHandler struct {
	mu          sync.Mutex
	marshalFunc MarshalFunc
	publishFunc PublishFunc
	topic       string
}

// NewApexLogNSQHandler returns a pointer to an apexovernsq.ApexLogNSQHandler that can
// in turn be passed to github.com/apex/log.SetHandler.
//
// The marshalFunc provided will be used to marshal a
// github.com/apex/log.Entry as the body of a message sent over nsq.
//
// The publishFunc is used to push a message onto the nsq.  For simple
// cases, with only one nsq endpoint using
// github.com/nsqio/go-nsq.Producer.Publish is fine.  For cases with
// multiple producers you'll want to wrap it.  See the examples
// directory for an implementation of this.
//
// The topic is a string determining the nsq topic the messages will
// be published to.
//
func NewApexLogNSQHandler(marshalFunc MarshalFunc, publishFunc PublishFunc, topic string) *ApexLogNSQHandler {
	return &ApexLogNSQHandler{
		marshalFunc: marshalFunc,
		publishFunc: publishFunc,
		topic:       topic,
	}
}

// HandleLog makes ApexLogNSQHandler fulfil the interface required by
// github.com/apex/log for handlers.  Each individual log entry made
// in client programs will eventually invoke this function when using
// this ApexLogNSQHandler.
func (h *ApexLogNSQHandler) HandleLog(e *log.Entry) error {
	h.mu.Lock()
	defer h.mu.Unlock()

	payload, err := h.marshalFunc(e)
	if err != nil {
		return err
	}
	return h.publishFunc(h.topic, payload)
}