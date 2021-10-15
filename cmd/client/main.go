// Latency benchmark for comparing Unix sockets with TCP sockets.
//
// Idea: ping-pong 128-byte packets between a goroutine acting as a Server and
// main acting as client. Measure how long it took to do 2*N ping-pongs and find
// the average latency.
//
// Eli Bendersky [http://eli.thegreenplace.net]
// This code is in the public domain.
package main

import (
	"flag"
	"github.com/pratikdeoghare/benchmate/latency"
	"github.com/pratikdeoghare/benchmate/throughput"
	"log"
)

func main() {
	flag.Parse()
	err := throughput.Client()
	if err != nil {
		log.Println(err)
	}
	err = latency.Client()
	if err != nil {
		log.Println(err)
	}
}
