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

// This package provides pporf like handlers for network performance analysis.
package benchmate

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"net/http"
)

// ThroughputRequest provides options for throughput benchmarks.
// Set Client to true if you want the handler to run a client.
type ThroughputRequest struct {
	ThroughputOptions
	Client bool `json:"client"`
}

// ThroughputHandler runs client/server for throughput benchmarks.
func ThroughputHandler(w http.ResponseWriter, r *http.Request) {
	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	req := new(ThroughputRequest)
	err = json.Unmarshal(body, req)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	if req.Client {
		log.Println("running throughput client")
		resp, err := NewThroughputMeter(req.ThroughputOptions).Client()
		if err != nil {
			log.Println(err)
		}
		err = json.NewEncoder(w).Encode(resp)
		if err != nil {
			log.Println(err)
		}
	} else {
		log.Println("running throughput server")
		if err != nil {
			panic(err)
		}
		err = NewThroughputMeter(req.ThroughputOptions).Server()
		if err != nil {
			log.Println(err)
		}
	}
}

// LatencyRequest provides options for latency benchmarks.
// Set Client to true if you want the handler to run a client.
type LatencyRequest struct {
	LatencyOptions
	Client bool `json:"client"`
}

// LatencyHandler runs client/server for latency benchmarks.
func LatencyHandler(w http.ResponseWriter, r *http.Request) {
	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		log.Println(err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	req := new(LatencyRequest)
	err = json.Unmarshal(body, req)
	if err != nil {
		log.Println(err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	if req.Client {
		log.Println("running latency client")
		result, err := NewLatencyMeter(req.LatencyOptions).Client()
		if err != nil {
			log.Println(err)
		}

		err = json.NewEncoder(w).Encode(result)
		if err != nil {
			log.Println(err)
		}
	} else {
		log.Println("running latency server")

		err = NewLatencyMeter(req.LatencyOptions).Server()
		if err != nil {
			log.Println(err)
		}
	}
}