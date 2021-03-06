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
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"math/rand"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestHandlers(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/benchmate/latency", LatencyHandler)
	mux.HandleFunc("/benchmate/throughput", ThroughputHandler)

	s := httptest.NewServer(mux)
	defer s.Close()

	c := httptest.NewServer(mux)
	defer c.Close()

	rand.Seed(time.Now().UnixNano())

	tpOpt := DefaultThroughputOptions()
	latOpt := DefaultLatencyOptions()

	// adjust for test

	tpOpt.ClientPort = randPort()
	tpOpt.MsgSize >>= 5
	tpOpt.Addr = fmt.Sprintf(":%d", randPort())
	tpOpt.NumMsg = 100

	latOpt.ClientPort = randPort()
	latOpt.Addr = fmt.Sprintf(":%d", randPort())
	latOpt.NumMsg = 100

	tests := []struct {
		name      string
		runServer func()
		runClient func()
	}{
		{
			name: "latency",
			runServer: func() {
				_, err := doReq(s.URL+"/benchmate/latency", &LatencyRequest{
					Options: latOpt,
				})
				if err != nil {
					t.Errorf("%s: %v", t.Name(), err)
				}
			},
			runClient: func() {
				data, err := doReq(s.URL+"/benchmate/latency", &LatencyRequest{
					Options: latOpt,
					Client:  true,
				})
				if err != nil {
					t.Error(err)
				}

				result := new(LatencyResult)
				err = json.Unmarshal(data, &result)
				if err != nil {
					t.Error(err)
				}

				if result.NumMsg != latOpt.NumMsg*2 {
					t.Errorf("%s: expected %d pings, got %d", t.Name(), latOpt.NumMsg*2, result.NumMsg)
				}

				if result.AvgLatency == 0 {
					t.Errorf("%s: %v", t.Name(), "latency is 0")
				}

				t.Log("latency result:", prettyJSON(result))
			},
		},
		{
			name: "throughput",
			runServer: func() {
				_, err := doReq(s.URL+"/benchmate/throughput", &ThroughputRequest{
					Options: tpOpt,
				})
				if err != nil {
					t.Errorf("%s: %v", t.Name(), err)
				}
			},
			runClient: func() {
				data, err := doReq(s.URL+"/benchmate/throughput", &ThroughputRequest{
					Options: tpOpt,
					Client:  true,
				})
				if err != nil {
					t.Error(err)
				}

				result := new(ThroughputResult)
				err = json.Unmarshal(data, &result)
				if err != nil {
					t.Error(err)
				}

				if result.NumMsg != tpOpt.NumMsg {
					t.Errorf("%s: expected %d messages, got %d", t.Name(), tpOpt.NumMsg, result.NumMsg)
				}

				if result.AvgThroughput == 0 {
					t.Errorf("%s: %v", t.Name(), "throughput is 0")
				}

				t.Log("throughput result:", prettyJSON(result))
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			go test.runServer()
			time.Sleep(time.Second) // give server time to start
			test.runClient()
		})
	}
}

func randPort() int {
	return 1234 + rand.Intn(1234)
}

func prettyJSON(x interface{}) string {
	b, err := json.MarshalIndent(x, "", "  ")
	if err != nil {
		panic(err)
	}
	return string(b)
}

func toReader(r interface{}) *bytes.Reader {
	data, err := json.Marshal(r)
	if err != nil {
		panic(err)
	}
	return bytes.NewReader(data)
}

func doReq(addr string, r interface{}) ([]byte, error) {
	req, err := http.NewRequest("GET", addr, toReader(r))
	if err != nil {
		return nil, err
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer func() { _ = resp.Body.Close() }()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	return body, nil
}
