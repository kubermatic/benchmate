/*
Copyright 2021 The Kubermatic Kubernetes Platform contributors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

// This is a program that can be used to benchmark the performance of the
// network. You can measure latency and throughput of the network.
//
// You run server somewhere lime
//	 $ benchmate
//
// You run client somewhere like
//
// 	$ benchmate -c
//
// As long as client can talk to the server, you can measure the latency and throughput.
//
// You can configure the details using json files and supply them as arguments.
//  $ benchmate -c --latOpt=latOpt.json --tpOpt=tpOpt.json
// This will read the the benchmark parameters such as message size from the json
// files and use them.
//
// Sample json files can be found i the hack/examples folder.
//
// benchmate is built using the library in https://pkg.go.dev/github.com/kubermatic/benchmate/pkg/latency/ and pkg/throughput.
//
// You can use the library to add network benchmarking to your application.
package main

import (
	"encoding/json"
	"flag"
	"github.com/kubermatic/benchmate/pkg/latency"
	"github.com/kubermatic/benchmate/pkg/throughput"
	"io/ioutil"
	"log"
	"sync"
)

func RunClients(tpOpt throughput.Options, latOpt latency.Options) {
	tpResult, err := throughput.NewThroughputMeter(tpOpt).Client()
	if err != nil {
		log.Println(err)
	}

	result, err := json.MarshalIndent(tpResult, "", "  ")
	if err != nil {
		log.Println(err)
	}

	log.Println(string(result))
	log.Println("throughput: ", tpResult.ThroughputMBPerSec, "MB/s")

	latResult, err := latency.NewLatencyMeter(latOpt).Client()
	if err != nil {
		log.Println(err)
	}

	result, err = json.MarshalIndent(latResult, "", "  ")
	if err != nil {
		log.Println(err)
	}
	log.Println(string(result))
	log.Println("average latency:", latResult.AvgLatency)
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
	var latencyOpts string
	var throughputOpts string
	var nodeIP string
	flag.BoolVar(&c, "c", false, "set the flag to run in client mode. Default is server mode. ")
	flag.StringVar(&latencyOpts, "latOpt", "", "set the latency options")
	flag.StringVar(&throughputOpts, "tpOpt", "", "set the throughput options")
	flag.StringVar(&nodeIP, "node-ip", "", "override the tcpaddr setting")
	flag.Parse()

	latOpt := latency.DefaultOptions()
	if latencyOpts != "" {
		data, err := ioutil.ReadFile(latencyOpts)
		if err != nil {
			panic(err)
		}
		err = json.Unmarshal(data, &latOpt)
		if err != nil {
			panic(err)
		}
	}

	tpOpt := throughput.DefaultOptions()
	if throughputOpts != "" {
		data, err := ioutil.ReadFile(throughputOpts)
		if err != nil {
			panic(err)
		}
		err = json.Unmarshal(data, &tpOpt)
		if err != nil {
			panic(err)
		}
	}

	if nodeIP != "" {
		tpOpt.TcpAddress = nodeIP + ":13500"
		latOpt.TcpAddress = nodeIP + ":13501"
	}

	if c {
		RunClients(tpOpt, latOpt)
	} else {
		RunServers(tpOpt, latOpt)
	}
	log.Println("Finished.")
}
