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

package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"github.com/kubermatic/benchmate/latency"
	"github.com/kubermatic/benchmate/throughput"
	"io/ioutil"
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
