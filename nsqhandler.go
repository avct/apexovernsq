package nsqhandler

import "sync"

type PublishFunc func(topic string, body []byte) error

type Handler struct {
	mu    sync.Mutex
	pfunc PublishFunc
}

func New(pfunc PublishFunc) *Handler {
	return &Handler{
		pfunc: pfunc,
	}
}
