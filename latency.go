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
	"fmt"
	"net"
	"time"
)

//// LatencyOptions holds the options for the latency benchmarks.
//type LatencyOptions struct {
//	MsgSize    int    `json:"msgSize"`    // size of messages in bytes
//	Addr       string `json:"tcpAddress"` // server listens on this address
//	UnixDomain bool   `json:"unixDomain"` // set to true when using unix domain sockets
//	ClientPort int    `json:"clientPort"` // local port used by client
//	Timeout    int    `json:"timeout"`    // in milliseconds
//}

// Result holds the results of a latency benchmark.
type LatencyResult struct {
	ElapsedTime time.Duration `json:"elapsedTime"` // time elapsed in nanoseconds
	NumMsg      int           `json:"numPings"`    // number of pings sent
	AvgLatency  time.Duration `json:"avgLatency"`  // average latency in nanoseconds
}

type LatencyServer struct {
	msgSize int
	numMsg  int
}

func NewLatencyServer(msgSize, numMsg int) LatencyServer {
	return LatencyServer{
		msgSize: msgSize,
		numMsg:  numMsg,
	}
}

func NewLatencyClient(msgSize, numMsg, timeout int) LatencyClient {
	return LatencyClient{
		msgSize: msgSize,
		timeout: timeout,
		numMsg:  numMsg,
	}
}

type LatencyClient struct {
	msgSize int
	numMsg  int
	timeout int
}

func (o LatencyServer) Run(l net.Listener) error {
	conn, err := l.Accept()
	if err != nil {
		return err
	}
	defer conn.Close()

	buf := make([]byte, o.msgSize)
	for i := 0; i < o.numMsg; i++ {
		nread, err := conn.Read(buf)
		if err != nil {
			return err
		}
		if nread != o.msgSize {
			return fmt.Errorf("bad nread = %d", nread)
		}
		nwrite, err := conn.Write(buf)
		if err != nil {
			return err
		}
		if nwrite != o.msgSize {
			return fmt.Errorf("bad nwrite = %d", nwrite)
		}
	}

	return nil
}

func (lm LatencyClient) Run(conn net.Conn) (*LatencyResult, error) {
	buf := make([]byte, lm.msgSize)
	t1 := time.Now()
	stopTime := t1.Add(time.Duration(lm.timeout) * time.Millisecond)
	pingsSent := 0
	for n := 0; n < lm.numMsg; n++ {
		nwrite, err := conn.Write(buf)
		if err != nil {
			return nil, err
		}
		if nwrite != lm.msgSize {
			return nil, fmt.Errorf("bad nwrite = %d", nwrite)
		}
		nread, err := conn.Read(buf)
		if err != nil {
			return nil, err
		}
		if nread != lm.msgSize {
			return nil, fmt.Errorf("bad nread = %d", nread)
		}

		pingsSent = n + 1
		if time.Now().After(stopTime) {
			break
		}
	}
	elapsed := time.Since(t1)
	totalpings := pingsSent * 2

	return &LatencyResult{
		ElapsedTime: elapsed,
		NumMsg:      totalpings,
		AvgLatency:  elapsed / time.Duration(totalpings),
	}, nil
}
