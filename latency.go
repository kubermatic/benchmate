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

// LatencyResult contains the details of a latency estimation run.
// AvgLatency = NumMsg / ElapsedTime.
type LatencyResult struct {
	ElapsedTime time.Duration `json:"elapsedTime"` // time elapsed in nanoseconds
	NumMsg      int           `json:"numPings"`    // number of pings sent
	AvgLatency  time.Duration `json:"avgLatency"`  // average latency in nanoseconds
}

// LatencyServer holds parameters for the server side of latency estimation.
type LatencyServer struct {
	msgSize int
	numMsg  int
}

// NewLatencyServer creates a new instance of LatencyServer.
func NewLatencyServer(msgSize, numMsg int) LatencyServer {
	return LatencyServer{
		msgSize: msgSize,
		numMsg:  numMsg,
	}
}

// Run waits to get connection from a client. It then reads the message sent
// by the client and replies back with the same message. This allows client to
// estimate the latency.
//
// It accepts a listener. The following code will run the server at port 8888.
//
//	l, _ := net.Listen("tcp", ":8888")
//	s.Run(l)
//
// This will run the server at unix domain socket.
//
//	l, _ := net.Listen("unix", "/tmp/tp-srv")
//	s.Run(l)
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

// LatencyClient holds parameters for the client side of latency estimation.
type LatencyClient struct {
	msgSize int
	numMsg  int
	timeout int
}

// NewLatencyClient returns an instance of LatencyClient. You can
// call its Run method to start client for latency estimation.
func NewLatencyClient(msgSize, numMsg, timeout int) LatencyClient {
	return LatencyClient{
		msgSize: msgSize,
		timeout: timeout,
		numMsg:  numMsg,
	}
}

// Run sends the messages over the connection and
// reads the reply back from the server. After the configured number
// of messages are exchanged or the timeout is reached, it estimates the
// latency by total time spent / ( 2 * # messages sent).
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
