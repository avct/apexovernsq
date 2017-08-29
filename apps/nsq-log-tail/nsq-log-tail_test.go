package main

import "testing"

func assertError(t *testing.T, err error, expected string) {
	if err == nil {
		t.Error("Expected error, nil returned")
		return
	}
	msg := err.Error()
	if msg != expected {
		t.Errorf("Expected %q, got %q", expected, msg)
	}
}

func TestCheckParameters(t *testing.T) {
	p := newParameters()
	*p.topic = ""
	err := p.check()
	assertError(t, err, "Please provide a topic")

	*p.topic = "thisworks"
	err = p.check()
	assertError(t, err, "--nsqd-tcp-address or --lookupd-http-address required")

	p.lookupdHTTPAddrs.Set("foo")
	p.nsqdTCPAddrs.Set("bar")
	err = p.check()
	assertError(t, err, "use --nsqd-tcp-address or --lookupd-http-address not both")

	p.lookupdHTTPAddrs = stringFlags{}
	err = p.check()
	if err != nil {
		t.Errorf("Expect nil, but got an error: %q", err.Error())
	}
}
