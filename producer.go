/*
Package apexovernsq provides a handler for github.com/apex/log.
It's intended to act as a transport to allow log.Entry structs to pass
through nsq and be reconstructed and passed to another handler on the
other side.
*/
package apexovernsq

import (
	"fmt"
	"math"
	"sync"
	"time"

	"github.com/apex/log"
	"github.com/apex/log/handlers/logfmt"
)

const maximumBackoffMultiple = 5

var (
	backupLogger = log.Logger{
		Handler: logfmt.Default,
		Level:   log.InfoLevel,
	}
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
	err = h.publishFunc(h.topic, payload)
	if err != nil {
		return err
	}
	return nil
}

// AsyncApexLogNSQHandler is a handler that can be passed to
// github.com/apex/log.SetHandler and will publish log entries on NSQ
// asynchronously.
type AsyncApexLogNSQHandler struct {
	mu       sync.Mutex
	wg       sync.WaitGroup
	logChan  chan *log.Entry
	stopChan chan bool
}

// NewAsyncApexLogNSQHandler returns a pointer to an
// apexovernsq.AsyncApexLogNSQHandler that can in turn be passed to
// github.com/apex/log.SetHandler.  The AsyncApexLogNSQHandler uses a
// goroutine and a channel to make the publication of NSQ message
// asynchronous to the act of logging.
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
func NewAsyncApexLogNSQHandler(marshalFunc MarshalFunc, publishFunc PublishFunc, topic string, bufferSize int) *AsyncApexLogNSQHandler {
	logChan := make(chan *log.Entry, bufferSize)
	stopChan := make(chan bool, 1)

	handler := &AsyncApexLogNSQHandler{
		logChan:  logChan,
		stopChan: stopChan,
	}

	// Form a closure over topic and publishFunc to keep the interface clean
	publishF := func(payload []byte) func() error {
		return func() error {
			return publishFunc(topic, payload)
		}
	}

	handler.wg.Add(1)
	go func(cLog chan *log.Entry, cStop chan bool) {
		var e *log.Entry
		for {
			select {
			case e = <-cLog:
				payload, err := marshalFunc(e)
				if err != nil {
					handler.mu.Lock()
					backupLogger.WithError(err).Error("cannot marshal log entry")
					handler.mu.Unlock()
					continue
				}
				err = publishOrRetry(
					time.Second*time.Duration(maximumBackoffMultiple),
					publishF(payload))
				if err != nil {
					handler.mu.Lock()
					backupLogger.WithError(err).Error("Publishing in AsyncApexLogNSQHander")
					handler.mu.Unlock()
					continue
				}
			case <-cStop:
				handler.wg.Done()
				return
			}

		}

	}(logChan, stopChan)
	return handler
}

func (h *AsyncApexLogNSQHandler) HandleLog(e *log.Entry) error {
	h.mu.Lock()
	defer h.mu.Unlock()

	select {
	case h.logChan <- e:
	default:
		backupLogger.Error("AsyncApexLogNSQHandler log channel is full")
		backupLogger.Handler.HandleLog(e)
	}

	return nil
}

func (h *AsyncApexLogNSQHandler) Stop() {
	h.stopChan <- true
	h.wg.Wait()
}

func publishOrRetry(maxBackoff time.Duration, fn func() error) error {
	var err error

	for i := 0; true; i++ {
		err = fn()
		if err == nil {
			break
		}
		backoff := time.Duration(math.Exp(float64(i))) * time.Second
		backupLogger.WithError(err).WithField("backoff", backoff).Info("failed to publish, backing off and retrying")
		if backoff > maxBackoff {
			err = fmt.Errorf("giving up after %v retries, too many errors. last error: %s", i, err)
			break
		}
		time.Sleep(backoff)
	}
	return err
}
