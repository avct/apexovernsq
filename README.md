# nsqhandler

## Overview
nsqhandler is a handler for [github.com/apex/log](https://github.com/apex/log).

In order to create a new nsqhandler instance you'll need to call `nsqhandler.New` and pass it three things:

   * A function with a signature matching `nsqhandler.MarshalFunc` to convert an apex `log.Entry` into a slice of bytes.
   * A function with a signature matching `nsqhandler.PublishFunc`. This is typically `go-nsq.Producer.Publish`, or a function that wraps it.
   * A string naming the nsq topic the log messages will be sent to.

Once you've got a handler, you can use it in apex/log by calling `github.com/apex/log.SetHandler`, with your handler instance as it's only argument.  For a more detailed usage example please look at the `log_to_nsq` program in the `examples` directory.
