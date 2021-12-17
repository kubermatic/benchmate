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
	"github.com/kubermatic/benchmate"
	"io"
	"io/ioutil"
	"log"
	"net"
	"sync"
	"time"
)

func prettyJSON(x interface{}) string {
	b, err := json.MarshalIndent(x, "", "  ")
	if err != nil {
		panic(err)
	}
	return string(b)
}

func runClients(tpOpt, latOpt benchmate.Options) {
	log.Println("running throughput client with:", prettyJSON(tpOpt))
	conn, err := net.Dial(tpOpt.Network, tpOpt.Addr)
	if err != nil {
		log.Fatal(err)
	}
	defer func() {
		err = conn.Close()
		if err != nil {
			log.Fatal(err)
		}
	}()
	tpResult, err := tpOpt.ThroughputClient().Run(conn)
	if err != nil {
		log.Println("throughput measurement failed:", err)
	} else {
		log.Println("throughput benchmark result:", prettyJSON(tpResult))
		log.Println("throughput: ", float64(tpResult.NumMsg*tpResult.MsgSize*1000)/float64(tpResult.Elapsed.Nanoseconds()), "MB/s")
		log.Println("throughput client done.")
	}

	log.Println("running latency client with:", prettyJSON(latOpt))
	conn, err = net.Dial(latOpt.Network, latOpt.Addr)
	if err != nil {
		log.Fatal(err)
	}
	latResult, err := latOpt.LatencyClient().Run(conn)
	if err != nil {
		log.Println("latency measurement failed:", err)
	} else {
		log.Println("latency benchmark result:", prettyJSON(latResult))
		log.Println("average latency:", time.Duration(float64(latResult.ElapsedTime.Nanoseconds())/float64(latResult.NumMsg)))
		log.Println("latency client done.")
	}
}

func runServers(tpOpt, latOpt benchmate.Options) {
	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		log.Println("running throughput server with:", prettyJSON(tpOpt))
		l, err := net.Listen(tpOpt.Network, tpOpt.Addr)
		if err != nil {
			log.Println("throughput server failed:", err)
			return
		}
		defer l.Close()
		err = tpOpt.ThroughputServer().Run(l)
		if err != nil {
			if err == io.EOF {
				log.Println("throughput server done.")
			} else {
				log.Println("throughput server failed:", err)
			}
		}
	}()
	wg.Add(1)
	go func() {
		defer wg.Done()
		log.Println("running latency server with:", prettyJSON(latOpt))
		l, err := net.Listen(latOpt.Network, latOpt.Addr)
		if err != nil {
			log.Println("latency server failed:", err)
			return
		}
		err = latOpt.LatencyServer().Run(l)
		if err != nil {
			log.Println("latency server:", err)
		} else {
			log.Println("latency server done.")
		}
	}()

	wg.Wait()
}

func main() {
	log.SetFlags(0)
	var c bool
	var latencyOpts string
	var throughputOpts string
	var nodeIP string
	flag.BoolVar(&c, "c", false, "set the flag to run in client mode. Default is server mode. ")
	flag.StringVar(&latencyOpts, "latOpt", "", "set the latency options")
	flag.StringVar(&throughputOpts, "tpOpt", "", "set the throughput options")
	flag.StringVar(&nodeIP, "node-ip", "", "override the tcpaddr setting")
	flag.Parse()

	latOpt := benchmate.DefaultLatencyOptions()
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

	tpOpt := benchmate.DefaultThroughputOptions()
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
		tpOpt.Addr = nodeIP + ":13500"
		latOpt.Addr = nodeIP + ":13501"
	}

	if c {
		runClients(tpOpt, latOpt)
	} else {
		runServers(tpOpt, latOpt)
	}

	log.Println("done.")
}
