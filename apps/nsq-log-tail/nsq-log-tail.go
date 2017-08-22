/*
nsq-log-tail is a program that will monitor a topic on one or more nsqd instances and attempt to convert messages on that topic into human readable log output.
*/

package main

import (
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
	"github.com/apex/log/handlers/text"
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

var (
	topic            = flag.String("topic", "", "NSQ topic to consume from [Required]")
	useCLIHandler    = flag.Bool("cli", false, "Use CLI output handler")
	nsqdTCPAddrs     = stringFlags{}
	lookupdHTTPAddrs = stringFlags{}
)

func init() {
	flag.Var(&nsqdTCPAddrs, "nsqd-tcp-address", "nsqd TCP address (may be given multiple times)")
	flag.Var(&lookupdHTTPAddrs, "lookupd-http-address", "lookupd HTTP address (may be given multiple times)")
}

func listenToNSQ(consumer *nsq.Consumer) error {
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	err := consumer.ConnectToNSQDs(nsqdTCPAddrs)
	if err != nil {
		return err
	}

	err = consumer.ConnectToNSQLookupds(lookupdHTTPAddrs)
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

func logFromNSQ() error {
	var handler alog.Handler
	cfg := nsq.NewConfig()
	channel := generateEphemeralChannelName()
	consumer, err := nsq.NewConsumer(*topic, channel, cfg)
	if err != nil {
		return err
	}
	handler = text.Default
	if *useCLIHandler {
		handler = cli.Default
	}
	consumer.AddHandler(apexovernsq.NewNSQApexLogHandler(handler, protobuf.Unmarshal))

	return listenToNSQ(consumer)
}

func checkParamters() error {
	if *topic == "" {
		return fmt.Errorf("--topic is required")
	}

	if len(nsqdTCPAddrs) == 0 && len(lookupdHTTPAddrs) == 0 {
		return fmt.Errorf("--nsqd-tcp-address or --lookupd-http-address required")
	}
	if len(nsqdTCPAddrs) > 0 && len(lookupdHTTPAddrs) > 0 {
		return fmt.Errorf("use --nsqd-tcp-address or --lookupd-http-address not both")
	}
	return nil
}

func main() {
	flag.Parse()
	if err := checkParamters(); err != nil {
		log.Fatal(err)
	}
	err := logFromNSQ()
	if err != nil {
		log.Fatal(err)
	}
}
