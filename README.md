# apexovernsq

## Overview
The apexovernsq package provides a mechanism to transfer structured log entries, generated with the Apex log package ( [github.com/apex/log](https://github.com/apex/log) ) over NSQ ( [github.com/nsqio](https://github.com/nsqio) ).  Specifically it allows Apex's `log.Entry` structs to be marshalled, published to an NSQ topic and then unmarshalled at the other end and injected into a local Apex log handler.

## Putting log messages onto the NSQ channel

To push log messages onto an NSQ channel we provide a type that implements the `github.com/apex/log.Handler` interface.  In order to create a new `ApexLogNSQHandler` instance you'll need to call `apexovernsq.NewApexLogNSQHandler` and pass it three things:

   * A function with a signature matching `apexovernsq.MarshalFunc` to convert an apex `log.Entry` into a slice of bytes.
   * A function with a signature matching `apexovernsq.PublishFunc`. This is typically `github.com/nsqio/go-nsq.Producer.Publish`, or a function that wraps it.
   * A string naming the nsq topic the log messages will be sent to.

Once you've got a handler, you can use it in apex/log by calling `github.com/apex/log.SetHandler`, with your handler instance as it's only argument. 

### Partial Example

```go
package main

import (
	"github.com/avct/apexovernsq"
	nsq "github.com/nsqio/go-nsq"
	alog "github.com/apex/log"
)


func main() {
	// This is a very minimal NSQ configuation, you'll need nsqd
	// running locally to make this work.
	cfg := nsq.NewConfig()
	nsqdAddress := "127.0.0.1:4150"
	producer := nsq.NewProducer(nsqdAddress, cfg)

	handler := apexovernsq.NewApexLogNSQHandler(json.Marshal, producer.Publish, "log")
	alog.SetHandler(handler)
	// From this point onward, logging via apex log will be forwarded over NSQ
}
```

For a more detailed usage example please look at the `log_to_nsq` program in the `examples` directory.

## Consuming apex log messages from NSQ

To consume apex log `Entry` structs from NSQ an NSQ handler is
provided.  To construct an `NSQApexLogHandler` you must call `apexovernsq.NewNSQApexLogHandler` with two arguments:

   * an `github.com/apex/log.Handler`implementation which will handle the log messages as they arrive.  For example, if you use the `github.com/apex/log/handlers/cli.Default` the log messages will be output to `os.Stderr` on the consuming process.
   * a function with a signature that matches `apexovernsq.UnmarshalFunc`- for example the `json.Unmarshal`.  Note, this must match to the function used to marshal the log entries before they are published on NSQ.
   
### Partial Example

```go
package main

import (
	"encoding/json"
	"os"
	"os/signal"
	"syscall"

	alog "github.com/apex/log"
	"github.com/apex/log/handlers/cli"
	"github.com/avct/apexovernsq"
	nsq "github.com/nsqio/go-nsq"
)

func main() {
	cfg := nsq.NewConfig()
	channel := "mychannel#ephemeral"

	// Note, it's important you consume from the same topic name
	// that you publish the messaegs to.
	consumer, err := nsq.NewConsumer("log", channel, cfg)
	if err != nil {
		alog.WithError(err).Error("error setting up NSQ consumer")
		os.Exit(1)
	}

	// We choose the apex log handler we'd like to pump our log
	// messages through.  They'll be alog.Entry instances passed
	// to the HandleLog function for the handler, exactly as if
	// they were produced locally.
	apexHandler := cli.Default

	// We create an NSQ handler that will unmarshal the entries
	// from NSQ and pump them through the provided apex log
	// handler.
	nsqHandler := apexovernsq.NewNSQApexLogHandler(apexHandler, json.Unmarshal)

	// .. and we tell the NSQ consumer to use our new handler.
	consumer.AddHandler(nsqHandler)

	// Note, this is a very simplistic NSQ setup.  You'll need
	// nsqd running on the localhost to make this work.
	err := consumer.ConnectToNSQD("127.0.0.1:4150")
	if err != nil {
		alog.WithError(err).Error("error connecting to NSQD")
		os.Exit(2)
	}

	// This block makes us loop listening to NSQ until the program is terminated.
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	for {
		select {
		case <-consumer.StopChan:
			return nil
		case <-sigChan:
			consumer.Stop()
		}
	}

}
```

For a more complete example look at the `nsq-log-tail`programg in the `apps` subdirectory.

# Nicities

Additionally we provide a few additional useful mecahisms.

## NewApexLogServiceContext

The function `apexovernsq.NewApexLogServiceContext`returns an apex log `Entry`with some standard fields set: 

   * "service" - the name of the process that is logging.
   * "hostname" - the hostname of the machine that created the log message.
   * "pid" - the process ID of the process that created the log message.
   
You can pass this `Entry` around and use it as a context for log calls (as per normal operaion with apex log).  Having these standard fields set is very helpful if, for example, you wish to aggregate the logs from multiple services and/or hosts.

## Protobuf

We provide a protobuf definition of the apex log `Entry` struct, which
generates a go library containing a `Marshal` and an `Unmarshal`
function that can be used by the producing and consuming handlers in
`apexovernsq`. You'll find these functions by importing
`github.com/avct/apexovernsq/protobuf`
