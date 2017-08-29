/*
nsq-log-tail is a program that will monitor a topic on one or more nsqd instances and attempt to convert messages on that topic into human readable log output.
*/

package main

import (
	"errors"
	"flag"
	"fmt"
	"log"
	"math/rand"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/avct/apexovernsq"
	"github.com/avct/apexovernsq/protobuf"

	alog "github.com/apex/log"
	"github.com/apex/log/handlers/cli"
	"github.com/apex/log/handlers/logfmt"
	nsq "github.com/nsqio/go-nsq"
)

type stringFlags []string

func (n *stringFlags) Set(value string) error {
	*n = append(*n, value)
	return nil
}

func (n *stringFlags) String() string {
	return strings.Join(*n, ",")
}

func listenToNSQ(consumer *nsq.Consumer, p *parameters) error {
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	err := consumer.ConnectToNSQDs(p.nsqdTCPAddrs)
	if err != nil {
		return err
	}

	err = consumer.ConnectToNSQLookupds(p.lookupdHTTPAddrs)
	if err != nil {
		return err
	}
	for {
		select {
		case <-consumer.StopChan:
			return nil
		case <-sigChan:
			consumer.Stop()
		}
	}
}

func generateEphemeralChannelName() string {
	rand.Seed(time.Now().UnixNano())
	return fmt.Sprintf("tail%06d#ephemeral", rand.Int()%999999)
}

func logFromNSQ(p *parameters) error {
	var handler alog.Handler
	var logHandler nsq.Handler

	cfg := nsq.NewConfig()
	channel := generateEphemeralChannelName()
	consumer, err := nsq.NewConsumer(*p.topic, channel, cfg)
	if err != nil {
		return err
	}
	handler = logfmt.New(os.Stdout)
	if *p.useCLIHandler {
		handler = cli.New(os.Stdout)
	}
	if p.services != nil {
		strings := []string(p.services)
		serviceFilter := apexovernsq.NewApexLogServiceFilterHandler(handler, &strings)
		logHandler = apexovernsq.NewNSQApexLogHandler(serviceFilter, protobuf.Unmarshal)
	} else {
		logHandler = apexovernsq.NewNSQApexLogHandler(handler, protobuf.Unmarshal)
	}
	consumer.AddHandler(logHandler)

	return listenToNSQ(consumer, p)
}

type parameters struct {
	topic            *string
	useCLIHandler    *bool
	services         stringFlags
	nsqdTCPAddrs     stringFlags
	lookupdHTTPAddrs stringFlags
}

func newParameters() *parameters {
	p := &parameters{
		topic:            flag.String("topic", "", "NSQ topic to consume from [Required]"),
		useCLIHandler:    flag.Bool("cli", false, "Use CLI output handler"),
		services:         stringFlags{},
		nsqdTCPAddrs:     stringFlags{},
		lookupdHTTPAddrs: stringFlags{},
	}
	flag.Var(&p.nsqdTCPAddrs, "nsqd-tcp-address", "nsqd TCP address (may be given multiple times)")
	flag.Var(&p.lookupdHTTPAddrs, "lookupd-http-address", "lookupd HTTP address (may be given multiple times)")
	flag.Var(&p.services, "service", "service to output logs for (may be given multiple times). If no service flag is specified, logs for all services will be output")

	return p
}

func (p *parameters) check() error {
	if *p.topic == "" {
		return errors.New("Please provide a topic")
	}

	if len(p.nsqdTCPAddrs) == 0 && len(p.lookupdHTTPAddrs) == 0 {
		return errors.New("--nsqd-tcp-address or --lookupd-http-address required")
	}
	if len(p.nsqdTCPAddrs) > 0 && len(p.lookupdHTTPAddrs) > 0 {
		return errors.New("use --nsqd-tcp-address or --lookupd-http-address not both")
	}
	return nil
}

func main() {
	p := newParameters()
	flag.Parse()
	if err := p.check(); err != nil {
		flag.PrintDefaults()
		log.Fatal(err)
	}
	err := logFromNSQ(p)
	if err != nil {
		log.Fatal(err)
	}
}
