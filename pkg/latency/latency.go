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

package latency

import (
	"fmt"
	"net"
	"time"
)

// Options holds the options for the latency benchmark.
type Options struct {
	MsgSize     int    `json:"msgSize"`
	NumPings    int    `json:"numPings"`
	TcpAddress  string `json:"tcpAddress"`
	UnixAddress string `json:"unixAddress"`
	UnixDomain  bool   `json:"unixDomain"`
	ClientPort  int    `json:"clientPort"`
	Timeout     int    `json:"timeout"` // in milliseconds
}

// DefaultOptions returns default latency benchmark options.
//	Options{
//		MsgSize:     128,                       // in bytes
//		NumPings:    1000,                      // number of pings to send
//		TcpAddress:  ":13501",                  // tcp address to send to
//		UnixAddress: "/tmp/lat_benchmark.sock", // unix address to send to
//		UnixDomain:  false,                     // whether to use unix domain socket
//		ClientPort:  13504,                     // port that client contacts from
//		Timeout:     120000,                    // in milliseconds
//	}
func DefaultOptions() Options {
	return Options{
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
type Result struct {
	ElapsedTime time.Duration `json:"elapsedTime"`
	NumPings    int           `json:"numPings"`
	AvgLatency  time.Duration `json:"avgLatency"`
}

type LatencyMeter struct {
	Options
}

// NewLatencyMeter creates a new latency meter.
func NewLatencyMeter(options Options) *LatencyMeter {
	return &LatencyMeter{
		Options: options,
	}
}

// domainAndAddress returns the domain,address pair for net functions to connect
// to, depending on the value of the benchmate.UnixDomain flag.
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

// Server starts a latency benchmark server. It returns once it participates
// in one benchmark with the client.
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

// ClientConn starts a latency benchmark client. It is like LatencyMeter.Client
// but it takes a net.Conn as an argument. This is useful when you want to custom
// net.Conn as in the case of konnectivity-benchmate.
func (lm *LatencyMeter) ClientConn(conn net.Conn) (*Result, error) {
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

		pingsSent = n
		if time.Now().After(stopTime) {
			break
		}
	}
	elapsed := time.Since(t1)
	totalpings := pingsSent * 2

	return &Result{
		ElapsedTime: elapsed,
		NumPings:    int(totalpings),
		AvgLatency:  time.Duration(int(elapsed.Nanoseconds()) / totalpings),
	}, nil
}

// Client starts a latency benchmark client and returns the results after
// the benchmark is complete.
func (lm *LatencyMeter) Client() (*Result, error) {
	dial, domain, address := lm.domainAndAddress()
	conn, err := dial(domain, address)
	if err != nil {
		return nil, err
	}
	defer conn.Close()
	return lm.ClientConn(conn)
}
