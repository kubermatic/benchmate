package main

import (
	"flag"
	"fmt"
	"github.com/kubermatic/benchmate/latency"
	"github.com/kubermatic/benchmate/throughput"
	"log"
	"sync"
)

func RunClients(tpOpt throughput.Options, latOpt latency.Options) {
	tpResult, err := throughput.NewThroughputMeter(tpOpt).Client()
	if err != nil {
		log.Println(err)
	}
	fmt.Println(tpResult)

	latResult, err := latency.NewLatencyMeter(latOpt).Client()
	if err != nil {
		log.Println(err)
	}
	fmt.Println(latResult)
}

func RunServers(tpOpt throughput.Options, latOpt latency.Options) {
	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		log.Println("throughput server running with: ", tpOpt)
		err := throughput.NewThroughputMeter(tpOpt).Server()
		if err != nil {
			log.Println("throughput server: ", err)
		}
	}()
	wg.Add(1)
	go func() {
		defer wg.Done()
		log.Println("latency server running with: ", latOpt)
		err := latency.NewLatencyMeter(latOpt).Server()
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
	var nodeIP string
	flag.BoolVar(&c, "c", false, "set the flag to run in client mode. Default is server mode. ")
	flag.StringVar(&nodeIP, "node-ip", "", "IP address of benchmate server")
	flag.Parse()
	log.Println("node-ip", nodeIP)
	tpOpt := throughput.DefaultOptions()
	latOpt := latency.DefaultOptions()
	if c {
		tpOpt.TcpAddress = nodeIP + ":13500"
		latOpt.TcpAddress = nodeIP + ":13501"
		RunClients(tpOpt, latOpt)
	} else {
		RunServers(tpOpt, latOpt)
	}
	log.Println("Finished.")
}
