package server

import (
	"encoding/json"
	"github.com/kubermatic/benchmate/latency"
	"github.com/kubermatic/benchmate/throughput"
	"io/ioutil"
	"log"
	"net/http"
)

type Request struct {
	ThroughputOptions *throughput.Options `json:"throughput"`
	LatencyOptions    *latency.Options    `json:"latency"`
	Client            bool                `json:"client"`
}

func StatsHandler(w http.ResponseWriter, r *http.Request) {
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
