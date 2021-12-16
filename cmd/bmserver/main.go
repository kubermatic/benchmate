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
	"flag"
	"log"
	"net/http"

	bmHandler "github.com/kubermatic/benchmate/pkg/handler"
)

func main() {
	var addr string
	flag.StringVar(&addr, "addr", ":8080", "Address to listen on")
	flag.Parse()

	mux := http.NewServeMux()
	mux.HandleFunc("/benchmate/throughput", bmHandler.Throughput)
	mux.HandleFunc("/benchmate/latency", bmHandler.Latency)
	log.Fatal(http.ListenAndServe(addr, mux))
}
