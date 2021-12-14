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
	"os"

	"github.com/kubermatic/benchmate/latency"
	"github.com/kubermatic/benchmate/throughput"
)

type Request struct {
	ThroughputOptions *throughput.Options `json:"throughput"`
	LatencyOptions    *latency.Options    `json:"latency"`
	Client            bool                `json:"client"`
}

func ExitHandler(w http.ResponseWriter, r *http.Request) {
	log.Println("Exiting...")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("Exiting..."))
	os.Exit(0)
}

func BenchmateHandler(w http.ResponseWriter, r *http.Request) {
	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		panic(err)
	}

	req := new(Request)
	err = json.Unmarshal(body, req)
	if err != nil {
		panic(err)
	}

	if req.ThroughputOptions != nil {

		if req.Client {
			log.Println("running throughput client")
			func() {
				resp, err := throughput.NewThroughputMeter(*req.ThroughputOptions).Client()
				if err != nil {
					log.Println(err)
				}
				err = json.NewEncoder(w).Encode(resp)
				if err != nil {
					log.Println(err)
				}
			}()
		} else {
			log.Println("running throughput server")
			if err != nil {
				panic(err)
			}
			go func() {
				err = throughput.NewThroughputMeter(*req.ThroughputOptions).Server()
				if err != nil {
					log.Println(err)
				}
			}()
		}
	}

	if req.LatencyOptions != nil {
		if req.Client {
			log.Println("running latency client")
			func() {
				result, err := latency.NewLatencyMeter(*req.LatencyOptions).Client()
				if err != nil {
					log.Println(err)
				}

				err = json.NewEncoder(w).Encode(result)
				if err != nil {
					log.Println(err)
				}
			}()
		} else {
			log.Println("running latency server")

			go func() {
				err = latency.NewLatencyMeter(*req.LatencyOptions).Server()
				if err != nil {
					log.Println(err)
				}
			}()
		}
	}
}
