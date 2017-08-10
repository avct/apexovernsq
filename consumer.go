package nsqhandler

import (
	alog "github.com/apex/log"
	nsq "github.com/nsqio/go-nsq"
)

// UnmarshalFunc is a function signature for any function that can
// unmarshal an arbitrary struct from a slice of bytes.  Whilst we
// only ever want to support github.com/apex/log.Entry structs, we
// support this interface because it allows using 3rd party
// Marshal/Unmarshal functions simply.
type UnmarshalFunc func(data []byte, v interface{}) error

// NSQApexLogHandler is a handler for NSQ that can consume messages
// who's Body is a marshalled github.com/apex/log.Entry.
type NSQApexLogHandler struct {
	logger        *alog.Logger
	handler       alog.Handler
	unmarshalFunc UnmarshalFunc
}

// NewNSQApexLogHandler creates a new NSQApexLogHandler with a
// provided github.com/apex/log.Handler and any function that satifies
// the UnmarshalFunc interface.
//
// The provided UnmarshalFunc will be used to unmarshal the
// github.com/apex/log.Entry from the NSQ Message.Body field.  It
// should match the marshal function used to publish the Message on
// the NSQ channel.  If you don't have any special requirement using
// the Marshal and Unmarshal functions from
// github.com/avct/nsqhandler/protobuf should work well.
//
// When the handler is invoked to consume a message, the provided
// github.com/apex/log.Handler will have it's HandleLog method called
// with the unmarshalled github.com/apex/log.Entry just as it would if
// you made a logging call locally.
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

// HandleMessage makes NSQApexLogHandler implement the
// github.com/nsqio/go-nsq.Handler interface and therefore,
// NSQApexLogHandler can be passed to the AddHandler function of a
// github.com/nsqio/go-nsq.Consumer.
//
// HandleMessage will unmarshal a github.com/apex/log.Entry from a
// github.com/nsqio/go-nsq.Message's Body and pass it into the
// github.com/apex/log.Handler provided when calling
// NewNSQApexLogHandler to construct the NSQApexLogHandler.
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
