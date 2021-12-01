package server

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/kubermatic/benchmate/latency"
	"github.com/kubermatic/benchmate/throughput"
	"io/ioutil"
	"math/rand"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
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

	c := httptest.NewServer(http.HandlerFunc(BenchmateHandler))
	defer c.Close()
	resp, err = doReq(c.URL, &Request{
		ThroughputOptions: &tpOpt,
		LatencyOptions:    &latOpt,
		Client:            true,
	})
	body, err := ioutil.ReadAll(resp.Body)
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
