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

// LatencyOptions holds the options for the latency benchmarks.
type LatencyOptions struct {
	MsgSize     int    `json:"msgSize"`     // size of messages in bytes
	NumPings    int    `json:"numPings"`    // number of pings to send
	TcpAddress  string `json:"tcpAddress"`  // server listens on this address
	UnixAddress string `json:"unixAddress"` // server listens on this address when UnixDomain is true
	UnixDomain  bool   `json:"unixDomain"`  // set to true when using unix domain sockets
	ClientPort  int    `json:"clientPort"`  // local port used by client
	Timeout     int    `json:"timeout"`     // in milliseconds
}

// DefaultOptions the default values of the options.
//  {
//		MsgSize:     128,
//		NumPings:    1000,
//		TcpAddress:  ":13501",
//		UnixAddress: "/tmp/lat_benchmark.sock",
//		UnixDomain:  false,
//		ClientPort:  13504,
//		Timeout:     120000,
//  }
func DefaultLatencyOptions() LatencyOptions {
	return LatencyOptions{
		MsgSize:     128,
		NumPings:    1000,
		TcpAddress:  ":13501",
		UnixAddress: "/tmp/lat_benchmark.sock",
		UnixDomain:  false,
		ClientPort:  13504,
		Timeout:     120000,
	}
}

// Result holds the results of a latency benchmark.
type LatencyResult struct {
	ElapsedTime time.Duration `json:"elapsedTime"` // time elapsed in nanoseconds
	NumPings    int           `json:"numPings"`    // number of pings sent
	AvgLatency  time.Duration `json:"avgLatency"`  // in nanoseconds ( elapsedTime / numPings )
}

// LatencyMeter allows you to run clients and servers to measure latency
// of the network between them.
type LatencyMeter struct {
	LatencyOptions
}

// NewLatencyMeter returns a new LatencyMeter instance.
func NewLatencyMeter(options LatencyOptions) *LatencyMeter {
	return &LatencyMeter{
		LatencyOptions: options,
	}
}

// Server starts a latency benchmark server.
// Once a client connects and starts sending messages, the server reads and replies with the same message.
// It returns when client closes the connection.
func (lm *LatencyMeter) Server() error {
	_, domain, address := lm.domainAndAddress()
	l, err := net.Listen(domain, address)
	if err != nil {
		return err
	}
	defer l.Close()

	conn, err := l.Accept()
	if err != nil {
		return err
	}
	defer conn.Close()

	buf := make([]byte, lm.MsgSize)
	for n := 0; n < lm.NumPings; n++ {
		nread, err := conn.Read(buf)
		if err != nil {
			return err
		}
		if nread != lm.MsgSize {
			return fmt.Errorf("bad nread = %d", nread)
		}
		nwrite, err := conn.Write(buf)
		if err != nil {
			return err
		}
		if nwrite != lm.MsgSize {
			return fmt.Errorf("bad nwrite = %d", nwrite)
		}
	}

	return nil
}

// ClientConn like Client with a connection argument.
func (lm *LatencyMeter) ClientConn(conn net.Conn) (*LatencyResult, error) {
	buf := make([]byte, lm.MsgSize)
	t1 := time.Now()
	stopTime := t1.Add(time.Duration(lm.Timeout) * time.Millisecond)
	pingsSent := 0
	for n := 0; n < lm.NumPings; n++ {
		nwrite, err := conn.Write(buf)
		if err != nil {
			return nil, err
		}
		if nwrite != lm.MsgSize {
			return nil, fmt.Errorf("bad nwrite = %d", nwrite)
		}
		nread, err := conn.Read(buf)
		if err != nil {
			return nil, err
		}
		if nread != lm.MsgSize {
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
		NumPings:    totalpings,
		AvgLatency:  time.Duration(int(elapsed.Nanoseconds()) / totalpings),
	}, nil
}

// Client sends a message of MsgSize bytes to the server and reads the reply from the server.
// It calculates average latency over all the request/responses and returns the result.
func (lm *LatencyMeter) Client() (*LatencyResult, error) {
	dial, domain, address := lm.domainAndAddress()
	conn, err := dial(domain, address)
	if err != nil {
		return nil, err
	}
	defer conn.Close()
	return lm.ClientConn(conn)
}

func (lm *LatencyMeter) domainAndAddress() (func(string, string) (net.Conn, error), string, string) {
	if lm.UnixDomain {
		return net.Dial, "unix", lm.UnixAddress
	} else {
		dialer := &net.Dialer{
			LocalAddr: &net.TCPAddr{
				Port: lm.ClientPort,
			},
		}
		return dialer.Dial, "tcp", lm.TcpAddress
	}
}
