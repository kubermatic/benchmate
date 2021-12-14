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
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"math/rand"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/kubermatic/benchmate/latency"
	"github.com/kubermatic/benchmate/throughput"
)

func TestName(t *testing.T) {
	s := httptest.NewServer(http.HandlerFunc(BenchmateHandler))
	defer s.Close()
	rand.Seed(time.Now().UnixNano())

	randPort := func() int {
		return 1234 + rand.Intn(1<<16)
	}

	tpOpt := throughput.DefaultOptions()
	latOpt := latency.DefaultOptions()
	tpOpt.ClientPort = randPort()
	tpOpt.TcpAddress = fmt.Sprintf(":%d", randPort())
	latOpt.ClientPort = randPort()
	latOpt.TcpAddress = fmt.Sprintf(":%d", randPort())
	latOpt.NumPings = 100000

	resp, err := doReq(s.URL, &Request{
		ThroughputOptions: &tpOpt,
		LatencyOptions:    &latOpt,
	})
	if err != nil {
		t.Fatal(err)
	}

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		t.Fatal(err)
	}
	fmt.Println(string(body))

	c := httptest.NewServer(http.HandlerFunc(BenchmateHandler))
	defer c.Close()
	resp, err = doReq(c.URL, &Request{
		ThroughputOptions: &tpOpt,
		LatencyOptions:    &latOpt,
		Client:            true,
	})
	body, err = ioutil.ReadAll(resp.Body)
	if err != nil {
		t.Fatal(err)
	}
	fmt.Println(string(body))
}

func toReader(r *Request) *bytes.Reader {
	data, err := json.Marshal(r)
	if err != nil {
		panic(err)
	}
	return bytes.NewReader(data)
}

func doReq(addr string, r *Request) (*http.Response, error) {
	req, err := http.NewRequest("GET", addr, toReader(r))
	if err != nil {
		return nil, err
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	return resp, nil
}
