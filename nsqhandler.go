package nsqhandler

import (
	"sync"

	"github.com/apex/log"
)

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
	return nil
}
