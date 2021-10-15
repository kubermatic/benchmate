package main

import (
	"flag"
	"io"
	"net"
	"os"
	"sync"
)

func main() {
	var from, to string
	flag.StringVar(&from, "from", "/tmp/udsToTcp.sock", "uds socket")
	flag.StringVar(&to, "to", "127.0.0.1:8080", "tcp socket")
	flag.Parse()

	if err := os.RemoveAll(from); err != nil {
		panic(err)
	}

	l, err := net.Listen("unix", from)
	if err != nil {
		panic(err)
	}
	defer l.Close()

	toConn, err := net.Dial("tcp", to)
	if err != nil {
		panic(err)
	}
	defer toConn.Close()

	conn, err := l.Accept()
	if err != nil {
		panic(err)
	}
	defer conn.Close()

	var wg sync.WaitGroup

	wg.Add(1)
	go func() {
		defer wg.Done()
		_, err := io.Copy(conn, toConn)
		if err != nil {
			panic(err)
		}
	}()

	wg.Add(1)
	go func() {
		defer wg.Done()
		_, err := io.Copy(toConn, conn)
		if err != nil {
			panic(err)
		}
	}()

	wg.Wait()

	//var buf [32 * 1024]byte
	//for {
	//
	//	n, err := conn.Read(buf[:])
	//	if err != nil {
	//		panic(err)
	//	}
	//
	//	// TODO: does this write all n bytes in single call?
	//	n, err = toConn.Write(buf[:n])
	//	if err != nil {
	//		panic(err)
	//	}
	//
	//}

}
