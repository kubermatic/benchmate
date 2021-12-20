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

// ThroughputResult contains the details of a throughput estimation run.
// AvgThroughput = MsgSize * NumMsg / Elapsed in MB/s.
type ThroughputResult struct {
	MsgSize       int           `json:"msgSize"`       // size of a message in bytes
	NumMsg        int           `json:"numMsg"`        // number of messages received from the client
	Elapsed       time.Duration `json:"elapsed"`       // total time
	AvgThroughput float64       `json:"avgThroughput"` // avg throughput in MB/s
}

// ThroughputClient holds parameters for the client side of throughput estimation.
type ThroughputClient struct {
	msgSize int
	numMsg  int
	timeout int
}

// NewThroughputClient returns an instance of ThroughputClient. You can
// call its Run method to start client for throughput estimation.
func NewThroughputClient(msgSize, numMsg, timeout int) ThroughputClient {
	return ThroughputClient{
		msgSize: msgSize,
		timeout: timeout,
		numMsg:  numMsg,
	}
}

func (c ThroughputClient) Run(conn net.Conn) (*ThroughputResult, error) {
	buf := make([]byte, c.msgSize)
	t1 := time.Now()
	stopTime := t1.Add(time.Duration(c.timeout) * time.Millisecond)
	msgSent := 0

	for n := 0; n < c.numMsg; n++ {
		nwrite, err := conn.Write(buf)
		if err != nil {
			return nil, err
		}
		if nwrite != c.msgSize {
			return nil, fmt.Errorf("bad nwrite = %d", nwrite)
		}

		msgSent = n + 1
		if time.Now().After(stopTime) {
			break
		}
	}

	elapsed := time.Since(t1)

	return &ThroughputResult{
		MsgSize:       c.msgSize,
		NumMsg:        msgSent,
		Elapsed:       elapsed,
		AvgThroughput: float64(msgSent*c.msgSize*1000) / float64(elapsed.Nanoseconds()),
	}, nil
}

// ThroughputServer holds parameters for the server side of throughput estimation.
type ThroughputServer struct {
	msgSize int
}

// NewThroughputServer creates a new instance of ThroughputServer.
func NewThroughputServer(msgSize int) ThroughputServer {
	return ThroughputServer{
		msgSize: msgSize,
	}
}

// Run waits to get connection from a client.
func (s ThroughputServer) Run(l net.Listener) error {
	conn, err := l.Accept()
	if err != nil {
		return err
	}
	defer conn.Close()
	buf := make([]byte, s.msgSize)
	k := 0
	for {
		nread, err := conn.Read(buf)
		if err != nil {
			return err
		}

		fmt.Println("got", k, nread, s.msgSize)
		k++

		if nread == 0 {
			break
		}
	}

	return nil
}
