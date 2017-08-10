package nsqhandler

import (
	alog "github.com/apex/log"
	nsq "github.com/nsqio/go-nsq"
)

type UnmarshalFunc func(data []byte, v interface{}) error

type NSQApexLogHandler struct {
	logger        *alog.Logger
	handler       alog.Handler
	unmarshalFunc UnmarshalFunc
}

func NewNSQApexLogHandler(handler alog.Handler, unmarshalFunc UnmarshalFunc) *NSQApexLogHandler {
	if logger, ok := alog.Log.(*alog.Logger); ok {
		return &NSQApexLogHandler{
			logger:        logger,
			handler:       handler,
			unmarshalFunc: unmarshalFunc,
		}
	}
	panic("alog.Log is not an *alog.Logger")
}

func (alh *NSQApexLogHandler) HandleMessage(m *nsq.Message) error {
	entry := alog.NewEntry(alh.logger)
	if err := alh.unmarshalFunc(m.Body, entry); err != nil {
		return err
	}

	if entry.Level < alh.logger.Level {
		return nil
	}

	return alh.handler.HandleLog(entry)
}
