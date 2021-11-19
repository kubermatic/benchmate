package main

import (
	"flag"
	"fmt"
	"github.com/kubermatic/benchmate/latency"
	"github.com/kubermatic/benchmate/throughput"
	"log"
	"sync"
)

func RunClients() {
	tpResult, err := throughput.NewThroughputMeter(throughput.DefaultOptions()).Client()
	if err != nil {
		log.Println(err)
	}
	fmt.Println(tpResult)

	latResult, err := latency.NewLatencyMeter(latency.DefaultOptions()).Client()
	if err != nil {
		log.Println(err)
	}
	fmt.Println(latResult)
}

func RunServers() {
	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		opt := throughput.DefaultOptions()
		log.Println("throughput server running with: %s", opt)
		err := throughput.NewThroughputMeter(opt).Server()
		if err != nil {
			log.Println("throughput server: ", err)
		}
	}()
	wg.Add(1)
	go func() {
		defer wg.Done()
		opt := latency.DefaultOptions()
		log.Println("latency server running with: %s", opt)
		err := latency.NewLatencyMeter(opt).Server()
		if err != nil {
			log.Println("latency server: ", err)
		}
	}()

	wg.Wait()
}

func main() {

	log.SetFlags(log.LstdFlags | log.Lshortfile)

	log.Println("Running...")

	var c bool
	flag.BoolVar(&c, "c", false, "set the flag to run in client mode. Default is server mode. ")
	flag.Parse()
	if c {
		RunClients()
	} else {
		RunServers()
	}
	log.Println("Finished.")
}
