package latency

import (
	"fmt"
	"log"
	"net"
	"time"
)

type Options struct {
	MsgSize     int    `json:"msgSize"`
	NumPings    int    `json:"numPings"`
	TcpAddress  string `json:"tcpAddress"`
	UnixAddress string `json:"unixAddress"`
	UnixDomain  bool   `json:"unixDomain"`
}

func DefaultOptions() Options {
	return Options{
		MsgSize:     128,
		NumPings:    50000,
		TcpAddress:  ":13501",
		UnixAddress: "/tmp/lat_benchmark.sock",
		UnixDomain:  false,
	}
}

type Result struct {
	ElapsedTime time.Duration `json:"elapsedTime"`
	NumPings    int           `json:"numPings"`
	AvgLatency  time.Duration `json:"AvgLatency"`
}

type latencyMeter struct {
	Options
}

func NewLatencyMeter(options Options) *latencyMeter {
	return &latencyMeter{
		Options: options,
	}
}

// DomainAndAddress returns the domain,address pair for net functions to connect
// to, depending on the value of the benchmate.UnixDomain flag.
func (lm *latencyMeter) DomainAndAddress() (func(string, string) (net.Conn, error), string, string) {
	if lm.UnixDomain {
		return net.Dial, "unix", lm.UnixAddress
	} else {
		dialer := &net.Dialer{
			LocalAddr: &net.TCPAddr{
				Port: 13504,
			},
		}

		return dialer.Dial, "tcp", lm.TcpAddress
	}
}

func (lm *latencyMeter) Server() error {
	_, domain, address := lm.DomainAndAddress()
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

func (lm *latencyMeter) ClientConn(conn net.Conn) (*Result, error) {
	buf := make([]byte, lm.MsgSize)
	t1 := time.Now()
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
	}
	elapsed := time.Since(t1)

	totalpings := lm.NumPings * 2
	log.Println("Client done")

	return &Result{
		ElapsedTime: elapsed,
		NumPings:    int(totalpings),
		AvgLatency:  time.Duration(int(elapsed.Nanoseconds()) / totalpings),
	}, nil
}

func (lm *latencyMeter) Client() (*Result, error) {
	log.Println("latency client running")
	dial, domain, address := lm.DomainAndAddress()
	conn, err := dial(domain, address)
	if err != nil {
		return nil, err
	}
	defer conn.Close()
	return lm.ClientConn(conn)
}
