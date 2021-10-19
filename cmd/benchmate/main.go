package main

import (
	"flag"
	"github.com/kubermatic/benchmate/latency"
	"github.com/kubermatic/benchmate/throughput"
	"log"
	"sync"
)

func RunClients() {
	err := throughput.Client()
	if err != nil {
		log.Println(err)
	}

	err = latency.Client()
	if err != nil {
		log.Println(err)
	}
}

func RunServers() {
	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		err := throughput.Server()
		if err != nil {
			log.Println("throughput server: ", err)
		}
	}()
	wg.Add(1)
	go func() {
		defer wg.Done()
		err := latency.Server()
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
