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

package benchmate

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"net"
	"net/http"
)

// ThroughputRequest provides options for throughput benchmarks.
// Set Client to true if you want the handler to run a client.
type ThroughputRequest struct {
	Options
	Client bool `json:"client"`
}

// ThroughputHandler can be added to HTTP mux. Like this,
//	mux := http.NewServeMux()
//	mux.HandleFunc("/benchmate/throughput", bmHandler.ThroughputHandler)
//	log.Fatal(http.ListenAndServe(addr, mux))
//It can be triggered to run client or server for throughput estimation.
//  # curl http://localhost:8888/benchmate/throughput --data '
//  {
//	 "msgSize": 128000,
//	 "numMsg": 10000,
//	 "network": "tcp",
//	 "addr": ":13500",
//	 "timeout": 120000
//  }
//  '
//
//  # curl http://localhost:9999/benchmate/throughput --data '
//  {
//   	"msgSize": 128000,
//  	"numMsg": 10000,
//  	"network": "tcp",
//  	"addr": ":13500",
//  	"timeout": 120000,
//  	"client": true
//  }
//  '
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
		conn, err := net.Dial(req.Network, req.Addr)
		if err != nil {
			log.Println(err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		defer conn.Close()
		resp, err := NewThroughputClient(req.MsgSize, req.NumMsg, req.Timeout).Run(conn)
		if err != nil {
			log.Println(err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		err = json.NewEncoder(w).Encode(resp)
		if err != nil {
			log.Println(err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
	} else {
		log.Println("running throughput server")
		if err != nil {
			log.Println(err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		l, err := net.Listen(req.Network, req.Addr)
		if err != nil {
			log.Println(err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		defer l.Close()

		err = NewThroughputServer(req.MsgSize).Run(l)
		if err != nil {
			log.Println(err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
	}
}

// LatencyRequest provides options for latency benchmarks.
// Set Client to true if you want the handler to run a client.
type LatencyRequest struct {
	Options
	Client bool `json:"client"`
}

// LatencyHandler can be added to HTTP mux. Like this,
//	mux := http.NewServeMux()
//	mux.HandleFunc("/benchmate/latency", bmHandler.LatencyHandler)
//	log.Fatal(http.ListenAndServe(addr, mux))
//It can be triggered to run client or server for latency estimation.
//  # curl http://localhost:8888/benchmate/latency --data '
//  {
//	 "msgSize": 128,
//	 "numMsg": 1000,
//	 "network": "tcp",
//	 "addr": ":13501",
//	 "timeout": 120000
//  }
//  '
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

		conn, err := net.Dial(req.Network, req.Addr)
		if err != nil {
			log.Println(err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		defer conn.Close()

		log.Println("running latency client")

		result, err := NewLatencyClient(req.MsgSize, req.NumMsg, req.Timeout).Run(conn)
		if err != nil {
			log.Println(err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		err = json.NewEncoder(w).Encode(result)
		if err != nil {
			log.Println(err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

	} else {

		l, err := net.Listen(req.Network, req.Addr)
		if err != nil {
			log.Println(err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		defer l.Close()

		log.Println("running latency server")
		err = NewLatencyServer(req.MsgSize, req.NumMsg).Run(l)
		if err != nil {
			log.Println(err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
	}
}
