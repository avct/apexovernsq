package protobuf

import (
	"fmt"

	alog "github.com/apex/log"
	proto "github.com/golang/protobuf/proto"
)

// Marshal is an implementation of a MarshalFunc specifically for use
// with this handler.  Although it accepts an empty interface type, it
// will only work with an apex.log.Entry type, and will panic if any
// other type is passed in.
func Marshal(x interface{}) ([]byte, error) {
	var logEntry *alog.Entry
	var timestamp []byte
	var ok bool
	if logEntry, ok = x.(*alog.Entry); !ok {
		return nil, fmt.Errorf("Attempted to marshal a type other than apex.log.Entry")
	}
	timestamp, err := logEntry.Timestamp.MarshalText()
	if err != nil {
		return nil, err
	}
	entry := &Entry{
		Level:     logEntry.Level.String(),
		Timestamp: timestamp,
		Message:   logEntry.Message,
	}
	return proto.Marshal(entry)
}

func Unmarshal(data []byte, v interface{}) error {
	var entry *Entry
	var logEntry *alog.Entry
	var ok bool

	if logEntry, ok = v.(*alog.Entry); !ok {
		return fmt.Errorf("Attempted to unmarshal to a type other than apex.log.Entry")
	}

	entry = &Entry{}
	err := proto.Unmarshal(data, entry)
	if err != nil {
		return err
	}
	logEntry.Level = alog.MustParseLevel(entry.Level)
	logEntry.Timestamp.UnmarshalText(entry.Timestamp)
	return nil
}
