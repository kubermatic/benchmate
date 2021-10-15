package throughput

import (
	"flag"
	"fmt"
	"github.com/pratikdeoghare/benchmate"
	"log"
	"net"
	"os"
	"time"
)

var TcpAddress = flag.String("tp_tcp_addr", "127.0.0.1:13500", "tcp addr of throughput server")
var UnixAddress = flag.String("tp_uds_addr", "/tmp/tp_benchmark.sock", "uds addr of throughput server")

var MsgSize = flag.Int("tp_msgsize", 256*1024, "Size of each message")
var NumMsg = flag.Int("tp_nummsg", 10000, "Number of messages to send")

// DomainAndAddress returns the domain,address pair for net functions to connect
// to, depending on the value of the benchmate.UnixDomain flag.
func DomainAndAddress() (func(string, string) (net.Conn, error), string, string) {
	if *benchmate.UnixDomain {
		return net.Dial, "unix", *UnixAddress
	} else {
		dialer := &net.Dialer{
			LocalAddr: &net.TCPAddr{
				IP:   net.ParseIP("127.0.0.1"),
				Port: 13503,
			},
		}

		return dialer.Dial, "tcp", *TcpAddress
	}
}

func Server() error {
	if *benchmate.UnixDomain {
		if err := os.RemoveAll(*UnixAddress); err != nil {
			panic(err)
		}
	}

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
	for {
		nread, err := conn.Read(buf)
		if err != nil {
			return err
		}
		if nread == 0 {
			break
		}
	}

	time.Sleep(50 * time.Millisecond)
	return nil
}

func Client() error {

	// This is the Client code in the main goroutine.
	dial, domain, address := DomainAndAddress()
	conn, err := dial(domain, address)
	if err != nil {
		return err
	}
	defer conn.Close()

	buf := make([]byte, *MsgSize)
	t1 := time.Now()
	for n := 0; n < *NumMsg; n++ {
		nwrite, err := conn.Write(buf)
		if err != nil {
			return err
		}
		if nwrite != *MsgSize {
			return fmt.Errorf("bad nwrite = %d")
		}
	}
	elapsed := time.Since(t1)

	totaldata := int64(*NumMsg * *MsgSize)
	fmt.Println("Client done")
	fmt.Printf("Sent %d msg in %d ns; throughput %d msg/sec (%d MB/sec)\n",
		*NumMsg, elapsed,
		(int64(*NumMsg)*1000000000)/elapsed.Nanoseconds(),
		(totaldata*1000)/elapsed.Nanoseconds())

	time.Sleep(50 * time.Millisecond)
	return nil
}
