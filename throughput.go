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

type ThroughputClient struct {
	msgSize int
	numMsg  int
	timeout int
}

type ThroughputServer struct {
	msgSize int
}

type ThroughputResult struct {
	MsgSize       int           `json:"msgSize"`       // size of the messages in bytes
	NumMsg        int           `json:"numMsg"`        // number of messages sent
	Elapsed       time.Duration `json:"elapsed"`       // total time
	AvgThroughput float64       `json:"avgThroughput"` // avg throughput in MB/s
}

func NewThroughputServer(msgSize int) ThroughputServer {
	return ThroughputServer{
		msgSize: msgSize,
	}
}

func NewThroughputClient(msgSize, numMsg, timeout int) ThroughputClient {
	return ThroughputClient{
		msgSize: msgSize,
		timeout: timeout,
		numMsg:  numMsg,
	}
}

func (s ThroughputServer) Run(l net.Listener) error {
	conn, err := l.Accept()
	if err != nil {
		return err
	}
	defer conn.Close()
	buf := make([]byte, s.msgSize)
	for {
		nread, err := conn.Read(buf)
		if err != nil {
			return err
		}
		if nread == 0 {
			break
		}
	}

	return nil
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
