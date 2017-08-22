/*
log_to_nsq is an example program that demonstrates the use of apexovernsq.  When invoked with the IP adress and port of one or more running nsqd and a topic name, it will pu
sh two structured log messages to that nsq daemon (or deamons) and then exit.

To see this working the following three things should be invoked.

1. Start the nsqdaemon:

    nsqd

... take note of the port number for TCP that it informs you of.

2. Start nsq_tail:

    nsq_tail -topic log --nsqd-tcp-address <IPADDRESS:PORT>

Not that <IPADDRESS:PORT> should be replaced with the IP address of the machine where nsqd is running and the port number you took note of in step one.

3. invoke this program:

    ./example -nsqd-address <IPADDRESS:PORT> -topic log

This program should exit almost immediately, but if you check the nsq_tail process you should see some output that looks like this:
{"fields":{"flavour":"pistachio","scoops":"two"},"level":"info","timestamp":"2017-08-04T15:48:22.044783085+02:00","message":"It's ice cream time!"}
{"fields":{"error":"ouch, brainfreeze"},"level":"error","timestamp":"2017-08-04T15:48:22.047870426+02:00","message":"Problem consuming ice cream"}

*/
package main

import (
	"flag"
	"fmt"
	"log"
	"strings"

	"github.com/avct/apexovernsq"
	"github.com/avct/apexovernsq/protobuf"

	alog "github.com/apex/log"
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
	topic         = flag.String("topic", "", "NSQ topic to publish to [Required]")
	nsqdAddresses = stringFlags{}
)

func init() {
	flag.Var(&nsqdAddresses, "nsqd-address", "The IP Address of a nsqd you with to publish to. Give this option once for every nsqd [1 or more required].")
}

func usage() {
	flag.PrintDefaults()
}

func makeProducers(addresses stringFlags, cfg *nsq.Config) []*nsq.Producer {
	var producer *nsq.Producer
	var err error
	producerCount := len(addresses)
	producers := make([]*nsq.Producer, producerCount, producerCount)
	for i, address := range addresses {
		producer, err = nsq.NewProducer(address, cfg)
		if err != nil {
			log.Fatalf("Error creating nsq.Producer: %s", err)
		}
		producers[i] = producer
	}
	return producers
}

func makePublisher(producers []*nsq.Producer) apexovernsq.PublishFunc {
	return func(topic string, body []byte) (err error) {
		for _, producer := range producers {
			err = producer.Publish(topic, body)
			if err != nil {
				return err
			}
		}
		return nil
	}
}

func main() {
	flag.Parse()
	if len(*topic) == 0 || len(nsqdAddresses) == 0 {
		usage()
		log.Fatal("Required parameters missing.")
	}

	cfg := nsq.NewConfig()
	producers := makeProducers(nsqdAddresses, cfg)
	publisher := makePublisher(producers)
	handler := apexovernsq.NewApexLogNSQHandler(protobuf.Marshal, publisher, "log")

	alog.SetHandler(handler)
	ctx := apexovernsq.NewServiceLogContext()
	ctx.WithFields(alog.Fields{
		"flavour": "pistachio",
		"scoops":  "two",
	}).Info("It's ice cream time!")
	ctx.WithError(fmt.Errorf("ouch, brainfreeze")).Error("Problem consuming ice cream")
}
