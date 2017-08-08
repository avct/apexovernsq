package nsqhandler

import (
	alog "github.com/apex/log"
	nsq "github.com/nsqio/go-nsq"
)

type UnmarshalFunc func(data []byte, v interface{}) error

type NSQApexLogHandler struct {
	handler       alog.Handler
	unmarshalFunc UnmarshalFunc
}

func NewNSQApexLogHandler(logger *alog.Logger, handler alog.Handler, unmarshalFunc UnmarshalFunc) *NSQApexLogHandler {
	return &NSQApexLogHandler{
		logger:        logger,
		handler:       handler,
		unmarshalFunc: unmarshalFunc,
	}
}

func (alh *NSQApexLogHandler) HandleMessage(m *nsq.Message) error {
	var entry *alog.Entry
	entry = alh.logger.WithField("remote", true)
	if err := alh.unmarshalFunc(m.Body, entry); err != nil {
		return err
	}
	return alh.handler.HandleLog(entry)
}
