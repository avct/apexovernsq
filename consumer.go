package nsqhandler

import nsq "github.com/nsqio/go-nsq"

type ApexLogHandler struct {
}

func (alh *ApexLogHandler) HandleMessage(m *nsq.Message) error {
	return nil
}
