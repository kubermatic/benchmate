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


package throughput

import (
	"fmt"
	"log"
	"net"
	"os"
	"time"
)

type Options struct {
	MsgSize     int    `json:"msgSize"`
	NumMsg      int    `json:"numMsg"`
	TcpAddress  string `json:"tcpAddress"`
	UnixAddress string `json:"unixAddress"`
	UnixDomain  bool   `json:"unixDomain"`
	ClientPort  int    `json:"clientPort"`
	Timeout     int    `json:"timeout"` // in milliseconds
}

func DefaultOptions() Options {
	return Options{
		MsgSize:     256 * 1024,
		NumMsg:      100000,
		TcpAddress:  ":13500",
		UnixAddress: "/tmp/tp_benchmark.sock",
		UnixDomain:  false,
		ClientPort:  13503,
		Timeout:     120000,
	}
}

type Result struct {
	MsgSize             int           `json:"msgSize"`
	NumMsg              int           `json:"numMsg"`
	TotalData           int           `json:"totalData"`
	Elapsed             time.Duration `json:"elapsed"`
	ThroughputMBPerSec  float64       `json:"throughputMbPerSec"`
	ThroughputMsgPerSec float64       `json:"throughputMsgPerSec"`
}

type throughputMeter struct {
	Options
}

// NewThroughputMeter returns a new throughputMeter.
func NewThroughputMeter(options Options) *throughputMeter {
	return &throughputMeter{
		Options: options,
	}
}

// domainAddress returns the domain,address pair for net functions to connect
// to, depending on the value of the benchmate.UnixDomain flag.
func (tm *throughputMeter) domainAddress() (func(string, string) (net.Conn, error), string, string) {
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

func (tm *throughputMeter) Server() error {
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

	log.Println("connected ", conn.LocalAddr(), conn.RemoteAddr())

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

func (tm *throughputMeter) ClientConn(conn net.Conn) (*Result, error) {
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

		msgSent = n
		if time.Now().After(stopTime) {
			break
		}
	}
	elapsed := time.Since(t1)

	totaldata := int64(msgSent * tm.MsgSize)
	fmt.Println("Client done")
	fmt.Printf("Sent %d msg in %d ns; throughput %d msg/sec (%d MB/sec)\n",
		msgSent, elapsed,
		(int64(msgSent)*1000000000)/elapsed.Nanoseconds(),
		(totaldata*1000)/elapsed.Nanoseconds())

	return &Result{
		MsgSize:             tm.MsgSize,
		NumMsg:              msgSent,
		TotalData:           int(totaldata),
		Elapsed:             elapsed,
		ThroughputMBPerSec:  float64((totaldata * 1000) / elapsed.Nanoseconds()),
		ThroughputMsgPerSec: float64((int64(msgSent) * 1000000000) / elapsed.Nanoseconds()),
	}, nil
}

func (tm *throughputMeter) Client() (*Result, error) {
	dial, domain, address := tm.domainAddress()
	conn, err := dial(domain, address)
	if err != nil {
		return nil, err
	}
	defer conn.Close()
	return tm.ClientConn(conn)
}
