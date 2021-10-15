package latency

import (
	"flag"
	"fmt"
	"github.com/pratikdeoghare/benchmate"
	"log"
	"net"
	"time"
)

var MsgSize = flag.Int("lat_msgsize", 128, "Message size in each ping")
var NumPings = flag.Int("lat_numping", 50000, "Number of pings to measure")

var TcpAddress = flag.String("lat_tcp_addr", "127.0.0.1:13501", "tcp addr of latency server")
var UnixAddress = flag.String("lat_uds_addr", "/tmp/lat_benchmark.sock", "uds addr of latency server")

// DomainAndAddress returns the domain,address pair for net functions to connect
// to, depending on the value of the benchmate.UnixDomain flag.
func DomainAndAddress() (func(string, string) (net.Conn, error), string, string) {
	if *benchmate.UnixDomain {
		return net.Dial, "unix", *UnixAddress
	} else {
		dialer := &net.Dialer{
			LocalAddr: &net.TCPAddr{
				IP:   net.ParseIP("127.0.0.1"),
				Port: 13504,
			},
		}

		return dialer.Dial, "tcp", *TcpAddress
	}
}
func Server() error {
	_, domain, address := DomainAndAddress()
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

	buf := make([]byte, *MsgSize)
	for n := 0; n < *NumPings; n++ {
		nread, err := conn.Read(buf)
		if err != nil {
			return err
		}
		if nread != *MsgSize {
			return fmt.Errorf("bad nread = %d", nread)
		}
		nwrite, err := conn.Write(buf)
		if err != nil {
			return err
		}
		if nwrite != *MsgSize {
			return fmt.Errorf("bad nwrite = %d", nwrite)
		}
	}

	time.Sleep(50 * time.Millisecond)
	return nil
}

func Client() error {
	// This is the client code in the main goroutine.
	dial, domain, address := DomainAndAddress()
	conn, err := dial(domain, address)
	if err != nil {
		return err
	}
	defer conn.Close()

	buf := make([]byte, *MsgSize)
	t1 := time.Now()
	for n := 0; n < *NumPings; n++ {
		nwrite, err := conn.Write(buf)
		if err != nil {
			return err
		}
		if nwrite != *MsgSize {
			return fmt.Errorf("bad nwrite = %d", nwrite)
		}
		nread, err := conn.Read(buf)
		if err != nil {
			return err
		}
		if nread != *MsgSize {
			return fmt.Errorf("bad nread = %d", nread)
		}
	}
	elapsed := time.Since(t1)

	totalpings := int64(*NumPings * 2)
	fmt.Println("Client done")
	fmt.Printf("%d pingpongs took %d ns; avg. latency %d ns\n",
		totalpings, elapsed.Nanoseconds(),
		elapsed.Nanoseconds()/totalpings)

	time.Sleep(50 * time.Millisecond)
	return nil
}
