package nsqhandler

import (
	"sync"
	"time"

	"github.com/apex/log"
)

var start = time.Now()

type PublishFunc func(topic string, body []byte) error

type MarshalFunc func(x interface{}) ([]byte, error)

type Handler struct {
	mu          sync.Mutex
	marshalFunc MarshalFunc
	publishFunc PublishFunc
	topic       string
}

func New(marshalFunc MarshalFunc, publishFunc PublishFunc, topic string) *Handler {
	return &Handler{
		marshalFunc: marshalFunc,
		publishFunc: publishFunc,
		topic:       topic,
	}
}

func (h *Handler) HandleLog(e *log.Entry) error {
	h.mu.Lock()
	defer h.mu.Unlock()

	payload, err := h.marshalFunc(e)
	if err != nil {
		return err
	}
	return h.publishFunc(h.topic, payload)
}
