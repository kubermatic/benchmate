package main

import (
	"context"
	"fmt"
	"google.golang.org/grpc"
	"k8s.io/klog/v2"
	"log"
	"net"
	"sigs.k8s.io/apiserver-network-proxy/konnectivity-client/pkg/client"
	"github.com/pratikdeoghare/benchmate/throughput"
	"github.com/pratikdeoghare/benchmate/latency"
)

func main() {
	proxyUDSName := "/tmp/uds-proxy"
	dialOption := grpc.WithContextDialer(func(context.Context, string) (net.Conn, error) {
		// Ignoring addr and timeout arguments:
		// addr - comes from the closure
		// timeout - is turned off as this is test code and eases debugging.
		c, err := net.DialTimeout("unix", proxyUDSName, 0)
		if err != nil {
			klog.ErrorS(err, "failed to create connection to uds", "name", proxyUDSName)
		}
		return c, err
	})
	ctx := context.Background()
	tunnel, err := client.CreateSingleUseGrpcTunnel(ctx, proxyUDSName, dialOption, grpc.WithInsecure(), grpc.WithUserAgent("o.userAgent"))
	if err != nil {
		panic(err)
	}

	requestAddress := fmt.Sprintf("%s:%d", "127.0.0.1", 13500)
	proxyConn, err := tunnel.DialContext(ctx, "tcp", requestAddress)
	if err != nil {
		panic(err)
	}


	err = throughput.ClientConn(proxyConn)
	if err != nil {
		log.Println(err)
	}

	requestAddress = fmt.Sprintf("%s:%d", "127.0.0.1", 13501)
	proxyConn, err = tunnel.DialContext(ctx, "tcp", requestAddress)
	if err != nil {
		panic(err)
	}


	err = latency.ClientConn(proxyConn)
	if err != nil {
		log.Println(err)
	}

	//
	//for {
	//	n, err := proxyConn.Write([]byte("Hello there"))
	//	if err != nil {
	//		panic(err)
	//	}
	//	fmt.Println("wrote ",n, " bytes")
	//	time.Sleep(time.Second)
	//}

}
