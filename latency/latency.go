package latency

import (
	"flag"
	"fmt"
	"github.com/pratikdeoghare/benchmate"
	"net"
	"os"
	"time"
)

var MsgSize = flag.Int("lat_msgsize", 128, "Message size in each ping")
var NumPings = flag.Int("lat_numping", 50000, "Number of pings to measure")

var TcpAddress = "127.0.0.1:13501"
var UnixAddress = "/tmp/lat_benchmark.sock"

// DomainAndAddress returns the domain,address pair for net functions to connect
// to, depending on the value of the UnixDomain flag.
func DomainAndAddress() (string, string) {
	if *benchmate.UnixDomain {
		return "unix", UnixAddress
	} else {
		return "tcp", TcpAddress
	}
}

func Server() error {
	if *benchmate.UnixDomain {
		if err := os.RemoveAll(UnixAddress); err != nil {
			panic(err)
		}
	}

	domain, address := DomainAndAddress()
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
	domain, address := DomainAndAddress()
	conn, err := net.Dial(domain, address)
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
