package nsqhandler

import (
	"sync"
	"time"

	"github.com/apex/log"
)

var start = time.Now()

type PublishFunc func(topic string, body []byte) error

type Handler struct {
	mu    sync.Mutex
	pfunc PublishFunc
	topic string
}

func New(pfunc PublishFunc, topic string) *Handler {
	return &Handler{
		pfunc: pfunc,
		topic: topic,
	}
}

func (h *Handler) HandleLog(e *log.Entry) error {
	h.mu.Lock()
	defer h.mu.Unlock()

	h.pfunc(h.topic, []byte(e.Message))

	return nil
}
