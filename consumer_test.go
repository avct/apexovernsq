package apexovernsq

import (
	"fmt"
	"testing"

	"github.com/avct/apexovernsq/protobuf"

	alog "github.com/apex/log"
	"github.com/apex/log/handlers/memory"
	nsq "github.com/nsqio/go-nsq"
)

type entryList []*alog.Entry

// simulateMessageFromNSQ packs an Apex Log Entry into an NSQ Message
// and pushes it through the NSQApexLogHandler.  The NSQApexLogHandler
// will use the Apex Log memory handler to return the Apex Log Entry
// that would be logged.
func simulateMessageFromNSQ(sourceEntry *alog.Entry, filter *[]string) (*alog.Entry, error) {
	var handler *NSQApexLogHandler
	marshalledEntry, err := protobuf.Marshal(sourceEntry)
	if err != nil {
		return nil, fmt.Errorf("Couldn't marshal log entry: %s", err.Error())
	}

	msg := nsq.NewMessage(nsq.MessageID{'a', 'b', 'c'}, []byte(marshalledEntry))

	innerHandler := memory.New()
	if filter == nil {
		handler = NewNSQApexLogHandler(innerHandler, protobuf.Unmarshal)
	} else {
		filterHandler := NewApexLogServiceFilterHandler(innerHandler, filter)
		handler = NewNSQApexLogHandler(filterHandler, protobuf.Unmarshal)
	}

	handler.HandleMessage(msg)
	return innerHandler.Entries[0], nil
}

// simulateMessagesFromNSQ packs an entryList into a series of NSQ Messages and pushes them through the NSQApexLogHandler.  The NSQApexLogHandler will use the Apex Log memory handler to return an entryList that would be logged locally.
func simulateMessagesFromNSQ(sourceEntries *entryList, filter *[]string) (*entryList, error) {
	var marshalledEntry []byte
	var err error
	var msg *nsq.Message
	var handler nsq.Handler

	innerHandler := memory.New()
	if filter == nil {
		handler = NewNSQApexLogHandler(innerHandler, protobuf.Unmarshal)
	} else {
		filterHandler := NewApexLogServiceFilterHandler(innerHandler, filter)
		handler = NewNSQApexLogHandler(filterHandler, protobuf.Unmarshal)
	}

	for _, sourceEntry := range *sourceEntries {
		marshalledEntry, err = protobuf.Marshal(sourceEntry)
		if err != nil {
			return nil, fmt.Errorf("Couldn't marshal log entry: %s", err.Error())
		}

		msg = nsq.NewMessage(nsq.MessageID{'a', 'b', 'c'}, marshalledEntry)
		handler.HandleMessage(msg)
	}
	result := innerHandler.Entries[:]
	return (*entryList)(&result), nil

}

// simulateEntry returns a finalised Entry from the memory handler as
// it would appear in normal logging.
func simulateEntry(logger alog.Interface, fields alog.Fielder, msg string) *alog.Entry {
	memoryHandler := memory.New()
	alog.SetHandler(memoryHandler)
	if fields == nil {
		logger.Info(msg)
	} else {
		logger.WithFields(fields).Info(msg)
	}
	return memoryHandler.Entries[0]
}

// simulateEntries makes a stipulated number of finalised Apex Log
// Entry instances and stores them in a provided store (a pointer to
// an entryList), starting from a given offset. ⚠ Danger Will
// Robinson, here be side-effects. The provided offset and store will
// both be mutated. ⚠
func simulateEntries(ctx alog.Interface, count int, offset *int, store *entryList) {
	var m int
	for m = 0; m < count; m++ {
		entry := simulateEntry(ctx, nil, fmt.Sprintf("%d", *offset))
		(*store)[*offset] = entry
		*offset++
	}
}

func TestHandleMessage(t *testing.T) {
	logger := alog.Log
	sourceEntry := simulateEntry(logger, alog.Fields{
		"flavour": "pistachio",
		"scoops":  "2",
	}, "it's ice cream time!")

	entry, err := simulateMessageFromNSQ(sourceEntry, nil)
	if err != nil {
		t.Fatalf("%s", err.Error())
	}
	if entry.Message != sourceEntry.Message {
		t.Errorf("Expected %q, got %q", sourceEntry.Message, entry.Message)
	}
	if entry.Level != sourceEntry.Level {
		t.Errorf("Expected %s, got %s", sourceEntry.Level, entry.Level)
	}
	if entry.Timestamp != sourceEntry.Timestamp {
		t.Errorf("Expected %q, got %q", sourceEntry.Timestamp, entry.Timestamp)
	}
	expectedFieldCount := len(sourceEntry.Fields)
	actualFieldCount := len(entry.Fields)
	if expectedFieldCount != actualFieldCount {
		t.Errorf("Expected %d fields, but got %d fields", expectedFieldCount, actualFieldCount)
	}
	for fieldName, value := range sourceEntry.Fields {
		recieved := entry.Fields.Get(fieldName)
		if recieved != value {
			t.Errorf("Expected %s=%q, got %s=%q", fieldName, value, fieldName, recieved)
		}
	}
}

func TestFilterMessagesByService(t *testing.T) {
	var ctx alog.Interface
	var entries *entryList
	var err error
	var generatedMessages int
	var resultEntryCount int
	var sourceEntries entryList
	var totalMessages int

	var thisProcessFilter = &[]string{processName()}
	var otherProcessFilter = &[]string{"other"}
	var unknownProcessFilter = &[]string{"unknown"}
	var multiServiceFilter = &[]string{processName(), "other"}

	var caseTable = []struct {
		serviceMessages      int
		otherServiceMessages int
		nonServiceMessages   int
		resultEntryCount     int
		filter               *[]string
		messages             []string
	}{
		{1, 1, 1, 3, nil, []string{"0", "1", "2"}},           // Default case, no filtering
		{1, 1, 1, 1, thisProcessFilter, []string{"0"}},       // Whitelist service messages
		{1, 1, 1, 1, otherProcessFilter, []string{"1"}},      // Whitelist other service messages
		{1, 1, 1, 0, unknownProcessFilter, []string{}},       // Whitelist service that isn't present
		{1, 1, 1, 2, multiServiceFilter, []string{"0", "1"}}, // Whitelist both this service and "other" service

	}

	for caseNum, testCase := range caseTable {
		totalMessages = testCase.serviceMessages + testCase.nonServiceMessages + testCase.otherServiceMessages
		generatedMessages = 0
		sourceEntries = make([]*alog.Entry, totalMessages)

		// Generate service messages
		ctx = NewApexLogServiceContext()
		simulateEntries(ctx, testCase.serviceMessages, &generatedMessages, &sourceEntries)
		// Generate other service messages
		ctx = ctx.WithFields(alog.Fields{"service": "other"})
		simulateEntries(ctx, testCase.otherServiceMessages, &generatedMessages, &sourceEntries)

		// Generate non-service messages
		ctx = alog.Log
		simulateEntries(ctx, testCase.nonServiceMessages, &generatedMessages, &sourceEntries)

		entries, err = simulateMessagesFromNSQ(&sourceEntries, testCase.filter)
		if err != nil {
			t.Fatalf("Error in message simulation: %s", err.Error())
		}
		if entries == nil && testCase.resultEntryCount > 0 {
			t.Errorf("Expected %d entries from test case %d but got nil. Message counts: service %d, other-service %d, non-service %d.  Filter: %+v.", testCase.resultEntryCount, caseNum, testCase.serviceMessages, testCase.otherServiceMessages, testCase.nonServiceMessages, testCase.filter)
			continue
		}
		resultEntryCount = len(*entries)
		if resultEntryCount != testCase.resultEntryCount {
			t.Errorf("[Case %d] Expected %d entries, but got %d. Message counts: service %d, other-service %d, non-service %d.  Filter: %+v.", caseNum+1, testCase.resultEntryCount, resultEntryCount, testCase.serviceMessages, testCase.otherServiceMessages, testCase.nonServiceMessages, testCase.filter)
			continue
		}

		for idx, entry := range *entries {
			testMessage := testCase.messages[idx]
			if entry.Message != testMessage {
				t.Errorf("[Case %d] Expected Entry[%d] to have message %q, but got %q", caseNum, idx, testMessage, entry.Message)
			}
		}
	}
}
