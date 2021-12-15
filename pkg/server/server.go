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

package server

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"net/http"

	"github.com/kubermatic/benchmate/pkg/latency"
	"github.com/kubermatic/benchmate/pkg/throughput"
)

type Request struct {
	ThroughputOptions *throughput.Options `json:"tpOpt"`
	LatencyOptions    *latency.Options    `json:"latOpt"`
	Client            bool                `json:"client"`
	ServerTimeout     int                 `json:"serverTimeout"`
}

func Throughput(w http.ResponseWriter, r *http.Request) {
	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	req := new(Request)
	err = json.Unmarshal(body, req)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	if req.ThroughputOptions == nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	if req.Client {
		log.Println("running throughput client")
		resp, err := throughput.NewThroughputMeter(*req.ThroughputOptions).Client()
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
		err = throughput.NewThroughputMeter(*req.ThroughputOptions).Server()
		if err != nil {
			log.Println(err)
		}
	}
}

func Latency(w http.ResponseWriter, r *http.Request) {
	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		log.Println(err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	req := new(Request)
	err = json.Unmarshal(body, req)
	if err != nil {
		log.Println(err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	if req.LatencyOptions == nil {
		log.Println(err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	if req.Client {
		log.Println("running latency client")
		result, err := latency.NewLatencyMeter(*req.LatencyOptions).Client()
		if err != nil {
			log.Println(err)
		}

		err = json.NewEncoder(w).Encode(result)
		if err != nil {
			log.Println(err)
		}
	} else {
		log.Println("running latency server")

		err = latency.NewLatencyMeter(*req.LatencyOptions).Server()
		if err != nil {
			log.Println(err)
		}
	}
}
