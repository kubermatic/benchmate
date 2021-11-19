package main

import (
	"encoding/json"
	"github.com/kubermatic/benchmate/latency"
	"github.com/kubermatic/benchmate/throughput"
	"io/ioutil"
	"log"
	"net/http"
)

type server struct {
	addr string
}

func NewServer(addr string) *server {
	return &server{
		addr: addr,
	}
}

type Request struct {
	Throughput bool `json:"throughput"`
	Latency    bool `json:"latency"`
	Server     bool `json:"server"`
	Client     bool `json:"client"`
}

func statsHandler(w http.ResponseWriter, r *http.Request) {
	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		panic(err)
	}

	req := new(Request)
	err = json.Unmarshal(body, req)
	if err != nil {
		panic(err)
	}

	if req.Throughput {
		if req.Server {
			log.Println("running throughput server")
			opt := throughput.DefaultOptions()
			err := json.NewEncoder(w).Encode(opt)
			if err != nil {
				panic(err)
			}
			go func() {
				err = throughput.NewThroughputMeter(opt).Server()
				if err != nil {
					log.Println(err)
				}
			}()
		}

		if req.Client {
			log.Println("running throughput client")
			func() {
				resp, err := throughput.NewThroughputMeter(throughput.DefaultOptions()).Client()
				if err != nil {
					log.Println(err)
				}
				err = json.NewEncoder(w).Encode(resp)
				if err != nil {
					log.Println(err)
				}
			}()
		}
	}

	if req.Latency {
		if req.Server {
			log.Println("running latency server")
			opt := latency.DefaultOptions()
			err := json.NewEncoder(w).Encode(opt)
			if err != nil {
				panic(err)
			}
			go func() {
				err = latency.NewLatencyMeter(opt).Server()
				if err != nil {
					log.Println(err)
				}
			}()
		}

		if req.Client {
			log.Println("running latency client")
			func() {
				result, err := latency.NewLatencyMeter(latency.DefaultOptions()).Client()
				if err != nil {
					log.Println(err)
				}

				err = json.NewEncoder(w).Encode(result)
				if err != nil {
					log.Println(err)
				}
			}()
		}
	}

}

func (s *server) Start() error {
	http.HandleFunc("/stats", statsHandler)
	return http.ListenAndServe(s.addr, nil)
}

func main() {
	s := NewServer(":8080")
	log.Fatal(s.Start())
}
