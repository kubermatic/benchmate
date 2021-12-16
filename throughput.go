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
	"os"
	"time"
)

// ThroughputOptions holds the options for a throughput benchmark.
type ThroughputOptions struct {
	MsgSize     int    `json:"msgSize"`     // size of messages in bytes
	NumMsg      int    `json:"numMsg"`      // number of messages to exchange before timeout
	TcpAddress  string `json:"tcpAddress"`  // server listens on this address
	UnixAddress string `json:"unixAddress"` // server listens on this address when UnixDomain is true
	UnixDomain  bool   `json:"unixDomain"`  // set to true when using unix domain sockets
	ClientPort  int    `json:"clientPort"`  // local port used by client
	Timeout     int    `json:"timeout"`     // in milliseconds
}

// DefaultOptions the default values of the options.
//	{
//		MsgSize:     512 * 1024,
//		NumMsg:      100000,
//		TcpAddress:  ":13500",
//		UnixAddress: "/tmp/tp_benchmark.sock",
//		UnixDomain:  false,
//		ClientPort:  13503,
//		Timeout:     120000,
//	}
func DefaultThroughputOptions() ThroughputOptions {
	return ThroughputOptions{
		MsgSize:     256 * 1024,
		NumMsg:      100000,
		TcpAddress:  ":13500",
		UnixAddress: "/tmp/tp_benchmark.sock",
		UnixDomain:  false,
		ClientPort:  13503,
		Timeout:     120000,
	}
}

// Result holds the results of a throughput benchmark.
type ThroughputResult struct {
	MsgSize             int           `json:"msgSize"`             // size of the messages in bytes
	NumMsg              int           `json:"numMsg"`              // number of messages exchanged
	TotalData           int           `json:"totalData"`           // total bytes exchanged
	Elapsed             time.Duration `json:"elapsed"`             // total time elapsed
	ThroughputMBPerSec  float64       `json:"throughputMBPerSec"`  // throughput in MB/s
	ThroughputMsgPerSec float64       `json:"throughputMsgPerSec"` // throughput in msg/s
}

// ThroughputMeter allows you to run clients and servers to measure throughput
// of the network between them.
type ThroughputMeter struct {
	ThroughputOptions
}

// NewThroughputMeter returns a new ThroughputMeter instance.
func NewThroughputMeter(options ThroughputOptions) *ThroughputMeter {
	return &ThroughputMeter{
		ThroughputOptions: options,
	}
}

// Server starts a throughput benchmark server.
// Once a client connects and starts sending data, the server reads as much data as it can.
// It returns when the client closes the connection.
func (tm *ThroughputMeter) Server() error {
	if tm.UnixDomain {
		if err := os.RemoveAll(tm.UnixAddress); err != nil {
			panic(err)
		}
	}

	_, domain, address := tm.domainAddress()
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

	buf := make([]byte, tm.MsgSize)
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

// ClientConn like Client with a connection argument.
func (tm *ThroughputMeter) ClientConn(conn net.Conn) (*ThroughputResult, error) {
	buf := make([]byte, tm.MsgSize)
	t1 := time.Now()
	stopTime := t1.Add(time.Duration(tm.Timeout) * time.Millisecond)
	msgSent := 0

	for n := 0; n < tm.NumMsg; n++ {
		nwrite, err := conn.Write(buf)
		if err != nil {
			return nil, err
		}
		if nwrite != tm.MsgSize {
			return nil, fmt.Errorf("bad nwrite = %d", nwrite)
		}

		msgSent = n + 1
		if time.Now().After(stopTime) {
			break
		}
	}

	elapsed := time.Since(t1)
	totaldata := int64(msgSent * tm.MsgSize)

	return &ThroughputResult{
		MsgSize:             tm.MsgSize,
		NumMsg:              msgSent,
		TotalData:           int(totaldata),
		Elapsed:             elapsed,
		ThroughputMBPerSec:  float64((totaldata * 1000) / elapsed.Nanoseconds()),
		ThroughputMsgPerSec: float64((int64(msgSent) * 1000000000) / elapsed.Nanoseconds()),
	}, nil
}

// Client tries to send NumMsg messages of size MsgSize to the server within Timeout.
// It returns the results containing throughput information.
func (tm *ThroughputMeter) Client() (*ThroughputResult, error) {
	dial, domain, address := tm.domainAddress()
	conn, err := dial(domain, address)
	if err != nil {
		return nil, err
	}
	defer conn.Close()
	return tm.ClientConn(conn)
}

func (tm *ThroughputMeter) domainAddress() (func(string, string) (net.Conn, error), string, string) {
	if tm.UnixDomain {
		return net.Dial, "unix", tm.UnixAddress
	} else {
		dialer := &net.Dialer{
			LocalAddr: &net.TCPAddr{
				Port: tm.ClientPort,
			},
		}
		return dialer.Dial, "tcp", tm.TcpAddress
	}
}
