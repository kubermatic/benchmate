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

//A toy program to demonstrate the usage of handlers LatencyHandler and ThroughputHandler.
//
// Run one instance at 8888.
//	#  ./bmserver --addr=:8888
//Run another instance at 9999.
//	#  ./bmserver --addr=:9999
//Start a latency server at 13501 by sending request to localhost:8888/benchmate/latency endpoint.
//
//	# curl http://localhost:8888/benchmate/latency --data '
//	{
//		"msgSize": 128,
//		"numMsg": 1000,
//		"network": "tcp",
//		"addr": ":13501",
//		"timeout": 120000
//	}
//	'
//Start a latency client by sending request to localhost:9999/benchmate/latency endpoint.
//
//  # curl http://localhost:9999/benchmate/latency --data '
//  {
//   	"msgSize": 128,
//  	"numMsg": 1000,
//  	"network": "tcp",
//  	"addr": ":13501",
//  	"timeout": 120000,
//  	"client": true
//  }
//  '
//  {"elapsedTime":29707159,"numPings":2000,"avgLatency":14853}
//
//Results of the benchmark run are printed to stdout. In the above example, the latency is 14853ns.
package main

import (
	"flag"
	"log"
	"net/http"

	"github.com/kubermatic/benchmate"
)

func main() {
	var addr string
	flag.StringVar(&addr, "addr", ":8080", "Address to listen on")
	flag.Parse()

	mux := http.NewServeMux()
	mux.HandleFunc("/benchmate/throughput", benchmate.ThroughputHandler)
	mux.HandleFunc("/benchmate/latency", benchmate.LatencyHandler)
	log.Fatal(http.ListenAndServe(addr, mux))
}
