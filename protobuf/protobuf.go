package protobuf

import (
	"fmt"

	alog "github.com/apex/log"
	proto "github.com/golang/protobuf/proto"
)

// Marshal is an implementation of a MarshalFunc specifically for use
// with this handler.  Although it accepts an empty interface type, it
// will only work with an apex.log.Entry type, and will panic if any
// other type is passed in.  Not that this mechanism also enforces the
// rule that any fields set must either be strings or satisfy the
// fmt.Stringer interface.
func Marshal(x interface{}) ([]byte, error) {
	var logEntry *alog.Entry
	var timestamp []byte
	var ok bool
	var fields map[string]string
	var stringer fmt.Stringer
	var str string

	if logEntry, ok = x.(*alog.Entry); !ok {
		return nil, fmt.Errorf("Attempted to marshal a type other than apex.log.Entry")
	}
	timestamp, err := logEntry.Timestamp.MarshalText()
	if err != nil {
		return nil, err
	}

	fields = make(map[string]string, len(logEntry.Fields))
	for key, value := range logEntry.Fields {
		// Enforcing the string or fmt.Stringer is a simple
		// way to ensure we can always push field data over
		// the line.  If we ever really want to reconstruct
		// types on the other side of NSQ then we'd probably
		// be just as well off wrapping gob.Encode and
		// gob.Decode as Marshal/Unmarshal.
		if str, ok = value.(string); ok {
			fields[key] = str
			continue
		}
		if stringer, ok = value.(fmt.Stringer); !ok {
			err := fmt.Errorf("Value for field %s is not a string, nor does it satisfy fmt.Stringer", key)
			return nil, err
		}
		fields[key] = stringer.String()
	}
	entry := &Entry{
		Level:     logEntry.Level.String(),
		Timestamp: timestamp,
		Message:   logEntry.Message,
		Fields:    fields,
	}
	return proto.Marshal(entry)
}

// Unmarshal is an implementation of a UnmarshalFunc specifically for
// unmarshalling an Entry back into apex.log.Entry.
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
	logEntry.Message = entry.Message
	if logEntry.Fields == nil {
		logEntry.Fields = make(map[string]interface{}, len(entry.Fields))
	}
	for key, value := range entry.Fields {
		logEntry.Fields[key] = value
	}
	return nil
}
